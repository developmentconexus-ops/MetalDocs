//go:build integration
// +build integration

// Tripwire-arm regression gates for the documents(UPDATE) arm (register the
// M2 F2.1 §1.1 fixes, migration 0271).
//
// enforce_capability_asserted()'s documents(UPDATE) arm was missing two caps
// that two write paths assert as their SOLE tier-2 capability (neither
// co-asserts document.edit, unlike publish/supersede/submit/decision):
//   - membership.manage: ForceReleaseSession / ForceReleaseSessionTx
//     (documents/repository/repository.go:798 / :828) — an admin
//     force-releasing another user's editor lock.
//   - document.obsolete: ObsoleteService.MarkObsolete
//     (documents/approval/application/obsolete_service.go:88) — transitioning
//     a published/superseded document to obsolete.
// A missing arm fail-closes the write with SQLSTATE P0001 for EVERY actor
// (the trigger checks the recorded asserted-cap set, not the role), bricking
// the path as a 500. Both were latent (never shipped fixed, found only by the
// M2 census), unlike the 0269/0270 incidents this mirrors
// (tests/integration/templates/tripwire_caps_test.go).
//
// Application tests are sqlmock, so only an integration drive against a
// tripwire-enforced DB can pin these. These tests drive the exact two write
// paths whose caps were missing from the arm; a future arm regression (e.g. a
// migration recreating the function from a stale copy) fails here with
// ErrCapabilityNotAsserted instead of in production.
package documents_test

import (
	"context"
	"database/sql"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	approvalapp "metaldocs/internal/modules/documents/approval/application"
	approvalrepo "metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	platformdb "metaldocs/internal/platform/db"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/tests/integration/testdb"
)

// TestTripwire_ForceReleaseWritesDocumentRow pins migration 0271's
// membership.manage arm: an admin force-releasing another user's editor
// session (Repository.ForceReleaseSession, asserts ONLY CapMembershipManage)
// must be able to UPDATE documents.active_session_id under the live tripwire.
func TestTripwire_ForceReleaseWritesDocumentRow(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

	tnt := testdb.NewTenant(t, db)
	// system_admin: authz.Require's bypass branch still asserts the real
	// capability (appendAssertedCap) before the UPDATE, so the tripwire is
	// genuinely exercised with membership.manage recorded — the bypass only
	// skips the grant-row lookup, not the GUC assertion the trigger reads.
	admin := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	editor := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID))

	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithTaxonomy(tax),
		testdb.WithOwner(editor.ID),
	)

	r := repository.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	// Seed a document WITH an active editing session (CreateDocumentTx sets
	// documents.active_session_id + editor_sessions.status='active' as part of
	// document creation), owned by editor. Created by the system_admin actor so
	// fixture-build does not itself depend on the behavior under test.
	profileCode := cd.ProfileCode
	processAreaCode := cd.ProcessAreaCode
	controlledDocID := cd.ID
	doc := &domain.Document{
		TenantID:                tnt.ID,
		TemplateVersionID:       testdb.DeterministicID(t, "template-version"),
		Name:                    "Tripwire Force-Release Doc",
		FormDataJSON:            []byte(`{}`),
		CreatedBy:               editor.ID,
		ControlledDocumentID:    &controlledDocID,
		ProfileCodeSnapshot:     &profileCode,
		ProcessAreaCodeSnapshot: &processAreaCode,
		Code:                    "TRIPWIRE-FORCE-RELEASE",
	}

	// Create as the admin identity (not editor) so document creation itself
	// does not depend on editor holding any grants — only force-release
	// (the behavior under test) is driven as a distinct actor.
	createTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin create tx: %v", err)
	}
	defer createTx.Rollback()
	if err := authz.SeedTxIdentity(ctx, createTx, tnt.ID, admin.ID); err != nil {
		t.Fatalf("seed create tx identity: %v", err)
	}
	docID, _, sessionID, err := r.CreateDocumentTx(ctx, createTx, doc, "hash-tripwire-force-release", "", nil)
	if err != nil {
		t.Fatalf("CreateDocumentTx fixture: %v", err)
	}
	if err := createTx.Commit(); err != nil {
		t.Fatalf("commit create tx: %v", err)
	}

	// Sanity: the document actually carries the active session before the
	// pinned write.
	var activeSessionBefore sql.NullString
	if err := db.QueryRowContext(ctx,
		`SELECT active_session_id::text FROM public.documents WHERE id = $1::uuid`, docID,
	).Scan(&activeSessionBefore); err != nil {
		t.Fatalf("read active_session_id before force-release: %v", err)
	}
	if !activeSessionBefore.Valid || activeSessionBefore.String != sessionID {
		t.Fatalf("active_session_id before force-release = %v, want %q", activeSessionBefore, sessionID)
	}

	// The pinned write: force-release UPDATEs documents.active_session_id
	// under asserted cap membership.manage ONLY (no co-asserted document.edit).
	// Pre-0271 this raised P0001 for every actor.
	if err := r.ForceReleaseSession(ctx, tnt.ID, admin.ID, sessionID); err != nil {
		t.Fatalf("ForceReleaseSession (0271 tripwire arm regression?): %v", err)
	}

	var activeSessionAfter sql.NullString
	if err := db.QueryRowContext(ctx,
		`SELECT active_session_id::text FROM public.documents WHERE id = $1::uuid`, docID,
	).Scan(&activeSessionAfter); err != nil {
		t.Fatalf("read active_session_id after force-release: %v", err)
	}
	if activeSessionAfter.Valid {
		t.Fatalf("active_session_id after force-release = %v, want NULL (cleared)", activeSessionAfter)
	}

	var sessStatus string
	if err := db.QueryRowContext(ctx,
		`SELECT status FROM public.editor_sessions WHERE id = $1::uuid`, sessionID,
	).Scan(&sessStatus); err != nil {
		t.Fatalf("read session status: %v", err)
	}
	if sessStatus != "force_released" {
		t.Fatalf("session status = %q, want force_released", sessStatus)
	}
}

