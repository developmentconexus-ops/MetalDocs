package http

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	controlleddocumentsapi "metaldocs/internal/modules/controlleddocuments/api"
	"metaldocs/internal/modules/controlleddocuments/application"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/tenant"
)

type tenantContextKey struct{}

type controlledDocumentService interface {
	Create(ctx context.Context, cmd application.CreateControlledDocumentCmd) (*application.CreateResult, error)
	List(ctx context.Context, tenantID string, filter application.CDFilter) ([]controlleddocumentsdomain.ControlledDocument, error)
	CreateRevision(ctx context.Context, cmd application.CreateRevisionCmd) (*controlleddocumentsdomain.DocumentRef, error)
	Get(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ControlledDocument, error)
	Obsolete(ctx context.Context, tenantID, id string) error
	Supersede(ctx context.Context, tenantID, id string) error
	PeekSeq(ctx context.Context, tenantID, profileCode, areaCode string) (int, error)
}

type Handler struct {
	svc           controlledDocumentService
	db            *sql.DB
	idempCreate   *idempotency.Store
	idempRevision *idempotency.Store
}

func NewHandler(svc *application.ControlledDocumentService, db *sql.DB) *Handler {
	return &Handler{
		svc:           svc,
		db:            db,
		idempCreate:   idempotency.New(db, "POST /api/v1/controlled-documents"),
		idempRevision: idempotency.New(db, "POST /api/v1/controlled-documents/{id}/revisions"),
	}
}

// injectTenant is a thin middleware that reads the tenant from context (set by
// auth middleware) and re-stores it under a local key so the idempotency actor
// closure can access it without a reference to the *http.Request.
func injectTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tid, err := tenant.FromContext(r.Context())
		if err != nil {
			httpresponse.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			return
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

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && r.URL.Path == "/api/v1/controlled-documents" {
				injectTenant(idempotency.Require(h.idempCreate, actorOf)(next)).ServeHTTP(w, r)
				return
			}
			if r.Method == http.MethodPost &&
				len(r.URL.Path) > len("/api/v1/controlled-documents/") &&
				r.URL.Path[:len("/api/v1/controlled-documents/")] == "/api/v1/controlled-documents/" &&
				len(r.URL.Path) >= len("/revisions") &&
				r.URL.Path[len(r.URL.Path)-len("/revisions"):] == "/revisions" {
				injectTenant(idempotency.Require(h.idempRevision, actorOf)(next)).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	controlleddocumentsapi.HandlerWithOptions(h, controlleddocumentsapi.StdHTTPServerOptions{
		BaseRouter: mux,
		Middlewares: []controlleddocumentsapi.MiddlewareFunc{
			middleware,
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			if requiresIdempotencyKey(r) && strings.Contains(err.Error(), "Idempotency-Key is required") {
				httpresponse.WriteError(w, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED", "Idempotency-Key header is required")
				return
			}
			httpresponse.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		},
	})
}

func requiresIdempotencyKey(r *http.Request) bool {
	if r.Method != http.MethodPost {
		return false
	}
	if r.URL.Path == "/api/v1/controlled-documents" {
		return true
	}
	return strings.HasPrefix(r.URL.Path, "/api/v1/controlled-documents/") &&
		strings.HasSuffix(r.URL.Path, "/revisions")
}
