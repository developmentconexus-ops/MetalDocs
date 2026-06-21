# Feature F3.2 — Plan: CD consumes `v_active_user_areas` (C1 + C2)

> Input: `spec.md` (approved 2026-06-21). Engine: `superpowers:writing-plans` shape, inline.

## Plan

### Files touched
- **NEW** `internal/modules/controlleddocuments/infrastructure/membership_view_parity_integration_test.go`
  — the parity gate (`//go:build integration`, package `infrastructure`).
- **EDIT** `internal/modules/controlleddocuments/infrastructure/repository.go`
  — repoint the two restricted-visibility membership EXISTS legs (`List` ~:150, `CanRead` ~:492)
  from `user_process_areas` → `metaldocs.v_active_user_areas`; drop the `effective_to IS NULL` clause;
  refresh the membership-leg anchor comments.
- **EDIT** `tools/cilint/internal/analyzers/hgcrossmodule.go`
  — delete the `{controlleddocuments/infrastructure/repository.go, user_process_areas}` ledger entry.
- **EDIT** `tools/cilint/internal/analyzers/hgcrossmodule_test.go`
  — realign `TestHGCrossModule_Negative_PendingBaseline` from the now-drained CD site to a still-pending
  C4 `search/.../reader.go × user_process_areas` row.

### Test strategy (TDD, D6 — parity green BEFORE raw deleted)
1. Write `membership_view_parity_integration_test.go` with two tests:
   - `TestCanRead_ViewParityWithRaw` — seed company / restricted-area-grant / restricted-area-grant+**revoked**-membership / restricted-user-grant / owner / no-access CDs; for each (actor, cd) assert `repo.CanRead(...)` == an inline `rawCanRead(...)` helper that runs the **verbatim deleted** `user_process_areas … effective_to IS NULL` membership form.
   - `TestList_ViewParityWithRaw` — for each actor assert the visible-id set from `repo.List(...)` == an inline `rawListIDs(...)` helper using the verbatim raw membership leg.
2. Run **pre-repoint**: both green (raw repo == raw baseline — sanity; the revoked row proves the baseline itself excludes revoked).
3. Repoint `repository.go` both legs to the view.
4. Run **post-repoint**: both green (view repo == raw baseline — parity, no authz drift). The revoked-membership case is the discriminator.
5. Delete is the repoint itself (raw `user_process_areas` read replaced in place). After green, drain ledger + realign cilint test.

### Ordering
spec✓ → plan✓ → parity test (RED-capable) → pre-repoint green → repoint → post-repoint green →
`go build ./...` → drain ledger → realign cilint unit test → `go test ./tools/cilint/...` green →
`go run ./tools/cilint ./...` exit 0 → `git grep` clean → evidence.md → commit.

### Risks / mitigations
- **Inline-raw baseline must match the deleted SQL exactly** (else parity is vacuous). Mitigation: copy the membership leg verbatim from the pre-edit `repository.go` into the test helper, keep the surrounding owned-leg SQL identical to the repo's.
- **`List` keyset/pagination differences** could perturb the id set independent of membership. Mitigation: use a large `Limit`, compare as a set (sorted id slice), seed few CDs.
- **HS-3** if PG :5434 down → mark parity not-run, do NOT delete raw, stop.
