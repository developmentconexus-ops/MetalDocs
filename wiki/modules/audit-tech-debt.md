# Tech Debt Register — audit

> Companion to `wiki/modules/audit.md`. Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in `wiki/backlog/audit-refactor.md`.

**Last verified:** 2026-05-11 (Plan 6a)

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
- **Surface:** `internal/modules/audit/delivery/http/handler.go:34-35` (route registration); `apps/api/cmd/metaldocs-api/permissions.go:211-221` (resolver default + public-path checker treat unregistered paths as public)
- **Observation:** The route is mounted via `mux.HandleFunc("/api/v1/audit/events", h.handleEvents)` with zero auth middleware. The capability resolver has no rule for the path; the public-path checker therefore admits the request as public. Verified by grep — no `audit/events` rule in `permissions.go`. Any network-reachable client can read up to 200 audit rows per call, filtered by `resource_type`/`resource_id`. Confidentiality breach + tampering reconnaissance vector. Trigger fired: authn/authz bypass.
- **Evidence:** `_artifacts/02-flow-list.md` §1, §6; `_artifacts/05-industry.md` IP-004.
- **Linked backlog row:** `backlog/audit-refactor.md#R-001`
- **Linked ADR:** missing-ADR

### T-002 · Legacy error envelope (RFC 9457 drift)
- **Severity:** major
- **Surface:** `internal/modules/audit/delivery/http/handler.go:48,60,97-105`
- **Observation:** Handler emits `{"error":{"code","message","details","trace_id"}}` instead of `application/problem+json` per RFC 9457. Mirrors documented drift in auth T-003, iam T-006, documents T-001. Consumer tooling that relies on `type`/`title`/`status`/`detail`/`instance` cannot parse audit errors uniformly. Trigger fired: contract violation with measurable consumer impact.
- **Evidence:** `_artifacts/02-flow-list.md` §5; `_artifacts/05-industry.md` IP-001.
- **Linked backlog row:** `backlog/audit-refactor.md#R-002`
- **Linked ADR:** missing-ADR

### T-003 · No retention or purge policy — CLOSED 2026-05-11 (Plan 6a, app-level goroutine)
- **Severity:** major (closed)
- **Surface:** `migrations/0004_init_audit_events.sql`; `migrations/0005_grant_workflow_audit_privileges.sql` (no follow-up retention migration); module code (no purge job, no partition, no TTL)
- **Observation:** `metaldocs.audit_events` grows monotonically. No partitioning, no `pg_cron` job, no soft-delete, no archive offload. Regulated data has both retention-floor (ISO 9001 §7.5.3 record retention) and retention-ceiling (LGPD/GDPR right-to-erasure for personal data inside `payload`) obligations. Without a retention strategy, both ends fail: old records cannot be selectively purged when erasure is requested; storage growth is unbounded. Trigger fired: governance/compliance gap on a regulated path with measurable consumer impact (legal).
- **Evidence:** `_artifacts/04-persistence.md` §6 ("none — table grows monotonically"); `_artifacts/05-industry.md` "Patterns deliberately NOT cited" — retention.
- **Linked backlog row:** `backlog/audit-refactor.md#R-003`
- **Linked ADR:** missing-ADR

### T-004 · No tamper-evidence on the audit trail
- **Severity:** critical
- **Surface:** `migrations/0004_init_audit_events.sql:1-14` (no hash column); `migrations/0005_grant_workflow_audit_privileges.sql:2` (grant is INSERT only — but schema owner / superuser retain UPDATE/DELETE); `internal/modules/audit/infrastructure/postgres/writer.go:20-42` (no signing, no hash chaining)
- **Observation:** Append-only is enforced at the application role (`metaldocs_app` has only `INSERT`). The schema owner role (used by migrations) and any Postgres superuser retain full `UPDATE/DELETE` on `metaldocs.audit_events`. There is no hash chain (`prev_hash`/`row_hash`), no row signing (no asymmetric signature column), no external WORM mirror, no Merkle root, and no integrity-validation job. A privileged actor can rewrite history undetected. Trigger fired: regulated audit-trail tampering path (user-supplied rubric: "audit-trail tampering path = Critical").
- **Evidence:** `_artifacts/04-persistence.md` §1, §3, §6; `_artifacts/05-industry.md` "Patterns deliberately NOT cited" — tamper-evidence.
- **Linked backlog row:** `backlog/audit-refactor.md#R-004`
- **Linked ADR:** missing-ADR

