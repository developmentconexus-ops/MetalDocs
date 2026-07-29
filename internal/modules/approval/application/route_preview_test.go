package application

import (
	"context"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/platform/db"
)

// ---------------------------------------------------------------------------
// Fakes for the PreviewRoute unit tests (SLICE 8a, unit 3.2).
//
// PreviewRoute delegates resolution to each service's OWN existing logic
// (SubmitService.resolveActiveRoute → resolver/cdRead; TemplateSubmitService →
// routeResolver.LoadActiveRouteIDBySubject) and then to the shared
// loadRoutePreviewByRouteID core (repo.LoadRoute). These fakes drive that
// wiring without a DB; the in-tx area lookup + authz.Require run against the
// in-memory submit test driver (newSubmitTestDBWithConn, submit_service_test.go).
// ---------------------------------------------------------------------------

// fakePreviewRepo serves a canned route for LoadRoute — the only repo method
// the shared preview core calls — and embeds the interface so any unexpected
// call panics rather than silently returning a zero value.
type fakePreviewRepo struct {
	infrastructure.ApprovalRepository
	route      domain.Route
	loadErr    error
	gotRouteID string
}

func (r *fakePreviewRepo) LoadRoute(_ context.Context, _ db.Tx, _, routeID string) (domain.Route, error) {
	r.gotRouteID = routeID
	if r.loadErr != nil {
		return domain.Route{}, r.loadErr
	}
	return r.route, nil
}

// fakePreviewResolver stands in for the SubmitDefaultsResolver used by both
// SubmitService.resolveActiveRoute (document path, via profile) and
// TemplateSubmitService.PreviewRoute (template path, via subject).
type fakePreviewResolver struct {
	cdID           string
	cdOK           bool
	profileRouteID string
	profileErr     error
	subjectRouteID string
	subjectErr     error

	gotSubjectKind string
	gotSubjectKey  string
}

func (f *fakePreviewResolver) LoadControlledDocumentID(_ context.Context, _ db.Tx, _, _ string) (string, bool, error) {
	return f.cdID, f.cdOK, nil
}

func (f *fakePreviewResolver) LoadActiveRouteIDByProfile(_ context.Context, _ db.Tx, _, _ string) (string, error) {
	return f.profileRouteID, f.profileErr
}

func (f *fakePreviewResolver) LoadActiveRouteIDBySubject(_ context.Context, _ db.Tx, _, subjectKind, subjectKey string) (string, error) {
	f.gotSubjectKind = subjectKind
	f.gotSubjectKey = subjectKey
	return f.subjectRouteID, f.subjectErr
}

func (f *fakePreviewResolver) LoadHeadContentHash(_ context.Context, _ db.Tx, _, _ string) (string, error) {
	return "", nil
}

// fakePreviewCDRead resolves a controlled document's profile_code for the
// document preview happy path.
type fakePreviewCDRead struct{ profileCode string }

func (f fakePreviewCDRead) ProfileCode(_ context.Context, _ db.DB, _, _ string) (string, error) {
	return f.profileCode, nil
}

func (f fakePreviewCDRead) ProcessAreaCode(_ context.Context, _ db.DB, _, _ string) (string, bool, error) {
	return "", false, nil
}

// previewRouteFixture is a two-stage route whose stage 2 carries a
// submit_choice selector with a role + area_code — the exact shape a
// chosen_actors picker must inspect before submit.
func previewRouteFixture() domain.Route {
	return domain.Route{
		ID:          "route-uuid-1",
		Name:        "ISO 9001 Two-Stage",
		TenantID:    "tenant-uuid-1",
		ProfileCode: "ISO_9001",
		Version:     1,
		Stages: []domain.Stage{
			{
				Order: 1,
				Name:  "QA Review",
				Selectors: []domain.ActorSelector{
					{Kind: domain.SelectorRoleInFixedArea, Role: "quality_approver", AreaCode: "QA"},
				},
			},
			{
				Order: 2,
				Name:  "Manager Sign-off",
				Selectors: []domain.ActorSelector{
					{Kind: domain.SelectorSubmitChoice, Role: "manager", AreaCode: "OPS"},
				},
			},
		},
	}
}

