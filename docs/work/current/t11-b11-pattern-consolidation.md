# T11 — B11 Access Administration — P10 Pattern Consolidation

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Locked P8:** `docs/work/current/t11-b11-access-administration-p8-r5.html`  
> **Locked blob:** `96094773435a88c357e308779639415d9853b327`  
> **P9:** `docs/work/current/t11-b11-screen-contract.md` — PASS.  
> **Implementation:** BLOCKED.

## 1. Consolidation rule

A shared frontend pattern graduates only when repeated **LOCKED protected semantics** prove the same behavior. Cosmetic similarity, repeated boxes/lists, or implementation convenience are insufficient.

B11 therefore reuses already-accepted global/interaction behavior where semantics genuinely match and keeps Access-specific structure local until another LOCKED block proves the same invariant.

## 2. Existing shared patterns reused

### P10-S1 — Global App Shell / primary IA

Source: B01 LOCK.

B11 reuses the accepted shell, global navigation grouping, `/admin/access` entry, responsive shell behavior and content-region relationship. B11 creates no alternate admin shell.

### P10-S2 — Global Notification chrome / Quick Inbox

Source: B01N LOCK.

B11 inherits the utility-header notification affordance and does not redefine Notification state, routing or interaction.

### P10-S3 — Deliberate consequential-action confirmation

Repeated protected semantic behavior across prior LOCKED administrative/governance blocks and B11:

```text
exact target identity
+ bounded consequence
+ deliberate confirm/cancel
```

B11 applies it to membership add/remove and exact RoleAssignment revoke. The pattern does not imply a generic confirmation component or identical copy.

### P10-S4 — Ambiguous idempotent-create recovery

Existing global wire/idempotency law reused by B11 op32:

```text
ambiguous transport outcome
→ preserve same logical command
→ retry with same Idempotency-Key
→ never silently create a second command
```

This is shared command-recovery semantics, not an Access-specific abstraction.

## 3. B11-local protected patterns — do not graduate yet

The following remain **B11 LOCAL**:

```text
L1  Por Área / Grupos / Funções access lenses
L2  Area-specific + Company-wide grants as separate regions
L3  Group access footprint across Company + multiple Areas
L4  Group footprint before security-bearing membership mutation
L5  fixed Role meaning shown read-only from RoleView
L6  contextual Area grant preselection
L7  contextual Group-subject grant preselection
L8  filtered canonical RoleAssignment slice pagination
L9  configuration-inspection vs per-User effective-access boundary
```

Each is justified by Access Administration semantics. No second LOCKED consumer currently proves these are reusable cross-block patterns.

## 4. Similarity deliberately rejected

### Generic admin list/detail

B10 and B11 both use list/detail-like low-fi regions, but their protected semantics differ:

```text
B10
  Organization identity/lifecycle + independent concurrency domains

B11
  Authorization configuration slices + security-bearing memberships/grants
```

Visual similarity does not justify a shared `AdminEntityManager` or generic CRUD abstraction.

### Search / discovery patterns

B02 search/discovery does not transfer to B11. B11 has no accepted global User/Group/Area search authority. Filtered op31 traversal is server-owned canonical slicing, not Library search.

### Work-queue / triage patterns

B05/B08 queue semantics do not transfer. Access assignments are configuration truth, not personal work/attention lifecycle.

### Audit Query Assist

B09 query-assist semantics do not transfer. Access does not gain a generic query DSL/reference-data/search framework.

## 5. B10 ownership boundaries preserved

B11 consumes User/Area/Group supporting reads but does not absorb B10 Organization semantics.

```text
Organization
  User / Area / Group identity + lifecycle

Access Administration
  GroupMembership mutation
  RoleAssignment configuration
```

Group identity and Group access footprint may be adjacent in UI without merging their owners.

## 6. False abstractions rejected

P10 explicitly rejects graduating or preparing:

```text
generic IAM console framework
generic access graph
generic permission matrix
custom Role/Permission builder
generic Admin Center framework
generic Entity list/detail manager
global subject/reference directory for UI convenience
access search framework without proven need
Group-to-Area ownership abstraction
bulk access mutation framework
access certification/review campaign framework
permission simulator/effective-access engine
atomic EditGrant abstraction over delete + create
```

These would add accidental complexity or duplicate authority without a current evidenced consumer.

## 7. Prototype-only constructs

The following are Evidence mechanics only and do not become production pattern authority:

```text
review fixture bar
one-shot fake 403/404/409/page-failure controls
fixture arrays / local mutation simulation
hard-coded low-fi grayscale styling
browser-only fake cursors/pages
prototype dialog implementation
```

Production implementation later realizes the LOCKED interaction semantics using the accepted implementation architecture; it does not port P8 code.

## 8. Shared vocabulary effect

No new cross-product frontend pattern name is required by B11.

Current useful protected vocabulary remains:

```text
App Shell
Quick Inbox
consequential confirmation
idempotent ambiguous-command retry
server-owned pagination
```

B11-specific terms remain Product/UX language, not reusable component architecture:

```text
Por Área
Grupos
Funções
access footprint
Company-wide grants
Area-specific grants
```

## 9. Anti-abstraction / Global Maximum check

Global Maximum for P10 is the smallest reusable set that preserves proven semantics.

```text
reuse four already-proven shared behaviors
+ keep nine B11 access semantics local
+ create zero new framework layers
```

Creating a reusable IAM/admin architecture now would optimize for hypothetical future screens rather than the accepted Product. The current structure can later graduate a pattern without dismantling authority if another LOCKED block proves the same semantics.

## 10. Reopen / graduation triggers

A B11-local pattern may be reconsidered when:

- another operator-LOCKED block proves the same protected semantics;
- a named Product requirement adds a second real consumer;
- P11 integration reveals duplicated semantic behavior that cannot stay local coherently.

B11 itself reopens on material Evidence such as:

- a real per-User effective-access/troubleshooting job;
- real scale/findability failure requiring search/read precision;
- access certification/compliance review requirement;
- changed Role/Permission model;
- proven organizational ownership need for Groups;
- P11 assembled-fidelity contradiction.

## 11. Closure

```text
existing shared protected behaviors reused      4
new shared semantic patterns graduated           0
B11-local protected patterns                     9
false abstractions introduced                    0
Organization/Authorization ownership merges      0
generic IAM frameworks                           0
prototype constructs promoted to authority       0
unresolved P10 findings                          0
```

**P10 verdict:** PASS. B11 remains a bounded Access Administration experience over existing shared shell/recovery semantics; no speculative frontend/IAM framework is justified.
