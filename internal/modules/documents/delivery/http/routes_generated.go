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

func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request, params documentsapi.ListDocumentsParams) {
	h.listDocuments(w, r, params)
}

func (h *Handler) DocumentStats(w http.ResponseWriter, r *http.Request, params documentsapi.DocumentStatsParams) {
	h.documentStats(w, r, params)
}

func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.getDocument(w, r)
}

func (h *Handler) RenameDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.renameDocument(w, r)
}

func (h *Handler) ArchiveDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.archiveDocument(w, r)
}

func (h *Handler) CommitDocumentAutosave(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.commitAutosave(w, r)
}

func (h *Handler) PresignDocumentAutosave(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.presignAutosave(w, r)
}

func (h *Handler) ListDocumentCheckpoints(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.listCheckpoints(w, r)
}

func (h *Handler) CreateDocumentCheckpoint(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.createCheckpoint(w, r)
}

func (h *Handler) RestoreDocumentCheckpoint(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, version int) {
	r.SetPathValue("id", id.String())
	r.SetPathValue("version", strconv.Itoa(version))
	h.restoreCheckpoint(w, r)
}

func (h *Handler) ListDocumentComments(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.listComments(w, r)
}

func (h *Handler) CreateDocumentComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.createComment(w, r)
}

func (h *Handler) DeleteDocumentComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, libraryId int) {
	r.SetPathValue("id", id.String())
	r.SetPathValue("library_id", strconv.Itoa(libraryId))
	h.deleteComment(w, r)
}

func (h *Handler) UpdateDocumentComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, libraryId int) {
	r.SetPathValue("id", id.String())
	r.SetPathValue("library_id", strconv.Itoa(libraryId))
	h.updateComment(w, r)
}

func (h *Handler) DuplicateDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.duplicateDocument(w, r)
}

func (h *Handler) GetDocumentDocxURL(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	if h.export == nil {
		writeNotFound(w, r)
		return
	}
	h.export.exportDocxURL(w, r)
}

func (h *Handler) ExportDocumentPDF(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	if h.export == nil {
		writeNotFound(w, r)
		return
	}
	h.export.exportPDF(w, r)
}

func (h *Handler) GetDocumentFillInSchema(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	if h.fillIn == nil {
		writeNotFound(w, r)
		return
	}
	h.fillIn.GetFillInSchema(w, r)
}

func (h *Handler) GetDocumentPlaceholderOptions(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, pid string) {
	r.SetPathValue("id", id.String())
	r.SetPathValue("pid", pid)
	if h.placeholderOpts == nil {
		writeNotFound(w, r)
		return
	}
	h.placeholderOpts.HandleGetOptions(w, r)
}

func (h *Handler) ListDocumentPlaceholderValues(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	if h.fillIn == nil {
		writeNotFound(w, r)
		return
	}
	h.fillIn.ListPlaceholderValues(w, r)
}

func (h *Handler) PutDocumentPlaceholderValue(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, pid string) {
	r.SetPathValue("id", id.String())
	r.SetPathValue("pid", pid)
	if h.fillIn == nil {
		writeNotFound(w, r)
		return
	}
	h.fillIn.PutPlaceholderValue(w, r)
}

func (h *Handler) ReconstructDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	if h.reconstruct == nil {
		writeNotFound(w, r)
		return
	}
	h.reconstruct.HandleReconstruct(w, r)
}

func (h *Handler) GetDocumentRevisionHistory(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.listRevisionHistory(w, r)
}

func (h *Handler) GetDocumentRevisionUrl(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, rid openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	r.SetPathValue("rid", rid.String())
	h.signedRevisionURL(w, r)
}

func (h *Handler) AcquireDocumentSession(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.acquireSession(w, r)
}

func (h *Handler) ForceReleaseDocumentSession(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.forceReleaseSession(w, r)
}

func (h *Handler) HeartbeatDocumentSession(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.heartbeatSession(w, r)
}

func (h *Handler) ReleaseDocumentSession(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	h.releaseSession(w, r)
}

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
	_ = problem.Write(w, problem.New(http.StatusNotFound, problem.CodeNotFound, "not found"))
}
