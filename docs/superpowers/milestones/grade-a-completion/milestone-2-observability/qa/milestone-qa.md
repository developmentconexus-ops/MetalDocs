# Milestone 2 — Validation Verdict (C1–C7) · RE-RUN after f2.5

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-16 (RE-RUN)  ·  **Prior run:** FAIL (C1/C2/C3/C6 on F2.3 span-name contract).
> **Verdict:** see C7 — **PASS**.
> The validator judges and writes this file; the **main session flips status only on a PASS**. The
> validator never edits code, fixes findings, or flips status.

## What this re-run validates

This is the close-out re-run after fix feature **f2.5-otelsql-span-name-truth** (commit `eed1d293`).
The prior FAIL was a single defect family: F2.3 declared the **wrong DB driver span names**
(`go.sql.*`, belonging to a *different* otelsql library) in its binding contract, ADR 0036, and
test — while the vendored library `github.com/XSAM/otelsql` emits `sql.*` names. The row-1 test
"passed" only by SKIP, masking a guaranteed assertion failure. F2.1, F2.2, F2.4 were clean.

## Inputs loaded

- Milestone spec `../milestone.md` — read.
- F2.1–F2.4 `spec.md` / `plan.md` / `evidence.md` — read.
- Vendored library `vendor/github.com/XSAM/otelsql/methods.go:23-48` — read (the contract source of truth).
- Production seam `internal/platform/db/postgres/connect.go` + test `connect_otel_test.go` — read.
- ADR `wiki/decisions/0036-otelsql-db-tracing.md` — read.
- Fix commit `eed1d293` (scope: spec.md, ADR 0036, connect_otel_test.go, evidence.md) — reviewed.

All inputs present and readable — no fail-fast.

## C1 — Spec & plan conformance (per feature) — **PASS**

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F2.1 scheduler-slog | ✅ | ✅ | ✅ | Re-run still passing (see C2). No change since prior PASS. |
| F2.2 scheduler-metrics | ✅ | ✅ | ✅ | Re-run still passing (see C2). No change since prior PASS. |
| F2.3 otel-app-spans | ✅ **(fixed)** | ✅ | ✅ | **Contract now correct.** spec.md §1 (lines 32-35) declares `sql.connector.connect` / `sql.conn.ping` / `sql.conn.query` / `sql.conn.exec` / `sql.conn.prepare` / `sql.conn.begin_tx` / `sql.conn.reset_session` / `sql.rows`, citing `vendor/.../methods.go:25-47`. Cross-checked against the actual library `methods.go:24-47` — exact match (`MethodConnectorConnect = "sql.connector.connect"`, `MethodConnQuery = "sql.conn.query"`, …). Validation-Gate rows 1 & 6 (lines 109, 114) corrected to `sql.*` / `sql.connector.connect` / `sql.conn.exec`. `cd.create` / `signoff.record` semantic spans unchanged and honored. |
| F2.4 metrics-completeness | ✅ | ✅ | ✅ | Re-run still passing (see C2). No change since prior PASS. |

**C1 result: PASS** — F2.3's binding consumer contract (spec §1 + Validation Gate) now matches the
producer (`otelsql.Open` in `connect.go`) and the vendored library it actually uses. The guessed
contract that drove the prior FAIL is resolved.

*Non-blocking finding (doc hygiene, not a contract defect):* two stale references to the old fact
remain in **non-binding** sections and should be cleaned by the main session:
(a) `f2.3-otel-app-spans/plan.md:72,82,671,680` — the pre-code TDD-red test snippet and a runtime
note still say `go.sql.*` (superseded by the actual test, which the plan's own line 87 flags as a
draft);
(b) `f2.3-otel-app-spans/spec.md:65-70` ("What this feature implements") still names the wrong
package `go.opentelemetry.io/contrib/instrumentation/database/sql/otelsql` and the `otelsql.Register`/`pgx-otel`
API (hedged with "or equivalent pattern"), whereas the code uses `github.com/XSAM/otelsql` with
`otelsql.Open`. These are historical/sketch sections, not the contract C1 judges against; they do
not contradict the **governing** contract or the ADR, so they are not a live split-brain — but
leaving them invites future drift. Recommend a doc-only cleanup (no new feature required).

