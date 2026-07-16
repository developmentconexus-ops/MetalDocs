# Harness Profile — MetalDocs

**Layer:** REPO (binds with `HARNESS-CORE.md` from the `mnfs-harness` plugin, v0.2.0).
Conflict order: this profile wins over core for in-flight missions (core §0).
Adopted 2026-07-16 at a clean unit boundary (unit 4.5 merged `3a2f9f1a`, nothing in flight),
superseding the combined `docs/superpowers/HARNESS.md` (now a pointer + method addenda pending
core upstream). The active queue for the current program is `docs/superpowers/ROADMAP.md`
(legacy mission layer — no `.mnfs/` yet; a future mission planned via `/mission-init` moves the
queue to `.mnfs/MIS-*/`).

Every section: `status: ratified | assumed | open` + provenance. Amendments per core §0 —
append-only log at the bottom.

---

## 1. Identity & stack
`status: ratified` · `provenance: 2026-07-16 · migrated from HARNESS.md + CLAUDE.md (program-proven since 2026-07-10)`

- **Languages/build:** Go modular monolith (4 binaries: `metaldocs-api`, `metaldocs-worker`,
  `metaldocs-jobs`, `docx-renderer`; 15 bounded-context modules under `internal/modules/`) +
  TypeScript frontend (`frontend/apps/web`, pnpm) + docx workspace (`npm run *:docx-v2`).
- **Contract-first:** routes change ONLY via `api/openapi` + `oapi-codegen`.
- **OS/shell binding:** Windows 11 + PowerShell for ALL stack ops and scripts. bash /
  `source .env` FORBIDDEN. Field note 2026-07-16: the Claude Bash tool can break machine-wide
  (shell-profile parse error `unexpected EOF while looking for matching`) — PowerShell tool is
  the reliable path; scripts authored as `.ps1`.
- **Default branch:** `main`. Hub checkout must be on `main` (never detached HEAD / chip
  scaffold branch) before EVERY hub commit — chip launch can silently switch the hub working
  dir (field gotcha 2026-07-14).
- **Queue (mission layer, current program):** `docs/superpowers/ROADMAP.md` — ordered unit
  table with per-unit context files + token budgets. First act of program work: read it;
  last act: update the unit row.

## 2. Verification ladder bindings (core §5)
`status: ratified` · `provenance: 2026-07-16 · migrated from HARNESS.md §5 (field-proven across GMR M0–M9 + units 2.x–4.6)`

- **L0** — `go build ./...` · `npm run typecheck:docx-v2` (when docx touched) · FE typecheck
  (when FE touched) · governance lanes: api-lint, `module-boundaries.yml`,
  check-test-discipline, capability-coherence. Zero findings.
- **L1** — `go test ./...` (unit) · integration via `.\scripts\test-integration.ps1` — NEVER
  hand-set `DATABASE_URL` (script derives from `.env` `POSTGRES_*`, probes postgres, fails
  loud; `testdb.Open` silently SKIPS when the var is missing/DB down = false green) · FE
  vitest via `make test`.
  **Selective policy (ratified 2026-07-13):** default = touched packages' integration tests +
  guard suites (`./tests/integration/tenantdata/...`, `./tests/integration/scenarios/...`,
  `./tests/integration/iam/...`); FULL `./...` integration only when `db/migrations`,
  `db/baseline`, or `internal/platform/**` touched. No blanket `-count=1` (Go test cache is
  sound — testdb schema fingerprint reads every migration file); `-count=1` only to re-prove
  a specific flake. Full-suite reference: 115 ok / 39 no-test, ~8.5–13 min wall at `-p 8`
  (post-4.6 leased-DB factory; TRUE-green baseline pinned `788b140a`, 2026-07-16).
- **L2** — full container stack via gateway `:80`, coded compose path
  (`build --progress plain` + tee logfile, or `dev-api.ps1`; never silent-background builds or
  short timeouts) · gateway smoke: logins, target routes, RFC 9457 `problem+json` shapes.
  Start scripts: `.\scripts\start-api.ps1` (`-Build` to rebuild),
  `.\scripts\check-system-runnable.ps1`.
