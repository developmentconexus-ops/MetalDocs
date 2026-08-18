# MetalDocs — Post-T5 Integrated T1→T5 Architecture — Independent Cold Adversarial Review

> **Status:** INDEPENDENT REVIEW — EVIDENCE ONLY / **NOT TARGET AUTHORITY**
> **Date:** 2026-08-18
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Reviewed checkpoint:** T1→T5 + Decision Registry, all OPERATOR-RATIFIED, @ `616f5ffd`
> **Review request:** `docs/superpowers/analysis/2026-08-18-t1-t5-integrated-fable-review-request.md`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** **CLOSED — this review changes no authority, code, schema, OpenAPI, frontend or provider configuration.**

---

# 0. Bootstrap and evidence base

True cold start from the repository at `616f5ffd`. Read in authority order: `AGENTS.md` → Method mirror → `wiki/references/current-agent-handoff.md` → `wiki/architecture/launch-v1-product-contract.md` → `wiki/architecture/whole-product-alignment-review.md` → `wiki/architecture/launch-v1-ownership-topology.md` → `r10-t1` → `r10-t2` → `r10-t3` → `r10-t4` → `r10-t5` → `rebaseline-decision-registry.md` → `r10-technical-architecture.md` → the review packet. Old R3–R9.5/R10 artifacts and current implementation were not consulted as authority; no claim below depends on current code shape.

Method posture applied: every material finding carries evidence → root cause → target invariant → local-vs-global → required correction → reopen scope → proof-to-close. Structural Inversion was run against each ratified pillar: the T1→T5 conclusions derive from the Product Contract and ratified topology, not from current implementation shape — inverting the current runtime (multi-tenant RLS substrate, 5×43 catalog, outbox/worker topology, nested approval module) changes no ratified conclusion. The rebaseline chain is genuinely code-independent.

All ten mandatory cross-stage attacks (packet §4 A–J), the authority-uniqueness sweep (§5), the registry audit (§6), the essential-vs-accidental attack (§7), the future-seam attack (§8) and the T6 readiness test (§9) were executed. Attack walkthroughs that produced no finding are summarized in §4 below rather than reproduced in full.

---

# 1. VERDICT

```text
APPROVE T1→T5 WITH MATERIAL FIXES
T6 remains blocked until the fixes below are adjudicated

BLOCKER = 0
MAJOR   = 3   (M1 Search latest-state convergence lacks the serialization law that makes it true;
               M2 restore resurrects revoked sessions/access teardown — readiness barrier covers PII only;
               M3 always-required search_refresh/projection machinery ratified before any consumer
                  requiring a materialized projection is named)
LOW     = 5   (L1–L5)
NOTE    = 3   (N1–N3)

MINIMAL REOPEN SET = NONE
  No ratified T-stage decision is invalidated. All three MAJORs are bounded amendments
  inside existing sections (T5-G; T4-M/T4-N; T5-B/T5-F + ASY-01). Everything else remains frozen.

ANTI-OVERENGINEERING VERDICT = PASS WITH ONE EXCEPTION
  Every retained mechanism except the Search projection (M3) has a named Launch consumer;
  T4 §18 demonstrates the discipline explicitly. No dormant module/table/permission/job
  beyond M3's contingency was found. No prohibited anti-legacy item leaks back in.

FUTURE-SEAM VERDICT = PASS
  All eleven named horizons attach by stable reference (checked individually, §6).
  No current decision rewrites immutable meaning or makes a named future materially
  harder than necessary.

T6 READINESS VERDICT = READY AFTER M1–M3 ADJUDICATION
  Semantic identity/state, transaction boundaries, AuthZ check sites, Audit obligations,
  exact-content ownership, viewer/rendition distinction and the Search authority boundary
  are coherent. None of the MAJORs forces API/journey redesign; adjudicating them first
  prevents T6 from encoding an unprovable convergence claim (M1) and an unowned
  machinery commitment (M3).
```

---

# 2. MAJOR findings

