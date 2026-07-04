# F3.3 plan — ADR 0027 + wiki amendment

> Contract: `../validation-contract.md` §3. Docs-only. Runs AFTER F3.1 + F3.2 committed (documents their
> shipped behavior). No code, no migration, no behavior change.

## Task list (ordered)

### T1 — ADR 0027 dated amendment
- Append `## Amendment 2026-07-03 (M3 tenancy chokepoint)` to
  `wiki/decisions/0027-rls-adoption-sequencing.md` (do NOT edit the original decision body). Cover the 5
  points from contract §3.1:
  1. NULL-permissive design is deliberate & load-bearing (GUC-unset → all rows, for GUC-less system/scan
     paths); not a bug; must not be removed.
  2. Pre-M3 sync↔async asymmetry: API seeded the GUC (real backstop); worker/jobs seeded nothing — async
     isolation rested on hand predicates; one bad worker join = silent cross-tenant leak, no gate.
  3. How M3 closes it: (a) TxRunner chokepoint autoseeds tenant+actor from the platform identity carrier
     (F3.1); (b) async single-tenant processing txs seed via `SeedTxTenant` (F3.2), completing ADR 0054
     rule 2; (c) two blocking lints (`SEED-CHOKEPOINT`, `ASYNC-TENANT-SEED`) + negative RLS integration
     proof make seeding structural, not discipline.
  4. Residual sanctioned GUC-unset surface: outbox claim, cross-tenant scans, `idempotency_keys` /
     `job_leases` (no `tenant_id` column) — enumerated, cross-referenced to ADR 0054.
  5. Cross-reference ADR 0054 + the M3 milestone folder.

### T2 — Wiki tenancy pages
- Find stale claims: `grep -rniE "async .*no .*backstop|~?85 sites|manually seed|only on controlled_documents|seed.*at.*sites" wiki`.
- Update each hit + the tenancy/RLS architecture pages so they reflect: chokepoint autoseed (API), async
  `SeedTxTenant` (worker/jobs), the two lints, and the per-binary posture (contract §4 table). Keep the
  wording aligned to shipped truth (read F3.1/F3.2 `evidence.md` before writing).

### T3 — wiki-curator pass
- Dispatch `wiki-curator` agent over the touched docs: refresh `Last verified` stamps, resolve file:line
  anchors, update `wiki/README.md` index if a new page/section was added.

### T4 — Gate + evidence
- PG-2 grep = 0 stale claims; PG-4 `git diff --stat` shows only `wiki/**`; curator verdict clean.

## Files expected to change
- `wiki/decisions/0027-rls-adoption-sequencing.md` (amendment)
- tenancy/RLS wiki pages (curator-identified) + possibly `wiki/README.md`

## Risk / ordering notes
- Runs last — it documents F3.1+F3.2 as shipped; author T1/T2 from their `evidence.md`, not from the plan
  (avoid documenting intended-but-changed behavior).
- Docs-only: any `.go`/SQL diff here is a scope breach — stop and reassess.
