# Dev-Velocity & Release Program

**Status:** AGREED v3 (Claude draft → Codex GPT-5.6 critique → reconciliation → Codex round-3 ACK 5/6 + 2 objections adopted) — **pending operator ratification only**. No unit executes before ratification.
**Date:** 2026-07-13
**Execution model:** each unit = a harness unit (`docs/superpowers/HARNESS.md` P0–P7); this plan's hub session dispatches per Transport B default. Wave R units are release-infrastructure — **single owner (hub), strictly serialized** (they share compose, scripts, env, runbooks). Wave V runs post-release.

**Goals (operator-stated):**
1. Release MetalDocs v1 THIS WEEK.
2. Solo developer ≥3x efficiency after release.
3. Architecture/coding standards enforced structurally, permanently.

---

## 0. Wave 0 — already landed (2026-07-13, NOT pushed)

| Commit | What |
|---|---|
| d5b450e5 | ops/smoke rewritten to real health routes; fictional approval-v2 roundtrip archived; smoke.yml → manual dispatch |
| 71315407 | ops/incident/RUNBOOK.md rewritten for compose reality; fictional freeze.sh archived |
| 2a12e7ed | fe-ci.yml blocking tsc+vitest gate; web `typecheck` script |

## 1. Evidence base (7-agent audit, 2026-07-13)

- **CI:** 19 workflows; strong custom lint wall (cilint ×9, api-lint ~20 blocking, module-boundaries, tripwire parity/drift, test-discipline). No git hooks. ~20 manual-only scripts. Integration tests not in CI. `make test` = FE only.
- **Process:** verification 6–8× per unit; 200 evidence.md files / 16,773 hand-typed lines. Biggest velocity cost.
- **Tests:** 512 files, 148 integration-tagged, zero `t.Parallel()` in integration; `newPermissiveMockDB` ×5 copies; `tests/docx_v2` 8/9 skipped; ADR 0034 describes nonexistent "IntegreSQL".
- **Deploy:** compose single host, `:dev` tags, no version stamping (stale-image false-greens ×2); worker/jobs no health surface (verified: zero HTTP listeners in either binary — only api owns METRICS_ADDR); migrations auto-apply at api boot (advisory-locked, exit 1 on failure).
- **Drift:** hand-synced enums (`subject_kind` ×3 sites, doc-status parity pin, notification status); CLAUDE.md "14 modules" (=15).
- **FE:** duplicate-React root cause = packages pin react 18.2.0 vs web 18.3.1 + raw-source alias + no dedupe + junction drift. Affects LOCAL dev tree; release image built in Docker (fresh install) — likely unaffected (verify in V0).
- **Verified for this plan:** merge commits e6e54813/bc8c351f/72f8bd5c/dc497fcf/a9afbeca (units 2.1–2.5 MERGED; ROADMAP rows stale). Health routes ARE in OpenAPI (`HealthResponse`). `db/dev-seeds/0001_local_dev_seed.sql` exists. Backup/restore scripts + `run-backup-restore-gate.ps1` exist.

## 2. Codex reconciliation record (round 2)

Codex verdicts: all units CHANGE (specifics adopted below), R7 DROP, 4 ADD units (R0/R0a/R0b/R0c).

**Adopted:** timestamp-freshness invalid → SHA identity (R3) sequenced before smoke freshness (R2); rehearsal isolation requires dedicated compose override (fixed `container_name`s + hard-coded host ports incl. MinIO `127.0.0.1:9000` collide under any `-p`); one host-validated SHA threaded release-script → compose `build.args` → `ARG` → ldflags + OCI `org.opencontainers.image.revision` on ALL images incl. jobs; `/health/live` liveness semantics unchanged (additive version exposure only); rollback = restore-from-backup + previous images, never SQL downgrade; async acceptance (real outbox/River roundtrip) replaces container healthchecks as the worker/jobs release gate; R5 gated on reproduction in release image; V1 reshaped (kill only duplicate command reruns; deterministic path-based risk classification; merge-SHA post-merge checks; 3-class representative trial); V5 inventory-first; V8 delete-after-mapping; secret rotation verification precedes any F-18 timing decision.

