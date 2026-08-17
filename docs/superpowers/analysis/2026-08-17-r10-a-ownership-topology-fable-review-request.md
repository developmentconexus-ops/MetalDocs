# R10-A Ownership Topology — Independent Fable Review Request

> **Status:** CANDIDATE / INDEPENDENT REVIEW REQUEST — **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Candidate baseline before this packet:** `ab0aba8a60ab3069cfa3b0187176187fe1e6b220`
> **Stage:** R10 — Technical Architecture / R10-A Ownership Topology
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** **CLOSED — design/documentation only.**
> **Authority note:** this packet is review/staging evidence only. It does not amend R9.5 or become target authority by existing. Accepted outcomes must be adjudicated against the Method and promoted deliberately into the active redesign authority/ledger.

---

## 0. Reviewer bootstrap — Fable cold start

**Reviewer role:** senior independent adversarial reviewer. Reconstruct the project from the repository. Do not use prior conversation memory, author explanations outside the repo, historical implementation shape, or this packet as requirement authority.

Read in the order required by `AGENTS.md`:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. this packet
7. current runtime/module/schema/OpenAPI/frontend evidence only when necessary to falsify or validate a specific technical claim

When auditing the R9.5 freeze, the review evidence is:

- `docs/superpowers/analysis/2026-08-17-r9.5-8-whole-product-adversarial-freeze.md`
- `docs/superpowers/analysis/2026-08-17-r9.5-8-independent-adversarial-challenge.md`

The R9.5 review artifacts are evidence, not parallel target authority.

### Required posture

Apply the DevelopmentConexus Engineering Method directly:

```text
Evidence
→ Known / Inferred / Unknown / Deferred
→ Root Cause
→ Target Invariant
→ Constraints
→ Credible Alternatives
→ Local Maximum vs Global Maximum
→ Essential vs Accidental Complexity
→ YAGNI / Overengineering / Future Cost
→ Authority / Boundary
→ Enforcement
→ Proof Strategy
→ Adversarial Challenge
→ Decision
→ Reopen Triggers
```

Apply the Structural Inversion Test aggressively:

> If the current implementation were opposite in every relevant respect, would the target ownership conclusion still be correct?

Do **not** reward the candidate for resembling the current source tree. Do **not** penalize it merely because migration would be large. The target is the smallest sustainable architecture for the frozen product semantics.

A reviewer finding is evidence, not requirement authority. Do not introduce new product requirements disguised as review fixes.

---

## 1. Verified stage gate

The candidate assumes and must not casually reopen:

```text
R3–R9   = LOCKED
R9.5-1  = LOCKED
R9.5-2  = LOCKED
R9.5-3  = LOCKED (refined by R9.5-8)
R9.5-4  = LOCKED
R9.5-5  = LOCKED (refined by R9.5-8)
R9.5-6  = LOCKED
R9.5-7  = LOCKED
R9.5-8  = CLOSED / APPROVED
R9.5    = FROZEN
R10     = DESIGN ONLY
implementation = BLOCKED
```

R9.5 independent review ended with `APPROVE / FREEZE R9.5` and an empty reopen set.

If this review believes R9.5 must reopen, it MUST provide exactly:

1. material new evidence;
2. frozen invariant/authority invalidated;
3. why the existing seam cannot solve the issue;
4. smallest decision set that must reopen;
5. everything that remains frozen.

Implementation inconvenience, package preference, table convenience, provider capability, current code shape, or a cleaner local implementation is not a reopen reason.

---

# 2. R10 integrated context

R10 was decomposed before microdecisions into six blocks:

```text
R10-A  Ownership Topology & Dependency DAG
R10-B  Transactional Domain State & DB Invariants
R10-C  Artifact / Records Physical Integrity
R10-D  Durable Async / Projections / External Effects
R10-E  Canonical Access / API / Frontend Journeys
R10-F  Historical Migration / Cutover / Final Deletion
```

Candidate closure order:

```text
R10-A → R10-B → R10-C → R10-D → R10-E → R10-F
```

R10-A must decide ownership before schema/endpoints because table ownership, transaction ownership, event producers, API ownership and migration destinations are downstream consequences of semantic ownership.

