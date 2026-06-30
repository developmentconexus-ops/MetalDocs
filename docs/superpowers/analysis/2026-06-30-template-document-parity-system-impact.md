# System-impact analysis — Template ↔ Document parity & lifecycle correctness

**Date:** 2026-06-30
**Intent (one line):** Make templates and documents share one controlled-artifact UX (view shell, sidebars, approval screen), stop templates from auto-spawning the next draft version on approve/publish (manual only), give templates a detail/view screen instead of jumping straight to the eigenpal editor, fix the creation-wizard document-number showing "???", and confirm documents do not auto-version — all Grade-A, dedup duplicated logic.
**Work type:** feature (multi-part; spans `templates` + `documents` + `controlleddocuments` backend and the templates/documents/approval/shared frontend)
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow *(see §10)*

> Ground truth gathered 2026-06-30 by 4 read-only investigators + wiki `templates.md`/`documents.md`. Anchors below are verified.

---

## 1. Classify & own

- **Work type:** feature bundle (5 operator findings + the dedup they imply). No new module.
- **Owning module(s):**
  - **#2 template auto-version** → `templates` (backend service + contract) + templates FE.
  - **#4 wizard "???"** → frontend `documents` wizard only. Backend `controlleddocuments` already exposes the preview (`PreviewCode`/`PeekSeq`); **no backend change**.
  - **#1 screen parity + #3 template detail screen** → frontend `templates`, `documents`, `approval`, `shared` (new shared controlled-artifact view layer).
  - **#5 document auto-version** → **no change** — `documents` is already manual-only (correct).
- **Explicitly NOT owning:**
  - **Do NOT merge `templates` and `documents` backend modules.** They are correct, distinct bounded contexts (authoring vs instances). Merging would violate invariant 6 (module boundaries). The duplication to remove is on the **frontend**, plus **semantic alignment** of the template lifecycle to the document lifecycle.
  - `controlleddocuments` numbering logic — already correct and reused; not re-implemented.
- **Cross-module edges (with direction):**
  - `documents → controlleddocuments` (preview-code, atomic create) — already via published port (`SequenceAllocator`, `CreateDocumentTx`). No new edge.
  - `documents → templates` (template version reads at create) — unchanged.
  - FE shared shell is frontend-only; introduces no backend cross-module edge.
- **Ambiguity?** None. No AS-3.

## 2. Foundation verdict (Global-Maximum)

- **Base today:**
  - Templates auto-spawn the next `v(n+1)` draft inside the publish/approve tx (`lifecycle.go` `spawnNextDraft`, returned as `next_draft`; added 2026-05-31 `feat/templates-approve-next-draft-response`). Documents do **not** — they only transition the single row (`publish_service.go:51-171`, no insert).
  - Frontend grew two **parallel bespoke** controlled-artifact experiences: documents have a detail route (`/documents/:id` → `DocumentDetailLayout`+`DocumentPublishedPage`), a right metadata sidebar (`EditorMetaSidebar`), and a dedicated approval screen (`/approvals/:id` → `SignoffDetailPage`+`ControlledDocumentDetailPanel`); templates have **none** of these — list-click jumps to `/templates/:id/edit`, approval is an inline `VersionActionPanel` at the bottom of the editor.
- **Sound, or legacy/patch?**
  - The template auto-spawn is a **local-maximum patch**: it diverged templates from the document model and from every mature QMS (new revision is always a deliberate manual act; releasing never auto-creates a draft). Removing it is reverting a wrong decision, not optimizing inside it. ✅ global-maximum move.
  - The FE duplication is a **local maximum**. Building *more* bespoke template screens (copy-paste of the document screens) to reach parity would lock in the duplication — a defect. Global-maximum = extract a **shared controlled-artifact view layer** (detail shell, approval screen, metadata sidebar) parameterized by `kind: template | document`, and render both features through it.
- **AS-2?** The FE-duplication is an AS-2-shaped finding (would otherwise optimize inside a patch) — **but the operator has explicitly pre-authorized the dedup/simplification**, so it is resolved as a **locked design constraint** handed to brainstorming, not a STOP.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | Yes (lightly) | #2 removes spawn only; publish/approve caps unchanged; manual `CreateNextVersion` already uses `authz.Require(CapTemplateEdit,"tenant")` (create.go). No new cap. | `authz.Require` |
| Contract-first (OpenAPI + oapi-codegen) | **Yes** | #2 changes `PublishTemplateVersion` 200 + `ApproveTemplateVersionResponse` (drop `next_draft*`). Edit `api/openapi/v1/openapi.yaml` → regen templates `api.gen.go` + FE api-types. #4 uses existing `PreviewCodeResponse` (no change). | `oapi-codegen`, `go generate` |
| Multi-tenant pooled | No | No new tables; no tenant-scoping change. | — |
| Async = transactional outbox | No | Publish lifecycle event fanout unchanged. | — |
| DB enforces invariants | No | No migration required by #2 (one-published-per-template partial-unique still holds; no constraint expects a spawned draft). | — |
| Cross-module via published interface only | No | No backend cross-module change; FE shell is FE-only. | — |

No AS-1 violation.

## 4. Capability wiring

**N/A** — no capability added or changed. (Manual revision creation reuses the existing `template.edit` path; publish/approve caps unchanged.)

## 5. Module wiring

