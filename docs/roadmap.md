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
T1 → T8-E                             CLOSED / OPERATOR-RATIFIED / INTEGRATED
T8-F FRONTEND REALIZATION             CLOSED / OPERATOR-RATIFIED / INTEGRATION PENDING
T8-G → T12                            NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T8-F — Frontend Realization is **CLOSED / OPERATOR-RATIFIED** as of 2026-08-21 on `arch/t8f-frontend-realization`. Integration into `main` remains pending through Draft PR #139 and requires a separate explicit operator merge authorization.

Ratified authority:

```text
docs/architecture/frontend.md
+ docs/decisions/frontend-read-symmetry.md   // bounded T8-E-FR precision discovered by T8-F
```

The ratified frontend result remains:

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
T8-G                                 not open
implementation                       blocked
```

### Independent challenge and adjudication

Fable Round 1 Evidence PR #140 challenged candidate `a32ba8b58f5574336f825f46bd552dd96246de7f` and returned F1–F6. Lead adjudication accepted all six and the operator approved the bounded correction package.

The only upstream precision was T8-E-FR:

```text
DocumentOfficialView
  + open_revision?: { revision:RevisionIdentity, state:OpenRevisionState }
  + active_obsolescence_request_id?: Uuid
```

These are disclosure-safe derived routing references, not persisted pointers. They add no Product capability, Permission, lifecycle state, T8-C contract class or application operation. T8-D already guarantees uniqueness of the current DRAFT/SUBMITTED Revision and ACTIVE ObsolescenceRequest per Document.

T8-F additionally closed the Round-1 frontend findings:

```text
navigation presence is not permission-filtering authority
permission.csrf_failed -> session/CSRF re-bootstrap before safe same-command retry
state.upload_expired -> preserve local bytes; fresh allocation/upload/complete; no capability revival
Audit -> inspection/paging only; no inferred filter
History/My Work -> never current-resource identity resolvers
```

Bounded Fable Round 2 Evidence PR #141 reviewed exact corrected candidate `e54986904063c982315129635191ebade8f9b9ed` and returned:

```text
CONVERGED
MATERIAL findings = 0
Round 3 = NOT JUSTIFIED
```

Lead independently revalidated the Round-2 result. Required CI #1047 was SUCCESS on the exact corrected HEAD and `main` remained `6443986672f4f183cff90b76e96e48ebe1c34594` during adjudication.

The operator explicitly ratified T8-F on **2026-08-21**.

## Ratified T8-E baseline

T8-E remains **CLOSED / OPERATOR-RATIFIED / INTEGRATED** with the bounded T8-E-FR member precision above.

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

T8-E-FR changes none of those counts.

## Exact next action

```text
revalidate final ratification-recording HEAD of PR #139
→ required repository CI SUCCESS on that exact HEAD
→ explicit operator merge authorization
→ mark PR #139 ready only when merge is authorized
→ squash merge PR #139 into main
→ revalidate main + merged tree + required CI + durable T8-F/T8-E-FR authorities
→ close/delete absorbed candidate/review branches where tooling permits
→ record T8-F integration closeout
→ only then may T8-G become the next stage; do not open it without explicit operator authorization
```

Do not reopen completed T1→T8-F or the 78-operation Product/T6 census by preference. New material evidence reopens only the authority it actually implicates.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED; T8-E-FR precision ratified with T8-F |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | CLOSED / OPERATOR-RATIFIED; integration pending PR #139 |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/provider boundary, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | NOT OPEN; may open only after T8-F integration/revalidation + explicit operator authorization |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, and runtime realization as one system | Opens after T8-A→T8-G close |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes | Opens after Whole T8 coherence |
| T10 — Transition / Cutover | Real current→target transition and rollback barriers | Opens after T9 baseline |
| T11 — Implementation Program & Execution Graph | Bounded work graph and proof obligations | Opens after T1→T10 accepted |
| T12 — Adversarial Implementation-Readiness | Independent implementation-readiness attack | Opens after T11 |

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

Completed Product/R10 decisions reopen only on material evidence defined by the owning authority or the DevelopmentConexus Engineering Method. Preference, sunk cost, old implementation shape, or hypothetical future capability are not reopen triggers.
