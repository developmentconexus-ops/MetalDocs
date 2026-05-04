# Group E Sub-Plan 2 — Editor Lifecycle UX Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make UI track backend document lifecycle so post-submit editor locks down (E1), registry surfaces published revisions (E10), and async PDF readiness is observable to the user (E11).

**Architecture:** Two thin backend changes (active-document SQL restructure; view endpoint adds `pdf_status`). One new frontend polling hook. Editor adds derived `isEditable` gate plus window focus refetch. Registry detail renders both active and published when present.

**Tech Stack:** Go 1.22, PostgreSQL 16, React 18 + TypeScript + Vite, vitest + msw, sqlmock, Sonner.

**Spec:** `docs/superpowers/specs/2026-05-04-group-e-editor-lifecycle-design.md` (commit `4d2c9a56`).

---

## Model Routing

| Phase | Role | Model |
|---|---|---|
| 0 | Worktree + codex spec validate | sonnet (controller) → codex (validator) |
| 1 | Backend E10 SQL + contract tests | **codex** (multi-edge query) |
| 2 | Backend E11 view_service + outbox reader + tests | **codex** (interface design + sqlmock) |
| 3a | Frontend `useDocumentPdfStatus` hook + tests | sonnet (mechanical) |
| 3b | Frontend E1 editor gate + focus refetch + tests | sonnet (mechanical) |
| 4 | Frontend E10 wiring + tests | sonnet |
| 5 | Frontend E11 wiring + tests | sonnet |
| 6 | Verify + smoke + codex audit | sonnet → **codex audit** |
| 7 | Audit doc closure + wiki-curator + branch-finishing | sonnet + wiki-curator agent |
| Phase reviews | between phases | **opus** (review only) |

Phase 3a ‖ 3b run in parallel — different files, no overlap.

---

## File Structure

| File | Responsibility | Action |
|---|---|---|
| `internal/modules/registry/delivery/http/routes.go` | `getActiveDocument` handler | Modify (E10) |
| `internal/modules/registry/delivery/http/routes_contract_test.go` | Contract tests | Modify (3 new cases) |
| `internal/modules/documents/http/view_handler.go` | View HTTP shape + error mapping | Modify (E11) |
| `internal/modules/documents/http/view_handler_test.go` | View handler tests | Modify (E11) |
| `internal/modules/documents/application/view_service.go` | View business logic | Modify (E11) |
| `internal/modules/documents/application/view_service_test.go` | View service sqlmock tests | Create or modify (E11) |
| `internal/modules/render/fanout/pdf_outbox_repository.go` | PDF outbox reader | Modify (add `ReadState`) |
| `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx` | Editor page | Modify (E1, E11) |
| `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.test.tsx` | Editor tests | Create |
| `frontend/apps/web/src/features/documents/v2/hooks/useDocumentPdfStatus.ts` | PDF polling hook | Create (E11) |
| `frontend/apps/web/src/features/documents/v2/hooks/useDocumentPdfStatus.test.ts` | Hook tests | Create |
| `frontend/apps/web/src/features/registry/api.ts` | Active doc client | Modify (E10) |
| `frontend/apps/web/src/features/registry/RegistryDetailPage.tsx` | Registry detail page | Modify (E10) |
| `frontend/apps/web/src/features/registry/RegistryDetailPage.test.tsx` | Registry tests | Create |
| `frontend/apps/web/src/features/approval/components/RegistryDetailPanel.tsx` | Active workflow panel | Modify (E11 PDF) |
| `wiki/concepts/document-lifecycle.md` | Lifecycle concept doc | Create or refresh |
| `wiki/bugs/audit-2026-05-03.md` | Audit closure | Modify |

---

## Phase 0 — Setup

### Task 0.1: Worktree

**Files:** none (workspace setup)

- [ ] **Step 1: Create worktree**

```bash
git worktree add -b phase-e-editor-lifecycle ../metaldocs-phase-e-editor main
cd ../metaldocs-phase-e-editor
```

- [ ] **Step 2: Verify clean baseline**

```bash
go test -mod=mod ./... 2>&1 | tail -5
cd frontend/apps/web && npx vitest run 2>&1 | tail -5
```

Expected: all tests pass on baseline.

### Task 0.2: Codex spec validation

**Files:** none (validator pass)

- [ ] **Step 1: Dispatch codex:codex-rescue subagent with prompt**

> Validate spec at `docs/superpowers/specs/2026-05-04-group-e-editor-lifecycle-design.md`. Report PASS/FAIL on: (1) E10 SQL safety with FULL OUTER JOIN; (2) view_service backward compat (existing `signed_url` field preserved); (3) hook polling logic correctness for race between unmount and async resolve; (4) E1 gate covers all submit paths (Finalizar button + out-of-band). Cite file:line for every concern.

- [ ] **Step 2: Address PASS/FAIL findings inline if any.** No commit needed.

---

## Phase 1 — Backend E10: Active-Document Restructure

### Task 1.1: Contract test — only published exists

**Files:**
- Modify: `internal/modules/registry/delivery/http/routes_contract_test.go`

- [ ] **Step 1: Add failing test**

```go
func TestActiveDocument_OnlyPublished_Returns200_WithPublishedID(t *testing.T) {
    db, mock := newSQLMock(t)
    defer db.Close()

    tenantID := "11111111-1111-1111-1111-111111111111"
    cdID := "22222222-2222-2222-2222-222222222222"
    pubID := "33333333-3333-3333-3333-333333333333"

    rows := sqlmock.NewRows([]string{"id", "content_hash", "rev", "approval_state", "published_id"}).
        AddRow(nil, nil, nil, nil, pubID)
    mock.ExpectQuery(`(?s)FROM \(.*FULL OUTER JOIN.*\) pub`).
        WithArgs(tenantID, cdID).
        WillReturnRows(rows)

    h := newTestHandler(db)
    req := newAuthedRequest(t, http.MethodGet, "/api/v2/controlled-documents/"+cdID+"/active-document", tenantID)
    rec := httptest.NewRecorder()
    h.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d, want 200", rec.Code)
    }
    var got struct {
        DocumentID          *string `json:"document_id"`
        PublishedDocumentID *string `json:"published_document_id"`
    }
    if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
        t.Fatal(err)
    }
    if got.DocumentID != nil {
        t.Errorf("DocumentID = %v, want nil", got.DocumentID)
    }
    if got.PublishedDocumentID == nil || *got.PublishedDocumentID != pubID {
        t.Errorf("PublishedDocumentID = %v, want %s", got.PublishedDocumentID, pubID)
    }
}
```

