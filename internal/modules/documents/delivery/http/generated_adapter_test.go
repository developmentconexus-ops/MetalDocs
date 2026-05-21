package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	documentsapi "metaldocs/internal/modules/documents/api"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

func TestGeneratedServerAdapter_ForwardsToLegacyMux(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		path    string
		pattern string
		invoke  func(*GeneratedServerAdapter, http.ResponseWriter, *http.Request)
	}{
		{
			name:    "ListDocuments",
			method:  http.MethodGet,
			path:    "/api/v1/documents",
			pattern: "GET /api/v1/documents",
			invoke: func(a *GeneratedServerAdapter, w http.ResponseWriter, r *http.Request) {
				a.ListDocuments(w, r, documentsapi.ListDocumentsParams{})
			},
		},
		{
			name:    "FinalizeDocument",
			method:  http.MethodPost,
			path:    "/api/v1/documents/11111111-1111-4111-8111-111111111111/finalize",
			pattern: "POST /api/v1/documents/{id}/finalize",
			invoke: func(a *GeneratedServerAdapter, w http.ResponseWriter, r *http.Request) {
				var id openapi_types.UUID
				a.FinalizeDocument(w, r, id, documentsapi.FinalizeDocumentParams{})
			},
		},
		{
			name:    "UpdateDocumentComment",
			method:  http.MethodPatch,
			path:    "/api/v1/documents/11111111-1111-4111-8111-111111111111/comments/7",
			pattern: "PATCH /api/v1/documents/{id}/comments/{libraryID}",
			invoke: func(a *GeneratedServerAdapter, w http.ResponseWriter, r *http.Request) {
				var id openapi_types.UUID
				a.UpdateDocumentComment(w, r, id, 7)
			},
		},
		{
			name:    "GetDocumentPlaceholderOptions",
			method:  http.MethodGet,
			path:    "/api/v1/documents/11111111-1111-4111-8111-111111111111/placeholder-options/p1",
			pattern: "GET /api/v1/documents/{id}/placeholder-options/{pid}",
			invoke: func(a *GeneratedServerAdapter, w http.ResponseWriter, r *http.Request) {
				var id openapi_types.UUID
				a.GetDocumentPlaceholderOptions(w, r, id, "p1")
			},
		},
		{
			name:    "ViewDocument",
			method:  http.MethodGet,
			path:    "/api/v1/documents/11111111-1111-4111-8111-111111111111/view",
			pattern: "GET /api/v1/documents/{id}/view",
			invoke: func(a *GeneratedServerAdapter, w http.ResponseWriter, r *http.Request) {
				var id openapi_types.UUID
				a.ViewDocument(w, r, id)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			legacy := http.NewServeMux()
			legacy.HandleFunc(tc.pattern, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Forwarded", "yes")
				w.WriteHeader(http.StatusNoContent)
			})

			adapter := NewGeneratedServerAdapter(legacy)
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			tc.invoke(adapter, rr, req)

			if rr.Code != http.StatusNoContent {
				t.Fatalf("expected 204, got %d", rr.Code)
			}
			if got := rr.Header().Get("X-Forwarded"); got != "yes" {
				t.Fatalf("expected forwarded header, got %q", got)
			}
		})
	}
}
