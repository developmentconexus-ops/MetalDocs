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

### Wave 2 evidence block (2026-06-12, fresh session #3)

**Branch:** `qa/iam-area-membership`. Orchestrator + balanced-model subagents (sonnet implement, haiku/sonnet review; never Fable for execution). One commit per item per spec §4; tracker rows final with hashes (same-commit per item). Docker stack up — runtime proofs executed live. All 11 Wave 2 items (2.1–2.11) closed.

| Item | Commit(s) | Verification commands → outcome |
|---|---|---|
| 2.1 ADR 0027 RLS sequencing (D-3, F-12, T-008) | `81213133` | ADR records auth_identities tenant-global by design (T-008 closed by-design) + RLS rollout sequence (2.3 = controlled_documents+audit_events, RF-6 = iam_users, trigger-gated = rest). GUC `metaldocs.tenant_id` confirmed in `iam/authz/context.go`. Decisions index + roadmap updated same commit |
| 2.2 in-tx audit/governance + cilint guard (F-07, D-01, D-2) | `6cc9595` + `510bc1b` (review) | `LogTx`/`RecordTx`/`WriteTx` on GovernanceLogger/AuditWriter; 11 taxonomy event types + templates lifecycle + documents ForceRelease/Archive + CD changeStatus moved in-tx; approval path was already the template. New `PostCommitAudit` cilint analyzer forbids audit-sink-after-Commit() in modules/**. Review fixes: LogTx nil-fallback removed (return error), atomicity rollback test added, analyzer test renamed. build+vet+targeted tests+cilint exit 0. **RUNTIME-VERIFIED:** taxonomy family create via API → governance row `family.created` present in `audit_events` written in the mutation tx |
| 2.3 RLS controlled_documents+audit_events; ExportJob.TenantID uuid (F-12) | `de0b44c2` | Migration `0234`: ENABLE+FORCE RLS + NULL-permissive `tenant_isolation` policy (GUC-less system paths preserved). Dev `metaldocs_app` is superuser → policies load-bearing only under prod NOSUPERUSER+NOBYPASSRLS (per ADR 0022 §Item 7 / ADR 0027). **RUNTIME-VERIFIED** (NOSUPERUSER `rls_test_role`): GUC-unset → both tenants; GUC=A → only A; GUC=B → only B. `ExportJob.TenantID` string→uuid.UUID (3 compile sites + test UUIDs). Migration in `schema_migrations`, idempotent. Probe rows cleaned up. build+audit/CD tests green |
| 2.4 capability fixes + tier-1 lint (F-11, REQ-AUTHZ-2, ADR 0022) | `784ce561` | `"template.admin"` (unknown cap, route permanently locked) → CapTemplateEdit (matches tier-2); publish tier-1 → CapTemplatePublish; all templates-delivery raw tier-1 literals typed; broken `containsRole(...,"admin")` post-publish gate → canonical system_admin/qms_admin; typed `route.config.*` EventTypes; new api-lint `no-rawstring-tier1-authz` rule (4 tests, would have caught the defect). No new cap minted (no ADR). 8-route tier-1/tier-2 pairing table all ✅. build+vet+tests+`api-lint -strict`=0 |
| 2.5 extract delivery SQL (F-06b, REQ-H-1, REQ-TOP-1) | `ea996da2` + `e9e6e2dc` (review) | CD `GetActiveDocument` (3 inline queries) → `GetActiveInstance` repo+service+`ActiveDocumentInstance` domain type; documents `finalizeDocument` (4 inline queries) → `GetFinalizePrereqs` repo+`FinalizePrereqs`+sentinels. Grep: 0 `QueryRowContext/QueryContext/ExecContext` in both delivery files; idempotency block untouched. Review fix: restored exact `NO_ACTIVE_INSTANCE` 404 contract code (refactor had drifted it to `CONTROLLED_DOCUMENT_NOT_FOUND`) via new `ErrNoActiveInstance` sentinel + service/handler tests. build+vet+tests green |
| 2.6 auth→iam via LoginContextPort (F-06c, REQ-TOP-1) | `07f914e9` | `iamdomain.LoginContextPort` + `iampg.LoginContextRepository` (exact SQL, zero-rows no-op); `RecordLastLoginContext` removed from authdomain.Repository + impls; `Service.loginCtxPort`+`WithLoginContextPort`; login swallows error as before; wired in main.go `if deps.SQLDB != nil`. Grep: no `UPDATE metaldocs.iam_users` under auth. build+vet+auth/iam tests green (13 pkgs) |
| 2.7 promote session types to auth/domain (F-06d, REQ-TOP-1) | `3c6ab235` | `SessionAdminQuery`/`SessionListItem` moved (single def) to `auth/domain/session_admin.go`; iam/delivery drops `authpg` import (structural typing keeps wiring). Grep: `auth/infrastructure` → 0 under modules/iam. build+vet+tests green |
| 2.8 delete legacy limiter; activate platform/ratelimit (F-05, D-04, REQ-TOP-2, REQ-MW-5) | `0b41d1c1` | `platform/security/ratelimit.go` (~200 lines, REQ-TOP-2 breach) deleted; `GlobalEnvelopeWrap` 120/min user→IP in the `rate_limit` chain slot (same position — chain-order test green); documents routes `RegisterRoutesWithRateLimit` (60/30/20); dead `METALDOCS_RATE_LIMIT_*` envs removed; `platform/security` dropped from cilint platformboundary baseline. build+vet+tests+cilint exit 0. **RUNTIME-VERIFIED:** 130-req burst on `/iam/users` → 64×200 then 429 problem+json `RATE_LIMITED` + `Retry-After`; recovers after refill |
| 2.9 N+1 fixes (F-10, REQ-DATA-2, REQ-AUTHZ-6) | `db384188` | (a) RolesByUserID 2→1 LEFT JOIN; (b) RolesByUserIDs batch + CachedRoleProvider read-through + auth ListUsers; (c) VerifyUserInTenant→EXISTS via TenantMemberChecker; (d) approval ListPendingForActor→LoadInstancesByIDs batch; (e) audit ILIKE→indexed cols only (payload scan dropped, undocumented). Authz semantics (a/b/c) byte-identical outputs, fewer queries. build+vet+tests green |
| 2.10 Postgres auth-failure counter (F-20e, REQ-REL-3, D-1) | `9921b323` + `6e2bf39d` (critical fix) | Migration `0235` `auth_failure_counters` + `PostgresAuthFailureRateLimiter` (exact in-memory semantics: 60s window, 5 threshold, actorID key, reset=delete); wired in reauth.go when db!=nil. **Critical fix:** original `timestamptz + $3(time.Duration)` errored at runtime under pgx (duration→bigint, no operator); sqlmock masked it. Fixed to precomputed-timestamp comparison. **LIVE integration probe** (`-tags integration`, all 5 state transitions pass). Dictionary page added. build+vet+16 sig tests green |
| 2.11 dead-code deletion + govLogger fallback (F-14, F-07-sub, D-06, CWE-561) | `63f74368` | Grep+compiler reference-checked (GitNexus stale). Deleted: CutoverService(+test), CompositionConfig(+test), AreaService.SetParent(+tests), resolvePermissionFallback, WorkerConfig.ReviewReminderDays(+env), coverage_boost CutoverService section, templates legacy INSERT columns (CreateTemplate), dead platform/config/ratelimit.go. DBGovernanceLogger nil-fallback in CD → fail-loud panic. pdfDispatcher nil-guard added. Out-of-scope left per card: FreezeService.Freeze, SnapshotService, document_subjects, RepositoryMemory. build+vet+tests+cilint exit 0 |
| review disposition (full-wave) | `5a6b407b` | 7-angle `/code-review` (high effort) over the wave diff `7f3e734cc..63f74368`. **Fixed by family:** (1) pdfDispatcher fully removed — field+param+guard+dead 45s-timeout dispatch block + test struct-literal initializers (the constructor guard was bypassable via struct literals; deprecated path now gone, not just guarded); (2) 2.10 limiter — DB error in `Allow()` now logged distinctly before fail-closed return; stale "prunes rows" comment corrected (rows persist+reset, not deleted); (3) DBGovernanceLogger.LogTx nil-tx guard added; (4) `GetActiveInstance` stale doc-comment fixed; (5) `LoadInstancesByIDs` dead `order` slice removed + input IDs deduped (duplicate IDs no longer yield duplicate rows); (6) TenantMemberChecker type-assertion wiring now logs a warning on miss (silent full-scan regression made observable). **Refuted/accepted with rationale:** RLS per-row `current_setting` cost (NULL-permissive policy is the ADR-0027 tripwire altitude); GlobalEnvelopeWrap double-limit on autosave (intentional — per-route + global envelope are layered by design); audit payload-search drop (undocumented, F-10e accepted); PG limiter check-then-act race (single-replica D-1; ±1 failure tolerance acceptable, fail-closed); UTF-8 truncation split (best-effort field, pre-existing helper behavior); unbounded ANY() in ListUsers (pre-existing full-tenant load, separate from this wave's batching). build+vet green; LoadInstancesByIDs+limiter behavior verified |

**Wave close:** `go build ./...` ✓ · `go vet ./...` ✓ · full touched-tree test run **84 packages all `ok`, 0 failed/skipped** (incl. `-tags integration` PG auth-failure probe + live RLS verification) · `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` → **0 violations** · `go run ./tools/cilint/...` exit 0 (now 4 CI guards: gitleaks, platformboundary, **PostCommitAudit**, chain-order test) · **OpenAPI contract surface untouched** (`git diff 7f3e734cc..HEAD -- api/openapi/` empty — expected, no contract changes this wave) · wiki sync via metaldocs-module-doc-sync for all 7 touched documented modules (taxonomy, templates, documents, approval, iam, auth, controlled-documents — stamps bumped 2026-06-12) · legacy-register resolution notes added (F-05, F-06b/c/d, F-07-atomicity, F-10, F-11, F-12-partial, F-14, F-20e, D-01, D-04, D-06) · ADR 0027 merged · tracker rows final with hashes. **NOT merged to integration branch (per directive — user review gate before Wave F).** Residual defers recorded inline per register finding (iam_users RLS → RF-6; remaining RLS tables → first-external-tenant; taxonomy DBGovernanceLogger dev-fallback → next-touch; F-06 structural residuals; CreateTemplateTx orphaned columns → drop-migration window; lower-priority F-10/F-20 rows → next-touch). **Environment:** `.gitnexus/` cache dir still breaks `git add -A` (staged explicitly all wave).

### Wave 2.12 / 2.13 evidence block (2026-06-12, fresh session #4 — fallback elimination + single-mode)

**Branch:** `qa/iam-area-membership`. Orchestrator + balanced-model subagents (sonnet implement/review; never Fable for execution), two-stage review (spec + code-quality) per commit. Executes the APPROVED spec `docs/superpowers/specs/2026-06-12-wave2-fallback-elimination-design.md` (FE-1..FE-7) as roadmap rows **2.12** (ports) and **2.13** (single-mode + dead-schema + CI locks). Docker up — runtime proofs live.

**Row 2.12 — fallback elimination (ports):**

| Item | Commit(s) | Verification → outcome |
|---|---|---|
| 2.12.1 `db.Tx`/`db.DB` interfaces (FE-1) | `c02c1d8f` | New `internal/platform/db`; `*sql.DB`/`*sql.Tx` satisfy structurally; compile-assert test. No callers yet (grep 0) |
| 2.12.2 domain ports speak `db.Tx`; delete Unwrap chain (FE-2, M-1) | `e3bf7b62` + `56477596` (tracker fix) | 56-file sweep across audit/governance/authz-path ports; `taxonomyTx.Unwrap`/`familyTx.Unwrap`/`sqlTxFromFamilyTx` deleted; `FamilyTx` embeds `db.Tx`; `mustSQLTx` shim contained to documents/repository (infra). `git grep '*sql.Tx' -- **/domain/` = 0. build+vet+test green |
| 2.12.3 delete `DBGovernanceLogger` (FE-3, F-07-sub) | `242b301a` | `governance_logger.go` deleted; `taxonomy/module.go` panics if `AuditWriter` nil. **Proof A re-run:** `POST /api/v1/taxonomy/families` → 201; in-tx `family.created` row in `audit_events` (`fa87d3a8…`). Zero `// Deprecated` in modules (one pre-existing on out-of-scope `SnapshotService`) |
| 2.12.4 `RoleProvider.UserActiveInTenant`; delete `TenantMemberChecker` (FE-3, F-10 tail, M-6) | `28b533a1` | Method on every provider; `TenantMemberChecker`/`WithTenantMemberChecker`/`tenantChecker`/ListUsers-fallback + main.go type-assert deleted. Memory `RolesByUserIDs` now filters tenantID (M-6). **Live probe** `TestRoleProvider_UserActiveInTenant_Live`: true/false PASS |

