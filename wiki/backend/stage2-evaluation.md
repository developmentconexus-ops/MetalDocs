# Stage-2 Evaluation: Backend vs Professional SaaS Standards

> **Last verified:** 2026-06-11
> **Scope:** Synthesized evaluation across all 7 themes. Consumes `wiki/backend/legacy-register.md` and all `wiki/backend/_artifacts/stage2/` theme artifacts. READ-ONLY on code; every verdict is grounded in a confirmed file:line anchor plus an external citation.
>
> **Goal (verbatim):** Take the Stage-1 legacy register (factual findings) and EVALUATE each finding against professional SaaS practice and industry standards, to decide how to restructure the backend so it can scale, be maintained, and run in production at the user's company. This is NOT a hyperscale system (yet) — size judgments for sane future growth, not for a thousand engineers. Apply the simplicity-first rule HARD: prefer the smallest change that reaches a professional bar; explicitly flag any finding where the current code is already fine or where a "fix" would be over-engineering. Every verdict must cite evidence (Go/PostgreSQL official docs, OWASP, ISO 27001/9001 where the QMS angle matters, RFC numbers, well-known engineering references) — no opinion without a reference.

---

## Executive Posture

The MetalDocs backend is in a legitimate early-production state: contract-first API, two-tier authz, a working transactional outbox for the approval pipeline, and a reasonable module layout. The api-contract-hardening program (Phases A–F, closed 2026-06-08) resolved the major structural contract issues and the backend standardization batches eliminated most casing and error-vocabulary drift. The code is buildable, testable, and passing CI.

What the Stage-2 evaluation reveals is a cluster of gaps that compound each other: plaintext credentials committed to VCS are a P0 prerequisite that must be resolved before any other security work is meaningful; a middleware chain inverted around the metrics layer produces misleading RED metrics and leaves the login endpoint under-protected against brute-force; post-commit audit/governance writes violate ISO 9001 §8.5.1 and 21 CFR Part 11 §11.10(e) atomicity requirements across five modules; and the jobs binary is absent from the deployment manifest, making scheduled-publish silently non-functional in every containerised deployment. These four items — credentials, middleware, audit atomicity, and jobs deployment — are not aspirational improvements; they are correctness and compliance defects at the professional SaaS bar.

The remaining findings (module boundary violations, N+1 query patterns, dead code, contract surface drift) are real but lower-urgency at current data volume and team size. The anti-over-engineering posture is maintained throughout: nine of the total verdict actions are SIMPLIFY or DELETE (reducing code), and every verdict explicitly bounds the fix to the smallest change that reaches the required bar. No finding in this evaluation warrants a distributed cache, an external search engine, an event bus, or a framework migration. The platform needs surgical repairs, not a rewrite.

Two ADRs are required. All other verdicts execute within existing normative requirements (REQ-* IDs in `backend-target-architecture.md`) and existing ADR decisions.

---

## P0 Prerequisites

These block the roadmap. No Wave 1 security work is meaningful until F-18 is resolved. No Wave 1 middleware work is clean until F-06a is resolved.

### P0-1: F-18 — Plaintext credentials committed to VCS

`cmd/seed-test-document/main.go:25-28` contains a plaintext Postgres password (`***REDACTED***`) and MinIO credentials. This is CWE-798 / OWASP ASVS V2.10.4 / OWASP Top 10 A07. Git history permanently retains the secret unless rewritten. The MinIO credentials are default values; the Postgres password is non-default and specific.

**Blocked-by:** nothing. Execute immediately.

**Fix sequence:**
1. Rotate the Postgres password if it matches any non-dev environment.
2. Delete `cmd/seed-test-document/` (dead binary, D-06 closes as side effect).
3. Rewrite git history (`git filter-repo --invert-paths --path cmd/seed-test-document/main.go`), force-push, notify cloners.
4. `git rm bin/metaldocs-api.exe`; add `bin/*.exe` to `.gitignore`.
5. Replace `scripts/api-lint/api-lint.exe` with a `go build`-produced artifact or add a SHA-256 manifest.
6. Delete `scripts/start-spec1-api.ps1` (confirmed dead).

