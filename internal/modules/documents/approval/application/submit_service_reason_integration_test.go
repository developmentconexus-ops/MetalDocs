//go:build integration
// +build integration

package application

// F6.3 — structured reason-for-change real-DB integration proof
// (validation-contract.md §5.4). Runs SubmitService.SubmitRevisionForReview
// against a real Postgres testdb (ADR 0034 harness), proving:
//
//  1. TestSubmitPersistsReason_RealDB — a REV>=1 submit with reason_for_change
//     (+ reason_category) persists BOTH columns on public.documents via the
//     SAME draft->under_review UPDATE that sets revision_title (no second
//     write; revision_title untouched by the reason value).
//  2. TestSubmitReasonOnAuditTrail_RealDB — the submit emits exactly one
//     governance_events row (event_type = 'approval_submitted') whose
//     payload_json carries reason_for_change (+ reason_category), inserted in
//     the SAME transaction as the documents UPDATE (both survive one commit).
//
// Unlike the fake-driver unit tests in submit_service_test.go (which assert on
// the bound SQL args), this proves the full path against real DB CHECKs
// (ck_documents_reason_category, migration 0274) and real governance_events
// persistence.
//
// system_admin bypasses the document.submit / document.edit capability checks
// (ADR 0022 Phase 11 F8), avoiding the need to seed role_capabilities grants;
// the bypass path still requires the tenant_id/actor_id GUCs, seeded here via
// platformtenant context (the same chokepoint TxRunner.Do uses in production).

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"

	approvalrepo "metaldocs/internal/modules/documents/approval/repository"
)

// seedSubmitRouteWithStage seeds one approval_routes row + one
// approval_route_stages row (quorum any_1_of, no quorum_m) so
// SubmitService.SubmitRevisionForReview's route.Validate() and stage-instance
// creation succeed. NewApprovalRoute (testdb factory) does not seed stages.
func seedSubmitRouteWithStage(t *testing.T, dbc *sql.DB, tenantID, profileCode string) string {
	t.Helper()
	route := testdb.NewApprovalRoute(t, dbc, testdb.WithTenant(tenantID), testdb.WithProfile(profileCode))
	if _, err := dbc.ExecContext(context.Background(),
		`INSERT INTO public.approval_route_stages
		   (route_id, stage_order, name, required_role, required_capability, area_code, quorum, on_eligibility_drift)
		 VALUES ($1::uuid, 1, 'Stage 1', 'approver', 'document.signoff', 'area-001', 'any_1_of', 'keep_snapshot')`,
		route.ID,
	); err != nil {
		t.Fatalf("seed approval_route_stages: %v", err)
	}
	return route.ID
}

// submitCtxWithIdentity builds the ctx TxRunner.Do's chokepoint auto-seed
// expects: tenant + actor identity, so seedTxIdentityFromContext seeds the
// metaldocs.tenant_id/actor_id GUCs authz.Require reads (F3.1/M3 contract).
func submitCtxWithIdentity(tenantID, actorID string) context.Context {
	ctx := platformtenant.WithTenantID(context.Background(), tenantID)
	ctx = platformtenant.WithActorID(ctx, actorID)
	return ctx
}