**Modified (Claude position, pending Codex ACK):**
1. R0/R0a/R0b/R0c fold into R1/R6/R2 as **named, non-droppable checklist items** instead of standalone units — solo-dev dispatch economy; content identical.
2. "Freeze all roadmap work during R2–R6" → **merge-freeze** (product work continues on worktree branches; main accepts only release-blocker merges during the release window). Operator decides (D5) — hard constraint from operator: development cannot stop.
3. Version exposure: additive optional `version` field in `HealthResponse` schema (contract-first, regen, additive = no consumer break) chosen over `X-MetalDocs-Build-SHA` header — field is typed + discoverable; header noted as fallback if regen ripple proves large.

## 3. Wave R — release week (serialized: R1 → R3 → R2 → R4 → R6)

### R1. Release freeze + gate definition (same-day gate, FIRST)
**Goal:** one signed decision: what v1 IS.
- Release-freeze checklist (absorbs Codex R0): record release SHA (clean tree asserted); lockfile/dependency freeze; deployment `.env` key INVENTORY (names only — never values); exact production compose command documented; **secret-rotation verification** — confirm the F-18 historical secret is revoked/rotated (D2 prerequisite; check without printing values).
- Blocker decision list: consolidated L3 browser QA (ROADMAP §5.6) and F2d.8 walkthrough = candidate blockers; 3.2 / 4.1–4.4 / Wave V = post-release (D1).
- ROADMAP reconciliation: rows 2.1–2.5 → merged status (commits above); 1.1/1.4 approved; dependencies/owners re-checked, not just status flags. Edits applied only after operator OK.
- **Merge-freeze proposal (D5):** during R-window, main accepts only release-blocker merges; product units keep working on worktree branches.
**Deliverable:** `docs/superpowers/reports/2026-07-XX-release-freeze.md` + ROADMAP row edits (post-OK).
**Context:** ROADMAP.md, git log, memory index. **Budget:** ≤40k. **Collisions:** ROADMAP.md (hub-owned).

### R2. Build identity — git-SHA end to end (before smoke; kills false-green class)
- `internal/platform/version` pkg: `var GitSHA = "dev"`, `var BuildTime = "unknown"` (ldflags `-X` targets; defaults tested).
- ONE host-validated value: release script computes `git rev-parse HEAD` (clean-tree assert) → compose `build.args: GIT_SHA` → each Dockerfile `ARG GIT_SHA` → Go ldflags + OCI label `org.opencontainers.image.revision`. ALL release images: api, worker, jobs, docx-renderer, web (web: build-time env → `version.json` served by nginx or baked env echo). Dockerfiles never run `git` themselves.
- Expose: additive optional `version` in `HealthResponse` (OpenAPI edit → oapi-codegen regen → handler reads version pkg); worker/jobs log SHA at boot; FE regenerated types (additive).
- `make up` stays dev-loose but gains `--build`; RELEASE path (release-smoke) separately asserts clean tree + explicit SHA — dev convenience ≠ release provenance.
**Acceptance:** stale image vs HEAD → smoke FAILS naming both SHAs; rebuild → green. OCI label inspectable on all 5 images.
**Context:** 3 Go Dockerfiles + web/docx Dockerfiles, compose yml, `api/openapi/v1/openapi.yaml` + regen chain, health handler (`internal/platform/observability/health.go`), Makefile. **Budget:** ≤150k. **Collisions:** OpenAPI regen touches generated FE types — dispatch only under R1 merge-freeze; serialize with any contract work.

### R3 → renumbered: R2 above is identity; R3 below is smoke. (Order: identity first — Codex.)