**N/A** — no new module. Backend modules stay as-is (explicitly: no templates↔documents merge).

## 6. Frameworks to reuse, not reinvent

- Backend #2: reuse `TxRunner` (`runner.Do`), `authz.Require`, `audit.RecordTx`, existing repo `CreateVersionTx` (for the manual path only). The change is **deletion** of `spawnNextDraft` calls in `Approve`/`PublishTemplateVersion` + their result fields + contract fields.
- FE: reuse existing **shared** primitives (`EditorChrome`, `WizardShell`, `StatusPill`, `CodeChip`, `VersionBadge`, `AutosaveStatus`). **Extract new shared** components into `features/shared/` for the controlled-artifact detail shell, approval screen, and metadata sidebar; both `templates` and `documents` consume them. `#4` reuses the existing `usePreviewCodeQuery` + the canonical `CodePreviewBanner` loading/data state pattern.

## 7. Contract & data

- **OpenAPI-first:** edit `api/openapi/v1/openapi.yaml` — remove `next_draft_id`/`next_draft_version_num` from `PublishTemplateVersion` 200 and `next_draft` from `ApproveTemplateVersionResponse`; regen templates module + FE types. **Expand/contract order:** (1) FE stops reading `next_draft*` and stops auto-navigating; (2) backend stops populating + remove fields + regen. No other route changes (#4 endpoint already exists).
- **Migration:** none required. (Dev DB may hold already-spawned orphan v2 drafts from the old behavior; that is dev-seed noise, not a prod migration — do not write a destructive prod migration for it.)
- **Destructive change?** Response-field removal — handled by the FE-first expand/contract above; all consumers are in-repo.

## 8. Test & QA plan

- **Canonical framework:** `testdb` integration factory + `//go:build integration`; FE vitest. R1–R4 discipline.
- **Gates that apply (feature subset):** contract (regen no-drift), authz (unchanged path still green), DB-invariant (one-published holds), docs. Multi-tenant/async gates **N/A**.
- **Backend:** flip the contract/invariant guard tests `templates/application/lifecycle_test.go:310-440` to assert **NO** next draft + **no new version row** after approve/publish (these are guards — repair, not delete). Add lifecycle e2e: create→submit→review→approve→publish ⇒ exactly one version, status `published`, no draft `v2`; then manual `POST .../versions` ⇒ draft `v2` appears.
- **FE:** vitest — `StepConfirm` shows code / "…" loading / fallback (never bare "???") mirroring `CodePreviewBanner`; template detail route + list-click-to-detail; shared approval screen renders for both kinds.
- **Evidence shape:** `go build ./...`, `go test ./...`, `go test -tags=integration ./...`, `make test`, `npm run typecheck` web, `oapi-codegen` no-drift, `.\scripts\check-system-runnable.ps1`.
- **Final acceptance:** Preview-driven real-user QA — template flow (create → detail screen → editor → submit → review → approve → publish, **assert no auto v2**, then manual new version) and document flow (wizard shows real number, create → detail → submit → signoff → publish), logged in with the `approver` QA account (SoD). Zero errors = PASS.

## 9. Docs / ADR

- **Wiki:** update `wiki/modules/templates.md` (remove "auto-spawn next draft" as intended behavior; document manual-only revision), `wiki/modules/documents.md` (cross-link shared shell), and a new/updated FE structure note for the shared controlled-artifact view layer. Refresh `Last verified`.
- **REQ IDs:** cite from `wiki/architecture/backend-target-architecture.md` (lifecycle/contract REQs) during review.
- **ADR required? YES (two):**
  1. **Revert template auto-next-draft** — supersedes the 2026-05-31 `feat/templates-approve-next-draft-response` documented decision (a standing policy change: release no longer creates a draft). Cite ISO/industry norm: new revision is a manual act.
  2. **Shared controlled-artifact FE view layer** — records the dedup decision (templates & documents render one parameterized shell), so future screens don't re-fork.

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — proceed to brainstorming. Two ADRs flagged; one contract change (regen) carried as a constraint. No Red.
- **Open hard-stops:** none. AS-2 (FE duplication) is resolved by explicit operator authorization → locked design constraint, not a STOP. No AS-1, no AS-3.
- **Locked constraints handed to brainstorming:**
  1. **Templates align TO documents**, never the reverse. Remove `spawnNextDraft` from `Approve`+`PublishTemplateVersion`; the **only** revision-creation path is manual `CreateNextVersion` (`POST .../versions`). Drop `next_draft*` from both response contracts + regen; FE-first expand/contract.
  2. **Do NOT merge backend modules.** Dedup happens in the **frontend** via a shared controlled-artifact view layer (detail shell + approval screen + metadata sidebar), parameterized by `kind`. Both features render through it. New FE ADR.
  3. **#4 is FE-only:** mirror `CodePreviewBanner` ready/isLoading/data states in `StepConfirm`; backend `PreviewCode` endpoint is correct and unchanged.
  4. **#5 is a no-op on documents** (already manual). Treat the asymmetry as the bug.
  5. **Decision to make in brainstorming:** unify lifecycle status string `in_review` (templates) → `under_review` (documents)? Blast radius = contract + DB enum + FE; **propose DEFER** unless free — not required for any operator ask; if done, expand/contract migration. Carry as an explicit yes/no.
  6. Contract-first + canonical test framework + RFC-9457 errors + two-tier authz remain inviolable.
