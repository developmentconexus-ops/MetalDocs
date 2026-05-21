package documents

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/application"
	dhttp "metaldocs/internal/modules/documents/delivery/http"
	"metaldocs/internal/modules/documents/domain"
	documentshttp "metaldocs/internal/modules/documents/http"
	"metaldocs/internal/modules/documents/repository"
	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesdomain "metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/ratelimit"
	"metaldocs/internal/platform/tenant"
)

type moduleTestService struct{}

func (s *moduleTestService) CreateDocument(context.Context, application.CreateDocumentInput) (*application.CreateDocumentResult, error) {
	return &application.CreateDocumentResult{DocumentID: "doc_1", InitialRevisionID: "rev_1", SessionID: "sess_1"}, nil
}
func (s *moduleTestService) GetDocument(context.Context, string, string) (*domain.Document, error) {
	return &domain.Document{ID: "doc_1", Name: "Doc", Status: domain.DocStatusDraft, FormDataJSON: []byte(`{}`), Code: "DOC-1"}, nil
}
func (s *moduleTestService) DuplicateDocument(context.Context, string, string, string) (*application.CreateDocumentResult, error) {
	return &application.CreateDocumentResult{DocumentID: "doc_2", InitialRevisionID: "rev_2", SessionID: "sess_2"}, nil
}
func (s *moduleTestService) RenameDocument(context.Context, string, string, string, string) error {
	return nil
}
func (s *moduleTestService) ListDocuments(context.Context, string) ([]domain.Document, error) {
	return nil, nil
}
func (s *moduleTestService) ListDocumentsForUser(context.Context, string, string) ([]domain.Document, error) {
	return nil, nil
}
func (s *moduleTestService) ListDocumentsPaginated(context.Context, string, string, application.ListOptions) ([]*domain.Document, int64, error) {
	return []*domain.Document{{ID: "doc_1"}}, 1, nil
}
func (s *moduleTestService) DocumentStats(context.Context, string, string, application.ListOptions) (*application.DocumentStats, error) {
	return &application.DocumentStats{ByStatus: map[string]int64{"draft": 1}}, nil
}
func (s *moduleTestService) IsDocumentOwner(context.Context, string, string, string) (bool, error) {
	return true, nil
}
func (s *moduleTestService) AcquireSession(context.Context, string, string, string) (*domain.Session, bool, error) {
	return &domain.Session{ID: "sess_1", UserID: "user_1", Status: domain.SessionActive}, false, nil
}
func (s *moduleTestService) HeartbeatSession(context.Context, string, string) error { return nil }
func (s *moduleTestService) ReleaseSession(context.Context, string, string, string, string) error {
	return nil
}
func (s *moduleTestService) ForceReleaseSession(context.Context, string, string, string, string) error {
	return nil
}
func (s *moduleTestService) PresignAutosave(context.Context, application.PresignAutosaveCmd) (*application.PresignAutosaveResult, error) {
	return &application.PresignAutosaveResult{UploadURL: "https://example/upload", PendingUploadID: "pending_1", ExpiresAt: time.Now().Add(time.Minute)}, nil
}
func (s *moduleTestService) CommitAutosave(context.Context, application.CommitAutosaveCmd) (*application.CommitResult, error) {
	return &application.CommitResult{RevisionID: "rev_2", RevisionNum: 2}, nil
}
func (s *moduleTestService) SyncArtifactMetadata(context.Context, application.SyncArtifactMetadataCmd) (*application.CommitResult, error) {
	return &application.CommitResult{RevisionID: "rev_2", RevisionNum: 2}, nil
}
func (s *moduleTestService) CreateCheckpoint(context.Context, string, string, string, string) (*domain.Checkpoint, error) {
	return &domain.Checkpoint{ID: "cp_1", VersionNum: 1}, nil
}
func (s *moduleTestService) ListCheckpoints(context.Context, string, string) ([]domain.Checkpoint, error) {
	return []domain.Checkpoint{{ID: "cp_1", VersionNum: 1}}, nil
}
func (s *moduleTestService) ListRevisionHistory(context.Context, string, string) ([]domain.RevisionHistoryItem, error) {
	return []domain.RevisionHistoryItem{{DocumentID: "doc_1", RevisionNumber: 0, RevisionTitle: "Criacao", Status: domain.DocStatusDraft, IsCurrent: true}}, nil
}
func (s *moduleTestService) RestoreCheckpoint(context.Context, string, string, string, int) (*application.RestoreResult, error) {
	return &application.RestoreResult{NewRevisionID: "rev_3", NewRevisionNum: 3}, nil
}
func (s *moduleTestService) Finalize(context.Context, string, string, string) error { return nil }
func (s *moduleTestService) Archive(context.Context, string, string, string) error  { return nil }
func (s *moduleTestService) SignedRevisionURL(context.Context, string, string, string) (string, error) {
	return "https://example/rev", nil
}
func (s *moduleTestService) ListDocumentComments(context.Context, string, string, string) ([]domain.Comment, error) {
	return nil, nil
}
func (s *moduleTestService) AddDocumentComment(context.Context, string, string, string, string, domain.CommentCreateInput) (*domain.Comment, error) {
	return &domain.Comment{}, nil
}
func (s *moduleTestService) UpdateDocumentComment(context.Context, string, string, string, int, domain.CommentUpdateInput) (*domain.Comment, error) {
	return &domain.Comment{}, nil
}
func (s *moduleTestService) DeleteDocumentComment(context.Context, string, string, string, int) error {
	return nil
}

