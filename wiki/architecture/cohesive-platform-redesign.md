# Cohesive Platform Redesign — Active Architecture Authority

> **Status:** Active design authority — **R9.5 FROZEN; R10-A/B1/B2 PROMOTED; R10-B3/B4/B5 ACCEPTED FOR R10 INTEGRATION / NON-FINAL; R10-B6 NEXT; NO PRODUCT IMPLEMENTATION AUTHORIZED**  
> **Established:** 2026-08-14  
> **R9.5 freeze ratified:** 2026-08-17  
> **R10-A/B1/B2 promoted:** 2026-08-17  
> **R10-B3 integration acceptance:** 2026-08-17 — non-final / not independently ratified  
> **R10-B4 integration acceptance:** 2026-08-18 — non-final / not independently ratified  
> **R10-B5 integration acceptance:** 2026-08-18 — non-final / not independently ratified  
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** `docs/engineering/standards/root-cause-global-maximum-method.md`  
> **Frozen product/domain ledger:** `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`  
> **Promoted R10 technical authority through B2:** `wiki/architecture/r10-technical-architecture.md`

Accepted current R10 integration inputs:

- `docs/superpowers/analysis/2026-08-17-r10-b3-controlled-information-artifact-integrated-candidate.md`
- `docs/superpowers/analysis/2026-08-18-r10-b4-approval-rendition-release-distribution-integrated-candidate.md`
- `docs/superpowers/analysis/2026-08-18-r10-b4-integration-acceptance.md`
- `docs/superpowers/analysis/2026-08-18-r10-b5-documentary-context-records-governance-artifact-closure-integrated-candidate.md`
- `docs/superpowers/analysis/2026-08-18-r10-b5-integration-acceptance.md`

---

## 1. Purpose / north star

> **MetalDocs is the system of record for product/organizational identity, governance, revision, evidence and documentary context. Authentication credential/upstream identity-provider truth may be provider-owned; physical storage, authoring/editor technology, viewers and upstream repositories are replaceable providers/connectors around the MetalDocs kernel.**

Target posture:

- smallest professional architecture preserving real invariants;
- one canonical authority per business/system fact;
- one company per V1 deployment; common code/build/migrations; no customer forks;
- commodity mechanism may be externalized without taking semantic authority;
- no speculative ECM/BPM/ReBAC/low-code/object/records platform;
- current implementation is evidence, never automatic target entitlement.

Fresh sessions follow `AGENTS.md` → Method → current handoff → this page → frozen ledger → promoted R10 authority → accepted non-final B3/B4/B5 working inputs.

R3–R9.5 remains frozen historical/product-domain authority except where an explicitly recorded bounded reopen/refinement is operator-approved for current R10 integration. B3/B4/B5 remain non-final and challengeable only by material later-stage counterexample.

---

## 2. Deployment / ownership / access posture

V1 deployment invariant:

> **One company per deployment. `Tenant` is the singleton company root and TenantScope target, not a database partition.**

Promoted B1 substrate:

```text
one PostgreSQL product DB / schema metaldocs
UUID PKs
ordinary typed FKs
cross-owner RESTRICT / NO ACTION
no universal tenant/company/deployment partition column
no Tenant/Area/role/Permission RLS policy engine
serving DB role non-owner + NOSUPERUSER
READ COMMITTED
same-local-commit frozen cross-owner invariants
no provider DB atomicity dependency
```

Business bounded contexts remain exactly 8:

```text
Authentication
Organization
Authorization
Controlled Information
Approval
Documentary Context
Records Governance
Distribution
```

Supporting semantic owners remain exactly 3:

```text
Artifact
Audit
Interchange
```

Notifications are attributed support; Search is rebuildable projection; composition/platform/workers/providers own mechanisms only.

Authentication uses Keycloak as V1 AuthN provider while MetalDocs owns provider-subject binding, local Session and bounded assurance consumption. Provider roles/groups/orgs/permissions never become canonical AuthZ.

Authorization remains live additive/default-deny from static Role/Permission catalogs + current RoleAssignment + scope + owning-domain relationship/governance predicate. `tenant_owner` is a grant bundle, never bypass.

---

## 3. R10-B3 accepted working target

Core:

```text
Document
→ business DocumentRevision
→ WorkingContent + monotonic working_version OCC
→ immutable RevisionSubmission + deterministic governed manifest/digest
→ exact provider-neutral Artifact
```

Key laws:

- Document identity ≠ Revision ≠ autosave ≠ exact bytes;
- WorkingContent is sole mutable DRAFT authority;
- SUBMIT freezes exactly one accepted generation and consumes the OCC generation;
- same-REV return/resubmit never mutates old Submission;
- Artifact is exact-byte identity, not provider location and not global content-hash business identity;
- Template reuses Document lifecycle; no parallel TemplateVersion aggregate;
- one-open/one-EFFECTIVE have structural backstops;
- downstream governance binds exact immutable Submission.

### B5-approved bounded B3 refinements

Current R10 working target additionally carries:

1. `DocumentOrigin` no longer holds a strong FK to source Submission that would force indefinite template-source retention. It keeps source Revision identity + exact immutable Submission/digest/hash provenance snapshots.
2. `DocumentRevision` owns one-shot canonical `cancelled_at` / `obsoleted_at`; native superseded time remains B4 ReleaseRecord authority and is not duplicated.
3. `RevisionDictionarySnapshot` is separated from the permanent Revision identity skeleton so lawful disposition may remove retained payload without deleting/reusing the revision ordinal or mutating immutable JSON in place.

These are bounded technical refinements exposed by real B5 Records-Governance counterexamples; no unrelated B3 semantic is reopened.

---

## 4. R10-B4 accepted working target

```text
RevisionSubmission
→ immutable SubmissionApprovalRequirement / ReleasePlan snapshots
→ one sequential Approval/Governance Step model
→ detached SubmissionFeedback
→ immutable Rendition when produced
→ automatic system-owned Release
→ concrete DistributionObligation snapshot in winning Release transaction
→ explicit immutable AcknowledgementRecord
```

Key laws:

- one Step semantic type; historical `review|approval` discriminator is removed in current R10 working target;
- Step carries label, ordered position, `NamedUser|Group|RoleInArea`, `ANY|ALL`, optional fresh-auth and due date;
- active participants may inspect exact Submission in-product, including exact PDF rendition when available, and may comment/annotate/suggest;
- feedback never mutates Submission; applying returned suggestion is later B3 OCC mutation;
- strict SoD has live/domain + structural backstops;
- required fresh-auth is Authentication-owned input and immutable Approval-owned consumption evidence;
- return-for-changes requires reason and returns same Revision to DRAFT without resetting B3 generation;
- Approval requirement/PolicyVersion + representation requirement are snapshotted per Submission;
- viewer capability is independent from official representation policy;
- Release is automatic/system-owned and sole effectivity transition;
- Release atomically establishes EFFECTIVE/SUPERSEDED and concrete Distribution obligations;
- Distribution never grants access; explicit acknowledgement is the only obligation completion signal.

### Explicit bounded R9.5 Approval refinement

Frozen historical ledger contains `Step.purpose = review | approval`. Operator-approved R10 overlay removes it because external evidence + Structural Inversion Test showed it encoded legacy editor/UI ceremony rather than an invariant. Collaboration and fresh-auth remain orthogonal capabilities; `PeriodicReview` remains the distinct CI review concept.

---

## 5. R10-B5 accepted working target

### Documentary Context

- Dossier is stable documentary context only; never physical folder/content owner/access grant/retention authority/ERP-PLM master.
- DossierType remains small with explicit eligible DocumentTypes/EvidenceTypes; no custom fields/forms/workflow/ACL/completeness engine.
- Dossier↔Document is M:N over stable Document identity, copies no content and changes no lifecycle/Area/AuthZ.
- Dossier scope = exactly one TenantScope|AreaScope; stable type/key/scope; title mutable; archive reversible navigation only.
- External source identity/provenance remains explicit; no heuristic source takeover/merge.

### Evidence

- lifecycle = `DRAFT → CAPTURED → VOIDED`; VOIDED means invalid MetalDocs capture only;
- CAPTURE freezes immutable payload/metadata + exactly one primary Artifact and creates RetentionBinding;
- every CAPTURED Evidence has exactly one immutable primary Dossier and may have secondary Dossier context;
- Evidence reuses primary-Dossier scope and does not use REV/Approval/Release by default;
- canonical name freezes at CAPTURE; user filename remains provenance;
- current `{SEQ}` candidate uses one monotonic EvidenceType series; Dossier-local reset is a reopen trigger.

### Records Governance

