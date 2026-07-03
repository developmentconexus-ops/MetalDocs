# F2.1 plan — tripwire arm generation + drift check

> Executes `spec.md` against `../validation-contract.md §1`. TDD per task (failing test first).
> Implementation via subagents; main session reviews + commits. **Never retype the trigger from
> memory — regenerate from `db/migrations/0270_*.sql` (the prior file is the template).**

## Design decisions (fixed)

- **Source of truth:** `internal/platform/tripwire/arms.go` — `TripwireArms` (ordered slice or map of
  `{Table, Op, Caps []iamdomain.Capability}`) referencing iam registry **consts**. One-way import
  (tripwire → iam/domain); no cycle. Content = contract §1.2 exactly (18 entries; only
  documents(UPDATE) widened vs 0270).
- **Generation:** `internal/platform/tripwire/render.go` — `RenderMigration() string` renders the full
  0271 SQL from an embedded Go template that is **0270's file body verbatim** with the 19
  `v_required_caps := ARRAY[...]` literals substituted from `TripwireArms` (and header/ledger text
  updated). A tiny `cmd/gen-tripwire` (or a `-write` flag) writes it to disk; committing the output IS
  the migration. Determinism: caps rendered in a stable sorted order.
- **Parity enforced in api-lint, not codegen-drift** (the drift job globs only `api.gen.go`): api-lint
  **imports** `internal/platform/tripwire` and compares `RenderMigration()` to the committed
  `db/migrations/0271_*.sql` bytes. This is `TRIPWIRE-ARM-PARITY`. No `.github` change needed (rule
  runs under the existing blocking `api-design-system-lint`).
- **Migration:** `db/migrations/0271_documents_update_tripwire_membership_obsolete.sql`, forward-only,
  `CREATE OR REPLACE`, ledger insert `('0271', …)`, `BEGIN/COMMIT`. Header documents both latent
  incidents (force-release repository.go:798/828; obsolete_service.go:88→93) in 0269/0270 house style.

## Tasks

### Stage 1 — foundation (one sonnet subagent, sequential; everything depends on it)
- **T1.1** Author `internal/platform/tripwire/arms.go` (`TripwireArms` = contract §1.2) +
  `render.go` (`RenderMigration()` from the 0270 template) + `cmd`/`-write` generator.
- **T1.2** Generate & commit `db/migrations/0271_*.sql`; verify `go build ./...` green and the diff vs
  0270 is **only** the documents(UPDATE) arm + header/ledger (POSITIVE proof, contract §1.4).
- **T1.3** `arms_test.go`: (a) every cap in `TripwireArms` is `IsValidCapability`; (b) `RenderMigration()`
  == committed 0271 bytes (golden). Failing-first: write test before render is correct.

### Stage 2a — lint rules (one sonnet subagent, after Stage 1)
- **T2.1** `TRIPWIRE-ARM-PARITY` in `scripts/api-lint` (import tripwire; render==committed-0271 +
  all caps registry-real). Register in `RunCodeRules`/`RunRegistryRules`. Blocking.
- **T2.2** `TRIPWIRE-ARM-DRIFT` — extend the existing `checkTripwirePairing` technique: AST-scan
  functions; for a function that calls `authz.Require(cap)` **and** runs `Exec/ExecContext` mutating
  SQL on a gated table T (parse table after `INTO`/`UPDATE`/`DELETE FROM`), require `cap ∈
  TripwireArms[T, op]`. Reuse `parseCapabilityConsts` const-resolution (registry_rules.go). Blocking.
- **T2.3** Rule tests + `testdata` fixtures: PARITY RED on a mutated arm literal / a non-registry cap;
  DRIFT RED on a synthetic function asserting a new cap + writing a gated table with no arm (the
  mission's required synthetic — contract §1.5.b NEGATIVE); both GREEN on the clean tree post-0271.
- **Detection proof:** capture DRIFT rule RED against the **pre-0271** tree on the force-release +
  obsolete functions (run the rule with the old arm) — this is the "it catches the real latent bug"
  evidence (contract §1.5.b POSITIVE note).

### Stage 2b — integration drives (one sonnet subagent, after Stage 1; parallel to 2a)
- **T2.4** `tests/integration/documents/tripwire_documents_test.go` (`//go:build integration`, testdb
  factory, mirroring `tripwire_caps_test.go`):
  - `TestTripwire_ForceReleaseWritesDocumentRow` — actor with `membership.manage` force-releases a
    stuck session; `documents` UPDATE succeeds under the live tripwire.
  - `TestTripwire_ObsoleteWritesDocumentRow` — `MarkObsolete` on a published doc asserting
    `document.obsolete` succeeds.
  - Both proven **RED against 0270** (capture `ErrCapabilityNotAsserted … {document.edit} … documents`)
    and **GREEN against 0271**. Targeted `-run 'Tripwire'` only. If the box can't run integration,
    author + record the bounded defer with run-trigger.

## Verification (before F2.1 commit)
`go build ./...` · `go test ./internal/platform/tripwire/...` · `go test ./scripts/api-lint/...` ·
`go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` (0 violations, new rules blocking) ·
`go test ./internal/modules/iam/...` (`TestCapabilityRegistrySize`=34) · 5 authz lints green ·
targeted integration drives (or defer recorded).
