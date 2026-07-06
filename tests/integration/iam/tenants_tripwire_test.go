//go:build integration

// tenants_tripwire_test.go — M7 F7.2 (ADR 0070) negative tripwire proof for
// the tenants/INSERT arm (migration 0277): metaldocs.tenants now carries
// trg_require_cap_asserted (tenant.onboard), BEFORE INSERT only.
//
// Two properties proven here, mirroring iam_users_tripwire_test.go's
// established direct-drive pattern (and tripwire_caps_test.go's SeedWithCaps
// idiom for the positive arm):
//
//   - insert_without_asserted_caps_rejected: a raw SQL INSERT against
//     metaldocs.tenants with no metaldocs.asserted_caps GUC set (bypassing the
//     Go authz layer entirely, e.g. a rogue migration or a future repository
//     method that forgets authz.Require) is rejected by the DB trigger itself
//     with ErrCapabilityNotAsserted (SQLSTATE P0001) — the defense-in-depth
//     property the arm exists for.
//
//   - insert_with_tenant_onboard_asserted_succeeds: the same INSERT succeeds
//     once tenant.onboard is asserted tx-locally (testdb.SeedWithCaps, exactly
//     as the production authz layer asserts caps), and the row is actually
//     written — proving the arm accepts the capability the OnboardTenant
//     workflow asserts (a wrong-cap arm would fail-close every onboarding
//     INSERT as P0001, the 0269/0270/0271/0275 defect class).
//
//     go test -tags=integration ./tests/integration/iam/... -run TestTenantsInsertTripwire
package iam_test

import (
	"context"
	"database/sql"
	"testing"

	"metaldocs/tests/integration/testdb"
)

// TestTenantsInsertTripwire drives the tenants/INSERT tripwire arm (0277)
// directly: without asserted caps the trigger fail-closes with P0001; with
// tenant.onboard asserted the provisioning INSERT lands.
func TestTenantsInsertTripwire(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()

	slug := "tripwire-probe-" + testdb.DeterministicID(t, "tenants-tripwire-slug")
	t.Cleanup(func() { cleanupTenant(db, slug) })

	t.Run("insert_without_asserted_caps_rejected", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer tx.Rollback() //nolint:errcheck

		_, err = tx.ExecContext(ctx,
			`INSERT INTO metaldocs.tenants (id, name, slug)
			 VALUES (gen_random_uuid(), 'Tripwire Probe Tenant', $1)`,
			slug,
		)
		if !isCapabilityNotAsserted(err) {
			t.Fatalf("raw INSERT into tenants without asserted_caps = %v, want ErrCapabilityNotAsserted (P0001)", err)
		}
	})

	t.Run("insert_with_tenant_onboard_asserted_succeeds", func(t *testing.T) {
		testdb.SeedWithCaps(t, db, `[{"cap":"tenant.onboard"}]`, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx,
				`INSERT INTO metaldocs.tenants (id, name, slug)
				 VALUES (gen_random_uuid(), 'Tripwire Probe Tenant', $1)`,
				slug,
			)
			return err
		})

		var count int
		if err := db.QueryRowContext(ctx,
			`SELECT count(*) FROM metaldocs.tenants WHERE slug = $1`, slug,
		).Scan(&count); err != nil {
			t.Fatalf("count tenants after asserted INSERT: %v", err)
		}
		if count != 1 {
			t.Fatalf("asserted INSERT wrote %d tenants rows, want 1", count)
		}
	})
}

// cleanupTenant removes the probe tenant row. DELETE on metaldocs.tenants is
// not gated (the 0277 trigger is BEFORE INSERT only — the arm matches the
// onboarding surface exactly; no update/offboard mutation exists yet), so a
// plain DELETE suffices. Error-discarding by design, matching the established
// cleanup convention in this package (cleanupIAMUser).
//
// The metaldocs.tenant_keys row (provisioned by a crypto-wired OnboardTenant
// — TestOnboardTenant_AuditPayloadSealedWhenCryptoWired) FKs metaldocs.tenants,
// so it must be deleted BEFORE the tenant row or the tenants DELETE fails on
// the FK and (being error-discarded) leaks the row — poisoning the next run's
// deterministic slug. Noop-provisioner tests write no tenant_keys row, so this
// extra DELETE is a harmless no-op for them.
func cleanupTenant(db *sql.DB, slug string) {
	_, _ = db.ExecContext(context.Background(),
		`DELETE FROM metaldocs.tenant_keys WHERE tenant_id IN (SELECT id FROM metaldocs.tenants WHERE slug = $1)`, slug)
	_, _ = db.ExecContext(context.Background(),
		`DELETE FROM metaldocs.tenants WHERE slug = $1`, slug)
}
