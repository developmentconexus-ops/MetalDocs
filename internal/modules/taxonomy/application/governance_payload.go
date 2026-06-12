package application

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"metaldocs/internal/modules/taxonomy/domain"
)

func marshalGovernancePayload(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal governance payload: %w", err)
	}
	return b, nil
}

// sqlTxFromFamilyTx extracts the underlying *sql.Tx from a FamilyTx.
// All taxonomy infrastructure adapters implement Unwrap() *sql.Tx; test
// doubles that do not are silently skipped (LogTx falls back to Log in the
// adapter if tx is nil, but that path should not occur in production).
func sqlTxFromFamilyTx(tx domain.FamilyTx) *sql.Tx {
	type unwrapper interface{ Unwrap() *sql.Tx }
	if u, ok := tx.(unwrapper); ok {
		return u.Unwrap()
	}
	return nil
}
