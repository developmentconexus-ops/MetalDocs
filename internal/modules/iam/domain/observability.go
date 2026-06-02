package domain

import "time"

// PlanTier is the v1 plan label surfaced to tenant admins. The Tier-A
// platform-owner surface owns upgrade / downgrade flows; Tier-B reads only.
type PlanTier string

const (
	PlanTierFree       PlanTier = "free"
	PlanTierPro        PlanTier = "pro"
	PlanTierEnterprise PlanTier = "enterprise"
)

func (p PlanTier) Valid() bool {
	switch p {
	case PlanTierFree, PlanTierPro, PlanTierEnterprise:
		return true
	}
	return false
}

// SeatUsage carries the seat side of the Usage card.
type SeatUsage struct {
	Used      int
	Allocated int
}

// StorageUsage carries the storage side of the Usage card. UsedBytes == -1
// means the source for "used" is not yet wired — the FE renders "—" in that
// case rather than fabricating a number.
type StorageUsage struct {
	UsedBytes      int64
	AllocatedBytes int64
}

// CountWindows is the standard 24h / 7d / 30d shape reused by apiCalls and
// activeUsers. A -1 in any field has the same semantics as StorageUsage.UsedBytes.
type CountWindows struct {
	Last24h int64
	Last7d  int64
	Last30d int64
}

// UsageSnapshot is the payload returned by GET /iam/usage.
type UsageSnapshot struct {
	Seats       SeatUsage
	Storage     StorageUsage
	APICalls    CountWindows
	ActiveUsers CountWindows
	PlanTier    PlanTier // empty string when no plan row exists.
}

// RoleCount is one bucket of the role distribution histogram.
type RoleCount struct {
	Role  Role
	Count int
}

// KpiSnapshot is the payload returned by GET /iam/kpi.
type KpiSnapshot struct {
	LockedAccounts       int
	MfaCoveragePct       float32
	FailedLogins24h      int64
	DormantUsers30d      int
	RoleDistribution     []RoleCount
	AuditEventsPerMinute float32
}

// TenantPlan is the persisted plan envelope row. The Tier-B observability
// service treats it as read-only.
type TenantPlan struct {
	TenantID               string
	Tier                   PlanTier
	SeatsAllocated         int
	StorageAllocatedBytes  int64
	APICallsAllocated      int
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
