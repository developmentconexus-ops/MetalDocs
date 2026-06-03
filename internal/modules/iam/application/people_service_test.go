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
func (s *stubAuth) UnlockUser(context.Context, string) error                  { return nil }
func (s *stubAuth) ListUsers(_ context.Context, _ string) ([]authdomain.ManagedUser, error) {
	return s.users, nil
}

type stubRoles struct{}

func (stubRoles) RolesByUserID(_ context.Context, _, _ string) ([]iamdomain.Role, error) {
	return []iamdomain.Role{iamdomain.RoleAuthor}, nil
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
