# Feature F3 — Evidence

> **Milestone:** 2b — Approval Kernel Backend  ·  **Feature:** `f3-capabilities-review-oversee`  ·
> **Closed:** 2026-07-07
> **Contract:** `spec.md`

## What was implemented

- `internal/modules/iam/domain/model.go`: two new capability consts, `CapApprovalReview =
  "approval.review"` and `CapApprovalOversee = "approval.oversee"`, both added to
  `validCapabilities`.
- `internal/modules/iam/domain/capability_scope.go`: `CapApprovalReview` classified `ScopeArea`,
  `CapApprovalOversee` classified `ScopeTenant`.
- `internal/modules/iam/domain/catalog.go`: pt-BR descriptions added for both.
- `internal/modules/iam/domain/model_test.go`: `TestCapabilityRegistrySize` `want` 38 → 40.
- `internal/modules/iam/domain/capability_scope_test.go`: `TestAreaGradeCapabilitySet`'s locked
  `wantArea` map gains `CapApprovalReview` (area-grade set 11 → 12).
- `apps/api/cmd/metaldocs-api/permissions.go`: deleted the 4 generic `/api/v1/approval/` prefix
  rows (P1); replaced with 4 explicit rows, one per real registered runtime route — signoff →
  `CapDocumentSignoff`, cancel → `CapDocumentEdit`, get-instance → `CapDocumentView`, inbox →
  `CapDocumentView`. Route-admin rows untouched (BE-9 already closed that gap).
- `internal/modules/documents/approval/application/read_service.go`: `LoadInstance` and
  `LoadActiveInstanceByDocument` each now try `authz.Require(CapDocumentView, "tenant")` first, and
  on failure try `authz.Require(CapApprovalOversee, "tenant")` — the first error is returned only if
  both fail. Explicit two-capability check, never a role check.
- `db/reference-data/0001_product_reference_data.sql`: 7 new `role_capabilities` seed rows —
  `approval.review` → `approver`/`area_admin`/`qms_admin`/`signer`/`system_admin` (mirrors the
  existing `document.signoff` pool); `approval.oversee` → `qms_admin`/`system_admin`.
- `scripts/api-lint/registry_rules_test.go`: `expectedRoleCapabilityRows` golden count 108 → 115
  (+7 new seed rows).
- ADR `wiki/decisions/0075-approval-oversee-visibility.md` (new); `wiki/decisions/index.md` updated
  (also backfilled the missing 0074 row, an F2 wiki-sync gap found in passing).
- No `db/migrations/*.sql` file — seed grants are reference-data only (verified: zero
  `role_capabilities` INSERT rows exist in any `db/migrations/*.sql` file, for any capability, ever).
- No `internal/platform/tripwire/arms.go` change — neither new capability needs a write-tripwire arm
  in this feature (see Acceptance table).
- Two pre-existing stale test fixtures in `apps/api/cmd/metaldocs-api/permissions_test.go` corrected
  to runtime truth (see Review disposition).

## Verification

| Check | Command | Result |
|-------|---------|--------|
| Build | `go build ./...` | clean, exit 0 |
| IAM domain suite | `go test ./internal/modules/iam/...` | all PASS (`TestEveryCapabilityClassified`, `TestAreaGradeCapabilitySet`, `TestCapabilityRegistrySize` green) |
| Permissions/tier-1 suite | `go test ./apps/api/cmd/metaldocs-api/...` | PASS — `TestPermissionResolver`, `TestRouteCoverage`, `TestEveryCapSeededOrDeferred`, `TestTier1Tier2CapabilityCoherence_F4Sites` all green |
| Approval module suite | `go test ./internal/modules/documents/approval/...` | all subpackages PASS |
| api-lint strict (TRIPWIRE-ARM-DRIFT + TRIPWIRE-ARM-PARITY + all other lints) | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` |
| api-lint unit suite (incl. golden seed-row-count parser) | `go test ./scripts/api-lint/...` | PASS (`TestSeedRowCount_GoldenMatchesParser` green post-bump to 115) |
| Integration tag build | `go build -tags integration ./...` | clean, exit 0 |
| Tripwire regression (packages that compile) | `go test -tags integration -run 'Tripwire' ./tests/integration/...` | `tests/integration/iam` PASS (tripwire tests ran and passed — `tenants_tripwire_test.go`, `iam_users_tripwire_test.go`); `tests/integration/approval`, `audit`, `migrations`, `scenarios`, `security`, `tenantdata`, `testdb` — `[no tests to run]` (no `Tripwire`-named test in these packages, expected); `tests/integration/controlleddocuments`, `tests/integration/documents`, `tests/integration/templates` — pre-existing compile failure, unrelated to this feature (see Bounded defers) |

Zero-arms-changed proof: `git diff --stat internal/platform/tripwire/arms.go` for this feature's
whole diff is empty — both lints pass with no arm added, confirming by construction (not assertion)
that neither new capability needs one yet.

Grep-zero proof (P1, generic prefix fallback removed):
```
$ grep -n 'pathPrefix: "/api/v1/approval/"' apps/api/cmd/metaldocs-api/permissions.go
(no matches — the old bare-prefix catch-all rows are gone)

