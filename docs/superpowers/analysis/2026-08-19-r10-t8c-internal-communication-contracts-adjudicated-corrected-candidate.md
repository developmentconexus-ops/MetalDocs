# R10-T8C — Internal Communication Contracts — Adjudicated Corrected Candidate

```text
ADJUDICATED CORRECTED CANDIDATE
NON-AUTHORITATIVE STAGING
BOUNDED ROUND-2 FABLE INPUT
NOT TARGET AUTHORITY
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Revalidated pre-correction remote HEAD:** `ffb6125a37bf32878f68bbc5109be17ee7bc2f83`  
> **Stage:** T8-C ACTIVE  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Round-1 evidence:** `docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-independent-fable-review.md`  
> **Original candidate:** `docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-global-maximum-candidate.md`  
> **Implementation:** BLOCKED

This document is the Lead-adjudicated corrected candidate for bounded Round 2. It preserves the Round-1-confirmed Global Maximum class and incorporates only corrections or explicit decisions that survived technical adjudication. It is not durable authority and does not open T8-D.

---

## 0. Adjudication result

Round-1 Fable verdict:

```text
APPROVE T8-C GLOBAL MAXIMUM CANDIDATE WITH MATERIAL FIXES
BLOCKER 5 / MAJOR 6 / LOW 5
GLOBAL MAXIMUM CONFIRMED
T8-B REOPEN NO
T1→T7 REOPEN NO
```

Lead disposition:

```text
B1  ACCEPT
B2  ACCEPT
B3  ACCEPT
B4  ACCEPT
B5  REJECT AS BLOCKER; preserve required concurrent replay behavior and defer SQL realization to T8-D

M1  ACCEPT
M2  ACCEPT
M3  ACCEPT WITH NARROWING
M4  ACCEPT
M5  ACCEPT AS A REAL UNMADE T8-C DECISION; SELECT PII-FREE REPLAY SNAPSHOT BY CONSTRUCTION
M6  ACCEPT

L1-L5 ACCEPT as bounded precision corrections
```

No accepted correction changes the selected ownership/model class. No correction requires a T8-B or T1→T7 reopen.

---

## 1. Confirmed Global Maximum

```text
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL

SEMANTIC OWNERS
  concrete producer-owned public APIs
  one public package path per owner
  no mandatory owner-side service interface

TECHNICAL DEPENDENCIES
  narrow consumer-owned interfaces only for real mechanism/mid-transition consumers
  no provider SDK types across semantic boundaries

TRANSACTION
  one caller-owned local Scope
  application owns begin/commit/rollback lifecycle
  owners participate explicitly
  database/sql-family internal transaction substrate selected deliberately
  semantic-owner public signatures never expose *sql.Tx / pgx.Tx / pool types

CROSS-OWNER FACTS
  known before mutation -> application gathers/maps
  discovered mid-transition -> bounded request-scoped consumer-owned resolver
  no owner->owner import

AUDIT
  mutating owner authors intrinsic evidence meaning
  application maps/routes only
  Audit append is same-Scope and required before commit

AUTHORIZATION
  Organization authors current subject/scope facts
  resource owner authors relationship/state/governance predicate facts
  Authorization alone evaluates final ALLOW/default-DENY

DURABLE INTENT
  named intent-specific ports only
  River remains mechanism underneath
  no EventBus/generic outbox contract

IDEMPOTENCY
  platform mechanism owns scoped claim/replay storage only
  application owns semantic fingerprint + operation-local ReplaySnapshot
  durable replay snapshots are PII-free by construction

READ PROJECTIONS
  Authorization can enumerate authorized scopes
  owner canonical facts are filtered/queried coherently
  application composes bounded lens results
  no foreign SQL / persistent duplicate truth

NO
  shared/contracts
  common/models
  generic UnitOfWork
  generic ServiceLocator
  generic Repository interface family
  generic policy language
  generic DomainEvent bus
```

---

## 2. Contract-placement law remains unchanged

### Semantic-owner API

Each Launch owner exposes a concrete public API from its single T8-B public package:

```text
authentication.Service
organization.Service
authorization.Service
controlleddocs.Service
audit.Service
```

Do not add owner interfaces merely for mocks.

### Real inversion only

Consumer-owned interfaces exist only for named current technical/mid-transition consumers such as:

```text
authentication.ProviderClient
controlleddocs.ManagedContentStore
controlleddocs.AdmissionClaims
controlleddocs.MalwareInspector
controlleddocs.EnabledGroupMembersResolver
application rendition renderer port
application OfficialRendition durable-intent sink
application managed-content GC mechanism port where required
```

No shared/common contract package is introduced.

---

## 3. T8C-D04 / D05 / D17 correction — txscope + River

### 3.1 Deliberate substrate selection

T8-C now explicitly selects:

```text
database/sql transaction family
```

as the Launch internal local-transaction substrate.

This is a deliberate constraint on T8-D, not an accidental consequence of method signatures. T8-D may select the PostgreSQL driver/stdlib realization behind `database/sql`, but a pgx-native-only transaction realization that cannot satisfy the frozen contract requires a bounded T8-C reopen with concrete benefit evidence.

Reason:

```text
current Go standard primitive is mature
current repo already proves database/sql execution shape
River v0.37.1 + riverdatabasesql has a named current consumer requiring *sql.Tx
inventing custom Row/Rows abstractions solely to preserve an unused pgx-native option adds accidental complexity
```

### 3.2 Scope surface

Conceptual target contract:

```go
type Scope interface {
    ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
    QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row

    // unexported marker in the owning package seals realizations to txscope
    isScope()
}

