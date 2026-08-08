//go:build integration
// +build integration

package application_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/application"
	docrepo "metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/docgenv2"
	"metaldocs/tests/integration/testdb"
)

func TestCreateDocumentTx_PopulatesAllSnapshotColumns(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

	// More than one connection is required here: the off-tx UserDisplayNameReader
	// port read and the templates GetPublishedVersion read both run on the pool
	// while the create tx holds a connection (H-PRE-1) — a single-connection pool
	// deadlocks.
	db.SetMaxIdleConns(0)
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(4)

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID

	actorID := testdb.DeterministicID(t, "actor")
	templateID := testdb.DeterministicID(t, "template")
	templateVersionID := testdb.DeterministicID(t, "template-version")

	const wantDisplayName = "Snapshot Author"
	// Seed the actor as a tenant-scoped system_admin (satisfies the create path's
	// app-level authz.Require via the ADR 0022 tier-2 bypass) with display_name set,
	// so the REAL UserDisplayNameReader port resolves a value to snapshot (F4.1 value
	// proof). iam_users is not tripwire-governed; the role write is user.manage tx-local.
	testdb.SeedSystemAdmin(t, db, tenantID, actorID, wantDisplayName)

	// SUT ctx must carry the seeded tenant + actor: CreateDocumentTx's tier-2
	// authz.Require(document.create) reads identity off the tx-local GUCs, and
	// downstream off-tx reads (UserDisplayNameReader, templates readers) seed
	// their own GUCs from this ctx via TxRunner.
	ctx = testdb.AuthzCtx(tenantID, actorID)

	// Seed taxonomy parents required by the controlled_documents FK chain.
	testdb.SeedGovernedTaxonomy(t, db, tenantID, "po", "quality")

	// Seed controlled_documents via factory (controlled_documents.create tripwire).
	// Supply the pre-seeded taxonomy so the factory reuses it instead of minting new codes.
	cdFixture := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithTaxonomy(testdb.Taxonomy{TenantID: tenantID, ProfileCode: "po", ProcessAreaCode: "quality"}),
		testdb.WithCode("PO-TEST-001"),
		testdb.WithOwner(actorID),
	)

	// Seed templates via a tenant+caps-seeded tx (no factory builder exists for
	// templates_template). templates_template carries template.create;
	// templates_template_version shares the same tripwire AND
	// trg_template_version_tenant_consistent, which additionally requires
	// metaldocs.tenant_id to be set tx-locally — so this seeds tenant identity
	// (authz.SeedTxTenant, the production chokepoint) alongside caps
	// (testdb.SetCapsOnTx), matching the sanctioned direct-tx pattern used
	// throughout tests/integration for tenant-consistency-guarded writes.
	func() {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin template seed tx: %v", err)
		}
		defer tx.Rollback()
		if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
			t.Fatalf("seed template tx tenant: %v", err)
		}
		testdb.SetCapsOnTx(t, tx, `[{"cap":"template.create"}]`)
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
				id, tenant_id, template_id, version_number, status, docx_storage_key, content_hash,
				metadata_schema, placeholder_schema, author_id, published_at
			) VALUES (
				$1::uuid, $4::uuid, $2::uuid, 1, 'published', 'templates/snapshot/body.docx', $5,
				'{}'::jsonb, '{"placeholders":[]}'::jsonb, $3, now()
			)`,
			templateVersionID, templateID, actorID, tenantID, strings.Repeat("a", 64),
		); err != nil {
			t.Fatalf("seed templates_template_version: %v", err)
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE templates_template
			   SET published_version_id = $1::uuid
			 WHERE id = $2::uuid`,
			templateVersionID, templateID,
		); err != nil {
			t.Fatalf("update templates_template published_version_id: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit template seed tx: %v", err)
		}
	}()

	cd := &controlleddocumentsdomain.ControlledDocument{
		ID:              cdFixture.ID,
		TenantID:        cdFixture.TenantID,
		ProfileCode:     cdFixture.ProfileCode,
		ProcessAreaCode: cdFixture.ProcessAreaCode,
		Code:            cdFixture.Code,
		Title:           "Snapshot Test Controlled Document",
		OwnerUserID:     actorID,
		Status:          controlleddocumentsdomain.CDStatusActive,
	}

	snapshotSvc := application.NewSnapshotService(
		docgenv2.NewTemplatesSnapshotReader(db),
	)
	svc := application.NewServiceWithSnapshot(
		docrepo.New(db, iampg.NewUserDisplayNameRepository(db), controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{}),
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
		  FROM public.documents
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
