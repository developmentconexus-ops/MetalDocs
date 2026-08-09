package observability

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// RuntimeStatusProvider serves liveness, readiness, and runtime metrics data
// for the /live, /ready, and /metrics endpoints.
type RuntimeStatusProvider interface {
	Live(ctx context.Context) (int, map[string]any)
	Ready(ctx context.Context) (int, map[string]any)
	RuntimeMetrics(ctx context.Context) map[string]any
}

// DependencyCheckResult is the outcome of one DependencyCheck invocation.
type DependencyCheckResult struct {
	Status string
	Detail string
	// Meta keys are defined in observability/keys.go (or the equivalent runtime payload contract).
	Meta map[string]any
}

// DependencyCheck is a named readiness probe run against an external dependency.
type DependencyCheck struct {
	Name  string
	Check func(context.Context) (DependencyCheckResult, error)
}

// StaticRuntimeStatusProvider implements RuntimeStatusProvider from
// statically configured repository/storage mode flags plus a set of
// dependency checks, with no live database access of its own.
type StaticRuntimeStatusProvider struct {
	repositoryMode  string
	storageProvider string
	authEnabled     bool
	checks          []DependencyCheck
}

// NewStaticRuntimeStatusProvider constructs a StaticRuntimeStatusProvider.
func NewStaticRuntimeStatusProvider(repositoryMode, storageProvider string, authEnabled bool, checks ...DependencyCheck) *StaticRuntimeStatusProvider {
	return &StaticRuntimeStatusProvider{
		repositoryMode:  repositoryMode,
		storageProvider: storageProvider,
		authEnabled:     authEnabled,
		checks:          checks,
	}
}

// Live returns the liveness status; it is process-local and never fails.
func (p *StaticRuntimeStatusProvider) Live(_ context.Context) (int, map[string]any) {
	return 200, map[string]any{
		"status": "live",
		"checks": []map[string]any{
			{"name": "process", "status": "up"},
		},
	}
}

// Ready returns the readiness status, running the configured dependency
// checks (this static provider always reports its own repository/storage/auth
// checks as up).
func (p *StaticRuntimeStatusProvider) Ready(ctx context.Context) (int, map[string]any) {
	// Exposure: the /api/v1/metrics payload is gated by CapMetricsView in
	// permissions.go (tier-1); health/readiness endpoints are intentionally
	// public for load-balancer probes and expose only coarse up/down status.
	status := "ready"
	code := 200
	checks := p.applyDependencyChecks(ctx, []map[string]any{
		{"name": "repository", "status": "up"},
		{"name": "storage", "status": "up"},
		{"name": "auth", "status": "up"},
	}, &status, &code)
	return code, map[string]any{
		"status": status,
		"checks": checks,
	}
}

// RuntimeMetrics returns static runtime metrics with zeroed counters (no
// database access).
func (p *StaticRuntimeStatusProvider) RuntimeMetrics(_ context.Context) map[string]any {
	// Exposure: the /api/v1/metrics payload is gated by CapMetricsView in
	// permissions.go (tier-1); health/readiness endpoints are intentionally
	// public for load-balancer probes and expose only coarse up/down status.
	return map[string]any{
		"repositoryMode":  p.repositoryMode,
		"storageProvider": p.storageProvider,
		"authEnabled":     p.authEnabled,
		"auth": map[string]any{
			"users": map[string]any{
				"active":             0,
				"inactive":           0,
				"mustChangePassword": 0,
				"locked":             0,
			},
			"sessions": map[string]any{
				"active":  0,
				"expired": 0,
				"revoked": 0,
			},
		},
		"worker": map[string]any{
			"outbox": map[string]any{
				"claimable":    0,
				"pending":      0,
				"deadLettered": 0,
			},
		},
	}
}

// PostgresRuntimeStatusProvider implements RuntimeStatusProvider backed by a
// live Postgres connection, overriding Ready and RuntimeMetrics to query the
// database while falling back to StaticRuntimeStatusProvider behavior.
type PostgresRuntimeStatusProvider struct {
	// embedded for JSON serialization; do not add methods to the embedded type.
	*StaticRuntimeStatusProvider
	db *sql.DB
}

// NewPostgresRuntimeStatusProvider constructs a PostgresRuntimeStatusProvider
// backed by db.
func NewPostgresRuntimeStatusProvider(db *sql.DB, repositoryMode, storageProvider string, authEnabled bool, checks ...DependencyCheck) *PostgresRuntimeStatusProvider {
	return &PostgresRuntimeStatusProvider{
		StaticRuntimeStatusProvider: NewStaticRuntimeStatusProvider(repositoryMode, storageProvider, authEnabled, checks...),
		db:                          db,
	}
}

