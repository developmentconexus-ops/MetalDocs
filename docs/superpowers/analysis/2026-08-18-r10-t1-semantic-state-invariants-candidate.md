# R10-T1 — Semantic State & Invariants — Integrated Candidate

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE CANDIDATE — OPERATOR REVIEW PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Product authority:** `wiki/architecture/launch-v1-product-contract.md`  
> **GCR authority:** `wiki/architecture/whole-product-alignment-review.md`  
> **Ownership authority:** `wiki/architecture/launch-v1-ownership-topology.md`  
> **Technical routing authority:** `wiki/architecture/r10-technical-architecture.md` after routing promotion  
> **Implementation:** BLOCKED

This is the first technical-descent candidate after the Whole-Product rebaseline. It defines the **minimum enduring semantic facts and invariants** required by the approved Launch V1 Product Contract and 4+1 ownership topology.

It intentionally does **not** define SQL, tables, indexes, Go packages, API routes, storage providers/keys, object-store topology, worker topology, lock order, exact authorization catalog, frontend flows, or migration execution machinery.

Old R10 B1–B6/C material is evidence only. A fact survives here only when the accepted Product Contract, GCR, ownership authority, or a reachable Launch correctness/future-seam failure proves that it deserves to survive.

---

## 1. T1 decision question

> **What semantic facts must MetalDocs be able to remember durably so that Launch truth is unambiguous, immutable history remains truthful, and named future capabilities can attach without duplicating or reinterpreting core authority?**

T1 is successful when every Launch lifecycle statement can point to exactly one semantic owner/fact family and no mechanism is promoted into authority merely to make persistence convenient.

---

## 2. Method classification

### Known

Accepted authority requires:

```text
Authentication != Organization != Authorization
Document != Revision != Working Content != Submission
Revision ordinals never reuse
Submission = immutable exact governed attempt
same-Revision resubmit = new Submission
Template = ordinary governed Document role
NoHumanApproval OR one sequential governance route
feedback binds exact Submission
withdraw attempt != cancel Revision
Release = system-owned effectivity authority
one current EFFECTIVE truth
replacement Release atomically supersedes prior EFFECTIVE truth
optional required official Rendition
OBSOLETE without successor = explicit governed journey
Audit = evidence, never state authority
Search = projection, never state/access authority
native history != imported history
storage/provider identity != semantic identity
governed history preserved in Launch
future capability must attach without duplicating Document/Revision/Submission authority
```

### Reopened from old R10

The following old conclusions are not automatically retained:

```text
standalone Artifact semantic row/owner
DocumentTypeCategory taxonomy
editable Tenant Dictionary/System Value platform
structured TemplateSpec platform
DRAFT EditorialComment platform
mandatory REV002+ reason-for-change
Periodic Review state
Approval as separate semantic owner
ApprovalPolicyVersion as mandatory architecture
ANY/ALL quorum
strict SoD
fresh-auth requirement
due/SLA/reassign/overseer semantics
scheduled ReleasePlan.not_before
SourceOnly auxiliary semantic PDF
Distribution state
Dossier/Evidence/Records state
generic Interchange state
global AuditChainHead/hash-chain
```

A later technical block may reopen one of these only with a concrete accepted Launch consumer/invariant.

### Deferred by technical stage

```text
T2 → exact lifecycle transactions, concurrency, governance participant semantics, locks/atomicity
T3 → concrete Launch roles/permissions/check sites + required-Audit census
T4 → storage/staging/byte validation/malware/restore/GC mechanisms
T5 → async intents/workers/Search projections/notifications/external effects
T6 → API/frontend/viewer/editor journeys
T7 → historical migration plan/cutover/idempotency/incomplete-history execution shape
```

---

## 3. Credible alternatives

### A — subtract future features from the old B1–B6 entity set

Keep the former relational families and delete only obvious Dossier/Evidence/Records/Distribution rows.

**Reject — Local Maximum.** The resulting model still inherits semantic facts created by old abstractions: Artifact ownership, Approval-owner snapshots, policy-version machinery, dictionary/taxonomy adjuncts and disposition-driven provenance choices.

### B — one generic document/workflow record model

Collapse Document/Revision/Submission/governance/history into configurable record/status/event structures.

**Reject — accidental generality.** It destroys the exact distinctions the Product Contract was created to preserve and trends toward ECM/BPM platform design.

