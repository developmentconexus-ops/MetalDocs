package http_test

import (
	"database/sql"

	platformdb "metaldocs/internal/platform/db"
)

func newTxRunner(database *sql.DB) platformdb.TxRunner { return platformdb.NewTxRunner(database) }
