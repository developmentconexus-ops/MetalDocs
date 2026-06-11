// Package httpdelivery exposes the Sessions & Security tab read endpoints
// (PR-7) over HTTP. Every endpoint depends on tenant.FromContext — the
// middleware in apps/api/cmd/metaldocs-api/permissions.go already enforces
// auth + CapUserView, so the handler can treat the context as trusted.
package httpdelivery

import (
	"log/slog"
	"net/http"
	"time"

	securityapp "metaldocs/internal/modules/security/application"
	securitydomain "metaldocs/internal/modules/security/domain"
	httpresponse "metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

var writeJSON = httpresponse.WriteJSON

type Handler struct {
	service *securityapp.Service
}

func NewHandler(service *securityapp.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/security/mfa-coverage", h.handleMfaCoverage)
	mux.HandleFunc("/api/v1/security/lockouts", h.handleLockouts)
	mux.HandleFunc("/api/v1/security/signals", h.handleSignals)
}

func (h *Handler) handleMfaCoverage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.methodNotAllowed(w)
		return
	}
	tenantID, ok := h.requireTenant(w, r)
	if !ok {
		return
	}
	if h.service == nil {
		writeMfaCoverageZero(w)
		return
	}
	coverage, err := h.service.MfaCoverage(r.Context(), tenantID)
	if err != nil {
		slog.Error("security: mfa coverage failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load MFA coverage"))
		return
	}
	writeJSON(w, http.StatusOK, mfaCoverageToJSON(coverage))
}

func (h *Handler) handleLockouts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.methodNotAllowed(w)
		return
	}
	tenantID, ok := h.requireTenant(w, r)
	if !ok {
		return
	}
	if h.service == nil {
		writeJSON(w, http.StatusOK, map[string]any{"items": []any{}})
		return
	}
	items, err := h.service.ListLockouts(r.Context(), tenantID)
	if err != nil {
		slog.Error("security: lockouts failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list lockouts"))
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, l := range items {
		row := map[string]any{
			"user_id":         l.UserID,
			"display_name":    l.DisplayName,
			"failed_attempts": l.FailedAttempts,
		}
		if l.LockedUntil != nil {
			row["locked_until"] = l.LockedUntil.Format(time.RFC3339)
		}
		if l.LastFailedAt != nil {
			row["last_failed_at"] = l.LastFailedAt.Format(time.RFC3339)
		}
		if l.LastFailedIP != "" {
			row["last_failed_ip"] = l.LastFailedIP
		}
		out = append(out, row)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handler) handleSignals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.methodNotAllowed(w)
		return
	}
	tenantID, ok := h.requireTenant(w, r)
	if !ok {
		return
	}
	if h.service == nil {
		writeJSON(w, http.StatusOK, map[string]any{"items": []any{}})
		return
	}
	signals, err := h.service.ListSignals(r.Context(), tenantID)
	if err != nil {
		slog.Error("security: signals failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list security signals"))
		return
	}
	out := make([]map[string]any, 0, len(signals))
	for _, s := range signals {
		row := map[string]any{
			"signal_id":   s.SignalID,
			"kind":       s.Kind,
			"severity":   s.Severity,
			"summary":    s.Summary,
			"detected_at": s.DetectedAt.Format(time.RFC3339),
		}
		if len(s.Evidence) > 0 {
			row["evidence"] = s.Evidence
		}
		out = append(out, row)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handler) requireTenant(w http.ResponseWriter, r *http.Request) (string, bool) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required"))
		return "", false
	}
	return tenantID, true
}

// methodNotAllowed writes the RFC 9457 405 response (D-03: every error path
// in this module emits problem+json — bare statuses were the one exception).
func (h *Handler) methodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Allow", http.MethodGet)
	h.writeProblem(w, problem.New(http.StatusMethodNotAllowed, problem.CodeMethodNotAllowed, "Method not allowed"))
}

func (h *Handler) writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if err := problem.Write(w, p); err != nil {
		slog.Warn("security: write response failed", "err", err)
	}
}

func mfaCoverageToJSON(c securitydomain.MfaCoverage) map[string]any {
	byRole := make([]map[string]any, 0, len(c.ByRole))
	for _, s := range c.ByRole {
		byRole = append(byRole, map[string]any{
			"role":       s.Role,
			"total":      s.Total,
			"mfa_enabled": s.MfaEnabled,
			"pct":        s.Pct,
		})
	}
	return map[string]any{
		"total_users":    c.TotalUsers,
		"mfa_enabled":    c.MfaEnabled,
		"mfa_enabled_pct": c.MfaEnabledPct,
		"by_role":        byRole,
	}
}

func writeMfaCoverageZero(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]any{
		"total_users":    0,
		"mfa_enabled":    0,
		"mfa_enabled_pct": 0,
		"by_role":        []any{},
	})
}
