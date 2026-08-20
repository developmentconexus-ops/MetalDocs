---
id: repository-reset-proposal
kind: work
owner: architecture
summary: Temporary corrected proposal for the clean-slate MetalDocs repository reset.
---

# Repository reset proposal

> Temporary / non-authoritative. Delete before merge after ratification.

## Proposed outcome

Replace the current repository tree with an architecture-first baseline containing only:

```text
root agent/navigation files
one allowlist-based required CI job
Product Contract + whole-product alignment
ratified architecture T1 → T8-D under semantic paths
remaining T8-F → T12 stage program
52 preserved forward decision obligations
current 78-operation API census precision
repository reset/documentation/engineering authorities
paused durable T8-E checkpoint
```

Everything else from the previous implementation is intentionally absent from the live tree.

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
unused static-doc build configuration
```

## Preserve

The reset rehomes current ratified product/architecture truth, the surviving stage program, the 52 proof-backed forward obligations, and the paused T8-E checkpoint.

Main history preserves the removed legacy implementation. Because the PR #131/#132 authority/review corpus was never merged to `main`, those source branches remain protected provenance refs and MUST NOT be deleted until equivalent immutable archival refs/tags exist.

## Important semantic distinction

This is a **source-tree reset**, not a production cutover. There is no Launch historical-business compatibility consumer and implementation is currently blocked. Any later runtime/data cutover remains a future transition decision.

## Acceptance tests

The final tree should make these statements true:

1. A new agent can orient from `AGENTS.md` → `docs/index.md` → `docs/status.md` without reading old roadmaps or code.
2. `wiki/`, application code, old OpenAPI, DB, frontend, deploy, scripts, tests and old verifier machinery are absent.
3. Current Product/R10 authorities remain reachable under semantic paths.
4. The accepted T8-E checkpoint is durably routed and explicitly paused, not lost or promoted.
5. The current API application census has one durable 78-operation authority.
6. Remaining T8-F→T12 stage ownership is defined without requiring PR archaeology.
7. The old registry is reduced to exactly the 21 PRESERVE + 4 REOPEN + 27 DEFERRED forward obligations still needed by future stages.
8. CI has one `required` allowlist-based repository-shape job and rejects arbitrary new implementation paths.
9. Unmerged PR #131/#132 authority provenance remains reachable until immutable archival refs replace the source branches.
10. No removed mechanism is implicitly promised reuse.
11. Implementation remains blocked.
12. Secret scanning must be reintroduced before the first future implementation/code/schema/runtime commit is authorized.

## Review result

Fable confirmed the clean-slate reset structure as the Global Maximum and returned `APPROVE CLEAN-SLATE REPOSITORY RESET WITH MATERIAL FIXES`.

The material corrections stay inside the selected reset:

```text
preserve unreachable PR provenance refs
route paused T8-E durably
make 78-operation precision durable
restore remaining-stage definitions
restore only forward registry obligations
invert CI to allowed-tree enforcement
remove unused MkDocs surface
record repository-ruleset binding
add explicit checkpoint document kind
centralize legacy-path routing law
restore secret scanning before implementation
```

No Product/R10 semantic reopen and no legacy implementation restoration is required.