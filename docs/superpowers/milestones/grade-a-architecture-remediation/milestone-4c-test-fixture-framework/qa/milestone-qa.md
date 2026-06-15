# Milestone 4c — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-15 (HEAD `9856ca65`)  ·  **Verdict:** PASS (see C7).
> The validator judged and wrote this file only; it edited no source and flipped no status.

## Inputs loaded

- Milestone spec `milestone.md`; all 5 features' `spec.md` + `evidence.md` (`plan.md` present for F4c.1–F4c.4; F4c.5 is documentation-only — spec declares TDD N/A, no `plan.md` required by its own contract).
- Program `README.md` + governing spec `2026-06-14-grade-a-architecture-remediation-design.md` (read via README + milestone linkage).
- Aggregate milestone diff `git diff 071931c9..HEAD` (prior milestone close = M4b drop `071931c9`) and `238ea15f..HEAD` (F4c.3-close phase baseline). No input missing.

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F4c.1 factory-framework | ✅ — factory API surface matches the contract read from the consumer tests (verbatim builder table) | ✅ — 9-subtest self-test green; all 6 ACs mapped to real template-cloned proof | ✅ — no consumer migration, no `db.go` edit, no tripwire weakening | `f4c.1-.../evidence.md` |
| F4c.2 migrate-blocker-files | ✅ — migrated tests genuinely consume `testdb.Open` + `NewTenant/NewUser/NewTaxonomy/NewControlledDoc/NewDocument` + `Scenario.ScheduledRevision` + `SeedWithCaps` (spot-verified in `postgres_approval_repository_integration_test.go`, `scheduled_publish_job_test.go`) | ✅ — 5 approval/repo + 3 jobs + 2 commit_upload + fillin GREEN; fillin classified empirically as a Family-A production-schema defect and **HS-2-escalated** (migration 0241 + ADR 0033), exactly as the spec pre-committed | ✅ — the one production-source change is the operator-approved HS-2 path; abandoned WIP discarded; 3 named helpers deleted | `f4c.2-.../evidence.md` |
| F4c.3 migrate-remaining | ✅ — declarative grep scope; migrated files consume the stable factory API (no new builder invented) | ✅ — AC1–AC5/AC7/AC8 PASS; AC6 honestly recorded PARTIAL (pre-existing M4b scenarios debt, baseline-equality proven at `4b5e2fc5`) | ✅ — HS-6 over-migration of sqlmock unit tests reverted (`4b5e2fc5`) and scope amended with `//go:build integration` filter; `db.go`/`factory.go` untouched | `f4c.3-.../evidence.md` |
| F4c.4 ci-grep-guards | ✅ — guard consumed by CI step + local dev; exactly R1–R4 + `testdb/**` exception per spec | ✅ — re-verified independently (see C2); pgtest retired (zero callers → dir deleted); discipline doc 134 lines + indexed | ✅ — additive only; the two test-file R1/R2/R4 fixes are recorded scope admits, not silent | `f4c.4-.../evidence.md` |
| F4c.5 docs-adr | ✅ — harness doc names all 9 shipped builders (no drift); ADR 0034 canonical headers | ✅ — AC1–AC8 re-verified structurally (see C2); wiki-curator dispatched + returned | ✅ — documentation-only; no production-source change | `f4c.5-.../evidence.md` |

No missing spec/evidence row. **C1 pass.**

## C2 — Gates re-run, isolated (validator-run, not trusted from transcript)

