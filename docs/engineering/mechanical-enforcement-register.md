# Mechanical-Enforcement Register

> **Opened:** 2026-08-06 · **Owner:** whoever is running the current standardization program
> **Status:** standing register — never "closed"; entries close individually when their firing mechanism lands.

## What this register is for

MetalDocs is mid-campaign on one idea: **replace hand-maintained agreement with mechanical
derivation.** The errors module was standardized. The HTTP/API surface became generated from the
spec (ADR 0090/0091). Approval routes were standardized and made config-first. AuthZ is next.

The meta-defect behind all of them was named in the 2026-07-03 final architecture review:
**hand-synced enumerations**. Two or more places that must agree, with nothing forcing them to.
Appendix B cause 2 of the defect-class catalog is the same finding from the defect side.

This register is where that observation gets *durable* instead of being re-derived each program.
Anything spotted in passing — during design, review, debugging, anything — goes here, whether or not
the current task touches it. Recording is cheap; rediscovering is not.

**Scope, widened 2026-08-06 by operator ruling.** Hand-maintained sync was the founding case, not
the boundary. This register takes anything worth standardizing or reconsidering: code, architecture,
logic, implementation — **and build-vs-adopt.** Where we hand-rolled something the industry already
sells hardened and certified (authentication is the archetype — Keycloak, Ory, Zitadel, WorkOS), that
belongs here too. Filing an entry is **not** a proposal to adopt. It is a commitment to study the
option on evidence rather than let "we already wrote it" decide by default.

**The operator's standing rule this register serves** (2026-08-06, restated from CLAUDE.md's
Global Maximum section): *never prefer a solution because it is what is already implemented.* If the
implemented thing is genuinely best, it stays and the entry says so with the reason. Otherwise the
global-maximum structure gets named here, even when it is out of the current boundary.

## Tracking

Every entry is also a GitHub issue, so the register carries the reasoning and the issue tracker
carries the state. The issue links back here; do **not** copy entry text into the issue beyond a
summary and its `file:line` evidence — duplicating it would recreate the exact hand-synced-copies
defect this register exists to eliminate.

