//go:build integration
// +build integration

package repository_test

// M6 F6.2 — DB-CHECK integration proof for the eQMS review/expiry + reason
// columns added by db/migrations/0274_document_review_and_reason.sql.
//
// Per the M6 validation contract §2.1 exit criteria, the review/expiry CHECKs
// must be enforced by the DATABASE, not merely by app code. This test seeds a
// valid base document row via the same fixture path as
// repository_create_integration_test.go, then drives raw UPDATEs against
// public.documents (satisfying the documents/UPDATE tripwire with document.edit)
// to prove each CHECK:
//
//   1. a valid review/expiry window UPDATE succeeds;
//   2. effective_to <= effective_from is rejected by ck_documents_effective_window;
//   3. an out-of-enum reason_category is rejected by ck_documents_reason_category;
//   4. review_due_at < effective_from is rejected by ck_documents_review_due_sane.
//
// The UPDATEs are status-preserving (they never touch documents.status), so the
// enforce_document_transition() trigger is a no-op and only the new CHECKs gate
// the write. The documents/UPDATE tripwire (enforce_capability_asserted) still
// requires an asserted cap in {document.edit, document.obsolete, membership.manage}
// (migration 0271), satisfied here with document.edit via testdb.SeedWithCaps.

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/documents/repository"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/tests/integration/testdb"
)

// seedReviewColumnsDoc creates one valid base document row (via the repo create
// path, exactly like repository_create_integration_test.go) and returns its id.
func seedReviewColumnsDoc(t *testing.T, db *sql.DB) (docID string) {
	t.Helper()
	ctx := context.Background()

	tnt := testdb.NewTenant(t, db)
	actor := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithTaxonomy(tax),
		testdb.WithOwner(actor.ID),
	)

	repo := repository.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	profileCode := cd.ProfileCode
	processAreaCode := cd.ProcessAreaCode
	controlledDocID := cd.ID
	doc := &domain.Document{
		TenantID:                tnt.ID,
		TemplateVersionID:       testdb.DeterministicID(t, "template-version"),
		Name:                    "Review Columns Doc",
		FormDataJSON:            []byte(`{}`),
		CreatedBy:               actor.ID,
		ControlledDocumentID:    &controlledDocID,
		ProfileCodeSnapshot:     &profileCode,
		ProcessAreaCodeSnapshot: &processAreaCode,
		Code:                    "PO-REVIEW-001",
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"document.create"},{"cap":"document.edit"}]`)
	newDocID, _, _, err := repo.CreateDocumentTx(ctx, tx, doc, "hash-review", "", nil)
	if err != nil {
		t.Fatalf("CreateDocumentTx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	return newDocID
}

// updateReviewColumns runs a single status-preserving UPDATE against
// public.documents inside a self-owned tx that asserts document.edit
// (satisfying the documents/UPDATE tripwire). Returns the DB error, if any.
//
// This test owns the tx directly (rather than using testdb.SeedWithCaps)
// precisely because it EXPECTS the UPDATE to fail for the CHECK cases: a failed
// ExecContext aborts the Postgres tx, and SeedWithCaps would then error on its
// unconditional Commit. Here a valid UPDATE is committed; an aborted one is
// rolled back — and either way the CHECK-violation error is returned verbatim.
func updateReviewColumns(t *testing.T, db *sql.DB, query string, args ...any) error {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("updateReviewColumns: begin tx: %v", err)
	}
	defer tx.Rollback()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"document.edit"}]`)
	if _, execErr := tx.ExecContext(ctx, query, args...); execErr != nil {
		return execErr
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("updateReviewColumns: commit valid UPDATE: %v", err)
	}
	return nil
}

// TestDocumentReviewCheckConstraints proves the 0274 CHECK constraints are
// DB-enforced (M6 contract §2.1).
func TestDocumentReviewCheckConstraints(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	docID := seedReviewColumnsDoc(t, db)

	// (1) Valid window: effective_from < effective_to, review_due_at >= effective_from,
	// reason_category in the enum. Must succeed.
	if err := updateReviewColumns(t, db,
		`UPDATE public.documents
		    SET effective_from   = timestamptz '2026-01-01 00:00:00+00',
		        effective_to     = timestamptz '2027-01-01 00:00:00+00',
		        review_due_at    = timestamptz '2026-06-01 00:00:00+00',
		        last_reviewed_at = timestamptz '2026-02-01 00:00:00+00',
		        reason_for_change = 'initial periodic review cycle set',
		        reason_category   = 'periodic_review'
		  WHERE id = $1::uuid`, docID,
	); err != nil {
		t.Fatalf("valid review/expiry window UPDATE = %v, want success", err)
	}

	// Confirm the valid values actually persisted.
	var gotReason, gotCategory string
	if err := db.QueryRowContext(context.Background(),
		`SELECT reason_for_change, reason_category FROM public.documents WHERE id = $1::uuid`, docID,
	).Scan(&gotReason, &gotCategory); err != nil {
		t.Fatalf("read back persisted row: %v", err)
	}
	if gotReason != "initial periodic review cycle set" || gotCategory != "periodic_review" {
		t.Fatalf("persisted (reason=%q, category=%q), want the valid values", gotReason, gotCategory)
	}

	// (2) effective_to <= effective_from -> ck_documents_effective_window rejects.
	// Use equal endpoints (<=) to prove the strict-greater-than boundary.
	err := updateReviewColumns(t, db,
		`UPDATE public.documents
		    SET effective_from = timestamptz '2026-01-01 00:00:00+00',
		        effective_to   = timestamptz '2026-01-01 00:00:00+00'
		  WHERE id = $1::uuid`, docID,
	)
	if err == nil || !strings.Contains(err.Error(), "ck_documents_effective_window") {
		t.Fatalf("effective_to <= effective_from UPDATE = %v, want ck_documents_effective_window violation", err)
	}

	// (3) out-of-enum reason_category -> ck_documents_reason_category rejects.
	err = updateReviewColumns(t, db,
		`UPDATE public.documents
		    SET reason_category = 'not_a_valid_category'
		  WHERE id = $1::uuid`, docID,
	)
	if err == nil || !strings.Contains(err.Error(), "ck_documents_reason_category") {
		t.Fatalf("out-of-enum reason_category UPDATE = %v, want ck_documents_reason_category violation", err)
	}

	// (4) review_due_at < effective_from -> ck_documents_review_due_sane rejects.
	// The row currently holds effective_from = 2026-01-01 (from step 1); a
	// review_due_at strictly before that must be rejected.
	err = updateReviewColumns(t, db,
		`UPDATE public.documents
		    SET review_due_at = timestamptz '2025-12-01 00:00:00+00'
		  WHERE id = $1::uuid`, docID,
	)
	if err == nil || !strings.Contains(err.Error(), "ck_documents_review_due_sane") {
		t.Fatalf("review_due_at < effective_from UPDATE = %v, want ck_documents_review_due_sane violation", err)
	}
}
