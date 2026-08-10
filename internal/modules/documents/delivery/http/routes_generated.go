package http

import (
	"net/http"
	"strconv"

	documentsapi "metaldocs/internal/modules/documents/api"
	"metaldocs/internal/platform/problem"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Compile-time assertion: Handler must satisfy the generated ServerInterface.
var _ documentsapi.ServerInterface = (*Handler)(nil)

// The following 30 methods implement documentsapi.ServerInterface.
//
// ARC-02: each method consumes the generated wrapper's typed path params by
// writing the canonical string form back onto the request via r.SetPathValue
// before delegating — mirroring the established codebase precedent in
// internal/modules/controlleddocuments/delivery/http/routes.go — rather than
// having the private (w,r) handlers re-parse r.PathValue/r.URL themselves.
// This lets ~20 existing private handler bodies keep their (w,r)-only
// signatures (untouched call sites) while still consuming the
// wrapper-validated value instead of a second, redundant parse. listDocuments
// and documentStats are the two exceptions that take the typed params
// directly, since their query-derived application.ListOptions could not be
// reconstructed from path values alone.
//
// Sub-handlers reachable only via WithSubHandlers (export, fillIn,
// placeholderOpts, view, reconstruct) may be nil in tests/wiring that don't
// set them (see module_wrapper_test.go newWrapperTestModule, which wires
// export/view/reconstruct as nil). HandlerWithOptions registers every
// ServerInterface route unconditionally (no per-route opt-out), so the old
// `if h.export != nil { register }` conditional-registration guards are
// replaced here with nil-guards that return an RFC 9457 404 instead of
// panicking on a nil dereference.

// ListDocuments implements documentsapi.ServerInterface for GET /documents,
// delegating to the private listDocuments handler.
func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request, params documentsapi.ListDocumentsParams) {
	h.listDocuments(w, r, params)
}

// DocumentStats implements documentsapi.ServerInterface for GET
// /documents/stats, delegating to the private documentStats handler.
func (h *Handler) DocumentStats(w http.ResponseWriter, r *http.Request, params documentsapi.DocumentStatsParams) {
	h.documentStats(w, r, params)
}

// GetDocument implements documentsapi.ServerInterface for GET
// /documents/{id}, delegating to the private getDocument handler.
func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.getDocument(w, r)
}

// RenameDocument implements documentsapi.ServerInterface for the rename
// route, delegating to the private renameDocument handler.
func (h *Handler) RenameDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.renameDocument(w, r)
}

// ArchiveDocument implements documentsapi.ServerInterface for the archive
// route, delegating to the private archiveDocument handler.
func (h *Handler) ArchiveDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.archiveDocument(w, r)
}

// CommitDocumentAutosave implements documentsapi.ServerInterface for the
// autosave commit route, delegating to the private commitAutosave handler.
func (h *Handler) CommitDocumentAutosave(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.commitAutosave(w, r)
}

// PresignDocumentAutosave implements documentsapi.ServerInterface for the
// autosave presign route, delegating to the private presignAutosave handler.
func (h *Handler) PresignDocumentAutosave(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.presignAutosave(w, r)
}

// ListDocumentCheckpoints implements documentsapi.ServerInterface for the
// checkpoint listing route, delegating to the private listCheckpoints handler.
func (h *Handler) ListDocumentCheckpoints(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.listCheckpoints(w, r)
}

// CreateDocumentCheckpoint implements documentsapi.ServerInterface for the
// checkpoint creation route, delegating to the private createCheckpoint handler.
func (h *Handler) CreateDocumentCheckpoint(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.createCheckpoint(w, r)
}

// RestoreDocumentCheckpoint implements documentsapi.ServerInterface for the
// checkpoint restore route, delegating to the private restoreCheckpoint handler.
func (h *Handler) RestoreDocumentCheckpoint(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, version int) {
	r.SetPathValue("id", id.String())
	r.SetPathValue("version", strconv.Itoa(version))
	h.restoreCheckpoint(w, r)
}

// ListDocumentComments implements documentsapi.ServerInterface for the
// comment listing route, delegating to the private listComments handler.
func (h *Handler) ListDocumentComments(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.listComments(w, r)
}

// CreateDocumentComment implements documentsapi.ServerInterface for the
// comment creation route, delegating to the private createComment handler.
func (h *Handler) CreateDocumentComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.createComment(w, r)
}

// DeleteDocumentComment implements documentsapi.ServerInterface for the
// comment deletion route, delegating to the private deleteComment handler.
func (h *Handler) DeleteDocumentComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, libraryID int) {
	r.SetPathValue("id", id.String())
	r.SetPathValue("library_id", strconv.Itoa(libraryID))
	h.deleteComment(w, r)
}

// UpdateDocumentComment implements documentsapi.ServerInterface for the
// comment update route, delegating to the private updateComment handler.
func (h *Handler) UpdateDocumentComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, libraryID int) {
	r.SetPathValue("id", id.String())
	r.SetPathValue("library_id", strconv.Itoa(libraryID))
	h.updateComment(w, r)
}

