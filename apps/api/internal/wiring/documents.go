package wiring

import (
	"context"

	docsapp "metaldocs/internal/modules/documents/application"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// capabilityServiceAdapter bridges *iamapp.CapabilityService (takes string capability)
// to docsapp.CapabilityChecker (takes iamdomain.Capability). Both types are
// string-backed but Go requires an explicit conversion at the boundary.
type capabilityServiceAdapter struct {
	svc *iamapp.CapabilityService
}

func (a capabilityServiceAdapter) CanDo(ctx context.Context, userID, tenantID string, cap iamdomain.Capability) error {
	return a.svc.CanDo(ctx, userID, tenantID, string(cap))
}

// NewCapabilityChecker returns a docsapp.CapabilityChecker bound to the
// production CapabilityService.
func NewCapabilityChecker(svc *iamapp.CapabilityService) docsapp.CapabilityChecker {
	return capabilityServiceAdapter{svc: svc}
}
