// Package http is the taxonomy module's HTTP delivery layer: it mounts the
// 16 oapi-codegen-generated /api/v1/taxonomy/* routes onto Handler, which
// dispatches to the profile/area/family application services and maps
// domain errors to the RFC 9457 problem+json envelope. Idempotency-Key
// handling for the three create routes (profile, area, family) is wired
// here via idempotency.Require (CON-06(d)) — a retried create after a
// dropped response would otherwise surface the second call's unique-index
// 409 instead of replaying the first call's 201.
package http

import (
	"context"
	"database/sql"
	"metaldocs/internal/platform/httprouter"
	"net/http"

	taxonomyapi "metaldocs/internal/modules/taxonomy/api"
	"metaldocs/internal/modules/taxonomy/application"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/apibase"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

type profileService interface {
	List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.DocumentProfile, error)
	Get(ctx context.Context, tenantID string, code domain.ProfileCode) (*domain.DocumentProfile, error)
	Create(ctx context.Context, p *domain.DocumentProfile) error
	Update(ctx context.Context, p *domain.DocumentProfile) error
	Reclassify(ctx context.Context, tenantID string, profileCode domain.ProfileCode, newClass domain.GovernanceClass, actorID string) error
	SetDefaultTemplate(ctx context.Context, tenantID string, profileCode domain.ProfileCode, templateVersionID, actorID string) error
	Archive(ctx context.Context, tenantID string, profileCode domain.ProfileCode, actorID string) error
	// RouteReadySubjects returns the profile codes with an active approval
	// route, per subject kind, backing DocumentProfileItem.has_active_route
	// and .has_active_template_route.
	RouteReadySubjects(ctx context.Context, tenantID string) (application.RouteReadiness, error)
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
// application services. It holds no persistence logic of its own beyond the
// three create-route idempotency stores.
type Handler struct {
	profiles     profileService
	areas        areaService
	families     familyService
	idempProfile *idempotency.Store
	idempArea    *idempotency.Store
	idempFamily  *idempotency.Store
}

// idempotentCreateRoutes lists the route templates (method + path, matching
// the net/http 1.22 mux pattern syntax) that require an Idempotency-Key per
// api/openapi/v1/openapi.yaml (CON-06(d)). Update/archive/deactivate routes
// are keyed by the path-param `code`, so a retry naturally targets the same
// row and is not included here — only the three creates carry a
// unique-index collision risk on retry.
var idempotentCreateRoutes = map[string]func(*Handler) *idempotency.Store{
	"POST /api/v1/taxonomy/profiles": func(h *Handler) *idempotency.Store { return h.idempProfile },
	"POST /api/v1/taxonomy/areas":    func(h *Handler) *idempotency.Store { return h.idempArea },
	"POST /api/v1/taxonomy/families": func(h *Handler) *idempotency.Store { return h.idempFamily },
}

// Name identifies this publisher in boot assertion messages.
func (h *Handler) Name() string { return "taxonomy" }

// Tag is the OpenAPI tag this publisher owns.
func (h *Handler) Tag() string { return "taxonomy" }

// Mount mounts the 16 generated taxonomy operations onto mux under
// the "/api/v1" base, translating unmarshalable request bodies to a 400
// VALIDATION_ERROR problem+json response. The three create routes are
// additionally gated by the platform idempotency middleware.
func (h *Handler) Mount(mux httprouter.Muxer) {
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// r.Pattern already carries the method prefix ("POST /api/v1/...")
			// because the generated router registers method-qualified patterns —
			// do NOT prepend r.Method again or the lookup silently misses.
			if storeOf, ok := idempotentCreateRoutes[r.Pattern]; ok {
				idempotency.Require(storeOf(h), actorFromCtx)(next).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	taxonomyapi.HandlerWithOptions(h, taxonomyapi.StdHTTPServerOptions{
		BaseRouter: mux,
		// AD-1: spec path keys are relative; the generated router prepends this
		// base so served routes stay /api/v1/* and the codegen matches the spec.
		BaseURL: apibase.BaseURL,
		Middlewares: []taxonomyapi.MiddlewareFunc{
			middleware,
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			writeError(w, r, http.StatusBadRequest, problem.CodeRequestInvalid, err.Error())
		},
	})
}

// actorFromCtx extracts (tenantID, actorID) for idempotency-key scoping,
// matching the templates/approval/controlleddocuments sibling handlers.
func actorFromCtx(ctx context.Context) (string, string) {
	tenantID, _ := tenant.FromContext(ctx)
	actorID, _ := authn.UserIDFromContext(ctx)
	return tenantID, actorID
}

// NewHandler builds a Handler wired to the given profile/area/family
// application services and db (used solely for the three create-route
// idempotency stores, mirroring templates/approval/controlleddocuments).
func NewHandler(profiles profileService, areas areaService, families familyService, db *sql.DB) *Handler {
	return &Handler{
		profiles:     profiles,
		areas:        areas,
		families:     families,
		idempProfile: idempotency.New(db, "POST /api/v1/taxonomy/profiles"),
		idempArea:    idempotency.New(db, "POST /api/v1/taxonomy/areas"),
		idempFamily:  idempotency.New(db, "POST /api/v1/taxonomy/families"),
	}
}
