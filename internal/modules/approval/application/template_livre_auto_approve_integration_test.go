//go:build integration
// +build integration

package application

// ADR 0087 §5 (P1-3) — a LIVRE profile's TEMPLATE route carries zero stages, so
// template submit itself reaches terminal approval: the instance completes
// in-tx and the template version travels its NORMAL completion path
// (under_review -> approved via the approval-owned TemplateCompletionWriter),
// exactly as a signoff-driven terminal decision would. Without this the
// zero-stage template route parked the instance in_progress forever with the
// version stuck under_review and no actor able to move it — a dead end, since
// there is no stage to sign.
//
// Proves:
//  1. submit against a zero-stage template route leaves the instance approved
//     and the template version approved (real templates-infrastructure
//     adapter, not a local double — the module seam is the thing under test);
//  2. the approval_auto_approved governance event exists, carrying the subject
//     (template) so the record is subject-identifiable;
//  3. replay of the same idempotency key adds no second event and no second
//     completion.

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
	iampostgres "metaldocs/internal/modules/iam/infrastructure/postgres"
	templatesinfra "metaldocs/internal/modules/templates/infrastructure"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"

	"github.com/google/uuid"
)

// seedZeroStageTemplateRoute inserts an ACTIVE (subject_kind='template')
// route with NO stages — the livre shape. It goes in directly rather than
// through RouteAdminService.Create so the test does not depend on the HTTP/
// admin path; the DB's own route-shape trigger still validates the committed
// shape, so an illegal shape here would fail loudly at COMMIT.
func seedZeroStageTemplateRoute(t *testing.T, dbc *sql.DB, tenantID, actorID, profileCode string) string {
	t.Helper()
	routeID := uuid.NewString()
	testdb.SeedWithCaps(t, dbc, `[{"cap":"approval.route.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO public.approval_routes
			   (id, tenant_id, name, profile_code, subject_kind, subject_key, active, created_by)
			 VALUES ($1::uuid, $2::uuid, 'Livre Template Route', $3, 'template', $3, true, $4)`,
			routeID, tenantID, profileCode, actorID,
		)
		return err
	})
	return routeID
}

func TestSubmitTemplateVersion_ZeroStageLivreRoute_AutoApproves_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	// Fixture default class is livre; its route must carry zero stages.
	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID))
	templateID, versionID := seedTemplateVersion(t, dbc, tnt.ID, actor.ID, tax.ProfileCode, "draft")
	routeID := seedZeroStageTemplateRoute(t, dbc, tnt.ID, actor.ID, tax.ProfileCode)

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, iampostgres.NewUserDisplayNameRepository(dbc))
	services := NewServices(repo, NewSQLEmitter(), RealClock{}, nil).
		WithTemplateVersionReader(templatesinfra.NewApprovalVersionReader()).
		WithTemplateCompletionWriter(templatesinfra.NewApprovalCompletionWriter())
	runner := db.NewTxRunner(dbc)

	idemKey := "idem-" + uuid.NewString()
	res, err := services.TemplateSubmit.SubmitTemplateVersionForReview(newTemplateSubmitCtx(tnt.ID, actor.ID), runner, TemplateSubmitRequest{
		TenantID:          tnt.ID,
		TemplateID:        templateID,
		TemplateVersionID: versionID,
		SubmittedBy:       actor.ID,
		IdempotencyKey:    idemKey,
	})
	if err != nil {
		t.Fatalf("SubmitTemplateVersionForReview: %v", err)
	}

	ctx := context.Background()

	var status string
	var completedAt any
	if err := dbc.QueryRowContext(ctx,
		`SELECT status, completed_at FROM public.approval_instances WHERE id = $1::uuid`,
		res.InstanceID,
	).Scan(&status, &completedAt); err != nil {
		t.Fatalf("read instance: %v", err)
	}
	if status != "approved" {
		t.Fatalf("instance status = %q, want approved (a zero-stage route auto-approves at submit; in_progress here means the template subject was left in a dead end)", status)
	}
	if completedAt == nil {
		t.Fatal("instance completed_at is NULL; the terminal-approval helper must stamp it (the render path reads it when there are no signoffs)")
	}
	if routeID == "" {
		t.Fatal("route id empty")
	}

	if got := templateVersionStatus(t, dbc, tnt.ID, versionID); got != "approved" {
		t.Fatalf("template version status = %q, want approved (normal completion path, not a hand-rolled UPDATE)", got)
	}

	var events int
	if err := dbc.QueryRowContext(ctx, `
		SELECT count(*) FROM public.governance_events
		 WHERE tenant_id = $1::uuid AND event_type = $2 AND resource_id = $3`,
		tnt.ID, string(EventTypeApprovalAutoApproved), res.InstanceID,
	).Scan(&events); err != nil {
		t.Fatalf("count auto-approve events: %v", err)
	}
	if events != 1 {
		t.Fatalf("approval_auto_approved events = %d, want 1 (the governance event IS the record of an auto-approval, ADR 0087 §5)", events)
	}

	// Replay: same idempotency key must not complete a second time.
	_, err = services.TemplateSubmit.SubmitTemplateVersionForReview(newTemplateSubmitCtx(tnt.ID, actor.ID), runner, TemplateSubmitRequest{
		TenantID:          tnt.ID,
		TemplateID:        templateID,
		TemplateVersionID: versionID,
		SubmittedBy:       actor.ID,
		IdempotencyKey:    idemKey,
	})
	if err == nil {
		t.Fatal("replayed submit succeeded; want a duplicate/precondition error")
	}
	if errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error kind: %v", err)
	}
	if err := dbc.QueryRowContext(ctx, `
		SELECT count(*) FROM public.governance_events
		 WHERE tenant_id = $1::uuid AND event_type = $2 AND resource_id = $3`,
		tnt.ID, string(EventTypeApprovalAutoApproved), res.InstanceID,
	).Scan(&events); err != nil {
		t.Fatalf("recount auto-approve events: %v", err)
	}
	if events != 1 {
		t.Fatalf("approval_auto_approved events after replay = %d, want 1", events)
	}
}
