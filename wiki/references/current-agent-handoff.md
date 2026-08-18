# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — Cohesive Platform Redesign / **R9.5 FROZEN + R10-A/B1/B2 PROMOTED + R10-B3/B4/B5/B6 ACCEPTED FOR R10 INTEGRATION / NON-FINAL + R10-B INTEGRATED DESIGN BLOCK COMPLETE / NON-FINAL + R10-C NEXT**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file — current checkpoint / exact next step
4. `wiki/architecture/cohesive-platform-redesign.md` — active program/global-coherence authority
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — frozen R3–R9.5 product/domain authority
6. `wiki/architecture/r10-technical-architecture.md` — promoted R10 authority through integrated B2
7. `docs/superpowers/analysis/2026-08-17-r10-b3-controlled-information-artifact-integrated-candidate.md`
8. `docs/superpowers/analysis/2026-08-18-r10-b4-approval-rendition-release-distribution-integrated-candidate.md`
9. `docs/superpowers/analysis/2026-08-18-r10-b4-integration-acceptance.md`
10. `docs/superpowers/analysis/2026-08-18-r10-b5-documentary-context-records-governance-artifact-closure-integrated-candidate.md`
11. `docs/superpowers/analysis/2026-08-18-r10-b5-integration-acceptance.md`
12. `docs/superpowers/analysis/2026-08-18-r10-b6-audit-interchange-cross-owner-atomicity-integrated-candidate.md`
13. `docs/superpowers/analysis/2026-08-18-r10-b6-integration-acceptance.md`
14. older review/current-state artifacts only when auditing how an already-promoted/accepted decision was challenged

Git history is archive. Current code/schema/OpenAPI/module docs remain current-state/migration evidence only.

---

## Current checkpoint

```text
R3–R9.5 = FROZEN / GCR-REFINED / SINGLE-COMPANY-REFINED

R10-A   = CLOSED / APPROVED / GCR-REFINED / SINGLE-COMPANY-REFINED

R10-B   = INTEGRATED DESIGN BLOCK COMPLETE / NON-FINAL
R10-B1  = CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED
R10-B2  = CLOSED / APPROVED / INTEGRATED
R10-B3  = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED
R10-B4  = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED
R10-B5  = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED
R10-B6  = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED

R10-C   = NEXT / DESIGN ONLY
R10-D   = NOT STARTED
R10-E   = NOT STARTED
R10-F   = NOT STARTED

implementation = BLOCKED
```

Promoted durable authority remains R3–R9.5 + R10-A/B1/B2. B3–B6 are current non-final integration authority. Later stages may reopen only the materially implicated earlier decision when a real counterexample appears.

B6 included the internal whole-R10-B coherence/adversarial challenge required to close the B design block for continued integration. It does **not** replace the later Whole-R10 Global Coherence Review + cold independent review before final ratification.

---

## R10 working-mode rule

```text
repo authority
→ evidence + directed external research where useful
→ DevelopmentConexus Method
→ integrated candidate
→ internal adversarial challenge
→ cross-stage coherence
→ operator understanding/adjudication
→ ACCEPTED FOR R10 INTEGRATION / NON-FINAL
```

Do not return to per-table/per-field microdecision gates. Do not reopen accepted decisions for implementation convenience or hypothetical futures.

---

## Accepted R10-B kernel

### B3 — Controlled Information + Artifact

```text
Document
→ business DocumentRevision
→ WorkingContent + monotonic OCC generation
→ immutable RevisionSubmission
→ exact provider-neutral Artifact
```

Key laws:

- exact Submission is native downstream governance identity;
- same-REV return/resubmit never mutates old Submission;
- SUBMIT consumes WorkingContent generation;
- Artifact is exact-byte identity, not storage location;
- one-open/one-EFFECTIVE have structural backstops;
- Template reuses Document lifecycle; no parallel TemplateVersion aggregate.

### B4 — Approval / Rendition / Release / Distribution

```text
RevisionSubmission
→ SubmissionApprovalRequirement / ReleasePlan snapshots
→ one sequential Governance Step model
→ detached SubmissionFeedback
→ optional immutable Rendition
→ automatic system-owned Release
→ concrete Distribution obligations
→ explicit acknowledgement
```

Key current overlay: historical `Step.purpose = review|approval` is removed from the current R10 working target. Collaboration and fresh-auth are orthogonal capabilities.

### B5 — Documentary Context / Evidence / Records Governance

- Dossier is documentary context only; never folder/content owner/access grant/retention authority.
- Evidence lifecycle = `DRAFT → CAPTURED → VOIDED`; CAPTURE freezes exact Artifact/payload and creates RetentionBinding.
- no generic Record declaration.
- retention policy = explicit `NoMinimum | KeepFor | Indefinite`; expiry never auto-deletes.
- LegalHold materializes exact RetentionBindings and continues capturing newly entering live scope.
- DispositionFence is the irreversible semantic authorization barrier; DispositionRecord means verified physical + semantic completion.
- every Artifact has exactly one semantic retention root: one DocumentRevision or one Evidence.

