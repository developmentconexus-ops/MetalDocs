//go:build integration
// +build integration

package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/modules/templates/infrastructure"
	"metaldocs/tests/integration/testdb"
)

// TestMarkTemplateVersionUnderReview_ContentHashPreconditionInCAS_Live is the
// real-DB proof of the submit-lock TOCTOU fix. The kernel submit path reads the
// version's content hash as a friendly fast path and only afterwards runs the
// draft -> under_review CAS; under Read Committed a concurrent author edit can
// clear content_hash in that window. This test reproduces exactly that state —
// a draft row whose hash was cleared after the fast-path read would have
// passed — and drives the REAL writer:
//
//   - the CAS must lose (0 rows) instead of flipping the row and tripping
//     chk_template_version_content_hash_non_draft as a raw 23514 (which used to
//     surface as a 500 with the constraint name in the problem title),
//   - the error must be approval's ErrTemplateVersionNoContent sentinel (409),
//   - the row must still be draft.
//
// No sleeps and no concurrent goroutine: the race's OUTCOME is what the fix is
// about, and clearing the hash before the writer call reproduces it
// deterministically. (content_hash is NOT NULL, so "cleared" is '' — the same
// value the legacy docx-upload-url direct-PUT path leaves behind.)
func TestMarkTemplateVersionUnderReview_ContentHashPreconditionInCAS_Live(t *testing.T) {
	ctx := context.Background()
	dbc, _ := testdb.Open(t)

	tnt := testdb.NewTenant(t, dbc)
	actorID := testdb.DeterministicID(t, "acw-actor")
	templateID := testdb.DeterministicID(t, "acw-template")
	racedVersionID := testdb.DeterministicID(t, "acw-version-raced")
	intactVersionID := testdb.DeterministicID(t, "acw-version-intact")
	contentHash := strings.Repeat("1", 64)

	// templates_template / templates_template_version carry the template.*
	// tripwire and trg_template_version_tenant_consistent, so both writes need
	// the cap asserted tx-locally AND metaldocs.tenant_id seeded on the same tx.
	testdb.SeedWithCaps(t, dbc, `[{"cap":"template.create"}]`, func(tx *sql.Tx) error {
		if err := authz.SeedTxIdentity(ctx, tx, tnt.ID, actorID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.templates_template (
				id, tenant_id, doc_type_code, key, name, latest_version, published_version_id, created_by
			) VALUES (
				$1::uuid, $2::uuid, 'po', 'acw-template', 'ACW Template', 2, NULL, $3
			)`,
			templateID, tnt.ID, actorID,
		); err != nil {
			return err
		}
		// docx_storage_key is globally unique
		// (uq_templates_template_version_docx_storage_key), so key it off the
		// per-test deterministic version id; revision_number is unique per
		// template (ux_templates_version_revision), so the two rows differ in
		// both version_number and revision_number.
		for revisionNumber, versionID := range map[int]string{0: racedVersionID, 1: intactVersionID} {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO public.templates_template_version (
					id, tenant_id, template_id, version_number, revision_number, status, docx_storage_key,
					content_hash, metadata_schema, placeholder_schema, author_id
				) VALUES (
					$1::uuid, $2::uuid, $3::uuid, $4, $5, 'draft', $6,
					$7, '{}'::jsonb, '{"placeholders":[]}'::jsonb, $8
				)`,
				versionID, tnt.ID, templateID, revisionNumber+1, revisionNumber,
				"templates/acw/"+versionID+"/body.docx", contentHash, actorID,
			); err != nil {
				return err
			}
		}
		return nil
	})

	// The concurrent edit: clear the committed hash on the raced version only.
	// Legal for a draft row (chk_template_version_content_hash_non_draft demands
	// the 64-char hash only for non-draft statuses) — which is precisely why the
	// unguarded CAS used to carry it into under_review and blow up mid-tx.
	testdb.SeedWithCaps(t, dbc, `[{"cap":"template.edit"}]`, func(tx *sql.Tx) error {
		if err := authz.SeedTxIdentity(ctx, tx, tnt.ID, actorID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `
			UPDATE public.templates_template_version
			   SET content_hash = ''
			 WHERE id = $1::uuid AND tenant_id = $2::uuid`,
			racedVersionID, tnt.ID,
		)
		return err
	})

	writer := infrastructure.NewApprovalCompletionWriter()

	markUnderReview := func(t *testing.T, versionID string) error {
		t.Helper()
		tx, err := dbc.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer tx.Rollback()

		// Mirrors the production submit tx: template.submit is already asserted
		// by TemplateSubmitService's authz.Require before the writer runs.
		testdb.SetCapsOnTx(t, tx, `[{"cap":"template.submit"}]`)
		if err := authz.SeedTxIdentity(ctx, tx, tnt.ID, actorID); err != nil {
			t.Fatalf("seed tx identity: %v", err)
		}

		markErr := writer.MarkTemplateVersionUnderReview(ctx, tx, tnt.ID, versionID)
		if markErr != nil {
			return markErr
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit: %v", err)
		}
		return nil
	}

	statusOf := func(t *testing.T, versionID string) string {
		t.Helper()
		var status string
		if err := dbc.QueryRowContext(ctx,
			`SELECT status FROM public.templates_template_version WHERE id = $1::uuid AND tenant_id = $2::uuid`,
			versionID, tnt.ID,
		).Scan(&status); err != nil {
			t.Fatalf("query status: %v", err)
		}
		return status
	}

	t.Run("cleared_hash_loses_cas_with_no_content_sentinel", func(t *testing.T) {
		err := markUnderReview(t, racedVersionID)
		if !errors.Is(err, approvaldomain.ErrTemplateVersionNoContent) {
			t.Fatalf("err = %v, want approval domain ErrTemplateVersionNoContent (never a 23514 constraint error)", err)
		}
		if strings.Contains(strings.ToLower(err.Error()), "chk_template_version_content_hash_non_draft") {
			t.Fatalf("constraint name leaked into the error: %v", err)
		}
		if got := statusOf(t, racedVersionID); got != "draft" {
			t.Fatalf("status = %q, want draft (the CAS must not have transitioned the row)", got)
		}
	})

	t.Run("intact_hash_still_transitions", func(t *testing.T) {
		if err := markUnderReview(t, intactVersionID); err != nil {
			t.Fatalf("MarkTemplateVersionUnderReview: %v", err)
		}
		if got := statusOf(t, intactVersionID); got != "under_review" {
			t.Fatalf("status = %q, want under_review (the new predicate must not block a valid submit)", got)
		}
	})
}
