# Feature F2.3 — OTel App Spans — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add driver-level DB call spans (all queries auto-traced via `otelsql`) and two manual semantic spans (`cd.create`, `signoff.record`) to `metaldocs-api`.

**Architecture:** `otelsql.Open` replaces `sql.Open` at `internal/platform/db/postgres/connect.go` (one line); `otel.Tracer().Start` wraps `ControlledDocumentService.Create` and `DecisionService.RecordSignoff` bodies. OTel remains inert (NOOP) when `OTEL_TRACES_EXPORTER` is unset.

**Tech Stack:** `github.com/XSAM/otelsql`, `go.opentelemetry.io/otel` v1.44.0 (already in go.mod), `go.opentelemetry.io/otel/sdk/trace/tracetest` (test recorder), `pgx/v5/stdlib` (already in go.mod).

**ADR:** [`wiki/decisions/0036-otelsql-db-tracing.md`](../../../../../wiki/decisions/0036-otelsql-db-tracing.md) — accepted 2026-06-16.

---

## File map

| Action | File |
|--------|------|
| Modify | `internal/platform/db/postgres/connect.go` |
| Modify | `internal/modules/controlleddocuments/application/service.go` |
| Modify | `internal/modules/documents/approval/application/decision_service.go` |
| Create | `internal/platform/db/postgres/connect_otel_test.go` |
| Create | `internal/modules/controlleddocuments/application/service_otel_test.go` |
| Create | `internal/modules/documents/approval/application/decision_otel_test.go` |

---

### Task 1: Write four failing tests (TDD red)

**Files:**
- Create: `internal/platform/db/postgres/connect_otel_test.go`
- Create: `internal/modules/controlleddocuments/application/service_otel_test.go`
- Create: `internal/modules/documents/approval/application/decision_otel_test.go`

- [ ] **Step 1: Write `connect_otel_test.go`**

```go
package postgres_test

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"metaldocs/internal/platform/db/postgres"
)

func TestOpen_EmitsOTelSpan(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))

	dsn := "postgres://metaldocs:metaldocs@localhost:5432/metaldocs_test?sslmode=disable"
	db, err := postgres.OpenWithTracerProvider(dsn, tp)
	if err != nil {
		t.Skipf("no test DB: %v", err)
	}
	defer db.Close()

	_, err = db.QueryContext(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	spans := sr.Ended()
	if len(spans) == 0 {
		t.Fatal("want ≥1 span, got 0")
	}
	found := false
	for _, s := range spans {
		if s.Name() == "go.sql.query" || s.Name() == "go.sql.exec" || s.Name() == "go.sql" {
			found = true
			break
		}
	}
	if !found {
		names := make([]string, len(spans))
		for i, s := range spans {
			names[i] = s.Name()
		}
		t.Fatalf("want go.sql.* span, got %v", names)
	}
}
```

> Note: `postgres.OpenWithTracerProvider` does not exist yet — test will fail to compile (TDD red). DSN points to test DB; test skips if no DB (not a hard fail — fixture label is for the recorder, not the DB connection).

- [ ] **Step 2: Write `service_otel_test.go`**

