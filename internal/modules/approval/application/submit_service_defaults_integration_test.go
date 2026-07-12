//go:build integration
// +build integration

package application

// ADR 0073 — canonical-submit in-tx RESOLUTION real-DB proof.
//
// The unit tests (fake driver, nil resolver) and the existing
// submit_service_reason_integration_test.go (explicit route_id, no resolver)
// never exercise the resolution path. These tests wire the concrete Postgres
// repository as SubmitDefaultsResolver + the concrete CDFieldReaderPG cdRead —
// exactly as services.go NewServices does in production — and drive
// SubmitService.SubmitRevisionForReview against a real Postgres testdb, proving:
//
//  1. TestSubmitResolvesActiveRoute_RealDB — RouteID:"" → resolveActiveRoute
//     picks the single active route for the document's profile in-tx; the
//     created approval_instances.route_id equals the seeded route.
//  2. TestSubmitBindsHeadContentHash_RealDB — ContentFormData with a
//     present-but-empty "_content_hash" → resolveContentFormData binds the head
//     autosaved revision's stored content_hash (LoadHeadContentHash); the
//     persisted content_hash_at_submit equals the head hash (NOT the empty-form
//     hash).
//  3. TestSubmitRev0NoGovernanceData_RealDB — fresh REV0 draft, no route_id, no
//     reason_for_change, empty "_content_hash" → succeeds; document goes
//     draft->under_review; revision_version bumped by 1.
//  4. TestSubmitRev1RequiresReason_RealDB — REV1 draft with reason succeeds;
//     REV1 draft without reason returns application.ErrReasonForChangeRequired.
//  5. TestSubmitExplicitRouteAndHash_RealDB — explicit RouteID + a real non-empty
//     "_content_hash" → resolution NOT triggered; succeeds and persists the
//     supplied hash's computed value.
//  6. TestSubmitReplayDuplicate_RealDB — same IdempotencyKey twice → second call
//     returns infrastructure.ErrDuplicateSubmission (UNIQUE(document_id,
//     idempotency_key) backstop).
//
// Resolver + cdRead are wired via the SAME assertions services.go uses, so this
// is a faithful production-composition proof, not a fixture-only shortcut.
// system_admin bypasses the document.submit/document.edit capability checks
// (ADR 0022 Phase 11 F8); the tenant/actor GUCs are seeded via submitCtxWithIdentity.

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	cdinfra "metaldocs/internal/modules/controlleddocuments/infrastructure"
	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// newSubmitServiceWithResolver builds a SubmitService wired exactly as
// application.NewServices wires it in production: the concrete Postgres
// repository doubles as the SubmitDefaultsResolver (in-tx route/hash
// resolution, ADR 0073) via interface assertion, and the concrete
// CDFieldReaderPG serves the cross-module profile_code read. Setting the
// unexported fields directly is legal here because the test is in package
// application.
func newSubmitServiceWithResolver(t *testing.T, dbc *sql.DB) *SubmitService {
	t.Helper()
	repo := approvalrepo.NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	resolver, ok := interface{}(repo).(SubmitDefaultsResolver)
	if !ok {
		t.Fatalf("postgres approval repository does not satisfy SubmitDefaultsResolver; production wiring in services.go would silently disable resolution")
	}
	return &SubmitService{
		repo:     repo,
		emitter:  NewSQLEmitter(),
		clock:    RealClock{},
		cdRead:   cdinfra.NewCDFieldReaderPG(),
		resolver: resolver,
	}
}

