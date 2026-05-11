package tenant

import (
	"context"
	"errors"
	"testing"
)

func TestFromContext_RoundTrip(t *testing.T) {
	ctx := WithTenantID(context.Background(), "tenant-a")
	got, err := FromContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "tenant-a" {
		t.Fatalf("got %q, want tenant-a", got)
	}
}

func TestFromContext_MissingReturnsSentinel(t *testing.T) {
	_, err := FromContext(context.Background())
	if !errors.Is(err, ErrTenantMissing) {
		t.Fatalf("got %v, want ErrTenantMissing", err)
	}
}

func TestFromContext_EmptyTreatedAsMissing(t *testing.T) {
	ctx := WithTenantID(context.Background(), "")
	_, err := FromContext(ctx)
	if !errors.Is(err, ErrTenantMissing) {
		t.Fatalf("got %v, want ErrTenantMissing", err)
	}
}
