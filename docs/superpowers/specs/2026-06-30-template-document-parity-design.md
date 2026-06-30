# Design — Template ↔ Document parity & lifecycle correctness

**Date:** 2026-06-30
**Status:** Approved (operator goal directive + AskUserQuestion decisions 2026-06-30)
**System-impact gate:** `docs/superpowers/analysis/2026-06-30-template-document-parity-system-impact.md` (🟡 Yellow)
**Branch:** `feat/template-document-parity`

## 1. Problem & success criteria

Five behaviours observed in QA must be corrected, and the duplicated template/document UX consolidated, at Grade-A (no patch/workaround/hardcode):

1. Templates and documents must **share one frontend view experience** — detail screen, top bar, side bars, approval screen.
2. Templates must **stop auto-creating the next draft version** on approve/release. New-version creation is a manual act only.
3. Clicking a template must open a **detail/view screen** (like documents), not jump straight to the eigenpal editor.
4. The document creation wizard must **show the real document number**, not `???`.
5. Documents must **not** auto-create the next version on approve (verify; treat the template/document asymmetry as the bug).

**Done =** all five corrected; backend + contract + DB + FE gates green; final Preview-driven real-user QA on both flows (with the `approver` QA account) passes with zero errors.

## 2. Locked decisions

- **D1 — Full shared shell.** Extract one shared *controlled-artifact* view layer; refactor **both** templates and documents to render through it. Maximum dedup; regression-covered.
- **D2 — Align status enum now.** Rename templates' `in_review` → `under_review` to match documents, via expand/contract (DB + Go + OpenAPI + FE). Other lifecycle states stay as-is (legitimate domain difference).
- **D3 — Document-pattern approval.** Templates move to a dedicated approval/sign-off screen (review canvas + decision sidebar), replacing the inline `VersionActionPanel`.

## 3. Non-goals (explicit boundaries)

