# Current Agent Handoff

> **Last verified:** 2026-08-17
> **Status:** ACTIVE — Cohesive Platform Redesign / **R9.5 FROZEN + GCR-REFINED + SINGLE-COMPANY-REFINED / R10-A CLOSED + SINGLE-COMPANY-REFINED / R10-B1 CLOSED + SINGLE-COMPANY-RESTRUCTURED / R10-B2 IN PROGRESS / R10-B2-1 CLOSED / R10-B2-2 NEXT**
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Start with `AGENTS.md` and follow its authority chain. For the current stage read:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. **this file** for current status / next step
4. `wiki/architecture/cohesive-platform-redesign.md` — program authority / global coherence
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — frozen R3–R9.5 product/domain authority including promoted GCR + single-company refinements
6. `wiki/architecture/r10-technical-architecture.md` — active R10 technical authority including B1 and promoted B2-1
7. review artifacts only when auditing how a promoted decision was challenged

Git history is archive. Current code/schema/OpenAPI/module docs are current-state evidence only for target design.

---

## Current checkpoint

```text
R3–R9   = LOCKED

R9.5-1  = LOCKED / SINGLE-COMPANY-REFINED where tenancy wording changed
R9.5-2  = LOCKED / GCR-REFINED / SINGLE-COMPANY-REFINED
R9.5-3  = LOCKED (refined by R9.5-8)
R9.5-4  = LOCKED / SINGLE-COMPANY-REFINED where uniqueness wording changed
R9.5-5  = LOCKED (refined by R9.5-8; single-company privacy re-anchor applied)
R9.5-6  = LOCKED / SINGLE-COMPANY-REFINED (Tenant Portability deferred)
R9.5-7  = LOCKED / GCR-REFINED
R9.5-8  = CLOSED / APPROVED
R9.5    = FROZEN / GCR-REFINED / SINGLE-COMPANY-REFINED
reopen set = EMPTY

GCR                         = CLOSED / APPROVED
Single-Company Rebaseline   = CLOSED / APPROVED

R10-A    = CLOSED / APPROVED / GCR-REFINED / SINGLE-COMPANY-REFINED
R10-B    = IN PROGRESS / DESIGN ONLY
R10-B1   = CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED
R10-B2   = IN PROGRESS / DESIGN ONLY
R10-B2-1 = CLOSED / APPROVED
R10-B2-2 = NEXT / DESIGN ONLY
R10-B2-3 = NOT STARTED
R10-B2-4 = NOT STARTED
R10-B3   = NOT STARTED
R10-B4   = NOT STARTED
R10-B5   = NOT STARTED
R10-B6   = NOT STARTED
R10-C    = NOT STARTED
R10-D    = NOT STARTED
R10-E    = NOT STARTED
R10-F    = NOT STARTED

implementation = BLOCKED
```

Promoted target authority:

- R3–R9.5 product/domain semantics: `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
- R10 technical architecture: `wiki/architecture/r10-technical-architecture.md`
- program/global-coherence mirror: `wiki/architecture/cohesive-platform-redesign.md`

---

## Single-Company Deployment / Tenancy Rebaseline — promoted outcome

V1 deployment invariant:

> **One MetalDocs deployment serves exactly one company. The same product codebase, build artifacts and migrations are reused for every deployment; customer-specific forks are forbidden. Shared/pooled multi-customer tenancy is deferred until measured evidence selects it.**

`Tenant` is the single company/organization root of a deployment, whole-company Authorization scope target and deployment↔DB identity anchor. It is not a V1 partition dimension. Exactly one Tenant row exists per product DB; `Tenant.id` is immutable. Deployment config pins `expected_tenant_id`; missing/multiple/mismatch fails closed.

B1 target substrate:

```text
one PostgreSQL product DB / schema metaldocs
id UUID PRIMARY KEY
ordinary typed FKs
cross-owner RESTRICT / NO ACTION only
no universal tenant_id/company_id/deployment_id partition column
no composite tenant PK/FK
no Tenant/Area/role/Permission RLS
no Tenant GUC/context/customer routing
serving DB role non-owner + NOSUPERUSER
READ COMMITTED
required Audit + durable intent in same local commit
provider DB atomicity dependency forbidden
```

Tenant customer lifecycle/deletion/tombstones and Tenant Portability Export are deferred. User/data-subject privacy, Retention, LegalHold, Disposition, Backup/Restore remain.

---

## R10-B2-1 Authentication Binding / Session / Assurance — promoted outcome

R10-B2-1 closed after candidate → independent cold review → operator adjudication/corrected target → bounded delta review.

Evidence chain:

```text
candidate                9cba3acd
independent review       361f6c8b  APPROVE WITH MATERIAL FIXES
corrected target          ee0a0ce0
bounded delta review      6593c471  APPROVE
BLOCKER                   0
MAJOR                     0
prior findings closed     8/8
new material contradiction NONE
new concurrency counterexample NONE
broad review required     NO
```

### Authentication persistent state

Exactly two semantic persistent families:

```text
ProviderSubjectBinding
ApplicationSession
```

No provider-sync semantic FSM, provider account mirror, claims mirror, assurance-event history or provider-token authority.

### ProviderSubjectBinding

```text
id          UUID PK
user_id     UUID FK → Organization.User
issuer      TEXT
subject     TEXT
created_at  TIMESTAMPTZ
disabled_at TIMESTAMPTZ NULL

