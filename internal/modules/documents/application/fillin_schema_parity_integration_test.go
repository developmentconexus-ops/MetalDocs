//go:build integration
// +build integration

package application_test

import (
	"context"
	"database/sql"
	"errors"
	"encoding/json"
	"reflect"
	"testing"

	"metaldocs/internal/modules/documents/application"
	templatesinfra "metaldocs/internal/modules/templates/infrastructure"
	templatesdomain "metaldocs/internal/modules/templates/domain"
	"metaldocs/tests/integration/testdb"
)

// F2.4 / N1 parity gate (D6: parity-before-delete). Proves the ported
// LoadFillInSchema (resolve template_version_id on the documents-owned table,
// then read placeholder_schema through templates' TemplateVersionPort) returns
// the EXACT []Placeholder the deleted cross-module JOIN produced —
//
//	SELECT tv.placeholder_schema
//	  FROM templates_template_version tv
//	  JOIN documents d ON d.template_version_id = tv.id
//	 WHERE d.id = $1::uuid AND d.tenant_id = $2::uuid
//
// across present-schema / absent-document / null-schema.

func rawFillInSchema(t *testing.T, db *sql.DB, tenantID, docID string) []templatesdomain.Placeholder {
	t.Helper()
	var pRaw []byte
	err := db.QueryRowContext(context.Background(), `
		SELECT tv.placeholder_schema
		  FROM templates_template_version tv
		  JOIN documents d ON d.template_version_id = tv.id
		 WHERE d.id = $1::uuid AND d.tenant_id = $2::uuid`,
		docID, tenantID,
	).Scan(&pRaw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		t.Fatalf("raw fill-in schema: %v", err)
	}
	var ph []templatesdomain.Placeholder
	if len(pRaw) > 0 {
		if err := json.Unmarshal(pRaw, &ph); err != nil {
			t.Fatalf("raw unmarshal: %v", err)
		}
	}
	return ph
}

// seedTemplateVersion inserts a templates_template + templates_template_version
// (template.create tripwire) with the given placeholder_schema jsonb, returning
// the version id. schemaJSON is a raw jsonb literal (e.g. an array or 'null').
func seedTemplateVersion(t *testing.T, db *sql.DB, tenantID, schemaJSON string) string {
	t.Helper()
	actorID := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID
	templateID := testdb.DeterministicID(t, "f24-template-"+schemaJSON)
	versionID := testdb.DeterministicID(t, "f24-version-"+schemaJSON)
	testdb.SeedWithCaps(t, db, `[{"cap":"template.create"}]`, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(context.Background(), `
			INSERT INTO public.templates_template (
				id, tenant_id, doc_type_code, key, name, latest_version, published_version_id, created_by
			) VALUES ($1::uuid, $2::uuid, 'po', $3, 'F24 Template', 1, NULL, $4)`,
			templateID, tenantID, "f24-key-"+versionID, actorID,
		); err != nil {
			return err
		}
		_, err := tx.ExecContext(context.Background(), `
			INSERT INTO public.templates_template_version (
				id, template_id, version_number, revision_number, status, docx_storage_key, content_hash,
				metadata_schema, placeholder_schema, author_id, published_at
			) VALUES ($1::uuid, $2::uuid, 1, 0, 'published', 'templates/f24/body.docx', 'f24-hash',
				'{}'::jsonb, `+schemaJSON+`::jsonb, $3, now())`,
			versionID, templateID, actorID,
		)
		return err
	})
	return versionID
}

func TestLoadFillInSchema_ParityWithRawJoin(t *testing.T) {
	ctx := context.Background()
	db, dbName := testdb.Open(t)
	// Bare templates_*/documents references must resolve to public.*; mirror the
	// templates reader integration test's search_path fix.
	if _, err := db.ExecContext(ctx, `ALTER DATABASE "`+dbName+`" SET search_path TO public, metaldocs`); err != nil {
		t.Fatalf("alter database search_path: %v", err)
	}
	db.SetMaxIdleConns(0)
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(4)

	reader := application.NewTemplateVersionSchemaReader(db, templatesinfra.NewTemplateVersionReader(db))

	t.Run("present_schema", func(t *testing.T) {
		tnt := testdb.NewTenant(t, db)
		vID := seedTemplateVersion(t, db, tnt.ID, `'[{"id":"field_a","label":"Field A","type":"text","required":true}]'`)
		doc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithTemplateVersionID(vID), testdb.WithStatus("draft"))

		want := rawFillInSchema(t, db, tnt.ID, doc.ID)
		got, err := reader.LoadFillInSchema(ctx, tnt.ID, doc.ID)
		if err != nil {
			t.Fatalf("LoadFillInSchema: %v", err)
		}
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("present_schema: raw=%#v port=%#v", want, got)
		}
		if len(got) != 1 || got[0].ID != "field_a" {
			t.Fatalf("present_schema: expected 1 placeholder field_a, got %#v", got)
		}
	})

	t.Run("absent_document", func(t *testing.T) {
		tnt := testdb.NewTenant(t, db)
		absentDocID := testdb.DeterministicID(t, "f24-absent-doc")
		want := rawFillInSchema(t, db, tnt.ID, absentDocID)
		got, err := reader.LoadFillInSchema(ctx, tnt.ID, absentDocID)
		if err != nil {
			t.Fatalf("LoadFillInSchema: %v", err)
		}
		if !reflect.DeepEqual(want, got) || got != nil {
			t.Fatalf("absent_document: expected nil==nil, raw=%#v port=%#v", want, got)
		}
	})

	t.Run("null_schema", func(t *testing.T) {
		tnt := testdb.NewTenant(t, db)
		vID := seedTemplateVersion(t, db, tnt.ID, `'null'`)
		doc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithTemplateVersionID(vID), testdb.WithStatus("draft"))

		want := rawFillInSchema(t, db, tnt.ID, doc.ID)
		got, err := reader.LoadFillInSchema(ctx, tnt.ID, doc.ID)
		if err != nil {
			t.Fatalf("LoadFillInSchema: %v", err)
		}
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("null_schema: raw=%#v port=%#v", want, got)
		}
	})
}
