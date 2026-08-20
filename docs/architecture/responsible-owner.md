# R10-T3 — D4 Responsible-Owner Eligibility Amendment

> **Status:** ACTIVE / OPERATOR-RATIFIED BOUNDED T3 AUTHORITY  
> **Ratified:** 2026-08-18  
> **Scope:** `r10-t3-authorization-audit-enforcement.md` §9 — phrase `eligible target User` only  
> **Implementation:** BLOCKED

This amendment exists because the T6 bounded coherence delta proved that T3's phrase `eligible target User` was underspecified for responsible-owner assignment. It changes **no Role, Permission, bundle, scope, authorization equation, lifecycle predicate or Audit rule**.

For this one meaning, this page is more specific than `r10-t3-authorization-audit-enforcement.md` and therefore controls.

## Decision

For create-time deliberate responsible-owner selection and later responsible-owner replacement:

```text
eligible target User
=
existing MetalDocs User
+ same Company
+ current User eligibility = ENABLED
```

No additional eligibility is implied.

Specifically, responsible-owner assignment:

```text
does NOT grant Role
does NOT grant Permission
does NOT import or depend on Keycloak/provider roles/groups
does NOT require the target already to hold document.edit
does NOT by itself grant document.read_working/document.edit/document.submit
```

The responsible-owner relation remains Controlled Documents current relationship truth. T3 consumes that relationship as one domain predicate, but every actual action still requires the normal authorization equation:

```text
current User eligibility
+ current RoleAssignments / GroupMembership grants
+ scope match
+ Controlled Documents relationship/state predicate
= allow / default deny
```

An already-recorded disabled/offboarded responsible User remains truthful historical/current relationship data until an authorized manager changes the relation, but a disabled User cannot be newly selected as responsible owner.

## Create-time behavior

Ordinary author creation continues to default the actor as responsible owner.

An actor with `document.owner.manage` may deliberately select another target only when the target satisfies the eligibility definition above and the actor's own `document.owner.manage` scope matches the new Document.

## Replacement behavior

```text
document.owner.manage
+ matching scope
+ target = existing ENABLED User in same Company
+ T6 current-resource If-Match precondition
→ replace current responsible owner
```

The replacement must serialize correctly with target User offboarding/eligibility change so that:

```text
assignment linearizes first
→ assignment may commit while target is ENABLED
→ later offboarding may disable that User

offboarding linearizes first
→ assignment observes DISABLED
→ assignment fails closed
```

This extends T3 §11's existing eligibility-serialization law to responsible-owner assignment without changing its general concurrency model.

## Reopen trigger

Reopen only if a concrete product/regulatory requirement proves that responsible ownership itself must imply a specific professional qualification, role, employment relation or organizational membership distinct from current ENABLED MetalDocs User identity.

Do not infer such requirements from current provider roles/groups or implementation convenience.
