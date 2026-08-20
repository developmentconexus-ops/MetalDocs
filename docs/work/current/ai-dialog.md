---
id: work-ai-dialog
kind: work
owner: architecture
summary: Temporary review, adjudication, and operator-ratification record for the repository documentation profile.
---

# AI dialogue

> **TEMPORARY / NON-AUTHORITATIVE / DELETE BEFORE MERGE**

## Review context

```text
Repository: developmentconexus-ops/MetalDocs
Branch: docs/repository-information-architecture
PR: #132
Fable reviewed HEAD: 8eb2e70d11917362669f279f5183ae8366759e99
Review commit / first post-review HEAD: 3b8a25488e1aed5edc6c2b83d64e802b8d66c1c0
Product implementation: not authorized
Legacy deletion: not started
PR #131: frozen provenance only
```

Review target:

```text
docs/development/documentation.md
```

Canonical method and review workflow:

```text
developmentconexus-ops/conexus-methodology/METHOD.md
developmentconexus-ops/conexus-methodology/README.md
```

## Fable review

```text
PRIMARY VERDICT:
APPROVE REPOSITORY DOCUMENTATION PROFILE WITH MATERIAL FIXES

BLOCKER  3
MAJOR    8
LOW      6

Direction / one docs root                         CONFIRMED
Naming and navigation model                      CONFIRMED
One proposal + one AI dialogue                   CONFIRMED
One coherent ratifiable gate per PR              CONFIRMED
Profile as originally written                    NOT YET PROMOTABLE
Lead adjudication                                MAY PROCEED
Another broad review round                       NOT REQUIRED
```

The complete independent review remains in Git history. Its material findings were:

| ID | Finding |
|---|---|
| B1 | Deleting `wiki/` would delete machine-consumed documentation without naming successor homes or repointing its verification gates. |
| B2 | The planned R10 parity census used `git grep` over ignored untracked input and could report an empty false-green. |
| B3 | The Draft/Ready temporary-work guard would not run on `ready_for_review`. |
| M1 | A zero-occurrence grep would expand G1 into code, OpenAPI, migration, and generated-file churn. |
| M2 | History-pinned secret-scan allowlist entries must survive old-path deletion. |
| M3 | Generated requirement traceability and mandatory frontmatter were mutually unsatisfiable. |
| M4 | The durable page would retain dangling temporary references and no durable promotion provenance. |
| M5 | Load-bearing repository rules/current-runtime orientation lacked successor owners before slimming `AGENTS.md`/`CLAUDE.md`. |
| M6 | The local proof ladder did not mirror the CI ladder. |
| M7 | `kind: authority` + `status: active` was undefined and mechanically undetectable. |
| M8 | Existing `docs/runbooks/**` and `docs/engineering/**` were outside both retention and deletion sweeps. |
| L1 | Rebase-without-rewriting wording contradicted the no-force-push rule. |
| L2 | Some `git rm` targets were absent or mislocated. |
| L3 | ADR navigation needed a directory/index rule. |
| L4 | `.claude/settings.json` permission removal was outside this gate. |
| L5 | The docs-hygiene negative fixture could fail for the wrong reason. |
| L6 | `mkdocs.yml` is currently a navigation manifest, not a deployed documentation platform. |

## Lead adjudication

Reviewer output is evidence, not authority. The Lead adjudicated every finding as follows.

