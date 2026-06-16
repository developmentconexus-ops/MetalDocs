# F2.4 Metrics Completeness — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add DB-pool stats to `/api/v1/metrics` payload and fix the misleading "Prometheus" comment.

**Architecture:** Same provider-interface pattern as F2.2 `SchedulerMetricsProvider` — `DBPoolStatsProvider` interface on `HTTPObservability`, thin `SQLDBPoolStatsAdapter` wrapping `*sql.DB`, wired in main.go after `httpObs` creation.

**Tech Stack:** Go stdlib `database/sql.DBStats`, `internal/platform/observability/http.go`, `internal/platform/db/postgres/`.

---

### Task 1: Write failing tests (TDD red)

**Files:**
- Create: `internal/platform/observability/http_dbpool_test.go`
- Create: `internal/platform/db/postgres/pool_stats_test.go`

- [ ] **Step 1: Write observability test file**

```go
// internal/platform/observability/http_dbpool_test.go
package observability_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/observability"
)

type stubDBPool struct {
	stats map[string]any
}

func (s *stubDBPool) DBPoolStats() map[string]any { return s.stats }

func TestMetricsHandler_IncludesDBPoolStats(t *testing.T) {
	obs := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	obs.SetDBPool(&stubDBPool{stats: map[string]any{"in_use": 2, "idle": 3}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rec := httptest.NewRecorder()
	obs.MetricsHandler().ServeHTTP(rec, req)

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	pool, ok := payload["db_pool"]
	if !ok {
		t.Fatal("expected db_pool key in metrics payload")
	}
	poolMap, ok := pool.(map[string]any)
	if !ok {
		t.Fatalf("db_pool should be map, got %T", pool)
	}
	if v, ok := poolMap["in_use"]; !ok || v != float64(2) {
		t.Errorf("expected in_use=2, got %v", poolMap["in_use"])
	}
}

func TestMetricsHandler_NoDBPoolKey_WhenNotWired(t *testing.T) {
	obs := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	// SetDBPool NOT called

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rec := httptest.NewRecorder()
	obs.MetricsHandler().ServeHTTP(rec, req)

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := payload["db_pool"]; ok {
		t.Fatal("db_pool key should be absent when SetDBPool not called")
	}
}
```

- [ ] **Step 2: Write adapter test file**

```go
// internal/platform/db/postgres/pool_stats_test.go
package postgres_test

import (
	"database/sql"
	"testing"

	"metaldocs/internal/platform/db/postgres"
)

func TestSQLDBPoolStatsAdapter_Maps(t *testing.T) {
	stats := sql.DBStats{
		MaxOpenConnections: 10,
		OpenConnections:    4,
		InUse:              2,
		Idle:               2,
		WaitCount:          5,
	}
	adapter := postgres.NewPoolStatsAdapterFromStats(stats)
	m := adapter.DBPoolStats()

	cases := []struct {
		key  string
		want float64
	}{
		{"max_open_connections", 10},
		{"open_connections", 4},
		{"in_use", 2},
		{"idle", 2},
		{"wait_count", 5},
	}
	for _, tc := range cases {
		v, ok := m[tc.key]
		if !ok {
			t.Errorf("missing key %q", tc.key)
			continue
		}
		if got, ok := v.(int); ok {
			if float64(got) != tc.want {
				t.Errorf("key %q: want %v got %v", tc.key, tc.want, got)
			}
		} else if got64, ok := v.(int64); ok {
			if float64(got64) != tc.want {
				t.Errorf("key %q: want %v got %v", tc.key, tc.want, got64)
			}
		} else {
			t.Errorf("key %q: unexpected type %T value %v", tc.key, v, v)
		}
	}
}
```

- [ ] **Step 3: Run tests — expect compile errors (TDD red)**

