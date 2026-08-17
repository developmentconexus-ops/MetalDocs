# MetalDocs R10-B2-1 — Authentication Binding / Session / Assurance — Independent Cold Adversarial Review

> **Status:** INDEPENDENT REVIEW — EVIDENCE ONLY — **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Reviewer:** Claude Fable 5 (cold session; no prior-conversation authority used)
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Candidate packet:** `docs/superpowers/analysis/2026-08-17-r10-b2-1-authentication-binding-session-assurance-fable-review-request.md` @ `9cba3acd`
> **Authority baseline:** `e1bd83ce9a0a9b70135e4bd6984d990a54ba6377`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** CLOSED — this review authorizes nothing.
> **Authority note:** findings below are evidence for operator adjudication. No finding amends R9.5, R10-A, R10-B1, the handoff, the ledger, or the candidate packet.

---

# 0. Bootstrap and evidence base

Read order executed fresh from repository truth at `9cba3acd` (candidate packet is the only delta over the
authority baseline `e1bd83ce`, verified via `git log e1bd83ce..HEAD`):

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md` (v1.0.0)
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — especially §1 (Authentication/Keycloak/anti-corruption contract, 43-permission catalog incl. `session.manage`), §2 (Approval, `requires_reauthentication`), §6 (privacy minimums), §14 (attestation evidence)
6. `wiki/architecture/r10-technical-architecture.md` — §2.1 Authentication ownership, §6 B1 substrate law, §7 B2-1 checklist
7. Candidate packet (full)
8. Current auth/IAM/security code as claim-specific evidence only

Code evidence consulted (evidence of current use, never target entitlement):

- `internal/modules/auth/domain/model.go:41-81` — current `Identity` couples password hash/algo, lockout
  (`FailedLoginAttempts`, `LockedUntil`), roles; current `Session` couples `TenantID`, `IPAddress`,
  `UserAgent`, `LastSeenAt`. Confirms the packet's legacy-coupling claim (§2).
- `internal/modules/auth/application/service.go:1264-1295` — 32-byte CSPRNG token, HMAC-SHA256 cookie
  framing, stored form = SHA-256 digest of raw token, constant-time signature check. Confirms the
  opaque-bearer / stored-digest property exists and works at current scale.
- `internal/modules/auth/application/service_session_opacity_test.go:32-154` — the four properties the
  packet cites are individually test-pinned: opacity/no-claims, CSPRNG entropy, stored-hash-never-raw,
  post-logout revocation (`REQ-AUTHN-3`).
- `internal/modules/auth/domain/session_admin.go:13-37` — current Sessions & Security admin listing
  consumes `IPAddress`/`UserAgent`/`LastSeenAt`. Relevant to P9/D12 disposition below.

Method applied proportionally: this is a material decision (authentication trust boundary, persistent data
meaning, concurrency/recovery correctness), so full decision-core depth was used, including Structural
Inversion, subtractive pass, and failure-mode pass.

---

# 1. VERDICT

```text
APPROVE R10-B2-1 AUTHENTICATION BINDING / SESSION / ASSURANCE
WITH MATERIAL FIXES
```

```text
BLOCKER = 0
MAJOR   = 3   (F1, F2, F3 — all bounded text/constraint-level corrections; no topology change)
LOW     = 5   (L1–L5 — recommendations/notes; none blocks promotion)
```

The two-family shape (`ProviderSubjectBinding` + `ApplicationSession`), the local opaque session, the
one-ACTIVE-binding rule, the assurance-as-bounded-session-state model, the provider availability posture,
and the concurrency invariants all **survive adversarial attack**. Three material defects were found; each
has an exact bounded correction in §7. None reopens authority outside B2-1.

---

# 2. Material findings

## F1 — MAJOR — Total `UNIQUE(issuer, subject)` creates a structural relink dead-end

**Claim.** Candidate §6 declares `UNIQUE (issuer, subject)` as a total table constraint (packet
lines 199-212, D3) while simultaneously requiring: mapping immutability (§6.1 — `user_id`/`issuer`/`subject`
never edited in place), replacement-by-new-row (§6.1), and a one-way lifecycle (§14 — "constrained
ACTIVE→DISABLED lifecycle").

**Failure scenario.** Any path that disables a binding and later needs to re-trust the *same*
`(issuer, subject)` is structurally impossible:

```text
admin performs explicit security unlink (mistaken or later reversed)
→ binding row (issuer,subject) now DISABLED, row retained as history
→ relink same (issuer,subject) to the same or another User
→ new-row creation violates total UNIQUE(issuer,subject)
→ in-place reactivation violates the one-way lifecycle
→ in-place user_id edit violates mapping immutability
```

The same wall blocks legitimate provider-account handover (same subject re-correlated to a different User
after a confirmed identity replacement) — packet §6.3 itself lists "identity replacement" as a supported
reason for replacement. The only escape is deleting the historical row under the privacy-cleanup clause,
which converts privacy erasure into an operational workaround — a structural smell, not a valid path.

**Root cause.** The candidate conflated two distinct invariants into one constraint: (a) *current
acceptance* — at most one User may currently be reachable from an `(issuer, subject)`; and (b) *historical
truth* — retained rows record past correlations. Only (a) needs uniqueness. Historical rows must not
occupy the uniqueness namespace.

**Target property.** At any instant, an `(issuer, subject)` resolves to at most one ACTIVE binding, and a
User holds at most one ACTIVE binding — while disabled history rows remain representable and relink
remains reachable through the ordinary new-row law.

**Fix (exact corrected target in §7).** Scope both uniqueness rules to ACTIVE rows:

```text
UNIQUE (issuer, subject) WHERE state = ACTIVE
UNIQUE (user_id)         WHERE state = ACTIVE
```

P1 remains fully protected: two concurrent binding creations for the same `(issuer, subject)` still
collide on the ACTIVE-scoped uniqueness and the loser fails closed at the database — the DB-backstop
property is preserved. Relink and confirmed handover become ordinary new-row operations. "Deployment-wide"
scoping (vs. the retired per-tenant scoping) is correct and unchanged; only the totality is wrong.

## F2 — MAJOR — Provisioning-intent reconciliation is an open path for email to become binding authority

**Claim.** Candidate §6.4 forbids email/username/display-name auto-binding, but then permits binding
creation from "reconciliation of a known provisioning intent" without constraining *how* the reconciler
identifies the provider subject. §11's "provider response uncertain" row and §12's transaction pattern
require a "confirmed stable `(issuer,subject)`" but never define what confirmation excludes.

**Failure scenario.** This is precisely the P4 attack, and it succeeds as written:

```text
Organization creates User (email = x@company.com) + durable provisioning intent
→ R10-D calls provider create → provider answers "conflict: email already exists"
  (account created manually in Keycloak by an admin — possibly for a different human)
