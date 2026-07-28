package application

import (
	"context"
	"errors"
	"testing"
	"time"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

// newGateService builds a Create-capable service whose only variable is the
// route-readiness reader, so each test below isolates the D2 gate.
func newGateService(t *testing.T, reader *fakeRouteReadinessReader) (*ControlledDocumentService, *fakeSequenceAllocator, *fakeControlledDocumentRepository) {
	t.Helper()
	repo := newFakeControlledDocumentRepository()
	seq := &fakeSequenceAllocator{next: 1}
	svc := NewControlledDocumentService(
		newTxRunner(newPermissiveMockDB(t)),
		repo,
		seq,
		&fakeTemplateVersionChecker{},
		&fakeProfileReader{},
		&fakeAreaReader{},
		&fakeGovernanceLogger{},
		newInvariantReadyDocumentInitializer(),
	)
	if reader != nil {
		svc.WithRouteReadinessReader(reader)
	}
	svc.now = func() time.Time { return time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC) }
	return svc, seq, repo
}

func autoCreateCmd() CreateControlledDocumentCmd {
	return CreateControlledDocumentCmd{
		TenantID:        "tenant-a",
		ProfileCode:     "po",
		ProcessAreaCode: "quality",
		Title:           "Welding Procedure",
		DocumentName:    "Welding Procedure v1",
		OwnerUserID:     "owner-1",
		ActorUserID:     "actor-1",
		VisibilityScope: "restricted",
	}
}

func manualCreateCmd() CreateControlledDocumentCmd {
	cmd := autoCreateCmd()
	cmd.ManualCode = stringPtr("PO-LEG-47")
	cmd.ManualCodeReason = stringPtr("Legacy migration from spreadsheet")
	cmd.VisibilityScope = "company"
	return cmd
}

// TestCreate_AutoCode_RejectedWhenProfileHasNoActiveRoute is the D2 gate on the
// auto-numbered path: a profile without an active approval route cannot receive
// a new controlled document. No fallback route, no soft warning.
func TestCreate_AutoCode_RejectedWhenProfileHasNoActiveRoute(t *testing.T) {
	reader := &fakeRouteReadinessReader{ready: false}
	svc, _, repo := newGateService(t, reader)

	_, err := svc.Create(authzCtx("tenant-a", "actor-1"), autoCreateCmd())
	if !errors.Is(err, controlleddocumentsdomain.ErrApprovalRouteMissing) {
		t.Fatalf("expected ErrApprovalRouteMissing, got %v", err)
	}
	if reader.calls != 1 {
		t.Fatalf("expected exactly 1 gate read, got %d", reader.calls)
	}
	if repo.created != nil {
		t.Fatalf("gate must run before insert; repo persisted %+v", repo.created)
	}
}

// TestCreate_AutoCode_GateRunsBeforeSequenceIsConsumed pins the gate's position
// in the auto branch: a route-less profile must burn no CD number.
func TestCreate_AutoCode_GateRunsBeforeSequenceIsConsumed(t *testing.T) {
	svc, seq, _ := newGateService(t, &fakeRouteReadinessReader{ready: false})

	_, err := svc.Create(authzCtx("tenant-a", "actor-1"), autoCreateCmd())
	if !errors.Is(err, controlleddocumentsdomain.ErrApprovalRouteMissing) {
		t.Fatalf("expected ErrApprovalRouteMissing, got %v", err)
	}
	if seq.next != 1 {
		t.Fatalf("sequence must not be consumed when the gate rejects; next=%d want 1", seq.next)
	}
}

