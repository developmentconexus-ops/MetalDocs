// Package dispatchjobs holds the River queue-job workers that dispatch
// staging pdf/materialize outbox rows as domain events. The staging outbox
// row (internal/modules/render/fanout.StagingOutboxRepository) remains the
// dedup/audit/retention record; River owns retry/backoff for delivery. Each
// worker publishes the corresponding messaging.Event and then marks the
// outbox row dispatched, or — on River's terminal attempt — dead-letters it.
package dispatchjobs

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/platform/messaging"
)

// ErrPDFDispatchJobNil is returned by PDFDispatchWorker.Work when River
// delivers a nil job.
var ErrPDFDispatchJobNil = errors.New("pdf dispatch job is nil")

// ErrMaterializeDispatchJobNil is returned by MaterializeDispatchWorker.Work
// when River delivers a nil job.
var ErrMaterializeDispatchJobNil = errors.New("materialize dispatch job is nil")

// outboxMarker captures the two StagingOutboxRepository methods the dispatch
// workers use. Kept unexported and minimal so tests can supply a fake without
// standing up a real *fanout.StagingOutboxRepository (whose methods are
// concrete, not behind an interface, at the repo's own package boundary).
type outboxMarker interface {
	MarkDispatched(ctx context.Context, tenantID, id string) error
	MarkFailed(ctx context.Context, tenantID, id, errStr string) error
}

// dispatchIdempotencyKey widens the delivery idempotency identity from
// revision-only to generation-aware (ADR 0085). Two approvals of the same
// revision are two DIFFERENT release generations and must each produce their
// own artifact set; keying on the revision alone would make the second one a
// duplicate and leave its generation stuck on the "materializing" hold.
// Generation-less (legacy, non-approval) renders keep the historical key
// verbatim, so nothing that already shipped changes shape.
func dispatchIdempotencyKey(prefix string, f dispatchFields) string {
	key := prefix + ":" + f.TenantID + ":" + f.RevisionID
	if f.ReleaseGenerationID != "" {
		key += ":" + f.ReleaseGenerationID
	}
	return key
}

// buildPDFEvent reproduces, verbatim, the pdf outbox buildEvent closure
// previously inlined in apps/api/cmd/metaldocs-api/main.go.
func buildPDFEvent(f dispatchFields) messaging.Event {
	return messaging.Event{
		EventID:        messaging.EventID(uuid.NewString()),
		EventType:      messaging.EventTypePDFConvert,
		AggregateType:  messaging.AggregateType("document_revision"),
		AggregateID:    messaging.AggregateID(f.RevisionID),
		IdempotencyKey: messaging.IdempotencyKey(dispatchIdempotencyKey("docgen_v2_pdf", f)),
		Payload: messaging.PDFConvertPayload{
			TenantID:            f.TenantID,
			RevisionID:          f.RevisionID,
			ContentHash:         hex.EncodeToString(f.ContentHash),
			FinalDocxS3Key:      f.FinalDocxS3Key,
			ReleaseGenerationID: f.ReleaseGenerationID,
		},
	}
}

// buildMaterializeEvent reproduces, verbatim, the materialize outbox
// buildEvent closure previously inlined in apps/api/cmd/metaldocs-api/main.go.
func buildMaterializeEvent(f dispatchFields) messaging.Event {
	return messaging.Event{
		EventID:        messaging.EventID(uuid.NewString()),
		EventType:      messaging.EventTypeMaterializeFanout,
		AggregateType:  messaging.AggregateType("document_revision"),
		AggregateID:    messaging.AggregateID(f.RevisionID),
		IdempotencyKey: messaging.IdempotencyKey(dispatchIdempotencyKey("materialize_fanout", f)),
		Payload: messaging.MaterializeFanoutPayload{
			TenantID:            f.TenantID,
			RevisionID:          f.RevisionID,
			ReleaseGenerationID: f.ReleaseGenerationID,
		},
	}
}

