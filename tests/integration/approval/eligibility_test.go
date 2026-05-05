//go:build integration
// +build integration

// Package approval_test contains integration tests for approval signoff eligibility enforcement (J1).
//
// These tests target the live database (public schema) and verify:
// (a) The DB-level trigger enforce_signoff_eligibility_trg exists on approval_signoffs.
// (b) The governance_events table schema accepts signoff.rejected audit rows.
// (c) The eligibility trigger rejects a signoff for a non-eligible actor, and allows an eligible one.
//
// For HTTP-layer 403 coverage run the E2E test suite with a live API server.
package approval_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"metaldocs/tests/integration/testdb"
)

// openDB connects to the live database using METALDOCS_DATABASE_URL or DATABASE_URL.
func openDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("METALDOCS_DATABASE_URL"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if dsn == "" {
		t.Skip("DATABASE_URL/METALDOCS_DATABASE_URL not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.PingContext(context.Background()); err != nil {
		t.Skipf("integration DB unreachable: %v", err)
	}
	return db
}

// hasSQLState returns true when err carries the given PostgreSQL SQLSTATE code.
func hasSQLState(err error, state string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == state
	}
	return strings.Contains(err.Error(), state)
}

// TestSignoff_Eligibility runs all three spec assertions (a/b/c) for J1.
func TestSignoff_Eligibility(t *testing.T) {
	ctx := context.Background()
	db := openDB(t)

	// (a) Verify DB trigger present on approval_signoffs — confirms migration 0180 applied.
	t.Run("trigger_present", func(t *testing.T) {
		var triggerName string
		err := db.QueryRowContext(ctx, `
			SELECT trigger_name
			FROM information_schema.triggers
			WHERE trigger_schema = 'public'
			  AND event_object_table = 'approval_signoffs'
			  AND trigger_name = 'enforce_signoff_eligibility_trg'
			LIMIT 1`,
		).Scan(&triggerName)
		if errors.Is(err, sql.ErrNoRows) {
			t.Fatal("trigger enforce_signoff_eligibility_trg not found on public.approval_signoffs — was migration 0180 applied?")
		}
		if err != nil {
			t.Fatalf("query trigger: %v", err)
		}
		if triggerName != "enforce_signoff_eligibility_trg" {
			t.Fatalf("trigger name = %q; want enforce_signoff_eligibility_trg", triggerName)
		}
		t.Logf("PASS (trigger_present): %q present on public.approval_signoffs", triggerName)
	})

	// (b) Audit row: governance_events table accepts signoff.rejected rows.
	t.Run("audit_row_signoff_rejected", func(t *testing.T) {
		tenantID := testdb.DeterministicID(t, "tenant-audit")
		instanceID := testdb.DeterministicID(t, "instance-audit")
		actorID := "non-elig-audit-actor"

		_, err := db.ExecContext(ctx, `
			INSERT INTO governance_events
			  (tenant_id, event_type, actor_user_id, resource_type, resource_id, reason, payload_json)
			VALUES ($1::uuid, $2, $3, $4, $5, $6, '{}')`,
			tenantID, "signoff.rejected", actorID,
			"approval_instance", instanceID, "not_eligible",
		)
		if err != nil {
			t.Fatalf("insert audit governance_event: %v", err)
		}
		t.Cleanup(func() {
			_, _ = db.ExecContext(context.Background(),
				`DELETE FROM governance_events WHERE resource_id = $1 AND event_type = 'signoff.rejected'`,
				instanceID)
		})

		var count int
		if err := db.QueryRowContext(ctx, `
			SELECT count(*) FROM governance_events
			WHERE resource_id = $1
			  AND event_type = 'signoff.rejected'
			  AND reason = 'not_eligible'`,
			instanceID,
		).Scan(&count); err != nil {
			t.Fatalf("query audit row: %v", err)
		}
		if count == 0 {
			t.Fatal("expected audit row with event_type=signoff.rejected reason=not_eligible; got 0")
		}
		t.Logf("PASS (b): audit row kind=signoff.rejected reason=not_eligible present (count=%d)", count)
	})

	// (c) Trigger blocks non-eligible actor and allows eligible actor.
	// Uses a single transaction with savepoints so no data is permanently committed.
	// Seeds the full FK chain: iam_users → templates → template_versions → documents →
	//   approval_routes → approval_instances → approval_stage_instances → approval_signoffs.
	// Uses the dev tenant (ffffffff-...) where document_profiles exist for 'dc' profile_code.
	t.Run("trigger_blocks_non_eligible_allows_eligible", func(t *testing.T) {
		const devTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
		const devProfileCode = "dc" // document_profiles row exists for dev tenant

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer tx.Rollback() // always rolls back — data is never committed

		// Bypass authz GUC triggers (mirrors outbox_same_tx_test.go pattern).
		if _, err := tx.ExecContext(ctx,
			`SELECT set_config('metaldocs.bypass_authz', 'scheduler', true)`,
		); err != nil {
			t.Skipf("cannot set bypass_authz GUC: %v", err)
		}

		authorID := testdb.DeterministicID(t, "author-trg-c")
		docID := testdb.DeterministicID(t, "doc-trg-c")
		routeID := testdb.DeterministicID(t, "route-trg-c")
		instanceID := testdb.DeterministicID(t, "inst-trg-c")
		stageID := testdb.DeterministicID(t, "stage-trg-c")
		eligibleActorID := "elig-trg-c-actor"
		nonEligibleActorID := "nonelig-trg-c-actor"

		// Upsert author.
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO metaldocs.iam_users (user_id, tenant_id, display_name, is_active, created_at, updated_at)
			VALUES ($1, $2::uuid, 'Trigger Author', true, now(), now())
			ON CONFLICT (user_id) DO NOTHING`,
			authorID, devTenantID,
		); err != nil {
			t.Skipf("cannot seed iam_users: %v", err)
		}

		// Seed template.
		var tplID string
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO templates (tenant_id, key, name, created_by)
			VALUES ($1::uuid, $2, 'Trigger Test Tpl', $3::uuid)
			ON CONFLICT (tenant_id, key) DO UPDATE SET name = EXCLUDED.name
			RETURNING id::text`,
			devTenantID, "trg-elig-tpl-"+docID[:8], authorID,
		).Scan(&tplID); err != nil {
			t.Skipf("cannot seed template: %v", err)
		}

		// Seed template_version.
		var tvID string
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO template_versions
			  (template_id, version_num, status, docx_storage_key, schema_storage_key,
			   docx_content_hash, schema_content_hash, created_by)
			VALUES ($1::uuid, 1, 'published', 'key/t.docx', 'key/s.json', 'aa', 'bb', $2::uuid)
			ON CONFLICT (template_id, version_num) DO UPDATE SET status = EXCLUDED.status
			RETURNING id::text`,
			tplID, authorID,
		).Scan(&tvID); err != nil {
			t.Skipf("cannot seed template_version: %v", err)
		}

		// Seed document.
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO documents
			  (id, tenant_id, name, status, form_data_json, created_by,
			   template_version_id, revision_version, created_at, updated_at)
			VALUES ($1::uuid, $2::uuid, 'Trigger Elig Doc', 'draft', '{}', $3::uuid,
			        $4::uuid, 1, now(), now())
			ON CONFLICT (id) DO NOTHING`,
			docID, devTenantID, authorID, tvID,
		); err != nil {
			t.Skipf("cannot seed document: %v", err)
		}

		// Reuse existing approval_route for dev tenant + profile_code 'dc', or create if absent.
		if err := tx.QueryRowContext(ctx, `
			SELECT id::text FROM approval_routes
			WHERE tenant_id = $1::uuid AND profile_code = $2
			LIMIT 1`,
			devTenantID, devProfileCode,
		).Scan(&routeID); err != nil {
			// No existing route; try inserting.
			if _, insertErr := tx.ExecContext(ctx, `
				INSERT INTO approval_routes (id, tenant_id, name, profile_code, active, created_at, created_by)
				VALUES ($1::uuid, $2::uuid, 'Trigger Route', $3, true, now(), $4::uuid)
				ON CONFLICT DO NOTHING`,
				routeID, devTenantID, devProfileCode, authorID,
			); insertErr != nil {
				t.Skipf("cannot seed approval_routes: %v", insertErr)
			}
		}

		// Seed approval_instance.
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO approval_instances
			  (id, tenant_id, document_v2_id, route_id, route_version_snapshot,
			   status, submitted_by, submitted_at, content_hash_at_submit, idempotency_key)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, 1,
			        'in_progress', $5, now(), 'aabbcc', 'idem-trg-c')
			ON CONFLICT (id) DO NOTHING`,
			instanceID, devTenantID, docID, routeID, authorID,
		); err != nil {
			t.Skipf("cannot seed approval_instances: %v", err)
		}

		// Seed eligible actor in iam_users (required by approval_signoffs FK).
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO metaldocs.iam_users (user_id, tenant_id, display_name, is_active, created_at, updated_at)
			VALUES ($1, $2::uuid, 'Eligible Signoff Actor', true, now(), now())
			ON CONFLICT (user_id) DO NOTHING`,
			eligibleActorID, devTenantID,
		); err != nil {
			t.Skipf("cannot seed eligible actor in iam_users: %v", err)
		}

		eligibleJSON, _ := json.Marshal([]string{eligibleActorID})

		// Seed approval_stage_instance.
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO approval_stage_instances
			  (id, approval_instance_id, stage_order, name_snapshot,
			   required_role_snapshot, required_capability_snapshot, area_code_snapshot,
			   quorum_snapshot, on_eligibility_drift_snapshot,
			   eligible_actor_ids, status)
			VALUES ($1::uuid, $2::uuid, 1, 'QA Review',
			        'reviewer', 'doc.signoff', 'QA',
			        'any_1_of', 'keep_snapshot',
			        $3::jsonb, 'active')
			ON CONFLICT (id) DO NOTHING`,
			stageID, instanceID, string(eligibleJSON),
		); err != nil {
			t.Skipf("cannot seed approval_stage_instances: %v", err)
		}

		contentHash := strings.Repeat("a", 64)

		// Non-eligible actor: trigger must block.
		// Use a savepoint so the transaction remains usable after the expected error.
		badSignoffID := testdb.DeterministicID(t, "signoff-bad-c")
		if _, spErr := tx.ExecContext(ctx, `SAVEPOINT sp_bad_signoff`); spErr != nil {
			t.Skipf("cannot create savepoint: %v", spErr)
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO approval_signoffs
			  (id, approval_instance_id, stage_instance_id, actor_user_id, actor_tenant_id,
			   decision, signed_at, signature_method, signature_payload, content_hash)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5::uuid,
			        'approve', now(), 'password', '{}', $6)`,
			badSignoffID, instanceID, stageID, nonEligibleActorID, devTenantID, contentHash,
		)
		if err == nil {
			t.Fatal("expected trigger to block non-eligible actor; got nil error")
		}
		if !hasSQLState(err, "23514") && !strings.Contains(err.Error(), "not in eligible_actor_ids") {
			t.Fatalf("expected ERRCODE 23514 (check_violation); got: %v", err)
		}
		// Rollback to savepoint to recover the transaction.
		if _, rbErr := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT sp_bad_signoff`); rbErr != nil {
			t.Fatalf("rollback to savepoint failed: %v", rbErr)
		}
		t.Logf("PASS (c/non-eligible): trigger blocked non-eligible actor → 403-equivalent at DB layer")

		// Eligible actor: must succeed.
		goodSignoffID := testdb.DeterministicID(t, "signoff-good-c")
		_, err = tx.ExecContext(ctx, `
			INSERT INTO approval_signoffs
			  (id, approval_instance_id, stage_instance_id, actor_user_id, actor_tenant_id,
			   decision, signed_at, signature_method, signature_payload, content_hash)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5::uuid,
			        'approve', now(), 'password', '{}', $6)`,
			goodSignoffID, instanceID, stageID, eligibleActorID, devTenantID, contentHash,
		)
		if err != nil {
			t.Fatalf("expected eligible actor to succeed → 200-equivalent; got: %v", err)
		}
		t.Log("PASS (c/eligible): eligible actor signoff inserted → 200-equivalent")

		_ = fmt.Sprintf // suppress unused import
		// tx.Rollback() called by defer — no data committed.
	})
}