// TestCreate_ManualCode_RejectedWhenProfileHasNoActiveRoute is the D2 gate on
// the manual-code path — symmetric with the auto branch, so an override code
// cannot be used to route around approval readiness.
func TestCreate_ManualCode_RejectedWhenProfileHasNoActiveRoute(t *testing.T) {
	reader := &fakeRouteReadinessReader{ready: false}
	svc, _, repo := newGateService(t, reader)

	_, err := svc.Create(authzCtx("tenant-a", "actor-1"), manualCreateCmd())
	if !errors.Is(err, controlleddocumentsdomain.ErrApprovalRouteMissing) {
		t.Fatalf("expected ErrApprovalRouteMissing, got %v", err)
	}
	if reader.calls != 1 {
		t.Fatalf("expected exactly 1 gate read, got %d", reader.calls)
	}
	if repo.created != nil {
		t.Fatalf("gate must run before insert; repo persisted %+v", repo.created)
	}
}

// TestCreate_GateAddressesDocumentSubjectByProfileCode pins the subject
// addressing the gate shares with the submit path: subject_kind='document',
// subject_key=<profile code>. A drift here would silently make the gate
// unreachable (always-false or always-true).
func TestCreate_GateAddressesDocumentSubjectByProfileCode(t *testing.T) {
	reader := routeReady()
	svc, _, _ := newGateService(t, reader)

	if _, err := svc.Create(authzCtx("tenant-a", "actor-1"), autoCreateCmd()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reader.gotTenantID != "tenant-a" {
		t.Errorf("gate tenant = %q, want tenant-a", reader.gotTenantID)
	}
	if reader.gotSubjectKind != string(approvaldomain.SubjectKindDocument) {
		t.Errorf("gate subject_kind = %q, want %q", reader.gotSubjectKind, approvaldomain.SubjectKindDocument)
	}
	if reader.gotSubjectKey != "po" {
		t.Errorf("gate subject_key = %q, want the profile code po", reader.gotSubjectKey)
	}
	if reader.gotExec == nil {
		t.Error("gate must run on the caller's tx executor, got nil")
	}
}

// TestCreate_FailsClosedWhenRouteReaderUnwired proves the gate is fail-closed:
// a composition-root wiring miss rejects creation rather than allowing it.
func TestCreate_FailsClosedWhenRouteReaderUnwired(t *testing.T) {
	svc, _, _ := newGateService(t, nil)

	_, err := svc.Create(authzCtx("tenant-a", "actor-1"), autoCreateCmd())
	if !errors.Is(err, controlleddocumentsdomain.ErrApprovalRouteMissing) {
		t.Fatalf("expected ErrApprovalRouteMissing with an unwired reader, got %v", err)
	}
}

// TestCreate_RouteReadErrorIsNotMaskedAsRouteMissing keeps an infrastructure
// failure distinguishable from a genuine "no route": a read error must surface
// as a 500-class error, never as the 409 the gate emits.
func TestCreate_RouteReadErrorIsNotMaskedAsRouteMissing(t *testing.T) {
	boom := errors.New("connection reset")
	svc, _, _ := newGateService(t, &fakeRouteReadinessReader{err: boom})

	_, err := svc.Create(authzCtx("tenant-a", "actor-1"), autoCreateCmd())
	if !errors.Is(err, boom) {
		t.Fatalf("expected the underlying read error to propagate, got %v", err)
	}
	if errors.Is(err, controlleddocumentsdomain.ErrApprovalRouteMissing) {
		t.Fatal("a route-read failure must not be reported as ErrApprovalRouteMissing (409)")
	}
}

// TestCreate_SucceedsWhenProfileHasActiveRoute is the positive control for the
// three tests above: with a route present the gate is transparent.
func TestCreate_SucceedsWhenProfileHasActiveRoute(t *testing.T) {
	svc, _, repo := newGateService(t, routeReady())

	res, err := svc.Create(authzCtx("tenant-a", "actor-1"), autoCreateCmd())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || res.ControlledDocument == nil {
		t.Fatal("expected a controlled document")
	}
	if repo.created == nil {
		t.Fatal("expected the controlled document to be persisted")
	}
}

// ---------------------------------------------------------------------------
// CreationContext (GET /controlled-documents/creation-context)
// ---------------------------------------------------------------------------

type fakeProfileLister struct {
	items []taxonomydomain.DocumentProfile
	err   error
}

func (f *fakeProfileLister) ListProfiles(_ context.Context, _ string) ([]taxonomydomain.DocumentProfile, error) {
	return f.items, f.err
}

type fakeAreaLister struct {
	items []taxonomydomain.ProcessArea
	err   error
}

func (f *fakeAreaLister) ListAreas(_ context.Context, _ string) ([]taxonomydomain.ProcessArea, error) {
	return f.items, f.err
}

type fakeAreaCapabilityReader struct {
	tenantWide bool
	areas      []string
	err        error

	gotCapability string
	gotActor      string
}

func (f *fakeAreaCapabilityReader) AreasWithCapability(_ context.Context, _, actorUserID, capability string, _ time.Time) (bool, []string, error) {
	f.gotActor = actorUserID
	f.gotCapability = capability
	if f.err != nil {
		return false, nil, f.err
	}
	return f.tenantWide, f.areas, nil
}

func actorCtx(userID string) context.Context {
	return iamdomain.WithAuthContext(context.Background(), userID, nil)
}

func archivedAt() *time.Time {
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return &ts
}

// newCreationContextService wires only the four creation-context readers; the
// Create-path collaborators are irrelevant here.
func newCreationContextService(
	t *testing.T,
	profiles *fakeProfileLister,
	areas *fakeAreaLister,
	caps *fakeAreaCapabilityReader,
	routes *fakeRouteReadinessReader,
) *ControlledDocumentService {
	t.Helper()
	svc := NewControlledDocumentService(
		newTxRunner(newPermissiveMockDB(t)),
		newFakeControlledDocumentRepository(),
		&fakeSequenceAllocator{next: 1},
		&fakeTemplateVersionChecker{},
		&fakeProfileReader{},
		&fakeAreaReader{},
		&fakeGovernanceLogger{},
		newInvariantReadyDocumentInitializer(),
	)
	if routes != nil {
		svc.WithRouteReadinessReader(routes)
	}
	if profiles != nil && areas != nil && caps != nil {
		svc.WithCreationContextReaders(profiles, areas, caps)
	}
	return svc
}

// TestCreationContext_NarrowsAreasToCallerGrants is deliverable C's core
// invariant: the response carries ONLY areas in which the caller holds
// controlled_documents.create — the client never receives the raw catalog.
func TestCreationContext_NarrowsAreasToCallerGrants(t *testing.T) {
	caps := &fakeAreaCapabilityReader{areas: []string{"quality"}}
	svc := newCreationContextService(t,
		&fakeProfileLister{items: []taxonomydomain.DocumentProfile{{Code: "po", Name: "Procedimento"}}},
		&fakeAreaLister{items: []taxonomydomain.ProcessArea{
			{Code: "quality", Name: "Qualidade"},
			{Code: "finance", Name: "Financeiro"},
		}},
		caps,
		&fakeRouteReadinessReader{readySet: map[string]struct{}{"po": {}}},
	)

	got, err := svc.CreationContext(actorCtx("actor-1"), "tenant-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Areas) != 1 || got.Areas[0].Code != "quality" || got.Areas[0].Name != "Qualidade" {
		t.Fatalf("areas = %+v, want only [quality]", got.Areas)
	}
	if caps.gotCapability != string(iamdomain.CapControlledDocumentCreate) {
		t.Errorf("capability = %q, want %q", caps.gotCapability, iamdomain.CapControlledDocumentCreate)
	}
	if caps.gotActor != "actor-1" {
		t.Errorf("actor = %q, want actor-1", caps.gotActor)
	}
}

// TestCreationContext_TenantWideActorSeesWholeActiveCatalog covers the
// system_admin short-circuit: such an actor bypasses tier-2, so narrowing to
// explicit per-area grants would under-report what they may actually do.
func TestCreationContext_TenantWideActorSeesWholeActiveCatalog(t *testing.T) {
	svc := newCreationContextService(t,
		&fakeProfileLister{},
		&fakeAreaLister{items: []taxonomydomain.ProcessArea{
			{Code: "quality", Name: "Qualidade"},
			{Code: "finance", Name: "Financeiro"},
			{Code: "legacy", Name: "Legado", ArchivedAt: archivedAt()},
		}},
		&fakeAreaCapabilityReader{tenantWide: true},
		&fakeRouteReadinessReader{},
	)

	got, err := svc.CreationContext(actorCtx("admin-1"), "tenant-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Areas) != 2 {
		t.Fatalf("areas = %+v, want the 2 active catalog entries", got.Areas)
	}
	for _, a := range got.Areas {
		if a.Code == "legacy" {
			t.Error("archived areas must not be offered")
		}
	}
}

// TestCreationContext_AnnotatesProfileRouteReadiness proves the profile list
// carries has_active_route so the UI can disable a profile that would 409.
func TestCreationContext_AnnotatesProfileRouteReadiness(t *testing.T) {
	svc := newCreationContextService(t,
		&fakeProfileLister{items: []taxonomydomain.DocumentProfile{
			{Code: "po", Name: "Procedimento"},
			{Code: "it", Name: "Instrucao"},
			{Code: "old", Name: "Obsoleto", ArchivedAt: archivedAt()},
		}},
		&fakeAreaLister{},
		&fakeAreaCapabilityReader{},
		&fakeRouteReadinessReader{readySet: map[string]struct{}{"po": {}}},
	)

	got, err := svc.CreationContext(actorCtx("actor-1"), "tenant-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Profiles) != 2 {
		t.Fatalf("profiles = %+v, want the 2 active ones", got.Profiles)
	}
	byCode := map[string]bool{}
	for _, p := range got.Profiles {
		byCode[p.Code] = p.HasActiveRoute
	}
	if !byCode["po"] {
		t.Error("po has an active route, want has_active_route=true")
	}
	if byCode["it"] {
		t.Error("it has no active route, want has_active_route=false")
	}
	if _, present := byCode["old"]; present {
		t.Error("archived profiles must not be offered")
	}
}

// TestCreationContext_RequiresActor fails closed when no authenticated
// principal is present — the response is a per-caller authorization view and
// has no meaning without one.
func TestCreationContext_RequiresActor(t *testing.T) {
	svc := newCreationContextService(t, &fakeProfileLister{}, &fakeAreaLister{}, &fakeAreaCapabilityReader{}, &fakeRouteReadinessReader{})

	if _, err := svc.CreationContext(context.Background(), "tenant-a"); !errors.Is(err, ErrActorMissing) {
		t.Fatalf("expected ErrActorMissing, got %v", err)
	}
}

// TestCreationContext_FailsClosedWhenUnwired refuses to return a partially
// computed (and therefore misleading) authorization view.
func TestCreationContext_FailsClosedWhenUnwired(t *testing.T) {
	svc := newCreationContextService(t, nil, nil, nil, nil)

	if _, err := svc.CreationContext(actorCtx("actor-1"), "tenant-a"); !errors.Is(err, ErrCreationContextUnconfigured) {
		t.Fatalf("expected ErrCreationContextUnconfigured, got %v", err)
	}
}

// TestCreationContext_PropagatesCapabilityReadError keeps an iam read failure
// loud instead of degrading to an empty (silently over-restrictive) area list.
func TestCreationContext_PropagatesCapabilityReadError(t *testing.T) {
	boom := errors.New("iam unavailable")
	svc := newCreationContextService(t,
		&fakeProfileLister{},
		&fakeAreaLister{items: []taxonomydomain.ProcessArea{{Code: "quality", Name: "Qualidade"}}},
		&fakeAreaCapabilityReader{err: boom},
		&fakeRouteReadinessReader{},
	)

	if _, err := svc.CreationContext(actorCtx("actor-1"), "tenant-a"); !errors.Is(err, boom) {
		t.Fatalf("expected the iam read error to propagate, got %v", err)
	}
}
