# Group D — Freeze Quality Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix 8 freeze-quality bugs (D1-D8) that erode the regulatory integrity of MetalDocs frozen DOCX artifacts.

**Architecture:** Three sub-areas — tx integrity (D7 via DBTX variadic propagation), resolver correctness (D1-D3, D8 via snapshot-at-action-time + strict NULL handling), PDF dispatch durability (D4-D6 via transactional outbox + background worker).

**Tech Stack:** Go 1.22, PostgreSQL 16, sqlmock for unit tests, `-tags=integration` for live-DB tests, `messaging.Publisher` event bus, Prometheus metrics.

**Spec:** `docs/superpowers/specs/2026-05-03-group-d-freeze-quality-design.md`

---

## Model Strategy

| Role | Model | Why |
|---|---|---|
| Implementer (complex: outbox, context builder, decision_service) | codex | Multi-file integration, design judgment |
| Implementer (simple: typed error, single-file resolver tweak, migration) | sonnet | Mechanical, single file |
| Wiki-curator agent | haiku | Mechanical doc updates |
| Phase review | opus | Architectural sanity check |
| Final review + plan validation | codex | Independent audit |
| Spec compliance reviewer | sonnet | Read spec + diff |
| Code quality reviewer | sonnet | Style + idiom check |

---

## File Map

**Created:**
- `migrations/0173_signoff_actor_displayname_snapshot.sql`
- `migrations/0174_documents_created_by_displayname_snapshot.sql`
- `migrations/0175_documents_area_name_snapshot.sql`
- `migrations/0176_pdf_dispatch_outbox.sql`
- `internal/modules/render/fanout/pdf_outbox_repository.go`
- `internal/modules/render/fanout/pdf_outbox_repository_test.go`
- `internal/modules/render/fanout/pdf_outbox_worker.go`
- `internal/modules/render/fanout/pdf_outbox_worker_test.go`
- `tests/integration/freeze/group_d_test.go`
- `wiki/decisions/0008-pdf-dispatch-outbox.md`

**Modified:**
- `internal/modules/documents/repository/fillin_repository.go` — UpsertValue tx variadic
- `internal/modules/documents/application/freeze_service.go` — pass tx to UpsertValue, drop misleading comment
- `internal/modules/render/resolvers/resolver.go` — add AreaNameSnapshot, ApprovalInstanceID to ResolveInput
- `internal/modules/render/resolvers/effective_date.go` — typed error on NULL
- `internal/modules/render/resolvers/controlled_by_area.go` — return name, hash on name
- `internal/modules/render/resolvers/approvers.go` — filter by approval_instance_id, read snapshot column
- `internal/modules/render/resolvers/author.go` — read created_by snapshot column
- `internal/modules/documents/domain/errors.go` — ErrEffectiveDateMissing
- `internal/modules/documents/approval/application/decision_service.go` — write actor_display_name_snapshot, replace pdfDispatcher.Dispatch with outbox.Enqueue
- `internal/modules/documents/application/service.go` (Create) — write created_by + area_name snapshots
- `internal/modules/documents/application/context_builder.go` — pass AreaNameSnapshot, ApprovalInstanceID
- `internal/platform/bootstrap/api.go` — wire outbox repo + worker

**Deleted (after Phase 4 verified):**
- nothing immediate; `PDFDispatcher` retained for backwards compat until Phase 6.

---

## Phase 0: Worktree + Plan Validation

### Task 0.1: Create worktree

- [ ] **Step 1:** Create worktree off main

```powershell
git worktree add ../MetalDocs-group-d -b group-d-freeze-quality main
```

Expected: worktree at `../MetalDocs-group-d`, branch `group-d-freeze-quality`.

- [ ] **Step 2:** Switch CWD into worktree for remaining tasks. Verify clean.

```powershell
cd ../MetalDocs-group-d
git status
```

Expected: "On branch group-d-freeze-quality" + clean.

### Task 0.2: Codex spec validation

- [ ] **Step 1:** Dispatch codex agent to audit spec for inconsistencies.

Dispatch (caveman prompt):
```
audit spec docs/superpowers/specs/2026-05-03-group-d-freeze-quality-design.md.

check:
1. all 8 bugs (D1-D8) covered with concrete files + tests
2. migration order 0173-0176 conflict-free with main (latest 0170)
3. no contradiction between sub-areas (e.g. D1 snapshot vs D8 instance filter overlap)
4. acceptance criteria all testable
5. outbox schema sound (status enum, indexes, FOR UPDATE SKIP LOCKED viability)
6. ResolveInput field additions don't break existing resolvers

report PASS/FAIL per check, file:line evidence. under 300 words.
```

Expected: 6/6 PASS. If FAIL: fix spec before proceeding.

### Task 0.3: Wiki-curator pre-check

- [ ] **Step 1:** Dispatch wiki-curator agent for current freeze docs state.

Dispatch:
```
report current state of wiki freeze pipeline docs.
list:
- wiki/concepts/freeze-pipeline.md (exists? Last verified date?)
- wiki/concepts/placeholders.md (last verified?)
- wiki/modules/render-* (which files? stamps?)
- wiki/decisions/ (highest ADR number?)

no changes — pre-check only. under 100 words.
```

Expected: report. Use to plan Phase 6 wiki updates.

---

## Phase 1: Parallel — D7, D3, D8

Three independent file sets. Dispatch all three subagents in parallel.

### Task 1.1: D7 — UpsertValue tx propagation (codex)

**Files:**
- Modify: `internal/modules/documents/repository/fillin_repository.go`
- Modify: `internal/modules/documents/application/freeze_service.go:145-156`
- Test: `internal/modules/documents/repository/fillin_repository_test.go` (new or extend existing)

- [ ] **Step 1: Write failing unit test** for UpsertValue tx routing.

`internal/modules/documents/repository/fillin_repository_test.go`:
```go
package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUpsertValue_UsesTxWhenProvided(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO document_placeholder_values").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	repo := NewFillInRepository(db)
	val := "x"
	v := PlaceholderValue{
		TenantID: "00000000-0000-0000-0000-000000000001",
		RevisionID: "00000000-0000-0000-0000-000000000002",
		PlaceholderID: "p1", ValueText: &val, Source: "computed",
	}
	if err := repo.UpsertValue(context.Background(), v, tx); err != nil {
		t.Fatalf("UpsertValue: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUpsertValue_FallsBackToDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO document_placeholder_values").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewFillInRepository(db)
	val := "x"
	v := PlaceholderValue{
		TenantID: "00000000-0000-0000-0000-000000000001",
		RevisionID: "00000000-0000-0000-0000-000000000002",
		PlaceholderID: "p1", ValueText: &val, Source: "computed",
	}
	if err := repo.UpsertValue(context.Background(), v); err != nil {
		t.Fatalf("UpsertValue: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
```