$ grep -n 'api/v1/approval' apps/api/cmd/metaldocs-api/permissions.go
245:	{method: http.MethodGet, pathPrefix: "/api/v1/approval/routes", capability: iamdomain.CapRouteManage, ...}   # BE-9, untouched
246:	{method: http.MethodPost, pathPrefix: "/api/v1/approval/routes", capability: iamdomain.CapRouteManage, ...}  # BE-9, untouched
247:	{method: http.MethodPut, pathPrefix: "/api/v1/approval/routes", capability: iamdomain.CapRouteManage, ...}   # BE-9, untouched
255:	{method: http.MethodPost, pathPrefix: "/api/v1/approval/instances/", pathSuffix: "/signoffs", capability: iamdomain.CapDocumentSignoff, ...}
256:	{method: http.MethodPost, pathPrefix: "/api/v1/approval/instances/", pathSuffix: "/cancel", capability: iamdomain.CapDocumentEdit, ...}
257:	{method: http.MethodGet, pathPrefix: "/api/v1/approval/instances/", capability: iamdomain.CapDocumentView, ...}
258:	{method: http.MethodGet, pathExact: "/api/v1/approval/inbox", capability: iamdomain.CapDocumentView, ...}
```
Only the 3 pre-existing route-admin rows (BE-9, unchanged) plus the 4 new explicit runtime-verb rows
remain; no bare catch-all prefix row survives.

## Acceptance vs spec Validation Gate

| Gate item | Met? | Evidence |
|-----------|------|----------|
| Registry size 38 → 40 | yes | `TestCapabilityRegistrySize` |
| `CapApprovalReview`/`CapApprovalOversee` classified (area/tenant) | yes | `TestEveryCapabilityClassified`, `TestAreaGradeCapabilitySet` |
| Generic `/api/v1/approval/` prefix fallback removed; zero grep hits for the old catch-all rows | yes | grep-zero proof above |
| `permissions_test.go` asserts explicit rows for every real runtime verb, matching each route's real tier-2 capability | yes | `TestPermissionResolver` cases: approval get instance/signoff/cancel/inbox |
| Oversight reads accept `CapApprovalOversee` as an explicit alternative to `CapDocumentView`, never a role check | yes | `read_service.go` two-`authz.Require` sequence; approval module suite green |
| Both authz lints (TRIPWIRE-ARM-DRIFT, TRIPWIRE-ARM-PARITY) pass with zero arm change | yes | `api-lint -strict` `0 violation(s)`; empty `arms.go` diff |
| No manual `arms.go` edit | yes | diff-stat empty |
| No regression | yes | full sweep above; only 2 intentional fixture corrections (documented below), both proven against real router truth |

## Review disposition

- Spec-compliance review: contract matched — capability naming avoids the `document.review` (ADR
  0069) collision as required; cancel's tier-2 stays `CapDocumentEdit` (not reassigned to
  `approval.oversee`, per the Interview record's #2 answer — that's F4/F8 scope); `CapApprovalReview`
  correctly ships with no live tier-1 route (deferred to F4, documented in spec.md's Bounded defers,
  not a silent gap).
- Two pre-existing stale test-fixture regressions found and root-caused, not symptom-patched:
  deleting the generic prefix fallback exposed that `permissions_test.go` asserted 3 synthetic
  routes (`PUT`/`DELETE /instances/{id}`, `POST /instances/{id}/decisions`) that were **never
  actually registered** by the real router — verified against
  `internal/modules/documents/approval/http/router.go` and its handler-registration tests. These
  only ever "passed" because the now-deleted fallback silently caught any method/path under the
  prefix. Fixed by replacing them, in both `TestPermissionResolver`'s case slice and the
  `registeredRoutes` slice feeding `TestRouteCoverage`, with the 4 real routes and their real
  capabilities — not by preserving the fictitious assertions. A stale line-number comment
  ("main.go:396") was also corrected to the real call site (`main.go:753`), re-verified by direct
  read rather than trusted from the old comment.
- Seed-grant role mapping is a judgment call, recorded in ADR 0075: the design's persona language
  ("reviewer pools", "quality-manager profile") doesn't map 1:1 to MetalDocs's 8 canonical roles
  (`reviewer` is explicitly decommissioned per `TestIsAreaRole`). `approval.review` mirrors the
  existing `document.signoff` pool exactly; `approval.oversee` goes to `qms_admin` (closest existing
  role to "quality-manager") + `system_admin`.
- Confirmed seed grants are reference-data only, never a migration file: grepped
  `db/migrations/*.sql` for any `role_capabilities` INSERT and found zero across the entire
  migration history — an initially-authored `db/migrations/0288_approval_caps_seed_tripwire.sql`
  was deleted after this check as the wrong artifact shape for seed grants.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|--------------------------|
| `CapApprovalReview` has no live tier-1 route in this feature | Its only consumer (`review-verdict`) doesn't exist yet | F4 introduces `POST /approval-instances/{id}/stages/{stageId}/review-verdict` and wires the tier-1 row + tier-2 `authz.Require(CapApprovalReview, area)` call together |
| Worklist `?scope=oversee` variant not implemented | Out of this feature's scope per spec.md non-goals | F8 (SLA + visibility gating + worklist) |
| `tests/integration/controlleddocuments`, `tests/integration/documents`, `tests/integration/templates` fail to compile (`package metaldocs/internal/modules/documents/repository is not in std`) | Pre-existing breakage, unrelated to this feature — these packages reference a `repository` path that no longer exists post the `repository/`→`infrastructure/` rename (ADR 0073); not touched or caused by F3 | Flag for a dedicated fix outside this milestone's scope; not a gate for F3's close since `tests/integration/approval` and `tests/integration/iam` (the packages this feature actually touches) build and pass clean |
