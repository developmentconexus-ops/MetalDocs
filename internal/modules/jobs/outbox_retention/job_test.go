package outbox_retention

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/riverqueue/river"
)

// fakePurger records what each purge arm was called with and can be made to
// fail per arm — mirrors internal/modules/render/fanout/retention.fakePurger.
type fakePurger struct {
	publishedCalls int
	deadCalls      int

	gotPublishedCutoff time.Time
	gotDeadCutoff      time.Time
	gotBatchSize       int
	gotMaxIterations   int

	publishedDeleted int
	deadDeleted      int
	publishedErr     error
	deadErr          error
}

func (f *fakePurger) PurgePublished(_ context.Context, cutoff time.Time, batchSize, maxIterations int) (int, error) {
	f.publishedCalls++
	f.gotPublishedCutoff = cutoff
	f.gotBatchSize = batchSize
	f.gotMaxIterations = maxIterations
	return f.publishedDeleted, f.publishedErr
}

func (f *fakePurger) PurgeDeadLettered(_ context.Context, cutoff time.Time, batchSize, maxIterations int) (int, error) {
	f.deadCalls++
	f.gotDeadCutoff = cutoff
	f.gotBatchSize = batchSize
	f.gotMaxIterations = maxIterations
	return f.deadDeleted, f.deadErr
}

func TestWorker_Work_PurgesBothClassesWithTheirOwnCutoffs(t *testing.T) {
	p := &fakePurger{publishedDeleted: 3, deadDeleted: 1}
	w := &Worker{retention: p}

	before := time.Now()
	err := w.Work(context.Background(), &river.Job[Args]{})
	after := time.Now()

	if err != nil {
		t.Fatalf("Work returned error: %v", err)
	}
	if p.publishedCalls != 1 || p.deadCalls != 1 {
		t.Fatalf("calls: published=%d dead=%d, want 1 and 1", p.publishedCalls, p.deadCalls)
	}
	if p.gotBatchSize != BatchSize {
		t.Errorf("batchSize = %d, want %d", p.gotBatchSize, BatchSize)
	}
	if p.gotMaxIterations != MaxIterations {
		t.Errorf("maxIterations = %d, want %d", p.gotMaxIterations, MaxIterations)
	}

	assertCutoffWithin(t, "published", p.gotPublishedCutoff, before, after, PublishedRetention)
	assertCutoffWithin(t, "dead-lettered", p.gotDeadCutoff, before, after, DeadLetterRetention)

	// The two arms must NOT share a cutoff: a dead-lettered row is forensic
	// evidence and ages out on the much longer window.
	if !p.gotDeadCutoff.Before(p.gotPublishedCutoff) {
		t.Errorf("dead-letter cutoff %v is not older than published cutoff %v; the two windows collapsed into one",
			p.gotDeadCutoff, p.gotPublishedCutoff)
	}
}

func assertCutoffWithin(t *testing.T, name string, got, before, after time.Time, window time.Duration) {
	t.Helper()
	min := before.Add(-window)
	max := after.Add(-window)
	if got.Before(min) || got.After(max) {
		t.Errorf("%s cutoff = %v, want within [%v, %v]", name, got, min, max)
	}
}

func TestWorker_Work_PublishedFails_StillRunsDeadLetterPurge(t *testing.T) {
	publishedErr := errors.New("published purge boom")
	p := &fakePurger{publishedErr: publishedErr, deadDeleted: 1}
	w := &Worker{retention: p}

	err := w.Work(context.Background(), &river.Job[Args]{})
	if !errors.Is(err, publishedErr) {
		t.Fatalf("Work error = %v, want %v", err, publishedErr)
	}
	if p.deadCalls != 1 {
		t.Errorf("dead-letter calls = %d, want 1 (must still run even though the published arm failed)", p.deadCalls)
	}
}

func TestWorker_Work_DeadLetterFails_ReturnsError(t *testing.T) {
	deadErr := errors.New("dead-letter purge boom")
	p := &fakePurger{publishedDeleted: 2, deadErr: deadErr}
	w := &Worker{retention: p}

	err := w.Work(context.Background(), &river.Job[Args]{})
	if !errors.Is(err, deadErr) {
		t.Fatalf("Work error = %v, want %v", err, deadErr)
	}
	if p.publishedCalls != 1 {
		t.Errorf("published calls = %d, want 1", p.publishedCalls)
	}
}

func TestWorker_Work_BothFail_ReturnsPublishedErrorFirst(t *testing.T) {
	publishedErr := errors.New("published purge boom")
	deadErr := errors.New("dead-letter purge boom")
	p := &fakePurger{publishedErr: publishedErr, deadErr: deadErr}
	w := &Worker{retention: p}

	err := w.Work(context.Background(), &river.Job[Args]{})
	if !errors.Is(err, publishedErr) {
		t.Fatalf("Work error = %v, want published error %v (checked first)", err, publishedErr)
	}
}

func TestArgs_KindIsStableJobName(t *testing.T) {
	// The kind is the River job identity: changing it orphans every already-
	// enqueued row of the old kind, so it is pinned here deliberately.
	if got := (Args{}).Kind(); got != "outbox-events-retention" {
		t.Errorf("Args.Kind() = %q, want %q", got, "outbox-events-retention")
	}
	if JobName != "outbox-events-retention" {
		t.Errorf("JobName = %q, want %q", JobName, "outbox-events-retention")
	}
}

func TestRetentionWindows_DeadLetterOutlivesPublished(t *testing.T) {
	if PublishedRetention != 7*24*time.Hour {
		t.Errorf("PublishedRetention = %v, want 7*24h (matches the staging outbox purge window)", PublishedRetention)
	}
	if DeadLetterRetention != 90*24*time.Hour {
		t.Errorf("DeadLetterRetention = %v, want 90*24h", DeadLetterRetention)
	}
	if DeadLetterRetention <= PublishedRetention {
		t.Errorf("DeadLetterRetention (%v) must exceed PublishedRetention (%v): dead-lettered rows are the only record that an event never reached its consumer",
			DeadLetterRetention, PublishedRetention)
	}
	if BatchSize != 5000 {
		t.Errorf("BatchSize = %d, want 5000", BatchSize)
	}
	if MaxIterations != 10 {
		t.Errorf("MaxIterations = %d, want 10", MaxIterations)
	}
}
