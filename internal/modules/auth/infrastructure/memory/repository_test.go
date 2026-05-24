package memory

import (
	"context"
	"errors"
	"sort"
	"sync"
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
		SessionID:  "test-session-id",
		UserID:     "test-user-id",
		TenantID:   "t1",
		CreatedAt:  mustParseTime(t, "2025-05-11T08:00:00Z"),
		ExpiresAt:  mustParseTime(t, "2025-05-12T08:00:00Z"),
		IPAddress:  "192.0.2.1",
		UserAgent:  "Test Agent",
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

func TestSessionMutations_UnknownSessionReturnsNotFound(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()
	now := time.Now().UTC()

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "touch",
			run: func() error {
				return repo.TouchSession(ctx, "missing-session", now)
			},
		},
		{
			name: "revoke",
			run: func() error {
				return repo.RevokeSession(ctx, "missing-session", now)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); !errors.Is(err, authdomain.ErrSessionNotFound) {
				t.Fatalf("expected ErrSessionNotFound, got %v", err)
			}
		})
	}
}

func TestRecordFailedLogin_ConcurrentIncrementsAndLockout(t *testing.T) {
	cases := []struct {
		name            string
		concurrentFails int
		maxAttempts     int
		wantLocked      bool
	}{
		{
			name:            "below-threshold",
			concurrentFails: 2,
			maxAttempts:     3,
			wantLocked:      false,
		},
		{
			name:            "at-threshold",
			concurrentFails: 4,
			maxAttempts:     3,
			wantLocked:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewRepository()
			ctx := context.Background()
			userID := "concurrent-fail-user"

			if err := repo.CreateUser(ctx, authdomain.CreateUserParams{
				UserID:       userID,
				Username:     userID,
				Email:        "concurrent@example.com",
				DisplayName:  "Concurrent Fail User",
				PasswordHash: "hash",
				PasswordAlgo: "bcrypt",
				IsActive:     true,
			}); err != nil {
				t.Fatalf("CreateUser: %v", err)
			}

			start := make(chan struct{})
			results := make(chan int, tc.concurrentFails)
			errs := make(chan error, tc.concurrentFails)
			var wg sync.WaitGroup
			for i := 0; i < tc.concurrentFails; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					attempts, _, err := repo.RecordFailedLogin(ctx, userID, tc.maxAttempts, 600)
					if err != nil {
						errs <- err
						return
					}
					results <- attempts
				}()
			}
			close(start)
			wg.Wait()
			close(results)
			close(errs)

			for err := range errs {
				if err != nil {
					t.Fatalf("RecordFailedLogin: %v", err)
				}
			}

			attempts := make([]int, 0, tc.concurrentFails)
			for attempt := range results {
				attempts = append(attempts, attempt)
			}
			if len(attempts) != tc.concurrentFails {
				t.Fatalf("expected %d successful updates, got %d", tc.concurrentFails, len(attempts))
			}
			sort.Ints(attempts)
			for i, attempt := range attempts {
				if want := i + 1; attempt != want {
					t.Fatalf("attempt sequence mismatch: got %v, want 1..%d", attempts, tc.concurrentFails)
				}
			}

			identity, err := repo.FindIdentityByUserID(ctx, userID)
			if err != nil {
				t.Fatalf("FindIdentityByUserID: %v", err)
			}
			if identity.FailedLoginAttempts != tc.concurrentFails {
				t.Fatalf("failed login attempts mismatch: got %d, want %d", identity.FailedLoginAttempts, tc.concurrentFails)
			}
			if tc.wantLocked && identity.LockedUntil == nil {
				t.Fatal("expected lockout to be set")
			}
			if !tc.wantLocked && identity.LockedUntil != nil {
				t.Fatalf("expected no lockout, got %v", identity.LockedUntil)
			}
		})
	}
}
