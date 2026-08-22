# T11 — MetalDocs Functional Wireframe Pack

> **TEMPORARY T11 CANDIDATE WORK / BRANCH-ONLY.** F5 realizes the closed F1→F4 frontend-readiness pack as functional low/mid-fidelity wireframes. These are interaction contracts, not visual-brand authority. Product/API/backend meaning remains in accepted owners.

## 1. Wireframe law

The wireframes freeze:

```text
screen hierarchy
material regions
material data labels
actions and their placement/context
forms/dialogs/drawers carrying semantic writes
navigation destinations
read-only vs actionable regions
material lifecycle/absence/denial/conflict/recovery states
editor/viewer mode
```

They deliberately do **not** freeze:

```text
final color palette
exact spacing/radius/shadow
font family beyond repository/product standards
micro-animation
non-material loading skeleton shape
specific component-library implementation
```

Reference layout is desktop-first because the Launch interaction density includes document editing and administration. Responsive reflow may stack/collapse regions without changing semantic state, action availability or data authority.

All controls below consume only the F3/F4 operation/identity graph. A label such as `[Save]` does not imply client authority; the owning server operation remains authoritative.

## 2. Global shell

### WF-00 — Unauthenticated / session gate — APP-01

```text
┌──────────────────────────────────────────────────────────────────────┐
│ MetalDocs                                                            │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│                    Controlled documents                              │
│                                                                      │
│             Your MetalDocs session is not active.                    │
│                                                                      │
│                         [ Sign in ]                                   │
│                                                                      │
│  Authentication is provided by the configured identity provider.     │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

Material variants:

```text
401 / no session
  show Sign in

OIDC/provider unavailable
  show sanitized "Sign in temporarily unavailable"
  [Try again]
  never show local-password fallback

callback rejected/invalid
  show sanitized authentication failure
  [Return to sign in]
```

### WF-01 — Authenticated application shell — APP-02

```text
┌───────────────┬──────────────────────────────────────────────────────┐
│ MetalDocs     │ Page title / route context                User ▾    │
│               ├──────────────────────────────────────────────────────┤
│ Library       │                                                      │
│ My Work       │                  ROUTE CONTENT                       │
│ Audit         │                                                      │
│ Administration│                                                      │
│               │                                                      │
│               │                                                      │
│               │                                                      │
│               │                                                      │
│               │                                                      │
│               │                                                      │
│               │                                                      │
│               │                                                      │
│               │                                                      │
│               │                                                      │
│ [Sign out]    │                                                      │
└───────────────┴──────────────────────────────────────────────────────┘
```

Navigation presence is not Authorization. Direct route loads may return a denied/not-found state.

Global material route states:

```text
401 → WF-00
403 route read → keep shell + denied content panel
404 resource route → keep shell + "Not found or unavailable" panel
```

## 3. Library — `/documents`

### WF-02 — Library discovery — LIB-01

```text
┌───────────────┬──────────────────────────────────────────────────────┐
│ shell         │ Library                              [New document]  │
│               ├──────────────────────────────────────────────────────┤
│               │ [ Search code or title.......................... ]  │
│               │ Status: [Effective ▾]                              │
│               │ Active filters: [Area ×] [Type ×] [Owner ×]       │
│               │                                                      │
│               │ ┌──────┬──────────────────┬──────┬──────┬────────┐ │
│               │ │Code  │Title             │Type  │Area  │Owner   │ │
│               │ ├──────┼──────────────────┼──────┼──────┼────────┤ │
│               │ │PO-001│Purchasing policy │PO    │ADM   │Ana     │ │
│               │ │...   │...               │...   │...   │...     │ │
│               │ └──────┴──────────────────┴──────┴──────┴────────┘ │
│               │                                                      │
│               │                      [Load more]                     │
└───────────────┴──────────────────────────────────────────────────────┘
```

Interaction law:

```text
row click
  → WF-04 Document Official using returned document_id

click displayed Type / Area / Owner reference
  → apply exact returned id as Library filter

no global ordinary-reader User/Area/DocumentType directory is invented

empty
  "No documents match this view"
  not "no documents exist"

obsolete/cancelled status option
  server may 403; UI does not infer entitlement
