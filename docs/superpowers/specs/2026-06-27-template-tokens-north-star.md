# Template Tokens — North-Star Architecture

**Date:** 2026-06-27
**Status:** North-star (program-level direction). Increments ship as their own spec → plan → implementation cycles.
**Supersedes / revisits:** ADR 0008 (fixed placeholder catalog) — see §7.

---

## 1. Problem & intent

Template authors need to place **tokens** in a template body so that, when a controlled
document is produced, those tokens are filled with trustworthy values. Two intents:

1. **Secure auto-fill.** Values the user must *not* type — `author`, `doc_code`,
   `revision_number`, `effective_date`, `approvers`, etc. The system pulls them so they
   cannot be forged or mistyped. (This is the original stated goal.)
2. **Author-defined reusable tokens.** Authors define tenant-wide constants such as
   `{company_slogan}` once, map them centrally, and reuse them across every template.

Today only intent (1) exists server-side, and the authoring surface for it is broken
(see §6). Intent (2) does not exist. This document defines the end-state and the
increments that get us there without architectural drift.

## 2. Binding constraints (non-negotiable)

- **Layering: MetalDocs → adapter → vendor.** The app speaks MetalDocs language
  (token keys, catalog entries). The adapter (`@metaldocs/editor-ui`) translates to/from
  Eigenpal / docxtemplater. The app MUST NOT reach the vendor directly (no
  `as any` on the editor ref, no docxtemplater `{}` syntax or `TemplateTag` leaking into
  app code).
- **Wire contract = docxtemplater `{token}` text tags.** Substitution happens
  server-side. Content controls are explicitly *not* adopted (they would blast the
  forensically-pinned `render/*` renderer). The token in the stored `.docx` is always
  literal `{key}` text.
- **Capability-oriented authz, not roles.** Every new operation maps to a typed
  capability in the IAM registry (`internal/modules/iam/domain/model.go`
  `validCapabilities`), scope-classified in `capability_scope.go`, route-mapped in the
  tier-1 rules, seeded in `db/reference-data/0001_product_reference_data.sql`, and held to
  the existing guard tests (every-cap-classified / in-registry / seeded-or-deferred /
  deny-by-default). "Who can manage tokens" = "who holds the capability".
- **Never persist a resolved value over its token.** Any draft-time preview of a
  resolved value is display-only (decoration), never written back into the stored buffer.
  (ADR 0008 landmine: 1500 ms autosave would otherwise destroy the `{key}` and lose
  re-resolvability.)
- **Runtime truth beats docs.** The mature parts (`render/resolvers/*`, `fillin_service`,
  fanout `{token}` substitution) stay authoritative; increments build on them.

## 3. End-state model

Two token classes. Both are placed as docxtemplater `{key}` text, both are validated
against an authoritative registry, both are substituted server-side at freeze.

| Class | Value source | Who sets it | Status |
|---|---|---|---|
| **Computed** (`author, doc_code, revision_number, effective_date, approvers, controlled_by_area, approval_date, doc_title`) | `render/resolvers/*` at freeze | system (resolver registry) | ✅ built |
| **Tenant dictionary** (`company_slogan`, …) | tenant-scoped name→value, capability-governed | whoever holds `token_dictionary.manage` | ❌ new |

> A third pre-existing class — **per-document fill-in** placeholders (text/date/select the
> *document creator* fills, via `fillin_service`) — is orthogonal and out of this program's
> scope except where the authoring surface naturally exposes it.

The **authoritative token registry** at any point is `computed catalog ∪ tenant dictionary`.
A `{key}` in a template body that is not in that union is an error surfaced as early as
possible (authoring warning) and caught at the latest by fanout `UnreplacedVars`.

## 4. Components (end-state)

- **Catalog endpoint** (`templates/.../routes_catalog.go`) — the computed catalog. Exists.
- **Tenant dictionary** (new module surface) — domain + tenant-scoped storage + repo +
  capability-gated CRUD + audit.
- **Token resolution at freeze** (`render/*` + fanout) — computed resolvers (exist) plus
  injection of dictionary values into `ResolvedValues`.
