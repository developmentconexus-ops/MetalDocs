# Stage-1 Audit Artifact — Repo Topology

> **Area:** Repository topology: build wiring, scripts, CI, orphan/legacy candidates
> **Last verified:** 2026-06-10
> **Scope:** Top-level directory layout, all binaries built and run, Go module configuration, CI pipeline shape, script entry points, and explicit orphan/legacy classification for every top-level directory. Does not descend into frontend/, node_modules/, vendor/, .worktrees/, .clone/, or non_git/ beyond identification.

---

## 1. Identity & purpose

MetalDocs is a **Go modular-monolith** with a companion Node.js micro-service. The repository is a single Go module (`module metaldocs`, `go.mod:1`) containing all backend source; there is no `go.work` workspace. The Go module produces **four binaries** from paths under `apps/` and one legacy seed tool under `cmd/`. A fifth runtime — `apps/docx-renderer` — is a Node.js (Fastify) service built and containerized separately from the Go build.

The repository is a **mixed-language monorepo**: Go owns all backend logic, TypeScript owns the docx-renderer service and all frontend packages, and SQL owns all schema migrations. Infrastructure is defined in `deploy/compose/docker-compose.yml` and driven in development by PowerShell scripts under `scripts/`. CI is GitHub Actions (`.github/workflows/`), with 14 workflow files enforcing correctness, contract, security, and release readiness gates.

---

## 2. File inventory

### Root-level files

| Path | Role |
|---|---|
| `go.mod` | Single Go module declaration: `module metaldocs`, Go 1.25, all dependency versions pinned |
| `go.sum` | Dependency checksum database |
| `tools.go` | `//go:build tools` file pinning `oapi-codegen` to `go.mod` so `go generate` resolves it |
| `staticcheck.conf` | staticcheck configuration — disables ST1000/ST1003/ST1005/ST1016/ST1020-22, all SA* checks on |
| `redocly.yaml` | Redocly OpenAPI lint config; points to `api/openapi/v1/openapi.yaml`; silences `operation-summary`, `security-defined`, `struct` for pre-existing debt |
| `Makefile` | Thin wrapper: `up/down/logs` via `deploy/compose/docker-compose.yml`, `test`/`test-watch` run frontend Vitest |
| `package.json` | npm workspace root; workspaces: `apps/docx-renderer`, `packages/*`; scripts: `build/test/typecheck:docx-v2` |
| `pnpm-lock.yaml` | pnpm lockfile for the frontend workspace (frontend uses pnpm; root/docx-renderer workspace uses npm) |
| `package-lock.json` | npm lockfile for root/docx-renderer workspace |
| `CLAUDE.md` | Agent operating instructions (not deployed) |
| `AGENTS.md` | Agent routing guide (not deployed) |
| `README.md` | Developer quick-start (not deployed) |

### apps/ — runnable application packages

#### apps/api/cmd/metaldocs-api/ — main API binary
| File | Role |
|---|---|
| `main.go` (943 lines) | Composition root: config load, bootstrap, all module wiring, middleware chain assembly, HTTP server |
| `permissions.go` (286 lines) | `newPermissionResolver()` and `newPublicPathChecker()` — route-to-capability mapping table |
| `reauth.go` (52 lines) | Re-authentication helper adapter wiring for sensitive approval signoff |
| `main_test.go` | Startup smoke test |
| `permissions_test.go` | Unit tests for permission resolver |
| `permissions_authz_scope_test.go` | Authz scope binding tests |
| `e2e_gate_test.go` | E2E gate assertions |
| `metaldocs-api.exe` | Compiled Windows dev binary (gitignored; present locally) |

#### apps/api/cmd/metaldocs-e2e-seed/ — E2E seed binary
| File | Role |
|---|---|
| `main.go` | Seeds a dev/E2E tenant with admin and approver users via `bootstrap.BuildAPIDependencies`; invoked by `scripts/e2e-seed.ps1` |

#### apps/api/internal/wiring/ — API-local wiring helpers
| File | Role |
|---|---|
| `documents.go` | `NewCapabilityChecker` — bridges `iamapp.CapabilityService` to the documents module interface |

#### apps/worker/cmd/metaldocs-worker/ — outbox-poll worker binary
| File | Role |
|---|---|
| `main.go` | Poll-loop worker: loads `workerapp.Service`, wires `PDFJobRunner` (Gotenberg) and `MaterializeJobRunner` (fanout/eigenpal) when deps are configured; exits clean on context cancel |
| `main_test.go` | Unit tests for run-loop helpers |

#### apps/jobs/cmd/metaldocs-jobs/ — River-based jobs binary
| File | Role |
|---|---|
| `main.go` | Starts `river.Client` processing the `temporal` queue; wires `SchedulerService` via `approvaljobs.NewWorkers`; graceful shutdown on SIGTERM with 15-second drain timeout |

#### apps/docx-renderer/ — Node.js DOCX/PDF rendering service
| Path | Role |
|---|---|
| `Dockerfile` | Node.js multi-stage build; owned by CI `ci.yml` (docx-renderer job) |
| `build.mjs` | esbuild bundler script; outputs CJS bundle |
| `package.json` | Fastify app; dependencies: minio-js, docxtemplater (eigenpal), @aws-sdk/client-s3 |
| `src/index.ts` | Fastify server entrypoint |
| `src/render/` | Core rendering logic: eigenpal template fill-in + DOCX generation |
| `src/routes/` | HTTP route handlers (`/render/fanout`, `/render/reconstruct`, etc.) |
| `src/s3.ts` | MinIO/S3 client wiring |
| `src/service-auth.ts` | `X-Service-Token` validation middleware |
| `src/env.ts` | Typed env-var config |
| `test/` | Vitest integration tests for docx-renderer |
| `vendor/` | Bundled eigenpal tarball (out of Go vendor to avoid confusion) |
| `vitest.config.ts` | Vitest config |
| `tsconfig.json` | TypeScript config |

