# Backlog — Detalhe Signoff (F5.1)

Deferred-with-trigger items surfaced while building the Detalhe Signoff cockpit
(`/approvals/:documentId`). Each row names the concrete unblock condition.

| Item | Why deferred | Unblock trigger | Owner |
|------|--------------|-----------------|-------|
| "Mudanças vs versão anterior" diff tab | No document-diff backend exists. `GET /documents/{id}/view` returns a single rendered PDF pointer; there is no endpoint that compares the under-review revision against the previously published one. Rendering a fabricated diff would violate the honesty rule. | A backend diff endpoint (e.g. `GET /documents/{id}/diff?against={revision}`) returning structured added/removed/changed regions or a diff artifact. When it ships, wire the tab to it. | Frontend (consume) + Backend (produce) |

Until then the tab renders an honest explanation and no invented data.

## Pre-existing shared-component debt (surfaced by F5.1 reviewers, NOT introduced here)

These live in components the cockpit **consumes** (`ControlledDocumentDetailPanel`,
`SignoffDialog`, `useDocumentPdfStatus`) or in adjacent inbox code. F5.1 mounted/extended
them additively only — fixing them means editing shared, separately-tested components,
which the plan's no-fork guardrail (HS-2) reserves for a follow-up that owns them. The
cockpit makes some of these more visible, so they are recorded with triggers for the
operator's HS-1 decision.

| Item | Where | Why deferred | Unblock trigger | Owner |
|------|-------|--------------|-----------------|-------|
| Mojibake / missing accents in error strings | `SignoffDialog.tsx` (`error_server`, `error_not_eligible`) | Double-encoded Portuguese (`aprovaÃ§Ã£o`…) and missing accents predate F5.1; file outside the plan's additive scope. | A pass that owns `SignoffDialog` re-saves the strings as UTF-8 (or moves them to `lib/api/errorMessages.ts`). Cheap — fold into the next `SignoffDialog` edit. | Frontend |
| Native `window.prompt` for cancel reason | `ControlledDocumentDetailPanel.tsx` `handleCancelInstance` | Violates "never raw alert/prompt" (error-ux); replacing it is a structural change to a shared tested component (HS-2 fork territory), not an additive prop. | Follow-up PR that owns the panel replaces it with an inline cancel form / lightweight dialog. | Frontend |
| Hardcoded error strings bypass `resolveErrorMessage` | `ControlledDocumentDetailPanel.tsx` `handleCancelInstance` / `handleSubmitForReview` catch blocks | Pre-existing; should route `ApiError` through `resolveQueryError`. Out of additive scope. | Same panel-owning follow-up wraps catches with `resolveQueryError(err, fallback)`. | Frontend |
| Raw-hex (non-token) colors | `ControlledDocumentDetailPanel.module.css` | Entire file uses blue-gray hex absent from the wine token set; predates F5.1. | Panel-owning follow-up maps each hex to a `var(--token)`. | Frontend |
| Non-label `<span className={styles.label}>` collides with real `<label>` style | `ControlledDocumentDetailPanel.tsx` metadata rows | A11y/semantic collision; pre-existing. | Panel-owning follow-up renames the meta class / uses `<dt>`/`<p>`. | Frontend |
| `ControlledDocumentDetailPanel` > 400 LOC | `ControlledDocumentDetailPanel.tsx` (446) | Crosses god-component threshold; additive props worsened it marginally. | Follow-up extracts the submit-for-review form + stale-clock hook (~360 LOC after). | Frontend |
| `useDocumentPdfStatus` polls via `useEffect`+`setTimeout` | `documents/hooks/editor/useDocumentPdfStatus.ts` | Pre-existing anti-pattern (server state should be TanStack Query w/ `refetchInterval`); the cockpit is the first non-editor consumer. | Migrate the hook to `features/documents/queries/` as a Query with `refetchInterval` stopping on ready/failed. | Frontend (documents) |
| Inbox resolves context via imperative `await getActiveDocumentContext` | `InboxPage.tsx` `openDecisionFlow` / `openDocument` | Mirrors the pre-existing `openDocument` pattern; could use `queryClient.fetchQuery` to hit the warm cache. Kept consistent with the sibling handler rather than diverging in F5.1. | A pass that converts both inbox handlers to `queryClient.fetchQuery(QK.approval.activeDocument(...))`. | Frontend |
