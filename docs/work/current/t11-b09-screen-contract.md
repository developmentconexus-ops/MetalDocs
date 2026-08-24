# T11 — B09 P9 Screen Contract / Bidirectional Trace

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B09 — Audit.  
> **Depends on:** operator LOCK of `t11-b09-audit-functional-wireframe.html`, `../../decisions/audit-investigation-read.md`, B01 global shell, B06 Governance Case, B07 Document History and B03 Document Official route boundaries.  
> **Locked P8 Git blob:** `7daa6054851e617aeacb95a28d907d0d6d4bd3d6`.  
> **Implementation:** BLOCKED.

## 1. Goal

Prove that every material region/control in the operator-locked B09 Audit Investigation Ledger is realizable by current authority without inventing Audit state, frontend Authorization, client-side completeness, a detail endpoint, generic search infrastructure or screen-shaped APIs.

## 2. Screen contract

| Surface / region | User goal | Current read truth | Material control / write | Identity source | Material failure / safe UX | Forbidden frontend authority | Status |
|---|---|---|---|---|---|---|---|
| B09-01 stable route | enter Audit investigation | op78 `listAuditEvents` under `audit.read` | route/read only | stable `/audit` route | route-level denial is not an empty result | no client `audit.read` evaluator | READY |
| B09-02 global shell | preserve Product orientation | B01 locked App Shell | normal global navigation | existing shell state | shell remains usable when Audit read fails | no Audit-specific primary IA | READY |
| B09-03 Audit heading | identify evidence/investigation purpose | ratified B09 Product job | none | stable route + Product terminology | unavailable state remains explicit | no analytics/dashboard semantics | READY |
| B09-04 investigation editor | narrow evidence by five admitted dimensions | op78 predicates + op87-89 assists | draft editing only | stable IDs/enums/UTC instants | invalid draft never changes applied results | no generic filter engine authority | READY |
| B09-05 Period | bound investigation by exact time interval | op78 `occurred_at_from/before` | draft preset/custom input, then Apply | canonical UTC instants | `from >= before` stays draft-only | relative label never becomes URL authority | READY |
| B09-06 Historical Scope | narrow to exact historical Area visibility | op87 + op78 `visibility_area_id` | select/clear exact Area | returned `area_id`; code/name presentation only | assist loading/empty/failure distinct | no Company-only invented filter | READY |
| B09-07 Actor | investigate SYSTEM or one exact USER | closed SYSTEM + op88 | select exact actor | `actor_kind` + returned `user_id` | typed-but-unselected text never filters | no all-human category / no admin directory | READY |
| B09-08 Action | investigate one or more closed Audit operation codes | closed 37-value `AuditOperationCode` | local label search + multi-select | exact enum values | local label search failure cannot affect evidence truth | no Audit full-text search / no group backend semantics | READY |
| B09-09 Resource kind | narrow evidence by admitted resource kind | op78 `resource_kind` | select/clear kind | closed wire enum | changing kind clears incompatible exact id | no generic entity ontology | READY |
| B09-10 Exact Resource | narrow to one Audit-visible resource | op89 + op78 `resource_id` | select exact result | returned `resource_kind + resource_id` | assist loading/empty/failure distinct | no universal resource directory/resolver | READY |
| B09-11 Apply | promote a valid draft question | op78 first-page traversal | read only | normalized draft stable values | first-page loading removes old ledger; failure preserves applied question | no optimistic evidence state | READY |
| B09-12 Clear editing | abandon un-applied changes | current applied query | local reset only | current applied query | no server effect | no hidden query mutation | READY |
| B09-13 applied chips | understand and remove active query dimensions | current applied query | immediate clear + op78 reread | semantic dimension identity | compound dimensions clear atomically | no wire-invalid partial actor/resource predicates | READY |
| B09-14 canonical URL | preserve/share applied evidence question | browser navigation state derived from admitted op78 predicates | History API presentation only | stable IDs/enums/UTC instants | History API failure leaves canonical evidence operable in-page | URL != Authorization / cursor authority | READY |
| B09-15 evidence ledger | scan admitted immutable evidence | op78 `AuditInspectionPage` | read/detail open only | returned event/resource/actor/visibility identities | initial failure != known-empty | no client post-filter of incomplete pages | READY |
| B09-16 human recognition | recognize actor/resource/scope without rewriting evidence | optional current recognition composed after admission | none | stable evidence ids + optional safe labels | absence falls back to neutral type + compact id | current label != historical fact | READY |
| B09-17 historical scope column | understand event-time visibility | immutable event visibility | none | COMPANY or AREA exact evidence | missing current Area name does not erase stable code/id | current ownership/membership not inferred | READY |
| B09-18 local-day separators | scan long evidence chronologically | op78 canonical order | presentation only | local rendering of `occurred_at` | crossing continuation inserts one separator | no server grouping/count semantics | READY |
| B09-19 contextual detail | inspect exact loaded evidence | already-loaded `AuditInspectionItem` | local drawer/surface only | selected `event_id` | no extra fetch is required | no `GET /audit/events/{id}` invention | READY |
| B09-20 canonical detail evidence | distinguish proof from recognition | loaded immutable evidence union | none | exact UTC/id/code/actor/resource/visibility | recognition absence has no effect | no raw-current-state reconstruction | READY |
| B09-21 typed facts | inspect closed event-specific facts | loaded closed Audit union | none | operation-specific typed facts | no fact -> explicit absence | no arbitrary JSON/schema browser | READY |
| B09-22 same actor | continue Audit-native investigation | op78 | immediate applied query + first-page reread | exact loaded actor identity | query failure preserves question | no owner navigation required | READY |
| B09-23 same resource | continue Audit-native investigation | op78 | immediate applied query + first-page reread | exact loaded kind/id | query failure preserves question | no generic deep-link resolver | READY |
| B09-24 same action | continue Audit-native investigation | op78 | immediate applied query + first-page reread | exact loaded operation code | query failure preserves question | no analytics/grouping semantics | READY |
| B09-25 Document handoff | inspect current owner context when relevant | already-admitted exact document identity | navigate to B03 boundary | exact `document_id` | destination independently handles unavailable/denied | Audit visibility never grants Document access | READY |
| B09-26 Document History handoff | inspect Controlled Documents history when admitted | exact document identity in event facts | navigate to B07 boundary | exact `document_id` | B07 rechecks its own disclosure | Audit != Document History | READY |
| B09-27 Governance handoff | inspect current Governance Case when admitted | exact `governance_attempt_id` in typed facts | navigate to B06 boundary | exact attempt id | B06 owns unavailable/denied state | Audit never becomes governance workspace | READY |
| B09-28 cursor continuation | continue canonical evidence traversal | op78 cursor page | `Carregar eventos anteriores` | opaque server cursor | loaded rows remain on continuation failure; retry | no page numbers/total/infinite-scroll authority | READY |
| B09-29 end of traversal | know no more admitted rows are available in this traversal | op78 page `has_more/next_cursor` semantics | none | cursor response | explicit end message only after successful page | no total-count inference | READY |
| B09-30 known-empty | distinguish successful zero match from failure | successful op78 page with zero items | clear filters | applied query | says no match, not no history | no hidden fallback query | READY |
| B09-31 first-page failure | recover a failed investigation | op78 Problem/transport failure | retry same applied read | applied stable query | no stale ledger under changed applied chips | no empty-success fabrication | READY |
| B09-32 Query Assist recovery | build valid exact predicates safely | op87/op88/op89 | retry assist | server-authored bounded options | loading/known-empty/failure separate | loaded ledger never becomes selector completeness | READY |
| B09-33 responsive/accessibility | preserve the same Audit meaning across viewport/input modes | same B09 truth | filter sheet, full-surface detail, focus/escape | no new Product identity | focus restoration + live status | no mobile-only business semantics | READY |

