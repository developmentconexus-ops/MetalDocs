package application_test

import (
	"context"
	"errors"
	"testing"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/db"
)

// fakeTemplateRouteReadinessReader records what the ADR 0086 creation gate asked
// approval, so the tests below can pin the subject addressing as well as the
// verdict.
type fakeTemplateRouteReadinessReader struct {
	ready bool
	err   error

	calls          int
	gotTenantID    string
	gotSubjectKind string
	gotSubjectKey  string
	gotExec        db.DB
}

func (f *fakeTemplateRouteReadinessReader) ActiveRouteSubjectKeys(_ context.Context, _, _ string) (map[string]struct{}, error) {
	return map[string]struct{}{}, nil
}

func (f *fakeTemplateRouteReadinessReader) HasActiveRoute(_ context.Context, exec db.DB, tenantID, subjectKind, subjectKey string) (bool, error) {
	f.calls++
	f.gotTenantID = tenantID
	f.gotSubjectKind = subjectKind
	f.gotSubjectKey = subjectKey
	f.gotExec = exec
	if f.err != nil {
		return false, f.err
	}
	return f.ready, nil
}

// newTemplateGateService builds a create-capable service whose only variable is
// the route-readiness reader, isolating the ADR 0086 gate.
func newTemplateGateService(t *testing.T, reader approvaldomain.RouteReadinessReader) (*application.Service, *fakeRepo) {
	t.Helper()
	repo := newFakeRepo()
	svc := application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}).
		WithRunner(newTxRunner(newPermissiveMockDB(t)))
	if reader != nil {
		svc = svc.WithRouteReadinessReader(reader)
	}
	return svc, repo
}

func gateCreateCmd() application.CreateTemplateCmd {
	return application.CreateTemplateCmd{
		TenantID:    "tenant-a",
		ActorUserID: "user-a",
		DocTypeCode: "po",
		Key:         "po-default",
		Name:        "PO Template",
	}
}

// TestCreateTemplate_RejectedWhenProfileHasNoActiveTemplateRoute is the ADR 0086
// config-first gate: a doc type without an ACTIVE template approval route cannot
// receive a new template, because such a template could never be submitted.
func TestCreateTemplate_RejectedWhenProfileHasNoActiveTemplateRoute(t *testing.T) {
	reader := &fakeTemplateRouteReadinessReader{ready: false}
	svc, repo := newTemplateGateService(t, reader)

	_, err := svc.CreateTemplate(context.Background(), gateCreateCmd())
	if !errors.Is(err, domain.ErrApprovalRouteMissing) {
		t.Fatalf("expected ErrApprovalRouteMissing, got %v", err)
	}
	if reader.calls != 1 {
		t.Fatalf("expected exactly 1 gate read, got %d", reader.calls)
	}
	if len(repo.templates) != 0 {
		t.Fatalf("gate must run before insert; repo persisted %d templates", len(repo.templates))
	}
}

// TestCreateTemplate_GateAddressesTemplateSubjectByDocTypeCode pins the subject
// addressing the gate shares with the template submit path: subject_kind
// 'template', subject_key = doc_type_code (ADR 0086 re-keying). A drift here
// would silently make the gate unreachable (always-false or always-true).
func TestCreateTemplate_GateAddressesTemplateSubjectByDocTypeCode(t *testing.T) {
	reader := &fakeTemplateRouteReadinessReader{ready: true}
	svc, _ := newTemplateGateService(t, reader)

	if _, err := svc.CreateTemplate(context.Background(), gateCreateCmd()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reader.gotTenantID != "tenant-a" {
		t.Errorf("gate tenant = %q, want tenant-a", reader.gotTenantID)
	}
	if reader.gotSubjectKind != string(approvaldomain.SubjectKindTemplate) {
		t.Errorf("gate subject_kind = %q, want %q", reader.gotSubjectKind, approvaldomain.SubjectKindTemplate)
	}
	if reader.gotSubjectKey != "po" {
		t.Errorf("gate subject_key = %q, want the doc type code po", reader.gotSubjectKey)
	}
	if reader.gotExec == nil {
		t.Error("gate must run on the caller's tx executor, got nil")
	}
}

// TestCreateTemplate_FailsClosedWhenRouteReaderUnwired proves the gate is
// fail-closed: a composition-root wiring miss rejects creation rather than
// allowing it.
func TestCreateTemplate_FailsClosedWhenRouteReaderUnwired(t *testing.T) {
	svc, _ := newTemplateGateService(t, nil)

	_, err := svc.CreateTemplate(context.Background(), gateCreateCmd())
	if !errors.Is(err, domain.ErrApprovalRouteMissing) {
		t.Fatalf("expected ErrApprovalRouteMissing with an unwired reader, got %v", err)
	}
}

// TestCreateTemplate_RouteReadErrorIsNotMaskedAsRouteMissing keeps an
// infrastructure failure distinguishable from a genuine "no route": a read error
// must surface as a 500-class error, never as the 409 the gate emits.
func TestCreateTemplate_RouteReadErrorIsNotMaskedAsRouteMissing(t *testing.T) {
	boom := errors.New("connection reset")
	svc, _ := newTemplateGateService(t, &fakeTemplateRouteReadinessReader{err: boom})

	_, err := svc.CreateTemplate(context.Background(), gateCreateCmd())
	if !errors.Is(err, boom) {
		t.Fatalf("expected the underlying read error to propagate, got %v", err)
	}
	if errors.Is(err, domain.ErrApprovalRouteMissing) {
		t.Fatal("a route-read failure must not be reported as ErrApprovalRouteMissing (409)")
	}
}

// TestCreateTemplate_BlankDocTypeCodeRejected pins the ADR 0086 extermination of
// generic templates: doc_type_code is mandatory, and a blank one is a 422
// validation failure that never reaches the route gate.
func TestCreateTemplate_BlankDocTypeCodeRejected(t *testing.T) {
	for _, tc := range []struct {
		name string
		code string
	}{
		{name: "empty", code: ""},
		{name: "whitespace", code: "   "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := &fakeTemplateRouteReadinessReader{ready: true}
			svc, repo := newTemplateGateService(t, reader)

			cmd := gateCreateCmd()
			cmd.DocTypeCode = tc.code

			_, err := svc.CreateTemplate(context.Background(), cmd)
			if !errors.Is(err, domain.ErrDocTypeCodeRequired) {
				t.Fatalf("expected ErrDocTypeCodeRequired, got %v", err)
			}
			if reader.calls != 0 {
				t.Errorf("validation must precede the route gate; gate read %d times", reader.calls)
			}
			if len(repo.templates) != 0 {
				t.Errorf("repo persisted %d templates, want 0", len(repo.templates))
			}
		})
	}
}

// TestCreateTemplate_SucceedsWhenDocTypeHasActiveTemplateRoute is the positive
// control for the tests above: with a route present the gate is transparent.
func TestCreateTemplate_SucceedsWhenDocTypeHasActiveTemplateRoute(t *testing.T) {
	svc, repo := newTemplateGateService(t, &fakeTemplateRouteReadinessReader{ready: true})

	got, err := svc.CreateTemplate(context.Background(), gateCreateCmd())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.Template == nil {
		t.Fatal("expected a created template")
	}
	if got.Template.DocTypeCode != "po" {
		t.Errorf("DocTypeCode = %q, want po", got.Template.DocTypeCode)
	}
	if len(repo.templates) != 1 {
		t.Fatalf("repo persisted %d templates, want 1", len(repo.templates))
	}
}