### P0-2: F-19-deployment — Jobs binary absent from Docker Compose

`apps/jobs/cmd/metaldocs-jobs/main.go` has Go source but no `deploy/docker/jobs.Dockerfile` and no service in `deploy/compose/docker-compose.yml`. Without the jobs service, `scheduled_publish_cutover` River rows accumulate indefinitely and documents scheduled for future publication never become `published`. This is a silent, user-visible feature failure in every containerised deployment. Twelve-Factor App Factor VI and riverqueue.com/docs require a declared, running worker process.

**Blocked-by:** nothing. Execute immediately after F-18.

**Fix:** add `jobs.Dockerfile` (copy `worker.Dockerfile`, substitute `metaldocs-jobs`) and add a `jobs` compose service with `depends_on: postgres (healthy), api (healthy)`.

### P0-3: F-06a — `platform/observability` imports `auth/domain` (REQ-TOP-2 MUST)

`internal/platform/observability/http.go:15` imports `authdomain.CurrentUserFromContext` — a platform package importing a module domain type. This is a hard REQ-TOP-2 inversion that blocks a clean middleware chain reorder (F-01). The pattern is already solved correctly in `platform/ratelimit.Middleware` which accepts a `userExtractor func(*http.Request) string` callback. Apply the same inversion here.

**Blocked-by:** nothing. Execute before F-01 middleware reorder.

---

## Master Verdict Table

