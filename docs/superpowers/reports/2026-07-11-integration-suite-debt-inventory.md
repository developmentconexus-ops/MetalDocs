# Integration-suite debt inventory — 2026-07-11

**Context.** The canonical runner `scripts/test-integration.ps1` (commit 04cee2f9) removed the
silent-skip false green (`testdb.Open` t.Skips when `DATABASE_URL` is unset or the DB is
unreachable). First honest full run of `go test -tags=integration ./...` against the live compose
postgres (post-merge main: units 1.2 + 1.3@e9461515 + G1) exposed pre-existing debt that skip had
hidden. Surfaced by unit 2.2 (G2) escalation; failures reproduce identically on main for tests G2
never touched.

**Totals:** 22 packages FAIL · 75 individual test failures · 7 packages fail at compile under the
`integration` tag (`[setup failed]`) · 92 packages ok.

## Compile-level (`[setup failed]`)

| Package | Known cause |
|---|---|
| `internal/modules/iam/infrastructure/postgres` | import cycle: `active_user_areas_view_parity_integration_test.go` → testdb → bootstrap → back into the package under test |
| `internal/platform/bootstrap` | TBD (classify) |
| `internal/platform/jobs/river` | TBD |
| `tests/docx_v2` | TBD |
| `tests/integration/controlleddocuments` | TBD |
| `tests/integration/documents` | TBD |
| `tests/integration/templates` | TBD |

## Runtime failures by package (75 tests)

- `internal/modules/controlleddocuments/application` — 6 (authz/caps, tenant isolation)
- `internal/modules/controlleddocuments/delivery/http` — 2 (preview-code 403/200)
- `internal/modules/documents/application` — snapshot columns, fill-in schema parity, SP2 dictionary pin (6+)
- `internal/modules/documents/approval/application` — freeze suite, review-verdict suite, submit `_RealDB` suite, cancel, SoD (20+)
- `internal/modules/documents/approval/infrastructure` — stage due_at snapshot recompute
- `internal/modules/iam` — grant-area-membership fn/idempotent
- `internal/modules/jobs/approval_sla_surfacer` — 4 (full tick, tenant seed, idempotent, alert-only)
- `internal/modules/jobs/stuck_instance_watchdog` — 2 (P1 alert-only)
- `internal/modules/security/infrastructure/postgres` — 2 (`_Live` port parity)
- `internal/modules/taxonomy/application` — code immutability
- `internal/modules/templates/infrastructure` — template-version state `_Live`
- `internal/modules/templates/jobs` — orphan sweeper
- `tests/integration/approval` — freeze/review-verdict/SoD-trigger (dominant class below)
- `tests/integration/migrations` — 0152/0153/0169 assertions
- `tests/integration/scenarios` — concurrency, idempotency (4), outbox (2)
- `tests/integration/tenantdata` — writer reads, trigger bypass, illegal transition, port coverage

## Observed root-cause classes (hypotheses to confirm, not conclusions)

1. **Identity/caps unseeded in fixtures** — `authz: metaldocs.actor_id GUC not set` and
   `P0001 ErrCapabilityNotAsserted ... asserted_caps is not set on documents` at
   `stamp process_area_code_snapshot`. Fixtures use `context.Background()` without identity
   seeding + per-actor capability grants. Dominant class (freeze, review-verdict, submit, cancel).
2. **Schema drift vs stale seeds** — `approval_stage_instances.required_capability_snapshot`
   NOT NULL (23502) + `_eligibility_drift_snapshot_check` violated by old seed rows
   (`sla_due_at` tests).
3. **Compile-level rot** — import cycle + 6 other `[setup failed]` packages; these tests have not
   compiled for an unknown period.
4. **Unverified** — chip-reported vendor gap (`otlp-logs`, `redis-maintnotifications` missing under
   `vendor/`); did NOT reproduce on main checkout (`go test -tags=integration -run xxx_none ./...`
   compiles except the import-cycle package). Verify per-worktree.

**Caution for the repair unit:** classify each failure as (a) fixture/harness debt, (b) test
asserting retired behavior (legacy-test-deletion rule may apply), or (c) REAL product defect.
Any (c) escalates to the hub immediately — do not silently "fix" the test.

Raw log: scratchpad `claude-integration-main-full.txt` (session-local); re-derivable via
`.\scripts\test-integration.ps1`.
