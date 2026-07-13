// Package http implements the templates module's HTTP delivery layer: it
// adapts the generated templatesapi.ServerInterface (oapi-codegen, from
// api/openapi) to the application.Service use cases, performing tier-1
// capability authz, request/response mapping to the OpenAPI-generated DTOs,
// and error translation to RFC 9457 problem+json responses.
package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	approvalapp "metaldocs/internal/modules/approval/application"
	approvaldomain "metaldocs/internal/modules/approval/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/application"
	platformdb "metaldocs/internal/platform/db"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// Narrow, testable seams onto the approval kernel's published services (M3
// P3.S2b-4, R2a). Mirrors approval/http's own submitService/decisionService/
// readService interfaces (internal/modules/approval/http/handler.go) — the
// concrete *approvalapp.TemplateSubmitService / *approvalapp.DecisionService /
// *approvalapp.ReadService types satisfy these implicitly, but tests can
// substitute a fake without exercising the real signature-reauth/DB path.
type approvalSubmitService interface {
	SubmitTemplateVersionForReview(ctx context.Context, runner platformdb.TxRunner, req approvalapp.TemplateSubmitRequest) (approvalapp.TemplateSubmitResult, error)
}

type approvalDecisionService interface {
	RecordSignoff(ctx context.Context, runner platformdb.TxRunner, req approvalapp.SignoffRequest) (approvalapp.SignoffResult, error)
}

type approvalReadService interface {
	LoadActiveInstanceBySubjectForMutation(ctx context.Context, runner platformdb.TxRunner, tenantID, subjectKind, subjectKey string) (*approvaldomain.Instance, error)
}

// AuthzFunc is the tier-1 route→capability check for template routes.
//
// All template call-sites pass area "*" (area-agnostic) on purpose: every
// CapTemplate* capability is ScopeTenant per ADR 0022, and the wired
// implementation (apps/api/cmd/metaldocs-api/main.go) deliberately ignores the
// area argument for these tenant-grade caps, resolving them with CanDo at
// tenant scope. The "*" is therefore an intentional "no narrower area applies"
// marker, not a missing area — passing "tenant" here would be equivalent. The
// real area check is the in-tx tier-2 authz.Require in the application layer.
// (F-T6 part b: documents the deliberate area mismatch flagged by the audit.)
type AuthzFunc func(r *http.Request, tenantID, area string, action string) error

// Handler implements the templates module's HTTP delivery layer. It wires
// the application.Service to the generated templatesapi.ServerInterface,
// enforcing tier-1 capability authz via AuthzFunc and mapping between the
// domain and the OpenAPI-generated wire DTOs.
type Handler struct {
	svc               *application.Service
	authz             AuthzFunc
	db                *sql.DB
	displayNameReader iamdomain.UserDisplayNameReader

	// Approval-kernel wiring (M3 P3.S2b-4, R2a): optional — nil until
	// WithApprovalKernel is called. The two thin kernel entry points
	// (submit-for-approval, signoff) delegate to these published
	// approval/application services; templates never imports
	// approval/infrastructure or approval/delivery.
	approvalSubmit   approvalSubmitService
	approvalDecision approvalDecisionService
	approvalRead     approvalReadService
	approvalRunner   platformdb.TxRunner
}

// New constructs a Handler wired to the given application service, tier-1
// authz function, and database handle (used for idempotency-key storage).
// Panics if authz is nil, since every route depends on it.
func New(svc *application.Service, authz AuthzFunc, database *sql.DB) *Handler {
	if authz == nil {
		panic("templates http: authz function is required")
	}
	return &Handler{svc: svc, authz: authz, db: database, displayNameReader: iamdomain.NoopUserDisplayNameReader{}}
}

// WithDisplayNameReader wires the iam-owned UserDisplayNameReader port (M4/F4.1)
// used to resolve TemplateDTO.created_by_display_name without templates
// querying iam tables/repos directly. Optional — defaults to
// iamdomain.NoopUserDisplayNameReader{} (created_by_display_name omitted) when
// never called, matching the pattern in documents/approval's handler wiring.
func (h *Handler) WithDisplayNameReader(r iamdomain.UserDisplayNameReader) *Handler {
	if r != nil {
		h.displayNameReader = r
	}
	return h
}

// WithApprovalKernel wires the approval module's published kernel services
// (M3 P3.S2b-4, R2a) used by the thin submit-for-approval/signoff routes.
// Optional — nil kernel fields cause those two routes to 500 rather than
// panic (see nilApprovalKernel guard in routes_approval_kernel.go); every
// other route is unaffected. runner is the shared db.TxRunner the kernel
// application-layer services expect (mirrors documents' composition-root
// wiring of the same approval services).
func (h *Handler) WithApprovalKernel(submit approvalSubmitService, decision approvalDecisionService, read approvalReadService, runner platformdb.TxRunner) *Handler {
	h.approvalSubmit = submit
	h.approvalDecision = decision
	h.approvalRead = read
	h.approvalRunner = runner
	return h
}

