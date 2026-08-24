# T11 — B10 P9 Screen Contract / Bidirectional Trace

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B10 — Organization Administration.  
> **Depends on:** operator LOCK of `t11-b10-organization-administration-p8.html`, accepted B01/B01N global chrome, Product/Organization authority, Authorization boundary, executable wire.  
> **Locked P8 Git blob:** `1d1cc7d5cb42e034ab9ee71a21c96918cdcf691d`.  
> **Implementation:** BLOCKED.

This file is temporary planning Evidence under `docs/work/**`. It must be absent from the eventual merge candidate/main.

## 1. Goal

Prove that every material B10 region/control is realizable by current operations 3–26 without flattening independent concurrency domains, inventing Organization search, creating frontend Authorization, leaking provider identity semantics, or crossing into B11 GroupMembership/RoleAssignment administration.

## 2. Screen contract

| Surface / region | User goal | Current read truth | Material control / write | Identity source | Material failure / safe UX | Forbidden frontend authority | Status |
|---|---|---|---|---|---|---|---|
| B10-01 stable route | enter Organization administration | accepted `/admin/organization` route | route/read only | stable route | route denial != empty organization | no client `organization.manage` evaluator | READY |
| B10-02 global shell | preserve Product orientation/chrome | B01 + B01N LOCKED shell | existing global navigation + Quick Inbox | existing shell state | shell remains independent from B10 read failures | no B10-specific global IA | READY |
| B10-03 local section navigation | move among Company/Users/Areas/Groups without new Product routes | B10 LOCKED local workspace | local/URL-addressable presentation state only | section key within `/admin/organization` | invalid local state falls back safely without changing Product route meaning | no new stable Product route family | READY |
| B10-04 Company current truth | inspect current Company display identity | op4 `getCompany` -> `CompanyView` + ETag | read | `company_id` | unavailable/denied explicit | no generic settings registry | READY |
| B10-05 Company edit | change accepted Company `display_name` only | op4 current representation + ETag | op5 `replaceCompany` + If-Match | current Company identity + exact ETag | 412 preserves intended input and requires refetch/reconciliation | no multi-field tenant/settings model | READY |
| B10-06 Company stale recovery | avoid lost update | op4 current truth | refetch, compare, deliberate resubmit | same Company + new ETag | stale != mutation success | no silent overwrite / client merge authority | READY |
| B10-07 Users collection | find current User candidates by admitted browse | op6 `listUsers` -> cursor page | page continuation only | returned `user_id` | failed page != known-empty; preserve loaded page on continuation failure | no fake page-local global search / total-count inference | READY |
| B10-08 User selection | inspect one exact User identity/eligibility | op8 `getUser` | select/detail presentation | returned `user_id` | 404 absent/non-disclosable closes stale selection | list row != complete profile/binding truth | READY |
| B10-09 Create User entry | begin one atomic User+Profile+Binding command | current operator intent | open local form | no semantic id before success | abandoned form has no server effect | no provisional User authority | READY |
| B10-10 provider-subject lookup for create | resolve exact external authentication subject without importing provider authority | op3 `searchProviderSubjects` | query/select one returned option | opaque `provider_subject_ref` | 503 provider unavailable stays preflight failure; no local User created | provider hints/ref never become Product identity | READY |
| B10-11 Create User command | establish enabled User + required Profile + Binding atomically | selected provider ref + local profile input | op7 `createUser` + Idempotency-Key | returned `user_id`; one logical-command key | ambiguous outcome retries same key; validation/provider failure never fabricates partial success | no client multi-step create transaction | READY |
| B10-12 User Profile read | inspect erasable profile enrichment | op9 `getUserProfile` + ETag | read | selected `user_id` | absent profile distinguished from User absence | User identity != Profile existence | READY |
| B10-13 User Profile replace/recreate | edit profile independently from eligibility/binding | op9/absence + profile conditional law | op10 `replaceUserProfile` with If-Match or If-None-Match:* | selected User + profile concurrency domain | 412 requires current-truth reconciliation | no single User aggregate Save | READY |
| B10-14 User Profile erase | deliberately remove erasable enrichment | op9 current/absence | op11 `deleteUserProfile` | selected `user_id` | absent profile -> 404; deletion != offboarding | no User deletion/history rewrite | READY |
| B10-15 Provider Binding read | inspect current authentication binding separately | op12 `getUserProviderBinding` + ETag | read | selected User + opaque ref | unavailable/denied explicit | provider ref not parsed/displayed as Product identity | READY |
| B10-16 provider lookup for rebind | choose exact replacement subject | op3 bounded lookup | query/select one option | opaque `provider_subject_ref` | 503 leaves current binding untouched | no provider directory synchronization | READY |
| B10-17 Provider Binding replace | deliberately rebind authentication and trigger current security law | op12 current + ETag | op13 `replaceUserProviderBinding` + If-Match | selected User + exact binding ETag | 412 zero mutation; provider failure before local transition | no client session-revocation/binding orchestration authority | READY |
| B10-18 Eligibility read | understand whether User can act now | op14 `getUserEligibility` + ETag | read | selected `user_id` | unavailable/denied explicit | eligibility != role/permission snapshot | READY |
| B10-19 Disable / offboard | stop future access/actions with truthful consequence | op14 current eligibility | op15 desired `disabled` + If-Match after explicit confirmation | selected User + eligibility ETag | stale rechecks; UI explains session/membership/direct-grant teardown | no frontend implementation of teardown / no fake deletion | READY |
| B10-20 Re-enable | restore eligibility only | op14 current eligibility | op15 desired `enabled` + If-Match | selected User + eligibility ETag | stale reconciliation; success explicitly does not restore old sessions/memberships/grants | no resurrection inference | READY |
| B10-21 Areas collection | browse current Area identities/lifecycle summaries | op16 `listAreas` -> cursor page | page continuation | returned `area_id` | failed continuation preserves loaded rows | no collection search/total inference | READY |
| B10-22 Create Area | establish new immutable code + name / ACTIVE initial state | local inputs | op17 `createArea` + Idempotency-Key | returned `area_id` | ambiguous retry same logical key; validation clear | no client code reservation/state authority | READY |
| B10-23 Area identity detail | inspect exact Area identity/name | op18 `getArea` + ETag | read | selected `area_id` | 404 closes stale selection | code/name projection not lifecycle authority | READY |
| B10-24 Area rename | change name while preserving immutable code | op18 current + ETag | op19 `replaceArea` + If-Match | selected Area + metadata ETag | 412 reconciliation | no editable code after creation | READY |
| B10-25 Area lifecycle read | inspect ACTIVE/RETIRED independently | op20 `getAreaLifecycle` + ETag | read | selected Area + lifecycle domain | unavailable/denied explicit | metadata form != lifecycle form | READY |
| B10-26 Area retire / re-enable | block/restore future use while preserving references/history | op20 current lifecycle | op21 `replaceAreaLifecycle` + If-Match | selected Area + lifecycle ETag | stale reconciliation; retirement never presented as deletion | no history rewrite / cascading delete inference | READY |
| B10-27 Groups collection | browse Group identity | op22 `listGroups` -> cursor page | page continuation | returned `group_id` | failed continuation preserves loaded rows | no membership/role projection as Group identity | READY |
| B10-28 Create Group | create Organization-owned Group identity | local name | op23 `createGroup` + Idempotency-Key | returned `group_id` | ambiguous retry same key; validation explicit | no membership defaults / provider mirroring | READY |
| B10-29 Group detail | inspect exact Group identity/name | op24 `getGroup` + ETag | read | selected `group_id` | 404 closes stale selection | Group identity != effective access | READY |
| B10-30 Group rename | maintain Group identity | op24 current + ETag | op25 `replaceGroup` + If-Match | Group + metadata ETag | 412 reconciliation | no membership mutation | READY |
| B10-31 Group delete | remove Group identity only when dependency law permits | op24 selection/current context | op26 `deleteGroup` | exact `group_id` | dependency 409 explains block; no destructive workaround | no client dependency bypass / cascade | READY |
| B10-32 B10/B11 boundary | understand that access-bearing membership/role work lives elsewhere | accepted Organization/AuthZ ownership | no B10 write; navigate to accepted Access space only if user chooses | stable global `Acesso` destination | B10 may explain boundary but cannot surface member/grant mutation controls | no op27+ consumption inside B10 | READY |
| B10-33 material failure states | understand safe next action after denial/absence/validation/dependency failures | current Problem contract | retry/refetch/edit/navigate as appropriate | operation/resource context | 403/404/409/412/422/503 remain distinct | no generic toast that erases consequential state | READY |
| B10-34 responsive/accessibility | preserve same Organization meaning across viewport/input modes | same B10 truth | global drawer + local section menu + dialog/focus-visible controls | no new Product identity | no mobile-only Product behavior | no hover-only material control | READY |

