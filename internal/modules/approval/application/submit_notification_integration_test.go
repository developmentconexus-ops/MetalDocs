//go:build integration
// +build integration

package application

// Task 6 — real-DB proof that submit now tells its approvers. Before this,
// submit emitted only a GovernanceEvent (an AUDIT record); the notification
// path was a separate call made only from decision sites, so a real
// submission produced zero notification rows and zero jobs. These tests pin
// that SubmitService.SubmitRevisionForReview and
// TemplateSubmitService.SubmitTemplateVersionForReview both now enqueue
// approval.pending to the ACTIVE stage's eligible actors, inside the same
// submit transaction, and that a livre (ADR 0087) zero-stage route enqueues
// nothing — the degenerate case falls out of an empty recipient list, no
// special branch.
//
// package application (internal), not application_test: this file reuses the
// package's existing unexported test fixtures (fakeTemplateVersionReader,
// fakeTemplateVersionSubmitWriter, seedTemplateVersion, seedTemplateRoute,
// seedZeroStageTemplateRoute, newTemplateSubmitCtx, grantAreaBlindRole,
// seedSubmitRouteWithStage, submitCtxWithIdentity) defined in the neighbouring
// *_integration_test.go files in this same package, rather than inventing a
// parallel external-test-package scaffold for a one-off notification check.

import (
	"context"
	"errors"
	"testing"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
	iampostgres "metaldocs/internal/modules/iam/infrastructure/postgres"
	templatesinfra "metaldocs/internal/modules/templates/infrastructure"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"

	"github.com/google/uuid"
)

// recordingNotifier is a test-local double for domain.ApprovalNotificationEnqueuer
// that records every call instead of enqueuing a River job, so these tests
// assert on what submit PROMISED to notify without needing a live River
// worker/queue.
type recordingNotifier struct {
	sent []approvaldomain.ApprovalNotificationArgs
}

func (r *recordingNotifier) EnqueueApprovalNotificationTx(_ context.Context, _ db.Tx, args approvaldomain.ApprovalNotificationArgs) error {
	r.sent = append(r.sent, args)
	return nil
}

// errEnqueueBroken is the sentinel a failingNotifier returns. The submit
// error returned to the caller must wrap it, so a caller can tell WHY the
// submission was refused rather than seeing an opaque failure.
var errEnqueueBroken = errors.New("river: enqueue broken (simulated)")

// failingNotifier is a test-local double for domain.ApprovalNotificationEnqueuer
// that always fails, simulating the one condition Task 6 exists to guard:
// the notification cannot be promised. Task 6's whole reason to exist is that
// a submit which cannot promise a notification must not commit — a
// half-transaction that reports success and tells nobody is exactly the
// defect this feature removes. These tests exist to prove the rollback is
// real, not merely asserted in a comment.
type failingNotifier struct{}

func (failingNotifier) EnqueueApprovalNotificationTx(_ context.Context, _ db.Tx, _ approvaldomain.ApprovalNotificationArgs) error {
	return errEnqueueBroken
}

// assertSameUsers fails the test unless got and want contain exactly the same
// user ids, order-insensitive. Not exported/shared: Task 7's tests define
// their own copy in a different package, which is correct Go (no importable
// test-only helper package exists here), not duplication.
func assertSameUsers(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("recipients = %v, want %v", got, want)
	}
	gotSet := make(map[string]int, len(got))
	for _, id := range got {
		gotSet[id]++
	}
	wantSet := make(map[string]int, len(want))
	for _, id := range want {
		wantSet[id]++
	}
	for id, n := range wantSet {
		if gotSet[id] != n {
			t.Fatalf("recipients = %v, want %v", got, want)
		}
	}
	for id, n := range gotSet {
		if wantSet[id] != n {
			t.Fatalf("recipients = %v, want %v", got, want)
		}
	}
}

