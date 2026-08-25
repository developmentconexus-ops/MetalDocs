---
id: access-assignment-read
kind: authority
owner: architecture
summary: Bounded T11 authority for human-recognizable RoleAssignment inspection by subject and scope discovered during B11 frontend planning.
---

# Access Assignment Read — bounded T11 authority

> **Status:** OPERATOR-RATIFIED / BOUNDED T11 REOPEN.  
> **Ratified:** 2026-08-24; revalidated 2026-08-25 against restored local methods.  
> **Trigger:** B11 functional P8 operator walkthrough.  
> **Method:** `docs/development/engineering-method.md` v1.0.0 + `docs/development/frontend-product-experience-planning-method.md` v2.3.  
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Authority and supersession

This page is the single bounded current authority for the B11 access-assignment read precision exposed by frontend planning.

It does not replace Product/T1→T10 wholesale. It supersedes only conflicting current-tense clauses concerning operation 31 `listRoleAssignments`, its first-page query, and its human-recognizable read projection in:

```text
../product/journeys.md
../architecture/wire-contract.md
../architecture/frontend.md
api-operation-census.md
```

All unchanged Authorization equations, Role/Permission vocabulary, GroupMembership semantics, mutation contracts, Audit, offboarding, scope compatibility and security laws remain current.

Historical stage snapshots remain truthful for the stage at which they were ratified.

The 2026-08-25 repository-method restoration changed operating mechanics only. Revalidation against local Engineering Method v1.0.0 and Frontend Product Experience Planning Method v2.3 found no Product, Authorization, wire, UX, proof or reopen-trigger contradiction in this bounded decision.

## 2. Proven human jobs

B11 must make the already-accepted access model inspectable through canonical configuration, not through a browser-computed effective-permission graph.

### Group access footprint

For one selected Group, an administrator must be able to inspect every current RoleAssignment directly granted to that Group across:

```text
COMPANY
AREA(area_id)
```

A Group may hold different Roles in different Areas and may also hold Company-scoped Roles.

Therefore:

```text
Group identity != one Area
Group.area_id does not exist
```

### Area / scope access configuration

For one selected Area, an administrator must be able to inspect:

```text
A. RoleAssignments scoped exactly to that Area
B. Company-scoped RoleAssignments, shown separately, because Company scope also applies across the Area
```

The UI must not merge these two sets into a fake single-scope record. Each row remains one canonical RoleAssignment with its actual scope.

### Fixed Role meaning

Existing `listRoles` remains sufficient to explain each fixed Role through server-returned:

```text
RoleView.code
RoleView.permissions
RoleView.allowed_scope_kinds
```

No Role/Permission editor is introduced.

## 3. Global-Maximum boundary

The smallest sustainable correction is:

```text
refine existing operation 31
+ exact server-side filters before pagination
+ human-recognizable read enrichment
```

Explicitly not required by current Evidence:

```text
new application operation
effective-access engine
"why can User X do permission Y?" troubleshooting engine
materialized permission matrix
custom Role / Permission editing
Group organizational Area ownership
browser crawl of all RoleAssignment pages
client-side post-filter presented as complete
```

A later effective-access/troubleshooting capability requires a separate material Product/read decision with a proven human/security consumer.

## 4. Application-operation delta

No operation is added or removed.

```text
31  GET /api/v1/role-assignments
    operationId listRoleAssignments
    REFINED — same semantic collection/read owner
```

Current cross-contract census remains:

```text
application operations           89
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4
```

## 5. op31 first-page query

When `cursor` is absent, operation 31 admits:

```text
user_id?       Uuid
group_id?      Uuid
scope_kind?    company | area
area_id?       Uuid
role?          RoleCode
limit?         integer 1..100; default 20
```

Filter laws:

```text
user_id + group_id together
  -> 400 request.invalid

scope_kind=area
  -> area_id REQUIRED

scope_kind=company
  -> area_id FORBIDDEN

scope_kind absent
  -> area_id FORBIDDEN
```

All supplied filters are conjunctive.

Examples:

```text
?group_id=<G>
  every current RoleAssignment whose subject is exactly Group G

?user_id=<U>
  every current direct RoleAssignment whose subject is exactly User U

?scope_kind=area&area_id=<A>
  every RoleAssignment scoped exactly to Area A

?scope_kind=company
  every Company-scoped RoleAssignment

?role=approver
  every RoleAssignment using the fixed approver Role
```

Filters execute server-side before seek pagination.

Collection order remains:

```text
assignment_id ASC
```

