//go:build integration
// +build integration

package application_test

// Etapa 5 §5.2 — real-DB proof of the freeze lineage pin end to end.
//
// The unit tests in freeze_pin_test.go prove the SERVICE logic against fakes.
// This file proves the part fakes cannot: that Pin's lineage pin survives a
// real UPDATE against public.documents (tripwire + FK + tenancy included), and
// that Materialize reads that pin back out of the database — not
// current_revision_id — and refuses to render when the pinned body's bytes do
// not hash to the pinned revision's recorded content_hash.
//
// Only the two genuinely-external edges are faked: the docx-renderer HTTP
// fanout and the blob store. Everything between the service and Postgres is
// real (SnapshotRepository, FillInRepository, SnapshotSchemaReader).

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	docapp "metaldocs/internal/modules/documents/application"
	docrepo "metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/resolvers"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// freezeLineageTenantID is the seeded dev tenant every documents integration
// fixture uses.
const freezeLineageTenantID = tenant.DevTenantID

// recordingFanout stands in for the docx-renderer HTTP call and records the
// body key it was asked to render — the assertion target for "did Materialize
// follow the pin?".
type recordingFanout struct {
	calls int
	req   fanout.FanoutRequest
	resp  fanout.FanoutResponse
}

func (f *recordingFanout) Fanout(_ context.Context, req fanout.FanoutRequest) (fanout.FanoutResponse, error) {
	f.calls++
	f.req = req
	return f.resp, nil
}

// mapBodyReader stands in for objectstore.VerifiedStore.
type mapBodyReader struct {
	bodies map[string][]byte
}

func (m mapBodyReader) AssertedReadObject(_ context.Context, _, key string, _ int64) ([]byte, error) {
	body, ok := m.bodies[key]
	if !ok {
		return nil, errors.New("no object at " + key)
	}
	return body, nil
}

// nopOutbox satisfies Pin's materialize dispatch enqueuer. The staging outbox
// itself is covered by the dispatchjobs integration suite; here it would only
// add unrelated schema surface to the assertion.
type nopOutbox struct{ calls int }

func (n *nopOutbox) EnqueueMaterializeTx(_ context.Context, _ db.Tx, _, _ string, _ []byte, _ string) error {
	n.calls++
	return nil
}

// staticResolverCtx is the minimal ResolverContextBuilder: the seeded template
// has an empty placeholder schema, so nothing is ever resolved from it.
type staticResolverCtx struct{}

func (staticResolverCtx) Build(_ context.Context, tenantID, documentID string, _ docapp.ApproverContext) (resolvers.ResolveInput, error) {
	return resolvers.ResolveInput{TenantID: resolvers.TenantID(tenantID), RevisionID: resolvers.RevisionID(documentID)}, nil
}

func (staticResolverCtx) BuildForDraft(_ context.Context, tenantID, documentID string) (resolvers.ResolveInput, error) {
	return resolvers.ResolveInput{TenantID: resolvers.TenantID(tenantID), RevisionID: resolvers.RevisionID(documentID)}, nil
}

// stageRevisionBody gives the fixture's revision a real storage key and the
// content hash of the given bytes, as VerifiedStore.Confirm would at upload
// time. public.document_revisions carries no tripwire (see
// testdb.InsertDraftDocument).
func stageRevisionBody(t *testing.T, db *sql.DB, docID string, body []byte) (revID, storageKey string) {
	t.Helper()
	ctx := context.Background()

	if err := db.QueryRowContext(ctx, `
		SELECT r.id::text
		  FROM public.documents d
		  JOIN public.document_revisions r ON r.id = d.current_revision_id
		 WHERE d.id = $1::uuid`, docID,
	).Scan(&revID); err != nil {
		t.Fatalf("load current revision: %v", err)
	}

	sum := sha256.Sum256(body)
	storageKey = "documents/" + docID + "/revisions/" + revID + ".docx"
	if _, err := db.ExecContext(ctx, `
		UPDATE public.document_revisions
		   SET storage_key = $2, content_hash = $3
		 WHERE id = $1::uuid`,
		revID, storageKey, hex.EncodeToString(sum[:]),
	); err != nil {
		t.Fatalf("stage revision body: %v", err)
	}
	return revID, storageKey
}