```

### WF-03 — Create Document dialog/flow — LIB-02

```text
┌──────────────────────────── New document ────────────────────────────┐
│ Document type   [ Select admitted type ▾ ]                           │
│ Area            [ Select admitted area ▾ ]                           │
│ Title           [..................................................] │
│ Template        [ None / admitted template ▾ ]                       │
│ Responsible     [ Current user / admitted candidate ▾ ]              │
│                                                                      │
│ Number preview  PO-ADM-042                                           │
│                 Preview only — final number is assigned on create.   │
│                                                                      │
│ [Cancel]                                             [Create]         │
└──────────────────────────────────────────────────────────────────────┘
```

Dynamic behavior:

```text
Area/Type change
→ refetch DocumentCreationOptionsView scoped to selections
→ refresh Template/Responsible candidate controls
→ get numbering preview when meaningful

ambiguous create transport outcome
→ show "Result not confirmed"
→ [Retry same request] uses same Idempotency-Key

success
→ navigate WF-07/08 Document Work, not Document Official
→ direct target resolver rereads getDocument.open_revision
```

## 4. Document Official — `/documents/:document_id`

### WF-04 — Official overview + official content — OFF-01/OFF-02/OFF-04

```text
┌───────────────┬──────────────────────────────────────────────────────┐
│ shell         │ PO-ADM-042                              EFFECTIVE    │
│               │ Purchasing policy                                  │
│               │ Type: PO   Area: ADM   Owner: Ana                  │
│               │                                                      │
│               │ [Open work / Create revision] [History] [Manage ▾] │
│               ├──────────────────────────────────────────────────────┤
│               │ Official revision  REV003                           │
│               │ Released 2026-08-...                                │
│               │                                                      │
│               │ ┌──────────────────────────────────────────────────┐ │
│               │ │                                                  │ │
│               │ │        READ-ONLY OFFICIAL VIEWER                │ │
│               │ │        DOCX or PDF per representation           │ │
│               │ │                                                  │ │
│               │ └──────────────────────────────────────────────────┘ │
│               │                                                      │
│               │ [Open exact source]  (when separately applicable)  │
└───────────────┴──────────────────────────────────────────────────────┘
```

State variants:

```text
status=effective
  official viewer primary

status=obsolete
  retained last official + clear OBSOLETE label

status=draft|submitted before first Release
  no official viewer
  show current document status only
  disclosed open work action only when server reference exists/admitted

status=cancelled before first Release
  no official viewer

exact-content integrity/dependency failure
  viewer region becomes error panel
  no partial bytes rendered as success
```

`[Open work / Create revision]` behavior:

```text
open_revision disclosed
  label Open work
  → Document Work

open_revision absent + create action attempted/admitted
  createDocumentRevision
  → Document Work after current resolver reread

absence never proves no work to an unauthorized caller
```

### WF-05 — Manage Document drawer — OFF-03/OFF-05

```text
┌────────────────────────── Manage document ───────────────────────────┐
│ [Responsible owner] [Obsolescence]                                  │
├──────────────────────────────────────────────────────────────────────┤
│ Responsible owner                                                    │
│ Current: Ana                                                         │
│ New: [ complete eligible candidate list ▾ ]                          │
│                                                                      │
│ [Cancel]                                             [Change owner]   │
└──────────────────────────────────────────────────────────────────────┘
```

Candidate source:

```text
DocumentOfficialView.responsible_owner_candidates?
→ present only when current document.owner.manage ALLOW
→ complete existing + same Company + ENABLED Users
```

Concurrency source remains separate:

```text
getDocumentResponsibleOwner → exact ETag
```

OCC variant:

```text
┌──────────────────── Responsible owner changed ───────────────────────┐
│ The current responsible owner changed while you were editing.        │
│                                                                      │
│ Current server value: Maria                                          │
│ Your selected value:  João                                           │
│                                                                      │
│ [Use current]               [Review selection against current]       │
└──────────────────────────────────────────────────────────────────────┘
```

No automatic overwrite.

Obsolescence tab:

```text
no active request
  Reason [..........................................................]
  [Start obsolescence]

active human-governed request
  State: Awaiting governance
  Reason: ...
  [Withdraw request] when currently admitted

synchronous NoHumanApproval result
  request completes; Document refetch shows OBSOLETE

returned/withdrawn/completed
  read-only result state; no fake pending Step
```

## 5. Document Work — `/documents/:document_id/work`

### WF-06 — Work route resolving state

```text
route entry
→ getDocument
→ open_revision.state

