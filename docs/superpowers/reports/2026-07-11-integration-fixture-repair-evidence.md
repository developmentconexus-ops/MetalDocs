# Integration-suite fixture-layer repair — evidence (unit 2.2b)

**Branch:** `claude/suspicious-murdock-c3bf25` (forked post-G2, `bc8c351f` ancestor).
**Mission:** make `go test -tags=integration ./...` honest-green on main-lineage code via
RCA-first, class-based fixture repair — not per-test patching. Unblocks G3 (2.3) and full
acceptance of G2's integration tests.

**Binding triage rule applied to every failure:** (a) fixture/harness debt → repair at the
framework/fixture level; (b) test asserting retired behavior → delete per legacy-test-deletion;
(c) REAL product defect → STOP, escalate here, do NOT silently fix product code. Cross-tenant /
RLS / tripwire / SoD tests are invariant guards — never deleted or weakened.

---

## Baseline → result (package counts)

| Stage | setup-failed | FAIL pkgs | ok pkgs |
|---|---|---|---|
| Vendor-masked (fork checkout) | all | — | — |
| Honest baseline (post go-mod-vendor) | 7 | ~23 | 92 |
| After A/B/C/D commits | 0 | 16 | 99 |
| After R1/R2/R3/R4/R5 (final L1) | 0 | 4 | 111 |

**Final L1:** 111 ok packages · 4 RED packages (9 failing tests), each a documented product
defect (E-PROD-1..5) — zero fixture/harness debt remaining. Honest-green-modulo-escalations.

---

## Root-cause classes (confirmed, not assumed)

- **Class D (vendor gap)** — `.gitignore` `logs/` pattern was gitignoring vendored
  otlp-proto-logs + redis-maintnotifications `logs/` subtrees, so `vendor/` was incomplete on a
  fresh checkout and the ENTIRE integration build failed `[setup failed]` (masking the real
  baseline). Fix: `.gitignore` negations `!vendor/**/logs/` + `!vendor/**/logs/**` and
  `go mod vendor` to materialize 5 `.pb.go` files. (commit `948d03a0`) — NOTE: inventory said "did
  not reproduce on main"; it DID reproduce on the fresh worktree because the vendor files were
  untracked-local in main only.
- **Class C (compile rot)** — 7 `[setup failed]` packages: stale import paths + an import cycle
  (`iam/infrastructure/postgres` parity test → testdb → bootstrap → back). Fix: path renames +
  moved parity/bootstrap/river tests to external `_test` packages. (commit `abce2419`)
- **Class A (identity/caps unseeded)** — dominant class. Split into:
  - **A1 (seed-tx GUCs)** — testdb factory now seeds `metaldocs.actor_id`/`tenant_id` via
    `authz.SeedTxIdentity`/`SeedTxTenant` through new `seedTx`/`seedWithCapsIdentity`/`seedTenantOnly`
    fixture primitives. (commits `eb079861`, `06e742e4`)
  - **A2 (ctx-seeded identity for SUT-owned txns)** — SUTs run their own `TxRunner.Do`, which reads
    identity from ctx (`platformtenant.ActorFromContext`); fixtures were passing bare
    `context.Background()`. Added `testdb.AuthzCtx(tenantID, actorID)` + per-actor ctx threading.
    (commits `eb186bf2`, `1bf179b6`)
  - **A2-caps (capability grants)** — beyond seeding identity, approval fixtures listed actors in
    `eligible_actor_ids` but never granted them a capability-bearing role in the area, so tier-2
    `authz.Require` denied them once identity was correct. Added `user_process_areas` role grants
    (author→area_admin, reviewer→approver) + `document_process_areas` FK rows. SoD self-verdict
    guard preserved (separate same-actor check). (commit `1bf179b6`)
- **Class B (schema drift vs stale seeds)** — `approval_stage_instances.required_capability_snapshot`
  NOT NULL (23502) + `_eligibility_drift_snapshot_check`; `uq_iam_user_roles_user_tenant` one-role-
  per-user-per-tenant; template-version tenant-consistency 23514. Fix: seeds updated to supply the
  now-required columns / respect the tightened constraints. (commit `06e742e4` + residual sweep)

