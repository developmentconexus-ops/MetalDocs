//go:build integration
// +build integration

// Package templates_test contains integration tests for the templates module lifecycle.
// This file asserts the M1·T2 contract: approve/publish must NOT auto-spawn a next
// draft version row. Manual CreateNextVersion is the only revision-creation path.
package templates_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	platformdb "metaldocs/internal/platform/db"

	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/modules/templates/infrastructure"
	"metaldocs/internal/platform/objectstore"
	platformtenant "metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// noopPresigner satisfies application.Presigner without touching any object store.
// Approve/publish no longer call spawnNextDraft, so Copy is never invoked in the hot path.
type noopPresigner struct{}

func (noopPresigner) PresignPut(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return "https://noop/put", nil
}
func (noopPresigner) PresignGet(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "https://noop/get", nil
}
func (noopPresigner) Confirm(_ context.Context, _, key, hash string) (objectstore.VerifiedPointer, error) {
	return objectstore.VerifiedPointer{StorageKey: key, ContentHash: hash, SizeBytes: 1}, nil
}
func (noopPresigner) Copy(_ context.Context, _ string, _, _ string) error { return nil }
func (noopPresigner) Delete(_ context.Context, _ string) error            { return nil }

// testClock satisfies application.Clock with a fixed timestamp.
type testClock struct{}

func (testClock) Now() time.Time { return time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC) }

// testUUID satisfies application.UUIDGen using the google/uuid package.
type testUUID struct{}

func (testUUID) New() string { return uuid.NewString() }

// countVersionRows returns the number of template_version rows for a given template.
func countVersionRows(t *testing.T, db *sql.DB, templateID string) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(context.Background(),
		`SELECT count(*) FROM public.templates_template_version WHERE template_id = $1::uuid`,
		templateID,
	).Scan(&n); err != nil {
		t.Fatalf("countVersionRows: %v", err)
	}
	return n
}

// versionStatuses returns a map of version_num → status for all versions of a template.
func versionStatuses(t *testing.T, db *sql.DB, templateID string) map[int]string {
	t.Helper()
	rows, err := db.QueryContext(context.Background(),
		`SELECT version_number, status FROM public.templates_template_version
		  WHERE template_id = $1::uuid ORDER BY version_number`,
		templateID,
	)
	if err != nil {
		t.Fatalf("versionStatuses query: %v", err)
	}
	defer rows.Close()
	m := make(map[int]string)
	for rows.Next() {
		var n int
		var s string
		if err := rows.Scan(&n, &s); err != nil {
			t.Fatalf("versionStatuses scan: %v", err)
		}
		m[n] = s
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("versionStatuses rows: %v", err)
	}
	return m
}

