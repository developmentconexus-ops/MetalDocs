package db_test

import (
	"database/sql"
	"testing"

	"metaldocs/internal/platform/db"
)

// Compile-time assertions: *sql.DB satisfies db.DB; *sql.Tx satisfies db.DB and db.Tx.
var (
	_ db.DB = (*sql.DB)(nil)
	_ db.DB = (*sql.Tx)(nil)
	_ db.Tx = (*sql.Tx)(nil)
)

func TestPackageCompiles(t *testing.T) {
	// Sentinel — interface assertions above are the real test.
	t.Log("db.Tx/db.DB compile asserts pass")
}