### R3. `scripts/release-smoke.ps1` — the codified L2 gate
Stages (script + `ops/DEPLOY.md` cross-reference; each stage logs to `logs/release-smoke-<ts>/`):
1. Preflight: `.env` present (names-only check), docker daemon up, clean tree + SHA recorded.
2. Backup (skippable `-SkipBackup` cold-stack): `scripts/backup-postgres.ps1`.
3. Build: `docker compose --env-file ../../.env build --progress plain 2>&1 | tee ...` with `GIT_SHA` arg.
4. Up + health-wait: compose health states (postgres 600s start_period honored) + gateway `GET :80/api/v1/health/live` + `/ready` (`status=="ready"`).
5. **Identity check:** `/health/live` `version` == release SHA; OCI labels of running containers == release SHA. FAIL on mismatch (no timestamp heuristics).
6. App-level: `check-system-runnable.ps1` gains `-BaseUrl` param (default unchanged `127.0.0.1:8081`; release mode `http://localhost`).
7. Gateway routes: web root 200; login roundtrip; `auth/me`; one governed API route; RFC 9457 problem+json shape on deliberate 401.
8. **Async acceptance (absorbs Codex R0c):** trigger one real outbox/River-relevant action via API, poll for its observable completion (worker consumed / job ran); container "running" is NOT evidence.
9. **Smoke credentials (absorbs R0b):** dev/rehearsal = documented seeded personas; production mode = operator-provided credential via prompt/env-var reference at runtime, never logged, never stored in repo. No bootstrap-admin secrets in output.
10. Any failure: `compose ps` + last 200 log lines of failing service; non-zero exit.
**Acceptance:** green run artifact; deliberately-broken run (api stopped) shows diagnostics + non-zero; deliberately-stale image fails stage 5.
**Budget:** ≤120k. **Collisions:** scripts-only + check-system-runnable param (additive).

### R4. Release procedure doc — backup → migrate → rollout → smoke
`ops/DEPLOY.md` binding order: backup → build (SHA-stamped) → `up -d` (api applies boot migrations advisory-locked) → full stack → release-smoke. **Rollback policy (Codex-corrected): forward-only migrations; rollback = restore data from backup + redeploy previous SHA images; NEVER SQL downgrade; RPO = last backup, RTO = restore-gate-measured.** Migration-compatibility note: schema changes shipping in a release must be expand-only relative to the previous release's code (documented rule, checked at R1 freeze).
**Budget:** ≤40k.

### R6. Release rehearsal — cold start to GO/NO-GO (absorbs Codex R0a)
1. **Dedicated `deploy/compose/docker-compose.rehearsal.yml` override** (Codex-corrected isolation): NO `container_name` entries; every published host port parameterized (gateway, postgres, redis, minio ×2, gotenberg, docx, api); rehearsal-scoped named volumes; distinct `COMPOSE_PROJECT_NAME`; distinct DB name, MinIO bucket, session-cookie NAME, session secret (rehearsal-generated), `METALDOCS_MINIO_PUBLIC_ENDPOINT` port-shifted. Pre-cleanup assert: every targeted resource carries the rehearsal project label before any `down -v`; rehearsal commands verify `COMPOSE_PROJECT_NAME` first.
2. Cold boot: empty volumes → initdb baseline (prerequisites→baseline→reference-data) → api boot migrations → minio-init blank template → all healthy. WSL2 realities budgeted: cold build ~10min, PG recovery slow paths; disk headroom checked first.
3. Seed personas (`db/dev-seeds/0001_local_dev_seed.sql`).
4. `release-smoke.ps1` against rehearsal gateway port (incl. identity + async stages).
5. **Migration + restore drill (R0a):** `run-backup-restore-gate.ps1` round-trip INTO the rehearsal stack; proves the R4 rollback policy executable; RPO/RTO recorded.
6. Browser QA limited to v1 CRITICAL journeys (Codex cut): consolidated L3 list — F2d.8 walkthrough, route builder v2, rebuilt template approval — per HARNESS §6 persona or operator session. Full exploratory QA post-GO.
7. Output: `docs/superpowers/reports/2026-07-XX-release-rehearsal.md` — GO/NO-GO per journey.
**Depends:** R2, R3, R4. **Budget:** ≤150k (+wall-clock). **Collisions:** new override file only; shares docker daemon (cache = feature).

### Dropped from Wave R
- **R7 worker/jobs healthchecks → V10.** Codex: healthcheck ≠ correctness; R3 stage 8 (async acceptance) is the release gate. V10 executes only on named trigger (async incident or R3-stage-8 flakiness).
- **R5 duplicate-React → V0.** Release image built in Docker from lockfile (no junctions); V0 first REPRODUCES in release image — if a v1 journey breaks, it re-enters release scope; else post-release velocity fix.

## 4. Wave V — velocity (post-release)

