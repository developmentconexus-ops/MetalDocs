package http

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"

	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/application"
)

func (h *Handler) submitForReview(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	templateID := r.PathValue("id")
	versionNum, err := strconv.Atoi(r.PathValue("n"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, codeTplInvalidParam, "version must be an integer")
		return
	}

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateSubmit)); err != nil {
		writeMappedErr(w, err)
		return
	}

	v, err := h.svc.SubmitForReview(r.Context(), application.SubmitForReviewCmd{
		TenantID:      tenantID,
		ActorUserID:   actorID,
		TemplateID:    templateID,
		VersionNumber: versionNum,
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	dto, err := toAPIVersionDTO(v)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	var resp templatesapi.TemplateVersionEnvelope
	resp.Data.Version = dto
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) review(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	templateID := r.PathValue("id")
	versionNum, err := strconv.Atoi(r.PathValue("n"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, codeTplInvalidParam, "version must be an integer")
		return
	}

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateReview)); err != nil {
		writeMappedErr(w, err)
		return
	}

	var req struct {
		Accept bool   `json:"accept"`
		Reason string `json:"reason"`
	}
	if err := readJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, codeTplInvalidBody, err.Error())
		return
	}

	v, err := h.svc.Review(r.Context(), application.ReviewCmd{
		TenantID:      tenantID,
		ActorUserID:   actorID,
		ActorRoles:    actorRolesFromReq(r),
		TemplateID:    templateID,
		VersionNumber: versionNum,
		Accept:        req.Accept,
		Reason:        req.Reason,
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	dto, err := toAPIVersionDTO(v)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	var resp templatesapi.TemplateVersionEnvelope
	resp.Data.Version = dto
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) approve(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	templateID := r.PathValue("id")
	versionNum, err := strconv.Atoi(r.PathValue("n"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, codeTplInvalidParam, "version must be an integer")
		return
	}

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateApprove)); err != nil {
		writeMappedErr(w, err)
		return
	}

	var req struct {
		Accept bool   `json:"accept"`
		Reason string `json:"reason"`
	}
	if err := readJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, codeTplInvalidBody, err.Error())
		return
	}

	res, err := h.svc.Approve(r.Context(), application.ApproveCmd{
		TenantID:      tenantID,
		ActorUserID:   actorID,
		ActorRoles:    actorRolesFromReq(r),
		TemplateID:    templateID,
		VersionNumber: versionNum,
		Accept:        req.Accept,
		Reason:        req.Reason,
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	verDTO, err := toAPIVersionDTO(res.Version)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	var resp templatesapi.ApproveTemplateVersionResponse
	resp.Data.Version = verDTO
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) archiveTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	templateID := r.PathValue("id")

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateArchive)); err != nil {
		writeMappedErr(w, err)
		return
	}

	tpl, err := h.svc.ArchiveTemplate(r.Context(), application.ArchiveCmd{
		TenantID:    tenantID,
		ActorUserID: actorID,
		TemplateID:  templateID,
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	dto, err := toAPITemplateDTO(tpl)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	var resp templatesapi.ArchiveTemplateResponse
	resp.Data.Template = dto
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) upsertApprovalConfig(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	actorID := userIDFromReq(r)
	templateID := r.PathValue("id")

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateEdit)); err != nil {
		writeMappedErr(w, err)
		return
	}

	var req struct {
		ReviewerRole *string `json:"reviewer_role"`
		ApproverRole string  `json:"approver_role"`
	}
	if err := readJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, codeTplInvalidBody, err.Error())
		return
	}

	cfg, err := h.svc.UpsertApprovalConfig(r.Context(), application.UpsertApprovalConfigCmd{
		TenantID:     tenantID,
		ActorUserID:  actorID,
		TemplateID:   templateID,
		ReviewerRole: req.ReviewerRole,
		ApproverRole: req.ApproverRole,
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	cfgTemplateID, parseErr := uuid.Parse(cfg.TemplateID)
	if parseErr != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	var resp templatesapi.UpsertTemplateApprovalConfigResponse
	resp.Data.ApprovalConfig = templatesapi.TemplateApprovalConfig{
		TemplateId:   cfgTemplateID,
		ReviewerRole: cfg.ReviewerRole,
		ApproverRole: cfg.ApproverRole,
	}
	writeJSON(w, http.StatusOK, resp)
}

func actorRolesFromReq(r *http.Request) []string {
	if ctxRoles := iamdomain.RolesFromContext(r.Context()); len(ctxRoles) > 0 {
		out := make([]string, len(ctxRoles))
		for i, role := range ctxRoles {
			out[i] = string(role)
		}
		return out
	}
	// Fallback: no context roles available (legacy callers may still rely on header-based role wiring upstream).
	return nil
}