// TestTripwire_ObsoleteWritesDocumentRow pins migration 0271's
// document.obsolete arm: ObsoleteService.MarkObsolete on a published document
// (asserts ONLY CapDocumentObsolete) must be able to UPDATE documents.status
// under the live tripwire.
func TestTripwire_ObsoleteWritesDocumentRow(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

	tnt := testdb.NewTenant(t, db)
	admin := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	// Scenario{}.PublishedDocument seeds a document already in 'published'
	// status (with the snapshot columns MarkObsolete's OCC UPDATE needs already
	// stubbed by the factory for any non-draft status), satisfying
	// obsolete_service.go:81's guard without driving the full lifecycle.
	doc := testdb.Scenario{}.PublishedDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithOwner(admin.ID))

	repo := approvalrepo.NewPostgresApprovalRepository(db, iamdomain.NoopUserDisplayNameReader{})
	services := approvalapp.NewServices(repo, approvalapp.NewSQLEmitter(), approvalapp.RealClock{}, controlleddocumentsdomain.NoopCDFieldReader{})
	runner := platformdb.NewTxRunner(db)

	// The pinned write: MarkObsolete UPDATEs documents.status under asserted
	// cap document.obsolete ONLY (no co-asserted document.edit). Pre-0271 this
	// raised P0001 unconditionally.
	res, err := services.Obsolete.MarkObsolete(ctx, runner, approvalapp.MarkObsoleteRequest{
		TenantID:        tnt.ID,
		DocumentID:      doc.ID,
		MarkedBy:        admin.ID,
		RevisionVersion: doc.RevisionVersion,
		Reason:          "tripwire regression drive",
	})
	if err != nil {
		t.Fatalf("MarkObsolete (0271 tripwire arm regression?): %v", err)
	}
	if res.PriorStatus != "published" {
		t.Fatalf("MarkObsolete PriorStatus = %q, want published", res.PriorStatus)
	}

	var status string
	if err := db.QueryRowContext(ctx,
		`SELECT status FROM public.documents WHERE id = $1::uuid`, doc.ID,
	).Scan(&status); err != nil {
		t.Fatalf("read status after MarkObsolete: %v", err)
	}
	if status != "obsolete" {
		t.Fatalf("documents.status after MarkObsolete = %q, want obsolete", status)
	}
}
