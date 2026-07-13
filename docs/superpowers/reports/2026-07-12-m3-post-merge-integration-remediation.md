# M3 post-merge integration remediation — shared-DB RED beyond baseline

**Date:** 2026-07-12
**Trigger:** After M3 kernel extraction merged to main (`8685b85f`) and the api was
rebuilt (migrations 0296–0300 applied), the full integration ladder showed **3 packages RED
beyond the 9-accepted baseline**. Hub demanded either (i) fix commits on a fresh branch off
current main restoring the suite to exactly the 9 accepted RED, or (ii) hard mechanistic
evidence the failures pre-date the M3 merge on this stack.
**Fix branch:** `fix/iam-integration-shared-db-hardening`, forked from **current main `f2680303`**
(contains merge `8685b85f`). Commits `052db933` (iam suite) + `42d3673a` (iam-infra + sequence).
**Verdict:** Failures are **DB-state (shared-DB pollution + dev-seed drift), not M3 code** —
proven mechanically below. Fixture-hardening (fixture-only, test-discipline R1–R4, no product
code, no reviewer required) restores the suite to exactly the 9-accepted baseline.

---

## 1. The beyond-baseline failing set (3 packages, 4 mechanisms)

Reproduced RED on the M3 stack (test files byte-identical to `8685b85f`):

| Package / test | Signature | Mechanism |
|----------------|-----------|-----------|
| `tests/integration/iam` — `TestRoleProvider_TenantIsolation` | `iam_user_roles_pkey` (23505) | leftover deterministic-row PK collision |
| `tests/integration/iam` — `TestHasAnyRole_TenantIsolation` | `iam_user_roles_pkey` (23505) | same |
| `tests/integration/iam` — `TestMembershipAreaScope_AreaAdmin_WithinManagedArea` | `user_process_areas_…_fkey` (23503) | missing dev-seed area (FK parent) |
| `tests/integration/iam` — `…_OutsideManagedArea` | 23503 | same |
| `tests/integration/iam` — `…_SystemAdmin_BypassNotBlockedByMissingArea` | 23503 | same |
| `tests/integration/iam` — `TestMembershipDirectory_AreaAdminScopedInSQL` | 23503 | same |
| `iam/infrastructure/postgres` — `TestRoleProvider_UserActiveInTenant_Live` | `active=false` (assert fail) | missing dev-seed `admin` user |
| `controlleddocuments/domain` — `TestSequenceAllocatorNextAndIncrement_Concurrent` | `document_profiles_family_code_fkey` (23503) | missing dev-seed `quality` family (FK parent) |

Every error is a **plain PK/FK constraint violation** (23505/23503) or a **missing-seed read**
— NOT the tripwire's `P0001 ErrCapabilityNotAsserted`. The capability tripwire is not on any of
these paths.

## 2. Why M3 code cannot be the cause (mechanistic proof, not classification prose)

M3 delta on the merged line = `72f8bd5c..1485bc2d` (merge-base of the two merge parents).

- **iam / taxonomy / controlleddocuments module source untouched:**
  `git diff --name-only 72f8bd5c..1485bc2d -- internal/modules/iam/** internal/modules/taxonomy/**`
  filtered to non-test = **∅**. RoleProvider, RoleAdminRepository, AreaMembershipService,
  UserAreaRepository, SequenceAllocator — none changed.
- **The failing test files untouched:**
  `git diff --stat 72f8bd5c..1485bc2d -- tests/integration/iam/ internal/modules/iam/infrastructure/postgres/ internal/modules/controlleddocuments/domain/`
  = **empty**. Same test bytes before and after the merge.
- **No M3 migration writes `iam_user_roles` / `user_process_areas` / `document_process_areas` /
  `document_families` / `document_profiles`.** 0296–0300 write only `approval_routes` /
  `approval_instances` / `approval_signoffs` (+ the tripwire function rewrite).
- **The tripwire rewrite is disjoint from the tables these tests touch.** 0299/0300 use
  `CREATE OR REPLACE FUNCTION public.enforce_capability_asserted()`, so the diff re-emits every
  table arm — the exact suspicion. Diffing the function body of the predecessor
  `0283_tripwire_delete_return_old.sql` against the final
  `0300_tripwire_signoff_parent_discriminator.sql` shows the **only** changes are: (a) one new
  `v_parent_subject_kind` var decl, (b) the `approval_instances` arm gains a `CASE NEW.subject_kind`,
  (c) the `approval_signoffs` arm gains a parent-lookup `CASE`. The `iam_user_roles`,
  `user_process_areas`, `document_process_areas`, and taxonomy arms are **byte-identical** across
  0283→0300.

Because no relevant *code* differs across the merge, the failure is **not git-bisectable to a
commit** — it is driven purely by shared-DB state, the same fact from the other side.

## 3. Root cause (DB state, all four mechanisms)

