# Cohesive Platform Redesign — Active Architecture Authority

> **Status:** Active design authority — **R9.5 FROZEN; R10-A/B1/B2 PROMOTED; R10-B3/B4 ACCEPTED FOR R10 INTEGRATION / NON-FINAL; R10-B5 NEXT; NO PRODUCT IMPLEMENTATION AUTHORIZED**  
> **Established:** 2026-08-14  
> **R9.5 freeze ratified:** 2026-08-17  
> **R10-A/B1/B2 promoted:** 2026-08-17  
> **R10-B3 integration acceptance:** 2026-08-17 — non-final / not independently ratified  
> **R10-B4 integration acceptance:** 2026-08-18 — non-final / not independently ratified  
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** `docs/engineering/standards/root-cause-global-maximum-method.md`  
> **Frozen product/domain ledger:** `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`  
> **Promoted R10 technical authority through B2:** `wiki/architecture/r10-technical-architecture.md`  
> **Accepted B3 working candidate:** `docs/superpowers/analysis/2026-08-17-r10-b3-controlled-information-artifact-integrated-candidate.md`  
> **Accepted B4 working candidate:** `docs/superpowers/analysis/2026-08-18-r10-b4-approval-rendition-release-distribution-integrated-candidate.md`  
> **B4 acceptance/adjudication record:** `docs/superpowers/analysis/2026-08-18-r10-b4-integration-acceptance.md`

## 1. Purpose / north star

MetalDocs is being redesigned as one coherent product before the next large implementation wave.

> **MetalDocs is the system of record for product/organizational identity, governance, revision, evidence and documentary context. Authentication credential/upstream identity-provider truth may be provider-owned; physical storage, authoring/editor technology, viewers and upstream repositories are replaceable providers/connectors around the MetalDocs kernel.**

Target posture:

- smallest professional architecture preserving real invariants;
- one authority for each business/system fact;
- one company per V1 deployment;
- same code/build/migrations for every deployment; no customer forks;
- commodity mechanisms may be externalized without surrendering domain authority;
- no speculative ECM/BPM/ReBAC/low-code/object platform;
- current implementation is evidence, never automatic target entitlement.

## 2. Authority / evidence

Fresh sessions follow `AGENTS.md` → Method → current handoff → this page → frozen ledger → promoted R10 authority → accepted non-final B3/B4 working candidates.

R3–R9.5 remains frozen product/domain authority except where an explicitly documented bounded reopen is operator-approved for current R10 integration. Accepted B3/B4 candidates are **working integration authority only**, not final independent ratification.

Whole-R10 Global Coherence Review + cold independent review occurs before final R10 ratification unless an exceptional material trust-boundary/irreversible/cross-repository blocker requires earlier independent review.

## 3. Deployment / Authentication / Organization / Authorization

V1 deployment invariant:

> **One company per deployment. `Tenant` is the singleton company root and TenantScope target, not a database partition.**

Promoted B1 substrate:

```text
one PostgreSQL product DB / schema metaldocs
UUID technical PKs
ordinary typed FKs
cross-owner RESTRICT / NO ACTION
no universal tenant/company/deployment partition column
no Tenant/Area/role/Permission RLS policy engine
serving DB role non-owner + NOSUPERUSER
READ COMMITTED
same-local-commit business state + required Audit/durable intent
no provider DB atomicity dependency
```

Authentication:

- Keycloak is V1 AuthN provider;
- MetalDocs owns only `ProviderSubjectBinding` + `ApplicationSession` semantic state;
- stable provider identity = `issuer + subject`;
- provider roles/groups/orgs/claims never become MetalDocs AuthZ authority;
- fresh-auth evidence is Authentication-owned input; consuming domains persist their own bounded consumption evidence.

Organization:

```text
Tenant
Area
User
UserProfile
Group
GroupMembership
```

Area retirement preserves history and blocks prohibited new references. UserProfile is erasable human-readable enrichment. Groups are flat/company-wide; hard delete fails while any live typed reference exists.

Authorization:

- static product Role/Permission catalogs;
- one persisted `RoleAssignment` family;
- five roles: viewer, author, approver, area_manager, tenant_owner;
- exact 43-permission catalog/bundles remain promoted B2 authority;
- `tenant_owner` is a role bundle, never bypass;
- canonical evaluation = live grants → static bundle → scope → domain relationship → governance → default deny;
- no custom roles/deny/ReBAC/effective-permission store/Session AuthZ snapshot/provider bridge.

## 4. R10-A ownership

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

Notifications are attributed support; Search is rebuildable projection.

## 5. R10-B3 accepted non-final working target

Core:

```text
Document
→ DocumentRevision
→ WorkingContent + monotonic working_version OCC
→ immutable RevisionSubmission + deterministic governed manifest/digest
→ exact provider-neutral Artifact
```

Small typed adjuncts cover DocumentType/category/numbering/dictionary, template Document role/use/spec/origin, EditorialComment and PeriodicReview.

Key B3 laws:

- Document identity is not bytes/revision/autosave;
- business REV is not technical autosave/checkpoint;
- WorkingContent is sole mutable DRAFT authority;
- every governed DRAFT mutation uses one OCC generation;
- SUBMIT freezes one coherent WorkingContent generation into immutable `RevisionSubmission` and consumes/increments the OCC generation;
- same-REV return/resubmit never mutates old Submission;
- Submission digest binds exact Artifact + governed state/provenance, never storage location;
- Artifact is exact-byte identity, provider-neutral and not globally hash-unique business identity;
- Template reuses Document lifecycle; no parallel TemplateVersion aggregate;
- one-open/one-EFFECTIVE have structural backstops;
- template creation and PeriodicReview serialize with B4 Release on the Document root;
- B5 must finish global Artifact typed-owner/disposition closure.

## 6. R10-B4 accepted non-final working target

Current working model:

```text
RevisionSubmission
→ SubmissionApprovalRequirement / ReleasePlan snapshots
→ one sequential Approval/Governance Step model
→ detached SubmissionFeedback
→ immutable Rendition when produced
→ automatic system-owned Release
→ concrete DistributionObligation snapshot in winning Release transaction
→ explicit immutable AcknowledgementRecord
```

### Approval

- ApprovalPolicy has stable identity + immutable numbered versions;
- each Submission snapshots `NO_HUMAN_APPROVAL | USE_POLICY(exact version)`;
- there is one Step semantic type, not `review|approval` types;
- Step fields: `label`, ordered position, `NamedUser|Group|RoleInArea`, `ANY|ALL`, optional fresh-auth, optional due date;
- Step label is human/business language only;
- actor rule resolves to concrete Users at Step activation; action-time AuthZ stays live;
- strict SoD: creator/submitter cannot ACCEPT own Submission; same User cannot ACCEPT two Steps; no role bypass;
- required fresh-auth evidence is one-shot/bounded and snapshotted in immutable decision evidence;
- `RETURN_FOR_CHANGES` requires reason and returns same Revision to DRAFT without changing old Submission or resetting B3 generation;
- qualified reassignment must currently satisfy the same frozen actor rule + enabled User + `approval.act` + SoD;
- no generic delegation/BPM task engine.

### In-product collaboration/viewing

- supported native governed formats must have a safe in-product inspection journey;
- approval-route participants may view exact source and/or exact PDF rendition derived from the same Submission;
- comments/annotations/suggestions are `SubmissionFeedback`, detached from submitted bytes;
- applying a returned suggestion is a later B3 WorkingContent OCC mutation;
- viewer/editor provider states never become Step/domain identity or mechanism permissions.

### Rendition / Release

- `OfficialRepresentationPolicy = SourceOnly | RequireRendition(ContentFormat)`;
- viewer capability is independent: SourceOnly may still have auxiliary PDF for in-product viewing;
- Rendition is immutable exact-Submission → exact output Artifact + generator/build provenance;
- successful Rendition confirmation creates Artifact + typed Rendition ownership in one local semantic transaction;
- ReleasePlan snapshots representation requirement + optional `not_before` per Submission;
- Release is automatic/system-owned; no publish button;
- winning Release is the only effectivity transition and atomically creates ReleaseRecord, makes candidate EFFECTIVE and predecessor SUPERSEDED;
- one semantic Release winner is enforced by Document serialization + unique ReleaseRecord + B3 one-EFFECTIVE backstop.

