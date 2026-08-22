# T10 Fable independent review — Round 1

> **Evidence only — non-authoritative. This review branch must never merge.**

## Review identity

```text
Repository                developmentconexus-ops/MetalDocs
Gate                      T10 — Transition / Cutover
Candidate branch          arch/t10-transition-cutover
Exact candidate HEAD      0b90f26690b2b2bbf627f0c72283ff14c0ce9b84
Required candidate CI     #1153 SUCCESS
Candidate Draft PR        #158
Review branch             review/t10-fable
Round                     1 / ADVERSARIAL
```

## Read route

Read strictly:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ docs/work/current/t10-transition-cutover.md
→ only the exact accepted owner needed to challenge a concrete claim
```

Do not recursively read history/legacy/closed PRs.

## Fixed envelope

```text
T1→T9                         CLOSED / OPERATOR-RATIFIED / INTEGRATED
T10                           OPEN / ACTIVE
T11→T12                       NOT OPEN
Product implementation         BLOCKED
legacy implementation live tree ABSENT
application operations         78
operation 79                   ABSENT
historical business corpus     NONE
required pre-R10 business corpus NONE
```

T7 remains binding: Launch has no historical business migration requirement. A review finding may not manufacture migration/compatibility machinery merely for symmetry.

## Candidate under attack

Selected posture:

```text
ONE-WAY GREENFIELD R10 ACTIVATION
+
PRIVATE PREPARATION
+
PROOF BEFORE AUTHORITATIVE BOOTSTRAP
+
FIRST-AUTHORITATIVE-MUTATION = POINT OF NO RETURN
+
SERVING ACTIVATION ONLY AFTER AUTHORITATIVE BASELINE EXISTS
+
ONE BUSINESS AUTHORITY AT A TIME
+
FAIL-CLOSED / R10 RECOVERY AFTER AUTHORITY BEGINS
```

Monotonic barriers:

```text
B0  source truth classified
B1  target privately prepared
B2  target proven while still non-authoritative
B3  first authoritative R10 Product mutation committed / point of no return
B4  canonical serving authority activated
```

## Adversarial questions

Do not optimize for agreement. Try to falsify the candidate.

Challenge especially:

1. **B0 source truth**
   - Can surviving DB/object/IdP/deploy state be incorrectly discarded even though it is actually authoritative?
   - Is the stop/reopen rule sufficient if a real pre-R10 business corpus appears?
   - Does the candidate silently assume an external estate that may not exist?

2. **B1/B2 private target and proof**
   - Can target proof require Product truth that would already cross B3?
   - Is there any accepted T8/T9 proof that cannot truthfully run while the target is still non-authoritative?
   - Does proof-before-B3 accidentally require a synthetic bootstrap or fake fixture that only proves itself?

3. **B3 point of no return**
   - Is “first authoritative Product mutation” the smallest mechanically identifiable boundary?
   - Can Company/User/ProviderSubjectBinding/configuration bootstrap commit partially or ambiguously so that the system crosses B3 without a coherent recoverable baseline?
   - Is any accepted Product truth established outside the definition of B3?
   - Is there a hidden need for an activation marker/table/state not already owned by Product? If yes, prove why rather than inventing it.

4. **B3→B4 serving activation**
   - After B3 but before B4, what failure modes exist while business truth exists but normal serving is still disabled?
   - Can external OIDC/DNS/ingress/config changes create competing authority or an unrecoverable half-cutover?
   - Is one business authority preserved even during retries/restarts?

5. **Rollback versus recovery**
   - Before B3, is reset/retry safe under every accepted target mechanism?
   - After B3, is destructive rollback truly forbidden in every path?
   - Can a binary/config rollback corrupt or misinterpret already-committed R10 state?
   - Are backup/restore/session invalidation/privacy/security readiness correctly inherited rather than weakened?

6. **Content / River / exact bytes**
   - Can content or River state become accidentally authoritative before B3?
   - Can pre-B3 content survive and later be mistaken for governed R10 content?
   - Can post-B3 cleanup remove required exact content or durable work evidence?

7. **Cleanup**
   - Are deletion preconditions strong enough to prevent removing recovery/provenance/security-critical resources?
   - Does cleanup need a waiting/observation barrier, or would that be ceremonial overengineering?

8. **Hidden migration/compatibility leakage**
   - Find any place where the candidate implicitly requires dual write, old/new reconciliation, legacy read fallback, schema translation, generic ETL/import, or historical reconstruction.
   - If such machinery is actually required, identify the exact accepted property that requires it.

9. **Stage leakage**
   - Identify any T11 implementation-decomposition or T12 readiness work smuggled into T10.
   - T10 may define required transition semantics/proof barriers, but must not implement or decompose Product code.

10. **Closed Product envelope**
    - 78 operations must remain exactly 78.
    - operation 79 must remain absent.
    - no new Permission, semantic owner, Product state, API operation, persistence authority or runtime capability may be invented solely for cutover.

## Finding standard

Classify every finding:

```text
MATERIAL
  candidate can violate accepted authority, lose/corrupt Product truth,
  cross an irreversible boundary ambiguously, or requires an upstream reopen

MINOR
  bounded precision/traceability issue that changes no accepted architecture

NOTE
  non-blocking observation
```

For each MATERIAL finding include:

```text
exact candidate claim
exact accepted authority/property contradicted or uncovered
reachable failure sequence
smallest correction or owning-authority reopen
why the finding is not preference/framework fashion/hypothetical scale
```

## Regression envelope

A correction must preserve unless a MATERIAL falsifier proves otherwise:

```text
barriers                         exactly B0→B4 / monotonic
historical business migration   absent
business authority              singular
application operations          78
operation 79                    absent
new Permission                  none
new semantic owner              none
T11/T12                         not open
Product implementation          blocked
```

## Required verdict

End exactly:

```text
VERDICT = CONVERGED | NOT CONVERGED
MATERIAL findings = N
Round 2 justified = YES | NO
```

Write review Evidence below this line only.

---
