# Documentation Governance

> **Last verified:** 2026-05-27
> **Scope:** Ownership model for durable project knowledge, section boundaries, and safe promotion rules.
> **Out of scope:** Rewriting every legacy document or mass-moving existing trees in one session.

## Ownership rule

- `wiki/` is the durable source of truth for maintained project knowledge.
- `docs/` is the staging and draft workspace for plans, specs, imported notes, prompts, templates, and non-canonical material that may later be promoted.
- Promotion goes by ownership and durability, not convenience. Durable material moves into the owning wiki domain folder. It does not get dumped at wiki root.

## Classification model

- `canonical wiki content` - maintained truth that should be the first-stop reference for the topic
- `staging/draft content` - design, planning, or research material that is still being shaped or not yet promoted
- `reference/archive content` - supporting evidence, handoffs, external-package notes, historical review packets, or archived context kept for re-entry
- `workflow/tooling gap` - missing index, governance rule, promotion path, or execution guidance that makes docs harder to trust
- `migration risk` - path, link, or ownership instability that makes a move unsafe without a phased pass
- `defer` - intentionally postponed cleanup or promotion

## Durable section boundaries inside `wiki/`

- `architecture/` - durable system structure, contracts, and boundary rules
- `modules/` - module-local living docs, debt registers, maturity evidence
- `database/` - schema ownership, dictionary, migration policy, reference data
- `standards/` - cross-cutting standards, including documentation governance
- `workflows/` - durable end-to-end flows
- `tests/` - repeatable acceptance and validation procedures
- `quality/` - QA operating-system home, release-quality rules, scenario-proof governance
- `decisions/` - durable ADRs
- `vision/` - durable product intent and target users
- `backlog/` - governed deferred work that remains active product memory
- `reviews/` - review packets and audit evidence
- `references/` - supporting operator references and historical aids that are useful but not primary canonical product truth

## What stays in `docs/`

- specs and plans under `docs/superpowers/`
- imported or exploratory research
- prompts and templates
- non-canonical runbooks that are phase-specific, transient, or still being normalized
- legacy ADR/history sets that have not been reconciled with the current wiki structure

`docs/` should not compete with `wiki/` for durable truth. If a `docs/` page becomes maintained truth, promote it into the owning wiki section and update indexes.

## Promotion rules

1. Confirm the content is durable enough to maintain.
2. Place it in the owning wiki domain folder.
3. Add it to the local folder index and root [../index.md](../index.md).
4. Leave a stable breadcrumb if the old path still has known consumers.
5. Avoid mass renames unless link impact is bounded and verified.

## Index rules

- `wiki/index.md` is the canonical root landing page.
- Each major domain folder should have its own canonical `index.md`.
- `README.md` inside wiki folders is optional and should exist only as a compatibility stub when older instructions or links still point there.
- Do not let any `README.md` evolve into a second full catalog once `index.md` exists.

## Current migration map

| Path | Classification | Decision |
|---|---|---|
| `wiki/references/ai-operating-system.md` | `reference/archive content` + `migration risk` + `compatibility bridge` | Keep path stable for now because `AGENTS.md`, `CLAUDE.md`, and handoff docs point here directly. It should explicitly bridge agents to the canonical QA loop in `wiki/quality/` until references are fully normalized. |
| `docs/superpowers/specs/2026-05-13-metaldocs-ai-operating-system-design.md` | `staging/draft content` | Keep in `docs/` as design input. It is not the live canonical wiki page. |
| `wiki/quality/deep-qa/index.md` | `canonical wiki content` | Canonical home for the promoted deep-QA artifact set. New durable links should point here. |
| `wiki/quality/deep-qa/runbook.md` | `canonical wiki content` | Canonical deep-QA execution runbook. |
| `wiki/quality/deep-qa/matrix.md` | `canonical wiki content` | Canonical deep-QA scenario matrix. |
| `wiki/quality/deep-qa/fixtures.md` | `canonical wiki content` | Canonical deep-QA fixture registry. |
| `wiki/quality/screen-qa-checklist.md` | `canonical wiki content` | Canonical reusable checklist for user-facing screen QA. |
| `wiki/quality/backend-api-qa-checklist.md` | `canonical wiki content` | Canonical reusable checklist for backend/API QA. |
| `wiki/quality/workflow-async-qa-checklist.md` | `canonical wiki content` | Canonical reusable checklist for async and workflow-owned QA. |
| `wiki/quality/release-closeout-checklist.md` | `canonical wiki content` | Canonical reusable checklist for final merge/release close-out. |
| `wiki/references/documents-approval-deep-qa/README.md` | `reference/archive content` + `compatibility breadcrumb` | Keep path stable because startup docs, module docs, and prompts still reference it directly. |
| `wiki/references/documents-approval-deep-qa/runbook.md` | `reference/archive content` + `compatibility breadcrumb` | Redirect-style compatibility page to the canonical QA location. |
| `wiki/references/documents-approval-deep-qa/matrix.md` | `reference/archive content` + `compatibility breadcrumb` | Redirect-style compatibility page to the canonical QA location. |
| `wiki/references/documents-approval-deep-qa/fixtures.md` | `reference/archive content` + `compatibility breadcrumb` | Redirect-style compatibility page to the canonical QA location. |
| `docs/superpowers/specs/2026-05-20-documents-approval-product-plus-qa-system-design.md` | `staging/draft content` | Keep in `docs/` as design/proposal input. Its durable operating outcome should land under `wiki/quality/`, not replace the draft spec. |
| `docs/runbooks/release-readiness.md` | `defer` | Candidate for promotion into `wiki/quality/` after normalization. Current file is still phase-specific and should not be moved blindly in the same session. |
| `docs/adr/*` | `reference/archive content` | Legacy ADR tree remains in `docs/` until reconciled against `wiki/decisions/`. Do not merge or rename blindly. |
| `docs/ck5-wiki/*` | `reference/archive content` | Keep in `docs/` as external/reference knowledge, not as top-level durable product truth. |

## Explicit defers

- Do not mass-rename `README.md` files to `index.md` where existing path consumers are unknown.
- Do not delete `wiki/references/documents-approval-deep-qa/` while startup docs, prompts, and module memory still point to it.
- Do not reconcile the legacy `docs/adr/` tree into `wiki/decisions/` without a dedicated review of overlap and historical status.
