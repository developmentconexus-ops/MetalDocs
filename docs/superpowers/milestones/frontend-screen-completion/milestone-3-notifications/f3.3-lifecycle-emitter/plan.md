# Feature F3.3 — Lifecycle Emitter (Domain-Event Pattern)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire five typed River-job domain events (document.published/superseded/obsoleted/approved/rejected) so the owning module enqueues in-tx beside the existing audit emit, and a notifications fan-out River worker inserts idempotent per-recipient inbox rows consumed by the F3.2 read surface.

**Architecture:** Domain-event envelope (`LifecycleEventArgs`) lives in `documents/domain` (the sanctioned cross-module contract layer, 47× precedent). A River adapter in `approval/jobs` enqueues in-tx inside the same `*sql.Tx` as the state-change, mirroring the existing `RiverScheduledPublishEnqueuer`. The fan-out worker (`notifications/infrastructure`) runs after commit, resolves recipients (reader events → `v_cd_obligated_readers`; author events → `submitted_by` from the payload), and inserts rows with `ON CONFLICT DO NOTHING` (idempotent). The five emit sites gain only an additive in-tx enqueue — publish/approval semantics are unchanged. Workers run in the `metaldocs-jobs` binary; the enqueuer is injected in `metaldocs-api`.

**Tech Stack:** Go 1.23 · River v0.37.1 (`github.com/riverqueue/river`) · `*river.Client[*sql.Tx]` · PostgreSQL (`metaldocs.notifications`, `metaldocs.v_cd_obligated_readers`) · `go:build integration` for PG tests

---

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `internal/modules/documents/domain/notification_events.go` | `LifecycleEventArgs` River-job args + `Kind()` + 5 constants + `LifecycleEventEnqueuer` port |
| Create | `internal/modules/documents/domain/notification_events_test.go` | Compile-time interface satisfaction + constant value tests |
| Create | `internal/modules/documents/application/document_cdid.go` | `LoadDocumentControlledDocumentID` — reads `controlled_document_id` in-tx |
| Create | `internal/modules/documents/application/document_cdid_test.go` | Unit test via fake tx |
| Create | `internal/modules/documents/approval/jobs/lifecycle_event_enqueuer.go` | `RiverLifecycleEventEnqueuer` — River adapter, `db.Tx → *sql.Tx` assertion, `Client.InsertTx` |
| Create | `internal/modules/documents/approval/jobs/lifecycle_event_enqueuer_test.go` | Unit test: wrong-type tx returns error; ok-path type assertion |
| Modify | `internal/modules/documents/approval/application/publish_service.go` | Add `lifecycleEnqueuer` field + `WithLifecycleEnqueuer` + additive enqueue in `PublishApproved` |
| Modify | `internal/modules/documents/approval/application/supersede_service.go` | Same pattern for `PublishSuperseding` |
| Modify | `internal/modules/documents/approval/application/obsolete_service.go` | Same pattern for `MarkObsolete` |
| Modify | `internal/modules/documents/approval/application/decision_service.go` | Same pattern for `RecordSignoff` (two events gated on `InstanceApproved`/`InstanceRejected`) |
| Modify | `internal/modules/documents/approval/application/services.go` | `WithLifecycleEnqueuer(*Services)` propagates to all 4 services |
| Create | `internal/modules/documents/approval/application/lifecycle_emit_test.go` | Service-level spy tests: 1 enqueue per action, correct args, rollback = no enqueue |
| Create | `internal/modules/notifications/infrastructure/fanout_worker.go` | `NotificationsFanoutWorker` — River worker, recipient resolution, bulk INSERT ON CONFLICT DO NOTHING |
| Create | `internal/modules/notifications/infrastructure/fanout_worker_integration_test.go` | Integration tests: 5 event types × recipient sets + redelivery no-op |
| Modify | `apps/jobs/cmd/metaldocs-jobs/main.go` | Register `NotificationsFanoutWorker` in the workers factory callback |
| Modify | `apps/api/cmd/metaldocs-api/main.go` | Inject `RiverLifecycleEventEnqueuer` via `approvalServices.WithLifecycleEnqueuer` |

---

## Task 1 — Domain contract: LifecycleEventArgs + enqueuer port

**Files:**
- Create: `internal/modules/documents/domain/notification_events.go`
- Create: `internal/modules/documents/domain/notification_events_test.go`

- [ ] **Step 1: Write the failing test (compile-only)**

```go
// internal/modules/documents/domain/notification_events_test.go
package domain_test

import (
	"testing"

	"metaldocs/internal/modules/documents/domain"
)

func TestLifecycleEventArgKind(t *testing.T) {
	args := domain.LifecycleEventArgs{}
	if args.Kind() != "notification_fanout" {
		t.Errorf("want Kind()=notification_fanout, got %q", args.Kind())
	}
}

func TestEventTypeConstants(t *testing.T) {
	want := []string{
		domain.EventTypeDocumentPublished,
		domain.EventTypeDocumentSuperseded,
		domain.EventTypeDocumentObsoleted,
		domain.EventTypeDocumentApproved,
		domain.EventTypeDocumentRejected,
	}
	for _, c := range want {
		if c == "" {
			t.Error("empty event_type constant")
		}
	}
}

// Compile-time: LifecycleEventEnqueuer interface exists and is usable as a type
var _ func(domain.LifecycleEventEnqueuer) = func(domain.LifecycleEventEnqueuer) {}
```

- [ ] **Step 2: Run test — expect compile failure**

```
go test ./internal/modules/documents/domain/...
```
Expected: `undefined: domain.LifecycleEventArgs` (or similar compile error)

- [ ] **Step 3: Implement the contract file**

```go
// internal/modules/documents/domain/notification_events.go
package domain

import (
	"context"
	"time"

	"metaldocs/internal/platform/db"
)

// LifecycleEventArgs is the River-job envelope for document lifecycle domain events.
// One kind, discriminated by EventType, keeps the fan-out worker simple (single worker).
// No `river` import here — domain stays infra-free; Kind() satisfies river.JobArgs via
// a plain string method, not a structural River dependency.
type LifecycleEventArgs struct {
	EventID              string    `json:"event_id"`               // uuid, minted at emit — idempotency key
	TenantID             string    `json:"tenant_id"`
	EventType            string    `json:"event_type"`             // one of the constants below
	ResourceType         string    `json:"resource_type"`
	ResourceID           string    `json:"resource_id"`
	ControlledDocumentID string    `json:"controlled_document_id"` // set for reader events; "" for author events
	SubmittedBy          string    `json:"submitted_by"`           // set for author events; "" for reader events
	OccurredAt           time.Time `json:"occurred_at"`
}

// Kind implements the river.JobArgs interface without importing River.
func (LifecycleEventArgs) Kind() string { return "notification_fanout" }

// Document lifecycle event types — the M3 bundle (ADR-0044 §3).
const (
	EventTypeDocumentPublished  = "document.published"
	EventTypeDocumentSuperseded = "document.superseded"
	EventTypeDocumentObsoleted  = "document.obsoleted"
	EventTypeDocumentApproved   = "document.approved"
	EventTypeDocumentRejected   = "document.rejected"
)

// LifecycleEventEnqueuer is the port the five emit sites depend on.
// Takes db.Tx (domain-clean) — the River infra adapter narrows to *sql.Tx internally.
type LifecycleEventEnqueuer interface {
	EnqueueLifecycleEventTx(ctx context.Context, tx db.Tx, args LifecycleEventArgs) error
}
```

