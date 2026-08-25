# T11 — B11 Access Administration — P10 Pattern Consolidation

> **Status:** COMPLETE / PASS / R3 POST-LOCK CONSOLIDATION.  
> **Current P8 locked package:** `docs/work/current/t11-b11-access-administration-p8.html`.  
> **Exact R3 locked blobs:** HTML `ea20912e5259f4f3f51df7ce09ee3f2e5cfc7540`; CSS `9ce012007613777187ae70956c2bfa09e7066c16`; JavaScript `670ff9b905d94014ff27698e2a23c868316030a4`.  
> **Operator LOCK:** `docs/work/current/t11-b11-p8-r3-operator-relock.md`.  
> **P9:** `docs/work/current/t11-b11-screen-contract.md` — COMPLETE / PASS.  
> **Implementation:** BLOCKED.

## 1. Consolidation rule

A shared frontend pattern graduates only when repeated operator-LOCKED protected semantics prove the same behavior. Cosmetic similarity, repeated boxes/lists/dialogs, or implementation convenience are insufficient.

B11 therefore reuses accepted cross-product interaction laws where the invariant genuinely repeats and keeps Access-specific composition local. P10 creates no speculative design system, IAM framework, Admin Center framework or production component architecture.

## 2. Existing shared protected behavior reused

### P10-S1 — Global App Shell / primary IA

Source: B01 LOCK.

B11 reuses the accepted shell, global navigation grouping, `/admin/access` entry, responsive shell behavior and content-region relationship. It creates no alternate administrative frame.

### P10-S2 — Notification chrome / Quick Inbox

Source: B01N LOCK.

B11 inherits the utility-header notification affordance. It does not redefine Notification state, routes, read/unread behavior or ownership.

### P10-S3 — Visible server-owned cursor traversal

Repeated across prior locked bounded collections and B11:

```text
render exactly the returned page
+ show whether continuation exists
+ continuation uses returned cursor
+ failed continuation preserves the current page
+ no invented total, offset or hidden complete-universe claim
```

B11 adds no reverse-cursor assumption. Previous revisits only already visited pages through bounded client navigation state. This is shared interaction/protection behavior, not a universal pager API and not permission to reuse operation filters on continuation.

### P10-S4 — Deliberate consequential-action confirmation

Repeated across prior locked administrative/governance blocks and B11:

```text
exact target identity
+ bounded consequence
+ residual-effect caveat where material
+ deliberate confirm / cancel
```

B11 applies it to security-bearing membership add/remove and exact RoleAssignment revoke. The pattern does not imply identical copy or a generic business-level “danger action” component.

### P10-S5 — Ambiguous idempotent-create recovery

Existing global wire/idempotency law reused by B11 op32:

```text
ambiguous transport outcome
→ preserve one logical command
→ retry same normalized fingerprint with same Idempotency-Key
→ recover exact stored success when committed
→ never silently create a second command
```

The clean P8 additionally proves the completed replay: same status/body/assignment identity and semantic mutation count 1→1.

### P10-S6 — Disclosure-safe failure distinction

Existing global request/disclosure behavior reused:

```text
403 denial != successful empty collection
404 absent/non-disclosable != proof of hidden existence
correctable command failure != erased draft
continuation failure != erased canonical page
```

This is shared safe communication behavior. It does not create a generic error-message authority divorced from operation-local contracts.

### P10-S7 — Responsive and keyboard semantic preservation

Existing locked structural expectation reused:

```text
same information identity
+ same consequence before confirmation
+ same available recovery
+ keyboard-operable navigation/dialogs
+ narrow layout may reflow, never change meaning
```

Repeated responsive geometry may later share low-level UI primitives, but P10 does not graduate one Product-semantic list/detail/dialog framework.

## 3. B11-local protected patterns — do not graduate

The following remain **B11 LOCAL**:

```text
L1   Por Área / Grupos / Funções access lenses
L2   Area-specific and Company-wide grants in separate labeled regions
L3   Group direct-grant footprint across Company and multiple Areas
L4   Group access footprint visible before security-bearing membership mutation
L5   fixed Role meaning read-only from RoleView
L6   contextual exact Area/Company grant preselection
L7   contextual exact Group-subject grant preselection
L8   op31 filtered canonical slices for Area, Company, Group and Role
L9   access configuration versus complete per-User effective-access boundary
L10  raw op6 UserPage picker with disabled-state guidance and no eligibility filter
L11  unknown membership reconciliation through op28 201/204
L12  initial success / completed replay / ambiguous recovery shown as distinct grant states
```

Each is justified by Access Administration semantics. No second LOCKED consumer currently proves the same full human job, owner, identity, read/write, disclosure and recovery invariant.

## 4. Similarity deliberately rejected

### Generic admin CRUD or entity manager

```text
B10
  Organization identity/lifecycle + independent concurrency domains

B11
  Authorization configuration slices + security-bearing memberships/grants
```

List/detail geometry does not justify `AdminEntityManager`, a generic CRUD domain model, one aggregate Save authority, or one client entity graph.

### Generic IAM wizard / permission matrix

