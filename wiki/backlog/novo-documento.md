# Backlog: Novo-Documento Wizard

> **Last verified:** 2026-05-28 (runtime QA pass post-PC-reset; happy path live-validated for `company` + `restricted` area scope, blank template real)
> **Scope:** Deferred items for the 4-step wizard at `/documents/new` (`NewDocumentWizardPage`). Each item corresponds to a `TODO(novo-documento:*)` comment in code.
> **Out of scope:** Library screen deferrals (`backlog/library-screen.md`), editor deferrals (`backlog/editor.md`).
> **Key files:**
> - `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx` — `handleCreate` + `buildVisibilityPayload` — single-call atomic submit via `createControlledDocumentAtomic`, visibility persisted
> - `frontend/apps/web/src/features/documents/lib/visibilityMeta.ts:1` — visibility SSOT
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepTemplate.tsx:1` — blank-template real path via `blankTemplate.templateVersionId` (`/api/v1/templates/system/blank`)
> - `frontend/apps/web/src/features/documents/components/wizard/CodePreviewBanner.tsx:1` — live preview endpoint wiring + fallback state

---

## Smoke summary (2026-05-07)

4-step flow smoke-tested (feat/cd-atomic-create). Single atomic POST returned 201. Editor opened with server-resolved code `PROC-02`. All items below are **intentional deferrals**, not regressions.

---

## Integration Audit (2026-05-28 runtime QA)

Runtime evidence from 2026-05-28 live browser pass (post-PC-reset, API `:8081`, web `:4173`):

- Created `PO-RH-003` (`a6bbd237-7210-42f5-9359-3e3523ba065b`, controlled-doc `4a3e9eb9-5c3c-445e-9814-a5f8c36634a6`) with `visibility=company`. Editor opened at `/documents/<id>/edit`.
- Created `PO-RH-004` (`ec71a499-58d2-4925-a758-fab175c4b514`) with restricted-area scope `Recursos Humanos`. Editor metadata showed `Restrito a area Recursos Humanos`.
- `GET /api/v1/controlled-documents/preview-code?profileCode=po&areaCode=rh` → `PO-RH-003`.
- `GET /api/v1/templates/system/blank` → `templateId=00000000-0000-0000-0000-000000000101`, `templateVersionId=00000000-0000-0000-0000-000000000102` (real blank-template path).

| Item | Source | Runtime/API reality | Frontend reality | Classification | Action |
|---|---|---|---|---|---|
| Step 1 profile cards | design + current screen | `GET /api/v1/taxonomy/profiles` exists, wired via `useProfilesQuery` | Real cards render from live data | implemented and aligned | keep |
| Step 1 profile count badge | design + backlog `profile-counts` | profiles response has no document-count field | UI shows `—` placeholder only | missing backend capability | keep deferred in backlog |
| Step 2 area selector + title | design + current screen | `GET /api/v1/taxonomy/areas` + atomic create accepts `processAreaCode` + `title` | Real selector/input gate progression | implemented and aligned | keep |
| Step 2 code preview | design + backlog `sequence-preview` | `GET /api/v1/controlled-documents/preview-code` returns real next code | Banner shows live preview | implemented and aligned | keep closed |
| Step 2 visibility — `company` | design + backlog `visibility` | `visibility` accepted by atomic create, persisted on `controlled_documents` | submitted + persisted | implemented and aligned | move to closed |
| Step 2 visibility — `area` / restricted-area scope | design + backlog `visibility` | atomic create accepts `visibilityAreaCodes`; restricted-area visibility persisted (editor metadata reflects it) | submitted + persisted | implemented and aligned | move to closed |
| Step 2 visibility — `people` / external subcontrols | design + backlog `visibility` | no invitee/external-share endpoints exist | rendered with `Em breve` for unsupported subcontrols | missing backend capability | keep deferred |
| Step 3 template list | design + current screen | `GET /api/v1/templates?doc_type=...` returns published version IDs | real list wired and selectable | implemented and aligned | keep |
| Step 3 per-version picker | design + backlog `template-versions` | no versions-list route for the wizard | wizard exposes only published-version selection | defer | preserve backlog item |
| Step 3 blank template | design + backlog `blank-template` | `GET /api/v1/templates/system/blank` returns real sentinel; atomic create accepts that `templateVersionId` | blank card selectable, real submit | implemented and aligned | move to closed |
| Step 4 summary card | design + current screen | preview endpoint + selected template available by Step 4 | summary mirrors real preview code + template | implemented and aligned | keep |
| Step 4 create action | design + current screen | `POST /api/v1/controlled-documents` returns `201` for valid atomic-create request via the real browser flow; landing redirects to `/documents/:id/edit` | wizard submits via `createControlledDocumentAtomic`, lands editor | implemented and aligned | keep |
| Template query/preview disabled-key wiring | frontend-only audit | n/a | canonical non-sentinel keys | implemented and aligned | keep |

Open caveat:
- A raw PowerShell `POST /api/v1/controlled-documents` was observed returning `403` during direct API probing while the real browser flow succeeded. Treat as a possible auth/session/tooling nuance, not a product blocker. Investigate before reopening the screen blocker.

Ready for implementation now:
- None outstanding for the screen happy path.

Prerequisites:
- People-invite + external-share subcontrols (`#visibility`, `#sharing`) still need backend capability.
- Per-version template picker still needs `template-versions` list surface.

