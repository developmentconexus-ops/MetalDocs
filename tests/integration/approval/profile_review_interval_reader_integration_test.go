//go:build integration
// +build integration

// Package approval_test — ADR 0085 review-cadence read port, wired to the REAL
// taxonomy repository. The coordinator's own suite stubs the port
// (fixedReviewInterval) so the review-cycle rule stays assertable without a
// seeded cadence; that stub cannot catch a context contract mismatch in the
// adapter, which is precisely what broke the release_evaluate job in
// production. This file covers the adapter itself under the job's real context
// shape.
package approval_test

import (
	"context"
	"testing"

	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	taxonomyrepo "metaldocs/internal/modules/taxonomy/infrastructure"
	"metaldocs/tests/integration/testdb"
)

// TestProfileReviewIntervalReader_BarrenBackgroundContext reproduces the live
// defect shape: the River release_evaluate worker marks its ctx with
// authz.WithBackgroundBypass and nothing else — no tenant, no actor — so the
// preflight's review-interval read must resolve the tenant from its explicit
// argument. Before the system read path this failed with "taxonomy: tenant
// context: tenant: not present in context".
func TestProfileReviewIntervalReader_BarrenBackgroundContext(t *testing.T) {
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tax := testdb.NewTaxonomy(t, db)
	reader := approvalrepo.NewProfileReviewIntervalReader(taxonomyrepo.NewProfileRepository(db))

	bgCtx := authz.WithBackgroundBypass(context.Background())

	days, err := reader.ReviewIntervalDays(bgCtx, tax.TenantID, tax.ProfileCode)
	if err != nil {
		t.Fatalf("ReviewIntervalDays under barren background ctx: %v", err)
	}
	if days != 365 {
		t.Fatalf("ReviewIntervalDays = %d, want 365 (NewTaxonomy cadence)", days)
	}
}
