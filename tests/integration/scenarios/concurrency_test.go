//go:build integration
// +build integration

package scenarios_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"metaldocs/tests/integration/testdb"

	"github.com/jackc/pgx/v5/pgconn"
)

const concurrencyTestSeed = 0xDEADBEEF
const occRaceWorkers = 2

func TestConcurrencyScenarios(t *testing.T) {
	t.Logf("testSeed=0x%X", concurrencyTestSeed)

	t.Run("OCC_StaleRevision", func(t *testing.T) {
		testOCCStaleRevision(t)
	})
	t.Run("SkipLocked_NoDuplicateProcessing", func(t *testing.T) {
		testSkipLockedNoDuplicateProcessing(t)
	})
	t.Run("SignoffUnique_DuplicateBlocked", func(t *testing.T) {
		testSignoffUniqueDuplicateBlocked(t)
	})
	t.Run("OCC_Race_N50", func(t *testing.T) {
		db, schema := testdb.Open(t)
		ctx := context.Background()
		workers := occRaceWorkers
		n := 50
		if v := os.Getenv("INTEGRATION_STRESS_N"); v != "" {
			if parsed, err := fmt.Sscanf(v, "%d", &n); parsed != 1 || err != nil {
				t.Logf("invalid INTEGRATION_STRESS_N=%q, using default 50", v)
				n = 50
			}
		}
		for i := 0; i < n; i++ {
			t.Run(fmt.Sprintf("iter_%02d", i+1), func(t *testing.T) {
				winners, losers := runSingleOCCRace(t, ctx, db, schema, fmt.Sprintf("n50-%02d", i+1), workers)
				if winners != 1 || losers != workers-1 {
					t.Fatalf("expected exactly one winner and %d stale loser(s); got winners=%d losers=%d", workers-1, winners, losers)
				}
			})
		}
	})
}

func testOCCStaleRevision(t *testing.T) {
	db, schema := testdb.Open(t)
	ctx := context.Background()
	winners, losers := runSingleOCCRace(t, ctx, db, schema, "single", occRaceWorkers)
	if winners != 1 || losers != occRaceWorkers-1 {
		t.Fatalf("expected exactly one winner and %d stale loser(s); got winners=%d losers=%d", occRaceWorkers-1, winners, losers)
	}
}

func runSingleOCCRace(t *testing.T, ctx context.Context, db *sql.DB, schema, suffix string, workers int) (int, int) {
	t.Helper()
	if workers < 2 {
		t.Fatalf("workers must be >=2, got %d", workers)
	}
	tenantID := testdb.DeterministicID(t, "tenant-"+suffix)
	docID := testdb.DeterministicID(t, "doc-"+suffix)
	userID := testdb.DeterministicID(t, "user-"+suffix)

	testdb.SeedUser(t, ctx, db, schema, userID, "Race User")
	testdb.SeedDocument(t, ctx, db, schema, docID, tenantID, userID)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`DELETE FROM %s WHERE id = $1::uuid`, testdb.Qualified(schema, "documents")), docID)
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`DELETE FROM %s WHERE tenant_id = $1::uuid`, testdb.Qualified(schema, "iam_users")), tenantID)
	})

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(workers)

	results := make([]int64, workers)
	errs := make([]error, workers)

	for i := 0; i < workers; i++ {
		i := i
		go func() {
			defer wg.Done()
			<-start

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				errs[i] = err
				return
			}
			defer tx.Rollback()

			testdb.SetCapsOnTx(t, tx, `[{"cap":"document.edit"}]`)

			res, err := tx.ExecContext(ctx, fmt.Sprintf(`
				UPDATE %s
				   SET revision_version = revision_version + 1
				 WHERE id = $1::uuid
				   AND tenant_id = $2::uuid
				   AND revision_version = $3`,
				testdb.Qualified(schema, "documents")),
				docID, tenantID, 1,
			)
			if err != nil {
				errs[i] = err
				return
			}
			ra, err := res.RowsAffected()
			if err != nil {
				errs[i] = err
				return
			}
			if err := tx.Commit(); err != nil {
				errs[i] = err
				return
			}
			results[i] = ra
		}()
	}

	close(start)
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			t.Fatalf("race update failed: %v", err)
		}
	}

	winners := 0
	losers := 0
	for _, ra := range results {
		if ra == 1 {
			winners++
		}
		if ra == 0 {
			losers++
		}
	}
	return winners, losers
}

