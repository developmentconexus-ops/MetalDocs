# F4 — Approval sidebar IA (C2)

> **Milestone:** M2c approval-screen-fe · **Consumer:** the approval cockpit
> (`ApprovalCockpitPage`), which replaces its `decisionExtras` slot + the shared screen's
> decision/flow rendering with a single self-contained `ApprovalSidebar`.
> **Status:** Approved — 2026-07-07. Approval line below.

## Problem (runtime truth today)

The cockpit aside is cluttered and duplicated: `ArtifactApprovalScreen` renders **two** "Fluxo de
aprovação" bands (`ArtifactApprovalScreen.tsx:174-179` decision branch, `:194-199` fallback branch),
`DocumentApprovalExtras` renders a second integrity block + a "Dados podem estar desatualizados"
stale banner (`:99-106`) driven by the old polling adapter, and the decision CTA is not pinned.
Spec §8 **C2** requires: **stage context → single timeline → integrity collapsed to an auditor
disclosure → decision CTA pinned at the footer**, sidebar scrolls internally (page doesn't), stale
banner dies (react-query invalidation from F2/signoff replaces polling).

`ArtifactApprovalScreen` is **shared with the template approval screen** — F4 must NOT edit it to
strip bands globally (that would regress templates, out of M2c scope).

## Consumer contract (what downstream requires, defined before producer)

### 1. `ApprovalSidebar` — the self-contained document-approval aside (new)

`features/approval/components/sidebar/ApprovalSidebar.tsx` (composition root). Container is a flex
column: **`overflow-y: auto`** on the scroll region, **`DecisionFooter` pinned `position: sticky;
bottom: 0`**. Renders top→bottom:

1. **`StageContextHeader`** — "Etapa N de M" (`activeStage.stage_index + 1` / `instance.stages.length`),
   stage `label`, **pool** (the active stage `actors[].display_name`, e.g. "Ana, Bruno +2"),
   **due chip** (relative PT-BR from `due_at`, see §5), **quorum** (derived — see §4), and the
   **lock badge** (reuse `LockBadge` when `lockedByInstanceId` set — the concurrent-edit signal,
   not lost).
2. **`ApprovalTimeline`** — the **single** timeline. Reuse `ApprovalTimelinePanel` internals as the
   base (submit event → per-stage signoffs → final status), **add** the `signature_meaning` label
   per signoff (`approval`→"Aprovação", `rejection`→"Rejeição"). No second band anywhere.
3. **`IntegrityDisclosure`** — collapsed `<details>`: summary "Conteúdo verificado ✓ · detalhes";
   expanded reveals `frozen_content_hash` (fallback to the active-context `content_hash` when the
   instance isn't frozen yet), `etag` (from `etagCache.get(documentId)`), each with a copy button
   **inside** the disclosure. Hash NOT visible while collapsed.
4. **`SuggestionList`** (review mode only) — cards from the `trackedChanges` prop (author, type,
   excerpt); below them, the document comments via `useDocumentCommentsQuery` + `commentPlainText`
   (author, relative/short date, resolved tag).
5. **`DecisionFooter`** (sticky) — **mode-aware**:
   - **review mode** → "Pronto para aprovação" (primary) + "Solicitar mudanças" (opens a dialog
     with a **mandatory** comment; submit disabled until non-empty; maps the 422 problem+json).
     Both call `useReviewVerdictMutation` with `{ instanceId: instance.id, stageId: activeStage.id,
     etag: instance.etag, body: { verdict: 'ready' | 'request_changes', comment } }`. **No** signoff
     button, **no** password field.
   - **approval mode** → renders the **existing** `ArtifactDecisionPanel model={model.decision}`
     (password re-auth + legal statement + decision options — the tested signoff flow, reused not
     rebuilt). **No** verdict CTAs, **no** suggestion cards.
   - Publish / cancel lifecycle actions (`model.actions`) render as a secondary "Outras ações"
     group in the footer region (not lost — they move here from `ArtifactApprovalScreen`'s band).

Props (shape, not final):
```tsx
interface ApprovalSidebarProps {
  editorMode: 'review' | 'readonly';        // 'readonly' here == approval/observer (from cockpit)
  isApprovalStage: boolean;                 // active stage_kind === 'approval' (decide footer variant)
  instance: ApprovalInstance;
  activeStage: StageInstance | undefined;
  documentId: string;
  contentHash: string | null;
  frozenContentHash: string | null;
  decision: ArtifactDecisionModel | null;   // model.decision for approval-mode signoff
  actions: ArtifactAction[];                 // publish/cancel etc.
  trackedChanges: TrackedChange[];           // review-mode suggestion cards
  onRefetchInstance: () => Promise<void> | void;
}
```

### 2. Cockpit rewiring (`ApprovalCockpitPage.tsx`)

- Build `screenModel` with **`decision: null`, `approvalChain: null`, `actions: []`** so
  `ArtifactApprovalScreen` renders **no** decision head, **no** flow band, **no**
  `ArtifactDecisionPanel`, **no** actions band — its `aside` contains ONLY our `ApprovalSidebar`
  (passed as `decisionExtras`). This kills the duplicate timeline **for the document cockpit
  without touching the shared component** (templates unaffected).
- Pass the real `decision`/`actions` into `ApprovalSidebar` (which now owns their rendering).
- `isApprovalStage = activeStage?.stage_kind === 'approval'`; `editorMode` from the existing
  `resolveEditorMode`. Keep the loading/context-error/no-active-context guard branches (they render
  in place of the sidebar body, as today).
- **Delete** `DocumentApprovalExtras.tsx` (fully absorbed: integrity→`IntegrityDisclosure`,
  timeline→`ApprovalTimeline`, lock→`StageContextHeader`, stale banner→**deleted**, state badge→
  `StageContextHeader`). Remove its import + the four-way `decisionExtras` branch's else-arm.

### 3. Stale banner dies

No "Dados podem estar desatualizados" anywhere. Freshness comes from react-query invalidation
(`useReviewVerdictMutation`/`useSignoffMutation` already invalidate `QK.approval.all` +
`QK.documents.all`). `isStale` is no longer rendered.

## Derivations & decisions (non-DTO facts)

### 4. Quorum is **derived**, not a DTO field

`ApprovalStageInstanceResponse` has **no** quorum/threshold field (investigator confirmed). Derive
progress for an **approval** stage as `count(actors where status === 'approved') de actors.length`
→ "N de M aprovaram". For a **review** stage, verdicts are not in the DTO — show the actor pool
count only, no quorum bar. If `actors` is empty, hide the quorum element (fail-safe, no fabricated
"0 de 0"). Label it as progress, never as an authoritative threshold.

### 5. Due-date relative formatter is **new** (none exists)

No "vence em N dias"/overdue formatter exists (`LockBadge.relativeTime` is past-only, unexported).
Add `formatDueRelative(dueAt: string | null, now?: number): { label: string; overdue: boolean }`
to `lib/format/dates.ts`:
- null/invalid → `{ label: '—', overdue: false }`.
- future same day → "vence hoje"; future → "vence em N dia(s)" (ceil of day diff).
- past → "atrasado há N dia(s)", `overdue: true` (drives the overdue chip style).
- `now` param (default `Date.now()`) so tests are deterministic.

### 6. **Meaning-of-signature copy — compliance contradiction, flagged for HS-1**

The master plan §F4 Step 2 literal copy cites **"21 CFR 11.50"**, but the shipped legal statement
(`documentSignoffDecision.ts:72-74`, rendered by `ArtifactDecisionPanel`) cites Brazilian **MP
2.200-2/2001** (ICP-Brasil). The API `signature_meaning` doc comment cites 21 CFR 11.50(a)(3). These
are **two different jurisdictions' regulations** for the same signature. **F4 will NOT silently
overwrite a live compliance statement with a second jurisdiction's citation.** Conservative default
(shipped in F4): keep the existing MP 2.200-2 `legal.text` untouched, and add a **meaning-of-
signature line derived from the decision** above the password field via the decision model —
approve → "Ao assinar, você declara **aprovação** deste documento." / reject → "…declara
**rejeição**…" (no regulation citation appended beyond the existing one). The 21 CFR-vs-MP 2.200-2
question is surfaced to the operator at **HS-1** for ratification; if the operator wants the 21 CFR
line, it is a one-line copy change post-ratification. (Global-Maximum rule: stop on the contradiction,
surface it, don't patch a compliance claim around it.)

## Non-goals

- **Editing the shared `ArtifactApprovalScreen`** to remove bands globally — templates use it; the
  cockpit neutralizes the bands via `decision/approvalChain/actions = null/[]` instead.
- **Rebuilding the signoff dialog** — reuse `ArtifactDecisionPanel` + the adapter's `model.decision`
  (password re-auth, legal, options, stale/error) verbatim inside `DecisionFooter`.
- **Delegation UI** — `useDelegationsMutations` exists (F2) but the admin/delegate surface is out of
  M2c (appetite rabbit-hole). Not wired here.
- **Rendering review verdicts as distinct timeline events** — the DTO exposes stage `status` +
  signoffs, not a verdict history array. Timeline shows submit + signoffs (with `signature_meaning`)
  + final status; verdict-event history is a data-gap noted, not invented.
- **Oversee dashboard** — F5 worklist territory, not the sidebar.

## Interview record (B1.5) — resolved design questions

| # | Question | Finding (file:line) | Decision |
|---|----------|--------------------|----------|
| 1 | Edit shared `ArtifactApprovalScreen` to remove the dup band? | It's shared with templates (`ArtifactApprovalScreen.tsx` consumed by template approval too). | No. Cockpit passes `decision/approvalChain/actions = null/[]` so bands don't render; the shared component is untouched. |
| 2 | Rebuild or reuse the signoff dialog? | `ArtifactDecisionPanel` (`:92-216`) already does password re-auth + legal + options + stale/error, driven by `model.decision` built by the adapter (`documentSignoffDecision.ts`). | Reuse it inside `DecisionFooter` for approval mode. Zero rebuild. |
| 3 | Where does quorum come from? | No quorum/threshold field on the stage schema (`api-types:3338-3357`). | Derive from `actors[]` (approved/total) for approval stages; pool-count only for review; hide when empty. |
| 4 | Existing due relative-time formatter? | None; `LockBadge.relativeTime` is past-only + unexported (`LockBadge.tsx:11-22`). | New `formatDueRelative` in `lib/format/dates.ts` with injectable `now`. |
| 5 | 21 CFR vs MP 2.200-2 legal copy? | FE ships MP 2.200-2 (`documentSignoffDecision.ts:72-74`); master plan literal says 21 CFR; API `signature_meaning` comment cites 21 CFR. | Contradiction. Keep MP 2.200-2 + add decision-derived meaning line; **flag to operator at HS-1**. No silent jurisdiction swap. |
| 6 | Stale banner replacement? | `DocumentApprovalExtras.tsx:99-106` renders it off `isStale` (polling adapter). F2 + signoff already invalidate `QK.approval.all`. | Delete the banner; rely on invalidation. |
| 7 | Lock/state badges — keep? | `DocumentApprovalExtras` renders `LockBadge` + `StateBadge`. | Fold `LockBadge` into `StageContextHeader` (keep the concurrent-lock signal); state conveyed by stage context + timeline. |

## Validation Gate

- **`sidebar/__tests__/ApprovalSidebar.test.tsx`** (render with a QueryClient; props-driven):
  - review mode → renders `SuggestionList` + the two verdict CTAs; **no** "Aprovar e assinar"
    button, **no** password input.
  - approval mode → renders `ArtifactDecisionPanel` (password input present) + "Outras ações";
    **no** verdict CTAs, **no** suggestion cards.
  - **exactly ONE** timeline in the DOM (`getAllByText('Fluxo de aprovação')`/timeline test-id
    length === 1).
  - integrity: `frozen_content_hash` NOT in the DOM while the `<details>` is collapsed; present
    after expanding (toggle `open`).
  - the scroll container has `overflow-y: auto`; `DecisionFooter` is `position: sticky`.
- **`sidebar/__tests__/DecisionFooter.test.tsx`**:
  - review: "Solicitar mudanças" dialog submit **disabled until the comment is non-empty**; on
    submit calls `useReviewVerdictMutation` with the right `{instanceId,stageId,etag,body}`.
  - approval: renders the decision panel; the meaning-of-signature line reads "declara aprovação"
    for an approve option, "declara rejeição" for reject.
- **`__tests__/formatDueRelative.test.ts`**: "vence hoje" / "vence em N dias" / "atrasado há N dias"
  + overdue flag, deterministic via injected `now`; null → "—".
- **Cockpit regression** (`ApprovalCockpitPage.test.tsx`): update the integrity assertions
  (hash now behind the disclosure) and confirm no toast / no stale banner; W2 session/autosave
  still call-count 0.
- `ReviewDocumentCanvas` grep-zero still holds; `DocumentApprovalExtras` deleted (grep-zero in
  approval feature).
- `vitest run` (targeted) PASS; `tsc --noEmit -p tsconfig.build.json` clean; wine tokens only on the
  new components (no `--slate-*` on the new sidebar chrome; reuse of `ApprovalTimelinePanel` keeps
  its existing tokens).

## Approval

- **Contract approved:** 2026-07-07 (main session, per ratified master plan §F4 + design spec §8 C2).
  Consumer = the cockpit aside; operator holds the HS-1 close gate (and the §6 21 CFR/MP 2.200-2
  ratification).