→ reconciler treats the existing subject with matching email as "the" subject
  for this known intent and binds it
→ email match silently became binding authority; the wrong human may now hold
  the User's access
```

Nothing in the candidate text forbids this: the intent is "known", the provider response is not
"uncertain", and the resulting `(issuer, subject)` is stable. The hard invariant of §6.4 is bypassed by
its own fourth bullet.

**Root cause.** §6.4 constrains *ordinary login* JIT-binding but leaves the trusted-provisioning
path's subject-identification rule unstated. The invariant was placed on a flow, not on the property.

**Target property.** No binding row is ever created whose subject was selected by attribute matching
(email/username/display name). A reconciler may bind only a subject that (a) the provider operation for
this specific intent itself created/returned, or (b) an explicit human trusted-correlation decision
designated. Attribute equality is corroborating display information, never selection authority.

**Fix.** Add the invariant to B2-1 semantic authority (exact text in §7). Mechanics (how R10-D surfaces
the conflict for human correlation) stay in R10-D — this is a semantic law, not retry machinery.

## F3 — MAJOR — Assurance satisfaction contract is under-specified; persisted `latest_*` invites "non-null = fresh"

**Claim.** Candidate §8.1 defines `requires_reauthentication` as "an explicit provider authentication
challenge was completed for the **operation/session context**" and simultaneously declines to freeze any
freshness window. Assurance is persisted as `latest_reauthenticated_at`/`latest_provider_auth_time`/
`latest_acr`/`latest_amr` on the Session (§7, §8).

**Failure scenario.** The persistence is genuinely required — the reauth callback and the governed
action arrive as different requests, so the evidence must survive the gap. But the combination of
(a) durable `latest_*` state, (b) the ambiguous "operation/**session** context" wording, and (c) "no
window is frozen" leaves the weakest legal reading available to implementation:

```text
user reauthenticates once (e.g. for an approval three days ago)
→ Session.latest_reauthenticated_at is non-NULL
→ a later requires_reauthentication step reads non-NULL as satisfied
→ sudo-forever within one Session — attestation freshness silently degraded
```

Under the frozen Approval attestation law (ledger §14: decisions preserve "required AuthN
assurance/fresh-auth evidence") this reading would produce formally complete but semantically stale
evidence. The candidate's own D15 (initial login does not satisfy) shows the intended strictness — but
D15's rationale is defeated if a *previous operation's* reauth satisfies a *later* operation indefinitely.

**Root cause.** The candidate specifies the *storage* of assurance and the *production* of assurance
(reauth flow laws 8.2/8.3) but not the *consumption* contract, and the one sentence touching consumption
("operation/session context") is ambiguous between the strict and the degenerate reading.

**Target property.** Persisted `latest_*` fields are evidence inputs, never satisfaction by themselves.
Satisfaction of `requires_reauthentication` is always bounded: either one-shot consumption tied to the
operation, or an explicitly configured freshness window applied by the consumer's frozen policy — decided
before implementation (B4/Approval owns the policy choice; B2-1 owns the "never bare non-null" law).

**Fix.** Exact corrected text in §7. No schema change: the fields are correct; the contract sentence is
what must change.

---

# 3. Low findings / notes

**L1 — LOW — Binding lifecycle can shed its state enum (subtractive).** The candidate proudly derives
Session lifecycle without a persisted enum (§7.5: `revoked_at` + `expires_at`), yet gives Binding a
`state TEXT CHECK (ACTIVE|DISABLED)` plus `state_changed_at`. The same law applied uniformly:
`disabled_at TIMESTAMPTZ NULL` — `NULL` = ACTIVE, `NOT NULL` = DISABLED — deletes one column, one CHECK,
and one enum vocabulary while gaining the disable instant for free, and the F1 partial uniqueness becomes
`WHERE disabled_at IS NULL`. Recommendation, not defect; if the enum is kept, `state_changed_at`
duplicates information Audit already records.

**L2 — LOW — `session.manage` listing UX will want distinguishing metadata.** The frozen catalog keeps
`session.manage` (ledger R9 catalog), and the current admin surface lists IP/User-Agent/LastSeen
(`internal/modules/auth/domain/session_admin.go:29-37`). Excluding these from *semantic* authority is
correct (D12 ACCEPTED — they are PII-bearing operational telemetry, and revoke-all/offboarding needs none
of them), but R10-E should expect to add bounded *mechanism/support* state when it designs the session
listing, exactly through the door the candidate leaves open in §7.6. No action for B2-1.

**L3 — LOW — Absolute TTL is the provider-disable staleness bound; say so where TTL is configured.**
Under D16 bounded staleness, the maximum window in which a provider-only-disabled subject retains access
equals the Session absolute lifetime (unless reconciliation revokes earlier). The TTL configuration owner
(R10-C/R10-E security config) should inherit this as an explicit input to choosing the value. The
candidate's contract is honest; this only routes the consequence to the value-picker.

**L4 — LOW — Provider logout may need transient `id_token_hint`; not session authority.** OIDC
RP-initiated logout conventionally wants an `id_token_hint` (providers may prompt or refuse without it —
external claim; R10-E must verify against current Keycloak documentation when it designs logout). If
seamless provider logout becomes an accepted R10-E journey, the ID token may need to transit request-scoped
or short-lived mechanism state. This does not contradict D11: the token never becomes ApplicationSession
authority and normal requests never need it. Flagged so D11's "later concrete consumer" clause has a named
candidate consumer.

**L5 — LOW — C1 minimal enforcement sketch for B2-4 (guidance, in scope as invariant-declaration
support).** The declared serialization invariant is sufficient; the minimal READ COMMITTED realization is:

```text
login tx:      SELECT user-eligibility row FOR SHARE (or FOR KEY SHARE)
               + SELECT binding FOR SHARE validating ACTIVE
               + INSERT ApplicationSession