- [ ] **Step 2: Run, expect FAIL** (handler not yet refactored)

```bash
go test -run TestActiveDocument_OnlyPublished -mod=mod ./internal/modules/registry/delivery/http/...
```

Expected: FAIL — current handler returns 404.

### Task 1.2: Refactor handler — single FULL OUTER JOIN query, additive response shape

**Files:**
- Modify: `internal/modules/registry/delivery/http/routes.go` (lines 91–160)

- [ ] **Step 1: Replace `activeDocumentResponse` and handler body**

```go
type activeDocumentResponse struct {
    DocumentID          *string `json:"document_id,omitempty"`
    ApprovalState       *string `json:"approval_state,omitempty"`
    ContentHash         *string `json:"content_hash,omitempty"`
    RevisionVersion     *int    `json:"revision_version,omitempty"`
    PublishedDocumentID *string `json:"published_document_id,omitempty"`
    ApprovalInstanceID  *string `json:"approval_instance_id,omitempty"`
}

func (h *Handler) getActiveDocument(w http.ResponseWriter, r *http.Request) {
    tenantID := tenantIDFromRequest(r)
    cdID := r.PathValue("id")

    var (
        activeID        sql.NullString
        contentHash     sql.NullString
        revisionVersion sql.NullInt64
        approvalState   sql.NullString
        publishedID     sql.NullString
    )
    err := h.db.QueryRowContext(r.Context(), `
        SELECT active.id, active.content_hash, active.rev, active.approval_state, pub.id
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
          ) pub ON TRUE`,
        tenantID, cdID,
    ).Scan(&activeID, &contentHash, &revisionVersion, &approvalState, &publishedID)

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            httpresponse.WriteError(w, http.StatusNotFound, "NO_ACTIVE_INSTANCE", "no active or published document for this controlled document")
            return
        }
        httpresponse.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
        return
    }

    if !activeID.Valid && !publishedID.Valid {
        httpresponse.WriteError(w, http.StatusNotFound, "NO_ACTIVE_INSTANCE", "no active or published document for this controlled document")
        return
    }

    var resp activeDocumentResponse
    if activeID.Valid {
        s := activeID.String
        resp.DocumentID = &s
        if contentHash.Valid {
            ch := contentHash.String
            resp.ContentHash = &ch
        }
        if revisionVersion.Valid {
            rv := int(revisionVersion.Int64)
            resp.RevisionVersion = &rv
        }
        if approvalState.Valid {
            as := approvalState.String
            resp.ApprovalState = &as
        }
        var apprID sql.NullString
        _ = h.db.QueryRowContext(r.Context(), `
            SELECT id::text
              FROM approval_instances
             WHERE document_v2_id = $1::uuid
               AND tenant_id = $2::uuid
               AND status = 'in_progress'
             ORDER BY submitted_at DESC LIMIT 1`,
            activeID.String, tenantID,
        ).Scan(&apprID)
        if apprID.Valid {
            s := apprID.String
            resp.ApprovalInstanceID = &s
        }
    }
    if publishedID.Valid {
        s := publishedID.String
        resp.PublishedDocumentID = &s
    }
    httpresponse.WriteJSON(w, http.StatusOK, resp)
}
```

- [ ] **Step 2: Run new test, expect PASS**

```bash
go test -run TestActiveDocument_OnlyPublished -mod=mod ./internal/modules/registry/delivery/http/...
```

- [ ] **Step 3: Run full handler tests, expect PASS**

```bash
go test -mod=mod ./internal/modules/registry/delivery/http/...
```

### Task 1.3: Contract tests — both / neither cases

**Files:**
- Modify: `internal/modules/registry/delivery/http/routes_contract_test.go`

- [ ] **Step 1: Add two more failing tests**

```go
func TestActiveDocument_BothActiveAndPublished_Returns200_WithBoth(t *testing.T) {
    db, mock := newSQLMock(t)
    defer db.Close()

    tenantID := "11111111-1111-1111-1111-111111111111"
    cdID := "22222222-2222-2222-2222-222222222222"
    activeID := "44444444-4444-4444-4444-444444444444"
    pubID := "33333333-3333-3333-3333-333333333333"

    mock.ExpectQuery(`(?s)FROM \(.*FULL OUTER JOIN.*\) pub`).
        WithArgs(tenantID, cdID).
        WillReturnRows(sqlmock.NewRows([]string{"id", "content_hash", "rev", "approval_state", "published_id"}).
            AddRow(activeID, "abc123", 2, "under_review", pubID))
    mock.ExpectQuery(`SELECT id::text\s+FROM approval_instances`).
        WithArgs(activeID, tenantID).
        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("ai-1"))

    h := newTestHandler(db)
    req := newAuthedRequest(t, http.MethodGet, "/api/v2/controlled-documents/"+cdID+"/active-document", tenantID)
    rec := httptest.NewRecorder()
    h.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d, want 200", rec.Code)
    }
    var got struct {
        DocumentID          *string `json:"document_id"`
        PublishedDocumentID *string `json:"published_document_id"`
        ApprovalInstanceID  *string `json:"approval_instance_id"`
    }
    _ = json.Unmarshal(rec.Body.Bytes(), &got)
    if got.DocumentID == nil || *got.DocumentID != activeID {
        t.Errorf("DocumentID = %v", got.DocumentID)
    }
    if got.PublishedDocumentID == nil || *got.PublishedDocumentID != pubID {
        t.Errorf("PublishedDocumentID = %v", got.PublishedDocumentID)
    }
    if got.ApprovalInstanceID == nil || *got.ApprovalInstanceID != "ai-1" {
        t.Errorf("ApprovalInstanceID = %v", got.ApprovalInstanceID)
    }
}

func TestActiveDocument_NoneExist_Returns404(t *testing.T) {
    db, mock := newSQLMock(t)
    defer db.Close()
    tenantID := "11111111-1111-1111-1111-111111111111"
    cdID := "22222222-2222-2222-2222-222222222222"

    mock.ExpectQuery(`(?s)FROM \(.*FULL OUTER JOIN.*\) pub`).
        WithArgs(tenantID, cdID).
        WillReturnError(sql.ErrNoRows)

    h := newTestHandler(db)
    req := newAuthedRequest(t, http.MethodGet, "/api/v2/controlled-documents/"+cdID+"/active-document", tenantID)
    rec := httptest.NewRecorder()
    h.ServeHTTP(rec, req)

    if rec.Code != http.StatusNotFound {
        t.Fatalf("status = %d, want 404", rec.Code)
    }
}
```

