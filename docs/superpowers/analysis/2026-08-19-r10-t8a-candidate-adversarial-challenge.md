# R10-T8A — Candidate Adversarial Challenge

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **INTERNAL CHALLENGE OF T8-A CANDIDATE**  
> **Date:** 2026-08-19  
> **Candidate under challenge:** `2026-08-19-r10-t8a-technical-authority-legacy-disposition-candidate.md`  
> **Implementation:** BLOCKED

This review attacks the preferred candidate rather than defending it.

## 1. Challenge: is “clean-slate inside current project” merely another sunk-cost compromise?

**Attack:** If the current repository is deeply contaminated by legacy architecture, perhaps a new repository is structurally cleaner and Candidate B preserves the local maximum at the repository boundary.

**Evidence:**
- accepted R10 authority does not identify repository layout/source-control history as a semantic boundary;
- Product Contract/T1→T7 do not require or forbid one repository;
- current repo already contains the durable R10 authority, Method mirror, verification control plane, OpenAPI/codegen mechanisms, tests/evidence and Git provenance needed for T10;
- no evidence shows the repository container itself prevents arbitrary package/schema/frontend/runtime replacement;
- T8-B is still free to replace package/module layout completely and T8-G is free to change process topology.

**Verdict:** Candidate survives. `current repository` is not being preserved as architecture; it is the change/control container. Reopen if T8 later proves the repository/toolchain itself prevents isolation or execution.

## 2. Challenge: does PRESERVE PostgreSQL prematurely lock a technology?

**Attack:** Global Maximum should not preserve Postgres because it is already implemented.

**Evidence:** PostgreSQL is not merely inherited current state. T2 explicitly ratified PostgreSQL READ COMMITTED + narrow serialization/OCC posture, and T5 selected PostgreSQL/River durable-job coherence. T4 backup/restore reasoning also consumes the product-state DB recovery point.

**Verdict:** Candidate survives. T8-A is preserving an already-ratified upstream decision, not rediscovering it. Reopening Postgres in T8-A without contradictory evidence would violate revalidation law.

## 3. Challenge: does PRESERVE River prematurely lock async implementation?

**Attack:** Current jobs code is messy; perhaps River should be removed with it.

**Evidence:** T5 independently selected PostgreSQL/River as the smallest durable-job mechanism after semantic/effect analysis. Current jobs registry/process shape is separately classified REWRITE. Candidate preserves only River mechanism, not old jobs/process topology.

**Verdict:** Candidate survives. Preserve upstream decision; rewrite consumers/wiring.

## 4. Challenge: is contract-first/codegen preservation just attachment to current tooling?

**Attack:** Generated packages and current module-tag protocol are cumbersome; perhaps a hand-written API would be simpler.

**Evidence:** T6 explicitly requires OpenAPI 3.0.3 and generated Go/TypeScript boundaries. Current verifier has real codegen-drift/contract-sync checks. Candidate does not preserve tag-per-legacy-module or current generated package layout.

**Verdict:** Candidate survives. Preserve the contract/generated-boundary property; rewrite current wire document and realization.

## 5. Challenge: is `tools/verify` itself accidental complexity?

**Attack:** The verifier is large and carries historical guards; rebuilding the architecture might justify deleting it and using ordinary CI scripts.

**Evidence:** The verifier has an explicit root cause: local/CI divergence and silent inert/skipped gates previously produced false confidence. Its registry centralizes the definition of verified, distinguishes PASS/FAIL/SKIP, preflights toolchains, enforces ordering/infra, audits CI mapping and requires negative fixtures/closed waivers for repo-authored blocking controls. These properties are independent of current business module topology.

**Counter-risk:** current registry is large and some checks encode superseded architecture. Preserving the whole registry unchanged would fossilize legacy policy.

**Verdict:** Candidate survives with existing distinction: `verification control-plane model = PRESERVE`; `target-specific checks = REFINE/REWRITE`. T9/T12 may simplify the check set but must preserve equivalent proof strength.

## 6. Challenge: does preserving least-privilege DB identity overfit the current db-provision implementation?

**Attack:** Maybe a separate runtime DB role and provisioning identity are unnecessary in a single-company product.

