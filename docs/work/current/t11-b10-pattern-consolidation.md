# T11 — B10 P10 Bounded Pattern Consolidation

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B10 — Organization Administration.  
> **Method:** pinned DevelopmentConexus `METHOD.md + FRONTEND-METHOD.md`.  
> **Locked P8 Git blob:** `1d1cc7d5cb42e034ab9ee71a21c96918cdcf691d`.  
> **Rule:** shared patterns graduate only from repeated LOCKED semantic/protected behavior; visual similarity is insufficient.

This file is temporary planning Evidence under `docs/work/**`. It must be absent from the eventual merge candidate/main.

## 1. Goal

Compare B10 against already-LOCKED B01/B01N/B02-B09, reuse only genuinely shared behavior, keep Organization-specific composition local, and reject abstractions that would blur Organization versus Access ownership or independent concurrency domains.

## 2. Existing shared patterns reused

### Global App Shell — REUSE

Origin: B01.

B10 reuses the operator-LOCKED application shell, global navigation hierarchy and responsive shell transformation. `Organização` remains under the accepted `Gestão` region and uses the stable `/admin/organization` Product route. B10 does not create an admin-specific global frame.

### Notification Quick Inbox — INHERITED GLOBAL CHROME

Origin: B01N.

The notification utility chrome remains present because it is part of the locked shell. B10 neither changes Notification semantics nor makes Organization actions/alerts Notification authority.

### Explicit stale-current-truth reconciliation — REUSE BEHAVIOR, NOT NEW FRAMEWORK

Earlier locked frontend blocks already establish that consequential stale state must preserve user intent, refetch canonical truth and require explicit reconciliation rather than silently overwrite. B10 applies the same protected behavior to independent Company/Profile/Binding/Eligibility/Area/Group ETag domains.

This reuse does **not** graduate a generic `ETagForm`, entity store or mutation framework as Product/frontend authority. Production implementation may share low-level primitives later if they preserve each owner's exact contract.

### Consequential action confirmation — REUSE INTERACTION LAW

Earlier locked flows already distinguish ordinary mutation from consequential irreversible/authority-changing intent. B10 uses explicit confirmation for User offboarding because the accepted transition tears down sessions/memberships/direct grants.

The semantic consequence remains User-eligibility/offboarding specific; no generic "danger action" business abstraction is created.

## 3. Existing patterns deliberately not imported

### Library discovery/search — NOT APPLICABLE

B02 proves an accepted discovery experience over official documents with real search/filter authority. B10 collections do not have accepted User/Area/Group search/filter operations.

```text
B02 search authority exists
  !=
B10 may imitate search locally
```

No search box is imported merely for visual consistency.

### My Work / Inbox queue patterns — NOT APPLICABLE

Users, Areas and Groups are current configuration collections, not actor-attention queues. B05/B08 assignment, unseen/read or triage semantics are not imported.

### Audit investigation ledger — NOT APPLICABLE

B09's structured query, evidence ledger and Query Assist are Audit-specific. B10 provider-subject lookup is bounded authentication preflight, not a reusable generic Query Assist/reference-data platform.

### Document/History/Governance content viewers — NOT APPLICABLE

B10 does not render controlled-document content, governed snapshots, semantic history or Audit evidence. Similar drawers/panels are geometry only.

## 4. B10-local semantic/composition patterns — DO NOT GRADUATE

### Organization local workspace

Status: **LOCAL B10 IA PATTERN**.

```text
/admin/organization
→ Company
→ Users
→ Areas
→ Groups
```

The four local sections are one Organization workspace because they serve one accepted administration route while preserving separate semantic/write owners beneath it. This does not become a generic tabbed-admin framework.

### Paginated collection → exact selected detail

Status: **LOCAL B10 COMPOSITION PATTERN**.

Users, Areas and Groups use cursor browse to select one stable identity and then load the exact current detail/concurrency representation needed for the task.

The pattern deliberately does not imply total count, global search, client cache completeness or one generic entity-detail API.

### Independent selected-User subresource editors

Status: **LOCAL B10 AUTHORITY-PRESERVING PATTERN**.

```text
User stable identity
├── UserProfile
├── ProviderSubjectBinding
└── UserEligibility
```

Each retains its own read/write/concurrency law. Visual co-location does not create a `UserAggregate`, one Save button or one client ETag.

### Provider-subject selection preflight

Status: **LOCAL AUTHENTICATION/ORGANIZATION COMPOSITION PATTERN**.

```text
bounded provider query
→ opaque provider_subject_ref selection
→ create/rebind command
→ Product User remains MetalDocs identity
```