## C2 — Gates re-run, isolated — **PASS**

| Feature | Command re-run (fresh, clean) | Real output | Pass? |
|---------|-------------------------------|-------------|-------|
| F2.3 (row 1, **the fixed gate**) | `go test ./internal/platform/db/postgres/ -run TestOpen_EmitsOTelSpan -count=1 -v` | `=== RUN   TestOpen_EmitsOTelSpan` → `--- PASS: TestOpen_EmitsOTelSpan (0.00s)`; `ok metaldocs/internal/platform/db/postgres 0.307s` | ✅ **PASS, not SKIP** |
| F2.3 (rows 2–4) | `go test ./internal/modules/controlleddocuments/application/ -run 'TestCreate_EmitsCdCreateSpan\|TestCreate_SpanStatusError_OnFailure'` + `…/documents/approval/application/ -run TestRecordSignoff_EmitsSignoffRecordSpan` | `ok …/controlleddocuments/application 1.963s`; `ok …/documents/approval/application 2.064s` | ✅ |
| F2.1 | `go test ./internal/modules/jobs/scheduler/ -run 'TestScheduler_LoggerEmitsJSON\|TestScheduler_New_RejectsNilLogger\|TestNew_LeaderIDRequired'` (bundled w/ F2.2) | `ok …/jobs/scheduler 1.230s` | ✅ |
| F2.2 | `go test ./internal/modules/jobs/scheduler/ -run …SchedulerMetrics_GroupedByJob` + `go test ./internal/platform/observability/` | `ok …/jobs/scheduler 1.230s`; `ok …/platform/observability 3.294s` | ✅ |
| F2.4 | `go test ./internal/platform/observability/` (db_pool + adapter tests) | `ok …/platform/observability 3.294s` | ✅ |
| all | `go build ./...` | `BUILD_EXIT=0` | ✅ |

**C2 result: PASS** — the row-1 gate that previously passed only by SKIP now executes against the
real vendored library (unreachable DSN `127.0.0.1:9999` + `PingContext` triggers
`sql.connector.connect` before the dial fails) and **asserts a `sql.*` prefix** — it can now fail,
and it passes. The green-by-skip is eliminated.

## C3 — Senior review of the aggregate milestone diff — **PASS**

Reviewed the f2.5 fix diff (`eed1d293`) and re-checked the four sites the prior FAIL named as the
split-brain:

- `spec.md:33` contract §1 — now `sql.connector.connect` / `sql.conn.*`. ✅
- `spec.md:109,114` Validation-Gate rows 1 & 6 — now `sql.*` / `sql.connector.connect` / `sql.conn.exec`. ✅
- `wiki/decisions/0036-otelsql-db-tracing.md:30,40` — now `sql.connector.connect` / `sql.conn.query` / `sql.conn.exec`; `go.sql.*` fully removed. ✅
- `connect_otel_test.go:46` — now asserts `strings.HasPrefix(s.Name(), "sql.")`, no skip path. ✅

These four — the **binding contract, the ADR, the test, and the evidence** — now agree with each
other, with the production code (`connect.go:45` `otelsql.Open("pgx", …)`), and with the
real-provider capture in `evidence.md:25,29` (`"sql.connector.connect"`, `"sql.conn.exec"`,
`InstrumentationScope.Name = "github.com/XSAM/otelsql"`). The governing source-of-truth split-brain
is resolved. The fix touched only docs + one test; no production behavior changed (instrumentation
was always correct). Diff is in-scope, no dead code, staff-engineer bar met for the corrected
artifacts. (Residual stale text in plan.md / spec-implementation-sketch noted under C1 as
non-blocking doc hygiene.)

**C3 result: PASS.**

## C4 — Workflow-class QA + regression — **PASS**

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (`backend-api-qa-checklist`, observability lens) | pass | Composition-root injection truth-table intact (logger F2.1; scheduler metrics + db_pool F2.2/F2.4; otelsql driver wrapper + global tracer F2.3). The F2.3 contract defect that downgraded this to pass-with-defect last run is cleared. `/api/v1/metrics` payload still purely additive. |
| Regression vs prior milestones (M0 auth/authz/session, M1 contract) | all pass | `go test ./...` → `GO_EXIT=0`, **85 packages `ok`, zero FAIL/panic** (remaining `?` lines are no-test-file packages). `go build ./...` exit 0. No payload-key rename/removal, no auth/route regression. |

