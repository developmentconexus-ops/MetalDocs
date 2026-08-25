---
id: engineering-rules
kind: authority
owner: engineering
summary: MetalDocs-specific execution, Git, CI, provenance, and implementation-gate specializations.
---

# Engineering rules

This page contains **only MetalDocs-specific controls**.

Reusable operation/reasoning lives in the local methods:

- [`engineering-method.md`](engineering-method.md) — material engineering reasoning / Global Maximum;
- [`repository-method.md`](repository-method.md) — repository continuity, selective context, documentation, Evidence and PR/Git lifecycle;
- [`frontend-product-experience-planning-method.md`](frontend-product-experience-planning-method.md) — frontend Product Experience P0–P14.

`AGENTS.md` routes to the applicable method(s). Do not restate those methods here.

## Implementation gate

`docs/roadmap.md` decides whether implementation is allowed. While it says implementation is blocked:

- do not add application code, schema/migrations, executable OpenAPI/generated code, frontend/runtime/deploy, dependency manifests, or dormant capability;
- do not restore removed implementation for convenience;
- historical mechanism reuse must pass the proof-backed gate in `docs/architecture/technical-baseline.md`.

Before the first future implementation/code/schema/runtime commit is authorized, restore a secret-scanning control appropriate to the new repository shape and prove its negative path.

## Frontend planning Evidence

While an architecture/planning PR remains **Draft**, functional P8 rendered wireframes may be tracked temporarily under:

```text
docs/work/current/*.html
```

Other temporary planning/proof notes may also use `docs/work/current/` when they have a real branch-local consumer.

This does not admit production frontend/runtime code. `docs/work/**` must be absent before a PR becomes a merge candidate and must never enter `main`.

## Forward decision obligations

Remaining stages consume current owning authorities plus `docs/decisions/forward-obligations.md`:

```text
PRESERVE → proof-backed baseline unless materially disproved
REOPEN   → owning stage must decide deliberately
DEFERRED → preserve seam/counterexample; no dormant implementation
```

When an obligation closes or materially refines, update the durable register rather than creating amendment chains.

## Independent adversarial review

MetalDocs retains the useful **ClaudeCode/FABLE-style review posture** without restoring the historical `.claude/skills` framework or introducing a fourth methodology.

Use an independent fresh challenger when required by the Engineering Method, when a material Product/architecture/trust-boundary decision is being ratified, at major stage/global-coherence closeout, at the integrated frontend P11/P12 challenge, and before implementation authorization.

The challenger must be read-only with respect to the candidate and should:

```text
verify anchors / premises first
→ root cause before patch
→ Local Maximum vs Global Maximum
→ simplify / YAGNI pass
→ classify findings against current authority
→ dispose prior findings explicitly
→ attack new material uncertainty
→ stop on convergence
```

Reviewer output is Evidence, not authority. A finding that implies new Product meaning returns to the owning decision. Repeated same-altitude findings are a signal to stop patching and revisit structure once, not to create unlimited review rounds.

Do **not** require a fresh independent review for every normal P7/P8 iteration, copy/layout correction, mechanical trace update or already-owned bounded fix. Use targeted proof during the inner loop and one strong independent challenge when a material candidate reaches its real gate.

If actual ClaudeCode/FABLE transport is unavailable, do not claim independent FABLE convergence. Lead/self-review may use the same doctrine, but independent gates remain independent when required.

## Repository protection

Current required aggregate status context:

```text
required
```

Do not rename/remove `required` without deliberately updating repository protection.

Normal integration is squash merge after explicit operator merge authorization. No direct commits to `main`; no force-push or shared-history rewrite.

Ordinary merged head branches should be deleted after successful integration once no independent provenance consumer remains. Archive/Evidence refs follow their named reachability/retirement law instead.

## Provenance

Durable Evidence locators may name exact historical refs/commits/blobs when a current frontend or provenance consumer needs byte-level recovery.

Required refs must remain reachable while named by current authority. They do **not** need permanent network/SHA checks in every unrelated CI run; verify exact identity when the claim or cleanup action actually depends on it.

## Verification

`.github/workflows/ci.yml` owns the small required repository safety net.

Required CI protects objective properties only:

- essential operating files;
- unresolved merge-conflict markers in the PR diff;
- implementation-block violations in the PR diff;
- temporary `docs/work/**` entering a non-Draft merge candidate.

Do not use CI to judge Global Maximum, evidence quality, UX/architecture quality, context size, number of files read, historical prose, reviewer preference, or Evidence-branch freshness. Those are method/review questions.

Run targeted proof when a specific Product/architecture/frontend claim requires it. A failing retired or historical checker is not a reason to restore old machinery; first prove it still protects a current property.
