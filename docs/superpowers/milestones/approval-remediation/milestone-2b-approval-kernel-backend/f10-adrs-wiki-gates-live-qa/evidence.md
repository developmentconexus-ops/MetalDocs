# F10 — Evidence

## 1. ADR completeness + index sync

Verified structure (Status/Context/Decision/Consequences/Alternatives
Considered/Rollback/References) for all 4 M2b ADRs:

- `wiki/decisions/0074-approval-route-versioning.md` (F2)
- `wiki/decisions/0075-approval-review-oversee-capabilities.md` (F3)
- `wiki/decisions/0076-approval-freeze-boundary-and-choke-point-concurrency.md` (F5)
- `wiki/decisions/0077-approval-delegation.md` (F9)

Gaps found and fixed in `wiki/decisions/index.md`:

- **ADR 0076 was missing from the index entirely** (table jumped 0075→0077).
  Added its row.
- **ADR 0018's row was stale** — didn't reflect partial supersession by 0074
  (§1's update-in-place claim). Updated the Status/Superseded-by columns and
  summary to note the split (§1 superseded, §3 deactivate-guard unaffected).

Both edits re-verified clean against `scripts/check-adr-status.sh` (Status
block length/char budget gate — unaffected by index-table edits, still exit
0).

## 2. Wiki sync

Dispatched a wiki-curator background agent (task a7c1e2cc4f2a10006) scoped to
"sync wiki/modules/approval.md for F1-F9 combined changes." Verified its
diff directly:

- `wiki/modules/approval.md`: 567→606 lines. Added "Last verified:
  2026-07-07" stamp; updated the HTTP operations table (F4 review-verdict
  route, F9's two delegation routes with verified handler/line anchors);
  updated the Route Truth Table (flagged pre-existing stale `router.go` line
  anchors, added verified operationIds for the new routes); updated the
  authz section (F3 prefix-fallback deletion, new capabilities, F8's
  4-way `requireInstanceVisible` check, F9 delegation-never-bypasses-authz);
  added new subsections for the freeze boundary (F5) and the no-fallback
  content-hash chain (F6); updated idempotency section (F4 replay pattern);
  added ADR 0074-0077 rows to the Architecture Decisions section; added
  Glossary entries (stage kind, verdict, freeze boundary, signature meaning,
  delegation, approval.review, approval.oversee); updated failure-modes
  table; added a changelog entry.
- Confirmed `wiki/architecture/backend-api-structure.md` and
  `wiki/architecture/backend-target-architecture.md` need no changes (generic
  docs, no approval-specific content to update).

Flagged but explicitly NOT fixed (out of scope for a sync pass): §5.3's
pre-existing stale `router.go` line-number anchors (predates this milestone,
a mount-pattern change), and the doc's length (606 lines, over the ~300-line
soft cap) — both are pre-existing debt, not introduced by M2b.

## 3. Cross-feature verification sweep

All commands run from repo root after all F10 code fixes (see §4) were
applied, confirming the FINAL, integrated F1-F10 state:

