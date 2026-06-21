# Milestone 4 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-21  ·  **Verdict:** see C7 — **PASS**.
> The validator judged and wrote this file only. It edited no source, fixed no findings, and did not
> flip status. The **main session flips status only on this PASS** and then presents the HS-1 gate.

## Inputs loaded

- `milestone.md` (M4 spec — objective, 4-feature decomposition, validation definition).
- All three closed features' `spec.md` + `plan.md` + `evidence.md` (F4.1, F4.2, F4.3).
- Program `README.md` (M0–M3 `passed` + HS-1; M4 `in-progress`) and governing `mission.md`
  (§8 terminal acceptance: H-G ledger empty).
- Aggregate M4 diff: commits `e147f33e` (F4.1) → `4eaee3fb` (F4.2) → `16cbd092` (F4.3); code files =
  the two migrations, `search/v2documents/reader.go` + tests, two new parity tests, and the cilint
  ledger + its test. No input missing or unreadable.

## C1 — Spec & plan conformance (per feature)

Each feature has `spec.md` (approval line filled, populated interview record, consumer contract
declared FIRST), an execution-shaped `plan.md` (TDD order, files touched, test strategy), and an
`evidence.md` whose acceptance table maps row-for-row to the `spec.md` Validation Gate. The consumer
contract was read from the consumer (`reader.go`), not guessed: F4.1/F4.2 specs cite
`reader.go:54-121` / `:89-118` as the source of truth and the producer views match exactly.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F4.1 cd-visibility-contract | ✅ — `v_cd_search_facts` (1:1 facts + `is_company` + `owner_user_id`) and `v_cd_grantee` (bounded restricted-CD edges) match the consumer's required shape; `is_company` replaces the scope-enum literal; no (cd×actor) cross-product (HS-2 avoided). | ✅ — all 5 Validation-Gate rows met (re-verified in C2). | ✅ — no search/CD-repository change; no cross-product; no new authz. | `spec.md` approved 2026-06-21/leandrotca; `evidence.md` |
| F4.2 documents-search-projection | ✅ — `v_document_search_facts` is the pure 1:1 projection of the exact 14 columns search reads; `archived_at` exposed (no baked filter), no COALESCE. | ✅ — migration applies, parity GREEN incl. archived + NULL-snapshot discriminators (re-verified C2). | ✅ — passthrough only; no CD columns; no >14 columns; no repository change. | `spec.md` approved 2026-06-21/leandrotca; `evidence.md` |
| F4.3 search-consume | ✅ — `reader.go` consumes exactly the three published views as-is; visibility decision composed per the published contract; param order `$1..$14`, ordering, pagination, family-port path unchanged (seam only). | ✅ — D6 frozen-raw parity + behavioral guard GREEN; ledger drained; baseline realigned (re-verified C2). | ✅ — no authz change; no new view/producer; no port/service/handler/family-port change. | `spec.md` approved 2026-06-21/leandrotca; `evidence.md` |

All three approval lines are filled with date + operator before code; interview records are populated
(F4.1's HS-2 cross-product question and F4.2's policy-vs-projection question are the decisive,
recorded fail-closed gates). C1 **pass**.

## C2 — Gates re-run, isolated

