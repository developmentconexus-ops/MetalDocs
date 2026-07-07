# F4 — Plan

Seeded from master plan §F4 + design spec §8 C2 + the F4 investigator map (runtime truth). TDD via a
fresh implementer subagent (sonnet) + independent reviewer subagent. **Implementer uses its OWN
tools; does NOT spawn sub-agents.**

## Ordered tasks

1. **[TDD] Failing tests first** (author all before impl; run RED):
   - `lib/format/__tests__/formatDueRelative.test.ts`
   - `features/approval/components/sidebar/__tests__/ApprovalSidebar.test.tsx`
   - `features/approval/components/sidebar/__tests__/DecisionFooter.test.tsx`
   Cases exactly per spec Validation Gate.
2. **`formatDueRelative`** in `lib/format/dates.ts` — `(dueAt, now = Date.now()) => { label, overdue }`
   per spec §5. Pure; injectable `now`.
3. **`sidebar/StageContextHeader.tsx`** — props `{ instance, activeStage, lockedByInstanceId }`.
   Renders "Etapa N de M", stage label, pool (`activeStage.actors[].display_name`, truncate "+K"),
   due chip (`formatDueRelative(activeStage.due_at)`, overdue style), quorum (derived §4: approval
   stage → "N de M aprovaram" from `actors[].status==='approved'`; review → pool count only; hide if
   no actors), and `<LockBadge>` when `lockedByInstanceId`.
4. **`sidebar/ApprovalTimeline.tsx`** — reuse `ApprovalTimelinePanel` as the base (import & wrap, or
   fork its render adding `signature_meaning`). Single timeline. Add the per-signoff
   `signature_meaning` label ("Aprovação"/"Rejeição"). No second band.
5. **`sidebar/IntegrityDisclosure.tsx`** — props `{ documentId, contentHash, frozenContentHash }`.
   `<details>` collapsed summary "Conteúdo verificado ✓ · detalhes"; expanded shows
   `frozenContentHash ?? contentHash` + `etagCache.get(documentId)`, each with an inside copy button
   (reuse the `navigator.clipboard.writeText` + "Copiado" 1200ms pattern from
   `DocumentApprovalExtras.tsx:55-63`). Hash absent from DOM while collapsed (render children only
   when open, or rely on `<details>` not mounting — assert collapsed-not-visible in test via
   `queryByText` on the closed element; prefer conditional render on `open` state for a hard guarantee).
6. **`sidebar/SuggestionList.tsx`** — props `{ documentId, trackedChanges }`. Cards: author, type,
   excerpt from each `TrackedChange`. Below: `useDocumentCommentsQuery(documentId)` list via
   `commentPlainText`, author + `formatSignedAt(created_at)` + resolved tag. Empty state PT-BR.
7. **`sidebar/DecisionFooter.tsx`** — props `{ isApprovalStage, decision, actions, instance,
   activeStage, onRefetchInstance }`. Sticky footer.
   - review (`!isApprovalStage`): "Pronto para aprovação" (primary) + "Solicitar mudanças" (opens a
     dialog: mandatory comment textarea, submit disabled until non-empty). Both call
     `useReviewVerdictMutation().mutateAsync({ instanceId: instance.id, stageId: activeStage.id,
     etag: instance.etag, body: { verdict, comment } })`; on success `onRefetchInstance()`; map
     problem+json (422) inline (not a toast).
   - approval (`isApprovalStage`): `<ArtifactDecisionPanel model={decision} />` (reuse). The
     meaning-of-signature line (§6) is injected via the decision model's description/legal — see
     task 8.
   - Both variants: render `actions` (publish/cancel) as a secondary "Outras ações" button group.
8. **Meaning-of-signature line (§6, conservative default)** — in `documentSignoffDecision.ts` (the
   decision-model builder), add a decision-derived meaning sentence to the model's `description` (or
   a new sub-line the panel already renders) — approve → "Ao assinar, você declara aprovação deste
   documento." / reject → "…declara rejeição…". **Do NOT touch the existing MP 2.200-2 `legal.text`.**
   Keep the change minimal + reversible. (If the builder can't be cleanly extended without touching
   the template path, add the line in `DecisionFooter` above the reused panel instead — pick the
   path that does not alter the template signoff.)
