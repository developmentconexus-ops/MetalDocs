package http

import (
	"context"
	"net/http"

	taxonomyapi "metaldocs/internal/modules/taxonomy/api"
	"metaldocs/internal/modules/taxonomy/application"
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

func NewHandler(profiles *application.ProfileService, areas *application.AreaService, families *application.FamilyService) *Handler {
	return &Handler{
		profiles: profiles,
		areas:    areas,
		families: families,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	generated := taxonomyapi.ServerInterfaceWrapper{
		Handler: h,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		},
	}
	mux.HandleFunc("GET /api/v1/taxonomy/profiles", generated.ListTaxonomyProfiles)
	mux.HandleFunc("POST /api/v1/taxonomy/profiles", generated.CreateTaxonomyProfile)
	mux.HandleFunc("GET /api/v1/taxonomy/profiles/{code}", generated.GetTaxonomyProfile)
	mux.HandleFunc("PATCH /api/v1/taxonomy/profiles/{code}", generated.UpdateTaxonomyProfile)
	mux.HandleFunc("DELETE /api/v1/taxonomy/profiles/{code}", generated.ArchiveTaxonomyProfile)
	mux.HandleFunc("PUT /api/v1/taxonomy/profiles/{code}/default-template", generated.SetTaxonomyProfileDefaultTemplate)
	mux.HandleFunc("GET /api/v1/taxonomy/areas", generated.ListTaxonomyAreas)
	mux.HandleFunc("POST /api/v1/taxonomy/areas", generated.CreateTaxonomyArea)
	mux.HandleFunc("GET /api/v1/taxonomy/areas/{code}", generated.GetTaxonomyArea)
	mux.HandleFunc("PUT /api/v1/taxonomy/areas/{code}", generated.UpdateTaxonomyArea)
	mux.HandleFunc("DELETE /api/v1/taxonomy/areas/{code}", generated.ArchiveTaxonomyArea)
	mux.HandleFunc("GET /api/v1/taxonomy/families", generated.ListTaxonomyFamilies)
	mux.HandleFunc("POST /api/v1/taxonomy/families", generated.CreateTaxonomyFamily)
	mux.HandleFunc("GET /api/v1/taxonomy/families/{code}", generated.GetTaxonomyFamily)
	mux.HandleFunc("PATCH /api/v1/taxonomy/families/{code}", generated.UpdateTaxonomyFamily)
	mux.HandleFunc("DELETE /api/v1/taxonomy/families/{code}", generated.DeactivateTaxonomyFamily)
}
