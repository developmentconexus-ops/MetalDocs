# Module #1 — `apps/api/cmd/metaldocs-api`

**Reviewed:** 2026-05-21
**Reviewers:** `ecc:go-reviewer`, `ecc:security-reviewer`, `ecc:silent-failure-hunter`, `ecc:type-design-analyzer`
**Static analysis:** `go vet ./apps/api/cmd/metaldocs-api/...` — clean. `staticcheck` / `govulncheck` not installed.
**Files in scope:**
- `apps/api/cmd/metaldocs-api/main.go` (593 lines)
- `apps/api/cmd/metaldocs-api/permissions.go` (255 lines)
- `apps/api/cmd/metaldocs-api/permissions_test.go` (147 lines)

Findings consolidated across reviewers. Where multiple reviewers flagged the same site, severity reflects the strongest lens (security/data-integrity > maintainability). Each entry cites at least one reviewer.

---

## Critical

### C1. `e2etest.RegisterE2EHandlers` mounted unconditionally
- **File:** `main.go:355`
- **Reviewer:** security-reviewer
- **Issue:** Called with no environment gate. The handlers are not listed in `permissions.go`, so the resolver returns `guarded=false` and `newPublicPathChecker` treats them as fully public. If the binary ships to prod, test-only destructive endpoints are reachable unauthenticated.
- **Recommend:** Gate behind `os.Getenv("METALDOCS_ENV") == "e2etest"` (or build tag). Add a startup assertion + a test that asserts the handlers are absent when the flag is unset.
- **Status:** Fixed in `6eb31ec7` (2026-05-21).
- **Resolution:** Defense-in-depth call-site gate added at [`apps/api/cmd/metaldocs-api/main.go:110-126`](../../../apps/api/cmd/metaldocs-api/main.go) (`e2eHandlersEnabled` + `mountE2EHandlersIfEnabled`). Used the existing `METALDOCS_E2E=1` convention rather than introducing a second env var — `internal/test/e2e_seed.go:81` already keys on the same flag (`grep -nR METALDOCS_E2E` returns 13 files; `METALDOCS_ENV` had zero usages outside the review doc). The call site at `main.go:370-372` now invokes `e2etest.RegisterE2EHandlers` only when the gate passes. `slog.Warn` on mount, `slog.Info` on skip — explicit operator signal in both directions.
- **Verification:** [`apps/api/cmd/metaldocs-api/e2e_gate_test.go`](../../../apps/api/cmd/metaldocs-api/e2e_gate_test.go) — `TestE2EHandlersGate_EnvUnset_DoesNotRegister` builds a fresh mux, calls the gated registrar with `METALDOCS_E2E` unset (via `t.Setenv("METALDOCS_E2E", "")`), asserts the registrar callback is never invoked AND `mux.Handler(req)` returns an empty pattern for all five e2etest paths (`/seed`, `/reset`, `/governance-events`, `/advance-clock`, `/trigger-scheduler-tick`). Three companion tests cover `=1`, whitespace trimming, and truthy-lookalike values (`true`, `yes`, `0`, etc.) — only the literal `"1"` enables the gate. `go vet ./apps/api/cmd/metaldocs-api/... && go test -race -count=1 ./apps/api/cmd/metaldocs-api/...` — clean.
- **Out of scope (deferred to C2):** the underlying structural risk — that any unenumerated route silently defaults to `(guarded=false)` — remains. The C1 gate prevents the specific e2etest exposure; C2's resolver-default flip + visibility enum is the durable fix.

### C2. Permission resolver defaults to fail-open
- **File:** `permissions.go:228-230` (function default `return "", false`) + `permissions.go:233-240` (`newPublicPathChecker` reads `!guarded`)
- **Reviewer:** security-reviewer (corroborated by type-design-analyzer)
- **Issue:** Any route registered on the mux but not enumerated in the resolver is silently treated as public. `RegisterRoutes` calls at `main.go:193-219, 337, 351, 354, 355` outnumber the resolver's matched paths; new routes added without a matching resolver entry bypass both auth and IAM middleware.
- **Recommend:** Flip the default to session-required (`return "", true`) and add an explicit `Public` enum value for genuinely public paths. Force opt-out, not opt-in. See M2 for the structural reshape that makes this easy.

