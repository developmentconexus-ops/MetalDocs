# Milestone 1 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + `../f1.1-resolution-constants/spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-20  ·  **Commit under test:** `4a21a7dc` (local)  ·  **Verdict:** see C7 — **PASS**.
> The validator judges and writes this file; the **main session flips status only on a PASS**. The validator
> never edits code, fixes findings, or flips status.

## Inputs loaded (fail-fast check)

All required inputs present and readable — judge was not blind:
- Milestone spec `../milestone.md`; mission `../../mission.md` (§2/§5 row 3/§7 M1/§9/D6); program `../../README.md`.
- F1.1 `spec.md` / `plan.md` / `evidence.md` (all present, all populated).
- Governing ADR-0039 (`wiki/decisions/0039-…md`, worked-classification row 3); M0/F0.2 `census.md` (Category A = A1–A3).
- Aggregate diff `git show 4a21a7dc` (2 source files + 4 milestone/feature docs).

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F1.1 | ✅ | ✅ | ✅ | See below |

- **spec.md approved before code:** `Approved before code: 2026-06-20 / leandrotca` — filled. ✅
- **Interview record populated:** 5-row Q&A table resolving vocabulary owner, owner-vs-CD-local, field-retype boundary, import cycle/collision, HS-6 completeness. ✅
- **Consumer contract honored (read from the consumer, not guessed):** the vocabulary owner is `templates/domain` (`version.go:14-15` `VersionStatusPublished="published"`, `VersionStatusObsolete="obsolete"`), populated into `TemplateVersionCandidate.Status` via the ADR-0030 `GetTemplateVersionState` port. The fix references **the owner's** constants (`templatesdomain.VersionStatus*`), not a CD-local copy — independently confirmed in `resolution.go` (import alias `templatesdomain` + 3 `string(templatesdomain.VersionStatus*)` comparisons). Source-of-truth line in spec matches the real owner. ✅
- **plan.md execution-shaped:** baseline → guard-test → 3 named line edits → green-after verify → evidence; files-touched list; test strategy (characterization parity lock + 1 boundary guard). Not a re-spec. ✅
- **evidence.md acceptance table maps row-for-row to spec Validation Gate** (behavior unchanged / regression guard / build / 0 literals / cilint). ✅
- **Non-goals respected:** no `Status` field retype (stays `*string`, confirmed in source); no port-signature change; no SQL/port/view/migration; no CD-local constants; no `CDStatus`/`api.gen.go` edits — verified against the diff. ✅

## C2 — Gates re-run, isolated

Re-run by the validator from clean state (not trusted from the evidence transcript):

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| F1.1 | `go build ./...` | `BUILD_EXIT:0` | ✅ |
| F1.1 | `go test ./internal/modules/controlleddocuments/domain/ -run TestResolve -v` | `10/10 PASS` (9 characterization + `TestResolve_UsesTemplatesVocabulary`); `ok …/domain` | ✅ |
| F1.1 | `go test ./internal/modules/controlleddocuments/... ./internal/modules/templates/...` | all `ok` / no-test-files; `WIDE_EXIT:0` | ✅ |
| F1.1 | `grep -nE '"published"\|"obsolete"' …/resolution.go` | no output (`EXIT:1` = 0 matches) | ✅ |
| F1.1 | `go run ./tools/cilint ./...` | `CILINT_EXIT:0` | ✅ |

All observable surface is in-memory domain logic exercised by the unit suite — no route, DB, or provider; no fixture-as-real substitution, no skipped integration step (none applies to F1.1).

## C3 — Senior review of the aggregate milestone diff

Whole-milestone diff (`git show 4a21a7dc`) reviewed as one unit. Source change is exactly: import alias `templatesdomain "metaldocs/internal/modules/templates/domain"` + 3 literal→constant swaps (`resolution.go` post-edit lines 46, 59, 62 — original :42,:55,:58, shifted +4 by the new import block) + 1 added regression-guard test. Stat: `+310/-4` across 6 files (4 are milestone/feature `.md`).

