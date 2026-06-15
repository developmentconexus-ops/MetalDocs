//go:build integration
// +build integration

package repository_test

import (
	"context"
	"database/sql"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/documents/repository"
	"metaldocs/tests/integration/testdb"
)

func TestGovernedRevisionTitleColumnExistsAndCanBeRead(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenant := testdb.NewTenant(t, db)
	doc := testdb.NewDocument(t, db, testdb.WithTenant(tenant.ID), testdb.WithStatus("draft"))

	testdb.SeedWithCaps(t, db, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`UPDATE public.documents SET revision_title = $2 WHERE id = $1::uuid`,
			doc.ID, "Correcao de procedimento",
		)
		return err
	})

	var title sql.NullString
	if err := db.QueryRowContext(ctx,
		`SELECT revision_title FROM public.documents WHERE id = $1::uuid`,
		doc.ID,
	).Scan(&title); err != nil {
		t.Fatalf("scan revision_title: %v", err)
	}

	if !title.Valid || title.String != "Correcao de procedimento" {
		t.Fatalf("revision_title = %#v", title)
	}
}

func TestListRevisionHistory_ReturnsGovernedDocumentsNotAutosaveRows(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenant := testdb.NewTenant(t, db)
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID))

	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenant.ID),
		testdb.WithTaxonomy(tax),
		testdb.WithCode("QA-OPS-001"),
		testdb.WithOwner(owner.ID),
	)

	firstDoc := testdb.NewDocument(t, db,
		testdb.WithTenant(tenant.ID),
		testdb.WithControlledDoc(cd),
		testdb.WithRevisionNumber(0),
		testdb.WithStatus("published"),
	)

	secondDoc := testdb.NewDocument(t, db,
		testdb.WithTenant(tenant.ID),
		testdb.WithControlledDoc(cd),
		testdb.WithRevisionNumber(1),
		testdb.WithStatus("draft"),
	)

	testdb.SeedWithCaps(t, db, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`UPDATE public.documents
			    SET revision_title = $2,
			        created_at = '2026-05-10T12:00:00Z'::timestamptz,
			        updated_at = '2026-05-10T12:00:00Z'::timestamptz
			  WHERE id = $1::uuid`,
			firstDoc.ID,
			"Primeira versao",
		)
		if err != nil {
			return err
		}

		for _, nextStatus := range []string{"under_review", "approved"} {
			if _, err := tx.ExecContext(ctx,
				`UPDATE public.documents SET status = $2 WHERE id = $1::uuid`,
				firstDoc.ID,
				nextStatus,
			); err != nil {
				return err
			}
		}

		_, err = tx.ExecContext(ctx,
			`UPDATE public.documents
			    SET revision_title = $2,
			        created_at = '2026-05-18T12:00:00Z'::timestamptz,
			        updated_at = '2026-05-18T12:00:00Z'::timestamptz
			  WHERE id = $1::uuid`,
			secondDoc.ID,
			"Ajuste operacional",
		)
		return err
	})

	for i := 0; i < 3; i++ {
		testdb.SeedWithCaps(t, db, `[{"cap":"document.create"}]`, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx,
				`INSERT INTO public.document_revisions
				    (document_id, parent_revision_id, session_id, storage_key, content_hash, form_data_snapshot)
				 VALUES ($1::uuid, NULL, (
				     SELECT active_session_id FROM public.documents WHERE id = $1::uuid
				 ), $2, $3, '{}'::jsonb)`,
				secondDoc.ID,
				"autosave/"+testdb.DeterministicID(t, "storage"),
				testdb.DeterministicID(t, "hash-"+string(rune('a'+i))),
			)
			return err
		})
	}

	repo := repository.New(db, iamdomain.NoopUserDisplayNameReader{})
	items, err := repo.ListRevisionHistory(ctx, tenant.ID, secondDoc.ID)
	if err != nil {
		t.Fatalf("ListRevisionHistory: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].DocumentID != secondDoc.ID || items[0].RevisionTitle != "Ajuste operacional" || !items[0].IsCurrent {
		t.Fatalf("items[0] = %#v", items[0])
	}
	if items[0].RevisionNumber != 1 {
		t.Fatalf("items[0].RevisionNumber = %d, want 1", items[0].RevisionNumber)
	}
	if items[1].DocumentID != firstDoc.ID || items[1].RevisionTitle != "Primeira versao" || items[1].IsCurrent {
		t.Fatalf("items[1] = %#v", items[1])
	}
	if items[1].RevisionNumber != 0 {
		t.Fatalf("items[1].RevisionNumber = %d, want 0", items[1].RevisionNumber)
	}
}
