# Backlog: Novo Template Wizard

> Created: 2026-05-09
> Feature: `/templates/new` (5-step wizard)
> Worksheet: `frontend/apps/web/design-source/novo-template-escopo/IMPLEMENTATION.md`

---

## Integration Audit - Plan 12.4 (2026-05-15)

**Scope:** one-screen finalization for the Template Creation Wizard at `/templates/new`.

**Design-source reality:** the requested `frontend/apps/web/design-source/novo-template-wizard/` directory does not exist. The committed Plan 12 artifacts are split across `frontend/apps/web/design-source/novo-template-*`, with the canonical worksheet at `frontend/apps/web/design-source/novo-template-escopo/IMPLEMENTATION.md`.

**Pre-code gates:**
- `scripts/check-system-runnable.ps1 -TargetRoute /api/v1/templates` - PASS (`auth/login`, `auth/me`, and `GET /api/v1/templates` returned HTTP 200).
- `scripts/check-module-contract-sync.ps1 -Module templates` - PASS (runtime owner, OpenAPI, backend generated package, frontend generated types, and feature wrapper present; manual drift review still required).

| Item | Source | Runtime/API reality | Frontend reality | Classification | Action |
|---|---|---|---|---|---|
| Wizard route `/templates/new` | routes + design | Route is frontend-only; primary API route is `POST /api/v1/templates` / `GET /api/v1/templates` | Route exists in `features/templates/routes.tsx` | implemented and aligned | keep |
| Step 1 profile scope | design + code | `GET /api/v1/taxonomy/profiles` exists as raw taxonomy route | `useProfilesQuery` uses taxonomy API + TanStack Query | implemented and aligned | keep |
| Profile-specific template creation | design + OpenAPI | `POST /api/v1/templates` accepts `doc_type_code?` | Wizard selected a profile but submit did not send `doc_type_code` | screen-local integration fix | include: pass selected profile code on create |
| Generic template creation | design + OpenAPI | Omitting `doc_type_code` creates generic/public template | Wizard supports `scopeType === "generic"` | implemented and aligned | keep |
| Template create API wrapper | frontend structure | `POST /api/v1/templates` is idempotency-wrapped and surfaced through shared client helpers | `createTemplate` used raw `fetch` on the wizard submit path | implemented but legacy-wired | include: switch this active create path to `apiFetch` + `Idempotency-Key` |
| Step 2 name + description | design + OpenAPI | `name` and `description?` exist in create body | Reducer and Step 2 are wired | implemented and aligned | keep |
| Step 2/5 `TPL-{PROFILE}-XXX` code preview | design + backlog | No next-code endpoint; backend create uses `key` from request | UI showed fake `TPL-*` values | screen-local integration fix | include: show the actual slug key preview instead of fake sequence |
| `key` UX | backlog | Backend requires `key`; server-side key generation does not exist | Wizard auto-slugs from name | defer | preserve blocker; do not expand product UX beyond honest slug preview |
| Step 3 blank start | design + runtime | Create endpoint creates a draft version; editor route exists | Wizard can redirect to editor after create | implemented and aligned | keep |
| Step 3 DOCX import | design + backlog | Presigned upload requires created template/version; no mid-wizard upload or handoff contract | UI accepted a local file name only | missing backend capability | defer; include local cut by disabling DOCX import in this screen |
| Step 3 placeholder extraction | design + backlog | No non-destructive extract-tokens endpoint | Cut from implementation | missing backend capability | defer |
| Step 4 public visibility | runtime + backlog | `CreateTemplate` hardcodes `VisibilityPublic` and `ApproverRole: "approver"` | Wizard exposed role/area/all choices that are ignored by backend | screen-local integration fix | include: replace mocked chooser with honest read-only current visibility |
| Step 4 roles mode | design + backlog | No personnel role endpoint or create-body roles field | Mock role cards and user counts existed | missing backend capability | defer; remove fake active UI |
| Step 4 areas mode | design + backlog | Taxonomy areas exist, but create body has no visibility/areas fields and no user-count aggregate | Mock area cards and counts existed | missing backend capability | defer; remove fake active UI |
| Step 4 company user count | design + backlog | No company-wide active-user-count endpoint | Mock `340` existed | missing backend capability | defer; remove fake count |
| Step 5 create + redirect | code + runtime | Runtime returns created template/version and editor route is `/templates/{id}/versions/{n}` | Submit calls create and redirects | implemented and aligned | keep after wrapper/profile-code fixes |
| Template count per profile card | backlog | No aggregate endpoint | UI shows `-` placeholder with TODO | missing backend capability | defer |
| CHK disabled state | backlog | Taxonomy profile API has no `enabled` flag | Hardcoded disabled `CHK` | defer | preserve TODO, do not broaden taxonomy/API |

