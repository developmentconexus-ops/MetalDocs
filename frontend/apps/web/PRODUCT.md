# Product

## Register

product

## Users

Quality/document-control professionals in metallurgy manufacturing (QMS operators):
document authors, approvers (quality managers/engineers), and system admins. They work
in an authenticated, task-driven app — authoring controlled documents, routing them for
approval with e-signature, distributing published revisions. ISO-9001-style compliance
context: every action is auditable; users are accountable for what they sign.

## Product Purpose

MetalDocs is a controlled-document management system (eQMS document control): create
documents from templates, revise with reason-for-change, submit through approval routes,
sign electronically, publish and distribute. Success = a compliant document lifecycle a
quality team trusts enough to replace paper/DocX + email chains.

## Brand Personality

Sober, precise, accountable. Wine palette (deep bordeaux `--brand: #6b1f2a` on warm
neutral surfaces) with Inter Tight; feels like a serious industrial compliance tool,
not a playful SaaS. PT-BR product language.

## Anti-references

- Generic Tailwind-slate admin dashboards (the approval feature currently drifts into
  exactly this — the `--slate-*` secondary palette is tolerated legacy, not endorsed).
- Consumer e-sign marketing gloss (DocuSign landing-page energy) — this is a workbench.
- AI-slop dashboard scaffolding: hero metrics, identical card grids, gradient text.

## Design Principles

1. **The document is the protagonist** — screens exist to read, judge, and act on a
   controlled document; chrome stays out of the way.
2. **One visual vocabulary** — Wine tokens (`src/styles/tokens.css`) everywhere; same
   button/badge/dialog shapes across features (ADR 0053 shared controlled-artifact layer).
3. **Accountability legible** — status, revision, who/when/why are always visible and
   human-readable; raw technical identifiers (hashes, UUIDs) don't lead.
4. **Density with hierarchy** — professionals scan queues and tables; compact, but every
   screen has one primary action.
5. **State-complete components** — loading (skeleton), empty (teaching), error, disabled
   are designed, not afterthoughts.

## Accessibility & Inclusion

WCAG AA target: ≥4.5:1 body-text contrast, full keyboard operability, visible focus,
`prefers-reduced-motion` respected. PT-BR copy throughout; monospace only for codes
(revision badges, document codes).
