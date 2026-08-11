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
| ME-11 anti-drift guard with a hand-synced allow-list | [#84](https://github.com/leandrotcawork/MetalDocs/issues/84) |
| ME-12 a string lint cannot see a constant-built bypass | [#85](https://github.com/leandrotcawork/MetalDocs/issues/85) |
| ME-13 an analysis that takes its subject as its own premise | [#86](https://github.com/leandrotcawork/MetalDocs/issues/86) |
| ME-14 a tenant table with no RLS, and no control that could see it | _issue pending_ |
| ME-15 the check registry is four hand-synced inventories, not one | _issue pending_ |
| ME-16 compose healthcheck probes are a second, unforced reading of two worker/jobs runtime facts | [#115](https://github.com/developmentconexus-ops/MetalDocs/issues/115) |
| ME-19 the negative fixture is the guard's guard, and nothing guards it | _issue pending_ |
| ME-20 a gate signal written by a later phase than the destructive one | _issue pending_ |

> **ME-17 and ME-18 are intentionally absent from this table on `main`.** They exist on the open
> `feat/a8-1-capability-bindings` branch (PR #113) and land with it. ME-19/ME-20 were allocated
> above that branch's high-water mark rather than reusing the next id readable from `main`, which
> was already taken — see "Note on allocating an entry id" below. The gap is a live claim about
> another branch, not a numbering error; do not "fix" it by renumbering.

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

## How to build a level-3 guard

The ordering above says *which* firing mechanism to prefer. It does not say how to build the one
you settle for. Level 3 is where most entries land in practice, and the 2026-08-11 remediation wave
is the evidence for why that needs its own bar: eight-plus confirmed instances of "the mechanism is
absent and nothing says so" were closed by writing guards, and then **the guards were found
defective in the same way, four times in one wave** — a Dockerfile guard blind to `FROM $VAR`, an
import guard blind to a single-line `import`, the same guard blind to an aliased import, and the
same guard again blind to a dot-import. ME-12 already named this as the wrong-noun failure and
called it "the default outcome" rather than a mistake to learn from. The wave confirmed that
prediction under authors who had read ME-12.

A guard is code, and it is code that nothing else guards. Three requirements, each earned live:

1. **Fail closed on what it cannot parse.** Most blind spots are not wrong answers; they are
   "could not see it, therefore passed". Ruled on PR #114: when the Dockerfile guard could not
   parse a multi-line `FROM` continuation, the correct fix was *not* to build a line-joiner in
   bash — more code than the check, its own untested edges, and it starts guessing. The fix is an
   explicit unparseable branch that exits non-zero and says which shape it refused. A guard with no
   "I do not know" branch answers "no violation" to every question it cannot understand.

2. **Ask semantics, not syntax, wherever a resolver exists.** The four blind spots above are one
   defect: a *syntactic* guard answering a *semantic* question. PR #120 is the template for the fix
   — the read-only-transaction rule stopped matching AST shapes and started resolving identifiers
   through `go/types`, asking whether a name *is* `database/sql.TxOptions` rather than whether it is
   spelled like one. That form is simultaneously blind to nothing (qualified, aliased, and
   dot-imported all resolve identically) and free of false positives (a locally shadowed
   `TxOptions` resolves to the local type and is correctly ignored) — which a syntactic guard cannot
   be at the same time. Where the artifact is buildable, prefer asking the built artifact over
   reading its source.

3. **Every exclusion needs `NotWant`.** A guard's scope exclusion is itself a rule, and `Want`
   cannot assert it: `Want` proves a guard FIRES, never that it stays SILENT. Delete an exclusion
   and the output only *gains* a line — every `Want` still matches and the harness reports ok. Put
   the excluded thing in the fixture tree in a form that WOULD fire, and name it in
   `Fixture.NotWant` (`tools/verify/fixtures.go`), so deleting the exclusion turns the harness red.

And the layer under all three: **the negative fixture is the guard's guard, and it is the least
guarded thing in the tree** — see ME-19.

### Note on allocating an entry id

ME ids are allocated by hand, and concurrent branches allocate them blind. Found live 2026-08-11:
`main` was at ME-16 while an open PR's branch already carried ME-17 and ME-18, so the next id
readable from `main` was already taken. **Before adding an entry, scan every open branch**, not
just `main`:

```bash
for b in $(git branch -r --format='%(refname:short)'); do
  git show "$b:docs/engineering/mechanical-enforcement-register.md" 2>/dev/null \
    | grep -oE '^## ME-[0-9]+' | tail -1 | sed "s|^|$b |"
done
```

Yes — this register's own ids are a hand-synced enumeration. It is recorded here rather than
quietly worked around, because the alternative is the exact defect the register exists to name.

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

## ME-11 — an anti-drift guard carrying a hand-synced allow-list

**Found** 2026-08-07 while auditing what `role_capabilities` really is.

`TestEveryCapSeededOrDeferred` (`apps/api/cmd/metaldocs-api/permissions_test.go:597`) is a genuine
level-3 mechanism: every registry capability must be seeded to ≥1 role, or sit in a `deferred`
allow-list with a documented reason. It works. It is also carrying, at `:609`, this comment:

> `// Mirrors deferredCaps in scripts/api-lint/registry_rules.go.`

A hand-synced enumeration **inside the mechanism whose purpose is preventing hand-synced
enumerations.** The two lists can disagree, and when they do, one of them silently stops guarding.
Prose is not a sync mechanism; `// keep in sync with` is scheduled drift.

**Firing mechanism** — level 1: one declaration, imported by both consumers. The allow-list is
program data, not test data; it belongs beside the capability registry it annotates, and the test and
the lint both read it. Deleting the second copy beats testing that the copies agree.

**Worth noting for the doctrine, not just the fix:** this guard is otherwise well built and was found
only by reading it closely for an unrelated reason. Mechanisms are not audited the way product code
is — nobody diffs a lint against its own principles. That is a gap in review practice, not in this
lint.

**Owner:** unassigned. Small.

---

## ME-12 — a lint on a string cannot see a bypass built from a constant

**Found** 2026-08-07 by the second-pass advisory audit of the authz design spec, testing F3 against
ME-02's own instruction.

The authz program's definition of done is an extinction lint: the literal `'system_admin'` must not
appear in production Go or SQL outside the role-catalog seed (ME-09). But the same program generates
Go role constants from `role_catalog`, so generated code legitimately contains that literal and the
lint must allowlist generated files. From that moment:

```go
if role == iamtypes.RoleSystemAdmin { /* ... */ }
```

is a brand-new bypass, with zero literals, and the lint stays green.

**The lint's noun is the string. The defect's noun is role-identity special-casing in enforcement
paths.** Different nouns — ME-02 exactly, discovered inside the mechanism written *because of* ME-02.
The lesson is that the wrong-noun failure is not a one-time mistake to learn from; it is the default
outcome whenever a mechanism targets a syntactic proxy for a semantic property, and it recurs even
under authors who know the rule.

**Firing mechanism** — level 3, and it is the harder, correct one: enforcement packages may consume
role identity **only through the evaluator**. Comparing a `role_code` value against any specific role
outside the evaluator, the seed, and generated code is the lint target. This names the semantic
property instead of a spelling of it.

**Interim, already landed in the design:** F3 claims literal extinction *only*, and the residue is
written into the guard and into the spec's accepted-residue section (§14 item 2) rather than left to
read as full coverage.

**Owner:** follow-up to the authz grant-unification program.

---

## ME-13 — an analysis that takes its subject as its own premise

**Found** 2026-08-07 by the operator, rejecting two independent advisory answers to the question
"is the `documents`/`controlleddocuments`/`templates` split a domain truth or an implementation
accident?"

Both arms answered with evidence about the system as built: table ownership is disjoint, no
`POST /documents` route exists, a prior design already ratified "no merge", the current transaction
creates the slot and the first document together. Each of those is a **consequence** of the split
being examined. The operator's objection, verbatim:

> "você está baseando num máximo local — baseando numa coisa que já está implementada em vez de
> analisar o melhor pra ser implementado."

An argument of that shape cannot return any answer but "leave it as it is", because it takes the
artifact under examination as its own premise. It is not a weak argument for the status quo; it is
**no argument at all**, wearing the costume of evidence — and it reads as rigorous precisely because
it cites `file:line`.

Two properties make this worse than ordinary bias. First, **it is selective**: the same analyst
reasons correctly about everything except the thing being judged, so the surrounding rigour vouches
for the one section that has none. Second, **the failure is invisible from inside the answer** —
nothing in "the tables are disjoint" announces that it is circular. It took an operator who was not
reading the evidence to catch it.

CLAUDE.md's Global-Maximum rule already names this. It fired here anyway, inside two analyses
commissioned specifically to avoid it, and it was the *rejected* answer that was later shown wrong on
the merits: re-asked with status-quo evidence ruled inadmissible, both arms reversed and converged on
the opposite conclusion (ADR 0093).

**Firing mechanism** — level 3, and it belongs to the brief, not to the analyst. Any brief asking
whether an existing structure is correct MUST carry an explicit inadmissible-evidence list (current
schema, module layout, import graph, route topology, existing ADRs, transaction boundaries,
doc-comments describing the code, migration cost) and MUST require an **inversion test**: state which
conclusions would survive if the current implementation were the opposite in every respect. An answer
without a passed inversion test is not a finding. `docs/superpowers/analysis/2026-08-07-controlled-information-greenfield-brief.md`
is the working template; the `adversarial-review` and `developing-new-work` skills are where the
requirement has to live to fire without being remembered.

**Owner:** unrouted. The skills change is small and blocks nothing; the risk is that every future
"should we restructure X" answer inherits the defect until it lands.

---

## ME-14 — a tenant table with no RLS, and no control that could have seen it

**Found** 2026-08-08 during the CI Phase-0 baseline work, by hand, while a *documentation* task
(the `db-docs-coverage` gap closure) read `db/baseline/0001_current_schema.sql` and refused to assume a
table matched its siblings.

**Surfaces**
| Surface | Claim |
|---|---|
| `db/baseline/0001_current_schema.sql:2084` `approval_route_stage_selectors` | `tenant_id uuid NOT NULL`, **no `ENABLE`, no `FORCE ROW LEVEL SECURITY`** |
| same file, sibling approval tables `:2025, 2055, 2077, 2144, 2171` | `FORCE ROW LEVEL SECURITY` |
| same file `:4807-4831` | the `ENABLE ROW LEVEL SECURITY` block — 37 tables |
| `internal/modules/iam/tenancy/tenant_data_port.go` | the table is wired into export **and** erase — the GDPR surface |

Measured first-hand against the baseline on 2026-08-08: **38 tables declare `tenant_id`; 37 enable
RLS.** The set difference is exactly one table, and the 37 are a strict subset of the 38. That
singleton status is what makes it credible as an oversight rather than a design choice — every other
RLS-less table in the schema inherits isolation through its FK chain and carries no `tenant_id` of
its own.

**Kept correct by** nothing. This is the part worth recording. The repo has a real, well-built RLS
control stack, and **not one piece of it could have caught this**:

- `scripts/api-lint/async_tenant_tables_schema_drift_test.go:48` compares
  `async-tenant-tables.txt` against the baseline's FORCE-RLS set. Both sides are derived from tables
  that *already have* RLS, so a table with none is absent from both and the test is green — a
  consistent, complete-looking agreement about an incomplete world.
- `tests/integration/security/rls_truth_test.go` proves enforcement is genuine for a table that has
  a policy (it killed the `metaldocs_app` false green, M7 F7.4). It says nothing about which tables
  should have one.
- `scripts/api-lint/sole_rls_read_rule.go` governs reads that rely on RLS as the sole isolation
  mechanism — again, downstream of RLS existing.
- `tools/verify/registry.go` — all 29 checks — contains **no RLS check of any kind** (`grep -i rls`
  returns nothing).

The M7 RLS-truth sweep missed it for the same structural reason: it reasoned about FORCE-vs-ENABLE
on tables already known to be RLS-enabled, not about tenant-scoped tables with no RLS at all. Every
control in the stack takes "the table has RLS" as its starting point. **The question none of them
asks is the one that fails.**

**Firing mechanism** — level 3, and the cheapest one in this register. A generator-diff test
asserting *set equality* between `{tables with a tenant_id column}` and `{tables with ENABLE ROW
LEVEL SECURITY}` in the baseline, registered in `tools/verify/registry.go` so it runs in `pr` and
`full`. Both sides come from the schema; neither is hand-maintained; new tenant tables are covered
the day they land. Deliberate exceptions go in a checked-in allow-list **with a reason per line**,
the same shape as `async-tenant-tables.txt`. Today that list would hold zero entries.

Level 1 is reachable and better if the schema ever moves to generated DDL: emit the `ENABLE`/`FORCE`
lines and the `tenant_isolation` policy *from* the `tenant_id` column declaration, so an RLS-less
tenant table is unrepresentable. Out of the current boundary; named here so the level-3 test is
understood as a step, not a destination.

**Verified** 2026-08-08, first-hand against `db/baseline/0001_current_schema.sql` and against the
four controls named above.

**Owner:** unrouted. Two pieces of work, and they are separable — the missing RLS on
`approval_route_stage_selectors` (a migration, plus a decision on whether its FK chain already makes
the exposure theoretical), and the drift test that makes the *class* impossible. The second is worth
more than the first: the table is one instance, and the absent control is why there could be others.

---

## ME-15 — the check registry is four hand-synced inventories, not one

**Found** 2026-08-09 during Task 13 of the CI restructure (§8.2 renames + ruleset re-export),
recorded per spec §9's own instruction to name this design's most likely lie: *"the gate enforces
every check."*

**Surfaces** — four places that must agree, by hand, for a registered check to actually gate a merge:

| Surface | What it claims |
|---|---|
| `tools/verify/registry.go` | the check exists, with an `ID`, a `Profiles` set, and a `CIJob` |
| each workflow job's `--only=`/`--profile=` invocation (`ci.yml`, `nightly.yml`, `docx-renderer.yml`) | which registered checks actually run, and in which job |
| `ci.yml`'s `required` job `needs:` list (`verify, test-integration, security, lint-go`), validated by `scripts/required-gate.jq` / the `required-gate-selftest` registry check | which of those jobs must succeed before `required` — the sole required ruleset context — reports success |
| the diff surface each check's `Paths` matches against under `--profile=changed` (`matchesPaths` in `tools/verify`) | whether a given PR's diff actually selected the check at all |

**Kept correct by** `--audit` (registry rules A1–A9) and `required-gate.jq`'s exact-set-equality
assertion — both real, both load-bearing, and both proven in this task's own verification run. But
per the ordering in "How to read an entry" above, that is level 3 (red build), not level 1
(unrepresentable): a fifth inventory — a check registered, wired into a profile, matched by a
job's `--only=`, inside `required`'s closure, but silently never selected because its `Paths` never
matches the diff that actually needs it — is exactly the shape A1–A9 were built to catch, and the
existence of a catching lint is not the same claim as the drift being impossible. `--audit` and the
set-equality guard make drift *visible*; they do not make it *impossible* (spec §9, quoted verbatim
in the Task 13 brief).

**Firing mechanism today** — level 3: `go run ./tools/verify --audit` (registry rules A1–A9),
red build in `ci.yml:verify`'s `--profile=changed` invocation on every PR. `required-gate-selftest`
pins the fourth inventory's aggregator (`scripts/required-gate.jq`) down the same way.

Extended 2026-08-09 by #87/A1 (Phase 1) with three rules covering gaps this entry named but did
not close: **A7** — a `pr`-profile check must declare either a negative fixture or a written
waiver, so "the guard is blocking" and "the guard has been observed to fail" stop being separate
claims; **A8** — duplicate check IDs are rejected (a duplicated ID silently shadows the original,
observed once during that work); **A9** — everything CI executes names an immutable version
(workflow `uses:` SHA-pinned, check `Argv` never `@latest`), because an unpinned tool changes what
the gate accepts with no diff to review. The fixture spine itself runs as the `guard-fixtures`
check in `ci.yml:verify`.

The same work moved the second inventory one rung closer to level 1: job routing is no longer
asserted twice. A workflow job passes `--ci-job=<file.yml:job>` and the verifier selects the checks
whose registry `CIJob` names that job, so "which job runs this check" is read from the registry
instead of being restated in the workflow and compared afterwards. The workflow still names its own
job (that string can still be wrong), so this is level 2, not level 1 — but a check can no longer
be run by a job that does not own it, which is how the first CI run of #87/A1 failed
(`ci.yml:verify` selected `golangci-lint`, whose binary only `ci.yml:lint-go` installs).

**Global-maximum structure** — one generated CI manifest that *owns* registry membership, job
routing (which job runs which check, via which `--only=`), and `required`'s gate dependency
(`needs:` closure), with the workflow YAML **generated from it** rather than hand-authored and
then audited for agreement after the fact. That converts all four inventories above into one
declaration with three projections, the same "unrepresentable beats guarded" move ME-01, ME-05,
and ME-11 already name for their own surfaces (no-fallback-principle doctrine, cited there).

**Follow-on milestone:** scheduled on `docs/superpowers/ROADMAP.md` §4, row **4.7 "generated CI
manifest"** (added 2026-08-09), queued after this CI-restructure program merges. Working name:
**"generated CI manifest"** — one manifest owning registry membership, job routing (`--only`
lists), the `required`-gate `needs`/jq set, and the workflow YAML generated from it; deletes this
entry on landing. Under this repository's Global Maximum rule (CLAUDE.md), shipping §4's rename +
re-export without this label would itself be the defect the rule exists to catch.

**Owner:** unrouted. Cheap to defer, expensive to forget — record here rather than let the next
`--audit` false-confidence read stand unqualified.

---

## ME-16 — compose healthcheck probes are a second, unforced reading of two worker/jobs runtime facts

**Found** 2026-08-11, round 5 of A7.1 (issue #95, PR #109), disposing of CodeRabbit review threads
against `deploy/compose/docker-compose.yml`. Two independent instances of the same shape, both
about the `worker`/`jobs` `healthcheck:` blocks encoding a runtime fact that a different part of the
same file (or the same binary) is the actual source of truth for.

**Surface A — the probed address.** `apps/worker/cmd/metaldocs-worker/infraserver.go:34` and
`apps/jobs/cmd/metaldocs-jobs/infraserver.go:53` read `WORKER_METRICS_ADDR` / `JOBS_METRICS_ADDR`
(via `config.LoadListenAddr`, defaulting to `:9091`/`:9092`) — the process's real, overridable
listen address. `deploy/compose/docker-compose.yml:344` and `:392` hard-code
`127.0.0.1:9091`/`:9092` in the healthcheck `test:` shell command, a second reading of the same
fact. They agree today only because neither service's `environment:` block currently sets
`WORKER_METRICS_ADDR`/`JOBS_METRICS_ADDR` (confirmed by reading both blocks in full,
2026-08-11) — an absence of override, not a guarantee against one. The moment either variable is
added to `environment:` without a matching edit to `healthcheck.test:`, `docker compose ps` starts
reporting `unhealthy` for a correctly-running process.

**Surface B — is `METALDOCS_WORKER_RUN_ONCE` a one-shot batch invocation, or a service `restart:
unless-stopped` should keep alive?** `apps/worker/cmd/metaldocs-worker/main.go:184-188` (`runWorkerBatch`)
calls `os.Exit` once the batch drains — by design, a single completed run, per the F7/F9 comments
already on `docker-compose.yml:295-344`. But `deploy/compose/docker-compose.yml:249` sets
`restart: unless-stopped` on that same `worker` service, and Compose's restart policy reacts to
**process exit**, independent of and unrelated to healthcheck status (confirmed against
`docs.docker.com/reference/compose-file/services` and `docs.docker.com/engine/containers/start-containers-automatically`:
`unless-stopped` restarts on any exit code, success or failure, unless the container was manually
stopped). The round-4 healthcheck case-guard at `docker-compose.yml:344` (`case "$$v" in
[Tt][Rr][Uu][Ee]|1) exit 0 ;; esac`) only changes what the healthcheck *reports* during batch mode;
it cannot and does not touch the restart decision, which Compose makes from the exit event alone.
Run `docker compose up worker` with `METALDOCS_WORKER_RUN_ONCE=true` today and the batch reruns in
a loop — the "one-shot" contract main.go documents and Surface A's neighbor comments assume is not
actually held by this compose file.

**Kept correct by** discipline plus, for Surface A, a coincidence (no override currently exists to
expose the drift). Surface B is not even currently correct — it is a live behavior gap, not a
latent one.

**Firing mechanism** — both level 1 candidates, not yet built:
- **A:** interpolate one compose variable (e.g. `${WORKER_METRICS_ADDR:-:9091}`) into both the
  `environment:` entry and the `healthcheck.test:` command, so a compose-level override reaches
  both readings by construction instead of by remembering to edit two lines.
- **B:** a batch-specific compose service (or a `docker-compose.override.yml` invoked instead of
  the base `worker` service for one-shot runs) carrying `restart: "no"`, so "is this a batch run"
  is answered once, by which service definition is invoked, not by a shell guard reinterpreting an
  env var the restart policy never reads.

**Why this round didn't fix it:** both surfaces require editing `environment:` or `restart:` /
service topology in `deploy/compose/docker-compose.yml`. PR #109's own diff on that file is fenced
to `#` comments and `healthcheck:`/`test:`/`interval:`/`timeout:`/`retries:`/`start_period:` only,
specifically so PR #110 (Lane C, same file, rebasing on top of PR #109) keeps a clean rebase.
Widening the fence to fix this now would break that property for a concurrent, already-in-flight PR.

**Subsumes a third instance found in round 4** (`docker-compose.yml:309-343`, F9): the healthcheck's
shell `case` guard and `config.LoadWorkerConfig`'s Go truthy check are two independently-maintained
readings of "is `METALDOCS_WORKER_RUN_ONCE` truthy," proven to agree value-by-value but not
structurally identical. Landing Surface B's firing mechanism (a batch-specific service/override
carrying `restart: "no"`) removes the shell guard's reason to exist at all — a batch-mode compose
service invoked instead of the continuous one needs no `/live`-vs-RunOnce case split, so the second
truthy reading is deleted as a side effect, not patched in place.

**Owner:** unrouted. **Closes when:** a follow-up slice lands either firing mechanism above and
deletes this entry; until then this is the named tracked item for the "TRANSITIONAL local maximum"
label on `docker-compose.yml:339-343` and in
`docs/runbooks/worker-jobs-liveness-healthchecks.md`'s "Deliberately deferred" section.

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

---

## ME-19 — the negative fixture is the guard's guard, and nothing guards it

**Found** 2026-08-11 across four PRs of the engineering-remediation wave (#114, #120, #122, #123),
recorded because the same shape was hit by four independent authors in one day.

Level-3 enforcement is a two-layer stack: the guard asserts the rule, and a **negative fixture**
asserts the guard. The register's preference order treats "red build" as one rung, but a red build
is only as trustworthy as the fixture that proves the build can go red. That second layer is where
this wave's failures actually concentrated.

**Surfaces — four distinct ways a fixture asserts nothing, all observed live:**

| Shape | Instance |
|---|---|
| A fixture added to an already-failing tree | PR #114: three new negative fixtures (lowercase `from`, second builder stage, `FROM --platform=`) whose `Want` still listed only the original `worker.Dockerfile` line. `exit != 0` was carried by the first bad file; **measured** — pre-fix parser + all four fixtures => harness reports `ok`. |
| An exclusion with no `NotWant` | PR #114 again, found by cold review: discovery legitimately skips `vendor/`, and nothing proved it kept skipping. `Want` cannot express silence. |
| A `Want` loose enough to match an unrelated failure | PR #119: `Want: ["EXPIRED"]` — a bare substring an unrelated failure path can also produce, pinning the fixture to the wrong cause. |
| A fixture never proven RED against pre-fix code | Wave-wide. A fixture authored alongside its fix has never been observed to fail; "the guard is blocking" and "the guard has been observed to fail" are separate claims (registry rule A7 names this for `pr`-profile checks). |

**Kept correct by** — today: `tools/verify` registry rule **A7** (a `pr`-profile check must declare
a fixture or a written waiver) plus reviewer attention. A7 forces a fixture to *exist*. Nothing
forces it to *assert*. Every shape in the table above satisfies A7.

**The convergent-invention evidence.** `Fixture.NotWant` was invented **twice, independently, on the
same day**: PR #114 added it (`tools/verify/fixtures.go`), and PR #123 added a byte-identical
implementation on a branch forked from the same commit, because both lanes were briefed from the
same written rule. `git merge-tree` reported it as a content conflict; it was a duplicate. Two
authors reaching for the same missing mechanism within hours is the strongest available evidence
that the mechanism was missing and that the rule alone was not carrying it.

**Firing mechanism** — level 3, extending the existing A-rules, in ascending order of value:

- **A-fixture-RED:** a fixture's assertions must be proven against the *pre-fix* guard, not only
  the fixed one. Mechanically: the check declares the commit or patch that restores the defect, and
  the harness runs the fixture against it and requires red. This is the one that would have caught
  all four shapes above.
- **A-fixture-specific:** reject a `Want` string that also appears in the guard's generic failure
  output — the "wrong cause" shape.
- **A-exclusion-NotWant:** a guard declaring a scope exclusion (a skip list, an allowlist, a
  `_test.go` carve-out) must declare a corresponding `NotWant`. Derivable: the exclusion is already
  data in the guard.

**Verified** 2026-08-11, first-hand: the #114 measurement is recorded in that PR's Control Tower
takeover note; the #114/#123 `NotWant` implementations were diffed directly and are byte-identical.

**Owner:** unassigned; belongs to whoever next touches `tools/verify`'s registry rules.

---

## ME-20 — a gate signal written by a later phase than the destructive one

**Found** 2026-08-11 on PR #121 (erased-tenant audit read-path gate), disposing of a Codex TOCTOU
finding. Recorded as its own entry because the fix that was initially accepted is a level-4 patch of
a level-1 defect, and the difference is the whole point of this register.

**Surface.** `internal/modules/iam/application/tenant_lifecycle_service.go:540-563` — tenant erasure
runs in phases. Phase 1 (`eraseTenantRowsTx`) deletes `audit_export_jobs` **and commits**. Phase 3
(`cryptoShredAndTombstoneTx`) sets `tenants.erased_at`, which is the **only** signal
`refuseIfTenantErased` reads. Between them sits Phase 2's blob-deletion network I/O, so the window
is not a few milliseconds — it is as wide as an object-store sweep.

An export landing in that window writes a new plaintext `audit_export_jobs` row *after* the cleanup
that was supposed to remove them, and nothing ever sweeps it again. No DB trigger backstops it.
Durable, compliance-relevant, and invisible: every phase reports success.

**Kept correct by** — nothing. The ordering is incidental.

**The generalization, which is why this is an entry and not a bug report:** the destructive phase
runs *before* the phase that publishes "this thing is being destroyed". Any multi-phase lifecycle
that raises its gate flag at the end has a window whose width is whatever work sits in between, and
that window is invisible to every phase's own success check.

**Firing mechanism** — the accepted disposition on #121 was *re-check erasure immediately before
persisting, fail closed*. That is level 4: it narrows the TOCTOU window, it does not close it, and it
adds a check that must be remembered at every future write site.

**Level 1 is available and is a smaller change: raise the gate flag first.** Mark the tenant
`erasing_at` (or set `erased_at`) in the transaction that *begins* erasure, before any destructive
phase runs, and let `refuseIfTenantErased` read that. The window then cannot exist — not because
something watches for it, but because there is no interval during which the tenant is being erased
and does not say so. The gate already reads that signal; only the write order changes.

Stated as doctrine: **the flag that gates access to a resource must be set by the first phase that
makes the resource unsafe, never by the last phase that finishes making it unavailable.**

**Boundary.** Out of PR #121's scope — that PR is `audit` module read-path enforcement; this is
`iam` module lifecycle phase ordering. #121's Codex thread at
`internal/modules/audit/infrastructure/postgres/writer.go:364` is **relocated, not undisposed**:
the finding is real and the relocation is the disposition.

**Verified** 2026-08-11, first-hand: phase ordering read directly in
`tenant_lifecycle_service.go`; `refuseIfTenantErased`'s single-signal dependency confirmed.

**Owner:** own unit in the `iam` module, not yet scheduled.
