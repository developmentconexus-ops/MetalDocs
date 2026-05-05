package application

import (
	"context"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// CapabilityChecker is the documents-module consumer port for tier-1
// (tenant-scoped) capability checks. Production wiring binds this to
// *iam/application.CapabilityService. See wiki/decisions/0007-two-tier-authz.md.
type CapabilityChecker interface {
	CanDo(ctx context.Context, userID, tenantID string, cap iamdomain.Capability) error
}
