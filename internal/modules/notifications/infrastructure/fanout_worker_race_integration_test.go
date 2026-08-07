//go:build integration
// +build integration

package notificationsinfra_test

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	documentsdomain "metaldocs/internal/modules/documents/domain"
	notificationsinfra "metaldocs/internal/modules/notifications/infrastructure"
	"metaldocs/tests/integration/testdb"
)

// notificationRow is the projection used by the race test to compare the
// resulting notification row set across interleaving orders.
type notificationRow struct {
	recipientUserID string
	sourceEventID   string
	eventType       string
}

// TestNotificationsFanoutWorker_ConcurrentRaceCommutativity is F5.5 T1: proves
// that two lifecycle events for the SAME controlled document, fanned out by two
// REAL goroutines racing against the SAME testdb, commute — the resulting
// notification row set does not depend on which tx commits first, and
// redelivery after the race stays idempotent.
//
// Each Work() call is single-tenant-scoped (args.TenantID) and only performs
// additive per-recipient INSERTs guarded by the partial unique index on
// (recipient_user_id, source_event_id) — there is no shared mutable
// projection row for the two events to contend over, so this test expects the
// commutativity to hold structurally rather than needing new locking.
func TestNotificationsFanoutWorker_ConcurrentRaceCommutativity(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	worker := notificationsinfra.NewNotificationsFanoutWorker(db)

	seedUserGrant := func(t *testing.T, tenantID, cdID, userID string) {
		t.Helper()
		if _, err := db.ExecContext(ctx,
			`INSERT INTO public.controlled_document_user_grants
			   (tenant_id, controlled_document_id, user_id)
			 VALUES ($1::uuid, $2::uuid, $3)`,
			tenantID, cdID, userID,
		); err != nil {
			t.Fatalf("seed user grant: %v", err)
		}
	}

	fetchRows := func(t *testing.T, tenantID, cdResourceID string, eventIDs []string) []notificationRow {
		t.Helper()
		rows, err := db.QueryContext(ctx,
			`SELECT recipient_user_id, source_event_id, event_type
			   FROM metaldocs.notifications
			  WHERE tenant_id = $1::uuid AND source_event_id = ANY($2::uuid[])`,
			tenantID, pgUUIDArray(eventIDs),
		)
		if err != nil {
			t.Fatalf("fetch rows: %v", err)
		}
		defer rows.Close()
		var out []notificationRow
		for rows.Next() {
			var r notificationRow
			if err := rows.Scan(&r.recipientUserID, &r.sourceEventID, &r.eventType); err != nil {
				t.Fatalf("scan row: %v", err)
			}
			out = append(out, r)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows err: %v", err)
		}
		sort.Slice(out, func(i, j int) bool {
			if out[i].recipientUserID != out[j].recipientUserID {
				return out[i].recipientUserID < out[j].recipientUserID
			}
			if out[i].sourceEventID != out[j].sourceEventID {
				return out[i].sourceEventID < out[j].sourceEventID
			}
			return out[i].eventType < out[j].eventType
		})
		return out
	}

	// runRace seeds a fresh tenant/CD/readers scenario and races Work(A) and
	// Work(B) as two real goroutines released simultaneously via a start
	// gate (closed channel), so both goroutines' BeginTx calls fire at
	// genuinely the same instant rather than being serialized by the test
	// itself. giveBHeadStart adds a tiny controlled delay before releasing
	// goroutine A, deterministically flipping which of the two txs is more
	// likely to reach BeginTx/commit first across the two calls of this
	// helper — exercising both interleaving orders instead of relying on
	// unpredictable scheduler variance alone.
	runRace := func(t *testing.T, giveBHeadStart bool) (tenantID string, readerIDs []string, eventIDA, eventIDB string, rows []notificationRow, worker2 func() error) {
		t.Helper()
		ten := testdb.NewTenant(t, db)
		cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(ten.ID))
		r1 := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		r2 := testdb.NewUser(t, db, testdb.WithTenant(ten.ID))
		seedUserGrant(t, ten.ID, cd.ID, r1.ID)
		seedUserGrant(t, ten.ID, cd.ID, r2.ID)

		eventA := uuid.NewString()
		eventB := uuid.NewString()

		argsA := documentsdomain.LifecycleEventArgs{
			EventID: eventA, TenantID: ten.ID,
			EventType:    documentsdomain.EventTypeDocumentPublished,
			ResourceType: "document", ResourceID: uuid.NewString(),
			ControlledDocumentID: cd.ID, OccurredAt: time.Now(),
		}
		argsB := documentsdomain.LifecycleEventArgs{
			EventID: eventB, TenantID: ten.ID,
			EventType:    documentsdomain.EventTypeDocumentSuperseded,
			ResourceType: "document", ResourceID: uuid.NewString(),
			ControlledDocumentID: cd.ID, OccurredAt: time.Now(),
		}

		jobA := &river.Job[documentsdomain.LifecycleEventArgs]{Args: argsA}
		jobB := &river.Job[documentsdomain.LifecycleEventArgs]{Args: argsB}

		start := make(chan struct{})
		var wg sync.WaitGroup
		errs := make(chan error, 2)

		wg.Add(2)
		go func() {
			defer wg.Done()
			if giveBHeadStart {
				<-start
				time.Sleep(2 * time.Millisecond)
			} else {
				<-start
			}
			errs <- worker.Work(ctx, jobA)
		}()
		go func() {
			defer wg.Done()
			if giveBHeadStart {
				<-start
			} else {
				<-start
				time.Sleep(2 * time.Millisecond)
			}
			errs <- worker.Work(ctx, jobB)
		}()

		close(start) // release both goroutines simultaneously
		wg.Wait()
		close(errs)

		for err := range errs {
			if err != nil {
				t.Fatalf("Work (race): %v", err)
			}
		}

		got := fetchRows(t, ten.ID, cd.ID, []string{eventA, eventB})
		return ten.ID, []string{r1.ID, r2.ID}, eventA, eventB, got, func() error { return worker.Work(ctx, jobB) }
	}

	var firstRunRows []notificationRow
	var firstTenant string
	var firstEventA, firstEventB string
	var redeliver func() error

	t.Run("order_A_head_start", func(t *testing.T) {
		tenantID, readers, eventA, eventB, rows, redo := runRace(t, false)
		firstTenant, firstEventA, firstEventB, firstRunRows, redeliver = tenantID, eventA, eventB, rows, redo

		if len(rows) != len(readers)*2 {
			t.Fatalf("want %d notification rows (readers=%d x 2 events), got %d", len(readers)*2, len(readers), len(rows))
		}
		assertPerReaderRows(t, rows, readers, eventA, eventB)
	})

	t.Run("order_B_head_start", func(t *testing.T) {
		_, readers, eventA, eventB, rows, _ := runRace(t, true)

		if len(rows) != len(readers)*2 {
			t.Fatalf("want %d notification rows (readers=%d x 2 events), got %d", len(readers)*2, len(readers), len(rows))
		}
		assertPerReaderRows(t, rows, readers, eventA, eventB)

		// Row SET shape (event-type/reader-count structure) must be identical
		// across both interleaving orders — same cardinality, same per-reader
		// per-event coverage, regardless of which tx committed first. Tenants
		// and event IDs differ per run (fresh fixtures), so we compare the
		// normalized shape: sorted (event_type) multiset per reader-count.
		if len(rows) != len(firstRunRows) {
			t.Fatalf("row set size differs across interleavings: order_A=%d order_B=%d", len(firstRunRows), len(rows))
		}
		typeCounts := func(rs []notificationRow) map[string]int {
			m := map[string]int{}
			for _, r := range rs {
				m[r.eventType]++
			}
			return m
		}
		firstCounts := typeCounts(firstRunRows)
		secondCounts := typeCounts(rows)
		for k, v := range firstCounts {
			if secondCounts[k] != v {
				t.Errorf("event_type %q count differs across interleavings: order_A=%d order_B=%d", k, v, secondCounts[k])
			}
		}
	})

	t.Run("redelivery_after_race_is_noop", func(t *testing.T) {
		if redeliver == nil {
			t.Fatal("redeliver not set — order_A_head_start subtest must run first")
		}
		if err := redeliver(); err != nil {
			t.Fatalf("Work (redelivery after race): %v", err)
		}
		rows := fetchRows(t, firstTenant, "", []string{firstEventA, firstEventB})
		if len(rows) != len(firstRunRows) {
			t.Fatalf("redelivery changed row count: before=%d after=%d", len(firstRunRows), len(rows))
		}
	})
}