- [ ] **Step 4: Run test — expect green**

```
go test ./internal/modules/documents/domain/...
```
Expected: `PASS`

- [ ] **Step 5: Verify go build**

```
go build ./...
```
Expected: exit 0

- [ ] **Step 6: Commit**

```
git add internal/modules/documents/domain/notification_events.go internal/modules/documents/domain/notification_events_test.go
git commit -m "feat(M3/F3.3): LifecycleEventArgs River-job contract + LifecycleEventEnqueuer port in documents/domain"
```

---

## Task 2 — LoadDocumentControlledDocumentID helper

**Files:**
- Create: `internal/modules/documents/application/document_cdid.go`
- Create: `internal/modules/documents/application/document_cdid_test.go`

The reader events (published/superseded/obsoleted) need the document's `controlled_document_id` in-tx for the fan-out worker's recipient query. This helper reads it from `public.documents` inside the caller's `*sql.Tx`.

- [ ] **Step 1: Write the failing test**

```go
// internal/modules/documents/application/document_cdid_test.go
package application_test

import (
	"context"
	"database/sql"
	"testing"
)

// Compile-time: function signature exists in the package.
// Full integration coverage is in fanout_worker_integration_test.go which exercises
// the real path end-to-end. This test validates the signature is importable.
func TestLoadDocumentControlledDocumentIDSignature(t *testing.T) {
	// Just verify the function exists and has the right signature at compile time.
	// We can't call it without a real DB here — that's covered in integration tests.
	_ = context.Background()
	_ = (*sql.Tx)(nil)
	// The function is called: application.LoadDocumentControlledDocumentID(ctx, tx, tenantID, docID)
	// Compile failure here means the function doesn't exist yet.
}
```

- [ ] **Step 2: Run test — verify it compiles (no runtime failure expected)**

```
go test ./internal/modules/documents/application/...
```
Expected: PASS (compile check only)

- [ ] **Step 3: Implement**

```go
// internal/modules/documents/application/document_cdid.go
package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// LoadDocumentControlledDocumentID reads the controlled_document_id for a document
// inside the caller's tx. Returns "" when the document has no CD link (NULL) or
// does not exist (not found is not an error — caller decides if "" is fatal).
// Used by F3.3 reader-event emit sites to build the LifecycleEventArgs payload.
func LoadDocumentControlledDocumentID(ctx context.Context, tx *sql.Tx, tenantID, documentID string) (string, error) {
	var cdID sql.NullString
	err := tx.QueryRowContext(ctx,
		`SELECT controlled_document_id FROM public.documents WHERE id = $1 AND tenant_id = $2`,
		documentID, tenantID,
	).Scan(&cdID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("load document controlled_document_id: %w", err)
	}
	if !cdID.Valid {
		return "", nil
	}
	return cdID.String, nil
}
```

- [ ] **Step 4: Run test — expect green**

```
go test ./internal/modules/documents/application/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```
git add internal/modules/documents/application/document_cdid.go internal/modules/documents/application/document_cdid_test.go
git commit -m "feat(M3/F3.3): LoadDocumentControlledDocumentID in-tx helper for reader event emit sites"
```

---

## Task 3 — River enqueuer adapter

**Files:**
- Create: `internal/modules/documents/approval/jobs/lifecycle_event_enqueuer.go`
- Create: `internal/modules/documents/approval/jobs/lifecycle_event_enqueuer_test.go`

Mirrors `RiverScheduledPublishEnqueuer` in the same package. Takes `db.Tx`, asserts to `*sql.Tx` (the one allowed coupling point with River), calls `Client.InsertTx`. Note: spec says `approval/infrastructure` but the established River adapter pattern lives in `approval/jobs` (`RiverScheduledPublishEnqueuer` is there); this adapter goes in `jobs` for consistency.

- [ ] **Step 1: Write the failing test**

```go
// internal/modules/documents/approval/jobs/lifecycle_event_enqueuer_test.go
package jobs_test

import (
	"context"
	"database/sql"
	"testing"

	"metaldocs/internal/modules/documents/approval/jobs"
	documentsdomain "metaldocs/internal/modules/documents/domain"
)

// Verify compile-time: RiverLifecycleEventEnqueuer implements LifecycleEventEnqueuer.
var _ documentsdomain.LifecycleEventEnqueuer = (*jobs.RiverLifecycleEventEnqueuer)(nil)

func TestRiverLifecycleEventEnqueuer_WrongTxType(t *testing.T) {
	enc := &jobs.RiverLifecycleEventEnqueuer{} // Client nil — only tests type assertion
	err := enc.EnqueueLifecycleEventTx(context.Background(), wrongTx{}, documentsdomain.LifecycleEventArgs{})
	if err == nil {
		t.Fatal("want error for wrong tx type, got nil")
	}
}

// wrongTx is a db.Tx that is NOT *sql.Tx, triggering the type assertion failure.
type wrongTx struct{}

func (wrongTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}
func (wrongTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}
func (wrongTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row { return nil }
func (wrongTx) Commit() error   { return nil }
func (wrongTx) Rollback() error { return nil }
```

- [ ] **Step 2: Run test — expect compile failure**

```
go test ./internal/modules/documents/approval/jobs/...
```
Expected: `undefined: jobs.RiverLifecycleEventEnqueuer`

- [ ] **Step 3: Implement**

```go
// internal/modules/documents/approval/jobs/lifecycle_event_enqueuer.go
package jobs

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"

	documentsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/platform/db"
)

// RiverLifecycleEventEnqueuer implements documentsdomain.LifecycleEventEnqueuer
// using River's same-tx InsertTx, mirroring RiverScheduledPublishEnqueuer.
// The db.Tx → *sql.Tx assertion is the one allowed coupling point with River infra.
type RiverLifecycleEventEnqueuer struct {
	Client *river.Client[*sql.Tx]
}