type moduleFillInService struct{}

func (s *moduleFillInService) SetPlaceholderValue(context.Context, string, string, string, string, string) error {
	return nil
}
func (s *moduleFillInService) GetPlaceholderValues(context.Context, string, string) ([]repository.PlaceholderValue, error) {
	return nil, nil
}
func (s *moduleFillInService) GetFillInSchema(context.Context, string, string) ([]templatesdomain.Placeholder, error) {
	return nil, nil
}

type moduleSchemaReader struct{}

func (r moduleSchemaReader) LoadPlaceholderSchema(context.Context, string, string) ([]templatesdomain.Placeholder, error) {
	return []templatesdomain.Placeholder{{ID: "p-sel", Type: templatesdomain.PHSelect, Options: []string{"A", "B"}}}, nil
}

type moduleIAMReader struct{}

func (r moduleIAMReader) ListUserOptions(context.Context, string) ([]documentshttp.UserOptionView, error) {
	return []documentshttp.UserOptionView{{UserID: "u1", DisplayName: "User 1"}}, nil
}

func newWrapperTestModule() *Module {
	return &Module{
		Handler:                   dhttp.NewHandler(&moduleTestService{}),
		FillInHandler:             documentshttp.NewFillInHandler(&moduleFillInService{}),
		PlaceholderOptionsHandler: documentshttp.NewPlaceholderOptionsHandler(moduleSchemaReader{}, moduleIAMReader{}),
	}
}

func withModuleAuth(req *http.Request, roles string) *http.Request {
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant_1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "user_1", []iamdomain.Role{iamdomain.Role(roles)}))
	req.Header.Set("X-User-Roles", roles)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestRegisterRoutes_WrapperForwardsRepresentativeRoutes(t *testing.T) {
	mod := newWrapperTestModule()
	mux := http.NewServeMux()
	mod.RegisterRoutes(mux)
	docID := "11111111-1111-4111-8111-111111111111"

	req := withModuleAuth(httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil), "document_filler")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	req = withModuleAuth(httptest.NewRequest(http.MethodGet, "/api/v1/documents/"+docID+"/placeholder-options/p-sel", nil), "document_filler")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRegisterRoutesWithRateLimit_WrapperForwardingStillWorks(t *testing.T) {
	mod := newWrapperTestModule()
	mux := http.NewServeMux()
	rl := ratelimit.New(ratelimit.DefaultConfig())
	docID := "11111111-1111-4111-8111-111111111111"
	mod.RegisterRoutesWithRateLimit(mux, rl, func(r *http.Request) string {
		return iamdomain.UserIDFromContext(r.Context())
	})

	reqBody := strings.NewReader(`{"session_id":"sess_1","base_revision_id":"rev_1","content_hash":"abc"}`)
	req := withModuleAuth(httptest.NewRequest(http.MethodPost, "/api/v1/documents/"+docID+"/autosave/presign", reqBody), "document_filler")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRegisterRoutes_WrapperReturnsValidationErrorOnMalformedPathParam(t *testing.T) {
	mod := newWrapperTestModule()
	mux := http.NewServeMux()
	mod.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/documents/not-a-uuid/comments/not-an-int", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("expected problem+json, got %s", ct)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode problem body: %v", err)
	}
	if got, _ := body["code"].(string); got != "VALIDATION_ERROR" {
		t.Fatalf("expected code VALIDATION_ERROR, got %q", got)
	}
}