- **Findings:** none.
  - No split-brain: the status vocabulary now has **one** source of truth (`templates/domain.VersionStatus*`); CD references it rather than duplicating. The pre-change literals were the split-brain — this removes it.
  - No dead code; no duplicated logic; no CD-local alias constants.
  - No feature broke another: CD + templates module suites all `ok`; full `go build ./...` clean.
  - **Regression vs M0 (C3 cross-check):** the F0.3 cilint H-G guard (`tools/cilint/internal/analyzers/hgcrossmodule.go`) was last touched by M0 commit `51ddc875`; M1's commit touches no file under `tools/cilint` and `resolution.go` is correctly **absent** from the guard's ledger (non-SQL site; `grep resolution` → 0). `hgPendingRemediation` debt ledger untouched; `go run ./tools/cilint ./...` exit 0 → no regression.
  - No import cycle: `templates/domain` imports only `errors`/`time` (verified in `version.go`) — nothing from `controlleddocuments`. The `domain`/`domain` package-name collision is correctly resolved by the `templatesdomain` alias.
- **Staff-engineer bar met?** ✅ Minimal, drift-proof, behavior-preserving; `string(...)` conversion is the correct minimal form given the field stays `*string` (retype is an HS-2 cross-module contract change, correctly deferred to M2).

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend code-quality / module-boundaries — DDD vocabulary ownership) | pass | No API/persistence/migration surface touched. Foreign vocabulary now referenced by the owner's typed constant, compiler-checked. |
| Regression vs prior milestone (M0) | all still pass | ADR-0039 unchanged; F0.3 cilint guard green on full tree (`CILINT_EXIT:0`); debt ledger untouched. |
| TDD honesty (C4 honesty rule) | pass | Disposition labeled honestly as **behavior-preserving refactor / characterization parity lock**, explicitly **not** RED-first (constant values equal the prior literals, so the new guard is green pre- and post-edit). No fake-RED claim. The 9 existing characterization tests legitimately use bare literals as wire-value inputs; the new guard wires the owner's constants — a real future-drift tripwire, not redundant. |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Category-A literal-coupling (census A1–A3) | 3 bare `"published"`/`"obsolete"` literals in `resolution.go` (foreign templates vocabulary duplicated as magic strings) | **0 sites** | `grep -nE '"published"\|"obsolete"' …/resolution.go` → 0 matches; the 3 comparisons now read `string(templatesdomain.VersionStatusPublished/Obsolete)` — the **owner's** published constants, not a CD-local copy. Root cause (duplicated foreign vocabulary that can silently drift) eliminated; not a symptom-patch (a CD-local alias const would have been the symptom-patch — explicitly forbidden by spec and confirmed absent). |

- **Could it be built better?** The only residual is the `string(...)` conversion, which exists because `TemplateVersionCandidate.Status` stays `*string`. Retyping it to `*VersionStatus` would drop the conversion but ripples into the ADR-0030 `GetTemplateVersionState` port return type — a cross-module contract change correctly bounded out of M1 (HS-2) and recorded as a bounded defer (owner: M2). The current construction is sound; no defect. No materially better in-scope construction exists.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean (per-criterion mapping in C1/C2)*
- [ ] Fixture/mock passed off as real-provider proof — *clean (pure in-memory domain logic; honestly labeled)*
- [ ] Consumer contract guessed rather than read from the consumer — *clean (owner read from `templates/domain` + ADR-0030/0039)*
- [ ] Split-brain (one fact, two sources of truth) — *clean (this milestone removes the split-brain)*
- [ ] Self-judged close / validator edited or fixed code — *clean (validator wrote only this verdict file)*
- [ ] Scope drift (work beyond the spec, no rationale) — *clean (diff = exactly the planned 2 files; A1–A3 only; no SQL/port/view/api.gen)*
- [ ] Symptom-patch (bar moved by masking) — *clean (root cause fixed via owner's constants, not a CD-local literal alias)*

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Every F1.1 acceptance cell in `milestone.md` is met by independently-reproduced evidence; the consumer contract was honored (constants are the templates owner's, not CD-local copies — root cause fixed); regression clean (M0 guard still green, ledger untouched); zero unplanned scope; TDD disposition honestly labeled. Both dimensions (code-wise and function-wise) pass.
- Handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only on PASS, by the main session
