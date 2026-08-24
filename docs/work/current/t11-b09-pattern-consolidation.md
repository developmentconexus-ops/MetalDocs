# T11 — B09 P10 Bounded Pattern Consolidation

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B09 — Audit.  
> **Method:** Frontend Product Experience Planning Method v2.3.  
> **Locked P8 Git blob:** `7daa6054851e617aeacb95a28d907d0d6d4bd3d6`.  
> **Rule:** shared patterns graduate only from repeated LOCKED semantic/protected behavior; visual similarity is insufficient.

## 1. Goal

Compare B09 against already-LOCKED B01/B01N/B02/B03/B04/B05/B06/B07/B08 and reuse or graduate only patterns whose semantics truly match.

## 2. Existing shared patterns reused

### Global App Shell — REUSE

Origin: B01.

B09 reuses the locked application shell, global navigation posture and responsive shell transformation. Audit remains under the accepted `Evidência` navigation home and stable `/audit` route; B09 creates no Audit-specific global frame.

### Notification Quick Inbox — GLOBAL CHROME ONLY

Origin: B01N.

B09 inherits the already-locked global Notification chrome because it is part of the application shell. It does not consume or alter Notification Product semantics, and no Audit state enters the Quick Inbox.

## 3. Existing patterns deliberately not imported

### B07 Revision Chapters / History timeline — NOT APPLICABLE

B07 is one Controlled Document's semantic lifecycle history. B09 is cross-product action evidence ordered by immutable Audit occurrence.

```text
B07 History
  semantic lifecycle story of one Document

B09 Audit
  cross-product action evidence under historical visibility
```

Shared chronological geometry is insufficient to merge these meanings.

### B08 triage Inbox — NOT APPLICABLE

Audit evidence is not personal attention. B08 unseen/read/archive lenses and engagement mutations do not belong in B09.

### B05 work queue / B06 Governance Case — NOT APPLICABLE

Audit has no assignment, responsibility, decision or case-management lifecycle. Owner handoffs are secondary current-context continuations only.

### Exact Read-Only Content Viewer Shell — NOT A B09 CONSUMER

B09 detail inspects structured Audit evidence already loaded in op78. It does not render exact Document bytes. Current owner context is handed off to B03/B07/B06 when admitted.

## 4. B09-local semantic patterns — DO NOT GRADUATE

### Audit Investigation Ledger

Status: **LOCAL B09 PATTERN**.

```text
five-dimensional structured investigation
+ applied-query chips
+ dense evidence ledger
+ exact evidence detail
+ explicit cursor continuation
```

This is specific to Audit and must not become a generic table/search platform.

### Draft vs applied investigation query

Status: **LOCAL B09 PATTERN**.

```text
draft editor
  != applied evidence question

applied query
  = URL + chips + ledger truth
```

This protects Audit investigation stability while editing filters. It does not graduate a generic form/query framework.

### Semantic-dimension applied chips

Status: **LOCAL B09 PATTERN**.

Compound wire members remain one user-semantic dimension:

```text
Period   = from + before
USER     = actor_kind + actor_user_id
Resource = resource_kind + optional exact resource_id
```

Immediate chip removal triggers a new canonical read and never constructs wire-invalid partial predicates.

### Query Assist constrained by Audit-visible evidence

Status: **LOCAL B09 COMPOSITION PATTERN**.

```text
op87 / op88 / op89
-> bounded server-authored candidate identities
-> exact stable predicate identity
-> op78 independently rechecks current audit.read + historical visibility
```

This is not a generic reference-data service or admin-directory pattern.

### Immutable evidence + optional current recognition

Status: **LOCAL B09 PRESENTATION PATTERN**.

```text
immutable Audit evidence
  = proof

optional current recognition
  = scan/readability context only
```

Similar presentation concerns exist elsewhere, but B09's historical-visibility/evidence boundary is distinct enough that no new generic Activity/Event row is frozen here.

### Audit-native same actor / resource / action continuation

Status: **LOCAL B09 INVESTIGATION PATTERN**.

The detail drawer can directly transform one admitted event identity into a new structured op78 investigation. This stays inside Audit and does not graduate a generic "related items" engine.