**Row 2.13 — single-mode + dead-schema + CI locks:**

| Item | Commit(s) | Verification → outcome |
|---|---|---|
| 2.13.1 single-mode (FE-4, F-08) | `75ad1298` + `15965ad0` (fillin tail) | All class-A `db==nil` branches removed (templates/documents/controlleddocuments/approval/auth); the CD `Create` **class-B authz-bypass** (`db==nil` skipped authn+authz — user-approved via fact-check gate) deleted, authz now unconditional; `InMemoryAuthFailureRateLimiter` + `RepositoryMemory` prod path deleted; `auth.Service.loginCtxPort` required (panic on nil); fillin `NewFillInServiceNoAuthz` authz-bypass eliminated. **Runtime:** API boots `repository=postgres`, login 200, authed `GET /iam/users` 200; live PG limiter probe (5 transitions) PASS |
| 2.13.2 dead-schema migration 0236 (FE-5, F-14 tail) | `dce4c81d` | Drops `public.templates_template.areas/visibility/specific_areas` (real table corrected from the plan's phantom `metaldocs.templates`), `document_profiles.is_active`, `document_subjects` (CASCADE). Writer removals same commit (`CreateTemplateTx`, reference-data 0001, e2e_seed). **Live-applied + ledgered**; post-drop `POST /templates` → 201. Dictionary pages updated/retired |
| 2.13.3 CI seam locks (FE-6) | `d7ab39ba` (CD `DBTX`→`db.Tx`) · `ab949015`+`49e329be` (audit/iam ports) · `0ab280b0` (analyzers) | `nosqltxindomain` (scoped to `sql.{Tx,DB,Rows,Row,Result}`, allows `sql.Null*`) + `nodualmode` (exempts fail-loud panic guards) wired into `RunAll`, CI-gated. **The analyzers surfaced 9 violations the Task-5 fact-check missed — its git pathspec `**` glob silently matched nothing.** Dispositioned: CD domain `DBTX` duplicate → `db.Tx`; iam membership logger required→best-effort (T-007, post-commit torn-write avoided); 4 justified `//cilint:allow-dualmode` (audit export feature-gate; freeze ×3 ADR-0015 optional-tx-enlistment). `go run ./tools/cilint/...` exit 0 |

**Process note — Task 6 salvage:** the first single-mode implementer bundled cross-branch contamination (6 frontend, 4 perf/runbook, 2 snake_case delivery files) from a stale `qa/std-execution` stash pop into its commit. The commit was `git reset --soft`, the 14 contamination/cosmetic files reverted to parent, and the verified single-mode work re-committed clean (`75ad1298`). No contamination reached the final history (`git diff 56c58f3a4..HEAD -- frontend/ scripts/perf/` empty).

**Wave close:** `go build ./...` ✓ · `go vet ./...` ✓ · `go test -p 2 ./...` **exit 0, zero failures** (incl. `-tags integration` live PG probes) · `go run ./scripts/api-lint/ -strict …` → **0 violations** · `go run ./tools/cilint/...` exit 0 (now **6 CI guards**: gitleaks, platformboundary, PostCommitAudit, chain-order, **nosqltxindomain**, **nodualmode**) · **OpenAPI surface untouched** (`git diff 56c58f3a4..HEAD -- api/openapi/` empty). **Final holistic review: WAVE VERIFIED ✅** — all 8 closure-claim items PASS (zero `// Deprecated` new, zero `*sql.Tx` in domain, zero unjustified dual-mode, analyzers green, 0236 applied, loginCtxPort required, memory tenantID honored, 4 removed types gone). Wiki synced (16 module docs + tech-debt registers: closed taxonomy T-010, iam T-007, CD T-008; opened documents T-013 freeze, audit T-012 writer, CD T-010 subject_code orphan). Legacy register: F-07-sub/M-1, F-08 RepositoryMemory, F-14 tail, F-10 tail/M-6 marked resolved. Tracker rows 2.12/2.13 ✅ with hashes.

**Deferred (next-touch / Wave-3 triggers, NONE blocking close):** audit `Service.writer` hard-required refactor of `WithExports`; `freeze_service` in-tx-only collapse (ADR-0015); iam membership in-tx atomic governance via `RecordTx` (T-007, currently best-effort post-commit); orphan `metaldocs.documents.subject_code` column+index (FK CASCADE-dropped by 0236); out-of-scope `SnapshotService` (pre-existing `// Deprecated`). **NOT merged to any integration branch — DONE_AWAITING_REVIEW per D-5 (user review gate before Wave F).** Environment: `.gitnexus/` still breaks `git add -A` (staged explicitly all session); C: SSD memory pressure → tests run `-p 2`.

### Wave F closing report — PROGRAM COMPLETE (2026-06-12, fresh session #5)

**Branch:** `qa/iam-area-membership`. Wave F is VERIFICATION-only (code touched solely to fix what verification found, each fix its own commit). Rows F.1–F.8 in [`wiki/backend/roadmap.md`](../backend/roadmap.md) carry full evidence; this is the program-level sign-off.

**Per-wave evidence summary:**
- **Wave 0 (P0 prerequisites):** F-18 secret scrubbed + gitleaks CI (history residual closed-by-plan, D-4b fresh-repo re-baseline at release); F-19 jobs Dockerfile+compose; F-06a observability layering. Owner waived rotation (0.3) and accepted residual.
- **Wave 1 (high-value/low-blast, 1.1–1.10):** middleware chain reorder + panic recovery + pre-auth login limit + REQ-MW-7 test; server timeouts; spec2/api-v2 deleted; idempotency typed codes; dead search fields + 405s; River single-migration owner; lease_reaper JOIN bug; templates audit sink split; .gitkeep cleanup; `platformboundary` CI guard. (1.11 gitleaks-full-history still open — out of session scope.)
- **Wave 2 (structural, 2.1–2.11):** ADR 0027; in-tx audit/governance + `PostCommitAudit` guard; RLS on controlled_documents+audit_events; capability-literal fixes + `no-rawstring-tier1-authz`; delivery-SQL extraction; auth→iam `LoginContextPort`; session-type promotion; rate-limiter swap (platform/security/ratelimit deleted); N+1 fixes; Postgres auth-failure counter; dead-code deletion.
- **Wave 2.12/2.13 (fallback elimination + single-mode):** `db.Tx` seam; Unwrap chain + `DBGovernanceLogger` + `TenantMemberChecker` deleted; class-A dual-mode + class-B authz-bypass removed; `RepositoryMemory` prod path deleted; dead-schema migration 0236; `nosqltxindomain` + `nodualmode` analyzers. **6 CI guards total.**
- **Wave F (this session):**
  - **F.1 — 6 CI guards green:** `go run ./tools/cilint/... ./...` exit 0 (platformboundary · postcommitaudit · nosqltxindomain · nodualmode + legacy) · chain-order test PASS · gitleaks accepted on Wave 0 evidence (D-4b).
  - **F.2 — static pass:** `go build ./...` 0 · `go vet ./...` 0 · `go test -p 2 ./...` **86 ok / 0 FAIL** · `api-lint -strict` **0 violations**.
  - **F.3 — runtime QA (Docker up):** login + authed smoke; **all 5 🔁 re-checks LIVE** (RLS cross-tenant block via NOSUPERUSER probe · in-tx `family.created` governance row · rate-limit 429 + account-lockout 403 · panic→500 problem+json + process alive · 401s counted in RED metrics); **worker** claims+relays `pdf_dispatch_outbox`→`outbox_events` LIVE; **jobs** host claims+executes a `scheduled_publish_cutover` River job LIVE; **full doc→PDF render** lands a real 11480-byte `%PDF` in MinIO via docx-renderer+gotenberg + `document_exports` row + warm-cache hit. All probe fixtures cleaned up.
  - **F.4 — legacy-register sweep:** all 20 F-* + 7 D-* entries now resolved (commit) or deferred (trigger); 7 silent leftovers (F-02/F-04/F-15/F-17/D-02/D-05/D-07) given explicit defer notes + a sweep tally.
  - **F.5 — blueprint re-score:** §7 Wave F re-score block (grade deltas: C6/D5/D3 ✅ now *earned*; D2/C4/D7/D8 clarified to trigger-gated; B2/A3 honestly NOT promoted) + 15-row REQ-* compliance table; D9 analyzer count 6→7; stamps bumped.
  - **F.6 — full-program code review:** 5 parallel sonnet subagents over the cumulative diff `d3f5dc62b..HEAD` (325 files/+26681). Fixed by family (commit `f698d1fd2`): outbox `Enqueue` fail-loud on nil tx (db.Tx atomicity) + regression test; audit `ExportEvents` logs dropped governance event. Dispositioned not-fixed: B-1 (CD ctor nil-db guard — over-guards impossible-in-prod scenario, breaks 19 tests), D (CD handler raw codes — correct values, guard-scope expansion → flagged as follow-up), PG-limiter race (accepted, D-1 single-replica).
  - **F.7 — wiki coherence:** wiki-curator refreshed `backend/index.md` (F-19/V-01/V-03/dead-code/governance closed), `audit.md` + `audit-tech-debt.md`, `render-fanout.md` (fail-loud contract); indexes/anchors/stamps/cross-links verified.

**All defers carry written triggers (NONE block program close):**
- **Wave 3 trigger-gated (by design):** F-17 OTel/W3C (→ "second host or first external-tenant SLA", D-1) · F-12 remaining RLS tables (→ first external tenant) · F-16C readiness concurrency (→ second dependency) · F-20b CD trigram index (→ >50k rows or p95>200ms) · F-04 outbox generics + F-06e taxonomy port + F-15 parseBoolEnv + D-02 MinIO clients + D-05 cache contract + D-07 status enum + F-13c/d/e (→ next touch of each area) · F-16B WS drain (→ K8s grace insufficient).
- **Next-touch tech debt:** audit `Service.writer` hard-required `WithExports` refactor (T-012) · `freeze_service` in-tx-only collapse (ADR-0015) · iam membership in-tx atomic governance via `RecordTx` (T-007) · orphan `metaldocs.documents.subject_code` column (FK CASCADE-dropped by 0236, CD T-010) · F-09 typed-code guard extension to controlled-documents HTTP handler (F.6 finding D) · 1.11 gitleaks full-history.
- **Owner-waived / resolved-by-plan:** F-18 credential rotation (waived) + git-history residual (fresh-repo re-baseline at v1 release, D-4b).

**ADR links:** [0009](../decisions/0009-pdf-dispatch-outbox.md) / [0015](../decisions/0015-async-freeze-pin-materialize.md) (transactional outbox) · [0022](../decisions/0022-authz-capability-coherence.md) (authz coherence — Phase 6 wiki-sync residual) · [0026](../decisions/0026-unified-authz-enforcement.md) (unified authz enforcement) · [0027](../decisions/0027-rls-adoption-sequencing.md) (RLS sequencing, merged this program). **ADR 0028 not minted** — F-11 reused `CapTemplateEdit`, no new capability (per ADR 0016 precedent).

**PROGRAM SIGN-OFF:** the MetalDocs backend meets the professional-SaaS bar this program defined (design spec §10 completion definition): Waves 0–2 executed with evidence; Wave 3 triggers documented; ADR 0027 merged (0028 N/A); **6 CI guards active and green**; every legacy-register entry resolved-or-deferred-with-trigger; blueprint maturity re-scored; handoff updated. **Verification-grade evidence recorded — no `done`/`green` claims without a command + output.** Further evolution goes through the normal REQ/ADR process; this roadmap is frozen as the historical record.

**NOT merged to any integration branch — awaiting the user's final sign-off before merge (per directive). Environment:** `.gitnexus/` still breaks `git add -A` (staged explicitly all session); C: SSD memory pressure → tests run `-p 2`.

### NEXT SESSION: Wave Z — sealed-scope final close-out (planned 2026-06-12, NOT yet executed)

The operator overrode D-1/D-3 and ordered a zero-defer finish (register-zero + ops-ready) before the v1 release (target Sunday 2026-06-14, operator performs the release re-baseline personally). Binding documents, read in order:

1. **Spec:** [`docs/superpowers/specs/2026-06-12-wave-z-sealed-closeout-design.md`](../../docs/superpowers/specs/2026-06-12-wave-z-sealed-closeout-design.md) — sealed manifest Z-1..Z-33, the **anti-circle rule**, machine-checkable DONE gate G1–G11.
2. **Plan:** [`wiki/backend/wave-z-plan.md`](../backend/wave-z-plan.md) — phases P0–P6, per-task files/steps/verify/commit. (Lives in the wiki because `docs/superpowers/plans/` is gitignored-ephemeral; this plan is durable program record, same precedent as the roadmap.)

Hard constraints: ultracode workflows with **sonnet** implement/review + **haiku** mechanical agents (**never fable workers**), **max 15 concurrent agents**, G1–G4 gates between phases, one commit per Z-task, anti-circle rule on every discovery (pre-existing finds → `wiki/backend/post-v1-backlog.md`, never block). Do NOT merge; present the DONE-gate evidence for user sign-off.

### Wave Z closing report — BACKEND CLOSED (2026-06-13, fresh session)

Wave Z is EXECUTED and the DONE gate is GREEN. The MetalDocs backend is closed for feature work — the only backend work left in this repository is feature work. Frozen Wave Z addendum table (Z-ID → commit → finding) lives in [`wiki/backend/roadmap.md`](../backend/roadmap.md); per-task plan in [`wave-z-plan.md`](../backend/wave-z-plan.md).

**Phases P0–P6 all executed** (ultracode workflows: sonnet implement/review + haiku verify, ≤15 concurrent, never fable; G1–G4 between phases; one commit per Z-task; explicit-path staging — `.gitnexus/` breaks `git add -A`; tests `-p 2` on the C: SSD):
- **P0** preflight: baseline G1–G4 green, Docker up, RLS census re-run (30 tenant_id cols → 27 base tables after excluding the 2 already-covered + 1 view), `post-v1-backlog.md` created (`836d57264`).
- **P1** (8 groups, fan-out): Z-4, Z-6, Z-7, Z-8/9, Z-11, Z-12/13, Z-15, Z-21, Z-23, Z-25, Z-17. One regression caught + fixed in-wave: the P1 Z-6 agent put `MembershipTx` in `iam/application` and made the postgres repo import it → test-binary import cycle; fixed by moving `MembershipTx` to `iam/domain` (`abc9afa48`).
- **P2** big two (sequential, runtime-proofed): **Z-2/Z-3** full-table RLS migration `0237` (all 27 remaining tenant tables, NULL-permissive tripwire policy = 0234 pattern; idempotency tenant FK; ADR 0027 amended executed-in-full) — NOSUPERUSER probe live (GUC-unset→all, GUC=A→only A, GUC=B→only B on iam_users + documents); GUC-less system paths preserved. **Z-1** minimal OTel (`otelhttp` + W3C `traceparent` + `autoexport`, env-gated INERT by default) — proved both inert boot (zero overhead, byte-identical) and console-exporter span (inbound traceparent TraceID bridged into access log + bounded span name).
- **P3** contract/codegen (serialized): Z-16 roles enum → canonical 8 + Z-18 deprecated flag + regen (Go + FE types, tsc clean); Z-19 split iam/audit/security codegen packages; Z-24 pg_trgm GIN index (EXPLAIN-proven); Z-26 gitleaks full-history (0 leaks over 4917 commits).
- **P4** structural M-items: Z-5 freeze tx-mandatory (ADR 0015 amended), Z-10 generic staging outbox + dead-loop delete + idemp dedup, Z-22 WS presence drain. After P4 the **last 3 `//cilint:allow-dualmode` directives are gone → repo-wide count 0** (Z-4 + Z-5). Z-10 left a stale `tripwire-allowlist.txt` (deleted files) → fixed in-wave (`029003c2d`).
- **P5** docs/governance (9 agents + index): Z-27 ADR lifecycle audit (every ADR now carries a canonical `Status:` header; stub 0003 → Historical; stray dated ADR → `0028`; `index.md` rebuilt as a status table; `README` gained the status vocabulary), Z-28 ADR 0022 Phase 6 wiki sync (status → "fully executed"), Z-29 cache contract + invalidation-path verification (all 7 invalidation callers traced, no missing path), Z-30 flag-lifecycle standard, Z-31 messaging/servicebus fence, Z-32 v1 re-baseline runbook.
- **P6 DONE gate (spec §2) — ALL GREEN 2026-06-13:**
  - **G1** `go build ./...` 0 · `go vet ./...` 0.
  - **G2** `go test -p 2 ./...` → **87 ok / 0 FAIL**.
  - **G3** `go run ./scripts/api-lint/ -strict …` → **0 violations**.
  - **G4** `go run ./tools/cilint/... ./...` → **exit 0** AND `git grep -c "cilint:allow-dualmode" -- internal/` → **0**.
  - **G5** manifest empty — every Z-1..Z-33 row has a commit (table above).
  - **G6** register CLOSED — every legacy-register entry resolved-with-Z-commit or KEEP/at-release (no defer-with-trigger remains).
  - **G7** blueprint all-green — scoreboard ✅ with the single documented F-18 "at-release" line; REQ-OBS/TEN/CACHE/REL/ASYNC flipped MET.
  - **G8** ADR index trustworthy — zero status-less ADR files; `index.md` rebuilt with status column.
  - **G9** runtime smoke (Docker up, clean restart): inert-OTel boot (no noise) · login 200 · authed GET 200 · in-tx `family.created` governance row · 401→403 lockout→429 rate-limit · RLS NOSUPERUSER on iam_users (unset=6/wrong=0/real≥1) · panic→500 problem+json + process alive · console-exporter span with bridged traceparent (cited from Z-1).
  - **G10** regression-only review (7 sonnet slices + adversarial verify, scoped strictly to "did Wave Z break what Wave F verified"): **ZERO Wave-Z-caused defects.** Two critical claims were raised and BOTH refuted with evidence — (1) "RLS `::uuid` cast breaks text-typed templates tenant_id" refuted by migration `0213` having converted those columns to uuid post-baseline (live information_schema confirms uuid; RLS evaluates clean under GUC); (2) "Z-22 CloseAll send-on-closed-channel race" refuted as pre-existing (the same `conn.Close()` path races with `broadcast` on client disconnect already; Z-22 did not introduce it). Both parked in `post-v1-backlog.md`.
  - **G11** deliverables exist: `post-v1-backlog.md` (12 parked pre-existing finds), `wiki/runbooks/v1-release-rebaseline.md` (Z-32), this close-out block.

**NOT merged to any integration branch — awaiting the operator's final sign-off, then the Sunday 2026-06-14 v1 release re-baseline (operator executes [`wiki/runbooks/v1-release-rebaseline.md`](../runbooks/v1-release-rebaseline.md) personally; that physical re-baseline closes the F-18 git-history residual permanently).** Environment notes unchanged: `.gitnexus/` cache breaks `git add -A` (staged explicit paths all session); C: SSD degraded writes → `go test -p 2`; vendored OTel deps added (`go.opentelemetry.io/otel@v1.44.0`, contrib `v0.69.0`) and committed under `vendor/`.

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

### Wave H — H-6b + H-PRE-1 close-out (2026-06-13)

- H-6b ✅: `DuplicateDocument` migrated to the atomic CD-create path; legacy `CreateDocument` chain deleted (`Service.CreateDocument`/`Repository.CreateDocument`/`Repository.SetRevisionStorageKey` + dead `ControlledDocumentReader`/`caps` + `service_caps_test.go`). `cloneIntoTx` is the sole seeding path. Net −406 LOC, contract-neutral.
- H-PRE-1 ✅ (distinct prereq, found during H-6b runtime-verify): system_admin CD-create self-deadlock on the audit hash-chain advisory lock `pg_advisory_xact_lock(90120260513004)`. The lock-holding atomic tx made a nested off-tx authz-recording taxonomy read (`GetByCode`) that re-grabbed the same lock on another connection. Fixed by hoisting `ensureTemplateArtifact` + template-version resolve off-tx before `runner.Do` (CD-create auto path + `CreateRevision`); `resolveTemplateVersionID` short-circuits on a concrete override. No authz/audit-model change.
- **Architectural rule learned:** never call an authz-recording read (one that records a system_admin bypass via `authz.Require`, which writes the audit hash-chain row under `pg_advisory_xact_lock(90120260513004)`) from inside a lock-holding atomic tx on a separate connection — it self-deadlocks undetectably. Keep such reads off-tx and pre-resolve their results.
- Runtime: 409 `template.artifact_missing` in 83ms (was 60–90s hang→500), 0 advisory locks before/after.
- Gates: build/vet/`test -p 2 ./...`/api-lint -strict 0/cilint 0. Review: sonnet APPROVE-with-nits (fixed).
- Defers (triggers in audit §H-PRE-1): happy-path clone runtime-exercise (seed gap); 409-before-403 ordering (ADR 0022 Phase 8); redundant bypass-audit (pre-existing cap-cache reset).
- One combined commit (H-6b titled, H-PRE-1 in body). NOT merged — operator review gate.
- Next Wave-H: **H-3b** → H-5 → H-2.
4. Do not re-execute anything from prior session summaries without verifying against the working tree — most of it is already done.

### Wave H — H-3b close-out (2026-06-13)

- H-3b ✅ (service-owned tx + in-tx audit for the 4 no-service-tx sites): Site 1 — new `SessionService` wraps `RevokeSessionTx` + `audit.RecordTx` in one `db.TxRunner.Do`; local `sessionRevoker` port avoids iam→auth-infra import cycle; rollback-on-audit-fail; 0-rows→`ErrSessionNotFound` preserved. Site 4 — `AdminService.UpsertUserAndAssignRole`/`ReplaceUserRoles` + new application-layer `roleAdminTxRepository` interface (domain never names `*sql.Tx` — REQ-FE-2/nosqltxindomain); `runner.Do` wraps repo + `RecordTx`; cache-invalidation post-commit. Site 3 — `PeopleService.WithTxAudit(runner, audit, updaterTx)` wraps Patch/Reset/Unlock/Invite with in-tx audit. Site 5 — auth `ChangePasswordForUser` + `AdminResetPassword` made atomic via hand-rolled `BeginTx`/`Commit` (auth has no injected TxRunner); new `updateUserTxRepository`/`revokeSessionsByUserIDTxRepository` ports; audit variadic on `NewService`; event `auth.password.changed`.
- New cross-function `DeliveryAuditSink` cilint analyzer: flags non-Tx audit/governance sink on audit-named receiver anywhere under `delivery/**`; registered in `analyzers.go` + table tests in `analyzers_test.go`.
- Review disposition: M1 ACCEPT (test-only else-branch — `sessionService` wired under same `SQLDB != nil` guard as handler registration, prod-unreachable; audit on a test-only path = needless complexity per CLAUDE.md 5.2). M2 FIXED (explicit `actorID` from `authn.UserIDFromContext` distinct from domain `assignedBy`; `buildAuditEvent` + `AdminService` methods take it; handler passes `authenticatedActor(r)`; runtime-proven `actor_id=admin`). M3 FIXED (`NewSessionsHandler` variadic tightened from `...interface{}` to `...auditdomain.Writer`; all 5 test sites still compile).
- Bounded DEFER: iam role repo seeds tier-2 authz principal from client-overridable `assigned_by` (pre-dates H-3b; fail-safe — tier-1 `CanDo(CapUserManage)` gates route + tier-2 re-checks, a forged `assigned_by` can only make tier-2 FAIL = fail-closed, no privilege escalation; post-M2 audit actor is the authenticated principal regardless; proper fix = decouple authz principal from `assigned_by` by threading `actorID` through `ReplaceUserRolesTx` — signature shared with `auth/application/service.go:128`/`:600` → cross-module authz change = CLAUDE.md hard-stop → DEFERRED; trigger: client-attributed `assigned_by` product feature, or auth/iam contract-consolidation wave; fix in ONE coordinated commit across iam-repo + iam-application-interface + iam-memory-repo + auth-interface(`service.go:128`) + auth-call(`:600`) + people_service Tx call, each passing its existing authenticated actor as the new `actorID`).
- Stage-3 analyzer repo-wide: 3 sites flagged, all suppressed with documented rationale, 0 unsuppressed, cilint exit 0 — (1) `auth/delivery/http/handler.go:200` login/logout best-effort [H-3b]; (2) `iam/delivery/http/people_handler.go:467` bulk-action envelope per-item already in-tx audited [H-3b]; (3) `controlleddocuments/application/service.go:412` pre-existing CD multi-leg-Create governance log, item-2.11 design defer [NOT H-3b, untouched].
- Gates green: `go build ./...` ok · `go vet ./...` ok · `go test -p 2 ./...` exit 0 (82 ok, 0 FAIL) · `api-lint -strict` 0 violations · `cilint` exit 0.
- Runtime: session revoke → 204; role upsert → 200 with in-tx audit `actor_id=admin`. Contract-neutral (raw-mux IAM/auth routes; no OpenAPI/codegen change). One H-3b family commit, NOT merged.
- Next Wave-H: H-5 → H-2.

### Wave H — H-5 close-out (2026-06-13)

- H-5.1 ✅: `RecordSignoff` kept **atomic** (async assessor: indivisible approval-tx complexity — do NOT fragment). Moved 6 raw-SQL read/resolve helpers (`loadPriorSignoffs`/`loadStageSignoffs`/`hasUnresolvedComments`/`loadActiveDocumentContentHash`+`scanSignoffs` from `decision_service.go`; `loadRoute`/`resolveEligibleActors` from `submit_service.go`) → `ApprovalRepository` (interface + postgres impl, all take `db.Tx`). SQL **verbatim** (adversarial reviewer char-confirmed vs HEAD). Repo sentinel `ErrNoActiveContentHash` maps to app `ErrContentHashMismatch` at the single call site (preserves both no-rows and null-hash). Deliberately kept inline (atomic + authz-ordering): display_name SELECT, the two RecordSignoff `UPDATE documents`, SubmitRevision `UPDATE documents`, `cancel_in_progress` GUC, all `authz.*`. Authz-ordering tripwire test end-anchor retargeted (deleted `loadRoute` → next-`\nfunc `); 3 ordering assertions preserved.
- H-5.2 ✅: documents `Handler.Service` 28-method fat interface split into 5 cohesive sub-interfaces (`documentReader`/`documentLifecycle`/`editingSessions`/`autosaveArtifacts`/`documentComments`) composed into `Service`; dropped 3 unused methods (`ListDocuments`/`ListDocumentsForUser`/`SyncArtifactMetadata` — zero `h.svc.` callers; audit's "4 unused" overcounted by 1) → 23 methods. Interface-only.
- H-5.3 partial: N+1 killed via batch `ListActiveForUsers`/`ListActiveBatch` (**D2**) + swallowed membership error now propagated (**D3**). **DEFER D1** (full SQL filter+paginate): `ListFiltered`'s username/email/lock filters read AUTH-owned `auth_identities` → self-contained IAM SQL projection is a cross-module hard-stop. Trigger: AUTH exposes `FilteredListUsers(ctx,tenantID,ListFilters)→(page,cursor,total,err)` port, OR `metaldocs.iam_users` gains `username`/`email`/`must_change_password`/`failed_login_attempts`/`locked_until`.
- Adversarial review (cavecrew-reviewer): 0 Critical, 2 Minor — both test-harness, accepted (tripwire end-anchor byte-scan functionally intact + documented + consistent with sibling tripwires; `fakeSubmitRepo` nil-embed pre-existing convention).
- Gates: build/vet/`test -p 2 ./...` 0 FAIL/api-lint -strict 0/cilint 0. Runtime (freshly-built binary, cookie session): login 200; `GET /api/v1/iam/users` 200 with batch memberships correct (`author-test=[rh]`, `approver-test`/`admin=[qualidade,rh]`); `?area_code=qualidade`→2 users, without_area=0; bogus area→200/0 (not 500 — new `ANY($2)` batch SQL valid + D3 doesn't false-trip). Contract-neutral. One H-5 family commit, NOT merged.
- Next Wave-H: **H-2** (composition/observability — the last Wave-H family), then the Wave H DONE gate (incl. A1 WS-upgrade 101 runtime proof).

### Wave H — H-2 close-out (2026-06-13) — FINAL Tier-2 family; Wave H Tier-1+Tier-2 complete

- H-2 ✅ (composition root + observability, C→A): **Pass 1** — 13 inline adapter types extracted from `apps/api/cmd/metaldocs-api/main.go` → `apps/api/internal/wiring/` (5 new files: `audit_adapters.go`, `documents_adapters.go`, `search_adapters.go`, `iam_adapters.go`, `clock.go`; private-struct+exported-ctor pattern beside `wiring/documents.go`). Plan named the 4th `taxonomy_adapters.go`; realized as `iam_adapters.go` — the taxonomy governance adapter already lived in wiring, the inline residue was the IAM membership + clock adapters. 9 bare `os.Getenv` → 4 typed loaders in `internal/platform/config/` (`server.go`=APP_PORT, `migration.go`=SKIP_STARTUP_MIGRATIONS+MIGRATIONS_DIR, `fanout.go`=FANOUT_URL+RENDERER_TOKEN, `retention.go`=AUDIT_RETENTION_DAYS), each +`_test.go`; `fanout`/`retention` fail loud on present-but-invalid (typo → error, not silent disable). `slog.SetDefault(JSON)` as first statement of all 3 binary mains (`log` import removed). Per-binary OTel `service.name`: `observability.SetupOTel` → 2-arg `(ctx, name)` (was hardcoded `metaldocs-api`) + `OTEL_SERVICE_NAME` override; `observability/http.go` dropped its private slog logger field → package `slog`. Worker graceful drain: in-flight batch finishes on SIGTERM via detached `context.WithTimeout(context.Background(), 30s)` before loop exit (bounded). Jobs cleanup-on-exit: fatal paths restructured so `defer deps.Cleanup()` always runs (os.Exit bypasses defers).
- **Pass 2** — main() decomposition: 5 order-safe subroutines extracted (`startPresence`, `buildFanoutComponents`, `startOutboxWorkers`, `registerScheduledJobs`, `startAuditRetention`) at exact prior call sites; interleaved documents/approval/fanout core left linear (load-bearing order); main() body 647→515 lines, file 1099→897 LOC. Load-bearing setters (`freezeService.WithMaterializeOutbox`, DecisionService wiring, `WithDocumentInitializer`, `WithRouteAdminIdempStore`), `var workerWG`, and the presenceHub drain capture kept inline in main.
- **Review disposition:** worker pre-batch racy `if ctx.Err() != nil { return }` guard removed (nondeterministically broke calls=1; the post-batch `select{<-ctx.Done()...}` already exits cleanly after the drain) + `main_test.go TestRunWorkerLoopStopsAfterCancellation` updated to assert the new detached-drain contract (batch ctx ≠ caller ctx + carries a bounded deadline); `TestRunWorkerBatchPassesCancellationContext` (RunOnce) unchanged. 2× `slog.Error(err.Error())` → structured `slog.Error(msg, "err", err)` (main.go 134/411). retention strict-parse behavior change (present+unparseable→error) intentional + self-documented — no edit.
- Gates: `go build ./...` 0 · `go vet ./...` 0 · `go test -p 2 ./...` 0 FAIL · `api-lint -strict` 0 · `cilint` exit 0. Adversarial review: faithful extraction, 0 Critical, 2 yellow both no-change.
- Runtime (`.\scripts\start-api.ps1 -Build`): all 3 binaries rebuilt + booted; API :8081 (migration 0224 applied → "MetalDocs API listening"); login 200; structured JSON logs throughout; worker PID 34124 alive; jobs PID 13816 alive; WS upgrade `/api/v1/iam/presence/stream` → **101 Switching Protocols** (residual A1 proof).
- Scope: 7 modified + 13 new files; no `api.gen.go`/`openapi.yaml` change → contract-neutral, no FE regen. Carried (not silent): D-H1 metrics ring-buffer NOT replaced (plan §H-2 NOT-to-do); no DI framework (manual wiring correct). One H-2 family commit, NOT merged — operator review gate.
- **Wave H status:** all Tier-1 (A1–A6) + Tier-2 (H-1..H-6) families ✅-or-deferred-with-trigger. Remaining: Wave H DONE-gate evidence consolidation + operator review. Deferred boundaries D-H1..D-H5 + H-5.3-D1 + H-3b tier-2-principal defer carry written triggers. NOT merged.

### v1 blocker remediation — close-out (2026-06-13) — branch `qa/iam-area-membership` → READY-for-v1 (operator review gate)

Plan: [`docs/superpowers/plans/2026-06-13-v1-blocker-remediation.md`](../../docs/superpowers/plans/2026-06-13-v1-blocker-remediation.md). 6 release blockers (B1–B6) + every Maximal-scope fix-now defer (H-A..H-G). One bounded commit per defect family. **NOT merged** — presented for operator review.

**Commit ledger** (`eae488e7c..HEAD`, 12 family commits; `a463ed375` baseline design doc precedes):

| Commit | Family | Subject |
|--------|--------|---------|
| `82420788a` | B1+B2 | fix(presence): snake_case wire keys + username |
| `c6ed8ceca` | B3 | fix(audit): FE consumes snake_case AuditEventItem |
| `7d5b98f3a` | B4 | fix(documents): form_data_json object not base64 |
| `b901c1567` | B5 | fix(documents): duplicate preserves visibility |
| `53b2495d8` | B6 | fix(search): escape LIKE/ILIKE wildcards |
| `3dec55631` | H-A | fix(audit): declare exportAuditEvents 202 status field |
| `c868c884b` | H-B | fix(observability): RFC9457 error bodies |
| `2307932ba` | H-C | fix(config): env guards |
| `f1ed57670` | H-D | docs(iam): reclassify phantom-perm row to TRIGGER (trackers only) |
| `b5c3b650c` | H-E | chore: remove dead code |
| `3216da809` | H-F | fix(metrics): declare p50/p95/p99 in MetricItem |
| `3e7fb9975` | H-G | docs(documents): reclassify profileDefaults status to TRIGGER (trackers only) |

**Per-commit gates:** every commit closed green — `go build ./...` 0 · `go vet ./...` 0 · `go test -p 2 ./...` 0 FAIL/panic · `api-lint -strict` 0 violations · `cilint` exit 0. Contract-changing commits (B4, H-A, H-F) carried FE regen (`npx openapi-typescript`) + `npx tsc --noEmit` 0 in the same commit. H-C (integration-tagged `e2e_seed.go`) also ran `go build -tags integration ./...` 0 + `go vet -tags integration ./internal/test/...` 0. Observable fixes runtime-verified by us against the running API (:8081, cookie session), never delegated:
- B1/B2 — presence stream wire keys snake_case + `username` present in admin overview (live JSON).
- B3 — audit list/export FE consumes snake_case `AuditEventItem` (live `GET /api/v1/audit/events`).
- B4 — `form_data_json` emitted as JSON object, not base64 (live document fetch).
- B5 — duplicated document preserves source visibility (live duplicate probe).
- B6 — `LIKE/ILIKE` wildcards escaped; `%`/`_` searched literally (live search probe).
- H-A — `exportAuditEvents` 202 body declares `status` field (live 202 export).
- H-B — error bodies are RFC9457 `application/problem+json` (live 4xx probes).
- H-F — `GET /api/v1/metrics` items all carry `p50/p95/p99_duration_ms` (present even at 0).

**Final gate (HEAD `3e7fb9975`):** full suite green incl. integration build/vet + FE tsc; `go generate ./...` → no codegen drift. **Cross-commit regression review (Opus):** APPROVE — 0 Critical, 0 Major over `git diff eae488e7c..HEAD` (36 files, +707/-328); 8/8 invariants PASS (codegen sync, snake_case wire casing, `required` correctness, RFC9457, advisory-lock deadlock untouched, two-tier authz untouched, cilint clean, surgical scope). 2 Minor, both accepted no-change: (1) presence `repository.go` Snapshot `INNER JOIN auth_identities` vs LEFT-JOIN precedent — proven unreachable (last_seen_at only set by authenticated requests, which require an auth_identities row), optional defensive nit; (2) the `admin_handler.go:228` 1-line `username` add is in B1/B2 not B3 — attribution note only, in-family benign. B6 `ESCAPE '\\'` flagged-then-confirmed FALSE POSITIVE (Go string literal → single backslash on the wire), not re-raised.

**TRIGGER register** (escalated out of fix-now scope — fixing each requires a boundary the plan forbids touching here):
- **F-001** (pre-existing) — phantom permission rows; trigger: IAM permission-catalog consolidation wave.
- **F-4a** (pre-existing) — see plan/roadmap; trigger as documented in roadmap.
- **concurrency** (pre-existing) — see plan/roadmap; trigger as documented in roadmap.
- **H-D** (contingent downgrade `f1ed57670`) — iam phantom-perm row was 🟠 fix-now → ⛔ TRIGGER: real fix needs cross-module permission-catalog change (CLAUDE.md hard-stop). Trackers-only commit, no code change.
- **H-G** (contingent downgrade `3e7fb9975`) — `profileDefaultsAdapter` hardcodes default-template `status:"published"`; real fix needs a template-version-status reader, `NewProfileDefaults` signature change + main.go wiring, AND collides with the H-PRE-1 advisory-lock deadlock constraint (status read must stay off the lock-holding atomic CD-create tx). 🟠 fix-now → ⛔ TRIGGER. Trackers-only commit, no code change. Trigger: `DocumentProfile`/taxonomy gains a default-template status field, or a status-aware template reader is threaded off-tx.

**KEEP register** (7 — verified correct as-is, no action):
1. B6 `ESCAPE '\\'` — correct (false-positive in review; single backslash on the wire).
2. Presence Snapshot `INNER JOIN auth_identities` — behavior identical to LEFT JOIN on every reachable path; not changed (surgical scope).
3. `admin_handler.go` `username` add — correct in-family B1/B2 enrichment, not B3 drift.
4. H-F metrics — Go codegen empirically neutral (tag-filtered embedded-spec); only FE types regen; handler already emitted percentiles, schema-only declaration.
5. H-C env guards — `fanout`/`retention` fail-loud-on-invalid behavior intentional + self-documented.
6. H-B RFC9457 — `application/problem+json` content-type + member shape conform; no over-broadening.
7. Advisory-lock deadlock constraint (H-PRE-1) — off-tx default-template fetch in documents service left untouched; H-G respects it.

**Result:** branch READY-for-v1 pending operator review. No merge performed (operator review gate). Resume point: operator approves → standard merge of `qa/iam-area-membership`; the two TRIGGER downgrades (H-D, H-G) carry written triggers and do not block v1.
