---
id: work-ai-dialog
kind: work
status: active
owner: architecture
summary: Temporary Lead and Fable review record for the repository documentation and agent-context governance decision.
---

# AI dialogue

> **TEMPORARY / NON-AUTHORITATIVE / DELETE BEFORE MERGE**

## Review context

```text
Repository: developmentconexus-ops/MetalDocs
Branch: docs/repository-information-architecture
PR: #132
Expected review HEAD: REVALIDATE REMOTELY BEFORE REVIEW
Current PR purpose: decide repository documentation and agent-context governance only
Product implementation: not authorized
Legacy deletion: not started
PR #131: frozen provenance only
```

Review target:

```text
docs/development/documentation.md
```

Supporting non-authoritative work:

```text
docs/work/current/proposal.md
docs/work/current/plan.md
```

Canonical engineering method and independent-review workflow:

```text
developmentconexus-ops/conexus-methodology/METHOD.md
developmentconexus-ops/conexus-methodology/README.md
```

Current MetalDocs repository files are evidence of present failure modes and current tool consumers; their existing documentation shape is not target authority.

## Review request

Perform one independent adversarial review of the proposed repository documentation profile.

Attack the following questions:

1. Does selecting one `docs/` root materially reduce authority ambiguity, context bloat, and Git conflict risk compared with retaining `wiki/` + `docs/`?
2. Are semantic filenames, frontmatter, an intent-based index, and explicit MkDocs navigation the smallest sustainable information architecture?
3. Does the proposal preserve accepted Product/R10 truth while deleting process/history artifacts from the live tree?
4. Is Git/closed-PR history sufficient as the archive, or is any named current consumer missing?
5. Is the proposed `AGENTS.md` model small enough for routine LLM orientation without hiding load-bearing authority?
6. Does the one-proposal/one-AI-dialog lifecycle eliminate review-artifact bloat without weakening independent challenge, Lead adjudication, or operator ratification?
7. Is one coherent ratifiable gate per PR the correct unit, and are S0/G0/G1/T8-E boundaries coherent and merge-safe?
8. Does the execution plan accidentally let a Writer invent a material Product/R10 decision during consolidation?
9. Is the allowlist deletion rule strong enough to remove legacy while preserving current runtime safety rails and runbooks?
10. Is the proposed docs-hygiene verifier structurally enforceable with the current Go verifier/negative-fixture spine?
11. Does any proposed mechanism add unnecessary framework/tooling complexity, especially MkDocs, Goldmark, frontmatter, or PR-draft-aware checks?
12. Would this profile transfer cleanly to Marketplace Central and other Conexus products without centralizing their product truth?
13. What can be removed from the proposal or plan without weakening a distinct material property?
14. Is there a materially better Global Maximum?

Required classification for each material finding:

```text
BLOCKER | MAJOR | LOW
claim
repo/source evidence
root cause
property at risk
smallest correction
upstream/product reopen required: yes/no
```

Required primary verdict — exactly one:

```text
APPROVE REPOSITORY DOCUMENTATION PROFILE
APPROVE REPOSITORY DOCUMENTATION PROFILE WITH MATERIAL FIXES
DO NOT APPROVE REPOSITORY DOCUMENTATION PROFILE
```

Explicitly report:

```text
Global Maximum confirmed: yes/no
one docs root confirmed: yes/no
naming/navigation model confirmed: yes/no
agent-context model confirmed: yes/no
AI-dialog/Fable lifecycle confirmed: yes/no
PR lifecycle confirmed: yes/no
allowlist deletion safe: yes/no
execution plan implementable: yes/no
another review round materially required: yes/no
Lead adjudication may proceed: yes/no
```

## Fable review

**Reviewer:** Fable — independent adversarial challenger. Non-authoritative review input.

**Method applied:** `developmentconexus-ops/conexus-methodology/METHOD.md` v1.0.0 (ACCEPTED) and the "Standard Fable review workflow" in that repository's `README.md`.

**Remote HEAD revalidated, not trusted from the handoff:** `git ls-remote origin refs/heads/docs/repository-information-architecture` → `8eb2e70d11917362669f279f5183ae8366759e99`. Matches the stated expected HEAD. Base `origin/main` = `7f5b8928cc5a13feb8ee3fa7c8ceb1c7d3655a18`. Branch delta vs base: 4 added files, 2,629 insertions, 0 deletions, no code/schema/contract files.

**PR #131 provenance revalidated:** head `d8b1c6d31e704e9552a14faa7764c634a29b081d` — matches the plan's pinned SHA — with 1,018 changed files, +39,173 / −128,852, so the proposal's §2 statistics are accurate. All 17 R10 source paths the plan maps exist at that SHA (`launch-v1-*`, `r10-t1` … `r10-t8d`, `r10-technical-architecture.md`). The parity **base** is sound; BLOCKER-2 below attacks the parity **proof**, not the base.

**Evidence discipline:** current `main` files are treated as evidence of present consumers and failure modes, never as target authority (METHOD §3, Evidence). Every mechanical claim below was executed in this repository at the reviewed HEAD.

---

### 1. Summary

The core decision survives adversarial attack. One `docs/` root, semantic filenames, five-field frontmatter, an intent index, explicit navigation, one bounded `AGENTS.md`, one temporary proposal plus one temporary AI dialogue, one ratifiable gate per PR, Git as the only archive, and predicate-based deletion is the right structure, and `RESTRUCTURE NOW` is the correct METHOD outcome. Repairing indexes inside the current two-root tree would leave current truth, active work, and historical archive competing in one tree — the defect class would survive. I found no credible alternative structure that is smaller and still preserves the invariants.

