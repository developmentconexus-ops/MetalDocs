# Grade-A Architecture Remediation — Design Spec

> **Status:** Proposed — awaiting operator review (brainstorming gate)
> **Date:** 2026-06-14
> **Branch of record:** qa/iam-area-membership
> **Author:** backend agent (Opus 4.8) + operator
> **Governs:** Milestones M0–M5 below. Each Milestone gets its own implementation plan (writing-plans), executed in a fresh session.
> **Supersedes for this program:** none. **Builds on:** `wiki/backend/_artifacts/architecture-audit-2026-06-13.md` (the independent audit), Wave H/V1 close-outs.

---

## 1. Problem

The independent 2026-06-13 audit graded the backend **B / "good senior Go"** — debt concentrated in three *structure* dimensions: **module-boundaries/DDD (C)**, **contract/API layer (C)**, **composition/observability (C)**. Wave H closed most structural debt; Wave V1 closed the v1 release blockers (B1–B6, H-A..H-G).

Two of those rows — **H-D** (contract tri-source drift) and **H-G** (hardcoded domain state / cross-module reach) — were reclassified `TRIGGER`, i.e. closed as *instances* but left open as **classes**. A follow-up architecture-wide sweep (46 agents, adversarial-verified) confirmed: **23 unique verified-real defects**, of which **4 literally block Grade A** and **19 are the quality tail** that hardens the three formerly-C dimensions.

This program closes the **classes**, not just the instances, and proves Grade A by an **independent re-audit gate** — never by self-assertion.

## 2. Goals / Non-Goals

**Goals**
- Take the three formerly-C dimensions to **Grade A−/A**, independently re-audited.
- Fully close the **H-D class** (handler/contract field drift; tri-source route drift) and the **H-G class** (hardcoded domain state + cross-module reach-without-a-port).
- Every fix carries **evidence** (gates + runtime proof + review/QA disposition). Symptom-patching is a hard-stop.
- A **human-in-the-loop hard-stop** at every Milestone boundary and whenever scope must be replanned.

**Non-Goals**
- No rewrite. Bounded, fixable list only.
- No new product features. No speculative abstractions.
- No data migrations and **no snapshot/denormalization semantics** for the H-G ports (reads stay live — Approach 2 was explicitly rejected absent a separate audit/legal "freeze actor name" product decision).
- No merges by the agent — the operator merges.

## 3. Locked decisions (operator-approved)

| # | Decision | Value |
|---|----------|-------|
| D1 | Scope | **Full A + close classes** — reach Grade A AND fully close H-D and H-G, including the systemic port wave |
| D2 | Execution | **One spec, per-Milestone plans + inter-Milestone operator gates**; each Milestone executed in a fresh session; no merge without approval |
| D3 | Proof of A | **Independent re-audit gate** — final multi-agent re-audit is the authoritative Grade-A sign-off |
| D4 | H-G fix shape | **Approach 3 (Hybrid ports)** — fix the contained instance first (M1), then generalize to shared ports `UserDisplayNameReader` + `TemplateVersionStateReader` (M4). Reads stay live; no migrations |
| D5 | Sequencing | **Docs first** (M0) — de-stale roadmap/backlog/ADRs before the architecture milestones; each milestone updates its own docs; light final curation |
| D6 | Structure | **Milestone → Feature** with per-Feature close-out and per-Milestone validation (QA + code review + focused audit slice proving root-cause-fixed + Grade-A for the touched dimension) |
| D7 | Control | **Human-in-the-loop hard-stop** between every Milestone and on any replan/off-track condition |

## 4. Program architecture

```
Program: Grade-A Architecture Remediation
└── Milestone (M0..M5)
    ├── Feature (Fx.y)  ── each runs the Feature Close-Out Loop
    │     implement → static+targeted verify → code review → product QA
    │     → classify findings by root-cause family → fix by family → rerun → evidence row
    └── Milestone Validation Gate
          all Feature close-outs green
          + product QA (canonical checklist for the milestone's workflow class)
          + independent code review of the milestone diff
          + FOCUSED AUDIT SLICE: re-grade the touched dimension(s) → must trend A−/A
            AND confirm the milestone's defect CLASS is fixed at ROOT CAUSE (not symptom)
          + evidence bundle assembled
          → HUMAN-IN-THE-LOOP HARD-STOP (operator review; no merge without approval)
```

### 4.1 Feature Close-Out Loop (canonical, every Feature)
Per `CLAUDE.md §4` default close-out loop:
1. Implement inside the bounded Feature.
2. Static + targeted verification for the touched slice.
3. Independent code review (spec-compliance, then code-quality).
4. Product QA using the canonical checklist for the workflow class.
5. Classify findings by **root-cause family** (F1–F9 taxonomy), not scattered symptoms.
6. Fix by family; rerun targeted review + QA + regression.
7. Broader regression when the change crossed a boundary.
8. Record the **evidence row**; declare bounded defers with written triggers.

