// Package http is the HTTP delivery layer for the tokens bounded context.
// It implements tokensapi.ServerInterface and maps domain errors to RFC 9457
// problem+json responses.
package http

import (
	"context"
	"errors"
	"log/slog"
	"metaldocs/internal/platform/httprouter"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	openapi_types "github.com/oapi-codegen/runtime/types"

	tokensapi "metaldocs/internal/modules/tokens/api"
	"metaldocs/internal/modules/tokens/application"
	"metaldocs/internal/modules/tokens/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/strictjson"
	"metaldocs/internal/platform/tenant"
)

// Re-export command types as aliases so test code can use tokenshttp.CreateCommand
// without importing the application package directly.
type CreateCommand = application.CreateCommand
type UpdateCommand = application.UpdateCommand

// Domain-specific error codes emitted by this handler.
//
// ADR 0089 execution step 9 (annex §2.7, rows #155-#158).
//
// ALREADY_EXISTS and NOT_FOUND were REDECLARATIONS of catalog strings (annex
// collision C-4): this package declared its own constants carrying the exact
// wire strings the platform catalog already owns. The registry's duplicate guard
// makes that a build break, so they reference the catalog registrations — same
// wire strings, same statuses, one declaration site.
//
// The two OWN codes were bare lower_snake, outside every family. They are
// re-homed under validation. (the request parses; a supplied value fails a
// business rule), which is also their existing 422 — so no status changes.
var (
	CodeTokenAlreadyExists  = problem.CodeConflictAlreadyExists
	CodeTokenNotFound       = problem.CodeNotFoundResource
	CodeTokenImmutableField = problem.Register("tokens", "validation.field_immutable", 422)
	CodeTokenReservedName   = problem.Register("tokens", "validation.name_reserved", 422)
)

// TokenService is the port the handler depends on. It matches the application
// service surface so the concrete *application.Service satisfies it directly.
type TokenService interface {
	Create(ctx context.Context, cmd application.CreateCommand) (*domain.Entry, error)
	Get(ctx context.Context, tenantID, id string) (*domain.Entry, error)
	List(ctx context.Context, tenantID string) ([]domain.Entry, error)
	Update(ctx context.Context, cmd application.UpdateCommand) (*domain.Entry, error)
	Delete(ctx context.Context, tenantID, actorID, id string) error
}

// Handler implements tokensapi.ServerInterface.
type Handler struct {
	svc TokenService
}

// NewHandler constructs a Handler wired to the given service.
func NewHandler(svc TokenService) *Handler {
	if svc == nil {
		panic("tokens/http: svc is required")
	}
	return &Handler{svc: svc}
}

// RegisterRoutes mounts the handler onto mux under /api/v1.
func (h *Handler) RegisterRoutes(mux httprouter.Muxer) {
	tokensapi.HandlerWithOptions(h, tokensapi.StdHTTPServerOptions{
		BaseRouter: mux,
		BaseURL:    "/api/v1",
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestInvalid, err.Error())
		},
	})
}

// ---- ServerInterface implementation ----------------------------------------

// ListTokens handles GET /tokens.
func (h *Handler) ListTokens(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}

	items, err := h.svc.List(r.Context(), tenantID)
	if err != nil {
		slog.Error("tokens: list", "err", err)
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}

	dtos := make([]tokensapi.TokenDictionaryEntry, len(items))
	for i := range items {
		dtos[i] = toDTO(&items[i])
	}
	httpresponse.WriteJSON(w, http.StatusOK, tokensapi.ListTokenDictionaryEntriesResponse{Items: dtos})
}

// CreateToken handles POST /tokens.
func (h *Handler) CreateToken(w http.ResponseWriter, r *http.Request) {
	var body tokensapi.CreateTokenJSONRequestBody
	if err := strictjson.Decode(r, &body); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "invalid JSON payload")
		return
	}

	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}
	actorID, _ := authn.UserIDFromContext(r.Context())

	entry, err := h.svc.Create(r.Context(), application.CreateCommand{
		TenantID:    tenantID,
		ActorID:     actorID,
		Name:        body.Name,
		Value:       body.Value,
		Label:       body.Label,
		Description: body.Description,
	})
	if err != nil {
		h.writeTokenError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, toDTO(entry))
}

// GetToken handles GET /tokens/{id}.
func (h *Handler) GetToken(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}

	entry, err := h.svc.Get(r.Context(), tenantID, id.String())
	if err != nil {
		h.writeTokenError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, toDTO(entry))
}

// UpdateToken handles PUT /tokens/{id}.
func (h *Handler) UpdateToken(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	var body tokensapi.UpdateTokenJSONRequestBody
	if err := strictjson.Decode(r, &body); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "invalid JSON payload")
		return
	}

	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}
	actorID, _ := authn.UserIDFromContext(r.Context())

	entry, err := h.svc.Update(r.Context(), application.UpdateCommand{
		TenantID:    tenantID,
		ActorID:     actorID,
		ID:          id.String(),
		Name:        body.Name,
		Value:       body.Value,
		Label:       body.Label,
		Description: body.Description,
	})
	if err != nil {
		h.writeTokenError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, toDTO(entry))
}

// DeleteToken handles DELETE /tokens/{id}.
func (h *Handler) DeleteToken(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}
	actorID, _ := authn.UserIDFromContext(r.Context())

	if err := h.svc.Delete(r.Context(), tenantID, actorID, id.String()); err != nil {
		h.writeTokenError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- error mapping ---------------------------------------------------------

func (h *Handler) writeTokenError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError
	var valErr *domain.ValidationError
	switch {
	case errors.Is(err, domain.ErrNotFound):
		httpresponse.WriteError(w, http.StatusNotFound, CodeTokenNotFound, "token dictionary entry not found")
	case errors.Is(err, domain.ErrImmutableName):
		_ = problem.Write(w, problem.
			New(http.StatusUnprocessableEntity, CodeTokenImmutableField, "name is immutable after creation").
			WithFieldError("name", CodeTokenImmutableField, "name is immutable after creation"))
	case errors.Is(err, domain.ErrReservedName):
		_ = problem.Write(w, problem.
			New(http.StatusUnprocessableEntity, CodeTokenReservedName, "name is reserved by a native token").
			WithFieldError("name", CodeTokenReservedName, "name collides with a built-in token; choose another"))
	// annex R-6: both 422 sites below carried request.invalid, whose registered
	// default is 400 — a code/status contradiction. The generic 422 is
	// validation.failed (annex row #121), and the status now comes from the
	// registration via NewFor instead of being restated at the call site.
	case errors.As(err, &valErr):
		_ = problem.Write(w, problem.NewFor(problem.CodeValidationFailed, valErr.Error()))
	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		httpresponse.WriteError(w, http.StatusConflict, CodeTokenAlreadyExists, "token name already exists for this tenant")
	case errors.As(err, &pgErr) && pgErr.Code == "23514":
		_ = problem.Write(w, problem.NewFor(problem.CodeValidationFailed, "request violates data constraints"))
	default:
		slog.Error("tokens: handler error", "err", err)
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
	}
}

// ---- DTO mapping -----------------------------------------------------------

func toDTO(e *domain.Entry) tokensapi.TokenDictionaryEntry {
	id, _ := uuid.Parse(e.ID)
	return tokensapi.TokenDictionaryEntry{
		Id:          openapi_types.UUID(id),
		Name:        e.Name,
		Value:       e.Value,
		Label:       e.Label,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}
