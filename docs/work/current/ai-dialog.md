---
id: work-ai-dialog
kind: work
owner: architecture
summary: Temporary Fable review and Lead adjudication record for repository documentation governance.
---

# AI dialogue

> **TEMPORARY / NON-AUTHORITATIVE / DELETE BEFORE MERGE**

## Review context

```text
Repository: developmentconexus-ops/MetalDocs
Branch: docs/repository-information-architecture
PR: #132
Fable-reviewed HEAD: 8eb2e70d11917362669f279f5183ae8366759e99
Fable review commit / post-review HEAD: 3b8a25488e1aed5edc6c2b83d64e802b8d66c1c0
Review target: docs/development/documentation.md
Product implementation: not authorized
Legacy deletion: not started
PR #131: frozen provenance only
```

The full independent Fable review is preserved in Git at commit `3b8a25488e1aed5edc6c2b83d64e802b8d66c1c0` and is not duplicated here after adjudication.

## Fable review summary

Primary verdict:

```text
APPROVE REPOSITORY DOCUMENTATION PROFILE WITH MATERIAL FIXES

BLOCKER 3
MAJOR   8
LOW     6
```

Fable confirmed the structural direction:

```text
one docs/ root                     YES
semantic naming/navigation         YES
one proposal + one AI dialogue     YES
one coherent ratifiable PR gate    YES
Git/closed PRs as archive          YES
```

The profile as reviewed was not yet promotable because deletion/retention, parity proof, and merge-ready enforcement contained bounded defects. Fable reported that another independent round was not materially required if Lead corrections stayed inside the selected information architecture.

## Lead adjudication

Reviewer output is evidence, never authority. Lead revalidated the findings against the reviewed repository state and selected the following dispositions.

### BLOCKER-1 — verification subjects under `wiki/`

**Disposition: ACCEPT.**

Root cause is valid: the previous plan expressed a retention predicate as a root allowlist and omitted machine-consumed documentation.

Selected correction:

- machine-consumed documentation is a first-class retained class until its gate is moved/retired;
- target homes include ADRs, database reference pages, problem-code reference and requirement traceability;
- module-debt material survives only with its gate;
- hard invariant: no verification gate subject is deleted without repoint/retire + proof in the same PR;
- G1 performs a complete document disposition census before any root deletion.

No Product/R10 reopen.

### BLOCKER-2 — false-green R10 parity census

**Disposition: ACCEPT WITH STRENGTHENING.**

The reported `git grep` behavior over ignored `.tmp/` is a real false-green risk.

Selected correction:

- closed source-to-target authority map is the primary parity mechanism;
- temporary extracted source is scanned with filesystem-capable `grep`/`rg`, not tracked-file search;
- normative vocabulary is widened to actual repository terms;
- empty source census is an explicit failure;
- textual census remains supporting evidence and cannot replace semantic parity review.

No Product/R10 reopen.

### BLOCKER-3 — merge-ready work-file guard cannot fire

**Disposition: ACCEPT.**

Selected correction:

- `docs/work/**` may exist only while the PR is Draft;
- GitHub Actions must include `ready_for_review` among `pull_request` activity types before this is treated as a merge guard;
- temporary work is deleted before marking Ready.

No upstream reopen.

### MAJOR-1 — zero textual old-path rule explodes G1 scope

**Disposition: ACCEPT.**

Selected correction:

```text
EXECUTABLE CONSUMER
→ repair or retire

PROVENANCE/HISTORY CITATION
→ no forced churn
```

G1 classifies occurrences instead of requiring zero textual references to old paths. Product/OpenAPI/runtime files are not touched merely to rewrite historical comments or citations.

### MAJOR-2 — `.gitleaksignore` historical fingerprints

**Disposition: ACCEPT.**

History-pinned security allowlist entries may continue referencing deleted historical paths while full-history scanning consumes them. No security gate is weakened to make path cleanup visually complete.

### MAJOR-3 — generated requirement report vs frontmatter

**Disposition: ACCEPT.**

Generated durable pages receive required metadata from their generator. Generated output is never hand-edited to satisfy documentation hygiene.

### MAJOR-4 — dangling temporary links / ratification provenance

**Disposition: ACCEPT.**

Selected correction:

- durable authority no longer links to temporary proposal/plan as current related documents;
- authority records G0 provenance directly until the compact decision registry exists;
- the promoting PR writes/retains decision provenance before temporary files are deleted;
- Git preserves the full review record.

### MAJOR-5 — current `AGENTS.md` / `CLAUDE.md` safety truth would evaporate

**Disposition: ACCEPT.**

Selected durable owners before bootstrap reduction:

```text
docs/development/engineering-rules.md
docs/reference/current-system.md
```

`AGENTS.md` routes to them instead of carrying their full prose. `CLAUDE.md` remains a minimal pointer only after its current load-bearing orientation has moved.

### MAJOR-6 — local proof ladder weaker than CI

**Disposition: ACCEPT.**

G0/G1 proof mirrors CI:

```text
go test ./...
go run ./tools/verify --require-infra --profile=pr
go run ./tools/verify --require-infra --only=golangci-lint
```

Affected non-PR/nightly governance checks are run explicitly when their documentation subjects move.

### MAJOR-7 — undefined `authority + active` metadata state

**Disposition: ACCEPT WITH SUBTRACTION.**

Remove the `status` field entirely. Closed kinds become:

```text
authority
work
```

A durable live page is current for its declared subject; temporary candidacy is represented by Draft-PR `work` files. This removes duplicate lifecycle metadata rather than adding another rule.

### MAJOR-8 — existing `docs/runbooks/**` / `docs/engineering/**` omitted

**Disposition: ACCEPT.**

G1 starts from a complete tracked Markdown census (`git ls-files '*.md'`) plus machine-consumed doc-like subjects. Every path gets explicit `KEEP`, `MERGE`, `GENERATED`, or `DELETE` disposition. Undispositioned paths stop the gate.

### LOW-1 — rebase without rewriting history

**Disposition: ACCEPT.**

Do not rebase/force-push shared pushed PR history merely to refresh base. Merge updated `main` or use the next clean branch according to repository Git policy.

### LOW-2 — brittle `git rm` unmatched paths

**Disposition: ACCEPT BY REMOVING THE DEFECT CLASS.**

G1 deletion commands are derived from the complete disposition census, not a fixed multi-path `git rm` command.

### LOW-3 — ADR navigation conflict

**Disposition: ACCEPT.**

Large governed collections may be reachable through a collection index. Individual ADRs/table pages do not each require top-level `mkdocs.yml` entries, but they must be reachable from the navigation graph.

### LOW-4 — `.claude/settings.json` permission cleanup

**Disposition: ACCEPT.**

Removed from this gate. Unrelated tool-permission cleanup needs its own justified change.

### LOW-5 — negative fixture can fail for wrong reason

**Disposition: ACCEPT.**

Docs-hygiene fixtures must include the minimum valid navigation/repository context and assert the intended finding, not merely non-zero exit.

### LOW-6 — MkDocs build semantics not actually used

**Disposition: ACCEPT PRECISION.**

`mkdocs.yml` is the explicit navigation manifest. Static publishing is not part of G0/G1; build-specific keys are future compatibility only.

## Lead result

```text
Fable blockers accepted and closed by selected corrections: 3 / 3
Fable majors accepted/accepted-with-correction:              8 / 8
Fable lows accepted/precision-subtracted:                    6 / 6

ONE docs/ ROOT                              CONFIRMED
SEMANTIC NAMES                              CONFIRMED
AGENT ROUTING MODEL                         CONFIRMED AFTER CORRECTION
ONE AI DIALOGUE                             CONFIRMED
ONE COHERENT PR GATE                        CONFIRMED
COMPLETE-DISPOSITION DELETION               SELECTED
GATE-SUBJECT PRESERVATION                   SELECTED
EXECUTABLE-CONSUMER PATH REPAIR             SELECTED
GLOBAL MAXIMUM                              CONFIRMED AFTER CORRECTIONS
PRODUCT/R10 REOPEN                          NO
SECOND FABLE ROUND                          NOT REQUIRED
```

The corrected durable candidate is `docs/development/documentation.md`. The corrected temporary proposal and execution plan implement the same decisions. No deletion or G1 work has started.

## Bounded round 2

Not required. No material contradiction survives Lead adjudication, and no correction reopens the selected documentation root, semantic naming model, or retention predicate.

## Operator decision

Pending explicit operator ratification of the corrected G0 profile.