### C3. Server-error shutdown path bypasses deferred cleanup
- **File:** `main.go:451-458, 463-465`
- **Reviewer:** go-reviewer
- **Issue:** `log.Fatalf` inside the `case err := <-serverErr` arm exits the process via `os.Exit(1)` after the `select`, skipping the deferred `stop()`, `shutdownCancel()`, and `defer deps.Cleanup()` at line 139. The scheduler `WaitGroup` is also leaked. Additionally `_ = server.Shutdown(shutdownCtx)` at line 463 discards the shutdown error — graceful-drain failures during deploy go invisible.
- **Recommend:** Move the fatal path into a post-select shutdown sequence that runs the same teardown as the `ctx.Done` arm. Use `errors.Is(err, http.ErrServerClosed)` instead of direct equality. Log shutdown error: `if err := server.Shutdown(shutdownCtx); err != nil { slog.Error("graceful shutdown incomplete", "err", err) }`.
- **Status:** FIXED 2026-05-21 — `shutdownServer` helper extracted; both arms run the same teardown (Shutdown + stop + schedulerWG.Wait + workerWG.Wait); explicit `deps.Cleanup()` before `os.Exit` on the error path. Smoke confirmed `INFO graceful shutdown complete` + `INFO scheduler stopped` on Ctrl-C. Unit tests in `main_test.go`.

### C4. `pdfOutboxWorker` goroutine fails silently, no restart, no shutdown coord
- **File:** `main.go:326-330`
- **Reviewers:** silent-failure-hunter (Critical), go-reviewer (High — no WaitGroup)
- **Issue:** Goroutine logs `err` and exits permanently. Document approval at the API layer continues to succeed while PDF delivery silently stops for the rest of the process lifetime. Worker is also not joined at shutdown — outbox can be half-flushed if process exits mid-write.
- **Recommend:** Add a restart loop with exponential backoff OR expose worker liveness to the health handler so `/api/v1/health/ready` degrades when the worker is down. Add a dedicated `WaitGroup` (matching the scheduler pattern at lines 392-397) and join it before exit.
- **Status:** FIXED 2026-05-21 — `workerWG` mirror of the scheduler pattern + capped-exponential-backoff restart loop (1s → 60s) bounded by `ctx`. Worker now joined at shutdown via `workerWG.Wait()` inside `shutdownServer`. Health-endpoint readiness exposure deferred (separate follow-up).

---

## High

### H1. Audit-retention goroutine has no shutdown coord or escalation
- **File:** `main.go:407-424`
- **Reviewers:** go-reviewer, security-reviewer, silent-failure-hunter
- **Issue:** Not tracked by any `WaitGroup`. `DELETE FROM metaldocs.audit_events` on the API DB role with only `slog.Warn` on failure — repeated failures silently grow the table without bound (compliance risk). No row-count logging on success.
- **Recommend:** Fold into the existing `jobscheduler` (line 358) so it participates in `schedulerWG` and inherits the lease/SkipOnPressure machinery. Log row count: `slog.Info("audit retention purge", "deleted", n)`. Consider an `AUDIT_RETENTION_CONFIRM=true` gate to prevent accidental enablement.

### H2. `log` and `log/slog` mixed in the same binary
- **File:** `main.go:8-9`, used at `main.go:115/119/123/127/131/137/147/153/231/304/308/316/432/443/456` (`log.Fatalf` / `log.Printf`) versus `slog.Default()` / `slog.Error` / `slog.Warn` used elsewhere.
- **Reviewer:** go-reviewer
- **Issue:** Mixed loggers produce inconsistent output formats and prevent structured log routing. The rest of `internal/` is standardized on `log/slog`.
- **Recommend:** Replace remaining `log.Printf`/`log.Fatalf` with `slog.Error(..., "err", err); os.Exit(1)` to match `slog.Default()` already threaded into `migrate.Apply`.

### H3. `jobEnabled` defaults to TRUE
- **File:** `main.go:547-549`
- **Reviewers:** security-reviewer, silent-failure-hunter
- **Issue:** Any value other than `"false"` enables the job — `"flase"`, `"0"`, `"no"`, missing-altogether all activate `stuck-instance-watchdog`, `idempotency-janitor`, `audit-integrity-validator`, `lease-reaper` (some destructive). Least-privilege violation: destructive background ops should be opt-in.
- **Recommend:** Invert default to opt-in (`return strings.EqualFold(strings.TrimSpace(os.Getenv(envName)), "true")`). Validate enum at startup; `slog.Warn` on unrecognised values.

