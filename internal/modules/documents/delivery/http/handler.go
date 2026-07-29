package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	documentsapi "metaldocs/internal/modules/documents/api"
	"metaldocs/internal/modules/documents/application"
	approvalapp "metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/documents/domain"
	iamapp "metaldocs/internal/modules/iam/application"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
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

type Handler struct {
	svc             Service
	db              *sql.DB
	caps            application.CapabilityChecker
	export          *ExportHandler
	fillIn          *FillInHandler
	placeholderOpts *PlaceholderOptionsHandler
	view            *ViewHandler
	reconstruct     *ReconstructHandler
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

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

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

func (h *Handler) RegisterRoutes(mux *http.ServeMux) { h.registerRoutes(mux, nil, nil) }

func (h *Handler) RegisterRoutesWithRateLimit(mux *http.ServeMux, rl *ratelimit.Middleware, userFn func(*http.Request) string) {
	h.registerRoutes(mux, rl, userFn)
}

func (h *Handler) registerRoutes(mux *http.ServeMux, rl *ratelimit.Middleware, userFn func(*http.Request) string) {
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
		BaseURL: "/api/v1",
		Middlewares: []documentsapi.MiddlewareFunc{
			middleware,
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			_ = problem.Write(w, problem.New(http.StatusBadRequest, problem.CodeValidationError, err.Error()))
		},
	})
}