// newFreezeSvcForIntegration mirrors the worker composition root
// (apps/worker/cmd/metaldocs-worker/main.go) with real repos.
func newFreezeSvcForIntegration(sqlDB *sql.DB, fanoutClient docapp.FanoutClient) *docapp.FreezeService {
	snapRepo := docrepo.NewSnapshotRepository(sqlDB)
	fillInRepo := docrepo.NewFillInRepository(sqlDB)
	reg := resolvers.NewRegistry()
	resolvers.RegisterBuiltins(reg)
	return docapp.NewFreezeService(
		docapp.NewSnapshotSchemaReader(sqlDB),
		fillInRepo, fillInRepo,
		reg,
		snapRepo,
		staticResolverCtx{},
		snapRepo,
		fanoutClient,
	)
}

// TestFreezeLineage_Integration_PinThenMaterializeFollowsPin drives the real
// pipeline: Pin inside a real tx, then Materialize. It asserts the pin landed
// in public.documents.frozen_revision_id and that Materialize rendered THAT
// revision's key.
func TestFreezeLineage_Integration_PinThenMaterializeFollowsPin(t *testing.T) {
	ctx := context.Background()
	sqlDB, _ := testdb.Open(t)
	// >1 conn is required: SnapshotSchemaReader reads through the POOL while
	// SeedWithCaps holds the business tx on another connection.
	sqlDB.SetMaxOpenConns(4)

	docID, tnt := testdb.InsertDraftDocument(t, sqlDB, "", freezeLineageTenantID)
	body := []byte("PK\x03\x04 approved revision body")
	revID, storageKey := stageRevisionBody(t, sqlDB, docID, body)

	fanoutClient := &recordingFanout{resp: fanout.FanoutResponse{
		ContentHash:    "deadbeef00000000000000000000000000000000000000000000000000000000",
		FinalDocxS3Key: "final/" + docID + ".docx",
	}}
	svc := newFreezeSvcForIntegration(sqlDB, fanoutClient).
		WithMaterializeOutbox(&nopOutbox{}).
		WithRevisionBodyReader(mapBodyReader{bodies: map[string][]byte{storageKey: body}})

	// Pin updates public.documents — guarded by the document.edit tripwire.
	testdb.SeedWithCaps(t, sqlDB, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		return svc.Pin(ctx, tx, tnt, docID, docapp.ApproverContext{}, "")
	})

	var gotPin *string
	if err := sqlDB.QueryRowContext(ctx, `
		SELECT frozen_revision_id::text FROM public.documents
		 WHERE tenant_id = $1::uuid AND id = $2::uuid`, tnt, docID,
	).Scan(&gotPin); err != nil {
		t.Fatalf("read frozen_revision_id: %v", err)
	}
	if gotPin == nil || *gotPin != revID {
		t.Fatalf("frozen_revision_id = %v, want %s", gotPin, revID)
	}

	res, err := svc.Materialize(ctx, tnt, docID)
	if err != nil {
		t.Fatalf("Materialize: %v", err)
	}
	if fanoutClient.calls != 1 {
		t.Fatalf("fanout calls = %d, want 1", fanoutClient.calls)
	}
	if fanoutClient.req.BodyDocxS3Key != storageKey {
		t.Fatalf("rendered body key = %q, want the pinned revision key %q",
			fanoutClient.req.BodyDocxS3Key, storageKey)
	}
	if res.FinalDocxS3Key != "final/"+docID+".docx" {
		t.Fatalf("FinalDocxS3Key = %q", res.FinalDocxS3Key)
	}
}

