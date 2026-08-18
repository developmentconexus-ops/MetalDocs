# R10-B5 — Integration Acceptance / Operator Adjudication

> **Status:** ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED

This record captures the operator's acceptance of the self-reviewed corrected R10-B5 candidate:

`docs/superpowers/analysis/2026-08-18-r10-b5-documentary-context-records-governance-artifact-closure-integrated-candidate.md`

Acceptance means **working authority for continued R10 integration**, not final independent ratification. B3–B5 remain challengeable only by a material later-stage counterexample. Whole-R10 Global Coherence Review + cold independent review remain required before final R10 ratification.

---

## 1. Accepted B5 target

### Documentary Context

- `Dossier` is stable documentary context, never a physical folder, content owner, access-grant source, ERP/CRM/PLM master, workflow container or retention authority.
- `DossierType` remains small: stable code/name/description/status plus explicit eligible DocumentTypes/EvidenceTypes.
- `Dossier↔Document` is M:N over stable Document identity, copies no content, changes no lifecycle/Area/AuthZ and never grants access.
- Dossier scope is exactly one `TenantScope | AreaScope`; stable type/key/scope remain immutable V1; title may change; archive is reversible navigation state only.
- External source identity uses explicit typed provenance/reference semantics; no heuristic merge or source-master takeover.

### Evidence

- Evidence is a separate capture lifecycle, not a DocumentRevision surrogate.
- lifecycle = `DRAFT → CAPTURED → VOIDED`, where VOIDED means invalid MetalDocs capture only.
- DRAFT carries mutable capture preparation; CAPTURE freezes immutable captured payload/metadata and exactly one primary Artifact.
- every CAPTURED Evidence has exactly one immutable primary Dossier and may have secondary contextual Dossier links.
- Evidence reuses the primary Dossier scope; no second independent Evidence scope authority.
- Evidence does not gain REV/Approval/Release by default.
- canonical Evidence name is frozen at CAPTURE; user filename remains provenance only.
- if `{SEQ}` is used, current V1 working target uses one monotonic series per EvidenceType; Dossier-local reset is a bounded reopen trigger.

### Records Governance

- no generic `Record` aggregate/declaration button exists.
- first DocumentRevision Submission and Evidence CAPTURE automatically create one immutable `RetentionBinding` using a closed typed union: exact DocumentRevision XOR exact Evidence.
- type-level retention is explicit `NoMinimum | KeepFor(value,DAYS|MONTHS|YEARS) | Indefinite`; no NULL-as-policy and no hardcoded legal periods.
- RetentionBinding snapshots policy; later type-policy changes affect future subjects only.
- retention anchors derive from canonical lifecycle facts, not Audit or duplicate mutable `retain_until` state.
- RetentionExtension is append-only, may only lengthen the minimum, requires a known anchor and is forbidden after a DispositionFence exists.
- expiry means disposition eligibility only; never automatic deletion.
- LegalHold is independent of retention and scopes only Evidence | stable Document | Dossier V1.
- `LegalHoldSubject(hold,binding)` is immutable materialization evidence; already materialized subjects never disappear because of unlink/lifecycle changes.
- active Document/Dossier holds continue materializing newly entering RetentionBindings while the live scope remains applicable.
- Hold activation is fail-closed/all-or-nothing for current preservable scope if a subject is already fenced/inconsistent.
- disposition uses semantic `DispositionFence` as the irreversible authorization barrier; worker/retry/lease state remains R10-D mechanism truth.
- completed `DispositionRecord` is written only after required physical removal/verification and semantic cleanup succeed.
- business lifecycle and records disposition are orthogonal; no `DISPOSED` DocumentRevision/Evidence state is introduced.

### Artifact closure

- no generic `ArtifactOwner(owner_type,owner_id)` or mutable reference-count authority.
- every confirmed Artifact has exactly one semantic retention root: one DocumentRevision **or** one Evidence.
- multiple semantic references inside that same root are allowed; cross-root Artifact-row reuse is rejected.
- identical bytes in another retention root may produce another Artifact semantic row with the same SHA-256; physical dedupe remains provider/mechanism freedom.
- Artifact survival/deletion is proven through typed semantic reachability, not independent Artifact retention policy.
- final semantic-reference removal also removes the Artifact semantic row in the same local semantic transaction; physical cleanup may reconcile later only where no governed disposition claim is being made.

---

## 2. Accepted B3 bounded refinements exposed by B5

The operator accepted exactly these three bounded refinements to the non-final B3 working target. No unrelated B3/B4 semantic is reopened.

### B3-R1 — DocumentOrigin retention-safe provenance

Old accepted working shape used a strong FK from `DocumentOrigin` to source `RevisionSubmission`. B5 proved that would silently force indefinite retention of template-source Submission/Artifact whenever a derived Document survives.

Current R10 working target:

```text
DocumentOrigin
  derived_document_id             FK Document
  source_template_revision_id     FK permanent DocumentRevision identity skeleton
  source_submission_id_snapshot   UUID value
  source_submission_digest        SHA-256
  source_artifact_sha256           SHA-256
  source_content_format
  created_at
```

Derived creation still proves the exact current-effective source Submission under B3/B4 serialization. Later lawful source-unit disposition cannot rewrite origin provenance and does not require keeping the source Submission/Artifact alive forever.

