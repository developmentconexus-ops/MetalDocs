# Root-Cause / Global-Maximum Engineering Method — Design Record

**Status:** ACCEPTED by operator on 2026-08-12  
**Classification:** staging/design record; not canonical maintained truth  
**Durable outcome:** `wiki/standards/root-cause-global-maximum-method.md`

## Problem

MetalDocs had root-cause, Global Maximum, YAGNI, review-convergence, and agent-routing rules distributed across `CLAUDE.md`, `AGENTS.md`, skills, and live onboarding. Some entrypoints also referenced repository-local skills that had already been removed.

That created two defect classes:

1. **multiple authorities** for the same engineering doctrine;
2. **dead workflow routing** that a fresh agent could follow into non-existent files.

## Accepted target

The operator accepted this design:

- one canonical engineering doctrine;
- `AGENTS.md` is a short model-agnostic routing bridge, not a second doctrine;
- `CLAUDE.md` owns MetalDocs-specific facts and invariants and bridges to the canonical doctrine;
- `developing-new-work` applies the doctrine before new feature/module design;
- `adversarial-review` applies it during bounded review/convergence;
- live onboarding/reference docs point only to real current workflow surfaces;
- retired `.agents/skills/` and `metaldocs-*` trees are not restored;
- no new verifier/framework is created merely to prove that agents read the method;
- after real use in MetalDocs, the project-neutral core may be promoted into MNFS/harness for reuse across projects.

## Binding engineering intent

The durable standard must preserve these decisions:

- root cause precedes patch selection;
- always test whether a proposed solution is only a local maximum;
- prefer the Global Maximum structure that resolves the cause and makes the defect class unrepresentable or mechanically impossible at the strongest reasonable boundary;
- simplify accidental complexity without simplifying correctness;
- YAGNI removes speculative capability, not invariants, fail-closed behavior, or proof for reachable states;
- sophisticated enforcement is valid when the material property genuinely requires it;
- transitional solutions name their successor and deletion condition;
- review is adversarial but bounded, and settled decisions reopen only on material findings or changed constraints.

The full maintained definitions and decision flow live only in `wiki/standards/root-cause-global-maximum-method.md`.

## Implementation ownership ruling

The initial design discussion considered `docs/engineering/root-cause-global-maximum-method.md` as the canonical location. During PR #129 review this was disproved against `wiki/standards/documentation-governance.md`:

- `wiki/` owns durable maintained project knowledge;
- `docs/` is staging/draft material;
- cross-cutting durable standards belong in `wiki/standards/`.

Therefore the final and binding implementation location is:

`wiki/standards/root-cause-global-maximum-method.md`

This ruling supersedes the earlier path assumption without changing the accepted engineering doctrine.

## Acceptance

The design is satisfied when:

1. the durable method exists under `wiki/standards/`;
2. live agent entrypoints bridge to it without duplicating its definitions;
3. workflow-specific skills consume it rather than compete with it;
4. dead repository-local skill routing is removed from live entrypoints;
5. no enforcement framework is added solely for this documentation method;
6. PR review/CI validates the final implementation head.