func newFixedClock() Clock {
	return fixedClock{t: time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)}
}

// ---------------------------------------------------------------------------
// SubmitService.PreviewRoute (document subject)
// ---------------------------------------------------------------------------

// TestDocumentPreviewRoute_NoActiveRoute proves that when the document has no
// linked profile (resolveActiveRoute → ErrProfileNotConfigured), PreviewRoute
// returns a zero-value RoutePreview and NO error — informational absence, not
// a failure.
func TestDocumentPreviewRoute_NoActiveRoute(t *testing.T) {
	repo := &fakePreviewRepo{route: previewRouteFixture()}
	resolver := &fakePreviewResolver{cdOK: false} // no controlled document → profileCode "" → ErrProfileNotConfigured

	svc := &SubmitService{repo: repo, emitter: &MemoryEmitter{}, clock: newFixedClock(), cdRead: fakePreviewCDRead{}, resolver: resolver}
	database := newSubmitTestDB(t, true)

	preview, err := svc.PreviewRoute(context.Background(), newTxRunner(database), "tenant-uuid-1", "doc-uuid-1")
	if err != nil {
		t.Fatalf("PreviewRoute: unexpected error: %v", err)
	}
	if preview.RouteID != "" || preview.RouteName != "" || len(preview.Stages) != 0 {
		t.Fatalf("expected zero-value RoutePreview when no active route resolves; got %+v", preview)
	}
	if repo.gotRouteID != "" {
		t.Fatalf("LoadRoute must not be called when no route resolves; got routeID %q", repo.gotRouteID)
	}
}

// TestDocumentPreviewRoute_HappyPath proves the resolved route's id, name, and
// stages+selectors are returned, including the submit_choice selector carrying
// role + area_code that a chosen_actors picker needs.
func TestDocumentPreviewRoute_HappyPath(t *testing.T) {
	repo := &fakePreviewRepo{route: previewRouteFixture()}
	resolver := &fakePreviewResolver{cdID: "cd-1", cdOK: true, profileRouteID: "route-uuid-1"}

	svc := &SubmitService{repo: repo, emitter: &MemoryEmitter{}, clock: newFixedClock(), cdRead: fakePreviewCDRead{profileCode: "ISO_9001"}, resolver: resolver}
	database := newSubmitTestDB(t, true)

	preview, err := svc.PreviewRoute(context.Background(), newTxRunner(database), "tenant-uuid-1", "doc-uuid-1")
	if err != nil {
		t.Fatalf("PreviewRoute: unexpected error: %v", err)
	}
	assertFixturePreview(t, preview)
	if repo.gotRouteID != "route-uuid-1" {
		t.Fatalf("LoadRoute called with routeID %q; want the resolved active route id", repo.gotRouteID)
	}
}

// ---------------------------------------------------------------------------
// TemplateSubmitService.PreviewRoute (template subject)
// ---------------------------------------------------------------------------

// TestTemplatePreviewRoute_NoActiveRoute proves the template path collapses
// ErrNoActiveApprovalRoute to a zero-value RoutePreview with no error.
func TestTemplatePreviewRoute_NoActiveRoute(t *testing.T) {
	repo := &fakePreviewRepo{route: previewRouteFixture()}
	resolver := &fakePreviewResolver{subjectErr: infrastructure.ErrNoActiveApprovalRoute}

	svc := &TemplateSubmitService{repo: repo, emitter: &MemoryEmitter{}, clock: newFixedClock(), routeResolver: resolver, versionReader: &stubTemplateVersionReader{}}
	database := newSubmitTestDB(t, true)

	preview, err := svc.PreviewRoute(context.Background(), newTxRunner(database), "tenant-uuid-1", "template-uuid-1")
	if err != nil {
		t.Fatalf("PreviewRoute: unexpected error: %v", err)
	}
	if preview.RouteID != "" || preview.RouteName != "" || len(preview.Stages) != 0 {
		t.Fatalf("expected zero-value RoutePreview when no active route resolves; got %+v", preview)
	}
	if repo.gotRouteID != "" {
		t.Fatalf("LoadRoute must not be called when no route resolves; got routeID %q", repo.gotRouteID)
	}
}

