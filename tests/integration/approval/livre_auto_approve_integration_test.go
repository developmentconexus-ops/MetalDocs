//go:build integration
// +build integration

// Package approval_test — ADR 0087 (livre = configured no-approval route).
// Exercises the submit-time auto-approval end-to-end against a real Postgres
// instance (testdb per-test clone):
//
//   - a livre profile's ACTIVE route carries ZERO stages, and submitting a
//     document against it completes in the SAME call/tx: the instance is
//     approved, the document is approved, and no stage instance exists;
//   - the completion runs through the shared terminal-approval path, so the
//     ADR 0085 release recorder fires (the F-QA4-14 defect class: one route to
//     terminal approval silently skipping publication side effects) and the
//     ADR 0044 document.approved lifecycle event is enqueued;
//   - the auto-approval is audited as approval_auto_approved, naming the route
//     and route version that authorized it;
//   - submit idempotency is unchanged: a retry with the same key returns the
//     same instance and does not double-approve or re-emit.
package approval_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	docsdomain "metaldocs/internal/modules/documents/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// recordingReleaseRecorder captures every ADR 0085 terminal-approval call so a
// test can assert the publication side effect fired (rather than only that the
// instance flipped status).
type recordingReleaseRecorder struct {
	calls []application.TerminalApprovalInput
}

func (r *recordingReleaseRecorder) RecordTerminalApproval(_ context.Context, _ db.Tx, in application.TerminalApprovalInput) (string, error) {
	r.calls = append(r.calls, in)
	return "", nil
}

// recordingLifecycleEnqueuer captures ADR 0044 lifecycle-event enqueues.
type recordingLifecycleEnqueuer struct {
	args []docsdomain.LifecycleEventArgs
}

func (r *recordingLifecycleEnqueuer) EnqueueLifecycleEventTx(_ context.Context, _ db.Tx, args docsdomain.LifecycleEventArgs) error {
	r.args = append(r.args, args)
	return nil
}

// livreFixture is a livre profile with a zero-stage active route and a
// submit-ready draft on it.
type livreFixture struct {
	tenantID   string
	authorID   string
	documentID string
	routeID    string
	areaCode   string
	services   *application.Services
	runner     db.TxRunner
	recorder   *recordingReleaseRecorder
	lifecycle  *recordingLifecycleEnqueuer
}

func seedLivreFixture(t *testing.T, database *sql.DB) livreFixture {
	t.Helper()
	ctx := context.Background()

	tenant := testdb.NewTenant(t, database)
	// system_admin bypasses the document.submit/document.edit capability
	// checks (ADR 0022 Phase 11 F8), mirroring the freeze integration test —
	// this test exercises the auto-approve path, not capability plumbing.
	author := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Author"), testdb.WithRole("system_admin"))

	tax := testdb.NewTaxonomy(t, database, testdb.WithTenant(tenant.ID), testdb.WithGovernanceClass("livre"))
	cd := testdb.NewControlledDoc(t, database,
		testdb.WithTenant(tenant.ID),
		testdb.WithTaxonomy(tax),
		testdb.WithOwner(author.ID),
	)
	doc := testdb.NewDocument(t, database,
		testdb.WithControlledDoc(cd),
		testdb.WithOwner(author.ID),
		testdb.WithStatus("draft"),
		testdb.WithSubmitReadySnapshots(),
	)

	// The submit path reads the document's process-area snapshot for the
	// area-scoped authz checks; bind it to the profile's own area.
	testdb.SeedWithCaps(t, database, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`UPDATE public.documents SET process_area_code_snapshot = $1 WHERE id = $2::uuid`,
			tax.ProcessAreaCode, doc.ID,
		)
		return err
	})

	// The livre route: active, profile-keyed, and deliberately STAGELESS. Under
	// ADR 0087 / migration 0316 this is the required shape, not a smuggled one.
	route := testdb.NewApprovalRoute(t, database,
		testdb.WithTenant(tenant.ID),
		testdb.WithProfile(tax.ProfileCode),
		testdb.WithOwner(author.ID),
	)

	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})
	runner := db.NewTxRunner(database)
	recorder := &recordingReleaseRecorder{}
	lifecycle := &recordingLifecycleEnqueuer{}
	services := application.NewServices(repo, application.NewSQLEmitter(), application.RealClock{}, nil)
	services.WithReleaseRecorder(recorder)
	services.WithLifecycleEnqueuer(lifecycle)

	return livreFixture{
		tenantID: tenant.ID, authorID: author.ID, documentID: doc.ID,
		routeID: route.ID, areaCode: tax.ProcessAreaCode,
		services: services, runner: runner,
		recorder: recorder, lifecycle: lifecycle,
	}
}

