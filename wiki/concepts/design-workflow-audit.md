# Design-vs-Workflow Audit

**Last verified:** 2026-05-06

When implementing any screen sourced from `frontend/apps/web/design-source/<slug>/` (the Claude-Design-generated mockups powering the screen-redesign initiative), you MUST audit the design against the real MetalDocs workflow before writing TSX.

## Why this exists

The redesign mockups (`*.jsx` references in `design-source/`) were produced by an AI designer with no operating knowledge of:

- MetalDocs document state machine (draft / review / approved / frozen / rejected / archived)
- RBAC roles (operators vs viewers vs approvers vs admins)
- Audit / approval models (who actually has pending items, who only consumes)
- Persona reality (most users land on a screen to *find a document and read it*, not to triage approvals)

Implementing those mockups 1:1 ships UX debt: stat cards counting things we don't track, filter tabs that map to nothing in our domain, sidebars whose contents assume privileges most users don't have.

## The audit (run before writing any TSX)

For every screen sourced from `design-source/`, walk every visible UI element and answer:

1. **Does it map to a real domain concept?**
   - Stat card "Em revisão" → maps to `document.status === 'review'`. ✅ Keep.
   - Stat card "Frozen este mês" → does our model track freeze date? If yes, ✅. If no, ❌ cut.
   - Filter tab "Aprovação pendente" → distinct from "Em revisão" in our state machine? If they collapse to the same state, ❌ cut one.

2. **Does the persona on this screen actually need it?**
   - Library list = primary entry point for *all* users (viewers, authors, approvers). A viewer doesn't need an approval-triage sidebar. Either gate the sidebar by role, or drop it from the default layout.
   - Approval-only widgets belong on the Approval Inbox screen, not Library.

3. **Are the counts/values derivable, or are they invented?**
   - Mock data in `*.jsx` (e.g. `count: 2847`, `count: 38`) is decorative. Don't carry the numbers into the code. Either derive from a real query, or cut the widget.

4. **Does the layout assume privileges we don't enforce?**
   - "Bulk freeze", "Export PDF for everyone", "See all areas" — verify against `wiki/concepts/authz-tiers.md` before exposing the affordance.

## Where to record findings

Each screen has `frontend/apps/web/design-source/<slug>/NOTES.md`. Add an **Audit findings** section with three subsections:

- **Keep** — elements that map to real workflow.
- **Cut** — elements that don't map; reason.
- **Defer** — elements that need backend work (new endpoint, new state) before they can be honest. Open a follow-up task.

## When to run

Run during the "Orient before coding" step of the `metaldocs-frontend` skill, after reading the design-source files and before placing any new file. Cross-reference:

- [wiki/concepts/controlled-documents.md](controlled-documents.md) — document model + states
- [wiki/concepts/authz-tiers.md](authz-tiers.md) — RBAC tiers
- [wiki/architecture/data-model.md](../architecture/data-model.md) — what fields actually exist

## Example: Library screen (2026-05-06)

The Library mockup ships:

- 4 stat cards (Em revisão / Aprovação pendente / Frozen este mês / Próx. revisão)
- 6 filter tabs with counts (Todos 2847 / Meus 38 / Em revisão 8 / Aprovação pendente 3 / Frozen 2612 / Rascunhos 23)
- Right activity sidebar (pending approvals + audit trail)
- Author column, Rev column, Updated column

A viewer-role user lands here just to find and read a document. The approval-pending sidebar and pending-approval stat card add no value to that user. The author/updated columns may matter to authors but not viewers. Findings get captured in `design-source/library/NOTES.md` under "Audit findings", and the LibraryPage implementation gates approval widgets behind role checks (or removes them from the default layout entirely).
