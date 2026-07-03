//go:build integration
// +build integration

// Tripwire-arm regression gates (register APP-08 / APP-09).
//
// enforce_capability_asserted()'s per-table cap lists are hand-synced to the
// capabilities the service layer asserts via authz.Require. A missing arm
// fail-closes the write with SQLSTATE P0001 for EVERY actor (the trigger
// checks the recorded asserted-cap set, not the role), bricking the path as
// a 500. This has happened twice, both found only by live QA:
//   - 0269: 'template.review' missing on templates_template_version — the
//     reviewer stage (Service.Review) was DB-bricked.
//   - 0270: 'template.archive' missing on templates_template —
//     POST /templates/{id}/archive returned 500 even for system_admin.
// Application tests are sqlmock, so only an integration drive against a
// tripwire-enforced DB can pin these. These tests drive the exact two
// write paths whose caps were missing; a future arm regression (e.g. a
// migration recreating the function from a stale copy) fails here with
// ErrCapabilityNotAsserted instead of in production.
package templates_test

import (
	"context"
	"database/sql"
	"testing"

	platformdb "metaldocs/internal/platform/db"

	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/modules/templates/repository"
	"metaldocs/tests/integration/testdb"
)

// seedContentHashTx stamps a committed content hash onto a version row so the
// submit gate (T-004) passes without a real autosave; same recipe as
// lifecycle_no_auto_draft_test.go.
func seedContentHashTx(t *testing.T, db *sql.DB, tenantID, versionID string) {
	t.Helper()
	const hash = "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
	testdb.SeedWithCaps(t, db, `[{"cap":"template.edit"}]`, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(context.Background(),
			`SELECT set_config('metaldocs.tenant_id', $1, true)`, tenantID,
		); err != nil {
			return err
		}
		_, err := tx.ExecContext(context.Background(),
			`UPDATE public.templates_template_version SET content_hash = $2 WHERE id = $1::uuid`,
			versionID, hash,
		)
		return err
	})
}

// TestTripwire_ReviewerStageWritesVersionRow pins migration 0269: the
// reviewer-stage accept (Service.Review, asserts template.review) must be
// able to UPDATE templates_template_version under the live tripwire.
func TestTripwire_ReviewerStageWritesVersionRow(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	author := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID)).ID
	reviewer := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID)).ID
	testdb.SeedSystemAdmin(t, db, tenant.ID, author, "Tripwire Test Author")
	testdb.SeedSystemAdmin(t, db, tenant.ID, reviewer, "Tripwire Test Reviewer")

	repo := repository.New(db)
	svc := application.New(repo, noopPresigner{}, testClock{}, testUUID{}).
		WithRunner(platformdb.NewTxRunner(db))

	reviewerRole := "reviewer"
	createRes, err := svc.CreateTemplate(ctx, application.CreateTemplateCmd{
		TenantID:     tenant.ID,
		ActorUserID:  author,
		DocTypeCode:  "po",
		Key:          "tpl-tripwire-rev-" + testdb.DeterministicID(t, "key")[:8],
		Name:         "Tripwire Reviewer-Stage Template",
		ApproverRole: "approver",
		ReviewerRole: &reviewerRole,
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	templateID := createRes.Template.ID
	seedContentHashTx(t, db, tenant.ID, createRes.Version.ID)

	if _, err := svc.SubmitForReview(ctx, application.SubmitForReviewCmd{
		TenantID:      tenant.ID,
		ActorUserID:   author,
		TemplateID:    templateID,
		VersionNumber: 1,
	}); err != nil {
		t.Fatalf("SubmitForReview: %v", err)
	}

	// The pinned write: reviewer accept UPDATEs templates_template_version
	// under asserted cap template.review. Pre-0269 this raised P0001.
	reviewed, err := svc.Review(ctx, application.ReviewCmd{
		TenantID:      tenant.ID,
		ActorUserID:   reviewer,
		ActorRoles:    []string{reviewerRole},
		TemplateID:    templateID,
		VersionNumber: 1,
		Accept:        true,
	})
	if err != nil {
		t.Fatalf("Review accept (0269 tripwire arm regression?): %v", err)
	}
	if reviewed.Status != domain.VersionStatusApproved {
		t.Fatalf("after review accept: status=%q want=approved", reviewed.Status)
	}
}

// TestTripwire_ArchiveWritesTemplateRow pins migration 0270: archive
// (Service.ArchiveTemplate, asserts template.archive) must be able to
// UPDATE templates_template under the live tripwire.
func TestTripwire_ArchiveWritesTemplateRow(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	actor := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID)).ID
	testdb.SeedSystemAdmin(t, db, tenant.ID, actor, "Tripwire Test Archiver")

	repo := repository.New(db)
	svc := application.New(repo, noopPresigner{}, testClock{}, testUUID{}).
		WithRunner(platformdb.NewTxRunner(db))

	createRes, err := svc.CreateTemplate(ctx, application.CreateTemplateCmd{
		TenantID:     tenant.ID,
		ActorUserID:  actor,
		DocTypeCode:  "po",
		Key:          "tpl-tripwire-arch-" + testdb.DeterministicID(t, "key")[:8],
		Name:         "Tripwire Archive Template",
		ApproverRole: "approver",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	templateID := createRes.Template.ID

	// The pinned write: archive UPDATEs templates_template under asserted
	// cap template.archive. Pre-0270 this raised P0001 for every actor.
	archived, err := svc.ArchiveTemplate(ctx, application.ArchiveCmd{
		TenantID:    tenant.ID,
		ActorUserID: actor,
		TemplateID:  templateID,
	})
	if err != nil {
		t.Fatalf("ArchiveTemplate (0270 tripwire arm regression?): %v", err)
	}
	if archived.ArchivedAt == nil {
		t.Fatalf("ArchiveTemplate returned nil ArchivedAt")
	}

	var archivedAt sql.NullTime
	if err := db.QueryRowContext(ctx,
		`SELECT archived_at FROM public.templates_template WHERE id = $1::uuid`,
		templateID,
	).Scan(&archivedAt); err != nil {
		t.Fatalf("read archived_at: %v", err)
	}
	if !archivedAt.Valid {
		t.Fatalf("archived_at not persisted on templates_template")
	}
}