## 3. Exact operation homes used by B10

```text
3   searchProviderSubjects
4   getCompany
5   replaceCompany
6   listUsers
7   createUser
8   getUser
9   getUserProfile
10  replaceUserProfile
11  deleteUserProfile
12  getUserProviderBinding
13  replaceUserProviderBinding
14  getUserEligibility
15  replaceUserEligibility
16  listAreas
17  createArea
18  getArea
19  replaceArea
20  getAreaLifecycle
21  replaceAreaLifecycle
22  listGroups
23  createGroup
24  getGroup
25  replaceGroup
26  deleteGroup
```

B10 consumes **no operation 27+**.

```text
27 listGroupMembers
28 addGroupMember
29 removeGroupMember
30+ roles / RoleAssignment
```

remain B11 / `access.manage`.

## 4. Bidirectional trace

### Product/backend -> frontend

```text
CompanyView + Company ETag
-> Company inspect/edit/stale-reconcile surface

UserPage + UserView
-> paginated Users collection + exact selected User shell

UserProfileView / UserProviderBindingView / UserEligibilityView
-> three independent selected-User detail regions
-> three independent mutation/concurrency paths

ProviderSubjectSearchView
-> exact opaque provider-subject selection for create/rebind only

AreaPage + AreaView + AreaLifecycleView
-> paginated Areas + immutable-code identity + separate lifecycle

GroupPage + GroupView
-> paginated Group identity list/detail

deleteGroup dependency conflict
-> dependency-blocked delete outcome, not client-side dependency authority
```

