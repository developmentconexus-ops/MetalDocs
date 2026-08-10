package approvalhttp

import (
	"errors"
	"net/http"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/http/contracts"
)

// DocumentApprovalPreviewHandler handles GET
// /api/v1/documents/{id}/approval-preview (SLICE 8a, unit 3.2): a read-only,
// submitter-scoped preview of the document's resolved active approval route
// (stages + each stage's ActorSelectors), so a chosen_actors picker can be
// built BEFORE the caller actually submits. Gated by the same
// CapDocumentSubmit capability the submit endpoint itself requires — no new
// capability, no writes, no If-Match/Idempotency-Key precondition.
func (h *Handler) DocumentApprovalPreviewHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	documentID := r.PathValue("id")

	submitSvc := h.submitSvc
	if submitSvc == nil && h.services != nil {
		submitSvc = h.services.Submit
	}
	if submitSvc == nil {
		WriteError(w, r, errors.New("submit service not configured"))
		return
	}

	preview, err := submitSvc.PreviewRoute(r.Context(), h.runner, tenantID, documentID)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, r, http.StatusOK, mapRoutePreviewToWire(preview))
}

// mapRoutePreviewToWire maps application.RoutePreview to the OpenAPI-shaped
// contracts.ApprovalRoutePreviewResponse, reusing mapSelectorsToWire
// (route_admin_handler.go) — the same domain.ActorSelector -> wire mapper the
// route-admin list/read responses already use — so the selector wire shape is
// defined exactly once in this package.
func mapRoutePreviewToWire(p application.RoutePreview) contracts.ApprovalRoutePreviewResponse {
	var routeID, routeName *string
	if p.RouteID != "" {
		id := p.RouteID
		routeID = &id
		name := p.RouteName
		routeName = &name
	}

	stages := make([]contracts.ApprovalRoutePreviewStage, 0, len(p.Stages))
	for _, s := range p.Stages {
		stages = append(stages, contracts.ApprovalRoutePreviewStage{
			StageOrder: s.StageOrder,
			Label:      s.Label,
			Selectors:  mapSelectorsToWire(s.Selectors),
		})
	}

	return contracts.ApprovalRoutePreviewResponse{
		RouteID:   routeID,
		RouteName: routeName,
		Stages:    stages,
	}
}
