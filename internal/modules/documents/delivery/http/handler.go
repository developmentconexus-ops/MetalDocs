package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"metaldocs/internal/modules/documents/application"
	approvalapp "metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/domain"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	registrydomain "metaldocs/internal/modules/registry/domain"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/ratelimit"
	"metaldocs/internal/platform/tenant"
)

const (
	roleAdmin          = "system_admin"
	roleTemplateAuthor = "template_author"
	roleDocumentFiller = "document_filler"
)

type Service interface {
	// Unused by HTTP handlers (no route). Kept for DuplicateDocument internal use.
	CreateDocument(ctx context.Context, cmd application.CreateDocumentInput) (*application.CreateDocumentResult, error)
	GetDocument(ctx context.Context, tenantID, id string) (*domain.Document, error)
	DuplicateDocument(ctx context.Context, tenantID, userID, docID string) (*application.CreateDocumentResult, error)
	RenameDocument(ctx context.Context, tenantID, userID, docID, newName string) error
	ListDocuments(ctx context.Context, tenantID string) ([]domain.Document, error)
	ListDocumentsForUser(ctx context.Context, tenantID, userID string) ([]domain.Document, error)
	ListDocumentsPaginated(ctx context.Context, tenantID, userID string, opts application.ListOptions) ([]*domain.Document, int64, error)
	DocumentStats(ctx context.Context, tenantID, userID string, opts application.ListOptions) (*application.DocumentStats, error)
	IsDocumentOwner(ctx context.Context, tenantID, docID, userID string) (bool, error)
	AcquireSession(ctx context.Context, tenantID, docID, userID string) (*domain.Session, bool, error)
	HeartbeatSession(ctx context.Context, sessionID, userID string) error
	ReleaseSession(ctx context.Context, tenantID, sessionID, userID, docID string) error
	ForceReleaseSession(ctx context.Context, tenantID, adminID, sessionID, docID string) error
	PresignAutosave(ctx context.Context, cmd application.PresignAutosaveCmd) (*application.PresignAutosaveResult, error)
	CommitAutosave(ctx context.Context, cmd application.CommitAutosaveCmd) (*application.CommitResult, error)
	CreateCheckpoint(ctx context.Context, tenantID, docID, actorID, label string) (*domain.Checkpoint, error)
	ListCheckpoints(ctx context.Context, tenantID, docID string) ([]domain.Checkpoint, error)
	ListRevisionHistory(ctx context.Context, tenantID, docID string) ([]domain.RevisionHistoryItem, error)
	RestoreCheckpoint(ctx context.Context, tenantID, docID, actorID string, versionNum int) (*application.RestoreResult, error)
	Finalize(ctx context.Context, tenantID, docID, actorID string) error
	Archive(ctx context.Context, tenantID, docID, actorID string) error
	SignedRevisionURL(ctx context.Context, tenantID, docID, revID string) (string, error)
	ListDocumentComments(ctx context.Context, tenantID, userID, documentID string) ([]domain.Comment, error)
	AddDocumentComment(ctx context.Context, tenantID, userID, authorDisplay, documentID string, in domain.CommentCreateInput) (*domain.Comment, error)
	UpdateDocumentComment(ctx context.Context, tenantID, userID, documentID string, libraryID int, in domain.CommentUpdateInput) (*domain.Comment, error)
	DeleteDocumentComment(ctx context.Context, tenantID, userID, documentID string, libraryID int) error
}

// approvalSubmitter is the subset of the approval submit service used by finalizeDocument.
type approvalSubmitter interface {
	SubmitRevisionForReview(ctx context.Context, db *sql.DB, req approvalapp.SubmitRequest) (approvalapp.SubmitResult, error)
}

type finalizeIdempotencyStore interface {
	CheckReplay(ctx context.Context, tenantID, actorID, key, payloadHash string) (*idempotency.Replay, error)
	RecordReplay(ctx context.Context, tenantID, actorID, key, payloadHash string, status int, body []byte) error
}