DRAFT      → WF-07 / WF-08 / WF-09
SUBMITTED  → WF-10
ABSENT     → "No current work is available" + [Back to Document]
```

No History fallback.

### WF-07 — DRAFT DOCX authoring — DW-01/DW-03

```text
┌───────────────┬──────────────────────────────────────────────────────┐
│ shell         │ PO-ADM-042 / REV004                     DRAFT       │
│               │ [Back to official]                                  │
│               ├──────────────────────────────────────────────────────┤
│               │ Title [Purchasing policy..........................] │
│               │                                                      │
│               │ ┌──────────────────────────────────────────────────┐ │
│               │ │                                                  │ │
│               │ │             EDITABLE DOCX                       │ │
│               │ │             exact DRAFT source                  │ │
│               │ │                                                  │ │
│               │ └──────────────────────────────────────────────────┘ │
│               │                                                      │
│               │ [Replace source]    [Save]    [Submit] [Cancel rev] │
└───────────────┴──────────────────────────────────────────────────────┘
```

Material local states:

```text
clean
unsaved title/editor bytes
saving
saved → authoritative DocumentWorkView + new ETag
```

Submit always uses current semantic DRAFT ETag + one logical Idempotency-Key.

### WF-08 — DRAFT PDF / source replacement — DW-01/DW-02

```text
┌──────────────────────────────────────────────────────────────────────┐
│ PO-ADM-042 / REV004                                      DRAFT      │
│                                                                      │
│ ┌──────────────────────────────────────────────────────────────────┐ │
│ │                    READ-ONLY PDF                                │ │
│ └──────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│ [Choose replacement file]                                           │
│                                                                      │
│ Upload:  ███████████░░░  78%                                        │
│                                                                      │
│ [Cancel upload]                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

Material progression shown to user:

```text
local file chosen
→ provider upload in progress
→ uploaded, verifying/admitting
→ admitted, attaching to current DRAFT
→ authoritative DRAFT updated
```

Never label provider PUT as "Saved".

Expired capability:

```text
Upload expired before it could be attached.
Your selected local file is still available.
[Re-upload]
```

`[Re-upload]` creates a new allocation and reuses the same intended local bytes; it never revives the old upload id.

### WF-09 — DRAFT OCC reconciliation — X-04

```text
┌──────────────────────── Document changed ────────────────────────────┐
│ A newer DRAFT exists on the server. Your local edits were preserved.│
│                                                                      │
│ CURRENT SERVER                         YOUR LOCAL EDITS               │
│ ┌──────────────────────┐              ┌───────────────────────────┐ │
│ │ title/source summary │              │ title/source summary      │ │
│ └──────────────────────┘              └───────────────────────────┘ │
│                                                                      │
│ [Discard local edits]       [Continue editing from current version]  │
└──────────────────────────────────────────────────────────────────────┘
```

`Continue editing` rebases the human editing session on the freshly loaded current DRAFT; no automatic merge or LWW occurs. Local editor content remains available for explicit copy/reconciliation.

### WF-10 — Submitted Revision / Submission — DW-04

```text
┌───────────────┬──────────────────────────────────────────────────────┐
│ shell         │ PO-ADM-042 / REV004                   SUBMITTED     │
│               ├──────────────────────────────────────────────────────┤
│               │ Human governance      Required / Pending            │
│               │ Official rendition    Required / Satisfied          │
│               │                                                      │
│               │ ┌──────────────────────────────────────────────────┐ │
│               │ │          READ-ONLY SUBMITTED SOURCE             │ │
│               │ └──────────────────────────────────────────────────┘ │
│               │                                                      │
│               │ [Withdraw submission]   [Cancel revision]           │
└───────────────┴──────────────────────────────────────────────────────┘
```

Outcome variants:

```text
governance pending
rendition pending
released → navigate/refetch Document Official
returned_for_changes
withdrawn
revision_cancelled
```

Buttons appear as UX affordances only; server remains authority.

## 6. My Work — `/work`

### WF-11 — My Work lanes — WRK-01/WRK-02

