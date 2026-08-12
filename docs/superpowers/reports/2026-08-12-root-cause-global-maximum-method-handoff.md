# Root-Cause / Global-Maximum Method — Implementation Handoff

**Date:** 2026-08-12  
**Base:** `main` at `8f4c7ac64d4a34eb5b7331cbe19ef4c0609a00a6`  
**Approved design:** `docs/superpowers/specs/2026-08-12-root-cause-global-maximum-method-design.md`  
**Implementation plan:** `docs/superpowers/plans/2026-08-12-root-cause-global-maximum-method.md`  
**Canonical method:** `wiki/standards/root-cause-global-maximum-method.md`

## Engineering Decision Record

### Symptom

Root-cause/global-maximum/YAGNI rules existed in several agent documents and skills, while `AGENTS.md` and live onboarding still routed to repository-local `metaldocs-*` skills that had been deliberately removed.

### Root cause

Engineering doctrine and workflow routing had multiple authorities. Pre-v1 re-baseline commits removed `.agents/skills/` and retired `metaldocs-*` skills, but live routing documents were not converged afterward.

### Target property

One canonical engineering doctrine defines root cause, local/global maximum, YAGNI, enforcement hierarchy, transitional design, and review convergence. Live entrypoints route to existing files only and project-specific skills consume rather than redefine the doctrine.

### Authority and boundary

- Durable generic engineering doctrine: `wiki/standards/root-cause-global-maximum-method.md`.
- Model-agnostic routing: `AGENTS.md`.
- MetalDocs invariants: `CLAUDE.md` and governed architecture/wiki.
- New-work workflow: `.claude/skills/developing-new-work/SKILL.md`.
- Review workflow: `.claude/skills/adversarial-review/SKILL.md`.
- `docs/superpowers/specs/**`, plans and reports remain staging/history, not durable canonical truth.

### Local-maximum candidate

Copy the new rules into `AGENTS.md`, `CLAUDE.md`, and each skill independently or restore the deleted `.agents/skills/` tree.

### Global-maximum candidate

One durable canonical standard in the governed wiki plus thin contextual bridges. Do not restore removed workflow trees.

### Decision

**Restructure now.** Canonicalize doctrine in `wiki/standards/` and correct live routing.

### Enforcement

Documentation/agent-routing is the appropriate layer for an agent decision method. No custom verifier was added merely to prove agents read it.

### Proof

- `wiki/standards/documentation-governance.md` explicitly assigns durable maintained truth to `wiki/` and cross-cutting standards to `wiki/standards/`.
- GitHub compare confirms the branch changes only agent/docs/wiki/method surfaces.
- Every new primary skill path was checked against repository truth.
- Pre-v1 removal history was verified: `c7f06f2e` removed `.agents/skills/`; `02ed1c24` removed retired `.claude/skills/metaldocs-*` trees.
- GitHub PR CI on the final head is the executable repository-level acceptance authority.

### Transitional exit

N/A. Future promotion of the project-neutral core to MNFS/harness is deliberate future work, not required for MetalDocs correctness.

## Implementation findings

### F1 — deleted skill routing

`.agents/skills/` and the retired `metaldocs-*` skill trees were deliberately removed during the pre-v1 re-baseline. The fix is truthful routing, not restoration of dead trees.

### F2 — planned bridge assumptions were stale

Two implementation-plan assumptions were disproved by repository truth:

1. `.claude/agents/wiki-curator.md` does not exist; documentation routing points directly to `wiki/standards/documentation-governance.md`.
2. `.claude/skills/gitnexus/SKILL.md` does not exist; impact analysis lives at `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md`.

### F3 — canonical path violated documentation ownership

The approved design/initial implementation target placed the method under `docs/engineering/`. PR #129 review found this conflicts with the repository's ownership rule: `wiki/` is durable truth and `docs/` is staging/draft material.

The finding was verified against `wiki/standards/documentation-governance.md`. The durable method was moved to `wiki/standards/root-cause-global-maximum-method.md`; the `docs/engineering/` copy is removed. The design spec and implementation plan remain in `docs/superpowers/` as the historical staging artifacts that led to the final ownership ruling.

### F4 — live screen reviewer depended on retired workflows

The live `frontend-screen-reviewer` depended on retired frontend/screen skills and a removed template. It was simplified to use the canonical standard, `AGENTS.md`, `CLAUDE.md`, frontend architecture, design audit, module docs, and actual preview tooling while preserving its read-only visual/numerical review role.

## Deliberate scope boundary

Historical plans, milestone evidence, and frozen sync artifacts retain old paths as historical evidence. This change does not rewrite history merely to make a repository-wide grep zero.

`wiki/architecture/frontend-structure.md` still contains legacy prose referring to retired frontend/TanStack skills. Authoritative routing is now `AGENTS.md`, onboarding, system map, and live workflow files. Rewriting that large architecture page solely for those routing mentions is intentionally outside this surgical change unless a material reviewer finding demonstrates that those references remain an active correctness hazard.