### V0. Duplicate-React + FE tree heal (was R5)
Reproduce first (release image + local). Fix: correct dependency OWNERSHIP (react as peerDependency in packages/editor-ui, eigenpal-adapter, form-ui — not blind bumps), align on ^18.3.1, `resolve.dedupe: ['react','react-dom']` in web vite config, full root `pnpm install` (heals junction drift per memory), verify vitest/tsc/build/docker-image. One commit; FE-quiet window; lockfile serialized with any FE unit. **Budget:** ≤100k.

### V1. Verification-ritual collapse (biggest lever; Codex-reshaped)
Principle: **eliminate duplicate command executions, never independent judgment.**
- Slice review: combined single pass for LOW-RISK slices; **deterministic risk classification** — task card auto-flags full two-stage rigor when touched paths match: `internal/modules/*/domain|application` authz/tenancy surfaces, `api/openapi/**`, `db/migrations/**`, outbox producers/consumers, middleware chain, tripwire arms. Path-rule list versioned in HARNESS.md — no judgment calls, no under-declaring.
- Hub acceptance: evidence audit + spot L0; full re-review only on flags (risk paths, 2×-reject, >500 lines — flags are OR, so small authz diffs still escalate via path rule).
- Post-merge: checks run at the MERGE SHA (not the branch SHA); L1 scoped to touched packages + their transitive reverse-DEPENDENTS (importer index built from `go list -json ./...` — Codex round-3 correction: `go list -deps` is forward-direction and wrong for "who is affected") + integration tags touching those packages; full `go test ./...` at milestone close.
- Milestone validator: keeps C1–C7 judgment; drops only command re-executions already evidenced at merge SHA; validates aggregate + samples.
- **Trial:** 3 representative unit classes (module-local / contract-crossing / tenancy-authz-outbox); compare escaped defects, review rejects, elapsed time vs baseline; operator ratifies permanence after trial (D4).
- Browser/product QA never reduced by code-check greenness.
**Deliverable:** HARNESS.md amendment diff + path-rule table + trial protocol. **Budget:** ≤100k (design).

### V2. Task facade — `scripts/task.ps1 <verb>`
Thin wrapper, stable exit codes, machine-readable `-Json` where cheap; wraps existing scripts (up/down/logs/api/test-go/test-int/test-fe/typecheck/lint/gen/seed/backup/restore/smoke/runnable/rehearse; `task help`). Incremental — NOT a second orchestration system (Codex guard). **Budget:** ≤80k.

### V3. Integration parallelism — folded into ROADMAP 4.3
Amendment to 4.3: benchmark before; after clone isolation, `t.Parallel()` package-by-package (3 slowest first); repeat-run flake thresholds + rollback criteria per package; re-benchmark; target ≥2x; template-clone contention = known ceiling. **+100k onto 4.3.**

### V4. Integration tests in CI
Real bootstrap path (prerequisites→baseline→reference-data SQL into postgres:16 service container) + `test-integration` equivalent. Start nightly + PR-label (`ci:integration`) + push-main; publish duration/flake metrics; main-protection decision AFTER V3 lands and metrics stabilize. **Budget:** ≤120k.

### V5. Canonical enum generation (after 3.2)
Slice 1 = consumer INVENTORY per enum (subject_kind, doc-status parity pin, notification status) + canonical-owner decision each; slice 2 = generator (GMR M2 pattern: Go registry → generated artifacts) + blocking drift lints. **Budget:** ≤200k. Hard-sequenced after 3.2.

### V6. Pre-commit hooks (<30s)
Staged-file tools only: gofmt/goimports, `go vet` touched packages, eslint invariant rule, gitleaks protect. Bootstrap-installed, documented emergency bypass for the human operator (agents: CLAUDE.md never-skip stays). **Budget:** ≤60k.