```text
┌───────────────┬──────────────────────────────────────────────────────┐
│ shell         │ My Work                                             │
│               │ [Authoring] [Governance]                            │
│               ├──────────────────────────────────────────────────────┤
│               │ Authoring                                           │
│               │ ┌──────┬──────────────────┬──────────┬───────────┐ │
│               │ │Code  │Title             │State     │Owner      │ │
│               │ ├──────┼──────────────────┼──────────┼───────────┤ │
│               │ │...   │...               │DRAFT     │...        │ │
│               │ └──────┴──────────────────┴──────────┴───────────┘ │
│               │                                          [More]     │
└───────────────┴──────────────────────────────────────────────────────┘
```

Governance lane:

```text
┌──────┬──────────────────┬──────────────┬─────────────┐
│Code  │Subject           │Created       │Open case    │
├──────┼──────────────────┼──────────────┼─────────────┤
│...   │Submission        │...           │[Open]       │
└──────┴──────────────────┴──────────────┴─────────────┘
```

Stale row target:

```text
Destination is no longer available.
[Refresh My Work]
```

Projection row never remains mutation authority.

## 7. Governance Case — `/work/governance/:attempt_id`

### WF-12 — Submission governance case — GOV-01/GOV-02/GOV-03

```text
┌───────────────┬──────────────────────────────────────────────────────┐
│ shell         │ Governance case                         ACTIVE      │
│               │ PO-ADM-042 / REV004                                │
│               │ Submitted by: Ana                                  │
│               ├──────────────────────┬───────────────────────────────┤
│               │                      │ Steps                         │
│               │ READ-ONLY EXACT      │ 1. Manager        DECIDED    │
│               │ SUBMISSION CONTENT   │ 2. Director       ACTIVE     │
│               │                      │ 3. Governance     PENDING     │
│               │                      │                               │
│               │                      │ Feedback                      │
│               │                      │ - ...                         │
│               │                      │ [Add feedback]                │
│               │                      │                               │
│               │                      │ Decision                      │
│               │                      │ [Accept] [Return for changes] │
└───────────────┴──────────────────────┴───────────────────────────────┘
```

`Return for changes` opens required reason form.

If `allowed_actions` is empty, decision/feedback controls are absent/disabled as UX only; direct operation still relies on server.

Already-decided conflict:

```text
This Step already has a decision.
[Reload case]
```

No "Publish" action exists.

### WF-13 — Obsolescence governance case — GOV-01/GOV-03

```text
┌──────────────────────────────────────────────────────────────────────┐
│ Governance case                                        ACTIVE       │
│ Obsolescence request                                                 │
│ Document: PO-ADM-042                         [View current document] │
│ Target revision: REV003                                           │
│ Initiator: Ana                                                     │
│ Reason: superseded policy                                          │
│                                                                      │
│ Steps                                                               │
│ ...                                                                 │
│                                                                      │
│ [Accept] [Return for changes] [Add feedback]                         │
└──────────────────────────────────────────────────────────────────────┘
```

The immutable case subject remains primary even if current Document Official later changes.

## 8. Document History — `/documents/:document_id/history`

### WF-14 — Semantic history timeline — HIS-01

```text
┌───────────────┬──────────────────────────────────────────────────────┐
│ shell         │ PO-ADM-042 / History                [Back document] │
│               ├──────────────────────────────────────────────────────┤
│               │ ○ REV000 created                                    │
│               │ │                                                    │
│               │ ○ Submission created             [Inspect source]    │
│               │ │                                                    │
│               │ ○ Governance accepted            [Open case]         │
│               │ │                                                    │
│               │ ○ Release completed              [Inspect release]   │
│               │ │                                                    │
│               │ ○ Obsolescence requested         [Inspect request]   │
│               │                                                      │
│               │                                             [More]   │
└───────────────┴──────────────────────────────────────────────────────┘
```

Inline detail drawer may use only ids returned by that exact history union item.

History never shows a button like "Open current work from this event" by reconstructing current state.

## 9. Audit — `/audit`

### WF-15 — Audit evidence — AUD-01

```text
┌───────────────┬──────────────────────────────────────────────────────┐
│ shell         │ Audit                                                │
│               ├──────────────────────────────────────────────────────┤
│               │ ┌────────────┬──────────────┬─────────┬────────────┐ │
│               │ │Occurred    │Actor         │Action   │Resource    │ │
│               │ ├────────────┼──────────────┼─────────┼────────────┤ │
│               │ │...         │...           │...      │...         │ │
│               │ └────────────┴──────────────┴─────────┴────────────┘ │
│               │                                             [More]   │
└───────────────┴──────────────────────────────────────────────────────┘
```

