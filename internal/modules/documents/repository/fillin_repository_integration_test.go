//go:build integration
// +build integration

package repository_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

const fillInTenantID = tenant.DevTenantID

func TestFillInRepository_UpsertValueAndListValues(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)

	docID, tenant := testdb.InsertDraftDocument(t, db, schema, fillInTenantID)

	var revID string
	if err := db.QueryRowContext(ctx,
		`SELECT current_revision_id::text FROM `+testdb.Qualified(schema, "documents")+` WHERE id=$1::uuid`,
		docID,
	).Scan(&revID); err != nil {
		t.Fatalf("get revision: %v", err)
	}

	repo := repository.NewFillInRepositoryWithSchema(db, schema)

	v1 := repository.PlaceholderValue{
		TenantID:      tenant,
		RevisionID:    revID,
		PlaceholderID: "ph-1",
		ValueText:     strPtr("A"),
		ValueTyped:    map[string]any{"raw": "A"},
		Source:        "user",
	}
	if err := repo.UpsertValue(ctx, v1); err != nil {
		t.Fatalf("upsert first: %v", err)
	}

	var createdAt, updatedAt time.Time
	if err := db.QueryRowContext(ctx,
		`SELECT created_at, updated_at FROM `+testdb.Qualified(schema, "document_placeholder_values")+`
		  WHERE tenant_id=$1::uuid AND revision_id=$2::uuid AND placeholder_id=$3`,
		tenant, revID, "ph-1",
	).Scan(&createdAt, &updatedAt); err != nil {
		t.Fatalf("timestamps first: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	v2 := repository.PlaceholderValue{
		TenantID:      tenant,
		RevisionID:    revID,
		PlaceholderID: "ph-1",
		ValueText:     strPtr("B"),
		ValueTyped:    map[string]any{"raw": "B", "n": float64(2)},
		Source:        "user",
	}
	if err := repo.UpsertValue(ctx, v2); err != nil {
		t.Fatalf("upsert second: %v", err)
	}

	var createdAt2, updatedAt2 time.Time
	var typedJSON []byte
	if err := db.QueryRowContext(ctx,
		`SELECT created_at, updated_at, value_typed FROM `+testdb.Qualified(schema, "document_placeholder_values")+`
		  WHERE tenant_id=$1::uuid AND revision_id=$2::uuid AND placeholder_id=$3`,
		tenant, revID, "ph-1",
	).Scan(&createdAt2, &updatedAt2, &typedJSON); err != nil {
		t.Fatalf("timestamps second: %v", err)
	}
	if !createdAt2.Equal(createdAt) {
		t.Fatalf("created_at changed: first=%v second=%v", createdAt, createdAt2)
	}
	if !updatedAt2.After(updatedAt) {
		t.Fatalf("updated_at did not advance: first=%v second=%v", updatedAt, updatedAt2)
	}

	values, err := repo.ListValues(ctx, tenant, revID)
	if err != nil {
		t.Fatalf("list values: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}
	if values[0].PlaceholderID != "ph-1" {
		t.Fatalf("placeholder_id = %q, want ph-1", values[0].PlaceholderID)
	}
	if values[0].ValueText == nil || *values[0].ValueText != "B" {
		t.Fatalf("value_text = %v, want B", values[0].ValueText)
	}
	if values[0].Source != "user" {
		t.Fatalf("source = %q, want user", values[0].Source)
	}

	var typed map[string]any
	if err := json.Unmarshal(typedJSON, &typed); err != nil {
		t.Fatalf("unmarshal typed json: %v", err)
	}
	if typed["raw"] != "B" {
		t.Fatalf("typed raw = %v, want B", typed["raw"])
	}
}

func strPtr(v string) *string { return &v }
