---
id: engineering-rules
kind: authority
owner: engineering
summary: MetalDocs-specific execution, Git, CI, provenance, and implementation-gate rules.
---

# Engineering rules

This page contains only MetalDocs-specific controls. Reusable engineering/frontend reasoning lives in the repository-local method files:

- [`engineering-method.md`](engineering-method.md) — DevelopmentConexus Engineering Method v1.0.0;
- [`frontend-product-experience-planning-method.md`](frontend-product-experience-planning-method.md) — Frontend Product Experience Planning Method v2.3.

`AGENTS.md` and `docs/index.md` route to these files directly. There is no external methodology router, pin, profile-selection step, file-count limit, owner-count limit, or context budget.

## Implementation gate

`docs/roadmap.md` decides whether implementation is allowed. While blocked:

- do not add application code, schema/migrations, executable OpenAPI/generated code, frontend/runtime/deploy, dependency manifests, or dormant capability;
- do not restore removed implementation for convenience;
- historical mechanism reuse must pass the proof-backed gate in `docs/architecture/technical-baseline.md`.

Before the first future implementation/code/schema/runtime commit is authorized, restore a secret-scanning control appropriate to the new repository shape.

## Frontend planning visual evidence

While an architecture/planning PR remains Draft, P8 rendered structural wireframes may be tracked as temporary HTML Evidence under:

```text
docs/work/current/*.html
```

This does not admit production frontend code or runtime assets. `docs/work/**` is temporary planning material and must be absent from a merge candidate/main.

## Forward decision obligations

Remaining architecture stages consume current owning authorities plus `docs/decisions/forward-obligations.md`:

```text
PRESERVE → proof-backed baseline unless materially disproved
REOPEN   → owning stage must decide deliberately
DEFERRED → preserve seam/counterexample; no dormant implementation
```

When an obligation closes or materially refines, update the durable register rather than creating amendment chains.

## Repository protection

Current protected aggregate status context remains:

```text
required
```

Do not rename/remove `required` without deliberately updating repository protection.

Normal integration is squash merge after explicit operator merge authorization. No direct commits to `main`; no force-push or shared-history rewrite.

## Provenance

Durable Evidence locators remain repository evidence and may name exact historical refs/blobs when needed to recover accepted LOCK evidence. They do not need to become permanent network/SHA checks in every unrelated CI run.

## Verification

`.github/workflows/ci.yml` owns the small required repository safety net.

Required CI protects objective properties only: essential operating files, unresolved merge-conflict markers, implementation-block violations in the PR diff, and temporary `docs/work/**` material entering a merge candidate.

Do not use CI to judge Global Maximum, evidence quality, UX/architecture quality, number of files read, methodology selection, documentation reachability, historical status prose, or Evidence-branch freshness. Those are engineering/review questions governed by the adopted methods.

Run targeted proof when a specific Product/architecture/wireframe claim requires it. A failing retired or historical checker is not a reason to restore old machinery; first prove the checker still protects a current property.
