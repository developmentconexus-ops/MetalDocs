# MetalDocs R10-B2-2 — Organization Singleton Root / People / Groups — Independent Cold Adversarial Review

> **Status:** INDEPENDENT REVIEW — EVIDENCE ONLY — **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Reviewer:** Claude Fable 5 (cold session; repository truth at `8dfffb65`; no prior-conversation authority used)
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Candidate packet:** `docs/superpowers/analysis/2026-08-17-r10-b2-2-organization-singleton-people-groups-fable-review-request.md` @ `8dfffb65`
> **Authority baseline:** `71791dfecd4cd185684373ffcdccbf256138b741` (B2-1 promotion)
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** CLOSED — this review authorizes nothing.
> **Authority note:** findings are evidence for operator adjudication; nothing here amends authority.

---

# 0. Bootstrap and evidence base

Read order executed fresh: `AGENTS.md` → Method v1.0.0 → `wiki/references/current-agent-handoff.md` →
`wiki/architecture/cohesive-platform-redesign.md` → frozen ledger (§1 Organization/roles/permissions, §2
Approval, §5 Distribution/values, §6 privacy/singleton root) → `wiki/architecture/r10-technical-architecture.md`
(promoted §7 B2-1, §8 B2-2 checklist) → candidate packet → current IAM/auth/taxonomy schema/code as
claim-specific evidence. Verified via `git log`: baseline `71791dfe` promoted B2-1 into authority; the only
delta to `8dfffb65` is this candidate packet.

Code/schema evidence consulted (evidence of current use, never target entitlement):

- `db/baseline/0001_current_schema.sql:1610-1619` — current `tenants(id, name, slug UNIQUE, created_at,
  updated_at, erased_at)`.
- `db/baseline/0001_current_schema.sql:1286-1301` — current `iam_users(user_id TEXT PK, display_name,
  is_active, deactivated_at, last_login_ip/user_agent/device_label, last_seen_at, mfa_enabled,
  mfa_enrolled_at, tenant_id)`.
- `db/baseline/0001_current_schema.sql:1130-1145` — current Area table `document_process_areas(code PK,
  name, description, is_active, parent_code, owner_user_id, default_approver_role, archived_at)` with
  lowercase code format CHECK.
- `db/baseline/0001_current_schema.sql:1230-1261, 2705` — `iam_group_members(group_id, user_id, tenant_id,
  granted_at, granted_by)`; `iam_groups(id, tenant_id, name, description)` with `UNIQUE(tenant_id, name)`.
- `db/baseline/0001_current_schema.sql:1651-1663` — `user_process_areas` role-in-area grants (Authorization
  family, not organizational placement).
- `internal/modules/taxonomy/delivery/http/routes_areas.go:32,129,185` — live `archiveArea` route,
  `include_archived` listing filter, and `409 "process area is archived"` conflict gate.
- `internal/modules/taxonomy/domain/area.go:115-123` — `IsActive()` gate consumed by mutation paths.
- `internal/modules/taxonomy/application/area_service.go:110-143` — `parent_code` is CRUD passthrough +
  cycle check only; no inheritance/permission semantics found.

Method applied at full depth: Organization identity is a material trust/persistence boundary.

---

# 1. VERDICT

```text
APPROVE R10-B2-2 ORGANIZATION SINGLETON / PEOPLE / GROUPS
WITH MATERIAL FIXES
```

```text
BLOCKER = 0
MAJOR   = 1   (F1 — Area minimal retirement lifecycle is required now, not later)
LOW     = 5   (L1–L5 — bounded wording/successor-gate notes)
```

The six-family shape (`Tenant, Area, User, UserProfile, Group, GroupMembership`), the singleton
enforcement, the User/UserProfile split, the attribute-never-identity laws, the User→Area omission, and
the pair-keyed current-only GroupMembership all **survive adversarial attack**. One material defect: the
candidate is **too small in exactly one place** — Area retirement.

