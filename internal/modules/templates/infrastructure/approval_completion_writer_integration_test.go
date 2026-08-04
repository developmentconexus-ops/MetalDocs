//go:build integration
// +build integration

package infrastructure_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/modules/templates/infrastructure"
	"metaldocs/tests/integration/testdb"
)

// TestMarkTemplateVersionUnderReview_ContentHashPreconditionInCAS_Live drives
// the REAL writer against a real DB.
//
// It used to reproduce a submit-lock TOCTOU: the kernel reads the version's
// content hash as a friendly fast path and only afterwards runs the draft ->
// under_review CAS, so under Read Committed a concurrent author edit could
// clear content_hash in that window and the CAS would carry an empty hash into
// under_review, tripping chk_template_version_content_hash_non_draft as a raw
// 23514. The fix put the hash predicate INTO the CAS, so it lost cleanly with
// approval's ErrTemplateVersionNoContent sentinel instead.
//
// ADR 0088 / migration 0317 moved that guarantee one rung further down: the
// CHECK is now unconditional (length(content_hash) = 64 in EVERY status), so
// the concurrent edit that this test used to perform is itself rejected by the
// database. The raced state is no longer constructible, and the arm that
// reproduced it is replaced by a proof of that — the writer's in-CAS predicate
// stays as the friendly first line (CLAUDE.md: the DB enforces, the app is
// polite about it), it is simply no longer reachable.
// TestMigration0317_RejectsContentLessInsert covers the INSERT half.
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

	// The TOCTOU premise itself is now unrepresentable: post-0317 the database
	// refuses the concurrent edit that used to create the raced state, so the
	// in-CAS predicate can never be the thing that catches it.
	t.Run("clearing_a_draft_hash_is_rejected_by_the_database", func(t *testing.T) {
		tx, err := dbc.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer tx.Rollback()

		testdb.SetCapsOnTx(t, tx, `[{"cap":"template.edit"}]`)
		if err := authz.SeedTxIdentity(ctx, tx, tnt.ID, actorID); err != nil {
			t.Fatalf("seed tx identity: %v", err)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE public.templates_template_version
			   SET content_hash = ''
			 WHERE id = $1::uuid AND tenant_id = $2::uuid`,
			racedVersionID, tnt.ID,
		)
		if err == nil {
			t.Fatal("clearing content_hash on a draft succeeded; ADR 0088 / migration 0317 requires length(content_hash) = 64 in every status")
		}
		if !strings.Contains(err.Error(), "chk_template_version_content_hash_non_draft") {
			t.Fatalf("err = %v, want chk_template_version_content_hash_non_draft", err)
		}
		if got := statusOf(t, racedVersionID); got != "draft" {
			t.Fatalf("status = %q, want draft", got)
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
