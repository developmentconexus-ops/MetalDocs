# Repository Documentation Rebaseline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the repository's duplicated `wiki/` + `docs/superpowers/` documentation estate with one governed `docs/` root, preserve every accepted Product/R10 decision through T8-D plus the accepted T8-E checkpoint, and make documentation/agent-context drift mechanically fail.

**Architecture:** Execute three clean pull-request gates from `main`: S0 restores a trustworthy green verification baseline; G0 ratifies the repository documentation profile in PR #132; G1 performs the atomic authority consolidation, legacy deletion, path repair, and verifier activation. Only after G1 is merged does a fresh T8-E PR recreate the active proposal under `docs/work/current/`.

**Tech Stack:** Markdown, YAML, MkDocs navigation syntax, Go 1.26.6+, `gopkg.in/yaml.v3`, `github.com/yuin/goldmark`, existing `tools/verify`, GitHub Actions, Git.

**Spec:** `docs/work/current/proposal.md`

## Global Constraints

- `docs/` is the only first-party documentation root after G1.
- `wiki/`, `docs/superpowers/`, `docs/operator/`, live archives, tombstones, and completed review artifacts do not survive G1.
- Durable filenames are semantic, lowercase, kebab-case, and do not carry dates, stage codes, versions, review states, or lifecycle adjectives.
- `README.md`, `AGENTS.md`, `CLAUDE.md`, `LICENSE`, `SECURITY.md`, `CONTRIBUTING.md`, `mkdocs.yml`, and `.github/*` may remain outside `docs/` because their locations are platform conventions.
- Each durable Markdown page has unique frontmatter fields: `id`, `kind`, `status`, `owner`, and `summary`.
- Closed frontmatter values are `kind: authority|guide|reference|work` and `status: current|active`.
- `docs/work/` is temporary, excluded from MkDocs navigation, and deleted before a governance/architecture PR becomes merge-ready.
- Git and closed pull requests are the historical archive; no repository-local archive is created.
- PR #131 is provenance only and is never merged into G0 or G1.
- G1 reads Product/R10 source authority from immutable PR #131 HEAD `d8b1c6d31e704e9552a14faa7764c634a29b081d`.
- No product code, schema, OpenAPI, frontend, runtime, or deployment behavior changes are authorized by G0/G1.
- Existing safety gates remain active until an equal-or-stronger replacement is proven to fire.
- Every PR remains Draft while temporary work files exist and is squash-merged only after the operator ratifies the result.

---

## Execution Graph

```text
S0 verification baseline
        ↓
G0 repository documentation profile — PR #132
        ↓
G1 docs-only authority consolidation + deletion + verifier
        ↓
close PR #131 as superseded
        ↓
fresh T8-E executable API-contract PR
```

Do not stack G1 on PR #132. Merge G0 first, then branch G1 from the updated `main`.

---

### Task 0: Restore a Trustworthy Green Baseline in a Separate S0 PR

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`
- Verify only: `pnpm-lock.yaml`
- Verify only: `.github/workflows/ci.yml`

**Interfaces:**
- Consumes: current CI failure evidence from PR #132 run `#833`.
- Produces: a `main` commit on which `govulncheck` and high-severity dependency scanning are green before documentation-governance changes are judged.

- [ ] **Step 1: Create the prerequisite branch from current `main`**

```bash
git fetch origin main
git switch -c fix/verification-baseline-2026-08 origin/main
```

Expected: branch contains no commits from PR #131 or PR #132.

- [ ] **Step 2: Prove the current failures before changing versions**

```bash
go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
```

Expected: non-zero with reachable Go 1.26.5 standard-library vulnerabilities including fixes available in Go 1.26.6.

```bash
docker run --rm \
  -v "$(pwd):/src" \
  anchore/grype@sha256:1e71065c0a4cff3e6bd3b8add525ffac4343eb4971694eb90a31cf6d4d3e85db \
  dir:/src --fail-on high -o table
```

Expected: non-zero with `golang.org/x/mod v0.37.0` reported as high severity and fixed in `v0.40.0`.

- [ ] **Step 3: Apply the minimum version corrections**

```bash
go mod edit -go=1.26.6
go get golang.org/x/mod@v0.40.0
go mod tidy
```

Expected `go.mod` changes:

```go
go 1.26.6
```

and:

```go
golang.org/x/mod v0.40.0 // indirect
```

Do not opportunistically upgrade unrelated dependencies in this PR.

- [ ] **Step 4: Run the focused proof**

```bash
go test ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
```

Expected: both exit zero; no reachable vulnerability remains from the Go 1.26.5 findings.

```bash
docker run --rm \
  -v "$(pwd):/src" \
  anchore/grype@sha256:1e71065c0a4cff3e6bd3b8add525ffac4343eb4971694eb90a31cf6d4d3e85db \
  dir:/src --fail-on high -o table
```

Expected: exit zero for high/critical findings. If a new high/critical finding remains, stop and repair that exact dependency in S0; do not waive or lower the severity threshold.

- [ ] **Step 5: Run the repository PR profile**

```bash
go run ./tools/verify --profile=pr
```

Expected: green, or a loud infrastructure limitation recorded without weakening a check.

- [ ] **Step 6: Commit and open the S0 PR**

```bash
git add -- go.mod go.sum
git commit -m "fix(security): restore verified dependency baseline"
git push -u origin fix/verification-baseline-2026-08
```

Open one Draft PR, obtain required review, then merge it before continuing G0. Rebase PR #132 onto the resulting `main` without rewriting shared history.

---

### Task 1: Promote the G0 Documentation Profile into Durable Authority

**Files:**
- Create: `docs/development/documentation.md`
- Modify: `docs/work/current/proposal.md`
- Modify: `docs/work/current/plan.md`

**Interfaces:**
- Consumes: the approved proposal and canonical DevelopmentConexus Method v1.0.0.
- Produces: one durable repository-level documentation authority; it does not yet migrate or delete the current documentation estate.

- [ ] **Step 1: Create the durable governance page**

Create `docs/development/documentation.md` with this exact frontmatter and top-level structure:

```markdown
---
id: development-documentation
kind: authority
status: current
owner: engineering
summary: Defines MetalDocs documentation placement, naming, lifecycle, agent routing, review, and pull-request governance.
---

# Documentation and agent-context governance

## Purpose
## Current decision
## Documentation root
## Naming and metadata
## Authority and navigation
## Agent context
## Active work and AI dialogue
## Pull-request lifecycle
## Deletion and retention
## Mechanical proof obligations
## Reopen triggers
## Related documents
```

The normative core under `## Current decision` is:

```text
ONE docs/ ROOT
+ SEMANTIC STABLE FILENAMES
+ TASK-ORIENTED INDEX
+ EXPLICIT NAVIGATION
+ SHORT AGENT BOOTSTRAP
+ ONE TEMPORARY PROPOSAL
+ ONE TEMPORARY AI DIALOGUE
+ ONE PR PER RATIFIABLE GATE
+ GIT AS THE ONLY ARCHIVE
+ ALLOWLIST-BASED LEGACY DELETION
- wiki/
- docs/superpowers/
- stage/date/version filenames
- amendment chains
- permanent per-round review artifacts
- live-tree archives
- duplicated status and authority
```

Copy the accepted rules from `docs/work/current/proposal.md`; remove PR #131 statistics, implementation sequencing, and temporary discussion. Do not add a second Method summary.

- [ ] **Step 2: Add an explicit source/provenance note to the temporary proposal**

Append this section to `docs/work/current/proposal.md`:

```markdown
## Promotion target

After independent review and operator ratification, the durable result is promoted to:

`docs/development/documentation.md`

The proposal, plan, and AI dialogue are then deleted. Git and PR #132 retain their provenance.
```

- [ ] **Step 3: Run the current changed-profile checks**

```bash
go run ./tools/verify --profile=changed
```

Expected: all checks except any pre-existing infrastructure failure already isolated in S0 are green. After S0 merges and PR #132 is rebased, the expected result is fully green.

- [ ] **Step 4: Commit the durable candidate**

```bash
git add -- docs/development/documentation.md docs/work/current/proposal.md docs/work/current/plan.md
git commit -m "docs(governance): promote documentation profile candidate"
```

---

### Task 2: Run the Final G0 Independent Review in One Temporary AI Dialogue

**Files:**
- Create temporarily: `docs/work/current/ai-dialog.md`
- Modify during review: `docs/work/current/ai-dialog.md`
- Modify only if findings require correction: `docs/development/documentation.md`
- Delete before merge: `docs/work/current/proposal.md`
- Delete before merge: `docs/work/current/plan.md`
- Delete before merge: `docs/work/current/ai-dialog.md`

**Interfaces:**
- Consumes: `docs/development/documentation.md`, PR #132, canonical Method/Fable workflow.
- Produces: an independently challenged, operator-ratified durable governance page with no review artifacts in the merge tree.

- [ ] **Step 1: Create the AI dialogue file**

```markdown
---
id: work-ai-dialog
kind: work
status: active
owner: architecture
summary: Temporary Lead and Fable review record for the repository documentation profile.
---

# AI dialogue

> Temporary, non-authoritative, and deleted before merge.

## Review request

Review `docs/development/documentation.md` against `docs/work/current/proposal.md`.
Attack authority duplication, naming ambiguity, agent-read bloat, unsafe deletion, PR sizing, review lifecycle, tooling feasibility, and portability to other Conexus repositories.

## Fable review

## Lead adjudication

## Bounded round 2

Use only if a material contradiction survives adjudication.

## Operator decision
```

- [ ] **Step 2: Send the bounded Fable handoff**

Use only:

```text
Repository: developmentconexus-ops/MetalDocs
Branch: docs/repository-information-architecture
PR: #132
Expected HEAD: <revalidated remote HEAD>
Read AGENTS.md, then docs/work/current/proposal.md.
Review docs/development/documentation.md and write only in docs/work/current/ai-dialog.md.
Apply the canonical DevelopmentConexus Method/Fable workflow.
Do not modify any other file.
```

- [ ] **Step 3: Adjudicate every material finding in the same file**

For each finding record:

```text
ACCEPT | ACCEPT WITH CORRECTION | REJECT
basis
exact correction or proof
upstream reopen required: yes/no
```

If no material contradiction survives, leave `## Bounded round 2` empty except for `Not required.`

- [ ] **Step 4: Obtain explicit operator ratification**

Record exactly one outcome in `## Operator decision`:

```text
APPROVED
```

or:

```text
NOT APPROVED — <specific decision reopened>
```

Do not promote on an ambiguous response.

- [ ] **Step 5: Remove temporary work files and prove the final G0 tree**

```bash
git rm -- docs/work/current/proposal.md docs/work/current/plan.md docs/work/current/ai-dialog.md
git status --short
```

Expected PR #132 documentation delta before merge:

```text
A  docs/development/documentation.md
```

plus no unrelated files.

- [ ] **Step 6: Run required verification and commit cleanup**

```bash
go run ./tools/verify --profile=pr
git add -- docs/development/documentation.md
git commit -m "docs(governance): ratify repository documentation profile"
```

Set PR #132 ready only after green checks, then squash merge it.

---

### Task 3: Start the G1 Consolidation from the Updated Main Branch

**Files:**
- Create: `docs/work/current/index.md`
- Create: `docs/work/current/proposal.md`

**Interfaces:**
- Consumes: merged G0 authority, immutable PR #131 source commit, clean `main`.
- Produces: one active consolidation package and one Draft G1 PR; no authority is duplicated outside the temporary proposal.

- [ ] **Step 1: Create the G1 branch**

```bash
git fetch origin main
git switch -c docs/repository-documentation-rebaseline origin/main
git fetch origin d8b1c6d31e704e9552a14faa7764c634a29b081d
```

- [ ] **Step 2: Create the active-work index**

```markdown
---
id: work-current
kind: work
status: active
owner: architecture
summary: Routes the active docs-only Product/R10 authority consolidation.
---

# Current work

| Field | Value |
|---|---|
| Topic | Product/R10 authority consolidation |
| Branch | `docs/repository-documentation-rebaseline` |
| Pull request | Draft G1 PR |
| Source authority | PR #131 HEAD `d8b1c6d31e704e9552a14faa7764c634a29b081d` |
| Proposal | `proposal.md` |
| Current checkpoint | Consolidation and deletion not yet ratified |
| Next action | Build semantic docs, repair consumers, activate proof, then final Fable review |
```

- [ ] **Step 3: Create the G1 proposal with the closed source-to-target map**

`docs/work/current/proposal.md` must contain this mapping:

