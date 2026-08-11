package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"metaldocs/internal/platform/httprouter"
	"net/http"
	"strconv"
	"strings"
	"time"

	approvalapp "metaldocs/internal/modules/approval/application"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	documentsapi "metaldocs/internal/modules/documents/api"
	"metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/domain"
	iamapp "metaldocs/internal/modules/iam/application"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/apibase"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/pagination"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/ratelimit"
	"metaldocs/internal/platform/tenant"

	"github.com/google/uuid"
)

// documentReader covers read/list/stats/ownership queries.
type documentReader interface {
	GetDocument(ctx context.Context, tenantID, id string) (*domain.Document, error)
	ListDocumentsPaginated(ctx context.Context, tenantID, userID string, opts application.ListOptions) ([]*domain.Document, int64, bool, error)
	DocumentStats(ctx context.Context, tenantID, userID string, opts application.ListOptions) (*application.DocumentStats, error)
	IsDocumentOwner(ctx context.Context, tenantID, docID, userID string) (bool, error)
	// RequireDocumentView asserts the actor holds CapDocumentView (tenant-grade).
	// Returns authz.ErrCapDenied when denied (maps to 403 via mapErr).
	RequireDocumentView(ctx context.Context, tenantID, actorID, docID string) error
}

// documentLifecycle covers mutation operations on a document's lifecycle.
type documentLifecycle interface {
	DuplicateDocument(ctx context.Context, tenantID, userID, docID string) (*application.CreateDocumentResult, error)
	RenameDocument(ctx context.Context, tenantID, userID, docID, newName string) error
	Archive(ctx context.Context, tenantID, docID, actorID string) error
	CreateCheckpoint(ctx context.Context, tenantID, docID, actorID, label string) (*domain.Checkpoint, error)
	ListCheckpoints(ctx context.Context, tenantID, actorID, docID string) ([]domain.Checkpoint, error)
	ListRevisionHistory(ctx context.Context, tenantID, docID string) ([]domain.RevisionHistoryItem, error)
	RestoreCheckpoint(ctx context.Context, tenantID, docID, actorID string, versionNum int) (*application.RestoreResult, error)
	SignedRevisionURL(ctx context.Context, tenantID, docID, revID string) (string, error)
}

// editingSessions covers real-time editing session management.
type editingSessions interface {
	AcquireSession(ctx context.Context, tenantID, docID, userID string) (*domain.Session, bool, error)
	HeartbeatSession(ctx context.Context, sessionID, userID string) error
	ReleaseSession(ctx context.Context, tenantID, sessionID, userID, docID string) error
	ForceReleaseSession(ctx context.Context, tenantID, adminID, sessionID, docID string) error
}

// autosaveArtifacts covers autosave presign/commit operations.
type autosaveArtifacts interface {
	PresignAutosave(ctx context.Context, cmd application.PresignAutosaveCmd) (*application.PresignAutosaveResult, error)
	CommitAutosave(ctx context.Context, cmd application.CommitAutosaveCmd) (*application.CommitResult, error)
}

// documentComments covers comment CRUD on a document.
type documentComments interface {
	ListDocumentComments(ctx context.Context, tenantID, documentID string) ([]domain.Comment, error)
	AddDocumentComment(ctx context.Context, tenantID, userID, authorDisplay, documentID string, in domain.CommentCreateInput) (*domain.Comment, error)
	UpdateDocumentComment(ctx context.Context, tenantID, userID, documentID string, libraryID int, in domain.CommentUpdateInput) (*domain.Comment, error)
	DeleteDocumentComment(ctx context.Context, tenantID, userID, documentID string, libraryID int) error
}

// Service is the application boundary consumed by Handler. It composes
// cohesive sub-interfaces so each responsibility group can be mocked or
// swapped independently.
type Service interface {
	documentReader
	documentLifecycle
	editingSessions
	autosaveArtifacts
	documentComments
}

// Handler serves the documents module's core HTTP routes (list/get/rename/
// archive/duplicate, autosave, checkpoints, sessions, comments) and delegates
// to optional sub-handlers (export, fillIn, placeholderOpts, view,
// reconstruct) wired via WithSubHandlers.
type Handler struct {
	svc             Service
	caps            application.CapabilityChecker
	export          *ExportHandler
	fillIn          *FillInHandler
	placeholderOpts *PlaceholderOptionsHandler
	view            *ViewHandler
	reconstruct     *ReconstructHandler
	// rl and userFn back the per-route rate limiting Mount applies (see
	// registerRoutes/rateLimitedRoutes below). Both nil is a valid,
	// non-rate-limited configuration (matches the prior RegisterRoutes'
	// (nil, nil) behavior); set together via WithRateLimit.
	rl     *ratelimit.Middleware
	userFn func(*http.Request) string
}

var writeJSON = httpresponse.WriteJSON

// WithCaps binds the tier-1 capability checker used to resolve system_admin.
func (h *Handler) WithCaps(c application.CapabilityChecker) *Handler {
	h.caps = c
	return h
}

// WithSubHandlers attaches optional sub-handlers that cover routes outside the
// core Handler. Routes for nil sub-handlers are skipped in registerRoutes.
func (h *Handler) WithSubHandlers(export *ExportHandler, fillIn *FillInHandler, placeholderOpts *PlaceholderOptionsHandler, view *ViewHandler, reconstruct *ReconstructHandler) *Handler {
	h.export = export
	h.fillIn = fillIn
	h.placeholderOpts = placeholderOpts
	h.view = view
	h.reconstruct = reconstruct
	return h
}

// isSystemAdmin resolves admin via the capability model. A nil checker fails
// closed (false) so callers fall back to the ownership-scoped path.
func (h *Handler) isSystemAdmin(ctx context.Context, userID, tenantID string) (bool, error) {
	if h.caps == nil {
		return false, nil
	}
	return h.caps.IsSystemAdmin(ctx, userID, tenantID)
}

