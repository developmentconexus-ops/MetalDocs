---
name: metaldocs-module-doc
description: Produces senior-grade living architecture documentation (Arc42 + C4 + ADR links + tech-debt register + refactor backlog) for one MetalDocs module per session. Use this skill whenever the user says any of "document module X", "write architecture doc for X", "map module X", "deep-dive X", "we need a real doc for X", or asks to add depth to any `wiki/modules/<m>.md` stub — even when the user does not say the word "skill". Dispatches Codex subagents for cheap research (AST scan, data-flow trace, import graph, persistence map). Main agent owns interpretation, gap analysis, industry comparison, and final composition. Enforces a mechanical tally gate (`scripts/tally_check.sh`) and a self-review pass before publish so severity counts and ADR links cannot drift between the composed doc and the register.
---

# MetalDocs Module Documentation

Build one canonical, living architecture doc per module — read by humans, parsed by LLM peers, kept fresh by `wiki-curator`.

## Why this skill is gated

Module docs that drift from code are worse than no docs — readers trust them, get burned, then stop trusting any doc. The phase gates exist so each step leaves an auditable artifact a future reviewer (human or LLM) can grep. The composer is fluent in English; English is great at hiding mismatched numbers. The gates put the mismatches on the floor before publish:

- Each phase deposits a numbered artifact under `wiki/modules/<module>/_artifacts/`. Missing artifact ⇒ phase not done.
- Coverage proof runs before publish — five mechanical checks (a–e) plus a self-review pass.
- Wiki integration runs LAST so cross-links cannot rot before the doc is final.

Skip a gate when the work genuinely doesn't need it (e.g. no state machine in the module ⇒ §6 row is "n/a"), but record the skip in `00-context.md`. Silent skips are how drift starts.

## Auto-invoke triggers

Use this skill when the user says any of:
- "Document module X" · "Write architecture doc for X" · "Map module X"
- "We need a real doc for X" · "Deep-dive X"
- Any prompt referencing `wiki/modules/<m>.md` and asking for more depth than a stub

If the user asks to document MULTIPLE modules in one go: refuse politely, do one at a time. Context pollution kills doc quality.

## Output contract

For module `M`, produce:
- `wiki/modules/M.md` — Arc42-shaped doc with embedded C4 Mermaid diagrams (THIS is the deliverable)
- `wiki/modules/M-tech-debt.md` — gap register (debt items only, no fix prescriptions)
- `wiki/backlog/M-refactor.md` — actionable refactor rows (one row = one PR)
- Update `wiki/README.md` index (Last verified stamps + new anchors)
- Cross-links from related concept/decision docs

Living doc. `wiki-curator` keeps it from rotting after merges.

## Subagent / main agent split

**Subagents (Codex via `codex:codex-rescue`) do research, NOT judgment.**
- AST scans · file enumeration · import graph extraction · raw SQL/migration list · route-by-route data flow
- Output = facts only (file:line, signatures, table names, call edges). No "should/recommend".
- Cap each subagent prompt at ~800 tokens. Point to templates by path, do not inline.

**Main agent (Opus-quality reasoning) does interpretation.**
- Picks what each fact MEANS for the module
- Decides which industry patterns are comparable (gated by `references/industry-patterns-index.md`)
- Writes Quality Goals, Risks, ADR amendments, gap callouts
- Composes the final doc and signs off

Rule: if a subagent ever writes a sentence containing "should", "recommend", "consider", or "professional", reject the artifact and re-dispatch with a stricter prompt.

## 8-phase workflow

| Phase | Owner | Inputs | Artifact | Notes |
|---|---|---|---|---|
| 0 — Load context | Main | module name | `_artifacts/00-context.md` | Read `wiki/README.md`, existing module stub, related ADRs, related concept docs |
| 1 — Surface scan | Codex subagent | module root path | `_artifacts/01-surface.md` | public exports, HTTP ops, file tree, migration list |
| 2 — Data-flow trace | 2–3 Codex subagents (parallel) | top operations from §1 | `_artifacts/02-flow-<op>.md` × N | HTTP → handler → service → repo → DB end-to-end |
| 3 — Cross-deps | Sonnet 4.6 subagent (`general-purpose`, `model: "sonnet"`) | module root | `_artifacts/03-deps.md` | imports IN / imports OUT, callers map. Sonnet, not Codex — Codex run on auth took ~3× longer than other phases because the IN-edge scan + 14 config-field trace is grep-bound, not AST-bound; Sonnet via Grep tool is faster. |
| 4 — Persistence map | Codex subagent | module root + migrations dir | `_artifacts/04-persistence.md` | tables, FKs, triggers, tripwire pairing, GUC reads |
| 5 — Industry comparison | Main + context7 + WebSearch | §1–§4 artifacts + `references/industry-patterns-index.md` | `_artifacts/05-industry.md` | gated: only patterns from the index, or explicit user opt-in |
| 6 — Compose | Main | all artifacts | `wiki/modules/M.md` + `M-tech-debt.md` + `backlog/M-refactor.md` | Arc42 + C4 Mermaid; coverage gate enforced |
| 6.5 — Mechanical tally | Main (runs `scripts/tally_check.sh`) | composed wiki files | stdout `[tally] PASS` | severity counts, missing-ADR count, debt↔backlog linkage |
| 6.75 — Self-review | Main | composed wiki files + artifacts | `_artifacts/06-selfreview.md` | catches what tally cannot: severity judgment, mermaid box drift, rubric application |
| 7 — Wiki integration | `wiki-curator` subagent | new doc paths | updated `wiki/README.md` + cross-links | bumps Last verified, adds anchors |
| 8 — Commit | Main | all wiki changes | git commit | conventional message, no Co-Authored-By unless user asks |

