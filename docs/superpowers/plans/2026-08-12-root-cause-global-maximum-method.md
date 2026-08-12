# Root-Cause / Global-Maximum Engineering Method — Executed Plan

**Status:** EXECUTED on PR #129  
**Accepted design:** `docs/superpowers/specs/2026-08-12-root-cause-global-maximum-method-design.md`  
**Durable outcome:** `wiki/standards/root-cause-global-maximum-method.md`

## Goal

Create one canonical root-cause/global-maximum engineering method, make live agent entrypoints consume it, and remove dead workflow routing without creating another framework.

## Global constraints applied

- simplify code/process, never correctness;
- root cause before patch;
- Global Maximum before optimizing a known workaround;
- YAGNI removes speculative/accidental complexity, not material enforcement;
- one authority for the doctrine;
- no restoration of retired `.agents/skills/` or `metaldocs-*` trees;
- no new verifier solely to enforce reading the method;
- historical/staging artifacts remain non-canonical.

## Executed work

### 1. Canonical standard

Promoted the accepted engineering doctrine into:

`wiki/standards/root-cause-global-maximum-method.md`

The standard owns:

- symptom/root-cause/target-property definitions;
- local maximum vs Global Maximum;
- essential vs accidental complexity;
- YAGNI boundary;
- enforcement hierarchy;
- legal decision outcomes;
- Engineering Decision Record;
- guard policy;
- transitional enforcement;
- bounded review/convergence;
- future cross-project promotion direction.

### 2. Agent entrypoints

Updated:

- `AGENTS.md` — model-agnostic mandatory gate + truthful routing;
- `CLAUDE.md` — MetalDocs facts/invariants + short bridge to the standard.

Dead local skill requirements were removed rather than restored.

### 3. Workflow-specific skills

Updated:

- `.claude/skills/developing-new-work/SKILL.md`;
- `.claude/skills/adversarial-review/SKILL.md`.

Both retain their workflow-specific mechanics and consume the canonical definitions instead of redefining them.

### 4. Live routing/documentation

Updated current entrypoints including:

- `wiki/ONBOARDING.md`;
- `wiki/architecture/system-map.md`;
- `wiki/references/ai-operating-system.md`;
- live frontend module routing pages;
- `wiki/concepts/design-workflow-audit.md`;
- `.claude/agents/frontend-screen-reviewer.md`.

Historical plans, old milestone evidence, and frozen artifacts were not mass-rewritten merely to make a grep result empty.

## Material findings during execution

### F1 — retired skill trees were intentionally removed

Repository history confirmed:

- `c7f06f2e` removed `.agents/skills/` during the pre-v1 re-baseline;
- `02ed1c24` removed retired `.claude/skills/metaldocs-*` trees.

**Disposition:** do not recreate them; repair live routing.

### F2 — planned helper paths were stale

Repository truth showed:

- `.claude/agents/wiki-curator.md` does not exist;
- `.claude/skills/gitnexus/SKILL.md` does not exist;
- impact analysis lives at `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md`.

**Disposition:** route to current real files only.

### F3 — canonical location violated documentation ownership

PR #129 review identified that the initial `docs/engineering/` target contradicted `wiki/standards/documentation-governance.md`.

**Root cause:** the implementation plan followed a generic documentation convention instead of the repository's ownership rule.

**Disposition:** move durable doctrine to `wiki/standards/root-cause-global-maximum-method.md`; remove the `docs/engineering/` copy; keep this plan/spec as staging history.

### F4 — live frontend screen reviewer depended on removed workflows

The reviewer agent still referenced retired skills/templates.

**Disposition:** preserve its read-only visual/numerical review role while routing it through live architecture, module docs, preview tools, and the canonical engineering standard.

## Verification policy

This executed plan intentionally contains no broad `git add <directory>` instructions and no dead-route grep that treats explanatory prose as a live dependency.

Final acceptance is based on the actual PR head:

1. GitHub diff review: only agent/workflow/documentation surfaces are changed;
2. referenced live paths verified against repository truth;
3. review threads dispositioned on the final head;
4. repository CI required gate green on the final head;
5. merge only after those conditions are met.

## Close-out

Detailed final evidence and review dispositions live in:

`docs/superpowers/reports/2026-08-12-root-cause-global-maximum-method-handoff.md`

The durable operating authority is the wiki standard, not this executed plan.