```powershell
go test ./internal/platform/observability/ -run "TestMetricsHandler_IncludesDBPoolStats|TestMetricsHandler_NoDBPoolKey_WhenNotWired" -count=1
go test ./internal/platform/db/postgres/ -run TestSQLDBPoolStatsAdapter_Maps -count=1
```

Expected: compile error — `observability.SetDBPool` and `postgres.NewPoolStatsAdapterFromStats` undefined.

- [ ] **Step 4: Commit red tests**

```bash
git add internal/platform/observability/http_dbpool_test.go internal/platform/db/postgres/pool_stats_test.go
git commit -m "test(f2.4): failing tests for db_pool metrics key and pool stats adapter [TDD red]"
```

---

### Task 2: Implement `DBPoolStatsProvider` + `SetDBPool` in observability/http.go

**Files:**
- Modify: `internal/platform/observability/http.go`

- [ ] **Step 1: Add interface, field, setter, and payload merge**

After the `SchedulerMetricsProvider` block (around line 22–30), add:

```go
// DBPoolStatsProvider exposes database connection pool stats for the metrics endpoint.
// *sql.DB does not implement this directly — use postgres.NewPoolStatsAdapter.
type DBPoolStatsProvider interface {
	DBPoolStats() map[string]any
}
```

Add field to `HTTPObservability` struct (after `schedulerMetrics`):

```go
dbPool DBPoolStatsProvider
```

Add setter (after `SetSchedulerMetrics`):

```go
// SetDBPool wires a DB pool stats provider. Call once during startup, before
// ListenAndServe. Absent → "db_pool" key omitted from MetricsHandler response.
func (o *HTTPObservability) SetDBPool(p DBPoolStatsProvider) {
	o.dbPool = p
}
```

Add nil-guard in `MetricsHandler()` body (after the `schedulerMetrics` nil-guard):

```go
if o.dbPool != nil {
	payload["db_pool"] = o.dbPool.DBPoolStats()
}
```

- [ ] **Step 2: Run tests — expect only the adapter test to still fail**

```powershell
go test ./internal/platform/observability/ -run "TestMetricsHandler_IncludesDBPoolStats|TestMetricsHandler_NoDBPoolKey_WhenNotWired" -count=1
```

Expected: PASS (observability tests now green). Adapter test still fails (postgres package not yet updated).

---

### Task 3: Implement `SQLDBPoolStatsAdapter` in postgres package

**Files:**
- Create: `internal/platform/db/postgres/pool_stats.go`

- [ ] **Step 1: Write adapter file**

```go
package postgres

import (
	"database/sql"
	"time"
)

// SQLDBPoolStatsAdapter implements observability.DBPoolStatsProvider by
// wrapping *sql.DB and mapping sql.DBStats fields to snake_case map keys.
type SQLDBPoolStatsAdapter struct {
	db *sql.DB
}

// NewPoolStatsAdapter wraps db for use as a DBPoolStatsProvider.
// Returns nil when db is nil (caller should nil-check before SetDBPool).
func NewPoolStatsAdapter(db *sql.DB) *SQLDBPoolStatsAdapter {
	if db == nil {
		return nil
	}
	return &SQLDBPoolStatsAdapter{db: db}
}

// NewPoolStatsAdapterFromStats builds an adapter from a known sql.DBStats value.
// Used in tests where a real *sql.DB is not available.
func NewPoolStatsAdapterFromStats(stats sql.DBStats) *sqlDBStatsAdapter {
	return &sqlDBStatsAdapter{stats: stats}
}

// DBPoolStats returns pool stats as a map[string]any with snake_case keys.
func (a *SQLDBPoolStatsAdapter) DBPoolStats() map[string]any {
	return dbStatsToMap(a.db.Stats())
}

// sqlDBStatsAdapter is the test-only variant backed by a pre-built sql.DBStats.
type sqlDBStatsAdapter struct {
	stats sql.DBStats
}

func (a *sqlDBStatsAdapter) DBPoolStats() map[string]any {
	return dbStatsToMap(a.stats)
}

func dbStatsToMap(s sql.DBStats) map[string]any {
	return map[string]any{
		"max_open_connections": s.MaxOpenConnections,
		"open_connections":     s.OpenConnections,
		"in_use":              s.InUse,
		"idle":                s.Idle,
		"wait_count":          s.WaitCount,
		"wait_duration_ms":    s.WaitDuration / time.Millisecond,
		"max_idle_closed":     s.MaxIdleClosed,
		"max_idle_time_closed": s.MaxIdleTimeClosed,
		"max_lifetime_closed": s.MaxLifetimeClosed,
	}
}
```

