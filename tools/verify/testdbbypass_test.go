package main

import (
	"strings"
	"testing"
)

const bypassingFixture = `package postgres_test

import (
	"database/sql"
	"os"
	"testing"
)

func TestBypassesFactory(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	_ = db
}
`

const testdbFixture = `package postgres_test

import (
	"testing"

	"metaldocs/tests/integration/testdb"
)

func TestUsesFactory(t *testing.T) {
	db := testdb.Open(t)
	_ = db
}
`

const commentOnlyFixture = `package postgres_test

import (
	"database/sql"
	"testing"
)

// This test used to read DATABASE_URL directly; it no longer does — see
// testdb.Open below.
func TestCommentMentionsDatabaseURLOnly(t *testing.T) {
	db, err := sql.Open("pgx", "postgres://localhost/testdb")
	if err != nil {
		t.Fatal(err)
	}
	_ = db
}
`

func TestFileBypassesTestdb_FlagsRawDSNPlusSQLOpen(t *testing.T) {
	line, bypass, err := fileBypassesTestdb("bypassing_test.go", []byte(bypassingFixture))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !bypass {
		t.Fatal("want bypass=true for a file that reads DATABASE_URL and calls sql.Open, got false")
	}
	if line != 11 {
		t.Errorf("line = %d, want 11 (the sql.Open call site)", line)
	}
}

func TestFileBypassesTestdb_SilentOnTestdbFactoryUsage(t *testing.T) {
	_, bypass, err := fileBypassesTestdb("testdb_test.go", []byte(testdbFixture))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if bypass {
		t.Fatal("want bypass=false for a file that only calls testdb.Open, got true")
	}
}

func TestFileBypassesTestdb_SilentWhenDatabaseURLOnlyInComment(t *testing.T) {
	// Guards the obvious false positive: DATABASE_URL appearing in prose,
	// with an unrelated sql.Open call using a literal DSN, must not fire —
	// the AST walk only inspects string-literal tokens, never comment text.
	_, bypass, err := fileBypassesTestdb("comment_test.go", []byte(commentOnlyFixture))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if bypass {
		t.Fatal("want bypass=false when DATABASE_URL only appears in a comment, got true")
	}
}

func TestScanTestdbBypassFindings_ReportsOffenderAndSkipsAllowlisted(t *testing.T) {
	// scanTestdbBypassFindings reads real files, so this drives it against
	// this package's own fixtures-as-strings by writing them to temp files
	// is unnecessary — fileBypassesTestdb is already proven pure above.
	// This test instead proves the composition: exclusion of the testdb
	// framework directory and of allowlisted paths happens before a file is
	// even read.
	entries := []testdbBypassEntry{
		{path: "tests/integration/scenarios/e2e_happy_test.go", reason: "allowlisted for this test"},
	}
	// Neither candidate path exists on disk relative to the test binary's
	// working directory, but both are filtered out before scanTestdbBypassFindings
	// ever calls os.ReadFile — so a missing-file error would mean the
	// exclusion did not run, not that the fixture setup is wrong.
	findings, err := scanTestdbBypassFindings(entries, []string{
		"tests/integration/testdb/does-not-matter_test.go",
		"tests/integration/scenarios/e2e_happy_test.go",
	})
	if err != nil {
		t.Fatalf("scanTestdbBypassFindings: %v (expected both paths to be excluded before any read)", err)
	}
	if len(findings) != 0 {
		t.Errorf("findings = %v, want none (excluded dir + allowlisted path)", findings)
	}
}

func TestValidateAllowlistEntries_FailsClosedOnStalePath(t *testing.T) {
	entries := []testdbBypassEntry{
		{path: "tests/integration/scenarios/renamed-or-deleted_test.go", reason: "was real once"},
	}
	findings := validateAllowlistEntries(entries, func(string) bool { return false })
	if len(findings) != 1 {
		t.Fatalf("findings = %v, want exactly 1 (fail-closed on a missing allowlist path)", findings)
	}
	if !strings.Contains(findings[0], "renamed-or-deleted_test.go") {
		t.Errorf("finding %q does not name the stale path", findings[0])
	}
}

func TestValidateAllowlistEntries_SilentWhenPathExists(t *testing.T) {
	entries := []testdbBypassEntry{
		{path: "tests/integration/scenarios/e2e_happy_test.go", reason: "real"},
	}
	findings := validateAllowlistEntries(entries, func(string) bool { return true })
	if len(findings) != 0 {
		t.Errorf("findings = %v, want none when the path exists", findings)
	}
}

func TestIsAllowlisted(t *testing.T) {
	entries := testdbBypassAllowlist
	if !isAllowlisted(entries, "tests/integration/scenarios/e2e_happy_test.go") {
		t.Error("expected the shipped allowlist entry to match its own path")
	}
	if isAllowlisted(entries, "tests/integration/scenarios/some_other_test.go") {
		t.Error("expected an unrelated path not to match")
	}
}
