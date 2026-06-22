# Feature F2.1a — Publish `metaldocs.v_cd_obligated_readers`

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Folder:** `f2.1a-cd-obligated-view`
> **Status:** Planning → ready for TDD

## Source

- Milestone spec row (F2.1a, `../milestone.md`): owner-published view `metaldocs.v_cd_obligated_readers` carrying `(tenant_id, controlled_document_id, user_id, area_code TEXT NULL, source TEXT)` with three UNION legs (`user_grant`/`area_grant`/`company_scope`), DISTINCT BY `(tenant_id, cd_id, user_id)` with source precedence `user_grant > area_grant > company_scope`. ADR-0040 + ADR-0039 inventory row. **No Go handler.** Search untouched.
- Validation Gate (spec.md §Validation Gate): migration applies + idempotent; view shape `\d+` exact; three-leg integration test against fixtured graph; `git diff db/migrations/0243* internal/modules/search` empty; ADR files present; `hgcrossmodule` green.
- Consumer (F2.1c+F2.2): new `internal/modules/distribution` reads only this view + F2.1b's `v_process_area_name` + ADR-0029 iam display-name port.
- Patterns to follow: migration `db/migrations/0243_cd_search_visibility_contract.sql` (sibling view publication); test `internal/modules/controlleddocuments/infrastructure/cd_visibility_contract_parity_integration_test.go` + `membership_view_parity_integration_test.go` (testdb factory framework per ADR-0034, build-tag `integration`, reusable `seedCDVisibility`).

---

## Plan

# Feature F2.1a Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish `metaldocs.v_cd_obligated_readers` — the controlleddocuments-owned read contract carrying the obligated-reader set (direct user-grant ∪ active area-grant member ∪ company-scope active user), DISTINCT by `(tenant_id, cd_id, user_id)` with source precedence `user_grant > area_grant > company_scope`. The new view is the F2.1c+F2.2 distribution module's denominator source. **No Go code; one migration; ADR-0040; ADR-0039 inventory bullet.**

**Architecture:** A single forward-only migration creates the view as a UNION of three legs over CD's own base tables + iam's already-published `metaldocs.v_active_user_areas` + the existing `metaldocs.v_cd_search_facts.is_company` (consumed instead of hardcoding CD's scope enum). DISTINCT-ON enforces one row per user per CD with the precedence-winning leg's `source`/`area_code`. Same publication pattern as migration `0243`; same security posture (no `security_invoker`); idempotency via the migration runner's `schema_migrations` ledger. `v_cd_grantee` is untouched.

**Tech Stack:** PostgreSQL 14+ (CREATE VIEW, DISTINCT ON, UNION ALL). Go integration test under build tag `integration` using the canonical `tests/integration/testdb` factory framework (ADR-0034). cilint `hgcrossmodule` analyzer for boundary verification.

---

### Task 1: Failing integration test — three-leg semantics, DISTINCT precedence, search untouched

**Files:**
- Create: `internal/modules/controlleddocuments/infrastructure/v_cd_obligated_readers_integration_test.go`

Test reuses the existing `seedCDVisibility` helper (`membership_view_parity_integration_test.go:44`) for the restricted+company CD pair and the active/revoked/user-grant/none users. It adds **one extra user** that has BOTH a direct user-grant AND active area membership — that's the discriminator for the `user_grant > area_grant` precedence rule. The test asserts the obligated-set rows row-by-row against the fixtured graph.

- [ ] **Step 1: Write the failing integration test**

```go
//go:build integration
// +build integration

package infrastructure

import (
	"context"
	"database/sql"
	"sort"
	"testing"

	"metaldocs/tests/integration/testdb"
)

// F2.1a — obligated-reader read contract.
//
// metaldocs.v_cd_obligated_readers must return, for every (tenant_id, cd):
//   - one row per active user_grant            (source='user_grant',    area_code=NULL)
//   - one row per active area_grant member     (source='area_grant',    area_code=<area>)
//   - one row per active tenant user, for each company-scope CD
//                                              (source='company_scope', area_code=NULL)
//
// DISTINCT BY (tenant_id, cd, user_id) with source precedence
// user_grant > area_grant > company_scope. Revoked area members MUST be excluded.
// v_cd_grantee MUST NOT be modified.

type obligatedRow struct {
	UserID   string
	AreaCode sql.NullString
	Source   string
}

func obligatedSet(t *testing.T, db *sql.DB, tenantID, cdID string) []obligatedRow {
	t.Helper()
	rows, err := db.QueryContext(context.Background(),
		`SELECT user_id::text, area_code, source
		   FROM metaldocs.v_cd_obligated_readers
		  WHERE tenant_id = $1 AND controlled_document_id = $2::uuid
		  ORDER BY user_id, source`,
		tenantID, cdID)
	if err != nil {
		t.Fatalf("obligatedSet query: %v", err)
	}
	defer rows.Close()
	var out []obligatedRow
	for rows.Next() {
		var r obligatedRow
		if err := rows.Scan(&r.UserID, &r.AreaCode, &r.Source); err != nil {
			t.Fatalf("obligatedSet scan: %v", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("obligatedSet rows: %v", err)
	}
	return out
}

// seedObligatedExtraOverlap adds one user that is BOTH user-granted AND an active
// area member on the restricted CD — the source-precedence discriminator.
func seedObligatedExtraOverlap(t *testing.T, db *sql.DB, sc cdScenario) string {
	t.Helper()
	ctx := context.Background()

	overlap := testdb.NewUser(t, db, testdb.WithTenant(sc.tenantID)).ID

	if _, err := db.ExecContext(ctx,
		`INSERT INTO public.controlled_document_user_grants
		   (tenant_id, controlled_document_id, user_id)
		 VALUES ($1::uuid,$2::uuid,$3)`,
		sc.tenantID, sc.cdRestricted, overlap); err != nil {
		t.Fatalf("seed overlap user grant: %v", err)
	}

	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1,$2::uuid,$3,'qms_admin', now() - interval '2 hours')`,
			overlap, sc.tenantID, sc.area,
		)
		return err
	})
	return overlap
}