func (h *Handler) listDocuments(w http.ResponseWriter, r *http.Request, params documentsapi.ListDocumentsParams) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, problem.CodeInternalError)
		return
	}
	callerUserID := userIDFromReq(r)
	isAdmin, err := h.isSystemAdmin(r.Context(), callerUserID, tenantID)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, problem.CodeInternalError)
		return
	}
	opts, effectiveUserID, err := listOptionsFromParams(r, params, callerUserID, isAdmin)
	if err != nil {
		slog.Warn("documents listDocuments invalid query params", "err", err)
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}

	items, total, hasMore, err := h.svc.ListDocumentsPaginated(r.Context(), tenantID, effectiveUserID, opts)
	if err != nil {
		if errors.Is(err, pagination.ErrInvalidCursor) {
			httpErr(w, http.StatusBadRequest, problem.CodeInvalidCursor)
			return
		}
		status, msg := mapErr(err)
		httpErr(w, status, msg)
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
			httpErr(w, http.StatusInternalServerError, problem.CodeInternalError)
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
		httpErr(w, http.StatusInternalServerError, problem.CodeInternalError)
		return
	}
	callerUserID := userIDFromReq(r)
	isAdmin, err := h.isSystemAdmin(r.Context(), callerUserID, tenantID)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, problem.CodeInternalError)
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
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}

	stats, err := h.svc.DocumentStats(r.Context(), tenantID, effectiveUserID, opts)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
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
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, _, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	doc, err := h.svc.GetDocument(r.Context(), tenantID, docID)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	resp, err := toDocumentDetailResponse(*doc)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, problem.CodeInternalError)
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
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}
	if !isValidBoundedText(req.Name, 255) {
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}

	if err := h.svc.RenameDocument(r.Context(), tenantID, userID, docID, req.Name); err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}

	// F1.2 / ADR 0035 — renameDocument returns 200 OK with empty body. OpenAPI declares no
	// response schema. Consumer FE adapter is Promise<void>. No domain.Document leak.
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) archiveDocument(w http.ResponseWriter, r *http.Request) {
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	if err := h.svc.Archive(r.Context(), tenantID, docID, userID); err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) duplicateDocument(w http.ResponseWriter, r *http.Request) {
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	res, err := h.svc.DuplicateDocument(r.Context(), tenantID, userID, docID)
	if err != nil {
		status, msg := mapErr(err)
		if status == http.StatusInternalServerError {
			slog.Error("documents duplicate failed", "doc_id", docID, "tenant_id", tenantID, "actor_id", userID, "err", err)
			httpErr(w, status, msg)
			return
		}
		httpErr(w, status, msg)
		return
	}
	docUUID, ridUUID, sidUUID, parseErr := parseCreateResultUUIDs(res.DocumentID, res.InitialRevisionID, res.SessionID)
	if parseErr != nil {
		slog.Error("documents duplicate produced unparseable id", "doc_id", docID, "tenant_id", tenantID, "actor_id", userID, "err", parseErr)
		httpErr(w, http.StatusInternalServerError, problem.CodeInternalError)
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
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	sess, readonly, err := h.svc.AcquireSession(r.Context(), tenantID, docID, userID)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
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
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}
	lastAckID, err := uuid.Parse(sess.LastAcknowledgedRevisionID)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
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
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	_, userID, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}

	if err := h.svc.HeartbeatSession(r.Context(), req.SessionID, userID); err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) releaseSession(w http.ResponseWriter, r *http.Request) {
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}

	if err := h.svc.ReleaseSession(r.Context(), tenantID, req.SessionID, userID, docID); err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) forceReleaseSession(w http.ResponseWriter, r *http.Request) {
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, problem.CodeInternalError)
		return
	}
	adminID := userIDFromReq(r)

	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}

	if err := h.svc.ForceReleaseSession(r.Context(), tenantID, adminID, req.SessionID, docID); err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) presignAutosave(w http.ResponseWriter, r *http.Request) {
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	var req struct {
		SessionID      string `json:"session_id"`
		BaseRevisionID string `json:"base_revision_id"`
		ContentHash    string `json:"content_hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}

	res, err := h.svc.PresignAutosave(r.Context(), application.PresignAutosaveCmd{
		TenantID:       tenantID,
		ActorUserID:    userID,
		DocumentID:     docID,
		SessionID:      req.SessionID,
		BaseRevisionID: req.BaseRevisionID,
		ContentHash:    req.ContentHash,
	})
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	pendingID, err := uuid.Parse(res.PendingUploadID)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, documentsapi.DocumentAutosavePresignResponse{
		UploadUrl:       res.UploadURL,
		PendingUploadId: pendingID,
		ExpiresAt:       res.ExpiresAt,
	})
}

func (h *Handler) commitAutosave(w http.ResponseWriter, r *http.Request) {
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
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
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}
	// form_data_snapshot is optional — absent means "this autosave carries no
	// form-data change" and the write path preserves the stored form data. When
	// it IS present the contract types it as an object, so reject null/scalars/
	// arrays here: persisting one would make every later read of the document
	// fail to decode form_data_json into the contract's object type (one bad
	// write poisoning a read path with 500s).
	if len(req.FormDataSnapshot) > 0 && !isJSONObject(req.FormDataSnapshot) {
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}

	res, err := h.svc.CommitAutosave(r.Context(), application.CommitAutosaveCmd{
		TenantID:         tenantID,
		ActorUserID:      userID,
		DocumentID:       docID,
		SessionID:        req.SessionID,
		PendingUploadID:  req.PendingUploadID,
		FormDataSnapshot: req.FormDataSnapshot,
		PageCount:        req.PageCount,
	})
	if err != nil {
		slog.Error("documents.commit_autosave failed", "doc_id", docID, "tenant_id", tenantID, "actor_id", userID, "session_id", redactID(req.SessionID), "pending_upload_id", redactID(req.PendingUploadID), "err", err)
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	commitRevID, err := uuid.Parse(res.RevisionID)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
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
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	items, err := h.svc.ListCheckpoints(r.Context(), tenantID, userID, docID)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	resp, err := toAPICheckpoints(items)
	if err != nil {
		slog.Error("documents.list_checkpoints malformed uuid", "doc_id", docID, "tenant_id", tenantID, "err", err)
		httpErr(w, http.StatusInternalServerError, problem.CodeInternalError)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) listRevisionHistory(w http.ResponseWriter, r *http.Request) {
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, _, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	items, err := h.svc.ListRevisionHistory(r.Context(), tenantID, docID)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}

	histItems, err := toAPIRevisionHistoryItems(items)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, problem.CodeInternalError)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, documentsapi.DocumentRevisionHistoryResponse{
		Items: histItems,
	})
}

type revisionHistoryItemResponse struct {
	DocumentID     string    `json:"document_id"`
	RevisionNumber int64     `json:"revision_number"`
	RevisionTitle  string    `json:"revision_title"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	IsCurrent      bool      `json:"is_current"`
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

func toRevisionHistoryResponse(items []domain.RevisionHistoryItem) []revisionHistoryItemResponse {
	out := make([]revisionHistoryItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, revisionHistoryItemResponse{
			DocumentID:     item.DocumentID,
			RevisionNumber: item.RevisionNumber,
			RevisionTitle:  item.RevisionTitle,
			Status:         string(item.Status),
			CreatedAt:      item.CreatedAt,
			IsCurrent:      item.IsCurrent,
		})
	}
	return out
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
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	var req struct {
		Label string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}
	if !isValidBoundedText(req.Label, 255) {
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}

	cp, err := h.svc.CreateCheckpoint(r.Context(), tenantID, docID, userID, req.Label)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	resp, err := toAPICheckpoint(*cp)
	if err != nil {
		slog.Error("documents.create_checkpoint malformed uuid", "doc_id", docID, "tenant_id", tenantID, "err", err)
		httpErr(w, http.StatusInternalServerError, problem.CodeInternalError)
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) restoreCheckpoint(w http.ResponseWriter, r *http.Request) {
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	versionNum, err := strconv.Atoi(r.PathValue("version"))
	if err != nil {
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}

	res, err := h.svc.RestoreCheckpoint(r.Context(), tenantID, docID, userID, versionNum)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	newRevID, err := uuid.Parse(res.NewRevisionID)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
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
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, _, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	url, err := h.svc.SignedRevisionURL(r.Context(), tenantID, docID, r.PathValue("rid"))
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, documentsapi.RevisionUrlResponse{Url: url})
}

func (h *Handler) listComments(w http.ResponseWriter, r *http.Request) {
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, _, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	comments, err := h.svc.ListDocumentComments(r.Context(), tenantID, docID)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	resp := make([]documentsapi.DocumentCommentResponse, 0, len(comments))
	for i := range comments {
		resp = append(resp, toCommentResponse(comments[i]))
	}
	httpresponse.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) createComment(w http.ResponseWriter, r *http.Request) {
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
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
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}

	comment, err := h.svc.AddDocumentComment(r.Context(), tenantID, userID, req.AuthorDisplay, docID, domain.CommentCreateInput{
		LibraryCommentID: req.LibraryCommentID,
		ParentLibraryID:  req.ParentLibraryID,
		AuthorDisplay:    req.AuthorDisplay,
		ContentJSON:      req.Content,
	})
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, toCommentResponse(*comment))
}

