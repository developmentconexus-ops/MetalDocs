# Tech Debt Register — audit

> Companion to `wiki/modules/audit.md`. Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in `wiki/backlog/audit-refactor.md`.

**Last verified:** 2026-06-11 (adversarial verification pass 2: T-005 anchor corrected to admin_handler.go:403 + main.go:816; T-006 anchor corrected to admin_handler.go:404, claim updated from timestamp-based to uuid.NewString(); T-003/T-010 migration paths corrected to archive/migrations/; T-010 payload anchor corrected to admin_handler.go:403-414 + main.go:816-828; T-007 wording corrected — tenant_id filter is unconditional; prior: T-001/T-005/T-007 closed-item observations labeled (original); T-008 anchor corrected to openapi.yaml:741-745; 2026-06-10 Stage-1 drift patch T-009)

## Severity scale

Triggers per `.claude/skills/metaldocs-module-doc/templates/tech-debt-register.md`.

- **Critical** — authn/authz bypass, regulated audit-trail tampering path, multi-tenant data leak, data-loss path, contract violation downstream consumers rely on, schema/version drift the boot-check misses.
- **Major** — defense-in-depth gap, governance sink wired to `nil` on regulated path, duplicated write surfaces with divergent semantics, documented contract not followed (measurable consumer impact), cross-module dep blocking another module's refactor.
- **Minor** — symbol-naming collision, missing doc comments, latent surface (no caller hits it today), bidirectional non-circular dep, missing standalone ADR for a rule already enforced by code + tests.

Pick highest trigger. Justify the call in `Observation`.

**Cross-module note (per user-supplied rubric):** missing audit-emission on a regulated mutation is **Major** in *this* register when the gap is consumer-side (audit module accepted the contract — caller did not call it). It is **Critical** in *this* register only if audit itself drops events that callers correctly emitted. T-005 is rated under this rule.

## Items

### T-001 · Unauthenticated `GET /api/v1/audit/events` — CLOSED 2026-05-11 (Plan 6a)
- **Severity:** critical (closed)
- **Surface (resolved):** `internal/modules/audit/delivery/http/handler.go:67-68` (route registration); `apps/api/cmd/metaldocs-api/permissions.go:232` (audit routeRules — `CapAuditRead` binding added at line 232 for `GET /api/v1/audit/events`).
- **Observation (original):** The route was mounted via `mux.HandleFunc("/api/v1/audit/events", h.handleEvents)` with zero auth middleware. The capability resolver had no rule for the path; the public-path checker therefore admitted the request as public. Any network-reachable client could read up to 200 audit rows per call, filtered by `resource_type`/`resource_id`. Confidentiality breach + tampering reconnaissance vector. Trigger fired: authn/authz bypass.
- **Evidence:** `_artifacts/02-flow-list.md` §1, §6; `_artifacts/05-industry.md` IP-004.
- **Linked backlog row:** `backlog/audit-refactor.md#R-001`
- **Linked ADR:** missing-ADR

### T-002 · Legacy error envelope (RFC 9457 drift) — CLOSED 2026-05-12 (Plan 7)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/audit/delivery/http/handler.go:88-101` — local `requestTraceID`/`writeAPIError` helpers deleted; all error paths now call `writeProblem(w, problem.New(...))` which delegates to `problem.Write`. `application/problem+json` is emitted on all 4xx/5xx paths.
- **Observation (original):** Handler emitted `{"error":{"code","message","details","trace_id"}}` instead of `application/problem+json`. Mirrored drift in auth T-003, iam T-006, documents T-001.
- **Evidence:** `_artifacts/02-flow-list.md` §5; `_artifacts/05-industry.md` IP-001.
- **Linked backlog row:** `backlog/audit-refactor.md#R-002` (merged Plan 7 2026-05-11, commit `2ca727d6`)
- **Linked ADR:** `wiki/architecture/api-design-system.md`

