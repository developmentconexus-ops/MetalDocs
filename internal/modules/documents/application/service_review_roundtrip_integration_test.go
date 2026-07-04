//go:build integration
// +build integration

package application_test

// TestReviewWriteRoundTrip exercises the reviewer-write path end-to-end against
// a real database (ADR 0034 testdb harness):
//
//  1. Seed: document under_review, eligible approver, open approval_stage_instance.
//  2. Approver: PresignAutosave + CommitAutosave succeed.
//     → new document_revisions row created (current_revision_id advances).
//     → documents.revision_number UNCHANGED (no new governed REV).
//  3. Non-eligible actor: PresignAutosave returns ErrInvalidStateTransition.

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/application"
	docrepo "metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/modules/documents/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/objectstore"
	"metaldocs/tests/integration/testdb"
)

func TestReviewWriteRoundTrip(t *testing.T) {
	ctx := context.Background()
	db, dbName := testdb.Open(t)

	// Set search_path and widen the pool so off-tx reads (H-PRE-1) don't deadlock.
	if _, err := db.ExecContext(ctx, `ALTER DATABASE "`+dbName+`" SET search_path TO public, metaldocs`); err != nil {
		t.Fatalf("alter database search_path: %v", err)
	}
	db.SetMaxIdleConns(0)
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(4)

	// Seed tenant, approver user, stranger user.
	tnt := testdb.NewTenant(t, db)
	approver := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	stranger := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	// Seed document under_review (auto-wires CD + snapshot columns).
	doc := testdb.NewDocument(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithOwner(approver.ID),
		testdb.WithStatus("under_review"),
	)

	// Look up the profile code for the document's controlled document so we can
	// seed an approval route that satisfies the document_profiles FK.
	var profileCode string
	if err := db.QueryRowContext(ctx,
		`SELECT profile_code FROM public.controlled_documents WHERE id=$1::uuid`,
		doc.ControlledDocumentID,
	).Scan(&profileCode); err != nil {
		t.Fatalf("lookup profile_code for CD %s: %v", doc.ControlledDocumentID, err)
	}

	// Seed approval route + in_progress instance + pending stage instance.
	route := testdb.NewApprovalRoute(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithProfile(profileCode),
	)
	inst := testdb.NewApprovalInstance(t, db,
		testdb.WithDocument(doc),
		testdb.WithRoute(route),
		testdb.WithStatus("in_progress"),
	)
	stageID := uuid.NewString()
	eligibleJSON := `["` + approver.ID + `"]`
	if _, err := db.ExecContext(ctx,
		`INSERT INTO public.approval_stage_instances
		   (id, approval_instance_id, stage_order, name_snapshot,
		    required_role_snapshot, required_capability_snapshot, area_code_snapshot,
		    quorum_snapshot, on_eligibility_drift_snapshot, eligible_actor_ids, status)
		 VALUES ($1::uuid, $2::uuid, 1, 'Stage 1',
		         'approver', 'document.signoff', 'area-001',
		         'any_1_of', 'keep_snapshot', $3::jsonb, 'pending')`,
		stageID, inst.ID, eligibleJSON,
	); err != nil {
		t.Fatalf("seed approval_stage_instances: %v", err)
	}

	// Assert autosave write capabilities at session level so CommitUpload's
	// authz.Require(document.edit) is satisfied (isolated DB, pool-safe).
	testdb.SetCapsOnDB(t, db, `[{"cap":"document.create"},{"cap":"document.edit"}]`)

	// NewDocument does not wire current_revision_id / active_session_id.
	// Seed a minimal session + initial revision so PresignReserve can verify them.
	initRevID, sessID := rtSeedSessionAndRevision(t, ctx, db, tnt.ID, doc.ID, approver.ID)

	// Build real repo + fake presigner (no S3 contact).
	repo := docrepo.New(db,
		iamdomain.NoopUserDisplayNameReader{},
		controlleddocumentsdomain.NoopCDFieldReader{},
		taxonomydomain.NoopAreaCatalogReader{},
	)
	const contentHash = "roundtrip-hash-abc123"
	presigner := &rtFakePresigner{
		hashReturn: contentHash,
		sizeReturn: 2048,
	}

	svc := application.New(repo, presigner, nil, nil, &noopAudit{}).
		WithEligibility(repo)

	// -----------------------------------------------------------------------
	// Scenario 1: PresignAutosave succeeds for eligible approver.
	// -----------------------------------------------------------------------
	presignRes, err := svc.PresignAutosave(ctx, application.PresignAutosaveCmd{
		TenantID:       tnt.ID,
		ActorUserID:    approver.ID,
		DocumentID:     doc.ID,
		SessionID:      sessID,
		BaseRevisionID: initRevID,
		ContentHash:    contentHash,
	})
	if err != nil {
		t.Fatalf("PresignAutosave(approver): %v", err)
	}
	if presignRes == nil || presignRes.PendingUploadID == "" {
		t.Fatal("PresignAutosave(approver): empty pending upload ID")
	}

	// -----------------------------------------------------------------------
	// Scenario 2: CommitAutosave succeeds; new revision row created;
	// revision_number UNCHANGED.
	// -----------------------------------------------------------------------
	commitRes, err := svc.CommitAutosave(ctx, application.CommitAutosaveCmd{
		TenantID:         tnt.ID,
		ActorUserID:      approver.ID,
		DocumentID:       doc.ID,
		SessionID:        sessID,
		PendingUploadID:  presignRes.PendingUploadID,
		FormDataSnapshot: []byte(`{"reviewed":true}`),
	})
	if err != nil {
		t.Fatalf("CommitAutosave(approver): %v", err)
	}
	if commitRes == nil || commitRes.RevisionID == "" {
		t.Fatal("CommitAutosave(approver): empty revision ID in result")
	}

	// Verify current_revision_id advanced.
	var newRevID string
	if err := db.QueryRowContext(ctx,
		`SELECT current_revision_id::text FROM public.documents WHERE id=$1::uuid`,
		doc.ID,
	).Scan(&newRevID); err != nil {
		t.Fatalf("query documents.current_revision_id: %v", err)
	}
	if newRevID == initRevID {
		t.Fatalf("current_revision_id did not advance after CommitAutosave; still %s", initRevID)
	}
	if newRevID != commitRes.RevisionID {
		t.Fatalf("current_revision_id = %q, want CommitResult.RevisionID = %q", newRevID, commitRes.RevisionID)
	}

	// Verify the new document_revisions row exists.
	var revRowCount int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM public.document_revisions WHERE document_id=$1::uuid AND id=$2::uuid`,
		doc.ID, commitRes.RevisionID,
	).Scan(&revRowCount); err != nil {
		t.Fatalf("query document_revisions count: %v", err)
	}
	if revRowCount != 1 {
		t.Fatalf("expected 1 document_revisions row for new revision, got %d", revRowCount)
	}

	// Verify revision_number UNCHANGED (working-content autosave must not bump
	// the governed sequence; only a finalize/approve path creates a new REV).
	var revNum int
	if err := db.QueryRowContext(ctx,
		`SELECT revision_number FROM public.documents WHERE id=$1::uuid`,
		doc.ID,
	).Scan(&revNum); err != nil {
		t.Fatalf("query documents.revision_number: %v", err)
	}
	if revNum != doc.RevisionNumber {
		t.Fatalf("revision_number changed %d→%d; approver autosave must not create a governed REV",
			doc.RevisionNumber, revNum)
	}

	// -----------------------------------------------------------------------
	// Scenario 3: PresignAutosave rejected for non-eligible actor.
	// -----------------------------------------------------------------------
	_, err = svc.PresignAutosave(ctx, application.PresignAutosaveCmd{
		TenantID:       tnt.ID,
		ActorUserID:    stranger.ID,
		DocumentID:     doc.ID,
		SessionID:      sessID,
		BaseRevisionID: initRevID,
		ContentHash:    contentHash,
	})
	if !errors.Is(err, domain.ErrInvalidStateTransition) {
		t.Fatalf("PresignAutosave(stranger) = %v, want ErrInvalidStateTransition", err)
	}

	// -----------------------------------------------------------------------
	// Scenario 4: 'rejected' is a removed DB status (migration
	// 0272_documents_remove_rejected.sql). The reject path collapses
	// under_review straight back to 'draft' (no intermediate 'rejected' hop),
	// matching documents/domain.CanTransitionDocumentStatus (state.go), which
	// has no rejected arcs. This asserts the DB trigger now REJECTS
	// under_review→rejected with a check_violation, proving app<->DB parity
	// at the schema boundary (not just the Go domain layer).
	// -----------------------------------------------------------------------

	_, err = db.ExecContext(ctx,
		`UPDATE public.documents SET status='rejected' WHERE id=$1::uuid AND status='under_review'`,
		doc.ID,
	)
	if err == nil {
		t.Fatal("scenario 4: under_review→rejected succeeded; want check_violation (rejected status removed by migration 0272)")
	}
	var pqErr interface{ SQLState() string }
	if errors.As(err, &pqErr) {
		if pqErr.SQLState() != "23514" && pqErr.SQLState() != "P0001" {
			t.Fatalf("scenario 4: under_review→rejected failed with unexpected SQLSTATE %q, want check_violation (23514) or raised exception (P0001): %v", pqErr.SQLState(), err)
		}
	}

	// Confirm the document is unaffected (still under_review) since the
	// UPDATE was rejected by the trigger/CHECK constraint.
	var statusAfter string
	if err := db.QueryRowContext(ctx,
		`SELECT status FROM public.documents WHERE id=$1::uuid`, doc.ID,
	).Scan(&statusAfter); err != nil {
		t.Fatalf("scenario 4: query status after rejected UPDATE attempt: %v", err)
	}
	if statusAfter != "under_review" {
		t.Fatalf("scenario 4: expected status still under_review after blocked transition, got %q", statusAfter)
	}
}

// rtSeedSessionAndRevision inserts an editor_session owned by userID, an initial
// document_revision, and wires documents.current_revision_id + active_session_id.
// Returns (revID, sessID).
//
// NewDocument does not wire the autosave-pointer columns, so this helper is
// required before any PresignReserve or CommitUpload call at the service level.
func rtSeedSessionAndRevision(t *testing.T, ctx context.Context, db *sql.DB, tenantID, docID, userID string) (string, string) {
	t.Helper()

	sessID := uuid.NewString()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO public.editor_sessions
		   (id, tenant_id, document_id, user_id, expires_at,
		    last_acknowledged_revision_id, status)
		 VALUES ($1::uuid, $2::uuid, $3::uuid, $4,
		         now() + interval '2 hours',
		         '00000000-0000-0000-0000-000000000000', 'active')`,
		sessID, tenantID, docID, userID,
	); err != nil {
		t.Fatalf("rtSeedSessionAndRevision: insert editor_session: %v", err)
	}

	revID := uuid.NewString()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO public.document_revisions
		   (id, document_id, parent_revision_id, session_id,
		    storage_key, content_hash, form_data_snapshot)
		 VALUES ($1::uuid, $2::uuid, NULL, $3::uuid,
		         $4, 'initial-hash', '{}')`,
		revID, docID, sessID, "tenants/test/roundtrip/"+revID+".docx",
	); err != nil {
		t.Fatalf("rtSeedSessionAndRevision: insert document_revision: %v", err)
	}

	// Wire document pointers (document.edit tripwire — satisfied by SetCapsOnDB
	// called before this helper).
	if _, err := db.ExecContext(ctx,
		`UPDATE public.documents
		    SET current_revision_id=$1::uuid, active_session_id=$2::uuid, updated_at=now()
		  WHERE id=$3::uuid`,
		revID, sessID, docID,
	); err != nil {
		t.Fatalf("rtSeedSessionAndRevision: update document pointers: %v", err)
	}

	// Ack the session to the initial revision so PresignReserve's stale-base
	// check (sessAck == baseRevisionID) passes.
	if _, err := db.ExecContext(ctx,
		`UPDATE public.editor_sessions
		    SET last_acknowledged_revision_id=$1::uuid
		  WHERE id=$2::uuid`,
		revID, sessID,
	); err != nil {
		t.Fatalf("rtSeedSessionAndRevision: update session ack: %v", err)
	}

	return revID, sessID
}

// rtFakePresigner satisfies application.Presigner without contacting S3.
// Confirm returns the pre-agreed hash so CommitAutosave's content-hash
// verification passes deterministically.
type rtFakePresigner struct {
	hashReturn string
	sizeReturn int64
}

func (p *rtFakePresigner) PresignPut(_ context.Context, _, key string, _ time.Duration) (string, error) {
	return "https://fake-s3/put/" + key, nil
}

func (p *rtFakePresigner) PresignGet(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://fake-s3/get/" + key, nil
}

func (p *rtFakePresigner) Confirm(_ context.Context, _, key, expected string) (objectstore.VerifiedPointer, error) {
	return objectstore.VerifiedPointer{StorageKey: key, ContentHash: expected, SizeBytes: p.sizeReturn}, nil
}

func (p *rtFakePresigner) Exists(_ context.Context, _ string) (bool, error) { return true, nil }

func (p *rtFakePresigner) Size(_ context.Context, _ string) (int64, error) {
	return p.sizeReturn, nil
}

func (p *rtFakePresigner) Delete(_ context.Context, _ string) error { return nil }

var _ application.Presigner = (*rtFakePresigner)(nil)
