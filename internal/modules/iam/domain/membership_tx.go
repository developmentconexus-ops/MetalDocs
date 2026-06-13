package domain

import "metaldocs/internal/platform/db"

// MembershipTx is the transaction handle services receive from the repository
// for in-tx mutation + governance writes (REQ-ASYNC-1). It embeds db.Tx so the
// governance logger can write audit rows inside the same database transaction.
// *sql.Tx satisfies this interface without an adapter.
type MembershipTx interface {
	db.Tx
	Commit() error
	Rollback() error
}
