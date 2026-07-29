package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/riverqueue/river"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
)

// ErrReleaseEvaluateJobNil is returned by Work when River delivers a nil job.
var ErrReleaseEvaluateJobNil = errors.New("release evaluate job is nil")

// ReleaseEvaluateWorker runs one release-coordinator evaluation.
type ReleaseEvaluateWorker struct {
	river.WorkerDefaults[ReleaseEvaluateArgs]

	coordinator *application.ReleaseCoordinator
	runner      db.TxRunner
}

// NewReleaseEvaluateWorker constructs the worker. Panics on nil dependencies —
// a release coordinator that cannot reach the database is a wiring error, not
// a runtime condition.
func NewReleaseEvaluateWorker(coordinator *application.ReleaseCoordinator, database *sql.DB) *ReleaseEvaluateWorker {
	if coordinator == nil {
		panic("release_evaluate_worker: coordinator is nil")
	}
	if database == nil {
		panic("release_evaluate_worker: db is nil")
	}
	return &ReleaseEvaluateWorker{coordinator: coordinator, runner: db.NewTxRunner(database)}
}

// Work evaluates one release generation under a background-bypass authz
// context (no HTTP-request identity exists here — ADR 0022 Phase 7).
//
// A generation that no longer exists is a success, not a failure: the tenant
// was erased or the row purged, and retrying forever would only fill the DLQ.
func (w *ReleaseEvaluateWorker) Work(ctx context.Context, job *river.Job[ReleaseEvaluateArgs]) error {
	if job == nil {
		return ErrReleaseEvaluateJobNil
	}
	ctx = authz.WithBackgroundBypass(ctx)

	key := argsToKey(job.Args)
	eval, err := w.coordinator.Evaluate(ctx, w.runner, key)
	if errors.Is(err, application.ErrReleaseGenerationNotFound) {
		slog.InfoContext(ctx, "release evaluation skipped",
			"document_id", key.DocumentID, "reason", "generation_not_found")
		return nil
	}
	if err != nil {
		return fmt.Errorf("release evaluate: %w", err)
	}
	slog.InfoContext(ctx, "release evaluated",
		"document_id", eval.DocumentID,
		"action", string(eval.Action),
		"hold", string(eval.Hold),
	)
	return nil
}

func argsToKey(a ReleaseEvaluateArgs) domain.ReleaseGenerationKey {
	kind := domain.SubjectKind(a.SubjectKind)
	if kind == "" {
		kind = domain.SubjectKindDocument
	}
	return domain.ReleaseGenerationKey{
		TenantID:           a.TenantID,
		SubjectKind:        kind,
		DocumentID:         a.DocumentID,
		ApprovalInstanceID: a.ApprovalInstanceID,
		RevisionID:         a.RevisionID,
		RevisionVersion:    a.RevisionVersion,
		FrozenContentHash:  a.FrozenContentHash,
	}
}

// DeferredReleaseEvaluationEnqueuer breaks the wiring cycle in the jobs
// binary: the worker set (which needs a coordinator, which needs an enqueuer)
// must be constructed BEFORE the River client that the enqueuer inserts into
// exists. Bind is called once, during startup, before the client starts
// working jobs — so no synchronisation is needed and an unbound use is a
// wiring bug that fails loud rather than silently dropping evaluations.
type DeferredReleaseEvaluationEnqueuer struct {
	inner application.ReleaseEvaluationEnqueuer
}

// NewDeferredReleaseEvaluationEnqueuer returns an unbound enqueuer.
func NewDeferredReleaseEvaluationEnqueuer() *DeferredReleaseEvaluationEnqueuer {
	return &DeferredReleaseEvaluationEnqueuer{}
}

// Bind supplies the River client once it exists.
func (e *DeferredReleaseEvaluationEnqueuer) Bind(client *river.Client[*sql.Tx]) {
	e.inner = NewReleaseEvaluationEnqueuer(client)
}

// EnqueueReleaseEvaluationTx delegates to the bound enqueuer.
func (e *DeferredReleaseEvaluationEnqueuer) EnqueueReleaseEvaluationTx(ctx context.Context, tx db.Tx, key domain.ReleaseGenerationKey, runAt time.Time) error {
	if e.inner == nil {
		return fmt.Errorf("release_evaluate: enqueuer used before Bind (startup wiring bug)")
	}
	return e.inner.EnqueueReleaseEvaluationTx(ctx, tx, key, runAt)
}

// RiverReleaseEvaluationEnqueuer implements
// application.ReleaseEvaluationEnqueuer using River's same-tx InsertTx, so the
// evaluation is enqueued in the same transaction as the fact that warrants it
// (transactional outbox).
type RiverReleaseEvaluationEnqueuer struct {
	Client riverInserter
}

// NewReleaseEvaluationEnqueuer wraps a River client as the application port.
func NewReleaseEvaluationEnqueuer(client *river.Client[*sql.Tx]) application.ReleaseEvaluationEnqueuer {
	return &RiverReleaseEvaluationEnqueuer{Client: client}
}

// EnqueueReleaseEvaluationTx enqueues an evaluation inside tx. A zero runAt
// means "as soon as possible"; a future runAt arms the effective-date timer.
func (e *RiverReleaseEvaluationEnqueuer) EnqueueReleaseEvaluationTx(ctx context.Context, tx db.Tx, key domain.ReleaseGenerationKey, runAt time.Time) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return fmt.Errorf("release_evaluate: river requires *sql.Tx, got %T", tx)
	}
	kind := key.SubjectKind
	if kind == "" {
		kind = domain.SubjectKindDocument
	}
	opts := &river.InsertOpts{Queue: "temporal"}
	if !runAt.IsZero() {
		opts.ScheduledAt = runAt.UTC()
	}
	_, err := e.Client.InsertTx(ctx, sqlTx, ReleaseEvaluateArgs{
		TenantID:           key.TenantID,
		SubjectKind:        string(kind),
		DocumentID:         key.DocumentID,
		ApprovalInstanceID: key.ApprovalInstanceID,
		RevisionID:         key.RevisionID,
		RevisionVersion:    key.RevisionVersion,
		FrozenContentHash:  key.FrozenContentHash,
	}, opts)
	if err != nil {
		return fmt.Errorf("release_evaluate: enqueue evaluation for document %s: %w", key.DocumentID, err)
	}
	return nil
}
