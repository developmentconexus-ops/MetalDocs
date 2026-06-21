package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"metaldocs/internal/platform/db"
)

// CDFieldReaderPG is the Postgres-backed controlleddocumentsdomain.CDFieldReader.
// It is the single home for cross-module reads of individual
// metaldocs.controlled_documents fields (profile_code, process_area_code) —
// the table token lives here, never in a consumer module (ADR-0039 D3(b),
// M2/F2.1). The reader is stateless: every method runs its SELECT on the
// db.DB executor the caller supplies (pool off-tx, or the caller's *sql.Tx
// in-tx), so one instance serves both the off-tx (B1) and in-tx (B5, B6)
// call sites. The SELECTs are plain and non-recording (HS-PRE-1 safe).
type CDFieldReaderPG struct{}

// NewCDFieldReaderPG constructs the stateless CD field reader.
func NewCDFieldReaderPG() *CDFieldReaderPG { return &CDFieldReaderPG{} }

// ProfileCode returns controlled_documents.profile_code for (tenantID, cdID),
// or ("", nil) when the controlled document is absent in that tenant.
func (CDFieldReaderPG) ProfileCode(ctx context.Context, exec db.DB, tenantID, cdID string) (string, error) {
	var profileCode string
	err := exec.QueryRowContext(ctx,
		`SELECT profile_code FROM controlled_documents WHERE id = $1 AND tenant_id = $2`,
		cdID, tenantID,
	).Scan(&profileCode)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read controlled document profile_code: %w", err)
	}
	return profileCode, nil
}

// ProcessAreaCode returns controlled_documents.process_area_code for
// (tenantID, cdID). found is false when the controlled document is absent; the
// caller coalesces a missing row to "" in its area-resolver COALESCE chain.
func (CDFieldReaderPG) ProcessAreaCode(ctx context.Context, exec db.DB, tenantID, cdID string) (string, bool, error) {
	var areaCode sql.NullString
	err := exec.QueryRowContext(ctx,
		`SELECT process_area_code FROM controlled_documents WHERE id = $1 AND tenant_id = $2`,
		cdID, tenantID,
	).Scan(&areaCode)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("read controlled document process_area_code: %w", err)
	}
	return areaCode.String, true, nil
}
