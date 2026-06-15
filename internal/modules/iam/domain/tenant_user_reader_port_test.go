package domain

import (
	"context"
	"testing"
)

func TestNoopTenantUserReader(t *testing.T) {
	ids, err := NoopTenantUserReader{}.TenantUserIDs(context.Background(), "any-tenant")
	if err != nil {
		t.Fatalf("Noop TenantUserIDs error = %v, want nil", err)
	}
	if ids == nil {
		t.Fatalf("Noop TenantUserIDs returned nil slice, want empty non-nil")
	}
	if len(ids) != 0 {
		t.Fatalf("Noop TenantUserIDs len = %d, want 0", len(ids))
	}
}
