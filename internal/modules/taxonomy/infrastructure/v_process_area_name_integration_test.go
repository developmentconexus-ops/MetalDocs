//go:build integration
// +build integration

package infrastructure

import (
	"context"
	"database/sql"
	"sort"
	"testing"

	"github.com/google/uuid"

	"metaldocs/tests/integration/testdb"
)

// F2.1b — taxonomy publishes metaldocs.v_process_area_name.
//
// Shape:  (tenant_id uuid NOT NULL, area_code text NOT NULL, area_name text NOT NULL)
// Rule:   one row per (tenant_id, area_code); 1:1 projection of
//         metaldocs.document_process_areas (code → area_code, name → area_name).
// No filter on is_active / archived_at — labels must resolve for archived areas too
// (parity with the existing AreaCatalogReader port at
// area_catalog_reader.go:32, which does not filter either).

type areaNameRow struct {
	TenantID string
	AreaCode string
	AreaName string
}

func seedExtraProcessArea(t *testing.T, db *sql.DB, tenantID string) (code, name string) {
	t.Helper()
	code = "f21b-" + uuid.NewString()[:8]
	name = "F2.1b Area " + code
	testdb.SeedWithCaps(t, db, `[{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO metaldocs.document_process_areas (tenant_id, code, name)
			 VALUES ($1::uuid, $2, $3)`,
			tenantID, code, name,
		)
		return err
	})
	return code, name
}

func areaNameRowsForTenant(t *testing.T, db *sql.DB, tenantID string) []areaNameRow {
	t.Helper()
	rows, err := db.QueryContext(context.Background(),
		`SELECT tenant_id::text, area_code, area_name
		   FROM metaldocs.v_process_area_name
		  WHERE tenant_id = $1::uuid
		  ORDER BY area_code`,
		tenantID)
	if err != nil {
		t.Fatalf("v_process_area_name query: %v", err)
	}
	defer rows.Close()
	var out []areaNameRow
	for rows.Next() {
		var r areaNameRow
		if err := rows.Scan(&r.TenantID, &r.AreaCode, &r.AreaName); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	return out
}

func TestProcessAreaName_OneToOneProjection_PerTenant(t *testing.T) {
	db, _ := testdb.Open(t)
	tax := testdb.NewTaxonomy(t, db) // seeds one area: code = tax.ProcessAreaCode, name = code
	c2, n2 := seedExtraProcessArea(t, db, tax.TenantID)
	c3, n3 := seedExtraProcessArea(t, db, tax.TenantID)

	got := areaNameRowsForTenant(t, db, tax.TenantID)

	want := map[string]string{
		tax.ProcessAreaCode: tax.ProcessAreaCode, // factory sets name = code
		c2:                  n2,
		c3:                  n3,
	}
	if len(got) != len(want) {
		t.Fatalf("rowcount: got %d want %d (rows=%+v)", len(got), len(want), got)
	}
	for _, r := range got {
		w, ok := want[r.AreaCode]
		if !ok {
			t.Fatalf("unexpected area_code in view: %+v", r)
		}
		if r.AreaName != w {
			t.Fatalf("area_name drift for code %q: got %q want %q", r.AreaCode, r.AreaName, w)
		}
		if r.TenantID != tax.TenantID {
			t.Fatalf("tenant_id drift: got %q want %q", r.TenantID, tax.TenantID)
		}
	}
}

func TestProcessAreaName_CrossTenantIsolation(t *testing.T) {
	db, _ := testdb.Open(t)
	taxA := testdb.NewTaxonomy(t, db)
	taxB := testdb.NewTaxonomy(t, db)

	gotA := areaNameRowsForTenant(t, db, taxA.TenantID)
	for _, r := range gotA {
		if r.AreaCode == taxB.ProcessAreaCode {
			t.Fatalf("tenant A view leaked tenant B's area_code %q: %+v", taxB.ProcessAreaCode, r)
		}
	}
}

func TestProcessAreaName_ViewShape(t *testing.T) {
	db, _ := testdb.Open(t)
	rows, err := db.QueryContext(context.Background(), `
	  SELECT column_name, data_type, is_nullable
	    FROM information_schema.columns
	   WHERE table_schema = 'metaldocs'
	     AND table_name   = 'v_process_area_name'
	   ORDER BY ordinal_position`)
	if err != nil {
		t.Fatalf("shape query: %v", err)
	}
	defer rows.Close()
	type col struct{ name, typ, nullable string }
	var got []col
	for rows.Next() {
		var c col
		if err := rows.Scan(&c.name, &c.typ, &c.nullable); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, c)
	}
	want := []col{
		{"tenant_id", "uuid", "NO"},
		{"area_code", "text", "NO"},
		{"area_name", "text", "NO"},
	}
	if len(got) != len(want) {
		t.Fatalf("column count: got %d want %d (got=%+v)", len(got), len(want), got)
	}
	sortCols := func(s []col) { sort.SliceStable(s, func(i, j int) bool { return s[i].name < s[j].name }) }
	sortCols(got)
	sortCols(want)
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("column %d drift: got %+v want %+v", i, got[i], want[i])
		}
	}
}
