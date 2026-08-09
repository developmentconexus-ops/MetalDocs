# Distribution Module

> **Last verified:** 2026-08-09 (Phase G governance reconciliation — Stage-1 stub created; the module existed in the tree with no wiki doc, PASS 14 finding D5)
> **Scope:** Durable knowledge for `internal/modules/distribution/`. Stage-1 stub — records what is verified about the module surface; deep-dive sections to be filled by the next feature touching the module.

## What it is

Distribution/read-acknowledgement surface for published controlled documents:
which users must read a published document, and whether they have acknowledged
it. One of the 15 bounded-context modules under `internal/modules/`.

## Verified facts (audit 2026-08-09, `main@418070bf`)

- Reads cross-module data through **published DB views** (`v_*` read-model
  projections, ADR 0039 family) rather than foreign raw-table SQL — one of the
  2 modules already using the read-model-projection escape hatch (with
  `search`); see `docs/superpowers/analysis/audit-2026-08-09/pass03-modules-support.md`.
- Not part of the module SCC of size 9 (PASS 2) and carries no foreign-table
  writes (PASS 5).
- No `TenantDataPort` ownership entry exists for the views it consumes —
  noted as a soft addition to #93/A4's data-ownership catalog work.

## Open items

- Full module doc (routes, capabilities, schema, workflows) — owed by the
  next unit of work that touches distribution.
- No `distribution-tech-debt.md` register yet.