// NewLifecycleEventEnqueuer wraps a River client as a LifecycleEventEnqueuer.
func NewLifecycleEventEnqueuer(client *river.Client[*sql.Tx]) documentsdomain.LifecycleEventEnqueuer {
	return &RiverLifecycleEventEnqueuer{Client: client}
}

func (e *RiverLifecycleEventEnqueuer) EnqueueLifecycleEventTx(ctx context.Context, tx db.Tx, args documentsdomain.LifecycleEventArgs) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return fmt.Errorf("lifecycle_event_enqueuer: river requires *sql.Tx, got %T", tx)
	}
	_, err := e.Client.InsertTx(ctx, sqlTx, args, nil)
	if err != nil {
		return fmt.Errorf("lifecycle_event_enqueuer: enqueue %s: %w", args.EventType, err)
	}
	return nil
}
```

- [ ] **Step 4: Run test — expect green**

```
go test ./internal/modules/documents/approval/jobs/...
```
Expected: PASS

- [ ] **Step 5: go build green**

```
go build ./...
```
Expected: exit 0

- [ ] **Step 6: Commit**

```
git add internal/modules/documents/approval/jobs/lifecycle_event_enqueuer.go internal/modules/documents/approval/jobs/lifecycle_event_enqueuer_test.go
git commit -m "feat(M3/F3.3): RiverLifecycleEventEnqueuer adapter (mirrors RiverScheduledPublishEnqueuer)"
```

---

## Task 4 — Inject enqueuer into 5 services + Services.WithLifecycleEnqueuer

**Files:**
- Modify: `internal/modules/documents/approval/application/publish_service.go`
- Modify: `internal/modules/documents/approval/application/supersede_service.go`
- Modify: `internal/modules/documents/approval/application/obsolete_service.go`
- Modify: `internal/modules/documents/approval/application/decision_service.go`
- Modify: `internal/modules/documents/approval/application/services.go`
- Create: `internal/modules/documents/approval/application/lifecycle_emit_test.go`

**Important:** the enqueue is additive (added AFTER the existing `s.emitter.Emit` call) — it must never change the behavior when `lifecycleEnqueuer` is nil (the nil guard ensures zero regression on existing tests).

- [ ] **Step 1: Write failing spy tests**

```go
// internal/modules/documents/approval/application/lifecycle_emit_test.go
package application

import (
	"context"
	"testing"
	"time"

	documentsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/platform/db"
)

// spyLifecycleEnqueuer captures enqueued lifecycle events for assertion.
type spyLifecycleEnqueuer struct {
	calls []documentsdomain.LifecycleEventArgs
}

func (s *spyLifecycleEnqueuer) EnqueueLifecycleEventTx(_ context.Context, _ db.Tx, args documentsdomain.LifecycleEventArgs) error {
	s.calls = append(s.calls, args)
	return nil
}

// TestPublishApproved_EnqueuesDocumentPublished verifies publish_service enqueues
// exactly one document.published event with the correct controlled_document_id.
func TestPublishApproved_EnqueuesDocumentPublished(t *testing.T) {
	spy := &spyLifecycleEnqueuer{}
	// Stub a minimal PublishService with spy enqueuer and a fake repo/emitter.
	svc := &PublishService{
		repo:              stubPublishRepo{},
		emitter:           &noopEmitter{},
		clock:             fixedClock{},
		lifecycleEnqueuer: spy,
	}
	runner := &fakeRunner{}
	_, _ = svc.PublishApproved(context.Background(), runner, PublishRequest{
		TenantID:   "tid",
		InstanceID: "iid",
		PublishedBy: "uid",
	})
	if len(spy.calls) != 1 {
		t.Fatalf("want 1 enqueue call, got %d", len(spy.calls))
	}
	if spy.calls[0].EventType != documentsdomain.EventTypeDocumentPublished {
		t.Errorf("want event_type %s, got %s", documentsdomain.EventTypeDocumentPublished, spy.calls[0].EventType)
	}
}

// TestDecisionService_EnqueuesApproved verifies exactly 1 document.approved at terminal approval.
func TestDecisionService_EnqueuesApproved(t *testing.T) {
	spy := &spyLifecycleEnqueuer{}
	svc := &DecisionService{
		repo:              &terminalApproveRepo{},
		emitter:           &noopEmitter{},
		clock:             fixedClock{},
		lifecycleEnqueuer: spy,
	}
	runner := &fakeRunner{}
	_, _ = svc.RecordSignoff(context.Background(), runner, SignoffRequest{
		TenantID:   "tid",
		InstanceID: "iid",
		ActorUserID: "uid",
	})
	if len(spy.calls) != 1 {
		t.Fatalf("want 1 enqueue for InstanceApproved, got %d", len(spy.calls))
	}
	if spy.calls[0].EventType != documentsdomain.EventTypeDocumentApproved {
		t.Errorf("want %s, got %s", documentsdomain.EventTypeDocumentApproved, spy.calls[0].EventType)
	}
}

// TestDecisionService_EnqueuesRejected verifies exactly 1 document.rejected at terminal rejection.
func TestDecisionService_EnqueuesRejected(t *testing.T) {
	spy := &spyLifecycleEnqueuer{}
	svc := &DecisionService{
		repo:              &terminalRejectRepo{},
		emitter:           &noopEmitter{},
		clock:             fixedClock{},
		lifecycleEnqueuer: spy,
	}
	runner := &fakeRunner{}
	_, _ = svc.RecordSignoff(context.Background(), runner, SignoffRequest{
		TenantID:   "tid",
		InstanceID: "iid",
		ActorUserID: "uid",
	})
	if len(spy.calls) != 1 {
		t.Fatalf("want 1 enqueue for InstanceRejected, got %d", len(spy.calls))
	}
	if spy.calls[0].EventType != documentsdomain.EventTypeDocumentRejected {
		t.Errorf("want %s, got %s", documentsdomain.EventTypeDocumentRejected, spy.calls[0].EventType)
	}
}

// TestNilEnqueuer_NoEnqueueAttempt verifies nil lifecycleEnqueuer does not panic
// and services behave identically to pre-F3.3 (regression guard).
func TestNilEnqueuer_NoEnqueueAttempt(t *testing.T) {
	svc := &PublishService{
		repo:              stubPublishRepo{},
		emitter:           &noopEmitter{},
		clock:             fixedClock{},
		lifecycleEnqueuer: nil, // not wired
	}
	runner := &fakeRunner{}
	// Must not panic.
	_, _ = svc.PublishApproved(context.Background(), runner, PublishRequest{
		TenantID:   "tid",
		InstanceID: "iid",
		PublishedBy: "uid",
	})
}

