# Feature F2.1b — Publish `metaldocs.v_process_area_name`

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Folder:** `f2.1b-taxonomy-area-name-view`
> **Status:** Planning → ready for TDD

## Source

- Milestone spec row (F2.1b, `../milestone.md`): taxonomy publishes `metaldocs.v_process_area_name (tenant_id uuid, area_code text, area_name text)` — 1:1 projection of taxonomy's process-area base table, one row per `(tenant_id, area_code)`. ADR-0041 + ADR-0039 inventory row. **No Go code.**
- Validation Gate (`spec.md` §Validation Gate): migration applies + idempotent; `\d+` view shape = 3 columns; 1:1 rowcount per tenant integration test; ADR-0041 present + ADR-0039 inventory updated; `hgcrossmodule` analyzer green.
- Consumer (F2.2): new `internal/modules/distribution` joins this view to F2.1a's `v_cd_obligated_readers` on `(tenant_id, area_code)` to populate `DistributionRecipient.area_name` + `DistributionAreaCoverage.area_name`.
- Patterns to follow:
  - Sibling F2.1a migration `db/migrations/0245_cd_obligated_readers_view.sql` (view publication + `schema_migrations` ledger + `COMMENT ON VIEW` + idempotent re-run).
  - Sibling F2.1a integration test `internal/modules/controlleddocuments/infrastructure/v_cd_obligated_readers_integration_test.go` (build-tag `integration`, testdb factory framework per ADR-0034).
  - Existing taxonomy test `internal/modules/taxonomy/infrastructure/area_catalog_reader_parity_integration_test.go` (uses `testdb.NewTaxonomy(t, db)` — already seeds an area with `name = code`).
- Base table recon (this session): `metaldocs.document_process_areas` — columns `code text NOT NULL`, `name text NOT NULL`, `tenant_id uuid NOT NULL DEFAULT '<system>'`, `description`, `is_active`, `parent_code`, `owner_user_id`, `default_approver_role`, `archived_at`. PK = `(code)` (global). CHECK `area_code_format` + `area_code_not_tenant`. RLS enabled (migration 0237). Already read raw by `internal/modules/taxonomy/infrastructure/area_catalog_reader.go:32` — same `(tenant_id, code)` predicate the view projects. Spec Q5 — **no extra columns** (no `is_active`, no `archived_at` filter; 1:1 projection, additive later).

---

## Plan

# Feature F2.1b Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish `metaldocs.v_process_area_name` — the taxonomy-owned read contract carrying `(tenant_id, area_code, area_name)`, one row per `(tenant_id, area_code)`. Pure 1:1 projection of `metaldocs.document_process_areas` (renaming `code → area_code`, `name → area_name`). Sole consumer is F2.2's distribution handler joining on `(tenant_id, area_code)`. **No Go code; one migration; ADR-0041; ADR-0039 inventory bullet.**

**Architecture:** A single forward-only migration creates the view as a plain `SELECT` over taxonomy's base table. Same publication pattern as sibling 0245 (no `security_invoker`, idempotent via the `schema_migrations` ledger, `COMMENT ON VIEW` with the ADR pointer). No filter on `is_active` / `archived_at` — labels resolve even for archived areas, matching the existing port behavior (`area_catalog_reader.go:32` does not filter either). The two minimal column renames (`code → area_code`, `name → area_name`) align the view shape with F2.1a's `v_cd_obligated_readers.area_code` so the F2.2 join is natural.

**Tech Stack:** PostgreSQL 14+ (CREATE VIEW). Go integration test under build tag `integration` using the canonical `tests/integration/testdb` factory framework (ADR-0034). cilint `hgcrossmodule` analyzer for boundary verification.

---

### Task 1: Failing integration test — view shape + 1:1 projection per tenant

**Files:**
- Create: `internal/modules/taxonomy/infrastructure/v_process_area_name_integration_test.go`

