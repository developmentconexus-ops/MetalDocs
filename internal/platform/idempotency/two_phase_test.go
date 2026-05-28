package idempotency_test

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/testsupport/pgtest"
)

const (
	twoPhaseTenant = "00000000-0000-4000-8000-0000000000aa"
	twoPhaseActor  = "actor-two-phase"
)

// uniqueKey returns a fresh idempotency key per test so test ordering does not
// matter and the table can accumulate rows from prior runs without skewing.
func uniqueKey(prefix string) string {
	return prefix + "-" + time.Now().Format("20060102150405.000000000")
}

func TestBeginReplay_FirstCall_ReturnsHandle(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /tp/{id}")
	ctx := context.Background()

	handle, replay, err := s.BeginReplay(ctx, twoPhaseTenant, twoPhaseActor, uniqueKey("first"), "hash-a")
	if err != nil {
		t.Fatalf("BeginReplay: %v", err)
	}
	if replay != nil {
		t.Fatalf("first call should not return a replay")
	}
	if handle == nil {
		t.Fatalf("first call must return a handle")
	}
	if err := s.FailReplay(handle, nil); err != nil {
		t.Fatalf("FailReplay: %v", err)
	}
}

func TestBeginReplay_InFlightRowAllowsNullResponseColumns(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /tp/{id}")
	ctx := context.Background()
	key := uniqueKey("in-flight")

	handle, replay, err := s.BeginReplay(ctx, twoPhaseTenant, twoPhaseActor, key, "hash-a")
	if err != nil {
		t.Fatalf("BeginReplay: %v", err)
	}
	if replay != nil {
		t.Fatalf("first call should not return a replay")
	}

	var (
		status                string
		responseStatusIsNull  bool
		responseBodyIsNull    bool
		completedResponseCode int
		completedResponseBody []byte
	)
	err = db.QueryRowContext(ctx, `
		SELECT status, response_status IS NULL, response_body IS NULL
		  FROM metaldocs.idempotency_keys
		 WHERE tenant_id = $1
		   AND actor_user_id = $2
		   AND route_template = $3
		   AND key = $4`,
		twoPhaseTenant, twoPhaseActor, "POST /tp/{id}", key,
	).Scan(&status, &responseStatusIsNull, &responseBodyIsNull)
	if err != nil {
		t.Fatalf("query in_flight row: %v", err)
	}
	if status != "in_flight" || !responseStatusIsNull || !responseBodyIsNull {
		t.Fatalf("want in_flight row with null response fields, got status=%q status_null=%v body_null=%v", status, responseStatusIsNull, responseBodyIsNull)
	}

	if err := s.CompleteReplay(handle, 201, []byte(`{"ok":true}`)); err != nil {
		t.Fatalf("CompleteReplay: %v", err)
	}

	err = db.QueryRowContext(ctx, `
		SELECT status, response_status, response_body
		  FROM metaldocs.idempotency_keys
		 WHERE tenant_id = $1
		   AND actor_user_id = $2
		   AND route_template = $3
		   AND key = $4`,
		twoPhaseTenant, twoPhaseActor, "POST /tp/{id}", key,
	).Scan(&status, &completedResponseCode, &completedResponseBody)
	if err != nil {
		t.Fatalf("query completed row: %v", err)
	}
	if status != "completed" {
		t.Fatalf("want completed row, got %q", status)
	}
	if completedResponseCode != 201 {
		t.Fatalf("response_status = %d, want 201", completedResponseCode)
	}
	if !bytes.Equal(completedResponseBody, []byte(`{"ok":true}`)) {
		t.Fatalf("response_body = %q, want stored bytes", completedResponseBody)
	}
}

func TestBeginCompleteReplay_RoundTrip(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /tp/{id}")
	ctx := context.Background()
	key := uniqueKey("round")

	handle, _, err := s.BeginReplay(ctx, twoPhaseTenant, twoPhaseActor, key, "hash-a")
	if err != nil {
		t.Fatalf("Begin1: %v", err)
	}
	if err := s.CompleteReplay(handle, 201, []byte(`{"ok":true}`)); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	_, replay, err := s.BeginReplay(ctx, twoPhaseTenant, twoPhaseActor, key, "hash-a")
	if err != nil {
		t.Fatalf("Begin2: %v", err)
	}
	if replay == nil || replay.Status != 201 {
		t.Fatalf("expected cached replay, got %+v", replay)
	}
}

