# MetalDocs R10-B2-1 — Authentication Binding / Session / Assurance — Independent Fable Review Request

> **Status:** CANDIDATE / INDEPENDENT REVIEW REQUEST — **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Authority baseline HEAD:** `e1bd83ce9a0a9b70135e4bd6984d990a54ba6377`
> **Stage:** R10-B2-1 — Authentication Binding / ApplicationSession / Assurance
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** **CLOSED — design/documentation only.**
> **Authority note:** this artifact is review evidence only. It does not amend R9.5, R10-A, R10-B1, R10-B2, handoff, code, schema, OpenAPI, frontend, Keycloak configuration or deployment.

---

# 0. Cold reviewer bootstrap

Reconstruct repository state fresh. Do not use prior conversation memory as authority.

Read `AGENTS.md` and follow its complete read order / authority chain. At minimum read:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. `wiki/architecture/r10-technical-architecture.md`
7. the Single-Company Deployment/Tenancy rebaseline evidence chain only as needed to understand the current B1 substrate
8. this candidate packet
9. current auth/IAM/security code/schema/tests only as claim-specific evidence
10. official/primary external documentation only if a material Keycloak/OIDC claim needs verification

Current implementation is evidence, never target entitlement.

Apply the DevelopmentConexus Engineering Method proportionally:

```text
Evidence
→ Known / Inferred / Unknown / Deferred
→ Root Cause
→ Target Invariant
→ Constraints
→ Credible Alternatives
→ Local Maximum vs Global Maximum
→ Essential vs Accidental Complexity
→ YAGNI / Future Cost
→ Authority / Boundary
→ Enforcement
→ Proof Strategy
→ Adversarial Challenge
→ Decision
→ Reopen Triggers
```

Use the Structural Inversion Test. If the current local-credential auth module did not exist, which authentication/session properties would still be necessary under the frozen target?

---

# 1. Current authority constraints

Current promoted authority already fixes:

```text
one company per deployment V1
one singleton Tenant root per deployment
no universal tenant_id partition column
no Tenant RLS / tenant context / company switching
Keycloak = V1 Authentication provider
Authentication ≠ Organization ≠ Authorization
provider roles/groups/organizations/permissions are not canonical MetalDocs AuthZ
no provider claims map / claim→permission bridge
no cross-provider DB atomicity / XA / 2PC
B2-1 owns provider subject binding + app Session + assurance/fresh-auth
B2-2 owns User lifecycle/offboarding
B2-4 owns cross-owner concurrency/transaction closure
implementation = BLOCKED
```

The candidate must not reopen these without a real material contradiction.

---

# 2. Current-state evidence — useful properties vs legacy coupling

Current `internal/modules/auth` mixes multiple target owners/mechanisms:

```text
local password credentials / password lifecycle
lockout
User administration
Tenant context
roles/capabilities
Session
```

The current model also stores `TenantID`, roles/capabilities and user/profile data in auth-facing structures. Those shapes are legacy evidence, not target entitlement.

One current property is independently valuable and should be attacked as a **property**, not preserved by implementation inertia:

```text
opaque high-entropy session bearer
server-side session state
stored verifier/hash rather than replayable raw bearer
server-side revocation
```

Current tests specifically exercise token opacity, CSPRNG material, stored-hash-not-raw and post-logout revocation. The candidate retains those properties while discarding the local-credential/tenant/AuthZ coupling around them.

---

# 3. Root cause

The B2-1 problem is not "integrate Keycloak".

Root cause question:

> How does MetalDocs convert a provider-authenticated external subject into the correct MetalDocs `Organization.User` and a locally revocable authenticated application context, without allowing the provider to become Organization/AuthZ authority, without persisting provider credential/token machinery, and without depending on provider availability for every ordinary authenticated request?

The target must distinguish four meanings:

```text
issuer + subject
  = external/provider authentication identity

Organization.User
  = person/user identity inside this MetalDocs company deployment

ProviderSubjectBinding
  = MetalDocs authority over the correlation between the two

ApplicationSession
  = MetalDocs authority over an established authenticated application session
```