// Ready returns the readiness status, pinging the database and running the
// configured dependency checks.
func (p *PostgresRuntimeStatusProvider) Ready(ctx context.Context) (int, map[string]any) {
	// Exposure: the /api/v1/metrics payload is gated by CapMetricsView in
	// permissions.go (tier-1); health/readiness endpoints are intentionally
	// public for load-balancer probes and expose only coarse up/down status.
	readyCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	checks := []map[string]any{
		{"name": "repository", "status": "up"},
		{"name": "storage", "status": "up"},
		{"name": "auth", "status": "up"},
	}
	status := "ready"
	code := 200

	if p.db == nil {
		status = "degraded"
		code = 503
		checks[0]["status"] = "down"
		checks[0]["detail"] = "runtime: database handle is not configured"
	} else if err := p.db.PingContext(readyCtx); err != nil {
		status = "degraded"
		code = 503
		checks[0]["status"] = "down"
		checks[0]["detail"] = truncateReadinessError(err)
	}

	statusPtr := &status
	codePtr := &code
	checks = p.applyDependencyChecks(readyCtx, checks, statusPtr, codePtr)
	status = *statusPtr
	code = *codePtr

	return code, map[string]any{
		"status": status,
		"checks": checks,
	}
}

// RuntimeMetrics returns live auth, session, and outbox counters queried from
// the database, omitting any section whose query failed.
func (p *PostgresRuntimeStatusProvider) RuntimeMetrics(ctx context.Context) map[string]any {
	// Exposure: the /api/v1/metrics payload is gated by CapMetricsView in
	// permissions.go (tier-1); health/readiness endpoints are intentionally
	// public for load-balancer probes and expose only coarse up/down status.
	metrics := p.StaticRuntimeStatusProvider.RuntimeMetrics(ctx)
	if p.db == nil {
		metrics["repositoryStatus"] = "down"
		return metrics
	}

	type authStats struct {
		active             int
		inactive           int
		mustChangePassword int
		locked             int
	}
	type sessionStats struct {
		active  int
		expired int
		revoked int
	}
	type outboxStats struct {
		claimable    int
		pending      int
		deadLettered int
	}

	auth := authStats{}
	sessions := sessionStats{}
	outbox := outboxStats{}

	authErr := queryRuntimeMetric(ctx, 3*time.Second, func(queryCtx context.Context) error {
		return p.db.QueryRowContext(queryCtx, `
SELECT
  COUNT(*) FILTER (WHERE is_active = TRUE),
  COUNT(*) FILTER (WHERE is_active = FALSE),
  COUNT(*) FILTER (WHERE must_change_password = TRUE),
  COUNT(*) FILTER (WHERE locked_until IS NOT NULL AND locked_until > NOW())
FROM metaldocs.auth_identities
`).Scan(&auth.active, &auth.inactive, &auth.mustChangePassword, &auth.locked)
	})
	sessionErr := queryRuntimeMetric(ctx, 3*time.Second, func(queryCtx context.Context) error {
		return p.db.QueryRowContext(queryCtx, `
SELECT
  COUNT(*) FILTER (WHERE revoked_at IS NULL AND expires_at > NOW()),
  COUNT(*) FILTER (WHERE expires_at <= NOW()),
  COUNT(*) FILTER (WHERE revoked_at IS NOT NULL)
FROM metaldocs.auth_sessions
`).Scan(&sessions.active, &sessions.expired, &sessions.revoked)
	})
	outboxErr := queryRuntimeMetric(ctx, 3*time.Second, func(queryCtx context.Context) error {
		return p.db.QueryRowContext(queryCtx, `
SELECT
  COUNT(*) FILTER (WHERE published_at IS NULL AND dead_lettered_at IS NULL AND (next_attempt_at IS NULL OR next_attempt_at <= NOW())),
  COUNT(*) FILTER (WHERE published_at IS NULL AND dead_lettered_at IS NULL),
  COUNT(*) FILTER (WHERE dead_lettered_at IS NOT NULL)
FROM metaldocs.outbox_events
`).Scan(&outbox.claimable, &outbox.pending, &outbox.deadLettered)
	})

	metrics["repositoryStatus"] = "up"
	authMetrics := map[string]any{}
	if authErr == nil {
		authMetrics["users"] = map[string]any{
			"active":             auth.active,
			"inactive":           auth.inactive,
			"mustChangePassword": auth.mustChangePassword,
			"locked":             auth.locked,
		}
	}
	if sessionErr == nil {
		authMetrics["sessions"] = map[string]any{
			"active":  sessions.active,
			"expired": sessions.expired,
			"revoked": sessions.revoked,
		}
	}
	if len(authMetrics) > 0 {
		metrics["auth"] = authMetrics
	} else {
		delete(metrics, "auth")
	}
	if outboxErr == nil {
		metrics["worker"] = map[string]any{
			"outbox": map[string]any{
				"claimable":    outbox.claimable,
				"pending":      outbox.pending,
				"deadLettered": outbox.deadLettered,
			},
		}
	} else {
		delete(metrics, "worker")
	}

	errors := map[string]string{}
	if authErr != nil {
		errors["auth"] = truncateReadinessError(authErr)
	}
	if sessionErr != nil {
		errors["sessions"] = truncateReadinessError(sessionErr)
	}
	if outboxErr != nil {
		errors["worker"] = truncateReadinessError(outboxErr)
	}
	if len(errors) > 0 {
		metrics["repositoryStatus"] = "degraded"
		metrics["errors"] = errors
	}

	return metrics
}

