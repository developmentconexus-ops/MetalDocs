# Rebaseline Decision Registry — D4 Bounded Amendment

> **Status:** ACTIVE / OPERATOR-RATIFIED REGISTRY RECONCILIATION  
> **Ratified:** 2026-08-18  
> **Parent registry:** `wiki/architecture/rebaseline-decision-registry.md`  
> **Detailed authority:** `wiki/architecture/r10-t3-d4-responsible-owner-eligibility-amendment.md`  
> **Implementation:** BLOCKED

This file is a bounded amendment to the current Rebaseline Decision Registry. It changes only the responsible-owner target eligibility precision exposed by the T6 bounded coherence delta. All parent-registry dispositions not named here remain unchanged.

Until the next full registry consolidation, this file is read immediately after `rebaseline-decision-registry.md` and is authoritative only for the amended meaning below.

## Amended current meaning

### DOC-08 — responsible owner

Parent meaning preserved:

```text
Responsible owner is current mutable Document relationship;
historical actions remain actor-bound.
```

Bounded precision added:

```text
new responsible-owner target eligibility
=
existing MetalDocs User
+ same Company
+ current eligibility = ENABLED
```

Assignment does not grant Role or Permission and does not depend on provider roles/groups. Actual document actions still require current T3 grant + scope + Controlled Documents predicate authority.

### AZ-16 — authoring relationship predicate

Parent meaning preserved:

```text
ordinary author working authority is bounded by current responsible-owner relationship unless actor also has document.owner.manage
```

Bounded precision:

```text
responsible-owner relationship itself is not an access grant
responsible-owner assignment does not manufacture document.edit/read_working/submit authority
```

## Cross-stage effect

```text
T1 = unchanged
T2 = unchanged
T3 = bounded §9 precision only
T4 = unchanged
T5 = unchanged
T6 = consumes this definition for create/options + owner replacement
```

No role catalog, permission catalog, bundle, scope matrix or provider-claim rule is reopened.
