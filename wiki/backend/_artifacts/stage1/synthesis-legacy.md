# Stage-1 Synthesis — Legacy & Duplication Register

> **Produced:** 2026-06-10
> **Input:** Structured flags from all 19 Stage-1 mapper runs (module-audit, module-auth, module-controlled-documents, module-documents-core, module-approval, module-iam, module-search, module-security, module-taxonomy, module-templates, async-runtime, render-pipeline, http-kernel, contract-surface, platform-identity-tenancy, platform-http-toolkit, platform-data-layer, platform-ops-config, repo-topology) plus `wiki/architecture/backend-target-architecture.md` RF register.
> **Purpose:** Cross-area duplicate and legacy finding consolidation for Stage-2 evaluation. No redesigns. No fixes. No opinions beyond factual flagging.
> **Legend:** Severity bands — **critical** (security / data-loss / correctness gaps that cannot be deferred), **high** (material production risk), **medium** (observable degradation or structural debt), **low** (hygiene, latent, info). RF-* numbers refer to the refactoring register in `wiki/architecture/backend-target-architecture.md §10`.

---

## Summary Table

| # | Family | Worst severity | Primary areas | RF mapping | Stage-2 question |
|---|---|---|---|---|---|
| F-01 | Middleware chain canon inversion | high | http-kernel, platform-identity-tenancy, platform-ops-config | RF-2 | Which chain positions must move and what is the blast radius of rewiring? |
| F-02 | Unstructured logging (log.Printf vs slog) | low | module-auth, module-approval, platform-data-layer | RF-1 | Is the full request path — including auth — now log-correlation-safe? |
| F-03 | Parallel contract surface (spec2 / api/v2) | high | contract-surface, repo-topology | RF-4 | Converge into v1 or formal ADR fence with sunset plan? |
| F-04 | Duplicate outbox/repository clones in render pipeline | medium | render-pipeline, async-runtime | RF-7 (adjacent) | Extract shared generic outbox worker / repo or collapse to one table? |
| F-05 | Duplicate rate-limiter implementations | medium | platform-identity-tenancy, platform-security | RF-2, RF-9 | Decommission fixed-window global limiter; activate per-route limiter in production? |
| F-06 | Cross-module SQL / infrastructure boundary violations | medium | module-auth, module-controlled-documents, module-documents-core, module-iam, module-security, module-taxonomy | REQ-TOP-1 | Where must repository / service boundaries be introduced to isolate cross-module SQL writes and reads? |
| F-07 | Post-commit audit / governance log (atomicity gap) | high | module-taxonomy, module-templates, module-documents-core, module-approval | REQ-ASYNC-1 | Which mutating flows have a silent audit-drop window and what is the outbox migration scope? |
| F-08 | Empty / vestigial platform scaffolds and dead binaries | medium | platform-data-layer, repo-topology, platform-ops-config | RF-7 | Which directories and binaries can be deleted vs need implementation? |
| F-09 | Idempotency inline re-implementation (diverges from middleware) | medium | platform-http-toolkit, module-documents-core | RF-10 | Migrate finalize handler to use `idempotency.Require` middleware; audit remaining raw calls? |
| F-10 | N+1 and full-scan read patterns across modules | medium | module-auth, module-iam, module-search, module-approval | REQ-DATA-2 adjacent | Which hot read paths need batch queries or SQL-side pagination to meet production scale? |
| F-11 | Capability string literals and undeclared capability constants | high | module-templates, module-approval, module-iam | REQ-AUTHZ-2 | Remove all raw capability literals; confirm CI guard covers all delivery packages? |
| F-12 | Tenant isolation gaps (no tenant_id column / no RLS) | medium | module-auth, module-security, module-controlled-documents | REQ-TEN-1, REQ-DATA-2 | Which tables are missing tenant_id or DB-layer tripwires that currently rely on query predicates alone? |
| F-13 | Missing or split OpenAPI schema coverage | medium | module-search, contract-surface, module-iam | RF-4, RF-5 | Which response schemas emit fields absent from the spec, and which deprecated routes lack machine-readable markers? |
| F-14 | Dead / superseded application code retained in source | medium | module-approval, module-templates, module-documents-core, render-pipeline, module-taxonomy | — | Safe-delete candidates after behavioral verification? |
| F-15 | Duplicate private helpers across platform packages | low | platform-ops-config, platform-identity-tenancy | RF-7 | Consolidate parseBoolEnv/splitCSV into one platform/config utility? |
| F-16 | Server timeout and graceful-shutdown gaps | medium | http-kernel | RF-9 | Which timeout fields (ReadTimeout, WriteTimeout, IdleTimeout) are absent; does SIGTERM drain correctly? |
| F-17 | No OpenTelemetry exporter or W3C trace propagation | high | platform-ops-config | RF-1 | What is the minimum instrumentation set to satisfy REQ-OBS-3? |
| F-18 | Hard-coded credentials and stale binaries in VCS | critical | repo-topology | REQ-SEC-1 | Immediate: scrub DSN / binary artifacts; confirm .gitignore coverage end-to-end. |
| F-19 | jobs binary absent from Docker Compose / no Dockerfile | high | async-runtime, repo-topology | REQ-ASYNC-4 | Define metaldocs-jobs deployment path; River rows accumulate silently without it. |
| F-20 | Correlated / sequential SQL performance patterns | low | module-audit, module-search, module-security, module-approval | — | Which SQL patterns (ILIKE on JSONB, correlated COUNT subqueries) need index or query rewrites at scale? |

---

## F-01 — Middleware Chain Canon Inversion

**Severity:** high | **RF:** RF-2

### Description

The composed HTTP middleware chain in `metaldocs-api` does not match the canonical target order defined in `wiki/architecture/backend-target-architecture.md §2.1`. Three structural inversions are confirmed by code reading:

1. **httpObs (metrics/logging) sits inside authMiddleware**, not outside it. Unauthenticated 401 and CORS-reject responses are therefore invisible to RED metrics (violates REQ-MW-4).
2. **No outermost panic-recovery or request-ID middleware.** A panicked handler crashes the connection with no `problem+json` body and no tagged log line (violates REQ-MW-1 and REQ-MW-2).
3. **No pre-auth IP-keyed rate limit for the login path.** The per-route rate limiter (`platform/ratelimit`) is fully implemented and tested but is never activated in production (see F-05). Login brute-force is throttled only by account-lockout logic (violates REQ-MW-5).

### Evidence

| Claim | Location |
|---|---|
| Chain order: cors → originProtection → authMiddleware → iamMiddleware → **httpObs → rateLimiter** → mux | `apps/api/cmd/metaldocs-api/main.go:598-602` |
| No panic recovery or outer request-ID middleware visible in chain | `apps/api/cmd/metaldocs-api/main.go:598-602` |
| httpObs.Wrap called inside authMiddleware.Wrap | `apps/api/cmd/metaldocs-api/main.go:598, 602` |
| httpObs creates trace ID at line 61-65, only after auth processing | `internal/platform/observability/http.go:61-65` |
| No chain-order assertion test | `apps/api/cmd/metaldocs-api/main.go:598-602` (no companion test) |

### Cross-area duplication

The same structural gap is flagged independently by http-kernel, platform-identity-tenancy, and platform-ops-config mappers — three areas all observed the same wiring defect from different vantage points. REQ-MW-7 (chain-order test) is also unmet.

---

## F-02 — Unstructured Logging (log.Printf vs slog)

**Severity:** low | **RF:** RF-1 (REQ-OBS-1)

### Description

Multiple packages use the stdlib `log` package (`log.Printf`) instead of the project-canonical `log/slog`. These call sites lose `request_id`, `tenant_id`, and `principal_id` correlation and cannot be integrated into structured log pipelines.

### Evidence

| Location | Detail |
|---|---|
| `internal/modules/auth/delivery/http/handler.go:81,106,149,192,211,217` | log.Printf on all auth handler error paths |
| `internal/modules/auth/delivery/http/middleware.go:72` | log.Printf in auth middleware |
| `internal/modules/documents/approval/http/doc_approval_handler.go:7,39,195` | log.Printf on approval handler WARN paths |
| `internal/modules/documents/approval/http/signoff_handler.go:5` | log package imported; log.Printf in signoff handler |
| `internal/platform/objectstore/document_presigner.go:10,80` | log.Printf in object store (platform package) |

### Observation

Auth is the highest-risk location: failed-login, session errors, and middleware rejections are among the most valuable log lines for security monitoring. Their loss of context prevents correlation with distributed traces (REQ-OBS-1).

---

## F-03 — Parallel Contract Surface (spec2.yaml / internal/api/v2)

**Severity:** high | **RF:** RF-4 (REQ-API-2)

### Description

Two parallel API contract surfaces coexist with no convergence plan, no formal fence, and no ADR:

- `api/openapi/spec2.yaml` — 13 approval routes overlapping v1 approval paths; uses a non-RFC-9457 error schema (`ErrorResponse{error{code,message,details},request_id}`); no global or per-operation security declarations; not consumed by any active codegen config.
- `internal/api/v2/types_gen.go` — hand-maintained file misnamed with `_gen.go` suffix (convention reserved for machine-generated files); consumed only by three delivery-layer contract test files; defines `apiv2.APIError` with flat `{Code,Message,Details,TraceID}` structure diverging from the canonical `Problem` type.

### Evidence

| Claim | Location |
|---|---|
| spec2.yaml defines 13 routes with non-RFC-9457 error schema | `api/openapi/spec2.yaml:1-30, 596-606` |
| spec2.yaml last touched by commit 'purge v2 routes' | `api/openapi/spec2.yaml` (git history: e1944bc4a) |
| types_gen.go has no //go:generate directive | `internal/api/v2/types_gen.go:1` |
| apiv2.APIError diverges from Problem{title,status,code,detail,instance,errors} | `internal/api/v2/types_gen.go:55-60` |
| Contract tests decode problem+json into APIError via tolerant json.Unmarshal | `internal/modules/iam/delivery/http/routes_memberships_contract_test.go:109-114` |
| Capability catalog CI gate references sql/seeds/capabilities_v2.sql which does not exist | `ops/CAPABILITY_CATALOG.sha256` (literal 'placeholder-hash-update-after-catalog-created'); `.github/workflows/invariants.yml:89-118` |

### Cross-area duplication

contract-surface and repo-topology mappers both observed this independently. The invariants.yml gate that was supposed to guard the capability catalog is confirmed non-functional (exits 0 with 'Catalog file not found').

---

## F-04 — Duplicate Outbox Worker / Repository Clones in Render Pipeline

**Severity:** medium | **RF:** RF-7 (adjacent; REQ-ASYNC-1)

### Description

The render pipeline contains two structurally identical outbox implementations that differ only in table name and row type:

- `PDFOutboxWorker` and `MaterializeOutboxWorker` — identical tick/claim/dispatch/backoff algorithms.
- `PDFOutboxRepository` and `MaterializeOutboxRepository` — all six repository methods differ only in table name; ~160 lines each duplicated.

Additionally, `pdf_dispatch_outbox` and `materialize_dispatch_outbox` are domain-level staging tables that relay rows into the generic `outbox_events` table, which the external worker binary then polls. This is a two-stage outbox chain: two domain outboxes → one platform outbox → external consumer. The relay logic is structurally identical across both domain outboxes, producing a three-way duplication at the code level.

