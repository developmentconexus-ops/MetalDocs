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

func NewEvent(tenantID, resourceType, resourceID, actorID string, payload any) (Event, error) {
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
		ResourceType: strings.TrimSpace(resourceType),
		ResourceID:   strings.TrimSpace(resourceID),
		PayloadJSON:  payloadJSON,
		OccurredAt:   time.Now().UTC(),
	}
	if event.TenantID == "" || event.ActorID == "" || event.ResourceType == "" || event.ResourceID == "" {
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

type ListEventsQuery struct {
	ResourceType string
	ResourceID   string
	TenantID     string
	Limit        int
}

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

type IntegrityValidator interface {
	ValidateIntegrity(ctx context.Context) ([]IntegrityIssue, error)
}