// NewHandler constructs a Handler backed by the given Service.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// WithRateLimit wires the per-route rate limiter and user-identity extractor
// Mount applies to rateLimitedRoutes. Both nil (the zero value) keeps Mount's
// prior no-rate-limit behavior -- this is a constructor field, not a second
// mounting mode: Mount always calls the same registerRoutes body.
func (h *Handler) WithRateLimit(rl *ratelimit.Middleware, userFn func(*http.Request) string) *Handler {
	h.rl = rl
	h.userFn = userFn
	return h
}

// rateLimitedRoutes maps each rate-limited route's method-qualified pattern
// (as it appears on r.Pattern once matched by the generated router — see the
// Middlewares closure below) to its ratelimit.RouteKey. CON-03/ARC-02: mounts
// the generated ServerInterface via HandlerWithOptions instead of a
// hand-written mux.HandleFunc list, so a route the spec adds can no longer
// silently go unserved (missing ServerInterface methods fail to compile).
// Mirrors the templates slice's idempotentRoutes map
// (internal/modules/templates/delivery/http/handler.go).
var rateLimitedRoutes = map[string]ratelimit.RouteKey{
	"POST /api/v1/documents/{id}/autosave/presign": ratelimit.RouteAutosavePresign,
	"POST /api/v1/documents/{id}/autosave/commit":  ratelimit.RouteAutosaveCommit,
	"POST /api/v1/documents/{id}/export/pdf":       ratelimit.RouteExportPDF,
}

// Name identifies this publisher in boot assertion messages.
func (h *Handler) Name() string { return "documents" }

// Tag is the OpenAPI tag this publisher owns.
func (h *Handler) Tag() string { return "documents" }

// Mount registers documents' routes on mux, applying rate limiting when
// WithRateLimit configured it (both fields nil is a valid unlimited mount).
func (h *Handler) Mount(mux httprouter.Muxer) { h.registerRoutes(mux, h.rl, h.userFn) }

func (h *Handler) registerRoutes(mux httprouter.Muxer, rl *ratelimit.Middleware, userFn func(*http.Request) string) {
	rateLimited := rl != nil && userFn != nil

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// r.Pattern already carries the method prefix ("POST /api/v1/...")
			// because the generated router registers method-qualified patterns —
			// do NOT prepend r.Method again or the lookup silently misses (see
			// templates/delivery/http/handler.go for the incident this guards).
			if rateLimited {
				if key, ok := rateLimitedRoutes[r.Pattern]; ok {
					rl.Limit(key, userFn, next).ServeHTTP(w, r)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}

	documentsapi.HandlerWithOptions(h, documentsapi.StdHTTPServerOptions{
		BaseRouter: mux,
		// AD-1: spec path keys are relative; the generated router prepends this
		// base so served routes stay /api/v1/* and the codegen matches the spec.
		BaseURL: apibase.BaseURL,
		Middlewares: []documentsapi.MiddlewareFunc{
			middleware,
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, err.Error()))
		},
	})
}

func (h *Handler) listDocuments(w http.ResponseWriter, r *http.Request, params documentsapi.ListDocumentsParams) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, problem.CodeInternalUnknown.String()))
		return
	}
	callerUserID, err := userIDFromReq(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
		return
	}
	isAdmin, err := h.isSystemAdmin(r.Context(), callerUserID, tenantID)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, problem.CodeInternalUnknown.String()))
		return
	}
	opts, effectiveUserID, err := listOptionsFromParams(r, params, callerUserID, isAdmin)
	if err != nil {
		slog.Warn("documents listDocuments invalid query params", "err", err)
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}

	items, total, hasMore, err := h.svc.ListDocumentsPaginated(r.Context(), tenantID, effectiveUserID, opts)
	if err != nil {
		if errors.Is(err, pagination.ErrInvalidCursor) {
			problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestCursorInvalid, problem.CodeRequestCursorInvalid.String()))
			return
		}
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}

	page := documentsapi.CursorPage{HasMore: hasMore}
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		nextCursor := pagination.EncodeCursor(last.UpdatedAt.UTC().Format(time.RFC3339Nano), last.ID)
		page.NextCursor = &nextCursor
	}

	summaries := make([]documentsapi.DocumentSummary, 0, len(items))
	for _, d := range items {
		s, err := toDocumentSummary(*d)
		if err != nil {
			problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, problem.CodeInternalUnknown.String()))
			return
		}
		summaries = append(summaries, s)
	}

	httpresponse.WriteJSON(w, http.StatusOK, documentsapi.DocumentListResponse{
		Items: summaries,
		Page:  page,
		Total: total,
	})
}

func (h *Handler) documentStats(w http.ResponseWriter, r *http.Request, _ documentsapi.DocumentStatsParams) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, problem.CodeInternalUnknown.String()))
		return
	}
	callerUserID, err := userIDFromReq(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
		return
	}
	isAdmin, err := h.isSystemAdmin(r.Context(), callerUserID, tenantID)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, problem.CodeInternalUnknown.String()))
		return
	}
	// DocumentStatsParams (operationId documentStats) is a strict subset of
	// ListDocumentsParams (status / area_code / profile_code only — no
	// cursor/limit/q/include_archived). parseListOptions already derives
	// status/area_code/profile_code from the same raw query values the wrapper
	// bound into params, including the repeated-?status= form the single-string
	// params.Status cannot represent (see listOptionsFromParams), so there is no
	// separate params-consuming path here to duplicate without diverging.
	opts, effectiveUserID, err := parseListOptions(r, callerUserID, isAdmin)
	if err != nil {
		slog.Warn("documents documentStats invalid query params", "err", err)
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}

	stats, err := h.svc.DocumentStats(r.Context(), tenantID, effectiveUserID, opts)
	if err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, documentsapi.DocumentStatsResponse{
		ByArea:   stats.ByArea,
		ByStatus: stats.ByStatus,
	})
}

