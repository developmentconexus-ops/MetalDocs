package http

import (
	"database/sql"

	platformdb "metaldocs/internal/platform/db"
)

func newTxRunner(database *sql.DB) platformdb.TxRunner {
	if database == nil {
		return nil
	}
	return platformdb.NewTxRunner(database)
}
