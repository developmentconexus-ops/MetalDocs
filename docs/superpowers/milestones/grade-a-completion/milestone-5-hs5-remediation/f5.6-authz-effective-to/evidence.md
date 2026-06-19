# Evidence — F5.6 authz-effective-to (re-audit Major #1 — REFUTED)

> **Status:** CLOSED 2026-06-19 · **Major #1 refuted with evidence; no predicate change** (Option A,
> operator-approved). Disposition: audit false-positive — `effective_to` is a soft-delete tombstone,
> not a future-dated validity-end. Documented in ADR 0037 + 6 code anchors.

## Disposition

The re-audit (2026-06-16 Major #1) claimed `authz.Require` "denies time-bounded active memberships"
and the active-now predicate should become `(effective_to IS NULL OR effective_to > now())`. The
investigation (spec.md) found the premise false: MetalDocs implements **soft-delete with a current
marker** (active ⟺ `effective_to IS NULL`), not bitemporal valid-time. Proof, three layers:

| Layer | Evidence | Source |
|-------|----------|--------|
| Schema | UNIQUE partial index `ux_user_process_areas_single_active … WHERE effective_to IS NULL` (+2 more partial indexes) defines "active". A future-dated row can't coexist with the current row under it. | `db/baseline/0001_current_schema.sql:3618,3667,3674` |
| Write path | `Grant` takes no end-date arg → inserts `effective_to = NULL`; `Revoke` sets `effective_to = time.Now().UTC()` (past) + `revoked_by`. No future `effective_to` is ever written. | `area_membership_service.go:78,130,245` |
| API | The DTO `effective_to` is read-only output (revoke timestamp of a closed row), never a grant input. | `api.gen.go:300`, `routes_memberships.go:62,75` |

So `effective_to > now()` matches zero rows. The proposed change would (1) grant no access, (2)
regress the authz hot path off its `WHERE effective_to IS NULL` partial indexes (OR-form not sargable
against a partial-NULL index), (3) contradict the unique index that *is* the definition of active —
introducing read/write split-truth. Rejected as a symptom-patch against an architecture contradiction
(HS-2; CLAUDE.md runtime-truth; ADR 0022 never-symptom-patch-authz). Industry framing: the design is
the canonical Postgres "one active row" pattern (partial-unique-on-active), same family as SCD-2
current-row / `discarded_at IS NULL` soft-delete — affirmed by operator as matching industry standard.

## Change (documentation + legibility only — no behavior, no SQL)

| File | Change |
|------|--------|
| `wiki/decisions/0037-membership-temporal-model.md` | **New ADR (Accepted)** — records Model A, refutes Major #1 with evidence, gates any future Model-B work behind a successor ADR. |
| `internal/modules/iam/authz/authz.go:114` | Go `//` comment above the `Require` active-now query: `effective_to IS NULL` is canonical (ADR 0037); do not change to interval form. |
| `internal/modules/controlleddocuments/infrastructure/repository.go` | Go comments above `List` (`:87`) and `CanRead` (`:470`) area-grant EXISTS subqueries. |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go` | Go comments above `MembershipDirectoryScope` (`:144`) and `ListByTenantInManagedAreas` (`:174`) — the latter notes the *intentional* dual predicate (outer as-of interval vs inner active-now `IS NULL`). |
| `internal/modules/search/infrastructure/v2documents/reader.go:29` | Go comment above `ListDocuments` visibility query. |
| `wiki/database/tables/user_process_areas.md` · `wiki/decisions/index.md` | Active-predicate rule documented; ADR 0037 indexed. |

**The 6 active-now SQL strings are byte-identical to HEAD.** Comments live in Go (`//`) above the
query literals, NOT inside the SQL — an earlier attempt to put them inside the strings broke the
`role_admin_repository_test.go` sqlmock exact-match tests; that is the regression-proof that no SQL
changed (in-string edits fail those tests; these pass).

## TDD record

**No behavior change → no new behavioral test.** The protection against a future blind predicate-flip
is the ADR + the 6 code anchors + the existing schema invariant (the partial unique index). A unit
test cannot meaningfully guard the regression direction: the proposed OR-form still grants live
memberships and denies revoked ones (the `OR` clause is dead), so a behavioral test would pass under
both forms. Honest label: this feature's deliverable is a refutation + legibility hardening, not a
code fix.

## Validation Gate results (real output)

1. **No SQL change** — sqlmock exact-match tests in `internal/modules/iam/infrastructure/postgres`
   pass (incl. `TestUpsertUserAndAssignRole_PassesTenantID`,
   `TestReplaceUserRoles_DeleteThenInsert_PersistsSingleRole`), which exercise the authz Require SQL
   verbatim → proves the active-now query strings are unchanged.
2. **ADR recorded** — `0037-membership-temporal-model.md` Accepted.
3. **Code anchors** — 6 sites carry the ADR-0037 pointer comment.
4. **Build** — `go build ./...` → `BUILD OK`.
5. **Tests** — `go test -count=1 ./internal/modules/iam/... ./internal/modules/controlleddocuments/...
   ./internal/modules/search/...` → all packages `ok`.

## Defers

None. The Model-B (time-bounded memberships) option is recorded in ADR 0037 D4 as a separate future
milestone/mission gated by a successor ADR — **not** a defer of F5.6 (it is out of scope by decision,
not deferred work).
