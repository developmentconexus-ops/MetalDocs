# CI restructure — design

Date: 2026-08-07
Status: proposed
Supersedes the CI structure delivered under axis A1 (`docs/superpowers/analysis/A1-handoff.md`, issue #87).
Adversarial review: gpt-5.6-sol, verdict REJECT on draft 1; this is draft 2 with every finding folded in.

## 1. Why

The verifier is not one trusted product. It is 20 workflow files emitting 32 check names, 21 of
them independently required, with no shared entry point and no staging. Measured on PR #96 with
`gh pr checks 96 --json name,state,startedAt,completedAt`:

| Fact | Value |
|---|---|
| Check names on one PR | 32, from 20 workflow files |
| Checks finishing in <=165s | 31 of 32 |
| `full` | 1176s (19.6 min), running **concurrently** with the other 31 |
| Required contexts in the `main` ruleset | 21, `bypass_actors: []` |
| Workflows that never call `tools/verify` | 10 of 20 |
| Workflows with a `concurrency:` group | 0 |
| Third-party actions pinned by SHA | 0 |

Four structural defects follow from that table.

**D-A. The expensive job does not know the cheap ones failed.** `full` starts at t=0 alongside a
25-second lint. When the lint fails, `full` still burns 20 minutes. There is no `needs:` edge
anywhere between a cheap check and an expensive one.

**D-B. Twenty-one required contexts is a deadlock surface.** The ruleset has no bypass actors, so
renaming any one of the 21 jobs leaves every PR permanently unmergeable with no escape hatch.
This is documented in `.github/rulesets/README.md`.

**D-C. `verify` is not the entry point A1 claimed.** Ten workflows bypass it entirely
(`golangci-lint`, `secret-scan`, `supply-chain`, `openapi-breaking`, `perf`, `smoke`,
`e2e-coverage-gate`, `phase3-hardening-gate`, `release-readiness`, `test-nightly`).
`printAudit` in `tools/verify/main.go:489` only tests `c.CIJob != ""` — it never reads a workflow
file, so it reported zero gaps while half the CI ran outside the product.

**D-D. A required check can go green without having tested anything.** `missingInfra`
(`tools/verify/main.go:341`) returns SKIP when `METALDOCS_DATABASE_URL` is unset, and `report`
exits non-zero only on failures. An integration job with a misconfigured database URL is green.
This is the sharpest form of the standing rule *a control that sometimes does not fire is not a
gate* — here it does not fire and reports success.

## 2. Evidence base

Live-inspected via `gh api` / `gh pr checks` on 2026-08-07: `coder/coder`, `go-gitea/gitea`,
`grafana/grafana`, `hashicorp/consul`, `prometheus/prometheus`, `PostHog/posthog`.

Unanimous across all six:

1. **One `if: always()` aggregator job is the sole required check.** Independently invented by
   five separate teams with near-identical code: coder `required`, consul `go-tests-success`,
   prometheus `build_all_status`, grafana `All backend unit tests complete`, posthog `... Pass`.
   The individual jobs are `needs:` entries and are *not* required.
2. **Never a trigger-level `paths:` on a required check.** A job skipped by a job-level `if:`
   reports `success`; a workflow that never dispatches reports *nothing* and hangs pending
   forever. Skip logic belongs in a cheap upstream `changes` job.
3. **Grouping by concern, never by tool.** `make lint-backend` runs N linters inside one job.
4. **`concurrency` + `cancel-in-progress` everywhere.**
5. **Third-party actions pinned by full commit SHA.**
6. **Zero of six use GitHub's merge queue.** gitea created a `merge-queue` ruleset and left it
   `"enforcement": "disabled"`.

PostHog lint-enforces one correctness detail (rule WF007) worth copying: the common guard
`contains(needs.*.result, 'failure')` is **wrong**. It denylists, so `cancelled` passes as green.

## 3. Decisions taken

**Merge queue: no.** `gh repo view` returns `{"isInOrganization":false,"owner":{"login":"leandrotcawork"}}`,
and GitHub offers merge queue only on organization-owned repositories. Transferring to a free
organization was available and was declined, on the evidence that zero of the six reference repos
use it either. Consequence: the merged result is never tested unless we compensate — see §6.

**Red controls: fixed, not excluded.** `test-discipline` (D-13), `db-dictionary` (D-5) and
`req-trace` (D-1) are red today. They are repaired in Phase 0 rather than declared out of scope.
The alternative — dropping them to obtain a green aggregate — would be the exact defect this
program exists to remove.

## 4. Target structure

Three workflow files. `ci.yml` is the only one that blocks a PR.

### 4.1 `ci.yml`

Triggers: `pull_request` on `main`. No trigger-level `paths:`.

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.head_ref || github.ref }}
  cancel-in-progress: ${{ github.event_name == 'pull_request' }}
