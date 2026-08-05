package httpdelivery

import (
	"log/slog"
	"metaldocs/internal/platform/httprouter"
	"net/http"

	iamapi "metaldocs/internal/modules/iam/api"
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

// NewObservabilityHandler constructs the handler. A nil service causes both
// endpoints to answer 501 NOT_IMPLEMENTED rather than panic.
func NewObservabilityHandler(service *iamapp.ObservabilityService) *ObservabilityHandler {
	return &ObservabilityHandler{service: service}
}

// RegisterRoutes mounts GET /iam/usage and GET /iam/kpi on mux. Both routes
// are permission-guarded by CapMetricsView at the tier-1 routing table.
func (h *ObservabilityHandler) RegisterRoutes(mux httprouter.Muxer) {
	mux.HandleFunc(http.MethodGet+" /api/v1/iam/usage", h.handleUsage)
	mux.HandleFunc(http.MethodGet+" /api/v1/iam/kpi", h.handleKpi)
}

func (h *ObservabilityHandler) handleUsage(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := h.requireTenant(w, r)
	if !ok {
		return
	}
	if h.service == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "Observability service is not configured"))
		return
	}
	usage, err := h.service.GetUsage(r.Context(), tenantID)
	if err != nil {
		slog.Error("iam observability: usage failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to load usage"))
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
		h.writeProblem(w, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "Observability service is not configured"))
		return
	}
	kpi, err := h.service.GetKpi(r.Context(), tenantID)
	if err != nil {
		slog.Error("iam observability: kpi failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to load KPI"))
		return
	}
	writeJSON(w, http.StatusOK, kpiToJSON(kpi))
}

func (h *ObservabilityHandler) requireTenant(w http.ResponseWriter, r *http.Request) (string, bool) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "Authentication required"))
		return "", false
	}
	return tenantID, true
}

func (h *ObservabilityHandler) writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if err := problem.Write(w, p); err != nil {
		slog.Warn("iam observability: write response failed", "err", err)
	}
}

func usageToJSON(u iamdomain.UsageSnapshot) iamapi.UsageSnapshot {
	// iamdomain.CountWindows and StorageUsage use int64; iamapi uses int.
	// Values are safe to narrow: seat counts and window counts never approach
	// int32 max in practice, and storage bytes is capped by the plan envelope.
	snap := iamapi.UsageSnapshot{
		ApiCalls: iamapi.UsageWindowCounts{
			Last24h: int(u.APICalls.Last24h),
			Last7d:  int(u.APICalls.Last7d),
			Last30d: int(u.APICalls.Last30d),
		},
		ActiveUsers: iamapi.UsageWindowCounts{
			Last24h: int(u.ActiveUsers.Last24h),
			Last7d:  int(u.ActiveUsers.Last7d),
			Last30d: int(u.ActiveUsers.Last30d),
		},
	}
	snap.Seats.Used = u.Seats.Used
	snap.Seats.Allocated = u.Seats.Allocated
	snap.Storage.UsedBytes = int(u.Storage.UsedBytes)
	snap.Storage.AllocatedBytes = int(u.Storage.AllocatedBytes)
	if u.PlanTier != "" {
		tier := iamapi.UsageSnapshotPlanTier(string(u.PlanTier))
		snap.PlanTier = &tier
	}
	return snap
}

func kpiToJSON(k iamdomain.KpiSnapshot) iamapi.IamKpiSnapshot {
	dist := make([]iamapi.IamKpiRoleCount, 0, len(k.RoleDistribution))
	for _, rc := range k.RoleDistribution {
		dist = append(dist, iamapi.IamKpiRoleCount{
			Role:  iamapi.UserRole(string(rc.Role)),
			Count: rc.Count,
		})
	}
	return iamapi.IamKpiSnapshot{
		LockedAccounts:       k.LockedAccounts,
		MfaCoveragePct:       k.MfaCoveragePct,
		FailedLogins24h:      int(k.FailedLogins24h),
		DormantUsers30d:      k.DormantUsers30d,
		RoleDistribution:     dist,
		AuditEventsPerMinute: k.AuditEventsPerMinute,
	}
}
