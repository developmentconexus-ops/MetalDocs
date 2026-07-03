# Feature F1.1 — oasdiff breaking-change gate — Evidence

> **Status:** CLOSED — all four Validation-Gate criteria met with recorded negative + positive proof.
> **Implemented by:** subagent (sonnet, subagent-driven-development). **Reviewed by:** main session.
> **Contract:** `../validation-contract.md §F1.1`; spec `spec.md`. Zero deviations.

## Files changed

- **`.github/workflows/openapi-breaking.yml`** (new) — blocking `oasdiff-breaking` job on
  `pull_request` modifying the spec or the workflow. No other files touched;
  `api/openapi/v1/openapi.yaml` untouched (its one struct fix is F1.2's).

## Tool

`github.com/oasdiff/oasdiff v1.21.0`, installed in CI via `go install github.com/oasdiff/oasdiff@latest`.
Command shape = `oasdiff breaking <base> <head> --fail-on ERR` (matches validation-contract §F1.1 verbatim).
Base spec materialized via `git show ${{ github.event.pull_request.base.sha }}:api/openapi/v1/openapi.yaml`
after `fetch-depth: 0` checkout so the base sha is reachable.

## Validation Gate — proof

| Criterion | Proof | Result |
|-----------|-------|--------|
| Gate red on synthetic breaking change | `oasdiff breaking original.yaml broken.yaml --fail-on ERR` where broken = `GET /documents` operation deleted | **exit 1**, `[api-path-removed-without-deprecation] … in API GET /documents` — fixture (discarded) |
| Gate green on identical spec | `oasdiff breaking original.yaml original.yaml --fail-on ERR` | `No changes detected`, **exit 0** — real |
| Gate green on compatible change | added optional `debug_note` (not in `required`) to `DocumentListResponse` | `No breaking changes to report…`, **exit 0** — fixture (discarded) |
| Wired blocking, correct path filter | workflow YAML present; `on.pull_request.paths = [api/openapi/v1/openapi.yaml, .github/workflows/openapi-breaking.yml]`; single job/step; no `continue-on-error` | real |
| Workflow syntactically valid | `yaml.safe_load` → top keys `[name, on, jobs]`, `jobs=[oasdiff-breaking]`, 5 steps | real |

## Design decisions (recorded)

- **`pull_request` trigger only.** No `push: branches:[main]` trigger. Rationale captured in the
  workflow's top-of-file comment: a same-branch parent-commit diff on `main` is not a true
  base-vs-proposed comparison and risks false positives on merge/squash commits. `main` is protected
  transitively — every change lands via a PR where this job is blocking.
- Style matched to `api-contract.yml`: `actions/checkout@v4`, `actions/setup-go@v5` (`go-version 1.25.x`,
  `cache: true`), `ubuntu-latest`.

## Bounded defers

None.

## Review disposition

Single new workflow file, isolated (no source/spec/generated changes). Contract §F1.1 satisfied
verbatim; negative + both positive proofs recorded. **APPROVED** — committed.
