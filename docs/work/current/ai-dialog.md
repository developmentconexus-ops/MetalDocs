---
id: repository-reset-ai-dialog
kind: work
owner: architecture
summary: Temporary final Fable review and Lead/operator decision record for the clean-slate repository reset.
---

# AI dialogue — repository reset

> TEMPORARY / NON-AUTHORITATIVE / DELETE BEFORE MERGE

## Review context

```text
Repository: developmentconexus-ops/MetalDocs
Branch: reset/clean-slate-repository
PR: #134
Expected HEAD: REVALIDATE BEFORE REVIEW
Current gate: clean-slate source-tree reset
Implementation: BLOCKED
T8-E: paused at docs/reference/t8e-checkpoint.md
```

## Review request

Perform one independent adversarial review of the proposed reset.

Attack especially:

1. Did the new tree preserve every current Product/R10 authority needed to continue architecture work?
2. Is any deleted implementation actually an externally binding compatibility contract rather than DEV/test history?
3. Is advancing source-tree cleanup before future T10 runtime cutover coherent with the ratified no-historical-business-migration posture?
4. Does any survivor still encode superseded Approval/Taxonomy/tenancy/local-password/frontend/runtime architecture as current truth?
5. Is the minimal CI genuinely sufficient for an architecture-only repository while implementation is blocked?
6. Are `AGENTS.md`, `docs/index.md`, `docs/status.md`, and the work router deterministic and non-duplicative?
7. Is the paused T8-E checkpoint complete enough to resume without conversation archaeology?
8. Can more be deleted without losing a distinct current property?
9. Is there any accepted decision present only in PR #131 that the semantic authority set failed to carry?
10. Is there a materially better Global Maximum than this source-tree reset?

Classify each material finding:

```text
BLOCKER | MAJOR | LOW
claim
evidence
property at risk
smallest correction
upstream reopen required: yes/no
```

Primary verdict — exactly one:

```text
APPROVE CLEAN-SLATE REPOSITORY RESET
APPROVE CLEAN-SLATE REPOSITORY RESET WITH MATERIAL FIXES
DO NOT APPROVE CLEAN-SLATE REPOSITORY RESET
```

## Fable review

Independent adversarial review. Remote HEAD revalidated before review:
`origin/reset/clean-slate-repository` = `feb6c27231667eced554867a499b6cb578ae8fc7`,
PR #134 OPEN / Draft / MERGEABLE, +10,344 / −1,709,872 over 7,612 files, live tree = 32
files. Method applied: DevelopmentConexus Engineering Method v1.0.0.

### Structural verdict first

The reset structure is correct and is the Global Maximum for the stated problem.
Delete-in-place on a branch, squash-merged onto `main`, with `main`'s own history as
rollback, beats the credible alternatives: an overlay on the existing tree preserves
exactly the false constraints the decision names; a new repository loses rollback
continuity, ruleset identity and PR lineage; an orphan-branch rewrite is forbidden by
`docs/development/engineering-rules.md` §Git. Outcome class is **CURRENT STRUCTURE
CONFIRMED**, not RESTRUCTURE.

Every finding below is a bounded correction inside that structure. Three are BLOCKER
because merge, or the next action merge schedules, destroys the property irreversibly.

### BLOCKER-1 — the provenance premise is false for the corpus this reset depends on

**Claim.** `docs/decisions/repository-reset.md` states "No deleted file is destroyed
historically: Git and closed pull requests retain it." That is false for the R10 authority
corpus, which was never merged to `main`.

**Evidence.**

```text
git merge-base origin/main origin/docs/a8-authz-approval-redesign-ledger
  = 7f5b8928cc5a13feb8ee3fa7c8ceb1c7d3655a18
PR #131 state = CLOSED, never merged
wiki/architecture/*.md added by PR #131          = 29 files
present anywhere in origin/main                  = 0
carried onto the reset branch as semantic paths  = 15
carried nowhere                                  = 14   (~128 KB)
```