### Frontend -> Product/backend

```text
edit Company name
-> getCompany ETag
-> replaceCompany If-Match

create User
-> searchProviderSubjects
-> selected opaque provider_subject_ref
-> createUser with one logical Idempotency-Key

edit/erase Profile
-> getUserProfile / absence
-> replaceUserProfile profile conditional law OR deleteUserProfile

rebind authentication
-> searchProviderSubjects
-> getUserProviderBinding ETag
-> replaceUserProviderBinding If-Match

disable/re-enable
-> getUserEligibility ETag
-> replaceUserEligibility If-Match

create/rename Area
-> createArea OR getArea ETag -> replaceArea

retire/re-enable Area
-> getAreaLifecycle ETag
-> replaceAreaLifecycle

create/rename/delete Group identity
-> createGroup OR getGroup ETag -> replaceGroup/deleteGroup

membership / RoleAssignment need
-> B11 boundary
-> never op27+ from B10 controls
```

## 5. Client state classes

B10 uses only accepted frontend state classes:

```text
SERVER STATE
  Company, User pages/detail, Profile, Binding, Eligibility,
  Area pages/detail/lifecycle, Group pages/detail, provider lookup responses

NAVIGATION / URL
  stable /admin/organization + local section/selection context when useful

FORM DRAFT
  Company/Profile/Binding/Eligibility/Area/Group edits and create forms

EPHEMERAL UI
  dialogs, confirmations, currently open local menu, failure-Evidence toggles,
  focus restoration, local fixture scenario state
```

No normalized Organization entity graph, permission matrix, provider mirror or access cache becomes client authority.

## 6. Ordering / pagination / identity mechanics

```text
Users
  cursor-paginated; returned User identity only

Areas
  cursor-paginated; canonical code/id order owned by wire

Groups
  cursor-paginated; canonical group identity order owned by wire

provider lookup
  bounded preflight, max provider result envelope; not a general directory page

collection search/filter
  absent from accepted B10 contract
  page-local filtering presented as global search is forbidden

B10-A1
  operator-validated for current Launch P8
  reopen only on material real scale/findability Evidence
```

## 7. Material failure intent

```text
403 permission.denied
  denied Organization context/action; never known-empty

404 notfound.*
  selected resource/profile absent or non-disclosable according to operation contract

409 Group dependency conflict
  preserve Group; explain blocking dependency class; no cascade workaround

412 precondition.resource_changed
  preserve intended form input; refetch exact concurrency domain; explicit reconcile/resubmit

422 validation.*
  keep form/input; bind field/business validation where available

503 provider dependency
  keep current local Product truth unchanged; lookup/rebind preflight can retry

ambiguous Idempotency-Key create outcome
  retry same logical command with same key; never issue second deliberate create
```

## 8. Access / disclosure proof

```text
visible control                    != Authorization
organization.manage               != access.manage
Group identity                     != GroupMembership
User eligibility                   != effective permission snapshot
provider_subject_ref               != MetalDocs User identity
Profile existence                  != User existence
Profile deletion                   != offboarding
re-enable                          != access restoration
Area retirement                    != historical deletion
Group delete UI                    != dependency authority
page of Users/Areas/Groups         != complete collection search universe
local section state                != new Product route authority
```

Every operation rechecks current server Authorization and owner truth.

## 9. B10 negative contract

The locked B10 contains no Product surface for:

```text
global collection search/filter for Users/Areas/Groups
bulk User/Area/Group actions
custom Role/Permission editor
GroupMembership add/remove
RoleAssignment grant/revoke
effective-permission explorer
provider group/role mirroring
identity-provider sync center
User deletion / historical identity erasure
multi-company/tenant switching
Company settings framework beyond accepted display_name
Area code editing after creation
Group dependency override/cascade
client-side access matrix
```

## 10. P9 closure

```text
material B10 regions/controls traced          34 / 34
accepted B10 operations bound                 24 / 24 (ops 3–26)
unbound material controls                     0
invented operations                           0
operation 27+ consumed by B10                  0
screen-shaped APIs                            0
frontend Authorization evaluator              0
fake page-local global search                 0
collapsed independent ETag/write domains      0
material B10 Screen Contract findings         0
```

P9 is complete for the operator-locked B10 scope.
