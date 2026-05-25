package application

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

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
		"submitted_by", "submitted_at", "stage_label", "required", "signed",
	}).AddRow(
		"inst-1", "doc-1", "CD-001", "Doc One", "finance",
		"user-1", submittedAt, "Stage 1", 2, 1,
	)

	mock.ExpectBegin()
	mock.ExpectExec(`SELECT set_config\('metaldocs\.tenant_id'`).
		WithArgs("tenant-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`SELECT set_config\('metaldocs\.actor_id'`).
		WithArgs("actor-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT[\s\S]+FROM approval_instances ai`).
		WithArgs("tenant-1", sqlmock.AnyArg(), "finance", 25, 0).
		WillReturnRows(rows)
	mock.ExpectCommit()

	svc := &ReadService{}
	items, err := svc.ListInboxItems(context.Background(), db, "tenant-1", "actor-1", "finance", 25, 0)
	if err != nil {
		t.Fatalf("ListInboxItems: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].DocumentTitle != "Doc One" {
		t.Errorf("DocumentTitle = %q, want %q", items[0].DocumentTitle, "Doc One")
	}
	if items[0].QuorumProgress != "1/2" {
		t.Errorf("QuorumProgress = %q, want %q", items[0].QuorumProgress, "1/2")
	}
	if items[0].InstanceID != "inst-1" || items[0].DocumentID != "doc-1" {
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
	mock.ExpectExec(`SELECT set_config\('metaldocs\.tenant_id'`).
		WithArgs("tenant-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`SELECT set_config\('metaldocs\.actor_id'`).
		WithArgs("actor-xyz").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`asi\.eligible_actor_ids @> \$2::jsonb`).
		WithArgs("tenant-1", []byte(`["actor-xyz"]`), "", 25, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "document_id", "doc_title", "area_code",
			"submitted_by", "submitted_at", "stage_label", "required", "signed",
		}))
	mock.ExpectCommit()

	svc := &ReadService{}
	if _, err := svc.ListInboxItems(context.Background(), db, "tenant-1", "actor-xyz", "", 0, 0); err != nil {
		t.Fatalf("ListInboxItems: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCountPendingForActor_ReturnsTotal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`SELECT set_config\('metaldocs\.tenant_id'`).
		WithArgs("tenant-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`SELECT set_config\('metaldocs\.actor_id'`).
		WithArgs("actor-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT ai\.id\)`).
		WithArgs("tenant-1", sqlmock.AnyArg(), "").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))
	mock.ExpectCommit()

	svc := &ReadService{}
	total, err := svc.CountPendingForActor(context.Background(), db, "tenant-1", "actor-1", "")
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
