package messaging

import (
	"context"
	"time"
)

type FailedEvent struct {
	EventID        EventID
	LastError      string
	NextAttemptAt  *time.Time
	DeadLetteredAt *time.Time
}

type Consumer interface {
	ClaimUnpublished(ctx context.Context, limit int) ([]Event, error)
	MarkPublished(ctx context.Context, eventIDs []EventID) error
	MarkFailed(ctx context.Context, failure FailedEvent) error
}