// profileCodeForDoc reads the controlled-document profile_code the submit
// route must target so the in-tx resolveActiveRoute finds it.
func profileCodeForDoc(t *testing.T, dbc *sql.DB, controlledDocumentID string) string {
	t.Helper()
	var profileCode string
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT profile_code FROM public.controlled_documents WHERE id = $1::uuid`,
		controlledDocumentID,
	).Scan(&profileCode); err != nil {
		t.Fatalf("lookup profile_code: %v", err)
	}
	return profileCode
}

// seedHeadRevision inserts one editor_sessions row + one document_revisions row
// carrying an explicit content_hash, so LoadHeadContentHash
// (postgres_approval_repository.go ~line 1250 — ORDER BY created_at DESC, id DESC
// LIMIT 1, joined to documents for tenant scope) returns exactly that hash. No
// factory helper seeds a head revision with a chosen content hash, and there is
// no tripwire on document_revisions/editor_sessions, so a raw INSERT is the
// correct seam (mirrors fixtures.go InsertDraftDocument's revision seeding).
func seedHeadRevision(t *testing.T, dbc *sql.DB, tenantID, documentID, ownerID, contentHash string) {
	t.Helper()
	ctx := context.Background()
	var sessID string
	if err := dbc.QueryRowContext(ctx,
		`INSERT INTO public.editor_sessions
		   (tenant_id, document_id, user_id, expires_at, last_acknowledged_revision_id, status)
		 VALUES ($1::uuid, $2::uuid, $3, now() + interval '1 hour',
		         '00000000-0000-0000-0000-000000000000', 'active')
		 RETURNING id::text`,
		tenantID, documentID, ownerID,
	).Scan(&sessID); err != nil {
		t.Fatalf("seed editor_sessions: %v", err)
	}
	if _, err := dbc.ExecContext(ctx,
		`INSERT INTO public.document_revisions
		   (document_id, parent_revision_id, session_id, storage_key, content_hash, form_data_snapshot)
		 VALUES ($1::uuid, NULL, $2::uuid, 'tenants/test/head.docx', $3, '{}')`,
		documentID, sessID, contentHash,
	); err != nil {
		t.Fatalf("seed document_revisions: %v", err)
	}
}

// instanceRouteAndHash reads back the persisted route_id + content_hash_at_submit
// for the created approval instance.
func instanceRouteAndHash(t *testing.T, dbc *sql.DB, instanceID string) (routeID, contentHash string) {
	t.Helper()
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT route_id::text, content_hash_at_submit
		   FROM public.approval_instances WHERE id = $1::uuid`,
		instanceID,
	).Scan(&routeID, &contentHash); err != nil {
		t.Fatalf("read back approval_instances row: %v", err)
	}
	return routeID, contentHash
}

// TestSubmitResolvesActiveRoute_RealDB — case 1: no-route resolution. RouteID:""
// forces resolveActiveRoute to pick the single active route for the document's
// profile in-tx; the created instance's route_id must equal the seeded route.
func TestSubmitResolvesActiveRoute_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	doc := testdb.NewDocument(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithOwner(actor.ID), testdb.WithStatus("draft"), testdb.WithRevisionNumber(0), testdb.WithSubmitReadySnapshots())

	profileCode := profileCodeForDoc(t, dbc, doc.ControlledDocumentID)
	wantRoute := seedSubmitRouteWithStage(t, dbc, tnt.ID, profileCode)

	svc := newSubmitServiceWithResolver(t, dbc)
	runner := db.NewTxRunner(dbc)

	req := SubmitRequest{
		TenantID:        tnt.ID,
		DocumentID:      doc.ID,
		RouteID:         "", // omitted → in-tx resolution
		SubmittedBy:     actor.ID,
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 0,
		IdempotencyKey:  "aaaa1111-1111-1111-1111-111111111111",
	}

	res, err := svc.SubmitRevisionForReview(submitCtxWithIdentity(tnt.ID, actor.ID), runner, req)
	if err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}
	if res.InstanceID == "" {
		t.Fatalf("InstanceID empty; want non-empty on success")
	}

	gotRoute, _ := instanceRouteAndHash(t, dbc, res.InstanceID)
	if gotRoute != wantRoute {
		t.Fatalf("approval_instances.route_id = %q, want resolved active route %q", gotRoute, wantRoute)
	}
}

