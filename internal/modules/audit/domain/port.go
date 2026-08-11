// Package domain declares the audit module's core contracts: the append-only
// Event record, the Writer/Reader/Counter ports that infrastructure adapters
// implement, and the export-job types used by the async-export pipeline.
package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/platform/db"
)

// Event is a single append-only audit record. Rows are persisted to
// metaldocs.audit_events with a prev_hash/row_hash chain for tamper-evidence;
// Event itself carries no hash fields — those are infrastructure-layer concerns.
type Event struct {
	ID           string
	OccurredAt   time.Time
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	PayloadJSON  string
	TraceID      string
	TenantID     string
}

// ErrInvalidEvent is returned by NewEvent when a required field (tenant,
// actor, action, resource type, or resource id) is blank after trimming.
var ErrInvalidEvent = errors.New("invalid event")

// NewEvent builds an Event with OccurredAt set to now (UTC), marshaling payload
// to JSON (defaulting to "{}" when nil). Returns ErrInvalidEvent if any of
// tenantID, actorID, action, resourceType, or resourceID is empty after
// trimming — callers must supply a fully-scoped event, never a partial one.
func NewEvent(tenantID, resourceType, resourceID, actorID, action string, payload any) (Event, error) {
	payloadJSON := "{}"
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return Event{}, fmt.Errorf("audit: marshal payload: %w", err)
		}
		payloadJSON = string(raw)
	}
	event := Event{
		TenantID:     strings.TrimSpace(tenantID),
		ActorID:      strings.TrimSpace(actorID),
		Action:       strings.TrimSpace(action),
		ResourceType: strings.TrimSpace(resourceType),
		ResourceID:   strings.TrimSpace(resourceID),
		PayloadJSON:  payloadJSON,
		OccurredAt:   time.Now().UTC(),
	}
	if event.TenantID == "" || event.ActorID == "" || event.Action == "" || event.ResourceType == "" || event.ResourceID == "" {
		return Event{}, fmt.Errorf("audit: %w", ErrInvalidEvent)
	}
	return event, nil
}

// IntegrityIssueKind classifies a single row-hash-chain discrepancy found by
// IntegrityValidator.ValidateIntegrity.
type IntegrityIssueKind string

const (
	// IntegrityIssuePrevHashMismatch marks a row whose stored prev_hash does not
	// equal the row_hash of its chain predecessor — the chain link is broken.
	IntegrityIssuePrevHashMismatch IntegrityIssueKind = "prev_hash_mismatch"
	// IntegrityIssueRowHashMismatch marks a row whose stored row_hash does not
	// match the hash recomputed from its own column values — the row was altered
	// after being written.
	IntegrityIssueRowHashMismatch IntegrityIssueKind = "row_hash_mismatch"
)

// IntegrityIssue reports one detected break in the audit_events hash chain,
// identifying the offending row (Sequence/EventID), the Kind of mismatch, and
// both the expected and actual hash values for diagnosis.
type IntegrityIssue struct {
	Sequence         int64
	EventID          string
	Kind             IntegrityIssueKind
	ExpectedHash     string
	ActualHash       string
	ExpectedPrevHash string
	ActualPrevHash   string
}

// ListEventsQuery captures every filter axis the Admin Center Audit Trail
// supports. TenantID is mandatory at the application layer; the repository
// trusts that the application has already enforced tenant scoping.
type ListEventsQuery struct {
	TenantID       string
	ResourceType   string
	ResourceID     string
	ActorID        string
	Action         string
	OccurredAfter  time.Time
	OccurredBefore time.Time
	Query          string
	Cursor         Cursor
	Limit          int
}

// Cursor encodes the (occurred_at, id) anchor used to keep pagination stable
// across pages. Zero value means "no cursor".
type Cursor struct {
	OccurredAt time.Time
	ID         string
}

// IsZero reports whether c carries no anchor (the "first page" cursor).
func (c Cursor) IsZero() bool { return c.ID == "" && c.OccurredAt.IsZero() }

// Writer stays in the domain package because audit append semantics are part of
// the cross-module contract today. Move it behind a narrower application port
// once write flows stop passing raw audit events across module boundaries.
//
// Record and RecordTx are fire-and-forget from the caller's perspective: the
// regulated mutation's own transaction must already have committed (or use
// RecordTx to bundle the audit insert into that same transaction) — an audit
// write failure must never roll back or block the regulated action it
// describes.
type Writer interface {
	// Record appends event in its own new transaction. Use for the standard
	// post-commit case: the caller's regulated mutation has already committed,
	// and this call is a separate, independent write.
	Record(ctx context.Context, event Event) error
	// RecordTx appends event inside the caller's own transaction tx, so the
	// audit insert and the regulated mutation commit or roll back atomically.
	// Used by callers (e.g. bypass/documents adapters) that need the audit
	// record to share fate with the mutation itself.
	RecordTx(ctx context.Context, tx db.Tx, event Event) error
}