### cmd/ — legacy root-level command

| Path | Role |
|---|---|
| `cmd/seed-test-document/main.go` | One-shot seed binary: creates a synthetic DOCX file and inserts it directly into Postgres + MinIO via hardcoded DSN constants (`dsn = "host=127.0.0.1 port=5433..."`). No CI reference found; never invoked by canonical scripts. |

### internal/ — shared Go source (not a binary)

#### internal/modules/ — 11 business modules
| Module | Role |
|---|---|
| `audit/` | Immutable audit event log, hash-chain integrity, export jobs |
| `auth/` | Login, session, JWT, credential management, admin bootstrap |
| `controlleddocuments/` | Controlled-document registry lifecycle |
| `documents/` | Core document editing, approval pipeline, freeze/materialize |
| `iam/` | Users, roles, capabilities, area memberships, presence, observability |
| `jobs/` | Scheduler framework, watchdog jobs, idempotency janitor, audit integrity validator |
| `render/` | Fanout client to docx-renderer, PDF outbox, materialize outbox, resolvers |
| `search/` | Cross-module full-text search (PG-backed) |
| `security/` | MFA coverage, security settings service |
| `taxonomy/` | Document profiles, process areas, document families |
| `templates/` | Template CRUD, versioning, DOCX import, publish state |

#### internal/platform/ — 26 cross-cutting packages
| Package | Role |
|---|---|
| `authn/` | JWT/session config, validation, context helpers |
| `bootstrap/` | `BuildAPIDependencies`, `BuildWorkerDependencies`, `BuildJobsDependencies`, `MigrateRiverSchema` |
| `cache/` | **Empty placeholder** — contains only `.gitkeep`; no Go source |
| `config/` | Typed env-var config loaders for CORS, rate-limit, attachments, feature-flags, jobs, worker |
| `db/postgres/` | pgx pool factory, connection helpers |
| `docgenv2/` | Template/snapshot readers for the fanout pipeline |
| `featureflags/` | Feature-flag config + HTTP handler (`GET /api/v1/feature-flags`) |
| `formval/` | JSON Schema form validation wrapper (gojsonschema) |
| `httpclient/` | `NewInternalClient()` — tuned HTTP/2 client with timeouts; no embedded retry |
| `httpresponse/` | Shared HTTP response helpers |
| `idempotency/` | Idempotency-key middleware and two-phase store |
| `jobs/river/` | River queue client bundle factory (`NewClientBundle`) |
| `messaging/` | `Publisher` interface + `noop/` and `outbox/` implementations; no servicebus adapter wiring in Go |
| `migrate/` | `migrate.Apply()` — forward-only migration runner reading `db/migrations/*.sql` |
| `objectstore/` | MinIO client factory, `DocumentPresigner` (presigned PUT/GET) |
| `observability/` | Metrics, tracing, structured logging, `NewHealthHandler`, `NewHTTPObservability` |
| `pagination/` | Cursor + offset pagination helpers |
| `problem/` | RFC 9457 `application/problem+json` builder with closed vocabulary |
| `ratelimit/` | Token-bucket rate limiter (wraps `golang.org/x/time/rate`) |
| `render/gotenberg/` | HTTP client for Gotenberg PDF conversion |
| `requesttrace/` | Request-ID propagation middleware |
| `security/` | CORS, origin protection, rate-limiter wiring |
| `servicebus/` | `GotenbergPDFConverter` — Gotenberg HTTP client used as PDF adapter |
| `sqlescape/` | SQL identifier escaping (values always via pgx params; identifiers via this) |
| `storage/minio/` | MinIO object store adapter |
| `tenant/` | `DevTenantID` constant + tenant-context helpers |
| `useragent/` | User-agent parsing helper |
| `worker/` | `workerapp.Service` — outbox-poll loop, PDF runner, materialize runner |

#### internal/test/ — E2E test helpers (build-tag gated)
| File | Role |
|---|---|
| `e2e_seed.go` | `RegisterE2EHandlers` — mounts destructive reset/governance endpoints; gated by `METALDOCS_E2E=1` |
| `e2e_seed_stub.go` | Build-tag stub for non-E2E builds |
| `e2e_clock_offset_nonprod.go` | Controllable clock for E2E (non-prod build tag) |
| `e2e_clock_offset_production.go` | Real clock for production build tag |

#### internal/testsupport/pgtest/ — Postgres test fixtures
| File | Role |
|---|---|
| `pgtest.go` | `NewTestDB()` — spins up an isolated Postgres database for integration tests via `METALDOCS_DATABASE_URL` |

#### internal/api/v2/ — spec2 generated types
| File | Role |
|---|---|
| `types_gen.go` | Handwritten response types for the spec2 surface (`ProfileResponse`, `AreaResponse`, `ControlledDocumentResponse`, etc.) |
| `contract_test.go` | Contract assertion tests for the above types |

### api/ — OpenAPI contract source

| Path | Role |
|---|---|
| `api/openapi/v1/openapi.yaml` | Primary contract source of truth; assembled by oapi-codegen |
| `api/openapi/v1/partials/` | Partial YAML fragments included by the main spec |
| `api/openapi/spec2.yaml` | Secondary/legacy spec targeting an alternate surface; last touched 2026-05-06 (commit `e1944bc4a` purged v2 routes) |

### db/ — database lifecycle artifacts

| Path | Role |
|---|---|
| `db/migrations/` | 233 forward-only SQL migration files (0001–0233); `platform/migrate` applies these at API startup |
| `db/baseline/0001_current_schema.sql` | Curated schema snapshot representing the baseline state |
| `db/prerequisites/0001_extensions.sql` | PG extension setup (run once before migrations) |
| `db/dev-seeds/0001_local_dev_seed.sql` | Local dev seed data |
| `db/reference-data/0001_product_reference_data.sql` | Product reference data applied via bootstrap |

