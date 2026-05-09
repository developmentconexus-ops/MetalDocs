# Backlog: Novo Template Wizard

> Created: 2026-05-09
> Feature: `/templates-v2/new` (5-step wizard)
> Worksheet: `frontend/apps/web/design-source/novo-template-escopo/IMPLEMENTATION.md`

---

## Deferred items

### template-counts
**Context:** Step 1 (Escopo) shows `— templates` per profile card.
**Blocked by:** No summary/aggregate endpoint — `GET /api/v2/templates` returns list, no per-profile count.
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

## Steps not yet implemented

Steps 2–5 of the wizard are stubs (placeholder "Em breve" content). Track implementation in separate tasks.

| Step | Name | Notes |
|---|---|---|
| 2 | Identidade | Name, description, version prefix |
| 3 | Estrutura | Block sections configuration |
| 4 | Permissões | Visibility / approval workflow |
| 5 | Confirmação | Summary + `POST /api/v2/templates` |
