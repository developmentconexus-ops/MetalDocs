package http

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"metaldocs/internal/modules/registry/application"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/tenant"
)

type tenantContextKey struct{}

type Handler struct {
	svc           *application.RegistryService
	db            *sql.DB
	idempCreate   *idempotency.Store
	idempRevision *idempotency.Store
}

func NewHandler(svc *application.RegistryService, db *sql.DB) *Handler {
	return &Handler{
		svc:           svc,
		db:            db,
		idempCreate:   idempotency.New(db, "POST /api/v2/controlled-documents"),
		idempRevision: idempotency.New(db, "POST /api/v2/controlled-documents/{id}/revisions"),
	}
}

// injectTenant is a thin middleware that reads X-Tenant-ID from the request
// header and stores it in the context so the idempotency actor closure can
// access it without a reference to the *http.Request.
func injectTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tid := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
		if tid == "" {
			tid = tenant.DevTenantID
		}
		ctx := context.WithValue(r.Context(), tenantContextKey{}, tid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func tenantIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(tenantContextKey{}).(string); ok && v != "" {
		return v
	}
	return tenant.DevTenantID
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	actorOf := func(ctx context.Context) (string, string) {
		return tenantIDFromContext(ctx), authn.UserIDFromContext(ctx)
	}

	mux.Handle("POST /api/v2/controlled-documents",
		injectTenant(idempotency.Require(h.idempCreate, actorOf)(http.HandlerFunc(h.atomicCreate))))
	mux.Handle("POST /api/v2/controlled-documents/{id}/revisions",
		injectTenant(idempotency.Require(h.idempRevision, actorOf)(http.HandlerFunc(h.createRevision))))
	mux.HandleFunc("GET /api/v2/controlled-documents/preview-code", h.previewCode)
	mux.HandleFunc("GET /api/v2/controlled-documents", h.listDocs)
	mux.HandleFunc("GET /api/v2/controlled-documents/{id}", h.getDoc)
	mux.HandleFunc("GET /api/v2/controlled-documents/{id}/active-document", h.getActiveDocument)
	mux.HandleFunc("PUT /api/v2/controlled-documents/{id}/obsolete", h.obsoleteDoc)
	mux.HandleFunc("PUT /api/v2/controlled-documents/{id}/supersede", h.supersedeDoc)
}