### T-005 · Fire-and-forget Record discards emission errors — CLOSED 2026-05-11 (Plan 6a)
- **Severity:** major (closed)
- **Surface:** `internal/modules/iam/delivery/http/admin_handler.go:457` (`_ = h.audit.Record(...)`); `apps/api/cmd/metaldocs-api/main.go:467` (adapter — logs error but emits no metric and does not propagate)
- **Observation:** All consumer call sites discard or only log the Record error. There is no failure metric, no dead-letter queue, no alarm, no retry. On Postgres unavailability or PK collision (see T-006), the regulated action persists but the audit row is silently lost. Operator cannot detect dropped trail entries. Per user-supplied rubric, this is rated **Major** here because the drop happens *consumer-side* (audit module's port contract correctly returns the error — callers ignore it); the Critical-rated mirror lives in the consumer registers. Trigger fired: defense-in-depth + observability sink gap on regulated path.
- **Evidence:** `_artifacts/02-flow-record.md` §6(c); `_artifacts/03-deps.md` §2b, §2c.
- **Linked backlog row:** `backlog/audit-refactor.md#R-005`
- **Linked ADR:** missing-ADR

### T-006 · App-side timestamp-based event ID is collision-prone
- **Severity:** minor
- **Surface:** `internal/modules/iam/delivery/http/admin_handler.go:458` — `"evt_" + strings.ReplaceAll(time.Now().UTC().Format("20060102150405.000000000"), ".", "")`; same shape in `apps/api/cmd/metaldocs-api/main.go:467` (adapter)
- **Observation:** Event ID derives from a single-process `time.Now()` timestamp at nanosecond resolution with a string-replace to drop the dot. On the same process, two emitters at the same wall-clock nanosecond collide on PK `id`. On distributed deploys (none today), collisions broaden. Resulting `unique_violation` is silently lost due to T-005. Latent today (single-process API, low audit-volume). Trigger fired: latent (surface exists; no caller hits it today at observable rate).
- **Evidence:** `_artifacts/04-persistence.md` §1 (event id generation fact).
- **Linked backlog row:** `backlog/audit-refactor.md#R-006`
- **Linked ADR:** missing-ADR

### T-007 · No `tenant_id` column or tenant-scoped query path — CLOSED 2026-05-11 (Plan 6a)
- **Severity:** major (closed)
- **Surface:** `migrations/0004_init_audit_events.sql:1-14` (schema); `internal/modules/audit/infrastructure/postgres/writer.go:50-57` (SELECT has no tenant filter)
- **Observation:** `metaldocs.audit_events` has no `tenant_id` column. `ListEvents` filters by `resource_type` and `resource_id` only. When multi-tenant lands (auth T-008 plans `tenant_id` on identity tables), a tenant-A admin reading `/api/v1/audit/events` would see Tenant-B events. Latent today (single-tenant deploy); load-bearing on multi-tenant cutover. Trigger fired: multi-tenant data leak path (latent — single-tenant today, but the leak path is in code today).
- **Evidence:** `_artifacts/04-persistence.md` §1; `_artifacts/05-industry.md` IP-008.
- **Linked backlog row:** `backlog/audit-refactor.md#R-007`
- **Linked ADR:** missing-ADR

### T-008 · Missing OpenAPI `operationId` for `/audit/events`
- **Severity:** minor
- **Surface:** `api/openapi/v1/openapi.yaml:1058-1103` (path entry present; `operationId` absent — verified via Phase 2 grep, marked `(unclear)`); `internal/modules/audit/delivery/http/handler.go:35` (route mounted via `http.ServeMux.HandleFunc` directly, not via oapi-codegen)
- **Observation:** The route exists in the spec but has no `operationId`, and is NOT served via generated handlers. Drift: spec promises a contract the code does not honour through codegen. Client-codegen tools (frontend `lib/api/openapi.gen.ts`) cannot bind a method name. Trigger fired: contract surface gap with measurable consumer impact (client codegen).
- **Evidence:** `_artifacts/02-flow-list.md` §1.
- **Linked backlog row:** `backlog/audit-refactor.md#R-008`
- **Linked ADR:** `wiki/decisions/0012-contract-first-api.md` (audit was never migrated; flagged here as the residual)

### T-009 · No explicit `SELECT` grant on `metaldocs.audit_events`
- **Severity:** minor
- **Surface:** `migrations/0005_grant_workflow_audit_privileges.sql:2` grants only `INSERT` to `metaldocs_app`; no other migration grants `SELECT`
- **Observation:** Audit reads succeed in dev today — implies `SELECT` is being granted through schema-owner pathways or `PUBLIC` defaults outside the migration set, or that `metaldocs_app` is identical to the schema owner in this deployment. The intent (read for query/export, write append-only) is not explicitly encoded in migrations. On Postgres permissions hardening (revoke PUBLIC, separate roles), reads will break silently. Trigger fired: latent (works today; breaks on a routine ops change).
- **Evidence:** `_artifacts/04-persistence.md` §6 ("No explicit GRANT SELECT … found in migrations").
- **Linked backlog row:** `backlog/audit-refactor.md#R-009`
- **Linked ADR:** missing-ADR

### T-010 · No payload size constraint
- **Severity:** minor
- **Surface:** `migrations/0004_init_audit_events.sql:8` (`payload JSONB NOT NULL DEFAULT '{}'::jsonb`); `internal/modules/iam/delivery/http/admin_handler.go:453` and `apps/api/cmd/metaldocs-api/main.go:467` (payload marshalled with no size check)
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