func parseListOptions(r *http.Request, callerUserID string, isAdmin bool) (application.ListOptions, string, error) {
	query := r.URL.Query()
	opts := application.ListOptions{
		PageSize: 20,
	}

	// FD-2: keyset cursor pagination. `cursor` is opaque (validated by the repo
	// on decode); `limit` clamps server-side to 1..100.
	opts.Cursor = strings.TrimSpace(query.Get("cursor"))

	if rawLimit := strings.TrimSpace(query.Get("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil {
			return opts, "", errors.New("limit must be a valid integer")
		}
		if limit < 1 {
			return opts, "", errors.New("limit must be >= 1")
		}
		if limit > 100 {
			return opts, "", errors.New("limit must be <= 100")
		}
		opts.PageSize = limit
	}

	statuses, err := statusFilterFromParams(r)
	if err != nil {
		return opts, "", err
	}
	opts.Status = statuses

	opts.AreaCode = strings.TrimSpace(query.Get("area_code"))
	opts.ProfileCode = strings.TrimSpace(query.Get("profile_code"))
	opts.Q = strings.TrimSpace(query.Get("q"))

	includeArchived := strings.TrimSpace(query.Get("include_archived"))
	if includeArchived != "" {
		v, err := strconv.ParseBool(includeArchived)
		if err != nil {
			return opts, "", errors.New("include_archived must be a valid boolean")
		}
		opts.IncludeArchived = v
	}

	reviewDue := strings.TrimSpace(query.Get("review_due"))
	if reviewDue != "" {
		v, err := strconv.ParseBool(reviewDue)
		if err != nil {
			return opts, "", errors.New("review_due must be a valid boolean")
		}
		opts.ReviewDue = v
	}

	effectiveUserID := ""
	if !isAdmin && callerUserID != "" {
		opts.CreatedBy = callerUserID
		effectiveUserID = callerUserID
	}

	return opts, effectiveUserID, nil
}

// statusFilterFromParams reads the `status` query parameter directly off the
// request, honoring both a single CSV value ("draft,active") and repeated
// occurrences (?status=draft&status=active) per the spec's documented
// behavior for GET /documents (operationId listDocuments) and its statuses
// query surface. Every known-status validation lives here so both
// parseListOptions and listOptionsFromParams share one source of truth.
func statusFilterFromParams(r *http.Request) ([]string, error) {
	statusValues := r.URL.Query()["status"]
	if len(statusValues) == 0 {
		return nil, nil
	}
	statuses := make([]string, 0, len(statusValues))
	for _, raw := range statusValues {
		for _, split := range strings.Split(raw, ",") {
			s := strings.TrimSpace(split)
			if s == "" {
				continue
			}
			if !isKnownDocumentStatus(s) {
				return nil, fmt.Errorf("invalid status %q", s)
			}
			statuses = append(statuses, s)
		}
	}
	return statuses, nil
}

// listOptionsFromParams builds application.ListOptions for listDocuments from the
// generated ListDocumentsParams the wrapper has already bound and type-checked
// (ARC-02 — no re-parsing r.URL.Query() for fields the typed params already carry).
// The one exception is `status`: the spec documents it as "CSV, also accepts
// repeated query params" (api/openapi/v1/openapi.yaml operationId listDocuments),
// but ListDocumentsParams.Status is a single *string (the form binder only captures
// one occurrence), so the repeated-param case is only observable via r.URL.Query()
// directly — statusFilterFromParams reads that raw multi-value form, which is a
// strict superset of params.Status's single-CSV-value case (identical result when
// the client sends exactly one ?status= occurrence).
func listOptionsFromParams(r *http.Request, params documentsapi.ListDocumentsParams, callerUserID string, isAdmin bool) (application.ListOptions, string, error) {
	opts := application.ListOptions{PageSize: 20}

	if params.Cursor != nil {
		opts.Cursor = strings.TrimSpace(*params.Cursor)
	}

	if params.Limit != nil {
		limit := *params.Limit
		if limit < 1 {
			return opts, "", errors.New("limit must be >= 1")
		}
		if limit > 100 {
			return opts, "", errors.New("limit must be <= 100")
		}
		opts.PageSize = limit
	}

	statuses, err := statusFilterFromParams(r)
	if err != nil {
		return opts, "", err
	}
	opts.Status = statuses

	if params.AreaCode != nil {
		opts.AreaCode = strings.TrimSpace(*params.AreaCode)
	}
	if params.ProfileCode != nil {
		opts.ProfileCode = strings.TrimSpace(*params.ProfileCode)
	}
	if params.Q != nil {
		opts.Q = strings.TrimSpace(*params.Q)
	}
	if params.IncludeArchived != nil {
		opts.IncludeArchived = *params.IncludeArchived
	}
	if params.ReviewDue != nil {
		opts.ReviewDue = *params.ReviewDue
	}

	effectiveUserID := ""
	if !isAdmin && callerUserID != "" {
		opts.CreatedBy = callerUserID
		effectiveUserID = callerUserID
	}

	return opts, effectiveUserID, nil
}

func (h *Handler) getDocument(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, _, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	doc, err := h.svc.GetDocument(ctx, tenantID, docID)
	if err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	resp, err := toDocumentDetailResponse(*doc)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, problem.CodeInternalUnknown.String()))
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, resp)
}