### H4. `documentsAuditAdapter` marshals payload, falls back to literal `"{}"` on error
- **File:** `main.go:501-504`, `main.go:527-530`
- **Reviewer:** silent-failure-hunter
- **Issue:** Audit row persisted with empty payload, silently dropping tenant_id, meta, and action context. Audit trail looks intact (row exists) but carries no usable data — compliance footgun.
- **Recommend:** In `WriteTx` return the marshal error (caller is in a tx and can roll back). In `Write` use `slog.Error` with structured fields (`tenantID`, `actorID`, `action`, `docID`) before falling back — or abort the write entirely.

### H5. `/api/v1/metrics` requires `CapUserManage`
- **File:** `permissions.go:24-26`, route registered at `main.go:403`
- **Reviewer:** security-reviewer
- **Issue:** Metrics route is inside the middleware chain (line 426 wraps the entire mux). Requiring `CapUserManage` means standard scrapers (Prometheus, uptime monitors) cannot scrape — they will be rejected when auth is enabled.
- **Recommend:** Expose metrics on a separate non-public port outside the auth chain, OR map to no capability and restrict by network policy. The current binding surfaces as "monitoring is broken" rather than as a security failure, but the design is wrong either way.

---

## Medium

### M1. `main()` is ~358 lines
- **File:** `main.go:109-466`
- **Reviewer:** go-reviewer
- **Recommend:** Extract at least: (a) lines 113-139 → `loadConfig + buildDeps`, (b) lines 192-354 → `registerRoutes(mux, deps, ...)`, (c) lines 357-398 → `registerScheduledJobs(...)`, (d) lines 437-465 → `startAndWait(ctx, server, &schedulerWG)`. Violates the <50-line guideline in `CLAUDE.md §5.2` / `coding-style.md`.

### M2. Permission resolver is a 200-line flat `if/switch` chain with dual-concern coupling
- **File:** `permissions.go:12-231`
- **Reviewers:** go-reviewer, type-design-analyzer
- **Issue:** The same `(Capability, bool)` tuple is read two different ways: `authMiddleware` consumes `!guarded` (visibility) while `iamMiddleware` consumes `Capability` (permission). A method falling through the `/api/v1/documents` switch (line 151) returns `CapDocumentEdit` to IAM but is treated as public by auth — invisible at the type level.
- **Recommend:** Replace with a `[]routeRule{method, prefix, suffix, cap}` table driving matching from a loop, returning `(Capability, Visibility)` where `Visibility` is a typed enum (`Public | SessionRequired | PermissionGuarded`). Resolver collapses to ~20 lines, missing coverage becomes statically enumerable, and the auth-vs-IAM divergence is impossible.

### M3. `documentsAuditAdapter` nil-receiver guards + hardcoded `TraceID: "trace-local"`
- **File:** `main.go:485-545`
- **Reviewers:** type-design-analyzer, security-reviewer (TraceID part)
- **Issue:** Pointer-receiver nil-checks (`if a == nil || a.writer == nil`) mean a `nil *documentsAuditAdapter` silently satisfies `application.Audit` and no-ops — broken invariant invisible at the type level. `TraceID: "trace-local"` is identical across every audit event; correlation impossible.
- **Recommend:** Mirror `noopFreezeInvoker` pattern with a `noopAudit` type for the disabled case; remove nil-receiver guards. Extract trace ID from `ctx` via the same plumbing already used by httpObs middleware.

### M4. `controlledDocumentDuplicatorAdapter.svc` nil-guard + missing error wrap
- **File:** `main.go:75-107`
- **Reviewers:** type-design-analyzer (nil-guard), go-reviewer (err wrap)
- **Issue:** Runtime `if a.svc == nil` guard at line 80 masks what should be a startup-time wiring failure. `controlledDocumentsModule.Service()` is always non-nil (line 213), so the guard creates false safety. Both error returns at lines 85 and 104 propagate bare `err`, breaking caller's ability to distinguish not-found vs create-failure.
- **Recommend:** Add `newControlledDocumentDuplicatorAdapter(svc)` constructor that panics on nil. Wrap errors: `fmt.Errorf("duplicate controlled document %s: %w", controlledDocumentID, err)`.

