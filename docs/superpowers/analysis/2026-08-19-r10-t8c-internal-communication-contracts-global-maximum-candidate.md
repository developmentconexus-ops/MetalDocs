# R10-T8C — Internal Communication Contracts — Global Maximum Candidate

```text
GLOBAL MAXIMUM CANDIDATE
NON-AUTHORITATIVE STAGING
INDEPENDENT-REVIEW INPUT
NOT TARGET AUTHORITY
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Revalidated baseline HEAD:** `6bc5e2299a89e0883d37fe0aac2d1e2905662899`  
> **Stage:** T8-C ACTIVE  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Upstream topology authority:** `wiki/architecture/r10-t8b-backend-module-package-topology.md`  
> **Implementation:** BLOCKED

This artifact is a candidate for independent challenge. It is not durable authority and does not open T8-D.

---

## 0. Authority and evidence reconstruction

Repository authority was revalidated from the live branch before deriving this candidate:

```text
AGENTS.md
docs/engineering/standards/root-cause-global-maximum-method.md
wiki/references/current-agent-handoff.md
wiki/architecture/r10-technical-architecture.md
Product Contract REV001
Whole-Product GCR + launch-v1-ownership-topology.md
T1→T7 durable authorities
T8-A durable authority + registry amendment
T8-B durable authority + registry amendment
r10-post-t6-implementation-readiness-program.md
active T8-C bootstrap
```

Current code is evidence only.

The external-reference pass used primary/current sources where the decision depended on language/database/tool behavior:

```text
Go:
  database/sql transaction documentation
  Go Code Review Comments — Interfaces / package names / Context

PostgreSQL:
  current Transaction Isolation documentation
  current locking documentation

River:
  current River documentation/source for Client.InsertTx
  repository actually pins River v0.37.1

HTTP:
  RFC 9110 for method idempotency / validators / If-Match semantics

OIDC:
  OpenID Connect Core 1.0 — Authorization Code Flow, issuer + subject identity

S3 reference provider:
  AWS S3 conditional-write documentation — If-None-Match for create-if-absent

OpenAPI:
  OAS 3.0.3 specification — wire description remains T8-E
```

Reference law:

```text
MetalDocs semantic authority > external pattern
primary standard/tool docs > remembered syntax
reference product/pattern = falsification evidence, never Product authority
```

No reference was used to import a capability absent the Product Contract.

---

## 1. T8-C question

> **What is the smallest complete set of internal contracts that lets the ratified owners and non-semantic application/mechanism layers realize T1→T8-B semantics without direct owner imports, foreign SQL, duplicate authority, hidden write ownership or unnecessary interface ceremony?**

---

## 2. Global Maximum

```text
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL

SEMANTIC OWNERS
  concrete producer-owned public APIs
  one public package path per owner
  no mandatory owner-side service interface

TECHNICAL DEPENDENCIES
  narrow consumer-owned interfaces only for real mechanism consumers
  primitive / standard-library / opaque technical values across mechanism ports
  provider SDK types never escape mechanism boundary

TRANSACTIONS
  shared platform/txscope contract
  application owns transaction lifecycle
  owners participate explicitly
  narrow database/sql executor shape
  no *sql.Tx / pgx.Tx on semantic-owner public signatures

CROSS-OWNER FACTS
  fact known before mutation -> application gathers and maps
  fact discovered mid-transition -> bounded request-scoped consumer-owned resolver
  no direct owner->owner imports

AUDIT
  mutating owner authors complete intrinsic evidence meaning
  application maps/routes only
  Audit appends inside same scope before commit

AUTHORIZATION
  Organization authors subject facts
  business owner authors relationship/state/governance predicate facts
  Authorization alone computes final ALLOW/default-DENY

DURABLE INTENT
  named intent-specific ports only
  River remains mechanism underneath
  no EventBus / generic outbox API

IDEMPOTENCY
  platform mechanism owns key/fingerprint/claim persistence
  operation-local application ReplaySnapshot owns replayable semantic result
  no HTTP response writer/body contract inside application

READ PROJECTIONS
  owners return bounded canonical facts
  application composes purpose-built lens result
  no foreign SQL / persistent duplicate truth

NO
  shared/contracts
  common/models
  generic ServiceLocator
  generic Repository interface family
  policy expression language
  generic DomainEvent bus
  generic UnitOfWork
