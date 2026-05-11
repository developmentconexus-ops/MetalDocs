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

### next-code-preview
**Context:** Step 2 (Identidade) shows code preview card "TPL-POP-009".
**Blocked by:** No `GET /api/v2/templates/next-code?profile=<CODE>` endpoint. Currently mocked client-side as `TPL-{PROFILE}-XXX` (placeholder digits).
**File:** `frontend/apps/web/src/features/templates/components/wizard/steps/StepIdentity.tsx`
**TODO tag:** `TODO(novo-template-wizard:next-code-preview)`
**Resolution:** When endpoint ships, replace mock with `useNextTemplateCodeQuery(profileCode)`. Show loading + error states.

---

### key-generation
**Context:** Backend `POST /api/v2/templates` requires `key` field (unique identifier). Design has no key input — derived auto from name.
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

### step3-docx-upload
**Context:** Step 3 lets user pick `.docx` as starting point. Real upload requires presigned URL flow — but template `id`+`n` only exist after Step 5 create.
**Blocked by:** Wizard ordering — `POST /api/v2/templates` runs at Step 5. Upload flow `POST /api/v2/templates/{id}/{n}/docx-upload-url` then PUT then `PUT /draft` happens **after** create.
**File:** `frontend/apps/web/src/features/templates/components/wizard/steps/StepStructure.tsx`
**TODO tag:** `TODO(novo-template-wizard:step3-docx-upload)`
**Resolution:** Step 3 currently captures `selectedDocxName` + `selectedDocxSize` only (filename echo, no upload). After Step 5 create, if `startingPoint === 'docx'`, redirect to editor with `?import=<file-blob-ref>` so editor performs presigned upload + draft save. Or stage file in IndexedDB between Step 3 and Step 5.

---

### step3-placeholder-extract
**Context:** Original design showed 7 placeholder chips after upload. Cut from impl.
**Blocked by:** No backend endpoint extracts tokens without publishing. `POST /publish` returns `missing_tokens` / `orphan_tokens` but is destructive.
**File:** N/A (cut at Phase 0).
**TODO tag:** `TODO(novo-template-wizard:step3-placeholder-extract)`
**Resolution:** Add `POST /api/v2/templates/{id}/{n}/extract-tokens` that runs docgen-v2's parser without publishing. Then Step 3 can preview placeholders inline. Auto-fill flag also requires schema metadata not yet defined — design separately.

---

### step3-editor-handoff
**Context:** After Step 5 create, user expects to land in editor (esp. for `'blank'` start, where there's no real docx to import).
**Blocked by:** No editor flow defined for templates wizard handoff. Today, `/templates-v2/new` Step 5 just calls create and (presumably) redirects to list.
**File:** `frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx` (Step 5 submit handler — not yet implemented).
**TODO tag:** `TODO(novo-template-wizard:step3-editor-handoff)`
**Resolution:** After successful create, redirect to `/templates-v2/<id>/edit?import=<blob-ref|blank>` based on `startingPoint`. Editor handles real upload (docx case) or stub-blank schema.json (blank case).

---

### permissions-roles-api
**Context:** Step 4 (Permissões) "Por funções" mode shows personnel role cards (QUA-INSP, QUA-ANA, etc.) with user counts. All mocked.
**Blocked by:** No personnel/roles endpoint. Taxonomy API exposes document-type profiles (POP, IT, etc.) — not user roles.
**File:** `frontend/apps/web/src/features/templates/components/wizard/steps/StepPermissions.tsx` (`MOCK_ROLES` constant)
**TODO tag:** `TODO(novo-template-wizard:permissions-roles-api)`
**Resolution:** When a user-roles/personas endpoint ships, replace `MOCK_ROLES` with `useRolesQuery()`. Wire loading + error states.

---

### permissions-area-counts
**Context:** Step 4 "Por área" mode shows area cards with user count per area (28, 89, 34…). Counts are mocked.
**Blocked by:** No user-count-per-area aggregate endpoint.
**File:** `frontend/apps/web/src/features/templates/components/wizard/steps/StepPermissions.tsx` (`MOCK_AREAS` constant)
**TODO tag:** `TODO(novo-template-wizard:permissions-area-counts)`
**Resolution:** Replace mock counts with real aggregate when API ships. Area names can come from existing `useAreasQuery`.

---

### permissions-user-count
**Context:** Step 4 "Todos" mode shows "~340 usuários ativos". Mocked constant.
**Blocked by:** No company-wide active user count endpoint.
**File:** `frontend/apps/web/src/features/templates/components/wizard/steps/StepPermissions.tsx` (`COMPANY_USER_COUNT` constant)
**TODO tag:** `TODO(novo-template-wizard:permissions-user-count)`
**Resolution:** Replace `COMPANY_USER_COUNT` with API-fetched total when endpoint ships.

---

### confirmacao-backend-submit
**Status: RESOLVED 2026-05-10.**
`TemplateWizardPage.handleSubmit` now calls `POST /api/v2/templates { key, name, description }` and redirects to `/templates-v2/<id>/versions/<n>` on success. Error state surfaces inline in `StepConfirmation` via `submitError` prop.

---

### template-create-visibility-api
**Context:** `POST /api/v2/templates` generated handler (`routes_generated.go`) only accepts `key`, `name`, `description?`, `doc_type_code?`. It hardcodes `Visibility: VisibilityPublic` and `ApproverRole: "approver"`. The wizard collects permissions (Step 4: by area / by role / all-company) and structure origin (Step 3: blank / docx) — none of those are forwarded to the create API.
**Blocked by:** Backend API contract. The generated OpenAPI spec (`api.gen.go` `CreateTemplateV2JSONBody`) does not expose `visibility`, `areas`, `specific_areas`, or `approver_role` in the create body.
**Files:**
- Backend: `internal/modules/templates_v2/api/api.gen.go` (`CreateTemplateV2JSONBody`)
- Backend: `internal/modules/templates_v2/delivery/http/routes_generated.go` (`CreateTemplateV2`)
- Frontend: `frontend/apps/web/src/features/templates/api/templatesV2.ts` (`createTemplate`)
- Frontend: `frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx` (`handleSubmit`)
**Resolution:**
1. Backend team expands OpenAPI spec to include `visibility`, `areas`, `specific_areas` (and optionally `approver_role`) in the create body — regenerate `api.gen.go`.
2. Frontend updates `createTemplate` type to include those fields.
3. `TemplateWizardPage.handleSubmit` passes `deriveVisibility(state.permissionsMode)` and area IDs. `selectedRoleIds` still needs a backend field (no `roles` parameter exists in domain model).
4. `permissionsMode === 'roles'` → either map to `specific_areas` or wait for a dedicated `roles` field in the domain.

---

## Steps not yet implemented

| Step | Name | Status |
|---|---|---|
| 2 | Identidade | **Done** (2026-05-09) |
| 3 | Estrutura | **Done** (2026-05-09) — mocked DOCX flow; real upload + placeholder extract deferred (see backlog items above) |
| 4 | Permissões | **Done** (2026-05-09) — mocked roles/areas/counts; real API deferred (see backlog items above) |
| 5 | Confirmação | **Done** (2026-05-10) — submit wired to API; redirects to editor on success. Visibility/permissions deferred: `template-create-visibility-api` |
