//go:build integration

package application

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"

	cdinfra "metaldocs/internal/modules/controlleddocuments/infrastructure"
	"metaldocs/tests/integration/testdb"
)

// F2.1 / B5 parity gate (D6: parity-before-delete). Proves the ported
// LoadDocumentAreaCode (own documents read + CDFieldReader.ProcessAreaCode
// COALESCE in Go, tx-aware) returns the EXACT result the deleted raw query
// produced —
//
//	SELECT COALESCE(d.process_area_code_snapshot, cd.process_area_code, '')
//	  FROM documents d LEFT JOIN controlled_documents cd
//	    ON cd.id = d.controlled_document_id AND cd.tenant_id = d.tenant_id
//	 WHERE d.id = $1 AND d.tenant_id = $2
//
// across the snapshot-precedence cases: non-NULL snapshot wins even when "",
// NULL snapshot falls through to the CD area, absent/unlinked CD ⇒ "".

// rawDocumentAreaCode runs the pre-port JOIN query (the term being deleted) so
// the test pins the port to the literal prior behavior, not a re-derivation.
func rawDocumentAreaCode(t *testing.T, db *sql.DB, tenantID, documentID string) (string, bool) {
	t.Helper()
	var area string
	err := db.QueryRowContext(context.Background(), `
		SELECT COALESCE(d.process_area_code_snapshot, cd.process_area_code, '')
		  FROM documents d
		  LEFT JOIN controlled_documents cd
		    ON cd.id = d.controlled_document_id AND cd.tenant_id = d.tenant_id
		 WHERE d.id = $1 AND d.tenant_id = $2`,
		documentID, tenantID,
	).Scan(&area)
	if err == sql.ErrNoRows {
		return "", false
	}
	if err != nil {
		t.Fatalf("raw document area query failed: %v", err)
	}
	return area, true
}

// seedDocWithArea inserts a draft documents row (document.create tripwire) with
// an explicit process_area_code_snapshot (NULL when snapshot.Valid is false) and
// controlled_document_id (NULL when cdID == ""), bypassing the higher-level
// builders so each COALESCE branch can be constructed directly.
func seedDocWithArea(t *testing.T, db *sql.DB, tenantID, createdBy, cdID string, snapshot sql.NullString) string {
	t.Helper()
	docID := uuid.NewString()
	var cdArg any
	if cdID != "" {
		cdArg = cdID
	}
	testdb.SeedWithCaps(t, db, `[{"cap":"document.create"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO public.documents
			   (id, tenant_id, template_version_id, name, status, form_data_json,
			    created_by, controlled_document_id, revision_number, process_area_code_snapshot)
			 VALUES ($1::uuid, $2::uuid, $3::uuid, $4, 'draft', '{}'::jsonb,
			    $5, $6::uuid, 0, $7)`,
			docID, tenantID, uuid.NewString(), "Parity Doc", createdBy, cdArg, snapshotArg(snapshot),
		)
		return err
	})
	return docID
}

func snapshotArg(s sql.NullString) any {
	if !s.Valid {
		return nil
	}
	return s.String
}

func TestLoadDocumentAreaCode_ParityPrePostPort(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	cdReader := cdinfra.NewCDFieldReaderPG()

	// A shared tenant; each linked case gets its OWN controlled document because
	// ux_documents_cd_active permits only one active document per CD.
	tenant := testdb.NewTenant(t, db)

	cases := []struct {
		name     string
		linked   bool // seed + link a fresh CD (non-empty area) for this doc
		snapshot sql.NullString
	}{
		{"nonnull_snapshot_wins_over_cd", true, sql.NullString{String: "snap-area-x", Valid: true}},
		{"empty_snapshot_wins_over_cd", true, sql.NullString{String: "", Valid: true}},
		{"null_snapshot_falls_through_to_cd", true, sql.NullString{Valid: false}},
		{"null_snapshot_no_cd_link", false, sql.NullString{Valid: false}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cdID := ""
			owner := ""
			if tc.linked {
				cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(tenant.ID))
				if cd.ProcessAreaCode == "" {
					t.Fatal("seeded CD has empty process_area_code; linked parity cases need a non-empty CD area")
				}
				cdID = cd.ID
				owner = cd.OwnerUserID
			} else {
				owner = testdb.NewUser(t, db, testdb.WithTenant(tenant.ID)).ID
			}

			docID := seedDocWithArea(t, db, tenant.ID, owner, cdID, tc.snapshot)

			wantArea, wantFound := rawDocumentAreaCode(t, db, tenant.ID, docID)

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatalf("begin tx: %v", err)
			}
			defer tx.Rollback()

			gotArea, gotFound, err := LoadDocumentAreaCode(ctx, tx, cdReader, tenant.ID, docID)
			if err != nil {
				t.Fatalf("LoadDocumentAreaCode error = %v", err)
			}
			if gotArea != wantArea || gotFound != wantFound {
				t.Fatalf("parity break: port=(%q,%v) raw=(%q,%v)", gotArea, gotFound, wantArea, wantFound)
			}
		})
	}

	// Absent document: both paths report not-found.
	absent := uuid.NewString()
	wantArea, wantFound := rawDocumentAreaCode(t, db, tenant.ID, absent)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()
	gotArea, gotFound, err := LoadDocumentAreaCode(ctx, tx, cdReader, tenant.ID, absent)
	if err != nil {
		t.Fatalf("LoadDocumentAreaCode(absent) error = %v", err)
	}
	if gotArea != wantArea || gotFound != wantFound || gotFound {
		t.Fatalf("parity break (absent): port=(%q,%v) raw=(%q,%v) want found=false", gotArea, gotFound, wantArea, wantFound)
	}
}