```

Method outcome at candidate stage:

```text
CURRENT T8-B STRUCTURE CONFIRMED
+ REWRITE INTERNAL CONTRACTS TO FIT RATIFIED AUTHORITY
```

No T8-B reopen is currently required.

---

## 3. Go contract law

### 3.1 Semantic-owner public APIs are concrete

Each owner exports concrete public types from its one T8-B public package:

```text
authentication.Service
organization.Service
authorization.Service
controlleddocs.Service
audit.Service
```

Do not add `AuthenticationService`, `OrganizationService`, `AuthorizationService`, `ControlledDocsService` or `AuditService` interfaces merely for mocking.

Reason:

```text
semantic owner is the producer/authority
there is exactly one product implementation at Launch
Go concrete types permit additive producer evolution
an implementor-side interface would add ceremony without a second consumer
```

### 3.2 Interfaces are consumer-owned when inversion is real

A narrow interface belongs to the package that materially consumes a replaceable mechanism or mid-transition external fact.

Examples accepted below:

```text
authentication.ProviderClient
controlleddocs.ManagedContent
controlleddocs.MalwareInspector
controlleddocs.EnabledGroupMembersResolver
application/... OfficialRenditionRenderer
application/... OfficialRenditionIntentSink
```

Do not define an interface before a real call site exists.

### 3.3 Context

Every request/job-sensitive public call takes `context.Context` first.

No custom Context type. No semantic authority is stored in context. Context may carry cancellation/deadline/correlation and already-authenticated request identity plumbing, but durable User/Role/Permission state remains owner truth.

### 3.4 Identifier/value posture

Cross-owner identity references use opaque UUID values without importing another owner's named semantic type.

```text
github.com/google/uuid.UUID
```

Owner-specific enums and result types stay owner-local.

No `common/types` package is introduced.

---

## 4. Provider-neutral transaction participation — exact T8-C contract

Home remains:

```text
internal/platform/txscope
```

### 4.1 Contract

Conceptual exact Go surface:

```go
type Scope interface {
    ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
    QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Runner interface {
    Within(ctx context.Context, fn func(Scope) error) error
}
```

`Scope` intentionally mirrors the narrow standard-library executor shape already proven in current code. `*sql.Tx` satisfies it structurally, but `*sql.Tx` is not part of any semantic-owner public signature.

### 4.2 Ownership and lifecycle

```text
application
  owns Runner
  invokes Within
  receives Scope
  passes Scope to owner mutation / in-transaction query calls

owner
  may use Scope only through owner-private persistence
  does not begin/commit/rollback

transport
  never opens Scope

platform/postgres
  realizes begin/commit/rollback in T8-D
```

`Runner.Within` contract:

```text
begin failure -> error; callback not invoked
callback returns error -> rollback; original error remains observable
callback panics -> rollback best-effort; panic propagates
callback nil -> commit
commit failure -> operation not reported successful
Scope does not escape callback lifetime
```

### 4.3 Application SQL prohibition

Application may hold/pass `txscope.Scope` but may not execute semantic SQL through its methods.

Target `cilint` proof must reject `ExecContext`, `QueryContext` or `QueryRowContext` invocation from `internal/application/**` except a deliberately proven technical fixture.

This closes the capability leak created by exposing an executor-shaped Scope while preserving the standard-library transaction primitive for owner-private persistence.

### 4.4 Deferred to T8-D

```text
actual *sql.Tx creation
pool/driver choice inside PostgreSQL realization
isolation options
lock clauses/order
serialization roots
SQL/query mapping
constraint/index design
```

The candidate does not require pgx-native transaction types.

---

## 5. Audit evidence contract

T3 ordering remains:

```text
business/security mutation
-> owning-domain evidence
-> AuditEvent(s)
-> COMMIT
```

### 5.1 Owner-local evidence result

Every mutation for which T3 requires same-local-commit Audit returns zero or more owner-authored `Evidence` values from that owner package.

The owner-local type carries these meanings using owner-local enums/structs and primitive values:

```text
actor attribution:
  USER + User id
  OR SYSTEM + stable product system-actor code

operation code
resource kind
stable resource id
visibility:
  Company id
  OR Company + Area id snapshot
bounded PII-minimized facts
```

The owner, not application, chooses:

```text
operation
resource identity
visibility attribution
fact keys/values
actor kind where SYSTEM action is involved
```

Trusted event id/time are not owner-generated.

### 5.2 Application mapping

Each application leaf performs a mechanical field-for-field conversion from owner-local evidence to `audit.AppendInput`.

Application may add only non-semantic request correlation metadata.

Application must not:

```text
change operation meaning
invent resource kind/id
invent facts
copy free-form reason/content into Audit
change visibility attribution
suppress required owner evidence
```

No shared `auditcontracts` package is introduced.

### 5.3 Audit public append contract

Conceptual exact surface:

```go
type AppendInput struct {
    ActorKind      ActorKind
    ActorUserID    uuid.UUID
    SystemActor    string
    Operation      string
    ResourceKind   string
    ResourceID     uuid.UUID
    CompanyID      uuid.UUID
    AreaID         *uuid.UUID
    Facts          json.RawMessage
    CorrelationID  string
}

type EventRef struct {
    EventID uuid.UUID
}

func (s *Service) AppendIn(
    ctx context.Context,
    scope txscope.Scope,
    in AppendInput,
) (EventRef, error)
```

`Audit` generates stable event identity and trusted event time during append. Exact database-time realization is T8-D.

Multiple required events are appended by repeated `AppendIn` calls in the same Scope. No generic domain-event batch framework is required.

### 5.4 Failure law

```text
required append error -> callback returns error -> entire Scope rolls back
owner mutation cannot be reported successful without all required appends
```

T9 must exercise a negative case for every T3 mandatory Audit mutation family.

---

## 6. Authorization contract

### 6.1 Authority partition

```text
Organization
  User existence / Company / enabled state / current GroupMembership facts

Authorization
  Role semantics
  Permission semantics
  RoleAssignments
  scope composition
  final ALLOW/default-DENY

Controlled Documents
  document relationship/state/governance predicate meaning

application
  map/route only
```

### 6.2 Organization subject snapshot

Organization exports:

```go
type SecuritySubject struct {
    UserID    uuid.UUID
    CompanyID uuid.UUID
    Enabled   bool
    GroupIDs  []uuid.UUID
}

func (s *Service) SecuritySubject(
    ctx context.Context,
    userID uuid.UUID,
) (SecuritySubject, error)

func (s *Service) SecuritySubjectIn(
    ctx context.Context,
    scope txscope.Scope,
    userID uuid.UUID,
) (SecuritySubject, error)
```

`GroupIDs` contains current memberships only. Provider groups never appear.

For governed/security-changing commands, use `SecuritySubjectIn` so actor eligibility/grant evaluation participates in the same local transaction.

### 6.3 Authorization input

Authorization owns its decision input vocabulary:

```go
type TargetScope struct {
    CompanyID uuid.UUID
    AreaID    *uuid.UUID
}

type Subject struct {
    UserID    uuid.UUID
    CompanyID uuid.UUID
    Enabled   bool
    GroupIDs  []uuid.UUID
}

type DomainPredicate struct {
    Provided  bool
    Satisfied bool
}

type Check struct {
    Permission      Permission
    Target          TargetScope
    DomainPredicate DomainPredicate
}

type Decision uint8

const (
    DENY Decision = iota
    ALLOW
)
```

Application converts `organization.SecuritySubject` into `authorization.Subject` mechanically.

Authorization owns the static rule indicating which permissions require domain predicate input. Application cannot weaken requiredness by setting `Provided=false`.

### 6.4 Decision methods

```go
func (s *Service) Decide(
    ctx context.Context,
    subject Subject,
    check Check,
) (Decision, error)

func (s *Service) DecideIn(
    ctx context.Context,
    scope txscope.Scope,
    subject Subject,
    check Check,
) (Decision, error)

func (s *Service) DecideMany(
    ctx context.Context,
    subject Subject,
    checks []Check,
) ([]Decision, error)

func (s *Service) DecideManyIn(
    ctx context.Context,
    scope txscope.Scope,
    subject Subject,
    checks []Check,
) ([]Decision, error)
```

Laws:

```text
subject.Enabled=false -> DENY
Company mismatch -> DENY
no live matching grants -> DENY
permission requiring domain predicate + Provided=false -> DENY
Provided=true + Satisfied=false -> DENY
technical failure loading Authorization-owned current grants -> error; never ALLOW
only canonical static role->permission mapping is evaluated
```

No decision cache is durable authority.

`DecideMany` is a batch evaluation over the same canonical evaluator, not a second ruleset.

### 6.5 Controlled Documents predicate facts

Controlled Documents exports closed, owner-local access-fact queries for the T3 document predicate vocabulary. The contract is not a generic policy language.

Conceptual closed action set:

```text
READ_EFFECTIVE
READ_HISTORY
READ_WORKING
EDIT_WORKING
SUBMIT_OR_WITHDRAW
CANCEL_REVISION
OBSOLETE
OWNER_MANAGE
GOVERNANCE_ACT
```

Conceptual result:

```go
type AccessFacts struct {
    TargetCompanyID uuid.UUID
    TargetAreaID    *uuid.UUID
    PredicateKnown  bool
    PredicateOK     bool
}
```

Controlled Documents exports ordinary and in-Scope single/batch fact queries. Exact resource references are owner-local tagged values limited to Document/Revision/Submission/GovernanceAttempt+Step/obsolescence request identities.

Rules:

```text
predicate not applicable/known for a required request -> PredicateKnown=false
application maps to authorization.DomainPredicate{Provided:false}
Authorization DENYs
technical owner read failure -> error; never synthetic true
```

`allowed_actions` uses these exact same owner facts plus `DecideMany`; no parallel role/action table is permitted.

---

## 7. Mid-transition GROUP snapshot resolver

Controlled Documents decides when a Governance Step activates. Organization owns who is currently an enabled Group member.

Direct import is forbidden, so Controlled Documents defines the consumer-owned request-scoped resolver:

```go
type EnabledGroupMembersResolver interface {
    EnabledGroupMembersIn(
        ctx context.Context,
        scope txscope.Scope,
        groupID uuid.UUID,
    ) ([]uuid.UUID, error)
}
```

Application supplies an adapter for the current invocation that delegates to:

```go
organization.Service.EnabledGroupMembersIn(...)
```

The resolver is passed only to Controlled Documents mutations capable of activating a GROUP Step:

```text
Submission creation activating first Step
Governance ACCEPT activating next Step
human-governed obsolescence initiation activating first Step
obsolescence ACCEPT activating next Step where applicable
```

Laws:

```text
Controlled Docs owns WHEN resolution occurs
Organization owns WHO the enabled members are
application owns neither meaning
snapshot is copied into Controlled Documents immutable/in-flight governance truth
later GroupMembership drift does not rewrite the activated snapshot
current Authorization is still rechecked when a candidate acts
```

If current enabled membership resolves to an empty set, the snapshot is truthfully empty. No fallback selector, System approver or implicit reassign is invented. The frozen route is then impossible to progress and the already-ratified bounded recovery is withdraw/fix/resubmit (or the dedicated obsolescence withdrawal path).

No generic service locator may grow from this resolver pattern.

---

## 8. Authentication public contracts

Authentication remains the owner of ProviderSubjectBinding, ApplicationSession and IdP anti-corruption meaning.

### 8.1 Provider protocol port

Authentication defines the consumer-owned port. It deliberately uses standard/primitive protocol data so the platform implementation need not import Authentication types:

```go
type ProviderClient interface {
    AuthorizationURL(
        ctx context.Context,
        state string,
        codeChallenge string,
        redirectURI string,
    ) (string, error)

    ExchangeAuthorizationCode(
        ctx context.Context,
        code string,
        codeVerifier string,
        redirectURI string,
    ) (issuer string, raw json.RawMessage, err error)

    SearchSubjects(
        ctx context.Context,
        query string,
    ) (issuer string, raw json.RawMessage, err error)
}
```

`platform/identityprovider` may satisfy this interface structurally. It owns protocol mechanics only.

Authentication parses/interprets provider response and owns:

```text
issuer + subject identity meaning
ProviderSubjectBinding
bounded provider-subject reference returned to product journeys
session issuance/revocation
assurance/fresh-auth meaning if later activated
```

Raw provider roles/groups/permissions never enter Organization/Authorization.

### 8.2 Browser/session capability census

Authentication concrete public API must cover exactly these Launch families:

```text
BeginBrowserLogin
VerifyBrowserCallbackPreflight
ResolveSession
SearchProviderSubjects
ProviderBinding
ResolveProviderBindingIn
BindProviderSubjectIn
ReplaceProviderBindingIn
IssueSessionIn
RevokeSessionIn
RevokeAllSessionsForUserIn
```

Provider/network exchange occurs before local semantic transaction.

`VerifyBrowserCallbackPreflight` returns an Authentication-owned verified provider-subject value. Application may carry it into the later transaction but does not interpret raw claims.

Session issuance flow:

```text
provider verification outside tx
-> application opens Scope
-> Authentication.ResolveProviderBindingIn
-> Organization current enabled User fact in same Scope
-> Authentication.IssueSessionIn using the resolved eligibility fact
-> commit
```

Unknown binding or disabled User fails closed.

Login/logout ordinary telemetry is not mandatory T3 semantic Audit.

### 8.3 User-eligibility fact consumed by Authentication

When issuing a session, application maps Organization current truth into an Authentication-owned input containing at least:

```text
resolved flag
User id
Company id
enabled flag
```

Authentication rejects unresolved, mismatched or disabled input.

---

## 9. Organization public contracts

Organization concrete API owns these public query/capability families required by T6/T8-C:

### 9.1 Queries

```text
Company
Users / User
UserProfile
User eligibility
Areas / Area
Groups / Group
Group members
SecuritySubject / SecuritySubjectIn
EnabledGroupMembersIn
UserReferences (batched bounded selection/display facts)
ResponsibleOwnerEligibilityIn
RoleAssignmentTargetFactsIn
```

No query returns provider roles/groups or Authorization grants as Organization truth.

### 9.2 Mutations

All semantic/security mutations execute inside caller-provided Scope:

```text
CreateUserIn
UpdateUserProfileIn
EraseUserProfileIn
SetUserEligibilityIn
UpdateCompanyIn
CreateAreaIn
UpdateAreaIn
SetAreaLifecycleIn
CreateGroupIn
UpdateGroupIn
DeleteGroupIn
PutGroupMemberIn
RemoveGroupMemberIn
```

Required T3 Audit mutations return Organization-authored evidence.

### 9.3 Responsible-owner eligibility

`ResponsibleOwnerEligibilityIn` returns exactly the D4 fact:

```text
resolved
existing User
same Company
enabled
```

It does not compute document permission or grant access.

Application maps this fact into the Controlled Documents owner-change/create command. Controlled Documents fails closed if the fact is unresolved or false.

### 9.4 RoleAssignment target facts

Authorization owns the grant, but Organization owns whether User/Group/Area identities exist in the Company.

`RoleAssignmentTargetFactsIn` returns only bounded identity/scope existence facts required to validate the grant target. Authorization must reject unresolved or invalid target facts.

### 9.5 Group deletion foreign dependencies

Before `DeleteGroupIn`, application gathers in the same Scope:

```text
Authorization: live Group RoleAssignment dependency fact
Controlled Documents: current GovernanceRoute / unresolved live route dependency fact
```

Application maps them into an Organization-owned `GroupDeletionDependencies` value with explicit `Resolved` markers for both foreign sources.

Organization rejects deletion when:

```text
foreign dependency source unresolved
OR live GroupMembership exists
OR live Group RoleAssignment exists
OR current GovernanceRoute references Group
OR live unactivated frozen GROUP Step still needs Group resolution
```

Application does not decide deletion; it supplies owner-authored dependency facts.

---

## 10. Authorization mutation/query contracts beyond Decide

Authorization concrete API covers:

### Queries

```text
Roles                         static product vocabulary
RoleAssignments              current grants
GroupRoleAssignmentUsageIn   bounded dependency fact for Group deletion
```

### Mutations

```text
CreateRoleAssignmentIn
RevokeRoleAssignmentIn
RevokeDirectUserAssignmentsIn
```

RoleAssignment creation consumes Organization-owned target facts mapped by application. Unresolved target identity/scope => no grant.

Offboarding uses `RevokeDirectUserAssignmentsIn` and does not remove Group RoleAssignments.

Every grant/revoke returns Authorization-authored required Audit evidence.

No owner query exposes an expanded materialized permission cache as authority.

---

## 11. Controlled Documents public contract census

Controlled Documents remains one public owner surface. The concrete Service may use owner-private responsibility packages without exposing them.

### 11.1 Canonical read/query families

The public API must cover:

```text
DocumentTypes / DocumentType
DocumentTypeGovernance
EligibleTemplates
NumberingPreview
TemplateConfigurationItems
DocumentCreationBaseOptions
LibraryCandidates
DocumentOfficial
DocumentWork
DocumentHistory
AuthoringWork
GovernanceWork
Submission
GovernanceCase
GovernanceFeedback
GovernanceStepDecision
Release
ObsolescenceRequest
AccessFacts / AccessFactsIn / batched variants
GroupGovernanceUsageIn
RenditionWorkCandidate
```

These are semantic facts/read projections owned by Controlled Documents. They do not include foreign Organization display/PII data by convenience.

### 11.2 Mutation families

Semantic mutations run inside caller-provided Scope:

```text
CreateDocumentIn
CreateNextRevisionIn
UpdateDraftIn
SubmitIn
WithdrawSubmissionIn
CancelRevisionIn
AddGovernanceFeedbackIn
DecideGovernanceStepIn
ChangeResponsibleOwnerIn
SetTemplateRoleIn
CreateDocumentTypeIn
UpdateDocumentTypeIn
ReplaceDocumentGovernanceIn
ReplaceEligibleTemplatesIn
InitiateObsolescenceIn
WithdrawObsolescenceIn
CompleteOfficialRenditionIn
```

Release is not a public application command. It occurs inside the appropriate Controlled Documents mutation when all gates are satisfied.

Required Audit evidence is returned with mutation result.

### 11.3 Mutation-result law

A mutation result contains only:

```text
canonical semantic result needed by application
zero-or-more owner-authored Audit evidence records
zero-or-one named durable intent when that semantic transition requires activated future work
```

It does not return generic `DomainEvent[]`.

### 11.4 External/preflight content preparation

Provider/scanner work never participates in the semantic transaction. Controlled Documents exposes preflight capabilities that use its technical ports and return owner-created opaque proof/candidate values that application can carry but cannot forge meaningfully:

```text
PrepareTemplateSeed
PrepareNextRevisionSeed
StartDraftUpload
CompleteDraftUpload
PrepareSubmissionGovernedContent
OpenSemanticSource
PrepareOfficialRenditionCandidate
```

Opaque preparation values use exported types with unexported internal fields or equivalent construction control. Only Controlled Documents creates/validates them.

Every later `...In` mutation revalidates the canonical semantic state and the exact prepared handle/descriptor/proof before commit.

Stale preparation creates no semantic fact; reclaimable bytes remain mechanism state.

---

## 12. Managed-content contract

Controlled Documents is the semantic consumer; managed content remains a mechanism.

### 12.1 Consumer-owned ManagedContent port

Conceptual exact minimal interface:

```go
type ManagedContent interface {
    Allocate(ctx context.Context) (uuid.UUID, error)

    PresignCreate(
        ctx context.Context,
        handle uuid.UUID,
        maxBytes int64,
        ttl time.Duration,
    ) (string, error)

    Stat(
        ctx context.Context,
        handle uuid.UUID,
    ) (exists bool, sizeBytes int64, err error)

    OpenExact(
        ctx context.Context,
        handle uuid.UUID,
    ) (io.ReadCloser, error)

    Copy(
        ctx context.Context,
        source uuid.UUID,
        destination uuid.UUID,
    ) error
}
```

No provider bucket/key/version/ETag appears.

Range read is deliberately not part of the required Launch interface because T4/T6 keep it optional. If a real viewer/file-size consumer activates range reads, add a second narrow consumer interface rather than widening every implementation now.

### 12.2 Create-once property

`PresignCreate` means create-once/no-overwrite semantics, not merely "return a PUT URL".

The S3 reference implementation must use a provider mechanism capable of enforcing create-if-absent (for example the provider's conditional-write primitive) and must prove the property. Exact provider realization remains implementation/T8-D/T8-G owned.

### 12.3 Descriptor derivation

The mechanism does not accept a client-authoritative semantic hash.

Controlled Documents admission reads exact bytes through `OpenExact` and derives/verifies:

```text
SHA-256
size
actual closed ContentFormat
```

Provider ETag/checksum remains evidence only.

---

## 13. Malware-inspection port

Controlled Documents owns the governed-boundary admission rule and defines the narrow consumer port:

```go
type MalwareInspector interface {
    Inspect(
        ctx context.Context,
        content io.Reader,
    ) (clean bool, err error)
}
```

Semantics:

```text
clean=true,nil   -> proof may be accepted for exact bytes
clean=false,nil  -> malicious; immutable governed admission forbidden
err!=nil         -> unavailable/incomplete; fail closed and retry visibly
```

The port owns no Document/Submission meaning and no scan lifecycle state.

---

## 14. OfficialRendition execution contracts

The renderer is a technical mechanism distinct from Controlled Documents OfficialRendition truth.

### 14.1 Renderer consumer port

The application leaf responsible for OfficialRendition work owns:

```go
type OfficialRenditionRenderer interface {
    RenderPDF(
        ctx context.Context,
        source io.Reader,
    ) (io.ReadCloser, error)
}
```

No renderer job/provider id becomes semantic identity.

### 14.2 Worker flow

```text
job adapter -> application leaf
-> ControlledDocs.RenditionWorkCandidate
-> ControlledDocs.OpenSemanticSource
-> renderer outside semantic tx
-> ControlledDocs.PrepareOfficialRenditionCandidate
-> application opens Scope
-> ControlledDocs.CompleteOfficialRenditionIn
-> required Audit AppendIn
-> possible system Release inside ControlledDocs mutation
-> commit
```

If canonical eligibility is gone at finalization, semantic result is no-op and produced bytes become reclaimable mechanism state.

---

## 15. Transaction-coupled durable-intent contract

Launch has one currently activated semantic durable-intent family:

```text
official_rendition_render(submission_id, required_format)
```

No generic EventBus/Outbox contract is created.

### 15.1 Submit result

When frozen representation policy requires OfficialRendition, `ControlledDocs.SubmitIn` returns an owner-authored intent description containing only:

```text
Submission id
required format
```

SourceOnly returns no rendition intent.

### 15.2 Consumer-owned sink

The submitting application leaf owns:

```go
type OfficialRenditionIntentSink interface {
    EnqueueOfficialRenditionIn(
        ctx context.Context,
        scope txscope.Scope,
        submissionID uuid.UUID,
        requiredFormat string,
    ) error
}
```

River implementation remains behind this interface.

### 15.3 Atomicity

```text
ControlledDocs semantic Submission requiring rendition commits
<=>
required durable intent row commits in same Scope
```

Intent insertion error rolls back the semantic transition.

River's `InsertTx` transaction-safe property is the preserved mechanism evidence; River client/driver/transaction types never enter the application contract.

No materialized Search intent exists because T6 keeps Search materialization OFF.

No routine IdP-disable intent exists because T5 keeps it OFF.

---

## 16. Idempotency internal contract

T6 owns which operations require durable `Idempotency-Key`. T8-C owns the internal application/mechanism contract.

### 16.1 No target HTTP middleware ownership

The current replay middleware hashes raw method/path/body and captures `http.ResponseWriter`. That contract is not inherited.

Target law:

```text
wire validates Idempotency-Key syntax in T8-E
application owns canonical operation id + semantic command fingerprint
platform/idempotency owns only scoped replay coordination/storage
```

### 16.2 Platform mechanism surface

Application may import `platform/idempotency` per T8-B.

Conceptual exact contract:

```go
type Claim struct { /* opaque */ }

