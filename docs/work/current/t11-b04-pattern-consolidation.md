# T11 — B04 P10 Bounded Pattern Consolidation

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B04 — Document Work / Authoring.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Rule:** shared patterns are derived from LOCKED semantic evidence; cosmetic similarity is insufficient.

## 1. Goal

Compare B04 against already-LOCKED B01/B01N/B02/B03 and graduate only patterns whose semantics now have at least two locked consumers.

## 2. Existing locked patterns reused

### Global application shell — REUSE

Origin:

```text
B01 App Shell + Global IA
```

B04 reuses the locked application header/sidebar/navigation posture and narrow-screen navigation transformation. No B04-specific shell variant is created.

### Notification Quick Inbox — REUSE

Origin:

```text
B01N Notification chrome + Quick Inbox
```

B04 inherits the global bell/Quick Inbox chrome only. It does not create a Work-specific Notification store, badge rule or source-disclosure rule.

## 3. Pattern graduated by the second locked consumer

### Exact read-only content viewer shell — GRADUATE TO SHARED PATTERN

Prior evidence:

```text
B03 P10
  Exact official viewer shell
  = LOCAL CANDIDATE because no second LOCKED consumer existed
```

B04 now supplies the second locked consumer:

```text
B03
  exact official source / OfficialRendition
  read-only
  owning Document context outside rendered bytes
  exact-content integrity/dependency failure must not become partial success

B04
  exact DRAFT PDF source or exact immutable Submission source
  read-only
  owning Work/Revision context outside rendered bytes
  exact-content integrity/dependency failure must not become partial success
```

Shared semantic core:

```text
one owning lens provides semantic identity/context
one exact-byte resource is rendered read-only
resource identity/authority remains outside the viewer component
viewer never mutates lifecycle/content
viewer never infers Product status
integrity failure cannot finish as successful partial content
explicit exit/back remains owned by the calling lens
responsive viewport behavior is presentation-only
```

Graduated vocabulary:

```text
ExactReadOnlyContentViewerShell
```

This is a planning/design pattern only. P10 does not authorize a production component or package yet.

The shared shell must be parameterized by the owning lens's already-authorized exact-resource loader/identity; it must not become a generic content resolver, BFF, storage adapter, lifecycle switch or provider URL parser.

B03's older `Exact official viewer shell` local-candidate disposition is therefore superseded **only at the pattern-vocabulary level** by this second-consumer evidence. B03 Product/UX LOCK remains unchanged.

## 4. B04-local patterns — DO NOT GRADUATE YET

### Content-first Work workspace

Status: **LOCAL B04 PATTERN**.

```text
minimal MetalDocs Work header
+ dominant editor/viewer canvas
+ right operational rail
```

No second locked block has the same authoring/current-work semantics. Do not generalize into a universal workspace layout.

### Eigenpal editable DOCX boundary

Status: **LOCAL B04 PATTERN**.

Semantics:

```text
DOCX DRAFT only
vendor/editor ergonomics inside Eigenpal boundary
MetalDocs owns persistence orchestration + Product state presentation
```

Do not create a generic editor abstraction merely for future imagined formats.

### Hybrid DRAFT persistence/status controller

Status: **LOCAL B04 PATTERN**.

Semantics include:

```text
local FORM DRAFT
coalesced background save
Salvar agora / Ctrl+S
one pipeline in flight
strong DRAFT ETag
submit force-flush
```

This is not a generic autosave hook until another locked editable resource proves the same OCC/failure/flush contract.

### Direct upload/admission progression

Status: **LOCAL B04 PATTERN**.

```text
allocation -> provider PUT -> server admission/READY -> attach under If-Match
```

The transport/content-integrity mechanism may be reused elsewhere in implementation, but frontend Product pattern graduation requires another locked UX consumer with the same user-visible progression and recovery semantics.

### DRAFT OCC reconciliation surface

Status: **LOCAL B04 PATTERN**.

```text
412 draft_changed
-> preserve local human input
-> refetch authoritative DRAFT
-> explicit reconciliation
-> no automatic merge
```

Do not collapse this with B03 responsible-owner 412 UX: both are OCC, but B04 preserves a potentially rich local content buffer while the owner flow preserves a selected User target. Their recovery semantics are materially different.

### Work operational rail

Status: **LOCAL B04 PATTERN**.

```text
Trabalho atual
Fonte
Ações
Contexto do documento collapsed
```

It is not a generic right sidebar; its content ownership and omission rules are Work-specific.

### SUBMITTED gate summary inside Work

Status: **LOCAL B04 PATTERN**.

It is orientation for an immutable Submission while remaining in the Work lens. It must not graduate into Governance UI or approval timeline.

### No-current-work terminal state

Status: **LOCAL B04 PATTERN**.

Semantics:

```text
current Work no longer resolvable
-> explicit B03 return
-> never History fallback
```

No generic missing-resource router abstraction is justified.

## 5. Similarity explicitly rejected as insufficient

```text
B03 management rail-like regions vs B04 Work rail
  -> different owners/purposes; no shared sidebar abstraction

B03 responsible-owner 412 vs B04 DRAFT 412
  -> same HTTP class, materially different preserved local intent/recovery

B03 allowed_actions vs B04 action group
  -> B04 has no equivalent DocumentWorkView allowed_actions contract; no generic action-menu authority

B03 Discussion composer vs Eigenpal editor
  -> both accept text, entirely different semantic objects/invariants

B04 upload progress vs generic async progress
  -> exact READY/attach truth ladder is domain-specific
```

## 6. Prototype-only constructs — NEVER GRADUATE

```text
review fixture bar
force-next-save-412 control
force-next-save-failure control
force-next-upload-expiry control
fixture toasts
local fake upload timing
local fake ETag increment logic
transition stub dialogs
```

These exist only to make P8 falsifiable/operable.

## 7. Pattern vocabulary effect

Existing shared patterns reused:

```text
Global App Shell
Notification Quick Inbox
```

New shared pattern graduated by B04:

```text
Exact Read-Only Content Viewer Shell
```

B04-local patterns retained without abstraction:

```text
Content-first Work workspace
Eigenpal editable DOCX boundary
Hybrid DRAFT persistence/status controller
Direct upload/admission progression
DRAFT OCC reconciliation surface
Work operational rail
SUBMITTED gate summary
No-current-work terminal state
```

## 8. Reopen / graduation triggers

A B04-local pattern may graduate only when another LOCKED block proves matching:

```text
purpose
state ownership
protected semantics
Authorization/disclosure posture
OCC/failure/recovery class
exact-content identity posture where relevant
accessibility behavior
responsive transformation
```

Shared code convenience alone is insufficient.

## 9. P10 closure

```text
existing locked shared patterns reused          2
new shared semantic pattern graduated           1
B04-local semantic patterns retained            8
false abstractions introduced                   0
unexplained duplicate locked semantic patterns  0
```

P10 is complete for B04.