```go
package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"metaldocs/internal/modules/controlleddocuments/application"
)

// stubFailRunner returns error immediately without calling tx func.
// Span is created BEFORE runner is called, so span is emitted regardless.
type stubFailRunner struct{ err error }

func (r *stubFailRunner) RunTx(ctx context.Context, fn func(context.Context) error) error {
	return r.err
}

func setupTracerProvider(t *testing.T) (*tracetest.SpanRecorder, func()) {
	t.Helper()
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	return sr, func() { otel.SetTracerProvider(prev) }
}

func TestCreate_EmitsCdCreateSpan(t *testing.T) {
	sr, cleanup := setupTracerProvider(t)
	defer cleanup()

	svc := application.NewControlledDocumentService(
		&stubFailRunner{err: errors.New("stub")},
		nil, nil, nil, nil, nil, nil, nil,
	)

	cmd := application.CreateControlledDocumentCmd{
		ProfileCode: "ISO-9001",
		TenantID:    "tenant-1",
	}
	_, _ = svc.Create(context.Background(), cmd)

	spans := sr.Ended()
	if len(spans) == 0 {
		t.Fatal("want ≥1 span, got 0")
	}
	var found *tracetest.SpanStub
	for i := range spans {
		if spans[i].Name() == "cd.create" {
			found = &spans[i]
			break
		}
	}
	if found == nil {
		names := make([]string, len(spans))
		for i, s := range spans {
			names[i] = s.Name()
		}
		t.Fatalf("want span named 'cd.create', got %v", names)
	}

	// Verify document.profile_code attribute
	attrFound := false
	for _, a := range found.Attributes() {
		if string(a.Key) == "document.profile_code" && a.Value.AsString() == "ISO-9001" {
			attrFound = true
			break
		}
	}
	if !attrFound {
		t.Fatalf("want attribute document.profile_code=ISO-9001, got %v", found.Attributes())
	}
}

func TestCreate_SpanStatusError_OnFailure(t *testing.T) {
	sr, cleanup := setupTracerProvider(t)
	defer cleanup()

	wantErr := errors.New("runner failure")
	svc := application.NewControlledDocumentService(
		&stubFailRunner{err: wantErr},
		nil, nil, nil, nil, nil, nil, nil,
	)

	cmd := application.CreateControlledDocumentCmd{ProfileCode: "X", TenantID: "t"}
	_, err := svc.Create(context.Background(), cmd)
	if err == nil {
		t.Fatal("want error, got nil")
	}

	spans := sr.Ended()
	var found *tracetest.SpanStub
	for i := range spans {
		if spans[i].Name() == "cd.create" {
			found = &spans[i]
			break
		}
	}
	if found == nil {
		t.Fatal("want cd.create span, got none")
	}
	if found.Status().Code != codes.Error {
		t.Fatalf("want status Error, got %v", found.Status().Code)
	}
}

// silence unused import
var _ = time.Now
```

> Note: `stubFailRunner` needs to implement the actual interface from `platformdb`. If the interface is `platformdb.TxRunner`, adjust the type assertion in the test. If `NewControlledDocumentService` rejects nil for non-runner deps (e.g. panics on nil check), adjust stubs accordingly — read the constructor to confirm nil is safe when runner errors first.

- [ ] **Step 3: Write `decision_otel_test.go`**