The review should challenge that decomposition only if a material failure class is split incorrectly or an owner decision cannot be judged without prematurely deciding a later mechanism.

---

# 3. Root cause R10-A is intended to solve

The candidate identifies the structural defect as:

> Incremental evolution allowed deployment/module boundaries, providers and historical nouns to acquire architectural authority that no longer matches the frozen product/domain model.

A local maximum would rename/reorganize the current modules and preserve their conceptual cuts.

The candidate Global Maximum is:

> Derive one semantic owner for every frozen business fact directly from R3–R9.5, classify replaceable mechanisms separately, publish an acyclic dependency direction, then make later data/transaction/API/migration design descend from those owners.

Review whether this is truly the root cause, or whether the candidate is merely moving accidental complexity into a new taxonomy.

---

# 4. Candidate R10-A target ownership

## 4.1 Candidate business bounded contexts

### A. Authentication

Owns:

```text
local credentials / identity binding
activation
opaque sessions
lockout / revocation
fresh-auth / reauthentication assurance
```

Does not own Authorization grants or Organization membership. External IdP remains a future adapter seam.

### B. Organization

Owns:

```text
Tenant
Area
User
Group
GroupMembership
Tenant lifecycle state: ACTIVE | SUSPENDED | ERASED
```

Does not own credentials or permission evaluation.

### C. Authorization

Owns:

```text
Permission
Role
RoleAssignment
subject = User | Group
scope = Tenant | Area
canonical grant evaluation
canonical resource/case relationship filtering contract
```

Frozen equation remains:

```text
Permission
+ required resource/case relationship
+ Domain Governance constraints
= ALLOW
```

No role bypass, generic ACL/ReBAC graph, explicit-deny engine or provider permissions V1.

### D. Controlled Information

Owns the governed Document/Revision lifecycle:

```text
DocumentType
Document
DocumentRevision
WorkingContent
WorkingSnapshot technical checkpoints
RevisionSubmission
Template designation / TemplateUse
numbering / revision labels
EditorialComment
PeriodicReviewRecord
Rendition business semantics
ReleasePlan / ReleaseRecord / effectivity
```

Negative boundary — deliberately does **not** own:

```text
Approval policy/decision semantics
Evidence / Dossier
Retention / LegalHold / Disposition
Distribution acknowledgement
Audit timeline
Historical Migration / portability / external publication
Search / notifications
storage-provider semantics
generic workflow
```

The candidate keeps Rendition and Release semantics here because they are defined only in relation to an exact `RevisionSubmission` and DocumentRevision effectivity. Rendering/execution mechanisms stay outside the domain owner.

### E. Approval

Owns:

```text
ApprovalPolicy(version)
ApprovalStep
ApprovalInstance
activated participant snapshots
ApprovalDecision
reassignment / cancellation / oversight
strict SoD
fresh-auth requirement at decision time
```

Approval binds a RevisionSubmission but does not own Document lifecycle/effectivity.

### F. Documentary Context

Owns:

```text
EvidenceType
Evidence
DossierType
Dossier
ExternalReference
Dossier↔Document contextual links
Evidence secondary context links
```

Dossier/context links never grant access. Evidence CAPTURED uses exactly one immutable primary Dossier and its scope.

### G. Records Governance

Owns:

```text
RetentionBinding / snapped policy
RetentionExtension
LegalHold
materialized held-subject relationship
prospective hold materialization
Disposition eligibility / authorization / completion truth
```

Does not own the underlying Document/Evidence lifecycle or physical byte identity.

### H. Distribution

Owns:

```text
released-document distribution obligation
audience snapshot
AcknowledgementRecord
coverage / completion semantics
```

It does not grant access and notification read/view/download is never acknowledgement.

---

## 4.2 Candidate supporting semantic owners

These own durable meaning but are not proposed as peer business bounded contexts.

### I. Artifact

Owns technical exact-byte identity and physical-content truth:

```text
Artifact identity
canonical SHA-256
size / ContentFormat / media type
technical provenance
staging → validation → confirmation state
managed physical location facts
relocation verification/cutover facts
restore integrity facts
```

Negative boundary:

- not a user-facing document/file library;
- no independent business lifecycle;
- no confirmed orphan Artifact;
- storage provider key/version/URL is never business identity;
- retention meaning derives from governed subjects, not Artifact itself.

Reason for separate supporting owner: both Controlled Information and Documentary Context/Evidence require exact immutable bytes without one business domain becoming authority over the other's content.

### J. Dictionary

Candidate owner for the small value-provider surface:

```text
tenant dictionary values
bounded product/system value catalog
snapshot/resolution provider contract
```

Values that become decision-relevant are snapshotted by the consuming governed lifecycle; later Dictionary mutation does not rewrite history.

The reviewer MUST challenge whether tenant dictionary and product/system catalog truly belong behind one owner or whether this combines two meanings merely because both produce values.

### K. Audit

Owns the transversal append-only/tamper-evident `AuditEvent` timeline and audit export/query semantics.

Domain evidence records remain authoritative for their own facts; Audit must never become a second business-state source.

The reviewer MUST challenge the distinction between `Audit` as a supporting semantic owner and audit-intent durability as transaction/event mechanism, especially because frozen authority requires critical governed mutation not to report success without durable audit intent/event in the same commit boundary.

### L. Interchange

Candidate narrow supporting owner for transfer-boundary process truth:

```text
Historical Migration batch / plan / dry-run / reconciliation
Tenant Portability Export package process
Governed Subject Export package process
External Repository IMPORT_COPY / PUBLISH_COPY process truth
source provenance / transfer attempt / reconciliation identity
```

Negative boundary:

```text
NOT ESB
NOT generic ETL
NOT workflow
NOT connector platform
NOT arbitrary transformation engine
NOT external system master-data authority
```

Imported Document/Evidence/Dossier/Approval/retention business facts remain owned by their respective domains. Interchange owns only the boundary/process/provenance needed to move truth without fabricating it.

The reviewer MUST challenge whether these four contracts actually share one coherent owner or whether `Interchange` is an abstraction caused by grouping unrelated edge operations.

---

# 5. Candidate mechanisms/projections that are NOT target authorities

The candidate explicitly refuses bounded-context status to:

```text
Search
Notifications
jobs / schedulers
outbox / queue
workers
storage providers
rendering providers
external repository adapters
RLS
HTTP routing
OpenAPI/codegen
cache
rate limiting
observability
```

## Search

Rebuildable/discovery projection/query component. It may consume published query/authz contracts but cannot own access semantics. Projection membership never proves canonical authorization.

## Notifications

Delivery/inbox projection. Producer owners resolve the business meaning/recipients required by their semantics; Notifications owns delivery/read-state only and cannot turn a notification read into acknowledgement/approval/business evidence.

## Jobs/outbox/workers

Execution mechanics. Owner-specific intents remain attributable to semantic producers; retry/scheduling/lease/DLQ may be shared machinery.

## Rendering providers

Rendition meaning belongs to Controlled Information. EigenPal/Gotenberg/docx-renderer/etc. are replaceable adapters/executors and receive no Document/Submission authority.

## Storage providers

Local/MinIO/AWS S3 are adapters behind Artifact/Managed Artifact Store mechanics. Provider location never becomes Artifact/Submission identity.

## External repository connectors

SharePoint/etc. are Interchange adapters for explicit `IMPORT_COPY` / `PUBLISH_COPY`; no silent synchronization or external-edit mutation of native immutable history.

---

# 6. Candidate cross-owner coordination — no new orchestration domains

The candidate deliberately avoids new `Governance`, `Release`, `Lifecycle`, `Workflow` or generic `Orchestrator` bounded contexts.

## Submission coordination

```text
accepted WorkingContent
→ immutable RevisionSubmission
→ first-submission RetentionBinding
→ ApprovalInstance when configured
→ audit + durable intents
```

`RevisionSubmission` remains Controlled Information authority.

## Release coordination

```text
exact RevisionSubmission
+ Approval satisfied when required
+ required Rendition ready when required
+ governing release preconditions
→ Controlled Information atomic effectivity transition
```

The coordinator does not own `EFFECTIVE`, `SUPERSEDED`, `ReleaseRecord`, ApprovalDecision or Rendition identity.

## Disposition coordination