## Run sequence

1. Confirm one module. If user names two, ask which first.
2. `mkdir wiki/modules/<m>/_artifacts/`. (Use plain folder, not hidden.)
3. **Phase 0** — Main reads existing wiki + ADRs. Drop summary into `00-context.md`. List open questions to the user (one batch, then proceed).
4. **Phase 1** — Dispatch `codex:codex-rescue` with `templates/subagent-surface-scan.md`. Provide module path. Receive `01-surface.md`.
5. **Phase 2** — From §1, pick the 2–3 most representative operations (one read, one write, one state-transition if it exists). Dispatch them in parallel (one Codex subagent per operation, all in a single message). Each uses `templates/subagent-data-flow-trace.md`.
6. **Phase 3** — Dispatch `general-purpose` subagent with `model: "sonnet"` and `templates/subagent-cross-deps.md`. Receive `03-deps.md`. (Sonnet 4.6, not Codex — see Phase-3 row in workflow table for rationale.)
7. **Phase 4** — Dispatch Codex with `templates/subagent-persistence-map.md`. Receive `04-persistence.md`.
8. **Phase 5** — Main does industry comparison. ONLY patterns named in `references/industry-patterns-index.md` are admissible by default. If a fresh comparison is genuinely warranted, ask the user once; on yes, use `context7` for current docs and WebSearch for source-of-truth, then add the pattern to the index in the same commit.
9. **Phase 6** — Main composes `wiki/modules/M.md` from `templates/module-doc.md`. Coverage gate (see below) must pass before continuing.
10. **Phase 6.5 — Mechanical tally.** Run `bash .claude/skills/metaldocs-module-doc/scripts/tally_check.sh <module>`. The script grep-counts severities, missing-ADR references, and validates debt↔backlog linkage. On `FAIL`: fix the composed doc, re-run until `PASS`. Do NOT proceed without `[tally] PASS` printed.
11. **Phase 6.75 — Self-review pass.** Main agent re-reads the composed `M.md` against the artifacts with fresh eyes and writes `_artifacts/06-selfreview.md` answering the checklist below. The point is to catch what the mechanical tally cannot: severity calls, mermaid boxes that name things the prose never explains, and rubric slack. If self-review surfaces a fix, apply it, re-run 6.5, then re-run 6.75 once more on the final state.
12. **Phase 7** — Dispatch `wiki-curator` to thread the new doc into the wiki graph.
13. **Phase 8** — Single commit; conventional message `docs(module-<m>): architecture documentation`.

## Coverage gate (Phase 6 → Phase 6.5)

Before mechanical tally can start, the composed `wiki/modules/M.md` must satisfy:

- **(a) Public surface** — every exported type/function in `01-surface.md` is named at least once in the doc (or explicitly listed in tech-debt as undocumented-on-purpose)
- **(b) Routes** — every HTTP operation in `01-surface.md` appears in the C4 Container or Component diagram
- **(c) Cross-deps** — every IN-edge and OUT-edge in `03-deps.md` is in §5 (Building Blocks) or §8 (Cross-cutting)
- **(d) State transitions** — every state machine surfaced in `02-flow-*.md` has a table in §6 (Runtime View)
- **(e) Decisions** — every architectural decision either links an existing ADR (`wiki/decisions/0xxx-*.md`) or is logged in `M-tech-debt.md` as missing-ADR

Grep the composed doc against the artifacts. If a coverage check fails, fix the doc — do not lower the gate. Lowering the gate is the slow-motion version of writing a vibes doc.

## Phase 6.75 — Self-review checklist

Write `_artifacts/06-selfreview.md` answering each item. One short sentence per row; "n/a — <reason>" is allowed but must say why.

