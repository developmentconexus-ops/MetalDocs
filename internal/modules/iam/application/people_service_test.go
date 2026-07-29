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

func (stubRoles) UserActiveInTenant(_ context.Context, _, _ string) (bool, error) {
	return true, nil
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

// spyMemberships records Grant calls so the invite complement can assert the
// validation layer let every area-assignable role through to the grant seam.
type spyMemberships struct{ granted []iamdomain.Role }

func (s *spyMemberships) Grant(_ context.Context, _, _, _ string, role iamdomain.Role, _ string) error {
	s.granted = append(s.granted, role)
	return nil
}

func (s *spyMemberships) ListActive(_ context.Context, _, _ string) ([]iamdomain.UserProcessArea, error) {
	return nil, nil
}

func (s *spyMemberships) ListActiveBatch(_ context.Context, _ string, _ []string) (map[string][]iamdomain.UserProcessArea, error) {
	return map[string][]iamdomain.UserProcessArea{}, nil
}

// F-QA4-2 (invite path): an invite-time area membership becomes a
// public.user_process_areas row, whose role CHECK accepts exactly the seven
// area-assignable roles. system_admin is a CANONICAL role, so it passes
// IsValidRole and used to reach the database, which rejected it as 23514 and
// surfaced a 500. Invite must reject it at the app layer with the same
// ErrUnknownRole sentinel the membership route uses (→ 400 UNKNOWN_ROLE), and
// must do so BEFORE any identity is created.
func TestInvite_SystemAdminIsNotAreaAssignable(t *testing.T) {
	spy := &spyInvalidator{}
	svc := iamapp.PeopleServiceFromInterfaces(
		&stubAuth{},
		stubRoles{},
		iammemory.NewRoleAdminRepository(),
		&spyMemberships{},
		iamapp.PermissiveAreaCatalog{},
		spy,
	)

	_, err := svc.Invite(context.Background(), "tenant-a", "admin", iamapp.InviteInput{
		Username:    "newbie",
		DisplayName: "New Bie",
		TenantRole:  iamdomain.RoleAuthor,
		AreaMemberships: []iamapp.InviteAreaInput{
			{AreaCode: "A1", Role: iamdomain.RoleSystemAdmin},
		},
	})
	if !errors.Is(err, iamapp.ErrUnknownRole) {
		t.Fatalf("expected ErrUnknownRole for system_admin membership, got %v", err)
	}
	// Validation runs before any write: no identity created, no cache flush.
	if len(spy.calls) != 0 {
		t.Fatalf("invite must reject before creating the user, got invalidator calls %v", spy.calls)
	}
}

// TestInvite_UnknownRole pins that a non-canonical role string fails on the
// same seam, so the vocabulary gate is not system_admin-specific.
func TestInvite_UnknownRole(t *testing.T) {
	svc := iamapp.PeopleServiceFromInterfaces(
		&stubAuth{},
		stubRoles{},
		iammemory.NewRoleAdminRepository(),
		&spyMemberships{},
		iamapp.PermissiveAreaCatalog{},
		nil,
	)

	_, err := svc.Invite(context.Background(), "tenant-a", "admin", iamapp.InviteInput{
		Username:    "newbie",
		DisplayName: "New Bie",
		TenantRole:  iamdomain.RoleAuthor,
		AreaMemberships: []iamapp.InviteAreaInput{
			{AreaCode: "A1", Role: iamdomain.Role("ghost")},
		},
	})
	if !errors.Is(err, iamapp.ErrUnknownRole) {
		t.Fatalf("expected ErrUnknownRole, got %v", err)
	}
}

// The complement: every area-assignable role must still be invitable, so the
// narrowing above cannot be over-tightened into a regression. The old hand
// predicate accepted only signer|area_admin|qms_admin — this pins all seven.
func TestInvite_AcceptsEveryAreaAssignableRole(t *testing.T) {
	roles := iamdomain.AreaRoles()
	if len(roles) != 7 {
		t.Fatalf("area-assignable roles = %d, want 7: %v", len(roles), roles)
	}
	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			memberships := &spyMemberships{}
			svc := iamapp.PeopleServiceFromInterfaces(
				&stubAuth{},
				stubRoles{},
				iammemory.NewRoleAdminRepository(),
				memberships,
				iamapp.PermissiveAreaCatalog{},
				nil,
			)

			_, err := svc.Invite(context.Background(), "tenant-a", "admin", iamapp.InviteInput{
				Username:    "newbie",
				DisplayName: "New Bie",
				TenantRole:  iamdomain.RoleAuthor,
				AreaMemberships: []iamapp.InviteAreaInput{
					{AreaCode: "A1", Role: role},
				},
			})
			if err != nil {
				t.Fatalf("area role %q was rejected: %v", role, err)
			}
			if len(memberships.granted) != 1 || memberships.granted[0] != role {
				t.Fatalf("grants = %v, want exactly [%s]", memberships.granted, role)
			}
		})
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