| Source at PR #131 HEAD | Target |
|---|---|
| `wiki/architecture/launch-v1-product-contract.md` | `docs/product/contract.md` |
| `wiki/architecture/whole-product-alignment-review.md` accepted conclusions | owning product/architecture pages |
| `wiki/architecture/launch-v1-scope-rebaseline.md` surviving scope | `docs/product/contract.md` |
| `wiki/architecture/launch-v1-ownership-topology.md` | `docs/architecture/ownership.md` |
| `wiki/architecture/r10-t1-semantic-state-invariants.md` | `docs/architecture/domain-model.md` |
| `wiki/architecture/r10-t2-governance-effectivity-transactions.md` | `docs/architecture/lifecycle.md` |
| `wiki/architecture/r10-t3-authorization-audit-enforcement.md` | `docs/architecture/authorization.md` + `docs/architecture/audit.md` |
| `wiki/architecture/r10-t3-d4-responsible-owner-eligibility-amendment.md` | merge into `docs/architecture/authorization.md` |
| `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md` | `docs/architecture/content-integrity.md` |
| `wiki/architecture/r10-t5-durable-async-search-external-effects.md` | `docs/architecture/async-and-search.md` |
| `wiki/architecture/r10-t6-canonical-api-frontend-journeys.md` | `docs/product/journeys.md`; closed principles referenced by active T8-E proposal |
| `wiki/architecture/r10-t7-historical-migration-truth-semantic-mapping.md` | `docs/architecture/transition.md` |
| `wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md` | surviving target/reuse laws in `docs/architecture/transition.md` |
| `wiki/architecture/r10-t8b-backend-module-package-topology.md` | `docs/architecture/backend.md` |
| `wiki/architecture/r10-t8c-internal-communication-contracts.md` | `docs/architecture/interfaces.md` |
| `wiki/architecture/r10-t8d-persistence-realization.md` | `docs/architecture/persistence.md` |
| Registry + all amendments through T8-D | `docs/decisions/index.md` |
| `wiki/architecture/r10-technical-architecture.md` | `docs/status.md` |
| accepted T8-E chat checkpoint | active T8-E proposal after G1 merge, not durable authority during G1 |

The proposal also records the exact deletion roots and the current-runtime documents retained by Tasks 6–7.

- [ ] **Step 4: Commit and open the Draft G1 PR**

```bash
git add -- docs/work/current/index.md docs/work/current/proposal.md
git commit -m "docs(governance): open authority consolidation gate"
git push -u origin docs/repository-documentation-rebaseline
```

---

### Task 4: Create the Minimal Docs Root and Explicit Navigation

**Files:**
- Create: `mkdocs.yml`
- Create: `docs/index.md`
- Create: `docs/status.md`
- Create: `docs/product/index.md`
- Create: `docs/architecture/index.md`
- Create: `docs/decisions/index.md`
- Create: `docs/development/index.md`
- Create: `docs/operations/index.md`
- Create: `docs/reference/index.md`

**Interfaces:**
- Consumes: G0 documentation authority.
- Produces: one navigable docs root; later tasks fill only pages with unique current meaning.

- [ ] **Step 1: Create `mkdocs.yml` with durable pages only**

```yaml
site_name: MetalDocs
site_description: Product, architecture, development, operations, and reference documentation for MetalDocs.
docs_dir: docs
site_dir: .build/docs-site
strict: true
use_directory_urls: true
nav:
  - Home: index.md
  - Status: status.md
  - Product:
      - Overview: product/index.md
      - Contract: product/contract.md
      - Journeys: product/journeys.md
      - Glossary: product/glossary.md
  - Architecture:
      - Overview: architecture/index.md
      - Ownership: architecture/ownership.md
      - Domain model: architecture/domain-model.md
      - Lifecycle: architecture/lifecycle.md
      - Authorization: architecture/authorization.md
      - Audit: architecture/audit.md
      - Content integrity: architecture/content-integrity.md
      - Async and search: architecture/async-and-search.md
      - Backend: architecture/backend.md
      - Interfaces: architecture/interfaces.md
      - Persistence: architecture/persistence.md
      - Transition: architecture/transition.md
  - Decisions: decisions/index.md
  - Development:
      - Overview: development/index.md
      - Setup: development/setup.md
      - Testing: development/testing.md
      - Verification: development/verification.md
      - Documentation: development/documentation.md
  - Operations:
      - Overview: operations/index.md
      - Restore: operations/runbooks/restore.md
      - Release backfill: operations/runbooks/release-backfill.md
  - Reference:
      - Overview: reference/index.md
      - Repository map: reference/repository-map.md
      - Configuration: reference/configuration.md
      - Current runtime requirements: reference/current-runtime-requirements.md
      - Requirement traceability: reference/requirement-traceability.md
```

Do not list `docs/work/`.

- [ ] **Step 2: Create `docs/index.md` as an intent map**

Use this table:

```markdown
| I need to know… | Read |
|---|---|
| Current project stage and implementation gate | [Status](status.md) |
| Product boundary | [Product contract](product/contract.md) |
| User/system journeys | [Product journeys](product/journeys.md) |
| Semantic ownership | [Ownership](architecture/ownership.md) |
| Domain state and invariants | [Domain model](architecture/domain-model.md) |
| Lifecycle and effectivity | [Lifecycle](architecture/lifecycle.md) |
| Authorization rules | [Authorization](architecture/authorization.md) |
| Persistence and concurrency | [Persistence](architecture/persistence.md) |
| Local setup | [Setup](development/setup.md) |
| Repository verification | [Verification](development/verification.md) |
| Current governed proposal | [Current work](work/current/index.md) |
```

The last row is removed before G1 merge because no active work survives the final merge tree.

- [ ] **Step 3: Create `docs/status.md` as the only current-stage authority**

```markdown
---
id: project-status
kind: authority
status: current
owner: architecture
summary: Defines the current MetalDocs architecture stage and implementation gate.
---

# Project status

```text
Product contract and ownership          operator-approved
Architecture through persistence        operator-ratified
Executable API contract                 active / resume in a fresh PR
Frontend/runtime/coherence/proof/cutover/planning/readiness  not open
Product implementation                  blocked
```

Current source of active work on a branch: `docs/work/current/index.md`.
Main contains no active proposal.
```

- [ ] **Step 4: Create focused folder indexes**

Each index contains only:

```text
purpose
intent → page table
what the folder does not own
```

Do not repeat architectural decisions or stage history.

- [ ] **Step 5: Commit the navigation spine**

```bash
git add -- mkdocs.yml docs/index.md docs/status.md docs/product/index.md docs/architecture/index.md docs/decisions/index.md docs/development/index.md docs/operations/index.md docs/reference/index.md
git commit -m "docs: establish semantic documentation navigation"
```

---

### Task 5: Consolidate Product and Architecture Authority Without Semantic Change

**Files:**
- Create: `docs/product/contract.md`
- Create: `docs/product/journeys.md`
- Create: `docs/product/glossary.md`
- Create: `docs/architecture/ownership.md`
- Create: `docs/architecture/domain-model.md`
- Create: `docs/architecture/lifecycle.md`
- Create: `docs/architecture/authorization.md`
- Create: `docs/architecture/audit.md`
- Create: `docs/architecture/content-integrity.md`
- Create: `docs/architecture/async-and-search.md`
- Create: `docs/architecture/backend.md`
- Create: `docs/architecture/interfaces.md`
- Create: `docs/architecture/persistence.md`
- Create: `docs/architecture/transition.md`
- Modify: `docs/decisions/index.md`