```

Jobs, and the exact registry IDs each one owns:

| Job | Registry IDs (`tools/verify/registry.go`) | Replaces |
|---|---|---|
| `changes` | — (dorny/paths-filter) | new |
| `lint-go` | `gofmt`, `go-vet`, `go-vet-integration`, `cilint`, `staticcheck`, `module-boundaries`, `test-discipline` | 3 check names |
| `lint-contract` | `problem-codes-fresh`, `api-lint-base-path-v1`, `api-lint-base-path-e2e`, `api-lint-strict`, `api-lint-selftest`, `contract-sync` | 6 check names |
| `lint-frontend` | `fe-eslint`, `css-token-discipline`, `eigenpal-selector-pin`, `docx-v2-typecheck` | 4 check names |
| `governance` | `adr-status`, `wiki-tally`, `db-dictionary`, `req-trace-selftest`, `req-trace` | 3 check names |
| `test-go` | `go-build`, `go-test-unit` | 2 check names |
| `test-frontend` | `fe-typecheck`, `fe-test`, `docx-v2-build`, `docx-v2-test` | 2 check names |
| `test-integration` | `go-test-integration` | the 19.6-min `full` |
| `security` | none — action-native, see §5 | 3 check names |
| **`CI v2 / gate`** | — | **the sole required context** |

7 + 6 + 4 + 5 + 2 + 4 + 1 = **29 registry IDs, all accounted for.** This table is the contract:
if a registry ID is not in exactly one row, the design is incomplete. Draft 1 silently dropped
five (`go-build`, `adr-status`, `db-dictionary`, `req-trace-selftest`, `req-trace`); the mapping
exists so that failure mode is visible rather than latent.

Two side effects worth naming. Putting `docx-v2-build` and `docx-v2-test` in the same job in that
order repairs the ordering defect admitted at `registry.go:273`. Putting `module-boundaries` and
`test-discipline` in `lint-go` keeps them at static-analysis cost rather than behind the test
matrix.

### 4.2 Staging — the fix for the actual complaint

```yaml
test-integration:
  needs: [changes, lint-go, lint-contract, lint-frontend, test-go]
```

The 20-minute job does not start until the cheap jobs pass. It stays **pre-merge and blocking** —
A1 item 4 requires the full suite pre-merge, and `test-full.yml:2` says the same. Deferring it to
a nightly cron was draft 1's proposal and was rejected as a control regression: it would trade a
real gate for a report nobody reads.

This is the whole of the fix. The complaint was never "the full suite is too slow", it was "the
final one runs when everything is already failing". A `needs:` edge answers exactly that, and
costs no coverage.

### 4.3 `CI v2 / gate`

The name is qualified because a job named `gate` already exists at `req-traceability.yml:33`, and
because a required context name is effectively permanent — with no bypass actors, renaming it
deadlocks every open PR.

```yaml
gate:
  name: CI v2 / gate
  if: ${{ always() }}
  needs: [changes, lint-go, lint-contract, lint-frontend, governance,
          test-go, test-frontend, test-integration, security]
  runs-on: ubuntu-latest
  env:
    NEEDS_JSON: ${{ toJSON(needs) }}
  steps:
    - name: Validate all CI results
      shell: bash
      run: |
        set -euo pipefail
        if ! jq -e '
          (keys | sort) == ([
            "changes","lint-go","lint-contract","lint-frontend","governance",
            "test-go","test-frontend","test-integration","security"
          ] | sort)
          and .changes.result == "success"
          and all(to_entries[];
                  if .key == "changes" then .value.result == "success"
                  else (.value.result == "success" or .value.result == "skipped") end)
        ' <<<"$NEEDS_JSON" >/dev/null; then
          jq -r 'to_entries[] | "\(.key)=\(.value.result // "null")"' <<<"$NEEDS_JSON"
          exit 1
        fi
