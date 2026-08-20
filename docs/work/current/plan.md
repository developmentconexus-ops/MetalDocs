# Repository documentation rebaseline plan

> Temporary execution plan. Delete this file before PR #132 becomes merge-ready.

## Goal

Move MetalDocs to one governed `docs/` root without losing accepted Product/R10 truth, current-runtime safety material, or verification subjects, then resume T8-E in a fresh small PR.

## Execution graph

```text
S0 verification baseline
→ G0 repository documentation profile (PR #132)
→ G1 complete census + consolidation + consumer repair + legacy deletion + verifier
→ close PR #131 as superseded provenance
→ fresh T8-E PR
```

Do not stack G1 on PR #132.

## Global constraints

- G0/G1 authorize documentation/governance only.
- No product code, schema, OpenAPI, frontend, runtime, or deployment behavior change is authorized merely to make documentation paths disappear.
- Existing verification gates remain in force until their subject is repointed or the gate is explicitly retired with equal-or-stronger proof.
- PR #131 is immutable provenance input and is never merged into G0/G1.
- Temporary work exists only while the PR is Draft.
- No force-push/rebase of shared pushed history merely to refresh the base; merge updated `main` or create the next clean branch.

## Current gate

The independent Fable review and Lead adjudication are complete. Explicit operator ratification of the corrected G0 profile is the only architecture/governance decision next.

Before that ratification:

```text
DO NOT start S0 implementation from this branch
DO NOT start G1
DO NOT delete wiki/ or legacy docs
DO NOT migrate Product/R10 authorities
DO NOT resume T8-E
DO NOT mark PR #132 ready
DO NOT merge PR #132
```

## S0 — trustworthy verification baseline

After operator ratification, create a separate prerequisite PR from current `main` and repair only the concrete security/toolchain failures proven by CI. Do not weaken thresholds or waive failing checks.

Required local proof mirrors CI:

```text
go test ./...
go run ./tools/verify --require-infra --profile=pr
go run ./tools/verify --require-infra --only=golangci-lint
```

Run the exact security scans required by the workflow. S0 merges before G0 is made merge-ready.

## G0 — finish PR #132 after S0

1. Merge updated `main` into the branch if necessary; do not rewrite shared history merely to refresh base.
2. Revalidate `docs/development/documentation.md` against the operator-ratified correction set.
3. Ensure the PR records its durable provenance.
4. Ensure the CI `pull_request` trigger includes `ready_for_review` before relying on a PR-state-aware `docs/work/**` guard. If that CI correction is not yet safely present, keep PR #132 Draft and perform it in the smallest coherent prerequisite/governance change rather than pretending the guard exists.
5. Delete all temporary `docs/work/current/*` files.
6. Run required checks green.
7. Mark Ready only after the work tree is clean and protections can re-evaluate.
8. Squash merge.

G0 does not delete `wiki/` or migrate Product/R10 authorities.

## G1 — complete documentation rebaseline

Branch from updated `main` after G0.

### 1. Complete tracked-document disposition census

Start from the full tracked Markdown estate, not a hand-written list:

```text
git ls-files '*.md'
```

Add documentation-like machine subjects discovered from verification/generator configuration.

Every page receives exactly one disposition:

```text
KEEP → target path
MERGE → target authority
GENERATED → generator + target path
DELETE → reason + proof no current consumer remains
```

An undispositioned path blocks G1. The census explicitly includes current `docs/runbooks/**`, `docs/engineering/**`, all `wiki/**` classes, and any other live Markdown already under `docs/`.

### 2. Verification-subject census

Inspect verifier registry, workflows, generators, scripts, and analyzers for documentation paths.

At minimum disposition current subjects for:

```text
problem-codes-drift
req-trace
adr-status
db-docs-coverage
wiki-debt-tally
```

For each gate:

```text
subject moves → repoint gate in same PR
subject dies  → retire gate in same PR with named rationale
```

Re-run the relevant negative proof. Execute affected non-PR/nightly governance checks explicitly, not only the PR profile.

### 3. Classify old-path consumers

Do not require zero textual occurrences of old paths.

```text
EXECUTABLE CONSUMER
  script/config/workflow/generator/verifier/navigation/maintained live link
  → repair or retire

PROVENANCE/HISTORY CITATION
  source comment, applied migration, generated historical artifact,
  commit-pinned secret-scan fingerprint, history-only record
  → leave unless also executable
```

History-pinned `.gitleaksignore` entries remain when the full-history scanner still consumes them.

