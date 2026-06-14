package wiring

import (
	"context"

	iamapp "metaldocs/internal/modules/iam/application"
	securityapp "metaldocs/internal/modules/security/application"
)

// mfaCoveragePctAdapter narrows securityapp.Service to the
// MfaCoveragePctReader port consumed by iamapp.ObservabilityService.
// Returns 0 when the security service is nil (in-memory / dev mode).
type mfaCoveragePctAdapter struct {
	svc *securityapp.Service
}

// NewMfaCoveragePctReader returns an iamapp.MfaCoveragePctReader backed by the
// given security service. A nil svc is accepted and yields 0 coverage (dev mode).
func NewMfaCoveragePctReader(svc *securityapp.Service) iamapp.MfaCoveragePctReader {
	return mfaCoveragePctAdapter{svc: svc}
}

func (a mfaCoveragePctAdapter) MfaCoveragePct(ctx context.Context, tenantID string) (float32, error) {
	if a.svc == nil {
		return 0, nil
	}
	cov, err := a.svc.MfaCoverage(ctx, tenantID)
	if err != nil {
		return 0, err
	}
	return cov.MfaEnabledPct, nil
}
