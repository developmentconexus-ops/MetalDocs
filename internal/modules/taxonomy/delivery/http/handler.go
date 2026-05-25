package http

import (
	"context"
	"net/http"

	taxonomyapi "metaldocs/internal/modules/taxonomy/api"
	"metaldocs/internal/modules/taxonomy/domain"
)

type profileService interface {
	List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.DocumentProfile, error)
	Get(ctx context.Context, tenantID, code string) (*domain.DocumentProfile, error)
	Create(ctx context.Context, p *domain.DocumentProfile) error
	Update(ctx context.Context, p *domain.DocumentProfile) error
	SetDefaultTemplate(ctx context.Context, tenantID, profileCode, templateVersionID, actorID string) error
	Archive(ctx context.Context, tenantID, profileCode, actorID string) error
}

type areaService interface {
	List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.ProcessArea, error)
	Get(ctx context.Context, tenantID, code string) (*domain.ProcessArea, error)
	Create(ctx context.Context, a *domain.ProcessArea) error
	Update(ctx context.Context, a *domain.ProcessArea) error
	Archive(ctx context.Context, tenantID, areaCode, actorID string) error
}

type familyService interface {
	List(ctx context.Context, includeInactive bool) ([]domain.DocumentFamily, error)
	Get(ctx context.Context, code string) (*domain.DocumentFamily, error)
	Create(ctx context.Context, f *domain.DocumentFamily) error
	Update(ctx context.Context, f *domain.DocumentFamily) (*domain.DocumentFamily, error)
	Deactivate(ctx context.Context, code string) error
}

type Handler struct {
	profiles profileService
	areas    areaService
	families familyService
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	taxonomyapi.HandlerWithOptions(h, taxonomyapi.StdHTTPServerOptions{
		BaseRouter: mux,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		},
	})
}

func NewHandler(profiles profileService, areas areaService, families familyService) *Handler {
	return &Handler{
		profiles: profiles,
		areas:    areas,
		families: families,
	}
}