B11 composes one exact `Subject(User|Group) × fixed Role × Scope(Company|Area)` command. It does not justify a configurable IAM workflow, custom roles, permission matrix, access graph, bulk assignment engine, certification campaign or effective-access simulator.

### Generic “eligible entity” picker

The picker deliberately renders raw op6 page/eligibility truth. It does not precompute membership completeness or client eligibility. A generic picker that silently filters disabled/already-related/unknown entries would destroy B11's protected 201/204 reconciliation and is rejected.

### Universal pager

A visual Previous/Next control cannot become an abstraction that assumes totals, offsets, reverse cursors, repeated first-page filters or interchangeable cursor semantics. Shared low-level rendering is allowed later only if each operation's exact first-page/continuation law remains explicit.

### Search / discovery

B02 search authority does not transfer to B11. No accepted global User/Group/Area search or global assignment matrix exists. Server-filtered op31 traversal is canonical slicing, not Library discovery.

### Queue / triage / Query Assist

B05/B08 attention lifecycle and B09 Audit query assistance do not transfer. Access assignments are configuration truth, not personal work, Notification state or a query DSL.

## 5. Organization / Authorization ownership boundary

B11 consumes Organization reads but does not absorb B10 ownership:

```text
Organization
  User / Area / Group identity + lifecycle

Access Administration
  GroupMembership security-bearing mutation
  RoleAssignment configuration
```

The raw User/Group/Area pages remain Organization truth. RoleAssignment and access-managing commands remain Authorization truth. UI adjacency and contextual preselection merge neither owner.

## 6. Prototype-only constructs

The following are Evidence mechanics only and do not become production pattern authority:

```text
review fixture bar and deterministic scenario selector
fixture arrays and in-browser simulated server filtering/pagination
fake cursors, response delays and mutation counters
forced op22/op31 continuation failures
Sofia op28 204 and Mariana op28 201 fixtures
ordinary completed-replay and ambiguous post-commit fixtures
hard-coded low-fi styling
prototype-only dialog/drawer implementation
```

Production implementation later realizes the LOCKED semantics through the accepted architecture. It does not port the P8 simulator or fixture data.

## 7. Pattern vocabulary effect

Existing shared protected behavior reused:

```text
Global App Shell
Notification Quick Inbox
visible server-owned cursor traversal
consequential-action confirmation
ambiguous idempotent-command recovery
disclosure-safe failure distinction
responsive/keyboard semantic preservation
```

New cross-product semantic patterns graduated by B11:

```text
none
```

B11-specific Product/UX language retained locally:

```text
Por Área
Grupos
Funções
access footprint
Company-wide grants
Area-specific grants
membership 201/204 reconciliation
```

## 8. False abstractions rejected

P10 explicitly rejects graduating or preparing:

```text
generic IAM console or IAM wizard
generic access graph / permission matrix
custom Role / Permission builder
generic Admin Center / CRUD / entity-manager framework
global subject/reference directory for frontend convenience
eligibility-filtered User picker abstraction
universal pager with totals/offset/reverse-cursor assumptions
access search framework without proven need
Group-to-Area ownership abstraction
bulk access mutation framework
access certification/review campaign framework
permission simulator / effective-access engine
atomic EditGrant abstraction over delete + create
generic replay store owned by presentation code
```

Low-level domain-agnostic UI and data-fetching primitives may be shared at implementation time only if they preserve every accepted owner, identity, cursor, failure and replay contract. Implementation remains blocked.

## 9. Global Maximum / anti-abstraction check

Global Maximum for this P10 is the smallest reusable set that preserves all proven semantics:

```text
reuse seven already-proven shared behaviors
+ keep twelve Access-specific protected behaviors local
+ introduce zero new semantic framework layers
+ preserve all four clean-rebaseline corrections
```

A reusable IAM/admin architecture now would optimize for hypothetical screens and erase known boundaries. The current structure can graduate a pattern later without dismantling authority if another operator-LOCKED block proves the same protected invariant.

## 10. Reopen / graduation triggers

A B11-local pattern may be reconsidered only when:

- another operator-LOCKED block proves the same human job and protected semantics;
- a named Product requirement adds a second real consumer;
- P11 integration reveals duplicate semantic behavior that cannot remain local coherently.

B11 itself reopens on material Evidence such as:

- a real per-User effective-access/troubleshooting job;
- a real scale/findability failure requiring new read/search precision;
- an access certification/compliance review requirement;
- a changed Role/Permission model;
- a proven organizational ownership need for Groups;
- a P11 assembled-fidelity contradiction.

## 11. Closure

```text
existing shared protected behaviors reused       7
new shared semantic patterns graduated           0
B11-local protected patterns                    12
clean-rebaseline corrections preserved           4 / 4
false abstractions introduced                     0
Organization/Authorization ownership merges      0
generic IAM/Admin frameworks                      0
prototype constructs promoted to authority       0
unresolved P10 findings                           0
```

**P10 status: COMPLETE / PASS.** R3 adds no shared pattern. It tightens the already protected ambiguous-command recovery law locally, preserves the seven accepted shared behaviors, and introduces no speculative IAM, Authorization or Admin framework.