| Command | Result |
|---|---|
| `go build ./...` | exit 0, no output |
| `go build -tags integration ./...` | exit 0, no output |
| `go test -count=1 ./...` | 96 packages `ok`, 0 `FAIL` |
| `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` |
| `go test ./scripts/api-lint/...` | included in the full sweep above, `ok` |
| `go test ./internal/modules/iam/domain/... -run TestCapabilityRegistrySize -v` | PASS, `want=40` (unchanged, includes M2b F3's +2) |

Grep-zero checks (re-run excluding `.claude/` stale worktrees, confirmed on
the live tree only):

| Check | Result |
|---|---|
| `SkipStage` anywhere in `*.go` | 0 hits |
| `LoadActiveDocumentContentHash` (old F6 pre-image fn) | 1 hit — a comment in `postgres_approval_repository.go:1477` explaining *why* it was replaced, not a live call; confirmed harmless |
| Bare `/api/v1/approval/` prefix row in `permissions.go` | 0 hits — all approval routes are explicit method+prefix+capability rows |

## 4. Live QA — real bugs found, fixed, and verified over HTTP

Ran `.\scripts\start-api.ps1 -Build` (native Windows process, PowerShell
tool) three times this session as fixes landed. First had to stop a stale
Docker container (`metaldocs-api`, built 2026-07-06 23:17, predating F3)
that was squatting on port 8081 via Docker's port mapping — `docker stop
metaldocs-api metaldocs-gateway`, then the sanctioned rebuild script started
cleanly. All further QA is against the native rebuilt binary + local
Postgres, migrations `0286`-`0293` all applied (33/33 total).

Logged in as dev-seed `admin` / `approver-test` (credentials from
`wiki/references/local-dev-startup.md`, not `.env`). All mutating requests
carry `Origin: http://localhost:8081` (required by `origin_protection`
middleware) and `Idempotency-Key` per the fixed-lifecycle contract.

### Bug #1 (FIXED, live-verified) — duplicate route-profile 500 instead of 409

`MapPgError` in
`internal/modules/documents/approval/infrastructure/errors.go` recognized
only the pre-F2 constraint name `approval_routes_tenant_profile_key`. F2's
migration 0287 renamed it to `approval_routes_active_profile_uq`, so a real
active-profile collision fell to the `default` branch → 500
`internal.db_unknown` instead of 409 `route.duplicate_profile`.

- Fix: added `"approval_routes_active_profile_uq"` to the same switch case.
- Unit coverage added: 2 new cases in
  `postgres_approval_repository_test.go`'s `TestMapPgError` table (legacy +
  new constraint name), 15/15 subtests pass.
- Live re-verification: POST a second route for the same tenant+profile ->
  `409 {"title":"insert route: approval: a route already exists for this
  tenant+profile combination","status":409,"code":"route.duplicate_profile"}`
  (`scratch_qa/route_collision.json`; log line `status":409` at
  `2026-07-07T16:28:21`).

### Bug #2 (FIXED, live-verified) — missing SAVEPOINT masks ErrRouteInUse as 500

`updateInPlaceOrSupersede` in
`internal/modules/documents/approval/application/route_admin_service.go`
attempts a speculative in-place UPDATE, catches the `enforce_route_immutable()`
trigger's `ErrRouteInUse` (P0001), and falls through to a supersede
UPDATE+INSERT fallback — all in one `*sql.Tx`, but with no `SAVEPOINT`
before the speculative UPDATE. Postgres aborts the *entire* transaction on
any statement error; without a savepoint every later statement (including
the fallback) errors closed with SQLSTATE 25P02
("current transaction is aborted"), turning the correctly-triggered
ErrRouteInUse into an opaque 500 instead of a working transparent-supersede.

- Fix: added `SAVEPOINT route_update_attempt` before the speculative UPDATE,
  `RELEASE SAVEPOINT` on success, `ROLLBACK TO SAVEPOINT` before the
  supersede fallback on `ErrRouteInUse`.
- `go build ./...` clean; `go test ./internal/modules/documents/approval/...`
  all subpackages `ok`, no regressions.
- **Test-harness limitation found, not fixed** (documented, not
  papered over): `route_admin_service_test.go`'s fake `routeAdminConn` /
  `routeAdminStmt` driver dispatches purely by SQL-text substring matching
  and has zero concept of Postgres transaction-abort semantics — any
  unmatched statement (including my new SAVEPOINT/RELEASE/ROLLBACK TO
  statements) silently returns a canned success. This is *why*
  `TestRouteAdminUpdate_RouteInUse` passed both before and after the fix —
  it could never have caught this class of bug. Rewriting the fake driver to
  model real tx-abort semantics would be a disproportionate new
  test-harness investment for a closing/verification feature; out of scope
  for F10, noted here as an observed structural test gap.