### C — small semantic fact set separated by meaning and mutation law

Keep stable business identities, mutable current truth, immutable governed attempts/evidence, explicit effectivity/obsolescence facts and minimal current configuration. Put physical bytes, staging, projections and execution machinery outside semantic authority.

**Recommended Global Maximum.**

---

# 4. Authentication — minimum semantic state

Authentication retains exactly two Launch semantic families unless T2/T3 proves another consumer.

## 4.1 Provider Subject Binding

Durable meaning:

> A trusted authentication-provider subject is bound to exactly one MetalDocs organizational User for the purposes of entering the product.

Minimum semantic facts:

```text
provider trust-domain / issuer identity
provider subject identity
MetalDocs User identity
binding lifecycle/current validity
```

Laws:

- email, username and display name are attributes, never provider-subject identity;
- provider roles/groups/orgs/permissions are never canonical product Authorization;
- changing IdP/provider mechanics must not rewrite User identity or historical document/governance actor identity;
- a provider binding may stop being valid without deleting the User whose historical attribution must remain.

Exact rebind/provisioning mechanics are not T1.

## 4.2 Application Session

Durable meaning:

> MetalDocs may terminate future product access independently of provider-side session lifetime.

Minimum semantic facts:

```text
session identity
bound authenticated User/binding
session validity/revocation state
essential created/expiry/revocation instants
```

Session is not authorization evidence and never carries durable Role/Permission authority.

## 4.3 Authentication assurance / fresh-auth

**DEFER SAFELY in T1.**

The approved Product Contract does not require fresh authentication for any Launch decision. The old B4 option therefore creates no Launch semantic fact yet.

Seam preserved:

- Authentication remains the future owner if T2/T3 later proves a named fresh-auth/e-signature consumer;
- Controlled Documents may consume bounded assurance evidence without ever storing credentials/password challenges.

---

# 5. Organization — minimum semantic state

## 5.1 Company root

Launch has one stable company identity per deployment.

Minimum semantic meaning:

```text
Company identity
current human-readable company identity/settings required by product operation
```

Laws:

- exactly one company root is valid in a Launch deployment;
- company identity is not a universal row-partition key by reflex;
- provider/Keycloak organization identity never becomes this authority;
- the stable company identity is deliberately preserved as the future attachment anchor if pooled multi-customer tenancy is later promoted.

T1 does not prescribe a universal `tenant_id` column or pooled-tenancy schema.

## 5.2 User

`User` is the stable organizational participant identity.

Minimum semantic facts:

```text
stable User identity
current organizational eligibility: enabled/disabled
```

Laws:

- offboarding disables future actions/access without deleting or rewriting historical attribution;
- provider identity, email and username are not User identity;
- disable/re-enable does not create a new historical actor identity.

## 5.3 User Profile

Current human-readable/contact enrichment is separately erasable from stable User identity.

Minimum semantic facts may include only product-needed current display/contact data.

Law:

> Erasing profile enrichment must not rewrite Submission decisions, Release/obsolescence history or Audit actor identity.

## 5.4 Area

Area is a stable organizational scope/reference used by Documents and scoped Authorization.

Minimum semantic facts:

```text
stable Area identity
stable code if numbering/display relies on it
current name
current enabled/retired eligibility
```

No Area hierarchy, default approver, dynamic metadata or generic taxonomy exists without a consumer.

Existing governed history may continue referencing a retired Area; retirement only blocks newly admitted relationships where the owning operation requires an active Area.

## 5.5 Group and membership

Group is a flat organizational grouping with current membership truth.

Minimum semantic facts:

```text
stable Group identity
current name
current User↔Group membership
```

No nested groups, dynamic rules, IdP-group mirroring or membership-history domain exists in Launch.

Audit may prove membership transitions later; it does not become current membership authority.

## 5.6 User↔Area organizational membership

**Not admitted as a T1 semantic fact yet.**

Area-scoped Authorization can assign grants directly to User/Group at an Area scope without requiring a separate organizational `UserAreaMembership` concept. If T2/T3 proves that reviewer eligibility or another accepted journey requires organizational Area membership independent of grants, reopen this one relationship deliberately.

---

# 6. Authorization — minimum semantic state

Authorization has one durable current-truth family plus product vocabulary.

## 6.1 Product Role / Permission vocabulary