func TestBeginReplay_ConflictOnHashMismatch(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /tp/{id}")
	ctx := context.Background()
	key := uniqueKey("conflict")

	handle, _, err := s.BeginReplay(ctx, twoPhaseTenant, twoPhaseActor, key, "hash-a")
	if err != nil {
		t.Fatalf("Begin1: %v", err)
	}
	if err := s.CompleteReplay(handle, 200, []byte(`{}`)); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	_, _, err = s.BeginReplay(ctx, twoPhaseTenant, twoPhaseActor, key, "hash-b")
	if !errors.Is(err, idempotency.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestBeginReplay_FailReleasesSlot(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /tp/{id}")
	ctx := context.Background()
	key := uniqueKey("fail")

	handle1, _, err := s.BeginReplay(ctx, twoPhaseTenant, twoPhaseActor, key, "hash-a")
	if err != nil {
		t.Fatalf("Begin1: %v", err)
	}
	if err := s.FailReplay(handle1, errors.New("simulated handler error")); err != nil {
		t.Fatalf("FailReplay: %v", err)
	}

	handle2, replay, err := s.BeginReplay(ctx, twoPhaseTenant, twoPhaseActor, key, "hash-a")
	if err != nil {
		t.Fatalf("Begin2: %v", err)
	}
	if replay != nil {
		t.Fatalf("retry after Fail should see a fresh slot, got replay %+v", replay)
	}
	if handle2 == nil {
		t.Fatalf("retry after Fail must return a new handle")
	}
	_ = s.FailReplay(handle2, nil)
}

// TestBeginReplay_ConcurrentSameKeySameHash is the critical C3 regression test.
// Two goroutines race for the same key with the same body hash; the store
// must serialize them so the second observes the cached response, and the
// handler-side "execute" counter must increment exactly once.
func TestBeginReplay_ConcurrentSameKeySameHash(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /tp/{id}")
	ctx := context.Background()
	key := uniqueKey("concurrent-same")

	var executed int32
	var wg sync.WaitGroup
	results := make([]string, 2)

	run := func(idx int) {
		defer wg.Done()
		handle, replay, err := s.BeginReplay(ctx, twoPhaseTenant, twoPhaseActor, key, "hash-a")
		if err != nil {
			results[idx] = "err:" + err.Error()
			return
		}
		if replay != nil {
			results[idx] = "replay"
			return
		}
		atomic.AddInt32(&executed, 1)
		// Hold the slot briefly so the loser is forced to block on FOR UPDATE.
		time.Sleep(150 * time.Millisecond)
		if err := s.CompleteReplay(handle, 201, []byte(`{"once":true}`)); err != nil {
			results[idx] = "complete-err:" + err.Error()
			return
		}
		results[idx] = "owner"
	}

	wg.Add(2)
	go run(0)
	go run(1)
	wg.Wait()

	if got := atomic.LoadInt32(&executed); got != 1 {
		t.Fatalf("handler should execute exactly once; ran %d times (results=%v)", got, results)
	}
	// Exactly one owner, one replay (order non-deterministic).
	owners, replays := 0, 0
	for _, r := range results {
		switch r {
		case "owner":
			owners++
		case "replay":
			replays++
		default:
			t.Fatalf("unexpected result %q", r)
		}
	}
	if owners != 1 || replays != 1 {
		t.Fatalf("want 1 owner / 1 replay, got %d / %d (results=%v)", owners, replays, results)
	}
}

// TestBeginReplay_ConcurrentSameKeyDifferentHash proves the loser sees
// ErrConflict when the winner committed a different payload hash.
func TestBeginReplay_ConcurrentSameKeyDifferentHash(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /tp/{id}")
	ctx := context.Background()
	key := uniqueKey("concurrent-diff")

	var wg sync.WaitGroup
	type outcome struct {
		role string
		err  error
	}
	results := make([]outcome, 2)
	hashes := []string{"hash-a", "hash-b"}

	run := func(idx int) {
		defer wg.Done()
		handle, replay, err := s.BeginReplay(ctx, twoPhaseTenant, twoPhaseActor, key, hashes[idx])
		if err != nil {
			results[idx] = outcome{role: "err", err: err}
			return
		}
		if replay != nil {
			results[idx] = outcome{role: "replay"}
			return
		}
		time.Sleep(100 * time.Millisecond)
		if err := s.CompleteReplay(handle, 200, []byte(`{}`)); err != nil {
			results[idx] = outcome{role: "complete-err", err: err}
			return
		}
		results[idx] = outcome{role: "owner"}
	}

	wg.Add(2)
	go run(0)
	go run(1)
	wg.Wait()

	owners, conflicts := 0, 0
	for _, r := range results {
		switch {
		case r.role == "owner":
			owners++
		case r.role == "err" && errors.Is(r.err, idempotency.ErrConflict):
			conflicts++
		default:
			t.Fatalf("unexpected outcome %+v", r)
		}
	}
	if owners != 1 || conflicts != 1 {
		t.Fatalf("want 1 owner / 1 conflict, got %d / %d (results=%+v)", owners, conflicts, results)
	}
}

// TestCompleteReplay_RejectsCompletedRow proves the M2 belt: once a row is
// completed, a stale concurrent writer cannot overwrite payload_hash via
// CompleteReplay against a handle that was already released.
func TestCompleteReplay_DoubleCallFails(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /tp/{id}")
	ctx := context.Background()
	key := uniqueKey("double-complete")

	handle, _, err := s.BeginReplay(ctx, twoPhaseTenant, twoPhaseActor, key, "hash-a")
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if err := s.CompleteReplay(handle, 200, []byte(`{}`)); err != nil {
		t.Fatalf("first Complete: %v", err)
	}
	if err := s.CompleteReplay(handle, 200, []byte(`{"steal":true}`)); err == nil {
		t.Fatalf("second Complete on same handle should fail")
	}
}

func TestCompleteReplay_RejectsInvalidStatus(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /tp/{id}")
	ctx := context.Background()
	key := uniqueKey("invalid-status")

	handle, _, err := s.BeginReplay(ctx, twoPhaseTenant, twoPhaseActor, key, "hash-a")
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if err := s.CompleteReplay(handle, 0, []byte(`{}`)); err == nil {
		t.Fatalf("expected error for status=0")
	}
}