type Handler struct {
	svc           Service
	db            *sql.DB
	submitSvc     approvalSubmitter
	idempFinalize finalizeIdempotencyStore
}

var writeJSON = httpresponse.WriteJSON

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// NewHandlerWithSubmit constructs a Handler with direct DB access and approval
// submit service — required for the finalize→submit flow.
func NewHandlerWithSubmit(svc Service, db *sql.DB, submitSvc approvalSubmitter) *Handler {
	return NewHandlerWithSubmitAndFinalizeStore(svc, db, submitSvc, nil)
}

func NewHandlerWithSubmitAndFinalizeStore(svc Service, db *sql.DB, submitSvc approvalSubmitter, store finalizeIdempotencyStore) *Handler {
	h := &Handler{svc: svc, db: db, submitSvc: submitSvc, idempFinalize: store}
	if h.idempFinalize == nil && db != nil {
		h.idempFinalize = idempotency.New(db, "POST /api/v1/documents/{id}/finalize")
	}
	return h
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/documents", h.listDocuments)
	mux.HandleFunc("GET /api/v1/documents/stats", h.documentStats)

	mux.HandleFunc("GET /api/v1/documents/{id}", h.getDocument)
	mux.HandleFunc("PATCH /api/v1/documents/{id}", h.renameDocument)
	mux.HandleFunc("POST /api/v1/documents/{id}/finalize", h.finalizeDocument)
	mux.HandleFunc("POST /api/v1/documents/{id}/archive", h.archiveDocument)
	mux.HandleFunc("POST /api/v1/documents/{id}/duplicate", h.duplicateDocument)

	mux.HandleFunc("POST /api/v1/documents/{id}/session/acquire", h.acquireSession)
	mux.HandleFunc("POST /api/v1/documents/{id}/session/heartbeat", h.heartbeatSession)
	mux.HandleFunc("POST /api/v1/documents/{id}/session/release", h.releaseSession)
	mux.HandleFunc("POST /api/v1/documents/{id}/session/force-release", h.forceReleaseSession)

	mux.HandleFunc("POST /api/v1/documents/{id}/autosave/presign", h.presignAutosave)
	mux.HandleFunc("POST /api/v1/documents/{id}/autosave/commit", h.commitAutosave)

	mux.HandleFunc("GET /api/v1/documents/{id}/checkpoints", h.listCheckpoints)
	mux.HandleFunc("POST /api/v1/documents/{id}/checkpoints", h.createCheckpoint)
	mux.HandleFunc("POST /api/v1/documents/{id}/checkpoints/{version}/restore", h.restoreCheckpoint)
	mux.HandleFunc("GET /api/v1/documents/{id}/revision-history", h.listRevisionHistory)

	mux.HandleFunc("GET /api/v1/documents/{id}/revisions/{rid}/url", h.signedRevisionURL)
	mux.HandleFunc("GET /api/v1/documents/{id}/comments", h.listComments)
	mux.HandleFunc("POST /api/v1/documents/{id}/comments", h.createComment)
	mux.HandleFunc("PATCH /api/v1/documents/{id}/comments/{libraryID}", h.updateComment)
	mux.HandleFunc("DELETE /api/v1/documents/{id}/comments/{libraryID}", h.deleteComment)
}

