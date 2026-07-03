# F2.1 — Tripwire arms generated from a Go source of truth + CI drift check

> **Milestone:** M2 · **Governing contract:** `../validation-contract.md` §0–§1 (binding; HS-7 on drift).
> **Approved for implementation:** 2026-07-03 (design fixed in the validation-contract before code).

## Consumer contract

The **consumer** is the CI system + the DB security trigger. It requires:

1. **A single Go source of truth `TripwireArms`** — map of gated `(table, op)` → required-cap set,
   referencing iam registry capability **consts** (never raw strings). Content = exactly the 18-entry
   table in `../validation-contract.md §1.2`. Only `documents(UPDATE)` differs from migration 0270:
   `{document.edit, document.obsolete, membership.manage}`.
2. **The trigger migration generated from that map** — forward-only migration `0271`, a
   `CREATE OR REPLACE FUNCTION public.enforce_capability_asserted()` **regenerated from 0270** (the
   prior file is the template; do not retype from memory), byte-faithful except the arm literals the
   map dictates. `go generate ./...` must reproduce the committed SQL (drift-checked by the existing
   `backend-codegen-drift` job); the SQL is a generated artifact, not hand-typed.
3. **A blocking CI drift check** (`scripts/api-lint`, registered in `RunCodeRules`/`RunRegistryRules`,
   runs under `api-design-system-lint`):
   - `TRIPWIRE-ARM-PARITY`: every `TripwireArms` cap is registry-real; the generated SQL matches the map.
   - `TRIPWIRE-ARM-DRIFT`: AST-scan `authz.Require` sites; a **function-local** gated-table write whose
     asserted cap is absent from that table's arm fails CI. (Function-local only — cross-layer is
     covered by integration drives; documented limitation, not a gap.)
4. **Integration proof** (only test class that can pin a `P0001` arm regression — app tests are sqlmock):
   two new drives (`TestTripwire_ForceReleaseWritesDocumentRow`, `TestTripwire_ObsoleteWritesDocumentRow`)
   GREEN against 0271, RED against 0270.

## Non-goals

- **No arm tightening / superset pruning** (e.g. `template.submit` on `templates_template` stays) —
  additive-only this milestone; pruning deferred to M9 (contract §1.3).
- **No cross-layer call-graph analysis** — the drift check is function-local by design (HS-2 boundary).
- **No trigger-attachment / down-migration / backfill changes** — function-body swap only.
- **No rewrite of the trigger to a generic `cap_table_requirements` lookup table** — that is a
  load-bearing security-trigger redesign (HS-2); generation keeps the proven trigger shape.

## Validation gate

Per `../validation-contract.md §1.7` (exit criteria) — copied as the acceptance checklist into
`evidence.md`. Positive + negative proof for each of: parity rule, drift rule (incl. the mission's
synthetic-unarmed-cap RED), generated 0271 diff-minimal-vs-0270, two integration drives (GREEN post /
RED pre), and the no-regression set (0269+0270 drives, `TestCapabilityRegistrySize`=34, 5 authz lints,
`go build ./...`).