### sql/ — supplementary SQL artifacts

| Path | Role |
|---|---|
| `sql/contracts/state_transitions.yaml` | State-machine contract definitions referenced by `scripts/api-lint` |

### scripts/ — tooling and dev scripts

| Path | Role |
|---|---|
| `start-api.ps1` | **Canonical API startup** (CLAUDE.md §1); builds metaldocs-api.exe/metaldocs-worker.exe/metaldocs-jobs.exe if stale; starts all three; tees output to `logs/api.log` |
| `start-worker.ps1` | Standalone worker startup (simpler, no staleness check) |
| `start-jobs.ps1` | Standalone jobs host startup with Access Denied fallback to `go run` |
| `api-lint/` | Custom Go linter program: `main.go`, `spec_rules.go`, `code_rules.go`, `registry_rules.go`, tests, testdata; invoked by `api-contract.yml` CI |
| `dev-bootstrap.ps1` | DB bootstrap helper for dev |
| `dev-bootstrap-baseline.ps1` | Baseline-specific bootstrap |
| `dev-migrate.ps1` | Run migrations against local DB |
| `dev-db-reset.ps1` | Reset local DB |
| `dev-api.ps1`, `dev-api-web.ps1`, `dev-api-perf.ps1` | Alternate dev launch wrappers |
| `dev-local.ps1` | Full local stack launcher |
| `dev-docx-renderer.ps1` | Start docx-renderer for local development |
| `e2e-seed.ps1` | Runs `go run ./apps/api/cmd/metaldocs-e2e-seed` |
| `e2e-smoke.ps1` | Runs E2E smoke against local API |
| `test.ps1` | Runs Go test suite |
| `tidy.ps1` | `go mod tidy` wrapper |
| `check-governance.ps1` | Contract-change governance rules (must update OpenAPI when delivery/http changes) |
| `check-module-boundaries.ps1` | Enforces no cross-module imports at wrong layer |
| `check-module-contract-sync.ps1` | Checks OpenAPI/module sync |
| `check-db-bootstrap.ps1` | Validates DB bootstrap state |
| `check-db-dictionary-coverage.ps1` | Validates schema dictionary coverage |
| `check-baseline-equivalence.ps1` | Validates baseline equivalence |
| `check-release-v2-names.ps1` | Release naming validation |
| `check-system-runnable.ps1` | System runnability gate |
| `contract-baseline.ps1` | Contract baseline capture |
| `export-schema-baseline.ps1` | Export schema baseline |
| `openapi-lint-local.ps1` | Local Redocly lint runner |
| `phase3-hardening-gate.ps1` | Phase 3 security hardening gate (gosec + govulncheck) |
| `phase3-release-readiness.ps1` | Full release readiness gate |
| `perf/` | k6 performance scripts (`k6-baseline.js`, `k6-light-concurrency.js`, `k6-write-concurrency.js`) |
| `sql/` | SQL utility scripts: `create-backup-role.sql`, `dev_reset_legacy_document_registry.sql`, `perf_db_query_analysis.sql` |
| `backup-postgres.ps1`, `restore-postgres.ps1`, `validate-backup.ps1`, `run-backup-restore-gate.ps1` | Postgres backup/restore tooling |
| `security-baseline.ps1` | Security baseline capture |
| `seed-system-blank-template.ps1` | Seeds the system blank template |
| `bench_authz.sh` | Authz benchmark shell script |
| `classify-test-failure.sh` | Test failure classifier |
| `dump-error-codes.go` | Go script to dump error code vocabulary |
| `cleanup_test_documents.sql` | Test data cleanup SQL |
| `verify-triggers.sql` | DB trigger verification SQL |
| `axe-diff.mjs` | Axe accessibility diff tool (frontend) |
| `build_dc_template.py` | Python helper for DC template building |
| `start-api-no-build.ps1` | Start API without build step |
| `start-api-planc.ps1` | Legacy: starts API on port 8083 for a worktree (last touched 2026-04-06) |
| `start-spec1-api.ps1` | Legacy: hardcoded worktree path on old machine hostname; dead |
| `w5-preflight-dump.sh`, `w5-rollback.sh` | Phase-5 preflight/rollback helpers |
| `docx-v2-seed-minio.sh`, `docx-v2-verify-migrations.sh` | docx-v2 seeding helpers |
| `spec2_phase1_smoke.sh` | spec2 phase 1 smoke test (legacy) |

### .github/workflows/ — CI gates

| Workflow | Trigger | Purpose |
|---|---|---|
| `ci.yml` | PR on docx-renderer/templates/featureflags paths | docx-renderer Node.js build + Go test subset |
| `test-smoke.yml` | PR → main/develop | Integration smoke: `TestTriggerBypass`, `TestMembership`, `TestSchemaLockdown`, `TestLegacy`, `TestE2E` with Postgres service |
| `test-full.yml` | Push to main | Full integration suite with race detector (`-tags integration -count=1 -race -timeout 600s`) |
| `test-nightly.yml` | Daily 02:00 UTC | Nightly stress: `INTEGRATION_STRESS_N=500`, 3600s timeout; opens GitHub issue on failure |
| `api-contract.yml` | PR touching API/codegen paths + push to main | Four jobs: backend codegen drift, frontend codegen drift, Redocly OpenAPI lint, api-lint PATH-BASE-PREFIX gate, api-lint design-system (BLOCKING) |
| `invariants.yml` | PR + push to main | cilint custom analyzers (SARIF), migration gapless sequence check, capability catalog SHA-256 integrity, staticcheck + go vet |
| `golangci-lint.yml` | PR + push to main | golangci-lint v2.11 on `apps/api/...` + `internal/...` + `tools/...`, diff-only on PRs |
| `governance-check.yml` | PR → main | Contract governance (`check-governance.ps1`), docx-v2 CK5 isolation guard |
| `module-boundaries.yml` | PR → main | `check-module-boundaries.ps1` cross-module import enforcement |
| `api-contract.yml` | PR/main | OpenAPI codegen drift (BE + FE), Redocly lint, api-lint |
| `e2e-coverage-gate.yml` | PR touching approval/documents paths + push to main | E2E coverage map check, axe baseline integrity, E2E Playwright smoke |
| `perf.yml` | PR touching approval/jobs paths + push to main | k6 perf benchmarks (reduced on PR, full on main) |
| `phase3-hardening-gate.yml` | PR → main | gosec + govulncheck gate |
| `release-readiness.yml` | Manual dispatch | Full release readiness (`phase3-release-readiness.ps1`) + artifact upload |
| `smoke.yml` | Schedule (every 5/10 min) + manual | Production and staging synthetic smoke probes with PagerDuty P1 alert on prod failure |
| `supply-chain.yml` | Push to main/tags, weekly, PR touching go.mod/package.json | SBOM (Syft, tag-only), CVE scan (Grype fail-build=high), Dependabot label |

