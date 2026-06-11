# Backend Professionalization — Execution Roadmap (LIVING TRACKER)

> **Status:** ACTIVE · Wave 0 not started
> **Last updated:** 2026-06-11 (created)
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
| **Wave 0 — P0 prerequisites** | ☐ | security unblocked, deployment complete, layering legal |
| **Wave 1 — high-value / low-blast** | ☐ | correctness + compliance defects < ~100 lines each |
| **Wave 2 — structural refactors** | ☐ | atomicity, RLS, boundaries, ADR 0027 |
| Wave 3 — trigger-gated | ➖ by design | executed only when triggers fire |
| **Wave F — final full review** | ☐ | program sign-off: backend verified solid |

---

## Wave 0 — P0 prerequisites (fresh session #1)

| # | Item | Findings | Status | Commit | Evidence / notes |
|---|------|----------|--------|--------|------------------|
| 0.1 | Delete seed binary; redact secret from 7 docs; rm committed .exe; .gitignore fix; delete dead script; pin api-lint.exe | F-18, D-4a | ✅ | (this commit) | Secret grep clean over tracked files AND full working tree (incl. `.env` — already holds a new value). Redacted 9 tracked files (2 more than planned: `scripts/start-api-no-build.ps1`, `scripts/start-worker.ps1` comments + `_artifacts/stage1/repo-topology.md`); `bin/metaldocs-api.exe` + `scripts/api-lint/api-lint.exe` were on-disk only, never git-tracked → file deleted / rebuilt from source (SHA-256 `5660…2C8D`); `bin/*.exe` added to `.gitignore`; stale gitignored agent-worktree copies redacted in place |
| 0.2 | gitleaks secret-scan in CI + D-4a rule in documentation-governance.md | F-18 | ✅ | (this commit) | `.github/workflows/secret-scan.yml` (gitleaks v8.24.3, working-tree mode until 0.4 rewrites history) + `.gitleaks.toml` allowlist (10 findings assessed: test fixtures, `.gen.go` spec chunks, stale plan-doc token ≠ current `.env`, runbook sample password). Verified: gitleaks over `git archive HEAD` + config = no leaks (exit 0); YAML parses |
| 0.3 | **USER:** rotate dev DB password in `.env` at Docker re-creation (+ anywhere reused) | F-18 | ⏸ user | — | unlocks 0.4 |
| 0.4 | History rewrite: `git filter-repo --replace-text` + force-push; user re-clones | F-18 | ⏸ on 0.3 | | `git log --all -S '<secret>'` empty |
| 0.5 | `jobs.Dockerfile` + `jobs` service in compose | F-19-deployment | ✅ | (this commit) | RUNTIME-VERIFIED (Docker up on this machine): `docker compose config` parses; `compose-jobs` image built; `docker compose up -d --no-deps jobs` → log "MetalDocs Jobs running (queues=temporal)", container stays Up. Service envs: `METALDOCS_JOBS_ENABLED`/`METALDOCS_JOBS_TEMPORAL_MAX_WORKERS` + PG* (mirrors worker); `depends_on: postgres(healthy), api(healthy)` |
| 0.6 | platform/observability → callback injection (drop auth/domain import) | F-06a | ✅ | (this commit) | `WithUserIDResolver(func(*http.Request) string)` setter (pattern: `platform/ratelimit`); wired at composition root `main.go`. Build + vet + `go test ./internal/platform/observability/...` OK; grep: observability clean. **Discovered for 1.10 scoping:** `platform/{authn,bootstrap,docgenv2,objectstore,security}` still import `internal/modules/**` — the 1.10 CI guard needs explicit per-package disposition (bootstrap/authn are composition-adjacent; `security/ratelimit.go` dies in 2.8) |

**Wave close:** evidence block in handoff · this file updated · user review → Wave 1.

## Wave 1 — high-value / low-blast (fresh session #2)

| # | Item | Findings | Status | Commit | Evidence / notes |
|---|------|----------|--------|--------|------------------|
| 1.1 | Middleware chain reorder + panic recovery + pre-auth login rate limit + REQ-MW-7 chain-order test | F-01 | ⏸ on 0.6 | | |
| 1.2 | http.Server Read/Write/Idle timeouts | F-16 | ☐ | | |
| 1.3 | Delete `spec2.yaml` + `internal/api/v2`; migrate 3 contract tests; fix capability-catalog CI gate | F-03 | ☐ | | −~1100 lines |
| 1.4 | Idempotency middleware codes → problem catalog; expand guard test | F-09 | ☐ | | |
| 1.5 | Remove dead search response fields + `businessUnit` param; fix bare 405s (search+security) | F-13a/b, D-03 | ☐ | | |
| 1.6 | River schema migration single owner (remove from bootstrap/jobs.go) | F-19 | ⏸ on 0.5 | | |
| 1.7 | lease_reaper JOIN bug → system-scoped governance event | F-19 | ☐ | | |
| 1.8 | Templates `ListAudit` reads `audit_events` (close read/write sink split) | F-07-sub-split | ☐ | | |
| 1.9 | Remove vestigial `.gitkeep` scaffolds | F-08 | ☐ | | |
| 1.10 | **CI guard:** `internal/platform/**` must not import `internal/modules/**` | F-06a lock-in | ☐ | | locks 0.6 forever |

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
