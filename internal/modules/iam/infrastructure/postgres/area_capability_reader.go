package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// AreaCapabilityReaderPG satisfies the iam-owned published port.
var _ iamdomain.AreaCapabilityReader = (*AreaCapabilityReaderPG)(nil)

// AreaCapabilityReaderPG is the Postgres-backed
// iamdomain.AreaCapabilityReader. It is the single cross-module home for the
// "which areas can this actor do X in?" read: the role_capabilities ⋈
// user_process_areas join lives here, never in a consumer module. The predicate
// is the all-areas form of the per-area grant check inside authz.Require, and
// the system_admin term reuses authz.SystemAdminExistsSQL so the two cannot
// drift (same DRY fix as UserAreaRepository.MembershipDirectoryScope).
type AreaCapabilityReaderPG struct {
	db *sql.DB
}

// NewAreaCapabilityReaderPG constructs a pool-backed AreaCapabilityReaderPG.
// db must be non-nil.
func NewAreaCapabilityReaderPG(db *sql.DB) *AreaCapabilityReaderPG {
	if db == nil {
		panic("iam/infrastructure/postgres: NewAreaCapabilityReaderPG requires a non-nil db")
	}
	return &AreaCapabilityReaderPG{db: db}
}

// AreasWithCapability resolves the actor's capability reach in one round trip:
// the system_admin tenant-wide bypass flag plus the distinct area codes with an
// active grant carrying capability.
func (r *AreaCapabilityReaderPG) AreasWithCapability(ctx context.Context, tenantID, actorUserID, capability string, now time.Time) (bool, []string, error) {
	// Arg order is positional and shared with SystemAdminExistsSQL:
	// $1=actorUserID, $2=tenantID, $3=capability, $4=now. Keep aligned if the
	// const changes.
	//
	// The subquery gates on `upa.effective_to IS NULL` — the canonical
	// active-now membership predicate under the soft-delete model (ADR 0037),
	// identical to MembershipDirectoryScope. Not an interval bug.
	const q = `
SELECT
  ` + authz.SystemAdminExistsSQL + ` AS tenant_wide,
  ARRAY(
    SELECT DISTINCT upa.area_code
      FROM metaldocs.role_capabilities rc
      JOIN public.user_process_areas upa
        ON upa.role = rc.role
       AND upa.tenant_id = $2::uuid
       AND upa.user_id   = $1
       AND upa.effective_from <= $4
       AND upa.effective_to IS NULL
     WHERE rc.capability = $3
     ORDER BY upa.area_code ASC
  ) AS area_codes
`
	var tenantWide bool
	var areaCodes []string
	if err := r.db.QueryRowContext(ctx, q, actorUserID, tenantID, capability, now).
		Scan(&tenantWide, pq.Array(&areaCodes)); err != nil {
		return false, nil, fmt.Errorf("iam: resolve areas with capability: %w", err)
	}
	if areaCodes == nil {
		areaCodes = []string{}
	}
	return tenantWide, areaCodes, nil
}
