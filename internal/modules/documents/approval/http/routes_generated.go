package approvalhttp

import (
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"
	approvalapi "metaldocs/internal/modules/documents/approval/api"
)

func (h *Handler) SubmitDocumentForApproval(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.SubmitDocumentForApprovalParams) {
	h.SubmitHandler(w, r)
}

func (h *Handler) RecordApprovalStageSignoff(w http.ResponseWriter, r *http.Request, instanceId openapi_types.UUID, stageId openapi_types.UUID, params approvalapi.RecordApprovalStageSignoffParams) {
	h.SignoffHandler(w, r)
}

func (h *Handler) PublishDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.PublishHandler(w, r)
}

func (h *Handler) ScheduleDocumentPublish(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.SchedulePublishHandler(w, r)
}

func (h *Handler) SupersedeDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.SupersedeHandler(w, r)
}

func (h *Handler) ObsoleteDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.ObsoleteHandler(w, r)
}

func (h *Handler) CancelApprovalInstance(w http.ResponseWriter, r *http.Request, instanceId openapi_types.UUID) {
	h.CancelHandler(w, r)
}

func (h *Handler) GetApprovalInstance(w http.ResponseWriter, r *http.Request, instanceId openapi_types.UUID) {
	h.GetInstanceHandler(w, r)
}

func (h *Handler) GetApprovalInstanceByDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.GetInstanceByDocumentHandler(w, r)
}

func (h *Handler) ListApprovalInbox(w http.ResponseWriter, r *http.Request) {
	h.InboxHandler(w, r)
}

func (h *Handler) RecordDocumentSignoff(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.RecordDocumentSignoffParams) {
	h.SignoffByDocumentHandler(w, r)
}

func (h *Handler) CancelDocumentApproval(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.CancelByDocumentHandler(w, r)
}

func (h *Handler) CreateApprovalRoute(w http.ResponseWriter, r *http.Request) {
	h.CreateRouteHandler(w, r)
}

func (h *Handler) UpdateApprovalRoute(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.UpdateRouteHandler(w, r)
}

func (h *Handler) DeactivateApprovalRoute(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.DeactivateRouteHandler(w, r)
}

func (h *Handler) ListApprovalRoutes(w http.ResponseWriter, r *http.Request) {
	h.ListRoutesHandler(w, r)
}