9. **`sidebar/ApprovalSidebar.tsx`** — composition root per spec §1: flex column, `overflow-y:auto`
   scroll region + sticky `DecisionFooter`. Order: StageContextHeader → ApprovalTimeline →
   IntegrityDisclosure → (review ? SuggestionList : null) → DecisionFooter. CSS module with wine
   tokens only.
10. **Rewire `ApprovalCockpitPage.tsx`**: `screenModel = { ...model, decision: null,
    approvalChain: null, actions: [] }`; pass `<ApprovalSidebar … />` as `decisionExtras` with the
    real `decision`/`actions`/`instance`/`activeStage`/`trackedChanges`/`contentHash`/
    `frozenContentHash={instance?.frozen_content_hash ?? null}`/`isApprovalStage`. Keep the
    loading/context-error/no-active-context guard arms (render in place of the sidebar). Remove the
    `DocumentApprovalExtras` import + its else-arm.
11. **Delete** `features/approval/components/DocumentApprovalExtras.tsx` (+ `.module.css` if now
    unused; check no other importer via grep first).
12. **Verify:** `grep 'DocumentApprovalExtras' frontend/apps/web/src/features/approval` → zero;
    `grep 'Dados podem estar desatualizados' frontend/apps/web/src` → zero; `grep
    'useDocumentSession|useDocumentAutosave' …/features/approval` (excl mocks) → still zero. Targeted
    `vitest run` GREEN (new sidebar tests + `ApprovalCockpitPage.test.tsx` updated); `tsc
    --noEmit -p tsconfig.build.json` clean.
13. **Review pass** — independent reviewer subagent (sonnet): spec §1-§6 compliance, single-timeline,
    shared-component-untouched (`git diff ArtifactApprovalScreen.tsx` empty), signoff-reuse (no
    rebuild), no stale banner, wine tokens, 21 CFR/MP flag honored (MP 2.200-2 untouched), test
    non-tautology. Apply accepted findings.

## Files touched

- `frontend/apps/web/src/lib/format/dates.ts` (add `formatDueRelative`)
- `frontend/apps/web/src/lib/format/__tests__/formatDueRelative.test.ts` (new)
- `frontend/apps/web/src/features/approval/components/sidebar/ApprovalSidebar.tsx` (new) + `.module.css`
- `.../sidebar/StageContextHeader.tsx` (new)
- `.../sidebar/ApprovalTimeline.tsx` (new)
- `.../sidebar/IntegrityDisclosure.tsx` (new)
- `.../sidebar/SuggestionList.tsx` (new)
- `.../sidebar/DecisionFooter.tsx` (new)
- `.../sidebar/__tests__/ApprovalSidebar.test.tsx` + `DecisionFooter.test.tsx` (new)
- `frontend/apps/web/src/features/approval/pages/ApprovalCockpitPage.tsx` (rewire)
- `frontend/apps/web/src/features/approval/pages/ApprovalCockpitPage.test.tsx` (update integrity assertions)
- `frontend/apps/web/src/features/documents/lib/documentSignoffDecision.ts` (meaning line, §6 — minimal)
- **Deleted:** `DocumentApprovalExtras.tsx` (+ `.module.css` if unused)

## Risks

- **Shared component regression** — the cockpit must neutralize bands via null model fields, NOT by
  editing `ArtifactApprovalScreen`. Guardrail: `git diff ArtifactApprovalScreen.tsx` MUST be empty.
- **Signoff flow regression** — reusing `ArtifactDecisionPanel` keeps password/legal/options intact.
  Do not fork it. The meaning line must not disturb the existing MP 2.200-2 legal text.
- **Timeline reuse** — if `ApprovalTimelinePanel` can't be cleanly wrapped, fork its render into
  `ApprovalTimeline` but keep behavior identical + add `signature_meaning`; delete the old panel only
  if it has no other importer (grep — `DocumentApprovalExtras` was its consumer; after deletion it
  may be orphaned → then move its internals into `ApprovalTimeline`).
- **`<details>` hash-visibility test** — jsdom renders `<details>` children regardless of `open`.
  Guarantee the collapsed-hidden assertion by conditionally rendering the hash only when `open`
  state is true (controlled `<details>` via `onToggle`), so `queryByText(hash)` is null when closed.
- **junction drift** — if vitest won't run, real fix = full `pnpm install` in `frontend/apps/web`;
  do NOT hack around it. Report the exact error.
- **21 CFR/MP 2.200-2** — conservative default only (MP untouched + derived meaning line); the
  jurisdiction choice is the operator's at HS-1. Do not append a 21 CFR citation.
