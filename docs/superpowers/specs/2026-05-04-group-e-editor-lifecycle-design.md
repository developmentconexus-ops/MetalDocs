# Group E (sub-plan 2) — Editor Lifecycle UX Design

> **Status:** approved 2026-05-04
> **Scope:** Fix bugs E1, E10, E11 from `wiki/bugs/audit-2026-05-03.md` (lines 271, 280, 281). Make UI truth track backend document lifecycle: editor mode follows doc status, registry surfaces published doc even when no active workflow exists, async PDF generation is observable to the user.
> **Out of scope:** E2/E3/E4 (sub-plan 1 — error UX, already shipped). E5/E7/E9/E12 (sub-plan 3 — misc UI fixes).

---

## Why This Spec Exists

Three independent bugs all stem from the UI failing to track post-submission backend state:

- **E1** — Editor mode is gated on session phase only (`DocumentEditorPage.tsx:233`). After `Finalizar`, doc status moves to `under_review` but the editor stays in `document-edit` mode until the user reloads. Risk: edits land on a frozen revision.
- **E10** — `getActiveDocument` SQL (`internal/modules/registry/delivery/http/routes.go:128`) excludes `published` from its `WHERE` clause. When a controlled document has only a published revision and no in-flight work, the endpoint 404s. Frontend treats null as "no instance" → "Nova Revisão" CTA shown but the existing published revision is invisible.
- **E11** — Freeze invokes synchronous DOCX fanout, then dispatches async `docgen_v2_pdf` via outbox. `final_pdf_s3_key` is NULL until the worker writes it. The view endpoint surfaces the URL only when present; the frontend never polls. Users see no download link post-freeze.

All three are fixable with thin, additive changes — no schema migration, no new infra.

---

## Architecture: UI Truth Follows Backend Lifecycle

```
                    ┌────────────────────┐
                    │  Backend (Go)      │
                    │  ─ active-doc API  │  E10: 200 + {active?,published?}
                    │  ─ view API        │  E11: + pdf_status field
                    │  ─ async PDF       │  (existing outbox worker)
                    └─────────┬──────────┘
                              │
       ┌──────────────────────┼──────────────────────┐
       │                      │                      │
   ┌───▼────────┐    ┌────────▼─────────┐    ┌──────▼─────────────┐
   │ Editor     │    │ useDocumentPdf   │    │ Registry detail    │
   │ docStatus  │    │ Status hook      │    │ active + published │
   │ gate (E1)  │    │ poll 3s/60s (E11)│    │ both visible (E10) │
   └────────────┘    └──────────────────┘    └────────────────────┘
```

**Backend touch:**
- `internal/modules/registry/delivery/http/routes.go` — restructure `getActiveDocument` query and response (E10)
- `internal/modules/documents/application/view_service.go` — add `pdf_status` field with `pending`/`ready`/`failed` semantics (E11)
- Zero migration, zero schema change

**Frontend touch:**
- `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx` — derive `isEditable` from `phase + docStatus`, refetch on window focus, surface PDF link (E1, E11)
- `frontend/apps/web/src/features/documents/v2/hooks/useDocumentPdfStatus.ts` — new polling hook (E11)
- `frontend/apps/web/src/features/registry/api.ts` — update `ActiveDocumentInstance` type contract (E10)
- `frontend/apps/web/src/features/registry/RegistryDetailPage.tsx` + `frontend/apps/web/src/features/approval/components/RegistryDetailPanel.tsx` — render published-only state, embed PDF download (E10, E11)

**Migration discipline:** Phase 1–2 build backend changes (codex). Phase 3 builds frontend hook in parallel with E1 editor gate (sonnet ‖ sonnet). Phases 4–5 wire frontend consumers (sonnet). Phase 6 verifies (vitest + go test + smoke + codex audit 3/3). Phase 7 closes audit + wiki.

---

## Per-Bug Fix Design

### E1 — Editor stays editable after submission

**Files:**
- Modify: `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx`

**Problem:** `mode={session.state.phase === 'writer' ? 'document-edit' : 'readonly'}` does not consult `docStatus`. After `Finalizar`, status moves to `under_review` but the writer session does not auto-release; mode stays `document-edit` until reload.

**Fix:**

1. Derive `isEditable` once:
   ```tsx
   const isEditable = session.state.phase === 'writer' && docStatus === 'draft';
   ```

2. Use it for both editor mode and finalize disabled state:
   ```tsx
   <button disabled={!isEditable}>Finalizar</button>
   <MetalDocsEditor mode={isEditable ? 'document-edit' : 'readonly'} ... />
   ```

