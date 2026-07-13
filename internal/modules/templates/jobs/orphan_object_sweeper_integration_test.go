//go:build integration
// +build integration

package jobs

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/infrastructure"
	"metaldocs/internal/platform/objectstore"
	"metaldocs/tests/integration/testdb"
)

// fakeOrphanStore is an in-memory OrphanObjectStore. It records the keys it holds
// (with last-modified times) and which keys Delete was called for, so the test can
// assert exactly which objects the sweep removed without a live MinIO. It mimics
// the VerifiedStore tenant-prefix guard on ListTenantObjects so the cross-tenant
// isolation is exercised against the same contract the real store enforces.
type fakeOrphanStore struct {
	objects map[string]time.Time // key -> last modified
	deleted map[string]bool
}

func newFakeOrphanStore() *fakeOrphanStore {
	return &fakeOrphanStore{objects: map[string]time.Time{}, deleted: map[string]bool{}}
}

func (f *fakeOrphanStore) put(key string, lastModified time.Time) {
	f.objects[key] = lastModified
}

func (f *fakeOrphanStore) ListTenantObjects(_ context.Context, tenantID, prefix string) ([]objectstore.ObjectInfo, error) {
	if !objectstore.KeyHasTenantPrefix(tenantID, prefix) {
		return nil, objectstore.ErrKeyOutsideTenant
	}
	var out []objectstore.ObjectInfo
	for key, lm := range f.objects {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			out = append(out, objectstore.ObjectInfo{Key: key, LastModified: lm})
		}
	}
	return out, nil
}

func (f *fakeOrphanStore) Delete(_ context.Context, key string) error {
	f.deleted[key] = true
	delete(f.objects, key)
	return nil
}

// seedTemplateWithVersion inserts a templates_template + templates_template_version
// row pair under tenantID, asserting the real template.* tripwire capability
// tx-locally via the canonical testdb.SetCapsOnTx helper. docxKey is the
// version's stored docx_storage_key.
func seedTemplateWithVersion(t *testing.T, db *sql.DB, tenantID, templateID string, versionNumber int, docxKey string) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seed begin tx: %v", err)
	}
	defer tx.Rollback()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"template.create"},{"cap":"tenant.onboard"}]`)
	if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
		t.Fatalf("seed tenant GUC: %v", err)
	}
	// templates_template_version.tenant_id FKs to metaldocs.tenants(id).
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO metaldocs.tenants (id, name, slug)
		 VALUES ($1::uuid, 'Orphan Sweeper Tenant '||$1, 'orphan-'||$1)
		 ON CONFLICT (id) DO NOTHING`,
		tenantID,
	); err != nil {
		t.Fatalf("seed tenant: %v", err)
	}
	// chk_template_version_content_hash requires either '' or exactly 64 chars
	// (sha256 hex digest length).
	contentHash := strings.Repeat("0", 64)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO public.templates_template
		   (id, tenant_id, doc_type_code, key, name, description,
		    latest_version, created_by, system_owned)
		 VALUES ($1::uuid, $2::uuid, 'system', $3, 'Test Template', '',
		         $4, 'tester', false)`,
		templateID, tenantID, "tpl-key-"+templateID, versionNumber,
	); err != nil {
		t.Fatalf("seed template: %v", err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO public.templates_template_version
		   (id, tenant_id, template_id, version_number, status, docx_storage_key, content_hash,
		    metadata_schema, placeholder_schema, author_id, lock_version)
		 VALUES ($1::uuid, $6::uuid, $2::uuid, $3, 'draft', $4, $5, '{}'::jsonb, '[]'::jsonb, 'tester', 0)`,
		uuid.NewString(), templateID, versionNumber, docxKey, contentHash, tenantID,
	); err != nil {
		t.Fatalf("seed version: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("seed commit: %v", err)
	}
}

func TestTemplateOrphanSweeper_DeletesOnlyAgedUnreferenced(t *testing.T) {
	db, _ := testdb.Open(t)
	repo := infrastructure.New(db)

	tenantA := uuid.NewString()
	tenantB := uuid.NewString()
	templateID := uuid.NewString()
	const versionNumber = 1

	// Tenant A owns one template version. Its docx key is the canonical key the
	// application layer would build; its schema key is derived (never stored).
	docxKey := application.TemplateVersionDocxKey(tenantA, templateID, versionNumber)
	schemaKey := application.TemplateVersionSchemaKey(tenantA, templateID, versionNumber)
	seedTemplateWithVersion(t, db, tenantA, templateID, versionNumber, docxKey)

	// A separate tenant also owns a template, so TenantIDsWithTemplates returns
	// more than one tenant and the per-tenant prefix scoping is exercised.
	// crossTenant is tenant B's OWN referenced docx object (sweepOnce reconciles
	// EVERY tenant, so B's own sweep would delete an aged UNREFERENCED object
	// under B's prefix — the cross-tenant survival being asserted holds because
	// tenant A's sweep is prefix-scoped and never enumerates B's prefix, and
	// because the key is referenced under B's own sweep).
	tenantBTemplateID := uuid.NewString()
	crossTenant := application.TemplateVersionDocxKey(tenantB, tenantBTemplateID, 1)
	seedTemplateWithVersion(t, db, tenantB, tenantBTemplateID, 1, crossTenant)

	now := time.Now().UTC()
	old := now.Add(-48 * time.Hour)  // older than maxAge
	young := now.Add(-1 * time.Hour) // younger than maxAge

	store := newFakeOrphanStore()
	store.put(docxKey, old)   // referenced docx (even though old) -> survives
	store.put(schemaKey, old) // referenced derived schema (old)  -> survives
	oldOrphan := application.TenantTemplatePrefix(tenantA) + templateID + "/versions/99.docx"
	store.put(oldOrphan, old) // unreferenced + old               -> DELETED
	youngOrphan := application.TenantTemplatePrefix(tenantA) + templateID + "/versions/100.docx"
	store.put(youngOrphan, young) // unreferenced + young             -> survives
	store.put(crossTenant, old)   // tenant B's prefix                -> never touched by A's sweep

	const maxAge = 24 * time.Hour
	deleted, err := sweepOnce(context.Background(), repo, store, maxAge)
	if err != nil {
		t.Fatalf("sweepOnce: %v", err)
	}

	if !store.deleted[oldOrphan] {
		t.Errorf("(a) old unreferenced orphan was NOT deleted: %s", oldOrphan)
	}
	if store.deleted[docxKey] {
		t.Errorf("(b) referenced docx was deleted: %s", docxKey)
	}
	if store.deleted[schemaKey] {
		t.Errorf("(b) referenced DERIVED schema key was deleted: %s", schemaKey)
	}
	if store.deleted[youngOrphan] {
		t.Errorf("(c) young orphan (< maxAge) was deleted: %s", youngOrphan)
	}
	if store.deleted[crossTenant] {
		t.Errorf("cross-tenant object was deleted by another tenant's sweep: %s", crossTenant)
	}
	if deleted != 1 {
		t.Errorf("expected exactly 1 deletion (the old orphan under tenant A), got %d", deleted)
	}
}
