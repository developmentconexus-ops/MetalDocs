package application_test

import (
	"database/sql"

	"metaldocs/internal/platform/db"
)

func newTxRunner(database *sql.DB) db.TxRunner { return db.NewTxRunner(database) }
