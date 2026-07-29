package infrastructure

import (
	"context"

	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

// ProfileReviewIntervalReader adapts the canonical taxonomy profile repository
// to the approval application.ProfileReviewIntervalReader port (ADR 0085).
//
// It reuses the same narrow taxonomyProfileGetter slice as ProfilePolicyReader,
// for the same reason: the read runs in taxonomy's own short transaction on a
// separate connection, so it never sits inside an approval lock-holding tx
// (H-PRE-1) — which is exactly why the release coordinator calls it in its
// off-lock preflight and never inside the release transaction.
type ProfileReviewIntervalReader struct {
	profiles taxonomyProfileGetter
}

// NewProfileReviewIntervalReader builds the adapter over the canonical
// taxonomy profile repository.
func NewProfileReviewIntervalReader(profiles taxonomyProfileGetter) *ProfileReviewIntervalReader {
	return &ProfileReviewIntervalReader{profiles: profiles}
}

// ReviewIntervalDays returns the profile's periodic-review cadence in days.
// A read error propagates unchanged so the caller fails closed rather than
// silently releasing a document with no review cycle.
func (r *ProfileReviewIntervalReader) ReviewIntervalDays(ctx context.Context, tenantID, profileCode string) (int, error) {
	profile, err := r.profiles.GetByCode(ctx, tenantID, taxonomydomain.ProfileCode(profileCode))
	if err != nil {
		return 0, err
	}
	if profile == nil {
		return 0, nil
	}
	return profile.ReviewIntervalDays, nil
}