No generic filters/search/export/retry/current-state controls are added.

Audit resource identity is evidence, not a baseline generic navigation resolver.

## 10. Admin / Organization — `/admin/organization`

### WF-16 — Company + route section navigation — ORG-01

```text
┌───────────────┬──────────────────────────────────────────────────────┐
│ shell         │ Administration / Organization                       │
│               │ [Company] [Users] [Areas] [Groups]                  │
│               ├──────────────────────────────────────────────────────┤
│               │ Company                                              │
│               │ Display name [Metal Nobre........................] │
│               │                                                      │
│               │                                      [Save]          │
└───────────────┴──────────────────────────────────────────────────────┘
```

Stale Company ETag uses the standard explicit reconcile pattern.

### WF-17 — Users + create User — ORG-02

```text
┌──────────────────────────────────────────────────────────────────────┐
│ Organization / Users                                  [New user]     │
│ ┌──────────────────┬───────────┬───────────────────────────────────┐ │
│ │User              │Eligibility│Open                               │ │
│ ├──────────────────┼───────────┼───────────────────────────────────┤ │
│ │Ana               │ENABLED    │[Manage]                           │ │
│ │...               │...        │...                                │ │
│ └──────────────────┴───────────┴───────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────┘
```

Create drawer:

```text
Provider subject search [............................................]
Results:
  ( ) provider display hints...
Profile display name [...............................................]
Email optional        [...............................................]

[Cancel] [Create user]
```

Provider ref stays opaque.

### WF-18 — User management drawer — ORG-03/04/05

```text
┌──────────────────────────── User: Ana ───────────────────────────────┐
│ Identity: stable User id / current eligibility                      │
│ [Profile] [Provider binding] [Eligibility]                           │
├──────────────────────────────────────────────────────────────────────┤
│ Profile                                                              │
│ Display name [.....................................................] │
│ Email        [.....................................................] │
│ [Erase profile]                                      [Save profile] │
└──────────────────────────────────────────────────────────────────────┘
```

Provider binding tab:

```text
Current provider binding: <presentation only>
Search provider subject [..............................]
[Replace binding]

Warning: replacing the binding revokes required existing sessions.
```

Eligibility tab:

```text
Current: ENABLED

[Disable user]

Disabling will revoke sessions, remove current memberships and direct
RoleAssignments. Re-enabling later will not restore them.
```

Confirmation text reflects accepted atomic security semantics; it does not promise provider-side disable.

### WF-19 — Areas / Groups — ORG-06/07/08

Areas view:

```text
Organization / Areas                                      [New area]
┌────────┬──────────────────────┬──────────┬───────────────┐
│Code    │Name                  │State     │Manage         │
├────────┼──────────────────────┼──────────┼───────────────┤
│ADM     │Administration        │ACTIVE    │[Open]         │
└────────┴──────────────────────┴──────────┴───────────────┘
```

Area drawer keeps separate ETag cards:

```text
Identity
  code read-only
  name [.....................] [Save name]

Lifecycle
  ACTIVE / RETIRED            [Change lifecycle]
```

Groups view:

```text
Organization / Groups                                    [New group]
┌────────────────────────┬─────────────────────┐
│Name                    │Manage               │
├────────────────────────┼─────────────────────┤
│Quality                 │[Open]               │
└────────────────────────┴─────────────────────┘
```

Delete conflict shows "Group has live dependencies" rather than generic validation.

## 11. Admin / Access — `/admin/access`

### WF-20 — Group memberships — ACC-01

```text
┌───────────────┬──────────────────────────────────────────────────────┐
│ shell         │ Administration / Access                             │
│               │ [Memberships] [Roles & assignments]                 │
│               ├───────────────────┬──────────────────────────────────┤
│               │ Groups            │ Quality members                  │
│               │ > Quality         │ Ana                    [Remove] │
│               │   Finance         │ João                   [Remove] │
│               │   ...             │                                  │
│               │                   │ [Add member ▾] [Add]             │
└───────────────┴───────────────────┴──────────────────────────────────┘
```

Add selector uses admitted User references; membership controls do not move semantic ownership out of Organization.

### WF-21 — Role catalog + assignments — ACC-02

