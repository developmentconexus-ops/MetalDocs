package worker

import (
	"context"
	"errors"
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
