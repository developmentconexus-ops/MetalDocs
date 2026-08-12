# Solo-Strong CI v1

## Scope and baseline

This report records the bounded CI/CD improvement from `origin/main` at
`8f4c7ac64d4a34eb5b7331cbe19ef4c0609a00a6`. The live `main` ruleset remains
unchanged: one required context (`required`), strict required checks, resolved
review threads, and no bypass actors.

The existing CI audit was run before implementation:

```text
go run ./tools/verify --audit
verify --audit: 50 checks, 14 workflow jobs, 0 findings
```

The historical wall-clock evidence is preserved rather than reinterpreted:
the roadmap records 17m17s for the pre-sharding PR run `31402055141` and 9m39s
for the four-shard run `31414404392`. Those are repository baseline evidence,
not a claim about this PR's final run.

## Implemented topology

- PR integration remains four `go list`-discovered shards, but the PR registry
  check now runs without `-race`.
- `go-test-integration-race` shares the same partition definition and runs in
  the automatic nightly full integration job and in the release profile with
  `-race`.
- `fe-build` runs the real `pnpm --filter @metaldocs/web run build`, separate
  from typecheck and Vitest.
- `docker-build` discovers tracked production Dockerfiles, rejects an
  unclassified production Dockerfile, builds only affected images, and never
  pushes them.
- `fe-boundary-allowlist` rejects growth of the shrink-only cross-feature
  allowlist or references to unknown feature directories; `eslint.config.mjs`
  discovers feature roots directly from the filesystem.
- `adr-status`, `wiki-debt-tally`, and `db-docs-coverage` are classified as
  documentation hygiene: they remain in fast/full/release and run in the
  explicit nightly `governance-hygiene` job, outside the PR `required` closure.
- `docx-renderer.yml` remains a main-only post-merge recheck; its PR trigger was
  removed because `ci.yml:verify` already owns the blocking DOCX checks.

## Known-red nightly work

The latest observed nightly run before this change was
[31580496114](https://github.com/developmentconexus-ops/MetalDocs/actions/runs/31580496114).
`perf` and `e2e` were red because their database/bootstrap and secret
prerequisites were not provisioned. They are retained for deliberate
`workflow_dispatch` repros behind the explicit `run_deferred=true` input, but
excluded from both the automatic schedule and a default manual dispatch.

Reactivation requires, at minimum:

- `perf`: a live pinned Postgres service, migrations/bootstrap, a configured
  `PERF_DATABASE_URL`, `METALDOCS_ATTACHMENTS_SIGNING_SECRET`, a green health
  probe, and a successful artifact-producing benchmark run.
- `e2e`: a live pinned E2E database, migrations/bootstrap, a configured
  `E2E_DATABASE_URL`, `METALDOCS_ATTACHMENTS_SIGNING_SECRET`, installed
  Playwright browsers, and a green manual run covering the approval flows.

The automatic schedule should only be restored after those prerequisites are
demonstrated and the failure-issue path is verified again.

## Metrics to complete at handoff

| Metric | Baseline before this PR | Final PR evidence |
| --- | ---: | --- |
| PR workflow triggers executing CI | 2 (`ci`, DOCX) | 1 blocking PR workflow; DOCX is main-only |
| Logical PR jobs | 6 | 5 in `ci.yml` |
| Integration shards | 4 | 4 |
| PR integration with `-race` | yes | no |
| Nightly/release full integration with `-race` | nightly stress only / release full | dedicated nightly + release registry check |
| Frontend production build gate | no | `fe-build` |
| Affected production Docker build gate | no | `docker-build` |
| DOCX checks duplicated in PR | 3 | 0 |
| Known-red jobs on automatic/default nightly execution | 2 | 0; explicit opt-in only |
| Documentation hygiene in PR required closure | yes | no; nightly/full/release |
| Required ruleset contexts | 1 (`required`) | 1 (`required`) |
| Final PR makespan | 17m17s pre-sharding PR (`31402055141`) | 4m39s (`31611387106`) |

The final makespan and check-level timings must be copied from the completed
GitHub Actions run in the PR; no estimate is a closure state. The final green
PR run at the implementation head was [31611387106](https://github.com/developmentconexus-ops/MetalDocs/actions/runs/31611387106),
with a measured 4m39s wall-clock duration (15:16:22Z–15:21:01Z). It passed
`verify`, all four integration shards, `security`, `lint-go`, and `required`.
The diff-scoped log selected `fe-boundary-allowlist` and `docker-build`; no
frontend production source/toolchain change or production image artifact was
in this PR, so `fe-build` and actual Docker image compilation were correctly
not selected.

The default manual nightly validation was [31609453086](https://github.com/developmentconexus-ops/MetalDocs/actions/runs/31609453086),
green from 14:55:18Z to 15:07:51Z. `integration-race`, `stress`,
`security-scan`, `governance-hygiene`, and `axe` passed; `perf` and `e2e` were
skipped by the explicit deferred-job gate.
