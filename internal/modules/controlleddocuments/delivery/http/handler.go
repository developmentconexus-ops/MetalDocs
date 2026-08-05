// Package http is the controlled-documents module's HTTP delivery layer:
// it mounts the oapi-codegen-generated /api/v1/controlled-documents/*
// routes onto Handler, which dispatches to
// application.ControlledDocumentService and maps domain errors to the
// RFC 9457 problem+json envelope. Idempotency-Key handling for the four
// mutating routes (atomic create, create revision, obsolete, supersede) is
// wired here via injectTenant + idempotency.Require (CON-06).
package http

import (
	"context"
	"database/sql"
	"metaldocs/internal/platform/httprouter"
	"net/http"
	"strings"

	controlleddocumentsapi "metaldocs/internal/modules/controlleddocuments/api"
	"metaldocs/internal/modules/controlleddocuments/application"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

type tenantContextKey struct{}

type controlledDocumentService interface {
	Create(ctx context.Context, cmd application.CreateControlledDocumentCmd) (*application.CreateResult, error)
	List(ctx context.Context, tenantID string, filter application.CDFilter) ([]controlleddocumentsdomain.ControlledDocument, bool, error)
	CreateRevision(ctx context.Context, cmd application.CreateRevisionCmd) (*controlleddocumentsdomain.DocumentRef, error)
	Get(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ControlledDocument, error)
	GetActiveInstance(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ActiveDocumentInstance, error)
	Obsolete(ctx context.Context, tenantID, id string) error
	Supersede(ctx context.Context, tenantID, id string) error
	PeekSeq(ctx context.Context, tenantID, profileCode, areaCode string) (int, error)
	CreationContext(ctx context.Context, tenantID string) (*application.CreationContext, error)
}

// Handler implements the generated controlled-documents ServerInterface,
// dispatching each route to svc and translating results/errors to
// controlleddocumentsapi response types.
type Handler struct {
	svc            controlledDocumentService
	db             *sql.DB
	idempCreate    *idempotency.Store
	idempRevision  *idempotency.Store
	idempObsolete  *idempotency.Store
	idempSupersede *idempotency.Store
}

// NewHandler builds a Handler backed by svc, with dedicated idempotency
// stores for each mutating route (each keyed by its own route string, so a
// replay of one route can never collide with another).
func NewHandler(svc *application.ControlledDocumentService, db *sql.DB) *Handler {
	return &Handler{
		svc:            svc,
		db:             db,
		idempCreate:    idempotency.New(db, "POST /api/v1/controlled-documents"),
		idempRevision:  idempotency.New(db, "POST /api/v1/controlled-documents/{id}/revisions"),
		idempObsolete:  idempotency.New(db, "PUT /api/v1/controlled-documents/{id}/obsolete"),
		idempSupersede: idempotency.New(db, "PUT /api/v1/controlled-documents/{id}/supersede"),
	}
}

// injectTenant is a thin middleware that reads the tenant from context (set by
// auth middleware) and re-stores it under a local key so the idempotency actor
// closure can access it without a reference to the *http.Request.
func injectTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tid, err := tenant.FromContext(r.Context())
		if err != nil {
			httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
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

// RegisterRoutes mounts the generated controlled-documents API surface
// onto mux under /api/v1, wrapping the two mutating POST routes with
// Idempotency-Key enforcement.
func (h *Handler) RegisterRoutes(mux httprouter.Muxer) {
	actorOf := func(ctx context.Context) (string, string) {
		// Idempotency scoping is best-effort here: a missing actor
		// produces a broader-but-still-safe key. The mutation handler
		// itself fail-closes on missing actor (see writeDomainError).
		userID, _ := authn.UserIDFromContext(ctx)
		return tenantIDFromContext(ctx), userID
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
			// CON-06(b): obsolete/supersede are PUT lifecycle routes whose
			// application-layer changeStatus rejects a second call with
			// ErrCDNotActive (409) once the first call already transitioned the
			// row out of "active" — a genuine network-timeout retry of a
			// succeeded request must replay the original 204, not surface that
			// 409 as if the retry itself were invalid.
			if r.Method == http.MethodPut &&
				len(r.URL.Path) > len("/api/v1/controlled-documents/") &&
				r.URL.Path[:len("/api/v1/controlled-documents/")] == "/api/v1/controlled-documents/" &&
				len(r.URL.Path) >= len("/obsolete") &&
				r.URL.Path[len(r.URL.Path)-len("/obsolete"):] == "/obsolete" {
				injectTenant(idempotency.Require(h.idempObsolete, actorOf)(next)).ServeHTTP(w, r)
				return
			}
			if r.Method == http.MethodPut &&
				len(r.URL.Path) > len("/api/v1/controlled-documents/") &&
				r.URL.Path[:len("/api/v1/controlled-documents/")] == "/api/v1/controlled-documents/" &&
				len(r.URL.Path) >= len("/supersede") &&
				r.URL.Path[len(r.URL.Path)-len("/supersede"):] == "/supersede" {
				injectTenant(idempotency.Require(h.idempSupersede, actorOf)(next)).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	controlleddocumentsapi.HandlerWithOptions(h, controlleddocumentsapi.StdHTTPServerOptions{
		BaseRouter: mux,
		// AD-1: spec path keys are relative; the generated router prepends this
		// base so served routes stay /api/v1/* and the codegen matches the spec.
		BaseURL: "/api/v1",
		Middlewares: []controlleddocumentsapi.MiddlewareFunc{
			middleware,
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			if requiresIdempotencyKey(r) && strings.Contains(err.Error(), "Idempotency-Key is required") {
				httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestIdempotencyKeyRequired, "Idempotency-Key header is required")
				return
			}
			httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestInvalid, err.Error())
		},
	})
}

func requiresIdempotencyKey(r *http.Request) bool {
	switch r.Method {
	case http.MethodPost:
		if r.URL.Path == "/api/v1/controlled-documents" {
			return true
		}
		return strings.HasPrefix(r.URL.Path, "/api/v1/controlled-documents/") &&
			strings.HasSuffix(r.URL.Path, "/revisions")
	case http.MethodPut:
		if !strings.HasPrefix(r.URL.Path, "/api/v1/controlled-documents/") {
			return false
		}
		return strings.HasSuffix(r.URL.Path, "/obsolete") || strings.HasSuffix(r.URL.Path, "/supersede")
	default:
		return false
	}
}
