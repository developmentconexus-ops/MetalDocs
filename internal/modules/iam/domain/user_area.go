package domain

import "time"

// UserProcessArea mirrors a metaldocs.user_process_areas row: one area-scoped
// role grant for a user within a tenant. EffectiveTo nil means the grant is
// open-ended (soft-delete model, ADR 0037); a non-nil EffectiveTo marks a
// revoked/expired grant.
type UserProcessArea struct {
	UserID        string
	TenantID      string
	AreaCode      string
	Role          Role
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
	GrantedBy     *string
}

// IsActive reports whether the grant is in effect at now: EffectiveFrom has
// passed and (EffectiveTo is nil or has not yet passed). Mirrors the
// `effective_from <= now() AND effective_to IS NULL` canonical "active"
// predicate used by authz.Require's tier-2 capability query.
func (m UserProcessArea) IsActive(now time.Time) bool {
	if m.EffectiveFrom.After(now) {
		return false
	}
	if m.EffectiveTo == nil {
		return true
	}
	return m.EffectiveTo.After(now)
}
