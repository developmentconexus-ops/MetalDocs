# System-impact analysis — Surface tenant dictionary tokens in the template editor palette

**Date:** 2026-06-29
**Intent (one line):** Tokens created in the token dictionary (`/templates/tokens`) should appear in the template editor's "Tokens disponíveis" sidebar palette (and count as known tokens during validation).
**Work type:** feature
**Author:** developing-new-work skill
**Verdict:** 🟢 Green  *(see §10)*

> Same ten sections for module and feature work. Module-only rows are marked **N/A** with a reason.

---

## 1. Classify & own
*(CLAUDE.md Orientation rule)*

- **Work type:** feature (no new module; no new capability; no contract change).
- **Owning module(s):** `templates` (frontend) — owns the editor and the `AvailableTokensPanel` ([frontend/apps/web/src/features/templates/AvailableTokensPanel.tsx](frontend/apps/web/src/features/templates/AvailableTokensPanel.tsx)), the placeholder-catalog query ([queries/usePlaceholderCatalogQuery.ts](frontend/apps/web/src/features/templates/queries/usePlaceholderCatalogQuery.ts)), and the known/unknown-token classification in [TemplateEditorPage.tsx](frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx). The palette is a templates-FE surface.
- **Explicitly NOT owning:** `tokens` — owns the dictionary *data* and the `GET /api/v1/tokens` read endpoint, but not the editor palette. We consume its existing published HTTP surface; we do NOT move palette logic into the tokens module. `documents`/`render` — not touched; dictionary values are pinned at document creation (SP-2) and render transports pre-resolved values; this feature is authoring-time discovery only.
- **Cross-module edges (with direction):** `templates-FE → tokens-HTTP` (the editor reads `GET /api/v1/tokens`, the tokens module's published REST surface — already consumed by the SP-3 UI via the existing FE client + `QK.tokens`). No backend cross-module Go coupling is introduced. The two catalogs stay **separate endpoints** (`/templates/placeholder-catalog` computed; `/tokens` dictionary) — composed on the FE, never merged server-side. This preserves the "two complementary catalogs" design ([wiki/modules/tokens.md §1, §8.9](wiki/modules/tokens.md)).
- **Ambiguity?** None. Palette = templates FE; data = tokens HTTP. No **AS-3**.

## 2. Foundation verdict
*(Global-Maximum rule)*

- **Base you'd build on:** `AvailableTokensPanel` (clean, prop-driven, flat catalog + `onInsert(key)`); `usePlaceholderCatalogQuery` (computed catalog); the tokens FE client + `QK.tokens` query keys shipped by SP-3 (2026-06-29).
- **Sound, or legacy/patch/workaround?** Sound. Both surfaces were delivered to the current Grade-A bar (tokens module L3; SP-3 just shipped). Not a patch base.
- **If patchy:** N/A. The global-maximum shape is already the natural one — compose the two existing read endpoints on the FE into two palette sections, and unify the FE "known keys" set for validation. No backend reshaping needed. No **AS-2**.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | Yes (read) | `GET /tokens` gated `CapTokenView`; granted to author/editor/approver/qms_admin/viewer/system_admin ([reference-data 0001:124-129](db/reference-data/0001_product_reference_data.sql)). Every authoring role already holds it — no grant change. | existing tier-1/tier-2 on `/tokens` |
| Contract-first (OpenAPI + oapi-codegen) | No | No route added/changed; `GET /tokens` already exists and is typed. | generated `paths['/tokens']` types |
| Multi-tenant pooled | Yes | `GET /tokens` is session-tenant-scoped (`tenant.FromContext`); returns only the caller's tenant rows. | existing endpoint |
| Async = transactional outbox | No | Read-only discovery; no writes, no side effects. | N/A |
| DB enforces invariants | No | No schema change. | N/A |
| Cross-module via published interface only | Yes | FE consumes tokens' published REST endpoint; no SQL/repo/domain reach-in; catalogs not merged server-side. | tokens FE client (SP-3) |

Any violation → none. No **AS-1**.

## 4. Capability wiring
**N/A** — no new/changed capability. `CapTokenView` already exists and is granted to all authoring roles (§3).

## 5. Module wiring
**N/A** — feature, no new module.

## 6. Frameworks to reuse, not reinvent

- React Query + `QK.tokens` (SP-3) for the dictionary fetch — reuse, do not add a parallel fetcher.
- `usePlaceholderCatalogQuery` pattern for the computed catalog — unchanged.
- The existing `AvailableTokensPanel` `onInsert(key)` seam — dictionary tokens insert as plain `{name}` exactly like computed tokens (the `PHDictionary` classification is a templates save-time concern, not an insert concern).
- Generated API types (`components['schemas']`) for the `/tokens` DTO — no hand-rolled shapes.

## 7. Contract & data

- **OpenAPI-first:** no change. `GET /api/v1/tokens` and `GET /templates/placeholder-catalog` both already exist.
- **Migration:** none.
- **Destructive change?** None — purely additive (a second palette section + widened known-key set).

## 8. Test & QA plan
*(feature; FE-only)*

- **Canonical framework:** Vitest + RTL for `AvailableTokensPanel` (extend existing [AvailableTokensPanel.test.tsx](frontend/apps/web/src/features/templates/AvailableTokensPanel.test.tsx)) and the editor classification. No Go tests (no backend change).
- **QA gates that apply:** FE typecheck, vitest, build; runtime preview drive (author sees dictionary section; a dictionary token in the body is NOT flagged "não reconhecido"). Backend gates N/A.
- **Evidence shape:** preview screenshots (palette two sections; used/known badges), console clean, network shows `GET /tokens` 200. Note: local vitest is currently blocked by node_modules junction drift ([memory: fe-node-modules-junction-drift]) — runtime preview is the primary evidence.

## 9. Docs / ADR

- **Wiki:** refresh [wiki/modules/templates.md](wiki/modules/templates.md) (palette now sources two catalogs) and bump `Last verified`; cross-link from [wiki/modules/tokens.md §1](wiki/modules/tokens.md) ("complementary catalogs" now realized in the authoring UI). [wiki/concepts/placeholders.md](wiki/concepts/placeholders.md) if it enumerates palette sources.
- **REQ IDs cited:** REQ-AUTHZ-1/2 (capability gate on `/tokens`), REQ-TEN-1 (tenant-scoped read).
- **ADR required?** No. No MUST-deviation, no policy change. SP-2's "complementary catalogs" design (ADR 0049) already anticipated authoring-side surfacing; this realizes it.

## 10. Verdict & locked constraints

- **Verdict:** 🟢 Green (proceed to design).
- **Open hard-stops:** none (AS-1/AS-2/AS-3 all clear).
- **Locked constraints handed to brainstorming:**
  1. **Do NOT merge the catalogs server-side.** Keep `/templates/placeholder-catalog` (computed, system-global) and `/tokens` (dictionary, tenant-scoped, mutable) as separate endpoints; compose on the FE as two labeled sections. (Module boundary + tokens.md "complementary catalogs".)
  2. **Two visual sections, distinct semantics.** Computed = "Preenchido pelo sistema (seguro)" (auto-filled). Dictionary = tenant-defined constants — label them so authors understand the difference (one is system-computed, the other a tenant value pinned at creation).
  3. **Unify the FE known-key set for validation.** A dictionary token used in the body must classify as *known*, not "Tokens não reconhecidos" — fold both catalogs into `usedKeys`/`unknownTokens` classification.
  4. **No new capability / no contract change / no migration.** If the design grows any of these, re-enter this gate.
  5. **Insert stays uniform** (`{name}` text via `onInsert`); `PHDictionary` schema classification is a separate templates save-time concern, out of scope for the palette itself.
