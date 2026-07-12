# M3 — Approval kernel extraction — EVIDENCE

**Unit:** ROADMAP 3.1 (approval-remediation M3) · **Branch:** `claude/nice-wu-353cd4`
**Spec:** `docs/superpowers/specs/2026-07-12-m3-approval-kernel-extraction-plan.md`
**P0 gate:** `docs/superpowers/analysis/2026-07-12-approval-kernel-extraction-system-impact.md` (🟡 Yellow)
**Ratification:** all 4 items APPROVED by operator via hub (as recommended) — logged below.

## Ratification log
- 2026-07-12 — ESCALATION sent (commit ffe604c6 base). ACK: R1 additive route-admin contract; R2(a)
  thin template entry points, retire parallel path; R3 3-phase relocate-then-generalize; R4 count
  in-flight first then hard-cutover/drain. openapi edits authorized within R1/R2 shapes.

## Dispatch ledger (HARNESS §4.4)

| Slice | Dispatch | Implementer | Reviewer | Gates | Commit | Status |
|-------|----------|-------------|----------|-------|--------|--------|
| P1.S1 relocate tree + imports | — | sonnet | sonnet (indep) | go build | — | pending |
| P1.S2 re-port audit edge | — | sonnet | sonnet (indep) | boundary-lint GREEN | — | pending |
| P1.S3 composition + codegen | — | sonnet | sonnet (indep) | api-lint, go test | — | pending |
| P1.S4 supersede ADR 0072 + guard | — | sonnet | sonnet (indep) | negative-plant proof | — | pending |
| P2.S1 migration + backfill | — | sonnet | sonnet (indep) | testdb backfill | — | pending |
| P2.S2 domain generalize | — | sonnet | sonnet (indep) | byte-equal doc path | — | pending |
| P2.S3 route-admin contract delta | — | sonnet | sonnet (indep) | additive-only diff | — | pending |
| P3.S1 in-flight count | — | main/haiku | — | count + query recorded | — | pending |
| P3.S2 template entry points | — | sonnet | sonnet (indep) | tier-1 caps, kernel wire | — | pending |
| P3.S3 config→route migration | — | sonnet | sonnet (indep) | cutover rule applied | — | pending |
| P3.S4 retire parallel path | — | sonnet | sonnet (indep) | contract diff | — | pending |

## Gate results (fill per slice)
_(commands + outcomes appended as slices close)_

## Baseline (pre-work)
- Accepted RED on main: exactly 9 tests / 4 pkgs (E-PROD-1..5: sla_surfacer ×4, controlleddocuments
  cross-tenant sequence ×1, scenarios ×3, tenantdata ×1). Bar for every slice: zero NEW failures.
- approval subtree: 164 Go files. Coupling: 2 inbound production files (documents→approval), 24
  outbound (approval→documents domain/application), 1 true re-port (audit→approval/http/router),
  3 external consumers on allowed layers (audit, jobs/approval_sla_surfacer, jobs/stuck_instance_watchdog).

## Defers / notes
- E-PROD-2 (document_profiles PK) untouched — operator decision pending.

## HS-1
- Operator sign-off gate: pending (milestone close).