```

Three properties, each load-bearing:

- **`if: always()` with no other predicate.** `always() && <anything>` can skip the job, and a
  skipped required check is not a verdict.
- **Exact set equality on `keys`, not just a failure scan.** A job silently dropped from `needs:`
  is caught. A failure scan would not catch it — the aggregate would go green over a shrinking
  set, which is precisely how this class of gate decays.
- **Allowlist, not denylist.** `needs.<job>.result` has exactly four values: `success`, `failure`,
  `cancelled`, `skipped`. Anything not in the allowlist — including `null` and any value GitHub
  adds later — fails. The tempting `any(.[]; .result == "failure")` accepts `cancelled` as green.

### 4.4 `nightly.yml`

Cron. Perf, e2e, axe, cross-platform. Advisory, never PR-blocking. Two jobs move here because
they are advisory *today* and pretending otherwise is the lie A1 was chartered to remove:
`Axe baseline` reads only the accepted-violations list and never an axe report (D-18), and `perf`
produces percentiles over failed requests.

### 4.5 What is not built

The consul-style `verify-ci.yml` no-op canary is **not** adopted as a required check. If `ci.yml`
never dispatches, `CI v2 / gate` never reports, stays pending, and blocks the merge — the canary
adds no safety guarantee, only a second permanent required name and a second deadlock surface.
It may exist as an advisory diagnostic; it will not be required.

## 5. Narrowing the `verify` claim honestly

A1 item 5 claimed `verify` is the entry point "CI calls and nothing else". That is false and the
correct repair is to narrow the claim, not to force unlike controls through an unsuitable
abstraction.

**`verify` owns portable deterministic checks** — anything a developer can run identically on a
laptop and in CI.

**Workflow YAML owns GitHub-native mechanics** — SARIF upload and `security-events: write`
(gitleaks, grype), `services:` provisioning, full-history checkout, fork-safe telemetry,
tag-triggered SBOM generation, Dependabot labeling, and the oasdiff breaking-change gate. These
are not deterministic local commands and do not belong in a Go runner.

Two things must then become true, and neither is true today:

**`verify --require-infra`.** In CI, a SKIP from `missingInfra` must be fatal. Today a
missing `METALDOCS_DATABASE_URL` yields SKIP, `report` exits 0, and the required integration job
is green over zero integration tests (D-D). Every `ci.yml` invocation passes `--require-infra`.
This must land *before* `CI v2 / gate` is promoted to required.

**`verify --audit` must read YAML.** The reverse direction — a workflow job that runs checks
outside the registry — is currently unenforced, which is how ten bypassing workflows went
unnoticed. `--audit` parses `.github/workflows/*.yml`, and fails when a job's `--only=` set
disagrees with the registry's `CIJob` mapping, or when a job named in the gate's `needs:` does not
exist. Until it parses YAML, the audit is a slogan.

## 6. Ruleset changes

**Required contexts: 21 → 1** (`CI v2 / gate`).

**`strict_required_status_checks_policy: false` → `true`.** Currently at `main.json:39`. Without a
merge queue, a PR can pass against a stale base and the combined merge result is never tested.
Requiring the branch to be current before merge is the available compensation. This was decided
together with §3's merge-queue decision — declining the queue is only safe with this flag on.

`bypass_actors: []`, `deletion`, `non_fast_forward`, and the `pull_request` rule are unchanged.

## 7. CodeRabbit

`.coderabbit.yaml` at the repository root. Free for public repositories, rate-limited. Custom
pre-merge checks require a paid plan and are out of scope under the zero-spend constraint.

```yaml
reviews:
  profile: chill
  commit_status: true
  auto_review: { enabled: true, drafts: false, base_branches: ["main"] }
  path_filters:
    - "!vendor/**"
    - "!third_party/**"
    - "!**/*.gen.go"
    - "!**/generated/**"
    - "!frontend/apps/web/node_modules/**"
  path_instructions:
    - path: "internal/modules/**"
      instructions: "Enforce ADR 0022 capability-based authz. Flag any reasoning of the form
        'admin/author/editor can X' — authorization is expressed in capabilities, never roles."
    - path: "api/openapi/**"
      instructions: "Routes change only by editing the spec and regenerating. Flag hand-edited
        generated client or server code."