---

## Triage table (repaired / deleted / escalated)

| Item | Package | Bucket | Disposition |
|---|---|---|---|
| approval verdict/freeze/cancel/SoD | tests/integration/approval | A2-caps | REPAIRED (`1bf179b6`) |
| submit `_RealDB` empty-pool | documents/approval/application | A2-caps | REPAIRED (`1bf179b6`) |
| security `_Live` port parity | security/infrastructure/postgres | A | REPAIRED (R3) |
| taxonomy code immutability | taxonomy/application | A | REPAIRED (R3) |
| iam ProbeA direct-insert blocked | iam | E (guard) | REPAIRED (stale SQLSTATE 42501→P0001; guard preserved) |
| iam tenant-user/display `_Live` | iam/infrastructure/postgres | A | REPAIRED (R3) |
| iam user-tenant multi-role | iam/infrastructure/postgres | B | REPAIRED (R3; seed 2 distinct tenants) |
| template orphan sweeper | templates/jobs | A | REPAIRED (R3) |
| preview-code 403/200 | controlleddocuments/delivery/http | A | REPAIRED (R3) |
| cross-tenant profile NotFound | controlleddocuments/application | A | REPAIRED (R3) |
| watchdog P1 alert-only | jobs/stuck_instance_watchdog | C | REPAIRED (R3; `metaldocs.documents`→`public.controlled_documents`) |
| api-lint clean-spec | scripts/api-lint | A | REPAIRED (R3; seed-chokepoint allowlist entry for testdb fixture) |
| scenarios E4/E5/idempotency | tests/integration/scenarios | A/path | REPAIRED (R2 `57dedd24`; `metaldocs.*`→`public.*`, tenant+cap seed) |
| documents/application Class-B | documents/application | A/B | REPAIRED (R2; BeginTx+SeedTxTenant+SetCapsOnTx, content_hash, Supersede) |
| migrations 0152/0153/0169 | tests/integration/migrations | C/name | REPAIRED (R2; `'public'` literal schema, real FK bug fixed, retired-assert deletes) |
| SeedRouteConfig FK / scenarios lowercase / OCC per-worker caps | testdb + scenarios | A/B | REPAIRED (R4 `13a333d2`) |
| publish/freeze/sla_due/ProbeG residuals | approval + iam | A2-caps | REPAIRED (R5 `dc167b8e`) |
| governance_class_policy inline set_config | tests/integration/approval | R1-discipline | REPAIRED (R4 `6e76f9ae`; `SetCapsOnTx`) |
| membership_fn / trigger_bypass / seq-counters / tenantdata / SLASurfacer | scenarios + controlleddocuments + tenantdata + jobs | **PRODUCT** | ESCALATED E-PROD-1..5 (left RED) |

---

## ESCALATIONS — REAL product defects (NOT fixed; hub decision required)

### E-PROD-1 — `sla_overdue_reader.go:37` ambiguous `status` column
- **File:** `internal/modules/documents/approval/infrastructure/sla_overdue_reader.go:37`
- **Defect:** `slaOverdueCorePredicate = "status = 'active' AND ..."` uses an unqualified `status`
  in a query joining `public.approval_stage_instances asi` and `public.approval_instances ai` —
  both carry a `status` column → `column reference "status" is ambiguous (SQLSTATE 42702)`.
- **Blast radius:** `ListTenantsWithOverdueStages` (lines 91-106) and the writer both consume the
  shared predicate; all 4 `TestIntegration_SLASurfacer_*` fail. This is a real runtime bug in the
  SLA surfacer read path — not a fixture issue.
- **Fix (product, out of scope here):** qualify the column (`asi.status`) in the predicate.

### E-PROD-2 — `document_profiles` primary key not tenant-scoped
- **File:** `db/baseline/0001_current_schema.sql:1116-1132,2606-2610` (unchanged through
  `db/migrations/0295_profile_governance_class.sql`).
- **Defect:** `document_profiles` carries `tenant_id` (FORCE ROW LEVEL SECURITY) but its primary key
  is `PRIMARY KEY (code)` only — never `(tenant_id, code)`. Two tenants therefore cannot use the
  same profile code; a legitimate cross-tenant reuse hits
  `duplicate key ... document_profiles_pkey (23505)`.