type Replay struct {
    Payload []byte
}

type Store struct { /* T8-D realization */ }

func (s *Store) BeginIn(
    ctx context.Context,
    scope txscope.Scope,
    actorID uuid.UUID,
    operationID string,
    key string,
    fingerprint []byte,
) (claim *Claim, replay *Replay, err error)

func (s *Store) CompleteIn(
    ctx context.Context,
    scope txscope.Scope,
    claim *Claim,
    payload []byte,
) error
```

Exactly one of `claim` or `replay` is non-nil on successful `BeginIn`.

Semantics:

```text
new key -> claim
existing completed same fingerprint -> replay
existing same key different fingerprint -> conflict error
transaction rollback -> no completed new replay record
no public/durable IN_PROGRESS business state
```

No `FailReplay` operation is required in the target contract because claim acquisition/completion occurs inside the same local transaction; rollback releases the uncommitted claim structurally.

### 16.3 Semantic fingerprint

Fingerprint is derived deterministically by the application use case from validated semantic command fields plus canonical operation identity.

It is not a hash of raw HTTP bytes/path/query by architecture requirement.

The idempotency mechanism treats fingerprint bytes as opaque equality material.

### 16.4 ReplaySnapshot

For each T6 operation requiring durable replay, the owning application leaf defines an operation-local, versioned `ReplaySnapshot` containing the exact semantic result needed for later deterministic wire replay.

Rules:

```text
snapshot is application-owned, not platform-owned
snapshot is not current-state authority
snapshot is not raw HTTP response bytes
snapshot contains no provider identity
snapshot contains no more PII than the original replayable result requires
snapshot version is explicit across deployments within retention window
platform stores encoded snapshot bytes opaquely
```

T8-E must define a deterministic mapping:

```text
ReplaySnapshot version
-> exact original status/body representation
```

T8-E may not force application to import generated wire DTOs.

### 16.5 Replay authorization

On replay hit:

```text
DO NOT re-run historical business mutation
DO NOT re-run original lifecycle preconditions
DO authenticate current request
DO re-check current permission/scope
DO re-check minimum current resource visibility needed to disclose snapshot
only then return encoded replay
```

Each application operation reuses the same current Authorization/domain-predicate path it would use to disclose the corresponding result; no generic replay ACL is created.

### 16.6 Success ordering

For a new idempotent creation:

```text
BEGIN Scope
BeginIn
current AuthZ / domain revalidation
semantic mutation(s)
required Audit append(s)
required durable intent(s)
CompleteIn with ReplaySnapshot
COMMIT
```

Any failure before commit rolls back semantic fact + Audit + durable intent + replay result together.

T9 must prove every idempotent creation family cannot commit a semantic result without its completed replay snapshot.

---

## 17. Cross-owner application flows

### 17.1 Request authentication

Transport may invoke the `application/session` leaf to resolve the cookie/session into an authenticated principal before invoking the requested application operation.

```text
transport
-> application/session
-> Authentication.ResolveSession
-> authenticated User id
```

No transport->Authentication direct import/call exists.

Authorization for a resource operation remains in the target application leaf and uses current Organization/AuthZ/domain facts.

### 17.2 User creation

```text
provider subject search/verification outside tx
-> application opens Scope
-> current actor SecuritySubjectIn
-> Authorization.DecideIn(organization.manage)
-> Organization.CreateUserIn
-> Authentication.BindProviderSubjectIn
-> owner-authored Audit evidence -> Audit.AppendIn
-> idempotency CompleteIn when applicable
-> commit
```

No half-created User/Profile/Binding can commit.

### 17.3 Offboarding

```text
application opens Scope
-> actor SecuritySubjectIn
-> Authorization.DecideIn(organization.manage)
-> serialize target eligibility through owner calls
-> Organization disable User
-> Authentication revoke all ApplicationSessions
-> Organization remove current GroupMemberships
-> Authorization revoke direct User RoleAssignments
-> append every required owner-authored Audit evidence
-> commit
```

Group RoleAssignments remain.
ProviderSubjectBinding remains.
Re-enable does not resurrect grants/memberships/sessions.

### 17.4 Group membership mutation

```text
Scope
-> actor SecuritySubjectIn
-> Authorization.DecideIn(access.manage @ Company)
-> Organization.Put/RemoveGroupMemberIn
-> Audit.AppendIn
-> commit
```

Membership remains Organization truth; `access.manage` protects mutation without transferring ownership.

### 17.5 RoleAssignment mutation

```text
Scope
-> actor SecuritySubjectIn
-> Authorization.DecideIn(access.manage @ Company)
-> Organization.RoleAssignmentTargetFactsIn
-> Authorization.Create/RevokeRoleAssignmentIn
-> Audit.AppendIn
-> commit
```

### 17.6 Group deletion

```text
Scope
-> actor SecuritySubjectIn
-> Authorization.DecideIn(organization.manage)
-> Authorization.GroupRoleAssignmentUsageIn
-> ControlledDocs.GroupGovernanceUsageIn
-> Organization.DeleteGroupIn with resolved dependency facts
-> Audit.AppendIn
-> commit
```

### 17.7 Document create

Provider/template copy preparation occurs outside tx when required.

```text
Scope
-> actor SecuritySubjectIn
-> ControlledDocs create/domain access facts
-> Authorization.DecideIn(document.create)
-> if deliberate other owner:
     Organization.ResponsibleOwnerEligibilityIn
     Authorization.DecideIn(document.owner.manage)
