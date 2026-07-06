# Milestone 9 — Governance & Hygiene Close-out

> **Program:** global-maximum-remediation  ·  **Governing spec:** `../mission.md` §7 M9
> **Status:** validator PASS (2026-07-06) — HS-1 operator gate pending
> **Authored:** 2026-07-06 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** M9 is, **which features** it contains,
> **what each feature implements**, and **what gets validated**. No execution steps — the "how" lives
> in each feature's `plan.md`. The close QA (`qa/milestone-qa.md`) validates M9 against *this* file and
> the binding `validation-contract.md` (D4).

## Objective

The **last** milestone of the mission: governance debt and structural hygiene cleared so the codebase's
*paper truth matches its runtime truth* and stays that way mechanically (findings 19–25 + D6). After M9:

- A reader opening any ADR sees its **decision state in ≤3 status lines** — never a changelog crammed
  into the status field — and the supersession chain (0013 included) is complete and navigable.
- A **CI-executable traceability gate** turns RED when a MUST-classified REQ ID in
  `wiki/architecture/backend-target-architecture.md` has zero citing test/commit — the convention
  becomes enforcement (the mission's meta-defect class: hand-synced truths).
- A maintainer hitting a broken legacy test has a **written decision procedure** (repair vs delete) in
  `wiki/quality`, replacing tribal memory; integration test wall-clock is **measurably** reduced via
  `t.Parallel` expansion (before/after numbers recorded).
- **CLAUDE.md tells the truth**: module inventory matches `internal/modules/` on disk (−`docs`,
  +`tokens`), idempotency described where it actually runs, janitor/scheduler wording matches the
  post-M5 River runtime; mission-touched wiki docs re-stamped.
- The backend has **one persistence-layer naming convention** (`infrastructure/`), and
  `documents/approval` — today a structural 15th module invisible to boundary guards — is either
  promoted or its exception is ADR-recorded, with boundary guards covering it either way.

**Bars moved:** ADR-status legibility (0 ADRs with >3-line status); traceability enforcement
(convention → RED/GREEN CI gate with negative proof); test-policy governance (unwritten → published
taxonomy); doc truth (CLAUDE.md drift = 0 vs disk); structural coherence (2 naming dialects → 1;
0 unguarded hidden modules).

## Appetite & rabbit holes

**Appetite:** 5 features, governance/docs/test/structure only. **No product behavior changes, no
schema migrations, no API contract changes, no new binaries.** Risk-isolating by design (D5): nothing
here may regress M0–M8 bars. Only F9.5 touches production Go code paths (mechanical package moves).

**Rabbit holes (do not chase):**
- **No If-Match/lock_version unification** (ADR 0066 named M9 a "candidate") — it is a cross-module
  wire-contract change, not hygiene; stays deferred with ADR 0066's own trigger. Touching it here
  violates the risk-isolation premise of M9.
- **No rewriting ADR bodies** beyond the status-field split — F9.1 relocates history, it does not
  re-litigate decisions.
- **No new REQ IDs / no re-grading MUST↔SHOULD** in the target architecture doc — F9.2 enforces the
  existing set; editing the normative doc is out of scope (exception: mechanical citation-anchor fixes
  if the gate needs them, recorded).
- **No mass test rewrites** — F9.3 writes policy + adds `t.Parallel` where safe; it does not migrate
  legacy tests to the canonical framework (that stays governed by the existing hard gate + new policy).
- **No repository-interface redesign in F9.5** — the rename is mechanical (package path/name); zero
  behavior or signature changes beyond what the move forces.
- **No frontend scope** — all findings 19–25 are backend/governance.
- **No Terminal Acceptance work inside M9** — mission §8 runs only after M9 passes HS-1, dispatched
  separately.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F9.1 | `f9.1-adr-hygiene` | Split ADR 0022's 2.7k-char status field: status ≤3 lines, execution history relocated to a linked companion doc (content preserved, not summarized away); same treatment for any other ADR whose status exceeds the rule (0027 flagged at 665 chars — audit all 70). Stamp ADR 0013 `Superseded` with a link to its superseding decision (chain researched, not guessed). Write the ADR status-field rule into `wiki/standards/documentation-governance.md`. **Consumer:** any maintainer/reviewer opening an ADR to learn its current state; docs-governance checks. | Scripted sweep over `wiki/decisions/*.md` shows every `> **Status:**` block ≤3 lines; 0013 reads `Superseded by …` and the superseding ADR back-links; 0022 companion doc exists, is linked from the status, and preserves the relocated history; governance doc states the rule; no ADR body decision content altered (diff review). |
| F9.2 | `f9.2-traceability-gate` | REQ-ID→evidence traceability automation for the 67 REQ IDs in `wiki/architecture/backend-target-architecture.md`: a gate (script + CI job, e2e-coverage-gate pattern) that extracts MUST-classified REQ IDs and fails when one has zero citing test file or commit; a committed map/report of current coverage. Baseline fact: ~10 REQ citations exist in `*_test.go` today — the gate's initial green must come from a legitimate evidence map (tests + commits + satisfied-by annotations already present in the doc), never from weakening the rule. **Consumer:** CI (`.github/workflows/`), reviewers citing REQ IDs per CLAUDE.md. | POSITIVE proof: gate green on clean tree with the committed map. NEGATIVE proof: planting a fake MUST REQ (or removing a citation) turns the gate RED — both runs recorded. Gate wired in CI; map regenerable by command. |
| F9.3 | `f9.3-test-policy` | Legacy-test deletion taxonomy published in `wiki/quality` (repair-class: REQ/tripwire/contract/invariant guards; delete-class: one-off task scaffolding; decision procedure + examples), codifying the existing memory rule. `t.Parallel` expansion across integration test files where isolation allows (baseline: 12 of 386 `_test.go` files under `internal/` call `t.Parallel`). Wall-clock impact **measured** on a named representative package set (full suite not run locally — box constraint). **Consumer:** maintainers triaging broken tests; CI wall-clock budget. | Policy doc exists in `wiki/quality` and is linked from `test-discipline.md`/index; before/after wall-clock numbers for the named package set recorded in evidence (same command, same box); expanded files still green; no test semantics changed (parallel-unsafe files left serial with reason noted). |
| F9.4 | `f9.4-doc-truth` | CLAUDE.md corrections to match runtime: module inventory (−`docs`, +`tokens`, count re-verified against `internal/modules/`); idempotency described per-handler (not a blanket chain link) if runtime confirms; janitor/scheduler wording matched to post-M5 reality (River-hosted schedules vs api-hosted leader-elected janitors — verified against binaries' main wiring). Wiki `Last verified` stamps refreshed for mission-touched docs via wiki-curator pass. **Consumer:** every future agent/maintainer session bootstrapped from CLAUDE.md. | Each corrected CLAUDE.md claim paired with its runtime-truth evidence (file:line or command output) in the feature evidence; `internal/modules/` listing matches inventory exactly; wiki-curator pass reports clean (stamps current, anchors valid) for the mission-touched doc list enumerated in the contract. |
| F9.5 | `f9.5-structure-hygiene` | One persistence-layer naming convention: rename `repository/` → `infrastructure/` in documents + templates (templates currently has BOTH dirs — fold correctly, not blindly); decide `documents/approval` (a full-layer hidden 15th module: api/application/domain/http/infrastructure/jobs/repository) via a **mini `developing-new-work` gate** → promote to top-level module OR record an ADR exception; extend module-boundary guards to cover approval either way. Mechanical moves only — no signature/behavior changes. **Consumer:** module-boundary CI guard (`scripts/check-module-boundaries.ps1`), future module work relying on layout conventions. | Zero `repository/` dirs remain under `internal/modules/` (or every survivor is ADR-recorded); approval decision documented (gate verdict + ADR if exception path); boundary guard demonstrably covers approval (negative proof: a planted violation is caught); `go build ./...` + `go test ./...` (targeted per box constraint) + `scripts/check-module-boundaries.ps1` green. |

> **Amendment 2026-07-06 (F9.2 row):** "POSITIVE proof: gate green on clean tree" is governed by
> `validation-contract.md` **Erratum E1** (operator-approved HS-7 re-open): first-population positive
> proof = anti-rot-clean run whose uncovered-MUST set equals exactly the 4 E1-ledgered defers
> (AUTHN-1, AUTHN-3, SEARCH-1, SEC-3 — genuine doc-falsity/coverage gaps, bounded with triggers).
> Gate strictness itself unchanged.

Order intentional: F9.1/F9.2 first (pure-docs + gate, zero code risk, and F9.2's gate becomes part of
the mission's terminal gate inventory); F9.3 next (policy + low-risk test-only edits); F9.4 after
F9.3/F9.5 facts exist but before close (doc truth reflects final state — its CLAUDE.md/wiki text must
describe post-F9.5 layout, so its **final verification** runs after F9.5; sequencing detail lives in
its plan); F9.5 last (only feature touching production code layout; its build/boundary gate doubles as
the milestone-close code gate).

## Milestone validation definition

Close gate run by the **`milestone-validator` subagent** (separation of powers — judges and writes
`qa/milestone-qa.md`; main session flips status only on PASS), per the binding C1–C7 checklist
(`.claude/skills/milestone/references/milestone-end-validation.md`). For M9:

1. **Per-feature acceptance** — every feature meets its "what to validate"; each feature's consumer
   contract (`spec.md`) honored. Checked **section-by-section against `validation-contract.md`** (D4)
   — any divergence is **HS-7**.
2. **Workflow-class QA** — docs governance (`wiki/standards/documentation-governance.md`) for
   F9.1/F9.3/F9.4 artifacts; `wiki/quality/test-discipline.md` conformance for F9.3's touched tests;
   backend structural checks for F9.5.
3. **Regression** — M0–M8 gates still pass from clean state: `go build ./...`,
   `scripts/check-module-boundaries.ps1`, api-lint blocking gate, targeted test suites named in the
   contract §6; **no M0–M8 evidence artifact or gate weakened by M9 edits** (risk-isolation premise).
4. **Root-cause check** — F9.2's gate kills the hand-sync class for REQ traceability (not a one-off
   report); F9.5 removes the naming split structurally (not a lint suppression); F9.1's rule is
   enforceable going forward (documented + sweepable), not a one-time cleanup.
5. **No unplanned scope** — anything beyond F9.1–F9.5 recorded with rationale; rabbit-hole list above
   is the drift baseline.

## Dependencies & constraints

- **Depends on:** M0–M8 all passed (operator HS-1 through M8, 2026-07-06). M9 is last; nothing depends
  on it except mission Terminal Acceptance (§8), which is **not** part of M9.
- **Quality goals (ranked):** truth-preservation (docs/ADR content never lost or silently rewritten) >
  regression-isolation (M0–M8 bars untouched) > mechanization (rules become gates).
- **Architectural constraints:** no OpenAPI/contract changes; no migrations; no capability/authz
  changes; modular-monolith boundaries hold (F9.5 *strengthens* REQ-TOP-1 guards, never relaxes);
  ADR/wiki edits follow docs governance; commits local, **never push**; never read/print/commit
  `.env`; full integration suite NOT run locally — targeted `-run`/package filters with bounded
  defers; subagent policy sonnet/haiku, never fable, ≤15 concurrent.
- **Risks:** (1) F9.5 rename breaks hidden import cycles or generated-code references → mitigation:
  mechanical move via compiler-guided sweep, full `go build ./...` gate, boundary check;
  (2) F9.2 gate too strict on day one (67 REQs, ~10 test citations) → mitigation: MUST-only scope +
  evidence map includes commits/doc-recorded satisfactions, negative proof keeps it honest;
  (3) F9.1 history relocation loses audit trail → mitigation: content moved verbatim to companion doc,
  diff-reviewed; (4) approval promotion churns imports across modules → mitigation: mini gate may
  legitimately choose the ADR-exception path — promotion is not presumed.

## Applicable hard-stops

- **HS-1** (milestone boundary) — after validator PASS, STOP; operator gate; **Terminal Acceptance
  (mission §8) does not start without explicit operator go.**
- **HS-2** (fix implies redesign outside boundary) — e.g. approval promotion turning into a cross-module
  authz or contract redesign; stop, surface, no symptom-patch.
- **HS-3** (prerequisite boundary fails) — build/boundary-script/CI-runner broken → repair first.
- **HS-4** (validator FAIL) — open named fix feature, re-run lifecycle, re-dispatch validator.
- **HS-6** (scope drift) — e.g. ADR sweep uncovering decision-content contradictions (not just status
  bloat), or F9.4 runtime check contradicting a mission-era doc: stop, surface, replan.
- **HS-7** (impl deviates from `validation-contract.md`) — fix work to contract, or re-open contract
  WITH operator approval; never silently edit contract to match work.
- **HS-8 analog (F9.5 mini gate)** — if the mini `developing-new-work` gate returns Red on approval
  promotion, the promotion is blocked; the ADR-exception path or operator escalation follows.
