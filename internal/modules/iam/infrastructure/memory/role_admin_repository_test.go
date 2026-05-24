package memory

import (
	"context"
	"testing"

	"metaldocs/internal/modules/iam/domain"
)

func TestReplaceUserRoles_PersistsAllRolesAndNormalisesInput(t *testing.T) {
	repo := NewRoleAdminRepository()
	ctx := context.Background()

	err := repo.ReplaceUserRoles(ctx, "alice", "Alice", "tenant-a", []domain.Role{
		domain.RoleEditor,
		"",
		" ",
		domain.RoleSystemAdmin,
		domain.RoleEditor,
		" system_admin ",
	}, "admin")
	if err != nil {
		t.Fatalf("ReplaceUserRoles: %v", err)
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	rec, ok := repo.users["alice"]
	if !ok {
		t.Fatal("expected user record")
	}
	if !rec.roles[domain.RoleEditor] {
		t.Fatal("expected editor role")
	}
	if !rec.roles[domain.RoleSystemAdmin] {
		t.Fatal("expected system_admin role")
	}
	if rec.roles[domain.Role("")] {
		t.Fatal("expected empty role to be ignored")
	}
	if len(rec.roles) != 2 {
		t.Fatalf("expected exactly 2 roles, got %d: %#v", len(rec.roles), rec.roles)
	}
}
