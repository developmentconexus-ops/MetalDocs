# T11 — Legacy Frontend Evidence Ledger

> **Status:** EVIDENCE / NOT AUTHORITY.  
> **Source:** `archive/r10-pr131-pre-reset-20260820` (pre-reset preserved implementation lineage).  
> **Purpose:** recover useful user-facing information architecture, screen responsibilities, information density and interaction lessons from the superseded MetalDocs frontend so T11 does not rediscover already-paid UX knowledge.  
> **Hard law:** current Product/T1→T10/T11 authority always wins. Nothing in this ledger restores legacy code, routes, lifecycle, permissions, semantic owners, API operations or Launch scope.

## 1. Why this evidence is being read

Direct operator evidence on 2026-08-22 identified that the first B03 candidate had collapsed two distinct user intentions:

```text
enter the Document record / understand the Document
vs
view/read the exact document content
```

The pre-reset frontend already contained useful separation and information that can falsify an over-simplified new wireframe even though its backend/domain model is superseded.

Use the legacy only for questions such as:

```text
what did users need to see?
which facts were grouped together?
which contexts were intentionally separate?
which views reduced navigation/friction?
which empty/error/loading states had useful teaching value?
```

Never use it to answer:

```text
what is Product authority now?
which route/API/permission/lifecycle must exist now?
which old module should be restored?
```

## 2. Cross-product UX evidence worth preserving

Legacy `frontend/apps/web/PRODUCT.md` explicitly emphasized:

```text
the document is the protagonist
accountability must be legible: status / revision / who / when / why
density with hierarchy for professional users
one clear primary action per screen
loading / empty / error / disabled states are designed, not afterthoughts
human-readable codes and labels lead; UUID/hash does not
```

Disposition: **KEEP-AS-UX-EVIDENCE**. These principles remain useful when compatible with current authority. The same legacy document also described a broader eQMS/e-sign/distribution product; that scope is superseded and has no independent authority.

## 3. Legacy route/surface census and current disposition

| Legacy surface | Legacy user intention | Current disposition | Current block/home |
|---|---|---|---|
| `/dashboard` | see pending work / quick context | **ADAPT** | B01 Home / B05 My Work |
| `/documents` | browse/filter document collection | **ADAPT** | B02 Library — already re-derived/LOCKED |
| `/documents/new` | guided creation | **ADAPT** | Library creation material surface; current create authority only |
| `/documents/:documentId` | mode-adaptive content workspace: edit/read/approval | **SPLIT** | B03 official record/read entry + B04 Document Work + B06 Governance Case |
| `/documents/:documentId/details` | full Document record/ficha and revision context | **STRONG ADAPT** | B03 Document Official record + bounded links/summary to B07 History |
| `/documents/:documentId/distribution` | distribution/read coverage | **DEFER** | Launch+ / future; not B03 Launch |
| `/approvals` | actor approval inbox | **ADAPT ERGONOMICS, DROP OWNER** | B01 Home + B05 My Work + B06 Governance |
| `/approval-routes` | generic workflow administration | **DROP/REPLACE** | current DocumentType governance config under B08; no generic workflow builder |
| `/templates` + `/templates/:id/*` | peer Template lifecycle/list/detail/editor/approval | **DROP PEER LIFECYCLE / ADAPT TASKS** | Template is ordinary governed Document role; creation + B08 config only |
| `/admin/*` | people, roles, memberships, audit, sessions, usage | **PARTIAL ADAPT** | B08 Organization / Access / Document Governance + B07 Audit |
| `/admin/taxonomy` | generic taxonomy administration | **DROP** | no Launch Taxonomy workspace |
| `/templates/tokens` | Token management | **DROP** | no Launch Token product authority |
| password-change/local-password surfaces | credential management | **DROP** | current OIDC provider owns credentials |

This route census is evidence of user intentions, not permission to add current routes. Current stable SPA Product routes remain the accepted ten.

## 4. Document-detail evidence — highest-value legacy input for B03

Legacy sources:

```text
frontend/apps/web/src/features/documents/routes.tsx
frontend/apps/web/src/features/documents/pages/DocumentDetailRoute.tsx
frontend/apps/web/src/features/shared/controlled-artifact/ArtifactDetailView.tsx
frontend/apps/web/src/features/documents/components/workspace/WorkspaceSidebar.tsx
frontend/apps/web/src/features/documents/pages/DocumentWorkspacePage.tsx
```

