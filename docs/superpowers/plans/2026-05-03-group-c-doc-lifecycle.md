# Group C — Document Lifecycle Correctness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix C1–C6 lifecycle bugs from `wiki/bugs/audit-2026-05-03.md` so documents are born complete (atomic snapshot INSERT), revisions allocate without races, archive uses timestamp not status, and `finalized_at` derives from event log.

**Architecture:** Two-pillar invariant — (1) documents born complete (snapshot resolved before INSERT, written in same tx as document row + placeholder seed), (2) lifecycle truth lives in `document_state_history` (drop denormalized `finalized_at`, soft-archive via `archived_at`). Localized to `internal/modules/documents/{application,repository}` plus one migration and one ADR.

**Tech Stack:** Go 1.22, PostgreSQL 16, `database/sql` + `pgx/v5`, sqlmock for unit tests, `-tags=integration` for live-DB tests.

**Spec:** `docs/superpowers/specs/2026-05-03-group-c-doc-lifecycle-design.md`

**Model selection:**
- Codex (complex multi-file work): C2/C4 atomic INSERT refactor, C5 advisory lock, caller-fan-out cleanup
- Sonnet (mechanical / single-file): C1 `MarkArchived` repo, list-query predicate sweep, `Service.Archive` signature change
- Haiku (one-shot scripts/diagnostics): C3 verification SQL
- Opus (phase review only — never coding): post-phase audits

**Codex parallelism:**
- Phase 1: C3 ‖ C6 migration ‖ C1 repo (zero file overlap)
- Phase 4: caller cleanup ‖ list-query sweep
- All other phases sequential due to file-conflict on `repository.go` / `service.go`

**Wiki maintenance:** dedicated `wiki-curator` subagent (already in project) dispatched in Phase 6 to refresh stamps, ADR creation, and audit-doc closure entries.

**Caveman prompts:** every subagent prompt drops articles/filler. `/simplify` directive applied to all generated code (least lines that solve the problem).

---

## File Structure

| File | Responsibility | Action |
|---|---|---|
| `internal/modules/documents/application/service.go` | Document orchestration | Modify Create flow (lines 308–357), Archive (line 594) |
| `internal/modules/documents/application/snapshot_service.go` | Template→document snapshot resolver+writer | Split: add `ResolveTemplate` (read-only), keep `SnapshotFromTemplate` for legacy callers but mark deprecated |
| `internal/modules/documents/repository/repository.go` | DB writes for documents | New `MarkArchived`; `CreateDocument` accepts `domain.TemplateSnapshot` + required placeholders, writes everything in one tx; advisory lock; remove `finalized_at` write |
| `internal/modules/documents/domain/model.go` | Document domain | Remove `FinalizedAt` field, update `TemplateSnapshot` payload usage |
| `internal/modules/documents/delivery/http/handler.go` | HTTP handler | `Archive(...)` signature drops `fromFinalized bool` parameter (lines 49, 375, 386) |
| `internal/modules/documents/delivery/http/handler_test.go` | Handler test fakes | Update `fakeSvc.Archive` signature (line 131) |
| `migrations/0171_drop_finalized_at.sql` | Schema migration | New: drop column, create view |
| `scripts/verify-triggers.sql` | Diagnostic | New: confirm `enforce_snapshot_on_submit_trg` present |
| `wiki/decisions/0008-soft-archive-via-timestamp.md` | ADR | New: rationale for soft-archive |
| `wiki/bugs/audit-2026-05-03.md` | Audit log | Close C1–C6 with commit SHAs |
| `tests/integration/documents/lifecycle_test.go` | Integration tests | New: cross-cutting C1+C5+C6 verification |

---

## Phase 0 — Worktree + Spec Validation

### Task 0.1: Create worktree

**Files:** none (git op)

- [ ] **Step 1:** Create worktree from main

```bash
git worktree add ../MetalDocs-group-c -b group-c-lifecycle main
cd ../MetalDocs-group-c
```

- [ ] **Step 2:** Verify clean state

Run: `git status`
Expected: `nothing to commit, working tree clean` on `group-c-lifecycle`.

### Task 0.2: Codex spec validation

**Files:** none (review)

- [ ] **Step 1:** Dispatch codex agent to validate spec

Prompt (caveman):
```
Validate spec docs/superpowers/specs/2026-05-03-group-c-doc-lifecycle-design.md
against code reality. For each bug C1-C6 confirm file paths + line refs match
current main. Report PASS/FAIL per bug with evidence. No fix attempts.
```

- [ ] **Step 2:** Block if any FAIL

If codex reports drift, halt and update spec before proceeding. Otherwise proceed.

### Task 0.3: Snapshot read-only audit

**Files:** none (read)

- [ ] **Step 1:** Confirm `SnapshotService` writes only to documents + placeholder_values

```bash
grep -n "INSERT\|UPDATE\|DELETE" internal/modules/documents/application/snapshot_service.go
grep -rn "WriteSnapshot\|SeedDefaults" internal/modules/documents/
```

Expected writes: `documents.*_snapshot` columns (via `SnapshotWriter`) and `placeholder_values` rows (via `PlaceholderValueSeeder`). No other tables.

- [ ] **Step 2:** Document finding in plan log

If any other table writes appear, halt and revise C2/C4 approach (use tx-propagation).

---

## Phase 1 — Parallel Tasks (C3, C6 migration, C1 repo)

Dispatch three subagents in parallel. No file overlap between tasks.

### Task 1.1: C3 — verify trigger present (Haiku)

