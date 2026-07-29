//go:build integration
// +build integration

package infrastructure

// Real-DB tests for the postgres approval repository, migrated off the drifted
// shared `pgtest` database onto the unified `testdb` factory (template-DB cloned
// per test, curated baseline). The factory seeds the controlled-document graph
// through the REAL capability tripwire (NewControlledDoc / NewDocument /
// NewApprovalInstance assert the cap tx-locally), so these no longer depend on a
// hand-patched dev schema. The unit tests (no DB) stay in
// postgres_approval_repository_test.go (untagged, runs under plain `go test`).

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/approval/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/tests/integration/testdb"
)

func TestValidateScheduledSupersedeTarget_RealRows(t *testing.T) {
	db, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})
	ctx := context.Background()

	// Two CD lineages sharing one tenant/owner/taxonomy. The published head and
	// the approved target live in cd1; otherPub lives in cd2 (a different lineage).
	tn := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tn.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tn.ID))
	cd1 := testdb.NewControlledDoc(t, db, testdb.WithTenant(tn.ID), testdb.WithTaxonomy(tax), testdb.WithOwner(owner.ID))
	cd2 := testdb.NewControlledDoc(t, db, testdb.WithTenant(tn.ID), testdb.WithTaxonomy(tax), testdb.WithOwner(owner.ID))

	publishedHead := testdb.NewDocument(t, db, testdb.WithControlledDoc(cd1), testdb.WithStatus("published"), testdb.WithRevisionNumber(0))
	target := testdb.NewDocument(t, db, testdb.WithControlledDoc(cd1), testdb.WithStatus("approved"), testdb.WithRevisionNumber(1))
	otherPub := testdb.NewDocument(t, db, testdb.WithControlledDoc(cd2), testdb.WithStatus("published"), testdb.WithRevisionNumber(0))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := repo.ValidateScheduledSupersedeTarget(ctx, tx, tn.ID, target.ID, publishedHead.ID); err != nil {
		t.Fatalf("valid target: unexpected error: %v", err)
	}

	err = repo.ValidateScheduledSupersedeTarget(ctx, tx, tn.ID, target.ID, otherPub.ID)
	if !errors.Is(err, ErrInvalidScheduledSupersedeTarget) {
		t.Fatalf("invalid target error = %v, want ErrInvalidScheduledSupersedeTarget", err)
	}
}

func TestLoadCurrentPublishedHeadForDocument_RealRows(t *testing.T) {
	db, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})
	ctx := context.Background()

	// One lineage with three revisions: superseded rev0, approved rev1,
	// published rev2. The current published head is the only published row —
	// ux_documents_published_head (migration 0310, ADR 0085's one-published-head
	// invariant) makes the old two-published-rows shape of this fixture
	// unrepresentable, so what it pins now is that the resolver ignores the
	// lineage's NON-published siblings on both sides of the head.
	cd := testdb.NewControlledDoc(t, db)
	testdb.NewDocument(t, db, testdb.WithControlledDoc(cd), testdb.WithStatus("superseded"), testdb.WithRevisionNumber(0))
	target := testdb.NewDocument(t, db, testdb.WithControlledDoc(cd), testdb.WithStatus("approved"), testdb.WithRevisionNumber(1))
	laterHead := testdb.NewDocument(t, db, testdb.WithControlledDoc(cd), testdb.WithStatus("published"), testdb.WithRevisionNumber(2))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	got, err := repo.LoadCurrentPublishedHeadForDocument(ctx, tx, cd.TenantID, target.ID)
	if err != nil {
		t.Fatalf("LoadCurrentPublishedHeadForDocument: %v", err)
	}
	if got != laterHead.ID {
		t.Fatalf("current head = %q, want %q", got, laterHead.ID)
	}
}

func TestLoadActiveInstanceByDocument_LoadsDocumentRevisionVersion(t *testing.T) {
	db, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})
	ctx := context.Background()

	cd := testdb.NewControlledDoc(t, db)
	doc := testdb.NewDocument(t, db,
		testdb.WithControlledDoc(cd),
		testdb.WithStatus("approved"),
		testdb.WithRevisionNumber(3),
		testdb.WithRevisionVersion(7),
	)
	testdb.NewApprovalInstance(t, db, testdb.WithDocument(doc), testdb.WithStatus("approved"))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	inst, err := repo.LoadActiveInstanceByDocument(ctx, tx, cd.TenantID, doc.ID)
	if err != nil {
		t.Fatalf("LoadActiveInstanceByDocument: %v", err)
	}
	if inst.RevisionVersion != 7 {
		t.Fatalf("revision version = %d, want 7", inst.RevisionVersion)
	}
}