## M1 — MAJOR — T5-G's convergence claim is not true as stated: concurrent refresh executions can permanently overwrite the projection with stale state; the per-Document projection-write serialization law that makes the claim true is absent

**Claim.** `r10-t5` §8 ratifies: worker reloads latest canonical state; "duplicate jobs → harmless; out-of-order jobs → converge to latest state; Per-Document FIFO/ordering infrastructure is not required." This is true for job *ordering* but false for concurrent job *execution overlap*.

**Evidence / counterexample (integrated attack D and E).** Duplicate jobs are explicitly normal. Take REV001 replacement Release at time t enqueuing job J2, while an earlier J1 (from any prior event) is mid-flight:

```text
J1 reads canonical state at t-ε   → sees REV000 EFFECTIVE
Release commits at t              → REV000 SUPERSEDED, REV001 EFFECTIVE; J2 visible
J2 runs at t+1                    → reads new state, writes correct projection
J1 writes projection at t+2       → overwrites with REV000-as-current
```

The projection now presents superseded (or, in the obsolescence variant, obsolete) truth as current **until the next unrelated refresh or full rebuild** — that is divergence, not convergence. T5-H's serve-time canonical/AuthZ recheck (§9) correctly prevents this from ever becoming served *authority*, so the leak is bounded to discovery-level staleness — but the ratified property "converge to latest state" is itself unprovable for a naive implementation that satisfies every stated law, and T5 §19 lists exactly that property as a proof obligation.

**Root cause.** T2 deliberately states its serialization laws (Document as lifecycle serialization root, OCC for DRAFT); T5-G asserts a convergence property without stating the analogous concurrency law for projection writes. The property was treated as a consequence of latest-state reads, but latest-state reads only converge when read-and-write are serialized per key.

**Target invariant.** For one Document, the canonical-state read and the projection write of a refresh execution are atomic with respect to every other refresh execution and rebuild write for that Document.

**Local vs global maximum.** Adding per-Document FIFO/ordering infrastructure is the rejected local maximum (T5-G already correctly rejects it). The global maximum is one sentence of law exploiting the already-chosen same-database design: acquire the per-Document projection write serialization *before* reading canonical state, in one local transaction (e.g. lock/upsert the projection row first, then read canonical, then write). Rebuild writes obey the same serialization.

**Required correction.** Amend T5-G (and the corresponding registry ASY-01 wording) with the serialization law above. No other T5 decision changes; FIFO remains rejected.

**Reopen.** None — bounded amendment to T5-G. **Proof to close:** amended text + later implementation proof obligation extended to a concurrent-overlap fixture (two overlapping refresh executions with an interleaved lifecycle transition must end at latest state).

## M2 — MAJOR — Restore resurrects revoked ApplicationSessions and post-snapshot access teardown; the ratified restore barrier covers lawful PII erasure only

**Claim.** Packet attack G names Session/RoleAssignment/GroupMembership resurrection explicitly. Current authority answers only the PII half.

**Evidence.** `r10-t4` §16 (T4-N) gates serving on reconciling post-snapshot lawful **UserProfile erasures**. §15 (T4-M) gates on exact-content verification. Nothing in T4-M/N, T3 §10–11 or the registry (STO-11, STO-18, SEC-07) addresses the security mirror image: restoring a recovery point taken *before* an offboarding revives that User's `ApplicationSession` rows (T3 §10 revoked them; restore un-revokes them; the bearer still holding the token regains live authenticated access the moment serving resumes — the AuthN provider is external and was never rewound), plus their memberships and direct grants. The same applies to any post-snapshot RoleAssignment revocation.

**Root cause.** T4-N was derived from the legal non-resurrection obligation (erasure survives restore as a matter of law), and the analysis stopped at PII. Access teardown is the same defect class — a post-snapshot fact whose *inverse* must not silently serve — but has no barrier. Point-in-time restore inherently rewinds facts; the question is which rewound facts are safe to serve without reconciliation, and revoked security state is not.