| Finding | Theme | Verdict | Priority | Effort | Blast Radius | Standard Cited | Needs ADR |
|---|---|---|---|---|---|---|---|
| F-18: Hardcoded credentials + dead binaries | security | REFACTOR | P0-prereq | S | contained | CWE-798; OWASP ASVS V2.10.4; OWASP Top 10 A07/A08; ISO 27001:2022 A.8.3; REQ-SEC-1 | No |
| F-19-deployment: Jobs binary absent from compose | async | REFACTOR | P0-prereq | S | contained | 12factor.net Factor VI; riverqueue.com/docs; ISO 9001:2015 §7.1.3; REQ-ASYNC-4 | No |
| F-06a: platform/observability → auth/domain import | boundaries | REFACTOR | P0-prereq | S | contained | REQ-TOP-2; Cockburn Hexagonal Architecture | No |
| F-01: Middleware chain inversion (metrics inside auth, no panic recovery, no pre-auth rate limit) | observability | REFACTOR | P1 | S | contained | REQ-MW-1/2/4/5/7; OWASP ASVS V2.2.1; CWE-307; Google SRE RED method | No |
| F-16: Missing ReadTimeout/WriteTimeout/IdleTimeout | observability | REFACTOR | P1 | S | contained | CWE-400; OWASP API4:2023; Cloudflare Go timeout guide; REQ-REL-1/2 | No |
| F-07: Post-commit audit/governance writes (taxonomy, templates, documents-core) | async | REFACTOR | P1 | M | cross-module | ISO 9001:2015 §8.5.1; 21 CFR Part 11 §11.10(e); Transactional Outbox pattern; REQ-ASYNC-1 | No |
| D-01: Same root cause as F-07 across all modules | async | REFACTOR | P1 | M | cross-module | ISO 9001:2015 §8.5.1; Transactional Outbox pattern; REQ-ASYNC-1 | No |
| F-07-sub-split: Templates read/write audit sink split | async | REFACTOR | P1 | S | module | ISO 9001:2015 §8.5.1; DB single-source-of-truth principle | No |
| F-12: Tenant isolation — no DB-layer RLS backstop on high-risk tables | security | REFACTOR | P1 | M | module | REQ-TEN-1; OWASP ASVS V4.1.3; OWASP Top 10 A01; PostgreSQL RLS docs; ISO 27001:2022 A.5.15 | Yes (ADR-A) |
| F-11: Capability string literals — correctness defect + tier mismatch | security | REFACTOR | P1 | M | module | REQ-AUTHZ-2; ADR 0022; OWASP ASVS V4.1.1; CWE-284 | Yes (ADR-B, conditional) |
| F-03: Dead parallel contract surface (spec2.yaml + internal/api/v2) | contract | DELETE | P1 | S | contained | REQ-API-2; OpenAPI 3.0.3 §4.1; OWASP API9:2023; RFC 9457 §3 | No |
| F-09: Raw string codes in idempotency middleware (REQ-H-2 violation) | contract | REFACTOR | P1 | S | contained | REQ-H-2; REQ-API-5; RF-10; RFC 9457 §5; Stripe idempotency guide | No |
| F-13a: Dead search struct fields (subject/business_unit/classification/tags) | contract | SIMPLIFY | P1 | S | contained | OpenAPI 3.0.3 §4.8.20; REQ-API-4 | No |
| F-13b: `businessUnit` camelCase undocumented query param | contract | REFACTOR | P1 | S | contained | REQ-API-4; OWASP API9:2023 | No |
| F-19-dual-migration: Both api + jobs binaries call MigrateRiverSchema | async | REFACTOR | P1 | S | contained | riverqueue.com/docs/migrations; single-owner schema principle; REQ-ASYNC-4 | No |
| F-19-lease-reaper-bug: Governance writes always no-op for maintenance jobs | async | REFACTOR | P1 | S | contained | ISO 9001:2015 §8.5.1; 21 CFR Part 11 §11.10(e); REQ-ASYNC-4 | No |
| F-06b: Delivery-layer raw SQL in CD + documents handlers | boundaries | REFACTOR | P1 | M | module | REQ-H-1; REQ-TOP-1; Cockburn Hexagonal Architecture | No |
| F-06c: auth repo writes to iam_users (RecordLastLoginContext) | boundaries | REFACTOR | P1 | M | cross-module | REQ-TOP-1; DDD Bounded Context (Evans 2003) | No |
| F-06d: iam/delivery imports auth/infrastructure/postgres | boundaries | REFACTOR | P1 | S | module | REQ-TOP-1; Cockburn — adapters depend on ports | No |
| F-05 + D-04: Delete security.RateLimiter; activate platform/ratelimit | boundaries | REFACTOR | P1 | M | module | REQ-TOP-2; REQ-MW-5; Go internal/ package cohesion | No |
| F-08: Empty scaffolds (.gitkeep) + committed binary (bin/metaldocs-api.exe) | hygiene | DELETE | P1 | S | contained | REQ-TOP-3; CWE-561; YAGNI | No |
| F-14: Dead application code (CutoverService, CompositionConfig, SetParent, etc.) | hygiene | DELETE/SIMPLIFY | P1 | M | module | CWE-561; YAGNI; Go effective idiom | No |
| D-06: cmd/ convention split (resolved by F-14 deletion) | hygiene | DELETE | P1 | S | contained | REQ-TOP-3; YAGNI; Go module layout | No |
| F-10: N+1 queries + full-scan reads (role, membership, approval inbox, audit ILIKE) | performance | REFACTOR | P1 | M | module | PostgreSQL §14.3; use-the-index-luke.com N+1; Winand "SQL Performance Explained" ch. 6; REQ-DATA-2; REQ-AUTHZ-6 | No |
| F-20e: InMemoryAuthFailureRateLimiter wired in production | performance | REFACTOR | P1 | M | module | OWASP ASVS v4.4 §2.2.1/§2.2.6; REQ-REL-3 | No |
| F-17: No OTel exporter / no W3C trace propagation | observability | REFACTOR | P2 | M | cross-module | W3C Trace Context L2; OTel Go SDK; Google SRE Four Golden Signals; REQ-OBS-1/2/3; RF-1 | No |
| F-06e: taxonomy.TemplateVersionChecker queries templates tables directly | boundaries | REFACTOR | P2 | S | cross-module | REQ-TOP-1; DDD Bounded Context | No |
| F-13c: Partial files dead with camelCase (3 files) | contract | DELETE | P2 | S | contained | OWASP API9:2023; REQ-API-4 | No |
| F-13d: POST /iam/users missing `deprecated: true` | contract | SIMPLIFY | P2 | S | contained | OpenAPI 3.0.3 §4.8.11.6 | No |
| F-07-sub-deprecated-logger: DBGovernanceLogger nil fallback | async | DELETE | P2 | S | module | Go doc comment convention; simplicity-first | No |
| F-04: Duplicate staging outbox worker/repo clones (render pipeline) | async | SIMPLIFY | P2 | M | module | DRY principle; Go generics spec go.dev/ref/spec; RF-7 | No |
| D-02: Three MinIO clients from same credentials | hygiene | SIMPLIFY | P2 | S | contained | minio-go v7 goroutine safety; Go net/http.Client concurrency | No |
| D-05: No written cache contract for CachedRoleProvider | hygiene | REFACTOR | P2 | S | contained | REQ-CACHE-1; REQ-TOP-3 | No |
| D-07: DocumentStatus enum missing 6 constants | hygiene | SIMPLIFY | P2 | S | module | Go effective idiom §Constants; staticcheck SA4003 | No |
| F-09a: Finalize handler inline idempotency | contract | KEEP | — | — | — | Structurally correct; middleware wrapping would be over-engineering | No |
| F-20f: Sequential security signal queries | performance | KEEP | — | — | — | Low-frequency admin surface; current latency acceptable; errgroup would add complexity | No |
| F-15: parseBoolEnv semantic drift | boundaries | SIMPLIFY | P3 | S | contained | Go internal/ module layout; Go proverb (go-proverbs.github.io) | No |
| F-15: splitCSV duplication | boundaries | KEEP | — | — | — | Go proverb applies; identical copies, 5 lines | No |
| F-16B: WebSocket presence drain on SIGTERM | observability | DEFER | P3 | M | module | Go stdlib net/http#Server.Shutdown re hijacked connections; K8s SIGKILL backstop | No |
| F-16C: Sequential readiness probe checks | observability | SIMPLIFY | P3 | S | contained | Correctness gap for N>1 deps; N=1 today | No |
| F-20b: CD leading-wildcard ILIKE | performance | DEFER | P2 | M | module | use-the-index-luke.com LIKE Performance; pg_trgm docs; trigger: >50k rows or p95 >200ms | No |
| F-09c: Idempotency TTL string duplicated | contract | SIMPLIFY | P3 | S | contained | DRY principle | No |
| F-09d: idempotency_keys.tenant_id missing FK | contract | DEFER | — | M | module | Trigger: F-12 tenant-isolation program | No |
| F-13e: iamapi three-domain codegen bundle | contract | DEFER | — | M | module | No correctness defect; trigger: merge-conflict bottleneck | No |
| F-04-dead-loop: Dead restart loop in startOutboxWorker | async | DELETE | P3 | S | contained | Go dead code; simplicity-first | No |

