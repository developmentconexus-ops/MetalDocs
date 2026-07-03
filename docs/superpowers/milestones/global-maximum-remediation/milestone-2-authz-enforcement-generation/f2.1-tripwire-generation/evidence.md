# F2.1 evidence — tripwire arms GENERATED from the Go registry + CI drift check

> Closes F2.1 against `../validation-contract.md §1`. Implemented via subagents (Stage 1 foundation,
> Stage 2a lint rules, Stage 2b integration drives); every gate independently re-run by the main
> session (anti-circle). **Additive-only** change: the single behavior delta is the documents(UPDATE)
> tripwire arm widening, closing two latent P0001 incidents.

## What shipped

| Layer | File | Role |
|---|---|---|
| Source of truth | `internal/platform/tripwire/arms.go` | `TripwireArms` — 18 ordered `{Table, Op, Caps}` arms referencing iam registry consts. Content = contract §1.2 exactly. |
| Generator | `internal/platform/tripwire/render.go` + `cmd/gen-tripwire/main.go` | `RenderMigration()` renders 0271 SQL from the 0270 body verbatim with arm literals substituted from `TripwireArms`. |
| Migration (generated) | `db/migrations/0271_documents_update_tripwire_membership_obsolete.sql` | forward-only `CREATE OR REPLACE enforce_capability_asserted()`; ledger `('0271', …)`. Only branch differing from 0270 = documents(UPDATE). |
| Parity rule | `scripts/api-lint/tripwire_arm_rules.go` → `TRIPWIRE-ARM-PARITY` | imports tripwire; every arm cap registry-real + `RenderMigration()` byte-equals committed 0271. Blocking. |
| Drift rule | same file → `TRIPWIRE-ARM-DRIFT` | function-local AST scan; `authz.Require(cap)` + mutating SQL on a gated table ⇒ `cap ∈ TripwireArms[table,op]`. Blocking. |
| Integration drives | `tests/integration/documents/tripwire_documents_test.go` | `//go:build integration`, testdb factory: force-release + obsolete write paths, RED-vs-0270 / GREEN-vs-0271. |

## The one behavior change (contract §1.1) — documents(UPDATE) arm widened

Census confirmed **two** function-local latent P0001 incidents (review predicted one):

| Path | Asserts | Wrote documents(UPDATE) | Pre-0271 arm `{document.edit}` verdict |
|---|---|---|---|
| `ForceReleaseSession` / `…Tx` (`repository.go:798,828`) | `CapMembershipManage` only | `:811 / :841` | fail-closed P0001, every actor |
| `MarkObsolete` (`obsolete_service.go:88→93`) | `CapDocumentObsolete` only | `:93` | fail-closed P0001, every actor |

0271 widens the arm additively to `{document.edit, document.obsolete, membership.manage}` (0269/0270
convention — match-one, no regression to existing `document.edit` writers). Contract §1.2 entry #6.

## Positive proof (contract §1.4 / §1.5.a) — independently re-run

```
go build ./...                                 → exit 0
go test ./internal/platform/tripwire/...       → ok   (golden: RenderMigration()==0271 bytes; every cap IsValidCapability; deterministic)
go test ./scripts/api-lint/...                 → ok   (15 tripwire-arm rule tests + suite)
go test ./internal/modules/iam/... -run TestCapabilityRegistrySize → ok (want=34)
go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .    → 0 violation(s)  (both new rules blocking, clean tree)
```

**0271 vs 0270 diff** = only the documents(UPDATE) `v_required_caps := ARRAY[...]` literal + header/ledger
text (POSITIVE proof, contract §1.4). Regenerating via `RenderMigration()` reproduces the committed file
byte-for-byte (`TestRenderMigration_MatchesCommittedFile` golden).

## Negative proof (contract §1.5.b) — the drift check actually catches the bug