- [ ] **Step 2: Run — expect FAIL** (signature doesn't match).

```powershell
go test -mod=mod ./internal/modules/documents/repository/ -run TestUpsertValue_ -v
```

Expected: compile error `too many arguments`.

- [ ] **Step 3: Implement variadic.** Replace `UpsertValue` body in `fillin_repository.go:73`.

```go
func (r *FillInRepository) UpsertValue(ctx context.Context, v PlaceholderValue, q ...DBTX) error {
	var valueTyped any
	if v.ValueTyped != nil {
		b, err := json.Marshal(v.ValueTyped)
		if err != nil {
			return err
		}
		valueTyped = b
	}

	exec := DBTX(r.db)
	if len(q) > 0 && q[0] != nil {
		exec = q[0]
	}

	_, err := exec.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s
		    (tenant_id, revision_id, placeholder_id, value_text, value_typed,
		     source, computed_from, resolver_version, inputs_hash, validated_at, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW(), NOW())
		ON CONFLICT (tenant_id, revision_id, placeholder_id) DO UPDATE SET
			value_text       = EXCLUDED.value_text,
			value_typed      = EXCLUDED.value_typed,
			source           = EXCLUDED.source,
			computed_from    = EXCLUDED.computed_from,
			resolver_version = EXCLUDED.resolver_version,
			inputs_hash      = EXCLUDED.inputs_hash,
			validated_at     = NOW(),
			updated_at       = NOW()`, r.table("document_placeholder_values")),
		v.TenantID, v.RevisionID, v.PlaceholderID, v.ValueText, valueTyped,
		v.Source, v.ComputedFrom, v.ResolverVersion, v.InputsHash,
	)
	return err
}
```

DBTX type already exists in `internal/modules/documents/repository/snapshot_repository.go:19`. Reuse — do NOT redefine.

- [ ] **Step 4: Update FillInWriter interface.** Find via grep:

```powershell
grep -rn "FillInWriter" internal/
```

Update `internal/modules/documents/application/freeze_service.go` (or wherever interface lives) — add variadic:

```go
type FillInWriter interface {
	UpsertValue(ctx context.Context, v repository.PlaceholderValue, q ...repository.DBTX) error
}
```

- [ ] **Step 5: Pass tx in freeze_service.** Modify `freeze_service.go:145-156`. Replace block:

```go
		// Persist within the approval tx so rollback discards stale computed values.
		if err := s.values.UpsertValue(ctx, repository.PlaceholderValue{
			TenantID: tenantID, RevisionID: revisionID, PlaceholderID: p.ID,
			ValueText: &strVal, Source: "computed",
			ComputedFrom: &key, ResolverVersion: &ver,
			InputsHash: rv.InputsHash,
		}, tx); err != nil {
			return err
		}
```

Drop the misleading comment block (lines 145-148 in original).

- [ ] **Step 6: Run unit tests — PASS.**

```powershell
go test -mod=mod ./internal/modules/documents/repository/ -run TestUpsertValue_ -v
go test -mod=mod ./internal/modules/documents/application/ -v
```

Expected: PASS.

- [ ] **Step 7: Commit.**

```powershell
git add internal/modules/documents/repository/fillin_repository.go internal/modules/documents/repository/fillin_repository_test.go internal/modules/documents/application/freeze_service.go
git commit -m "fix(D7): propagate tx through UpsertValue, ensure rollback discards computed values"
```

### Task 1.2: D3 — Effective date NULL → typed error (sonnet)

**Files:**
- Modify: `internal/modules/documents/domain/errors.go`
- Modify: `internal/modules/render/resolvers/effective_date.go`
- Test: `internal/modules/render/resolvers/effective_date_test.go`

- [ ] **Step 1: Add typed error** to `internal/modules/documents/domain/errors.go`:

```go
// ErrEffectiveDateMissing indicates a freeze attempt where the document
// revision has no effective_from date set. Surfaced as 422 to the caller;
// upstream workflow must populate the date before freeze.
var ErrEffectiveDateMissing = errors.New("effective_date missing")
```

If file or import doesn't exist, create with `package domain` + `import "errors"`.

- [ ] **Step 2: Write failing test** `internal/modules/render/resolvers/effective_date_test.go`:

```go
package resolvers

import (
	"context"
	"errors"
	"testing"
	"time"

	v2dom "metaldocs/internal/modules/documents/domain"
)

type stubRevReaderZero struct{}

func (stubRevReaderZero) GetRevisionNumber(_ context.Context, _, _ string) (int64, error) {
	return 0, nil
}
func (stubRevReaderZero) GetEffectiveFrom(_ context.Context, _, _ string) (time.Time, error) {
	return time.Time{}, nil
}
func (stubRevReaderZero) GetAuthor(_ context.Context, _, _ string) (AuthorInfo, error) {
	return AuthorInfo{}, nil
}

func TestEffectiveDateResolver_NullReturnsTypedError(t *testing.T) {
	r := EffectiveDateResolver{}
	in := ResolveInput{
		TenantID: "t1", RevisionID: "r1",
		RevisionReader: stubRevReaderZero{},
	}
	_, err := r.Resolve(context.Background(), in)
	if !errors.Is(err, v2dom.ErrEffectiveDateMissing) {
		t.Fatalf("want ErrEffectiveDateMissing, got %v", err)
	}
}
```

- [ ] **Step 3: Run — expect FAIL** (resolver returns time-formatted zero).

```powershell
go test -mod=mod ./internal/modules/render/resolvers/ -run TestEffectiveDateResolver_NullReturnsTypedError -v
```

Expected: FAIL (current resolver formats `time.Time{}.UTC().Format(...)`).

- [ ] **Step 4: Implement.** Replace `effective_date.go` Resolve body:

```go
func (EffectiveDateResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	effectiveFrom, err := in.RevisionReader.GetEffectiveFrom(ctx, in.TenantID, in.RevisionID)
	if err != nil {
		return ResolvedValue{}, err
	}
	if effectiveFrom.IsZero() {
		return ResolvedValue{}, v2dom.ErrEffectiveDateMissing
	}

	inputsHash, err := hashInputs(struct {
		TenantID   string `json:"tenant_id"`
		RevisionID string `json:"revision_id"`
	}{in.TenantID, in.RevisionID})
	if err != nil {
		return ResolvedValue{}, err
	}

	return ResolvedValue{
		Value:       effectiveFrom.UTC().Format("2006-01-02"),
		ResolverKey: "effective_date",
		ResolverVer: 1,
		InputsHash:  inputsHash,
		ComputedAt:  time.Now().UTC(),
	}, nil
}
```

Add import `v2dom "metaldocs/internal/modules/documents/domain"`.

- [ ] **Step 5: Drop COALESCE in RevisionReader SQL.** Find via grep:

```powershell
grep -rn "GetEffectiveFrom" internal/
grep -rn "COALESCE.*effective_from" internal/
```

If SQL has `COALESCE(effective_from, NOW())`, replace with bare `effective_from`. Adjust scan to use `*time.Time` or `sql.NullTime`, return zero `time.Time{}` for NULL.

- [ ] **Step 6: Run — PASS.**

```powershell
go test -mod=mod ./internal/modules/render/resolvers/ -run TestEffectiveDateResolver_ -v
go test -mod=mod ./internal/modules/render/... -v
```

- [ ] **Step 7: Commit.**

```powershell
git add internal/modules/documents/domain/errors.go internal/modules/render/resolvers/effective_date.go internal/modules/render/resolvers/effective_date_test.go
git commit -m "fix(D3): return ErrEffectiveDateMissing on NULL effective_from instead of NOW() fallback"
```

### Task 1.3: D8 — Approvers filter by instance (sonnet)

**Files:**
- Modify: `internal/modules/render/resolvers/resolver.go` — add `ApprovalInstanceID` to `ResolveInput`
- Modify: `internal/modules/render/resolvers/approvers.go` — pass instance ID
- Modify: `WorkflowReader.GetApprovers` — accept instance ID parameter
- Find caller (postgres impl of WorkflowReader) and update SQL to filter by instance
- Test: `internal/modules/render/resolvers/approvers_test.go` (extend)

- [ ] **Step 1: Extend ResolveInput** in `resolver.go`:

```go
type ResolveInput struct {
	TenantID, RevisionID, ControlledDocumentID string
	ProfileCodeSnapshot, AreaCodeSnapshot      string
	AreaNameSnapshot                           string // D2 (added Phase 3)
	ApprovalInstanceID                         string // D8
	RegistryReader                             RegistryReader
	RevisionReader                             RevisionReader
	WorkflowReader                             WorkflowReader
	DocumentReader                             DocumentReader
}
```

(Add `AreaNameSnapshot` now to avoid Phase 3 conflict; populate in Phase 3.)

- [ ] **Step 2: Update WorkflowReader interface** in `resolver.go`:

```go
type WorkflowReader interface {
	GetApprovers(ctx context.Context, tenantID, revisionID, approvalInstanceID string) ([]ApproverInfo, error)
	GetFinalApprovalDate(ctx context.Context, tenantID, revisionID string) (time.Time, error)
}
```

- [ ] **Step 3: Write failing test** `approvers_test.go` (extend):

```go
func TestApproversResolver_FiltersByInstanceID(t *testing.T) {
	want := []ApproverInfo{{UserID: "u-1", DisplayName: "Jane", SignedAt: time.Now()}}
	wr := &fakeWorkflowReader{
		approversByInstance: map[string][]ApproverInfo{
			"inst-current": want,
			"inst-stale":   {{UserID: "u-2", DisplayName: "Stale"}},
		},
	}
	r := ApproversResolver{}
	in := ResolveInput{
		TenantID: "t1", RevisionID: "r1",
		ApprovalInstanceID: "inst-current",
		WorkflowReader:     wr,
	}
	out, err := r.Resolve(context.Background(), in)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got, _ := out.Value.(string); got != "Jane" {
		t.Fatalf("want Jane, got %q", got)
	}
}
```

`fakeWorkflowReader` implementation:
```go
type fakeWorkflowReader struct {
	approversByInstance map[string][]ApproverInfo
}

