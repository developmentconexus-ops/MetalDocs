# T11 B10 — Organization Administration P7 Candidate

Status: `CANDIDATE / OPERATOR-APPROVED FOR P8`  
Method profile: pinned `METHOD.md + FRONTEND-METHOD.md`  
Repository stage: `T11 / FP1 / B10 OPEN / ACTIVE`  
Implementation: `BLOCKED`

This file is temporary planning Evidence under `docs/work/**`. It is not Product/architecture authority and must be absent from the eventual merge candidate/main.

## 1. Authority recovery

Current authority was recovered from accepted `main` through:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ conexus-methodology ROUTER.md @ 9c7210d1504bef01c0d134a6c3ae8627deebb535
→ METHOD.md + FRONTEND-METHOD.md
→ docs/product/contract.md
→ docs/product/journeys.md
→ docs/architecture/authorization-and-audit.md
→ docs/architecture/frontend.md / wire-contract.md as bounded realization evidence
→ docs/decisions/api-operation-census.md
→ docs/decisions/forward-obligations.md
```

Read expansion beyond the normal repository-local budget was bounded to resolve two material questions: the current 89-operation census versus historical 78-operation snapshots, and the exact B10/B11 boundary around Group identity versus GroupMembership.

## 2. User / job

Primary actor:

```text
Governance Admin
+ current ENABLED MetalDocs User
+ organization.manage @ Company
```

B10 serves the human need to maintain current Organization identity/lifecycle truth without becoming a parallel access-administration or identity-provider authority.

Material jobs:

```text
maintain current Company display name
create and inspect Users
maintain erasable UserProfile enrichment
replace provider subject binding deliberately
manage User eligibility/offboarding/re-enable
create and maintain Areas including retirement/re-enable
create, rename and delete Group identity when dependencies allow
```

B10 does not administer GroupMembership or RoleAssignment; those belong B11 under `access.manage`.

## 3. Exact accepted boundary

Current Organization operation boundary is:

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

B11 begins at:

```text
27  listGroupMembers
28  addGroupMember
29  removeGroupMember
30+ Role / RoleAssignment administration
```

Group identity remains Organization-owned semantic state. Membership mutation is intentionally access-sensitive because it can change effective authority.

## 4. P6 — reference study

Disposition:

```text
NOT TRIGGERED
```

Reason: current uncertainty is primarily authority/boundary correctness, not an unfamiliar interaction pattern. External admin products would create feature pressure before a concrete UX uncertainty requires reference Evidence.

Reopen P6 only if P8 operation exposes a material interaction/scale/failure-pattern ambiguity that current accepted authority cannot resolve cleanly.

## 5. P7 — credible structures considered

### A — one long administration form

Disposition: `REJECTED`.

It would visually and behaviorally flatten Company, UserProfile, provider binding, User eligibility, Area metadata/lifecycle and Group identity into a false shared write model despite independent ETag/concurrency domains.

### B — Organization workspace with local navigation + list/detail

Disposition: `LEADING CANDIDATE / OPERATOR-APPROVED FOR P8`.

```text
/admin/organization

Organization
├── Company
├── Users
├── Areas
└── Groups
```

The stable Product route remains `/admin/organization`. Local state may be URL-addressable when useful, but B10 does not invent new stable Product route families merely to organize the workspace.

### C — four new stable Product subroutes

Disposition: `REJECTED FOR NOW`.

No current Evidence requires expanding the accepted 11-route Product IA. A future material navigation/deep-link failure may reopen only the affected frontend/route decision.

## 6. Surface candidate

### Company

```text
current Company display truth
→ inspect
→ edit display_name
→ save with current ETag / If-Match
→ stale result requires explicit current-truth reconciliation
```

No tenant switcher, multi-company console or generic settings framework.

### Users

```text
paginated user list
→ select User
→ inspect current User identity/eligibility

Create User
→ provider-subject query
→ exact provider result selection
→ initial display_name
→ optional email
→ create enabled User + required profile + binding atomically

Selected User
├── Profile
│   ├── replace profile under profile concurrency law
│   └── erase profile when explicitly intended
├── Authentication binding
│   └── provider-subject query → exact provider result → replace binding under its own ETag domain
└── Eligibility
    ├── disable / offboard
    └── re-enable