### Distribution

- live V1 audience currently includes typed Group configuration only where evidenced;
- audience mutations and Release serialize on the same Document root;
- winning Release snapshots concrete Users and inserts DistributionObligation rows in the same commit;
- later Group membership/rename/delete never rewrites historical denominator;
- Distribution never grants access;
- only explicit authenticated acknowledgement by the obligated User completes the obligation;
- view/download/notification/search do not acknowledge.

## 7. Explicit bounded R9.5 Approval refinement

The frozen ledger historically states:

```text
ApprovalPolicy Step purpose = review | approval
```

On 2026-08-18 the operator approved a bounded reopen after directed external research + DevelopmentConexus Method review.

Current R10 working target **removes this discriminator** and uses one governance Step model.

Reason:

- new evidence showed comparable mature systems do not establish a universal `review|approval` kernel taxonomy;
- collaboration, verdict and stronger authentication are orthogonal task capabilities;
- the old distinction primarily encoded legacy editor/UI behavior and therefore failed the Structural Inversion Test;
- keeping it would create accidental complexity and ambiguity beside Controlled Information `PeriodicReview`.

This bounded refinement does not reopen exact Submission binding, actor rules, ANY/ALL, SoD, fresh-auth, return/resubmit, Rendition, Release, Distribution, B1, B2 or B3.

The frozen ledger remains the historical decision record; this section + B4 acceptance record are the explicit current-R10 overlay so fresh sessions must not resurrect `purpose=review|approval` as the working target.

## 8. Storage / records / privacy / migration posture carried forward

- Artifact exact-byte identity/hash remains provider-independent;
- ManagedArtifactStore port/conformance remains first-class; Local dev/test and AWS S3 reference production posture remain;
- production malware inspection remains mandatory/fail-closed before confirming untrusted Artifact bytes;
- no generic Record declaration; governed objects become retention subjects according to frozen rules;
- retention expiry never auto-deletes; LegalHold blocks disposition, not business lifecycle;
- User/data-subject privacy remains separable from immutable governance evidence; surviving Audit skeleton must be PII-minimized/non-PII;
- Backup/Restore, Historical Migration, Governed Subject Export and explicit IMPORT_COPY/PUBLISH_COPY remain; Tenant Portability Export remains deferred;
- no PKI/TSA/HSM/eDiscovery/crypto-erasure platform without a named real trigger.

## 9. Implementation gate

**CLOSED.** B3/B4 are accepted only for continued R10 integration, not final ratification.

Before implementation:

```text
B5
→ B6
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

## 10. Exact next step — R10-B5 Documentary Context + Records Governance + Artifact closure

Open **R10-B5** in the same integrated research-heavy design mode. Consume promoted B1/B2 plus accepted non-final B3/B4; do not reopen them for implementation convenience.

B5 must jointly cover:

```text
DossierType / Dossier
Dossier↔Document contextual relation
EvidenceType / Evidence
primary Dossier + secondary contextual links
Evidence capture immutability + exact Artifact ownership
ExternalReference/provenance where B5-owned

Retention policy selection + RetentionBinding snapshot
DocumentRevision retention unit
Evidence retention unit
RetentionExtension
LegalHold scopes + materialized held subjects
Disposition eligibility + explicit DispositionRecord
no automatic deletion

final typed Artifact ownership/reference closure
preservation across retained/held subjects
no generic owner_type/id registry
no confirmed orphan semantic Artifact
```

Cross-stage review must explicitly attack:

```text
B3 Submission/Artifact ownership
B4 Approval/Rendition/Release evidence as part of DocumentRevision retention unit
Release/effectivity anchors for retention clocks
LegalHold over newly entering subjects
Artifact shared-reference preservation/disposition
Distribution/Acknowledgement retention interaction where material
```

Route later work correctly:

```text
Audit/Interchange/final cross-owner matrix              → B6
physical storage/malware/relocation/restore             → R10-C
async jobs/projections/notifications/provider effects   → R10-D
API/frontend/viewer/editor journeys                     → R10-E
historical migration/cutover/deletion                   → R10-F
```

Implementation remains **BLOCKED**.