```text
Records Governance eligibility + no active hold
+ governed subject still valid for disposal
+ Artifact physical removal
→ Records Governance verified completion / DispositionRecord
```

Records Governance owns the disposition decision/completion semantics; Artifact owns physical byte truth.

## Tenant Erasure coordination

```text
Organization deletion request reaches execution threshold
+ Records Governance blocker evaluation
+ Authentication session revocation
+ eligible business rows/blob destruction
+ DEK destruction only when lawful
+ erasure tombstone / restore reconciliation
→ Organization marks Tenant ERASED
```

Organization is the only owner of Tenant lifecycle state.

Review whether these compositions preserve ownership or conceal a missing first-class owner/transaction boundary.

---

# 7. Candidate authority dependency direction

This is a semantic dependency sketch, not the final R10-B/R10-D call/event topology:

```text
Organization
  ├─> Authentication       (organization principal/lifecycle references; AuthN remains separate)
  ├─> Authorization
  └─> bounded tenant-scoped configuration owners

Authorization
  └─> canonical authorization service consumed by request/query owners

Artifact
  ├─> Controlled Information consumes exact-byte references
  └─> Documentary Context consumes exact-byte references

Controlled Information
  ├─> Approval binds exact RevisionSubmission
  ├─> Records Governance observes/receives retention-subject facts
  ├─> Distribution receives released-revision obligation facts
  └─> Interchange imports/exports through published contracts

Documentary Context
  ├─> Records Governance observes/receives Evidence retention-subject facts
  └─> Interchange imports/exports through published contracts

Search / Notifications
  └─> consume published projections/events; never become upstream authority
```

The final package/import DAG must be acyclic. A conceptual consumer relation does not automatically imply a Go import in that direction; published narrow interfaces/events/composition may invert technical dependencies while preserving semantic authority.

The reviewer must distinguish semantic dependency from compile-time dependency and identify any hidden cycle the candidate has hand-waved away.

---

# 8. Candidate filesystem/package classification

Names are mechanically reversible; ownership is the material decision. Candidate layout:

```text
internal/
  modules/
    authentication/
    organization/
    authorization/
    controlledinformation/
    approval/
    documentarycontext/
    recordsgovernance/
    distribution/
    dictionary/
    audit/
    interchange/

  support/
    artifacts/

  projections/
    search/
    notifications/

  composition/
    <only concrete cross-owner application compositions>

  platform/
    <commodity mechanics: db/http/async/observability/etc.>
```

Within an owner, use only as justified by actual code/consumer need:

```text
domain/
application/
infrastructure/
delivery/
api/
public/          # only with a real second-package/cross-owner consumer
```

Do not treat package naming preference as a blocker. Attack whether these classifications make illegal dependencies obvious and whether `support/artifacts` is the correct structural signal for a semantic-but-non-business owner.

Provider placement candidate:

```text
Local / MinIO / AWS S3 adapters
  → Artifact infrastructure

EigenPal / render / Gotenberg adapters
  → Controlled Information infrastructure/execution side

SharePoint / external repository adapters
  → Interchange infrastructure
```

---

# 9. Candidate disposition of all 15 current backend modules

Current implementation is evidence only. Candidate target disposition:

| Current module | Candidate target disposition |
|---|---|
| `approval` | Replace/converge to frozen Approval V1 |
| `audit` | Retain as supporting semantic owner; rederive durability contract |
| `auth` | Rename/converge to Authentication |
| `controlleddocuments` | Delete as target boundary; legitimate identity/code/numbering/effectivity responsibilities move to Controlled Information |
| `distribution` | Retain semantic owner; rederive from frozen obligation/ack semantics |
| `documents` | Delete as legacy boundary; legitimate responsibilities move to Controlled Information |
| `iam` | Delete/split into Organization + Authorization |
| `jobs` | Delete as business module; owner-specific work + shared async mechanics |
| `notifications` | Reclassify as projection/delivery component |
| `render` | Dismantle mixed boundary; Rendition semantics to Controlled Information, providers/execution to infrastructure/mechanisms |
| `search` | Reclassify as projection/query component |
| `security` | Delete as business BC; authentication/account facts to Authentication, tenant lifecycle to Organization, pure operational security to platform/supporting mechanics |
| `taxonomy` | Delete/split: Area→Organization; DocumentType/classification→Controlled Information; frozen-deleted governance duplicates do not survive |
| `templates` | Delete parallel lifecycle; template role/designation/TemplateUse move to Controlled Information |
| `tokens` | Rename/converge to Dictionary candidate |