```go
package application_test

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/application/repository"
	"metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/platform/db"
)

// stubFailTxRunner returns error immediately.
type stubFailTxRunner struct{ err error }

func (r *stubFailTxRunner) RunTx(ctx context.Context, fn func(context.Context) error) error {
	return r.err
}

// stubEmitter satisfies EventEmitter interface.
type stubEmitter struct{}
func (s *stubEmitter) Emit(ctx context.Context, events ...any) error { return nil }

// stubClock satisfies Clock interface.
type stubClock struct{}
func (c *stubClock) Now() time.Time { return time.Time{} }

// stubFreezeInvoker satisfies FreezeInvoker interface.
type stubFreezeInvoker struct{}
func (s *stubFreezeInvoker) Freeze(ctx context.Context, tenantID, instanceID string) error { return nil }

// stubApprovalRepo satisfies repository.ApprovalRepository — returns error on first call.
type stubFailRepo struct{}
func (r *stubFailRepo) GetInstance(ctx context.Context, tenantID, instanceID string) (*domain.ApprovalInstance, error) {
	return nil, errors.New("stub repo error")
}
// Implement all other ApprovalRepository methods as stubs returning zero/nil.
// (Add as needed based on the interface definition.)

func setupDecisionTracerProvider(t *testing.T) (*tracetest.SpanRecorder, func()) {
	t.Helper()
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	return sr, func() { otel.SetTracerProvider(prev) }
}

func TestRecordSignoff_EmitsSignoffRecordSpan(t *testing.T) {
	sr, cleanup := setupDecisionTracerProvider(t)
	defer cleanup()

	svc := application.NewDecisionService(
		&stubFailRepo{},
		&stubEmitter{},
		&stubClock{},
		&stubFreezeInvoker{},
	)

	runner := &stubFailTxRunner{err: errors.New("tx stub")}
	req := application.SignoffRequest{
		TenantID:   "tenant-1",
		InstanceID: "inst-1",
		Decision:   domain.Decision("approved"),
	}
	_, _ = svc.RecordSignoff(context.Background(), runner, req)

	spans := sr.Ended()
	var found *tracetest.SpanStub
	for i := range spans {
		if spans[i].Name() == "signoff.record" {
			found = &spans[i]
			break
		}
	}
	if found == nil {
		names := make([]string, len(spans))
		for i, s := range spans {
			names[i] = s.Name()
		}
		t.Fatalf("want span 'signoff.record', got %v", names)
	}

	attrFound := false
	for _, a := range found.Attributes() {
		if string(a.Key) == "signoff.verdict" && a.Value.AsString() == "approved" {
			attrFound = true
			break
		}
	}
	if !attrFound {
		t.Fatalf("want attribute signoff.verdict=approved, got %v", found.Attributes())
	}
}
```

> **Implementation note:** Read `internal/modules/documents/approval/application/repository/` to find the `ApprovalRepository` interface and implement all methods on `stubFailRepo`. The interface may have 5–15 methods; all can return zero values / nil except `GetInstance` (which returns an error to stop the flow early, before any nil-dep panic). If `DecisionService` uses `runner` (passed as parameter to `RecordSignoff`) to start a tx, the `stubFailTxRunner` will short-circuit the whole flow after the span is created.

- [ ] **Step 4: Run tests — confirm compile failure (TDD red)**

```
go test ./internal/platform/db/postgres/ -run TestOpen_EmitsOTelSpan -count=1
go test ./internal/modules/controlleddocuments/application/ -run TestCreate_EmitsCdCreateSpan -count=1
go test ./internal/modules/documents/approval/application/ -run TestRecordSignoff_EmitsSignoffRecordSpan -count=1
```

Expected: compile errors (`OpenWithTracerProvider undefined`, `undefined: SetSchedulerMetrics` gone; now `cd.create` span not emitted). All three must fail.

- [ ] **Step 5: Commit failing tests**

```
git add internal/platform/db/postgres/connect_otel_test.go \
        internal/modules/controlleddocuments/application/service_otel_test.go \
        internal/modules/documents/approval/application/decision_otel_test.go
git commit -m "test(f2.3): failing OTel span tests — TDD red (connect, cd.create, signoff.record)"
```

---

### Task 2: Add `otelsql` dependency + wire `connect.go`

**Files:**
- Modify: `internal/platform/db/postgres/connect.go`

- [ ] **Step 1: Add `otelsql` dependency**

```
go get github.com/XSAM/otelsql
```

Confirm it appears in `go.mod` and `go.sum`.

- [ ] **Step 2: Add `OpenWithTracerProvider` and update `Open` in `connect.go`**

Current file content to replace:
```go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}
```

Replace with:
```go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/XSAM/otelsql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Open opens a Postgres connection with OTel driver-level tracing using the global TracerProvider.
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	return openWithOptions(ctx, dsn)
}

// OpenWithTracerProvider opens a Postgres connection using the provided TracerProvider.
// Used in tests to inject a tracetest.SpanRecorder.
func OpenWithTracerProvider(dsn string, tp *trace.TracerProvider, _ ...any) (*sql.DB, error) {
	db, err := otelsql.Open("pgx", dsn,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithDBConnectStringAttribute(false),
		otelsql.WithTracerProvider(tp),
	)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	return db, nil
}

func openWithOptions(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := otelsql.Open("pgx", dsn,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithDBConnectStringAttribute(false),
	)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}
```

