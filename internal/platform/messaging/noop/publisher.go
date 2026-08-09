package noop

import (
	"context"

	"metaldocs/internal/platform/messaging"
)

// Publisher is a no-op messaging.Publisher that discards every event.
type Publisher struct{}

// NewPublisher constructs a no-op Publisher.
func NewPublisher() *Publisher {
	return &Publisher{}
}

// Publish discards event and always returns nil.
func (p *Publisher) Publish(_ context.Context, _ messaging.Event) error {
	return nil
}