- **Invariant violated:** CLAUDE.md multi-tenant-pooled ("every tenant table has `tenant_id`" +
  tenant-scoped uniqueness).
- **Surfaced by:** `TestTenantIsolation_SequenceCounters_CrossTenant`
  (`internal/modules/controlleddocuments/application/tenant_isolation_test.go:216`) — left FAILING
  on purpose; patching the test to avoid code reuse would hide the gap.
- **Fix (product, needs architecture decision):** migration to composite PK / tenant-scoped unique.

### E-PROD-3 — TenantDataPort missing for two tenant-scoped tables (CONFIRMED)
- **Files:** `internal/platform/tenantdata/registry/registry.go` +
  `internal/modules/documents/approval/infrastructure/tenant_data_port.go` (covers only
  `approval_instances`/`approval_routes`/`approval_signoffs`).
- **Defect:** `public.approval_delegations` (migration `0293_approval_delegations.sql`, has
  `tenant_id`) and `public.approval_review_verdicts` (migration `0288_review_verdicts_changes_requested.sql`,
  has `actor_tenant_id`) are tenant-scoped PII tables with **no registered `TenantDataPort`** → they
  are silently excluded from GDPR export + crypto-shred erasure.
- **Surfaced by:** `TestTenantDataPortCoverage` (`tests/integration/tenantdata/coverage_test.go`) —
  left FAILING. This is a real GDPR/tenant-lifecycle coverage gap, not test debt.
- **Fix (product):** register tenant-data ports for both tables.

### E-PROD-4 — `grant_area_membership` area_code CHECK contradiction
- **File:** `db/baseline/0001_current_schema.sql:204-206` (the `metaldocs.grant_area_membership`
  SECURITY DEFINER function).
- **Defect:** the function validates `_area_code` against `^[A-Z0-9_]+$` (UPPERCASE-only), but the
  row it inserts is FK-constrained (`user_process_areas_tenant_id_area_code_fkey`) to
  `metaldocs.document_process_areas(tenant_id, code)`, whose CHECK `area_code_format`
  (`0001_current_schema.sql:1060`) requires `^[a-z][a-z0-9_-]{1,63}$` (LOWERCASE-only). **No string
  satisfies both** → the function is unconditionally broken against the current schema.
- **Blast radius:** narrow — `internal/test/e2e_seed.go:574-584` only falls back to this function if a
  direct insert fails. But the function is dead code today.
- **Surfaced by:** `TestGrantAreaMembershipFn` + `TestGrantAreaMembershipIdempotent`
  (`tests/integration/scenarios/membership_fn_test.go`) — left FAILING
  (`invalid area_code ... SQLSTATE 22023`).
- **Fix (product):** reconcile the case rule — lowercase the function's regex to match
  `area_code_format`.

### E-PROD-5 — legal-transition trigger bypassable via `session_replication_role`
- **File:** `db/baseline/0001_current_schema.sql:3858` (`CREATE TRIGGER trg_documents_legal_transition
  ... BEFORE UPDATE`).
- **Defect:** the trigger is a plain trigger, not `ENABLE ALWAYS`, so
  `SET LOCAL session_replication_role = 'replica'` disables it entirely — and the dev/test connecting
  role has enough privilege to set that GUC (same privilege-escalation class as the memorialized
  "metaldocs_app superuser+BYPASSRLS → RLS inert in dev"). A raw `UPDATE documents SET
  status='published'` on a fresh draft then succeeds with no error. No defense-in-depth
  (`ENABLE ALWAYS`) layer protects this legal-lifecycle invariant.
- **Surfaced by:** `TestTriggerBypassBlocked` (`tests/integration/scenarios/trigger_bypass_test.go:18`)
  — left FAILING (guard NOT weakened). `TestIllegalTransitionBlocked` still PASSES (blocked by a
  different guard, `enforce_snapshot_on_submit`).
- **Fix (product):** `ALTER TABLE documents ENABLE ALWAYS TRIGGER trg_documents_legal_transition`
  (and audit other invariant triggers for the same exposure).

---

