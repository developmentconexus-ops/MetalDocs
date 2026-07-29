package application

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	v2dom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/render/resolvers"
	"metaldocs/internal/platform/db"
)

// RepairMaterialization is the operator-only re-freeze path (scripts/release-backfill).
//
// It is the ONLY sanctioned way a document frozen before migration 0313
// (values_hash written, frozen_revision_id NULL) can materialize again: it takes
// a FRESH pin from the current revision inside the repair transaction and only
// then enqueues, so the pin always names a revision that really is the body
// about to be rendered — true lineage, never fabricated history. What it must
// never do is recompute the values_hash the approver signed.

// repairTx returns a real *sql.Tx (sqlmock-backed) whose single expected query
// answers RepairMaterialization's freeze-state read. frozenAt may be nil to
// model a document that was never frozen.
func repairTx(t *testing.T, valuesHash []byte, frozenAt any) db.Tx {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = mockDB.Close() })
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT values_hash, values_frozen_at").
		WillReturnRows(sqlmock.NewRows([]string{"values_hash", "values_frozen_at"}).
			AddRow(valuesHash, frozenAt))
	tx, err := mockDB.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	return tx
}

// repairOrderOutbox records how many freeze writes had happened BEFORE the
// dispatch was enqueued, so the ordering invariant (pin first, enqueue second)
// is asserted rather than assumed: an enqueue that beat the pin would dispatch
// a render against the old — possibly NULL — lineage.
type repairOrderOutbox struct {
	finalize          *fakeFreezeFinalizer
	calls             int
	pinCallsAtEnqueue int
	hash              []byte
	generationIDs     []string
}

func (o *repairOrderOutbox) EnqueueMaterializeTx(_ context.Context, _ db.Tx, _, _ string, valuesHash []byte, releaseGenerationID string) error {
	o.calls++
	o.pinCallsAtEnqueue = o.finalize.calls
	o.hash = append([]byte(nil), valuesHash...)
	o.generationIDs = append(o.generationIDs, releaseGenerationID)
	return nil
}

func newRepairSvc(finalize *fakeFreezeFinalizer, snapReader fakeSnapshotReader, outbox *repairOrderOutbox) *FreezeService {
	return NewFreezeService(
		fakeSchemaReader{},
		&fakeFillInWriter{},
		&fakeValuesReader{},
		resolvers.NewRegistry(),
		finalize,
		&fakeResolverContextBuilder{},
		snapReader,
		&fakeFanoutClient{},
	).WithMaterializeOutbox(outbox)
}

func TestFreezeService_RepairMaterialization_PinsBeforeEnqueue(t *testing.T) {
	signedHash := []byte("signed-values-hash")
	frozenAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	finalize := &fakeFreezeFinalizer{}
	outbox := &repairOrderOutbox{finalize: finalize}
	head := v2dom.RevisionRef{ID: "22222222-2222-4222-8222-222222222222", StorageKey: "documents/d/revisions/head.docx"}
	// The document already carries a pin naming an older revision: choosing to
	// repair IS the decision to re-freeze, so that pin is superseded, not defended.
	stale := v2dom.RevisionRef{ID: "11111111-1111-4111-8111-111111111111"}
	svc := newRepairSvc(finalize, fakeSnapshotReader{currentRef: head, frozenRef: stale}, outbox)

	if err := svc.RepairMaterialization(context.Background(), repairTx(t, signedHash, frozenAt), "t", "doc-1", "gen-1"); err != nil {
		t.Fatalf("RepairMaterialization: %v", err)
	}

	if finalize.calls != 1 {
		t.Fatalf("WriteFreeze calls = %d, want 1", finalize.calls)
	}
	if finalize.pinnedRevisionID != head.ID {
		t.Fatalf("pinned revision = %q, want the current revision %q", finalize.pinnedRevisionID, head.ID)
	}
	// values_hash and values_frozen_at are the approver's: carried over verbatim.
	if string(finalize.hash) != string(signedHash) {
		t.Fatalf("repair rewrote values_hash: %q", finalize.hash)
	}
	if !finalize.at.Equal(frozenAt) {
		t.Fatalf("repair restamped values_frozen_at: %v, want %v", finalize.at, frozenAt)
	}
	if outbox.calls != 1 {
		t.Fatalf("enqueue calls = %d, want 1", outbox.calls)
	}
	if outbox.pinCallsAtEnqueue != 1 {
		t.Fatal("dispatch was enqueued before the lineage pin was written")
	}
	if string(outbox.hash) != string(signedHash) {
		t.Fatalf("enqueued hash = %q, want the signed values_hash", outbox.hash)
	}
	if len(outbox.generationIDs) != 1 || outbox.generationIDs[0] != "gen-1" {
		t.Fatalf("generation tag = %v, want [gen-1]", outbox.generationIDs)
	}
}

