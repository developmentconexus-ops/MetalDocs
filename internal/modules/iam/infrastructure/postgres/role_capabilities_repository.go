package postgres

import (
	"context"
	"database/sql"
	"fmt"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// RoleCapabilitiesRepository reads the global role↔capability matrix from
// metaldocs.role_capabilities. The matrix is tenant-agnostic; the HTTP edge
// still requires authenticated tenant context per Admin Center middleware
// policy.
type RoleCapabilitiesRepository struct {
	db *sql.DB
}

func NewRoleCapabilitiesRepository(db *sql.DB) *RoleCapabilitiesRepository {
	return &RoleCapabilitiesRepository{db: db}
}

// ListRoleCapabilities returns every (role, capability) row from
// metaldocs.role_capabilities, ordered for deterministic output.
func (r *RoleCapabilitiesRepository) ListRoleCapabilities(ctx context.Context) ([]iamdomain.RoleCapabilityLink, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("role_capabilities repository: db is nil")
	}
	const query = `
SELECT role, capability
  FROM metaldocs.role_capabilities
 ORDER BY role, capability
`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query role_capabilities: %w", err)
	}
	defer rows.Close()

	out := make([]iamdomain.RoleCapabilityLink, 0, 64)
	for rows.Next() {
		var role, capability string
		if err := rows.Scan(&role, &capability); err != nil {
			return nil, fmt.Errorf("scan role_capabilities row: %w", err)
		}
		out = append(out, iamdomain.RoleCapabilityLink{
			Role:       iamdomain.Role(role),
			Capability: iamdomain.Capability(capability),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate role_capabilities: %w", err)
	}
	return out, nil
}
