# Milestone 4 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md` + `../validation-contract.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-04  ·  **Verdict:** see C7.
> **Re-validation:** this is the re-validation after fix-feature **F4.5** (authz soft-GUC NULL
> hardening, commit `4ce4d308`) closed the F4.2 live-run defer (HS-2, operator-directed). Prior PASS
> was after F4.4 (HEAD `e5d0d91a`); current HEAD `cd1f3d64`. The two new commits since prior PASS are
> `4ce4d308` (F4.5) and `cd1f3d64` (README). Aggregate M4 diff measured from M3 close `2e4411ca`.

## Inputs loaded

- `milestone.md`, `validation-contract.md`, program `README.md`, mission spec (§7 M4 via README).
- Every feature `spec.md`/`plan.md`/`evidence.md`: F4.1, F4.2, F4.3, F4.4, F4.5.
- Aggregate diff `git diff 2e4411ca..HEAD` (56 files, +2397/−212).
- No input missing or unreadable.

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F4.1 state-machine-unification | ✅ (openapi enum = domain `CanTransitionDocumentStatus` = DB-trigger parity) | ✅ 9-status exhaustive fn; 0 scattered `if status !=` guards (census below) | ✅ | `state.go`, `state_parity_test.go`, `state_test.go` |
| F4.2 publish-race | ✅ scheduler no-op sentinel + repo `ErrStaleRevision` read, not guessed | ✅ real concurrent test, single winner, correct terminal state — **now live-green** | ✅ no prod code changed | `f4.2/evidence.md`, my C2 re-run |
| F4.3 concurrency-idiom | ✅ contract §3.5 fallback (ADR-split); premise-false surfaced honestly | ✅ ADR 0066 records intentional split, If-Match target; HS-7 operator-ratified | ✅ no wire change | `f4.3/evidence.md`, ADR 0066 |
| F4.4 drop-stray-release-file | ✅ | ✅ `docs/release/` net-zero in M4 diff (verified) | ✅ | `f4.4/evidence.md` |
| F4.5 authz-guc-null-hardening | ✅ consumer = all `MustActorID`/`MustTenantID` callers; contract read, not guessed | ✅ TDD pins green; unit+integration green; downstream F4.2 live-green | ✅ stricter-or-equal, no `authz.Require`/RLS change | `f4.5/evidence.md`, my C2 re-runs |

All five features carry `spec.md` (approval line filled), `plan.md` (execution-shaped), `evidence.md`
(acceptance table matching gate). F4.5 interview record populated (3 rows). No missing artifacts.

## C2 — Gates re-run, isolated (by the validator, from clean state)

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| build | `go build ./...` | BUILD_EXIT=0 | ✅ |
| F4.5 unit | `go test ./internal/modules/iam/authz/ -count=1` | `ok … 1.369s` | ✅ |
| F4.5 NULL pins | `go test ./internal/modules/iam/authz/ -run Null -count=1 -v` | `PASS: TestMustActorID_ReturnsErrWhenGUCNull`, `PASS: TestMustTenantID_ReturnsErrWhenGUCNull` | ✅ |
| F4.5 integration | `go test ./internal/modules/iam/authz/ -tags integration -count=1` (real Postgres) | `ok … 205.735s` | ✅ |
| F4.1 state+parity | `go test ./internal/modules/documents/domain/ -run 'State|Transition|Parity' -v` | `PASS: TestCanTransitionDocumentStatus_DBTriggerParity`, `PASS: TestCanTransitionDocumentStatus` | ✅ |
| F4.1 guard census | grep for scattered lifecycle status-equality guards in approval services | 0 matches | ✅ |
| F4.2 live race | `go test …/approval/application/ -tags integration -run TestPublishRace -count=1 -v` (real Postgres) | `--- PASS: TestPublishRace (239.65s)` + all 4 subtests PASS; `ok … exit 0` | ✅ |

F4.2 note: one `db.go:101` line ("drop isolated test database … timeout") is a **teardown-only**
cleanup message (DB drop after the assertions), not a test assertion — the enclosing test and all 4
subtests reported PASS and the package returned `ok` (exit 0). Genuinely live-green on real Postgres,
re-run by me, not trusted from the transcript.

## C3 — Senior review of the aggregate milestone diff

Whole M4 diff (`2e4411ca..HEAD`) reviewed as one unit.

