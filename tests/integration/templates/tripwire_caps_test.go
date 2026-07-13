//go:build integration
// +build integration

// Tripwire-arm regression gate (register APP-09; APP-08's gate retired).
//
// enforce_capability_asserted()'s per-table cap lists are hand-synced to the
// capabilities the service layer asserts via authz.Require. A missing arm
// fail-closes the write with SQLSTATE P0001 for EVERY actor (the trigger
// checks the recorded asserted-cap set, not the role), bricking the path as
// a 500. This has happened twice, both found only by live QA:
//   - 0269 (APP-08): 'template.review' missing on templates_template_version —
//     the reviewer stage (Service.Review) was DB-bricked. Its gate test was
//     retired with ADR 0082 phase (a): CreateTemplate no longer seeds the
//     approval config the legacy submit→review flow reads, so that path can
//     no longer be driven end-to-end from a fresh template (slice S5 retires
//     the template.review arm itself).
//   - 0270 (APP-09): 'template.archive' missing on templates_template —
//     POST /templates/{id}/archive returned 500 even for system_admin.
//
// Application tests are sqlmock, so only an integration drive against a
// tripwire-enforced DB can pin this. TestTripwire_ArchiveWritesTemplateRow
// drives the exact write path whose cap was missing; a future arm regression
// (e.g. a migration recreating the function from a stale copy) fails here
// with ErrCapabilityNotAsserted instead of in production.
package templates_test

import (
	"context"
	"database/sql"
	"testing"

	platformdb "metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"

	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/infrastructure"
	"metaldocs/tests/integration/testdb"
)

// TestTripwire_ArchiveWritesTemplateRow pins migration 0270: archive
// (Service.ArchiveTemplate, asserts template.archive) must be able to
// UPDATE templates_template under the live tripwire.
func TestTripwire_ArchiveWritesTemplateRow(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	actor := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID)).ID
	testdb.SeedSystemAdmin(t, db, tenant.ID, actor, "Tripwire Test Archiver")

	// Bind ctx identity for F3.1 TxRunner auto-seed (see reviewer test above).
	ctx = platformtenant.WithActorID(
		platformtenant.WithTenantID(ctx, tenant.ID), actor)

	repo := infrastructure.New(db)
	svc := application.New(repo, noopPresigner{}, testClock{}, testUUID{}).
		WithRunner(platformdb.NewTxRunner(db))

	createRes, err := svc.CreateTemplate(ctx, application.CreateTemplateCmd{
		TenantID:    tenant.ID,
		ActorUserID: actor,
		DocTypeCode: "po",
		Key:         "tpl-tripwire-arch-" + testdb.DeterministicID(t, "key")[:8],
		Name:        "Tripwire Archive Template",
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