// toDocumentSummary maps a domain document to the generated DocumentSummary
// response type (A6 — the list endpoint must emit generated types, not raw
// domain structs, so FE codegen stays in sync).
func toDocumentSummary(doc domain.Document) (documentsapi.DocumentSummary, error) {
	formData := doc.FormDataJSON
	if len(formData) == 0 {
		formData = []byte(`{}`)
	}
	var formMap map[string]interface{}
	if err := json.Unmarshal(formData, &formMap); err != nil {
		return documentsapi.DocumentSummary{}, fmt.Errorf("invalid document form_data_json for document %s: %w", doc.ID, err)
	}
	return documentsapi.DocumentSummary{
		ActiveSessionId:         doc.ActiveSessionID,
		ArchivedAt:              doc.ArchivedAt,
		Code:                    doc.Code,
		ControlledDocumentId:    doc.ControlledDocumentID,
		CreatedAt:               doc.CreatedAt,
		CreatedBy:               doc.CreatedBy,
		CurrentRevisionId:       doc.CurrentRevisionID,
		EffectiveFrom:           doc.EffectiveFrom,
		EffectiveTo:             doc.EffectiveTo,
		FormDataJson:            formMap,
		Id:                      doc.ID,
		LastReviewedAt:          doc.LastReviewedAt,
		Name:                    doc.Name,
		ProcessAreaCodeSnapshot: doc.ProcessAreaCodeSnapshot,
		ProfileCodeSnapshot:     doc.ProfileCodeSnapshot,
		ReviewDueAt:             doc.ReviewDueAt,
		ReviewSurfacedAt:        doc.ReviewSurfacedAt,
		RevisionNumber:          doc.RevisionNumber,
		RevisionTitle:           doc.RevisionTitle,
		RevisionVersion:         doc.RevisionVersion,
		Status:                  documentsapi.DocumentSummaryStatus(doc.Status),
		TemplateVersionId:       doc.TemplateVersionID,
		TenantId:                doc.TenantID,
		UpdatedAt:               doc.UpdatedAt,
		ValuesFrozenAt:          doc.ValuesFrozenAt,
	}, nil
}

// toDocumentDetailResponse maps a domain document to the generated
// DocumentDetailResponse type (A6). form_data_json is parsed into the object
// the generated contract declares (map[string]interface{}).
func toDocumentDetailResponse(doc domain.Document) (*documentsapi.DocumentDetailResponse, error) {
	formData := doc.FormDataJSON
	if len(formData) == 0 {
		formData = []byte(`{}`)
	}
	var formMap map[string]interface{}
	if err := json.Unmarshal(formData, &formMap); err != nil {
		return nil, fmt.Errorf("invalid document form_data_json for document %s: %w", doc.ID, err)
	}

	var pageCountSource *documentsapi.DocumentDetailResponseCurrentRevisionPageCountSource
	if doc.CurrentRevisionPageCountSource != nil {
		v := documentsapi.DocumentDetailResponseCurrentRevisionPageCountSource(*doc.CurrentRevisionPageCountSource)
		pageCountSource = &v
	}

	release, err := toDocumentReleaseProjection(doc)
	if err != nil {
		return nil, err
	}

	return &documentsapi.DocumentDetailResponse{
		ActiveSessionId:                doc.ActiveSessionID,
		ArchivedAt:                     doc.ArchivedAt,
		Code:                           doc.Code,
		ControlledDocumentId:           doc.ControlledDocumentID,
		CreatedAt:                      doc.CreatedAt,
		CreatedBy:                      doc.CreatedBy,
		CurrentRevisionFileSizeBytes:   doc.CurrentRevisionFileSizeBytes,
		CurrentRevisionId:              doc.CurrentRevisionID,
		CurrentRevisionPageCount:       doc.CurrentRevisionPageCount,
		CurrentRevisionPageCountSource: pageCountSource,
		EffectiveFrom:                  doc.EffectiveFrom,
		EffectiveTo:                    doc.EffectiveTo,
		FormDataJson:                   formMap,
		Id:                             doc.ID,
		LastReviewedAt:                 doc.LastReviewedAt,
		Name:                           doc.Name,
		ProcessAreaCodeSnapshot:        doc.ProcessAreaCodeSnapshot,
		ProfileCodeSnapshot:            doc.ProfileCodeSnapshot,
		ReviewDueAt:                    doc.ReviewDueAt,
		Release:                        release,
		ReviewSurfacedAt:               doc.ReviewSurfacedAt,
		RevisionNumber:                 doc.RevisionNumber,
		RevisionTitle:                  doc.RevisionTitle,
		RevisionVersion:                doc.RevisionVersion,
		Status:                         string(doc.Status),
		TemplateVersionId:              doc.TemplateVersionID,
		TenantId:                       doc.TenantID,
		UpdatedAt:                      doc.UpdatedAt,
		ValuesFrozenAt:                 doc.ValuesFrozenAt,
	}, nil
}

// toDocumentReleaseProjection maps the ADR 0085 Stage C readiness-hold
// projection composed by GetDocument onto the generated wire type. Returns
// (nil, nil) when the document carries no release generation — the wire field
// is required+nullable, so that serializes as `"release": null` (present-and-
// null, ADR 0035), never an absent key and never a synthesized hold.
func toDocumentReleaseProjection(doc domain.Document) (*documentsapi.DocumentReleaseProjection, error) {
	if doc.Release == nil {
		return nil, nil
	}
	generationID, err := uuid.Parse(doc.Release.GenerationID)
	if err != nil {
		return nil, fmt.Errorf("invalid release generation id %q for document %s: %w", doc.Release.GenerationID, doc.ID, err)
	}

	var holdReason *documentsapi.DocumentReleaseProjectionHoldReason
	if doc.Release.HoldReason != nil {
		v := documentsapi.DocumentReleaseProjectionHoldReason(*doc.Release.HoldReason)
		holdReason = &v
	}

	return &documentsapi.DocumentReleaseProjection{
		GenerationId:         generationID,
		State:                documentsapi.DocumentReleaseProjectionState(doc.Release.State),
		HoldReason:           holdReason,
		HoldDetail:           doc.Release.HoldDetail,
		PlannedEffectiveFrom: doc.Release.PlannedEffectiveFrom,
		ReleasedAt:           doc.Release.ReleasedAt,
		LastEvaluatedAt:      doc.Release.LastEvaluatedAt,
	}, nil
}