knowledge_base:
  code_guidelines:
    enabled: true
    filePatterns: ["CLAUDE.md", "wiki/architecture/backend-target-architecture.md"]
```

**Advisory only, at first.** It is promoted to a required context only after its actual check
context string has been *observed* on a real PR in this repository. This is the D-17 lesson:
promotion follows an observed green, never a predicted one. Draft 1 of the A1 work promoted
`hardening` on the assumption it would pass once a fix landed; it had never been green, and every
PR was blocked.

Note that `knowledge_base.filePatterns` sends the named files to CodeRabbit's servers. Both files
are already public in a public repository, so this adds no exposure — but the list is deliberate
rather than wildcarded, so that stays true.

## 8. Sequencing

**Phase 0 — make the baseline honestly green.** A gate cannot be promoted over red controls, and
a control that fires into a red baseline is absent.

- D-13: port 13 `check-test-discipline.sh` violations to the canonical fixture seam
  (R1 raw `set_config('metaldocs.asserted_caps', ...)` ×8, R4 raw `documents` SQL from approval
  tests ×4, R3 ×1).
- D-5: write the 10 missing `wiki/database/tables` pages (`audit_export_jobs`,
  `materialize_dispatch_outbox`, `tenant_keys`, `tenant_lifecycle_jobs`, `tenant_plans`,
  `token_dictionary_entries`, `approval_delegations`, `approval_review_verdicts`,
  `approval_route_stage_selectors`, `release_generations`).
- D-1: give each of the 4 MUST REQs a live test to cite.
- Diagnose why `full` is red. Never diagnosed; it is a precondition, not an afterthought.

**Phase 1 — `verify` fitness.** Land `--require-infra` and the YAML-parsing `--audit`. Neither the
gate nor any promotion happens before these exist.

**Phase 2 — build alongside.** PR A adds `ci.yml`, `nightly.yml`, `.coderabbit.yaml`, SHA-pins
every action, and adds `concurrency` groups. It deletes and renames **nothing**. It merges under
the existing 21 required contexts.

**Phase 3 — observe.** `CI v2 / gate` must be seen green on a real PR before it is named anywhere
in the ruleset.

**Phase 4 — swap the ruleset, in two atomic API edits.** First: add `CI v2 / gate` while retaining
all 21 old contexts. Then, once every open PR's latest SHA carries the new context: remove the 21
while retaining `CI v2 / gate`, and flip `strict_required_status_checks_policy` to `true`.
Enforcement is never disabled at any point. There is no window in which `main` is unprotected and
no window in which every PR is stuck pending.

**Phase 5 — delete.** PR B removes the superseded workflows and re-exports the live ruleset to
`.github/rulesets/main.json` with a corrected README (which currently says 22 required checks
while the JSON contains 21 — the drift this whole design exists to stop).

## 9. Known decay mechanism

This design's own most likely lie, in six months, is "the gate enforces every check". Four
inventories must agree by hand: `registry.go`, each job's `--only=` arguments, `gate.needs`, and
the `changes` outputs. Adding a check to one does not update the other three.

The §4.1 mapping table and the exact-set-equality guard in §4.3 make a drift *visible* — they do
not make it *impossible*. The unrepresentable-state answer is one generated manifest that owns
registry membership, job routing and gate dependencies, with the workflow YAML generated from it.
That is the global maximum here and it is **not** in this design's scope. Per the repository's
local-maximum rule, this is therefore labelled transitional: the structure that deletes it is the
generated CI manifest, and the milestone that must build it is the follow-on to this one. Shipping
§4 without that label would itself be the defect.

## 10. Non-goals

- Fixing the defects the newly-honest checks reveal (axes A2–A9), beyond the Phase 0 three.
- Any paid tier of any tool. Zero spend stands.
- Sharding the integration suite. Go already runs packages concurrently; naive sharding is not
  verified to help, and the `needs:` edge in §4.2 removes the pain without touching the suite.