| Entry | Issue |
|---|---|
| ME-01 role vocabulary on six surfaces | [#74](https://github.com/leandrotcawork/MetalDocs/issues/74) |
| ME-02 REQ-AUTHZ-5 covers the wrong noun | [#75](https://github.com/leandrotcawork/MetalDocs/issues/75) |
| ME-03 document-list visibility binary | [#76](https://github.com/leandrotcawork/MetalDocs/issues/76) |
| ME-04 params → filter fields, no completeness guard | [#77](https://github.com/leandrotcawork/MetalDocs/issues/77) |
| ME-05 are the tripwire arms generated? | [#78](https://github.com/leandrotcawork/MetalDocs/issues/78) |
| ME-06 per-document sharing retires RBAC-with-scope | [#79](https://github.com/leandrotcawork/MetalDocs/issues/79) |
| ME-07 authentication build-vs-adopt | [#80](https://github.com/leandrotcawork/MetalDocs/issues/80) |
| ME-08 MFA is a dashboard, not a control | [#81](https://github.com/leandrotcawork/MetalDocs/issues/81) |
| ME-09 a kill list is a hand-synced enumeration | [#82](https://github.com/leandrotcawork/MetalDocs/issues/82) |
| ME-10 tenant-defined roles retire the generated enums | [#83](https://github.com/leandrotcawork/MetalDocs/issues/83) |

## How to read an entry

Every entry must carry a **firing mechanism** — the concrete thing that makes drift a red build, a
boot fatal, or an unrepresentable state. An entry without one is a wish, not a plan.

Preference order for firing mechanisms, strongest first. This ordering is doctrine
([[no-fallback-principle]], "prefer unrepresentable over guarded"), not taste:

1. **Unrepresentable** — the drift cannot be expressed. One declaration, everything else derived
   from it (FK instead of repeated CHECK; one function with a parameter instead of two functions).
2. **Boot fatal** — the process refuses to start. `assertSurface` (ADR 0091) is the template.
3. **Red build** — a lint or generator-diff test fails in CI.
4. **Runtime assertion** — fails closed at the point of use.
5. *(not acceptable alone)* A doc, a comment, or a checklist.

| Field | Meaning |
|---|---|
| **Surfaces** | The places that must currently agree, with `file:line`. |
| **Kept correct by** | What holds it together *today* — usually "discipline". |
| **Firing mechanism** | What should hold it, per the order above. |
| **Verified** | Whether the claim was read first-hand, and when. |

---

## Open entries

### ME-01 — Role vocabulary declared on six surfaces

**Surfaces**
| Surface | Roles |
|---|---|
| `internal/platform/iamtypes/role.go:39` `validRoles` | 8 |
| `internal/platform/iamtypes/role.go:75` `areaRoles` | 7 |
| `api/openapi/v1/openapi.yaml:4836` `UserRole` | 8 |
| `api/openapi/v1/openapi.yaml:4845` `AreaRole` | 7 |
| `db/baseline/0001_current_schema.sql:1662` `user_process_areas` CHECK | 7 |
| `db/baseline/0001_current_schema.sql:1276` `iam_user_roles` CHECK | **5** |
| `db/baseline/0001_current_schema.sql:1245` `iam_group_roles.role` | **no constraint** |

**Kept correct by** discipline, and it already failed. Four surfaces agree in two consistent pairs.
The `iam_user_roles` CHECK of 5 is mirrored by no Go set and no OpenAPI enum, and the three roles it
omits — `area_admin`, `qms_admin`, `signer` — are the three the rest of the system treats as
first-class. That unmirrored table is the one tier-1 reads.

**Firing mechanism** — level 1. A `role_catalog` table referenced by FK; the Go sets and OpenAPI
enums generated from it. Adding a role becomes an INSERT, not a seven-place edit.

**Verified** 2026-08-06, first-hand. **Owner:** the authz grant-unification program (in design).

---

### ME-02 — REQ-AUTHZ-5's CI binding covers capabilities, not roles

REQ-AUTHZ-5 (`wiki/architecture/backend-target-architecture.md:162-169`) requires the declaration
surfaces be CI-bound so drift is a red build. That guard exists and works — **for capabilities.**
No guard compares the seven role surfaces in ME-01, which is why ME-01 survived a requirement
written to catch exactly its shape.

**This is the most instructive entry in the register.** A firing mechanism that covers the wrong
noun reads, in every audit, exactly like one that works. When adding a guard, write down what it
does *not* cover, in the guard.

**Firing mechanism** — level 3, extending the existing surface-coherence lint to the role catalog.
Subsumed by ME-01's level-1 fix if that lands first; keep this entry open until one of them does.

**Verified** 2026-08-06. **Owner:** authz grant-unification program.

---

### ME-03 — Document list visibility is a hand-written binary, not a permission query

`internal/modules/documents/delivery/http/handler.go:435-439` is the entire visibility model for
listing documents:

```go
if !isAdmin && callerUserID != "" {
    opts.CreatedBy = callerUserID
}
```

Either the caller is `system_admin` and sees everything, or sees only what they created. An area
`viewer` holding a grant in `user_process_areas` does **not** see that area's documents.

This is the disjoint-grant-tables defect surfacing as a product bug rather than a permission bug:
the list can only ask the question tier-1 can answer, and tier-1 cannot see the area model. The port
that answers it correctly already exists and is unused here — `AreaCapabilityReader`
(`internal/modules/iam/domain/area_capability_reader_port.go:19`) returns
`(tenant_wide bool, areas []string)` for a capability.

**Firing mechanism** — level 1: delete `IsSystemAdmin` entirely so "sees everything" is not
expressible as a boolean, only as "holds `document.read` at tenant scope". Ratified 2026-08-06.

**Verified** 2026-08-06, first-hand. **Owner:** authz grant-unification program.

---

### ME-04 — Request params → domain filter fields have no completeness guard

The search `Query` struct carried four fields (`Subject`, `BusinessUnit`, `Classification`, `Tag`)
that the handler populated and no reader ever consumed. A caller passing `classification=CONFIDENTIAL`
got unfiltered results and **no error** — a silent no-op on a field whose name implies access
control. Deleted 2026-08-06 (`6425b21a`), but the class is open: nothing prevents the next one.

Two independent gaps, and they need different mechanisms:
- a declared parameter the query layer never reads (silent no-op);
- a domain filter field nothing populates (dead weight that reads as capability).

**Firing mechanism** — level 3. A lint pairing declared query parameters against the fields the SQL
builder actually consumes. The generated-params direction (level 1) is the better answer if the
codegen can be pushed that far — worth a spike before settling for the lint.

**Verified** 2026-08-06, first-hand. **Owner:** unassigned.

---

### ME-05 — Confirm the tripwire CASE arms are generated, not hand-written

`public.enforce_capability_asserted()` (`db/baseline/0001_current_schema.sql:304-515`) is a large
per-table `CASE` mapping tables to required capabilities. GMR M2 recorded that tripwire arms became
generated from the Go capability registry with two blocking drift/parity lints — so this is very
likely already mechanized and the baseline is generator *output*.

**Not verified.** Recorded as a question, not a finding. If it is generated, close this entry and
note it in ME-02 as a positive example of a correctly-scoped guard. If any arm is hand-maintained,
it becomes a level-1/level-3 entry in its own right.

**Owner:** unassigned. Cheap to settle — find the generator or fail to.

---

### ME-06 — Watch for per-document sharing; it is the trigger that retires RBAC-with-scope

Not a defect. A recorded **decision trigger**, so the choice gets revisited on evidence instead of
inertia.

The authz model (capability bundles + scope on the binding) is NIST-RBAC-shaped and was chosen over
a policy engine (Cedar, OPA) and over Zanzibar/OpenFGA/SpiceDB on merit: MetalDocs' policy has no
attribute conditions and no relationship traversal, so those engines buy unused expressiveness while
charging a second runtime, a policy language, and — decisively — a **synchronization problem between
the grants in Postgres and the engine's own data**, which is the Class 35 defect re-created at a new
boundary.

**The trigger that flips this:** a requirement for per-document or per-object sharing ("share this
SOP with these three people"). RBAC-with-scope cannot express it, and the workaround — a role per
document — is the direct road back to the defect this program removes. On that day, a relation-tuple
engine becomes the global maximum and this entry becomes a program.

**Verified** as a decision, 2026-08-06. **Owner:** whoever gets that requirement first.

---

## Build vs. adopt

A different kind of entry, so it gets its own section and its own questions. A firing mechanism is
not the right field here — these are not drift problems. The questions that decide them:

1. **Commodity or domain?** If every serious product in the category implements the same thing the
   same way, it is commodity and owning it earns nothing. If the logic encodes *this product's*
   rules, it is domain and must stay ours.
2. **Does buying shift a compliance burden?** In a regulated eQMS, a vendor-validated component can
   move part of the validation evidence off our side. That is a real, costed argument, not a vague
   one.
3. **What does integration cost, and does it re-create a sync problem?** An external system holding
   a copy of data we also hold is Class 35 at a new boundary. This is what killed the policy-engine
   option in ME-06, and it is the first thing to check for any adoption.
4. **What is the exit?** Adopting is also a coupling. An option we cannot leave is a bigger decision
   than an option we can.

---

### ME-07 — `auth` is a hand-rolled identity provider (~4,378 LOC)

**What we own today** (`internal/modules/auth/application/service.go`): password hashing (bcrypt,
cost 12), credential verification, a constant-time login path with a dummy-hash timing-oracle
mitigation (`:127`, `:273-300`), session token minting + signing (`:1110-1142`), session cookies,
account lockout and unlock (`:840`), password policy (`:1013`), admin password reset (`:743`),
local-admin bootstrap (`:198`), plus `pre_auth_login_rate_limit` in the middleware chain.

**This entry is not a criticism of the code.** It is careful work — the timing-oracle handling is
better than most hand-rolled auth, bcrypt cost 12 is right, and the failure paths are deliberate.
The question is not quality. It is **category**.

**The line worth drawing:** authentication is commodity; authorization is domain. MetalDocs' capability
model, area scoping, and separation-of-duties encode *this product's* rules and must stay ours
(ME-06 covers why even the authZ engine stays in-house). Proving someone is who they say they are
encodes nothing of ours.

**Arguments to study, not conclusions:**
- **21 CFR Part 11** puts identity controls (§11.10(d),(g),(h), §11.300) in audit scope. A
  vendor-validated IdP moves part of that evidence burden off us — question 2, and the strongest
  argument on the adopt side.
- **SSO / SAML / SCIM is table stakes** for pharma and medical-device buyers. Today that is a
  from-scratch build; with an IdP it is configuration. This is a *feature* argument, not only an
  architecture one, which makes it the one most likely to arrive as a customer requirement.
- **Against:** integration is not free, and question 3 bites — an external IdP holding users while
  `iam_users` also holds them is a sync boundary. It is tractable (the IdP owns the credential, we
  own the profile and the grants, joined by a stable subject id) but it must be designed, not
  assumed.
- **Exit** (question 4) is genuinely good here: OIDC is a standard, so the coupling is to a protocol
  rather than to a vendor.

**Related and blocking-adjacent:** ME-08. Whatever is decided, MFA is currently a hole, and an IdP
would close it as a side effect.

**Verified** 2026-08-06, first-hand. **Owner:** unassigned — study, not scheduled.

---

### ME-08 — MFA is a dashboard, not a control

**Severity: this is the most serious entry in the register.**

`security` publishes per-tenant and per-role MFA coverage
(`internal/modules/security/delivery/http/handler.go:112`, `domain/model.go:13-23`), reading through
`iam`'s `MfaUserReader` port (`internal/modules/iam/domain/mfa_user_reader_port.go:16-22`). That
port has exactly two operations: `TenantMfaCounts` and `TenantMfaCountsByRole`. Both are counts.

There is **no enrollment path, no factor verification, and no MFA challenge in the login flow**
(`auth/application/service.go:265` `Authenticate` has no MFA step). No TOTP, no WebAuthn, anywhere in
the repo. `iam_users.mfa_enabled` and `mfa_enrolled_at` are columns nothing can set to true except
hand-written SQL.

**Why this is worse than a missing feature.** An absent control is visibly absent. A control that
reports its own coverage looks *present and measured* — the dashboard renders, the percentage is
real arithmetic over a real column, and an auditor reading "MFA coverage: 0%" concludes rollout is
incomplete, not that the mechanism does not exist. The reporting layer lends credibility the
enforcement layer never earned.

**Firing mechanism** — level 1 for the reporting half, and it is available immediately regardless of
what happens to the control itself: an endpoint must not be able to report on a control with no
enforcement path. Either MFA gains enrollment + challenge, or the coverage endpoint is deleted until
it does. Reporting a hollow control is a defect; the honest interim state is silence, not a zero.

**Candidate defect class** — this generalizes past MFA and past this repo: *observability for a
control that was never implemented*. Nothing in the existing catalog covers it, and the mechanism is
distinct from Class 12 (debt written up as policy) because no one ever claimed the control worked —
the dashboard makes the claim structurally, without a sentence. Evaluate for Part I on the next
catalog pass.

**Verified** 2026-08-06, first-hand: read the port, the repository, the handler, and the login path.
**Owner:** unassigned. **Do not let this one sit unrouted.**
**Filed:** [#81](https://github.com/leandrotcawork/MetalDocs/issues/81).

---

## ME-09 — a kill list is a hand-synced enumeration

**Found** 2026-08-07, by the independent advisory review of DD-1
(`docs/superpowers/analysis/2026-08-07-authz-advisory-opinion.md`).

DD-1 ruled that "all three `system_admin` bypasses die" and listed them. The advisor found **five**:
the two unnamed ones are `UserAreaRepository.MembershipDirectoryScope`
(`internal/modules/iam/infrastructure/postgres/user_area_repository.go:159`, which embeds
`SystemAdminExistsSQL` as its tenant-wide branch) and `CapsByUserID`
(`internal/modules/iam/application/capability_service.go:132-139`, which returns `AllCapabilities()`
on the admin branch — the `/auth/me` surface, so leaving it makes the frontend render capabilities by
a different rule than enforcement uses).

**The generalizable finding is not the two misses.** It is that a **removal inventory is the same
artifact class as the duplication it removes**: a hand-written enumeration of "every place this fact
lives", produced by reading, believed because it is written down. Every argument in this register
against hand-synced declaration surfaces applies verbatim to the list that claims to have found them
all — and this one was wrong on its first outing, by 40%, while reading as complete.

The failure mode is worse than ordinary drift because a kill list is consumed **once**, at the moment
it is believed, and then discarded. A drifting declaration surface at least stays around to be caught
later. An incomplete kill list closes the program.

**Firing mechanism** — level 3, and it is the criterion of completion rather than a check on it:
after the program, a lint asserting the literal `'system_admin'` appears in no production Go or SQL
outside the role-catalog seed. `SOLE-RLS` is the in-house template for exactly this shape. That
converts "we think we got them all" from a claim into a red build.

**Generalized rule:** when a program's charter is *deleting* a construct, the definition of done is a
mechanism that makes the construct unwritable — never an inventory of its known instances.

**Owner:** the authz grant-unification program (in design).

---

## ME-10 — watch for tenant-defined roles; they retire the generated role enums

**Recorded** 2026-08-07 by operator ruling DD-6. **Not a defect** — a decision trigger, in the ME-06
mold, so the choice is revisited on evidence instead of inertia.

The role vocabulary is **product reference data**: `role_catalog` is the single upstream, the OpenAPI
`UserRole` / `AreaRole` enums are generated from it, and adding a role is an INSERT plus a spec
regen, an FE regen and a release. Chosen on merit — in a validated eQMS each tenant-defined role is
un-validated configuration inside the customer's own CSV package, and the frontend's exhaustiveness
proof (see the positive template below) only exists because the vocabulary is closed.

**The trigger that flips this:** the first customer requiring their own role vocabulary. At that
point the industry shape becomes the global maximum — one type, `role_code` a string FK to the
catalog, product roles merely seeded rows. Kubernetes does not distinguish `cluster-admin` from a
ClusterRole you wrote; Keycloak realm roles are all data; AWS managed and customer-managed policies
are the same object. The enum in the contract is the piece none of them has.

**The migration cost, recorded now so it is not discovered later:** `role_code` leaves the enum space,
`role-vocabulary.ts`'s `satisfies Record<UserRole, …>` guard stops proving anything, the FE must fetch
the catalog at runtime, PT-BR labels move from code to data, and a role-composition admin UI (compose
a role from capabilities) becomes a required product feature.

**Related correction, filed here because it is the kind of claim that ages badly:** D3 argument 5
justified rejecting direct per-user grants partly on "a new role is an INSERT, not a migration." No
DDL, but a release — the escape valve for a genuine one-off is narrower than the phrasing implies.
D3's case stands on auditability and reviewability, which are untouched.

---

## Positive template — what a landed firing mechanism looks like

Recorded because "what good looks like" is easier to copy than to describe, and because the register
should not read as a list of failures only.

**`assertSurface`** (`apps/api/cmd/metaldocs-api/surface.go:25-30`, ADR 0091) — four boot checks
proving the mounted HTTP surface is exactly the declared one, aggregated into one error so a failure
is fixable in one restart, ending in `os.Exit(1)`. Level 2. Two properties worth copying:

1. **Ownership is evaluated per-publisher, not against the global union.** A global-set check would
   let every publisher claim a distinct tag while silently mounting each other's routes, and all the
   other checks would still pass. The mechanism was designed against the way it could be fooled, not
   only against the way it should work.
2. **Its ADR states what it does not claim** — that the surface is *exactly the declared one*, not
   that any handler is correct. A guard whose scope is written down cannot be mistaken for a
   stronger one later. ME-02 is what happens when that is left implicit.

**`role-vocabulary.ts`** (`frontend/apps/web/src/lib/iam/role-vocabulary.ts`, F-QA4-2) — the frontend
role vocabulary, single-sourced from the generated contract types, with every runtime list proved
against them:

```ts
const ROLE_LABELS = { … } as const satisfies Record<UserRole, string>;
```

Miss a role and `tsc` reports the missing property; invent one and `tsc` rejects the excess property.
Level 3, and 18 files depend on it. Two properties worth copying:

1. **It solves the type-only problem instead of working around it.** Generated OpenAPI enums are
   type-only, so *some* runtime array has to be written by hand — the design accepts that one literal
   and makes it undriftable, rather than pretending the hand-written list does not exist.
2. **It is the reason ME-10 has a cost.** A guard this good is an argument in the design record:
   DD-6 kept the role vocabulary closed partly because opening it degrades `UserRole` to `string` and
   this proof silently evaluates to nothing. Knowing what a mechanism buys is what makes trading it
   away a decision rather than an accident.
