package messaging

import (
	"context"
	"time"
)

// FailedEvent describes an outbox event that failed publication, for
// recording the failure and scheduling (or exhausting) a retry.
type FailedEvent struct {
	EventID        EventID
	LastError      string
	NextAttemptAt  *time.Time
	DeadLetteredAt *time.Time
}

// Consumer abstracts the transactional outbox consumer: claiming unpublished
// events, and marking them published or failed.
type Consumer interface {
	ClaimUnpublished(ctx context.Context, limit int) ([]Event, error)
	MarkPublished(ctx context.Context, eventIDs []EventID) error
	MarkFailed(ctx context.Context, failure FailedEvent) error
}