func TestSubmitPersistsReason_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	// WithRevisionNumber(1): the row must be a governed revision >= 1 so
	// SubmitRevisionForReview's in-tx derived revision_number (T8b) drives the
	// REV>=1 gates — the request itself never supplies a revision number.
	doc := testdb.NewDocument(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithOwner(actor.ID), testdb.WithStatus("draft"), testdb.WithRevisionNumber(1))

	var profileCode string
	if err := dbc.QueryRowContext(ctx,
		`SELECT profile_code FROM public.controlled_documents WHERE id = $1::uuid`, doc.ControlledDocumentID,
	).Scan(&profileCode); err != nil {
		t.Fatalf("lookup profile_code: %v", err)
	}
	routeID := seedSubmitRouteWithStage(t, dbc, tnt.ID, profileCode)

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	svc := &SubmitService{repo: repo, emitter: NewSQLEmitter(), clock: RealClock{}}
	runner := db.NewTxRunner(dbc)

	req := SubmitRequest{
		TenantID:        tnt.ID,
		DocumentID:      doc.ID,
		RouteID:         routeID,
		SubmittedBy:     actor.ID,
		RevisionTitle:   "Revisao corretiva",
		ReasonForChange: "Corrected dimensional tolerance per customer complaint #482",
		ReasonCategory:  "corrective",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 0,
		IdempotencyKey:  "11111111-1111-1111-1111-111111111111",
	}

	callCtx := submitCtxWithIdentity(tnt.ID, actor.ID)
	if _, err := svc.SubmitRevisionForReview(callCtx, runner, req); err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}

	var gotTitle, gotReason, gotCategory string
	if err := dbc.QueryRowContext(ctx,
		`SELECT revision_title, reason_for_change, reason_category FROM public.documents WHERE id = $1::uuid`,
		doc.ID,
	).Scan(&gotTitle, &gotReason, &gotCategory); err != nil {
		t.Fatalf("read back documents row: %v", err)
	}
	if gotTitle != req.RevisionTitle {
		t.Fatalf("revision_title = %q, want %q (must not be overwritten by reason_for_change)", gotTitle, req.RevisionTitle)
	}
	if gotReason != req.ReasonForChange {
		t.Fatalf("reason_for_change = %q, want %q", gotReason, req.ReasonForChange)
	}
	if gotCategory != req.ReasonCategory {
		t.Fatalf("reason_category = %q, want %q", gotCategory, req.ReasonCategory)
	}
}

func TestSubmitReasonOnAuditTrail_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	// WithRevisionNumber(1): see TestSubmitPersistsReason_RealDB — the derived
	// in-tx revision_number (T8b) must come from the row, not the request.
	doc := testdb.NewDocument(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithOwner(actor.ID), testdb.WithStatus("draft"), testdb.WithRevisionNumber(1))

	var profileCode string
	if err := dbc.QueryRowContext(ctx,
		`SELECT profile_code FROM public.controlled_documents WHERE id = $1::uuid`, doc.ControlledDocumentID,
	).Scan(&profileCode); err != nil {
		t.Fatalf("lookup profile_code: %v", err)
	}
	routeID := seedSubmitRouteWithStage(t, dbc, tnt.ID, profileCode)

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	svc := &SubmitService{repo: repo, emitter: NewSQLEmitter(), clock: RealClock{}}
	runner := db.NewTxRunner(dbc)

	req := SubmitRequest{
		TenantID:        tnt.ID,
		DocumentID:      doc.ID,
		RouteID:         routeID,
		SubmittedBy:     actor.ID,
		RevisionTitle:   "Revisao regulatoria",
		ReasonForChange: "Updated per new ISO 9001:2026 clause 8.5.6",
		ReasonCategory:  "regulatory",
		ContentFormData: map[string]any{"title": "My Doc"},
		RevisionVersion: 0,
		IdempotencyKey:  "22222222-2222-2222-2222-222222222222",
	}

	callCtx := submitCtxWithIdentity(tnt.ID, actor.ID)
	if _, err := svc.SubmitRevisionForReview(callCtx, runner, req); err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}

	rows, err := dbc.QueryContext(ctx,
		`SELECT payload_json FROM public.governance_events
		  WHERE tenant_id = $1::uuid AND resource_id = $2::uuid AND event_type = 'approval_submitted'`,
		tnt.ID, doc.ID,
	)
	if err != nil {
		t.Fatalf("query governance_events: %v", err)
	}
	defer rows.Close()

	var payloads []map[string]any
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			t.Fatalf("scan payload_json: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Fatalf("unmarshal payload_json: %v", err)
		}
		payloads = append(payloads, payload)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}

	if len(payloads) != 1 {
		t.Fatalf("governance_events approval_submitted rows for this document = %d, want exactly 1", len(payloads))
	}
	if payloads[0]["reason_for_change"] != req.ReasonForChange {
		t.Fatalf("payload.reason_for_change = %v, want %q", payloads[0]["reason_for_change"], req.ReasonForChange)
	}
	if payloads[0]["reason_category"] != req.ReasonCategory {
		t.Fatalf("payload.reason_category = %v, want %q", payloads[0]["reason_category"], req.ReasonCategory)
	}
}
