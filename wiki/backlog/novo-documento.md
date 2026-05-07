# Backlog: Novo-Documento Wizard

> **Last verified:** 2026-05-07
> **Scope:** Deferred items for the 4-step wizard at `/documents-v2/new` (`NewDocumentWizardPage`). Each item corresponds to a `TODO(novo-documento:*)` comment in code.
> **Out of scope:** Library screen deferrals (`backlog/library-screen.md`), editor deferrals (`backlog/editor.md`).
> **Key files:**
> - `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:112` — `handleCreate` — 2-call submit; slot-rollback TODO at :129
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

### sequence-preview {#sequence-preview}

**What:** Steps 2–4 show a code preview of the form `{profile}-{area}-???`. The `???` is a placeholder because no preview endpoint exists to reserve or estimate the next sequence number without committing it.

**Why deferred:** server resolves the sequence at `POST /api/v2/controlled-documents` create time. A preview would require either a stateless estimation endpoint (race-prone) or a two-phase hold that is out of scope.

**User impact:** preview is informational; the actual code (e.g. `PROC-02`) is visible in the editor after redirect.

**Backend prereq:** `GET /api/v2/controlled-documents/preview-code?profileCode=…&areaCode=…` or similar.

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

### slot-rollback {#slot-rollback}

**What:** The 2-call create sequence in `handleCreate` (`NewDocumentWizardPage.tsx:129`) does:
1. `POST /api/v2/controlled-documents` — creates and returns a CD slot with a consumed sequence number.
2. `POST /api/v2/documents` — clones template into a draft doc.

If step 2 fails (network error, template clone error, etc.) **after** step 1 succeeds, the CD slot remains in the registry and its sequence number is permanently consumed. No automatic compensation or rollback exists.

**User impact:** orphan CD slot visible in the registry. Sequence counter gap (e.g. `PROC-02` exists with no document). Manual cleanup by `system_admin` required.

**Backend prereq:** transactional create endpoint that performs both operations atomically, e.g. `POST /api/v2/documents/from-profile` that creates both CD and document in a single DB transaction.

---

### profile-counts {#profile-counts}

**What:** Step 1 (profile radio cards) does not show how many existing documents each profile has. The design mockup included a count badge per card.

**Why deferred:** `GET /api/v2/taxonomy/profiles` does not return document counts. Aggregating counts requires a JOIN against `documents` (or `controlled_documents`) which was deferred per the design-workflow audit (2026-05-07) — cost vs value unclear for the initial release.

**Backend prereq:** add `document_count` (or `controlled_document_count`) to the profiles list response.
