# Milestone 5 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-10 (re-run #2, post F5.3 remediation)  ·  **Verdict:** see C7 → **PASS**.
> Supersedes the prior FAIL verdict on this milestone. The validator judges and writes this file
> only; it never edits code, fixes findings, or flips status. The main session flips status **only
> on a PASS**.

## Inputs loaded

- `milestone.md` ✓ · `f5.1-signoff-detail/spec.md` ✓ · `f5.1-signoff-detail/evidence.md` (superseded
  banner) ✓ · `f5.2-taxonomy-restyle/evidence.md` ✓ · `f5.3-signoff-reconcile/spec.md` ✓ ·
  `f5.3-signoff-reconcile/evidence.md` ✓ · program `README.md` ✓ · governing `mission.md` (present;
  §5/§7/§8 bars quoted verbatim in `milestone.md`) ✓ · aggregate milestone diff — validator ran
  `git status`/`git log`/`git diff` on branch `claude/lucid-heisenberg-d6c94f` ✓.
- All required inputs present and readable → not blind. No fail-fast.
- **Artifact-completeness note (recorded, non-decisive — see C1):** no `plan.md` for any of the three
  features; F5.2 has no `spec.md`. All three features are verify-only / superseded closes, and
  equivalent execution-shaped content exists inline (see C1).

## Headline

The prior FAIL was driven entirely by F5.1's evidence proving against files **retired by ADR 0080**
(`SignoffDetailPage.tsx` / `ControlledDocumentDetailPanel.tsx` + their tests deleted; the
`/approvals/:documentId` route is now a redirect into the mode-adaptive workspace). The verify-only
feature **F5.3 signoff-reconcile** has now (a) annotated the F5.1 evidence as superseded with history
preserved and a pointer to current-state proof, (b) reconciled `milestone.md` objective/status to
ADR-0080 runtime truth, (c) re-proven the **current** workspace signoff surface with runnable tests
including a new `useSignoffMutation.test.tsx` re-establishing the `POST /signoff` + If-Match + 412→stale
guard, (d) removed the last orphan cockpit CSS, (e) recorded two reviewer APPROVEs on the current
surface. The signoff **function is live** (relocated, not lost); **0 backend changed**. Every
substantive protection the gate exists for is now satisfied against runtime truth.

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F5.1 | ✅ (superseded; contract honored at close 2026-06-23; its *implementation* was retired by ADR 0080 and is now correctly recorded as history, with current proof relocated to F5.3) | ✅ (via F5.3 — original gate rows are history-only behind the superseded banner; the objective is met by the current surface) | ✅ (no cockpit rebuilt; redirect preserves `?decision=`) | `f5.1/evidence.md` banner → `f5.3/evidence.md` |
| F5.2 | ✅ (`taxonomy/api/taxonomy.ts` untouched by styling; generated types consumed) | ✅ (validator re-verified: grep inline-style over `features/taxonomy` = **0**; `module.css` 0 hex, 17 `var(--…)` tokens) | ✅ (presentation-only; no behavior/contract change) | `f5.2/evidence.md`; greps in C2 |
| F5.3 | ✅ (contracts read from the live callers — `useSignoffMutation.ts` body `content_hash`/If-Match `"v{rev}"` verified against source, not guessed) | ✅ (all 5 spec acceptance rows re-verified by the validator — see C2) | ✅ (verify-only, zero behavior change; real defects deferred to chips, not inlined) | `f5.3/spec.md` + `evidence.md` |

**Consumer-contract cross-check (C1.3):** the new `useSignoffMutation.test.tsx` asserts the exact
producer shape — validator opened `useSignoffMutation.ts:41-51` and confirmed the source really sends
`content_hash: contentHash` in the body and `ifMatch: "v${revisionVersion}"`; the test mocks **only**
the `approvalApi.signoff` network boundary (`importOriginal` spread) leaving `ApprovalError` /
`SignoffError` / `mapSignoffError` as real code. This is a characterization test against real contract
logic, not a self-referential mock. Contract read from the consumer, not guessed → fail-closed honored.

