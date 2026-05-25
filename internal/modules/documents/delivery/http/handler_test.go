package http_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/documents/application"
	approvalapp "metaldocs/internal/modules/documents/approval/application"
	httphandler "metaldocs/internal/modules/documents/delivery/http"
	"metaldocs/internal/modules/documents/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/tenant"
)

type fakeSvc struct {
	listDocs       []domain.Document
	listForUser    []domain.Document
	listErr        error
	listForUserErr error

	acquireSession *domain.Session
	acquireRO      bool
	acquireErr     error

	commitResult *application.CommitResult
	commitErr    error
	commitCmd    application.CommitAutosaveCmd
	syncResult   *application.CommitResult
	syncErr      error
	syncCmd      application.SyncArtifactMetadataCmd

	renameErr       error
	renameName      string
	duplicateErr    error
	checkpointLabel string
	checkpointCalls int

	listPaginatedItems []*domain.Document
	listPaginatedTotal int64
	listPaginatedErr   error
	listPaginatedOpts  application.ListOptions
	listPaginatedUser  string

	revisionHistory []domain.RevisionHistoryItem
	revisionHistErr error

	statsResult *application.DocumentStats
	statsErr    error

	getDocumentResult *domain.Document
}

var _ httphandler.Service = (*fakeSvc)(nil)

func (f *fakeSvc) CreateDocument(_ context.Context, _ application.CreateDocumentInput) (*application.CreateDocumentResult, error) {
	return &application.CreateDocumentResult{DocumentID: "doc_1", InitialRevisionID: "rev_1", SessionID: "sess_1"}, nil
}

func (f *fakeSvc) GetDocument(_ context.Context, _, _ string) (*domain.Document, error) {
	if f.getDocumentResult != nil {
		return f.getDocumentResult, nil
	}
	return &domain.Document{
		ID:                "doc_1",
		Name:              "Doc",
		Status:            domain.DocStatusDraft,
		FormDataJSON:      []byte(`{"foo":"bar"}`),
		CurrentRevisionID: "rev_1",
		RevisionVersion:   1,
		RevisionNumber:    0,
		CreatedBy:         "user_1",
		Code:              "DOC-1",
	}, nil
}

func (f *fakeSvc) RenameDocument(_ context.Context, _, _, _, newName string) error {
	f.renameName = newName
	return f.renameErr
}

func (f *fakeSvc) ListDocuments(_ context.Context, _ string) ([]domain.Document, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.listDocs == nil {
		return []domain.Document{{ID: "doc_1"}}, nil
	}
	return f.listDocs, nil
}

func (f *fakeSvc) ListDocumentsForUser(_ context.Context, _, _ string) ([]domain.Document, error) {
	if f.listForUserErr != nil {
		return nil, f.listForUserErr
	}
	if f.listForUser == nil {
		return []domain.Document{{ID: "doc_1"}}, nil
	}
	return f.listForUser, nil
}

func (f *fakeSvc) ListDocumentsPaginated(_ context.Context, _, userID string, opts application.ListOptions) ([]*domain.Document, int64, error) {
	f.listPaginatedOpts = opts
	f.listPaginatedUser = userID
	if f.listPaginatedErr != nil {
		return nil, 0, f.listPaginatedErr
	}
	if f.listPaginatedItems == nil {
		return []*domain.Document{{ID: "doc_1"}}, 1, nil
	}
	return f.listPaginatedItems, f.listPaginatedTotal, nil
}

func (f *fakeSvc) DocumentStats(_ context.Context, _, _ string, _ application.ListOptions) (*application.DocumentStats, error) {
	if f.statsErr != nil {
		return nil, f.statsErr
	}
	if f.statsResult == nil {
		return &application.DocumentStats{
			ByStatus: map[string]int64{"draft": 1},
			ByArea:   map[string]int64{"RH": 1},
		}, nil
	}
	return f.statsResult, nil
}

func (f *fakeSvc) IsDocumentOwner(_ context.Context, _, _, _ string) (bool, error) {
	return true, nil
}

func (f *fakeSvc) AcquireSession(_ context.Context, _, _, _ string) (*domain.Session, bool, error) {
	if f.acquireErr != nil {
		return nil, false, f.acquireErr
	}
	if f.acquireSession == nil {
		return &domain.Session{ID: "sess_1", DocumentID: "doc_1", UserID: "user_1", Status: domain.SessionActive}, f.acquireRO, nil
	}
	return f.acquireSession, f.acquireRO, nil
}

