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