**Interfaces:**
- Consumes: immutable ratified source pages from PR #131 HEAD.
- Produces: semantic subject-based authority pages preserving the accepted meaning exactly.

- [ ] **Step 1: Materialize immutable source files locally without merging PR #131**

```bash
mkdir -p .tmp/r10-source
for path in \
  wiki/architecture/launch-v1-product-contract.md \
  wiki/architecture/whole-product-alignment-review.md \
  wiki/architecture/launch-v1-scope-rebaseline.md \
  wiki/architecture/launch-v1-ownership-topology.md \
  wiki/architecture/r10-t1-semantic-state-invariants.md \
  wiki/architecture/r10-t2-governance-effectivity-transactions.md \
  wiki/architecture/r10-t3-authorization-audit-enforcement.md \
  wiki/architecture/r10-t3-d4-responsible-owner-eligibility-amendment.md \
  wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md \
  wiki/architecture/r10-t5-durable-async-search-external-effects.md \
  wiki/architecture/r10-t6-canonical-api-frontend-journeys.md \
  wiki/architecture/r10-t7-historical-migration-truth-semantic-mapping.md \
  wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md \
  wiki/architecture/r10-t8b-backend-module-package-topology.md \
  wiki/architecture/r10-t8c-internal-communication-contracts.md \
  wiki/architecture/r10-t8d-persistence-realization.md; do
  git show d8b1c6d31e704e9552a14faa7764c634a29b081d:"$path" > ".tmp/r10-source/$(basename "$path")"
done
```

`.tmp/` remains untracked and is deleted after parity review.

- [ ] **Step 2: Apply one fixed authority-page format**

Every architecture authority page uses:

```markdown
# <Stable subject>

## Purpose
## Current decision
## Invariants
## Structure
## Boundaries
## Proof obligations
## Reopen triggers
## Related documents
```

Rules:

```text
preserve every binding law and normative enum
preserve every accepted absence/defer decision
preserve every proof obligation and reopen trigger
remove review counts, candidate history, branch names, and stage-routing prose
replace old cross-links with semantic target links
never silently reconcile a source contradiction; stop and surface it in the active proposal
```

- [ ] **Step 3: Split T3 mechanically by meaning**

`docs/architecture/authorization.md` receives:

```text
Role/Permission/RoleAssignment meaning
scope and evaluation
DomainPredicate ownership
eligibility serialization
Group/User/offboarding authorization effects
minimum disclosure rules
D4 responsible-owner eligibility
```

`docs/architecture/audit.md` receives:

```text
AuditEvent meaning
same-transaction evidence law
required event census
visibility attribution
append-only/read behavior
privacy exclusions
```

No rule is duplicated between both pages; cross-reference instead.

- [ ] **Step 4: Build the compact decision index**

Use these stable IDs:

```text
product-contract
ownership
domain-model
lifecycle
authorization
audit
content-integrity
async-and-search
migration-truth
backend-topology
internal-interfaces
persistence
```

Each row has:

```text
ID | decision summary | status | authority link | source provenance | reopen trigger
```

Provenance may mention R10/T-stage IDs; filenames may not.

- [ ] **Step 5: Run a normative-token census**

```bash
git grep -nE '\b(MUST|MUST NOT|REQUIRE|FORBID|BLOCKED|CLOSED|SELECT|REJECT)\b' .tmp/r10-source > .tmp/source-normative.txt
git grep -nE '\b(MUST|MUST NOT|REQUIRE|FORBID|BLOCKED|CLOSED|SELECT|REJECT)\b' docs/product docs/architecture docs/decisions > .tmp/target-normative.txt
```

Review every source-only line. Each must be either represented in the owning target page or explicitly documented as process metadata removed without semantic loss.

- [ ] **Step 6: Commit product and architecture authority in two reviewable commits**

```bash
git add -- docs/product docs/architecture/ownership.md docs/architecture/domain-model.md docs/architecture/lifecycle.md docs/architecture/authorization.md docs/architecture/audit.md docs/architecture/content-integrity.md docs/architecture/async-and-search.md
git commit -m "docs(product): consolidate product and semantic authority"

git add -- docs/architecture/backend.md docs/architecture/interfaces.md docs/architecture/persistence.md docs/architecture/transition.md docs/decisions/index.md
git commit -m "docs(architecture): consolidate technical realization authority"
```

---

### Task 6: Preserve Only Current Development, Operations, and Reference Consumers

**Files:**
- Create: `docs/development/setup.md`
- Create: `docs/development/testing.md`
- Create: `docs/development/verification.md`
- Create: `docs/operations/runbooks/restore.md`
- Create: `docs/operations/runbooks/release-backfill.md`
- Create: `docs/reference/repository-map.md`
- Create: `docs/reference/configuration.md`
- Create: `docs/reference/current-runtime-requirements.md`
- Move/rewrite: `wiki/architecture/req-trace-map.yaml` → `docs/reference/requirement-trace-map.yaml`
- Generate: `docs/reference/requirement-traceability.md`
- Modify: `scripts/req-trace/main.go`
- Modify: `scripts/req-trace/extract.go`
- Modify: `scripts/req-trace/extract_test.go`
- Modify: `scripts/req-trace/gate_test.go`
- Modify: `scripts/req-trace/report_test.go`
- Modify: `scripts/check-governance.ps1`

**Interfaces:**
- Consumes: current runtime/tool consumers on `main`.
- Produces: minimum operational/reference material needed until T10, without retaining legacy module/architecture narratives.

- [ ] **Step 1: Identify every hard-coded documentation consumer**

```bash
git grep -nE '(wiki/|docs/superpowers/|docs/operator/|docs/runbooks/)' -- \
  ':!docs/work/current/**' \
  ':!vendor/**' \
  ':!third_party/**' \
  > .tmp/doc-path-consumers.txt
cat .tmp/doc-path-consumers.txt
```

Every result must be repaired in Tasks 6–9 or explicitly removed with its obsolete consumer. Do not add compatibility stubs by default.

- [ ] **Step 2: Consolidate current setup/testing/verification guidance**

Sources with named current consumers include:

```text
wiki/references/local-dev-startup.md
wiki/references/environment-setup.md
wiki/references/how-to-run-tests.md
AGENTS.md verification section
tools/verify registry/profile behavior
```