The test seeds N process areas in one tenant via the taxonomy factory + direct inserts (the factory only seeds one), then asserts:
1. View shape exact (`information_schema.columns`).
2. Per-tenant rowcount = number of areas the test seeded (1:1 projection isolated to the test's tenant).
3. Name + code round-trip for the seeded rows (the renames `code → area_code`, `name → area_name` carry the values verbatim).
4. Cross-tenant isolation: an area seeded under a second tenant does not appear in the first tenant's result.

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

	"github.com/google/uuid"

	"metaldocs/tests/integration/testdb"
)

// F2.1b — taxonomy publishes metaldocs.v_process_area_name.
//
// Shape:  (tenant_id uuid NOT NULL, area_code text NOT NULL, area_name text NOT NULL)
// Rule:   one row per (tenant_id, area_code); 1:1 projection of
//         metaldocs.document_process_areas (code → area_code, name → area_name).
// No filter on is_active / archived_at — labels must resolve for archived areas too
// (parity with the existing AreaCatalogReader port at
// area_catalog_reader.go:32, which does not filter either).

type areaNameRow struct {
	TenantID string
	AreaCode string
	AreaName string
}

func seedExtraProcessArea(t *testing.T, db *sql.DB, tenantID string) (code, name string) {
	t.Helper()
	code = "f21b-" + uuid.NewString()[:8]
	name = "F2.1b Area " + code
	if _, err := db.ExecContext(context.Background(),
		`INSERT INTO metaldocs.document_process_areas (tenant_id, code, name)
		 VALUES ($1::uuid, $2, $3)`,
		tenantID, code, name,
	); err != nil {
		t.Fatalf("seed extra process area: %v", err)
	}
	return code, name
}