- **L3/L4** — QA persona per core; QA state and identities are REPO KNOWLEDGE:
  `wiki/references/local-dev-startup.md` (seeded personas table — dev-only, documented,
  NOT secrets; scripted login with seeded creds is REQUIRED self-sufficiency) +
  `wiki/quality/qa-operating-system.md` + surface checklist. Curl-only = FAIL (M2c field
  lesson). Asking the operator to type a wiki-documented seed password = QA failure.

## 3. Fresh-workspace bootstrap
`status: ratified` · `provenance: 2026-07-16 · field findings 2026-07-12..16 (units 4.2–4.5, Transport A probes)`

Before any lane in a fresh worktree/chip:

1. `.claude/settings.json` must have `worktree.baseRef: "head"` — default "fresh" forks from
   `origin/main`, which is ~20+ commits stale here.
2. Copy `.env` (+`.env.local` if present) from the main checkout into the worktree — file copy
   only, contents never read/printed. Missing `.env` = `REQUEST env`, never hand-set
   connection strings.
3. `node_modules` is SYMLINKED from the main checkout — NEVER run `pnpm install`/`npm install`
   inside a worktree (dep change = `REQUEST`).
4. **Worktree cwd drift trap:** shell-tool cwd silently resets to the main repo between calls —
   `Set-Location` to the worktree in EVERY path-relative call, or suites run against wrong code
   (field: unit 4.5 run-1 invalidated by exactly this).

False-alarm signatures (check BEFORE debugging a "failure"):
- Integration tests all "pass" instantly / skip → `DATABASE_URL` not derived (see L1) — false
  green, not success.
- FE vitest/vite breaks with module resolution errors → pnpm nested-junction drift in
  `frontend/apps/web`; real fix = complete `pnpm install` in the MAIN checkout (operator gate).
- Suite results contradict the diff → cwd drift (item 4).

## 4. Test database / integration strategy
`status: ratified` · `provenance: 2026-07-16 · units 4.6 (leased-DB factory) + 4.5 (GC design defect, fix A 572f6827) + parallel-track field findings 2026-07-14`

- **Infra:** postgres:16, Docker Desktop/WSL2, `:5433`, `POSTGRES_USER=metaldocs_app`,
  `POSTGRES_DB=metaldocs`, `synchronous_commit=off` (dev). Note: dev role is
  superuser+BYPASSRLS → RLS inert in dev; RLS-truth tests use the non-owner `metaldocs_ci`
  role.
- **Isolation unit:** content-addressed TEMPLATE DB per schema fingerprint (fingerprint reads
  every migration file) + leased per-test clone DBs (`metaldocs_test%`), post-4.6 factory.
- **GC contract (fix A, 4.5):** `GCRetiredDatabases` NEVER runs mid-suite or on a hot cluster —
  its wall time scales with cluster debris and `DROP DATABASE WITH (FORCE)` wedges minutes on
  WSL2 checkpoints. Guard tests use the pure `classifyGCCandidate`; the e2e full-GC pass runs
  ONLY behind `TESTDB_GC_E2E=1` on an idle cluster. Budget bumps to absorb GC hangs are
  local-max patching = defect.
- **Cross-track serialization (core §3 rule instantiated):** divergent-fingerprint tracks on
  the shared `:5433` mutually delete templates → `3D000`/timeouts. Hub SERIALIZES cross-track
  integration runs when migrations diverge; same-fingerprint concurrent runs are safe.
- **Orphan sweep (R1 policy):** hub sweeps orphan `metaldocs_test%` DBs (unmarked or
  `retired:`-marked) ONLY when no track is running (per-DB `pg_stat_activity` idle check,
  fail-loud) — before every full-suite window, never mid-run (DROP storm hangs live tracks).
  Script pattern: scratchpad `sweep-orphans.ps1` via `docker exec metaldocs-postgres psql`.
- **First-boot race:** cold template rebuild after a fingerprint rotation is expensive —
  don't misread the first run's clone-gate waits as hangs; TRUE-green reference times in §2.

## 5. Collision axes — instantiation (core §3)
`status: ratified` · `provenance: 2026-07-16 · migrated from HARNESS.md §9.1 (proven across Track A/B/C parallel units 4.2–4.4)`