// The complaint that started this: submitting a template told nobody.
func TestTemplateSubmit_EnqueuesApprovalPending_ToEligibleActors(t *testing.T) {
	dbc, _ := testdb.Open(t)
	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithGovernanceClass("controlado"))
	templateID, versionID := seedTemplateVersion(t, dbc, tnt.ID, actor.ID, tax.ProfileCode, "draft")
	seedTemplateRoute(t, dbc, tnt.ID, actor.ID, tax.ProfileCode)

	// templateRouteStages(): stage 1 (the one that opens ACTIVE at submit)
	// requires role "approver". Grant it to two users so the active stage's
	// eligible pool — and therefore the notification recipients — is exactly
	// those two. Stage 2 (pending) requires role "author" and must ALSO
	// resolve a non-empty pool: submit fails closed with
	// domain.ErrEmptyEligiblePool for any stage with zero eligible actors, not
	// just the active one (W6), independent of whether that stage has opened.
	approver1 := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	approver2 := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	pendingStageActor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	grantAreaBlindRole(t, dbc, tnt.ID, approver1.ID, "approver")
	grantAreaBlindRole(t, dbc, tnt.ID, approver2.ID, "approver")
	grantAreaBlindRole(t, dbc, tnt.ID, pendingStageActor.ID, "author")

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	notifier := &recordingNotifier{}
	svc := NewTemplateSubmitService(repo, NewSQLEmitter(), RealClock{}, fakeTemplateVersionReader{}, notifier)
	svc.versionWriter = fakeTemplateVersionSubmitWriter{}
	runner := db.NewTxRunner(dbc)

	res, err := svc.SubmitTemplateVersionForReview(newTemplateSubmitCtx(tnt.ID, actor.ID), runner, TemplateSubmitRequest{
		TenantID:          tnt.ID,
		TemplateID:        templateID,
		TemplateVersionID: versionID,
		SubmittedBy:       actor.ID,
		IdempotencyKey:    "idem-" + uuid.NewString(),
	})
	if err != nil {
		t.Fatalf("SubmitTemplateVersionForReview: %v", err)
	}
	if res.InstanceID == "" {
		t.Fatal("SubmitTemplateVersionForReview: empty InstanceID")
	}

	if len(notifier.sent) != 1 {
		t.Fatalf("enqueued %d events, want 1", len(notifier.sent))
	}
	got := notifier.sent[0]
	if got.EventType != approvaldomain.EventTypeApprovalPending {
		t.Fatalf("event type = %q, want %q", got.EventType, approvaldomain.EventTypeApprovalPending)
	}
	if got.SubjectKind != "template" {
		t.Fatalf("subject kind = %q, want template", got.SubjectKind)
	}
	if got.SubjectID != versionID {
		t.Fatalf("subject id = %q, want %q", got.SubjectID, versionID)
	}
	if got.ActorUserID != actor.ID {
		t.Fatalf("actor user id = %q, want %q", got.ActorUserID, actor.ID)
	}
	if got.EventID == "" {
		t.Fatal("event id is empty; it is the consumer's idempotency key")
	}
	assertSameUsers(t, got.RecipientUserIDs, approver1.ID, approver2.ID)
}

// This is the whole point of Task 6, proven for real: if the notification
// cannot be promised, the submission must not stick. A test that only checks
// the returned error would pass even if submit committed a half-transaction
// behind that error — so this asserts directly against the database that NO
// approval_instance was persisted and the template version never left draft.
func TestTemplateSubmit_EnqueueFailure_RollsBackSubmission(t *testing.T) {
	dbc, _ := testdb.Open(t)
	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithGovernanceClass("controlado"))
	templateID, versionID := seedTemplateVersion(t, dbc, tnt.ID, actor.ID, tax.ProfileCode, "draft")
	seedTemplateRoute(t, dbc, tnt.ID, actor.ID, tax.ProfileCode)

	approver1 := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	approver2 := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	pendingStageActor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	grantAreaBlindRole(t, dbc, tnt.ID, approver1.ID, "approver")
	grantAreaBlindRole(t, dbc, tnt.ID, approver2.ID, "approver")
	grantAreaBlindRole(t, dbc, tnt.ID, pendingStageActor.ID, "author")

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	svc := NewTemplateSubmitService(repo, NewSQLEmitter(), RealClock{}, fakeTemplateVersionReader{}, failingNotifier{})
	svc.versionWriter = fakeTemplateVersionSubmitWriter{}
	runner := db.NewTxRunner(dbc)

	_, err := svc.SubmitTemplateVersionForReview(newTemplateSubmitCtx(tnt.ID, actor.ID), runner, TemplateSubmitRequest{
		TenantID:          tnt.ID,
		TemplateID:        templateID,
		TemplateVersionID: versionID,
		SubmittedBy:       actor.ID,
		IdempotencyKey:    "idem-" + uuid.NewString(),
	})
	if err == nil {
		t.Fatal("SubmitTemplateVersionForReview: got nil error, want the enqueue failure to be surfaced — a swallowed error here means submit told nobody AND said nothing about it")
	}
	if !errors.Is(err, errEnqueueBroken) {
		t.Fatalf("SubmitTemplateVersionForReview error = %v, want it to wrap %v so a caller can tell why the submission was refused", err, errEnqueueBroken)
	}

	// The part that matters: prove the tx actually rolled back, not just that
	// an error came back. An error return with a committed row IS the
	// half-transaction this feature exists to prevent.
	var instanceCount int
	if err := dbc.QueryRow(
		`SELECT count(*) FROM approval_instances WHERE tenant_id = $1 AND subject_kind = 'template' AND subject_key = $2`,
		tnt.ID, versionID,
	).Scan(&instanceCount); err != nil {
		t.Fatalf("query approval_instances: %v", err)
	}
	if instanceCount != 0 {
		t.Fatalf("approval_instances rows for this version = %d, want 0 — an approval instance was persisted despite the enqueue failure, exactly the half-transaction Task 6 was built to prevent", instanceCount)
	}

	var versionStatus string
	if err := dbc.QueryRow(
		`SELECT status FROM templates_template_version WHERE id = $1 AND tenant_id = $2`,
		versionID, tnt.ID,
	).Scan(&versionStatus); err != nil {
		t.Fatalf("query templates_template_version: %v", err)
	}
	if versionStatus != "draft" {
		t.Fatalf("template version status = %q, want %q — the version was locked into under_review even though the submission that should have caused it failed", versionStatus, "draft")
	}
}

