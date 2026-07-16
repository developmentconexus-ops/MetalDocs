//go:build integration
// +build integration

package testdb

import (
	"os"
	"strings"
	"testing"
	"time"
)

// traceEnabled gates per-phase instrumentation of the test-database factory.
// Set METALDOCS_TESTDB_TRACE=1 (and run with -v) to emit one TESTDB_PHASE line
// per phase per test. Off by default: the lines are pure noise unless the
// suite's wall-clock budget is being measured.
//
// Phase timings are what separate storage cost (the physical CREATE DATABASE
// clone) from contention (the same clone slowed down by concurrent packages).
// A single aggregate package time cannot tell those apart, and the whole
// suite-runtime budget turns on which one dominates.
var traceEnabled = strings.TrimSpace(os.Getenv("METALDOCS_TESTDB_TRACE")) != ""

// tracePhase emits a tab-separated phase record. Format is fixed and machine-
// readable on purpose — before/after comparisons are summed from the run log:
//
//	TESTDB_PHASE\t<test name>\t<phase>\t<seconds>
func tracePhase(t *testing.T, phase string, d time.Duration) {
	if !traceEnabled {
		return
	}
	t.Helper()
	t.Logf("TESTDB_PHASE\t%s\t%s\t%.4f", t.Name(), phase, d.Seconds())
}
