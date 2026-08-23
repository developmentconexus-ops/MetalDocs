# MetalDocs Launch V1 — Ownership Topology

> **Status:** ACTIVE / OPERATOR-APPROVED ARCHITECTURE AUTHORITY  
> **Accepted baseline:** 2026-08-18  
> **T11 bounded reopen:** 2026-08-22 — stable-Document Discussion / `@mention` / Notifications, Lead GCR + fresh Fable CONVERGED  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Implementation:** BLOCKED

This page defines **semantic ownership only**. It does not define package layout, tables, schemas, storage ports, API routes, workers or implementation structure.

The governing evolution law remains:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

The T11 bounded reopen exercised the existing trigger exactly as intended: Notifications was previously deferred because no concrete consumer or independent lifecycle existed; explicit Document `@mention` plus a persistent recipient-controlled Inbox created both. The topology therefore changes only by the smallest proven owner addition.

---

## 1. Decision

Launch V1 has exactly **four business semantic owners plus two supporting semantic owners**:

```text
BUSINESS
1. Authentication
2. Organization
3. Authorization
4. Controlled Documents

SUPPORTING SEMANTIC
5. Audit
6. Notifications
```

Everything else is mechanism, projection, operations/cutover capability, Launch+ or Future until a concrete consumer proves an independent semantic lifecycle.

Current Method outcome:

```text
prior Launch topology  4 + 1
new concrete consumer  stable-Document @Mention + persistent Inbox engagement
bounded restructure    4 + 2
```

The new supporting owner does not imply another service, database, broker or runtime.

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

Authentication does **not** own organizational User identity, Area/Group, Role/Permission, document access policy, document governance or Notification engagement.

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

It does not decide what an actor may do and does not own the recipient's Notification attention lifecycle. Organizational relationships are inputs to Authorization, Controlled Documents and Notifications choreography without transferring ownership.

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

Authorization owns grants and final ALLOW/default-DENY. The owning business domain owns case/resource relationship meaning and lifecycle/disclosure predicate facts.

Conceptually:

```text
Authorization grant + scope
+ Controlled Documents relationship/state/disclosure predicate
= ALLOW or default DENY
```

The T11 Discussion capability adds `document.discuss` as a write-participation Permission; reading the stable-Document Discussion remains governed by the canonical current Document Discussion disclosure composition, not by a second read Permission.

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
stable-Document Discussion
immutable DiscussionMessage
semantic Mention(user_id)
```

`Controlled Documents` is one **semantic authority**, not one giant aggregate, file, package or transaction. Internal responsibility clusters may remain small and independently testable; they do not become separate semantic owners merely because code is separated.

### Discussion placement

Discussion belongs to stable Document identity across Revisions. A message may carry the exact official Revision identity that existed when posted as contextual provenance, but Discussion is not Revision history, WorkingContent/editor comments, SubmissionFeedback or GovernanceCase feedback.

A message may reference one prior message; no generic Thread/chat domain is introduced. Accepted Launch messages are immutable; correction is a new message.

### Governance placement

Launch has two proven document-governance consumers:

```text
1. govern an immutable Submission
2. govern obsolescence of the current EFFECTIVE Document without a successor
```

The smallest common sequential governance semantics may be reused by both journeys. Launch does **not** create a generic arbitrary-subject BPM/workflow platform.

### Exact content

There is no standalone `Artifact` semantic owner.

Exact byte facts such as hash, size, format and governed-content identity belong to the semantic record that freezes or owns that content. Storage handle, provider key, staging object, upload state, scanner execution and physical location remain mechanisms.

---

## 6. Audit

Audit is a supporting semantic owner because it has independent transversal meaning:

```text
immutable action/timeline evidence
actor attribution
trusted action time
operation/resource attribution
bounded PII-minimized audit facts
```

Audit never owns or reconstructs current business state.

DiscussionMessage already owns immutable author/time/content truth; Launch does not duplicate each message into Audit solely because it exists. Notification creation/engagement/realtime also has no mandatory semantic AuditEvent absent a concrete assurance requirement.

---

## 7. Notifications

Notifications is the second supporting semantic owner because the current Launch Inbox has independent persistent recipient-controlled meaning that cannot be rebuilt from its source Mention.

Notifications owns:

```text
Notification identity
recipient User reference
closed Notification kind
closed source reference
created_at
deduplication identity
seen_at
read_at
archived_at
Inbox ordering/current engagement state
```

Current Launch kind:

```text
DOCUMENT_MENTION
  document_id
  message_id
```

Notifications does **not** own:

```text
Document identity/title/lifecycle
DiscussionMessage content/validity
Mention target eligibility
User/Profile identity
Authorization/disclosure
Audit/history
email/push transport
SSE/realtime transport
River jobs
event bus
```

A Notification never grants or preserves source access and never becomes equivalent to document read/acknowledgement/governance evidence.

Current source presentation is resolved under current disclosure at read time. Persistent Notification rows do not copy source ACL/presentability truth merely to make Inbox reads convenient.

---

## 8. Not semantic owners in Launch

The following remain explicitly outside semantic ownership unless future evidence promotes them:

```text
storage / staging / byte integrity / malware inspection
rendering / viewers / editor providers
Search
async jobs / outbox / retry / lease / DLQ
realtime connection / SSE wake-up transport
backup / restore transport
Historical Migration execution/cutover machinery
```

Classification:

```text
storage/integrity     → mechanism
render/view/editor    → mechanism
Search                → rebuildable projection
async                 → durable/ephemeral mechanism as required
realtime wake-up      → ephemeral mechanism
Historical Migration → cutover capability that writes through owning semantic seams
backup/restore        → operations/readiness concern
```

Historical Migration never becomes a generic `Interchange` domain. Imported enduring truth belongs to the semantic owner whose truth was imported.

---

## 9. Semantic dependency shape

Conceptually:

```text
Authentication
    ↓ binds authenticated subject to product User
