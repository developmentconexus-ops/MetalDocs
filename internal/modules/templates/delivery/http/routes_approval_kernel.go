package http

import (
	"net/http"

	"github.com/google/uuid"

	approvalapp "metaldocs/internal/modules/approval/application"
	approvaldomain "metaldocs/internal/modules/approval/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
)

// SubmitTemplateVersionForApproval and SignoffTemplateVersion (this file) are
// the two thin kernel HTTP entry points (M3 P3.S2b-4, R2a): they resolve the
// caller-addressed (template_id, version_n) pair to the approval kernel's
// subject-generic identifiers and delegate the actual state transition to
// approval/application's published TemplateSubmitService / DecisionService.

// nilApprovalKernel reports whether the approval-kernel wiring
// (WithApprovalKernel) has not been configured on this Handler. The two
// kernel routes 500 rather than panic when unwired, so a Handler built
// without kernel wiring (e.g. an older test fixture) still serves every
// other route.
func (h *Handler) nilApprovalKernel() bool {
	return h.approvalSubmit == nil || h.approvalDecision == nil || h.approvalRead == nil || h.approvalRunner == nil
}

// SubmitTemplateVersionForApproval implements
// templatesapi.ServerInterface.SubmitTemplateVersionForApproval: resolves
// (id, n) -> template_version_id via the templates read side, then delegates
// to approvalSubmit.SubmitTemplateVersionForReview.
func (h *Handler) SubmitTemplateVersionForApproval(w http.ResponseWriter, r *http.Request, id uuid.UUID, n int, params templatesapi.SubmitTemplateVersionForApprovalParams) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	if h.nilApprovalKernel() {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "approval kernel not configured")
		return
	}
	actorID := userIDFromReq(r)
	templateID := id.String()

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateSubmit)); err != nil {
		writeMappedErr(w, err)
		return
	}

	version, err := h.svc.GetVersion(r.Context(), tenantID, templateID, n)
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	// requestBody is OPTIONAL (M4, unit 3.2, slice 5): existing callers that
	// never submit a submit_choice-governed route send no body at all. Only
	// decode when the client actually sent one — an empty/absent body is not
	// an error here (a missing chosen_actors entry for a submit_choice stage
	// is caught downstream by the application service's fail-closed check).
	var chosenActors []approvalapp.StageChosenActors
	if r.ContentLength != 0 {
		var body templatesapi.TemplateSubmitForApprovalRequest
		if err := readJSON(r, &body); err != nil {
			writeErr(w, http.StatusBadRequest, codeTplInvalidBody, err.Error())
			return
		}
		if body.ChosenActors != nil {
			for _, c := range *body.ChosenActors {
				chosenActors = append(chosenActors, approvalapp.StageChosenActors{
					StageOrder: c.StageOrder,
					UserIDs:    c.UserIds,
				})
			}
		}
	}

	res, err := h.approvalSubmit.SubmitTemplateVersionForReview(r.Context(), h.approvalRunner, approvalapp.TemplateSubmitRequest{
		TenantID:          tenantID,
		TemplateID:        templateID,
		TemplateVersionID: version.ID,
		SubmittedBy:       actorID,
		IdempotencyKey:    params.IdempotencyKey.String(),
		ChosenActors:      chosenActors,
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	instanceID, err := uuid.Parse(res.InstanceID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	var resp templatesapi.TemplateApprovalSubmitResponse
	resp.Data.InstanceId = instanceID
	resp.Data.VersionStatus = "under_review"
	writeJSON(w, http.StatusOK, resp)
}

// SignoffTemplateVersion implements
// templatesapi.ServerInterface.SignoffTemplateVersion: resolves the
// template's active approval instance + active stage for (id, n), then
// delegates to approvalDecision.RecordSignoff with a TEMPLATE subject.
// Unlike the document signoff route, the request carries no _content_hash —
// a template version's content is locked (not frozen) once under review, so
// DecisionService reads the content hash itself via TemplateVersionReader.
func (h *Handler) SignoffTemplateVersion(w http.ResponseWriter, r *http.Request, id uuid.UUID, n int, params templatesapi.SignoffTemplateVersionParams) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	if h.nilApprovalKernel() {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "approval kernel not configured")
		return
	}
	actorID := userIDFromReq(r)
	templateID := id.String()

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateApprove)); err != nil {
		writeMappedErr(w, err)
		return
	}

	var body templatesapi.SignoffTemplateVersionRequest
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, codeTplInvalidBody, err.Error())
		return
	}
	decision := approvaldomain.Decision(body.Decision)
	if decision != approvaldomain.DecisionApprove && decision != approvaldomain.DecisionReject {
		writeErr(w, http.StatusBadRequest, codeTplInvalidBody, "decision must be 'approve' or 'reject'")
		return
	}
	reason := ""
	if body.Reason != nil {
		reason = *body.Reason
	}

	version, err := h.svc.GetVersion(r.Context(), tenantID, templateID, n)
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	inst, err := h.approvalRead.LoadActiveInstanceBySubjectForMutation(r.Context(), h.approvalRunner, tenantID, string(approvaldomain.SubjectKindTemplate), version.ID)
	if err != nil {
		writeMappedErr(w, err)
		return
	}
	activeStage := inst.Active()
	if activeStage == nil {
		writeMappedErr(w, approvaldomain.ErrNoActiveStage)
		return
	}

	_ = params // Idempotency-Key is enforced by the idempotency.Require middleware (idempotentRoutes), not read here.

	res, err := h.approvalDecision.RecordSignoff(r.Context(), h.approvalRunner, approvalapp.SignoffRequest{
		TenantID:        tenantID,
		InstanceID:      inst.ID,
		StageInstanceID: activeStage.ID,
		ActorUserID:     actorID,
		Decision:        decision,
		Comment:         reason,
		SignatureMethod: "password_reauth",
		SignaturePayload: map[string]any{
			"password_token": body.Password,
		},
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	outcome := "recorded"
	switch {
	case res.InstanceApproved:
		outcome = "instance_approved"
	case res.InstanceRejected:
		outcome = "instance_rejected"
	case res.StageCompleted:
		outcome = "stage_completed"
	}
	resp := templatesapi.SignoffDocumentResponse{
		SignoffId: "",
		WasReplay: false,
		Outcome:   outcome,
	}
	writeJSON(w, http.StatusOK, resp)
}