3. Refetch document on window focus to pick up out-of-band status changes (submit from another tab, inbox, etc.):
   ```tsx
   useEffect(() => {
     const onFocus = () => {
       void getDocument(documentID).then(setDoc).catch(() => {});
     };
     window.addEventListener('focus', onFocus);
     return () => window.removeEventListener('focus', onFocus);
   }, [documentID]);
   ```

**Why not auto-release session on status change:** Risk of losing reacquire if status reverts to draft (rejected approval). Status gate alone is sufficient; session lifecycle stays unchanged.

**Test:** vitest mounts `DocumentEditorPage` with mocked `getDocument` returning `status='draft'` and writer phase, asserts editor renders with `mode='document-edit'`. Update mock to return `status='under_review'`, dispatch focus event, assert `mode='readonly'` after re-render.

---

### E10 — Active-document endpoint hides published revisions

**Files:**
- Modify: `internal/modules/registry/delivery/http/routes.go` (lines 91–160)
- Modify: `internal/modules/registry/delivery/http/routes_contract_test.go` (add three cases)
- Modify: `frontend/apps/web/src/features/registry/api.ts` (response type)
- Modify: `frontend/apps/web/src/features/registry/RegistryDetailPage.tsx` (consume new shape)

**Problem:** SQL `WHERE d.status IN ('draft','under_review','approved','rejected','scheduled')` excludes `published`. When a controlled document has only a published revision (no in-flight cycle), the query returns zero rows → 404 → frontend treats response as null → "Nova Revisão" CTA shown but `publishedDocumentId` never populated.

**Fix — Backend:**

Restructure handler to merge active and published lookups into one round-trip. Always return 200 if either side has data.

New response shape (additive — `documentId` becomes nullable):
```go
type activeDocumentResponse struct {
    DocumentID          *string `json:"document_id,omitempty"`
    ApprovalState       *string `json:"approval_state,omitempty"`
    ContentHash         *string `json:"content_hash,omitempty"`
    RevisionVersion     *int    `json:"revision_version,omitempty"`
    PublishedDocumentID *string `json:"published_document_id,omitempty"`
    ApprovalInstanceID  *string `json:"approval_instance_id,omitempty"`
}
```

New query (`FULL OUTER JOIN` of two single-row subqueries):
```sql
SELECT
  active.id, active.content_hash, active.rev, active.approval_state,
  pub.id AS published_id
FROM (
  SELECT d.id::text AS id,
         COALESCE(d.content_hash_at_submit,
                  (SELECT r.content_hash FROM document_revisions r
                    WHERE r.document_id = d.id
                    ORDER BY r.created_at DESC LIMIT 1),
                  '') AS content_hash,
         COALESCE(d.revision_version, 0) AS rev,
         COALESCE(
           (SELECT CASE ai.status
              WHEN 'in_progress' THEN 'under_review'
              WHEN 'approved'    THEN 'approved'
              WHEN 'scheduled'   THEN 'scheduled'
              WHEN 'rejected'    THEN 'rejected'
              WHEN 'cancelled'   THEN 'cancelled'
            END
            FROM approval_instances ai
            WHERE ai.document_v2_id = d.id
            ORDER BY ai.submitted_at DESC LIMIT 1),
           'draft') AS approval_state
    FROM documents d
   WHERE d.tenant_id = $1::uuid
     AND d.controlled_document_id = $2::uuid
     AND d.status IN ('draft','under_review','approved','rejected','scheduled')
   LIMIT 1
) active
FULL OUTER JOIN (
  SELECT id::text AS id
    FROM documents
   WHERE tenant_id = $1::uuid
     AND controlled_document_id = $2::uuid
     AND status = 'published'
   ORDER BY revision_number DESC LIMIT 1
) pub ON TRUE;
```

If `Scan` returns `sql.ErrNoRows` (no row at all) → 404. Else build response, populating only fields that are non-null. Lookup of `approval_instance_id` (existing second query at line 146) only runs when `active.id` is present.

**Fix — Frontend:**

Update type:
```ts
export interface ActiveDocumentResponse {
  documentId?: string;
  approvalState?: string;
  contentHash?: string;
  revisionVersion?: number;
  publishedDocumentId?: string;
  approvalInstanceId?: string;
}

export async function fetchActiveDocumentInstance(
  controlledDocumentId: string,
): Promise<ActiveDocumentResponse | null> {
  const res = await apiFetch<ActiveDocumentResponse>(
    `${BASE}/${encodeURIComponent(controlledDocumentId)}/active-document`,
  );
  return res; // null only if backend returns 404
}
```

`RegistryDetailPage.tsx` rendering logic:
- If `documentId` present → render existing `RegistryDetailPanel` (active flow)
- If `publishedDocumentId` present (with or without active) → render published banner with embedded PDF download via `useDocumentPdfStatus`
- If neither (404 case) → existing empty state with "Nova Revisão"

When both are present, show stacked panels: published reference banner above active workflow panel.

