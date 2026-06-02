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
		w.WriteHeader(http.StatusMethodNotAllowed)
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
		w.WriteHeader(http.StatusMethodNotAllowed)
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
			"userId":         l.UserID,
			"displayName":    l.DisplayName,
			"failedAttempts": l.FailedAttempts,
		}
		if l.LockedUntil != nil {
			row["lockedUntil"] = l.LockedUntil.Format(time.RFC3339)
		}
		if l.LastFailedAt != nil {
			row["lastFailedAt"] = l.LastFailedAt.Format(time.RFC3339)
		}
		if l.LastFailedIP != "" {
			row["lastFailedIp"] = l.LastFailedIP
		}
		out = append(out, row)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handler) handleSignals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
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
			"signalId":   s.SignalID,
			"kind":       s.Kind,
			"severity":   s.Severity,
			"summary":    s.Summary,
			"detectedAt": s.DetectedAt.Format(time.RFC3339),
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
			"mfaEnabled": s.MfaEnabled,
			"pct":        s.Pct,
		})
	}
	return map[string]any{
		"totalUsers":    c.TotalUsers,
		"mfaEnabled":    c.MfaEnabled,
		"mfaEnabledPct": c.MfaEnabledPct,
		"byRole":        byRole,
	}
}

func writeMfaCoverageZero(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]any{
		"totalUsers":    0,
		"mfaEnabled":    0,
		"mfaEnabledPct": 0,
		"byRole":        []any{},
	})
}
