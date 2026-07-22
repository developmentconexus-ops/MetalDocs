package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/messaging"
	"metaldocs/internal/platform/servicebus"
)

type fakeConsumer struct {
	events              []messaging.Event
	published           []messaging.EventID
	failures            []messaging.FailedEvent
	markPublishedErrors map[messaging.EventID]error
	markFailedErrors    map[messaging.EventID]error
}

func (f *fakeConsumer) ClaimUnpublished(_ context.Context, _ int) ([]messaging.Event, error) {
	return f.events, nil
}

func (f *fakeConsumer) MarkPublished(_ context.Context, eventIDs []messaging.EventID) error {
	f.published = append(f.published, eventIDs...)
	if len(eventIDs) == 1 && f.markPublishedErrors != nil {
		if err := f.markPublishedErrors[eventIDs[0]]; err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeConsumer) MarkFailed(_ context.Context, failure messaging.FailedEvent) error {
	f.failures = append(f.failures, failure)
	if f.markFailedErrors != nil {
		if err := f.markFailedErrors[failure.EventID]; err != nil {
			return err
		}
	}
	return nil
}

func TestWorkerService_RoutesPDFEventToPDFRunner(t *testing.T) {
	consumer := &fakeConsumer{events: []messaging.Event{
		{
			EventID:   "e1",
			EventType: messaging.EventTypePDFConvert,
			Payload: messaging.PDFConvertPayload{
				TenantID:       "t1",
				RevisionID:     "r1",
				FinalDocxS3Key: "tenants/t1/revisions/r1/frozen.docx",
			},
		},
	}}

	converter := &fakePDFConverter{result: servicebus.ConvertPDFResult{
		OutputKey:   "tenants/t1/revisions/r1/final.pdf",
		ContentHash: "deadbeef",
	}}
	persister := &fakePDFPersister{}
	runner := NewPDFJobRunner(converter, persister)

	cfg := config.WorkerConfig{MaxAttempts: 3, RetryBaseSeconds: 10, RetryMaxSeconds: 300}
	svc := NewService(consumer, cfg).WithPDFRunner(runner)

	if err := svc.RunOnce(context.Background(), 10); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if len(persister.calls) != 1 {
		t.Errorf("PDFJobRunner.Handle called %d times, want 1", len(persister.calls))
	}
}

func TestWorkerService_RunOnceContinuesAfterPersistenceErrors(t *testing.T) {
	consumer := &fakeConsumer{
		events: []messaging.Event{
			{
				EventID:      "publish-fails",
				EventType:    messaging.EventTypePDFConvert,
				AttemptCount: 1,
				Payload: messaging.PDFConvertPayload{
					TenantID:       "t1",
					RevisionID:     "r1",
					FinalDocxS3Key: "tenants/t1/revisions/r1/frozen.docx",
				},
			},
			{
				EventID:      "mark-failed-fails",
				EventType:    messaging.EventTypePDFConvert,
				AttemptCount: 1,
				Payload: messaging.PDFConvertPayload{
					TenantID:       "t1",
					RevisionID:     "r2",
					FinalDocxS3Key: "tenants/t1/revisions/r2/frozen.docx",
				},
			},
			{
				EventID:      "publish-succeeds",
				EventType:    messaging.EventTypePDFConvert,
				AttemptCount: 1,
				Payload: messaging.PDFConvertPayload{
					TenantID:       "t1",
					RevisionID:     "r3",
					FinalDocxS3Key: "tenants/t1/revisions/r3/frozen.docx",
				},
			},
		},
		markPublishedErrors: map[messaging.EventID]error{
			"publish-fails": errors.New("mark published failed"),
		},
		markFailedErrors: map[messaging.EventID]error{
			"mark-failed-fails": errors.New("mark failed persistence failed"),
		},
	}

	converter := &fakePDFConverter{
		result: servicebus.ConvertPDFResult{
			OutputKey:   "tenants/t1/revisions/r1/final.pdf",
			ContentHash: strings.Repeat("c", 64),
		},
	}
	persister := &fakePDFPersister{}
	cfg := config.WorkerConfig{MaxAttempts: 3, RetryBaseSeconds: 10, RetryMaxSeconds: 300}
	svc := NewService(consumer, cfg).WithPDFRunner(&PDFJobRunner{
		converter: failingSecondCallConverter{delegate: converter},
		persister: persister,
	})

	err := svc.RunOnce(context.Background(), 10)
	if err == nil {
		t.Fatal("expected aggregated persistence error")
	}
	if !strings.Contains(err.Error(), "publish-fails") || !strings.Contains(err.Error(), "mark-failed-fails") {
		t.Fatalf("expected both persistence errors, got %v", err)
	}
	if len(consumer.published) != 2 {
		t.Fatalf("published calls = %d, want 2", len(consumer.published))
	}
	if consumer.published[1] != "publish-succeeds" {
		t.Fatalf("last published event = %s, want publish-succeeds", consumer.published[1])
	}
	if len(consumer.failures) != 1 || consumer.failures[0].EventID != "mark-failed-fails" {
		t.Fatalf("failures = %#v", consumer.failures)
	}
	if len(persister.calls) != 2 {
		t.Fatalf("WritePDF calls = %d, want 2 successful conversions", len(persister.calls))
	}
}

func TestBackoffDurationClampsWithoutOverflow(t *testing.T) {
	delay := backoffDuration(128, 10, 300)
	if delay != 300*time.Second {
		t.Fatalf("delay = %v, want %v", delay, 300*time.Second)
	}
}

func TestWorkerService_UnknownEventDeadLettersInsteadOfPublishing(t *testing.T) {
	consumer := &fakeConsumer{
		events: []messaging.Event{
			{
				EventID:      "unknown-event",
				EventType:    messaging.EventType("future_event"),
				AttemptCount: 1,
				TraceID:      "trace-1",
			},
		},
	}

	cfg := config.WorkerConfig{MaxAttempts: 5, RetryBaseSeconds: 10, RetryMaxSeconds: 300}
	svc := NewService(consumer, cfg)

	if err := svc.RunOnce(context.Background(), 10); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if len(consumer.published) != 0 {
		t.Fatalf("published = %#v, want none", consumer.published)
	}
	if len(consumer.failures) != 1 {
		t.Fatalf("failures = %#v, want one dead-lettered failure", consumer.failures)
	}
	if consumer.failures[0].DeadLetteredAt == nil {
		t.Fatalf("failure = %#v, want DeadLetteredAt to be set", consumer.failures[0])
	}
	if !strings.Contains(consumer.failures[0].LastError, "unsupported event type") {
		t.Fatalf("LastError = %q, want unsupported event type", consumer.failures[0].LastError)
	}
}

// fakeClassifiedError mirrors the shape of fanout.RenderError (Status/Kind +
// Retryable()) without importing the render module — platform/worker must
// stay decoupled from module internals (see markFailure's structural match).
type fakeClassifiedError struct {
	kind      string
	retryable bool
}

func (e *fakeClassifiedError) Error() string   { return "render failed (" + e.kind + ")" }
func (e *fakeClassifiedError) Retryable() bool { return e.retryable }

func TestWorkerService_NonRetryableRenderError_DeadLettersOnFirstAttempt(t *testing.T) {
	consumer := &fakeConsumer{
		events: []messaging.Event{
			{
				EventID:      "materialize-permanent",
				EventType:    messaging.EventTypeMaterializeFanout,
				AttemptCount: 1,
				TraceID:      "trace-permanent",
				Payload: messaging.MaterializeFanoutPayload{
					TenantID:   "t1",
					RevisionID: "r1",
				},
			},
		},
	}

	// Simulate the real wrapping chain: fanout.Client returns *RenderError,
	// FreezeService.Materialize wraps with "materialize: fanout: %w", and
	// MaterializeJobRunner.Handle wraps again with "materialize job runner: %w".
	permanent := &fakeClassifiedError{kind: "template_parse", retryable: false}
	wrapped := fmt.Errorf("materialize job runner: %w", fmt.Errorf("materialize: fanout: %w", permanent))

	// markFailure is the classification seam under test (RunOnce's handler
	// dispatch to *MaterializeJobRunner is exercised separately in
	// materialize_job_runner_test.go via the concrete-type dependency).
	cfg := config.WorkerConfig{MaxAttempts: 5, RetryBaseSeconds: 10, RetryMaxSeconds: 300}
	svc := NewService(consumer, cfg)
	dlq, err := svc.markFailure(context.Background(), consumer.events[0], wrapped)
	if err != nil {
		t.Fatalf("markFailure: %v", err)
	}
	if !dlq {
		t.Fatal("expected non-retryable classified error to dead-letter on first attempt")
	}
	if len(consumer.failures) != 1 {
		t.Fatalf("failures = %#v, want 1", consumer.failures)
	}
	failure := consumer.failures[0]
	if failure.DeadLetteredAt == nil {
		t.Fatal("expected DeadLetteredAt to be set on first attempt for a non-retryable error")
	}
	if failure.NextAttemptAt != nil {
		t.Fatalf("expected no backoff scheduling for a non-retryable error, got NextAttemptAt=%v", *failure.NextAttemptAt)
	}
	if !strings.Contains(failure.LastError, "template_parse") {
		t.Fatalf("LastError = %q, want it to name the defect class", failure.LastError)
	}
}

func TestWorkerService_RetryableRenderError_SchedulesBackoffLikeBefore(t *testing.T) {
	consumer := &fakeConsumer{
		events: []messaging.Event{
			{
				EventID:      "materialize-transient",
				EventType:    messaging.EventTypeMaterializeFanout,
				AttemptCount: 1,
				TraceID:      "trace-transient",
				Payload: messaging.MaterializeFanoutPayload{
					TenantID:   "t1",
					RevisionID: "r1",
				},
			},
		},
	}

	transient := &fakeClassifiedError{kind: "renderer_unavailable", retryable: true}
	wrapped := fmt.Errorf("materialize job runner: %w", fmt.Errorf("materialize: fanout: %w", transient))

	cfg := config.WorkerConfig{MaxAttempts: 5, RetryBaseSeconds: 10, RetryMaxSeconds: 300}
	svc := NewService(consumer, cfg)
	dlq, err := svc.markFailure(context.Background(), consumer.events[0], wrapped)
	if err != nil {
		t.Fatalf("markFailure: %v", err)
	}
	if dlq {
		t.Fatal("expected a retryable classified error not to dead-letter on first attempt")
	}
	if len(consumer.failures) != 1 {
		t.Fatalf("failures = %#v, want 1", consumer.failures)
	}
	failure := consumer.failures[0]
	if failure.DeadLetteredAt != nil {
		t.Fatal("expected no DeadLetteredAt for a retryable error on first attempt")
	}
	if failure.NextAttemptAt == nil {
		t.Fatal("expected NextAttemptAt to be scheduled (existing backoff behavior)")
	}
}

// TestWorkerService_NilPDFRunner_FailsLoud pins F-QA2-1: a PDFConvert event
// arriving while no pdfRunner is configured must FAIL LOUD (mark the event
// failed → retry/DLQ), never be silently marked published with no PDF produced.
func TestWorkerService_NilPDFRunner_FailsLoud(t *testing.T) {
	consumer := &fakeConsumer{events: []messaging.Event{
		{
			EventID:      "pdf-no-runner",
			EventType:    messaging.EventTypePDFConvert,
			AttemptCount: 1,
			Payload: messaging.PDFConvertPayload{
				TenantID:       "t1",
				RevisionID:     "r1",
				FinalDocxS3Key: "tenants/t1/revisions/r1/frozen.docx",
			},
		},
	}}

	cfg := config.WorkerConfig{MaxAttempts: 5, RetryBaseSeconds: 10, RetryMaxSeconds: 300}
	svc := NewService(consumer, cfg) // no WithPDFRunner — pdfRunner is nil

	if err := svc.RunOnce(context.Background(), 10); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if len(consumer.published) != 0 {
		t.Fatalf("published = %#v, want none (nil runner must not mark published)", consumer.published)
	}
	if len(consumer.failures) != 1 {
		t.Fatalf("failures = %#v, want one failure", consumer.failures)
	}
	if !strings.Contains(consumer.failures[0].LastError, "pdf runner not configured") {
		t.Fatalf("LastError = %q, want it to name the missing pdf runner", consumer.failures[0].LastError)
	}
}

// TestWorkerService_NilMaterializeRunner_FailsLoud pins F-QA2-1 for the
// materialize arm of the same switch.
func TestWorkerService_NilMaterializeRunner_FailsLoud(t *testing.T) {
	consumer := &fakeConsumer{events: []messaging.Event{
		{
			EventID:      "materialize-no-runner",
			EventType:    messaging.EventTypeMaterializeFanout,
			AttemptCount: 1,
			Payload: messaging.MaterializeFanoutPayload{
				TenantID:   "t1",
				RevisionID: "r1",
			},
		},
	}}

	cfg := config.WorkerConfig{MaxAttempts: 5, RetryBaseSeconds: 10, RetryMaxSeconds: 300}
	svc := NewService(consumer, cfg) // no WithMaterializeRunner — materializeRunner is nil

	if err := svc.RunOnce(context.Background(), 10); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if len(consumer.published) != 0 {
		t.Fatalf("published = %#v, want none (nil runner must not mark published)", consumer.published)
	}
	if len(consumer.failures) != 1 {
		t.Fatalf("failures = %#v, want one failure", consumer.failures)
	}
	if !strings.Contains(consumer.failures[0].LastError, "materialize runner not configured") {
		t.Fatalf("LastError = %q, want it to name the missing materialize runner", consumer.failures[0].LastError)
	}
}

type failingSecondCallConverter struct {
	delegate *fakePDFConverter
}

func (f failingSecondCallConverter) ConvertPDF(ctx context.Context, req servicebus.ConvertPDFRequest) (servicebus.ConvertPDFResult, error) {
	if f.delegate.calls == 1 {
		f.delegate.calls++
		return servicebus.ConvertPDFResult{}, errors.New("convert failed")
	}
	return f.delegate.ConvertPDF(ctx, req)
}