The profile as written is not yet safe to promote. Three findings are blocking, and all three are specification defects rather than direction defects — each has a small, local correction that does not reopen the docs-root decision.

The load-bearing one: **the profile's own retention rule and its own root-deletion rule contradict each other on `main`'s machine-consumed documentation.** Five verification gates declare `wiki/**` subtrees as their subject, so by the profile's survival rule 1 ("a named current consumer exists") those documents must survive; by "the final live tree does not retain `wiki/`" they are deleted. The target information architecture names no home for any of them. That contradiction is invisible in the proposal and the plan because the target IA was derived from PR #131's R10 stage documents and was never reconciled against the tree it will actually delete.

---

### 2. Findings

#### BLOCKER-1 — Deleting `wiki/` removes the declared subject of five verification gates, and the target IA names no home for the documentation classes those gates read

**Claim.** `documentation.md` "Deletion and retention" requires a document to survive when it has a named current consumer. Four `wiki/` subtrees have named current consumers that are blocking or nightly-blocking checks. "Documentation root" deletes the tree they live in. "Authority and navigation", proposal §5/§7/§8, and plan Tasks 3/5/6 assign them no target. The two rules cannot both be satisfied.

**Repo/source evidence.**

| Gate | Declared subject | Where |
|---|---|---|
| `problem-codes-drift` (ProfilePR, `ci.yml:verify`) | `wiki/references/problem-codes.md` | `tools/verify/registry.go:376`; hard-coded `wikiPath = "wiki/references/problem-codes.md"` at `cmd/problem-codes-dump/main.go:54` |
| `req-trace` (ProfilePR, `ci.yml:verify`) | `wiki/architecture/` | `tools/verify/registry.go:1586`; `scripts/req-trace/main.go:37,42,45` |
| `adr-status` (`nightly.yml:governance-hygiene`) | `wiki/decisions/` — **95 tracked ADRs** | `tools/verify/registry.go:790` |
| `db-docs-coverage` (`nightly.yml:governance-hygiene`) | `wiki/database/tables/` — **76 tracked pages** | `tools/verify/registry.go:815`; the per-table `**Owner:**` line is parsed and enforced at `tools/cilint/internal/analyzers/table_ownership_parity_test.go:75,81` — "no governed doc at `wiki/database/tables/%s.md` — ownership must be governed before it is enforced" |
| `wiki-debt-tally` (`nightly.yml:governance-hygiene`) | `wiki/modules/` | `tools/verify/registry.go:804` |

Only `req-trace` is retargeted anywhere in the package (plan Task 6 Step 5). The other four appear in neither the proposal, the plan, nor `documentation.md`. Plan Task 9 Step 2 removes them with `git rm -r -- wiki`. Plan Task 5 Step 4's decision index is twelve R10 stage IDs (`product-contract` … `persistence`) — it is not, and does not claim to be, the 95-ADR corpus.

Three of the five carry `Profiles: {fast, full}` on `nightly.yml`, **not** `ProfilePR`. The plan's entire proof ladder is `--profile=pr` (Tasks 0/2/10), which is structurally blind to them: G1 would go green on the PR and break the nightly governance job after merge.

**Root cause.** The target IA was derived from PR #131's R10 stage documents and never reconciled against `main`'s machine-consumed documentation. The survival rule is a **predicate over documents**; the plan implements it as a **hand-written allowlist of roots**. Converting a predicate into an enumeration is where the hole entered.

**Property at risk.** Verification truth — five gates lose their subject, three of them silently. Governed authority for 95 ADRs and 76 table dictionary pages, both of which the profile's own naming rule anticipates (numbered ADRs are an explicit filename exception) but whose home the profile never names.

**Smallest correction.** Two clauses in `documentation.md`, no new mechanism:

1. Name machine-verified documentation as a first-class retained class and give it a home — ADRs at `docs/decisions/adr-NNNN-<slug>.md`, table dictionary at `docs/reference/database/<table>.md`, problem codes at `docs/reference/problem-codes.md`, module debt registers either homed or deleted together with `wiki-debt-tally`.
2. Add one invariant to "Deletion and retention": *no verification gate's declared subject may be deleted without retiring or repointing that gate — and re-proving its negative fixture — in the same pull request.* This generalizes past the four subtrees found here and makes the whole defect class fail loudly instead of silently.

**Upstream/product reopen required:** no. This creates documentation-placement authority inside the gate already under review; it changes no Product/R10 semantics.

---

#### BLOCKER-2 — The R10 authority-parity census cannot execute; it returns empty and reads as parity

**Claim.** Plan Task 5 Step 5 is the single control proving that no operator-ratified Product/R10 decision was lost in consolidation. As written it produces an empty file, and the reviewer reads that as "no source-only lines".

**Repo/source evidence.** Task 5 Step 1 stages the sources under `.tmp/r10-source` and states ".tmp/ remains untracked". `.gitignore:58` is `.tmp/`. Task 5 Step 5 then runs `git grep -nE '…' .tmp/r10-source`. `git grep` searches tracked content; `--untracked` still honours `.gitignore`. Executed at this HEAD against a probe file containing the literal token `MUST`:

```text
git grep -nE '\b(MUST|REQUIRE)\b' .tmp/probe             -> exit 1, no output
git grep --untracked -nE '\b(MUST|REQUIRE)\b' .tmp/probe -> exit 1, no output
grep -rnE '\b(MUST|REQUIRE)\b' .tmp/probe                -> .tmp/probe/sample.md:1: ... exit 0
```

