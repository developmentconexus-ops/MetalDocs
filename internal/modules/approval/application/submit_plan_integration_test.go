//go:build integration
// +build integration

package application

// ADR 0085 Stage B — submit-time publication plan, real-DB proof.
//
// Publication stopped being an endpoint: the plan (when it should become
// effective, when it expires, when it must be reviewed, and which OTHER
// document it retires) is stated once, at submission, and the release
// coordinator later reacts to it. These two tests pin the two halves of that
// move against a real Postgres testdb (ADR 0034 harness):
//
//  1. TestSubmitPersistsPublicationPlan_RealDB — the four plan values land on
//     the SAME draft->under_review UPDATE that already writes revision_title /
//     reason_for_change, and `effective_from` (the ACTUAL release instant,
//     owned solely by the release transaction) stays NULL. Plan is not actual.
//
//  2. TestSubmitCrossDocumentSupersedeDeniedOnTargetArea_RealDB — naming a
//     cross-document supersede target requires `document.supersede` on the
//     TARGET's area, not the submitting document's. The actor here holds
//     area_admin (which carries document.submit + document.edit +
//     document.supersede) in the SOURCE area and nothing at all in the target's
//     area, so a submit that would otherwise succeed must fail closed — proving
//     the capability is evaluated against the target, not the source.

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"

	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
)

// pinDocumentArea stamps process_area_code_snapshot on a documents row so
// resolveSubjectAreaCode (submit_service.go step 4a / 4a-bis) resolves a REAL
// area instead of the fail-closed "". public.documents carries the
// trg_require_cap_asserted tripwire on UPDATE (document.edit is one of its
// legal arms) and is tenant-scoped, so the write runs with both the capability
// assertion and the tenant GUC seeded through the production seams.
func pinDocumentArea(t *testing.T, dbc *sql.DB, tenantID, documentID, areaCode string) {
	t.Helper()
	testdb.SeedWithCaps(t, dbc, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		if err := authz.SeedTxTenant(context.Background(), tx, tenantID); err != nil {
			return err
		}
		_, err := tx.ExecContext(context.Background(),
			`UPDATE public.documents
			    SET process_area_code_snapshot = $3
			  WHERE id = $1::uuid AND tenant_id = $2::uuid`,
			documentID, tenantID, areaCode,
		)
		return err
	})
}

// grantAreaRole gives userID the given role in areaCode. The area row itself is
// assumed already seeded (seedSubmitRouteWithStage seeds 'area-001' for the
// tenant); user_process_areas FK-references metaldocs.document_process_areas.
func grantAreaRole(t *testing.T, dbc *sql.DB, tenantID, userID, areaCode, role string) {
	t.Helper()
	testdb.SeedWithCaps(t, dbc, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO metaldocs.user_process_areas
			    (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
			 VALUES ($1, $2::uuid, $3, $4, now(), NULL, $1)`,
			userID, tenantID, areaCode, role,
		)
		return err
	})
}

// documentProfileCode reads the controlled document's profile so the seeded
// approval route matches the document (route selection is profile-scoped).
func documentProfileCode(t *testing.T, dbc *sql.DB, controlledDocumentID string) string {
	t.Helper()
	var profileCode string
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT profile_code FROM public.controlled_documents WHERE id = $1::uuid`, controlledDocumentID,
	).Scan(&profileCode); err != nil {
		t.Fatalf("lookup profile_code: %v", err)
	}
	return profileCode
}

func TestSubmitPersistsPublicationPlan_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	doc := testdb.NewDocument(t, dbc,
		testdb.WithTenant(tnt.ID), testdb.WithOwner(actor.ID), testdb.WithStatus("draft"),
		testdb.WithRevisionNumber(1), testdb.WithSubmitReadySnapshots())

	// The supersede target is a PUBLISHED document under a DIFFERENT controlled
	// document (own CD, minted by the factory): same-lineage supersession is
	// implicit in the release transaction and must never be named in a plan.
	target := testdb.NewDocument(t, dbc,
		testdb.WithTenant(tnt.ID), testdb.WithOwner(actor.ID), testdb.WithStatus("published"))

	routeID := seedSubmitRouteWithStage(t, dbc, tnt.ID, documentProfileCode(t, dbc, doc.ControlledDocumentID))

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	svc := &SubmitService{repo: repo, emitter: NewSQLEmitter(), clock: RealClock{}}
	runner := db.NewTxRunner(dbc)

	now := time.Now().UTC().Truncate(time.Second)
	planned := now.Add(72 * time.Hour)
	expires := now.Add(400 * 24 * time.Hour)
	reviewDue := now.Add(180 * 24 * time.Hour)

	req := SubmitRequest{
		TenantID:             tnt.ID,
		DocumentID:           doc.ID,
		RouteID:              routeID,
		SubmittedBy:          actor.ID,
		RevisionTitle:        "Revisao com plano de publicacao",
		ReasonForChange:      "Planned release window agreed with quality",
		ReasonCategory:       "content",
		ContentFormData:      map[string]any{"title": "My Doc"},
		RevisionVersion:      0,
		IdempotencyKey:       "33333333-3333-3333-3333-333333333333",
		PlannedEffectiveFrom: &planned,
		EffectiveTo:          &expires,
		ReviewDueAt:          &reviewDue,
		SupersedeTargetID:    target.ID,
	}

	if _, err := svc.SubmitRevisionForReview(submitCtxWithIdentity(tnt.ID, actor.ID), runner, req); err != nil {
		t.Fatalf("SubmitRevisionForReview: unexpected error: %v", err)
	}

	var (
		status                 string
		gotPlanned             sql.NullTime
		gotEffectiveTo         sql.NullTime
		gotReviewDue           sql.NullTime
		gotSuperseded          sql.NullString
		gotActualEffectiveFrom sql.NullTime
	)
	if err := dbc.QueryRowContext(ctx,
		`SELECT status, planned_effective_from, effective_to, review_due_at,
		        superseded_document_id::text, effective_from
		   FROM public.documents WHERE id = $1::uuid`,
		doc.ID,
	).Scan(&status, &gotPlanned, &gotEffectiveTo, &gotReviewDue, &gotSuperseded, &gotActualEffectiveFrom); err != nil {
		t.Fatalf("read back documents row: %v", err)
	}

	if status != "under_review" {
		t.Fatalf("status = %q, want under_review", status)
	}
	if !gotPlanned.Valid || !gotPlanned.Time.Equal(planned) {
		t.Fatalf("planned_effective_from = %v (valid=%v), want %v", gotPlanned.Time, gotPlanned.Valid, planned)
	}
	if !gotEffectiveTo.Valid || !gotEffectiveTo.Time.Equal(expires) {
		t.Fatalf("effective_to = %v (valid=%v), want %v", gotEffectiveTo.Time, gotEffectiveTo.Valid, expires)
	}
	if !gotReviewDue.Valid || !gotReviewDue.Time.Equal(reviewDue) {
		t.Fatalf("review_due_at = %v (valid=%v), want %v", gotReviewDue.Time, gotReviewDue.Valid, reviewDue)
	}
	if !gotSuperseded.Valid || gotSuperseded.String != target.ID {
		t.Fatalf("superseded_document_id = %q (valid=%v), want %q", gotSuperseded.String, gotSuperseded.Valid, target.ID)
	}
	// ADR 0085's planned/actual split: submission states a PLAN. The actual
	// release instant is written only by the release transaction.
	if gotActualEffectiveFrom.Valid {
		t.Fatalf("effective_from = %v, want NULL — submit must never stamp the ACTUAL release instant", gotActualEffectiveFrom.Time)
	}
}