func (h *Handler) renameDocument(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}
	if !isValidBoundedText(req.Name, 255) {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}

	if err := h.svc.RenameDocument(ctx, tenantID, userID, docID, req.Name); err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}

	// F1.2 / ADR 0035 — renameDocument returns 200 OK with empty body. OpenAPI declares no
	// response schema. Consumer FE adapter is Promise<void>. No domain.Document leak.
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) archiveDocument(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	if err := h.svc.Archive(ctx, tenantID, docID, userID); err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) duplicateDocument(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	res, err := h.svc.DuplicateDocument(ctx, tenantID, userID, docID)
	if err != nil {
		status, msg := mapErr(err)
		if status == http.StatusInternalServerError {
			slog.Error("documents duplicate failed", "doc_id", docID, "tenant_id", tenantID, "actor_id", userID, "err", err) //nolint:gosec // G706: slog default is JSONHandler (set at process start) — control chars are JSON-escaped, log-line injection not possible
			problem.Respond(w, r, problem.New(status, msg, msg.String()))
			return
		}
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	docUUID, ridUUID, sidUUID, parseErr := parseCreateResultUUIDs(res.DocumentID, res.InitialRevisionID, res.SessionID)
	if parseErr != nil {
		slog.Error("documents duplicate produced unparseable id", "doc_id", docID, "tenant_id", tenantID, "actor_id", userID, "err", parseErr) //nolint:gosec // G706: slog default is JSONHandler (set at process start) — control chars are JSON-escaped, log-line injection not possible
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, problem.CodeInternalUnknown.String()))
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, documentsapi.DocumentCreateResult{
		DocumentId:        docUUID,
		InitialRevisionId: ridUUID,
		SessionId:         sidUUID,
	})
}

// parseCreateResultUUIDs parses the three server-generated id strings of a
// duplicate/create result into typed UUIDs for the generated DocumentCreateResult
// body. Any parse failure is a server-side invariant violation (the ids are
// freshly generated), surfaced as a 500 by the caller.
func parseCreateResultUUIDs(documentID, initialRevisionID, sessionID string) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	docUUID, err := uuid.Parse(documentID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("document_id: %w", err)
	}
	ridUUID, err := uuid.Parse(initialRevisionID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("initial_revision_id: %w", err)
	}
	sidUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("session_id: %w", err)
	}
	return docUUID, ridUUID, sidUUID, nil
}

func (h *Handler) acquireSession(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	sess, readonly, err := h.svc.AcquireSession(ctx, tenantID, docID, userID)
	if err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}

	if readonly {
		httpresponse.WriteJSON(w, http.StatusOK, documentsapi.DocumentSessionReadonlyResponse{
			Mode:      documentsapi.Readonly,
			HeldBy:    sess.UserID,
			HeldUntil: sess.ExpiresAt,
		})
		return
	}
	sessID, err := uuid.Parse(sess.ID)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	lastAckID, err := uuid.Parse(sess.LastAcknowledgedRevisionID)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, documentsapi.DocumentSessionWriterResponse{
		Mode:              documentsapi.Writer,
		SessionId:         sessID,
		ExpiresAt:         sess.ExpiresAt,
		LastAckRevisionId: lastAckID,
	})
}

func (h *Handler) heartbeatSession(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	_, userID, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}

	if err := h.svc.HeartbeatSession(ctx, req.SessionID, userID); err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) releaseSession(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}

	if err := h.svc.ReleaseSession(ctx, tenantID, req.SessionID, userID, docID); err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) forceReleaseSession(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, problem.CodeInternalUnknown.String()))
		return
	}
	adminID, err := userIDFromReq(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}

	if err := h.svc.ForceReleaseSession(ctx, tenantID, adminID, req.SessionID, docID); err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) presignAutosave(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	var req struct {
		SessionID      string `json:"session_id"`
		BaseRevisionID string `json:"base_revision_id"`
		ContentHash    string `json:"content_hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}

	res, err := h.svc.PresignAutosave(ctx, application.PresignAutosaveCmd{
		TenantID:       tenantID,
		ActorUserID:    userID,
		DocumentID:     docID,
		SessionID:      req.SessionID,
		BaseRevisionID: req.BaseRevisionID,
		ContentHash:    req.ContentHash,
	})
	if err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	pendingID, err := uuid.Parse(res.PendingUploadID)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, documentsapi.DocumentAutosavePresignResponse{
		UploadUrl:       res.UploadURL,
		PendingUploadId: pendingID,
		ExpiresAt:       res.ExpiresAt,
	})
}

func (h *Handler) commitAutosave(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	var req struct {
		SessionID        string          `json:"session_id"`
		PendingUploadID  string          `json:"pending_upload_id"`
		FormDataSnapshot json.RawMessage `json:"form_data_snapshot"`
		PageCount        *int            `json:"page_count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}
	// form_data_snapshot is optional — absent means "this autosave carries no
	// form-data change" and the write path preserves the stored form data. When
	// it IS present the contract types it as an object, so reject null/scalars/
	// arrays here: persisting one would make every later read of the document
	// fail to decode form_data_json into the contract's object type (one bad
	// write poisoning a read path with 500s).
	if len(req.FormDataSnapshot) > 0 && !isJSONObject(req.FormDataSnapshot) {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}

	res, err := h.svc.CommitAutosave(ctx, application.CommitAutosaveCmd{
		TenantID:         tenantID,
		ActorUserID:      userID,
		DocumentID:       docID,
		SessionID:        req.SessionID,
		PendingUploadID:  req.PendingUploadID,
		FormDataSnapshot: req.FormDataSnapshot,
		PageCount:        req.PageCount,
	})
	if err != nil {
		slog.Error("documents.commit_autosave failed", "doc_id", docID, "tenant_id", tenantID, "actor_id", userID, "session_id", redactID(req.SessionID), "pending_upload_id", redactID(req.PendingUploadID), "err", err) //nolint:gosec // G706: slog default is JSONHandler (set at process start) — control chars are JSON-escaped, log-line injection not possible
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	commitRevID, err := uuid.Parse(res.RevisionID)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	idempotentReplay := res.AlreadyConsumed
	commitRevNum := int(res.RevisionNum)
	var commitPageCountSource *documentsapi.CommitDocumentAutosave200JSONResponseBodyPageCountSource
	if res.PageCountSource != nil {
		src := documentsapi.CommitDocumentAutosave200JSONResponseBodyPageCountSource(*res.PageCountSource)
		commitPageCountSource = &src
	}
	httpresponse.WriteJSON(w, http.StatusOK, documentsapi.CommitDocumentAutosave200JSONResponse{
		RevisionId:       commitRevID,
		RevisionNum:      commitRevNum,
		IdempotentReplay: &idempotentReplay,
		FileSizeBytes:    res.FileSizeBytes,
		PageCount:        res.PageCount,
		PageCountSource:  commitPageCountSource,
	})
}

func (h *Handler) listCheckpoints(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	items, err := h.svc.ListCheckpoints(ctx, tenantID, userID, docID)
	if err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	resp, err := toAPICheckpoints(items)
	if err != nil {
		slog.Error("documents.list_checkpoints malformed uuid", "doc_id", docID, "tenant_id", tenantID, "err", err) //nolint:gosec // G706: slog default is JSONHandler (set at process start) — control chars are JSON-escaped, log-line injection not possible
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, problem.CodeInternalUnknown.String()))
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) listRevisionHistory(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, _, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	items, err := h.svc.ListRevisionHistory(ctx, tenantID, docID)
	if err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}

	histItems, err := toAPIRevisionHistoryItems(items)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, problem.CodeInternalUnknown.String()))
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, documentsapi.DocumentRevisionHistoryResponse{
		Items: histItems,
	})
}