type Runner interface {
    Within(ctx context.Context, fn func(Scope) error) error
}
```

The unexported marker means arbitrary first-party packages cannot implement alternative Scope realizations by convenience. Exact concrete wrapper/constructor fields remain T8-D.

### 3.3 Platform-only native binding

Some platform mechanisms require the concrete `*sql.Tx`, currently River transactional insertion.

`platform/txscope` therefore owns one explicit native SQL binding operation conceptually equivalent to:

```go
func SQLTx(scope Scope) *sql.Tx
```

Binding laws:

```text
Scope is sealed by txscope
SQLTx is valid for a live Scope created by the target Runner
semantic owners MUST NOT call SQLTx
application MUST NOT call SQLTx
only named platform mechanisms whose external API requires the concrete database/sql transaction may call it
first current consumer = platform River adapter
```

`tools/cilint` must reject `SQLTx` calls outside the explicit platform mechanism allowlist.

This removes the current runtime downcast pattern while keeping provider identity out of semantic-owner signatures.

### 3.4 Application SQL remains forbidden

Application may hold/pass `Scope` but may not call its SQL methods. Static architecture proof must reject application calls to `ExecContext`, `QueryContext`, `QueryRowContext` and `SQLTx`.

### 3.5 Runner law unchanged

```text
begin failure -> callback not invoked
callback error -> rollback; original error observable
callback panic -> rollback best-effort; panic propagates
callback nil -> commit
commit failure -> operation not reported successful
Scope cannot remain valid outside callback lifetime
```

Exact SQL transaction creation/options remain T8-D.

---

## 4. T8C-D06 correction — Audit write + read contracts

### 4.1 Same-transaction write

Round-1 direction remains accepted:

```text
owner mutation
-> owner-local required Evidence values
-> application mechanical mapping
-> Audit.AppendIn(scope,...)
-> commit only after all required append(s) succeed
```

Owner owns operation/resource/visibility/facts meaning. Audit owns immutable AuditEvent identity and trusted event time. Application may add bounded correlation metadata only.

### 4.2 Audit read visibility

Audit reads must not require Audit to import Authorization and must not make application a historical-visibility evaluator.

Authorization first resolves where the current actor holds `audit.read`; application maps that result into an Audit-owned bounded visibility input.

Conceptual Audit-owned types:

```go
type ReadVisibility struct {
    CompanyID   uuid.UUID
    CompanyWide bool
    AreaIDs     []uuid.UUID
}

type PageCursor string // Audit-owned opaque continuation token

type ListEventsQuery struct {
    Visibility ReadVisibility
    Cursor     PageCursor
    Limit      int
}

type EventPage struct {
    Events     []Event
    NextCursor PageCursor
    HasMore    bool
}

func (s *Service) ListEvents(
    ctx context.Context,
    q ListEventsQuery,
) (EventPage, error)
```

Minimum law:

```text
CompanyWide=true -> Company-attributed + every Area-attributed event in Company
CompanyWide=false -> only historical event attribution in AreaIDs
current resource relocation never rewrites visibility
Audit applies historical attribution filter before pagination
application never post-filters a paginated Audit page
```

Exact public filter/header/JSON shape remains T8-E. Additional Audit filters are not invented in T8-C absent a ratified consumer.

---

## 5. T8C-D07 / D22 correction — Authorization scope enumeration

`Decide`, `DecideIn`, `DecideMany` and `DecideManyIn` remain canonical per-target evaluators.

T8-C adds one Authorization-owned query for the different question: **where is this actor granted this permission?**

Conceptual target:

```go
type AuthorizedScopeSet struct {
    CompanyWide bool
    AreaIDs     []uuid.UUID
}

