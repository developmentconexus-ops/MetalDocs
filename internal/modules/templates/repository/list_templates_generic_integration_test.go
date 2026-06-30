//go:build integration
// +build integration

package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/modules/templates/repository"
	"metaldocs/tests/integration/testdb"
)

// TestListTemplates_ProfileFilterIncludesGeneric proves the document-creation
// template-picker contract: when a document is created for profile P, the
// template list filtered by doc_type=P must surface (a) templates scoped to P
// AND (b) generic templates (doc_type_code = ''). The template wizard's
// "Genérico" scope promises "Templates genéricos aparecem para todos os
// perfis", so a generic published template is a valid clone base for every
// profile. A profile-scoped template for a *different* profile must NOT appear.
//
// Regression guard: the filter predicate was previously a strict equality
// (doc_type_code = $2), which silently hid every generic template from the
// new-document wizard — leaving "Em branco" as the only selectable base even
// when a generic template had been published.
func TestListTemplates_ProfileFilterIncludesGeneric(t *testing.T) {
	ctx := context.Background()
	db, dbName := testdb.Open(t)

	// Bare runtime tables (templates_*) must resolve to public.*; evict the
	// connection that ran the ALTER so the pool reopens with the new default.
	if _, err := db.ExecContext(ctx, `ALTER DATABASE "`+dbName+`" SET search_path TO public, metaldocs`); err != nil {
		t.Fatalf("alter database search_path: %v", err)
	}
	db.SetMaxIdleConns(0)
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(4)

	tnt := testdb.NewTenant(t, db)
	actor := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID)).ID

	tplPO := testdb.DeterministicID(t, "list-generic-po")
	tplGeneric := testdb.DeterministicID(t, "list-generic-blank")
	tplDC := testdb.DeterministicID(t, "list-generic-dc")

	// Seed three templates: one scoped to 'po', one generic (''), one scoped to
	// 'dc'. The raw INSERT is cap-gated (template.create), mirroring the
	// production write path's authz context.
	testdb.SeedWithCaps(t, db, `[{"cap":"template.create"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO public.templates_template (
				id, tenant_id, doc_type_code, key, name, latest_version, published_version_id, created_by
			) VALUES
				($1::uuid, $4::uuid, 'po', 'list-generic-po',    'PO Template',      1, NULL, $5),
				($2::uuid, $4::uuid, '',   'list-generic-blank', 'Generic Template', 1, NULL, $5),
				($3::uuid, $4::uuid, 'dc', 'list-generic-dc',    'DC Template',      1, NULL, $5)`,
			tplPO, tplGeneric, tplDC, tnt.ID, actor,
		)
		return err
	})

	repo := repository.New(db)

	collectKeys := func(ts []*domain.Template) map[string]bool {
		out := map[string]bool{}
		for _, tpl := range ts {
			out[tpl.Key] = true
		}
		return out
	}

	t.Run("profile_filter_returns_profile_plus_generic", func(t *testing.T) {
		po := "po"
		got, err := repo.ListTemplates(ctx, application.ListFilter{
			TenantID:    tnt.ID,
			DocTypeCode: &po,
			Limit:       50,
		})
		if err != nil {
			t.Fatalf("ListTemplates(po): %v", err)
		}
		keys := collectKeys(got)
		if !keys["list-generic-po"] {
			t.Error("profile-scoped template (po) must appear when filtering by po")
		}
		if !keys["list-generic-blank"] {
			t.Error("generic template must appear when filtering by po (contract: genéricos aparecem para todos os perfis)")
		}
		if keys["list-generic-dc"] {
			t.Error("a different profile's template (dc) must NOT appear when filtering by po")
		}
	})

	t.Run("nil_filter_returns_all_non_system", func(t *testing.T) {
		got, err := repo.ListTemplates(ctx, application.ListFilter{
			TenantID: tnt.ID,
			Limit:    50,
		})
		if err != nil {
			t.Fatalf("ListTemplates(nil): %v", err)
		}
		keys := collectKeys(got)
		for _, k := range []string{"list-generic-po", "list-generic-blank", "list-generic-dc"} {
			if !keys[k] {
				t.Errorf("management listing (nil filter) must include %q", k)
			}
		}
	})
}