- no generic Record aggregate/declaration operation;
- first DocumentRevision Submission / Evidence CAPTURE automatically create immutable typed RetentionBinding;
- policy = explicit `NoMinimum | KeepFor(value,DAYS|MONTHS|YEARS) | Indefinite`;
- anchor derives from canonical owner lifecycle facts; Audit never becomes retention-clock authority;
- RetentionExtension only lengthens and cannot be added after DispositionFence;
- LegalHold scopes V1 = Evidence | stable Document | Dossier and materialize exact RetentionBindings, including newly entering live scope;
- Hold activation is fail-closed/all-or-nothing for current preservable scope when a fence/inconsistent subject is encountered;
- expiry only means eligibility; never automatic delete;
- DispositionFence is semantic irreversible authorization barrier; worker/retry state belongs R10-D;
- DispositionRecord means verified physical + semantic completion, not merely requested deletion;
- business lifecycle and records disposition are orthogonal axes.

### Artifact closure

- every confirmed Artifact has exactly one semantic retention root: one DocumentRevision or one Evidence;
- multiple references within that root are allowed; cross-root Artifact-row reuse is rejected;
- identical bytes may have separate Artifact semantic rows across roots; physical dedupe remains mechanism freedom;
- no Artifact retention policy, generic owner registry or ref-count authority;
- semantic Artifact survival/deletion is derived from typed reachability.

### Retained DocumentRevision unit

The retained Revision unit includes its subordinate immutable governance history, including B3 Submission/dictionary/provenance/PeriodicReview state and B4 Approval/fresh-auth/feedback/Rendition/Release/Distribution/Acknowledgement evidence. Independent authorities such as User, Group, DocumentType and ApprovalPolicy are not recursively swallowed into the retention unit merely because referenced.

---

## 6. Storage / privacy / migration posture carried forward

- ManagedArtifactStore port/conformance remains first-class; production malware inspection remains mandatory/fail-closed before untrusted bytes become confirmed Artifact;
- provider WORM/ObjectLock/Purview may enforce physical retention but never becomes MetalDocs Records-Governance authority;
- User/data-subject privacy remains separable from immutable governance evidence; B6 must classify surviving Audit/post-disposition skeleton fields;
- no generic privacy workflow or mandatory crypto-erasure platform without named immutable Target Data;
- Backup/Restore, Historical Migration, Governed Subject Export and explicit IMPORT_COPY/PUBLISH_COPY remain; Tenant Portability Export is deferred.

---

## 7. Implementation gate

**CLOSED.** B3/B4/B5 are accepted only for continued R10 integration, not final ratification.

Before implementation:

```text
B6
→ R10-C
→ R10-D
→ R10-E
→ R10-F
→ Whole-R10 integration
→ Global Coherence Review
→ cold independent review
→ operator adjudication
→ final R10 ratification
→ implementation specification/plan
→ code
```

---

## 8. Exact next step — R10-B6 Audit + Interchange + Cross-owner Atomicity

Open **R10-B6** as one integrated research-heavy design batch from promoted B1/B2 + accepted non-final B3/B4/B5.

B6 must jointly close:

```text
Audit
  append-only AuditEvent semantic skeleton
  trusted actor/resource/operation/time meaning
  field-by-field privacy / PII-minimized surviving skeleton
  erasable human-readable enrichment boundary
  forensic grant/revoke reconstruction after current-state deletion
  tamper-evidence/query/export ownership and separate Audit retention regime

Interchange
  Historical Migration batch/plan/dry-run/item outcome/reconciliation state
  imported history vs native domain-fact boundary
  Governed Subject Export process/package semantics
  External Repository IMPORT_COPY / PUBLISH_COPY truth
  required connection/reference state used by Dossier ExternalReference / publication
  no Tenant Portability Export V1

Cross-owner coherence
  final B1–B5 same-local-commit matrix
  exact required Audit append points
  exact required durable-intent points routed to R10-D
  published transaction-composition seams; no nested owner commits/repository imports
  final READ COMMITTED lock-order/deadlock challenge
  imported-history DB coherence with native constraints without fabricating native facts
```

Explicitly challenge:

```text
B2 offboarding/privacy vs immutable Audit evidence
B3/B4/B5 critical mutations vs same-commit Audit
B5 disposition vs Audit's separate retention regime
B5 post-disposition skeleton privacy
Historical Migration without synthetic native Approval/Release/actors/timestamps
Export completeness + canonical AuthZ
External repository effects without provider atomicity
whole B1–B5 lock graph
```

Route later work correctly:

```text
physical store/object-lock/malware/restore       → R10-C
async jobs/retry/lease/projections/effects       → R10-D
API/frontend/viewer/editor journeys              → R10-E
historical cutover/legacy deletion/bootstrap     → R10-F
```

Implementation remains **BLOCKED**.