> **Check `otelsql.WithTracerProvider` signature:** the XSAM/otelsql API may take `oteltrace.TracerProvider` (the interface from `go.opentelemetry.io/otel/trace`) not the concrete `*sdktrace.TracerProvider`. If so, adjust the `OpenWithTracerProvider` signature to accept `oteltrace.TracerProvider`. The test passes `*trace.TracerProvider` (SDK type) which implements the interface. **Check the actual otelsql package API before writing this — run `go doc github.com/XSAM/otelsql.WithTracerProvider` after `go get`.**

- [ ] **Step 3: Run DB span test — confirm green**

```
go test ./internal/platform/db/postgres/ -run TestOpen_EmitsOTelSpan -count=1 -v
```

Expected: `--- PASS: TestOpen_EmitsOTelSpan` (or SKIP if no test DB — either is acceptable; compile must succeed).

- [ ] **Step 4: Confirm `go build ./apps/api/...` still passes**

```
go build ./apps/api/...
```

Expected: exit 0.

- [ ] **Step 5: Commit**

```
git add internal/platform/db/postgres/connect.go go.mod go.sum
git commit -m "feat(db): wire otelsql driver-level OTel tracing in postgres.Open (F2.3)"
```

---

### Task 3: Add `cd.create` span to `ControlledDocumentService.Create`

**Files:**
- Modify: `internal/modules/controlleddocuments/application/service.go`

- [ ] **Step 1: Add imports to `service.go`**

Add to the import block (do not duplicate existing otel imports if any):
```go
"go.opentelemetry.io/otel"
"go.opentelemetry.io/otel/attribute"
"go.opentelemetry.io/otel/codes"
oteltrace "go.opentelemetry.io/otel/trace"
```

- [ ] **Step 2: Wrap `Create` body with `cd.create` span**

Find `func (s *ControlledDocumentService) Create(ctx context.Context, cmd CreateControlledDocumentCmd) (*CreateResult, error) {` (line ~147).

Add at the very start of the function body (before any existing code):
```go
ctx, span := otel.Tracer("metaldocs/controlleddocuments").Start(ctx, "cd.create",
    oteltrace.WithAttributes(attribute.String("document.profile_code", cmd.ProfileCode)),
)
defer span.End()
```

Add error handling: find all `return nil, err` and `return nil, fmt.Errorf(...)` statements in the `Create` function body. Before each, add:
```go
span.RecordError(err)
span.SetStatus(codes.Error, err.Error())
```

> **Surgical change:** only modify the `Create` method. Do not touch other methods. Do not restructure the existing function body.

- [ ] **Step 3: Run CD span tests — confirm green**

```
go test ./internal/modules/controlleddocuments/application/ -run "TestCreate_EmitsCdCreateSpan|TestCreate_SpanStatusError_OnFailure" -count=1 -v
```

Expected: both tests PASS.

- [ ] **Step 4: Commit**

```
git add internal/modules/controlleddocuments/application/service.go
git commit -m "feat(cd): add cd.create OTel span to ControlledDocumentService.Create (F2.3)"
```

---

### Task 4: Add `signoff.record` span to `DecisionService.RecordSignoff`

**Files:**
- Modify: `internal/modules/documents/approval/application/decision_service.go`

- [ ] **Step 1: Add imports to `decision_service.go`**

Add to the import block:
```go
"go.opentelemetry.io/otel"
"go.opentelemetry.io/otel/attribute"
"go.opentelemetry.io/otel/codes"
oteltrace "go.opentelemetry.io/otel/trace"
```

- [ ] **Step 2: Wrap `RecordSignoff` body with `signoff.record` span**

Find `func (s *DecisionService) RecordSignoff(ctx context.Context, runner db.TxRunner, req SignoffRequest) (SignoffResult, error) {` (line ~150).

