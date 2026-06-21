//go:build integration

package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"

	"metaldocs/tests/integration/testdb"
)

// F2.1 / M2 Category-B parity gate (D6: parity-before-delete). These tests pin
// CDFieldReaderPG — the single owner-side home for cross-module reads of
// controlled_documents fields — to the exact result the raw cross-module SELECTs
// produce at the consumer sites being ported (B1 profile_code; B5/B6
// process_area_code). The raw read MUST NOT be deleted until the matching parity
// test here is green.

// rawProfileCode runs the literal B1 query
// (documents/repository/repository.go:1701) and returns "" on no-row, matching
// that call site's sql.ErrNoRows tolerance.
func rawProfileCode(t *testing.T, db *sql.DB, tenantID, cdID string) string {
	t.Helper()
	var pc string
	err := db.QueryRowContext(context.Background(),
		`SELECT profile_code FROM controlled_documents WHERE id = $1 AND tenant_id = $2`,
		cdID, tenantID,
	).Scan(&pc)
	if errors.Is(err, sql.ErrNoRows) {
		return ""
	}
	if err != nil {
		t.Fatalf("raw profile_code query failed: %v", err)
	}
	return pc
}

// rawProcessAreaCode runs the cross-module term of the B5/B6 COALESCE
// (cd.process_area_code) and reports presence, matching the LEFT JOIN fallback:
// missing CD row ⇒ found=false ⇒ caller coalesces to "".
func rawProcessAreaCode(t *testing.T, db *sql.DB, tenantID, cdID string) (string, bool) {
	t.Helper()
	var ac sql.NullString
	err := db.QueryRowContext(context.Background(),
		`SELECT process_area_code FROM controlled_documents WHERE id = $1 AND tenant_id = $2`,
		cdID, tenantID,
	).Scan(&ac)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false
	}
	if err != nil {
		t.Fatalf("raw process_area_code query failed: %v", err)
	}
	return ac.String, true
}

func TestCDFieldReader_ProfileCode_ParityWithRawSQL(t *testing.T) {
	db, _ := testdb.Open(t)
	reader := NewCDFieldReaderPG()
	ctx := context.Background()

	cd := testdb.NewControlledDoc(t, db)

	// Present: port result == raw result == the seeded profile_code.
	got, err := reader.ProfileCode(ctx, db, cd.TenantID, cd.ID)
	if err != nil {
		t.Fatalf("ProfileCode(present) error = %v", err)
	}
	if raw := rawProfileCode(t, db, cd.TenantID, cd.ID); got != raw {
		t.Fatalf("parity break (present): port=%q raw=%q", got, raw)
	}
	if got != cd.ProfileCode {
		t.Fatalf("ProfileCode = %q, want seeded %q", got, cd.ProfileCode)
	}

	// Absent CD: port returns ("", nil), matching the B1 ErrNoRows-tolerant "".
	absentID := uuid.NewString()
	got, err = reader.ProfileCode(ctx, db, cd.TenantID, absentID)
	if err != nil {
		t.Fatalf("ProfileCode(absent) error = %v", err)
	}
	if raw := rawProfileCode(t, db, cd.TenantID, absentID); got != raw || got != "" {
		t.Fatalf("parity break (absent): port=%q raw=%q want \"\"", got, raw)
	}

	// Wrong tenant (tenant isolation): present id, foreign tenant ⇒ "".
	got, err = reader.ProfileCode(ctx, db, uuid.NewString(), cd.ID)
	if err != nil {
		t.Fatalf("ProfileCode(foreign tenant) error = %v", err)
	}
	if got != "" {
		t.Fatalf("ProfileCode(foreign tenant) = %q, want \"\" (tenant isolation)", got)
	}
}

func TestCDFieldReader_ProcessAreaCode_ParityWithRawSQL(t *testing.T) {
	db, _ := testdb.Open(t)
	reader := NewCDFieldReaderPG()
	ctx := context.Background()

	cd := testdb.NewControlledDoc(t, db)

	// Present: (areaCode, found=true) == raw, and areaCode == seeded value.
	gotArea, gotFound, err := reader.ProcessAreaCode(ctx, db, cd.TenantID, cd.ID)
	if err != nil {
		t.Fatalf("ProcessAreaCode(present) error = %v", err)
	}
	rawArea, rawFound := rawProcessAreaCode(t, db, cd.TenantID, cd.ID)
	if gotArea != rawArea || gotFound != rawFound {
		t.Fatalf("parity break (present): port=(%q,%v) raw=(%q,%v)", gotArea, gotFound, rawArea, rawFound)
	}
	if !gotFound || gotArea != cd.ProcessAreaCode {
		t.Fatalf("ProcessAreaCode = (%q,%v), want (%q,true)", gotArea, gotFound, cd.ProcessAreaCode)
	}

	// Absent CD: found=false (caller coalesces to ""), parity with raw.
	absentID := uuid.NewString()
	gotArea, gotFound, err = reader.ProcessAreaCode(ctx, db, cd.TenantID, absentID)
	if err != nil {
		t.Fatalf("ProcessAreaCode(absent) error = %v", err)
	}
	rawArea, rawFound = rawProcessAreaCode(t, db, cd.TenantID, absentID)
	if gotArea != rawArea || gotFound != rawFound || gotFound {
		t.Fatalf("parity break (absent): port=(%q,%v) raw=(%q,%v) want found=false", gotArea, gotFound, rawArea, rawFound)
	}
}
