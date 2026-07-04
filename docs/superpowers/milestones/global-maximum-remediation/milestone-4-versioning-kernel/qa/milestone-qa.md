# Milestone 4 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + `../validation-contract.md` (D4 binding) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-04 (re-dispatch after F4.4 fix)  ·  **Verdict:** see C7 — **PASS**.
> This is the re-validation following the prior FAIL, whose single blocker (F-1: stray tracked `docs/release/v2-name-inventory.md`) is now resolved by fix-feature `f4.4-drop-stray-release-file` (commit `c7aebfe3`). Run only after every feature is closed. The validator judges and writes this file; the main session flips status only on a PASS. The validator never edits code, fixes findings, or flips status.

## Inputs loaded

All present and readable: milestone spec, `validation-contract.md` (incl. the §3.7 HS-7 erratum that supersedes §3.4/§3.5/§3.6), program README + `mission.md`, all four features' (`f4.1`, `f4.2`, `f4.3`, `f4.4`) `spec.md`/`plan.md`/`evidence.md`, and the aggregate M4 diff (commits `a83a560e` … `c7aebfe3`). No blind-judge FAIL.

## Prior verdict & delta

The 2026-07-04 verdict was **FAIL** on a **single** blocker: commit `51950c26` accidentally tracked `docs/release/v2-name-inventory.md` (316 lines), breaching validation-contract §5 / CLAUDE.md ("never commit `docs/release/`"), no rationale recorded (finding F-1, C3/C6). All of C1/C2/C4/C5 passed on independent re-run. This re-validation re-runs the binding gate and confirms F-1 is cleared with nothing new surfaced.

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F4.1 | ✅ — approval services (consumer) route through `domain.CanTransitionDocumentStatus`; DB trigger (consumer) is exact mirror, parity pinned; wire/FE no longer list `rejected` | ✅ — all §1.6 exit criteria met (see C2) | ✅ — no route added for DB-legal-unused arcs; `archived` retained; enforcement stayed in DB; approval-INSTANCE FSM untouched | spec/plan/evidence present, `Approved for code: 2026-07-04`, interview record populated |
| F4.2 | ✅ — real concurrent testdb integration test proving exactly-one-winner + terminal state + single side-effect, matching §2.2 cell-for-cell | ✅ (test authored + compiles + vets under `-tags integration`; live-green is a contract-§6-permitted bounded defer, see C4) | ✅ — targeted `-run` only; no publish business logic changed (safe-by-construction proven, no choke point needed) | spec/plan/evidence present, approved 2026-07-04, interview record populated |
| F4.3 | ✅ — decision recorded (ADR 0066) + contract §3 re-opened with loud dated non-silent erratum; the binding F4.3 decision is the §3.7 HS-7 erratum, not the struck §3.4/§3.5 | ✅ — §3.7 corrected exit criteria met: ADR 0066 landed, erratum dated, spec/plan/evidence updated, zero templates/documents wire change, HS-7 headlined for HS-1 gate | ✅ — no templates migration, no `lock_version` rename, no documents change | spec/plan/evidence present, approved 2026-07-04, interview record populated |
| F4.4 | ✅ — consumer = repo history + M4 gate; required outcome (`docs/release/` untracked, kept on disk, dedicated commit) delivered | ✅ — `git ls-files docs/release/` empty; file `??` on disk; `go build ./...` green; fix commit touches no other functional path (see C2/C3) | ✅ — not deleted from disk, no `.gitignore` add, no source/test/contract change, no history rewrite | spec/plan/evidence present, `Approved for code: 2026-07-04`, interview record populated (1 row) |

All four features have complete spec/plan/evidence with filled approval lines and non-empty interview records. C1 conformance: **PASS**.

## C2 — Gates re-run, isolated (validator re-ran from clean state, not trusted from transcript)

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| all | `go build ./...` | `BUILD_EXIT=0` | ✅ |
| F4.1 | `go test ./internal/modules/documents/domain/ -run 'CanTransitionDocumentStatus\|DBTriggerParity' -count=1` | `ok metaldocs/internal/modules/documents/domain 1.227s` | ✅ |
| F4.1 | `grep -rn DocStatusRejected internal/ \| grep -v _test.go` | none remain | ✅ |
| F4.2 | `go vet -tags integration ./internal/modules/documents/approval/application/` | `VET_EXIT=0` | ✅ |
| F4.4 | `git ls-files docs/release/` | empty (no tracked file) | ✅ |
| F4.4 | `test -f docs/release/v2-name-inventory.md` | `YES on disk` (untracked `??`) | ✅ |
| F4.4 | `git show --stat c7aebfe3` | touches exactly: `docs/release/v2-name-inventory.md` (316 del, index removal) + the 3 f4.4 records (spec/plan/evidence) — no other path | ✅ |
| M2 regression | `go test ./internal/modules/iam/authz/... -count=1` | `ok metaldocs/internal/modules/iam/authz 1.284s` | ✅ |

C2 re-run: **PASS** — every named gate re-run green from clean state.

**F4.4 fix cross-checks (independently verified, not trusted from evidence):**
- `docs/release/` touches across the full M4 commit set: only `51950c26` (+1, the accidental add) and `c7aebfe3` (+1, the index removal); the other eight commits = 0. No M4 commit *going forward* touches a `docs/release/` path.
- **Net aggregate M4 diff** (`git diff --stat a83a560e^..c7aebfe3 -- docs/release/`) = **empty**: the stray file is net-zero over the milestone — as if never added. The F-1 breach is fully undone in the tracked history.
- `c7aebfe3` is `HEAD`; it is the milestone tip. The fix is a dedicated commit; no source/test/contract/generated change rode along.