// DuplicateDocument implements documentsapi.ServerInterface for the
// duplicate route, delegating to the private duplicateDocument handler.
func (h *Handler) DuplicateDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.duplicateDocument(w, r)
}

// GetDocumentDocxURL implements documentsapi.ServerInterface for the signed
// DOCX URL route, delegating to the export sub-handler (404 when unwired).
func (h *Handler) GetDocumentDocxURL(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	if h.export == nil {
		writeNotFound(w, r)
		return
	}
	h.export.exportDocxURL(w, r)
}

// ExportDocumentPDF implements documentsapi.ServerInterface for the PDF
// export route, delegating to the export sub-handler (404 when unwired).
func (h *Handler) ExportDocumentPDF(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	if h.export == nil {
		writeNotFound(w, r)
		return
	}
	h.export.exportPDF(w, r)
}

// GetDocumentFillInSchema implements documentsapi.ServerInterface for the
// fill-in schema route, delegating to the fillIn sub-handler (404 when unwired).
func (h *Handler) GetDocumentFillInSchema(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	if h.fillIn == nil {
		writeNotFound(w, r)
		return
	}
	h.fillIn.GetFillInSchema(w, r)
}

// GetDocumentPlaceholderOptions implements documentsapi.ServerInterface for
// the placeholder options route, delegating to the placeholderOpts
// sub-handler (404 when unwired).
func (h *Handler) GetDocumentPlaceholderOptions(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, pid string) {
	r.SetPathValue("id", id.String())
	r.SetPathValue("pid", pid)
	if h.placeholderOpts == nil {
		writeNotFound(w, r)
		return
	}
	h.placeholderOpts.HandleGetOptions(w, r)
}

// ListDocumentPlaceholderValues implements documentsapi.ServerInterface for
// the placeholder value listing route, delegating to the fillIn sub-handler
// (404 when unwired).
func (h *Handler) ListDocumentPlaceholderValues(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	if h.fillIn == nil {
		writeNotFound(w, r)
		return
	}
	h.fillIn.ListPlaceholderValues(w, r)
}

// PutDocumentPlaceholderValue implements documentsapi.ServerInterface for
// the placeholder value write route, delegating to the fillIn sub-handler
// (404 when unwired).
func (h *Handler) PutDocumentPlaceholderValue(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, pid string) {
	r.SetPathValue("id", id.String())
	r.SetPathValue("pid", pid)
	if h.fillIn == nil {
		writeNotFound(w, r)
		return
	}
	h.fillIn.PutPlaceholderValue(w, r)
}

// ReconstructDocument implements documentsapi.ServerInterface for the
// reconstruction route, delegating to the reconstruct sub-handler (404 when unwired).
func (h *Handler) ReconstructDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	if h.reconstruct == nil {
		writeNotFound(w, r)
		return
	}
	h.reconstruct.HandleReconstruct(w, r)
}

// GetDocumentRevisionHistory implements documentsapi.ServerInterface for the
// revision history route, delegating to the private listRevisionHistory handler.
func (h *Handler) GetDocumentRevisionHistory(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.listRevisionHistory(w, r)
}

// GetDocumentRevisionUrl implements documentsapi.ServerInterface for the
// signed revision URL route, delegating to the private signedRevisionURL
// handler.
//
//nolint:revive // name pinned by OpenAPI operationId via generated ServerInterface
func (h *Handler) GetDocumentRevisionUrl(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, rid openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	r.SetPathValue("rid", rid.String())
	h.signedRevisionURL(w, r)
}

// AcquireDocumentSession implements documentsapi.ServerInterface for the
// session acquisition route, delegating to the private acquireSession handler.
func (h *Handler) AcquireDocumentSession(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.acquireSession(w, r)
}

// ForceReleaseDocumentSession implements documentsapi.ServerInterface for
// the admin force-release route, delegating to the private
// forceReleaseSession handler.
func (h *Handler) ForceReleaseDocumentSession(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.forceReleaseSession(w, r)
}

// HeartbeatDocumentSession implements documentsapi.ServerInterface for the
// session heartbeat route, delegating to the private heartbeatSession handler.
func (h *Handler) HeartbeatDocumentSession(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.heartbeatSession(w, r)
}

// ReleaseDocumentSession implements documentsapi.ServerInterface for the
// session release route, delegating to the private releaseSession handler.
func (h *Handler) ReleaseDocumentSession(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.releaseSession(w, r)
}

// ViewDocument implements documentsapi.ServerInterface for the document view
// route, delegating to the view sub-handler (404 when unwired).
func (h *Handler) ViewDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	if h.view == nil {
		writeNotFound(w, r)
		return
	}
	h.view.HandleView(w, r)
}

// writeNotFound reports a route whose optional sub-handler was not wired
// (WithSubHandlers received nil for this slot) as an RFC 9457 404, standing
// in for the removed `if h.xxx != nil { mux.Handle(...) }` conditional
// registration that HandlerWithOptions's unconditional route mounting no
// longer allows.
func writeNotFound(w http.ResponseWriter, r *http.Request) {
	problem.Respond(w, r, problem.New(http.StatusNotFound, problem.CodeNotFoundResource, "not found"))
}