offboarding tx: SELECT same user row FOR UPDATE
               + set ineligible + UPDATE ... revoke all Sessions of that User
```

Whichever commits second observes the other: a blocked `FOR SHARE` re-evaluates the updated row on lock
release and sees ineligibility; an offboarding that acquires its lock after login-commit sweeps the new
Session in its revoke-all statement. A conditional `INSERT ... WHERE eligible` **alone** is insufficient
under READ COMMITTED (statement snapshot races the revoke sweep) — the row lock is the load-bearing part.
The identical pattern serializes C3 on the binding row. This confirms the candidate's enforcement
direction; the exact mechanism remains B2-4's to fix.

---

# 4. Disposition — B2-1-D1..D20

| ID | Decision | Disposition |
|---|---|---|
| D1 | exactly two semantic families | **ACCEPT** — every third-family pressure point (assurance history → consumer snapshot per ledger §14; provisioning state → R10-D mechanism; provider mirror → rejected shadow authority) was attacked and none requires semantic persistence. |
| D2 | stable identity = issuer + subject | **ACCEPT** — verbatim frozen authority (ledger §1). |
| D3 | UNIQUE(issuer,subject) deployment-wide | **ACCEPT scope / REJECT totality — F1.** Deployment-wide (vs per-tenant) is correct; the constraint must be ACTIVE-scoped or relink/handover dead-ends. |
| D4 | max one ACTIVE binding per User V1 | **ACCEPT** — see §5.2. |
| D5 | immutable mapping + new-row replacement | **ACCEPT** — with F1 fixed, replacement/relink is uniformly the new-row law; history rows stay truthful. |
| D6 | no email/username/display-name auto-binding, no login JIT | **ACCEPT with F2** — the invariant is right but currently bypassable through intent reconciliation; the corrected text closes the last attribute-authority path. |
| D7 | opaque local ApplicationSession + server-side revocation | **ACCEPT** — see §5.1; strongest alternative loses. |
| D8 | stored one-way verifier/digest, never raw bearer | **ACCEPT** — property already proven live (`service.go:1264-1295`, opacity test suite); DB-row disclosure yields no replayable bearer; `UNIQUE(credential_digest)` doubles as lookup and collision backstop. |
| D9 | finite absolute expiry; multiple Sessions/User | **ACCEPT** — see §5.4/L3. |
| D10 | no Tenant/AuthZ snapshot in Session | **ACCEPT** — single-company substrate fixes company context; canonical live Authorization makes snapshots a second authority. Resolution `Session → Binding → User` is one authority chain; duplicating `user_id` on Session adds a redundant (if immutability holds, merely dead) authority edge with no consumer — correctly omitted. |
| D11 | provider tokens are not Session authority | **ACCEPT** — no accepted V1 flow requires stored provider tokens (reauth = new challenge; logout = L4 transient candidate at most; no provider-API consumer V1). |
| D12 | no IP/UA/LastSeen semantic fields | **ACCEPT** — L2 note; current consumers are UX/presence, not frozen semantics. |
| D13 | assurance = bounded latest Session facts, no third table | **ACCEPT with F3** — storage shape correct and required (cross-request gap is the real consumer); the consumption contract must be stated. |
| D14 | reauth must prove same issuer+subject | **ACCEPT** — fail-closed identity pinning; combined with D18, a replaced/disabled binding's sessions are already revoked, so the C2 recheck subsumes binding recheck. |
| D15 | initial login ≠ explicit reauthentication | **ACCEPT with F3** — the strict default is right (challenge bound to operation intent, sudo-mode semantics); F3 prevents the same strictness being voided by stale prior reauth. `latest_provider_auth_time` from initial login remains stored, so a future deliberate freshness-window policy needs no schema change. |
| D16 | provider-only disable = bounded staleness | **ACCEPT** — see §5.5/L3. |
| D17 | MetalDocs offboarding revokes local Sessions immediately | **ACCEPT** — matches frozen privacy minimum (ledger §6) and does not inherit provider staleness. |
| D18 | binding disable/replacement revokes Sessions in same local tx | **ACCEPT** — load-bearing: same-tx atomicity is what makes C2/C3 rechecks sufficient and keeps a DISABLED binding's sessions unreachable even across crashes. |
| D19 | role/grant changes never revoke Sessions | **ACCEPT** — safe *because* D10 holds: no cached AuthZ exists to go stale; canonical evaluation applies downgrades on the next check. D10 and D19 stand or fall together — flagged as a paired invariant. |
| D20 | uncertain provider outcome never fabricates binding | **ACCEPT** — §12 two-transaction pattern + F2's confirmation rule make fabrication structurally excluded. |

---

# 5. Required sub-verdicts

## 5.1 Strongest alternative to local ApplicationSession — and why it loses

The strongest alternative is not the packet's Alternative A as written, but its best variant: **provider
access/ID token as bearer, validated locally by signature (stateless), with a local revocation list for
the frozen revocation requirements.** It removes per-request DB session lookup and one table.

It loses on structure, not taste:

1. **Revocation is a frozen requirement, and it regenerates the session table.** `session.manage`
   (catalog) and offboarding-revokes-sessions (ledger §6 privacy minimum) require server-side immediate
   revocation. A stateless token needs a revocation list consulted per request — which is a session table
   with a worse name, so the claimed deletion is illusory (duplicate-authority smell).
2. **Provider lifecycle becomes application lifecycle.** Token expiry/refresh semantics are provider
   authority; adopting them couples every request to provider policy and pushes refresh-token custody
   into MetalDocs — exactly the custody D11 excludes.
3. **Availability coupling.** Refresh flows put the provider on the ordinary-request path; the frozen
   posture (established sessions survive provider outage) dies or requires more machinery.
4. **Anti-corruption distance.** A claims-bearing token at the request path is a standing invitation for
   claim consumption; the frozen structural contract (no provider claims into AuthZ) is easiest to prove
   when no claims-bearing artifact exists past the login boundary.
5. **Scale reality.** One company per deployment; the current runtime already does per-request session
   lookup with no evidenced cost (runtime evidence). The stateless benefit answers a scale problem this
   architecture does not have — a textbook accidental-complexity trade.

**Local opaque ApplicationSession SURVIVES as the global maximum for V1.**

## 5.2 One-active-binding/User — SURVIVES

Keycloak-brokered upstream federation (SAML/LDAP/AD upstream) presents to MetalDocs as the *same*
Keycloak issuer+subject — brokering multiplies upstream IdPs, not MetalDocs-facing issuers, so federation
does not require simultaneous bindings now. Provider/realm migration is a cutover, served by the
replacement law; a migration needing *overlapping* validity of two issuers would be the real trigger to
relax D4, and relaxation is purely additive (drop the ACTIVE-scoped user uniqueness), so the seam is
already prepared. Binding-as-row-family (not columns on User) is the correct minimal structure that makes
the future relaxation non-breaking. One-active is the smallest sustainable V1.

## 5.3 Binding lifecycle — SOUND with F1 correction (+ L1 recommendation)

ACTIVE/DISABLED as *MetalDocs acceptance* (never provider mirror) is the correct anti-corruption reading;
no provider-shadow status fields. Offboarding-does-not-disable-binding is correct: eligibility is
Organization's fact, correlation truth is Authentication's, and login fails closed on either. Mapping
immutability + new-row replacement is correct history law. The one defect is F1's totality; L1 optionally
simplifies the representation.

## 5.4 Session minimum-field verdict — CORRECT SET after review

`id, subject_binding_id, credential_digest, created_at, expires_at, revoked_at` + four `latest_*`
assurance fields: every field has a named consumer or invariant; every excluded field (tenant, user_id
duplicate, roles/grants, provider tokens, email, IP/UA/LastSeen, idle state) was attacked and correctly
stays out (D10/D11/D12). Nothing further is deletable without losing a distinct property — see §6. Derived
three-state lifecycle without a persisted enum is the right law; `revoked_at` terminal is correct.

Finite absolute expiry without idle timeout is sufficient for V1: no frozen requirement demands idle
timeout; absolute TTL bounds exposure; adding idle state later is additive (L3 routes the TTL
consequence). Multiple sessions per User is correct; one-session-per-user has no frozen requirement and
would fight ordinary multi-device use.

## 5.5 Assurance / fresh-auth verdict — SHAPE CORRECT, CONTRACT MUST BE FIXED (F3)

Bounded latest-facts on Session + transient `FreshAuthEvidence` + consumer (Approval) snapshotting its own
decision evidence is the right ownership split — the ledger already places durable per-decision evidence
in Approval (§14 attestation), so an Authentication-owned event history would be a duplicate evidence
authority (correctly rejected). `acr`/`amr`/`auth_time` + local verification instant exactly matches the
frozen provider-contract enumeration (ledger §1: `issuer, subject, authenticated_at, auth_time, acr?,
amr?`) — neither over- nor under-frozen; `amr` bounded-representation deferral is legitimate
implementation-spec detail. Initial-login-does-not-satisfy is the right strict default. Same-issuer+subject
reauth pinning and no-revival-after-revocation are correct fail-closed laws. F3 is the single gap.

## 5.6 Provider-disable / live-session verdict — ACCEPT bounded staleness

Per-request introspection reintroduces the availability coupling the whole design exists to remove;
mandatory back-channel logout V1 has no frozen requirement demanding immediate provider-initiated
revocation and adds a provider-driven mutation path into local session authority. Bounded staleness with
(a) new login/reauth failing closed at the provider consult, (b) reconciliation able to revoke earlier,
and (c) authoritative offboarding never inheriting the staleness (9.3), is the honest, smallest V1
contract. L3 names the bound. Back-channel logout remains a clean future *addition* (it would only add an
earlier revocation trigger, not change authority) — the deferral is safe.

## 5.7 Offboarding / session concurrency verdict — INVARIANT SUFFICIENT

C1's forbidden-outcome + total-order statement is the correct invariant and correctly placed (B2-1
declares, B2-4 enforces). The minimal enforcement strategy exists and is cheap (L5: row-lock
serialization on the User eligibility row; login `FOR SHARE`, offboarding `FOR UPDATE` + same-tx revoke
sweep) — no isolation-level escalation, no advisory locks, no serializable retries needed. C2 is closed by
the conditional final assurance update; C3 by the same row-lock pattern on the binding row plus D18's
same-tx revocation. No surviving-session counterexample was found against the declared invariants.

## 5.8 Provider provisioning / reconciliation verdict — COMPLETE with F2

The six-case matrix covers the reachable semantic outcomes; cases probed for absence (attribute drift on a
bound subject — a non-event, since attributes are not identity; intent targeting an already-bound subject
— rejected by uniqueness; issuer-string migration — an operational replacement flow) need no seventh
semantic case. Rejecting a `provider_sync_status` semantic family is correct: attempt/retry/error state is
execution truth with mechanism lifecycle, and the durable-intent pattern (§12, two local transactions,
provider work between, no provider call inside any DB tx) carries uncertainty without fabricating
semantics. F2 closes the one hole (subject selection during reconciliation). Provisioning intent belongs
in R10-D mechanism state — confirmed; it is not Authentication semantic authority (Mechanism ≠ Authority).

## 5.9 Privacy verdict — ACCEPT

Both families erasable under lawful cleanup; Audit's surviving skeleton is PII-minimized and attributes
actors without FK-depending on Authentication rows (audit attribution is non-authoritative per B1 §6.2),
so erasing binding/session rows cannot orphan governed evidence. The `ON DELETE RESTRICT` edges force the
correct erasure order (sessions before binding) rather than permitting silent cascade — consistent with B1
reference law. The candidate correctly refuses to make Authentication rows a records-retention authority;
provider-subject↔User correlation is authentication plumbing, not governed decision evidence (governed
evidence references the MetalDocs User). No conflict with Retention/LegalHold because binding/session rows
are never governed retention subjects.

## 5.10 Structural anti-corruption / no cross-DB atomicity — VERIFIED

Every candidate operation was walked for hidden Keycloak↔MetalDocs atomicity (P25): login/callback
(provider validation strictly precedes the local tx), reauth (same), binding disable/replacement (local tx
only; provider confirmed beforehand), provisioning (two local txs bridged by durable intent),
reconciliation (reads provider truth, then local tx). None requires cross-DB atomicity; no XA/2PC pressure
exists anywhere in the candidate. No provider claim, role, group, or organization crosses into AuthZ; no
claims map exists; P26 found no shadow directory, no generic IAM platform, no duplicated identity
authority, and no new local maximum.

---

# 6. Subtractive / YAGNI pass — "what can still be deleted?"

Attacked every field/table/lifecycle for deletion without weakening a distinct property:

| Candidate element | Deletable? | Reason |
|---|---|---|
| `ProviderSubjectBinding.state` + `state_changed_at` | **YES → fold into `disabled_at`** (L1) | Same information, one column fewer, uniform with Session's enum-free law. Only recommended, not required. |
| Binding `created_at` | No | Correlation-establishment truth; cheap; ordering evidence. |
| `ApplicationSession.latest_acr` / `latest_amr` / `latest_provider_auth_time` | No | Frozen provider-contract facts (ledger §1) with a real cross-request consumer (reauth callback → governed action gap) and required by Approval attestation snapshotting. |
| `latest_reauthenticated_at` | No | The local verification instant is the only provider-independent freshness anchor; F3's bounded consumption reads it. |
| `credential_digest` uniqueness | No | Lookup index + collision/duplicate-insert backstop in one constraint. |
| Session `created_at` | No | `CHECK (expires_at > created_at)` anchor + issuance truth. |
| Third assurance table | Already absent | Correctly rejected — would duplicate Approval's evidence authority. |
| `provider_sync_status` family | Already absent | Correctly rejected — mechanism state, R10-D. |
| `user_id` on Session | Already absent | Correctly rejected — redundant authority edge, no consumer. |
| Idle-timeout / LastSeen semantics | Already absent | Correctly rejected — no frozen consumer; additive later. |

Nothing else survives deletion without weakening stable identity binding, local revocation, Approval
fresh-auth, offboarding correctness, provider failure honesty, or the prepared federation/migration seam.
The candidate is already near its subtractive floor — the notable residue is L1's enum.

**Structural Inversion Test (independently executed).** If MetalDocs had used Keycloak from day one with
no local password tables, the local model would still require: Organization.User (product identity),
a durable subject↔User correlation with MetalDocs-owned acceptance (binding), a locally revocable
application session (session.manage + offboarding + outage decoupling are product properties, not legacy
residue), bounded fresh-auth evidence for Approval attestation, and the anti-corruption boundary. It would
not invent: password/lockout tables, role/group mirrors, claims bridges, token vaults, sync state
machines, or per-request introspection. The candidate's §16 conclusion is **confirmed, not merely
plausible** — the retained elements are essential complexity; the deleted ones were accidental.

**Failure-mode pass.** Attacked for: unauthorized access (F2 was the one live vector — closed by fix;
F3 degrades attestation freshness, not access), false identity binding (F2), stale access after
offboarding (none — D17/C1/L5 close it locally), provider availability coupling (none — 9.1 posture),
fabricated provider truth (none — §12/D20). No BLOCKER-grade path found.

---

# 7. Exact corrected target (the three material fixes)

Operator adjudication should amend the B2-1 target as follows — everything else stands as written.

**Fix 1 (F1) — replace the binding uniqueness law:**

```text
UNIQUE (issuer, subject) applies only among ACTIVE bindings
UNIQUE (user_id)         applies only among ACTIVE bindings

