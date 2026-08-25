# T11 — B11 Access Administration — PR #173 review finding

> **Status:** MATERIAL P8/P9 FINDING / BOUNDED REOPEN.  
> **Trigger:** unresolved PR #173 review thread `PRRT_kwDORsgkIs6cGr9V`.  
> **Affected baseline:** operator-LOCKED P8 R5 + post-LOCK P9 only where collection continuation was not actually operable.  
> **Unaffected:** B11-F1, `Por Área / Grupos / Funções` IA, R4/R5 frame, Role meaning, membership consequence, exact revoke, Idempotency recovery, Authorization authority, 89-operation census.

## Evidence

The review correctly identified a contradiction between R5/P9 and accepted paginated reads:

```text
op27 listGroupMembers
  P9 declared Group member truth READY
  but R5 member list had no continuation control

op6 / op22 / op16 in Grant composer
  P9 bound User / Group / Area identity selection to paginated supporting reads
  but R5 rendered those collections as complete in plain <select> controls
```

The add-member User picker already had visible pagination in R5, but its continuation/failure semantics were not explicit in P9. Role selection remains a fixed, non-paginated six-Role Product vocabulary and is unaffected.

## Root cause

R5 proved filtered RoleAssignment pagination but reused a convenient in-memory fixture representation for supporting identity reads. P9 then described those supporting reads as sufficient without proving how the user reaches identities or memberships beyond the first page.

That creates a real implementation fork:

```text
omit valid later-page entity
OR
hidden full-collection crawl
```

Both violate the accepted experience/authority boundary.

## Global-Maximum disposition

```text
CURRENT STRUCTURE CONFIRMED
+ bounded P8/P9 correction
```

No backend/API reopen is justified. Existing operations already expose the required paginated truth.

Correct only:

```text
Group members      op27 visible cursor traversal
Add-member Users   op6 visible cursor traversal + continuation failure
Grant User subject op6 visible cursor traversal
Grant Group subject op22 visible cursor traversal
Grant Area scope   op16 visible cursor traversal
```

Do not add search, operation 90+, client crawl, total-count dependency, `Group.area_id`, effective-access engine, custom Role/Permission editing or a new UI topology.

## Proof strategy

P8 R6 must make the defect falsifiable with deterministic data beyond page one:

1. selected Group has more members than one page;
2. operator reaches/removes a member on a later op27 page;
3. grant User picker reaches and reviews a later-page User;
4. grant Group picker reaches and reviews a later-page Group;
5. grant Area picker reaches and reviews a later-page Area;
6. one-shot continuation failure preserves the currently loaded page and does not label it complete;
7. contextual Group/Area grant preselection still works;
8. add-member User picker still reaches a later-page User;
9. no new search/backend/effective-access authority appears.

## Re-LOCK boundary

Only the affected pagination/selection surfaces are reopened. The operator must operate and explicitly re-LOCK this bounded delta before the PR thread can be resolved and the B11 acceptance increment can return to merge-ready state.