### T-003 · No retention or purge policy — CLOSED 2026-05-11 (Plan 6a, app-level goroutine)
- **Severity:** major (closed)
- **Surface:** `archive/migrations/0004_init_audit_events.sql`; `archive/migrations/0005_grant_workflow_audit_privileges.sql` (no follow-up retention migration); module code (no purge job, no partition, no TTL)
- **Observation:** `metaldocs.audit_events` grows monotonically. No partitioning, no `pg_cron` job, no soft-delete, no archive offload. Regulated data has both retention-floor (ISO 9001 §7.5.3 record retention) and retention-ceiling (LGPD/GDPR right-to-erasure for personal data inside `payload`) obligations. Without a retention strategy, both ends fail: old records cannot be selectively purged when erasure is requested; storage growth is unbounded. Trigger fired: governance/compliance gap on a regulated path with measurable consumer impact (legal).
- **Evidence:** `_artifacts/04-persistence.md` §6 ("none — table grows monotonically"); `_artifacts/05-industry.md` "Patterns deliberately NOT cited" — retention.
- **Linked backlog row:** `backlog/audit-refactor.md#R-003`
- **Linked ADR:** missing-ADR

### T-004 · No tamper-evidence on the audit trail — CLOSED 2026-05-13 (audit T-004 follow-up)
- **Severity:** critical (closed)
- **Surface (resolved):** `migrations/0193_audit_events_hash_chain.sql` adds `audit_sequence`, `prev_hash`, `row_hash`, backfills existing rows, and defines `metaldocs.audit_event_row_hash(...)`; `internal/modules/audit/infrastructure/postgres/writer.go` now serializes audit inserts with a transaction-scoped advisory lock and writes the row-hash chain; `internal/modules/jobs/audit_integrity_validator/job.go` exposes the integrity validation job.
- **Observation (original):** Append-only is enforced at the application role (`metaldocs_app` has only `INSERT`). The schema owner role (used by migrations) and any Postgres superuser retain full `UPDATE/DELETE` on `metaldocs.audit_events`. Before this follow-up there was no hash chain (`prev_hash`/`row_hash`) and no integrity-validation job, so a privileged actor could rewrite history undetected. Trigger fired: regulated audit-trail tampering path (user-supplied rubric: "audit-trail tampering path = Critical").
- **Evidence:** `_artifacts/04-persistence.md` §1, §3, §6; `_artifacts/05-industry.md` "Patterns deliberately NOT cited" — tamper-evidence.
- **Linked backlog row:** `backlog/audit-refactor.md#R-004`
- **Linked ADR:** missing-ADR