A second, independent defect in the same step: the token set is `MUST|MUST NOT|REQUIRE|FORBID|BLOCKED|CLOSED|SELECT|REJECT`. `MUST NOT` is already subsumed by `MUST`, and the set omits `SHOULD` and `MAY` — which plan Task 6 Step 4 itself treats as normative classes to preserve, and which the current corpus uses (`wiki/architecture/req-traceability.md` totals: 62 MUST, **7 SHOULD**, 0 MAY). `SHALL`, `NEVER`, `ALWAYS`, `ONLY`, and `PROHIBIT` are also absent.

**Root cause.** A control was specified without being executed once. METHOD §3, Enforcement: "A control that cannot be shown to fire is not proven."

**Property at risk.** Proposal invariant 8 — "No operator-ratified Product/R10 decision is lost." The false green is silent and arrives exactly at the irreversible step.

**Smallest correction.** Use `grep -rnE` (or `git grep --no-index`) on the source side, keep `git grep` for the tracked target side, widen the pattern to the repository's actual normative vocabulary, and add an explicit non-empty assertion: if `.tmp/source-normative.txt` has zero lines the census failed — stop, do not read it as parity.

**Upstream/product reopen required:** no.

---

#### BLOCKER-3 — The merge-ready control cannot fire in CI; the stated proof obligation is decorative

**Claim.** `documentation.md` "Mechanical proof obligations" requires verification to fail on "temporary work files in a merge-ready PR". The mechanism the plan specifies never executes in the plan's own happy path.

**Repo/source evidence.** `.github/workflows/ci.yml:2-4`:

```yaml
on:
  pull_request:
    branches: [main]
```

No `types:` and no `push:`. GitHub's default activity types for `pull_request` are `opened`, `synchronize`, and `reopened` — `ready_for_review` is not among them, and CI never runs on `main` at all. Plan Task 8 Step 11 derives `MergeReady` from `github.event.pull_request.draft`. Plan Task 2 Step 6 and Task 10 Step 7 push the deletion commit **while the PR is still Draft** (`synchronize` → `draft=true` → `MergeReady=false` → the rule does not evaluate), then flip the PR to Ready — which produces no new run, so the required checks keep their previous green conclusion and the squash merge lands. The rule's only execution is the local, human-typed `METALDOCS_PR_DRAFT=false go run ./scripts/docs-hygiene` at Task 10 Step 6.

Corroboration in this PR's own check list: `CodeRabbit  pass  Review skipped: draft pull request` — a draft-aware consumer that has not re-evaluated and will not until an event fires.

**Root cause.** The control's trigger condition was bound to PR metadata without checking which events deliver that metadata. Same class as BLOCKER-2: mechanism specified, firing never demonstrated.

**Property at risk.** A ratifiable governance gate can merge with `proposal.md`, `plan.md`, and `ai-dialog.md` in the tree — the exact review-artifact bloat this profile exists to prevent — with the guard reporting green.

**Smallest correction.** Add `types: [opened, synchronize, reopened, ready_for_review]` to `ci.yml`'s `pull_request` trigger. State the rule state-free in `documentation.md` — *`docs/work/**` may exist only while the pull request is Draft* — so the obligation is a property of the tree and the PR state together, not of one environment variable.

**Upstream/product reopen required:** no.

---

#### MAJOR-1 — The zero-results grep rule turns a documentation reorganization into a ~198-file code, contract, and security-config PR, contradicting the plan's own constraint

**Claim.** Plan Task 9 Step 4 requires `git grep -nE '(wiki/|docs/superpowers/|docs/operator/|docs/HARNESS-PROFILE\.md|…)'`, excluding only `vendor/**` and `third_party/**`, to return **zero results**, and forbids satisfying it with stubs. Measured at this HEAD, that obligation reaches 198 tracked files far outside documentation.

**Repo/source evidence.** `git grep -l` for those paths, excluding the legacy trees themselves, `vendor/`, `third_party/`, and `docs/work/`: **198 files** — 66 `.md`, **57 `.go`** (65 of the Go hits are comment lines), **36 `.sql`** (almost all applied migrations under `archive/migrations/**`), 10 `.tsx`, 8 `.ts`, 4 `.ps1`, 3 `.sh`, 3 `.css`, plus `api/openapi/v1/openapi.yaml` (lines 4695, 4812, 7485), `.gitleaks.toml`, `.gitleaksignore`, `.golangci.yml:2`, `.coderabbit.yaml:21`, `.gitignore:84-89`, and `.github/workflows/ci.yml`.

`api/openapi/v1/openapi.yaml` is decisive: `oapi-codegen` embeds the whole spec — comments included — into each module's generated `swaggerSpec`, and **16 `api.gen.go` files** carry it. A comment-only edit to the spec therefore churns sixteen generated files and selects the contract checks.

That collides with the plan's own Global Constraint — "No product code, schema, OpenAPI, frontend, runtime, or deployment behavior changes are authorized by G0/G1" — and with Task 11 Step 1's merge assertion of the same. A 200-file, code-and-contract-touching G1 also reproduces the property the proposal cites as the reason PR #131 is unsafe (1,018 changed files). The cure would inherit the disease.

**Root cause.** The repair obligation is defined on **textual occurrence** instead of on **consumer class**. A Go comment citing a design document is a provenance citation, and the profile's own rule is that Git preserves provenance — so rewriting it serves no invariant.

**Property at risk.** Gate coherence and reviewability; the one-coherent-gate-per-PR rule this same document ratifies.

**Smallest correction.** Bound the obligation by class. An **executable consumer** — any path a script, config, tool, or gate resolves at runtime — MUST be repaired or retired. A **provenance citation** — a path appearing only in a source comment, an applied migration, a generated artifact, or a history-pinned allowlist — is out of scope by the same "Git is the archive" rule and is left alone. The gate then reads "zero executable consumers", proven by a census that classifies each hit rather than counting them.

