# Current Agent Handoff

> **Last verified:** 2026-06-10
> **Scope:** Recovery context for a fresh agent session after the user reinstalls Claude/Codex. Captures workspace state, in-flight programs, the new architecture reference stack, and the agent-memory contents that do not survive a tool reinstall.
> **Out of scope:** Implementation details — follow the linked ADRs, backlog programs, and architecture docs.

## Why this exists (2026-06-10 rewrite)

The user is uninstalling Claude Code / Codex. The agent's persistent memory directory (`~/.claude/projects/.../memory/`) will be wiped, and session history will be gone. This file is the durable landing document. The previous version of this file (Plan 12 screen-finalization era, 2026-05-27) is superseded; Plan 12 context lives in git history if ever needed.

Start a fresh session by reading: `CLAUDE.md` → `wiki/README.md` → this file → the three-doc architecture stack (below).

## Workspace state at handoff

- Repository: `C:\Users\leandro.theodoro\Documents\MetalDocs`
- Branch: `qa/iam-area-membership` (main branch for PRs: `main`)
- Latest commit: `791cc0850 docs(wiki): sync route templates + query params to snake_case (Family 5)`
- **Uncommitted work in the tree (this session's deliverable — commit it):**
  - `wiki/standards/backend-canon.md` (new)
  - `wiki/architecture/backend-blueprint.md` (new)
  - `wiki/architecture/backend-target-architecture.md` (new)
  - `wiki/architecture/index.md`, `wiki/standards/index.md` (index entries)
  - `wiki/references/current-agent-handoff.md` (this rewrite)
  - Suggested commit: `docs(wiki): backend architecture reference stack (canon → blueprint → target) + handoff`

## The architecture reference stack (created 2026-06-10)

The user's directive: before further "industry-grade" refactoring, define everything — what a backend is composed of, how ours maps, and how it must behave. Three documents now form the canonical chain. **This is the reference for the whole refactoring effort:**

1. **[`wiki/standards/backend-canon.md`](../standards/backend-canon.md)** — implementation-independent definition of what a backend is composed of. Planes (data/control/management), layers, full concern catalog (edge, API, middleware, identity, domain/DDD, data, async, reliability, observability, security, config, integrations), exact identity vocabulary (authn vs authz, permission vs capability vs role vs scope, PEP/PDP, RBAC/ABAC/ReBAC), canonical sources, 10 litmus tests.
2. **[`wiki/architecture/backend-blueprint.md`](../architecture/backend-blueprint.md)** — current MetalDocs mapped onto the canon. Composition diagram, middleware chain (`chain.go` — reordered Wave 1), async write path, per-concern maturity grades (✅/🟡/🔴), standards register, scoreboard with owning programs.
3. **[`wiki/architecture/backend-target-architecture.md`](../architecture/backend-target-architecture.md)** — **normative target spec.** ~60 REQ-* requirements (RFC 2119), target diagrams/workflows (topology, middleware chain, login, two-tier authz sequence, contract-first workflow, outbox/retry/DLQ), **Refactoring register RF-1..RF-10** with sequencing, litmus tests bound to REQ IDs as release gates, governance rules (PRs cite REQ IDs; MUST deviations need an ADR).

Flow: **canon defines target → blueprint locates gap → target specifies behavior → ADR records decision → program executes.**

## In-flight programs (status at handoff)

### 1. ADR 0022 — Authz capability coherence
`wiki/decisions/0022-authz-capability-coherence.md`. Phases 1-5 and 7-13 COMPLETE (registry SSOT, typed scope, membership area-scoping, directory SQL-scoping, CI binding guards, runtime scope binding, raw-string dialect closed, CI coherence-net revived). **Phase 6 (wiki sync via `wiki-curator`) still pending — deliberately sequenced LAST.** Residual: `authz-call-present` call-graph lint rewrite (deferred), authz-cache invalidation contract (RF-3). The original `area_admin` membership 403 is closed at root. **Never symptom-patch authz — this ADR is the boundary.**

### 2. API contract hardening
`wiki/backlog/api-contract-hardening.md`. 6 phases from the 2026-06-05 audit: A base-path ✅, B authz security gaps ✅, C dead-path prune ✅ (2026-06-07), D envelope / E hygiene / F cleanup — F partially done (FD-1 landed 2026-06-08). ~397 reported-only (non-blocking) api-lint spec-drift hits remain; blocking guards are at 0. Maps to RF-4 (fence/converge `spec2.yaml` + `internal/api/v2`) and RF-5.

### 3. Backend standardization (final polish)
6 families: auth hardening / pagination / error-vocab / casing / CI / dead-code. Executed as batched workflows on `qa/std-execution`, **merged into `qa/iam-area-membership`** (last session). Recent commits on this branch are its output: RFC 9457 error vocabulary (`ef696a177`), constant-time login paths (`f792d64de`), snake_case params spec-first (`16b72f81f`, `2369a02bf`), api-lint CI cleanup (`8e3b99727`), wiki sync (`791cc0850`). Batch A review findings addressed (`7faa7fb3d`). Status: Batches A+B landed; verify whether any family has open tail work before claiming the program closed.

## Next steps (agreed sequencing, from the target doc §10)

1. Commit the doc stack (above).
2. **Close ADR 0022 Phase 6** (wiki sync — runs last, now unblocked).
3. Finish api-contract-hardening tail (RF-5) + decide RF-4 (fence or converge the v2 spec surface).
4. **RF-1 (observability depth audit — standalone)** — RF-2 was CLOSED in Wave 1 (F-01): chain reordered, recovery/obs outermost, pre-auth login limit added, order test in `chain_test.go`. RF-1 remains: OTel exporter wiring, trace propagation api→worker→docx-renderer through outbox rows, readiness-probe depth. New chain (verified): `panicRecovery → httpObs → cors → origin → preAuthLoginLimit → authn → iam → presence → ratelimit → mux` (`apps/api/cmd/metaldocs-api/chain.go`).
5. RF-3 (authz cache contract), RF-9 (graceful shutdown/timeout audit), RF-10 (idempotency coverage audit).
6. RF-7 (delete or implement empty `platform/cache`, `platform/storage`; document-or-fence `platform/messaging` servicebus), RF-8 (feature-flag lifecycle doc).

### Stage-1 backend audit — COMPLETE (2026-06-11)

Stage 1 (backend audit and mapping) is complete. The full atlas is at:

- **[`wiki/backend/index.md`](../backend/index.md)** — atlas root: binary map, repo topology, HTTP kernel, domain modules, platform packages, contract surface, flows, and navigation to all detail pages
- **[`wiki/backend/legacy-register.md`](../backend/legacy-register.md)** — complete legacy/duplication/smell register with RF-* cross-references and prioritized remediation sequencing

### Stage-2 backend evaluation — COMPLETE (2026-06-11)

Stage 2 (evaluation against professional SaaS and industry standards) is complete. The canonical evaluation artifact is at:

- **[`wiki/backend/stage2-evaluation.md`](../backend/stage2-evaluation.md)** — master verdict table (all findings), executive posture, P0 prerequisites, 4-wave prioritized roadmap, and ADR backlog

**Summary:** 3 P0 prerequisites (F-18 credentials, F-19 jobs deployment, F-06a platform layering inversion), 4 roadmap waves, 2 ADRs proposed (ADR 0027: RLS adoption sequencing; ADR 0028: upsertApprovalConfig capability, conditional).

**Next step: execute the roadmap starting with Wave 0 — F-18 credential prerequisite** (rotate Postgres password, delete `cmd/seed-test-document/`, rewrite git history, fix `.gitignore`, replace `api-lint.exe`). No other work is meaningful for the security program until this lands.

### Design spec — APPROVED (2026-06-11): execution contract for Waves 0-3

Brainstorm with the user is complete; all judgment calls are decided and recorded in:

- **[`docs/superpowers/specs/2026-06-11-backend-professionalization-design.md`](../../docs/superpowers/specs/2026-06-11-backend-professionalization-design.md)** — binding decisions (D-1..D-5), execution protocol, and per-wave cards.

Key decisions: single-server Docker Compose deployment (D-1); audit-facing compliance, no ceremony (D-2); multi-tenant design kept, RLS as defense-in-depth on 2 tables (D-3); FULL F-18 remedy incl. history rewrite gated on user rotation (D-4); secrets never quoted in docs (D-4a); **wave-by-wave execution, each wave in a FRESH session, review gate between waves (D-5)**.

**A fresh wave session reads:** `CLAUDE.md` → `wiki/README.md` → this file → the design spec → **the living roadmap tracker** [`wiki/backend/roadmap.md`](../backend/roadmap.md) → its wave card. Wave closes only with the evidence block recorded here AND the tracker rows updated in the same commit. The program ends with the Wave F final full review (all CI guards green, runtime QA, legacy-register sweep, blueprint re-score, full-program code review).

### Wave 0 evidence block (2026-06-11, fresh session #1)

**Branch:** `qa/iam-area-membership`. One commit per item per spec §4. Pre-existing uncommitted `scripts/check-system-runnable.ps1` modification left untouched (not mine).

| Item | Commit | Verification commands → outcome |
|---|---|---|
| 0.1 secret removal + redaction (F-18, D-4a) | `58cbf9943` | `go build ./...` OK · secret grep: `git grep -F <secret>` = 0 hits AND full working-tree scan (incl. hidden, `.env`) = 0 hits. Beyond-card finds redacted: `scripts/start-api-no-build.ps1:5`, `scripts/start-worker.ps1:5`, `_artifacts/stage1/repo-topology.md`. Card deviations (verified): `bin/metaldocs-api.exe` and `scripts/api-lint/api-lint.exe` were never git-tracked (`git ls-files`/`git log` empty) → on-disk delete + source rebuild (SHA-256 `5660295764021D0DA9783EC27CE44EC40511AA0AA9A0BC5535351EAB43952C8D`) instead of `git rm` |
| 0.2 gitleaks CI + D-4a rule (F-18) | `3402a8bbd` | Workflow YAML parses (yq) · gitleaks v8.24.3 over `git archive HEAD` + `.gitleaks.toml`: **no leaks, exit 0**. 10 raw findings dispositioned: test fixtures, `.gen.go` spec chunks, stale plan-doc token (verified ≠ current `.env` value), runbook sample password — allowlisted with rationale in `.gitleaks.toml` |
| 0.5 jobs deployment (F-19-deployment, REQ-ASYNC-4) | `baf6e1b78` | **Runtime-verified** (Docker up): `docker compose config` parses (11 services) · `compose-jobs` image built · `up -d --no-deps jobs` → log `MetalDocs Jobs running (queues=temporal)`, container stays Up. Left running |
| 0.6 observability layering (F-06a, REQ-TOP-2) | `8e0aa9eb4` | `go build ./...` · `go vet` · `go test ./internal/platform/observability/...` OK · grep: zero `modules/` imports in `platform/observability`. **Discovered:** `platform/{authn,bootstrap,docgenv2,objectstore,security}` still import modules → recorded for Wave 1 item 1.10 guard scoping |
| review disposition | `e2b5b2aa4` | `/code-review` (7 finder angles) over wave diff. Fixed: dead cilint tx-allowlist entry; unanchored gitleaks path regexes; WithUserIDResolver set-before-traffic contract doc; `.dockerignore` agent dirs. Refuted with evidence: gitleaks config auto-detection (empirical), Dockerfile flags/base mirror worker.Dockerfile per card, exes never tracked, `depends_on api(healthy)` spec-mandated. Deferred (recorded in tracker/register): compose PG-env YAML anchors (style-consistent with existing), jobs healthcheck (worker has none either), `METALDOCS_JOBS_RIVER_SCHEMA` plumbing → Wave 1.6, tracked `docs/superpowers/plans/` leftovers (stale tokens, verified non-live) |
| 0.3 rotation (USER) | — | Owner **declined rotation** and explicitly accepted residual risk ("keep this work professional, no secret"); `.env` working value verified ≠ leaked string. Recorded as ➖ owner-waived |
| 0.4 history rewrite | this session | `git filter-repo --replace-text` (3057 commits) on a fresh mirror clone → force-pushed all `refs/heads/*`+`refs/tags/*`. Executed AFTER the evidence push. **Verification corrected in round-2 review:** a fresh *mirror* clone (fetches PR refs) proves all 45 branches + tag are clean — surviving secret commits are reachable ONLY via `refs/pull/*` (73 immutable PR refs). The original "plain-clone empty" claim was insufficient (a plain clone skips PR refs). Leaked value ≠ current `.env` (stale). Residual = the 73 PR refs. **RESOLVED-BY-PLAN (owner, 2026-06-11):** closed permanently at first release via re-baseline into a NEW repo (first commit = v1 release), abandoning all prior history with the old repo — no GitHub-support purge needed. Dead dev credential meanwhile (≠ current `.env`). ~22 stale LOCAL branches hold it in this clone's object DB (local-only, gone at re-baseline) |

**Wave close:** full `go vet ./...` clean · touched-package tests green (`observability`, `scripts/api-lint`, `tools/cilint`) · api-lint not run (no contract surface touched) · tracker rows updated same-commit per item · legacy-register F-06/F-18/F-19 resolution notes added. **After the 0.4 force-push every other clone must be re-cloned; this working repo was reset onto the rewritten origin.**

### Wave 0 round-2 review (2026-06-11, 5-subagent audit vs definitions + standards)

Independent audit (one sonnet reviewer per item + a protocol auditor) vs the cited standards (CWE-798, OWASP ASVS, 12-Factor, Hexagonal). Verdicts: 0.1 met; 0.6 met-for-card; 0.2 / 0.5 / 0.4 partial. Net: engineering goals met, nothing redesign-grade. Dispositions applied this round:

- **0.4 verification corrected** (above) — the headline finding. The original closure overstated cleanliness; the honest end-state (branches/tags clean, secret survives only in immutable PR refs, leaked value is stale) is now recorded. No new secret exposure; the value is dead.
- **0.5 completed:** added `METALDOCS_JOBS_RIVER_SCHEMA` passthrough to BOTH the `api` and `jobs` compose services (api reads the same var via `LoadJobsConfig`, `main.go:443`) so the River schema cannot silently diverge. The dual-`MigrateRiverSchema` owner fix stays at Wave 1.6.
- **0.6 hardened:** `WithUserIDResolver` setter → **required `NewHTTPObservability` constructor param**, so a caller can't silently omit it and degrade every access log to `anonymous` (audit-attribution loss). Sole call site updated; no tests construct it.
- **0.2 reassessed (kept as-is + documented):** attempted allowlist scoping via `[[allowlists]]` AND-form FAILED — gitleaks v8.24.3 silently ignores the plural schema (empirically: 10 leaks vs 0 with the singular `[allowlist]`). Full git-history mode deferred to **Wave 1.11** (surfaces ~15 historical dummy-token findings needing triage). The shipped singular allowlist (test/gen/plans/artifacts paths + StrongPass dummy-literal regex) is clean (exit 0) and its low-risk exemptions are now documented in `.gitleaks.toml` with rationale (plans/ is gitignored-going-forward; StrongPass exempts only that one dummy literal everywhere).
- **F-18 history residual — resolved by plan, not a trigger:** the owner will re-baseline into a fresh repo at first release (first commit = v1 release), so ALL prior history — PR refs, stale local branches, the lot — is abandoned with the old repo and the maintained project starts clean. This permanently closes the git-history half of F-18; no GitHub-support purge is pursued. Until release it is a dead dev credential in a single-user repo.
- **Deferred (Wave items):** gitleaks full-history (Wave 1.11); platform-wide REQ-TOP-2 (Wave 1.10 / Wave 2).

### Wave 1 evidence block (2026-06-11, fresh session #2)

**Branch:** `qa/iam-area-membership`. One commit per item per spec §4; tracker rows updated same-commit. Scope: items 1.1–1.10 (tracker row 1.11 gitleaks-full-history was NOT in this session's assignment — still open).

| Item | Commit | Verification commands → outcome |
|---|---|---|
| 1.1 chain reorder + recovery + pre-auth login limit + REQ-MW-7 test (F-01) | `58d7009d9` + probe `2d8728c6b` | Chain `panicRecovery → httpObs → cors → origin → preAuthLoginLimit(10/min IP) → authn → iam → presence → rateLimiter → mux` via declarative `apiChain`/`buildChain` + order test. New `platform/middleware.Recovery`; httpObs outside authn with panic-safe 500 metric recording; **principal slot** (`observability.SetPrincipal` from authn) preserves access-log attribution. **RUNTIME-VERIFIED** (start-api.ps1 -Build, :8081): login 200 · authed panic probe → 500 problem+json, process alive · metrics show 401s (`iam/users` errors=2) + panics · panicked request logged `user_id="admin"` with same trace_id as the stack-tagged panic line · 12 rapid bad logins → 11×401 then **429** |
| 1.2 server timeouts (F-16) | `1b3f0d11d` | `ReadTimeout 30s/WriteTimeout 60s/IdleTimeout 90s`; WS presence unaffected (hijacked conns). build+vet |
| 1.3 delete spec2.yaml + internal/api/v2 + CI gate (F-03) | `880f33c4a` | −1249 lines. 3 contract tests → `problem.Problem`; golangci exclusion dropped. `capability-catalog-hash` job DELETED (hashed a never-existing seed; always exit 0) — real REQ-AUTHZ-5 guard = api-lint registry rules (ADR 0022). Both YAMLs parse (python yaml); tests green |
| 1.4 idempotency codes (F-09 half) | `ed7890597` | `writeErrJSON` typed `problem.Code`; new consts `IDEMPOTENCY_KEY_INVALID`/`REQUEST_BODY_TOO_LARGE`; conflict → `IDEMPOTENCY_KEY_REUSED` (fixes pre-existing FE mismatch); guard test covers `platform/idempotency`. FE catalog regen (112→113 codes) + PT-BR messages; FE coverage test green |
| 1.5 search contract + 405s (F-13a/b, D-03) | `2e977845b` + `c76fd6cfc` | Response aligned TO spec (4 never-populated fields dropped; **no spec change/regen**); camelCase `businessUnit` removed; 4 bare 405s → problem+json + `Allow` via shared `httpresponse.WriteMethodNotAllowed`. **Runtime:** authed POST search → `405 {"code":"METHOD_NOT_ALLOWED"}` + `Allow: GET` |
| 1.6 River migration single owner (F-19) | `38d7cc8ce` | Call removed from `BuildJobsDependencies`; API sole owner; ordering = compose `depends_on: api(healthy)`. **RUNTIME-VERIFIED:** jobs image rebuilt on wave code, container Up, "MetalDocs Jobs running" |
| 1.7 lease_reaper JOIN bug (F-19) | `2512f502f` + `c76fd6cfc` | `public.documents` subquery removed (every reap errored + dropped its event while deleting the lease). System-scoped → structured slog (card option; `governance_events.tenant_id` NOT NULL, no system tenant). Regression test bans the cross-schema lookup. **Integration probe run against live Docker PG: PASS** (pre-wave code could not pass it) |
| 1.8 templates audit sink split (F-07-sub-split) | `3a257a9cf` | `ListAudit` reads `metaldocs.audit_events` (`resource_type='template'`); `version_id` now carried in payload on write (old write path silently dropped it) and lifted on read. Historical `templates_audit_log` rows = accepted seam per card |
| 1.9 .gitkeep cleanup (F-08) | `9ff2c0d6c` | 4 files removed; empty `platform/cache/` dir deleted |
| 1.10 import-boundary CI guard (F-06a lock-in) | `84f465efc` | New cilint `platformboundary` analyzer (invariants.yml blocking gate over `./...` — chosen over a Go test because ci.yml tests only a platform subset). Frozen per-package baseline with triggers: bootstrap/authn (Wave 2 review), docgenv2/objectstore (Wave 2 F-06), security (dies 2.8). 4 unit tests; repo run exit 0 |
| review disposition | `c76fd6cfc` | 7-angle `/code-review` over wave diff. **Fixed by family:** test-contract drift (integration probe asserted the removed `lease.reaped` row — updated + live-DB-verified); 405 duplication (shared helper); login-path literal duplication (`authdelivery.PathLogin`). **Refuted with evidence:** rows.Close/tx.Rollback defer order (LIFO runs Close first); trace-ID divergence (`Resolve` reads ctx first — idempotent); TraceID assertion loss (old tests never asserted it); change-password pre-auth gap (route is session-authenticated → post-auth territory, F-20e/Wave 2.10). **Accepted/deferred with rationale:** ErrAbortHandler counted as 500 in RED metrics (rare; defer can't see panic value without recover-repanic complexity); `TrustedProxyCIDRs` nil collapses login buckets behind an LB (config risk, single-server D-1 deployment, API directly exposed; revisit at RF-2 tail); recovery write-after-partial-response corruption (inherent, documented in code); main.go `userIDResolver` fallback effectively dead in prod chain (kept — Wave 0 round-2 constructor contract; revisit Wave 2); platformboundary package-level (not per-import) allowlist + test-file exclusion (baseline freezes Wave-2-doomed packages; tightening noted for Wave 2); version_id-in-payload schema implicitness (per-card accepted seam); raw `"IDEMPOTENCY_KEY_REQUIRED"` literal in controlleddocuments handler (pre-existing, outside wave scope → guard-scope expansion candidate, Wave 2); e2e panic probe env-gated but not build-tag-gated (authn-required + METALDOCS_E2E-gated; accepted for Wave F runtime QA usability) |

**Wave close:** full `go vet ./...` clean · 63 touched-tree test packages all `ok` (incl. `-tags integration` vet + live-DB lease-reaper probe) · `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` → **0 violations** · `go test ./scripts/api-lint/...` ok · FE error-message coverage test green · legacy-register resolution notes added (F-01, F-03, F-07-sub, F-08-partial, F-09-half, F-13a/b, F-16-half, F-19-complete, D-03) · tracker rows final with hashes. **Environment note:** a `.gitnexus/` tool-cache directory (193MB `lbug` file) appeared untracked in the worktree mid-session and breaks `git add -A` with an mmap error — files were staged explicitly; consider gitignoring `.gitnexus/`.

## Recovered agent memory (wiped by the reinstall — re-create if a memory system returns)

These facts lived in the agent's persistent memory and are NOT derivable from the repo:

- **Machine:** C: SSD (XPG S50 Lite) has degraded writes (~7-16 MB/s; reads fine). Hardware/firmware issue. Prefer `D:` for heavy-write work (builds, caches) when possible.
- **Impeccable adoption plan:** adopt the "impeccable" design skill only AFTER IAM PR-2, as a guidance layer over MetalDocs design tokens (not a replacement); needs a bridge doc first.
- **Authz program intent:** the user explicitly chose root-cause coherence (ADR 0022) over symptom-patching the membership 403 — keep honoring that.
- **Backend standardization parameters:** approved as a 6-family batched-workflow program; H2 finding downgraded to LOW; NEW-1 was the real HIGH; idle timeout policy 30m sliding.
- **User working style:** wants industry-grade/professional standards, definitions-before-implementation, evidence-based closure; caveman (terse) chat mode was active in Claude sessions — irrelevant for docs/code, which stay normal.

## Operating rules that remain binding (pointers, not copies)

- `CLAUDE.md` — skill routing table, mandatory gates, hard-stop rule, evidence rule, close-out loop. Read first, always.
- `wiki/quality/qa-operating-system.md` — canonical QA/close-out policy.
- Startup: `.\scripts\start-api.ps1` only (script-truth policy). Dev login: `POST /api/v1/auth/login` `{"identifier":"admin","password":"AdminMetalDocs123!"}`.
- Wiki drift policy: code change touching a documented surface bumps `Last verified` same change; dispatch `wiki-curator` after refactors.
- **New since 2026-06-10:** backend work is reviewed against `backend-target-architecture.md` REQ IDs; deviating from a MUST requires an ADR.

## How to proceed in a fresh session

1. Read `CLAUDE.md` → `wiki/README.md` → this file → the three-doc stack.
2. `git status --short`; confirm branch; commit the doc stack if still uncommitted.
3. Pick up at "Next steps" item 2 (ADR 0022 Phase 6) unless the user redirects.
4. Do not re-execute anything from prior session summaries without verifying against the working tree — most of it is already done.