| Axis | Concrete binding in this repo |
|---|---|
| Contract artifacts | `api/openapi/**` + oapi-codegen output (`*.gen.go`); one owner at a time (contract lock) |
| FE surface | `frontend/apps/web/src/features/<domain>` component trees; router/nav/layout = named owned seams |
| Migration | `db/migrations/` 4-digit sequential; hub pre-allocates disjoint blocks per track in chip prompts; unplanned need = `REQUEST migration-number` |
| DB shape | per-table exclusive ownership; tenant-port registrations (`TenantDataPort`) = shared file, one named owner |
| Module | `internal/modules/<name>` internals; cross-module edges only via application service / published Go interface (module-boundaries lint enforces) |

## 6. Shared seams & owners
`status: ratified` · `provenance: 2026-07-16 · migrated from HARNESS.md §2/§9.2`

Hub-owned (chips `REQUEST`, never take):
- `:80` container stack (rebuild/restart/reseed) and `:5433` shared postgres windows
  (full-suite scheduling + orphan sweeps per §4).
- Contract lock on `api/openapi`.
- Migration number blocks; `db/baseline` folds + `scripts/check-baseline-equivalence.ps1` gate.
- `docs/superpowers/ROADMAP.md` (queue rows) and this profile (amendments).
- Merge-back: ONE acceptance gate, serialized, smallest/lowest-risk first; post-merge ladder on
  integrated `main`.
- `HUB_SESSION_ID` resolution rule: the `local_…` id from a message this hub session previously
  SENT (never the scratchpad/transcript UUID); title-match fallback embedded in chip prompts.

## 7. Non-negotiables (per-endpoint / per-write, core §5)
`status: ratified` · `provenance: 2026-07-16 · migrated from HARNESS.md §5 + CLAUDE.md invariants (ADR-anchored)`

Re-checked at L0–L2 for every touched endpoint (miss-nothing list):
- AuthZ = capabilities, never roles (ADR 0022): tier-1 route→capability wired · tier-2
  `authz.Require` in-tx after `SeedTxIdentity` (needs a WRITABLE tx — G1 ruling) · DB tripwire
  arm exists (generated from Go registry, drift/parity lints).
- Tenant predicate on every query; cross-tenant URL → 404; tx-local GUCs only;
  tenant-namespaced blob keys.
- RFC 9457 `problem+json` errors; fixed middleware chain (`chain.go`) — new routes never
  reinvent auth/validation/errors.
- Async = transactional outbox; consumers idempotent; idempotency per-handler where the
  contract requires it.
- H-PRE-1: no authz-recording read inside a lock-holding atomic tx.
- OpenAPI ↔ generated ↔ handler parity; generated-file edits forbidden.
- No-fallback principle: integrity-critical reads fail closed (unknown ≠ zero); legacy-fallback
  extermination — migrating a consumer DROPs old wire/DTO fields clean, never relax-to-optional.
- Actor ids are TEXT (`iam_users.user_id` PK) — never type actor columns uuid.
- Test framework hard gate: new tests use the canonical framework for their class (testdb
  factory for DB integration); drive-by repair only for pre-existing.
- DB enforces invariants (triggers/constraints); app checks are the friendly first line.

## 8. Truth order (core §6)
`status: ratified` · `provenance: 2026-07-16 · CLAUDE.md "runtime truth beats docs" + backend-target-architecture governance (program-proven)`

On conflict, higher wins; stop-and-classify against this list (outside current boundary = STOP):
1. Runtime behavior (live system).
2. `api/openapi` contract (route truth).
3. Generated code.
4. `wiki/architecture/backend-target-architecture.md` (REQ IDs) + ADRs (`wiki/decisions/`).
5. Wiki module/architecture docs.
6. Memory / evidence files.
Tests assert truth, they don't define it — legacy one-off tests are deletable
(legacy-test-deletion rule); contract/invariant guard tests are repaired, not deleted.

## 9. Human gates
`status: ratified` · `provenance: 2026-07-16 · standing operator rulings 2026-07-10..16`

Operator-only, never assumed:
- **Push** — never without explicit permission per push; prior authorization is consumed by use.
- HS-1 gates, spec/model ratifications (AskUserQuestion), pause/resume dispatch.
- Chip launches (Transport B default — operator launches on Opus, approves each
  cross-session send; client-enforced, never rerouted through side channels).
