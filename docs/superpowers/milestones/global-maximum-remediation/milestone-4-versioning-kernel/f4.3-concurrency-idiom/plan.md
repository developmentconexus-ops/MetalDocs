# F4.3 plan (corrected — decision-only)

> Original plan ("migrate templates to If-Match") superseded 2026-07-04 by the contract §3.7 HS-7 erratum.
> Corrected decision: **ADR-record the two-idiom split** (ADR 0066). No product code changes.

## Task

Record the OCC-transport decision. No BE/FE/openapi/regen work.

Files:
- `wiki/decisions/0066-optimistic-concurrency-transport-split.md` — ADR: intentional split
  (documents=`If-Match`, templates=body `expected_lock_version`), `If-Match` named as target, full
  unification deferred to its own cross-module change. RFC 7232 / AIP-154 / Zalando basis. **DONE.**
- `../validation-contract.md` §3 — re-opened with the dated §3.7 erratum documenting the false premise
  and switching to the ADR-split decision (HS-7). **DONE.**
- `spec.md` (this feature) — updated to the corrected decision. **DONE.**
- `evidence.md` — records verification commands (zero templates If-Match; endpoint tags), the
  false-premise correction, and the HS-7 disposition.

## Verification evidence (the analysis that flipped the decision)

- `grep -rln "If-Match\|IfMatch\|parseIfMatch" internal/modules/templates/ frontend/apps/web/src/features/templates/` → **empty** (zero templates If-Match).
- The 3 If-Match OpenAPI operations are tagged `[documents]` / `[approval]` — none `[templates]`.
- templates `expected_lock_version` is its only OCC write (`routes_schema.go`), self-consistent with
  `lock_version` + `stale_lock_version`.
- "DEC-01" = CON-01 (`wiki/modules/documents.md`), a documents-internal decision, not a system-wide ADR.

## Gate

ADR 0066 present + cited · contract §3.7 erratum present + dated · spec/plan/evidence consistent · no
templates/documents wire change · HS-7 re-open headlined at the HS-1 gate. (No build/test steps — this
feature ships documentation only.)

## HS-7 disposition

Contract re-open done via loud dated erratum (not silent edit), per the contract §0 HS-7 rule and the
§3.5 fallback clause. Operator ratification deferred to the M4 HS-1 gate (nothing pushed/merged before).