---

## Prioritized Roadmap

### Wave 0 — P0 Prerequisites (execute before anything else; unblock all other waves)

| Item | Finding IDs | Smallest Correct Fix | REQ/RF IDs | Dependencies |
|---|---|---|---|---|
| Rotate credentials + delete seed binary + rewrite git history | F-18, D-06 | Delete `cmd/seed-test-document/`; rotate Postgres password; `git filter-repo` history scrub; `git rm bin/metaldocs-api.exe`; `.gitignore` fix; delete `scripts/start-spec1-api.ps1`; replace or SHA-256-pin `api-lint.exe` | REQ-SEC-1 | None — execute first |
| Add jobs.Dockerfile + compose service | F-19-deployment | Copy `worker.Dockerfile` → `jobs.Dockerfile`; add `jobs` service in `docker-compose.yml` with `depends_on: api (healthy)` | REQ-ASYNC-4 | None |
| Fix platform/observability → auth/domain import (prereq for middleware reorder) | F-06a | Replace `authdomain.CurrentUserFromContext` with a `func(*http.Request) string` callback injected at construction; identical pattern to `platform/ratelimit.Middleware` | REQ-TOP-2 | None; must land before F-01 |

**Code reduced:** 1 dead binary directory deleted (`cmd/`). `bin/metaldocs-api.exe` deleted. 2 dead scripts deleted. Net: significant dead artifact removal.

---

### Wave 1 — High-Value / Low-Blast (contained changes, high severity)

Target: eliminate the correctness and compliance defects that are each fewer than ~100 lines of change.