- [ ] **Step 2: Run, expect PASS** (handler already supports both shapes)

```bash
go test -run "TestActiveDocument_(Both|NoneExist)" -mod=mod ./internal/modules/registry/delivery/http/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/modules/registry/delivery/http/routes.go internal/modules/registry/delivery/http/routes_contract_test.go
git commit -m "fix(registry): active-document endpoint returns 200 with published_document_id when only published exists (E10)"
```

---

## Phase 2 — Backend E11: View Endpoint pdf_status

### Task 2.1: PDF outbox state reader

**Files:**
- Modify: `internal/modules/render/fanout/pdf_outbox_repository.go`

- [ ] **Step 1: Add reader method**

```go
// ReadState returns the latest pdf_outbox row state for the given revision.
// Returns ("", nil) when no row exists.
func (r *PDFOutboxRepository) ReadState(ctx context.Context, tenantID, revisionID string) (string, error) {
    var state string
    err := r.db.QueryRowContext(ctx, `
        SELECT state
          FROM pdf_outbox
         WHERE tenant_id = $1::uuid AND revision_id = $2::uuid
         ORDER BY created_at DESC LIMIT 1`,
        tenantID, revisionID,
    ).Scan(&state)
    if errors.Is(err, sql.ErrNoRows) {
        return "", nil
    }
    if err != nil {
        return "", fmt.Errorf("pdf outbox read state: %w", err)
    }
    return state, nil
}
```

- [ ] **Step 2: Run package tests, expect PASS**

```bash
go test -mod=mod ./internal/modules/render/fanout/...
```

### Task 2.2: ViewResult adds PDFStatus, drop ErrPDFPending sentinel

**Files:**
- Modify: `internal/modules/documents/http/view_handler.go`
- Modify: `internal/modules/documents/http/view_handler_test.go`

- [ ] **Step 1: Update result type + handler shape**

```go
package documentshttp

import (
    "context"
    "errors"
    "net/http"

    v2domain "metaldocs/internal/modules/documents/domain"
    "metaldocs/internal/modules/iam/authz"
)

type ViewResult struct {
    PDFStatus string // "pending" | "ready" | "failed"
    SignedURL string // populated only when PDFStatus == "ready"
}

type ViewService interface {
    GetViewURL(ctx context.Context, tenantID, actorID, docID string) (ViewResult, error)
}

type ViewHandler struct{ svc ViewService }

func NewViewHandler(svc ViewService) *ViewHandler { return &ViewHandler{svc: svc} }

func (h *ViewHandler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("GET /api/v2/documents/{id}/view", h.HandleView)
}

func (h *ViewHandler) HandleView(w http.ResponseWriter, r *http.Request) {
    result, err := h.svc.GetViewURL(r.Context(), tenantID(r), actorID(r), r.PathValue("id"))
    if err != nil {
        writeViewError(w, err)
        return
    }
    payload := map[string]any{"pdf_status": result.PDFStatus}
    if result.PDFStatus == "ready" && result.SignedURL != "" {
        payload["signed_url"] = result.SignedURL
        payload["pdf_url"] = result.SignedURL
    }
    writeFillInJSON(w, http.StatusOK, payload)
}

func writeViewError(w http.ResponseWriter, err error) {
    switch {
    case errors.As(err, &authz.ErrCapabilityDenied{}):
        writeFillInJSON(w, http.StatusForbidden, map[string]any{"error": "forbidden"})
    case errors.Is(err, v2domain.ErrNotFound):
        writeFillInJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
    default:
        writeFillInJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal"})
    }
}
```

`ErrPDFPending` is removed — pending no longer surfaces as 404; the handler returns 200 with `pdf_status: "pending"`.

- [ ] **Step 2: Update view handler test**

Replace the existing pending-case test with:

```go
func TestHandleView_PendingReturns200WithStatus(t *testing.T) {
    h := NewViewHandler(fakeViewService{result: ViewResult{PDFStatus: "pending"}})
    req := httptest.NewRequest(http.MethodGet, "/api/v2/documents/abc/view", nil)
    req.SetPathValue("id", "abc")
    rec := httptest.NewRecorder()
    h.HandleView(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d, want 200", rec.Code)
    }
    var got map[string]any
    _ = json.Unmarshal(rec.Body.Bytes(), &got)
    if got["pdf_status"] != "pending" {
        t.Errorf("pdf_status = %v, want pending", got["pdf_status"])
    }
    if _, ok := got["signed_url"]; ok {
        t.Errorf("signed_url should be absent when pending")
    }
}

func TestHandleView_ReadyReturnsURL(t *testing.T) {
    h := NewViewHandler(fakeViewService{result: ViewResult{PDFStatus: "ready", SignedURL: "https://s3.example/x.pdf"}})
    req := httptest.NewRequest(http.MethodGet, "/api/v2/documents/abc/view", nil)
    req.SetPathValue("id", "abc")
    rec := httptest.NewRecorder()
    h.HandleView(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d", rec.Code)
    }
    var got map[string]any
    _ = json.Unmarshal(rec.Body.Bytes(), &got)
    if got["pdf_status"] != "ready" || got["pdf_url"] == nil {
        t.Errorf("payload = %v", got)
    }
}
```

