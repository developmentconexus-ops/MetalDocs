# ADR 0038 — `FamilyCodeResolver`: cross-module document-profile→family reads go through a taxonomy-owned port

> **Status:** Accepted 2026-06-20
> **Last verified:** 2026-06-20
> **Scope:** How modules other than taxonomy resolve a document profile code to its family code (and the inverse — the profile codes belonging to a family) without reaching into `metaldocs.document_profiles`. The owning module (taxonomy), the port shape (two batch methods), the sentinel-tenant fallback precedence, the `archived_at`-agnostic semantics, and the reads-live / off-tx (H-PRE-1) constraint. Sibling of ADR 0031 (`TenantUserReader`).

## Context

`metaldocs.document_profiles` is owned by the **taxonomy** module (`taxonomy/infrastructure/repository.go`).
The **search** v2 documents reader resolved each row's `family_code` — and applied the family filter — with
two raw correlated subqueries against `metaldocs.document_profiles`
(`search/infrastructure/v2documents/reader.go`):

```sql
COALESCE((
  SELECT dp.family_code FROM metaldocs.document_profiles dp
  WHERE dp.code = COALESCE(d.profile_code_snapshot, cd.profile_code)
    AND dp.tenant_id IN (d.tenant_id, 'ffffffff-…'::uuid)
  ORDER BY CASE WHEN dp.tenant_id = d.tenant_id THEN 0 ELSE 1 END
  LIMIT 1
), '')
```

This is a cross-module / cross-schema reach (the honest **H-G** finding the post-M7 re-audit flagged): a
module reads another module's owned table directly, coupling search to taxonomy's storage shape. It is the
same class ADR 0031 closed for `iam_users`.

## Decision

Search resolves family codes through a **taxonomy-owned port**, never via raw `document_profiles` SQL.

```go
// taxonomy/domain
type FamilyCodeResolver interface {
    // Projection: precedence-resolved family for each requested code.
    // Missing codes are absent from the map (caller defaults to "").
    ResolveFamilyCodes(ctx context.Context, tenantID string, codes []ProfileCode) (map[ProfileCode]FamilyCode, error)
    // Filter: the profile codes whose precedence-resolved family == family
    // (case-insensitive). Empty family or no match → empty slice.
    ProfileCodesForFamily(ctx context.Context, tenantID string, family FamilyCode) ([]ProfileCode, error)
}
```

Two methods, because the family is used two ways and they have different pagination constraints:

- **Projection** (each row's `family_code`) is resolved in Go, batched over the page's distinct profile
  codes — it does not participate in `LIMIT/OFFSET`.
- **Filter** (`?documentFamily=`) is a `WHERE` predicate that participates in pagination, so it **stays in
  SQL**: search replaces the correlated subquery with `COALESCE(d.profile_code_snapshot, cd.profile_code)
  = ANY($codes)`, where `$codes = ProfileCodesForFamily(tenant, family)`. Page contents and counts are
  identical to the subquery form. Doing the filter in Go would silently corrupt pagination.

### Semantics (wire-equivalent to the replaced subqueries)

- **Sentinel fallback:** both methods read `tenant_id IN (tenant, 'ffffffff-ffff-ffff-ffff-ffffffffffff')`,
  so a code owned by the global sentinel tenant resolves for any querying tenant, and a code owned by the
  querying tenant resolves for that tenant. (`document_profiles.tenant_id` DEFAULTs to the sentinel —
  baseline `0001_current_schema.sql:1130` — so most profiles are sentinel-owned and reachable only via the
  fallback arm.)
- **Precedence is defensive / currently unreachable.** The `DISTINCT ON (code) … ORDER BY code, CASE WHEN
  tenant_id = $tenant THEN 0 ELSE 1 END` tie-break preserves the *exact shape* of the replaced subquery, but
  it can never actually fire: `document_profiles_pkey PRIMARY KEY (code)` (baseline
  `0001_current_schema.sql:2403-2407`) makes `code` a **global** primary key, so two rows with the same code
  under different tenants cannot coexist — there is never a tenant-vs-sentinel pair to rank. The ORDER BY is
  retained verbatim (a) for byte-for-byte parity with the original reader subquery and (b) to stay correct if
  the PK is ever widened to `(tenant_id, code)`. **It is not a live behavior**; the schema invariant is
  pinned by a resolver test that asserts seeding a duplicate `code` across tenants raises `23505`.
- **`archived_at`-agnostic:** the original subqueries had **no** `archived_at` predicate, so neither port
  method filters archived profiles (unlike `FamilyRepository.HasActiveProfiles`). This is deliberate parity.
- **Case:** code match is exact (mirrors `dp.code = key`); the filter's family comparison is
  case-insensitive (mirrors `LOWER(family) = $5`).
- **Reads live, off-tx:** the implementation reads from the connection pool, never from a caller's
  lock-holding transaction (H-PRE-1).

### Ownership & dependency direction

The port lives in `taxonomy/domain` (the module that **owns** `document_profiles`); search imports the
interface and depends on it. The Postgres implementation lives in `taxonomy/infrastructure` and is wired
into search's reader at the composition root (`apps/api/cmd/metaldocs-api/main.go`). This mirrors ADR 0031
exactly.

## Consequences

- **Positive:** the honest H-G class for `document_profiles` is closed; search no longer encodes taxonomy's
  storage shape, sentinel-fallback rule, or precedence — that logic has one home. Symmetric with ADR 0031.
- **Cost:** one extra round trip per search page (the projection batch) and one per request when a family
  filter is present (the filter resolve). Both are small bounded lookups (number of profiles per tenant),
  acceptable at search page sizes.
- **Constraint inherited:** callers must not invoke either method inside a lock-holding transaction
  (H-PRE-1); the port reads from the pool.

## Alternatives considered

- **Sanction the coupling via an allowlist ADR** (leave the raw SQL, document it as accepted). Rejected by
  the operator (2026-06-20, option A): close the class, do not grow the allowlist.
- **Single projection-only method, filter in Go.** Rejected: breaks `LIMIT/OFFSET` pagination (see Decision).
- **A taxonomy-side denormalized `family_code` column on documents.** Rejected: a schema/migration change,
  out of M8 appetite and the non-goals.

## Related

- ADR 0031 — `TenantUserReader` (the sibling H-G remediation this mirrors).
- ADR 0037 — soft-delete / active-now membership (the search reader's visibility predicate; untouched here).
- Mission `grade-a-completion` M8 F8.3.