func TestObligatedReaders_RestrictedCD_ThreeLegsWithPrecedence(t *testing.T) {
	db, _ := testdb.Open(t)
	sc := seedCDVisibility(t, db)
	overlap := seedObligatedExtraOverlap(t, db, sc)

	got := obligatedSet(t, db, sc.tenantID, sc.cdRestricted)

	// Expected, DISTINCT by user_id:
	//   areaMember   → source='area_grant',  area_code='quality'
	//   userGrant    → source='user_grant',  area_code=NULL
	//   overlap      → source='user_grant',  area_code=NULL    (precedence: user_grant > area_grant)
	// EXCLUDED: revokedMem (effective_to NOT NULL), owner (no grant edge), none (no grant edge).
	want := map[string]obligatedRow{
		sc.areaMember: {UserID: sc.areaMember, AreaCode: sql.NullString{String: sc.area, Valid: true}, Source: "area_grant"},
		sc.userGrant:  {UserID: sc.userGrant, AreaCode: sql.NullString{}, Source: "user_grant"},
		overlap:       {UserID: overlap, AreaCode: sql.NullString{}, Source: "user_grant"},
	}

	if len(got) != len(want) {
		t.Fatalf("restricted obligated set size: got %d want %d (rows=%+v)", len(got), len(want), got)
	}
	for _, r := range got {
		w, ok := want[r.UserID]
		if !ok {
			t.Fatalf("unexpected obligated row: %+v", r)
		}
		if r.Source != w.Source {
			t.Fatalf("source drift for user %s: got %q want %q", r.UserID, r.Source, w.Source)
		}
		if r.AreaCode.Valid != w.AreaCode.Valid || r.AreaCode.String != w.AreaCode.String {
			t.Fatalf("area_code drift for user %s: got %+v want %+v", r.UserID, r.AreaCode, w.AreaCode)
		}
	}

	// Discriminators: revoked area member must NOT appear; owner/none must NOT appear.
	for _, bad := range []string{sc.revokedMem, sc.owner, sc.none} {
		for _, r := range got {
			if r.UserID == bad {
				t.Fatalf("user %s must NOT be in restricted obligated set, got %+v", bad, r)
			}
		}
	}
}

