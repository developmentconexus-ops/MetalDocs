# Backlog: Novo-Documento Wizard

> **Last verified:** 2026-05-07 (sequence-preview + slot-rollback closed by feat/cd-atomic-create)
> **Scope:** Deferred items for the 4-step wizard at `/documents-v2/new` (`NewDocumentWizardPage`). Each item corresponds to a `TODO(novo-documento:*)` comment in code.
> **Out of scope:** Library screen deferrals (`backlog/library-screen.md`), editor deferrals (`backlog/editor.md`).
> **Key files:**
> - `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:151` — `handleCreate` — 2-call submit; slot-rollback TODO at :113
> - `frontend/apps/web/src/features/documents/lib/visibilityMeta.ts:1` — visibility SSOT; top-of-file TODO notes backend prereq
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepTemplate.tsx:1` — blank-template disabled state
> - `frontend/apps/web/src/features/documents/components/wizard/CodePreviewBanner.tsx:1` — `???` placeholder for unresolved sequence

---

## Smoke summary (2026-05-07)

4-step flow smoke-tested. Both POSTs returned 201. Editor opened with server-resolved code `PROC-02`. All items below are **intentional deferrals**, not regressions.

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

`GET /api/v2/controlled-documents/preview-code?profileCode=…&areaCode=…` was shipped. Returns next code preview read-only (no reservation). Wizard can now show a live preview instead of `{profile}-{area}-???`. See `concepts/controlled-documents.md` for endpoint details.

---

### template-versions {#template-versions}

**What:** Step 3 lists only templates with a `published_version_id`. Rollback to a draft or previous published template version is not supported.

**Why deferred:** the template selector calls `GET /api/v2/templates?profileCode=…` and filters client-side to `published_version_id != null`. No version-picker UI exists.

**Backend prereq:** none blocking — UI work only to expose version picker.

---

### blank-template {#blank-template}

**What:** "Em branco" (blank template option) is rendered disabled in Step 3. Selecting it has no effect.

**Why deferred:** creating a document from a blank template requires a true empty-document clone path on the backend (`POST /api/v2/documents` with no `template_version_id`, or a dedicated blank-template sentinel). The backend currently requires a valid `template_version_id`.

**Backend prereq:** support `template_version_id: null` in `POST /api/v2/documents`, or seed a blank sentinel template.

---

### ~~slot-rollback~~ {#slot-rollback} — CLOSED (feat/cd-atomic-create, 2026-05-07)

`POST /api/v2/controlled-documents` now creates the CD slot + first document revision in a single DB transaction. The two-call sequence no longer exists; orphan slots are structurally impossible. The legacy `POST /api/v2/documents` (create from CD) endpoint was deleted. See ADR 0009.

---

### profile-counts {#profile-counts}

**What:** Step 1 (profile radio cards) does not show how many existing documents each profile has. The design mockup included a count badge per card.

**Why deferred:** `GET /api/v2/taxonomy/profiles` does not return document counts. Aggregating counts requires a JOIN against `documents` (or `controlled_documents`) which was deferred per the design-workflow audit (2026-05-07) — cost vs value unclear for the initial release.

**Backend prereq:** add `document_count` (or `controlled_document_count`) to the profiles list response.

---

## Follow-ups from major-findings PR (2026-05-07)

- **Typography token cluster** — declare `--fs-1..--fs-7` in `tokens.css`; replace raw `font-size: NNpx` in StepProfile / StepAreaCodeVisibility / StepTemplate / StepConfirm CSS Modules. Estimated 45m.
- **Kicker variants** — add `.kicker--block` + `.kicker--callout` modifiers in `src/styles.css`; drop `:global(.kicker)` reaches from StepAreaCodeVisibility + StepConfirm modules. Estimated 30m.
- **InlineAlert primitive** — extract `<InlineAlert role="alert" tone="error|warning|info" message />` in `components/ui/`; replace 4 wizard sites + future Document/Approval flows. Estimated 30m.
- **ESLint rule** — `no-restricted-syntax` banning inline tuple `queryKey` in `features/**/queries/`. Forces QK registry use. Estimated 30m.
- **SelectableCardGroup** — wrapper exposing `name value onChange` so future radio-group consumers do not stitch manually. Defer until 2nd consumer appears.
- **Stepper numeric ids** — `Stepper` primitive currently takes `id: string` for steps; `WizardShell.tsx:35` round-trips `String(currentStep)`. Either accept `id: string | number` in Stepper or wrap the wizard with a typed `<NumericStepper currentStep: number />`. Touches a shared primitive — defer to dedicated PR. Estimated 30m.