### deploy/ — container and infrastructure definitions

| Path | Role |
|---|---|
| `deploy/compose/docker-compose.yml` | Full stack: postgres:16, redis:7, minio, minio-init, gotenberg:8, docx-renderer (local build) |
| `deploy/docker/api.Dockerfile` | Multi-stage Go build: `./apps/api/cmd/metaldocs-api`; copies `db/migrations/` into image; exposes 8081 |
| `deploy/docker/worker.Dockerfile` | Multi-stage Go build: `./apps/worker/cmd/metaldocs-worker`; no migration copy |
| `deploy/nginx/nginx.conf` | Reverse-proxy / TLS termination config |

### docker/ — additional container definitions

| Path | Role |
|---|---|
| `docker/gotenberg/Dockerfile` | Custom Gotenberg image with Carlito font |
| `docker/gotenberg/verify-carlito.sh` | Font verification script |

### ops/ — operational tooling

| Path | Role |
|---|---|
| `ops/CAPABILITY_CATALOG.sha256` | SHA-256 pin of `sql/seeds/capabilities_v2.sql` (currently contains placeholder string `placeholder-hash-update-after-catalog-created`; the catalog file itself does not exist) |
| `ops/DEPLOY.md` | Deployment runbook |
| `ops/smoke/healthz.sh` | Healthz probe script (used by `smoke.yml`) |
| `ops/smoke/approval_roundtrip.sh` | Synthetic approval roundtrip smoke (used by `smoke.yml`) |
| `ops/chaos/SCENARIOS.md`, `ops/chaos/kill_scheduler.sh` | Chaos engineering scenarios |
| `ops/incident/RUNBOOK.md`, `ops/incident/freeze.sh` | Incident response runbook |
| `ops/dashboards/approval.json` | Grafana/observability dashboard definition |

### tests/ — test suites

| Path | Role |
|---|---|
| `tests/integration/scenarios/` | Integration test scenarios (trigger bypass, membership, schema lockdown, e2e happy path, concurrency, idempotency, outbox) |
| `tests/integration/testdb/` | `NewTestDB()` helper (wraps `internal/testsupport/pgtest`) |
| `tests/integration/fixtures/` | `seed.go` — fixture seeding |
| `tests/integration/approval/`, `documents/`, `iam/`, `migrations/` | Module-level integration suites |
| `tests/unit/` | Unit-level test files (cross-cutting: auth, IAM, CORS, origin protection, rate limit) |
| `tests/docx_v2/` | docx-v2 integration tests: documents, exports, templates, scaffold smoke |
| `tests/contract/` | Contract tests (last activity: 2026-06-07 commit `ae8d5e4e5`) |
| `tests/e2e/` | Empty (no files) |

### tools/ — custom Go tooling

| Path | Role |
|---|---|
| `tools/cilint/main.go` | Custom Go linter binary; runs `analyzers.go`, `legacyvocab.go`, `outboxpair.go`, `txownership.go` on `./...`; outputs SARIF; used by `invariants.yml` |
| `tools/cilint/internal/analyzers/` | Five custom AST analyzers enforcing architecture invariants |
| `tools/perfbench/` | k6 performance benchmark scripts: `submit.js`, `signoff.js`, `publish.js`, `scheduler_tick.js`, `thresholds.json` |

### packages/ — frontend npm packages

All under `packages/` are TypeScript packages for the frontend; outside this audit's scope beyond identification.

| Package | Role |
|---|---|
| `packages/docx-editor/` | DOCX editor React component library |
| `packages/editor-ui/` | Editor UI component library |
| `packages/form-ui/` | Form UI component library |
| `packages/shared-tokens/` | Shared design tokens |
| `packages/shared-types/` | Shared TypeScript types |

### shared/ — cross-language shared packages

| Path | Role |
|---|---|
| `shared/mddm-layout-tokens/` | TypeScript layout tokens (npm package; used by docx-renderer + frontend) |
| `shared/mddm-pagination-types/` | TypeScript pagination transport types with compile-time contract assertions |
| `shared/schemas/` | MDDM JSON Schema (`mddm.schema.json`) + `embed.go` (embeds the schema into Go); TypeScript canonicalize/validate scripts |

### archive/ — historical artifacts

| Path | Role |
|---|---|
| `archive/migrations/` | Legacy pre-baseline migration tree (0001–0142b); archived by commit `266bd4132` 2026-05. No longer applied by the migration runner. |

### bin/ — local build artifacts

| Path | Role |
|---|---|
| `bin/metaldocs-api.exe` | Stale local build artifact; gitignored pattern covers `*.exe` but this specific path is committed (last git touch: initial commit `912879cba`); functionally dead |

