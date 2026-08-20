# MetalDocs Launch V1 — Ownership Topology

> **Status:** ACTIVE / OPERATOR-APPROVED ARCHITECTURE AUTHORITY  
> **Accepted:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Product authority:** `wiki/architecture/launch-v1-product-contract.md`  
> **GCR authority:** `wiki/architecture/whole-product-alignment-review.md`  
> **Implementation:** BLOCKED

This page replaces the prior R10-A Launch ownership conclusion. It defines **semantic ownership only**. It does not define package layout, tables, schemas, storage ports, API routes, workers or implementation structure.

The operator approved this topology on 2026-08-18 together with an explicit evolution constraint: known future capabilities must not be forgotten or made structurally expensive merely because they are deferred from Launch. The governing law is therefore:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

---

## 1. Decision

Launch V1 has exactly **four business semantic owners plus one supporting semantic owner**:

```text
BUSINESS
1. Authentication
2. Organization
3. Authorization
4. Controlled Documents

SUPPORTING SEMANTIC
5. Audit
```

Everything else is mechanism, projection, operations/cutover capability, Launch+ or Future until a concrete consumer proves an independent semantic lifecycle.

Method outcome:

```text
RESTRUCTURE NOW
prior R10-A 8+3 topology → superseded for Launch
replacement              → 4+1 semantic ownership topology
```

This minimizes accidental cross-owner atomicity without collapsing independently evolving authorities.

---

## 2. Authentication

Authentication owns the minimum MetalDocs-facing authentication truth:

```text
provider-subject binding
application Session
Session lifecycle / revocation
authentication-assurance / fresh-auth evidence when a named consumer requires it
IdP anti-corruption boundary
```

The authentication provider owns credentials, password policy/recovery, MFA/passkeys, upstream federation and provider authentication journeys.

Authentication does **not** own:

```text
organizational User identity
Area / Group
Role / Permission
document access policy
document governance
```

Provider roles/groups/organizations/permissions never become canonical MetalDocs Authorization merely because the provider exposes them.

---

## 3. Organization

Organization owns who exists in the company and how people are organized:

```text
single-company root
User
separately erasable User profile/enrichment
Area
Group
GroupMembership
organizational User lifecycle / offboarding identity
```

It does not decide what an actor may do. Organizational relationships are inputs to Authorization and Controlled Documents without transferring ownership.

---

## 4. Authorization

Authorization owns product access authority:

```text
product Role semantics
product Permission semantics
RoleAssignment / grant-revoke current truth
scope semantics
canonical grant evaluation
```

Exact Launch roles, permissions and bundles are **not inherited from the prior 5×43 catalog**. They are re-derived later from accepted Launch journeys and must include a least-privilege Auditor / Governance Viewer path.

Authorization owns grants. The owning business domain owns case/resource relationship meaning and lifecycle predicates.

Conceptually:

```text
Authorization grant + scope
+ Controlled Documents relationship/state predicate
= ALLOW or default DENY
```

No provider-role bridge, generic ACL/ReBAC graph or hidden bypass is implied.

---

## 5. Controlled Documents

Controlled Documents owns the complete Launch controlled-document meaning and lifecycle:

```text
Document Type + numbering
Controlled Document stable identity
Business Revision
mutable DRAFT Working Content + concurrency/recovery semantics
immutable Submission attempt
Template role / eligibility / origin semantics required by Launch
sequential document-governance semantics
Submission governance
feedback / ACCEPT / RETURN_FOR_CHANGES
withdraw governance attempt
Revision cancellation
required official Rendition semantics
system-owned Release / effectivity
EFFECTIVE / SUPERSEDED
explicit governed OBSOLETE without replacement
revision / lifecycle / provenance history
exact-content facts attached to the semantic record that freezes them
```

`Controlled Documents` is one **semantic authority**, not one giant aggregate, file, package or transaction. Internal responsibility clusters may remain small and independently testable; they do not become separate semantic owners merely because code is separated.

### Governance placement

Launch has two proven document-governance consumers:

```text
1. govern an immutable Submission
2. govern obsolescence of the current EFFECTIVE Document without a successor
```

The smallest common sequential governance semantics may be reused by both journeys. Launch does **not** create a generic arbitrary-subject BPM/workflow platform.

A future capability with a genuinely independent governance lifecycle may later justify a separate owner; current separation is not preserved by sunk cost.

### Exact content

There is no standalone `Artifact` semantic owner.

Exact byte facts such as hash, size, format and governed-content identity belong to the semantic record that freezes or owns that content. Storage handle, provider key, staging object, upload state, scanner execution and physical location remain mechanisms.

This preserves:

```text
semantic content identity != storage/provider identity
```

without creating an intermediate domain solely to mediate bytes.

---

## 6. Audit

Audit is the one supporting semantic owner because it has independent transversal meaning:

```text
immutable action/timeline evidence
actor attribution
trusted action time
operation/resource attribution
bounded PII-minimized audit facts
```

Audit never owns or reconstructs current business state.

Required governed operations may require same-local-commit Audit evidence. The deployment-wide cryptographic `AuditChainHead` / global-lock design is **not** a Launch requirement unless a concrete assurance requirement reopens it.

---

## 7. Not semantic owners in Launch

The following remain explicitly outside semantic ownership unless future evidence promotes them:

```text
storage / staging / byte integrity / malware inspection
rendering / viewers / editor providers
Search
async jobs / outbox / retry / lease / DLQ
notifications
backup / restore transport
Historical Migration execution/cutover machinery
```

Classification:

```text
storage/integrity     → mechanism
render/view/editor    → mechanism
Search                → rebuildable projection
async                 → durable/ephemeral mechanism as required
Historical Migration → cutover capability that writes through owning semantic seams
backup/restore        → operations/readiness concern
```