**C4 result: PASS** — regression clean across the whole repo; no prior milestone regressed.

## C5 — Quality-bar re-measure + retrospective — **PASS**

| Bar / class | After f2.5 | Root-cause-fixed evidence |
|-------------|-----------|---------------------------|
| D1 hardcoded text logger | injected `slog.Default()` JSON | unchanged from prior PASS. ✅ |
| D2 job metrics not exposed | `scheduler.jobs.*` on `/api/v1/metrics` | unchanged from prior PASS. ✅ |
| D3 no app-level spans | DB spans (`sql.*`) + `cd.create`/`signoff.record` | spans exist at runtime AND the declared contract now names them correctly. Root cause fixed at the `otelsql.Open` seam; the documentation/test correctness defect (the prior ⚠) is now resolved at the true source — the wrong-library guess was corrected, not symptom-patched. ✅ |
| D4 pool stats absent / misleading comment | `db_pool` key; comment corrected | unchanged from prior PASS. ✅ |

- **Symptom-patch check:** the fix did NOT weaken the test to make it pass against a wrong name, and
  did NOT change production code to match a guessed contract. It corrected the *documents and test*
  to the *real library's emitted names* (read from `methods.go`) — root cause (a guessed contract
  against the wrong library) eliminated. This is a root-cause fix.
- **Could it be built better?** The minor residual: the f2.5 fix corrected the binding contract, ADR,
  test, and evidence but left two non-binding sketch/historical sections (plan.md, spec
  implementation-sketch) carrying the stale `go.sql.*` / wrong-package text. A tidy close would have
  swept those too. Recommend a doc-only cleanup; not a milestone blocker.

## C6 — Forbidden-list (any hit = FAIL) — **clean**

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — (each row mapped to a named command + real output above).
- [ ] Fixture/mock passed off as real-provider proof — (F2.3 real-provider artifacts honestly labeled; fixture rows labeled fixture).
- [ ] Consumer contract guessed rather than read from the consumer — **RESOLVED.** Contract now matches `vendor/.../XSAM/otelsql/methods.go:24-47`, the actual library. (Prior hit.)
- [ ] Split-brain (one fact, two sources of truth) — **RESOLVED** in the governing artifacts (contract + ADR + test + evidence now agree). Non-binding stale text in plan.md / spec sketch noted as doc hygiene under C1, not a live governing-source split-brain.
- [ ] Self-judged close / validator edited or fixed code — (validator wrote only this file).
- [ ] Scope drift — (f2.5 diff is docs + one test, in scope of the named fix feature).
- [ ] Symptom-patch — **RESOLVED.** Root cause fixed (corrected to real names); test now genuinely fails-able and green. (Prior green-by-skip hit eliminated.)

No forbidden-list hits.

## C7 — Verdict

- **VERDICT: PASS**
- **Prior failed checks now cleared:** C1 (F2.3 contract honored — matches the real vendored
  library), C2 (row-1 gate executes and passes, no longer SKIP-masked), C3 (governing split-brain
  resolved across spec contract / ADR / test / evidence), C6 (guessed-contract + split-brain +
  green-by-skip all resolved).
- **All other gates re-run green:** F2.1, F2.2, F2.4 named tests pass from clean state; `go build ./...`
  exit 0; whole-repo `go test ./...` exit 0 (85 ok, zero FAIL/panic) — no regression of M0/M1.
- **Non-blocking defer (recommend doc-only cleanup, NOT a new fix feature):** sweep the residual
  `go.sql.*` / wrong-package references from `f2.3-otel-app-spans/plan.md:72,82,671,680` and the
  `spec.md:65-70` "What this feature implements" sketch so all F2.3 docs name `github.com/XSAM/otelsql`
  + `sql.*` consistently. These are historical/sketch sections, not the binding contract, and do not
  block the milestone.
- Milestone may advance. The **main session** flips status (README + roadmap) and presents the HS-1
  operator gate — the validator does not.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending main-session action.
> - Status flip in `README.md` / roadmap: main session, on this PASS.