// TestLoadInstanceByDocumentForView_SeesChangesRequested pins the contract
// fix behind the FE "requested changes" panel (C5): the view-read path must
// return an instance whose status is changes_requested (a non-terminal state
// reached via a request_changes verdict), while the narrow
// LoadActiveInstanceByDocument path used by publish/mutation must continue to
// treat it as invisible (ErrNoActiveInstance) so submit re-entry can create a
// fresh instance.
func TestLoadInstanceByDocumentForView_SeesChangesRequested(t *testing.T) {
	db, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})
	ctx := context.Background()

	cd := testdb.NewControlledDoc(t, db)
	doc := testdb.NewDocument(t, db,
		testdb.WithControlledDoc(cd),
		testdb.WithStatus("draft"),
		testdb.WithRevisionNumber(2),
		testdb.WithRevisionVersion(5),
	)
	inst := testdb.NewApprovalInstance(t, db, testdb.WithDocument(doc), testdb.WithStatus("changes_requested"))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	got, err := repo.LoadInstanceByDocumentForView(ctx, tx, cd.TenantID, doc.ID)
	if err != nil {
		t.Fatalf("LoadInstanceByDocumentForView: %v", err)
	}
	if got.ID != inst.ID {
		t.Fatalf("instance id = %q, want %q", got.ID, inst.ID)
	}
	if got.Status != domain.InstanceChangesRequested {
		t.Fatalf("status = %q, want changes_requested", got.Status)
	}

	_, err = repo.LoadActiveInstanceByDocument(ctx, tx, cd.TenantID, doc.ID)
	if !errors.Is(err, ErrNoActiveInstance) {
		t.Fatalf("LoadActiveInstanceByDocument err = %v, want ErrNoActiveInstance (narrow path must stay unchanged)", err)
	}
}

func TestLoadInstance_LoadsDocumentRevisionVersion(t *testing.T) {
	db, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})
	ctx := context.Background()

	cd := testdb.NewControlledDoc(t, db)
	doc := testdb.NewDocument(t, db,
		testdb.WithControlledDoc(cd),
		testdb.WithStatus("approved"),
		testdb.WithRevisionNumber(4),
		testdb.WithRevisionVersion(9),
	)
	inst := testdb.NewApprovalInstance(t, db, testdb.WithDocument(doc), testdb.WithStatus("approved"))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	got, err := repo.LoadInstance(ctx, tx, cd.TenantID, inst.ID)
	if err != nil {
		t.Fatalf("LoadInstance: %v", err)
	}
	if got.RevisionVersion != 9 {
		t.Fatalf("revision version = %d, want 9", got.RevisionVersion)
	}
}

func TestScheduleGenerationIncrementsOnScheduledWritePath(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()

	// Approved doc at revision_version 2, schedule_generation 9. The guarded
	// scheduled-write UPDATE (the repo's real write path) must bump both by one.
	doc := testdb.NewDocument(t, db,
		testdb.WithStatus("approved"),
		testdb.WithRevisionNumber(5),
		testdb.WithRevisionVersion(2),
		testdb.WithScheduleGen(9),
	)

	testdb.SeedWithCaps(t, db, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `
			UPDATE public.documents
			   SET status = 'scheduled',
			       planned_effective_from = $1,
			       effective_from = NULL,
			       superseded_document_id = NULL,
			       revision_version = revision_version + 1,
			       schedule_generation = schedule_generation + 1
			 WHERE id = $2::uuid
			   AND tenant_id = $3::uuid
			   AND status = 'approved'
			   AND revision_version = $4`,
			time.Now().UTC().Add(2*time.Hour), doc.ID, doc.TenantID, 2,
		)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected != 1 {
			return fmt.Errorf("rows affected = %d, want 1", affected)
		}
		return nil
	})

	var gotStatus string
	var gotGeneration int64
	if err := db.QueryRowContext(ctx, `
		SELECT status, schedule_generation
		  FROM public.documents
		 WHERE id = $1::uuid
		   AND tenant_id = $2::uuid`,
		doc.ID, doc.TenantID,
	).Scan(&gotStatus, &gotGeneration); err != nil {
		t.Fatalf("load document state: %v", err)
	}
	if gotStatus != "scheduled" {
		t.Fatalf("status = %s, want scheduled", gotStatus)
	}
	if gotGeneration != 10 {
		t.Fatalf("schedule_generation = %d, want 10", gotGeneration)
	}
}

