//go:build integration
// +build integration

package application_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/application"
	docrepo "metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/platform/docgenv2"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	templatesdomain "metaldocs/internal/modules/templates/domain"
	"metaldocs/tests/integration/testdb"
)

// fakeDictLookup is a controllable DictionaryValueReader. It stores a per-(tenantID,name)
// map; a missing key returns (found=false), satisfying D7 and isolation tests.
type fakeDictLookup struct {
	entries map[string]string // key: tenantID+":"+name
}

func newFakeDictLookup() *fakeDictLookup { return &fakeDictLookup{entries: map[string]string{}} }
func (f *fakeDictLookup) set(tenantID, name, value string) {
	f.entries[tenantID+":"+name] = value
}
func (f *fakeDictLookup) Lookup(_ context.Context, tenantID, name string) (string, bool, error) {
	v, ok := f.entries[tenantID+":"+name]
	return v, ok, nil
}

// newDictionarySvcForIntegration builds a real documents Service with the given dictReader.
// Mirrors the wiring in create_document_snapshot_integration_test.go.
func newDictionarySvcForIntegration(db *sql.DB, dictReader application.DictionaryValueReader, publishedTemplateVersionID string) *application.Service {
	snapshotSvc := application.NewSnapshotService(docgenv2.NewTemplatesSnapshotReader(db))
	svc := application.NewServiceWithSnapshot(
		docrepo.New(db, iampg.NewUserDisplayNameRepository(db), controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{}),
		nil,
		docgenv2.NewTemplatesTemplateReader(db),
		fakeFormVal{valid: true},
		&noopAudit{},
		&fakeProfileDefaultTemplateReader{id: strptr(publishedTemplateVersionID), status: strptr("published")},
		snapshotSvc,
	)
	svc.WithDictionaryReader(dictReader)
	return svc
}

// seedDictTemplate seeds a template + version with a single PHDictionary placeholder.
// Returns the templateVersionID.
func seedDictTemplate(t *testing.T, db *sql.DB, tenantID, actorID, templateKey, phID, phName string) string {
	t.Helper()
	ctx := context.Background()
	templateID := testdb.DeterministicID(t, "dtpl-"+templateKey)
	templateVersionID := testdb.DeterministicID(t, "dtplv-"+templateKey)

	schema, err := json.Marshal([]templatesdomain.Placeholder{
		{ID: phID, Type: templatesdomain.PHDictionary, Name: phName, Label: phName},
	})
	if err != nil {
		t.Fatalf("marshal placeholder schema: %v", err)
	}

	// templates_template_version carries trg_template_version_tenant_consistent,
	// which requires metaldocs.tenant_id set tx-locally in addition to the
	// template.create tripwire cap — so this seeds tenant identity
	// (authz.SeedTxTenant) alongside caps (testdb.SetCapsOnTx) on a
	// self-managed tx, per the sanctioned direct-tx pattern.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin template seed tx: %v", err)
	}
	defer tx.Rollback()
	if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
		t.Fatalf("seed template tx tenant: %v", err)
	}
	testdb.SetCapsOnTx(t, tx, `[{"cap":"template.create"}]`)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO templates_template (
			id, tenant_id, doc_type_code, key, name, latest_version, published_version_id, created_by
		) VALUES ($1::uuid, $2, 'po', $3, $3, 1, NULL, $4)`,
		templateID, tenantID, templateKey, actorID,
	); err != nil {
		t.Fatalf("seed templates_template: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO templates_template_version (
			id, tenant_id, template_id, version_number, status, docx_storage_key, content_hash,
			metadata_schema, placeholder_schema, author_id, published_at
		) VALUES ($1::uuid, $5::uuid, $2::uuid, 1, 'published', $7, $6,
			'{}'::jsonb, $3::jsonb, $4, now())`,
		templateVersionID, templateID, string(schema), actorID, tenantID, strings.Repeat("a", 64), "templates/dict-test/"+templateKey+"/body.docx",
	); err != nil {
		t.Fatalf("seed templates_template_version: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE templates_template SET published_version_id = $1::uuid WHERE id = $2::uuid`,
		templateVersionID, templateID,
	); err != nil {
		t.Fatalf("update templates_template published_version_id: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit template seed tx: %v", err)
	}
	return templateVersionID
}

