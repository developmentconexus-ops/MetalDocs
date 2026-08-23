# T11 — B06 P10 Bounded Pattern Consolidation

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B06 — Governance Case.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Rule:** shared patterns graduate only from repeated LOCKED semantic behavior; geometric similarity is insufficient.

## 1. Goal

Compare B06 against already-LOCKED B01/B01N/B02/B03/B04/B05 and reuse or graduate only patterns whose protected semantics truly match.

## 2. Existing shared patterns reused

### Global App Shell — REUSE

Origin: B01.

B06 reuses the application shell, navigation posture and responsive shell transformation. It does not create a Governance-specific application frame.

### Notification Quick Inbox — REUSE

Origin: B01N.

B06 inherits global notification chrome only. Notification persistence/presentability/engagement remain outside B06.

### Exact Read-Only Content Viewer Shell — REUSE

Graduated at B04 after matching locked evidence from B03 + B04.

B06 is now a third locked consumer:

```text
B03  exact current official bytes
B04  exact DRAFT PDF / immutable Submission bytes
B06  exact immutable governed Submission or exact obsolescence target bytes
```

Shared semantic core still holds:

```text
owning lens supplies exact semantic identity
exact bytes rendered read-only
resource identity remains outside viewer internals
viewer never mutates Product lifecycle/content
integrity/dependency failure cannot become partial-success truth
caller owns exit/context/actions
responsive viewport behavior is presentation-only
```

B06 does not broaden this pattern into a generic content resolver or review/editor platform.

## 3. Similarity evaluated but NOT graduated

### B04 Content-first Work workspace vs B06 Content-first Governance Workspace

**DO NOT GRADUATE.**

Both place dominant content beside a right-side rail, but protected semantics differ:

```text
B04
  mutable/current work
  persistence/OCC/upload/submit concerns
  Work operational rail

B06
  immutable governed subject
  Step/feedback/Decision concerns
  Governance Case rail
```

A universal `ContentWorkspace` abstraction would hide owner/state differences that implementation must keep explicit.

### B04 Work rail vs B06 Governance rail

**DO NOT GRADUATE.**

Geometry is similar; content ownership, writes, failures and omission laws are not.

### B03 `allowed_actions` management affordances vs B06 `allowed_actions` Decision affordances

**DO NOT GRADUATE AS ONE ACTION ENGINE.**

Both consume server-derived hints, but each lens has a closed action vocabulary and different command/failure semantics. The shared law is architectural guidance only:

```text
hint != Authorization
command rechecks server truth
```

No generic frontend action registry/matrix is justified.

### B03 Discussion vs B06 GovernanceFeedback

**DO NOT GRADUATE.**

```text
Document Discussion = stable-Document conversation
GovernanceFeedback  = immutable exact-attempt context
```

The B06-F2 future Review Layer seam reinforces this separation.

## 4. B06-local semantic patterns retained

### Content-first Governance Workspace

```text
minimal case orientation
+ dominant exact governed content
+ B06-local governance context rail
```

Local because its purpose/state/action model is governance-specific.

### Governance subject summary

Submission and obsolescence cases expose only the exact subject identity needed to understand the Decision target. This is not a generic dossier pattern.

### Ordered Step + deadline context

```text
route-order Steps
active Step semantic current indication
optional exact due_at after activation
prior Decision context
overdue presentation without lifecycle mutation
```

Local to Governance Case until another locked surface proves the same protected process semantics.

### Governance feedback timeline/composer

Immutable attempt feedback, cursor continuation and same-logical-key ambiguous retry remain B06-local. It is not stable-Document Discussion.

### Deliberate Decision zone

```text
ACCEPT confirmation
RETURN_FOR_CHANGES mandatory reason + confirmation
singleton Decision semantics
```

Local to governance decision-making.

### Authoritative Decision reconciliation

```text
409 winner reread
403 current-authority reread
no silent changed-outcome retry
```

Do not generalize with unrelated OCC/conflict surfaces merely because they also refetch.

### Disclosure-neutral unavailable case

```text
404 absent/non-disclosable
-> same neutral terminal presentation
-> safe B05 return
```

Local until another locked lens proves the same identity/disclosure boundary and recovery path.

### Exact-content decision hold

When exact governed bytes fail to load, locked R1 keeps case context but makes the Decision surface unavailable until exact bytes return. This remains a B06-local UX rule, not a shared Authorization or viewer rule.

## 5. Future Review Layer seam is NOT a current UI pattern

`docs/decisions/governance-review-layer-seam.md` preserves future semantics only.

Do not graduate or implement:

```text
anchored comment rail
review thread
resolve/reply controls
tracked-change UI
suggestion acceptance UI
DRAFT anchor-remapping UI
```

until a future bounded promotion creates current capability and new locked evidence.

## 6. Prototype-only constructs — NEVER GRADUATE

```text
review fixture bar
clock +20h control
force next Decision 409/403
force lost feedback response
content-failure toggle
fixture-generated document pages
fixture toasts/local fake state transitions
```

These exist only to make P8 falsifiable.

## 7. Pattern vocabulary effect

Existing shared patterns reused:

```text
Global App Shell
Notification Quick Inbox
Exact Read-Only Content Viewer Shell
```

New shared semantic patterns graduated by B06:

```text
none
```

B06-local patterns retained:

```text
Content-first Governance Workspace
Governance subject summary
Ordered Step + deadline context
Governance feedback timeline/composer
Deliberate Decision zone
Authoritative Decision reconciliation
Disclosure-neutral unavailable case
Exact-content decision hold
```

## 8. Reopen / graduation triggers

A B06-local pattern may graduate only when another LOCKED block proves matching:

```text
human purpose
owner/state authority
closed action vocabulary
identity source
Authorization/disclosure posture
conflict/failure recovery
exact-content posture where relevant
responsive/accessibility transformation
```

Shared code convenience or visual similarity alone is insufficient.

## 9. P10 closure

```text
existing locked shared patterns reused          3
new shared semantic patterns graduated           0
B06-local semantic patterns retained             8
false abstractions introduced                    0
unexplained duplicate locked semantic patterns  0
```

P10 is complete for B06.
