// Package memory provides in-process, non-persistent implementations of the
// audit domain ports (Writer/Reader/Counter, ExportJobRepository). These are
// test fixtures only — not wired by bootstrap — and carry no hash-chain or
// integrity guarantees; ValidateIntegrity always fails accordingly.
package memory

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"

	"metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/platform/db"
)

// Writer is an in-memory implementation of domain.Writer/Reader/Counter/
// IntegrityValidator, backed by a mutex-guarded slice. Suitable for unit
// tests only — see the package doc.
type Writer struct {
	mu               sync.Mutex
	events           []domain.Event
	erased           map[string]bool
	erasedCheckErr   error
	erasedCheckCalls int
}

// NewWriter constructs an empty in-memory Writer.
func NewWriter() *Writer {
	return &Writer{events: make([]domain.Event, 0, 16)}
}

// Record appends event to the in-memory store. Never returns an error.
func (w *Writer) Record(_ context.Context, event domain.Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.events = append(w.events, event)
	return nil
}

// RecordTx ignores tx (the in-memory store has no transactional semantics)
// and delegates to Record.
func (w *Writer) RecordTx(ctx context.Context, tx db.Tx, event domain.Event) error {
	return w.Record(ctx, event)
}

// ListEvents filters and sorts the in-memory events to match domain.Reader's
// contract: newest-first, with a limit+1 probe row used to compute hasMore so
// an exact-multiple last page does not falsely advertise a next page.
func (w *Writer) ListEvents(_ context.Context, query domain.ListEventsQuery) ([]domain.Event, bool, error) {
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

	// Collect one extra row past the page (limit+1 probe) to mirror the postgres
	// reader's hasMore semantics; trim the probe before returning.
	items := make([]domain.Event, 0, limit+1)
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
		// Paired checks: the in-loop break stops once the limit+1 probe is collected
		// so the scan is bounded to at most limit+1 matches, and the post-loop check
		// below does the actual trim + hasMore decision.
		if len(items) > limit {
			break
		}
	}
	if len(items) > limit {
		return items[:limit], true, nil
	}
	return items, false, nil
}

// CountEvents counts in-memory events matching query's filter (cursor/limit
// ignored, matching the postgres Counter's semantics).
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
	if !matchesIdentity(e, q) {
		return false
	}
	if !matchesAction(e, q) {
		return false
	}
	if !matchesTimeWindow(e, q) {
		return false
	}
	return matchesSearch(e, q)
}

// matchesIdentity checks the tenant/resource-type/resource-id/actor-id filters.
func matchesIdentity(e domain.Event, q domain.ListEventsQuery) bool {
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
	return true
}

// matchesAction checks the action filter, including trailing-"*" prefix matches.
func matchesAction(e domain.Event, q domain.ListEventsQuery) bool {
	act := strings.TrimSpace(q.Action)
	if act == "" {
		return true
	}
	if strings.HasSuffix(act, "*") {
		return strings.HasPrefix(e.Action, strings.TrimSuffix(act, "*"))
	}
	return strings.EqualFold(e.Action, act)
}

// matchesTimeWindow checks the OccurredAfter/OccurredBefore bounds.
func matchesTimeWindow(e domain.Event, q domain.ListEventsQuery) bool {
	if !q.OccurredAfter.IsZero() && e.OccurredAt.Before(q.OccurredAfter) {
		return false
	}
	if !q.OccurredBefore.IsZero() && !e.OccurredAt.Before(q.OccurredBefore) {
		return false
	}
	return true
}

// matchesSearch checks the free-text Query filter against action/actor/resource/payload.
func matchesSearch(e domain.Event, q domain.ListEventsQuery) bool {
	needle := strings.TrimSpace(q.Query)
	if needle == "" {
		return true
	}
	hay := strings.ToLower(strings.Join([]string{e.Action, e.ActorID, e.ResourceID, e.PayloadJSON}, " "))
	return strings.Contains(hay, strings.ToLower(needle))
}

// ValidateIntegrity always fails: the in-memory store keeps no hash chain, so
// integrity validation is not meaningful here.
func (w *Writer) ValidateIntegrity(context.Context) ([]domain.IntegrityIssue, error) {
	return nil, errors.New("integrity validation not supported by memory writer")
}

// IsErased satisfies domain.ErasureChecker (embedded in domain.Writer). No
// tenant is erased by default; tests that need to simulate an erased tenant
// or a failed lookup use SetErased / SetErasedCheckError. Same fail-closed
// contract as the real writer: a configured error is returned as-is, never
// silently treated as "not erased".
func (w *Writer) IsErased(_ context.Context, tenantID string) (bool, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.erasedCheckCalls++
	if w.erasedCheckErr != nil {
		return false, w.erasedCheckErr
	}
	return w.erased[tenantID], nil
}

// ErasedCheckCalls reports how many times IsErased has been called.
// Test-only helper — used to assert a re-check happens exactly once per
// export, not zero or repeated times.
func (w *Writer) ErasedCheckCalls() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.erasedCheckCalls
}

// SetErased marks tenantID as erased (or not) for subsequent IsErased calls.
// Test-only control hook.
func (w *Writer) SetErased(tenantID string, erased bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.erased == nil {
		w.erased = make(map[string]bool)
	}
	w.erased[tenantID] = erased
}

// SetErasedCheckError makes every subsequent IsErased call fail with err,
// simulating an erasure-status lookup failure. Test-only control hook.
func (w *Writer) SetErasedCheckError(err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.erasedCheckErr = err
}
