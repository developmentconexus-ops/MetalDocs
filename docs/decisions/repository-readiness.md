---
id: repository-readiness
kind: authority
owner: engineering
summary: Repository closeout and readiness decision for resuming MetalDocs Product planning after the 2026-08-25 governance rebaseline.
---

# Repository readiness closeout

> **Status:** READY TO RESUME PRODUCT PLANNING / IMPLEMENTATION STILL BLOCKED.  
> **Date:** 2026-08-25.  
> **Scope:** repository health, continuity and review posture only; no new Product capability or implementation authorization.

## 1. Decision

MetalDocs repository governance is no longer a blocker to Product work.

```text
current main
+ local Engineering / Repository / Frontend methods
+ selective authority routing
+ compact roadmap / decision register
+ zero open PRs before this closeout increment
+ valid B11-F1 authority integrated
= READY TO CONTINUE T11 / FP1 PRODUCT PLANNING
```

Product/runtime implementation remains blocked by `docs/roadmap.md` until the existing T11/T12 and later implementation-readiness gates close.

## 2. Verified integrated baseline

Before this closeout increment:

```text
main                         c0a206f56596bd886011817da9c5f48c48dc58a9
PR #174 governance          MERGED
PR #175 B11-F1             MERGED
PR #173 old B11 workspace  CLOSED / SUPERSEDED / NOT MERGED
open PRs                    0
application operations     89
operation 90+              0
```

Current active routing contains no external methodology router/pin or `REPOSITORY-STANDARD` dependency. Historical branch refs may still exist on GitHub, but they are not current Product/status authority and do not block new work from `main`.

## 3. ClaudeCode / FABLE review doctrine retained

The historical ClaudeCode/FABLE workflow contributed useful review doctrine. The old `.claude/skills` transport/framework is **not** restored as a fourth methodology or repository dependency.

When an independent challenge is materially required, use a fresh challenger with this bounded posture:

```text
verify anchors before accepting claims
→ root cause before patch
→ ask Local Maximum vs Global Maximum
→ run a subtractive / YAGNI pass
→ classify findings against current authority
→ dispose prior findings explicitly
→ attack only new/material uncertainty
→ stop when material findings converge
```

A reviewer finding is Evidence, not requirement authority. Repeated same-altitude findings are a structural signal: stop patching and reopen the smallest owning decision once instead of creating endless rounds.

### Trigger independent review when

- the Engineering Method's independent-review floor is met;
- a material Product/architecture authority or trust boundary is created/moved;
- a major stage/global-coherence gate is closing;
- the final integrated frontend Product (P11/P12 boundary) is being challenged;
- implementation authorization is about to be granted.

### Do not use independent review as an inner-loop tax

Normal P7/P8 iteration, copy/layout correction, already-owned bounded correction and mechanical contract tracing do not need a fresh FABLE round every time. They need the applicable method, targeted proof, operator-only LOCK where required, and one strong final challenge when the material candidate is ready.

If actual ClaudeCode/FABLE transport is unavailable, do not claim independent FABLE convergence. The doctrine may still guide lead/self-review, but an independent-review gate remains independent when the Engineering Method requires one.

## 4. Closeout adversarial findings

A FABLE-style root-cause / Global-Maximum / simplify pass over the current repository produced:

### R1 — repository hygiene was incorrectly gating B11

```text
severity    MAJOR
root cause  administrative branch cleanup was promoted into Product-stage sequencing
fix         make ordinary branch/ref cleanup non-blocking; continue B11 from current main
```

Old branch refs may be deleted later when suitable tooling is available and their provenance law permits it. They must not stop Product planning merely because the names still exist.

### R2 — independent-review usage was no longer discoverable

```text
severity    MAJOR
root cause  useful ClaudeCode/FABLE doctrine survived only in Git history
fix         retain a compact review trigger/convergence rule in engineering-rules.md
```

This does not restore the old global methodology/router or permanent review artifacts.

### R3 — old B11 workspace is historical, not a current baseline

```text
severity    CLOSED BY STRUCTURE
root cause  the prior PR became a long-lived 80+ commit workspace
fix         PR #173 closed as superseded; B11-F1 extracted to main; B11 restarts cleanly
```

Known frontend findings from that workspace are inputs to the clean B11 rebaseline, not reasons to continue the old branch.

### Simplify verdict

Do **not** add:

```text
another methodology repository
another router/profile selector
per-round permanent FABLE documents
branch-cleanup gate before Product work
new Evidence refs for every intermediate P8 revision
```

## 5. B11 continuation contract

B11 may resume as a new acceptance increment from current `main`.

It must:

```text
preserve integrated B11-F1 op31 precision
+ use prior B11 work only as Evidence
+ address known failure classes in the first clean candidate
+ produce one coherent current P8/P9/P10 line
```

Known failure classes to carry forward:

1. paginated member/selector reads must use visible server-page traversal; no hidden all-page crawl;
2. `listUsers` grant selection preserves raw server page boundaries — DISABLED Users remain visible but unavailable, never pre-filtered before pagination;
3. add-member UX must not assume complete Group membership knowledge; idempotent op28 `PUT` reconciles first-add vs already-member truth;
4. repeated grant confirmation / same Idempotency-Key must produce zero second semantic mutation.

These are bounded frontend realization constraints. They do not create operation 90, `Group.area_id`, a custom Role editor or a browser effective-access engine.

## 6. Review convergence verdict

```text
BLOCKER  0
MAJOR    0 after R1/R2 closeout corrections
MINOR    historical ordinary branch refs remain as administrative residue

VERDICT  PROCEED WITH PRODUCT PLANNING
```

Repository administration should not preempt the next Product acceptance increment unless new Evidence proves it can change current authority, correctness or recoverability.

## 7. Reopen triggers

Reopen repository governance/readiness only if concrete Evidence shows one of:

- a fresh actor cannot recover current state from `AGENTS.md` + repository authority without chat archaeology;
- duplicate/stale authority again changes or obscures Product meaning;
- context inflation again requires broad recursive reading for ordinary tasks;
- PR/review mechanics repeatedly create long-lived workspaces or patch-on-patch loops;
- historical refs become the only remaining path to required unmerged provenance;
- repository mechanics materially block correct Product planning or implementation.

Preference for more ceremony, more reviewers, or a cleaner branch list is not by itself a Product-work stop condition.
