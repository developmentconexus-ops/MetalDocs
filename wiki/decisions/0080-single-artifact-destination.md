# ADR 0080 — Single Artifact Destination (Cockpit Pattern Retired)

> **Status:** Accepted
> **Date:** 2026-07-09
> **Scope:** Frontend routing for the document artifact. One mode-adaptive screen owns every
> relationship a user (author, reviewer, approver, observer) can have to a document, mounted at
> the canonical URL `/documents/:id`. The standalone approval cockpit route is retired; the
> record surface (revisions/distribution/lineage) moves one path segment deeper.
> **Milestone:** `docs/superpowers/milestones/approval-remediation/milestone-2d-workflow-coherence-fe/f5-single-screen-shell/`
> **Governing spec:** `docs/superpowers/specs/2026-07-08-approval-workflow-coherence-design.md` §8 item 2
> ("Single artifact destination (A): one screen per artifact, mode-adaptive; cockpit pattern
> retired.")
> **Design brief:** `docs/superpowers/specs/2026-07-08-single-screen-design-brief.md` (destination
> pinned by operator 2026-07-08 — canonical artifact URL)
> **Related:** ADR 0078 (viewer-facts contract — the mode this screen renders is derived, not
> role-guessed), ADR 0022 (capabilities)
> **Key files:**
> - `frontend/apps/web/src/features/documents/routes.tsx` — `/documents/:id` → `DocumentWorkspacePage`;
>   `/documents/:id/details` → `DocumentDetailLayout`; `/documents/:id/edit` → redirect
> - `frontend/apps/web/src/features/approval/routes.tsx` — `/approvals/:documentId` → redirect
> - `frontend/apps/web/src/features/documents/pages/DocumentWorkspacePage.tsx` — the mode-adaptive
>   screen owner (F2d.5 S1/S2)
> - `frontend/apps/web/src/features/approval/pages/InboxPage.tsx` — worklist deep-links retargeted
> - `frontend/apps/web/src/features/documents/pages/DocumentDistributionPage.tsx` — breadcrumb
>   retargeted to `/documents/:id/details`

## Context

M2c shipped two parallel URLs for the same underlying artifact: `/documents/:id` (the record
surface — revisions, distribution, lineage) and `/approvals/:documentId` (the standalone approval
cockpit, plus `/documents/:id/edit` for the author's writable editor). Reviewers, approvers, and
authors each landed on a different page for the same document depending on which capability they
held. This is the "second approval system" coherence debt named in the governing spec §1.3: an
author polling a reviewed document sees the cockpit's read-only shell rather than their own
document; a reviewer who wants to see the record surface has no path there without leaving the
cockpit. The multiple destinations also multiplied the audit-trail-adjacent UI (timeline,
verdicts, signature panel) across two independently-maintained page trees.

The design brief (§1) settles this: one screen, `/documents/:id`, that adapts by
`deriveWorkspaceMode` (F2d.3) — author-editing, author-changes-requested, author-waiting,
reviewing, approving, observing, lifecycle. Author, reviewer, approver, and observer see the SAME
shell (`DocumentShell` header + canvas + sidebar); what changes is the canvas mode, the sidebar
decision footer, and the contextual panels — the Google Docs editing/suggesting/viewing model,
not a page-per-role split. F2d.5 (S1–S2) built that screen (`DocumentWorkspacePage`,
`ModeChip`, `WorkspaceSidebar`, `EditorCanvas`). This ADR (S3) flips routing to make it the
canonical destination.

## Decision

**1 — Canonical artifact URL.** `/documents/:id` renders `DocumentWorkspacePage` — the single
mode-adaptive screen. This is now the primary destination for every worklist/deep-link entry
point regardless of the viewer's relationship to the document.

**2 — Record surface moves one segment deeper.** The existing `DocumentDetailLayout` (tabs:
Documento / Distribuição — revisions, distribution, lineage) is unchanged in content and
behavior; it remounts at `/documents/:id/details`. It stops competing with the workspace for the
canonical path.

**3 — Cockpit and editor routes retired as *routes*, not as files.** `/approvals/:documentId` and
`/documents/:id/edit` become redirects to `/documents/:id` (the editor redirect preserves the
`:documentId` param; the approval redirect additionally preserves `location.search`, e.g.
`?decision=approve`, so an in-flight signature deep-link still lands on the right screen with the
right decision pre-seeded). `ApprovalCockpitPage` and `DocumentEditorRoutePage` are **not
deleted** by this ADR — `DocumentEditorRoutePage`'s writable canvas lives on as the extracted
`EditorCanvas` component inside the workspace; `ApprovalCockpitPage`'s file deletion is scoped to
a later feature (F2d.7) once nothing routes to it. This ADR only stops routing traffic to the
cockpit pattern; it does not yet remove the dead file.

**4 — Deep-link retargeting.** The approval worklist (`InboxPage`) now navigates directly to
`/documents/:id` (with `?decision=` preserved for the approve/reject actions) instead of
`/approvals/:id`, so no in-app navigation round-trips through the redirect. The distribution page
breadcrumb now points back at `/documents/:id/details` (the record surface it lives under), not
the workspace.

**5 — Signature ceremony stays in-screen.** The approval signature panel remains a re-auth
control WITHIN the workspace sidebar (approving mode), never a separate URL — consistent with the
Qualio/Veeva pattern this ADR follows. **Reopen trigger:** an external-signer persona (no app
account) would justify a DocuSign-style standalone ceremony page; not before that need exists.

**Amendment (F2d.5b re-scope, 2026-07-09 — supersedes the earlier same-day note):** viewing
policy for the single destination is settled as follows. **In-approval viewing (draft /
under_review, all read modes) stays on the in-app source canvas** (read-only `DocumentShell`);
**the PDF is the official post-approval artifact** — the workspace canvas renders it
(`PdfCanvas` via `GET /documents/{id}/view`) only for `approved / scheduled / published`.
Rationale: a pre-freeze PDF would be falsely official (computed tokens resolve only at freeze),
and the Veeva-style continuous rendition solves an uploaded-binary-source gap MetalDocs (in-app
editing) doesn't have. The signature binds the source `content_hash` (existing If-Match
contract), consistent with 21 CFR Part 11 (signature binds the record; the rendition is its
human-readable form). Bundle note: `DocumentShell`'s TipTap import becomes a lazy chunk in
F2d.5b — docx read modes fetch it on canvas mount; the PDF canvas never fetches it.
**Reopen trigger:** customer demand for print-fidelity preview *during* approval → Veeva-pattern
continuous rendition (eager render at submit/resubmit via outbox, keyed by `content_hash`) —
researched and shaped 2026-07-09, design record
`docs/superpowers/specs/2026-07-09-f5b-pdf-official-view-design.md`.

**Closure (F2d.7 cockpit retirement, 2026-07-09):** the migration is now physically complete. The
retired `ApprovalCockpitPage`, its wholly-cockpit-only adapter `useDocumentApprovalArtifact`, the
unmounted `DocumentEditorRoutePage`, and the cockpit-only sidebar stack (`ApprovalSidebar` +
`StageContextHeader` + `SuggestionList`) were **deleted** — leaving zero dead cockpit code. The
`/approvals/:documentId` and `/documents/:id/edit` **redirects are KEPT** as the bookmark/deep-link
preservation surface (the `/approvals` one still forwards `location.search`, e.g. `?decision=`). All
live navigation now targets `/documents/:id` directly instead of bouncing through a redirect
(`getActiveSiblingDestination` collapsed to the canonical path; `DocumentDetailRoute` +
`NewDocumentWizardPage` constructors retargeted). Deferred (not deleted): `CancelInstanceDialog`
remains for the yet-to-be-wired single-screen "Cancelar instância" action (`WorkspaceSidebar.tsx`
S2b note). No new ADR — this decision governs. Evidence:
`docs/superpowers/milestones/approval-remediation/milestone-2d-workflow-coherence-fe/f7-cockpit-retirement/evidence.md`.

## Consequences

- One screen, one mounted route, one accountability column (timeline + verdicts + signature) per
  artifact — no more author/reviewer page fork.
- Worklist and distribution deep-links resolve to the correct segment on the first hop; only
  external/stale links (e.g. a bookmarked `/approvals/:id?decision=approve`, or an existing
  "open editor" CTA that still targets `/documents/:id/edit`) pay one extra client-side redirect.
- `ApprovalCockpitPage` and `DocumentEditorRoutePage` become dead code reachable only by direct
  URL entry until F2d.7 deletes them; both remain fully functional (and covered by their existing
  test suites) in the interim — this ADR is a routing change, not a deletion.
- The record surface's own internal links (e.g. distribution) must now target `/details`
  explicitly; a raw `/documents/:id` link from within the record surface would (correctly) land on
  the workspace, not loop back to the record surface.

## Alternatives rejected

- **Keep both routes, add cross-links.** Preserves the coherence debt (two independently
  evolving page trees for one artifact) that the governing spec names as the root defect; only
  patches the *navigation* symptom, not the destination.
- **Delete `ApprovalCockpitPage`/`DocumentEditorRoutePage` now.** Out of this feature's boundary
  (F2d.5 S3 owns the route flip only; F2d.7 owns deletion) and unnecessarily couples a routing
  change to a file-removal change — the redirect alone fully retires the pattern's *reachability*.
- **Make `/documents/:id` the record surface and add a new path for the workspace.** Rejected by
  the operator-pinned destination (design brief, 2026-07-08): the workspace — not the record
  view — is the artifact's canonical identity; the record surface is a secondary, explicitly-
  labeled view of it.