Historical Migration never becomes a generic `Interchange` domain. Imported enduring truth belongs to the semantic owner whose truth was imported.

---

## 8. Semantic dependency shape

Conceptually:

```text
Authentication
    ↓ binds authenticated subject to product User
Organization
    ↓ supplies Users / Groups / Areas
Authorization
    ↓ supplies grants / scopes
Controlled Documents
    ↓ emits required transversal action evidence
Audit
```

This is an **authority dependency shape**, not a package-import or transaction prescription.

Search, storage, rendering, migration, async execution and backup/restore sit around these owners as mechanisms/projections/cutover/operations and may not acquire business meaning by convenience.

---

## 9. Known future capability horizon

Deferral does **not** mean the future capability is forgotten or architecturally irrelevant.

The following are declared expected evolution and therefore count as **evidence of future direction** under the DevelopmentConexus Engineering Method. They justify preserving seams where doing so is materially cheaper than later dismantling core authority; they do **not** justify dormant modules/tables/permissions/jobs today.

### Launch+

| Capability | Required attachment seam | Must not become |
|---|---|---|
| Distribution / Read & Acknowledge | Released Document/Revision + concrete User/Group audience | effectivity authority or access grant |
| Periodic Review | stable Document + exact current EFFECTIVE Revision | Approval route or effectivity authority |

### Future product capabilities

| Capability | Expected attachment seam | Boundary to preserve |
|---|---|---|
| Dossier / documentary context | stable Document identity and future Evidence identity | context must not own content or grant access |
| Evidence / quality records | Organization/AuthZ + exact-content mechanism; likely independent lifecycle when promoted | must not be forced through Document REV/Release without requirement |
| Retention / Legal Hold / Disposition | stable governed-subject identities and immutable lifecycle history | records policy must not become Document lifecycle or storage-provider authority |
| Governed Export | stable semantic identities, relationships and exact-content facts | export packaging must not become source authority |
| External Repository IMPORT/PUBLISH | target-owner import seams and exact-content snapshots | provider object identity must not become MetalDocs identity |
| Training/LMS | effective document/release and future distribution obligations | training competence must not become document effectivity |
| Generic/multi-document Change Control | stable Document/Revision identity and explicit change initiation seams | orchestration must not take over Document/Revision authority |
| pooled multi-customer tenancy | durable company-root identity and deliberately reopenable deployment/substrate boundary | no universal partition machinery before the requirement exists |
| realtime coauthoring / CRDT | replaceable DRAFT Working Content concurrency mechanism | must not change business Revision or immutable Submission identity |

These seams describe **what must remain attachable**, not future table/module shapes.

---

## 10. Future-evolution law for remaining technical design

Every material technical decision after this topology must run this check:

1. **Launch correctness first:** does the decision satisfy the accepted Launch invariant without relying on a future capability?
2. **Named-future compatibility:** can the known future capabilities above attach without changing the meaning/identity of existing Document, Revision, Submission, Release, User/Group/Area or Audit history?
3. **No history rewrite:** would adding the future feature require rewriting immutable governed history or fabricating facts? If yes, current design is presumptively wrong.
4. **Additive evolution where reasonable:** prefer a seam that permits future capability addition through new owner/state plus bounded migration, rather than dismantling current authority.
5. **No dormant implementation:** do not create unused modules, tables, permissions, workers, generic registries or feature flags solely for the future.
6. **No generic framework by anticipation:** known future direction justifies an attachment seam, not a generic ECM/BPM/records/integration platform.
7. **Record unavoidable future cost:** if a current decision knowingly makes a named future capability materially expensive, record why the Launch benefit outweighs the future cost and state the reopen trigger before accepting it.

A later capability may become a **new semantic owner** when it gains a real independent lifecycle/consumer. The 4+1 topology is the smallest Launch authority set, not a claim that MetalDocs will forever have only five owners.

---

## 11. Future-proofing invariants

Technical architecture must preserve these stable anchors unless later evidence explicitly reopens them:

```text
User / Group / Area identity remains independent of AuthN provider identity
Document identity remains stable across Revisions
Revision remains a business change cycle
Working Content remains replaceable DRAFT authority/mechanism boundary
Submission remains immutable exact governed-attempt identity
Release remains effectivity authority
Audit remains evidence, not current state
storage/provider identity never becomes semantic identity
future contexts attach by reference rather than duplicate core authority
```

This is the intended balance:

```text
YAGNI today
+ explicit known future horizon
+ stable semantic anchors
+ replaceable mechanisms
= sustainable evolution without speculative platform code
```

---

## 12. Reopen triggers

Reopen this ownership topology when material evidence shows one of:

- a deferred/Launch+ capability now has a concrete consumer and independent lifecycle that merits a new owner;
- Controlled Documents accumulates unrelated authority rather than cohesive controlled-document semantics;
- a future capability cannot attach without duplicating or rewriting core authority;
- AuthN, Organization or Authorization cease to evolve independently in a way that justifies their boundary;
- a new cross-repository/trust boundary creates independently owned truth;
- implementation evidence shows a boundary creates materially more accidental cross-owner complexity than it prevents.

Do not reopen merely because a future feature exists on the horizon; the explicit seam is already the preparation.

---

## 13. Gate / next step

Ownership/topology is closed and operator-approved.

```text
Product Contract
→ Whole-Product GCR / A1–A10
→ Launch V1 ownership topology 4+1  ✅
→ re-derive remaining technical architecture  NEXT
→ Whole-R10 Global Coherence Review
→ cold independent review
→ final operator ratification
→ implementation spec/plan
→ code
```

Implementation remains **BLOCKED**.