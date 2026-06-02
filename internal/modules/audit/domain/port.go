package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

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

var ErrInvalidEvent = errors.New("invalid event")

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

type IntegrityIssueKind string

const (
	IntegrityIssuePrevHashMismatch IntegrityIssueKind = "prev_hash_mismatch"
	IntegrityIssueRowHashMismatch  IntegrityIssueKind = "row_hash_mismatch"
)

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

func (c Cursor) IsZero() bool { return c.ID == "" && c.OccurredAt.IsZero() }

// Writer stays in the domain package because audit append semantics are part of
// the cross-module contract today. Move it behind a narrower application port
// once write flows stop passing raw audit events across module boundaries.
type Writer interface {
	Record(ctx context.Context, event Event) error
	RecordTx(ctx context.Context, tx *sql.Tx, event Event) error
}

// ListEvents intentionally lives on Reader rather than Writer so write-only
// implementations are allowed even when they do not support query workloads.
type Reader interface {
	ListEvents(ctx context.Context, query ListEventsQuery) ([]Event, error)
}

// Counter estimates how many rows a filter yields. Used to decide whether an
// export runs synchronously inline.
type Counter interface {
	CountEvents(ctx context.Context, query ListEventsQuery) (int64, error)
}

type IntegrityValidator interface {
	ValidateIntegrity(ctx context.Context) ([]IntegrityIssue, error)
}

// ---- Export job types -----------------------------------------------------

type ExportFormat string

const (
	ExportFormatCSV   ExportFormat = "csv"
	ExportFormatJSONL ExportFormat = "jsonl"
)

func (f ExportFormat) Valid() bool {
	return f == ExportFormatCSV || f == ExportFormatJSONL
}

type ExportStatus string

const (
	ExportStatusPending ExportStatus = "pending"
	ExportStatusRunning ExportStatus = "running"
	ExportStatusReady   ExportStatus = "ready"
	ExportStatusFailed  ExportStatus = "failed"
)

// ExportJob is the persisted row for an audit export request. Inline (sync)
// exports are written with status=ready in the same transaction.
type ExportJob struct {
	ID            string
	TenantID      string
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

var ErrExportJobNotFound = errors.New("audit: export job not found")
