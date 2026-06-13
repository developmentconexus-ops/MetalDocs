package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	v2domain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesdomain "metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

type FillInService interface {
	SetPlaceholderValue(ctx context.Context, tenantID, actorID, revisionID, placeholderID, value string) error
	GetPlaceholderValues(ctx context.Context, tenantID, docID string) ([]repository.PlaceholderValue, error)
	GetFillInSchema(ctx context.Context, tenantID, docID string) ([]templatesdomain.Placeholder, error)
}

type FillInHandler struct {
	service FillInService
}

func NewFillInHandler(service FillInService) *FillInHandler {
	return &FillInHandler{service: service}
}

func (h *FillInHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/documents/{id}/fill-in-schema", h.GetFillInSchema)
	mux.HandleFunc("GET /api/v1/documents/{id}/placeholders", h.ListPlaceholderValues)
	mux.HandleFunc("PUT /api/v1/documents/{id}/placeholders/{pid}", h.PutPlaceholderValue)
}

func (h *FillInHandler) GetFillInSchema(w http.ResponseWriter, r *http.Request) {
	tid, err := tenantID(r)
	if err != nil {
		writeFillInError(w, requestID(r), err)
		return
	}
	docID := r.PathValue("id")
	phs, err := h.service.GetFillInSchema(r.Context(), tid, docID)
	if err != nil {
		writeFillInError(w, requestID(r), err)
		return
	}
	if phs == nil {
		phs = []templatesdomain.Placeholder{}
	}
	writeFillInJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"placeholder_schema": phs,
		},
	})
}

func (h *FillInHandler) ListPlaceholderValues(w http.ResponseWriter, r *http.Request) {
	docID := r.PathValue("id")
	tid, err := tenantID(r)
	if err != nil {
		writeFillInError(w, requestID(r), err)
		return
	}

	vals, err := h.service.GetPlaceholderValues(r.Context(), tid, docID)
	if err != nil {
		writeFillInError(w, requestID(r), err)
		return
	}
	type out struct {
		PlaceholderID string  `json:"placeholder_id"`
		ValueText     *string `json:"value_text"`
		Source        string  `json:"source"`
	}
	res := make([]out, len(vals))
	for i, v := range vals {
		res[i] = out{PlaceholderID: v.PlaceholderID, ValueText: v.ValueText, Source: v.Source}
	}
	writeFillInJSON(w, http.StatusOK, res)
}

func (h *FillInHandler) PutPlaceholderValue(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Value string `json:"value"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeFillInError(w, requestID(r), err)
		return
	}
	tenantID, err := tenantID(r)
	if err != nil {
		writeFillInError(w, requestID(r), err)
		return
	}

	err = h.service.SetPlaceholderValue(r.Context(),
		tenantID,
		actorID(r),
		r.PathValue("id"),
		r.PathValue("pid"),
		body.Value,
	)
	if err != nil {
		writeFillInError(w, requestID(r), err)
		return
	}

	writeFillInJSON(w, http.StatusOK, map[string]any{
		"placeholder_id": r.PathValue("pid"),
		"updated_at":     time.Now().UTC().Format(time.RFC3339),
	})
}

var ErrBadContentType = errors.New("content-type must be application/json")

// mapFillInError maps a service error to its RFC 9457 (status, code) pair. The
// codes are the module's dot-notation taxonomy (see internal/platform/problem).
func mapFillInError(err error) (int, problem.Code) {
	switch {
	case errors.As(err, &authz.ErrCapDenied{}):
		return http.StatusForbidden, "authz.capability_denied"
	case errors.As(err, &notChoicePlaceholderError{}):
		return http.StatusBadRequest, "not_a_choice_placeholder"
	case errors.Is(err, v2domain.ErrNotFound):
		return http.StatusNotFound, "not_found.revision"
	case errors.Is(err, v2domain.ErrInvalidStateTransition):
		return http.StatusConflict, "state.revision_not_draft"
	case errors.Is(err, v2domain.ErrValidationFailed):
		return http.StatusUnprocessableEntity, "validation.failed"
	case errors.Is(err, io.EOF):
		return http.StatusBadRequest, "validation.empty_body"
	case errors.Is(err, ErrBadContentType):
		return http.StatusUnsupportedMediaType, "validation.bad_content_type"
	case looksLikeDecodeError(err):
		return http.StatusBadRequest, "validation.json_decode"
	default:
		return http.StatusInternalServerError, "internal.unknown"
	}
}

// writeFillInError emits the unified RFC 9457 application/problem+json error
// shape (AD-2). The request id, when present, rides in the Problem `instance`.
func writeFillInError(w http.ResponseWriter, reqID string, err error) {
	status, code := mapFillInError(err)
	prob := problem.New(status, code, errorMessage(err, status))
	if reqID != "" {
		prob = prob.WithInstance(reqID)
	}
	_ = problem.Write(w, prob)
}

func writeFillInJSON(w http.ResponseWriter, status int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		_ = problem.Write(w, problem.New(http.StatusInternalServerError, "internal.unknown", "internal error"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

func decodeJSON(r *http.Request, out any) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return ErrBadContentType
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(out)
}

func errorMessage(err error, status int) string {
	if status >= http.StatusInternalServerError {
		return "internal error"
	}
	return err.Error()
}

func looksLikeDecodeError(err error) bool {
	if err == nil {
		return false
	}
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	return errors.As(err, &syntaxErr) || errors.As(err, &typeErr) || errors.Is(err, io.ErrUnexpectedEOF)
}

func requestID(r *http.Request) string {
	if id := strings.TrimSpace(r.Header.Get("X-Request-ID")); id != "" {
		return id
	}
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

func tenantID(r *http.Request) (string, error) {
	return tenant.FromContext(r.Context())
}

func actorID(r *http.Request) string {
	return iamdomain.UserIDFromContext(r.Context())
}
