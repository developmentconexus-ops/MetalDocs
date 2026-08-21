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
T8-G RUNTIME / PROCESS / DEPLOYMENT   CLOSED / OPERATOR-RATIFIED / INTEGRATION PENDING
T8-H → T12                            NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T8-G — Runtime / Process / Deployment is **CLOSED / OPERATOR-RATIFIED / INTEGRATION PENDING** as of 2026-08-21.

The stage was opened from freshly revalidated:

```text
main @ 8f39184a2b2e2d07a48ff6796dc9efa77c5c3aac
```

Candidate branch:

```text
arch/t8g-runtime-deployment
```

Ratified durable authority:

```text
docs/architecture/runtime.md
+ docs/decisions/t8g-ratification.md
```

The operator explicitly ratified the converged T8-G design after required CI #1066 passed on independently reviewed candidate `2f6c6f084fa368cceeef5e97b0c846cc381f4ab1` and Fable Round 2 returned **CONVERGED / MATERIAL findings = 0 / Round 3 NOT JUSTIFIED**. The current ratification carrier must pass required Repository Standard CI before integration; integration remains a separate operator gate.

Ratified Global Maximum:

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

Ratified subtraction remains:

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

### T8-G independent review closure

Fable Round 1 Evidence PR #145 returned:

```text
F1  MATERIAL — deployment substrate omitted required writable ephemeral scratch capability
F2  MINOR    — substrate contract omitted per-workload egress-control capability
F3  MINOR    — health probes were described too broadly as public-origin surfaces
F4  MINOR    — renderer/scanner bounds lacked an explicit coherence law with the accepted content profile
verdict      — NOT CONVERGED
```

Lead adjudication accepted F1–F4 as bounded T8-G-only self-coherence corrections. The corrected candidate added no Product capability, semantic owner, application operation or upstream reopen and changed no Global Maximum topology.

Fable Round 2 Evidence PR #146 reviewed exact corrected candidate `2f6c6f084fa368cceeef5e97b0c846cc381f4ab1` and returned:

```text
F1 CLOSED
F2 CLOSED
F3 CLOSED
F4 CLOSED
CONVERGED
MATERIAL findings = 0
Round 3 = NOT JUSTIFIED
```

Both T8-G Evidence PRs are **CLOSED / UNMERGED**. Evidence never becomes authority.

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
revalidate exact ratified T8-G carrier HEAD
→ required Repository Standard CI must pass on that exact Draft HEAD
→ await explicit operator integration authorization
→ only then mark PR #144 ready
→ required post-ready CI must pass on the exact authorized HEAD
→ squash-merge PR #144 only with explicit operator integration authorization
→ revalidate resulting main tree/CI
→ do not open T8-H automatically; T8-H requires its own explicit operator authorization after T8-G integration
```

Candidate/review branch cleanup is non-authoritative housekeeping and does not change stage state.

Do not reopen completed T1→T8-F or the 78-operation Product/T6 census by preference. New material evidence reopens only the authority it actually implicates.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED; T8-E-FR precision ratified with T8-F |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/scanner/provider boundaries, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | CLOSED / OPERATOR-RATIFIED / INTEGRATION PENDING; integration remains a separate operator gate |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, and runtime realization as one system | NOT OPEN; may open only after T8-G integration and explicit operator authorization |
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
