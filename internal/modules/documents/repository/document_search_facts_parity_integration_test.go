//go:build integration
// +build integration

package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"metaldocs/tests/integration/testdb"
)

// F4.2 documents search projection contract parity gate (M4 Category C4a; ADR-0039
// D3a). documents publishes metaldocs.v_document_search_facts — a 1:1 column
// projection of public.documents (the 14 columns the search consumer
// search/infrastructure/v2documents/reader.go reads). The view is a PURE passthrough:
// no WHERE, no COALESCE — archived_at is exposed as a column and search keeps its own
// `archived_at IS NULL` filter (F4.3). This test proves ZERO drift: every view row ==
// the base-table row for the same (tenant_id, id) across the discriminator docs, the
// row counts match, and archived rows are PRESENT (no hidden filter).
//
// D6: GREEN proves view == raw base before F4.3 repoints `FROM public.documents` to
// `FROM metaldocs.v_document_search_facts`. The archived + NULL-snapshot docs are the
// anti-drift discriminators (catch a sneaky filter or COALESCE baked into the view).

// dsfDoc is one seeded discriminator document.
type dsfDoc struct {
	id        string
	cdID      sql.NullString
	profile   sql.NullString
	area      sql.NullString
	code      sql.NullString
	archived  bool
	effective bool
}

// insertSearchDoc inserts one public.documents row with only the columns the contract
// projects (plus the NOT-NULL columns the table requires) and returns its id.
func insertSearchDoc(t *testing.T, db *sql.DB, tenantID string, d dsfDoc) string {
	t.Helper()
	const q = `
INSERT INTO public.documents
  (id, tenant_id, template_version_id, name, status, form_data_json, created_by,
   controlled_document_id, profile_code_snapshot, process_area_code_snapshot, code,
   archived_at, effective_from, effective_to, revision_number)
VALUES
  (gen_random_uuid(), $1, gen_random_uuid(), $2, 'draft', '{}'::jsonb, $3,
   $4::uuid, $5, $6, $7,
   CASE WHEN $8 THEN now() ELSE NULL END,
   CASE WHEN $9 THEN now() ELSE NULL END,
   CASE WHEN $9 THEN now() ELSE NULL END,
   3)
RETURNING id::text`
	var id string
	if err := db.QueryRowContext(context.Background(), q,
		tenantID, "doc-"+t.Name(), tenantID, // created_by reuses tenantID string (a text col; FK-free)
		d.cdID, d.profile, d.area, d.code, d.archived, d.effective,
	).Scan(&id); err != nil {
		t.Fatalf("insertSearchDoc: %v", err)
	}
	return id
}

// dsfRow holds the 14 projected columns, scanned NULL-safely for view==base comparison.
type dsfRow struct {
	tenantID, id                       string
	cdID, name, status                 sql.NullString
	profile, area, createdBy, code     sql.NullString
	revision                           sql.NullInt64
	effectiveFrom, effectiveTo, archAt sql.NullTime
	createdAt                          sql.NullTime
}

func scanDSF(t *testing.T, db *sql.DB, table, tenantID, id string) dsfRow {
	t.Helper()
	q := `SELECT tenant_id::text, id::text, controlled_document_id::text, name, status,
	             profile_code_snapshot, process_area_code_snapshot, created_by, code,
	             revision_number, effective_from, effective_to, created_at, archived_at
	        FROM ` + table + ` WHERE tenant_id = $1 AND id = $2::uuid`
	var r dsfRow
	if err := db.QueryRowContext(context.Background(), q, tenantID, id).Scan(
		&r.tenantID, &r.id, &r.cdID, &r.name, &r.status,
		&r.profile, &r.area, &r.createdBy, &r.code,
		&r.revision, &r.effectiveFrom, &r.effectiveTo, &r.createdAt, &r.archAt,
	); err != nil {
		t.Fatalf("scanDSF(%s): %v", table, err)
	}
	return r
}

func TestDocumentSearchFacts_ParityWithBaseTable(t *testing.T) {
	db, _ := testdb.Open(t)
	// Pin to one physical conn so the session-level capability assertion persists
	// across the raw inserts (public.documents has a capability-assertion trigger).
	db.SetMaxOpenConns(1)
	tnt := testdb.NewTenant(t, db)
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	owner := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tnt.ID), testdb.WithTaxonomy(tax), testdb.WithOwner(owner.ID))

	// Assert document.create for the raw inserts below (mirrors the production authz layer).
	testdb.SetCapsOnDB(t, db, `[{"cap":"document.create"}]`)

	ns := func() sql.NullString { return sql.NullString{} }
	str := func(s string) sql.NullString { return sql.NullString{String: s, Valid: true} }

	docs := map[string]string{} // label -> id
	docs["standalone"] = insertSearchDoc(t, db, tnt.ID, dsfDoc{
		cdID: ns(), profile: str("PROF-A"), area: str("AREA-A"), code: str("STD-1"), effective: true})
	docs["cd-linked"] = insertSearchDoc(t, db, tnt.ID, dsfDoc{
		cdID: str(cd.ID), profile: str("PROF-B"), area: str("AREA-B"), code: str("CDL-1"), effective: true})
	docs["archived"] = insertSearchDoc(t, db, tnt.ID, dsfDoc{
		cdID: ns(), profile: str("PROF-C"), area: str("AREA-C"), code: str("ARC-1"), archived: true})
	docs["null-snapshot"] = insertSearchDoc(t, db, tnt.ID, dsfDoc{
		cdID: ns(), profile: ns(), area: ns(), code: ns()})

	// Per-doc NULL-safe column equality: view row == base row across all 14 columns.
	for label, id := range docs {
		v := scanDSF(t, db, "metaldocs.v_document_search_facts", tnt.ID, id)
		b := scanDSF(t, db, "public.documents", tnt.ID, id)
		if v != b {
			t.Fatalf("projection drift (%s):\n view=%+v\n base=%+v", label, v, b)
		}
	}

	// Row counts equal — the view is 1:1 with the base table for this tenant
	// (no hidden filter drops or duplicates rows).
	var vCount, bCount int
	if err := db.QueryRowContext(context.Background(),
		`SELECT count(*) FROM metaldocs.v_document_search_facts WHERE tenant_id=$1`, tnt.ID).Scan(&vCount); err != nil {
		t.Fatalf("view count: %v", err)
	}
	if err := db.QueryRowContext(context.Background(),
		`SELECT count(*) FROM public.documents WHERE tenant_id=$1`, tnt.ID).Scan(&bCount); err != nil {
		t.Fatalf("base count: %v", err)
	}
	if vCount != bCount {
		t.Fatalf("row count drift: view=%d base=%d (passthrough must be 1:1)", vCount, bCount)
	}

	// Discriminator: the archived doc MUST be present in the view (passthrough, not
	// pre-filtered — search applies `archived_at IS NULL` itself in F4.3).
	var archInView bool
	if err := db.QueryRowContext(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM metaldocs.v_document_search_facts
		   WHERE tenant_id=$1 AND id=$2::uuid AND archived_at IS NOT NULL)`,
		tnt.ID, docs["archived"]).Scan(&archInView); err != nil {
		t.Fatalf("archived presence: %v", err)
	}
	if !archInView {
		t.Fatal("archived doc must be PRESENT in v_document_search_facts (view is passthrough, not active-only)")
	}
}