-> ControlledDocs.CreateDocumentIn
-> Audit.AppendIn
-> idempotency CompleteIn
-> commit
```

### 17.8 Next Revision

```text
exact current-effective source copy outside tx
-> Scope
-> actor subject + ControlledDocs edit predicate facts
-> Authorization.DecideIn(document.edit)
-> ControlledDocs.CreateNextRevisionIn revalidates source/current state
-> Audit.AppendIn
-> idempotency CompleteIn
-> commit
```

### 17.9 DRAFT mutation

Upload/admission preparation occurs outside semantic tx.

```text
Scope
-> actor subject + ControlledDocs working/edit facts
-> Authorization.DecideIn(document.edit)
-> ControlledDocs.UpdateDraftIn with expected generation + prepared upload when any
-> commit
```

WorkingContent autosave is not mandatory semantic Audit.

### 17.10 SUBMIT

Malware/exact-content preflight occurs outside tx.

```text
Scope
-> actor subject + ControlledDocs submit predicate facts
-> Authorization.DecideIn(document.submit)
-> ControlledDocs.SubmitIn
   may freeze Submission
   may create GovernanceAttempt + GROUP snapshot through resolver
   may system-Release if gates are synchronous
   may return OfficialRendition durable intent
-> enqueue named intent in same Scope when returned
-> append required evidence(s)
-> idempotency CompleteIn
-> commit
```

### 17.11 Governance decision

```text
Scope
-> actor SecuritySubjectIn
-> ControlledDocs governance-action facts
-> Authorization.DecideIn(governance.act)
-> ControlledDocs.DecideGovernanceStepIn
   records immutable Decision
   may activate next GROUP Step via resolver
   may system-Release if final gate now satisfied