| Gate | Command re-run | Real output | Pass? |
|------|----------------|-------------|-------|
| Discipline guard clean on HEAD | `bash scripts/check-test-discipline.sh` | `test-discipline: clean (63 integration test files checked)` exit 0 | ✅ |
| Guard trips each rule + clears | planted a fresh `//go:build integration` file with R1+R2+R3+R4 violations | reported `R1`,`R2`,`R3`,`R4` on the planted lines, exit 1; after revert `clean` exit 0 | ✅ |
| Allowlist files hold real pre-existing violations | grep each R3/R4 allowlist file for the literal/bare-`documents` | every allowlist file contains ≥1 genuine pre-existing hit (debt-tracking, not debt-hiding) | ✅ |
| Harness doc ≥ 80 lines | `wc -l wiki/quality/integration-test-harness.md` | `225` | ✅ |
| ADR 0034 canonical headers | `grep -E "^## (Context\|Decision\|Consequences\|References)" wiki/decisions/0034-…md` | 4 matches | ✅ |
| ADR 0034 + harness indexed | `grep 0034 wiki/decisions/index.md`; `grep integration-test-harness wiki/quality/index.md` | both match | ✅ |
| 9 factory builders shipped | `grep -E "^func (New\|.*Scenario)" tests/integration/testdb/factory.go` | 7 `New*` + 2 `Scenario` = 9, all named in harness doc | ✅ |
| pgtest retired | `grep -rn "pgtest\." --include=*.go internal/`; `ls internal/testsupport/pgtest/` | zero real refs (only 2 migration-comment mentions); directory absent | ✅ |
| `db.go` unchanged (fix-not-adapt) | `git diff --stat 071931c9 HEAD -- tests/integration/testdb/db.go` and `238ea15f..HEAD` | empty (both) | ✅ |
| Builds (integration + plain) | `go build -tags integration ./tests/integration/testdb/... .../approval/...`; `go build ./...`; `go vet` untagged approval repo | all exit 0 (pgtest deletion left no dangling import; split-file keeps both tag states compiling) | ✅ |

Live-DB test re-execution (`go test -tags integration`) was **not** re-run by the validator — no operator DSN/IntegreSQL in this sandbox. The named integration GREENs are taken from feature evidence (wall-clock, non-zero seconds, real-DB labeled, with an explicit anti-skip pre-flight DSN gate in F4c.3 after a false-GREEN-on-SKIP was caught and reverted). Every command the validator *could* run from clean state passed. **C2 pass** (with the live-DB caveat noted; see C6 honesty assessment).

## C3 — Senior review of the aggregate milestone diff

Diff `071931c9..HEAD`: 33 files, +2328/−1238. Scoped exactly to test infrastructure (migrated `*_test.go`), the `testdb` framework (`factory.go` new, `fixtures.go` extended, `db.go` untouched), the CI guard (`check-test-discipline.sh`, one workflow step), docs/ADRs, and two authorized production-source touches.

- **Only two non-test production-source changes:** `db/migrations/0241_*` (HS-2 fillin trigger fix, ADR 0033) and deletion of `internal/testsupport/pgtest/pgtest.go` (F4c.4 Q4 zero-callers). Both authorized and recorded.
- **Baseline trigger "split-brain" examined and cleared.** `db/baseline/0001_current_schema.sql:620` still holds the *old* buggy trigger body while migration 0241 corrects it. This is **not** split-brain: ADR 0033 (and the prior ADR 0032 precedent) records the curated baseline as a **frozen historical snapshot**; the migration tail carries forward state. The `testdb` harness `ApplyCuratedBootstrap` applies baseline → reference-data → **migrations**, so 0241 overrides the stale body in every clone — the fillin GREEN exercises the corrected trigger for real. Convention is consistent and documented.
- `SetCapsOnDB` (`is_local=false`) lives **only** inside `tests/integration/testdb/fixtures.go` (the R2 exception zone), docstring-constrained to MaxOpenConns=1 isolated per-test DBs where leak-across-sessions is impossible by construction; flagged as a bounded HS-2 defer (refactor production to DBTX-variadic). R2 scan of test files outside `testdb/` is empty. Sound.
- No duplicated logic across features, no dead code left by a superseded approach, no feature breaking another (F4c.3 verified F4c.2 paths still GREEN after the `fixtures.go` extension).

Staff-engineer bar met? ✅

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Workflow class = test-infrastructure + tooling | pass | CI guard RED/GREEN re-confirmed by validator; factory builders seeded+returned in migrated tests; docs accurate to shipped API |
| `backend-api-qa-checklist` (changed code is backend test infra) | pass | no production runtime/route/contract change (only the HS-2 trigger + pgtest deletion) |
| Database rules (`wiki/database/`) for the schema touch | pass | migration 0241 forward-only/idempotent, ADR-recorded, baseline-frozen convention honored |
| Regression vs prior milestones (M0–M3 gates; M4 F4.1–F4.6; M4b drop `071931c9`) | pass | M4-blocker set (incl. `TestCreateDocumentTx_PopulatesAllSnapshotColumns` + the 8 approval tests) recorded GREEN in F4c.3/F4c.4 evidence under operator DSN with `db.go` unchanged; both build states clean; M4b drop untouched |