// A livre route (ADR 0087) has zero stages, so there are no eligible actors
// and nothing is emitted. The degenerate case needs no special branch — it
// falls out of the recipient list being empty.
func TestTemplateSubmit_LivreRoute_EnqueuesNothing(t *testing.T) {
	dbc, _ := testdb.Open(t)
	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	// Fixture default governance class is livre; its route must carry zero
	// stages (ADR 0087 / migration 0316).
	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID))
	templateID, versionID := seedTemplateVersion(t, dbc, tnt.ID, actor.ID, tax.ProfileCode, "draft")
	seedZeroStageTemplateRoute(t, dbc, tnt.ID, actor.ID, tax.ProfileCode)

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, iampostgres.NewUserDisplayNameRepository(dbc))
	notifier := &recordingNotifier{}
	services := NewServices(repo, NewSQLEmitter(), RealClock{}, nil).
		WithTemplateVersionReader(templatesinfra.NewApprovalVersionReader()).
		WithTemplateCompletionWriter(templatesinfra.NewApprovalCompletionWriter()).
		WithApprovalNotificationEnqueuer(notifier)
	runner := db.NewTxRunner(dbc)

	res, err := services.TemplateSubmit.SubmitTemplateVersionForReview(newTemplateSubmitCtx(tnt.ID, actor.ID), runner, TemplateSubmitRequest{
		TenantID:          tnt.ID,
		TemplateID:        templateID,
		TemplateVersionID: versionID,
		SubmittedBy:       actor.ID,
		IdempotencyKey:    "idem-" + uuid.NewString(),
	})
	if err != nil {
		t.Fatalf("SubmitTemplateVersionForReview: %v", err)
	}
	if res.InstanceID == "" {
		t.Fatal("SubmitTemplateVersionForReview: empty InstanceID")
	}
	if len(notifier.sent) != 0 {
		t.Fatalf("enqueued %d events, want 0 — a livre route has no one to notify", len(notifier.sent))
	}
}

