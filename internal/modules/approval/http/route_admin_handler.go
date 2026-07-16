package approvalhttp

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/http/contracts"
	"metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/platform/strictjson"
)

// CreateRouteHandler creates a new approval route for a document profile.
// Requires an Idempotency-Key header; replays return the original result.
func (h *Handler) CreateRouteHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		WriteError(w, ErrIdempotencyRequired)
		return
	}

	var req contracts.CreateRouteRequest
	if err := strictjson.Decode(r, &req); err != nil {
		WriteError(w, err)
		return
	}
	if err := req.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	routeAdminSvc := h.routeAdmin
	if routeAdminSvc == nil {
		WriteError(w, errors.New("route admin service not configured"))
		return
	}

	result, err := routeAdminSvc.Create(r.Context(), h.runner, application.CreateRouteInput{
		TenantID:       tenantID,
		ProfileCode:    req.ProfileCode,
		Name:           req.Name,
		ActorUserID:    actorID,
		SubjectKind:    req.SubjectKind,
		SubjectKey:     req.SubjectKey,
		IdempotencyKey: idempotencyKey,
		Stages:         mapStageRequests(req.Stages),
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, contracts.RouteResponse{
		RouteID: result.RouteID,
	})
}

// UpdateRouteHandler updates an existing approval route. Requires both an
// Idempotency-Key header and a valid If-Match header (OCC precondition
// against the route's version).
func (h *Handler) UpdateRouteHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)
	routeID := r.PathValue("id")
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		WriteError(w, ErrIdempotencyRequired)
		return
	}
	expectedVersion, err := parseIfMatch(r.Header.Get("If-Match"))
	if err != nil {
		WriteError(w, err)
		return
	}

	var req contracts.UpdateRouteRequest
	if err := strictjson.Decode(r, &req); err != nil {
		WriteError(w, err)
		return
	}
	if err := req.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	routeAdminSvc := h.routeAdmin
	if routeAdminSvc == nil {
		WriteError(w, errors.New("route admin service not configured"))
		return
	}

	result, err := routeAdminSvc.Update(r.Context(), h.runner, application.UpdateRouteInput{
		TenantID:        tenantID,
		RouteID:         routeID,
		Name:            req.Name,
		ActorUserID:     actorID,
		IdempotencyKey:  idempotencyKey,
		ExpectedVersion: expectedVersion,
		Stages:          mapStageRequests(req.Stages),
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contracts.RouteResponse{
		RouteID:    result.RouteID,
		NewVersion: intPtr(result.NewVersion),
	})
}

// DeactivateRouteHandler deactivates an approval route. Requires both an
// Idempotency-Key header and a valid If-Match header (OCC precondition
// against the route's version).
func (h *Handler) DeactivateRouteHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)
	routeID := r.PathValue("id")
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		WriteError(w, ErrIdempotencyRequired)
		return
	}
	expectedVersion, err := parseIfMatch(r.Header.Get("If-Match"))
	if err != nil {
		WriteError(w, err)
		return
	}

	var body contracts.DeactivateRouteRequest
	if err := strictjson.Decode(r, &body); err != nil {
		WriteError(w, err)
		return
	}
	if err := body.Validate(); err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}
	reason := strings.TrimSpace(body.Reason)

	routeAdminSvc := h.routeAdmin
	if routeAdminSvc == nil {
		WriteError(w, errors.New("route admin service not configured"))
		return
	}

	result, err := routeAdminSvc.Deactivate(r.Context(), h.runner, application.DeactivateRouteInput{
		TenantID:        tenantID,
		RouteID:         routeID,
		ActorUserID:     actorID,
		IdempotencyKey:  idempotencyKey,
		Reason:          reason,
		ExpectedVersion: expectedVersion,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, contracts.RouteResponse{
		RouteID: result.RouteID,
	})
}

// ListRoutesHandler returns all approval routes configured for the tenant.
func (h *Handler) ListRoutesHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := actorIDFromRequest(r)

	routeAdminSvc := h.routeAdmin
	if routeAdminSvc == nil {
		WriteError(w, errors.New("route admin service not configured"))
		return
	}

	out, err := routeAdminSvc.List(r.Context(), h.runner, tenantID, actorID)
	if err != nil {
		WriteError(w, err)
		return
	}

	routes := make([]contracts.ListRouteItem, 0, len(out.Routes))
	for _, route := range out.Routes {
		routes = append(routes, mapListRoute(route))
	}

	total := len(out.Routes)
	if len(out.Routes) > 0 {
		total = out.Routes[0].Total
	}
	WriteJSON(w, http.StatusOK, contracts.ListRoutesResponse{
		Routes: routes,
		Total:  total,
	})
}