---

# 2. Material finding

## F1 — MAJOR — Area needs a minimal retirement lifecycle now; the candidate's own trigger is already satisfied

**Claim.** Candidate D5/§7 omits any Area lifecycle and states: "a real requirement to preserve an Area
while prohibiting future assignment is the trigger for a minimal lifecycle" — classifying that requirement
as future.

**Evidence that the trigger is already met.**

1. The shipped product already has exactly this capability: a live archive route
   (`internal/modules/taxonomy/delivery/http/routes_areas.go:129`), an `include_archived` listing filter
   (`:32`), a `409 "process area is archived"` fail-closed gate on further use (`:185`), and a domain
   `IsActive()` gate (`internal/modules/taxonomy/domain/area.go:115-123`), backed by
   `archived_at`/`is_active` columns (`db/baseline/0001_current_schema.sql:1134,1140`). This is not an
   unused legacy column — it is an operating product capability with an OpenAPI contract and enforcement.
   Classification: **KNOWN REQUIREMENT**, not legacy mechanism.
2. Frozen authority makes retirement-by-deletion structurally impossible: Document Area is a stable
   Controlled Information fact (ledger §3: "Document code/type/Area are stable V1") and cross-owner FKs
   are RESTRICT/NO ACTION only (B1 §6.2). Any Area ever referenced by a Document can never be deleted.
3. Therefore, under the candidate as written, an organizationally obsolete Area **remains assignable
   forever** — new Documents, new `AreaScope` grants (B2-3), new `RoleInArea` policy references — with no
   representable way to retire it. "FK prevents delete" was conflated with "must remain usable"; the
   packet itself names this trap (§11) and then walks into it.

**Root cause.** The candidate classified current runtime archive state as unproven legacy instead of
running the packet's own §13 classification against it; the deletion-impossibility consequence of frozen
Document-Area stability then went unexamined.

**Target property.** An Area can be preserved as referenced historical identity while structurally
refusing new assignment, reversibly, without deleting anything.

**Fix (exact corrected target).** Add one field and one law, uniform with the already-promoted
`disabled_at` pattern (User, ProviderSubjectBinding):

```text
Area
  id          UUID PRIMARY KEY
  code        TEXT NOT NULL
  name        TEXT NOT NULL
  disabled_at TIMESTAMPTZ NULL

UNIQUE(code)

disabled_at IS NULL     → Area accepts new references/assignments
disabled_at IS NOT NULL → Area is retired: existing references remain valid;
                          new Document assignment, new AreaScope grants and new
                          policy references fail closed at their owning boundaries
                          (Controlled Information, B2-3, B4)

id/code immutable; name mutable; disabled_at reversible; Audit owns transitions.
```

This adds no state machine, no hierarchy, no second authority; consumers enforce the no-new-assignment
gate at their own boundaries exactly as they consume Area identity today. Severity is MAJOR, not BLOCKER:
bounded one-field correction, no topology movement.

---

# 3. Low findings

