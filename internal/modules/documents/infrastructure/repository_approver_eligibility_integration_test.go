//go:build integration
// +build integration

package infrastructure_test

// TestIsEligibleApprover verifies the three invariants of IsEligibleApprover:
//   - returns true  when the actor is in eligible_actor_ids of the open stage
//   - returns false when the actor is NOT in eligible_actor_ids (stranger)
//   - returns false when queried against a different document (wrong doc)

import (
	"context"
	"testing"

	"github.com/google/uuid"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/tests/integration/testdb"
)

func TestIsEligibleApprover(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	approver := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	stranger := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	// Document under review (non-draft status satisfies snapshot constraints via
	// NewDocument's internal stub path).
	doc := testdb.NewDocument(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithOwner(approver.ID),
		testdb.WithStatus("under_review"),
	)

	// A second document in the same tenant — IsEligibleApprover must return false
	// for this docID even though the approver is eligible on the first document.
	otherDoc := testdb.NewDocument(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithOwner(approver.ID),
		testdb.WithStatus("under_review"),
	)

	// Seed an in_progress approval_instance for doc. NewApprovalInstance accepts
	// WithStatus to set the status column directly (document.submit tripwire).
	// No explicit route: letting NewApprovalInstance auto-wire one resolves the
	// route's profile_code from the document's controlled-doc via its internal
	// profileForCD lookup, satisfying approval_routes_document_profile_fk.
	inst := testdb.NewApprovalInstance(t, db,
		testdb.WithDocument(doc),
		testdb.WithStatus("in_progress"),
	)

	// Seed the open (pending) approval_stage_instance referencing this instance.
	// eligible_actor_ids is a JSONB array of user-ID strings.
	// No dedicated factory for stage instances — insert directly.
	// approval_stage_instances carries no authz tripwire of its own; the parent
	// approval_instances insert (above, via document.submit) is the governed gate.
	stageID := uuid.NewString()
	eligibleJSON := `["` + approver.ID + `"]`
	if _, err := db.ExecContext(ctx,
		`INSERT INTO public.approval_stage_instances
		   (id, approval_instance_id, stage_order, name_snapshot,
		    required_role_snapshot, required_capability_snapshot, area_code_snapshot,
		    quorum_snapshot, on_eligibility_drift_snapshot, eligible_actor_ids, status)
		 VALUES ($1::uuid, $2::uuid, 1, 'Stage 1',
		         'approver', 'document.signoff', 'area-001',
		         'any_1_of', 'keep_snapshot', $3::jsonb, 'pending')`,
		stageID, inst.ID, eligibleJSON,
	); err != nil {
		t.Fatalf("seed approval_stage_instances: %v", err)
	}

	repo := infrastructure.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	t.Run("EligibleApprover_True", func(t *testing.T) {
		got, err := repo.IsEligibleApprover(ctx, tnt.ID, doc.ID, approver.ID)
		if err != nil {
			t.Fatalf("IsEligibleApprover(approver): %v", err)
		}
		if !got {
			t.Fatal("IsEligibleApprover(approver) = false, want true")
		}
	})

	t.Run("Stranger_False", func(t *testing.T) {
		got, err := repo.IsEligibleApprover(ctx, tnt.ID, doc.ID, stranger.ID)
		if err != nil {
			t.Fatalf("IsEligibleApprover(stranger): %v", err)
		}
		if got {
			t.Fatal("IsEligibleApprover(stranger) = true, want false")
		}
	})

	t.Run("WrongDoc_False", func(t *testing.T) {
		got, err := repo.IsEligibleApprover(ctx, tnt.ID, otherDoc.ID, approver.ID)
		if err != nil {
			t.Fatalf("IsEligibleApprover(otherDoc): %v", err)
		}
		if got {
			t.Fatal("IsEligibleApprover(approver, otherDoc) = true, want false")
		}
	})
}
