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

Twenty workflow files become four: `ci.yml`, `nightly.yml`, `release.yml`, and `smoke.yml` (kept
as-is — `workflow_dispatch` only, honest header, targets a deployed environment). `ci.yml` is the
only one that blocks a PR.

### 4.1 `ci.yml`

Triggers: `pull_request` on `main`. No trigger-level `paths:`.

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.head_ref || github.ref }}
  cancel-in-progress: ${{ github.event_name == 'pull_request' }}
```

Jobs, and every control each one owns. **Registry IDs are not the unit of account.** Draft 2
made that mistake: it balanced 29 registry IDs and declared the table complete, while seven
controls that bypass the registry — five of them *required contexts today* — had no row at all.
Phase 5 would then have deleted `invariants.yml` and dropped the migration-monotonicity gate
without anything noticing. That is the same blindness as `printAudit` testing only
`CIJob != ""` (D-C), reproduced inside the instrument meant to catch it. The unit of account is
**every control that runs in CI today**, registry-backed or not.

| Job | Registry IDs | Non-registry controls it also owns |
|---|---|---|
| `changes` | — | dorny/paths-filter (new) |
| `lint-go` | `gofmt`, `go-vet`, `go-vet-integration`, `cilint`, `staticcheck`, `module-boundaries`, `test-discipline` | **`golangci-lint`** (see §4.6) |
| `lint-contract` | `problem-codes-fresh`, `api-lint-base-path-e2e`, `api-lint-strict`, `api-lint-selftest`, `contract-sync` | **`openapi-lint`** (redocly), **`backend-codegen-drift`**, **`frontend-codegen-drift`**, **`oasdiff-breaking`** |
| `lint-frontend` | `fe-eslint`, `css-token-discipline`, `eigenpal-selector-pin`, `docx-v2-typecheck` | — |
| `governance` | `adr-status`, `wiki-tally`, `db-dictionary`, `req-trace-selftest`, `req-trace` | **`Migration monotonicity check`**, **`check-governance.ps1`**, the invariant→spec-ID coverage grep |
| `test-go` | `go-test-unit` | — |
| `test-frontend` | `fe-typecheck`, `fe-test`, `docx-v2-build`, `docx-v2-test` | — |
| `test-integration` | `go-test-integration` | — |
| `security` | — | gitleaks, grype |
| **`required`** | — | **the sole required context** |

Registry IDs: 7 + 5 + 4 + 5 + 1 + 4 + 1 = **27 placed, 2 deleted** (`go-build`,
`api-lint-base-path-v1` — see §4.5), totalling the 29 in `registry.go`. Non-registry controls:
**11 placed.** Both counts are part of the contract. A control absent from this table is a
control Phase 5 deletes.

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

### 4.3 `required`

The name is `required`, matching coder/coder exactly. Draft 2 proposed `required`; that was
rejected on two counts. A slash faking a namespace inside a job name is a homegrown convention no
external reader has seen — every context in `main.json` today is a bare job name. And a required
context name is effectively permanent, so baking a migration epoch ("v2") into it means the repo
carries a 2026 rewrite in its most durable identifier forever. The collision with the job named
`gate` at `req-traceability.yml:33` disappears in Phase 5 when that file is deleted, and during
Phases 2–4 the two coexist harmlessly because only `required` is ever promoted.

```yaml
required:
  name: required
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

### 4.4 `nightly.yml` and `release.yml`

`nightly.yml` — cron. Perf, e2e, axe, cross-platform, gosec/govulncheck. Advisory, never
PR-blocking. Jobs move here because they are advisory *today* and pretending otherwise is the lie
A1 was chartered to remove. `perf.yml` carries a header saying "KNOWN RED: this job cannot pass
today" **on a `pull_request` trigger** — a check guaranteed red on every approval-touching PR
teaches the whole team to ignore red, which is the most corrosive thing a CI surface can do.

