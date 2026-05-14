# Backlog: Novo-Documento Wizard

> **Last verified:** 2026-05-14 (Plan 12.3 integration audit + screen-local sync)
> **Scope:** Deferred items for the 4-step wizard at `/documents-v2/new` (`NewDocumentWizardPage`). Each item corresponds to a `TODO(novo-documento:*)` comment in code.
> **Out of scope:** Library screen deferrals (`backlog/library-screen.md`), editor deferrals (`backlog/editor.md`).
> **Key files:**
> - `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:146` — `handleCreate` — single-call atomic submit via `createControlledDocumentAtomic`
> - `frontend/apps/web/src/features/documents/lib/visibilityMeta.ts:1` — visibility SSOT; top-of-file TODO notes backend prereq
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepTemplate.tsx:1` — blank-template disabled state
> - `frontend/apps/web/src/features/documents/components/wizard/CodePreviewBanner.tsx:1` — live preview endpoint wiring + fallback state

---

## Smoke summary (2026-05-07)

4-step flow smoke-tested (feat/cd-atomic-create). Single atomic POST returned 201. Editor opened with server-resolved code `PROC-02`. All items below are **intentional deferrals**, not regressions.

---

## Integration Audit (2026-05-14)

| Item | Source | Runtime/API reality | Frontend reality | Classification | Action |
|---|---|---|---|---|---|
| Step 1 profile cards | design + current screen | `GET /api/v1/taxonomy/profiles` exists and is already wired through `useProfilesQuery` | Real cards render from live data | implemented and aligned | keep |
| Step 1 profile count badge | design + backlog `profile-counts` | profiles response has no document-count field | UI shows `—` placeholder only | missing backend capability | keep deferred in backlog |
| Step 2 area selector + title | design + current screen | `GET /api/v1/taxonomy/areas` exists and atomic create request accepts `processAreaCode` + `title` | Real selector/input gate progression | implemented and aligned | keep |
| Step 2 code preview | design + backlog `sequence-preview` | `GET /api/v1/controlled-documents/preview-code` exists in runtime/spec/frontend wrapper | Step 2 banner already uses the real preview endpoint | implemented and aligned | keep closed |
| Step 2 visibility radios | design + backlog `visibility` | no persisted `visibility` field or sharing endpoints exist on create | UI captures local state only, not submitted | missing backend capability | keep deferred in backlog |
| Step 2 people/external subcontrols | design + backlog `visibility` | no invitee/external-share endpoints exist | rendered disabled with `Em breve` state | missing backend capability | keep deferred in backlog |
| Step 3 template list | design + current screen | `GET /api/v1/templates?doc_type=...` exists and returns published version IDs | real list is wired and selectable | implemented and aligned | keep |
| Step 3 per-version picker | design + backlog `template-versions` | no versions-list route is available for the wizard to enumerate template versions | wizard exposes only published-version selection | defer | preserve backlog item |
| Step 3 blank template | design + backlog `blank-template` | atomic create still requires a valid `templateVersionId` | blank card stays disabled | missing backend capability | preserve backlog item |
| Step 4 summary card | design + current screen | preview endpoint and selected template are available by Step 4 | summary was partially wired and needed real code/template sync | screen-local integration fix | implement in this PR |
| Step 4 create action | design + current screen | `POST /api/v1/controlled-documents` exists, but runtime smoke returned `500 INTERNAL_ERROR` for a valid atomic-create request | wizard submits via `createControlledDocumentAtomic`; UI now surfaces the backend failure honestly | shared contract prerequisite | preserve blocker; no backend change in this PR |
| Template query/preview disabled-key wiring | frontend-only audit | no shared contract issue; purely local TanStack Query key usage | disabled queries used sentinel-style keys | screen-local integration fix | normalize in this PR |

Ready for implementation now:
- Step 4 summary sync with real preview/template data.
- Local query-key cleanup for wizard-owned disabled queries.

Prerequisites:
- Runtime create flow: `POST /api/v1/controlled-documents` returned `500 INTERNAL_ERROR` during smoke, so end-to-end creation remains blocked outside this screen-local scope.
- Visibility persistence/share semantics need backend capability.
- Blank-template creation needs backend support for `templateVersionId: null`.

Deferred:
- Profile counts stay deferred until a counts field exists on the profiles response.
- Per-version template picking stays deferred until a versions-list surface exists for the wizard.

Verification needed next:
- `cd frontend/apps/web`
- `pnpm.cmd tsc --noEmit -p tsconfig.build.json`
- `pnpm test`
- Runtime smoke for `/documents-v2/new`
- Screenshot capture for Step 1 through Step 4

---

## Items

### visibility

**What:** Step 2 exposes four visibility options (area / people / company / external) via `VISIBILITY_META` in `visibilityMeta.ts`. The selected value is stored in wizard state but **not submitted** to the backend.

**Why deferred:** `controlled_documents` has no `visibility` column. The `visibility TEXT DEFAULT 'area'` column was added by migration 0164 to the now-dropped `public.documents_v2`; it was never ported to `public.documents` or `controlled_documents`.

**Subcontrols that are no-op:**
- "Apenas minha área" — no ACL enforcement
- "Pessoas específicas" — no `/acl` or sharing endpoint; invitees field captured but not sent
- "Compartilhamento externo" — no link-generation, no password/watermark/expiry endpoint

**Backend prereq:** add `visibility` column to `controlled_documents` (or `documents`) + ACL/sharing endpoints.

---

### ~~sequence-preview~~ {#sequence-preview} — CLOSED (feat/cd-atomic-create, 2026-05-07)

`GET /api/v1/controlled-documents/preview-code?profileCode=…&areaCode=…` was shipped. Returns next code preview read-only (no reservation). Wizard now shows a live preview instead of `{profile}-{area}-???`. See `concepts/controlled-documents.md` for endpoint details.

---

### template-versions {#template-versions}

**What:** Step 3 lists only templates with a `published_version_id`. Rollback to a draft or previous published template version is not supported.

**Why deferred:** the template selector calls `GET /api/v1/templates?doc_type=…` and uses `published_version_id` to select the real published version. No version-picker UI exists.

**Backend prereq:** none blocking — UI work only to expose version picker.

---

### blank-template {#blank-template}

**What:** "Em branco" (blank template option) is rendered disabled in Step 3. Selecting it has no effect.

**Why deferred:** creating a document from a blank template requires a true empty-document clone path on the atomic create endpoint. The backend currently requires a valid `templateVersionId` in `POST /api/v1/controlled-documents`.

**Backend prereq:** support `templateVersionId: null` in `POST /api/v1/controlled-documents` (atomic create), or seed a blank sentinel template.

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
