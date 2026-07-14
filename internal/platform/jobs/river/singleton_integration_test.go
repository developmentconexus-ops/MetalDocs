//go:build integration

package riverjobs_test

// M5 F5.2 T6 — P4 proof (THE load-bearing one, ADR 0067 §H-PRE-1 retirement).
//
// Before this migration, cluster-wide single-execution of the janitor jobs
// was enforced by a custom lease scheduler taking a Postgres advisory lock.
// That scheduler + its advisory lock are now deleted (metaldocs.job_leases +
// lease functions dropped). This test proves River's queue dequeue semantics
// alone — FOR UPDATE SKIP LOCKED row-claim, no advisory lock anywhere in this
// test or in production code — give the same cluster-wide singleton
// guarantee.
//
// Guarantee asserted: two independent river.Client[*sql.Tx] instances (same
// DB/schema/queue, both registered with the SAME counting worker) run
// concurrently against ONE pre-inserted job row. Only one of them can claim
// the row (FOR UPDATE SKIP LOCKED on river_job), so the worker body executes
// exactly once — never twice, never zero times. This is the "insert ONE job
// row, run both clients' queues concurrently" deterministic variant named in
// the task (chosen over the periodic-enqueuer timing variant because a fixed
// single-row assertion has no scheduler-jitter flakiness).

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/riverqueue/river"

	riverjobs "metaldocs/internal/platform/jobs/river"
	"metaldocs/tests/integration/testdb"
)

type singletonProbeArgs struct{}

func (singletonProbeArgs) Kind() string { return "h_pre_1_singleton_probe" }

// countingWorker increments a shared atomic counter once per Work() call —
// standing in for the watchdog worker body. What's under test is River's
// single-claim guarantee, not the watchdog's own logic (already covered by
// the P1 proof), so a minimal counting worker keeps this test deterministic
// and fast.
type countingWorker struct {
	river.WorkerDefaults[singletonProbeArgs]
	count *int64
	done  chan struct{}
}

func (w *countingWorker) Work(_ context.Context, _ *river.Job[singletonProbeArgs]) error {
	if atomic.AddInt64(w.count, 1) == 1 {
		close(w.done)
	}
	return nil
}

func TestSingleton_P4_HPRE1_ExactlyOnceAcrossTwoClients(t *testing.T) {
	// testdb.Open returns a private clone whose template already bakes the
	// River schema into the default schema (ApplyCuratedBootstrap ->
	// bootstrap.MigrateRiverSchema), so no per-test River migration is needed.
	db, _ := testdb.Open(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var count int64
	done := make(chan struct{})

	newBundle := func() *riverjobs.ClientBundle {
		workers := river.NewWorkers()
		river.AddWorker(workers, &countingWorker{count: &count, done: done})
		bundle, err := riverjobs.NewClientBundle(db, riverjobs.Config{
			Schema:              "",
			SkipUnknownJobCheck: true,
			Queues: map[string]river.QueueConfig{
				"maintenance": {MaxWorkers: 2},
			},
		}, workers)
		if err != nil {
			t.Fatalf("new client bundle: %v", err)
		}
		return bundle
	}

	// Two independent clients — the H-PRE-1 proof point: no advisory lock
	// coordinates them; only River's FOR UPDATE SKIP LOCKED row-claim does.
	bundleA := newBundle()
	bundleB := newBundle()

	if err := bundleA.Client.Start(ctx); err != nil {
		t.Fatalf("start client A: %v", err)
	}
	defer func() { _ = bundleA.Client.Stop(context.Background()) }()

	if err := bundleB.Client.Start(ctx); err != nil {
		t.Fatalf("start client B: %v", err)
	}
	defer func() { _ = bundleB.Client.Stop(context.Background()) }()

	// Insert exactly ONE job row. Both clients' producers poll/claim from the
	// same queue concurrently; River's claim query uses FOR UPDATE SKIP LOCKED
	// so only one producer can win the row.
	if _, err := bundleA.Client.Insert(ctx, singletonProbeArgs{}, &river.InsertOpts{
		Queue: "maintenance",
	}); err != nil {
		t.Fatalf("insert probe job: %v", err)
	}

	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("timed out waiting for the probe job to run")
	}

	// Give any (incorrect) double-claim a window to occur before asserting —
	// a second execution would race to increment count past 1.
	time.Sleep(500 * time.Millisecond)

	if got := atomic.LoadInt64(&count); got != 1 {
		t.Fatalf("worker body executed %d times across two concurrent River clients sharing one job row, want exactly 1 "+
			"(H-PRE-1 retirement: FOR UPDATE SKIP LOCKED must give single-claim with no advisory lock)", got)
	}
}
