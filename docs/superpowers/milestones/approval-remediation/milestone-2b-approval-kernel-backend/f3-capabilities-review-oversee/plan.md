# Feature F3 — Plan

> Input: `spec.md` (this folder). Engine: inline — contract fully prescriptive from verified
> runtime truth (permissions.go grep, read_service.go read, tripwire pipeline research), no design
> exploration needed.

## Files touched

- Modify: `internal/modules/iam/domain/model.go` — add `CapApprovalReview`/`CapApprovalOversee`
  const + `validCapabilities` entries.
- Modify: `internal/modules/iam/domain/capability_scope.go` — classify `CapApprovalReview` =
  `ScopeArea`, `CapApprovalOversee` = `ScopeTenant`.
- Modify: `internal/modules/iam/domain/catalog.go` — `capabilityDescriptions` entries (pt-BR),
  category derivation covers the `approval.*` prefix.
- Modify: `apps/api/cmd/metaldocs-api/permissions.go` — delete lines 250-253 (generic `/approval/`
  fallback); add 4 explicit rows: `POST .../instances/{id}/stages/{sid}/signoffs` →
  `CapDocumentSignoff`; `POST .../instances/{id}/cancel` → `CapDocumentEdit`;
  `GET .../instances/{id}` → `CapDocumentView`; `GET /approval/inbox` → `CapDocumentView`.
- Modify: `internal/modules/documents/approval/application/read_service.go` — `LoadInstance` /
  `LoadActiveInstanceByDocument`: replace the single `authz.Require(CapDocumentView, "tenant")` call
  with an explicit two-capability check (`CapDocumentView` OR `CapApprovalOversee`).
- Create: `db/migrations/0288_approval_caps_seed_tripwire.sql` — seed grants only (quality-manager →
  `approval.oversee`; reviewer pools → `approval.review`). **No new tripwire arm** — neither
  capability gates a new INSERT/UPDATE surface yet (`CapApprovalReview`'s write lands in F4 when
  `review-verdict` is created; `CapApprovalOversee` is read-only by contract, tripwires never gate
  reads). Confirmed by re-running the parity/drift lints with zero `arms.go` change (below) —
  expected clean, proving no arm was silently required.
- Modify: `internal/modules/iam/domain/model_test.go` — `TestCapabilityRegistrySize` want 38 → 40;
  the failure-string comment gains ` + F3: +2 approval.review/approval.oversee`.
- Modify: `apps/api/cmd/metaldocs-api/permissions_test.go` — assert the 4 new explicit rows;
  assert the generic `/api/v1/approval/` prefix row is gone.
- Modify: `internal/modules/documents/approval/application/read_service_test.go` — add
  oversee-only-actor-succeeds / no-cap-actor-denied cases for `LoadInstance`.
- Create: `wiki/decisions/00XX-approval-oversee-visibility.md` (ADR 3).

## Ordering (TDD)

1. **Failing tests first.**
   - Bump `TestCapabilityRegistrySize` want to 40 (fails: registry still 38).
   - Add `TestEveryCapabilityClassified`/`TestAreaGradeCapabilitySet` expectations for the two new
     caps (fails: consts don't exist yet — compile-fails first, which is the correct RED for a new
     const).
   - Add `permissions_test.go` assertions for the 4 explicit rows + generic-row-absence (fails:
     rows don't exist, generic row still present).
   - Add `read_service_test.go` oversee-alternative cases (fails: `CapApprovalOversee` doesn't
     exist / read_service doesn't accept it yet).
2. Add the two consts + `validCapabilities` entries + scope classification + catalog descriptions
   (`model.go`, `capability_scope.go`, `catalog.go`). Re-run — registry/classification tests green.
3. `permissions.go`: delete lines 250-253; insert the 4 explicit rows immediately above the
   deleted block's former position (keep the "route-admin must precede generic" ordering comment,
   adjust since the generic block is gone). Re-run `permissions_test.go` — green.
4. `read_service.go`: change `LoadInstance`/`LoadActiveInstanceByDocument` tier-2 gate to the
   explicit two-capability check. Re-run `read_service_test.go` — green.
5. Migration `0288`: seed grants only, per `db/reference-data/0001_product_reference_data.sql`
   dev-seed parity convention (quality-manager profile, reviewer pools).
6. Run both lints with zero `internal/platform/tripwire/arms.go` change to confirm no new arm is
   actually required (proves the bounded-defer reasoning in spec.md, not just asserts it):
   ```
   go run ./scripts/api-lint -only TRIPWIRE-ARM-PARITY api/openapi/v1/openapi.yaml .
   go run ./scripts/api-lint -only TRIPWIRE-ARM-DRIFT api/openapi/v1/openapi.yaml .
   ```
7. Full suite: `go build ./...`; `go test ./internal/modules/iam/... ./apps/api/... ./internal/modules/documents/approval/...`; `go test ./scripts/api-lint/...`; `go test -tags integration -run 'Tripwire' ./tests/integration/...` (existing hand-maintained tripwire regression suite, unaffected — proves no regression on the pre-existing arms).
8. Write ADR (`approval.oversee` + visibility model); commit
   `feat(iam,approval): F3 approval.review+approval.oversee caps, explicit tier-1, prefix fallback deleted (P1/P5)`.

## Notes

- Tripwire-arm regeneration command (for when F4 actually needs a new arm for `CapApprovalReview`):
  `go run ./cmd/gen-tripwire` — renders `internal/platform/tripwire.RenderMigration()` from the
  `TripwireArms` Go slice to the canonical migration path (check `defaultRelPath` in
  `cmd/gen-tripwire/main.go` before running — it advances each milestone, currently past 0283).
  Not invoked in F3 since no arm changes.
- `submit`/`publish`/`schedule-publish`/`supersede`/`obsolete`/route-admin tier-1 rows are
  untouched — verified already explicit and correct, no mismatch found (Interview record item 1).
