//go:build integration
// +build integration

// Package approval_test contains integration tests for approval signoff eligibility enforcement (J1).
//
// These tests target a testdb-leased factory database (public schema) and verify:
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
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"metaldocs/tests/integration/testdb"
)

// hasSQLState returns true when err carries the given PostgreSQL SQLSTATE code.
func hasSQLState(err error, state string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == state
	}
	return strings.Contains(err.Error(), state)
}

// TestSignoff_Eligibility runs all three spec assertions (a/b/c) for J1.
// Uses testdb.Open — a leased database built from the migrated factory
// template (ADR 0034), reset on t.Cleanup — not the shared dev database.
func TestSignoff_Eligibility(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

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
	// Seeds the FK chain via the canonical testdb factory (tenant, author,
	// eligible/non-eligible actors, document, route, in-progress instance) —
	// NewDocument free-mints template_version_id (no FK to either template
	// family, so no templates_template row is needed at all, mirroring
	// fixtures.go's SeedDocument precedent). approval_stage_instances has no
	// factory builder yet (mirrors review_verdict_integration_test.go's
	// seedReviewVerdictFixture), so it stays a raw INSERT with only the
	// eligible actor in eligible_actor_ids — the non-eligible actor is
	// deliberately excluded so the trigger has something to reject.
	t.Run("trigger_blocks_non_eligible_allows_eligible", func(t *testing.T) {
		database, _ := testdb.Open(t)

		tenant := testdb.NewTenant(t, database)
		author := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Trigger Author"))
		eligibleActor := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Eligible Signoff Actor"))
		nonEligibleActor := testdb.NewUser(t, database, testdb.WithTenant(tenant.ID), testdb.WithDisplayName("Non-Eligible Signoff Actor"))

		doc := testdb.NewDocument(t, database, testdb.WithTenant(tenant.ID), testdb.WithOwner(author.ID))
		route := testdb.NewApprovalRoute(t, database, testdb.WithTenant(tenant.ID))
		instance := testdb.NewApprovalInstance(t, database,
			testdb.WithDocument(doc),
			testdb.WithRoute(route),
			testdb.WithOwner(author.ID),
			testdb.WithStatus("in_progress"),
		)

		eligibleJSON, err := json.Marshal([]string{eligibleActor.ID})
		if err != nil {
			t.Fatalf("marshal eligible_actor_ids: %v", err)
		}

		var stageID string
		if err := database.QueryRowContext(ctx, `
			INSERT INTO public.approval_stage_instances
			  (approval_instance_id, stage_order, name_snapshot,
			   required_role_snapshot, required_capability_snapshot, area_code_snapshot,
			   quorum_snapshot, on_eligibility_drift_snapshot,
			   eligible_actor_ids, status)
			VALUES ($1::uuid, 1, 'QA Review',
			        'reviewer', 'document.signoff', 'QA',
			        'any_1_of', 'keep_snapshot',
			        $2::jsonb, 'active')
			RETURNING id::text`,
			instance.ID, string(eligibleJSON),
		).Scan(&stageID); err != nil {
			t.Fatalf("seed approval_stage_instances: %v", err)
		}

		contentHash := strings.Repeat("a", 64)

		// Non-eligible actor: trigger must block. approval_signoffs carries
		// trg_require_cap_asserted_signoffs (document.signoff), so the caps
		// assertion runs tx-local (pool-safe) even for the attempt expected to
		// fail — a single autocommit statement inside its own short tx means a
		// failed INSERT never poisons a later statement (unlike a shared,
		// explicit multi-statement transaction, which Postgres aborts on the
		// first error).
		badSignoffID := testdb.DeterministicID(t, "signoff-bad-c")
		badErr := func() error {
			tx, txErr := database.BeginTx(ctx, nil)
			if txErr != nil {
				t.Fatalf("begin tx (non-eligible signoff): %v", txErr)
			}
			defer tx.Rollback()
			testdb.SetCapsOnTx(t, tx, `[{"cap":"document.signoff"}]`)
			_, execErr := tx.ExecContext(ctx, `
				INSERT INTO public.approval_signoffs
				  (id, approval_instance_id, stage_instance_id, actor_user_id, actor_tenant_id,
				   actor_display_name_snapshot, decision, signed_at, signature_method,
				   signature_payload, content_hash)
				VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5::uuid,
				        $6, 'approve', now(), 'password', '{}', $7)`,
				badSignoffID, instance.ID, stageID, nonEligibleActor.ID, tenant.ID,
				nonEligibleActor.DisplayName, contentHash,
			)
			return execErr
		}()
		if badErr == nil {
			t.Fatal("expected trigger to block non-eligible actor; got nil error")
		}
		if !hasSQLState(badErr, "23514") && !strings.Contains(badErr.Error(), "not in eligible_actor_ids") {
			t.Fatalf("expected ERRCODE 23514 (check_violation); got: %v", badErr)
		}
		t.Logf("PASS (c/non-eligible): trigger blocked non-eligible actor → 403-equivalent at DB layer")

		// Eligible actor: must succeed — the non-vacuity control. Without this
		// half, a trigger that rejected every signoff (not just non-eligible
		// ones) would also pass the block above.
		goodSignoffID := testdb.DeterministicID(t, "signoff-good-c")
		tx, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx (eligible signoff): %v", err)
		}
		defer tx.Rollback()
		testdb.SetCapsOnTx(t, tx, `[{"cap":"document.signoff"}]`)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.approval_signoffs
			  (id, approval_instance_id, stage_instance_id, actor_user_id, actor_tenant_id,
			   actor_display_name_snapshot, decision, signed_at, signature_method,
			   signature_payload, content_hash)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5::uuid,
			        $6, 'approve', now(), 'password', '{}', $7)`,
			goodSignoffID, instance.ID, stageID, eligibleActor.ID, tenant.ID,
			eligibleActor.DisplayName, contentHash,
		); err != nil {
			t.Fatalf("expected eligible actor to succeed → 200-equivalent; got: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit eligible signoff: %v", err)
		}
		t.Log("PASS (c/eligible): eligible actor signoff inserted → 200-equivalent")
	})
}
