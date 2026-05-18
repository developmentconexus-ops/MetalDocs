//go:build integration
// +build integration

package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/repository"
	"metaldocs/tests/integration/testdb"
)

func TestCommitUpload_PersistsRevisionAndFormDataSnapshot(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, fmt.Sprintf(`SET search_path TO %q`, schema)); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	tenantID := testdb.DeterministicID(t, "tenant")
	docID, _ := testdb.InsertDraftDocument(t, db, schema, tenantID)

	var sessionID, userID, baseRevisionID string
	if err := db.QueryRowContext(ctx,
		`SELECT active_session_id::text, created_by::text, current_revision_id::text
		 FROM documents WHERE id=$1::uuid`,
		docID,
	).Scan(&sessionID, &userID, &baseRevisionID); err != nil {
		t.Fatalf("load document session/base: %v", err)
	}

	seedTenantRole(t, db, schema, userID, tenantID, "system_admin")

	pendingID := testdb.DeterministicID(t, "pending")
	contentHash := "hash-commit-upload"
	storageKey := "tenants/test/documents/commit-upload/revisions/hash-commit-upload.docx"
	if _, err := db.ExecContext(ctx,
		`INSERT INTO autosave_pending_uploads
		   (id, session_id, document_id, base_revision_id, content_hash, storage_key, expires_at)
		 VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, $7)`,
		pendingID, sessionID, docID, baseRevisionID, contentHash, storageKey, time.Now().Add(time.Hour),
	); err != nil {
		t.Fatalf("insert pending upload: %v", err)
	}

	repo := repository.New(db)
	formSnapshot := []byte(`{"field":"value"}`)
	res, err := repo.CommitUpload(ctx, tenantID, sessionID, userID, docID, pendingID, contentHash, formSnapshot)
	if err != nil {
		t.Fatalf("CommitUpload: %v", err)
	}
	if res.RevisionNum != 2 {
		t.Fatalf("revision_num = %d, want 2", res.RevisionNum)
	}

	var gotDocumentJSON string
	if err := db.QueryRowContext(ctx,
		`SELECT form_data_json::text FROM documents WHERE id=$1::uuid`,
		docID,
	).Scan(&gotDocumentJSON); err != nil {
		t.Fatalf("query documents.form_data_json: %v", err)
	}
	if gotDocumentJSON != `{"field": "value"}` && gotDocumentJSON != `{"field":"value"}` {
		t.Fatalf("documents.form_data_json = %s", gotDocumentJSON)
	}

	var gotRevisionJSON string
	if err := db.QueryRowContext(ctx,
		`SELECT form_data_snapshot::text FROM document_revisions WHERE id=$1::uuid`,
		res.RevisionID,
	).Scan(&gotRevisionJSON); err != nil {
		t.Fatalf("query document_revisions.form_data_snapshot: %v", err)
	}
	if gotRevisionJSON != `{"field": "value"}` && gotRevisionJSON != `{"field":"value"}` {
		t.Fatalf("document_revisions.form_data_snapshot = %s", gotRevisionJSON)
	}
}

func seedTenantRole(t *testing.T, db *sql.DB, schema, userID, tenantID, roleCode string) {
	t.Helper()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx,
		`INSERT INTO `+testdb.Qualified(schema, "iam_users")+` (user_id, display_name, tenant_id)
		 VALUES ($1, $2, $3::uuid)
		 ON CONFLICT (user_id) DO UPDATE SET display_name = EXCLUDED.display_name, tenant_id = EXCLUDED.tenant_id`,
		userID, userID, tenantID,
	); err != nil {
		t.Fatalf("insert iam_users: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO `+testdb.Qualified(schema, "iam_user_roles")+` (user_id, role_code, tenant_id, assigned_by)
		 VALUES ($1, $2, $3::uuid, $1)
		 ON CONFLICT (tenant_id, user_id) DO UPDATE SET role_code = EXCLUDED.role_code, assigned_by = EXCLUDED.assigned_by`,
		userID, roleCode, tenantID,
	); err != nil {
		t.Fatalf("insert iam_user_roles: %v", err)
	}
}
