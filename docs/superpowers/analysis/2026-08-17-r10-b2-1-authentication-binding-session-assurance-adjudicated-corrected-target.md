# MetalDocs R10-B2-1 — Authentication Binding / Session / Assurance — Adjudicated Corrected Target

> **Status:** ADJUDICATED CORRECTED CANDIDATE — **PENDING BOUNDED DELTA REVIEW — NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Authority baseline:** `e1bd83ce9a0a9b70135e4bd6984d990a54ba6377`
> **Candidate packet:** `docs/superpowers/analysis/2026-08-17-r10-b2-1-authentication-binding-session-assurance-fable-review-request.md` @ `9cba3acd`
> **Independent review:** `docs/superpowers/analysis/2026-08-17-r10-b2-1-authentication-binding-session-assurance-independent-fable-review.md` @ `361f6c8b`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** **CLOSED — design/documentation only.**
> **Authority note:** this file records operator adjudication/corrected target only. It does not amend R10 authority, handoff, ledger, program authority, code, schema, OpenAPI, frontend or Keycloak configuration.

---

## 1. Independent review result

```text
VERDICT = APPROVE R10-B2-1 AUTHENTICATION BINDING / SESSION / ASSURANCE
          WITH MATERIAL FIXES

BLOCKER = 0
MAJOR   = 3
LOW     = 5
```

The independent review confirmed that the B2-1 topology is sound:

```text
semantic persistent families = exactly two
  ProviderSubjectBinding
  ApplicationSession

local opaque ApplicationSession = KEEP
one enabled binding per User V1 = KEEP
assurance as bounded Session state + transient evidence = KEEP
bounded provider-disable staleness = KEEP
MetalDocs-authoritative offboarding = KEEP
no provider token vault = KEEP
no provider-role/group/org/claim authority = KEEP
no cross-provider DB atomicity = KEEP
```

The review found three material defects. The operator adjudicates them below; findings are evidence, not authority.

---

## 2. Operator adjudication summary

| Finding | Decision | Corrected target |
|---|---|---|
| F1 — total subject uniqueness + terminal disable creates relink dead-end | **ACCEPT defect / MODIFY reviewer fix** | Keep total `UNIQUE(issuer,subject)` because provider subject identity must never normally migrate between MetalDocs Users. Remove ACTIVE/DISABLED enum and use reversible `disabled_at`; re-enable same row for same User. Keep `UNIQUE(user_id) WHERE disabled_at IS NULL`. Handover of same `(issuer,subject)` to a different User is a bounded reopen, not V1 entitlement. |
| F2 — reconciliation can silently select subject by email/attribute | **ACCEPT** | Subject selection must have causal/verifiable correlation to the exact provisioning/correlation intent or an explicit trusted-human correlation decision. Email/username/display name never select subject, including provider “already exists” conflicts. |
| F3 — persisted `latest_*` assurance can be misread as indefinitely fresh | **ACCEPT** | Persisted assurance fields are evidence inputs only. Their presence never satisfies `requires_reauthentication`. Satisfaction must be explicitly bounded to an operation or to a deliberately configured freshness policy owned by the consuming policy authority. |
| L1 — binding state enum is removable | **ACCEPT** | Replace `state + state_changed_at` with `disabled_at NULL/NOT NULL`. |
| L2 — session-management UX may need IP/UA/last-seen support state | **ACCEPT / DEFER** | Keep out of semantic ApplicationSession. R10-E may add bounded support/mechanism telemetry if a concrete UX/security consumer requires it. |
| L3 — Session TTL is also provider-disable staleness bound | **ACCEPT / ROUTE** | B2-1 requires finite absolute TTL; R10-E/deployment security configuration must choose the value knowing it bounds provider-only-disable exposure absent earlier reconciliation. |
| L4 — provider logout may need transient ID token hint | **ACCEPT / DEFER** | Provider tokens remain outside ApplicationSession authority. R10-E may introduce request-scoped/short-lived mechanism state if current Keycloak logout semantics require it. |
| L5 — login/offboarding row-lock strategy | **ACCEPT AS B2-4 CANDIDATE** | B2-1 retains the concurrency invariant; B2-4 chooses exact locking/transaction enforcement. READ COMMITTED row-lock serialization is a credible minimal realization. |

