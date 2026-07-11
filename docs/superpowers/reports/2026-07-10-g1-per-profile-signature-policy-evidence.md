# G1 — Per-profile signature policy — evidence (COMPLETE)

**ROADMAP unit:** 2.1 (G1). **Status:** implemented, reviewed, L0/L1/L2 green. Committed on branch `claude/vibrant-lovelace-834af5`. **NOT pushed.**

- Ratified model: `docs/superpowers/specs/2026-07-10-review-approval-workflow-model.md` (R1 + G1).
- Implementation design (LOCKED): `docs/superpowers/specs/2026-07-10-g1-per-profile-signature-policy-design.md`.
- ADR: `wiki/decisions/0081-per-profile-signature-policy.md`.
- Scope guard: G2 (quorum rounds) and G3 (SoD) are OUT OF SCOPE — one gated feature only.

## What the feature is

Each `document_profile` carries a `governance_class ∈ {controlado, simples, livre}` that fixes its active approval route's signature policy:

| class | route-shape rule | app sentinel | HTTP |
|---|---|---|---|
| `controlado` | active route MUST contain ≥1 `approval`-kind stage | `ErrApprovalStageRequired` / `ErrRouteViolatesProfilePolicy` | 422 |
| `simples` | review-only route permitted; no approval stage required | — (never rejects) | — |
| `livre` | no approval route permitted at all | `ErrRouteNotPermittedForProfile` | 422 |
| reclassify to a class the current active route violates | conflict | `ErrClassChangeRouteConflict` | 409 |

Two enforcement lines: the friendly first line is domain `Route.Validate(policy)` at the route-admin write sites; the authoritative last line is the migration-0295 `DEFERRABLE INITIALLY DEFERRED` constraint triggers, which fire in both directions (route/stage writes → direction A/A′; profile reclassification → direction B) against the shared `assert_route_shape()` helper.

## Ratified decision (submit-site enforcement) — deviation of record

The LOCKED design piece 3 said "enforce `Route.Validate(policy)` at submit too." Submit cannot satisfy all three ratified invariants at once — ADR 0073 (route/profile resolved in-tx), H-PRE-1 (no recording CapTaxonomyView read inside the lock-holding submit tx), and the approval→taxonomy module boundary.

**Operator-ratified resolution: option 1.** Submit does structural-only `Validate("")`; the per-profile invariant is held at the write boundaries — route-admin `Validate(policy)` resolved **off-tx** (H-PRE-1) plus the both-direction DB trigger. Redundancy proof: "every active route conforms to its profile's current class" is authoritatively held by (route-admin Validate + direction-A/A′ trigger) and (reclassify guard + direction-B trigger). Submit only ever binds an **active** route ⇒ the bound route already conforms; `livre` has no active route ⇒ submit fails at route resolution. So a submit-time policy read is provably redundant. Recorded as an explicit deviation from the locked "enforce at submit" wording, citing H-PRE-1 + ADR 0073 + ADR 0081. **Option 2 (in-tx taxonomy port) permanently rejected** — it adds cross-module read surface the invariant makes unnecessary. Submit carries a marker comment (`submit_service.go` step 6) pinning this shape.

## Changeset

**New:**
- `db/migrations/0295_profile_governance_class.sql` — column + CHECK + `assert_route_shape()` + `enforce_profile_route_policy()` + 3 deferrable constraint triggers (`approval_route_stages`, `approval_routes`, `document_profiles`). Expand-only, idempotent, behavior-preserving backfill (every existing profile → `controlado`).
- `internal/modules/taxonomy/domain/governance_class.go` (+ `_test.go`) — GovernanceClass type + `RoutePolicy` derivation.
- `internal/modules/documents/approval/application/ports.go` — narrow `ProfilePolicyReader` port (approval never sees GovernanceClass, only RoutePolicy).
- `internal/modules/documents/approval/infrastructure/profile_policy_reader.go` (+ `_test.go`) — adapter over taxonomy `GetByCode`, own short tx, CapTaxonomyView, off-tx per H-PRE-1.
- `internal/modules/documents/approval/application/profile_policy_wiring_test.go` — nil-reader ⇒ friendly check skipped, trigger authoritative.
- `internal/modules/taxonomy/infrastructure/policy_error_test.go` — P0001 prefix → sentinel mapping.
- `tests/integration/approval/governance_class_policy_test.go` — DB-level race/enforcement integration (see L1).
- `wiki/decisions/0081-per-profile-signature-policy.md`.

**Modified (non-vendor):** `api/openapi/v1/openapi.yaml` (+ regenerated `internal/modules/taxonomy/api/api.gen.go`), taxonomy delivery/application/domain/infrastructure (governance_class in create/update/reclassify + Reclassify route + governance event), approval domain `route.go`/`errors.go` + `route_admin_service.go`/`submit_service.go`/`services.go` (`WithProfilePolicyReader`), `apps/api/cmd/metaldocs-api/main.go` (wiring), `tests/integration/testdb/factory.go` + `fixtures.go` (governance_class support), `wiki/modules/{approval,taxonomy}.md`, `wiki/decisions/index.md`.