func (fx livreFixture) submit(t *testing.T, idempKey string) (application.SubmitResult, error) {
	t.Helper()
	return fx.services.Submit.SubmitRevisionForReview(
		testdb.AuthzCtx(fx.tenantID, fx.authorID), fx.runner,
		application.SubmitRequest{
			TenantID:        fx.tenantID,
			DocumentID:      fx.documentID,
			RouteID:         fx.routeID,
			SubmittedBy:     fx.authorID,
			RevisionTitle:   "Initial",
			ContentFormData: map[string]any{"title": "Doc"},
			RevisionVersion: 0,
			IdempotencyKey:  idempKey,
		})
}

// TestLivreSubmit_ZeroStageRoute_AutoApprovesInSameCall is the ADR 0087 happy
// path: submit against a stageless route and the instance comes back approved,
// with the document approved and no stage instance ever created.
func TestLivreSubmit_ZeroStageRoute_AutoApprovesInSameCall(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedLivreFixture(t, database)

	result, err := fx.submit(t, "livre-auto-approve-1")
	if err != nil {
		t.Fatalf("SubmitRevisionForReview against a livre zero-stage route: %v", err)
	}

	if got := instanceStatus(t, database, result.InstanceID); got != string(domain.InstanceApproved) {
		t.Fatalf("instance status = %q; want approved in the same submit call", got)
	}
	var docStatus string
	if err := database.QueryRowContext(context.Background(),
		`SELECT status FROM public.documents WHERE id = $1::uuid`, fx.documentID,
	).Scan(&docStatus); err != nil {
		t.Fatalf("query document status: %v", err)
	}
	if docStatus != "approved" {
		t.Fatalf("document status = %q; want approved", docStatus)
	}

	var stageCount int
	if err := database.QueryRowContext(context.Background(),
		`SELECT count(*) FROM public.approval_stage_instances WHERE approval_instance_id = $1::uuid`,
		result.InstanceID,
	).Scan(&stageCount); err != nil {
		t.Fatalf("count stage instances: %v", err)
	}
	if stageCount != 0 {
		t.Fatalf("stage instances = %d; want 0 (a livre route has no stages to instantiate)", stageCount)
	}
}

// TestLivreSubmit_AutoApprove_EmitsAuditEvent asserts the auto-approval is on
// the record with the route + route version that authorized it — the answer to
// "why is this approved with an empty signoff set?".
func TestLivreSubmit_AutoApprove_EmitsAuditEvent(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedLivreFixture(t, database)

	result, err := fx.submit(t, "livre-auto-approve-audit-1")
	if err != nil {
		t.Fatalf("SubmitRevisionForReview: %v", err)
	}

	var payload []byte
	if err := database.QueryRowContext(context.Background(), `
		SELECT payload_json FROM public.governance_events
		 WHERE tenant_id = $1::uuid AND event_type = $2 AND resource_id = $3`,
		fx.tenantID, string(application.EventTypeApprovalAutoApproved), result.InstanceID,
	).Scan(&payload); err != nil {
		t.Fatalf("query approval_auto_approved event: %v", err)
	}

	var got struct {
		RouteID      string `json:"route_id"`
		RouteVersion int    `json:"route_version"`
		StageCount   int    `json:"stage_count"`
		Reason       string `json:"reason"`
	}
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("unmarshal payload %s: %v", payload, err)
	}
	if got.RouteID != fx.routeID {
		t.Errorf("payload route_id = %q; want %q", got.RouteID, fx.routeID)
	}
	if got.RouteVersion < 1 {
		t.Errorf("payload route_version = %d; want the route's version", got.RouteVersion)
	}
	if got.StageCount != 0 {
		t.Errorf("payload stage_count = %d; want 0", got.StageCount)
	}
	if got.Reason != "zero_stage_route" {
		t.Errorf("payload reason = %q; want zero_stage_route", got.Reason)
	}

	// The submission itself is still recorded — the trail reads
	// submitted-then-approved, not approved-out-of-nowhere.
	var submitted int
	if err := database.QueryRowContext(context.Background(), `
		SELECT count(*) FROM public.governance_events
		 WHERE tenant_id = $1::uuid AND event_type = 'approval_submitted' AND resource_id = $2`,
		fx.tenantID, fx.documentID,
	).Scan(&submitted); err != nil {
		t.Fatalf("count approval_submitted: %v", err)
	}
	if submitted != 1 {
		t.Fatalf("approval_submitted events = %d; want 1", submitted)
	}
}

