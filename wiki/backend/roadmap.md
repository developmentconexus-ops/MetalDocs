# Backend Professionalization — Execution Roadmap (LIVING TRACKER)

> **Status:** ACTIVE · Wave 0 executed — awaiting user review gate before Wave 1
> **Last updated:** 2026-06-11 (Wave 0 session; post-rewrite hash stamp)
> **NOTE:** the 0.4 history rewrite (2026-06-11) changed every commit hash in the repo. Hashes below are the REWRITTEN ones. Any other clone of this repo must be re-cloned.
> **Contract:** [`docs/superpowers/specs/2026-06-11-backend-professionalization-design.md`](../../docs/superpowers/specs/2026-06-11-backend-professionalization-design.md) (decisions D-1..D-5, protocol)
> **Verdicts:** [`stage2-evaluation.md`](stage2-evaluation.md) · **Evidence:** [`legacy-register.md`](legacy-register.md)
>
> **Update protocol (mandatory):** every wave session updates this file IN THE SAME COMMIT as its close-out — flip statuses, fill the Commit and Evidence columns, bump "Last updated". A wave is not closed while its rows are stale. The final review (Wave F) cannot start until every Wave 0–2 row is ✅ or explicitly deferred with a trigger.

**Status legend:** ☐ pending · ▶ in progress · ✅ done · ⏸ gated (waiting on dependency/user) · 🔁 done, runtime re-check pending (Docker) · ➖ deferred (trigger recorded)

---

## Program map

| Phase | Status | Output |
|---|---|---|
| Stage 1 — map the backend | ✅ `d7489d5a2` | `wiki/backend/` atlas, legacy register (212 flags) |
| Stage 2 — evaluate vs standards | ✅ `e5a162867` | 34 verdicts, 4-wave plan, 2-ADR backlog |
| Brainstorm — decisions D-1..D-5 | ✅ `432e24405` | Design spec (the contract) |
| **Wave 0 — P0 prerequisites** | ✅ evidence in handoff · user review gate pending | security unblocked, deployment complete, layering legal |
| **Wave 1 — high-value / low-blast** | ☐ | correctness + compliance defects < ~100 lines each |
| **Wave 2 — structural refactors** | ☐ | atomicity, RLS, boundaries, ADR 0027 |
| Wave 3 — trigger-gated | ➖ by design | executed only when triggers fire |
| **Wave F — final full review** | ☐ | program sign-off: backend verified solid |

---

## Wave 0 — P0 prerequisites (fresh session #1)

