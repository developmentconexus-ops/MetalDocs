# Feature F4.2 — Spec

> **Milestone:** 4 — Documento Publicado completion + Documento Obsoleto  ·  **Folder:** `f4.2-obsoleto-variant`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-23 / leandrotca — *contract decisions below confirmed via interview; no implementation begins until this line is filled.*

> The feature's **contract**, approved **before any code**. Defines what the obsolete variant must do
> and how it is proven — not how it is built (`plan.md`). The milestone-validator judges against *this* file (C1).

## Interview record (fail-closed gate)

Driven via `superpowers:brainstorming` as the interview engine, seeded with the F4.2 row of `../milestone.md`.
Recon established the obsolete state is already a **status-driven branch of `DocumentPublishedPage`** (no fork):
`getDocumentStatusPresentation` has an `'obsolete'` case, and an `isObsolete` watermark banner already renders
(`.obsoleteStamp` CSS is a faithful port of the design's rotated OBSOLETO overlay). The genuine open
decisions — parity depth and how/whether to gate the obsolete "Visualizar" action — were put to the operator.

| # | Question | Answer |
|---|----------|--------|
| 1 | Parity depth vs `design-source/documento-obsoleto` (page-wide grayscale 0.65 + opacity 0.85 dim, hidden "vigente" pill, all hero buttons disabled, rotated OBSOLETO watermark). Current page has only the watermark + status text. | **Full design parity.** Apply the page-wide grayscale+opacity dim, hide the green "vigente" pill, disable the hero actions, keep the rotated OBSOLETO watermark (already ported). |
| 2 | On an obsolete doc, should "Visualizar documento" stay enabled? Design disables every hero button; QMS norms keep obsolete docs viewable for audit. | **Enabled, but capability-gated** — viewable only by users holding a privileged capability, not everyone. Other hero actions (Baixar PDF, Iniciar revisão, Copiar link) are disabled when obsolete. |
| 3 | Which capability gates obsolete "Visualizar"? It must use the dynamic roles+capabilities system (`useHasCapability`), like all other permissions — not a hardcoded role list. | The system is capability-based (roles are composed of assignable capabilities); the gate must check a **capability**, not role names. |
| 4 | No dedicated "view-obsolete" capability exists in the backend capability catalog (`internal/modules/iam/domain/model.go:97`). Canonical doc caps: `document.view` (tenant-grade, gates all viewing), `edit, create, submit, signoff, publish, supersede, obsolete, docx, rename`. A new cap = backend work (HS-2, out of frontend-only milestone). Which to use? | **Reuse `document.obsolete`.** Gate obsolete "Visualizar" on `useHasCapability('document.obsolete')` — only users who manage obsoletion see it enabled. Existing cap, no backend change, frontend-only, fits "not everyone". Backend stays the sole authz boundary (UX hint only). |

## Consumer contract (FIRST — before any producer)

F4.2 has **no new producer**. It consumes two already-shipped, contract-frozen inputs and reuses the
existing `DocumentPublishedPage` component (no second page file — reuse is itself an acceptance criterion).

- **Consumer:** `DocumentPublishedPage` (`features/documents/pages/DocumentPublishedPage.tsx`) — the root
  container, the hero "vigente" pill, the hero action buttons, rendered at route `documents/:documentId`
  (`features/documents/routes.tsx:25`).
- **Contract — obsolete state:** `DocumentResponse.status === 'obsolete'` (already on the payload via
  `useDocumentDetailQuery`). The obsolete variant is driven by this **real status field**, not a prop hack
  or a forked route. Source of truth: `lib/api-types/index.d.ts` `DocumentResponse.status`; openapi.yaml
  documents status enum (includes `obsolete`).
- **Contract — capability gate:** `user.capabilities: string[]` (on `CurrentUser`,
  `lib/types/index.ts:39`), consumed via the existing `useHasCapability('document.obsolete')` hook
  (`features/iam/hooks/useHasCapability.ts`). Precedent for capability-as-UX-hint:
  `features/templates/lib/canActOnVersion.ts` (header: "UX hint only — the backend remains the sole
  authorization boundary", `wiki/concepts/authz-tiers.md`). `document.obsolete` is a canonical capability in
  the backend catalog — verified: 4 backend literals incl. `obsolete_service_test.go:365`
  (`denied.Capability == "document.obsolete"`).
- **Contract — design parity:** `design-source/documento-obsoleto/documento-obsoleto.html` →
  `onda1-v5.jsx` `PublicadoV5 obsolete={true}` (the obsolete treatment: root `grayscale(0.65)`+`opacity 0.85`
  at onda1-v5.jsx:746-747; `{!obsolete && <vigente pill>}` at :190; `disabled={obsolete}` on all hero
  buttons at :220-230; rotated OBSOLETO overlay at :805-819).

## What this feature implements

On the published-document screen when `status === 'obsolete'` (route `documents/:documentId`):

1. **Page-wide dim** — the root container renders the design's obsolete treatment
   (`grayscale(0.65)` + `opacity 0.85`) via a conditional CSS-module class, so the whole screen reads as
   superseded. Removed/absent when not obsolete.
2. **"Vigente" pill hidden** — the green `vigenteBadge` (current-version pill) is not rendered when obsolete
   (a document cannot be both obsolete and "vigente").
3. **Hero actions gated** —
   - **Visualizar documento:** enabled only when `useHasCapability('document.obsolete')` is true; otherwise
     `disabled` with a Portuguese tooltip naming the missing capability (mirrors `canActOnVersion` reason
     copy). Backend `document.view` remains the real boundary.
   - **Baixar PDF** and **Copiar link:** `disabled` when obsolete.
   - **Iniciar revisão / Publicar:** already disabled for obsolete status (not published/approved); kept
     consistent.
4. **Rotated OBSOLETO watermark** — unchanged (the existing `.obsoleteBanner`/`.obsoleteStamp` overlay is a
   faithful port of the design); confirmed still rendered.
5. **One component** — all of the above is a `status === 'obsolete'` branch inside the existing
   `DocumentPublishedPage`. No new page file, no forked route.

## Non-goals (mandatory)

- **No new backend capability / Go source change.** Reuses the existing `document.obsolete` cap. A dedicated
  "view-obsolete" capability is explicitly out of scope (HS-2 — would be a separate full-stack feature).
- **No second Obsoleto page / route component.** Obsolete is a status-driven branch; forking is the anti-goal (R2).
- **No Publicado layout restyle** beyond the obsolete dim/disable treatment — gaps only.
- **No change to the F4.1 wirings** (coverage, PDF client, páginas/tamanho) — F4.2 only adds the obsolete
  presentation on top.
- **No backend enforcement claim.** The capability gate is a UX hint; F4.2 does not assert backend authz.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| A doc with `status:'obsolete'` renders the OBSOLETO watermark + obsolete status presentation, driven by real `status` | new `describe('F4.2 — obsolete variant')` in `DocumentPublishedPage.test.tsx`: fixture `status:'obsolete'` ⇒ OBSOLETO text present, obsolete subtitle/badge present | fixture (vitest) |
| Root applies the obsolete dim treatment only when obsolete | vitest: obsolete fixture ⇒ root carries the obsolete class; non-obsolete (published) fixture ⇒ it does not | fixture (vitest) |
| The "vigente" pill is hidden when obsolete | vitest: obsolete fixture ⇒ no "vigente" text; published fixture ⇒ "vigente" present | fixture (vitest) |
| Visualizar is enabled with `document.obsolete` cap and disabled without it (obsolete only) | vitest: obsolete + caps `['document.obsolete']` ⇒ Visualizar enabled; obsolete + caps `[]` ⇒ Visualizar disabled (+ tooltip) | fixture (vitest) |
| Baixar PDF + Copiar link are disabled when obsolete | vitest: obsolete fixture ⇒ both buttons `disabled` | fixture (vitest) |
| Obsolete path reuses `DocumentPublishedPage` — no new page file | `ls features/documents/pages` shows no new obsolete page; grep confirms the obsolete branch lives in `DocumentPublishedPage.tsx` | real (fs + grep) |
| Generated types consumed directly; capability via `useHasCapability`; no hand-written mapper | `npx tsc --noEmit` → exit 0 | real |
| No F4.1 regression | re-run `DocumentPublishedPage.test.tsx` + `documentDetailMeta.test.ts` ⇒ all green | fixture (vitest) |
| Both reviewer agents APPROVE | `frontend-screen-reviewer` (visual parity vs `design-source/documento-obsoleto`) + `frontend-code-reviewer` reports on record | real (review) |

> TDD: failing vitest case first, then wire to green. Vitest cases are fixture-level (mocked query +
> `useHasCapability`) — they prove the consumer branch + capability gate wiring, not backend enforcement
> (which is the backend's boundary, per `wiki/concepts/authz-tiers.md`). The reuse (no-new-page) and tsc
> checks are real.

## ADR needed?

- [x] No durable architectural decision — skip. F4.2 consumes existing contracts (`DocumentResponse.status`,
  `document.obsolete` capability) and reuses an existing component. The capability-as-UX-hint pattern is
  already governed by `wiki/concepts/authz-tiers.md` + the `canActOnVersion` precedent. The reuse-of-cap
  decision is recorded in the interview table above; no new authz primitive is introduced.