func toAPICheckpoint(cp domain.Checkpoint) (documentsapi.DocumentCheckpoint, error) {
	id, err := uuid.Parse(cp.ID)
	if err != nil {
		return documentsapi.DocumentCheckpoint{}, fmt.Errorf("checkpoint id %q: %w", cp.ID, err)
	}
	docID, err := uuid.Parse(cp.DocumentID)
	if err != nil {
		return documentsapi.DocumentCheckpoint{}, fmt.Errorf("checkpoint document_id %q: %w", cp.DocumentID, err)
	}
	revID, err := uuid.Parse(cp.RevisionID)
	if err != nil {
		return documentsapi.DocumentCheckpoint{}, fmt.Errorf("checkpoint revision_id %q: %w", cp.RevisionID, err)
	}
	return documentsapi.DocumentCheckpoint{
		Id:         id,
		DocumentId: docID,
		RevisionId: revID,
		VersionNum: cp.VersionNum,
		Label:      cp.Label,
		CreatedAt:  cp.CreatedAt,
		CreatedBy:  cp.CreatedBy,
	}, nil
}

func toAPICheckpoints(cps []domain.Checkpoint) ([]documentsapi.DocumentCheckpoint, error) {
	out := make([]documentsapi.DocumentCheckpoint, 0, len(cps))
	for _, c := range cps {
		mapped, err := toAPICheckpoint(c)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}
	return out, nil
}

func toAPIRevisionHistoryItems(items []domain.RevisionHistoryItem) ([]documentsapi.DocumentRevisionHistoryItem, error) {
	out := make([]documentsapi.DocumentRevisionHistoryItem, 0, len(items))
	for _, item := range items {
		docID, err := uuid.Parse(item.DocumentID)
		if err != nil {
			return nil, fmt.Errorf("revision history document_id %q: %w", item.DocumentID, err)
		}
		out = append(out, documentsapi.DocumentRevisionHistoryItem{
			DocumentId:     docID,
			RevisionNumber: item.RevisionNumber,
			RevisionTitle:  item.RevisionTitle,
			Status:         string(item.Status),
			CreatedAt:      item.CreatedAt,
			IsCurrent:      item.IsCurrent,
		})
	}
	return out, nil
}

func (h *Handler) createCheckpoint(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	var req struct {
		Label string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}
	if !isValidBoundedText(req.Label, 255) {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}

	cp, err := h.svc.CreateCheckpoint(ctx, tenantID, docID, userID, req.Label)
	if err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	resp, err := toAPICheckpoint(*cp)
	if err != nil {
		slog.Error("documents.create_checkpoint malformed uuid", "doc_id", docID, "tenant_id", tenantID, "err", err) //nolint:gosec // G706: slog default is JSONHandler (set at process start) — control chars are JSON-escaped, log-line injection not possible
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, problem.CodeInternalUnknown.String()))
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) restoreCheckpoint(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	versionNum, err := strconv.Atoi(r.PathValue("version"))
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}

	res, err := h.svc.RestoreCheckpoint(ctx, tenantID, docID, userID, versionNum)
	if err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	newRevID, err := uuid.Parse(res.NewRevisionID)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	newRevNum := int(res.NewRevisionNum)
	httpresponse.WriteJSON(w, http.StatusOK, documentsapi.RestoreDocumentCheckpoint200JSONResponse{
		NewRevisionId:              newRevID,
		NewRevisionNum:             newRevNum,
		SourceCheckpointVersionNum: &versionNum,
		Idempotent:                 res.Idempotent,
	})
}

func (h *Handler) signedRevisionURL(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, _, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	url, err := h.svc.SignedRevisionURL(ctx, tenantID, docID, r.PathValue("rid"))
	if err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, documentsapi.RevisionUrlResponse{Url: url})
}