func (s *Service) AuthorizedScopes(
    ctx context.Context,
    subject Subject,
    permission Permission,
) (AuthorizedScopeSet, error)
```

Laws:

```text
same canonical current direct/group RoleAssignment authority
same static Role->Permission bundles
same scope semantics
no domain predicate evaluation in scope enumeration
CompanyWide dominates individual AreaIDs for that Company
AreaIDs contain only currently granted matching Areas
technical failure -> error, never synthetic grant
```

This is not a second evaluator and not a generic policy query language.

Use cases include:

```text
document-creation/options eligible Area prefilter
audit.read historical visibility input
Library/catalog scope prefilter before owner pagination
bounded admin selection/read paths where T6 needs "where may actor do X?"
```

Resource/domain predicates remain owner-authored and are evaluated through `Decide`/`DecideMany` when required.

### DecideMany correspondence

```text
len(decisions) == len(checks)
decisions[i] corresponds exactly to checks[i]
```

A mismatch is an internal error; never silently reorders authorization results.

---

## 6. T8C-D08 / M1 correction — eligibility serialization contract

T3 requires certain actions to **serialize with offboarding on User eligibility**. Same-Scope participation alone is insufficient.

Organization therefore distinguishes an ordinary in-Scope subject read from a protected read.

Conceptual public contracts:

```go
func (s *Service) SecuritySubjectIn(
    ctx context.Context,
    scope txscope.Scope,
    userID uuid.UUID,
) (SecuritySubject, error)

func (s *Service) ProtectedSecuritySubjectIn(
    ctx context.Context,
    scope txscope.Scope,
    userID uuid.UUID,
) (SecuritySubject, error)
```

`ProtectedSecuritySubjectIn` semantic guarantee:

```text
returns current User eligibility + GroupMembership facts
AND
serializes that User's eligibility against concurrent offboarding/eligibility-disable
until caller Scope completes
```

Exact PostgreSQL row lock/advisory/serialization mechanism remains T8-D.

Protected form is required at least where T3 already names the dependency across owners:

```text
new ApplicationSession issuance
governance ACCEPT / RETURN_FOR_CHANGES
Submission / withdraw / cancel / obsolescence actor mutations
other governed/security-changing owner mutations whose correctness requires the actor to remain enabled through commit
```

Same-owner Organization operations such as adding GroupMembership may establish required target eligibility serialization internally without a cross-owner contract. Authorization RoleAssignment creation consumes Organization target facts that carry the same protected eligibility guarantee when the subject is a User.

Responsible-owner eligibility remains the D4-specific protected target check.

---

## 7. M2 correction — owner-side optimistic-concurrency token

T6's strong ETag / If-Match laws represent owner concurrency meaning at the wire; T8-C freezes the owner-side token so T8-E does not invent concurrency authority.

Each affected owner exports an opaque current-version token conceptually:

```go
type VersionToken string
```

`VersionToken` is semantically opaque outside the owner. Exact generation/storage representation remains T8-D; exact HTTP ETag encoding remains T8-E.

Affected reads return current value + VersionToken. Corresponding replacement/mutation receives `expected VersionToken` and fails with producer-owned stale/precondition error when current differs.

At minimum this applies to the ratified T6 set:

```text
Company
UserProfile
User eligibility
ProviderSubjectBinding
Area metadata/lifecycle
Group metadata
DocumentType base
DocumentType governance
DocumentType eligible-template set
Document responsible owner
Document Template role
```

DRAFT remains governed by WorkingContent generation, its already-ratified separate OCC token.

Exact already-current repeats may be owner no-op and must not fabricate duplicate Audit evidence.

---

## 8. M3 correction — Authentication/provider protocol seam

T8-B requires platform identity-provider code to own raw provider protocol handling while Authentication owns MetalDocs anti-corruption meaning.

The consumer-owned `authentication.ProviderClient` therefore uses only stdlib/primitive protocol values so `platform/identityprovider` can satisfy it without importing Authentication.

Corrected conceptual surface:

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
    ) (issuer string, subject string, err error)

    SearchSubjects(
        ctx context.Context,
        query string,
        emit func(ref string, displayHint string) error,
    ) error

    ResolveSubjectRef(
        ctx context.Context,
        ref string,
    ) (issuer string, subject string, err error)
}
```

Laws:

```text
platform verifies/parses raw OIDC/directory protocol structure
verified issuer + subject cross the seam explicitly
raw provider JSON/claims are not required by Launch Authentication contract
provider roles/groups/permissions never cross as MetalDocs truth
provider subject ref is opaque product-facing selection handle only
Authentication owns ProviderSubjectBinding + ApplicationSession meaning
```

If a future named assurance consumer needs a verified claim beyond issuer+subject, add the smallest primitive/consumer-specific seam then; do not expose raw claim bags now.

---

## 9. T8C-D13 precision — foreign Organization facts

### Responsible owner

`ResponsibleOwnerEligibilityIn` preserves D4 exactly and its call is protected against target offboarding:

```text
resolved
existing User
same Company
ENABLED
+ serialization against concurrent target offboarding until Scope completes
```

It grants no permission.

### RoleAssignment target

`RoleAssignmentTargetFactsIn` returns bounded existence/company/scope facts and, when subject is a User, protected current enabled eligibility sufficient for T3 security-sensitive grant creation.