---

# 4. Credible alternatives

## A — provider token/JWT is the MetalDocs session

Browser/provider token is sent to MetalDocs on ordinary requests; MetalDocs validates/introspects provider state.

Potential benefit: fewer local session rows.

Material costs to attack:

```text
provider session lifecycle becomes application-session lifecycle
session.manage/local revocation becomes awkward or provider-dependent
provider outage can become request-path outage
claims-bearing tokens sit close to product AuthZ
refresh/access token custody enters MetalDocs
provider expiry semantics dictate application semantics
```

Candidate verdict: **REJECT** unless reviewer finds a stronger total-complexity argument.

## B — provider authenticates; MetalDocs issues its own opaque ApplicationSession

```text
Keycloak/provider auth
→ bounded trusted subject/assurance facts
→ ProviderSubjectBinding
→ Organization.User
→ MetalDocs ApplicationSession
```

Candidate verdict: **RECOMMENDED**.

## C — provider shadow directory / generic identity graph

ProviderIdentity/ProviderAccount/ProviderOrganization/GroupMirror/ClaimMirror/SyncState/etc.

Candidate verdict: **REJECT / YAGNI**. Provider lifecycle remains provider-owned; synchronization attempts are mechanism state in R10-D, not Authentication semantic authority.

---

# 5. Candidate persistent families — exactly two

The candidate target contains exactly two Authentication semantic-persistence families:

```text
ProviderSubjectBinding
ApplicationSession
```

Fresh-auth/assurance is bounded session state plus a transient evidence/value object; no third semantic `AuthenticationAssuranceEvent` table is proposed.

Provider provisioning/retry/reconciliation attempt state belongs to R10-D `DURABLE MECHANISM`, not B2-1 semantic authority.

---

# 6. ProviderSubjectBinding candidate

Conceptual relational target:

```text
ProviderSubjectBinding

id UUID PRIMARY KEY
user_id UUID NOT NULL
issuer TEXT NOT NULL
subject TEXT NOT NULL
state TEXT NOT NULL CHECK (state IN ('ACTIVE','DISABLED'))
created_at TIMESTAMPTZ NOT NULL
state_changed_at TIMESTAMPTZ NOT NULL

FOREIGN KEY (user_id)
  REFERENCES organization_user(id)
  ON DELETE RESTRICT

UNIQUE (issuer, subject)
```

Candidate active-cardinality rule:

```text
one Organization.User
→ at most one ACTIVE ProviderSubjectBinding V1
```

Conceptually enforce with a partial uniqueness rule such as:

```text
UNIQUE(user_id) WHERE state = 'ACTIVE'
```

The exact table/constraint syntax remains B2 implementation-spec detail; the semantic rule is under review here.

## 6.1 Binding mapping immutability

For one binding row, these mapping fields define its identity meaning:

```text
user_id
issuer
subject
```

They are not silently edited in place. Identity replacement uses a new binding row:

```text
old binding ACTIVE → DISABLED
new binding created ACTIVE
```

The mapping is immutable while the row exists; the lifecycle state may change.

## 6.2 Binding `ACTIVE|DISABLED` is MetalDocs acceptance, not Keycloak state mirror

`state` means whether MetalDocs accepts that binding for new ApplicationSession creation.

It does **not** claim to mirror provider account enabled/disabled/locked state. Provider credential/account status remains provider authority.

Do not add provider-shadow status fields merely to make reconciliation convenient.

## 6.3 User offboarding is not binding deletion/disable by default

Organization User offboarding/ineligibility is a separate fact owned by Organization.

A binding may remain a truthful correlation even when the User is offboarded. Offboarding must revoke ApplicationSessions and block new Session issuance; it does not automatically rewrite external identity history.

Binding disable/replacement is for identity-correlation trust changes such as:

```text
provider/realm migration
explicit security unlink
confirmed provider subject removal/invalidity
identity replacement
administrative disable of the correlation
```

