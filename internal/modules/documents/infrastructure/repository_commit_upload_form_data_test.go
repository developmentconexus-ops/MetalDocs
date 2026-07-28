package infrastructure

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

const (
	preserveFormDataQuery = "SELECT form_data_json FROM documents WHERE id=$1 AND tenant_id=$2::uuid"
	updateDocumentPointer = "UPDATE documents SET current_revision_id=$1, form_data_json=$2, updated_at=now() WHERE id=$3 AND tenant_id=$4::uuid"
)

// expectCommitUploadPreamble mocks everything CommitUpload runs before the
// revision INSERT: the authz preamble, the pending-upload lock (which also
// locks the documents row via the join) and the session re-verification.
func expectCommitUploadPreamble(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	mock.ExpectBegin()
	expectDocumentEditAuthz(t, mock)
	mock.ExpectQuery(`FROM autosave_pending_uploads p`).
		WithArgs("pending-1", "tenant-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "session_id", "document_id", "base_revision_id", "content_hash",
			"storage_key", "expires_at", "consumed_at",
		}).AddRow("pending-1", "sess-1", "doc-1", "rev-0", "hash-1",
			"tenant-1/doc-1/hash-1", time.Now().Add(time.Hour), nil))
	mock.ExpectQuery(`FROM editor_sessions WHERE id=\$1`).
		WithArgs("sess-1", "tenant-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "last_acknowledged_revision_id", "status"}).
			AddRow("user-1", "rev-0", "active"))
}

// expectCommitUploadTail mocks the revision INSERT + the three writes that
// follow it, asserting the exact form-data bytes reaching document_revisions
// and documents.
func expectCommitUploadTail(t *testing.T, mock sqlmock.Sqlmock, wantFormData []byte) {
	t.Helper()
	mock.ExpectQuery(`INSERT INTO document_revisions`).
		WithArgs("doc-1", "rev-0", "sess-1", "tenant-1/doc-1/hash-1", "hash-1",
			wantFormData, int64(2048), nil, nil).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "revision_num", "file_size_bytes", "page_count", "page_count_source",
		}).AddRow("rev-1", int64(1), int64(2048), nil, nil))
	mock.ExpectExec(regexp.QuoteMeta(updateDocumentPointer)).
		WithArgs("rev-1", wantFormData, "doc-1", "tenant-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE editor_sessions SET last_acknowledged_revision_id=\$1`).
		WithArgs("rev-1", "sess-1", "tenant-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE autosave_pending_uploads SET consumed_at=now\(\)`).
		WithArgs("pending-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}

// TestCommitUpload_AbsentFormDataSnapshotPreservesStoredFormData locks F-QA4-9:
// form_data_snapshot is OPTIONAL in the contract, so an autosave commit that
// omits it must NOT write NULL into documents.form_data_json (NOT NULL ->
// constraint violation -> 500) and must NOT substitute a fabricated `{}`
// (silent data loss). The stored form data is read under the lock already held
// on the documents row and re-stamped on both the new revision and the
// document, so the commit is a pure artifact update.
func TestCommitUpload_AbsentFormDataSnapshotPreservesStoredFormData(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	r := New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	stored := []byte(`{"kept":"yes"}`)
	expectCommitUploadPreamble(t, mock)
	mock.ExpectQuery(regexp.QuoteMeta(preserveFormDataQuery)).
		WithArgs("doc-1", "tenant-1").
		WillReturnRows(sqlmock.NewRows([]string{"form_data_json"}).AddRow(stored))
	expectCommitUploadTail(t, mock, stored)

	res, err := r.CommitUpload(context.Background(), "tenant-1", "sess-1", "user-1", "doc-1", "pending-1", "hash-1", nil, 2048, nil, nil)
	if err != nil {
		t.Fatalf("CommitUpload with absent form_data_snapshot: %v", err)
	}
	if res.RevisionID != "rev-1" || res.RevisionNum != 1 {
		t.Fatalf("unexpected commit result: %#v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// TestCommitUpload_PresentFormDataSnapshotIsWritten proves the preserve branch
// is scoped to the absent case: a supplied snapshot is written verbatim and no
// preserving read is issued (ordered expectations would reject an extra query).
func TestCommitUpload_PresentFormDataSnapshotIsWritten(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	r := New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	snapshot := []byte(`{"field":"new"}`)
	expectCommitUploadPreamble(t, mock)
	expectCommitUploadTail(t, mock, snapshot)

	if _, err := r.CommitUpload(context.Background(), "tenant-1", "sess-1", "user-1", "doc-1", "pending-1", "hash-1", snapshot, 2048, nil, nil); err != nil {
		t.Fatalf("CommitUpload with present form_data_snapshot: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