func (h *Handler) RegisterRoutesWithRateLimit(mux *http.ServeMux, rl *ratelimit.Middleware, userFn func(*http.Request) string) {
	mux.HandleFunc("GET /api/v1/documents", h.listDocuments)
	mux.HandleFunc("GET /api/v1/documents/stats", h.documentStats)

	mux.HandleFunc("GET /api/v1/documents/{id}", h.getDocument)
	mux.HandleFunc("PATCH /api/v1/documents/{id}", h.renameDocument)
	mux.HandleFunc("POST /api/v1/documents/{id}/finalize", h.finalizeDocument)
	mux.HandleFunc("POST /api/v1/documents/{id}/archive", h.archiveDocument)
	mux.HandleFunc("POST /api/v1/documents/{id}/duplicate", h.duplicateDocument)

	mux.HandleFunc("POST /api/v1/documents/{id}/session/acquire", h.acquireSession)
	mux.HandleFunc("POST /api/v1/documents/{id}/session/heartbeat", h.heartbeatSession)
	mux.HandleFunc("POST /api/v1/documents/{id}/session/release", h.releaseSession)
	mux.HandleFunc("POST /api/v1/documents/{id}/session/force-release", h.forceReleaseSession)

	mux.Handle(
		"POST /api/v1/documents/{id}/autosave/presign",
		rl.Limit(ratelimit.RouteAutosavePresign, userFn, http.HandlerFunc(h.presignAutosave)),
	)
	mux.Handle(
		"POST /api/v1/documents/{id}/autosave/commit",
		rl.Limit(ratelimit.RouteAutosaveCommit, userFn, http.HandlerFunc(h.commitAutosave)),
	)

	mux.HandleFunc("GET /api/v1/documents/{id}/checkpoints", h.listCheckpoints)
	mux.HandleFunc("POST /api/v1/documents/{id}/checkpoints", h.createCheckpoint)
	mux.HandleFunc("POST /api/v1/documents/{id}/checkpoints/{version}/restore", h.restoreCheckpoint)
	mux.HandleFunc("GET /api/v1/documents/{id}/revision-history", h.listRevisionHistory)

	mux.HandleFunc("GET /api/v1/documents/{id}/revisions/{rid}/url", h.signedRevisionURL)
	mux.HandleFunc("GET /api/v1/documents/{id}/comments", h.listComments)
	mux.HandleFunc("POST /api/v1/documents/{id}/comments", h.createComment)
	mux.HandleFunc("PATCH /api/v1/documents/{id}/comments/{libraryID}", h.updateComment)
	mux.HandleFunc("DELETE /api/v1/documents/{id}/comments/{libraryID}", h.deleteComment)
}

func (h *Handler) listDocuments(w http.ResponseWriter, r *http.Request) {
	if !hasAnyRole(r, roleAdmin, roleDocumentFiller) {
		httpErr(w, http.StatusForbidden, "forbidden")
		return
	}

	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, "internal_error")
		return
	}
	callerUserID := userIDFromReq(r)
	isAdmin := hasRole(r, roleAdmin)
	opts, effectiveUserID, err := parseListOptions(r, callerUserID, isAdmin)
	if err != nil {
		httpErr(w, http.StatusBadRequest, err.Error())
		return
	}

	items, total, err := h.svc.ListDocumentsPaginated(r.Context(), tenantID, effectiveUserID, opts)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"items":    items,
		"page":     opts.Page,
		"pageSize": opts.PageSize,
		"total":    total,
	})
}

func (h *Handler) documentStats(w http.ResponseWriter, r *http.Request) {
	if !hasAnyRole(r, roleAdmin, roleDocumentFiller) {
		httpErr(w, http.StatusForbidden, "forbidden")
		return
	}

	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, "internal_error")
		return
	}
	callerUserID := userIDFromReq(r)
	isAdmin := hasRole(r, roleAdmin)
	opts, effectiveUserID, err := parseListOptions(r, callerUserID, isAdmin)
	if err != nil {
		httpErr(w, http.StatusBadRequest, err.Error())
		return
	}

	stats, err := h.svc.DocumentStats(r.Context(), tenantID, effectiveUserID, opts)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, stats)
}