func (h *Handler) listComments(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, _, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	comments, err := h.svc.ListDocumentComments(ctx, tenantID, docID)
	if err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	resp := make([]documentsapi.DocumentCommentResponse, 0, len(comments))
	for i := range comments {
		resp = append(resp, toCommentResponse(comments[i]))
	}
	httpresponse.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) createComment(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	var req struct {
		LibraryCommentID int             `json:"library_comment_id"`
		ParentLibraryID  *int            `json:"parent_library_id"`
		AuthorDisplay    string          `json:"author_display"`
		Content          json.RawMessage `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}

	comment, err := h.svc.AddDocumentComment(ctx, tenantID, userID, req.AuthorDisplay, docID, domain.CommentCreateInput{
		LibraryCommentID: req.LibraryCommentID,
		ParentLibraryID:  req.ParentLibraryID,
		AuthorDisplay:    req.AuthorDisplay,
		ContentJSON:      req.Content,
	})
	if err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, toCommentResponse(*comment))
}

func (h *Handler) updateComment(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	libraryID, err := strconv.Atoi(r.PathValue("library_id"))
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}
	var req struct {
		Content *json.RawMessage `json:"content"`
		Done    *bool            `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}

	comment, err := h.svc.UpdateDocumentComment(ctx, tenantID, userID, docID, libraryID, domain.CommentUpdateInput{
		ContentJSON: req.Content,
		Done:        req.Done,
	})
	if err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, toCommentResponse(*comment))
}

func (h *Handler) deleteComment(w http.ResponseWriter, r *http.Request) {
	ctx := withAdminCtx(r.Context(), r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, ctx, docID)
	if !ok {
		return
	}

	libraryID, err := strconv.Atoi(r.PathValue("library_id"))
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, problem.CodeRequestInvalid.String()))
		return
	}
	if err := h.svc.DeleteDocumentComment(ctx, tenantID, userID, docID, libraryID); err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func toCommentResponse(c domain.Comment) documentsapi.DocumentCommentResponse {
	return documentsapi.DocumentCommentResponse{
		Id:               c.ID,
		LibraryCommentId: c.LibraryCommentID,
		ParentLibraryId:  c.ParentLibraryID,
		Author:           c.AuthorDisplay,
		AuthorId:         c.AuthorID,
		Content:          decodeCommentContent(c.ContentJSON),
		Done:             c.ResolvedAt != nil,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
		ResolvedAt:       c.ResolvedAt,
	}
}

// decodeCommentContent decodes a comment's stored content JSON (an array of
// rich-text nodes) into the generated typed node slice. The generated Content
// field is non-omitempty, so a nil/empty/invalid blob serializes as [] (never
// null) to honor the published contract.
func decodeCommentContent(raw json.RawMessage) []documentsapi.DocumentCommentContentNode {
	nodes := []documentsapi.DocumentCommentContentNode{}
	if len(raw) == 0 {
		return nodes
	}
	if err := json.Unmarshal(raw, &nodes); err != nil {
		return []documentsapi.DocumentCommentContentNode{}
	}
	return nodes
}

// authorizeDocumentScope asserts CapDocumentView (tenant-grade, ADR 0022) for the
// actor identified in the request. isSystemAdmin (ADR-0022-permitted admin shortcut
// via iam_user_roles) bypasses the capability gate when the caps checker is wired.
// Non-admin actors are gate-checked via RequireDocumentView (SeedTxIdentity +
// authz.Require(CapDocumentView, "tenant") inside a RW tx), mirroring ViewService.
func (h *Handler) authorizeDocumentScope(w http.ResponseWriter, r *http.Request, ctx context.Context, docID string) (tenantID string, userID string, ok bool) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, problem.CodeInternalUnknown.String()))
		return "", "", false
	}
	userID, err = userIDFromReq(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
		return "", "", false
	}
	admin, err := h.isSystemAdmin(ctx, userID, tenantID)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, problem.CodeInternalUnknown.String()))
		return "", "", false
	}
	if admin {
		return tenantID, userID, true
	}

	if err := h.svc.RequireDocumentView(ctx, tenantID, userID, docID); err != nil {
		status, msg := mapErr(err)
		problem.Respond(w, r, problem.New(status, msg, msg.String()))
		return "", "", false
	}
	return tenantID, userID, true
}

// withAdminCtx extends the caller-supplied ctx (the request's inherited
// context, passed explicitly so contextcheck can trace the chain across this
// call boundary) with the caller's admin auth info (userID + roles), used by
// authorizeDocumentScope's isSystemAdmin shortcut and by the handler's
// downstream service calls.
func withAdminCtx(ctx context.Context, r *http.Request) context.Context {
	// A3.3 class B (identity OPTIONAL by policy): this helper only ENRICHES a
	// context that already carries auth. It performs no permission decision and
	// writes nothing — the actor-required decision belongs to
	// authorizeDocumentScope, which fails closed on its own. Absence therefore
	// returns ctx unchanged rather than a blank actor; the presence result is
	// read explicitly so "" can never be forwarded to WithAuthContext.
	userID, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		return ctx
	}

	roles := iamdomain.RolesFromContext(ctx)
	if len(roles) == 0 {
		return ctx
	}
	return iamdomain.WithAuthContext(ctx, userID, roles)
}

func tenantIDFromReq(r *http.Request) (string, error) {
	return tenant.FromContext(r.Context())
}

// userIDFromReq resolves the authenticated actor for the documents handlers
// that attribute a read/mutation to a person — export, list/stats admin
// shortcut, force-release, document scope authorization. A3.3: it returns
// authn.ErrMissingActor rather than "" when the principal is absent, so a
// blank actor can no longer reach isSystemAdmin, RequireDocumentView or an
// export service call.
func userIDFromReq(r *http.Request) (string, error) {
	return authn.RequireUserID(r.Context())
}

