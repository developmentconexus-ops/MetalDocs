# F5 — Plan

Seeded from master plan §F5 + design §8 C3 + F5 investigator map (agentId ac9e41eaead6680a0). TDD via
fresh implementer subagent (sonnet) + independent reviewer subagent. **Implementer uses its OWN tools;
does NOT spawn sub-agents.**

## Ground truth (from investigator — do not re-derive)
- API layer is F2-complete: `ListInboxParams{stage_kind?,due_before?,scope?,area_code?,limit?,offset?}`
  (`approvalTypes.ts:55-62`), `listInbox` forwards them (`approvalApi.ts:62-86`),
  `useInboxQuery(params)` parameterized (`useInboxQuery.ts`), `QK.inbox(params)` keyed. **F5 adds no
  API/type code** — it supplies params from new UI.
- `InboxPage.tsx:28` calls `useInboxQuery()` **bare** — the wire-up gap.
- `InboxPage.openDocument` (`:48-60`) navigates `/documents/${document_id}/edit` — **the C3 bug to fix**.
- `openDecisionFlow` (`:62-74`) already navigates `/approvals/${document_id}?decision=…` — keep.
- Item DTO `ApprovalInboxItem`: `document_id`, `document_title`, `stage_kind?`, `due_at:string|null`,
  `quorum_progress`, `stage_label`, `submitted_at` (no `status`, no `doc_type`).
- `InboxToolbar` has a **disabled** "Filtros" button (`InboxToolbar.tsx:33-35`, `title="Em breve"`).
- `formatDueRelative(dueAt, now?) → {label, overdue}` exists (`lib/format/dates.ts:57-85`); unused in inbox.
- InboxStack owns loading/error/empty strings (`InboxStack.tsx:56-68`); InboxTimeline gets no state props.
- Wine tokens already in place; no `--slate-*`.

## Ordered tasks

1. **[TDD] Failing tests first** (author all before impl; run RED):
   - `features/approval/components/__tests__/InboxFilters.test.tsx` (new)
   - Extend `features/approval/pages/InboxPage.test.tsx` with the 5 Validation-Gate cases (single
     destination, filter→param, due-asc null-last, teaching empty ×2, overdue chip).
   Cases exactly per spec Validation Gate.
2. **`lib/inbox/sortByDue.ts`** (or inline in InboxPage) — pure `sortByDueAsc(items)`: ascending
   `due_at`, `null` last. Unit-testable; keep pure. (If trivial, colocate + test via InboxPage.)
3. **`components/InboxFilters.tsx`** (new) — props `{ value: InboxFilterState, onChange }`.
   `InboxFilterState = { stageKind?: 'review'|'approval'; due?: 'all'|'next7'|'overdue'; oversee: boolean }`.
   Controls: stage-kind `<select>` [Todos|Revisão|Aprovação]; due `<select>` [Todas|Vence em 7 dias|
   Atrasadas]; oversee toggle (checkbox/switch). Wine tokens, visible focus, PT-BR sentence case,
   labelled for a11y. Emits filter state up; does NOT own query.
4. **Filter→param mapping** in `InboxPage` — `toParams(state, now)`: `stageKind→stage_kind`;
   `due==='next7'→due_before = new Date(now+7*864e5).toISOString()`; `due==='overdue'→due_before =
   new Date(now).toISOString()` (server-side pre-filter) **and** a client `overdue` guard on the
   result (belt+braces, since `due_before` is inclusive of not-yet-overdue-today); `oversee→scope:'oversee'`.
   Wire `const params = toParams(filters, Date.now()); useInboxQuery(params)`.
