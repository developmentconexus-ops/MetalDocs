package requesttrace

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestNormalize_RejectsUnsafeTraceIDs(t *testing.T) {
	if _, ok := Normalize("trace\r\npoison"); ok {
		t.Fatal("expected poisoned trace id to be rejected")
	}
}

func TestResolve_ReusesContextTraceID(t *testing.T) {
	ctx := WithTraceID(context.Background(), "trace-123")

	if got := Resolve(ctx); got != "trace-123" {
		t.Fatalf("Resolve() = %q, want %q", got, "trace-123")
	}
}

func TestResolve_GeneratesTraceIDWhenMissing(t *testing.T) {
	got := Resolve(context.Background())
	if _, err := uuid.Parse(got); err != nil {
		t.Fatalf("Resolve() = %q, want generated uuid: %v", got, err)
	}
}
