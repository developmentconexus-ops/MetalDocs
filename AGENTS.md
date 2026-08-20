# MetalDocs Agent Routing

## Entry point

Read `docs/index.md` first. Read `docs/status.md` when the current stage matters. If `docs/work/current/` exists, read its `index.md` before touching the active Draft gate.

Use only the smallest set of authorities required for the task. Do not reconstruct context by reading Git history, closed PRs, or removed implementation unless a current authority names that evidence as necessary.

## Engineering method

For material engineering or architecture decisions, apply the canonical DevelopmentConexus Engineering Method from `developmentconexus-ops/conexus-methodology/METHOD.md`.

Repository/product truth remains local to this repository. External documentation and prior implementation are evidence, never product authority.

## Current hard stops

- `docs/status.md` is the sole stage / implementation-gate authority.
- If status says implementation is blocked, do not create application code, schemas, runtime, deployment, or dormant implementation.
- Do not resurrect removed legacy implementation because it existed, was tested, or is easy to copy.
- Reuse a historical mechanism only when a current authority names a current consumer and reuse is independently smaller than rewrite.
- A material contradiction with a ratified authority is a STOP / bounded reopen, not a local patch.
- Any embedded `wiki/...` path in a carried pre-reset authority is provenance only; current routing comes from `docs/index.md`, `docs/status.md`, and `docs/decisions/index.md`.

## Remaining-stage input

Before a remaining T-stage, read the owning semantic authorities plus `docs/decisions/forward-obligations.md`.

- `PRESERVE` = proof-backed baseline evidence unless materially disproved.
- `REOPEN` = the owning stage must decide deliberately.
- `DEFERRED` = preserve the seam/counterexample; do not build dormant implementation.

Stage ownership after T8-E is defined in `docs/decisions/stage-program.md`.

## Pull requests and review

One coherent ratifiable gate uses one branch and one Draft PR.

For governed architecture work:

```text
proposal converges
→ final independent Fable review in docs/work/current/ai-dialog.md
→ Lead adjudication in the same file
→ bounded Round 2 only if a material contradiction survives
→ explicit operator decision
→ promote durable authority
→ delete temporary work files
→ required checks
→ squash merge
```

Reviewer output is evidence, not authority.

## Git safety

- Never commit directly to `main`.
- Do not force-push or rewrite shared history.
- Git history and explicitly retained refs preserve provenance; do not create live-tree archive/tombstone folders.
- Do not delete an unmerged authority branch that is the only reachable provenance ref until an equivalent immutable archival ref exists.
- Do not merge a governance/architecture PR while temporary `docs/work/**` review artifacts remain.

## Verification

The live repository is intentionally architecture-first while implementation is blocked. Run the checks defined by current repository CI and the active gate. Do not preserve or recreate old verification machinery merely to satisfy superseded implementation assumptions.

The repository ruleset requires status context `required` and resolved review conversations before merge; see `docs/development/engineering-rules.md`.

## Context routing

| Need | Read |
|---|---|
| Product boundary | `docs/product/contract.md` |
| Whole-product alignment | `docs/product/alignment.md` |
| User/API journeys | `docs/product/journeys.md` |
| Current API operation census | `docs/decisions/api-operation-census.md` |
| Semantic ownership | `docs/architecture/ownership.md` |
| Domain state | `docs/architecture/domain-model.md` |
| Lifecycle / transactions | `docs/architecture/lifecycle.md` |
| Authorization / Audit | `docs/architecture/authorization-and-audit.md` |
| Exact content | `docs/architecture/content-integrity.md` |
| Async / Search | `docs/architecture/async-and-search.md` |
| Clean-slate technical posture | `docs/architecture/technical-baseline.md` |
| Backend topology | `docs/architecture/backend.md` |
| Internal contracts | `docs/architecture/interfaces.md` |
| Persistence | `docs/architecture/persistence.md` |
| Remaining stage definitions | `docs/decisions/stage-program.md` |
| Cross-stage forward obligations | `docs/decisions/forward-obligations.md` |
| Paused T8-E checkpoint | `docs/reference/t8e-checkpoint.md` |
| Current gate | `docs/status.md`; plus `docs/work/current/index.md` only when it exists |

`CLAUDE.md` has no independent authority.