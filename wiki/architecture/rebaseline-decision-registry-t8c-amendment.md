# Rebaseline Decision Registry — T8-C Closure Amendment

> **Status:** ACTIVE / OPERATOR-RATIFIED REGISTRY RECONCILIATION  
> **Ratified:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **T8-C authority:** `wiki/architecture/r10-t8c-internal-communication-contracts.md`

This bounded amendment reconciles the Decision Registry after T8-C closure. It changes internal realization contracts only. Product Contract and T1→T8-B semantic/topology decisions remain unchanged except for the explicitly ratified T8-C refinement that `internal/application/maintenance` is a non-semantic application leaf inside the already-ratified T8-B `application` class.

Registry authority chain is now:

```text
rebaseline-decision-registry.md
→ rebaseline-decision-registry-d4-amendment.md
→ rebaseline-decision-registry-t6-amendment.md
→ rebaseline-decision-registry-post-t6-amendment.md
→ rebaseline-decision-registry-t7-amendment.md
→ rebaseline-decision-registry-t8a-amendment.md
→ rebaseline-decision-registry-t8b-amendment.md
→ rebaseline-decision-registry-t8c-amendment.md
```

## 1. T8-C stage disposition

```text
T8-C Internal Communication Contracts = CLOSED / OPERATOR-RATIFIED / PROMOTED
```

Ratified Global Maximum:

```text
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL
```

Meaning:

```text
concrete semantic-owner public APIs
+ narrow consumer-owned technical/mid-transition ports only for real consumers
+ stateless application choreography
+ application-routed cross-owner facts
+ database/sql-family shared transaction participation
+ owner-authored same-transaction Audit evidence
+ Authorization as sole final ALLOW/default-DENY authority
+ named transaction-coupled durable intents only
+ self-contained PII-free ReplaySnapshot
+ bounded owner facts + application read composition
-
shared/common semantic contract packages
-
generic UnitOfWork/EventBus/policy language/ServiceLocator
```

## 2. Transaction contract

```text
platform/txscope owns Scope + Runner
application owns begin/commit/rollback lifecycle
semantic owners participate in caller-provided Scope
Scope uses database/sql executor family
semantic-owner public signatures expose no *sql.Tx/pgx.Tx/pool type
T2 READ COMMITTED posture remains binding
```

`txscope.SQLTx(scope) (*sql.Tx,error)` is the sole target native binding for explicitly catalogued platform mechanisms whose external API requires `*sql.Tx`; current named consumer is River.

Fail-closed and enforcement laws:

```text
non-target/foreign Scope -> SQLTx error
application/owners cannot call SQLTx
non-txscope packages cannot embed txscope.Scope
application cannot execute Scope SQL
```

Exact SQL/locks/schema remain T8-D.

## 3. Audit contracts

Mutation evidence:

```text
mutating owner owns intrinsic evidence meaning
application maps/routes only
Audit.AppendIn participates in same Scope
required append failure rolls back transition
```

Read contract:

```text
Authorization.AuthorizedScopes(audit.read)
→ application maps current scope authority
→ Audit.ListEvents applies historical Company/Area attribution before pagination
```

Audit never reconstructs current state or current grants.

## 4. Authorization / owner-fact contracts

Authorization retains sole final decision authority through:

```text
Decide
DecideIn
DecideMany
DecideManyIn
AuthorizedScopes
```

`AuthorizedScopes` is grant/scope prefilter only and never substitutes exact owner-predicate-aware decision.

Organization owns current User/Group facts. Controlled Documents owns document relationship/state/governance predicate meaning. Missing/invalid/unverifiable required predicate remains DENY.

T3 eligibility serialization is represented through protected Organization reads whose semantic guarantee is serialization against concurrent offboarding/disable; exact lock mechanism remains T8-D.

## 5. Controlled Documents / Organization cross-owner facts

Ratified seams include:

```text
EnabledGroupMembersResolver for GROUP-Step activation
responsible-owner eligibility = existing User + same Company + ENABLED
RoleAssignment target identity/eligibility facts
Group deletion dependency facts from Authorization + Controlled Documents
closed ControlledDocs AccessFacts vocabulary
```

Empty GROUP enabled-member snapshot remains truthfully empty. No fallback/System approver/reassign engine is created.

## 6. Version / OCC contract

For T6 whole-replacement mutable resources:

```text
owner read -> current truth + opaque VersionToken
owner mutation -> expected VersionToken
stale -> zero mutation / precondition error
exact already-current repeat -> no version/Audit fabrication + current VersionToken
```

Exact ETag/If-Match wire representation remains T8-E. DRAFT WorkingContent retains its own generation/CAS authority.

## 7. Provider/content mechanism contracts

Authentication consumes verified primitive provider identity coordinates only; raw protocol/provider claims stay inside the IdP protocol mechanism boundary.

Managed-content contract includes:

```text
Allocate
PresignCreate = create-once/no-overwrite
Stat
OpenExact
CopyToNewHandle
DeleteReclaimable
AdmissionClaim Reserve/ProveLive/ConsumeIn/Release
```

No provider key/ETag becomes product identity. No Artifact/Retention semantic owner is created.

Malware CLEAN is valid only when the inspector-returned digest equals the exact admitted-content SHA-256.

## 8. OfficialRendition / River

Launch retains one named transaction-coupled durable-intent family:

```text
official_rendition_render
```

