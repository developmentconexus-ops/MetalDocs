package memory

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"

	"metaldocs/internal/modules/audit/domain"
)

type Writer struct {
	mu     sync.Mutex
	events []domain.Event
}

func NewWriter() *Writer {
	return &Writer{events: make([]domain.Event, 0, 16)}
}

func (w *Writer) Record(_ context.Context, event domain.Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.events = append(w.events, event)
	return nil
}

func (w *Writer) RecordTx(ctx context.Context, tx *sql.Tx, event domain.Event) error {
	return w.Record(ctx, event)
}

func (w *Writer) ListEvents(_ context.Context, query domain.ListEventsQuery) ([]domain.Event, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}

	resourceType := strings.TrimSpace(query.ResourceType)
	resourceID := strings.TrimSpace(query.ResourceID)
	items := make([]domain.Event, 0, len(w.events))
	for i := len(w.events) - 1; i >= 0; i-- {
		event := w.events[i]
		if resourceType != "" && !strings.EqualFold(event.ResourceType, resourceType) {
			continue
		}
		if resourceID != "" && !strings.EqualFold(event.ResourceID, resourceID) {
			continue
		}
		if query.TenantID != "" && event.TenantID != query.TenantID {
			continue
		}
		items = append(items, event)
		if len(items) >= limit {
			break
		}
	}
	return items, nil
}

func (w *Writer) ValidateIntegrity(context.Context) ([]domain.IntegrityIssue, error) {
	return nil, errors.New("integrity validation not supported by memory writer")
}