Targets:

```text
docs/development/setup.md
docs/development/testing.md
docs/development/verification.md
```

Each command is verified against current scripts/config before publication. Remove old roadmap/status prose.

- [ ] **Step 3: Retain only current runbooks**

`docs/operations/runbooks/release-backfill.md` is derived from the existing release-backfill runbook and current source under `scripts/release-backfill/`.

`docs/operations/runbooks/restore.md` consolidates only current backup/restore/startup-readiness behavior still required by the running system and T4/T8-D. Do not copy historical migration plans.

- [ ] **Step 4: Replace the monolithic legacy requirement source with a compact current-runtime reference**

Create `docs/reference/current-runtime-requirements.md` by copying every existing normative `REQ-*` bullet from `wiki/architecture/backend-target-architecture.md` exactly, preserving class (`MUST|SHOULD|MAY`) and ID.

Frontmatter:

```yaml
---
id: reference-current-runtime-requirements
kind: reference
status: current
owner: verification
summary: Lists the current-runtime requirement identifiers still cited by tests until T10 replaces their implementation baseline.
---
```

Body note:

```text
This page is current verification input, not target architecture authority.
Deletion trigger: T10/T11 replace every remaining current-runtime requirement citation with the accepted target proof program.
```

Expected extraction count remains the current tested count of 69 unique REQ IDs.

- [ ] **Step 5: Retarget `scripts/req-trace`**

Change defaults in `scripts/req-trace/main.go` to:

```go
doc = filepath.Join(root, "docs", "reference", "current-runtime-requirements.md")
ReportPath: filepath.Join(root, "docs", "reference", "requirement-traceability.md")
MapPath: filepath.Join(root, "docs", "reference", "requirement-trace-map.yaml")
```

Update user-facing stale-report text to name the new report path.

Change `extract.go` comments to say “governing current-runtime requirement reference” rather than “backend target architecture”.

- [ ] **Step 6: Prove requirement-trace parity**

```bash
go test ./scripts/req-trace
go run ./scripts/req-trace -write
go run ./scripts/req-trace
```

Expected:

```text
69 REQ IDs
0 MUST uncovered
stale=false
```

If the count differs, stop; do not weaken the test or silently drop IDs.

- [ ] **Step 7: Retarget the runbook governance rule**

In `scripts/check-governance.ps1`, replace:

```powershell
'(?m)^docs/runbooks/'
```

with:

```powershell
'(?m)^docs/operations/runbooks/'
```

Update the corresponding guard fixture expected tree/path.

- [ ] **Step 8: Commit current-consumer documentation and path repairs**

```bash
git add -- docs/development docs/operations docs/reference scripts/req-trace scripts/check-governance.ps1 scripts/testdata/guard-fixtures/governance-diff-rules
git commit -m "docs(reference): preserve current operational consumers"
```

---

### Task 7: Replace Agent Bootstrap and Repository Entry Points

**Files:**
- Modify: `AGENTS.md`
- Keep minimal: `CLAUDE.md`
- Modify: `README.md`
- Modify: `.github/PULL_REQUEST_TEMPLATE.md`
- Modify: `.claude/settings.json`
- Delete: `.claude/skills/adversarial-review/`
- Delete: `.claude/skills/developing-new-work/`
- Delete after consumer census: `.claude/agents/frontend-screen-reviewer.md`
- Delete after consumer census: `.claude/agents/milestone-validator.md`
- Delete after consumer census: `.claude/agents/mission-validator.md`

**Interfaces:**
- Consumes: new semantic docs navigation and canonical external Method/Fable workflow.
- Produces: deterministic low-token agent orientation with no duplicated architecture/method/process authority.

- [ ] **Step 1: Replace `AGENTS.md` with the bounded bootstrap**

```markdown
# MetalDocs agent bootstrap

## Start

1. Read `docs/index.md`.
2. Read `docs/status.md` only when stage or implementation authority matters.
3. Read `docs/work/current/index.md` only when it exists on the active branch.
4. Use the intent table to open only the authority pages needed for the task.
5. Inspect code/schema/OpenAPI/frontend/runtime evidence only for concrete current-state claims.

## Material decisions

Apply `developmentconexus-ops/conexus-methodology/METHOD.md` v1.0.0.
Repository/product authority remains local. External practice is evidence, never product authority.

## Independent review

For material Fable review, follow the canonical workflow in `developmentconexus-ops/conexus-methodology/README.md` and use only `docs/work/current/ai-dialog.md`.
Reviewer output is evidence until Lead adjudication and operator ratification.

## Stable rules

- Never expose secrets, credentials, tokens, PII, or private keys.
- Treat generated files as generated.
- Do not weaken tests, verifiers, or safety controls to obtain green.
- Never push directly to `main`, force-push shared history, or stage unrelated files.
- Git history and closed PRs are the archive; do not create live archive/tombstone documentation.

## Verification

- Targeted iteration: run the owning package/check.
- PR gate: `go run ./tools/verify --profile=pr`.
- Report infrastructure limits; never bypass them silently.
```

No stage summary, authority chain, or architecture prose is copied into this file.

- [ ] **Step 2: Keep `CLAUDE.md` as a pointer only**

```markdown
# MetalDocs Claude bootstrap

Read and follow [`AGENTS.md`](AGENTS.md) first.

This file has no independent methodology, architecture, status, workflow, or next-step authority.
```

- [ ] **Step 3: Rewrite `README.md` for humans**

Keep only:

```text
what MetalDocs is
current implementation-blocked notice linking docs/status.md
read-first links to AGENTS.md and docs/index.md
local setup link
verification link
```

Delete cockpit, R10 stage narrative, and duplicated authority summaries.

- [ ] **Step 4: Replace the PR template**

Use this structure:

```markdown
## Gate / purpose

## Materiality and authority
- [ ] Trivial/mechanical
- [ ] Material — authority/invariant: ...

## Scope

## Proof
- [ ] Targeted checks
- [ ] `go run ./tools/verify --profile=pr`
- [ ] Documentation navigation/links when docs changed

## Independent review
- [ ] Not required
- [ ] Required and completed in `docs/work/current/ai-dialog.md`

## Merge hygiene
- [ ] Operator ratification recorded when required
- [ ] Temporary proposal/plan/AI dialogue deleted
- [ ] No unrelated files

## Risk / rollback / reopen trigger
```

- [ ] **Step 5: Remove obsolete local review/process skills**

Delete the local adversarial-review and developing-new-work skills because the canonical Method/Fable workflow now owns that process and the old skills create permanent staging files.

For each `.claude/agents/*.md` file, run:

```bash
git grep -n "$(basename <file> .md)" -- ':!<file>'
```

