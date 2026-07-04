package retention

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/riverqueue/river"
)

// fakePurger records the cutoff/batchSize/maxIterations it was called with
// and can be made to fail.
type fakePurger struct {
	calls int

	gotCutoff        time.Time
	gotBatchSize     int
	gotMaxIterations int

	deleted int
	err     error
}

func (f *fakePurger) PurgeDispatched(_ context.Context, cutoff time.Time, batchSize, maxIterations int) (int, error) {
	f.calls++
	f.gotCutoff = cutoff
	f.gotBatchSize = batchSize
	f.gotMaxIterations = maxIterations
	return f.deleted, f.err
}

func TestPurgeWorker_Work_BothSucceed(t *testing.T) {
	pdf := &fakePurger{deleted: 3}
	mat := &fakePurger{deleted: 7}
	w := &PurgeWorker{pdfRepo: pdf, materializeRepo: mat}

	before := time.Now()
	err := w.Work(context.Background(), &river.Job[PurgeArgs]{})
	after := time.Now()

	if err != nil {
		t.Fatalf("Work returned error: %v", err)
	}
	if pdf.calls != 1 || mat.calls != 1 {
		t.Fatalf("calls: pdf=%d mat=%d, want 1 and 1", pdf.calls, mat.calls)
	}

	for name, f := range map[string]*fakePurger{"pdf": pdf, "materialize": mat} {
		if f.gotBatchSize != BatchSize {
			t.Errorf("%s batchSize = %d, want %d", name, f.gotBatchSize, BatchSize)
		}
		if f.gotMaxIterations != MaxIterations {
			t.Errorf("%s maxIterations = %d, want %d", name, f.gotMaxIterations, MaxIterations)
		}
		wantCutoff := before.Add(-RetentionPeriod)
		maxCutoff := after.Add(-RetentionPeriod)
		if f.gotCutoff.Before(wantCutoff) || f.gotCutoff.After(maxCutoff) {
			t.Errorf("%s cutoff = %v, want within [%v, %v]", name, f.gotCutoff, wantCutoff, maxCutoff)
		}
	}
}

func TestPurgeWorker_Work_PDFFails_ReturnsError_StillRunsMaterialize(t *testing.T) {
	pdfErr := errors.New("pdf purge boom")
	pdf := &fakePurger{err: pdfErr}
	mat := &fakePurger{deleted: 1}
	w := &PurgeWorker{pdfRepo: pdf, materializeRepo: mat}

	err := w.Work(context.Background(), &river.Job[PurgeArgs]{})
	if !errors.Is(err, pdfErr) {
		t.Fatalf("Work error = %v, want %v", err, pdfErr)
	}
	if mat.calls != 1 {
		t.Errorf("materialize calls = %d, want 1 (must still run even though pdf failed)", mat.calls)
	}
}

func TestPurgeWorker_Work_MaterializeFails_ReturnsError(t *testing.T) {
	matErr := errors.New("materialize purge boom")
	pdf := &fakePurger{deleted: 2}
	mat := &fakePurger{err: matErr}
	w := &PurgeWorker{pdfRepo: pdf, materializeRepo: mat}

	err := w.Work(context.Background(), &river.Job[PurgeArgs]{})
	if !errors.Is(err, matErr) {
		t.Fatalf("Work error = %v, want %v", err, matErr)
	}
	if pdf.calls != 1 {
		t.Errorf("pdf calls = %d, want 1", pdf.calls)
	}
}

func TestPurgeWorker_Work_BothFail_ReturnsPDFErrorFirst(t *testing.T) {
	pdfErr := errors.New("pdf purge boom")
	matErr := errors.New("materialize purge boom")
	pdf := &fakePurger{err: pdfErr}
	mat := &fakePurger{err: matErr}
	w := &PurgeWorker{pdfRepo: pdf, materializeRepo: mat}

	err := w.Work(context.Background(), &river.Job[PurgeArgs]{})
	if !errors.Is(err, pdfErr) {
		t.Fatalf("Work error = %v, want pdf error %v (checked first)", err, pdfErr)
	}
}

func TestConstants_MatchBindingContractValues(t *testing.T) {
	if RetentionPeriod != 7*24*time.Hour {
		t.Errorf("RetentionPeriod = %v, want 7*24h", RetentionPeriod)
	}
	if BatchSize != 5000 {
		t.Errorf("BatchSize = %d, want 5000", BatchSize)
	}
	if MaxIterations != 10 {
		t.Errorf("MaxIterations = %d, want 10", MaxIterations)
	}
}