Add at the very start of the function body:
```go
ctx, span := otel.Tracer("metaldocs/documents/approval").Start(ctx, "signoff.record",
    oteltrace.WithAttributes(attribute.String("signoff.verdict", string(req.Decision))),
)
defer span.End()
```

Add error handling on all `return SignoffResult{}, err` statements in the `RecordSignoff` function body:
```go
span.RecordError(err)
span.SetStatus(codes.Error, err.Error())
```

> **Surgical change:** only modify the `RecordSignoff` method. Do not touch other methods.

- [ ] **Step 3: Run signoff span test — confirm green**

```
go test ./internal/modules/documents/approval/application/ -run TestRecordSignoff_EmitsSignoffRecordSpan -count=1 -v
```

Expected: `--- PASS: TestRecordSignoff_EmitsSignoffRecordSpan`.

- [ ] **Step 4: Commit**

```
git add internal/modules/documents/approval/application/decision_service.go
git commit -m "feat(approval): add signoff.record OTel span to DecisionService.RecordSignoff (F2.3)"
```

---

### Task 5: Update ADR index + link spec

**Files:**
- Modify: `wiki/decisions/index.md` (if it exists and lists ADRs)
- Modify: `wiki/decisions/README.md`
- Modify: `docs/superpowers/milestones/grade-a-completion/milestone-2-observability/f2.3-otel-app-spans/spec.md` — fill ADR link

- [ ] **Step 1: Add ADR 0036 to the decisions index**

Read `wiki/decisions/README.md` and `wiki/decisions/index.md`. Add entry for ADR 0036:
```
0036 — Driver-Level DB Tracing via `otelsql` — Accepted 2026-06-16
```

- [ ] **Step 2: Update spec.md ADR section**

In `spec.md`, find:
```
- [ ] Durable decision — `otelsql` driver-level instrumentation as the standard DB tracing
      pattern for MetalDocs. Record ADR under `wiki/decisions/` and link it here before implementation.
```

Replace with:
```
- [x] Durable decision — recorded as **ADR 0036** [`wiki/decisions/0036-otelsql-db-tracing.md`](../../../../../wiki/decisions/0036-otelsql-db-tracing.md) — Accepted 2026-06-16.
```

- [ ] **Step 3: Commit**

```
git add wiki/decisions/0036-otelsql-db-tracing.md wiki/decisions/README.md wiki/decisions/index.md \
        docs/superpowers/milestones/grade-a-completion/milestone-2-observability/f2.3-otel-app-spans/spec.md
git commit -m "docs(adr): 0036 — driver-level DB tracing via otelsql (F2.3)"
```

---

### Task 6: Regression + runtime proof + commit evidence

**Files:**
- Create: `docs/superpowers/milestones/grade-a-completion/milestone-2-observability/f2.3-otel-app-spans/evidence.md`

- [ ] **Step 1: Whole-repo regression**

```
go test ./...
```

Expected: no FAIL lines.

- [ ] **Step 2: Run 4 named acceptance tests**

```
go test ./internal/platform/db/postgres/ -run TestOpen_EmitsOTelSpan -count=1 -v
go test ./internal/modules/controlleddocuments/application/ -run "TestCreate_EmitsCdCreateSpan|TestCreate_SpanStatusError_OnFailure" -count=1 -v
go test ./internal/modules/documents/approval/application/ -run TestRecordSignoff_EmitsSignoffRecordSpan -count=1 -v
```

Expected: all PASS (or TestOpen_EmitsOTelSpan SKIP if no test DB — note in evidence).

- [ ] **Step 3: Build fresh binary**

```
.\scripts\start-api.ps1 -Build
```