### 4. Preserve repository engineering/current-runtime law

Before shrinking `AGENTS.md` and `CLAUDE.md`, move surviving load-bearing content to:

```text
docs/development/engineering-rules.md
docs/reference/current-system.md
```

Then reduce root agent files to routing/bootstrap only.

### 5. Product/R10 authority consolidation

Read accepted source authority from immutable PR #131 HEAD:

```text
d8b1c6d31e704e9552a14faa7764c634a29b081d
```

Maintain a closed source-to-target map covering Product Contract, ownership, T1→T8-D, surviving Decision Registry meaning, and the accepted T8-E checkpoint.

Never silently reconcile contradictory source statements. Stop and surface the exact changed decision.

Support semantic parity with a normative census. For ignored temporary source extraction use filesystem search (`grep -rnE`, `rg`, or equivalent), not `git grep`.

The source census must:

- be explicitly non-empty;
- cover actual normative vocabulary including `MUST`, `SHOULD`, `MAY`, `SHALL`, `REQUIRE`, `FORBID`, `PROHIBIT`, `NEVER`, `ALWAYS`, `ONLY`, `SELECT`, `REJECT`, `BLOCKED`, `CLOSED` where normative;
- serve as supporting evidence rather than substitute for semantic mapping review.

### 6. Machine-consumed target homes

As required by current consumers, target homes include:

```text
docs/decisions/adr-NNNN-<slug>.md
docs/reference/database/<subject>.md
docs/reference/problem-codes.md
docs/reference/requirement-traceability.md
```

Generated durable pages receive frontmatter from their generator. Large governed collections use collection indexes and need not each appear as top-level `mkdocs.yml` entries.

### 7. Navigation and status

Create the smallest set actually required by surviving content, including:

```text
docs/index.md
docs/status.md
docs/decisions/index.md
mkdocs.yml
```

`mkdocs.yml` is the navigation manifest; static publishing is not a G1 requirement.

### 8. Docs hygiene guard

Implement on the existing Go verification spine.

It proves at least:

```text
single docs root in final tree
valid durable filenames
unique id/kind/owner/summary frontmatter
kind ∈ authority|work
generated-page metadata emitted by generator
broken maintained links fail
orphan durable pages/collections fail
work pages excluded from durable navigation
docs/work/** fails when PR is not Draft
```

Fixtures contain a minimal valid navigation/repository context and assert the intended finding.

Before treating PR state as merge protection, GitHub Actions must trigger on `ready_for_review`.

### 9. Delete legacy only after proofs

Only after Steps 1–8 are satisfied:

- delete paths classified `DELETE`, including legacy roots no longer consumed;
- derive deletion commands from the disposition census;
- do not remove provenance-only source text or history-pinned security entries merely to make old-path grep output empty.

### 10. G1 proof ladder

Run:

```text
go test ./...
go run ./tools/verify --require-infra --profile=pr
go run ./tools/verify --require-infra --only=golangci-lint
```

Also execute every affected non-PR/nightly documentation-governance check.

Prove:

- Product/R10 authority parity;
- complete path disposition;
- zero undispositioned documentation;
- all executable consumers repaired/retired;
- all affected gate subjects repaired/retired;
- negative fixtures fire for the intended rule;
- temporary G1 work removed before Ready;
- required CI green.

Then run one final Fable review through a temporary `ai-dialog.md`, Lead adjudication, operator ratification, cleanup, and squash merge.

## After G1

Close PR #131 as superseded provenance only after the new tree proves accepted authority parity.

Create a fresh T8-E PR from updated `main`. Resume the accepted checkpoint:

```text
one OpenAPI application SSOT
one generated Go wire boundary
one generated TypeScript boundary
semantic operation IDs and purpose-built responses
strong ETag/If-Match rules
bounded T6 resource precision
Idempotency-Key matrix
session-bound CSRF
stateless cursor pagination
closed RFC 9457 Problem catalog
errors.conexus.fun/{product}/{code}
create-only direct upload
server-authoritative completion
exact-byte response contract
Submission/Governance/Release/Rendition wire shapes
```

Measured document/upload-size limits remain open and must not be guessed during cleanup.

## Stop conditions

Stop instead of guessing if:

- a current verification subject has no safe target;
- a current-runtime safety document has no named successor;
- Product/R10 sources conflict semantically;
- a path repair would require product/OpenAPI/runtime behavior changes;
- the parity census is empty or incomplete;
- required CI/negative proof cannot be demonstrated to fire.