// errorMapping is one entry of documentsErrorMappings: a predicate over err
// plus the HTTP status and problem.Code to use when it matches. Same
// ordered-dispatch-table shape as approval/http/errors.go's errorMapping
// (that module hit the same gocyclo ceiling first, for the identical
// authn.ErrMissingActor addition) — an ordered slice, evaluated top to
// bottom, first match wins, exactly reproducing the switch statement this
// table replaced.
type errorMapping struct {
	match  func(err error) bool
	status int
	code   problem.Code
}

// matchIs reports errors.Is(err, target) for any target in targets — an
// error matches the mapping if it matches ANY of them (mirrors a switch case
// with multiple comma-separated conditions sharing one body).
func matchIs(targets ...error) func(error) bool {
	return func(err error) bool {
		for _, target := range targets {
			if errors.Is(err, target) {
				return true
			}
		}
		return false
	}
}

// matchAs reports errors.As(err, &target) for a fresh zero-value T — the
// generic counterpart of a `var x T; errors.As(err, &x)` case clause.
func matchAs[T error]() func(error) bool {
	return func(err error) bool {
		var target T
		return errors.As(err, &target)
	}
}

// documentsErrorMappings is the ordered dispatch table mapErr evaluates top
// to bottom, taking the status/code of the first match. Entries are
// transcribed in their original switch-case order and must not be reordered.
var documentsErrorMappings = []errorMapping{
	// A3.3 / #108 review round 1: no authenticated principal in context.
	// Every documents route is session-required, so this is the platform's
	// 401 + auth.unauthenticated, not a documents dialect — the same case
	// approval/http (errors.go) and templates/delivery/http (errors.go)
	// already carry for the identical sentinel.
	{matchIs(authn.ErrMissingActor), http.StatusUnauthorized, problem.CodeAuthUnauthenticated},
	{matchIs(domain.ErrForbidden, domain.ErrDocumentNotOwner), http.StatusForbidden, problem.CodePermissionDenied},
	{matchIs(
		domain.ErrPendingNotFound,
		domain.ErrCheckpointNotFound,
		domain.ErrCommentNotFound,
		domain.ErrNotFound,
		controlleddocumentsdomain.ErrCDNotFound,
	), http.StatusNotFound, problem.CodeNotFoundResource},
	{matchIs(
		domain.ErrInvalidName,
		domain.ErrInvalidPageCount,
		application.ErrControlledDocumentRequired,
		approvalapp.ErrRevisionTitleRequired,
		domain.ErrCommentInvalid,
	), http.StatusBadRequest, problem.CodeRequestInvalid},
	// Both tier-1 (iamapp.ErrCapabilityDenied) and tier-2 (authz.ErrCapDenied)
	// denials are the same semantic — "you lack this capability" — so they must
	// surface the same problem code to clients regardless of which PDP tier denied.
	{matchIs(iamapp.ErrCapabilityDenied), http.StatusForbidden, problem.CodePermissionCapabilityDenied},
	{matchAs[authz.ErrCapDenied](), http.StatusForbidden, problem.CodePermissionCapabilityDenied},
	{matchIs(domain.ErrExpiredUpload), http.StatusGone, problem.CodeStateUploadExpired},
	// 409, not the pre-ADR-0089 410: the upload never happened, so 410 Gone
	// (which asserts the resource existed and was removed) is false. The
	// status is bound to state.upload_missing at registration (annex R-1).
	{matchIs(domain.ErrUploadMissing), http.StatusConflict, problem.CodeStateUploadMissing},
	{matchIs(domain.ErrUploadTooLarge), http.StatusRequestEntityTooLarge, problem.CodeRequestBodyTooLarge},
	{matchIs(domain.ErrContentHashMismatch), http.StatusUnprocessableEntity, problem.CodeRequestInvalid},
	{matchIs(
		domain.ErrSessionTaken,
		domain.ErrSessionInactive,
		domain.ErrSessionNotHolder,
		domain.ErrMisbound,
		controlleddocumentsdomain.ErrProfileHasNoDefaultTemplate,
	), http.StatusConflict, problem.CodeConflictGeneric},
	{matchIs(domain.ErrStaleBase), http.StatusConflict, problem.CodeConflictConcurrentModification},
	{matchIs(domain.ErrInvalidStateTransition, controlleddocumentsdomain.ErrCDNotActive), http.StatusConflict, problem.CodeStateTransitionInvalid},
	// Not a sentinel: a string-prefix predicate on err.Error(), same as the
	// original `case strings.HasPrefix(...)` clause. Kept as an inline
	// predicate func rather than pulled out of the table, mirroring
	// approval/http/errors.go's identical `json: unknown field` prefix entry.
	{func(err error) bool { return err != nil && strings.HasPrefix(err.Error(), "form_data_invalid") }, http.StatusUnprocessableEntity, problem.CodeRequestInvalid},
}

func mapErr(err error) (int, problem.Code) {
	if err == nil {
		return http.StatusOK, problem.Code{}
	}
	for _, m := range documentsErrorMappings {
		if m.match(err) {
			return m.status, m.code
		}
	}
	return http.StatusInternalServerError, problem.CodeInternalUnknown
}

func isKnownDocumentStatus(status string) bool {
	switch documentsapi.DocumentSummaryStatus(status) {
	case documentsapi.Approved,
		documentsapi.Draft,
		documentsapi.Obsolete,
		documentsapi.Published,
		documentsapi.Scheduled,
		documentsapi.Superseded,
		documentsapi.UnderReview:
		return true
	default:
		return status == string(domain.DocStatusArchived)
	}
}

// isJSONObject reports whether raw is a JSON object literal. The body was
// already syntax-validated by the decoder, so the leading token is decisive.
func isJSONObject(raw json.RawMessage) bool {
	return strings.HasPrefix(strings.TrimSpace(string(raw)), "{")
}

func isValidBoundedText(value string, max int) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed != "" && len(trimmed) <= max
}

func redactID(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "..."
	}
	return value[:8] + "..."
}
