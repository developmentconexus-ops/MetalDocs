package http

import (
	"net/http"

	documentsapi "metaldocs/internal/modules/documents/api"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Compile-time assertion: Handler must satisfy the generated ServerInterface.
var _ documentsapi.ServerInterface = (*Handler)(nil)

// The following 30 methods implement documentsapi.ServerInterface.
// Each ignores the typed path/query params (the raw handler re-parses from r)
// and delegates directly to the existing func(w,r) handler — identical
// semantics to the previous GeneratedServerAdapter shim, but without the
// intermediate legacy mux.

func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request, _ documentsapi.ListDocumentsParams) {
	h.listDocuments(w, r)
}

func (h *Handler) DocumentStats(w http.ResponseWriter, r *http.Request, _ documentsapi.DocumentStatsParams) {
	h.documentStats(w, r)
}

func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.getDocument(w, r)
}

func (h *Handler) RenameDocument(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.renameDocument(w, r)
}

func (h *Handler) ArchiveDocument(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.archiveDocument(w, r)
}

func (h *Handler) CommitDocumentAutosave(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.commitAutosave(w, r)
}

func (h *Handler) PresignDocumentAutosave(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.presignAutosave(w, r)
}

func (h *Handler) ListDocumentCheckpoints(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.listCheckpoints(w, r)
}

func (h *Handler) CreateDocumentCheckpoint(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.createCheckpoint(w, r)
}

func (h *Handler) RestoreDocumentCheckpoint(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ int) {
	h.restoreCheckpoint(w, r)
}

func (h *Handler) ListDocumentComments(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.listComments(w, r)
}

func (h *Handler) CreateDocumentComment(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.createComment(w, r)
}

func (h *Handler) DeleteDocumentComment(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ int) {
	h.deleteComment(w, r)
}

func (h *Handler) UpdateDocumentComment(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ int) {
	h.updateComment(w, r)
}

func (h *Handler) DuplicateDocument(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.duplicateDocument(w, r)
}

func (h *Handler) GetDocumentDocxURL(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.export.exportDocxURL(w, r)
}

func (h *Handler) ExportDocumentPDF(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.export.exportPDF(w, r)
}

func (h *Handler) GetDocumentFillInSchema(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.fillIn.GetFillInSchema(w, r)
}

func (h *Handler) FinalizeDocument(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ documentsapi.FinalizeDocumentParams) {
	h.finalizeDocument(w, r)
}

func (h *Handler) GetDocumentPlaceholderOptions(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ string) {
	h.placeholderOpts.HandleGetOptions(w, r)
}

func (h *Handler) ListDocumentPlaceholderValues(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.fillIn.ListPlaceholderValues(w, r)
}

func (h *Handler) PutDocumentPlaceholderValue(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ string) {
	h.fillIn.PutPlaceholderValue(w, r)
}

func (h *Handler) ReconstructDocument(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.reconstruct.HandleReconstruct(w, r)
}

func (h *Handler) GetDocumentRevisionHistory(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.listRevisionHistory(w, r)
}

func (h *Handler) GetDocumentRevisionUrl(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID, _ openapi_types.UUID) {
	h.signedRevisionURL(w, r)
}

func (h *Handler) AcquireDocumentSession(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.acquireSession(w, r)
}

func (h *Handler) ForceReleaseDocumentSession(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.forceReleaseSession(w, r)
}

func (h *Handler) HeartbeatDocumentSession(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.heartbeatSession(w, r)
}

func (h *Handler) ReleaseDocumentSession(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.releaseSession(w, r)
}

func (h *Handler) ViewDocument(w http.ResponseWriter, r *http.Request, _ openapi_types.UUID) {
	h.view.HandleView(w, r)
}
