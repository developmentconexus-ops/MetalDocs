# R10 T8-C — Internal Communication Contracts

> **Status:** CLOSED / OPERATOR-RATIFIED / PROMOTED  
> **Ratified:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Upstream topology authority:** `wiki/architecture/r10-t8b-backend-module-package-topology.md`  
> **Implementation:** BLOCKED

This page is the durable target authority for R10 **T8-C — Internal Communication Contracts**.

It freezes how the five ratified semantic owners, stateless application choreography and non-semantic platform mechanisms communicate inside the T8-B modular-monolith topology. It does not freeze persistence schema/SQL/lock realization owned by T8-D, exact HTTP/OpenAPI wire encoding owned by T8-E, frontend realization owned by T8-F, runtime/process/deployment owned by T8-G, transition owned by T10 or implementation decomposition owned by T11.

The operator ratified the final package after:

```text
Global Maximum candidate
→ independent Fable Round 1
→ Lead adjudication
→ corrected candidate
→ bounded Fable Round 2
→ final Lead adjudication
→ explicit operator ratification
```

Round-1 and Round-2 reviewer artifacts remain provenance/evidence only. The staging chain is not target authority.

---

## 1. Ratified Global Maximum

```text
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL

SEMANTIC OWNERS
  concrete producer-owned public APIs
  one public package path per owner
  no owner-side service interface by default

TECHNICAL DEPENDENCIES
  narrow consumer-owned ports only for real current consumers
  stdlib / primitive / opaque technical values across mechanism seams
  provider SDK types remain inside mechanisms

APPLICATION
  stateless choreography
  sole semantic inbound orchestration class
  cross-owner facts are gathered/mapped, never re-owned

TRANSACTION
  one caller-owned local Scope
  application owns begin/commit/rollback lifecycle
  database/sql-family substrate
  no *sql.Tx / pgx.Tx / pool types in semantic-owner public signatures

AUDIT
  mutating owner authors intrinsic evidence meaning
  application mechanically maps/routes
  Audit appends in the same Scope before commit

AUTHORIZATION
  Organization authors current subject/group facts
  resource owner authors relationship/state/governance predicate facts
  Authorization alone computes final ALLOW/default-DENY

DURABLE EFFECTS
  named current intent ports only
  River remains mechanism underneath
  no EventBus / generic outbox contract

IDEMPOTENCY
  application-owned semantic fingerprint + self-contained ReplaySnapshot
  platform-owned opaque claim/replay mechanism
  durable ReplaySnapshot PII-free by construction

READS
  Authorization scope prefilter where needed
  owner canonical filtered/paginated truth
  exact current Decide/DecideMany where required
  application composition

NO
  shared/contracts
  common/models
  generic UnitOfWork
  generic ServiceLocator
  generic Repository framework
  generic policy language
  generic DomainEvent bus
```

No materially superior contract-placement model survived the Method challenge.

Method outcome:

```text
CURRENT T8-B STRUCTURE CONFIRMED
+ INTERNAL CONTRACTS REFINED TO FIT RATIFIED AUTHORITY
```

No T8-B or T1→T7 reopen is required.

---

## 2. Contract-placement law

### 2.1 Semantic-owner public APIs are concrete

Each Launch owner exposes one concrete public Go API from its single T8-B public package:

```text
authentication.Service
organization.Service
authorization.Service
controlleddocs.Service
audit.Service
```

Do not add owner interfaces merely for mocking or hypothetical alternate implementations.

A second public package path for one semantic owner remains a material T8-B reopen trigger.

### 2.2 Interfaces exist only where inversion is real

A narrow interface belongs to the package that consumes a replaceable technical mechanism or a mid-transition external fact.

Current accepted consumer-owned seam classes include:

```text
authentication.ProviderClient
controlleddocs.ManagedContentStore
controlleddocs.AdmissionClaims
controlleddocs.MalwareInspector
controlleddocs.EnabledGroupMembersResolver
application OfficialRenditionRenderer
application OfficialRenditionIntentSink
application managed-content GC mechanism port
```

Do not define an interface before a real call site exists.

### 2.3 No shared contract authority

Forbidden as cross-owner semantic dumping grounds:

```text
internal/contracts
common/models
common/types
common/errors
shared/domain
shared/repositories
shared/services
```

Primitive/stdlib types such as `uuid.UUID`, `context.Context`, `io.Reader`, byte slices and closed owner-local enums may cross the appropriate seams without creating a sixth semantic home.

### 2.4 Context

Request/job-sensitive public calls take `context.Context` first.

Context may carry cancellation, deadline, correlation and already-authenticated request plumbing. Durable User/Group/Role/Permission/document truth is never stored in Context as authority.

---

## 3. Transaction participation — `platform/txscope`

### 3.1 Deliberate substrate selection

T8-C selects the **Go `database/sql` transaction family** as the Launch internal local-transaction substrate.

Reasons:

```text
Go database/sql is a mature standard transaction/query primitive
current repository evidence already proves the narrow executor property
no named Launch consumer requires a pgx-native-only capability
one transaction substrate is simpler than maintaining parallel Row/Rows abstractions
pgx may still participate beneath database/sql through a compatible driver realization
```

River compatibility is supporting evidence, not the independent reason for the selection.