func truncateReadinessError(err error) string {
	if err == nil {
		return ""
	}
	msg := "runtime: " + err.Error()
	if len(msg) > 160 {
		return msg[:160]
	}
	return msg
}

func queryRuntimeMetric(ctx context.Context, timeout time.Duration, query func(context.Context) error) error {
	queryCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return query(queryCtx)
}

// applyDependencyChecks runs every configured dependency check and appends
// its result entry to checks. status/code are optional out-params so Ready
// can reuse this helper without allocating a result wrapper for the common
// static-provider fast path.
func (p *StaticRuntimeStatusProvider) applyDependencyChecks(ctx context.Context, checks []map[string]any, status *string, code *int) []map[string]any {
	if len(p.checks) == 0 {
		return checks
	}

	active := activeDependencyChecks(p.checks)
	if len(active) == 0 {
		return checks
	}

	// All dependency checks share a single 5s budget via errgroup so the probe
	// latency is bounded by the slowest single check, not their sum.
	budgetCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for _, entry := range runDependencyChecksConcurrently(budgetCtx, active) {
		checks = append(checks, entry)
		degradeStatusIfUnhealthy(entry, status, code)
	}
	return checks
}

// activeDependencyChecks filters out entries with a nil Check func.
func activeDependencyChecks(all []DependencyCheck) []DependencyCheck {
	active := make([]DependencyCheck, 0, len(all))
	for _, c := range all {
		if c.Check != nil {
			active = append(active, c)
		}
	}
	return active
}

// runDependencyChecksConcurrently runs each check under budgetCtx via
// errgroup and returns one result entry per check, in the same order as
// active.
func runDependencyChecksConcurrently(budgetCtx context.Context, active []DependencyCheck) []map[string]any {
	results := make([]map[string]any, len(active))
	var mu sync.Mutex

	g, gCtx := errgroup.WithContext(budgetCtx)
	for i, check := range active {
		i, check := i, check
		g.Go(func() error {
			entry := runSingleDependencyCheck(gCtx, check)
			mu.Lock()
			results[i] = entry
			mu.Unlock()
			return nil
		})
	}
	// errgroup.Go callbacks never return a non-nil error here; wait only for
	// completion.
	_ = g.Wait()
	return results
}

// runSingleDependencyCheck invokes check.Check and shapes its result/error
// into the wire entry: {"name", "status", "detail"?, ...Meta}.
func runSingleDependencyCheck(ctx context.Context, check DependencyCheck) map[string]any {
	result, err := check.Check(ctx)
	entry := map[string]any{
		"name":   check.Name,
		"status": "up",
	}
	if result.Status != "" {
		entry["status"] = result.Status
	}
	if result.Detail != "" {
		entry["detail"] = result.Detail
	}
	for key, value := range result.Meta {
		entry[key] = value
	}
	if err != nil {
		entry["status"] = "down"
		entry["detail"] = truncateReadinessError(err)
	}
	return entry
}

// degradeStatusIfUnhealthy flips the optional out-params to "degraded"/503
// when entry's status is neither "up" nor "skipped".
func degradeStatusIfUnhealthy(entry map[string]any, status *string, code *int) {
	state := entry["status"]
	if state == "up" || state == "skipped" {
		return
	}
	if status != nil {
		*status = "degraded"
	}
	if code != nil {
		*code = 503
	}
}
