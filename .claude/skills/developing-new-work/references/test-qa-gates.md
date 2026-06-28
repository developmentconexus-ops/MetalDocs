# Test & QA gates

**Last verified:** 2026-06-28

## Canonical test framework
New tests MUST use the canonical framework for their class (test-framework hard gate). Do not write a
bespoke harness.
- **Integration:** `tests/integration/testdb/` — `Open(t)`, factory builders, `SeedWithCaps`,
  `Qualified`. Tag with `//go:build integration`.
- **Discipline R1–R4:** enforced by `scripts/check-test-discipline.sh`; rules in
  `wiki/quality/test-discipline.md` and ADR 0034 (`wiki/decisions/0034-integration-test-fixture-framework.md`).
- **Legacy tests:** many existing tests are one-off task scaffolding — delete (don't maintain) when
  they break; repair only contract/invariant guards. Drive-by repair of a pre-existing test must still
  land on the canonical framework.

## The 6 QA gates
The QA operating system (`wiki/quality/qa-operating-system.md`) defines the gates; the relevant
per-class checklists live in `wiki/quality/*-checklist.md`. Decide which apply:
- A **new module** typically exercises all gates (contract, authz, multi-tenant isolation, async/
  idempotency, DB-invariant, docs).
- A **feature** exercises the subset its change touches — name them, mark the rest N/A.

## Evidence shape (CLAUDE.md "evidence before closure")
Report, before saying done: **commands run + their outcomes + QA/review disposition + any bounded
defers**. No bare "done".

Baseline commands:
- `go build ./...`
- `go test ./...`
- `.\scripts\check-system-runnable.ps1` (PowerShell; never `source .env` / bash for startup)
- integration: `go test -tags=integration ./...`

For runtime-observable work, drive the running app and capture evidence (screenshot/network/logs) per
the verify discipline — don't ask the operator to check manually.