A future pgx-native-only realization that cannot satisfy this contract requires a bounded T8-C reopen backed by concrete benefit/consumer evidence.

### 3.2 Scope and Runner

Normative conceptual Go surface:

```go
type Scope interface {
    ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
    QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
    isScope()
}

type Runner interface {
    Within(ctx context.Context, fn func(Scope) error) error
}
```

The unexported marker prevents first-party packages outside `platform/txscope` from implementing `Scope` from scratch. It is defense-in-depth, not a claim that Go prevents interface embedding.

Mechanical law:

```text
no first-party package outside platform/txscope may embed txscope.Scope
```

`tools/cilint` must reject this with a RED/negative fixture.

### 3.3 Lifecycle ownership

```text
application
  owns Runner
  invokes Within
  receives Scope
  passes Scope to in-transaction owner/mechanism calls

semantic owner
  may use Scope only through owner-private persistence
  does not begin/commit/rollback

transport
  never opens Scope

platform/postgres
  realizes begin/commit/rollback in T8-D
```

`Runner.Within` guarantees:

```text
begin failure -> callback not invoked
callback error -> rollback; operation not reported successful
callback panic -> rollback best-effort; panic propagates
callback success -> commit attempted
commit failure -> operation not reported successful
Scope lifetime is bounded to the callback
```

Application may hold/pass `Scope` but may not invoke its SQL methods. This is mechanically enforced because Go cannot express "may hold this interface value but may not call its methods" through visibility alone.

### 3.4 Platform-only native SQL binding

A named platform mechanism whose external library API requires the concrete `*sql.Tx` may obtain it only through `platform/txscope`:

```go
func SQLTx(scope Scope) (*sql.Tx, error)
```

Laws:

```text
only a live Scope created by the target Runner is accepted
nil / foreign / embedded-wrapper / unrecognized Scope -> explicit fail-closed error
application may not call SQLTx
semantic owners may not call SQLTx
only explicitly catalogued platform mechanisms may call SQLTx
current named consumer = platform River adapter
```

The call-site allowlist is a target architecture rule with negative fixtures.

Distributed ad-hoc `scope.(*sql.Tx)` / `tx.(*sql.Tx)` downcasts are forbidden in the target.

### 3.5 Isolation inheritance

T8-C does not create a new isolation decision. D19 and every T8-C transaction contract inherit T2's ratified posture:

```text
PostgreSQL READ COMMITTED
+ narrow explicit serialization where required
+ OCC/CAS
+ structural constraints where required
```

Exact SQL, lock clauses/order and persistence mapping remain T8-D.

---

## 4. Same-transaction Audit evidence

T3 ordering remains:

```text
business/security mutation
→ owner-authored required evidence
→ Audit append(s)
→ COMMIT
```

### 4.1 Owner-authored evidence

Every mutation for which T3 requires same-local-commit Audit returns the owner-local evidence records required by that semantic transition.

The owner owns:

```text
actor attribution meaning
operation code
resource kind
stable resource id
historical visibility attribution input
bounded PII-minimized facts
SYSTEM-vs-USER actor meaning where applicable
```

Audit owns:

```text
AuditEvent identity
trusted event time
immutable evidence persistence/history semantics
```

Application owns neither meaning.

### 4.2 Application mapping

Application performs a mechanical field mapping from owner-local evidence to `audit.AppendInput` and may add only non-semantic correlation metadata.

Application must not:

```text
invent/change operation meaning
invent/change resource identity
invent/change owner facts
copy free-form governed content/reason into Audit unless owner evidence explicitly requires a bounded fact
change historical visibility attribution
suppress mandatory owner evidence
```

### 4.3 Audit append

Normative conceptual surface:

```go
func (s *Service) AppendIn(
    ctx context.Context,
    scope txscope.Scope,
    in AppendInput,
) (EventRef, error)
```

Multiple events are appended by multiple calls in the same Scope; no generic domain-event framework is created.

Required append failure returns an error and the caller's transaction rolls back.

---

## 5. Audit read contract

Audit owns historical event visibility attribution. Authorization owns the actor's current `audit.read` grants/scopes.

Application flow:

```text
Organization current subject
→ Authorization.AuthorizedScopes(audit.read)
→ application maps to Audit.ReadVisibility
→ Audit.ListEvents applies historical visibility filter BEFORE pagination
```

Conceptual Audit read contract includes:

```text
ReadVisibility:
  Company-wide historical visibility
  OR bounded Area-id set

ListEvents:
  Audit-owned filters permitted by T6
  ReadVisibility
  Audit-owned opaque PageCursor
  bounded limit

EventPage:
  events
  next cursor
  has_more
```

Laws:

```text
Audit evaluates historical attribution, never current grants
Authorization evaluates current grants/scopes, never historical attribution
application never post-filters a paginated Audit page
current Area/User relocation never rewrites historical Audit visibility
```

Exact wire cursor syntax/defaults remain T8-E.

---

## 6. Authorization decision contract

### 6.1 Authority partition

```text
Organization
  User existence / Company / enabled state / current GroupMembership facts

Authorization
  Role semantics
  Permission semantics
  RoleAssignments
  scope composition
  canonical grant evaluation
  final ALLOW/default-DENY

Controlled Documents
  document relationship/state/governance predicate meaning

application
  maps/routes facts only
```

### 6.2 Organization SecuritySubject