func TestSubmitCrossDocumentSupersedeDeniedOnTargetArea_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ctx := context.Background()

	tnt := testdb.NewTenant(t, dbc)
	// No WithRole("system_admin"): this actor must go through the real
	// capability query, not the tier-2 admin bypass.
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithDisplayName("Area Admin"))
	doc := testdb.NewDocument(t, dbc,
		testdb.WithTenant(tnt.ID), testdb.WithOwner(actor.ID), testdb.WithStatus("draft"),
		testdb.WithRevisionNumber(1), testdb.WithSubmitReadySnapshots())
	target := testdb.NewDocument(t, dbc,
		testdb.WithTenant(tnt.ID), testdb.WithOwner(actor.ID), testdb.WithStatus("published"))

	// seedSubmitRouteWithStage also seeds metaldocs.document_process_areas
	// ('area-001') for this tenant, which grantAreaRole's FK needs.
	routeID := seedSubmitRouteWithStage(t, dbc, tnt.ID, documentProfileCode(t, dbc, doc.ControlledDocumentID))

	// Source in area-001 where the actor is area_admin (document.submit +
	// document.edit + document.supersede); target in area-002 where the actor
	// holds nothing.
	pinDocumentArea(t, dbc, tnt.ID, doc.ID, "area-001")
	pinDocumentArea(t, dbc, tnt.ID, target.ID, "area-002")
	grantAreaRole(t, dbc, tnt.ID, actor.ID, "area-001", "area_admin")

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	svc := &SubmitService{repo: repo, emitter: NewSQLEmitter(), clock: RealClock{}}
	runner := db.NewTxRunner(dbc)

	req := SubmitRequest{
		TenantID:          tnt.ID,
		DocumentID:        doc.ID,
		RouteID:           routeID,
		SubmittedBy:       actor.ID,
		RevisionTitle:     "Revisao que tenta substituir fora da area",
		ReasonForChange:   "Attempts to retire a document in another area",
		ReasonCategory:    "content",
		ContentFormData:   map[string]any{"title": "My Doc"},
		RevisionVersion:   0,
		IdempotencyKey:    "44444444-4444-4444-4444-444444444444",
		SupersedeTargetID: target.ID,
	}

	_, err := svc.SubmitRevisionForReview(submitCtxWithIdentity(tnt.ID, actor.ID), runner, req)
	if err == nil {
		t.Fatal("SubmitRevisionForReview succeeded, want capability denial on the TARGET's area")
	}
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("error = %T %v, want authz.ErrCapDenied", err, err)
	}
	if denied.Capability != string(iamdomain.CapDocumentSupersede) {
		t.Fatalf("denied capability = %q, want %q", denied.Capability, iamdomain.CapDocumentSupersede)
	}
	if denied.AreaCode != "area-002" {
		t.Fatalf("denied area = %q, want area-002 (the TARGET's area, not the source's)", denied.AreaCode)
	}

	// The whole submit transaction must have rolled back: no plan, no status
	// change, no approval instance.
	var status string
	var gotSuperseded sql.NullString
	if err := dbc.QueryRowContext(ctx,
		`SELECT status, superseded_document_id::text FROM public.documents WHERE id = $1::uuid`, doc.ID,
	).Scan(&status, &gotSuperseded); err != nil {
		t.Fatalf("read back documents row: %v", err)
	}
	if status != "draft" {
		t.Fatalf("status = %q, want draft (denied submit must not commit)", status)
	}
	if gotSuperseded.Valid {
		t.Fatalf("superseded_document_id = %q, want NULL (denied submit must not commit)", gotSuperseded.String)
	}
}
