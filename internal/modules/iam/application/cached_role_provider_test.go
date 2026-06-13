package application

import (
	"context"
	"testing"
	"time"

	"metaldocs/internal/modules/iam/domain"
)

// stubRoleProvider is a minimal RoleProvider for cache unit tests. It counts
// RolesByUserID calls so tests can verify cache hits vs source reads.
type stubRoleProvider struct {
	calls  int
	roles  []domain.Role
	active bool
}

func (s *stubRoleProvider) RolesByUserID(_ context.Context, _, _ string) ([]domain.Role, error) {
	s.calls++
	return cloneRoles(s.roles), nil
}

func (s *stubRoleProvider) RolesByUserIDs(_ context.Context, _ string, userIDs []string) (map[string][]domain.Role, error) {
	s.calls++
	out := map[string][]domain.Role{}
	for _, uid := range userIDs {
		out[uid] = cloneRoles(s.roles)
	}
	return out, nil
}

func (s *stubRoleProvider) UserActiveInTenant(_ context.Context, _, _ string) (bool, error) {
	return s.active, nil
}

// TestCachedRoleProvider_CacheHitAndInvalidate is the focused RF-3 contract test:
//
//  1. First call on a cold key hits the base provider.
//  2. Second call within TTL is served from cache (base call count stays at 1).
//  3. InvalidateUserTenant evicts the entry.
//  4. Call after eviction hits the base provider again (count becomes 2).
func TestCachedRoleProvider_CacheHitAndInvalidate(t *testing.T) {
	base := &stubRoleProvider{roles: []domain.Role{domain.RoleEditor}, active: true}
	// Use a very long TTL so the entry never expires during the test.
	provider := NewCachedRoleProvider(context.Background(), base, 10*time.Minute)

	// Cold read — must hit source.
	roles, err := provider.RolesByUserID(context.Background(), "u1", "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 1 || roles[0] != domain.RoleEditor {
		t.Fatalf("unexpected roles: %v", roles)
	}
	if base.calls != 1 {
		t.Fatalf("want 1 base call after cold read, got %d", base.calls)
	}

	// Warm read — must be served from cache (no additional base call).
	_, err = provider.RolesByUserID(context.Background(), "u1", "t1")
	if err != nil {
		t.Fatalf("unexpected error on warm read: %v", err)
	}
	if base.calls != 1 {
		t.Fatalf("want still 1 base call after cache hit, got %d", base.calls)
	}

	// Invalidate the entry.
	provider.InvalidateUserTenant("u1", "t1")

	// Post-invalidation read — must hit source again.
	_, err = provider.RolesByUserID(context.Background(), "u1", "t1")
	if err != nil {
		t.Fatalf("unexpected error after invalidation: %v", err)
	}
	if base.calls != 2 {
		t.Fatalf("want 2 base calls after invalidation+read, got %d", base.calls)
	}
}

// TestCachedRoleProvider_TenantIsolation verifies that a cache entry for
// (u1, t1) does not satisfy a read for (u1, t2) — keys are scoped to tenant.
func TestCachedRoleProvider_TenantIsolation(t *testing.T) {
	base := &stubRoleProvider{roles: []domain.Role{domain.RoleEditor}, active: true}
	provider := NewCachedRoleProvider(context.Background(), base, 10*time.Minute)

	// Warm the (u1, t1) entry.
	if _, err := provider.RolesByUserID(context.Background(), "u1", "t1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A read for (u1, t2) must miss the cache and hit the base provider.
	if _, err := provider.RolesByUserID(context.Background(), "u1", "t2"); err != nil {
		t.Fatalf("unexpected error for t2: %v", err)
	}
	// Two distinct keys: two base calls.
	if base.calls != 2 {
		t.Fatalf("want 2 base calls for distinct tenants, got %d", base.calls)
	}
}

// TestCachedRoleProvider_TTLExpiry verifies that a cached entry is not served
// after its TTL has elapsed.
func TestCachedRoleProvider_TTLExpiry(t *testing.T) {
	base := &stubRoleProvider{roles: []domain.Role{domain.RoleViewer}, active: true}
	// Very short TTL so the entry expires almost immediately.
	ttl := 10 * time.Millisecond
	provider := NewCachedRoleProvider(context.Background(), base, ttl)

	// Cold read to populate cache.
	if _, err := provider.RolesByUserID(context.Background(), "u1", "t1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base.calls != 1 {
		t.Fatalf("want 1 base call, got %d", base.calls)
	}

	// Wait for TTL to expire.
	time.Sleep(ttl + 5*time.Millisecond)

	// Post-TTL read must hit source.
	if _, err := provider.RolesByUserID(context.Background(), "u1", "t1"); err != nil {
		t.Fatalf("unexpected error after TTL: %v", err)
	}
	if base.calls != 2 {
		t.Fatalf("want 2 base calls after TTL expiry, got %d", base.calls)
	}
}

// TestCachedRoleProvider_InvalidateDoesNotAffectOtherKeys checks that evicting
// (u1, t1) leaves an unrelated (u2, t1) entry intact.
func TestCachedRoleProvider_InvalidateDoesNotAffectOtherKeys(t *testing.T) {
	base := &stubRoleProvider{roles: []domain.Role{domain.RoleViewer}, active: true}
	provider := NewCachedRoleProvider(context.Background(), base, 10*time.Minute)

	// Warm two distinct entries.
	if _, err := provider.RolesByUserID(context.Background(), "u1", "t1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := provider.RolesByUserID(context.Background(), "u2", "t1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base.calls != 2 {
		t.Fatalf("want 2 base calls, got %d", base.calls)
	}

	// Evict only u1/t1.
	provider.InvalidateUserTenant("u1", "t1")

	// u2/t1 should still be cached — no additional base call.
	if _, err := provider.RolesByUserID(context.Background(), "u2", "t1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base.calls != 2 {
		t.Fatalf("u2/t1 should remain cached; base calls = %d, want 2", base.calls)
	}

	// u1/t1 should be a cache miss and require a base call.
	if _, err := provider.RolesByUserID(context.Background(), "u1", "t1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base.calls != 3 {
		t.Fatalf("u1/t1 should miss after invalidation; base calls = %d, want 3", base.calls)
	}
}