The signoff and route-admin idempotency stores in the approval module exhibit the same pattern: ~95% duplicate `ReplayHandle` adapter and JSON envelope code across `postgres_signoff_idemp_store.go` and `postgres_route_admin_idemp_store.go`.

### Evidence

| Claim | Location |
|---|---|
| PDFOutboxWorker / MaterializeOutboxWorker structural clones | `internal/modules/render/fanout/pdf_outbox_worker.go` vs `materialize_outbox_worker.go` |
| PDFOutboxRepository / MaterializeOutboxRepository structural clones (~160 lines each) | `internal/modules/render/fanout/pdf_outbox_repository.go` vs `materialize_outbox_repository.go` |
| Two-stage outbox relay chain | `internal/platform/messaging/outbox/postgres/consumer.go` |
| Approval idempotency store duplication (T-014) | `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go` vs `postgres_route_admin_idemp_store.go` |
| startOutboxWorker restart loop is dead code — Run() methods never return non-nil | `apps/api/cmd/metaldocs-api/main.go:462-486`; `pdf_outbox_worker.go:41`; `materialize_outbox_worker.go:40` |

---

## F-05 — Duplicate Rate-Limiter Implementations

**Severity:** medium | **RF:** RF-2, RF-9

### Description

Two separate rate-limiter implementations exist in the platform layer with no consolidation plan:

- `internal/platform/security/ratelimit.go` — fixed-window, global; imports domain packages (auth, iam), violating REQ-TOP-2 (platform packages must be domain-free).
- `internal/platform/ratelimit/middleware.go` — token-bucket, per-route; avoids domain imports via a `userExtractor` callback; the cleaner pattern.

The per-route limiter is fully implemented and tested but **never activated in production**: `apps/api/cmd/metaldocs-api/main.go:501` calls `RegisterRoutes(mux)` not `RegisterRoutesWithRateLimit`; `internal/modules/documents/module.go:118-119` passes `nil` for the limiter.

The global security limiter has a secondary defect: its `requestIdentity` function checks `authdomain.CurrentUserFromContext` then falls back to `iamdomain.UserIDFromContext` — both set simultaneously by auth middleware — making the fallback dead code.

### Evidence

| Claim | Location |
|---|---|
| Fixed-window global limiter with domain imports | `internal/platform/security/ratelimit.go:12-13, 182-186` |
| Per-route token-bucket limiter (clean pattern) | `internal/platform/ratelimit/middleware.go` |
| RegisterRoutes called in production (not RegisterRoutesWithRateLimit) | `apps/api/cmd/metaldocs-api/main.go:501` |
| Documents module passes nil limiter | `internal/modules/documents/module.go:118-119` |
| Dead fallback identity check | `internal/platform/security/ratelimit.go:182-186` |
| Known intentional deferral per wiki | `wiki/architecture/rate-limiting.md §4` |

---

## F-06 — Cross-Module SQL and Infrastructure Boundary Violations

**Severity:** medium | **RF:** REQ-TOP-1

### Description

Multiple modules reach directly into another module's tables or infrastructure layer without going through that module's published Go interface. This is confirmed across five distinct cross-boundary write and read paths.

### Evidence

| Violation | Location | Direction |
|---|---|---|
| auth postgres repository writes to `metaldocs.iam_users` (last_login_ip, user_agent) | `internal/modules/auth/infrastructure/postgres/repository.go:211-235` | auth → IAM table |
| GetActiveDocument issues 2-3 raw SQL queries directly against `h.db` inside the delivery layer (no repository or service indirection) | `internal/modules/controlleddocuments/delivery/http/routes.go:266-438` | delivery → DB |
| CanRead visibility SQL duplicated verbatim in repository and delivery handler; delivery holds *sql.DB solely for this | `internal/modules/controlleddocuments/infrastructure/repository.go:469-510`; `delivery/http/routes.go:281-317` | delivery → DB |
| Inline SQL in finalizeDocument handler (4 raw db.QueryRowContext calls) | `internal/modules/documents/delivery/http/handler.go:512-573` | delivery → DB |
| security module achieves cross-tenant isolation via JOIN to iam_users only (no tenant_id on auth_identities) | `internal/modules/security/infrastructure/postgres/repository.go:84-100` | security → auth table structure |
| Second standalone `PostgresControlledDocumentRepository` constructed in main.go outside module boundary | `apps/api/cmd/metaldocs-api/main.go:357` | main → infra package |
| TemplateVersionChecker queries templates table with no tx and no authz GUC, from within taxonomy module | `internal/modules/taxonomy/infrastructure/template_version_checker.go:30-48` | taxonomy → templates table |
| sessions_handler.go (IAM delivery) directly imports auth/infrastructure/postgres types | `internal/modules/iam/delivery/http/sessions_handler.go:19` | IAM delivery → auth infra |
| `platform/observability` imports `modules/auth/domain` | `internal/platform/observability/http.go:15` | platform → module domain (REQ-TOP-2) |

### Observation

The observability package violation (last row) is a REQ-TOP-2 breach: a platform package must not import module internals. The auth→IAM SQL write (first row) means a schema change to `iam_users.last_login_ip/user_agent` requires updating the auth repository — a non-obvious coupling with no domain-port abstraction.

---

## F-07 — Post-Commit Audit / Governance Log (Atomicity Gap)

**Severity:** high | **RF:** REQ-ASYNC-1

### Description

Multiple modules write audit or governance events **after** `tx.Commit()`, outside any transaction. A crash, OOM, or context cancellation between commit and the audit write silently drops the event with no retry and no outbox. This pattern recurs across at least five modules.

For a QMS-regulated audit trail this is a data-loss path on every mutating operation.

### Evidence