Controlled Documents returns the named intent when the frozen representation policy requires it; application enqueues through a narrow `OfficialRenditionIntentSink` inside the same Scope. River remains mechanism underneath and obtains `*sql.Tx` only through `txscope.SQLTx`.

OfficialRendition exact-content read is a distinct semantic byte resource; current EFFECTIVE reads use effective-read authority and historical reads use history authority.

No generic EventBus/outbox is introduced.

## 9. Idempotency / replay

Internal replay law:

```text
BeginIn / CompleteIn share the semantic Scope
no target FailReplay
same scoped key+fingerprint requests serialize without poisoning Scope
same key+different fingerprint -> conflict / no business mutation
T2 READ COMMITTED posture is inherited
```

Replay representation:

```text
versioned operation-local ReplaySnapshot
PII-free by construction for erasable UserProfile data
self-contained snapshot-only historical reconstruction
no current-state re-projection
no Launch replay purge/redaction subsystem
```

Live disclosure authorization is rechecked; historical mutation/preconditions are not re-executed.

## 10. Read-composition law

```text
Authorization scope prefilter when needed
→ owner canonical filtered/paginated truth
→ owner domain facts
→ exact Authorization Decide/DecideMany where required
→ bounded Organization display refs
→ application lens
→ T8-E wire
```

No foreign SQL, filter-after-pagination workaround, persistent cross-owner read truth or Search owner is created.

## 11. T5-J managed-content GC

GC choreography is hosted by:

```text
internal/application/maintenance
```

This is a non-semantic application leaf inside the existing T8-B `application` class; it creates no T8-B reopen/new owner/new architecture class/new dependency direction.

Two-phase proof is binding:

```text
phase 1: prove semantic references/claims/backup pins absent -> mark GC_PENDING
phase 2 immediately before delete: re-prove GC_PENDING + all semantic/live references + claims + backup protection absent -> DeleteReclaimable
```

Safe failure is leaked storage, never lost governed truth.

## 12. Contract-placement / reuse disposition

```text
owner interfaces by default                    REJECT
shared contracts/common models                REJECT
generic UnitOfWork                            REJECT
generic EventBus/outbox                       REJECT
generic policy/fact language                  REJECT
generic ServiceLocator                        REJECT
current IAM authz.Require exact contract       REWRITE
current objectstore exact contract             REWRITE
current idempotency HTTP middleware contract  REWRITE
current tx lifecycle property                 PRESERVE
narrow database/sql executor property         PRESERVE / REHOME / REFINE
River transactional-insert property           PRESERVE behind named port
OpenAPI/codegen contract-first property       PRESERVE; exact wire T8-E
```

Existence/tests do not create reuse entitlement.

## 13. Ratified decision ledger

```text
D01→D03  authority-aligned hybrid ownership + concrete owner APIs + real consumer-owned ports
D04→D05  database/sql-family txscope + no concrete tx type in owner signatures
D06       owner Audit evidence + same-Scope append + Audit read
D07→D09  canonical Authorization decisions + AuthorizedScopes + Organization subject + closed owner predicates
D10→D11  GROUP resolver + truthful empty snapshot/no fallback
D12→D13  primitive provider identity seam + bounded Organization target/dependency facts
D14→D15  opaque preparation/admission claims + ManagedContent/Malware/DeleteReclaimable
D16→D18  rendition renderer + one named River intent + no EventBus
D19→D21  same-Scope idempotency + PII-free replay + live disclosure recheck
D22→D25  scope-prefiltered read composition + producer errors + external-outside-tx + selective reuse
D26       owner VersionToken/expected-version contract
D27       PII-free replay/no purge subsystem
D28       admission-claim + T5-J GC family
D29       protected eligibility serialization semantics
D30       Audit historical-visibility read
D31       Authorization AuthorizedScopes
D32       D19 inherits T2 READ COMMITTED
D33       Scope sealing defense-in-depth + no embedding + SQLTx fail-closed
D34       application/maintenance hosts T5-J without T8-B reopen
D35       full two-phase immediate pre-delete GC re-proof
D36       self-contained snapshot-only replay reconstruction
D37       free-form replay exclusion by snapshot minimality
D38       database/sql selected as single standard substrate; no current pgx-native consumer
D39       PresignCreate create-once/no-overwrite
D40       bounded synchronous provider-directory callback
D41       proof AuthorizedScopes never grants exact resource action
D42       no-op replacement returns current VersionToken without version/Audit fabrication
```

## 14. Independent review convergence

```text
Round 1: Global Maximum class CONFIRMED
Round 2: BLOCKER 0 / surviving material contradiction 0
Round-1 B5 blocker: NOT SUSTAINED
PII-free replay selection: independently UPHELD
third Fable round: NOT REQUIRED
```

Reviewer output remained evidence until Lead adjudication/operator ratification.

## 15. Stage-boundary reconciliation

```text
T8-C = CLOSED / PROMOTED
T8-D = ACTIVE / Persistence Realization
T8-E = NOT OPEN
T8-F = NOT OPEN
T8-G = NOT OPEN
T8-H = NOT OPEN
T9→T12 = NOT OPEN
implementation = BLOCKED
```

T8-D consumes T1→T8-C and may not alter semantic ownership/contracts by convenience. It owns persistent structure, constraints, SQL/query realization and exact transaction/lock mapping.

T8-E later owns exact HTTP/OpenAPI representation.

No implementation task is authorized by T8-C closure.
