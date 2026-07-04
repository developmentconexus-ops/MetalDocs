# F4.5 plan

Root-cause, class-level fix. One shared reader; four call sites unified; two TDD pins.

## Tasks (TDD)

1. **Failing pins first** — add to `internal/modules/iam/authz/context_test.go`:
   `TestMustActorID_ReturnsErrWhenGUCNull` and `TestMustTenantID_ReturnsErrWhenGUCNull`
   (sqlmock `AddRow(nil)` → SQL NULL → assert `ErrActorContextMissing`/`ErrTenantContextMissing`).
   Confirm they FAIL against current code (driver error).
2. **Shared reader** — add `readSoftGUC(ctx, tx, query) (string, error)` to `context.go`: scan into
   `sql.NullString`; `!Valid ⇒ ""`; propagate scan errors. Doc the PostgreSQL NULL-vs-'' subtlety and
   the byte-identical-SQL / no-injection constraint.
3. **Rewire the readers** to the helper:
   - `context.go`: `MustActorID` / `MustTenantID` → `readSoftGUC` then `if v=="" → sentinel`.
   - `bypass_audit.go`: `softGUC` → `readSoftGUC`, keep fail-soft `err||v==""` → default.
   - `authz.go`: `loadAssertedCaps` → `readSoftGUC` (drop its local `sql.NullString`).
4. **Green pins** — `go build ./...`; `go test ./internal/modules/iam/authz/ -count=1`.
5. **Regression (real DB)** — `go test ./internal/modules/iam/authz/ -tags integration -count=1`.
6. **Downstream proof** — run F4.2 `TestPublishRace -tags integration` live green (its harness also
   gets the production-faithful identity seeding; see F4.2 evidence).
7. Independent review pass on the aggregate diff.

## Files touched

- `internal/modules/iam/authz/context.go` — new `readSoftGUC`; `MustActorID`/`MustTenantID` rewired.
- `internal/modules/iam/authz/bypass_audit.go` — `softGUC` rewired.
- `internal/modules/iam/authz/authz.go` — `loadAssertedCaps` rewired.
- `internal/modules/iam/authz/context_test.go` — two NULL-case pins.
- (F4.2 test-harness identity seeding is recorded under F4.2, not here.)

## Gate

Two new pins RUN + PASS · empty-string/happy/seed tests still green · unit + integration authz green
on real Postgres · F4.2 live green · accept path unchanged (stricter-or-equal) · no prod publish-path
change.