func (h *Handler) updateComment(w http.ResponseWriter, r *http.Request) {
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	libraryID, err := strconv.Atoi(r.PathValue("library_id"))
	if err != nil {
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}
	var req struct {
		Content *json.RawMessage `json:"content"`
		Done    *bool            `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}

	comment, err := h.svc.UpdateDocumentComment(r.Context(), tenantID, userID, docID, libraryID, domain.CommentUpdateInput{
		ContentJSON: req.Content,
		Done:        req.Done,
	})
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, toCommentResponse(*comment))
}

func (h *Handler) deleteComment(w http.ResponseWriter, r *http.Request) {
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	libraryID, err := strconv.Atoi(r.PathValue("library_id"))
	if err != nil {
		httpErr(w, http.StatusBadRequest, problem.CodeValidationError)
		return
	}
	if err := h.svc.DeleteDocumentComment(r.Context(), tenantID, userID, docID, libraryID); err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
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
func (h *Handler) authorizeDocumentScope(w http.ResponseWriter, r *http.Request, docID string) (tenantID string, userID string, ok bool) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, problem.CodeInternalError)
		return "", "", false
	}
	userID = userIDFromReq(r)
	admin, err := h.isSystemAdmin(r.Context(), userID, tenantID)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, problem.CodeInternalError)
		return "", "", false
	}
	if admin {
		return tenantID, userID, true
	}

	if err := h.svc.RequireDocumentView(r.Context(), tenantID, userID, docID); err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return "", "", false
	}
	return tenantID, userID, true
}

func withAdminCtx(r *http.Request) *http.Request {
	userID := userIDFromReq(r)
	if userID == "" {
		return r
	}

	roles := iamdomain.RolesFromContext(r.Context())
	if len(roles) == 0 {
		return r
	}
	ctx := iamdomain.WithAuthContext(r.Context(), userID, roles)
	return r.WithContext(ctx)
}

func tenantIDFromReq(r *http.Request) (string, error) {
	return tenant.FromContext(r.Context())
}

func userIDFromReq(r *http.Request) string {
	return iamdomain.UserIDFromContext(r.Context())
}

func mapErr(err error) (int, problem.Code) {
	var capDenied authz.ErrCapDenied
	switch {
	case err == nil:
		return http.StatusOK, ""
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrDocumentNotOwner):
		return http.StatusForbidden, problem.CodeAuthForbidden
	case errors.Is(err, domain.ErrPendingNotFound),
		errors.Is(err, domain.ErrCheckpointNotFound),
		errors.Is(err, domain.ErrCommentNotFound),
		errors.Is(err, domain.ErrNotFound),
		errors.Is(err, controlleddocumentsdomain.ErrCDNotFound):
		return http.StatusNotFound, problem.CodeNotFound
	case errors.Is(err, domain.ErrInvalidName),
		errors.Is(err, domain.ErrInvalidPageCount),
		errors.Is(err, application.ErrControlledDocumentRequired),
		errors.Is(err, approvalapp.ErrRevisionTitleRequired),
		errors.Is(err, domain.ErrCommentInvalid):
		return http.StatusBadRequest, problem.CodeValidationError
	// Both tier-1 (iamapp.ErrCapabilityDenied) and tier-2 (authz.ErrCapDenied)
	// denials are the same semantic — "you lack this capability" — so they must
	// surface the same problem code to clients regardless of which PDP tier denied.
	case errors.Is(err, iamapp.ErrCapabilityDenied):
		return http.StatusForbidden, problem.CodeForbiddenCapability
	case errors.As(err, &capDenied):
		return http.StatusForbidden, problem.CodeForbiddenCapability
	case errors.Is(err, domain.ErrExpiredUpload):
		return http.StatusGone, problem.CodeUploadExpired
	case errors.Is(err, domain.ErrUploadMissing):
		return http.StatusGone, problem.CodeUploadMissing
	case errors.Is(err, domain.ErrUploadTooLarge):
		return http.StatusRequestEntityTooLarge, problem.CodeRequestBodyTooLarge
	case errors.Is(err, domain.ErrContentHashMismatch):
		return http.StatusUnprocessableEntity, problem.CodeValidationError
	case errors.Is(err, domain.ErrSessionTaken),
		errors.Is(err, domain.ErrSessionInactive),
		errors.Is(err, domain.ErrSessionNotHolder),
		errors.Is(err, domain.ErrMisbound),
		errors.Is(err, controlleddocumentsdomain.ErrProfileHasNoDefaultTemplate):
		return http.StatusConflict, problem.CodeConflict
	case errors.Is(err, domain.ErrStaleBase):
		return http.StatusConflict, problem.CodeConcurrentModification
	case errors.Is(err, domain.ErrInvalidStateTransition),
		errors.Is(err, controlleddocumentsdomain.ErrCDNotActive):
		return http.StatusConflict, problem.CodeStateTransitionInvalid
	case strings.HasPrefix(err.Error(), "form_data_invalid"):
		return http.StatusUnprocessableEntity, problem.CodeValidationError
	default:
		return http.StatusInternalServerError, problem.CodeInternalError
	}
}

func httpErr(w http.ResponseWriter, status int, code problem.Code) {
	_ = problem.Write(w, problem.New(status, code, string(code)))
}

// httpErrDetail writes a canonical-code problem carrying runtime context in the
// RFC 9457 detail field (never in the code field).
func httpErrDetail(w http.ResponseWriter, status int, code problem.Code, detail string) {
	_ = problem.Write(w, problem.New(status, code, string(code)).WithDetail(detail))
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