// TestFreezeLineage_Integration_MaterializeFailsClosedOnHashMismatch proves the
// fail-closed half against a real database: the pin is recorded, but the bytes
// under the pinned key no longer hash to the recorded content_hash. Materialize
// must return ErrFrozenBodyHashMismatch, never call the renderer, and leave the
// document's artifact columns untouched.
func TestFreezeLineage_Integration_MaterializeFailsClosedOnHashMismatch(t *testing.T) {
	ctx := context.Background()
	sqlDB, _ := testdb.Open(t)
	// >1 conn is required: SnapshotSchemaReader reads through the POOL while
	// SeedWithCaps holds the business tx on another connection.
	sqlDB.SetMaxOpenConns(4)

	docID, tnt := testdb.InsertDraftDocument(t, sqlDB, "", freezeLineageTenantID)
	approved := []byte("the bytes the approver signed")
	_, storageKey := stageRevisionBody(t, sqlDB, docID, approved)

	// The object under the pinned key was substituted after the freeze.
	tampered := []byte("substituted bytes nobody approved")

	fanoutClient := &recordingFanout{}
	svc := newFreezeSvcForIntegration(sqlDB, fanoutClient).
		WithMaterializeOutbox(&nopOutbox{}).
		WithRevisionBodyReader(mapBodyReader{bodies: map[string][]byte{storageKey: tampered}})

	testdb.SeedWithCaps(t, sqlDB, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		return svc.Pin(ctx, tx, tnt, docID, docapp.ApproverContext{}, "")
	})

	_, err := svc.Materialize(ctx, tnt, docID)
	if !errors.Is(err, docapp.ErrFrozenBodyHashMismatch) {
		t.Fatalf("Materialize error = %v, want ErrFrozenBodyHashMismatch", err)
	}
	if fanoutClient.calls != 0 {
		t.Fatalf("renderer was called %d time(s) on a mismatch, want 0", fanoutClient.calls)
	}

	// No artifact may exist: Materialize returns before the fanout, so the
	// caller (MaterializeJobRunner) never reaches WriteFinalDocx.
	var finalKey string
	var contentHash []byte
	if err := sqlDB.QueryRowContext(ctx, `
		SELECT coalesce(final_docx_s3_key, ''), content_hash FROM public.documents
		 WHERE tenant_id = $1::uuid AND id = $2::uuid`, tnt, docID,
	).Scan(&finalKey, &contentHash); err != nil {
		t.Fatalf("read artifact columns: %v", err)
	}
	if finalKey != "" || len(contentHash) != 0 {
		t.Fatalf("artifact written despite fail-closed: key=%q hash=%x", finalKey, contentHash)
	}
}

// TestFreezeLineage_Integration_MaterializeRefusesUnpinnedDocument covers the
// pre-0313 shape: a document frozen without a pin (NULL frozen_revision_id)
// must fail closed rather than silently falling back to current_revision_id.
func TestFreezeLineage_Integration_MaterializeRefusesUnpinnedDocument(t *testing.T) {
	ctx := context.Background()
	sqlDB, _ := testdb.Open(t)
	// >1 conn is required: SnapshotSchemaReader reads through the POOL while
	// SeedWithCaps holds the business tx on another connection.
	sqlDB.SetMaxOpenConns(4)

	docID, tnt := testdb.InsertDraftDocument(t, sqlDB, "", freezeLineageTenantID)
	body := []byte("head revision body")
	_, storageKey := stageRevisionBody(t, sqlDB, docID, body)

	fanoutClient := &recordingFanout{}
	svc := newFreezeSvcForIntegration(sqlDB, fanoutClient).
		WithMaterializeOutbox(&nopOutbox{}).
		WithRevisionBodyReader(mapBodyReader{bodies: map[string][]byte{storageKey: body}})

	// Freeze WITHOUT a pin — exactly what a pre-migration freeze left behind.
	testdb.SeedWithCaps(t, sqlDB, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		return docrepo.NewSnapshotRepository(sqlDB).WriteFreeze(
			ctx, tnt, docID, bytes.Repeat([]byte{0xAA}, 32), "", time.Now().UTC(), tx)
	})

	_, err := svc.Materialize(ctx, tnt, docID)
	if err == nil {
		t.Fatal("Materialize on an unpinned document: got nil error, want fail-closed")
	}
	if fanoutClient.calls != 0 {
		t.Fatalf("renderer was called %d time(s) without a pin, want 0", fanoutClient.calls)
	}
}