- [ ] **Step 3: Run, expect PASS**

```bash
go test -mod=mod ./internal/modules/documents/http/...
```

### Task 2.3: ViewService produces pdf_status

**Files:**
- Modify: `internal/modules/documents/application/view_service.go`
- Create or modify: `internal/modules/documents/application/view_service_test.go`

- [ ] **Step 1: Add reader interface and update service**

```go
// PDFOutboxStateReader returns the latest pdf_outbox state for a revision/document.
// Returns "" + nil when no row exists.
type PDFOutboxStateReader interface {
    ReadState(ctx context.Context, tenantID, revisionID string) (string, error)
}

type ViewService struct {
    db        *sql.DB
    presigner ViewPresigner
    outbox    PDFOutboxStateReader // optional; nil → assume pending
}

func NewViewService(db *sql.DB, presigner ViewPresigner, outbox PDFOutboxStateReader) *ViewService {
    return &ViewService{db: db, presigner: presigner, outbox: outbox}
}

func (s *ViewService) GetViewURL(ctx context.Context, tenantID, actorID, docID string) (documentshttp.ViewResult, error) {
    tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
    if err != nil {
        return documentshttp.ViewResult{}, fmt.Errorf("view: begin tx: %w", err)
    }
    defer func() { _ = tx.Rollback() }()

    ctx = authz.WithCapCache(ctx)
    if err := setAuthzGUC(ctx, tx, tenantID, actorID); err != nil {
        return documentshttp.ViewResult{}, err
    }

    var status, areaCode string
    var pdfKey sql.NullString
    err = tx.QueryRowContext(ctx, `
        SELECT status, coalesce(process_area_code_snapshot,''), final_pdf_s3_key
          FROM documents
         WHERE tenant_id=$1::uuid AND id=$2::uuid`,
        tenantID, docID,
    ).Scan(&status, &areaCode, &pdfKey)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return documentshttp.ViewResult{}, v2dom.ErrNotFound
        }
        return documentshttp.ViewResult{}, fmt.Errorf("view: load document: %w", err)
    }

    area := areaCode
    if area == "" {
        area = "tenant"
    }
    if err := authz.Require(ctx, tx, "doc.view_published", area); err != nil {
        return documentshttp.ViewResult{}, err
    }
    if _, ok := viewableStatuses[status]; !ok {
        return documentshttp.ViewResult{}, v2dom.ErrNotFound
    }

    if pdfKey.Valid && pdfKey.String != "" {
        url, err := s.presigner.PresignObjectGET(ctx, pdfKey.String)
        if err != nil {
            return documentshttp.ViewResult{}, fmt.Errorf("view: presign: %w", err)
        }
        return documentshttp.ViewResult{PDFStatus: "ready", SignedURL: url}, nil
    }

    pdfStatus := "pending"
    if s.outbox != nil {
        if state, err := s.outbox.ReadState(ctx, tenantID, docID); err == nil && state == "failed" {
            pdfStatus = "failed"
        }
    }
    return documentshttp.ViewResult{PDFStatus: pdfStatus}, nil
}
```

- [ ] **Step 2: Add sqlmock tests**

```go
func TestViewService_PDFKeySet_ReturnsReady(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()

    mock.ExpectBegin()
    mock.ExpectExec(`SET LOCAL`).WillReturnResult(sqlmock.NewResult(0,0))
    mock.ExpectQuery(`SELECT status`).WillReturnRows(
        sqlmock.NewRows([]string{"status","area","pdf_key"}).AddRow("published","ops","s3://k.pdf"))
    mock.ExpectQuery(`authz.require`).WillReturnRows(sqlmock.NewRows([]string{"ok"}).AddRow(true))

    svc := NewViewService(db, fakePresigner{url: "https://s3/example"}, nil)
    res, err := svc.GetViewURL(context.Background(), "t1", "a1", "d1")
    if err != nil { t.Fatal(err) }
    if res.PDFStatus != "ready" || res.SignedURL == "" {
        t.Errorf("res = %+v", res)
    }
}

func TestViewService_PDFKeyNull_OutboxFailed_ReturnsFailed(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()

    mock.ExpectBegin()
    mock.ExpectExec(`SET LOCAL`).WillReturnResult(sqlmock.NewResult(0,0))
    mock.ExpectQuery(`SELECT status`).WillReturnRows(
        sqlmock.NewRows([]string{"status","area","pdf_key"}).AddRow("approved","ops",nil))
    mock.ExpectQuery(`authz.require`).WillReturnRows(sqlmock.NewRows([]string{"ok"}).AddRow(true))

    svc := NewViewService(db, fakePresigner{}, fakeOutbox{state: "failed"})
    res, err := svc.GetViewURL(context.Background(), "t1", "a1", "d1")
    if err != nil { t.Fatal(err) }
    if res.PDFStatus != "failed" {
        t.Errorf("PDFStatus = %s, want failed", res.PDFStatus)
    }
}

func TestViewService_PDFKeyNull_OutboxNil_ReturnsPending(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()

    mock.ExpectBegin()
    mock.ExpectExec(`SET LOCAL`).WillReturnResult(sqlmock.NewResult(0,0))
    mock.ExpectQuery(`SELECT status`).WillReturnRows(
        sqlmock.NewRows([]string{"status","area","pdf_key"}).AddRow("approved","ops",nil))
    mock.ExpectQuery(`authz.require`).WillReturnRows(sqlmock.NewRows([]string{"ok"}).AddRow(true))

    svc := NewViewService(db, fakePresigner{}, nil)
    res, _ := svc.GetViewURL(context.Background(), "t1", "a1", "d1")
    if res.PDFStatus != "pending" {
        t.Errorf("PDFStatus = %s, want pending", res.PDFStatus)
    }
}

type fakePresigner struct{ url string }
func (f fakePresigner) PresignObjectGET(_ context.Context, _ string) (string, error) { return f.url, nil }

type fakeOutbox struct{ state string }
func (f fakeOutbox) ReadState(_ context.Context, _, _ string) (string, error) { return f.state, nil }
```

