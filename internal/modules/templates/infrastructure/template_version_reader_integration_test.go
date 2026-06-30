//go:build integration
// +build integration

package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"

	"metaldocs/internal/modules/templates/infrastructure"
	"metaldocs/tests/integration/testdb"
)

// TestTemplateVersionReader_GetTemplateVersionState_Live proves the raw-state
// primitive returns the *raw* status string (not a published bool), scoped to
// tenant, with the owning template's doc_type_code — the contract F4.2's
// controlled-documents override resolution and profile-default adapter depend on.
func TestTemplateVersionReader_GetTemplateVersionState_Live(t *testing.T) {
	ctx := context.Background()
	db, dbName := testdb.Open(t)

	// Bare runtime tables must resolve to public.* (templates_*); evict the
	// connection that ran the ALTER so the pool reopens with the new default.
	if _, err := db.ExecContext(ctx, `ALTER DATABASE "`+dbName+`" SET search_path TO public, metaldocs`); err != nil {
		t.Fatalf("alter database search_path: %v", err)
	}
	db.SetMaxIdleConns(0)
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(4)

	tnt := testdb.NewTenant(t, db)
	actorID := testdb.DeterministicID(t, "tvs-actor")
	templateID := testdb.DeterministicID(t, "tvs-template")
	publishedVersionID := testdb.DeterministicID(t, "tvs-version-published")
	obsoleteVersionID := testdb.DeterministicID(t, "tvs-version-obsolete")
	absentVersionID := testdb.DeterministicID(t, "tvs-version-absent")

	// templates_template and templates_template_version carry the template.create
	// authz tripwire; assert the cap tx-locally via SeedWithCaps (pool-safe,
	// mirrors production — assertion is discarded on commit, never leaks).
	testdb.SeedWithCaps(t, db, `[{"cap":"template.create"}]`, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.templates_template (
				id, tenant_id, doc_type_code, key, name, latest_version, published_version_id, created_by
			) VALUES (
				$1::uuid, $2::uuid, 'po', 'tvs-template', 'TVS Template', 1, NULL, $3
			)`,
			templateID, tnt.ID, actorID,
		); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.templates_template_version (
				id, tenant_id, template_id, version_number, revision_number, status, docx_storage_key, content_hash,
				metadata_schema, placeholder_schema, author_id, published_at
			) VALUES (
				$1::uuid, $4::uuid, $2::uuid, 1, 0, 'published', 'templates/tvs/body.docx', 'body-hash-1',
				'{}'::jsonb, '{"placeholders":[]}'::jsonb, $3, now()
			)`,
			publishedVersionID, templateID, actorID, tnt.ID,
		); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.templates_template_version (
				id, tenant_id, template_id, version_number, revision_number, status, docx_storage_key, content_hash,
				metadata_schema, placeholder_schema, author_id, published_at
			) VALUES (
				$1::uuid, $4::uuid, $2::uuid, 2, 1, 'obsolete', 'templates/tvs/body2.docx', 'body-hash-2',
				'{}'::jsonb, '{"placeholders":[]}'::jsonb, $3, now()
			)`,
			obsoleteVersionID, templateID, actorID, tnt.ID,
		); err != nil {
			return err
		}
		return nil
	})

	reader := infrastructure.NewTemplateVersionReader(db)

	t.Run("published_returns_raw_status_and_doctype", func(t *testing.T) {
		status, docType, err := reader.GetTemplateVersionState(ctx, tnt.ID, publishedVersionID)
		if err != nil {
			t.Fatalf("GetTemplateVersionState: %v", err)
		}
		if status == nil || *status != "published" {
			t.Fatalf("status = %v, want \"published\"", status)
		}
		if docType != "po" {
			t.Fatalf("docType = %q, want \"po\"", docType)
		}
	})

	t.Run("obsolete_returns_raw_status_not_bool", func(t *testing.T) {
		status, docType, err := reader.GetTemplateVersionState(ctx, tnt.ID, obsoleteVersionID)
		if err != nil {
			t.Fatalf("GetTemplateVersionState: %v", err)
		}
		// The whole point of the raw-state primitive: a non-published status is
		// returned verbatim, not collapsed to false.
		if status == nil || *status != "obsolete" {
			t.Fatalf("status = %v, want \"obsolete\"", status)
		}
		if docType != "po" {
			t.Fatalf("docType = %q, want \"po\"", docType)
		}
	})

	t.Run("absent_returns_nil_empty_nil", func(t *testing.T) {
		status, docType, err := reader.GetTemplateVersionState(ctx, tnt.ID, absentVersionID)
		if err != nil {
			t.Fatalf("GetTemplateVersionState: %v", err)
		}
		if status != nil || docType != "" {
			t.Fatalf("absent = (%v, %q), want (nil, \"\")", status, docType)
		}
	})

	t.Run("other_tenant_returns_nil", func(t *testing.T) {
		otherTnt := testdb.NewTenant(t, db)
		status, docType, err := reader.GetTemplateVersionState(ctx, otherTnt.ID, publishedVersionID)
		if err != nil {
			t.Fatalf("GetTemplateVersionState: %v", err)
		}
		if status != nil || docType != "" {
			t.Fatalf("other-tenant = (%v, %q), want (nil, \"\")", status, docType)
		}
	})
}