func parseListOptions(r *http.Request, callerUserID string, isAdmin bool) (application.ListOptions, string, error) {
	query := r.URL.Query()
	opts := application.ListOptions{
		Page:     1,
		PageSize: 20,
	}

	if rawPage := strings.TrimSpace(query.Get("page")); rawPage != "" {
		page, err := strconv.Atoi(rawPage)
		if err != nil {
			return opts, "", errors.New("page must be a valid integer")
		}
		if page < 1 {
			return opts, "", errors.New("page must be >= 1")
		}
		opts.Page = page
	}

	if rawPageSize := strings.TrimSpace(query.Get("pageSize")); rawPageSize != "" {
		pageSize, err := strconv.Atoi(rawPageSize)
		if err != nil {
			return opts, "", errors.New("pageSize must be a valid integer")
		}
		if pageSize < 1 {
			return opts, "", errors.New("pageSize must be >= 1")
		}
		if pageSize > 50 {
			return opts, "", errors.New("pageSize must be <= 50")
		}
		opts.PageSize = pageSize
	}

	statusValues := query["status"]
	if len(statusValues) > 0 {
		statuses := make([]string, 0, len(statusValues))
		for _, raw := range statusValues {
			for _, split := range strings.Split(raw, ",") {
				s := strings.TrimSpace(split)
				if s != "" {
					statuses = append(statuses, s)
				}
			}
		}
		opts.Status = statuses
	}

	opts.AreaCode = strings.TrimSpace(query.Get("areaCode"))
	opts.ProfileCode = strings.TrimSpace(query.Get("profileCode"))
	opts.Q = strings.TrimSpace(query.Get("q"))

	includeArchived := strings.TrimSpace(query.Get("includeArchived"))
	if includeArchived != "" {
		v, err := strconv.ParseBool(includeArchived)
		if err != nil {
			return opts, "", errors.New("includeArchived must be a valid boolean")
		}
		opts.IncludeArchived = v
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
		httpErr(w, http.StatusInternalServerError, "internal_error")
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, resp)
}

type documentDetailResponse struct {
	ID                      string                `json:"ID"`
	TenantID                string                `json:"TenantID"`
	TemplateVersionID       string                `json:"TemplateVersionID"`
	Name                    string                `json:"Name"`
	Status                  domain.DocumentStatus `json:"Status"`
	FormDataJSON            json.RawMessage       `json:"FormDataJSON"`
	CurrentRevisionID       string                `json:"CurrentRevisionID"`
	RevisionVersion         int64                 `json:"RevisionVersion"`
	ActiveSessionID         string                `json:"ActiveSessionID"`
	ValuesFrozenAt          *time.Time            `json:"ValuesFrozenAt"`
	ArchivedAt              *time.Time            `json:"ArchivedAt"`
	CreatedAt               time.Time             `json:"CreatedAt"`
	UpdatedAt               time.Time             `json:"UpdatedAt"`
	CreatedBy               string                `json:"CreatedBy"`
	ControlledDocumentID    *string               `json:"ControlledDocumentID"`
	RevisionTitle           *string               `json:"RevisionTitle"`
	ProfileCodeSnapshot     *string               `json:"ProfileCodeSnapshot"`
	ProcessAreaCodeSnapshot *string               `json:"ProcessAreaCodeSnapshot"`
	Code                    string                `json:"Code"`
}

func toDocumentDetailResponse(doc domain.Document) (*documentDetailResponse, error) {
	formData := json.RawMessage(doc.FormDataJSON)
	if len(formData) == 0 {
		formData = json.RawMessage(`{}`)
	}
	if !json.Valid(formData) {
		return nil, fmt.Errorf("invalid document form_data_json for document %s", doc.ID)
	}

	return &documentDetailResponse{
		ID:                      doc.ID,
		TenantID:                doc.TenantID,
		TemplateVersionID:       doc.TemplateVersionID,
		Name:                    doc.Name,
		Status:                  doc.Status,
		FormDataJSON:            formData,
		CurrentRevisionID:       doc.CurrentRevisionID,
		RevisionVersion:         doc.RevisionVersion,
		ActiveSessionID:         doc.ActiveSessionID,
		ValuesFrozenAt:          doc.ValuesFrozenAt,
		ArchivedAt:              doc.ArchivedAt,
		CreatedAt:               doc.CreatedAt,
		UpdatedAt:               doc.UpdatedAt,
		CreatedBy:               doc.CreatedBy,
		ControlledDocumentID:    doc.ControlledDocumentID,
		RevisionTitle:           doc.RevisionTitle,
		ProfileCodeSnapshot:     doc.ProfileCodeSnapshot,
		ProcessAreaCodeSnapshot: doc.ProcessAreaCodeSnapshot,
		Code:                    doc.Code,
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
		httpErr(w, http.StatusBadRequest, "invalid_body")
		return
	}

	if err := h.svc.RenameDocument(r.Context(), tenantID, userID, docID, req.Name); err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}

	doc, err := h.svc.GetDocument(r.Context(), tenantID, docID)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, doc)
}