Conceptual producer-owned fact:

```go
type SecuritySubject struct {
    UserID    uuid.UUID
    CompanyID uuid.UUID
    Enabled   bool
    GroupIDs  []uuid.UUID
}
```

Organization exposes ordinary and in-Scope subject reads.

For T3 operations whose correctness depends on eligibility serializing with offboarding, Organization exposes a protected form conceptually named:

```text
ProtectedSecuritySubjectIn
```

Its semantic guarantee is:

```text
return current User eligibility + current GroupMembership facts
AND
serialize that User's eligibility against concurrent offboarding/eligibility-disable until caller Scope completes
```

The exact PostgreSQL lock mechanism remains T8-D.

### 6.3 Authorization input/result

Conceptual closed vocabulary:

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

Authorization owns which permissions require a domain predicate. Application cannot weaken requiredness by omitting the fact.

Laws:

```text
subject disabled -> DENY
Company mismatch -> DENY
no matching current grant -> DENY
required predicate + missing/unverifiable fact -> DENY
required predicate + false fact -> DENY
technical failure proving current authority -> error / never ALLOW
```

### 6.4 Decision methods

Conceptual surface:

```text
Decide
DecideIn
DecideMany
DecideManyIn
```

`DecideMany`/`DecideManyIn` use the exact same evaluator as `Decide`; they are batch execution, not another ruleset.

Batch correspondence law:

```text
len(decisions) == len(checks)
decisions[i] corresponds exactly to checks[i]
correspondence/length failure is an internal error; never implicit ALLOW
```

No durable decision cache is authority.

### 6.5 AuthorizedScopes

Authorization additionally exposes a bounded canonical query conceptually:

```text
AuthorizedScopes(subject, permission)
→ CompanyWide OR []AreaID
```

It answers only:

> Where does current grant/scope authority potentially permit this permission?

It evaluates direct/group RoleAssignments, static Role→Permission bundles and scope semantics through the same canonical Authorization authority.

It does **not** evaluate resource-owner domain predicates and never substitutes for required exact-resource `Decide`/`DecideMany`.

Target proof must include a negative case where a resource lies inside an authorized scope but fails the owner-authored predicate and is still DENIED.

---

## 7. Controlled Documents access facts

Controlled Documents exposes closed owner-local access-fact queries for the ratified T3 predicate vocabulary. There is no generic policy expression language.

Conceptual action vocabulary includes:

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

```text
Target Company scope
optional Area scope
PredicateKnown
PredicateOK
```

Rules:

```text
required fact unavailable/invalid/unverifiable -> PredicateKnown=false
application maps missing fact as not provided
Authorization DENYs when the permission requires it
technical owner read failure -> error; never synthetic true
```

`allowed_actions` consumes these exact same owner facts plus the canonical Authorization evaluator. No parallel role/action table exists.

Single/batch and ordinary/in-Scope fact queries may be provided where a named current consumer requires them.

---

## 8. GROUP-Step enabled-member resolver

Controlled Documents owns **when** a sequential Governance Step activates. Organization owns **who** is currently an enabled Group member.

Controlled Documents therefore defines the request-scoped consumer-owned seam:

```go
type EnabledGroupMembersResolver interface {
    EnabledGroupMembersIn(
        ctx context.Context,
        scope txscope.Scope,
        groupID uuid.UUID,
    ) ([]uuid.UUID, error)
}
```

Application supplies an invocation-local adapter delegating to Organization.

Laws:

```text
snapshot is taken in the same Scope as Step activation
later GroupMembership drift cannot rewrite the activated snapshot
current Authorization is still rechecked when a candidate acts
empty current enabled-member set freezes truthfully as empty
no fallback selector
no System approver
no implicit reassign/overseer engine
```

If the frozen route becomes impossible, the already-ratified recovery is withdraw/fix/resubmit or the dedicated obsolescence withdrawal path.

The resolver pattern is not a ServiceLocator and may not become one.

---

## 9. Organization cross-owner fact contracts

Organization owns bounded facts needed by other owner capabilities without deciding the consuming owner's semantics.

### 9.1 Responsible-owner eligibility

Conceptual in-Scope fact:

```text
resolved
existing User
same Company
enabled
```

This is exactly the D4 target. It grants no document access.

Where D4/T3 requires concurrency with offboarding, the target eligibility read is protected/serializing.

### 9.2 RoleAssignment target facts

Organization supplies only bounded identity/scope existence and eligibility facts required by Authorization to validate a grant target.

New direct User RoleAssignment uses protected target eligibility.

Authorization owns the grant decision and persistent RoleAssignment truth.

### 9.3 Group deletion dependencies

Before Organization deletes a Group, application gathers in one Scope:

```text
Authorization live Group RoleAssignment dependency fact
Controlled Documents current GovernanceRoute / unresolved live GROUP-Step dependency fact
```

Organization receives explicit resolved dependency facts and rejects deletion when any required source is unresolved or when a live dependency exists.

Organization remains the deletion owner; application does not decide.

---

## 10. Owner VersionToken / replacement preconditions

For T6 whole-replacement mutable resources, the owning semantic owner exposes an opaque owner-side `VersionToken` with the current read and consumes an expected token on replacement.

Semantic law:

```text
read -> current owner truth + VersionToken
mutation -> expected VersionToken
expected != current -> stale/precondition error + zero mutation
```

Exact token storage/generation remains T8-D. ETag quoting/encoding and `If-Match` wire syntax remain T8-E.

This law covers the T6 resources requiring strong preconditions, including Company/Profile/eligibility/binding/Area/Group/DocumentType/governance/eligible-templates/responsible-owner/template-role replacements.

DRAFT WorkingContent retains its separately ratified monotonic generation/OCC authority.

Exact already-current repeat may be an owner no-op:

```text
no duplicate semantic Audit evidence
no version advance merely for the repeat
return the current authoritative VersionToken
```

---

## 11. Authentication provider-protocol seam

`platform/identityprovider` owns raw provider protocol handling. Authentication owns ProviderSubjectBinding, ApplicationSession and MetalDocs anti-corruption meaning.

Authentication defines a consumer-owned seam using primitive/stdlib values only.

Conceptual capabilities:

```text
AuthorizationURL(...)

ExchangeAuthorizationCode(...)
  -> verified issuer string
  -> verified subject string

SearchSubjects(ctx, query, emit)
  emit(ref string, displayHints []string) error

ResolveSubjectRef(ref)
  -> verified issuer string
  -> verified subject string
```

Directory-search laws:

```text
bounded selection journey
emit is synchronous enumeration, not durable streaming
non-nil emit error aborts enumeration and propagates
ref is opaque provider-subject reference
displayHints are bounded presentation hints only
```

Provider roles/groups/permissions/raw claim bags do not cross as Organization/Authorization truth.

Authentication public capabilities cover browser login/callback preflight, binding resolution/create/replace, session issue/resolve/revoke and current provider-subject search required by T6.

Provider/network exchange occurs outside the local semantic transaction; binding/session state mutations occur in caller-provided Scope.

Session issuance requires protected current enabled-User truth from Organization.

---

## 12. Controlled Documents public capability/query census

Controlled Documents remains one public owner surface; owner-private responsibility decomposition is ungated.

Its public API must cover the Launch families already ratified by T6/T1→T5, including at least:

### Read/query families

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
ResponsibleOwner
TemplateRole
OfficialRenditionContent
AccessFacts / AccessFactsIn / batch variants
GroupGovernanceUsageIn
RenditionWorkCandidate
managed-content semantic-reference proof for GC
```

### Mutation families

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

Release is system-owned and occurs inside the appropriate Controlled Documents transition; no user publish mutation is created.

No separate obsolescence-completion command is created:

```text
NoHumanApproval completion -> InitiateObsolescenceIn
human final ACCEPT completion -> DecideGovernanceStepIn
```

Both emit the required owner-authored completion evidence when completion occurs.

A mutation result contains only:

```text
canonical semantic result needed by application
required owner-authored Audit evidence
zero-or-one named durable intent when that transition activates one
```

No generic `DomainEvent[]` is returned.

---

## 13. External/preflight content preparation

Provider/scanner/renderer work never participates as a required external effect inside the semantic transaction.

Controlled Documents may expose bounded preparation capabilities for current consumers such as:

```text
PrepareTemplateSeed
PrepareNextRevisionSeed
StartDraftUpload
CompleteDraftUpload
PrepareSubmissionGovernedContent
OpenSemanticSource
PrepareOfficialRenditionCandidate
```

Prepared values must be unforgeable/opaque from the caller's semantic perspective and are revalidated against canonical owner state at final commit.

Stale/abandoned preparation creates no semantic fact; leftover bytes remain reclaimable mechanism state.

---

## 14. ManagedContentStore

Controlled Documents is the semantic consumer. Managed content remains mechanism only.

Conceptual consumer-owned port:

```go
type ManagedContentStore interface {
    Allocate(ctx context.Context) (uuid.UUID, error)
    PresignCreate(ctx context.Context, handle uuid.UUID, maxBytes int64, ttl time.Duration) (string, error)
    Stat(ctx context.Context, handle uuid.UUID) (exists bool, sizeBytes int64, err error)
    OpenExact(ctx context.Context, handle uuid.UUID) (io.ReadCloser, error)
    CopyToNewHandle(ctx context.Context, source uuid.UUID, destination uuid.UUID) error
    DeleteReclaimable(ctx context.Context, handle uuid.UUID) error
}
```

No provider bucket/key/version/ETag is product identity.

`PresignCreate` means:

```text
create once / no overwrite
```

Every production provider implementation must prove the property using an appropriate provider primitive.

`OpenRange` is not a Launch baseline contract because T4/T6 keep range reads optional until a real viewer/file-size consumer activates it.

ExactContentDescriptor remains server-derived semantic truth:

```text
SHA-256
size_bytes
closed actual ContentFormat
```

Client-declared/provider checksums are hints/evidence only.

---

## 15. AdmissionClaims

Managed-content admission uses a separate bounded mechanism seam; there is no generic semantic ownership registry.

Conceptual operations use opaque primitive claim identity:

```text
Reserve(handle) -> claim id
ProveLive(claim id)
ConsumeIn(scope, claim id)
Release(claim id)
```

Laws:

```text
claim binds one handle to one in-flight authorized semantic attachment attempt
live claim protects content from GC eligibility
claim is unforgeable from ordinary caller input
ConsumeIn shares the semantic attachment transaction
rollback cannot produce committed attachment without consumed binding
explicit release makes abandoned content reclaimable
unconsumed claim has bounded mechanism expiry
```

No `owner_type/owner_id` generic registry is created. Claim duration/storage/expiry realization belongs to T8-D/T8-G as appropriate.

---

## 16. Malware inspection

Controlled Documents defines the narrow mechanism port:

```go
type MalwareInspector interface {
    Inspect(ctx context.Context, content io.Reader) (digest [32]byte, clean bool, err error)
}
```

Laws:

```text
clean=false,nil -> malicious; immutable governed admission forbidden
err!=nil -> unavailable/incomplete; fail closed and retry visibly
clean=true,nil -> usable only when returned digest == ExactContentDescriptor.sha256
```

The scanner owns no Document/Submission meaning and no business scan lifecycle.

---

## 17. OfficialRendition execution and read contracts

### 17.1 Renderer

The rendition application path owns a narrow renderer port conceptually:

```go
type OfficialRenditionRenderer interface {
    RenderPDF(ctx context.Context, source io.Reader) (io.ReadCloser, error)
}
```

Renderer/provider job ids never become semantic identity.

### 17.2 Durable intent

When a frozen Submission requires OfficialRendition, Controlled Documents returns a named intent containing only the stable Submission identity and required format.

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

The platform River adapter obtains the native `*sql.Tx` only through `txscope.SQLTx` and uses River's transaction-coupled insert primitive.

Atomicity law:

```text
required-rendition Submission semantic transition commits
<=>
required durable rendition intent commits in the same Scope
```

No generic EventBus/outbox is added.

### 17.3 Worker flow

```text
transport/jobs
→ rendition application leaf
→ ControlledDocs.RenditionWorkCandidate
→ open exact source
→ renderer outside semantic tx
→ prepare exact rendition candidate
→ Scope
→ ControlledDocs.CompleteOfficialRenditionIn
→ required Audit append
→ system Release inside ControlledDocs if all gates now pass
→ commit
```

If final canonical eligibility is gone, semantic finalization is no-op/fail-closed as applicable and produced bytes remain reclaimable mechanism state.

### 17.4 Rendition content read

OfficialRendition is a distinct semantic byte resource, not Release source.

```text
current EFFECTIVE rendition -> READ_EFFECTIVE authorization path
historical rendition -> READ_HISTORY authorization path
governance decision view -> exact immutable Submission source, not rendition
```

Provider handle/key never becomes public identity.

---

## 18. Idempotency

T6 remains authority for which operations require durable `Idempotency-Key`. T8-C freezes only the internal contract.

### 18.1 Platform mechanism surface

Application may import `platform/idempotency` per T8-B.

Conceptual contract:

```text
BeginIn(scope, actor, operationID, key, semanticFingerprint)
  -> exactly one of:
       claim
       completed replay payload
       semantic fingerprint conflict
       technical error

