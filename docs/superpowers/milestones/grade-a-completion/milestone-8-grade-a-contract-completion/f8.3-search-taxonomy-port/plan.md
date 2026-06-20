# Feature F8.3 — Plan

> Engine: inline (superpowers:writing-plans). Spec: `./spec.md` (approved 2026-06-20). ADR 0038.
> Mirrors ADR 0031 `TenantUserReader` H-G remediation.

## Files touched

| File | Change |
|------|--------|
| `internal/modules/taxonomy/domain/family_code_resolver_port.go` | NEW — `FamilyCodeResolver` interface (2 methods) + `NoopFamilyCodeResolver` null-object. |
| `internal/modules/taxonomy/infrastructure/family_code_resolver.go` | NEW — `FamilyCodeResolverRepository` Postgres impl: `ResolveFamilyCodes` (DISTINCT ON precedence, `code = ANY`), `ProfileCodesForFamily` (precedence-resolved family = family). Reads from pool (H-PRE-1). No `archived_at` filter. |
| `internal/modules/search/infrastructure/v2documents/reader.go` | Inject `taxonomydomain.FamilyCodeResolver`; remove both `metaldocs.document_profiles` subqueries; SELECT raw `d.profile_code_snapshot`,`cd.profile_code` for the key; resolve filter codes up front → `= ANY($14)`; batch-resolve family projection in Go post-scan. |
| `apps/api/cmd/metaldocs-api/main.go` | Construct `taxonomyinfra.NewFamilyCodeResolverRepository(sqlDB)`; pass into `searchdocs.NewReader(db, resolver)`. |
| `internal/modules/taxonomy/infrastructure/family_code_resolver_integration_test.go` | NEW — port impl: precedence, sentinel fallback, archived-included, filter inverse. |
| `internal/modules/search/infrastructure/v2documents/reader_family_integration_test.go` | NEW — projection + filter parity (tenant + sentinel profiles, with/without family filter, pagination). |

## Test strategy

- **Class:** DB integration (`//go:build integration`, `tests/integration/testdb`) — the only way to prove
  byte-identical against the real schema + sentinel fallback. testdb reachable (docker `metaldocs-postgres`).
- **red→green:** `grep -rn 'document_profiles' internal/modules/search` → 1 file (`reader.go`) before, 0 after.
- Port integration test: seed tenant + sentinel profiles for a code (different families) → assert precedence;
  archived profile still resolves; `ProfileCodesForFamily` returns the precedence-correct code set.
- Reader integration test: seed documents whose family comes from tenant profile, sentinel fallback, and a
  null snapshot (cd.profile_code fallback); assert `family_code` projection per row; assert the family filter
  returns exactly the matching docs and that pagination (`LIMIT/OFFSET`) is unaffected.

## Task order

1. Confirm grep baseline (`document_profiles` in search → reader.go).
2. Port interface + Noop (taxonomy/domain).
3. Postgres impl (taxonomy/infrastructure) + its integration test → green.
4. Rework reader.go (key columns, projection batch, filter ANY) + inject resolver.
5. Reader integration test (projection + filter parity, pagination) → green.
6. Wire main.go.
7. `go build ./...`; `go test -count=1 ./...` (unit) + `-tags integration` for the two new tests + the
   existing visibility test (regression); grep = 0.
8. Evidence + commit.

## Risk / rollback

- **Highest-risk feature** (visibility-critical paginated query). Mitigations: filter stays in SQL (pagination
  preserved); parity pinned by integration tests vs the documented original semantics; existing
  `reader_visibility_integration_test.go` must stay green (regression).
- Rollback = `git checkout` the 6 files (+ ADR). No schema/migration change to revert.
