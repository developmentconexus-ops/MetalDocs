//go:build integration

package domain_test

import (
	"context"
	"sync"
	"testing"

	"metaldocs/internal/modules/controlleddocuments/infrastructure"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// This test previously hand-rolled its own database handle: it read DATABASE_URL
// (or METALDOCS_DATABASE_URL) straight from the environment, sql.Open'd the
// SHARED dev database, and t.Skip'd when the variable was unset or unreachable.
// Every part of that was a defect (ADR 0034 — the testdb factory is the only
// sanctioned way to get an integration database):
//
//   - It skipped rather than failed when the env was not set, so a run in which
//     this test never executed was indistinguishable from a run in which it
//     passed. It was green because it was absent.
//   - It ran against the shared dev database, which is not migrated by the test
//     harness. It was passing against a schema 47 migrations stale — that is not
//     a passing test, it is a test asserting against the wrong schema.
//   - Sharing a database with every other test forced workarounds INTO the test:
//     a timestamped profile code to dodge PK collisions, and ON CONFLICT DO
//     NOTHING on every seed to tolerate rows it did not create.
//
// testdb.Open(t) hands over a private leased database built from the migrated
// template, so all three disappear: the run fails loudly if postgres is
// unreachable, the schema is current by construction, and the seeds are plain
// inserts of rows this test provably owns. The workarounds are removed rather
// than kept "just in case" — carrying them would preserve the shared-DB
// assumption in a test that no longer shares anything, and the next reader would
// reasonably infer isolation is not guaranteed here.
//
// Reset-safe, so Open and not OpenFreshDatabase: it does no DDL, creates no
// roles, and touches no database-level settings — only ordinary DML on
// document_families / document_profiles / document_process_areas /
// cd_sequence_counters, all of which the lease reset empties.
func TestSequenceAllocatorNextAndIncrement_Concurrent(t *testing.T) {
	db, _ := testdb.Open(t)

	// tenant.DevTenantID is the 'System Tenant' row the curated bootstrap seeds
	// into the template, and metaldocs.tenants is snapshot-restored by the lease
	// reset — so it is present on every lease, not merely usually present.
	tenantID := tenant.DevTenantID
	// Deterministic, no timestamp: document_profiles' primary key is on
	// (tenant_id, code), so on the old shared database two runs in the same
	// second under the same tenant collided. On a private lease this test is
	// the only writer.
	profileCode := "seqtest"
	areaCode := "rh"

	// The taxonomy tables (document_profiles, document_process_areas) and the
	// cd_sequence_counters table each carry the capability tripwire
	// (trg_require_cap_asserted): document_profiles/process_areas require
	// taxonomy.manage, cd_sequence_counters requires controlled_documents.create.
	// The service layer normally asserts these tx-locally via authz.Require; this
	// raw-SQL setup must do the same or every INSERT trips P0001. Seed all three
	// rows in ONE tx that asserts both caps. This also replaces the old
	// allocator.EnsureCounter call, which ran on the pool (autocommit) and so
	// could never carry a tx-local assertion.
	setupTx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin setup tx: %v", err)
	}
	// Assert the setup caps tx-locally via the sanctioned testdb helper (no inline
	// tripwire config in the test file — R1). Same effect the service layer's
	// authz.Require has: the asserted-caps GUC set transaction-locally.
	testdb.SetCapsOnTx(t, setupTx, `[{"cap":"taxonomy.manage"},{"cap":"controlled_documents.create"}]`)
	// document_profiles.family_code FKs document_families(code), and the template
	// seeds no families, so this test provisions its own FK parent. It carries
	// trg_require_cap_asserted (taxonomy.manage), already asserted on this tx.
	// Plain INSERT, no ON CONFLICT: on a lease no other writer exists, so a
	// duplicate here would mean the reset failed to empty the table and must fail
	// the test rather than be swallowed.
	if _, err := setupTx.ExecContext(context.Background(), `
		INSERT INTO metaldocs.document_families
			(code, tenant_id, name)
		VALUES
			('quality', $1, 'Quality')`,
		tenantID,
	); err != nil {
		_ = setupTx.Rollback()
		t.Fatalf("seed family: %v", err)
	}
	if _, err := setupTx.ExecContext(context.Background(), `
		INSERT INTO metaldocs.document_profiles
			(code, tenant_id, family_code, name, description, review_interval_days, editable_by_role, alias)
		VALUES
			($1, $2, 'quality', 'Seq Test', 'integration sequence test', 30, 'admin', $1)`,
		profileCode, tenantID,
	); err != nil {
		_ = setupTx.Rollback()
		t.Fatalf("insert profile: %v", err)
	}
	if _, err := setupTx.ExecContext(context.Background(), `
		INSERT INTO metaldocs.document_process_areas
			(code, tenant_id, name, description)
		VALUES
			($1, $2, 'Resources & HR', 'integration test area')`,
		areaCode, tenantID,
	); err != nil {
		_ = setupTx.Rollback()
		t.Fatalf("insert process area: %v", err)
	}
	if _, err := setupTx.ExecContext(context.Background(), `
		INSERT INTO public.cd_sequence_counters
			(tenant_id, profile_code, process_area_code, next_seq)
		VALUES
			($1, $2, $3, 1)`,
		tenantID, profileCode, areaCode,
	); err != nil {
		_ = setupTx.Rollback()
		t.Fatalf("seed counter: %v", err)
	}
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("commit setup tx: %v", err)
	}

	allocator := infrastructure.NewPostgresSequenceAllocator(db)

	const workers = 50
	results := make([]int, 0, workers)
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			// Each allocation runs in its own committed tx — the production shape.
			// The FOR UPDATE lock is held until Commit, serializing the 50
			// concurrent workers into a gap-free 1..50 (the allocator now rejects
			// a nil tx, so autocommit is no longer an option).
			tx, err := db.BeginTx(context.Background(), nil)
			if err != nil {
				t.Errorf("begin tx: %v", err)
				return
			}
			// NextAndIncrement writes cd_sequence_counters (ensure-INSERT + locked
			// UPDATE), both gated by the controlled_documents.create tripwire — the
			// service asserts it tx-locally, so the test must too. Sanctioned helper,
			// no inline set_config (R1).
			testdb.SetCapsOnTx(t, tx, `[{"cap":"controlled_documents.create"}]`)
			next, err := allocator.NextAndIncrement(context.Background(), tx, tenantID, profileCode, areaCode)
			if err != nil {
				_ = tx.Rollback()
				t.Errorf("next and increment: %v", err)
				return
			}
			if err := tx.Commit(); err != nil {
				t.Errorf("commit: %v", err)
				return
			}
			mu.Lock()
			results = append(results, next)
			mu.Unlock()
		}()
	}
	wg.Wait()

	if len(results) != workers {
		t.Fatalf("expected %d results, got %d", workers, len(results))
	}

	seen := map[int]bool{}
	for _, v := range results {
		if seen[v] {
			t.Fatalf("duplicate sequence value found: %d", v)
		}
		seen[v] = true
	}
	for i := 1; i <= workers; i++ {
		if !seen[i] {
			t.Fatalf("missing sequence value %d in %v", i, results)
		}
	}
}
