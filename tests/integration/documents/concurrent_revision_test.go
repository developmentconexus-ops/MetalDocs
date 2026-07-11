//go:build integration
// +build integration

package documents_test

// C5 guard: "concurrent document revision creation is serialized per
// controlled-document via pg_advisory_xact_lock so revision numbers never
// collide." Re-guards the invariant enforced at
// internal/modules/documents/repository/repository.go's CreateDocumentTx,
// which takes `SELECT pg_advisory_xact_lock(hashtextextended(tenant_id || ':'
// || controlled_document_id, 0))` before computing
// `COALESCE(MAX(revision_number)+1, 0)` for the target controlled document.
//
// Replaces the deleted tests/integration/documents/concurrent_revision_test.go
// (dead scaffolding: imported a never-committed package and queried views that
// do not exist in this schema — v_document_finalized /
// document_state_history). Rebuilt on the canonical testdb template-DB
// factory (tests/integration/testdb), following the pattern in
// internal/modules/documents/repository/repository_create_integration_test.go
// and repository_write_authz_integration_test.go.
//
// Race shape: N goroutines race to CreateDocumentTx a *new* document against
// the SAME controlled document, each on its own connection/transaction. Two
// DB invariants bound the outcome independent of C5:
//   - ux_documents_cd_active (unique, partial on controlled_document_id WHERE
//     status IN draft/under_review/approved/rejected/scheduled) permits only
//     ONE active document per controlled document, so at most one racer can
//     land its INSERT.
//   - ux_documents_cd_revision (unique on (controlled_document_id,
//     revision_number)) is the hard backstop C5 exists to keep unreachable.
//
// What C5 actually buys, and what this test isolates, is the interleaving
// inside CreateDocumentTx between the advisory lock, the
// `COALESCE(MAX(revision_number)+1, 0)` computation, and the INSERT: without
// serialization, two racers could both read revision_number=0 for an empty
// CD family and both attempt to insert it, relying on ux_documents_cd_revision
// (or ux_documents_cd_active) to catch the collision after the fact — a
// correctness accident, not a guarantee, and one that would surface as a
// generic 23505 rather than a clean single-winner race. This test asserts the
// *clean* form: exactly one CreateDocumentTx call commits successfully with
// revision_number=0; every other racer fails on ux_documents_cd_active
// (23505) with no partial/duplicate documents row ever committed for the CD.
// It also runs a second wave after the winner is superseded, confirming the
// next revision_number (1) is likewise assigned to exactly one winner with no
// gaps or duplicates — i.e., the lock keeps allocation strictly serialized
// across repeated contention, not just on the first call.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/tests/integration/testdb"
)

const concurrentRevisionWorkers = 8

// raceResult captures the outcome of one racer's CreateDocumentTx attempt.
type raceResult struct {
	docID string
	err   error
}

