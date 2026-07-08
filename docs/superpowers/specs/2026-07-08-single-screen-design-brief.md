# Design Brief — Single Document Screen (M2d / A3)

- **Date:** 2026-07-08
- **Status:** RATIFIED by operator (both open questions answered 2026-07-08)
- **Produced by:** `/impeccable shape` (routed from `polish`; target did not exist yet)
- **Governing spec:** `2026-07-08-approval-workflow-coherence-design.md` §4 (Milestone A / M2d)
- **Consumes:** PRODUCT.md (`frontend/apps/web/PRODUCT.md`, register `product`), Wine tokens
  (`src/styles/tokens.css`), M2c F4/F5/F6 ratified interaction designs

## 1. Summary

One screen (`/documents/:id`) that adapts by derived workspace mode (`deriveWorkspaceMode`)
— the Google Docs editing/suggesting/viewing model. Author, reviewer, approver, and
observer see the SAME shell; what changes is the canvas mode, the sidebar footer, and the
contextual panels. `ApprovalCockpitPage` is deleted; `/approvals/:documentId` and
`/documents/:id/edit` become redirects to `/documents/:id` (destination pinned by operator
2026-07-08 — canonical artifact URL). The record surface (revisions/distribution/lineage)
survives unchanged at `/documents/:id/details`. One surface, modular domains: approval
components (`features/approval`) compose into the document screen per mode; the signature
ceremony stays a re-auth panel WITHIN the screen, never a separate URL.

## 2. Primary action per mode

| Mode | Canvas | Sidebar footer | Contextual panel | Primary action |
|---|---|---|---|---|
| `author-editing` | **edits** (autosave) | — | meta + route preview | Submit for review |
| `author-changes-requested` | edits | — | RequestedChangesPanel (F6): verdict reason | Resubmit |
| `author-waiting` | read + **comment replies** | — | timeline + stage status | follow (no CTA) |
| `reviewing` | **suggests** (comments) | Verdict CTAs: "Pronto para aprovação" / "Solicitar mudanças" | timeline | record verdict |
| `approving` | reads **frozen content** (integrity disclosure F4) | Signature panel: password + legal effect; delegation disclosure when `via_delegation_from` | timeline | sign |
| `observing` | read | — | timeline; "Cancelar instância" when capability allows | oversight |
| `lifecycle` | read | Publish/Schedule per TRANSITION_POLICY | timeline + lineage | publish |

## 3. Visual direction

Restrained; Wine tokens; Inter Tight — PRODUCT.md governs, zero new palette. Scene:
quality manager on a factory floor / industrial office, fluorescent light, deciding and
signing with legal accountability — light theme, dense, sober (the current system).
Anchors: Google Docs (mode indicator), Veeva Vault (doc-control rigor), Linear (calm
sidebar density). Anti: slate drift (tolerated legacy palette — do not extend), generic
admin dashboard.

## 4. Scope

Production-ready; whole surface (1 screen × 7 modes); shipped-quality interactive; feeds
M2d feature specs — not a prototype.

## 5. Layout strategy

- **Constant DocumentShell**: header (code, StateBadge, revision), central canvas, right
  sidebar. The skeleton never changes — the user never "switches screens", the screen
  switches mode.
- **Mode indicator** in the header: discreet chip ("Editando" / "Revisando" / "Aprovando" /
  "Visualizando" / "Aguardando revisão"), Wine tokens, tooltip explaining WHY (derived
  from server `viewer` facts).
- **Sidebar = accountability column**: single timeline (submission → verdicts[] →
  signatures), always present when an instance exists; the footer is the ONLY decision
  surface (variant by `stage_kind` + eligibility — F4 contract honored).
- Document is the protagonist: chrome does not grow; contextual panels stack in the
  sidebar, never over the canvas.

## 6. Key states

- Loading: shell skeleton (no central spinner).
- No instance + draft = `author-editing`; no instance + non-author = `observing` read.
- Instance error: sidebar error panel; canvas stays readable.
- `changes_requested`: highlight banner on top of canvas + F6 panel.
- Active delegation: "assinando por delegação de X" badge on the signature footer.
- SoD: author NEVER sees the signature panel (server `viewer.eligible_for_active_stage`
  already excludes; FE renders facts, never derives).
- Empty/edge: route without active stage, cancelled instance — teaching copy PT-BR
  (C3 worklist pattern).

## 7. Interaction model

Single entry: worklist, dashboards, notifications → `/documents/:id`. Decision
(verdict/signature) → mutation → `QK.approval` invalidation → react-query refetches
instance → mode re-derives on its own → screen transitions (e.g. `reviewing` →
`observing` after verdict). Mode transition: 150–200ms crossfade;
`prefers-reduced-motion` → instant.

## 8. Content

PT-BR; mode labels above; verdicts in the timeline with display name (A1 supplies),
reason, timestamp; signature legal-effect copy unchanged (M2b); no UUID/hash leading
(readable accountability — hash lives in the integrity disclosure).

## 9. Ratified decisions (operator, 2026-07-08)

1. **Author comment participation during review — YES.** In `author-waiting` the author
   can reply to / resolve instance comments (never edit content). Rationale: backend
   holds freeze on unresolved comments (`HasUnresolvedInstanceComments`); author replies
   accelerate resolution; Google Docs model. **Caveat:** authz surface for author comment
   replies must be verified against the existing instance-comments capability model
   during the M2d feature spec — if a capability gap exists, it is an M2d contract item,
   not a client-side workaround.
2. **`author-editing` adopts the unified shell — YES.** `DocumentEditorPage`'s own
   `ArtifactMetaSidebar` composition is replaced by the same right-sidebar architecture as
   every other mode (meta + route preview as sidebar panels). More refactor inside M2d,
   accepted — it IS the single-screen point.
3. **Destination — `/documents/:id` canonical (Option 1.5) — YES.** Working screen owns the
   artifact's canonical URL; record surface moves unchanged to `/documents/:id/details`
   (+ `distribution` child); `/edit` and `/approvals/:documentId` redirect. Surface unified,
   domains modular (approval components compose in from `features/approval`); signature =
   ceremony within the screen. Reopen trigger: external-signer persona → standalone ceremony
   page. Approver/observer modes lazy-load editing chunks (bundle discipline — F2d.5 spec item).
   Working screen's sidebar meta panel carries a discoverable link to the record view
   (`/documents/:id/details`); worklist `?decision=` preselect deep-link survives on the new
   destination (`/documents/:id?decision=...`).

## 10. Implementation references

impeccable `product.md` (register rules), interaction/harden guidance for signature form
states, layout guidance for dense sidebar. All M2d FE feature specs cite this brief.