This is not a generic external-directory browser, identity-sync center or provider mirroring pattern.

### Eligibility/offboarding consequence panel

Status: **LOCAL B10 SECURITY-UX PATTERN**.

Disable communicates the accepted teardown consequence; re-enable explicitly communicates non-resurrection. The frontend explains Product truth but does not execute or mirror teardown semantics.

### Area identity + lifecycle split

Status: **LOCAL B10 AUTHORITY-PRESERVING PATTERN**.

Area immutable code/current name and Area ACTIVE/RETIRED lifecycle remain separate current truths/concurrency domains. Retirement is presented as lifecycle, never delete.

### Group identity boundary to Access

Status: **LOCAL B10/B11 OWNERSHIP PATTERN**.

```text
B10 Group identity
──── explicit boundary ────
B11 GroupMembership + RoleAssignment
```

This is intentionally visible enough to prevent an operator from assuming membership controls are missing by accident, while B10 never consumes access-managing operations.

### Dependency-blocked destructive action

Status: **LOCAL B10 GROUP PATTERN**.

Group delete can fail because live semantic/security/governance dependencies protect the Group. The UI preserves the Group and explains the block; it does not expose cascade, force-delete or dependency mutation controls.

### Browse-without-fake-search

Status: **LOCAL B10 SCALE/YAGNI PATTERN**.

The locked P8 proves cursor browse without pretending a loaded page is a complete search universe. B10-A1 is validated for current Launch and remains reopenable only on real scale/findability Evidence.

This must not graduate into a general rule that all future admin collections should omit search.

## 5. Similarity explicitly rejected as insufficient

```text
B02 list/search vs B10 list/browse
  -> official-document discovery authority != Organization collection authority

B08 inbox rows vs B10 User/Area/Group rows
  -> personal attention lifecycle != configuration identity

B09 Query Assist vs B10 provider lookup
  -> Audit-visible predicate construction != authentication subject preflight

repeated list/detail geometry
  -> does not create generic EntityListDetail Product pattern

repeated ETag forms
  -> does not create one client concurrency owner

Company/User/Area/Group CRUD-like verbs
  -> does not create generic CRUD semantic platform
```

## 6. Prototype-only constructs — NEVER GRADUATE

```text
Evidence scenario buttons for 403/404/412/422/503
fixture Users/Areas/Groups and page sizes
fixture provider subject matches
fixture Idempotency-Key ambiguity message
forced Group dependency conflict
review-only B10-A1 Rafael Siqueira locate task
local simulated latency/result messages
```

They exist only to make P8 falsifiable.

## 7. Pattern vocabulary effect

Existing locked shared behavior reused/inherited:

```text
Global App Shell
Notification Quick Inbox global chrome
explicit stale-current-truth reconciliation behavior
consequential-action confirmation interaction law
```

New shared semantic patterns graduated by B10:

```text
none
```

B10-local semantic/composition patterns retained:

```text
Organization local workspace
paginated collection → exact selected detail
independent selected-User subresource editors
provider-subject selection preflight
eligibility/offboarding consequence panel
Area identity + lifecycle split
Group identity boundary to Access
dependency-blocked Group delete
browse-without-fake-search
```

## 8. Anti-abstraction decisions

Rejected:

```text
Generic Admin Center framework
Generic CRUD framework
Generic Entity list/detail model
Generic Search/Filter control over incomplete pages
Generic Reference Data / Directory service
Generic ETag Form / aggregate Save authority
Generic User aggregate/client entity graph
Generic Lifecycle editor
Generic destructive-action business engine
Generic Permission/Authorization-aware frontend router
provider synchronization/mirroring framework
```

Production implementation may reuse domain-agnostic UI primitives where useful. P10 freezes no cross-Product semantic/component framework from B10 alone.

## 9. Reopen / graduation triggers

A B10-local pattern may graduate only after another LOCKED block proves materially matching:

```text
human job
semantic owner
stable identity source
read/write boundary
concurrency domain
access/disclosure posture
failure/recovery class
pagination/findability law
responsive/accessibility meaning
```

B10-A1 specifically reopens on material scale/findability Evidence; future B11/B12 UX must not inherit browse sufficiency merely because B10 accepted it.

## 10. P10 closure

```text
existing locked shared behaviors reused/inherited  4
new shared semantic patterns graduated              0
B10-local semantic/composition patterns             9
false abstractions introduced                       0
Organization/Access ownership merges                0
independent concurrency-domain collapses             0
generic admin/CRUD/search frameworks                0
```

P10 is complete for the operator-locked B10 scope.