// TestSubmitBindsHeadContentHash_RealDB — case 2: head-hash resolution. A
// present-but-empty "_content_hash" forces resolveContentFormData to bind the
// head autosaved revision's stored hash. The persisted content_hash_at_submit
// must equal the ComputeContentHash of the head hash — and must NOT equal the
// hash computed with an empty "_content_hash".
func TestSubmitBindsHeadContentHash_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	doc := testdb.NewDocument(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithOwner(actor.ID), testdb.WithStatus("draft"), testdb.WithRevisionNumber(0), testdb.WithSubmitReadySnapshots())

	profileCode := profileCodeForDoc(t, dbc, doc.ControlledDocumentID)
	routeID := seedSubmitRouteWithStage(t, dbc, tnt.ID, profileCode)

	const headHash = "head-content-hash-abc123"
	seedHeadRevision(t, dbc, tnt.ID, doc.ID, actor.ID, headHash)

	svc := newSubmitServiceWithResolver(t, dbc)
	runner := db.NewTxRunner(dbc)

	req := SubmitRequest{
		TenantID:   tnt.ID,
		DocumentID: doc.ID,
		RouteID:    routeID,
		SubmittedBy: actor.ID,
		// Present-but-empty → resolveContentFormData substitutes the head hash.
		ContentFormData: map[string]any{"_content_hash": ""},
		RevisionVersion: 0,
		IdempotencyKey:  "bbbb2222-2222-2222-2222-222222222222",
	}

	res, err := svc.SubmitRevisionForReview(submitCtxWithIdentity(tnt.ID, actor.ID), runner, req)
	if err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}

	// Recompute the two candidate hashes the same way the service does. The
	// service derives revision_number in-tx (T8b) — a REV0 draft → 0.
	wantBound, err := ComputeContentHash(ContentHashInput{
		TenantID: tnt.ID, DocumentID: doc.ID, RevisionNumber: 0,
		FormData: map[string]any{"_content_hash": headHash},
	})
	if err != nil {
		t.Fatalf("compute bound hash: %v", err)
	}
	emptyForm, err := ComputeContentHash(ContentHashInput{
		TenantID: tnt.ID, DocumentID: doc.ID, RevisionNumber: 0,
		FormData: map[string]any{"_content_hash": ""},
	})
	if err != nil {
		t.Fatalf("compute empty-form hash: %v", err)
	}
	if wantBound == emptyForm {
		t.Fatalf("test setup invalid: bound and empty-form hashes collide (%q)", wantBound)
	}

	_, gotHash := instanceRouteAndHash(t, dbc, res.InstanceID)
	if gotHash != wantBound {
		t.Fatalf("content_hash_at_submit = %q, want head-bound hash %q (NOT empty-form %q)", gotHash, wantBound, emptyForm)
	}
}