### B6 — Audit / Interchange / final cross-owner atomicity

Audit:

- AuditEvent is forensic/timeline evidence, never domain state;
- required Audit appends in the same local transaction as governed mutation;
- ordinary serving SQL cannot directly mutate AuditEvent/AuditChainHead; a narrow Audit-owned DB append primitive owns sequence/hash-chain mutation;
- JCS + SHA-256 chain; `AuditChainHead` is the final semantic lock;
- immutable Audit skeleton is PII-minimized/non-human-readable and `Indefinite` V1; human-readable UserProfile enrichment is read-time/erasable;
- Audit is not outbox, telemetry, event sourcing, Approval/effectivity/AuthZ/retention authority.

Interchange:

- Historical Migration, Governed Subject Export, IMPORT_COPY and PUBLISH_COPY remain distinct contracts;
- migration never fabricates native Submission/Approval/Release/User actions;
- true dry-run + deterministic outcomes + reconciliation; APPLY atomic per semantic import unit;
- complete export snapshot uses short `REPEATABLE READ`, provider-independent JCS/SHA-256 manifest and fails closed on authorization-incomplete closure;
- PUBLISH_COPY request ≠ success receipt; IMPORT_COPY target mutation + receipt commit atomically.

Cross-owner:

- one local PostgreSQL transaction through published owner seams for frozen atomicity;
- required durable async intents insert before Audit and execute later under R10-D;
- no nested owner commit/repository import to fake atomicity;
- provider/network effects never join DB commit;
- before implementation the whole B1→B6 lock graph must be mechanically shown acyclic.

---

## Current bounded refinements carried by R10-B

### Approval bounded refinement

Historical frozen `review|approval` Step discriminator is replaced by one Governance Step semantic model in current R10 integration authority.

### B5-approved B3 refinements

1. `DocumentOrigin` uses permanent source Revision identity + exact immutable source provenance snapshots rather than a retention-blocking source-Submission FK.
2. `DocumentRevision.cancelled_at/obsoleted_at` are native lifecycle facts; supersession time remains B4 ReleaseRecord authority.
3. immutable dictionary payload lives in `RevisionDictionarySnapshot`, separable from permanent Revision identity for disposition.

### B6-approved imported-history refinements

4. `RevisionOrdinalReservation` preserves historically real ordinals that cannot lawfully become fake Revisions.
5. `RevisionImportedContent` represents exact imported Revision content without fake native Submission.
6. imported governance/history uses `history_kind=NATIVE|IMPORTED` + target-owner imported governance state rather than synthetic Approval/Release facts.
7. Template origin may pin `NATIVE_SUBMISSION | IMPORTED_REVISION_CONTENT` source kind through immutable provenance snapshots.
8. native terminal timestamps remain native-only; imported historical terminal timestamps/unknown stay imported provenance.
9. imported Evidence uses `history_kind` + immutable imported capture state rather than fake native capture actor/time.
10. imported Revision/Evidence obtains RetentionBinding in the migration semantic unit even when no native SUBMIT/CAPTURE event exists; unknown anchor never silently becomes disposition-eligible.
11. current Tenant Dictionary is never resolved merely to manufacture imported historical dictionary state.

---

## Exact next step — R10-C Artifact / Records Physical Integrity

Open **R10-C** in the same integrated research-heavy design mode from promoted B1/B2 + accepted non-final B3–B6.

R10-C owns physical integrity/mechanism semantics, not business governance meaning.

At minimum close:

```text
ManagedArtifactStore
  provider-neutral physical-location/reference model
  provider conformance contract
  Local dev/test and reference-production adapter posture

Ingest/confirmation
  temporary staging
  malware inspection / fail-closed production gate
  exact bytes/hash/size/format verification
  semantic confirmation handoff without confirmed orphan Artifact

Physical integrity
  read verification policy
  relocation/copy verification
  restore verification
  storage-provider version/object identity treatment
  no change to Artifact business identity or Submission digest on relocation

Records enforcement
  Object Lock/WORM integration as enforcement only
  B5 LegalHold/Retention/Disposition remain business authority
  physical delete after DispositionFence
  provider-neutral deletion verification before DispositionRecord completion

Recovery/cleanup
  failed staging/render/export physical cleanup
  orphan mechanism-state reconciliation
  backup/restore integrity
  privacy non-resurrection constraints where physical restore is implicated
```

Route correctly:

```text
async retry/lease/jobs/reconciliation execution → R10-D
API/frontend/viewer/download journeys           → R10-E
historical cutover/legacy deletion/bootstrap    → R10-F
```

Do not turn provider buckets/keys/version IDs/ObjectLock flags into Artifact/Records business identity or authority.

Current implementation is evidence only. Implementation remains **BLOCKED**.