**Files:**
- Create: `scripts/verify-triggers.sql`
- Modify: `wiki/bugs/audit-2026-05-03.md` (C3 entry)

- [ ] **Step 1: Write diagnostic**

```sql
-- scripts/verify-triggers.sql
-- Confirms enforce_snapshot_on_submit_trg exists on metaldocs.documents.
-- Run: psql $DATABASE_URL -f scripts/verify-triggers.sql
SELECT
    tgname,
    tgrelid::regclass AS target,
    pg_get_triggerdef(oid) AS definition
FROM pg_trigger
WHERE tgname = 'enforce_snapshot_on_submit_trg';
```

- [ ] **Step 2: Run against dev DB**

```powershell
.\scripts\start-api.ps1  # ensures DB running
psql $env:DATABASE_URL -f scripts/verify-triggers.sql
```

Expected: 1 row returned. `target = metaldocs.documents`. Definition includes `BEFORE INSERT OR UPDATE`.

- [ ] **Step 3: Update audit doc**

In `wiki/bugs/audit-2026-05-03.md`, change C3 entry to:
```markdown
- [x] **C3** ~~enforce_snapshot_on_submit_trg missing~~ — **FALSE POSITIVE**.
  Trigger present in migration 0152 lines 46-49. Verified via
  `scripts/verify-triggers.sql` on 2026-05-03. Closed without code change.
```

- [ ] **Step 4: Commit**

```bash
git add scripts/verify-triggers.sql wiki/bugs/audit-2026-05-03.md
git commit -m "fix(docs): C3 enforce_snapshot_on_submit_trg verified present (false positive)"
```

### Task 1.2: C6 — drop finalized_at column + view (Codex)

**Files:**
- Create: `migrations/0171_drop_finalized_at.sql`
- Test: `tests/integration/documents/finalized_view_test.go`

- [ ] **Step 1: Write integration test (failing)**

```go
// tests/integration/documents/finalized_view_test.go
//go:build integration

package documents_integration

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"metaldocs/internal/testdb"
)

func TestVDocumentFinalized_ReturnsApprovalChangedAt(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()

	tenantID, docID := testdb.InsertDraftDocument(t, db, nil, testdb.DevTenantID)
	approvedAt := time.Now().UTC().Truncate(time.Microsecond)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO metaldocs.document_state_history
			(document_id, from_status, to_status, changed_by, changed_at)
		VALUES ($1, 'under_review', 'approved', 'tester', $2)`,
		docID, approvedAt); err != nil {
		t.Fatalf("seed history: %v", err)
	}

	var got sql.NullTime
	if err := db.QueryRowContext(ctx,
		`SELECT finalized_at FROM metaldocs.v_document_finalized WHERE document_id=$1`,
		docID,
	).Scan(&got); err != nil {
		t.Fatalf("query view: %v", err)
	}
	if !got.Valid || !got.Time.Equal(approvedAt) {
		t.Fatalf("finalized_at = %v, want %v", got, approvedAt)
	}
	_ = tenantID
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -tags=integration ./tests/integration/documents/ -run TestVDocumentFinalized_ReturnsApprovalChangedAt -v`
Expected: FAIL — `relation "metaldocs.v_document_finalized" does not exist`.

- [ ] **Step 3: Write migration**

```sql
-- migrations/0171_drop_finalized_at.sql
-- Group C / C6: drop denormalized finalized_at, derive from state history.
BEGIN;

CREATE OR REPLACE VIEW metaldocs.v_document_finalized AS
SELECT
    d.id AS document_id,
    (SELECT h.changed_at
       FROM metaldocs.document_state_history h
      WHERE h.document_id = d.id
        AND h.to_status = 'approved'
      ORDER BY h.changed_at ASC
      LIMIT 1) AS finalized_at
FROM metaldocs.documents d;

ALTER TABLE metaldocs.documents DROP COLUMN IF EXISTS finalized_at;

