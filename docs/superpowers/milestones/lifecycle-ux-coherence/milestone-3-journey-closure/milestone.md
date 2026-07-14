# M3 — Journey Closure

**Program:** lifecycle-ux-coherence · **Governing spec:** `docs/superpowers/specs/2026-07-06-lifecycle-ux-coherence-design.md` §3 (findings 9–12, 20)
**Roadmap unit:** 4.1 · **Status:** in progress

## Objective (outcome)
Every lifecycle event/screen links to the next action; zero dead FE affordances.

## Findings in scope + runtime-truth disposition

The spec was authored 2026-07-06, **before ADR 0080** retired the standalone approval
cockpit and collapsed it into the mode-adaptive `DocumentWorkspacePage`
(`/documents/:id`). Grounding the findings against current code:

| # | Spec finding | Disposition |
|---|---|---|
| 9 | Cockpit → detail: no link | **Already resolved by ADR 0080 / M2** — `WorkspaceSidebar.tsx:148` renders `<Link to="/documents/:id/details">`. The workspace (post-cockpit) links to the detail record surface. Evidence-only, no code. |
| 10 | Detail → cockpit: no link when instance exists | **Already resolved by ADR 0080 / M2** — `DocumentDetailRoute` "Visualizar documento" (`handleView` → `/documents/:id`) links the detail surface to the workspace, which IS the approval surface under 0080 (mode-adaptive: renders approving mode when an instance + eligible viewer exist). Evidence-only, no code. |
| 11 | Notifications not clickable (resource_id ignored) | **BUILD** — `NotificationRow` ignores `resource_type`/`resource_id`. Backend emits `resource_type="document"`, `resource_id=<document id>` for the 5 user-facing lifecycle events (`fanout_worker.go` Work switch). Add a fail-closed deep-link mapper → `/documents/{id}`; row becomes a link, still marks read. |
| 12 | "Abrir Fanout" → nonexistent `/distribution` → dashboard | **BUILD** — two spots (`DocumentDetailRoute.tsx:270` navigate, `useDocumentArtifact.ts:235` kpi href) point at `/distribution` (no such route → wildcard → dashboard). Working child route is `/documents/:id/details/distribution` (relative `distribution`). Two vitest suites pin the broken absolute path — updated. |
| 20 | Dead FE: obsolete(), archiveDocument(), ActivityPanel, "···" stub | **DELETE** — `approvalApi.ts obsolete()` (+re-export), `documents.ts archiveDocument()` (both zero callers), `ActivityPanel` (permanent "Em breve" placeholder + LibraryPage wiring), LibraryPage "···" `moreBtn` stub (stopPropagation only). |

## Slices
- **A** — F12 fanout CTA fix (+ pinning-test updates)
- **B** — F11 notification deep-link mapper + clickable row
- **C** — F20 dead-affordance deletion

## QA
L3 browser QA is operator-deferred (consolidated session). Feature-unit targeted UI
walk attempted only if trivially possible on a free port; CANNOT is acceptable.

## Close-out
_(filled at unit close)_