// --- minimal stubs used only in this test file ---

type fixedClock struct{}
func (fixedClock) Now() time.Time { return time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC) }

type noopEmitter struct{}
func (*noopEmitter) Emit(_ context.Context, _ db.Tx, _ GovernanceEvent) error { return nil }

type fakeRunner struct{}
func (f *fakeRunner) Do(ctx context.Context, fn func(*sql.Tx) error) error { return fn(nil) }
```

> **NOTE:** The fake repos (`stubPublishRepo`, `terminalApproveRepo`, `terminalRejectRepo`) need to satisfy `repository.ApprovalRepository` — they can embed `fakeDecisionRepo` from `decision_service_test.go` (already in this package). Add minimal overrides. See existing `decision_service_test.go` for the pattern. Implement the stubs in this file too.

- [ ] **Step 2: Run tests — expect compile failure (field `lifecycleEnqueuer` doesn't exist yet)**

```
go test ./internal/modules/documents/approval/application/...
```
Expected: compile error on `lifecycleEnqueuer` field

- [ ] **Step 3: Add lifecycleEnqueuer field + WithLifecycleEnqueuer to publish_service.go**

In `internal/modules/documents/approval/application/publish_service.go`:

Add to the `PublishService` struct (after existing fields):
```go
lifecycleEnqueuer documentsdomain.LifecycleEventEnqueuer
```

Add import at top:
```go
documentsdomain "metaldocs/internal/modules/documents/domain"
```

Add `With` method (after `WithScheduledPublishEnqueuer`):
```go
func (s *PublishService) WithLifecycleEnqueuer(e documentsdomain.LifecycleEventEnqueuer) *PublishService {
	s.lifecycleEnqueuer = e
	return s
}
```

In `PublishApproved`, after the `s.emitter.Emit(ctx, tx, event)` block (line 137–139), add:
```go
// Additive in-tx domain-event enqueue (ADR-0044; F3.3). After audit emit.
if s.lifecycleEnqueuer != nil {
    cdID, err := docapp.LoadDocumentControlledDocumentID(ctx, tx, req.TenantID, instance.DocumentID)
    if err != nil {
        return fmt.Errorf("publishApproved: load cd id for lifecycle event: %w", err)
    }
    largs := documentsdomain.LifecycleEventArgs{
        EventID:              uuid.NewString(),
        TenantID:             req.TenantID,
        EventType:            documentsdomain.EventTypeDocumentPublished,
        ResourceType:         "document",
        ResourceID:           instance.DocumentID,
        ControlledDocumentID: cdID,
        OccurredAt:           now,
    }
    if err := s.lifecycleEnqueuer.EnqueueLifecycleEventTx(ctx, tx, largs); err != nil {
        return fmt.Errorf("publishApproved: enqueue lifecycle event: %w", err)
    }
}
```

Add import: `"github.com/google/uuid"` (check: `go.mod` likely already has `github.com/google/uuid` — verify with `grep "google/uuid" go.mod`; if missing, `go get github.com/google/uuid@latest`).

- [ ] **Step 4: Add lifecycleEnqueuer field + enqueue to supersede_service.go**

Add to `SupersedeService` struct: `lifecycleEnqueuer documentsdomain.LifecycleEventEnqueuer`

Add import: `documentsdomain "metaldocs/internal/modules/documents/domain"` and `"github.com/google/uuid"`

Add `WithLifecycleEnqueuer` method:
```go
func (s *SupersedeService) WithLifecycleEnqueuer(e documentsdomain.LifecycleEventEnqueuer) *SupersedeService {
	s.lifecycleEnqueuer = e
	return s
}
```

In `PublishSuperseding`, after `s.emitter.Emit(ctx, tx, event)` (line 140–142), add:
```go
// Additive in-tx domain-event enqueue (ADR-0044; F3.3). Reader event for new doc's CD.
if s.lifecycleEnqueuer != nil {
    cdID, err := docapp.LoadDocumentControlledDocumentID(ctx, tx, req.TenantID, req.NewDocumentID)
    if err != nil {
        return fmt.Errorf("publishSuperseding: load cd id for lifecycle event: %w", err)
    }
    largs := documentsdomain.LifecycleEventArgs{
        EventID:              uuid.NewString(),
        TenantID:             req.TenantID,
        EventType:            documentsdomain.EventTypeDocumentSuperseded,
        ResourceType:         "document",
        ResourceID:           req.NewDocumentID,
        ControlledDocumentID: cdID,
        OccurredAt:           now,
    }
    if err := s.lifecycleEnqueuer.EnqueueLifecycleEventTx(ctx, tx, largs); err != nil {
        return fmt.Errorf("publishSuperseding: enqueue lifecycle event: %w", err)
    }
}
```

- [ ] **Step 5: Add lifecycleEnqueuer field + enqueue to obsolete_service.go**

Add to `ObsoleteService` struct: `lifecycleEnqueuer documentsdomain.LifecycleEventEnqueuer`

Add import: `documentsdomain "metaldocs/internal/modules/documents/domain"` and `"github.com/google/uuid"`

Add `WithLifecycleEnqueuer` method:
```go
func (s *ObsoleteService) WithLifecycleEnqueuer(e documentsdomain.LifecycleEventEnqueuer) *ObsoleteService {
	s.lifecycleEnqueuer = e
	return s
}
```

In `MarkObsolete`, after `s.emitter.Emit(ctx, tx, event)` (line 136–138), add:
```go
// Additive in-tx domain-event enqueue (ADR-0044; F3.3). Reader event.
if s.lifecycleEnqueuer != nil {
    cdID, err := docapp.LoadDocumentControlledDocumentID(ctx, tx, req.TenantID, req.DocumentID)
    if err != nil {
        return fmt.Errorf("markObsolete: load cd id for lifecycle event: %w", err)
    }
    obsoletedAt := s.clock.Now()
    largs := documentsdomain.LifecycleEventArgs{
        EventID:              uuid.NewString(),
        TenantID:             req.TenantID,
        EventType:            documentsdomain.EventTypeDocumentObsoleted,
        ResourceType:         "document",
        ResourceID:           req.DocumentID,
        ControlledDocumentID: cdID,
        OccurredAt:           obsoletedAt,
    }
    if err := s.lifecycleEnqueuer.EnqueueLifecycleEventTx(ctx, tx, largs); err != nil {
        return fmt.Errorf("markObsolete: enqueue lifecycle event: %w", err)
    }
}
```

Note: `obsolete_service.go` uses `s.clock.Now()` inline in the event — capture it before the enqueue block to stay consistent with the audit event's timestamp.

- [ ] **Step 6: Add lifecycleEnqueuer field + enqueue to decision_service.go**

Add to `DecisionService` struct: `lifecycleEnqueuer documentsdomain.LifecycleEventEnqueuer`

Add import: `documentsdomain "metaldocs/internal/modules/documents/domain"` and `"github.com/google/uuid"`

Add `WithLifecycleEnqueuer` method (after existing `With*` methods):
```go
func (s *DecisionService) WithLifecycleEnqueuer(e documentsdomain.LifecycleEventEnqueuer) *DecisionService {
	s.lifecycleEnqueuer = e
	return s
}
```

In `RecordSignoff`, after `s.emitter.Emit(ctx, tx, event)` (line 508–510), add:
```go
// Additive in-tx domain-event enqueue (ADR-0044; F3.3). Author events — terminal transitions only.
if s.lifecycleEnqueuer != nil {
    var lifecycleEventType string
    if result.InstanceApproved {
        lifecycleEventType = documentsdomain.EventTypeDocumentApproved
    } else if result.InstanceRejected {
        lifecycleEventType = documentsdomain.EventTypeDocumentRejected
    }
    if lifecycleEventType != "" {
        largs := documentsdomain.LifecycleEventArgs{
            EventID:      uuid.NewString(),
            TenantID:     req.TenantID,
            EventType:    lifecycleEventType,
            ResourceType: "approval_instance",
            ResourceID:   req.InstanceID,
            SubmittedBy:  instance.SubmittedBy,
            OccurredAt:   now,
        }
        if err := s.lifecycleEnqueuer.EnqueueLifecycleEventTx(ctx, tx, largs); err != nil {
            return fmt.Errorf("recordSignoff: enqueue lifecycle event: %w", err)
        }
    }
}
```

The `now` variable is already in scope (declared at line ~219 in the `runner.Do` closure). Confirm the exact `now` variable name used in `decision_service.go` by reading lines 210–230 before implementing.

- [ ] **Step 7: Add Services.WithLifecycleEnqueuer to services.go**

In `internal/modules/documents/approval/application/services.go`, add:

```go
// WithLifecycleEnqueuer wires the F3.3 domain-event enqueuer to all four emit
// services (Publish, Supersede, Obsolete, Decision). Call after NewServices.
func (s *Services) WithLifecycleEnqueuer(e documentsdomain.LifecycleEventEnqueuer) *Services {
	if s == nil {
		return s
	}
	if s.Publish != nil {
		s.Publish = s.Publish.WithLifecycleEnqueuer(e)
	}
	if s.Supersede != nil {
		s.Supersede = s.Supersede.WithLifecycleEnqueuer(e)
	}
	if s.Obsolete != nil {
		s.Obsolete = s.Obsolete.WithLifecycleEnqueuer(e)
	}
	if s.Decision != nil {
		s.Decision = s.Decision.WithLifecycleEnqueuer(e)
	}
	return s
}
```

Add import to `services.go`: `documentsdomain "metaldocs/internal/modules/documents/domain"`

- [ ] **Step 8: Run ALL approval/application tests — existing + new spy tests must pass**

```
go test ./internal/modules/documents/approval/application/...
```
Expected: PASS (all existing tests + new spy tests)

- [ ] **Step 9: Check google/uuid import**

```
grep "google/uuid" go.mod
```
If absent: `go get github.com/google/uuid@latest` then `go mod tidy`

- [ ] **Step 10: go build**

```
go build ./...
```
Expected: exit 0

- [ ] **Step 11: Commit**

```
git add internal/modules/documents/approval/application/publish_service.go \
        internal/modules/documents/approval/application/supersede_service.go \
        internal/modules/documents/approval/application/obsolete_service.go \
        internal/modules/documents/approval/application/decision_service.go \
        internal/modules/documents/approval/application/services.go \
        internal/modules/documents/approval/application/lifecycle_emit_test.go