func assertPerReaderRows(t *testing.T, rows []notificationRow, readers []string, eventA, eventB string) {
	t.Helper()
	byReader := map[string]map[string]string{} // reader -> sourceEventID -> eventType
	for _, r := range rows {
		if byReader[r.recipientUserID] == nil {
			byReader[r.recipientUserID] = map[string]string{}
		}
		byReader[r.recipientUserID][r.sourceEventID] = r.eventType
	}
	for _, readerID := range readers {
		m, ok := byReader[readerID]
		if !ok {
			t.Errorf("reader %s: missing all notification rows", readerID)
			continue
		}
		if _, ok := m[eventA]; !ok {
			t.Errorf("reader %s: missing row for event A (%s)", readerID, eventA)
		}
		if _, ok := m[eventB]; !ok {
			t.Errorf("reader %s: missing row for event B (%s)", readerID, eventB)
		}
		if len(m) != 2 {
			t.Errorf("reader %s: want exactly 2 distinct source_event_id rows, got %d", readerID, len(m))
		}
	}
}

// pgUUIDArray formats a Go string slice as a Postgres uuid[] literal argument
// usable with ANY($n::uuid[]) — avoids pulling in pq/pgtype array helpers for
// this one query.
func pgUUIDArray(ids []string) string {
	s := "{"
	for i, id := range ids {
		if i > 0 {
			s += ","
		}
		s += id
	}
	s += "}"
	return s
}
