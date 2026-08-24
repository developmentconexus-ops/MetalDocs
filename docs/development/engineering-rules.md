---
id: engineering-rules
kind: authority
owner: engineering
summary: MetalDocs-specific engineering, Git, CI, provenance, and implementation-gate rules layered on the pinned organizational methodology.
---

# Engineering rules

The exact organizational methodology pin and selection route are owned by `AGENTS.md`. This page contains only MetalDocs-specific controls and MUST NOT duplicate or redefine reusable organizational methodology.

## Implementation gate

`docs/roadmap.md` decides whether implementation is allowed. While blocked:

- do not add application code, schema/migrations, executable OpenAPI/generated code, frontend/runtime/deploy, dependency manifests, or dormant capability;
- do not restore removed implementation for convenience;
- historical mechanism reuse must pass the proof-backed gate in `docs/architecture/technical-baseline.md`.

Before the first future implementation/code/schema/runtime commit is authorized, restore a secret-scanning control appropriate to the new repository shape.

## Frontend planning visual evidence

While an architecture/planning PR remains Draft, P8 rendered structural wireframes may be tracked as temporary HTML Evidence only under:

```text
docs/work/current/*.html
```

This is a MetalDocs repository specialization for preserving operator-reviewable rendered Evidence. It does not admit production frontend code, scripts, dependencies, runtime assets, or HTML outside that bounded temporary-work root. `docs/work/**` remains branch-only and must be absent from a merge candidate/main.

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

Repository-host settings may expose additional merge mechanisms; MetalDocs policy remains squash as normal integration. Host-setting hardening is an administrative control, not a reason to duplicate Git policy in Product/architecture authority.

## Provenance

Required unmerged provenance/Evidence refs are named by durable repository authority and exact-SHA checked by the aggregate gate while they have a current consumer. Do not delete or move them until the owning authority's retirement condition is satisfied.

## Verification

`.github/workflows/ci.yml` owns the executable repository-conformance proof. A control counts only when its negative path is demonstrably capable of firing.

A failing retired/legacy check is not a reason to restore old machinery; first prove the check still protects a current property.