```text
┌──────────────────────────────────────────────────────────────────────┐
│ Access / Roles & assignments                     [New assignment]    │
│                                                                      │
│ Role catalog                                                         │
│ governance_admin  Company only   organization.manage, access.manage… │
│ area_manager      Area only      ...                                 │
│ author            Company/Area   ...                                 │
│                                                                      │
│ Current assignments                                                  │
│ ┌─────────────┬────────────┬────────────┬───────────┬──────────────┐ │
│ │Subject      │Kind        │Role        │Scope      │Action        │ │
│ ├─────────────┼────────────┼────────────┼───────────┼──────────────┤ │
│ │Ana          │User        │author      │ADM        │[Revoke]      │ │
│ └─────────────┴────────────┴────────────┴───────────┴──────────────┘ │
└──────────────────────────────────────────────────────────────────────┘
```

Assignment drawer constrains scope UI from returned RoleView but server revalidates:

```text
Subject kind [User|Group]
Subject      [admitted reference ▾]
Role         [fixed role ▾]
Scope kind   [allowed by selected role ▾]
Area         [when area scope ▾]
[Grant]
```

No custom role/permission editor.

## 12. Admin / Document Governance — `/admin/document-governance`

### WF-22 — Document Types list + create/base config — DGV-01/DGV-02/DGV-05

```text
┌───────────────┬──────────────────────────────────────────────────────┐
│ shell         │ Administration / Document Governance                │
│               │ [Document types] [Templates]                         │
│               ├──────────────────────────────────────────────────────┤
│               │ Document types                         [New type]    │
│               │ ┌──────┬────────────────┬────────────┬─────────────┐ │
│               │ │Code  │Name            │Active      │Open         │ │
│               │ ├──────┼────────────────┼────────────┼─────────────┤ │
│               │ │PO    │Policy          │Yes         │[Manage]     │ │
│               │ └──────┴────────────────┴────────────┴─────────────┘ │
└───────────────┴──────────────────────────────────────────────────────┘
```

Create drawer includes all required initial authority:

```text
Code
Name
Numbering scope
Active
Governance mode + route when required
Representation policy
[Create]
```

No hidden default governance/representation.

Existing type drawer:

```text
Base configuration
  code/name/numbering scope/active
  [Save base]

Number preview
  Area [when relevant]
  Preview: PO-ADM-042
  "Preview only; not reserved"
```

### WF-23 — Governance + representation route editor — DGV-03

```text
┌──────────────────── Governance configuration ────────────────────────┐
│ Mode: ( ) No human approval  (•) Use governance route              │
│                                                                      │
│ Steps                                                                │
│ 1. [Manager approval........] Selector [Named user ▾] [Ana ▾]      │
│ 2. [Quality approval........] Selector [Group ▾]      [Quality ▾]  │
│                                      [↑] [↓] [Remove]               │
│ [Add step]                                                           │
│                                                                      │
│ Representation: ( ) Source only  (•) Require official PDF rendition │
│                                                                      │
│ [Cancel]                                                [Save]       │
└──────────────────────────────────────────────────────────────────────┘
```

Step order is semantic. This is not a generic workflow designer; only the closed accepted selector/mode vocabulary exists.

Stale governance ETag shows current route beside preserved local route before any re-submit.

### WF-24 — Eligible Templates — DGV-04

```text
┌──────────────────────── Eligible templates ──────────────────────────┐
│ Document type: PO                                                    │
│                                                                      │
│ [✓] TMP-001  Standard policy                                        │
│ [ ] TMP-002  Short policy                                           │
│                                                                      │
│ [Cancel]                                                [Save set]   │
└──────────────────────────────────────────────────────────────────────┘
```

Selection sources come from admitted Template configuration projection. Empty set is valid.

### WF-25 — Template configuration / Template role — DGV-06

```text
┌──────────────────────────────────────────────────────────────────────┐
│ Document Governance / Templates                                     │
│ ┌──────────┬──────────────────┬──────────┬─────────────┬───────────┐ │
│ │Code      │Current title     │Template  │Effective    │Manage     │ │
│ ├──────────┼──────────────────┼──────────┼─────────────┼───────────┤ │
│ │TMP-001   │Standard policy   │Yes       │Yes          │[Open]     │ │
│ └──────────┴──────────────────┴──────────┴─────────────┴───────────┘ │
└──────────────────────────────────────────────────────────────────────┘
```