func (f *fakeSvc) HeartbeatSession(_ context.Context, _, _ string) error { return nil }

func (f *fakeSvc) ReleaseSession(_ context.Context, _, _, _, _ string) error { return nil }

func (f *fakeSvc) ForceReleaseSession(_ context.Context, _, _, _, _ string) error { return nil }

func (f *fakeSvc) PresignAutosave(_ context.Context, _ application.PresignAutosaveCmd) (*application.PresignAutosaveResult, error) {
	return &application.PresignAutosaveResult{UploadURL: "https://example/upload", PendingUploadID: "pending_1", ExpiresAt: time.Now().Add(time.Minute)}, nil
}

func (f *fakeSvc) CommitAutosave(_ context.Context, cmd application.CommitAutosaveCmd) (*application.CommitResult, error) {
	f.commitCmd = cmd
	if f.commitErr != nil {
		return nil, f.commitErr
	}
	if f.commitResult == nil {
		return &application.CommitResult{RevisionID: "rev_2", RevisionNum: 2}, nil
	}
	return f.commitResult, nil
}

func (f *fakeSvc) SyncArtifactMetadata(_ context.Context, cmd application.SyncArtifactMetadataCmd) (*application.CommitResult, error) {
	f.syncCmd = cmd
	if f.syncErr != nil {
		return nil, f.syncErr
	}
	if f.syncResult == nil {
		fileSizeBytes := int64(1304)
		pageCount := 3
		pageCountSource := "eigenpal_client"
		f.syncResult = &application.CommitResult{
			RevisionID:      "rev_1",
			FileSizeBytes:   &fileSizeBytes,
			PageCount:       &pageCount,
			PageCountSource: &pageCountSource,
		}
	}
	return f.syncResult, nil
}

func (f *fakeSvc) CreateCheckpoint(_ context.Context, _, _, _, label string) (*domain.Checkpoint, error) {
	f.checkpointCalls++
	f.checkpointLabel = label
	return &domain.Checkpoint{ID: "cp_1", VersionNum: 1}, nil
}

func (f *fakeSvc) ListCheckpoints(_ context.Context, _, _ string) ([]domain.Checkpoint, error) {
	return []domain.Checkpoint{{ID: "cp_1", VersionNum: 1}}, nil
}

func (f *fakeSvc) ListRevisionHistory(_ context.Context, _, _ string) ([]domain.RevisionHistoryItem, error) {
	if f.revisionHistErr != nil {
		return nil, f.revisionHistErr
	}
	if f.revisionHistory == nil {
		return []domain.RevisionHistoryItem{{
			DocumentID:     "doc_1",
			RevisionNumber: 2,
			RevisionTitle:  "Ajuste operacional",
			Status:         domain.DocStatusDraft,
			CreatedAt:      time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
			IsCurrent:      true,
		}}, nil
	}
	return f.revisionHistory, nil
}

func (f *fakeSvc) RestoreCheckpoint(_ context.Context, _, _, _ string, _ int) (*application.RestoreResult, error) {
	return &application.RestoreResult{NewRevisionID: "rev_3", NewRevisionNum: 3}, nil
}

func (f *fakeSvc) DuplicateDocument(_ context.Context, _, _, _ string) (*application.CreateDocumentResult, error) {
	if f.duplicateErr != nil {
		return nil, f.duplicateErr
	}
	return &application.CreateDocumentResult{DocumentID: "doc_dup", InitialRevisionID: "rev_dup", SessionID: "sess_dup"}, nil
}

func (f *fakeSvc) Finalize(_ context.Context, _, _, _ string) error { return nil }

func (f *fakeSvc) Archive(_ context.Context, _, _, _ string) error { return nil }

func (f *fakeSvc) SignedRevisionURL(_ context.Context, _, _, _ string) (string, error) {
	return "https://example/rev", nil
}

func (f *fakeSvc) ListDocumentComments(_ context.Context, _, _, _ string) ([]domain.Comment, error) {
	return nil, nil
}

func (f *fakeSvc) AddDocumentComment(_ context.Context, _, _, _, _ string, _ domain.CommentCreateInput) (*domain.Comment, error) {
	return &domain.Comment{}, nil
}

func (f *fakeSvc) UpdateDocumentComment(_ context.Context, _, _, _ string, _ int, _ domain.CommentUpdateInput) (*domain.Comment, error) {
	return &domain.Comment{}, nil
}

