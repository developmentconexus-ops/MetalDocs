//go:build integration
// +build integration

package distributioninfra_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	distributiondomain "metaldocs/internal/modules/distribution/domain"
	distributioninfra "metaldocs/internal/modules/distribution/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/pagination"
	"metaldocs/tests/integration/testdb"
)

// seedDistributionGrants seeds the distribution data for a restricted CD:
//   - userGrantID: user_grant + area_grant member (user_grant wins → precedence test)
//   - areaUserID:  area_grant member only
//   - companyUserID: area member (for company-scope CDs, used in other tests)
//
// Returns (tenantID, cdID, userGrantID, areaUserID, areaCode).
func seedDistributionGrants(
	t *testing.T, db *sql.DB,
) (tenantID, cdID, userGrantID, areaUserID, areaCode string) {
	t.Helper()
	ctx := context.Background()

	ten := testdb.NewTenant(t, db)
	tenantID = ten.ID

	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))
	areaCode = tax.ProcessAreaCode

	// Seed two users.
	uGrant := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	uArea := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	userGrantID = uGrant.ID
	areaUserID = uArea.ID

	// Both users need to be active area members of areaCode so v_active_user_areas
	// includes them (area_grant leg2 requires active membership via v_active_user_areas).
	// user_process_areas carries the membership.manage tripwire — use SeedWithCaps.
	for _, uid := range []string{uArea.ID, uGrant.ID} {
		uidCopy := uid
		testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx,
				`INSERT INTO public.user_process_areas
				   (user_id, tenant_id, area_code, role, effective_from, granted_by)
				 VALUES ($1, $2::uuid, $3, 'viewer', now(), $1)
				 ON CONFLICT DO NOTHING`,
				uidCopy, tenantID, areaCode,
			)
			return err
		})
	}

	// Build a restricted CD.
	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithTaxonomy(tax),
		testdb.WithOwner(uGrant.ID),
	)
	cdID = cd.ID

	// Leg 1: user_grant for uGrant.
	if _, err := db.ExecContext(ctx,
		`INSERT INTO public.controlled_document_user_grants
		   (tenant_id, controlled_document_id, user_id)
		 VALUES ($1::uuid, $2::uuid, $3)
		 ON CONFLICT DO NOTHING`,
		tenantID, cdID, uGrant.ID,
	); err != nil {
		t.Fatalf("seedDistributionGrants: user_grant: %v", err)
	}

	// Leg 2: area_grant for areaCode.
	// Both uArea and uGrant are area members → view emits user_grant for uGrant
	// (source_rank=1 beats source_rank=2) and area_grant for uArea (source_rank=2).
	if _, err := db.ExecContext(ctx,
		`INSERT INTO public.controlled_document_area_grants
		   (tenant_id, controlled_document_id, area_code)
		 VALUES ($1::uuid, $2::uuid, $3)
		 ON CONFLICT DO NOTHING`,
		tenantID, cdID, areaCode,
	); err != nil {
		t.Fatalf("seedDistributionGrants: area_grant: %v", err)
	}

	return tenantID, cdID, userGrantID, areaUserID, areaCode
}

