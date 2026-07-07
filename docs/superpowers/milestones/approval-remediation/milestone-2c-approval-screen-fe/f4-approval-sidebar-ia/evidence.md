# F4 — Evidence

## Commands + real output

- `cd frontend/apps/web && npx vitest run src/features/approval src/features/documents src/lib/format`
  → **Test Files 51 passed (51) · Tests 340 passed (340)**, 0 failed. Includes the 4
  directly-authored/updated suites: `formatDueRelative.test.ts` (8), `DecisionFooter.test.tsx` (9),
  `ApprovalSidebar.test.tsx` (5), `ApprovalCockpitPage.test.tsx` (14) + zero regressions across the
  rest of both features. Re-run independently by the reviewer from clean state — same result.
- `npx tsc --noEmit -p tsconfig.build.json` → **clean, zero output** (implementer + reviewer both).
  Two real type errors caught + fixed during the pass: (1) `ApprovalCockpitPage.tsx:196`
  `screenModel.decision` set to `null` but `ArtifactViewModel.decision` is `ArtifactDecisionModel |
  undefined` → fixed to `decision: undefined`; (2) `DecisionFooter.test.tsx` passed a stale
  `isApprovalStage` prop at 10 sites → stripped.
- `grep -rn "DocumentApprovalExtras" src/` → **ZERO** (main-session verified). Component +
  `.module.css` deleted; 2 stale doc-comment mentions (`approvalWorkflow.ts`, `tokens.css`) also
  cleaned. `TemplateApprovalExtras` left untouched (different, still-live component).
- `grep -rni "dados podem estar desatualizados" src --include=*.tsx --include=*.ts | grep -v .test.`
  → **ZERO non-test matches** (main-session + reviewer verified). The single match is the *negative*
  assertion (`expect(...).toBeNull()`) proving the stale banner does NOT render.
- `git diff --stat -- .../shared/controlled-artifact/ArtifactApprovalScreen.tsx` → **EMPTY** — shared
  template/document component never edited (main-session + reviewer verified by reading the gating
  logic, not just trusting the diff).
- `git diff .../documents/lib/documentSignoffDecision.ts` → **EMPTY** — MP 2.200-2/2001 `legal.text`
  untouched (§6 conservative default honored).

## TDD proof

- Implementer confirmed RED-first — all three authored tests failed before implementation:
  `TypeError: formatDueRelative is not a function`; `Cannot find module '../DecisionFooter'`;
  `Cannot find module '../ApprovalSidebar'`. GREEN after implementing. Tests authored first.

## Runtime proof (observable change) + fixture-vs-real

- **Single timeline (the C2 point):** the cockpit now renders exactly ONE approval timeline. The
  shared `ArtifactApprovalScreen` bands are suppressed by passing `decision: undefined,
  approvalChain: null, actions: []` — its `decision`/`hasChain`/`hasActions` render gates all fall,
  leaving only `decisionExtras` (the new `ApprovalSidebar`) in the `aside`. Reviewer confirmed by
  reading the gating logic in the shared component (not trusting the neutralization comment).
  `ApprovalSidebar.test.tsx` asserts `getAllByRole('region', {name:/timeline de aprovação/i})`
  length === 1 — runtime-proven by render. Labeled **fixture/mock** (vitest + jsdom).
- **Integrity behind disclosure, hash not leaked:** `IntegrityDisclosure` is a controlled
  `<details open={open} onToggle>` that conditionally renders the hash body (`{open && …}`) — the
  hash is absent from the DOM tree while collapsed (not merely visually hidden). Both the component
  test and the `ApprovalCockpitPage` regression assert `queryByText(hash)` null pre-expand, present
  post-`fireEvent.click` + `waitFor` — runtime-proven, not fixture-asserted.
- **Signoff reuse, not rebuild:** `DecisionFooter` approval mode renders `<ArtifactDecisionPanel
  model={decision} />` verbatim (shared import) — password re-auth + legal + options intact. No
  forked password/legal UI (grep confirms `ArtifactDecisionPanel` is the sole decision surface).
