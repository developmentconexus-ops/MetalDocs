//go:build integration
// +build integration

package application

// P2.S3 (M3 kernel extraction, ADR 0082) — real-DB proof that
// RouteAdminService.Create's subject_kind/subject_key plumbing is byte-equal
// for the existing document route-admin path when the fields are omitted,
// and persists faithfully when supplied. Runs against a real Postgres testdb
// (ADR 0034 harness) so it exercises the actual approval_routes
// subject_kind/subject_key columns (migration from P2.S1) rather than the
// fake-driver captured-args assertions in route_admin_service_test.go.

import (
	"context"
	"testing"

	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

func routeAdminSubjectCtx(tenantID, actorID string) context.Context {
	ctx := platformtenant.WithTenantID(context.Background(), tenantID)
	ctx = platformtenant.WithActorID(ctx, actorID)
	return ctx
}

// TestRouteAdminCreate_SubjectDefault_RealDB proves the byte-equal default:
// a CreateRouteInput with no SubjectKind/SubjectKey persists
// subject_kind='document', subject_key=profile_code.
func TestRouteAdminCreate_SubjectDefault_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID))

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	svc := &RouteAdminService{repo: repo, emitter: NewSQLEmitter(), clock: RealClock{}}
	runner := db.NewTxRunner(dbc)

	out, err := svc.Create(routeAdminSubjectCtx(tnt.ID, actor.ID), runner, CreateRouteInput{
		TenantID:    tnt.ID,
		ProfileCode: tax.ProfileCode,
		Name:        "Default Subject Route",
		ActorUserID: actor.ID,
		Stages:      validRouteStages(),
	})
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}

	var subjectKind, subjectKey string
	if err := dbc.QueryRowContext(ctx,
		`SELECT subject_kind, subject_key FROM public.approval_routes WHERE id = $1::uuid`,
		out.RouteID,
	).Scan(&subjectKind, &subjectKey); err != nil {
		t.Fatalf("query persisted subject: %v", err)
	}
	if subjectKind != "document" {
		t.Errorf("subject_kind = %q, want %q", subjectKind, "document")
	}
	if subjectKey != tax.ProfileCode {
		t.Errorf("subject_key = %q, want %q (profile_code)", subjectKey, tax.ProfileCode)
	}
}

// TestRouteAdminCreate_ExplicitSubject_RealDB proves an explicit
// subject_kind=document + subject_key is accepted and persisted as given
// (not silently overridden by the profile_code default).
func TestRouteAdminCreate_ExplicitSubject_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID))

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	svc := &RouteAdminService{repo: repo, emitter: NewSQLEmitter(), clock: RealClock{}}
	runner := db.NewTxRunner(dbc)

	out, err := svc.Create(routeAdminSubjectCtx(tnt.ID, actor.ID), runner, CreateRouteInput{
		TenantID:    tnt.ID,
		ProfileCode: tax.ProfileCode,
		Name:        "Explicit Subject Route",
		ActorUserID: actor.ID,
		SubjectKind: "document",
		SubjectKey:  "explicit-key-123",
		Stages:      validRouteStages(),
	})
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}

	var subjectKind, subjectKey string
	if err := dbc.QueryRowContext(ctx,
		`SELECT subject_kind, subject_key FROM public.approval_routes WHERE id = $1::uuid`,
		out.RouteID,
	).Scan(&subjectKind, &subjectKey); err != nil {
		t.Fatalf("query persisted subject: %v", err)
	}
	if subjectKind != "document" {
		t.Errorf("subject_kind = %q, want %q", subjectKind, "document")
	}
	if subjectKey != "explicit-key-123" {
		t.Errorf("subject_key = %q, want %q", subjectKey, "explicit-key-123")
	}
}

// TestRouteAdminList_ExposesSubjectFields_RealDB proves the list read model
// (ListRoutesTx → infrastructure.Route) surfaces subject_kind/subject_key for
// an existing document route.
func TestRouteAdminList_ExposesSubjectFields_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID))

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	svc := &RouteAdminService{repo: repo, emitter: NewSQLEmitter(), clock: RealClock{}}
	runner := db.NewTxRunner(dbc)

	_, err := svc.Create(routeAdminSubjectCtx(tnt.ID, actor.ID), runner, CreateRouteInput{
		TenantID:    tnt.ID,
		ProfileCode: tax.ProfileCode,
		Name:        "List Subject Route",
		ActorUserID: actor.ID,
		Stages:      validRouteStages(),
	})
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}

	listOut, err := svc.List(routeAdminSubjectCtx(tnt.ID, actor.ID), runner, tnt.ID, actor.ID)
	if err != nil {
		t.Fatalf("List: unexpected error: %v", err)
	}
	if len(listOut.Routes) != 1 {
		t.Fatalf("routes len = %d, want 1", len(listOut.Routes))
	}
	got := listOut.Routes[0]
	if got.SubjectKind != "document" || got.SubjectKey != tax.ProfileCode {
		t.Fatalf("subject fields = kind=%q key=%q; want kind=document key=%q", got.SubjectKind, got.SubjectKey, tax.ProfileCode)
	}
}
