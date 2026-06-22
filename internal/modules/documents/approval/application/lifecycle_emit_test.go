package application

import (
	"context"
	"testing"

	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/platform/db"
)

// spyLifecycleEnqueuer captures enqueued lifecycle events for assertion.
type spyLifecycleEnqueuer struct {
	calls []docsdomain.LifecycleEventArgs
}

func (s *spyLifecycleEnqueuer) EnqueueLifecycleEventTx(_ context.Context, _ db.Tx, args docsdomain.LifecycleEventArgs) error {
	s.calls = append(s.calls, args)
	return nil
}

// TestWithLifecycleEnqueuer_Services verifies Services.WithLifecycleEnqueuer
// wires the enqueuer to all 4 emit services and returns the same *Services.
func TestWithLifecycleEnqueuer_Services(t *testing.T) {
	spy := &spyLifecycleEnqueuer{}
	svc := NewServices(nil, &noopSQLEmitter{}, RealClock{}, nil)
	got := svc.WithLifecycleEnqueuer(spy)
	if got != svc {
		t.Error("WithLifecycleEnqueuer must return the same *Services pointer")
	}
	// Each field must now have the enqueuer set.
	if svc.Publish.lifecycleEnqueuer != spy {
		t.Error("Publish.lifecycleEnqueuer not wired")
	}
	if svc.Supersede.lifecycleEnqueuer != spy {
		t.Error("Supersede.lifecycleEnqueuer not wired")
	}
	if svc.Obsolete.lifecycleEnqueuer != spy {
		t.Error("Obsolete.lifecycleEnqueuer not wired")
	}
	if svc.Decision.lifecycleEnqueuer != spy {
		t.Error("Decision.lifecycleEnqueuer not wired")
	}
}

// TestWithLifecycleEnqueuer_NilServices verifies nil *Services does not panic.
func TestWithLifecycleEnqueuer_NilServices(t *testing.T) {
	var svc *Services
	got := svc.WithLifecycleEnqueuer(&spyLifecycleEnqueuer{})
	if got != nil {
		t.Error("nil receiver must return nil")
	}
}

// noopSQLEmitter satisfies EventEmitter with a no-op.
type noopSQLEmitter struct{}

func (noopSQLEmitter) Emit(_ context.Context, _ db.Tx, _ GovernanceEvent) error { return nil }
