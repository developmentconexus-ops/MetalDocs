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

```text
PRIMARY VERDICT
APPROVE REPOSITORY DOCUMENTATION PROFILE WITH MATERIAL FIXES

BLOCKER 3
MAJOR   8
LOW     6
```

Fable confirmed the structural direction—one `docs/` root, semantic names/navigation, one proposal + AI dialogue, one coherent PR gate, and Git/closed PRs as archive—but found bounded execution defects in retention, parity proof, agent-context preservation, and merge-ready enforcement.

## Lead adjudication

Reviewer output is evidence, never authority. Lead selected these dispositions after confronting the findings with the reviewed repository state.

| Finding | Disposition | Selected correction |
|---|---|---|
| B1 machine-consumed `wiki/` subjects | ACCEPT | complete document disposition; explicit homes for ADR/database/problem/trace docs; gate-subject repoint-or-retire invariant |
| B2 false-green R10 census | ACCEPT WITH STRENGTHENING | closed source→target map + filesystem census + real normative vocabulary + empty-source failure |
| B3 Ready-state guard cannot fire | ACCEPT | `docs/work/**` Draft-only + CI `ready_for_review` trigger before using it as merge protection |
| M1 zero textual old-path rule | ACCEPT | repair live executable consumers only; provenance/history citations do not force churn |
| M2 `.gitleaksignore` historical paths | ACCEPT | preserve history-pinned fingerprints while full-history scanner consumes them |
| M3 generated report/frontmatter conflict | ACCEPT | generator emits durable frontmatter; no hand-edits to generated output |
| M4 provenance disappears with work cleanup | ACCEPT | durable authority records G0 provenance; future decision registry retains compact row |
| M5 AGENTS/CLAUDE safety truth would vanish | ACCEPT | move repository law to `docs/development/engineering-rules.md`; current orientation to `docs/reference/current-system.md` before shrinking bootstrap |
| M6 local proof weaker than CI | ACCEPT | mirror `--require-infra`, lint, security, and affected non-PR governance checks |
| M7 `authority + active` ambiguity | ACCEPT WITH SUBTRACTION | remove `status`; closed kinds = `authority | work` |
| M8 existing `docs/runbooks/**` / `docs/engineering/**` omitted | ACCEPT | complete `git ls-files '*.md'` disposition census; undispositioned path blocks G1 |
| L1 pushed-branch rebase contradiction | ACCEPT | no force-rebase merely to refresh base; merge updated main or use next clean branch |
| L2 brittle fixed `git rm` | ACCEPT BY REMOVAL | derive deletions from disposition census |
| L3 ADR navigation conflict | ACCEPT | governed collection index satisfies reachability for large collections |
| L4 unrelated `.claude` permission cleanup | ACCEPT | remove from gate |
| L5 fixture can fail for wrong reason | ACCEPT | fixture includes minimal valid nav/tree and asserts intended finding |
| L6 MkDocs semantics overstated | ACCEPT PRECISION | `mkdocs.yml` is navigation manifest; publishing is out of scope |

### Lead result

```text
Fable blockers closed by selected corrections    3 / 3
Fable majors accepted/corrected                   8 / 8
Fable lows accepted/precision-subtracted          6 / 6

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

The corrections are materialized in:

```text
docs/development/documentation.md
docs/work/current/proposal.md
docs/work/current/plan.md
```

No legacy deletion, Product/R10 authority migration, G1 implementation, or T8-E resumption has begun.

## Bounded round 2

Not required. No material contradiction survives Lead adjudication, and no correction reopens the selected documentation root, semantic naming model, or retention predicate.

## Operator decision

**PENDING — explicit operator ratification of the corrected G0 profile is the next gate.**
