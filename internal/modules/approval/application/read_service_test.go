package application

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
)

type readServiceRepoSpy struct {
	infrastructure.ApprovalRepository
	called bool
	inst   *domain.Instance
}

func (s *readServiceRepoSpy) LoadInstance(context.Context, db.Tx, string, string) (*domain.Instance, error) {
	return nil, nil
}

func (s *readServiceRepoSpy) LoadActiveInstanceByDocument(_ context.Context, _ db.Tx, _, _ string) (*domain.Instance, error) {
	s.called = true
	return s.inst, nil
}

func TestReadService_LoadsLockedApprovalInstancesOutsideReadOnlyTransactions(t *testing.T) {
	source, err := os.ReadFile("read_service.go")
	if err != nil {
		t.Fatalf("read read_service.go: %v", err)
	}

	text := string(source)
	for _, fn := range []string{"LoadInstance", "LoadActiveInstanceByDocument", "ListPendingForActor"} {
		start := strings.Index(text, "func (s *ReadService) "+fn)
		if start < 0 {
			t.Fatalf("missing %s", fn)
		}
		body := text[start:]
		next := strings.Index(body[len("func "):], "\nfunc ")
		if next >= 0 {
			body = body[:len("func ")+next]
		}
		if strings.Contains(body, "ReadOnly: true") {
			t.Fatalf("%s must not open a read-only transaction because approval repository stage loads use SELECT ... FOR UPDATE", fn)
		}
	}
}

