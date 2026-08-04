//go:build integration
// +build integration

package infrastructure_test

// ADR 0087 (P2-6) — render resolver readers for a LIVRE document.
//
// A livre profile's active route carries zero stages, so submit auto-approves
// the document inside the submit transaction: the approval instance reaches
// status='approved' with completed_at set and NO approval_signoffs rows at
// all. That shape used to be pathological; it is now the normal case, so both
// workflow readers must handle it truthfully:
//
//   1. GetFinalApprovalDate must return the instance's completed_at, NOT a
//      fabricated now() (the superseded coalesce(max(signed_at), now()) would
//      have silently stamped render time onto every livre document);
//   2. GetFinalApprovalDate must FAIL CLOSED (ErrNoApprovalDate) when neither
//      a signoff nor a completed approved instance exists — never substitute;
//   3. GetApprovers must render an explicit auto-approval indication instead
//      of an empty block, which would read as "nobody approved this".
//
// A staged (controlado) document is asserted alongside as the control: the
// signoff path is unchanged.

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/tests/integration/testdb"
)

func TestWorkflowReader_LivreAutoApprovedDocument(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, db)
	// Default fixture class is livre, whose route is stageless by construction.
	doc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("approved"))
	inst := testdb.NewApprovalInstance(t, db, testdb.WithDocument(doc), testdb.WithStatus("approved"))

	// Auto-approve completes the instance in the submit tx; no signoff rows.
	completedAt := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	if _, err := db.ExecContext(ctx,
		`UPDATE public.approval_instances SET completed_at = $2 WHERE id = $1::uuid`,
		inst.ID, completedAt,
	); err != nil {
		t.Fatalf("set completed_at: %v", err)
	}

	reader := infrastructure.NewWorkflowReader(db)

	got, err := reader.GetFinalApprovalDate(ctx, tnt.ID, doc.ID)
	if err != nil {
		t.Fatalf("GetFinalApprovalDate: %v", err)
	}
	if !got.UTC().Equal(completedAt) {
		t.Fatalf("GetFinalApprovalDate = %s, want the instance completed_at %s (never a fabricated now())",
			got.UTC(), completedAt)
	}

	approvers, err := reader.GetApprovers(ctx, tnt.ID, doc.ID, inst.ID)
	if err != nil {
		t.Fatalf("GetApprovers: %v", err)
	}
	if len(approvers) != 1 {
		t.Fatalf("GetApprovers returned %d rows, want exactly 1 auto-approval row (a blank block reads as 'nobody approved this')", len(approvers))
	}
	if approvers[0].DisplayName != infrastructure.AutoApprovalDisplayName {
		t.Fatalf("approver display name = %q, want %q", approvers[0].DisplayName, infrastructure.AutoApprovalDisplayName)
	}
	if !approvers[0].SignedAt.UTC().Equal(completedAt) {
		t.Fatalf("auto-approval SignedAt = %s, want completed_at %s", approvers[0].SignedAt.UTC(), completedAt)
	}
}

func TestWorkflowReader_NoApprovalAtAll_FailsClosed(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, db)
	doc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("draft"))

	reader := infrastructure.NewWorkflowReader(db)
	if _, err := reader.GetFinalApprovalDate(ctx, tnt.ID, doc.ID); !errors.Is(err, infrastructure.ErrNoApprovalDate) {
		t.Fatalf("GetFinalApprovalDate error = %v, want ErrNoApprovalDate (no-fallback: never substitute now())", err)
	}
}