## 6.4 No email/username/display-name auto-binding

Hard candidate invariant:

```text
email       != stable provider identity
username    != stable provider identity
displayName != stable provider identity
```

A successful provider authentication with an email matching a User MUST NOT create a binding automatically.

Trusted binding creation may result from an explicit provisioning/correlation flow, trusted invitation/account activation, controlled migration, or reconciliation of a known provisioning intent. Exact UI/provider mechanics belong R10-D/R10-E.

## 6.5 Provider-authenticated subject without binding

```text
provider authentication valid
+ no ACTIVE ProviderSubjectBinding
= no MetalDocs ApplicationSession
```

No JIT User creation, email auto-binding, role import or provider-organization import occurs from ordinary login.

---

# 7. ApplicationSession candidate

Conceptual relational target:

```text
ApplicationSession

id UUID PRIMARY KEY
subject_binding_id UUID NOT NULL
credential_digest BYTEA NOT NULL
created_at TIMESTAMPTZ NOT NULL
expires_at TIMESTAMPTZ NOT NULL
revoked_at TIMESTAMPTZ NULL

latest_reauthenticated_at TIMESTAMPTZ NULL
latest_provider_auth_time TIMESTAMPTZ NULL
latest_acr TEXT NULL
latest_amr <bounded representation> NULL

FOREIGN KEY (subject_binding_id)
  REFERENCES provider_subject_binding(id)
  ON DELETE RESTRICT

UNIQUE (credential_digest)
CHECK (expires_at > created_at)
```

The exact bounded representation for `amr` and whether optional assurance fields are separate columns or one bounded whole snapshot is reviewable; arbitrary provider claims JSON is forbidden.

## 7.1 Do not duplicate User/Tenant/AuthZ state in Session

Candidate Session contains no:

```text
tenant_id
user_id duplicate of binding.user_id
roles
permissions
groups
Area grants
provider roles/groups/organizations/permissions
provider access token
provider refresh token
provider ID token
email/username/display name
```

User is resolved through:

```text
Session → Binding → Organization.User
```

The one-company deployment already fixes company context; no Tenant dimension exists in the Session.

Role/grant/group changes therefore take effect through current canonical Authorization without session regeneration/revocation.

## 7.2 Opaque bearer + non-replayable DB disclosure

Browser/client receives an opaque high-entropy bearer credential.

The product DB stores only a one-way credential digest/verifier sufficient to resolve the Session. Disclosure of an ApplicationSession DB row must not yield a replayable browser credential.

The candidate intentionally does not freeze cookie encoding, HMAC framing, exact token byte count or delivery format. Those are R10-E/implementation mechanics.

## 7.3 Multiple Sessions per User

Multiple active Sessions for one User are allowed (e.g. workstation + notebook + mobile). Do not impose one-session-per-user without a real requirement.

`session.manage` may later expose one-session and all-user-session revocation semantics through R10-E without changing this authority.

## 7.4 Finite absolute lifetime

```text
expires_at IS NOT NULL
```

Infinite-until-logout Session is excluded. Exact duration is security/product configuration, not B2-1 authority.

Idle timeout / `LastSeenAt` is **not** a required B2-1 semantic property. Current runtime support is evidence only. Add later only for a concrete consumer/security policy.

## 7.5 Minimal Session lifecycle

No persisted state enum is required:

```text
revoked_at IS NULL && now < expires_at → ACTIVE
revoked_at IS NOT NULL                 → REVOKED
now >= expires_at                      → EXPIRED
```

`revoked_at` is terminal; reauthentication never revives a revoked/expired Session. A new login creates a new Session.

## 7.6 IP/User-Agent/LastSeen not frozen as semantic state

Current runtime records IP/User-Agent/LastSeen. Candidate does not promote them because no frozen property currently depends on them and they increase PII/mutation/write/privacy surface.

Telemetry or session-device UX may later introduce mechanism/support state if a real consumer appears.

---

# 8. Authentication assurance / fresh-auth

