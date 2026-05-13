---
name: metaldocs-module-doc
description: "Produces mature MetalDocs module wiki memory for one module per session: Arc42/C4 doc, API route truth table, runtime flows, public surface, persistence map, cross-dependencies, ADR links, tech-debt register, refactor backlog, artifacts, and validation gates. Use when the user asks to document, mature, promote, deep-dive, map, rebuild, or make a real technical wiki page for a module, especially when a wiki/modules module page is a stub/partial page or lacks the doc trio/artifacts. Use metaldocs-module-doc-sync instead for bounded post-implementation updates to an already mature module."
---

# MetalDocs Module Documentation

Build one canonical, mature technical wiki entry per module: read by humans, used as memory by LLM peers, and kept fresh by `metaldocs-module-doc-sync`.

## Maturity Target

A mature module wiki is not just architecture prose. It is an implementation memory packet that lets a future agent answer:
- What owns this behavior?
- Which route/API/spec/codegen surface touches it?
- Which service/repository/database rows change?
- Which authz, tenant, audit, idempotency, and error-envelope rules apply?
- What legacy paths, spec gaps, and debt shape the next implementation?

Read `references/module-wiki-maturity.md` during Phase 0. Use that standard as the acceptance target.

## Maturity Levels

Use these labels in `00-context.md`, `06-selfreview.md`, and the final report:

| Level | Meaning |
|---|---|
| L0 stub | A page exists but cannot guide implementation. |
| L1 partial | Useful facts exist, but no full trio/artifact set. |
| L2 living doc | Trio plus artifacts exist; main facts are usable. |
| L3 mature | Meets the maturity standard and passes gates. |
| L4 current | L3 plus synced after the latest implementation. |

## One Module Per Run

If the user asks to mature multiple modules, do one module at a time. Module docs are dense; mixing modules pollutes context and lowers quality. For a whole-module-wiki maturity program, first audit all modules, then promote or refresh them one by one.

## Output Contract

For module `M`, produce:
- `wiki/modules/M.md` - Arc42/C4 living doc with key files, API route truth table, runtime flows, persistence, authz, dependencies, decisions, risks, glossary, cross-links, and changelog.
- `wiki/modules/M-tech-debt.md` - evidence-backed gap register, debt only, no fix prescriptions.
- `wiki/backlog/M-refactor.md` - PR-sized refactor rows tied to debt IDs or approved `maint:<kind>` IDs.
- `wiki/modules/M/_artifacts/00-context.md` - context, scope, existing maturity level, related docs, and open questions.
- `wiki/modules/M/_artifacts/01-surface.md` - public exports, route owners, generated API surface, jobs, and file tree.
- `wiki/modules/M/_artifacts/02-flow-*.md` - representative read/write/state-transition traces.
- `wiki/modules/M/_artifacts/03-deps.md` - inbound/outbound imports, composition-root wiring, config/env edges, and cross-module calls.
- `wiki/modules/M/_artifacts/04-persistence.md` - tables, migrations, constraints, indexes, triggers, tenant keys, tripwire/GUC behavior, retention/deletion facts.
- `wiki/modules/M/_artifacts/05-industry.md` - comparison only against approved patterns in `references/industry-patterns-index.md`.
- `wiki/modules/M/_artifacts/06-selfreview.md` - final maturity and quality review.
- `wiki/modules/M/_artifacts/sync-log.md` - append-only sync log header for future incremental updates.
- Index/cross-link updates in `wiki/README.md`, `wiki/modules/README.md`, and directly related module/concept/architecture/decision docs.

After this skill promotes a module to L3, use `metaldocs-module-doc-sync` for cheap post-implementation refreshes.

## Research Split

Subagents collect facts only:
- AST/file/symbol scans.
- HTTP route and generated API inventory.
- Route-by-route data flow.
- Import graph and composition-root wiring.
- SQL/migration/persistence map.

Main agent performs judgment:
- Module boundary and purpose.
- Which facts matter for implementation memory.
- Severity ratings from the tech-debt rubric.
- ADR/missing-ADR classification.
- Industry pattern applicability.
- Final composition and self-review.

Reject or rewrite any subagent artifact that contains prescriptive prose such as "should", "recommend", "consider", or unsupported "industry standard" claims.

## Workflow

| Phase | Owner | Artifact | Purpose |
|---|---|---|---|
| 0 - Load context | Main | `00-context.md` | Read existing wiki, ADRs, concepts, maturity standard; classify L0-L4. |
| 1 - Surface scan | Research subagent | `01-surface.md` | Exports, routes, codegen, jobs, file tree, key tests. |
| 2 - Data-flow traces | Research subagents | `02-flow-*.md` | End-to-end read/write/state-transition behavior. |
| 3 - Cross-deps | Research subagent | `03-deps.md` | IN/OUT imports, callers, DI/composition wiring, env/config edges. |
| 4 - Persistence map | Research subagent | `04-persistence.md` | Tables, migrations, constraints, tripwire, tenant keys. |
| 5 - Industry comparison | Main | `05-industry.md` | Use only approved pattern index unless user approves adding a new source. |
| 6 - Compose | Main | module doc, register, backlog, sync-log header | Build the mature wiki entry. |
| 6.5 - Tally | Main/script | stdout | Run `scripts/tally_check.sh M`; fix mismatches. |
| 6.75 - Self-review | Main | `06-selfreview.md` | Verify maturity, anchors, links, counts, diagrams, severity, artifacts. |
| 7 - Wiki integration | Main | indexes/cross-links | Thread the module into the wiki graph. |
| 8 - Commit/report | Main | final report | Report maturity level and gate results; commit only when asked. |