**L1 — Area.code normalization law.** Current runtime enforces lowercase codes via CHECK
(`0001_current_schema.sql:1141`) and frozen numbering consumes `{AREA}` verbatim into generated document
codes. The candidate is silent on case/normalization. One sentence suffices: code normalization/format is
fixed at creation as implementation-spec detail; rendered numbering consumes the stored code verbatim
(Controlled Information's concern). Prevents a silent format drift between Area identity and governed
numbering.

**L2 — Offboarding disposition + re-enable policy must be a named B2-4 gate decision.** The candidate's
deferral (§15) is **safe** for access correctness: retained memberships/grants are dormant while disabled
because every access path is session-gated and issuance requires eligibility (promoted §7 B2-1; system
paths run as PlatformOperator/SystemPrincipal outside company RBAC). But re-enabling a User silently
restores whatever memberships/grants were retained — right for short leave, wrong for a two-year rehire
(stale privilege). This restore-vs-regrant policy is deliberately deferred; make it an explicit mandatory
B2-4 closure item ("decide retained-vs-removed on offboarding and restore-vs-regrant on re-enable;
silent-default retention may not ship unexamined") so the deferral cannot decay into an accident. Not
MAJOR: no reachable unsafe state exists before B2-4 closes, and implementation is blocked until whole-B2
closes in order.

**L3 — Record the GroupMembership pair-PK ruling against B1 explicitly.** `(user_id, group_id)` is a
composite of two **internal technical UUIDs** — not a business/provider/external identifier — identifying
a relationship fact, not an ordinary durable entity. That is compatible with B1's identity law, but B2-4's
persistence classification should record this ruling in one line, so "every table gets a surrogate UUID"
drift does not later reopen it in either direction.

**L4 — Profile-presence expectation.** State explicitly: an eligible User is normally profile-complete;
profile absence means lawful erasure or provisioning-in-progress, and UI renders a neutral fallback for
absent profiles. One sentence; prevents divergent consumer assumptions about `User` without `UserProfile`.

**L5 — Group/Tenant name normalization.** Blank/whitespace/case handling for `Group.name`,
`Tenant.display_name`, `Area.name` is implementation-spec (current runtime has not-blank CHECKs,
`0001_current_schema.sql:1617-1618`); no semantic law needed — note only so implementers do not invent
case-insensitive uniqueness as accidental identity semantics.

---

# 4. Disposition — B2-2-D1..D12

| ID | Decision | Disposition |
|---|---|---|
| D1 | Tenant = `id + display_name` only | **ACCEPT** — `display_name` has a real consumer (company identity under `tenant.settings.manage`); slug/status/settings-JSON all correctly dead (see §13 sweep). |
| D2 | exactly-one = constant-expression unique + readiness | **ACCEPT** — see §5.1. |
| D3 | no generic settings platform | **ACCEPT** — typed facts with real consumers only; future typed settings join their semantic owner. |
| D4 | Area id/code/name, code immutable V1 | **ACCEPT** — `{AREA}` numbering and sequence-bucket continuity make casual recode unsafe; explicit-recode-as-material-operation is the right escape hatch (L1 note). |
| D5 | no Area hierarchy or lifecycle | **SPLIT** — hierarchy omission ACCEPT (`parent_code` is CRUD passthrough with no semantic consumer, `area_service.go:110-143`; no frozen consumer); lifecycle omission **REJECT → F1**. |
| D6 | User minimal root | **ACCEPT** — no credential/provider/AuthZ/tenant/HR fact has a frozen consumer on User; B2-1 FKs `User.id` cleanly. |
| D7 | reversible `disabled_at` eligibility | **ACCEPT** — uniform with promoted Binding law; Audit owns transitions; no accepted consumer distinguishes SUSPENDED/TERMINATED/etc.; a terminal state adds nothing an erased profile + disabled flag does not already express. |
| D8 | User/UserProfile split | **ACCEPT — required verdict in §5.4.** |
| D9 | attributes never identity; no `UNIQUE(email)` | **ACCEPT** — identity-level email uniqueness would recreate email-as-authority pressure against promoted B2-1 §7.3 and break legitimate duplicate/recycled-email realities; operational duplicate warnings are R10-E UX validation without identity semantics. |
| D10 | no User→Area / home_area | **ACCEPT** — see §5.6. |
| D11 | Group = flat, company-wide, `UNIQUE(name)` | **ACCEPT** — see §5.7. |
| D12 | GroupMembership current-only, pair-keyed | **ACCEPT — required verdict in §5.8.** |

---

# 5. Required sub-verdicts

## 5.1 Tenant singleton / enforcement — APPROVE

`CREATE UNIQUE INDEX ... ON tenant ((true))` is valid PostgreSQL (immutable constant expression index;
table-level UNIQUE constraints cannot hold expressions, so an index is the correct realization — the
packet's "conceptually" wording already allows this). It is directly falsifiable: a second INSERT fails
with a uniqueness violation regardless of isolation or concurrency. The credible alternative — a
`singleton boolean PRIMARY KEY DEFAULT TRUE CHECK (singleton)` column — is equal in strength but adds a
fake business column; a trigger adds nothing the index does not already guarantee and is accidental
complexity. A CHECK pinning `id` to a constant is impossible because `Tenant.id` varies per deployment.
Constant-expression unique = minimal, strongest-reasonable enforcement. **At-most-one (DB) + at-least-one
(promoted fail-closed readiness handshake) genuinely compose to exactly-one for every serving deployment**
— the DB alone cannot demand at-least-one (an empty database during provisioning must be representable),
so the split enforcement is not a gap but the correct boundary assignment. `created_at` on Tenant:
correctly absent — no consumer; Audit owns provisioning events.

## 5.2 Area identity / lifecycle — APPROVE WITH F1

Identity shape (UUID id + immutable unique code + mutable name) is correct: code is genuine business
identity (frozen `{TYPE}/{AREA}/{SEQ}` numbering consumes it; recoding mid-life would fork sequence
buckets and governed prefixes — history is not corrupted because generated codes are stored, but future
semantics silently shift, which is exactly why casual mutation is forbidden). Deployment-wide
`UNIQUE(code)` is the correct re-derivation. Hierarchy stays out. Lifecycle must come in — F1.

## 5.3 User lifecycle — APPROVE

`disabled_at` reversible eligibility is sufficient V1: session issuance gates on it (promoted B2-1),
offboarding sets it, re-enable clears it, Audit owns history. No created_at needed (Audit owns creation).
No human/business key beside UUID: every governed reference is by UUID; human identification is profile
display data. A User without profile validly exists (erased or mid-provisioning — L4). Re-enable
introduces no identity ambiguity because `User.id` never changes meaning. No employee-number requirement
exists in frozen authority or runtime evidence (`iam_users` has none). Service identities stay outside:
PlatformOperator/SystemPrincipal are already frozen as outside company RBAC — modeling them as Users
would create implicit content authority; correctly absent.

## 5.4 User + UserProfile split — required explicit verdict

```text
USER + USERPROFILE SPLIT = APPROVE
```

**Concrete property the monolithic alternative cannot express structurally.** The split makes erasure a
**structural absence** instead of a sentinel: with `UserProfile(display_name NOT NULL)` 1:1-subordinate,
the DB can enforce simultaneously (a) an existing profile is complete (NOT NULL holds for every profile
row) and (b) an erased person has no profile (row absent — unambiguous, fail-closed). The monolithic
`User(id, display_name, email, disabled_at)` cannot hold both: it must either make `display_name`
nullable for everyone (losing the completeness invariant for active users and making "erased" and
"never filled" indistinguishable) or scrub with placeholder sentinel values — fabricated data that
violates the B1 primitive law "real unknown = NULL" and the no-fallback posture. Additional supporting
properties: the PII boundary becomes a table boundary (B6 field-classification and the R10-C
restore-non-resurrection proof reason about row presence, not per-column scrub state), and governed
owners FK the stable `User` root (B2-1 `ProviderSubjectBinding.user_id` already does) so RESTRICT
references never collide with erasure — sessions/bindings/profile erase while `User.id` survives as the
PII-minimized skeleton anchor.

Guardrails verified: `UserProfile` is 1:1 keyed by `user_id` (no second UUID → cannot drift into a second
person authority); classification vocabulary (semantic-authority-subordinate vs attributed support) does
not matter to correctness — what matters is one owner + erasability, both present (P23 holds).
Alternative D (Audit snapshots all human names) is strictly worse: it duplicates PII into immutable
evidence and directly contradicts the frozen PII-minimized-skeleton law. What survives lawful erasure:
`User.id`, `disabled_at`, governed domain evidence referencing `User.id`, Audit's PII-minimized skeleton.
What dies: profile row, eligible Authentication rows (per promoted B2-1 §7.9). UI renders historical
actors via an opaque/neutral fallback — sufficient (P10 holds).

## 5.5 Email / username — APPROVE

Email login/recovery/uniqueness is wholly Keycloak's problem (frozen provider ownership). Product-side
`UNIQUE(email)` would silently rebuild email-as-identity — the exact pressure B2-1 F2 closed — and would
break real duplicate/recycled-email cases. Invitation/provisioning flows correlate by intent-bound
mechanisms, never email (promoted §7.3), so no flow needs uniqueness. Username has no Organization
consumer anywhere in frozen authority; provider/UI projections may display provider attributes without
persistence. Admin-flow disambiguation between same-email Users uses display name + context — a UX
concern, not identity law.

## 5.6 User→Area omission — APPROVE

Frozen authority reads `RoleInArea` as an Authorization-derived resolution ("Approval actor resolution"
reuses Area as organizational truth via grants; Step `actor_rule: RoleInArea` consumes Authorization +
Area), not as workforce placement. Current runtime confirms: the only user↔area fact is
`user_process_areas(user_id, area_code, role, ...)` — role-scoped **grants** (`0001_current_schema.sql:
1651-1663`), i.e. the B2-3 RoleAssignment ancestor; `iam_users` has no home-area column at all. Document
owner/responsibility is a Controlled Information fact about the Document, not the person. Distribution
audiences resolve from grants/groups snapshots. Area "manager" is the `area_manager` role bundle — a
grant, not an employment fact. Adding `home_area` now would create two Area-truths (placement vs grants)
with zero consumers for the first — the exact duplicate-authority defect the Method forbids. Reopen
trigger (real HR/reporting placement fact) is correctly named.

## 5.7 Group shape / lifecycle — APPROVE

Flat, company-wide, `id UUID + name UNIQUE`: sufficient. Stable-code attack fails: no external/business
contract references Groups by code anywhere in frozen authority or runtime (`iam_groups` has no code
column today — `0001_current_schema.sql:1255-1261`); admin/API references use UUID; historical evidence
does not need Group identity because Approval activation and Distribution release snapshot **resolved
concrete Users** (frozen ledger), never live group references. UUID already provides stable identity
under rename. `UNIQUE(tenant_id, name)` re-derives to deployment-wide `UNIQUE(name)` correctly. Lifecycle
omission is genuinely YAGNI for Group — unlike Area, no shipped retirement capability exists (no archive
route found), and deletion is reachable: after removing memberships and (B2-3) role assignments, no
historical FK pins a Group, so RESTRICT never permanently blocks deletion. Same company-wide Group taking
RoleAssignments at different scopes is exactly the frozen model (grant carries the scope; Group carries
none). Case-insensitivity of name uniqueness = implementation detail (L5).

## 5.8 GroupMembership — required explicit verdict

```text
GROUPMEMBERSHIP SURROGATE UUID = YAGNI
```

Attacked hard, per packet §14 and prompt §8:

- **B1 compatibility:** B1 mandates UUID PKs for *ordinary durable entities* and forbids
  *business/provider/external* identifiers as technical PKs. `(user_id, group_id)` is neither: it is a
  composite of two internal technical UUIDs identifying a pure relationship fact. No law is violated
  (L3 records the ruling).
- **No consumer references a Membership as an entity:** Audit events carry the (user, group, actor, time)
  facts — they do not FK the membership row (which is deleted on removal; a FK would make current-only
  semantics impossible). Approval snapshots resolved participants at Step activation; Distribution
  snapshots concrete audiences at release — both frozen, both eliminate any need for membership history
  or membership identity. Authorization (B2-3) consumes *current* membership only, by pair. B2-4
  transactions target the pair directly.
- **Remove + re-add ambiguity:** none — with no external references to the row, recreating the same pair
  is semantically idempotent; interval history is Audit's, and Audit needs no row identity to record
  transitions (current runtime's `granted_at/granted_by` on the membership row are the Audit-duplicate
  being correctly deleted).
- **"Cheap seam" argument rejected:** a surrogate key invites future consumers to FK a row whose
  current-only delete semantics forbid stable reference — the seam would be a trap. If a real consumer
  ever needs addressable membership (named reopen trigger), it needs interval semantics too, which is a
  different family, not a column.
- `created_at` on the row: convenience duplicate of Audit; correctly absent.

Current-only + Audit is the sufficient historical model (P18 holds).

## 5.9 Offboarding boundary — APPROVE with L2

B2-2 fixes eligibility; B2-1 fixes race-safe Session revocation; B2-4 fixes transactions and the
retention/disposition policy. Verified no hidden access path from retained rows while disabled: every
product access path traverses Session issuance (eligibility-gated) or an existing Session (revoked at
offboarding); background execution runs as SystemPrincipal outside company RBAC; canonical Authorization
is only ever evaluated behind an authenticated context. Deferral is Method-clean (safe-now + named owner
+ trigger) — with L2 converting the re-enable restore-vs-regrant policy into an explicit B2-4 gate item.

## 5.10 Privacy — APPROVE

End-to-end walk: `UserProfile` (direct PII) — erasable by row delete; `ApplicationSession` /
`ProviderSubjectBinding` — erasable per promoted B2-1 §7.9 in RESTRICT order; `GroupMembership` — current
state, deletable, no PII beyond pseudonymous UUIDs; future RoleAssignment — B2-3, references User by
UUID; Audit — PII-minimized skeleton per frozen law, no FK-dependence on erased rows; Approval/
Distribution evidence — reference `User.id`, no embedded profile PII (B6 proves field-by-field);
Document responsibility — CI fact by UUID. Lawful erasure never requires deleting the `User` root — the
root UUID *is* the PII-minimized reference that lets governed history survive; physical User deletion
would cascade-orphan governed evidence and is structurally blocked by RESTRICT, which is correct, not a
defect. Architecturally, a bare UUID with erased enrichment is exactly the "PII-minimized skeleton"
posture frozen authority already accepts; whether residual linkability requires more is B6's field-by-field
proof, not a B2-2 schema question. Scrub-in-place is strictly weaker than structural absence (§5.4). No
privacy platform needed; none proposed. Broken-journey check: absent profile degrades to fallback
rendering only (L4).

---

# 6. Current-implementation evidence sweep (packet §13 / P24)

| Current fact | Classification | Target disposition |
|---|---|---|
| `tenants.slug` (+ UNIQUE) | LEGACY MECHANISM | dead — multi-tenant routing/login selector; single-company + provider-hosted journeys removed the consumer |
| `tenants.erased_at` | DEFERRED | customer lifecycle deferred by authority |
| `tenants.created_at/updated_at` | ACCIDENTAL | Audit owns; ops convenience only |
| `iam_users.display_name` | KNOWN REQUIREMENT | → `UserProfile.display_name` |
| `iam_users.is_active/deactivated_at` | KNOWN REQUIREMENT | → `User.disabled_at` |
| `iam_users.last_login_ip/user_agent/device_label/last_seen_at` | LEGACY MECHANISM | telemetry; R10-E support state if a concrete consumer appears (mirrors promoted B2-1 ruling) |
| `iam_users.mfa_enabled/mfa_enrolled_at` | LEGACY MECHANISM | provider-owned (Keycloak MFA); delete at cutover |
| auth `users` credential table (username/email/password/lockout) | LEGACY MECHANISM | deleted per promoted target; email survives only as `UserProfile.email` attribute |
| `document_process_areas.archived_at/is_active` + archive route | **KNOWN REQUIREMENT** | → **F1** `Area.disabled_at` |
| `document_process_areas.parent_code` | ACCIDENTAL | CRUD passthrough, no semantic consumer; no frozen hierarchy |
| `document_process_areas.owner_user_id/default_approver_role` | LEGACY MECHANISM | approval routing → ApprovalPolicy (B4) |
| `document_process_areas.description` | ACCIDENTAL | no frozen consumer; omit |
| `iam_groups.description` | ACCIDENTAL | omit |
| `iam_group_members.granted_at/granted_by` | ACCIDENTAL | Audit-duplicate; Audit owns transitions |
| `user_process_areas` (role, effective intervals) | out of B2-2 scope | B2-3 RoleAssignment ancestor |
| every `tenant_id` column | LEGACY MECHANISM | removed by single-company substrate |

No legacy field with an evidenced V1 consumer was removed by the candidate **except** Area
archival (F1) — the sweep found exactly one miss.

---

# 7. Structural Inversion Test

Designed from day one with Keycloak, separate Authorization, and one company per deployment, Organization
would still necessarily contain: a singleton company root with immutable identity (deployment↔DB
handshake + company display identity); an Area registry with stable code (numbering, document ownership,
scope target) **including a way to retire an Area** (organizations restructure; references are forever);
a stable person identity for governed attribution (`User.id`); erasable human enrichment separate from
that identity (privacy law predates any legacy); Groups + current memberships (bundled grant subjects).
It would not invent: credential columns, provider mirrors, username identity, home-area placement, HR
directory, group hierarchy, membership history engines, tenant partition columns. The inversion
reproduces the candidate's six families **plus Area retirement** — independently confirming both the
shape and F1.

---

# 8. Alternatives (packet §20)

- **A — monolithic User row:** REJECTED — loses the structural-absence erasure property (§5.4).
- **B — generic organization directory:** REJECTED — zero V1 consumers; enterprise-directory YAGNI.
- **C — Keycloak as User/Group authority:** REJECTED — directly violates frozen anti-corruption law
  (provider groups/accounts never product Organization/AuthZ authority; provider independence dies).
- **D — no split, Audit snapshots names:** REJECTED — duplicates PII into immutable evidence, defeating
  the frozen PII-minimized-skeleton requirement and making erasure materially harder.

---

# 9. Subtractive pass (P26)

| Element | Deletable? | Consuming property |
|---|---|---|
| `Tenant.display_name` | No | company identity fact; the concrete `tenant.settings.manage` consumer |
| `Area.code` | No | frozen `{AREA}` numbering + stable scope/business identity |
| `Area.name` | No | human-readable display for governed UX |
| `User.disabled_at` | No | eligibility gate consumed by promoted Session issuance |
| `UserProfile` table | No | structural-absence erasure (§5.4) |
| `UserProfile.email` | No (nullable) | contact/corroboration display for provisioning + admin UX |
| `UserProfile.display_name` | No | human rendering of active Users |
| `Group.name` | No | only human identity of a Group |
| `GroupMembership` explicit table | No | the current-membership truth Authorization/B2-3 consumes |
| Any `created_at` in B2-2 | Already absent | Audit owns creation instants — consistent subtractive posture (asymmetry with B2-1's `created_at` fields is justified there by correlation-establishment/CHECK-anchor consumers) |
| Any uniqueness rule | No | each is a fail-closed identity/duplication backstop (singleton, code, name, pair, one-profile) |

Nothing further is deletable. With F1's one field added, B2-2 is at its subtractive floor.

---

# 10. Proof-obligation results — P1–P26

| P | Result | P | Result |
|---|---|---|---|
| P1 | HOLDS (valid, minimal, falsifiable — §5.1) | P14 | HOLDS (flat/company-wide; scope lives on the grant) |
| P2 | HOLDS (correct boundary split — §5.1) | P15 | HOLDS (no code consumer; UUID is the stable identity) |
| P3 | HOLDS (typed facts only) | P16 | HOLDS (Group deletion reachable; no shipped retirement) |
| P4 | HOLDS (numbering + scope identity) | P17 | RESOLVED: surrogate UUID = YAGNI (§5.8) |
| P5 | **FAILS for lifecycle → F1**; holds for hierarchy | P18 | HOLDS (snapshots + Audit) |
| P6 | HOLDS | P19 | HOLDS (dormancy proof — §5.9) |
| P7 | HOLDS | P20 | HOLDS with L2 (Method-clean deferral) |
| P8 | HOLDS (stable UUID; Audit history) | P21 | HOLDS (no tenant FK has semantic meaning on any B2-2 row) |
| P9 | HOLDS (split is essential — §5.4) | P22 | HOLDS (§5.10) |
| P10 | HOLDS (fallback rendering; UUID refs) | P23 | HOLDS (1:1, no second authority) |
| P11 | HOLDS (no phone/title/username consumer) | P24 | HOLDS except the one F1 miss (§6) |
| P12 | HOLDS (uniqueness would rebuild email authority) | P25 | HOLDS (RESTRICT everywhere cross-owner; within-owner subordinate cascade for Profile/Membership is B1-legal and B2-4 detail) |
| P13 | HOLDS (grants ≠ placement; runtime confirms) | P26 | subtractive floor reached (§9) |

---

# 11. Required outputs

- **Verdict:** `APPROVE R10-B2-2 ORGANIZATION SINGLETON / PEOPLE / GROUPS WITH MATERIAL FIXES`.
- **Counts:** BLOCKER 0 / MAJOR 1 (F1) / LOW 5 (L1–L5).
- **Tenant singleton/enforcement:** APPROVE (constant-expression unique index + fail-closed readiness =
  exactly-one for serving deployments; trigger = accidental complexity).
- **Area identity/lifecycle:** identity APPROVE; lifecycle **MATERIAL FIX F1** — add reversible
  `Area.disabled_at` retirement now (shipped-capability evidence + deletion impossibility).
- **User lifecycle:** APPROVE (reversible `disabled_at`, uniform with promoted law).
- **User/UserProfile split:** **APPROVE** — structural-absence erasure property; monolithic alternative
  rejected with counterexample (§5.4).
- **Email/username:** APPROVE candidate (no identity uniqueness; username absent).
- **User→Area omission:** APPROVE (grants ≠ placement; two-truths defect avoided).
- **Group shape/lifecycle:** APPROVE (flat, name-unique, no code, no lifecycle).
- **GroupMembership:** **surrogate UUID = YAGNI**; pair-keyed current-only + Audit approved (L3 records
  the B1 ruling).
- **Offboarding boundary:** APPROVE deferral (safe; L2 makes the B2-4 policy decision a named gate item).
- **Privacy:** APPROVE (structural absence for PII; User root survives as skeleton anchor; no platform).
- **Subtractive/YAGNI:** at floor after F1; one-miss evidence sweep otherwise clean (§6, §9).
- **Material reopen outside B2-2:** **NONE.** F1's no-new-assignment gates land in the consumers'
  existing design surfaces (CI/B2-3/B4 checklists) as successor obligations, not reopens. No B2-1
  contradiction found (integration facts — eligibility gate, FK target, offboarding revocation — all
  consumed consistently).
- **Exact corrected target:** §2 F1 block (one field + one law); L1/L4 one-sentence additions optional
  at operator discretion.
- **Broad review:** NOT required.
- **Bounded delta review after correction:** SUFFICIENT — verify the amended Area section and its
  consistency with D4/D5 and the consumer gate wording; nothing else moved.

**Convergence:** one round; 1 MAJOR (missing capability, bounded fix) + 5 LOW notes; altitude mechanical.
Findings are evidence; operator adjudication is the next authority gate.