### V7. Shared authz mock extraction
Define narrow test-support API first; migrate package-by-package (permissive mock must not conceal authz behavior — each migration re-runs that package's suite). **Budget:** ≤60k.

### V8. ADR 0034 truth repair + docx_v2 mapping
ADR 0034 → real testdb template-clone factory NOW. docx_v2: map each skipped test to obsolete-behavior or replacement-coverage; delete only mapped-obsolete (D3). **Budget:** ≤60k.

### V9. Evidence auto-capture
Captured per command: command line, commit SHA, clean/dirty state, exit code, elapsed, artifact paths (tee convention + template). Templates alone ≠ trust — capture is the point. Feeds V1. **Budget:** ≤40k.

### V10. worker/jobs health surface (trigger-gated)
Only on named trigger (async incident / smoke stage-8 flakiness): minimal `/healthz` listener per binary (api METRICS_ADDR dedicated-listener pattern, infra-net only) + compose healthchecks. **Budget:** ≤80k.

**V order:** V1 → V2 → V9 → V0 → V6 → V8 → V7 → V4 → [4.3+V3] → V5 (after 3.2). V10 trigger-gated.

## 5. Ownership & serialization matrix (Codex-upgraded from surface list)

| Unit | Exact surfaces | Owner | Serialization rule |
|---|---|---|---|
| R1 | ROADMAP.md, reports/ | hub | edits post-operator-OK only |
| R2 | Dockerfiles ×5, compose yml, openapi.yaml + api.gen + FE generated types, health handler, Makefile | hub-dispatched unit | merge-freeze active; serialize vs ANY contract-touching unit |
| R3 | scripts/ (new + check-system-runnable param) | same release-infra owner | after R2 |
| R4 | ops/DEPLOY.md, RUNBOOK pointer | same | after R3 |
| R6 | new rehearsal override yml, reports/ | same | last; needs R2–R4 |
| V0 | pnpm-lock, 3 packages, web vite config | FE-quiet window | serialize vs 4.1 + any FE unit |
| V1/V2/V6/V9 | HARNESS.md, scripts, hooks — the operating machine | hub + operator HS-1 | cutover point declared; already-dispatched units finish on old rules |
| V3/V4 | test infra + CI | folded/sequenced with 4.3 | V4 consumes V3's finalized design |
| V5 | kernel enums + all consumers (inventory first) | after 3.2 | inventory gates sequencing |
| V7 | 5+ test packages | package-by-package | not while a package's tests under active modification |
| V8 | ADR 0034, tests/docx_v2 | mapping gates deletion | — |

## 6. NOT-DOING register (binding)
Coverage floor · ROADMAP CI lint (manual freeze-report instead) · defer/stamp automation · weekly autonomous drift-audit agent · GHCR/registry pipeline (revisit at multi-host/second-operator) · secrets-management redesign (rotation check in R1 + F-18 re-baseline cover it) · K8s.

## 7. Operator decisions (ratification gate)

| # | Decision | Recommendation |
|---|---|---|
| D1 | v1 release gate contents | Ship on main + Wave R green; 3.2/4.x post-release |
| D2 | F-18 secret + re-baseline | Verify rotation NOW (R1 checklist — rotation IS the release blocker). Re-baseline strictly POST-release (Codex round-3: re-baselining after R6 changes the shipped SHA vs the rehearsed SHA — never inside release week; before-R1 only if operator declares it release-blocking) |
| D3 | docx_v2 skipped tests | Map then delete obsolete (V8) |
| D4 | V1 trust-boundary trial | Approve 3-class trial post-release; permanence only after trial metrics |
| D5 | Release-window freeze | Merge-freeze on main during R2–R6 (product work continues on branches) |
| D6 | Wave R order + drops | R1→R2(identity)→R3(smoke)→R4→R6; R7 dropped to V10; R5 deferred to V0 |

## 8. Week shape (solo-dev realism, post-Codex cuts)

Day 1: R1 (same-day freeze gate; operator answers D1–D6). Day 1–2: R2. Day 2–3: R3 + R4. Day 3–4: R6 rehearsal + fix loop. Day 5: buffer / GO decision. F-18 re-baseline: NOT in release week (post-release; rotation already verified in R1). Product work continues on branches throughout (merge-freeze per D5). Wave V starts after GO.

## 9. Agreement record

Round 1 (pre-plan): joint improvement list. Round 2: Codex full critique — all units CHANGE with specifics (adopted), R7 DROP, R0/R0a/R0b/R0c ADD (folded as named checklist items). Round 3: 6 deltas — ACK ×5; OBJECT ×2 both adopted (reverse-dependents direction; F-18 out of release week). Claude + Codex AGREED. Pending: operator ratification of D1–D6.