| Location | Mutation committed | Audit sink | Drop window |
|---|---|---|---|
| `internal/modules/taxonomy/application/family_service.go:161` (and all post-commit `govLogger.Log` calls across 11 event types) | taxonomy mutations committed | `govLogger.Log` called after commit | crash between commit and log |
| `internal/modules/taxonomy/application/profile_service.go:193` | taxonomy mutations committed | govLogger.Log | same |
| `internal/modules/taxonomy/application/area_service.go:199` | taxonomy mutations committed | govLogger.Log | same |
| `internal/modules/templates/application/lifecycle.go:70, 92` | state transition to in_review committed at :70 | AppendAudit at :92 | crash between :70 and :92 |
| `internal/modules/documents/application/service.go:803-808` | ForceReleaseSession committed | audit.Write after | crash between |
| `internal/modules/documents/application/service.go:843-848` | MarkArchived committed | audit.Write after | crash between |
| `internal/modules/documents/approval/application/decision_service.go` | approval decisions committed | audit write post-commit | same pattern |

### Additional observation

The taxonomy `DBGovernanceLogger` is marked `// Deprecated` (`internal/modules/taxonomy/application/governance_logger.go:15-17`) but remains the active sink when `AuditWriter` is nil. The `controlled-documents` module hard-imports it as a fallback (`internal/modules/controlleddocuments/module.go:31`). Two parallel audit sinks (`governance_events` vs `audit_events`) with divergent schemas and no shared retention policy coexist.

The templates audit read/write split compounds this: `AppendAudit[Tx]` now writes to `audit_events` (canonical sink) while `ListAudit` reads from the legacy `templates_audit_log` table (`internal/modules/templates/repository/postgres.go:631-650` vs `:676-714`). Events written after the write-sink migration are invisible to the read endpoint.

---

## F-08 — Empty / Vestigial Platform Scaffolds and Dead Binaries

**Severity:** medium | **RF:** RF-7 (REQ-TOP-3)

### Description

Several directories and files constitute empty scaffolding or dead artifacts that violate REQ-TOP-3 (every platform package either has production consumers or does not exist).

### Evidence

| Artifact | Status | Location |
|---|---|---|
| `internal/platform/cache/` | Empty — `.gitkeep` only, no Go source; implies caching infrastructure that does not exist | `internal/platform/cache/.gitkeep` (since commit 912879cba) |
| `internal/platform/db/.gitkeep` | Extra nesting level with no Go package at that path | `internal/platform/db/.gitkeep` |
| `internal/platform/objectstore/.gitkeep` | Superfluous — directory contains Go source files | `internal/platform/objectstore/.gitkeep` |
| `internal/platform/observability/.gitkeep` | Residual scaffold predates real package files | `internal/platform/observability/.gitkeep` |
| `cmd/seed-test-document/main.go` | Dead binary; hardcodes DSN with plaintext password and MinIO credentials; no CI reference; last touched commit c4a7d9a93 | `cmd/seed-test-document/main.go:25-30` |
| `bin/metaldocs-api.exe` | Stale compiled binary committed to git from initial commit 912879cba; .gitignore excludes root-level .exe but not `bin/` | `bin/metaldocs-api.exe` |
| `apps/api/cmd/metaldocs-api/metaldocs-api.exe` | Build artifact in source tree; gitignored but present | `apps/api/cmd/metaldocs-api/metaldocs-api.exe` |
| `apps/api/.gocache-build/` | Build cache directory in source tree | `apps/api/.gocache-build/` |
| `RepositoryMemory` mode in platform config | Dead production path — main.go:677-680 fatal-exits for any non-postgres mode; only reachable from tests | `internal/platform/config/repository.go:9`; `internal/platform/bootstrap/api.go:127-153` |

---

## F-09 — Idempotency Inline Re-Implementation

**Severity:** medium | **RF:** RF-10 (REQ-API-5)

### Description

The `platform/idempotency` package provides an `idempotency.Require` middleware that should be the canonical entry point for idempotency enforcement. Two divergence patterns exist:

1. **Finalize handler hand-rolls idempotency inline** (`internal/modules/documents/delivery/http/handler.go:440-495`), calling `IsValidKey`, `RequestHash`, `BeginReplay`, `CompleteReplay`, and `FailReplay` directly rather than delegating to the middleware. This creates a divergent implementation and makes it easy to miss `FailReplay` on error paths.

2. **Platform idempotency middleware itself uses raw string codes** bypassing its own catalog guard (`internal/platform/idempotency/middleware.go:93,97,108,111,117,123`): `"IDEMPOTENCY_KEY_INVALID"`, `"BAD_REQUEST"`, `"INTERNAL"` are inline literals with no entry in `codes.go`; `"INTERNAL"` diverges from the canonical `CodeInternalError="INTERNAL_ERROR"`. An inline comment at line 212-215 acknowledges this as a sweep miss.

### Evidence

| Claim | Location |
|---|---|
| Finalize handler raw idempotency calls | `internal/modules/documents/delivery/http/handler.go:440-495` |
| Platform middleware raw string codes | `internal/platform/idempotency/middleware.go:93,97,108,111,117,123` |
| codes.go catalog (does not list the raw strings above) | `internal/platform/problem/codes.go` |
| Idempotency TTL hard-coded as duplicate SQL string literals | `internal/platform/idempotency/postgres_store.go:91,229` |
| idempotency_keys table missing tenant_id FK | `internal/platform/idempotency/postgres_store.go:54,87` |

---

## F-10 — N+1 and Full-Scan Read Patterns

**Severity:** medium | **RF:** none explicit; adjacent to REQ-DATA-2

### Description

Multiple modules exhibit N+1 query patterns or full in-memory scans that degrade linearly with data volume and are documented as known deferred items.

### Evidence

