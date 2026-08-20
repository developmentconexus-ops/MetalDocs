---
id: repository-reset-proposal
kind: work
owner: architecture
summary: Temporary proposal for the clean-slate MetalDocs repository reset.
---

# Repository reset proposal

> Temporary / non-authoritative. Delete before merge after ratification.

## Proposed outcome

Replace the current repository tree with an architecture-first baseline containing only:

```text
root agent/navigation files
minimal one-job CI
Product Contract + whole-product alignment
ratified architecture T1 → T8-D under semantic paths
repository reset/documentation/engineering authorities
paused T8-E checkpoint
```

Everything else from the previous implementation is intentionally absent.

## Delete by construction

```text
.claude/
.superpowers/
.qa-reports/
legacy .github workflows/config
API/OpenAPI/generated code
apps/cmd/internal Go implementation
DB baseline/migrations
frontend/packages
Docker/deploy/ops
Go/Node manifests and vendor
scripts/tests/tools
legacy quality/dead-code/lint baselines
wiki/
old docs/superpowers roadmap/harness/milestones/reports/reviews
archive/reviews/scratch material
legacy env/runtime examples
```

## Preserve

The reset rehomes only current ratified product/architecture truth and the paused T8-E checkpoint. Git preserves every deleted source file and historical PR.

## Important semantic distinction

This is a **source-tree reset**, not a production cutover. There is no Launch historical-business compatibility consumer and implementation is currently blocked. Any later runtime/data cutover remains a future transition decision.

## Acceptance tests

The final tree should make these statements true:

1. A new agent can orient from `AGENTS.md` → `docs/index.md` → `docs/status.md` without reading old roadmaps or code.
2. `wiki/`, application code, old OpenAPI, DB, frontend, deploy, scripts, tests and old verifier machinery are absent.
3. Current Product/R10 authorities remain reachable under semantic paths.
4. The accepted T8-E checkpoint is preserved and explicitly paused, not lost or promoted.
5. CI has one required repository-shape job and no implementation-era toolchain burden.
6. Git history remains the only legacy archive.
7. No removed mechanism is implicitly promised reuse.
8. Implementation remains blocked.

## Review focus

Attack decision loss, accidental deletion of current authority, hidden external compatibility obligations, ambiguous agent routing, unnecessary survivors, and any old implementation assumption still present in the live tree.