## C3 — Senior review of the aggregate milestone diff

- **F-1 (prior blocker) — RESOLVED.** The stray `docs/release/v2-name-inventory.md` is untracked again (`git ls-files docs/release/` empty), removed via `git rm --cached` in a dedicated, rationale-carrying commit (`c7aebfe3`). Net M4 diff no longer adds any `docs/release/` path. The contract §5 / CLAUDE.md "never commit `docs/release/`" constraint is honored across the aggregate.
- **Split-brain:** none — F4.1 closed the app-vs-DB transition split-brain; the friendly-first-line function is the single app authority, pinned to the DB trigger by the parity test.
- **Dead code:** dead `CanTransitionDocument` FSM removed; no second document-status FSM remains.
- **Duplication / one feature breaking another:** F4.1 refactor left the approval unit suite green; F4.2 changed no production code; F4.3 changed no product code; F4.4 changed only the git index + its own records.
- Staff-engineer bar on the deliverable code **and** on the committed change set as a whole: now met.

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| backend-api / contract-first (`api/openapi` + regen, zero hand-edits) | pass | document-status enum tightened; regen pure; approval-DECISION enums correctly retain `rejected` |
| test-discipline (testdb factory for the race integration; targeted `-run`; no full suite) | pass | F4.2 uses `tests/integration/testdb`; full 20-min integration suite correctly not run |
| F4.2 live-green run | pass-with-defer | contract §6-permitted bounded defer (20-min box + pre-existing authz NULL-GUC cold-connection scan gap reproducing byte-identically on a reference integration test); honestly disclosed, fixture-vs-real distinguished, flagged `task_e03a4383` for M8. Legitimately bounded, not a masked failure. |
| Regression vs M0–M3 | all still pass | `go build ./...` green; `DocStatusRejected` fully gone; `iam/authz` tests green (M2 tripwire/M3 seed not regressed); no `.env`/`node_modules`/`docs/release` leaks in the aggregate diff |

C4: **pass**. The F4.2 defer is contract-legitimate.

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| One exhaustive transition function; zero scattered lifecycle guards | dead 3-of-9 FSM called 0×; real legality scattered as `if status != X` | single `CanTransitionDocumentStatus`, routed by all 9 services, census = 0 | root cause fixed (parity test pins fn == DB trigger); not symptom-patched |
| Publish race proven safe (real concurrent test) | asserted-but-unverified | real testdb barrier test, exactly-one-winner + terminal + single side-effect | proven by construction; no production change needed |
| One OCC idiom or ADR-recorded exception | undocumented accidental split | ADR 0066 records intentional split + If-Match target + deferred-unification charter | undocumented drift closed by a decision record; HS-7 handled by loud dated erratum, ratification headlined for HS-1 |
| Change-set hygiene (no stray `docs/release/`) | F-1: stray file tracked in `51950c26` | untracked via `c7aebfe3`; net M4 diff = 0 for `docs/release/` | root cause (accidental `git add` sweep) undone at the index; kept on disk per non-goal |

- Could it be built better? Deliverables are sound. Optional hardening surfaced by F4.4 (add `docs/release/` to `.gitignore` to prevent a future re-stage) is correctly left to the operator at HS-1, out of the minimal fix scope. Non-blocking.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence
- [ ] Fixture/mock passed off as real-provider proof — *(F4.2 distinguishes fixture from real and discloses the live-green defer honestly)*
- [ ] Consumer contract guessed rather than read from the consumer
- [ ] Split-brain (one fact, two sources of truth) — *(the milestone closes one; parity test pins it)*
- [ ] Self-judged close / validator edited or fixed code
- [ ] Scope drift (work beyond the spec, no rationale) — *(prior F-1 breach RESOLVED by f4.4; net M4 diff no longer touches `docs/release/`; f4.4 itself is a validator-named fix feature with recorded rationale)*
- [ ] Symptom-patch (bar "moved" by masking, root cause intact)

No box checked → C6 clean → no forbidden-list hit.

## C7 — Verdict

- **VERDICT: PASS**
- The sole prior blocker (F-1 / C3 / C6 — stray tracked `docs/release/v2-name-inventory.md` in `51950c26`) is **resolved** by fix-feature `f4.4-drop-stray-release-file` (commit `c7aebfe3`): `git ls-files docs/release/` empty, file kept on disk (`??`), dedicated commit touching only the index removal + its own f4.4 records, net aggregate M4 diff for `docs/release/` = zero. No M4 commit going forward touches a `docs/release/` path.
- All feature deliverables re-verified green from clean state: F4.1 unified `CanTransitionDocumentStatus` + DB-trigger parity (`ok` on `CanTransitionDocumentStatus|DBTriggerParity`), `rejected` fully removed (no `DocStatusRejected` refs); F4.2 real concurrent race harness (`go vet -tags integration` clean, single-winner contract per §2.2, bounded live defer legitimate); F4.3 ADR 0066 intentional-split decision (the §3.7 HS-7 erratum is the binding decision). `go build ./...` green; M0–M3 not regressed (`iam/authz` tests green).
- Nothing new surfaced. Both dimensions (code-wise + function-wise/QA) pass.
- The milestone may advance. The main session may now flip status and present the **HS-1 operator gate** (including the F4.3 §3.7 HS-7 ratification and the F4.2 live-green defer for acknowledgement).

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): now reachable — present for approval
> - Status flip in `README.md` / roadmap: main session's action on this PASS (the validator does not flip status; a `README.md` modification is currently pending in the worktree — that is the main session's to commit, not the validator's)
