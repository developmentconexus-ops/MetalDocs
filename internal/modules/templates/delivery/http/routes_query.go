package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/application"
)

const (
	systemBlankTemplateTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	systemBlankTemplateID       = "00000000-0000-0000-0000-000000000101"
)

func (h *Handler) listTemplates(w http.ResponseWriter, r *http.Request, params templatesapi.ListTemplatesParams) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateView)); err != nil {
		writeMappedErr(w, err)
		return
	}

	// Query params are now contract-declared (F-C1); the generated wrapper binds
	// and type-checks them, so the handler only enforces the business bound
	// (max page size) and applies defaults for omitted params.
	limit := 50
	if params.Limit != nil {
		limit = *params.Limit
	}
	if limit > 200 {
		writeErr(w, http.StatusBadRequest, codeTplInvalidLimit, "limit must be less than or equal to 200")
		return
	}
	offset := 0
	if params.Offset != nil {
		offset = *params.Offset
	}

	var docTypeCode *string
	if params.DocType != nil {
		if docType := strings.TrimSpace(*params.DocType); docType != "" {
			docTypeCode = &docType
		}
	}

	templates, err := h.svc.ListTemplates(r.Context(), application.ListFilter{
		TenantID:    tenantID,
		DocTypeCode: docTypeCode,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	createdByIDs := make([]string, 0, len(templates))
	for _, tpl := range templates {
		createdByIDs = append(createdByIDs, tpl.CreatedBy)
	}
	displayNames := h.resolveCreatedByDisplayNames(r.Context(), tenantID, createdByIDs)

	out := make([]templatesapi.TemplateDTO, 0, len(templates))
	for _, tpl := range templates {
		dto, err := toAPITemplateDTO(tpl, displayNames[tpl.CreatedBy])
		if err != nil {
			writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
			return
		}
		out = append(out, dto)
	}

	var resp templatesapi.ListTemplatesResponse
	resp.Data.Templates = out
	resp.Meta.Limit = limit
	resp.Meta.Offset = offset
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetSystemBlankTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateView)); err != nil {
		writeMappedErr(w, err)
		return
	}
	tpl, err := h.svc.GetTemplate(r.Context(), systemBlankTemplateTenantID, systemBlankTemplateID)
	if err != nil {
		writeMappedErr(w, err)
		return
	}
	ver, err := h.svc.GetVersion(r.Context(), systemBlankTemplateTenantID, tpl.ID, tpl.LatestVersion)
	if err != nil {
		writeMappedErr(w, err)
		return
	}
	tplUUID, err := uuid.Parse(tpl.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	verUUID, err := uuid.Parse(ver.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	var resp templatesapi.SystemBlankTemplateResponse
	resp.TemplateId = tplUUID
	resp.TemplateVersionId = verUUID
	resp.Name = tpl.Name
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) getTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	templateID := r.PathValue("id")
	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateView)); err != nil {
		writeMappedErr(w, err)
		return
	}

	tpl, err := h.svc.GetTemplate(r.Context(), tenantID, templateID)
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	latest, err := h.svc.GetVersion(r.Context(), tenantID, templateID, tpl.LatestVersion)
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	tplDTO, err := toAPITemplateDTO(tpl, h.resolveCreatedByDisplayName(r.Context(), tenantID, tpl.CreatedBy))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	latestDTO, err := toAPIVersionDTO(latest)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	var resp templatesapi.GetTemplateResponse
	resp.Data.Template = tplDTO
	resp.Data.LatestVersion = latestDTO
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) getVersion(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	templateID := r.PathValue("id")
	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateView)); err != nil {
		writeMappedErr(w, err)
		return
	}
	versionNum, err := strconv.Atoi(r.PathValue("n"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, codeTplInvalidParam, "version must be an integer")
		return
	}

	v, err := h.svc.GetVersion(r.Context(), tenantID, templateID, versionNum)
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	dto, err := toAPIVersionDTO(v)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, dto)
}

func (h *Handler) getDocxURL(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	templateID := r.PathValue("id")
	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateView)); err != nil {
		writeMappedErr(w, err)
		return
	}
	versionNum, err := strconv.Atoi(r.PathValue("n"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, codeTplInvalidParam, "version must be an integer")
		return
	}

	url, err := h.svc.GetDocxURL(r.Context(), application.GetDocxURLCmd{
		TenantID:      tenantID,
		TemplateID:    templateID,
		VersionNumber: versionNum,
	})
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	var resp templatesapi.GetTemplateDocxUrlResponse
	resp.Data.Url = url
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) listAudit(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	templateID := r.PathValue("id")
	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateView)); err != nil {
		writeMappedErr(w, err)
		return
	}
	q := r.URL.Query()

	limit, ok := readQueryInt(q.Get("limit"), 50)
	if !ok {
		writeErr(w, http.StatusBadRequest, codeTplInvalidLimit, "limit must be an integer")
		return
	}
	offset, ok := readQueryInt(q.Get("offset"), 0)
	if !ok {
		writeErr(w, http.StatusBadRequest, codeTplInvalidParam, "offset must be an integer")
		return
	}

	// TODO(pagination): migrate templates audit listing to keyset pagination.
	events, err := h.svc.ListAudit(r.Context(), tenantID, templateID, limit, offset)
	if err != nil {
		writeMappedErr(w, err)
		return
	}

	out := make([]templatesapi.TemplateAuditEvent, 0, len(events))
	for _, event := range events {
		tenantUUID, err := uuid.Parse(event.TenantID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
			return
		}
		tplUUID, err := uuid.Parse(event.TemplateID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
			return
		}
		actorUUID, err := uuid.Parse(event.ActorID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
			return
		}
		e := templatesapi.TemplateAuditEvent{
			TenantId:   tenantUUID,
			TemplateId: tplUUID,
			ActorId:    actorUUID,
			Action:     string(event.Action),
			OccurredAt: event.OccurredAt.UTC(),
		}
		if event.VersionID != nil {
			verUUID, err := uuid.Parse(*event.VersionID)
			if err != nil {
				writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
				return
			}
			e.VersionId = &verUUID
		}
		if event.Details != nil {
			d := map[string]interface{}(event.Details)
			e.Details = &d
		}
		out = append(out, e)
	}

	var resp templatesapi.ListTemplateAuditResponse
	resp.Data.Audit = out
	resp.Meta.Limit = limit
	resp.Meta.Offset = offset
	writeJSON(w, http.StatusOK, resp)
}

func readQueryInt(raw string, fallback int) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, true
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return v, true
}
