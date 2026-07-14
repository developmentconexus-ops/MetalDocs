//go:build integration
// +build integration

package scenarios_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"metaldocs/tests/integration/testdb"
)

func TestOutbox_ApprovalInstanceInsertHasGovernanceEvent(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)

	tenantID := testdb.DeterministicID(t, "tenant")
	authorID := testdb.DeterministicID(t, "author")
	docID := testdb.DeterministicID(t, "doc")
	routeID := testdb.DeterministicID(t, "route")
	instanceID := testdb.DeterministicID(t, "instance")

	// approval_instances_submitted_by_tenant_fkey requires (tenant_id,
	// submitted_by) to match an iam_users(tenant_id, user_id) row; plain
	// SeedUser leaves iam_users.tenant_id at its default (dev tenant), which
	// does not match this test's randomly-minted tenantID. SeedSystemAdmin
	// seeds iam_users with the correct tenant_id (mirrors this test's own
	// Cleanup, which already deletes iam_users by tenant_id=tenantID).
	testdb.SeedSystemAdmin(t, db, tenantID, authorID, "Outbox Author")
	testdb.SeedDocument(t, ctx, db, schema, docID, tenantID, authorID)
	testdb.SeedRouteConfig(t, ctx, db, schema, routeID, tenantID, "outbox_flow")

	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`DELETE FROM %s WHERE id = $1::uuid`, testdb.Qualified(schema, "approval_instances")), instanceID)
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`DELETE FROM %s WHERE tenant_id = $1::uuid AND event_type = 'doc.submitted' AND resource_id = $2`, testdb.Qualified(schema, "governance_events")), tenantID, docID)
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`DELETE FROM %s WHERE id = $1::uuid`, testdb.Qualified(schema, "approval_routes")), routeID)
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`DELETE FROM %s WHERE id = $1::uuid`, testdb.Qualified(schema, "documents")), docID)
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`DELETE FROM %s WHERE tenant_id = $1::uuid`, testdb.Qualified(schema, "iam_users")), tenantID)
	})

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `SELECT set_config('metaldocs.bypass_authz', 'scheduler', true)`); err != nil {
		_ = tx.Rollback()
		t.Fatalf("set bypass_authz: %v", err)
	}

	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s
			(id, tenant_id, document_id, route_id, route_version_snapshot, status, submitted_by, submitted_at, content_hash_at_submit, idempotency_key)
		VALUES
			($1::uuid, $2::uuid, $3::uuid, $4::uuid, 1, 'in_progress', $5, now(), 'outbox-hash', 'outbox-idem')`,
		testdb.Qualified(schema, "approval_instances")),
		instanceID, tenantID, docID, routeID, authorID,
	); err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert approval_instance: %v", err)
	}

	payload, _ := json.Marshal(map[string]any{
		"instance_id": instanceID,
		"document_id": docID,
	})
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s
			(tenant_id, event_type, actor_user_id, resource_type, resource_id, reason, payload_json)
		VALUES
			($1::uuid, 'doc.submitted', $2, 'document', $3, 'integration outbox pairing', $4::jsonb)`,
		testdb.Qualified(schema, "governance_events")),
		tenantID, authorID, docID, string(payload),
	); err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert governance_event: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx: %v", err)
	}

	var instanceTenant string
	if err := db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT tenant_id::text
		  FROM %s
		 WHERE id = $1::uuid`,
		testdb.Qualified(schema, "approval_instances")),
		instanceID,
	).Scan(&instanceTenant); err != nil {
		t.Fatalf("read approval_instance: %v", err)
	}

	var eventTenant string
	if err := db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT tenant_id::text
		  FROM %s
		 WHERE tenant_id = $1::uuid
		   AND event_type = 'doc.submitted'
		   AND resource_type = 'document'
		   AND resource_id = $2`,
		testdb.Qualified(schema, "governance_events")),
		tenantID, docID,
	).Scan(&eventTenant); err != nil {
		t.Fatalf("read governance_event: %v", err)
	}

	if instanceTenant != eventTenant {
		t.Fatalf("tenant mismatch: approval_instance=%s governance_event=%s", instanceTenant, eventTenant)
	}
}

func TestOutbox_RollbackOmitsEvent(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)
	table := testdb.Qualified(schema, "governance_events")

	tenantID := testdb.DeterministicID(t, "tenant-rollback")
	resourceID := testdb.DeterministicID(t, "resource-rollback")
	actorID := "outbox-rollback-user"

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s
			(tenant_id, event_type, actor_user_id, resource_type, resource_id, reason, payload_json)
		VALUES
			($1::uuid, 'doc.submitted', $2, 'document', $3, 'rollback-test', '{"rollback":true}'::jsonb)`, table),
		tenantID, actorID, resourceID,
	); err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert event in tx: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback tx: %v", err)
	}

	var count int
	if err := db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT count(*)
		  FROM %s
		 WHERE tenant_id = $1::uuid
		   AND event_type = 'doc.submitted'
		   AND resource_id = $2`, table),
		tenantID, resourceID,
	).Scan(&count); err != nil {
		t.Fatalf("count events after rollback: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected rollback to remove event row; found %d row(s)", count)
	}
}

func TestOutbox_DedupeKey(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)
	table := testdb.Qualified(schema, "governance_events")

	// governance_events lives in public (db/baseline/0001_current_schema.sql:2268),
	// not metaldocs; on the factory database dedupe_key is a real, always-present
	// column (not a feature flag), so this test asserts it directly instead of
	// probing information_schema and skipping when absent.
	tenantID := testdb.DeterministicID(t, "tenant-dedupe")
	actorID := "outbox-dedupe-user"
	resourceID := "outbox-dedupe-resource"
	dedupeKey := "test-dedup-key-1"
	eventType := "doc.submitted"

	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf(`
			DELETE FROM %s
			 WHERE tenant_id = $1::uuid
			   AND event_type = $2
			   AND dedupe_key = $3`, table),
			tenantID, eventType, dedupeKey,
		)
	})

	if _, err := db.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s
			(tenant_id, event_type, actor_user_id, resource_type, resource_id, reason, payload_json, dedupe_key)
		VALUES
			($1::uuid, $2, $3, 'document', $4, 'dedupe-1', '{}'::jsonb, $5)`, table),
		tenantID, eventType, actorID, resourceID, dedupeKey,
	); err != nil {
		t.Fatalf("seed dedupe event: %v", err)
	}

	if _, err := db.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s
			(tenant_id, event_type, actor_user_id, resource_type, resource_id, reason, payload_json, dedupe_key)
		VALUES
			($1::uuid, $2, $3, 'document', $4, 'dedupe-2', '{}'::jsonb, $5)
		ON CONFLICT DO NOTHING`, table),
		tenantID, eventType, actorID, resourceID, dedupeKey,
	); err != nil {
		t.Fatalf("insert duplicate dedupe key: %v", err)
	}

	var count int
	if err := db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT count(*)
		  FROM %s
		 WHERE tenant_id = $1::uuid
		   AND event_type = $2
		   AND dedupe_key = $3`, table),
		tenantID, eventType, dedupeKey,
	).Scan(&count); err != nil {
		t.Fatalf("count dedupe rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one deduped governance_event row, got %d", count)
	}
}
