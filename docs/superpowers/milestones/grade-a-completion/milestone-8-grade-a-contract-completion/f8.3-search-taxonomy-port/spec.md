# Feature F8.3 — Spec (APPROVED)

> **Milestone:** 8 — Grade-A Contract & Boundary Completion  ·  **Folder:** `f8.3-search-taxonomy-port`
> **Status:** Approved 2026-06-20 (execution session) — seed confirmed against `reader.go:38-70`, taxonomy `repository.go`, ADR 0031 `TenantUserReader` precedent. Port refined to **two methods** (projection + filter) so pagination stays in SQL (Q4); no-archived-filter parity pinned (Q5).
> **Approved before code:** ✅ 2026-06-20 — testdb reachable (docker `metaldocs-postgres`), so the wire-equivalence gate is provable against the real schema. No code written before this line.
> **Execution note (2026-06-20):** integration testing revealed `document_profiles.code` is a GLOBAL primary key (baseline `0001:2403`), so the sentinel-tenant *precedence* the seed assumed is **unreachable** (a code can exist under only one tenant). The sentinel *fallback* is reachable. Production SQL keeps the precedence ORDER BY for exact parity/forward-safety; ADR 0038 + code comments corrected to runtime truth; invariant pinned by `TestFamilyCodeResolverRepository_CodeIsGlobalPrimaryKey`. Acceptance relaxed from "byte-identical" to **wire-equivalent** (M7 F7.1 precedent), since the family projection moved from per-row SQL to a Go batch resolve over the identical key.
> **Long pole of M8.** Heaviest feature; operator chose option A (close the class).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Close the coupling (port) or sanction it (ADR allowlist)? | **Close it** (operator decision 2026-06-20): add a taxonomy read-port; search drops the raw `metaldocs.document_profiles` SQL. |
| 2 | How to preserve the sentinel-tenant fallback + the family **filter** path? | Port resolves `profileCode → familyCode` honoring the `ffffffff-…` global fallback (same `ORDER BY` precedence); the `$5` family filter resolves family→profile-codes via the port, applied in Go or pushed as a resolved code set. Result byte-identical. |
| 3 | Perf concern (inline subquery → batch lookup)? | Batch-resolve distinct profile codes per page; acceptable for search page sizes. Confirm with a benchmark/listing test in execution. |
| 4 | Can the **family filter** (`$5`) be applied in Go after the query? | **No** — `$5` is a WHERE predicate participating in `LIMIT/OFFSET` pagination; filtering in Go would change page contents/counts. **Resolution:** the port exposes a second method `ProfileCodesForFamily(tenantID, family) → []ProfileCode` (precedence-resolved); search keeps the filter **in SQL** by replacing the correlated subquery with `COALESCE(d.profile_code_snapshot, cd.profile_code) = ANY($resolvedCodes)`. Pagination semantics preserved exactly. Projection (`family_code` per row) stays a Go-side batch resolve (`ResolveFamilyCodes(codes) → map`), which does not affect pagination. |
| 5 | Does the original subquery filter archived profiles? | **No** — it has no `archived_at` predicate. Both port methods must therefore ignore `archived_at` (unlike `HasActiveProfiles`) to stay byte-identical. Precedence (`tenant_id` over sentinel) replicated via `DISTINCT ON (code) … ORDER BY code, CASE WHEN tenant_id = $tenant THEN 0 ELSE 1 END`. Code match is exact-case (mirrors `dp.code = key`); family compare in the filter is case-insensitive (mirrors `LOWER(family)=$5`). |

## Consumer contract (FIRST)

- **Consumer(s):** `search` list query (`search/.../v2documents/reader.go`) needs each document's `family_code` (projection) and a family filter.
- **Contract:** a taxonomy-owned port (provider-owned, ADR 0031 `TenantUserReader` precedent):
  ```go
  // taxonomy/domain
  type FamilyCodeResolver interface {
      // Projection: precedence-resolved family for each code (missing code → absent).
      ResolveFamilyCodes(ctx context.Context, tenantID string, codes []ProfileCode) (map[ProfileCode]FamilyCode, error)
      // Filter: profile codes whose precedence-resolved family == family (case-insensitive).
      ProfileCodesForFamily(ctx context.Context, tenantID string, family FamilyCode) ([]ProfileCode, error)
  }
  ```
  honoring the sentinel-tenant (`ffffffff-ffff-ffff-ffff-ffffffffffff`) global fallback with tenant-own precedence, **ignoring `archived_at`** (parity). Search depends on this interface; taxonomy supplies the implementation (reads from the pool, never a lock-holding tx — H-PRE-1).
- **Source of truth:** taxonomy owns `metaldocs.document_profiles` (`taxonomy/infrastructure/repository.go`); existing `ProfileRepository.GetByCode` (`taxonomy/domain/port.go:10`) is the single-row precedent.

## What this feature implements

- New `FamilyCodeResolver` port + Postgres impl in taxonomy (batch, sentinel-aware).
- `search/.../v2documents/reader.go:38-45,63-70` — remove both `metaldocs.document_profiles` subqueries; select profile codes only; resolve `family_code` via the injected port; apply the family filter against resolved codes.
- Wire the taxonomy resolver into search's composition root (`search module.go` + `apps/api/cmd/metaldocs-api/main.go`).

## Non-goals (mandatory)

- No taxonomy schema/migration change.
- No change to search result shape, ordering, or visibility predicate ($13).
- No new taxonomy product capability.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| No `document_profiles` SQL in search | `grep -rn 'document_profiles' internal/modules/search` → none | real |
| Search imports a taxonomy port (no raw cross-schema read) | read reader.go imports + constructor | real |
| Family projection + filter results wire-equivalent | search listing test (with/without family filter; tenant + global-fallback profiles) | real |
| Honest H-G = 0 | widened H-G grep (F8.6) returns 0 | real |

## ADR needed?

- [ ] No durable decision — skip.
- [x] **Durable decision** — new cross-module taxonomy read-port consumed by search → ADR recorded: [`wiki/decisions/0038-family-code-resolver-port.md`](../../../../../wiki/decisions/0038-family-code-resolver-port.md) (port ownership, sentinel-fallback precedence, two-method batch contract, `archived_at`-agnostic parity, H-PRE-1).
