# Branch rulesets

`main.json` is an export of the ruleset applied to the default branch. It is
checked in so the gate configuration is reviewable in a diff instead of living
only in the repository settings UI, where a change leaves no trace.

**This file is documentation, not the source of truth.** GitHub does not read it.
The live ruleset is the one that enforces. Re-export after any settings change:

```bash
gh api repos/developmentconexus-ops/MetalDocs/rulesets/20560142 | jq -S . > .github/rulesets/main.json
```

## What it enforces

- `main` cannot be deleted or force-pushed.
- All changes arrive through a pull request. Review threads must be resolved;
  approvals are set to 0 because a solo author cannot approve their own PR, and
  a rule that cannot be satisfied is a rule that gets bypassed.
- **One** status check is required: `required`. `bypass_actors` is `[]` and
  `strict_required_status_checks_policy` is `true`.

## `required` is an aggregator, not a single check

A required check is matched **by name**, and a check that never reports counts
as pending forever — so a workflow with a `paths:` filter cannot be required
without deadlocking every PR that does not touch those paths (see the A1
verification ledger §1b). Naming individual job or check IDs directly in the
ruleset also meant every rename of any one of them was a ruleset edit that had
to land in lockstep with the workflow change, on pain of deadlocking every
open PR — the exact failure mode the "If a required check name changes"
section below used to warn about for 22 separate names.

The fix collapses the ruleset's surface to one name. `ci.yml`'s `required` job
is the sole required context; it is `if: always()` so it reports even when an
upstream job failed, and its own step fails the job (not GitHub's "skipped
still counts" default) unless every upstream job succeeded. Its `needs:` list
is the membership set:

```yaml
required:
  needs: [verify, test-integration, security, lint-go]
```

`verify` is itself an aggregate: `go run ./tools/verify --profile=changed`
runs every owned check registered in `tools/verify/registry.go` whose declared
`Paths` the diff touches; the four PR integration shards invoke the explicit
non-race check, and `security` invokes the complete PR security selection.
Together they cover — gofmt, go vet, arch-lint, module-imports, test-conventions, the
contract/codegen/openapi lints, eslint, the frontend boundary ratchet,
frontend production build, css-tokens, fe-typecheck, fe-test,
affected production Docker builds, docx-typecheck/build/test, migration-gapless,
governance-diff-rules,
invariant-coverage-map, and the registry's own self-tests, among others. A
check with no declared `Paths`
always runs (`matchesPaths`' fail-closed default). Renaming, adding, or
removing one of those checks is a `tools/verify/registry.go` + `ci.yml`
`--only=`/`Paths` change; the ruleset itself never needs to know a check's
name and cannot deadlock on one changing.

`adr-status`, `wiki-debt-tally`, and `db-docs-coverage` are documentation
hygiene rather than merge correctness. They remain in the local fast profile,
the full/release profile, and the explicit `nightly.yml:governance-hygiene`
job; they are intentionally outside `ci.yml:required` so documentation debt
cannot be confused with a broken contract, test, security control, or build.

The membership set is not just eyeballed against `ci.yml`: `required`'s own
step evaluates `scripts/required-gate.jq` against `toJSON(needs)`, which
asserts **exact set equality** — `(keys | sort) == (["lint-go", "security",
"test-integration", "verify"] | sort)` — and that every one of those four
results is `"success"` (no "skipped is green" allowance). The
`required-gate-selftest` registry check (`scripts/check-required-gate.sh`)
pins that predicate itself down with fixtures for the accept/reject cases, so
a PR that quietly loosens the jq expression is caught the same way a PR that
quietly narrows the `needs:` list is.

## Security checks inside the required closure

`gosec` and `govulncheck` are PR-blocking registry checks owned by
`ci.yml:security`, and that job is itself a dependency of `ci.yml:required`.
Therefore both checks are inside the required closure and can block a merge.
They are not direct GitHub ruleset contexts: the live ruleset still requires
only the single `required` context. The same scanners are repeated by
`nightly.yml:security-scan` against the live vulnerability database so new
disclosures after merge are caught without changing the merge-gate topology.
See their entries in `tools/verify/registry.go` for pins and scope.

## If a job inside `required`'s `needs:` closure is renamed

There is exactly one required context (`required`) and no bypass actors, so
this ruleset itself never needs an edit for that rename — `required`'s own
name never changes. What must move together, in the same commit, is:
`tools/verify/registry.go` (the check's `ID`), every `--only=` list in
`ci.yml`/`nightly.yml`/`docx-renderer.yml` that names it, and any doc or
script that references the old ID. `--audit` (registry rule A1) catches an
`--only=` still naming an ID the registry no longer has.

If the **job name** inside `required`'s `needs:` list changes (`verify`,
`test-integration`, `security`, `lint-go`), update both `ci.yml`'s `needs:`
list and `scripts/required-gate.jq`'s literal array in the same commit —
`required-gate-selftest` fails otherwise, and it fails locally, not just in
CI, because it is in the `fast` profile.

```bash
gh api -X PUT repos/developmentconexus-ops/MetalDocs/rulesets/20560142 --input .github/rulesets/main.json
```