### Contextual detail from already-loaded evidence

Status: **LOCAL B09 PATTERN**.

B09 detail uses the already-loaded `AuditInspectionItem`. No detail route or backend read is created merely because the UI presents more fields.

### Bounded owner-lens handoff

Status: **LOCAL B09 COMPOSITION PATTERN**.

```text
admitted exact Document / governance identity
-> existing B03 / B07 / B06 route boundary
-> destination rechecks current AuthZ/disclosure
```

This preserves owner separation and rejects a generic deep-link resolver.

### Presentation-only local-day separators

Status: **LOCAL B09 PATTERN**.

Day separators improve scanability without changing op78 canonical ordering, server grouping, cursor meaning or counts.

### Cursor continuation preserving loaded evidence

Status: **LOCAL B09 PATTERN**.

A failed continuation preserves already-loaded immutable evidence and offers retry. Although B07/B08 also preserve loaded rows on continuation failure, their owner/order/disclosure semantics differ; no generic Product pagination abstraction is frozen by planning.

### Canonical applied-query URL round-trip

Status: **LOCAL B09 NAVIGATION PATTERN**.

Only stable IDs/enums/UTC instants enter the applied URL. Draft edits and cursor/load depth remain local. URL reconstruction reproduces the evidence question, never authorization.

## 5. Similarity explicitly rejected as insufficient

```text
B07 chronological rows vs B09 chronological rows
  -> lifecycle history != transversal action evidence

B08 list + filters/lenses vs B09 ledger + investigation
  -> personal attention lifecycle != evidence predicates

B05/B06 row action geometry vs B09 owner handoffs
  -> work/governance responsibility != read-only Audit context

Generic typeahead similarity
  -> Audit Query Assist candidate law is evidence-visibility bounded

Generic detail drawer geometry
  -> B09 detail is loaded Audit evidence, not a reusable detail-domain contract
```

## 6. Prototype-only constructs — NEVER GRADUATE

```text
review-only failure toggles
forced known-empty state
fake local request timing
fixture actor/resource/Area candidates
fixture History API failure path
fixture route-denied switch
prototype-generated event rows
```

They exist only to make P8 falsifiable.

## 7. Pattern vocabulary effect

Existing locked shared patterns reused:

```text
Global App Shell
Notification Quick Inbox as inherited global chrome only
```

New shared semantic patterns graduated by B09:

```text
none
```

B09-local semantic/composition patterns retained:

```text
Audit Investigation Ledger
Draft vs applied investigation query
Semantic-dimension applied chips
Audit-visible Query Assist
Immutable evidence + optional current recognition
Audit-native same actor/resource/action continuation
Contextual detail from loaded evidence
Bounded owner-lens handoff
Presentation-only local-day separators
Cursor continuation preserving loaded evidence
Canonical applied-query URL round-trip
```

## 8. Anti-abstraction decisions

Rejected:

```text
Generic Audit/Search framework
Generic Filter/Chip engine as Product authority
Generic Reference Data / Directory service
Generic Activity/Event feed
Generic Entity Detail endpoint
Generic DeepLink resolver
Generic Evidence/History timeline
Generic Pagination framework from UX similarity
Generic Authorization-aware frontend router
```

Implementation may later share low-level UI primitives where convenient, but P10 freezes no new cross-Product semantic/component contract from B09 alone.

## 9. Reopen / graduation triggers

A B09-local pattern may graduate only after another LOCKED block proves materially matching:

```text
human job
truth owner
stable identity source
query semantics
ordering/pagination law
access/disclosure posture
failure/recovery class
historical/current-recognition separation
responsive/accessibility meaning
```

Repeated CSS or component convenience alone is insufficient.

## 10. P10 closure

```text
existing locked shared patterns reused/inherited  2
new shared semantic patterns graduated             0
B09-local semantic/composition patterns           11
false abstractions introduced                      0
Audit/History semantic merges                      0
Audit/Inbox or Audit/work semantic merges          0
generic search/reference/deep-link frameworks      0
```

P10 is complete for the operator-locked B09 scope.