**Ready for implementation:**
- Pass `doc_type_code` for profile-scoped templates.
- Use the shared API client and an idempotency key on the active `createTemplate` submit path.
- Replace fake `TPL-*` sequence previews with the actual slug key already submitted today.
- Disable unsupported DOCX import and replace mocked Step 4 permissions with an honest current public-visibility state.

**Prerequisites / backend capabilities preserved:**
- `next-code-preview`
- `key-generation` UX decision
- `step3-docx-upload`
- `step3-placeholder-extract`
- `step3-editor-handoff` richer import handoff
- `permissions-roles-api`
- `permissions-area-counts`
- `permissions-user-count`
- `template-create-visibility-api`
- `template-counts`
- `chk-disabled`

**Verification needed next:**
- `cd frontend/apps/web; pnpm.cmd tsc --noEmit -p tsconfig.build.json`
- `cd frontend/apps/web; pnpm test`
- Runtime smoke for `/templates/new` including profile-scoped create payload, blank-start flow, disabled DOCX import, and read-only permissions step.
- Screenshots for PR evidence.

**Implementation result (2026-05-15):**
- Implemented screen-local fixes only: `doc_type_code` submit wiring, shared `apiFetch` create wrapper with `Idempotency-Key`, honest slug-key preview, disabled DOCX import, read-only public visibility, and create-to-editor redirect.
- Runtime prerequisite discovered during smoke: `POST /api/v1/templates` reached `Service.CreateTemplate`, but the transaction lacked `metaldocs.tenant_id` / `metaldocs.actor_id` GUCs before `authz.Require`, returning HTTP 500. Fixed inside the templates create path with transaction-local authz context setup and a regression test.
- Post-repair gates passed: `scripts/check-system-runnable.ps1 -TargetRoute /api/v1/templates` and `scripts/check-module-contract-sync.ps1 -Module templates`.
- Runtime smoke passed: profile `qa_seed`, payload included `doc_type_code: "qa_seed"` and `Idempotency-Key`, backend returned HTTP 201, and the wizard redirected to `/templates/{id}/versions/1`.
- Evidence captured under `frontend/apps/web/design-source/novo-template-escopo/artifacts/screenshots/plan-12-4/`.

**Rename consistency checkpoint (2026-05-15):**
- Searched the active touched wizard path for `documents_v2` references after implementation.
- No `documents_v2` references were found in `frontend/apps/web/src/features/templates`, `wiki/backlog/novo-template-wizard.md`, or the concrete `frontend/apps/web/design-source/novo-template-*` artifacts touched for this screen.
- No out-of-scope rename sweep was performed.

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
**Blocked by:** No editor flow defined for templates wizard handoff. Today, `/templates/new` Step 5 just calls create and (presumably) redirects to list.
**File:** `frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx` (Step 5 submit handler — not yet implemented).
**TODO tag:** `TODO(novo-template-wizard:step3-editor-handoff)`
**Resolution:** After successful create, redirect to `/templates/<id>/edit?import=<blob-ref|blank>` based on `startingPoint`. Editor handles real upload (docx case) or stub-blank schema.json (blank case).

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
`TemplateWizardPage.handleSubmit` now calls `POST /api/v2/templates { key, name, description }` and redirects to `/templates/<id>/versions/<n>` on success. Error state surfaces inline in `StepConfirmation` via `submitError` prop.

---

### template-create-visibility-api
**Context:** `POST /api/v2/templates` generated handler (`routes_generated.go`) only accepts `key`, `name`, `description?`, `doc_type_code?`. It hardcodes `Visibility: VisibilityPublic` and `ApproverRole: "approver"`. The wizard collects permissions (Step 4: by area / by role / all-company) and structure origin (Step 3: blank / docx) — none of those are forwarded to the create API.
**Blocked by:** Backend API contract. The generated OpenAPI spec (`api.gen.go` `CreateTemplateJSONBody`) does not expose `visibility`, `areas`, `specific_areas`, or `approver_role` in the create body.
**Files:**
- Backend: `internal/modules/templates/api/api.gen.go` (`CreateTemplateJSONBody`)
- Backend: `internal/modules/templates/delivery/http/routes_generated.go` (`CreateTemplate`)
- Frontend: `frontend/apps/web/src/features/templates/api/templates.ts` (`createTemplate`)
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
