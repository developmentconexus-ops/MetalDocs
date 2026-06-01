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
	approvalinfra "metaldocs/internal/modules/documents/approval/infrastructure"
	"metaldocs/internal/modules/documents/approval/repository"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

type submitService interface {
	SubmitRevisionForReview(ctx context.Context, db *sql.DB, req application.SubmitRequest) (application.SubmitResult, error)
}

type decisionService interface {
	RecordSignoff(ctx context.Context, db *sql.DB, req application.SignoffRequest) (application.SignoffResult, error)
}

type readService interface {
	LoadInstance(ctx context.Context, db *sql.DB, tenantID, instanceID string) (*domain.Instance, error)
	LoadActiveInstanceByDocument(ctx context.Context, db *sql.DB, tenantID, documentID string) (*domain.Instance, error)
	ListPendingForActor(ctx context.Context, db *sql.DB, tenantID, actorID string, areaCode string, limit, offset int) ([]domain.Instance, error)
	ListInboxItems(ctx context.Context, db *sql.DB, tenantID, actorID, areaCode string, limit, offset int) ([]application.InboxView, error)
	CountPendingForActor(ctx context.Context, db *sql.DB, tenantID, actorID, areaCode string) (int, error)
}

type cancelService interface {
	CancelInstance(ctx context.Context, db *sql.DB, req application.CancelInput) (application.CancelResult, error)
}

type obsoleteService interface {
	MarkObsolete(ctx context.Context, db *sql.DB, req application.MarkObsoleteRequest) (application.MarkObsoleteResult, error)
}

type supersedeService interface {
	PublishSuperseding(ctx context.Context, db *sql.DB, req application.SupersedeRequest) (application.SupersedeResult, error)
}

type routeAdminService interface {
	Create(ctx context.Context, db *sql.DB, in application.CreateRouteInput) (application.CreateRouteResult, error)
	Update(ctx context.Context, db *sql.DB, in application.UpdateRouteInput) (application.UpdateRouteResult, error)
	Deactivate(ctx context.Context, db *sql.DB, in application.DeactivateRouteInput) (application.DeactivateRouteResult, error)
}

type routeListRepository interface {
	ListRoutes(ctx context.Context, tenantID string) ([]repository.Route, error)
}

var (
	ErrIfMatchRequired  = errors.New("precondition: If-Match header required")
	ErrIfMatchMalformed = errors.New("precondition: If-Match header malformed; expected \"v<N>\" or \"*\"")
)

// signoffIdempStore backs idempotent replay for the signoff handlers. Slots are
// keyed by (tenantID, actorID, route template, idempotency key); payloadHash is a
// misuse guard that must be derived only from client-stable request inputs.
type signoffIdempStore interface {
	BeginDocumentReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (approvalinfra.SignoffReplayCommitter, *approvalinfra.SignoffReplay, error)
	BeginStageReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (approvalinfra.SignoffReplayCommitter, *approvalinfra.SignoffReplay, error)
}

type Handler struct {
	services     *application.Services
	db           *sql.DB
	submitSvc    submitService
	decisionSvc  decisionService
	readSvc      readService
	cancelSvc    cancelService
	obsoleteSvc  obsoleteService
	supersedeSvc supersedeService
	routeAdmin   routeAdminService
	routeRepo    routeListRepository
	idempStore   signoffIdempStore
}

func NewHandler(services *application.Services, db *sql.DB, idempStore signoffIdempStore) *Handler {
	h := &Handler{
		services:   services,
		db:         db,
		idempStore: idempStore,
	}
	if services != nil {
		h.submitSvc = services.Submit
		h.decisionSvc = services.Decision
		h.readSvc = services.Read
		h.cancelSvc = services.Cancel
		h.obsoleteSvc = services.Obsolete
		h.supersedeSvc = services.Supersede
		h.routeAdmin = services.RouteAdmin
	}
	if db != nil {
		h.routeRepo = repository.NewPostgresApprovalRepository(db)
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

func parseIfMatch(header string) (int, error) {
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
	if err != nil || version <= 0 {
		return -1, ErrIfMatchMalformed
	}
	return version, nil
}
