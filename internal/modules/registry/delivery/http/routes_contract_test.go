package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	apiv2 "metaldocs/internal/api/v2"
	"metaldocs/internal/modules/registry/application"
	registrydomain "metaldocs/internal/modules/registry/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

type fakeRegistryDocs struct{}

func (f fakeRegistryDocs) GetByID(ctx context.Context, tenantID, id string) (*registrydomain.ControlledDocument, error) {
	return nil, registrydomain.ErrCDNotFound
}

func (f fakeRegistryDocs) GetByCode(ctx context.Context, tenantID, profileCode, code string) (*registrydomain.ControlledDocument, error) {
	return nil, nil
}

func (f fakeRegistryDocs) CodeExists(ctx context.Context, tenantID, profileCode, code string) (bool, error) {
	return false, nil
}

func (f fakeRegistryDocs) List(ctx context.Context, tenantID string, filter registrydomain.CDFilter) ([]registrydomain.ControlledDocument, error) {
	return nil, nil
}

func (f fakeRegistryDocs) Create(ctx context.Context, doc *registrydomain.ControlledDocument) error {
	return nil
}

func (f fakeRegistryDocs) CreateTx(ctx context.Context, tx *sql.Tx, doc *registrydomain.ControlledDocument) error {
	return nil
}

func (f fakeRegistryDocs) UpdateStatus(ctx context.Context, tenantID, id string, status registrydomain.CDStatus, updatedAt time.Time) error {
	return nil
}

type fakeSequenceAllocator struct{}

func (f fakeSequenceAllocator) NextAndIncrement(ctx context.Context, tx registrydomain.DBExecutor, tenantID, profileCode, areaCode string) (int, error) {
	return 1, nil
}

func (f fakeSequenceAllocator) Peek(ctx context.Context, tenantID, profileCode, areaCode string) (int, error) {
	return 1, nil
}

func (f fakeSequenceAllocator) EnsureCounter(ctx context.Context, tenantID, profileCode, areaCode string) error {
	return nil
}

type fakeTemplateChecker struct{}

func (f fakeTemplateChecker) GetTemplateVersionState(ctx context.Context, templateVersionID string) (*string, string, error) {
	return nil, "", nil
}

type fakeProfileReader struct{}

func (f fakeProfileReader) GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.DocumentProfile, error) {
	return nil, nil
}

type fakeAreaReader struct{}

func (f fakeAreaReader) GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.ProcessArea, error) {
	return nil, nil
}

type fakeGovernanceLogger struct{}

func (f fakeGovernanceLogger) Log(ctx context.Context, e taxonomydomain.GovernanceEvent) error {
	return nil
}

// helpers

func newSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func newTestHandler(db *sql.DB) *Handler {
	svc := application.NewRegistryService(
		nil,
		fakeRegistryDocs{},
		fakeSequenceAllocator{},
		fakeTemplateChecker{},
		fakeProfileReader{},
		fakeAreaReader{},
		fakeGovernanceLogger{},
		nil,
	)
	return NewHandler(svc, db)
}

func newAuthedRequest(t *testing.T, method, url, tenantID string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, url, nil)
	req.Header.Set("X-Tenant-ID", tenantID)
	// extract {id} path value from URL pattern /api/v2/controlled-documents/{id}/active-document
	// httptest doesn't set path values automatically; set manually
	// URL format: /api/v2/controlled-documents/<id>/active-document
	// Parse id from URL
	const prefix = "/api/v2/controlled-documents/"
	const suffix = "/active-document"
	if len(url) > len(prefix)+len(suffix) {
		id := url[len(prefix) : len(url)-len(suffix)]
		req.SetPathValue("id", id)
	}
	return req
}

// existing tests