- Dependency changes (`pnpm install`, go.mod additions).
- Interactive `/codex:setup` (codex sandbox broken machine-wide until run).
- `settings.local.json` secret hygiene (user-only).
- Fable is never a worker; Claude worker models: sonnet implement/review fallback, haiku
  mechanical, cold Opus dual-gate; ≤15 concurrent workers.

## 10. Superseded protocols denylist
`status: ratified` · `provenance: 2026-07-16 · supersede events dated inline`

Chip prompts pin this list verbatim (core §2 item h). Never rely on on-disk skill discovery in
worktrees.

| Retired | Since | Superseded by |
|---|---|---|
| Combined `docs/superpowers/HARNESS.md` as binding doctrine | 2026-07-16 | HARNESS-CORE (plugin 0.2.0) + this profile |
| Transport A (background unit agents) as default | 2026-07-13 | Transport B spawn_task chips (A = explicit operator opt-in per unit) |
| MNFS execution engine (`/milestone-start`, `/feature-context`, `/feature-accept`, milestone-orchestrator/feature-implementer/correction-worker agents) | 2026-07-15 | harness chips + dispatched workers (role binding in mnfs shared-standards) |
| `/milestone-validate` ★ crew at milestone CLOSE | 2026-07-16 | P6 dual gate + P7 QA live-drive (crew survives only as legacy in-flight milestone option) |
| Postgres-lease scheduler + `lease-reaper` | M5 | River periodic jobs (ADR 0067/0068) |
| `DoReadOnly` tx path | G1 (817abd59) | `Do` + api-lint guard (authz.Require needs writable tx) |
| deep-research workflow | 2026-07 | lean inline research (workflow ≈4M tokens) |
| Codex dispatch for PLANNER and IMPLEMENTER roles — **operator ruling 2026-07-16, refined same day** | 2026-07-16 | **Planner = Opus subagent** (one batch P2 pass) · **Implementers = sonnet subagents** (TDD slices) · per-slice reviewer = independent sonnet · investigator = sonnet (haiku trivial greps) · mechanical = haiku. **Dual gate at CLOSED UNCHANGED from core §4: cold Opus subagent + GPT-5.6 Sol medium review (via codex, git read-only, fixed-SHA diff)** — the GPT gate arm is explicitly MAINTAINED; codex stays legal for the dual gate only. |
| testdb ownership-gate retirement + mid-suite GC | 2026-07-15/16 | guard-first virginity + `classifyGCCandidate` + `TESTDB_GC_E2E=1` idle-only (§4) |
| Blanket `-count=1` on integration runs | 2026-07-13 | cache-sound selective policy (§2 L1) |

## Amendment log

```
2026-07-16 · all sections · ratified · profile created at 4.5 close boundary; content migrated
  from docs/superpowers/HARNESS.md (2026-07-10..14) + memory field findings (GMR M0–M9,
  units 2.x–4.6); no scout-assumed content — everything above is program-proven.
2026-07-16 · §10 · corrected · codex sandbox fixed by operator; hub probe PROBE-OK — unit-side
  codex exec DISCHARGED from denylist; stdin evidence-pack retained as fallback pattern.
2026-07-16 · §9/§10 · ratified (operator, hub session) · codex/GPT subagents RETIRED for this
  repo (all roles). Claude-only worker matrix: sonnet plan/implement/review (implementer ≠
  reviewer), haiku mechanical, cold Opus dual-gate arm 1 + independent sonnet arm 2 (replaces
  GPT-5.6 Sol arm). Profile wins over core §1 model matrix (core §0 conflict order).
2026-07-16 · §10 · REFINED (operator, same hub session — supersedes the entry above) · scope of
  the codex retirement narrowed to PLANNER + IMPLEMENTER only: planner = Opus subagent,
  implementers = sonnet subagents. Dual gate at CLOSED reverts to core §4 canon: cold Opus +
  GPT-5.6 Sol medium via codex (git read-only) — GPT arm MAINTAINED. Codex legal for the dual
  gate only; per-slice reviewer/investigator stay sonnet, mechanical haiku.
```