`release.yml` — the fourth file, forced into existence by §5's decision to keep tag-time SBOM
generation workflow-native. Draft 2 said "three workflow files" and named two; an unnamed file is
an unowned file.

### 4.5 Deletions

Six controls cease to exist because they cannot fail for the reason they claim, or because
another control strictly subsumes them. Each is evidenced.

| Control | Why it dies |
|---|---|
| `go-build` (registry) | `go test ./...` compiles every package including those with no test files, and `go vet ./...` compiles them again. Neither links binaries, so `go build ./...` verifies nothing the other two lack. |
| `api-lint-base-path-v1` (registry) | `-only` is a *filter*, not a mode (`scripts/api-lint/main.go:21,64-67`), so PATH-BASE-PREFIX already runs inside `api-lint-strict` on the same file. `api-lint-base-path-e2e` survives — `-strict` never touches `internal-e2e.yaml`. |
| `test-smoke.yml:smoke` | A hand-listed `go test -run "TestTriggerBypass\|TestMembership\|..."` regex that exits 0 when every named test is renamed away. Subsumed by `go-test-integration` over a superset of paths. |
| `governance-check.yml:docx-v2-isolation` | Fires only when the PR *title* contains "docx-v2" or the branch starts with `feat/docx-v2-` (`governance-check.yml:69`). A guard that runs only when the PR self-identifies is not a gate. |
| `e2e-coverage-gate.yml` checkbox job | Its own header says "Treat this gate as a reminder, not a coverage guarantee". A control that documents itself as not a control has settled the question. The invariant→spec-ID grep beside it is real and moves to `governance`. |
| `supply-chain.yml:dependabot-label` | Applies a `needs-staging-soak-7d` label. No staging environment exists (`smoke.yml` says so) and nothing enforces the label. |

Two whole workflows die with no successor: `phase3-hardening-gate.yml` (D-17/D-19 — two of four
steps duplicate other checks, one targets a suite deleted in `dc0572f6`, the fourth disables its
only vulnerability scanner by default parameter) and `release-readiness.yml` (a
`workflow_dispatch` wrapper around a PowerShell script writing to `non_git/` — an operator
runbook wearing a workflow costume).

One control is **broken, not inert, and must be fixed rather than deleted**:
`e2e-coverage-gate.yml:156` runs `go run ./cmd/api`. That path does not exist; the binary is
`apps/api/cmd/metaldocs-api`, which `perf.yml:46` in the same repository uses correctly. This job
has not booted a backend since the layout change. It moves to `nightly.yml` with the path fixed.

### 4.6 The two Go lint stacks

