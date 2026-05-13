# Module Wiki Maturity Standard

Use this standard for `wiki/modules/*` living docs. The target is operational memory: a future LLM or developer should be able to plan a feature, fix a bug, spot legacy behavior, and avoid broken assumptions without rereading the whole module.

## Mature Module Doc

A mature module wiki entry has all of these:

1. Complete doc set
   - `wiki/modules/<m>.md`
   - `wiki/modules/<m>-tech-debt.md`
   - `wiki/backlog/<m>-refactor.md`
   - `wiki/modules/<m>/_artifacts/00-context.md`
   - `01-surface.md`, `02-flow-*.md`, `03-deps.md`, `04-persistence.md`, `05-industry.md`, `06-selfreview.md`
   - `_artifacts/sync-log.md` after the first incremental sync

2. Implementation map
   - Key files with verified `path:line` anchors.
   - Public surface grouped by file, with exported symbols summarized or intentionally excluded.
   - Runtime owner for each important HTTP route, handler, service, repo, job, and generated API method.

3. API and contract truth
   - HTTP operations table.
   - API Route Truth Table with runtime route, OpenAPI path, operationId, codegen method, status, and notes.
   - Explicit contract status such as `Contracted`, `Spec missing`, `Runtime missing`, `Bootstrap only`, or `Legacy/manual`.
   - Permission/capability mapping for guarded routes.

4. Runtime behavior
   - Sequence diagrams or step traces for representative read, write, and state-transition flows.
   - State transition table when the module owns lifecycle state.
   - Failure modes and current error envelope, including gaps against the target contract.
   - Transaction boundaries, idempotency behavior, audit emission, and authz layers.

5. Persistence and data ownership
   - Owned tables, read tables, migrations, constraints, FKs, triggers, indexes, tenant columns, and retention/deletion behavior.
   - Tripwire/GUC pairing for regulated mutations.
   - Cross-module data contracts and snapshots.

6. Cross-dependencies
   - Inbound and outbound Go package edges.
   - Composition-root wiring.
   - Cross-module conceptual dependencies and wiki cross-links.
   - C4 relations that match the dependency facts.

7. Decisions and debt
   - Load-bearing decisions link to ADRs or are logged as `missing-ADR`.
   - Tech debt rows are evidence-backed, severity-rated by rubric, and linked to backlog rows when actionable.
   - Backlog rows are PR-sized and tied to debt IDs or approved `maint:<kind>` IDs.
   - Summary counts in the module doc match the register.

8. LLM usability
   - The doc answers: what owns this behavior, what route/API touches it, what DB rows change, what authz applies, what can break, what legacy path remains, where to implement next.
   - Names are consistent across code, OpenAPI, wiki, tech debt, and backlog.
   - Stale predecessor pages are marked deprecated or retired.

9. Validation
   - `tally_check.sh <module>` passes or pre-existing failures are explicitly recorded.
   - Self-review confirms key anchors, cross-links, diagram/prose alignment, severity calls, and artifact consistency.

## Maturity Levels

Use these labels during audits:

| Level | Meaning |
|---|---|
| L0 stub | A page exists but cannot guide implementation. |
| L1 partial | Some useful facts exist, but no full trio/artifact set. |
| L2 living doc | Trio plus artifacts exist; main facts are usable. |
| L3 mature | Meets this standard and passes gates. |
| L4 current | L3 plus recently synced after the latest implementation. |

## Current Direction For MetalDocs

The backend/core module docs are already approaching L2/L3. Partial pages such as frontend-only stubs, predecessor pages, or modules without tech-debt/backlog/artifacts should not be patched as if they were mature. First promote them with `metaldocs-module-doc`, then keep them current with `metaldocs-module-doc-sync`.
