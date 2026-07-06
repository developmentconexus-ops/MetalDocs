# F9.5 — structure-hygiene (feature spec)

> **Milestone:** M9 governance-hygiene · **Contract:** `../validation-contract.md` §5 (binding)
> **Mini-gate:** `docs/superpowers/analysis/2026-07-06-f95-approval-structure-system-impact.md`
> (YELLOW — locked constraints §10 govern this spec)
> **Approved:** 2026-07-06 — approved against mission.md M9 row + contract §5 + mini-gate verdict
> (operator-locked sources; autonomous session per mission D2). Code may start.

## Consumer contract (first)

**Consumer 1 — module-boundary CI guard** (`scripts/check-module-boundaries.ps1`, invoked by
`module-boundaries.yml` + milestone close gate): after F9.5 it encodes REQ-TOP-1's actual rule —
cross-module imports may target only a module's **published surface**; imports of another module's
`repository`/`infrastructure` (persistence), SQL, or non-published internals FAIL. It treats
`documents/approval` per the ADR: internal to documents for intra-module edges; its **external
surface** explicitly listed for outside consumers. GREEN on the final tree; RED on a planted
violation.

**Consumer 2 — future module work relying on layout convention:** exactly one persistence-layer dir
name exists under `internal/modules/*`: `infrastructure/`. Zero `repository/` dirs remain
(documents, templates, documents/approval).

**Consumer 3 — architecture record:** a new ADR states (a) approval-nested exception + external
surface + promotion trigger, (b) the guard-model realignment rationale (REQ-TOP-1), (c) the true-
violation debt list (each: edge, why it exists, fix-now or trigger). Status field per F9.1 rule.

## Interview record (B1.5 — resolved from normative sources + mini-gate)

| Q | A | Source |
|---|---|--------|
| Promote approval? | No — ADR-recorded exception. Coupling evidence: documents→approval/{repository×42, domain×34, application×29}; approval→documents/{domain×16, application×6}; one aggregate. Promotion trigger recorded. | Mini-gate §2/§4 |
| Guard allow-model? | Published-surface: cross-module target must be in {`domain`, `application`, `api`} of the owner (domain carries the published ports per invariant-checklist §6; historical type-sharing), plus tool packages published as interfaces (`iam/authz`, `render/fanout`, `render/resolvers` — verified real + sanctioned). Persistence (`repository`/`infrastructure`), `delivery`, `jobs`, `http` of another module: FORBIDDEN. | REQ-TOP-1; runtime sweep 2026-07-06 |
| Stricter-or-equal proof? | Old model allowed cross-module `domain` only; new model still forbids every class REQ-TOP-1 names (persistence/SQL/internals). The 55 baseline reds are reclassified: sanctioned-published (majority: iam/authz etc.) vs true violations (e.g. stuck_instance_watchdog → `documents/approval/repository`). True violations: fix in-feature if mechanical (consume approval's application/port instead), else debt-list in the ADR with trigger — NEVER silent allow-list. | Mini-gate §10.3 |
| approval external surface? | Derived from real external consumers (audit, watchdog, render/dispatchjobs, platform registry, apps main): allowed = `approval/application`, `approval/domain`, `approval/api`, `approval/http/contracts` (wire DTO package consumed by audit/main wiring; verify at impl). External `approval/repository`/`infrastructure` imports = violations. | Import sweep 2026-07-06 |
| Rename mechanics? | `git mv` + package rename `repository`→`infrastructure`; templates: fold files into existing `infrastructure/` (flat, no nesting) — resolve any filename collisions by rename-with-suffix; imports updated repo-wide; alias names in importers updated only where the compiler forces. | Contract §5.1 |
| apps/ and tests? | All importers updated including tests and binaries; `go build ./...` is the completeness proof. | Contract §5.4 |

## Non-goals (mandatory)

- No promotion of approval; no new module.
- No exported-signature, behavior, SQL, contract, capability, or schema changes.
- No rewrite of the boundary script's language/tooling (stays PowerShell; model change only).
- No fixing of debt-listed violations beyond mechanical port-consumption swaps (anything needing a
  new port method or redesign → debt list with trigger, HS-2 respected).
- No wiki restructure (stamps/path fixes belong to F9.4's final pass).

## Validation Gate

1. `find internal/modules -type d -name repository` → empty.
2. `go build ./...` exit 0; targeted `go test ./internal/modules/documents/... ./internal/modules/templates/... -count=1` green (DB-dependent suites: run those runnable on this box; label runs real-vs-skip in evidence).
3. `scripts/check-module-boundaries.ps1` → GREEN on final tree.
4. **Negative proof:** plant a forbidden import (an external file importing
   `documents/approval/infrastructure`) → script RED naming it → revert; both outputs captured.
5. ADR committed (exception + guard model + debt list), status per F9.1 rule; sweep still GREEN.
6. api-lint blocking gate `0 blocking` unchanged.
7. Diff class: moves/renames/imports + script + ADR only (validator checks no behavior diffs).
