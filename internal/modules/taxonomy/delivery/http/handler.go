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
	mux.HandleFunc("GET /api/v2/taxonomy/profiles", generated.ListTaxonomyProfilesV2)
	mux.HandleFunc("POST /api/v2/taxonomy/profiles", generated.CreateTaxonomyProfileV2)
	mux.HandleFunc("GET /api/v2/taxonomy/profiles/{code}", generated.GetTaxonomyProfileV2)
	mux.HandleFunc("PATCH /api/v2/taxonomy/profiles/{code}", generated.UpdateTaxonomyProfileV2)
	mux.HandleFunc("DELETE /api/v2/taxonomy/profiles/{code}", generated.ArchiveTaxonomyProfileV2)
	mux.HandleFunc("PUT /api/v2/taxonomy/profiles/{code}/default-template", generated.SetTaxonomyProfileDefaultTemplateV2)
	mux.HandleFunc("GET /api/v2/taxonomy/areas", generated.ListTaxonomyAreasV2)
	mux.HandleFunc("POST /api/v2/taxonomy/areas", generated.CreateTaxonomyAreaV2)
	mux.HandleFunc("GET /api/v2/taxonomy/areas/{code}", generated.GetTaxonomyAreaV2)
	mux.HandleFunc("PUT /api/v2/taxonomy/areas/{code}", generated.UpdateTaxonomyAreaV2)
	mux.HandleFunc("DELETE /api/v2/taxonomy/areas/{code}", generated.ArchiveTaxonomyAreaV2)
	mux.HandleFunc("GET /api/v2/taxonomy/families", generated.ListTaxonomyFamiliesV2)
	mux.HandleFunc("POST /api/v2/taxonomy/families", generated.CreateTaxonomyFamilyV2)
	mux.HandleFunc("GET /api/v2/taxonomy/families/{code}", generated.GetTaxonomyFamilyV2)
	mux.HandleFunc("PATCH /api/v2/taxonomy/families/{code}", generated.UpdateTaxonomyFamilyV2)
	mux.HandleFunc("DELETE /api/v2/taxonomy/families/{code}", generated.DeactivateTaxonomyFamilyV2)
}