- **Tag validation** — template save validates referenced tokens ⊆ registry.
- **Adapter token capability** (`@metaldocs/editor-ui`) — `insertToken(key)` /
  `getUsedTokens()` in MetalDocs language; vendor translation hidden.
- **Authoring surface** (`features/templates`) — the "available tokens" palette: discover +
  insert + validate.
- **Dictionary management UI** — capability-gated CRUD surface.
- **Draft-time preview** — decoration-only resolution of the resolvable subset.

## 5. Decomposition (dependency-ordered sub-projects)

Each is its own spec → plan → implementation cycle.

- **SP-0 — Available-tokens palette over the computed catalog.** *(detailed spec:
  `2026-06-27-sp0-available-tokens-palette-design.md`)* Frontend + adapter only; no new
  caps, no backend change. Fixes the live gambiarra. **Independent — built first.**
- **SP-1 — Tenant dictionary backend.** Domain + tenant-scoped storage + repo + new IAM
  capabilities (`token.view`, `token_dictionary.manage`) with scope/seed/guard wiring +
  capability-gated CRUD + audit. No render yet.
- **SP-2 — Freeze substitution + tag validation.** Inject dictionary values into
  `ResolvedValues`; validate template tokens ⊆ registry. Depends on SP-1.
- **SP-3 — Dictionary management UI.** Capability-gated CRUD surface. Depends on SP-1.
- **SP-4 — Authoring surface absorbs dictionary tokens.** Add the dictionary as a second
  data source to the SP-0 palette. Depends on SP-0 + SP-1. Small.
- **SP-5 — Draft-time token preview.** Resolve the *resolvable subset* (author, doc_code,
  date — known at draft) and show real values via **decoration overlay, never persisted**.
  Resolvable-when matrix is part of its spec. Independent of the dictionary; can follow SP-0.

```
SP-0 ──────────────► SP-4
                      ▲
SP-1 ─► SP-2          │
   └──► SP-3          │
   └───────────────► SP-4
SP-5  (independent; needs a draft-resolve API)
```

**Recommended order:** SP-0 → (SP-5 and SP-1 in either order) → SP-2 → SP-3 → SP-4.
SP-0 first: immediate value, zero blast radius, de-risks the native-primitive approach.

## 6. Why the current implementation is a local maximum

`TemplateEditorPage.tsx` ignores the adapter and the vendor's native `templatePlugin`:

- `(editorRef.current as any).getAgent().getVariables()` — reaches **MetalDocs → vendor**
  directly (violates §2 layering), calls an API the adapter does not expose, and resolves
  to `[]` every time. Detection is dead.
- That dead `[]` is written into `localSchemas.placeholders` and fed to the 400 ms schema
  autosave → latent **schema corruption** (zeroes the persisted placeholder schema).
- `readHeadings` / custom outline ride the same dead `getAgent` pattern.
- `PlaceholderCatalogPanel` is a passive list with inline styles, a duplicate catalog
  fetch, and a "detected" highlight that never fires.

The vendor already provides the right primitives natively: `templatePlugin` detection +
canvas chips, `getTemplateTags(state)`, and `insertTemplateVariable`. The global maximum
is to **bridge MetalDocs' catalog to those native primitives through the adapter**, not to
hand-roll detection in app code.

## 7. Relationship to ADR 0008

ADR 0008 ("fixed placeholder catalog", 7 computed tokens, author types raw `{token}`,
no user-fill) predates Eigenpal adoption. It remains correct about **computed tokens being
server-resolved and secure**. It is **revisited** in two ways:

1. Authors no longer hand-type raw `{token}` — they insert from a palette (SP-0), with the
   raw-text path remaining valid but no longer the primary UX.
2. The "fixed catalog, no author-defined tokens" stance is superseded by the tenant
   dictionary (SP-1+), which adds author/tenant-defined tokens under capability governance.

A formal ADR superseding 0008 should be written when SP-1 lands (it is the point where the
catalog stops being fixed). SP-0 alone does not require it.