| ID | Disposition | Corrected decision |
|---|---|---|
| B1 | **ACCEPT** | Machine-consumed documentation is a retained class. ADRs, database dictionary, problem codes, requirement traceability, and other gate subjects receive explicit target homes. No gate subject may be deleted unless that gate is repointed or retired and its negative proof is re-established in the same PR. |
| B2 | **ACCEPT** | R10 parity uses a closed source→target map, semantic comparison, a non-empty normative census over actual source files, and an explicit failure when source extraction is empty. |
| B3 | **ACCEPT** | `docs/work/**` may exist only while a PR is Draft; CI must subscribe to `ready_for_review`, and the Ready transition must produce a fresh required run. |
| M1 | **ACCEPT** | G1 repairs **executable path consumers**, not mere provenance citations in comments, applied migrations, generated history, or history-pinned records. |
| M2 | **ACCEPT** | History-pinned security allowlists are preserved unless the history-scanning gate itself proves a safe replacement. |
| M3 | **ACCEPT** | Generated maintained pages receive metadata from their generator; generated output is never hand-edited. |
| M4 | **ACCEPT** | Temporary links are removed before merge. Promotion provenance is recorded in the durable authority and then indexed in `docs/decisions/index.md` during G1. |
| M5 | **ACCEPT** | Before `AGENTS.md`/`CLAUDE.md` are reduced, repository engineering law moves to `docs/development/engineering-rules.md` and current-runtime orientation moves to `docs/reference/current-system.md`. |
| M6 | **ACCEPT** | The proof ladder mirrors required CI and uses `--require-infra`; the Go lint job is proved separately where CI does so. |
| M7 | **ACCEPT WITH SUBTRACTION** | Metadata is reduced to `id`, `kind`, `owner`, and `summary`; `kind` is only `authority | work`. Candidate state belongs to the Draft work lifecycle, not durable authority metadata. |
| M8 | **ACCEPT** | G1 begins from complete `git ls-files '*.md'` enumeration. Every first-party Markdown path receives `KEEP → target` or `DELETE`; an undispositioned path is a hard stop. |
| L1 | **ACCEPT** | A pushed branch incorporates new `main` through a normal merge unless a separately authorized single-owner rewrite policy applies. |
| L2 | **ACCEPT** | Deletion commands operate only on census-confirmed paths or use safe ignore-unmatched behavior. |
| L3 | **ACCEPT** | ADR pages are reachable through `docs/decisions/index.md`; they need not each appear as top-level MkDocs navigation entries. |
| L4 | **ACCEPT** | `.claude/settings.json` changes are removed from this gate. |
| L5 | **ACCEPT** | Negative fixtures include a minimal otherwise-valid documentation tree so the intended rule is the demonstrated failure. |
| L6 | **ACCEPT** | `mkdocs.yml` is explicitly a navigation manifest in the current baseline; publishing/Backstage deployment is not introduced. |

### Corrected Global Maximum

```text
ONE docs/ ROOT
+
SEMANTIC STABLE FILENAMES
+
MINIMAL AUTHORITY | WORK METADATA
+
INTENT-BASED INDEX + EXPLICIT NAVIGATION MANIFEST
+
SHORT AGENT BOOTSTRAP WITH DURABLE RULE/ORIENTATION OWNERS
+
ONE TEMPORARY PROPOSAL + ONE TEMPORARY AI DIALOGUE
+
ONE COHERENT RATIFIABLE GATE PER PR
+
GIT / CLOSED PRs AS ARCHIVE
+
COMPLETE TRACKED-DOCUMENT CENSUS
+
EXPLICIT KEEP→TARGET OR DELETE DISPOSITION
+
VERIFICATION-GATE SUBJECT PRESERVATION
+
EXECUTABLE-CONSUMER PATH REPAIR
-
TWO DOCUMENTATION ROOTS
-
HAND-WRITTEN ROOT ALLOWLISTS
-
TEXT-OCCURRENCE CLEANUP ACROSS CODE/HISTORY
-
PER-ROUND PERMANENT REVIEW ARTIFACTS
-
LIVE-TREE ARCHIVES
-
DUPLICATED STATUS / AUTHORITY
```

No Product/R10 decision was reopened. No second Fable round is materially required because every correction remains inside the selected information architecture and closes a defect against its own invariants.

## Bounded round 2

```text
NOT REQUIRED
```

## Operator decision

```text
APPROVED
Operator ratification date: 2026-08-20
```

The operator ratifies the corrected repository documentation profile and authorizes the sequence:

```text
S0 trustworthy verification baseline
→ finish and squash-merge G0 / PR #132
→ G1 complete census, authority consolidation, gate/consumer repair,
  legacy deletion, and documentation verifier
→ close PR #131 as superseded provenance after parity proof
→ resume T8-E in a fresh small PR
```

This approval does **not** authorize G1 deletion inside PR #132, product implementation, Product/R10 semantic changes, PR #131 merge, or T8-E resumption before the clean baseline is established.