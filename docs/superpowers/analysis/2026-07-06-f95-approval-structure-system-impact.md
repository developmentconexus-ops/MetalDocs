# System-impact analysis — F9.5 structure-hygiene (approval decision + layer rename + boundary-guard truth)

> **Date:** 2026-07-06 · **Class:** feature (M9 F9.5 mini-gate per mission.md M9 row)
> **Verdict:** **YELLOW** — proceed with the ADR-recorded exception path + guard realignment as locked
> constraints. Promotion of `documents/approval` is **rejected by evidence** (would trip AS-1/HS-2).

## 1. Orientation

- **Work:** (a) decide `documents/approval` top-level promotion vs ADR-recorded nested exception;
  (b) `repository/` → `infrastructure/` rename in documents + templates; (c) boundary-guard coverage
  for approval — plus the discovered prerequisite: `scripts/check-module-boundaries.ps1` is **RED on
  the current tree (55 violations)** because its allow-model (cross-module imports may target only
  `<module>/domain`) is **inverted** relative to REQ-TOP-1 ("access via application service or
  published Go interface — never repository, SQL, or domain internals").
- **Owning modules:** documents, templates (renames); the boundary guard is repo-level tooling
  (`scripts/`), governance-owned. Not-owners: no other module's code moves.
- **Cross-module edges (measured 2026-07-06):**
  - documents → approval: `approval/repository` ×42, `approval/domain` ×34, `approval/application`
    ×29, `approval/http/contracts` ×20 — and approval → documents: `documents/domain` ×16,
    `documents/application` ×6. **Bidirectional, dense, same-aggregate coupling.**
  - External consumers of approval: audit (handler), jobs/stuck_instance_watchdog (incl.
    `approval/repository` — a violation-class import), render/fanout/dispatchjobs,
    platform/tenantdata/registry, apps/api main.

## 2. Foundation judgment (global maximum, not local)

- **Approval nesting is NOT a patch.** DDD lens: approval is the approval *workflow of the documents
  aggregate* (frozen eligible-actor snapshots, SoD, dialect-B area-scoped authz per ADR 0022 — all
  keyed to document versions). Splitting one aggregate across two bounded contexts is a known
  anti-pattern; the 100+ bidirectional edges are the empirical proof this is one context.
  **Promotion would be structure-worship, not structure.** The global-maximum move is: keep nested,
  make the exception *explicit* (ADR), and guard approval's **external** surface.
- **The boundary script IS a patch-era artifact** (phase3, predates the target-architecture doc).
  Its only-`domain` allow-model contradicts REQ-TOP-1 and the living system (55 red edges, most of
  them the sanctioned `iam/authz` tier-2 interface). Optimizing around it (allow-listing 55 lines)
  would be a local maximum. The global maximum: realign the guard's allow-model to REQ-TOP-1's
  published-surface rule, with an explicit debt list for true violations found in the sweep.

## 3. Invariants (6)

| # | Touched? | Disposition |
|---|----------|-------------|
| 1 AuthZ capabilities | No caps added/changed. Rename must not alter any `authz.Require` call. | Satisfied by mechanical-move constraint |
| 2 Contract-first | No route/spec change. | N/A (locked: zero `api/openapi` diffs) |
| 3 Multi-tenant | No queries added/changed. | N/A |
| 4 Outbox | Untouched. | N/A |
| 5 DB invariants | No schema change. | N/A |
| 6 Cross-module published-interface | **Core of the work.** Guard realigned to REQ-TOP-1; approval external surface defined; true violations (e.g. watchdog → `approval/repository`) fixed-or-debt-listed explicitly, never silently allow-listed. | AS-1 risk if guard is *weakened*: locked constraint — new model must be stricter-or-equal on the violation classes REQ-TOP-1 names (repository/SQL/domain-internals) |

## 4. Wiring

- No capability wiring (skip `capability-wiring.md`).
- Module birth: **rejected** — `module-wiring.md` walked; promotion would require re-porting 100+
  edges through published ports (repository-class imports forbidden by invariant 6) — an interface
  redesign outside M9's boundary (HS-2). Recorded as the named alternative with its trigger:
  *if approval ever needs an independent lifecycle (own deploy cadence, own team, or a second
  consumer context), the promotion plan starts from this analysis.*

## 5. Module-birth rows

N/A — verdict is no new module (see §4).

## 6. Frameworks

No new frameworks. Rename reuses nothing new; guard realignment edits the existing PowerShell script
(kept — CI already invokes it; rewriting the tool is out of scope/YAGNI).

## 7. Test/QA gates

- `go build ./...` + targeted `go test` (documents/templates/approval packages) after rename.
- Boundary guard: RED→GREEN on realigned model **plus negative proof** (planted violation caught) —
  contract §5.3.
- api-lint blocking gate stays `0 blocking`.

## 8. Docs/ADR governance

- **ADR required (exception path):** new ADR — "documents/approval stays a nested subdomain of the
  documents bounded context (explicit exception to one-module-one-directory); external surface =
  published packages listed; boundary guard enforces both directions." Status per F9.1 rule.
- The same ADR (or its §) records the **guard-model realignment to REQ-TOP-1** and the debt list of
  true violations with owners/triggers.
- Wiki: `wiki/modules/documents.md` + architecture docs referencing `repository/` paths get F9.4
  stamps after rename.

## 9. Consumers

- Boundary guard consumers: CI `module-boundaries.yml`, milestone close gate §6.2.
- Rename consumers: every import site (39 files import `documents/repository`, 6 import
  `templates/repository`) — compiler-verified.

## 10. Verdict & locked constraints for design

**YELLOW.** Proceed with:
1. **Exception path** for approval (ADR-recorded), promotion rejected-with-trigger.
2. Rename `repository/`→`infrastructure/` mechanical-only; templates folds into its existing
   `infrastructure/` without nesting; approval's own `repository/` folds into approval's
   `infrastructure/` (one convention everywhere).
3. Guard realignment to REQ-TOP-1 published-surface model; **stricter-or-equal** on
   repository/SQL/domain-internal imports; true violations fixed or explicitly debt-listed in the
   ADR — silent allow-listing forbidden.
4. Zero contract/schema/capability diffs.

Hard-stops armed: AS-1/HS-2 if the guard realignment tempts an interface redesign of live modules;
HS-6 if the violation sweep uncovers a defect class beyond import hygiene.