**Upstream/product reopen required:** no.

---

#### MAJOR-2 — Zero-results would delete live secret-scan allowlist entries pinned to history

**Claim.** `.gitleaksignore` contains commit-pinned fingerprints whose paths live under `docs/superpowers/**`. Those files leave the tree; the commits do not leave history. Removing the entries to satisfy the grep re-arms findings that a merge-blocking, history-walking security gate will then report.

**Repo/source evidence.** `.gitleaksignore:19,24,31` — e.g. `21f52f29…:docs/superpowers/reports/2026-07-22-qa3-browser-qa-evidence.md:generic-api-key:40`. The `security` job's checkout in `.github/workflows/ci.yml` sets `fetch-depth: 0` with the comment "secret-scan walks full git history; on a shallow clone it would report a green it did not earn". `secret-scan` runs at `ci.yml:260` via `--require-infra --ci-job=ci.yml:security --profile=pr` — merge-blocking.

**Root cause.** Same as MAJOR-1: an occurrence-based rule applied to an artifact whose subject is history rather than the tree.

**Property at risk.** A security gate — and the standing temptation to "fix" it by weakening the scan, which `AGENTS.md` and the plan's own stable rules forbid.

**Smallest correction.** Name history-pinned security allowlists explicitly as out of scope for path repair in `documentation.md`'s deletion rule. One sentence.

**Upstream/product reopen required:** no.

---

#### MAJOR-3 — The generated requirement report and the frontmatter mandate are mutually unsatisfiable

**Claim.** Plan Task 6 Step 5 moves the generated REQ traceability report to `docs/reference/requirement-traceability.md`, and Task 4 Step 1 lists it in `mkdocs.yml` nav. `documentation.md` requires frontmatter on every maintained page under `docs/`. The generator does not emit frontmatter, and its staleness gate forbids hand-editing the file.

**Repo/source evidence.** `scripts/req-trace/report.go:81` emits `# REQ Traceability Report (generated)` as the first line — no frontmatter block; the current artifact at `wiki/architecture/req-traceability.md` confirms it. `scripts/req-trace/main.go:64`: `STALE REPORT: committed … does not match a fresh regeneration.` `req-trace` is ProfilePR / `ci.yml:verify` (`tools/verify/registry.go:1575-1586`). Hand-adding frontmatter makes the page permanently stale and fails `req-trace`; omitting it fails `docs.frontmatter`. Two blocking gates, no satisfying state.

**Root cause.** The frontmatter rule was written for hand-maintained pages and never quantified over generated pages, of which this repository already has at least one inside the target tree.

**Property at risk.** G1 cannot reach green; the likely field fix is an ad-hoc exemption, which reintroduces an unnamed hole in the metadata rule.

**Smallest correction.** `documentation.md` names generated pages as a class: their frontmatter is emitted by their generator, the generator is the owner, and the page is never hand-edited. The plan adds the emit step to `report.go` before the move.

**Upstream/product reopen required:** no.

---

#### MAJOR-4 — The promoted authority merges with dangling references and no durable ratification record of its own gate

**Claim.** As specified, `docs/development/documentation.md` lands on `main` pointing at two files the same plan deletes, and the ratification that created its authority survives nowhere in the repository.

**Repo/source evidence.** `documentation.md` "Related documents" cites `docs/work/current/proposal.md` and `docs/work/current/plan.md`. Plan Task 2 Step 5 runs `git rm` on both, and states the expected pre-merge delta is `A docs/development/documentation.md` alone. Both citations are inline code rather than Markdown links, so the planned `docs.link` rule (a Goldmark `*ast.Link`/`*ast.Image` walk, Task 8 Step 6) will not catch them. Separately, `docs/decisions/index.md` — the page `documentation.md` designates as the home for "decision identity, status, authority link, provenance, and reopen trigger" — is not created until G1 Task 4, so G0's own decision has no provenance row anywhere. With a squash merge (Task 2 Step 6) the branch commits carrying the Fable review, the Lead adjudication, and the operator ratification survive only under GitHub's `refs/pull/132/head`, which an ordinary clone does not fetch.

METHOD §3, Complexity law: YAGNI "MUST NOT remove … evidence/provenance".

**Property at risk.** Auditability of the act that created the authority — and the profile's first claim about itself.

**Smallest correction.** Replace the two "Related documents" rows with provenance — PR #132, ratification date, review reference — and add one rule to `documentation.md`: the pull request that promotes an authority writes that authority's provenance row into `docs/decisions/index.md` in the same PR. For G0, whose registry does not exist yet, carry the row in `documentation.md`'s own provenance line and move it at G1.

**Upstream/product reopen required:** no.

---

#### MAJOR-5 — Load-bearing current-runtime operating rules are removed with no successor owner, and the profile's closed `AGENTS.md` list forbids keeping them

**Claim.** Today's `AGENTS.md` and `CLAUDE.md` own repository-specific safety rules and current-system orientation that appear in no target page in the package. `documentation.md` "Agent context" is a **closed** enumeration of what `AGENTS.md` may contain, and none of this content fits it.

