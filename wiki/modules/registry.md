# Module: Registry (Controlled Documents)

> **Last verified:** 2026-05-04
> **Scope:** Controlled-document catalog — code generation, CRUD routes, active-document lookup (FULL OUTER JOIN for published-only state), frontend registry pages.
> **Out of scope:** Document versions / editor (see `modules/documents.md`), taxonomy profiles + areas (see `modules/taxonomy.md`), approval (see `modules/approval.md`).
> **Key files:**
> - `internal/modules/registry/delivery/http/routes.go:31` — `activeDocumentResponse` struct — all fields pointer+omitempty (E10 fix)
> - `internal/modules/registry/delivery/http/routes.go:91` — `getActiveDocument` handler — FULL OUTER JOIN so published-only CDs return 200 with `publishedDocumentId`
> - `internal/modules/registry/delivery/http/routes_contract_test.go` — E10 contract tests (3 new cases)
> - `frontend/apps/web/src/features/registry/api.ts:63` — `ActiveDocumentResponse` interface — all fields optional; `ActiveDocumentInstance` kept as deprecated alias
> - `frontend/apps/web/src/features/registry/RegistryDetailPage.tsx:27` — detail page; renders published banner + Nova Revisão form when only published revision exists
> - `frontend/apps/web/src/features/registry/PublishedDownloadCell.tsx:4` — polls PDF status for a published document via `useDocumentPdfStatus`
> - `frontend/apps/web/src/features/registry/RegistryListPage.tsx` — CD list

---

## What the registry module does

The **registry** module owns the `controlled_documents` table — the catalog of code-numbered document slots. Each CD slot:

- Binds a **profile** + **area** → auto-generates a unique code (`{profile_code}-{area_code}-{seq}`).
- Has a lifecycle status: `active | obsolete | superseded`.
- Can have zero or more **document versions** attached (those live in `public.documents`).

The registry module does **not** own document content or approval state — it is the index.

## `getActiveDocument` endpoint (E10 fix)

`GET /api/v2/controlled-documents/{id}/active-document`

**Before this fix:** used a simple `WHERE status IN ('draft',...)` query. If the only version for the CD was `published`, the query returned no rows → 404. Frontend could not render the published download link.

**After this fix (commit 1dfcf3da):** uses a FULL OUTER JOIN between:
- `active` — any version with status in `draft | under_review | approved | rejected | scheduled` (LIMIT 1)
- `pub` — the latest `published` version (ORDER BY `revision_number DESC`, LIMIT 1)

Returns 200 as long as at least one side is non-NULL. If only the `pub` side exists, the response contains only `publishedDocumentId`. If both sides exist, the response contains both `documentId` and `publishedDocumentId`.

Returns 404 only when both sides are NULL (CD truly has no versions at all).

### Response shape (`activeDocumentResponse`)

All fields are pointer+omitempty. Consumers must treat all fields as optional:

```go
// internal/modules/registry/delivery/http/routes.go:31
type activeDocumentResponse struct {
    DocumentID          *string `json:"documentId,omitempty"`
    ApprovalState       *string `json:"approvalState,omitempty"`
    ContentHash         *string `json:"contentHash,omitempty"`
    RevisionVersion     *int    `json:"revisionVersion,omitempty"`
    PublishedDocumentID *string `json:"publishedDocumentId,omitempty"`
    ApprovalInstanceID  *string `json:"approvalInstanceId,omitempty"`
}
```

```typescript
// frontend/apps/web/src/features/registry/api.ts:63
export interface ActiveDocumentResponse {
  documentId?: string;
  approvalState?: string;
  contentHash?: string;
  revisionVersion?: number;
  publishedDocumentId?: string;
  approvalInstanceId?: string;
}
/** @deprecated use ActiveDocumentResponse */
export type ActiveDocumentInstance = ActiveDocumentResponse;
```

## Frontend: RegistryDetailPage

`RegistryDetailPage` (`frontend/apps/web/src/features/registry/RegistryDetailPage.tsx`) shows a controlled document's metadata and its current revision state. After the E10 fix it handles three cases:

| Scenario | What renders |
|---|---|
| Active draft/in-review version exists | `RegistryDetailPanel` with edit/approval actions |
| Only a published version exists | Green "Revisão publicada" banner + `PublishedDownloadCell` |
| No active version + CD is active | "Nenhuma revisão ativa" placeholder + **Nova Revisão** inline form |

**Nova Revisão form:** lets the user name a new document and calls `POST /api/v2/documents` (createDocument). On success, navigates to the editor if `onOpenDocumentEditor` prop is set, otherwise reloads the detail page.

## PublishedDownloadCell

`frontend/apps/web/src/features/registry/PublishedDownloadCell.tsx` — thin wrapper that uses `useDocumentPdfStatus` (always enabled) and renders `PDFCell`. It shows the PDF download link for a published revision directly on the registry detail page, polling until the PDF is ready.

## API routes

| Method | Path | Handler |
|---|---|---|
| GET | `/api/v2/controlled-documents` | `listDocs` |
| POST | `/api/v2/controlled-documents` | `createDoc` |
| GET | `/api/v2/controlled-documents/{id}` | `getDoc` |
| GET | `/api/v2/controlled-documents/{id}/active-document` | `getActiveDocument` (FULL OUTER JOIN) |
| PUT | `/api/v2/controlled-documents/{id}/obsolete` | `obsoleteDoc` |
| PUT | `/api/v2/controlled-documents/{id}/supersede` | `supersedeDoc` |

## Cross-refs

- [concepts/controlled-documents.md](../concepts/controlled-documents.md) — what a CD is, code format
- [modules/documents.md](documents.md) — document versions that hang off a CD
- [modules/taxonomy.md](taxonomy.md) — profiles + areas that CDs bind to
- [workflows/user-onboarding.md](../workflows/user-onboarding.md) — Step 5 (register CD) + Step 6 (generate version)
