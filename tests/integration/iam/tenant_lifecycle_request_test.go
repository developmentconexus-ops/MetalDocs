//go:build integration

// tenant_lifecycle_request_test.go — M7 F7.3 Task E's named integration-test
// deliverable: "enqueue path: cap denied / unknown tenant 404 / job row +
// outbox row" (the "outbox row" here is the same-tx River river_job row —
// see the package doc comment on tenant_lifecycle_service.go for the
// outbox-vs-River architectural finding: tenant_lifecycle_jobs IS the
// outbox/dedup record, paired with a same-tx river.InsertTx, no separate
// metaldocs.outbox_events row for this feature).
//
// Mirrors onboard_tenant_test.go's real-Postgres idiom (openDB/seedIdentity/
// seedSystemAdminRole/db.NewTxRunner) plus river/client_test.go's
// riverjobs.NewClientBundle construction for a real *river.Client[*sql.Tx].
//
// go test -tags=integration ./tests/integration/iam/... -run TestRequestTenantLifecycle
package iam_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	auditpg "metaldocs/internal/modules/audit/infrastructure/postgres"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	iamjobs "metaldocs/internal/modules/iam/jobs"
	"metaldocs/internal/platform/db"
	riverjobs "metaldocs/internal/platform/jobs/river"
	"metaldocs/tests/integration/testdb"
)

func newTenantLifecycleServiceForTest(t *testing.T, sqlDB *sql.DB) *iamapp.TenantLifecycleService {
	t.Helper()
	bundle, err := riverjobs.NewClientBundle(sqlDB, riverjobs.Config{SkipUnknownJobCheck: true}, nil)
	if err != nil {
		t.Fatalf("new river client bundle: %v", err)
	}
	tenantRepo := iampg.NewTenantRepository(sqlDB)
	lifecycleRepo := iampg.NewTenantLifecycleRepository(sqlDB)
	return iamapp.NewTenantLifecycleService(
		tenantRepo,
		lifecycleRepo,
		lifecycleRepo,
		iamjobs.NewTenantLifecycleEnqueuer(bundle.Client),
		db.NewTxRunner(sqlDB),
		auditpg.NewWriter(sqlDB),
		sqlDB,
		nil, // ports: not exercised by the enqueue-side test (RunJob is not called here)
		nil, // blobs: not exercised
		nil, // crypto: not exercised
	)
}

// TestRequestTenantLifecycle_UnknownTenantIs404 proves RequestExport returns
// iamdomain.ErrTenantNotFound (404 at the handler) for a tenant id that has
// no metaldocs.tenants row, without seeding any capability GUC beyond the
// actor's own (the lookup happens inside the SAME tx as authz.Require, per
// the task's prescribed sequence: authz.Require(actor's tenant) -> lookup
// target tenant -> 404/409 short-circuit BEFORE authz.SeedTxTenant(target)).
func TestRequestTenantLifecycle_UnknownTenantIs404(t *testing.T) {
	sqlDB := openDB(t)
	ctx := context.Background()

	actor := testdb.DeterministicID(t, "tenant-lifecycle-404-actor")
	seedIdentity(t, sqlDB, actor)
	seedSystemAdminRole(t, sqlDB, actor)

	svc := newTenantLifecycleServiceForTest(t, sqlDB)

	unknownTenantID := "00000000-0000-0000-0000-000000000999"
	_, err := svc.RequestExport(ctx, unknownTenantID, actor, devTenant)
	if !isTenantNotFound(err) {
		t.Fatalf("RequestExport(unknown tenant) = %v, want iamdomain.ErrTenantNotFound", err)
	}
}

