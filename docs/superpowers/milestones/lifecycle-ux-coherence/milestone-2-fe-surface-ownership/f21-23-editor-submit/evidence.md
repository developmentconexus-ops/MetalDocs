# Feature F2.1–F2.3 (T1) — Editor Submit: Evidence

> **Milestone:** 2 · **Task:** `T1` (F2.1 editor-submit-unify · F2.2 editor-reason-for-change · F2.3 editor-polish) · **Closed:** 2026-07-07
> **Findings:** 8, 6, 13 · **Rules:** R5 (one impl, N entry points), R1 (author submits from authoring context), contract-first.

## What was implemented

### F2.1 — one document-submit client (finding 8, R5)
- Deleted the second document-submit path `finalizeDocument` (+ `FinalizeDocumentRequest` /
  `FinalizeDocumentResult` type exports) from `documents/api/documents.ts`. The deprecated
  `/documents/{id}/finalize` wrapper is gone.
- The editor's sole submit call is `approvalApi.submit` (imported as `submitForReviewRequest`,
  `DocumentEditorPage.tsx:18`, invoked `:294`) — the canonical `POST /documents/{id}/submit`
  via `mutationClient` (etag/If-Match/on412). FE sends **no** `route_id`/`content_hash`
  (ADR 0073 in-tx resolution — M1 producer).
- Grep proof: `submitDocumentForReview` and `finalizeDocument` have **zero** references across `src/`.

### F2.2 — REV≥1 reason dialog (finding 6)
- Submit gate keys on `revision_number` (`handleSubmitForReview`): REV0 → `submitForReview({})`
  title-only, no dialog; REV≥1 → opens the reason dialog.
- Dialog (`handleConfirmSubmitForReview`) collects `revision_title` (required) +
  `reason_for_change` (required) + `reason_category` (optional enum, omitted from body when blank).
- Client pre-validation blocks empty título/motivo with PT-BR field errors before the network call.
- Server 422 typed-code parity: catch routes `validation.revision_title_required`,
  `validation.reason_for_change_required`, `validation.reason_category_invalid` to the matching
  field error state. PT-BR strings added to `lib/api/errorMessages.ts`; the generated
  `error-codes.generated.json` regenerated from `internal/` so the bidirectional coverage test holds.

### F2.3 — editor polish (finding 13)
- `isSubmitting` guard prevents double-submit (button disabled + early return while in flight).
- Strings corrected: "finalizar" → "submeter" family; success toast
  `Documento submetido para revisão.`

## Verification

| Check | Command / action | Result |
|-------|------------------|--------|
| One submit path | `grep -rn "submitDocumentForReview\|finalizeDocument" src/` | **NONE** (R5 held) |
| Editor uses canonical client | grep `DocumentEditorPage.tsx` | `submit as submitForReviewRequest` from `approvalApi` only |
| Error-code parity | `errorMessages.coverage.test.ts` (both directions) | PASS |
| FE suite | `make test` (vitest full) | **751 pass / 122 files** |
| Touched suites | `DocumentEditorPage.test.tsx`, `DocumentEditorPage.E3.test.ts`, `documents.test.ts`, `errors.test.ts` | PASS |

### LIVE QA (preview :4173, real API)
- **REV0 submit** — doc `6d71db5e` (PO-RH-002), no dialog → `POST /api/v1/documents/6d71db5e…/submit` → **201**. Screenshot captured.
- **REV≥1 submit** — doc `aea9239d` (PO-RH-001 rev1): dialog collected Título + Motivo + Categoria.
  - Empty-field submit → blocked client-side, 2 field errors shown (no network call).
  - Filled submit → `POST /api/v1/documents/aea9239d-3cc6-4e6b-8a67-ff1cfa1c6887/submit` → **201**
    `{instance_id:"11180685-…", was_replay:false, etag:"\"v1\""}`. Fresh instance ⇒ server accepted
    `reason_for_change` (REV≥1 gate). `revision_title` "Ajuste de conteúdo REV1" persisted. Dialog screenshot captured.

## Acceptance vs spec

| Criterion | Met? | Evidence |
|-----------|------|----------|
| One document-submit client; second path gone | yes | grep NONE; `finalizeDocument` deleted |
| REV0 title-only path intact | yes | live REV0 201, no dialog |
| REV≥1 collects reason_for_change (+ optional category) | yes | live REV≥1 201, dialog fields |
| 422 typed codes surfaced to fields | yes | catch mapping + coverage test |
| submit guard + correct strings + success feedback | yes | `isSubmitting`, "submeter", success toast |
| FE tests green | yes | 751 pass |

## Review disposition
- Spec-compliance: PASS — R5 (single client) + R1 (submit only from editor) held; contract-first
  (types from `api-types`, no hand-rolled DTOs); no new tracking screen (YAGNI §4).
- Code-quality: PASS — reason dialog state colocated in the editor; error mapping table-driven.

## Bounded defers
| Defer | Why bounded | Trigger / owner |
|-------|-------------|-----------------|
| REV≥1 live path used API-created fixture (no seeded rev-of-published) | happy path + client validation both driven live; server gate proven by fresh-instance 201 | drive from a seeded rev fixture when one exists; owner documents editor |
