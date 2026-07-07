// Package approvalhttp implements the approval subsystem's HTTP handlers:
// submit, decision, read, cancel, obsolete, supersede, and route admin. Every
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

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/tenant"
)

type submitService interface {
	SubmitRevisionForReview(ctx context.Context, runner db.TxRunner, req application.SubmitRequest) (application.SubmitResult, error)
}

type decisionService interface {
	RecordSignoff(ctx context.Context, runner db.TxRunner, req application.SignoffRequest) (application.SignoffResult, error)
}

type readService interface {
	LoadInstance(ctx context.Context, runner db.TxRunner, tenantID, instanceID string) (*domain.Instance, error)
	LoadActiveInstanceByDocument(ctx context.Context, runner db.TxRunner, tenantID, documentID string) (*domain.Instance, error)
	ListPendingForActor(ctx context.Context, runner db.TxRunner, tenantID, actorID string, areaCode string, limit, offset int) ([]domain.Instance, error)
	ListInboxItems(ctx context.Context, runner db.TxRunner, tenantID, actorID, areaCode string, limit, offset int) ([]application.InboxView, error)
	ListInboxItemsWithTotal(ctx context.Context, runner db.TxRunner, tenantID, actorID, areaCode string, limit, offset int) ([]application.InboxView, int, error)
	CountPendingForActor(ctx context.Context, runner db.TxRunner, tenantID, actorID, areaCode string) (int, error)
}

type cancelService interface {
	CancelInstance(ctx context.Context, runner db.TxRunner, req application.CancelInput) (application.CancelResult, error)
}

type reviewVerdictService interface {
	RecordVerdict(ctx context.Context, runner db.TxRunner, req application.ReviewVerdictRequest) (application.ReviewVerdictResult, error)
}

type obsoleteService interface {
	MarkObsolete(ctx context.Context, runner db.TxRunner, req application.MarkObsoleteRequest) (application.MarkObsoleteResult, error)
}

type supersedeService interface {
	PublishSuperseding(ctx context.Context, runner db.TxRunner, req application.SupersedeRequest) (application.SupersedeResult, error)
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
	// ErrIfMatchMalformed is returned when the If-Match header is present but not "v<N>" or "*".
	ErrIfMatchMalformed = errors.New("precondition: If-Match header malformed; expected \"v<N>\" or \"*\"")
)

// signoffIdempStore backs idempotent replay for the signoff handlers. Slots are
// keyed by (tenantID, actorID, route template, idempotency key); payloadHash is a
// misuse guard that must be derived only from client-stable request inputs.
type signoffIdempStore interface {
	BeginDocumentReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (application.SignoffReplayCommitter, *application.SignoffReplay, error)
	BeginStageReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (application.SignoffReplayCommitter, *application.SignoffReplay, error)
}

// Handler implements the approval subsystem's HTTP endpoints (submit, decision,
// read, cancel, obsolete, supersede, route admin). It wraps *application.Services
// behind narrow per-endpoint interfaces so handlers can be tested against fakes.
type Handler struct {
	services          *application.Services
	db                *sql.DB
	runner            db.TxRunner
	submitSvc         submitService
	decisionSvc       decisionService
	readSvc           readService
	cancelSvc         cancelService
	reviewVerdictSvc  reviewVerdictService
	obsoleteSvc       obsoleteService
	supersedeSvc      supersedeService
	routeAdmin        routeAdminService
	idempStore        signoffIdempStore
	displayNameReader iamdomain.UserDisplayNameReader
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
	if database != nil {
		h.runner = db.NewTxRunner(database)
	}
	if services != nil {
		h.submitSvc = services.Submit
		h.decisionSvc = services.Decision
		h.readSvc = services.Read
		h.cancelSvc = services.Cancel
		h.reviewVerdictSvc = services.ReviewVerdict
		h.obsoleteSvc = services.Obsolete
		h.supersedeSvc = services.Supersede
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
// is rejected as malformed. The "*" wildcard (match-any) maps to 0.
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
func parseIfMatchMin(header string, minVersion int) (int, error) {
	value := strings.TrimSpace(header)
	if value == "" {
		return -1, ErrIfMatchRequired
	}
	if value == "*" {
		return 0, nil
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