| # | Item | Findings | Status | Commit | Evidence / notes |
|---|------|----------|--------|--------|------------------|
| 0.1 | Delete seed binary; redact secret from 7 docs; rm committed .exe; .gitignore fix; delete dead script; pin api-lint.exe | F-18, D-4a | ✅ | `58cbf9943` | Secret grep clean over tracked files AND full working tree (incl. `.env` — already holds a new value). Redacted 9 tracked files (2 more than planned: `scripts/start-api-no-build.ps1`, `scripts/start-worker.ps1` comments + `_artifacts/stage1/repo-topology.md`); `bin/metaldocs-api.exe` + `scripts/api-lint/api-lint.exe` were on-disk only, never git-tracked → file deleted / rebuilt from source (SHA-256 `5660…2C8D`); `bin/*.exe` added to `.gitignore`; stale gitignored agent-worktree copies redacted in place |
| 0.2 | gitleaks secret-scan in CI + D-4a rule in documentation-governance.md | F-18 | ✅ | `3402a8bbd` + `e2b5b2aa4` | `.github/workflows/secret-scan.yml` (gitleaks v8.24.3, working-tree mode until 0.4 rewrites history) + `.gitleaks.toml` allowlist (10 findings assessed: test fixtures, `.gen.go` spec chunks, stale plan-doc token ≠ current `.env`, runbook sample password). Verified: gitleaks over `git archive HEAD` + config = no leaks (exit 0); YAML parses |
| 0.3 | **USER:** rotate dev DB password in `.env` at Docker re-creation (+ anywhere reused) | F-18 | ➖ owner-waived | — | User declined rotation in the Wave 0 session ("keep this work professional, no secret") and explicitly accepted the residual risk of the historical value; `.env` working value verified ≠ the leaked string. 0.4 unlocked by this directive |
| 0.4 | History rewrite: `git filter-repo --replace-text` + force-push; user re-clones | F-18 | ✅ | (rewrite itself) | Executed 2026-06-11 with explicit user authorization: `git filter-repo --replace-text` (old secret → `***REDACTED***`) on a fresh mirror clone, 3057 commits rewritten; force-pushed all `refs/heads/*` + `refs/tags/*` to origin. **VERIFICATION (corrected in round-2 review):** a fresh *mirror* clone of origin (which fetches PR refs, unlike a plain clone) confirms all 45 branches + the tag are clean — every surviving secret commit is reachable ONLY from `refs/pull/*` (73 immutable GitHub PR refs, `heads=0 tags=0` for all). The earlier "plain-clone `git log --all -S` = empty" check was insufficient because a plain clone does not fetch PR refs. Also confirmed: leaked value ≠ current `.env` password (stale credential). Residual: the stale value persists in the 73 PR refs. **RESOLVED-BY-PLAN (owner decision, 2026-06-11):** F-18 history residual is permanently closed at first release — the owner will re-baseline into a NEW repo whose first commit is the v1 release, abandoning all prior history (PR refs included) with the old repo. No GitHub-support purge needed. Until then it is a dead dev credential in a single-user repo (leaked value ≠ current `.env`). Local repo reset onto rewritten origin (~22 stale LOCAL branches still hold the value in this clone's object DB — local-only, gone at re-baseline) |
| 0.5 | `jobs.Dockerfile` + `jobs` service in compose | F-19-deployment | ✅ | `baf6e1b78` | RUNTIME-VERIFIED (Docker up on this machine): `docker compose config` parses; `compose-jobs` image built; `docker compose up -d --no-deps jobs` → log "MetalDocs Jobs running (queues=temporal)", container stays Up. Service envs: `METALDOCS_JOBS_ENABLED`/`METALDOCS_JOBS_TEMPORAL_MAX_WORKERS` + PG* (mirrors worker); `depends_on: postgres(healthy), api(healthy)`. Round-2: added `METALDOCS_JOBS_RIVER_SCHEMA` passthrough to api+jobs so the schema cannot diverge |
| 0.6 | platform/observability → callback injection (drop auth/domain import) | F-06a | ✅ | `8e0aa9eb4` + `e2b5b2aa4` | `WithUserIDResolver(func(*http.Request) string)` setter (pattern: `platform/ratelimit`); wired at composition root `main.go`. Build + vet + `go test ./internal/platform/observability/...` OK; grep: observability clean. **Discovered for 1.10 scoping:** `platform/{authn,bootstrap,docgenv2,objectstore,security}` still import `internal/modules/**` — the 1.10 CI guard needs explicit per-package disposition (bootstrap/authn are composition-adjacent; `security/ratelimit.go` dies in 2.8). Round-2: `WithUserIDResolver` setter promoted to a **required `NewHTTPObservability` constructor param** so the resolver can't be silently omitted |

**Wave close:** evidence block in handoff · this file updated · user review → Wave 1.

## Wave 1 — high-value / low-blast (fresh session #2)

| # | Item | Findings | Status | Commit | Evidence / notes |
|---|------|----------|--------|--------|------------------|
| 1.1 | Middleware chain reorder + panic recovery + pre-auth login rate limit + REQ-MW-7 chain-order test | F-01 | ✅ | `58d7009d9` + `2d8728c6b` | Chain now `panicRecovery → httpObs → cors → origin → preAuthLoginLimit → authn → iam → presenceBump → rateLimiter → mux` via declarative `apiChain`/`buildChain` (`apps/api/cmd/metaldocs-api/chain.go`) + REQ-MW-7 order test (`chain_test.go`). New `platform/middleware.Recovery` (trace-ID + panic → 500 problem+json; re-panics ErrAbortHandler). httpObs: records panicked requests as 500 in RED metrics (defer; panic NOT swallowed); user attribution preserved across the authn boundary via `observability.SetPrincipal` slot (authn reports outward — REQ-MW-4 without losing audit attribution). Pre-auth login limit: `platform/ratelimit` instance, `auth_login` 10/min IP-keyed, login path only. build+vet+tests green (`apps/api/cmd`, `platform/{middleware,observability,ratelimit}`, `modules/auth`, `tests/unit`). **RUNTIME-VERIFIED** (start-api.ps1 -Build, :8081): login 200 · authed `GET /internal/test/panic` (METALDOCS_E2E probe, commit `2d8728c6b`) → 500 problem+json `INTERNAL_ERROR`, process alive (health 200 after) · metrics show 401s (`/api/v1/iam/users` errors=2) and panics (panic route errors=4) · panic log line trace-ID-tagged with full stack; stack shows live chain order = canon · `http_request` for the panicked authed request has `user_id":"admin"` (principal slot works through a panic) · 12 rapid bad logins → 11×401 then 429 (pre-auth IP limit live) |
| 1.2 | http.Server Read/Write/Idle timeouts | F-16 | ✅ | `1b3f0d11d` | `ReadTimeout 30s / WriteTimeout 60s / IdleTimeout 90s` added at the `http.Server` literal in `main.go` (ReadHeaderTimeout 5s kept). WebSocket presence unaffected (hijacked conns manage own deadlines). `go build`+`go vet` clean. F-16B (WS drain) stays deferred per D-1; F-16C (readiness concurrency) Wave 3 trigger |
| 1.3 | Delete `spec2.yaml` + `internal/api/v2`; migrate 3 contract tests; fix capability-catalog CI gate | F-03 | ✅ | (this commit) | Parallel surface gone: `api/openapi/spec2.yaml` (1061 lines) + `internal/api/v2/` deleted; 3 contract tests (CD/taxonomy/iam) decode `problem.Problem` instead of `apiv2.APIError`; `.golangci.yml` types_gen exclusion dropped. CI gate **deleted not fixed**: `capability-catalog-hash` pinned `sql/seeds/capabilities_v2.sql` which never existed (always exit 0) — real REQ-AUTHZ-5 guard is api-lint `registry_rules.go` (ADR 0022); placeholder `ops/CAPABILITY_CATALOG.sha256` removed; rationale comment left in `invariants.yml`. v1 spec untouched (no regen needed). build+vet+3 delivery test pkgs green; both YAMLs parse |
| 1.4 | Idempotency middleware codes → problem catalog; expand guard test | F-09 | ✅ | (this commit) | `writeErrJSON` signature now takes `problem.Code`; all 6 raw literals replaced: `IDEMPOTENCY_KEY_REQUIRED`→const, `IDEMPOTENCY_KEY_INVALID`→new const, `REQUEST_BODY_TOO_LARGE`→new const, `BAD_REQUEST`→`CodeValidationError`, `IDEMPOTENCY_KEY_CONFLICT`→`CodeIdempotencyKeyReused` (fixes pre-existing FE mismatch — FE already mapped REUSED), `INTERNAL`→`CodeInternalError`. Guard test now covers `internal/platform/idempotency` + `writeErrJSON`. FE catalog regenerated (`dump-error-codes.go`, 112 codes) + 2 PT-BR messages; FE coverage test green. Go build/vet/tests green. Finalize-handler inline idempotency (F-09 half 1) untouched — RF-10 scope, not this card |
| 1.5 | Remove dead search response fields + `businessUnit` param; fix bare 405s (search+security) | F-13a/b, D-03 | ✅ | (this commit) | `SearchDocumentResponse` now mirrors spec `SearchDocumentItem` exactly (subject/business_unit/classification/tags dropped — SQL reader never populated them; spec never declared them → code aligned TO spec, **no spec change/regen**). camelCase `businessUnit` param read removed (undocumented; reader no-ops it). Bare 405s → RFC 9457 `METHOD_NOT_ALLOWED` problem+json + `Allow: GET` in search (1×) and security (3×, shared helper); new catalog const + FE message + regen (113 codes). Tests: 405 contract test added; legacy-fields test inverted to assert absence; businessUnit filter assertion inverted. build/vet/tests + FE coverage green. Remaining no-op params subject/classification/tag (snake_case, documented) out of card scope |
| 1.6 | River schema migration single owner (remove from bootstrap/jobs.go) | F-19 | 🔁 | (this commit) | `MigrateRiverSchema` call removed from `BuildJobsDependencies` (jobs binary); API binary is sole owner (`main.go` startup). Ordering guaranteed by compose `depends_on: api (healthy)` (0.5) + shared `METALDOCS_JOBS_RIVER_SCHEMA` passthrough (0.5 round-2). Rationale comment left at the removal site. build/vet/bootstrap tests green; jobs container re-check at wave close |
| 1.7 | lease_reaper JOIN bug → system-scoped governance event | F-19 | ☐ | | |
| 1.8 | Templates `ListAudit` reads `audit_events` (close read/write sink split) | F-07-sub-split | ☐ | | |
| 1.9 | Remove vestigial `.gitkeep` scaffolds | F-08 | ☐ | | |
| 1.10 | **CI guard:** `internal/platform/**` must not import `internal/modules/**` | F-06a lock-in | ☐ | | locks 0.6 forever. Round-2 note: guard cannot be a blanket rule — `platform/{authn,bootstrap,docgenv2,objectstore,security}` still import modules; needs per-package disposition (bootstrap/authn composition-adjacent; `security/ratelimit.go` dies in 2.8) |
| 1.11 | gitleaks **full git-history** mode (drop `--no-git`, `fetch-depth: 0`) | F-18 round-2 | ☐ | | Closes the commit-then-amend bypass. BLOCKED ON a triage pass: git-mode over the rewritten branch surfaces ~15 historical test-fixture/dummy-token findings (all in already-path-allowlisted classes, but the singular `[allowlist]` schema needs per-finding review for git-mode). Also: v8.24.3 ignores the plural `[[allowlists]]` AND-form, so per-file regex scoping needs a gitleaks upgrade or rule-level allowlist |

## Wave 2 — structural refactors (fresh session #3)

| # | Item | Findings | Status | Commit | Evidence / notes |
|---|------|----------|--------|--------|------------------|
| 2.1 | **ADR 0027** — auth_identities tenant-global by design; RLS sequencing (write BEFORE 2.3) | D-3, F-12 | ☐ | | |
| 2.2 | In-tx audit/governance writes (`RecordTx`/`LogTx`) across taxonomy/templates/documents-core + **CI guard vs post-commit audit calls** | F-07, D-01, D-2 | ☐ | | approval module is the template |
| 2.3 | RLS on `controlled_documents` + `audit_events`; `ExportJob.TenantID` → uuid.UUID | F-12 | ⏸ on 2.1 | | rest of tables ➖ trigger: first external tenant |
| 2.4 | Capability literals: resolve `"template.admin"`; typed EventType consts; tier-1/2 pairing audit (ADR 0028 only if new cap minted) | F-11 | ☐ | | ask user if new capability needed |
| 2.5 | Extract delivery-layer raw SQL (CD routes.go, documents handler.go) to repositories | F-06b | ☐ | | |
| 2.6 | auth → iam_users write via `LoginContextPort` | F-06c | ☐ | | |
| 2.7 | iam/delivery drops auth/infrastructure import (promote types to auth/domain) | F-06d | ☐ | | |
| 2.8 | Delete `platform/security/ratelimit.go`; activate `platform/ratelimit` in prod wiring | F-05, D-04 | ⏸ on 1.1 | | −~200 lines |
| 2.9 | N+1 fixes: RolesByUserID join, batch variant, VerifyUserInTenant EXISTS, approval inbox batch-load, audit ILIKE restriction | F-10 | ☐ | | authz path first |
| 2.10 | Postgres-backed auth-failure counter (replace in-memory) | F-20e | ☐ | | |
| 2.11 | Dead-code deletion PRs (CutoverService, CompositionConfig, SetParent, fallbacks, coverage_boost_test) + remove deprecated govLogger fallback | F-14, F-07-sub | ⏸ logger on 2.2 | | |

## Wave 3 — trigger-gated (NOT executed proactively)

Tracked as written triggers; full table in `stage2-evaluation.md` §Wave 3. Headline triggers: OTel/W3C (F-17) → "second host or first external-tenant SLA" · CD trigram index (F-20b) → ">50k rows or p95 >200ms" · readiness concurrency (F-16C) → "second dependency added" · RLS remaining tables (F-09d/F-12 tail) → "first external tenant" · others → "next touch of that area".

## Wave F — FINAL FULL REVIEW (fresh session, after user approves Wave 2 close)

The user-mandated gate: prove the backend is solid before declaring the program done.

| # | Check | Status | Evidence |
|---|-------|--------|----------|
| F.1 | All 4 CI guards active and green: gitleaks · platform-import boundary · post-commit-audit ban · middleware chain-order test | ☐ | |
| F.2 | Full static pass: `go build ./...` · `go vet ./...` · full test suite · `api-lint` — zero blocking findings | ☐ | |
| F.3 | Runtime QA (Docker up): `.\scripts\start-api.ps1` · login · backend-api-qa-checklist smoke · worker+jobs containers process real work · scheduled publish fires · every 🔁 `[runtime-unverified]` item from Waves 0–2 re-checked live | ☐ | |
| F.4 | Legacy-register sweep: every F-*/D-* entry marked **resolved (commit)** or **deferred (trigger)** — no silent leftovers | ☐ | |
| F.5 | `backend-blueprint.md` maturity grades re-scored; REQ-* compliance checked against `backend-target-architecture.md`; deltas explained | ☐ | |
| F.6 | Full-program code review (`/code-review` over the cumulative program diff); findings dispositioned by family | ☐ | |
| F.7 | Wiki coherence pass (wiki-curator): anchors, stamps, indexes match post-refactor reality | ☐ | |
| F.8 | Closing report in handoff: every wave's evidence, all defers with triggers, ADR links, program sign-off | ☐ | |

**Program DONE = all F-rows ✅.** After that, evolution goes through the normal REQ/ADR process — this roadmap freezes as the historical record.