Role and Permission meanings belong to the product, but T1 does not require editable deployment rows or preserve the old 5×43 catalog.

Current decision:

```text
Role semantics       → product-owned vocabulary
Permission semantics → product-owned vocabulary
Role→Permission      → product-owned bundle definition
```

Whether these definitions are code/static configuration versus persisted immutable catalog is a T3 realization decision. They are **not** customer-configurable custom-role/custom-permission platform facts in Launch.

## 6.2 Role Assignment

RoleAssignment is the durable product grant truth.

Minimum semantics:

```text
subject = User | Group
role
scope = Company | Area
current grant existence
```

Laws:

- additive grants + default deny;
- removal means the grant is no longer current; history is reconstructed from Audit/domain evidence where required, not from interval/tombstone grant rows unless T3 proves a consumer;
- `tenant_owner`/admin-style roles are grant bundles, never domain-governance bypass;
- Authorization answers grant/scope questions; Controlled Documents still owns resource relationship/state predicates.

Exact roles, permissions, allowed scope matrix and administration law belong to T3.

---

# 7. Controlled Documents — minimum semantic state

This is the Launch product kernel. The goal is not minimum entity count; it is minimum **independent meaning**.

## 7.1 Document Type

Document Type owns current controlled-document configuration needed to create/govern future work.

Minimum semantic facts:

```text
stable Document Type identity/code
current display/status eligibility
current numbering rule
current template eligibility relation
current governance mode/configuration
current official-representation requirement
```

Required governance mode vocabulary:

```text
NoHumanApproval
UseGovernanceRoute
```

Required representation posture:

```text
SourceOnly
or
RequireOfficialRendition(format/representation requirement)
```

Laws:

- changing current Document Type configuration never rewrites existing Submission/governance/Release history;
- in-flight governed attempts freeze all decision-relevant configuration needed to interpret their history;
- an inactive type may block new Documents without invalidating existing governed history.

Not T1 facts:

```text
DocumentTypeCategory taxonomy
editable generic metadata dictionary
custom forms/rules
```

## 7.2 Number allocation

Two meanings must remain distinct:

```text
numbering rule/configuration = semantic Document Type truth
allocation counter/locking   = durable mechanism
```

Document owns the final allocated stable official code. Committed codes are never reused/rebound.

T1 does not require a first-class `DocumentNumberSeries` semantic owner. T2 may require durable monotonic allocation state as mechanism to prove uniqueness/concurrency.

## 7.3 Document

Document is the stable official identity across all business revisions.

Minimum semantic facts:

```text
stable Document identity
stable official code
Document Type
Area/responsibility context required by accepted search/access journeys
current responsible owner when product uses that relationship
ordinary Document vs Template role
immutable creation provenance needed by the product
```

Laws:

- Document is not a file;
- Document is not replaced when content changes;
- Document must not duplicate current Revision/effectivity authority through a second independent status truth;
- future Dossier, Records, Training, Change Control and repository capabilities attach by reference to stable Document identity rather than copying it.

### Document Area

Product search/access explicitly uses Area, so the Document→Area relationship is admitted.

### Responsible owner

Product personas/search explicitly reference responsible owner; current responsibility is admitted as a mutable Document relationship. Historical governance actions continue pointing to the actor identities that actually acted and are never rewritten when responsibility changes.

## 7.4 Business Revision

Revision is one business change cycle.

Minimum semantic facts:

```text
stable Revision identity
owning Document
business ordinal
current lifecycle state
creation attribution/time
native vs imported provenance classification sufficient to prevent fabricated native history
```

Lifecycle vocabulary:

```text
DRAFT
SUBMITTED
EFFECTIVE
SUPERSEDED
OBSOLETE
CANCELLED
```

Laws:

- revision ordinals never reuse;
- at most one open business Revision (`DRAFT|SUBMITTED`) per Document;
- at most one EFFECTIVE Revision per Document;
- zero open is valid;
- zero EFFECTIVE is valid before first Release, after governed obsolescence, or where imported truth legitimately says so;
- a newer open Revision may coexist with the prior EFFECTIVE Revision;
- state transitions never imply physical byte deletion.

### Lifecycle evidence versus current state

Current Revision state is canonical current lifecycle truth. Immutable transition evidence is held by the fact that caused the terminal/effectivity transition:

```text
Submission/governance evidence → submitted/returned/withdrawn journey
RevisionCancellation          → CANCELLED reason/actor/time
Release                       → EFFECTIVE and predecessor SUPERSEDED
Obsolescence                  → OBSOLETE reason/governance/time
```

This avoids using Audit as lifecycle authority and avoids a second generic event-store authority.

## 7.5 Working Content

Working Content is the sole mutable authoring truth for an open DRAFT Revision.

Minimum semantic facts:

```text
owning Revision
current governed source/content identity facts
current bounded governed metadata required by the document itself
monotonic working generation/version for concurrency
last material author/update attribution needed by product operation
```

Laws:

- mutable only while Revision is DRAFT;
- autosave/recovery never consumes a Revision number and never creates official governed history;
- browser/editor/provider state is not authority;
- Working Content may survive temporarily across SUBMITTED to support truthful return/withdraw-to-DRAFT, but during SUBMITTED the immutable Submission is governance truth;
- a future CRDT/coauthoring implementation may replace DRAFT concurrency mechanics without changing Revision or Submission meaning.

Not admitted without a Launch consumer:

```text
semantic WorkingSnapshot/checkpoint history
EditorSession as business truth
DRAFT EditorialComment platform
structured TemplateSpec platform
mandatory REV002+ reason-for-change rule
Tenant Dictionary snapshot platform
```

Recovery snapshots/editor leases may exist later as mechanisms if T2/T4/T5 prove need.

## 7.6 Submission

Submission is an immutable exact governed attempt for one Revision.

Minimum semantic facts:

```text
stable Submission identity
owning Revision
accepted Working Content generation
exact submitted source-content identity facts
all decision-relevant governed metadata/provenance frozen for this attempt
submitted actor/time
frozen governance requirement/config snapshot
frozen official-representation requirement
```

Exact submitted source-content identity includes provider-neutral properties sufficient to prove sameness/difference of governed content, such as:

```text
cryptographic digest
size
format/media semantics as required
```

The physical locator/handle used to retrieve bytes is a T4 mechanism, not Submission identity.

Laws:

- immutable after creation;
- same Revision may have multiple Submissions;
- separate legitimate attempts remain separate identities even if exact bytes are equal;
- RETURN or WITHDRAW never mutates an older Submission;
- every governance decision/feedback binds the exact Submission it concerned;
- later DocumentType/policy edits do not reinterpret the frozen attempt.

T1 does not require the old generic JSON/JCS manifest shape or global `Artifact` reference. T4 may choose a deterministic canonical descriptor/digest proof mechanism, but the semantic requirement is exact frozen content + decision-relevant state.

## 7.7 Template semantics

Template remains an ordinary governed Document role.

Minimum semantic facts:

```text
Document has/does not have Template role
current eligibility: which target Document Types may use it
immutable origin provenance on a derived Document
```

Create-from-template must pin the exact source that was current EFFECTIVE at creation.

Recommended semantic origin anchor:

```text
source Template Document
source exact EFFECTIVE Revision
source exact-content identity snapshot
creation time
```

Do **not** require a strong dependency on a native Submission: imported truthful history may have an EFFECTIVE Revision without fabricated native Submission history.

Later template changes never rebind or rewrite the derived Document.

## 7.8 Governance Route — current configuration

Controlled Documents owns a small sequential route definition for document governance.

Minimum current configuration semantics:

```text
stable route/policy identity only if reusable configuration needs it
ordered Step definitions
business labels
participant-selection rule sufficient for the accepted Launch route
```

No structural `review|approval` Step type exists.

### Versioned policy object is not mandatory in T1

The actual invariant is:

> every governed attempt freezes the exact route/configuration it must satisfy.

Therefore T1 does **not** require the old mandatory `ApprovalPolicyVersion` family. A mutable current route plus immutable per-attempt snapshot is sufficient unless T2/T6 proves a product journey that needs independently browsable/versioned reusable policy history.

## 7.9 Governance Attempt — bounded shared semantic

The Product Contract has two proven governance consumers:

```text
Submission governance
Obsolescence governance
```

Both need a stable execution identity independent of the subject they govern, because attempts may be active, completed or terminated while the governed subject remains historical truth.

T1 therefore admits one internal semantic concept:

```text
GovernanceAttempt
```

with a **closed subject universe**:

```text
SUBMISSION
OBSOLESCENCE
```