**Target invariant.** A restored deployment may not serve with authentication/authorization state that current operations already revoked, without that resurrection being either structurally impossible (sessions) or explicitly reviewed (grants/memberships/offboarding).

**Local vs global maximum.** A full independent "security teardown journal" mirroring the erasure journal is overengineering — no named consumer justifies journaling every grant revocation. The global maximum is asymmetric and cheap:

```text
1. sessions — structural: restore readiness requires invalidation of ALL restored
   ApplicationSessions before serving. Zero semantic loss (Sessions are revocable
   mechanism-adjacent state by T3 baseline; users re-authenticate), closes the
   bearer-token hole completely.
2. offboarding/grants — procedural: restore choreography (T7/operations, which T4
   already delegates) must include re-application of known post-snapshot User
   offboardings and security revocations before serving, from the same
   independently retained control-plane source T4-N already requires for erasures;
   residual risk for unknown revocations is stated, not silent.
```

**Required correction.** Amend T4-M/T4-N readiness with law 1; add law 2 to the T7 REOPEN set line "concrete restore/erasure reconciliation choreography" so it becomes "restore/erasure **and post-snapshot security-teardown** reconciliation choreography". Registry: STO-11/STO-18 wording follows.

**Reopen.** None — bounded amendment to T4 §15/§16 + one T7 REOPEN-set line + registry wording. **Proof to close:** amended text; later restore-readiness test proving a restored pre-offboarding session cannot authenticate.

## M3 — MAJOR — `search_refresh` + materialized projection are ratified as always-required machinery while the only consumer that could require them is deferred to T6; every Product-Contract-named search fact is a canonical column

**Claim.** T5 ratifies machinery whose justifying consumer is not yet named — the exact defect class (mechanism before consumer) the Method prohibits and this checkpoint exists to catch.

**Evidence.** `r10-t5` §4 (T5-B): "always-required durable job: search_refresh(document_id)". §7 (T5-F): "Launch Search is one PostgreSQL-backed rebuildable discovery projection", while "the exact searchable field set belongs to T6". Product Contract §5-K names the Launch search facts: "code/title, Document Type, Area, responsible owner and status" plus current-effective favoritism — **all of which are canonical Document/Revision columns in the same PostgreSQL database**. A plain canonical query (or view) serves every named Launch search journey with zero jobs, zero staleness laws, zero rebuild machinery and zero convergence proofs (M1 included). The one consumer class that would genuinely require a materialized projection plus async refresh is a *derived* searchable fact — above all extracted content text for full-text search — and no current authority names content search as Launch scope.

**Root cause.** Stage sequencing: T5 (async/effects) closed before T6 (journeys) decides the field set, so the projection was ratified defensively rather than from a named consumer. The Method's own question — "what concrete consumer requires it now?" — has no recorded answer for the materialization itself; T5-F only answers why *external engines* are rejected.

**Target invariant.** Launch machinery exists only for a named consumer (Method complexity law; ownership-topology future-law §10.5 "no dormant implementation").

**Local vs global maximum.** Keeping the machinery "because T6 probably needs it" is a local maximum that also inherits M1's proof burden for free. Two legal global-maximum resolutions, either acceptable:

```text
(a) name the consumer now — operator declares content full-text search (or another
    derived fact) Launch scope; the projection + search_refresh stand as ratified,
    because text extraction from DOCX/PDF bytes is genuinely async and derived;
(b) conditionalize — T5-B/T5-F are amended so that the projection and search_refresh
    are contingent on T6 naming at least one derived searchable fact; if the T6 field
    set proves fully canonical, Search degrades to a canonical query/view and the
    always-required job kind is deleted rather than shipped dormant.
```

