---
id: repository-roadmap
kind: authority
owner: architecture
summary: Sole mutable MetalDocs stage, gate, implementation-status, and next-action authority.
---

# MetalDocs roadmap

## Current state

```text
REPOSITORY MODE                       CLEAN-SLATE / ARCHITECTURE-FIRST
REPOSITORY RESET                      MERGED / OPERATOR-RATIFIED
REPOSITORY STANDARD V1 ALIGNMENT      MERGED
PRODUCT / OWNERSHIP                   OPERATOR-APPROVED
T1 → T8-F                             CLOSED / OPERATOR-RATIFIED / INTEGRATED
T8-G RUNTIME / PROCESS / DEPLOYMENT   OPEN / ACTIVE CANDIDATE / NOT RATIFIED
T8-H → T12                            NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T8-G — Runtime / Process / Deployment is **OPEN / ACTIVE CANDIDATE / NOT RATIFIED** as of 2026-08-21.

The stage was opened by explicit operator authorization from freshly revalidated:

```text
main @ 8f39184a2b2e2d07a48ff6796dc9efa77c5c3aac
```

Candidate branch:

```text
arch/t8g-runtime-deployment
```

Candidate durable authority:

```text
docs/architecture/runtime.md
```

The operator approved the complete design derivation before repository materialization, including the bounded correction that added the T4-required MalwareInspector and the final reuse-first third-party mechanism law. **Design approval is not T8-G ratification.** Ratification remains intentionally gated on repository verification, full independent Fable review, Lead adjudication and explicit operator ratification.

Current candidate Global Maximum:

```text
one modular-monolith application runtime
+ one PostgreSQL product-state database
+ River workers in the application process
+ one active ManagedContentStore
+ one private MalwareInspector
+ one private conditional DOCX→PDF renderer
+ external OIDC provider boundary
+ verified ephemeral exact-byte spool
+ fail-closed recovery profile
+ OpenTelemetry / OTLP observability baseline
+ one-shot migration / job / recovery operations
+ proven third-party mechanisms before local infrastructure
```

Current subtraction remains:

```text
Redis                             absent
separate worker deployment         absent
BFF / SSR                         absent
WebSocket/realtime                absent
external Search                   absent
custom scheduler/event bus         absent
service mesh                      absent
Kubernetes requirement             absent
custom telemetry framework         absent
custom queue/migration/OIDC/S3     absent
operation 79                       absent
Product implementation             BLOCKED
```

### T8-G independent review gate

A **full T8-G Fable review is mandatory before ratification**.

Repository Standard isolation remains:

```text
exact candidate branch/HEAD
→ review/t8g-fable
→ Evidence PR against the exact candidate
→ only docs/work/current/ai-dialog.md differs from candidate
→ Fable reviews whole T8-G authority and its cross-T8 coherence
→ Evidence PR never merges
```

Lead adjudicates every finding against current repository authority. Corrections are applied only to the smallest implicated authority. If material corrections change the candidate, an additional independent Fable round reviews the exact corrected HEAD. T8-G ratification requires convergence with no surviving MATERIAL contradiction plus explicit operator ratification.

## Integrated T8-F baseline

T8-F — Frontend Realization remains **CLOSED / OPERATOR-RATIFIED / INTEGRATED** as of 2026-08-21.

PR #139 was squash-merged into `main` as:

```text
711b8526ebacea9e15034459f85fc37707a32ab4
```

The merge commit carries tree:

```text
ea128d0900667846f33ef188552624886d7885b0
```

That tree is exactly the tree of the final authorized T8-F HEAD `ac3a4b23655e9529e73a6208d41f1804b27855f4`. The same HEAD passed required CI #1055 while Draft and required CI #1056 after the PR was marked ready.

Ratified durable authority:

```text
docs/architecture/frontend.md
+ docs/decisions/frontend-read-symmetry.md
+ docs/decisions/t8f-ratification.md
```

The integrated frontend result remains:

```text
accepted application operations      78 / 78 covered
orphaned operations                  0
invented application operations      0
operation 79                         absent
stable SPA route meanings            exact accepted T6 route set
frontend semantic owner added        none
server-state authority               TanStack Query / server truth
parallel global server store         absent
manual parallel DTO/API authority    absent
frontend Authorization engine        absent
interactive DOCX runtime             one adapter boundary
```

Fable Round 1 Evidence PR #140 returned F1–F6; Lead accepted all six and the operator approved the bounded correction package. Bounded Fable Round 2 Evidence PR #141 reviewed the exact corrected candidate and returned:

```text
CONVERGED
MATERIAL findings = 0
Round 3 = NOT JUSTIFIED
```

Both T8-F Evidence PRs remain closed and unmerged.

## Ratified T8-E baseline

T8-E remains **CLOSED / OPERATOR-RATIFIED / INTEGRATED** with the bounded T8-E-FR member precision ratified with T8-F.

Application census remains exactly 78 operations. Operation 79 remains a material Product/T6 reopen.

T8-E proof remains:

```text
ledger rows                         78 / exact 1..78
operationId                         78 unique
method + path                       78 unique
Idempotency-Key creations           exact 10
ETag read / mutation domains        13 / 13
exact-byte resources                exact 4
Audit operation codes               37 unique
Problem namespace                   https://errors.metaldocs.io/{code}
ShortText / LongText                256 / 4096
attention_required                  absent
PROFILE_REPLACE If-Match+absent     412 precondition.resource_changed
rows 3 / 45 validation.failed       absent
row 42 validation.failed            present
forward obligations                 21 / 3 / 27 = 51
```

T8-G changes none of those counts.

## Exact next action

```text
complete/revalidate exact T8-G candidate HEAD
→ required Repository Standard CI must pass
→ open isolated review/t8g-fable Evidence PR from exact candidate
→ run full independent Fable review of T8-G and cross-T8 coherence
→ Lead adjudicates every finding
→ apply only evidence-backed bounded corrections
→ if candidate changes materially, run another independent Fable round on exact corrected HEAD
→ converge with no surviving MATERIAL contradiction
→ explicit operator T8-G ratification
→ integration authorization / squash merge
→ do not open T8-H before T8-G is ratified and integrated
```

Candidate/review branch cleanup is non-authoritative housekeeping and does not change stage state.

Do not reopen completed T1→T8-F or the 78-operation Product/T6 census by preference. New material evidence reopens only the authority it actually implicates.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED; T8-E-FR precision ratified with T8-F |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/scanner/provider boundaries, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | OPEN / ACTIVE CANDIDATE / NOT RATIFIED; closes only after required CI + converged independent Fable review + explicit operator ratification + integration |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, and runtime realization as one system | NOT OPEN; opens only after T8-G ratification/integration |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes | NOT OPEN; opens after Whole T8 coherence |
| T10 — Transition / Cutover | Real current→target transition and rollback barriers | NOT OPEN; opens after T9 baseline |
| T11 — Implementation Program & Execution Graph | Bounded work graph and proof obligations | NOT OPEN; opens after T1→T10 accepted |
| T12 — Adversarial Implementation-Readiness | Independent implementation-readiness attack | NOT OPEN; opens after T11 |

## Final implementation gate

Implementation remains blocked until all are true:

```text
T8  CLOSED / OPERATOR-RATIFIED
T9  CLOSED / OPERATOR-RATIFIED
T10 CLOSED / OPERATOR-RATIFIED
T11 CLOSED / OPERATOR-RATIFIED
T12 CLOSED / OPERATOR-RATIFIED
Integrated Whole-R10 coherence = PASS
fresh independent challenge = converged
operator implementation authorization = explicit
```

## Reopen law

Completed Product/R10 decisions reopen only on material evidence defined by the owning authority or the DevelopmentConexus Engineering Method. Preference, sunk cost, old implementation shape, hypothetical future capability or generic infrastructure fashion are not reopen triggers.