**Repo/source evidence.** `AGENTS.md` currently owns the truth hierarchy (runtime → contract → wiki → execution), eight required mismatch classifications, the prerequisite hard stop, and the critical-contradiction stop rule with its trigger list (route ownership/prefix, plan or prerequisite status, startup instructions, module ownership, API contract expectations, verification expectations). `CLAUDE.md` owns current-system facts: four binaries, fifteen bounded-context modules, capabilities-never-roles authorization, the fixed middleware chain (`CLAUDE.md` cites `apps/api/cmd/metaldocs-api/chain.go:25`), contract-first route change, pooled multi-tenancy, transactional outbox, DB-enforced invariants — plus the Context Map row routing to `docs/engineering/defect-class-catalog.md`. Plan Task 7 Step 1 replaces `AGENTS.md` with a bootstrap containing none of it; Step 2 reduces `CLAUDE.md` to a pointer with "no independent methodology, architecture, status, workflow, or next-step authority". Plan Tasks 3, 5, and 6 map no target for any of it. `METHOD.md` §4's `STOP / SPLIT PREREQUISITE` absorbs the generic shape but carries none of the repository-specific triggers, and no method owns MetalDocs' truth hierarchy.

METHOD §3, Complexity law: YAGNI "MUST NOT remove a known invariant, safety property, required isolation/recoverability/auditability".

**Root cause.** The proposal treats `AGENTS.md` as pure routing. It is currently routing **plus** a small body of repository engineering law, and the reduction was scoped as if only routing were present.

**Property at risk.** The prerequisite-stop discipline — the rule that halts feature work on an untrustworthy baseline — and current-runtime orientation for every agent after G1.

**Smallest correction.** `documentation.md` names one durable owner for repository engineering rules and one for current-system orientation (`docs/development/engineering-rules.md` and `docs/architecture/overview.md` are the natural slots in the IA as drawn), and the closed `AGENTS.md` list gains a routing entry to them. The plan adds the corresponding mapping rows so the content moves rather than evaporates.

**Upstream/product reopen required:** no.

---

#### MAJOR-6 — The plan's proof ladder does not mirror the CI ladder

**Claim.** The plan proves G0/G1 with `go run ./tools/verify --profile=pr` alone. CI runs three different invocations, all with `--require-infra`. The plan's command can report green over checks that silently skipped, and never exercises the job that will lint the new Go package.

**Repo/source evidence.** `ci.yml:109` `--require-infra --profile=changed --ci-job=ci.yml:verify`; `ci.yml:260` `--require-infra --ci-job=ci.yml:security --profile=pr`; `ci.yml:302` `--require-infra --only=golangci-lint`. The `security` job's own comment: "without it a missing docker daemon turns vuln-scan into a SKIP and the job exits 0 over an unrun scan." The plan uses bare `--profile=pr` at Task 0 Step 5, Task 2 Step 6, and Task 10 Step 7; only Task 0 Step 5 hedges ("or a loud infrastructure limitation"). Task 8 introduces `scripts/docs-hygiene/*.go` — a new Go package that `lint-go` lints whole-tree-blocking — and the plan never runs golangci-lint.

**Root cause.** The local ladder was written from the profile name rather than from the workflow.

**Property at risk.** False-green handoffs at every gate boundary, and a G1 that is locally green and red on `lint-go`.

**Smallest correction.** The plan's proof command becomes `go run ./tools/verify --require-infra --profile=pr` plus `go run ./tools/verify --require-infra --only=golangci-lint`.

**Upstream/product reopen required:** no.

---

#### MAJOR-7 — `kind: authority` with `status: active` is an undefined state, the candidate is in it, and nothing can detect it

**Claim.** The closed vocabulary admits `authority` + `active` but assigns it no meaning; the artifact under review is in that state; the plan says it should be in a different one; and no rule fires either way.

**Repo/source evidence.** `docs/development/documentation.md` frontmatter reads `kind: authority` / `status: active`, with the body note "Promote `status` to `current` only after independent review, operator ratification, and merge." Plan Task 1 Step 1 specifies creating the file with `status: current`. No plan step flips it, so plan and artifact already disagree and G0 merges an authority page in an undefined status. Plan Task 8 Step 4 validates only that `status ∈ {current, active}`; Step 7 constrains `status` only for `kind: work`. An unratified authority reaching `main` is therefore unrepresentable as a finding.

**Root cause.** `status` is doing two jobs — lifecycle class and ratification state — and the vocabulary was closed around the first.

**Smallest correction.** Pick one. Either (a) `status: active` is illegal for `kind: authority`, the promotion step flips it to `current`, and candidacy is carried by the work file that proposes it — which is the profile's own model; or (b) keep the candidate state and add the rule that fails a merge-ready tree containing `kind: authority, status: active`. (a) is smaller and composes with the subtractive finding in §4.

**Upstream/product reopen required:** no.

---

#### MAJOR-8 — `docs/runbooks/**` and `docs/engineering/**` have no disposition and are invisible to both of the plan's sweeps

**Claim.** Nine tracked pages already under `docs/` are neither mapped, retained, nor deleted, and neither sweep can see them. They would survive into a tree whose verifier requires frontmatter and navigation membership for every `docs/**/*.md`.

**Repo/source evidence.** Tracked at this HEAD: `docs/runbooks/{ci-required-gate-and-hardening, db-identity-provisioning, dockerfile-go-version-pin, release-backfill, replay-materialize-pdf-deadletters, worker-jobs-liveness-healthchecks}.md` and `docs/engineering/{defect-class-catalog, mechanical-enforcement-register, repo-audit-playbook}.md`. Only `release-backfill` is sourced (Task 6 Step 3). Task 9 Step 2 deletes `wiki`, `docs/superpowers`, `docs/operator`, `docs/HARNESS-PROFILE.md`, and one `docs/engineering/standards/…` path — not these. Task 9 Step 3's sweep is scoped to Markdown "outside `docs/**`", so it structurally cannot reach them. Task 9 Step 4's grep does not name `docs/runbooks/` or `docs/engineering/`. `docs/engineering/defect-class-catalog.md` is a named current consumer — `CLAUDE.md`'s Context Map routes to it. Separately, `scripts/check-governance.ps1:181` currently asserts `'(?m)^docs/runbooks/'`; plan Task 6 Step 7 repoints the rule, but nothing repoints the six pages.

