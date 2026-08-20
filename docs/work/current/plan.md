# Repository documentation rebaseline plan

> Temporary execution plan. This file is deleted before PR #132 becomes merge-ready.

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

## S0 — trustworthy verification baseline

Create a separate prerequisite PR from current `main` and repair the concrete security/toolchain failures already proven by CI. Do not weaken thresholds or waive the failing checks.

Required proof mirrors CI rather than using a weaker local shorthand:

```text
go test ./...
go run ./tools/verify --require-infra --profile=pr
go run ./tools/verify --require-infra --only=golangci-lint
```

Run the exact security scans required by the workflow. S0 merges before G0 is made merge-ready.

## G0 — PR #132

PR #132 decides only the repository documentation profile.

### Durable result

`docs/development/documentation.md` is the sole durable candidate in this gate.

### Temporary evidence

```text
docs/work/current/proposal.md
docs/work/current/plan.md
docs/work/current/ai-dialog.md
```

### Review/adjudication flow

1. Fable review is already preserved in Git at PR #132 HEAD `3b8a25488e1aed5edc6c2b83d64e802b8d66c1c0`.
2. Lead adjudication is recorded in `ai-dialog.md`.
3. No second Fable round is required unless adjudication reopens the docs root, naming model, or retention predicate.
4. Operator ratifies or reopens the corrected profile.
5. After S0 is merged, merge updated `main` into this branch if necessary.
6. Record durable provenance.
7. Delete all `docs/work/current/*` files.
8. Ensure CI includes `pull_request` activity `ready_for_review` before relying on PR-state-sensitive hygiene checks.
9. Run required checks green.
10. Mark ready and squash merge.

G0 must not delete `wiki/` or migrate Product/R10 authorities.

## G1 — complete documentation rebaseline

Branch from updated `main` after G0.

### Step 1 — complete tracked-document disposition census

Enumerate the entire tracked documentation estate rather than a hand-written list. At minimum start from:

```text
git ls-files '*.md'
```

Include other documentation-like machine subjects discovered from verification/generator configuration.

Every tracked page receives exactly one explicit disposition:

```text
KEEP → target path
MERGE → target authority
GENERATED → generator + target path
DELETE → reason + proof no current consumer remains
```

An undispositioned path blocks the PR.

The census must cover existing `docs/runbooks/**`, `docs/engineering/**`, all `wiki/**` document classes, and any other live Markdown already under `docs/`.

### Step 2 — verification-subject census

Inspect verifier registry, workflow configuration, generators, scripts, and analyzers for documentation paths.

At minimum explicitly disposition the current subjects identified by Fable:

```text
problem-codes-drift → problem-code reference
a req-trace gate → governing requirements + generated trace report
adr-status → ADR collection
db-docs-coverage → database table dictionary/ownership pages
wiki-debt-tally → module-debt subject or gate retirement
```

For every gate:

```text
subject moves → repoint gate in same PR
subject dies  → retire gate with named rationale in same PR
```

Re-run the gate's relevant negative fixture/proof. Include affected non-PR/nightly governance checks, not only the PR profile.

### Step 3 — classify old-path consumers

Do not require zero textual occurrences of old paths.

Classify each discovered occurrence as one of:

```text
EXECUTABLE CONSUMER
  script/config/workflow/generator/verifier/navigation/maintained link
  → repair or retire

PROVENANCE/HISTORY CITATION
  source comment, applied migration, generated historical artifact,
  commit-pinned secret-scan fingerprint, history-only record
  → leave unless it is also executable
```

History-pinned `.gitleaksignore` entries remain when the full-history secret scan still consumes them.

### Step 4 — preserve repository engineering/current-runtime law

Before shrinking `AGENTS.md` and `CLAUDE.md`, move surviving load-bearing content to:

```text
docs/development/engineering-rules.md
docs/reference/current-system.md
```

Then reduce root agent files to pointers/bootstrap only.

### Step 5 — Product/R10 authority consolidation

Read accepted durable source authority from immutable PR #131 HEAD `d8b1c6d31e704e9552a14faa7764c634a29b081d`.