### T-005 · Fire-and-forget Record discards emission errors — CLOSED 2026-05-11 (Plan 6a)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/iam/delivery/http/admin_handler.go:403` (audit.Record call); `apps/api/cmd/metaldocs-api/main.go:816` (documentsAuditAdapter.Record call) — error surfacing work merged per `backlog/audit-refactor.md#R-005`.
- **Observation (original):** All consumer call sites discarded or only logged the Record error (`_ = h.audit.Record(...)`). There was no failure metric, no dead-letter queue, no alarm, no retry. On Postgres unavailability or PK collision (see T-006), the regulated action persisted but the audit row was silently lost. Operator could not detect dropped trail entries. Per user-supplied rubric, rated **Major** here because the drop happened *consumer-side* (audit module's port contract correctly returned the error — callers ignored it); the Critical-rated mirror lives in the consumer registers. Trigger fired: defense-in-depth + observability sink gap on regulated path.
- **Evidence:** `_artifacts/02-flow-record.md` §6(c); `_artifacts/03-deps.md` §2b, §2c.
- **Linked backlog row:** `backlog/audit-refactor.md#R-005`
- **Linked ADR:** missing-ADR

### T-006 · App-side timestamp-based event ID is collision-prone
- **Severity:** minor
- **Surface:** `internal/modules/iam/delivery/http/admin_handler.go:404` — `"evt_" + uuid.NewString()`; `apps/api/cmd/metaldocs-api/main.go:761` (`bypassAuditAdapter`, same `"evt_" + uuid.NewString()` pattern); `apps/api/cmd/metaldocs-api/main.go:817` (`documentsAuditAdapter`, bare `uuid.NewString()` without `"evt_"` prefix)
- **Observation:** Event ID generation is not uniform across call sites. The IAM handler and bypass adapter prefix the UUID with `"evt_"` (`admin_handler.go:404`, `main.go:761`), while the documents adapter emits a bare UUID with no prefix (`main.go:817`). This inconsistency means the `id` column holds heterogeneous formats and consumers cannot reliably parse the prefix as a stable type marker. A prior version of this item documented a timestamp-based ID scheme (`"evt_" + strings.ReplaceAll(time.Now()...)`); that scheme no longer exists — all sites now use `uuid.NewString()`. The residual debt is the prefix inconsistency, not collision risk. Trigger fired: latent (surface exists; no caller relies on the prefix today, but it is part of the implicit event-ID contract).
- **Evidence:** `_artifacts/04-persistence.md` §1 (event id generation fact).
- **Linked backlog row:** `backlog/audit-refactor.md#R-006`
- **Linked ADR:** missing-ADR

### T-007 · No `tenant_id` column or tenant-scoped query path — CLOSED 2026-05-11 (Plan 6a)
- **Severity:** major (closed)
- **Surface (resolved):** `archive/migrations/0190_audit_events_tenant_id.sql:4-8` adds `tenant_id TEXT NOT NULL DEFAULT ''` column and index; `archive/migrations/0193_audit_events_hash_chain.sql:60` includes `tenant_id` in hash-chain CTE; `internal/modules/audit/infrastructure/postgres/writer.go:232` (`buildWhere` unconditionally adds `tenant_id = $N` as the first filter clause).
- **Observation (original):** `metaldocs.audit_events` had no `tenant_id` column. `ListEvents` filtered by `resource_type` and `resource_id` only. When multi-tenant lands (auth T-008 plans `tenant_id` on identity tables), a tenant-A admin reading `/api/v1/audit/events` would see Tenant-B events. Latent at time of writing (single-tenant deploy); load-bearing on multi-tenant cutover. Trigger fired: multi-tenant data leak path (latent — single-tenant today, but the leak path was in code). Note: the resolved `buildWhere` at `writer.go:232` applies the `tenant_id` filter unconditionally (not guarded by a non-empty check), consistent with `TenantID` being mandatory at the application layer per `internal/modules/audit/domain/port.go:69-70`.
- **Evidence:** `_artifacts/04-persistence.md` §1; `_artifacts/05-industry.md` IP-008.
- **Linked backlog row:** `backlog/audit-refactor.md#R-007`
- **Linked ADR:** missing-ADR

### T-008 · Audit route served outside oapi-codegen (manual `HandleFunc`, not generated handler)
- **Severity:** minor
- **Surface:** `api/openapi/v1/openapi.yaml:741-745` (path `/audit/events`, `operationId: listAuditEvents` — present); `internal/modules/audit/delivery/http/handler.go:67-68` (route mounted via `http.ServeMux.HandleFunc` directly, not via oapi-codegen generated handler)
- **Observation:** The `operationId` `listAuditEvents` exists in the spec at `openapi.yaml:745`. The drift is not a missing `operationId` but a wiring gap: the handler is registered manually via `HandleFunc` rather than through the oapi-codegen `StrictServerInterface` path used by other modules. Client codegen can bind a method name from the spec, but the server-side implementation is not generated/validated by oapi-codegen. This means request/response type checking at the codegen boundary is absent for this route. Trigger fired: contract surface gap — spec and implementation diverge in binding mechanism.
- **Evidence:** `_artifacts/02-flow-list.md` §1.
- **Linked backlog row:** `backlog/audit-refactor.md#R-008`
- **Linked ADR:** `wiki/decisions/0012-contract-first-api.md` (audit was never migrated; flagged here as the residual)

### T-009 · No explicit `SELECT` grant on `metaldocs.audit_events` — CLOSED 2026-06-10 (Stage-1 drift patch)
- **Severity:** minor (closed)
- **Surface (resolved):** `migrations/0193_audit_events_hash_chain.sql:110` — `GRANT INSERT, SELECT ON TABLE metaldocs.audit_events TO metaldocs_app`. The SELECT grant exists in the archived migration ledger; it was added together with the hash-chain extension in archived migration 0193.
- **Observation (original):** Written before migration 0193 was applied. At that point only `migrations/0005_grant_workflow_audit_privileges.sql:2` (INSERT only) existed. 0193 upgraded the grant to `INSERT, SELECT`. The item was not closed when 0193 landed. Stage-1 backend audit confirmed the grant at `archive/migrations/0193_audit_events_hash_chain.sql:110`. Note: `migrations/0224_audit_export_jobs_pr6.sql` creates `metaldocs.audit_export_jobs` without an explicit GRANT (tracked as flag F-07 in the Stage-1 artifact).
- **Evidence:** `wiki/backend/_artifacts/stage1/module-audit.md` §6 ("Grants: GRANT INSERT, SELECT…added in archived 0193").
- **Linked backlog row:** `backlog/audit-refactor.md#R-009`
- **Linked ADR:** missing-ADR

### T-010 · No payload size constraint
- **Severity:** minor
- **Surface:** `archive/migrations/0004_init_audit_events.sql:8` (`payload JSONB NOT NULL DEFAULT '{}'::jsonb`); `internal/modules/iam/delivery/http/admin_handler.go:403-414` and `apps/api/cmd/metaldocs-api/main.go:816-828` (payload marshalled with no size check before passing to `audit.Record`)
- **Observation:** `payload JSONB` has no `CHECK (octet_length(payload::text) < N)` constraint and no application-side size cap. A misbehaving consumer can emit unbounded payloads (e.g. dumping a full document body). Today consumer payloads are small (`map[string]any{}`-shaped), but the surface is open. Trigger fired: latent (surface exists; no caller hits it today).
- **Evidence:** `_artifacts/04-persistence.md` §1 ("Payload size constraint fact").
- **Linked backlog row:** `backlog/audit-refactor.md#R-010`
- **Linked ADR:** missing-ADR

### T-011 · Missing-ADR for append-only-by-grant + port-and-adapter shape
- **Severity:** minor
- **Surface:** module shape itself — `internal/modules/audit/{domain,application,delivery,infrastructure}/`; `migrations/0005:2` (grant strategy)
- **Observation:** Three load-bearing decisions are encoded in code with no ADR backing: (a) `Writer` and `Reader` as Go ports, with no transaction-accepting variant; (b) append-only enforced by `GRANT INSERT` only, not by trigger or RLS; (c) the same `*postgres.Writer` satisfies both ports. Each is defensible; none is documented. Future refactors (tamper-evidence, tenant scoping, tx-bundled emit) need an ADR to evaluate against. Trigger fired: missing standalone ADR for rules already enforced by code.
- **Evidence:** `_artifacts/03-deps.md` §3 (DI wiring confirms same instance into both slots); `_artifacts/04-persistence.md` §3 (no triggers/RLS).
- **Linked backlog row:** `backlog/audit-refactor.md#R-011`
- **Linked ADR:** missing-ADR

### T-012 · No Go doc comments on exported symbols
- **Severity:** minor
- **Surface:** every file under `internal/modules/audit/` (all exported symbols undocumented per Phase 1 surface scan — all entries marked `(undocumented)` implicitly: only signatures were extracted, no doc comments)
- **Observation:** `Event`, `ListEventsQuery`, `Writer`, `Reader`, `Service`, `Handler`, `EventResponse`, `postgres.Writer`, `memory.Writer`, and all methods have zero Go doc comments. `go doc` returns signatures only. Reduces discoverability and breaks lint rules other modules apply (`golint` exported-doc rule). Trigger fired: missing doc comments on exported symbols.
- **Evidence:** `_artifacts/01-surface.md` §2 (Doc-comment column empty across the board).
- **Linked backlog row:** `backlog/audit-refactor.md#R-012`
- **Linked ADR:** missing-ADR

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: 12 / 12 (all exported symbols lack doc comments — see T-012)
- Operations missing C4 placement: 0 / 1 (`GET /api/v1/audit/events` is in §5.3 + §6.2)
- Cross-deps missing in §5/§8: 0 / 3 (bootstrap, iam admin handler, documents adapter — all referenced)
- State transitions missing in §6: 0 / 0 (n/a — append-only sink)
- Decisions without ADR link: 11 / 12 (T-001..T-007, T-009..T-012 = 11 missing-ADR; T-008 links ADR 0012 as residual)
