# MetalDocs Agent Bootstrap

## Fresh-session recovery

Before relying on chat, handoff, or remembered state, revalidate as applicable:

```text
repository / remote identity
current branch + HEAD
remote main HEAD
active PR base/head
required aggregate check
unowned worktree state
```

Then read, in order:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ pinned methodology ROUTER.md
→ selected methodology profile
→ only the 1–2 repository-local owners required by the task
```

Default repository-local task context is at most five files. Do not recursively read `docs/`, Git history, closed PRs, Evidence, or removed implementation without a named material reason.

## Organizational methodology pin

Canonical methodology repository:

```text
developmentconexus-ops/conexus-methodology
@ 9c7210d1504bef01c0d134a6c3ae8627deebb535
```

Start method selection at `ROUTER.md` from that exact commit. Never auto-follow moving methodology `main`, and do not maintain a parallel reusable methodology copy in MetalDocs.

Repository-local Product, architecture, status, and engineering specialization remain owned by MetalDocs; organizational methods do not replace Product meaning.

## MetalDocs hard stops

- `docs/roadmap.md` is the sole mutable stage/status/next-action authority.
- If the roadmap blocks implementation, do not add application code, schema/migrations, executable OpenAPI/generated code, frontend/runtime/deploy, dependency manifests, or dormant capability.
- Do not restore removed legacy implementation merely because it existed or was tested.
- Historical mechanism reuse requires a current named consumer and the proof-backed gate in `docs/architecture/technical-baseline.md`.
- Embedded pre-reset `wiki/...` paths are provenance only, not current routing.
- A material contradiction with current Product/R10 authority is a STOP / bounded reopen, never a silent local patch.

## Repository-local Git and verification

- No direct commits to `main`; no force-push or shared-history rewrite.
- Normal integration is squash merge after explicit operator merge authorization.
- Current aggregate gate is GitHub Actions job `required` in `.github/workflows/ci.yml`.
- Required unmerged provenance/Evidence refs are recorded in current repository authority and must remain reachable while they have a named consumer.
- Independent review uses the pinned `ADVERSARIAL-REVIEW-METHOD.md` selected through `ROUTER.md`; review transport remains Evidence and never enters candidate/main.

`CLAUDE.md` has no independent Product, architecture, roadmap, status, or methodology authority.
