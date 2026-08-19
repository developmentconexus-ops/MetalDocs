# Current Agent Handoff

> **Last verified:** 2026-08-19  
> **Status:** ACTIVE — **T1→T8-B CLOSED / OPERATOR-RATIFIED; T8-C ACTIVE / GLOBAL MAXIMUM CANDIDATE MATERIALIZED / INDEPENDENT REVIEW NEXT; T8-D→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. `wiki/architecture/r10-technical-architecture.md` — sole status/next-action router
5. Product Contract REV001 + Whole-Product GCR + 4+1 ownership
6. T1→T8-B durable authorities
7. Decision Registry + amendments through T8-B
8. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
9. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
10. `docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-bootstrap.md`
11. `docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-global-maximum-candidate.md`
12. current interfaces/code only when a concrete T8-C evidence claim needs them

Do not route target design through superseded/historical architecture or current module/interface existence.

## Current checkpoint

```text
Product Contract                         REV001 / OPERATOR-APPROVED
Whole-Product GCR / 4+1 ownership       CLOSED / OPERATOR-APPROVED
T1→T8-B                                  CLOSED / OPERATOR-RATIFIED
Decision Registry                        CURRENT + amendments through T8-B
Post-T6 Stage-Decomposition GCR          RESTRUCTURE NOW / OPERATOR-RATIFIED
TRRB                                     CLOSED / OPERATOR-RATIFIED / PROMOTED

T8-C                                     ACTIVE / CANDIDATE MATERIALIZED / INDEPENDENT REVIEW NEXT
T8-D                                     NOT OPEN
T8-E                                     NOT OPEN
T8-F                                     NOT OPEN
T8-G                                     NOT OPEN
T8-H                                     NOT OPEN
T9                                       NOT OPEN
T10                                      NOT OPEN
T11                                      NOT OPEN
T12                                      NOT OPEN
implementation                           BLOCKED
```

## Binding execution law

> **No Writer task may contain a material architecture decision that should have been decided before execution.**

## Reference-backed decision law

For material technical decisions, use this order:

```text
MetalDocs semantic/product authority
→ DevelopmentConexus Engineering Method
→ current repository evidence
→ primary/current standards and official tool/library documentation
→ relevant reference products/patterns as falsification evidence
→ credible alternatives + Global Maximum
→ proof/adversarial review
→ operator ratification
```

Do not reinvent a technical mechanism when the selected language/database/standard/tool already provides the required property. Do not import an external "best practice" when it conflicts with MetalDocs authority. Reference-product private architecture is never inferred.

`AGENTS.md` requires Context7 for current external library/framework/SDK/API/CLI/cloud behavior when that behavior is load-bearing.

## T7 — CLOSED

```text
NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH
```

Current MetalDocs DB/content/history is DEV/test/throwaway and gives no business-data compatibility entitlement.

## T8-A — CLOSED

Durable authority:

`wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md`

Registry amendment:

`wiki/architecture/rebaseline-decision-registry-t8a-amendment.md`

Ratified law:

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

Current implementation remains evidence only. PRESERVE requires all five T8-A proofs.

## T8-B — CLOSED

Durable authority:

`wiki/architecture/r10-t8b-backend-module-package-topology.md`

Registry amendment:

`wiki/architecture/rebaseline-decision-registry-t8b-amendment.md`

Ratified Global Maximum:

```text
ONE GO MODULE FOR BACKEND GO CODE
+ OWNER-FIRST MODULAR MONOLITH
+ ONE IMPORTABLE PUBLIC SURFACE PER SEMANTIC OWNER
+ STATELESS APPLICATION LEAF ORCHESTRATION
+ ONE SEMANTIC INBOUND DOOR THROUGH APPLICATION
+ NON-SEMANTIC PLATFORM MECHANISMS
+ WIRING-ONLY COMPOSITION ROOT
+ CLOSED-WORLD / DEFAULT-DENY FIRST-PARTY DEPENDENCY GRAPH
```