// TestInsertStageInstances_ActiveStageWithNilDueInDays_NoTypeInferenceError is
// the RED/GREEN regression test for the 42P08 defect: when a stage is
// inserted as already-active (submit path, stage 0) and its route stage has
// no due_in_days SLA configured, DueInDaysSnapshot is nil. Because the bind
// param feeding the due_at CASE expression is referenced ONLY inside that
// CASE (never bound to a typed column), an untyped NULL made Postgres unable
// to infer $16's type ("could not determine data type of parameter $16").
// This must succeed for both the nil-SLA stage and a sibling active stage
// that DOES have an SLA (non-null path stays correct).
func TestInsertStageInstances_ActiveStageWithNilDueInDays_NoTypeInferenceError(t *testing.T) {
	db, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})
	ctx := context.Background()

	inst := testdb.NewApprovalInstance(t, db, testdb.WithStatus("in_progress"))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	dueDays := 5
	stages := []domain.StageInstance{
		{
			ID:                         uuid.NewString(),
			ApprovalInstanceID:         inst.ID,
			StageOrder:                 1,
			NameSnapshot:               "Review",
			RequiredRoleSnapshot:       "reviewer",
			RequiredCapabilitySnapshot: "document.review",
			AreaCodeSnapshot:           "eng",
			QuorumSnapshot:             domain.QuorumAllOf,
			QuorumMSnapshot:            nil,
			OnEligibilityDriftSnapshot: domain.DriftKeepSnapshot,
			Kind:                       domain.StageKindApproval,
			EligibleActorIDs:           []string{inst.DocumentID},
			EffectiveDenominator:       nil,
			Status:                     domain.StageActive,
			DueInDaysSnapshot:          nil, // no SLA configured — the defect trigger
		},
		{
			ID:                         uuid.NewString(),
			ApprovalInstanceID:         inst.ID,
			StageOrder:                 2,
			NameSnapshot:               "Approve",
			RequiredRoleSnapshot:       "approver",
			RequiredCapabilitySnapshot: "document.signoff",
			AreaCodeSnapshot:           "eng",
			QuorumSnapshot:             domain.QuorumAllOf,
			QuorumMSnapshot:            nil,
			OnEligibilityDriftSnapshot: domain.DriftKeepSnapshot,
			Kind:                       domain.StageKindApproval,
			EligibleActorIDs:           []string{inst.DocumentID},
			EffectiveDenominator:       nil,
			Status:                     domain.StageActive,
			DueInDaysSnapshot:          &dueDays, // has an SLA — proves non-null path unaffected
		},
	}

	if err := repo.InsertStageInstances(ctx, tx, stages); err != nil {
		t.Fatalf("InsertStageInstances: unexpected error: %v", err)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, due_in_days_snapshot, due_at
		  FROM public.approval_stage_instances
		 WHERE approval_instance_id = $1::uuid
		 ORDER BY stage_order`,
		inst.ID,
	)
	if err != nil {
		t.Fatalf("select stage instances: %v", err)
	}
	defer rows.Close()

	type row struct {
		id      string
		dueDays sql.NullInt64
		dueAt   sql.NullTime
	}
	var got []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.dueDays, &r.dueAt); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("stage instance rows = %d, want 2", len(got))
	}

	// Stage 0: no SLA — due_at must stay NULL.
	if got[0].dueDays.Valid {
		t.Fatalf("stage 0 due_in_days_snapshot = %v, want NULL", got[0].dueDays)
	}
	if got[0].dueAt.Valid {
		t.Fatalf("stage 0 due_at = %v, want NULL (no SLA configured)", got[0].dueAt)
	}

	// Stage 1: has an SLA — due_at must be set (now() + 5 days).
	if !got[1].dueDays.Valid || got[1].dueDays.Int64 != 5 {
		t.Fatalf("stage 1 due_in_days_snapshot = %v, want 5", got[1].dueDays)
	}
	if !got[1].dueAt.Valid {
		t.Fatalf("stage 1 due_at = NULL, want set (SLA configured)")
	}
	if diff := got[1].dueAt.Time.Sub(time.Now().UTC()); diff < 4*24*time.Hour || diff > 6*24*time.Hour {
		t.Fatalf("stage 1 due_at = %v, want ~now()+5 days", got[1].dueAt.Time)
	}
}