Or if port is occupied: kill old PID, then `go build -o bin\metaldocs-api.exe .\apps\api\cmd\metaldocs-api\`.

- [ ] **Step 4: Start API with console exporter**

```powershell
$env:OTEL_TRACES_EXPORTER = "console"
$env:OTEL_SERVICE_NAME = "metaldocs-api"
# Load other env vars from .env
Get-Content .env | ForEach-Object { if ($_ -match '^([^#=][^=]*)=(.*)$') { [System.Environment]::SetEnvironmentVariable($Matches[1].Trim(), $Matches[2].Trim()) } }
.\bin\metaldocs-api.exe 2>&1 | Tee-Object -FilePath f2.3-smoke.log
```

Wait for `"msg":"MetalDocs API listening"`.

- [ ] **Step 5: Trigger CD create via curl + capture trace output**

```powershell
# 1. Login
$body = '{"identifier":"admin","password":"AdminMetalDocs123!"}'
Invoke-WebRequest -Method Post -Uri "http://localhost:8081/api/v1/auth/login" -ContentType "application/json" -Body $body -SessionVariable "sess" | Out-Null

# 2. Trigger any endpoint that calls ControlledDocumentService.Create or just hit metrics
# For CD create, use the appropriate endpoint. Check the API routes for POST /api/v1/controlled-documents.
# Alternatively, trigger a simpler DB-hitting endpoint (e.g., GET /api/v1/metrics) to verify go.sql.* spans.

# 3. After request, check console output in f2.3-smoke.log for span JSON
Get-Content f2.3-smoke.log | Select-String '"cd.create"\|"signoff.record"\|"go.sql"'
```

- [ ] **Step 6: Capture verbatim span JSON for evidence**

The console exporter emits JSON per span. Capture one snippet showing:
- `cd.create` span OR `go.sql.*` span
- Parent span id present (child of HTTP envelope)

Paste verbatim into evidence.md labeled **real-provider**.

- [ ] **Step 7: Write evidence.md**

Create `docs/superpowers/milestones/grade-a-completion/milestone-2-observability/f2.3-otel-app-spans/evidence.md` following the same structure as `f2.2-scheduler-metrics/evidence.md`:
- What was implemented
- Verification table (command + result + real-vs-fixture label)
- Verbatim console span JSON (**real-provider**)
- Acceptance vs spec Validation Gate table (rows 1–6)
- Review disposition
- Bounded defers (if any)

- [ ] **Step 8: Final commit**

```
git add docs/superpowers/milestones/grade-a-completion/milestone-2-observability/f2.3-otel-app-spans/evidence.md
git commit -m "docs(f2.3): evidence — OTel spans on db + cd.create + signoff.record"
```

---

## Self-review against spec

| Spec requirement | Task that implements it |
|-----------------|------------------------|
| All `database/sql` queries emit child spans (`otelsql`) | Task 2 (`connect.go`) |
| `cd.create` span with `document.profile_code` attribute | Task 3 (`service.go`) |
| `signoff.record` span with `signoff.verdict` attribute | Task 4 (`decision_service.go`) |
| Error path sets span status = Error | Tasks 3 + 4 (RecordError + SetStatus) |
| `otelsql` at composition root, not in repo/service layer | Task 2 (`connect.go` only) |
| OTel inert when `OTEL_TRACES_EXPORTER` unset | No change to `SetupOTel` — inherits existing behavior |
| Existing HTTP envelope spans unmodified | No touch to `otel.go` |
| `TestOpen_EmitsOTelSpan` | Tasks 1 (red) + 2 (green) |
| `TestCreate_EmitsCdCreateSpan` | Tasks 1 (red) + 3 (green) |
| `TestCreate_SpanStatusError_OnFailure` | Tasks 1 (red) + 3 (green) |
| `TestRecordSignoff_EmitsSignoffRecordSpan` | Tasks 1 (red) + 4 (green) |
| `go test ./...` green | Task 6 |
| Runtime proof — console exporter trace with child spans | Task 6 |
| ADR 0036 written before implementation | Task 5 (written before Tasks 2–4) |
