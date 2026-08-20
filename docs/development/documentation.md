---
id: documentation-governance
kind: authority
owner: engineering
summary: MetalDocs-specific documentation placement, provenance-routing, and checkpoint rules layered on Repository Standard v1.0.0.
---

# Documentation governance

Canonical repository organization/lifecycle is defined by `developmentconexus-ops/conexus-methodology/REPOSITORY-STANDARD.md` v1.0.0. This page contains only MetalDocs-specific specialization.

## Local semantic layout

MetalDocs currently has named consumers for:

```text
docs/product/
docs/architecture/
docs/decisions/
docs/development/
docs/reference/
docs/work/            temporary branch-only work only
```

Do not add another documentation root, live archive/tombstone tree, or new category without a real consumer.

Durable paths use lowercase kebab-case semantic names. Existing imported Product/R10 authorities may retain their historical internal title/provenance blocks until substantively rewritten; cosmetic normalization must not risk decision loss.

## Local metadata deviation — checkpoint

Repository Standard permits local semantic surfaces when a real consumer exists. MetalDocs uses minimal frontmatter:

```yaml
id: unique-id
kind: authority | checkpoint | work
owner: owner-name
summary: one sentence
```

`checkpoint` is a MetalDocs-specific durable, non-authoritative accepted-work snapshot. Current consumer: `docs/reference/t8e-checkpoint.md`, which preserves already-accepted T8-E design across repository-governance gates without promoting it to final authority.

Removal trigger: when the owning stage is ratified and its durable authority fully absorbs the checkpoint.

`work` remains temporary and branch-only under `docs/work/`; it never enters a merge candidate or `main`.

## Carried pre-reset provenance strings

Some ratified authorities imported by the clean-slate reset still contain historical `wiki/...` strings. They are non-navigational provenance, even when old prose calls them current/program/decision authority.

Current routing is only through `docs/index.md`, `docs/roadmap.md`, and `docs/decisions/index.md`. Important replacements include:

```text
old technical-architecture router → docs/roadmap.md
old post-T6 stage program          → docs/roadmap.md
old decision registry              → current semantic authorities + docs/decisions/forward-obligations.md
```

Do not recursively repair provenance-only strings unless a live consumer depends on them.

## Temporary work and review

Temporary work follows Repository Standard v1. Future independent Fable review uses an isolated `review/<gate>-fable` branch whose only delta from the candidate is `docs/work/current/ai-dialog.md`; the candidate branch and `main` never absorb that review artifact.

The reset review predates this standard and remains historical provenance; do not fabricate a second historical review branch.

## Source of current status

Only `docs/roadmap.md` owns mutable stage/gate/implementation status and exact next action. README, AGENTS, indexes, Product/Architecture pages, checkpoints, and PR descriptions must not become parallel mutable status authorities.