Pre-existing `tests/integration/scenarios/` failures are **carried M4b post-teardown debt** (baseline-equality proven), not an M4c regression. **C4 pass-with-defers.**

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| M4 close-gate blocker (cross-test state leakage + `search_path` drift on shared `pgtest` DB) | two ungoverned harnesses; `document_profiles_pkey` 23505 collisions; F4.1a Gate #5 environment-coupled | one unified template-DB-per-test framework; all stateful tests migrated; pgtest retired | **structural**, not symptom-patched: `db.go` empty-diff (no harness adapt); per-clone isolation + minted-unique codes; CI guard R1–R4 prevents regression; fillin Family-A defect fixed at root via 0241 (a genuinely dead security trigger restored), not masked |
| Discipline enforced not asserted | inline `set_config`/`is_local=false`/hardcoded tenant UUID/bare `documents` drift across ~35+ sites | guard fails-on-violation/passes-on-clean (validator-re-verified); 5 named local seed helpers deleted | grep-proof + planted-violation proof |

Could it be built better? The `SetCapsOnDB` `is_local=false` helper is the one residual: a production `DBTX`-variadic refactor would let every guarded-write SUT assert tx-locally and retire the helper. Correctly captured as a bounded HS-2 defer — does not make the current construction unsound. Recommend it as M5 input / a post-v1 follow-up.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — **clear.** AC6 suite-level RED is honestly disclosed as pre-existing M4b debt with baseline-equality proof; per-feature ACs each map to named real-DB runs.
- [ ] Fixture/mock passed off as real-provider proof — **clear.** Real (template-cloned Postgres) vs sqlmock-unit explicitly separated; the HS-6 false-GREEN-on-SKIP was caught and reverted (`4b5e2fc5`).
- [ ] Consumer contract guessed — **clear.** Factory API read verbatim from consumer tests; migrated tests verified to consume it.
- [ ] Split-brain — **clear.** Baseline-vs-migration trigger divergence is the documented frozen-baseline convention (ADR 0032/0033), not two competing live sources.
- [ ] Self-judged close / validator edited code — **clear.** Validator wrote only this verdict; the planted-violation file was created and removed within a single check, leaving the tree at HEAD.
- [ ] Scope drift — **clear.** Two HS-6 events (over-migration; false-GREEN) surfaced, reverted, and recorded; the HS-2 production touch was operator-approved and ADR-recorded.
- [ ] Symptom-patch — **clear.** Root cause closed structurally; fillin trigger fixed at root.

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass: **code-wise** (scoped, contract-clean, no split-brain, no dead code, both build states green, guard sound) and **function-wise** (the unified framework does what the milestone promised — `db.go` empty-diff structural proof, M4-blocker set GREEN at root cause, discipline enforced not asserted, pgtest retired, docs accurate).
- Handed back to the main session to flip status and present the HS-1 operator gate.

### Non-blocking notes for the main session (validator findings, not fixes)

1. **README hard-stop table is behind.** The program `README.md` hard-stop log stops at the M4c open gate; it does not yet record (a) the F4c.2 HS-2 fillin-trigger escalation + operator approval (ADR 0033 / migration 0241), (b) the two F4c.3 HS-6 events, or (c) the per-feature close gates. The main session should append these on close (they are fully recorded in the feature evidence + ADRs; this is a logging-currency note, not a defect).
2. **Carried defers to track at program close:** `tests/integration/scenarios/` M4b post-teardown refresh; the `SetCapsOnDB` `is_local=false` → DBTX-variadic HS-2 candidate; the 5 allowlist files (R3×3/R4×3) with written removal triggers; iam-membership seed helpers (F4c.3b micro-task) — each already has a written trigger.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending operator review of this PASS.
> - Status flipped in `README.md`: not yet — only on PASS + operator go (then re-dispatch the M4 validator per the milestone's stated unblock chain).
