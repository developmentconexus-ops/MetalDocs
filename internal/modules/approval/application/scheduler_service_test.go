package application

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
)

// M3 F3.2 PG-2 (validation-contract.md §2.2 site 3) — RunScheduledPublishJob
// must call authz.SeedTxTenant with the River-args tenant BEFORE the
// loadScheduledDocumentState `FOR UPDATE` query (H-PRE-1: seed before any
// lock). This is a sqlmock ordering test: the seed's set_config statement
// must be observed before the FOR UPDATE SELECT.

func TestRunScheduledPublishJob_SeedsTenantBeforeLock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer mockDB.Close()

	mock.ExpectBegin()
	mock.MatchExpectationsInOrder(true)
	// 1) SeedTxTenant must run FIRST, before any lock (H-PRE-1).
	mock.ExpectExec(`SELECT set_config\('metaldocs\.tenant_id', \$1, true\)`).
		WithArgs("tenant-sched-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 2) BypassSystem's tripwire audit posture — its own SQL is exercised by
	//    its own unit tests, so keep this a permissive matcher; ordering
	//    relative to the seed above is what this test pins.
	mock.ExpectExec(`set_config`).WillReturnResult(sqlmock.NewResult(0, 1)) // BypassSystem
	// loadScheduledDocumentState's FOR UPDATE must come strictly after both
	// of the above (MatchExpectationsInOrder enforces this).
	mock.ExpectQuery(`FOR UPDATE`).WillReturnError(sql.ErrNoRows)
	// document-not-found is handled as a successful no-op (nil error), so the
	// runner commits the (no-op) tx rather than rolling back.
	mock.ExpectCommit()

	runner := &fakeTxRunnerFromSQLDB{db: mockDB}
	svc := &SchedulerService{repo: nil, emitter: nil, clock: fixedClockForTest{t: time.Now()}}

	ctx := authz.WithBackgroundBypass(context.Background())
	err = svc.RunScheduledPublishJob(ctx, runner, ScheduledPublishJobInput{
		TenantID:                "tenant-sched-1",
		DocumentID:              "doc-1",
		ExpectedRevisionVersion: 1,
		ScheduledEffectiveAt:    time.Now(),
		ScheduleGeneration:      1,
	})
	// document-not-found short-circuits to nil (see RunScheduledPublishJob).
	if err != nil {
		t.Fatalf("RunScheduledPublishJob: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations (seed-before-lock ordering violated?): %v", err)
	}
}

type fixedClockForTest struct{ t time.Time }

func (c fixedClockForTest) Now() time.Time { return c.t }

// fakeTxRunnerFromSQLDB adapts a sqlmock *sql.DB into a db.TxRunner for
// this ordering test — Do begins a real *sql.Tx against the mock so the
// callback observes the mock's expectation sequence.
type fakeTxRunnerFromSQLDB struct {
	db *sql.DB
}

func (r *fakeTxRunnerFromSQLDB) Do(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *fakeTxRunnerFromSQLDB) DoReadOnly(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return r.Do(ctx, fn)
}

var _ db.TxRunner = (*fakeTxRunnerFromSQLDB)(nil)
