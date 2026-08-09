# PASS 1 — Audit Baseline and Repository/Runtime Topology

**Date:** 2026-08-09
**Status:** reproduced-current (fresh local measurement in an isolated worktree)

## 1. Measurement baseline

| Fact | Value |
|---|---|
| `main` SHA (== `origin/main` after `git fetch --prune`) | `418070bf38a9f358f9131bcc36b7a6bcbc069273` |
| Audit branch | `docs/architecture-audit-current-state` @ `9e48a6a1f6e08182b5af91d7251da727dec2ca36` |
| Branch vs `main` | 7 commits ahead, **0 behind**; merge-base == `main` HEAD → all product code on the branch is byte-identical to `main`; the branch adds docs only |
| Go toolchain | `go version go1.26.5 windows/amd64` |
| Dirty status | clean (0 modified files at measurement start) |
| Isolation | dedicated `git worktree` at `.claude/worktrees/arch-audit`; main checkout untouched |

Evidence-class labels used across all PASS artifacts: `reproduced-current`
(measured at this baseline), `historical` (true at filing time, superseded),
`stale` (contradicted by current runtime truth).

## 2. Binaries / executable roots

| Binary | Root | Role |
|---|---|---|
| `metaldocs-api` | `apps/api/cmd/metaldocs-api` | sync HTTP data/control plane + authz; joins River leader election to enqueue periodic maintenance jobs (ADR 0067 dual-define) |
| `metaldocs-worker` | `apps/worker/cmd/metaldocs-worker` | async transactional-outbox consumers |
| `metaldocs-jobs` | `apps/jobs/cmd/metaldocs-jobs` | hosts + executes River periodic jobs, scheduled publish, notifications fanout |
| `metaldocs-e2e-seed` | `apps/api/cmd/metaldocs-e2e-seed` | test/E2E seeding utility (not a production plane) |
| `docx-renderer` | `apps/docx-renderer` (Node.js/TypeScript) | internal-only DOCX rendering service |

## 3. Backend source topology

- Single Go module `metaldocs`; 158 first-party packages (see PASS 2).
- `internal/modules/` — **15 module directories** (mechanically confirmed):
  approval, audit, auth, controlleddocuments, distribution, documents, iam,
  jobs, notifications, render, search, security, taxonomy, templates, tokens.
  Any document stating 11/12 modules or a nested `documents/approval` is stale.
- `internal/platform/` — **37 packages**: apibase, authn, bootstrap, config,
  crypto, db, docgenv2, featureflags, formval, httpclient, httpresponse,
  httprouter, iamtypes, idempotency, jobs, legacystatus, messaging, middleware,
  migrate, objectstore, observability, pagination, passwordhash, problem,
  ratelimit, render, requesttrace, security, servicebus, sqlescape, storage,
  strictjson, tenant, tenantdata, tripwire, useragent, worker.
- `internal/composition/` — exactly one root: `tenantdata` (registry wiring
  12 of 15 module infrastructures).
- `tools/` — `cilint`, `perfbench`, `verify` (Go verifier/registry).
- `scripts/` — 58 PowerShell/utility scripts.

## 4. Contract, data, frontend, deploy surfaces

| Surface | Location | Note |
|---|---|---|
| OpenAPI SSOT | `api/openapi/v1/` + `api/openapi/internal-e2e.yaml` | routes change only via spec + oapi-codegen |
| DB | `db/{baseline,migrations,grants,prerequisites,reference-data,dev-seeds}` | 4-stage bootstrap; migrations folded 2026-07-29 (next = 0316+) |
| Frontend | `frontend/apps/web/src/{features,components,lib,app,routing,store,styles}` | React + TanStack Query |
| Deploy | `deploy/compose/docker-compose.yml`, `deploy/docker/`, `deploy/nginx/` | single compose, self-labelled dev/test |
| CI | `.github/workflows/`: `ci.yml`, `docx-renderer.yml`, `nightly.yml`, `release.yml`, `smoke.yml` | 5 workflows (former 20-workflow layout collapsed — #87 staleness correction) |
| Infra services | Postgres, Redis, MinIO, River (in-DB queue) | |

## 5. Runtime shape (as wired today)

- API: middleware chain fixed at `apps/api/cmd/metaldocs-api/chain.go` —
  `panic_recovery → otel → http_obs → cors → origin_protection →
  pre_auth_login_rate_limit → authn → iam_authz → presence_bump → rate_limit →
  method_not_allowed`; RFC 9457 problem+json errors.
- Async: transactional outbox (worker) + River periodic jobs (jobs binary);
  API enqueues, jobs executes (`maintenance` queue), per ADR 0067/0068.
- Multi-tenant pooled Postgres, tx-local GUCs, tenant-namespaced MinIO keys.

Per-plane observability/health findings live in PASS 12; guard reachability in
PASS 13.