| Item | Finding IDs | Smallest Correct Fix | REQ/RF IDs | Dependencies |
|---|---|---|---|---|
| Middleware chain reorder + panic recovery + login rate limit | F-01 | Reorder `main.go:598-602` to `panicRecovery → httpObs → cors → originProtection → preAuthRateLimit(login) → authn → iam → presenceBump → rateLimiter → mux`; add `platform/middleware/recovery.go` (~40 lines); wire `platform/ratelimit` pre-auth instance for login | REQ-MW-1/2/4/5/7 | F-06a must land first |
| Server timeout fields | F-16 | Add `ReadTimeout: 30s`, `WriteTimeout: 60s`, `IdleTimeout: 90s` to `http.Server` at `main.go:613-617` (3 lines) | REQ-REL-1/2 | None |
| Delete spec2.yaml + internal/api/v2 + fix CI gate | F-03 | `git rm api/openapi/spec2.yaml`; migrate 3 contract tests from `apiv2.APIError` to `problem.Problem`; delete `internal/api/v2/`; delete or fix `capability-catalog-hash` CI job | REQ-API-2; RF-4 | None |
| Idempotency middleware raw string codes | F-09 | Add `CodeIdempotencyKeyInvalid`, `CodeRequestBodyTooLarge` to `codes.go`; use `CodeIdempotencyKeyReused` for conflict; replace `"INTERNAL"` with `CodeInternalError`; expand `guardedPackages` in catalog guard test | REQ-H-2; REQ-API-5; RF-10 | None |
| Dead search struct fields + camelCase param removal | F-13a, F-13b | Remove `subject`/`business_unit`/`classification`/`tags` from `SearchDocumentResponse`; remove `businessUnit` from `handler.go:99`; fix bare 405 in search (D-03 cross-finding) | REQ-API-4 | None |
| River dual migration ownership | F-19-dual-migration | Remove `MigrateRiverSchema` call from `bootstrap/jobs.go` (2 lines) | REQ-ASYNC-4 | F-19-deployment |
| lease_reaper JOIN bug | F-19-lease-reaper-bug | Remove `public.documents` subquery; emit system-scoped governance event or structured slog line for maintenance job lease reaps | REQ-ASYNC-4; ISO 9001 §8.5.1 | None |
| Templates audit read/write sink split | F-07-sub-split | Migrate `ListAudit` to query `audit_events` filtered by `resource_type='template'`; accept historical seam in `templates_audit_log` | REQ-ASYNC-1 | None |
| Empty scaffolds + dead gitkeep files | F-08 | `git rm` four `.gitkeep` files; `git rm bin/metaldocs-api.exe` (included in Wave 0) | REQ-TOP-3 | Wave 0 (bin/ covered there) |

**Code reduced:** `internal/api/v2/` package deleted (~60 lines). `spec2.yaml` deleted (~1061 lines). 4 dead struct fields removed. 1 dead query parameter removed. `bin/metaldocs-api.exe` deleted.

---

### Wave 2 — Structural Refactors (module-scoped or cross-module; each bounded, no architecture redesign)

Target: fix the cross-module boundary violations and compliance-grade atomicity gaps. Each item is a contained refactor within named boundaries.