```

Profile deletion is not User deletion/offboarding. Offboarding preserves stable historical User identity/provider correlation while revoking current sessions/memberships/direct grants. Re-enable restores eligibility only; old access does not resurrect.

### Areas

```text
paginated Area list
→ create {code,name}
→ selected Area
   ├── rename metadata
   └── lifecycle ACTIVE | RETIRED
```

Area code is immutable after creation. Retirement is not deletion and preserves existing references/history while blocking future use as authority requires.

### Groups

```text
paginated Group list
→ create
→ selected Group
   ├── rename
   └── delete when no live dependency prevents it
```

B10 stops at Group identity. It must not expose membership add/remove or RoleAssignment controls.

## 7. Client-state / authority constraints

B10 must preserve:

```text
server read models remain canonical current truth
separate ETag domains remain separate forms/commands
frontend does not derive effective permissions
frontend does not mirror a durable User/Organization entity graph as authority
allowed UI presence never substitutes for server authorization
provider_subject_ref remains opaque anti-corruption data, never Product identity
stale whole-replacement mutation never silently overwrites current truth
```

No single "Save Organization" command exists.

## 8. Failures that must be operable in P8

At minimum the low-fidelity artifact must make these materially understandable:

```text
403 permission.denied
404 non-disclosable / absent selected resource
409 Group deletion blocked by live dependency
412 precondition.resource_changed on ETag-protected edits
422 validation errors on creation/replacement
503 provider dependency unavailable during provider-subject lookup/rebinding
ambiguous transport outcome for Idempotency-Key creations without issuing a second logical command
```

Offboarding confirmation must communicate its destructive access consequence. Re-enable must communicate that memberships/direct RoleAssignments/sessions are not restored.

## 9. B10-A1 — material open assumption

```text
B10-A1
Paginated browse is sufficient for Launch V1 administration of Users, Areas and Groups without collection-specific global search/filter operations.
```

Current authority provides paginated collection reads but no accepted `q`/search/filter contract for these collections.

Forbidden workaround:

```text
filter only the currently loaded page while presenting the control as global search
```

P8 falsification rule:

```text
if realistic operator use shows paginated browse sufficient
→ CURRENT AUTHORITY CONFIRMED

if realistic operator use shows finding a target is materially impractical
→ B10 UPSTREAM FINDING
→ stop only affected B10 scope
→ prove human need / scale problem
→ reopen smallest owning Product/read-contract decision
→ never invent convenience API from screen shape
```

The assumption remains OPEN through initial P8 operation. It cannot receive `ACCEPT_FOR_LOCK_WITH_LATER_PROBE` except by explicit operator disposition at lock time.

## 10. Global Maximum / YAGNI disposition

Deliberately excluded from B10 candidate:

```text
administrative dashboard metrics
generic CRUD framework
bulk actions
custom Role/Permission editor
client permission matrix
multi-company/tenant switching
identity-provider sync center
provider group/role mirroring
collection search/filter API not already authorized
GroupMembership administration
RoleAssignment administration
Document Governance administration
```

These exclusions remove accidental complexity without suppressing an already-proven Launch requirement.

## 11. P8 target

Canonical next Evidence, after review of this written candidate:

```text
docs/work/current/t11-b10-organization-administration-p8.html
```

It must be browser-operable low-fidelity HTML/CSS/vanilla JS with deterministic local fixtures only. It is disposable planning Evidence, never production React/API integration/business authority.

Material interactions to make operable include:

```text
Organization section switching
collection pagination
list → selected-detail transitions
create User provider-subject lookup/selection flow
provider-binding lookup/selection/replacement flow
create Area / create Group
separate User Profile / Binding / Eligibility actions
Company/Area/Group edits
stale ETag conflict presentation/recovery path
User disable confirmation + re-enable explanation
Group delete dependency failure
responsive navigation/detail behavior
keyboard/focus-visible interaction structure
```

P8 exits only after operator operation and explicit disposition. Assistant/reviewer cannot set `LOCKED`.
