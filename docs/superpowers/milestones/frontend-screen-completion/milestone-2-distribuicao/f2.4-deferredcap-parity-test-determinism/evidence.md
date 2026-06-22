# Feature F2.4 — Evidence

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Feature:** `f2.4-deferredcap-parity-test-determinism`  ·  **Closed:** 2026-06-22
> **Opened under HS-4** by the M2 milestone-validator FAIL (`../qa/milestone-qa.md`, C7). **Spec:** `spec.md` · **Plan:** `plan.md`.
> Pure test-determinism fix — no product, contract, migration, authz, or registry behavior change.

## What was implemented

One change, one file: `scripts/api-lint/registry_rules_test.go`,
`TestSeedRegistryParity_RegistryNotSeeded`. Replaced the flaky `missing := string(caps[0])` (which indexed the
**map-randomized** output of `iamdomain.AllCapabilities()`) with a deterministic selection: sort the capability
slice lexicographically, then pick the first cap **not** in `deferredCaps`. Added the `sort` import + a
root-cause comment + a `t.Fatal` guard for the (hypothetical) all-deferred case.

Why this is the right fix: `checkSeedRegistryParity` legitimately **exempts deferred caps** (`registry_rules.go:700`).
F2.1c introduced the program's first-ever deferred cap (`CapDistributionRead`; `deferredCaps` was empty `{}` pre-M2),
so a random `caps[0]==CapDistributionRead` omitted from the seed produced **0** violations → `want 1, got 0`
(~1/30 full-suite runs). The product code is correct; only the test's fixture-selection was non-deterministic.
The fix is selection-side only — `AllCapabilities()` ordering, the checker, and `deferredCaps` are untouched.

### Commit
```
<this>  fix(M2/F2.4): make seed-registry-parity test pick a deterministic non-deferred cap
```

## Verification (real output)

| # | Criterion | Command | Result |
|---|-----------|---------|--------|
| G1 | Flaky test deterministic | `go test ./scripts/api-lint/... -run TestSeedRegistryParity -count=200` | `ok metaldocs/scripts/api-lint 6.772s`, exit 0 |
| G2 | Package suite stable | `go test ./scripts/api-lint/... -count=50` | `ok metaldocs/scripts/api-lint 225.489s`, exit 0 |
| G3 | Full suite deterministically green | `go test ./...` ×10 consecutive (`-count` fresh each) | **run 1..10: PASS** (10/10), exit 0 |
| G4 | Still bites on real drift | The chosen cap is non-deferred + genuinely unseeded → checker fires exactly 1 violation, message names it; assertions (`countRule==1`, `containsMsg(missing)`) unchanged in intent | held (test asserts the same; passes deterministically) |
| G5 | No out-of-scope change | `git status --short` / commit `--stat` | only `scripts/api-lint/registry_rules_test.go` (+ this feature's docs + the validator's `qa/milestone-qa.md`) — no product/contract/migration/authz/registry change |

Old flake probability was ~1/30 per full-suite run. **250 clean package runs (G1+G2) + 10 clean full-suite runs (G3)** =
zero failures across 260 executions → the flake is eliminated, not merely unobserved.

## Bounded defers

None. The fix is complete and self-contained.

## Next

Re-dispatch `milestone-validator` (HS-4). The M2 substance already PASSed C1/C3/C5 on the prior run with live-PG
evidence; this feature clears the sole blocker (C2/C4/C6 — the non-deterministic CI guard).