| Item | Finding IDs | Smallest Correct Fix | REQ/RF IDs | Dependencies |
|---|---|---|---|---|
| Move post-commit audit/governance writes inside transactions (taxonomy, templates, documents-core) | F-07, D-01 | Add `LogTx(ctx, tx, event)` to GovernanceLogger interface and `RecordTx` to AuditWriter interface; replace post-commit calls in taxonomy (11 methods), templates lifecycle, documents-core service with in-tx calls; approval module is the canonical correct pattern (decision_service.go:545) | REQ-ASYNC-1; ISO 9001 §8.5.1; 21 CFR Part 11 §11.10(e) | None |
| RLS on controlled_documents + audit_events; type-safety fix | F-12 | Change `TenantID string` → `TenantID uuid.UUID` in `audit/domain/port.go:ExportJob` (S); add RLS migration for `controlled_documents` and `audit_events` using `current_setting('metaldocs.tenant_id')` GUC pattern (M); close T-008 as by-design with wiki note; defer `iam_users` trigger to RF-6 | REQ-TEN-1; REQ-DATA-2; OWASP ASVS V4.1.3 | ADR-A required |
| Fix capability string literals (correctness defect + tier mismatch) | F-11 | Decide intended cap for `upsertApprovalConfig` (replace `"template.admin"` literal); add 3 `EventType` constants in `events.go`; audit tier-1/tier-2 pairings across template lifecycle routes and fix mismatches; expand api-lint scope | REQ-AUTHZ-2; REQ-AUTHZ-5; ADR 0022 | ADR-B if new cap minted |
| Delivery-layer SQL in CD + documents handlers | F-06b | Extract inline `h.db.QueryRowContext` calls from `controlleddocuments/delivery/http/routes.go:281-317` and `documents/delivery/http/handler.go:512-536` to repository methods called via application service | REQ-H-1; REQ-TOP-1 | None |
| auth repo writing to iam_users | F-06c | Define narrow `LoginContextPort` interface in iam/domain; implement in iam/infra; auth calls port — or move write to iam/application invoked as service call | REQ-TOP-1; DDD Bounded Context | None |
| iam/delivery imports auth/infrastructure/postgres | F-06d | Promote `SessionAdminQuery`/`SessionListItem` to `auth/domain`; remove `authpg` import from `iam/delivery/http/sessions_handler.go` | REQ-TOP-1 | None |
| Delete security.RateLimiter; activate platform/ratelimit | F-05, D-04 | Delete `platform/security/ratelimit.go`; replace `security.NewRateLimiter` with `ratelimit.New` in `main.go`; call `RegisterRoutesWithRateLimit` | REQ-TOP-2; REQ-MW-5 | F-01 (rate limit wiring coordinates) |
| N+1 role queries + VerifyUserInTenant full scan + approval inbox load-in-loop | F-10 | Sequence: (a) `RolesByUserID` → single LEFT JOIN round trip; (b) add `RolesByUserIDs` batch variant; (c) `VerifyUserInTenant` → `EXISTS` point-lookup; (d) `ListPendingForActor` → batch `LoadInstancesByIDs`; (e) audit ILIKE → restrict to indexed columns action/actor_id/resource_id | REQ-DATA-2; REQ-AUTHZ-6 | F-10a is prerequisite for F-10b/c |
| InMemoryAuthFailureRateLimiter in production | F-20e | Replace with Postgres-backed counter table (no Redis required now); 5-line migration + 2 SQL statements | REQ-REL-3; OWASP ASVS §2.2.1 | None |
| Delete dead application code | F-14 | PR1: delete `CutoverService` + test, `CompositionConfig` + test, `AreaService.SetParent` + test, `resolvePermissionFallback`, `WorkerConfig.ReviewReminderDays`, `coverage_boost_test.go`; PR2: add nil-guard fail-loud in `DecisionService` constructor for `pdfDispatcher`; PR3: remove legacy column writes from templates INSERT | CWE-561; YAGNI | None |
| Delete deprecated govLogger nil fallback | F-07-sub-deprecated-logger | Remove `DBGovernanceLogger` from `controlleddocuments/module.go`; wire `AuditWriter` unconditionally | Go doc convention; simplicity-first | F-07 tx fix should land first |

**Code reduced:** `platform/security/ratelimit.go` deleted (~204 lines). 6 dead application types and test files deleted. Deprecated logger removed. Legacy INSERT columns removed.

---

### Wave 3 — Deferrable (real but trigger-gated or low-urgency)