This is not a generic workflow/BPM subject registry. Adding another subject kind is a material reopen requiring a named product capability.

Minimum immutable/frozen semantics per attempt:

```text
attempt identity
exact governed subject identity
frozen governance mode/route steps relevant to the attempt
started actor/system attribution and time
attempt lifecycle/outcome sufficient to distinguish active/completed/withdrawn/returned/terminated truth
```

T2 decides the smallest exact state vocabulary separately for Submission governance and Obsolescence governance; T1 does not force all operations to be legal for both subject kinds.

### Why attempt identity is distinct from Submission

Submission is immutable candidate identity. GovernanceAttempt is the human/system governance execution over a candidate or obsolescence request. Conflating them prevents the required second obsolescence journey and encourages Submission-specific workflow assumptions.

## 7.10 Governance Step activation/decision evidence

A route Step is one ordered governance gate.

T1 admits immutable decision evidence sufficient to answer:

```text
which attempt/Step?
which actor actually decided?
what outcome?
when?
what exact governed subject did the attempt bind?
```

Normal human Submission outcomes remain:

```text
ACCEPT
RETURN_FOR_CHANGES
```

T1 intentionally does **not** admit as baseline semantic requirements:

```text
ANY|ALL quorum
strict creator/submitter SoD
cross-Step same-user SoD
fresh-auth
SLA/due date
overseer/reassign model
complex actor activation pool
```

T2 may promote only the dimensions needed to make the named Launch journeys correct.

## 7.11 Submission feedback

Feedback is immutable/detached evidence bound to an exact Submission/governance attempt.

Minimum semantics:

```text
submission/attempt/Step context
actor
time
feedback body or bounded product-supported payload
```

Feedback never mutates Submission or Working Content. Applying feedback occurs only after the Revision lawfully returns to DRAFT and produces later Working Content / a later Submission.

No generic threaded annotation/suggestion platform is admitted in T1.

## 7.12 Revision cancellation

Cancellation is not governance rejection and not Submission withdrawal.

T1 admits an immutable cancellation fact:

```text
cancelled Revision
actor
time
mandatory explicit reason
```

Law:

- cancellation terminates the business change cycle;
- an older EFFECTIVE Revision remains EFFECTIVE;
- prior Submissions/decisions remain history.

Exact eligibility/atomic transition is T2.

## 7.13 Release

Release is the sole normal native effectivity authority.

T1 admits an immutable Release fact:

```text
released Revision
winning exact Submission
system-owned release identity/time
prior EFFECTIVE Revision identity when replacement occurred
required official Rendition identity when representation policy required one
```

Laws:

- no human “publish latest file” semantic exists;
- first Release establishes first EFFECTIVE Revision;
- replacement Release is one business transition that makes the successor EFFECTIVE and predecessor SUPERSEDED;
- Release cannot point to content different from the governed winning Submission;
- Distribution is not part of Launch-Core Release semantic/atomicity.

No scheduled `not_before` fact is admitted without a Launch consumer.

## 7.14 Official Rendition

Rendition is admitted only when a Document Type/Submission froze a requirement for one official derived representation.

Minimum semantic facts:

```text
exact source Submission
required output representation/format
exact immutable derived-content identity facts
generator/provenance facts necessary to prove it corresponds to that Submission
successful creation time
```

Laws:

- Rendition never replaces or mutates the governed source Submission;
- a SourceOnly document has no semantic Rendition merely because UI preview/rendering exists;
- preview/view conversions remain T4/T5/T6 mechanisms unless they are the required official representation.

## 7.15 Governed obsolescence without replacement

Obsolescence is an explicit second governance journey over the current EFFECTIVE Revision.

T1 admits an immutable obsolescence intent/result fact family sufficient to preserve:

```text
exact target Document/current EFFECTIVE Revision
mandatory reason
initiator/time
governance attempt identity/frozen route
successful completion actor/decision evidence through GovernanceAttempt
actual obsoleted time
```

Laws:

- until governance succeeds, the target Revision remains EFFECTIVE;
- success changes that exact Revision to OBSOLETE and leaves no EFFECTIVE Revision;
- no raw status toggle is valid;
- an unresolved open replacement Revision prevents completion as required by the Product Contract;
- reactivation of OBSOLETE is absent from Launch.

### NoHumanApproval for obsolescence

**UNKNOWN / T2 decision.**

