//go:build integration
// +build integration

package testdb

import (
	"context"
	"database/sql"
	"testing"
)

// InsertDraftDocument inserts a minimal draft document with an initial revision
// into the test schema. Returns (docID, tenantID).
// tenantID must be a valid UUID string.
func InsertDraftDocument(t *testing.T, db *sql.DB, schema, tenantID string) (docID, tenant string) {
	t.Helper()
	ctx := context.Background()

	userID := DeterministicID(t, "user")
	templateKey := "test-template-" + randomSuffix(t)

	// Insert minimal stub template.
	var tplID string
	if err := db.QueryRowContext(ctx,
		`INSERT INTO `+Qualified(schema, "templates")+
			` (tenant_id, key, name, created_by)
		 VALUES ($1::uuid, $2, 'Test Template', $3::uuid)
		 RETURNING id::text`,
		tenantID, templateKey, userID,
	).Scan(&tplID); err != nil {
		t.Fatalf("InsertDraftDocument: insert template: %v", err)
	}

	// Insert minimal published template version.
	var tvID string
	if err := db.QueryRowContext(ctx,
		`INSERT INTO `+Qualified(schema, "template_versions")+
			` (template_id, version_num, status, docx_storage_key, schema_storage_key,
			   docx_content_hash, schema_content_hash, created_by)
		 VALUES ($1::uuid, 1, 'published', 'key/tpl.docx', 'key/schema.json',
			   'aabbcc', 'ddeeff', $2::uuid)
		 RETURNING id::text`,
		tplID, userID,
	).Scan(&tvID); err != nil {
		t.Fatalf("InsertDraftDocument: insert template_version: %v", err)
	}

	// Insert document.
	if err := db.QueryRowContext(ctx,
		`INSERT INTO `+Qualified(schema, "documents")+
			` (tenant_id, template_version_id, name, status, form_data_json, created_by)
		 VALUES ($1::uuid, $2::uuid, 'Test Doc', 'draft', '{}', $3::uuid)
		 RETURNING id::text`,
		tenantID, tvID, userID,
	).Scan(&docID); err != nil {
		t.Fatalf("InsertDraftDocument: insert document: %v", err)
	}

	// Insert editor session.
	var sessID string
	if err := db.QueryRowContext(ctx,
		`INSERT INTO `+Qualified(schema, "editor_sessions")+
			` (tenant_id, document_id, user_id, expires_at, last_acknowledged_revision_id, status)
		 VALUES ($1::uuid, $2::uuid, $3::uuid, now() + interval '1 hour',
		         '00000000-0000-0000-0000-000000000000', 'active')
		 RETURNING id::text`,
		tenantID, docID, userID,
	).Scan(&sessID); err != nil {
		t.Fatalf("InsertDraftDocument: insert session: %v", err)
	}

	// Insert initial revision.
	var revID string
	if err := db.QueryRowContext(ctx,
		`INSERT INTO `+Qualified(schema, "document_revisions")+
			` (document_id, parent_revision_id, session_id, storage_key, content_hash, form_data_snapshot)
		 VALUES ($1::uuid, NULL, $2::uuid, '', 'aabbcc', '{}')
		 RETURNING id::text`,
		docID, sessID,
	).Scan(&revID); err != nil {
		t.Fatalf("InsertDraftDocument: insert revision: %v", err)
	}

	// Update document pointers.
	if _, err := db.ExecContext(ctx,
		`UPDATE `+Qualified(schema, "documents")+
			` SET current_revision_id=$1::uuid, active_session_id=$2::uuid, updated_at=now()
		 WHERE id=$3::uuid`,
		revID, sessID, docID,
	); err != nil {
		t.Fatalf("InsertDraftDocument: update document pointers: %v", err)
	}

	// Update session ack.
	if _, err := db.ExecContext(ctx,
		`UPDATE `+Qualified(schema, "editor_sessions")+
			` SET last_acknowledged_revision_id=$1::uuid WHERE id=$2::uuid AND tenant_id=$3::uuid`,
		revID, sessID, tenantID,
	); err != nil {
		t.Fatalf("InsertDraftDocument: update session ack: %v", err)
	}

	return docID, tenantID
}

// seedWithCaps runs fn inside its own transaction with the given capabilities
// asserted transaction-locally (set_config third arg true), exactly as the
// production authz layer asserts them (authz.appendAssertedCap). Doing the
// set_config and the governed write in one transaction makes the helper safe
// against connection pools of any size: the tripwire reads the caps on the same
// connection that performs the write, and the assertion is discarded on commit
// so it never leaks into the caller's session. capsJSON is the JSONB array
// literal, e.g. `[{"cap":"taxonomy.manage"}]`.
func seedWithCaps(t *testing.T, db *sql.DB, capsJSON string, fn func(tx *sql.Tx) error) {
	t.Helper()
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seedWithCaps: begin tx: %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', $1, true)`, capsJSON,
	); err != nil {
		t.Fatalf("seedWithCaps: assert caps %s: %v", capsJSON, err)
	}
	if err := fn(tx); err != nil {
		t.Fatalf("seedWithCaps: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("seedWithCaps: commit: %v", err)
	}
}

