package httpdelivery

import (
	"log/slog"
	"net/http"

	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// ObservabilityHandler exposes the two Tier-B observability endpoints
// introduced in PR-8: GET /iam/usage and GET /iam/kpi. Both endpoints are
// permission-guarded by CapMetricsView at the routing table; the handler
// trusts the context to carry a tenant id.
type ObservabilityHandler struct {
	service *iamapp.ObservabilityService
}

func NewObservabilityHandler(service *iamapp.ObservabilityService) *ObservabilityHandler {
	return &ObservabilityHandler{service: service}
}

func (h *ObservabilityHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodGet+" /api/v1/iam/usage", h.handleUsage)
	mux.HandleFunc(http.MethodGet+" /api/v1/iam/kpi", h.handleKpi)
}

func (h *ObservabilityHandler) handleUsage(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := h.requireTenant(w, r)
	if !ok {
		return
	}
	if h.service == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "Observability service is not configured"))
		return
	}
	usage, err := h.service.GetUsage(r.Context(), tenantID)
	if err != nil {
		slog.Error("iam observability: usage failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load usage"))
		return
	}
	writeJSON(w, http.StatusOK, usageToJSON(usage))
}

func (h *ObservabilityHandler) handleKpi(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := h.requireTenant(w, r)
	if !ok {
		return
	}
	if h.service == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "Observability service is not configured"))
		return
	}
	kpi, err := h.service.GetKpi(r.Context(), tenantID)
	if err != nil {
		slog.Error("iam observability: kpi failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load KPI"))
		return
	}
	writeJSON(w, http.StatusOK, kpiToJSON(kpi))
}

func (h *ObservabilityHandler) requireTenant(w http.ResponseWriter, r *http.Request) (string, bool) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required"))
		return "", false
	}
	return tenantID, true
}

func (h *ObservabilityHandler) writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if err := problem.Write(w, p); err != nil {
		slog.Warn("iam observability: write response failed", "err", err)
	}
}

func usageToJSON(u iamdomain.UsageSnapshot) map[string]any {
	tier := any(nil)
	if u.PlanTier != "" {
		tier = string(u.PlanTier)
	}
	return map[string]any{
		"seats": map[string]any{
			"used":      u.Seats.Used,
			"allocated": u.Seats.Allocated,
		},
		"storage": map[string]any{
			"usedBytes":      u.Storage.UsedBytes,
			"allocatedBytes": u.Storage.AllocatedBytes,
		},
		"apiCalls": map[string]any{
			"last24h": u.APICalls.Last24h,
			"last7d":  u.APICalls.Last7d,
			"last30d": u.APICalls.Last30d,
		},
		"activeUsers": map[string]any{
			"last24h": u.ActiveUsers.Last24h,
			"last7d":  u.ActiveUsers.Last7d,
			"last30d": u.ActiveUsers.Last30d,
		},
		"planTier": tier,
	}
}

func kpiToJSON(k iamdomain.KpiSnapshot) map[string]any {
	dist := make([]map[string]any, 0, len(k.RoleDistribution))
	for _, rc := range k.RoleDistribution {
		dist = append(dist, map[string]any{
			"role":  string(rc.Role),
			"count": rc.Count,
		})
	}
	return map[string]any{
		"lockedAccounts":       k.LockedAccounts,
		"mfaCoveragePct":       k.MfaCoveragePct,
		"failedLogins24h":      k.FailedLogins24h,
		"dormantUsers30d":      k.DormantUsers30d,
		"roleDistribution":     dist,
		"auditEventsPerMinute": k.AuditEventsPerMinute,
	}
}