**Required correction.** Operator picks (a) or (b); amend T5-B/T5-F + registry ASY-01 (and ASY-09's rebuild mandate follows the same contingency) accordingly.

**Reopen.** None — bounded amendment; Search's authority boundary (never grants access/effectivity, serve-time canonical recheck) is unaffected and remains frozen either way. **Proof to close:** amended text naming the consumer or the contingency.

---

# 3. LOW findings

## L1 — LOW — Document title has no named owner in T1

Product Contract §4 defines Document identity by example as "`PO-001 — Procedimento de Compras`" and journey K searches/filters by "code/title", but the T1 Controlled Documents family (`r10-t1` §1) models code, Area/responsibility, Template role, origin — never title. Whether title is stable Document identity, mutable Document metadata (changed under which permission, audited how), or governed revision content (retitle requires a revision) is undecided, and each answer has different search/audit/UX consequences. Required: one sentence in T1 §1 naming the owner and mutability law. No reopen — completeness amendment.

## L2 — LOW — OfficialRendition finalization "eligibility" is undefined for terminated attempts

`r10-t5` §6 requires finalization to "reload current eligibility → prove required rendition still absent/eligible", but never defines eligibility when the Submission's attempt was withdrawn/returned or the Revision cancelled before the render lands. Both resolutions are safe (Release-eligibility is independently rechecked under Document serialization), but they differ materially in outcome: creating the semantic OfficialRendition anyway permanently freezes render output as un-GC-able immutable governed content (T4 §13) for a dead attempt; no-op leaves the output reclaimable mechanism state. State the intended rule (no-op recommended). No reopen.

## L3 — LOW — GC eligibility lacks a liveness rule for admitted-but-not-yet-referenced handles

T4-K eligibility + T5-J recheck prove only "not current WorkingContent + no immutable reference + no backup pin". A handle that is READY but *seconds away* from its first reference (autosave swap, SUBMIT preflight) satisfies both proofs and may be marked GC_PENDING, forcing spurious fail-closed autosave/SUBMIT failures (no corruption is reachable: attachment requires READY, and the state machine is one-way). Add the natural completion: eligibility also requires no live T4-F admission binding (or an equivalent bounded grace). No reopen.

## L4 — LOW — No initiator withdraw for an active obsolescence attempt

T2 §8 defines withdraw for Submission attempts only; T2 §7 lets a *participant* RETURN an obsolescence attempt. An initiator who regrets a human-governed obsolescence request cannot terminate it, and while it is active, new-Revision creation is blocked (T2 §10) — the document is frozen until some approver RETURNs. An escape exists, so this is asymmetry, not deadlock; either ratify participant-RETURN as the intended sole escape or add bounded initiator cancellation of an obsolescence attempt. No reopen.

## L5 — LOW — T3/T5 wording tension on provider-disable intent

T3 §10 offboarding transaction includes "insert durable provider-disable intent only if a provider-side effect is required"; T5-L rules no mandatory durable IdP-disable job as Launch baseline. Read together they are consistent (baseline: not required → no intent), but T3's line invites re-adding the job by local reading. Align T3 §10 wording to reference T5-L. No reopen.

---

# 4. Attacks executed with no finding (disposition record)

```text
A  DRAFT→SUBMIT→governance→Rendition→Release→Search — no duplicate authority, no
   impossible intermediate state; tx-coupled enqueue means no lost required work;
   both Release trigger orders (final ACCEPT vs rendition finalization) revalidate
   canonical eligibility under Document serialization; races reduce to M1/L2.
B  RETURN/withdraw/resubmit — old Submissions immutable; stale rendition/search jobs
   cannot become authority (workers reload canonical state; Release binds winning
   Submission under serialization); residual ambiguity is L2 only.
C  Offboarding vs governance/async — T3 §11 eligibility serialization is a complete
   answer for user-driven actions; Step candidate snapshots are not grants and current
   AuthZ is rechecked at action; jobs carry bounded IDs and never authority; an
   all-candidates-offboarded route resolves through the withdraw→fix→resubmit escape.
F  GC vs WorkingContent/Submission/Rendition/backup — one-way OPEN→READY→GC_PENDING +
   in-tx READY/reference proofs + backup pin + immediate pre-delete recheck close every
   corruption path; only the L3 liveness nit remains; crash-safe failure mode is leaked
   storage, correct choice.
H  Viewer vs OfficialRendition — the RV split is clean; SourceOnly release gate is
   satisfied-by-absence so no viewer can become release authority; no silent
   RequireOfficialRendition→SourceOnly downgrade path exists.
I  Audit + async composition — mutation + domain evidence + AuditEvents + River insert
   compose in one local tx by construction (same DB); Audit has no job consumers (not a
   bus); job rows are prunable mechanism (not evidence).
J  Search/AuthZ — the projection stores no authorization data and serve-time re-resolves
   canonical lifecycle + current T3 grants; the stale-ACL-in-index defect class is
   structurally dissolved; Area is stable per DOC-07 so no Area-drift path exists.
§5 Authority sweep — every durable semantic fact has exactly one owner; representation
   policy → Submission snapshot; candidates → GovernanceAttempt; effectivity → Release
   via Revision lifecycle (DOC-09 kills the second authority); descriptors → owning
   semantic record; mechanism facts stay mechanism. Erasure barrier/journal is named
   as independently-retained mechanism with T7 owning choreography — acceptable at this
   altitude (M2 extends its cargo, not its ownership). Document title is the one
   ownerless fact found (L1).
§6 Registry audit — no T1→T5 decision incorrectly REOPEN; no SUPERSEDED leak-back into
   CURRENT/PRESERVE; DEFERRED items create no backward pressure; anti-legacy list
   consistent with all five T-authorities; T6 REOPEN set routes correctly (N3).
§8 Future seams — Distribution (Release+audience), Periodic Review (Document+effective
   Revision), Dossier (reference-only), Evidence (own descriptor+handle), Records
   (immutable history + never-deleted governed content), Export (descriptors), repository
   connectors (copy, external ID never identity), Training (effective truth), Change
   Control (T2 §15 already names the replace+obsolete coexistence reopen trigger),
   pooled tenancy (Company root), CRDT (WorkingContent mechanism boundary) — all attach
   by stable reference; none requires rewriting immutable meaning; none is being
   implemented early.
```

Subtractive pass: no second authority, no compat layer, no duplicate enforcement proving the same property, no obsolete transitional mechanism found; the single subtractive candidate that survives scrutiny is M3.

---

# 5. NOTES

## N1 — Same-database durable jobs give restore coherence for free — record it

Because River rows live in the product database and required intents are tx-coupled (T5-C), any DB recovery point is automatically consistent between semantic facts and their required durable work: restored pending jobs re-run idempotently; post-snapshot jobs vanish *together with* the facts that required them. This is a genuine architectural strength of the one-mechanism/one-DB choice and belongs in T4-L/T5 restore evidence so a future mechanism swap cannot silently lose it.

## N2 — Registry rows whose disposition label and current-meaning column point in opposite directions

OBS-03, GOV-14, DOC-03 are labeled `SUPERSEDED`/`SUPERSEDED / REJECTED` while their "Current meaning" cell states the *current binding law* (the superseded thing being the old inverse decision). Internally coherent once the convention is understood, but a cold reader can misread OBS-03 as "no-human obsolescence is superseded". Cosmetic wording tightening only.

## N3 — T6 REOPEN set does not explicitly name the upload/ingest journey

"editor/viewer provider behavior" plausibly covers it, but source upload/admission UX (T4-E's browser side) is a distinct journey; naming it avoids a scope debate inside T6.

---

# 6. Adjudication guidance

Per the Independence Law (packet §12), this review is evidence, not authority. Recommended disposition: adjudicate M1/M2 as ACCEPT-WITH-FIX bounded amendments (no operator-visible design choice inside them beyond wording), M3 as an explicit operator choice between (a) naming the consumer and (b) conditionalizing, then close L1–L5 as one-line amendments or recorded accepts. None of this requires reopening any ratified stage decision or delaying T6 beyond the adjudication itself. A delta review is warranted only if the M3 adjudication chooses (a) with a materially expanded search scope.

**Everything not named in §2–§3 remains frozen.**

Implementation remains **BLOCKED**.
