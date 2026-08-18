# MetalDocs — Post-T5 Integrated T1→T5 — Independent Fable Delta Review

> **Status:** INDEPENDENT DELTA REVIEW — EVIDENCE ONLY / **NOT TARGET AUTHORITY**
> **Date:** 2026-08-18
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Reviewed delta:** `bdef5fc3..094d67da` (ratified round-1 bounded amendments)
> **Original review:** `docs/superpowers/analysis/2026-08-18-t1-t5-integrated-independent-fable-review.md` @ `bdef5fc3`
> **Adjudication:** `docs/superpowers/analysis/2026-08-18-t1-t5-fable-author-adjudication-round1.md` — OPERATOR-RATIFIED
> **Delta request:** `docs/superpowers/analysis/2026-08-18-t1-t5-fable-delta-review-request.md`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** **CLOSED — this review changes no authority, code, schema, OpenAPI, frontend or provider configuration.**

---

# 1. Required verdict

```text
DELTA VERDICT = APPROVE

ORIGINAL FINDINGS:
M1 = CLOSED
M2 = CLOSED
M3 = CLOSED
L1 = CLOSED
L2 = CLOSED
L3 = CLOSED
L4 = CLOSED
L5 = CLOSED

NEW MATERIAL FINDINGS = 0
DISAGREEMENT SET = EMPTY
T6 READINESS = MAY OPEN
```

---

# 2. Why the delta closes the original defect classes

Every claim below was verified against the promoted authority text at `094d67da` (diffs `bdef5fc3..HEAD` read in full for the Product Contract, T1–T5 and the Decision Registry; router/handoff checked for gate consistency).

**M1 — CLOSED.** T5 §8 now requires the per-Document projection-write serialization to be acquired **before** the canonical-state read and held through rewrite/removal, with rebuild under the same law (ASY-11). This is the exact ordering that kills the original counterexample: a worker that blocks on the serialization performs its canonical read only after the earlier writer releases, so under READ COMMITTED it observes at least that writer's committed state — an older observation can no longer commit last. FIFO/broker rejection is correctly preserved; the lock/upsert primitive is correctly left to implementation. The proof-obligation list adds the concurrent-overlap fixture.

**M2 — CLOSED.** T4-M invalidates all restored ApplicationSessions before ordinary serving (structural closure of the bearer-token hole); T4-N gates ordinary authenticated serving on reconciliation of required known post-snapshot offboardings/access teardown, fails closed when the chosen proof cannot establish safe access state, and explicitly separates a non-serving operations trust surface. The seam law is respected: no generic per-grant security journal is frozen (it joins the anti-resurrect list), and T7 owns the smallest concrete recovery evidence/choreography via its amended REOPEN line. Registry AUTH-05/ORG-10/STO-11/STO-18/SEC-07 follow coherently.

**M3 — CLOSED (option b).** The baseline is now a canonical PostgreSQL query/view; `search_refresh`, the materialized projection, rebuild and the M1 serialization law are one coherent conditional seam activated only by a T6-named derived/expensive fact or measured need. The activation seam is properly guarded in both directions: the anti-legacy list forbids materialization "merely because Search exists", and T5-B forbids inventing a derived field to justify the projection. The L1 resolution reinforces the baseline's sufficiency: title as Revision-governed metadata makes every Product-Contract-named search fact a canonical column. T5-A/ASY-04 retain River's justification through an independent consumer (policy-required OfficialRendition — a Launch-Core product capability, so the mechanism is not dormant even in an all-SourceOnly deployment). Durable-intent, observability and proof-obligation sections are conditionalized consistently.

**L1 — CLOSED.** Product Contract REV001 §4 + invariant 20 + Retitle scenario, and T1's amended family/laws, bind human-readable title to the Revision; ordinary readers/search follow the current EFFECTIVE Revision's title; historical Revisions keep their governed titles; a DRAFT/SUBMITTED retitle cannot rewrite reader truth.

**L2 — CLOSED.** T5 §6 late-renderer no-op (no OfficialRendition, no Release, output reclaimable after claim release/expiry), mirrored in T2 §9/§12 failure laws and registry REL-09; finalization now proves "exact Submission remains the current eligible pre-Release candidate", removing the original ambiguity of "eligible".

**L3 — CLOSED.** T4-F defines the live bounded admission claim/binding protecting in-flight READY content; T4-K eligibility and the T5-J pre-delete recheck both prove claim absence (plus backup pin, now stated in the eligibility transaction as well); STO-14/STO-16/ASY-06 follow. Bounded expiry keeps this liveness state, not business retention.

**L4 — CLOSED.** Product Contract REV001 journey J.10 + T2 §8 (withdraw active human-governed obsolescence request; target remains EFFECTIVE; no fabricated verdict; retry = new request) + T3 authorization predicate (`document.obsolete` + initiator-or-`document.owner.manage`) + Audit census "obsolescence withdrawn" + Document-serialization list including obsolescence withdrawal + OBS-07/GOV-12. The NoHumanApproval no-window clarification is correct (synchronous completion).

**L5 — CLOSED.** The provider-disable intent line was removed from the T3 §10 offboarding transaction; T3 now names T5-L as baseline with a bounded future activation shape. No wording path back to a mandatory IdP-disable job remains.

**Notes N1–N3** were also landed: same-DB durable-intent restore coherence recorded as a guarded property (T4-L + STO-20, with the re-prove obligation on any future job-substrate move); ambiguous `SUPERSEDED` registry labels tightened (GOV-14, DOC-03, OBS-03); T6 REOPEN set now names source upload/T4 admission UX and the Search-materialization proof question.

No new material contradiction was found in the amended text or its registry reconciliation. The amendments are genuinely bounded: no semantic owner, lifecycle state, transaction law or authority boundary changed beyond the adjudicated findings.

---

# 3. Non-blocking observation (not a finding, no correction required)

Title is now Revision-governed and decision-relevant, so SUBMIT freezes it into the Submission snapshot (CNT-06 covers this). The one mechanism detail intentionally left open is *how* a DRAFT retitle is mutated and concurrency-protected — inside WorkingContent under the existing OCC generation, or as Revision metadata under Document serialization. Either satisfies T2's "no silent last-write-wins for governed DRAFT content" law if applied; T6/implementation should place DRAFT retitle explicitly under one of those two existing laws rather than leaving it outside both. This is journey/mechanism design squarely inside T6's existing REOPEN scope; it does not reopen T1/T2 and does not block the checkpoint.

---

# 4. Gate consequence

```text
Post-T5 checkpoint  → may close on adjudication of this delta verdict
T6                  → MAY OPEN after explicit checkpoint closure
implementation      → remains BLOCKED
```

This delta review is evidence, not authority. Checkpoint closure and T6 opening remain operator actions.