| Module | Pattern | Location |
|---|---|---|
| module-auth | N+1 role queries in ListUsers: one `RolesByUserID` call per user in a full `auth_identities` scan | `internal/modules/auth/application/service.go:432-451` (TODO at :432) |
| module-iam | `PeopleService.ListFiltered` loads entire tenant user list then filters/paginates in Go | `internal/modules/iam/application/people_service.go:509-601` |
| module-iam | `VerifyUserInTenant` does a full-list scan per call | `internal/modules/iam/application/people_service.go` |
| module-iam | `RolesByUserID` makes two DB round trips on cache-miss (liveness check then roles query) | `internal/modules/iam/infrastructure/postgres/role_provider.go:20-57` |
| module-approval | `ListPendingForActor` uses SELECT DISTINCT IDs then `LoadInstance` per ID in a loop | `internal/modules/documents/approval/application/read_service.go:172-232` |
| module-search | No pagination; offset hard-coded to 0 across the full call chain | `internal/modules/search/application/service.go:53`; `domain/port.go:9`; `infrastructure/v2documents/reader.go:21` |
| module-audit | ILIKE on `payload::text` (full-table scan on JSONB, no GIN index) | `internal/modules/audit/infrastructure/postgres/writer.go:260-264` |
| module-search | `document_profiles` subquery duplicated identically in SELECT and WHERE clauses of same SQL statement | `internal/modules/search/infrastructure/v2documents/reader.go:34-41, 59-66` |

---

## F-11 — Capability String Literals and Undeclared Capability Constants

**Severity:** high | **RF:** REQ-AUTHZ-2

### Description

Multiple delivery and application packages use raw capability string literals that bypass the typed registry. REQ-AUTHZ-2 requires that every capability referenced anywhere is a typed registry constant and that raw strings are CI-rejected. Three confirmed violations:

1. `"template.admin"` is passed to the tier-1 authz check for `PUT /approval-config` but no `CapTemplateAdmin` constant exists in `internal/modules/iam/domain/model.go`. Runtime behavior when `CapabilityService.CheckCapability` receives an unknown capability is **runtime-unverified** — the route may be permanently locked or permanently open.

2. Route admin event types use raw string literals for event type constants (`"route.config.created"`, `"route.config.updated"`, `"route.config.deactivated"`) rather than typed `EventType` constants.

3. The codegen enum `CreateManagedUserRequestRoles` lists only 5 of 8 canonical roles — omits signer, area_admin, qms_admin — meaning the OpenAPI-generated surface is behind the domain model for user creation.

### Evidence

| Claim | Location |
|---|---|
| `"template.admin"` literal with no matching CapTemplateAdmin const | `internal/modules/templates/delivery/http/routes_lifecycle.go:192`; `internal/modules/iam/domain/model.go` (8 CapTemplate* consts declared, none for admin) |
| Route admin inline event type literals | `internal/modules/documents/approval/application/route_admin_service.go:223,375,520` |
| Tier-1/tier-2 capability mismatch on publish route (`"template.approve"` at tier-1 vs `CapTemplatePublish` at tier-2) | `internal/modules/templates/delivery/http/routes_generated.go:203` |
| Codegen enum missing 3 roles | `internal/modules/iam/api/api.gen.go:51-75` |
| Stale Phase-2 residual-gap comment after ADR 0022 Phase 7 | `internal/modules/iam/domain/capability_scope.go:31-35` |

---

## F-12 — Tenant Isolation Gaps

**Severity:** medium | **RF:** REQ-TEN-1, REQ-DATA-2

### Description

Several tables or modules enforce tenant isolation solely via application-layer query predicates with no DB-level backstop (no RLS, no trigger tripwire, no `tenant_id` column).

### Evidence

| Gap | Location | Nature |
|---|---|---|
| `auth_identities` has no `tenant_id` column — identity is tenant-global | `internal/modules/auth/infrastructure/postgres/repository.go:379-408`; `archive/migrations/0021_init_auth_identities_and_sessions.sql` | structural; tracked T-008 |
| `controlled_documents` and `cd_sequence_counters` have no GUC + RLS backstop (T-005) | `internal/modules/controlleddocuments/infrastructure/repository.go` (all query methods) | predicate-only isolation |
| Security module cross-tenant isolation achieved via JOIN to `iam_users` only | `internal/modules/security/infrastructure/postgres/repository.go:84-100` | structural JOIN dependency |
| `audit_export_jobs.tenant_id` is UUID NOT NULL in Postgres but `string` in Go domain type; insert passes string directly | `db/migrations/0224_audit_export_jobs_pr6.sql:22`; `internal/modules/audit/domain/port.go:146`; `infrastructure/postgres/exports.go:43` | type mismatch; silently works for valid UUIDs |
| `audit_events.tenant_id` cast inconsistency between queries in security module | `internal/modules/security/infrastructure/postgres/repository.go:266` vs `:32, 97` | [runtime-unverified] |
| MembershipGovernanceLogger wired nil in production — grant/revoke produce no governance log | `apps/api/cmd/metaldocs-api/main.go:325` | T-007 open; asymmetric audit trail |
| `iam_users` INSERT has no DB tripwire trigger (`trg_require_cap_asserted` not attached to iam_users) | `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:52-58` | DB enforcement floor absent for user-record write |

---

## F-13 — Missing or Split OpenAPI Schema Coverage

**Severity:** medium | **RF:** RF-4, RF-5

### Description

The OpenAPI spec surface has several confirmed gaps between what the code emits, what the spec declares, and what codegen consumes.

### Evidence

