//go:build integration
// +build integration

package application

// Task 1 (SDD plan, approval-accountability-loop) — real-DB proof that the
// area pre-check in loadInstanceAreaCode stops 404-ing template approval
// instances. A template instance has document_id IS NULL (ADR 0082, the M3
// kernel extraction); the INNER JOIN documents in the pre-check dropped
// every such row before the precedence chain (stage snapshot -> document
// snapshot -> CD area) ever ran, so ReadService callers saw
// ErrNoActiveInstance and 404'd a roster the repository layer one rung below
// already resolved correctly (that layer was corrected to a LEFT JOIN during
// the extraction; this pre-check was missed).
//
// This file is package application (white-box), not application_test: the
// task brief's suggested harness names (env.Factory(), f.ApprovalRoute,
// testdb.SubjectKindTemplate, testdb.WithActiveStageArea,
// env.LoadInstanceAreaCodeForTest) do not exist under those names in
// tests/integration/testdb — the real API is package-level builders
// (testdb.NewTenant, testdb.NewApprovalRoute, ...) returning a bare *sql.DB.
// Rather than invent a parallel harness or export loadInstanceAreaCode
// solely for this test, it is driven directly here — the same white-box
// access pattern every other *_integration_test.go file in this package
// already uses (e.g. template_submit_service_integration_test.go).

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/tests/integration/testdb"
)

// seedTemplateApprovalInstance inserts a public.approval_instances row with
// subject_kind='template' and document_id left NULL — the exact shape the
// INNER JOIN bug dropped (approval_instances_template_subject_projection_check
// requires document_id IS NULL whenever subject_kind='template'). Wired to
// routeID; status must be a legal approval_instances status.
func seedTemplateApprovalInstance(t *testing.T, dbc *sql.DB, tenantID, actorID, routeID, status string) string {
	t.Helper()
	id := uuid.NewString()
	subjectKey := uuid.NewString()
	idemKey := "idem-" + uuid.NewString()

	testdb.SeedWithCaps(t, dbc, `[{"cap":"template.submit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(), `
			INSERT INTO public.approval_instances
			  (id, tenant_id, route_id, route_version_snapshot, status, subject_kind, subject_key,
			   submitted_by, submitted_at, content_hash_at_submit, idempotency_key)
			VALUES ($1::uuid, $2::uuid, $3::uuid, 1, $4, 'template', $5, $6, now(), repeat('a', 64), $7)`,
			id, tenantID, routeID, status, subjectKey, actorID, idemKey,
		)
		return err
	})
	return id
}

// seedActiveStageSnapshot inserts an active public.approval_stage_instances
// row carrying areaCode as its area_code_snapshot — the value
// loadInstanceAreaCode's precedence chain must resolve for a template
// instance, since a template row has no documents row to fall back to.
func seedActiveStageSnapshot(t *testing.T, dbc *sql.DB, instanceID, areaCode string) {
	t.Helper()
	testdb.SeedWithCaps(t, dbc, `[{"cap":"template.submit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(), `
			INSERT INTO public.approval_stage_instances
			  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
			   area_code_snapshot, quorum_snapshot, on_eligibility_drift_snapshot,
			   eligible_actor_ids, status, stage_kind, required_capability_snapshot)
			VALUES ($1::uuid, 1, 'Stage 1', 'reviewer', $2, 'any_1_of', 'keep_snapshot',
			        '[]'::jsonb, 'active', 'approval', 'template.approve')`,
			instanceID, areaCode,
		)
		return err
	})
}

// TestLoadInstanceAreaCode_TemplateInstance_ResolvesFromStageSnapshot is the
// RED->GREEN target: before the LEFT JOIN fix, the INNER JOIN documents
// drops the template instance's only row entirely and the resolver reports
// found=false — the exact ErrNoActiveInstance 404 this task closes.
func TestLoadInstanceAreaCode_TemplateInstance_ResolvesFromStageSnapshot(t *testing.T) {
	dbc, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tenant.ID))
	route := testdb.NewApprovalRoute(t, dbc, testdb.WithTenant(tenant.ID), testdb.WithSubjectKind("template"))

	instanceID := seedTemplateApprovalInstance(t, dbc, tenant.ID, actor.ID, route.ID, "in_progress")
	seedActiveStageSnapshot(t, dbc, instanceID, "qualidade")

	tx, err := dbc.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	area, found, err := loadInstanceAreaCode(context.Background(), tx, controlleddocumentsdomain.NoopCDFieldReader{}, tenant.ID, instanceID)
	if err != nil {
		t.Fatalf("loadInstanceAreaCode: %v", err)
	}
	if !found {
		t.Fatal("found = false, want true — a template instance row exists and must be visible to the area pre-check")
	}
	if area != "qualidade" {
		t.Fatalf("area = %q, want %q (the active stage snapshot)", area, "qualidade")
	}
}

// TestLoadInstanceAreaCode_CompletedTemplate_EmptyAreaButFound proves the
// resolver bakes in no default: a completed template instance has NO active
// stage, so every branch of the precedence chain is NULL and area resolves
// to "". found must still be true — "" means "no area anywhere", which
// callers widen to "tenant" EXPLICITLY at the call site, never inside the
// resolver.
func TestLoadInstanceAreaCode_CompletedTemplate_EmptyAreaButFound(t *testing.T) {
	dbc, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tenant.ID))
	route := testdb.NewApprovalRoute(t, dbc, testdb.WithTenant(tenant.ID), testdb.WithSubjectKind("template"))

	instanceID := seedTemplateApprovalInstance(t, dbc, tenant.ID, actor.ID, route.ID, "approved")
	// No stage instance seeded: a completed instance carries none active.

	tx, err := dbc.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	area, found, err := loadInstanceAreaCode(context.Background(), tx, controlleddocumentsdomain.NoopCDFieldReader{}, tenant.ID, instanceID)
	if err != nil {
		t.Fatalf("loadInstanceAreaCode: %v", err)
	}
	if !found {
		t.Fatal("found = false, want true — the row exists even with no area")
	}
	if area != "" {
		t.Fatalf("area = %q, want \"\" — no default may be baked into the resolver", area)
	}
}