### backups/ — database backups

Gitignored (`backups/` in `.gitignore`). Contains local Postgres dump files (`*.dump`) used with `scripts/backup-postgres.ps1` / `scripts/restore-postgres.ps1`. Not tracked in git.

### non_git/ — untracked evidence and release artifacts

Gitignored. Used by `release-readiness.yml` to collect evidence JSON. No git history.

### Top-level log files (untracked, development noise)

`api-bg.err.log`, `api-bg.out.log`, `api-dev.log`, `worker.err.log`, etc. — local dev run log files, gitignored. Not part of the committed artifact.

---

## 3. Public surface

This area (repo topology) has no exported Go types or HTTP routes of its own. The public surface it _governs_ is:

**Binaries produced:**

| Binary | Build path | Canonical build command |
|---|---|---|
| `metaldocs-api.exe` | `./apps/api/cmd/metaldocs-api/...` | `scripts/start-api.ps1` (timestamp-stale check + `go build`) |
| `metaldocs-worker.exe` | `./apps/worker/cmd/metaldocs-worker/...` | `scripts/start-api.ps1` (co-started), `scripts/start-worker.ps1` |
| `metaldocs-jobs.exe` | `./apps/jobs/cmd/metaldocs-jobs/...` | `scripts/start-api.ps1` (co-started), `scripts/start-jobs.ps1` |
| `metaldocs-e2e-seed` (no binary on disk) | `./apps/api/cmd/metaldocs-e2e-seed` | `scripts/e2e-seed.ps1` via `go run` |
| `seed-test-document` (no binary on disk) | `./cmd/seed-test-document/...` | No canonical script; built locally as `seed-test-document.exe` |
| `api-lint` (scripts/api-lint/api-lint.exe) | `./scripts/api-lint/` | `go run ./scripts/api-lint/` in CI; locally via `go test ./scripts/api-lint/...` |
| `cilint` (via `go run`) | `./tools/cilint/` | `go run ./tools/cilint ./...` in CI |
| `docx-renderer` (Node.js) | `apps/docx-renderer/` | `npm run build:docx-v2` → `docker build` via `deploy/docker/` or `deploy/compose/docker-compose.yml` |

**Docker images built:**
- `deploy/docker/api.Dockerfile` → API container (EXPOSE 8081, runs migrations at startup)
- `deploy/docker/worker.Dockerfile` → Worker container
- `docker/gotenberg/Dockerfile` → Custom Gotenberg with Carlito font
- `apps/docx-renderer/Dockerfile` → docx-renderer Node.js service

---

## 4. Logic flows

### Flow 1: Binary build and startup (local development)

1. Developer runs `scripts/start-api.ps1` (`start-api.ps1:1-354`)
2. Script reads `.env` and exports env vars into the process (`start-api.ps1:235-239`)
3. Script calls `Get-BinaryFreshness` for each binary — compares binary mtime against all source files under critical paths (`apps/api/cmd/metaldocs-api`, `internal/modules`, `internal/platform`, `db`, `scripts/start-api.ps1`) (`start-api.ps1:66-91`)
4. If stale or missing, script calls `go build -o metaldocs-api.exe ./apps/api/cmd/metaldocs-api/...` (`start-api.ps1:100-106`)
5. Same check/build cycle for worker (`./apps/worker/cmd/metaldocs-worker/...`) and jobs (`./apps/jobs/cmd/metaldocs-jobs/...`) unless `-NoWorker` / `-NoJobs` flags are set (`start-api.ps1:278-296`)
6. Worker and jobs binaries are launched in the background as separate Windows processes (`start-api.ps1:302-322`)
7. API binary runs in the foreground with stdout/stderr tee'd to `logs/api.log` (`start-api.ps1:328-353`)
8. Port 8081 is set via `APP_PORT=8081` env var; the Go binary reads `os.Getenv("APP_PORT")` at `main.go:605`

### Flow 2: API binary composition root startup

1. `main()` at `apps/api/cmd/metaldocs-api/main.go:148` begins signal context setup
2. Config is loaded: `config.RepositoryMode()`, `config.LoadRateLimitConfig()`, `config.LoadCORSConfig()`, `config.LoadAttachmentsConfig()`, `authn.LoadRuntimeConfig()`, `config.LoadFeatureFlagsConfig()` (`main.go:152-178`)
3. `bootstrap.BuildAPIDependencies(ctx, repoMode, attachmentsCfg)` wires all infrastructure: PG pool, MinIO client, Redis (if configured), audit repos (`main.go:180-184`)
4. `migrate.Apply(ctx, deps.SQLDB, "db/migrations", slog.Default())` runs all pending migrations forward (`main.go:191-194`)
5. All module services are instantiated (auth → audit → search → IAM → taxonomy → controlled-documents → documents → templates → approval → jobs scheduler) (`main.go:196-560`)
6. Middleware chain assembled at `main.go:598-602`:
   - `cors.Wrap(originProtection.Wrap(authMiddleware.Wrap(iamMiddleware.Wrap(presenceWrapped))))`
   - where `presenceWrapped = presenceBump.Wrap(httpObs.Wrap(rateLimiter.Wrap(mux)))`
   - Outermost-first: CORS → origin protection → AuthN middleware → IAM middleware → presence bump → HTTP observability → rate limiter → router mux
7. In-process scheduler goroutine starts (`main.go:563-566`) — runs `stuck-instance-watchdog`, `idempotency-janitor`, `audit-integrity-validator`, `lease-reaper` on ticker intervals
8. Outbox worker goroutines start for PDF and materialize outboxes (`main.go:488-492`)
9. HTTP server starts on `:8081`; graceful shutdown on SIGTERM drains in-flight requests

### Flow 3: Docker image build (CI/production)

