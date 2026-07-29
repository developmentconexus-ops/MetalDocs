// Package approvalhttp implements the approval subsystem's HTTP handlers:
// submit, decision, read, cancel, obsolete, and route admin. Every
// mutating handler enforces the If-Match / Idempotency-Key preconditions
// (OCC + safe-retry) before invoking the application layer, and errors are
// mapped to RFC 9457 problem+json via MapErrorToResponse.
package approvalhttp

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	approvalidempinfra "metaldocs/internal/modules/approval/infrastructure/idempotency"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/tenant"
)

type submitService interface {
	SubmitRevisionForReview(ctx context.Context, runner db.TxRunner, req application.SubmitRequest) (application.SubmitResult, error)
	// PreviewRoute (SLICE 8a, unit 3.2) is the read-only route-preview method
	// DocumentApprovalPreviewHandler calls — resolves the same active route
	// SubmitRevisionForReview would, without submitting.
	PreviewRoute(ctx context.Context, runner db.TxRunner, tenantID, documentID string) (application.RoutePreview, error)
}

type decisionService interface {
	RecordSignoff(ctx context.Context, runner db.TxRunner, req application.SignoffRequest) (application.SignoffResult, error)
}

type readService interface {
	LoadInstance(ctx context.Context, runner db.TxRunner, tenantID, instanceID string) (*domain.Instance, error)
	LoadActiveInstanceByDocument(ctx context.Context, runner db.TxRunner, tenantID, documentID string) (*domain.Instance, error)
	// LoadInstanceByDocumentForView is the view-only sibling used by the FE
	// GET handler: its status filter also admits changes_requested (unlike
	// LoadActiveInstanceByDocument, which stays narrow for publish/mutation).
	LoadInstanceByDocumentForView(ctx context.Context, runner db.TxRunner, tenantID, documentID string) (*domain.Instance, error)
	// LoadInstanceByDocumentForViewWithViewer (F2d.1, ADR 0078) is
	// LoadInstanceByDocumentForView's viewer-facts sibling: same load +
	// visibility gate, plus the server-derived domain.ViewerFacts for the
	// caller, computed in-tx via the shared write-path eligibility primitives.
	LoadInstanceByDocumentForViewWithViewer(ctx context.Context, runner db.TxRunner, tenantID, documentID string) (*domain.Instance, domain.ViewerFacts, []domain.ReviewVerdict, error)
	ListPendingForActor(ctx context.Context, runner db.TxRunner, tenantID, actorID string, areaCode string, limit, offset int) ([]domain.Instance, error)
	ListInboxItems(ctx context.Context, runner db.TxRunner, tenantID, actorID, areaCode string, limit, offset int) ([]application.InboxView, error)
	ListInboxItemsWithTotal(ctx context.Context, runner db.TxRunner, tenantID, actorID, areaCode string, limit, offset int) ([]application.InboxView, int, error)
	// ListWorklist (F8, spec.md §4/W4, §6.3 P2/P3/P8) extends the inbox listing
	// with stage_kind/due_before/scope=oversee filters.
	ListWorklist(ctx context.Context, runner db.TxRunner, tenantID, actorID, areaCode string, filter application.InboxFilter, limit, offset int) ([]application.InboxView, int, error)
	CountPendingForActor(ctx context.Context, runner db.TxRunner, tenantID, actorID, areaCode string) (int, error)
}

type cancelService interface {
	CancelInstance(ctx context.Context, runner db.TxRunner, req application.CancelInput) (application.CancelResult, error)
}

type reviewVerdictService interface {
	RecordVerdict(ctx context.Context, runner db.TxRunner, req application.ReviewVerdictRequest) (application.ReviewVerdictResult, error)
}

type fastForwardService interface {
	RecordFastForward(ctx context.Context, runner db.TxRunner, req application.FastForwardRequest) (application.FastForwardResult, error)
}

type obsoleteService interface {
	MarkObsolete(ctx context.Context, runner db.TxRunner, req application.MarkObsoleteRequest) (application.MarkObsoleteResult, error)
}

type routeAdminService interface {
	Create(ctx context.Context, runner db.TxRunner, in application.CreateRouteInput) (application.CreateRouteResult, error)
	Update(ctx context.Context, runner db.TxRunner, in application.UpdateRouteInput) (application.UpdateRouteResult, error)
	Deactivate(ctx context.Context, runner db.TxRunner, in application.DeactivateRouteInput) (application.DeactivateRouteResult, error)
	List(ctx context.Context, runner db.TxRunner, tenantID, actorID string) (application.ListRoutesResult, error)
}

var (
	// ErrIfMatchRequired is returned when a write endpoint is called without an If-Match header (OCC precondition).
	ErrIfMatchRequired = errors.New("precondition: If-Match header required")
	// ErrIfMatchMalformed is returned when the If-Match header is present but not "v<N>".
	ErrIfMatchMalformed = errors.New("precondition: If-Match header malformed; expected \"v<N>\"")
)

// signoffIdempStore backs idempotent replay for the signoff handlers. Slots are
// keyed by (tenantID, actorID, route template, idempotency key); payloadHash is a
// misuse guard that must be derived only from client-stable request inputs.
type signoffIdempStore interface {
	BeginDocumentReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (application.SignoffReplayCommitter, *application.SignoffReplay, error)
	BeginStageReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (application.SignoffReplayCommitter, *application.SignoffReplay, error)
}

// fastForwardIdempStore backs idempotent replay for the fast-forward handler.
// It is kept distinct from signoffIdempStore (its own route template) rather
// than reusing BeginStageReplay's stage-signoff template — see
// fastForwardRouteTemplate's doc comment for why the two must not collide.
type fastForwardIdempStore interface {
	BeginFastForwardReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (application.SignoffReplayCommitter, *application.SignoffReplay, error)
}