func mapListRoute(route infrastructure.Route) contracts.ListRouteItem {
	createdAt := route.CreatedAt.UTC().Format(time.RFC3339)
	updatedAt := route.UpdatedAt.UTC().Format(time.RFC3339)
	if route.UpdatedAt.IsZero() {
		updatedAt = createdAt
	}

	stages := make([]contracts.ListStageItem, 0, len(route.Stages))
	for _, stage := range route.Stages {
		// Selectors is the sole source of truth for a stage's actor pool
		// (unit 3.2 slice 7a, wire contract extermination): the flat
		// required_role/area_code wire fields are gone entirely, no
		// derived-back compat pair.
		stages = append(stages, contracts.ListStageItem{
			Order:              stage.Order,
			Name:               stage.Name,
			RequiredCapability: stage.RequiredCapability,
			Quorum:             contracts.QuorumKind(stage.Quorum),
			QuorumM:            stage.QuorumM,
			DriftPolicy:        contracts.DriftPolicyKind(stage.DriftPolicy),
			StageKind:          mapStageKind(stage.Kind),
			Selectors:          mapSelectorsToWire(stage.Selectors),
		})
	}

	return contracts.ListRouteItem{
		ID:          route.ID,
		Name:        route.Name,
		TenantID:    route.TenantID,
		ProfileCode: listRouteProfileCode(route),
		SubjectKind: route.SubjectKind,
		SubjectKey:  route.SubjectKey,
		Active:      route.Active,
		Version:     route.Version,
		Stages:      stages,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// listRouteProfileCode carries the repository scan's NULL-vs-empty truth
// (postgres_approval_repository.go scanRouteListRows: profile_code sql.NullString
// collapsed to route.ProfileCode string) back out honestly on the wire. A
// template route has no profile by DB constraint
// (approval_routes_template_subject_projection_check, ADR 0082); infrastructure.Route
// carries SubjectKind alongside ProfileCode on the same scanned row, so
// SubjectKind=="template" is an exact, row-local proxy for "profile_code was
// SQL NULL" — no separate Valid flag needs to be threaded through the
// repository projection to make this honest.
func listRouteProfileCode(route infrastructure.Route) *string {
	if route.SubjectKind == string(domain.SubjectKindTemplate) {
		return nil
	}
	code := route.ProfileCode
	return &code
}

func mapStageRequests(stages []contracts.StageRequest) []domain.Stage {
	out := make([]domain.Stage, 0, len(stages))
	for _, s := range stages {
		// required_capability is validated as non-empty + pattern-checked by
		// CreateRouteRequest/UpdateRouteRequest.Validate before we reach here,
		// so no silent default is needed (and a defaulted code may not even be
		// a registered capability).
		cap := strings.TrimSpace(s.RequiredCapability)
		// Selectors is the sole source of truth for domain.Stage (unit 3.2
		// slice 7a, wire contract extermination): the wire's selectors field
		// is contract-required and non-empty (validateStages), so no
		// flat->selector synthesis is needed or performed here anymore.
		out = append(out, domain.Stage{
			Order:              s.Order,
			Name:               s.Name,
			RequiredCapability: cap,
			Quorum:             domain.QuorumPolicy(s.Quorum),
			QuorumM:            s.QuorumM,
			OnEligibilityDrift: domain.DriftPolicy(s.DriftPolicy),
			// F0: empty stage_kind stays the domain zero value ("") and is
			// defaulted to approval at the persistence layer (insertRouteStages,
			// migration 0286 DEFAULT 'approval'); a supplied review makes a
			// review-kind stage. Validated as review|approval|"" in contracts.
			Kind:      domain.StageKind(s.StageKind),
			Selectors: mapSelectorsFromWire(s.Selectors),
		})
	}
	return out
}

// mapSelectorsFromWire maps wire ActorSelectors to domain ActorSelectors
// (M4, unit 3.2, slice 4). Role and AreaCode are normalized the same way the
// synthesized flat->selector fallback is above (lowercase + trim); UserID is
// trimmed only — user ids are not role/area codes and are not lowercased
// elsewhere in this module.
func mapSelectorsFromWire(selectors []contracts.ActorSelector) []domain.ActorSelector {
	if len(selectors) == 0 {
		return nil
	}
	out := make([]domain.ActorSelector, 0, len(selectors))
	for _, s := range selectors {
		out = append(out, domain.ActorSelector{
			Kind:     domain.SelectorKind(s.Kind),
			UserID:   strings.TrimSpace(s.UserID),
			Role:     strings.ToLower(strings.TrimSpace(s.Role)),
			AreaCode: strings.ToLower(strings.TrimSpace(s.AreaCode)),
		})
	}
	return out
}

// mapSelectorsToWire maps domain ActorSelectors to their wire representation
// (M4, unit 3.2, slice 4) — the read-side inverse of mapSelectorsFromWire.
func mapSelectorsToWire(selectors []domain.ActorSelector) []contracts.ActorSelector {
	if len(selectors) == 0 {
		return nil
	}
	out := make([]contracts.ActorSelector, 0, len(selectors))
	for _, s := range selectors {
		out = append(out, contracts.ActorSelector{
			Kind:     contracts.SelectorKind(s.Kind),
			UserID:   s.UserID,
			Role:     s.Role,
			AreaCode: s.AreaCode,
		})
	}
	return out
}

// mapStageKind normalizes a persisted stage-kind string for the wire. The DB
// column is NOT NULL DEFAULT 'approval', but a legacy row read before the
// backfill (or a zero value) is defaulted here so the response is never an
// empty enum value (no-fallback: a known-safe canonical default, not a masked
// unknown — the two kinds are exhaustive).
func mapStageKind(kind string) contracts.StageKind {
	if kind == string(contracts.StageKindReview) {
		return contracts.StageKindReview
	}
	return contracts.StageKindApproval
}

func intPtr(v int) *int {
	return &v
}