Re-run by the validator from clean state (`-count=1`, `GOCACHE="$PWD/.gocache"`,
`METALDOCS_DATABASE_URL=postgres://metaldocs:metaldocs@127.0.0.1:5434/...`, `-tags integration` for
PG). Test PG `:5434` container `metaldocs-test-pg` confirmed UP and reachable — no HS-3.

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| Bootstrap (F4.1+F4.2 migrations) | `go test -count=1 -tags integration ./tests/integration/testdb/...` | `ok …/testdb 3.560s` (0243+0244 apply in full bootstrap) | ✅ real |
| F4.1 | `… ./controlleddocuments/infrastructure/ -run 'TestCDSearchFacts_ParityWithBaseTable\|TestCDGrantee_BoundedSetExcludesRevokedAndUngranted\|TestCDVisibilityContract_ComposedDecisionParityWithRaw'` | 3/3 PASS (0.26s / 0.09s / 0.09s) | ✅ real |
| F4.2 | `… ./documents/repository/ -run TestDocumentSearchFacts_ParityWithBaseTable` | PASS (0.27s) | ✅ real |
| F4.3 D6 + behavioral | `… ./search/infrastructure/v2documents/ -run 'ContractParityWithFrozenRaw\|EnforcesUnifiedVisibility'` | `TestListDocuments_ContractParityWithFrozenRaw` PASS (0.36s); `TestListDocuments_EnforcesUnifiedVisibility` PASS (0.11s) | ✅ real |
| F4.3 full search module | `go test -count=1 -tags integration ./internal/modules/search/...` | `ok` application / delivery/http / v2documents | ✅ real |
| H-G guard | `go run ./tools/cilint ./...` | `cilint-exit=0` | ✅ real |
| cilint unit suite | `go test -count=1 ./tools/cilint/...` | `ok …/internal/analyzers 2.466s` (uncached) | ✅ real |
| Build | `go build ./...` | exit 0 | ✅ |
| Full unit suite | `go test ./...` | no FAIL lines (exit 0) | ✅ real |

Every feature's named gate passed on isolated re-run. C2 **pass**.

## C3 — Senior review of the aggregate milestone diff

Reviewed the whole M4 code diff as one unit.

- **No split-brain.** The visibility decision now has one source: the published views. The frozen-raw
  predicate exists only inside the parity test as a permanent regression baseline, not as a second
  live source of truth.
- **No dead code.** Five raw base-table reads and the `'company'`/`'restricted'` literals are fully
  removed from `reader.go` (grep confirms NONE remain; only `metaldocs.v_*` views referenced). The
  drained `hgPendingRemediation` slice is empty (comments only).
- **No duplication / no feature breaking another.** F4.1 and F4.2 are independent producers; F4.3
  composes both + iam's M3 view. The producer views read only their own module's base tables (+ iam's
  *published* `v_active_user_areas` in `v_cd_grantee`) — the documents-view does NOT join CD, so the
  document↔CD correlation correctly stays in search's JOIN (the milestone's #1 rabbit hole avoided).
- **Contract defined once.** `v_cd_grantee` gates on `visibility_scope = 'restricted'` to mirror the
  inline predicate exactly; `is_company` cleanly replaces the scope-enum literal in the consumer.
  Migrations are forward-only, transactional, `COMMENT ON VIEW`-documented, idempotent
  (`schema_migrations … ON CONFLICT DO NOTHING`).
- Findings: none.
- Staff-engineer bar met? ✅

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api / persistence-migration / module-boundaries ADR-0039) | pass | Migrations apply in full bootstrap; views are SELECT-only owner-published D3a contracts; no producer view reads a third module's base table; active-membership leg is `v_active_user_areas` (`effective_to IS NULL`, no interval reinterpretation). |
| Regression vs prior milestones | all prior gates still pass | M2 read-port parity (`TestActiveInstanceReader_*`, `TestCDFieldReader_*`) and M3 `v_active_user_areas` CD consumers (`TestCanRead_ViewParityWithRaw`, `TestList_ViewParityWithRaw`) GREEN. cilint H-G guard exit 0. `go build ./...` + `go test ./...` green. |