func testSkipLockedNoDuplicateProcessing(t *testing.T) {
	db, schema := testdb.Open(t)
	ctx := context.Background()
	tenantID := testdb.DeterministicID(t, "tenant-skiplocked")
	actorID := testdb.DeterministicID(t, "actor-skiplocked")
	routeTemplate := "POST /integration/skiplocked"

	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`
			DELETE FROM %s
			 WHERE tenant_id = $1::uuid
			   AND actor_user_id = $2
			   AND route_template = $3`,
			testdb.Qualified(schema, "idempotency_keys")),
			tenantID, actorID, routeTemplate,
		)
	})

	seedIdempotencyTenant(t, db, tenantID)

	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("k-%d", i)
		_, err := db.ExecContext(ctx, fmt.Sprintf(`
			INSERT INTO %s
				(tenant_id, actor_user_id, route_template, key, payload_hash, response_status, response_body, status, expires_at)
			VALUES
				($1::uuid, $2, $3, $4, $5, 200, '\x7b7d'::bytea, 'completed', now() - interval '1 hour')`,
			testdb.Qualified(schema, "idempotency_keys")),
			tenantID, actorID, routeTemplate, key, fmt.Sprintf("h-%d", i),
		)
		if err != nil {
			t.Fatalf("seed idempotency key %d: %v", i, err)
		}
	}

	type rowRef struct {
		tenantID      string
		actorUserID   string
		routeTemplate string
		key           string
	}

	var mu sync.Mutex
	processed := map[string]int{}

	start := make(chan struct{})
	selected := sync.WaitGroup{}
	selected.Add(3)
	var wg sync.WaitGroup
	errs := make([]error, 3)
	wg.Add(3)
	for w := 0; w < 3; w++ {
		w := w
		go func() {
			defer wg.Done()
			<-start

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				errs[w] = fmt.Errorf("worker %d begin tx: %w", w, err)
				return
			}
			defer tx.Rollback()

			rows, err := tx.QueryContext(ctx, fmt.Sprintf(`
				SELECT tenant_id::text, actor_user_id, route_template, key
				  FROM %s
				 WHERE tenant_id = $1::uuid
				   AND actor_user_id = $2
				   AND route_template = $3
				   AND status = 'completed'
				 FOR UPDATE SKIP LOCKED
				 LIMIT 2`,
				testdb.Qualified(schema, "idempotency_keys")),
				tenantID, actorID, routeTemplate,
			)
			if err != nil {
				errs[w] = fmt.Errorf("worker %d select for update skip locked: %w", w, err)
				return
			}

			picked := make([]rowRef, 0, 2)
			for rows.Next() {
				var r rowRef
				if err := rows.Scan(&r.tenantID, &r.actorUserID, &r.routeTemplate, &r.key); err != nil {
					_ = rows.Close()
					errs[w] = fmt.Errorf("worker %d scan: %w", w, err)
					return
				}
				picked = append(picked, r)
			}
			if err := rows.Err(); err != nil {
				_ = rows.Close()
				errs[w] = fmt.Errorf("worker %d rows err: %w", w, err)
				return
			}
			_ = rows.Close()

			// Signal this goroutine has finished selecting, then wait for all.
			selected.Done()
			selected.Wait()

			for _, r := range picked {
				res, err := tx.ExecContext(ctx, fmt.Sprintf(`
					UPDATE %s
					   SET status = 'failed'
					 WHERE tenant_id = $1::uuid
					   AND actor_user_id = $2
					   AND route_template = $3
					   AND key = $4
					   AND status = 'completed'`,
					testdb.Qualified(schema, "idempotency_keys")),
					r.tenantID, r.actorUserID, r.routeTemplate, r.key,
				)
				if err != nil {
					errs[w] = fmt.Errorf("worker %d update key %s: %w", w, r.key, err)
					return
				}
				ra, err := res.RowsAffected()
				if err != nil {
					errs[w] = fmt.Errorf("worker %d rows affected key %s: %w", w, r.key, err)
					return
				}
				if ra != 1 {
					errs[w] = fmt.Errorf("worker %d expected rows affected=1 for key %s, got %d", w, r.key, ra)
					return
				}
				mu.Lock()
				processed[r.key]++
				mu.Unlock()
			}

			if err := tx.Commit(); err != nil {
				errs[w] = fmt.Errorf("worker %d commit: %w", w, err)
				return
			}
		}()
	}

	close(start)
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatalf("skip-locked worker failed: %v", err)
		}
	}

	if len(processed) != 5 {
		t.Fatalf("expected 5 processed keys, got %d (%v)", len(processed), processed)
	}
	for k, cnt := range processed {
		if cnt != 1 {
			t.Fatalf("key %s processed %d times (expected 1)", k, cnt)
		}
	}
}

