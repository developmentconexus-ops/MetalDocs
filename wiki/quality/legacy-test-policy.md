# Legacy Test Policy — Repair or Delete

> **Last verified:** 2026-07-06 — authored by M9 F9.3 (governance-hygiene).
> **Status:** This page **supersedes the informal memory rule** ("many tests are one-off task
> scaffolding; delete when they break, repair only contract/invariant guards") and is the
> canonical triage procedure for any legacy test that breaks, flakes, or blocks a change.
> **See also:** [test-discipline.md](test-discipline.md) (R1–R4 harness rules),
> [integration-test-harness.md](integration-test-harness.md) (how to write on the framework),
> [ADR 0034](../decisions/0034-integration-test-fixture-framework.md) (the canonical framework).

MetalDocs' test tree contains two very different populations that look identical from the
outside: **guard tests** that pin an architectural or regulatory invariant, and **scaffolding
tests** that were written to drive one task's fix and were never promoted to a guard. Treating
them the same wastes effort in both directions — scaffolding gets lovingly repaired forever,
and guards get deleted because "the test was old anyway". This page makes the classification
mechanical.

---

## The taxonomy

A broken/blocking legacy test is exactly one of two classes:

| Class | Trigger (any ONE suffices) | Disposition |
|---|---|---|
| **Repair-class** | Guards a **REQ ID** — cites or demonstrably covers a `REQ-*` line in [backend-target-architecture.md](../architecture/backend-target-architecture.md) (check the F9.2 evidence map `wiki/architecture/req-trace-map.yaml` and generated report `wiki/architecture/req-traceability.md`) | **Repair on the canonical framework.** Never delete. |
| | Guards a **tripwire arm** — asserts capability-tripwire firing/ordering, arm coverage, or arm↔registry parity (e.g. `tripwire_caps_test.go`) | |
| | Guards a **wire-contract shape** — an OpenAPI-generated DTO/route shape, RFC 9457 problem shape, or generated↔handwritten parity | |
| | Guards a **DB invariant** — a trigger, constraint, RLS policy, GUC requirement, or transition function enforced in the database | |
| **Delete-class** | **One-off task scaffolding** — drove a single task's fix and pins nothing above: assert-nothing walks, skip-only placeholders, commented-out "promise" files, error-propagation tours, duplicates of coverage a sibling already holds, superseded bespoke harnesses | **Delete**, with a one-line commit rationale. |

Classification is **per test function, not per file**. One file can contain both classes —
extract the repair-class remnant before deleting the shell (see worked example 1).

## The decision procedure

Run this when a legacy test breaks, flakes, or blocks your change:

1. **Does it guard a REQ ID, a tripwire arm, a wire-contract shape, or a DB invariant?**
   Check the test's assertions (not its name or age). Consult the REQ evidence map for citations.
   - **Yes → repair-class.** Go to step 2.
   - **No → delete-class.** Go to step 3.
2. **Repair on the canonical framework** — this is the test-framework **hard gate**:
   any rewritten or new DB-integration test MUST use the `testdb` factory
   (template-DB-per-test, [ADR 0034](../decisions/0034-integration-test-fixture-framework.md));
   unit tests use the standard per-test-instance patterns (e.g. per-test `sqlmock`). Repairing
   *in place* on a bypassed/bespoke harness is a defect. R1–R4
   ([test-discipline.md](test-discipline.md)) apply; the CI guard enforces them.
   Repair the **observation mechanism** if that is what broke, never weaken the **invariant**
   (see worked example 2).
3. **Delete with a one-line commit rationale** naming the class:
   ```
   test: delete one-off scaffolding <name> (legacy-test-policy delete-class)
   ```
   Never delete to make a suite green faster — deletion is only for tests that pin nothing.
   If any single assertion in the file is repair-class, extract it first.

**Fail closed:** if you cannot decide, treat it as repair-class and escalate. A wrongly-kept
scaffold costs minutes; a wrongly-deleted guard costs an invariant.

---

## Worked examples (this repo)

### 1. Delete-class with a repair-class remnant — `coverage_boost_test.go` (TST-10, commit `99f6f8cc`)

`internal/modules/documents/approval/application/coverage_boost_test.go` was 3,521 lines of
assert-nothing error-propagation walks, bespoke fake-driver rigs, getter smoke tests, and
duplicates of domain coverage — textbook one-off scaffolding written to inflate coverage, not
to pin behavior. The same sweep found 4 skip-only placeholder files and 2 fully commented-out
"promise" files.

Applying this policy retroactively: the file as a whole was delete-class, **but 18 of its tests
were repair-class in disguise** — they pinned coverage no sibling held (SQLEmitter governance
INSERT, float ban at all 4 service boundaries, `RecordSignoff` state legality, signoff
idempotency, multi-stage progression, OCC `ErrStaleRevision`, `ErrNoRows` mapping). The
disposition was: extract the remnant to
`internal/modules/documents/approval/application/service_invariants_test.go`, then delete the
shell. This is why classification is **per test, not per file**.

### 2. Repair-class where the *mechanism* was broken, not the invariant — ratelimit sweeper (TST-04, commit `695bd8e0`)

`internal/platform/ratelimit/eviction_test.go` guarded a real lifecycle contract: the eviction
sweeper goroutine must exit when its context is cancelled (a leaked goroutine per middleware
instance is a production defect). The test flaked constantly in full-suite runs — but the
sweeper was correct. The flake was the test's **observation mechanism**:
`runtime.NumGoroutine()` is process-global, so parallel tests' goroutine churn polluted both
the startup poll and the leak-check window.

Because the invariant is real, delete was not an option. The repair replaced the observation
mechanism with the deterministic contract the middleware already exposed — `Wait()` joins the
sweeper's `WaitGroup`, `Done()`d only when the loop returns; pre-cancel liveness asserted as
Wait-not-returned, exit asserted as Wait-returned within 1s of cancel. The flake class is
structurally removed with no isolate-retry policy.

### 3. Repair-class, currently broken at HEAD — `tenant_id_rls_integration_test.go`

`internal/modules/templates/infrastructure/tenant_id_rls_integration_test.go`
(`TestTemplateVersion_TenantID_RLSParity`) guards a **DB invariant** and a **REQ ID**: migration
0256's own-column `tenant_id` + FORCE-RLS `tenant_isolation` policy on
`templates_template_version` (REQ-TEN-1 / F-DB5). As of 2026-07-06 it fails at seeding: the
test drives `repo.CreateVersionTx` inside `testdb.SeedWithCaps` without seeding the
`metaldocs.tenant_id` GUC, and the (later-hardened) trigger
`enforce_template_version_tenant_consistent` now rejects the write. The *production* path seeds
that GUC via the TxRunner chokepoint (M3); the test predates that hardening.

Disposition under this policy: **repair-class — never delete.** It cites a REQ ID and pins an
RLS invariant; the failure is in the test's seeding, not the invariant. The repair belongs on
the canonical framework (seed the tenant GUC alongside the caps within the same tx), preserving
both assertions (a) insert persists `tenant_id`, (b) RLS filters cross-tenant reads under a
NOBYPASSRLS role.

---

## Anti-pattern appendix — `runtime.NumGoroutine()` in parallel suites

Never assert goroutine lifecycle via `runtime.NumGoroutine()` (or any process-global counter)
in a suite that runs alongside parallel tests. The counter observes the whole process: other
tests' goroutine churn makes both "goroutine started" polls and "goroutine leaked" windows
non-deterministic. Symptom: a test that passes `-count=5` in isolation and flakes in full-suite
runs (the TST-04 case above).

Assert lifecycle through a **deterministic, owned signal** instead: a `WaitGroup` the SUT
`Done()`s, a closed channel, or a context the SUT signals on exit. If the SUT exposes no such
signal, add one — that is a repair of the SUT's observability, not a test hack.

This matters doubly now that `t.Parallel()` is expanding across integration packages
(F9.3): any process-global observation (goroutine counts, global metrics registries,
`os.Setenv`) is unsound in a parallel suite by construction. Note the Go-level hard
constraint: `t.Parallel()` **panics** when combined with `t.Setenv` in the same test — files
using `t.Setenv` stay serial.

---

## Relationship to the test-framework hard gate

This policy **cites and does not weaken** the standing hard gate: every new or rewritten test
uses the canonical framework for its class — the `testdb` factory
([ADR 0034](../decisions/0034-integration-test-fixture-framework.md)) for DB integration,
per-test-instance mocks for unit tests — and complies with R1–R4
([test-discipline.md](test-discipline.md)), enforced by `scripts/check-test-discipline.sh` in
CI. Drive-by repair of pre-existing violations is allowed only on structural touch (the
allowlist can only shrink). Repair-class work that lands **off** the framework fails review.