All three packages are **non-isolated**: `testdb.DSN(t)` returns the raw `DATABASE_URL` — no
per-test `CREATE DATABASE` clone (unlike the testdb-factory suites). They use deterministic /
dev-seed identities and rely on manual `t.Cleanup` DELETEs and on dev-seed rows existing.

- **Class 1 (iam 23505):** `insertUserRoleForTenant` upserts with `ON CONFLICT (tenant_id, user_id)`,
  which does **not** cover the PK `(user_id, role_code)`. A prior run killed mid-test leaves an
  `(user_id, role_code)` row (cleanup never ran). Next run uses a **fresh random tenant**
  (`NewTenant`), so the ON CONFLICT arm misses the leftover and the INSERT collides on the PK.
- **Class 2 (iam 23503):** `user_process_areas` FKs `document_process_areas(tenant_id, code)`. The
  tests assume dev-seeded areas `qualidade`/`rh` for `devTenant`. The rebuilt shared DB never
  (re)applied that dev-seed → FK parent absent → 23503.
- **Class 3 (iam-infra `_Live`):** `UserActiveInTenant` runs `SELECT EXISTS(… WHERE user_id=$1 AND
  tenant_id=$2 AND deactivated_at IS NULL)`. The probe reads a dev-seeded `admin`; rebuilt DB omits
  it → `active=false` → assert fail.
- **Class 4 (sequence allocator 23503):** `document_profiles.family_code` FKs
  `document_families(code)`. The setup assumes the dev-seeded `quality` family; rebuilt DB omits it
  → FK parent absent → 23503.

All four are properties of the shared DB + fixtures, independent of any M3 diff.

## 4. Fix (fixture-only; commits `052db933`, `42d3673a`)

- **Class 1:** pre-clean the deterministic user's `iam_user_roles` rows inside the seed tx
  (`tenant_isolation_test.go`) → seed self-healing against prior pollution.
- **Class 2:** new `seedProcessAreas` helper provisions the FK-parent `document_process_areas`
  rows in-fixture (idempotent, under the scheduler bypass GUC), called by the 4 area-dependent
  membership tests (`membership_area_scope_test.go`) → hermetic w.r.t. the FK parent.
- **Class 3:** the `_Live` probe now provisions an active `admin` in the system tenant if absent
  (seeded with `user.manage` asserted, per SEC-05/0259) and removes ONLY what it created; the
  pool close moved from `defer` to a first-registered `t.Cleanup` so the LIFO teardown DELETE runs
  before the pool closes.
- **Class 4:** the sequence setup tx now provisions the `quality` `document_families` row in-fixture
  (idempotent, under the already-asserted `taxonomy.manage`) before the `document_profiles` insert.

No product code touched → no independent reviewer required (hub rule). All three packages verified
GREEN individually:
`ok metaldocs/tests/integration/iam 9.325s`,
`ok metaldocs/internal/modules/iam/infrastructure/postgres 8.227s`,
`ok metaldocs/internal/modules/controlleddocuments/domain 7.891s`.

## 5. Flagged, not taken (global-maximum)

The real defect is that these suites run against a **shared, non-isolated DB** with deterministic
IDs / dev-seed assumptions and manual cleanup — the same fragility will resurface for any fixture
that assumes exact DB state. The global-max fix is **per-test DB isolation** (testdb-clone, matching
the isolated suites) for the iam / iam-infra / controlleddocuments-domain live suites. That is a
separate unit, out of M3 remediation scope; flagged here for the roadmap.

## 6. Full-suite delta (post-fix)

Full integration ladder re-run on the fix branch. Remaining RED == **exactly the 9-accepted
baseline**, no other RED. The 3 remediated packages are GREEN.

**Remaining RED (9 tests / 4 packages — all pre-existing accepted baseline):**

| Package | Tests | Baseline ID |
|---------|-------|-------------|
| `internal/modules/controlleddocuments/application` | `TestTenantIsolation_SequenceCounters_CrossTenant` | E-PROD |
| `internal/modules/jobs/approval_sla_surfacer` | `…_FullTick_IteratesAllTenants`, `…_Writer_TenantSeed_DoesNotSurfaceOtherTenant`, `…_Idempotent_SecondRunNoOp`, `…_AlertOnly_DoesNotMutateStatusOrDueAt` | E-PROD |
| `tests/integration/scenarios` | `TestGrantAreaMembershipFn`, `TestGrantAreaMembershipIdempotent`, `TestTriggerBypassBlocked` | E-PROD |
| `tests/integration/tenantdata` | `TestTenantDataPortCoverage` | E-PROD |

**Now GREEN (the 3 beyond-baseline packages, remediated by this branch):**

- `tests/integration/iam` — was 6 RED
- `internal/modules/iam/infrastructure/postgres` — was 1 RED (`_Live`)
- `internal/modules/controlleddocuments/domain` — was 1 RED (sequence allocator)

Suite restored to the accepted baseline. Remediation CLOSED.