// Reader is the audit query port. ListEvents intentionally lives on Reader
// rather than Writer so write-only implementations are allowed even when they
// do not support query workloads. It returns up to query.Limit events plus
// hasMore: implementations fetch a limit+1 probe row and report hasMore=true
// only when that extra row exists, so an exact-multiple last page no longer
// falsely advertises a next page (AIP-158; matches the documents /
// controlled-documents keyset shape).
type Reader interface {
	ListEvents(ctx context.Context, query ListEventsQuery) (items []Event, hasMore bool, err error)
}

// Counter estimates how many rows a filter yields. Used to decide whether an
// export runs synchronously inline.
type Counter interface {
	CountEvents(ctx context.Context, query ListEventsQuery) (int64, error)
}

// IntegrityValidator verifies the audit_events row-hash chain is unbroken —
// used by the audit-integrity-validator janitor to detect tampering.
type IntegrityValidator interface {
	ValidateIntegrity(ctx context.Context) ([]IntegrityIssue, error)
}

// ErasureChecker reports whether a tenant has been erased (ADR 0070
// crypto-shred). application.Service.ExportEvents uses it to re-check status
// immediately before persisting an export job, closing the window where a
// tenant's erasure commits after the export's rows were fetched (an earlier,
// separate read) but before the export job is written to durable storage
// (PR #121 review round 1, P1 — export-persistence half). Implementations
// must fail closed: a lookup error must not be treated as "not erased".
type ErasureChecker interface {
	IsErased(ctx context.Context, tenantID string) (bool, error)
}

// ---- Export job types -----------------------------------------------------

// ExportFormat is the requested serialization for an audit export (csv or
// jsonl). Use Valid to check a value parsed from an external request.
type ExportFormat string

const (
	// ExportFormatCSV requests a comma-separated-values export.
	ExportFormatCSV ExportFormat = "csv"
	// ExportFormatJSONL requests a newline-delimited-JSON export.
	ExportFormatJSONL ExportFormat = "jsonl"
)

// Valid reports whether f is a recognized export format (csv or jsonl).
func (f ExportFormat) Valid() bool {
	return f == ExportFormatCSV || f == ExportFormatJSONL
}

// ExportStatus is the lifecycle state of an ExportJob.
type ExportStatus string

const (
	// ExportStatusPending is the initial state of an export job before it
	// starts running.
	ExportStatusPending ExportStatus = "pending"
	// ExportStatusRunning means the export job is actively producing its
	// payload.
	ExportStatusRunning ExportStatus = "running"
	// ExportStatusReady means the export payload is persisted and downloadable.
	// Synchronous exports move straight to Ready in the same call that creates
	// the job.
	ExportStatusReady ExportStatus = "ready"
	// ExportStatusFailed means the export attempt errored and produced no
	// downloadable payload.
	ExportStatusFailed ExportStatus = "failed"
)

// ExportJob is the persisted row for an audit export request. Inline (sync)
// exports are written with status=ready in the same transaction.
type ExportJob struct {
	ID            string
	TenantID      uuid.UUID
	ActorID       string
	Format        ExportFormat
	FilterJSON    string
	Status        ExportStatus
	ObjectKey     string
	DownloadToken string
	ExpiresAt     time.Time
	ErrorMessage  string
	EstimatedRows int64
	ActualRows    int64
	Payload       []byte
	CreatedAt     time.Time
	CompletedAt   time.Time
}

// ExportJobRepository persists export jobs and the rendered payload.
type ExportJobRepository interface {
	Save(ctx context.Context, job ExportJob) error
	Get(ctx context.Context, tenantID, exportID string) (ExportJob, error)
	GetByDownloadToken(ctx context.Context, exportID, token string) (ExportJob, error)
}

// ErrExportJobNotFound is returned when an export job lookup finds no row for
// the given (tenant, id) or (id, token) pair — including when the caller is
// not the job's owning actor, so ownership checks fail closed with the same
// not-found error rather than leaking existence.
var ErrExportJobNotFound = errors.New("audit: export job not found")