## Invariant guards preserved (never weakened)
- iam ProbeA direct-insert-blocked — kept as guard; only stale expected SQLSTATE updated.
- Approval SoD self-verdict block — still asserted after capability grants added.
- Cross-tenant isolation tests — repaired to seed tenant ctx correctly, assertions unchanged.

---

## Gates (L0 / L1)

### L0 — all GREEN
| Gate | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go build -tags=integration ./...` | exit 0 |
| `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | 0 violations |
| `scripts/check-module-boundaries.ps1` | `[module-boundaries] OK` |
| `scripts/check-test-discipline.sh` | clean (136 files) — after fixing the one pre-existing G1 R1 violation (`governance_class_policy_test.go:366`, commit `6e76f9ae`) |

### L1 — full `go test -tags=integration ./...` (canonical `scripts/test-integration.ps1`)
**Final post-R5 rerun: 111 ok packages · 4 RED packages · 9 failing tests.** Every RED test maps
1:1 to an escalated product defect — zero fixture/harness debt remains:

| RED package | Failing tests | Escalation |
|---|---|---|
| `jobs/approval_sla_surfacer` | 4× `TestIntegration_SLASurfacer_*` | E-PROD-1 (ambiguous `status`) |
| `controlleddocuments/application` | `TestTenantIsolation_SequenceCounters_CrossTenant` | E-PROD-2 (profile PK not tenant-scoped) |
| `tests/integration/tenantdata` | `TestTenantDataPortCoverage` | E-PROD-3 (missing tenant ports) |
| `tests/integration/scenarios` | `TestGrantAreaMembershipFn`, `TestGrantAreaMembershipIdempotent` | E-PROD-4 (area_code CHECK contradiction) |
| `tests/integration/scenarios` | `TestTriggerBypassBlocked` | E-PROD-5 (trigger not `ENABLE ALWAYS`) |

## Commit ledger
- `948d03a0` — .gitignore vendor-logs negations + go mod vendor (Class D)
- `abce2419` — Class C compile-rot (import paths + cycle break)
- `eb079861`, `06e742e4` — Class A1 + Class B fixture framework
- `eb186bf2` — Class A2 AuthzCtx + ctx threading
- `1bf179b6` — Class A2-caps approval capability grants
- `4775d6c5` — HARNESS §2 comms protocol adoption
- `57dedd24` — R2 scenarios/migrations/documents-application (`metaldocs.*`→`public.*`, schema literal, FK fix)
- `3aee8d7f` — R3 security/taxonomy/iam/templates-jobs/controlleddocuments/watchdog/api-lint
- `13a333d2` — R4 SeedRouteConfig FK · scenarios lowercase codes · OCC per-worker caps
- `6e76f9ae` — R4 governance_class_policy inline set_config → `testdb.SetCapsOnTx` (last R1-discipline violation)
- `dc167b8e` — R5 publish/freeze/sla_due/ProbeG residuals
- _evidence doc commit follows_

## Dispatch ledger
- R1 (approval capability grants) — done, committed `1bf179b6`.
- R2 (scenarios E4/E5/idempotency · documents/application Class-B · migrations E2/E6) — done, `57dedd24`.
- R3 (security · taxonomy · iam · templates/jobs · controlleddocuments · watchdog · api-lint) — done, `3aee8d7f`.
- R4 (SeedRouteConfig FK · scenarios lowercase · OCC caps · discipline fix) — done, `13a333d2` + `6e76f9ae`.
- R5 (publish/freeze/sla_due/ProbeG) — done, `dc167b8e`.
- Investigator (read-only classification of 10 hand-rolled-connection packages) — done.

## HS-1 items (operator)
- **5 escalated product defects (E-PROD-1..5)** require product decisions before fix — not touched here.
  E-PROD-2 (profile PK) needs an architecture/migration decision; E-PROD-5 (`ENABLE ALWAYS`) is a
  defense-in-depth hardening; E-PROD-1/3/4 are direct product fixes.
- Branch `claude/suspicious-murdock-c3bf25` NOT pushed (per standing rule).
- Co-author trailer on R-batch commits reads "Claude Fable 5"; harness convention is "Claude Opus 4.8"
  — cosmetic, commits already sealed.
