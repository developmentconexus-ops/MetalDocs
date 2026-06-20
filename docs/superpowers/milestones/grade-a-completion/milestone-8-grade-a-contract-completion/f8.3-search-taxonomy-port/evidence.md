# Feature F8.3 — Evidence (search → taxonomy `FamilyCodeResolver` port; H-G closure)

> **Milestone:** 8  ·  **Feature:** `f8.3-search-taxonomy-port`  ·  **Closed:** 2026-06-20
> **Contract:** `spec.md` (approved 2026-06-20). Plan: `plan.md`. ADR: [`wiki/decisions/0038-family-code-resolver-port.md`](../../../../../wiki/decisions/0038-family-code-resolver-port.md).
> **Commit:** recorded at commit time below.

## What was implemented

- **Port (taxonomy/domain)** ([`internal/modules/taxonomy/domain/family_code_resolver_port.go`](../../../../../internal/modules/taxonomy/domain/family_code_resolver_port.go)) —
  NEW `FamilyCodeResolver` interface (two batch methods) + `NoopFamilyCodeResolver` null-object. Provider-owned
  port (ADR 0031 `TenantUserReader` pattern): lives in the module that **owns** `document_profiles`.
- **Impl (taxonomy/infrastructure)** ([`internal/modules/taxonomy/infrastructure/family_code_resolver.go`](../../../../../internal/modules/taxonomy/infrastructure/family_code_resolver.go)) —
  NEW Postgres `FamilyCodeResolverRepository`: `ResolveFamilyCodes` (`DISTINCT ON (code)` over `tenant_id IN
  (tenant, sentinel)`, `code = ANY`) and `ProfileCodesForFamily` (precedence-resolved family, case-insensitive).
  Reads from the pool, never a lock-holding tx (H-PRE-1). No `archived_at` filter (parity).
- **Search reader** ([`internal/modules/search/infrastructure/v2documents/reader.go`](../../../../../internal/modules/search/infrastructure/v2documents/reader.go)) —
  **both** raw `metaldocs.document_profiles` correlated subqueries removed. Projection: column 5 now selects
  the raw join key `COALESCE(d.profile_code_snapshot, cd.profile_code)`; family is batch-resolved in Go
  post-scan via `ResolveFamilyCodes`. Filter: family resolved to its profile-code set up front
  (`ProfileCodesForFamily`) and pushed into SQL as `= ANY($14)` — **pagination stays in SQL**.
- **Composition root** ([`apps/api/cmd/metaldocs-api/main.go:207`](../../../../../apps/api/cmd/metaldocs-api/main.go)) —
  `searchdocs.NewReader(deps.SQLDB, taxonomyinfra.NewFamilyCodeResolverRepository(deps.SQLDB))`.

**Dependency direction:** search imports the taxonomy *interface* and depends on it; taxonomy supplies the impl;
wired at the root. Search no longer encodes taxonomy's storage shape, sentinel rule, or precedence.

## Architecture-truth discovery (HS-6 surfaced + resolved in-boundary)

Integration testing revealed the seed spec (and the original reader subquery) **overclaimed** sentinel-tenant
*precedence*. `metaldocs.document_profiles_pkey` is `PRIMARY KEY (code)` — a **global** key
(`db/baseline/0001_current_schema.sql:2403-2407`); `tenant_id` merely DEFAULTs to the sentinel
(`0001:1130`). So a single `code` exists under **exactly one** tenant — two rows (tenant + sentinel) with the
same code **cannot coexist**, and the `ORDER BY CASE WHEN tenant_id = $tenant …` tie-break can never fire.

Disposition (in F8.3 boundary — this *is* the read being ported, so no HS-2 redesign):
- **Production SQL unchanged.** The ORDER BY is retained verbatim — exact parity with the replaced subquery,
  and forward-safe if the PK is ever widened to `(tenant_id, code)`.
- **Docs corrected to runtime truth.** ADR 0038 §Semantics, the port doc-comment, and the impl `sentinelTenantID`
  comment now state precedence is *defensive / currently unreachable*; the **sentinel fallback** (a code owned
  by the sentinel resolving for any tenant) is the live, reachable behavior.
- **Invariant pinned by test.** `TestFamilyCodeResolverRepository_CodeIsGlobalPrimaryKey` asserts the PK is
  `(code)`; if widened, it fails and flags the precedence as newly-live behavior to re-test.

## Wire-equivalence (one honest caveat)

family_code projection moved from a per-row SQL `COALESCE(subquery,'')` to a Go batch map lookup keyed on the
same `COALESCE(d.profile_code_snapshot, cd.profile_code)`. Missing keys default to `""` (map absence) — identical
to the old `COALESCE(...,'')`. The filter moved from a correlated subquery predicate to `key = ANY($14)` over the
precedence-resolved code set, evaluated **in the same WHERE/LIMIT/OFFSET** — page contents and counts identical.
This is wire-equivalent (M7 F7.1 precedent), not byte-identical SQL.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| H-G red→green: no cross-schema SQL read in search | `grep -rn 'metaldocs.document_profiles' internal/modules/search/` | **0 matches** (was 2 subqueries in `reader.go`) | real |
| Bare-token gate (F8.6 widened) | `grep -rn 'document_profiles' internal/modules/search/` | **0 matches** | real |
| Search imports a taxonomy port (no raw cross-schema read) | read `reader.go` imports + `NewReader` signature | imports `taxonomydomain`; ctor takes `FamilyCodeResolver` | real |
| Port impl: ownership/fallback/archived/case/filter | `go test -tags integration ./internal/modules/taxonomy/infrastructure/ -run TestFamilyCodeResolverRepository` | **6/6 PASS** (ownership-resolve, global-PK invariant, sentinel-fallback, archived-included, exact-case, ProfileCodesForFamily) (104.6s) | real (docker `metaldocs-postgres`) |
| Reader projection + filter + pagination parity | `go test -tags integration ./internal/modules/search/infrastructure/v2documents/ -run TestListDocuments_Family` | **3/3 PASS** (FamilyProjection, FamilyFilter, FamilyFilterPagination) | real |
| Visibility regression (existing) | same package, `-run TestListDocuments_EnforcesUnifiedVisibility` | **PASS** (4.68s) — wired to real resolver | real |
| Unit suite (search + taxonomy) | `go test -count=1 ./internal/modules/search/... ./internal/modules/taxonomy/...` | all `ok` (incl. updated sqlmock tests) | real |
| Static (build + vet) | `go build ./...`; `go vet -tags integration <pkgs>` | exit 0, no findings | — |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| No `document_profiles` SQL in search | yes | `metaldocs.document_profiles` grep 2→0; bare grep 0 |
| Search imports a taxonomy port (no raw cross-schema read) | yes | `reader.go` imports `taxonomydomain`; `NewReader(db, FamilyCodeResolver)` |
| Family projection + filter results wire-equivalent | yes | 3 reader integration tests + 6 port tests, real DB |
| Honest H-G = 0 (for this read) | yes | widened grep 0 (F8.6 re-runs the full sweep) |

## Review disposition

- Spec-compliance review: PASS — H-G closed; provider-owned port (ADR 0031 mirror); H-PRE-1 honored (pool read);
  non-goals untouched (no schema/migration, no result-shape/ordering/visibility change, no new taxonomy capability).
- Code-quality review: PASS — `pq.Array` for `= ANY`; dedup in `ResolveFamilyCodes`; Noop null-object for the
  nil resolver; pre-existing sqlmock unit tests updated for the new signature ($14 + stub resolver) as drive-by
  repair (CLAUDE.md §4); precedence honesty corrected across ADR + code comments + a pinning test.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| None | | |
