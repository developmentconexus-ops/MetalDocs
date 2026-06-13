package application

import (
	"database/sql"

	"metaldocs/internal/platform/db"
)

// newTxRunner wraps a test *sql.DB in the db.TxRunner port so service-method
// call sites compile after the H-1d *sql.DB -> db.TxRunner refactor. Named to
// avoid the local `db` var that shadows the db package inside test builders.
func newTxRunner(database *sql.DB) db.TxRunner { return db.NewTxRunner(database) }