// TestLivreSubmit_AutoApprove_FiresPublicationSideEffects is the F-QA4-14 guard:
// the auto-approve leg must reach the SAME shared terminal-approval path the
// signoff and review-verdict legs reach, so the ADR 0085 release recorder and
// the ADR 0044 document.approved lifecycle event both fire.
func TestLivreSubmit_AutoApprove_FiresPublicationSideEffects(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedLivreFixture(t, database)

	result, err := fx.submit(t, "livre-auto-approve-sideeffects-1")
	if err != nil {
		t.Fatalf("SubmitRevisionForReview: %v", err)
	}

	if len(fx.recorder.calls) != 1 {
		t.Fatalf("RecordTerminalApproval calls = %d; want exactly 1 (ADR 0085 approval fact + pin + coordinator)", len(fx.recorder.calls))
	}
	call := fx.recorder.calls[0]
	if call.InstanceID != result.InstanceID {
		t.Errorf("recorded instance_id = %q; want %q", call.InstanceID, result.InstanceID)
	}
	if call.DocumentID != fx.documentID {
		t.Errorf("recorded document_id = %q; want %q", call.DocumentID, fx.documentID)
	}
	if call.FinalApproverID != fx.authorID {
		t.Errorf("recorded final approver = %q; want the submitter %q (on a zero-stage route the submission IS the approving act)", call.FinalApproverID, fx.authorID)
	}
	// The pin is what the release/publication path renders and verifies
	// against; an empty hash here is the F-QA4-14 shape (approved document
	// that can never materialize), so assert it explicitly.
	if call.FrozenContentHash == "" {
		t.Errorf("recorded frozen_content_hash is empty; want the freeze pin taken at submit (ADR 0085 no-fallback: the release path reads ONLY this pin)")
	}

	if len(fx.lifecycle.args) != 1 {
		t.Fatalf("lifecycle enqueues = %d; want exactly 1", len(fx.lifecycle.args))
	}
	if got := fx.lifecycle.args[0].EventType; got != docsdomain.EventTypeDocumentApproved {
		t.Fatalf("lifecycle event type = %q; want %q", got, docsdomain.EventTypeDocumentApproved)
	}
}

// TestLivreSubmit_AutoApprove_IdempotentRetry asserts submit idempotency is
// untouched by ADR 0087: replaying the same key is still rejected by the
// service-level duplicate-submission sentinel (unchanged pre-existing
// semantics — the HTTP boundary maps it to conflict.duplicate_submission), and
// crucially the replay neither re-approves nor re-fires the terminal-approval
// side effects.
func TestLivreSubmit_AutoApprove_IdempotentRetry(t *testing.T) {
	database, _ := testdb.Open(t)
	fx := seedLivreFixture(t, database)

	if _, err := fx.submit(t, "livre-auto-approve-idemp-1"); err != nil {
		t.Fatalf("first submit: %v", err)
	}
	if _, err := fx.submit(t, "livre-auto-approve-idemp-1"); !errors.Is(err, infrastructure.ErrDuplicateSubmission) {
		t.Fatalf("replayed submit err = %v; want ErrDuplicateSubmission", err)
	}

	if len(fx.recorder.calls) != 1 {
		t.Fatalf("RecordTerminalApproval calls after replay = %d; want 1", len(fx.recorder.calls))
	}
	var events int
	if err := database.QueryRowContext(context.Background(), `
		SELECT count(*) FROM public.governance_events
		 WHERE tenant_id = $1::uuid AND event_type = $2`,
		fx.tenantID, string(application.EventTypeApprovalAutoApproved),
	).Scan(&events); err != nil {
		t.Fatalf("count approval_auto_approved: %v", err)
	}
	if events != 1 {
		t.Fatalf("approval_auto_approved events after replay = %d; want 1", events)
	}
}