// The document path must gain the same behaviour, not a different one.
func TestDocumentSubmit_EnqueuesApprovalPending(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	doc := testdb.NewDocument(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithOwner(actor.ID), testdb.WithStatus("draft"), testdb.WithRevisionNumber(0), testdb.WithSubmitReadySnapshots())

	var profileCode string
	if err := dbc.QueryRowContext(ctx,
		`SELECT profile_code FROM public.controlled_documents WHERE id = $1::uuid`, doc.ControlledDocumentID,
	).Scan(&profileCode); err != nil {
		t.Fatalf("lookup profile_code: %v", err)
	}
	routeID := seedSubmitRouteWithStage(t, dbc, tnt.ID, profileCode)

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	notifier := &recordingNotifier{}
	svc := &SubmitService{repo: repo, emitter: NewSQLEmitter(), clock: RealClock{}, notifier: notifier}
	runner := db.NewTxRunner(dbc)

	req := SubmitRequest{
		TenantID:        tnt.ID,
		DocumentID:      doc.ID,
		RouteID:         routeID,
		SubmittedBy:     actor.ID,
		RevisionTitle:   "Initial",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 0,
		IdempotencyKey:  "idem-" + uuid.NewString(),
	}

	if _, err := svc.SubmitRevisionForReview(submitCtxWithIdentity(tnt.ID, actor.ID), runner, req); err != nil {
		t.Fatalf("SubmitRevisionForReview: %v", err)
	}

	if len(notifier.sent) != 1 || notifier.sent[0].SubjectKind != "document" {
		t.Fatalf("sent = %+v, want exactly one document-subject event", notifier.sent)
	}
	if notifier.sent[0].SubjectID != doc.ID {
		t.Fatalf("subject id = %q, want %q", notifier.sent[0].SubjectID, doc.ID)
	}
	if notifier.sent[0].EventType != approvaldomain.EventTypeApprovalPending {
		t.Fatalf("event type = %q, want %q", notifier.sent[0].EventType, approvaldomain.EventTypeApprovalPending)
	}
}

// The document path must gain the exact same rollback guarantee as the
// template path: if the notification cannot be promised, the document
// submission must not stick either. Proven against the database, not just
// against the returned error.
func TestDocumentSubmit_EnqueueFailure_RollsBackSubmission(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	doc := testdb.NewDocument(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithOwner(actor.ID), testdb.WithStatus("draft"), testdb.WithRevisionNumber(0), testdb.WithSubmitReadySnapshots())

	var profileCode string
	if err := dbc.QueryRowContext(ctx,
		`SELECT profile_code FROM public.controlled_documents WHERE id = $1::uuid`, doc.ControlledDocumentID,
	).Scan(&profileCode); err != nil {
		t.Fatalf("lookup profile_code: %v", err)
	}
	routeID := seedSubmitRouteWithStage(t, dbc, tnt.ID, profileCode)

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	svc := &SubmitService{repo: repo, emitter: NewSQLEmitter(), clock: RealClock{}, notifier: failingNotifier{}}
	runner := db.NewTxRunner(dbc)

	req := SubmitRequest{
		TenantID:        tnt.ID,
		DocumentID:      doc.ID,
		RouteID:         routeID,
		SubmittedBy:     actor.ID,
		RevisionTitle:   "Initial",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 0,
		IdempotencyKey:  "idem-" + uuid.NewString(),
	}

	_, err := svc.SubmitRevisionForReview(submitCtxWithIdentity(tnt.ID, actor.ID), runner, req)
	if err == nil {
		t.Fatal("SubmitRevisionForReview: got nil error, want the enqueue failure to be surfaced — a swallowed error here means submit told nobody AND said nothing about it")
	}
	if !errors.Is(err, errEnqueueBroken) {
		t.Fatalf("SubmitRevisionForReview error = %v, want it to wrap %v so a caller can tell why the submission was refused", err, errEnqueueBroken)
	}

	// Prove the tx actually rolled back: no approval instance persisted, and
	// the document never left draft. An error return with a committed row IS
	// the half-transaction Task 6 exists to prevent.
	var instanceCount int
	if err := dbc.QueryRowContext(ctx,
		`SELECT count(*) FROM approval_instances WHERE tenant_id = $1 AND document_id = $2`,
		tnt.ID, doc.ID,
	).Scan(&instanceCount); err != nil {
		t.Fatalf("query approval_instances: %v", err)
	}
	if instanceCount != 0 {
		t.Fatalf("approval_instances rows for this document = %d, want 0 — an approval instance was persisted despite the enqueue failure, exactly the half-transaction Task 6 was built to prevent", instanceCount)
	}

	var docStatus string
	var revisionVersion int
	if err := dbc.QueryRowContext(ctx,
		`SELECT status, revision_version FROM documents WHERE id = $1 AND tenant_id = $2`,
		doc.ID, tnt.ID,
	).Scan(&docStatus, &revisionVersion); err != nil {
		t.Fatalf("query documents: %v", err)
	}
	if docStatus != "draft" {
		t.Fatalf("document status = %q, want %q — the document was pushed into under_review even though the submission that should have caused it failed", docStatus, "draft")
	}
	if revisionVersion != req.RevisionVersion {
		t.Fatalf("document revision_version = %d, want unchanged %d — the OCC counter advanced despite the rolled-back submission", revisionVersion, req.RevisionVersion)
	}
}