### M5. Switch-blocks for templates/documents have no `default` catch-all
- **File:** `permissions.go:93-124` (templates), `permissions.go:125-172` (documents)
- **Reviewer:** security-reviewer
- **Issue:** A future POST sub-route that doesn't match any suffix falls through to the outer resolver default (`"", false`) — silent public exposure (see C2). Structural risk independent of any specific future route.
- **Recommend:** Add an explicit `default` case in each block returning a safe fallback (`CapDocumentView` / `CapTemplateView`) instead of relying on the outer default.

### M6. `METALDOCS_FANOUT_URL` / `METALDOCS_DOCGEN_V2_SERVICE_TOKEN` missing only warns
- **File:** `main.go:228-238`
- **Reviewers:** security-reviewer, silent-failure-hunter
- **Issue:** Deployment with URL set but token unset boots cleanly and fanout silently 401s. URL-unset bypasses approval freeze step (the `noopFreezeInvoker` at line 320-323). Operator misconfiguration discovered only when users finalize a document.
- **Recommend:** Promote token-missing to `log.Fatalf` when `fanoutURL != ""` (combination is always a misconfiguration). URL-missing warn is acceptable only when `METALDOCS_REQUIRE_FANOUT` is explicitly `"false"`.

---

## Low

### L1. Direct equality on sentinel `http.ErrServerClosed`
- **File:** `main.go:453` (`err != http.ErrServerClosed`)
- **Reviewer:** go-reviewer
- **Recommend:** `!errors.Is(err, http.ErrServerClosed)`. Consistent with the rest of `internal/`.

### L2. `documentsAuditAdapter.Write` uses `log.Printf` + duplicates `WriteTx`
- **File:** `main.go:543` (logger), `main.go:493-545` (duplication)
- **Reviewer:** go-reviewer
- **Recommend:** Replace `log.Printf` with `slog.Error` (see H2). Extract `buildEvent(...)` helper to remove the WriteTx/Write duplication.

### L3. `strconv.Atoi` error discarded for `AUDIT_RETENTION_DAYS`
- **File:** `main.go:406`
- **Reviewer:** silent-failure-hunter
- **Issue:** `AUDIT_RETENTION_DAYS=30d` silently disables retention rather than failing loudly.
- **Recommend:** Capture the error and `slog.Warn` when the env var is non-empty but unparseable.

### L4. PATCH negative-suffix guard for IAM user update
- **File:** `permissions.go:51`
- **Reviewer:** security-reviewer
- **Recommend:** Replace `!strings.HasSuffix(path, "/roles")` with explicit path structure. Negative guards are fragile against new sub-paths.

### L5. `permissions_test.go` mixes table-driven and standalone tests
- **File:** `permissions_test.go:75-94, 126-146`
- **Reviewer:** go-reviewer
- **Recommend:** Fold into the table at `TestPermissionResolver`. File already uses `t.Parallel()` correctly elsewhere.

### L6. `realClock` / `realUUIDGen` duplicate concepts already exported by `approvalapp`
- **File:** `main.go:468, 481`
- **Reviewer:** type-design-analyzer
- **Issue:** `approvalapp.RealClock` is exported and used at lines 301/332; `realClock` in `main` exists only for `templatesapp`. Pure surface duplication.
- **Recommend:** Promote `realClock` and `realUUIDGen` to a shared `internal/platform/clock` package (or reuse `approvalapp.RealClock` directly if dependency direction allows).

---

## Summary

| Severity | Count |
|----------|-------|
| Critical | 4     |
| High     | 5     |
| Medium   | 6     |
| Low      | 6     |
| **Total** | **21** |

**Verdict:** **BLOCK merge of any new prod deploy from this branch until C1-C4 are remediated.** (C1 fixed in `6eb31ec7`; C2-C4 outstanding.)

- **C1 + C2** combine into the single highest-priority security thesis: *any new route on this codebase is silently public unless the operator remembers two separate places (route registration + resolver entry).* Fix C2 first — the structural reshape (M2) makes C1 a one-line `e2etest.IsEnabled()` check rather than a defensive scattering.
- **C3 + C4** are correctness gaps in process lifetime — they will not corrupt data in steady state but they corrupt failure modes (crashes look clean, workers die invisibly).
- **H1-H5** are pre-prod-discipline gaps: each one is recoverable but each one reduces the operator's ability to diagnose a misbehaving deployment.

**Action items spawned:** none auto-spawned during review. Per the initiative's severity convention, Critical findings should each get a fix task. Recommend opening four fix branches off this commit, one per Critical, with the corresponding M-tier reshape bundled where it reduces total churn (M2 bundled with C2; M3 bundled with H4).