| Gap | Location |
|---|---|
| `SearchDocumentItem` schema omits subject, business_unit, classification, tags fields emitted by the handler response struct (always zero-valued at runtime, but schema and wire format diverge) | `internal/modules/search/delivery/http/handler.go:24-43`; `api/openapi/v1/openapi.yaml:4785-4820` |
| Four search query parameters (subject, businessUnit, classification, tag) accepted and forwarded but silently no-op at SQL reader | `internal/modules/search/delivery/http/handler.go:96-103`; `infrastructure/v2documents/reader.go:23-27` |
| `businessUnit` query parameter is camelCase (all others snake_case) and undocumented in spec | `internal/modules/search/delivery/http/handler.go:99` |
| Partial files (`partials/controlled-documents.yaml`, `partials/documents.yaml`, `partials/templates.yaml`) use camelCase schema properties and `/api/v1/` path prefix — both violations of canonical conventions; CASING-DRIFT and PATH-BASE-PREFIX lint rules do not run on these files | `api/openapi/v1/partials/controlled-documents.yaml:207-305`; `partials/documents.yaml:47-51`; all three partial files |
| `ControlledDocumentVisibility` in partial has `areaCodes/userIds` (camelCase); canonical spec has `area_codes/user_ids` (snake_case) — two live definitions of same resource | `api/openapi/v1/partials/controlled-documents.yaml:221`; `api/openapi/v1/openapi.yaml:5020-5030` |
| POST /iam/users lacks machine-readable `deprecated: true` despite description calling it legacy | `api/openapi/v1/openapi.yaml:213` |
| iamapi codegen bundles three distinct domain concerns (iam, audit, security) into one 5691-line generated file | `internal/modules/iam/api/cfg.yaml:9-11`; `internal/modules/iam/api/api.gen.go` |
| api-lint.exe pre-built Windows binary committed to repository | `scripts/api-lint/api-lint.exe` |

---

## F-14 — Dead / Superseded Application Code Retained in Source

**Severity:** medium (for correctness risk) / low (for pure dead code) | **RF:** none direct

### Description

Multiple files or types are confirmed dead code — either explicitly deprecated, post-migration one-time utilities, or never wired in production.

### Evidence

| Artifact | Status | Location |
|---|---|---|
| `CutoverService` | Dead post-migration-0142 (applied ~1 year ago); not in `Services` struct; no HTTP route; reachable only from `coverage_boost_test.go` | `internal/modules/documents/approval/application/cutover_service.go:1-80` |
| `Deprecated PDFDispatchInvoker` post-commit dispatch path | Compiled but inactive when `pdfOutbox` is wired (production case); silently activates if caller omits outbox wiring | `internal/modules/documents/approval/application/decision_service.go:43,60,64,566-573` |
| `FreezeService.Freeze` synchronous path | Annotated 'New code should use Pin + Materialize instead'; unclear if any production callsite still exercises it | `internal/modules/documents/application/freeze_service.go:302` |
| `CompositionConfig` struct | ADR 0008 removed composition 2026-04-27; struct exported with no inbound callers | `internal/modules/templates/domain/schemas.go:81` |
| `AreaService.SetParent` (cycle-safe, transactional) | Only called from `area_service_test.go:49`; production `updateArea` handler calls `AreaService.Update` instead (without cycle check) | `internal/modules/taxonomy/application/area_service.go:111` |
| `SnapshotService.SnapshotFromTemplate` | Marked deprecated; retained for unnamed backfill scripts | `internal/modules/documents/application/snapshot_service.go:46` |
| `WorkerConfig.ReviewReminderDays` | Parsed and logged at startup but referenced by nothing; planned feature never implemented | `internal/platform/config/worker.go:13,25,50`; `apps/worker/cmd/metaldocs-worker/main.go:109` |
| Legacy `areas`, `visibility`, `specific_areas` columns on templates | Written by INSERT as empty values; not scanned; no domain.Template fields | `internal/modules/templates/repository/postgres.go:52-53` |
| `document_profiles.is_active` column | Added by migration 0023; superseded by `archived_at` in 0122; never read or written by Go code | `archive/migrations/0023_init_document_family_and_profile_registry.sql:14`; `archive/migrations/0122_taxonomy_extend_document_profiles.sql:13` |
| `document_subjects` table | Created by migration 0025; zero Go code references it | `archive/migrations/0025_init_document_taxonomy.sql:9-16` |
| `resolvePermissionFallback` function | Switch with only a default case and a discarded path parameter; leftover scaffolding | `apps/api/cmd/metaldocs-api/permissions.go:270-279` |
| `coverage_boost_test.go` | Explicitly created to push coverage to ≥90%; duplicates setup from primary test files | `internal/modules/documents/approval/application/coverage_boost_test.go:1` |

---

## F-15 — Duplicate Private Helpers Across Platform Packages

**Severity:** low | **RF:** RF-7 (REQ-TOP-3 adjacent)

### Description

Identical private helper functions have been independently re-implemented in multiple packages rather than sharing a common utility. Divergence risk exists if the trimming logic ever changes.

### Evidence

| Helper | Duplicated at |
|---|---|
| `parseBoolEnv` | `internal/platform/config/attachments.go:108` and `internal/platform/authn/config.go:213` |
| `splitCSV` | `internal/platform/config/cors.go:63` and `internal/platform/authn/config.go:222` |

`authn/config.go` appears to predate the `platform/config` package restructure and was never updated to import the canonical helpers.

---

## F-16 — Server Timeout and Graceful-Shutdown Gaps

**Severity:** medium | **RF:** RF-9 (REQ-REL-1, REQ-REL-2)

### Description

The `http.Server` configuration in the API binary sets only `ReadHeaderTimeout: 5s`; `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` are absent. Slow-body or slow-read clients hold connections indefinitely, exhausting the connection pool under moderate load.

### Evidence

| Claim | Location |
|---|---|
| http.Server missing ReadTimeout, WriteTimeout, IdleTimeout | `apps/api/cmd/metaldocs-api/main.go:613-617` |

### Related observations

