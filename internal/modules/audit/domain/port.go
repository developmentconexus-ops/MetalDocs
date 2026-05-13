package domain

import (
	"context"
	"database/sql"
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

type ListEventsQuery struct {
	ResourceType string
	ResourceID   string
	TenantID     string
	Limit        int
}

type Writer interface {
	Record(ctx context.Context, event Event) error
	RecordTx(ctx context.Context, tx *sql.Tx, event Event) error
}

type Reader interface {
	ListEvents(ctx context.Context, query ListEventsQuery) ([]Event, error)
}

type IntegrityValidator interface {
	ValidateIntegrity(ctx context.Context) ([]IntegrityIssue, error)
}
