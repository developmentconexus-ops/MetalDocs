# T11 — B11 final adversarial challenge R1

> **Status:** NOT CONVERGED / SMALLEST P8-P9 REOPEN.  
> **Reviewer posture:** independent fresh read-only challenger.  
> **Reviewed locked blobs:** HTML `bbf4018739f6abdf4299adbeb8b056125cee0dd7`; CSS `3336919859fa9fefff7019525c251e20853d9bc9`; JavaScript `6c1e28dae16957a37f1e29d335d9e20a796e57c6`.  
> **Authority:** reviewer output is Evidence, not Product/roadmap authority.

## 1. Verdict and adjudication

The final R1 challenge returned **NOT CONVERGED**. Lead adjudication accepted both MATERIAL findings and the four lower-severity proof mismatches as valid candidate defects.

```text
M1  RoleView fixture bundles contradict fixed Permission authority       ACCEPTED / P8 REOPEN
M2  403/404 controls log text but do not operate denial/reconciliation   ACCEPTED / P8 REOPEN
I1  P9 Previous claim exceeds numeric fixture-window proof               ACCEPTED / P9 NARROW
I2  P9 claims operable 400/422 fixtures that do not exist                ACCEPTED / P9 NARROW
I3  responsive drawer lacks focus/Escape/underlying-focus management     ACCEPTED / P8 REOPEN
m1  P9 names op6 field `state` instead of `eligibility`                  ACCEPTED / P9 CORRECT
```

The operator's 2026-08-25 LOCK remains valid historical Evidence for its exact blobs, but those blobs cannot proceed to integration after material falsification. Any corrected P8 bytes require a new operator walkthrough and LOCK.

## 2. Four clean-rebaseline failures explicitly disposed

The reviewer independently closed the four findings that caused the clean rebaseline:

```text
visible op6/op16/op22/op27/op31 traversal; continuation failure preserves page    CLOSED
raw op6 page boundaries; DISABLED User visible and unavailable                    CLOSED
unknown membership; op28 201/204 reconciliation without complete cache            CLOSED
op32 initial success/completed replay/ambiguous same-key retry; mutations 1→1      CLOSED
```

The R2 correction must preserve all four closures.

## 3. Smallest correction boundary

Authorized by current method after material Evidence:

```text
P8
  exact six Role bundles from current Authorization authority
  operation-targeted /admin/access 403 denial state
  selected-Group 404 owner-truth reconciliation state
  mobile drawer aria-controls + focus entry/return + Escape + underlying inert

P9
  `eligibility` field name
  fixture-window Previous claim narrowed from production cursor-history proof
  operable failure-fixture claim narrowed to states actually exposed

P10
  semantic consolidation unchanged; identity/status revalidated only after re-LOCK
```

Not authorized:

```text
new operation
effective-access engine
custom Role/Permission editing
generic IAM/Admin framework
implementation, B12 or P11
```

## 4. Next gate

```text
correct and verify P8 R2
→ operator operates R2
→ operator-only re-LOCK on exact new blobs
→ reclose exact P9/P10
→ one fresh final adversarial challenge
```