-> required Audit evidence(s)
-> commit
```

### 17.12 Withdraw / cancel / obsolescence

Each follows the same pattern:

```text
Scope
-> current actor subject
-> owner-authored exact lifecycle/relationship facts
-> Authorization.DecideIn required T3 permission
-> ControlledDocs mutation
-> required Audit evidence
-> idempotency snapshot only for T6 POST creations
-> commit
```

### 17.13 Responsible-owner replacement

```text
Scope
-> actor subject
-> ControlledDocs owner-manage facts
-> Authorization.DecideIn(document.owner.manage)
-> Organization.ResponsibleOwnerEligibilityIn(target)
-> ControlledDocs.ChangeResponsibleOwnerIn with resolved eligibility fact
-> Audit.AppendIn
-> commit
```

If target offboarding linearizes first, eligibility is disabled and replacement fails closed.

---

## 18. Read projections and query composition

Purpose-built T6 views remain application lens results, never owner authority.

### 18.1 General law

```text
owner canonical facts
-> application mapping/composition
-> current Authorization decision
-> bounded Organization display references where needed
-> application read result
-> T8-E wire encoding
```

No application SQL.
No foreign owner SQL.
No persistent cross-owner projection merely for convenience.

### 18.2 Batched reference/decision support

To avoid N+1 orchestration without creating a shared read database:

```text
Organization.UserReferences supports batched User ids
Authorization.DecideMany supports batched checks
ControlledDocs access-fact queries support bounded batch input where a list lens needs it
```

Batch methods are the same owner authority evaluated in bulk, not a second cache/ruleset.

### 18.3 Lens source map

```text
SessionView
  Authentication session truth
  + bounded Organization User reference when displayed

