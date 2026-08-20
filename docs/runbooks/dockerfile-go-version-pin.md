# Runbook — Dockerfile Go builder version vs. go.mod

**Scripts:** `scripts/check-dockerfile-go-version.sh`
**Check:** `tools/verify` check `dockerfile-go-version-drift`
**Owner:** on-call (whoever is triaging a red image build or a red
`dockerfile-go-version-drift` check)
**Context:** 2026-08-11 trunk-health incident — `go.mod`'s `go` directive had
moved to `go 1.26.5` to remediate GO-2026-5856 (deliberately the `go`
directive and not `toolchain`; see the comment above it in `go.mod`), while
`deploy/docker/api.Dockerfile`, `jobs.Dockerfile`, and `worker.Dockerfile`
still hardcoded `FROM golang:1.25-alpine`. Nothing forced the two to agree,
and nothing in CI noticed: every GitHub Actions workflow derives its Go
toolchain via `go-version-file: go.mod` (`.github/workflows/*.yml`), so CI's
own Go steps tracked the bump correctly — but no CI job builds container
images, so the drift only surfaced the first time someone actually built one,
as `go: go.mod requires go >= 1.26.5 (running go 1.25.12; GOTOOLCHAIN=local)`
failing inside `go mod download` mid-build. This is the repo's own named
"hand-synced enumeration" defect class
(`docs/engineering/defect-class-catalog.md`, Class 2) applied to a value that
happens to live in a Dockerfile instead of a Go source file.

## Why this is a labelled-transitional literal, not a derived value

Docker's `FROM` instruction resolves its image reference at parse time, from
a literal or from an `ARG` supplied via `--build-arg` — it cannot read
`go.mod` to compute its own value, so true derivation (the kind
`go-version-file: go.mod` gives CI for free) is not available inside a
Dockerfile. The current Go-builder Dockerfiles carry a
`FROM golang:1.26.7-alpine` line, restated by hand. The api/jobs/worker files
also carry a comment block naming this as a TRANSITIONAL local maximum under
CLAUDE.md's "Global Maximum, Not Local Maximum" rule: the global-maximum
structure is one parameterized multi-stage Dockerfile shared by the Go
binaries with `GO_VERSION` threaded from `go.mod` through the compose build
stanza. That consolidation is outside the security-baseline patch; until it
happens, `dockerfile-go-version-drift` keeps the restatement honest.

A module-version change must also run `go mod tidy` and `go mod vendor`; the
repository commits `vendor/`, so updating `go.mod` alone makes every Go job
fail closed with an inconsistent-vendoring error before tests or scanners run.

## What `dockerfile-go-version-drift` does

Static, no Docker daemon, no network. Parses `go.mod`'s `^go X.Y(.Z)?$`
directive line, discovers every tracked Dockerfile via
`git ls-files -- '*.Dockerfile' '*Dockerfile*'` (deliberately not a
hand-kept list — see the comment in the script), and for each one that
declares a `FROM golang:<version>` builder stage, fails if that version is
less than `go.mod`'s. `scripts/testdata/guard-fixtures/` is excluded from
discovery, because that directory deliberately contains a drifted-on-purpose
fixture Dockerfile the check must reject in isolation but must never treat
as a real file in the tree it is checking.

## When it goes red

- **A real Dockerfile pins a `golang:` version below `go.mod`'s `go`
  directive.** This is the check working as intended — bump that
  Dockerfile's `FROM golang:X.Y.Z-alpine` line to match (or exceed) `go.mod`,
  in the same commit that changed `go.mod` if that is what caused the drift.
- **`go.mod`'s `go` directive changed and no Dockerfile was touched.** Same
  fix: bump every `deploy/docker/*.Dockerfile` (and
  `apps/docx-renderer/Dockerfile` if it ever grows a `golang:` builder stage
  — today it only has a `node:` stage) to the new version.
- **The check itself looks wrong** (false positive on a correct tree, or a
  false negative on a drifted one): run
  `go run ./tools/verify -guard-fixtures -only dockerfile-go-version-drift`
  against the fixture in
  `scripts/testdata/guard-fixtures/dockerfile-go-version-drift/`, and/or run
  `bash scripts/check-dockerfile-go-version.sh` directly from repo root and
  read its `DOCKERFILE-GO-VERSION-DRIFT:` lines. A prior version of this
  script had exactly this failure mode: its discovery glob matched the
  fixture's own `.txt` file by substring, so it went red on a clean checkout
  for the wrong reason — that is why the fixture directory is excluded.

## How to confirm an image actually builds after fixing a drift

This check is static and does not build anything. To prove the real defect
this runbook exists for is gone, build the affected images from repo root:

```sh
docker compose -f deploy/compose/docker-compose.yml --env-file .env build db-provision api worker jobs
```

A pre-fix tree fails inside `go mod download` with the `GOTOOLCHAIN=local`
error quoted above; a post-fix tree completes and `docker images` shows the
freshly built MetalDocs Go-service images.
