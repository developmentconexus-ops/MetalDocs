//go:build integration
// +build integration

package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"

	"metaldocs/internal/modules/templates/infrastructure"
	"metaldocs/internal/platform/tenant"
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

	tenantID := tenant.DevTenantID
	actorID := testdb.DeterministicID(t, "tvs-actor")
	templateID := testdb.DeterministicID(t, "tvs-template")
	publishedVersionID := testdb.DeterministicID(t, "tvs-version-published")
	obsoleteVersionID := testdb.DeterministicID(t, "tvs-version-obsolete")
	absentVersionID := testdb.DeterministicID(t, "tvs-version-absent")

	seedTemplateVersionStateRows(t, ctx, db, tenantID, actorID, templateID, publishedVersionID, obsoleteVersionID)

	reader := infrastructure.NewTemplateVersionReader(db)

	t.Run("published_returns_raw_status_and_doctype", func(t *testing.T) {
		status, docType, err := reader.GetTemplateVersionState(ctx, tenantID, publishedVersionID)
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
		status, docType, err := reader.GetTemplateVersionState(ctx, tenantID, obsoleteVersionID)
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
		status, docType, err := reader.GetTemplateVersionState(ctx, tenantID, absentVersionID)
		if err != nil {
			t.Fatalf("GetTemplateVersionState: %v", err)
		}
		if status != nil || docType != "" {
			t.Fatalf("absent = (%v, %q), want (nil, \"\")", status, docType)
		}
	})

	t.Run("other_tenant_returns_nil", func(t *testing.T) {
		otherTenant := "11111111-1111-1111-1111-111111111111"
		status, docType, err := reader.GetTemplateVersionState(ctx, otherTenant, publishedVersionID)
		if err != nil {
			t.Fatalf("GetTemplateVersionState: %v", err)
		}
		if status != nil || docType != "" {
			t.Fatalf("other-tenant = (%v, %q), want (nil, \"\")", status, docType)
		}
	})
}

func seedTemplateVersionStateRows(t *testing.T, ctx context.Context, db *sql.DB, tenantID, actorID, templateID, publishedVersionID, obsoleteVersionID string) {
	t.Helper()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seedTemplateVersionStateRows: begin tx: %v", err)
	}
	defer tx.Rollback()

	// templates_template(_version) carry the template.create authz tripwire; assert
	// the cap tx-locally so the seed writes are pool-safe (mirrors production).
	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', '[{"cap":"template.create"}]', true)`,
	); err != nil {
		t.Fatalf("seedTemplateVersionStateRows: assert caps: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO templates_template (
			id, tenant_id, doc_type_code, key, name, latest_version, published_version_id, created_by
		) VALUES (
			$1::uuid, $2, 'po', 'tvs-template', 'TVS Template', 1, NULL, $3
		)`,
		templateID, tenantID, actorID,
	); err != nil {
		t.Fatalf("seed templates_template: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO templates_template_version (
			id, template_id, version_number, revision_number, status, docx_storage_key, content_hash,
			metadata_schema, placeholder_schema, author_id, published_at
		) VALUES (
			$1::uuid, $2::uuid, 1, 0, 'published', 'templates/tvs/body.docx', 'body-hash-1',
			'{}'::jsonb, '{"placeholders":[]}'::jsonb, $3, now()
		)`,
		publishedVersionID, templateID, actorID,
	); err != nil {
		t.Fatalf("seed published version: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO templates_template_version (
			id, template_id, version_number, revision_number, status, docx_storage_key, content_hash,
			metadata_schema, placeholder_schema, author_id, published_at
		) VALUES (
			$1::uuid, $2::uuid, 2, 1, 'obsolete', 'templates/tvs/body2.docx', 'body-hash-2',
			'{}'::jsonb, '{"placeholders":[]}'::jsonb, $3, now()
		)`,
		obsoleteVersionID, templateID, actorID,
	); err != nil {
		t.Fatalf("seed obsolete version: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("seedTemplateVersionStateRows: commit: %v", err)
	}
}