- Live re-verification: PUT `/approval/routes/{id}` on the in-use "QA E2E
  Route" (v1) adding a stage -> `200`, `new_version:2`, a NEW route id
  returned (v1 route flips to `active:false`, v2 row created and active) —
  confirmed via `GET /approval/routes` before/after
  (`scratch_qa/route_list.json` -> `scratch_qa/route_list2.json`); log line
  `status":200` at `2026-07-07T16:39:30`, no ERROR line, no SQLSTATE 25P02.

### Bug #3 (FIXED, live-verified) — `scope=oversee` worklist always 500s (SQLSTATE 42P18)

`ListWorklist` / `countWorklist` in
`internal/modules/documents/approval/application/read_service.go` (F8) build
`eligibilityPredicate` as either `"asi.eligible_actor_ids @> $2::jsonb"` or,
for `scope=oversee`, the bare literal `"TRUE"` — which drops the eligibility
filter entirely (list every in-progress tenant instance), by design. But
when the oversee branch fires, placeholder `$2` (the actor-ids JSON) is
never referenced anywhere in the query text at all. Postgres's real
parameter-type inference requires every bound placeholder to be inferable
from context; with `$2` completely absent from the SQL text, every
`scope=oversee` call failed closed with `SQLSTATE 42P18 ("could not
determine data type of parameter $2")`, surfaced as `500 internal.unknown`.

