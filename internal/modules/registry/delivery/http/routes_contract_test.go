package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	apiv2 "metaldocs/internal/api/v2"
	registryapi "metaldocs/internal/modules/registry/api"
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

type spyRegistryService struct {
	gotCreate application.CreateControlledDocumentCmd
}

func (s *spyRegistryService) Create(ctx context.Context, cmd application.CreateControlledDocumentCmd) (*application.CreateResult, error) {
	s.gotCreate = cmd
	return &application.CreateResult{
		ControlledDocument: &registrydomain.ControlledDocument{ID: "cd-1"},
		DocumentRef:        &registrydomain.DocumentRef{ID: "doc-1", ContentHash: "hash-1"},
	}, nil
}

func (s *spyRegistryService) List(ctx context.Context, tenantID string, filter application.CDFilter) ([]registrydomain.ControlledDocument, error) {
	return nil, nil
}

func (s *spyRegistryService) CreateRevision(ctx context.Context, cmd application.CreateRevisionCmd) (*registrydomain.DocumentRef, error) {
	return nil, nil
}

func (s *spyRegistryService) Get(ctx context.Context, tenantID, id string) (*registrydomain.ControlledDocument, error) {
	return nil, registrydomain.ErrCDNotFound
}

func (s *spyRegistryService) Obsolete(ctx context.Context, tenantID, id string) error {
	return nil
}

func (s *spyRegistryService) Supersede(ctx context.Context, tenantID, id string) error {
	return nil
}

func (s *spyRegistryService) PeekSeq(ctx context.Context, tenantID, profileCode, areaCode string) (int, error) {
	return 1, nil
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

func TestAtomicCreate_MissingDocumentName_Returns400(t *testing.T) {
	handler := &Handler{svc: &spyRegistryService{}}
	req := httptest.NewRequest(http.MethodPost, "/api/v2/controlled-documents", strings.NewReader(`{
		"profileCode":"DC",
		"processAreaCode":"RH",
		"title":"Policy",
		"ownerUserId":"user-1"
	}`))
	rec := httptest.NewRecorder()

	handler.AtomicCreateControlledDocument(rec, req, registryapi.AtomicCreateControlledDocumentParams{})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "documentName") {
		t.Fatalf("body %q does not mention documentName", rec.Body.String())
	}
}

func TestAtomicCreate_UnknownField_Returns400(t *testing.T) {
	handler := &Handler{svc: &spyRegistryService{}}
	req := httptest.NewRequest(http.MethodPost, "/api/v2/controlled-documents", strings.NewReader(`{
		"documentName":"Policy v1",
		"profileCode":"DC",
		"processAreaCode":"RH",
		"title":"Policy",
		"ownerUserId":"user-1",
		"evilField":true
	}`))
	rec := httptest.NewRecorder()

	handler.AtomicCreateControlledDocument(rec, req, registryapi.AtomicCreateControlledDocumentParams{})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "evilField") {
		t.Fatalf("body %q does not mention evilField", rec.Body.String())
	}
}

func TestAtomicCreate_ForwardsGeneratedOnlyFields(t *testing.T) {
	spy := &spyRegistryService{}
	handler := &Handler{svc: spy}
	req := httptest.NewRequest(http.MethodPost, "/api/v2/controlled-documents", strings.NewReader(`{
		"documentName":"Policy v1",
		"profileCode":"DC",
		"processAreaCode":"RH",
		"title":"Policy",
		"ownerUserId":"user-1",
		"templateVersionId":"11111111-1111-1111-1111-111111111111",
		"formData":{"summary":"hello","count":2}
	}`))
	rec := httptest.NewRecorder()

	handler.AtomicCreateControlledDocument(rec, req, registryapi.AtomicCreateControlledDocumentParams{})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}
	if spy.gotCreate.DocumentName != "Policy v1" {
		t.Fatalf("DocumentName = %q, want Policy v1", spy.gotCreate.DocumentName)
	}
	if spy.gotCreate.TemplateVersionID == nil || *spy.gotCreate.TemplateVersionID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("TemplateVersionID = %v, want 11111111-1111-1111-1111-111111111111", spy.gotCreate.TemplateVersionID)
	}
	if spy.gotCreate.FormData["summary"] != "hello" {
		t.Fatalf("FormData[summary] = %v, want hello", spy.gotCreate.FormData["summary"])
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

// TestPostControlledDocuments_MissingIdempotencyKey_400: POST to atomic-create
// endpoint without Idempotency-Key header must return 400 with code IDEMPOTENCY_KEY_REQUIRED.
func TestPostControlledDocuments_MissingIdempotencyKey_400(t *testing.T) {
	handler := newTestHandler(nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v2/controlled-documents", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	if body["code"] != "IDEMPOTENCY_KEY_REQUIRED" {
		t.Fatalf("code = %q, want IDEMPOTENCY_KEY_REQUIRED", body["code"])
	}
}

// TestGetPreviewCode_200: GET /api/v2/controlled-documents/preview-code with valid
// query params returns 200 with profileCode, areaCode, nextSeq, and code fields.
func TestGetPreviewCode_200(t *testing.T) {
	handler := newTestHandler(nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/controlled-documents/preview-code?profileCode=DC&areaCode=RH", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	for _, field := range []string{"profileCode", "areaCode", "nextSeq", "code"} {
		if _, ok := body[field]; !ok {
			t.Errorf("missing field %q in response: %s", field, rec.Body.String())
		}
	}
}

// TestGetPreviewCode_MissingParams_400: GET /api/v2/controlled-documents/preview-code
// without required query params returns 400.
func TestGetPreviewCode_MissingParams_400(t *testing.T) {
	handler := newTestHandler(nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/controlled-documents/preview-code?profileCode=DC", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
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