conceptually:
  UNIQUE (issuer, subject) WHERE state = 'ACTIVE'
  UNIQUE (user_id)         WHERE state = 'ACTIVE'

DISABLED rows are retained history and never occupy the uniqueness namespace.
Re-trusting a previously disabled (issuer,subject) — same or different User —
uses the ordinary new-row replacement law. In-place reactivation and in-place
mapping edits remain forbidden.
```

(If L1 is adopted, the predicate becomes `WHERE disabled_at IS NULL`; identical semantics.)

**Fix 2 (F2) — add one binding-creation invariant to §6.4:**

```text
Subject selection is never attribute matching. A binding may only be created for
an (issuer,subject) that either:
  a) the provider operation executed for this specific provisioning/correlation
     intent itself created or explicitly returned as this intent's outcome; or
  b) an explicit trusted human correlation decision designated.
Email/username/display-name equality — including during provisioning-intent
reconciliation and provider "already exists" conflict handling — is corroborating
display information only and never selects the subject. A provider conflict on
attempted provisioning resolves to a pending explicit-correlation decision, never
to silent adoption of the existing provider account.
```

**Fix 3 (F3) — replace the §8.1 satisfaction sentence:**

```text
requires_reauthentication means an explicit provider authentication challenge was
completed for the operation context.

