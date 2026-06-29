package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/tenant"
)

func TestGeneratedTemplatesRoutes_ContractHappyPaths(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["11111111-1111-1111-1111-111111111111"] = &domain.Template{
		ID:            "11111111-1111-1111-1111-111111111111",
		TenantID:      "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		Key:           "contract",
		Name:          "Contract",
		LatestVersion: 1,
		CreatedBy:     "user-a",
		CreatedAt:     fakeClock{}.Now(),
	}
	repo.versions["22222222-2222-4222-8222-222222222222"] = &domain.TemplateVersion{
		ID:             "22222222-2222-4222-8222-222222222222",
		TemplateID:     "11111111-1111-1111-1111-111111111111",
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/11111111-1111-1111-1111-111111111111/versions/1.docx",
		ContentHash:    "hash_abc",
		AuthorID:       "original-author",
		CreatedAt:      fakeClock{}.Now(),
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	tests := []struct {
		name   string
		method string
		path   string
		body   []byte
		want   int
	}{
		{name: "listTemplates", method: http.MethodGet, path: "/api/v1/templates", want: http.StatusOK},
		{name: "createTemplate", method: http.MethodPost, path: "/api/v1/templates", body: jsonBody(t, map[string]any{"key": "new-contract", "name": "New Contract", "description": "Default"}), want: http.StatusCreated},
		{name: "getTemplateVersion", method: http.MethodGet, path: "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1", want: http.StatusOK},
		{name: "presignTemplateDocxUploadUrl", method: http.MethodPost, path: "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/docx-upload-url", want: http.StatusOK},
		{name: "presignTemplateSchemaUploadUrl", method: http.MethodPost, path: "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/schema-upload-url", want: http.StatusOK},
{name: "updateTemplateSchema", method: http.MethodPut, path: "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/schema", body: jsonBody(t, map[string]any{
			"metadata_schema":       map[string]any{"retention_days": 1},
			"placeholder_schema":    []any{},
			"expected_lock_version": 0,
		}), want: http.StatusOK},
		// publishTemplateVersion requires a real DB transaction — excluded from no-DB contract harness.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(tt.body))
			withHeaders(req)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)
			if rr.Code != tt.want {
				t.Fatalf("expected %d, got %d body=%s", tt.want, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestGeneratedTemplatesRoutes_RejectInvalidBodies(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["11111111-1111-1111-1111-111111111111"] = &domain.Template{ID: "11111111-1111-1111-1111-111111111111", TenantID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}
	repo.versions["22222222-2222-4222-8222-222222222222"] = &domain.TemplateVersion{
		ID:             "22222222-2222-4222-8222-222222222222",
		TemplateID:     "11111111-1111-1111-1111-111111111111",
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/11111111-1111-1111-1111-111111111111/versions/1.docx",
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "create unknown field", method: http.MethodPost, path: "/api/v1/templates", body: `{"key":"contract","name":"Contract","extra":true}`},
		{name: "create missing key", method: http.MethodPost, path: "/api/v1/templates", body: `{"name":"Contract"}`},
		{name: "publish unknown field", method: http.MethodPost, path: "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/publish", body: `{"schema_key":"s","extra":true}`},
		{name: "schema missing expected_lock_version", method: http.MethodPut, path: "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/schema", body: `{"metadata_schema":{},"placeholder_schema":[]}`},
		{name: "schema negative expected_lock_version", method: http.MethodPut, path: "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/schema", body: `{"metadata_schema":{},"placeholder_schema":[],"expected_lock_version":-1}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			withHeaders(req)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestUpdateTemplateSchema_StaleLockVersion_412(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["11111111-1111-1111-1111-111111111111"] = &domain.Template{
		ID:       "11111111-1111-1111-1111-111111111111",
		TenantID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
	}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:            "ver-1",
		TemplateID:    "11111111-1111-1111-1111-111111111111",
		VersionNumber: 1,
		Status:        domain.VersionStatusDraft,
	}
	repo.lockVersions["ver-1"] = 5
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	body := jsonBody(t, map[string]any{
		"metadata_schema":       map[string]any{"retention_days": 1},
		"placeholder_schema":    []any{},
		"expected_lock_version": 1,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/schema", bytes.NewReader(body))
	withHeaders(req)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412, got %d body=%s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("content-type")
	if ct == "" || ct[:len("application/problem+json")] != "application/problem+json" {
		t.Fatalf("expected application/problem+json, got %q", ct)
	}
	var problem struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decode problem: %v", err)
	}
	if problem.Code != "CONCURRENT_MODIFICATION" {
		t.Fatalf("expected code CONCURRENT_MODIFICATION, got %q", problem.Code)
	}
}

func TestGeneratedTemplatesRoutes_RejectValidation(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["11111111-1111-1111-1111-111111111111"] = &domain.Template{ID: "11111111-1111-1111-1111-111111111111", TenantID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}
	repo.versions["22222222-2222-4222-8222-222222222222"] = &domain.TemplateVersion{
		ID:             "22222222-2222-4222-8222-222222222222",
		TemplateID:     "11111111-1111-1111-1111-111111111111",
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/11111111-1111-1111-1111-111111111111/versions/1.docx",
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "listTemplates invalid limit", method: http.MethodGet, path: "/api/v1/templates?limit=bad"},
		{name: "createTemplate missing key", method: http.MethodPost, path: "/api/v1/templates", body: `{"name":"Contract"}`},
		{name: "getTemplateVersion invalid version", method: http.MethodGet, path: "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/not-int"},
		{name: "presignTemplateDocxUploadUrl invalid version", method: http.MethodPost, path: "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/not-int/docx-upload-url"},
		{name: "presignTemplateSchemaUploadUrl invalid version", method: http.MethodPost, path: "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/not-int/schema-upload-url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			withHeaders(req)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestGeneratedTemplatesRoutes_IdempotencyKeyRequired(t *testing.T) {
	repo := newFakeRepo()
	repo.templates["11111111-1111-1111-1111-111111111111"] = &domain.Template{ID: "11111111-1111-1111-1111-111111111111", TenantID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}
	repo.versions["22222222-2222-4222-8222-222222222222"] = &domain.TemplateVersion{
		ID:             "22222222-2222-4222-8222-222222222222",
		TemplateID:     "11111111-1111-1111-1111-111111111111",
		VersionNumber:  1,
		Status:         domain.VersionStatusDraft,
		DocxStorageKey: "templates/11111111-1111-1111-1111-111111111111/versions/1.docx",
		ContentHash:    "hash_abc",
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "create", method: http.MethodPost, path: "/api/v1/templates", body: `{"key":"k1","name":"n1"}`},
		{name: "publish", method: http.MethodPost, path: "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/publish", body: `{"schema_key":"s"}`},
		{name: "submit", method: http.MethodPost, path: "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/submit", body: `{}`},
		{name: "review", method: http.MethodPost, path: "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/review", body: `{"accept":true}`},
		{name: "approve", method: http.MethodPost, path: "/api/v1/templates/11111111-1111-1111-1111-111111111111/versions/1/approve", body: `{"accept":true}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("content-type", "application/json")
			*req = *req.WithContext(tenant.WithTenantID(req.Context(), "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
			*req = *req.WithContext(iamdomain.WithAuthContext(req.Context(), "user-a", []iamdomain.Role{}))
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
			}
		})
	}
}

func jsonBody(t *testing.T, body map[string]any) []byte {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	return raw
}