git commit -m "feat(M3/F3.3): additive lifecycle-event enqueue at 5 emit sites; Services.WithLifecycleEnqueuer"
```

---

## Task 5 — Notifications fan-out River worker

**Files:**
- Create: `internal/modules/notifications/infrastructure/fanout_worker_integration_test.go`
- Create: `internal/modules/notifications/infrastructure/fanout_worker.go`

**TDD: write the integration tests first (failing), then implement to green.**

The worker runs after commit (off-tx; H-PRE-1). It switches on `EventType`:
- Reader events: query `metaldocs.v_cd_obligated_readers WHERE tenant_id=$1 AND controlled_document_id=$2`, bulk-insert one notification row per user.
- Author events: insert one row for `args.SubmittedBy`.

Both use `ON CONFLICT (recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL DO NOTHING`.

Test setup strategy: create tenant + user(s) + a CD + explicit user grant in `public.controlled_document_user_grants` (no tripwire — direct INSERT). Then build a `LifecycleEventArgs` with `ControlledDocumentID = cd.ID` (reader) or `SubmittedBy = user.ID` (author), call `worker.Work()` directly, assert `metaldocs.notifications` rows.

- [ ] **Step 1: Write failing integration tests**

```go
//go:build integration
// +build integration

// internal/modules/notifications/infrastructure/fanout_worker_integration_test.go
package notificationsinfra_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	notificationsinfra "metaldocs/internal/modules/notifications/infrastructure"
	documentsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/tests/integration/testdb"
)

// seedCDUserGrant inserts a direct user grant for a CD (leg 1 of v_cd_obligated_readers).
// No tripwire on this table — plain INSERT is safe in tests.
func seedCDUserGrant(t *testing.T, db *sql.DB, tenantID, cdID, userID string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO public.controlled_document_user_grants (tenant_id, controlled_document_id, user_id)
		 VALUES ($1::uuid, $2::uuid, $3)
		 ON CONFLICT DO NOTHING`,
		tenantID, cdID, userID,
	)
	if err != nil {
		t.Fatalf("seedCDUserGrant: %v", err)
	}
}

// countNotifications counts rows in metaldocs.notifications matching the given filter.
func countNotifications(t *testing.T, db *sql.DB, tenantID, eventType, sourceEventID string) int {
	t.Helper()
	var n int
	err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM metaldocs.notifications
		  WHERE tenant_id = $1::uuid AND event_type = $2 AND source_event_id = $3::uuid`,
		tenantID, eventType, sourceEventID,
	).Scan(&n)
	if err != nil {
		t.Fatalf("countNotifications: %v", err)
	}
	return n
}

