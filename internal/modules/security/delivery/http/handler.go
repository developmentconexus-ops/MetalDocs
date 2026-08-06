// Package httpdelivery exposes the Sessions & Security tab read endpoints
// (PR-7) over HTTP. Every endpoint depends on tenant.FromContext — the
// middleware in apps/api/cmd/metaldocs-api/permissions.go already enforces
// auth + CapUserView, so the handler can treat the context as trusted.
//
// Routes are mounted through the generated securityapi.ServerInterface router
// (HandlerWithOptions) rather than bare mux.HandleFunc — see RegisterRoutes.
// Handler satisfies securityapi.ServerInterface directly (not the strict
// variant): the three operations already write typed JSON/problem+json
// bodies straight to http.ResponseWriter via the existing writeJSON/
// writeProblem helpers, which the strict (ctx, RequestObject) ->
// (ResponseObject, error) shape would force to be re-expressed as
// XxxResponseObject wrappers for no behavioral gain. Mirrors the audit
// module's precedent (internal/modules/audit/delivery/http/handler.go) and
// Task 6's auth migration.
package httpdelivery

import (
	"log/slog"
	"metaldocs/internal/platform/httprouter"
	"net/http"
	"time"

	securityapi "metaldocs/internal/modules/security/api"
	securityapp "metaldocs/internal/modules/security/application"
	securitydomain "metaldocs/internal/modules/security/domain"
	"metaldocs/internal/platform/apibase"
	httpresponse "metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// --- typed response structs (F6.4) ---

type mfaCoverageByRoleItem struct {
	Role       string  `json:"role"`
	Total      int     `json:"total"`
	MfaEnabled int     `json:"mfa_enabled"`
	Pct        float32 `json:"pct"`
}

type mfaCoverageResponse struct {
	TotalUsers    int                     `json:"total_users"`
	MfaEnabled    int                     `json:"mfa_enabled"`
	MfaEnabledPct float32                 `json:"mfa_enabled_pct"`
	ByRole        []mfaCoverageByRoleItem `json:"by_role"`
}

type lockoutItem struct {
	UserID         string  `json:"user_id"`
	DisplayName    string  `json:"display_name"`
	FailedAttempts int     `json:"failed_attempts"`
	LockedUntil    *string `json:"locked_until,omitempty"`
	LastFailedAt   *string `json:"last_failed_at,omitempty"`
	LastFailedIP   string  `json:"last_failed_ip,omitempty"`
}

type lockoutsResponse struct {
	Items []lockoutItem `json:"items"`
}

type signalItem struct {
	SignalID   string         `json:"signal_id"`
	Kind       string         `json:"kind"`
	Severity   string         `json:"severity"`
	Summary    string         `json:"summary"`
	DetectedAt string         `json:"detected_at"`
	Evidence   map[string]any `json:"evidence,omitempty"`
}

type signalsResponse struct {
	Items []signalItem `json:"items"`
}

var writeJSON = httpresponse.WriteJSON

type Handler struct {
	service *securityapp.Service
}

func NewHandler(service *securityapp.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux httprouter.Muxer) {
	securityapi.HandlerWithOptions(h, securityapi.StdHTTPServerOptions{
		BaseURL:    apibase.BaseURL,
		BaseRouter: mux,
	})
}

// ─── securityapi.ServerInterface adapter ───────────────────────────────────

// GetMfaCoverage adapts GET /security/mfa-coverage to the existing handler.
func (h *Handler) GetMfaCoverage(w http.ResponseWriter, r *http.Request) {
	h.handleMfaCoverage(w, r)
}

// ListLockouts adapts GET /security/lockouts to the existing handler.
func (h *Handler) ListLockouts(w http.ResponseWriter, r *http.Request) {
	h.handleLockouts(w, r)
}

// ListSecuritySignals adapts GET /security/signals to the existing handler.
func (h *Handler) ListSecuritySignals(w http.ResponseWriter, r *http.Request) {
	h.handleSignals(w, r)
}

func (h *Handler) handleMfaCoverage(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := h.requireTenant(w, r)
	if !ok {
		return
	}
	if h.service == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "Security service is not configured"))
		return
	}
	coverage, err := h.service.MfaCoverage(r.Context(), tenantID)
	if err != nil {
		slog.Error("security: mfa coverage failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to load MFA coverage"))
		return
	}
	writeJSON(w, http.StatusOK, mfaCoverageToJSON(coverage))
}

func (h *Handler) handleLockouts(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := h.requireTenant(w, r)
	if !ok {
		return
	}
	if h.service == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "Security service is not configured"))
		return
	}
	items, err := h.service.ListLockouts(r.Context(), tenantID)
	if err != nil {
		slog.Error("security: lockouts failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to list lockouts"))
		return
	}
	out := make([]lockoutItem, 0, len(items))
	for _, l := range items {
		row := lockoutItem{
			UserID:         l.UserID,
			DisplayName:    l.DisplayName,
			FailedAttempts: l.FailedAttempts,
			LastFailedIP:   l.LastFailedIP,
		}
		if l.LockedUntil != nil {
			s := l.LockedUntil.Format(time.RFC3339)
			row.LockedUntil = &s
		}
		if l.LastFailedAt != nil {
			s := l.LastFailedAt.Format(time.RFC3339)
			row.LastFailedAt = &s
		}
		out = append(out, row)
	}
	writeJSON(w, http.StatusOK, lockoutsResponse{Items: out})
}

func (h *Handler) handleSignals(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := h.requireTenant(w, r)
	if !ok {
		return
	}
	if h.service == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "Security service is not configured"))
		return
	}
	signals, err := h.service.ListSignals(r.Context(), tenantID)
	if err != nil {
		slog.Error("security: signals failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to list security signals"))
		return
	}
	out := make([]signalItem, 0, len(signals))
	for _, s := range signals {
		row := signalItem{
			SignalID:   s.SignalID,
			Kind:       s.Kind,
			Severity:   s.Severity,
			Summary:    s.Summary,
			DetectedAt: s.DetectedAt.Format(time.RFC3339),
		}
		if len(s.Evidence) > 0 {
			row.Evidence = s.Evidence
		}
		out = append(out, row)
	}
	writeJSON(w, http.StatusOK, signalsResponse{Items: out})
}

func (h *Handler) requireTenant(w http.ResponseWriter, r *http.Request) (string, bool) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
		return "", false
	}
	return tenantID, true
}

func (h *Handler) writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if err := problem.Write(w, p); err != nil {
		slog.Warn("security: write response failed", "err", err)
	}
}

func mfaCoverageToJSON(c securitydomain.MfaCoverage) mfaCoverageResponse {
	byRole := make([]mfaCoverageByRoleItem, 0, len(c.ByRole))
	for _, s := range c.ByRole {
		byRole = append(byRole, mfaCoverageByRoleItem{
			Role:       s.Role,
			Total:      s.Total,
			MfaEnabled: s.MfaEnabled,
			Pct:        s.Pct,
		})
	}
	return mfaCoverageResponse{
		TotalUsers:    c.TotalUsers,
		MfaEnabled:    c.MfaEnabled,
		MfaEnabledPct: c.MfaEnabledPct,
		ByRole:        byRole,
	}
}