The 14 uncarried files exist on exactly one ref,
`refs/heads/docs/a8-authz-approval-redesign-ledger`: `rebaseline-decision-registry.md`
(39,014 B), its eight closure amendments (48,659 B),
`r10-post-t6-implementation-readiness-program.md` (15,443 B),
`r10-technical-architecture.md` (10,156 B),
`r10-technical-realization-reconciliation-baseline.md` (8,987 B),
`launch-v1-scope-rebaseline.md`, `cohesive-platform-redesign.md`.

Two surviving authorities additionally cite ratified Git blobs by SHA:

```text
docs/product/journeys.md:10         blob 5f3f0ec93bf94f586eafd341d72ec484ef2ec848  45,434 B
docs/architecture/transition.md:10  blob 9ae3cce4b25d6824a45bbb4872d21e558f6c6763   1,754 B
docs/architecture/transition.md:11  blob cfda127151d55c2de28737fc4e692d1b5bf603fa   5,005 B
```

All three verified present; all three verified **unreachable** from `origin/main` and from
the reset branch, reachable only from the PR #131 branch.

`docs/status.md:31` schedules "close superseded PR #131 / #132" as the action immediately
after merge. Deleting that branch makes every object above unreachable and GC-eligible.

**Property at risk.** Git-as-rollback/provenance — the reset's own central safety premise,
and the sole justification offered for deleting the decision corpus.