- **Review verdict wiring:** `DecisionFooter` review mode calls
  `useReviewVerdictMutation().mutateAsync({instanceId, stageId, etag, body:{verdict, comment}})` with
  exact args (asserted in `DecisionFooter.test.tsx`); "Solicitar mudanças" submit disabled until the
  comment textarea is non-empty (disabled-state transition asserted).
- **Quorum derived (no DTO field):** `StageContextHeader.formatQuorum` → "N de M aprovaram" from
  `activeStage.actors[].status==='approved'`; review stage → pool count only; returns `null` when
  `actors` empty (no "0 de 0" leak). Matches spec §4.
- **W2 not regressed (F3 invariant held):** `ApprovalCockpitPage` still passes `DocumentShell` no
  `onAutoSave`; `useDocumentSession`/`useDocumentAutosave` appear only as test mocks. Grep confirms.
- **Real end-to-end** (live approval-stage cockpit sidebar + review-stage suggest+verdict against the
  running stack) is exercised in the **F8 live-QA walkthrough** — registered there, not claimed here.

## Key design decisions (verified against runtime truth)

- **Shared component neutralized, never edited.** Cockpit kills the duplicate "Fluxo de aprovação"
  band by nulling the model fields the shared screen gates on — `ArtifactApprovalScreen` stays
  byte-identical (empty diff), so the template approval screen is untouched.
- **Reuse `ArtifactDecisionPanel` for signoff** — no rebuilt password/legal form.
- **Meaning-of-signature line placed in `DecisionFooter` (`MeaningOfSignatureLine`), not the
  decision-model builder.** Plan task 8 pre-authorized this fallback ("If the builder can't be
  cleanly extended… add the line in DecisionFooter instead"). Better isolation: zero diff on
  `documentSignoffDecision.ts`, which both document + template flows share. Approve → "declara
  aprovação", reject → "declara rejeição". No regulation citation appended.
- **`formatDueRelative`** new pure util in `lib/format/dates.ts`, `now` injectable (deterministic
  tests). null→`—`; same-day→"vence hoje"; future→"vence em N dia(s)"; past→"atrasado há N dia(s)"
  overdue:true.

## Review / QA disposition

- Independent reviewer subagent (separate from implementer, own tools, no edits): **APPROVE**,
  **0 Critical, 0 Major, 3 Minor**. All 12 adversarial checks PASS, each runtime-proven (executed
  tests + direct reads of gating logic, not taken on faith from comments). Reviewer re-ran vitest
  (340/340) + tsc (clean) itself from clean state.
- **3 Minor findings (all accepted as non-blocking, no fix):**
  1. `ApprovalSidebar.tsx` / `DecisionFooter.tsx` use inline `style` for `overflowY:auto` /
     `position:sticky` alongside the CSS module — cosmetic duplication of layout intent; tests target
     `.style.*` so fixture-consistent. Optional cleanup.
  2. `formatDueRelative` same-day check uses local-tz calendar fields against a UTC epoch `now` —
     inherited pre-existing pattern (all `dates.ts` formatters are implicitly local-tz), not a new
     defect; midnight-rollover boundary is env-tz-dependent. Awareness only.
  3. `ApprovalTimeline.tsx:60` doc-comment "Forked from the former `ApprovalTimelinePanel`…" reads
     like a live dependency — stale phrasing, harmless. No action.

## §6 flag (surface at HS-1, unresolved by design)

FE legal text (`documentSignoffDecision.ts`) cites **MP 2.200-2/2001** (Brazilian ICP-Brasil);
backend `signature_meaning` doc-comment cites **21 CFR 11.50(a)(3)**. NOT reconciled by F4 — the
conservative default keeps MP 2.200-2 untouched and adds only a decision-derived meaning supplement.
The jurisdiction choice is the operator's decision at the M2c HS-1 gate.

## Bounded defers

- None new. Durable server-authoritative suggestion persistence remains the F0/HS-2 program-level
  bounded defer (client-authoritative track-changes rendered by `SuggestionList` + comments +
  freeze-hash chain cover today).
- Minor #1 (inline-style vs CSS-module dedup) and Minor #2 (local-tz boundary in `formatDueRelative`)
  logged as cosmetic/awareness; no trigger — pick up opportunistically in F7 (visual + a11y polish).
