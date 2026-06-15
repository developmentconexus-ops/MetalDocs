//go:build integration
// +build integration

package application_test

import (
	"context"
	"database/sql"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/application"
	docrepo "metaldocs/internal/modules/documents/repository"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/platform/docgenv2"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

const createSnapshotTenantID = tenant.DevTenantID

func TestCreateDocumentTx_PopulatesAllSnapshotColumns(t *testing.T) {
	ctx := context.Background()
	db, dbName := testdb.Open(t)

	// All NEW pool connections must resolve bare runtime tables to public.* (the real
	// documents / controlled_documents / templates_*); metaldocs.documents is a dead
	// legacy duplicate lacking controlled_document_id. ALTER DATABASE sets the default
	// for sessions opened after it, so evict the connection that ran the ALTER (it
	// predates the new default) and let the pool reopen. More than one connection is
	// required here: the off-tx UserDisplayNameReader port read and the templates
	// GetPublishedVersion read both run on the pool while the create tx holds a
	// connection (H-PRE-1) — a single-connection pool deadlocks.
	if _, err := db.ExecContext(ctx, `ALTER DATABASE "`+dbName+`" SET search_path TO public, metaldocs`); err != nil {
		t.Fatalf("alter database search_path: %v", err)
	}
	db.SetMaxIdleConns(0)
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(4)

	tenantID := createSnapshotTenantID
	actorID := testdb.DeterministicID(t, "actor")
	templateID := testdb.DeterministicID(t, "template")
	templateVersionID := testdb.DeterministicID(t, "template-version")
	controlledDocumentID := testdb.DeterministicID(t, "controlled-document")

	const wantDisplayName = "Snapshot Author"
	// Seed the actor as a tenant-scoped system_admin (satisfies the create path's
	// app-level authz.Require via the ADR 0022 tier-2 bypass) with display_name set,
	// so the REAL UserDisplayNameReader port resolves a value to snapshot (F4.1 value
	// proof). iam_users is not tripwire-governed; the role write is user.manage tx-local.
	testdb.SeedSystemAdmin(t, db, tenantID, actorID, wantDisplayName)

	testdb.SeedGovernedTaxonomy(t, db, tenantID, "po", "quality")
	seedCreateDocumentSnapshotRows(t, ctx, db, tenantID, actorID, templateID, templateVersionID, controlledDocumentID)

	cd := &controlleddocumentsdomain.ControlledDocument{
		ID:              controlledDocumentID,
		TenantID:        tenantID,
		ProfileCode:     "po",
		ProcessAreaCode: "quality",
		Code:            "PO-TEST-001",
		Title:           "Snapshot Test Controlled Document",
		OwnerUserID:     actorID,
		Status:          controlleddocumentsdomain.CDStatusActive,
	}

	snapshotSvc := application.NewSnapshotService(
		docgenv2.NewTemplatesSnapshotReader(db),
	)
	svc := application.NewServiceWithSnapshot(
		docrepo.New(db, iampg.NewUserDisplayNameRepository(db)),
		nil,
		docgenv2.NewTemplatesTemplateReader(db),
		fakeFormVal{valid: true},
		&noopAudit{},
		&fakeProfileDefaultTemplateReader{id: strptr(templateVersionID), status: strptr("published")},
		snapshotSvc,
	)
	initializer := application.NewCDDocumentInitializer(svc)

	req, err := controlleddocumentsdomain.NewCloneTemplateRequest(nil, "Snapshot Integration Test", nil)
	if err != nil {
		t.Fatalf("NewCloneTemplateRequest: %v", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	ref, err := initializer.CloneTemplate(ctx, tx, cd, req)
	if err != nil {
		t.Fatalf("CloneTemplate: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	var (
		placeholderSchemaSnapshot sql.NullString
		placeholderSchemaHash     []byte
		compositionConfigSnapshot sql.NullString
		compositionConfigHash     []byte
		bodyDocxSnapshotS3Key     sql.NullString
		bodyDocxHash              []byte
		profileCodeSnapshot       sql.NullString
		processAreaCodeSnapshot   sql.NullString
		createdByDisplayNameSnap  sql.NullString
	)
	if err := db.QueryRowContext(ctx, `
		SELECT placeholder_schema_snapshot::text,
		       placeholder_schema_hash,
		       composition_config_snapshot::text,
		       composition_config_hash,
		       body_docx_snapshot_s3_key,
		       body_docx_hash,
		       profile_code_snapshot,
		       process_area_code_snapshot,
		       created_by_display_name_snapshot
		  FROM documents
		 WHERE id = $1::uuid`,
		ref.ID,
	).Scan(
		&placeholderSchemaSnapshot,
		&placeholderSchemaHash,
		&compositionConfigSnapshot,
		&compositionConfigHash,
		&bodyDocxSnapshotS3Key,
		&bodyDocxHash,
		&profileCodeSnapshot,
		&processAreaCodeSnapshot,
		&createdByDisplayNameSnap,
	); err != nil {
		t.Fatalf("query snapshot columns: %v", err)
	}

	assertNotNullString(t, "placeholder_schema_snapshot", placeholderSchemaSnapshot)
	assertNotNullBytes(t, "placeholder_schema_hash", placeholderSchemaHash)
	assertNotNullString(t, "composition_config_snapshot", compositionConfigSnapshot)
	assertNotNullBytes(t, "composition_config_hash", compositionConfigHash)
	assertNotNullString(t, "body_docx_snapshot_s3_key", bodyDocxSnapshotS3Key)
	assertNotNullBytes(t, "body_docx_hash", bodyDocxHash)
	assertNotNullString(t, "profile_code_snapshot", profileCodeSnapshot)
	assertNotNullString(t, "process_area_code_snapshot", processAreaCodeSnapshot)

	// F4.1 value proof: the real UserDisplayNameReader port resolved the actor's
	// display name and CreateDocumentTx snapshotted it (read off-tx, H-PRE-1).
	assertNotNullString(t, "created_by_display_name_snapshot", createdByDisplayNameSnap)
	if createdByDisplayNameSnap.String != wantDisplayName {
		t.Fatalf("created_by_display_name_snapshot = %q, want %q", createdByDisplayNameSnap.String, wantDisplayName)
	}
}

func seedCreateDocumentSnapshotRows(t *testing.T, ctx context.Context, db *sql.DB, tenantID, actorID, templateID, templateVersionID, controlledDocumentID string) {
	t.Helper()

	// templates_template(_version) and controlled_documents carry the authz tripwire
	// (template.create / controlled_documents.create). Assert both transaction-locally
	// in one tx so the writes are pool-safe and the assertion never leaks (mirrors the
	// production authz.appendAssertedCap pattern).
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seedCreateDocumentSnapshotRows: begin tx: %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', '[{"cap":"template.create"},{"cap":"controlled_documents.create"}]', true)`,
	); err != nil {
		t.Fatalf("seedCreateDocumentSnapshotRows: assert caps: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO templates_template (
			id, tenant_id, doc_type_code, key, name, latest_version, published_version_id, created_by
		) VALUES (
			$1::uuid, $2, 'po', 'snapshot-integration-template', 'Snapshot Integration Template',
			1, NULL, $3
		)`,
		templateID, tenantID, actorID,
	); err != nil {
		t.Fatalf("seed templates_template: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO templates_template_version (
			id, template_id, version_number, status, docx_storage_key, content_hash,
			metadata_schema, placeholder_schema, author_id, published_at
		) VALUES (
			$1::uuid, $2::uuid, 1, 'published', 'templates/snapshot/body.docx', 'body-hash',
			'{}'::jsonb, '{"placeholders":[]}'::jsonb, $3, now()
		)`,
		templateVersionID, templateID, actorID,
	); err != nil {
		t.Fatalf("seed templates_template_version: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE templates_template
		   SET published_version_id = $1::uuid
		 WHERE id = $2::uuid`,
		templateVersionID, templateID,
	); err != nil {
		t.Fatalf("seed template published_version_id: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO controlled_documents (
			id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, status
		) VALUES (
			$1::uuid, $2::uuid, 'po', 'quality', 'PO-TEST-001',
			'Snapshot Test Controlled Document', $3, 'active'
		)`,
		controlledDocumentID, tenantID, actorID,
	); err != nil {
		t.Fatalf("seed controlled_documents: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("seedCreateDocumentSnapshotRows: commit: %v", err)
	}
}

func assertNotNullString(t *testing.T, name string, got sql.NullString) {
	t.Helper()
	if !got.Valid {
		t.Fatalf("%s is NULL", name)
	}
}

func assertNotNullBytes(t *testing.T, name string, got []byte) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s is NULL", name)
	}
}