1. `deploy/docker/api.Dockerfile` uses `golang:1.25-alpine` as builder
2. `go build -o /out/metaldocs-api ./apps/api/cmd/metaldocs-api` produces a static binary (`api.Dockerfile:6`)
3. Final image is `alpine:3.21` + `ca-certificates`; `db/migrations/` is copied into `/app/db/migrations/` so the binary can run migrations at container start (`api.Dockerfile:10-14`)
4. Worker Dockerfile follows the same pattern but without copying migrations (`worker.Dockerfile:1-11`)

### Flow 4: CI gate sequence on a PR

1. `golangci-lint.yml` runs diff-only lint over `apps/api/...`, `internal/...`, `tools/...`
2. `invariants.yml` runs: `cilint` custom analyzers (SARIF upload + hard fail), migration gapless sequence check, capability catalog SHA-256 check, staticcheck + go vet
3. `module-boundaries.yml` runs `check-module-boundaries.ps1` — scans all `internal/modules/**/*.go` for cross-module imports at wrong layers
4. `governance-check.yml` runs `check-governance.ps1` — if `internal/modules/.*/delivery/http/.*\.go` changed, OpenAPI must also have changed
5. `api-contract.yml` runs: backend codegen drift (`go generate ./...` + `git diff api.gen.go`), frontend codegen drift, Redocly lint, PATH-BASE-PREFIX gate, full api-lint design-system BLOCKING gate
6. `test-smoke.yml` runs targeted integration test subset against a live Postgres service container
7. `phase3-hardening-gate.yml` runs gosec + govulncheck

### Flow 5: Outbox worker polling (metaldocs-worker)

1. `metaldocs-worker` starts, loads `config.LoadWorkerConfig()`, calls `bootstrap.BuildWorkerDependencies` (`apps/worker/cmd/metaldocs-worker/main.go:55-63`)
2. `workerapp.NewService(deps.Consumer, workerCfg)` creates the poll loop service
3. If `METALDOCS_FANOUT_URL` is configured and SQLDB is available, a `MaterializeJobRunner` is wired to call the docx-renderer `/render/fanout` endpoint and write results back to Postgres (`worker/main.go:73-97`)
4. Ticker fires every `PollIntervalSeconds` seconds; `runWorkerBatch` calls `workerSvc.RunOnce(ctx, batchSize)` each tick (`worker/main.go:106-111`)
5. Worker exits cleanly on context cancellation

---

## 5. Dependencies

### Outbound (what repo topology / build wiring imports)

**Go module direct dependencies** (from `go.mod`):

| Dependency | Purpose |
|---|---|
| `github.com/jackc/pgx/v5` | Postgres driver (pgx pool) |
| `github.com/lib/pq` | Legacy `database/sql` Postgres driver (used in seed tools) |
| `github.com/minio/minio-go/v7` | MinIO/S3 object store client |
| `github.com/riverqueue/river` + `riverdatabasesql` | Durable job queue (River) for scheduled publish |
| `github.com/oapi-codegen/oapi-codegen/v2` | OpenAPI codegen (pinned via `tools.go`) |
| `github.com/oapi-codegen/runtime` | oapi-codegen runtime types |
| `github.com/getkin/kin-openapi` | OpenAPI spec parsing (used by api-lint) |
| `github.com/santhosh-tekuri/jsonschema/v6` | JSON Schema validation (formval) |
| `golang.org/x/crypto` | Argon2 / bcrypt password hashing |
| `golang.org/x/sync` | `errgroup` / synchronized work |
| `golang.org/x/time` | Rate-limit token bucket |
| `golang.org/x/text` | Text normalization |
| `gopkg.in/yaml.v3` | YAML parsing (config, api-lint) |
| `github.com/google/uuid` | UUID generation |
| `github.com/DATA-DOG/go-sqlmock` | SQL mock for tests |

**Node.js workspace** (from `package.json`): workspaces `apps/docx-renderer`, `packages/*`; CI uses npm; frontend uses pnpm (separate lockfile).

### Inbound (who uses these top-level wiring files)

The binaries and scripts in this area are the final consumers — nothing in the Go codebase imports `apps/api/cmd/metaldocs-api`, `apps/worker/cmd/metaldocs-worker`, or `apps/jobs/cmd/metaldocs-jobs`. They are build roots.

`internal/api/v2` is imported by (verified via grep):
- `internal/modules/controlleddocuments/delivery/http/routes_contract_test.go`
- `internal/modules/iam/delivery/http/routes_memberships_contract_test.go`
- `internal/modules/taxonomy/delivery/http/routes_profiles_contract_test.go`

`shared/schemas/embed.go` is a Go file in an otherwise TypeScript package — it imports `embed` and uses `//go:embed mddm.schema.json`; it is consumed by the Go backend modules that need the MDDM schema at runtime.

---

## 6. Persistence

The repo topology area is stateless in itself. It governs the artifacts that _drive_ persistence:

- `db/migrations/` (233 SQL files) is the migration corpus; `platform/migrate.Apply()` reads it
- `db/baseline/0001_current_schema.sql` is the curated schema snapshot
- `db/prerequisites/0001_extensions.sql` runs `CREATE EXTENSION IF NOT EXISTS` for PG extensions
- `db/dev-seeds/`, `db/reference-data/` are seeded by bootstrap scripts
- The `deploy/docker/api.Dockerfile` copies `db/migrations/` into the container so the API binary can run migrations at startup without a mounted volume

---

## 7. Config & environment

The `start-api.ps1` script sources `.env` and sets env vars for the process. The canonical env vars consumed by the Go binaries (from `platform/config/`, `platform/authn/`, `main.go`):