// queryPlaceholderRow reads (source, value_text) from document_placeholder_values directly.
// Runs as superuser so RLS does not filter.
func queryPlaceholderRow(t *testing.T, db *sql.DB, tenantID, revisionID, placeholderID string) (source string, valueText sql.NullString) {
	t.Helper()
	err := db.QueryRowContext(context.Background(), `
		SELECT source, value_text
		  FROM public.document_placeholder_values
		 WHERE tenant_id = $1::uuid AND revision_id = $2::uuid AND placeholder_id = $3
	`, tenantID, revisionID, placeholderID).Scan(&source, &valueText)
	if err != nil {
		t.Fatalf("queryPlaceholderRow(tenant=%s rev=%s ph=%s): %v", tenantID, revisionID, placeholderID, err)
	}
	return source, valueText
}

// openDictDB opens a test DB with the search_path and pool config that the
// snapshot integration test uses.
func openDictDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	db, dbName := testdb.Open(t)
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `ALTER DATABASE "`+dbName+`" SET search_path TO public, metaldocs`); err != nil {
		t.Fatalf("alter search_path: %v", err)
	}
	db.SetMaxIdleConns(0)
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(4)
	return db, dbName
}

// TestSP2_DictionaryPin_CreatePinsValue: D1 pin-at-creation + D8 per-revision immunity.
// Template has a PHDictionary placeholder company_name. Tenant dictionary entry = "ACME".
// After create: placeholder row has source='dictionary', value_text='ACME'.
// After updating the fake dictionary to "ACME-2": the existing row is unchanged (D8).
func TestSP2_DictionaryPin_CreatePinsValue(t *testing.T) {
	ctx := context.Background()
	db, _ := openDictDB(t)

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	actorID := testdb.DeterministicID(t, "actor-d1-pin")
	testdb.SeedSystemAdmin(t, db, tenantID, actorID, "D1 Author")
	testdb.SeedGovernedTaxonomy(t, db, tenantID, "po", "quality")

	cdFixture := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithTaxonomy(testdb.Taxonomy{TenantID: tenantID, ProfileCode: "po", ProcessAreaCode: "quality"}),
		testdb.WithCode("PO-D1-001"),
		testdb.WithOwner(actorID),
	)

	const phID = "ph-company"
	const phName = "company_name"
	tplVersionID := seedDictTemplate(t, db, tenantID, actorID, "d1-pin-test", phID, phName)

	dictReader := newFakeDictLookup()
	dictReader.set(tenantID, phName, "ACME")

	svc := newDictionarySvcForIntegration(db, dictReader, tplVersionID)
	initializer := application.NewCDDocumentInitializer(svc)

	cd := &controlleddocumentsdomain.ControlledDocument{
		ID:              cdFixture.ID,
		TenantID:        tenantID,
		ProfileCode:     cdFixture.ProfileCode,
		ProcessAreaCode: cdFixture.ProcessAreaCode,
		Code:            cdFixture.Code,
		OwnerUserID:     actorID,
		Status:          controlleddocumentsdomain.CDStatusActive,
	}

	// OFF-TX resolve (H-PRE-1).
	dv, err := initializer.ResolveDictionaryValues(ctx, tenantID, tplVersionID)
	if err != nil {
		t.Fatalf("ResolveDictionaryValues: %v", err)
	}
	if dv[phID] != "ACME" {
		t.Fatalf("want dv[%q]=ACME, got %q", phID, dv[phID])
	}

	// In-tx clone with pinned values.
	req, err := controlleddocumentsdomain.NewCloneTemplateRequest(&tplVersionID, "D1 Pin Test", nil)
	if err != nil {
		t.Fatalf("NewCloneTemplateRequest: %v", err)
	}
	req = req.WithDictionaryValues(dv)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()
	ref, err := initializer.CloneTemplate(ctx, tx, cd, req)
	if err != nil {
		t.Fatalf("CloneTemplate: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// D1: assert pin row has source='dictionary' and value_text='ACME'.
	source, vt := queryPlaceholderRow(t, db, tenantID, ref.RevisionID, phID)
	if source != "dictionary" {
		t.Fatalf("D1: want source='dictionary', got %q", source)
	}
	if !vt.Valid || vt.String != "ACME" {
		t.Fatalf("D1: want value_text='ACME', got %+v", vt)
	}

	// D8: mutate the dictionary; re-read the same revision row — must still be 'ACME'.
	dictReader.set(tenantID, phName, "ACME-2")
	sourceAfter, vtAfter := queryPlaceholderRow(t, db, tenantID, ref.RevisionID, phID)
	if sourceAfter != "dictionary" || !vtAfter.Valid || vtAfter.String != "ACME" {
		t.Fatalf("D8: existing revision must stay pinned at ACME; got source=%q value_text=%+v", sourceAfter, vtAfter)
	}
}

// TestSP2_DictionaryPin_RevisionReResolves: D10 — CreateRevision re-resolves the
// dictionary so the new revision captures the CURRENT value while the old revision
// retains its pinned value from creation time.
func TestSP2_DictionaryPin_RevisionReResolves(t *testing.T) {
	ctx := context.Background()
	db, _ := openDictDB(t)

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	actorID := testdb.DeterministicID(t, "actor-d10")
	testdb.SeedSystemAdmin(t, db, tenantID, actorID, "D10 Author")
	testdb.SeedGovernedTaxonomy(t, db, tenantID, "po", "quality")

	cdFixture := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithTaxonomy(testdb.Taxonomy{TenantID: tenantID, ProfileCode: "po", ProcessAreaCode: "quality"}),
		testdb.WithCode("PO-D10-001"),
		testdb.WithOwner(actorID),
	)

	const phID = "ph-corp"
	const phName = "company_name"
	tplVersionID := seedDictTemplate(t, db, tenantID, actorID, "d10-rerev-test", phID, phName)

	dictReader := newFakeDictLookup()
	dictReader.set(tenantID, phName, "V1_Corp")

	svc := newDictionarySvcForIntegration(db, dictReader, tplVersionID)
	initializer := application.NewCDDocumentInitializer(svc)

	cd := &controlleddocumentsdomain.ControlledDocument{
		ID:              cdFixture.ID,
		TenantID:        tenantID,
		ProfileCode:     cdFixture.ProfileCode,
		ProcessAreaCode: cdFixture.ProcessAreaCode,
		Code:            cdFixture.Code,
		OwnerUserID:     actorID,
		Status:          controlleddocumentsdomain.CDStatusActive,
	}

	// V1 create.
	dv1, err := initializer.ResolveDictionaryValues(ctx, tenantID, tplVersionID)
	if err != nil {
		t.Fatalf("ResolveDictionaryValues V1: %v", err)
	}
	req1, _ := controlleddocumentsdomain.NewCloneTemplateRequest(&tplVersionID, "D10 Rev1", nil)
	req1 = req1.WithDictionaryValues(dv1)
	tx1, _ := db.BeginTx(ctx, nil)
	ref1, err := initializer.CloneTemplate(ctx, tx1, cd, req1)
	if err != nil {
		tx1.Rollback()
		t.Fatalf("CloneTemplate V1: %v", err)
	}
	if err := tx1.Commit(); err != nil {
		t.Fatalf("commit V1: %v", err)
	}

	// Update dictionary to V2.
	dictReader.set(tenantID, phName, "V2_Corp")

	// ux_documents_cd_active allows only one active document per controlled
	// document; V1's clone left an active document under cd, so it must be
	// superseded before V2's clone creates the next active document.
	testdb.SupersedeActiveDocumentForCD(t, db, cdFixture.ID)

	// V2 revision: re-resolve then clone.
	dv2, err := initializer.ResolveDictionaryValues(ctx, tenantID, tplVersionID)
	if err != nil {
		t.Fatalf("ResolveDictionaryValues V2: %v", err)
	}
	req2, _ := controlleddocumentsdomain.NewCloneTemplateRequest(&tplVersionID, "D10 Rev2", nil)
	req2 = req2.WithDictionaryValues(dv2)
	tx2, _ := db.BeginTx(ctx, nil)
	ref2, err := initializer.CloneTemplate(ctx, tx2, cd, req2)
	if err != nil {
		tx2.Rollback()
		t.Fatalf("CloneTemplate V2: %v", err)
	}
	if err := tx2.Commit(); err != nil {
		t.Fatalf("commit V2: %v", err)
	}

	// D10: new revision captures current value.
	_, vt2 := queryPlaceholderRow(t, db, tenantID, ref2.RevisionID, phID)
	if !vt2.Valid || vt2.String != "V2_Corp" {
		t.Fatalf("D10: new revision must have V2_Corp, got %+v", vt2)
	}

	// D8 / D10 immunity: old revision retains V1_Corp.
	_, vt1 := queryPlaceholderRow(t, db, tenantID, ref1.RevisionID, phID)
	if !vt1.Valid || vt1.String != "V1_Corp" {
		t.Fatalf("D10/D8: old revision must still have V1_Corp, got %+v", vt1)
	}
}