Maintain a closed source-to-target map covering Product Contract, ownership, T1→T8-D, Decision Registry amendments with surviving meaning, and accepted T8-E checkpoint work.

Never silently reconcile contradictory source statements. Stop and surface the exact changed decision.

Support the semantic parity review with a normative census. If sources are extracted under ignored `.tmp/`, use a filesystem search such as `grep -rnE`/`rg`, not `git grep`.

The source census must:

- assert non-empty output;
- cover actual normative vocabulary, including `MUST`, `SHOULD`, `MAY`, `SHALL`, `REQUIRE`, `FORBID`, `PROHIBIT`, `NEVER`, `ALWAYS`, `ONLY`, `SELECT`, `REJECT`, `BLOCKED`, `CLOSED` where normative;
- compare against the new authorities only as supporting evidence;
- never substitute for semantic source-to-target review.

### Step 6 — machine-consumed target homes

Target homes include, as needed by current consumers:

```text
docs/decisions/adr-NNNN-<slug>.md
docs/reference/database/<subject>.md
docs/reference/problem-codes.md
docs/reference/requirement-traceability.md
```

Generated requirement/reference pages receive frontmatter from the generator itself. Do not hand-edit generated output.

Large collections are linked through governed collection indexes and need not each appear as a top-level `mkdocs.yml` nav entry.

### Step 7 — navigation and status

Create:

```text
docs/index.md
docs/status.md
docs/decisions/index.md
mkdocs.yml
```

`mkdocs.yml` is the explicit navigation manifest; static publishing is not part of G1.

`docs/index.md` routes by reader intent. `docs/status.md` owns current stage. Indexes do not duplicate authority prose.

### Step 8 — docs hygiene guard

Implement a repository-native verifier on the existing Go verification spine.

It must prove at least:

```text
single docs root in final tree
valid durable filenames
unique id/kind/owner/summary frontmatter
kind ∈ authority|work
generated-page metadata generated by owner tool
broken maintained links fail
orphan durable pages/collections fail
work pages excluded from durable navigation
docs/work/** fails when PR is not Draft
```

The fixture includes a minimal valid navigation/document tree so a negative case fails for the intended rule, not missing setup.

Add `ready_for_review` to the GitHub Actions `pull_request` activity types before treating the work-file guard as merge protection.

### Step 9 — delete legacy only after proofs

Only after Steps 1–8 are green:

- remove `wiki/`, `docs/superpowers/`, `docs/operator/`, completed review/candidate/tombstone trees, and other paths marked DELETE;
- use deletion commands generated from the disposition census rather than a brittle hard-coded `git rm` list;
- do not remove history-pinned security entries or provenance-only source text to make a grep return zero.

### Step 10 — G1 proof ladder

Run the workflow-equivalent ladder:

```text
go test ./...
go run ./tools/verify --require-infra --profile=pr
go run ./tools/verify --require-infra --only=golangci-lint
```

Also execute every affected non-PR/nightly documentation-governance check explicitly.

Prove:

- Product/R10 authority parity;
- complete path disposition;
- no undispositioned documentation;
- all executable consumers repaired/retired;
- all affected gate subjects repaired/retired;
- negative fixtures fire;
- temporary G1 work removed before ready-for-review;
- required CI green.

Then final Fable review through one temporary `ai-dialog.md`, Lead adjudication, operator ratification, cleanup, and squash merge.

## After G1

Close PR #131 as superseded provenance only after the new tree proves accepted authority parity.

Create a fresh T8-E PR from updated `main`. Its proposal resumes the already accepted checkpoint including:

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

The measured document/upload-size limits remain open and must not be guessed during repository cleanup.

## Stop conditions

Stop instead of guessing if:

- a current verification subject has no safe target;
- a current-runtime safety document has no named successor;
- Product/R10 sources conflict semantically;
- a proposed path repair requires product/OpenAPI/runtime behavior changes;
- the parity census is empty or incomplete;
- required CI/negative proof cannot be made to fire.
