---
id: repository-reset-ai-dialog
kind: work
owner: architecture
summary: Temporary independent-review, Lead-adjudication, and operator-decision record for the clean-slate repository reset.
---

# AI dialogue — repository reset

> **TEMPORARY / NON-AUTHORITATIVE / DELETE BEFORE MERGE**

## Review context

```text
Repository: developmentconexus-ops/MetalDocs
Branch: reset/clean-slate-repository
PR: #134
Fable reviewed HEAD: feb6c27231667eced554867a499b6cb578ae8fc7
Fable review commit: 622700fb943b89686354462f7693ca7e3ca51b72
Current gate: clean-slate source-tree reset
Implementation: BLOCKED
T8-E: paused at docs/reference/t8e-checkpoint.md
```

The complete independent Fable review is preserved in Git at commit `622700fb943b89686354462f7693ca7e3ca51b72`.

## Fable review summary

```text
PRIMARY VERDICT:
APPROVE CLEAN-SLATE REPOSITORY RESET WITH MATERIAL FIXES

STRUCTURE / GLOBAL MAXIMUM    CONFIRMED
BLOCKER                       3
MAJOR                         3
LOW                           6
PRODUCT/R10 REOPEN            NO
LEGACY IMPLEMENTATION RETURN  NO
```

The reviewer confirmed delete-in-place on a branch + squash merge as the Global Maximum. Findings were bounded gaps in provenance reachability, durable routing, decision carry-forward, stage definitions, and enforcement.

## Lead adjudication

| Finding | Disposition | Correction |
|---|---|---|
| B1 — unmerged PR #131/#132 provenance could become unreachable | **ACCEPT WITH BOUNDED CORRECTION** | Source branches and exact SHAs are now protected provenance refs in `docs/status.md` and `docs/decisions/repository-reset.md`; they MUST NOT be deleted until equivalent immutable archival tags/refs exist. Merge no longer schedules branch deletion. |
| B2 — durable navigation depended on `docs/work/current/` | **ACCEPT** | Durable routing now points to `docs/reference/t8e-checkpoint.md`; README/index/decision registry no longer depend on temporary work. |
| B3 — 76 vs 78 operation census conflict | **ACCEPT WITH STRONGER OWNER PAGE** | Added `docs/decisions/api-operation-census.md`, a semantic authority that owns only the approved +2 read-symmetry precision and establishes the current 78-operation census. T8-E consumes it. |
| M1 — remaining T8-F→T12 program missing | **ACCEPT** | Added `docs/decisions/stage-program.md`; `docs/architecture/transition.md` now points to it and the final implementation gate is defined there. |
| M2 — old decision registry forward obligations lost | **ACCEPT WITH SUBTRACTION** | Added `docs/decisions/forward-obligations.md` with exactly 21 PRESERVE + 4 REOPEN + 27 DEFERRED = 52 rows. CURRENT rows remain in owning authorities; SUPERSEDED rows remain deleted. Consumption law moved to `engineering-rules.md`. |
| M3 — CI denylist could not reject new implementation names | **ACCEPT** | CI is now an allowed-tree guard over all tracked paths, self-proves that `src/main.go` is rejected, and retains the required `required` status context. |
| L1 — T8-E status/routing contradictions | **ACCEPT** | README, status, decisions and checkpoint uniformly say T8-E is paused during reset and resumes only after ratification cleanup. |
| L2 — unused/broken `mkdocs.yml` | **ACCEPT BY DELETION** | `mkdocs.yml` removed; `docs/index.md` is the sole durable human/agent navigation surface until a real publication consumer exists. |
| L3 — external repository rule binding undocumented | **ACCEPT** | `engineering-rules.md` records ruleset `20560142`, required context name `required`, and resolved-review-thread merge precondition. |
| L4 — `kind: work` checkpoint outside temporary work | **ACCEPT WITH STRUCTURAL CORRECTION** | Added closed `kind: checkpoint`; T8-E checkpoint is durable non-authoritative state and may survive merge without pretending to be active work. |
| L5 — stale `wiki/...` routing prose survives in copied authorities | **ACCEPT WITH GLOBAL CORRECTION** | Documentation governance now makes every embedded `wiki/...` reference non-navigational provenance and names the semantic successors; transition's program pointer was directly repointed. This removes the defect class without risky cosmetic rewrites of copied authorities. |
| L6 — secret scanning retired without reopen trigger | **ACCEPT** | Repository-reset authority now requires an appropriate secret-scanning control before the first future implementation/code/schema/runtime commit is authorized. |

### Lead result

```text
BLOCKERS CLOSED                    3 / 3
MAJORS CLOSED                      3 / 3
LOWS CLOSED                        6 / 6
SURVIVING MATERIAL CONTRADICTION   0
CLEAN-SLATE GLOBAL MAXIMUM         CONFIRMED
UPSTREAM PRODUCT/R10 REOPEN        NO
SECOND FABLE ROUND                 NOT REQUIRED
```

No correction restored application code, old DB/API/frontend/deploy, old CI registry, roadmap, Harness, QA system, or any superseded module.

## Bounded round 2

```text
NOT REQUIRED
```

Reason: Fable confirmed the selected structure as the Global Maximum. Every correction is either the reviewer's smallest correction or a strictly stronger bounded correction inside the same structure. No new product semantics, trust boundary, implementation topology, or stage ordering was invented by adjudication.

## Operator decision

```text
APPROVED
Operator ratification date: 2026-08-20
```

The operator explicitly ratifies the corrected clean-slate repository reset and authorizes finalization:

```text
delete docs/work/current/**
→ change docs/status.md to T8-E ACTIVE
→ keep PR #131/#132 source branches protected until immutable archival refs exist
→ required CI green on final tree
→ resolve any remaining review conversations
→ squash merge PR #134
→ resume T8-E from docs/reference/t8e-checkpoint.md in a fresh small PR
```

This ratification does **not** authorize product implementation, schema/runtime/deploy work, T8-F, or restoration of superseded legacy code.