// TestRequestTenantLifecycle_CapDeniedBeforeAnyWrite proves an actor without
// tenant.export/tenant.erase is denied via authz.ErrCapDenied and that no
// tenant_lifecycle_jobs row (nor river_job row) is created — the tier-2
// authz.Require call happens before the tenant lookup or any INSERT. Target
// tenant is a freshly onboarded one (not devTenant, whose metaldocs.tenants
// row this suite does not assume) — the FK on tenant_lifecycle_jobs.tenant_id
// is irrelevant here anyway since the denial must short-circuit before that
// lookup/INSERT is reached.
func TestRequestTenantLifecycle_CapDeniedBeforeAnyWrite(t *testing.T) {
	sqlDB := openDB(t)
	ctx := context.Background()

	targetTenantID := onboardTenantForLifecycleTest(t, sqlDB, "tenant-lifecycle-denied")

	// Seeded with NO role at all — holds zero capabilities anywhere.
	actor := testdb.DeterministicID(t, "tenant-lifecycle-denied-actor")
	seedIdentity(t, sqlDB, actor)

	svc := newTenantLifecycleServiceForTest(t, sqlDB)

	_, err := svc.RequestExport(ctx, targetTenantID, actor, devTenant)
	if !isCapDenied(err) {
		t.Fatalf("RequestExport(no capability) = %v, want authz.ErrCapDenied", err)
	}

	var jobCount int
	if err := sqlDB.QueryRowContext(ctx,
		`SELECT count(*) FROM metaldocs.tenant_lifecycle_jobs WHERE requested_by = $1`, actor,
	).Scan(&jobCount); err != nil {
		t.Fatalf("count tenant_lifecycle_jobs: %v", err)
	}
	if jobCount != 0 {
		t.Fatalf("tenant_lifecycle_jobs rows for denied actor = %d, want 0 (authz.Require must short-circuit before any INSERT)", jobCount)
	}
}

// TestRequestTenantLifecycle_SuccessCreatesJobRowAndRiverJob proves the full
// enqueue-side happy path: a system_admin actor requesting export against
// their OWN (freshly onboarded) tenant gets back a job id, and BOTH the
// tenant_lifecycle_jobs row and the same-tx river_job row exist afterward —
// the outbox-via-River pairing this feature relies on (no
// metaldocs.outbox_events row is expected; see package doc comment).
func TestRequestTenantLifecycle_SuccessCreatesJobRowAndRiverJob(t *testing.T) {
	sqlDB := openDB(t)
	ctx := context.Background()

	// The onboarding actor becomes the new tenant's system_admin (F7.2
	// OnboardTenantService grants system_admin atomically), so it can also
	// act as the export-requesting actor here without extra role seeding.
	onboardActor := testdb.DeterministicID(t, "tenant-lifecycle-ok-onboarder")
	seedIdentity(t, sqlDB, onboardActor)
	seedSystemAdminRole(t, sqlDB, onboardActor)

	slug := "tenant-lifecycle-ok-" + testdb.DeterministicID(t, "tenant-lifecycle-ok-slug")
	adminUserID := testdb.DeterministicID(t, "tenant-lifecycle-ok-admin")
	t.Cleanup(func() { cleanupOnboardedTenant(sqlDB, slug, adminUserID) })

	onboardSvc := newOnboardTenantService(sqlDB)
	result, err := onboardSvc.OnboardTenant(ctx, iamapp.OnboardTenantInput{
		Name:             "Tenant Lifecycle OK Tenant",
		Slug:             slug,
		AdminUserID:      adminUserID,
		AdminDisplayName: "Tenant Lifecycle OK Admin",
		AdminPassword:    "correct-horse-battery-staple-lifecycle",
	}, onboardActor, devTenant)
	if err != nil {
		t.Fatalf("OnboardTenant() (test setup) = %v, want nil", err)
	}
	targetTenantID := result.TenantID

	// The export request below inserts a tenant_lifecycle_jobs row whose FK
	// references the tenants row — without this cleanup (LIFO: runs BEFORE
	// cleanupOnboardedTenant above) the tenants DELETE fails silently inside
	// withBypassErr and the deterministic slug leaks into the shared dev DB,
	// breaking the next run's OnboardTenant with "tenant slug already exists".
	t.Cleanup(func() {
		_ = withBypassErr(sqlDB, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(context.Background(),
				`DELETE FROM metaldocs.tenant_lifecycle_jobs WHERE tenant_id = $1::uuid`, targetTenantID)
			return err
		})
	})

	svc := newTenantLifecycleServiceForTest(t, sqlDB)

	// adminUserID holds system_admin in targetTenantID (granted by
	// OnboardTenant) and is requesting export against that SAME tenant —
	// actorTenantID == targetTenantID here, unlike the cross-tenant denial
	// case above.
	jobID, err := svc.RequestExport(ctx, targetTenantID, adminUserID, targetTenantID)
	if err != nil {
		t.Fatalf("RequestExport() = %v, want nil", err)
	}
	if jobID == "" {
		t.Fatalf("RequestExport() job id is empty")
	}

	var kind, status, requestedBy string
	if err := sqlDB.QueryRowContext(ctx,
		`SELECT kind, status, requested_by FROM metaldocs.tenant_lifecycle_jobs WHERE id = $1::uuid`, jobID,
	).Scan(&kind, &status, &requestedBy); err != nil {
		t.Fatalf("read tenant_lifecycle_jobs row: %v", err)
	}
	if kind != "export" {
		t.Fatalf("tenant_lifecycle_jobs.kind = %q, want %q", kind, "export")
	}
	if status != "pending" {
		t.Fatalf("tenant_lifecycle_jobs.status = %q, want %q", status, "pending")
	}
	if requestedBy != adminUserID {
		t.Fatalf("tenant_lifecycle_jobs.requested_by = %q, want %q", requestedBy, adminUserID)
	}

	var riverJobCount int
	if err := sqlDB.QueryRowContext(ctx,
		`SELECT count(*) FROM river_job WHERE kind = 'tenant_lifecycle' AND args->>'job_id' = $1`, jobID,
	).Scan(&riverJobCount); err != nil {
		t.Fatalf("count river_job rows: %v", err)
	}
	if riverJobCount != 1 {
		t.Fatalf("river_job rows for job %s = %d, want 1 (same-tx InsertTx pairing)", jobID, riverJobCount)
	}

	var auditCount int
	if err := sqlDB.QueryRowContext(ctx,
		// The event's resource is the TENANT (resource_id = tenant id); the job
		// id travels in the payload (buildLifecycleRequestedEvent).
		`SELECT count(*) FROM metaldocs.audit_events WHERE action = 'tenant.lifecycle.requested' AND resource_id = $1 AND payload->>'job_id' = $2`, targetTenantID, jobID,
	).Scan(&auditCount); err != nil {
		t.Fatalf("count audit_events rows: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("audit_events tenant.lifecycle.requested rows for job %s = %d, want 1", jobID, auditCount)
	}
}