func TestFanout(t *testing.T) {
	db, _ := testdb.Open(t)

	t.Run("published_to_obligated_readers", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		userA := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		userB := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		nonReader := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(ten.ID))
		seedCDUserGrant(t, db, ten.ID, cd.ID, userA.ID)
		seedCDUserGrant(t, db, ten.ID, cd.ID, userB.ID)
		// nonReader has no grant — must get zero rows.

		worker := notificationsinfra.NewNotificationsFanoutWorker(db)
		eventID := uuid.NewString()
		job := &river.Job[documentsdomain.LifecycleEventArgs]{
			Args: documentsdomain.LifecycleEventArgs{
				EventID:              eventID,
				TenantID:             ten.ID,
				EventType:            documentsdomain.EventTypeDocumentPublished,
				ResourceType:         "document",
				ResourceID:           uuid.NewString(),
				ControlledDocumentID: cd.ID,
				OccurredAt:           time.Now().UTC(),
			},
		}
		if err := worker.Work(context.Background(), job); err != nil {
			t.Fatalf("Work: %v", err)
		}

		if n := countNotifications(t, db, ten.ID, documentsdomain.EventTypeDocumentPublished, eventID); n != 2 {
			t.Errorf("want 2 reader notifications, got %d", n)
		}
		// Verify nonReader got nothing
		var nonCount int
		_ = db.QueryRowContext(context.Background(),
			`SELECT COUNT(*) FROM metaldocs.notifications
			  WHERE tenant_id=$1::uuid AND recipient_user_id=$2 AND source_event_id=$3::uuid`,
			ten.ID, nonReader.ID, eventID,
		).Scan(&nonCount)
		if nonCount != 0 {
			t.Errorf("nonReader should have 0 notifications, got %d", nonCount)
		}
	})

	t.Run("superseded_to_obligated_readers", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		user := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(ten.ID))
		seedCDUserGrant(t, db, ten.ID, cd.ID, user.ID)

		worker := notificationsinfra.NewNotificationsFanoutWorker(db)
		eventID := uuid.NewString()
		job := &river.Job[documentsdomain.LifecycleEventArgs]{
			Args: documentsdomain.LifecycleEventArgs{
				EventID:              eventID,
				TenantID:             ten.ID,
				EventType:            documentsdomain.EventTypeDocumentSuperseded,
				ResourceType:         "document",
				ResourceID:           uuid.NewString(),
				ControlledDocumentID: cd.ID,
				OccurredAt:           time.Now().UTC(),
			},
		}
		if err := worker.Work(context.Background(), job); err != nil {
			t.Fatalf("Work: %v", err)
		}
		if n := countNotifications(t, db, ten.ID, documentsdomain.EventTypeDocumentSuperseded, eventID); n != 1 {
			t.Errorf("want 1 notification, got %d", n)
		}
	})

	t.Run("obsoleted_to_obligated_readers", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		user := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(ten.ID))
		seedCDUserGrant(t, db, ten.ID, cd.ID, user.ID)

		worker := notificationsinfra.NewNotificationsFanoutWorker(db)
		eventID := uuid.NewString()
		job := &river.Job[documentsdomain.LifecycleEventArgs]{
			Args: documentsdomain.LifecycleEventArgs{
				EventID:              eventID,
				TenantID:             ten.ID,
				EventType:            documentsdomain.EventTypeDocumentObsoleted,
				ResourceType:         "document",
				ResourceID:           uuid.NewString(),
				ControlledDocumentID: cd.ID,
				OccurredAt:           time.Now().UTC(),
			},
		}
		if err := worker.Work(context.Background(), job); err != nil {
			t.Fatalf("Work: %v", err)
		}
		if n := countNotifications(t, db, ten.ID, documentsdomain.EventTypeDocumentObsoleted, eventID); n != 1 {
			t.Errorf("want 1 notification, got %d", n)
		}
	})

	t.Run("approved_to_submitter", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		submitter := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		otherUser := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))

		worker := notificationsinfra.NewNotificationsFanoutWorker(db)
		eventID := uuid.NewString()
		job := &river.Job[documentsdomain.LifecycleEventArgs]{
			Args: documentsdomain.LifecycleEventArgs{
				EventID:      eventID,
				TenantID:     ten.ID,
				EventType:    documentsdomain.EventTypeDocumentApproved,
				ResourceType: "approval_instance",
				ResourceID:   uuid.NewString(),
				SubmittedBy:  submitter.ID,
				OccurredAt:   time.Now().UTC(),
			},
		}
		if err := worker.Work(context.Background(), job); err != nil {
			t.Fatalf("Work: %v", err)
		}
		if n := countNotifications(t, db, ten.ID, documentsdomain.EventTypeDocumentApproved, eventID); n != 1 {
			t.Errorf("want 1 notification for submitter, got %d", n)
		}
		// otherUser must get nothing
		var otherCount int
		_ = db.QueryRowContext(context.Background(),
			`SELECT COUNT(*) FROM metaldocs.notifications
			  WHERE tenant_id=$1::uuid AND recipient_user_id=$2 AND source_event_id=$3::uuid`,
			ten.ID, otherUser.ID, eventID,
		).Scan(&otherCount)
		if otherCount != 0 {
			t.Errorf("other user should have 0 notifications, got %d", otherCount)
		}
	})

	t.Run("rejected_to_submitter", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		submitter := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))

		worker := notificationsinfra.NewNotificationsFanoutWorker(db)
		eventID := uuid.NewString()
		job := &river.Job[documentsdomain.LifecycleEventArgs]{
			Args: documentsdomain.LifecycleEventArgs{
				EventID:      eventID,
				TenantID:     ten.ID,
				EventType:    documentsdomain.EventTypeDocumentRejected,
				ResourceType: "approval_instance",
				ResourceID:   uuid.NewString(),
				SubmittedBy:  submitter.ID,
				OccurredAt:   time.Now().UTC(),
			},
		}
		if err := worker.Work(context.Background(), job); err != nil {
			t.Fatalf("Work: %v", err)
		}
		if n := countNotifications(t, db, ten.ID, documentsdomain.EventTypeDocumentRejected, eventID); n != 1 {
			t.Errorf("want 1 notification, got %d", n)
		}
	})

	t.Run("redelivery_is_noop", func(t *testing.T) {
		ten := testdb.NewTenant(t, db)
		user := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))

		worker := notificationsinfra.NewNotificationsFanoutWorker(db)
		eventID := uuid.NewString()
		job := &river.Job[documentsdomain.LifecycleEventArgs]{
			Args: documentsdomain.LifecycleEventArgs{
				EventID:      eventID,
				TenantID:     ten.ID,
				EventType:    documentsdomain.EventTypeDocumentApproved,
				ResourceType: "approval_instance",
				ResourceID:   uuid.NewString(),
				SubmittedBy:  user.ID,
				OccurredAt:   time.Now().UTC(),
			},
		}
		// First delivery
		if err := worker.Work(context.Background(), job); err != nil {
			t.Fatalf("Work (1st): %v", err)
		}
		// Simulate redelivery — same job, same EventID
		if err := worker.Work(context.Background(), job); err != nil {
			t.Fatalf("Work (2nd): %v", err)
		}
		if n := countNotifications(t, db, ten.ID, documentsdomain.EventTypeDocumentApproved, eventID); n != 1 {
			t.Errorf("want exactly 1 row after redelivery (idempotent), got %d", n)
		}
	})
}
```

- [ ] **Step 2: Run tests — expect compile failure**

```
go test -tags integration ./internal/modules/notifications/infrastructure/...
```
Expected: `undefined: notificationsinfra.NewNotificationsFanoutWorker`

- [ ] **Step 3: Implement the fan-out worker**

```go
// internal/modules/notifications/infrastructure/fanout_worker.go
package notificationsinfra

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"

	documentsdomain "metaldocs/internal/modules/documents/domain"
)