Authorization remains owner of the RoleAssignment decision/mutation.

### Group deletion

Application gathers same-Scope owner-authored dependency facts from Authorization + Controlled Documents and maps them into Organization's deletion input. Any unresolved source => deletion fails closed.

---

## 10. T8C-D14 / D15 / B4 correction — Managed Content admission claims

Managed Content remains mechanism, not Artifact/domain authority.

### 10.1 Serving/content mechanism port

Controlled Documents consumer-owned minimal store remains conceptually:

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

`OpenRange` stays outside Launch baseline until a real viewer/file-size consumer activates it.

### 10.2 Admission-claim mechanism contract

Knowing a handle UUID is never attachment authority. Controlled Documents therefore consumes an opaque technical claim contract conceptually:

```go
type AdmissionClaim interface { /* opaque/unforgeable */ }

type AdmissionClaims interface {
    Reserve(
        ctx context.Context,
        handle uuid.UUID,
    ) (AdmissionClaim, error)

    ProveLive(
        ctx context.Context,
        claim AdmissionClaim,
        handle uuid.UUID,
    ) error

    ConsumeIn(
        ctx context.Context,
        scope txscope.Scope,
        claim AdmissionClaim,
        handle uuid.UUID,
    ) error

    Release(
        ctx context.Context,
        claim AdmissionClaim,
    ) error
}
```

Laws:

```text
claim binds one handle to one in-flight authorized attachment intent without owner_type/owner_id registry
claim is unforgeable/opaque to the client
live claim prevents GC eligibility
ConsumeIn is atomic with successful semantic attachment in the caller Scope
rollback does not leave semantic attachment committed without the corresponding claim consumption
Release ends an abandoned claim
unconsumed/unreleased claim has bounded mechanism expiry
exact expiry duration/persistence schema = T8-D/T8-G
```

Cross-root/template copy creates a new handle + new claim.

---

## 11. B4 correction — T5-J managed-content GC contract path

T5-J remains technical reclamation and creates no retention/records domain.

T8-C freezes the required coordination contract, not storage schema or scheduler topology.

### 11.1 Managed-content GC mechanism side

A bounded GC mechanism port must support conceptually:

```text
claim bounded technical GC candidate(s)
mark/retain candidate in GC_PENDING technical state
prove no live admission claim
prove no backup pin/exclusion
re-prove GC_PENDING immediately before provider delete
DeleteReclaimable(handle)
finalize/remove technical mechanism state after confirmed absence
```

Exact SQL/table/lease representation is T8-D; periodic process/runtime placement is T8-G.

### 11.2 Controlled Documents semantic-reference check

Controlled Documents exports a bounded batch query for candidate handles:

```text
for each handle:
  current WorkingContent reference?
  immutable Submission reference?
  OfficialRendition reference?
  imported governed-content reference?
```

It returns only canonical reference-presence facts. It does not own GC decisions.

### 11.3 Coordination law

The T5-J application operation coordinates:

```text
technical GC candidate
-> ControlledDocs canonical reference check
-> mechanism live-claim + backup-pin check
-> if any protection exists: not deletable
-> technical GC_PENDING / immediate recheck
-> provider DeleteReclaimable outside semantic tx
-> finalize technical state
```

Safe failure remains leaked storage, never lost governed truth.

The exact existing T8-B application leaf hosting this non-semantic maintenance choreography is not redefined here; no new semantic owner or product route is created. If later runtime/package proof shows no existing T8-B application surface can legally host the T5-J invocation without violating the frozen dependency graph, that concrete contradiction reopens only the implicated T8-B leaf decision.

---

## 12. M4 correction — malware verdict binds exact bytes

Controlled Documents defines the consumer-owned technical port:

```go
type MalwareInspector interface {
    Inspect(
        ctx context.Context,
        content io.Reader,
    ) (digest [32]byte, clean bool, err error)
}
```

Laws:

```text
clean=false,nil -> malicious; governed admission forbidden
err!=nil -> unavailable/incomplete; fail closed/retry visibly
clean=true,nil -> only usable when returned digest equals ExactContentDescriptor.sha256 of the bytes being admitted
```

The scanner owns no Document/Submission meaning. No scan business lifecycle is created.

---

## 13. T8C-D16 / D17 — OfficialRendition execution + River binding

Renderer remains a consumer-owned technical port in the rendition application path.

Named durable-intent sink remains:

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

The platform River adapter obtains the concrete `*sql.Tx` only through the txscope-owned platform-only native binding from §3.3, then uses River's transactional insert primitive.

No River client/driver/transaction type crosses into application or owner signatures.

T5 still names only this activated semantic durable-intent family for Launch. No generic EventBus/outbox is added.

---

## 14. T8C-D19 adjudication — concurrent idempotency behavior, B5 rejected