No third semantic table is proposed.

ApplicationSession stores only the latest bounded assurance needed by the product:

```text
latest_reauthenticated_at?
latest_provider_auth_time?
latest_acr?
latest_amr?
```

A published Authentication result may expose bounded value/evidence such as:

```text
FreshAuthEvidence
  session_id
  verified_at
  provider_auth_time?
  acr?
  amr?
```

Approval or another owner snapshots whatever frozen decision evidence it is required to preserve.

## 8.1 Initial login is not automatically explicit reauthentication

Candidate does **not** automatically set `latest_reauthenticated_at = created_at`.

`requires_reauthentication` means an explicit provider authentication challenge was completed for the operation/session context. A future policy may deliberately define an acceptable freshness window, but no hidden time window is frozen here.

## 8.2 Reauth must prove the same provider identity

A forced reauth callback for an existing Session must verify:

```text
reauth issuer  == Session.Binding.issuer
reauth subject == Session.Binding.subject
```

Mismatch fails closed. Reauthentication cannot change Session identity; authenticating another subject requires a new login/new Session.

## 8.3 Reauth cannot revive revoked/expired Session

The assurance update succeeds only if the Session is still active at the final local mutation. If revoked/expired while the provider redirect was in flight, callback fails and does not restore the Session.

## 8.4 Provider token material is not ApplicationSession authority

After the provider response is validated and bounded facts are extracted, Keycloak access/refresh/ID token material is not retained as the MetalDocs ApplicationSession source of truth unless a later concrete provider API consumer proves a need. Normal authenticated requests do not require provider introspection.

---

# 9. Provider availability / disable posture

## 9.1 Provider unavailable

```text
existing valid MetalDocs ApplicationSession → continues locally until revoke/expiry
new login                              → fails visibly
forced reauth                          → fails visibly
provider provisioning/reconciliation  → retries via R10-D
```

Keycloak/provider availability is therefore not a mandatory dependency for every ordinary request after an ApplicationSession is established.

## 9.2 Provider-only disable/removal while Session is live

Three alternatives must be attacked:

A. per-request provider introspection — candidate rejects due to availability/coupling cost;
B. mandatory back-channel logout V1 — candidate defers until a real immediate-provider-revocation requirement;
C. bounded staleness — **candidate recommendation**.

Candidate V1 semantics:

```text
provider-only disable/removal
→ new login/reauth fails when provider is consulted
→ existing MetalDocs Session may remain valid until local revoke/expiry
→ reconciliation may revoke earlier
```

This is an explicit bounded-staleness contract, not an implied immediate-revocation guarantee.

## 9.3 Authoritative MetalDocs offboarding

For company access, the authoritative offboarding path is Organization User lifecycle, not "disable only in Keycloak".

B2-4 must guarantee:

```text
Organization User becomes authentication-ineligible
→ all local ApplicationSessions revoked atomically/coherently
→ required Audit/durable intent recorded
→ provider disable/deprovision may execute asynchronously
```

Thus normal employee offboarding does not inherit provider-only bounded staleness.

---

# 10. Binding/session local mutation laws

## 10.1 Binding disable

When MetalDocs disables a trusted binding:

```text
BEGIN local MetalDocs transaction
  Binding ACTIVE → DISABLED
  revoke all Sessions referencing binding
  required Audit
COMMIT
```

A disabled binding cannot issue new Sessions.

## 10.2 Binding replacement

When a confirmed new provider subject replaces an old one:

```text
BEGIN local MetalDocs transaction
  old binding → DISABLED
  new binding → ACTIVE
  revoke Sessions on old binding
  required Audit
COMMIT
```

Do not perform the local replacement until the new provider subject is known/confirmed. Provider calls are outside this transaction.

## 10.3 Role/grant/group changes do not require Session revocation

Because Session carries no canonical AuthZ snapshot, Authorization/Organization changes take effect on subsequent canonical checks. Do not revoke Sessions merely to refresh permissions.

---

# 11. Provider provisioning / reconciliation failure matrix

