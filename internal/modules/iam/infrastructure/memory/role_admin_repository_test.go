package memory

import (
	"context"
	"testing"

	"metaldocs/internal/modules/iam/domain"
)

func TestReplaceUserRoles_ReplacesWithSingleRole(t *testing.T) {
	repo := NewRoleAdminRepository()
	ctx := context.Background()

	err := repo.ReplaceUserRoles(ctx, "alice", "Alice", "tenant-a", domain.RoleSystemAdmin, "admin")
	if err != nil {
		t.Fatalf("ReplaceUserRoles: %v", err)
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	rec, ok := repo.users["alice"]
	if !ok {
		t.Fatal("expected user record")
	}
	if !rec.roles[domain.RoleSystemAdmin] {
		t.Fatal("expected system_admin role")
	}
	if len(rec.roles) != 1 {
		t.Fatalf("expected exactly 1 role, got %d: %#v", len(rec.roles), rec.roles)
	}
}
