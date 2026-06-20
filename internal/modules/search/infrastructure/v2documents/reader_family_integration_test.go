//go:build integration
// +build integration

package v2documents

import (
	"context"
	"sort"
	"testing"

	searchdomain "metaldocs/internal/modules/search/domain"
	taxonomyinfra "metaldocs/internal/modules/taxonomy/infrastructure"
	"metaldocs/tests/integration/testdb"
)

// TestListDocuments_FamilyProjection proves that each returned document's
// DocumentFamily equals its taxonomy-resolved FamilyCode.
func TestListDocuments_FamilyProjection(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID

	// Seed two taxonomies in distinct families.
	taxA := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))
	taxB := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))

	cdA := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithOwner(owner),
		testdb.WithTaxonomy(testdb.Taxonomy{
			TenantID:        taxA.TenantID,
			FamilyCode:      taxA.FamilyCode,
			ProcessAreaCode: taxA.ProcessAreaCode,
			ProfileCode:     taxA.ProfileCode,
		}),
	)
	cdB := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithOwner(owner),
		testdb.WithTaxonomy(testdb.Taxonomy{
			TenantID:        taxB.TenantID,
			FamilyCode:      taxB.FamilyCode,
			ProcessAreaCode: taxB.ProcessAreaCode,
			ProfileCode:     taxB.ProfileCode,
		}),
	)

	testdb.NewDocument(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithOwner(owner),
		testdb.WithControlledDoc(cdA),
	)
	testdb.NewDocument(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithOwner(owner),
		testdb.WithControlledDoc(cdB),
	)

	reader := NewReader(db, taxonomyinfra.NewFamilyCodeResolverRepository(db))
	docs, err := reader.ListDocuments(ctx, searchdomain.Query{TenantID: tenantID, ActorUserID: owner}, 50, 0)
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if len(docs) < 2 {
		t.Fatalf("expected at least 2 docs, got %d", len(docs))
	}

	// Build a profile→family map from what we seeded.
	wantFamily := map[string]string{
		taxA.ProfileCode: taxA.FamilyCode,
		taxB.ProfileCode: taxB.FamilyCode,
	}

	for _, d := range docs {
		want, ok := wantFamily[d.DocumentProfile]
		if !ok {
			// doc whose profile we didn't seed with a known family; skip.
			continue
		}
		if d.DocumentFamily != want {
			t.Errorf("doc %s: DocumentFamily=%q, want %q (profile=%q)", d.ID, d.DocumentFamily, want, d.DocumentProfile)
		}
	}
}

// TestListDocuments_FamilyFilter proves that querying with DocumentFamily returns
// only docs whose resolved family matches (case-insensitive), and an unmatched
// family returns empty.
func TestListDocuments_FamilyFilter(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID

	taxA := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))
	taxB := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))

	cdA := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithOwner(owner),
		testdb.WithTaxonomy(testdb.Taxonomy{
			TenantID:        taxA.TenantID,
			FamilyCode:      taxA.FamilyCode,
			ProcessAreaCode: taxA.ProcessAreaCode,
			ProfileCode:     taxA.ProfileCode,
		}),
	)
	cdB := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithOwner(owner),
		testdb.WithTaxonomy(testdb.Taxonomy{
			TenantID:        taxB.TenantID,
			FamilyCode:      taxB.FamilyCode,
			ProcessAreaCode: taxB.ProcessAreaCode,
			ProfileCode:     taxB.ProfileCode,
		}),
	)

	docA := testdb.NewDocument(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithOwner(owner),
		testdb.WithControlledDoc(cdA),
	)
	testdb.NewDocument(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithOwner(owner),
		testdb.WithControlledDoc(cdB),
	)

	reader := NewReader(db, taxonomyinfra.NewFamilyCodeResolverRepository(db))

	// Filter by family A using upper-case (case-insensitive).
	filterFamily := taxA.FamilyCode // e.g. "fam-abc12"
	// Pass upper-cased to exercise case-insensitivity.
	import_upper := toUpper(filterFamily)
	docsA, err := reader.ListDocuments(ctx, searchdomain.Query{
		TenantID:       tenantID,
		ActorUserID:    owner,
		DocumentFamily: import_upper,
	}, 50, 0)
	if err != nil {
		t.Fatalf("ListDocuments(familyA): %v", err)
	}

	// Only docs in family A should be returned.
	for _, d := range docsA {
		if d.ID == docA.ID {
			continue
		}
		// Any other doc returned must also resolve to family A (not B).
		if d.DocumentFamily != filterFamily {
			t.Errorf("family filter leak: doc %s has family %q, want %q", d.ID, d.DocumentFamily, filterFamily)
		}
	}

	// Exactly one doc (cdA / docA) should match.
	found := false
	for _, d := range docsA {
		if d.ID == docA.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("family filter: expected docA %s in results, got %v", docA.ID, docIDs(docsA))
	}

	// A non-existent family → empty.
	none, err := reader.ListDocuments(ctx, searchdomain.Query{
		TenantID:       tenantID,
		ActorUserID:    owner,
		DocumentFamily: "fam-does-not-exist-xyz",
	}, 50, 0)
	if err != nil {
		t.Fatalf("ListDocuments(nofamily): %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("non-existent family must return 0 docs, got %d", len(none))
	}
}