Round-1 B5 assumed the losing concurrent request must encounter a unique-violation that aborts its caller transaction. That failure mode is not required by the frozen behavior.

Current PostgreSQL documentation establishes:

```text
INSERT ... ON CONFLICT DO NOTHING provides an alternative to raising unique violation
under READ COMMITTED an insert may not proceed because of another transaction's outcome even when that row was not visible in the command's original snapshot
successive commands in the same READ COMMITTED transaction may observe later commits
```

Therefore T8-C does **not** freeze savepoints or application-level retry merely to avoid an avoidable SQL realization.

### 14.1 Required BeginIn postcondition

The internal contract remains:

```go
BeginIn(...) -> exactly one of:
  claim
  completed replay
  semantic fingerprint conflict error
  technical error
```

Additional frozen concurrency law:

```text
concurrent same scoped key+fingerprint requests serialize
winner commit -> loser can return completed replay without leaving caller Scope unusable
winner rollback -> one waiting contender may become the new claim owner
same key + different fingerprint -> conflict; no business mutation
no public/durable IN_PROGRESS business state
```

T8-D chooses the PostgreSQL SQL/locking/upsert realization that satisfies this contract. A realization that aborts the transaction on an expected same-key race does not satisfy T8-C.

`FailReplay` remains unnecessary because an uncommitted claim belongs to the same transaction and disappears on rollback.

---

## 15. T8C-D20 / M5 new decision — PII-free replay by construction

This is a real T8-C decision, not a textual correction.

### 15.1 Alternatives

```text
A. persist replay payload containing erasable PII
   + add purge/redaction coordination with UserProfile erasure

B. keep durable ReplaySnapshot PII-free by construction
   + require T8-E replay-required success representations to be reconstructible from stable non-erasable semantic result facts
```

### 15.2 Selection

```text
SELECT B — PII-FREE REPLAY SNAPSHOT BY CONSTRUCTION
REJECT baseline replay-purge/redaction subsystem
```

Reason:

```text
T6 requires exact replay of the completed result but does not require erasable profile enrichment to be part of every POST creation result
T3 deliberately separates erasable UserProfile from stable User identity/history
storing erasable PII creates a concrete replay retention root
purge/redaction machinery adds cross-owner coupling and failure modes with no current consumer once snapshots are PII-free
```

### 15.3 ReplaySnapshot law

For every T6 durable-Idempotency-Key operation, application owns a versioned operation-local snapshot containing only stable result facts required for deterministic replay.

Forbidden in durable snapshot payload:

```text
erasable UserProfile fields such as display enrichment/name/email where classified as erasable profile data
free-form governance/cancellation/obsolescence reason text
provider claims/tokens/raw directory payload
request body/header copies
raw HTTP response bytes
full governed content
```

Stable semantic identifiers and non-erasable immutable result facts may be stored.

T8-E must design the exact success representation for replay-required POSTs so exact status/body replay can be deterministically reconstructed from the PII-free snapshot without querying current mutable PII state.

If a future promoted Product requirement requires a durable-idempotent success response to contain erasable PII that must be replayed exactly after lawful erasure, reopen this decision and select the smallest explicit purge/redaction design then.

---

## 16. Replay authorization remains unchanged

Completed replay:

```text
re-authenticate current request
re-check current permission/scope
re-check minimum current resource visibility required to disclose stored result
DO NOT re-run historical mutation
DO NOT re-run original lifecycle preconditions
only after disclosure authorization -> encode stored ReplaySnapshot through T8-E mapping
```

No generic replay ACL is created.

---

## 17. M6 correction — OfficialRendition exact-content read

T6 requires a distinct semantic rendition-content resource.

Controlled Documents therefore exposes a distinct rendition read family in addition to source reads:

```text
OfficialRenditionContent
```

Conceptual flow:

```text
application requests rendition identity
-> ControlledDocs returns target Document/Revision scope + rendition semantic facts/access facts
-> Authorization evaluates current READ_EFFECTIVE or READ_HISTORY path as appropriate
-> after ALLOW, ControlledDocs opens exact OfficialRendition bytes through its ManagedContentStore consumer port
```

The rendition is not treated as Release source and provider handle/key never becomes public identity.

Current EFFECTIVE official rendition uses effective-read semantics. Historical rendition access uses historical-read semantics. Governance decision content remains exact Submission source, not rendition content.

---

## 18. Read-projection correction

General law:

```text
Authorization.AuthorizedScopes prefilters where a scoped permission is granted
-> owning domain performs canonical filtered/paginated query
-> owner access facts/domain predicates where required
-> Authorization Decide/DecideMany for exact candidate actions
-> bounded Organization display references
-> application lens result
-> T8-E wire encoding
```

This avoids:

```text
unbounded enumerate-all-then-filter
filter-after-pagination incoherence
application re-deriving RoleAssignment semantics
foreign SQL
persistent duplicate read authority
```

