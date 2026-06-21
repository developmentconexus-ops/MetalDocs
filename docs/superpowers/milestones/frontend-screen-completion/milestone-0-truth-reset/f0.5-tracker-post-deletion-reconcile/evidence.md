# Feature F0.5 — Evidence

> **Milestone:** 0  ·  **Feature:** `f0.5-tracker-post-deletion-reconcile`  ·  **Closed:** 2026-06-21
> **Contract:** `spec.md` (consumer = the tracker resume doc + validator RM3).
> **Origin:** opened per the M0 milestone-validator FAIL (HS-4) — split-brain between the F0.1 tracker and the F0.3 deletions.

## What was implemented

Reconciled the two stale rows the validator's RM3 flagged in
`wiki/implementation/screen-redesign-tracker.md`:

- **Operations row:** `stub → cut`; Route → "— (route removed M0/F0.3)"; Component →
  `operations/pages/OperationsPage.tsx` *(deleted)*; Milestone → `cut (M0/F0.3 — deleted)`; Notes
  rewritten to past tense (was an empty `OperationsCenter` shell with the dup root index; F0.2 removed
  the index, F0.3 deleted the page + route; IAM owns metrics/audit/sessions).
- **Audit row:** same treatment — `cut`, route removed, file marked *(deleted)*, Notes past tense.
- **Evidence-base note:** corrected "`ls` of every cited page file (all 20 present)" → the 18 routed
  pages present after F0.3; the two `cut` rows cite their now-deleted files for lineage only, marked
  *(deleted)*.

Doc-only; no code or router touched (F0.3 already executed the deletion).

## Verification

| Check | Command | Result | Real vs fixture |
|-------|---------|--------|-----------------|
| Operations + Audit rows read `cut` | `grep -nE "^\| (Operations\|Audit) " …tracker.md` | both rows show **`cut`** + "route removed" + *(deleted)* | real |
| RM3 — every live-presented row's file exists | `ls` of all 17 non-cut routed page files | **17/17 ok, 0 MISSING** | real |
| Deleted files absent (and only cited as `cut`) | `ls` of `OperationsPage`/`AuditPage` | **both gone(correct)**; cited only on `cut` rows, never as live `stub` | real |
| Evidence-base count corrected | read of the evidence-base note | claims 18 routed pages present, two deleted cited *(deleted)* | real |

> This directly closes the validator's RM3 MISSING finding (2/9 sampled files missing while presented as
> live → now 0 live rows cite a missing file). Deterministic `grep`/`ls`; doc-only.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| No live row cites a deleted Operations/Audit file | yes | both cited only on `cut` rows, marked *(deleted)* |
| Operations + Audit rows read `cut` | yes | grep shows `cut` on both |
| RM3 clean — every live row's file exists | yes | 17/17 ok, 0 MISSING |
| Evidence-base count corrected | yes | note now says 18 present + *(deleted)* lineage |

All 4 criteria **met**.

## Review disposition

- Spec-compliance review: self-review against `spec.md` — PASS. Only the two flagged rows + the
  evidence-base note changed; no other row, header, lineage, or the DoD doc touched; no code/router
  change.
- Code-quality review: n/a (docs). Re-judged by the re-dispatched M0 `milestone-validator`.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| (none) | F0.5 is the complete fix for the named FAIL | — |

## Root-cause note (for the milestone retro)

The split-brain arose because F0.1 (tracker rewrite) ran **before** F0.3 (deletion) within the same
milestone and nothing reconciled the doc afterward. The durable lesson: when a milestone both *records
truth* and *changes truth*, the truth-record feature must run **after** the truth-changing feature, or a
reconcile step must close the loop. Recorded so M1+ sequence doc-vs-code features accordingly.