- [ ] **Step 3: Update all `NewViewService` call sites to pass new arg**

```bash
grep -rn "NewViewService(" internal/ cmd/ 2>/dev/null
```

For each call site, pass `nil` (legacy) or wire the outbox repository. In `cmd/api/wire.go` (or equivalent assembly file) wire the real `PDFOutboxRepository`. Example:

```go
// cmd/api/wire.go (or equivalent)
viewSvc := documentsapp.NewViewService(db, presigner, fanoutpkg.NewPDFOutboxRepository(db))
```

- [ ] **Step 4: Run tests, expect PASS**

```bash
go test -mod=mod ./internal/modules/documents/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/modules/documents/http/view_handler.go internal/modules/documents/http/view_handler_test.go internal/modules/documents/application/view_service.go internal/modules/documents/application/view_service_test.go internal/modules/render/fanout/pdf_outbox_repository.go cmd/
git commit -m "feat(documents): view endpoint emits pdf_status (pending/ready/failed) for async PDF readiness (E11)"
```

---

## Phase 3 — Frontend Hook + Editor Gate (parallel)

### Task 3a.1: Hook test — pending → ready

**Files:**
- Create: `frontend/apps/web/src/features/documents/v2/hooks/useDocumentPdfStatus.test.ts`

- [ ] **Step 1: Write failing test**

```ts
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';
import { useDocumentPdfStatus } from './useDocumentPdfStatus';

const server = setupServer();
beforeEach(() => server.listen());
afterEach(() => { server.resetHandlers(); server.close(); vi.useRealTimers(); });

describe('useDocumentPdfStatus', () => {
  it('polls until ready and exposes URL', async () => {
    let calls = 0;
    server.use(
      http.get('/api/v2/documents/:id/view', () => {
        calls += 1;
        if (calls < 3) return HttpResponse.json({ pdf_status: 'pending' });
        return HttpResponse.json({ pdf_status: 'ready', pdf_url: 'https://s3/x.pdf' });
      }),
    );

    vi.useFakeTimers();
    const { result } = renderHook(() => useDocumentPdfStatus('doc-1', true));
    await vi.runOnlyPendingTimersAsync();
    await vi.advanceTimersByTimeAsync(3000);
    await vi.advanceTimersByTimeAsync(3000);
    await waitFor(() => expect(result.current.status).toBe('ready'));
    expect(result.current.url).toBe('https://s3/x.pdf');
  });

  it('does not poll when disabled', async () => {
    let calls = 0;
    server.use(http.get('/api/v2/documents/:id/view', () => { calls += 1; return HttpResponse.json({ pdf_status: 'pending' }); }));
    renderHook(() => useDocumentPdfStatus('doc-1', false));
    await new Promise((r) => setTimeout(r, 50));
    expect(calls).toBe(0);
  });

  it('marks failed after 60s timeout', async () => {
    server.use(http.get('/api/v2/documents/:id/view', () => HttpResponse.json({ pdf_status: 'pending' })));
    vi.useFakeTimers();
    const { result } = renderHook(() => useDocumentPdfStatus('doc-1', true));
    await vi.advanceTimersByTimeAsync(61_000);
    await waitFor(() => expect(result.current.status).toBe('failed'));
  });

  it('retry restarts polling', async () => {
    let calls = 0;
    server.use(http.get('/api/v2/documents/:id/view', () => {
      calls += 1;
      return HttpResponse.json(calls >= 2 ? { pdf_status: 'ready', pdf_url: 'u' } : { pdf_status: 'pending' });
    }));
    vi.useFakeTimers();
    const { result } = renderHook(() => useDocumentPdfStatus('doc-1', true));
    await vi.advanceTimersByTimeAsync(61_000);
    await waitFor(() => expect(result.current.status).toBe('failed'));
    calls = 0;
    result.current.retry();
    await vi.advanceTimersByTimeAsync(3000);
    await waitFor(() => expect(result.current.status).toBe('ready'));
  });
});
```

- [ ] **Step 2: Run, expect FAIL** (hook missing)

```bash
cd frontend/apps/web && npx vitest run src/features/documents/v2/hooks/useDocumentPdfStatus.test.ts
```

### Task 3a.2: Implement hook

**Files:**
- Create: `frontend/apps/web/src/features/documents/v2/hooks/useDocumentPdfStatus.ts`

- [ ] **Step 1: Write hook**

```ts
import { useEffect, useRef, useState } from 'react';
import { apiFetch } from '../../../../lib/api/client';

export type PDFStatus = 'pending' | 'ready' | 'failed';
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
        const v = await apiFetch<ViewResponse>(`/api/v2/documents/${encodeURIComponent(documentID)}/view`);
        if (cancelled) return;
        setData({ status: v.pdf_status, url: v.pdf_url });
        if (v.pdf_status === 'ready' || v.pdf_status === 'failed') return;
        if (Date.now() - startedAt.current > TIMEOUT_MS) {
          setData({ status: 'failed' });
          return;
        }
      } catch {
        // network glitch — retry next tick
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

- [ ] **Step 2: Run tests, expect PASS**

```bash
cd frontend/apps/web && npx vitest run src/features/documents/v2/hooks/useDocumentPdfStatus.test.ts
```

- [ ] **Step 3: Commit**

```bash
git add frontend/apps/web/src/features/documents/v2/hooks/useDocumentPdfStatus.ts frontend/apps/web/src/features/documents/v2/hooks/useDocumentPdfStatus.test.ts
git commit -m "feat(documents): useDocumentPdfStatus hook polls view endpoint until pdf ready (E11)"
```

### Task 3b.1: Editor gate test

**Files:**
- Create: `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.test.tsx`

- [ ] **Step 1: Write failing tests**

```tsx
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';
import { DocumentEditorPage } from './DocumentEditorPage';