// run is the shared body for both dispatch workers: publish the built event,
// then mark the outbox row dispatched. On any error, if this was River's
// terminal attempt (attempt >= maxAttempts), the row is dead-lettered
// (MarkFailed with finalize=true) so it isn't retried after River gives up;
// the original error is always returned so River records its own attempt
// history. A MarkFailed error at that point is intentionally not fatal to
// the return value — River's discard decision still needs the original err.
func run(ctx context.Context, pub messaging.Publisher, repo outboxMarker, fields dispatchFields, buildEvent func(dispatchFields) messaging.Event, attempt, maxAttempts int) error {
	ctx = authz.WithBackgroundBypass(ctx)

	evt := buildEvent(fields)
	if err := pub.Publish(ctx, evt); err != nil {
		terminalDeadLetter(ctx, repo, fields, err, attempt, maxAttempts)
		return err
	}

	if err := repo.MarkDispatched(ctx, fields.TenantID, fields.OutboxID); err != nil {
		terminalDeadLetter(ctx, repo, fields, err, attempt, maxAttempts)
		return err
	}
	return nil
}

// terminalDeadLetter dead-letters the outbox row when attempt is River's
// last attempt for this job. Any error from MarkFailed itself is swallowed:
// the caller already has the original error to return to River.
func terminalDeadLetter(ctx context.Context, repo outboxMarker, fields dispatchFields, cause error, attempt, maxAttempts int) {
	if attempt < maxAttempts {
		return
	}
	_ = repo.MarkFailed(ctx, fields.TenantID, fields.OutboxID, cause.Error())
}

// PDFDispatchWorker is the River worker that dispatches one pdf staging
// outbox row as a docgen_v2_pdf event.
type PDFDispatchWorker struct {
	river.WorkerDefaults[PDFDispatchArgs]

	pub  messaging.Publisher
	repo outboxMarker
}

// NewPDFDispatchWorker constructs a PDFDispatchWorker.
func NewPDFDispatchWorker(pub messaging.Publisher, repo *fanout.StagingOutboxRepository) *PDFDispatchWorker {
	return &PDFDispatchWorker{pub: pub, repo: repo}
}

// Work publishes the pdf-convert event for the delivered job and marks the
// underlying staging outbox row dispatched (or dead-letters it on River's
// terminal attempt).
func (w *PDFDispatchWorker) Work(ctx context.Context, job *river.Job[PDFDispatchArgs]) error {
	if job == nil {
		return ErrPDFDispatchJobNil
	}
	return run(ctx, w.pub, w.repo, job.Args.dispatchFields, buildPDFEvent, job.Attempt, job.MaxAttempts)
}

// MaterializeDispatchWorker is the River worker that dispatches one
// materialize staging outbox row as a docx_materialize event.
type MaterializeDispatchWorker struct {
	river.WorkerDefaults[MaterializeDispatchArgs]

	pub  messaging.Publisher
	repo outboxMarker
}

// NewMaterializeDispatchWorker constructs a MaterializeDispatchWorker.
func NewMaterializeDispatchWorker(pub messaging.Publisher, repo *fanout.StagingOutboxRepository) *MaterializeDispatchWorker {
	return &MaterializeDispatchWorker{pub: pub, repo: repo}
}

// Work publishes the materialize-fanout event for the delivered job and
// marks the underlying staging outbox row dispatched (or dead-letters it on
// River's terminal attempt).
func (w *MaterializeDispatchWorker) Work(ctx context.Context, job *river.Job[MaterializeDispatchArgs]) error {
	if job == nil {
		return ErrMaterializeDispatchJobNil
	}
	return run(ctx, w.pub, w.repo, job.Args.dispatchFields, buildMaterializeEvent, job.Attempt, job.MaxAttempts)
}

// NewWorkers returns the River worker set for staging pdf/materialize
// dispatch, ready to register with a river.Client. Not yet wired into any
// binary — a later task adds registration + enqueue call sites.
func NewWorkers(pub messaging.Publisher, pdfRepo, materializeRepo *fanout.StagingOutboxRepository) *river.Workers {
	workers := river.NewWorkers()
	river.AddWorker(workers, NewPDFDispatchWorker(pub, pdfRepo))
	river.AddWorker(workers, NewMaterializeDispatchWorker(pub, materializeRepo))
	return workers
}
