# F-R3 — Test-discipline green + defer ratification (Dim 8 → CONFIRMED)

> Consumer-contract-first. Approved **before** the feature's commit.
> **Approval:** APPROVED 2026-07-06 (operator "Go" on M10; contract self-reviewed against
> `validation-contract.md` C5). — *filled before the F-R3 commit.*

## Problem (the DEBT this closes)

`bash scripts/check-test-discipline.sh` was **RED at HEAD** (4 violations), so the Dim-8 gate did not
actually pass from a clean state. Three violations predate the mission; **one is mission-introduced**
(M9 F9.5 renamed `templates/repository/` → `templates/infrastructure/` but left the R2 allowlist path
stale). A gate that is red at HEAD is not a gate.

Separately, two §8 uncovered-MUST requirements (REQ-SEARCH-1, REQ-SEC-3) are absent **product
features**, not hygiene — they must be **ratified as bounded defers with triggers**, not silently
carried.

## Consumers & the contract each requires

| Consumer | Required shape |
|----------|----------------|
| **`check-test-discipline.sh` CI gate** | Exits 0 (`test-discipline: clean`) at HEAD — all 4 violations resolved at root. |
| **Future maintainer** | The R2 allowlist path matches the file's real location (post-F9.5); allowlists did not widen (only a path correction). |
| **Terminal acceptance / `req-trace`** | The uncovered-MUST set stays exactly `{REQ-AUTHN-1, REQ-AUTHN-3, REQ-SEARCH-1, REQ-SEC-3}` — F-R3 closes none and adds none; REQ-SEARCH-1/REQ-SEC-3 are ratified defers, not implementations. |
| **Program record** | A defer ledger entry for REQ-SEARCH-1 and REQ-SEC-3 with `{finding, why-absent, trigger, owner}`. |

## The 4 violations and their root-cause fixes

| # | Location | Rule | Fix (class) |
|---|----------|------|-------------|
| 1 | `templates/infrastructure/tenant_id_rls_integration_test.go:148` | R2 | **Allowlist path correction** — update the stale `repository/` path to `infrastructure/` (F9.5-rename reconciliation). The RLS probe is a legitimate sanctioned single-pinned-conn GUC set; it was allowlisted before the rename. Not a widening. |
| 2 | `jobs/stuck_instance_watchdog/job_integration_test.go:186` | R4 | **Qualify** the bare `documents` → `metaldocs.documents`. |
| 3–4 | `controlleddocuments/domain/sequence_test.go:57,123` | R1 | **Migrate** the two inline `asserted_caps` set_config sites to `testdb.SetCapsOnTx(t, tx, …)` — the sanctioned tx-local primitive. Repair-class: same tripwire assertion, sanctioned helper. |

## Non-goals (mandatory)

- No change to the discipline rules R1–R4 or the legacy-test-policy taxonomy.
- No allowlist **widening** — only the one F9.5 path correction (allowlists may only shrink or be
  path-corrected).
- No migration of the whole `sequence_test.go` to the testdb factory (it stays a DATABASE_URL-gated
  integration test); only the R1 inline set_config is replaced.
- No implementation of REQ-SEARCH-1 or REQ-SEC-3 — ratify as defers only.
- No drive-by repair of any test outside the 4 named violators.

## Validation Gate

1. `bash scripts/check-test-discipline.sh` → exit 0, `test-discipline: clean`.
2. `go vet -tags integration` on the three edited packages → exit 0 (compiles; no testdb import cycle).
3. Allowlist diff shows only the one path correction (no new entries).
4. `go run ./scripts/req-trace` → uncovered set unchanged (4), `stale=false`, exit 1.
5. Defer ledger records REQ-SEARCH-1 + REQ-SEC-3 with finding/trigger/owner.

## ADR?

No. Repair + allowlist reconciliation + defer ratification; no durable new decision.