func (h *Handler) finalizeDocument(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		RevisionTitle string `json:"revisionTitle"`
	}

	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		httpErr(w, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED")
		return
	}
	if !idempotency.IsValidKey(idempotencyKey) {
		httpErr(w, http.StatusBadRequest, "IDEMPOTENCY_KEY_INVALID")
		return
	}

	payloadHash, err := idempotency.RequestHash(r)
	if err != nil {
		httpErr(w, http.StatusBadRequest, "invalid_body")
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		httpErr(w, http.StatusBadRequest, "invalid_body")
		return
	}
	revisionTitle := strings.TrimSpace(reqBody.RevisionTitle)

	if h.submitSvc == nil || h.db == nil {
		// Fallback: legacy status-only transition when submit service not wired.
		r = withAdminCtx(r)
		docID := r.PathValue("id")
		tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
		if !ok {
			return
		}
		if err := h.svc.Finalize(r.Context(), tenantID, docID, userID); err != nil {
			status, msg := mapErr(err)
			httpErr(w, status, msg)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
		return
	}

	tenantForReplay, err := tenantIDFromReq(r)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, "internal_error")
		return
	}
	actorForReplay := userIDFromReq(r)
	idempStore := h.idempFinalize
	if idempStore == nil && h.db != nil {
		idempStore = idempotency.New(h.db, "POST /api/v1/documents/{id}/finalize")
	}
	if idempStore != nil {
		replay, err := idempStore.CheckReplay(r.Context(), tenantForReplay, actorForReplay, idempotencyKey, payloadHash)
		if errors.Is(err, idempotency.ErrConflict) {
			httpErr(w, http.StatusUnprocessableEntity, "IDEMPOTENCY_KEY_REUSED")
			return
		}
		if err != nil {
			httpErr(w, http.StatusInternalServerError, "internal_error")
			return
		}
		if replay != nil {
			w.Header().Set("Idempotent-Replay", "true")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(replay.Status)
			_, _ = w.Write(replay.Body)
			return
		}
	}

	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, actorID, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	// Load document revision version and controlled document ID.
	var revisionVersion int64
	var revisionNumber int64
	var cdID sql.NullString
	if err := h.db.QueryRowContext(r.Context(),
		`SELECT revision_version, revision_number, controlled_document_id::text
		   FROM documents
		  WHERE id = $1 AND tenant_id = $2 AND status = 'draft'`,
		docID, tenantID,
	).Scan(&revisionVersion, &revisionNumber, &cdID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpErr(w, http.StatusConflict, "document not in draft state")
		} else {
			httpErr(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	// Load profile code from controlled document.
	var profileCode string
	if cdID.Valid && cdID.String != "" {
		if err := h.db.QueryRowContext(r.Context(),
			`SELECT profile_code FROM controlled_documents WHERE id = $1 AND tenant_id = $2`,
			cdID.String, tenantID,
		).Scan(&profileCode); err != nil && !errors.Is(err, sql.ErrNoRows) {
			httpErr(w, http.StatusInternalServerError, "internal error")
			return
		}
	}
	if profileCode == "" {
		http.Error(w, `{"error":"profile not found"}`, http.StatusBadRequest)
		return
	}

	// Find the most recent active approval route for this profile.
	// Column name is `active` (migration 0146), not `is_active`.
	var routeID string
	if err := h.db.QueryRowContext(r.Context(),
		`SELECT id FROM approval_routes
		  WHERE tenant_id    = $1
		    AND profile_code = $2
		    AND active       = true
		  ORDER BY version DESC
		  LIMIT 1`,
		tenantID, profileCode,
	).Scan(&routeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpErr(w, http.StatusConflict, "no active approval route for profile "+profileCode)
		} else {
			httpErr(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	// Retrieve the latest content hash from the most recent autosaved revision.
	var contentHash string
	_ = h.db.QueryRowContext(r.Context(),
		`SELECT COALESCE(content_hash, '') FROM document_revisions
		  WHERE document_id = $1
		  ORDER BY created_at DESC LIMIT 1`,
		docID,
	).Scan(&contentHash)

	result, err := h.submitSvc.SubmitRevisionForReview(r.Context(), h.db, approvalapp.SubmitRequest{
		TenantID:        tenantID,
		DocumentID:      docID,
		RouteID:         routeID,
		SubmittedBy:     actorID,
		RevisionTitle:   revisionTitle,
		ContentFormData: map[string]any{"_content_hash": contentHash},
		RevisionVersion: int(revisionVersion),
		RevisionNumber:  int(revisionNumber),
	})
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	respBody := map[string]string{"instanceId": result.InstanceID}
	if idempStore != nil {
		body, _ := json.Marshal(respBody)
		if err := idempStore.RecordReplay(r.Context(), tenantID, actorID, idempotencyKey, payloadHash, http.StatusCreated, body); err != nil {
			log.Printf("documents finalize idempotency record error: %v", err)
		}
	}
	httpresponse.WriteJSON(w, http.StatusCreated, respBody)
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
			http.Error(w, `{"error":"`+msg+`","detail":"`+err.Error()+`"}`, status)
			return
		}
		httpErr(w, status, msg)
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, map[string]string{
		"document_id":         res.DocumentID,
		"initial_revision_id": res.InitialRevisionID,
		"session_id":          res.SessionID,
	})
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
		httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
			"mode":       "readonly",
			"held_by":    sess.UserID,
			"held_until": sess.ExpiresAt,
		})
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, map[string]any{
		"mode":                 "writer",
		"session_id":           sess.ID,
		"expires_at":           sess.ExpiresAt,
		"last_ack_revision_id": sess.LastAcknowledgedRevisionID,
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
		httpErr(w, http.StatusBadRequest, "invalid_body")
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
		httpErr(w, http.StatusBadRequest, "invalid_body")
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
	if !hasRole(r, roleAdmin) {
		httpErr(w, http.StatusForbidden, "forbidden")
		return
	}
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, "internal_error")
		return
	}
	adminID := userIDFromReq(r)

	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpErr(w, http.StatusBadRequest, "invalid_body")
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
		httpErr(w, http.StatusBadRequest, "invalid_body")
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
	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"upload_url":        res.UploadURL,
		"pending_upload_id": res.PendingUploadID,
		"expires_at":        res.ExpiresAt,
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
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpErr(w, http.StatusBadRequest, "invalid_body")
		return
	}

	res, err := h.svc.CommitAutosave(r.Context(), application.CommitAutosaveCmd{
		TenantID:         tenantID,
		ActorUserID:      userID,
		DocumentID:       docID,
		SessionID:        req.SessionID,
		PendingUploadID:  req.PendingUploadID,
		FormDataSnapshot: req.FormDataSnapshot,
	})
	if err != nil {
		log.Printf("documents.commit_autosave failed: doc_id=%s tenant_id=%s actor_id=%s session_id=%s pending_upload_id=%s err=%v",
			docID, tenantID, userID, req.SessionID, req.PendingUploadID, err)
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"revision_id":       res.RevisionID,
		"revision_num":      res.RevisionNum,
		"idempotent_replay": res.AlreadyConsumed,
	})
}

