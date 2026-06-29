//go:build integration
// +build integration

package repository_test

import (
	"context"
	"testing"

	"metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

const guardTenantID = tenant.DevTenantID

// TestFillInRepository_UpsertAuthorValue_GovernedSourceBlocked asserts that
// UpsertAuthorValue returns 0 rows affected when the existing row carries a
// governed source ('dictionary' or 'computed'), and leaves the row unchanged.
// This is the DB enforcement layer of SP-2 D11.
func TestFillInRepository_UpsertAuthorValue_GovernedSourceBlocked(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)
	db.SetMaxOpenConns(1)

	testdb.SetCapsOnDB(t, db, `[{"cap":"document.create"},{"cap":"document.edit"}]`)

	docID, tid := testdb.InsertDraftDocument(t, db, schema, guardTenantID)

	var revID string
	if err := db.QueryRowContext(ctx,
		`SELECT current_revision_id::text FROM `+testdb.Qualified(schema, "documents")+` WHERE id=$1::uuid`,
		docID,
	).Scan(&revID); err != nil {
		t.Fatalf("get revision: %v", err)
	}

	repo := repository.NewFillInRepository(db)

	// Seed a governed row with source='dictionary'.
	originalText := "governed-value"
	governed := repository.PlaceholderValue{
		TenantID:      tid,
		RevisionID:    revID,
		PlaceholderID: "ph-governed",
		ValueText:     &originalText,
		Source:        "dictionary",
	}
	if err := repo.UpsertValue(ctx, governed); err != nil {
		t.Fatalf("seed governed row: %v", err)
	}

	// Attempt an author write — must return 0 rows affected.
	authorText := "author-hack"
	affected, err := repo.UpsertAuthorValue(ctx, repository.PlaceholderValue{
		TenantID:      tid,
		RevisionID:    revID,
		PlaceholderID: "ph-governed",
		ValueText:     &authorText,
		Source:        "user",
	})
	if err != nil {
		t.Fatalf("UpsertAuthorValue: unexpected error: %v", err)
	}
	if affected != 0 {
		t.Fatalf("D11 DB guard: want 0 rows affected, got %d", affected)
	}

	// Confirm the row is unchanged.
	src, exists, err := repo.CurrentSource(ctx, tid, revID, "ph-governed")
	if err != nil {
		t.Fatalf("CurrentSource: %v", err)
	}
	if !exists {
		t.Fatal("row unexpectedly absent")
	}
	if src != "dictionary" {
		t.Fatalf("D11: source must remain 'dictionary', got %q", src)
	}

	// Confirm value_text is also unchanged (row not mutated even partially).
	vals, err := repo.ListValues(ctx, tid, revID)
	if err != nil {
		t.Fatalf("ListValues: %v", err)
	}
	if len(vals) != 1 {
		t.Fatalf("expected 1 row, got %d", len(vals))
	}
	if vals[0].ValueText == nil || *vals[0].ValueText != originalText {
		t.Fatalf("D11: value_text must remain %q, got %v", originalText, vals[0].ValueText)
	}
}

// TestFillInRepository_UpsertAuthorValue_DefaultSourceAllowed asserts that
// UpsertAuthorValue successfully flips a source='default' row to source='user'.
// This is the no-regression case: author writes on author-editable rows still work.
func TestFillInRepository_UpsertAuthorValue_DefaultSourceAllowed(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)
	db.SetMaxOpenConns(1)

	testdb.SetCapsOnDB(t, db, `[{"cap":"document.create"},{"cap":"document.edit"}]`)

	docID, tid := testdb.InsertDraftDocument(t, db, schema, guardTenantID)

	var revID string
	if err := db.QueryRowContext(ctx,
		`SELECT current_revision_id::text FROM `+testdb.Qualified(schema, "documents")+` WHERE id=$1::uuid`,
		docID,
	).Scan(&revID); err != nil {
		t.Fatalf("get revision: %v", err)
	}

	repo := repository.NewFillInRepository(db)

	// Seed a row with source='default'.
	defaultText := "default-value"
	defaultRow := repository.PlaceholderValue{
		TenantID:      tid,
		RevisionID:    revID,
		PlaceholderID: "ph-default",
		ValueText:     &defaultText,
		Source:        "default",
	}
	if err := repo.UpsertValue(ctx, defaultRow); err != nil {
		t.Fatalf("seed default row: %v", err)
	}

	// Author write on a 'default' row must succeed (1 row affected).
	authorText := "author-edit"
	affected, err := repo.UpsertAuthorValue(ctx, repository.PlaceholderValue{
		TenantID:      tid,
		RevisionID:    revID,
		PlaceholderID: "ph-default",
		ValueText:     &authorText,
		Source:        "user",
	})
	if err != nil {
		t.Fatalf("UpsertAuthorValue on default row: %v", err)
	}
	if affected == 0 {
		t.Fatal("author write on default row must succeed (>0 rows affected)")
	}

	// Confirm the row is updated to source='user' with new value.
	src, exists, err := repo.CurrentSource(ctx, tid, revID, "ph-default")
	if err != nil {
		t.Fatalf("CurrentSource: %v", err)
	}
	if !exists {
		t.Fatal("row unexpectedly absent")
	}
	if src != "user" {
		t.Fatalf("row source must be 'user' after author write, got %q", src)
	}
	vals, err := repo.ListValues(ctx, tid, revID)
	if err != nil {
		t.Fatalf("ListValues: %v", err)
	}
	if len(vals) != 1 {
		t.Fatalf("expected 1 row, got %d", len(vals))
	}
	if vals[0].ValueText == nil || *vals[0].ValueText != authorText {
		t.Fatalf("value_text must be updated to %q, got %v", authorText, vals[0].ValueText)
	}
}

// TestFillInRepository_UpsertAuthorValue_NewRow asserts that UpsertAuthorValue
// creates a new row (no pre-existing conflict) and reports 1 row affected.
func TestFillInRepository_UpsertAuthorValue_NewRow(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)
	db.SetMaxOpenConns(1)

	testdb.SetCapsOnDB(t, db, `[{"cap":"document.create"},{"cap":"document.edit"}]`)

	docID, tid := testdb.InsertDraftDocument(t, db, schema, guardTenantID)

	var revID string
	if err := db.QueryRowContext(ctx,
		`SELECT current_revision_id::text FROM `+testdb.Qualified(schema, "documents")+` WHERE id=$1::uuid`,
		docID,
	).Scan(&revID); err != nil {
		t.Fatalf("get revision: %v", err)
	}

	repo := repository.NewFillInRepository(db)

	newText := "brand-new"
	affected, err := repo.UpsertAuthorValue(ctx, repository.PlaceholderValue{
		TenantID:      tid,
		RevisionID:    revID,
		PlaceholderID: "ph-new",
		ValueText:     &newText,
		Source:        "user",
	})
	if err != nil {
		t.Fatalf("UpsertAuthorValue new row: %v", err)
	}
	if affected != 1 {
		t.Fatalf("new row insert: want 1 rows affected, got %d", affected)
	}

	src, exists, err := repo.CurrentSource(ctx, tid, revID, "ph-new")
	if err != nil {
		t.Fatalf("CurrentSource: %v", err)
	}
	if !exists || src != "user" {
		t.Fatalf("new row: want exists=true source='user', got exists=%v source=%q", exists, src)
	}
}