5. **Rewire `openDocument`** — navigate `/approvals/${item.document_id}` (the cockpit), NOT
   `/documents/:id/edit`. Remove the `fetchActiveDocumentInstance`→editor path from the primary open.
   (Keep `openDecisionFlow`'s cockpit deep-link.) Verify no `/edit` navigation remains in InboxPage.
6. **Enable Filtros** — replace the disabled `InboxToolbar` button with a control that toggles the
   `InboxFilters` panel (or always-render the panel above the list — pick the lower-friction path;
   panel visible is fine). Wire `filters` state + `onChange` in InboxPage.
7. **Due chip + overdue** — render `formatDueRelative(item.due_at)` on each item (InboxStack card +
   InboxTimeline row). `overdue` → `--danger` styling. Add to InboxApprovalCard / InboxStack render.
8. **Sort applied** — `sortByDueAsc(items)` before render in both stack + timeline paths.
9. **Teaching empty state** — InboxStack empty branch: if any filter active → "Nenhuma aprovação
   corresponde aos filtros."; else → "Nenhuma aprovação pendente. Documentos submetidos a rotas onde
   você é revisor ou aprovador aparecem aqui." Pass an `isFiltered` flag down (or a resolved
   `emptyMessage`). Give InboxTimeline the same loading/error/empty states it currently lacks.
10. **oversee 403 reactive** — `useInboxQuery` error where `scope==='oversee'` and status 403 →
    render inline `role="alert"` "Você não tem permissão de supervisão." + reset `filters.oversee`
    to false. No toast. (Detect 403 via the problem+json error shape the query throws — check how
    `isError`/error surfaces in the existing hook; may need to read `error` not just `isError`.)
11. **Verify:** `grep '/edit' src/features/approval/pages/InboxPage.tsx` → zero;
    `grep 'useInboxQuery()' src/features/approval` → zero (all callers pass params). Targeted
    `vitest run src/features/approval` GREEN; `tsc --noEmit -p tsconfig.build.json` clean.
12. **Review pass** — independent reviewer subagent (sonnet): C3 single-destination (row→cockpit, no
    editor), filter→param correctness, due-asc null-last, teaching empty ×2, overdue chip, oversee
    reactive-403 (not probe), no backend/type changes, no doc-type filter smuggled in, test
    non-tautology, wine tokens. Apply accepted findings.

## Files touched
- `features/approval/pages/InboxPage.tsx` (rewire nav + filters state + params + sort + empty)
- `features/approval/pages/InboxPage.test.tsx` (extend)
- `features/approval/components/InboxFilters.tsx` (new) + `.module.css`
- `features/approval/components/__tests__/InboxFilters.test.tsx` (new)
- `features/approval/components/InboxToolbar.tsx` (enable Filtros toggle)
- `features/approval/components/InboxStack.tsx` (due chip, teaching empty, sort consume)
- `features/approval/components/InboxApprovalCard.tsx` (due chip/overdue) — if card owns the row
- `features/approval/components/InboxTimeline.tsx` (state props + due chip + sort)
- `lib/inbox/sortByDue.ts` (+ test) OR inline pure fn in InboxPage
- possibly `lib/queryKeys.ts` — NO (already keyed on params)

## Risks
- **C3 single-destination regression** — the core change is nav retargeting. Guardrail: test asserts
  `/approvals/:documentId` and `grep '/edit'` → zero. Confirm the cockpit (F3/F4) actually resolves
  mode from `document_id` alone (it does — F3 map).
- **doc-type scope creep** — do NOT invent a client-side doc-type filter or a backend param. Deferred,
  flagged D1. Reviewer checks none was smuggled in.
- **oversee 403** — must read the query `error` object (not just `isError`) to detect 403 vs other
  failures; reverting the toggle must not loop (reset once, don't re-trigger the oversee query).
- **due_before semantics** — `due_before` is a server hint; "Atrasadas" needs the client `overdue`
  guard too, else today's not-yet-due items with `due_at < now+ε` could leak. Keep both.
- **junction drift** — vitest broken → full `pnpm install` in `frontend/apps/web`; do NOT hack config.
- **Timeline parity** — adding states/sort to InboxTimeline must not break its existing rendering;
  keep its structure, add the missing branches.