- **F4.5 authz change** — one shared `readSoftGUC` (`sql.NullString`) now backs all four soft-GUC
  readers (`MustActorID`, `MustTenantID`, `softGUC`, `loadAssertedCaps`); this **collapses** a prior
  three-way behavior split into one idiom (anti-split-brain, positive). Accept path byte-identical;
  emitted SQL unchanged (constant literals). Reject path swaps a driver crash for the documented
  sentinel — **stricter-or-equal**, fail-closed preserved. `softActorID`/`softTenantID` still fail-soft
  to their defaults (`"system"`/`""`). No dead code.
- **F4.2 harness change** — test-only: seeds identity exactly as production middleware does (manual
  path `platformtenant.WithTenantID/WithActorID`; scheduler path `authz.WithBackgroundBypass` +
  in-tx `SeedTxTenant`), µs-truncated timestamp, text `resource_id` compare. **No production
  publish-path code changed** (confirmed: publish/scheduler service diffs are F4.1 routing only).
- **F4.1 `rejected` removal** is consistent across openapi enum, `api.gen.go`, FE parsers, and
  `DocumentStatus` domain — NOT split-brain. The surviving `"rejected"` literals live in the distinct
  **approval-instance / signoff** types (`InstanceStatus`, `DocState`, `ApprovalInstance…Status`),
  which are a different concept from document lifecycle status and out of F4.1 scope.
- **Module boundary** — `internal/platform/db` still does not import `internal/modules/iam`
  (grep = only the guard comment, no import). Boundary intact.
- **`docs/release/`** — net-zero in the M4 diff (F4.4 fix holds; `git ls-files docs/release/` empty).

- Findings: none.
- Staff-engineer bar met? ✅

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api + test-discipline) | pass | Contract-first: openapi edit + regen only, zero hand-edits to gen; testdb factory used for F4.2/authz integration (not sqlmock); targeted `-run` only, full suite not run (per box constraint). |
| Regression vs prior milestones | all still pass | F4.1 state-machine + DB-trigger parity re-run green; F4.5 did not weaken authz (fail-closed preserved, tenancy/RLS untouched — M3 chokepoint intact; authz integration green on real DB exercising RLS backstop + bypass). M0 VersionRef wire shape unaffected (F4.3 = no wire change). `go build ./...` green. |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| M4.1 one exhaustive 9-status transition fn, 0 scattered guards | scattered `if status != X` across services | one `CanTransitionDocumentStatus` + 0 census hits | census grep = 0; parity test green |
| M4.2 publish race proven safe by a real concurrent test | argued safe-by-construction; live run deferred | **live-green on real Postgres** (my re-run, 4/4 subtests) | `TestPublishRace` PASS exit 0 |
| F4.5 authz NULL-GUC gap (HS-2) | driver crash on cold-connection identity read | correct `ErrActorContextMissing`/`ErrTenantContextMissing` sentinel | 2 NULL pins green; class collapsed to one reader (root cause, not symptom-patch) |

- Could it be built better? The OCC transport unification (F4.3) remains a real defer — ADR 0066
  records the concrete `platform/occ` kernel target design and names M9 as the candidate change. Not a
  soundness defect for M4 (intentional, operator-ratified split). No other retrospective item.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — clean (per-feature C2 rows)
- [ ] Fixture/mock passed off as real-provider proof — clean (F4.2 + authz integration re-run on **real Postgres** by me)
- [ ] Consumer contract guessed rather than read from the consumer — clean (F4.5 sentinel + F4.2 no-op read from source)
- [ ] Split-brain (one fact, two sources of truth) — clean (F4.5 collapses the reader split; `rejected` refs are a distinct type)
- [ ] Self-judged close / validator edited or fixed code — clean (validator only judged + wrote this file)
- [ ] Scope drift (work beyond the spec, no rationale) — clean (F4.5 recorded with HS-2 operator-direction rationale)
- [ ] Symptom-patch (bar moved by masking, root cause intact) — clean (F4.5 fixes the reader class; F4.2 live-proven)

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass. Code-wise: senior-level, contract-clean, no split-brain, no dead code,
  fail-closed preserved, module boundary intact, `docs/release/` net-zero. Function-wise: F4.1 coverage
  + parity green, F4.2 publish race **live-green on real Postgres** (my isolated re-run), F4.3 ADR-split
  recorded + operator-ratified, F4.5 NULL-GUC class fixed at root and TDD-pinned. No fix feature needed.
- Handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending — includes F4.3 HS-7 ADR-split ratification (README notes "HS-7 okay").
> - Status flipped in `README.md`: main session only, on this PASS.