**Artifact-completeness (C1.4 / missing-artifact clause) — recorded, non-decisive.** No `plan.md`
exists for F5.1/F5.2/F5.3, and F5.2 has no `spec.md`. Under C1's governing principle ("C1 binds on
artifacts, not on which skill produced them — equivalent inline output present = PASS"), this does not
fail the milestone here: all three features are **verify-only or superseded** closes with no net build
to plan, and the plan-equivalent content (files-touched table, gate commands = test strategy, ordered
acceptance) is present inline — in `f5.3/spec.md` + `evidence.md`, and for F5.2 in the `milestone.md`
F5.2 contract row + `f5.2/evidence.md`. The verify-only path is itself operator/orchestrator-recorded
(F5.2 "operator chose the verify-only close"; F5.3 "orchestrator authorized 2026-07-10"), which is the
justification for a build-`plan.md` being N/A. Carried as **process-artifact debt** (not a fail-closed
violation): a future close-out may backfill named `plan.md`/`spec.md` stubs or an explicit operator N/A.

## C2 — Gates re-run, isolated (validator ran these fresh, not trusted from transcript)

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| F5.3/all | `pnpm.cmd tsc --noEmit` (triggered pnpm install first) | **EXIT=0** | ✅ |
| F5.3 | `pnpm.cmd vitest run` over the 6 current-surface suites | **81/81** — useSignoffMutation **3**, DocumentWorkspacePage **26**, DecisionFooter **12**, WorkspaceSidebar **8**, signoffErrors **12**, InboxPage **20** (exact split matches evidence) | ✅ |
| F5.3 | `grep -rn "SignoffDetailPage.module" src` | **0** (exit 1) — orphan reference gone | ✅ |
| F5.3 | `grep -rn "SignoffDetailPage" src` | 2 hits, both **doc-comments** (`useSignoffMutation.test.tsx:3`, `documents/lib/documentSignoffDecision.ts:4`); no import/mount — no dead code | ✅ |
| F5.2 | `grep -rnE 'style=\{\{\|style="' src/features/taxonomy` | **0** (exit 1) — all inline styles gone | ✅ |
| F5.2 | `grep -nE '#[0-9a-fA-F]{3,8}' pages/TaxonomyAdminPage.module.css` | **0 hex** (exit 1); 17 `var(--…)` token refs; residual px = `max-width:1100px` + `2px` tab underline (disclosed token-coverage debt, structural not color) | ✅ |
| F5.2 | inline `style=` in `pages/TaxonomyAdminPage.tsx` | **0** (exit 1) | ✅ |
| all | `git status --short` / `git diff --stat` | working tree = docs (banner + milestone.md + f5.3 spec/evidence) + **1 new test** (`useSignoffMutation.test.tsx`) + **1 deleted orphan CSS**; **zero product `.tsx`/`.ts` behavior touched; no `*.go`** → 0 backend regression | ✅ |

All named gates re-ran green from clean state. The previously un-re-runnable F5.1 suites are correctly
retired (history-only behind the superseded banner) and **replaced** by the runnable current-surface
suites above — the C2 defect that drove the prior FAIL is cleared.

## C3 — Senior review of the aggregate milestone diff

- Working-tree diff for the remediation is minimal and clean: docs reconciliation + one real
  characterization test + one dead-asset removal. No product behavior changed (verified via `git
  status` — no product `.tsx`/`.ts` modified).
- **Split-brain — RESOLVED.** The prior "docs say cockpit / code says redirect" split is gone: F5.1
  evidence now carries a superseded banner pointing at the current surface, `milestone.md` objective +
  status describe the workspace surface, and F5.3 maps every current owner (routes.tsx redirect →
  DocumentWorkspacePage → WorkspaceSidebar → DecisionFooter → ApprovalModeFooter → ArtifactDecisionPanel
  → useSignoffMutation → signoffErrors). One fact, one source of truth.
- **Dead code:** the orphan `SignoffDetailPage.module.css` is removed; the only surviving
  `SignoffDetailPage` strings are explanatory doc-comments. No dangling import.
