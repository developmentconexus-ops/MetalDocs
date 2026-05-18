package http_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

	renameErr  error
	renameName string

	listPaginatedItems []*domain.Document
	listPaginatedTotal int64
	listPaginatedErr   error
	listPaginatedOpts  application.ListOptions
	listPaginatedUser  string

	statsResult *application.DocumentStats
	statsErr    error
}

var _ httphandler.Service = (*fakeSvc)(nil)

func (f *fakeSvc) CreateDocument(_ context.Context, _ application.CreateDocumentInput) (*application.CreateDocumentResult, error) {
	return &application.CreateDocumentResult{DocumentID: "doc_1", InitialRevisionID: "rev_1", SessionID: "sess_1"}, nil
}

func (f *fakeSvc) GetDocument(_ context.Context, _, _ string) (*domain.Document, error) {
	return &domain.Document{
		ID:                "doc_1",
		Name:              "Doc",
		Status:            domain.DocStatusDraft,
		FormDataJSON:      []byte(`{"foo":"bar"}`),
		CurrentRevisionID: "rev_1",
		RevisionVersion:   1,
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

func (f *fakeSvc) CommitAutosave(_ context.Context, _ application.CommitAutosaveCmd) (*application.CommitResult, error) {
	if f.commitErr != nil {
		return nil, f.commitErr
	}
	if f.commitResult == nil {
		return &application.CommitResult{RevisionID: "rev_2", RevisionNum: 2}, nil
	}
	return f.commitResult, nil
}

func (f *fakeSvc) CreateCheckpoint(_ context.Context, _, _, _, _ string) (*domain.Checkpoint, error) {
	return &domain.Checkpoint{ID: "cp_1", VersionNum: 1}, nil
}

func (f *fakeSvc) ListCheckpoints(_ context.Context, _, _ string) ([]domain.Checkpoint, error) {
	return []domain.Checkpoint{{ID: "cp_1", VersionNum: 1}}, nil
}

func (f *fakeSvc) RestoreCheckpoint(_ context.Context, _, _, _ string, _ int) (*application.RestoreResult, error) {
	return &application.RestoreResult{NewRevisionID: "rev_3", NewRevisionNum: 3}, nil
}

func (f *fakeSvc) DuplicateDocument(_ context.Context, _, _, _ string) (*application.CreateDocumentResult, error) {
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

func (f *fakeFinalizeIdempotencyStore) CheckReplay(_ context.Context, _, _, _, _ string) (*idempotency.Replay, error) {
	return f.replay, nil
}

func (f *fakeFinalizeIdempotencyStore) RecordReplay(_ context.Context, _, _, _, _ string, _ int, _ []byte) error {
	return nil
}

func withAuthHeaders(req *http.Request, roles string) {
	req.Header.Set("content-type", "application/json")
	*req = *req.WithContext(tenant.WithTenantID(req.Context(), "tenant_1"))
	*req = *req.WithContext(iamdomain.WithAuthContext(req.Context(), "user_1", []iamdomain.Role{}))
	req.Header.Set("X-User-Roles", roles)
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

func TestFinalizeDocument_RequiresRevisionTitle(t *testing.T) {
	mux := newMux(t, &fakeSvc{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/finalize", bytes.NewReader([]byte(`{}`)))
	withAuthHeaders(req, "document_filler")
	req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("expected problem content-type, got %q", ct)
	}
	if !strings.Contains(rr.Body.String(), "revisionTitle") {
		t.Fatalf("expected revisionTitle validation error, got %s", rr.Body.String())
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

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/finalize", nil)
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