Contract-first honored: routes changed only via `openapi.yaml` + oapi-codegen regen; `api.gen.go` never hand-edited.

## Review dispatch ledger (retroactive per-slice independent reviews)

| Slice | Verdict | Disposition |
|---|---|---|
| S1 domain GovernanceClass/RoutePolicy | CLEAN | — |
| S2 migration 0295 + triggers | **CRITICAL → fixed** | reviewer found the zero-stage/livre/activate-later route-row path the stage-only trigger could not see; fixed by adding the `approval_routes` direction-A′ constraint trigger. Re-reviewed CLEAN. |
| S3 persistence + Reclassify + P0001 mapping | CLEAN | MINOR deferred: suggested unit asserting setAuthzGUC + `authz.Require(CapTaxonomyManage)` fire before UPDATE in `SetGovernanceClassTx` — code correct, integration covers it. |
| S4 read port + adapter + wiring | CLEAN (inline) | — |
| S5 Route.Validate + route-admin sites | CLEAN | MINOR deferred: wording nit on `route.go:132-133` "fail-closing". |

## Verification ladder

**L0 (static) — all green:**
- `go build ./...` — clean.
- `api-lint -strict` — 0 violations (capability/tripwire coherence incl.).
- `scripts/check-module-boundaries.ps1` — OK (approval→taxonomy edge is domain/infra port only).
- `bash scripts/check-test-discipline.sh` — clean (135 integration test files checked).
- contract-sync (taxonomy) — OK; `go generate ./...` codegen — no drift (`api.gen.go` byte-identical).

**L1 (in-process) — all green:**
- `go test ./...` unit suite — green.
- Integration cascade cleared: new migration-0295 route-shape trigger initially rejected `NewApprovalRoute`'s bare route. Fix: flipped fixture default to `simples` (`factory.go` `NewTaxonomy` + `fixtures.go` `SeedGovernedTaxonomy`) + made the 6 governance-policy test spots opt into `controlado` explicitly via new `WithGovernanceClass` Opt. `simples` imposes no route-shape constraint ⇒ never rejects a fixture route.
- `tests/integration/approval/governance_class_policy_test.go` — green vs docker pg (proves DB-level enforcement + reclassify guard in-process).
- Residual approval/scenarios/tenantdata failures proven **pre-existing** (reproduced identically at the pre-G1 baseline via path-scoped stash + migration/​test move-aside; baseline had one *more* failure). Out of the G1 boundary.

**L2 (live container stack) — all green:**
- api rebuilt from worktree (`docker compose build --no-cache api`), force-recreated; migration 0295 applied on boot (`applied_now:1, total:35`); DB artifacts verified live (governance_class column, both route triggers, both functions, backfill 0 non-controlado).
- **Live-DB trigger enforcement smoke** (against the running `metaldocs-postgres`, each proof in its own tx forced via `SET CONSTRAINTS ALL IMMEDIATE` then `ROLLBACK` — enforcement observed, zero persistence):
  - P1 controlado + zero-stage active route → `ERROR ErrRouteViolatesProfilePolicy: profile is controlado; route … must contain at least one approval-kind stage` (direction A′). ✓
  - P2 livre + active route → `ERROR ErrRouteViolatesProfilePolicy: profile is livre; no approval route is permitted` (direction A′). ✓
  - P3 controlado + route WITH an approval-kind stage → committed clean (positive control — enforcement does not over-reject). ✓
  - P4 reclassify `po` (controlado, has active route) → livre → `ERROR ErrClassChangeRouteConflict: profile is livre; no approval route is permitted (route …)` (direction B; maps to app 409). ✓
  - DB tripwire (ADR 0022 last line) independently confirmed: raw `document_profiles` write without `metaldocs.asserted_caps` → `ErrCapabilityNotAsserted: one of {taxonomy.manage} required`.
  - Post-smoke: zero pollution (only `po` remains, still `controlado`); api restored to non-E2E state; 0295 persists.

## Bounded defers
- S3 MINOR: unit asserting authz GUC + `authz.Require` ordering in `SetGovernanceClassTx` (integration covers; code correct).
- S5 MINOR: `route.go:132-133` "fail-closing" wording nit.

## Security / process compliance
- No `.env` secret read, printed, committed, or exposed (env supplied to `docker compose` via `--env-file`, never sourced/echoed).
- PowerShell/coded-compose used for stack ops; DB proofs via `docker exec … psql`.
- Commit excludes the 3 spurious vendor CRLF-drift files (`cpuid/v2/CONTRIBUTING.txt`, `yaml-jsonpath/{LICENSE,NOTICE}`). **NOT pushed.**