| Env Var | Consumer | Effect |
|---|---|---|
| `APP_PORT` | `main.go:605` | HTTP listen port (default 8080; script sets 8081) |
| `METALDOCS_DATABASE_URL` | `platform/bootstrap` | Postgres connection string |
| `METALDOCS_FANOUT_URL` | `main.go:362` | docx-renderer endpoint; if empty, approval runtime disabled |
| `METALDOCS_DOCX_RENDERER_SERVICE_TOKEN` | `main.go:366` | Service-to-service auth token for docx-renderer |
| `METALDOCS_E2E` | `main.go:136` | Gates E2E destructive handlers; must be `"1"` |
| `METALDOCS_SKIP_STARTUP_MIGRATIONS` | `main.go:186` | Bypasses migration runner at startup |
| `METALDOCS_MIGRATIONS_DIR` | `main.go:187` | Override for migration directory (default `db/migrations`) |
| `AUDIT_RETENTION_DAYS` | `main.go:575` | Days to retain audit events; 0 = disabled |
| `ENABLE_JOB_*` | `main.go:528-559` | Feature flags for each scheduler job |
| Config loaded via `config.Load*Config()` | `platform/config/` | CORS, rate-limit, attachments, feature flags, jobs, worker settings |

Worker-specific env:
| Env Var | Consumer |
|---|---|
| `METALDOCS_FANOUT_URL` | `worker/main.go:73` |
| `METALDOCS_WORKER_*` | `platform/config.LoadWorkerConfig()` |

---

## 8. Concurrency & async

Goroutines started by the API composition root (`main.go`):

| Goroutine | Location | Lifecycle |
|---|---|---|
| PDF outbox worker | `main.go:488-489` | Runs until ctx cancels; supervised by `startOutboxWorker` with exponential backoff restart |
| Materialize outbox worker | `main.go:491-492` | Same as above |
| Scheduler | `main.go:563-566` | `s.Start(ctx)` — tick-based job runner; exits on ctx cancel |
| Session sweeper | `main.go:568` | 60-second interval, sweeps expired editor sessions |
| Orphan pending sweeper | `main.go:569` | Hourly, sweeps orphaned pending uploads older than 24h |
| Presence hub Run | `main.go:306` | Presence connection hub; exits on ctx cancel |
| Presence hub RunHeartbeat | `main.go:307` | 30-second heartbeat tick; exits on ctx cancel |
| Audit retention purge | `main.go:578-593` | 24-hour ticker; only if `AUDIT_RETENTION_DAYS > 0` |
| HTTP server | `main.go:623-625` | Blocked goroutine running `server.ListenAndServe()` |

The `workerWG` and `schedulerWG` sync.WaitGroups are used during graceful shutdown (`main.go:627`) to drain outbox workers and the scheduler before process exit.

---

## 9. Error handling & observability

**Error handling patterns in composition root:**

- Config load failures call `log.Fatalf` — process exits immediately, no partial startup (`main.go:160-178`)
- `bootstrap.BuildAPIDependencies` failure is fatal (`main.go:182`)
- Migration failure is fatal (`main.go:194`)
- Individual module initialization failures are fatal with descriptive messages
- Outbox worker goroutines use exponential backoff restart (base 1s, cap 60s) logged via `slog.Error` (`main.go:470-485`)

**Observability wiring:**
- `platform/observability.NewHTTPObservability` wraps the mux with metrics (`main.go:275`)
- `slog.Default()` is used for structured logging throughout the composition root; the startup script tees this to `logs/api.log`
- `platform/requesttrace` propagates request IDs through context
- `/api/v1/metrics` Prometheus-compatible endpoint mounted at `main.go:572`
- `/healthz` and `/readyz` registered via `observability.NewHealthHandler` (`main.go:222`)

**RFC 9457 usage:** All module handlers use `platform/problem` to return `application/problem+json`; the error vocabulary is closed (completed 2026-06, commits `ef696a177`, `2369a02bf`).

---

## 10. Legacy / duplication / smell flags

- **DEAD BINARY: `cmd/seed-test-document/`** — `cmd/seed-test-document/main.go` has hardcoded DSN (`host=127.0.0.1 port=5433 ... password='<redacted — see .env>'`), hardcoded MinIO credentials, and no reference in any CI workflow, canonical script, or README. Last git touch: commit `c4a7d9a93` (2026-04, extracted `DevTenantID` constant). Binary `seed-test-document.exe` appears in `.gitignore`. No active consumer found. RF candidate for deletion.

- **LEGACY STARTUP SCRIPT: `scripts/start-spec1-api.ps1`** — Contains hardcoded absolute path to a different machine's username (`C:\Users\leandro.theodoro.MN-NTB-LEANDROT\Documents\MetalDocs\...`). This script cannot function on the current machine. Last git touch: commit `9c62bd3a2` (early 2026). Dead in its current form; confusing to future developers.

- **LEGACY STARTUP SCRIPT: `scripts/start-api-planc.ps1`** — A Plan-C variant startup targeting port 8083 for a specific worktree. Last git touch: `403ad2eef` (2026-04). Not referenced by CI. Reasonable to retain for context or delete.

- **EMPTY PLATFORM PACKAGE: `internal/platform/cache/`** — Contains only `.gitkeep` (added initial commit `912879cba`). No Go source. Registered in `backend-blueprint.md` as a known gap (C4 caching). Drift bait: its presence implies caching infrastructure that does not exist. Either implement or delete the directory.

- **ORPHAN CONTRACT SURFACE: `api/openapi/spec2.yaml` + `internal/api/v2/`** — `spec2.yaml` last touched commit `e1944bc4a` ("purge v2 routes and apply plan10 schema cleanup"). `internal/api/v2/types_gen.go` is handwritten (not generated) and is consumed only by contract tests in three module delivery-layer test files. The `invariants.yml` comment explicitly notes the former `openapi-drift` job was removed because `spec2.yaml` "is referenced only by this workflow" and "the path no longer exists". The spec2 surface has no active route coverage. `backend-blueprint.md` flags it as needing convergence or explicit fencing.