Library / DocumentSummary
  ControlledDocs current official candidates
  + Authorization DecideMany
  + Organization User/Area references only for display

My Work / authoring
  ControlledDocs actor-related work facts
  + current Authorization checks

My Work / governance
  ControlledDocs live governance participation facts
  + current Authorization checks

Document Official
  ControlledDocs stable Document + current EFFECTIVE truth
  + current Authorization

Document Work
  ControlledDocs current open Revision/WorkingContent truth
  + current Authorization

Governance Case
  ControlledDocs exact attempt/step/immutable subject
  + current Authorization

Document History
  ControlledDocs immutable/history facts
  + current Authorization

AuditEventView
  Audit facts
  + current Authorization audit.read scope

Document Creation Options
  ControlledDocs active DocumentType/template facts
  + Organization Areas/UserReference facts
  + current Authorization

Administration
  each owner supplies its own bounded current facts
  application co-locates UX only
```

Search remains the canonical PostgreSQL query/view baseline under Controlled Documents/Library read orchestration; no Search owner and no materialized projection is activated.

---

## 19. Inside/outside transaction classification

### Outside local semantic transaction

```text
OIDC authorization/token/provider-directory network calls
provider subject search
managed-content upload
managed-content exact-byte read/copy preparation
malware inspection
OfficialRendition renderer execution
provider physical delete
ordinary non-linearizable read projections unless a command requires in-Scope truth
```

### Inside caller-provided Scope

```text
all semantic/security mutations
command-time Organization subject/eligibility facts
command-time Authorization decisions
command-time ControlledDocs relationship/state/governance facts
Group Step enabled-member resolution at activation
required Audit append
required River durable intent insert
idempotency acquire/complete for T6 idempotent POST creation
```

External/provider call required for success inside Scope is forbidden.

---

## 20. Failure / fail-closed semantics

### 20.1 Owner errors

Errors remain producer-owned. Do not create `common/errors`.

Owners expose typed/sentinel errors sufficient for application to distinguish material categories such as:

```text
not found
invalid/stale state
precondition conflict
eligibility failure
duplicate/singleton conflict
invalid semantic input
```

T8-E maps application outcomes to exact RFC 9457 problem codes; owner packages do not import generated HTTP errors.

### 20.2 Authorization

```text
known insufficient authority -> DENY
missing required domain fact -> DENY
known disabled subject -> DENY
technical failure proving current authority -> error / no ALLOW
```

### 20.3 External mechanisms

```text
IdP unavailable -> no login/provider preflight success
malware inspector unavailable -> no governed admission
managed content missing/corrupt -> no semantic attachment/retrieval success
renderer failure -> no silent SourceOnly downgrade
River enqueue failure -> semantic transaction rollback when intent required
```

### 20.4 No fallback authority

Never convert mechanism failure into:

```text
fake System approval
implicit ALLOW
silent Release
provider claim as MetalDocs role
stale replay disclosure
alternate content object
```

---

## 21. Current-code selective reuse disposition

T8-A five-part gate applied at contract level:

| Current unit/property | T8-C disposition | Reason |
|---|---|---|
| `internal/platform/db.Tx` narrow Exec/Query shape | **PRESERVE PROPERTY / REHOME+REFINE** | standard-library-shaped, current R10 consumer, no legacy semantic authority required |
| current `TxRunner` lifecycle property | **PRESERVE PROPERTY** | one begin/commit/rollback owner remains required |
| current `TxRunner` exact contract / `*sql.Tx` callback / tenant-GUC behavior | **REWRITE** | leaks concrete tx and legacy tenancy/RLS assumptions |
| current IAM `authz.Require` | **REWRITE** | old capability/system_admin/GUC/RLS authority contradicts T3/T8-B target |
| current Audit application/export service | **REWRITE PUBLIC CONTRACT** | Launch Audit owner survives; legacy export/non-Launch shape does not |
| current objectstore public contract | **REWRITE** | provider keys/tenant prefix/caller expected hash do not match T4 target contract |
| current idempotency HTTP middleware | **REWRITE CONTRACT** | raw HTTP hashing/response capture violates application-first target layering |
| current idempotency DB/concurrency implementation ideas | **EVIDENCE ONLY / selective reuse later** | may inform T8-D if target semantics/proof remain exact |
| River v0.37.1 transactional insertion property | **PRESERVE / REFINE BEHIND PORT** | T5 named current consumer + proven transaction-coupled enqueue |
| River concrete Client/driver/transaction types | **DO NOT EXPOSE** | mechanism identity must not become application contract |
| current OpenAPI/codegen mechanism | **PRESERVE PROPERTY; T8-E owns exact** | T6/T8-A already ratified |

No current interface survives merely because tests exist.

---

## 22. External reference check — decision impact

### Go standard/library guidance

Observed reference properties:

```text
interfaces generally belong with consumers
avoid interfaces invented only for mocking
avoid defining interfaces before real use
avoid generic package names such as util/common/types/interfaces
context.Context passed explicitly
standard database/sql transaction API provides begin/use Tx/commit/rollback model
```

Decision impact:

```text
concrete semantic owner Services
consumer-owned narrow mechanism ports
no shared/contracts package
reuse standard SQL executor shape instead of inventing Row/Rows abstractions
```

### PostgreSQL

Observed current reference properties:

```text
READ COMMITTED is default
successive reads may see newer commits
explicit row/table locks exist when stronger serialization is required
```

Decision impact:

```text
T8-C carries one transaction Scope
T8-D still owns exact lock/isolation/serialization realization
no global SERIALIZABLE assumption is smuggled into contracts
```

### River

Observed reference property:

```text
InsertTx inserts on caller transaction
job remains invisible until commit
rollback removes job with transaction
```

Decision impact:

```text
preserve transaction-coupled job property
hide River behind named OfficialRenditionIntentSink
```

### HTTP / OpenAPI

Observed reference properties:

```text
HTTP already defines natural idempotent method semantics and validators
OpenAPI is language-agnostic wire description
```

Decision impact:

```text
T6 natural idempotency remains preferred
T8-C replay mechanism is only for T6 named non-idempotent POST creation
wire DTOs/status/body remain T8-E and never enter owner APIs
```

### OIDC

Observed reference property:

```text
Authorization Code Flow returns code to client, tokens at Token Endpoint
issuer + subject are protocol identity coordinates
```

Decision impact:

```text
use provider protocol rather than invent authentication protocol
Authentication translates protocol identity into ProviderSubjectBinding
```

### S3 reference implementation

Observed reference property:

```text
ordinary same-key write can overwrite
conditional create-if-absent is available through provider conditional write semantics
```

Decision impact:

```text
ManagedContent contract requires create-once property
provider implementation must prove it using provider primitive; owner contract does not expose provider header/key
```

### Reference-product boundary

Google Docs / Office / SharePoint-like UX is intentionally not load-bearing to T8-C internal Go contracts. Those products become more relevant to T8-F editor/version/conflict/review UX. Their private backend architecture is not inferred.

---

## 23. Credible alternatives challenged

### Alternative A — producer-owned interfaces/DTOs everywhere

Rejected.

Failure classes:

```text
implementor-side interface ceremony
mechanism types creep into owner contract
mock-driven abstraction
```

### Alternative B — consumer-owned interfaces for every owner call

Rejected.

Failure classes:

```text
application leaf interface explosion
same owner semantics redescribed by multiple consumers
mock/type vocabulary becomes accidental authority
```

### Alternative C — shared `internal/contracts` / common DTO package

Rejected.

Failure classes:

```text
hidden sixth semantic home
ownership ambiguity
circular semantic pressure moves into shared types
future junk-drawer growth
```

### Alternative D — raw domain types cross owners

Rejected.

Would recreate direct owner coupling through types even if import calls were abstracted.

### Alternative E — generic EventBus/outbox

Rejected.

Launch has one named transaction-coupled durable effect family. Generic event infrastructure has no current consumer.

### Alternative F — HTTP response replay inside application

Rejected.

Would reverse T8-B transport/application direction and make T8-E wire authority leak inward.

### Selected

```text
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL
```

No materially stronger alternative has yet survived with less total complexity.

---

## 24. Structural Inversion / subtractive pass

### Structural Inversion

If current implementation instead used:

```text
pgx-native transactions instead of database/sql
no current idempotency middleware
another object provider
another IdP client
no current IAM module
```

these conclusions would remain:

```text
application owns one local transaction lifecycle
owner public APIs stay concrete
no owner->owner import
owner facts route through application
Authorization is sole ALLOW/DENY
Audit is same-commit evidence
required durable intent is transaction-coupled
mechanisms stay behind narrow ports
```

Therefore the candidate is not legacy-shape inheritance.

### Subtractive results

Deleted from the candidate:

```text
owner service interfaces
shared contract package
generic unit-of-work type
custom Row/Rows abstraction
FailReplay operation
generic event bus/outbox
generic policy/fact language
generic mechanism registry/service locator
range-read requirement
Search refresh contract
provider-disable durable intent
```

Every retained interface has a named current consumer.

---

## 25. Mechanical enforcement / proof

### T8-C architecture rules to add to target proof catalog

```text
owner public packages do not import another owner
owner public signatures contain no another-owner named type
semantic owner public signatures contain no *sql.Tx / pgx.Tx / DB pool type
provider SDK types remain inside platform mechanisms
application does not invoke Scope SQL methods
application leaf does not import another application leaf
no `common/contracts`, `common/models`, `service_locator`, generic event-bus semantic package
only application may coordinate owner->owner facts
```

### Falsifiable contract tests later

```text
transaction callback rollback prevents all cross-owner partial state
mandatory Audit failure rolls back owning mutation
durable enqueue failure rolls back required-rendition Submission
same idempotency key + different semantic fingerprint conflicts
same completed key does not re-run mutation
replay without current disclosure authority is denied
GROUP snapshot is frozen and membership drift cannot rewrite it
empty GROUP snapshot creates no fallback authority
provider/network failure never participates in local commit
missing required predicate never ALLOWs
DecideMany and Decide use identical canonical evaluator
```

### T9 composed proofs carried forward

```text
cross-owner offboarding atomicity
User/Profile/Binding create atomicity
same-tx Audit omission impossible for required census
required durable intent cannot be lost on semantic commit
no second Authorization evaluator
```

---

## 26. Stage-boundary audit

### T8-C freezes

```text
owner public capability/query families
exact transaction Scope + Runner contract
exact Audit handoff shape/direction
exact Authorization decision input/result contract
exact GROUP mid-transition resolver direction
material technical mechanism ports
transaction-coupled OfficialRendition intent contract
idempotency acquire/complete + ReplaySnapshot ownership law
read-projection composition ownership
inside/outside transaction classification
fail-closed contract semantics
```

### T8-D remains untouched

```text
tables/schemas
repository SQL
indexes/constraints
exact lock clauses/order
transaction isolation mapping
Scope-to-*sql.Tx concrete realization
idempotency schema/retention indexes
Audit tables
River tables/schema ownership
managed-content mechanism persistence
```

### T8-E remains untouched

```text
exact HTTP paths spelling/operationIds
JSON field schemas
status codes per operation
ETag/header syntax details
Idempotency-Key wire validation grammar
exact ReplaySnapshot -> HTTP encoder mapping
OpenAPI generated Go/TS types
```

### T8-F/T8-G remain untouched

Frontend and process/deployment realization remain later stages.

No current contract forces T8-B reopen.

---

## 27. Candidate decision set

```text
T8C-D01  SELECT
  authority-aligned hybrid contract ownership