// TestSP2_DictionaryPin_MissingTokenD7: a template references a token with no
// dictionary entry → ResolveDictionaryValues returns controlleddocuments ErrDictionaryTokenMissing
// (the CD sentinel, which the HTTP layer maps to 422).
func TestSP2_DictionaryPin_MissingTokenD7(t *testing.T) {
	ctx := context.Background()
	db, _ := openDictDB(t)

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	actorID := testdb.DeterministicID(t, "actor-d7")
	testdb.SeedSystemAdmin(t, db, tenantID, actorID, "D7 Author")
	testdb.SeedGovernedTaxonomy(t, db, tenantID, "po", "quality")

	const phID = "ph-unknown"
	const phName = "unknown_token"
	tplVersionID := seedDictTemplate(t, db, tenantID, actorID, "d7-missing-test", phID, phName)

	// DictReader: no entry for unknown_token in any tenant.
	svc := newDictionarySvcForIntegration(db, newFakeDictLookup(), tplVersionID)
	initializer := application.NewCDDocumentInitializer(svc)

	_, err := initializer.ResolveDictionaryValues(ctx, tenantID, tplVersionID)
	if !errors.Is(err, controlleddocumentsdomain.ErrDictionaryTokenMissing) {
		t.Fatalf("D7: want ErrDictionaryTokenMissing, got %v", err)
	}
}

