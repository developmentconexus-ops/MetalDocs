package approvalhttp

import (
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"
	approvalapi "metaldocs/internal/modules/documents/approval/api"
	"metaldocs/internal/platform/problem"
)

func (h *Handler) SubmitDocumentForApproval(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.SubmitDocumentForApprovalParams) {
	h.SubmitHandler(w, r)
}

func (h *Handler) RecordApprovalStageSignoff(w http.ResponseWriter, r *http.Request, instanceId openapi_types.UUID, stageId openapi_types.UUID, params approvalapi.RecordApprovalStageSignoffParams) {
	h.SignoffHandler(w, r)
}

func (h *Handler) PublishDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.PublishDocumentParams) {
	h.PublishHandler(w, r)
}

func (h *Handler) ScheduleDocumentPublish(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.ScheduleDocumentPublishParams) {
	h.SchedulePublishHandler(w, r)
}

func (h *Handler) SupersedeDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.SupersedeDocumentParams) {
	h.SupersedeHandler(w, r)
}

// MarkDocumentReviewed is the spec-generated adapter for the F6.2 mark-reviewed
// operation: records a periodic-review completion on a live published
// revision (last_reviewed_at + next review_due_at) under CapDocumentReview.
// Not a status transition — the document stays published (T5).
func (h *Handler) MarkDocumentReviewed(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.MarkDocumentReviewedParams) {
	h.MarkReviewedHandler(w, r)
}

func (h *Handler) ObsoleteDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.ObsoleteDocumentParams) {
	h.ObsoleteHandler(w, r)
}

func (h *Handler) CancelApprovalInstance(w http.ResponseWriter, r *http.Request, instanceId openapi_types.UUID, params approvalapi.CancelApprovalInstanceParams) {
	h.CancelHandler(w, r)
}

// RecordApprovalReviewVerdict is the spec-generated adapter for the F4
// review-stage runtime-verdict operation (M2b): records a ready/request_changes
// verdict against a review-kind stage.
func (h *Handler) RecordApprovalReviewVerdict(w http.ResponseWriter, r *http.Request, instanceId openapi_types.UUID, stageId openapi_types.UUID, params approvalapi.RecordApprovalReviewVerdictParams) {
	h.ReviewVerdictHandler(w, r)
}

func (h *Handler) GetApprovalInstance(w http.ResponseWriter, r *http.Request, instanceId openapi_types.UUID) {
	h.GetInstanceHandler(w, r)
}

func (h *Handler) GetApprovalInstanceByDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.GetInstanceByDocumentHandler(w, r)
}

func (h *Handler) ListApprovalInbox(w http.ResponseWriter, r *http.Request, params approvalapi.ListApprovalInboxParams) {
	h.InboxHandler(w, r)
}

func (h *Handler) RecordDocumentSignoff(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.RecordDocumentSignoffParams) {
	h.SignoffByDocumentHandler(w, r)
}

func (h *Handler) CancelDocumentApproval(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.CancelDocumentApprovalParams) {
	h.CancelByDocumentHandler(w, r)
}

func (h *Handler) CreateApprovalRoute(w http.ResponseWriter, r *http.Request, params approvalapi.CreateApprovalRouteParams) {
	h.CreateRouteHandler(w, r)
}

func (h *Handler) UpdateApprovalRoute(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.UpdateApprovalRouteParams) {
	h.UpdateRouteHandler(w, r)
}

func (h *Handler) DeactivateApprovalRoute(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params approvalapi.DeactivateApprovalRouteParams) {
	h.DeactivateRouteHandler(w, r)
}

func (h *Handler) ListApprovalRoutes(w http.ResponseWriter, r *http.Request) {
	h.ListRoutesHandler(w, r)
}

// RecordApprovalFastForward is a build-satisfying placeholder for the R5
// "Aprovar já" contract path (unit 2.3 G3, slice S3). CON-03 mounts the
// generated approvalapi.ServerInterface directly (router.go), so a route
// added to the spec that lacks a Handler method fails to compile — this stub
// exists only to keep that guarantee while the real handler (If-Match /
// Idempotency-Key enforcement, application.Services.FastForward wiring,
// response mapping) ships in slice S4. It intentionally fails closed with 501
// rather than silently succeeding or 404ing.
func (h *Handler) RecordApprovalFastForward(w http.ResponseWriter, r *http.Request, instanceId openapi_types.UUID, stageId openapi_types.UUID, params approvalapi.RecordApprovalFastForwardParams) {
	problem.Write(w, problem.New(http.StatusNotImplemented, "approval.fast_forward_not_implemented", "fast-forward endpoint not yet wired"))
}