// TestTemplatePreviewRoute_HappyPath proves the template path resolves via
// LoadActiveRouteIDBySubject and returns the same stages+selectors shape,
// including the submit_choice selector. It also pins the ADR 0086 addressing:
// the subject key is the template's doc_type_code read through the SAME
// TemplateVersionReader port the submit path uses — never the template id —
// which is what keeps preview and submit on the same route.
func TestTemplatePreviewRoute_HappyPath(t *testing.T) {
	repo := &fakePreviewRepo{route: previewRouteFixture()}
	resolver := &fakePreviewResolver{subjectRouteID: "route-uuid-1"}

	svc := &TemplateSubmitService{repo: repo, emitter: &MemoryEmitter{}, clock: newFixedClock(), routeResolver: resolver, versionReader: &stubTemplateVersionReader{}}
	database := newSubmitTestDB(t, true)

	preview, err := svc.PreviewRoute(context.Background(), newTxRunner(database), "tenant-uuid-1", "template-uuid-1")
	if err != nil {
		t.Fatalf("PreviewRoute: unexpected error: %v", err)
	}
	assertFixturePreview(t, preview)
	if repo.gotRouteID != "route-uuid-1" {
		t.Fatalf("LoadRoute called with routeID %q; want the resolved active route id", repo.gotRouteID)
	}
	if resolver.gotSubjectKind != string(domain.SubjectKindTemplate) {
		t.Errorf("resolver subject_kind = %q; want %q", resolver.gotSubjectKind, domain.SubjectKindTemplate)
	}
	if resolver.gotSubjectKey != "po" {
		t.Errorf("resolver subject_key = %q; want the template's doc_type_code po, not the template id", resolver.gotSubjectKey)
	}
}

// assertFixturePreview asserts a RoutePreview matches previewRouteFixture —
// the shared assertion proving BOTH delivery paths map route → stages →
// selectors identically (resolution parity at the mapping layer; the DB-backed
// parity guard lives in tests/integration/approval).
func assertFixturePreview(t *testing.T, preview RoutePreview) {
	t.Helper()
	if preview.RouteID != "route-uuid-1" {
		t.Errorf("RouteID = %q; want route-uuid-1", preview.RouteID)
	}
	if preview.RouteName != "ISO 9001 Two-Stage" {
		t.Errorf("RouteName = %q; want ISO 9001 Two-Stage", preview.RouteName)
	}
	if len(preview.Stages) != 2 {
		t.Fatalf("Stages len = %d; want 2", len(preview.Stages))
	}

	s1 := preview.Stages[0]
	if s1.StageOrder != 1 || s1.Label != "QA Review" {
		t.Errorf("stage 1 = {order %d, label %q}; want {1, QA Review}", s1.StageOrder, s1.Label)
	}
	if len(s1.Selectors) != 1 || s1.Selectors[0].Kind != domain.SelectorRoleInFixedArea {
		t.Errorf("stage 1 selectors = %+v; want one role_in_fixed_area", s1.Selectors)
	}

	s2 := preview.Stages[1]
	if s2.StageOrder != 2 || s2.Label != "Manager Sign-off" {
		t.Errorf("stage 2 = {order %d, label %q}; want {2, Manager Sign-off}", s2.StageOrder, s2.Label)
	}
	if len(s2.Selectors) != 1 {
		t.Fatalf("stage 2 selectors len = %d; want 1", len(s2.Selectors))
	}
	sc := s2.Selectors[0]
	if sc.Kind != domain.SelectorSubmitChoice {
		t.Errorf("stage 2 selector kind = %q; want submit_choice", sc.Kind)
	}
	if sc.Role != "manager" || sc.AreaCode != "OPS" {
		t.Errorf("submit_choice selector = {role %q, area %q}; want {manager, OPS} (role+area_code must survive for the picker)", sc.Role, sc.AreaCode)
	}
}

// Compile-time proof the docsdomain sentinel PreviewRoute collapses is the one
// resolveActiveRoute actually returns (keeps the no-route test honest if the
// sentinel is ever renamed).
var _ = docsdomain.ErrProfileNotConfigured
