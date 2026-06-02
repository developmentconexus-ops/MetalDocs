package memory

import (
	"context"
	"database/sql"
	"errors"
	"sort"
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
	snapshot := make([]domain.Event, len(w.events))
	copy(snapshot, w.events)
	w.mu.Unlock()

	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}

	sort.SliceStable(snapshot, func(i, j int) bool {
		if snapshot[i].OccurredAt.Equal(snapshot[j].OccurredAt) {
			return snapshot[i].ID > snapshot[j].ID
		}
		return snapshot[i].OccurredAt.After(snapshot[j].OccurredAt)
	})

	items := make([]domain.Event, 0, len(snapshot))
	for _, event := range snapshot {
		if !matches(event, query) {
			continue
		}
		if !query.Cursor.IsZero() {
			if event.OccurredAt.After(query.Cursor.OccurredAt) {
				continue
			}
			if event.OccurredAt.Equal(query.Cursor.OccurredAt) && event.ID >= query.Cursor.ID {
				continue
			}
		}
		items = append(items, event)
		if len(items) >= limit {
			break
		}
	}
	return items, nil
}

func (w *Writer) CountEvents(_ context.Context, query domain.ListEventsQuery) (int64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	var n int64
	for _, e := range w.events {
		if matches(e, query) {
			n++
		}
	}
	return n, nil
}

func matches(e domain.Event, q domain.ListEventsQuery) bool {
	if q.TenantID != "" && e.TenantID != q.TenantID {
		return false
	}
	if rt := strings.TrimSpace(q.ResourceType); rt != "" && !strings.EqualFold(e.ResourceType, rt) {
		return false
	}
	if rid := strings.TrimSpace(q.ResourceID); rid != "" && !strings.EqualFold(e.ResourceID, rid) {
		return false
	}
	if a := strings.TrimSpace(q.ActorID); a != "" && !strings.EqualFold(e.ActorID, a) {
		return false
	}
	if act := strings.TrimSpace(q.Action); act != "" {
		if strings.HasSuffix(act, "*") {
			if !strings.HasPrefix(e.Action, strings.TrimSuffix(act, "*")) {
				return false
			}
		} else if !strings.EqualFold(e.Action, act) {
			return false
		}
	}
	if !q.OccurredAfter.IsZero() && e.OccurredAt.Before(q.OccurredAfter) {
		return false
	}
	if !q.OccurredBefore.IsZero() && !e.OccurredAt.Before(q.OccurredBefore) {
		return false
	}
	if needle := strings.TrimSpace(q.Query); needle != "" {
		hay := strings.ToLower(strings.Join([]string{e.Action, e.ActorID, e.ResourceID, e.PayloadJSON}, " "))
		if !strings.Contains(hay, strings.ToLower(needle)) {
			return false
		}
	}
	return true
}

func (w *Writer) ValidateIntegrity(context.Context) ([]domain.IntegrityIssue, error) {
	return nil, errors.New("integrity validation not supported by memory writer")
}