func TestRegistryHandler_ErrorEnvelopeContract(t *testing.T) {
	svc := application.NewRegistryService(
		nil,
		fakeRegistryDocs{},
		fakeSequenceAllocator{},
		fakeTemplateChecker{},
		fakeProfileReader{},
		fakeAreaReader{},
		fakeGovernanceLogger{},
		nil,
	)
	handler := NewHandler(svc, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/controlled-documents/missing", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var apiErr apiv2.APIError
	if err := json.Unmarshal(rec.Body.Bytes(), &apiErr); err != nil {
		t.Fatalf("unmarshal api error: %v body=%s", err, rec.Body.String())
	}
	if apiErr.Code == "" {
		t.Fatalf("expected non-empty code in API error: %s", rec.Body.String())
	}
}

func TestActiveDocumentResponse_IncludesApprovalInstanceID(t *testing.T) {
	approvalInstanceID := "approval-instance-1"
	docID := "document-1"
	approvalState := "under_review"
	contentHash := "hash-1"
	revVersion := 2
	resp := activeDocumentResponse{
		DocumentID:         &docID,
		ApprovalState:      &approvalState,
		ContentHash:        &contentHash,
		RevisionVersion:    &revVersion,
		ApprovalInstanceID: &approvalInstanceID,
	}

	body, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got["approvalInstanceId"] != approvalInstanceID {
		t.Fatalf("approvalInstanceId = %v, want %s", got["approvalInstanceId"], approvalInstanceID)
	}
}

// E10 contract tests

// TestActiveDocument_OnlyPublished_Returns200_WithPublishedID: controlled document has only a
// published revision (no draft/under_review/etc.) — getActiveDocument must return 200 with
// publishedDocumentId set and documentId absent.
func TestActiveDocument_OnlyPublished_Returns200_WithPublishedID(t *testing.T) {
	db, mock := newSQLMock(t)
	handler := newTestHandler(db)

	tenantID := "tenant-1"
	cdID := "cd-1"
	publishedDocID := "pub-doc-1"

	// Main FULL OUTER JOIN query: no active doc row — only published side
	mainRows := sqlmock.NewRows([]string{
		"doc_id", "content_hash", "revision_version", "approval_state", "published_doc_id",
	}).AddRow(nil, nil, nil, nil, publishedDocID)
	mock.ExpectQuery(`SELECT`).
		WithArgs(tenantID, cdID).
		WillReturnRows(mainRows)

	// approval_instances query — no active instance
	aiRows := sqlmock.NewRows([]string{"id"})
	mock.ExpectQuery(`SELECT`).
		WithArgs(sqlmock.AnyArg(), tenantID).
		WillReturnRows(aiRows)

	req := newAuthedRequest(t, http.MethodGet,
		"/api/v2/controlled-documents/"+cdID+"/active-document", tenantID)
	rec := httptest.NewRecorder()
	handler.getActiveDocument(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["publishedDocumentId"] != publishedDocID {
		t.Errorf("publishedDocumentId = %v, want %s", body["publishedDocumentId"], publishedDocID)
	}
	if _, ok := body["documentId"]; ok {
		t.Errorf("documentId should be absent (omitempty), got %v", body["documentId"])
	}
}

// TestActiveDocument_BothActiveAndPublished_Returns200_WithBoth: active draft + published revision
// both exist — both documentId and publishedDocumentId must be present.
func TestActiveDocument_BothActiveAndPublished_Returns200_WithBoth(t *testing.T) {
	db, mock := newSQLMock(t)
	handler := newTestHandler(db)

	tenantID := "tenant-2"
	cdID := "cd-2"
	activeDocID := "active-doc-2"
	publishedDocID := "pub-doc-2"
	contentHash := "abc123"
	revVersion := 3
	approvalState := "draft"

	mainRows := sqlmock.NewRows([]string{
		"doc_id", "content_hash", "revision_version", "approval_state", "published_doc_id",
	}).AddRow(activeDocID, contentHash, revVersion, approvalState, publishedDocID)
	mock.ExpectQuery(`SELECT`).
		WithArgs(tenantID, cdID).
		WillReturnRows(mainRows)

	// no in-progress approval instance
	aiRows := sqlmock.NewRows([]string{"id"})
	mock.ExpectQuery(`SELECT`).
		WithArgs(activeDocID, tenantID).
		WillReturnRows(aiRows)

	req := newAuthedRequest(t, http.MethodGet,
		"/api/v2/controlled-documents/"+cdID+"/active-document", tenantID)
	rec := httptest.NewRecorder()
	handler.getActiveDocument(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["documentId"] != activeDocID {
		t.Errorf("documentId = %v, want %s", body["documentId"], activeDocID)
	}
	if body["publishedDocumentId"] != publishedDocID {
		t.Errorf("publishedDocumentId = %v, want %s", body["publishedDocumentId"], publishedDocID)
	}
}

// TestActiveDocument_NoneExist_Returns404: no active doc and no published revision —
// must return 404.
func TestActiveDocument_NoneExist_Returns404(t *testing.T) {
	db, mock := newSQLMock(t)
	handler := newTestHandler(db)

	tenantID := "tenant-3"
	cdID := "cd-3"

	mainRows := sqlmock.NewRows([]string{
		"doc_id", "content_hash", "revision_version", "approval_state", "published_doc_id",
	})
	mock.ExpectQuery(`SELECT`).
		WithArgs(tenantID, cdID).
		WillReturnRows(mainRows)

	req := newAuthedRequest(t, http.MethodGet,
		"/api/v2/controlled-documents/"+cdID+"/active-document", tenantID)
	rec := httptest.NewRecorder()
	handler.getActiveDocument(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d, want 404; body=%s", rec.Code, rec.Body.String())
	}
}