`.golangci.yml` enables `staticcheck`, `govet` and `gosec`, so the repository runs staticcheck
twice (golangci diff-scoped via `only-new-issues: true`; the registry's `staticcheck` whole-tree)
and `go vet` twice. Not strictly redundant — different scoping — but two overlapping stacks with
different scope is the same "two controls, one claim" shape as §3's boundary pair, and draft 2
left `golangci-lint` unplaced entirely.

Resolution: **golangci-lint becomes the single Go lint umbrella**, whole-tree, with
`only-new-issues` dropped, and the registry's standalone `staticcheck` ID is deleted. This is
conditional on the whole tree being clean at whole-tree scope, which is **not verified** and must
be measured in Phase 0 — dropping `only-new-issues` is exactly the kind of change that surfaces a
pre-existing backlog. If the tree is not clean, the finding is real work, not a reason to keep the
diff-scoping. Deleting golangci-lint instead is not an option: it carries unique linters
(`errcheck`, `nilerr`, `exhaustive`) that nothing else runs.

### 4.7 What is not built

The consul-style `verify-ci.yml` no-op canary is **not** adopted as a required check. If `ci.yml`
never dispatches, `required` never reports, stays pending, and blocks the merge — the canary
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
This must land *before* `required` is promoted to required.

**`verify --audit` must read YAML.** The reverse direction — a workflow job that runs checks
outside the registry — is currently unenforced, which is how ten bypassing workflows went
unnoticed. `--audit` parses `.github/workflows/*.yml`, and fails when a job's `--only=` set
disagrees with the registry's `CIJob` mapping, or when a job named in the gate's `needs:` does not
exist. Until it parses YAML, the audit is a slogan.

## 6. Ruleset changes

**Required contexts: 21 → 1** (`required`).

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

The three bullets below were re-measured on 2026-08-07 while writing the Phase 0 plan, and each
one moved. They are restated here as measured, not as first estimated. A spec whose enumeration
drifts from the tree is the same meta-defect §9 names.

- **D-13 — 10 real violations, not 13.** `bash scripts/check-test-discipline.sh` reports 13, but
  three are the checker itself: R1/R3/R4 grep raw lines, so a Go comment *describing* SQL
  (`// … the INNER JOIN documents in the pre-check dropped`) is reported as a violation. Those
  three are at `read_service_template_area_integration_test.go:9,83` and
  `read_service_worklist_subject_generic_integration_test.go:99`. The fix is to make the checker
  skip comments — rewording prose to satisfy a grep would be bending the file to fit the rule
  instead of fixing the rule. The 10 real ones are R1 ×8, R4 ×1, R3 ×1 (not R4 ×4).
- **D-5 — 9 missing pages plus 1 wrong heading.** `tenant_plans.md` exists; its heading does not
  carry the qualified name the coverage checker demands. The 9 genuinely missing:
  `audit_export_jobs`, `materialize_dispatch_outbox`, `tenant_keys`, `tenant_lifecycle_jobs`,
  `token_dictionary_entries`, `approval_delegations`, `approval_review_verdicts`,
  `approval_route_stage_selectors`, `release_generations`.
- **D-1 — not a traceability gap.** "Give each of the 4 MUST REQs a live test to cite" was the
  wrong instruction. Three of the four describe a system MetalDocs did not build:
  REQ-AUTHN-1 demands Argon2id and the code uses bcrypt
  (`internal/modules/auth/infrastructure/postgres/repository.go:544` writes
  `password_algo = 'bcrypt'`); REQ-AUTHN-3 demands RFC 8725 token handling in a tree with zero
  JWTs; REQ-SEARCH-1 demands a derived rebuildable index where search is live-table escaped
  `ILIKE`. The fourth, REQ-SEC-3, is a process requirement ("OWASP ASVS is the review checklist")
  that `req-trace`'s test/commit evidence model structurally cannot express.
  `wiki/architecture/req-trace-map.yaml`'s own header declares the anti-gaming boundary; adding
  entries to make these go green would game the gate the map exists to protect. Phase 0 therefore
  produces a **written disposition, and `req-trace` stays red** pending operator rulings.
  REQ-AUTHN-1 in particular must be closed by implementing Argon2id, never by amending the REQ to
  accept bcrypt — that is relaxing a security requirement to match the code.
- Diagnose why `full` is red. Never diagnosed; it is a precondition, not an afterthought. The
  diagnosis must separate real failures from `missingInfra` skips (§5), because a skip exits 0.
- Measure golangci-lint whole-tree with `only-new-issues` off (§4.6). The result sizes real work;
  it does not license keeping the diff-scoping.

Phase 0's exit criteria are `test-discipline`, `db-dictionary`, and `cilint` green; `req-trace`
red **with a written disposition**; and both measurements recorded. `req-trace` staying red is not
a Phase 0 failure — it was already required and already red, and Phase 0 replaces an undiagnosed
red with a diagnosed one.

### 8.1 The `cilint` ownership map — Phase 0's first task, and the cheapest

`tools/cilint/baseline.json` suppresses 102 findings (`hgcrossmodule` 101, `platformboundary` 1).
That number is misleading, and the misleading part is fixable in hours.

`hgcrossmodule` scans SQL string literals for `FROM`/`JOIN <table>` and compares the table's owner
against the reading module, using the hardcoded census at `hgcrossmodule.go:23`. That census
reads:

```go
// documents (incl. the approval sub-context)
"approval_instances":       "documents",
"approval_routes":          "documents",
"approval_route_stages":    "documents",
"approval_stage_instances": "documents",
"approval_signoffs":        "documents",
```

The comment states the model it encodes: approval as a *sub-context of documents*, which was the
ADR 0072 world. **ADR 0082 promoted approval to a first-class top-level module — the 15th — and
the census was never updated.** So the analyzer accuses the approval module of cross-module
access when it reads its own tables. Those five tables account for roughly 68 of the 101 findings.

Schema ownership cannot contradict this: the schema is a single baseline file
(`db/baseline/0001_current_schema.sql`), so there are no per-module migrations to move. Ownership
is expressed *only* in this map, and ADR 0082 is the fact it is supposed to encode.

This is the same meta-defect the 2026-07-03 architecture review named: **hand-synced
enumerations**. An ADR changes, an enumeration does not follow, and the control silently begins
measuring a world that no longer exists. It is the identical failure class as §9 of this document,
which is why §9 is not decoration. The stale `hgExempt` path in the table below is a second
instance in the same file — and `scripts/check-test-discipline.sh:59` reconciled that exact rename
on 2026-07-06, so one enumeration followed the rename and the other did not.

A third finding, measured while writing the Phase 0 plan: **34 of `baseline.json`'s 35 entries
carry the same copy-pasted reason**, attributing the debt to "M3 approval kernel extraction" —
including entries in `auth`, `iam`, and `templates` that M3 does not touch. A reason that is
pasted rather than written is not a reason, and it is what let ~68 fabricated findings sit in the
file looking like real debt. The regenerated baseline must carry per-entry reasons that are true
of that entry, or the honest string `unclassified`. Its `_doc` must also name the milestone that
deletes it; today it declares the file transitional and shrink-only while naming no such
milestone, which is an unlabelled local maximum.

Correcting the census is expected to take the baseline from 35 entries to roughly 10–12. The
residue is genuine debt in ~5 clusters:

| Cluster | Findings | Shape of the fix |
|---|---|---|
| approval → `documents` / `document_comments` / `document_revisions` base tables | ~29 | one published read-port on `documents`, consumed by `postgres_approval_repository.go` and `read_service.go` |
| approval → `controlled_documents` | 1 | likely reuse the already-published `v_cd_search_facts` view |
| templates → `audit_events` | 1 | stale exemption: `hgExempt` names `templates/repository/postgres.go`; `internal/modules/templates/repository/` no longer exists (the F9.5 rename moved it to `templates/infrastructure/`). Locate the live read by grep before rewriting the entry — if no `audit_events` read remains in `templates/`, delete the entry rather than repoint it. A permanent allowlist matching nothing is the same class of lie. |
| auth → `iam_users`, iam → `governance_events` | 2 | isolated, one small port each |
| `platformboundary`: tripwire → `iam/domain` | 1 | needs an ADR; carve into that ADR's task |

A secondary defect in the ledger itself: 34 of 35 entries carry the identical copy-pasted reason
blaming "M3 approval kernel extraction", including entries in `auth`, `iam` and `templates` that
M3 does not touch. A false reason on a suppression is worse than no reason — it defeats triage.

**The deleting milestone, named as CLAUDE.md requires.** The baseline names its global-maximum end
state (empty entries) but names no milestone that deletes it, which under the local-maximum rule
makes it a defect today rather than a sanctioned transitional. The milestone is
**M3-final: cross-module SQL closure**, starting when approval HTTP (#19) merges, with three
ordered deliverables: (1) re-derive the census for the ADR 0082 world and fix the stale assertion
at `hgcrossmodule_test.go:51`; (2) port the residue to published read-ports; (3) delete
`baseline.json` and make `--update-baseline` refuse to write a non-empty entries array — turning
"shrink-only" from a comment into a mechanism.

The same treatment applies to the 27-file allowlist in `check-css-token-discipline.sh:21`, which
says "MUST only shrink" with nothing enforcing shrinkage.

Neither list blocks this CI restructure. Both must carry a named deleting milestone before the
restructure ships, because shipping an unlabelled local maximum is itself the defect.

**Phase 1 — `verify` fitness.** Land `--require-infra` and the YAML-parsing `--audit`. Neither the
gate nor any promotion happens before these exist.

**Phase 2 — build alongside.** PR A adds `ci.yml`, `nightly.yml`, `.coderabbit.yaml`, SHA-pins
every action, and adds `concurrency` groups. It deletes and renames **nothing**. It merges under
the existing 21 required contexts.

**Phase 3 — observe.** `required` must be seen green on a real PR before it is named anywhere
in the ruleset.

**Phase 4 — swap the ruleset, in two atomic API edits.** First: add `required` while retaining
all 21 old contexts. Then, once every open PR's latest SHA carries the new context: remove the 21
while retaining `required`, and flip `strict_required_status_checks_policy` to `true`.
Enforcement is never disabled at any point. There is no window in which `main` is unprotected and
no window in which every PR is stuck pending.

**Phase 5 — delete and rename.** PR B removes the superseded workflows, applies the renames in
§8.2, resolves §4.6, and re-exports the live ruleset to `.github/rulesets/main.json` with a
corrected README (which currently says 22 required checks while the JSON contains 21 — the drift
this whole design exists to stop).

Renames are safe only here. Once `required` is the sole required context, check IDs and job names
are internal and a rename cannot deadlock anything. Renaming earlier is the deadlock.

### 8.2 Renames

Convention: lowercase kebab-case; workflow files named by lifecycle stage; jobs named by concern;
checks named by **the claim they verify, not the tool that verifies it** — except where the tool
name *is* the universally understood claim (`gofmt`, `staticcheck`, `eslint`).

| Old | New | Why |
|---|---|---|
| `cilint` | `arch-lint` | named for a homegrown binary; the claim is that the architecture analyzers hold |
| `problem-codes-fresh` | `problem-codes-drift` | joins the drift family; "fresh" names the desired state, "drift" names what fails |
| `api-lint-strict` | `api-lint` | once `-base-path-v1` is deleted, strict is the only mode that runs |
| `api-lint-base-path-e2e` | `api-lint-e2e-base-path` | subject first |
| `module-boundaries` | `module-imports` | **misleading today**: the script reads Go import paths only; the invariant it claims covers SQL, which it cannot see (§3) |
| `test-discipline` | `test-conventions` | "discipline" is moralistic house vocabulary |
| `wiki-tally` | `wiki-debt-tally` | "tally" alone is opaque |
| `db-dictionary` | `db-docs-coverage` | "dictionary" is house vocabulary; the claim is documentation coverage |
| `fe-eslint` | `eslint` | tool = claim |
| `css-token-discipline` | `css-tokens` | as above |
| `docx-v2-typecheck` / `-build` / `-test` | `docx-typecheck` / `docx-build` / `docx-test` | there is no v1; the repo ships a `legacyvocab` analyzer while its own CI names carry a dead epoch suffix. Cascades to the npm scripts — coordinate, do not block on it |

The remaining 14 IDs keep their names.

Workflow-file names to kill, beyond the merge itself:

- **`ci.yml` whose `name:` is "docx-renderer CI"** — the file every external contributor opens
  first, named by the universal convention, containing 5% of the CI. The new `ci.yml` takes the
  name it always implied.
- **`phase3-hardening-gate.yml`, `release-readiness.yml`** — "phase3" is a dead internal program
  epoch baked into a public filename. Both are deleted outright anyway (§4.5).
- **`invariants.yml` / "CI Invariants"** — describes nothing a reader can predict.
- **`module-boundaries.yml:conformance`** — "conformance" to what? It is a required context name
  today carrying zero information.
- **the job display name `gofmt + go vet + staticcheck`** — a name that is a list goes stale the
  day the list changes, and it already omits `go-vet-integration`, which the job runs.

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

---

## 11. Amendment — 2026-08-08, after Phase 0

Phase 0 executed (branch `ci/a1-verify-single-entry-point`, 31 commits). It did what it was for:
it replaced estimates with measurements. Two of this spec's premises did not survive that, and
the first real CI run against the branch produced evidence about the report tier that changes
Phase 1's contents. Recorded here rather than silently absorbed, because a spec that quietly
absorbs a falsified premise is the §9 decay mechanism operating on this document.

### 11.1 golangci-lint cannot be a blocking gate in this restructure

§8's Phase 0 bullet said to measure whole-tree golangci-lint because "the result sizes real work".
It was measured. The result is larger than the sentence anticipated:

| Scope | Findings |
|---|---|
| Configured scope | **1078** |
| Whole tree | **1173** |

`revive` (582) and `errcheck` (267) are 79% of it; `documents` is the top package at 291.

Two corrections follow. First, the number previously carried in the Phase 0 planning notes — 214 —
was not a count. It was golangci-lint v2's **default output cap** (`max-issues-per-linter: 50`,
`max-same-issues: 3`), which truncates with no `+N more` indicator. Any future measurement of this
tool must pass `--max-issues-per-linter=0 --max-same-issues=0` or it is reading a ceiling and
reporting it as a total.

Second, and worse for the plan: **`apps/worker` and `apps/jobs` have zero lint coverage in CI
today** — two of the four binaries. Turning golangci-lint into a required blocking gate is
therefore not "tightening an existing gate", it is discovering two unlinted binaries at the moment
the gate goes live.

**Amended:** golangci-lint stays in the report tier for the duration of this restructure and is
**not** promoted in Phase 4. Reducing 1078 to a gateable number is real work with its own
sequencing, and it is not this design's work. §4.6's resolution of the dual Go lint stack is
unaffected — that is about which linter owns which rule, not about promotion.

### 11.2 `--require-infra` is confirmed in code, not proposed

§5 argued for `--require-infra` from the shape of the runner. The mechanism is now located
exactly: `tools/verify/main.go:341` returns SKIP when infra is missing, and `report()` at
`:423-466` returns non-zero only when `len(failed) > 0`. A SKIP therefore exits 0.

**A job promoted to `required` can report green over zero executed tests, with no code change and
nothing in its output that looks wrong.** `test-integration` is the only check in the `full`
profile and not in `pr`, which is precisely the check this would silently hollow out.

Phase 1 keeps `--require-infra` unchanged in substance; the justification is upgraded from
inference to `file:line`.

### 11.3 Three inherited report-tier reds, named

PR #96's first full run against the Phase 0 branch: 24 pass, 5 fail. One failure was Phase 0's own
(`check`/governance — a false positive fixed at the rule, see §11.5). One is red by design
(`gate`/req-traceability — the four dispositioned MUSTs). The remaining three are inherited, and
none of them is a required context, so none has ever blocked a merge. They are named here because
Phase 0 existed to end accepted red without an owner, and inheriting three unnamed would reproduce
exactly that.

**R-1 — `hardening` runs tests that were deleted.** `scripts/contract-baseline.ps1:42` invokes
`go test ./tests/contract`; the directory contains only `.gitkeep`, the contract tests having been
removed in `dc0572f6`. It fails with `no Go files in .../tests/contract` and has failed loudly for
long enough that the failure is ambient. This is §9's decay mechanism already realised in another
control: a check measuring a world that no longer exists. Phase 1 resolves it in one direction or
the other — restore the contract tests, or delete the step and the directory. Leaving a third
option open ("someone will look at it") is what produced the current state.

**R-2 — `E2E smoke` has no stack.** `webServer` exits 1 with `DATABASE_URL` empty.

**R-3 — `Perf suite (reduced — PR gate)` has no stack.** The API refuses to boot:
`METALDOCS_ATTACHMENTS_SIGNING_SECRET is required for provider local`. The workflow's own comment
concedes the position — *"Standing up that stack is the perf axis's work; A1's job is to stop the
failure from masquerading as a passing measurement."*

R-2 and R-3 share a cause with §11.2: this repository's CI has no runnable application stack, and
every control that needs one is either red or skipping. That is one problem with three faces, not
three problems.

### 11.4 Tier placement for E2E and perf, decided on industry practice

The operator asked directly whether running E2E and perf in GitHub CI is what large engineering
organisations do. The answer differs for the two, and the difference is a design decision this
spec had not made explicitly.

**E2E — yes, staged; never full-breadth blocking on every PR.** The standard shape is a small
smoke subset presubmit, the full suite on merge queue or post-merge, and the remainder nightly.
Google's small/medium/large/enormous size taxonomy encodes the same rule: large and enormous tests
do not run presubmit by default. The reason is behavioural rather than technical. E2E is the
flakiest tier by construction, and a blocking gate that fails on flake teaches the team to press
re-run — after which they re-run *every* gate. A flaky required check does not merely fail to
guard its own subject; it degrades the signal of every other required check beside it. This is the
same failure this design's own §4.3 exists to prevent from the other direction.

**Perf — never a PR gate on shared runners, stack or no stack.** The check is currently named
`Perf suite (reduced — PR gate)`. Fixing R-3's stack would make it *run*, and it would still be
wrong: GitHub-hosted runners are shared, unpinned and noisy-neighbour prone, so run-to-run variance
on an identical commit exceeds the regression such a gate is meant to catch. It measures which
machine it landed on. Industry practice puts perf on dedicated pinned hardware, or on statistical
microbenchmark comparison (`benchstat` in Go — significance, not raw numbers), or on a nightly
trend that alerts rather than blocks.

**Amended:** perf leaves the pre-merge tier permanently and belongs in `nightly.yml` (§4.4) as an
alerting trend. It is not a candidate for `required` in Phase 4 or afterwards. E2E keeps a
presubmit smoke and moves its breadth to nightly. R-2 and R-3 are therefore not "restore these two
checks to the gate" — they are "put each one in the tier where its result means something."

### 11.5 One Phase 0 defect worth carrying forward

`scripts/check-governance.ps1` failed the branch by flagging three static fixture files under
`scripts/testdata/` as an infra/ops change requiring a runbook edit. The rule excludes what is
*not* ops (`check-`, `api-lint/`, `req-trace/`), so every new non-ops path under `scripts/` must be
remembered into that list — the hand-synced-enumeration class again, inside a governance control.
Fixed by adding `testdata/` as a **category** (Go's reserved, toolchain-ignored directory: a path
under it is test input by definition) rather than as a fourth remembered name, and proved in both
directions. The inverted enumeration remains and will drift again; it is filed in
`docs/engineering/mechanical-enforcement-register.md`.

### 11.6 Out of this design's scope, stated so it is not lost

ME-14: `approval_route_stage_selectors` declares `tenant_id uuid NOT NULL` and enables no RLS —
38 tenant-scoped tables in the baseline, 37 with RLS. None of the repository's four RLS controls
could have caught it; all four take "the table has RLS" as their premise, and `tools/verify`'s 29
checks contain no RLS check at all. The firing mechanism is a set-equality drift test between
`{tables with tenant_id}` and `{tables with ENABLE ROW LEVEL SECURITY}`.

It is deliberately **not** queued behind this restructure. It is security work that would merely
*use* CI as its trigger, it is cheap, and the table is wired into the tenant export and erasure
paths. Placing it in the CI queue would make it wait five phases for no reason.