func (f *fakeWorkflowReader) GetApprovers(_ context.Context, _, _, instanceID string) ([]ApproverInfo, error) {
	return f.approversByInstance[instanceID], nil
}
func (f *fakeWorkflowReader) GetFinalApprovalDate(_ context.Context, _, _ string) (time.Time, error) {
	return time.Time{}, nil
}
```

- [ ] **Step 4: Run — FAIL** (signature mismatch).

```powershell
go test -mod=mod ./internal/modules/render/resolvers/ -run TestApproversResolver_FiltersByInstanceID -v
```

- [ ] **Step 5: Update resolver** `approvers.go`:

```go
func (ApproversResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	approvers, err := in.WorkflowReader.GetApprovers(ctx, in.TenantID, in.RevisionID, in.ApprovalInstanceID)
	if err != nil {
		return ResolvedValue{}, err
	}
	// rest unchanged
```

- [ ] **Step 6: Update postgres impl of WorkflowReader.** Find:

```powershell
grep -rn "GetApprovers" internal/ --include="*.go"
```

Locate the postgres implementer. Update SQL:

```sql
SELECT s.actor_user_id, s.actor_display_name_snapshot, s.signed_at
  FROM approval_signoffs s
 WHERE s.tenant_id = $1::uuid
   AND s.approval_instance_id = $2::uuid
   AND s.decision = 'approved'
 ORDER BY s.signed_at ASC
```

Note: `actor_display_name_snapshot` column ships in Phase 2 (Task 2.1). For Phase 1, leave the column reference but add `COALESCE(s.actor_display_name_snapshot, s.actor_user_id::text)` so missing column build fails fast → forces Phase 2 ordering. If migration not yet applied locally, the unit test using fakeWorkflowReader still passes.

Update Go signature:
```go
func (r *WorkflowReaderImpl) GetApprovers(ctx context.Context, tenantID, revisionID, approvalInstanceID string) ([]resolvers.ApproverInfo, error) { ... }
```

- [ ] **Step 7: Update context_builder** to populate `ApprovalInstanceID`. Find:

```powershell
grep -rn "ResolveInput{" internal/ --include="*.go"
```

In `context_builder.go` `Build` method, add SQL to load active instance:

```go
const activeInstSQL = `
SELECT id FROM approval_instances
 WHERE tenant_id = $1::uuid AND document_id = (
   SELECT document_id FROM document_revisions WHERE id = $2::uuid AND tenant_id = $1::uuid
 )
   AND status IN ('approved','in_progress')
 ORDER BY created_at DESC LIMIT 1`
var instanceID sql.NullString
_ = b.db.QueryRowContext(ctx, activeInstSQL, tenantID, revisionID).Scan(&instanceID)
```

Set `out.ApprovalInstanceID = instanceID.String` (empty string OK if no instance — resolvers handle gracefully).

- [ ] **Step 8: Run all resolver + builder tests — PASS.**

```powershell
go test -mod=mod ./internal/modules/render/... ./internal/modules/documents/application/... -v
```

- [ ] **Step 9: Commit.**

```powershell
git add internal/modules/render/resolvers/resolver.go internal/modules/render/resolvers/approvers.go internal/modules/render/resolvers/approvers_test.go internal/modules/documents/application/context_builder.go [postgres workflow reader file]
git commit -m "fix(D8): filter approvers by approval_instance_id, prevent stale cycle bleed"
```

### Phase 1 Review (Opus)

- [ ] **Step 1:** Dispatch opus phase review.

Prompt:
```
review phase 1 commits (D7, D3, D8) on branch group-d-freeze-quality.

git log main..HEAD --oneline

per commit: verify
1. tests assert correct behavior (not just compile)
2. existing pattern followed (DBTX variadic matches snapshot_repository.go:19)
3. no unrelated changes leaked in
4. caller updates complete (no orphan signatures)

PASS or list issues. under 200 words.
```

Address any issues before Phase 2.

---

## Phase 2: D1 — Snapshot author + signoff display names (codex)

### Task 2.1: Migration 0173 — signoff actor display name snapshot

**File:** `migrations/0173_signoff_actor_displayname_snapshot.sql`

- [ ] **Step 1: Write migration.**

```sql
-- 0173_signoff_actor_displayname_snapshot.sql
-- Adds actor_display_name_snapshot to approval_signoffs for QMS-grade attribution.
-- Backfills from current iam_users.display_name (best-effort historical).

ALTER TABLE metaldocs.approval_signoffs
    ADD COLUMN IF NOT EXISTS actor_display_name_snapshot TEXT;

UPDATE metaldocs.approval_signoffs s
   SET actor_display_name_snapshot = u.display_name
  FROM metaldocs.iam_users u
 WHERE u.user_id = s.actor_user_id
   AND s.actor_display_name_snapshot IS NULL;
```

- [ ] **Step 2: Apply locally + smoke verify.**

```powershell
.\scripts\start-api.ps1 -Build
```

API runs migrations on start. Verify column:

```sql
SELECT column_name FROM information_schema.columns
 WHERE table_schema='metaldocs' AND table_name='approval_signoffs'
   AND column_name='actor_display_name_snapshot';
```

Expected: 1 row.

- [ ] **Step 3: Commit.**

```powershell
git add migrations/0173_signoff_actor_displayname_snapshot.sql
git commit -m "feat(D1): migration 0173 add approval_signoffs.actor_display_name_snapshot"
```

### Task 2.2: Migration 0174 — documents created_by display name snapshot

**File:** `migrations/0174_documents_created_by_displayname_snapshot.sql`

- [ ] **Step 1: Write migration.**

```sql
-- 0174_documents_created_by_displayname_snapshot.sql
ALTER TABLE metaldocs.documents
    ADD COLUMN IF NOT EXISTS created_by_display_name_snapshot TEXT;

UPDATE metaldocs.documents d
   SET created_by_display_name_snapshot = u.display_name
  FROM metaldocs.iam_users u
 WHERE u.user_id = d.created_by
   AND d.created_by_display_name_snapshot IS NULL;
```

- [ ] **Step 2: Apply + verify** (same procedure as 2.1).

- [ ] **Step 3: Commit.**

```powershell
git add migrations/0174_documents_created_by_displayname_snapshot.sql
git commit -m "feat(D1): migration 0174 add documents.created_by_display_name_snapshot"
```

### Task 2.3: Write actor snapshot in decision_service

**File:** `internal/modules/documents/approval/application/decision_service.go`

- [ ] **Step 1: Locate signoff insert SQL.** Find:

```powershell
grep -n "INSERT INTO approval_signoffs\|INSERT INTO metaldocs.approval_signoffs" internal/modules/documents/approval/
```

- [ ] **Step 2: Write failing integration test** `tests/integration/freeze/group_d_test.go` (new):

```go
//go:build integration

package freeze_integration

import (
	"context"
	"database/sql"
	"testing"

	"metaldocs/tests/integration/testdb"
)

func TestRecordSignoff_PersistsActorDisplayNameSnapshot(t *testing.T) {
	ctx := context.Background()
	db := testdb.MustOpen(t)
	defer db.Close()

	// Arrange: seed iam_user with display_name
	mustExec(t, db, `INSERT INTO metaldocs.iam_users (user_id, display_name, is_active) VALUES ('signer-1','Alice Approver',TRUE) ON CONFLICT (user_id) DO UPDATE SET display_name=EXCLUDED.display_name`)

	// Act: invoke recordSignoff path (use real DecisionService against test DB)
	// ... seeded approval_instance + stage; call svc.RecordSignoff with actor=signer-1

	// Assert: approval_signoffs row has actor_display_name_snapshot = 'Alice Approver'
	var snap sql.NullString
	row := db.QueryRowContext(ctx, `SELECT actor_display_name_snapshot FROM metaldocs.approval_signoffs WHERE actor_user_id='signer-1' ORDER BY signed_at DESC LIMIT 1`)
	if err := row.Scan(&snap); err != nil {
		t.Fatalf("query: %v", err)
	}
	if !snap.Valid || snap.String != "Alice Approver" {
		t.Fatalf("want 'Alice Approver', got %v", snap)
	}
}
```

(Helper `mustExec` + DecisionService wiring follows existing patterns in `tests/integration/`.)

- [ ] **Step 3: Run — FAIL** (column not populated).

```powershell
go test -mod=mod -tags=integration ./tests/integration/freeze/ -run TestRecordSignoff_PersistsActorDisplayNameSnapshot -v
```

- [ ] **Step 4: Modify signoff insert.** In `decision_service.go`, before the INSERT, add:

```go
var actorDisplayName sql.NullString
if err := tx.QueryRowContext(ctx, `SELECT display_name FROM metaldocs.iam_users WHERE user_id = $1`, req.ActorUserID).Scan(&actorDisplayName); err != nil && err != sql.ErrNoRows {
	_ = tx.Rollback()
	return SignoffResult{}, fmt.Errorf("recordSignoff: lookup actor display name: %w", err)
}
```

Update INSERT to include `actor_display_name_snapshot`:

```go
const insertSignoff = `
INSERT INTO metaldocs.approval_signoffs (
    id, approval_instance_id, stage_instance_id, actor_user_id, actor_tenant_id,
    actor_display_name_snapshot,
    decision, comment, signed_at, signature_method, signature_payload, content_hash
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
```

Pass `actorDisplayName` (NullString) as parameter. NULL when IAM lookup empty (silent fallback per spec).

- [ ] **Step 5: Run — PASS.**

```powershell
go test -mod=mod -tags=integration ./tests/integration/freeze/ -run TestRecordSignoff_ -v
```

- [ ] **Step 6: Commit.**

```powershell
git add internal/modules/documents/approval/application/decision_service.go tests/integration/freeze/group_d_test.go
git commit -m "fix(D1): persist actor_display_name_snapshot at signoff time"
```

### Task 2.4: Write created_by snapshot in documents.Service.Create

**File:** `internal/modules/documents/application/service.go`

- [ ] **Step 1: Locate Create method + INSERT into documents.** Find:

```powershell
grep -n "INSERT INTO.*documents\b\|INSERT INTO metaldocs.documents" internal/modules/documents/application/
grep -n "func.*Create" internal/modules/documents/application/service.go
```

- [ ] **Step 2: Write failing integration test** (extend `group_d_test.go`):

```go
func TestCreateDocument_PersistsCreatedByDisplayNameSnapshot(t *testing.T) {
	// ... seed iam_user 'author-1' with display_name 'Bob Author'
	// ... call documents.Service.Create with createdBy='author-1'
	// ... assert documents.created_by_display_name_snapshot = 'Bob Author'
}
```

- [ ] **Step 3: Run — FAIL.**

- [ ] **Step 4: Modify Create.** Before document INSERT, lookup display_name + write to new column. Pattern mirrors Task 2.3.

- [ ] **Step 5: Run — PASS.**

- [ ] **Step 6: Commit.**

```powershell
git commit -m "fix(D1): persist created_by_display_name_snapshot at document creation"
```

### Task 2.5: Resolvers read snapshot columns

**Files:**
- `internal/modules/render/resolvers/author.go` — read `documents.created_by_display_name_snapshot`
- `internal/modules/render/resolvers/approvers.go` — already reads via WorkflowReader (Phase 1 Task 1.3 set up SQL)
- WorkflowReader postgres impl — confirm SELECT includes `actor_display_name_snapshot`
- RevisionReader postgres impl — `GetAuthor` returns AuthorInfo with snapshot DisplayName

- [ ] **Step 1: Update `GetAuthor` postgres SQL.** Find:

```powershell
grep -rn "GetAuthor" internal/ --include="*.go"
```

Update SQL to JOIN `documents.created_by_display_name_snapshot`:

```sql
SELECT d.created_by, COALESCE(d.created_by_display_name_snapshot, d.created_by::text) AS display_name
  FROM document_revisions r
  JOIN documents d ON d.id = r.document_id
 WHERE r.tenant_id = $1::uuid AND r.id = $2::uuid
```

- [ ] **Step 2: Confirm WorkflowReader.GetApprovers SQL** (from Phase 1 Task 1.3) selects `COALESCE(s.actor_display_name_snapshot, s.actor_user_id::text)`. If still using `actor_user_id`, fix.

- [ ] **Step 3: Integration test — snapshot retention across rename.**

```go
func TestApproversResolver_UsesSnapshotNotLiveIAM(t *testing.T) {
	// seed user 'u-1' display_name 'Original Name'
	// record signoff for instance inst-1 actor=u-1 → snapshot='Original Name'
	// UPDATE iam_users SET display_name='Renamed' WHERE user_id='u-1'
	// invoke ApproversResolver.Resolve with ApprovalInstanceID=inst-1
	// assert output contains 'Original Name', NOT 'Renamed'
}
```

- [ ] **Step 4: Run — PASS.**

- [ ] **Step 5: Commit.**

```powershell
git commit -m "fix(D1): resolvers read display_name snapshot columns, immune to live IAM renames"
```

### Phase 2 Review (Opus)

- [ ] **Step 1: Dispatch opus.**

Prompt:
```
review phase 2 (D1 snapshots).

verify:
1. snapshot written ATOMICALLY with signoff/create (same tx)
2. resolvers never call live iam_users at freeze time
3. backfill migrations idempotent + handle NULL gracefully
4. AuthorInfo/ApproverInfo DisplayName populated from snapshot column

PASS or list issues. under 200 words.
```

---

## Phase 3: D2 — Area name snapshot (codex)

### Task 3.1: Migration 0175

**File:** `migrations/0175_documents_area_name_snapshot.sql`

- [ ] **Step 1: Write migration.**

```sql
-- 0175_documents_area_name_snapshot.sql
ALTER TABLE metaldocs.documents
    ADD COLUMN IF NOT EXISTS area_name_snapshot TEXT;

UPDATE metaldocs.documents d
   SET area_name_snapshot = pa.name
  FROM metaldocs.process_areas pa
 WHERE pa.tenant_id = d.tenant_id
   AND pa.code      = d.area_code_snapshot
   AND d.area_name_snapshot IS NULL;
```

- [ ] **Step 2: Apply + verify column exists.**

- [ ] **Step 3: Commit.**

```powershell
git commit -m "feat(D2): migration 0175 add documents.area_name_snapshot"
```

### Task 3.2: Populate area_name_snapshot at document create

**File:** `internal/modules/documents/application/service.go`

- [ ] **Step 1: Failing integration test** (extend group_d_test.go):

```go
func TestCreateDocument_PersistsAreaNameSnapshot(t *testing.T) {
	// seed process_areas tenant=t1 code=QA-01 name='Quality Assurance — Area 1'
	// create document with area_code='QA-01'
	// assert documents.area_name_snapshot = 'Quality Assurance — Area 1'
}
```

- [ ] **Step 2: Run — FAIL.**

- [ ] **Step 3: Implement.** In `Service.Create`, before/in INSERT:

```go
var areaName sql.NullString
if err := tx.QueryRowContext(ctx, `SELECT name FROM metaldocs.process_areas WHERE tenant_id=$1::uuid AND code=$2`, tenantID, areaCode).Scan(&areaName); err != nil && err != sql.ErrNoRows {
	_ = tx.Rollback()
	return nil, fmt.Errorf("create document: lookup area name: %w", err)
}
```

Add `area_name_snapshot` column to INSERT.

- [ ] **Step 4: Run — PASS.**

- [ ] **Step 5: Commit.**

```powershell
git commit -m "fix(D2): populate area_name_snapshot at document creation"
```

### Task 3.3: ControlledByAreaResolver returns area name

**File:** `internal/modules/render/resolvers/controlled_by_area.go`

- [ ] **Step 1: Failing test** in `controlled_by_area_test.go`:

```go
func TestControlledByAreaResolver_ReturnsAreaName(t *testing.T) {
	r := ControlledByAreaResolver{}
	in := ResolveInput{
		TenantID: "t1",
		AreaCodeSnapshot: "QA-01",
		AreaNameSnapshot: "Quality Assurance — Area 1",
	}
	out, err := r.Resolve(context.Background(), in)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got := out.Value.(string); got != "Quality Assurance — Area 1" {
		t.Fatalf("want area name, got %q", got)
	}
}

func TestControlledByAreaResolver_FallsBackToCode(t *testing.T) {
	r := ControlledByAreaResolver{}
	in := ResolveInput{TenantID: "t1", AreaCodeSnapshot: "QA-01"}
	out, _ := r.Resolve(context.Background(), in)
	if got := out.Value.(string); got != "QA-01" {
		t.Fatalf("fallback to code, got %q", got)
	}
}
```

- [ ] **Step 2: Run — FAIL.**

- [ ] **Step 3: Implement.** Replace resolver body:

```go
func (ControlledByAreaResolver) Resolve(ctx context.Context, in ResolveInput) (ResolvedValue, error) {
	value := in.AreaNameSnapshot
	if value == "" {
		value = in.AreaCodeSnapshot
	}
	inputsHash, err := hashInputs(struct {
		TenantID         string `json:"tenant_id"`
		AreaNameSnapshot string `json:"area_name_snapshot"`
		AreaCodeFallback string `json:"area_code_fallback,omitempty"`
	}{
		TenantID:         in.TenantID,
		AreaNameSnapshot: in.AreaNameSnapshot,
		AreaCodeFallback: in.AreaCodeSnapshot,
	})
	if err != nil {
		return ResolvedValue{}, err
	}
	return ResolvedValue{
		Value:       value,
		ResolverKey: "controlled_by_area",
		ResolverVer: 2, // bumped — output semantics changed
		InputsHash:  inputsHash,
		ComputedAt:  time.Now().UTC(),
	}, nil
}

func (ControlledByAreaResolver) Version() int { return 2 }
```

- [ ] **Step 4: Populate AreaNameSnapshot in context_builder.** In `loadDocumentSnapshot` (or sibling), SELECT `area_name_snapshot`:

```sql
SELECT controlled_document_id, area_code_snapshot, area_name_snapshot
  FROM documents WHERE id = (SELECT document_id FROM document_revisions WHERE id=$1::uuid AND tenant_id=$2::uuid) AND tenant_id=$2::uuid
```

Set `out.AreaNameSnapshot`.

- [ ] **Step 5: Run — PASS.**

- [ ] **Step 6: Commit.**

```powershell
git commit -m "fix(D2): controlled_by_area resolver returns area NAME from snapshot, version bumped to 2"
```

### Phase 3 Review (Opus)

Same pattern as Phase 1/2. Verify version bump documented + hash stability across area rename.

---

## Phase 4: D4 + D5 + D6 — PDF dispatch outbox (codex)

### Task 4.1: Migration 0176 — outbox table

**File:** `migrations/0176_pdf_dispatch_outbox.sql`

- [ ] **Step 1: Write migration.**

```sql
-- 0176_pdf_dispatch_outbox.sql
CREATE TABLE IF NOT EXISTS metaldocs.pdf_dispatch_outbox (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    revision_id     UUID NOT NULL,
    content_hash    BYTEA NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','processing','dispatched','failed')),
    attempts        INT  NOT NULL DEFAULT 0,
    last_error      TEXT,
    claimed_at      TIMESTAMPTZ,
    next_retry_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    dispatched_at   TIMESTAMPTZ,
    CONSTRAINT ux_pdf_dispatch_outbox_revision UNIQUE (tenant_id, revision_id)
);

CREATE INDEX IF NOT EXISTS ix_pdf_dispatch_outbox_pending
    ON metaldocs.pdf_dispatch_outbox (next_retry_at)
 WHERE status IN ('pending','processing');
```

- [ ] **Step 2: Apply + verify.**

- [ ] **Step 3: Commit.**

```powershell
git commit -m "feat(D4): migration 0176 pdf_dispatch_outbox table"
```

### Task 4.2: Outbox repository

**Files:**
- Create: `internal/modules/render/fanout/pdf_outbox_repository.go`
- Create: `internal/modules/render/fanout/pdf_outbox_repository_test.go`

- [ ] **Step 1: Write failing tests.** Six tests:

```go
package fanout

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPDFOutboxRepository_Enqueue_UsesTx(t *testing.T) { /* ExpectExec INSERT inside tx */ }
func TestPDFOutboxRepository_Enqueue_Idempotent(t *testing.T) { /* ON CONFLICT DO NOTHING */ }
func TestPDFOutboxRepository_ClaimPending_SkipLocks(t *testing.T) { /* SELECT ... FOR UPDATE SKIP LOCKED returns rows, sets status=processing */ }
func TestPDFOutboxRepository_MarkDispatched(t *testing.T) { /* UPDATE status=dispatched */ }
func TestPDFOutboxRepository_MarkFailed_AppliesBackoff(t *testing.T) { /* attempts++, next_retry_at advance */ }
func TestPDFOutboxRepository_ResetStaleClaims(t *testing.T) { /* UPDATE WHERE claimed_at < threshold */ }
```

(Each test: sqlmock.New, set expectations, call method, assert.ExpectationsWereMet.)

- [ ] **Step 2: Run — FAIL** (no impl).

- [ ] **Step 3: Implement.** `pdf_outbox_repository.go`:

```go
package fanout

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"metaldocs/internal/modules/documents/repository"
)

type OutboxRow struct {
	ID           string
	TenantID     string
	RevisionID   string
	ContentHash  []byte
	Attempts     int
}

type PDFOutboxRepository struct{ db *sql.DB }

func NewPDFOutboxRepository(db *sql.DB) *PDFOutboxRepository {
	return &PDFOutboxRepository{db: db}
}

func (r *PDFOutboxRepository) Enqueue(ctx context.Context, tx repository.DBTX, tenantID, revisionID string, contentHash []byte) error {
	exec := repository.DBTX(r.db)
	if tx != nil {
		exec = tx
	}
	_, err := exec.ExecContext(ctx, `
INSERT INTO metaldocs.pdf_dispatch_outbox (tenant_id, revision_id, content_hash)
VALUES ($1::uuid, $2::uuid, $3)
ON CONFLICT (tenant_id, revision_id) DO NOTHING`,
		tenantID, revisionID, contentHash)
	if err != nil {
		return fmt.Errorf("pdf outbox enqueue: %w", err)
	}
	return nil
}

func (r *PDFOutboxRepository) ClaimPending(ctx context.Context, limit int) ([]OutboxRow, error) {
	rows, err := r.db.QueryContext(ctx, `
WITH claimed AS (
  SELECT id FROM metaldocs.pdf_dispatch_outbox
   WHERE status = 'pending' AND next_retry_at <= NOW() AND attempts < 5
   ORDER BY next_retry_at ASC
   LIMIT $1
   FOR UPDATE SKIP LOCKED
)
UPDATE metaldocs.pdf_dispatch_outbox o
   SET status='processing', claimed_at=NOW()
  FROM claimed c
 WHERE o.id = c.id
RETURNING o.id, o.tenant_id, o.revision_id, o.content_hash, o.attempts`, limit)
	if err != nil {
		return nil, fmt.Errorf("claim pending: %w", err)
	}
	defer rows.Close()
	var out []OutboxRow
	for rows.Next() {
		var r OutboxRow
		if err := rows.Scan(&r.ID, &r.TenantID, &r.RevisionID, &r.ContentHash, &r.Attempts); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (r *PDFOutboxRepository) MarkDispatched(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.pdf_dispatch_outbox
   SET status='dispatched', dispatched_at=NOW()
 WHERE id=$1::uuid`, id)
	return err
}

