# T11 — B11 final adversarial challenge R2

> **Status:** NOT CONVERGED / OP32 AMBIGUITY SCOPE REOPENED.  
> **Reviewer posture:** new independent fresh read-only challenger.  
> **Reviewed locked blobs:** HTML `c0dfe7b942b83f53374307dbdb3d3524b7d47c69`; CSS `9ce012007613777187ae70956c2bfa09e7066c16`; JavaScript `3e923746cfea01142249b8166500833c807c5ce5`.  
> **Authority:** reviewer output is Evidence, not Product/roadmap authority.

## 1. Verdict and adjudication

The R2 final challenge returned **NOT CONVERGED** with one MATERIAL candidate defect and no other IMPORTANT/MINOR finding. Lead adjudication accepted it.

Reachable invalid path in the locked R2:

```text
commit command K1
→ response ambiguous / mutation count 1
→ click Revisar grant again
→ grantReview allocates fresh K2
→ confirm K2
→ duplicate semantic RoleAssignment / mutation count 2
```

The visible same-key retry path itself remained correct at 1→1, but the alternate fresh-review path meant the original op32 failure class was not globally closed.

## 2. Explicit dispositions

Challenge R1 findings independently closed by the R2 challenger:

```text
exact six Role bundles/scopes                                      CLOSED
operable /admin/access 403 denial                                 CLOSED
selected-Group 404 reconciliation                                 CLOSED
P9 Previous fixture-window versus production cursor precision      CLOSED
accurate operable failure-fixture coverage                         CLOSED
drawer aria-controls/focus/inert/Escape/return                     CLOSED
op6 `eligibility` field                                            CLOSED
```

Original clean-rebaseline findings:

```text
visible traversal / continuation page preservation                 CLOSED
raw op6 boundaries / DISABLED visible and unavailable              CLOSED
op28 201/204 without complete membership cache                     CLOSED
op32 demonstrated initial/completed/ambiguous same-key path        WORKS
op32 all reachable ambiguous-outcome paths                         REOPENED
```

No contradiction was found in authz/disclosure, fixed Role compatibility, fixture/production boundary, responsive/keyboard structure, P9/P10 ownership, operation census or anti-framework/Global Maximum posture.

## 3. Smallest R3 correction

```text
ambiguous result
→ mark command pendingAmbiguous
→ disable Subject/Role/Scope/review/confirm/close
→ prevent Escape close
→ leave same-key retry as the only reachable resolution
→ on stored-success replay, unlock close and remain terminal
```

No operation, Product meaning, backend authority or framework changes.

## 4. Next gate

```text
verify the alternate path cannot produce K2 / mutation 2
→ operator operates R3
→ operator-only re-LOCK on exact R3 blobs
→ reclose P9/P10 identity and ambiguity trace
→ fresh final challenge
```
