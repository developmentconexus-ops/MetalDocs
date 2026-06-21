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

// F2.1 / B6 parity gate (D6: parity-before-delete). Proves the ported
// loadInstanceAreaCode (own approval/documents read + CDFieldReader.ProcessAreaCode
// for the CD fallback, COALESCE precedence resolved in Go, tx-aware) returns the
// EXACT result the deleted JOIN query produced —
//
//	SELECT COALESCE(asi.area_code_snapshot, d.process_area_code_snapshot, cd.process_area_code, '')
//	  FROM approval_instances ai
//	  JOIN documents d ON d.id = ai.document_id AND d.tenant_id = ai.tenant_id
//	  LEFT JOIN controlled_documents cd ON cd.id = d.controlled_document_id
//	  LEFT JOIN approval_stage_instances asi
//	    ON asi.approval_instance_id = ai.id AND asi.status = 'active'
//	 WHERE ai.id = $1 AND ai.tenant_id = $2 LIMIT 1
//
// across the precedence chain: active-stage snapshot wins (even when ""), then
// document snapshot wins (even when ""), then the CD area, then "".

func rawInstanceAreaCode(t *testing.T, db *sql.DB, tenantID, instanceID string) (string, bool) {
	t.Helper()
	var area string
	err := db.QueryRowContext(context.Background(), `
		SELECT COALESCE(asi.area_code_snapshot, d.process_area_code_snapshot, cd.process_area_code, '')
		  FROM approval_instances ai
		  JOIN documents d
		    ON d.id = ai.document_id
		   AND d.tenant_id = ai.tenant_id
		  LEFT JOIN controlled_documents cd
		    ON cd.id = d.controlled_document_id
		  LEFT JOIN approval_stage_instances asi
		    ON asi.approval_instance_id = ai.id
		   AND asi.status = 'active'
		 WHERE ai.id = $1
		   AND ai.tenant_id = $2
		 LIMIT 1`,
		instanceID, tenantID,
	).Scan(&area)
	if err == sql.ErrNoRows {
		return "", false
	}
	if err != nil {
		t.Fatalf("raw instance area query failed: %v", err)
	}
	return area, true
}

func b6SnapshotArg(s sql.NullString) any {
	if !s.Valid {
		return nil
	}
	return s.String
}

// seedInstanceAreaChain seeds tenant-scoped CD → document → approval_instance and,
// when stage.Valid, an ACTIVE approval_stage_instance carrying area_code_snapshot.
// A fresh CD per call satisfies ux_documents_cd_active; the CD also supplies the
// route profile. When link is false the document is NOT linked to the CD (its
// controlled_document_id is NULL), exercising the no-CD branch.
func seedInstanceAreaChain(t *testing.T, db *sql.DB, tenantID string, link bool, docSnap, stage sql.NullString) string {
	t.Helper()
	ctx := context.Background()
	cd := testdb.NewControlledDoc(t, db, testdb.WithTenant(tenantID))
	route := testdb.NewApprovalRoute(t, db, testdb.WithTenant(tenantID), testdb.WithProfile(cd.ProfileCode))
	owner := cd.OwnerUserID

	docID := uuid.NewString()
	var cdArg any
	if link {
		cdArg = cd.ID
	}
	testdb.SeedWithCaps(t, db, `[{"cap":"document.create"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO public.documents
			   (id, tenant_id, template_version_id, name, status, form_data_json,
			    created_by, controlled_document_id, revision_number, process_area_code_snapshot)
			 VALUES ($1::uuid, $2::uuid, $3::uuid, $4, 'draft', '{}'::jsonb,
			    $5, $6::uuid, 0, $7)`,
			docID, tenantID, uuid.NewString(), "B6 Doc", owner, cdArg, b6SnapshotArg(docSnap),
		)
		return err
	})

	instID := uuid.NewString()
	testdb.SeedWithCaps(t, db, `[{"cap":"document.submit"}]`, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.approval_instances
			   (id, tenant_id, document_id, route_id, route_version_snapshot, status,
			    submitted_by, submitted_at, content_hash_at_submit, idempotency_key)
			 VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, 1, 'in_progress',
			    $5, now(), repeat('a', 64), $6)`,
			instID, tenantID, docID, route.ID, owner, "idem-"+instID,
		); err != nil {
			return err
		}
		if stage.Valid {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO public.approval_stage_instances
				   (id, approval_instance_id, stage_order, name_snapshot,
				    required_role_snapshot, required_capability_snapshot, area_code_snapshot,
				    quorum_snapshot, on_eligibility_drift_snapshot, eligible_actor_ids, status)
				 VALUES ($1::uuid, $2::uuid, 1, 'S',
				    'reviewer', 'doc.signoff', $3,
				    'any_1_of', 'keep_snapshot', '[]'::jsonb, 'active')`,
				uuid.NewString(), instID, stage.String,
			); err != nil {
				return err
			}
		}
		return nil
	})
	return instID
}

func TestLoadInstanceAreaCode_ParityPrePostPort(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	cdReader := cdinfra.NewCDFieldReaderPG()
	tenant := testdb.NewTenant(t, db)

	valid := func(s string) sql.NullString { return sql.NullString{String: s, Valid: true} }
	null := sql.NullString{Valid: false}

	cases := []struct {
		name    string
		link    bool
		docSnap sql.NullString
		stage   sql.NullString
	}{
		{"stage_snapshot_wins", true, valid("DOC"), valid("STAGE")},
		{"empty_stage_snapshot_wins", true, valid("DOC"), valid("")},
		{"doc_snapshot_wins_no_active_stage", true, valid("DOCAREA"), null},
		{"empty_doc_snapshot_wins", true, valid(""), null},
		{"cd_area_fallback", true, null, null},
		{"unlinked_all_null", false, null, null},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			instID := seedInstanceAreaChain(t, db, tenant.ID, tc.link, tc.docSnap, tc.stage)

			wantArea, wantFound := rawInstanceAreaCode(t, db, tenant.ID, instID)

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatalf("begin tx: %v", err)
			}
			defer tx.Rollback()

			gotArea, gotFound, err := loadInstanceAreaCode(ctx, tx, cdReader, tenant.ID, instID)
			if err != nil {
				t.Fatalf("loadInstanceAreaCode error = %v", err)
			}
			if gotArea != wantArea || gotFound != wantFound {
				t.Fatalf("parity break: port=(%q,%v) raw=(%q,%v)", gotArea, gotFound, wantArea, wantFound)
			}
			if tc.name == "cd_area_fallback" && wantArea == "" {
				t.Fatal("cd_area_fallback expected a non-empty CD area to exercise the port; seeded CD area was empty")
			}
		})
	}

	// Absent instance: both paths report not-found.
	absent := uuid.NewString()
	wantArea, wantFound := rawInstanceAreaCode(t, db, tenant.ID, absent)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()
	gotArea, gotFound, err := loadInstanceAreaCode(ctx, tx, cdReader, tenant.ID, absent)
	if err != nil {
		t.Fatalf("loadInstanceAreaCode(absent) error = %v", err)
	}
	if gotArea != wantArea || gotFound != wantFound || gotFound {
		t.Fatalf("parity break (absent): port=(%q,%v) raw=(%q,%v) want found=false", gotArea, gotFound, wantArea, wantFound)
	}
}