func TestListInboxItems_PopulatesTitleAndQuorumProgress(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	submittedAt := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"id", "document_id", "controlled_document_id", "doc_title", "area_code",
		"submitted_by", "submitted_at", "stage_label", "required", "signed", "total_count",
	}).AddRow(
		"inst-1", "doc-1", "CD-001", "Doc One", "finance",
		"user-1", submittedAt, "Stage 1", 2, 1, 0,
	)

	mock.ExpectBegin()
	mock.ExpectExec(`set_config\('metaldocs\.tenant_id'`).
		WithArgs("tenant-1", "actor-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT[\s\S]+FROM approval_instances ai`).
		WithArgs("tenant-1", sqlmock.AnyArg(), "finance", 25, 0).
		WillReturnRows(rows)
	mock.ExpectCommit()

	svc := &ReadService{}
	items, err := svc.ListInboxItems(authzCtx("tenant-1", "actor-1"), newTxRunner(db), "tenant-1", "actor-1", "finance", 25, 0)
	if err != nil {
		t.Fatalf("ListInboxItems: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].SubjectTitle != "Doc One" {
		t.Errorf("SubjectTitle = %q, want %q", items[0].SubjectTitle, "Doc One")
	}
	if items[0].QuorumProgress != "1/2" {
		t.Errorf("QuorumProgress = %q, want %q", items[0].QuorumProgress, "1/2")
	}
	if items[0].InstanceID != "inst-1" || items[0].SubjectKey != "doc-1" {
		t.Errorf("ID mapping wrong: %+v", items[0])
	}
	if items[0].StageLabel != "Stage 1" || items[0].AreaCode != "finance" {
		t.Errorf("Stage/Area mapping wrong: %+v", items[0])
	}
	if !items[0].SubmittedAt.Equal(submittedAt) {
		t.Errorf("SubmittedAt = %v, want %v", items[0].SubmittedAt, submittedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestListInboxItems_FiltersByActor(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// We assert the actorID is JSON-marshalled into the eligible_actor_ids @> filter arg ($2).
	mock.ExpectBegin()
	mock.ExpectExec(`set_config\('metaldocs\.tenant_id'`).
		WithArgs("tenant-1", "actor-xyz").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// limit=0 clamps to pagination.DefaultLimit (20), not the old handler-only
	// default of 25 — the service layer now owns clamping via
	// internal/platform/pagination.ClampLimit (T-005 hardening).
	mock.ExpectQuery(`asi\.eligible_actor_ids @> \$2::jsonb`).
		WithArgs("tenant-1", []byte(`["actor-xyz"]`), "", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "document_id", "doc_title", "area_code",
			"submitted_by", "submitted_at", "stage_label", "required", "signed", "total_count",
		}))
	mock.ExpectCommit()

	svc := &ReadService{}
	if _, err := svc.ListInboxItems(authzCtx("tenant-1", "actor-xyz"), newTxRunner(db), "tenant-1", "actor-xyz", "", 0, 0); err != nil {
		t.Fatalf("ListInboxItems: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestListInboxItemsWithTotal_SingleQueryCarriesTotal pins the T-005 fix: the
// total is read via COUNT(*) OVER() on the SAME query/snapshot as the page,
// not a second independent query. It also pins the ai.id DESC ORDER BY
// tiebreaker added to prevent OFFSET-page interleaving on submitted_at ties.
func TestListInboxItemsWithTotal_SingleQueryCarriesTotal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	submittedAt := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"id", "document_id", "controlled_document_id", "doc_title", "area_code",
		"submitted_by", "submitted_at", "stage_label", "required", "signed", "total_count",
	}).AddRow(
		"inst-1", "doc-1", "CD-001", "Doc One", "finance",
		"user-1", submittedAt, "Stage 1", 2, 1, 5,
	).AddRow(
		"inst-2", "doc-2", "CD-002", "Doc Two", "finance",
		"user-2", submittedAt, "Stage 1", 1, 1, 5,
	)

	mock.ExpectBegin()
	mock.ExpectExec(`set_config\('metaldocs\.tenant_id'`).
		WithArgs("tenant-1", "actor-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`COUNT\(\*\) OVER\(\) AS total_count[\s\S]+ORDER BY ai\.submitted_at DESC, ai\.id DESC`).
		WithArgs("tenant-1", sqlmock.AnyArg(), "finance", 25, 0).
		WillReturnRows(rows)
	mock.ExpectCommit()

	svc := &ReadService{}
	items, total, err := svc.ListInboxItemsWithTotal(authzCtx("tenant-1", "actor-1"), newTxRunner(db), "tenant-1", "actor-1", "finance", 25, 0)
	if err != nil {
		t.Fatalf("ListInboxItemsWithTotal: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if total != 5 {
		t.Errorf("total = %d, want 5 (from COUNT(*) OVER(), not len(items))", total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestListInboxItemsWithTotal_EmptyPageFallsBackToCount pins the empty-page
// edge case: COUNT(*) OVER() never appears when zero rows match (e.g. offset
// past the end), so the total must fall back to CountPendingForActor rather
// than silently reporting 0.
func TestListInboxItemsWithTotal_EmptyPageFallsBackToCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`set_config\('metaldocs\.tenant_id'`).
		WithArgs("tenant-1", "actor-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`FROM approval_instances ai`).
		WithArgs("tenant-1", sqlmock.AnyArg(), "finance", 25, 1000).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "document_id", "controlled_document_id", "doc_title", "area_code",
			"submitted_by", "submitted_at", "stage_label", "required", "signed", "total_count",
		}))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`set_config\('metaldocs\.tenant_id'`).
		WithArgs("tenant-1", "actor-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT ai\.id\)`).
		WithArgs("tenant-1", sqlmock.AnyArg(), "finance").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectCommit()

	svc := &ReadService{}
	items, total, err := svc.ListInboxItemsWithTotal(authzCtx("tenant-1", "actor-1"), newTxRunner(db), "tenant-1", "actor-1", "finance", 25, 1000)
	if err != nil {
		t.Fatalf("ListInboxItemsWithTotal: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("got %d items, want 0", len(items))
	}
	if total != 3 {
		t.Errorf("total = %d, want 3 (from CountPendingForActor fallback)", total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestListInboxItemsWithTotal_ClampsLimit pins pagination.ClampLimit behavior:
// limit<=0 clamps to DefaultLimit (20), limit>MaxLimit clamps to MaxLimit
// (100), negative offset clamps to 0.
func TestListInboxItemsWithTotal_ClampsLimit(t *testing.T) {
	cases := []struct {
		name       string
		inLimit    int
		inOffset   int
		wantLimit  int
		wantOffset int
	}{
		{"zero limit clamps to default", 0, 0, 20, 0},
		{"negative limit clamps to default", -5, 0, 20, 0},
		{"over-max limit clamps to max", 500, 0, 100, 0},
		{"in-range limit passes through", 40, 0, 40, 0},
		{"negative offset clamps to zero", 10, -1, 10, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()

			mock.ExpectBegin()
			mock.ExpectExec(`set_config\('metaldocs\.tenant_id'`).
				WithArgs("tenant-1", "actor-1").
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectQuery(`FROM approval_instances ai`).
				WithArgs("tenant-1", sqlmock.AnyArg(), "", tc.wantLimit, tc.wantOffset).
				WillReturnRows(sqlmock.NewRows([]string{
					"id", "document_id", "controlled_document_id", "doc_title", "area_code",
					"submitted_by", "submitted_at", "stage_label", "required", "signed", "total_count",
				}).AddRow(
					"inst-1", "doc-1", "CD-001", "Doc One", "finance",
					"user-1", time.Now().UTC(), "Stage 1", 1, 0, 1,
				))
			mock.ExpectCommit()

			svc := &ReadService{}
			_, _, err = svc.ListInboxItemsWithTotal(authzCtx("tenant-1", "actor-1"), newTxRunner(db), "tenant-1", "actor-1", "", tc.inLimit, tc.inOffset)
			if err != nil {
				t.Fatalf("ListInboxItemsWithTotal: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations (limit/offset not clamped as expected): %v", err)
			}
		})
	}
}

func TestCountPendingForActor_ReturnsTotal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`set_config\('metaldocs\.tenant_id'`).
		WithArgs("tenant-1", "actor-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT ai\.id\)`).
		WithArgs("tenant-1", sqlmock.AnyArg(), "").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))
	mock.ExpectCommit()

	svc := &ReadService{}
	total, err := svc.CountPendingForActor(authzCtx("tenant-1", "actor-1"), newTxRunner(db), "tenant-1", "actor-1", "")
	if err != nil {
		t.Fatalf("CountPendingForActor: %v", err)
	}
	if total != 7 {
		t.Errorf("total = %d, want 7", total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestListWorklist_ZeroFilter_MatchesBaseInboxShape pins that an empty
// InboxFilter produces the same eligibility-pool WHERE shape as
// ListInboxItemsWithTotal (F8 must not change default worklist behavior).
func TestListWorklist_ZeroFilter_MatchesBaseInboxShape(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	submittedAt := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "subject_kind", "subject_key", "document_id", "controlled_document_id",
		"controlled_document_code", "doc_title", "area_code",
		"submitted_by", "submitted_at", "stage_label", "required", "signed",
		"stage_kind", "due_at", "total_count",
	}).AddRow(
		"inst-1", "document", "doc-1", "doc-1", "cd-uuid-1",
		"POP-QUA-0148", "Doc One", "finance",
		"user-1", submittedAt, "Stage 1", 2, 1, "approval", nil, 1,
	)

	mock.ExpectBegin()
	mock.ExpectExec(`set_config\('metaldocs\.tenant_id'`).
		WithArgs("tenant-1", "actor-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`asi\.eligible_actor_ids @> \$2::jsonb`).
		WithArgs("tenant-1", sqlmock.AnyArg(), "finance", 25, 0, "", sqlmock.AnyArg()).
		WillReturnRows(rows)
	mock.ExpectCommit()

	svc := &ReadService{}
	items, total, err := svc.ListWorklist(authzCtx("tenant-1", "actor-1"), newTxRunner(db), "tenant-1", "actor-1", "finance", InboxFilter{}, 25, 0)
	if err != nil {
		t.Fatalf("ListWorklist: %v", err)
	}
	if len(items) != 1 || total != 1 {
		t.Fatalf("got items=%d total=%d, want 1/1", len(items), total)
	}
	if items[0].StageKind != domain.StageKindApproval {
		t.Errorf("StageKind = %q, want %q", items[0].StageKind, domain.StageKindApproval)
	}
	if items[0].DueAt != nil {
		t.Errorf("DueAt = %v, want nil (no-fallback: NULL due_at stays nil)", items[0].DueAt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestListWorklist_StageKindFilter_PassesArgThrough pins that a non-empty
// StageKind filter is threaded to the query as the stage_kind predicate arg.
func TestListWorklist_StageKindFilter_PassesArgThrough(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`set_config\('metaldocs\.tenant_id'`).
		WithArgs("tenant-1", "actor-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`asi\.stage_kind = \$6`).
		WithArgs("tenant-1", sqlmock.AnyArg(), "", 25, 0, "review", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "subject_kind", "subject_key", "document_id", "controlled_document_id",
			"controlled_document_code", "doc_title", "area_code",
			"submitted_by", "submitted_at", "stage_label", "required", "signed",
			"stage_kind", "due_at", "total_count",
		}))
	mock.ExpectCommit()
	// Empty-page fallback (countWorklist) — no rows matched above.
	mock.ExpectBegin()
	mock.ExpectExec(`set_config\('metaldocs\.tenant_id'`).
		WithArgs("tenant-1", "actor-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT ai\.id\)`).
		WithArgs("tenant-1", sqlmock.AnyArg(), "", "review", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectCommit()

	svc := &ReadService{}
	_, _, err = svc.ListWorklist(authzCtx("tenant-1", "actor-1"), newTxRunner(db), "tenant-1", "actor-1", "", InboxFilter{StageKind: domain.StageKindReview}, 25, 0)
	if err != nil {
		t.Fatalf("ListWorklist: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestListWorklist_Oversee_DropsEligibilityPredicate pins that scope=oversee
// switches the eligibility predicate to TRUE (list every in-progress instance
// in the tenant) rather than filtering by the actor's own pool membership —
// but ONLY after authz.Require(CapApprovalOversee) succeeds. This test uses a
// system_admin-style short-circuit is NOT exercised here (that requires a
// real DB); this sqlmock test instead pins the SQL predicate shape by
// asserting the query text contains "TRUE" in place of the eligible_actor_ids
// check. A full authz.Require pass/fail matrix for scope=oversee is covered
// by the integration test (read_service_worklist_oversee_integration_test.go).
func TestListWorklist_Oversee_QueryShapeDropsEligibilityPredicate(t *testing.T) {
	source, err := os.ReadFile("read_service.go")
	if err != nil {
		t.Fatalf("read read_service.go: %v", err)
	}
	text := string(source)
	// The oversee-scope predicate must be a tautology that does NOT filter by
	// eligible_actor_ids (list every in-progress instance), but it must still
	// reference $2 with an inferable type — a bare "TRUE" literal leaves $2
	// completely unreferenced in the query text, which Postgres's real
	// parameter-type inference rejects with SQLSTATE 42P18 (F10 live-QA
	// finding; sqlmock cannot catch this class of bug).
	if !strings.Contains(text, `eligibilityPredicate = "($2::jsonb IS NOT NULL)"`) {
		t.Fatalf("ListWorklist must switch off the eligibility predicate entirely for scope=oversee (list every in-progress instance), not pass a wildcard actor value, while keeping $2 referenced with an inferable type (see SQLSTATE 42P18)")
	}
}

func TestLoadActiveInstanceByDocument_RequiresDocumentViewBeforeRepoLoad(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repoSpy := &readServiceRepoSpy{inst: &domain.Instance{ID: "inst-1"}}
	svc := &ReadService{repo: repoSpy}

	mock.ExpectBegin()
	mock.ExpectExec(`set_config\('metaldocs\.tenant_id'`).
		WithArgs("tenant-1", "actor-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT current_setting\('metaldocs\.actor_id', true\)`).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow("actor-1"))
	mock.ExpectQuery(`SELECT current_setting\('metaldocs\.tenant_id', true\)`).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow("tenant-1"))
	mock.ExpectQuery(`SELECT EXISTS \([\s\S]*iam_user_roles`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(`SELECT EXISTS \([\s\S]*role_capabilities`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectRollback()

	ctx := tenant.WithTenantID(context.Background(), "tenant-1")
	ctx = tenant.WithActorID(ctx, "actor-1")
	ctx = iamdomain.WithAuthContext(ctx, "actor-1", nil)
	_, err = svc.LoadActiveInstanceByDocument(ctx, newTxRunner(db), "tenant-1", "doc-1")
	if err == nil {
		t.Fatal("expected authz denial")
	}
	var denied authz.ErrCapDenied
	if !errors.As(err, &denied) {
		t.Fatalf("expected ErrCapDenied, got %v", err)
	}
	// F0.3: CapDocumentView is tenant-grade — denial envelope carries "tenant" sentinel,
	// not the document's resolved area code.
	if denied.AreaCode != "tenant" {
		t.Fatalf("ErrCapDenied.AreaCode = %q, want %q (tenant-grade)", denied.AreaCode, "tenant")
	}
	if repoSpy.called {
		t.Fatal("repo must not be called when document.view is denied")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestLoadActiveInstanceByDocumentForMutation_DoesNotRequireDocumentView(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repoSpy := &readServiceRepoSpy{inst: &domain.Instance{ID: "inst-1"}}
	svc := &ReadService{repo: repoSpy}

	mock.ExpectBegin()
	mock.ExpectExec(`set_config\('metaldocs\.tenant_id'`).
		WithArgs("tenant-1", "actor-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx := tenant.WithTenantID(context.Background(), "tenant-1")
	ctx = tenant.WithActorID(ctx, "actor-1")
	ctx = iamdomain.WithAuthContext(ctx, "actor-1", nil)
	inst, err := svc.LoadActiveInstanceByDocumentForMutation(ctx, newTxRunner(db), "tenant-1", "doc-1")
	if err != nil {
		t.Fatalf("LoadActiveInstanceByDocumentForMutation: %v", err)
	}
	if inst == nil || inst.ID != "inst-1" {
		t.Fatalf("instance = %#v, want inst-1", inst)
	}
	if !repoSpy.called {
		t.Fatal("repo must be called on mutation lookup")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