**Root cause.** Same as BLOCKER-1: enumeration standing in for a predicate, this time on the `docs/` side.

**Smallest correction.** Drive the disposition pass from a complete enumeration — `git ls-files '*.md'` — with an explicit KEEP-with-target or DELETE decision recorded for every path, and treat an undispositioned path as a stop condition. This replaces both hand-written lists with one mechanism and closes BLOCKER-1's `wiki/` hole and this one together.

**Upstream/product reopen required:** no.

---

#### LOW-1 — "Rebase … without rewriting shared history" is self-contradictory

Plan Task 0 Step 6 instructs rebasing the already-pushed PR #132 onto the new `main` "without rewriting shared history". Rebasing a pushed branch requires a force-push, which the plan's own stable rules (Task 7 Step 1) forbid. **Correction:** merge `origin/main` into the branch, or state explicitly that a single-owner pre-merge PR branch is not shared history and a force-push there is admissible.

#### LOW-2 — Two of three `git rm` targets in Task 9 Step 2 do not exist, and `git rm` aborts atomically

`docs/operator/` is absent from the tracked tree at this HEAD, and the root-cause method lives at `wiki/standards/root-cause-global-maximum-method.md`, not `docs/engineering/standards/root-cause-global-maximum-method.md`. `git rm` fails the whole invocation on an unmatched pathspec, so the step's "record that fact and continue" prose cannot happen as written. **Correction:** `--ignore-unmatch`, and correct the path.

#### LOW-3 — Every ADR added after G1 becomes an orphan under the navigation rule

Proposal §5 includes `decisions/adr-0001-short-title.md` and the naming rule exempts numbered ADRs, but Task 4's nav carries a single `Decisions: decisions/index.md` leaf while Task 8 Step 5 requires every durable page in nav exactly once. The first ADR fails the build. **Correction:** declare `docs/decisions/` a directory exemption whose registry is its index, or require nav insertion on ADR creation and say so.

#### LOW-4 — The `.claude/settings.json` permission removal is outside the declared scope

Plan Task 7 Step 6 strips the CCD session-message MCP permissions. The declared scope is documentation and agent-context governance; a tool-permission change is neither, and no finding in the package motivates it. **Correction:** drop the step, or file it as its own trivial change.

#### LOW-5 — The `docs-hygiene` negative fixture may fail for the wrong reason

The fixture tree is `scripts/testdata/guard-fixtures/docs-hygiene/wiki/bad.md.txt` only, with `Want: ["wiki/bad.md", "docs.root"]`. If `Run` returns an `error` for the sandbox's absent `mkdocs.yml` before emitting findings, `main` prints to stderr and exits 1 — non-zero for a reason unrelated to the rule, and the `Want` strings never appear. **Correction:** give the fixture a minimal valid `mkdocs.yml` and `docs/index.md`, or have `Run` degrade a missing `mkdocs.yml` to a `docs.nav` finding rather than an error. The mechanism itself is sound: `Dir`, `ArgvOverride`, `Want`, and the `{{fixture}}` token all exist (`tools/verify/fixtures.go:54,96-138`), and the `req-trace` entry already uses the identical `ArgvOverride` + `{{fixture}}` shape.

#### LOW-6 — `mkdocs.yml` is a navigation manifest wearing a build tool's clothes

Nothing in the package builds the site: no MkDocs step in any workflow, no Python dependency, and `mkdocs` appears in exactly one tracked file at this HEAD (`docs/development/documentation.md`). The only consumer is a Go `nav:` parser. `strict: true`, `site_description`, `site_dir`, and `use_directory_urls` therefore assert a build no gate performs. This is not overengineering — a YAML list is the cheapest possible explicit navigation, and the format keeps a real future option open — but it should be stated honestly. **Correction:** one line in `documentation.md` — `mkdocs.yml` is the navigation manifest; publishing is not in scope, and the build-only keys are forward compatibility, not a current obligation.

---

### 3. Attack questions

