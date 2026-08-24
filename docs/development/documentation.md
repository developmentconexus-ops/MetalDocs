---
id: documentation-governance
kind: authority
owner: engineering
summary: MetalDocs-specific documentation placement and provenance-routing rules layered on the pinned organizational Repository Standard.
---

# Documentation governance

The exact organizational methodology pin and selection route are owned by `AGENTS.md`; repository organization/lifecycle is selected from the pinned `REPOSITORY-STANDARD.md`. This page contains only MetalDocs-specific specialization.

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

## Local metadata

MetalDocs uses minimal frontmatter only where it improves routing:

```yaml
id: unique-id
kind: authority | work
owner: owner-name
summary: one sentence
```

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

Temporary work follows the pinned Repository Standard. Independent review follows the pinned `ADVERSARIAL-REVIEW-METHOD.md`; MetalDocs adds only the repository-specific executable isolation/protection checks in `.github/workflows/ci.yml`.

The reset review predates the current review method and remains historical provenance; do not fabricate a second historical review branch.

## Source of current status

Only `docs/roadmap.md` owns mutable stage/gate/implementation status and exact next action. README, AGENTS, indexes, Product/Architecture pages, checkpoints, and PR descriptions must not become parallel mutable status authorities.