**Synthetic (mission-required):** rule-test fixtures assert `TRIPWIRE-ARM-DRIFT` RED on a synthetic
function asserting a new cap while writing a gated table with no matching arm, and `TRIPWIRE-ARM-PARITY`
RED on a mutated arm literal / non-registry cap — both GREEN on the clean post-0271 tree
(`scripts/api-lint/tripwire_arm_rules_test.go`, 15 tests). A synthetic new asserted cap fails **CI**, not
runtime.

**Real-bug detection proof — independently reproduced by the main session:** reverted
`internal/platform/tripwire/arms.go` documents(UPDATE) arm to `{CapDocumentEdit}` (the pre-0271 state),
ran `go run ./scripts/api-lint -only TRIPWIRE-ARM-DRIFT …`:
```
3 violation(s)
  internal/modules/documents/approval/application/obsolete_service.go:93  MarkObsolete            asserts {document.obsolete}   writes documents(UPDATE)
  internal/modules/documents/repository/repository.go:811                 ForceReleaseSession     asserts {membership.manage}   writes documents(UPDATE)
  internal/modules/documents/repository/repository.go:841                 ForceReleaseSessionTx   asserts {membership.manage}   writes documents(UPDATE)
  … "the DB tripwire will reject this write with P0001 for every actor — widen the arm in internal/platform/tripwire/arms.go"
```
`git checkout -- internal/platform/tripwire/arms.go` → re-ran → `0 violation(s)`. The rule fires on
exactly the two real latent incidents this feature fixed — it is not a tautology against its own arm.

## Integration drives (contract §1.6) — authored + bounded-defer

`tests/integration/documents/tripwire_documents_test.go` compile-verified (`//go:build integration`,
package `documents_test`, testdb factory mirroring `tripwire_caps_test.go`):
`TestTripwire_ForceReleaseWritesDocumentRow` (drives `repository.ForceReleaseSession`),
`TestTripwire_ObsoleteWritesDocumentRow` (drives `approvalapp Obsolete.MarkObsolete`).

**Live run bounded-deferred (contract §1.6 / §4).** No DB URL in the process env
(`METALDOCS_DATABASE_URL` / `DATABASE_URL` both unset — confirmed) and creds live only in `.env` (hard
rule: never read/source `.env`); no sanctioned env-loading PowerShell test runner exists on this box
(`scripts/test.ps1` is a bare `go test ./...`, no `-tags integration`, no env load). Author-and-defer is
explicitly permitted by contract §1.6/§4.
- **Run-trigger:** `METALDOCS_DATABASE_URL=<tenant DSN> go test -tags integration ./tests/integration/documents/ -run Tripwire`.
- **Expected:** RED against 0270 (`ErrCapabilityNotAsserted … {document.edit} … documents`), GREEN
  against 0271.
- **Compensating proof:** the `TRIPWIRE-ARM-DRIFT` detection proof above independently demonstrates the
  same two write paths were fail-closed pre-0271 and are covered post-0271 — static, box-runnable, and
  already GREEN in CI.

## Review / QA disposition
Stage 1 / 2a / 2b each: sonnet implement + review; every gate independently re-executed by the main
session from a clean tree (not trusting subagent reports). Stage 2a subagent additionally fixed a real
interop bug between the missing-migration check and the `e2e_test`/`main_test` fixtures
(`requireCoreFile` strict-hard-error / non-strict-skip convention) — surfaced, not silently absorbed.

## Bounded defers
1. **Integration live run** — trigger + expected + compensating proof above (contract §1.6/§4).
2. **`templates_template` arm `template.submit` prune** — deliberately retained harmless superset (no
   writer asserts submit while writing that table); tightening deferred to **M9 arm-hygiene**
   (contract §1.3). Not a defect; recorded so the superset is not mistaken for drift.

## Acceptance (contract §1.5) — MET
Arms generated from the same registry Go reads (not hand-synced) · drift check RED on synthetic new
asserted cap lacking an arm + on the real pre-0271 latent bugs, GREEN on clean tree · existing
`tripwire_caps_test.go` still GREEN (unchanged) · synthetic cap fails CI not runtime · 5 authz lints +
2 new tripwire-arm lints all green · registry pin 34 intact.