func (h *Handler) listCheckpoints(w http.ResponseWriter, r *http.Request) {
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, _, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	items, err := h.svc.ListCheckpoints(r.Context(), tenantID, docID)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, items)
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

	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"items": toRevisionHistoryResponse(items),
	})
}

type revisionHistoryItemResponse struct {
	DocumentID     string    `json:"documentId"`
	RevisionNumber int64     `json:"revisionNumber"`
	RevisionTitle  string    `json:"revisionTitle"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	IsCurrent      bool      `json:"isCurrent"`
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
		httpErr(w, http.StatusBadRequest, "invalid_body")
		return
	}

	cp, err := h.svc.CreateCheckpoint(r.Context(), tenantID, docID, userID, req.Label)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, cp)
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
		httpErr(w, http.StatusBadRequest, "invalid_version")
		return
	}

	res, err := h.svc.RestoreCheckpoint(r.Context(), tenantID, docID, userID, versionNum)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"new_revision_id":               res.NewRevisionID,
		"new_revision_num":              res.NewRevisionNum,
		"source_checkpoint_version_num": versionNum,
		"idempotent":                    res.Idempotent,
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
	httpresponse.WriteJSON(w, http.StatusOK, map[string]string{"url": url})
}

func (h *Handler) listComments(w http.ResponseWriter, r *http.Request) {
	r = withAdminCtx(r)
	docID := r.PathValue("id")
	tenantID, userID, ok := h.authorizeDocumentScope(w, r, docID)
	if !ok {
		return
	}

	comments, err := h.svc.ListDocumentComments(r.Context(), tenantID, userID, docID)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	resp := make([]commentResponse, 0, len(comments))
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
		httpErr(w, http.StatusBadRequest, "invalid_body")
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

	libraryID, err := strconv.Atoi(r.PathValue("libraryID"))
	if err != nil {
		httpErr(w, http.StatusBadRequest, "invalid_library_comment_id")
		return
	}
	var req struct {
		Content *json.RawMessage `json:"content"`
		Done    *bool            `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpErr(w, http.StatusBadRequest, "invalid_body")
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

	libraryID, err := strconv.Atoi(r.PathValue("libraryID"))
	if err != nil {
		httpErr(w, http.StatusBadRequest, "invalid_library_comment_id")
		return
	}
	if err := h.svc.DeleteDocumentComment(r.Context(), tenantID, userID, docID, libraryID); err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type commentResponse struct {
	ID               string          `json:"id"`
	LibraryCommentID int             `json:"library_comment_id"`
	ParentLibraryID  *int            `json:"parent_library_id"`
	Author           string          `json:"author"`
	AuthorID         string          `json:"author_id"`
	Content          json.RawMessage `json:"content"`
	Done             bool            `json:"done"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
	ResolvedAt       *time.Time      `json:"resolved_at"`
}

func toCommentResponse(c domain.Comment) commentResponse {
	return commentResponse{
		ID:               c.ID.String(),
		LibraryCommentID: c.LibraryCommentID,
		ParentLibraryID:  c.ParentLibraryID,
		Author:           c.AuthorDisplay,
		AuthorID:         c.AuthorID,
		Content:          c.ContentJSON,
		Done:             c.ResolvedAt != nil,
		CreatedAt:        c.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        c.UpdatedAt.UTC().Format(time.RFC3339),
		ResolvedAt:       c.ResolvedAt,
	}
}

func (h *Handler) authorizeDocumentScope(w http.ResponseWriter, r *http.Request, docID string) (tenantID string, userID string, ok bool) {
	if !hasAnyRole(r, roleAdmin, roleDocumentFiller) {
		httpErr(w, http.StatusForbidden, "forbidden")
		return "", "", false
	}
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, "internal_error")
		return "", "", false
	}
	userID = userIDFromReq(r)
	if hasRole(r, roleAdmin) {
		return tenantID, userID, true
	}

	owner, err := h.svc.IsDocumentOwner(r.Context(), tenantID, docID, userID)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return "", "", false
	}
	if !owner {
		httpErr(w, http.StatusForbidden, "forbidden")
		return "", "", false
	}
	return tenantID, userID, true
}

func withAdminCtx(r *http.Request) *http.Request {
	userID := userIDFromReq(r)
	if userID == "" {
		return r
	}

	roles := rolesFromHeader(r.Header.Get("X-User-Roles"))
	if len(roles) == 0 {
		return r
	}
	ctxRoles := make([]iamdomain.Role, 0, len(roles))
	for _, role := range roles {
		ctxRoles = append(ctxRoles, iamdomain.Role(role))
	}
	ctx := iamdomain.WithAuthContext(r.Context(), userID, ctxRoles)
	return r.WithContext(ctx)
}

func hasAnyRole(r *http.Request, want ...string) bool {
	for _, w := range want {
		if hasRole(r, w) {
			return true
		}
	}
	return false
}

func hasRole(r *http.Request, want string) bool {
	for _, role := range rolesFromHeader(r.Header.Get("X-User-Roles")) {
		if role == want {
			return true
		}
	}
	for _, role := range iamdomain.RolesFromContext(r.Context()) {
		if string(role) == want {
			return true
		}
	}
	return false
}

func rolesFromHeader(header string) []string {
	if strings.TrimSpace(header) == "" {
		return nil
	}
	parts := strings.Split(header, ",")
	roles := make([]string, 0, len(parts))
	for _, part := range parts {
		role := strings.ToLower(strings.TrimSpace(part))
		if role != "" {
			roles = append(roles, role)
		}
	}
	return roles
}

func tenantIDFromReq(r *http.Request) (string, error) {
	return tenant.FromContext(r.Context())
}

func userIDFromReq(r *http.Request) string {
	return iamdomain.UserIDFromContext(r.Context())
}

func mapErr(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, ""
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrDocumentNotOwner):
		return http.StatusForbidden, "forbidden"
	case errors.Is(err, domain.ErrPendingNotFound):
		return http.StatusNotFound, "pending_not_found"
	case errors.Is(err, domain.ErrCheckpointNotFound):
		return http.StatusNotFound, "checkpoint_not_found"
	case errors.Is(err, domain.ErrCommentNotFound):
		return http.StatusNotFound, "comment_not_found"
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, domain.ErrInvalidName):
		return http.StatusBadRequest, "invalid_name"
	case errors.Is(err, application.ErrControlledDocumentRequired):
		return http.StatusBadRequest, "controlled_document_required"
	case errors.Is(err, approvalapp.ErrRevisionTitleRequired):
		return http.StatusBadRequest, "revision_title_required"
	case errors.Is(err, domain.ErrCommentInvalid):
		return http.StatusBadRequest, "comment_invalid"
	case errors.Is(err, iamapp.ErrCapabilityDenied):
		return http.StatusForbidden, "forbidden"
	case errors.Is(err, domain.ErrExpiredUpload):
		return http.StatusGone, "expired_upload"
	case errors.Is(err, domain.ErrUploadMissing):
		return http.StatusGone, "upload_missing"
	case errors.Is(err, domain.ErrContentHashMismatch):
		return http.StatusUnprocessableEntity, "content_hash_mismatch"
	case errors.Is(err, domain.ErrSessionTaken):
		return http.StatusConflict, "session_taken"
	case errors.Is(err, domain.ErrSessionInactive):
		return http.StatusConflict, "session_inactive"
	case errors.Is(err, domain.ErrSessionNotHolder):
		return http.StatusConflict, "session_not_holder"
	case errors.Is(err, domain.ErrStaleBase):
		return http.StatusConflict, "stale_base"
	case errors.Is(err, domain.ErrMisbound):
		return http.StatusConflict, "misbound"
	case errors.Is(err, domain.ErrInvalidStateTransition):
		return http.StatusConflict, "invalid_state_transition"
	case errors.Is(err, registrydomain.ErrCDNotFound):
		return http.StatusNotFound, "controlled_document_not_found"
	case errors.Is(err, registrydomain.ErrCDNotActive):
		return http.StatusConflict, "controlled_document_not_active"
	case errors.Is(err, registrydomain.ErrProfileHasNoDefaultTemplate):
		return http.StatusConflict, "profile_has_no_default_template"
	case strings.HasPrefix(err.Error(), "form_data_invalid"):
		return http.StatusUnprocessableEntity, "form_data_invalid"
	default:
		return http.StatusInternalServerError, "internal_error"
	}
}

func httpErr(w http.ResponseWriter, code int, msg string) {
	_ = problem.Write(w, problem.New(code, msg, msg))
}