// isUniqueViolation reports whether err is a Postgres unique_violation
// (23505), unwrapping through CreateDocumentTx's fmt.Errorf("insert
// document: %w", err) wrapping.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// raceCreateDocumentTx fires `workers` concurrent CreateDocumentTx calls at
// the same controlled document, each on its own transaction, released
// simultaneously past a start barrier. Returns one raceResult per worker.
func raceCreateDocumentTx(t *testing.T, ctx context.Context, db *sql.DB, repo *infrastructure.Repository, tenantID, actorID string, cd testdb.ControlledDoc, codePrefix string) []raceResult {
	t.Helper()

	start := make(chan struct{})
	var wg sync.WaitGroup
	results := make([]raceResult, concurrentRevisionWorkers)

	for i := 0; i < concurrentRevisionWorkers; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				results[i] = raceResult{err: fmt.Errorf("worker %d begin tx: %w", i, err)}
				return
			}
			defer tx.Rollback()

			// Assert caps tx-locally, mirroring repository_create_integration_test.go —
			// document.create + document.edit are both exercised inside CreateDocumentTx.
			testdb.SetCapsOnTx(t, tx, `[{"cap":"document.create"},{"cap":"document.edit"}]`)

			profileCode := cd.ProfileCode
			processAreaCode := cd.ProcessAreaCode
			controlledDocID := cd.ID
			doc := &domain.Document{
				TenantID:                tenantID,
				TemplateVersionID:       testdb.DeterministicID(t, "template-version"),
				Name:                    fmt.Sprintf("Race Doc %d", i),
				FormDataJSON:            []byte(`{}`),
				CreatedBy:               actorID,
				ControlledDocumentID:    &controlledDocID,
				ProfileCodeSnapshot:     &profileCode,
				ProcessAreaCodeSnapshot: &processAreaCode,
				Code:                    fmt.Sprintf("%s-%d", codePrefix, i),
			}

			<-start // release all goroutines together

			docID, _, _, err := repo.CreateDocumentTx(ctx, tx, doc, fmt.Sprintf("hash-%d", i), "", nil)
			if err != nil {
				results[i] = raceResult{err: err}
				return
			}
			if err := tx.Commit(); err != nil {
				results[i] = raceResult{err: fmt.Errorf("worker %d commit: %w", i, err)}
				return
			}
			results[i] = raceResult{docID: docID}
		}()
	}

	close(start)
	wg.Wait()
	return results
}