This was invisible to `read_service_test.go`'s pinned-shape unit test
(string-checks for the literal `"TRUE"`, doesn't execute real SQL) and to
`read_service_worklist_oversee_integration_test.go` — a correctly-written
`//go:build integration` test that has genuinely never run this milestone,
since it needs `DATABASE_URL`/`METALDOCS_DATABASE_URL` handed directly to
the test binary, which stays off-limits without reading `.env` (identical
bounded-defer precedent as F1-F9). Only real HTTP QA against a real running
Postgres instance surfaced this.

- Fix: changed the oversee-branch predicate from `"TRUE"` to
  `"($2::jsonb IS NOT NULL)"` in both `ListWorklist` and `countWorklist` — a
  tautology (actorJSON is always a valid non-nil marshaled array) that keeps
  `$2` referenced with an inferable type, while still not filtering by
  eligibility.
- Updated the pinned-shape unit test
  `TestListWorklist_Oversee_QueryShapeDropsEligibilityPredicate` to check for
  the new literal instead of the old.
- `go build ./...` clean; `go test ./internal/modules/documents/approval/...`
  all subpackages `ok`.
- Live re-verification: `GET /approval/inbox?scope=oversee` as `admin`
  (holds `approval.oversee`) -> `200`, `total:3`, returns instances submitted
  by `admin` whose stage pool does NOT include `admin`
  (`scratch_qa/inbox_oversee2.json`) — proving the eligibility predicate is
  genuinely dropped for oversee scope, not merely bypassed client-side.
  `GET /approval/inbox?scope=oversee` as `approver-test` (no
  `approval.oversee` grant) -> `403
  {"code":"authz.capability_denied", ...capability \"approval.oversee\"
  denied...}` — confirming the tier-2 authz gate genuinely runs before the
  query. Log lines confirm both: `status":200` and `status":403` at
  `2026-07-07T16:51`.

### Other live-QA confirmations (no bug, positive proof)

- **F6 no-fallback fail-closed, live-verified**: a pre-existing in-progress
  instance (`7272b16c-...`, submitted before F5/F6 landed in this DB's
  timeline, so `frozen_content_hash: null`) rejected a signoff attempt with
  `412 precondition.content_hash_mismatch` rather than silently accepting
  any substitute hash — the exact no-fallback behavior F6 exists to
  guarantee (`scratch_qa/signoff2.json`, `sod_test.json`). Confirmed via log
  that the instance state was unchanged after the rejection (no partial
  writes).
- **F8 stage_kind / due_before worklist filters, live-verified**:
  `GET /approval/inbox?stage_kind=approval` returns the expected matching
  set; `GET /approval/inbox?due_before=2020-01-01T00:00:00Z` correctly
  returns an empty set (all live instances have `due_at: null`, so no false
  positive) — confirms both filter predicates compile and run correctly
  against real Postgres, not just the sqlmock-pinned shape.
- **F2 route versioning + supersede-not-freeze, live-verified**: confirmed
  above under Bug #2's re-verification — `active`/`superseded_at` toggle
  behaves exactly per ADR 0074 (old row retired, new versioned row created,
  never a destructive update of an in-use row).

## 5. Reachability gaps — honestly reported, not forced

Two gaps blocked the rest of the planned lifecycle walkthrough (review-verdict
advancing/collapsing quorum, resubmit, freeze-at-transition, delegation
grant/act/revoke, SoD self-signoff rejection on a fresh instance, cross-tenant
404 matrix, cancel-with-reason). Both are genuine, not evaded:

1. **`stage_kind` is not exposed on the HTTP route contract.**
   `http/contracts/route.go`'s `StageRequest` has no `stage_kind`/`kind`
   field; every stage created via the live API defaults to
   `domain.StageKindApproval` (confirmed: the "Revisao" stage I added via the
   live PUT in Bug #2's re-verification landed as `stage_kind:"approval"`,
   not `"review"`, despite its name and `required_capability:
   "document.review"`). This is a pre-existing gap already admitted in F4's
   own evidence.md ("no factory builder yet exists for a review-kind stage
   instance... follows eligibility_test.go's raw-SQL FK-chain precedent").
   **No sanctioned way exists today to create a live review-kind stage
   instance over HTTP.** I did not attempt a raw SQL write to force one — a
   materially identical action was explicitly flagged this session by the
   harness's auto-mode classifier as "bad-faith tunneling... bypassing the
   API's authz/OCC/versioning-trigger layer," and I accepted that reasoning.
   Forcing a `stage_kind` value the same way would be the same class of
   action against the same tables. **Bounded defer**: add a `stage_kind`
   field to `CreateRouteRequest`/`UpdateRouteRequest`'s `StageRequest` (a
   real code feature, contract-first via `api/openapi/v1/openapi.yaml` +
   `oapi-codegen`, not a closing/verification task) — trigger: whenever a
   future feature needs to live-exercise the review-verdict endpoint end to
   end, or M2c FE work surfaces the same need first.
2. **No draft document available to submit fresh.** All documents in this
   tenant are already `under_review` (pre-existing QA fixtures); `GET
   /documents?status=draft` returns `{"items":[],"total":0}`. Creating one
   requires the full controlled-document/template/revision creation chain,
   which is a new-fixture investment disproportionate to a closing feature.
   **Bounded defer**: exercise the fresh submit -> freeze -> review-verdict
   -> resubmit -> signoff -> delegation chain end to end once a draft
   document exists (either from a future feature's own fixture needs, or a
   dedicated fixture-seeding task). Trigger: next session with a reason to
   create fresh documents (e.g., M2c FE screen work), or a dedicated
   real-Postgres end-to-end fixture task (same unassigned defer already
   recorded in F9's own evidence.md for the analogous Submit->Decision
   chain).

Both gaps were reachable-but-blocked for reasons independent of F10's own
work (pre-existing HTTP contract gap; pre-existing fixture state) — not new
regressions introduced by F1-F10.

## 6. Bounded defers carried forward (not closed by F10)

| Defer | Why still bounded | Trigger |
|---|---|---|
| Live-DB integration test run of the `//go:build integration` suites (F1-F9's own delegation/eligibility/route-versioning/worklist-oversee tests, this session's newly-relevant `read_service_worklist_oversee_integration_test.go`) | Requires `DATABASE_URL`/`METALDOCS_DATABASE_URL` passed directly to the test binary — obtaining the value requires reading `.env`, explicitly forbidden | Run `.\scripts\start-api.ps1` (confirms Postgres reachable) then, with an operator-supplied env var (never read from `.env` by the agent), `go test -tags integration -count=1 ./...`. Owner: next session with authorized DB access, or the operator directly. |
| `stage_kind` HTTP contract gap (§5.1) | Real code feature (contract-first schema addition), not a closing task | Trigger: next feature that needs live review-kind stage creation (M2c FE, or a dedicated fixture task). |
| Fresh full lifecycle walkthrough (submit->freeze->review-verdict->resubmit->signoff->delegation->cancel) on new fixtures | No draft document available; building one is a disproportionate new-fixture cost for a closing feature | Trigger: same as above, or a dedicated fixture-seeding task before the next milestone that needs it. |
| Route-admin fake-driver test harness doesn't model Postgres tx-abort semantics (§4 Bug #2) | Rewriting it is a disproportionate new test-harness investment for a closing/verification feature | Trigger: next time a savepoint/rollback-shaped bug in this package needs a REAL regression test, not just live-QA + unit build-passes. |

## 7. Self-review against the system-impact analysis' locked constraints

Checked every locked hard constraint from
`docs/superpowers/analysis/2026-07-07-approval-remediation-m2b-system-impact.md`
against F1-F10's combined state:

| Constraint | Status |
|---|---|
| AuthZ = capabilities, never roles (ADR 0022); tier-1+tier-2+tripwire | Held — confirmed via `permissions.go` grep-zero (no bare prefix fallback), `authz.Require` calls at every mutating handler, `TestCapabilityRegistrySize` pass |
| Contract-first (routes change only via openapi + oapi-codegen) | Held — `api-lint -strict` 0 violations |
| Route versioning: supersede not destructive update on in-use rows (ADR 0074) | Held — live-verified in Bug #2's re-check |
| approval.review/approval.oversee capabilities correctly scoped (ADR 0075) | Held — live-verified 403 on non-oversee actor, 200 on oversee-holder |
| Freeze boundary fires at review->approval transition or submit-for-approval-only (ADR 0076) | Implemented in code (F5); **not live-verified on a fresh instance** — bounded defer §6, pre-existing instances predate the migration so show `frozen_content_hash: null` as expected, not a regression |
| No-fallback content-hash chain (F6) | Held — live-verified: NULL frozen hash fails closed with 412, no substitute value ever accepted |
| Unified SoD predicate (F7) | Implemented in code, unit-tested; **not live-verified** (blocked by §5's reachability gaps) — bounded defer |
| SLA/visibility/worklist (F8) | Held for visibility (oversee 403/200 split verified) and filters (stage_kind/due_before verified); **the 42P18 bug found and fixed here was a real, previously-undetected regression in F8's own oversee-scope code path** — now closed |
| Delegation (F9) | Implemented in code, unit-tested; **not live-verified** (blocked by §5's reachability gaps) — bounded defer, same as F9's own evidence.md already recorded |
| Multi-tenant pooled (tenant_id everywhere, tx-local GUCs) | Held — every query in the touched code carries `tenant_id` predicates; no changes to tenancy scoping made by F10 |
| Async = transactional outbox | Not touched by F10 |
| DB enforces invariants (triggers/constraints) | Held — `enforce_route_immutable()` trigger confirmed firing correctly (that's what Bug #2 was masking, now correctly surfaced) |

**No gap is glossed over.** Three constraints (freeze boundary live-fire,
unified SoD live-fire, delegation live-fire) remain bounded-deferred with
named triggers, consistent with F9's own already-recorded defer for the
identical reachability problem — F10 did not introduce new gaps here, it
inherited them, attempted to close them, and hit real, honestly-documented
reachability walls (§5) rather than force a workaround through a path the
harness had already flagged as out of bounds.