### 4.1 Proven legacy separation

The old product deliberately distinguished:

```text
Library
→ content workspace / editor / review context
→ full Document detail/ficha
```

The workspace sidebar even exposed a literal `Ver ficha completa do documento` navigation to `/documents/:id/details`.

**B03 implication:** the initial T11 viewer-first candidate was too collapsed. The new architecture should distinguish:

```text
Document Official record/ficha
vs
official-content viewing surface
vs
open Revision work
vs
governance case
vs
full history
```

without necessarily adding a new stable route. Material surfaces may live within the accepted route/lens structure.

### 4.2 Information shown on the legacy ficha

The legacy detail surface carried:

```text
Document identity
  code
  human title
  status
  revision/version

About / responsibility
  responsible owner
  type/profile
  area
  effective date
  next review date
  file size
  classification placeholder

Version governance
  approval/sign-off chain by stage
  actor
  decision state
  signed/decided time

Lineage
  revision/version timeline
  creation time
  lifecycle/status per revision
  current marker
  governed title/summary

Content actions
  view PDF/content
  download

Lifecycle/work actions
  create/continue revision
  review/publish according to legacy state
  copy link

Additional sections
  coverage/distribution
  related artifacts placeholder
  comments placeholder
```

### 4.3 Current disposition of each information family

| Legacy information | Disposition now | Reason/current seam |
|---|---|---|
| code + governed official title | **KEEP** | current `DocumentOfficialView` / `ReleasedRevisionView` |
| official status + revision | **KEEP** | current official lens truth |
| Document Type | **KEEP** | current `DocumentTypeReference` |
| Area | **KEEP** | current `AreaReference` |
| responsible owner | **KEEP** | current `UserReference` + owner management precision |
| release/effective timestamp | **KEEP** | current `released_at` |
| exact source / official representation | **KEEP + ADAPT** | current exact-content/representation authority |
| file size/format | **KEEP WHEN USEFUL** | current `ContentSummary`; presentation must not expose hash as primary UX |
| open work/revision existence | **KEEP + ADAPT** | disclosure-safe `open_revision` routing reference |
| active obsolescence context | **KEEP + ADAPT** | current disclosure-safe request reference |
| revision lineage summary | **ADAPT / TRACE REQUIRED** | current Document History owns semantic history; B03 may consume only a bounded supporting summary/read if current trace permits, while B07 owns full history composition |
| approval/sign-off chain for effective revision | **ADAPT / TRACE REQUIRED** | useful accountability evidence; current immutable governance facts live in history, not `DocumentOfficialView`; must not resurrect old Approval owner |
| next periodic-review date | **DEFER** | Periodic Review is Launch+ |
| distribution/coverage | **DEFER** | Distribution/Read & Acknowledge is Launch+ |
| comments | **DROP FOR LAUNCH** | no current Launch comment capability |
| related artifacts | **DROP/DEFER** | no current relationships capability admitted |
| classification/tags | **DROP** | no current Launch classification/taxonomy/tag authority |
| copy/public link | **DO NOT ASSUME** | no current public-link/share authority |
| legacy publish command | **DROP** | current Release is system-owned, no user publish operation |

## 5. Workspace/editor evidence — useful for B04/B06, not a route template

Legacy `DocumentWorkspacePage` was one mode-adaptive screen that switched among edit, read and approval modes. It showed:

```text
content/editor as protagonist
Document title + code/revision/state badges
sync/save state
sidebar identification
revision lineage
approval context
current decision context
```

Useful information/ergonomics: **KEEP-AS-EVIDENCE**.

Legacy structural mistake under current authority: one route decided edit/read/approval mode from user/state and thereby mixed semantic lenses. Current architecture deliberately separates:

```text
Document Official
Document Work
Governance Case
```

Disposition: **DO NOT RESTORE MODE-ADAPTIVE AUTHORITY**.

## 6. Approval inbox evidence — use the queue ergonomics, not legacy Approval semantics

Legacy inbox provided:

```text
queue rail
one selected decision card
code/title
area
author/submitter
submitted time
current stage label
due/overdue context
stack vs timeline presentation
filter/empty/error states
keyboard next/previous
open-document action
```

Useful for B05/B06: queue orientation, one-decision-at-a-time focus, area/title/code prominence, teaching empty state.

Superseded concepts:

```text
peer Approval workspace/owner
quorum progress
quick generic approve/reject semantics
legacy template approval route
oversee model
localStorage preference as durable product preference
```

Disposition: **ADAPT UX / REJECT OLD AUTHORITY**.

## 7. Creation wizard evidence

Legacy creation was a four-step wizard with deliberate progression and confirmation. Useful interaction lessons:

```text
choose document classification/type context
choose area + title
choose template/blank seed
confirm before atomic creation
show code preview
fail closed when prior selection becomes unavailable
```

Current authority already owns a different and cleaner create model:

```text
Document Type
Area
REV000 title
optional eligible current-EFFECTIVE Template
responsible owner when permitted
non-reserving code preview
```

Legacy visibility/profile/family/external mechanics are not current authority. Disposition: **ADAPT WIZARD RHYTHM ONLY**.

## 8. Templates evidence

Legacy had independent list/detail/editor/approval routes and a full version lifecycle for Templates. Current Product explicitly rejects a peer Template lifecycle: Template is an ordinary governed Document role.

Reusable evidence:

```text
template selection benefits from code/name/title recognition
detail information and version clarity matter to governance admins
creation from template should explain what source is being used
```

Rejected structure:

```text
peer Template route tree
Template-specific approval owner
separate Template version authority
```

## 9. Admin evidence

Legacy Admin Center grouped overview, people, roles, memberships, audit, sessions and usage. Current authority should preserve the **human grouping lesson** but map it only to accepted owners:

```text
Organization
  Company
  Users / Profile / Provider Binding / Eligibility
  Areas
  Groups

Access
  Group membership
  Role assignments

Document Governance
  Document Types
  Governance route / representation
  eligible Templates / Template role

Audit
  separate evidence lens
```

Sessions/usage dashboards are not current Product admin surfaces.

## 10. Capabilities deliberately NOT resurrected

Legacy presence is not a reopen trigger for:

```text
generic Approval/workflow builder
peer Template lifecycle
Distribution / acknowledgement in Launch
generic Taxonomy
Tokens/placeholders platform
comments
artifact relationships
favorites/recents/sharing/public links
local password management
session administration
usage/metrics Product dashboard
periodic review in Launch
```

Any one of these needs a current named consumer/material reopen under the owning authority.

## 11. Per-block use of this ledger

Before each remaining material frontend block, the P0/P1 evidence recovery must include the relevant rows from this ledger when a corresponding legacy surface exists:

```text
B03 Document Official
  legacy detail/ficha + viewer/workspace separation

B04 Document Work
  legacy workspace/editor information + save/status ergonomics

B05 My Work
  dashboard + approval inbox queue ergonomics

B06 Governance
  inbox decision card + legacy approval-context presentation, not workflow semantics

B07 History / Audit
  legacy revision timeline/sign-off presentation + Admin Audit presentation

B08 Administration
  legacy Admin Center grouping + Template-governance configuration information
```

This does not make recursive legacy archaeology a default read. This bounded ledger is the reusable evidence summary; return to legacy source files only when a specific unresolved question materially needs exact evidence.

## 12. B03 correction produced by legacy evidence

The initial B03 viewer-first hypothesis is **REJECTED AS TOO COLLAPSED** by direct operator feedback plus legacy evidence.

Correct B03 design question becomes:

```text
What belongs on the stable Document Official record/ficha
so the user can understand identity, responsibility, official state,
current official revision and bounded governance/lineage context?

and

How does the user deliberately enter a distinct official-content viewer
without confusing that viewer with Document Work or Governance Case?
```

Constraints:

```text
no new stable Product route merely for visual convenience
viewer may be a material surface/modal/full-screen mode owned by B03
B04 remains the only current open-Revision work lens
B06 remains the exact Governance Case lens
B07 remains the full history owner
```

The next B03 P7/P8 iteration must compare record-first structures and a distinct content-viewing transition; it must not reuse the rejected viewer-first whole page as baseline.
