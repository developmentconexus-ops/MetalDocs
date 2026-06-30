//go:build integration
// +build integration

package jobs

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/repository"
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
// tx-locally (mirroring testdb.seedWithCaps, which is unexported). docxKey is the
// version's stored docx_storage_key.
func seedTemplateWithVersion(t *testing.T, db *sql.DB, tenantID, templateID string, versionNumber int, docxKey string) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seed begin tx: %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', $1, true)`,
		`[{"cap":"template.create"}]`,
	); err != nil {
		t.Fatalf("seed assert caps: %v", err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO public.templates_template
		   (id, tenant_id, doc_type_code, key, name, description,
		    visibility, latest_version, created_by, system_owned)
		 VALUES ($1::uuid, $2::uuid, 'system', $3, 'Test Template', '',
		         'internal', $4, 'tester', false)`,
		templateID, tenantID, "tpl-key-"+templateID, versionNumber,
	); err != nil {
		t.Fatalf("seed template: %v", err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO public.templates_template_version
		   (id, template_id, version_number, status, docx_storage_key, content_hash,
		    metadata_schema, placeholder_schema, author_id, pending_approver_role, lock_version)
		 VALUES ($1::uuid, $2::uuid, $3, 'draft', $4, 'hash', '{}'::jsonb, '[]'::jsonb, 'tester', 'approver', 0)`,
		uuid.NewString(), templateID, versionNumber, docxKey,
	); err != nil {
		t.Fatalf("seed version: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("seed commit: %v", err)
	}
}

func TestTemplateOrphanSweeper_DeletesOnlyAgedUnreferenced(t *testing.T) {
	db, _ := testdb.Open(t)
	repo := repository.New(db)

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
	seedTemplateWithVersion(t, db, tenantB, uuid.NewString(), 1,
		application.TenantTemplatePrefix(tenantB)+"other/versions/1.docx")

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
	crossTenant := application.TenantTemplatePrefix(tenantB) + "x/versions/1.docx"
	store.put(crossTenant, old) // tenant B's prefix                -> never touched by A's sweep

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