// TestWorkflowReader_ReviewOnlyRoute_NoAutoApprovalRow guards the inference the
// livre rendering must NOT make: "no signoffs" is not the same as "auto
// approved". A review-only route (all stages review-kind, e.g. a simples
// profile) reaches terminal approval through review verdicts and legitimately
// carries zero signoff rows. Rendering the livre auto-approval line there would
// fabricate a governance fact — and attribute it to the submitter. The gate is
// the instance's pinned shape: it has stage instances, therefore it is NOT the
// ADR 0087 stageless route, therefore the approvers block stays as it was
// before ADR 0087 (empty).
func TestWorkflowReader_ReviewOnlyRoute_NoAutoApprovalRow(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, db)
	doc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("approved"))
	inst := testdb.NewApprovalInstance(t, db, testdb.WithDocument(doc), testdb.WithStatus("approved"))

	completedAt := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	if _, err := db.ExecContext(ctx,
		`UPDATE public.approval_instances SET completed_at = $2 WHERE id = $1::uuid`,
		inst.ID, completedAt,
	); err != nil {
		t.Fatalf("set completed_at: %v", err)
	}
	reviewer := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithDisplayName("Rui Reviewer"))
	if _, err := db.ExecContext(ctx, `
		INSERT INTO public.approval_stage_instances
		  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
		   required_capability_snapshot, area_code_snapshot, quorum_snapshot,
		   on_eligibility_drift_snapshot, eligible_actor_ids, status, stage_kind)
		VALUES ($1::uuid, 1, 'Revisão', 'reviewer', 'document.review', 'area-1',
		        'any_1_of', 'keep_snapshot', jsonb_build_array($2::text), 'completed', 'review')`,
		inst.ID, reviewer.ID,
	); err != nil {
		t.Fatalf("seed review stage instance: %v", err)
	}

	reader := infrastructure.NewWorkflowReader(db)
	approvers, err := reader.GetApprovers(ctx, tnt.ID, doc.ID, inst.ID)
	if err != nil {
		t.Fatalf("GetApprovers: %v", err)
	}
	if len(approvers) != 0 {
		t.Fatalf("GetApprovers = %+v, want empty: a review-only route has no signoffs but is NOT a livre auto-approval", approvers)
	}

	// The date still resolves from the instance completion — that part is
	// shape-independent and was already correct.
	got, err := reader.GetFinalApprovalDate(ctx, tnt.ID, doc.ID)
	if err != nil {
		t.Fatalf("GetFinalApprovalDate: %v", err)
	}
	if !got.UTC().Equal(completedAt) {
		t.Fatalf("GetFinalApprovalDate = %s, want completed_at %s", got.UTC(), completedAt)
	}
}

func TestWorkflowReader_SignedDocument_UsesSignoffDate(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, db)
	doc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("approved"))
	inst := testdb.NewApprovalInstance(t, db, testdb.WithDocument(doc), testdb.WithStatus("approved"))

	signer := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithDisplayName("Ana Signer"))
	signedAt := time.Now().UTC().Add(-30 * time.Minute).Truncate(time.Second)
	completedAt := signedAt.Add(time.Minute)
	if _, err := db.ExecContext(ctx,
		`UPDATE public.approval_instances SET completed_at = $2 WHERE id = $1::uuid`,
		inst.ID, completedAt,
	); err != nil {
		t.Fatalf("set completed_at: %v", err)
	}
	var stageInstanceID string
	if err := db.QueryRowContext(ctx, `
		INSERT INTO public.approval_stage_instances
		  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
		   required_capability_snapshot, area_code_snapshot, quorum_snapshot,
		   on_eligibility_drift_snapshot, eligible_actor_ids, status)
		VALUES ($1::uuid, 1, 'Stage 1', 'approver', 'document.signoff', 'area-1',
		        'any_1_of', 'keep_snapshot', jsonb_build_array($2::text), 'completed')
		RETURNING id::text`,
		inst.ID, signer.ID,
	).Scan(&stageInstanceID); err != nil {
		t.Fatalf("seed approval_stage_instances: %v", err)
	}
	testdb.SeedWithCaps(t, db, `[{"cap":"document.signoff"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO public.approval_signoffs
			  (approval_instance_id, stage_instance_id, actor_tenant_id, actor_user_id,
			   actor_display_name_snapshot, decision, signed_at, signature_method,
			   signature_payload, content_hash)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4, 'Ana Signer', 'approve', $5,
			        'password_reauth', '{}'::jsonb, repeat('a', 64))`,
			inst.ID, stageInstanceID, tnt.ID, signer.ID, signedAt,
		)
		return err
	})

	reader := infrastructure.NewWorkflowReader(db)
	got, err := reader.GetFinalApprovalDate(ctx, tnt.ID, doc.ID)
	if err != nil {
		t.Fatalf("GetFinalApprovalDate: %v", err)
	}
	if !got.UTC().Equal(signedAt) {
		t.Fatalf("GetFinalApprovalDate = %s, want the signoff date %s (signoff wins over completed_at)", got.UTC(), signedAt)
	}

	approvers, err := reader.GetApprovers(ctx, tnt.ID, doc.ID, inst.ID)
	if err != nil {
		t.Fatalf("GetApprovers: %v", err)
	}
	if len(approvers) != 1 || approvers[0].DisplayName != "Ana Signer" {
		t.Fatalf("GetApprovers = %+v, want the real signer row (auto-approval line must NOT appear)", approvers)
	}
}