- `applyDependencyChecks` in the readiness probe runs checks sequentially under individual 2-second timeouts rather than a shared budget; worst-case `/api/v1/health/ready` latency is `2N + 3` seconds for N dependency checks (`internal/platform/observability/runtime.go:286-323`).
- `METALDOCS_ATTACHMENTS_SIGNING_SECRET` is required unconditionally even for memory/local storage providers that do not use HMAC-signed URLs (`internal/platform/config/attachments.go:63-68`).

---

## F-17 — No OpenTelemetry Exporter or W3C Trace Propagation

**Severity:** high | **RF:** RF-1 (REQ-OBS-3)

### Description

There is no OpenTelemetry instrumentation anywhere in the Go codebase. The `backend-blueprint.md §D2` lists observability as 🟡 and mentions "confirm exporter wiring." This audit confirms the gap is total: no OTLP exporter, no Prometheus endpoint, no W3C `traceparent` propagation, and no span creation.

RED metrics are produced as structured log aggregates only. There is no distributed trace that spans edge → api → outbox → worker → docx-renderer. REQ-OBS-3 is fully open.

### Evidence

| Claim | Location |
|---|---|
| Zero `go.opentelemetry.io` imports in platform/observability | `internal/platform/observability/` (all files — grep confirmed) |
| normalizeRoute only covers 5 path patterns; all other routes log raw paths with IDs (metric cardinality / PII leak in logs) | `internal/platform/observability/http.go:178-208` |
| httpObs creates a trace ID internally but it is process-local only (no W3C propagation) | `internal/platform/observability/http.go:61-65` |

---

## F-18 — Hard-Coded Credentials and Stale Binaries in VCS

**Severity:** critical | **RF:** REQ-SEC-1

### Description

A committed source file contains plaintext credentials. This is the highest-severity finding in the register.

### Evidence

| Claim | Location |
|---|---|
| `cmd/seed-test-document/main.go` contains hardcoded DSN with plaintext password (`***REDACTED***`) and hardcoded MinIO credentials | `cmd/seed-test-document/main.go:25-30` |
| `bin/metaldocs-api.exe` is a stale compiled binary committed from the initial commit; .gitignore covers root .exe but not `bin/` | `bin/metaldocs-api.exe` |
| `scripts/api-lint/api-lint.exe` — pre-built Windows binary committed; provenance cannot be verified | `scripts/api-lint/api-lint.exe` |
| `scripts/start-spec1-api.ps1` hardcodes an absolute path to a different machine's username | `scripts/start-spec1-api.ps1:2` |

### Stage-2 action

The plaintext password in `cmd/seed-test-document/main.go` must be assessed for whether the credential has been rotated since commit c4a7d9a93. This is a pre-Stage-2 prerequisite, not a Stage-2 deferral.

---

## F-19 — jobs Binary Absent from Docker Compose / No Dockerfile

**Severity:** high | **RF:** REQ-ASYNC-4

### Description

`metaldocs-jobs` is a distinct deployable binary (River job host for scheduled publish). It has no Dockerfile and no service entry in `docker-compose.yml`. Without it running, `scheduled_publish_cutover` River rows accumulate and documents never auto-publish in containerized deployments.

The API binary and the jobs binary both call `MigrateRiverSchema` independently at startup, meaning River schema lifecycle has no single declared owner.

### Evidence

| Claim | Location |
|---|---|
| No `jobs.Dockerfile` in deploy/docker/ | `deploy/docker/` (only api.Dockerfile, worker.Dockerfile) |
| No jobs service in docker-compose.yml | `deploy/compose/docker-compose.yml` |
| Local-only `start-jobs.ps1` entrypoint | `scripts/start-jobs.ps1` |
| River schema migration called from both binaries | `apps/api/cmd/metaldocs-api/main.go:439`; `internal/platform/bootstrap/jobs.go:36` |
| `lease_reaper` subquery matches job names against `public.documents.id` — always returns NULL for scheduler job names | `internal/modules/jobs/scheduler/lease_reaper.go:37-86` |

---

## F-20 — Correlated / Sequential SQL Performance Patterns

**Severity:** low | **RF:** none; adjacent to REQ-DATA-2

### Description

Several confirmed SQL patterns produce unnecessary query load. All are currently low-impact due to data volume but degrade linearly or super-linearly with table growth.

### Evidence

| Pattern | Location |
|---|---|
| ILIKE on `payload::text` (full-table sequential scan on JSONB; no GIN/GiST index) | `internal/modules/audit/infrastructure/postgres/writer.go:260-264` |
| `TODO(phase11)` leading-wildcard ILIKE for CD full-text search | `internal/modules/controlleddocuments/infrastructure/repository.go:128` |
| `listRoutesQuery` correlated `SELECT COUNT(*)` re-evaluated per row | `internal/modules/documents/approval/repository/postgres_approval_repository.go:439` |
| `document_profiles` family subquery duplicated identically in SELECT and WHERE | `internal/modules/search/infrastructure/v2documents/reader.go:34-41, 59-66` |
| `listRoutesQuery` aliases `created_at AS updated_at`; no `updated_at` column exists on `approval_routes` | `internal/modules/documents/approval/repository/postgres_approval_repository.go:437` |
| `InMemoryAuthFailureRateLimiter` wired in production — process-local, no cross-replica coordination | `internal/modules/documents/approval/infrastructure/signature/password_reauth.go:118-127`; `apps/api/cmd/metaldocs-api/reauth.go:49` |
| Sequential security rule evaluation (4 independent SQL queries in series, no early exit) | `internal/modules/security/application/service.go:93-149` |

---

## Cross-Area Duplications Not Visible to Individual Mappers

The following duplications required comparing artifacts from multiple mapper runs to detect. No single mapper could observe these because each mapper had bounded scope.

### D-01: Governance log written post-commit in five independent modules

