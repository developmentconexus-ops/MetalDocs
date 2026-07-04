# F4.3 — concurrency-idiom decision (ADR-record the split)

> **Contract:** `../validation-contract.md` §3 — see the **§3.7 HS-7 erratum (2026-07-04)**: the original
> "migrate templates to If-Match" decision was superseded after verification proved its premise false.
> **Approved for code: 2026-07-04** (decision-only feature — no product code changes).

## Consumer contract

- **Consumer = API clients + the generated FE client + future maintainers.** They require the OCC
  transport situation to be **decided and recorded**, not left as an undocumented accidental split.
- **Decision (contract §3.7, ADR 0066):** the two OCC transports are an **intentional, documented split**:
  documents/approval use the `If-Match` header; templates uses the body `expected_lock_version` field.
  Each is self-consistent within its module. `If-Match` is named the **long-term target** transport; full
  unification is deferred to its own deliberate cross-module change (candidate M9), not M4.
- **No wire change ships in M4.** documents and templates are both unchanged. The deliverable is ADR 0066
  + the contract erratum.

## Why the original decision was dropped (false-premise correction)

The original §3.4/§3.5 decision ("unify on If-Match, migrate templates the minority") assumed templates
was already near the If-Match convention and just needed a bounded cleanup. Verification (2026-07-04):

- templates has **zero** If-Match usage (BE + FE); `expected_lock_version` is its **only** OCC write,
  self-consistent with its own `lock_version` column + `stale_lock_version` error.
- The cited "system-wide If-Match standard" is **CON-01** in `wiki/modules/documents.md` — a
  documents-**module-internal** decision, not a cross-module OCC ADR. Nothing templates violates.
- Migrating templates would therefore **create a new cross-module standard**, not finish a convergence —
  a first-class architectural change, out of scope for a **versioning-kernel-correctness** milestone, and
  needless regression risk on a correct module. CLAUDE.md: stop on architecture contradictions; respect
  module boundaries.

This is the contract's own §3.5-permitted fallback ("ADR-record the split"), taken via HS-7 (loud dated
erratum, ratified at the HS-1 gate).

## Non-goals

- NOT migrating templates to `If-Match` in M4 (deferred to its own change).
- NOT renaming the `lock_version` column.
- NOT changing documents (already on `If-Match`).
- NOT the state machine (F4.1) or the publish race (F4.2).

## Validation gate

Per contract §3.7 (corrected exit criteria): ADR 0066 landed under `wiki/decisions/` recording the
intentional split + If-Match target + deferred-unification charter · contract §3 re-opened with the dated
erratum (not silent) · this spec + plan updated · `evidence.md` records the false-premise correction + HS-7
disposition · **no templates/documents wire change in M4** · HS-7 re-open headlined at the HS-1 operator
gate. Decision-only feature — no BE/FE/openapi build steps apply.

## Interview record

| Q | Operator answer |
|---|---|
| Idiom choice | "Best solution — full analysis: what we have, what we want, a fresh professional impl." → delegated to the §3 analysis. |
| (analysis outcome, corrected) | Analysis first decided "unify on If-Match + migrate templates." Verification proved the "templates straggler" premise false → **corrected to ADR-record the split** (ADR 0066), If-Match named as target, unification deferred. HS-7 re-open surfaced for ratification at the HS-1 gate. |