UNIQUE(issuer, subject)
UNIQUE(user_id) WHERE disabled_at IS NULL
```

`issuer+subject` is stable provider identity. `user_id/issuer/subject` are immutable mapping fields. `disabled_at` is reversible MetalDocs acceptance state for the same mapping; Audit owns disable/re-enable history. At most one enabled binding/User V1.

A retained `(issuer,subject)` cannot normally be handed to another MetalDocs User. Same-subject re-trust for the same User re-enables the same row. Handover/User merge/provider subject reuse is a material reopen. After lawful erasure of the row, a later correlation is a new trusted decision and the old DB structural no-recorrelation guarantee has intentionally been surrendered.

Subject selection must be causally/verifiably tied to the exact provisioning/correlation intent or an explicit trusted-human correlation decision. Email, username, display name, similar name or provider "already exists" attribute matching never select a provider subject.

### ApplicationSession

```text
id                        UUID PK
subject_binding_id        UUID FK
credential_digest         BYTEA
created_at                TIMESTAMPTZ
expires_at                TIMESTAMPTZ
revoked_at                TIMESTAMPTZ NULL
latest_reauthenticated_at TIMESTAMPTZ NULL
latest_provider_auth_time TIMESTAMPTZ NULL
latest_acr                bounded nullable
latest_amr                bounded nullable
```

Properties:

- high-entropy opaque browser bearer;
- database stores only one-way verifier/digest, never replayable raw bearer;
- finite absolute expiry;
- multiple Sessions/User allowed;
- terminal revocation; reauth never revives revoked/expired Session;
- no Tenant/company dimension;
- no duplicated persisted `user_id`;
- no roles/permissions/groups/AuthZ snapshot;
- no normal provider access/refresh/ID-token authority;
- no semantic IP/User-Agent/LastSeen/idle state.

Runtime authenticated context resolves `Session → Binding → User` and canonical Authorization reads current authority.

### Fresh-auth

Persisted Session `latest_*` fields are evidence inputs only. Bare non-NULL never satisfies `requires_reauthentication`.

A consumer must apply an explicit bounded rule:

```text
one-shot operation-linked evidence
OR
explicit configured freshness window owned by that consumer's policy authority
```

Initial login does not automatically satisfy later reauthentication. Forced reauth must prove the same `(issuer,subject)` and only updates a still-valid Session. Authentication may emit transient `FreshAuthEvidence`; Approval/B4 owns approval freshness policy and snapshots its consumed evidence.

### Provider disable / offboarding / availability

```text
provider-only disable/removal
→ new login/reauth fail
→ reconciliation may revoke earlier
→ existing Session survives at most until local revoke or finite expiry
```

Finite Session TTL is therefore the provider-only-disable staleness upper bound absent earlier reconciliation; R10-E/security configuration chooses the value later.

Keycloak outage does not by itself invalidate established local Sessions; new login/reauth fail visibly and provider operations retry through R10-D.

MetalDocs User offboarding is authoritative for MetalDocs access and must revoke local Sessions without waiting for Keycloak. Role/group/grant mutation does not revoke Sessions because AuthZ is never cached in Session.

### Binding mutations / reconciliation

Disable binding → revoke all its Sessions in the same local DB transaction. Replacement to a proven new subject atomically disables old binding, enables/creates new binding, revokes old-bound Sessions and writes required Audit. Re-enable is also an acceptance mutation and never revives Sessions.

Six reconciliation outcomes remain authoritative: User/subject absent combinations, removed/disabled provider subject, duplicate issuer+subject, provider unavailable, uncertain provider response. Uncertain provider truth never fabricates binding truth; retry/attempt/error/pending-correlation state belongs R10-D mechanism persistence. No provider call participates in MetalDocs DB atomicity.

### B2-4/B6 successor notes

- B2-4 must serialize login/session issuance vs User offboarding (C1).
- B2-4 must serialize Session issuance vs binding disable/re-enable/replacement (C3); re-enable is explicitly included.
- reauth callback cannot overwrite revoked/expired Session (C2).
- grant mutation vs Session is safe because Session has no AuthZ snapshot (C4).
- lawful Binding erasure surrenders structural no-recorrelation for the erased subject; B2-4/B6 privacy classification must state this explicitly.
- exact row-lock/transaction mechanics remain B2-4; B2-1 only fixes the invariants.

---

## Exact next step — R10-B2-2 Organization

Open **R10-B2-2 — Organization singleton root / people / groups** in design-only mode. B2-1 is closed; do not rediscover its decisions.

B2-2 intake must cover:

```text
Tenant
  exact singleton persistent shape
  structural exactly-one enforcement
  immutable Tenant.id
  editable company identity/settings
  expected_tenant_id startup/readiness consumer surface

Area
  identity/stable fields
  actual lifecycle requirement if any
  deployment-wide uniqueness

User
  technical identity
  authentication eligibility/offboarding state
  erasable profile/enrichment boundary
  Area relationship only if semantically justified
  no provider-state mirror

Group
  flat-group identity
  deployment-wide uniqueness
  actual lifecycle requirement if any

GroupMembership
  User↔Group relation
  mutation/evidence semantics
  no nested groups

B2-1 integration
  ProviderSubjectBinding FK/integrity to User
  Session issuance consumes current User eligibility + accepted Binding
  offboarding revokes ApplicationSessions race-safely
```

Do not design RoleAssignment/grants (B2-3), final transaction/lock matrix (B2-4), provider retry mechanics (R10-D), admin/frontend journeys (R10-E), or resurrect customer Tenant lifecycle/partitioning/RLS.

Implementation remains **BLOCKED**.
