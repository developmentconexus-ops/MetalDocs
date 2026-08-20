# MetalDocs Agent Bootstrap

## Fresh-actor route

Read, in order:

1. `AGENTS.md`
2. `docs/index.md`
3. `docs/roadmap.md`
4. only the 1–2 owning documents named for the task

Default task context is at most five files. Do not recursively read `docs/`, Git history, closed PRs, or removed implementation without a named material reason.

## Organizational standards

Canonical engineering reasoning: `developmentconexus-ops/conexus-methodology/METHOD.md` — DevelopmentConexus Engineering Method v1.0.0.

Canonical repository operating model: `developmentconexus-ops/conexus-methodology/REPOSITORY-STANDARD.md` — DevelopmentConexus Repository Standard v1.0.0.

Repository-local Product and architecture authority remains in this repository; external standards do not replace Product meaning.

## MetalDocs hard stops

- `docs/roadmap.md` is the sole mutable stage/status/next-action authority.
- If the roadmap blocks implementation, do not add application code, schema, OpenAPI implementation, frontend/runtime/deploy, dependency manifests, or dormant capability.
- Do not restore removed legacy implementation from Git merely because it existed or was tested.
- Historical mechanism reuse requires a current named consumer and the proof-backed reuse gate in `docs/architecture/technical-baseline.md`.
- Embedded pre-reset `wiki/...` paths are provenance only, not current routing.
- A material contradiction with current Product/R10 authority is a STOP / bounded reopen, never a silent local patch.

## Repository-local Git and verification

- No direct commits to `main`; no force-push or shared-history rewrite.
- Normal integration is squash merge after explicit operator merge authorization.
- Current aggregate gate is GitHub Actions job `required` in `.github/workflows/ci.yml`.
- Required unmerged provenance refs are recorded in `docs/decisions/repository-reset.md`; do not delete them while they have a named consumer.

Future independent Fable review follows Repository Standard v1 isolation: candidate branch → `review/<gate>-fable`, with only `docs/work/current/ai-dialog.md` differing from the candidate. Reviewer output is Evidence, never authority.

`CLAUDE.md` has no independent Product, architecture, roadmap, or status authority.