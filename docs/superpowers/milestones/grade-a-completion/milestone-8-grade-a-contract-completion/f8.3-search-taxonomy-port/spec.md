# Feature F8.3 — Spec (SEED — approval pending)

> **Milestone:** 8 — Grade-A Contract & Boundary Completion  ·  **Folder:** `f8.3-search-taxonomy-port`
> **Status:** Drafting (seed from post-M7 re-audit Module-boundaries Major #3 / honest H-G; **interview + approval pending**)
> **Approved before code:** PENDING — *no implementation begins until this line is filled (Phase 3, fresh session).*
> **Long pole of M8.** Heaviest feature; operator chose option A (close the class).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Close the coupling (port) or sanction it (ADR allowlist)? | **Close it** (operator decision 2026-06-20): add a taxonomy read-port; search drops the raw `metaldocs.document_profiles` SQL. |
| 2 | How to preserve the sentinel-tenant fallback + the family **filter** path? | Port resolves `profileCode → familyCode` honoring the `ffffffff-…` global fallback (same `ORDER BY` precedence); the `$5` family filter resolves family→profile-codes via the port, applied in Go or pushed as a resolved code set. Result byte-identical. |
| 3 | Perf concern (inline subquery → batch lookup)? | Batch-resolve distinct profile codes per page; acceptable for search page sizes. Confirm with a benchmark/listing test in execution. |

## Consumer contract (FIRST)

- **Consumer(s):** `search` list query (`search/.../v2documents/reader.go`) needs each document's `family_code` (projection) and a family filter.
- **Contract:** a taxonomy-owned port:
  ```go
  // taxonomy/domain
  type FamilyCodeResolver interface {
      ResolveFamilyCodes(ctx context.Context, tenantID string, codes []ProfileCode) (map[ProfileCode]FamilyCode, error)
  }
  ```
  honoring the sentinel-tenant (`ffffffff-ffff-ffff-ffff-ffffffffffff`) global fallback with tenant-own precedence. Search depends on this interface; taxonomy supplies the implementation.
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
| Family projection + filter results byte-identical | search listing test (with/without family filter; tenant + global-fallback profiles) | real |
| Honest H-G = 0 | widened H-G grep (F8.6) returns 0 | real |

## ADR needed?

- [ ] No durable decision — skip.
- [x] **Durable decision** — new cross-module taxonomy read-port consumed by search → record an ADR under `wiki/decisions/` (port ownership, sentinel-fallback semantics, batch contract); link here: _TBD in execution_.