func isTenantNotFound(err error) bool {
	return errors.Is(err, iamdomain.ErrTenantNotFound)
}

// onboardTenantForLifecycleTest onboards a throwaway tenant (via the real
// F7.2 OnboardTenantService, same as the success-path test) purely to obtain
// a valid metaldocs.tenants row id to target — used where the test does not
// care about the tenant's own admin/roles (e.g. the cap-denied case, where
// the actor is denied before the target tenant's identity matters at all).
func onboardTenantForLifecycleTest(t *testing.T, sqlDB *sql.DB, namePrefix string) string {
	t.Helper()
	ctx := context.Background()

	onboardActor := testdb.DeterministicID(t, namePrefix+"-onboarder")
	seedIdentity(t, sqlDB, onboardActor)
	seedSystemAdminRole(t, sqlDB, onboardActor)

	slug := namePrefix + "-" + testdb.DeterministicID(t, namePrefix+"-slug")
	adminUserID := testdb.DeterministicID(t, namePrefix+"-admin")
	t.Cleanup(func() { cleanupOnboardedTenant(sqlDB, slug, adminUserID) })

	onboardSvc := newOnboardTenantService(sqlDB)
	result, err := onboardSvc.OnboardTenant(ctx, iamapp.OnboardTenantInput{
		Name:             namePrefix + " Tenant",
		Slug:             slug,
		AdminUserID:      adminUserID,
		AdminDisplayName: namePrefix + " Admin",
		AdminPassword:    "correct-horse-battery-staple-" + namePrefix,
	}, onboardActor, devTenant)
	if err != nil {
		t.Fatalf("OnboardTenant() (test setup for %s) = %v, want nil", namePrefix, err)
	}
	return result.TenantID
}