Delete when no current tool/config consumer exists. If a real consumer exists, move the bounded instruction into that consumer's configuration; do not retain a general historical agent role.

- [ ] **Step 6: Remove CCD session-message permissions from `.claude/settings.json`**

The final file retains only worktree settings that are currently used:

```json
{
  "worktree": {
    "baseRef": "head",
    "symlinkDirectories": [
      "node_modules",
      "frontend/apps/web/node_modules"
    ]
  }
}
```

Lead↔Fable communication now uses `ai-dialog.md`, not session-message MCP permissions.

- [ ] **Step 7: Commit bootstrap changes**

```bash
git add -- AGENTS.md CLAUDE.md README.md .github/PULL_REQUEST_TEMPLATE.md .claude
git commit -m "docs(governance): simplify repository and agent entry points"
```

---

### Task 8: Implement the Docs Hygiene Verifier with Negative Proof

**Files:**
- Create: `scripts/docs-hygiene/main.go`
- Create: `scripts/docs-hygiene/check.go`
- Create: `scripts/docs-hygiene/frontmatter.go`
- Create: `scripts/docs-hygiene/links.go`
- Create: `scripts/docs-hygiene/mkdocs.go`
- Create: `scripts/docs-hygiene/names.go`
- Create: `scripts/docs-hygiene/check_test.go`
- Create: `scripts/docs-hygiene/testdata/valid/`
- Create: `scripts/docs-hygiene/testdata/invalid-*/`
- Create: `scripts/testdata/guard-fixtures/docs-hygiene/wiki/bad.md.txt`
- Modify: `go.mod`
- Modify: `go.sum`
- Modify: `tools/verify/registry.go`
- Modify: `.github/workflows/ci.yml`

**Interfaces:**
- Consumes: tracked repository tree, `mkdocs.yml`, Markdown frontmatter, PR draft state from `METALDOCS_PR_DRAFT`.
- Produces: deterministic sorted findings; exit zero only when the live documentation tree satisfies the accepted profile.

- [ ] **Step 1: Add the Markdown parser dependency**

```bash
go get github.com/yuin/goldmark@v1.8.5
go mod tidy
```

`gopkg.in/yaml.v3` is already present and is used for frontmatter and `mkdocs.yml`.

- [ ] **Step 2: Define the public checker interface**

```go
package main

type Config struct {
    Root       string
    MergeReady bool
}

type Finding struct {
    Path    string
    Rule    string
    Message string
}

func Run(cfg Config) ([]Finding, error)
```

`main.go`:

```go
func main() {
    root := flag.String("root", ".", "repository root")
    flag.Parse()

    mergeReady := strings.EqualFold(os.Getenv("METALDOCS_PR_DRAFT"), "false")
    findings, err := Run(Config{Root: *root, MergeReady: mergeReady})
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    for _, f := range findings {
        fmt.Printf("%s: %s: %s\n", f.Path, f.Rule, f.Message)
    }
    if len(findings) > 0 {
        os.Exit(1)
    }
}
```

- [ ] **Step 3: Implement tracked-file discovery and root rules**

Use:

```bash
git -C <root> ls-files -z
```

Rules:

```text
first-party Markdown lives under docs/
root exceptions: README.md, AGENTS.md, CLAUDE.md, SECURITY.md, CONTRIBUTING.md, .github/*.md
third_party/** and vendor/** are third-party-managed exclusions
wiki/**, docs/superpowers/**, docs/operator/** are always forbidden
archive, legacy, historical, tombstone directory segments are forbidden under docs/
```

- [ ] **Step 4: Implement frontmatter and naming rules**

Frontmatter model:

```go
type Frontmatter struct {
    ID      string `yaml:"id"`
    Kind    string `yaml:"kind"`
    Status  string `yaml:"status"`
    Owner   string `yaml:"owner"`
    Summary string `yaml:"summary"`
}
```

Validate:

```text
all five values nonblank
kind in authority|guide|reference|work
status in current|active
id globally unique
summary one physical line
```

Durable filename regex:

```text
^[a-z0-9]+(?:-[a-z0-9]+)*\.md$
```

Allowed exceptions:

```text
docs/decisions/adr-[0-9]{4}-<slug>.md
docs/operations/incidents/YYYY-MM-DD-<slug>.md
```

Rejected durable filename tokens:

```text
candidate corrected review adjudication amendment tombstone legacy historical final new old
```

Reject stage/date/version patterns outside the explicit exceptions.

- [ ] **Step 5: Implement MkDocs navigation parity**

Parse `mkdocs.yml` with:

```go
type MkDocs struct {
    Nav []any `yaml:"nav"`
}
```

Recursively collect every string leaf under `nav`.

Rules:

```text
every durable docs/**/*.md page appears exactly once in nav
no docs/work/** page appears in nav
all nav targets exist
no duplicate nav target
```

- [ ] **Step 6: Implement internal-link validation**

Use Goldmark AST to collect `*ast.Link` and `*ast.Image` destinations.

Ignore:

```text
http://
https://
mailto:
urn:
fragment-only links
```

Resolve relative file targets against the source page's directory. Strip query/fragment before file existence checks. Directory links resolve to `index.md`.

Fail on missing file targets and on links that escape the repository root.

- [ ] **Step 7: Implement active-work rules**

Always enforce:

```text
only docs/work/current/ may contain work pages
at most one proposal.md
at most one plan.md
at most one ai-dialog.md
all work pages use kind=work,status=active
```

When `MergeReady` is true, also enforce:

```text
docs/work/current/proposal.md absent
docs/work/current/plan.md absent
docs/work/current/ai-dialog.md absent
```

- [ ] **Step 8: Write focused unit tests first**

Required tests:

```go
func TestRunAcceptsValidTree(t *testing.T)
func TestRunRejectsWikiRoot(t *testing.T)
func TestRunRejectsMissingFrontmatter(t *testing.T)
func TestRunRejectsDuplicateID(t *testing.T)
func TestRunRejectsProhibitedDurableName(t *testing.T)
func TestRunRejectsOrphanDurablePage(t *testing.T)
func TestRunRejectsWorkPageInNavigation(t *testing.T)
func TestRunRejectsBrokenRelativeLink(t *testing.T)
func TestRunRejectsTemporaryFilesWhenMergeReady(t *testing.T)
```

Each test creates an isolated git repository fixture, commits its files, calls `Run`, and asserts the exact rule ID.

Rule IDs:

```text
docs.root
docs.frontmatter
docs.id
docs.name
docs.nav
docs.link
docs.work
```

- [ ] **Step 9: Run tests RED, then implement GREEN**

