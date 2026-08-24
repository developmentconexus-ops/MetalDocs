# T11 B10 — Organization Administration P8 Realization Plan

> **Status:** TEMPORARY PLANNING EVIDENCE  
> **Method:** pinned `METHOD.md + FRONTEND-METHOD.md`  
> **Spec:** `docs/work/current/t11-b10-organization-administration-p7.md`  
> **Production implementation:** BLOCKED

## Goal

Produce one browser-operable low-fidelity HTML artifact for B10 that preserves the LOCKED global shell/chrome, realizes only the accepted Organization surface, and makes `B10-A1` plus material failure/recovery states falsifiable by operator use.

## Artifact boundary

Create exactly:

```text
docs/work/current/t11-b10-organization-administration-p8.html
```

No external assets, dependency manifests, production React/API code, reusable component framework, or repository script is permitted. HTML/CSS/vanilla JS and deterministic local fixtures only.

## Inherited structure

Preserve the exact semantic shape already LOCKED by B01/B01N:

```text
MetalDocs utility header
notification bell before session
left global navigation
Organization entry under Gestão active for this block
main content inside the accepted shell
```

Do not redesign Home, Notifications, global IA, or route inventory.

## Task 1 — Structural verifier first

Create an **ephemeral local verifier outside the repository** that fails until the P8 HTML exists and then checks the browser artifact through headless Chromium.

Required assertions:

```text
route marker = /admin/organization
Organization global nav active
local sections = Company | Users | Areas | Groups
no Access/Role/GroupMembership controls inside B10 content
Users collection paginates without fake search
B10-A1 falsification target is present across later pages
Company edit + stale conflict state operable
User create provider-subject lookup/selection operable
Profile, Binding, Eligibility are separate controls
Disable confirmation and re-enable consequence operable
Area create/rename/lifecycle actions operable
Group create/rename/delete dependency failure operable
403 / 404 / 409 / 412 / 422 / 503 Evidence states inspectable
keyboard-focusable controls exist
responsive layout remains usable at narrow viewport
```

Run verifier before HTML creation and confirm expected RED because the artifact is absent.

## Task 2 — Build the minimum P8 artifact

Use the LOCKED B01/B01N visual/structural grammar only as inherited Evidence. Implement:

```text
Global shell
  header + notification bell + session
  sidebar with Organization active

Organization workspace
  local tabs Company / Users / Areas / Groups

Company
  inspect display_name
  edit/save
  deterministic stale-ETag conflict/reload path

Users
  paginated list
  deterministic B10-A1 target on a later page
  list → detail
  create User provider lookup/selection/form
  Profile edit/erase
  Provider Binding lookup/replace
  Eligibility disable/re-enable

Areas
  paginated list
  create
  rename
  ACTIVE/RETIRED lifecycle transition

Groups
  paginated list
  create
  rename
  delete success candidate
  dependency-blocked delete state
```

Evidence controls may switch deterministic failure scenarios but must be visibly outside Product semantics.

## Task 3 — Browser verification

Run the ephemeral verifier against the completed local artifact. It must exercise real DOM events rather than static string checks for material interactions.

Minimum interaction sequence:

```text
open Users
advance pagination until B10-A1 target is visible
select target
open create-user flow
query provider subject and select one result
switch between Profile / Authentication binding / Eligibility
trigger disable confirmation, cancel once, confirm once
re-enable and inspect non-restoration explanation
open Company and force stale-save state, then reload current truth
open Areas, create an Area, select it, retire and re-enable it
open Groups, select protected Group, trigger blocked delete
switch Evidence scenario to permission denied / provider unavailable
verify narrow viewport local navigation remains reachable
```

Any failure blocks publishing the artifact.

## Task 4 — Publish exact verified bytes

After GREEN verification:

1. create `docs/work/current/t11-b10-organization-administration-p8.html` on the B10 branch using the exact verified local bytes;
2. fetch the committed file and compare its content hash with the verified local artifact;
3. do **not** set `LOCKED` or update P9/P10;
4. stop for operator walkthrough/disposition.

## P8 gate

P8 remains `CANDIDATE / OPERATOR REVIEW` after publication.

Only the operator may disposition:

```text
LOCK
REVISE
UPSTREAM FINDING
```

`B10-A1` remains OPEN until operator use establishes whether paginated browse is sufficient or materially inadequate.