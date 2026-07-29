//go:build integration
// +build integration

package infrastructure

// Real-DB coverage for the approval-owned RouteReadinessReader port
// (route_readiness_reader.go), the read the controlled-documents creation gate
// (D2) and the taxonomy has_active_route badge both depend on. The sqlmock
// sibling (route_readiness_reader_test.go) pins the SQL shape and the
// no-rows/executor semantics; this file proves the predicate is load-bearing
// against Postgres:
//
//  1. A seeded active document-subject route is discoverable by profile code
//     through BOTH port methods (the subject_key alias trigger from migration
//     0296/0298 backfills subject_key := profile_code).
//  2. A profile with no route is absent from the bulk set and reports false —
//     the gate's reject path is reachable, not vacuously true.
//  3. Cross-tenant isolation: an identical subject under another tenant is
//     invisible, proving the tenant_id predicate is real and not decorative.
//  4. HasActiveRoute observes the caller's *sql.Tx, which is how the in-tx
//     creation gate consumes it.

import (
	"context"
	"database/sql"
	"testing"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/tests/integration/testdb"
)

func TestRouteReadinessReader_ActiveDocumentRoute_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()
	reader := NewRouteReadinessReaderPG(dbc)

	route := testdb.NewApprovalRoute(t, dbc)
	routeless := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(route.TenantID))

	keys, err := reader.ActiveRouteSubjectKeys(ctx, route.TenantID, string(domain.SubjectKindDocument))
	if err != nil {
		t.Fatalf("ActiveRouteSubjectKeys: %v", err)
	}
	if _, ok := keys[route.ProfileCode]; !ok {
		t.Fatalf("expected profile %q in the active-route set, got %+v", route.ProfileCode, keys)
	}
	if _, ok := keys[routeless.ProfileCode]; ok {
		t.Fatalf("profile %q has no route and must not appear in %+v", routeless.ProfileCode, keys)
	}

	ready, err := reader.HasActiveRoute(ctx, dbc, route.TenantID, string(domain.SubjectKindDocument), route.ProfileCode)
	if err != nil {
		t.Fatalf("HasActiveRoute(routed): %v", err)
	}
	if !ready {
		t.Fatal("expected the seeded active route to be reported ready")
	}

	ready, err = reader.HasActiveRoute(ctx, dbc, route.TenantID, string(domain.SubjectKindDocument), routeless.ProfileCode)
	if err != nil {
		t.Fatalf("HasActiveRoute(routeless): %v", err)
	}
	if ready {
		t.Fatalf("profile %q has no route; the gate's reject path must be reachable", routeless.ProfileCode)
	}
}

// TestRouteReadinessReader_ActiveTemplateRoute_RealDB is the ADR 0086 arm: a
// template route is keyed by the profile (doc_type_code) it governs, so the
// SAME port answers the templates creation gate. It also proves the two kinds
// are independent — a profile with a document route only is NOT template-ready,
// which is what makes the templates gate reachable rather than vacuously true.
func TestRouteReadinessReader_ActiveTemplateRoute_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()
	reader := NewRouteReadinessReaderPG(dbc)

	docRoute := testdb.NewApprovalRoute(t, dbc)
	tmplTax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(docRoute.TenantID))
	tmplRoute := testdb.NewApprovalRoute(t, dbc,
		testdb.WithTenant(docRoute.TenantID),
		testdb.WithProfile(tmplTax.ProfileCode),
		testdb.WithSubjectKind(string(domain.SubjectKindTemplate)))

	keys, err := reader.ActiveRouteSubjectKeys(ctx, docRoute.TenantID, string(domain.SubjectKindTemplate))
	if err != nil {
		t.Fatalf("ActiveRouteSubjectKeys(template): %v", err)
	}
	if _, ok := keys[tmplRoute.ProfileCode]; !ok {
		t.Fatalf("expected doc type %q in the active template-route set, got %+v", tmplRoute.ProfileCode, keys)
	}
	if _, ok := keys[docRoute.ProfileCode]; ok {
		t.Fatalf("profile %q has only a DOCUMENT route and must not appear in the template set %+v", docRoute.ProfileCode, keys)
	}

	ready, err := reader.HasActiveRoute(ctx, dbc, tmplRoute.TenantID, string(domain.SubjectKindTemplate), tmplRoute.ProfileCode)
	if err != nil {
		t.Fatalf("HasActiveRoute(template, routed): %v", err)
	}
	if !ready {
		t.Fatal("expected the seeded active template route to be reported ready")
	}

	ready, err = reader.HasActiveRoute(ctx, dbc, docRoute.TenantID, string(domain.SubjectKindTemplate), docRoute.ProfileCode)
	if err != nil {
		t.Fatalf("HasActiveRoute(template, document-only profile): %v", err)
	}
	if ready {
		t.Fatalf("profile %q has no TEMPLATE route; the templates creation gate must reject it", docRoute.ProfileCode)
	}
}

func TestRouteReadinessReader_CrossTenantIsolation_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()
	reader := NewRouteReadinessReaderPG(dbc)

	mine := testdb.NewApprovalRoute(t, dbc)
	theirs := testdb.NewApprovalRoute(t, dbc)
	if mine.TenantID == theirs.TenantID {
		t.Fatal("factory produced a shared tenant; the isolation assertion would be vacuous")
	}

	keys, err := reader.ActiveRouteSubjectKeys(ctx, mine.TenantID, string(domain.SubjectKindDocument))
	if err != nil {
		t.Fatalf("ActiveRouteSubjectKeys: %v", err)
	}
	if _, leaked := keys[theirs.ProfileCode]; leaked {
		t.Fatalf("another tenant's subject %q leaked into %+v", theirs.ProfileCode, keys)
	}

	ready, err := reader.HasActiveRoute(ctx, dbc, mine.TenantID, string(domain.SubjectKindDocument), theirs.ProfileCode)
	if err != nil {
		t.Fatalf("HasActiveRoute: %v", err)
	}
	if ready {
		t.Fatalf("another tenant's route for %q must be invisible", theirs.ProfileCode)
	}
}

func TestRouteReadinessReader_HasActiveRoute_InCallerTx_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()
	reader := NewRouteReadinessReaderPG(dbc)
	route := testdb.NewApprovalRoute(t, dbc)

	tx, err := dbc.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	ready, err := reader.HasActiveRoute(ctx, tx, route.TenantID, string(domain.SubjectKindDocument), route.ProfileCode)
	if err != nil {
		t.Fatalf("HasActiveRoute on caller tx: %v", err)
	}
	if !ready {
		t.Fatal("expected ready=true when reading through the caller's transaction")
	}
}