### Library

```text
subject
-> Authorization.AuthorizedScopes(document.read_effective)
-> ControlledDocs.LibraryCandidates(scopes, canonical filters, owner cursor)
-> current effective-domain truth is enforced by ControlledDocs
-> bounded final authorization checks where material
-> Organization refs for display
```

Search remains canonical PostgreSQL query/view under Controlled Documents owner-private persistence; materialized Search remains OFF.

### Audit

```text
subject
-> Authorization.AuthorizedScopes(audit.read)
-> map to Audit.ReadVisibility
-> Audit.ListEvents applies historical attribution before pagination
```

---

## 19. LOW precision corrections

### L1 — RoleAssignment POST idempotency

`POST /role-assignments` is explicitly included in the general T6 durable idempotency flow. Its named cross-owner flow now includes `BeginIn`/`CompleteIn` and a PII-free ReplaySnapshot.

### L2 — DecideMany correspondence

Specified in §5.

### L3 — singleton reads

Controlled Documents public read census explicitly includes:

```text
ResponsibleOwner
TemplateRole
```

and both return owner-side VersionToken for T6 If-Match workflows.

Authentication `ProviderBinding` read likewise returns VersionToken.

### L4 — obsolescence completion host

No separate completion command is invented.

```text
NoHumanApproval obsolescence completion
  hosted by InitiateObsolescenceIn

human-governed final ACCEPT completion
  hosted by DecideGovernanceStepIn
```

Both return the mandatory owner-authored `obsolescence completed` Audit evidence when completion actually occurs.

### L5 — ProviderSubjectBinding disabled/replaced wording

No unratified standalone disable API is created.

```text
replacement
  replaces/deactivates prior current binding as one Authentication mutation and emits required binding evidence

offboarding
  preserves ProviderSubjectBinding where required for truthful historical correlation
```

If a future named product operation needs independent provider-binding disable, that is a new operation and requires the appropriate T6/T8-C reopen.

---

## 20. Corrected cross-owner critical flows

### Session issuance

```text
provider verification outside Scope
-> Scope
-> Authentication.ResolveProviderBindingIn
-> Organization.ProtectedSecuritySubjectIn(bound User)
-> Authentication.IssueSessionIn using resolved enabled fact
-> commit
```

### User create

```text
provider-directory resolve outside Scope
-> Scope
-> current actor protected subject + Authorization organization.manage
-> idempotency BeginIn
-> Organization.CreateUserIn
-> Authentication.BindProviderSubjectIn
-> required owner Evidence -> Audit.AppendIn
-> CompleteIn(PII-free ReplaySnapshot)
-> commit
```

### Offboarding

```text
Scope
-> current actor protected subject + Authorization organization.manage
-> serialize target eligibility
-> Organization disable User
-> Authentication revoke all ApplicationSessions
-> Organization remove current GroupMemberships
-> Authorization revoke direct User RoleAssignments
-> all owner-authored required evidence -> Audit.AppendIn
-> commit
```

### Group membership

Organization ensures target eligibility serialization internally for the same-owner User+GroupMembership mutation; current actor uses protected subject + `access.manage` decision.

### RoleAssignment create

```text
Scope
-> current actor protected subject + Authorization access.manage
-> idempotency BeginIn
-> Organization.RoleAssignmentTargetFactsIn (protected if User subject)
-> Authorization.CreateRoleAssignmentIn
-> Audit.AppendIn
-> CompleteIn(PII-free ReplaySnapshot)
-> commit
```

### Document create / next Revision / SUBMIT / governance / owner change

All governed actor-sensitive mutations use `ProtectedSecuritySubjectIn`. Specialized target eligibility checks remain protected where T3/D4 require serialization.

SUBMIT requiring OfficialRendition uses named intent sink; the platform River adapter uses txscope native SQL binding only inside platform.

---

## 21. Corrected operation coverage

The corrected candidate explicitly covers all previously identified Round-1 omissions:

```text
Audit list/read                         -> Audit.ListEvents + Authorization scopes
creation/options scope enumeration      -> Authorization.AuthorizedScopes
OfficialRendition content read          -> ControlledDocs OfficialRenditionContent
ManagedContent claim lifecycle          -> AdmissionClaims
DeleteReclaimable                       -> ManagedContentStore
T5-J GC coordination                    -> bounded mechanism + ControlledDocs reference query
If-Match owner version                  -> owner VersionToken + expected version
eligibility serialization               -> ProtectedSecuritySubjectIn / specialized protected target facts
provider issuer+subject protocol split  -> typed primitive coordinates
malware exact-byte correlation          -> returned digest
idempotent replay PII/erasure            -> PII-free snapshot selection
```

All original T6/T2/T3/T4/T5 operation families remain governed by the unchanged original candidate paths unless amended here.

---

## 22. Current-code reuse disposition after adjudication