1. **Severity rubric application.** For every Critical or Major row in `M-tech-debt.md`, does the rubric in `templates/tech-debt-register.md` actually map to that severity? If a row claims Major but the trigger list says Critical (e.g. regulated audit-trail gap, multi-tenant data leak), re-rate it now.
2. **Mermaid box ↔ prose.** Every box in the §3 and §5 diagrams is named at least once in the surrounding prose. Stray boxes get either an explanation or removal.
3. **Top-3 in §11.** The "Top 3" list is by severity, then by blast-radius — not by order of authorship.
4. **Cross-link existence.** Every wiki link in the doc points at a file that exists today (`ls` it, don't guess).
5. **Key Files freshness.** Every `path:LL` anchor opens to the symbol it claims. Sample at least 3.
6. **Backlog ↔ debt linkage.** Every `T-NNN` has a matching backlog row OR is explicitly "no backlog row yet — latent" (acceptable for Minor only).
7. **Industry citations.** Every §5 industry-comparison citation traces back to a row in `references/industry-patterns-index.md` or to a new row added in the same commit.
8. **Subagent purity.** Re-skim `_artifacts/02-flow-*.md`, `03-deps.md`, `04-persistence.md`. If any contains "should / recommend / professional / industry-standard", that prose came from a subagent that broke the research-only rule; flag and strip.

## Subagent dispatch patterns

**Phases 1, 2, 4 — Codex (`codex:codex-rescue`):**

```
Agent({
  subagent_type: "codex:codex-rescue",
  description: "<phase>: <module>",
  prompt: "[paste subagent template from .claude/skills/metaldocs-module-doc/templates/subagent-<X>.md]\n\nModule path: internal/modules/<m>\nArtifact path: wiki/modules/<m>/_artifacts/<NN>-<X>.md\nModel: --model gpt-5.3-codex"
})
```

**Phase 3 — Sonnet 4.6 (`general-purpose`):**

```
Agent({
  subagent_type: "general-purpose",
  model: "sonnet",
  description: "phase 3 cross-deps: <module>",
  prompt: "[paste subagent template from .claude/skills/metaldocs-module-doc/templates/subagent-cross-deps.md]\n\nModule path: internal/modules/<m>\nArtifact path: wiki/modules/<m>/_artifacts/03-deps.md"
})
```

Rationale: cross-deps work is grep-bound (IN-edge scan across whole repo, config-var trace to env-var sites, DI wiring across composition roots). Sonnet via the Grep tool finishes in a fraction of the Codex wall-clock time. Other phases stay on Codex because they benefit from its AST/code-execution sandbox.

Subagent never edits the wiki doc directly. Only writes its artifact file.

## Industry comparison guard

The cheapest filler is "Stripe does X" with no link, no version, no quote. To block that:

1. `references/industry-patterns-index.md` lists pre-vetted patterns with source URL, version pinned, and quote snippet. Use those.
2. Any new pattern requires: source URL · accessed date · quoted snippet · why it applies HERE (one sentence anchored to a MetalDocs file:line).
3. If the user opts to add a new pattern mid-session, append the row to the index in the same commit.

No vibes-based "industry standard" lines. Either it is in the index, or it is footnoted with source.

## Anti-patterns (instant rewrite)

- Documenting two modules in one session.
- Subagent writing prose with "should/recommend".
- Skipping Phase 0 (you will rediscover stale wiki context the slow way).
- "Industry standard" sentences without source link in §5.
- ADR claim without `wiki/decisions/0xxx-*.md` link.
- Mermaid diagrams that name boxes the doc never explains.
- Tech-debt items written as fixes ("we should refactor X") — they go in `backlog/M-refactor.md` instead.
- File:line anchors not verified (Last verified stamp not bumped).
- Commit that merges doc + code changes. Doc commit is doc-only.

## Red flags — STOP

| Thought | Reality |
|---|---|
| "I already know this module, skip Phase 1" | Codex AST scan catches what you forgot. Run it. |
| "I'll add ‘industry standard says…' without a link" | Index or footnote. Or delete the sentence. |
| "Coverage gate is overkill" | Gate is the only thing separating this skill from a vibes doc. |
| "Skill prompt is long, drop the artifact list" | Artifact list IS the deliverable shape. Keep. |
| "Two modules at once is faster" | Token bleed kills the second one. One at a time. |
| "Subagent prompt can be looser" | Loose subagent = prose artifact = main agent rewrites it = double cost. |

## Output expectations

End-of-run report covers: module name · 5 artifact paths · final 3 wiki paths · coverage gate results (a–e pass/fail) · industry references used (with URLs) · open questions deferred to tech-debt.

## Changelog

- 1.2 (2026-05-10) — After auth first-run review: moved Phase 3 (cross-deps) off Codex onto Sonnet 4.6 (`general-purpose`, `model: "sonnet"`). Codex took ~3× longer than other phases on auth because cross-deps work is grep-bound (whole-repo IN-edge scan + 14 config-field env-var trace + DI wiring across composition roots), not AST-bound. Sonnet via the Grep tool is the fast path. Updated workflow table, run sequence, dispatch-pattern section, and `templates/subagent-cross-deps.md`. Other phases stay on Codex.
- 1.1 (2026-05-10) — Hardening pass after iam first-run review. Added Phase 6.5 mechanical tally (`scripts/tally_check.sh`) to catch severity-count drift and ADR-count mismatches. Added Phase 6.75 self-review checklist to catch severity-judgment slack and stray mermaid boxes. Tightened severity rubric in tech-debt template with concrete triggers (regulated audit-trail gap → Critical, etc.). Loosened refactor-backlog `debt_id` schema with `maint:<kind>` enum for maintenance rows that have no debt origin. Softened opening "Iron Law" framing with rationale (why gates exist, when skipping is OK if recorded).
- 1.0 (2026-05-10) — initial release. Arc42 + C4 + ADR scaffold; subagent/main split; coverage gate; industry-patterns-index guard.