- [ ] **Step 2: Run adapter test**

```powershell
go test ./internal/platform/db/postgres/ -run TestSQLDBPoolStatsAdapter_Maps -count=1
```

Expected: PASS.

---

### Task 4: Wire `SetDBPool` in main.go + fix Prometheus comment

**Files:**
- Modify: `apps/api/cmd/metaldocs-api/main.go`
- Modify: `apps/api/cmd/metaldocs-api/permissions.go`

- [ ] **Step 1: Wire adapter in main.go**

After the `httpObs` creation block (around line 285), add:

```go
if deps.SQLDB != nil {
	httpObs.SetDBPool(postgres.NewPoolStatsAdapter(deps.SQLDB))
}
```

Check existing imports for the postgres package. The import alias used elsewhere in main.go:

```bash
grep -n 'postgres\|pgconn\|pgdb' apps/api/cmd/metaldocs-api/main.go | head -10
```

Add import if not already present:
```go
"metaldocs/internal/platform/db/postgres"
```

- [ ] **Step 2: Fix Prometheus comment**

In `apps/api/cmd/metaldocs-api/permissions.go` line 95, change:

```go
// Metrics — Prometheus scrape endpoint; read-only.
```

to:

```go
// Metrics — JSON scrape endpoint; read-only.
```

- [ ] **Step 3: Build check**

```powershell
go build ./apps/api/cmd/metaldocs-api/
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add apps/api/cmd/metaldocs-api/main.go apps/api/cmd/metaldocs-api/permissions.go internal/platform/db/postgres/pool_stats.go internal/platform/observability/http.go
git commit -m "feat(f2.4): db_pool stats in /api/v1/metrics + fix Prometheus comment"
```

---

### Task 5: Regression + runtime proof + final commit

**Files:** None new.

- [ ] **Step 1: Full regression**

```powershell
go test ./... 2>&1 | Select-String -Pattern "FAIL|ok" | head -60
```

Expected: no FAIL lines.

- [ ] **Step 2: Verify Prometheus comment removed**

```powershell
Select-String -Path apps/api/cmd/metaldocs-api/permissions.go -Pattern "Prometheus scrape endpoint"
```

Expected: no output (0 matches).

- [ ] **Step 3: Runtime proof**

Start API (if not running):
```powershell
.\scripts\start-api.ps1
```

Login and get token:
```powershell
$r = Invoke-RestMethod -Uri http://localhost:8081/api/v1/auth/login -Method Post -ContentType "application/json" -Body '{"identifier":"admin","password":"AdminMetalDocs123!"}'
$tok = $r.session_token
```

Scrape metrics:
```powershell
Invoke-RestMethod -Uri http://localhost:8081/api/v1/metrics -Headers @{"Authorization"="Bearer $tok"} | ConvertTo-Json -Depth 5 | Select-String -Pattern "db_pool" -A 10
```

Expected: `"db_pool"` object with numeric values (at least `open_connections` > 0 after startup queries).

Paste verbatim output snippet into `evidence.md`.

- [ ] **Step 4: Commit evidence**

```bash
git add docs/superpowers/milestones/grade-a-completion/milestone-2-observability/f2.4-metrics-completeness/evidence.md
git commit -m "docs(f2.4): evidence — db_pool stats key + comment fix verified"
```