// A legacy NULL-pin document (frozen before 0313) is the §5.3 case: the absent
// pin is what repair FIXES, never a reason to refuse.
func TestFreezeService_RepairMaterialization_LegacyNullPinBecomesRepairable(t *testing.T) {
	finalize := &fakeFreezeFinalizer{}
	outbox := &repairOrderOutbox{finalize: finalize}
	head := v2dom.RevisionRef{ID: "33333333-3333-4333-8333-333333333333"}
	// frozenRef zero = no pin recorded: the pre-0313 shape.
	svc := newRepairSvc(finalize, fakeSnapshotReader{currentRef: head}, outbox)

	if err := svc.RepairMaterialization(context.Background(), repairTx(t, []byte("h"), time.Now().UTC()), "t", "doc-1", ""); err != nil {
		t.Fatalf("legacy NULL-pin document must be repairable, got: %v", err)
	}
	if finalize.pinnedRevisionID != head.ID {
		t.Fatalf("pinned revision = %q, want %q", finalize.pinnedRevisionID, head.ID)
	}
	if outbox.calls != 1 {
		t.Fatalf("enqueue calls = %d, want 1", outbox.calls)
	}
}

func TestFreezeService_RepairMaterialization_RefusesNeverFrozen(t *testing.T) {
	finalize := &fakeFreezeFinalizer{}
	outbox := &repairOrderOutbox{finalize: finalize}
	svc := newRepairSvc(finalize, fakeSnapshotReader{currentRef: v2dom.RevisionRef{ID: "rev-1"}}, outbox)

	err := svc.RepairMaterialization(context.Background(), repairTx(t, nil, nil), "t", "doc-1", "")
	if err == nil || !containsStr(err.Error(), "was never frozen") {
		t.Fatalf("expected never-frozen refusal, got %v", err)
	}
	if finalize.calls != 0 || outbox.calls != 0 {
		t.Fatalf("refused repair still wrote state: writes=%d enqueues=%d", finalize.calls, outbox.calls)
	}
}

// No current revision means there is no honest body to pin. Repair refuses
// rather than writing a NULL pin and letting Materialize fail later.
func TestFreezeService_RepairMaterialization_RefusesWithoutCurrentRevision(t *testing.T) {
	finalize := &fakeFreezeFinalizer{}
	outbox := &repairOrderOutbox{finalize: finalize}
	svc := newRepairSvc(finalize, fakeSnapshotReader{}, outbox)

	err := svc.RepairMaterialization(context.Background(), repairTx(t, []byte("h"), time.Now().UTC()), "t", "doc-1", "")
	if err == nil || !containsStr(err.Error(), "no current revision") {
		t.Fatalf("expected missing-revision refusal, got %v", err)
	}
	if finalize.calls != 0 || outbox.calls != 0 {
		t.Fatalf("refused repair still wrote state: writes=%d enqueues=%d", finalize.calls, outbox.calls)
	}
}
