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

// GetTemplateVersionApprovalPreview implements
// templatesapi.ServerInterface.GetTemplateVersionApprovalPreview (SLICE 8a,
// unit 3.2): a read-only, submitter-scoped preview of the template's resolved
// active approval route (stages + each stage's ActorSelectors), so a
// chosen_actors picker can be built BEFORE the caller actually calls
// SubmitTemplateVersionForApproval. Gated by the same CapTemplateSubmit
// capability the submit endpoint itself requires — no new capability, no
// writes, no Idempotency-Key precondition. n identifies the specific version
// the caller is previewing for; GetVersion validates (id, n) resolves to a
// real version for this tenant (404 otherwise) — the route itself is governed
// by the template id, not the version, mirroring
// SubmitTemplateVersionForApproval's own resolve-then-delegate shape.
func (h *Handler) GetTemplateVersionApprovalPreview(w http.ResponseWriter, r *http.Request, id uuid.UUID, n int) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, codeTplInternalError, "internal server error"))
		return
	}
	if h.nilApprovalKernel() {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, codeTplInternalError, "approval kernel not configured"))
		return
	}
	templateID := id.String()

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateSubmit)); err != nil {
		writeMappedErr(w, r, err)
		return
	}

	if _, err := h.svc.GetVersion(r.Context(), tenantID, templateID, n); err != nil {
		writeMappedErr(w, r, err)
		return
	}

	preview, err := h.approvalSubmit.PreviewRoute(r.Context(), h.approvalRunner, tenantID, templateID)
	if err != nil {
		writeMappedErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, mapRoutePreviewToWire(preview))
}

// mapRoutePreviewToWire maps approvalapp.RoutePreview to the OpenAPI-shaped
// templatesapi.ApprovalRoutePreview. templatesapi.ActorSelector is a distinct
// generated type from approval/http/contracts.ActorSelector (each module
// codegens its own models from its own include-tags filter, ADR-governed
// module boundary) — this is the templates-side sibling of approval/http's
// mapSelectorsToWire, doing the identical field-for-field copy against the
// templates package's own generated wire type.
func mapRoutePreviewToWire(p approvalapp.RoutePreview) templatesapi.ApprovalRoutePreview {
	var routeID *uuid.UUID
	var routeName *string
	if p.RouteID != "" {
		if parsed, err := uuid.Parse(p.RouteID); err == nil {
			routeID = &parsed
		}
		name := p.RouteName
		routeName = &name
	}

	stages := make([]templatesapi.ApprovalRoutePreviewStage, 0, len(p.Stages))
	for _, s := range p.Stages {
		stages = append(stages, templatesapi.ApprovalRoutePreviewStage{
			StageOrder: s.StageOrder,
			Label:      s.Label,
			Selectors:  mapApprovalSelectorsToWire(s.Selectors),
		})
	}

	return templatesapi.ApprovalRoutePreview{
		RouteId:   routeID,
		RouteName: routeName,
		Stages:    stages,
	}
}

// mapApprovalSelectorsToWire maps approval/domain.ActorSelector to the
// templates-generated wire ActorSelector — the templates-side sibling of
// approval/http's mapSelectorsToWire (route_admin_handler.go), which maps the
// same domain type to approval/http/contracts.ActorSelector. Both do the same
// trivial field copy; each targets its own module's generated type.
func mapApprovalSelectorsToWire(selectors []approvaldomain.ActorSelector) []templatesapi.ActorSelector {
	if len(selectors) == 0 {
		return nil
	}
	out := make([]templatesapi.ActorSelector, 0, len(selectors))
	for _, s := range selectors {
		sel := templatesapi.ActorSelector{Kind: templatesapi.ActorSelectorKind(s.Kind)}
		if s.UserID != "" {
			userID := s.UserID
			sel.UserId = &userID
		}
		if s.Role != "" {
			role := s.Role
			sel.Role = &role
		}
		if s.AreaCode != "" {
			areaCode := s.AreaCode
			sel.AreaCode = &areaCode
		}
		out = append(out, sel)
	}
	return out
}