**Evidence:** Company count is unrelated to database least privilege. Serving processes needing DML but not schema ownership is a security/trust-boundary property. Current implementation proves the role split is workable. Candidate does not require a separate long-lived db-provision service; T8-G/T10 rederive choreography.

**Verdict:** Candidate survives as a property, not topology.

## 7. Challenge: is “reuse algorithms/tests” a loophole that lets legacy semantics creep back in?

**Attack:** Selective reuse sounds safe but can become an excuse to copy old code wholesale.

**Required tightening:** Reuse must satisfy all of:

```text
named current R10 consumer
no legacy semantic authority in its public contract
no dependency inversion against T8 target
proof/test asserts target property rather than legacy shape
cheaper/simpler than rewrite after transition cost
```

Otherwise classify REWRITE even if the implementation is locally high quality.

**Verdict:** Candidate survives, with this five-part reuse gate binding to its interpretation.

## 8. Challenge: should one Go module be PRESERVE?

**Attack:** Current single module may hide coupling and could constrain package isolation.

**Evidence:** No target property requires one or multiple Go modules. T8-B owns physical package/repository layout.

**Verdict:** Do **not** preserve yet. Current one-Go-module fact remains `CURRENT-STATE / PRESERVE CANDIDATE ONLY`, and T8-B may choose otherwise. Candidate already leaves this undecided.

## 9. Challenge: should React Router / TanStack Query be preserve candidates?

**Attack:** They are familiar and installed, but T6 ratified frontend lenses, not these libraries.

**Evidence:** No upstream authority requires them. Current frontend structure demonstrates them but also demonstrates strong coupling to legacy features.

**Verdict:** Do **not** preserve in T8-A. Keep `CURRENT-STATE ONLY`; T8-F compares the smallest frontend realization. Generated TypeScript transport boundary remains upstream-required, distinct from query/router libraries.

## 10. Challenge: does rejecting full greenfield ignore the psychological/structural benefit of a clean break?

**Attack:** New project tree could reduce accidental imports and deletion burden.

**Evidence:** T8-B/T10 can achieve hard structural cutover inside the existing repo, including deleting old trees and creating new roots, while retaining Git provenance and verification authority. No current evidence shows a second repository improves a required property enough to justify operational/provenance cost.

**Verdict:** Candidate survives. Full greenfield remains a valid reopen outcome if T8-B/G prove substrate obstruction, not a default purity move.

## 11. Challenge: is exact foreign-SQL remeasurement required before T8-A closure?

**Attack:** TRRB marked old 55/12 counts LAST-REPRODUCED; perhaps T8-A cannot classify boundaries without fresh exact counts.

**Evidence:** Current `tools/cilint/baseline.json` directly proves live cross-module raw-SQL reads/writes across multiple owner tables. The target ownership topology is independently different from the legacy module topology. Whether current leakage is 40, 67 or 80 statements does not change the T8-A target-disposition decision.

**Verdict:** Exact count is **not load-bearing to T8-A Global Maximum**. Preserve LAST-REPRODUCED label; remeasure later if T10 transition sizing/proof needs the exact current blast radius.

## 12. Challenge: are there hidden requirements to preserve current data/API behavior?

**Attack:** Even DEV systems may have user habits/tests that make compatibility valuable.

**Evidence:** T6 explicitly permits pre-launch `/api/v1` rebuild in place with no compatibility; T7 proves current business data is disposable; Product Contract is Launch authority. Tests against superseded behavior are evidence, not requirements.

**Verdict:** Candidate survives. User-observable behavior must match Product Contract/T6, not legacy compatibility.

## 13. Adversarial outcome

No material challenge disproves Candidate B.

The challenge **tightens** it with the following law:

> **Selective reuse is allowed only when a current R10 consumer exists, the reusable unit's public contract is free of legacy authority, dependency direction fits the target, its proof asserts the target property, and reuse remains simpler than rewrite after transition cost.**

Therefore final T8-A candidate direction remains:

```text
clean-slate physical target freedom
+ selective proof-backed mechanism reuse
- legacy shape inheritance
- full-greenfield purity reset without evidence
```

No T8-B→G physical design has been selected. Implementation remains BLOCKED.
