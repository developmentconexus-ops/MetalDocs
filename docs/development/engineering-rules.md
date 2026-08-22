---
id: engineering-rules
kind: authority
owner: engineering
summary: MetalDocs-specific engineering, Git, CI, provenance, and implementation-gate rules layered on organizational standards.
---

# Engineering rules

Canonical reasoning: `developmentconexus-ops/conexus-methodology/METHOD.md` v1.0.0.

Canonical repository operation: `developmentconexus-ops/conexus-methodology/REPOSITORY-STANDARD.md` v1.0.0.

This page contains only MetalDocs-specific controls.

## Implementation gate

`docs/roadmap.md` decides whether implementation is allowed. While blocked:

- do not add application code, schema/migrations, executable OpenAPI/generated code, frontend/runtime/deploy, dependency manifests, or dormant capability;
- do not restore removed implementation for convenience;
- historical mechanism reuse must pass the proof-backed gate in `docs/architecture/technical-baseline.md`.

Before the first future implementation/code/schema/runtime commit is authorized, restore a secret-scanning control appropriate to the new repository shape.

## Frontend planning visual evidence

While an architecture/planning PR remains Draft, P8 rendered structural wireframes may be tracked as temporary HTML evidence only under:

```text
docs/work/current/*.html
```

This exception exists so the operator-reviewed visual artifact remains reproducible in its actual rendered medium. It does not admit production frontend code, scripts, dependencies, runtime assets, or HTML outside that bounded temporary-work root. The existing `docs/work/**` merge-candidate rule still requires all such temporary planning artifacts to be absent before integration.

## Forward decision obligations

Remaining architecture stages consume current owning authorities plus `docs/decisions/forward-obligations.md`:

```text
PRESERVE → proof-backed baseline unless materially disproved
REOPEN   → owning stage must decide deliberately
DEFERRED → preserve seam/counterexample; no dormant implementation
```

When an obligation closes or materially refines, update the durable register rather than creating amendment chains.

## Repository protection

Current external repository binding:

```text
required aggregate status context: required
GitHub ruleset id:                 20560142
review conversations:              resolved before merge
```

Do not rename/remove `required` without deliberately updating repository protection.

Normal governance/architecture integration uses squash merge after explicit operator merge authorization. No direct commits to `main`; no force-push or shared-history rewrite.

Repository settings still permit merge-commit and rebase methods at the GitHub repository level; after this governance alignment is later merge-authorized and merged, settings should be tightened to squash-only normal integration, PR-only protected `main`, no force-push/delete, automatic head-branch deletion, and the required aggregate gate, as supported by the hosting configuration.

## Provenance

Unique unmerged Product/R10 governance provenance is protected by explicit archive refs recorded in `docs/decisions/repository-reset.md`. Do not delete those refs while a current authority names the byte-level provenance as required.

## Verification

`.github/workflows/ci.yml` owns the current executable repository-conformance proof. A control counts only when its negative path is demonstrably capable of firing.

A failing retired/legacy check is not a reason to restore old machinery; first prove the check still protects a current property.