1. **One `docs/` root vs. retaining `wiki/` + `docs/`.** Yes, materially. 568 tracked `wiki/` files and 906 tracked `docs/` files address overlapping subjects; the root-cause method alone exists at four tracked paths (`wiki/standards/…` plus three under `docs/superpowers/`). `AGENTS.md` "Boundary Routing" and `CLAUDE.md` "Context Map" are two non-identical routers for the same task→document mapping. Authority is currently inferred from filename and chronology, which is the defect itself, not a symptom of it.
2. **Semantic filenames + frontmatter + intent index + explicit nav — the smallest sustainable IA?** Yes, with one subtraction (§4). Five scalar frontmatter fields, a Markdown table, and a YAML list are the minimum that makes one-meaning-one-authority mechanically checkable. No plugin, no service, no schema registry.
3. **Preserves accepted Product/R10 truth while deleting process/history?** For the R10 stage corpus the design is sound and the base is verified — all 17 mapped sources exist at the pinned immutable SHA. For the rest of accepted truth it does not: BLOCKER-1 (95 ADRs, 76 table pages, module debt, problem codes) and MAJOR-8. And the proof that nothing was lost does not execute (BLOCKER-2).
4. **Is Git/closed-PR history sufficient as the archive, or is a named current consumer missing?** Git history is sufficient *as an archive*. Named current consumers **are** missing: the five gates in BLOCKER-1; `CLAUDE.md` → `docs/engineering/defect-class-catalog.md`; the `.gitleaksignore` history fingerprints (MAJOR-2); `.coderabbit.yaml:21` → `wiki/architecture/backend-target-architecture.md`.
5. **Is the `AGENTS.md` model small enough without hiding load-bearing authority?** Small enough, yes. Without hiding load-bearing authority, no — MAJOR-5. The routing spine (`AGENTS.md` → `docs/index.md` → status/work pointer → 1–3 authorities → code evidence) is correct and is a genuine improvement over the present two-router arrangement.
6. **Does the one-proposal/one-dialog lifecycle remove bloat without weakening challenge, adjudication, or ratification?** The challenge and adjudication mechanics are preserved and match the canonical workflow. The *record* of ratification is weakened — MAJOR-4. Fix that and the answer is yes.
7. **Is one ratifiable gate per PR the correct unit, and are S0/G0/G1/T8-E coherent and merge-safe?** The unit is correct. S0/G0/G1/T8-E are coherent as a sequence and correctly refuse to stack G1 on #132. They are **not** merge-safe as specified: G0 lands an authority page that is unreachable from any navigation (no `docs/index.md`, no `mkdocs.yml` until G1) and that the still-unmodified `AGENTS.md` contradicts by routing every task to `wiki/`, with no transitional exit recorded in the durable page — METHOD §4 requires the property protected now, why the target cannot land now, the successor, and the deletion condition. A three-line transitional note in `documentation.md` naming G1 as the successor closes that window cleanly. G1 itself is not merge-safe until BLOCKER-1, MAJOR-1, and MAJOR-8 are corrected.
8. **Can a Writer invent a material decision during consolidation?** Not a Product/R10 decision — Task 5 Step 2's "never silently reconcile a source contradiction; stop and surface it" is the right guard, and Task 5 Step 5's census is the right idea. But the inverse happens: Task 9 Step 4's zero-results requirement hands a Writer unadjudicated **deletions**. `.claude/skills/harness-hub/SKILL.md` and the harness program machinery cite `docs/superpowers/ROADMAP.md` and `docs/HARNESS-PROFILE.md`; with both deleted and no successor named, "update the consumer to the semantic target or delete the obsolete consumer" resolves to deleting the program operating model on a Writer's judgment. Same shape as MAJOR-5. The corrections in MAJOR-1 (classify consumers) and MAJOR-8 (enumerate dispositions) remove the discretion.
9. **Is the allowlist deletion rule strong enough?** No — BLOCKER-1, MAJOR-1, MAJOR-2, MAJOR-8. The *predicate* is right and is the best part of the profile; its expression as a hand-written list of roots is what fails. Note also that three of the six survival criteria (named consumer, unique meaning, clear lifecycle) are not machine-checkable under the proposed verifier while the other three are — the document should say which are review-time judgments and who adjudicates them, so the six-item list is not read as a mechanical gate.
10. **Is the docs-hygiene verifier structurally enforceable on the current spine?** Yes. `Check` (`tools/verify/registry.go:64-140`), `Fixture` with `Dir`/`ArgvOverride`/`Want`/`NotWant` and the `{{fixture}}` token (`tools/verify/fixtures.go:54-138`), the sandbox-as-git-repo harness, and the profile constants all exist, and the plan's registry entry uses them correctly. Audit rules A6 (a ProfilePR check's `CIJob` must sit inside `ci.yml:required`'s needs closure) and A7 (fixture or classified waiver) are satisfied by `CIJob: "ci.yml:verify"` and the declared `Fixture`. `gopkg.in/yaml.v3` is already in `go.mod:41`; only `goldmark` is new. Caveats: LOW-5 and the MAJOR-3 collision.
11. **Does any mechanism add unnecessary framework or tooling complexity?** MkDocs: no framework is actually adopted — LOW-6; the cost is one YAML file parsed by a Go tool. Goldmark: justified — link validation over an AST rather than a regex is the difference between a control and a guess, and the alternative is a hand-rolled Markdown parser. Frontmatter: justified, with one field removable (§4). **PR-draft-aware checks: not justified as specified** — BLOCKER-3. That is the one mechanism whose complexity buys nothing today, because it cannot fire.
12. **Does the profile transfer to Marketplace Central and other Conexus products without centralizing their truth?** Yes for the portable artifact — root, naming, frontmatter, index, navigation, work lifecycle, one gate per PR, Git as archive, deletion predicate. The *enforcement* does not transfer: `tools/verify`, the registry shape, `METALDOCS_PR_DRAFT`, and `scripts/docs-hygiene` are MetalDocs-specific, and each product must implement the rules on its own spine. Proposal §15 already says the shared profile is structure and lifecycle; `documentation.md` should say so too, in one sentence, so a second repository does not read the Go verifier as part of the standard.
13. **What can be removed?** See §4.
14. **Is there a materially better Global Maximum?** See §5.

---

### 4. Subtractive pass

Removable without weakening a distinct material property:

- **The `status` field, or its second job.** Across the entire package, `status: active` occurs only with `kind: work` and `status: current` only with the other kinds. `status` is a function of `kind` and carries no independent information — until MAJOR-7's candidate state gives it one. So: delete `status` and derive it, **or** keep it and let it mean exactly "ratified vs. not". Two fields encoding one fact is the duplicate-authority defect this document exists to remove, applied to its own metadata.
- **Two of the four `kind` values.** Nothing in `documentation.md` or the proposed verifier branches on `guide` vs. `reference`; only `work` is read (Task 8 Step 7). A closed vocabulary nothing consumes is ceremony that will generate a classification argument on every page. Collapse to `authority | work` unless a named consumer for the distinction is produced.
- **Proposal §3 "Industry-aligned basis"** and its six URLs. METHOD §1: references are evidence, never requirements, and this section changes no rule. `documentation.md` correctly already omits it.
- **Proposal §2's PR #131 statistics.** Motivational, not normative; correctly already absent from the durable page.
- **The `## Current decision` block's overlap with the sections beneath it** (`ONE PR PER RATIFIABLE GATE` vs. `## Pull-request lifecycle`, and so on). A compact normative core plus elaboration is defensible on one page; if it is kept, say that the block indexes the sections rather than restating them.