CompleteIn(scope, claim, replayPayload)
```

No target `FailReplay` operation is required because an uncommitted claim belongs to the same local transaction and disappears on rollback.

### 18.2 Semantic fingerprint

Application derives the fingerprint deterministically from validated semantic command fields plus canonical operation identity.

It is not architecturally defined as a hash of raw HTTP path/query/body bytes.

The platform treats fingerprint bytes as opaque equality material.

### 18.3 Concurrent same-key law

D19 inherits T2's ratified PostgreSQL READ COMMITTED posture.

```text
same scoped key + same fingerprint requests serialize
winner commit -> loser returns completed replay without poisoning its Scope
winner rollback -> a waiting contender may become claim owner
same key + different fingerprint -> conflict; no business mutation
no public/durable IN_PROGRESS business state
```

T8-D chooses the SQL/upsert/locking realization that satisfies this law. A realization that turns an expected same-key race into an aborted caller Scope does not satisfy T8-C.

### 18.4 ReplaySnapshot

For each T6 durable-idempotent POST, the owning application leaf defines a versioned operation-local ReplaySnapshot.

Ratified representation law:

```text
PII-FREE BY CONSTRUCTION with respect to erasable UserProfile data
SELF-CONTAINED
SNAPSHOT-ONLY RECONSTRUCTION
```

A completed durable replay response is reconstructed deterministically from the stored ReplaySnapshot version alone.

Replay must not query current mutable state or depend on later canonical resource existence to reconstruct the historical success response.

Durable snapshot may contain stable semantic identifiers and stable non-erasable immutable result facts required for exact replay.

Baseline snapshot excludes:

```text
erasable UserProfile fields
provider claims/tokens/raw provider directory payload
request body/header copies
raw HTTP response bytes
full governed content
free-form governance/cancellation/obsolescence/feedback text when not required for response reconstruction
```

Free-form exclusion is a snapshot-minimality/duplicate-retention decision, not a claim that all such text is erasable PII.

There is no Launch replay purge/redaction subsystem.

T8-E must design replay-required success representations so exact status/body reconstruction is possible from the self-contained allowed snapshot facts.

If a future promoted requirement forces excluded erasable/free-form data into a replay-required exact success body, reopen this decision and select the smallest explicit representation/erasure mechanism then.

### 18.5 Replay disclosure

On completed replay:

```text
re-authenticate current request
re-check current permission/scope
re-check minimum current resource visibility necessary to disclose the stored result
DO NOT re-run historical mutation
DO NOT re-run original lifecycle preconditions
only after live disclosure authorization -> encode stored ReplaySnapshot
```

There is no generic replay ACL.

---

## 19. Read-projection composition

Purpose-built T6 views remain application lens results, never persistent cross-owner truth.

General law:

```text
Authorization.AuthorizedScopes where a scope prefilter is required
→ owning domain canonical filtered/paginated query
→ owner access facts/domain predicates where required
→ Authorization Decide/DecideMany for exact candidate actions
→ bounded Organization display references
→ application lens result
→ T8-E wire encoding
```

This prevents:

```text
unbounded enumerate-all-then-filter
filter-after-pagination incoherence
application re-deriving RoleAssignment semantics
foreign SQL
persistent duplicate read authority
```

### 19.1 Library

```text
subject
→ AuthorizedScopes(document.read_effective)
→ ControlledDocs.LibraryCandidates(scopes, canonical filters, owner cursor)
→ exact owner/domain truth
→ final exact Authorization checks where material
→ bounded Organization refs for display
```

Search remains canonical PostgreSQL query/view under Controlled Documents owner-private persistence; materialized Search remains OFF.

### 19.2 Audit

```text
subject
→ AuthorizedScopes(audit.read)
→ Audit.ReadVisibility
→ Audit.ListEvents historical filter before pagination
```

### 19.3 Other lenses

The same law covers SessionView, My Work, Document Official, Document Work, Governance Case, Document History, Audit and Administration without turning any lens into a semantic owner.

---

## 20. T5-J managed-content GC choreography

T5-J is hosted by:

```text
internal/application/maintenance
```

This is a non-semantic maintenance application leaf inside the existing T8-B `application` class. It is not a product route/workspace, owner, storage domain or retention domain. No T8-B reopen is required.

### Phase 1 — eligibility transaction

```text
technical bounded GC candidate
→ ControlledDocs proves no current WorkingContent reference
→ ControlledDocs proves no immutable Submission/Rendition/imported governed reference
→ AdmissionClaims proves no live claim/binding
→ backup mechanism proves no pin/exclusion
→ mark technical GC_PENDING
→ commit
```

### Phase 2 — immediate pre-delete proof

Immediately before provider delete:

```text
re-prove GC_PENDING
→ re-prove no current WorkingContent reference
→ re-prove no immutable Submission/Rendition/imported governed reference
→ re-prove no live admission claim/binding
→ re-prove no backup pin/exclusion
→ only then DeleteReclaimable outside semantic transaction
→ finalize technical state
```

The second Controlled Documents semantic-reference proof is mandatory.

Safe failure is leaked storage, never lost governed truth.

Exact GC tables/leases/locks/expiry and runtime schedule remain T8-D/T8-G.

---

## 21. Inside / outside transaction classification

### Outside local semantic transaction

```text
OIDC authorization/token/provider-directory network calls
managed-content browser upload
managed-content exact-byte provider read/copy preparation
malware inspection
OfficialRendition renderer execution
provider physical delete
ordinary non-linearizable read projections unless a command requires in-Scope truth
```

### Inside caller-provided Scope

```text
all semantic/security mutations
command-time protected Organization eligibility facts where required
command-time Authorization decisions
command-time ControlledDocs relationship/state/governance facts
GROUP-Step enabled-member snapshot resolution at activation
required Audit append
required River durable-intent insert
idempotency BeginIn/CompleteIn for T6 durable-idempotent creations
AdmissionClaim consume when coupled to semantic attachment
```

External/provider call required for semantic success inside Scope is forbidden.

---

## 22. Critical cross-owner flows

### Session issuance

```text
provider verification outside Scope
→ Scope
→ Authentication.ResolveProviderBindingIn
→ Organization.ProtectedSecuritySubjectIn(bound User)
→ Authentication.IssueSessionIn using resolved enabled fact
→ commit
```

Unknown binding / disabled User fails closed.

### User creation

```text
provider subject resolve outside Scope
→ Scope
→ protected current actor subject
→ Authorization.DecideIn(organization.manage)
→ idempotency BeginIn when T6 requires it
→ Organization.CreateUserIn
→ Authentication.BindProviderSubjectIn
→ owner evidence → Audit.AppendIn
→ CompleteIn(PII-free ReplaySnapshot)
→ commit
```

No half-created User/Profile/Binding may commit.

### Offboarding

```text
Scope
→ protected current actor + authorization
→ serialize target eligibility
→ Organization disable User
→ Authentication revoke all ApplicationSessions
→ Organization remove current GroupMemberships
→ Authorization revoke direct User RoleAssignments
→ append all required owner-authored evidence
→ commit
```

Group RoleAssignments and ProviderSubjectBinding remain according to T3. Re-enable never resurrects removed memberships/grants/sessions.

### GroupMembership mutation

Organization serializes target eligibility internally for its same-owner User+GroupMembership mutation. Actor authorization uses protected current subject.

### Direct User RoleAssignment create

```text
Scope
→ protected actor + access.manage
→ idempotency BeginIn when required
→ protected Organization target facts
→ Authorization.CreateRoleAssignmentIn
→ Audit.AppendIn
→ CompleteIn(PII-free snapshot)
→ commit
```

### Group deletion

```text
Scope
→ protected actor + organization.manage
→ Authorization GroupRoleAssignment usage fact
→ ControlledDocs Group governance usage fact
→ Organization.DeleteGroupIn(resolved dependency facts)
→ Audit.AppendIn
→ commit
```

### Document create / next Revision / SUBMIT

Provider/template/content/scanner preparation occurs outside Scope where required.

Inside Scope:

```text
protected actor subject
→ ControlledDocs exact access facts
→ Authorization DecideIn
→ protected responsible-owner eligibility where deliberately selecting another owner
→ ControlledDocs mutation
→ GROUP resolver if Step activation occurs
→ named rendition intent enqueue when returned
→ required Audit append
→ idempotency CompleteIn when required
→ commit
```

### Governance decision

```text
Scope
→ protected actor subject
→ ControlledDocs governance-action facts
→ Authorization.DecideIn(governance.act)
→ ControlledDocs.DecideGovernanceStepIn
→ next GROUP snapshot resolver where activated
→ system Release if final gate now satisfied
→ required Audit append
→ commit
```

### Responsible-owner replacement

```text
Scope
→ protected actor + exact owner-manage facts
→ Authorization.DecideIn(document.owner.manage)
→ protected Organization target eligibility
→ ControlledDocs.ChangeResponsibleOwnerIn
→ Audit.AppendIn
→ commit
```

If offboarding linearizes first, target eligibility is false and replacement fails closed.

---

## 23. Failure / fail-closed law

Owner errors remain producer-owned. No common error package is created.

Owners expose enough typed/sentinel distinctions for application to distinguish material categories such as:

```text
not found
invalid/stale state
precondition conflict
eligibility failure
duplicate/singleton conflict
invalid semantic input
```

T8-E maps outcomes to exact RFC 9457 Problem Details/code/status. Owners do not import wire errors.

Fail-closed examples:

```text
missing required AuthZ/domain truth -> DENY/no mutation
IdP unavailable -> no successful login/provider preflight
malware unavailable -> no governed admission
managed content missing/corrupt -> no semantic attachment/read success
renderer failure -> no silent SourceOnly downgrade
required River enqueue failure -> semantic rollback
required Audit append failure -> semantic rollback
foreign/unrecognized tx Scope native binding -> error
```

Never convert mechanism failure into fake approval, implicit ALLOW, silent Release, provider-role authority, stale replay disclosure or alternate content identity.

---

## 24. Selective reuse disposition

T8-A gate applied to contracts/properties:

```text
current narrow database/sql executor property
  PRESERVE PROPERTY / REHOME + REFINE

