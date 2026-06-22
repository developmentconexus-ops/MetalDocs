# Feature F2.4 — Plan (deferredcap-parity-test-determinism)

> **Spec:** `./spec.md`. Single-file test-determinism fix; controller-executed (mechanical, root cause
> confirmed at source). The re-dispatched milestone-validator is the independent check (separation of powers).

## T1 — Make the fixture-cap selection deterministic + non-deferred
`scripts/api-lint/registry_rules_test.go` ~L136-137, in `TestSeedRegistryParity_RegistryNotSeeded`:
- Replace `missing := string(caps[0])` with: sort `caps` lexicographically, then pick the first cap **not** in
  `deferredCaps` (package var, same `package main`). Guard with `t.Fatal` if none found.
- Add the `sort` import.
- Add a one-line comment citing the map-randomization + deferred-cap-exemption root cause so the next reader
  doesn't reintroduce the flake.

## T2 — Verify determinism (controller)
- `go test ./scripts/api-lint/... -run TestSeedRegistryParity -count=200` → green (G1).
- `go test ./scripts/api-lint/... -count=50` → green (G2).
- `go test ./...` ×≥10 consecutive → all green, no parity flake (G3).
- Confirm the test still asserts 1 violation + names the omitted cap (G4 — assertions unchanged).
- `git show --stat` confirms only the test file + docs touched (G5).

## T3 — Close
- Write `evidence.md` (real captured output: the 200×/50×/≥10× runs, diff scope). Commit
  `fix(M2/F2.4): make seed-registry-parity test pick a deterministic non-deferred cap`.
- Re-dispatch `milestone-validator` (HS-4): the M2 substance already PASSed C1/C3/C5; this clears C2/C4/C6.

## Risk
- **R1 — over-fix.** Do NOT change `AllCapabilities()` to return sorted order (would touch product code + risk
  other call sites). Keep the determinism local to the test. The fix is selection-side only.
