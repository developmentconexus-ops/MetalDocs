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
T1 → T8-E                             CLOSED / OPERATOR-RATIFIED
T8-F FRONTEND REALIZATION             OPEN / ACTIVE
T8-G → T12                            NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T8-F — Frontend Realization is **OPEN / ACTIVE** as of 2026-08-21 on isolated branch `arch/t8f-frontend-realization`, opened from revalidated `main @ 6443986672f4f183cff90b76e96e48ebe1c34594`.

The bounded T8-F authority is `docs/architecture/frontend.md`, derived from accepted Product/frontend journeys plus the ratified T8-E wire. Its derivation order is:

```text
coverage
→ human goals/flows
→ screens/routes
→ vertical traces
→ state/transport/read-model/editor boundaries
→ package topology
→ adversarial subtraction
```

Initial candidate proof remains:

```text
accepted application operations      78 / 78 have frontend consumers
orphaned operations                  0
invented application operations      0
operation 79                         absent
stable SPA route meanings            exact accepted T6 route set
frontend semantic owner added        none
server-state authority               TanStack Query / server truth
parallel global server store         absent
manual parallel DTO/API authority    absent
generic frontend Authorization       absent
interactive DOCX runtime             one adapter boundary
T8-G                                 not open
implementation                       blocked
```

### Fable Round 1 / Lead adjudication

Independent Evidence PR #140 challenged candidate `a32ba8b58f5574336f825f46bd552dd96246de7f` and returned 3 MATERIAL + 3 MINOR findings.

Lead adjudication accepted all six. The three material findings reduced to one upstream class plus one T8-F precision:

```text
F1  stable /documents/:document_id/work lacked a direct current open-Revision identity
F2  shell promised permission-filtered navigation but SessionView has no permission snapshot
F3  Document Official lacked a discoverable active ObsolescenceRequest id
```

The operator approved the correction package explicitly on 2026-08-21.

Bounded T8-E precision is now recorded in `docs/decisions/frontend-read-symmetry.md`:

```text
DocumentOfficialView
  + open_revision?: { revision:RevisionIdentity, state:OpenRevisionState }
  + active_obsolescence_request_id?: Uuid
```

Both members are disclosure-safe derived read truth. They add no persisted Document pointer, no Permission, no Product capability and no application operation. T8-D already guarantees at most one current DRAFT/SUBMITTED Revision and at most one ACTIVE ObsolescenceRequest per Document.

The T8-F candidate also now closes:

```text
navigation presence is not permission-filtering authority
permission.csrf_failed -> session/CSRF re-bootstrap before safe same-command retry
state.upload_expired -> preserve local bytes; allocate/upload/complete again; never revive expired capability
Audit -> inspection/paging only; no inferred filter
History/My Work -> never current-resource identity resolvers
```

Current bounded reopen result:

```text
Product/T6 scope reopen               NO
operation 79                          NO
T8-C contract reopen                  NO
T8-D persistence reopen               NO
T8-E read-model precision             YES / OPERATOR-APPROVED
```

T8-F is **not ratified yet**. A bounded Fable Round 2 must attack only the corrected read symmetry/disclosure, navigation behavior and minor recovery/subtraction corrections before operator ratification.

## Ratified T8-E baseline

T8-E — Executable Wire Contract remains **CLOSED / OPERATOR-RATIFIED / INTEGRATED** except for the bounded operator-approved T8-E-FR precision above.

PR #136 was squash-merged into `main` as `5568788d6322396f230db82e0cd0da027778f55e`; its exact ratified tree passed required CI #1032.

The ratified durable authority remains `docs/architecture/wire-contract.md`; the application census remains 78 operations. The T8-F-discovered member precision is owned by `docs/decisions/frontend-read-symmetry.md` and supersedes only the `DocumentOfficialView` member set.

Ratified T8-E proof remains:

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
required CI                          #1032 SUCCESS on exact merged tree
```

The T8-E-FR precision changes none of those counts.

## Exact next action

```text
revalidate corrected arch/t8f-frontend-realization candidate
→ required repository CI on exact corrected HEAD
→ close Round-1 Evidence PR #140 unmerged after adjudication preservation
→ open isolated bounded Fable Round 2 against exact corrected candidate
→ attack only F1–F6 corrections and regression of 78/78 / route / ownership / T8-G boundaries
→ Lead adjudication
→ if no material finding survives: explicit operator ratification of T8-F before integration
→ do not open T8-G and do not implement Product code
```

Do not reopen completed T1→T8-E or the 78-operation Product/T6 census by preference. New material evidence reopens only the authority it actually implicates.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED; bounded T8-E-FR read-model precision operator-approved |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | OPEN / ACTIVE; Round 2 required before explicit operator ratification |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/provider boundary, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | Opens only after T8-F ratification |
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