The Product Contract says obsolescence is governed but does not explicitly state whether a Document Type configured `NoHumanApproval` may obsolete without human decision. T1 preserves the same governance configuration seam but does not silently answer this product-semantic question. T2 must surface it for operator adjudication if the current authorities do not determine it.

## 7.16 Current-effective truth

T1 deliberately rejects a second independent `Document.current_revision/current_status` semantic authority.

Canonical meaning:

```text
current EFFECTIVE Revision
= the unique Revision whose canonical lifecycle state is EFFECTIVE,
  established by Release and removable only by replacement Release or governed obsolescence.
```

A physical pointer/cache may be introduced later only as a structurally synchronized optimization, never a second authority.

Search indexes remain projections and must re-resolve canonical state before serving current official content.

---

# 8. Imported/native history seam

Historical Migration is T7 execution, but the target must be able to represent imported truth without fabrication.

T1 preserves these semantic requirements:

```text
native vs imported history distinguishable
reliable imported Revision identity/ordinal/state/provenance may be preserved
exact imported content, when available, belongs to the imported Revision semantic history
source actor/time/label may remain source provenance rather than native User action
unknown remains NULL/unknown, never synthesized
next native Revision ordinal must remain above every reliable real historical ordinal
```

T1 does **not** yet require old B6 families such as:

```text
RevisionOrdinalReservation
RevisionImportedContent
RevisionImportedGovernanceSnapshot
HistoricalSourceBinding
```

Those are possible T7 realizations. T7 may introduce the smallest target-owned imported-history fact shape needed by actual migration evidence, but it may not fabricate native Submission/Release/Approval/User actions.

Important compatibility rule:

> Template origin and future contexts should prefer stable Document/Revision + exact-content provenance anchors over “must have native Submission” assumptions, because truthful imported history may not have one.

---

# 9. Audit — minimum semantic state

Audit remains an independent supporting semantic owner.

## 9.1 Audit Event

Minimum meaning:

```text
immutable event identity
trusted occurrence time
actor kind + stable User identity or bounded System actor
operation vocabulary
resource kind + stable resource identity
bounded PII-minimized facts necessary to prove the action
correlation identity when materially useful
```

Laws:

- append-only;
- no UPDATE/DELETE through ordinary serving path once implementation exists;
- human-readable profile data, passwords, tokens, provider claims, request bodies and free-form domain reasons do not get copied into indefinite Audit by convenience;
- Audit references owning-domain evidence when a reason/comment exists rather than becoming its duplicate authority;
- Audit never reconstructs current Revision/AuthZ/Organization state by “latest event wins.”

## 9.2 Global cryptographic chain

**DEFER SAFELY.**

No accepted Launch assurance requirement currently justifies `AuditChainHead`, deployment-wide serialization or mandatory cryptographic event chaining. Same-local-commit append-only Audit remains required where T3 marks an operation as critical/governed.

Future tamper-evidence/non-repudiation requirements may reopen this with concrete assurance evidence.

---

# 10. Minimum semantic family set

This list is conceptual. It does **not** mean one table/type/package per line.

```text
Authentication
  ProviderSubjectBinding
  ApplicationSession

Organization
  Company
  User
  UserProfile
  Area
  Group
  GroupMembership

Authorization
  product Role vocabulary
  product Permission vocabulary
  RoleAssignment

Controlled Documents
  DocumentType
  DocumentType↔Template eligibility configuration
  Document
  Document Template role
  DocumentOrigin
  Revision
  WorkingContent
  Submission
  current GovernanceRoute configuration
  frozen GovernanceAttempt
  governance Step/Decision evidence
  SubmissionFeedback
  RevisionCancellation
  Release
  OfficialRendition              // only when required
  Obsolescence governance/result
  imported/native provenance seam

Audit
  AuditEvent
```

Explicitly **not** in the T1 Launch semantic family set:

```text
Artifact
DocumentTypeCategory
TenantDictionaryValue/System Value platform
TemplateSpec
EditorialComment platform
PeriodicReviewPolicy/Record
Distribution obligation/acknowledgement
Dossier/Evidence
Retention/Hold/Disposition
Interchange
Governed Export
Repository connection/receipt
AuditChainHead
WorkingSnapshot business history
EditorSession business authority
```

---

# 11. Cross-family invariants

These are T1 semantic laws; T2–T4 choose enforcement.