Persisted Session assurance fields (latest_reauthenticated_at, latest_provider_
auth_time, latest_acr, latest_amr) are evidence inputs; their mere presence
(non-NULL) never satisfies requires_reauthentication. A consumer accepts them
only under an explicitly bounded rule: one-shot consumption tied to the operation,
or a deliberately configured freshness window owned by the consumer's policy
authority (Approval/B4 for approval steps). No implicit or unbounded window
exists. Initial login does not satisfy requires_reauthentication (unchanged).
```

---

# 8. Reopen assessment / review routing

- **Material reopen outside B2-1:** **NONE.** All three fixes live inside B2-1's own text. No finding
  contradicts R9.5, GCR, the single-company rebaseline, R10-A ownership, or B1 substrate law. The frozen
  provider-contract enumeration, the 43-permission catalog, Approval attestation semantics, and the
  privacy minimums were used as constraints and all held.
- **New authority overlap introduced by the candidate:** none found (P26 clean).
- **Broad review required:** **NO.** The candidate's structure survived; the defects are bounded
  constraint/wording corrections with no topology or ownership movement.
- **After correction:** a **bounded delta review is sufficient** — verify the three amended passages and
  their internal consistency (F1×D5 replacement law; F2×§11/§12 reconciliation rows; F3×D15/§8) plus any
  operator-chosen L1 adoption. A full re-review is not warranted.

Per the packet's own rule: these findings are evidence; operator adjudication is the next authority gate.
No finding herein amends target authority.

---

# 9. Attack-question coverage record (P1–P26)

| P | Outcome |
|---|---|
| P1 | HELD (with F1 fix; ACTIVE-scoped uniqueness still DB-rejects the two-User race — verified both orderings) |
| P2 | HELD — brokered federation shares the MetalDocs-facing issuer; overlap-migration is the named future relax trigger; seam prepared |
| P3 | HELD — history rows erasable under §14, so no unnecessary permanence; F1 removes the one dead-end |
| P4 | **FALSIFIED as written → F2** — reconciliation adoption-on-conflict was a live attribute-authority path |
| P5 | HELD — no JIT path exists; login without ACTIVE binding fails closed with no side effects |
| P6 | HELD — no tenant/user/AuthZ duplication; single resolution chain |
| P7 | HELD — digest of 256-bit CSPRNG bearer; DB disclosure non-replayable (pattern live-proven in current runtime) |
| P8 | HELD — absolute expiry sufficient V1; idle timeout additive later (L3) |
| P9 | HELD — no frozen consumer; session.manage listing metadata is future mechanism state (L2) |
| P10 | HELD — issuer+subject pinning; identity fixed via immutable binding reference |
| P11 | HELD — conditional final mutation; revocation/expiry wins |
| P12 | HELD — established sessions provider-independent |
| P13 | HELD — bounded staleness acceptable; no frozen immediate-revocation requirement; back-channel additive later |
| P14 | HELD — offboarding revocation is purely local |
| P15 | HELD — C1 invariant sufficient; minimal enforcement exists (L5) |
| P16 | HELD — C3 + D18 same-tx sweep |
| P17 | HELD — matrix complete; no semantic sync-status family needed |
| P18 | HELD — §12 confirmed-subject precondition (tightened by F2) |
| P19 | HELD — no stored-token consumer V1; L4 names the only future candidate (transient, non-authority) |
| P20 | HELD — Approval owns durable decision evidence (ledger §14); latest+snapshot suffices; history table would duplicate authority |
| P21 | HELD — exactly the frozen assurance-fact enumeration; amr representation is legitimate impl detail |
| P22 | HELD — no frozen use case requires initial login to count; strict default correct (F3 protects it) |
| P23 | HELD — live canonical evaluation; D10/D19 paired invariant |
| P24 | HELD — PII-minimized Audit skeleton independent of auth rows; RESTRICT forces correct erasure order |
| P25 | HELD — all flows walked; no hidden cross-DB atomicity |
| P26 | HELD — no shadow authority, no platform, no new local maximum |

**Convergence:** single round; findings count 3 MAJOR / 5 LOW, altitude mechanical-to-textual (constraint
scoping + two invariant wordings), no structural recurrence. Stop condition met.
