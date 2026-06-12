package application_test

import (
	"context"
	"errors"
	"testing"

	authdomain "metaldocs/internal/modules/auth/domain"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iammemory "metaldocs/internal/modules/iam/infrastructure/memory"
)

// stubAuth returns a deterministic ManagedUser slice from ListUsers.
type stubAuth struct{ users []authdomain.ManagedUser }

func (s *stubAuth) CreateUserWithInput(context.Context, authdomain.CreateUserInput) error {
	return nil
}
func (s *stubAuth) UpdateUser(context.Context, authdomain.UpdateUserParams, string) error {
	return nil
}
func (s *stubAuth) AdminResetPassword(context.Context, string, string) error { return nil }
func (s *stubAuth) UnlockUser(context.Context, string) error                 { return nil }
func (s *stubAuth) ListUsers(_ context.Context, _ string) ([]authdomain.ManagedUser, error) {
	return s.users, nil
}

type stubRoles struct{}

func (stubRoles) RolesByUserID(_ context.Context, _, _ string) ([]iamdomain.Role, error) {
	return []iamdomain.Role{iamdomain.RoleAuthor}, nil
}

func (stubRoles) RolesByUserIDs(_ context.Context, _ string, userIDs []string) (map[string][]iamdomain.Role, error) {
	out := make(map[string][]iamdomain.Role, len(userIDs))
	for _, uid := range userIDs {
		out[uid] = []iamdomain.Role{iamdomain.RoleAuthor}
	}
	return out, nil
}

// spyInvalidator records InvalidateUserTenant calls so the create-path guard
// (A3) can assert Invite flushes the role cache after the user is created.
type spyInvalidator struct {
	calls [][2]string
}

func (s *spyInvalidator) InvalidateUserTenant(userID, tenantID string) {
	s.calls = append(s.calls, [2]string{userID, tenantID})
}

// TestInvite_InvalidatesRoleCache asserts that provisioning a user (which
// assigns the tenant role) flushes that user's cached roles, so the freshly
// granted role is authoritative immediately rather than after the cache TTL (A3).
func TestInvite_InvalidatesRoleCache(t *testing.T) {
	spy := &spyInvalidator{}
	svc := iamapp.PeopleServiceFromInterfaces(
		&stubAuth{},
		stubRoles{},
		iammemory.NewRoleAdminRepository(),
		nil,
		iamapp.PermissiveAreaCatalog{},
		spy,
	)

	_, err := svc.Invite(context.Background(), "tenant-a", "admin", iamapp.InviteInput{
		Username:    "newbie",
		DisplayName: "New Bie",
		TenantRole:  iamdomain.RoleAuthor,
	})
	if err != nil {
		t.Fatalf("Invite: %v", err)
	}
	if want := [][2]string{{"newbie", "tenant-a"}}; len(spy.calls) != 1 || spy.calls[0] != want[0] {
		t.Fatalf("invalidator calls = %v, want %v", spy.calls, want)
	}
}

func TestListFiltered_ReturnsCursorExpiredWhenAnchorMissing(t *testing.T) {
	auth := &stubAuth{users: []authdomain.ManagedUser{
		{UserID: "alice", Username: "alice", DisplayName: "Alice", IsActive: true},
		{UserID: "bob", Username: "bob", DisplayName: "Bob", IsActive: true},
	}}
	svc := iamapp.PeopleServiceFromInterfaces(
		auth,
		stubRoles{},
		iammemory.NewRoleAdminRepository(),
		nil,
		iamapp.PermissiveAreaCatalog{},
		nil,
	)

	missing := "ghost-user"
	_, err := svc.ListFiltered(context.Background(), "tenant-a", iamapp.ListFilters{Cursor: &missing})
	if !errors.Is(err, iamapp.ErrCursorExpired) {
		t.Fatalf("expected ErrCursorExpired for stale cursor anchor, got %v", err)
	}
}

func TestListFiltered_HappyCursorReturnsNextPage(t *testing.T) {
	auth := &stubAuth{users: []authdomain.ManagedUser{
		{UserID: "alice", Username: "alice", DisplayName: "Alice", IsActive: true},
		{UserID: "bob", Username: "bob", DisplayName: "Bob", IsActive: true},
		{UserID: "carol", Username: "carol", DisplayName: "Carol", IsActive: true},
	}}
	svc := iamapp.PeopleServiceFromInterfaces(
		auth,
		stubRoles{},
		iammemory.NewRoleAdminRepository(),
		nil,
		iamapp.PermissiveAreaCatalog{},
		nil,
	)

	cursor := "alice"
	res, err := svc.ListFiltered(context.Background(), "tenant-a", iamapp.ListFilters{Cursor: &cursor, Limit: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Items) != 1 || res.Items[0].UserID != "bob" {
		t.Fatalf("expected bob as next page, got %+v", res.Items)
	}
}