### 4.2 Evidence row (what every Feature/Milestone records)
- Commands: `go build` · `go vet` · `go test -p 2 ./...` · `api-lint -strict` · `cilint` — each **0**.
- FE (only when codegen touched): `gen:api` (intentional byte-diff noted) · `tsc` **0**.
- **Runtime proof** for every observable change (we run it; never delegated to the operator): exact request + response status/body excerpt.
- Review dispositions (spec + quality), QA outcome, classified findings, fixes by family.
- Bounded defers, each with a written **trigger**.
- `done`/`green`/`looks good` alone is **not** closure.

### 4.3 Focused audit slice (per Milestone, non-terminal)
A scoped re-run of the relevant dimension(s) of the 2026-06-13 10-dimension method — enough agents to re-grade only the dimensions the Milestone touched, with an adversarial skeptic per finding. It answers two questions:
- **Root cause fixed?** The defect *class* this Milestone targeted is gone, not symptom-patched (e.g. M1: zero bare-405 sites remain anywhere in the swept packages; M4: zero reach-without-a-port remains).
- **Grade-A for the touched dimension?** The dimension trends A−/A.

A slice failure on either question is **HS-4** (replan the Feature). The slice is *not* the authoritative Grade-A gate — that is **M5**.

## 5. Defect inventory (source: 2026-06-13 audit + Grade-A sweep)

### 5.1 The 4 that block Grade A
| # | Site | Class | Locality | Milestone |
|---|------|-------|----------|-----------|
| 1 | `auth/delivery/http/handler.go:66,101,121,134` — login/logout/me/change-password bare 405 | error-contract (H-B) | mechanical | M1 / F1.1 |
| 2 | `iam/delivery/http/admin_handler.go:149` — admin/overview path-only → bare 405 | error-contract (H-B) | mechanical | M1 / F1.1 |
| 3 | `iam/presence/handler.go:66` — `GET /iam/presence/stream` live + FE-consumed but **absent from openapi.yaml, no stub** | contract tri-source (H-D, stronger form) | needs-contract-change | M1 / F1.2 |
| 4 | `documents/approval/application/decision_service.go:266-272` — raw `SELECT … FROM metaldocs.iam_users` inside the signoff tx | cross-module (H-G shape) | bounded cross-module | M1 / F1.3 (contained) → M4 / F4.1 (generalized) |

### 5.2 The 19 quality tail
| Group | Sites | Class | Milestone |
|---|---|---|---|
| error-contract tail | `iam/sessions_handler.go:67,143`, `featureflags/handler.go:33`, `observability/http.go:149` (bare-405) | H-B | M1 / F1.1 (same sweep) |
| H-D tail (handler emits undeclared field) | `planTier` on `/iam/usage` (breaks `UsageGauges.tsx`), `status` on `OnlinePresenceItem`, `/templates` envelope-vs-bare-array | H-D | M2 |
| H-G tail (reach-without-a-port) | `controlleddocuments/repository.go:705-710` reads templates-owned tables; `iam_users` direct-read systemic 5+ sites (`documents/repository.go:134`, `approval/get_instance_handler.go:127`, + security + presence) | H-G | M4 |
| H-G hardcode | `apps/api/internal/wiring/documents_adapters.go:113` hardcoded `status := "published"` | H-G (hardcoded-domain-state) | M4 / F4.2 |
| dead/test-only surface | 7 deletes (post-Wave-H orphans) | legacy/dead-code | M3 / F3.1 |
| tx-hazard | CD `service.go:308,331` off-tx reads inside lock-holding tx → hoist-to-preflight | persistence (H-PRE-1) | M3 / F3.2 |
| misc | auth dead camelCase `MarshalJSON`, CD `routes.go:123` raw `map[string]any`, documents pagination TOCTOU (`COUNT(*) OVER()`), `DeleteObject` silent error | code-quality / contract | M3 |

## 6. Milestones

### M0 — Docs Progression De-Staling
**Objective:** one unambiguous progression surface so the architecture milestones run against clean docs and stale material stops polluting agent context.