T8C-D02  SELECT
  concrete semantic-owner Services; no owner interface by default

T8C-D03  SELECT
  consumer-owned interfaces only for real technical/mid-transition consumers

T8C-D04  SELECT
  txscope.Scope = narrow database/sql executor shape
  txscope.Runner.Within = application-owned transaction lifecycle

T8C-D05  SELECT
  no *sql.Tx / pgx.Tx on semantic-owner public signatures

T8C-D06  SELECT
  owner-local Audit evidence -> mechanical application mapping -> Audit.AppendIn

T8C-D07  SELECT
  Authorization Subject/Check/Decision + Decide/DecideIn/DecideMany/DecideManyIn

T8C-D08  SELECT
  Organization SecuritySubject / SecuritySubjectIn is canonical grant-subject fact source

T8C-D09  SELECT
  Controlled Documents closed access-fact vocabulary; no generic policy language

T8C-D10  SELECT
  request-scoped EnabledGroupMembersResolver for GROUP activation

T8C-D11  SELECT
  empty enabled GROUP snapshot remains empty; no fallback/reassign; existing withdrawal recovery applies

T8C-D12  SELECT
  Authentication-owned ProviderClient using raw/primitive protocol data

T8C-D13  SELECT
  Organization role-target/owner-eligibility/deletion-dependency fact contracts

