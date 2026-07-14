# M3 — Journey Closure

**Program:** lifecycle-ux-coherence · **Governing spec:** `docs/superpowers/specs/2026-07-06-lifecycle-ux-coherence-design.md` §3 (findings 9–12, 20)
**Roadmap unit:** 4.1 · **Status:** CLOSED (2026-07-14)

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
CLOSED 2026-07-14. Branch `unit-4.1-journey-closure`, base `a684cb4e`, range
`a684cb4e..daf471a5` (4 commits; F12/F11/F20 slices + 1 restoration fix). Net 17 files,
+177/−380. Gates: L0 tsc exit 0; L1 vitest 81 files / 546 tests passed. Each slice + final
net diff independently reviewed (cavecrew-reviewer, all "No issues"). One defect found &
remediated: slice A's `git add -A` on an incomplete worktree checkout committed 2237 spurious
tracked-file deletions → restored from base in `daf471a5`. F9/F10 already-resolved by ADR 0080
(evidence-only). L3 browser QA → operator (login-password prohibition). Not pushed.
Evidence: `docs/superpowers/reports/2026-07-14-unit-4.1-evidence.md`.