The F-07 family spans taxonomy (11 event types), templates (SubmitForReview), documents-core (ForceRelease, Archive), and approval (decision). Each mapper identified its own instance. The cross-area picture is: **no module using a govLogger or audit.Write call currently has atomic outbox semantics.** The only exception is the platform-level outbox workers, which are themselves subject to the two-stage chain duplication in F-04.

### D-02: Three separate MinIO client instances from the same credentials

Each of `buildMinioClients:158-178` (internal + public presigning clients) and `miniostore.NewStore` (byte I/O client) creates a distinct `*minio.Client` from the same endpoint and credentials. This is invisible to any single module mapper because all three are wired at the platform bootstrap layer (`internal/platform/bootstrap/api.go:85-103`). No connection pool is shared.

**Location:** `internal/platform/bootstrap/api.go:85-103, 158-178`

### D-03: 405 Method Not Allowed returned as bare status with no RFC 9457 body in two modules

Both the search module (`internal/modules/search/delivery/http/handler.go:55-57`) and the security module (`internal/modules/security/delivery/http/handler.go:37-39, 58-60, 98-100`) return a bare `405` status with no `problem+json` body. Every other error path in both modules uses `problem.Write` / `httpresponse.WriteError`. This is a cross-module consistency defect affecting the same response type.

### D-04: `platform/security` vs `platform/ratelimit` — same concern, different packages

`internal/platform/security/` contains rate limiting, origin protection, and CORS helpers. `internal/platform/ratelimit/` contains a separate rate-limiting middleware. These two packages implement the same concern at different levels of maturity and with different dependency models. The security package imports domain packages (violation of REQ-TOP-2); the ratelimit package correctly uses callbacks. The relationship between these two packages is not documented anywhere.

**Locations:** `internal/platform/security/ratelimit.go`; `internal/platform/ratelimit/middleware.go`

### D-05: `platform/cache` placeholder vs `CachedRoleProvider` (no declared cache contract)

`platform/cache` is an empty scaffold (F-08). The only production cache is the `CachedRoleProvider` in the IAM module (`internal/modules/iam/application/cached_role_provider.go`), which is an in-process `sync.Map` TTL cache. REQ-CACHE-1 requires a written cache contract (TTL, invalidation, staleness bound, failure behavior). No such contract exists for the role provider. The `platform/cache` placeholder may have been intended to house this — its emptiness means the pattern is module-local with no platform support.

**Location:** `internal/platform/cache/.gitkeep`; `internal/modules/iam/application/cached_role_provider.go:80-83`

### D-06: `cmd/` root vs `apps/*/cmd/` — two binary entrypoint conventions

`cmd/seed-test-document/` lives at the repository root under `cmd/`, while all active binaries live under `apps/<name>/cmd/<binary>/`. `go build ./...` traverses both locations. The `cmd/` root convention is a Go standard-library holdover; the `apps/` convention is MetalDocs-canonical. The only inhabitant of the root `cmd/` is the dead seed binary flagged in F-18.

**Location:** `cmd/seed-test-document/`; `apps/api/cmd/metaldocs-api/`; `apps/worker/cmd/metaldocs-worker/`; `apps/jobs/cmd/metaldocs-jobs/`

### D-07: Domain status enum fragmentation across modules

`internal/modules/documents/domain/model.go:8-13` defines only 3 of 8 live `DocumentStatus` values; the remaining 5 (approved, published, superseded, obsolete, scheduled, rejected) exist only in `api.gen.go`. In the taxonomy module, `document_profiles.is_active` was superseded by `archived_at` (migration 0122) but the orphaned column and default remain. Both patterns represent the same root cause: incremental migration with no cleanup of the original type surface.

---

## Appendix: Flags With No Family Assignment

The following low-severity flags are informational and do not fit cleanly into any family above. They are recorded here for completeness.

| Flag | Location | Severity |
|---|---|---|
| All exported symbols lack Go doc comments (affects: audit, auth, CD, documents, templates, taxonomy, iam, security, approval) | All affected modules | info |
| `application.Service` exported struct allows direct zero-value construction in tests | `internal/modules/audit/application/service_test.go:36` | info |
| `RecordFailedLogin` concrete method on `*postgres.Repository` not in interface | `internal/modules/auth/infrastructure/postgres/repository.go:365-377` | info |
| Migration history deployment warning embedded in domain model Go source | `internal/modules/auth/domain/model.go:31-34` | info |
| `memory/repository.go` implements both auth and IAM interfaces in one 540-line struct | `internal/modules/auth/infrastructure/memory/repository.go:1-540` | info |
| `DBExecutor` type alias is a redundant exported name for `DBTX` | `internal/modules/controlleddocuments/domain/sequence.go:14` | info |
| `crypto/sha1` used in `stableID` (content-addressing only; may trigger static analysis) | `internal/modules/security/application/service.go:179` | info |
| Hardcoded Portuguese fallback string `'[aguardando aprovação]'` in ApproversResolver | `internal/modules/render/resolvers/approvers.go:34` | info |
| `controlled_by_area` resolver is version 2; all others version 1; no versioning policy | `internal/modules/render/resolvers/controlled_by_area.go:14` | info |
| Package name `docgenv2` is a pre-rename artifact; service is now called `docx-renderer` | `internal/platform/docgenv2/` | info |
| TemplateReader magic sentinel UUID for system templates | `internal/platform/docgenv2/template_reader.go:13` | info |
| Stale file:line anchors in permissions_test.go comments | `apps/api/cmd/metaldocs-api/permissions_test.go:202-298, 428` | info |
| Duplicate snapshot repository construction in main() (3 instances, same underlying pool) | `apps/api/cmd/metaldocs-api/main.go:372, 398, 420` | info |
| e2e-seed binary opens two independent DB pools in sequence | `apps/api/cmd/metaldocs-e2e-seed/main.go:52-56, 104-114` | info |
| Large archive migration tree adjacent to active migrations | `archive/migrations/` | info |