1. Provider identity, organizational User identity and Authorization grant are distinct authorities.
2. User offboarding removes future eligibility without rewriting immutable historical actor references.
3. Document code/identity remains stable across all Revisions.
4. Revision ordinal is business history and never reused.
5. Working Content is the only mutable current DRAFT content authority.
6. Submission freezes one exact accepted Working Content generation and decision-relevant state.
7. Submission is immutable; resubmit creates a new Submission under the same Revision.
8. Governance attempt is execution truth over an exact closed-set subject, never content identity.
9. Governance feedback/decisions cannot mutate Submission.
10. Cancellation ends a Revision change cycle; withdrawal ends an active Submission governance attempt; neither rewrites prior Submission history.
11. Release is the only normal native operation establishing EFFECTIVE.
12. Replacement Release changes predecessor SUPERSEDED + successor EFFECTIVE as one business transition.
13. At most one Revision per Document is EFFECTIVE.
14. Obsolescence is governed; success targets the current EFFECTIVE Revision and leaves no EFFECTIVE Revision.
15. Required official Rendition binds the exact winning Submission; preview rendering does not become a semantic Rendition.
16. Search/index state cannot establish effectivity or access.
17. Audit proves that actions occurred but never owns current domain state.
18. Storage/provider handle/key/location cannot become Document/Revision/Submission/Rendition identity.
19. Imported history never fabricates native actor/approval/release facts.
20. Future contexts attach to stable existing identities rather than copy core authority.

---

# 12. Named-future compatibility test

| Future capability | T1 attachment anchor | T1 protection |
|---|---|---|
| Distribution / Read & Acknowledge | Release + exact effective Revision + User/Group | no Distribution state inside Release; effectivity remains Controlled Documents |
| Periodic Review | stable Document + exact current EFFECTIVE Revision | no review state folded into Revision lifecycle; later review cannot become effectivity |
| Dossier | stable Document identity | Dossier can reference without owning content/access |
| Evidence | Organization/AuthZ + shared exact-content mechanism | Evidence may gain independent lifecycle rather than being forced into Document Revision |
| Records / Hold / Disposition | stable Document/Revision/Submission/Release identities + immutable history | no storage/provider authority or disposition-driven rewrite in Launch core |
| Governed Export | stable semantic relationships + exact-content identities | export remains consumer/package, not source authority |
| External repository connectors | target-owner import/publish seams + exact-content descriptors | repository object IDs cannot become MetalDocs IDs |
| Training/LMS | Release/effective Document + future Distribution | training competence remains separate from effectivity |
| Generic Change Control | stable Document/Revision + explicit lifecycle seams | orchestration can reference core without taking its authority |
| pooled tenancy | stable Company identity + reopenable substrate | no premature universal partitioning; company anchor survives |
| realtime CRDT | Working Content boundary | replace concurrency mechanism without changing Revision/Submission identity |

This is the binding balance:

```text
known future = architecture counterexample
known future != dormant Launch implementation
```

---

# 13. Adversarial challenge

## Attack 1 — “GovernanceAttempt is the generic workflow abstraction we just rejected.”

Counter-test: its subject universe is closed to `SUBMISSION|OBSOLESCENCE`, both owned by Controlled Documents and both explicitly required by Launch. It owns no arbitrary workflow designer, branching, plugin tasks, cross-domain subject registry or generic orchestration. If another subject appears, the concept reopens rather than silently extending.

**Result:** bounded common semantic survives.

## Attack 2 — “Removing Artifact makes exact content impossible to reference consistently.”

Exact content still has semantic identity facts on Submission/Rendition/Working Content/imported Revision history. T4 must design a shared storage/integrity mechanism that can locate/verify bytes without owning their meaning. A generic Artifact semantic row is not required for shared physical storage.

**Result:** Artifact owner remains removed; T4 receives proof obligation.

## Attack 3 — “Without policy versions, changing a route destroys history.”

No: each governed attempt freezes the exact route/configuration it must satisfy. Independent versioned reusable-policy history is only necessary if a product journey needs to browse/reuse policy versions as first-class objects.

**Result:** per-attempt snapshot is essential; mandatory PolicyVersion aggregate is not yet proven.

## Attack 4 — “One Revision.state duplicates Release/Obsolescence evidence.”

