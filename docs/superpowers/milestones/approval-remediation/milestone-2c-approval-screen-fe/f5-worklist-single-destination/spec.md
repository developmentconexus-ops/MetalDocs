# F5 — Worklist single destination (C3) — Spec

**Feature:** M2c `approval-screen-fe` / F5. Governing: design spec §8 **C3**
(`docs/superpowers/specs/2026-07-07-approval-remediation-design.md:145`) + master plan §F5.

> C3 (verbatim): "*worklist destino único: `/approvals` lists instances where actor is eligible on
> the active stage (or oversee). Real filters (stage kind, due, doc type), due-date sort, teaching
> empty state, items deep-link into cockpit in the right mode. Notifications point here.*"

## Interview record (fail-closed contract discovery)

| # | Question | Resolution |
|---|----------|------------|
| 1 | Does a single worklist + single cockpit route already hold? | **Yes** — `routes.tsx`: `/approvals` → InboxPage, `/approvals/:documentId` → ApprovalCockpitPage. No legacy duplicate list to collapse (investigator §8). "Single destination" is a *behavior* gap, not a routing one. |
| 2 | Where does a worklist item currently open? | **The bug.** `InboxPage.openDocument` navigates `/documents/${document_id}/edit` — the **author editor** (writable session). Sending a reviewer/approver there re-opens the exact W2 vector F3 killed at the cockpit. Only the explicit Approve/Reject buttons go to the cockpit (`/approvals/:id?decision=…`). |
| 3 | So what is the C3 fix? | Row/primary-open on a worklist item MUST navigate to the **cockpit** (`/approvals/:documentId`), which already resolves review/readonly/oversee mode (F3/F4). No `?decision` needed for a plain open; keep the decision deep-link as an optional convenience on the action buttons. Remove all `/documents/:id/edit` navigation from the worklist. |
| 4 | Is the API layer ready? | **Yes (F2 landed it).** `ListInboxParams` has `stage_kind?`, `due_before?`, `scope?`; `listInbox` forwards them; `useInboxQuery(params)` is parameterized. **But no caller passes params** — `InboxPage` calls `useInboxQuery()` bare. F5 adds filter state + UI only. |
| 5 | "doc type" filter — deliverable? | **No — contract gap.** No `doc_type` param exists in `ListInboxParams` or `ApprovalInboxItem`; adding one is a backend/contract change (HS-2 boundary, outside M2c FE scope). **Deviation:** F5 delivers stage-kind + due filters; doc-type is deferred and flagged at HS-1. |
| 6 | oversee toggle — how gated? | Design says "eligible… (or oversee)". No capability-probe pattern exists. **Decision (global-max):** render an oversee toggle unconditionally; when on, pass `scope=oversee`. Handle a 403 **reactively** — surface "Você não tem permissão de supervisão." and auto-revert the toggle. NO preemptive probe request (master plan's probe-gate is rejected: a probe on every mount is wasted latency; reactive 403 is the cleaner engineering). |
| 7 | Sort order? | Due-date **ascending**, `due_at === null` sorted **last** (nulls have no urgency). Client-side (backend returns response order; investigator §3 confirms no server sort). Overdue items therefore surface at the top. |
| 8 | Overdue display? | Reuse `formatDueRelative` (F4, `lib/format/dates.ts`) — `{label, overdue}`. Render the due chip on each item; `overdue` styles it with `--danger`. `due_at` is currently rendered NOWHERE in the inbox (investigator §7). |
| 9 | Tokens? | Worklist already on wine tokens (`--brand`); zero `--slate-*` (investigator §6). No migration in F5; C4 audit (F7) unaffected. |

## Consumer contract (what F5 must deliver)

**Consumer = the reviewer/approver/oversee user landing on `/approvals`.**

1. **Single destination:** clicking a worklist item's primary open action navigates to
   `/approvals/:documentId` (the cockpit), never `/documents/:id/edit`. The cockpit resolves the
   right mode. Approve/Reject action buttons continue to deep-link the cockpit with
   `?decision=approve|reject` (still the cockpit — one destination, optional intent preselect).
2. **Filters** (`InboxFilters` — new component): stage-kind select `[Todos | Revisão | Aprovação]`
   → `stage_kind`; due select `[Todas | Vence em 7 dias | Atrasadas]` → `due_before` (7-days = now+7d
   ISO; Atrasadas = now ISO, client also filters `overdue`); oversee toggle → `scope=oversee`.
   Filter state lives in `InboxPage`, flows into `useInboxQuery(params)`.
3. **Due-date ascending sort**, nulls last; overdue chip via `formatDueRelative`.
4. **Teaching empty state** (replaces the bare "Nenhuma aprovação pendente."):
   *"Nenhuma aprovação pendente. Documentos submetidos a rotas onde você é revisor ou aprovador
   aparecem aqui."* When a filter is active and yields nothing: *"Nenhuma aprovação corresponde aos
   filtros."* (distinct — teaches the filter is the reason, not absence of work).
5. **Loading + error states** preserved (existing InboxStack strings); Timeline view also gets the
   states it currently lacks (investigator §1 — Timeline receives no state props today).
6. **oversee 403** → inline `role="alert"` note + toggle auto-reverts; no toast.

## Non-goals (explicit)

- **No doc-type filter** (contract gap — deferred, flagged HS-1).
- **No preemptive oversee capability probe** (reactive 403 only).
- **No backend/contract changes** — API layer is F2-complete; F5 is FE UI + wiring.
- **No token migration** (already wine).
- **No new worklist/cockpit routes** — single destination already holds structurally.
- **No full a11y audit** (that is F7/C4); F5 keeps parity + adds visible focus on new controls only.
- **No stack/timeline toggle removal** — kept.

## Validation Gate

- **New:** `InboxFilters.test.tsx` — each control maps to the correct `ListInboxParams` field; oversee
  403 reverts + shows the note.
- **Extend `InboxPage.test.tsx`:** (a) primary item open navigates `/approvals/:documentId` NOT
  `/documents/:id/edit`; (b) filter selection re-queries with mapped params; (c) due-asc sort with
  null-last ordering; (d) teaching empty state (no-work vs filtered-empty, distinct copy); (e) overdue
  chip renders for a past `due_at`.
- **Regression:** existing InboxPage assertions (view persistence, counter nav, approve/reject
  deep-link) stay green.
- **Proof commands:** `npx vitest run src/features/approval` GREEN; `npx tsc --noEmit -p
  tsconfig.build.json` clean; `grep -rn "documents/.*edit\|/edit" src/features/approval/pages/InboxPage.tsx`
  → ZERO (no author-editor navigation left in the worklist).

## Deviations to surface at HS-1

- **D1 — doc-type filter deferred** (backend/contract gap; C3 lists it, no param exists).
- **D2 — oversee gated reactively, not by probe** (rejects master plan's probe-gate for a
  lower-latency design; behavior equivalent from the user's view).