## Run Sequence

1. Confirm exactly one module.
2. Create or reuse `wiki/modules/M/_artifacts/`.
3. Read `references/module-wiki-maturity.md`, `references/industry-patterns-index.md`, existing `wiki/modules/M*`, related ADRs/concepts/architecture docs, and related modules.
4. Write `00-context.md` with current maturity level, target level, source files, related docs, and any justified skips.
5. Generate or refresh research artifacts using the templates in `templates/`.
6. Compose the three canonical wiki files and sync-log header.
7. Apply the coverage gate.
8. Run the tally gate. On Windows, Git Bash may be required:
   `& 'C:\Program Files\Git\bin\bash.exe' .claude/skills/metaldocs-module-doc/scripts/tally_check.sh M`
9. Write `06-selfreview.md`. If it finds issues, fix them, rerun tally, and update self-review.
10. Update indexes and direct cross-links.
11. End with a concise report: initial/final maturity level, files changed, gates, and deferred gaps.

## Coverage Gate

Before publish, `wiki/modules/M.md` must satisfy:

- **(a) Public surface:** every exported/public symbol in `01-surface.md` is named, grouped, or explicitly excluded as undocumented-on-purpose.
- **(b) Routes/API:** every HTTP operation has runtime owner, handler, authz, spec path, operationId, codegen method, status, and notes in the HTTP operations table and API Route Truth Table.
- **(c) Runtime behavior:** representative read/write/state-transition flows are traced; state machines, transactions, idempotency, audit, error envelopes, and authz layers are documented when applicable.
- **(d) Persistence:** every owned/read table, migration family, trigger, index, tenant key, and tripwire/GUC fact in `04-persistence.md` is represented or explicitly n/a.
- **(e) Cross-deps:** every IN-edge and OUT-edge in `03-deps.md` appears in C4/prose/cross-links or is explicitly excluded with reason.
- **(f) Decisions:** every load-bearing decision links an ADR or appears as `missing-ADR` debt.
- **(g) Debt/backlog:** every actionable T-row has a backlog row, except explicitly latent Minor debt.
- **(h) LLM usability:** the doc answers where to implement, what can break, what is legacy/partial/deprecated, and which debt/backlog rows govern the next PR.

Do not lower the gate to make a page pass. Record justified n/a cases in `00-context.md` and `06-selfreview.md`.

## Self-Review Checklist

Write `06-selfreview.md` with one short answer per item:

1. Initial and final maturity level.
2. Severity rubric application for every Critical and Major row.
3. Mermaid diagram/prose alignment.
4. Top-3 debt ordered by severity, then blast radius.
5. Cross-link existence.
6. Key file anchor freshness, with at least three sampled anchors.
7. Backlog/debt linkage.
8. Route/API truth table completeness.
9. Persistence/tripwire/tenant-key completeness.
10. Industry citations trace to `references/industry-patterns-index.md`.
11. Subagent purity: facts only, no prescriptive prose.
12. Remaining gaps preventing L4 current status.

## Mature Wiki Guard

Do not call a module mature because it has a long page. Mature means the doc is useful for implementation without fresh archaeology:
- It maps code, runtime behavior, API contract, data ownership, cross-deps, decisions, debt, and backlog.
- It records what is legacy, partial, deprecated, or intentionally missing.
- It can be mechanically checked by tally and manually checked by self-review.
- It is ready for cheap maintenance by `metaldocs-module-doc-sync`.

If a module is frontend-only or non-Go, adapt the same maturity target: component surface replaces exported Go symbols, UI workflows replace HTTP flows, and state/store/contracts replace persistence where applicable. Do not force irrelevant backend sections; record n/a with evidence.

## Industry Guard

Use only patterns in `references/industry-patterns-index.md` by default. Add new industry sources only with explicit user approval, and then add the source, access date, quote snippet, and applicability note to the index in the same change.

## Anti-Patterns

- Documenting multiple modules in one run.
- Producing prose without source artifacts.
- Skipping route/API truth for modules with HTTP surfaces.
- Treating OpenAPI, runtime routes, and generated code as interchangeable without checking all three.
- Inventing debt without concrete evidence.
- Writing fixes in the tech-debt register instead of the backlog.
- Claiming ADR coverage without a real `wiki/decisions/*` link.
- Bumping `Last verified` without checking anchors.
- Calling a partial/stub page mature.

## Output

End-of-run report includes:
- Module name.
- Initial and final maturity level.
- Artifacts written.
- Wiki files written.
- Coverage gate results.
- Tally result.
- Industry references used.
- Deferred gaps and why they remain.

## Changelog

- 1.3 (2026-05-13) - Defined module wiki maturity as implementation memory for humans and LLMs. Added `references/module-wiki-maturity.md`; expanded output contract with API route truth table, sync log, indexes, persistence/cross-dep/runtime gates; changed target from long-form architecture doc to L3 mature module wiki.
- 1.2 (2026-05-10) - Moved cross-deps phase to the fastest suitable research path; preserved Codex for AST/code-heavy phases.
- 1.1 (2026-05-10) - Added mechanical tally and self-review gates.
- 1.0 (2026-05-10) - Initial Arc42/C4 module documentation workflow.
