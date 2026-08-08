package main

import "testing"

// A check that cannot run must not be reported as a hole in the claim when
// --require-infra is set. In CI, "not verified" and "verified" must not both
// exit 0, because the gate cannot tell them apart.
func TestRequireInfraTurnsSkipIntoFail(t *testing.T) {
	t.Setenv("METALDOCS_DATABASE_URL", "")

	needsDB := Check{
		ID:    "fixture-needs-postgres",
		Desc:  "fixture",
		Argv:  []string{"go", "version"},
		Needs: []string{needsPostgres},
	}

	got := run([]Check{needsDB}, 1, false, false)
	if len(got) != 1 || got[0].status != statusSkip {
		t.Fatalf("without --require-infra: want one SKIP, got %+v", got)
	}

	got = run([]Check{needsDB}, 1, false, true)
	if len(got) != 1 || got[0].status != statusFail {
		t.Fatalf("with --require-infra: want one FAIL, got %+v", got)
	}
	if got[0].reason == "" {
		t.Error("the FAIL must carry the same reason the SKIP carried; an unexplained FAIL is worse than a SKIP")
	}
	if report(got, ProfileFull) != 1 {
		t.Error("report() must exit non-zero when infra is required and missing")
	}
}

// The absence of the flag must not change existing behaviour.
func TestRunnableCheckPassesUnderRequireInfra(t *testing.T) {
	ok := Check{ID: "fixture-runnable", Desc: "fixture", Argv: []string{"go", "version"}}
	got := run([]Check{ok}, 1, false, true)
	if len(got) != 1 || got[0].status != statusPass {
		t.Fatalf("want one PASS, got %+v", got)
	}
}