func (r *PDFOutboxRepository) MarkFailed(ctx context.Context, id string, errStr string, nextRetryAt time.Time, finalize bool) error {
	if finalize {
		_, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.pdf_dispatch_outbox
   SET status='failed', last_error=$2, attempts=attempts+1
 WHERE id=$1::uuid`, id, errStr)
		return err
	}
	_, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.pdf_dispatch_outbox
   SET status='pending', last_error=$2, attempts=attempts+1, next_retry_at=$3, claimed_at=NULL
 WHERE id=$1::uuid`, id, errStr, nextRetryAt)
	return err
}

func (r *PDFOutboxRepository) ResetStaleClaims(ctx context.Context, olderThan time.Duration) (int, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.pdf_dispatch_outbox
   SET status='pending', claimed_at=NULL
 WHERE status='processing' AND claimed_at < NOW() - $1::interval`,
		fmt.Sprintf("%d milliseconds", olderThan.Milliseconds()))
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
```

- [ ] **Step 4: Run — PASS.**

```powershell
go test -mod=mod ./internal/modules/render/fanout/ -run TestPDFOutboxRepository_ -v
```

- [ ] **Step 5: Commit.**

```powershell
git commit -m "feat(D4): PDFOutboxRepository — enqueue/claim/mark with FOR UPDATE SKIP LOCKED"
```

### Task 4.3: Outbox worker

**Files:**
- Create: `internal/modules/render/fanout/pdf_outbox_worker.go`
- Create: `internal/modules/render/fanout/pdf_outbox_worker_test.go`

- [ ] **Step 1: Write failing tests** for worker behavior:

```go
func TestPDFOutboxWorker_PublishSuccessMarksDispatched(t *testing.T) { /* fake repo + fake publisher; one row claimed → publish OK → MarkDispatched called */ }
func TestPDFOutboxWorker_PublishFailIncrementsAttemptsWithBackoff(t *testing.T) { /* publisher returns err → MarkFailed(finalize=false), nextRetryAt > now */ }
func TestPDFOutboxWorker_MaxAttemptsMarksFinal(t *testing.T) { /* attempts=4 + fail → MarkFailed(finalize=true) */ }
func TestPDFOutboxWorker_StopOnContext(t *testing.T) { /* cancel ctx → Run returns nil */ }
```

- [ ] **Step 2: Implement worker.**

```go
package fanout

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"time"

	"metaldocs/internal/platform/messaging"
)

type outboxRepoAPI interface {
	ClaimPending(ctx context.Context, limit int) ([]OutboxRow, error)
	MarkDispatched(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id, errStr string, nextRetryAt time.Time, finalize bool) error
	ResetStaleClaims(ctx context.Context, olderThan time.Duration) (int, error)
}

type PDFOutboxWorker struct {
	repo       outboxRepoAPI
	pub        messaging.Publisher
	pollEvery  time.Duration
	batchSize  int
	maxAttempt int
	staleAfter time.Duration
	log        *slog.Logger
}

func NewPDFOutboxWorker(repo outboxRepoAPI, pub messaging.Publisher, log *slog.Logger) *PDFOutboxWorker {
	return &PDFOutboxWorker{
		repo: repo, pub: pub,
		pollEvery: 5 * time.Second, batchSize: 10,
		maxAttempt: 5, staleAfter: 5 * time.Minute,
		log: log,
	}
}

func (w *PDFOutboxWorker) Run(ctx context.Context) error {
	t := time.NewTicker(w.pollEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			w.tick(ctx)
		}
	}
}

func (w *PDFOutboxWorker) tick(ctx context.Context) {
	if _, err := w.repo.ResetStaleClaims(ctx, w.staleAfter); err != nil {
		w.log.Warn("reset stale claims", "err", err)
	}
	rows, err := w.repo.ClaimPending(ctx, w.batchSize)
	if err != nil {
		w.log.Warn("claim pending", "err", err)
		return
	}
	for _, r := range rows {
		w.dispatchOne(ctx, r)
	}
}

func (w *PDFOutboxWorker) dispatchOne(ctx context.Context, r OutboxRow) {
	err := w.pub.Publish(ctx, messaging.Event{
		EventType:      "docgen_v2_pdf",
		AggregateType:  "document_revision",
		AggregateID:    r.RevisionID,
		IdempotencyKey: "docgen_v2_pdf:" + r.TenantID + ":" + r.RevisionID,
		Payload: map[string]any{
			"tenant_id":    r.TenantID,
			"revision_id":  r.RevisionID,
			"content_hash": r.ContentHash,
		},
	})
	if err == nil {
		if mErr := w.repo.MarkDispatched(ctx, r.ID); mErr != nil {
			w.log.Error("mark dispatched", "id", r.ID, "err", mErr)
		}
		return
	}
	finalize := r.Attempts+1 >= w.maxAttempt
	backoff := time.Duration(math.Min(float64(30*time.Minute), float64(time.Duration(1<<r.Attempts)*30*time.Second)))
	nextRetry := time.Now().Add(backoff)
	if mErr := w.repo.MarkFailed(ctx, r.ID, err.Error(), nextRetry, finalize); mErr != nil {
		w.log.Error("mark failed", "id", r.ID, "err", mErr)
	}
	if finalize {
		w.log.Error("pdf dispatch permanently failed", "id", r.ID, "revision_id", r.RevisionID, "err", err)
	}
}

var ErrPublish = errors.New("publish failed")
```

- [ ] **Step 3: Run worker tests — PASS.**

- [ ] **Step 4: Commit.**

```powershell
git commit -m "feat(D6): PDFOutboxWorker — bounded retry, backoff, structured logging"
```

### Task 4.4: Wire outbox into decision_service

**File:** `internal/modules/documents/approval/application/decision_service.go`

- [ ] **Step 1: Add interface field.**

```go
type PDFOutboxEnqueuer interface {
	Enqueue(ctx context.Context, tx repository.DBTX, tenantID, revisionID string, contentHash []byte) error
}

type DecisionService struct {
	// ... existing fields ...
	pdfOutbox PDFOutboxEnqueuer // replaces pdfDispatcher for production path
}
```

- [ ] **Step 2: Failing integration test** — outbox row exists post-commit, none if tx fails.

```go
func TestSignoffApproval_OutboxRowEnqueuedInSameTx(t *testing.T) {
	// approve → query pdf_dispatch_outbox WHERE revision_id=...; expect 1 row, status='pending'
}
func TestSignoffApproval_TxRollback_NoOutboxRow(t *testing.T) {
	// force WriteFinalDocx error → expect 0 rows in outbox
}
```

- [ ] **Step 3: Replace post-commit dispatch.** Find lines 354-356 (`if shouldDispatchPDF && s.pdfDispatcher != nil { _ = s.pdfDispatcher.Dispatch(ctx, ...) }`).

Move enqueue INTO the tx, BEFORE `tx.Commit()`. Right after WriteFinalDocx (around line 230 area depending on file state):

```go
// Enqueue PDF dispatch in the same tx — outbox guarantees durability.
if shouldDispatchPDF && s.pdfOutbox != nil {
    if err := s.pdfOutbox.Enqueue(ctx, tx, pdfTenantID, pdfRevisionID, contentHash); err != nil {
        _ = tx.Rollback()
        return SignoffResult{}, fmt.Errorf("recordSignoff: enqueue pdf outbox: %w", err)
    }
}
```

Remove the post-commit `_ = s.pdfDispatcher.Dispatch(...)` block entirely.

- [ ] **Step 4: Update DecisionService constructor + tests.** All test files reference `pdfDispatcher: &fakePDFDispatchInvoker{}` — keep field for backwards compat OR add `pdfOutbox: &fakePDFOutbox{}` alongside. Recommend: keep `pdfDispatcher` field and zero it at construction; new path uses `pdfOutbox` only. Mark `pdfDispatcher` as deprecated in doc comment.

- [ ] **Step 5: Run all approval tests + new integration tests — PASS.**

```powershell
go test -mod=mod ./internal/modules/documents/approval/... -v
go test -mod=mod -tags=integration ./tests/integration/freeze/ -v
```

- [ ] **Step 6: Commit.**

```powershell
git commit -m "fix(D4/D5/D6): replace post-commit pdfDispatcher with transactional outbox enqueue"
```

### Task 4.5: Bootstrap wiring

**File:** `internal/platform/bootstrap/api.go`

- [ ] **Step 1: Wire repo + worker.** In `NewAPI` (or equivalent):

```go
pdfOutboxRepo := fanout.NewPDFOutboxRepository(db)
pdfOutboxWorker := fanout.NewPDFOutboxWorker(pdfOutboxRepo, eventPublisher, logger)
go func() {
    if err := pdfOutboxWorker.Run(ctx); err != nil {
        logger.Error("pdf outbox worker exited", "err", err)
    }
}()
```

Pass `pdfOutboxRepo` into `DecisionService` constructor.

- [ ] **Step 2: Manual smoke.**

```powershell
.\scripts\start-api.ps1 -Build
```

Submit doc → approve → check outbox table:

```sql
SELECT status, attempts, dispatched_at FROM metaldocs.pdf_dispatch_outbox ORDER BY created_at DESC LIMIT 5;
```

Expected: row appears with `status='pending'`, then `'dispatched'` within 5-10s.

- [ ] **Step 3: Commit.**

```powershell
git commit -m "feat(D4): wire PDFOutboxRepository + Worker in bootstrap"
```

### Phase 4 Review (Opus)

- [ ] **Step 1: Dispatch opus.**

Prompt:
```
review phase 4 — pdf dispatch outbox.

verify:
1. enqueue inside tx (tx.Rollback discards row)
2. unique (tenant_id, revision_id) prevents duplicate
3. FOR UPDATE SKIP LOCKED viable for multi-replica
4. backoff bounded (max 30min)
5. terminal status='failed' after 5 attempts
6. structured logging on permanent failure
7. ResetStaleClaims handles crashed workers
8. metric instrumentation present (or flag if missing)

PASS or list issues. under 250 words.
```

---

## Phase 5: Verify

### Task 5.1: Full test suite

- [ ] **Step 1: Unit + module tests.**

```powershell
go test -mod=mod ./...
```

Expected: ALL PASS, no FAIL.

- [ ] **Step 2: Integration tests.**

```powershell
go test -mod=mod -tags=integration ./tests/integration/...
```

Expected: ALL PASS.

- [ ] **Step 3: Vet + lint.**

```powershell
go vet -mod=mod ./...
```

Expected: zero warnings.

### Task 5.2: Codex independent audit

- [ ] **Step 1: Dispatch codex audit subagent.**

Prompt (caveman):
```
audit group D fixes on branch group-d-freeze-quality vs main.

git log main..HEAD --oneline

per bug D1-D8: PASS/FAIL with file:line evidence.

D1: actor_display_name_snapshot + created_by_display_name_snapshot written in same tx as signoff/create. resolvers read snapshot columns not live IAM.
D2: documents.area_name_snapshot column + populated at create. controlled_by_area resolver returns name.
D3: ErrEffectiveDateMissing typed error. zero time.Time → return error.
D4: pdf_dispatch_outbox table. enqueue inside approval tx. worker poll loop.
D5: outbox row keyed by (tenant_id, revision_id) — distinct per revision.
D6: MarkFailed records last_error + attempts. structured log on permanent failure.
D7: UpsertValue accepts DBTX variadic. freeze_service passes tx.
D8: ApproversResolver filters by approval_instance_id.

8/8 PASS expected. report under 400 words.
```

Expected: 8/8 PASS. If FAIL: fix before merge.

### Task 5.3: Smoke test

- [ ] **Step 1: Bootstrap fresh dev DB + run smoke.**

```powershell
.\scripts\start-api.ps1 -Build
```

- [ ] **Step 2:** Login as approver (`POST /api/v1/auth/login {identifier:"approver", password:"<dev>"}`).

- [ ] **Step 3:** Create document → submit → approve (multi-stage if applicable) → freeze.

- [ ] **Step 4:** Verify frozen DOCX placeholder values via inspection endpoint or DB:

```sql
SELECT placeholder_id, value_text FROM metaldocs.document_placeholder_values WHERE revision_id='<rev>' ORDER BY placeholder_id;
```

Expected: author/approver fields show display NAMES (not UUIDs). `controlled_by_area` shows area NAME. `effective_date` deterministic ISO date.

- [ ] **Step 5:** Force NULL effective_from on a draft, attempt freeze → expect 422.

- [ ] **Step 6:** Verify outbox dispatch within 10s post-approval.

```sql
SELECT status, dispatched_at - created_at AS latency FROM metaldocs.pdf_dispatch_outbox ORDER BY created_at DESC LIMIT 1;
```

Expected: `dispatched`, latency < 10s.

---

## Phase 6: Wiki + Finishing

### Task 6.1: Update audit doc

**File:** `wiki/bugs/audit-2026-05-03.md`

- [ ] **Step 1: Mark D1-D8 fixed with commit SHAs.** Replace open status with `fixed` + commit SHA.

- [ ] **Step 2: Add session history row.**

```markdown
| 2026-05-03 | Group D freeze quality | Fixed D1 (snapshot at signoff/create — <SHA1>), D2 (area_name_snapshot — <SHA2>), D3 (ErrEffectiveDateMissing — <SHA3>), D4/D5/D6 (pdf_dispatch_outbox — <SHA4..6>), D7 (UpsertValue tx variadic — <SHA7>), D8 (approvers instance filter — <SHA8>). All unit/integration tests pass. Codex audit 8/8 PASS. |
```

- [ ] **Step 3: Update progress summary table** (Group D 8/8/0/0).

- [ ] **Step 4: Update scope line** at top.

- [ ] **Step 5: Commit.**

```powershell
git commit -m "docs(wiki): mark Group D bugs fixed with commit SHAs"
```

### Task 6.2: Wiki-curator full pass

- [ ] **Step 1: Dispatch wiki-curator agent (haiku).**

Prompt (caveman):
```
group D freeze quality closed. update wiki for drift.

recent commits — git log main..HEAD --oneline

verify + fix:
1. wiki/concepts/freeze-pipeline.md — anchors valid? Last verified: 2026-05-03? mention outbox?
2. wiki/concepts/placeholders.md — controlled_by_area now returns NAME (was CODE)
3. wiki/modules/render-* — file:line anchors after resolver edits
4. wiki/decisions/ — add ADR 0008 pdf-dispatch-outbox if pattern significant
5. wiki/README.md — index new ADR + any new docs
6. wiki/references/local-dev-credentials.md — no change needed

create wiki/decisions/0008-pdf-dispatch-outbox.md (~80 lines) summarizing: problem (post-commit dispatch loses on crash), decision (transactional outbox + worker), consequences (worker process required, latency 5-10s instead of immediate).

report changes in under 200 words.
```

- [ ] **Step 2: Review wiki-curator diff, commit.**

### Task 6.3: Final Opus review

- [ ] **Step 1: Dispatch opus full-branch review.**

Prompt:
```
final review group-d-freeze-quality before merge.

git diff main..HEAD --stat
git log main..HEAD --oneline

verify:
1. spec coverage 100% (8/8 bugs addressed in code, not just tests)
2. no scope creep beyond spec
3. tests assert behavior not implementation detail
4. migrations idempotent + ordered
5. backwards-compat path for legacy documents (NULL snapshot columns)
6. outbox worker started by bootstrap, stopped on shutdown ctx
7. structured errors propagate (ErrEffectiveDateMissing surfaces as 422 not 500)
8. wiki updated, audit doc updated

PASS for merge OR list blocking issues. under 300 words.
```

- [ ] **Step 2: Address blockers if any. Re-review.**

### Task 6.4: Finishing branch

- [ ] **Step 1: Invoke superpowers:finishing-a-development-branch.** Verify tests pass → present 4 options to user → execute chosen path.

Recommended option: **1. Merge back to main locally** — Group D is internal cleanup, no PR needed.

- [ ] **Step 2: After merge, cleanup worktree.**

```powershell
git worktree remove ../MetalDocs-group-d
```

---

## Done Definition

- All 8 bugs (D1-D8) fixed and verified
- `go test -mod=mod ./...` PASS
- `go test -mod=mod -tags=integration ./...` PASS
- Codex audit 8/8 PASS with file:line evidence
- Smoke test PASS (display names not UUIDs, area name, deterministic effective_date, 422 on NULL, outbox dispatch < 10s)
- `wiki/bugs/audit-2026-05-03.md` updated with commit SHAs
- ADR 0008 (pdf-dispatch-outbox) committed
- Wiki stamps refreshed on freeze/render docs
- Branch merged to main, worktree cleaned

---

## Self-Review

**Spec coverage:** All 8 bugs in spec mapped to tasks: D1 → 2.1-2.5, D2 → 3.1-3.3, D3 → 1.2, D4 → 4.1-4.5, D5 → 4.2 (column naming), D6 → 4.3 (worker), D7 → 1.1, D8 → 1.3. Acceptance criteria from spec all reflected in Done definition.

**Placeholder scan:** No "TBD"/"TODO"/"implement later". Code blocks complete. Test bodies concrete (sqlmock + integration patterns). Test stubs in Task 4.4-4.5 reference fake types — fakes are conventional in this codebase (already used in `decision_service_test.go`).

**Type consistency:** `DBTX` used consistently (defined `internal/modules/documents/repository/snapshot_repository.go:19`, reused everywhere). `OutboxRow` fields stable across repo + worker. `MarkFailed(id, errStr, nextRetryAt, finalize bool)` signature consistent between Task 4.2 (impl) and 4.3 (worker call). `ResolveInput.AreaNameSnapshot` + `ApprovalInstanceID` defined Phase 1 (Task 1.3) and used Phase 3.

**Migration ordering:** 0173 → 0174 (Phase 2) → 0175 (Phase 3) → 0176 (Phase 4). Each idempotent (`IF NOT EXISTS` / `ON CONFLICT DO NOTHING`).

**Phase ordering:** Phase 1 parallel safe (no shared files). Phase 2 introduces snapshot columns referenced by Phase 1 SQL — Phase 1 SQL uses `COALESCE` so missing column at Phase 1 commit time would only break runtime, not compile. Acceptable: Phase 1 + 2 verified together at end of Phase 2 review.