Binding owner homes:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit — supporting evidence authority
```

Binding realization consequences:

```text
exactly one public package path per owner
owner-private decomposition is ungated and may evolve internally
no direct owner→owner imports
transport → application is the only semantic inbound door
application leaves are stateless; application→application imports forbidden
platform owns mechanisms, never semantic truth
composition owns construction only
foreign SQL / hidden shared write authority forbidden
package classification and dependency edges are closed-world/default-deny
```

Required T8-B seam classes carried into T8-C:

```text
provider-neutral transaction participation
same-transaction owner evidence → Audit coordination
owner-authored predicate facts → Authorization final decision
```

T8-B freezes their ownership/direction only. Exact contracts are T8-C.

Other key closure laws:

```text
session/AuthN has an application leaf; omitted use cases never justify transport bypass
generated Go OpenAPI boundary is transport/wire technical, exact content T8-E
IdP protocol client = platform; anti-corruption meaning = Authentication
OfficialRendition mechanism != interactive editor/viewer
idempotency = opaque + live-authorized + erasure-safe
owner persistence adapters private; T8-D owns mapping, not private folder layout
tools/cilint = analyzer host; tools/verify = verification SSOT
stale architecture exception = verifier FAIL, not alert
missing required domain predicate fact = DENY
```

T8-B bootstrap/candidate/reviewer artifacts are no longer active staging. Their live-tree tombstones route to durable authority; full evidence remains in Git history.

## T8-C — ACTIVE

Bootstrap:

`docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-bootstrap.md`

Current non-authoritative candidate:

`docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-global-maximum-candidate.md`

Candidate Global Maximum:

```text
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL

concrete producer-owned semantic-owner APIs
+ narrow consumer-owned mechanism/resolver interfaces only for real consumers
+ application-routed cross-owner facts
+ provider-neutral txscope with application-owned lifecycle
+ owner-authored Audit evidence → application mapping → same-tx Audit append
+ Organization/owner facts → Authorization-owned final ALLOW/default-DENY
+ named transaction-coupled durable-intent ports only
+ operation-local ReplaySnapshot + opaque idempotency mechanism
+ bounded owner facts + application read-projection composition
-
shared contracts/common models
-
generic UnitOfWork/EventBus/policy language/service locator
```

Candidate/reference pass includes current Go/database/sql, PostgreSQL, River, HTTP, OIDC, S3 and OpenAPI behavior only where technically load-bearing. Current repo pins River v0.37.1; its transaction-coupled enqueue property is preserved behind a named port, not exposed as application authority.

### Candidate status

```text
interaction census                   COMPLETE AT CANDIDATE LEVEL
owner/mechanism contract families    MATERIALIZED
inside/outside transaction law       MATERIALIZED
fail-closed law                      MATERIALIZED
T8-A contract reuse disposition      MATERIALIZED
primary/current reference pass       MATERIALIZED
credible alternatives                COMPARED
subtractive/Structural Inversion     COMPLETE AT CANDIDATE LEVEL
independent Fable review             NEXT
```

### Exact next action

```text
independent Fable review of the T8-C Global Maximum candidate
→ reconstruct authority independently
→ challenge selected contract-placement model and T8C-D01→D25
→ test T8-B reopen / T8-D or T8-E trespass
→ search for a materially superior Global Maximum
→ Lead adjudicates reviewer evidence technically
→ bounded correction only where a material finding survives
→ explicit operator ratification before durable T8-C promotion
```

Do **not** start T8-D from the candidate. It remains non-authoritative staging.

### Do not decide by stealth

```text
schema/tables/constraints/locks        → T8-D
exact OpenAPI/wire                     → T8-E
frontend realization                  → T8-F
runtime/process/deployment            → T8-G
transition/deletion                   → T10
implementation tasks                  → T11
```

T8-C may reopen T8-B only on a concrete required-contract contradiction, not preference.

Implementation remains **BLOCKED**.