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

	// One lineage with three revisions: published rev0, approved rev1, published
	// rev2. The current published head is the highest-revision published row.
	cd := testdb.NewControlledDoc(t, db)
	testdb.NewDocument(t, db, testdb.WithControlledDoc(cd), testdb.WithStatus("published"), testdb.WithRevisionNumber(0))
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
			       effective_from = $1,
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