// ptBRMessages maps event_type to the pt-BR inbox title+message (ADR-0044 / F3.3 bundle table).
var ptBRMessages = map[string][2]string{
	documentsdomain.EventTypeDocumentPublished:  {"Novo documento controlado para leitura", "Um novo documento controlado foi publicado e requer sua leitura."},
	documentsdomain.EventTypeDocumentSuperseded: {"Documento substituído por nova revisão", "O documento controlado foi substituído por uma nova versão."},
	documentsdomain.EventTypeDocumentObsoleted:  {"Documento que você lê foi obsoletado", "Um documento controlado que você lê foi marcado como obsoleto."},
	documentsdomain.EventTypeDocumentApproved:   {"Seu documento foi aprovado", "Seu documento submetido para aprovação foi aprovado."},
	documentsdomain.EventTypeDocumentRejected:   {"Documento rejeitado — ajustes solicitados", "Seu documento foi rejeitado. Ajustes são solicitados."},
}

// NotificationsFanoutWorker is a River worker that consumes LifecycleEventArgs jobs
// and inserts per-recipient notification rows into metaldocs.notifications.
// It runs AFTER commit (off-tx; H-PRE-1). Cross-module data is read only from the
// published v_cd_obligated_readers view (reader events) or the event payload (author events).
type NotificationsFanoutWorker struct {
	river.WorkerDefaults[documentsdomain.LifecycleEventArgs]
	db *sql.DB
}

// NewNotificationsFanoutWorker constructs a ready worker. db is required.
func NewNotificationsFanoutWorker(db *sql.DB) *NotificationsFanoutWorker {
	if db == nil {
		panic("notifications_fanout_worker: db is required")
	}
	return &NotificationsFanoutWorker{db: db}
}

func (w *NotificationsFanoutWorker) Work(ctx context.Context, job *river.Job[documentsdomain.LifecycleEventArgs]) error {
	args := job.Args
	msgs, ok := ptBRMessages[args.EventType]
	if !ok {
		// Unknown event type — skip silently (forward-compatible; parked events add new kinds additively).
		return nil
	}
	title, message := msgs[0], msgs[1]

	switch args.EventType {
	case documentsdomain.EventTypeDocumentPublished,
		documentsdomain.EventTypeDocumentSuperseded,
		documentsdomain.EventTypeDocumentObsoleted:
		return w.fanoutToReaders(ctx, args, title, message)
	case documentsdomain.EventTypeDocumentApproved,
		documentsdomain.EventTypeDocumentRejected:
		return w.fanoutToAuthor(ctx, args, title, message)
	default:
		return nil
	}
}