Organization
    ↓ supplies Users / Groups / Areas
Authorization
    ↓ supplies grants / scopes / final decisions
Controlled Documents
    ↓ owns source Document/Discussion/Mention facts
Notifications
    ↓ owns persistent recipient attention/engagement derived from accepted source facts
Audit
    ↓ owns required transversal evidence only where explicitly required
```

This is an **authority dependency shape**, not a package-import or transaction prescription.

For explicit accepted Mentions, application choreography composes Controlled Documents + Notifications in one caller-owned local PostgreSQL transaction so accepted Mention and required persistent in-app Notification cannot diverge. Owners do not import one another.

Search, storage, rendering, migration, async execution, realtime transport and backup/restore sit around owners as mechanisms/projections/cutover/operations and may not acquire business meaning by convenience.

---

## 10. Known future capability horizon

Deferral does **not** mean future capability is forgotten or architecturally irrelevant.

### Launch+

| Capability | Required attachment seam | Must not become |
|---|---|---|
| Distribution / Read & Acknowledge | Released Document/Revision + concrete User/Group audience | effectivity authority or access grant |
| Periodic Review | stable Document + exact current EFFECTIVE Revision | Approval route or effectivity authority |

Notification `READ` is explicitly **not** Read & Acknowledge.

### Future product capabilities

| Capability | Expected attachment seam | Boundary to preserve |
|---|---|---|
| additional Notification channels | persistent Notification intent + named durable delivery effect | channel transport must not own Notification/Product truth |
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

## 11. Future-evolution law for remaining technical design

Every material technical decision after this topology must run this check:

1. **Launch correctness first:** does the decision satisfy the accepted Launch invariant without relying on a future capability?
2. **Named-future compatibility:** can known future capabilities attach without changing the meaning/identity of existing core semantic records?
3. **No history rewrite:** would adding the future feature require rewriting immutable governed history or fabricating facts? If yes, current design is presumptively wrong.
4. **Additive evolution where reasonable:** prefer a seam that permits future capability addition through new owner/state plus bounded migration, rather than dismantling current authority.
5. **No dormant implementation:** do not create unused modules, tables, permissions, workers, generic registries or feature flags solely for the future.
6. **No generic framework by anticipation:** known future direction justifies an attachment seam, not a generic ECM/BPM/records/integration/event platform.
7. **Record unavoidable future cost:** if a current decision knowingly makes a named future capability materially expensive, record why the Launch benefit outweighs the future cost and state the reopen trigger before accepting it.

The 4+2 topology is the smallest current Launch authority set, not a claim that MetalDocs will forever have six owners.

---

## 12. Future-proofing invariants

Technical architecture must preserve these stable anchors unless later evidence explicitly reopens them:

```text
User / Group / Area identity remains independent of AuthN provider identity
Document identity remains stable across Revisions
Revision remains a business change cycle
Working Content remains replaceable DRAFT authority/mechanism boundary
Submission remains immutable exact governed-attempt identity
Release remains effectivity authority
Discussion remains stable-Document conversation, not DRAFT/Submission/Governance authority
Mention identity remains stable User reference
Notification remains recipient attention state, not access/acknowledgement authority
Audit remains evidence, not current state
storage/provider identity never becomes semantic identity
future contexts attach by reference rather than duplicate core authority
```

---

## 13. Reopen triggers

Reopen this ownership topology when material evidence shows one of:

- a deferred/Launch+ capability now has a concrete consumer and independent lifecycle that merits another owner;
- Controlled Documents accumulates unrelated authority rather than cohesive controlled-document semantics;
- Notifications loses its independent persistent engagement lifecycle and becomes truthfully rebuildable;
- a future capability cannot attach without duplicating or rewriting core authority;
- AuthN, Organization or Authorization cease to evolve independently in a way that justifies their boundary;
- a new cross-repository/trust boundary creates independently owned truth;
- implementation evidence shows a boundary creates materially more accidental cross-owner complexity than it prevents.

Do not reopen merely because a future feature exists on the horizon; explicit seams are already the preparation.

---

## 14. Review proof

The T11 Notifications promotion was challenged by:

```text
Lead GCR Round 1   NOT CONVERGED / MATERIAL=3 / IMPORTANT=6
corrected Round 2  CONVERGED / MATERIAL=0 / IMPORTANT=0
fresh Fable #165   CONVERGED / MATERIAL=0 / IMPORTANT=1 / OPTIONAL=3
```

The sole Fable IMPORTANT was consolidation-target completeness and changed no ownership decision. Fable explicitly confirmed `4+2 owners` survives the independent challenge. No second Fable round was justified.

Current bounded cross-layer details are routed by `docs/decisions/discussion-notifications-launch.md`.