| Feature | Work | Done when |
|---|---|---|
| F0.1 ADR audit + ledger | Run `decisions/README.md` status-gate; verify each ADR's decision still matches code; flag drift | Every `wiki/decisions/` ADR has `> **Status:**` + matches code; drift marked Historical/Superseded/amended; `decisions/index.md` ledger refreshed |
| F0.2 Stale-ref repair | Fix the 12 refs to deleted `docs/` trees (`documentation-governance.md:73-96`, `decisions/index.md:34`, `quality/*`, `architecture/data-model.md`, `modules/frontend/iam.md`) | `grep` proves 0 wiki links to deleted `docs/` paths |
| F0.3 Roadmap consolidation | Mark `wiki/backlog/roadmap.md` (May) + `wiki/backend/roadmap.md` (June) **historical**; create one **forward roadmap** carrying this program + post-v1 progression | 1 forward roadmap; 2 old roadmaps clearly historical |
| F0.4 Backlog hygiene | Close/archive done items in `wiki/backlog/*`; keep active deferred work | Backlog = active deferred work only |
| F0.5 Archive convention | Establish `wiki/_archive/`; move superseded-historical docs; update domain `index.md`s + governance migration map to post-deletion reality | Archive tree + accurate indexes |

**Milestone gate:** doc-QA (ADR status-gate script + 0-stale-ref grep + broken-link sweep) + operator review. Root-cause criterion = ambiguity/stale eliminated. No code-audit slice (docs-only). **HS-1** at close.

### M1 — Reach-A Blockers (Wave A)
**Objective:** close all 4 Grade-A blockers + the error-contract tail.

| Feature | Closes | Fix shape |
|---|---|---|
| F1.1 bare-405 sweep | blockers #1,#2 + error-contract tail (4 sites) | Every bare `w.WriteHeader(405)` → `problem+json` 405 via canonical helper. Kills the *class* across auth/iam/featureflags/observability delivery |
| F1.2 presence/stream spec | blocker #3 | Declare `GET /iam/presence/stream` in `openapi.yaml` + regen stub + FE type; 101 upgrade stays live (statusWriter `Unwrap()` already merged) |
| F1.3 approval display-name reach (contained) | blocker #4 | Move the raw `iam_users` SELECT out of the signoff tx into a contained `ApprovalRepository` method (Approach-3 step 1) |

**Milestone gate:** Feature close-outs green + backend-api-qa-checklist + code review + **focused audit slice** on the **contract** + **error-contract** facets: zero bare-405 anywhere in swept packages (root cause), `/iam/presence/stream` now tri-source-consistent. **HS-1** at close.

### M2 — Contract Tail (H-D class) (Wave B)
**Objective:** eliminate handler-emits-undeclared-field drift; batched behind **one** FE regen.

| Feature | Work |
|---|---|
| F2.1 `planTier` | Declare `planTier` on `/iam/usage` response schema; unblock `UsageGauges.tsx` |
| F2.2 `OnlinePresenceItem.status` | Declare `status` on `OnlinePresenceItem` |
| F2.3 `/templates` envelope | Declare the `/templates` list envelope (vs bare array) |

Single FE `gen:api` after all three; `tsc` 0. **Milestone gate:** backend-api-qa-checklist + FE screen-impact check (UsageGauges) + code review + **focused audit slice** on the **contract** dimension: H-D class = 0 undeclared-field drift across delivery handlers (root cause). **HS-1** at close.

### M3 — Mechanical Quality (Wave C)
**Objective:** harden code-quality + persistence dimensions; all bounded, single-purpose.

| Feature | Work |
|---|---|
| F3.1 dead-surface deletes | 7 post-Wave-H orphans; zero-caller proof per symbol |
| F3.2 tx-hazard hoist ×2 | CD `service.go:308,331` off-tx reads hoisted to pre-flight — **respect H-PRE-1** (hoist off-tx, never a Tx-variant inside the lock); runtime-prove CD-create stays fast, no deadlock |
| F3.3 CD raw-map → generated type | `routes.go:123` raw `map[string]any` → generated response type |
| F3.4 documents pagination | TOCTOU fix via `COUNT(*) OVER()` |
| F3.5 `DeleteObject` | silent error → WARN log |
| F3.6 auth dead method | delete dead camelCase `MarshalJSON` |

**Milestone gate:** backend-api-qa-checklist + workflow-async-qa-checklist (tx-hoist touches CD-create) + code review + **focused audit slice** on **code-quality** + **persistence**: each delete proven caller-free, tx-hazard gone with H-PRE-1 intact (root cause). **HS-1** at close.

### M4 — Systemic Ports (H-G class) (Wave D)
**Objective:** close the H-G class by generalizing to shared ports. Last, so it cannot regress the grade.