**Test:**
- Go: `routes_contract_test.go` adds three cases:
  - `TestActiveDocument_OnlyPublished_Returns200_WithPublishedID`
  - `TestActiveDocument_BothActiveAndPublished_Returns200_WithBoth`
  - `TestActiveDocument_NoneExist_Returns404`
- vitest: `RegistryDetailPage` with mocked `{publishedDocumentId only}` asserts published banner renders, Nova Revisão CTA visible, `RegistryDetailPanel` not mounted.

---

### E11 — Async PDF readiness invisible to user

**Files:**
- Modify: `internal/modules/documents/application/view_service.go`
- Modify: `internal/modules/documents/application/view_service_test.go`
- Create: `frontend/apps/web/src/features/documents/v2/hooks/useDocumentPdfStatus.ts`
- Create: `frontend/apps/web/src/features/documents/v2/hooks/useDocumentPdfStatus.test.ts`
- Modify: `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx` (consumer)
- Modify: `frontend/apps/web/src/features/approval/components/RegistryDetailPanel.tsx` (consumer)

**Problem:** Freeze fanout produces DOCX synchronously, dispatches `docgen_v2_pdf` event to outbox. PDF worker eventually writes `final_pdf_s3_key`. View endpoint already presigns when key is set, but returns no signal otherwise. Frontend has no way to know when to retry.

**Fix — Backend:**

Extend `ViewResponse`:
```go
type ViewResponse struct {
    Status    string  `json:"status"`
    AreaCode  string  `json:"area_code"`
    PDFStatus string  `json:"pdf_status"` // "pending" | "ready" | "failed"
    PDFURL    *string `json:"pdf_url,omitempty"`
}
```

`view_service.go` logic:
```go
if !pdfKey.Valid || pdfKey.String == "" {
    // probe outbox: if row state="failed" after retries exhausted → "failed", else "pending"
    state, err := s.pdfOutboxState.Read(ctx, tenantID, documentID)
    if err == nil && state == "failed" {
        resp.PDFStatus = "failed"
    } else {
        resp.PDFStatus = "pending"
    }
} else {
    resp.PDFStatus = "ready"
    url, err := s.presigner.PresignObjectGET(ctx, pdfKey.String)
    if err != nil {
        return ViewResponse{}, err
    }
    resp.PDFURL = &url
}
```

New dependency `pdfOutboxState` is a thin reader interface implemented by `pdf_outbox_repository`:
```go
type PDFOutboxStateReader interface {
    Read(ctx context.Context, tenantID, revisionID string) (string, error)
}
```

If outbox reader is nil (test fixtures), default to `"pending"`.

**Fix — Frontend hook:**

```ts
// hooks/useDocumentPdfStatus.ts
import { useEffect, useRef, useState } from 'react';
import { apiFetch } from '../../../../lib/api/client';

type PDFStatus = 'pending' | 'ready' | 'failed';
type ViewResponse = { pdf_status: PDFStatus; pdf_url?: string };

export type DocumentPdfStatus = {
  status: PDFStatus;
  url?: string;
  retry: () => void;
};

const POLL_INTERVAL_MS = 3_000;
const TIMEOUT_MS = 60_000;

export function useDocumentPdfStatus(documentID: string, enabled: boolean): DocumentPdfStatus {
  const [data, setData] = useState<{ status: PDFStatus; url?: string }>({ status: 'pending' });
  const [tick, setTick] = useState(0);
  const startedAt = useRef(0);

  useEffect(() => {
    if (!enabled || !documentID) return;
    let cancelled = false;
    let timer = 0;
    startedAt.current = Date.now();
    setData({ status: 'pending' });

    const poll = async () => {
      try {
        const v = await apiFetch<ViewResponse>(
          `/api/v2/documents/${encodeURIComponent(documentID)}/view`,
        );
        if (cancelled) return;
        setData({ status: v.pdf_status, url: v.pdf_url });
        if (v.pdf_status === 'ready' || v.pdf_status === 'failed') return;
        if (Date.now() - startedAt.current > TIMEOUT_MS) {
          setData({ status: 'failed' });
          return;
        }
      } catch {
        // network error — try again next tick
      }
      timer = window.setTimeout(poll, POLL_INTERVAL_MS);
    };
    void poll();

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [documentID, enabled, tick]);

  return { ...data, retry: () => setTick((n) => n + 1) };
}
```

**Consumers:**

`DocumentEditorPage.tsx` — when `docStatus !== 'draft'`, render PDF cell in `overlayRight`:
```tsx
const pdf = useDocumentPdfStatus(documentID, docStatus !== 'draft');

{docStatus !== 'draft' && (
  <PDFCell status={pdf.status} url={pdf.url} onRetry={pdf.retry} />
)}
```