A syntactically valid filter identity that currently matches no disclosable assignment returns an ordinary empty page. Identity existence remains owned by Organization reads; this collection does not become a generic identity resolver.

## 6. Cursor law

The existing global pagination law remains current.

For a continuation request:

```text
cursor + optional limit only
```

The cursor authenticates:

```text
operationId
+ normalized op31 filters
+ seek position
```

Repeating `user_id`, `group_id`, `scope_kind`, `area_id` or `role` with a cursor is `400 request.invalid`.

Changing `limit` on continuation remains permitted by the existing global pagination law.

## 7. Human-recognizable read projection

Mutation/input identity remains ID-based. Read enrichment does not transfer identity ownership to Authorization.

Add read-only reference:

```text
GroupReference {
  group_id: Uuid,
  name: ShortText
}
```

Operation 31 returns a purpose-built read subject:

```text
RoleAssignmentSubjectView
  { kind:user,  user:UserReference }
  { kind:group, group:GroupReference }
```

and purpose-built read scope:

```text
RoleAssignmentScopeView
  { kind:company }
  { kind:area, area:AreaReference }
```

Refined operation-31 item:

```text
RoleAssignmentView {
  assignment_id: Uuid,
  subject: RoleAssignmentSubjectView,
  role: RoleCode,
  scope: RoleAssignmentScopeView
}
```

Existing mutation command remains unchanged:

```text
CreateRoleAssignmentRequest {
  subject: RoleAssignmentSubject,   // id-only USER | GROUP
  role: RoleCode,
  scope: RoleAssignmentScope        // COMPANY | AREA(area_id)
}
```

No mutable display label becomes Authorization authority. User/Group/Area labels in op31 are bounded current recognition data for administration only.

## 8. B11 consumption law

### Group lens

For selected Group G:

```text
listRoleAssignments?group_id=G
```

The result is the Group's canonical direct RoleAssignment footprint. It does not expand Group members into per-User effective access.

### Area lens

For selected Area A, the frontend uses two explicit canonical traversals:

```text
Area-scoped
  listRoleAssignments?scope_kind=area&area_id=A

Company-wide
  listRoleAssignments?scope_kind=company
```

The two result sets are rendered in separate labeled regions.

This is not an effective-access engine. It shows canonical grants whose real scopes are already authoritative. It does not answer whether an arbitrary User currently receives a Permission after GroupMembership expansion and domain predicates.

### Role lens

`listRoles` remains read-only fixed-role meaning. A Role may link to filtered canonical assignments through:

```text
listRoleAssignments?role=<RoleCode>
```

No Role mutation is added.

## 9. Security / authority invariants preserved

This precision does not reopen:

```text
access.manage owns GroupMembership + RoleAssignment mutation
Role vocabulary is static and product-owned
Permission vocabulary is static and product-owned
Role grants are additive and non-hierarchical
subject = User | Group
scope = Company | Area
provider roles/groups/claims are never canonical Authorization
frontend never evaluates Authorization as authority
commands always recheck current truth
GroupMembership add / direct User grant serialize against offboarding
User re-enable never resurrects old grants/memberships
```

Read filters never grant access and never widen current disclosure.

## 10. Proof strategy

The B11 functional P8 must be capable of falsifying this decision with deterministic fixtures.

Required probes:

```text
1. one Group has different Roles in at least two Areas plus a Company-scoped Role;
2. Group view shows all those canonical assignments without browser post-filter over incomplete pages;
3. one Area view shows Area-scoped grants and Company-wide grants in separate regions;
4. Role meaning comes only from RoleView and remains non-editable;
5. revocation targets one exact assignment_id from the read projection;
6. pagination remains real under active filters;
7. no screen claims complete per-User effective access from memberships + grants;
8. no Group.area_id or equivalent single-Area ownership is introduced.
```

Earlier B11 P8/P9 work exercised this op31 decision, but later review exposed separate frontend interaction contradictions elsewhere in the B11 candidate. Those findings did **not** falsify the op31 precision above; final whole-B11 P8/P9 completion remains pending a clean B11 rebaseline.

A control/projection counts only if the P8 can make an omission or false-completeness defect visible to the operator.

## 11. Reopen triggers

Reopen only on material Evidence such as:

```text
operator cannot administer safely without per-User effective-access explanation;
real scale proves User/Group/Area selection itself requires new search/read capability;
a named compliance/security consumer requires access certification or permission troubleshooting;
Product changes the static Role/Permission model;
Group acquires a real organizational ownership concept independent of Authorization scope.
```

Preference for a richer enterprise-IAM console is not a reopen trigger.