```text
current db.Tx executor shape
  PRESERVE PROPERTY / REHOME+REFINE
  now with explicit database/sql-family selection + sealed Scope

current TxRunner lifecycle property
  PRESERVE PROPERTY

current runtime downcast db.Tx -> *sql.Tx for River
  REJECT / REWRITE
  target uses txscope-owned platform-only native binding

current IAM authz.Require
  REWRITE

current Audit export/application shape
  REWRITE public contract

current objectstore contract
  REWRITE

current idempotency HTTP middleware
  REWRITE

current idempotency persistence/concurrency ideas
  evidence only for T8-D; expected races must not abort target Scope

River v0.37.1 transactional InsertTx property
  PRESERVE / REFINE behind named sink

OpenAPI/codegen property
  PRESERVE; exact wire remains T8-E
```

Existence/tests do not create reuse entitlement.

---

## 23. Reference evidence relevant to the corrected delta

### River

Pinned repository evidence + current River docs confirm transactional `InsertTx` requires the driver's concrete transaction; for `riverdatabasesql` that is `*sql.Tx`.

Decision impact: explicit platform-only native SQL binding; no runtime type assertion invented by a Writer.

### PostgreSQL

Current PostgreSQL docs establish both:

```text
ON CONFLICT DO NOTHING is an alternative to unique-violation failure
READ COMMITTED commands may observe effects committed after earlier commands/snapshots
```

Decision impact: B5's assumed unavoidable transaction-abort path is rejected. T8-C freezes outcome semantics; T8-D chooses the SQL realization.

### Go

Go standard `database/sql` remains the selected transaction/executor primitive. No custom row/result framework is introduced merely to preserve unused provider choices.

External references remain evidence, never Product authority.

---

## 24. Stage-boundary check

### T8-C freezes

```text
contract ownership/direction
one local transaction participation contract
database/sql-family substrate decision + platform-only native binding law
Audit write/read contracts
Authorization decision + scope-enumeration contracts
protected eligibility serialization semantic guarantee
owner-side VersionToken/precondition contract
provider protocol consumer seam
ManagedContent + admission-claim + GC contract families
malware byte correlation
OfficialRendition execution/read contracts
named transaction-coupled intent
idempotency outcome/concurrency/replay representation laws
read-projection composition
fail-closed semantics
```

### T8-D still owns

```text
schema/table/index/constraint design
exact lock clauses/order
exact eligibility serialization SQL mechanism
Scope concrete wrapper/internal fields
actual database/sql driver/pool configuration
idempotency table/unique index/upsert statements
admission-claim storage/expiry persistence
GC_PENDING tables/leases
Audit table/query SQL
River table/schema realization
managed-content persistence
```

### T8-E still owns

```text
exact HTTP status/path/operationId/schema/header encoding
ETag quoting/format from VersionToken
Idempotency-Key wire grammar
exact ReplaySnapshot -> HTTP body/status encoder
OpenAPI generated types
```

No frontend/runtime/transition decision is promoted by this candidate.

---

## 25. Corrected decision set

Original T8C-D01→D25 remain selected except where the text below explicitly refines them.

```text
D01 SELECT authority-aligned hybrid ownership                                  UNCHANGED
D02 SELECT concrete semantic-owner Services                                    UNCHANGED
D03 SELECT consumer-owned interfaces only for real technical consumers         UNCHANGED

D04 SELECT database/sql-family sealed txscope + Runner.Within
    + platform-only native SQL binding
D05 SELECT no *sql.Tx/pgx.Tx in semantic-owner public signatures               UNCHANGED

D06 SELECT owner-local Audit evidence -> app mapping -> AppendIn
    + Authorization-scoped Audit.ListEvents

D07 SELECT Decide/DecideIn/DecideMany/DecideManyIn
    + AuthorizedScopes from same canonical Authorization authority
D08 SELECT Organization SecuritySubject
    + ProtectedSecuritySubjectIn when eligibility must serialize with offboarding
D09 SELECT closed ControlledDocs access-fact vocabulary                        UNCHANGED
D10 SELECT request-scoped EnabledGroupMembersResolver                           UNCHANGED
D11 SELECT empty GROUP snapshot remains empty/no fallback                       UNCHANGED

D12 SELECT consumer-owned provider seam with verified primitive issuer+subject
    + opaque ref/display search; no raw claim bag in Launch contract
D13 SELECT bounded Organization target/dependency facts
    + protected target eligibility where T3/D4 require serialization

D14 SELECT opaque preflight values + explicit AdmissionClaim lifecycle
D15 SELECT ManagedContentStore + AdmissionClaims + MalwareInspector
    + DeleteReclaimable + exact-byte malware digest correlation
D16 SELECT OfficialRenditionRenderer                                             UNCHANGED
D17 SELECT one named OfficialRenditionIntentSink + txscope platform native binding
D18 SELECT no generic EventBus/outbox                                             UNCHANGED

D19 SELECT BeginIn/CompleteIn same Scope, no FailReplay
    + concurrent same-key outcome must not leave Scope unusable
    + SQL realization deferred T8-D
D20 SELECT operation-local versioned PII-FREE ReplaySnapshot
D21 SELECT live disclosure recheck, no mutation/precondition re-execution       UNCHANGED

D22 SELECT Authorization scope prefilter + owner canonical pagination
    + owner facts + application composition + exact final decision
D23 SELECT producer-owned errors                                                 UNCHANGED
D24 SELECT external/provider execution outside semantic Scope                    UNCHANGED
D25 SELECT selective reuse only through T8-A gate                                REFINED for txscope/River

D26 SELECT owner-side VersionToken + expected-version mutation contract
D27 SELECT PII-free replay by construction; no Launch replay-purge subsystem
D28 SELECT ManagedContent admission-claim + T5-J GC contract family
D29 SELECT protected eligibility-read semantic guarantee; lock realization T8-D
D30 SELECT Audit historical-visibility read contract
D31 SELECT Authorization AuthorizedScopes query from canonical evaluator
```