// Compile-time guard: the production idempotency store must keep implementing
// fastForwardIdempStore. NewHandler's type-assertion below is otherwise a
// silent no-op if PostgresSignoffIdempStore's method set ever drifts (e.g. a
// rename of BeginFastForwardReplay) — this assertion turns that into a build
// failure instead of a latent fail-open on fast-forward replay protection.
var _ fastForwardIdempStore = (*approvalidempinfra.PostgresSignoffIdempStore)(nil)

// Handler implements the approval subsystem's HTTP endpoints (submit, decision,
// read, cancel, obsolete, route admin). It wraps *application.Services
// behind narrow per-endpoint interfaces so handlers can be tested against fakes.
type Handler struct {
	services              *application.Services
	db                    *sql.DB
	runner                db.TxRunner
	submitSvc             submitService
	decisionSvc           decisionService
	readSvc               readService
	cancelSvc             cancelService
	reviewVerdictSvc      reviewVerdictService
	fastForwardSvc        fastForwardService
	obsoleteSvc           obsoleteService
	routeAdmin            routeAdminService
	idempStore            signoffIdempStore
	fastForwardIdempStore fastForwardIdempStore
	displayNameReader     iamdomain.UserDisplayNameReader
}

// NewHandler constructs the approval HTTP handler. displayName is a required
// collaborator for resolveEligibleActorNames; pass iamdomain.NoopUserDisplayNameReader{}
// for tests or paths that do not need display-name resolution.
func NewHandler(services *application.Services, database *sql.DB, idempStore signoffIdempStore, displayName iamdomain.UserDisplayNameReader) *Handler {
	h := &Handler{
		services:          services,
		db:                database,
		idempStore:        idempStore,
		displayNameReader: displayName,
	}
	if store, ok := idempStore.(fastForwardIdempStore); ok {
		h.fastForwardIdempStore = store
	}
	if database != nil {
		h.runner = db.NewTxRunner(database)
	}
	if services != nil {
		h.submitSvc = services.Submit
		h.decisionSvc = services.Decision
		h.readSvc = services.Read
		h.cancelSvc = services.Cancel
		h.reviewVerdictSvc = services.ReviewVerdict
		h.fastForwardSvc = services.FastForward
		h.obsoleteSvc = services.Obsolete
		h.routeAdmin = services.RouteAdmin
	}
	return h
}

func instanceETag(revisionVersion int) string {
	return "\"v" + strconv.Itoa(revisionVersion) + "\""
}

func actorIDFromRequest(r *http.Request) string {
	return iamdomain.UserIDFromContext(r.Context())
}

func tenantIDFromReq(r *http.Request) (string, error) {
	return tenant.FromContext(r.Context())
}

// idempotentHandler wraps next with the platform idempotency middleware, keyed
// to routeTemplate (used to scope the replay store, matching the templates
// sibling pattern). It validates presence + UUID format and provides
// replay-on-retry; callers no longer need to manually extract or validate the
// header. Named distinctly from the generated ServerInterface method set (no
// collision today, but router.go's Middlewares closure and this method both
// live on *Handler, so an explicit name keeps the two call paths unambiguous).
func (h *Handler) idempotentHandler(routeTemplate string, next http.Handler) http.Handler {
	store := idempotency.New(h.db, routeTemplate)
	return idempotency.Require(store, func(ctx context.Context) (string, string) {
		tenantID, _ := tenant.FromContext(ctx)
		return tenantID, iamdomain.UserIDFromContext(ctx)
	})(next)
}

// parseIfMatch parses the If-Match OCC precondition for handlers acting on an
// already-governed document — signoff, cancel, decision. Those documents have
// been through submit, which transitions draft→under_review and bumps
// revision_version to >= 1, so an explicit "v0" precondition is always stale and
// is rejected as malformed.
func parseIfMatch(header string) (int, error) {
	return parseIfMatchMin(header, 1)
}

// parseSubmitIfMatch parses the If-Match header for the canonical /submit
// handler (ADR 0073), whose target is a fresh draft at revision_version = 0.
// It therefore accepts an explicit "v0" precondition that parseIfMatch rejects;
// the two valid-version domains must stay separate (submit: v>=0, everything
// else: v>=1).
func parseSubmitIfMatch(header string) (int, error) {
	return parseIfMatchMin(header, 0)
}

// parseIfMatchMin is the shared If-Match parser; minVersion is the lowest
// explicit "v<N>" the caller accepts (1 for governed transitions, 0 for submit).
//
// The RFC 7232 "*" wildcard is deliberately NOT accepted. It used to be encoded
// as the sentinel version 0, which is not a separate state but a valid version:
// only mark-reviewed translated it back to "current", every other OCC site fed
// it straight into `WHERE revision_version = 0` (so "*" silently meant "expect
// v0" and 409'd), and on /submit — whose minVersion is 0 — it was flatly
// indistinguishable from an explicit "v0". Beyond the encoding, a wildcard on a
// controlled-document lifecycle transition is a request to write regardless of
// who wrote last, i.e. exactly the lost update the OCC guard exists to prevent.
// Preconditions on these routes fail closed: send the concrete "v<N>".
func parseIfMatchMin(header string, minVersion int) (int, error) {
	value := strings.TrimSpace(header)
	if value == "" {
		return -1, ErrIfMatchRequired
	}

	value = strings.Trim(value, "\"")
	if !strings.HasPrefix(value, "v") {
		return -1, ErrIfMatchMalformed
	}

	version, err := strconv.Atoi(strings.TrimPrefix(value, "v"))
	if err != nil || version < minVersion {
		return -1, ErrIfMatchMalformed
	}
	return version, nil
}
