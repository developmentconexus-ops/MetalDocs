package wiring

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	authdomain "metaldocs/internal/modules/auth/domain"
	docapp "metaldocs/internal/modules/documents/application"
)

// authListUsersFunc is the narrow function-typed seam the adapter consumes.
// In production it is satisfied by *auth.Service.ListUsers. The test stands a
// trivial fake against it; no sloppy strings, UUIDs for UserID per CLAUDE.md §4.
type authListUsersFunc func(ctx context.Context, tenantID string) ([]authdomain.ManagedUser, error)

func (f authListUsersFunc) ListUsers(ctx context.Context, tenantID string) ([]authdomain.ManagedUser, error) {
	return f(ctx, tenantID)
}

func newManagedUser(displayName string, isActive bool) authdomain.ManagedUser {
	return authdomain.ManagedUser{
		UserID:      uuid.New().String(),
		DisplayName: displayName,
		IsActive:    isActive,
	}
}

func TestDocumentsIAMUserOptions(t *testing.T) {
	t.Parallel()

	tenantA := uuid.New().String()

	zoeActive := newManagedUser("Zoe", true)
	aliceActive := newManagedUser("alice", true) // lower-case to exercise case-insensitive sort
	bobInactive := newManagedUser("Bob", false)
	mikeActive := newManagedUser("Mike", true)
	mikeTwinActive := authdomain.ManagedUser{
		// Same DisplayName as mikeActive to exercise UserID tie-break.
		// Force UserID lexicographically smaller than mikeActive.UserID.
		UserID:      "00000000-0000-0000-0000-000000000001",
		DisplayName: "Mike",
		IsActive:    true,
	}

	sentinelErr := errors.New("auth boom")

	cases := []struct {
		name         string
		tenantID     string
		fake         authListUsersFunc
		wantUserIDs  []string // order-significant; UserID is unique
		wantErr      error
		assertNotNil bool // empty result must still be a non-nil slice
	}{
		{
			name:     "filters inactive — Bob dropped",
			tenantID: tenantA,
			fake: func(_ context.Context, _ string) ([]authdomain.ManagedUser, error) {
				return []authdomain.ManagedUser{zoeActive, bobInactive, aliceActive}, nil
			},
			wantUserIDs: []string{aliceActive.UserID, zoeActive.UserID}, // case-insensitive: alice < Zoe
		},
		{
			name:     "sorts case-insensitive ASC, tie-break by UserID ASC",
			tenantID: tenantA,
			fake: func(_ context.Context, _ string) ([]authdomain.ManagedUser, error) {
				return []authdomain.ManagedUser{mikeActive, aliceActive, mikeTwinActive}, nil
			},
			// alice < Mike; among the two "Mike"s, the lex-smaller UUID wins.
			wantUserIDs: []string{aliceActive.UserID, mikeTwinActive.UserID, mikeActive.UserID},
		},
		{
			name:     "empty result returns non-nil empty slice",
			tenantID: tenantA,
			fake: func(_ context.Context, _ string) ([]authdomain.ManagedUser, error) {
				return nil, nil
			},
			wantUserIDs:  []string{},
			assertNotNil: true,
		},
		{
			name:     "propagates underlying error and returns nil slice",
			tenantID: tenantA,
			fake: func(_ context.Context, _ string) ([]authdomain.ManagedUser, error) {
				return nil, sentinelErr
			},
			wantErr: sentinelErr,
		},
		{
			name:     "forwards tenantID verbatim (tenant isolation)",
			tenantID: tenantA,
			fake: func(_ context.Context, gotTenant string) ([]authdomain.ManagedUser, error) {
				if gotTenant != tenantA {
					t.Fatalf("adapter did not forward tenantID: got %q want %q", gotTenant, tenantA)
				}
				return []authdomain.ManagedUser{aliceActive}, nil
			},
			wantUserIDs: []string{aliceActive.UserID},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewDocumentsIAMUserOptions(tc.fake)

			got, err := adapter.ListUserOptions(context.Background(), tc.tenantID)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err: got %v want %v", err, tc.wantErr)
				}
				if got != nil {
					t.Fatalf("on error want nil slice, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}

			if tc.assertNotNil && got == nil {
				t.Fatalf("empty result must be non-nil []UserOption{}, got nil")
			}

			if len(got) != len(tc.wantUserIDs) {
				t.Fatalf("len: got %d (%+v) want %d (%v)", len(got), got, len(tc.wantUserIDs), tc.wantUserIDs)
			}
			// Build a lookup map from UserID to DisplayName for all active users in this case,
			// so we can assert each returned element has the correct DisplayName.
			allUsers := []authdomain.ManagedUser{zoeActive, aliceActive, bobInactive, mikeActive, mikeTwinActive}
			displayNameByUserID := make(map[string]string, len(allUsers))
			for _, u := range allUsers {
				displayNameByUserID[u.UserID] = u.DisplayName
			}

			for i, want := range tc.wantUserIDs {
				if got[i].UserID != want {
					t.Errorf("idx %d UserID: got %q want %q", i, got[i].UserID, want)
				}
				wantDisplayName := displayNameByUserID[want]
				if got[i].DisplayName != wantDisplayName {
					t.Errorf("idx %d DisplayName: got %q want %q", i, got[i].DisplayName, wantDisplayName)
				}
			}

			// Type guard: result must satisfy the consumer port.
			var _ docapp.IAMUserOptionsReader = adapter
		})
	}
}
