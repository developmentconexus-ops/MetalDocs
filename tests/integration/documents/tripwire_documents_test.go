//go:build integration
// +build integration

// Tripwire-arm regression gates for the documents(UPDATE) arm (register the
// M2 F2.1 §1.1 fixes, migration 0271).
//
// enforce_capability_asserted()'s documents(UPDATE) arm accepted only
// {document.edit}, but two write paths assert a DIFFERENT single capability as
// their SOLE tier-2 capability (neither co-asserts document.edit, unlike
// publish/supersede/submit/decision):
//   - membership.manage: ForceReleaseSession / ForceReleaseSessionTx
//     (documents/repository/repository.go:798 / :828) — an admin
//     force-releasing another user's editor lock (UPDATE active_session_id).
//   - document.obsolete: ObsoleteService.MarkObsolete
//     (documents/approval/application/obsolete_service.go:88 -> :93) —
//     transitioning a document to obsolete (UPDATE status/revision_version).
// A missing arm fail-closes the write with SQLSTATE P0001 for EVERY actor (the
// trigger checks the recorded asserted-cap set, not the role), bricking the
// path as a 500. Both were latent (never shipped fixed; found only by the M2
// census), the same defect class as the 0269/0270 incidents this mirrors
// (tests/integration/templates/tripwire_caps_test.go).
//
// These drive the DB arm DIRECTLY (the proven sibling pattern in
// tripwire_caps_test.go): seed a document, then run the guarded documents
// UPDATE under ONLY the single capability the real path asserts. The status
// column is left unchanged so the sibling triggers on public.documents that
// fire BEFORE trg_require_cap_asserted (trg_documents_legal_transition,
// enforce_snapshot_on_submit_trg, trg_documents_revision_version_monotonic) are
// not the thing under test — the capability arm is. RowsAffected==1 is asserted
// inside the tx so a row hidden by RLS (0 rows, trigger never fires) cannot pass
// as a false green. RED pre-0271 (arm {document.edit} -> P0001); GREEN post-0271.
//
// Application tests are sqlmock and cannot exercise the live trigger; only an
// integration drive against a tripwire-enforced DB pins these. A future arm
// regression (e.g. a migration recreating the function from a stale copy) fails
// here with ErrCapabilityNotAsserted instead of in production.
package documents_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"metaldocs/tests/integration/testdb"
)

// driveGuardedDocumentUpdate runs the given neutral-column documents UPDATE
// under exactly capsJSON, tenant-scoped, and fails the test if the guarded
// write does not affect exactly one row. SeedWithCaps asserts the caps
// tx-locally (pool-safe) and t.Fatalf's if the tripwire rejects the write, so a
// P0001 (the pre-0271 failure mode) surfaces as a test failure with the exact
// ErrCapabilityNotAsserted message.
func driveGuardedDocumentUpdate(t *testing.T, db *sql.DB, tenantID, docID, capsJSON string) {
	t.Helper()
	ctx := context.Background()
	testdb.SeedWithCaps(t, db, capsJSON, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`SELECT set_config('metaldocs.tenant_id', $1, true)`, tenantID,
		); err != nil {
			return fmt.Errorf("set tenant_id GUC: %w", err)
		}
		// active_session_id is a neutral column (no legal-transition / snapshot /
		// monotonic guard); the cap trigger fires on ANY documents UPDATE, so this
		// exercises the arm without dragging in the status state machine.
		res, err := tx.ExecContext(ctx,
			`UPDATE public.documents SET active_session_id = NULL, updated_at = now() WHERE id = $1::uuid`,
			docID,
		)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("rows affected: %w", err)
		}
		if n != 1 {
			return fmt.Errorf("guarded UPDATE affected %d rows, want 1 (RLS hid the row / trigger never fired -> false green)", n)
		}
		return nil
	})
}

// TestTripwire_DocumentsUpdate_MembershipManageArm pins migration 0271's
// membership.manage arm: the ForceReleaseSession write path asserts ONLY
// membership.manage before UPDATEing documents.active_session_id. Pre-0271
// ({document.edit}) this raised P0001 for every actor.
func TestTripwire_DocumentsUpdate_MembershipManageArm(t *testing.T) {
	db, _ := testdb.Open(t)
	tnt := testdb.NewTenant(t, db)
	docID, tenantID := testdb.InsertDraftDocument(t, db, "", tnt.ID)

	driveGuardedDocumentUpdate(t, db, tenantID, docID, `[{"cap":"membership.manage"}]`)
}

// TestTripwire_DocumentsUpdate_DocumentObsoleteArm pins migration 0271's
// document.obsolete arm: the MarkObsolete write path asserts ONLY
// document.obsolete before UPDATEing documents. Pre-0271 ({document.edit}) this
// raised P0001 for every actor.
func TestTripwire_DocumentsUpdate_DocumentObsoleteArm(t *testing.T) {
	db, _ := testdb.Open(t)
	tnt := testdb.NewTenant(t, db)
	docID, tenantID := testdb.InsertDraftDocument(t, db, "", tnt.ID)

	driveGuardedDocumentUpdate(t, db, tenantID, docID, `[{"cap":"document.obsolete"}]`)
}
