package tenant

import (
	"context"
	"errors"
	"strings"
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

func TestWithTenantID_EmptyPanics(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic for empty tenant id")
		}
		msg, ok := recovered.(string)
		if !ok {
			t.Fatalf("panic type = %T, want string", recovered)
		}
		if !strings.Contains(msg, "empty tenantID") {
			t.Fatalf("panic = %q, want empty tenantID message", msg)
		}
	}()

	_ = WithTenantID(context.Background(), "")
}