- **STALE CAPABILITY CATALOG PIN: `ops/CAPABILITY_CATALOG.sha256`** — Contains the string `placeholder-hash-update-after-catalog-created` instead of an actual SHA-256 hash. The `invariants.yml` CI gate reads this file and auto-creates it if missing but does not fail on placeholder content — the check compares `ACTUAL=$(sha256sum $CATALOG)` against this literal, so it will always fail if `sql/seeds/capabilities_v2.sql` exists. The referenced file `sql/seeds/capabilities_v2.sql` does not exist in the working tree (`sql/` contains only `sql/contracts/state_transitions.yaml`). The gate silently passes with "Catalog file not found" and exits 0. **The capability catalog integrity check is currently a no-op.**

- **DUPLICATE LOG FILE SPRAWL: root-level `*.log` files** — Dozens of `api-bg.err.log`, `api-main-8081.combined.log`, `api-smoke.err.log`, etc. accumulate in the repository root. Most patterns are gitignored (`tmp-*.log`, `api_err.log`, etc.) but several named variants (e.g. `api-bg.err.log`, `api-main-8081.combined.log`) are not in `.gitignore` and appear untracked (`??` in git status). These clutter the working tree and may confuse `git status` for new contributors. Not in-repo; dev ops hygiene issue.

- **STALE BIN ARTIFACT: `bin/metaldocs-api.exe`** — A compiled binary committed to git (last touched initial commit `912879cba`). The `.gitignore` excludes `metaldocs-api.exe` at root but not `bin/metaldocs-api.exe`. This binary is stale and was committed before the gitignore was tightened.

- **ARCHIVE MIGRATIONS IN TREE: `archive/migrations/`** — 142 SQL files representing the pre-baseline migration tree, preserved as historical record. Last git touch: `266bd4132` ("archive the legacy pre-baseline migration tree", 2026-05). Not applied by `platform/migrate`; correct per intent, but the directory is large (142 files) and adjacent to `db/migrations/`, creating potential confusion about which migrations are active.

- **DEPLOYMENT DOCKERFILE MISSING JOBS BINARY** — `deploy/docker/` has `api.Dockerfile` and `worker.Dockerfile` but no `jobs.Dockerfile`. The `metaldocs-jobs` binary has no container image definition. If jobs is intended to run in production as a separate container, its Dockerfile is absent. `deploy/compose/docker-compose.yml` also has no `jobs` service. [runtime-unverified: whether jobs is co-started within the API container or via an untracked mechanism]

- **GOVERNANCE: `scripts/check-governance.ps1` requires tests/ change for any internal/modules/ change** — The script enforces: if `internal/modules/` changed then `tests/` must also change. This creates a coarse-grained gate that may reject refactors that genuinely do not need new tests (e.g. renaming an unexported function). Not a smell in isolation but worth noting as a potential friction point.

- **NON-CANONICAL SCRIPT: `scripts/dev-api.ps1`, `dev-api-web.ps1`, etc.** — Multiple alternate dev startup scripts exist alongside the canonical `start-api.ps1`. CLAUDE.md §1 explicitly states "Scripts must rebuild or explicitly prove freshness" and "Ad hoc startup commands are not authoritative." These alternate scripts do not have the freshness checks of `start-api.ps1`. They are likely legacy convenience scripts from before the canonical script was hardened.

---

## 11. Wiki drift

No existing wiki doc covers repository topology specifically. `wiki/architecture/backend-blueprint.md` references `apps/api/cmd/metaldocs-api/main.go:595-602` for the middleware chain, which matches what is found at `main.go:598-602` (slight offset — the chain is assembled starting at line 598 in the current file). Minor drift; no material mismatch.

`backend-blueprint.md` section C4 states "`platform/cache` directory exists but is empty — either implement or delete". Confirmed: only `.gitkeep` present. Wiki is accurate.

`backend-blueprint.md` section A3 states "`api/openapi/spec2.yaml` + `internal/api/v2/` exist as a parallel surface — must converge or be explicitly fenced." Confirmed present and unfenced. Wiki is accurate.

---

## 12. Open questions

- **[runtime-unverified]** How is `metaldocs-jobs` deployed in production? No `jobs.Dockerfile` exists in `deploy/docker/` and no `jobs` service in `deploy/compose/docker-compose.yml`. The River jobs client may be co-started inside the API process (the `apps/api/cmd/metaldocs-api/main.go` does wire a River client bundle for scheduled publish enqueueing, but the full River worker host is in `metaldocs-jobs`). Whether `metaldocs-jobs` runs as a sidecar, in the same container, or is simply not in production yet is unknown without a live environment.

- **[runtime-unverified]** The `sql/seeds/capabilities_v2.sql` file referenced by `ops/CAPABILITY_CATALOG.sha256` and `invariants.yml` does not exist in the working tree. It may have been deleted, never created, or generated at runtime. The CI gate silently passes with "Catalog file not found". Confirm whether this file should exist and if so what its source is.

- **[runtime-unverified]** `bin/metaldocs-api.exe` is a committed binary. Without running it, its exact provenance (which commit's source it was compiled from) is unknown. Recommend `git rm bin/metaldocs-api.exe` and adding `bin/*.exe` to `.gitignore`.

- **[runtime-unverified]** The `deploy/docker/api.Dockerfile` exposes port 8081 but the Go binary defaults to `:8080` unless `APP_PORT` is set. In the container, `APP_PORT` must be set to `8081` to match the exposed port. Whether the Docker entrypoint or compose file sets this env var is not confirmed without running the compose stack.

- **Naming drift:** The root module is `module metaldocs` but the binary is `metaldocs-api`, worker is `metaldocs-worker`, jobs is `metaldocs-jobs`, and the `cmd/seed-test-document/` binary uses `package main` with no module-level alignment. The top-level `cmd/` directory coexists with `apps/*/cmd/` — two `cmd/` conventions in one repo.
