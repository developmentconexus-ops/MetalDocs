package analyzers_test

import (
	"os"
	"path/filepath"
	"testing"

	"metaldocs/tools/cilint/internal/analyzers"
)

// nosqlTxInDomainFixture writes a Go source file at the given relative path
// inside a temp directory so path-based filtering works correctly.
func nosqlTxInDomainFixture(t *testing.T, relPath, src string) string {
	t.Helper()
	dir := t.TempDir()
	full := filepath.Join(dir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(src), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return full
}

// ─── NoSQLTxInDomain ─────────────────────────────────────────────────────────

func TestNoSQLTxInDomain_Positive_SqlTx(t *testing.T) {
	src := `package domain
import "database/sql"
type Repository interface {
	LogTx(ctx interface{}, tx *sql.Tx, msg string) error
}
`
	path := nosqlTxInDomainFixture(t, "internal/modules/foo/domain/port.go", src)
	findings := analyzers.NoSQLTxInDomain([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected a finding for *sql.Tx in a domain port")
	}
}

func TestNoSQLTxInDomain_Positive_ExecutionTypes(t *testing.T) {
	// A locally-defined DBTX-style interface using sql.Result/*sql.Rows is the
	// exact leak Fix B removed from controlleddocuments — it must flag.
	src := `package domain
import "database/sql"
type DBTX interface {
	ExecContext(ctx interface{}, q string, args ...any) (sql.Result, error)
	QueryContext(ctx interface{}, q string, args ...any) (*sql.Rows, error)
}
`
	path := nosqlTxInDomainFixture(t, "internal/modules/bar/domain/seq.go", src)
	findings := analyzers.NoSQLTxInDomain([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected findings for sql.Result / *sql.Rows in a domain interface")
	}
}

func TestNoSQLTxInDomain_Negative_NullableHelpers(t *testing.T) {
	// sql.Null* nullable helpers are a legitimate domain row/DTO idiom — the
	// rule locks the transaction seam, not all of database/sql (this is the
	// auth/domain/session_admin.go case).
	src := `package domain
import "database/sql"
type SessionListItem struct {
	CreatedAt sql.NullTime
	Label     sql.NullString
}
`
	path := nosqlTxInDomainFixture(t, "internal/modules/auth/domain/session_admin.go", src)
	findings := analyzers.NoSQLTxInDomain([]string{path})
	if len(findings) != 0 {
		t.Fatalf("sql.Null* helpers must not be flagged, got %d: %+v", len(findings), findings)
	}
}

func TestNoSQLTxInDomain_Negative_DbTxNoDatabaseSql(t *testing.T) {
	src := `package domain
import "metaldocs/internal/platform/db"
type Repository interface {
	LogTx(ctx interface{}, tx db.Tx, msg string) error
}
`
	path := nosqlTxInDomainFixture(t, "internal/modules/foo/domain/port.go", src)
	findings := analyzers.NoSQLTxInDomain([]string{path})
	if len(findings) != 0 {
		t.Fatalf("expected no findings for db.Tx-only domain file, got %d: %+v", len(findings), findings)
	}
}

func TestNoSQLTxInDomain_OutOfScope_Infrastructure(t *testing.T) {
	src := `package postgres
import "database/sql"
func (r *Repo) Save(tx *sql.Tx) error { return nil }
`
	path := nosqlTxInDomainFixture(t, "internal/modules/foo/infrastructure/repo.go", src)
	findings := analyzers.NoSQLTxInDomain([]string{path})
	if len(findings) != 0 {
		t.Fatalf("infrastructure layer is out of scope, got %d findings: %+v", len(findings), findings)
	}
}

func TestNoSQLTxInDomain_Negative_AllowDirective(t *testing.T) {
	src := `package domain
import "database/sql" //cilint:allow-sqltx-in-domain
type Repository interface {
	LogTx(ctx interface{}, tx *sql.Tx, msg string) error //cilint:allow-sqltx-in-domain
}
`
	path := nosqlTxInDomainFixture(t, "internal/modules/foo/domain/port.go", src)
	findings := analyzers.NoSQLTxInDomain([]string{path})
	if len(findings) != 0 {
		t.Fatalf("expected no findings with allow directives, got %d: %+v", len(findings), findings)
	}
}
