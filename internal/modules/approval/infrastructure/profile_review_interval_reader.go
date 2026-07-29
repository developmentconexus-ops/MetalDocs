package infrastructure

import (
	"context"

	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

// taxonomySystemProfileGetter is the narrow background read slice of the
// taxonomy profile repository. It is deliberately separate from
// taxonomyProfileGetter: the system read seeds its tenant from the explicit
// parameter and is fail-closed outside a background context, so depending on
// the slice keeps the request-path reader and the job-path reader from sharing
// a surface (interface segregation).
type taxonomySystemProfileGetter interface {
	GetByCodeSystem(ctx context.Context, tenantID string, code taxonomydomain.ProfileCode) (*taxonomydomain.DocumentProfile, error)
}

// ProfileReviewIntervalReader adapts the canonical taxonomy profile repository
// to the approval application.ProfileReviewIntervalReader port (ADR 0085).
//
// Its ONLY consumer is the release coordinator's preflight, which runs in the
// River release_evaluate job and therefore carries no HTTP identity — hence the
// system read slice rather than the identity-gated GetByCode. The read still
// runs in taxonomy's own short transaction on a separate connection, so it never
// sits inside an approval lock-holding tx (H-PRE-1) — which is exactly why the
// coordinator calls it in its off-lock preflight and never inside the release
// transaction.
type ProfileReviewIntervalReader struct {
	profiles taxonomySystemProfileGetter
}

// NewProfileReviewIntervalReader builds the adapter over the canonical
// taxonomy profile repository.
func NewProfileReviewIntervalReader(profiles taxonomySystemProfileGetter) *ProfileReviewIntervalReader {
	return &ProfileReviewIntervalReader{profiles: profiles}
}

// ReviewIntervalDays returns the profile's periodic-review cadence in days.
// A read error propagates unchanged so the caller fails closed rather than
// silently releasing a document with no review cycle.
func (r *ProfileReviewIntervalReader) ReviewIntervalDays(ctx context.Context, tenantID, profileCode string) (int, error) {
	profile, err := r.profiles.GetByCodeSystem(ctx, tenantID, taxonomydomain.ProfileCode(profileCode))
	if err != nil {
		return 0, err
	}
	if profile == nil {
		return 0, nil
	}
	return profile.ReviewIntervalDays, nil
}