## 3. Exact operation homes used by B09

```text
78  listAuditEvents
87  listAuditQueryAreas
88  searchAuditQueryActors
89  searchAuditQueryResources
```

All four are reads. B09 introduces no business-domain write.

Owner-lens handoffs reuse already accepted frontend route boundaries; they do not add Audit API operations or transfer authority to Audit.

No operation 90+ is required.

## 4. Bidirectional trace

### Product/backend -> frontend

```text
op78 admitted structured predicates + AuditInspectionPage
-> applied investigation + ledger + detail + cursor continuation

op87 Audit-visible historical Area candidates
-> Historical Scope selector

op88 Audit-visible USER candidates
-> exact actor Query Assist

op89 Audit-visible resource candidates by kind
-> exact resource Query Assist

immutable evidence + optional current recognition
-> proof-first detail + human-readable ledger fallback
```

### Frontend -> Product/backend

```text
Apply / remove chip / Audit-native shortcut
-> exact normalized op78 first-page query

load older events
-> op78 cursor continuation

select historical Area
-> op87 candidate identity
-> op78 visibility_area_id

select USER actor
-> op88 candidate identity
-> op78 actor_kind=user + actor_user_id

select resource kind/exact resource
-> op89 candidate identity where exact lookup is used
-> op78 resource_kind + optional resource_id

owner-context action
-> existing B03/B07/B06 route boundary
-> destination rechecks its own current authority
```