func (f *fakeSvc) DeleteDocumentComment(_ context.Context, _, _, _ string, _ int) error {
	return nil
}

func newMux(t *testing.T, svc *fakeSvc) *http.ServeMux {
	t.Helper()
	h := httphandler.NewHandler(svc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

type fakeApprovalSubmitter struct {
	called bool
}

func (f *fakeApprovalSubmitter) SubmitRevisionForReview(_ context.Context, _ *sql.DB, _ approvalapp.SubmitRequest) (approvalapp.SubmitResult, error) {
	f.called = true
	return approvalapp.SubmitResult{InstanceID: "inst_1"}, nil
}

type fakeFinalizeIdempotencyStore struct {
	replay *idempotency.Replay
}

func (f *fakeFinalizeIdempotencyStore) BeginReplay(_ context.Context, _, _, _, _ string) (*idempotency.ReplayHandle, *idempotency.Replay, error) {
	if f.replay != nil {
		return nil, f.replay, nil
	}
	return &idempotency.ReplayHandle{}, nil, nil
}

func (f *fakeFinalizeIdempotencyStore) CompleteReplay(_ *idempotency.ReplayHandle, _ int, _ []byte) error {
	return nil
}

func (f *fakeFinalizeIdempotencyStore) FailReplay(_ *idempotency.ReplayHandle, _ error) error {
	return nil
}

func withAuthHeaders(req *http.Request, roles string) {
	req.Header.Set("content-type", "application/json")
	*req = *req.WithContext(tenant.WithTenantID(req.Context(), "tenant_1"))
	ctxRoles := make([]iamdomain.Role, 0)
	for _, part := range strings.Split(roles, ",") {
		role := strings.TrimSpace(part)
		if role != "" {
			ctxRoles = append(ctxRoles, iamdomain.Role(role))
		}
	}
	*req = *req.WithContext(iamdomain.WithAuthContext(req.Context(), "user_1", ctxRoles))
}

func TestListDocuments_Happy(t *testing.T) {
	mux := newMux(t, &fakeSvc{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetDocument_EmbedsFormDataJSON(t *testing.T) {
	svc := &fakeSvc{}
	mux := newMux(t, svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/doc_1", nil)
	req.SetPathValue("id", "doc_1")
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	formData, ok := body["FormDataJSON"].(map[string]any)
	if !ok {
		t.Fatalf("expected FormDataJSON object, got %#v", body["FormDataJSON"])
	}
	if got := formData["foo"]; got != "bar" {
		t.Fatalf("expected FormDataJSON.foo=bar, got %#v", got)
	}
}

func TestGetDocument_ReturnsCurrentRevisionArtifactMetadata(t *testing.T) {
	fileSizeBytes := int64(1304)
	pageCount := 3
	pageCountSource := "eigenpal_client"
	svc := &fakeSvc{getDocumentResult: &domain.Document{
		ID:                             "doc_1",
		Name:                           "Doc",
		Status:                         domain.DocStatusDraft,
		FormDataJSON:                   []byte(`{"foo":"bar"}`),
		CurrentRevisionID:              "rev_1",
		RevisionVersion:                1,
		CreatedBy:                      "user_1",
		Code:                           "DOC-1",
		CurrentRevisionFileSizeBytes:   &fileSizeBytes,
		CurrentRevisionPageCount:       &pageCount,
		CurrentRevisionPageCountSource: &pageCountSource,
	}}
	mux := newMux(t, svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/doc_1", nil)
	req.SetPathValue("id", "doc_1")
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got := body["currentRevisionFileSizeBytes"]; got != float64(1304) {
		t.Fatalf("currentRevisionFileSizeBytes = %v", got)
	}
	if got := body["currentRevisionPageCount"]; got != float64(3) {
		t.Fatalf("currentRevisionPageCount = %v", got)
	}
	if got := body["currentRevisionPageCountSource"]; got != "eigenpal_client" {
		t.Fatalf("currentRevisionPageCountSource = %v", got)
	}
}

func TestListDocuments_Forbidden(t *testing.T) {
	mux := newMux(t, &fakeSvc{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	withAuthHeaders(req, "template_author")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("want application/problem+json, got %s", ct)
	}
}

func TestAcquireSession_Happy(t *testing.T) {
	mux := newMux(t, &fakeSvc{acquireSession: &domain.Session{ID: "sess_1"}})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/session/acquire", bytes.NewReader([]byte(`{}`)))
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
}

func TestAcquireSession_Forbidden(t *testing.T) {
	mux := newMux(t, &fakeSvc{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/session/acquire", bytes.NewReader([]byte(`{}`)))
	withAuthHeaders(req, "template_author")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestCommitAutosave_IdempotentReplay_Returns200(t *testing.T) {
	mux := newMux(t, &fakeSvc{commitResult: &application.CommitResult{RevisionID: "rev_2", RevisionNum: 2, AlreadyConsumed: true}})

	body := []byte(`{"session_id":"sess_1","pending_upload_id":"pending_1","form_data_snapshot":{"a":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/autosave/commit", bytes.NewReader(body))
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if ok, _ := out["idempotent_replay"].(bool); !ok {
		t.Fatalf("expected idempotent_replay=true, got %v", out["idempotent_replay"])
	}
}

func TestCommitAutosave_AcceptsPageCountAndReturnsArtifactMetadata(t *testing.T) {
	fileSizeBytes := int64(1304)
	pageCount := 3
	pageCountSource := "eigenpal_client"
	svc := &fakeSvc{commitResult: &application.CommitResult{
		RevisionID:      "rev_2",
		RevisionNum:     2,
		FileSizeBytes:   &fileSizeBytes,
		PageCount:       &pageCount,
		PageCountSource: &pageCountSource,
	}}
	mux := newMux(t, svc)

	body := []byte(`{"session_id":"sess_1","pending_upload_id":"pending_1","form_data_snapshot":{"a":1},"page_count":3}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/autosave/commit", bytes.NewReader(body))
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if svc.commitCmd.PageCount == nil || *svc.commitCmd.PageCount != 3 {
		t.Fatalf("page_count not passed to service: %#v", svc.commitCmd.PageCount)
	}

	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got := out["file_size_bytes"]; got != float64(1304) {
		t.Fatalf("file_size_bytes = %v", got)
	}
	if got := out["page_count"]; got != float64(3) {
		t.Fatalf("page_count = %v", got)
	}
	if got := out["page_count_source"]; got != "eigenpal_client" {
		t.Fatalf("page_count_source = %v", got)
	}
}

func TestCommitAutosave_InvalidPageCountUsesProblemEnvelope(t *testing.T) {
	mux := newMux(t, &fakeSvc{commitErr: domain.ErrInvalidPageCount})

	body := []byte(`{"session_id":"sess_1","pending_upload_id":"pending_1","page_count":0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/autosave/commit", bytes.NewReader(body))
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("want application/problem+json, got %s", ct)
	}
	if !strings.Contains(rr.Body.String(), "invalid_page_count") {
		t.Fatalf("expected invalid_page_count problem, got %s", rr.Body.String())
	}
}

func TestForceReleaseSession_RequiresAdmin(t *testing.T) {
	mux := newMux(t, &fakeSvc{})

	body := []byte(`{"session_id":"sess_1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/session/force-release", bytes.NewReader(body))
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestRenameDocument_Happy(t *testing.T) {
	svc := &fakeSvc{}
	mux := newMux(t, svc)

	body := []byte(`{"name":"Updated Name"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/documents/doc_1", bytes.NewReader(body))
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if svc.renameName != "Updated Name" {
		t.Fatalf("expected rename name to be passed to service, got %q", svc.renameName)
	}
}

func TestRenameDocument_EmptyName_Returns400(t *testing.T) {
	svc := &fakeSvc{renameErr: domain.ErrInvalidName}
	mux := newMux(t, svc)

	body := []byte(`{"name":"   "}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/documents/doc_1", bytes.NewReader(body))
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRenameDocument_NameTooLong_Returns400(t *testing.T) {
	svc := &fakeSvc{}
	mux := newMux(t, svc)

	body := []byte(`{"name":"` + strings.Repeat("a", 256) + `"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/documents/doc_1", bytes.NewReader(body))
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if svc.renameName != "" {
		t.Fatalf("rename service should not be called for invalid payload, got %q", svc.renameName)
	}
}

func TestCreateCheckpoint_EmptyLabel_Returns400(t *testing.T) {
	svc := &fakeSvc{}
	mux := newMux(t, svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/checkpoints", bytes.NewReader([]byte(`{"label":"   "}`)))
	withAuthHeaders(req, "document_filler")
	req.SetPathValue("id", "doc_1")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if svc.checkpointCalls != 0 {
		t.Fatalf("create checkpoint service should not be called, calls=%d", svc.checkpointCalls)
	}
}

func TestCreateCheckpoint_LabelTooLong_Returns400(t *testing.T) {
	svc := &fakeSvc{}
	mux := newMux(t, svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/checkpoints", bytes.NewReader([]byte(`{"label":"`+strings.Repeat("b", 256)+`"}`)))
	withAuthHeaders(req, "document_filler")
	req.SetPathValue("id", "doc_1")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if svc.checkpointCalls != 0 {
		t.Fatalf("create checkpoint service should not be called, calls=%d", svc.checkpointCalls)
	}
}

func TestFinalizeDocument_MissingIdempotencyKey_Returns400(t *testing.T) {
	mux := newMux(t, &fakeSvc{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/finalize", nil)
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestFinalizeDocument_InvalidIdempotencyKey_Returns400(t *testing.T) {
	mux := newMux(t, &fakeSvc{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/finalize", nil)
	withAuthHeaders(req, "document_filler")
	req.Header.Set("Idempotency-Key", "not-a-uuid")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestFinalizeDocument_AllowsMissingRevisionTitleAtHTTPBoundary(t *testing.T) {
	mux := newMux(t, &fakeSvc{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/finalize", bytes.NewReader([]byte(`{}`)))
	withAuthHeaders(req, "document_filler")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code == http.StatusBadRequest && strings.Contains(rr.Body.String(), "revisionTitle") {
		t.Fatalf("revisionTitle should be conditionally validated by the submit service, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFinalizeDocument_ProfileNotFoundUsesProblemEnvelope(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT revision_version, revision_number, controlled_document_id::text").
		WithArgs("doc_1", "tenant_1").
		WillReturnRows(sqlmock.NewRows([]string{"revision_version", "revision_number", "controlled_document_id"}).
			AddRow(int64(0), int64(1), nil))

	h := httphandler.NewHandlerWithSubmitAndFinalizeStore(&fakeSvc{}, db, &fakeApprovalSubmitter{}, &fakeFinalizeIdempotencyStore{})
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/finalize", bytes.NewReader([]byte(`{"revisionTitle":"Ajuste"}`)))
	withAuthHeaders(req, "document_filler")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("want application/problem+json, got %s", ct)
	}
	if !strings.Contains(rr.Body.String(), "profile_not_found") {
		t.Fatalf("expected profile_not_found problem, got %s", rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestFinalizeDocument_ReplayReturnsCreatedAndHeader(t *testing.T) {
	key := "11111111-1111-4111-8111-111111111111"
	body := []byte(`{"instanceId":"inst_1"}`)

	submitter := &fakeApprovalSubmitter{}
	store := &fakeFinalizeIdempotencyStore{replay: &idempotency.Replay{Status: http.StatusCreated, Body: body}}
	h := httphandler.NewHandlerWithSubmitAndFinalizeStore(&fakeSvc{}, &sql.DB{}, submitter, store)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/finalize", bytes.NewReader([]byte(`{"revisionTitle":"Ajuste operacional"}`)))
	withAuthHeaders(req, "document_filler")
	req.Header.Set("Idempotency-Key", key)
	if tid, err := tenant.FromContext(req.Context()); err != nil || tid == "" {
		t.Fatalf("tenant missing before ServeHTTP: tid=%q err=%v", tid, err)
	}
	if uid := iamdomain.UserIDFromContext(req.Context()); uid == "" {
		t.Fatalf("user missing before ServeHTTP")
	}
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Idempotent-Replay"); got != "true" {
		t.Fatalf("expected Idempotent-Replay=true, got %q", got)
	}
	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got, _ := out["instanceId"].(string); got != "inst_1" {
		t.Fatalf("expected instanceId=inst_1, got %q", got)
	}
	if submitter.called {
		t.Fatalf("submit service should not be called on replay")
	}
}

func TestFinalizeDocument_ContentHashQueryNoRows_ContinuesWithEmptyHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT revision_version, revision_number, controlled_document_id::text").
		WithArgs("doc_1", "tenant_1").
		WillReturnRows(sqlmock.NewRows([]string{"revision_version", "revision_number", "controlled_document_id"}).
			AddRow(int64(0), int64(1), "cd_1"))
	mock.ExpectQuery("SELECT profile_code FROM controlled_documents WHERE id = \\$1 AND tenant_id = \\$2").
		WithArgs("cd_1", "tenant_1").
		WillReturnRows(sqlmock.NewRows([]string{"profile_code"}).AddRow("QA"))
	mock.ExpectQuery("SELECT id FROM approval_routes").
		WithArgs("tenant_1", "QA").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("route_1"))
	mock.ExpectQuery("SELECT COALESCE\\(content_hash, ''\\) FROM document_revisions").
		WithArgs("doc_1").
		WillReturnError(sql.ErrNoRows)

	submitter := &fakeApprovalSubmitter{}
	h := httphandler.NewHandlerWithSubmitAndFinalizeStore(&fakeSvc{}, db, submitter, &fakeFinalizeIdempotencyStore{})
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/finalize", bytes.NewReader([]byte(`{"revisionTitle":"Ajuste operacional"}`)))
	withAuthHeaders(req, "document_filler")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !submitter.called {
		t.Fatal("expected submit service to be called")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestFinalizeDocument_ContentHashQueryError_Returns500(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT revision_version, revision_number, controlled_document_id::text").
		WithArgs("doc_1", "tenant_1").
		WillReturnRows(sqlmock.NewRows([]string{"revision_version", "revision_number", "controlled_document_id"}).
			AddRow(int64(0), int64(1), "cd_1"))
	mock.ExpectQuery("SELECT profile_code FROM controlled_documents WHERE id = \\$1 AND tenant_id = \\$2").
		WithArgs("cd_1", "tenant_1").
		WillReturnRows(sqlmock.NewRows([]string{"profile_code"}).AddRow("QA"))
	mock.ExpectQuery("SELECT id FROM approval_routes").
		WithArgs("tenant_1", "QA").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("route_1"))
	mock.ExpectQuery("SELECT COALESCE\\(content_hash, ''\\) FROM document_revisions").
		WithArgs("doc_1").
		WillReturnError(errors.New("db offline"))

	submitter := &fakeApprovalSubmitter{}
	h := httphandler.NewHandlerWithSubmitAndFinalizeStore(&fakeSvc{}, db, submitter, &fakeFinalizeIdempotencyStore{})
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/finalize", bytes.NewReader([]byte(`{"revisionTitle":"Ajuste operacional"}`)))
	withAuthHeaders(req, "document_filler")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", rr.Code, rr.Body.String())
	}
	if submitter.called {
		t.Fatal("submit service should not be called on content-hash query failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestDuplicateDocument_InternalError_DoesNotLeakDetail(t *testing.T) {
	mux := newMux(t, &fakeSvc{duplicateErr: errors.New("sensitive db detail")})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/duplicate", nil)
	req.SetPathValue("id", "doc_1")
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", rr.Code, rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "sensitive db detail") {
		t.Fatalf("response must not leak internal error details: %s", rr.Body.String())
	}
}

func TestFinalizeDocument_SubmitPathDoesNotCallLegacyFinalize(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}

	body := string(src)
	start := strings.Index(body, "func (h *Handler) finalizeDocument")
	if start == -1 {
		t.Fatal("finalizeDocument not found")
	}
	end := strings.Index(body[start:], "func (h *Handler) archiveDocument")
	if end == -1 {
		t.Fatal("archiveDocument not found")
	}
	finalize := body[start : start+end]

	submitCall := strings.Index(finalize, "h.submitSvc.SubmitRevisionForReview")
	if submitCall == -1 {
		t.Fatal("finalizeDocument submit path not found")
	}
	legacyFinalize := strings.Index(finalize[submitCall:], "h.svc.Finalize")
	if legacyFinalize != -1 {
		t.Fatal("submit-backed finalize path must not call legacy Finalize after approval submit")
	}
}

func TestListRevisionHistory_ReturnsGovernedItems(t *testing.T) {
	mux := newMux(t, &fakeSvc{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/doc_1/revision-history", nil)
	req.SetPathValue("id", "doc_1")
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var body struct {
		Items []struct {
			DocumentID     string `json:"documentId"`
			RevisionNumber int64  `json:"revisionNumber"`
			RevisionTitle  string `json:"revisionTitle"`
			Status         string `json:"status"`
			IsCurrent      bool   `json:"isCurrent"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(body.Items))
	}
	if body.Items[0].RevisionNumber != 2 {
		t.Fatalf("revision number = %d", body.Items[0].RevisionNumber)
	}
	if body.Items[0].RevisionTitle != "Ajuste operacional" {
		t.Fatalf("revision title = %q", body.Items[0].RevisionTitle)
	}
	if body.Items[0].Status != "draft" {
		t.Fatalf("status = %q", body.Items[0].Status)
	}
	if !body.Items[0].IsCurrent {
		t.Fatal("expected current revision item")
	}
}
