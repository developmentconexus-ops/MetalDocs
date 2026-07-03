// Package http is the taxonomy module's HTTP delivery layer: it mounts the
// 16 oapi-codegen-generated /api/v1/taxonomy/* routes onto Handler, which
// dispatches to the profile/area/family application services and maps
// domain errors to the RFC 9457 problem+json envelope.
package http

import (
	"context"
	"net/http"

	taxonomyapi "metaldocs/internal/modules/taxonomy/api"
	"metaldocs/internal/modules/taxonomy/domain"
)

type profileService interface {
	List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.DocumentProfile, error)
	Get(ctx context.Context, tenantID string, code domain.ProfileCode) (*domain.DocumentProfile, error)
	Create(ctx context.Context, p *domain.DocumentProfile) error
	Update(ctx context.Context, p *domain.DocumentProfile) error
	SetDefaultTemplate(ctx context.Context, tenantID string, profileCode domain.ProfileCode, templateVersionID, actorID string) error
	Archive(ctx context.Context, tenantID string, profileCode domain.ProfileCode, actorID string) error
}

type areaService interface {
	List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.ProcessArea, error)
	Get(ctx context.Context, tenantID string, code domain.AreaCode) (*domain.ProcessArea, error)
	Create(ctx context.Context, a *domain.ProcessArea) error
	Update(ctx context.Context, a *domain.ProcessArea) error
	Archive(ctx context.Context, tenantID string, areaCode domain.AreaCode, actorID string) error
}

type familyService interface {
	List(ctx context.Context, tenantID string, includeInactive bool) ([]domain.DocumentFamily, error)
	Get(ctx context.Context, tenantID string, code domain.FamilyCode) (*domain.DocumentFamily, error)
	Create(ctx context.Context, f *domain.DocumentFamily) error
	Update(ctx context.Context, f *domain.DocumentFamily) (*domain.DocumentFamily, error)
	Deactivate(ctx context.Context, code domain.FamilyCode) error
}

// Handler wires the generated taxonomyapi.ServerInterface to the taxonomy
// application services. It holds no persistence or authz logic of its own.
type Handler struct {
	profiles profileService
	areas    areaService
	families familyService
}

// RegisterRoutes mounts the 16 generated taxonomy operations onto mux under
// the "/api/v1" base, translating unmarshalable request bodies to a 400
// VALIDATION_ERROR problem+json response.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	taxonomyapi.HandlerWithOptions(h, taxonomyapi.StdHTTPServerOptions{
		BaseRouter: mux,
		// AD-1: spec path keys are relative; the generated router prepends this
		// base so served routes stay /api/v1/* and the codegen matches the spec.
		BaseURL: "/api/v1",
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		},
	})
}

// NewHandler builds a Handler wired to the given profile/area/family
// application services.
func NewHandler(profiles profileService, areas areaService, families familyService) *Handler {
	return &Handler{
		profiles: profiles,
		areas:    areas,
		families: families,
	}
}
