# T11 — B11 P8 R1 Operator Finding

> **Status:** OPERATOR EVIDENCE / P8 R1 REVISE / UPSTREAM FINDING.  
> **Block:** B11 — Access Administration.  
> **Exact P8 R1 artifact:** `docs/work/current/t11-b11-access-administration-p8.html`.  
> **Exact P8 R1 Git blob:** `c04ff56efa7aae72c59dc0ee9c4d56c9357c4de7`.  
> **Implementation:** BLOCKED.  
> **Authority:** Evidence only; the durable owning decision is `../../decisions/access-assignment-read.md`.

## Operator Evidence

During the B11 P8 R1 walkthrough, the operator identified that the current wireframe exposes membership and individual RoleAssignment mutations but does not make the existing access model human-readable enough to administer.

The missing operational questions are:

```text
For a selected Group:
- which Areas / Company scope does this Group have access to?
- which fixed Role does it hold in each scope?
- what does each fixed Role mean?

For a selected Area:
- which Groups have Area-scoped grants here?
- which Users have direct Area-scoped grants here?
- which Company-scoped grants also apply across this Area?
```

The operator explicitly agreed that a Group may legitimately hold different Roles in different Areas. Therefore a single `Group.area_id` would be semantically wrong.

## P8 R1 disposition

```text
B11-A1 findability / inspectability
  FAIL
  The raw paginated RoleAssignment ledger does not make access configuration discoverable by Group or Area/Scope.

B11-A2 membership consequence
  FAIL
  Safe membership administration requires seeing the selected Group's canonical access footprint before adding/removing a member; generic consequence copy alone is insufficient.

B11-A3 access explanation
  REFINED / PARTIALLY FALSIFIED
  A full effective-access / "why can this User do X?" engine is NOT proven necessary.
  What is proven necessary is complete human-recognizable inspection of canonical RoleAssignments by Group and by Scope.
```

## UPSTREAM FINDING

```text
B11-F1 — Access Assignment Read Precision
```

Protected outcome:

> An access administrator must be able to inspect canonical access configuration by Group and by Area/Company scope before making security-bearing membership/grant decisions, without the browser manufacturing completeness or effective Authorization truth.

Smallest owning correction:

```text
refine existing op31 listRoleAssignments
+ server-side exact filters for User / Group / Scope / Role
+ human-recognizable subject/scope projection in read items
+ preserve existing mutation model and role vocabulary
```

Explicit non-solutions:

```text
NO Group.area_id
NO client crawl + post-filter over incomplete pages
NO custom Role/Permission editor
NO effective-permission matrix computed in frontend
NO new effective-access engine
NO new application operation merely for screen convenience
```

## Bounded rebaseline law

Only B11 is affected. B01-B10 remain accepted/locked/integrated. B12, FP2/P11, T12 and Product implementation remain unopened/blocked.

P8 R2 must not begin until B11-F1 is durably reconciled in the owning read authority. After reconciliation, P7 is rebaselined only where this Finding changes the Access information architecture.
