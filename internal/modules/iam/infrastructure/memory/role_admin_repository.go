// Test fixture only; not wired by bootstrap.
package memory

import (
	"context"
	"database/sql"
	"sync"

	"metaldocs/internal/modules/iam/domain"
)

type RoleAdminRepository struct {
	mu    sync.Mutex
	users map[string]userRecord
}

type userRecord struct {
	displayName string
	roles       map[domain.Role]bool
}

func NewRoleAdminRepository() *RoleAdminRepository {
	return &RoleAdminRepository{users: map[string]userRecord{}}
}

func (r *RoleAdminRepository) HasAnyRole(_ context.Context, role domain.Role, _ string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, rec := range r.users {
		if rec.roles[role] {
			return true, nil
		}
	}
	return false, nil
}

func (r *RoleAdminRepository) UpsertUserAndAssignRole(_ context.Context, userID, displayName, _ string, role domain.Role, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Replace semantics: the Postgres implementation deletes prior roles before
	// inserting the new one, so the in-memory double must do the same (single
	// active role per user) — accumulating would silently diverge from prod.
	r.users[userID] = userRecord{displayName: displayName, roles: map[domain.Role]bool{role: true}}
	return nil
}

func (r *RoleAdminRepository) ReplaceUserRoles(_ context.Context, userID, displayName, _ string, role domain.Role, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec, ok := r.users[userID]
	if !ok {
		rec = userRecord{displayName: displayName, roles: map[domain.Role]bool{}}
	}
	rec.displayName = displayName
	rec.roles = map[domain.Role]bool{role: true}
	r.users[userID] = rec
	return nil
}

// UpsertUserAndAssignRoleTx is the tx-aware variant. In the memory repo (test
// fixture only) the *sql.Tx is ignored — no real database is involved.
func (r *RoleAdminRepository) UpsertUserAndAssignRoleTx(_ context.Context, _ *sql.Tx, userID, displayName, _ string, role domain.Role, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[userID] = userRecord{displayName: displayName, roles: map[domain.Role]bool{role: true}}
	return nil
}

// ReplaceUserRolesTx is the tx-aware variant. In the memory repo (test
// fixture only) the *sql.Tx is ignored — no real database is involved.
func (r *RoleAdminRepository) ReplaceUserRolesTx(_ context.Context, _ *sql.Tx, userID, displayName, _ string, role domain.Role, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec, ok := r.users[userID]
	if !ok {
		rec = userRecord{displayName: displayName, roles: map[domain.Role]bool{}}
	}
	rec.displayName = displayName
	rec.roles = map[domain.Role]bool{role: true}
	r.users[userID] = rec
	return nil
}