func TestDistributionCoverage(t *testing.T) {
	db, _ := testdb.OpenFreshDatabase(t)
	ctx := context.Background()

	tenantID, cdID, userGrantID, areaUserID, areaCode := seedDistributionGrants(t, db)

	nameReader := iamdomain.NoopUserDisplayNameReader{}
	repo := distributioninfra.NewCoverageRepository(db, nameReader)

	t.Run("summary_total", func(t *testing.T) {
		total, err := repo.Summary(ctx, tenantID, cdID)
		if err != nil {
			t.Fatalf("Summary: %v", err)
		}
		// userGrantID (user_grant), areaUserID (area_grant) = 2 distinct users.
		if total < 2 {
			t.Errorf("want >=2 total targets, got %d", total)
		}
	})

	t.Run("recipients_three_legs_precedence", func(t *testing.T) {
		res, err := repo.Recipients(ctx, tenantID, cdID, "", 100)
		if err != nil {
			t.Fatalf("Recipients: %v", err)
		}

		byUser := make(map[string]distributiondomain.RecipientRow)
		for _, r := range res.Items {
			byUser[r.UserID] = r
		}

		// userGrantID must be user_grant (not area_grant even though they are an area member).
		rg, ok := byUser[userGrantID]
		if !ok {
			t.Fatalf("userGrantID %s not found in recipients", userGrantID)
		}
		if rg.Source != "user_grant" {
			t.Errorf("userGrantID: want source=user_grant, got %s", rg.Source)
		}
		if rg.AreaCode != nil {
			t.Errorf("userGrantID: want area_code=nil for user_grant, got %v", *rg.AreaCode)
		}
		if rg.AreaName != nil {
			t.Errorf("userGrantID: want area_name=nil for user_grant, got %v", *rg.AreaName)
		}

		// areaUserID must be area_grant with the correct area_code.
		ra, ok := byUser[areaUserID]
		if !ok {
			t.Fatalf("areaUserID %s not found in recipients", areaUserID)
		}
		if ra.Source != "area_grant" {
			t.Errorf("areaUserID: want source=area_grant, got %s", ra.Source)
		}
		if ra.AreaCode == nil || *ra.AreaCode != areaCode {
			t.Errorf("areaUserID: want area_code=%s, got %v", areaCode, ra.AreaCode)
		}
		// area_name must be non-nil (resolved via v_process_area_name JOIN).
		if ra.AreaName == nil || *ra.AreaName == "" {
			t.Errorf("areaUserID: want non-nil area_name, got %v", ra.AreaName)
		}
	})

	t.Run("recipients_keyset_pagination", func(t *testing.T) {
		// Fetch with limit=1 → expect has_more=true and a next cursor.
		page1, err := repo.Recipients(ctx, tenantID, cdID, "", 1)
		if err != nil {
			t.Fatalf("Recipients page1: %v", err)
		}
		if len(page1.Items) != 1 {
			t.Fatalf("want 1 item on page1, got %d", len(page1.Items))
		}
		if !page1.HasMore {
			t.Fatalf("want has_more=true on page1")
		}
		if page1.NextCursor == "" {
			t.Fatalf("want non-empty next_cursor on page1")
		}

		// Fetch page2 with the cursor.
		page2, err := repo.Recipients(ctx, tenantID, cdID, page1.NextCursor, 100)
		if err != nil {
			t.Fatalf("Recipients page2: %v", err)
		}

		// Union of pages must equal full set; no duplicates.
		all, err := repo.Recipients(ctx, tenantID, cdID, "", 100)
		if err != nil {
			t.Fatalf("Recipients all: %v", err)
		}
		combined := make(map[string]bool)
		for _, r := range page1.Items {
			combined[r.UserID] = true
		}
		for _, r := range page2.Items {
			if combined[r.UserID] {
				t.Errorf("duplicate user_id across pages: %s", r.UserID)
			}
			combined[r.UserID] = true
		}
		if len(combined) != len(all.Items) {
			t.Errorf("pages union size %d != full set size %d", len(combined), len(all.Items))
		}

		// Order must be stable: re-fetch page2 with same cursor, same result.
		page2again, err := repo.Recipients(ctx, tenantID, cdID, page1.NextCursor, 100)
		if err != nil {
			t.Fatalf("Recipients page2again: %v", err)
		}
		if len(page2again.Items) != len(page2.Items) {
			t.Errorf("unstable pagination: page2 size changed %d→%d", len(page2.Items), len(page2again.Items))
		}
	})

	t.Run("recipients_bad_cursor", func(t *testing.T) {
		_, err := repo.Recipients(ctx, tenantID, cdID, "!!!not-valid-base64!!!", 10)
		if err == nil {
			t.Fatal("want error for bad cursor, got nil")
		}
		if !errors.Is(err, pagination.ErrInvalidCursor) {
			t.Errorf("want ErrInvalidCursor-like error, got %v", err)
		}
	})

	// recipients_pagination_null_bucket: proves the NULL-area-name cursor round-trips.
	// Seeds ≥3 obligated users that ALL have NULL area_name (direct user_grants only,
	// no area_grant), then pages through with limit=1.  Each intermediate NextCursor
	// must be non-empty AND feeding it back must NOT error; the union of all pages must
	// equal the full seeded set with no duplicates and no skips; the final page must
	// have HasMore=false.
	//
	// Before the fix this sub-test FAILS on the second request because DecodeCursor
	// rejects a token whose first part is "" — which is exactly what the pre-fix code
	// encodes for NULL area_name.
	t.Run("recipients_pagination_null_bucket", func(t *testing.T) {
		// Seed a fresh tenant + 3 users, each with a direct user_grant → NULL area_name.
		nbTenant := testdb.NewTenant(t, db)
		nbCD := testdb.NewControlledDoc(t, db, testdb.WithTenant(nbTenant.ID))

		var nbUserIDs []string
		for i := 0; i < 3; i++ {
			u := testdb.NewUser(t, db, testdb.WithTenant(nbTenant.ID))
			nbUserIDs = append(nbUserIDs, u.ID)
			if _, err := db.ExecContext(ctx,
				`INSERT INTO public.controlled_document_user_grants
				   (tenant_id, controlled_document_id, user_id)
				 VALUES ($1::uuid, $2::uuid, $3)
				 ON CONFLICT DO NOTHING`,
				nbTenant.ID, nbCD.ID, u.ID,
			); err != nil {
				t.Fatalf("recipients_pagination_null_bucket: seed user_grant user %d: %v", i, err)
			}
		}

		nbRepo := distributioninfra.NewCoverageRepository(db, iamdomain.NoopUserDisplayNameReader{})

		// Collect all pages by following the cursor chain with limit=1.
		seen := make(map[string]int) // userID → page index seen
		cursor := ""
		pageIdx := 0
		for {
			page, err := nbRepo.Recipients(ctx, nbTenant.ID, nbCD.ID, cursor, 1)
			if err != nil {
				t.Fatalf("recipients_pagination_null_bucket: page %d with cursor %q: %v", pageIdx, cursor, err)
			}
			if len(page.Items) == 0 {
				t.Fatalf("recipients_pagination_null_bucket: page %d returned 0 items (want 1)", pageIdx)
			}
			for _, item := range page.Items {
				if prevPage, dup := seen[item.UserID]; dup {
					t.Errorf("recipients_pagination_null_bucket: duplicate user_id %s on pages %d and %d", item.UserID, prevPage, pageIdx)
				}
				seen[item.UserID] = pageIdx
				// Every item must have NULL area_name (user_grant source).
				if item.AreaName != nil {
					t.Errorf("recipients_pagination_null_bucket: page %d item %s: want nil area_name, got %q", pageIdx, item.UserID, *item.AreaName)
				}
			}
			if !page.HasMore {
				break
			}
			if page.NextCursor == "" {
				t.Fatal("recipients_pagination_null_bucket: HasMore=true but NextCursor is empty")
			}
			cursor = page.NextCursor
			pageIdx++
		}

		if len(seen) != len(nbUserIDs) {
			t.Errorf("recipients_pagination_null_bucket: want %d distinct users across pages, got %d (seen: %v)", len(nbUserIDs), len(seen), seen)
		}
		for _, uid := range nbUserIDs {
			if _, ok := seen[uid]; !ok {
				t.Errorf("recipients_pagination_null_bucket: seeded user %s never appeared in any page", uid)
			}
		}
	})

	t.Run("coverage_by_area", func(t *testing.T) {
		areas, err := repo.Coverage(ctx, tenantID, cdID)
		if err != nil {
			t.Fatalf("Coverage: %v", err)
		}
		// For restricted CD with area_grant rows there should be at least 1 area.
		if len(areas) == 0 {
			t.Error("want >=1 area coverage rows for restricted CD with area_grant, got 0")
		}
		for _, a := range areas {
			if a.AreaCode == "" {
				t.Error("area_code must not be empty")
			}
			if a.AreaName == "" {
				t.Error("area_name must not be empty")
			}
			if a.Total < 1 {
				t.Errorf("area %s: want total>=1, got %d", a.AreaCode, a.Total)
			}
		}
	})

	t.Run("coverage_company_scope_empty", func(t *testing.T) {
		// Create a company-scope CD (no area grants) → coverage must be empty.
		companyCDID := seedCompanyScopeCD(t, db, tenantID)

		areas, err := repo.Coverage(ctx, tenantID, companyCDID)
		if err != nil {
			t.Fatalf("Coverage for company-scope CD: %v", err)
		}
		if len(areas) != 0 {
			t.Errorf("want empty coverage for company-scope CD, got %d rows", len(areas))
		}
	})

	t.Run("unknown_document_empty", func(t *testing.T) {
		// An unknown tenant+CD produces an empty result without error (fail-closed: no
		// data returned for unauthorized/unknown resource). The repo must not panic or
		// produce a misleading result.
		badTenantID := uuid.NewString()
		badCDID := uuid.NewString()

		res, err := repo.Recipients(ctx, badTenantID, badCDID, "", 10)
		if err != nil {
			t.Fatalf("unknown_document_empty: empty tenant/cd should not error, got %v", err)
		}
		if len(res.Items) != 0 {
			t.Errorf("unknown_document_empty: want 0 items for unknown tenant/cd, got %d", len(res.Items))
		}

		total, err := repo.Summary(ctx, badTenantID, badCDID)
		if err != nil {
			t.Fatalf("unknown_document_empty: Summary for unknown tenant/cd should not error, got %v", err)
		}
		if total != 0 {
			t.Errorf("unknown_document_empty: want total=0 for unknown tenant/cd, got %d", total)
		}

		cov, err := repo.Coverage(ctx, badTenantID, badCDID)
		if err != nil {
			t.Fatalf("unknown_document_empty: Coverage for unknown tenant/cd should not error, got %v", err)
		}
		if len(cov) != 0 {
			t.Errorf("unknown_document_empty: want 0 coverage rows for unknown tenant/cd, got %d", len(cov))
		}
	})

	// fail_closed: G5 / interview #7 — the repo must REJECT a row whose source
	// value is outside the published enum {user_grant, area_grant, company_scope}.
	// The production view is a fixed UNION of 3 hard-coded source literals, so an
	// unknown source is unreachable through the real view by construction. We prove
	// the guard fires by temporarily replacing the view with a one-row literal that
	// carries source='future_grant' for a freshly-seeded tenant/CD/user, calling
	// repo.Recipients, and asserting the unknown-source error.
	//
	// The real view is restored in t.Cleanup before this sub-test exits. All other
	// sub-tests are already complete at this point (sub-tests run sequentially and
	// fail_closed is last), but the cleanup is correct regardless of ordering.
	//
	// Proof: this test goes RED if validateSource is removed from coverage_repository.go.
	t.Run("fail_closed", func(t *testing.T) {
		// Seed a fresh tenant, user, and CD whose IDs we will embed in the fake view.
		fcTenant := testdb.NewTenant(t, db)
		fcUser := testdb.NewUser(t, db, testdb.WithTenant(fcTenant.ID))
		fcCD := testdb.NewControlledDoc(t, db, testdb.WithTenant(fcTenant.ID))

		// Save the real view DDL so we can restore it in cleanup.
		// We recreate the real view verbatim from migration 0245.
		// MUST match migration 0245 / metaldocs.v_cd_obligated_readers exactly — keep in sync.
		const realViewDDL = `
			CREATE VIEW metaldocs.v_cd_obligated_readers AS
			WITH legs AS (
			  SELECT cdug.tenant_id,
			         cdug.controlled_document_id,
			         cdug.user_id,
			         NULL::text          AS area_code,
			         'user_grant'::text  AS source,
			         1                   AS source_rank
			    FROM public.controlled_document_user_grants cdug
			  UNION ALL
			  SELECT cdag.tenant_id,
			         cdag.controlled_document_id,
			         upa.user_id,
			         upa.area_code       AS area_code,
			         'area_grant'::text  AS source,
			         2                   AS source_rank
			    FROM public.controlled_document_area_grants cdag
			    JOIN metaldocs.v_active_user_areas upa
			      ON upa.tenant_id = cdag.tenant_id
			     AND upa.area_code = cdag.area_code
			  UNION ALL
			  SELECT f.tenant_id,
			         f.controlled_document_id,
			         tu.user_id,
			         NULL::text             AS area_code,
			         'company_scope'::text  AS source,
			         3                      AS source_rank
			    FROM metaldocs.v_cd_search_facts f
			    JOIN (SELECT DISTINCT tenant_id, user_id FROM metaldocs.v_active_user_areas) tu
			      ON tu.tenant_id = f.tenant_id
			   WHERE f.is_company
			)
			SELECT DISTINCT ON (tenant_id, controlled_document_id, user_id)
			       tenant_id,
			       controlled_document_id,
			       user_id,
			       area_code,
			       source
			  FROM legs
			 ORDER BY tenant_id, controlled_document_id, user_id, source_rank, area_code NULLS LAST`

		// Replace the view with a synthetic one that emits one row carrying an
		// unknown source for our specific tenant/CD/user. The view is narrowed to
		// that pair so it does not interfere with the real data already seeded by
		// the other sub-tests.
		if _, err := db.ExecContext(ctx, `DROP VIEW metaldocs.v_cd_obligated_readers`); err != nil {
			t.Fatalf("fail_closed: drop real view: %v", err)
		}
		t.Cleanup(func() {
			if _, err := db.ExecContext(ctx, `DROP VIEW IF EXISTS metaldocs.v_cd_obligated_readers`); err != nil {
				t.Logf("fail_closed cleanup: drop synthetic view: %v", err)
			}
			if _, err := db.ExecContext(ctx, realViewDDL); err != nil {
				t.Logf("fail_closed cleanup: restore real view: %v", err)
			}
		})

		fakeViewDDL := `
			CREATE VIEW metaldocs.v_cd_obligated_readers AS
			SELECT $1::uuid AS tenant_id,
			       $2::uuid AS controlled_document_id,
			       $3       AS user_id,
			       NULL::text AS area_code,
			       'future_grant'::text AS source`
		// PostgreSQL views cannot use query parameters; embed the UUIDs as literals.
		fakeViewDDL = `
			CREATE VIEW metaldocs.v_cd_obligated_readers AS
			SELECT '` + fcTenant.ID + `'::uuid AS tenant_id,
			       '` + fcCD.ID + `'::uuid AS controlled_document_id,
			       '` + fcUser.ID + `'       AS user_id,
			       NULL::text               AS area_code,
			       'future_grant'::text     AS source`
		if _, err := db.ExecContext(ctx, fakeViewDDL); err != nil {
			t.Fatalf("fail_closed: create synthetic view: %v", err)
		}

		// Now call Recipients for the seeded tenant+CD. The scan will hit
		// 'future_grant', validateSource will reject it, and repo must return an error.
		_, err := repo.Recipients(ctx, fcTenant.ID, fcCD.ID, "", 10)
		if err == nil {
			t.Fatal("fail_closed: want error for unknown source 'future_grant', got nil — validateSource guard is not firing")
		}
		if !strings.Contains(err.Error(), "unknown source value") {
			t.Errorf("fail_closed: want error containing 'unknown source value', got: %v", err)
		}
	})
}

// seedCompanyScopeCD creates a company-scope controlled document (visibility_scope='company')
// in the given tenant. Returns the CD id.
func seedCompanyScopeCD(t *testing.T, db *sql.DB, tenantID string) string {
	t.Helper()
	ctx := context.Background()

	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenantID))

	id := uuid.NewString()
	code := "cd-co-" + uuid.NewString()[:8]
	testdb.SeedWithCaps(t, db, `[{"cap":"controlled_documents.create"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO public.controlled_documents
			   (id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, status, visibility_scope)
			 VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, 'active', 'company')`,
			id, tenantID, tax.ProfileCode, tax.ProcessAreaCode, code, "Company CD", owner.ID,
		)
		return err
	})
	return id
}