---

# 3. B2-1 root cause and target invariant

Root cause:

> MetalDocs needs to bind externally authenticated Keycloak subjects to MetalDocs organizational Users and issue independently revocable product Sessions without moving credential authority, Organization membership, or Authorization into Keycloak.

Target invariant:

> A valid MetalDocs ApplicationSession proves that one currently accepted provider subject binding maps to one authentication-eligible MetalDocs User. Provider credentials and provider session/token mechanics remain external; MetalDocs owns only the binding, local application-session lifecycle/revocation and bounded authentication-assurance evidence.

The design must preserve the accepted single-company substrate:

```text
one company per deployment
no universal tenant/company partition column
no Tenant RLS/customer routing
Keycloak = V1 AuthN provider
Authentication != Organization != Authorization
no provider roles/groups/orgs/claims as canonical AuthZ
no provider DB atomicity / XA / 2PC
```

---

# 4. Exactly two semantic persistent families

B2-1 introduces exactly:

```text
ProviderSubjectBinding
ApplicationSession
```

No third semantic family is introduced for:

```text
provider sync state
provider account mirror
provider claims
fresh-auth history
authentication assurance events
provider tokens
provider Organizations/groups/roles
```

Provider provisioning/retry/attempt/error state belongs to R10-D durable mechanism state. Approval snapshots the per-decision authentication evidence it owns; Authentication does not create a competing attestation history.

---

# 5. ProviderSubjectBinding — corrected target

Conceptual target:

```text
ProviderSubjectBinding

id          UUID PRIMARY KEY
user_id     UUID NOT NULL REFERENCES Organization.User(id)
issuer      TEXT NOT NULL
subject     TEXT NOT NULL
created_at  TIMESTAMPTZ NOT NULL
disabled_at TIMESTAMPTZ NULL

UNIQUE (issuer, subject)
UNIQUE (user_id) WHERE disabled_at IS NULL
```

Cross-owner FK action follows B1:

```text
ON DELETE RESTRICT / NO ACTION
ON UPDATE RESTRICT / NO ACTION
```

## 5.1 Stable identity

Stable external identity is:

```text
(issuer, subject)
```

Never:

```text
email
username
display name
provider role/group/org
```

`user_id`, `issuer`, and `subject` are immutable mapping fields. They are never edited in-place to make the same row represent a different correlation.

## 5.2 Acceptance lifecycle

There is no `ACTIVE|DISABLED` state enum.

```text
disabled_at IS NULL     => MetalDocs currently accepts this binding
disabled_at IS NOT NULL => MetalDocs does not currently accept this binding
```

The acceptance flag is reversible for the **same mapping**:

```text
same issuer + subject
same User

accepted → disabled → accepted again
```

Re-enabling clears `disabled_at` on the same row. Audit records the disable/re-enable operations.

## 5.3 One enabled binding per User V1

```text
UNIQUE(user_id) WHERE disabled_at IS NULL
```

V1 therefore permits historical disabled bindings for a User while allowing at most one currently accepted provider subject.

This is sufficient because upstream federation occurs inside Keycloak and normally remains one MetalDocs-facing Keycloak issuer+subject. A real requirement for simultaneous overlapping MetalDocs-facing issuers is the reopen trigger.

## 5.4 Total external-subject uniqueness remains

The operator deliberately does **not** adopt the reviewer's ACTIVE-scoped `(issuer,subject)` uniqueness.

Target:

```text
UNIQUE(issuer, subject)
```

Rationale:

- accepted authority already treats issuer+subject as stable provider identity;
- V1 has no consumer requiring the same provider subject to become a different MetalDocs User;
- allowing historical handover makes actor/correlation meaning more ambiguous;
- same-subject re-trust for the same User is already supported by reversible `disabled_at`;
- User identity merge/provider subject reuse/handover is a materially exceptional capability and must be reopened explicitly if evidenced.

Forbidden V1 normal path:

```text
issuer X + subject Y → User A
later
issuer X + subject Y → User B
```

Reopen trigger: a proved provider subject-reuse guarantee, User merge/identity-consolidation requirement, or migration case that genuinely requires subject handover between MetalDocs Users.

## 5.5 Offboarding does not rewrite correlation truth

Organization User offboarding does not automatically disable the binding. User eligibility and provider-subject correlation are separate authorities.

A User may become authentication-ineligible while the binding remains the truthful external identity correlation. Offboarding immediately revokes local ApplicationSessions through the cross-owner flow defined by B2-4.

Binding disable is reserved for authentication-correlation decisions such as:

```text
security unlink
provider/realm migration
confirmed subject retirement/removal
explicit admin disable
identity replacement to a different subject
```

---

# 6. Binding-creation authority — F2 closure

A binding may never be created by human identity attribute matching.

Binding subject selection must have a **causal and verifiable correlation** to the specific provisioning/correlation intent or explicit trusted-human decision.

Permitted subject authority:

1. the provider operation executed for this exact provisioning/correlation intent creates and returns the subject;
2. reconciliation proves the subject through a unique provider-side correlation/idempotency mechanism belonging to that exact intent;
3. an explicit trusted human correlation decision designates the provider subject.

Never subject-selection authority:

```text
matching email
matching username
matching display name
similar name
provider "already exists" plus matching email
```

Those attributes may be displayed as corroborating information only.

Provider conflict example:

```text
provision intended User
→ provider: email/account already exists
```

Target result:

```text
NO automatic binding
→ pending explicit correlation/reconciliation outcome
```

The mechanics of how R10-D represents the pending intent/attempt remain mechanism design, not Authentication semantic state.

Valid Keycloak authentication with no accepted binding fails closed:

```text
provider authentication succeeds
+ no enabled ProviderSubjectBinding
= no MetalDocs ApplicationSession
```

No JIT User, no JIT binding, no role import.

---

# 7. ApplicationSession — corrected target

Conceptual target:

```text
ApplicationSession

id                          UUID PRIMARY KEY
subject_binding_id          UUID NOT NULL REFERENCES ProviderSubjectBinding(id)
credential_digest           BYTEA NOT NULL
created_at                  TIMESTAMPTZ NOT NULL
expires_at                  TIMESTAMPTZ NOT NULL
revoked_at                  TIMESTAMPTZ NULL
latest_reauthenticated_at   TIMESTAMPTZ NULL
latest_provider_auth_time   TIMESTAMPTZ NULL
latest_acr                  <bounded nullable representation>
latest_amr                  <bounded nullable representation>

UNIQUE (credential_digest)
CHECK (expires_at > created_at)
```

Exact wire encoding, cookie framing, hash/verifier algorithm and TTL value are implementation/R10-E decisions; the semantic properties are fixed here.

## 7.1 Local opaque Session is the chosen Global Maximum

Keycloak authenticates; MetalDocs creates its own opaque server-revocable ApplicationSession.

Rejected strongest alternative:

```text
provider JWT as bearer
+ local signature validation
+ local revocation list
```

That alternative removes one Session table only to re-create local revocation state, couples product-session lifetime to provider token mechanics and moves provider claims closer to product authority. No evidenced scale problem justifies the stateless trade.

## 7.2 Session references Binding, not duplicated User/Tenant

Session stores `subject_binding_id` only.

```text
Session → Binding → Organization.User
```

No duplicated `user_id` consumer exists in B2-1, and no `tenant_id` exists under the single-company substrate.

## 7.3 Multiple Sessions per User

Multiple concurrent Sessions are permitted. No one-session-per-User constraint exists.