The reviewer must attack every row for retained hidden meaning, dropped meaning, duplicated owner or inertia-driven preservation.

---

# 10. Candidate OpenAPI/frontend ownership consequences — classification only

R10-A is not selecting final endpoints/DTOs/routes. It only classifies owner surfaces.

Candidate OpenAPI tag direction:

```text
auth                    → authentication
iam                     → organization + authorization

documents
controlled-documents
templates
taxonomy(document side) → controlled-information

taxonomy(area side)     → organization
approval                → approval
distribution            → distribution
tokens                  → dictionary
audit                   → audit
search                  → search projection
notifications           → notifications projection
security                → retire/split
configuration/health/observability → platform

new semantic surfaces where needed:
documentary-context
records-governance
interchange
```

Candidate negative rule: no generic public business `/artifacts` API that creates an orphan Artifact library.

Frontend feature ownership should follow the same semantic decomposition; frontend pages do not create product authority.

R10-E will choose exact public contracts/journeys after data/transaction/async semantics close.

---

# 11. Candidate invariants / R10-A proof obligations

Before R10-A may close, the architecture must demonstrate:

1. **owner completeness:** every frozen durable/business fact has exactly one semantic owner;
2. **owner uniqueness:** no fact has two independent writers/authorities;
3. **DAG:** no unavoidable semantic or package cycle;
4. **published seams:** cross-owner consumers can use narrow published contracts/events instead of repositories/SQL/domain internals;
5. **mechanism separation:** providers/jobs/projections cannot mutate or redefine owner meaning;
6. **negative boundaries:** Controlled Information, Artifact and Interchange cannot expand into God/ECM/integration-platform contexts;
7. **legacy disposition totality:** every current module has a target disposition, not an indefinite compatibility status;
8. **downstream feasibility:** the ownership split leaves a credible route for R10-B transaction/DB invariants and R10-F migration/deletion without inventing dual authority;
9. **YAGNI:** no target module exists solely for imagined future consumers;
10. **Structural Inversion:** the ownership result still follows if current modules/tables/providers were shaped differently.

---

# 12. Required Fable adversarial attack matrix

Do not perform a polite consistency review. Try to falsify the candidate.

## A. Missing / duplicate authority

- Is any frozen fact ownerless?
- Does any owner duplicate another owner or a frozen cross-cutting authority?
- Is `Artifact` an authority that should instead be a mechanism, or does collapsing it create wrongful Document→Evidence ownership?
- Does `Audit` own durable meaning distinct enough to justify its classification without becoming business authority?

## B. Wrong bounded-context cuts

- Is Controlled Information too large and actually several independent models, or is splitting it what would recreate duplicate lifecycle authority?
- Should Distribution be a business BC or a supporting concern?
- Does Documentary Context incorrectly combine Evidence and Dossier?
- Should Records Governance own retention + hold + disposition together?
- Does Authentication vs Organization split leave User identity ambiguous?
- Is Authorization sufficiently separate from Organization without creating cycles?

## C. Interchange challenge

- Are Historical Migration, portability export, governed export and external repository copy actually one coherent semantic owner?
- Does this become a generic integration platform by abstraction?
- Would smaller owner-specific adapters plus a common technical package be more sustainable?
- If split, who owns transfer attempt/provenance/idempotency/reconciliation truth without duplication?

## D. Dependency/cycle challenge

Construct the strongest plausible cycles, especially:

```text
Organization ↔ Authentication
Organization ↔ Authorization
Controlled Information ↔ Approval
Controlled Information ↔ Records Governance
Controlled Information ↔ Artifact
Documentary Context ↔ Records Governance
Interchange ↔ every business owner
Audit ↔ every mutation owner
```

For each, decide whether it is a semantic cycle, technical import cycle, transaction-coordination issue, or harmless reference relation. Demand an explicit seam where needed.