// SeedGovernedTaxonomy seeds the governed FK-parent chain required before any
// controlled_documents insert: a shared document_families row, plus the
// (tenantID, processAreaCode) document_process_areas row and the
// (tenantID, profileCode) document_profiles row the controlled document
// references. All three tables carry the metaldocs authz tripwire
// (trg_require_cap_asserted -> taxonomy.manage), so the writes run inside a
// transaction that asserts taxonomy.manage transaction-locally (pool-safe; the
// assertion never leaks into the caller's session).
//
// Idempotent (ON CONFLICT DO NOTHING) so repeated calls within a test are safe.
func SeedGovernedTaxonomy(t *testing.T, db *sql.DB, tenantID, profileCode, processAreaCode string) {
	t.Helper()
	ctx := context.Background()

	const familyCode = "test-family"

	seedWithCaps(t, db, `[{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO `+Qualified("", "document_families")+` (code, name)
			 VALUES ($1, $1)
			 ON CONFLICT (code) DO NOTHING`,
			familyCode,
		); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO `+Qualified("", "document_process_areas")+` (code, tenant_id, name)
			 VALUES ($1, $2::uuid, $1)
			 ON CONFLICT (tenant_id, code) DO NOTHING`,
			processAreaCode, tenantID,
		); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO `+Qualified("", "document_profiles")+`
			    (code, tenant_id, family_code, name, review_interval_days, alias)
			 VALUES ($1, $2::uuid, $3, $1, 365, $1)
			 ON CONFLICT (tenant_id, code) DO NOTHING`,
			profileCode, tenantID, familyCode,
		); err != nil {
			return err
		}
		return nil
	})
}

// SupersedeActiveDocumentForCD walks the single active document of the given
// controlled document through the legal status lifecycle to 'superseded', so a
// successor revision can be created. ux_documents_cd_active permits only one
// active document (draft/under_review/approved/rejected/scheduled) per CD, and
// trg_documents_legal_transition allows no shortcut out of 'draft' — the only
// exit is the full path draft -> under_review -> approved -> published ->
// superseded. This mirrors production: a prior revision is published and then
// superseded before its successor exists. enforce_snapshot_on_submit requires
// the six snapshot columns for under_review/approved/published, so they are
// stubbed (32-byte hashes satisfy the *_hash_len checks).
//
// Each UPDATE carries the documents-UPDATE tripwire (document.edit), asserted
// transaction-locally so the helper is pool-safe.
func SupersedeActiveDocumentForCD(t *testing.T, db *sql.DB, controlledDocumentID string) {
	t.Helper()
	ctx := context.Background()

	seedWithCaps(t, db, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		var docID string
		if err := tx.QueryRowContext(ctx,
			`SELECT id::text FROM `+Qualified("", "documents")+`
			  WHERE controlled_document_id = $1::uuid
			    AND status = ANY (ARRAY['draft','under_review','approved','rejected','scheduled'])`,
			controlledDocumentID,
		).Scan(&docID); err != nil {
			return err
		}

		// Stub the snapshot columns (still 'draft', so no transition / no snapshot
		// gate yet), then walk the legal transition graph to 'superseded'.
		if _, err := tx.ExecContext(ctx,
			`UPDATE `+Qualified("", "documents")+` SET
			    placeholder_schema_snapshot = '{}'::jsonb,
			    placeholder_schema_hash     = decode(repeat('ab',32),'hex'),
			    composition_config_snapshot = '{}'::jsonb,
			    composition_config_hash     = decode(repeat('cd',32),'hex'),
			    body_docx_snapshot_s3_key   = 'tenants/test/superseded/body.docx',
			    body_docx_hash              = decode(repeat('ef',32),'hex')
			  WHERE id = $1::uuid`,
			docID,
		); err != nil {
			return err
		}
		for _, status := range []string{"under_review", "approved", "published", "superseded"} {
			if _, err := tx.ExecContext(ctx,
				`UPDATE `+Qualified("", "documents")+` SET status = $2 WHERE id = $1::uuid`,
				docID, status,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

// SeedSystemAdmin seeds the actor as a tenant-scoped system_admin so the
// application-level authz check inside the create path (authz.Require, ADR 0022
// tier-2) is satisfied via the system_admin inheritance bypass — without
// granting per-area memberships. It upserts the iam_users row (display_name is
// preserved for the iam UserDisplayNameReader port) and the iam_user_roles row.
// iam_users carries no tripwire; iam_user_roles requires user.manage, so the
// role write runs inside a user.manage transaction (pool-safe, tx-local).
func SeedSystemAdmin(t *testing.T, db *sql.DB, tenantID, userID, displayName string) {
	t.Helper()
	ctx := context.Background()

	// iam_users.tenant_id FK -> metaldocs.tenants; seed the tenant parent first.
	// tenants carries no tripwire. slug = tenantID guarantees uniqueness/non-blank.
	// ON CONFLICT (id) DO NOTHING leaves an already-seeded reference tenant (e.g. the
	// dev tenant) untouched.
	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.tenants (id, name, slug)
		 VALUES ($1::uuid, $2, $1::text)
		 ON CONFLICT (id) DO NOTHING`,
		tenantID, "Test Tenant "+tenantID,
	); err != nil {
		t.Fatalf("SeedSystemAdmin: tenants: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO `+Qualified("", "iam_users")+` (user_id, display_name, tenant_id)
		 VALUES ($1, $2, $3::uuid)
		 ON CONFLICT (user_id) DO UPDATE SET display_name = EXCLUDED.display_name, tenant_id = EXCLUDED.tenant_id`,
		userID, displayName, tenantID,
	); err != nil {
		t.Fatalf("SeedSystemAdmin: iam_users: %v", err)
	}

	seedWithCaps(t, db, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO `+Qualified("", "iam_user_roles")+` (user_id, role_code, tenant_id, assigned_by)
			 VALUES ($1, 'system_admin', $2::uuid, $1)
			 ON CONFLICT (tenant_id, user_id) DO UPDATE SET role_code = EXCLUDED.role_code, assigned_by = EXCLUDED.assigned_by`,
			userID, tenantID,
		)
		return err
	})
}