Manage drawer:

```text
Template role: [On/Off]
Eligible for Document Types: read/composed configuration context
[Save Template role]
```

Template configuration access never unlocks content/history viewer automatically.

## 13. Cross-cutting material overlays

These overlays are reused only where their F3 contract reaches the corresponding state.

### WF-X1 — Permission denied

```text
┌──────────────────────────────────────────────────────────────────────┐
│ You do not currently have access to this action or view.             │
│                                                                      │
│ [Back]                                                               │
└──────────────────────────────────────────────────────────────────────┘
```

No client explanation of role logic is manufactured.

### WF-X2 — Not found / non-disclosable

```text
This item is not available.
```

Do not distinguish "does not exist" from "you cannot know it exists".

### WF-X3 — CSRF re-bootstrap

Normally no dedicated dialog is needed:

```text
unsafe request receives permission.csrf_failed
→ re-read SessionView/CSRF
→ retry the SAME logical command when the accepted retry law permits
```

If recovery cannot complete, show ordinary request failure without changing the command key/condition silently.

### WF-X4 — Whole-replacement OCC conflict

```text
Current server value changed while you were editing.
Your local form was preserved.

[Discard local changes]
[Review current + local and submit again]
```

Fresh re-submit uses the new current ETag after explicit review.

### WF-X5 — Idempotent ambiguous outcome

```text
The result of this request could not be confirmed.
[Retry same request]
```

No key is shown as business identity.

### WF-X6 — Dependency unavailable

```text
This external dependency is temporarily unavailable.
[Retry] only when the owning interaction's retry law permits it
```

Raw provider/scanner/storage errors never appear.

### WF-X7 — Exact-content integrity failure

```text
Document content could not be verified.
Reference: <sanitized trace context>
```

Viewer stays empty; no partial authoritative display.

## 14. Surface coverage reconciliation

| Surface | Wireframe home |
|---|---|
| APP-01 | WF-00 |
| APP-02 | WF-01 |
| LIB-01 | WF-02 |
| LIB-02 | WF-03 |
| OFF-01 | WF-04 |
| OFF-02 | WF-04 |
| OFF-03 | WF-05 |
| OFF-04 | WF-04 |
| OFF-05 | WF-05 |
| HIS-01 | WF-14 |
| WRK-01 | WF-11 |
| WRK-02 | WF-11 |
| DW-01 | WF-06/WF-07/WF-08 |
| DW-02 | WF-08 |
| DW-03 | WF-07 |
| DW-04 | WF-06/WF-10 |
| GOV-01 | WF-12/WF-13 |
| GOV-02 | WF-12/WF-13 |
| GOV-03 | WF-12/WF-13 |
| ORG-01 | WF-16 |
| ORG-02 | WF-17 |
| ORG-03 | WF-18 |
| ORG-04 | WF-18 |
| ORG-05 | WF-18 |
| ORG-06 | WF-19 |
| ORG-07 | WF-19 |
| ORG-08 | WF-19 |
| ACC-01 | WF-20 |
| ACC-02 | WF-21 |
| DGV-01 | WF-22 |
| DGV-02 | WF-22 |
| DGV-03 | WF-23 |
| DGV-04 | WF-24 |
| DGV-05 | WF-22 |
| DGV-06 | WF-25 |
| AUD-01 | WF-15 |

```text
F2 material surfaces        36 / 36 wireframed
primary/variant frames      WF-00 → WF-25
cross-cutting overlays      7
new Product routes          0
new application operations  0
operation 79                absent
```

## 15. F5 finding law

A visual inconvenience is not a backend requirement.

If a later review of these wireframes proves that a material action/data block cannot be populated through its F3/F4 trace:

```text
STOP that wireframe
→ classify the exact accepted human goal
→ smallest owner analysis
→ bounded correction only when genuinely required
```

No screen-shaped endpoint, generic reference-data platform or client semantic authority is allowed as a convenience repair.

## 16. F5 status

```text
F5 wireframes                         COMPLETE CANDIDATE
material surfaces covered             36 / 36
stable routes preserved               10 / 10
unresolved known material wireframe gap 0
next                                  F6 Material Interaction Ledger
```

The wireframe pack must still survive F6/F7 trace reconciliation and later independent review; `COMPLETE CANDIDATE` is not T11 ratification.
