# Feature F0.5 — Spec

> **Milestone:** 0 — Truth reset & structural cleanup  ·  **Folder:** `f0.5-tracker-post-deletion-reconcile`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-21 / leandrotca — *fix feature opened per the M0 milestone-validator FAIL (HS-4); contract is "the tracker must match post-F0.3 reality." No open decision.*
> **Origin:** M0 milestone-validator verdict FAIL (C1/C2-RM3/C6 split-brain) in `../qa/milestone-qa.md`.

> This is the feature's **contract**, written and approved **before any code**. The milestone-validator
> judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | none needed — why | The defect is mechanical and the correct end-state is unambiguous: F0.1 rewrote the tracker *before* F0.3 deleted Operations/Audit, so rows 22–23 still present them as live routed `stub` screens citing deleted files. Reality (post-F0.3): both deleted, both routes unmounted. The fix is to reconcile those two rows to that verified truth. No decision to interview. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** the operator + every later milestone reads `screen-redesign-tracker.md` as the durable
  resume doc; the M0 milestone-validator's RM3 (tracker-row-vs-implemented-page sample) binds on it.
- **Contract:** **no tracker row may cite a source file that does not exist** or present a deleted/unmounted
  screen as live. After F0.5, the Operations and Audit rows reflect their F0.3 deletion (status `cut`,
  route removed, component marked deleted), and the evidence-base note matches the post-deletion page count.
  RM3 must return **0 MISSING** for any row presented as live.
- **Source of truth for the contract:** F0.3 evidence (`OperationsPage`/`AuditPage`/`OperationsCenter`
  deleted, routes unwired — verified by grep=0 this session) + mission D7.

## What this feature implements

Reconcile the two stale rows in `wiki/implementation/screen-redesign-tracker.md` to post-F0.3 reality:

1. **Operations row (22):** Status `stub` → `cut`; Milestone `M0 / F0.3 (delete)` → `cut (M0/F0.3 — deleted)`;
   Route → mark the route removed; Component → mark the file deleted; Notes → past tense ("Was an empty
   `OperationsCenter` shell; **deleted** in M0/F0.3 per D7. IAM Admin Center owns metrics/audit/sessions.").
2. **Audit row (23):** same treatment — Status `cut`, deleted, route removed, file deleted, Notes past tense.
3. **Evidence-base note:** correct "`ls` of every cited page file (all 20 present)" to the post-deletion
   count (the two deleted pages no longer cited as live; remaining routed pages present).

## Non-goals (mandatory)

- **Not** touching any other tracker row, the header, the governing lineage, or the DoD doc.
- **Not** re-deleting code or changing the router (F0.3 already did that; F0.5 is doc-only).
- **Not** re-classifying any other screen's status.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| No live row cites a deleted Operations/Audit file | `grep -nE "operations/pages/OperationsPage\|audit/pages/AuditPage" wiki/implementation/screen-redesign-tracker.md` → only on rows marked `cut`/deleted, never as a live `stub` | real |
| Operations + Audit rows read `cut` | `grep -nE "^\| Operations \|^\| Audit " …` → both rows show `cut` status | real |
| RM3 clean — every live-presented row's file exists | re-sample: for each row NOT marked `cut`/`not-started`/`out-of-scope`, the cited page file exists (`ls`) → **0 MISSING** | real |
| Evidence-base count corrected | the "(all N present)" note no longer claims the deleted pages are present | real |

> Doc-only; no automated unit test. Deterministic `grep`/`ls`. This directly closes the validator's RM3
> MISSING finding.

## ADR needed?

- [x] No durable decision — skip. Mechanical truth-sync of two rows to a deletion already decided by D7
  and executed by F0.3.