T8C-D14  SELECT
  Controlled Documents preflight opaque content proof/candidate values

T8C-D15  SELECT
  consumer-owned ManagedContent + MalwareInspector ports

T8C-D16  SELECT
  consumer-owned OfficialRenditionRenderer

T8C-D17  SELECT
  one named OfficialRenditionIntentSink; River hidden underneath

T8C-D18  SELECT
  no generic EventBus/outbox contract

T8C-D19  SELECT
  idempotency BeginIn/CompleteIn inside same Scope; no target FailReplay

T8C-D20  SELECT
  operation-local versioned application ReplaySnapshot, platform stores opaque payload

T8C-D21  SELECT
  replay rechecks live disclosure authority but never reexecutes historical mutation

T8C-D22  SELECT
  owner-fact + application composition + DecideMany read projection law

T8C-D23  SELECT
  producer-owned errors; no common errors package

T8C-D24  SELECT
  external/provider execution outside semantic Scope

T8C-D25  SELECT
  current contract reuse only where T8-A gate passes; existence/tests grant no entitlement
```

---

## 28. Reopen triggers

Reopen only the implicated T8-C decision on evidence that:

```text
standard database/sql executor shape cannot realize required PostgreSQL correctness without material adapter cost
an owner requires a second real consumer whose contract cannot sustainably use the concrete producer API
a required cross-owner interaction cannot be expressed by application routing/resolver without duplicating owner semantics
Audit evidence mapping cannot preserve owner-authored meaning mechanically
Authorization needs more than bounded owner predicate truth to evaluate T3 without absorbing domain semantics
more than a small named set of durable intents becomes Launch scope and specific ports create worse duplication than a generic mechanism
exact idempotency replay cannot be encoded deterministically from application ReplaySnapshot without wire dependency
selected interactive editor proves a backend contract materially outside existing platform mechanism seam
range-read becomes a named current viewer/file-size requirement
```

Preference, mock convenience, legacy API shape and hypothetical future integrations are not reopen triggers.

---

## 29. Independent review contract

A material independent Fable review is required before any T8-C promotion.

Reviewer must reconstruct repository authority independently and attack at minimum:

```text
1. whether concrete owner Services + consumer-owned mechanism interfaces is the Go/Method Global Maximum
2. whether txscope's database/sql executor shape overcommits T8-D or leaks SQL into application
3. whether Audit owner-local evidence -> application mapping creates a drop/semantic-mutation channel
4. whether Authorization Check + DomainPredicate preserves sole ALLOW/DENY authority without policy-language drift
5. whether GROUP resolver preserves Organization/ControlledDocs ownership and handles empty snapshots truthfully
6. whether Authentication ProviderClient can remain provider-neutral without weak map/stringly-typed authority
7. whether Group deletion / owner eligibility / RoleAssignment target fact routing remains fail-closed
8. whether ManagedContent/Malware ports preserve T4 exact-content law without provider leakage
9. whether OfficialRendition renderer + intent ports are minimal and River-compatible
10. whether idempotency BeginIn/CompleteIn + application ReplaySnapshot really satisfies T6 crash consistency and live disclosure without HTTP leaking inward
11. whether read projections avoid N+1, foreign SQL and duplicate truth without a hidden Search/read-model owner
12. whether any required T6 operation family lacks a legal internal contract path
13. whether candidate steals T8-D persistence or T8-E wire decisions
14. whether any shared abstraction can be deleted without weakening a ratified invariant
15. whether a materially better fourth contract-placement model exists
```

Required verdict:

```text
APPROVE T8-C GLOBAL MAXIMUM CANDIDATE
or
APPROVE T8-C GLOBAL MAXIMUM CANDIDATE WITH MATERIAL FIXES
or
DO NOT APPROVE T8-C GLOBAL MAXIMUM CANDIDATE
```

Reviewer output is evidence only.

---

## 30. Gate

```text
T8-B  CLOSED / OPERATOR-RATIFIED / PROMOTED

T8-C  ACTIVE
      interaction census COMPLETE at candidate level
      Global Maximum candidate MATERIALIZED AS NON-AUTHORITATIVE STAGING
      independent adversarial review = NEXT

T8-D  NOT OPEN
T8-E  NOT OPEN
T8-F  NOT OPEN
T8-G  NOT OPEN
T8-H  NOT OPEN
T9→T12 NOT OPEN
implementation BLOCKED
```

Do not promote this file to `wiki/` or open T8-D before independent review, Lead adjudication and explicit operator ratification.

---

**End of non-authoritative T8-C candidate.**