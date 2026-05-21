package http

import (
	"net/http"

	documentsapi "metaldocs/internal/modules/documents/api"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// GeneratedServerAdapter bridges generated wrapper methods to the existing mux-mounted handlers.
type GeneratedServerAdapter struct {
	legacy http.Handler
}

func NewGeneratedServerAdapter(legacy http.Handler) *GeneratedServerAdapter {
	return &GeneratedServerAdapter{legacy: legacy}
}

func (a *GeneratedServerAdapter) forward(w http.ResponseWriter, r *http.Request) {
	a.legacy.ServeHTTP(w, r)
}

func (a *GeneratedServerAdapter) ListDocuments(w http.ResponseWriter, r *http.Request, _ documentsapi.ListDocumentsParams) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) DocumentStats(w http.ResponseWriter, r *http.Request, _ documentsapi.DocumentStatsParams) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) GetDocument(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) RenameDocument(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) ArchiveDocument(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) CommitDocumentAutosave(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) PresignDocumentAutosave(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) ListDocumentCheckpoints(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) CreateDocumentCheckpoint(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) RestoreDocumentCheckpoint(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ int) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) ListDocumentComments(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) CreateDocumentComment(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) DeleteDocumentComment(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ int) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) UpdateDocumentComment(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ int) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) DuplicateDocument(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) GetDocumentDocxURL(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) ExportDocumentPDF(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) GetDocumentFillInSchema(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) FinalizeDocument(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ documentsapi.FinalizeDocumentParams) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) GetDocumentPlaceholderOptions(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ string) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) ListDocumentPlaceholderValues(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) PutDocumentPlaceholderValue(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ string) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) ReconstructDocument(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) GetDocumentRevisionHistory(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) GetDocumentRevisionUrl(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) AcquireDocumentSession(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) ForceReleaseDocumentSession(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) HeartbeatDocumentSession(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) ReleaseDocumentSession(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}

func (a *GeneratedServerAdapter) ViewDocument(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	a.forward(w, r)
}