B2-1 defines semantic outcomes; R10-D owns retry/attempt mechanics.

| Case | Candidate semantic result |
|---|---|
| User exists / provider subject absent | User remains; no binding; login impossible until trusted provisioning/correlation completes |
| provider subject exists / binding absent | no MetalDocs access; trusted reconciliation/provisioning may create binding |
| binding exists / provider subject removed or disabled | reconciliation may disable binding + revoke Sessions; provider-only detection is not synchronous request-path guarantee |
| duplicate `(issuer,subject)` attempted | database uniqueness rejects / fail closed |
| provider unavailable | existing local Sessions continue; new provider-dependent operations fail/retry honestly |
| provider response uncertain after create/update | do not fabricate binding or assume failure; reconcile provider truth before local semantic write |

No `provider_sync_status = PENDING/SYNCED/FAILED/ORPHANED` semantic lifecycle is introduced. Execution attempt/retry/error state is R10-D mechanism state.

---

# 12. Provider provisioning transaction pattern

When Organization creates a User that requires provider provisioning:

```text
BEGIN MetalDocs
  Organization.User
  required Audit
  durable provider-provisioning intent
COMMIT
```

R10-D later performs provider work and resolves stable provider identity.

Only after a confirmed stable `(issuer,subject)` exists:

```text
BEGIN MetalDocs
  ProviderSubjectBinding
  required Audit
COMMIT
```

No provider HTTP call participates in a MetalDocs DB transaction. No binding is created from an uncertain provider response.

---

# 13. Critical concurrency cases

These cases cross into B2-4, but B2-1 must declare the invariant that B2-4 must enforce.

## C1 — new Session issuance vs User offboarding

Forbidden outcome:

```text
login reads User eligible
→ concurrent offboarding completes
→ login commits a new surviving valid Session afterward
```

Required total-order outcome:

```text
login/session issuance first
→ Session exists
→ offboarding revokes it

OR

offboarding first
→ login sees ineligible User
→ no Session
```

Candidate enforcement direction: Session issuance and User offboarding serialize on the same User eligibility authority/row inside the local MetalDocs transaction. Exact lock/atomic-update mechanism belongs B2-4.

## C2 — reauth callback vs Session revoke/expiry

Final assurance update must recheck Session active state. Revocation/expiry wins; reauth cannot resurrect.

## C3 — binding replacement/disable vs Session issuance

A Session may be issued only against an ACTIVE binding whose state is validated in the same local transaction/serialization boundary. Binding disable/replacement revokes old-bound Sessions atomically.

---

# 14. Persistence classification / privacy

Candidate classification:

```text
ProviderSubjectBinding
  semantic class = SEMANTIC AUTHORITY
  mutation law   = immutable mapping + constrained ACTIVE→DISABLED lifecycle
  privacy        = erasable when lawful User/data-subject cleanup permits/requires it

ApplicationSession
  semantic class = SEMANTIC AUTHORITY
  mutation law   = constrained mutable lifecycle; revocation terminal; assurance monotonic only from validated provider auth
  privacy        = operational/auth state; erasable after expiry/revocation subject to required Audit evidence
```

Do not preserve provider subject, email, IP/User-Agent or session rows eternally merely to satisfy Audit. Audit/domain evidence owns its own PII-minimized immutable history.

If retained governed evidence needs an actor reference after personal enrichment is erased, use the B6 privacy-safe Audit/evidence design rather than converting Authentication rows into records-retention authorities.

---

# 15. Explicit non-target / deferred items

B2-1 does not design or promote:

```text
local password / password-policy / lockout / MFA / passkey tables
provider account/credential state mirror
Keycloak roles/groups/Organizations as product authority
provider claim map
claim→permission mapping
provider access/refresh token vault
per-request Keycloak introspection V1
mandatory back-channel logout V1
provider-sync semantic state machine
identity graph / generic principal registry
email-based JIT binding/User creation
Tenant/company dimension in binding/session
IP/User-Agent/LastSeen semantic Session state
idle-timeout semantic requirement
exact Session TTL value
exact cookie/token wire format
OIDC endpoints/state/nonce/PKCE/CSRF mechanics
Keycloak Admin API choreography
reconciliation cadence
login/provider logout UI
realm/config automation
```

