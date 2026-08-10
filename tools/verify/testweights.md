# Refreshing `testweights.json`

`testweights.json` maps a Go import path to its measured wall-clock seconds.
It exists for one purpose: telling `shardOf` (partition.go) which packages are
expensive, so the four CI shards finish at roughly the same time instead of
three finishing early while one carries `tests/integration/approval`.

**It is advisory.** The set of packages that runs comes from `go list` on the
commit under test, never from this file. A missing entry is weighted at the
median of the entries that exist; a stale entry makes a shard lumpy. Neither
can stop a package from running. If you are unsure whether to refresh it, the
answer is that nothing breaks if you do not — the worst shard just drifts
slowly upward, and nothing fails to warn you, which is exactly why this
procedure is written down instead of remembered.

## Regenerate from a local run

Requires a live Postgres (`METALDOCS_DATABASE_URL`) and cgo for `-race`:

```bash
go test -tags integration -count=1 -race -timeout 900s -json \
  $(go list -tags integration -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./tests/... ./internal/... ./apps/...) \
  | jq -s 'map(select(.Action=="pass" and .Test==null)) | map({(.Package): .Elapsed}) | add' \
  > tools/verify/testweights.json
```

`.Test==null` is what makes this per-PACKAGE rather than per-test: `go test
-json` emits one `pass` event per test and one per package, and only the
package-level event carries the package's total elapsed time.

## Regenerate from a CI run

Local timings are a usable proxy for relative cost but not for absolute cost —
the runner is a different machine, with a different storage stack and a
different Postgres. Real CI numbers are preferable, and the shard step passes
`verify -v` precisely so they survive: `-v` streams a passing check's output,
and a non-verbose `go test` prints exactly one `ok <package> <seconds>` line
per package. Without it the verifier swallows the child's stdout on PASS and
the only number left is the shard total, from which per-package weights cannot
be recovered.

Every package appears in exactly one shard, so one pass over the whole run's
log collects all four legs at once:

```bash
gh run view <run-id> --log \
  | grep -oE 'ok[[:space:]]+metaldocs/[^[:space:]]+[[:space:]]+[0-9.]+s' \
  | awk '{sub(/s$/,"",$3); print $2, $3}' \
  | jq -Rn '[inputs | split(" ") | {(.[0]): (.[1]|tonumber)}] | add' \
  > tools/verify/testweights.json
```

## The version of this that does not rot

Hand-refreshing a timing file is the transitional arrangement, not the target.
Every CI provider with native test splitting (CircleCI, Buildkite, Knapsack)
closes this loop automatically: each shard publishes its timings as an
artifact and the next run consumes them, so the file is never edited by a
person and can never be stale. Doing that here means each shard leg uploading
its `-json` output and a post-merge job committing the merged result. Until
that exists, this document is the procedure — and its absence was the
difference between an advisory file and a maintenance trap.