vi.mock('@metaldocs/editor-ui', () => ({
  MetalDocsEditor: (props: any) => <div data-testid="editor" data-mode={props.mode} />,
}));

vi.mock('./hooks/useDocumentSession', () => ({
  useDocumentSession: () => ({
    state: { phase: 'writer', sessionID: 's1', lastAckRevisionID: 'r1' },
    setLastAck: vi.fn(),
    release: vi.fn(),
  }),
}));

const server = setupServer();
beforeEach(() => server.listen());
afterEach(() => { server.resetHandlers(); server.close(); });

describe('DocumentEditorPage E1 gate', () => {
  it('renders document-edit when status=draft', async () => {
    server.use(
      http.get('/api/v2/documents/:id', () => HttpResponse.json({
        Status: 'draft', CurrentRevisionID: 'r1', Name: 'X', Code: 'C-001', RevisionVersion: 1,
      })),
      http.get('/api/v2/documents/:id/revisions/:rid/signed-url', () => HttpResponse.json({ url: 'blob:x' })),
      http.get('blob:x', () => new HttpResponse(new ArrayBuffer(8))),
    );
    render(<DocumentEditorPage documentID="d1" onDone={() => {}} />);
    await waitFor(() => expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'));
  });

  it('renders readonly when status=under_review', async () => {
    server.use(
      http.get('/api/v2/documents/:id', () => HttpResponse.json({
        Status: 'under_review', CurrentRevisionID: 'r1', Name: 'X', Code: 'C-001', RevisionVersion: 1,
      })),
      http.get('/api/v2/documents/:id/revisions/:rid/signed-url', () => HttpResponse.json({ url: 'blob:x' })),
      http.get('blob:x', () => new HttpResponse(new ArrayBuffer(8))),
    );
    render(<DocumentEditorPage documentID="d1" onDone={() => {}} />);
    await waitFor(() => expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('readonly'));
  });

  it('refetches doc on window focus', async () => {
    let status = 'draft';
    server.use(
      http.get('/api/v2/documents/:id', () => HttpResponse.json({
        Status: status, CurrentRevisionID: 'r1', Name: 'X', Code: 'C-001', RevisionVersion: 1,
      })),
      http.get('/api/v2/documents/:id/revisions/:rid/signed-url', () => HttpResponse.json({ url: 'blob:x' })),
      http.get('blob:x', () => new HttpResponse(new ArrayBuffer(8))),
    );
    render(<DocumentEditorPage documentID="d1" onDone={() => {}} />);
    await waitFor(() => expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'));
    status = 'under_review';
    fireEvent(window, new FocusEvent('focus'));
    await waitFor(() => expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('readonly'));
  });
});
```

- [ ] **Step 2: Run, expect FAIL** (mode still phase-only, no focus refetch)

```bash
cd frontend/apps/web && npx vitest run src/features/documents/v2/DocumentEditorPage.test.tsx
```

### Task 3b.2: Apply editor gate + focus refetch

**Files:**
- Modify: `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx`

- [ ] **Step 1: Add `isEditable` derivation, focus refetch effect, update mode + finalize disabled**

Replace at line ~151-167 area (after `docStatus` is computed):

```tsx
const docStatus = doc?.Status ?? doc?.status ?? '';
const isEditable = session.state.phase === 'writer' && docStatus === 'draft';
```

Replace the existing finalize button:

```tsx
<button
  type="button"
  className={styles.editorSubmitBtn}
  onClick={() => void handleFinalize()}
  disabled={!isEditable}
>
  Finalizar
</button>
```

Replace editor mode (line ~233):

```tsx
<MetalDocsEditor
  ref={editorRef}
  mode={isEditable ? 'document-edit' : 'readonly'}
  ...
/>
```

Add focus refetch effect (near other useEffects, e.g. after the existing load effect at line 57):

```tsx
useEffect(() => {
  const onFocus = () => {
    void getDocument(documentID).then((d) => setDoc(d)).catch(() => {});
  };
  window.addEventListener('focus', onFocus);
  return () => window.removeEventListener('focus', onFocus);
}, [documentID]);
```

- [ ] **Step 2: Run tests, expect PASS**

```bash
cd frontend/apps/web && npx vitest run src/features/documents/v2/DocumentEditorPage.test.tsx
```

- [ ] **Step 3: Commit**

```bash
git add frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx frontend/apps/web/src/features/documents/v2/DocumentEditorPage.test.tsx
git commit -m "fix(editor): lock editor mode on non-draft status + refetch on focus (E1)"
```

---

## Phase 4 — Frontend E10 Wiring

### Task 4.1: Update `ActiveDocumentResponse` type and client

**Files:**
- Modify: `frontend/apps/web/src/features/registry/api.ts`

- [ ] **Step 1: Replace type and fetcher**

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
  try {
    return await apiFetch<ActiveDocumentResponse>(
      `${BASE}/${encodeURIComponent(controlledDocumentId)}/active-document`,
    );
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  }
}
```

Update imports at top:

```ts
import { apiFetch, ApiError } from '../../lib/api/client';
```

Remove the old fetch-based implementation. Keep `BASE` constant.

- [ ] **Step 2: Find consumers and update field references**

```bash
grep -rn "ActiveDocumentInstance\|approvalState\|publishedDocumentId" frontend/apps/web/src/features/registry/ frontend/apps/web/src/features/approval/
```

Rename `ActiveDocumentInstance` → `ActiveDocumentResponse` everywhere; treat `documentId` as optional.

- [ ] **Step 3: Run typecheck**

```bash
cd frontend/apps/web && npx tsc --noEmit
```

### Task 4.2: RegistryDetailPage renders published-only state

**Files:**
- Modify: `frontend/apps/web/src/features/registry/RegistryDetailPage.tsx`
- Create: `frontend/apps/web/src/features/registry/RegistryDetailPage.test.tsx`

- [ ] **Step 1: Write failing test**

```tsx
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';
import { RegistryDetailPage } from './RegistryDetailPage';

const server = setupServer();
beforeEach(() => server.listen());
afterEach(() => { server.resetHandlers(); server.close(); });

describe('RegistryDetailPage E10', () => {
  it('shows published banner when only published exists', async () => {
    server.use(
      http.get('/api/v2/controlled-documents/:id', () => HttpResponse.json({
        id: 'cd-1', code: 'SOP-001', title: 'T', status: 'active', profileCode: 'sop', processAreaCode: 'ops',
        ownerUserId: 'u', createdAt: '', updatedAt: '',
      })),
      http.get('/api/v2/controlled-documents/:id/active-document', () => HttpResponse.json({
        published_document_id: 'pub-1',
      })),
    );
    render(<RegistryDetailPage id="cd-1" onBack={() => {}} />);
    await screen.findByText(/Revisão publicada/i);
    expect(screen.getByRole('button', { name: /Nova Revisão/i })).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run, expect FAIL**

```bash
cd frontend/apps/web && npx vitest run src/features/registry/RegistryDetailPage.test.tsx
```

- [ ] **Step 3: Update render logic**

Replace the conditional block at lines ~143-209 with:

```tsx
const hasActive = !!instance?.documentId;
const hasPublished = !!instance?.publishedDocumentId;

{hasActive && (
  <div style={{ marginTop: 32 }}>
    <RegistryDetailPanel
      documentId={instance!.documentId!}
      approvalState={instance!.approvalState ?? 'draft'}
      contentHash={instance!.contentHash ?? ''}
      revisionVersion={instance!.revisionVersion ?? 0}
      publishedDocumentId={instance!.publishedDocumentId}
      lockedByInstanceId={instance!.approvalInstanceId}
    />
  </div>
)}

{hasPublished && (
  <div style={{ marginTop: 32, padding: 16, border: '1px solid #2a7a2a', borderRadius: 4, background: '#f4faf4' }}>
    <p style={{ margin: '0 0 8px', fontWeight: 600, color: '#2a7a2a' }}>Revisão publicada</p>
    <PublishedDownloadCell documentId={instance!.publishedDocumentId!} />
  </div>
)}

{!hasActive && doc.status === 'active' && (
  <div style={{ marginTop: 32, padding: 16, border: '1px dashed #ccc', borderRadius: 4, color: '#888' }}>
    {/* existing Nova Revisão form unchanged */}
  </div>
)}
```

`PublishedDownloadCell` is a small inline component (next task).

- [ ] **Step 4: Commit (test still fails — depends on PublishedDownloadCell)**

Skip commit until task 4.3.

### Task 4.3: PublishedDownloadCell component using PDF status hook

**Files:**
- Create: `frontend/apps/web/src/features/registry/PublishedDownloadCell.tsx`

- [ ] **Step 1: Write component**

```tsx
import { useDocumentPdfStatus } from '../documents/v2/hooks/useDocumentPdfStatus';

export function PublishedDownloadCell({ documentId }: { documentId: string }) {
  const pdf = useDocumentPdfStatus(documentId, true);
  if (pdf.status === 'ready' && pdf.url) {
    return <a href={pdf.url} download style={{ color: '#2a7a2a', fontWeight: 600 }}>Baixar PDF</a>;
  }
  if (pdf.status === 'failed') {
    return (
      <span style={{ color: '#c00' }}>
        Falha ao gerar PDF.{' '}
        <button type="button" onClick={pdf.retry} style={{ marginLeft: 4 }}>Tentar novamente</button>
      </span>
    );
  }
  return <span style={{ color: '#888' }}>Gerando PDF…</span>;
}
```

- [ ] **Step 2: Run RegistryDetailPage test, expect PASS**

```bash
cd frontend/apps/web && npx vitest run src/features/registry/RegistryDetailPage.test.tsx
```

- [ ] **Step 3: Commit**

```bash
git add frontend/apps/web/src/features/registry/api.ts frontend/apps/web/src/features/registry/RegistryDetailPage.tsx frontend/apps/web/src/features/registry/RegistryDetailPage.test.tsx frontend/apps/web/src/features/registry/PublishedDownloadCell.tsx
git commit -m "fix(registry): render published banner with PDF download when only published revision exists (E10)"
```

---

## Phase 5 — Frontend E11 Wiring (DocumentEditorPage)

### Task 5.1: Editor PDF cell test

**Files:**
- Modify: `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.test.tsx`

- [ ] **Step 1: Append test**

```tsx
it('shows PDF download link when status published and pdf ready', async () => {
  server.use(
    http.get('/api/v2/documents/:id', () => HttpResponse.json({
      Status: 'published', CurrentRevisionID: 'r1', Name: 'X', Code: 'C', RevisionVersion: 1,
    })),
    http.get('/api/v2/documents/:id/revisions/:rid/signed-url', () => HttpResponse.json({ url: 'blob:x' })),
    http.get('blob:x', () => new HttpResponse(new ArrayBuffer(8))),
    http.get('/api/v2/documents/:id/view', () => HttpResponse.json({ pdf_status: 'ready', pdf_url: 'https://s3/p.pdf' })),
  );
  render(<DocumentEditorPage documentID="d1" onDone={() => {}} />);
  const link = await screen.findByRole('link', { name: /Baixar PDF/i });
  expect(link.getAttribute('href')).toBe('https://s3/p.pdf');
});
```

- [ ] **Step 2: Run, expect FAIL**

### Task 5.2: Wire `useDocumentPdfStatus` into editor

**Files:**
- Modify: `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx`

- [ ] **Step 1: Import hook + cell**

```tsx
import { useDocumentPdfStatus } from './hooks/useDocumentPdfStatus';
import { PDFCell } from './PDFCell';
```

- [ ] **Step 2: Add hook + cell in overlayRight**

After `isEditable` derivation:

```tsx
const pdf = useDocumentPdfStatus(documentID, docStatus !== 'draft');
```

In `overlayRight` (next to ExportMenuButton):

```tsx
{docStatus !== 'draft' && <PDFCell status={pdf.status} url={pdf.url} onRetry={pdf.retry} />}
```

- [ ] **Step 3: Create `PDFCell` component**

`frontend/apps/web/src/features/documents/v2/PDFCell.tsx`:

```tsx
import type { PDFStatus } from './hooks/useDocumentPdfStatus';

type Props = { status: PDFStatus; url?: string; onRetry: () => void };

export function PDFCell({ status, url, onRetry }: Props) {
  if (status === 'ready' && url) {
    return <a href={url} download style={{ color: '#2a7a2a', fontWeight: 600 }}>Baixar PDF</a>;
  }
  if (status === 'failed') {
    return (
      <span style={{ color: '#c00' }}>
        Falha ao gerar PDF.
        <button type="button" onClick={onRetry} style={{ marginLeft: 4 }}>Tentar novamente</button>
      </span>
    );
  }
  return <span style={{ color: '#888' }}>Gerando PDF…</span>;
}
```

Replace the inline `PublishedDownloadCell` body to delegate to `PDFCell` (DRY):

```tsx
import { PDFCell } from '../documents/v2/PDFCell';

export function PublishedDownloadCell({ documentId }: { documentId: string }) {
  const pdf = useDocumentPdfStatus(documentId, true);
  return <PDFCell status={pdf.status} url={pdf.url} onRetry={pdf.retry} />;
}
```

- [ ] **Step 4: Run tests, expect PASS**

```bash
cd frontend/apps/web && npx vitest run src/features/documents/v2/DocumentEditorPage.test.tsx
```

- [ ] **Step 5: Commit**

```bash
git add frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx frontend/apps/web/src/features/documents/v2/PDFCell.tsx frontend/apps/web/src/features/registry/PublishedDownloadCell.tsx frontend/apps/web/src/features/documents/v2/DocumentEditorPage.test.tsx
git commit -m "feat(editor): show PDF download link post-finalize via polling hook (E11)"
```

---

## Phase 6 — Verify

### Task 6.1: Full test sweep

- [ ] **Step 1: Backend**

```bash
go test -mod=mod ./...
```

Expected: PASS.

- [ ] **Step 2: Frontend**

```bash
cd frontend/apps/web && npx vitest run && npx tsc --noEmit && npm run lint
```

Expected: PASS, zero new lint warnings.

### Task 6.2: Smoke flows

Use `.\scripts\start-api.ps1` and dev frontend.

- [ ] **Smoke E1:** open draft doc, click Finalizar, observe editor flips readonly without reload.
- [ ] **Smoke E10:** in DB, set a controlled doc to have only `status='published'` revisions. Open registry detail, observe published banner + Nova Revisão CTA.
- [ ] **Smoke E11:** finalize a doc. Observe `Gerando PDF…` for ~10s, then download link appears.

### Task 6.3: Codex independent audit

Dispatch codex:codex-rescue subagent with prompt:

> Independent audit of Group E sub-plan 2 implementation against `docs/superpowers/specs/2026-05-04-group-e-editor-lifecycle-design.md`. For each of E1, E10, E11 produce PASS/FAIL with file:line evidence. Verify (a) no raw `fetch` introduced in changed frontend files (must use `apiFetch`); (b) `ErrPDFPending` sentinel removed from view handler; (c) FULL OUTER JOIN preserves single-row contract; (d) hook polling cancels on unmount (no leaked timers).

Required: 3/3 PASS before Phase 7.

---

## Phase 7 — Closure

### Task 7.1: Audit doc closure

**Files:**
- Modify: `wiki/bugs/audit-2026-05-03.md` (lines 271, 280, 281)

- [ ] **Step 1: Mark closed**

Change status column to `fixed` and Fix-commit column to the relevant SHAs for E1, E10, E11.

```bash
git add wiki/bugs/audit-2026-05-03.md
git commit -m "docs(audit): close E1/E10/E11 with fix SHAs"
```

### Task 7.2: Wiki-curator dispatch

- [ ] **Step 1: Dispatch wiki-curator agent with prompt**

> Update wiki for Group E sub-plan 2. Create or refresh `wiki/concepts/document-lifecycle.md` to document: (a) editor mode is gated on `phase + docStatus`; (b) active-document endpoint contract returns `{document_id?, published_document_id?}`; (c) PDF readiness contract via `pdf_status` field with polling expectation. Refresh `Last verified` stamps on touched module docs. Update `wiki/README.md` index.

### Task 7.3: Branch finishing

- [ ] **Step 1: Use superpowers:finishing-a-development-branch**

Default to option 2 (push + PR) unless user requests otherwise.

---

## Acceptance Criteria

- [ ] E1: post-finalize editor renders readonly without reload
- [ ] E1: window focus refetches doc status
- [ ] E10: active-document endpoint returns 200 with `published_document_id` when only published exists
- [ ] E10: 404 only when zero rows of either kind
- [ ] E10: registry detail shows published banner + Nova Revisão when only published exists
- [ ] E10: registry detail shows both banners when active and published coexist
- [ ] E11: view endpoint returns `pdf_status` ∈ {pending, ready, failed}
- [ ] E11: editor + registry panel poll until ready, show download link
- [ ] E11: 60s timeout shows retry button; retry re-polls
- [ ] All vitest pass
- [ ] `go test -mod=mod ./...` passes
- [ ] No new lint warnings
- [ ] Codex audit returns 3/3 PASS
- [ ] Audit doc updated, E1/E10/E11 closed with commit SHAs
- [ ] Wiki refreshed by wiki-curator
