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
	"database/sql"
	"errors"
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
// (not silently overridden by the profile_code default) — as long as the
// explicit key equals profile_code. Per rail R1 (M3 P3.S2b-2 regression fix:
// application.ErrDocumentSubjectKeyMismatch), a document route's subject_key
// is the backward-compat ALIAS for profile_code, so an explicit key that
// diverges from profile_code is rejected before it ever reaches this
// assertion — see TestRouteAdminCreate_ExplicitDivergentSubjectKey_Rejected_RealDB
// for that negative case.
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
		SubjectKey:  tax.ProfileCode,
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
		t.Errorf("subject_key = %q, want %q", subjectKey, tax.ProfileCode)
	}
}

// TestRouteAdminCreate_ExplicitDivergentSubjectKey_Rejected_RealDB is the
// real-DB negative counterpart: an explicit subject_kind=document with a
// subject_key that diverges from profile_code must be rejected with
// ErrDocumentSubjectKeyMismatch (M3 P3.S2b-2 regression fix) — and no row
// must be written.
func TestRouteAdminCreate_ExplicitDivergentSubjectKey_Rejected_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID))

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	svc := &RouteAdminService{repo: repo, emitter: NewSQLEmitter(), clock: RealClock{}}
	runner := db.NewTxRunner(dbc)

	_, err := svc.Create(routeAdminSubjectCtx(tnt.ID, actor.ID), runner, CreateRouteInput{
		TenantID:    tnt.ID,
		ProfileCode: tax.ProfileCode,
		Name:        "Divergent Subject Route",
		ActorUserID: actor.ID,
		SubjectKind: "document",
		SubjectKey:  "explicit-key-123",
		Stages:      validRouteStages(),
	})
	if !errors.Is(err, ErrDocumentSubjectKeyMismatch) {
		t.Fatalf("expected ErrDocumentSubjectKeyMismatch; got %v", err)
	}

	var count int
	if err := dbc.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM public.approval_routes WHERE tenant_id = $1::uuid AND name = $2`,
		tnt.ID, "Divergent Subject Route",
	).Scan(&count); err != nil {
		t.Fatalf("count check: %v", err)
	}
	if count != 0 {
		t.Fatalf("row count for rejected create = %d, want 0", count)
	}
}

// TestRouteAdminCreate_DocumentRoute_FoundByBothProfileAndSubjectLookup_RealDB
// is the M3 P3.S2b-2 regression closure: a document route created through the
// real application entry point (RouteAdminService.Create →
// resolveCreateRouteSubject) — the exact path where a divergent document
// subject_key used to be silently accepted pre-fix — must be found by BOTH
// the legacy alias method (LoadActiveRouteIDByProfile) and the generic
// subject method (LoadActiveRouteIDBySubject), proving the delegation
// (LoadActiveRouteIDByProfile → LoadActiveRouteIDBySubject(tenantID,
// "document", profileCode), added in c0ae114a) stays correct once the
// ErrDocumentSubjectKeyMismatch guard holds.
func TestRouteAdminCreate_DocumentRoute_FoundByBothProfileAndSubjectLookup_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID))

	baseRepo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	svc := &RouteAdminService{repo: baseRepo, emitter: NewSQLEmitter(), clock: RealClock{}}
	runner := db.NewTxRunner(dbc)

	out, err := svc.Create(routeAdminSubjectCtx(tnt.ID, actor.ID), runner, CreateRouteInput{
		TenantID:    tnt.ID,
		ProfileCode: tax.ProfileCode,
		Name:        "Delegation Equivalence Route",
		ActorUserID: actor.ID,
		Stages:      validRouteStages(),
	})
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}

	// LoadActiveRouteIDByProfile / LoadActiveRouteIDBySubject are declared on
	// application.SubmitDefaultsResolver, not on the broader ApprovalRepository
	// interface — same asserted-interface pattern as
	// TestLoadActiveRouteIDBySubject_TemplateSubject_RealDB in
	// infrastructure/postgres_approval_repository_subject_integration_test.go.
	repo, ok := interface{}(baseRepo).(interface {
		LoadActiveRouteIDByProfile(ctx context.Context, tx db.Tx, tenantID, profileCode string) (string, error)
		LoadActiveRouteIDBySubject(ctx context.Context, tx db.Tx, tenantID, subjectKind, subjectKey string) (string, error)
	})
	if !ok {
		t.Fatalf("postgres approval repository does not satisfy the route-selection interface; production wiring would silently break")
	}

	tx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	viaProfile, err := repo.LoadActiveRouteIDByProfile(ctx, tx, tnt.ID, tax.ProfileCode)
	if err != nil {
		t.Fatalf("LoadActiveRouteIDByProfile: %v", err)
	}
	viaSubject, err := repo.LoadActiveRouteIDBySubject(ctx, tx, tnt.ID, "document", tax.ProfileCode)
	if err != nil {
		t.Fatalf("LoadActiveRouteIDBySubject: %v", err)
	}
	if viaProfile != out.RouteID || viaSubject != out.RouteID || viaProfile != viaSubject {
		t.Fatalf("route ids diverge: viaProfile=%q viaSubject=%q want=%q", viaProfile, viaSubject, out.RouteID)
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

// TestRouteAdminCreate_TemplateSubject_ProfileCodeNull_RealDB is the S1 (F18)
// failing-test-first case (a): a template-subject route created with an empty
// ProfileCode must succeed and persist profile_code as SQL NULL (not empty
// string), matching the DB truth in migration 0297
// (approval_routes_template_subject_projection_check forbids a non-NULL
// profile_code when subject_kind='template'). Before the S1 fix this either
// violated the DB check (in.ProfileCode bound as "" is non-NULL) or was
// unreachable because the contract layer required profile_code
// unconditionally.
func TestRouteAdminCreate_TemplateSubject_ProfileCodeNull_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	svc := &RouteAdminService{repo: repo, emitter: NewSQLEmitter(), clock: RealClock{}}
	runner := db.NewTxRunner(dbc)

	out, err := svc.Create(routeAdminSubjectCtx(tnt.ID, actor.ID), runner, CreateRouteInput{
		TenantID:    tnt.ID,
		Name:        "Template Route",
		ActorUserID: actor.ID,
		SubjectKind: "template",
		SubjectKey:  "tmpl-version-1",
		Stages:      validRouteStages(),
	})
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}

	var profileCode sql.NullString
	var subjectKind, subjectKey string
	if err := dbc.QueryRowContext(ctx,
		`SELECT profile_code, subject_kind, subject_key FROM public.approval_routes WHERE id = $1::uuid`,
		out.RouteID,
	).Scan(&profileCode, &subjectKind, &subjectKey); err != nil {
		t.Fatalf("query persisted route: %v", err)
	}
	if profileCode.Valid {
		t.Errorf("profile_code = %q, want SQL NULL", profileCode.String)
	}
	if subjectKind != "template" {
		t.Errorf("subject_kind = %q, want %q", subjectKind, "template")
	}
	if subjectKey != "tmpl-version-1" {
		t.Errorf("subject_key = %q, want %q", subjectKey, "tmpl-version-1")
	}
}

// TestRouteAdminCreate_DocumentSubject_ProfileCodePopulated_RealDB is the S1
// (F18) regression guard, case (d): a document-subject (absent/explicit
// subject_kind=document) route create with a valid ProfileCode still
// succeeds and persists profile_code populated — the template-NULL binding
// introduced by S1 must not regress the pre-existing document path.
func TestRouteAdminCreate_DocumentSubject_ProfileCodePopulated_RealDB(t *testing.T) {
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
		Name:        "Document Route Regression Guard",
		ActorUserID: actor.ID,
		Stages:      validRouteStages(),
	})
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}

	var profileCode sql.NullString
	if err := dbc.QueryRowContext(ctx,
		`SELECT profile_code FROM public.approval_routes WHERE id = $1::uuid`,
		out.RouteID,
	).Scan(&profileCode); err != nil {
		t.Fatalf("query persisted route: %v", err)
	}
	if !profileCode.Valid || profileCode.String != tax.ProfileCode {
		t.Errorf("profile_code = %+v, want valid %q", profileCode, tax.ProfileCode)
	}
}
