# Backlog: Novo-Documento Wizard

> **Last verified:** 2026-06-21 (verify-and-archive sweep; see _cleanup-2026-06-21.md)
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
| Step 1 profile count badge | design + backlog `profile-counts` | profiles response has no document-count field | UI shows `—` placeholder only | missing backend capability | keep deferred in backlog |
| Step 2 visibility — `people` / external subcontrols | design + backlog `visibility` | no invitee/external-share endpoints exist | rendered with `Em breve` for unsupported subcontrols | missing backend capability | keep deferred |
| Step 3 per-version picker | design + backlog `template-versions` | no versions-list route for the wizard | wizard exposes only published-version selection | defer | preserve backlog item |

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

### Closed (see git history)

sequence-preview (feat/cd-atomic-create, 2026-05-07), blank-template (2026-05-28), slot-rollback (feat/cd-atomic-create, 2026-05-07) — all closed; details recoverable from git history.

---

### template-versions {#template-versions}

**What:** Step 3 lists only templates with a `published_version_id`. Rollback to a draft or previous published template version is not supported.

**Why deferred:** the template selector calls `GET /api/v1/templates?doc_type=…` and uses `published_version_id` to select the real published version. No version-picker UI exists.

**Backend prereq:** none blocking — UI work only to expose version picker.

---

### profile-counts {#profile-counts}

**What:** Step 1 (profile radio cards) does not show how many existing documents each profile has. The design mockup included a count badge per card.

**Why deferred:** `GET /api/v1/taxonomy/profiles` does not return document counts. Aggregating counts requires a JOIN against `documents` (or `controlled_documents`) which was deferred per the design-workflow audit (2026-05-07) — cost vs value unclear for the initial release.

**Backend prereq:** add `document_count` (or `controlled_document_count`) to the profiles list response.

---

## Follow-ups from major-findings PR (2026-05-07)

- ~~**Typography token cluster**~~ — **closed 2026-07-02 (FE-06/FE-07).** Deviated from the literal `--fs-1..--fs-7` ask: densified the existing, already-60-file-adopted `--font-size-*`/`--sp-*` scales instead of introducing a parallel `--fs-*` namespace (Global Maximum call — a second naming convention would fragment the token system). Added `--font-size-2xs-2`, `--font-size-xs-2`, `--font-size-sm-2`, `--font-size-md-2`, `--font-size-md-3` plus `--sp-0-5`/`--sp-1-5` to `tokens.css`, and replaced raw `font-size`/padding px values in `StepProfile.module.css` and `StepConfirm.module.css`. **Not fully closed**: `StepTemplate.module.css` (`font-size: 14px` line 38, `11px` line 45) and `StepAreaCodeVisibility.module.css` (`font-size: 13px` at lines 45/94/128) still have raw literals — they map cleanly onto the now-existing `--font-size-md`/`--font-size-xs`/`--font-size-sm-2` tokens respectively but were not swapped in this pass (priority-order cutoff; these 2 files are mechanical follow-up, same pattern as StepProfile/StepConfirm).
- **Kicker variants** — **partially closed 2026-07-02.** Added `.kicker--callout` modifier to `src/styles.css` and applied it in `StepConfirm.tsx`/`StepConfirm.module.css`, replacing that file's single-purpose `:global(.kicker) { line-height: 1; }` override. `.kicker--block` not added — no call site needed it. The two remaining `:global(.kicker)` reaches in `WizardShell.module.css` (`.header :global(.kicker)`, `.container > :global(.card) > :global(.kicker)`) were left as-is: they are structural/positional descendant selectors (DOM-ancestry-based layout composition), not single-purpose style overrides, so they don't fit the modifier-class pattern the same way.
- ~~**InlineAlert primitive**~~ — **closed 2026-07-02 (FE-06/FE-07).** Added `components/ui/InlineAlert.tsx` (`tone: 'error'|'warning'|'info'`, role/aria-live derived per tone, optional `action` slot) + barrel export + 6 unit tests in `components/ui/primitives.test.tsx`. Migrated 3 wizard sites: `StepConfirm.tsx`, `StepProfile.tsx`, `StepTemplate.tsx`. Not migrated: `TemplateEditorPage.tsx`'s alert slot (owned by a concurrent agent this pass, and its alert shape didn't cleanly match `InlineAlert`'s `message`/`action` contract — flagged as follow-up, not fixed) and Document/Approval-flow future sites (not yet built).
- **ESLint rule** — still open, **report-only per FE-07 scope** (`no-restricted-syntax` banning inline tuple `queryKey` in `features/**/queries/**`, forcing use of the `QK` registry in `frontend/apps/web/src/lib/queryKeys.ts`). Not implemented this pass — no ESLint config changes were in the FE-06/FE-07 boundary.
- **SelectableCardGroup** — still open, unchanged this pass (defer until 2nd consumer appears — no new consumer surfaced during FE-06/FE-07).
- **Stepper numeric ids** — still open, unchanged this pass (out of FE-06/FE-07 scope — touches a shared primitive's public type).