func testSignoffUniqueDuplicateBlocked(t *testing.T) {
	db, schema := testdb.Open(t)
	ctx := context.Background()

	tenantID := testdb.DeterministicID(t, "tenant-signoff")
	authorID := testdb.DeterministicID(t, "author-signoff")
	actorID := testdb.DeterministicID(t, "actor-signoff")
	docID := testdb.DeterministicID(t, "doc-signoff")
	routeID := testdb.DeterministicID(t, "route-signoff")
	instanceID := testdb.DeterministicID(t, "instance-signoff")
	stageID := testdb.DeterministicID(t, "stage-signoff")

	// approval_instances_submitted_by_tenant_fkey / approval_signoffs_actor_tenant_fkey
	// both require (tenant_id, user_id) to match an iam_users row; plain
	// SeedUser leaves iam_users.tenant_id at its default (dev tenant), which
	// does not match this test's randomly-minted tenantID. SeedSystemAdmin
	// seeds iam_users with the correct tenant_id (mirrors this test's own
	// Cleanup, which already deletes iam_users by tenant_id=tenantID).
	testdb.SeedSystemAdmin(t, db, tenantID, authorID, "Doc Author")
	testdb.SeedSystemAdmin(t, db, tenantID, actorID, "Signer")
	testdb.SeedDocument(t, ctx, db, schema, docID, tenantID, authorID)
	testdb.SeedRouteConfig(t, ctx, db, schema, routeID, tenantID, "int_signoff")

	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`DELETE FROM %s WHERE approval_instance_id = $1::uuid`, testdb.Qualified(schema, "approval_signoffs")), instanceID)
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`DELETE FROM %s WHERE id = $1::uuid`, testdb.Qualified(schema, "approval_stage_instances")), stageID)
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`DELETE FROM %s WHERE id = $1::uuid`, testdb.Qualified(schema, "approval_instances")), instanceID)
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`DELETE FROM %s WHERE id = $1::uuid`, testdb.Qualified(schema, "approval_routes")), routeID)
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`DELETE FROM %s WHERE id = $1::uuid`, testdb.Qualified(schema, "documents")), docID)
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`DELETE FROM %s WHERE tenant_id = $1::uuid`, testdb.Qualified(schema, "iam_users")), tenantID)
	})

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin setup tx: %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT set_config('metaldocs.bypass_authz', 'scheduler', true)`); err != nil {
		t.Fatalf("set bypass_authz for setup: %v", err)
	}

	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s
			(id, tenant_id, document_id, route_id, route_version_snapshot, status, submitted_by, submitted_at, content_hash_at_submit, idempotency_key)
		VALUES
			($1::uuid, $2::uuid, $3::uuid, $4::uuid, 1, 'in_progress', $5, now(), 'hash-signoff', 'idem-signoff')`,
		testdb.Qualified(schema, "approval_instances")),
		instanceID, tenantID, docID, routeID, authorID,
	); err != nil {
		t.Fatalf("seed approval_instance: %v", err)
	}

	eligible, _ := json.Marshal([]string{actorID})
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s
			(id, approval_instance_id, stage_order, name_snapshot, required_role_snapshot, required_capability_snapshot, area_code_snapshot, quorum_snapshot, quorum_m_snapshot, on_eligibility_drift_snapshot, eligible_actor_ids, effective_denominator, status, opened_at)
		VALUES
			($1::uuid, $2::uuid, 1, 'Stage 1', 'reviewer', 'document.signoff', 'AREA_INT', 'any_1_of', NULL, 'keep_snapshot', $3::jsonb, 1, 'active', now())`,
		testdb.Qualified(schema, "approval_stage_instances")),
		stageID, instanceID, string(eligible),
	); err != nil {
		t.Fatalf("seed stage_instance: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit setup tx: %v", err)
	}

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	errs := make([]error, 2)
	for i := 0; i < 2; i++ {
		i := i
		go func() {
			defer wg.Done()
			<-start

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				errs[i] = err
				return
			}
			defer tx.Rollback()

			if _, err := tx.ExecContext(ctx, `SELECT set_config('metaldocs.bypass_authz', 'scheduler', true)`); err != nil {
				errs[i] = err
				return
			}

			_, err = tx.ExecContext(ctx, fmt.Sprintf(`
				INSERT INTO %s
					(id, approval_instance_id, stage_instance_id, actor_user_id, actor_tenant_id, actor_display_name_snapshot, decision, comment, signed_at, signature_method, signature_payload, content_hash)
				VALUES
					($1::uuid, $2::uuid, $3::uuid, $4, $5::uuid, 'Signer', 'approve', 'integration race', now(), 'password', '{}'::jsonb, 'content-hash')`,
				testdb.Qualified(schema, "approval_signoffs")),
				testdb.DeterministicID(t, fmt.Sprintf("signoff-%d", i)), instanceID, stageID, actorID, tenantID,
			)
			if err != nil {
				errs[i] = err
				return
			}

			if err := tx.Commit(); err != nil {
				errs[i] = err
				return
			}
		}()
	}

	close(start)
	wg.Wait()

	uniqueViolations := 0
	successes := 0
	for _, err := range errs {
		if err == nil {
			successes++
			continue
		}
		if hasSQLState(err, "23505") {
			uniqueViolations++
			continue
		}
		t.Fatalf("unexpected signoff insert error: %v", err)
	}
	if successes != 1 || uniqueViolations != 1 {
		t.Fatalf("expected one success and one 23505 violation, got successes=%d uniqueViolations=%d errs=%v", successes, uniqueViolations, errs)
	}

	var count int
	if err := db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT count(*)
		  FROM %s
		 WHERE approval_instance_id = $1::uuid
		   AND stage_instance_id = $2::uuid
		   AND actor_user_id = $3`,
		testdb.Qualified(schema, "approval_signoffs")),
		instanceID, stageID, actorID,
	).Scan(&count); err != nil {
		t.Fatalf("count signoffs: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 signoff row, got %d", count)
	}
}

func hasSQLState(err error, state string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == state
	}
	return strings.Contains(err.Error(), state)
}