// TestSubmitRev0NoGovernanceData_RealDB — case 3: fresh REV0 draft, no route_id,
// no reason_for_change, empty "_content_hash" → succeeds; document transitions
// draft->under_review; revision_version bumped by 1.
func TestSubmitRev0NoGovernanceData_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	doc := testdb.NewDocument(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithOwner(actor.ID), testdb.WithStatus("draft"), testdb.WithRevisionNumber(0), testdb.WithRevisionVersion(0), testdb.WithSubmitReadySnapshots())

	profileCode := profileCodeForDoc(t, dbc, doc.ControlledDocumentID)
	seedSubmitRouteWithStage(t, dbc, tnt.ID, profileCode)
	// A head revision so the empty "_content_hash" resolves to a real value.
	seedHeadRevision(t, dbc, tnt.ID, doc.ID, actor.ID, "rev0-head-hash")

	svc := newSubmitServiceWithResolver(t, dbc)
	runner := db.NewTxRunner(dbc)

	req := SubmitRequest{
		TenantID:        tnt.ID,
		DocumentID:      doc.ID,
		RouteID:         "",
		SubmittedBy:     actor.ID,
		ContentFormData: map[string]any{"_content_hash": ""},
		RevisionVersion: 0,
		IdempotencyKey:  "cccc3333-3333-3333-3333-333333333333",
	}

	res, err := svc.SubmitRevisionForReview(submitCtxWithIdentity(tnt.ID, actor.ID), runner, req)
	if err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}
	if res.InstanceID == "" {
		t.Fatalf("InstanceID empty; want non-empty")
	}

	var gotStatus string
	var gotRevVersion int
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT status, revision_version FROM public.documents WHERE id = $1::uuid`, doc.ID,
	).Scan(&gotStatus, &gotRevVersion); err != nil {
		t.Fatalf("read back documents row: %v", err)
	}
	if gotStatus != "under_review" {
		t.Fatalf("document status = %q, want under_review", gotStatus)
	}
	if gotRevVersion != 1 {
		t.Fatalf("revision_version = %d, want 1 (bumped from 0)", gotRevVersion)
	}
}

// TestSubmitRev1RequiresReason_RealDB — case 4: REV1 draft with reason succeeds;
// REV1 draft without reason returns application.ErrReasonForChangeRequired.
func TestSubmitRev1RequiresReason_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)

	// Sub-case A: REV1 WITH reason → succeeds.
	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	doc := testdb.NewDocument(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithOwner(actor.ID), testdb.WithStatus("draft"), testdb.WithRevisionNumber(1), testdb.WithSubmitReadySnapshots())
	profileCode := profileCodeForDoc(t, dbc, doc.ControlledDocumentID)
	seedSubmitRouteWithStage(t, dbc, tnt.ID, profileCode)
	seedHeadRevision(t, dbc, tnt.ID, doc.ID, actor.ID, "rev1-head-hash")

	svc := newSubmitServiceWithResolver(t, dbc)
	runner := db.NewTxRunner(dbc)

	okReq := SubmitRequest{
		TenantID:        tnt.ID,
		DocumentID:      doc.ID,
		RouteID:         "",
		SubmittedBy:     actor.ID,
		RevisionTitle:   "Revisao corretiva",
		ReasonForChange: "Corrected tolerance per complaint #900",
		ReasonCategory:  "corrective",
		ContentFormData: map[string]any{"_content_hash": ""},
		RevisionVersion: 0,
		IdempotencyKey:  "dddd4444-4444-4444-4444-444444444444",
	}
	if _, err := svc.SubmitRevisionForReview(submitCtxWithIdentity(tnt.ID, actor.ID), runner, okReq); err != nil {
		t.Fatalf("REV1 with reason: unexpected error: %v", err)
	}

	// Sub-case B: a distinct REV1 draft WITHOUT reason → ErrReasonForChangeRequired.
	tnt2 := testdb.NewTenant(t, dbc)
	actor2 := testdb.NewUser(t, dbc, testdb.WithTenant(tnt2.ID), testdb.WithRole("system_admin"))
	doc2 := testdb.NewDocument(t, dbc, testdb.WithTenant(tnt2.ID), testdb.WithOwner(actor2.ID), testdb.WithStatus("draft"), testdb.WithRevisionNumber(1), testdb.WithSubmitReadySnapshots())
	profileCode2 := profileCodeForDoc(t, dbc, doc2.ControlledDocumentID)
	seedSubmitRouteWithStage(t, dbc, tnt2.ID, profileCode2)
	seedHeadRevision(t, dbc, tnt2.ID, doc2.ID, actor2.ID, "rev1-head-hash-2")

	badReq := SubmitRequest{
		TenantID:        tnt2.ID,
		DocumentID:      doc2.ID,
		RouteID:         "",
		SubmittedBy:     actor2.ID,
		RevisionTitle:   "Revisao sem motivo",
		ContentFormData: map[string]any{"_content_hash": ""},
		RevisionVersion: 0,
		IdempotencyKey:  "eeee5555-5555-5555-5555-555555555555",
	}
	_, err := svc.SubmitRevisionForReview(submitCtxWithIdentity(tnt2.ID, actor2.ID), runner, badReq)
	if !errors.Is(err, ErrReasonForChangeRequired) {
		t.Fatalf("REV1 without reason: err = %v, want ErrReasonForChangeRequired", err)
	}
}

// TestSubmitExplicitRouteAndHash_RealDB — case 5: explicit RouteID + a real
// non-empty "_content_hash" → resolution NOT triggered; succeeds and the
// persisted content_hash_at_submit is the ComputeContentHash of the supplied
// (non-empty) form data, unchanged by any head-hash binding.
func TestSubmitExplicitRouteAndHash_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	doc := testdb.NewDocument(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithOwner(actor.ID), testdb.WithStatus("draft"), testdb.WithRevisionNumber(0), testdb.WithSubmitReadySnapshots())

	profileCode := profileCodeForDoc(t, dbc, doc.ControlledDocumentID)
	routeID := seedSubmitRouteWithStage(t, dbc, tnt.ID, profileCode)
	// Seed a DIFFERENT head hash to prove it is ignored when an explicit
	// non-empty "_content_hash" is supplied.
	seedHeadRevision(t, dbc, tnt.ID, doc.ID, actor.ID, "should-be-ignored-head-hash")

	svc := newSubmitServiceWithResolver(t, dbc)
	runner := db.NewTxRunner(dbc)

	const explicitHash = "explicit-client-hash-xyz789"
	req := SubmitRequest{
		TenantID:        tnt.ID,
		DocumentID:      doc.ID,
		RouteID:         routeID,
		SubmittedBy:     actor.ID,
		ContentFormData: map[string]any{"_content_hash": explicitHash},
		RevisionVersion: 0,
		IdempotencyKey:  "ffff6666-6666-6666-6666-666666666666",
	}

	res, err := svc.SubmitRevisionForReview(submitCtxWithIdentity(tnt.ID, actor.ID), runner, req)
	if err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}

	wantHash, err := ComputeContentHash(ContentHashInput{
		TenantID: tnt.ID, DocumentID: doc.ID, RevisionNumber: 0,
		FormData: map[string]any{"_content_hash": explicitHash},
	})
	if err != nil {
		t.Fatalf("compute explicit hash: %v", err)
	}

	gotRoute, gotHash := instanceRouteAndHash(t, dbc, res.InstanceID)
	if gotRoute != routeID {
		t.Fatalf("approval_instances.route_id = %q, want explicit %q", gotRoute, routeID)
	}
	if gotHash != wantHash {
		t.Fatalf("content_hash_at_submit = %q, want explicit-form hash %q (resolution must NOT rebind)", gotHash, wantHash)
	}
}

// TestSubmitReplayDuplicate_RealDB — case 6: the same IdempotencyKey submitted
// twice on the same document. The first succeeds; the second returns
// infrastructure.ErrDuplicateSubmission (UNIQUE(document_id, idempotency_key)
// backstop), assertable via errors.Is.
func TestSubmitReplayDuplicate_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	doc := testdb.NewDocument(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithOwner(actor.ID), testdb.WithStatus("draft"), testdb.WithRevisionNumber(0), testdb.WithSubmitReadySnapshots())

	profileCode := profileCodeForDoc(t, dbc, doc.ControlledDocumentID)
	routeID := seedSubmitRouteWithStage(t, dbc, tnt.ID, profileCode)

	svc := newSubmitServiceWithResolver(t, dbc)
	runner := db.NewTxRunner(dbc)

	req := SubmitRequest{
		TenantID:        tnt.ID,
		DocumentID:      doc.ID,
		RouteID:         routeID,
		SubmittedBy:     actor.ID,
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 0,
		IdempotencyKey:  "77777777-7777-7777-7777-777777777777",
	}

	ctx := submitCtxWithIdentity(tnt.ID, actor.ID)
	if _, err := svc.SubmitRevisionForReview(ctx, runner, req); err != nil {
		t.Fatalf("first submit: unexpected error: %v", err)
	}
	_, err := svc.SubmitRevisionForReview(ctx, runner, req)
	if !errors.Is(err, approvalrepo.ErrDuplicateSubmission) {
		t.Fatalf("replay submit: err = %v, want infrastructure.ErrDuplicateSubmission", err)
	}
}