Deferred:
- Profile counts (`#profile-counts`).
- Per-version template picking (`#template-versions`).

Verification needed next:
- `cd frontend/apps/web`
- `pnpm.cmd tsc --noEmit -p tsconfig.build.json`
- `pnpm.cmd test -- src/features/documents/pages/NewDocumentWizardPage.test.tsx`
- Runtime smoke for `/documents/new` (re-run after PC reset / fresh dev server)
- Screenshot capture for Step 1 through Step 4 if running a full evidence pass

---

## Items

### visibility {#visibility}

**Status (2026-05-28):** PARTIALLY CLOSED — `company` and restricted-area (`area`) scopes are submitted via the atomic-create payload and persisted on `controlled_documents`. Runtime QA verified both: `PO-RH-003` (company) and `PO-RH-004` (restricted to area `rh`) survived create + reload; editor metadata shows `Restrito a area Recursos Humanos`.

**Still deferred:**
- "Pessoas específicas" — no `/acl` or sharing endpoint; invitees field captured but not sent
- "Compartilhamento externo" — no link-generation, no password/watermark/expiry endpoint

**Backend prereq for remaining subcontrols:** invitee ACL endpoints + external-share / password / watermark / expiry endpoints (see `#sharing`).

---

### ~~sequence-preview~~ {#sequence-preview} — CLOSED (feat/cd-atomic-create, 2026-05-07)

`GET /api/v1/controlled-documents/preview-code?profileCode=…&areaCode=…` was shipped. Returns next code preview read-only (no reservation). Wizard now shows a live preview instead of `{profile}-{area}-???`. See `concepts/controlled-documents.md` for endpoint details.

---

### template-versions {#template-versions}

**What:** Step 3 lists only templates with a `published_version_id`. Rollback to a draft or previous published template version is not supported.

**Why deferred:** the template selector calls `GET /api/v1/templates?doc_type=…` and uses `published_version_id` to select the real published version. No version-picker UI exists.

**Backend prereq:** none blocking — UI work only to expose version picker.

---

### ~~blank-template~~ {#blank-template} — CLOSED (2026-05-28)

`GET /api/v1/templates/system/blank` ships a real sentinel: `templateId=00000000-0000-0000-0000-000000000101`, `templateVersionId=00000000-0000-0000-0000-000000000102`. Step 3 "Em branco" card is selectable and submits that sentinel `templateVersionId` through the standard atomic-create payload. Runtime QA verified blank-template creation ends on the real editor route.

---

### ~~slot-rollback~~ {#slot-rollback} — CLOSED (feat/cd-atomic-create, 2026-05-07)

`POST /api/v1/controlled-documents` now creates the CD slot + first document revision in a single DB transaction. The two-call sequence no longer exists; orphan slots are structurally impossible. The legacy create-from-CD path was deleted. See ADR 0011.

---

### profile-counts {#profile-counts}

**What:** Step 1 (profile radio cards) does not show how many existing documents each profile has. The design mockup included a count badge per card.

**Why deferred:** `GET /api/v1/taxonomy/profiles` does not return document counts. Aggregating counts requires a JOIN against `documents` (or `controlled_documents`) which was deferred per the design-workflow audit (2026-05-07) — cost vs value unclear for the initial release.

**Backend prereq:** add `document_count` (or `controlled_document_count`) to the profiles list response.

---

## Follow-ups from major-findings PR (2026-05-07)

- **Typography token cluster** — declare `--fs-1..--fs-7` in `tokens.css`; replace raw `font-size: NNpx` in StepProfile / StepAreaCodeVisibility / StepTemplate / StepConfirm CSS Modules. Estimated 45m.
- **Kicker variants** — add `.kicker--block` + `.kicker--callout` modifiers in `src/styles.css`; drop `:global(.kicker)` reaches from StepAreaCodeVisibility + StepConfirm modules. Estimated 30m.
- **InlineAlert primitive** — extract `<InlineAlert role="alert" tone="error|warning|info" message />` in `components/ui/`; replace 4 wizard sites + future Document/Approval flows. Estimated 30m.
- **ESLint rule** — `no-restricted-syntax` banning inline tuple `queryKey` in `features/**/queries/`. Forces QK registry use. Estimated 30m.
- **SelectableCardGroup** — wrapper exposing `name value onChange` so future radio-group consumers do not stitch manually. Defer until 2nd consumer appears.
- **Stepper numeric ids** — `Stepper` primitive currently takes `id: string` for steps; `WizardShell.tsx:35` round-trips `String(currentStep)`. Either accept `id: string | number` in Stepper or wrap the wizard with a typed `<NumericStepper currentStep: number />`. Touches a shared primitive — defer to dedicated PR. Estimated 30m.