## E. Transaction-boundary preview

Without designing R10-B fully, test whether the ownership split makes frozen atomicity impossible or forces cross-owner distributed transactions for:

- Submission + RetentionBinding + ApprovalInstance + audit/outbox intent;
- Release + Approval/Rendition preconditions;
- LegalHold prospective materialization;
- verified disposition;
- tenant erasure.

If a cross-owner invariant requires one local SQL transaction, determine whether shared database transaction coordination can preserve semantic ownership or whether the owner split is wrong.

Do not demand microservices or separate databases. Deployment topology is not bounded-context authority.

## F. AuthZ / isolation

- Can canonical Authorization support query/search/export without foreign domains reimplementing visibility predicates?
- Do resource/case relationship checks remain with owning domains or accidentally move business semantics into Authorization?
- Are Dossier links structurally unable to grant access?
- Can projection code avoid becoming a second policy engine?

## G. Provider/editor/storage overfit

- Would reversing MinIO/S3/local, EigenPal/other editor, Gotenberg/other renderer or SharePoint/other repository alter any ownership conclusion?
- If yes, identify provider leakage.

## H. Legacy-inertia challenge

For every retained candidate owner, ask whether it exists because a current module exists.
For every deleted module, ask whether legitimate independent semantics have been lost.

Explicitly challenge:

```text
Distribution
Audit
Dictionary
Artifact
Interchange
```

because they are the least obvious classifications.

## I. YAGNI / subtractive pass

Try to remove a candidate owner without weakening a distinct material property. Conversely, identify any merged owner that reduces file/module count while increasing semantic ambiguity or future structural cost.

## J. Hardest future evidenced changes

Only use future seams already evidenced in frozen authority, such as:

```text
external IdP
realtime collaboration
SharePoint Embedded enterprise profile
advanced content security
PKI/signature level
ArtifactPackage after real multi-file requirement
second records/export/integration role
```

Do not invent speculative futures. Check whether the candidate seams allow these without rewriting core identity/authority.

---

# 13. Reviewer output contract

Return one independent review with this structure.

## 13.1 Authority reconstruction

State whether the repo authority/status was reconstructed successfully and whether any material contradiction exists before judging the candidate.

## 13.2 Material findings

For each finding:

```text
F<N> — BLOCKER | MAJOR | LOW
Claim
Repository authority/evidence anchors
Root cause
Frozen/target invariant affected
Strongest credible alternatives
Global Maximum analysis
YAGNI / essential-vs-accidental complexity
Required change
R9.5 reopen? YES/NO
Proof needed to close
```

Do not include stylistic/package-name preferences as material findings.

## 13.3 Candidate disposition table

For each candidate owner:

```text
CONFIRM | MERGE | SPLIT | RECLASSIFY | REMOVE | UNKNOWN
```

with one-line basis.

## 13.4 DAG verdict

State whether the semantic DAG can be made acyclic with published seams. List every material edge that must be explicit in R10-A before closure.

## 13.5 Legacy disposition verdict

Attack the complete 15-module delete/re-home map and identify any missing current-state meaning that later migration must preserve.

## 13.6 R9.5 reopen set

Expected default is EMPTY. If non-empty, use the strict five-part reopen contract in §1.

## 13.7 Final verdict

Exactly one:

```text
VERDICT: APPROVE R10-A
VERDICT: APPROVE R10-A WITH MATERIAL FIXES
VERDICT: DO NOT APPROVE R10-A
```

Then state the single biggest remaining structural risk.

---

# 14. Stop condition

Stop when:

- authority is reconstructed;
- every candidate owner has been challenged;
- duplicate/missing authority and cycles have been attacked;
- strongest credible alternative decompositions were compared;
- the subtractive/YAGNI pass is complete;
- any R9.5 reopen claim is fully substantiated;
- no material finding remains unstated merely because implementation would be difficult.

Do **not** design R10-B schemas/transactions, write implementation code, propose migration SQL, change OpenAPI or frontend, or author an implementation plan.

This review exists to determine whether R10-A is the **Global Maximum ownership topology** for the already-frozen product — not whether it is the easiest architecture to implement from today's repository.