They have different mutation laws: Revision state is current lifecycle truth; Release/Obsolescence/Cancellation are immutable evidence/causal facts that explain terminal/effectivity transitions. Audit cannot substitute for those domain facts.

**Result:** both current state and owning-domain transition evidence are justified, provided T2 enforces one transition atomically.

## Attack 5 — “Future Records will need retention roots, so keep Artifact/retention fields now.”

The future horizon only proves stable governed-subject identities/history must survive. Records can later attach new policy/hold/disposition state by reference. Prebuilding retention roots would reintroduce the exact backward pressure the GCR removed.

**Result:** seam preserved, dormant Records state rejected.

## Attack 6 — “Imported history needs the old B6 imported tables now.”

T1 must preserve the semantic distinction and non-fabrication law, but the exact shape depends on actual migration source completeness. T7 is the proper stage to choose full imported Revision, partial historical evidence, ordinal reservation or another bounded form.

**Result:** semantic seam now; realization deferred to T7.

---

# 14. T1 outcome candidate

Method outcome:

```text
CURRENT STRUCTURE CONFIRMED:
  AuthN / Organization / AuthZ semantic separation
  Document != Revision != WorkingContent != Submission
  Template as ordinary Document role
  one sequential Step semantic
  system-owned Release
  Audit separate from domain truth

RESTRUCTURE NOW:
  exact-content facts move onto semantic content-owning/frozen records
  Approval execution folds into Controlled Documents as bounded GovernanceAttempt semantics
  current-effective truth is not duplicated at Document/Search layer
  old mandatory versioned Approval-policy machinery is not automatic

DEFER SAFELY:
  fresh-auth/SoD/quorum/SLA/reassign richness until T2/T3 proves it
  physical content/storage/restore mechanics → T4
  imported-history persistence shape → T7
  global Audit cryptographic chain absent assurance requirement
  Launch+/Future capability state
```

---

# 15. T1 explicit non-decisions

T1 does **not** decide:

```text
SQL/table/index/check/trigger layout
one-table-vs-multiple-table persistence shape
Go package/module layout
aggregate roots
lock order/isolation level
exact Step participant selector
ANY vs ALL
SoD rules
fresh-auth
reassignment/oversight
exact governance-attempt state enum
exact NoHumanApproval behavior for obsolescence
exact role/permission catalog
same-commit Audit census
storage handle/reference design
hash/canonicalization encoding
malware/scan policy
object-store/provider topology
async/outbox model
Search index technology
API/frontend routes
migration execution/idempotency/partial-history model
```

These are deliberately routed to T2–T7.

---

# 16. T1 gate / operator adjudication packet

Recommended operator dispositions:

```text
T1-A ACCEPT — minimum semantic family set in §10.
T1-B ACCEPT — GovernanceAttempt as bounded Controlled-Documents semantic with closed subjects SUBMISSION|OBSOLESCENCE; no generic BPM.
T1-C ACCEPT — current route config + immutable per-attempt snapshot is required; standalone mandatory ApprovalPolicyVersion family is not required unless later consumer proves it.
T1-D ACCEPT — exact-content identity facts belong on WorkingContent/Submission/Rendition/imported Revision history; physical locator remains T4 mechanism.
T1-E ACCEPT — Document has no independent current-revision/effectivity status authority; canonical current-effective truth comes from Revision lifecycle established by Release/Obsolescence.
T1-F ACCEPT — TemplateOrigin anchors source Template Document + exact effective Revision/content provenance, not mandatory native Submission.
T1-G ACCEPT — imported/native semantic distinction is required now; exact incomplete-history persistence shape stays T7.
T1-H ACCEPT — AuditEvent remains semantic supporting evidence; global AuditChainHead/hash-chain remains deferred.
T1-I ACCEPT — old taxonomy/dictionary/TemplateSpec/DRAFT-comments/PeriodicReview/Distribution/Records/Interchange families are absent from Launch T1 unless later stages produce a named Launch consumer.
T1-J OPEN PRODUCT-SEMANTIC QUESTION FOR T2 — whether NoHumanApproval may govern obsolescence with zero human gate or obsolescence always requires at least one human Step. Do not infer silently.
```

If T1-A through T1-I are accepted, T1 closes except the bounded T1-J question routed into T2. T2 then derives lifecycle transactions/concurrency/governance participant semantics from this semantic fact set.

Implementation remains **BLOCKED**.