current TxRunner begin/commit/rollback lifecycle property
  PRESERVE PROPERTY

current TxRunner exact contract / *sql.Tx callback / tenant-GUC behavior
  REWRITE

current distributed tx.(*sql.Tx) River downcasts
  REWRITE / REPLACE WITH txscope-owned native binding

current IAM authz.Require
  REWRITE

current Audit application/export public contract
  REWRITE

current objectstore contract
  REWRITE

current idempotency HTTP middleware contract
  REWRITE

current idempotency concurrency ideas
  EVIDENCE ONLY for T8-D; expected same-key races may not poison target Scope

River transactional insertion property
  PRESERVE / REFINE behind named sink

River concrete client/driver/transaction types
  DO NOT EXPOSE

OpenAPI/codegen contract-first property
  PRESERVE; exact wire remains T8-E
```

No current interface survives merely because tests exist.

---

## 25. Mechanical enforcement and proof obligations

T8-C carries these falsifiable architecture/proof obligations forward:

```text
closed-world package/edge classification
application cannot invoke Scope SQL methods
owner/application cannot call SQLTx
non-txscope packages cannot embed txscope.Scope
foreign/unrecognized Scope -> SQLTx fails closed
required Audit failure rolls back business mutation
Audit historical visibility filters before pagination
AuthorizedScopes cannot substitute for exact Decide/DecideMany
protected eligibility read serializes with offboarding
VersionToken stale replacement -> zero mutation
exact no-op replacement returns current token / no duplicate Audit
GROUP membership snapshot frozen; empty remains empty
ManagedContent create-once/no-overwrite
AdmissionClaim rollback/consume/release safety
GC performs initial and immediate pre-delete semantic-reference proofs
malware CLEAN digest equals exact admitted bytes
required OfficialRendition enqueue shares semantic transaction
idempotency concurrent loser does not poison Scope under T2 READ COMMITTED
same-key different semantic fingerprint performs no business mutation
completed replay is self-contained, exact, live-authorized and PII-free
OfficialRendition read never exposes provider identity
```

`tools/cilint`/`tools/verify` exact target rules are materialized by the later proof/execution stages; T8-C freezes the protected properties.

---

## 26. Ratified decision set

The complete T8-C decision set is:

```text
D01  authority-aligned hybrid contract ownership
D02  concrete semantic-owner Services; no owner interface by default
D03  consumer-owned interfaces only for real technical/mid-transition consumers
D04  database/sql-family txscope + application-owned Runner.Within
D05  no *sql.Tx/pgx.Tx in semantic-owner public signatures
D06  owner-local Audit evidence -> application mapping -> same-Scope Audit append + Audit read contract
D07  canonical Decide/DecideIn/DecideMany/DecideManyIn + AuthorizedScopes
D08  Organization SecuritySubject + protected eligibility form
D09  closed ControlledDocs access-fact vocabulary; no policy language
D10  request-scoped EnabledGroupMembersResolver
D11  empty GROUP snapshot remains empty; no fallback/reassign
D12  consumer-owned provider seam with verified primitive issuer+subject + bounded directory refs
D13  bounded Organization target/dependency facts; protected target eligibility where required
D14  opaque preparation values + explicit admission-claim lifecycle
D15  ManagedContentStore + AdmissionClaims + MalwareInspector + DeleteReclaimable
D16  consumer-owned OfficialRenditionRenderer
D17  one named OfficialRenditionIntentSink + txscope native binding
D18  no generic EventBus/outbox
D19  BeginIn/CompleteIn same Scope; no FailReplay; non-poisoning concurrent outcome under T2 READ COMMITTED
D20  operation-local versioned PII-free ReplaySnapshot
D21  live replay-disclosure recheck; no mutation/precondition re-execution
D22  AuthorizedScopes prefilter + owner canonical pagination + exact final decision/read composition
D23  producer-owned errors; no common errors package
D24  external/provider execution outside semantic Scope
D25  selective reuse only where T8-A five-part gate passes
D26  owner VersionToken + expected-version replacement contract
D27  PII-free replay by construction; no Launch replay-purge subsystem
D28  ManagedContent admission-claim + T5-J GC contract family
D29  protected eligibility-read semantic guarantee; lock realization T8-D
D30  Audit historical-visibility read contract
D31  Authorization AuthorizedScopes query from canonical evaluator
D32  D19 explicitly inherits ratified T2 READ COMMITTED posture
D33  Scope sealing is defense-in-depth + no external embedding + SQLTx fail-closed
D34  application/maintenance hosts T5-J GC choreography without T8-B reopen
D35  two-phase GC with full immediate pre-delete semantic/live-reference re-proof
D36  self-contained snapshot-only replay reconstruction
D37  free-form replay exclusion is snapshot minimality; T8-E response must remain reconstructible
D38  database/sql selected from standard primitive + single-substrate + no current pgx-native consumer evidence
D39  PresignCreate = create-once/no-overwrite property
D40  bounded synchronous provider-directory callback + propagated callback failure + bounded display hints
D41  explicit proof AuthorizedScopes never grants exact resource action
D42  exact owner no-op replacement returns current VersionToken with no version/Audit fabrication
```

---

## 27. Stage boundaries

### T8-C closed authority

T8-C owns:

```text
internal contract ownership/direction
owner public capability/query families
transaction participation contract
Audit handoff/read contracts
Authorization decision/scope contracts
cross-owner fact/resolver contracts
material mechanism ports
named transaction-coupled intent contract
idempotency internal/replay representation law
read-composition law
inside/outside transaction classification
fail-closed contract semantics
```

### T8-D owns next

```text
target schemas/tables
persistent state ownership mapping
PK/FK/unique/check constraints
owner-private SQL/query realization
immutable/history relational shapes
WorkingContent/OCC persistence
Submission/Release/effectivity constraints
Organization/AuthZ/ApplicationSession persistence
Audit evidence persistence
managed-content technical persistence + admission claims + GC_PENDING
idempotency persistence
River technical persistence boundary
exact database/sql Runner realization
exact transaction/serialization/lock mapping under T2/T8-C laws
```

### T8-E owns later

```text
exact HTTP paths/operationIds/schemas
JSON fields/enums/nullability
status/problem mapping
ETag/If-Match encoding
Idempotency-Key wire grammar
pagination cursor encoding
exact ReplaySnapshot -> status/body encoder
OpenAPI generated Go/TS boundary
```

T8-F/T8-G/T10/T11 remain later stages.

---

## 28. Reopen triggers

Reopen only the implicated T8-C decision on material evidence such as:

```text
database/sql family cannot realize required PostgreSQL correctness sustainably
named pgx-native-only/current consumer makes the single-substrate choice materially worse
required cross-owner interaction cannot be expressed through application routing/resolver without duplicating semantics
Audit evidence mapping cannot preserve mandatory owner meaning mechanically
Authorization requires more than bounded owner predicate truth without absorbing resource-domain semantics
specific durable-intent ports become materially worse than a generic technical mechanism because multiple real Launch consumers appear
exact replay cannot be reconstructed from PII-free self-contained ReplaySnapshot without violating promoted wire/product requirements
selected provider/editor/viewer creates a backend mechanism contract outside the existing platform class
range reads become a named current viewer/file-size requirement
T8-D proves the accepted transaction/persistence realization cannot satisfy a frozen T8-C contract
```

Preference, mock convenience, legacy API shape and hypothetical future integrations are not reopen triggers.

---

## 29. Closure

```text
T8-C Internal Communication Contracts = CLOSED / OPERATOR-RATIFIED / PROMOTED
```

Independent review convergence at closure:

```text
Round 1  Global Maximum class confirmed
Round 2  BLOCKER 0 / surviving material contradiction 0
B5       Round-1 blocker not sustained
PII-free replay selection independently upheld
third Fable round not required
```

No upstream reopen occurred.

The next active stage is:

```text
T8-D — Persistence Realization
```

Implementation remains **BLOCKED**.