// resolveCreatedByDisplayName resolves a single created_by display name via
// the iam-owned port (FE-08). Returns "" (field omitted on the wire) on any
// resolution failure or miss — display-name enrichment is best-effort and
// must never fail the request that carries the real payload.
func (h *Handler) resolveCreatedByDisplayName(ctx context.Context, tenantID, userID string) string {
	if userID == "" {
		return ""
	}
	name, err := h.displayNameReader.DisplayName(ctx, tenantID, userID)
	if err != nil {
		return ""
	}
	return name
}

// resolveCreatedByDisplayNames batch-resolves created_by display names for a
// list of templates via the iam-owned port (FE-08), avoiding one DisplayName
// call per row. Returns userID -> displayName; misses are simply absent, so
// callers fall back to the raw id (mirrors resolveEligibleActorNames in
// documents/approval/http/get_instance_handler.go).
func (h *Handler) resolveCreatedByDisplayNames(ctx context.Context, tenantID string, userIDs []string) map[string]string {
	unique := make(map[string]struct{}, len(userIDs))
	ids := make([]string, 0, len(userIDs))
	for _, id := range userIDs {
		if id == "" {
			continue
		}
		if _, ok := unique[id]; ok {
			continue
		}
		unique[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return map[string]string{}
	}
	names, err := h.displayNameReader.DisplayNames(ctx, tenantID, ids)
	if err != nil {
		return map[string]string{}
	}
	return names
}

// idempotentRoutes lists the route templates (method + path, matching the
// net/http 1.22 mux pattern syntax) that require an Idempotency-Key per
// api/openapi/v1/openapi.yaml. Kept as a set so the Middlewares closure below
// can dispatch per-route without hand-registering each mutation separately —
// CON-03 templates slice: mounts the generated ServerInterface via
// HandlerWithOptions instead of a hand-written route list, so a spec'd route
// can no longer silently go unserved (missing ServerInterface methods fail
// to compile). Mirrors the controlleddocuments pattern
// (internal/modules/controlleddocuments/delivery/http/handler.go).
var idempotentRoutes = map[string]bool{
	"POST /api/v1/templates":                                       true,
	"POST /api/v1/templates/{id}/versions/{n}/publish":             true,
	"POST /api/v1/templates/{id}/versions/{n}/submit":              true,
	"POST /api/v1/templates/{id}/versions/{n}/review":              true,
	"POST /api/v1/templates/{id}/versions/{n}/approve":             true,
	"POST /api/v1/templates/{id}/versions/{n}/submit-for-approval": true,
	"POST /api/v1/templates/{id}/versions/{n}/signoff":             true,
}

// Register mounts the generated templates routes onto mux via
// templatesapi.HandlerWithOptions, applying an idempotency-key middleware to
// the mutation routes listed in idempotentRoutes.
func (h *Handler) Register(mux *http.ServeMux) {
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// r.Pattern already carries the method prefix ("POST /api/v1/...")
			// because the generated router registers method-qualified patterns —
			// do NOT prepend r.Method again or the lookup silently misses.
			routeTemplate := r.Pattern
			if idempotentRoutes[routeTemplate] {
				h.idempotent(routeTemplate, next).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	templatesapi.HandlerWithOptions(h, templatesapi.StdHTTPServerOptions{
		BaseRouter: mux,
		// AD-1: spec path keys are relative; the generated router prepends this
		// base so served routes stay /api/v1/* and the codegen matches the spec.
		BaseURL: "/api/v1",
		Middlewares: []templatesapi.MiddlewareFunc{
			middleware,
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			writeErr(w, http.StatusBadRequest, codeTplInvalidRequest, err.Error())
		},
	})
}

func (h *Handler) idempotent(routeTemplate string, next http.Handler) http.Handler {
	store := idempotency.New(h.db, routeTemplate)
	return idempotency.Require(store, func(ctx context.Context) (string, string) {
		tenantID, _ := tenant.FromContext(ctx)
		return tenantID, iamdomain.UserIDFromContext(ctx)
	})(next)
}

var (
	writeJSON = httpresponse.WriteJSON
	readJSON  = httpresponse.ReadJSON
)

func readStrictJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return errors.New("request body must contain a single JSON value")
		}
		return err
	}
	return nil
}

func tenantIDFromReq(r *http.Request) (string, error) {
	return tenant.FromContext(r.Context())
}

func userIDFromReq(r *http.Request) string {
	return iamdomain.UserIDFromContext(r.Context())
}

func writeErr(w http.ResponseWriter, status int, code problem.Code, message string) {
	if err := problem.Write(w, problem.New(status, code, message)); err != nil {
		slog.Warn("templates http: problem write failed", "err", err, "status", status, "code", code)
	}
}

var friendlyMsg = map[problem.Code]string{
	codeTplUploadMissing: "DOCX file not yet uploaded. Please upload the template file before submitting for review.",
}

func writeMappedErr(w http.ResponseWriter, err error) {
	status, code := MapErr(err)
	msg := friendlyMsg[code]
	if msg == "" {
		msg = err.Error()
	}
	if msg == "" {
		msg = string(code)
	}
	writeErr(w, status, code, msg)
}