- **No backend module merge.** `templates` (authoring) and `documents` (instances) remain distinct bounded contexts. Merging would violate the module-boundary invariant. Dedup is frontend + lifecycle-semantic only.
- **No full state-machine unification.** Only the `in_review`→`under_review` string aligns (D2). Document-only states (`scheduled`, `rejected`, `superseded`, `archived`) and template-only `obsolete` stay.
- **No numbering rework.** `controlleddocuments` numbering + the `preview-code` endpoint are correct and reused unchanged.
- **No documents lifecycle change.** Documents are already manual-only (#5 is verify-only).
- No unrelated refactoring outside the controlled-artifact surface.

## 4. Architecture

### 4.1 Shared controlled-artifact frontend layer (D1, D3)

New directory `frontend/apps/web/src/features/shared/controlled-artifact/`. **Presentational + kind-agnostic** components fed by a normalized view-model; **per-kind adapters** own data-fetching and the available lifecycle actions. The shell must not branch on `kind` beyond declarative config (tabs, labels).

**Normalized view-model (the interface boundary):**

```ts
type ArtifactKind = 'document' | 'template';

interface ArtifactViewModel {
  kind: ArtifactKind;
  id: string;
  code: string | null;            // document: CD code (DC-RH-001); template: code/name
  title: string;
  status: LifecycleStatus;        // unified vocabulary incl. under_review
  versionNumber: number;
  revisionLabel: string | null;   // REVnn (shared formatRevisionCode)
  hero: ArtifactHeroModel;        // breadcrumb, badges, subtitle
  meta: ArtifactMetaModel;        // profile/area/visibility labels, file size, page count, dates
  approvalChain: ApprovalChainItem[] | null;
  lineage: VersionHistoryItem[];
  tabs: ArtifactTab[];            // document: Documento + Distribuição; template: Documento
  actions: ArtifactActionSet;     // available actions + handlers (kind-specific)
}
```

**Shared components (generalized from the existing document components):**

| New shared component | Generalizes (document source) |
|---|---|
| `ArtifactDetailLayout` (tabbed shell) | `documents/pages/DocumentDetailLayout.tsx` |
| `ArtifactDetailView` (hero, KPI strip, About, approval chain, lineage) | `documents/pages/DocumentPublishedPage.tsx` |
| `ArtifactHero` | `documents/components/DocumentHero.tsx` |
| `ArtifactMetaSidebar` (metadata + history + approval chain) | `documents/components/EditorMetaSidebar.tsx` |
| `ArtifactApprovalScreen` (read-only canvas + decision sidebar) | `approval/pages/SignoffDetailPage.tsx` + `approval/components/ControlledDocumentDetailPanel.tsx` |

**Adapters (own kind-specific queries + action handlers):**
- `useDocumentArtifact(id) → ArtifactViewModel` — wraps existing document detail/approval queries + signoff endpoints.
- `useTemplateArtifact(templateId) → ArtifactViewModel` — wraps template `GetTemplate`/versions/audit + version review/approve/reject/publish endpoints.

The approval backends differ (documents = approval routes/stages/instances/signoffs; templates = reviewer→approver on the version row). The adapter normalizes both into `ArtifactActionSet` (available decisions + submit handlers); the shared `ArtifactApprovalScreen` is dumb.

### 4.2 Routing changes

- **Add** `/templates/:templateId` → `ArtifactDetailLayout` + `ArtifactDetailView` (template adapter). Template list-click navigates **here** (was `/templates/:id/edit`).
- **Add** a dedicated template approval route (e.g. `/templates/:templateId/versions/:n/approval`) → `ArtifactApprovalScreen` (template adapter). **Remove** the inline `VersionActionPanel` from `TemplateEditorPage`.
- **Re-point** documents' `/documents/:id` and `/approvals/:id` to render through the shared components (behaviour preserved; regression-covered).
- Template editor gains `ArtifactMetaSidebar` (right) next to its existing template-only Variáveis panel (left).
- Template detail screen surfaces a **"Criar nova versão"** action → manual `POST /api/v1/templates/{id}/versions` (the only revision path after WS-A).

### 4.3 Backend lifecycle correctness

**WS-A — remove auto-version (`templates`):**
- `templates/application/lifecycle.go`: delete the `spawnNextDraft` calls in `Approve` (Accept path, ~:256-298) and `PublishTemplateVersion` (~:411-454); remove `NextDraft` from `ApproveResult` and `PublishTemplateVersionResult`. `nextVersionNumber` + `spawnNextDraft` remain (still used by manual `CreateNextVersion`).
- OpenAPI `api/openapi/v1/openapi.yaml`: drop `next_draft_id`/`next_draft_version_num` from `PublishTemplateVersion` 200 and `next_draft` from `ApproveTemplateVersionResponse`. Regenerate templates `api.gen.go` + FE `lib/api-types`.
- Tests `templates/application/lifecycle_test.go:310-440`: flip to assert no next draft + no extra version row; add integration e2e (publish ⇒ 1 version; manual create ⇒ draft v2).

**WS-B — align status enum (`templates`):**
- DB migration (forward-only, expand/contract): add `under_review` to the template-version status CHECK, backfill `in_review`→`under_review`, drop `in_review`. Mirror the documents status check pattern.
- Go: `domain/version.go` `VersionStatusInReview` value → `"under_review"`; update `CanTransition` + audit/labels.
- OpenAPI status enum + regen; FE literals (`features/templates/lib/canActOnVersion.ts`, gates, badges).

### 4.4 Wizard number fix (WS-D, FE only)

`documents/components/wizard/steps/StepConfirm.tsx:60` mirrors the canonical `CodePreviewBanner.tsx` states:
```ts
const codePreview = !ready
  ? `${PLACEHOLDER}-${PLACEHOLDER}-${PLACEHOLDER}`
  : isLoading ? `${profile.code}-${area.code}-…`
  : previewCode ?? `${profile.code}-${area.code}-${PLACEHOLDER}`;
```
Backend `GET /controlled-documents/preview-code` is correct and unchanged.

## 5. Contract & data changes

- OpenAPI: template approve/publish response shrink (WS-A) + template status enum value (WS-B). Single regen batch. `oapi-codegen` must show **no drift** after.
- DB: one forward-only migration for the template-version status CHECK (WS-B). No other migration. No new tables, no tenant-scoping change.
- Expand/contract ordering for WS-A: (1) FE stops consuming `next_draft*` + stops auto-nav; (2) backend removes fields + regen.

## 6. Test & QA plan

- **Backend:** `go build ./...`, `go test ./...`, `go test -tags=integration ./...`. Flip/extend lifecycle guard tests; integration e2e for manual-only versioning; migration applies clean + baseline (`check-system-runnable.ps1`).
- **Contract:** regen no-drift check.
- **Frontend:** `make test` / web vitest — shared shell components (both adapters), template detail route, shared approval screen, `StepConfirm` three states; `npm run typecheck` clean.
- **Final acceptance (Preview, real user):** log in as `approver` (SoD). Template flow: list → **detail screen** → editor → submit → review → approve → publish ⇒ **assert no auto v2** → manual "Criar nova versão" ⇒ v2 draft. Document flow: wizard shows **real number** → create → detail → submit → signoff → publish. Both render the **same shell/bars/sidebars/approval screen**. Zero errors = PASS; any error = honest FAIL.

## 7. Docs & ADRs

- **ADR-A:** revert template auto-next-draft (supersedes the 2026-05-31 `feat/templates-approve-next-draft-response` decision; rationale: new revision is a manual act per QMS norm + parity with documents).
- **ADR-B:** shared controlled-artifact frontend view layer (records the dedup so future screens don't re-fork).
- Update `wiki/modules/templates.md` (manual-only revision; `under_review`), `wiki/modules/documents.md` (shared shell cross-link), FE structure doc; refresh `Last verified`.

## 8. Milestone decomposition

- **M1 — Lifecycle correctness (backend + contract + thin FE + wizard).** WS-A (remove auto-version, expand/contract, tests), WS-B (enum align + migration + regen), WS-D (wizard fix). Lands the unified status vocabulary the shell depends on. Gate: backend + integration + contract-no-drift + typecheck green.
- **M2 — Shared controlled-artifact view layer.** WS-C: extract shared components + adapters, re-point documents (regression-covered), add template detail + approval routes, remove inline panel, add template editor sidebar, manual "Criar nova versão" action. Gate: web vitest + typecheck + visual parity review.
- **M3 — Docs/ADR + final acceptance.** WS-E ADRs + wiki; full Preview real-user QA on both flows; #5 confirm-only. Gate: mission/milestone validator + Preview QA zero-errors.

Each milestone runs through the `milestone` workflow with a separation-of-powers validation gate; fresh subagent sessions per feature; two-stage (spec-compliance then code-quality) review; sonnet implement/review, haiku mechanical, never fable, ≤15 concurrent.

## 9. Risks & mitigations

- **Re-pointing working document screens (M2)** → regression. Mitigate: snapshot/vitest + Preview visual parity before/after; adapter keeps document data flow identical.
- **Status enum migration (WS-B)** → stale literals. Mitigate: grep sweep for `in_review`; expand/contract; integration coverage.
- **Approval backend divergence** → leaky abstraction in `ArtifactApprovalScreen`. Mitigate: normalize via `ArtifactActionSet`; screen stays presentational; adapter owns endpoints.
- **Dev-seed `approver` authz** (memory `dev-seed-template-approval-authz-gap`) → blocks Preview approval QA. Verify the F-IAM1 seed fix is present before M3 QA; if not, repair the seed (never force-grant).