// TestSP2_DictionaryPin_TwoTenantIsolation: tenant A's dictionary entry must not
// satisfy tenant B's resolve (same token name, different tenant ID).
func TestSP2_DictionaryPin_TwoTenantIsolation(t *testing.T) {
	ctx := context.Background()
	db, _ := openDictDB(t)

	tenantA := testdb.NewTenant(t, db)
	tenantB := testdb.NewTenant(t, db)
	actorA := testdb.DeterministicID(t, "actor-iso-a")
	actorB := testdb.DeterministicID(t, "actor-iso-b")
	testdb.SeedSystemAdmin(t, db, tenantA.ID, actorA, "Iso Author A")
	testdb.SeedSystemAdmin(t, db, tenantB.ID, actorB, "Iso Author B")
	// document_profiles.code is the table's sole PRIMARY KEY (globally unique
	// across tenants, not (tenant_id, code) as SeedGovernedTaxonomy's ON
	// CONFLICT target assumes) — reusing the same profile code "po" for both
	// tenants collides on document_profiles_pkey, so each tenant gets its own
	// code here.
	testdb.SeedGovernedTaxonomy(t, db, tenantA.ID, "po-iso-a", "quality")
	testdb.SeedGovernedTaxonomy(t, db, tenantB.ID, "po-iso-b", "quality")

	const phID = "ph-iso"
	const phName = "company_name"
	tplVersionA := seedDictTemplate(t, db, tenantA.ID, actorA, "iso-test-a", phID, phName)
	tplVersionB := seedDictTemplate(t, db, tenantB.ID, actorB, "iso-test-b", phID, phName)

	// Only tenant A has the entry.
	dictReader := newFakeDictLookup()
	dictReader.set(tenantA.ID, phName, "TenantA_Corp")

	// Tenant A resolve succeeds.
	svcA := newDictionarySvcForIntegration(db, dictReader, tplVersionA)
	initA := application.NewCDDocumentInitializer(svcA)
	dv, err := initA.ResolveDictionaryValues(ctx, tenantA.ID, tplVersionA)
	if err != nil {
		t.Fatalf("tenant A resolve: %v", err)
	}
	if dv[phID] != "TenantA_Corp" {
		t.Fatalf("tenant A: expected TenantA_Corp, got %q", dv[phID])
	}

	// Tenant B resolve fails — A's entry must not bleed into B.
	svcB := newDictionarySvcForIntegration(db, dictReader, tplVersionB)
	initB := application.NewCDDocumentInitializer(svcB)
	_, err = initB.ResolveDictionaryValues(ctx, tenantB.ID, tplVersionB)
	if !errors.Is(err, controlleddocumentsdomain.ErrDictionaryTokenMissing) {
		t.Fatalf("two-tenant isolation: tenant B must return ErrDictionaryTokenMissing, got %v", err)
	}
}