```bash
go test ./scripts/docs-hygiene -run TestRun -count=1
```

Expected before implementation: compile/test failure.

After implementation:

```bash
go test ./scripts/docs-hygiene -count=1
```

Expected: PASS.

- [ ] **Step 10: Register the blocking check and a command-level negative fixture**

Add to `tools/verify/registry.go`:

```go
{
    ID:       "docs-hygiene",
    Desc:     "documentation has one governed docs root, valid metadata, complete navigation, resolvable internal links, and no merge-ready temporary work",
    Profiles: []string{ProfileFast, ProfilePR, ProfileFull},
    Argv:     []string{"go", "run", "./scripts/docs-hygiene"},
    Paths: []string{
        "docs/**",
        "wiki/**",
        "AGENTS.md",
        "CLAUDE.md",
        "README.md",
        "mkdocs.yml",
        ".github/**",
        "scripts/docs-hygiene/**",
    },
    CIJob: "ci.yml:verify",
    Fixture: &Fixture{
        Dir:          "docs-hygiene",
        ArgvOverride: []string{"go", "run", "./scripts/docs-hygiene", "--root", "{{fixture}}"},
        Want:         []string{"wiki/bad.md", "docs.root"},
    },
},
```

The fixture tree contains a minimal git repository with `wiki/bad.md.txt`; the harness strips `.txt` and proves the actual command exits non-zero for the expected rule.

- [ ] **Step 11: Pass PR draft state to the verifier**

In `.github/workflows/ci.yml`, set this environment variable on the verify invocation/job:

```yaml
env:
  METALDOCS_PR_DRAFT: ${{ github.event.pull_request.draft }}
```

On non-PR events the value is empty, so the check performs structural validation without PR-finalization semantics.

- [ ] **Step 12: Prove the verifier and fixture**

```bash
go test ./scripts/docs-hygiene ./tools/verify
go run ./tools/verify --guard-fixtures --only=docs-hygiene
go run ./scripts/docs-hygiene
```

Expected: all pass while G1 remains Draft. The merge-ready mode is intentionally tested in Task 10 after temporary files are removed.

- [ ] **Step 13: Commit the verifier**

```bash
git add -- go.mod go.sum scripts/docs-hygiene scripts/testdata/guard-fixtures/docs-hygiene tools/verify/registry.go .github/workflows/ci.yml
git commit -m "feat(verify): enforce documentation hygiene"
```

---

### Task 9: Delete the Legacy Documentation Estate and Repair Every Live Consumer

**Files:**
- Delete: `wiki/`
- Delete: `docs/superpowers/`
- Delete: `docs/operator/`
- Delete: `docs/HARNESS-PROFILE.md`
- Delete: local Method mirror superseded by canonical `conexus-methodology/METHOD.md`
- Delete: any first-party Markdown outside approved roots after consumer repair
- Modify: every tracked file returned by the path-consumer census

**Interfaces:**
- Consumes: complete replacement docs, repaired scripts, active docs-hygiene verifier.
- Produces: one live documentation tree with zero compatibility stubs unless a real external consumer was proven.

- [ ] **Step 1: Record the pre-deletion tracked Markdown census**

```bash
git ls-files '*.md' '*.yaml' '*.yml' > .tmp/pre-deletion-doc-files.txt
```

- [ ] **Step 2: Delete the known legacy roots atomically**

```bash
git rm -r -- wiki docs/superpowers docs/operator
git rm -- docs/HARNESS-PROFILE.md
git rm -- docs/engineering/standards/root-cause-global-maximum-method.md
```

If a path is absent on the G1 base, record that fact and continue; do not recreate it.

- [ ] **Step 3: Delete remaining obsolete first-party Markdown**

```bash
git ls-files '*.md' | sort
```

For every first-party Markdown file outside:

```text
docs/**
README.md
AGENTS.md
CLAUDE.md
SECURITY.md
CONTRIBUTING.md
.github/*.md
```

move unique current meaning into the owning target page or delete the file. Exclude only `vendor/**` and `third_party/**` as third-party-managed trees.

- [ ] **Step 4: Repair all path consumers**

```bash
git grep -nE '(wiki/|docs/superpowers/|docs/operator/|docs/HARNESS-PROFILE\.md|docs/engineering/standards/root-cause-global-maximum-method\.md)' -- \
  ':!vendor/**' ':!third_party/**'
```

Expected: zero results.

Do not satisfy this with redirect stubs. Update the consuming code/config/doc to the semantic target or delete the obsolete consumer.

- [ ] **Step 5: Prove navigation and requirement trace after deletion**

```bash
go run ./scripts/docs-hygiene
go run ./scripts/req-trace
go test ./scripts/req-trace ./scripts/docs-hygiene
```

Expected: all pass.

- [ ] **Step 6: Commit the atomic deletion**

```bash
git add --all -- docs wiki AGENTS.md CLAUDE.md README.md .github .claude scripts tools go.mod go.sum mkdocs.yml
git commit -m "docs: remove legacy documentation estate"
```

Before committing, inspect `git diff --cached --name-status`; unstage any unrelated product file. Do not use `git add -A` outside the exact path list above.

---

### Task 10: Prove Authority Parity and Run the Final G1 Fable Review

**Files:**
- Create temporarily: `docs/work/current/ai-dialog.md`
- Modify only if findings require correction: durable `docs/**`, scripts, verifier
- Delete before merge: `docs/work/current/index.md`
- Delete before merge: `docs/work/current/proposal.md`
- Delete before merge: `docs/work/current/ai-dialog.md`

**Interfaces:**
- Consumes: PR #131 immutable source authority and G1 target tree.
- Produces: independently challenged proof that deletion lost no accepted decision or current operational safety.

- [ ] **Step 1: Run the complete local proof**

```bash
go test ./scripts/docs-hygiene ./scripts/req-trace ./tools/verify
go run ./scripts/docs-hygiene
go run ./scripts/req-trace
go run ./tools/verify --guard-fixtures --only=docs-hygiene
go run ./tools/verify --profile=pr
```

- [ ] **Step 2: Run the old-path and unindexed-page proof**

```bash
git grep -nE '(wiki/|docs/superpowers/|docs/operator/)' -- ':!vendor/**' ':!third_party/**'
```

Expected: no output.

```bash
git ls-files 'docs/**/*.md' | sort > .tmp/tracked-docs.txt
```

`go run ./scripts/docs-hygiene` proves every durable file in that list is navigated and linked correctly.

- [ ] **Step 3: Create the final AI dialogue**

The review request must require Fable to compare:

