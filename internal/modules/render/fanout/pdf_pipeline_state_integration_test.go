//go:build integration
// +build integration

package fanout

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"metaldocs/internal/platform/messaging"
	"metaldocs/tests/integration/testdb"
)

// openPipelineStateDB opens a template-cloned DB pinned to a single
// connection, mirroring dispatchjobs/dispatch_integration_test.go's
// openDispatchDB so the runtime search path resolves the metaldocs-schema
// outbox_events table.
func openPipelineStateDB(t *testing.T) *sql.DB {
	t.Helper()
	db, _ := testdb.OpenFreshDatabase(t)
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(context.Background(), `SET search_path TO metaldocs, public`); err != nil {
		t.Fatalf("set search_path: %v", err)
	}
	return db
}

// insertOutboxEvent inserts a row directly into metaldocs.outbox_events —
// outbox_events carries no tripwire (see postgres.Publisher.Publish, which
// inserts without asserting caps), so a plain INSERT is the real write path.
// deadLettered controls whether dead_lettered_at is set.
func insertOutboxEvent(t *testing.T, ctx context.Context, db *sql.DB, eventID, eventType, aggregateID, tenantID string, deadLettered bool) {
	t.Helper()
	var deadLetteredAt sql.NullTime
	if deadLettered {
		deadLetteredAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	}
	payload := `{"tenant_id":"` + tenantID + `","revision_id":"` + aggregateID + `"}`
	if _, err := db.ExecContext(ctx, `
INSERT INTO metaldocs.outbox_events (
  event_id, event_type, aggregate_type, aggregate_id, occurred_at, version,
  idempotency_key, producer, trace_id, payload, published_at, dead_lettered_at
)
VALUES ($1, $2, 'document_revision', $3, now(), 1, $1, 'test', 'trace-'||$1, $4::jsonb, NULL, $5)`,
		eventID, eventType, aggregateID, payload, deadLetteredAt,
	); err != nil {
		t.Fatalf("insert outbox_events row %s: %v", eventID, err)
	}
}

// TestPDFPipelineState_DeadLetteredMaterialize proves ReadState surfaces
// "failed" when the docx_materialize hop is dead-lettered.
func TestPDFPipelineState_DeadLetteredMaterialize(t *testing.T) {
	ctx := context.Background()
	db := openPipelineStateDB(t)

	tenant := testdb.NewTenant(t, db)
	revisionID := testdb.DeterministicID(t, "pipeline-materialize")

	insertOutboxEvent(t, ctx, db, testdb.DeterministicID(t, "evt-materialize"),
		string(messaging.EventTypeMaterializeFanout), revisionID, tenant.ID, true)

	reader := NewPDFPipelineStateReader(db)
	state, err := reader.ReadState(ctx, tenant.ID, revisionID)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}
	if state != "failed" {
		t.Fatalf("state = %q, want %q", state, "failed")
	}
}

// TestPDFPipelineState_DeadLetteredPDFConvert proves ReadState surfaces
// "failed" when the docgen_v2_pdf hop is dead-lettered.
func TestPDFPipelineState_DeadLetteredPDFConvert(t *testing.T) {
	ctx := context.Background()
	db := openPipelineStateDB(t)

	tenant := testdb.NewTenant(t, db)
	revisionID := testdb.DeterministicID(t, "pipeline-pdfconvert")

	insertOutboxEvent(t, ctx, db, testdb.DeterministicID(t, "evt-pdfconvert"),
		string(messaging.EventTypePDFConvert), revisionID, tenant.ID, true)

	reader := NewPDFPipelineStateReader(db)
	state, err := reader.ReadState(ctx, tenant.ID, revisionID)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}
	if state != "failed" {
		t.Fatalf("state = %q, want %q", state, "failed")
	}
}

// TestPDFPipelineState_LiveRowOnly proves a non-dead-lettered row (still in
// flight) does not surface "failed".
func TestPDFPipelineState_LiveRowOnly(t *testing.T) {
	ctx := context.Background()
	db := openPipelineStateDB(t)

	tenant := testdb.NewTenant(t, db)
	revisionID := testdb.DeterministicID(t, "pipeline-live")

	insertOutboxEvent(t, ctx, db, testdb.DeterministicID(t, "evt-live"),
		string(messaging.EventTypeMaterializeFanout), revisionID, tenant.ID, false)

	reader := NewPDFPipelineStateReader(db)
	state, err := reader.ReadState(ctx, tenant.ID, revisionID)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}
	if state != "" {
		t.Fatalf("state = %q, want empty", state)
	}
}

// TestPDFPipelineState_NoRows proves an aggregate with no outbox rows at all
// reads as pending ("").
func TestPDFPipelineState_NoRows(t *testing.T) {
	ctx := context.Background()
	db := openPipelineStateDB(t)

	tenant := testdb.NewTenant(t, db)
	revisionID := testdb.DeterministicID(t, "pipeline-norows")

	reader := NewPDFPipelineStateReader(db)
	state, err := reader.ReadState(ctx, tenant.ID, revisionID)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}
	if state != "" {
		t.Fatalf("state = %q, want empty", state)
	}
}

// TestPDFPipelineState_CrossTenantDeadLetterIgnored proves the tenant
// predicate is enforced: a dead-lettered row for the SAME aggregate_id but a
// DIFFERENT tenant must not leak "failed" across tenants.
func TestPDFPipelineState_CrossTenantDeadLetterIgnored(t *testing.T) {
	ctx := context.Background()
	db := openPipelineStateDB(t)

	tenant := testdb.NewTenant(t, db)
	otherTenant := testdb.NewTenant(t, db)
	revisionID := testdb.DeterministicID(t, "pipeline-crosstenant")

	insertOutboxEvent(t, ctx, db, testdb.DeterministicID(t, "evt-crosstenant"),
		string(messaging.EventTypeMaterializeFanout), revisionID, otherTenant.ID, true)

	reader := NewPDFPipelineStateReader(db)
	state, err := reader.ReadState(ctx, tenant.ID, revisionID)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}
	if state != "" {
		t.Fatalf("state = %q, want empty (cross-tenant dead-letter must not leak)", state)
	}
}