Explicitly **not** removable: `docs/status.md` as a page distinct from `docs/index.md`. A single citable stage authority that a PR body or `AGENTS.md` can reference without loading the index is a real property, not ceremony.

---

### 5. Global Maximum

**Direction: confirmed.** Comparing the credible alternatives —

- *Repair indexes inside the two-root tree* (local maximum): leaves current truth, active work, and archive competing in one tree. Preserves the defect class. Correctly rejected.
- *Keep `wiki/` as governed technical memory and `docs/` as program material*: the status quo with a rationale attached. Fails on the same evidence — the root-cause method exists in four places today under exactly this split, and two routers already disagree.
- *Publish through Backstage or a central Conexus documentation service*: adds a service and centralizes product truth; correctly declared a non-goal.
- *Chosen — one governed `docs/` root with mechanical hygiene*: the smallest structure that makes one-meaning-one-authority checkable, keeps product truth local, and stays portable.

**Specification: not yet at the maximum.** Three structural upgrades take it there, and each subtracts rather than adds:

1. **Predicate over enumeration.** Drive both retention and deletion from a complete `git ls-files` enumeration with an explicit disposition per path, and treat an undispositioned path as a stop. One mechanism replaces two hand-written lists and closes BLOCKER-1 and MAJOR-8 together.
2. **Gate-subject invariant.** No verification gate's declared subject may be deleted without retiring or repointing that gate, and re-proving its negative fixture, in the same pull request. One sentence, and it generalizes past every instance found here — it is the rule this repository's own registry design (`Check.Paths`, `Fixture`, audit rules A6/A7) already implies but has never stated for documentation.
3. **Consumer classes, not text occurrences.** Executable consumers must be repaired; provenance citations are Git-archived and left alone. This keeps G1 doc-only, which is the property that makes the gate ratifiable at all.

With those three, the profile is the global maximum for this repository and transfers cleanly as a rule set.

**Independent-challenge floor.** METHOD §3 requires an independent challenger before ratification when a decision creates or moves authority boundaries or binds multiple repositories. This decision does both. That floor is met by this review.

---

### 6. Required report

```text
PRIMARY VERDICT: APPROVE REPOSITORY DOCUMENTATION PROFILE WITH MATERIAL FIXES

Global Maximum confirmed: no
one docs root confirmed: yes
naming/navigation model confirmed: yes
agent-context model confirmed: no
AI-dialog/Fable lifecycle confirmed: yes
PR lifecycle confirmed: yes
allowlist deletion safe: no
execution plan implementable: no
another review round materially required: no
Lead adjudication may proceed: yes

BLOCKER: 3
MAJOR:   8
LOW:     6
```

The material fixes are **pre-promotion**, not post-merge. `documentation.md` should not be promoted to `status: current` until BLOCKER-1 is closed in the document itself.

**Reading of each line.**

- *Global Maximum confirmed: no.* The direction is confirmed and no better structure exists. The profile **as written** is not the global maximum, because "Deletion and retention" and "Documentation root" contradict each other on machine-consumed documentation. Flips to yes with BLOCKER-1's two clauses.
- *agent-context model confirmed: no.* The routing spine is correct; the closed content list has no slot for the load-bearing rules it displaces (MAJOR-5). Flips to yes when those rules are given a named owner.
- *allowlist deletion safe: no.* BLOCKER-1, MAJOR-1, MAJOR-2, MAJOR-8. The predicate is right; its expression as a root list is not.
- *execution plan implementable: no.* BLOCKER-2 (parity census returns empty), BLOCKER-3 (merge-ready control cannot fire), MAJOR-3 (two gates mutually unsatisfiable), MAJOR-6 (local ladder ≠ CI ladder), LOW-2 (`git rm` aborts). Every one is a bounded repair; none requires redesign.
- *another review round materially required: no* — conditional, and the condition is stated so the Lead can falsify it. All findings are defects against the candidate's own stated rules and are correctable by adjudication. **The one exception:** BLOCKER-1's first clause names homes for four documentation classes, which is new placement authority. METHOD §3 requires that a correction creating new authority return to decision rather than enter disguised as a fix. If the Lead's correction stays inside the IA as drawn (adds page homes plus one invariant), it is a decision inside this gate and needs no second challenger. If it instead reopens the root, the naming rule, or the retention predicate, Round 2 applies.
- *Lead adjudication may proceed: yes.* Nothing here blocks adjudication, and nothing here promotes anything.

**Scope statement.** This review modified only `docs/work/current/ai-dialog.md`. No durable document, code, schema, OpenAPI, frontend, runtime, deployment file, or PR metadata was touched. No legacy documentation was deleted, no candidate promoted, no PR marked ready, merged, or closed, no gate started or resumed, no history rewritten. Reviewer output is evidence, never authority, until Lead adjudication and operator ratification.

## Lead adjudication

Lead confronts every material finding here. Reviewer output is evidence, never authority.

## Bounded round 2

Use only if a real material contradiction survives Lead adjudication.

## Operator decision

Record final operator ratification or the exact bounded reopen here.
