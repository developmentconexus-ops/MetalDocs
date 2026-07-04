# F4.3 evidence — concurrency-idiom decision (ADR-record the split)

> **Feature type:** decision-only (documentation). No product code, no tests, no build steps.
> **Outcome:** contract §3.4/§3.5 ("migrate templates to If-Match") **superseded** after verification
> proved its premise false; corrected to **ADR-record the intentional split** (ADR 0066) via HS-7.

## What shipped

| Artifact | State |
|---|---|
| `wiki/decisions/0066-optimistic-concurrency-transport-split.md` | ADR: intentional documents=If-Match / templates=body split; If-Match named target; full unification deferred to its own cross-module change (not M4). |
| `../validation-contract.md` §3.7 | HS-7 erratum: loud, dated (2026-07-04), non-silent re-open; §3.4/§3.5/§3.6 struck-through + superseded; corrected binding decision + exit criteria. |
| `spec.md`, `plan.md` (this feature) | Updated to the corrected decision. |
| documents + templates product code | **Unchanged** — no wire change in M4. |

## Verification (the analysis that flipped the decision)

Commands run 2026-07-04, real output:

```
$ grep -rln "If-Match\|IfMatch\|parseIfMatch" internal/modules/templates/ frontend/apps/web/src/features/templates/
(empty — zero templates If-Match usage, BE + FE)

$ # tags of the 3 If-Match OpenAPI operations (api/openapi/v1/openapi.yaml)
  near line 2572 → tags: [documents]  (operationId: finalizeDocument)
  near line 3371 → tags: [approval]   (operationId: submitDocumentForApproval)
  near line 3877 → tags: [approval]   (operationId: updateApprovalRoute)
(none tagged [templates])

$ grep -rln "expected_lock_version\|ExpectedLockVersion" internal/modules/templates/ frontend/apps/web/src/features/templates/
internal/modules/templates/.../routes_schema.go        (UpdateSchemas — templates' ONLY OCC write)
frontend/apps/web/src/features/templates/api/templates.ts
... (self-consistent with lock_version + stale_lock_version)
```

**Conclusion:** the original premise ("templates is the minority straggler near the If-Match convention")
is false. templates has zero If-Match; `expected_lock_version` is its sole, self-consistent OCC idiom;
the cited "system-wide If-Match standard" is CON-01 (`wiki/modules/documents.md`), a documents-**internal**
decision, not a cross-module ADR. Migrating templates would create a new cross-module standard, not finish
a convergence — out of scope for a versioning-kernel-correctness milestone.

## Subagent disposition

F4.3 implementation subagent (a98c531f) returned **STOP — architecture contradiction**, made **no edits**,
and recommended either dropping the migration or opening a separate system-wide ADR milestone. Main session
independently re-ran the grep/tag verification (above), confirmed the finding, and took the contract
§3.5-permitted fallback.

## HS-7 disposition (headline for the HS-1 operator gate)

- The contract §3 decision was operator-delegated ("best solution, full analysis"). Verification changed
  the analysis outcome, so the decision changed with it.
- Re-open handled per the contract §0 HS-7 rule: **loud dated erratum, not a silent edit to match code.**
- Full ratification is **deferred to the M4 HS-1 operator gate** — nothing pushed or merged before it.
  Operator may reject/redirect (e.g. mandate the templates migration now, or schedule it into a named
  milestone) at that gate.

## Bounded defers

| Defer | Rationale | Trigger / owner |
|---|---|---|
| Full unification onto `If-Match` (migrate templates `UpdateSchemas` contract-first) | Cross-module wire refactor; out of scope for M4 correctness; ADR 0066 names the target + charter | M9 governance-hygiene or a standalone milestone; retires ADR 0066 when done |