- Staff-engineer bar on the diff **and** on the milestone close: **met** — the close-out now describes
  the shipped system.

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Screen DoD (D2) — F5.2 | pass | tokenized `module.css` (0 hex, 17 tokens), inline-style grep 0, tsc clean, 23/23 (prior run), both reviewers APPROVE on the current page, rendered GREEN (disclosed) |
| Screen DoD (D2) — F5.3 current signoff surface | pass | current surface re-proven 81/81 runnable; both `frontend-screen-reviewer` + `frontend-code-reviewer` APPROVE **the current surface** (verdicts recorded in `f5.3/evidence.md`, independently traced the full runtime path) |
| Backend regression | pass | working tree touches no `*.go`; §3 "0 backend regressions" holds |
| M0 regression (truth / no dead stub / accurate tracker) | pass — now **repaired** | the M0 truth-accuracy concern from the prior run (docs mis-describing the `/approvals/:documentId` route) is exactly what F5.3 reconciled; tracker now matches runtime |
| M2 sacred views / taxonomy contract | pass | `taxonomy/api/taxonomy.ts` untouched by styling |
| M1/M3/M4 gates | not re-run (out of changed surface) | no M5 change touches those surfaces (frontend test + docs + dead-CSS only) |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| §8 "Taxonomy on redesign tokens, inline `style=`=0" | 11 inline styles (stale premise) | **0** inline styles; token-backed `module.css`, 0 hex | grep exit 1; root cause fixed by FE-14 tokenization, not masked — **PASS** |
| §8 "Detalhe Signoff: real API, both reviewers APPROVE, tests green" | screen never built (prior run: on-record APPROVE was for a **deleted** cockpit → not certifiable) | current workspace signoff surface (approving mode) proven by **81/81** runnable tests incl. the `POST /signoff` + If-Match guard; **both reviewers APPROVE the current surface** | root cause ("screen never built / then relocated, never re-certified") fixed by reconciling to the live surface + runnable proof, **not** masked by stale evidence — **PASS** |

- Could it be built better? The ADR-0080 consolidation into one mode-adaptive workspace is the
  better construction, and the milestone now documents it rather than riding on stale cockpit
  evidence. Residual improvement (already disclosed, deferred to chips, not blocking): direct unit
  coverage for `ArtifactDecisionPanel` and a transport-layer `approvalApi.signoff` test; plus the
  process-artifact `plan.md`/`spec.md` backfill from C1. None indicate an unsound current construction.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clear: every
  acceptance row mapped to a re-run command/output above.*
- [ ] Fixture/mock passed off as real-provider proof — *clear: the live-e2e signoff gap (empty dev DB)
  is honestly disclosed; the contract is proven by a characterization test that exercises real
  `content_hash`/If-Match/412→stale code paths and is labeled fixture-level, matching the F5.1 spec
  which explicitly accepted fixture proof for these rows.*
- [ ] Consumer contract guessed rather than read from the consumer — *clear: verified against
  `useSignoffMutation.ts` source.*
- [ ] Split-brain — *RESOLVED by F5.3 (see C3).*
- [ ] Self-judged close / validator edited or fixed code — *clear: validator only judged; wrote this
  verdict file only.*
- [ ] Scope drift — *clear: F5.3 is verify-only, zero behavior change; real defects pushed to chips.*
- [ ] Symptom-patch — *clear: stale-docs root cause fixed by reconciliation to runtime truth, not
  masked.*

(All unchecked = clean. The two hits that produced the prior FAIL — fixture/record-as-current-proof and
split-brain — are both cleared.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass. **Code-wise:** clean minimal diff (one real characterization test + one
  dead-asset removal), no split-brain, no dead code, contracts read from source. **Function-wise:** the
  current workspace signoff surface renders/gates/records decisions via `useSignoffMutation` proven by
  81/81 runnable tests + two reviewer APPROVEs on the current surface; Taxonomy Admin is fully
  tokenized (inline-style grep 0, 0 hex). The prior FAIL's failed checks (C2 deleted-test re-run, C4/C5
  deleted-screen certification, C6 fixture-as-current + split-brain) are all cleared by F5.3.
- **Recorded non-blocking debt** (does not block PASS): missing `plan.md` (all features) + F5.2
  `spec.md` — process-artifact debt justified by the verify-only/superseded nature (C1); plus disclosed
  coverage chips (`ArtifactDecisionPanel`, transport-layer `approvalApi.signoff`). Carry these forward
  to the program close-out / a coverage chip, not a milestone blocker.
- Handed back to the main session to flip status (`README.md` M5 → passed on PASS) and present the
  **HS-1** operator gate. M5 is the **last** milestone — on HS-1 approval the program proceeds to
  terminal acceptance (`mission-validator`), not another milestone.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending.
> - Status flipped in `README.md`: pending — only on PASS + HS-1.
