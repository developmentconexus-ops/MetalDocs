package scheduler

import "testing"

func TestScheduler_SchedulerMetrics_GroupedByJob(t *testing.T) {
	snap := MetricsSnapshot{
		RunsTotal:   map[string]int64{"job-a": 3, "job-b": 1},
		ErrorsTotal: map[string]int64{"job-b": 1},
		SkipsTotal:  map[string]int64{"job-c": 2},
	}
	got := schedulerMetricsFromSnapshot(snap)

	jobsRaw, ok := got["jobs"]
	if !ok {
		t.Fatal("result missing 'jobs' key")
	}
	jobs, ok := jobsRaw.(map[string]any)
	if !ok {
		t.Fatalf("jobs is not map[string]any: %T", jobsRaw)
	}

	cases := []struct {
		name                string
		runs, errors, skips int64
	}{
		{"job-a", 3, 0, 0},
		{"job-b", 1, 1, 0},
		{"job-c", 0, 0, 2},
	}
	for _, tc := range cases {
		entryRaw, ok := jobs[tc.name]
		if !ok {
			t.Fatalf("jobs[%q] missing", tc.name)
		}
		entry, ok := entryRaw.(map[string]int64)
		if !ok {
			t.Fatalf("jobs[%q] is not map[string]int64: %T", tc.name, entryRaw)
		}
		if entry["runs"] != tc.runs {
			t.Errorf("jobs[%q].runs = %d; want %d", tc.name, entry["runs"], tc.runs)
		}
		if entry["errors"] != tc.errors {
			t.Errorf("jobs[%q].errors = %d; want %d", tc.name, entry["errors"], tc.errors)
		}
		if entry["skips"] != tc.skips {
			t.Errorf("jobs[%q].skips = %d; want %d", tc.name, entry["skips"], tc.skips)
		}
	}
}
