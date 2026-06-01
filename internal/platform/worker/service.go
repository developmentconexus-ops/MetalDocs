package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/messaging"
)

var errUnsupportedEventType = errors.New("worker: unsupported event type")

type Service struct {
	consumer          messaging.Consumer
	pdfRunner         *PDFJobRunner
	materializeRunner *MaterializeJobRunner
	cfg               config.WorkerConfig
}

func NewService(consumer messaging.Consumer, cfg config.WorkerConfig) *Service {
	return &Service{
		consumer: consumer,
		cfg:      cfg,
	}
}

// WithPDFRunner attaches a PDFJobRunner that handles docgen_v2_pdf events.
func (s *Service) WithPDFRunner(r *PDFJobRunner) *Service {
	s.pdfRunner = r
	return s
}

// WithMaterializeRunner attaches a MaterializeJobRunner that handles docx_materialize events (ADR 0015).
func (s *Service) WithMaterializeRunner(r *MaterializeJobRunner) *Service {
	s.materializeRunner = r
	return s
}

func (s *Service) RunOnce(ctx context.Context, batchSize int) error {
	if s.consumer == nil {
		return fmt.Errorf("worker dependencies not configured")
	}

	start := time.Now()
	events, err := s.consumer.ClaimUnpublished(ctx, batchSize)
	if err != nil {
		return err
	}

	processed := 0
	failed := 0
	deadLettered := 0
	var runErr error
	for _, event := range events {
		var handleErr error
		switch event.EventType {
		case messaging.EventTypePDFConvert:
			if s.pdfRunner != nil {
				handleErr = s.pdfRunner.Handle(ctx, event)
			}
		case messaging.EventTypeMaterializeFanout:
			if s.materializeRunner != nil {
				handleErr = s.materializeRunner.Handle(ctx, event)
			}
		default:
			handleErr = fmt.Errorf("%w: %s", errUnsupportedEventType, event.EventType)
		}
		if handleErr != nil {
			failed++
			markedDLQ, markErr := s.markFailure(ctx, event, handleErr)
			if markErr != nil {
				runErr = errors.Join(runErr, fmt.Errorf("worker: mark failed %s: %w", event.EventID, markErr))
				continue
			}
			if markedDLQ {
				deadLettered++
			}
			continue
		}

		if err := s.consumer.MarkPublished(ctx, []messaging.EventID{event.EventID}); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("worker: mark published %s: %w", event.EventID, err))
			continue
		}
		processed++
		log.Printf("worker_event event_id=%s event_type=%s attempt_count=%d result=published trace_id=%s",
			event.EventID, event.EventType, event.AttemptCount, event.TraceID)
	}

	log.Printf("worker_batch result=completed processed=%d failed=%d dead_lettered=%d duration_ms=%d",
		processed, failed, deadLettered, time.Since(start).Milliseconds())
	return runErr
}

func (s *Service) markFailure(ctx context.Context, event messaging.Event, handleErr error) (bool, error) {
	now := time.Now().UTC()
	attempt := event.AttemptCount
	if attempt < 1 {
		attempt = 1
	}
	if errors.Is(handleErr, errUnsupportedEventType) {
		attempt = s.cfg.MaxAttempts
	}

	failure := messaging.FailedEvent{
		EventID:   event.EventID,
		LastError: truncateError(handleErr),
	}

	if attempt >= s.cfg.MaxAttempts {
		failure.DeadLetteredAt = &now
		log.Printf("worker_event event_id=%s event_type=%s attempt_count=%d result=dead_lettered trace_id=%s error=%q",
			event.EventID, event.EventType, attempt, event.TraceID, failure.LastError)
	} else {
		nextAttempt := now.Add(backoffDuration(attempt, s.cfg.RetryBaseSeconds, s.cfg.RetryMaxSeconds))
		failure.NextAttemptAt = &nextAttempt
		log.Printf("worker_event event_id=%s event_type=%s attempt_count=%d result=retry_scheduled trace_id=%s next_attempt_at=%s error=%q",
			event.EventID, event.EventType, attempt, event.TraceID, nextAttempt.Format(time.RFC3339), failure.LastError)
	}

	if err := s.consumer.MarkFailed(ctx, failure); err != nil {
		return false, err
	}
	return failure.DeadLetteredAt != nil, nil
}

func backoffDuration(attempt, baseSeconds, maxSeconds int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if baseSeconds < 1 {
		baseSeconds = 10
	}
	if maxSeconds < baseSeconds {
		maxSeconds = baseSeconds
	}

	delaySeconds := int64(baseSeconds)
	maxDelaySeconds := int64(maxSeconds)
	for i := 1; i < attempt && delaySeconds < maxDelaySeconds; i++ {
		if delaySeconds > maxDelaySeconds/2 {
			delaySeconds = maxDelaySeconds
			break
		}
		delaySeconds *= 2
	}
	return time.Duration(delaySeconds) * time.Second
}

func truncateError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if len(message) > 512 {
		return message[:512]
	}
	return message
}
