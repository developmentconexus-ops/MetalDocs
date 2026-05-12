package approvalhttp

import (
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *Handler) SubmitDocumentForApprovalV2(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.SubmitHandler(w, r)
}

func (h *Handler) RecordApprovalStageSignoffV2(w http.ResponseWriter, r *http.Request, instanceId openapi_types.UUID, stageId openapi_types.UUID) {
	h.SignoffHandler(w, r)
}

func (h *Handler) PublishDocumentV2(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.PublishHandler(w, r)
}

func (h *Handler) ScheduleDocumentPublishV2(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.SchedulePublishHandler(w, r)
}

func (h *Handler) SupersedeDocumentV2(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.SupersedeHandler(w, r)
}

func (h *Handler) ObsoleteDocumentV2(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.ObsoleteHandler(w, r)
}

func (h *Handler) CancelApprovalInstanceV2(w http.ResponseWriter, r *http.Request, instanceId openapi_types.UUID) {
	h.CancelHandler(w, r)
}

func (h *Handler) GetApprovalInstanceV2(w http.ResponseWriter, r *http.Request, instanceId openapi_types.UUID) {
	h.GetInstanceHandler(w, r)
}

func (h *Handler) GetApprovalInstanceByDocumentV2(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.GetInstanceByDocumentHandler(w, r)
}

func (h *Handler) ListApprovalInboxV2(w http.ResponseWriter, r *http.Request) {
	h.InboxHandler(w, r)
}

func (h *Handler) RecordDocumentSignoffV2(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.SignoffByDocumentHandler(w, r)
}

func (h *Handler) CancelDocumentApprovalV2(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.CancelByDocumentHandler(w, r)
}

func (h *Handler) CreateApprovalRouteV2(w http.ResponseWriter, r *http.Request) {
	h.CreateRouteHandler(w, r)
}

func (h *Handler) UpdateApprovalRouteV2(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.UpdateRouteHandler(w, r)
}

func (h *Handler) DeactivateApprovalRouteV2(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.DeactivateRouteHandler(w, r)
}

func (h *Handler) ListApprovalRoutesV2(w http.ResponseWriter, r *http.Request) {
	h.ListRoutesHandler(w, r)
}