func TestObligatedReaders_CompanyCD_AllActiveTenantUsers(t *testing.T) {
	db, _ := testdb.Open(t)
	sc := seedCDVisibility(t, db)

	got := obligatedSet(t, db, sc.tenantID, sc.cdCompany)

	// Company-scope leg = DISTINCT user_id from metaldocs.v_active_user_areas in this tenant
	//                     × this company-scope CD; source='company_scope', area_code=NULL.
	// In the fixture only `areaMember` has an active area membership. revokedMem is excluded;
	// owner/userGrant/none have no area memberships → not "active tenant users" under the
	// v_active_user_areas definition (ADR 0037 D1).
	if len(got) != 1 {
		t.Fatalf("company obligated set size: got %d want 1 (rows=%+v)", len(got), got)
	}
	r := got[0]
	if r.UserID != sc.areaMember {
		t.Fatalf("company obligated user: got %s want %s", r.UserID, sc.areaMember)
	}
	if r.Source != "company_scope" {
		t.Fatalf("company source: got %q want %q", r.Source, "company_scope")
	}
	if r.AreaCode.Valid {
		t.Fatalf("company area_code: got %v want NULL", r.AreaCode)
	}
}

func TestObligatedReaders_ViewShape(t *testing.T) {
	db, _ := testdb.Open(t)
	const q = `
	  SELECT column_name, data_type, is_nullable
	    FROM information_schema.columns
	   WHERE table_schema = 'metaldocs'
	     AND table_name   = 'v_cd_obligated_readers'
	   ORDER BY ordinal_position`
	rows, err := db.QueryContext(context.Background(), q)
	if err != nil {
		t.Fatalf("view shape query: %v", err)
	}
	defer rows.Close()
	type col struct{ name, typ, nullable string }
	var got []col
	for rows.Next() {
		var c col
		if err := rows.Scan(&c.name, &c.typ, &c.nullable); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, c)
	}
	want := []col{
		{"tenant_id", "uuid", "NO"},
		{"controlled_document_id", "uuid", "NO"},
		{"user_id", "uuid", "NO"},
		{"area_code", "text", "YES"},
		{"source", "text", "NO"},
	}
	if len(got) != len(want) {
		t.Fatalf("column count: got %d want %d (got=%+v)", len(got), len(want), got)
	}
	sortCols := func(s []col) { sort.SliceStable(s, func(i, j int) bool { return s[i].name < s[j].name }) }
	sortCols(got)
	sortCols(want)
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("column %d drift: got %+v want %+v", i, got[i], want[i])
		}
	}
}
```

- [ ] **Step 2: Run the test — verify it fails because the view does not exist yet**

Run:
```bash
go test -tags=integration -run "TestObligatedReaders" ./internal/modules/controlleddocuments/infrastructure/... -v
```
Expected: FAIL — `pq: relation "metaldocs.v_cd_obligated_readers" does not exist` (or equivalent on all three tests).

- [ ] **Step 3: Commit the failing test**

```bash
git add internal/modules/controlleddocuments/infrastructure/v_cd_obligated_readers_integration_test.go
git commit -m "test(M2/F2.1a): failing parity test for v_cd_obligated_readers — three legs + precedence"
```

---

### Task 2: Forward-only migration `0245` — publish the view

**Files:**
- Create: `db/migrations/0245_cd_obligated_readers_view.sql`

- [ ] **Step 1: Write the migration**

```sql
-- 0245: controlleddocuments publishes metaldocs.v_cd_obligated_readers — the
-- denominator read contract for the distribution module (mission frontend-
-- screen-completion, M2/F2.1a; ADR-0040; ADR-0039 D3a/D4 inventory).
--
-- ADR-0039 D3a/D4: the distribution module (internal/modules/distribution, built in
-- F2.1c/F2.2) reads THIS view (not CD's base tables) to derive the obligated-reader
-- set for /api/v1/documents/:id/distribution*. Three legs UNION'd, DISTINCT BY
-- (tenant_id, controlled_document_id, user_id) with source precedence
-- user_grant > area_grant > company_scope (most-specific wins). For area_grant
-- rows on a user with multiple granting areas: lowest area_code wins
-- (deterministic).
--
-- Why a new sibling view (not extending v_cd_grantee):
--   v_cd_grantee is restricted-only by design (migration 0243 COMMENT + the
--   WHERE visibility_scope='restricted' gate) — that gate is the search-semantic
--   contract making search's EXISTS predicate (search/.../v2documents/reader.go)
--   correct-by-construction. Mutating it forces search to carry distribution-domain
--   knowledge → module-boundary leak. There is also zero DROP/ALTER VIEW precedent
--   across 244 migrations (wiki/database/migration-policy.md is forward-only). New
--   sibling view = clean. See ADR-0040.
--
-- Reads (compliant per ADR-0039):
--   own base tables: public.controlled_documents,
--                    public.controlled_document_user_grants,
--                    public.controlled_document_area_grants
--   published views: metaldocs.v_active_user_areas (iam, D3a — encodes
--                    effective_to IS NULL, ADR 0037 D1),
--                    metaldocs.v_cd_search_facts (CD-owned, is_company scalar)
--
-- Security posture matches the underlying tables (no security_invoker), identical
-- to 0242/0243.

BEGIN;

CREATE VIEW metaldocs.v_cd_obligated_readers AS
WITH legs AS (
  -- Leg 1: direct user-grant. source_rank=1 (highest precedence).
  SELECT cdug.tenant_id,
         cdug.controlled_document_id,
         cdug.user_id,
         NULL::text          AS area_code,
         'user_grant'::text  AS source,
         1                   AS source_rank
    FROM public.controlled_document_user_grants cdug

  UNION ALL

  -- Leg 2: area-grant ⋈ active area membership. source_rank=2.
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

  -- Leg 3: company-scope CDs × DISTINCT active tenant users.
  -- Consumes CD's own v_cd_search_facts.is_company (1:1 over controlled_documents)
  -- — no hardcoded scope literal.
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
 ORDER BY tenant_id, controlled_document_id, user_id, source_rank, area_code NULLS LAST;

COMMENT ON VIEW metaldocs.v_cd_obligated_readers IS
  'Published CD obligated-reader read contract (ADR-0040; ADR-0039 D3a/D4): one (tenant_id, controlled_document_id, user_id, area_code NULL, source) row per user obligated to read a CD. Three legs UNION''d (user_grant ∪ active area_grant member ∪ company-scope active tenant user) and DISTINCT BY (tenant_id, cd, user_id) with source precedence user_grant > area_grant > company_scope. Non-owner modules (distribution) read THIS view, never CD''s grant base tables. v_cd_grantee remains restricted-only by design (search semantics, migration 0243). Mission frontend-screen-completion M2/F2.1a.';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0245', 'controlleddocuments publishes metaldocs.v_cd_obligated_readers obligated-reader view (M2/F2.1a, ADR-0040)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
```

- [ ] **Step 2: Apply the migration on a fresh template DB and re-run the tests**

Run:
```bash
.\scripts\start-api.ps1 -Build
```
Expected: all migrations apply green; logs show `0245 controlleddocuments publishes metaldocs.v_cd_obligated_readers …`.

Then:
```bash
go test -tags=integration -run "TestObligatedReaders" ./internal/modules/controlleddocuments/infrastructure/... -v
```
Expected: PASS on all three tests.

- [ ] **Step 3: Idempotency check — re-running the migration is a no-op**

Run the API start script a second time (it re-applies the ledger):
```bash
.\scripts\start-api.ps1
```
Then verify the ledger has exactly one row:
```bash
psql "$env:DATABASE_URL" -c "SELECT version, description FROM public.schema_migrations WHERE version='0245'"
```
Expected: exactly **one** row with version `0245`.

- [ ] **Step 4: Verify search + v_cd_grantee untouched**

Run:
```bash
git diff db/migrations/0243_cd_search_visibility_contract.sql
git diff internal/modules/search
```
Both expected: empty.

- [ ] **Step 5: Commit migration**

```bash
git add db/migrations/0245_cd_obligated_readers_view.sql
git commit -m "feat(M2/F2.1a): publish metaldocs.v_cd_obligated_readers view (ADR-0040)"
```

---

### Task 3: ADR-0040 + ADR-0039 inventory bullet

**Files:**
- Create: `wiki/decisions/0040-cd-obligated-readers-view.md`
- Modify: `wiki/decisions/0039-cross-module-base-table-read-boundary.md` (add a `Related code` bullet pointing at the new migration)
- Modify: `wiki/decisions/index.md` (insert the ADR-0040 entry per existing format) — verify against the file before editing; skip if the index is auto-generated.

- [ ] **Step 1: Write ADR-0040**

```markdown
# ADR 0040 — controlleddocuments publishes `metaldocs.v_cd_obligated_readers` (obligated-reader read contract)

> **Status:** Accepted 2026-06-21
> **Last verified:** 2026-06-21
> **Deciders:** leandrotca.work (operator), MetalDocs backend
> **Context window:** Mission `frontend-screen-completion` · Milestone M2 (Distribuição coverage-scope) · Feature F2.1a (re-decomposition under HS-6 path A).
> **Related ADRs:** [0037 — Membership temporal model](./0037-membership-temporal-model.md) (active-now ⟺ `effective_to IS NULL`); [0039 — Cross-module read boundary](./0039-cross-module-base-table-read-boundary.md) (this view is a D3(a) exemption — the compliant mechanism distribution uses); ADR-0041 (taxonomy publishes `v_process_area_name`, F2.1b, sibling view); ADR-0042 (distribution module + denominator-only contract, F2.1c).
> **Related code (Last verified 2026-06-21):**
> - `db/migrations/0245_cd_obligated_readers_view.sql` — this view.
> - `db/migrations/0243_cd_search_visibility_contract.sql` — `v_cd_grantee` (restricted-only by design, **not** modified by this ADR).
> - `db/migrations/0242_iam_v_active_user_areas_view.sql` — iam's active-membership view consumed by the area-grant + company-scope legs.

## Context

M2/F2.1a (under mission `frontend-screen-completion`) builds a Grade-A, read-only denominator endpoint for the Distribuição & Cobertura screen: who must read a given CD, broken down by area, with a true total. The endpoint is implemented by a new `distribution` module (F2.1c/F2.2) which, per ADR-0039 D3a, must read another module's **published view**, never its base tables.

The existing CD-published view `metaldocs.v_cd_grantee` (migration 0243) is **not sufficient**:

1. It carries only `(tenant_id, controlled_document_id, grantee_user_id)` — distribution needs `area_code` (which area granted access) and a `source` discriminator (`user_grant`/`area_grant`/`company_scope`).
2. It is gated `visibility_scope = 'restricted'` **by design** — that predicate is the search-semantic contract that makes search's `EXISTS` predicate (`search/infrastructure/v2documents/reader.go:99-106`) correct-by-construction. Company-scope CDs contribute zero rows; search handles those via `v_cd_search_facts.is_company` separately.

Extending `v_cd_grantee` in place would force search to carry a distribution-domain `source` discriminator → module-boundary leak. There is also zero `DROP`/`ALTER VIEW` precedent across the 244 prior migrations (`wiki/database/migration-policy.md` is forward-only; non-additive view DDL is policy-hostile).

## Decision

### D1 — Publish a new sibling view, do not mutate `v_cd_grantee`

Migration 0245 creates `metaldocs.v_cd_obligated_readers`. `v_cd_grantee` is untouched. Both views coexist: search reads `v_cd_grantee` (restricted edges); distribution reads `v_cd_obligated_readers` (full obligated set with per-row `area_code`/`source`).

### D2 — Shape

```
metaldocs.v_cd_obligated_readers (
  tenant_id              uuid     not null,
  controlled_document_id uuid     not null,
  user_id                uuid     not null,
  area_code              text         null,   -- null when source ∈ {user_grant, company_scope}
  source                 text     not null    -- 'user_grant' | 'area_grant' | 'company_scope'
)
-- DISTINCT BY (tenant_id, controlled_document_id, user_id)
-- Source precedence: user_grant > area_grant > company_scope (most-specific wins)
-- For area_grant on a user with multiple granting areas: lowest area_code wins (deterministic)
```

### D3 — Three legs (UNION ALL → DISTINCT ON)

1. **`user_grant`** — `public.controlled_document_user_grants`; `area_code=NULL`.
2. **`area_grant`** — `public.controlled_document_area_grants` ⋈ `metaldocs.v_active_user_areas` on `(tenant_id, area_code)`; `area_code=upa.area_code`. (`v_active_user_areas` encodes `effective_to IS NULL` per ADR 0037 D1 — revoked members are excluded.)
3. **`company_scope`** — `metaldocs.v_cd_search_facts` where `is_company=true`, cross-joined to DISTINCT `(tenant_id, user_id)` from `metaldocs.v_active_user_areas`. Consumes the existing `is_company` scalar (no hardcoded scope literal).

### D4 — Source precedence rule (DISTINCT ON via `source_rank`)

`user_grant=1, area_grant=2, company_scope=3` (numerical rank, **low rank wins**). The view body applies `SELECT DISTINCT ON (tenant_id, controlled_document_id, user_id) … ORDER BY tenant_id, controlled_document_id, user_id, source_rank, area_code NULLS LAST` so a user appearing through multiple legs collapses to the lowest-rank leg's `source`/`area_code`. The `area_code` tiebreaker keeps the chosen `area_grant` row deterministic when a user is an active member of multiple granting areas (lowest `area_code`).

### D5 — Security posture

Plain `CREATE VIEW` (no `security_invoker`) — identical to `v_active_user_areas` (0242) and `v_cd_grantee` (0243). RLS over the base tables continues to apply unchanged.

### D6 — No index on the view

Views cannot carry indexes. If F2.2's integration test or production telemetry surfaces a real latency problem with the company-scope leg (cross-join of company-scope CDs × active tenant users), the remediation is a separate, evidence-backed decision (materialized view, per-CD lateral, or selective index on the underlying tables) — premature optimization is out of scope for this ADR.

## Consequences

- The `distribution` module (F2.1c/F2.2) is ADR-0039 D3a compliant by construction — it reads only published views + the ADR-0029 iam display-name port; `hgcrossmodule` analyzer holds.
- Search is untouched; `v_cd_grantee`'s restricted-only semantic contract is preserved.
- The denominator-only contract is **forward-compatible** for the parked `document-distribution-mission`: that mission may add columns to the consumer DTO (read/ack/overdue), but the obligated set itself remains a function of the same three legs.

## Verification

- Migration applies cleanly + idempotent on re-run (the `schema_migrations` ledger enforces this).
- `internal/modules/controlleddocuments/infrastructure/v_cd_obligated_readers_integration_test.go` asserts three-leg semantics + DISTINCT precedence + view shape against a fixtured graph.
- `git diff db/migrations/0243*` = empty; `git diff internal/modules/search` = empty.
- `go run ./tools/cilint/...` = 0 H-G under both `controlleddocuments` (publishes) and `distribution` (reads, once F2.2 lands).
```

- [ ] **Step 2: Add the ADR-0039 inventory bullet**

Open `wiki/decisions/0039-cross-module-base-table-read-boundary.md`. Locate the `Related code (Last verified …)` list at the top of the front matter — it currently ends with the `0244_documents_search_projection.sql` bullet. Append, on a new line immediately after that bullet (keeping the same `> - ` prefix):

```
> - `db/migrations/0245_cd_obligated_readers_view.sql` — the D3(a)/D4 controlleddocuments-published obligated-reader read contract, built in mission frontend-screen-completion M2/F2.1a: `metaldocs.v_cd_obligated_readers` (per-CD obligated set with `area_code` + `source` discriminator across `user_grant`/`area_grant`/`company_scope` legs, DISTINCT by user with source precedence). The distribution module (M2/F2.1c+F2.2) reads this view instead of CD's grant base tables. Sibling of `v_cd_grantee` (restricted-only, search-owned semantics, untouched).
```

Also flip the front-matter `Last verified` date to `2026-06-21` if it was earlier.

- [ ] **Step 3: Verify the inventory bullet renders and the ADR file is reachable**

Run:
```bash
grep -n "v_cd_obligated_readers" wiki/decisions/0039-cross-module-base-table-read-boundary.md
ls wiki/decisions/0040-cd-obligated-readers-view.md
```
Expected: at least one hit in 0039; the 0040 file exists.

- [ ] **Step 4: Update the ADR index entry (if the index is hand-maintained)**

Read `wiki/decisions/index.md`. If it has a numbered list of ADRs in order, insert an entry for `0040` between 0039 and the previous tail. **If the file is auto-generated (has a banner saying so), skip — do not hand-edit.**

- [ ] **Step 5: Commit ADRs**

```bash
git add wiki/decisions/0040-cd-obligated-readers-view.md wiki/decisions/0039-cross-module-base-table-read-boundary.md
git add wiki/decisions/index.md  # only if step 4 modified it
git commit -m "docs(adr): ADR-0040 v_cd_obligated_readers + ADR-0039 inventory (M2/F2.1a)"
```

---

### Task 4: Run the full Validation Gate

Mirrors the spec's Validation Gate table row-by-row. No code changes here — just evidence collection.

- [ ] **Step 1: Migration applies + idempotent (re-confirm)**

```bash
psql "$env:DATABASE_URL" -c "SELECT version FROM public.schema_migrations WHERE version='0245'"
```
Expected: exactly one row.

- [ ] **Step 2: View shape exact**

```bash
psql "$env:DATABASE_URL" -c "\d+ metaldocs.v_cd_obligated_readers"
```
Expected: 5 columns — `tenant_id uuid NOT NULL`, `controlled_document_id uuid NOT NULL`, `user_id uuid NOT NULL`, `area_code text NULL`, `source text NOT NULL`.

- [ ] **Step 3: Three-leg integration test green (re-run)**

```bash
go test -tags=integration -run "TestObligatedReaders" ./internal/modules/controlleddocuments/infrastructure/... -v
```
Expected: PASS — `TestObligatedReaders_RestrictedCD_ThreeLegsWithPrecedence`, `TestObligatedReaders_CompanyCD_AllActiveTenantUsers`, `TestObligatedReaders_ViewShape`.

- [ ] **Step 4: Search + v_cd_grantee untouched**

```bash
git diff origin/main -- db/migrations/0243_cd_search_visibility_contract.sql
git diff origin/main -- internal/modules/search
```
Expected: both empty.

- [ ] **Step 5: `hgcrossmodule` analyzer green**

```bash
go run ./tools/cilint/...
```
Expected: 0 H-G findings. The new view publishes from CD's own base tables + iam's `v_active_user_areas` (D3a) + CD's own `v_cd_search_facts` — fully compliant.

- [ ] **Step 6: Build + vet + non-integration tests green (regression)**

```bash
go build ./...
go vet ./...
go test ./...
```
Expected: all green.

- [ ] **Step 7: Capture evidence**

Copy `.claude/skills/milestone/templates/feature-evidence.md` into `docs/superpowers/milestones/frontend-screen-completion/milestone-2-distribuicao/f2.1a-cd-obligated-view/evidence.md` and fill the rows with the exact commands run + their real output (Steps 1–6 above). Label every piece of proof `real` (no fixtures used at the gate level — the integration test uses a fixtured graph but the view + migration evidence is real-DB).

- [ ] **Step 8: Commit evidence + close the feature**

```bash
git add docs/superpowers/milestones/frontend-screen-completion/milestone-2-distribuicao/f2.1a-cd-obligated-view/evidence.md
git commit -m "docs(M2/F2.1a): close evidence — v_cd_obligated_readers gate green"
```

---

## Execution notes

Filled during `superpowers:subagent-driven-development`:

- Model: Sonnet 4.6 per `[[workflow-model-balancing]]` + operator directive 2026-06-21.
- Deviations from plan: <none yet>.
- Open questions answered: <none yet>.