func areaNameRowsForTenant(t *testing.T, db *sql.DB, tenantID string) []areaNameRow {
	t.Helper()
	rows, err := db.QueryContext(context.Background(),
		`SELECT tenant_id::text, area_code, area_name
		   FROM metaldocs.v_process_area_name
		  WHERE tenant_id = $1::uuid
		  ORDER BY area_code`,
		tenantID)
	if err != nil {
		t.Fatalf("v_process_area_name query: %v", err)
	}
	defer rows.Close()
	var out []areaNameRow
	for rows.Next() {
		var r areaNameRow
		if err := rows.Scan(&r.TenantID, &r.AreaCode, &r.AreaName); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	return out
}

func TestProcessAreaName_OneToOneProjection_PerTenant(t *testing.T) {
	db, _ := testdb.Open(t)
	tax := testdb.NewTaxonomy(t, db) // seeds one area: code = tax.ProcessAreaCode, name = code
	c2, n2 := seedExtraProcessArea(t, db, tax.TenantID)
	c3, n3 := seedExtraProcessArea(t, db, tax.TenantID)

	got := areaNameRowsForTenant(t, db, tax.TenantID)

	want := map[string]string{
		tax.ProcessAreaCode: tax.ProcessAreaCode, // factory sets name = code
		c2:                  n2,
		c3:                  n3,
	}
	if len(got) != len(want) {
		t.Fatalf("rowcount: got %d want %d (rows=%+v)", len(got), len(want), got)
	}
	for _, r := range got {
		w, ok := want[r.AreaCode]
		if !ok {
			t.Fatalf("unexpected area_code in view: %+v", r)
		}
		if r.AreaName != w {
			t.Fatalf("area_name drift for code %q: got %q want %q", r.AreaCode, r.AreaName, w)
		}
		if r.TenantID != tax.TenantID {
			t.Fatalf("tenant_id drift: got %q want %q", r.TenantID, tax.TenantID)
		}
	}
}

func TestProcessAreaName_CrossTenantIsolation(t *testing.T) {
	db, _ := testdb.Open(t)
	taxA := testdb.NewTaxonomy(t, db)
	taxB := testdb.NewTaxonomy(t, db)

	gotA := areaNameRowsForTenant(t, db, taxA.TenantID)
	for _, r := range gotA {
		if r.AreaCode == taxB.ProcessAreaCode {
			t.Fatalf("tenant A view leaked tenant B's area_code %q: %+v", taxB.ProcessAreaCode, r)
		}
	}
}

func TestProcessAreaName_ViewShape(t *testing.T) {
	db, _ := testdb.Open(t)
	rows, err := db.QueryContext(context.Background(), `
	  SELECT column_name, data_type, is_nullable
	    FROM information_schema.columns
	   WHERE table_schema = 'metaldocs'
	     AND table_name   = 'v_process_area_name'
	   ORDER BY ordinal_position`)
	if err != nil {
		t.Fatalf("shape query: %v", err)
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
		{"area_code", "text", "NO"},
		{"area_name", "text", "NO"},
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
go test -tags=integration -run "TestProcessAreaName" ./internal/modules/taxonomy/infrastructure/... -v
```
Expected: FAIL — `pq: relation "metaldocs.v_process_area_name" does not exist` (or equivalent) on all three tests.

- [ ] **Step 3: Commit the failing test**

```bash
git add internal/modules/taxonomy/infrastructure/v_process_area_name_integration_test.go
git commit -m "test(M2/F2.1b): failing test for v_process_area_name — shape + 1:1 + isolation"
```

---

### Task 2: Forward-only migration `0246` — publish the view

**Files:**
- Create: `db/migrations/0246_taxonomy_process_area_name_view.sql`

- [ ] **Step 1: Write the migration**

```sql
-- 0246: taxonomy publishes metaldocs.v_process_area_name — the per-area
-- human-label read contract for the distribution module (mission frontend-
-- screen-completion, M2/F2.1b; ADR-0041; ADR-0039 D3a/D4 inventory).
--
-- ADR-0039 D3a/D4: the distribution module (internal/modules/distribution, built
-- in F2.1c/F2.2) joins THIS view to F2.1a's metaldocs.v_cd_obligated_readers on
-- (tenant_id, area_code) to populate DistributionRecipient.area_name and
-- DistributionAreaCoverage.area_name. The base table metaldocs.document_process_areas
-- is taxonomy-owned and must not be read directly by non-taxonomy modules
-- (hgcrossmodule analyzer; ADR-0039).
--
-- Why a 1:1 projection (no is_active / archived_at filter):
--   The existing taxonomy AreaCatalogReader port
--   (internal/modules/taxonomy/infrastructure/area_catalog_reader.go:32) reads the
--   base table without any active/archived filter — names resolve for archived
--   areas too. The view preserves that semantic. Adding a filter now would change
--   contract on the existing port's behavior and is YAGNI (spec.md Q5: minimal
--   contract, additive later if a consumer actually needs it).
--
-- Renames (code → area_code, name → area_name):
--   align the published shape with F2.1a's v_cd_obligated_readers.area_code so the
--   F2.2 join is natural.
--
-- Reads (compliant per ADR-0039):
--   own base tables: metaldocs.document_process_areas (taxonomy-owned)
--
-- Security posture matches the underlying table (no security_invoker), identical
-- to 0242 / 0243 / 0245.

BEGIN;

CREATE VIEW metaldocs.v_process_area_name AS
SELECT tenant_id,
       code AS area_code,
       name AS area_name
  FROM metaldocs.document_process_areas;

COMMENT ON VIEW metaldocs.v_process_area_name IS
  'Published taxonomy per-area human-label read contract (ADR-0041; ADR-0039 D3a/D4): one (tenant_id, area_code, area_name) row per (tenant_id, area_code), 1:1 projection of metaldocs.document_process_areas (code → area_code, name → area_name). No is_active / archived_at filter — labels resolve for archived areas (parity with internal/modules/taxonomy/infrastructure/area_catalog_reader.go). Sole consumer: distribution module (F2.2), joining on (tenant_id, area_code) to F2.1a''s v_cd_obligated_readers. Mission frontend-screen-completion M2/F2.1b.';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0246', 'taxonomy publishes metaldocs.v_process_area_name per-area label view (M2/F2.1b, ADR-0041)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
```

- [ ] **Step 2: Apply the migration on a fresh template DB and re-run the tests**

Run:
```bash
.\scripts\start-api.ps1 -Build
```
Expected: all migrations apply green; logs show `0246 taxonomy publishes metaldocs.v_process_area_name …`.

Then:
```bash
go test -tags=integration -run "TestProcessAreaName" ./internal/modules/taxonomy/infrastructure/... -v
```
Expected: PASS on all three tests.

- [ ] **Step 3: Idempotency check — re-running the migration is a no-op**

Run the API start script a second time (it re-applies the ledger):
```bash
.\scripts\start-api.ps1
```
Then verify the ledger has exactly one row:
```bash
psql "$env:DATABASE_URL" -c "SELECT version, description FROM public.schema_migrations WHERE version='0246'"
```
Expected: exactly **one** row with version `0246`.

- [ ] **Step 4: Verify taxonomy base table untouched + no Go runtime drift**

Run:
```bash
git diff -- internal/modules/taxonomy
git diff db/migrations/0228_authz_p11_reserve_tenant_area_code.sql
```
Both expected: empty (the only taxonomy-tree change should be the new test file from Task 1; the only `db/migrations/` change in this feature is `0246_taxonomy_process_area_name_view.sql`).

- [ ] **Step 5: Commit migration**

```bash
git add db/migrations/0246_taxonomy_process_area_name_view.sql
git commit -m "feat(M2/F2.1b): publish metaldocs.v_process_area_name view (ADR-0041)"
```

---

### Task 3: ADR-0041 + ADR-0039 inventory bullet

**Files:**
- Create: `wiki/decisions/0041-taxonomy-process-area-name-view.md`
- Modify: `wiki/decisions/0039-cross-module-base-table-read-boundary.md` (add a `Related code` bullet pointing at the new migration; flip the front-matter `Last verified` date if earlier than today)
- Modify: `wiki/decisions/index.md` (insert the ADR-0041 entry per existing format) — verify against the file before editing; skip if the index is auto-generated.

- [ ] **Step 1: Write ADR-0041**

```markdown
# ADR 0041 — taxonomy publishes `metaldocs.v_process_area_name` (per-area label read contract)

> **Status:** Accepted 2026-06-21
> **Last verified:** 2026-06-21
> **Deciders:** leandrotca.work (operator), MetalDocs backend
> **Context window:** Mission `frontend-screen-completion` · Milestone M2 (Distribuição coverage-scope) · Feature F2.1b (re-decomposition under HS-6 path A).
> **Related ADRs:** [0039 — Cross-module read boundary](./0039-cross-module-base-table-read-boundary.md) (this view is a D3(a) exemption — the compliant mechanism distribution uses to read area names); [0040 — `v_cd_obligated_readers`](./0040-cd-obligated-readers-view.md) (sibling published view F2.2 joins on `(tenant_id, area_code)`); ADR-0042 (distribution module + denominator-only contract, F2.1c).
> **Related code (Last verified 2026-06-21):**
> - `db/migrations/0246_taxonomy_process_area_name_view.sql` — this view.
> - `internal/modules/taxonomy/infrastructure/area_catalog_reader.go:32` — the existing taxonomy port reading the same base table with the same (no-filter) semantic.
> - `db/migrations/0245_cd_obligated_readers_view.sql` — sibling F2.1a view the F2.2 consumer joins on `(tenant_id, area_code)`.

## Context

M2/F2.1c's `DistributionRecipient` and `DistributionAreaCoverage` schemas carry an `area_name` field (the human label rendered next to each recipient + per-area total on the Distribuição screen). The base table holding it is `metaldocs.document_process_areas` (`name text NOT NULL`), owned by the taxonomy module.

ADR-0039 forbids the new `distribution` module from reading taxonomy's base table raw (`hgcrossmodule` H-G violation). Taxonomy must therefore publish a read contract.

Recon (this session, HEAD post-F2.1a): no existing `metaldocs.v_*` view exposes `area_name`. The taxonomy module has only one published port today — the in-Go `AreaCatalogReader` port (ADR-0039 D3b) — which serves single-area lookups inside the documents create tx and the iam pre-invite check; it is not appropriate for the distribution module's set-shaped `JOIN` use case (distribution joins thousands of obligated-reader rows to area names in a single query, not one lookup at a time).

## Decision

### D1 — Publish a minimal sibling view

Migration 0246 creates `metaldocs.v_process_area_name` as a plain `CREATE VIEW` over `metaldocs.document_process_areas`. No new base table; no Go code; no port change. The taxonomy in-Go port (`AreaCatalogReader`) is untouched — it continues to serve its existing two callers (B7 in-tx documents create, B8 off-tx iam pre-invite).

### D2 — Shape

```
metaldocs.v_process_area_name (
  tenant_id  uuid not null,
  area_code  text not null,
  area_name  text not null
)
-- One row per (tenant_id, area_code). 1:1 projection of
-- metaldocs.document_process_areas.
```

Renames: `code → area_code`, `name → area_name`. The `area_code` rename aligns the shape with F2.1a's `v_cd_obligated_readers.area_code` so the F2.2 join is natural; the `area_name` rename disambiguates "name" at the consumer (distribution may project area names alongside user names).

### D3 — No `is_active` / `archived_at` filter

The existing taxonomy port (`internal/modules/taxonomy/infrastructure/area_catalog_reader.go:32`) reads the base table without any active/archived filter — names resolve for archived areas too. The view preserves that semantic so label rendering remains consistent across both reader paths. Adding a filter now would change behavior at one of two readers and is YAGNI (spec.md Q5: minimal contract, additive later if a real consumer surfaces).

### D4 — No additional columns

`description`, `parent_code`, `owner_user_id`, `default_approver_role`, `archived_at`, `is_active`, `created_at` are deliberately omitted. F2.1c's contract consumes only `area_name`. Adding columns now would grow the published surface beyond its consumer's need and complicate future additive evolution.

### D5 — Security posture

Plain `CREATE VIEW` (no `security_invoker`) — identical to `v_active_user_areas` (0242), `v_cd_grantee` (0243), `v_document_search_facts` (0244), and `v_cd_obligated_readers` (0245). RLS over the base table continues to apply unchanged.

## Consequences

- The `distribution` module (F2.1c/F2.2) is ADR-0039 D3a compliant by construction — it reads only published views (`v_cd_obligated_readers`, `v_process_area_name`) + the ADR-0029 iam display-name port; `hgcrossmodule` analyzer holds.
- The taxonomy in-Go port (`AreaCatalogReader`) is unchanged — no risk to the existing B7 in-tx + B8 off-tx callers.
- The contract is **forward-compatible** for the parked `document-distribution-mission`: that mission may add columns to the consumer DTO without altering the published view shape.

## Verification

- Migration applies cleanly + idempotent on re-run (the `schema_migrations` ledger enforces this).
- `internal/modules/taxonomy/infrastructure/v_process_area_name_integration_test.go` asserts view shape + 1:1 per-tenant projection + cross-tenant isolation against a fixtured graph.
- `git diff -- internal/modules/taxonomy` consists only of the new test file (no runtime Go code).
- `go run ./tools/cilint/...` = 0 H-G under both `taxonomy` (publishes) and `distribution` (reads, once F2.2 lands).
```

- [ ] **Step 2: Add the ADR-0039 inventory bullet**

Open `wiki/decisions/0039-cross-module-base-table-read-boundary.md`. Locate the `Related code (Last verified …)` list at the top of the front matter — it currently ends with the `0245_cd_obligated_readers_view.sql` bullet. Append, on a new line immediately after that bullet (keeping the same `> - ` prefix):

```
> - `db/migrations/0246_taxonomy_process_area_name_view.sql` — the D3(a)/D4 taxonomy-published per-area label read contract, built in mission `frontend-screen-completion` M2/F2.1b (ADR-0041): `metaldocs.v_process_area_name` (one `(tenant_id, area_code, area_name)` row per `(tenant_id, area_code)`, 1:1 projection of `metaldocs.document_process_areas` with `code → area_code` + `name → area_name`; no `is_active` / `archived_at` filter — parity with the existing in-Go `AreaCatalogReader` port). Consumer module: `distribution` (built in M2/F2.1c + F2.2) — joins this view to `v_cd_obligated_readers` on `(tenant_id, area_code)` to populate `DistributionRecipient.area_name` + `DistributionAreaCoverage.area_name`.
```

Also flip the front-matter `Last verified` date to `2026-06-21` if it was earlier.

- [ ] **Step 3: Verify the inventory bullet renders and the ADR file is reachable**

Run:
```bash
grep -n "v_process_area_name" wiki/decisions/0039-cross-module-base-table-read-boundary.md
ls wiki/decisions/0041-taxonomy-process-area-name-view.md
```
Expected: at least one hit in 0039; the 0041 file exists.

- [ ] **Step 4: Update the ADR index entry (if the index is hand-maintained)**

Read `wiki/decisions/index.md`. If it has a numbered list of ADRs in order, insert an entry for `0041` between 0040 and the next tail. **If the file is auto-generated (has a banner saying so), skip — do not hand-edit.**

- [ ] **Step 5: Commit ADRs**

```bash
git add wiki/decisions/0041-taxonomy-process-area-name-view.md wiki/decisions/0039-cross-module-base-table-read-boundary.md
git add wiki/decisions/index.md  # only if step 4 modified it
git commit -m "docs(adr): ADR-0041 v_process_area_name + ADR-0039 inventory (M2/F2.1b)"
```

---

### Task 4: Run the full Validation Gate

Mirrors `spec.md` §Validation Gate row-by-row. No code changes here — just evidence collection.

- [ ] **Step 1: Migration applies + idempotent (re-confirm)**

```bash
psql "$env:DATABASE_URL" -c "SELECT version FROM public.schema_migrations WHERE version='0246'"
```
Expected: exactly one row.

- [ ] **Step 2: View shape exact**

```bash
psql "$env:DATABASE_URL" -c "\d+ metaldocs.v_process_area_name"
```
Expected: 3 columns — `tenant_id uuid NOT NULL`, `area_code text NOT NULL`, `area_name text NOT NULL`.

- [ ] **Step 3: 1:1 per-tenant integration test green (re-run)**

```bash
go test -tags=integration -run "TestProcessAreaName" ./internal/modules/taxonomy/infrastructure/... -v
```
Expected: PASS — `TestProcessAreaName_OneToOneProjection_PerTenant`, `TestProcessAreaName_CrossTenantIsolation`, `TestProcessAreaName_ViewShape`.

- [ ] **Step 4: Taxonomy runtime + sibling F2.1a artifacts untouched**

```bash
git diff origin/main -- internal/modules/taxonomy ':(exclude)internal/modules/taxonomy/infrastructure/v_process_area_name_integration_test.go'
git diff origin/main -- db/migrations/0245_cd_obligated_readers_view.sql
```
Both expected: empty.

- [ ] **Step 5: `hgcrossmodule` analyzer green**

```bash
go run ./tools/cilint/...
```
Expected: 0 H-G findings. The new view publishes from taxonomy's own base table — fully compliant.

- [ ] **Step 6: Build + vet + non-integration tests green (regression)**

```bash
go build ./...
go vet ./...
go test ./...
```
Expected: all green.

- [ ] **Step 7: Capture evidence**

Copy `.claude/skills/milestone/templates/feature-evidence.md` into `docs/superpowers/milestones/frontend-screen-completion/milestone-2-distribuicao/f2.1b-taxonomy-area-name-view/evidence.md` and fill the rows with the exact commands run + their real output (Steps 1–6 above). Label every piece of proof `real` (no fixtures used at the gate level — the integration test uses a fixtured graph but the view + migration evidence is real-DB).

- [ ] **Step 8: Commit evidence + close the feature**

```bash
git add docs/superpowers/milestones/frontend-screen-completion/milestone-2-distribuicao/f2.1b-taxonomy-area-name-view/evidence.md
git commit -m "docs(M2/F2.1b): close evidence — v_process_area_name gate green"
```

---

## Execution notes

Filled during `superpowers:subagent-driven-development`:

- Model: Sonnet 4.6 per `[[workflow-model-balancing]]` + operator directive 2026-06-21.
- Deviations from plan: <none yet>.
- Open questions answered: <none yet>.
