# Root-Cause / Global-Maximum Method — Implementation Handoff

**Date:** 2026-08-12  
**Base:** `main` at `8f4c7ac64d4a34eb5b7331cbe19ef4c0609a00a6`  
**Approved design:** `docs/superpowers/specs/2026-08-12-root-cause-global-maximum-method-design.md`  
**Canonical method:** `docs/engineering/root-cause-global-maximum-method.md`

## Engineering Decision Record

### Symptom

Root-cause/global-maximum/YAGNI rules existed in several agent documents and skills, while `AGENTS.md` and live onboarding still routed to repository-local `metaldocs-*` skills that had been deliberately removed.

### Root cause

Engineering doctrine and workflow routing had multiple authorities. Pre-v1 re-baseline commits removed `.agents/skills/` and retired `metaldocs-*` skills, but live routing documents were not converged afterward.

### Target property

One canonical engineering doctrine defines root cause, local/global maximum, YAGNI, enforcement hierarchy, transitional design, and review convergence. Live entrypoints route to existing files only and project-specific skills consume rather than redefine the doctrine.

### Authority and boundary

- Generic engineering doctrine: `docs/engineering/root-cause-global-maximum-method.md`.
- Model-agnostic routing: `AGENTS.md`.
- MetalDocs invariants: `CLAUDE.md` and governed architecture/wiki.
- New-work workflow: `.claude/skills/developing-new-work/SKILL.md`.
- Review workflow: `.claude/skills/adversarial-review/SKILL.md`.

### Local-maximum candidate

Copy the new rules into `AGENTS.md`, `CLAUDE.md`, and each skill independently or restore the deleted `.agents/skills/` tree.

### Global-maximum candidate

One canonical method plus thin contextual bridges. Do not restore removed workflow trees.

### Decision

**Restructure now.** Canonicalize doctrine and correct live routing.

### Enforcement

Documentation/agent-routing is the appropriate layer for an agent decision method. No custom verifier was added merely to prove agents read it.

### Proof

- GitHub compare confirms the branch modifies only agent/docs/method surfaces.
- Every new primary skill path was checked against repository truth.
- Pre-v1 removal history was verified: `c7f06f2e` removed `.agents/skills/`; `02ed1c24` removed retired `.claude/skills/metaldocs-*` trees.
- GitHub PR CI is the executable repository-level acceptance authority for the final head.

### Transitional exit

N/A. Future promotion of the project-neutral core to MNFS/harness is deliberate future work, not required for MetalDocs correctness.

## Implementation notes

During execution, two planned routing assumptions were disproved by repository truth:

1. `.claude/agents/wiki-curator.md` does not exist; documentation routing therefore points directly to `wiki/standards/documentation-governance.md`.
2. `.claude/skills/gitnexus/SKILL.md` does not exist; the current impact-analysis skill is `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md`.

The live `frontend-screen-reviewer` agent also depended on retired screen/frontend skills and a removed template. It was simplified to depend on the canonical method, `AGENTS.md`, `CLAUDE.md`, frontend architecture, design audit, and actual preview tooling while preserving its read-only visual/numerical review role.

## Deliberate scope boundary

Historical plans, milestone evidence, and frozen sync artifacts retain old paths as historical evidence. This change does not rewrite history merely to make a repository-wide grep zero.

`wiki/architecture/frontend-structure.md` still contains legacy prose referring to the retired frontend/TanStack skills. The authoritative routing now lives in `AGENTS.md`, onboarding, system map, and the live reviewer agent. Rewriting that large architecture page solely for routing text was not required to establish the single engineering-method authority and was intentionally left outside this surgical change.