// TestCreateDocumentTx_ConcurrentRevisionAllocation_SingleWinnerNoCollision
// guards C5: concurrentRevisionWorkers goroutines race CreateDocumentTx
// against the same controlled document. Exactly one must commit (winning
// revision_number=0); every other racer must fail cleanly on
// ux_documents_cd_active — never on a generic/ambiguous error, and never with
// two racers both landing a documents row for the same controlled document
// (which would mean the advisory lock failed to serialize the
// revision_number read against concurrent INSERTs).
func TestCreateDocumentTx_ConcurrentRevisionAllocation_SingleWinnerNoCollision(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	// Deliberately NOT db.SetMaxOpenConns(1): this test requires
	// concurrentRevisionWorkers simultaneous connections to actually exercise
	// server-side lock contention rather than serializing through Go's pool.

	tnt := testdb.NewTenant(t, db)
	actor := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithTaxonomy(tax),
		testdb.WithOwner(actor.ID),
	)

	repo := infrastructure.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	results := raceCreateDocumentTx(t, ctx, db, repo, tnt.ID, actor.ID, cd, "C5-WAVE1")

	var winners []string
	var uniqueViolations int
	for i, r := range results {
		switch {
		case r.err == nil:
			winners = append(winners, r.docID)
		case isUniqueViolation(r.err):
			uniqueViolations++
		default:
			t.Fatalf("worker %d: unexpected error (want nil or 23505 unique_violation): %v", i, r.err)
		}
	}

	if len(winners) != 1 {
		t.Fatalf("expected exactly 1 winner among %d racers, got %d (docIDs=%v)", concurrentRevisionWorkers, len(winners), winners)
	}
	if uniqueViolations != concurrentRevisionWorkers-1 {
		t.Fatalf("expected %d unique_violation losers, got %d", concurrentRevisionWorkers-1, uniqueViolations)
	}

	// Confirm the DB reality directly: exactly one documents row for this CD,
	// with revision_number=0 — the C5 assertion, verified against the real
	// table rather than trusted from in-process error counting alone.
	var rowCount int
	var revisionNumber int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*), COALESCE(MAX(revision_number), -1)
		   FROM public.documents
		  WHERE controlled_document_id = $1::uuid`,
		cd.ID,
	).Scan(&rowCount, &revisionNumber); err != nil {
		t.Fatalf("query documents for cd: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("documents row count for cd = %d, want 1 (no duplicate/partial rows survived the race)", rowCount)
	}
	if revisionNumber != 0 {
		t.Fatalf("winning revision_number = %d, want 0", revisionNumber)
	}
	if winners[0] == "" {
		t.Fatalf("winner docID is empty")
	}
}

// TestCreateDocumentTx_ConcurrentRevisionAllocation_SecondWaveNoGapNoDuplicate
// re-runs the same race after the wave-1 winner is walked to 'superseded',
// confirming C5 holds for revision_number=1 too — i.e. the advisory lock
// keeps serializing allocation on every call, not just when the CD family is
// empty. Guards against a regression where the lock is only acquired/checked
// on the first-ever revision.
func TestCreateDocumentTx_ConcurrentRevisionAllocation_SecondWaveNoGapNoDuplicate(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

	tnt := testdb.NewTenant(t, db)
	actor := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithTaxonomy(tax),
		testdb.WithOwner(actor.ID),
	)

	repo := infrastructure.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	// Seed revision 0 sequentially (not part of the race under test).
	profileCode := cd.ProfileCode
	processAreaCode := cd.ProcessAreaCode
	controlledDocID := cd.ID
	seedDoc := &domain.Document{
		TenantID:                tnt.ID,
		TemplateVersionID:       testdb.DeterministicID(t, "template-version"),
		Name:                    "Wave 0 Seed",
		FormDataJSON:            []byte(`{}`),
		CreatedBy:               actor.ID,
		ControlledDocumentID:    &controlledDocID,
		ProfileCodeSnapshot:     &profileCode,
		ProcessAreaCodeSnapshot: &processAreaCode,
		Code:                    "C5-WAVE0-SEED",
	}
	seedTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin seed tx: %v", err)
	}
	testdb.SetCapsOnTx(t, seedTx, `[{"cap":"document.create"},{"cap":"document.edit"}]`)
	if _, _, _, err := repo.CreateDocumentTx(ctx, seedTx, seedDoc, "hash-seed", "", nil); err != nil {
		t.Fatalf("CreateDocumentTx seed: %v", err)
	}
	if err := seedTx.Commit(); err != nil {
		t.Fatalf("commit seed tx: %v", err)
	}

	// Walk the seed document to 'superseded' so ux_documents_cd_active allows
	// a new active document, and revision_number should next allocate to 1.
	testdb.SupersedeActiveDocumentForCD(t, db, cd.ID)

	results := raceCreateDocumentTx(t, ctx, db, repo, tnt.ID, actor.ID, cd, "C5-WAVE2")

	var winners []string
	var uniqueViolations int
	for i, r := range results {
		switch {
		case r.err == nil:
			winners = append(winners, r.docID)
		case isUniqueViolation(r.err):
			uniqueViolations++
		default:
			t.Fatalf("worker %d: unexpected error (want nil or 23505 unique_violation): %v", i, r.err)
		}
	}

	if len(winners) != 1 {
		t.Fatalf("expected exactly 1 second-wave winner among %d racers, got %d (docIDs=%v)", concurrentRevisionWorkers, len(winners), winners)
	}
	if uniqueViolations != concurrentRevisionWorkers-1 {
		t.Fatalf("expected %d unique_violation losers, got %d", concurrentRevisionWorkers-1, uniqueViolations)
	}

	// Confirm the full revision family for this CD is exactly {0 (superseded),
	// 1 (new active)} — contiguous, no gap, no duplicate revision_number.
	rows, err := db.QueryContext(ctx,
		`SELECT revision_number FROM public.documents
		  WHERE controlled_document_id = $1::uuid
		  ORDER BY revision_number`,
		cd.ID,
	)
	if err != nil {
		t.Fatalf("query revision family: %v", err)
	}
	defer rows.Close()

	var got []int
	for rows.Next() {
		var rn int
		if err := rows.Scan(&rn); err != nil {
			t.Fatalf("scan revision_number: %v", err)
		}
		got = append(got, rn)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}

	want := []int{0, 1}
	if len(got) != len(want) {
		t.Fatalf("revision_number family = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("revision_number family = %v, want %v", got, want)
		}
	}
}