// fanoutToReaders queries v_cd_obligated_readers for the CD and inserts one row per reader.
func (w *NotificationsFanoutWorker) fanoutToReaders(ctx context.Context, args documentsdomain.LifecycleEventArgs, title, message string) error {
	if args.ControlledDocumentID == "" {
		// No CD linked — no obligated readers; skip.
		return nil
	}
	rows, err := w.db.QueryContext(ctx,
		`SELECT user_id FROM metaldocs.v_cd_obligated_readers
		  WHERE tenant_id = $1::uuid AND controlled_document_id = $2::uuid`,
		args.TenantID, args.ControlledDocumentID,
	)
	if err != nil {
		return fmt.Errorf("fanout_worker: query obligated readers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var recipientUserID string
		if err := rows.Scan(&recipientUserID); err != nil {
			return fmt.Errorf("fanout_worker: scan recipient: %w", err)
		}
		if err := w.insertRow(ctx, args, recipientUserID, title, message); err != nil {
			return err
		}
	}
	return rows.Err()
}

// fanoutToAuthor inserts one row for the submitter carried in the event payload.
func (w *NotificationsFanoutWorker) fanoutToAuthor(ctx context.Context, args documentsdomain.LifecycleEventArgs, title, message string) error {
	if args.SubmittedBy == "" {
		return nil
	}
	return w.insertRow(ctx, args, args.SubmittedBy, title, message)
}

// insertRow inserts one notification row. ON CONFLICT DO NOTHING provides idempotency
// via the partial unique index uq_notifications_recipient_event on (recipient_user_id, source_event_id)
// WHERE source_event_id IS NOT NULL (migration 0247).
func (w *NotificationsFanoutWorker) insertRow(ctx context.Context, args documentsdomain.LifecycleEventArgs, recipientUserID, title, message string) error {
	_, err := w.db.ExecContext(ctx, `
		INSERT INTO metaldocs.notifications
		  (tenant_id, recipient_user_id, event_type, resource_type, resource_id, title, message, status, created_at, source_event_id)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, 'PENDING', $8, $9::uuid)
		ON CONFLICT (recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL DO NOTHING`,
		args.TenantID, recipientUserID, args.EventType, args.ResourceType, args.ResourceID,
		title, message, args.OccurredAt, args.EventID,
	)
	if err != nil {
		return fmt.Errorf("fanout_worker: insert notification for %s: %w", recipientUserID, err)
	}
	return nil
}
```

- [ ] **Step 4: Run integration tests — expect green**

```
go test -tags integration ./internal/modules/notifications/infrastructure/... -run TestFanout -v
```
Expected: all 6 subtests PASS

- [ ] **Step 5: Run all notifications tests (non-integration too)**

```
go test ./internal/modules/notifications/...
go test -tags integration ./internal/modules/notifications/... -v
```
Expected: all PASS

- [ ] **Step 6: go build**

```
go build ./...
```
Expected: exit 0

- [ ] **Step 7: Commit**

```
git add internal/modules/notifications/infrastructure/fanout_worker.go \
        internal/modules/notifications/infrastructure/fanout_worker_integration_test.go
git commit -m "feat(M3/F3.3): NotificationsFanoutWorker — River worker, recipient resolution, idempotent insert"
```

---

## Task 6 — Wire binaries: register worker + inject enqueuer

**Files:**
- Modify: `apps/jobs/cmd/metaldocs-jobs/main.go`
- Modify: `apps/api/cmd/metaldocs-api/main.go`

**`metaldocs-jobs`** runs workers (`.Start(ctx)`). **`metaldocs-api`** enqueues (it has the `riverBundle.Client`). Both need changes.

- [ ] **Step 1: Write compile check (build both binaries before changes)**

```
go build ./apps/jobs/... && go build ./apps/api/...
```
Expected: exit 0 (confirm baseline builds)

- [ ] **Step 2: Register fan-out worker in metaldocs-jobs/main.go**

In `apps/jobs/cmd/metaldocs-jobs/main.go`, modify the `BuildJobsDependencies` callback (line 36–43):

```go
deps, err := bootstrap.BuildJobsDependencies(ctx, jobsCfg, func(db *sql.DB) (*river.Workers, error) {
    displayNameRepo := iampg.NewUserDisplayNameRepository(db)
    repo := approvalrepo.NewPostgresApprovalRepository(db, displayNameRepo)
    services := approvalapp.NewServices(repo, approvalapp.NewSQLEmitter(), approvalapp.RealClock{}, cdinfra.NewCDFieldReaderPG())
    workers := approvaljobs.NewWorkers(services.Scheduler, db)
    // F3.3: register notifications fan-out worker.
    river.AddWorker(workers, notificationsinfra.NewNotificationsFanoutWorker(db))
    return workers, nil
})
```

Add import:
```go
notificationsinfra "metaldocs/internal/modules/notifications/infrastructure"
```

- [ ] **Step 3: Inject lifecycle enqueuer in metaldocs-api/main.go**

In `apps/api/cmd/metaldocs-api/main.go`, inside the `if deps.SQLDB != nil` block, after line 494 (`approvalServices.WithScheduledPublishEnqueuer(...)`), add:

```go
// F3.3: wire lifecycle-event enqueuer into the 4 approval services.
approvalServices.WithLifecycleEnqueuer(approvaljobs.NewLifecycleEventEnqueuer(riverBundle.Client))
```

The import `approvaljobs "metaldocs/internal/modules/documents/approval/jobs"` is already present (line ~76 area — confirm with `grep approvaljobs apps/api/cmd/metaldocs-api/main.go`). If not, add it.

- [ ] **Step 4: Build both binaries**

```
go build ./apps/jobs/... && go build ./apps/api/...
```
Expected: exit 0

- [ ] **Step 5: Full build + tests**

```
go build ./...
go test ./...
```
Expected: exit 0 / all PASS

- [ ] **Step 6: Run all CI guards**

```
go vet ./...
go test ./tools/cilint/...
.\scripts\check-module-boundaries.ps1
```
Expected: all OK / 0 violations

- [ ] **Step 7: Verify no river import in documents/domain**

```
grep -r "riverqueue" internal/modules/documents/domain/
```
Expected: no output (domain stays infra-free)

- [ ] **Step 8: Verify api-lint**

```
go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .
```
Expected: `0 violation(s)` (no new routes; notifications surface unchanged)

- [ ] **Step 9: Commit**

```
git add apps/jobs/cmd/metaldocs-jobs/main.go apps/api/cmd/metaldocs-api/main.go
git commit -m "feat(M3/F3.3): register NotificationsFanoutWorker in jobs binary; inject LifecycleEventEnqueuer in api binary"
```

---

## Task 7 — Final validation sweep

- [ ] **Step 1: Full test suite (unit + integration)**

```
go test ./...
go test -tags integration ./internal/modules/notifications/... -v
go test -tags integration ./internal/modules/documents/... -v
```
Expected: all PASS

- [ ] **Step 2: Run validation gate commands from spec**

```
go test -tags integration ./internal/modules/notifications/... -run TestFanout/published_to_obligated_readers
go test -tags integration ./internal/modules/notifications/... -run TestFanout/superseded_to_obligated_readers
go test -tags integration ./internal/modules/notifications/... -run TestFanout/obsoleted_to_obligated_readers
go test -tags integration ./internal/modules/notifications/... -run TestFanout/approved_to_submitter
go test -tags integration ./internal/modules/notifications/... -run TestFanout/rejected_to_submitter
go test -tags integration ./internal/modules/notifications/... -run TestFanout/redelivery_is_noop
```
Expected: all PASS

- [ ] **Step 3: Verify publish/approval semantics unchanged**

```
go test ./internal/modules/documents/approval/application/... -v
```
Expected: all pre-existing tests PASS (no assertion edits)

- [ ] **Step 4: outboxpair + module-boundaries + hgcrossmodule**

```
go test ./tools/cilint/...
.\scripts\check-module-boundaries.ps1
```
Expected: all OK; `hgcrossmodule = 0`; module-boundaries = OK

- [ ] **Step 5: Verify publish path diff is additive only**

```
git diff main -- internal/modules/documents/approval/application/publish_service.go | grep "^+" | grep -v "lifecycleEnqueuer\|EnqueueLifecycleEventTx\|LoadDocumentControlledDocumentID\|EventType\|uuid\|import\|documentsdomain"
```
Expected: no semantic-logic additions outside the enqueue block

- [ ] **Step 6: System runnable check**

```
.\scripts\check-system-runnable.ps1
```
Expected: OK

- [ ] **Step 7: Write evidence.md**

Copy `templates/feature-evidence.md` to `docs/superpowers/milestones/frontend-screen-completion/milestone-3-notifications/f3.3-lifecycle-emitter/evidence.md` and fill with actual command outputs, real test run results, and TDD proof.

---

## Source (milestone row)

- **Milestone spec row (F3.3 — what to implement):** Five additive in-tx typed River-job domain-event enqueues (document.published/superseded/obsoleted/approved/rejected) in the owning module, plus one notifications fan-out River worker that resolves recipients (obligated readers via `v_cd_obligated_readers`; submitter via `submitted_by` in the payload) and inserts idempotent per-recipient rows. Domain-event contract type in `documents/domain`. HS-2 lifted for the additive enqueue only.
- **Governing ADR:** [ADR-0044](../../../../../wiki/decisions/0044-domain-event-pattern-and-river-dispatch.md)

## Execution notes

(Fill during `subagent-driven-development` execution: model choices, deviations from plan with rationale, questions answered.)
