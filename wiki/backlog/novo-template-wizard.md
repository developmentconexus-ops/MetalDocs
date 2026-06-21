# Backlog: Novo Template Wizard

> Last verified: 2026-06-21 (verify-and-archive sweep; see _cleanup-2026-06-21.md)
> Created: 2026-05-09
> Feature: `/templates/new` (4-step wizard)
> Worksheet: `frontend/apps/web/design-source/novo-template-escopo/IMPLEMENTATION.md`

---

## Deferred items

### template-counts
**Context:** Step 1 (Escopo) shows `— templates` per profile card.
**Blocked by:** No summary/aggregate endpoint — `GET /api/v1/templates` returns list, no per-profile count.
**File:** `frontend/apps/web/src/features/templates/components/wizard/steps/StepScope.tsx`
**TODO tag:** `TODO(novo-template-wizard:template-counts)`
**Resolution:** When API ships a count field or aggregate endpoint, replace `—` placeholder with real count.

---

### chk-disabled
**Context:** CHK profile is disabled in Step 1 (`DISABLED_PROFILES = new Set(['CHK'])`).
**Blocked by:** Taxonomy API does not expose an `enabled` flag per profile. CHK disabled until Checklist feature ships.
**File:** `frontend/apps/web/src/features/templates/components/wizard/steps/StepScope.tsx`
**TODO tag:** `TODO(novo-template-wizard:chk-disabled)`
**Resolution:** When taxonomy API exposes `profile.enabled`, remove hardcode. Drive disabled state from API flag.

---

### next-code-preview
**Context:** Step 2 (Identidade) shows code preview card "TPL-POP-009".
**Blocked by:** No `GET /api/v1/templates/next-code?profile=<CODE>` endpoint. Currently mocked client-side as `TPL-{PROFILE}-XXX` (placeholder digits).
**File:** `frontend/apps/web/src/features/templates/components/wizard/steps/StepIdentity.tsx`
**TODO tag:** `TODO(novo-template-wizard:next-code-preview)`
**Resolution:** When endpoint ships, replace mock with `useNextTemplateCodeQuery(profileCode)`. Show loading + error states.

---

### key-generation
**Context:** Backend `POST /api/v1/templates` requires `key` field (unique identifier). Design has no key input — derived auto from name.
**Blocked by:** No design decision on key UX. Auto-slug from name is fragile (collisions, edits break links).
**File:** Step 5 (Confirmação) submit handler — `frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx`
**TODO tag:** `TODO(novo-template-wizard:key-generation)`
**Resolution:** Either (a) add advanced "Identificador técnico" field with auto-suggest + manual override, or (b) backend generates key from name server-side. Decide before Step 5 implementation.

---

### font-size-hero
**Context:** Step 2 code preview value uses raw `font-size: 26px` — sits between `--font-size-lg` (22px) and `--font-size-xl` (32px). No matching token.
**File:** `frontend/apps/web/src/features/templates/components/wizard/steps/StepIdentity.module.css` (`.codePreviewValue`)
**TODO tag:** `TODO(novo-template-wizard:font-size-hero)` — captured as `/* design-exact */` comment in CSS.
**Resolution:** Either (a) add `--font-size-hero: 26px` to `tokens.css` if 26px is a recurring hero-numeric pattern, or (b) accept as design-exact magic number for this single use.

---

### step3-placeholder-extract
**Context:** Original design showed 7 placeholder chips after upload. Cut from impl.
**Blocked by:** No backend endpoint extracts tokens without publishing. `POST /publish` returns `missing_tokens` / `orphan_tokens` but is destructive.
**File:** N/A (cut at Phase 0).
**TODO tag:** `TODO(novo-template-wizard:step3-placeholder-extract)`
**Resolution:** Add `POST /api/v1/templates/{id}/{n}/extract-tokens` that runs docgen-v2's parser without publishing. Then Step 3 can preview placeholders inline. Auto-fill flag also requires schema metadata not yet defined — design separately.

---




## Active wizard steps

Current runtime has four steps. Historical rows below for the old 5-step implementation are superseded by the 2026-05-17 checkpoint: Step 3 supports blank + DOCX import handoff, Step 4 is Confirmação, and the former `Permissões` step was removed.

| Step | Name | Status |
|---|---|---|
| 2 | Identidade | **Done** (2026-05-09) |
| 3 | Estrutura | **Done** (2026-05-17) - blank start and real DOCX import handoff to Eigenpal |
| 4 | Confirmação | **Done** (2026-05-17) - create, optional DOCX import, redirect to editor |

