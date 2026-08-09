package approvalhttp

import (
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"
	approvalapi "metaldocs/internal/modules/approval/api"
)

// SubmitDocumentForApproval is the spec-generated adapter for the
// submit-for-approval operation.
func (h *Handler) SubmitDocumentForApproval(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.SubmitDocumentForApprovalParams) {
	h.SubmitHandler(w, r)
}

// RecordApprovalStageSignoff is the spec-generated adapter for the
// stage-scoped signoff operation.
func (h *Handler) RecordApprovalStageSignoff(w http.ResponseWriter, r *http.Request, instanceID openapi_types.UUID, stageID openapi_types.UUID, params approvalapi.RecordApprovalStageSignoffParams) {
	h.SignoffHandler(w, r)
}

// MarkDocumentReviewed is the spec-generated adapter for the F6.2 mark-reviewed
// operation: records a periodic-review completion on a live published
// revision (last_reviewed_at + next review_due_at) under CapDocumentReview.
// Not a status transition — the document stays published (T5).
func (h *Handler) MarkDocumentReviewed(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.MarkDocumentReviewedParams) {
	h.MarkReviewedHandler(w, r)
}

// ObsoleteDocument is the spec-generated adapter for the obsolete-document
// operation.
func (h *Handler) ObsoleteDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.ObsoleteDocumentParams) {
	h.ObsoleteHandler(w, r)
}

// CancelApprovalInstance is the spec-generated adapter for the
// cancel-approval-instance operation.
func (h *Handler) CancelApprovalInstance(w http.ResponseWriter, r *http.Request, instanceID openapi_types.UUID, params approvalapi.CancelApprovalInstanceParams) {
	h.CancelHandler(w, r)
}

// ExtendApprovalSLA is the spec-generated adapter for the Task 9 (approval
// accountability loop) SLA-extension operation: postpones due_at on the
// instance's active stage.
func (h *Handler) ExtendApprovalSLA(w http.ResponseWriter, r *http.Request, instanceID openapi_types.UUID, params approvalapi.ExtendApprovalSLAParams) {
	h.ExtendSLAHandler(w, r)
}

// RecordApprovalReviewVerdict is the spec-generated adapter for the F4
// review-stage runtime-verdict operation (M2b): records a ready/request_changes
// verdict against a review-kind stage.
func (h *Handler) RecordApprovalReviewVerdict(w http.ResponseWriter, r *http.Request, instanceID openapi_types.UUID, stageID openapi_types.UUID, params approvalapi.RecordApprovalReviewVerdictParams) {
	h.ReviewVerdictHandler(w, r)
}

// GetApprovalInstance is the spec-generated adapter for the
// get-approval-instance-by-id operation.
func (h *Handler) GetApprovalInstance(w http.ResponseWriter, r *http.Request, instanceID openapi_types.UUID) {
	h.GetInstanceHandler(w, r)
}

// GetApprovalInstanceByDocument is the spec-generated adapter for the
// get-approval-instance-by-document operation.
func (h *Handler) GetApprovalInstanceByDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.GetInstanceByDocumentHandler(w, r)
}

// ListApprovalInbox is the spec-generated adapter for the
// list-approval-inbox operation.
func (h *Handler) ListApprovalInbox(w http.ResponseWriter, r *http.Request, params approvalapi.ListApprovalInboxParams) {
	h.InboxHandler(w, r)
}

// RecordDocumentSignoff is the spec-generated adapter for the
// document-scoped signoff operation.
func (h *Handler) RecordDocumentSignoff(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.RecordDocumentSignoffParams) {
	h.SignoffByDocumentHandler(w, r)
}

// CancelDocumentApproval is the spec-generated adapter for the
// document-scoped cancel-approval operation.
func (h *Handler) CancelDocumentApproval(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.CancelDocumentApprovalParams) {
	h.CancelByDocumentHandler(w, r)
}

// CreateApprovalRoute is the spec-generated adapter for the
// create-approval-route operation.
func (h *Handler) CreateApprovalRoute(w http.ResponseWriter, r *http.Request, params approvalapi.CreateApprovalRouteParams) {
	h.CreateRouteHandler(w, r)
}

// UpdateApprovalRoute is the spec-generated adapter for the
// update-approval-route operation.
func (h *Handler) UpdateApprovalRoute(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.UpdateApprovalRouteParams) {
	h.UpdateRouteHandler(w, r)
}

// DeactivateApprovalRoute is the spec-generated adapter for the
// deactivate-approval-route operation.
func (h *Handler) DeactivateApprovalRoute(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.DeactivateApprovalRouteParams) {
	h.DeactivateRouteHandler(w, r)
}

// ListApprovalRoutes is the spec-generated adapter for the
// list-approval-routes operation.
func (h *Handler) ListApprovalRoutes(w http.ResponseWriter, r *http.Request) {
	h.ListRoutesHandler(w, r)
}

// RecordApprovalFastForward is the spec-generated adapter for the R5
// "Aprovar já" contract path (unit 2.3 G3): records a `ready` review verdict
// and an approve signoff as two ledger writes in one transaction.
func (h *Handler) RecordApprovalFastForward(w http.ResponseWriter, r *http.Request, instanceID openapi_types.UUID, stageID openapi_types.UUID, params approvalapi.RecordApprovalFastForwardParams) {
	h.FastForwardHandler(w, r)
}

// GetDocumentApprovalPreview is the spec-generated adapter for the SLICE 8a
// (unit 3.2) read-only route-preview operation.
func (h *Handler) GetDocumentApprovalPreview(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.DocumentApprovalPreviewHandler(w, r)
}