// TestListDocuments_FamilyFilterPagination proves the family filter participates
// in SQL LIMIT/OFFSET — not a post-Go filter — by verifying paginated windows
// across ≥3 family-A docs cover the full set with no overlap.
func TestListDocuments_FamilyFilterPagination(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID

	// One taxonomy/family for 3 docs in family A, one doc in family B.
	taxA := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))
	taxB := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))

	cdA1 := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID), testdb.WithOwner(owner),
		testdb.WithTaxonomy(testdb.Taxonomy{
			TenantID:        taxA.TenantID,
			FamilyCode:      taxA.FamilyCode,
			ProcessAreaCode: taxA.ProcessAreaCode,
			ProfileCode:     taxA.ProfileCode,
		}),
	)
	cdA2 := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID), testdb.WithOwner(owner),
		testdb.WithTaxonomy(testdb.Taxonomy{
			TenantID:        taxA.TenantID,
			FamilyCode:      taxA.FamilyCode,
			ProcessAreaCode: taxA.ProcessAreaCode,
			ProfileCode:     taxA.ProfileCode,
		}),
	)
	cdA3 := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID), testdb.WithOwner(owner),
		testdb.WithTaxonomy(testdb.Taxonomy{
			TenantID:        taxA.TenantID,
			FamilyCode:      taxA.FamilyCode,
			ProcessAreaCode: taxA.ProcessAreaCode,
			ProfileCode:     taxA.ProfileCode,
		}),
	)
	cdB := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID), testdb.WithOwner(owner),
		testdb.WithTaxonomy(testdb.Taxonomy{
			TenantID:        taxB.TenantID,
			FamilyCode:      taxB.FamilyCode,
			ProcessAreaCode: taxB.ProcessAreaCode,
			ProfileCode:     taxB.ProfileCode,
		}),
	)

	docA1 := testdb.NewDocument(t, db, testdb.WithTenant(tenantID), testdb.WithOwner(owner), testdb.WithControlledDoc(cdA1))
	docA2 := testdb.NewDocument(t, db, testdb.WithTenant(tenantID), testdb.WithOwner(owner), testdb.WithControlledDoc(cdA2))
	docA3 := testdb.NewDocument(t, db, testdb.WithTenant(tenantID), testdb.WithOwner(owner), testdb.WithControlledDoc(cdA3))
	testdb.NewDocument(t, db, testdb.WithTenant(tenantID), testdb.WithOwner(owner), testdb.WithControlledDoc(cdB))

	allFamilyA := map[string]bool{
		docA1.ID: true,
		docA2.ID: true,
		docA3.ID: true,
	}

	reader := NewReader(db, taxonomyinfra.NewFamilyCodeResolverRepository(db))

	// Page 1: limit=2 offset=0.
	page1, err := reader.ListDocuments(ctx, searchdomain.Query{
		TenantID:       tenantID,
		ActorUserID:    owner,
		DocumentFamily: taxA.FamilyCode,
	}, 2, 0)
	if err != nil {
		t.Fatalf("page1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("page1: expected 2 docs, got %d: %v", len(page1), docIDs(page1))
	}

	// Page 2: limit=2 offset=2.
	page2, err := reader.ListDocuments(ctx, searchdomain.Query{
		TenantID:       tenantID,
		ActorUserID:    owner,
		DocumentFamily: taxA.FamilyCode,
	}, 2, 2)
	if err != nil {
		t.Fatalf("page2: %v", err)
	}
	if len(page2) != 1 {
		t.Fatalf("page2: expected 1 doc, got %d: %v", len(page2), docIDs(page2))
	}

	// No overlap between pages.
	p1IDs := make(map[string]bool, len(page1))
	for _, d := range page1 {
		p1IDs[d.ID] = true
	}
	for _, d := range page2 {
		if p1IDs[d.ID] {
			t.Fatalf("overlap: doc %s appears in both pages", d.ID)
		}
	}

	// Union of both pages == all 3 family-A docs.
	union := make(map[string]bool)
	for _, d := range append(page1, page2...) {
		union[d.ID] = true
	}
	for id := range allFamilyA {
		if !union[id] {
			t.Fatalf("family-A doc %s missing from union of pages", id)
		}
	}
	if len(union) != 3 {
		// Extra docs appeared (family B leaked in).
		got := make([]string, 0, len(union))
		for id := range union {
			got = append(got, id)
		}
		sort.Strings(got)
		t.Fatalf("union has %d docs (expected 3): %v", len(union), got)
	}
}

// toUpper is a local helper to upper-case a family code for case-insensitivity tests.
func toUpper(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			result[i] = c - 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func docIDs(docs []searchdomain.Document) []string {
	ids := make([]string, len(docs))
	for i, d := range docs {
		ids[i] = d.ID
	}
	return ids
}