`RegistryDetailPanel.tsx` — when `publishedDocumentId` present, render PDF cell using that ID.

`PDFCell` is a small inline render (no separate file needed):
- `pending` → spinner + `"Gerando PDF..."`
- `ready` → `<a href={url} download>Baixar PDF</a>`
- `failed` → `"Falha ao gerar PDF"` + retry button

**Test — Backend:**
- sqlmock cases:
  - `pdfKey` NULL + outbox state `pending` → `"pending"`, no URL
  - `pdfKey` NULL + outbox state `failed` → `"failed"`, no URL
  - `pdfKey` set → `"ready"` + presigned URL
  - `pdfKey` NULL + outbox reader nil → `"pending"`

**Test — Frontend:**
- vitest + msw + fake timers:
  - Hook polls `pending` twice, then `ready` → asserts URL exposed, polling stopped
  - Hook polls past 60s timeout → asserts `status='failed'`
  - `retry()` resets and re-polls
  - `enabled=false` → no requests

---

## Rollout Plan

| Phase | Tasks | Parallelism | Model |
|---|---|---|---|
| 0 | Worktree, codex spec validate | sequential | sonnet |
| 1 | Backend E10: SQL restructure + contract tests (3 new cases) | sequential | codex |
| 2 | Backend E11: view_service `pdf_status` + outbox reader + sqlmock tests | sequential | codex |
| 3 | Frontend `useDocumentPdfStatus` hook + tests ‖ Frontend E1 editor gate + tests | parallel | sonnet ‖ sonnet |
| 4 | Frontend E10: api.ts type + RegistryDetailPage + RegistryDetailPanel published banner + tests | sequential | sonnet |
| 5 | Frontend E11 wire-up: DocumentEditorPage PDFCell + window focus refetch + tests | sequential | sonnet |
| 6 | Verify: vitest + go test + smoke flows + codex audit 3/3 | sequential | sonnet → codex audit |
| 7 | Audit doc closure + wiki-curator + finishing-a-development-branch | sequential | sonnet |

**Phase review after each:** opus.

---

## Testing Strategy

**Per-bug:** see "Per-Bug Fix Design".

**Cross-cutting:**
- `npx vitest run` — full pass
- `go test -mod=mod ./...` — backend regression + new contract/sqlmock cases
- Smoke E1: open draft doc → click Finalizar → editor flips to readonly without reload
- Smoke E10: registry detail of doc with only a published revision → published banner with download link visible + Nova Revisão CTA still works
- Smoke E11: freeze a doc → observe `Gerando PDF...` spinner → ~10s later download link appears (depends on docgen worker latency)
- Codex independent audit: PASS/FAIL per bug with file:line evidence
- Wiki-curator: refresh stamps on `wiki/modules/frontend-*.md` (if exists), update `wiki/concepts/document-lifecycle.md` (or create if absent) to document the active+published+pdf_status contract

**Coverage targets:** new `useDocumentPdfStatus.ts` ≥85% line; `view_service.go` additions ≥85% line. No new lint warnings.

---

## Acceptance Criteria

- [ ] E1: post-finalize editor renders readonly without reload (status gate)
- [ ] E1: window focus refetches doc status (out-of-band submit case)
- [ ] E10: `active-document` endpoint returns 200 with `published_document_id` when only published revision exists
- [ ] E10: 404 only when zero rows of either kind
- [ ] E10: registry detail shows published banner + Nova Revisão button when only published exists
- [ ] E10: registry detail shows both banners when active and published coexist
- [ ] E11: view endpoint returns `pdf_status` ∈ {pending, ready, failed}
- [ ] E11: editor + registry panel poll until ready, show download link
- [ ] E11: 60s timeout shows retry button; retry re-polls
- [ ] All vitest pass
- [ ] `go test -mod=mod ./...` passes
- [ ] Codex independent audit returns 3/3 PASS
- [ ] Audit doc updated, E1/E10/E11 closed with commit SHAs

---

## Open Questions

None.

---

## References

- Audit: `wiki/bugs/audit-2026-05-03.md` (E1 line 271, E10 line 280, E11 line 281)
- Sub-plan 1 (error UX): `docs/superpowers/specs/2026-05-03-group-e-error-ux-design.md`
- Sub-plan 1 plan: `docs/superpowers/plans/2026-05-03-group-e-error-ux.md`
- Sub-plan 3 (misc): `docs/superpowers/specs/2026-05-04-group-e-misc-design.md` (TBD)
- Editor entry: `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx`
- Active-doc handler: `internal/modules/registry/delivery/http/routes.go:91`
- View handler: `internal/modules/documents/application/view_service.go:49`
- PDF outbox: `internal/modules/render/fanout/pdf_outbox_repository.go`, `pdf_outbox_worker.go`