| Feature | Work | Root cause closed |
|---|---|---|
| F4.1 `UserDisplayNameReader` | iam/domain-owned port; migrate 5+ direct `iam_users` read sites (incl. generalizing F1.3's contained fix) | cross-module reach-without-a-port |
| F4.2 `TemplateVersionStateReader` | templates/domain-owned port; replace `controlleddocuments/repository.go:705-710` reach **and** the `documents_adapters.go:113` hardcoded `"published"`; status read stays **off** the lock-holding tx (H-PRE-1) | reach-without-a-port + hardcoded-domain-state |
| F4.3 ADRs | one ADR per port, into the now-clean `wiki/decisions/` ledger | governance |

One PR per port. Reads stay live; no migrations; no snapshot semantics. **Milestone gate:** backend-api-qa-checklist + code review + **focused audit slice** on **module-boundaries**: H-G class = 0 reach-without-port + 0 hardcoded-domain-state (root cause); ADRs present. **HS-1** at close.

### M5 — Independent Re-Audit (authoritative Grade-A gate)
**Objective:** prove Grade A by an independent fresh read — the authoritative sign-off.

| Feature | Work |
|---|---|
| F5.1 re-audit | Re-run the **full 10-dimension multi-agent audit method** (Workflow tool) on the post-M4 branch; adversarial skeptic per finding |
| F5.2 micro-wave loop | Any dimension < A− → its findings become a bounded remediation micro-wave → re-audit again (loop until pass); operator decides continue/replan at each loop |
| F5.3 operator sign-off | Operator owns the final Grade-A declaration |

**Pass bar:**
- 3 formerly-C dimensions (module-boundaries, contract, composition) all **≥ A−**, none below.
- **0** new Critical/Major.
- H-D class = 0 tri-source drift; H-G class = 0 reach-without-port + 0 hardcoded-domain-state.

This is the only milestone that uses a full Workflow fan-out; M0–M4 use targeted per-slice verification + focused slices.

## 7. Human-in-the-loop hard-stop catalog

| ID | Trigger | Action |
|---|---|---|
| HS-1 | Every Milestone boundary | Operator review gate; **no merge without approval** |
| HS-2 | A fix implies shared-API redesign, cross-module auth/authz model change, storage/provider redesign, or workflow-semantic redesign outside the assigned boundary (`CLAUDE.md` hard-stop) | **Stop**; report the architecture boundary + minimum prerequisite plan; do not symptom-patch |
| HS-3 | A prerequisite boundary fails (build/runnable/auth-session/route/contract truth) | Switch to `runtime-contract-prereq`; repair; rerun the failed checkpoint; return to the Feature |
| HS-4 | A Milestone focused-audit slice finds symptom-patch (root cause NOT fixed) or a dimension not trending A− | **Stop**; replan the offending Feature; re-run its close-out |
| HS-5 | M5 re-audit finds any dimension < A− | Bounded micro-wave + re-audit loop; **operator decides** continue vs replan |
| HS-6 | Scope drift / off-plan discovery mid-Milestone ("goes off") | **Stop**; surface the deviation; replan before continuing |

## 8. Constraints respected
- **H-PRE-1** advisory-lock deadlock: never call an authz-recording read on a fresh connection inside the audit-lock atomic tx; fix tx-hazards by hoisting reads off-tx (applies F3.2, F4.2).
- **CLAUDE.md hard-stops** (HS-2): none of the planned Features trip a redesign boundary — all are additive/bounded/contained. If one is discovered to, HS-2 fires.
- **Contract-first**: route/codegen/spec changes (F1.2, M2, F3.3) follow the build-route-truth-table → compare runtime/spec/codegen/wiki → regen flow.
- **No-merge**: operator merges; agent never does.
- **D4/Approach-3**: contained fix (F1.3) precedes generalization (F4.1); no snapshot semantics.

## 9. End-state ("all we planned was executed")
Closure of the program requires a **final reconciliation**:
- Every Feature in M0–M4 has an evidence row.
- Zero unplanned scope merged; every bounded defer carries a written trigger.
- M5 re-audit passes the §6 pass bar; operator signs off Grade A.
- The forward roadmap (F0.3) reflects the executed program and any deferred triggers.

## 10. Execution model
- **One spec** (this) governs all Milestones. Each Milestone → its own plan via **writing-plans**, executed in a **fresh session**, subagent-driven (implementer + spec-review + quality-review per Feature).
- **Operator gate between every Milestone** (HS-1). **No merge** by the agent.
- **Token discipline:** Workflow fan-out only where it pays (M5 re-audit, focused slices); everything else direct tools.

## 11. Open risks / non-goals restated
- Focused audit slices (M0–M4) are *indicative*, not authoritative; only M5 declares Grade A.
- If M5 loops more than twice on the same dimension, treat as a design-boundary signal (HS-2/HS-5) rather than continued patching.
- No FE feature work beyond the codegen-type unblocks (UsageGauges) that the contract fixes require.