---

## 26. Findings closed by this correction

```text
B1  CLOSED by explicit database/sql substrate + platform-only native binding
B2  CLOSED by Audit.ListEvents + resolved historical visibility
B3  CLOSED by Authorization.AuthorizedScopes + owner-side pagination prefilter
B4  CLOSED by AdmissionClaims + DeleteReclaimable + T5-J coordination contracts
B5  REJECTED AS BLOCKER; concurrent outcome law frozen, SQL strategy T8-D

M1  CLOSED by protected eligibility serialization contract
M2  CLOSED by owner VersionToken / expected-version law
M3  CLOSED with primitive verified issuer+subject seam
M4  CLOSED by inspected-byte digest correlation
M5  DECIDED: PII-free replay by construction
M6  CLOSED by OfficialRendition content read contract

L1-L5 CLOSED by explicit precision text
```

This status is candidate/adjudication status only. Round 2 must independently attack the material delta before any promotion.

---

## 27. Bounded Round-2 review contract

Round 2 is intentionally **not** a broad re-review of T8C-D01→D25. The Global Maximum class and all unchanged decisions were already independently confirmed.

Round 2 must attack only the surviving material delta:

```text
R2-1  txscope correction:
      deliberate database/sql-family selection
      sealed Scope
      platform-only native SQL binding
      River compatibility
      no owner/application native-binding escape
      no T8-D trespass beyond declared substrate constraint

R2-2  Audit read contract:
      historical attribution
      Authorization scope routing
      pagination before/after visibility

R2-3  Authorization AuthorizedScopes:
      same canonical evaluator
      no second policy engine
      pagination/list coherence

R2-4  ManagedContent claims + GC:
      T4-D/T4-F/T5-J completeness
      no Artifact/retention owner resurrection
      no hidden T8-D schema design
      legal application/runtime invocation path under T8-B

R2-5  protected eligibility serialization:
      semantic guarantee sufficient for T3 §11
      lock mechanism correctly remains T8-D

R2-6  owner VersionToken:
      concurrency authority remains owner
      wire ETag remains T8-E

R2-7  ProviderClient correction:
      typed enough to keep raw protocol in platform
      primitives avoid platform->Authentication import
      no provider claims leak

R2-8  malware exact-byte digest correlation

R2-9  idempotency B5 disagreement:
      verify PostgreSQL primary behavior
      attack whether BeginIn can meet concurrent winner/loser law without savepoint/retry contract
      distinguish contract outcome from SQL realization

R2-10 PII-free ReplaySnapshot decision:
      exact replay after profile erasure
      no current-state re-projection
      whether T8-E can encode exact response without erasable PII
      whether purge/redaction is actually avoidable

R2-11 OfficialRendition content read

R2-12 operation-census delta closure + L1-L5 precision
```

Round 2 must report whether any real contradiction survives. Another broad round is not justified unless this corrected delta changes the confirmed model class.

---

## 28. Gate

```text
T8-B  CLOSED / OPERATOR-RATIFIED / PROMOTED

T8-C  ACTIVE
      Global Maximum class CONFIRMED by Round 1
      Round-1 review COMPLETE
      Lead adjudication COMPLETE
      corrected candidate MATERIALIZED AS NON-AUTHORITATIVE STAGING
      bounded Fable Round 2 = NEXT

T8-D  NOT OPEN
T8-E  NOT OPEN
T8-F  NOT OPEN
T8-G  NOT OPEN
T8-H  NOT OPEN
T9→T12 NOT OPEN
implementation BLOCKED
```

Do not promote this candidate, open T8-D, or implement product code before bounded Round 2, final Lead adjudication and explicit operator ratification.

---

**End of adjudicated corrected candidate.**