**Smallest correction.** Before closing or deleting the PR #131 and #132 branches, create
annotated tags on `d8b1c6d3` (PR #131) and `b0ebe54c` (PR #132) using the convention this
repository already has (`refs/tags/archive/main-pre-qa-20260614`,
`refs/tags/archive/phase10-integration`), and record both tag names in
`docs/decisions/repository-reset.md`. No file content moves; no removed file returns.

**Upstream reopen required:** no.

### BLOCKER-2 — durable navigation is routed through the directory this gate must delete

**Claim.** Three durable pages route through `docs/work/current/`, which governance
requires deleting before merge, and `docs/reference/` has no durable route at all. Merge
orphans the T8-E checkpoint.

**Evidence.**

```text
docs/decisions/index.md:30  | Executable API wire contract | active | ../work/current/proposal.md |
docs/index.md:34            | Active T8-E work | Current work (work/current/index.md) |
README.md:12                - [Active work](docs/work/current/index.md)
```

`docs/work/current/proposal.md` is the **repository-reset** proposal, not the executable
wire contract — the registry row names the wrong subject today and dangles tomorrow.
`docs/development/documentation.md` §Active work: "Temporary work files are deleted before
a governance/architecture PR is merged." `AGENTS.md` §Git safety: "Do not merge a
governance/architecture PR while temporary `docs/work/**` review artifacts remain."

`docs/reference/t8e-checkpoint.md` is named by exactly two files in the tree —
`docs/work/current/index.md:19` and `docs/work/current/ai-dialog.md:21` — both deleted at
merge. It appears in no `docs/index.md` row, no `AGENTS.md` routing row, no `mkdocs.yml`
nav entry and no decision-registry row.

**Property at risk.** Proposal acceptance test 4 ("the accepted T8-E checkpoint is
preserved and explicitly paused, not lost"); `docs/status.md:32` "restore T8-E checkpoint
as current work".

**Smallest correction.** Point `docs/decisions/index.md:30` at
`../reference/t8e-checkpoint.md`; add a `docs/index.md` row for the paused checkpoint;
make the `README.md` and `docs/index.md` entries for `work/current/` conditional on that
directory existing, or drop them — `AGENTS.md:5` already states the conditional correctly.

**Upstream reopen required:** no.

### BLOCKER-3 — two conflicting operation censuses; the amendment lives only in a `kind: work` file

**Claim.** The tree ships a ratified **closed 76-operation** census and an unratified
**78-operation** census, with no amendment recorded in the authority.

**Evidence.** `docs/product/journeys.md:1301` — "## 29. Closed Launch `/api/v1` operation
census" — enumerates 76 operations (verified by count: 3 session + 24 organization + 4
authorization + 10 governance-config + 34 controlled-documents + 1 audit) and states that
new route families require a named Product Contract journey or an explicit T6 reopen.

`docs/reference/t8e-checkpoint.md:32-39`:

```text
The original T6 census contained 76 operations. A bounded precision amendment adds:
GET /api/v1/users/{user_id}/profile         getUserProfile
GET /api/v1/areas/{area_id}/lifecycle       getAreaLifecycle
Current census: 78 operations.
```

Neither operation appears in §29 — journeys carries `PUT`/`DELETE .../profile` and
`PUT .../lifecycle`, no `GET` on either. Adding an operation is not the "minor path
spelling/operationId normalization" §29 permits. In PR #131 such amendments were recorded
in dedicated `rebaseline-decision-registry-*-amendment.md` files; none were carried. The
`+2` therefore survives only inside a `kind: work` document that BLOCKER-2 orphans.

**Property at risk.** One meaning, one authority; the "78 operations" input contract for
the next design layer at `docs/reference/t8e-checkpoint.md:202`.

**Smallest correction.** Add a one-line amendment note to `docs/product/journeys.md` §29
recording the two added operations and the current count of 78, citing
`docs/reference/t8e-checkpoint.md`.

**Upstream reopen required:** no — the amendment is already carried as accepted by the
checkpoint's own status; only its recording in the owning authority is missing.

### MAJOR-1 — the stage program that `docs/status.md` and five survivors depend on has no successor

**Claim.** The live tree asserts a stage roadmap it cannot define.

**Evidence.** `docs/architecture/transition.md:8` declares
`> **Program authority:** wiki/architecture/r10-post-t6-implementation-readiness-program.md`.
That file (15,443 B, uncarried) is where T8-A…T8-H, T9 Golden Flows & Validation Baseline,
T10 Transition/Cutover, T11 Implementation Program & Execution Graph and T12 Adversarial
Implementation-Readiness are defined, together with the close order and the rule that only
after T12 may implementation execute the T11 graph.

`docs/status.md:17` asserts "T8-F → T12 NOT OPEN". Survivors reference
T8-F/T8-G/T8-H/T9/T10/T11/T12 as decision owners at 20+ sites across `backend.md`,
`interfaces.md`, `persistence.md`, `technical-baseline.md` and `transition.md`. The live
tree defines none of them; `T8-H` and `T12` have zero definitional coverage anywhere.
`docs/architecture/persistence.md:1689-1690` splits the range as `T8-F→T8-H` plus
`T9→T12`, which cannot be reconciled with status's single `T8-F → T12` band from the tree
alone.

This also trips the reset's own boundary: `AGENTS.md:7` forbids reconstructing context from
Git history or closed PRs "unless a current authority names that evidence as necessary" —
and `transition.md:8` names it.

**Property at risk.** Deterministic agent routing after T8-E ratification; the
no-archaeology rule.

**Smallest correction.** Carry the stage definitions into one durable page — a ~15-line
stage table in `docs/status.md`, or `docs/decisions/stage-program.md` — then repoint
`transition.md:8` at it.

**Upstream reopen required:** no.

### MAJOR-2 — the ratified decision-disposition authority was replaced by a page that inherits its name but not its function

**Claim.** `docs/decisions/index.md` is called the decision registry but carries none of
the registry's forward obligations.

**Evidence.** `wiki/architecture/rebaseline-decision-registry.md` (39,014 B, uncarried) is
headed "ACTIVE / OPERATOR-RATIFIED DECISION DISPOSITION AUTHORITY" and carries 171
dispositioned decision rows:

```text
CURRENT 82 | PRESERVE 21 | REFINED 12 | REOPEN 4 | DEFERRED 27 | SUPERSEDED 25
```

plus eight closure amendments (48,659 B) and an operator-ratified consumption law: "Before
every remaining T-stage: read this registry → consume CURRENT/PRESERVE/REFINED as baseline
→ deliberately design only that T-stage's REOPEN set → keep DEFERRED as future
counterexample/seam → reject SUPERSEDED inheritance."

`docs/decisions/index.md` is a 17-row pointer table with no decision IDs, no dispositions
and no consumption law. Three survivors still cite the deleted registry as their decision
baseline: `async-and-search.md:8`, `authorization-and-audit.md:8`,
`content-integrity.md:8`.

Applying the Structural Inversion Test honestly: the 82 CURRENT rows are genuinely
redundant — their content sits inside the 15 carried authorities. The 25 SUPERSEDED rows
are largely self-neutralizing, because the implementation and documentation that could
have inherited those shapes are exactly what this reset deletes. The forward obligations
are not neutralized by any deletion: 4 REOPEN (DOC-12 numbering grammar, CNT-03 editor
lease, AUD-06 retention, MIG-10 import families), 27 DEFERRED seams, 21 PRESERVE
baselines. Spot-check of survivor coverage: `Keycloak`, `River`, `EigenPal`, `Dossier`,
`Acknowledgement`, `numbering` and `EditorSession` all appear in carried authorities;
`LegalHold` and `ObjectLock` — FUT-05, "LegalHold business preservation != ObjectLock /
provider physical enforcement" — have zero occurrences in the entire live tree.

Method §Complexity law: YAGNI "MUST NOT remove a known invariant, safety property, …
or a seam justified by evidenced evolution." Method §1: "Summaries … are derived aids …
MUST NOT become a second authority." Here a summary replaced the authority and took its
name.

**Property at risk.** The ratified input contract for every remaining T-stage, starting
with the T8-E resumption this reset is clearing the way for.

**Smallest correction.** Carry the 52 REOPEN + DEFERRED + PRESERVE rows into
`docs/decisions/index.md` or a sibling durable page; restate the consumption law in
`docs/development/engineering-rules.md`; repoint the three "Decision baseline" citations.
Deliberately dropping the CURRENT and SUPERSEDED rows is defensible and should be stated
as a decision rather than left as an omission.

**Upstream reopen required:** no.

### MAJOR-3 — the repository-shape guard is a legacy-name denylist and cannot fire on new implementation

**Claim.** CI protects the absence of fourteen specific old directory names, not the
invariant "architecture-first, implementation blocked".

**Evidence.** `.github/workflows/ci.yml:16-35` asserts `test ! -e` for `wiki`,
`docs/superpowers`, `api`, `apps`, `cmd`, `db`, `deploy`, `frontend`, `internal`,
`packages`, `scripts`, `tests`, `tools`, `vendor`.

`origin/main`'s root also contains `sql`, `ops`, `third_party`, `archive`, `tasks`,
`reviews`, `scratch_qa`, `test-results`, `.claude`, `.superpowers`, `.qa-reports`,
`Makefile`, `go.mod`, `package.json` and `pnpm-workspace.yaml` — all removed by this
reset, none covered by the guard. Restoring any of them passes CI, as would any new
`src/`, `pkg/`, `web/`, `migrations/` or `openapi/`. The proposal's own "delete by
construction" list names `.claude/`, `.superpowers/` and `.qa-reports/`; the guard checks
none of the three.

`git diff --check` runs on a fresh `actions/checkout` with a clean worktree, so it compares
worktree to index and is always empty. Method §Enforcement: "A control that cannot be shown
to fire is not proven."

CI also does not check `docs/work/**` absence, so the merge precondition BLOCKER-2 depends
on has no mechanical enforcement at all.

**Property at risk.** The implementation gate itself — the single invariant this repository
currently exists to hold.

**Smallest correction.** Invert to an allowlist: fail if any tracked path falls outside
`{AGENTS.md, CLAUDE.md, README.md, mkdocs.yml, .gitignore, .gitattributes, .github/**,
docs/**}`, or if any tracked file is not `.md`/`.yml`/`.yaml`; add `test ! -e docs/work`;
drop `git diff --check`. This is strictly shorter than the current list and complete over
all paths that can reach the protected state.

**Upstream reopen required:** no.

### LOW-1 — three survivors contradict the sole stage authority

`docs/status.md:16`, the declared sole stage/implementation-gate authority, says T8-E is
"PAUSED AT APPROVED CHECKPOINT". Contradicted by `README.md:18` "T8-E executable API
contract ACTIVE", `docs/decisions/index.md:30` status "active", and
`docs/reference/t8e-checkpoint.md:10` "> Active work." Separately, `docs/status.md:31`
lists "close superseded PR #131 / #132" as a future action; both are already CLOSED.
Correction: say PAUSED in all three; reword the status step to "tag and delete the PR #131
/ #132 branches" per BLOCKER-1.

**Upstream reopen required:** no.

### LOW-2 — `mkdocs.yml` is a survivor with no named consumer, and it is broken

No CI job builds documentation, no Python dependency manifest exists, no pages workflow
exists. It is a fourth routing surface after `AGENTS.md`, `docs/index.md` and
`docs/decisions/index.md`, and it already drifts: nav omits `docs/reference/` entirely.
With no `docs_dir` key, mkdocs defaults to `docs_dir: docs`, so every nav entry
`docs/x.md` resolves to `docs/docs/x.md` and the build fails on first use. The reset's own
survival rule is "only current truth with a named consumer survives". Correction: delete
it, or set `docs_dir: .` and add the checkpoint to nav.

**Upstream reopen required:** no.

### LOW-3 — an externally binding contract survived undocumented

Ruleset `20560142` (active, `~DEFAULT_BRANCH`) requires exactly one status context,
`required`, satisfied by the job id `required` at `.github/workflows/ci.yml:11` — verified
SUCCESS on PR #134. `.github/rulesets/main.json`, the tracked record of that binding, is
deleted and nothing in the survivor set records it; renaming that job silently un-gates
`main`. This is the one genuinely external contract review question 2 asked for, and it is
not in the removed implementation — it is in what survived.

The same ruleset sets `required_review_thread_resolution: true` with
`required_approving_review_count: 0`, so unresolved bot threads block merge independently
of checks — CodeRabbit is installed and has already commented on PR #134 — while
`docs/development/engineering-rules.md` §Closure names only "required checks". Correction:
two lines in `engineering-rules.md` recording the `required` context name and the
thread-resolution precondition.

**Upstream reopen required:** no.

### LOW-4 — `docs/reference/` is an undeclared class holding a `kind: work` file

`docs/development/documentation.md` §Navigation and §Active work enumerate
`docs/index.md`, `docs/status.md`, `docs/decisions/index.md`, `mkdocs.yml` and
`docs/work/`. `docs/reference/t8e-checkpoint.md` carries `kind: work` while sitting
outside `docs/work/`, where `work` is defined as temporary material inside a Draft PR that
is deleted before merge. Correction: name `docs/reference/` in the governance page and
state that `kind: work` outside `docs/work/` denotes a paused checkpoint that survives
merge.

**Upstream reopen required:** no.

### LOW-5 — stale routing assertions in body prose, outside the reset's own waiver

The waiver in `docs/development/documentation.md` covers "internal title/status/provenance
block". These three sites are body prose and fall outside it:

```text
docs/product/alignment.md:10   "current technical-stage routing is now owned by
                                wiki/architecture/r10-technical-architecture.md"
docs/product/alignment.md:197  "This sequence is now owned by ...r10-technical-architecture.md"
docs/product/contract.md:481   "The active technical architecture is routed by ...r10-technical-architecture.md"
```

That file is deleted; its function is now `docs/status.md`. In total 25 `wiki/` citations
survive across 12 files; 9 of them target the three uncarried files. The remaining
provenance-header citations may stand under the waiver. Correction: repoint the three
body-prose sites at `docs/status.md`.

**Upstream reopen required:** no.

### LOW-6 — a secret-scanning control was retired without a recorded reopen trigger

`.gitleaks.toml` and `.gitleaksignore` are deleted from a public repository with no
replacement control. Retirement is proportionate right now — the live tree is 32 Markdown
files with no secret surface — but `docs/development/engineering-rules.md` §Closure
requires the retire-or-replace reasoning to be explicit. Correction: one line under
`docs/decisions/repository-reset.md` §Reopen triggers, restoring secret scanning before
the first implementation commit.

**Upstream reopen required:** no.

### Attacked and found sound

Recorded so these are not re-litigated, with what would have falsified each.

**Q1 — current authority preservation.** All fifteen ratified Product/R10 authorities are
present under semantic paths and byte-substantial (`persistence.md` 50,936 B,
`interfaces.md` 47,899 B, `journeys.md` 44,217 B). What is missing is the registry and the
stage program, handled as MAJOR-1 and MAJOR-2, not the authorities themselves. Falsifier
would have been any T1→T8-D subject with no semantic-path successor; none found.

**Q2 — externally binding compatibility contracts in the deleted implementation.** None
found. The deleted runtime is DEV/test with no historical-business consumer, and
`docs/architecture/transition.md:70` — "For Launch, the T10 historical-business-import
branch is empty" — supports the decision directly. The repository license was already
`null` on `origin/main`, so its absence is not a reset regression. The only external
binding located is LOW-3, which lives in the survivor set rather than the removed tree.

**Q3 — source-tree cleanup before T10 runtime cutover.** Coherent.
`docs/architecture/transition.md:57-70` and `docs/architecture/technical-baseline.md:46`
both keep T10 mandatory for technical transition while confirming the import branch is
empty. A source-tree reset and a runtime/data cutover are different objects, and the
decision says so explicitly.

**Q4 — superseded architecture asserted as current.** None found. `tenant`/RLS/GUC,
local-password authentication, peer Approval/Templates/Taxonomy/Tokens modules and the
legacy frontend appear only inside explicit DELETE/REJECT dispositions —
`persistence.md:1515`, `:1600`, `technical-baseline.md:77-98`,
`journeys.md:47,237,1442-1450`, `backend.md:524`. `NoHumanApproval` is current governance
vocabulary, not the retired Approval module.

**Q5 — CI sufficiency, mechanical half.** The workflow does produce the exact context the
active ruleset requires, and PR #134 shows it green. The insufficiency is semantic, and is
MAJOR-3.

**Q6 — routing determinism.** `AGENTS.md` and `docs/index.md` are keyed differently — need
versus subject — and do not contradict each other. `AGENTS.md:5` states the
`docs/work/current/` conditional correctly, which is precisely what `docs/index.md:34` and
`README.md:12` fail to do (BLOCKER-2). `CLAUDE.md` correctly disclaims authority.

**Q7 — checkpoint resumability.** The checkpoint is substantively complete: four approved
layers, creation-result rules, the open measurement obligation, and a ten-item next design
layer. Its one external dependency — the 76-operation census it amends — survives at
`docs/product/journeys.md:1301`. Its resumability failures are BLOCKER-2 (reachability)
and BLOCKER-3 (the unrecorded amendment), not its content.

**Q8 — further subtraction.** Only `mkdocs.yml` fails the survival test (LOW-2).
`CLAUDE.md` correctly survives: Claude Code loads it automatically and cannot be repaired
in this gate, which is exactly the external-consumer exception
`docs/development/documentation.md` §"One meaning, one authority" allows. The
implementation-era entries in `.gitignore` (`node_modules/`, `dist/`, `coverage/`) are
inert and not worth a change.

**Q10 — better Global Maximum.** None found; see the structural verdict above.

### Verdict

```text
APPROVE CLEAN-SLATE REPOSITORY RESET WITH MATERIAL FIXES
```

The structure is right and should not be reopened. Three blockers must land in this gate,
because merge — or the branch cleanup `docs/status.md` schedules immediately after it —
makes each one irreversible: unmerged provenance becomes GC-eligible once the PR #131
branch goes, the T8-E checkpoint is orphaned once `docs/work/current/` goes, and the
78-operation amendment goes with it. The three MAJORs are decision-carrying gaps that the
T8-E session would otherwise rediscover as archaeology. None of the ten findings requires
restoring a removed legacy file, and none requires an upstream reopen.

## Lead adjudication

## Bounded round 2

Use only if a material contradiction survives adjudication.

## Operator decision