// TestLifecycle_NoAutoNextDraft — M1·T2 contract gate.
//
// Verifies that after create → direct publish (the KEPT path to published:
// ADR 0082 phase (a) retired CreateTemplate's approval-config seeding, so the
// legacy submit → review → approve flow no longer runs on a fresh template):
//   - exactly ONE version row exists (the published v1)
//   - the template has published_version_id pointing to v1
//   - no v2 draft row was created automatically
//
// Then exercises the manual path CreateNextVersion and asserts v2 exists with status=draft.
func TestLifecycle_NoAutoNextDraft(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

	tenant := testdb.NewTenant(t, db)
	author := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID)).ID
	approver := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID)).ID

	// Grant system_admin so authz.Require passes for all template capabilities
	// (template.create, template.publish) without per-cap grants.
	testdb.SeedSystemAdmin(t, db, tenant.ID, author, "Integration Test Author")
	testdb.SeedSystemAdmin(t, db, tenant.ID, approver, "Integration Test Approver")

	// TxRunner chokepoint (F3.1) seeds GUC identity from ctx only — bind
	// tenant+actor per caller so authz.Require sees the right identity.
	authorCtx := platformtenant.WithActorID(
		platformtenant.WithTenantID(context.Background(), tenant.ID), author)
	approverCtx := platformtenant.WithActorID(
		platformtenant.WithTenantID(context.Background(), tenant.ID), approver)

	repo := infrastructure.New(db)
	svc := application.New(repo, noopPresigner{}, testClock{}, testUUID{}).
		WithRunner(platformdb.NewTxRunner(db))

	// ── Step 1: create template with v1 draft ──────────────────────────────
	createRes, err := svc.CreateTemplate(authorCtx, application.CreateTemplateCmd{
		TenantID:    tenant.ID,
		ActorUserID: author,
		DocTypeCode: "po",
		Key:         "tpl-no-auto-" + testdb.DeterministicID(t, "key")[:8],
		Name:        "No-Auto-Draft Template",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	templateID := createRes.Template.ID
	v1ID := createRes.Version.ID

	// Verify exactly 1 version row after create.
	if n := countVersionRows(t, db, templateID); n != 1 {
		t.Fatalf("after create: want 1 version row, got %d", n)
	}

	// ── Step 2: seed content hash (simulates autosave commit) ─────────────
	// The publish gate (T-004) requires a non-empty content_hash. Update directly
	// via SQL since the presigner is a no-op and there is no autosave service here.
	// content_hash must satisfy chk_template_version_content_hash (length 64) and
	// chk_template_version_content_hash_non_draft once the version leaves draft.
	const seedContentHash = "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
	testdb.SeedWithCaps(t, db, `[{"cap":"template.edit"}]`, func(tx *sql.Tx) error {
		// trg_template_version_tenant_consistent reads metaldocs.tenant_id (set by
		// authz.SeedTxIdentity in production). Assert it tx-locally for this raw write
		// so the trigger sees the parent template's tenant.
		if _, err := tx.ExecContext(ctx,
			`SELECT set_config('metaldocs.tenant_id', $1, true)`, tenant.ID,
		); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx,
			`UPDATE public.templates_template_version
			    SET content_hash = $2
			  WHERE id = $1::uuid`,
			v1ID, seedContentHash,
		)
		return err
	})

	// ── Step 3: direct publish ────────────────────────────────────────────
	// ADR 0082 phase (a) part 1: CreateTemplate no longer seeds an approval
	// config, so the legacy submit → review → approve flow cannot run against
	// a fresh template (SubmitForReview reads the retired config). The KEPT
	// path to published is PublishTemplateVersion on the draft: committed
	// content hash (T-004), publisher ≠ author (SoD via CheckSegregation),
	// and capability template.publish. The never-submitted version carries no
	// pending role binding (RoleBindingFor returns ""), so the capability
	// tier is the sole gate — no ActorRoles are passed.
	publishRes, err := svc.PublishTemplateVersion(approverCtx, application.PublishTemplateVersionCmd{
		TenantID:      tenant.ID,
		ActorUserID:   approver,
		TemplateID:    templateID,
		VersionNumber: 1,
	})
	if err != nil {
		t.Fatalf("PublishTemplateVersion: %v", err)
	}
	if publishRes.PublishedVersion.Status != domain.VersionStatusPublished {
		t.Fatalf("after publish: status=%q want=published", publishRes.PublishedVersion.Status)
	}

	// ── M1·T2 assertion: NO auto next-draft ───────────────────────────────
	// The result struct no longer carries a NextDraft field (dropped in M1·T2);
	// the absence of an auto-spawned row is proven directly against the DB below.

	// Exactly ONE version row: the published v1. No v2 was auto-created.
	if n := countVersionRows(t, db, templateID); n != 1 {
		t.Fatalf("M1·T2 violated: after approve-publish want 1 version row, got %d", n)
	}
	statuses := versionStatuses(t, db, templateID)
	if statuses[1] != "published" {
		t.Fatalf("v1 status = %q, want published", statuses[1])
	}
	if _, exists := statuses[2]; exists {
		t.Fatalf("M1·T2 violated: v2 must not exist after approve-publish (auto-spawn removed)")
	}

	// ── Step 5: manual CreateNextVersion → v2 draft ───────────────────────
	v2, err := svc.CreateNextVersion(authorCtx, application.CreateVersionCmd{
		TenantID:    tenant.ID,
		ActorUserID: author,
		TemplateID:  templateID,
	})
	if err != nil {
		t.Fatalf("CreateNextVersion: %v", err)
	}
	if v2.VersionNumber != 2 {
		t.Fatalf("v2.VersionNumber = %d, want 2", v2.VersionNumber)
	}
	if v2.Status != domain.VersionStatusDraft {
		t.Fatalf("v2.Status = %q, want draft", v2.Status)
	}

	// Now exactly TWO version rows exist: published v1 + draft v2.
	if n := countVersionRows(t, db, templateID); n != 2 {
		t.Fatalf("after CreateNextVersion want 2 version rows, got %d", n)
	}
	statuses = versionStatuses(t, db, templateID)
	if statuses[1] != "published" {
		t.Fatalf("v1 status = %q, want published", statuses[1])
	}
	if statuses[2] != "draft" {
		t.Fatalf("v2 status = %q, want draft", statuses[2])
	}
}