## 5. Client state classes

B09 uses only the accepted four frontend state classes:

```text
SERVER STATE
  op78 pages, op87/88/89 Query Assist responses, optional recognition

NAVIGATION / URL
  /audit + applied structured query only

FORM DRAFT
  un-applied five-dimension investigation editor

EPHEMERAL UI
  open filter panel/sheet, selected row, detail drawer,
  focus restoration, retry/review-only fixture toggles
```

Cursor/load-depth is ephemeral traversal state and is not serialized into the canonical applied-query URL.

## 6. Ordering / pagination / identity mechanics

```text
op78 canonical order
  occurred_at DESC,event_id DESC

first page
  admitted structured filters + optional limit

continuation
  cursor + optional limit only

filter identity
  stable IDs/enums/UTC instants

recognition labels
  presentation only
  never cursor/filter authority
```

Multiple `operation_codes` are canonicalized into the accepted closed enum order. USER actor identity is one compound `actor_kind=user + actor_user_id` dimension. Exact resource identity is one compound `resource_kind + resource_id` dimension.

## 7. Material failure intent

```text
route denial
  stop at shell/route boundary; never render a successful empty Audit

first-page failure
  keep applied URL/chips/question; no old rows; retry

known-empty
  successful query with no matching admitted evidence; clear/refine filters

continuation failure
  preserve every loaded row and current position; retry continuation

Query Assist failure
  keep editor usable; distinguish failure from no candidates; retry assist

recognition unavailable
  keep immutable evidence; use stable/neutral identity fallback

owner destination unavailable
  destination owns current denial/unavailable semantics

History API failure
  applied in-page canonical evidence remains operable; routing architecture is not inferred
```

## 8. Access / disclosure proof

```text
query candidate                 != authorization
historical Audit visibility     != current owner access
recognition label               != event-time evidence
loaded page                     != complete selector universe
URL query                       != server admission
Audit evidence                  != current business state
Audit                           != Document History
detail drawer                   != detail API
owner handoff                   != generic deep-link resolver
```

Server-side current `audit.read` plus historical visibility filtering precedes pagination and all public evidence/query-assist disclosure.

## 9. B09 negative contract

The locked B09 contains no Product surface for:

```text
free-text generic Audit search
query DSL
saved searches
custom sort
analytics/dashboard
export
column chooser/reorder/pin
bulk row selection
comments/annotations
case management
mark reviewed/read
operational mutation
raw JSON/developer mode
generic resource/entity resolver
admin-directory browsing
all-human actor category filter
Company-only historical-scope filter
Audit detail route/endpoint
```

## 10. P9 closure

```text
material B09 regions/controls traced         33 / 33
unbound material controls                    0
invented operations                          0
operation 90+                                absent
screen-shaped APIs                           0
frontend audit.read evaluator                0
client post-filter completeness authority    0
Audit/History semantic merges                0
generic entity/deep-link resolver            0
material B09 Screen Contract findings        0
```

P9 is complete for the operator-locked B09 scope.
