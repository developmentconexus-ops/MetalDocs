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
	"metaldocs/internal/platform/apibase"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

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
	idempCreate    *idempotency.Store
	idempRevision  *idempotency.Store
	idempObsolete  *idempotency.Store
	idempSupersede *idempotency.Store
}

// NewHandler builds a Handler backed by svc, with dedicated idempotency
// stores for each mutating route (each keyed by its own route string, so a
// replay of one route can never collide with another). db is used only to
// construct those four stores (#90/A3.5 drive-by: the Handler used to also
// keep db on the struct, unread after construction — dead weight predating
// this change, deleted here rather than left for a future audit to
// rediscover) — the idempotency middleware resolves tenant/actor through its
// package-owned identity boundary, not from a handle threaded through Handler.
func NewHandler(svc *application.ControlledDocumentService, db *sql.DB) *Handler {
	return &Handler{
		svc:            svc,
		idempCreate:    idempotency.New(db, "POST /api/v1/controlled-documents"),
		idempRevision:  idempotency.New(db, "POST /api/v1/controlled-documents/{id}/revisions"),
		idempObsolete:  idempotency.New(db, "PUT /api/v1/controlled-documents/{id}/obsolete"),
		idempSupersede: idempotency.New(db, "PUT /api/v1/controlled-documents/{id}/supersede"),
	}
}

// injectTenant is a thin middleware that fails closed (500) if the tenant
// claim is absent from context before any mutating route runs. It used to
// also re-store the tenant under a local context key so the (now-deleted)
// idempotencyActor closure could read it without a *http.Request — that
// indirection is gone (#90/A3.5): Require reads tenant.FromContext(ctx)
// through its package-owned identity boundary, and the context this
// middleware passes through is unmodified, so it sees the same claim
// injectTenant already validated. injectTenant itself stays, unchanged in behavior, as the
// pre-existing fail-closed guard for this module's four mutating routes.
func injectTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := tenant.FromContext(r.Context()); err != nil {
			problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Name identifies this publisher in boot assertion messages.
func (h *Handler) Name() string { return "controlled-documents" }

// Tag is the OpenAPI tag this publisher owns.
func (h *Handler) Tag() string { return "controlled-documents" }

// matchesControlledDocumentSubPath reports whether path is a
// "/api/v1/controlled-documents/{id}/<suffix>"-shaped route: it starts with
// the tenant-scoped controlled-documents prefix and ends with suffix. Same
// index-based comparison the revisions/obsolete/supersede route matches in
// Mount used inline before extraction (not strings.HasPrefix/HasSuffix, to
// keep the match behavior byte-identical) — factored out once so Mount's own
// body stops repeating it three times.
func matchesControlledDocumentSubPath(path, suffix string) bool {
	const prefix = "/api/v1/controlled-documents/"
	return len(path) > len(prefix) &&
		path[:len(prefix)] == prefix &&
		len(path) >= len(suffix) &&
		path[len(path)-len(suffix):] == suffix
}

// Mount mounts the generated controlled-documents API surface
// onto mux under /api/v1, wrapping the two mutating POST routes with
// Idempotency-Key enforcement.
func (h *Handler) Mount(mux httprouter.Muxer) {
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && r.URL.Path == "/api/v1/controlled-documents" {
				injectTenant(idempotency.Require(h.idempCreate)(next)).ServeHTTP(w, r)
				return
			}
			if r.Method == http.MethodPost && matchesControlledDocumentSubPath(r.URL.Path, "/revisions") {
				injectTenant(idempotency.Require(h.idempRevision)(next)).ServeHTTP(w, r)
				return
			}
			// CON-06(b): obsolete/supersede are PUT lifecycle routes whose
			// application-layer changeStatus rejects a second call with
			// ErrCDNotActive (409) once the first call already transitioned the
			// row out of "active" — a genuine network-timeout retry of a
			// succeeded request must replay the original 204, not surface that
			// 409 as if the retry itself were invalid.
			if r.Method == http.MethodPut && matchesControlledDocumentSubPath(r.URL.Path, "/obsolete") {
				injectTenant(idempotency.Require(h.idempObsolete)(next)).ServeHTTP(w, r)
				return
			}
			if r.Method == http.MethodPut && matchesControlledDocumentSubPath(r.URL.Path, "/supersede") {
				injectTenant(idempotency.Require(h.idempSupersede)(next)).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	controlleddocumentsapi.HandlerWithOptions(h, controlleddocumentsapi.StdHTTPServerOptions{
		BaseRouter: mux,
		// AD-1: spec path keys are relative; the generated router prepends this
		// base so served routes stay /api/v1/* and the codegen matches the spec.
		BaseURL: apibase.BaseURL,
		Middlewares: []controlleddocumentsapi.MiddlewareFunc{
			middleware,
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			if requiresIdempotencyKey(r) && strings.Contains(err.Error(), "Idempotency-Key is required") {
				problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestIdempotencyKeyRequired, "Idempotency-Key header is required"))
				return
			}
			problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, err.Error()))
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
