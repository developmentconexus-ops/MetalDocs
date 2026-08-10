package http

import (
	"net/http"

	"github.com/google/uuid"

	approvalapp "metaldocs/internal/modules/approval/application"
	approvaldomain "metaldocs/internal/modules/approval/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/platform/problem"
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
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, codeTplInternalError, "internal server error"))
		return
	}
	if h.nilApprovalKernel() {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, codeTplInternalError, "approval kernel not configured"))
		return
	}
	actorID := userIDFromReq(r)
	templateID := id.String()

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateSubmit)); err != nil {
		writeMappedErr(w, r, err)
		return
	}

	version, err := h.svc.GetVersion(r.Context(), tenantID, templateID, n)
	if err != nil {
		writeMappedErr(w, r, err)
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
			problem.Respond(w, r, problem.New(http.StatusBadRequest, codeTplInvalidBody, err.Error()))
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
		writeMappedErr(w, r, err)
		return
	}

	instanceID, err := uuid.Parse(res.InstanceID)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, codeTplInternalError, "internal server error"))
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
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, codeTplInternalError, "internal server error"))
		return
	}
	if h.nilApprovalKernel() {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, codeTplInternalError, "approval kernel not configured"))
		return
	}
	actorID := userIDFromReq(r)
	templateID := id.String()

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateApprove)); err != nil {
		writeMappedErr(w, r, err)
		return
	}

	var body templatesapi.SignoffTemplateVersionRequest
	if err := readJSON(r, &body); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, codeTplInvalidBody, err.Error()))
		return
	}
	decision := approvaldomain.Decision(body.Decision)
	if decision != approvaldomain.DecisionApprove && decision != approvaldomain.DecisionReject {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, codeTplInvalidBody, "decision must be 'approve' or 'reject'"))
		return
	}
	reason := ""
	if body.Reason != nil {
		reason = *body.Reason
	}

	version, err := h.svc.GetVersion(r.Context(), tenantID, templateID, n)
	if err != nil {
		writeMappedErr(w, r, err)
		return
	}

	inst, err := h.approvalRead.LoadActiveInstanceBySubjectForMutation(r.Context(), h.approvalRunner, tenantID, string(approvaldomain.SubjectKindTemplate), version.ID)
	if err != nil {
		writeMappedErr(w, r, err)
		return
	}
	activeStage := inst.Active()
	if activeStage == nil {
		writeMappedErr(w, r, approvaldomain.ErrNoActiveStage)
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
		writeMappedErr(w, r, err)
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
	// F-E4-3: thread the persisted approval_signoffs row id out to the wire,
	// exactly as the document signoff delivery does
	// (approval/http/doc_approval_handler.go). SignoffResult.SignoffID is
	// already populated by DecisionService.RecordSignoff — including the
	// ORIGINAL row id on the DB-level replay branch (F-QA4-7) — so the template
	// entry point was simply dropping it. was_replay stays false here: unlike
	// the document handler, this route has no application-level replay store;
	// its idempotency is enforced by the idempotency.Require middleware.
	resp := templatesapi.SignoffDocumentResponse{
		SignoffId: res.SignoffID,
		WasReplay: false,
		Outcome:   outcome,
	}
	writeJSON(w, http.StatusOK, resp)
}
