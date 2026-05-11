package memory

import (
	"context"
	"testing"
	"time"

	authdomain "metaldocs/internal/modules/auth/domain"
)

func mustParseTime(t *testing.T, s string) time.Time {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse time %q: %v", s, err)
	}
	return tm
}

// TestGetUserTenants_EmptyForNewUser verifies that GetUserTenants returns
// nil/empty for a fresh user with no roles in the memory repository.
func TestGetUserTenants_EmptyForNewUser(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()

	tenants, err := repo.GetUserTenants(ctx, "new-user-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenants != nil && len(tenants) != 0 {
		t.Errorf("expected empty/nil slice, got %v", tenants)
	}
}

// TestGetUserTenants_ConsistentWithEmptyUsers verifies that GetUserTenants
// handles multiple different users consistently.
func TestGetUserTenants_ConsistentWithEmptyUsers(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()

	// Call GetUserTenants for multiple different users that don't exist.
	for _, userID := range []string{"user1", "user2", "user3"} {
		tenants, err := repo.GetUserTenants(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", userID, err)
		}
		if tenants != nil && len(tenants) != 0 {
			t.Errorf("expected empty/nil for %s, got %v", userID, tenants)
		}
	}
}

// TestCreateAndFindSession_PreservesRoundTrip verifies that a session
// with TenantID is correctly persisted and retrieved.
func TestCreateAndFindSession_PreservesRoundTrip(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()

	session := authdomain.Session{
		SessionID: "test-session-id",
		UserID:    "test-user-id",
		TenantID:  "t1",
		CreatedAt: mustParseTime(t, "2025-05-11T08:00:00Z"),
		ExpiresAt: mustParseTime(t, "2025-05-12T08:00:00Z"),
		IPAddress: "192.0.2.1",
		UserAgent: "Test Agent",
		LastSeenAt: mustParseTime(t, "2025-05-11T08:00:00Z"),
	}

	// Create the session
	if err := repo.CreateSession(ctx, session); err != nil {
		t.Fatalf("CreateSession error: %v", err)
	}

	// Find it back
	found, err := repo.FindSession(ctx, session.SessionID)
	if err != nil {
		t.Fatalf("FindSession error: %v", err)
	}

	// Verify TenantID is preserved
	if found.TenantID != "t1" {
		t.Errorf("TenantID mismatch: got %q, want %q", found.TenantID, "t1")
	}
	if found.UserID != "test-user-id" {
		t.Errorf("UserID mismatch: got %q, want %q", found.UserID, "test-user-id")
	}
}
