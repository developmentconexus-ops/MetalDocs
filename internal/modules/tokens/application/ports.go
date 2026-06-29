package application

import (
	"context"
	"database/sql"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/platform/db"
)

// txRunner is the platform tx boundary. Re-declared minimally so the service
// depends on the behaviour it needs, not the whole platform package surface.
type txRunner interface {
	Do(ctx context.Context, fn func(tx *sql.Tx) error) error
	DoReadOnly(ctx context.Context, fn func(tx *sql.Tx) error) error
}

// auditRecorder records a state change inside the business tx. tx is typed
// db.Tx (audit's real signature) — the *sql.Tx from txRunner satisfies it
// structurally and is passed through with no cast.
type auditRecorder interface {
	RecordTx(ctx context.Context, tx db.Tx, event auditdomain.Event) error
}

// ReservedNames is the consumer port the service uses to reject dictionary entry
// names that collide with native/computed token names. The composition root backs
// it with the render resolver registry's known keys, so tokens stays a leaf — no
// tokens -> render import (SP-2 §5.1, §11).
type ReservedNames interface {
	IsReserved(name string) bool
}
