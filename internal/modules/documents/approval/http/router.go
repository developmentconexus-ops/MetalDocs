package approvalhttp

import (
	"net/http"

	approvalapi "metaldocs/internal/modules/documents/approval/api"
)

// RegisterRoutes wires all approval routes onto mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	wrapper := approvalapi.ServerInterfaceWrapper{
		Handler: h,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			WriteError(w, err)
		},
	}

	mux.HandleFunc("GET /api/v1/approval/inbox", wrapper.ListApprovalInbox)
	mux.HandleFunc("GET /api/v1/approval/instances/{instance_id}", wrapper.GetApprovalInstance)
	mux.HandleFunc("POST /api/v1/approval/instances/{instance_id}/cancel", wrapper.CancelApprovalInstance)
	mux.HandleFunc("POST /api/v1/approval/instances/{instance_id}/stages/{stage_id}/signoffs", wrapper.RecordApprovalStageSignoff)
	mux.HandleFunc("GET /api/v1/approval/routes", wrapper.ListApprovalRoutes)
	mux.HandleFunc("POST /api/v1/approval/routes", wrapper.CreateApprovalRoute)
	mux.HandleFunc("DELETE /api/v1/approval/routes/{id}", wrapper.DeactivateApprovalRoute)
	mux.HandleFunc("PUT /api/v1/approval/routes/{id}", wrapper.UpdateApprovalRoute)
	mux.HandleFunc("GET /api/v1/documents/{id}/approval-instance", wrapper.GetApprovalInstanceByDocument)
	mux.HandleFunc("POST /api/v1/documents/{id}/cancel", wrapper.CancelDocumentApproval)
	mux.HandleFunc("POST /api/v1/documents/{id}/obsolete", wrapper.ObsoleteDocument)
	mux.HandleFunc("POST /api/v1/documents/{id}/publish", wrapper.PublishDocument)
	mux.HandleFunc("POST /api/v1/documents/{id}/schedule-publish", wrapper.ScheduleDocumentPublish)
	mux.HandleFunc("POST /api/v1/documents/{id}/signoff", wrapper.RecordDocumentSignoff)
	mux.HandleFunc("POST /api/v1/documents/{id}/submit", wrapper.SubmitDocumentForApproval)
	mux.HandleFunc("POST /api/v1/documents/{id}/supersede", wrapper.SupersedeDocument)
}