| Item | Finding IDs | Smallest Correct Fix | Trigger |
|---|---|---|---|
| OTel SDK + W3C traceparent + Prometheus metrics endpoint | F-17 | `otelhttp.NewHandler` for HTTP middleware auto-instrumentation (~60 lines `platform/observability/otel.go`); OTLP export via `autoexport` (env-configured, no backend compiled in); replace custom `X-Trace-Id` with W3C `traceparent` propagation; expand `normalizeRoute` or use Go 1.22 `r.Pattern` | Infrastructure decision on OTLP backend |
| taxonomy.TemplateVersionChecker queries templates tables | F-06e | Define `TemplateVersionPort` interface in templates/domain; implement in templates/infra; taxonomy calls port | Next time templates or taxonomy is touched |
| parseBoolEnv semantic drift consolidation | F-15 | Export `ParseBoolEnv` from `platform/config` with 4-value POSIX semantics; update 2 callers | Next time either caller is touched |
| Simplify duplicate staging outbox worker/repo (render pipeline) | F-04 | Generic `StagingOutboxWorker[R]` and `StagingOutboxRepository[R]`; delete dead restart loop in `startOutboxWorker` | Next time render/fanout is touched |
| MinIO client consolidation (3 → 2) | D-02 | Change `miniostore.NewStore` to accept `*minio.Client`; pass internal client from `buildMinioClients` | Next time bootstrap wiring is touched |
| CachedRoleProvider cache contract documentation | D-05 | Add `// CacheContract:` doc block to `cached_role_provider.go` (15–20 lines, no code change) | Before next IAM module feature work |
| DocumentStatus enum completeness | D-07 | Add 6 missing `DocumentStatus` constants to `documents/domain/model.go`; update string-literal callsites in documents module | Next documents domain touch |
| OpenAPI partial files deletion + deprecated:true on createManagedUser | F-13c, F-13d | `git rm` 3 partial files; add `deprecated: true` to `createManagedUser` in spec + regen | Next spec regen pass |
| Sequential readiness probe checks | F-16C | Replace sequential loop in `observability/runtime.go:281-325` with shared-budget concurrent checks via errgroup | When a second dependency (beyond Gotenberg) is added |
| WebSocket presence drain on SIGTERM | F-16B | Hub connection tracking + `server.RegisterOnShutdown` close signal | If K8s termination grace period proves insufficient |
| CD leading-wildcard ILIKE | F-20b | `pg_trgm` GIN index + rewrite search to use trigram operator | When p95 on CD list >200ms or table >50k rows |
| idempotency_keys.tenant_id FK | F-09d | Migration adding FK constraint | When F-12 RLS program runs |
| iamapi three-domain codegen bundle split | F-13e | Split `iam/api/cfg.yaml` into `iam`, `audit`, `security` configs | When merge-conflict bottleneck appears |
| TTL string constant in postgres_store.go | F-09c | `const idempotencyTTL = "'24 hours'"` (1 line) | Next postgres_store.go touch |

---

## ADR Backlog

Continuing from the highest existing ADR number (`0026-unified-authz-enforcement.md`):

**ADR-A: ADR 0027 — auth_identities tenant-global by design; RLS adoption sequencing**
Records that `auth_identities` has no `tenant_id` by deliberate design (identity is tenant-global; scoping is via JOIN to `iam_users`). Closes T-008 as by-design. Documents the RLS rollout sequence: `controlled_documents` and `audit_events` in the first migration; `iam_users` deferred to the RF-6 authz tripwire program; other tables sequenced by risk tier. Required before executing the F-12 RLS migration so future reviewers understand the partial coverage.

**ADR-B: ADR 0028 — intended capability for upsertApprovalConfig (conditional)**
Required only if the F-11 fix introduces a new `CapTemplateManage` capability constant. If the fix reuses an existing capability (e.g., `CapTemplateArchive`), no ADR is needed — the code change and a lock test are sufficient per ADR 0016 precedent. The ADR, if needed, documents the capability name, its tier-1/tier-2 grant semantics, and the route(s) it governs.

---

## Sources

- Theme artifacts: `wiki/backend/_artifacts/stage2/security-secrets.md`, `observability-middleware.md`, `async-integrity.md`, `contract-surface.md`, `module-boundaries.md`, `performance-scale.md`, `dead-code-hygiene.md`
- Input register: `wiki/backend/legacy-register.md`
- Normative requirements: `wiki/architecture/backend-target-architecture.md` (REQ-* / RF-*)
- Vocabulary: `wiki/standards/backend-canon.md`
- Maturity grades: `wiki/architecture/backend-blueprint.md`
- ADR chain: `wiki/decisions/` (highest existing: 0026)