**Pre-existing environmental failures (NOT M4 regressions) — independently confirmed.** Integration
runs surface failures in untouched modules (`iam` Probe A–H; `iam/infrastructure/postgres`
`*_Live`; `controlleddocuments/domain` `TestSequenceAllocatorNextAndIncrement_Concurrent`;
`documents/approval/.../signature` `TestPostgresLimiter_Live`; `documents/approval/repository`
`TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema`). Failure mode is uniform:
`relation "metaldocs.iam_users" does not exist (SQLSTATE 42P01)` — these tests connect to the raw
base DSN whose schema only materializes inside testdb per-test clones. I did **not** trust the
evidence claim: I created a git worktree at the pre-M4 parent (`e147f33e^`) and re-ran the affected
packages — **the identical failure set reproduces with all M4 changes absent**. They are pre-existing,
in modules the M4 diff never touches (the M4 code diff is exactly: the two migrations, `reader.go` +
its tests, two new parity tests, cilint ledger + test). The task's "3 known" list was non-exhaustive,
but every additional failure is the same environmental class and is verified pre-existing. The one M3
artifact that cannot be exercised here — `documents/approval/repository`'s eligible-actors parity
test — is blocked solely by this environmental setup failure (HS-3 not-run for that single test), not
by M4; its M3 sibling CD-consumer parity tests run in a clean package and pass. C4 **pass** (no M4
regression; pre-existing env failures recorded, not false-greened).

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| H-G debt ledger `hgPendingRemediation` (mission §8 terminal precondition) | 5 live C4 entries (C4a–C4e) | **EMPTY** (comments only) | `tools/cilint/internal/analyzers/hgcrossmodule.go` slice empty; `go run ./tools/cilint ./...` exit 0 |
| search reads no foreign base table | 5 raw reads + `'company'`/`'restricted'` literals in `reader.go` | 0 raw reads, 0 literals; only `metaldocs.v_*` views | grep NONE for literals and for `FROM/JOIN public.{documents,controlled_documents,grant tables,user_process_areas}`; root cause (search re-implementing CD's predicate) eliminated by consuming published views |
| anti-symptom-patch | — | clean | No `//cilint:allow-hgcrossmodule` added; ledger rows drained only because the raw reads are actually gone (verified). Negative-baseline test realigned to the empty-ledger end-state: `TestHGCrossModule_LedgerDrained_EmptyAtMissionEnd` asserts the formerly-pending C4d site now FLAGS (not a silent FAIL); suppression mechanism still covered by `TestHGCrossModule_Negative_Exempt`. |
| authz/visibility drift (goal #1) | inline predicate | view-composed | Proven set-identical at SQL level (F4.1 `TestCDVisibilityContract_ComposedDecisionParityWithRaw`, 5 actors × {company,restricted}) AND end-to-end (F4.3 frozen-raw parity), both with revoked-member + ungranted-user discriminators. |

- Could it be built better? No material rebuild needed. The two-view shape (1:1 facts + bounded
  grantee) is the correct construction — it avoids the company-scope cross-product blow-up while
  keeping the projection JOIN at 1 row/CD. Optional future input (not a defer-blocker): the
  environmental raw-base-DSN integration tests across iam/CD/documents/approval should be re-pointed
  at the testdb template clone so a full `-tags integration` run is green — that is outside the M4
  boundary and already recorded as a bounded defer in F4.3's evidence.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean: every
  feature's named acceptance was re-run and mapped (C1/C2).*
- [ ] Fixture/mock passed off as real-provider proof — *clean: all parity/behavioral gates ran on real
  PG `:5434`; sqlmock unit guards are labeled as contract guards, not behavioral proof.*
- [ ] Consumer contract guessed rather than read from the consumer — *clean: contracts cite
  `reader.go` line ranges as source of truth; producers match.*
- [ ] Split-brain (one fact, two sources of truth) — *clean: one live visibility source (the views);
  frozen raw exists only as a test baseline.*
- [ ] Self-judged close / validator edited or fixed code — *clean: validator wrote only this verdict.*
- [ ] Scope drift — *clean: exactly the 4-feature (delivered as 3 producer/consumer features) scope;
  the five C4 sites trace to mission §5 + F0.2 census C4a–C4e.*
- [ ] Symptom-patch — *clean: ledger drained because raw reads removed; no suppression directive; bar
  re-measured at root cause.*

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass. Code-wise: senior-level, contract-clean, no split-brain, no dead code, no
  guessed contract. Function-wise: search consumes only published views end-to-end with proven
  set-identical visibility (revoked + ungranted discriminators) and the H-G ledger is EMPTY — the
  mission's §8 terminal precondition is met. No M4 regression; the integration failures observed are
  pre-existing environmental (raw base-DSN) issues independently confirmed at the pre-M4 commit.
- Handed back to the main session to flip M4 status and present the HS-1 operator gate. M4's HS-1 is
  also the gate before the mission's terminal acceptance (re-run of the F5.1 10-dimension re-audit →
  `mission-validator`).

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only on PASS, by the main session