### B3-R2 — canonical terminal retention anchors

`DocumentRevision` gains domain-owned state-coupled lifecycle instants:

```text
cancelled_at
obsoleted_at
```

They are written exactly once with their corresponding CI lifecycle transition. Audit is never the retention-anchor authority.

Native supersession time is **not duplicated**; it remains B4 `ReleaseRecord.released_at` for the Release that names the revision as `prior_effective_revision_id`.

### B3-R3 — permanent revision identity vs disposable dictionary payload

`DocumentRevision` must retain its minimal identity/lifecycle skeleton after lawful disposition so REV ordinals can never be reused. Therefore immutable dictionary content may not remain co-located permanently on that skeleton.

Current working target:

```text
DocumentRevision
  permanent identity/lifecycle skeleton

RevisionDictionarySnapshot
  revision_id FK DocumentRevision
  immutable snapshot
```

`RevisionDictionarySnapshot` remains part of Submission construction and belongs to the DocumentRevision retention unit; completed records disposition may remove it while the revision identity row survives.

---

## 3. DocumentRevision retention-unit closure accepted

A retained DocumentRevision unit includes the immutable governed history whose meaning is subordinate to that Revision, including at minimum:

### B3

- RevisionDictionarySnapshot;
- all RevisionSubmissions/manifests/digests;
- exact Submission source Artifacts;
- revision-bound immutable submitted/template structured provenance where not already wholly represented in Submission;
- PeriodicReviewRecords;
- other Revision-governance evidence that has no independent lifecycle.

### B4

- SubmissionApprovalRequirement;
- ApprovalInstance/Step participant/decision/reassignment evidence;
- bounded consumed fresh-auth decision evidence;
- SubmissionFeedback;
- Renditions and their Artifacts;
- ReleasePlan / ReleaseRecord;
- DistributionObligations;
- AcknowledgementRecords.

This does **not** recursively retain independent authorities such as User, Group, DocumentType or ApprovalPolicy merely because their identities were referenced. Necessary historical meaning uses stable IDs/bounded snapshots while those independent owners retain their own lifecycle.

---

## 4. Accepted concurrency/failure laws

### First Submission / CAPTURE

- first RevisionSubmission creates the Revision RetentionBinding in the same local semantic transaction;
- Evidence CAPTURE freezes the EvidenceCapture, exact Artifact relation and RetentionBinding atomically;
- applicable active hold materialization occurs in the same semantic boundary when a new binding enters held scope.

### Hold vs disposition

The semantic ordering is fail-closed:

```text
subject root
→ relevant Dossier/context roots where needed
→ RetentionBinding
→ hold/fence decision
```

DispositionFence creation must recompute eligibility and reconcile both already-materialized and currently applicable live hold scopes before crossing the irreversible barrier.

A competing hold that loses the race to an existing fence fails visibly; it may not partially activate over the remainder of the current preservable scope.

### External physical deletion

No cross-database/object-store atomicity is claimed:

```text
semantic fence commit
→ external physical removal/verification with retry/reconcile
→ semantic payload cleanup + immutable DispositionRecord commit
```

A fence without a DispositionRecord means disposal is in-progress/incomplete, never falsely completed.

---

## 5. Method disposition

The accepted result is the current **Global Maximum working candidate** because it preserves the essential domain distinctions while rejecting generic ECM/records-platform machinery:

```text
Dossier context
+ specialized Evidence capture
+ closed typed RetentionBinding
+ materialized LegalHold
+ explicit disposition fence
+ one semantic Artifact retention root
```

Rejected/deferred accidental complexity includes:

- folder/binder ownership model;
- generic Record declaration aggregate;
- polymorphic owner/subject registry;
- generic retention expression engine;
- Artifact-specific retention authority;
- automatic expiry deletion;
- generic eDiscovery/custodian graph;
- provider ObjectLock as business authority;
- ref-count correctness authority;
- generic multi-stage disposition workflow without a real consumer.

---

## 6. Residual proof obligations / reopen triggers

B5 acceptance does not waive later proof obligations:

- B6 must produce the final whole B1–B5 same-commit/cross-owner matrix and classify surviving immutable Audit/privacy evidence;
- the whole lock graph must be demonstrated acyclic under `READ COMMITTED` before implementation;
- R10-C must define physical delete/verification, provider ObjectLock interaction, malware/storage/restore guarantees;
- R10-D must define disposition retry/lease/reconciliation without becoming business authority;
- R10-E owns user journeys and fail-visible disposition/hold UX;
- R10-F owns historical migration, target-vs-legacy deletion mapping and post-disposition/privacy cutover rules.

Reopen B5 only for material evidence such as a real cross-retention-root Artifact semantic consumer, required Dossier-local sequence semantics, a legally required distinct disposition approval workflow, a new real retention subject family, or a failure mode proving the fence/hold model cannot preserve the target invariant.

---

## 7. Status transition

Operator adjudication:

```text
R10-B3 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
         with B5-approved bounded refinements B3-R1/R2/R3

R10-B4 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL

R10-B5 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED

R10-B6 = NEXT / DESIGN ONLY

implementation = BLOCKED
```

Whole-R10 independent review remains deferred until the integrated design is complete unless a truly exceptional material blocker arises.