Those belong to provider/R10-D/R10-E/implementation when a real consumer or security property requires them.

---

# 16. Structural Inversion Test

If the legacy system had always used Keycloak with no local passwords:

Would the target still need:

```text
Organization.User                         YES
stable provider subject correlation       YES
local application Session                 YES — local revocation/session.manage/availability boundary
fresh-auth assurance for Approval         YES
provider anti-corruption boundary         YES
```

Would it naturally invent:

```text
password tables                           NO
provider role/group mirrors               NO
claims-to-permission bridge               NO
provider token vault                      NO absent another provider-API consumer
provider sync business state machine      NO
per-request introspection                 NO absent immediate-provider-revocation requirement
```

The candidate therefore intends to preserve essential authentication/session properties while deleting legacy/provider accidental complexity.

---

# 17. Candidate decisions to attack

```text
B2-1-D1  exactly two persistent Authentication semantic families:
         ProviderSubjectBinding + ApplicationSession

B2-1-D2  stable external identity = issuer + subject

B2-1-D3  UNIQUE(issuer,subject) deployment-wide

B2-1-D4  max one ACTIVE ProviderSubjectBinding per User V1

B2-1-D5  binding mapping fields user_id/issuer/subject immutable in-place;
         replacement creates new row

B2-1-D6  no email/username/display-name auto-binding or ordinary-login JIT User

B2-1-D7  opaque local ApplicationSession with server-side revocation

B2-1-D8  DB stores non-replayable bearer verifier/digest, never raw session bearer

B2-1-D9  finite absolute Session expiry; multiple Sessions/User allowed

B2-1-D10 Session carries no Tenant or canonical AuthZ snapshot

B2-1-D11 provider access/refresh/ID tokens are not ApplicationSession authority

B2-1-D12 no required IP/User-Agent/LastSeen semantic fields V1

B2-1-D13 assurance stored only as bounded latest Session facts; no third semantic table

B2-1-D14 explicit forced reauth must authenticate same issuer+subject as Session binding

B2-1-D15 initial login does not automatically satisfy explicit `requires_reauthentication`

B2-1-D16 provider-only disable uses bounded staleness; no per-request introspection V1

B2-1-D17 authoritative MetalDocs User offboarding revokes local Sessions immediately/coherently

B2-1-D18 binding disable/replacement revokes bound Sessions in same local transaction

B2-1-D19 role/group/grant changes do not revoke Sessions merely to refresh AuthZ

B2-1-D20 uncertain provider outcome never fabricates local binding; R10-D reconciles first
```

---

# 18. Required proof/adversarial questions

The independent reviewer must attempt to falsify at least:

```text
P1  Can one issuer+subject become two Users despite DB/application races?
P2  Is one ACTIVE binding/User actually the smallest sustainable V1, or does Keycloak federation/migration require simultaneous bindings now?
P3  Does immutable mapping + new-row replacement preserve history without unnecessary permanent rows?
P4  Can email/username/provider claims accidentally acquire binding authority through another path?
P5  Can valid provider auth with no binding create a User/Session indirectly?
P6  Does ApplicationSession accidentally duplicate User/Tenant/AuthZ authority?
P7  Does DB disclosure of session state enable bearer replay?
P8  Is finite absolute expiry sufficient or is idle timeout essential now?
P9  Are IP/User-Agent/LastSeen truly nonessential, or does session.manage/audit/security require them?
P10 Can explicit reauth authenticate a different provider subject and mutate the existing Session identity?
P11 Can reauth revive a revoked/expired Session due to callback race?
P12 Does provider outage unnecessarily break established application sessions?
P13 Is bounded provider-disable staleness acceptable, or is back-channel/introspection a real V1 requirement?
P14 Does Organization offboarding revoke all local Sessions without provider dependency?
P15 Can new Session issuance race offboarding and survive incorrectly?
P16 Can binding disable/replacement race Session issuance and leave a valid Session on an untrusted binding?
P17 Does the provider failure matrix require a semantic provider-sync status family after all?
P18 Can an uncertain provider response create/fabricate a local binding?
P19 Is provider-token non-persistence compatible with all accepted V1 consumers, including reauth and logout?
P20 Does assurance need immutable event history in Authentication, or is session-latest + consumer snapshot sufficient?
P21 Are `acr/amr/auth_time` bounded facts enough, or is the candidate underspecifying assurance semantics?
P22 Does initial login need to count as fresh auth for any frozen use case?
P23 Does role/grant change require Session revocation for any accepted security property?
P24 Can Binding/Session privacy cleanup occur without erasing retained governed evidence or breaking Audit references?
P25 Does any candidate operation accidentally require atomicity across Keycloak/provider DB and MetalDocs DB?
P26 Did the candidate introduce another local maximum, generic IAM platform, shadow provider authority or duplicated identity authority?
```

Also perform a subtractive pass:

> What candidate field/table/lifecycle can still be deleted without weakening stable identity binding, local revocation, Approval fresh-auth, offboarding correctness, provider failure honesty or future federation/provider migration?

And a failure-mode pass:

> Which authentication failure can produce unauthorized access, false identity binding, stale access after offboarding, provider availability coupling, or fabricated provider truth under this candidate?

---

# 19. Review scope fences

Do not reopen by preference:

```text
single-company deployment substrate
singleton Tenant root
Keycloak selection
8+3 ownership topology
five roles / 43 permissions
Document/Revision/WorkingContent/Submission
Approval specialized workflow/SoD
Artifact/ManagedArtifactStore/malware gate
Dossier/Evidence
Retention/LegalHold/Disposition
Audit ownership
Interchange/Historical Migration
```

A material contradiction may reopen the minimum owning decision, but the review must show the counterexample rather than broadening scope automatically.

Do not design B2-2/B2-3 final schemas in this review. Cross-owner facts may be referenced only enough to validate B2-1 invariants/concurrency.

---

# 20. Required review output

Required verdict:

```text
APPROVE R10-B2-1 AUTHENTICATION BINDING / SESSION / ASSURANCE

or

APPROVE R10-B2-1 AUTHENTICATION BINDING / SESSION / ASSURANCE
WITH MATERIAL FIXES

or

DO NOT APPROVE R10-B2-1 AUTHENTICATION BINDING / SESSION / ASSURANCE
```

Required output must include:

```text
verdict
BLOCKER / MAJOR / LOW counts
B2-1-D1..D20 disposition individually or by clearly lossless groups
strongest alternative to local ApplicationSession and why it wins/loses
one-active-binding/User verdict
binding state/lifecycle verdict
Session minimal-field verdict
assurance/fresh-auth verdict
provider-disable/live-session verdict
offboarding/session-race verdict
provider provisioning/reconciliation verdict
privacy verdict
subtractive/YAGNI findings
any new authority overlap or material reopen
exact corrected target if fixes are required
whether another broad review is required
whether a bounded delta review would be sufficient after correction
```

If approved with fixes, distinguish:

```text
review finding = evidence
operator adjudication = next authority gate
```

No review finding itself amends target authority.

---

# 21. Write authorization for the independent reviewer

The operator authorization represented by this packet is only to:

1. create one independent review artifact;
2. commit it on the current branch;
3. push the same branch.

Suggested review artifact:

`docs/superpowers/analysis/2026-08-17-r10-b2-1-authentication-binding-session-assurance-independent-fable-review.md`

Do **not**:

```text
alter R10 authority
alter handoff/program/ledger authority
alter this candidate packet
continue into B2-2/B2-3/B2-4
change code/schema/OpenAPI/frontend
configure Keycloak/deployment
implement
merge PR
```

Implementation remains blocked.