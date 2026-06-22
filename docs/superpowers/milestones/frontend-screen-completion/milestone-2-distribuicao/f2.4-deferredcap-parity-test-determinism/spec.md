# Feature F2.4 — Spec (deferredcap-parity-test-determinism)

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Folder:** `f2.4-deferredcap-parity-test-determinism`
> **Status:** Approved (pre-code) — 2026-06-22. **Opened under HS-4** by the milestone-validator FAIL (`../qa/milestone-qa.md`, C7).
> **Class:** correction / test-determinism. **No product, contract, migration, or authz behavior change.**

> This is the minimum fix feature the M2 milestone-validator named. It exists to make one CI-gating test
> deterministic so the milestone's own "6 CI guards green / `go test ./...` green" bar holds on every run, not ~97%.

## Root cause (validator-found, controller-confirmed at source)

`scripts/api-lint/registry_rules_test.go:136-137` — `TestSeedRegistryParity_RegistryNotSeeded` does:
```go
caps := iamdomain.AllCapabilities()   // iterates a Go map → randomized order (model.go:173-179)
missing := string(caps[0])            // omit this one cap from the seed; expect exactly 1 parity violation
```
But `checkSeedRegistryParity` (`registry_rules.go:700`) **legitimately exempts deferred caps**. F2.1c introduced
the program's **first-ever** deferred cap (`CapDistributionRead`; `deferredCaps` was empty `{}` at the pre-M2
baseline). When the randomized `caps[0]` lands on `CapDistributionRead`, omitting it yields **0** violations →
`want 1, got 0` FAIL. Probability ≈ 1/30 per full-suite run. **Product code is correct** (the checker rightly
exempts deferred caps); the defect is the test's non-determinism, newly *exposed* by an M2 change.

`fullSeed` (line 94) also iterates `AllCapabilities()` but in an order-independent range loop — **not** affected.
Only the `caps[0]` index (line 137) is flaky.

## What to implement

Make the "missing-from-seed" capability **deterministic and guaranteed non-deferred**: sort the capability
slice and pick the first cap that is not in `deferredCaps`. `deferredCaps` is a package-level var in the same
`package main`, directly referencable from the test. Add a guard `t.Fatal` if (hypothetically) every cap were
deferred. No change to product code, the checker, the registry, or any other test.

## Validation Gate (acceptance — validator-specified)

| # | Criterion | How proven |
|---|-----------|------------|
| G1 | The flaky test is deterministic | `go test ./scripts/api-lint/... -run TestSeedRegistryParity -count=200` → green (was ~1/30 fail) |
| G2 | Package suite stable | `go test ./scripts/api-lint/... -count=50` → green |
| G3 | Full suite deterministically green | `go test ./...` green across **≥10 consecutive** full-suite runs (no `TestSeedRegistryParity_RegistryNotSeeded` flake) |
| G4 | Test still bites on real drift | The omitted cap is genuinely unseeded and non-deferred, so the checker still fires exactly 1 violation and the message still names that cap (assertions unchanged in intent) |
| G5 | No out-of-scope change | `git show --stat` of the F2.4 commit touches only `scripts/api-lint/registry_rules_test.go` (+ this feature's docs). No product/contract/migration/authz/registry change. |

## Non-goals

No change to `AllCapabilities()` ordering, the parity checker, `deferredCaps` contents, the registry seed, or
any other test. This does not alter what the guard enforces — only how the test selects its fixture cap.