```text
source: PR #131 HEAD d8b1c6d31e704e9552a14faa7764c634a29b081d
against:
Product contract
ownership
domain model
lifecycle
authorization
audit
content integrity
async/search
backend
interfaces
persistence
transition
decision index
T8-E checkpoint in the active proposal
current runtime requirements/runbooks
```

Required verdict:

```text
APPROVE DOCS-ONLY CONSOLIDATION
APPROVE DOCS-ONLY CONSOLIDATION WITH MATERIAL FIXES
DO NOT APPROVE DOCS-ONLY CONSOLIDATION
```

- [ ] **Step 4: Adjudicate findings and obtain operator ratification**

A finding may reopen only the implicated consolidation decision. Product/R10 semantics are not changed under the label of documentation correction.

- [ ] **Step 5: Remove all temporary work files**

```bash
git rm -- docs/work/current/index.md docs/work/current/proposal.md docs/work/current/ai-dialog.md
```

Remove the now-empty `docs/work/current/` directory from the tracked tree.

- [ ] **Step 6: Prove merge-ready mode**

```bash
METALDOCS_PR_DRAFT=false go run ./scripts/docs-hygiene
```

Expected: PASS, proving no proposal, plan, or AI dialogue survives.

- [ ] **Step 7: Run the final required gate and commit**

```bash
go run ./tools/verify --profile=pr
git add -- docs mkdocs.yml AGENTS.md CLAUDE.md README.md .github .claude scripts tools go.mod go.sum
git commit -m "docs: finalize documentation rebaseline"
```

Set the G1 PR ready only after the remote required checks are green.

---

### Task 11: Squash-Merge G1 and Close PR #131 as Superseded

**Files:**
- Repository files: none after G1 finalization.
- GitHub metadata: G1 PR, PR #131.

**Interfaces:**
- Consumes: operator-ratified, green G1 PR.
- Produces: clean `main`; PR #131 preserved as closed provenance, never merged.

- [ ] **Step 1: Verify the G1 remote head and changed-file scope**

```text
all required checks green
no temporary work files
no product code/schema/OpenAPI/frontend/runtime/deploy changes
```

- [ ] **Step 2: Squash-merge G1**

Squash title:

```text
docs: rebaseline repository documentation and governance
```

- [ ] **Step 3: Verify the new `main`**

```bash
git fetch origin main
git switch main
git reset --hard origin/main
go run ./scripts/docs-hygiene
go run ./tools/verify --profile=pr
```

Do not run `reset --hard` in a worktree containing unrelated local work.

- [ ] **Step 4: Close PR #131 without merging**

Comment:

```text
Superseded by the docs-only authority consolidation merged from G1.
Accepted Product/R10 truth through T8-D and the T8-E checkpoint were preserved in the semantic docs structure. This PR remains provenance and was intentionally not merged.
```

Closing PR #131 requires explicit operator authorization at execution time.

---

### Task 12: Resume T8-E in a Fresh PR from Clean Main

**Files:**
- Modify: `docs/status.md`
- Create: `docs/work/current/index.md`
- Create: `docs/work/current/proposal.md`

**Interfaces:**
- Consumes: clean merged docs baseline and the accepted T8-E checkpoint preserved by G1 parity review.
- Produces: one small active T8-E proposal without reintroducing stage-named durable files or permanent review artifacts.

- [ ] **Step 1: Create the fresh branch**

```bash
git fetch origin main
git switch -c arch/executable-api-contract origin/main
```

- [ ] **Step 2: Create the active-work index**

```markdown
---
id: work-current
kind: work
status: active
owner: architecture
summary: Routes the active executable API contract proposal.
---

# Current work

| Field | Value |
|---|---|
| Topic | Executable API contract |
| Branch | `arch/executable-api-contract` |
| Pull request | Draft T8-E PR |
| Proposal | `proposal.md` |
| Current checkpoint | Layers 1–4 accepted; operation/schema ledger and size measurement remain |
| Next action | Complete the 78-operation contract ledger and measured upload limits |
```

- [ ] **Step 3: Materialize the accepted T8-E checkpoint in one proposal**

The proposal includes, without re-deriving:

```text
one OpenAPI /api/v1 SSOT
one generated Go wire boundary
one generated TypeScript boundary
purpose-built schemas and semantic operationIds
78-operation census after UserProfile/AreaLifecycle precision
strong opaque ETag and If-Match rules
UserProfile If-None-Match:* recreation
10-operation Idempotency-Key matrix
createSubmission Idempotency-Key + DRAFT If-Match
session-bound CSRF
stateless opaque pagination
RFC 9457 closed Problem catalog
https://errors.conexus.fun/{product}/{code}
create-only direct upload
server-authoritative completion
exact-byte response contract
Submission/Governance/Release/Rendition views
open measurement: raw/expanded DOCX/PDF limits
```

No historical T8-E review/candidate files are recreated.

- [ ] **Step 4: Update only the status pointer**

`docs/status.md` remains the sole status authority and points to `docs/work/current/index.md` for the active branch.

- [ ] **Step 5: Open the Draft T8-E PR and continue the architecture process**

```bash
git add -- docs/status.md docs/work/current/index.md docs/work/current/proposal.md
git commit -m "docs(api): resume executable contract design"
git push -u origin arch/executable-api-contract
```

Use `docs/work/current/ai-dialog.md` only after the coherent T8-E proposal is ready for its final independent challenge.

---

## Plan Self-Review Results

### Spec coverage

- One docs root: Tasks 4, 9.
- Semantic naming/frontmatter: Tasks 4, 8.
- Task-oriented index and MkDocs navigation: Task 4.
- Short AGENTS/CLAUDE entrypoints: Task 7.
- One proposal and one AI dialogue: Tasks 2, 3, 10, 12.
- One PR per coherent gate: execution graph and Tasks 0–12.
- Git as archive and PR #131 supersession: Tasks 9–11.
- Allowlist deletion with current consumer preservation: Tasks 6, 9.
- Mechanical proof with negative fixtures: Task 8.
- R10 parity and T8-E checkpoint transfer: Tasks 5, 10, 12.

### Placeholder scan

The plan contains no `TBD`, `TODO`, unspecified implementation step, or Writer-owned material decision. Unknown repository consumers are handled by a fail-closed grep census: any unresolved consumer stops deletion rather than being guessed.

### Interface consistency

- Temporary work paths are consistently `docs/work/current/{index,proposal,plan,ai-dialog}.md`.
- Durable documentation authority is consistently `docs/development/documentation.md`.
- Documentation guard command is consistently `go run ./scripts/docs-hygiene` with `--root` for fixtures and `METALDOCS_PR_DRAFT=false` for merge-ready proof.
- Requirement trace paths are consistently under `docs/reference/`.