COMMIT;
```

- [ ] **Step 4: Apply migration**

```powershell
.\scripts\start-api.ps1 -Build  # picks up new migration
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test -tags=integration ./tests/integration/documents/ -run TestVDocumentFinalized_ReturnsApprovalChangedAt -v`
Expected: PASS.

> **Note:** Subsequent tasks (Task 4.1, 4.2) remove every Go reference to `documents.finalized_at`. Until those land, `go build` may fail. Run Task 1.2 + Task 4.1 + Task 4.2 in the same Phase 1 batch but commit Task 1.2 last to keep main buildable. If executed via subagent-driven-development, controller orders commits accordingly.

- [ ] **Step 6: Commit migration only (build cleanup in Phase 4)**

```bash
git add migrations/0171_drop_finalized_at.sql tests/integration/documents/finalized_view_test.go
git commit -m "feat(migration): C6 drop documents.finalized_at, add v_document_finalized view"
```

### Task 1.3: C1 — MarkArchived repo + service (Sonnet)

**Files:**
- Modify: `internal/modules/documents/repository/repository.go` (add `MarkArchived`, `Unarchive` methods after `UpdateDocumentStatus`)
- Test: `internal/modules/documents/repository/repository_archive_test.go` (new)

- [ ] **Step 1: Write failing unit test**

```go
// internal/modules/documents/repository/repository_archive_test.go
package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMarkArchived_StampsTimestampWithoutStatusChange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := New(db)

	mock.ExpectExec(`UPDATE metaldocs\.documents`).
		WithArgs("tenant-1", "doc-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := r.MarkArchived(context.Background(), "tenant-1", "doc-1", "user-1"); err != nil {
		t.Fatalf("MarkArchived: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUnarchive_ClearsTimestamp(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	r := New(db)

	mock.ExpectExec(`UPDATE metaldocs\.documents`).
		WithArgs("tenant-1", "doc-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := r.Unarchive(context.Background(), "tenant-1", "doc-1", "user-1"); err != nil {
		t.Fatalf("Unarchive: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}
```

- [ ] **Step 2: Run to verify FAIL**

Run: `go test ./internal/modules/documents/repository/ -run TestMarkArchived -v`
Expected: FAIL — `r.MarkArchived undefined`.

- [ ] **Step 3: Implement MarkArchived + Unarchive**

Append to `internal/modules/documents/repository/repository.go` (after line 213, end of `UpdateDocumentStatus`):

```go
// MarkArchived sets archived_at on a document without changing status.
// Idempotent: succeeds with rows-affected=0 if already archived.
// Soft-archive semantics — terminal status (published/superseded/obsolete)
// is preserved so audit history remains accurate.
func (r *Repository) MarkArchived(ctx context.Context, tenantID, docID, actorID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE metaldocs.documents
		   SET archived_at = now(), updated_at = now()
		 WHERE tenant_id = $1 AND id = $2 AND archived_at IS NULL`,
		tenantID, docID)
	_ = actorID // reserved for future audit column; audit currently via Service.audit
	return err
}

// Unarchive clears archived_at, restoring the document to active queries.
func (r *Repository) Unarchive(ctx context.Context, tenantID, docID, actorID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE metaldocs.documents
		   SET archived_at = NULL, updated_at = now()
		 WHERE tenant_id = $1 AND id = $2 AND archived_at IS NOT NULL`,
		tenantID, docID)
	_ = actorID
	return err
}
```

- [ ] **Step 4: Run to verify PASS**

Run: `go test ./internal/modules/documents/repository/ -run TestMarkArchived -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/modules/documents/repository/repository.go internal/modules/documents/repository/repository_archive_test.go
git commit -m "feat(documents): C1 add MarkArchived/Unarchive (timestamp-only soft archive)"
```

---

## Phase 1 Review (Opus)

- [ ] **Dispatch Opus reviewer**

Prompt:
```
Review group-c-lifecycle commits since Phase 0. Confirm: (a) C3 audit
update accurate, (b) migration 0171 idempotent and matches spec view
shape, (c) MarkArchived semantics correct (no status change, archived_at
nullable). Flag any issues. PASS/FAIL.
```

Block on FAIL.

---

## Phase 2 — C5 Advisory Lock (Codex, sequential)

### Task 2.1: Add advisory lock to CreateDocument

**Files:**
- Modify: `internal/modules/documents/repository/repository.go:35-87` (CreateDocument)
- Test: `tests/integration/documents/concurrent_revision_test.go` (new)

- [ ] **Step 1: Write failing concurrency test**

```go
// tests/integration/documents/concurrent_revision_test.go
//go:build integration

package documents_integration

import (
	"context"
	"sync"
	"testing"

	"metaldocs/internal/testdb"
)

func TestCreateDocument_ConcurrentSubmitsSerialised(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()

	cdID := testdb.InsertControlledDocument(t, db, testdb.DevTenantID)

	const N = 10
	var wg sync.WaitGroup
	errs := make(chan error, N)
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			_, _, _, err := testdb.CreateDocumentForCD(ctx, db, testdb.DevTenantID, cdID)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent create failed: %v", err)
		}
	}

	// Assert revisions 1..N exist exactly once each
	rows, err := db.QueryContext(ctx,
		`SELECT revision_number FROM metaldocs.documents
		  WHERE controlled_document_id=$1 ORDER BY revision_number`, cdID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	got := []int{}
	for rows.Next() {
		var n int
		_ = rows.Scan(&n)
		got = append(got, n)
	}
	if len(got) != N {
		t.Fatalf("expected %d rows, got %d (%v)", N, len(got), got)
	}
	for i, n := range got {
		if n != i+1 {
			t.Fatalf("revision_number[%d] = %d, want %d", i, n, i+1)
		}
	}
}
```

- [ ] **Step 2: Run to verify FAIL**

Run: `go test -tags=integration ./tests/integration/documents/ -run TestCreateDocument_ConcurrentSubmitsSerialised -v`
Expected: FAIL with `pq: duplicate key value violates unique constraint "ux_documents_v2_cd_revision"` on at least one goroutine.

- [ ] **Step 3: Add advisory lock inside CreateDocument tx**

In `internal/modules/documents/repository/repository.go`, modify `CreateDocument` to insert advisory lock AFTER `BeginTx` and BEFORE the document INSERT:

```go
func (r *Repository) CreateDocument(ctx context.Context, d *domain.Document, initialContentHash string) (docID, revID, sessionID string, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return "", "", "", err
	}
	defer tx.Rollback()

	// Serialise revision_number allocation per (tenant, controlled_document).
	// Auto-released on COMMIT/ROLLBACK. hashtextextended hashes any text.
	if d.ControlledDocumentID != nil && *d.ControlledDocumentID != "" {
		lockKey := d.TenantID + ":" + *d.ControlledDocumentID
		if _, err := tx.ExecContext(ctx,
			`SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, lockKey,
		); err != nil {
			return "", "", "", fmt.Errorf("acquire revision lock: %w", err)
		}
	}

	// (existing INSERT documents ... unchanged)
	if err := tx.QueryRowContext(ctx,
```

(rest of method unchanged)

- [ ] **Step 4: Run to verify PASS**

Run: `go test -tags=integration ./tests/integration/documents/ -run TestCreateDocument_ConcurrentSubmitsSerialised -v`
Expected: PASS. All 10 rows present with monotonic revision_number 1..10.

- [ ] **Step 5: Run full unit suite (regression check)**

Run: `go test ./internal/modules/documents/repository/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/modules/documents/repository/repository.go tests/integration/documents/concurrent_revision_test.go
git commit -m "fix(documents): C5 advisory lock serialises concurrent revision allocation"
```

---

## Phase 2 Review (Opus)

- [ ] **Dispatch Opus reviewer**

Prompt:
```
Review C5 advisory-lock commit. Confirm: (a) lock key includes tenant
prefix, (b) lock acquired before MAX(revision_number) read, (c) no lock
leak (auto-release at tx end), (d) integration test asserts monotonic
sequence not just absence of error. PASS/FAIL.
```

---

## Phase 3 — C2/C4 Atomic Snapshot (Codex, sequential)

### Task 3.1: Add ResolveTemplate to SnapshotService

**Files:**
- Modify: `internal/modules/documents/application/snapshot_service.go`
- Test: `internal/modules/documents/application/snapshot_resolver_test.go` (new)

- [ ] **Step 1: Write failing test**

```go
// internal/modules/documents/application/snapshot_resolver_test.go
package application

import (
	"context"
	"errors"
	"testing"

	"metaldocs/internal/modules/documents/domain"
	templatesdomain "metaldocs/internal/modules/templates_v2/domain"
)

type stubReader struct {
	snap domain.TemplateSnapshot
	err  error
}

func (s stubReader) LoadForSnapshot(ctx context.Context, tenantID, templateID string) (domain.TemplateSnapshot, error) {
	return s.snap, s.err
}

func TestResolveTemplate_ReturnsSnapshotAndPlaceholders(t *testing.T) {
	want := domain.TemplateSnapshot{
		PlaceholderSchemaJSON: []byte(`[{"key":"author","required":true}]`),
	}
	svc := NewSnapshotService(stubReader{snap: want}, nil)

	got, phs, err := svc.ResolveTemplate(context.Background(), "tenant-1", "tmpl-1")
	if err != nil {
		t.Fatalf("ResolveTemplate: %v", err)
	}
	if string(got.PlaceholderSchemaJSON) != string(want.PlaceholderSchemaJSON) {
		t.Fatalf("snapshot mismatch")
	}
	if len(phs) != 1 || phs[0].Key != "author" {
		t.Fatalf("placeholders = %+v", phs)
	}
	_ = templatesdomain.Placeholder{}
}

func TestResolveTemplate_ReaderError(t *testing.T) {
	svc := NewSnapshotService(stubReader{err: errors.New("boom")}, nil)
	if _, _, err := svc.ResolveTemplate(context.Background(), "t", "x"); err == nil {
		t.Fatal("expected error")
	}
}
```

- [ ] **Step 2: Run to verify FAIL**

Run: `go test ./internal/modules/documents/application/ -run TestResolveTemplate -v`
Expected: FAIL — `svc.ResolveTemplate undefined`.

- [ ] **Step 3: Add ResolveTemplate (pure read)**

Append to `internal/modules/documents/application/snapshot_service.go`:

```go
// ResolveTemplate reads the template snapshot and required-placeholder list
// without writing to the DB. Used by the document Create flow so the snapshot
// can be written atomically with the documents row.
func (s *SnapshotService) ResolveTemplate(ctx context.Context, tenantID, templateID string) (domain.TemplateSnapshot, []templatesdomain.Placeholder, error) {
	snap, err := s.templates.LoadForSnapshot(ctx, tenantID, templateID)
	if err != nil {
		return domain.TemplateSnapshot{}, nil, err
	}
	phs, err := parseRequiredPlaceholders(snap.PlaceholderSchemaJSON)
	if err != nil {
		return domain.TemplateSnapshot{}, nil, fmt.Errorf("parse placeholder schema: %w", err)
	}
	return snap, phs, nil
}
```

- [ ] **Step 4: Run to verify PASS**

Run: `go test ./internal/modules/documents/application/ -run TestResolveTemplate -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/modules/documents/application/snapshot_service.go internal/modules/documents/application/snapshot_resolver_test.go
git commit -m "feat(documents): C2 add SnapshotService.ResolveTemplate (pure read for atomic create)"
```

### Task 3.2: Extend CreateDocument repo to write snapshot + placeholder seed in tx

**Files:**
- Modify: `internal/modules/documents/repository/repository.go:35-87` (CreateDocument signature + body)
- Modify: `internal/modules/documents/domain/model.go` (add `TemplateSnapshot` field on `Document`)
- Test: `internal/modules/documents/repository/repository_create_atomic_test.go` (new sqlmock)

- [ ] **Step 1: Write failing sqlmock test**

```go
// internal/modules/documents/repository/repository_create_atomic_test.go
package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/documents/domain"
	templatesdomain "metaldocs/internal/modules/templates_v2/domain"
)

func TestCreateDocument_WritesSnapshotInSameTx(t *testing.T) {
	db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	defer db.Close()
	r := New(db)

	cdID := "cd-1"
	doc := &domain.Document{
		TenantID:             "t1",
		TemplateVersionID:    "tv1",
		Name:                 "n",
		FormDataJSON:         []byte(`{}`),
		CreatedBy:            "u1",
		ControlledDocumentID: &cdID,
		TemplateSnapshot: domain.TemplateSnapshot{
			PlaceholderSchemaJSON: []byte(`[{"key":"a","required":true}]`),
			BodyDocxKey:           "key",
		},
	}
	phs := []templatesdomain.Placeholder{{Key: "a", Required: true}}

	mock.ExpectBegin()
	mock.ExpectExec(`SELECT pg_advisory_xact_lock`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("INSERT INTO documents").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("doc-1"))
	mock.ExpectQuery("INSERT INTO editor_sessions").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("sess-1"))
	mock.ExpectQuery("INSERT INTO document_revisions").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("rev-1"))
	mock.ExpectExec("UPDATE editor_sessions").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE documents SET current_revision_id").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE documents SET placeholder_schema_snapshot").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO placeholder_values").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if _, _, _, err := r.CreateDocument(context.Background(), doc, "hash", phs); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run to verify FAIL**

Run: `go test ./internal/modules/documents/repository/ -run TestCreateDocument_WritesSnapshotInSameTx -v`
Expected: FAIL — signature mismatch (CreateDocument doesn't accept `[]Placeholder`).

- [ ] **Step 3: Add TemplateSnapshot field to domain.Document**

In `internal/modules/documents/domain/model.go`, add to `Document` struct:

```go
// TemplateSnapshot is the frozen template payload. Populated by Service.Create
// from SnapshotService.ResolveTemplate before CreateDocument INSERT, so all
// snapshot columns are written in the same tx as the row itself.
TemplateSnapshot TemplateSnapshot `json:"-"`
```

- [ ] **Step 4: Modify CreateDocument signature + body**

Update `internal/modules/documents/repository/repository.go`:

```go
func (r *Repository) CreateDocument(
	ctx context.Context,
	d *domain.Document,
	initialContentHash string,
	requiredPlaceholders []templatesdomain.Placeholder,
) (docID, revID, sessionID string, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return "", "", "", err
	}
	defer tx.Rollback()

	// (advisory lock from Phase 2 stays here)

	// (document INSERT — unchanged, returns docID)

	// (editor_sessions INSERT — unchanged)

	// (document_revisions INSERT — unchanged)

	// (UPDATE editor_sessions ack — unchanged)

	// (UPDATE documents pointers — unchanged)

	// NEW: write snapshot columns in same tx
	if _, err := tx.ExecContext(ctx, `
		UPDATE documents
		   SET placeholder_schema_snapshot = $1,
		       placeholder_schema_hash     = $2,
		       composition_config_snapshot = $3,
		       body_docx_snapshot_s3_key   = $4
		 WHERE id = $5`,
		d.TemplateSnapshot.PlaceholderSchemaJSON,
		d.TemplateSnapshot.PlaceholderSchemaHash,
		d.TemplateSnapshot.CompositionConfigJSON,
		d.TemplateSnapshot.BodyDocxKey,
		docID,
	); err != nil {
		return "", "", "", fmt.Errorf("write snapshot: %w", err)
	}

	// NEW: seed placeholder_values rows in same tx
	for _, p := range requiredPlaceholders {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO placeholder_values (revision_id, placeholder_key, value)
			VALUES ($1, $2, '')`,
			revID, p.Key,
		); err != nil {
			return "", "", "", fmt.Errorf("seed placeholder %q: %w", p.Key, err)
		}
	}

	return docID, revID, sessionID, tx.Commit()
}
```

Add import: `templatesdomain "metaldocs/internal/modules/templates_v2/domain"`.

- [ ] **Step 5: Run to verify PASS**

Run: `go test ./internal/modules/documents/repository/ -run TestCreateDocument_WritesSnapshotInSameTx -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/modules/documents/repository/repository.go internal/modules/documents/domain/model.go internal/modules/documents/repository/repository_create_atomic_test.go
git commit -m "feat(documents): C2/C4 CreateDocument writes snapshot + placeholder seed atomically"
```

### Task 3.3: Switch Service.Create to use ResolveTemplate + atomic CreateDocument

**Files:**
- Modify: `internal/modules/documents/application/service.go:308-357` (both Create code paths)
- Test: existing `create_document_snapshot_integration_test.go` validates

- [ ] **Step 1: Replace post-commit SnapshotFromTemplate calls with pre-INSERT resolve**

In `service.go` around lines 308–331 (docgen path) and 339–357 (fallback path), replace:

```go
doc := buildDocumentForCreate(cmd, cd, resolvedTemplateVersionID)
docID, revID, sessionID, err := s.repo.CreateDocument(ctx, &doc, contentHash)
if err != nil {
	return nil, err
}
// ... S3 adoption / SetRevisionStorageKey ...
if s.snapshotSvc != nil {
	if err := s.snapshotSvc.SnapshotFromTemplate(ctx, cmd.TenantID, docID, revID, resolvedTemplateVersionID); err != nil {
		return nil, fmt.Errorf("snapshot template: %w", err)
	}
}
```

with:

```go
var snap domain.TemplateSnapshot
var phs []templatesdomain.Placeholder
if s.snapshotSvc != nil {
	var err error
	snap, phs, err = s.snapshotSvc.ResolveTemplate(ctx, cmd.TenantID, resolvedTemplateVersionID)
	if err != nil {
		return nil, fmt.Errorf("resolve template snapshot: %w", err)
	}
}
doc := buildDocumentForCreate(cmd, cd, resolvedTemplateVersionID)
doc.TemplateSnapshot = snap
docID, revID, sessionID, err := s.repo.CreateDocument(ctx, &doc, contentHash, phs)
if err != nil {
	return nil, err
}
// ... S3 adoption / SetRevisionStorageKey ...
// (snapshotSvc.SnapshotFromTemplate post-commit call removed)
```

Apply same change to BOTH code paths (docgen-configured and fallback).

Add import to `service.go`: `templatesdomain "metaldocs/internal/modules/templates_v2/domain"`.

- [ ] **Step 2: Run integration tests**

Run: `go test -tags=integration ./internal/modules/documents/application/ -run TestCreateDocument -v`
Expected: PASS. All snapshot columns populated post-create.

- [ ] **Step 3: Run full unit suite**

Run: `go test ./internal/modules/documents/...`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/modules/documents/application/service.go
git commit -m "refactor(documents): C2/C4 Service.Create uses ResolveTemplate + atomic INSERT"
```

### Task 3.4: Mark legacy SnapshotFromTemplate deprecated

**Files:**
- Modify: `internal/modules/documents/application/snapshot_service.go:45-48`

- [ ] **Step 1: Add deprecation comment**

```go
// Deprecated: SnapshotFromTemplate writes snapshot post-commit, breaking
// atomicity guarantees (see audit C2/C4). Use ResolveTemplate + pass payload
// to Repository.CreateDocument instead. Retained only for backfill scripts.
func (s *SnapshotService) SnapshotFromTemplate(ctx context.Context, tenantID, docID, revisionID, templateID string) error {
```

- [ ] **Step 2: Commit**

```bash
git add internal/modules/documents/application/snapshot_service.go
git commit -m "chore(documents): deprecate SnapshotFromTemplate (post-commit path)"
```

---

## Phase 3 Review (Opus)

- [ ] **Dispatch Opus reviewer**

Prompt:
```
Review Phase 3 commits (C2/C4). Confirm: (a) snapshot resolution happens
before CreateDocument INSERT, (b) snapshot + placeholder_values writes
inside same tx as documents row, (c) failure of resolve aborts before any
row created, (d) S3 AdoptTempObject still runs after commit (acceptable
since it's an idempotent move), (e) deprecated SnapshotFromTemplate kept
for backfill only. PASS/FAIL.
```

---

## Phase 4 — Caller Cleanup (parallel)

### Task 4.1: Drop fromFinalized from Service.Archive + delivery (Codex)

**Files:**
- Modify: `internal/modules/documents/application/service.go:594-604`
- Modify: `internal/modules/documents/delivery/http/handler.go:49,375,386`
- Modify: `internal/modules/documents/delivery/http/handler_test.go:131`

- [ ] **Step 1: Update Service.Archive to use MarkArchived**

Replace lines 594–604 in `service.go`:

```go
// Archive soft-archives a document (sets archived_at). Status field is
// unchanged — terminal states remain immutable for audit. Idempotent.
func (s *Service) Archive(ctx context.Context, tenantID, docID, actorID string) error {
	if err := s.repo.MarkArchived(ctx, tenantID, docID, actorID); err != nil {
		return err
	}
	s.audit.Write(ctx, tenantID, actorID, "document.archived", docID, nil)
	return nil
}
```

- [ ] **Step 2: Update interface in handler.go:49**

```go
Archive(ctx context.Context, tenantID, docID, actorID string) error
```

- [ ] **Step 3: Update call sites in handler.go**

Lines 375, 386 — replace with single path (no fromFinalized branching):

```go
if err := h.svc.Archive(r.Context(), tenantID, docID, userID); err != nil {
	writeErr(w, err)
	return
}
```

(Remove the dual-call pattern entirely; archive is one operation.)

- [ ] **Step 4: Update handler_test.go fake**

Line 131:
```go
func (f *fakeSvc) Archive(_ context.Context, _, _, _ string) error { return nil }
```

- [ ] **Step 5: Run tests**

Run: `go test ./internal/modules/documents/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/modules/documents/application/service.go internal/modules/documents/delivery/http/handler.go internal/modules/documents/delivery/http/handler_test.go
git commit -m "refactor(documents): C1 drop fromFinalized; Archive uses MarkArchived"
```

### Task 4.2: Remove finalized_at writes/reads from repository + domain (Codex)

**Files:**
- Modify: `internal/modules/documents/repository/repository.go:111,115,145,156,172,183,192-201` (any `finalized_at` SELECT/UPDATE)
- Modify: `internal/modules/documents/domain/model.go` (drop `FinalizedAt` field)
- Modify: any caller of `doc.FinalizedAt` (use grep)

- [ ] **Step 1: Find all references**

Run: `grep -rn "finalized_at\|FinalizedAt" internal/modules/documents/`

- [ ] **Step 2: Remove `finalized_at` from SELECTs**

In `repository.go` lines 109–125 (`GetDocument`), 141–163 (`ListDocuments`), 168–190 (`ListDocumentsForUser`): drop `finalized_at` from SELECT column list and corresponding `Scan` field. Drop `&d.FinalizedAt` from each `rows.Scan(...)` call.

- [ ] **Step 3: Remove finalized_at branch from UpdateDocumentStatus**

`repository.go:192-213`:

```go
func (r *Repository) UpdateDocumentStatus(ctx context.Context, tenantID, id string, cur, next domain.DocumentStatus, stampTime bool) error {
	col := ""
	if stampTime && next == domain.DocStatusArchived {
		col = "archived_at = now(),"
	}
	res, err := r.db.ExecContext(ctx,
		fmt.Sprintf(`UPDATE documents SET status=$1, %s updated_at=now() WHERE id=$2 AND tenant_id=$3 AND status=$4`, col),
		next, id, tenantID, cur)
	// ... rest unchanged
}
```

(Note: `archived_at` write here is now redundant since `MarkArchived` is the canonical path, but leave it as belt-and-suspenders for any caller still using `UpdateDocumentStatus(... → archived)`.)

- [ ] **Step 4: Remove FinalizedAt field from domain.Document**

In `internal/modules/documents/domain/model.go`, drop the `FinalizedAt sql.NullTime` field.

- [ ] **Step 5: Compile + test**

Run: `go build ./... && go test ./internal/modules/documents/...`
Expected: PASS. Any external readers of `doc.FinalizedAt` must switch to `v_document_finalized` view (separate task if found).

- [ ] **Step 6: Commit**

```bash
git add internal/modules/documents/
git commit -m "refactor(documents): C6 remove FinalizedAt from domain + repo (use v_document_finalized)"
```

### Task 4.3: List-query archived_at filter (Sonnet)

**Files:**
- Modify: `internal/modules/documents/repository/repository.go:141-163` (`ListDocuments`)
- Modify: `internal/modules/documents/repository/repository.go:168-190` (`ListDocumentsForUser`)
- Test: `internal/modules/documents/repository/repository_list_test.go` (new sqlmock)

- [ ] **Step 1: Write failing test**

```go
// internal/modules/documents/repository/repository_list_test.go
package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListDocuments_ExcludesArchivedByDefault(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	r := New(db)

	mock.ExpectQuery(`archived_at IS NULL`).
		WithArgs("tenant-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	if _, err := r.ListDocuments(context.Background(), "tenant-1"); err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run to verify FAIL**

Run: `go test ./internal/modules/documents/repository/ -run TestListDocuments_ExcludesArchived -v`
Expected: FAIL — query did not contain `archived_at IS NULL`.

- [ ] **Step 3: Add `AND archived_at IS NULL` predicate**

Both `ListDocuments` and `ListDocumentsForUser` queries — append to `WHERE` clauses:

```go
WHERE tenant_id=$1 AND archived_at IS NULL ORDER BY updated_at DESC
```
and
```go
WHERE tenant_id=$1 AND created_by=$2 AND archived_at IS NULL ORDER BY updated_at DESC
```

- [ ] **Step 4: Run to verify PASS**

Run: `go test ./internal/modules/documents/repository/ -run TestListDocuments_ExcludesArchived -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/modules/documents/repository/
git commit -m "feat(documents): C1 list queries exclude archived by default (archived_at IS NULL)"
```

---

## Phase 4 Review (Opus)

- [ ] **Dispatch Opus reviewer**

Prompt:
```
Review Phase 4 commits. Confirm: (a) Archive call sites use new
no-fromFinalized signature, (b) all FinalizedAt references removed,
(c) list queries default-filter archived_at IS NULL, (d) build is green
end-to-end. PASS/FAIL.
```

---

## Phase 5 — Full Verification

### Task 5.1: Run full test suite

- [ ] **Step 1: Unit + integration**

```bash
go test -mod=mod ./...
go test -tags=integration ./tests/integration/documents/...
```

Expected: all PASS.

- [ ] **Step 2: Smoke test**

```powershell
.\scripts\start-api.ps1 -Build
# Login, create doc from template, submit, approve, archive
# Verify via /api/v1/documents/{id}: archived_at populated, status preserved
# Verify list does not include archived
```

### Task 5.2: Codex independent audit

- [ ] **Dispatch codex auditor**

Prompt (caveman):
```
Independent audit Group C bugs C1-C6 against branch group-c-lifecycle.
For each bug: produce PASS/FAIL with file:line evidence.
- C1: Service.Archive uses MarkArchived, no status change, list excludes
- C2: snapshot resolved pre-INSERT
- C3: trigger verified false positive
- C4: snapshot + placeholder_values inside CreateDocument tx
- C5: pg_advisory_xact_lock present, key has tenant prefix
- C6: finalized_at column dropped, view exists, signature dropped fromFinalized
No fixes. Report only.
```

Block on any FAIL.

### Task 5.3: Lint + coverage

- [ ] **Step 1: Lint**

```bash
go vet ./...
golangci-lint run ./internal/modules/documents/...
```

Expected: zero new warnings vs main.

- [ ] **Step 2: Coverage**

```bash
go test -cover ./internal/modules/documents/...
```

Expected: new code ≥80% line coverage.

---

## Phase 5 Review (Opus)

- [ ] **Dispatch Opus reviewer**

Prompt:
```
Final review Group C. Confirm tests, codex audit, lint, coverage all green.
Identify any spec acceptance criterion not yet met. PASS/FAIL.
```

---

## Phase 6 — Wiki + Audit + Merge

### Task 6.1: ADR for soft-archive

**Files:**
- Create: `wiki/decisions/0008-soft-archive-via-timestamp.md`

- [ ] **Step 1: Write ADR**

```markdown
# ADR 0008 — Soft-archive documents via archived_at timestamp

> Status: accepted 2026-05-03

## Context

Migration 0142 enforces a strict status-transition trigger
(`enforce_document_transition`). It defines no transition into `archived`.
The original `Service.Archive` attempted `UpdateDocumentStatus(... → archived)`
and was rejected at runtime (audit C1).

QMS regulatory requirement: a controlled document's terminal status
(`published`, `superseded`, `obsolete`) must remain visible in the audit
trail unchanged. Replacing it with `archived` discards evidence of the
document's lifecycle outcome.

## Decision

Archive is a soft-hide via `documents.archived_at` timestamp. Status field
is never changed by archive. Default list/search queries filter
`archived_at IS NULL`. Admin endpoints opt in to include archived.

`Service.Archive(tenantID, docID, actorID)` — no `fromFinalized` parameter.
Symmetric `Unarchive` clears the timestamp.

`finalized_at` is **not** retained as a denormalized column (see C6) —
finalization timestamp derives from `document_state_history` via the
`v_document_finalized` view. Different from `archived_at` because
`archived_at` is a hot-path filter predicate (queried per list) while
`finalized_at` is cold-path audit data.

## Consequences

- No trigger 0142 change required
- Status field remains source of truth for lifecycle outcome
- `archived_at IS NULL` predicate added to default queries
- Two readers of "is archived?" (status vs timestamp) collapsed to one
- Frontend list views unchanged (default filter applied at repo level)

## References

- Spec: `docs/superpowers/specs/2026-05-03-group-c-doc-lifecycle-design.md`
- Audit: `wiki/bugs/audit-2026-05-03.md` (C1)
- Trigger: `migrations/0142_disable_legacy_compat.sql`
```

- [ ] **Step 2: Commit**

```bash
git add wiki/decisions/0008-soft-archive-via-timestamp.md
git commit -m "docs(adr): 0008 soft-archive documents via archived_at timestamp"
```

### Task 6.2: Close audit entries

**Files:**
- Modify: `wiki/bugs/audit-2026-05-03.md` (C1, C2, C4, C5, C6 entries — C3 already closed in Task 1.1)

- [ ] **Step 1: Mark each closed**

For each of C1, C2, C4, C5, C6 in the audit doc, add `[x]` and a closure line:
```markdown
- [x] **C1** — fixed in <commit-sha>. archived_at soft-archive, status preserved.
```

(Substitute actual commit SHAs from `git log --oneline group-c-lifecycle ^main`.)

- [ ] **Step 2: Commit**

```bash
git add wiki/bugs/audit-2026-05-03.md
git commit -m "docs(audit): close C1-C6 with commit SHAs"
```

### Task 6.3: Wiki-curator subagent dispatch

- [ ] **Dispatch wiki-curator agent**

Prompt:
```
Group C lifecycle work merged. Refresh wiki for:
- wiki/concepts/document-lifecycle.md (if exists; else create)
- wiki/modules/documents-*.md (Last verified stamps)
- wiki/decisions/README.md index (add 0008)
- wiki/README.md index entry for ADR 0008
Verify all file:line anchors still valid post-refactor. Commit changes.
```

### Task 6.4: Finish branch

- [ ] **Step 1: Invoke finishing-a-development-branch skill**

Run tests one more time, choose Option 1 (merge locally) or Option 2 (PR), per user preference.

---

## Self-Review Checklist

**Spec coverage:**
- [x] C1 → Tasks 1.3, 4.1, 4.3
- [x] C2 → Tasks 3.1, 3.2, 3.3
- [x] C3 → Task 1.1
- [x] C4 → Tasks 3.2, 3.3 (atomic placeholder seed)
- [x] C5 → Task 2.1
- [x] C6 → Tasks 1.2, 4.1, 4.2
- [x] ADR 0008 → Task 6.1
- [x] Audit closure → Task 6.2

**Placeholder scan:** none — all steps contain code or exact commands.

**Type consistency:**
- `MarkArchived(ctx, tenantID, docID, actorID string) error` — Task 1.3 + 4.1 match
- `Unarchive(ctx, tenantID, docID, actorID string) error` — Task 1.3 only (used by future endpoint)
- `Archive(ctx, tenantID, docID, actorID string) error` — Task 4.1 service + handler interface match
- `CreateDocument(ctx, *Document, hash, []Placeholder) (docID, revID, sessionID, err)` — Task 3.2 + 3.3 match
- `ResolveTemplate(ctx, tenantID, templateID) (TemplateSnapshot, []Placeholder, error)` — Task 3.1 + 3.3 match
- `domain.Document.TemplateSnapshot` — Task 3.2 + 3.3 match

---

## Acceptance Criteria (from spec)

- [ ] C1: `Service.Archive(published_doc)` succeeds, `archived_at` populated, status unchanged, default list excludes
- [ ] C2/C4: Create with bad `template_version_id` fails before any row written; successful Create populates all 6 snapshot columns + placeholder_values rows in single tx
- [ ] C3: `pg_trigger` query returns trigger row; audit doc updated with evidence
- [ ] C5: 10 concurrent submits same `controlled_document_id` produce monotonic revisions 1..10, zero duplicate-key errors
- [ ] C6: `finalized_at` column dropped; `v_document_finalized` returns approval timestamp; `Service.Archive` signature drops `fromFinalized`
- [ ] All Go tests pass with `-mod=mod`
- [ ] Integration tests pass with `-tags=integration`
- [ ] Codex audit returns 6/6 PASS (with C3 marked false-positive-with-evidence)
- [ ] Smoke test: full create→submit→approve→archive flow completes
- [ ] Audit doc updated, all 6 entries closed with commit SHAs

---

## References

- Spec: `docs/superpowers/specs/2026-05-03-group-c-doc-lifecycle-design.md`
- Audit: `wiki/bugs/audit-2026-05-03.md`
- Migration 0142: `migrations/0142_disable_legacy_compat.sql`
- Migration 0152: `migrations/0152_placeholder_fillin_columns.sql`
- Group A spec: `docs/superpowers/specs/2026-05-03-group-a-blockers.md`
- Group B spec: `docs/superpowers/specs/2026-05-03-group-b-authz-cleanup-design.md`
- Group D spec: `docs/superpowers/specs/2026-05-03-group-d-freeze-quality-design.md`
- Phase 12 column-drop alignment: `docs/superpowers/plans/2026-04-21-foundation-doc-approval-state-machine.md`