`session.manage` may later support one-session or all-session revocation through the canonical Session/Binding/User relationships.

## 7.4 Bearer secrecy

The browser receives an opaque high-entropy bearer. The database stores only a one-way verifier/digest.

Invariant:

> Disclosure of an ApplicationSession database row must not yield a replayable browser credential.

Current runtime tests prove this property is already practical evidence; exact current token framing is not target authority.

## 7.5 Finite absolute lifetime

Every Session has finite absolute expiry:

```text
expires_at NOT NULL
expires_at > created_at
```

No infinite-until-logout Session. Idle timeout is not a B2-1 requirement because no accepted product requirement needs it. It may be added later as support/mechanism state if a concrete security/UX need appears.

The configured absolute TTL is also the maximum provider-only-disable staleness bound absent earlier reconciliation/local revocation; R10-E/deployment security configuration must choose its value with that consequence explicit.

## 7.6 Derived lifecycle

No persisted Session state enum:

```text
revoked_at IS NULL AND now < expires_at => active
revoked_at IS NOT NULL                 => revoked
now >= expires_at                      => expired
```

Revocation is terminal. Reauthentication never revives a revoked or expired Session.

## 7.7 Explicitly excluded semantic fields

ApplicationSession contains no canonical:

```text
Tenant/company id
User id duplicate
roles
permissions
groups
Area grants
Keycloak realm/client roles
provider claims map
access token
refresh token
ID token
email
username
IP address
User-Agent
LastSeenAt
```

IP/User-Agent/LastSeen may later exist as bounded support/mechanism telemetry for a concrete `session.manage` UX; they are not Authentication semantic authority.

Provider tokens may transiently exist for a concrete OIDC operation. If R10-E proves RP-initiated logout needs an ID-token hint, that is request-scoped/short-lived mechanism state, never normal ApplicationSession authority.

---

# 8. Authorization boundary

ApplicationSession never snapshots canonical Authorization:

```text
no roles
no permissions
no Groups
no Area assignments
no provider claims
```

Authenticated application context contains only identity/session/authentication facts such as:

```text
user_id
session_id
bounded authentication assurance
```

Authorization evaluates current Permission/Role/RoleAssignment/domain relationship state when an operation is attempted.

Therefore ordinary Role/Group/grant mutations do **not** require Session revocation. Their effect is visible on the next canonical authorization evaluation.

Paired invariant:

> Role/grant changes may avoid Session revocation only because ApplicationSession contains no authoritative Authorization snapshot.

---

# 9. Fresh-auth / assurance — F3 closure

## 9.1 Bounded Session assurance state

Authentication may persist the latest validated provider-assurance inputs:

```text
latest_reauthenticated_at
latest_provider_auth_time
latest_acr
latest_amr
```

They are **evidence inputs only**.

Their presence, including non-NULL `latest_reauthenticated_at`, never by itself satisfies a `requires_reauthentication` requirement.

## 9.2 Explicit reauthentication meaning

`requires_reauthentication` means an explicit provider authentication challenge was completed for a bounded operation context.

Initial login does not automatically satisfy a later `requires_reauthentication` operation.

A reauthentication callback must prove:

```text
callback issuer+subject == Session Binding issuer+subject
Session still active at final local update
```

Different subject → fail closed. Revoked/expired Session → fail closed. Reauthentication never switches identity and never revives Session.

## 9.3 Satisfaction contract

A consumer may accept Authentication assurance only under an explicitly bounded rule:

```text
A. one-shot operation-linked consumption
or
B. deliberately configured freshness window owned by the consuming policy authority
```

For Approval, B4/Approval owns the policy choice for Approval steps.

Forbidden:

```text
latest_reauthenticated_at IS NOT NULL
→ automatically satisfied forever
```

No implicit/unbounded freshness window exists.

## 9.4 FreshAuthEvidence

Authentication may publish a transient/value-object evidence result such as:

```text
FreshAuthEvidence
  session_id
  verified_at
  provider_auth_time?
  acr?
  amr?
```

Approval or another consumer snapshots whatever evidence its own authoritative record requires. Authentication does not create a third semantic assurance-event table.

Nonce/state/challenge persistence needed to bridge redirects is mechanism state owned by the later authentication/API journey design.

---

# 10. Provider availability / disable posture

## 10.1 Provider outage

Existing valid ApplicationSessions continue until local revocation or finite expiry.

```text
existing Session        => continues
new login               => fails visibly
explicit reauth         => fails visibly
provider provisioning   => R10-D retry/reconciliation
```

No per-request Keycloak introspection is required.

## 10.2 Provider-only disable

Provider-only disable/removal does not promise synchronous local Session revocation.

```text
provider-only disable
→ new login/reauth fails when provider consulted
→ existing local Session may remain until local revoke/expiry
→ reconciliation may revoke earlier
```

This bounded staleness is deliberate and bounded by finite local Session TTL.

Back-channel logout is not required V1; a future implementation may add it only as an earlier local-revocation trigger without transferring Session authority to Keycloak.

## 10.3 MetalDocs offboarding

Organization User offboarding is the authoritative MetalDocs access-removal flow:

```text
mark User authentication-ineligible
+ revoke all User ApplicationSessions
+ required Audit
```

These local effects must be one coherent MetalDocs transaction as finalized in B2-4. Provider disable/provisioning is an external effect and never part of that local atomicity claim.

---

# 11. Binding disable/replacement

Disabling a binding means MetalDocs no longer accepts that correlation for new authentication.

Binding disable must locally revoke all Sessions referencing the binding in the same MetalDocs transaction.

Replacement to a different provider subject occurs only after the replacement subject has been causally/explicitly confirmed under §6:

```text
BEGIN local MetalDocs transaction
  disable old binding
  enable/create new binding for same User
  revoke Sessions referencing old binding
  required Audit
COMMIT
```

Provider calls do not occur inside this transaction.

If replacement provider outcome is uncertain, the local replacement transaction does not execute yet.

Re-enabling the same disabled binding for the same User is allowed and does not mutate mapping fields.

---

# 12. Provider provisioning / reconciliation

The six required semantic cases remain:

```text
1. User exists / provider subject absent
2. provider subject exists / binding absent
3. binding exists / provider subject removed or disabled
4. duplicate issuer+subject attempt
5. provider unavailable
6. uncertain provider response / retry
```

Target outcomes:

### 12.1 User exists / subject absent

User remains Organization truth. No binding means no provider login access. If provisioning is required, the creating mutation may atomically write a durable provider-provisioning intent; the provider effect happens later.

### 12.2 Subject exists / binding absent

No access until a trusted correlation satisfying §6 creates the binding. No email/username JIT adoption.

### 12.3 Binding exists / subject removed or disabled

Reconciliation may disable the binding and revoke its Sessions locally. Provider-only disable may otherwise exhibit the bounded staleness in §10.2.

### 12.4 Duplicate subject attempt

Total `UNIQUE(issuer,subject)` fails closed. The existing binding remains the unique correlation authority.

### 12.5 Provider unavailable

Existing local Sessions continue. New provider-dependent work fails visibly/retries as appropriate. No fabricated local provider truth.

### 12.6 Uncertain response

An uncertain provider outcome never creates a binding by assumption. R10-D reconciles the exact intent; subject selection must satisfy §6.

No semantic `provider_sync_status` table/family is added. Retry/error/attempt/lease state is R10-D durable mechanism state.

---

# 13. Local transaction / concurrency invariants

B2-1 declares the correctness outcomes; B2-4 chooses exact enforcement.

## C1 — login Session issuance vs User offboarding

Forbidden outcome:

```text
offboarding commits
AND
a new valid Session created from pre-offboarding eligibility survives
```

Required total-order behavior:

```text
login first
→ Session may be created
→ later offboarding must revoke it

OR

offboarding first
→ later login observes ineligible User
→ no Session
```

A minimal credible READ COMMITTED realization exists via row-lock serialization on the same User eligibility row; exact lock/query shape remains B2-4.

## C2 — reauth callback vs Session revoke/expiry

Final assurance update succeeds only if the Session remains active at the local update point. Reauth callback cannot revive or mutate a revoked/expired Session.

## C3 — binding disable/replacement vs Session issuance

Session issuance must validate/serialize against the same accepted binding state that disable/replacement locks. A binding disable/replacement cannot commit while leaving a concurrent newly-issued Session on the disabled binding valid afterward.

The same local transaction that disables/replaces the binding revokes Sessions on the old binding.

## C4 — Role/grant mutation vs Session

Role/grant changes do not revoke Session because Session stores no AuthZ authority. Canonical current authorization applies immediately at the next protected operation.

---

# 14. Persistence class / mutation law

```text
ProviderSubjectBinding
  semantic class = SEMANTIC AUTHORITY
  mapping identity fields = IMMUTABLE
  disabled_at = constrained reversible acceptance fact
  erasable under lawful user/data-subject cleanup when no longer required

ApplicationSession
  semantic class = SEMANTIC AUTHORITY
  lifecycle = constrained active/revoked/expired
  revoked_at = terminal revocation marker
  assurance fields = mutable only from validated authentication events
  erasable after operational/security retention no longer requires it
```

Authentication rows are not Records Governance retention subjects by default. Durable governed evidence lives in owning domain/Audit records and must not depend on retaining provider binding or Session rows forever.

---

# 15. Privacy

Provider subject identity and Session state may contain personal/security-sensitive data.

V1 requires that lawful User/data-subject cleanup can remove Authentication rows/enrichment without rewriting retained governed business evidence.

Erasure order respects B1 `RESTRICT/NO ACTION`:

```text
ApplicationSessions first
→ ProviderSubjectBinding
```

Audit's allowed surviving skeleton must remain PII-minimized/non-PII and cannot require live FK-dependent Authentication rows. Human-readable enrichment is separately erasable/read-derived per B6.

No generic privacy workflow/domain is introduced.

---

# 16. Structural anti-corruption proof

No B2-1 canonical model contains:

```text
provider role
realm role
client role
provider Group
provider Organization
provider Permission
realm_access
resource_access
generic claims map
claim-to-permission mapping
provider-token authority
provider-account shadow lifecycle
```

The provider-facing result may expose only enumerated identity/assurance facts accepted by frozen authority, e.g.:

```text
issuer
subject
authenticated_at/auth_time
acr?
amr?
```

MetalDocs Organization/User/Group/RoleAssignment/Permission remain product authorities.

---

# 17. No cross-provider transaction

Every provider interaction is ordered outside local MetalDocs atomicity:

```text
login/provider validation
→ local Session transaction

reauth/provider validation
→ local assurance update transaction

provider provisioning intent local commit
→ provider effect/reconciliation
→ later local binding commit

binding replacement
→ provider subject confirmed first
→ local replacement/revoke commit
```

No invariant requires Keycloak DB + MetalDocs DB atomic commit.

---

# 18. Explicit exclusions / YAGNI

B2-1 does not introduce:

```text
local password/hash/policy/lockout/MFA/passkey state
provider account mirror
provider sync semantic FSM
AuthenticationAssuranceEvent history
provider token vault
per-request provider introspection
mandatory back-channel logout
idle Session state
IP/User-Agent/LastSeen semantic authority
one Session per User
multiple simultaneous active provider bindings/User
provider subject handover between Users
email/username attribute binding
JWT/provider-role Authorization snapshot
generic IAM/identity graph
XA / 2PC
```

Later seams:

- R10-D: provider intent/attempt/retry/reconciliation mechanism state;
- R10-E: OIDC login/logout/callback/nonce/cookie mechanics, TTL value, optional session-list telemetry;
- B2-4: exact row locking/transaction mechanics for concurrency invariants;
- B4: consumer policy for one-shot vs explicit bounded reauth freshness.

---

# 19. Proof obligations for bounded delta / promotion

A bounded delta reviewer must verify at minimum:

```text
P1  exactly two semantic persistent Authentication families remain
P2  issuer+subject is stable external identity
P3  total UNIQUE(issuer,subject) plus reversible disabled_at has no internal contradiction
P4  max one enabled binding/User holds under concurrency
P5  same binding can disable/re-enable without changing mapping fields
P6  same issuer+subject cannot silently move to another User
P7  no email/username/display-name path selects binding subject
P8  provisioning conflict/uncertainty cannot silently adopt provider account
P9  valid provider auth without accepted binding creates no Session/User
P10 Session contains no Tenant/AuthZ snapshot or provider token authority
P11 Session→Binding→User is sufficient without duplicated user_id
P12 stored Session row cannot reveal replayable bearer
P13 every Session has finite absolute expiry
P14 multiple Sessions/User are permitted
P15 fresh-auth fields are evidence inputs only; non-NULL is never satisfaction
P16 initial login does not automatically satisfy requires_reauthentication
P17 explicit reauth proves same issuer+subject and active Session
P18 consumer freshness is explicit bounded policy, never implicit/unbounded
P19 provider outage preserves existing local Sessions but blocks provider-dependent new work
P20 provider-only disable has honest bounded staleness
P21 MetalDocs offboarding revokes local Sessions independent of provider availability
P22 login/offboarding race has a reachable minimal serializing enforcement
P23 reauth/revoke race cannot resurrect Session
P24 binding disable/replacement/session-issuance race cannot leave surviving invalid Session
P25 role/grant changes are safe without Session revoke because no AuthZ snapshot exists
P26 six reconciliation cases remain complete
P27 no provider_sync semantic family is needed
P28 uncertain provider result never fabricates binding
P29 privacy cleanup can erase Authentication rows without destroying governed evidence
P30 no hidden Keycloak↔MetalDocs atomicity requirement exists
P31 no third field/table/lifecycle can still be deleted without losing a named property
```

---

# 20. Reopen triggers

Reopen only the smallest invalidated B2-1 decision on material evidence such as:

- simultaneous overlapping MetalDocs-facing provider identities are required for one User;
- real provider subject reuse/User merge demands `(issuer,subject)` handover between MetalDocs Users;
- a provider-driven immediate revocation SLA makes bounded staleness unacceptable;
- a concrete accepted journey requires durable provider token material rather than transient mechanism state;
- `session.manage` requires semantic, not support/mechanism, device/session metadata;
- an accepted non-Approval consumer requires durable Authentication-owned assurance history;
- measured Session lookup/revocation architecture becomes an actual scaling defect.

Preference, current legacy schema inconvenience or hypothetical enterprise IAM features are not reopen evidence.

---

# 21. Decision / next gate

Corrected B2-1 candidate:

```text
ProviderSubjectBinding
  total stable UNIQUE(issuer,subject)
  max one enabled binding/User
  mapping immutable
  reversible disabled_at acceptance
  no attribute-based subject selection

ApplicationSession
  local opaque server-revocable session
  multiple/User
  finite absolute TTL
  bearer never persisted replayably
  no tenant/AuthZ/provider-token snapshots

Fresh auth
  latest bounded Session evidence inputs
  explicit same-subject provider reauth
  never bare-non-NULL satisfaction
  consumer-owned explicit bounded freshness/one-shot rule

Provider disable
  bounded staleness

MetalDocs offboarding
  immediate local access revocation

Provider provisioning/reconciliation
  mechanism-driven, causally correlated, no fabricated binding
```

This artifact is not authority. The next gate is one independent **bounded delta review** of the adjudicated corrected target. B2-2/B2-3/B2-4 do not advance, and product implementation remains blocked.