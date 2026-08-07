# AuthZ grant unification — design spec

**Date:** 2026-08-07 · **Revision:** v2 (post-audit) · **Status:** design, pending operator review
**Gate:** `developing-new-work` passed 🟡 Yellow —
`docs/superpowers/analysis/2026-08-06-authz-grant-unification-system-impact.md`
**Inputs:** `docs/superpowers/analysis/2026-08-06-authz-grant-unification-decisions.md` (D1–D4),
`docs/superpowers/analysis/2026-08-07-authz-advisory-opinion.md` (advisory rounds 1–2, DD-5, DD-6)
**Next ADR number:** 0092 · **Next migration:** 0318

> **v2 changelog.** v1 was audited against the standard "does this actually delete the manual and the
> legacy, or does it read well?" and came back **(b): a large real improvement, but N things still
> rely on discipline.** v2 closes every finding. Material changes: F3 and F5 had **overstated levels**
> and are corrected (§3.4, §8) — F5's "unrepresentable" was outright false because areas are
> soft-archived, so the FK it relied on never fires; §5 **violated the module boundary** and is
> rewritten (§5); `recordBypass` deletion was **overbroad** and would have unaudited every background
> bypass (§6); the **tripwire re-point** (§6.2), the **HTTP contract surface** (§9), **`iam_group_members`**
> (§3.5) and the **transaction seam** (§4.2) were absent and are now specified; `role_capabilities`
> was dispositioned "unchanged" and is not (§3.3); `subject_id` chose polymorphism over referential
> integrity and no longer does (§3.2). Accepted residue is enumerated in §14 rather than left
> implicit.

---

## 1. The defect, in one paragraph

Tier-1 (`CapabilityService.CanDo`, `internal/modules/iam/application/capability_service.go:48`) reads
`iam_user_roles` ∪ `iam_group_members`⋈`iam_group_roles`. Tier-2 (`authz.Require`,
`internal/modules/iam/authz/authz.go:145`) reads `role_capabilities`⋈`user_process_areas`. **Disjoint
tables.** Nothing forces tier-1 to be a relaxation of tier-2, so a principal holding only an area
membership is refused at the HTTP edge before tier-2 ever runs — the area model is unreachable except
to further narrow someone tier-1 already admitted. `TestNoDeclaredOperationIsUnreachable` reports the
consequence and is red by design.

ADR 0007's own Context records this as an unfinished 2026-05-02 migration; its Decision then ratifies
the split as architecture. That is an unlabelled local maximum under CLAUDE.md's Global Maximum rule.

## 2. Ratified decisions this design implements

| ID | Decision | Source |
|---|---|---|
| D1 | One `capability_bindings` relation; scope lives on the binding | 2026-08-06 |
| D2 | No bypass; `system_admin` holds every capability as ordinary grants | 2026-08-06 |
| D3 | No direct per-user capability grants | 2026-08-06 |
| D4 | Groups (who) × areas (where) are orthogonal; both survive | 2026-08-06 |
| DD-1 | All `system_admin` bypasses die; visibility becomes `document.read` at tenant scope | 2026-08-06 |
| DD-2 | Hard break: one transaction, migration 0318, legacy tables dropped | 2026-08-06 |
| DD-4 | One Go evaluator, one query builder; `tier1 := ∃scope. tier2` | 2026-08-06 |
| DD-5 | Authorship floor is **revision-scoped**, superseding DD-3 | 2026-08-07 |
| DD-6 | Role catalog is **product reference data**; enums stay generated | 2026-08-07 |

Not open for re-debate. Everything below is *how*.

## 3. Target model

### 3.1 `role_catalog` — the single role upstream

```
role_catalog(
  code            text primary key,
  label_pt_br     text not null,
  area_assignable boolean not null,   -- false only for system_admin
  display_order   int  not null,
  description     text
)
```

Product reference data, not tenant data (DD-6). Seeded, not tenant-writable. This kills ME-01: the
seven role declaration surfaces collapse to one upstream, with `iamtypes` sets and the OpenAPI
`UserRole` / `AreaRole` enums **generated** from it. `area_assignable` replaces the hand-maintained
`areaRoles` subset and makes the `AreaRole ⊂ UserRole` relationship data rather than a coincidence
two enums happen to preserve — `role-vocabulary.ts:34-36` currently *proves* that subset relation at
compile time and composes with a generated upstream unchanged.

**The physical upstream is a checked-in artifact, not the table.** A DB table cannot be a codegen
input at build time. The canonical artifact is `db/reference-data/role_catalog.yaml`; the generator
emits (a) the `role_catalog` seed SQL, (b) `iamtypes`' Go sets, (c) the OpenAPI `UserRole` /
`AreaRole` enums. Stating this matters: left unnamed, F1 degrades to "a generator with a
hand-maintained input nobody declared," which is the ME-01 defect wearing a generator's clothes.

### 3.2 `capability_bindings` — the single assignment relation

```
capability_bindings(
  id             uuid primary key,
  tenant_id      uuid not null,
  subject_kind   text not null check (subject_kind in ('user','group')),
  subject_user_id  text references iam_users(user_id),
  subject_group_id uuid references iam_groups(id),
  role_code      text not null references role_catalog(code),
  scope_kind     text not null check (scope_kind in ('tenant','area')),
  scope_ref      text references process_areas(code),
  effective_from timestamptz not null default now(),
  effective_to   timestamptz,          -- soft revocation; never DELETE
  granted_by     text not null,
  revoked_by     text,
  reason         text,

  check ((subject_kind = 'user')  = (subject_user_id  is not null)),
  check ((subject_kind = 'group') = (subject_group_id is not null)),
  check ((scope_kind   = 'tenant') = (scope_ref is null))
)
```

Kubernetes-shaped: `role_catalog` is the bundle, `capability_bindings` is the binding, scope rides on
the binding. `subject_kind` is what makes D4 expressible — **binding a group to an area** is a
capability the current schema cannot state at all.

**Two subject columns, not one polymorphic column.** A single `subject_id text` forfeits referential
integrity: a binding to a deleted or never-existing user or group would be representable, insertable,
and invisible to every guard in §8. The discriminated pair keeps real FKs and costs nothing, since
`subject_kind` already discriminates. This is the register's own doctrine — prefer unrepresentable
over guarded — applied to the *new* schema rather than only to the legacy it replaces.

`subject_user_id` is `text`, not `uuid`: `iam_users.user_id` is the TEXT PK, and typing actor columns
as uuid is a known repeat defect (memory: tokens actor-id text contract).

**Invariants re-homed from `user_process_areas`, not silently dropped.** The legacy table carries
three DB-level protections that must survive onto `capability_bindings` or the migration is a net
loss of enforcement:

- `enforce_user_process_areas_update_contract` (`0001_current_schema.sql:877`) — identity columns are
  immutable on UPDATE. Re-home: `(tenant_id, subject_*, role_code, scope_*, effective_from,
  granted_by)` immutable.
- `reject_user_process_areas_delete` (`:952`) — DELETE is refused; revocation is an UPDATE. Re-home
  verbatim.
- The soft-delete revocation function (`:141`) — sets `effective_to` + `revoked_by`. Re-home.

Revocation therefore sets `effective_to` and rows are never deleted (ADR 0037 soft-delete model).
Grant history is Part 11 evidence: "who could do X on date D" must stay answerable retroactively.

### 3.3 `role_capabilities` — kept, but not untouched

The three-part shape (`role_catalog` = roles, `role_capabilities` = role→capability edges, the Go
registry = capabilities) is the NIST-RBAC shape and matches Kubernetes' `ClusterRole` rules. It does
**not** collapse into `capability_bindings`. But "unchanged" was the wrong disposition: the table is
hundreds of hand-typed `INSERT` rows
(`db/reference-data/0001_product_reference_data.sql:156+`) and is where the next drift lives.

Three changes, all mechanical, none touching bundle *content* (which is a human judgment and stays
one):

1. **`role_capabilities.role` gets the FK to `role_catalog(code)`.** F1 otherwise covers only
   `capability_bindings.role_code`; a typo'd role in the seed still inserts fine today.
2. **`role_capabilities.capability` gets a referential parity guard against the Go registry.** A
   typo'd capability is a silently dead grant row that *reads* as a grant — the ME-04 shape. Validity
   is mechanical even though correctness is not.
3. **`system_admin`'s bundle is generator output, never a hand-seeded list of N rows** (advisory
   round-1 P2). Otherwise capability N+1 is added to the Go registry and `system_admin` silently
   lacks it — the disjoint-tables defect reborn as a stale bundle. Generator + blocking parity lint,
   or a boot check asserting `bundle(system_admin) == registry`.

#### 3.3.1 D2 silently disarms an existing guard — this must be fixed in the same change

`TestEveryCapSeededOrDeferred` (`apps/api/cmd/metaldocs-api/permissions_test.go:597`) is a real,
landed level-3 mechanism: every registry capability must be seeded to ≥1 role or sit in a documented
`deferred` allow-list. That allow-list currently holds `distribution.read` and `notification.read` —
**the exact finding that originated this program.**

Under D2, `system_admin` holds every capability, so `seededCaps` contains everything by construction.
The test goes green forever and the allow-list becomes dead code. A program whose purpose is
mechanical enforcement would have deleted one of the few guards already catching this class.

**Required:** `seededCaps` excludes `system_admin` rows, so the guard keeps asking its real question —
"is this capability reachable by a *tenant* role?" Without this, §11 acceptance is satisfied by a
test that can no longer fail.

Related, and filed rather than fixed here: that same allow-list carries the comment *"Mirrors
deferredCaps in `scripts/api-lint/registry_rules.go`"* — a hand-synced enumeration living inside the
anti-drift mechanism. Prose is not a sync mechanism. Register entry, not this program's scope.

### 3.4 Dangling scope — the F5 correction

`tier1 := ∃scope. tier2` reintroduces unreachability through one door: a binding whose `scope_ref`
names a dead area. Tier-1 admits, tier-2 denies for every live area — the program's own defect,
per-user.

**v1 claimed an FK plus a revoke-on-retire rule made this unrepresentable. That was false.** Areas are
**soft-archived**, not deleted (ADR 0010; `internal/modules/taxonomy/domain/area.go:120-128`
stamps `ArchivedAt`). An FK constrains DELETE and therefore fires *never* on the path that actually
occurs. The whole claim rested on "retiring an area revokes its bindings in the same transaction" —
a rule a human must remember to implement in the taxonomy archive service. Level 5 wearing level 1's
clothes, which is precisely the ME-02 failure this register exists to catch.

**Correction — the invariant moves inside the evaluator.** Tier-1's existential joins live areas:

```
tier1(cap) := ∃b ∈ bindings(actor, cap).
                b.scope_kind = 'tenant'
             OR b.scope_ref ∈ (areas where archived_at is null)
```

A binding scoped to an archived area now grants nothing at **either** tier, by construction, in the
one query builder the whole design already trusts. It degrades safely under any future archive path,
including ones nobody has written yet.

Revoke-on-archive stays as hygiene — keeping `effective_to` honest for audit — but correctness no
longer depends on anyone remembering it.

### 3.5 `iam_group_members` — survives, and is now load-bearing

v1's backfill read this table and never stated its fate. It **must survive**: `subject_kind='group'`
bindings resolve to users only through membership, which makes membership **half of every
group-derived authorization decision**. That is strictly more load than it carries today.

Consequences, all of which v1 omitted:

- The evaluator's subject predicate is `subject_user_id = :actor OR subject_group_id IN (:actor's
  groups)`. The membership join belongs **inside the builder** and is covered by the sole-reader rule
  (§8 F2), exactly like `capability_bindings`.
- Editing membership is editing effective grants. It is already tripwire-armed
  (`0001_current_schema.sql:414`, `user.manage`) — that arm stays, and §6.2 confirms rather than
  assumes it.
- Cache invalidation (§10) keys on **two** tables, not one.

## 4. The evaluator

### 4.1 The two predicates

One package owns the grant query:

```go
// tier-2, the binding decision
Granted(ctx, tx, tenantID, subjectID, capability, scope) (bool, error)

// tier-1, the cheap edge reject
GrantedAnyScope(ctx, db, tenantID, subjectID, capability) (bool, error)  // ≡ ∃scope. Granted(...)
```

`GrantedAnyScope` is `Granted` with the scope conjunct dropped, built by the same builder. That makes
**tier-1 ⊇ tier-2 a theorem of the code shape**, not a property a test hopes holds. The permanent
property test stays anyway (gate constraint 4) — the theorem argues the test is cheap, not absent.

### 4.2 The transaction seam — stated, not improvised

v1 gave signatures without a `tx` and left the seam to implementation time. Explicitly:

- **Tier-2 is in-transaction, always.** `authz.Require(ctx, tx, cap, area)` keeps its exported
  signature and composes the builder internally. It must run on the caller's tx because
  `appendAssertedCap` writes the tx-local `metaldocs.asserted_caps` GUC (`authz.go:256-294`) that the
  DB tripwire reads. Nothing about that changes.
- **Tier-1 is a pooled read.** There is no transaction at middleware time, and there must not be one.
  `GrantedAnyScope` takes a `*sql.DB`.
- Read-skew between the two within one request is inherent and harmless: tier-2 is the binding
  decision.

### 4.3 Sole-reader rule

The identifiers `capability_bindings` and `iam_group_members` may appear in production Go and SQL
only inside the evaluator package. Every read model composes the builder; none re-implements the
join. Without this, `AreaCapabilityReader`, `MembershipDirectoryScope` and the document-list query
drift from the evaluator and ME-03 recurs as "the UI offers areas the evaluator denies."

The lint also scans `db/migrations`, because a migration creating a **view** over
`capability_bindings` would let readers hit the view with the identifier lint still green.

**Read-model re-points** (all currently embed the grant join or a bypass):
- `internal/modules/iam/infrastructure/postgres/area_capability_reader.go:49-64`
- `internal/modules/iam/infrastructure/postgres/user_area_repository.go:159` (`MembershipDirectoryScope`)
- `internal/modules/iam/application/capability_service.go:132-139` (`CapsByUserID`)
- the document-list visibility query (§5)

## 5. Document-list visibility (ME-03 / issue #76) — corrected for module boundary

Today `internal/modules/documents/delivery/http/handler.go:435-439` is the whole model: admin sees
everything, everyone else sees only what they created. An area `viewer` holding a real grant sees
nothing.

The target relation is:

```
visible(actor, D) :=
      D.area ∈ areas where actor holds document.read     -- tenant scope ⇒ all areas
   ∪  D has a revision authored or signed by actor        -- DD-5 floor
```

**v1 ruled this lives "inside the evaluator, never in the handler" — half right, and the wrong half
broke a module boundary.** Revision provenance lives in `documents`; signoffs live in `approval`
(`approval_signoffs`, the module promoted by ADR 0082). Putting the union in `iam` means iam's SQL
joining two other modules' tables — the invariant CLAUDE.md states in bold. v1 conflated *not in the
handler* (correct — that is how ME-03 was born) with *in iam* (wrong module).

**Corrected structure. Visibility is the `documents` module's question:**

- `iam` publishes **scope resolution only** — `AreaCapabilityReader`'s existing contract,
  `(tenant_wide bool, areas []string)` for a capability, re-pointed at the builder. Nothing about
  documents enters iam.
- `documents`' repository composes that with its own revision provenance.
- The signed-revision leg needs a **published port from `approval`** — unless signoff is already
  denormalized into revision provenance at signing time, in which case no new port. **Implementation
  step 1 is to determine which**, and the spec is not finishable by guessing.

"One query answers who can read D" survives — one query *in the owning module*, built from published
ports.

The `WHERE created_by` shortcut stays forbidden.

## 6. Bypass extinction (DD-1, ME-09 / issue #82)

Five sites, not the three originally listed:

| | Site | Disposition |
|---|---|---|
| a | `capability_service.go:56-68` — tier-1 UNION admin arms | delete; ordinary bindings |
| b | `authz.go:113-136` — tier-2 short-circuit | delete the **arm**; see §6.1 |
| c | `handler.go:436` + `area_capability_reader.go:51` | delete; §5 |
| d | `user_area_repository.go:159` — `SystemAdminExistsSQL` branch | delete; compose the builder |
| e | `capability_service.go:132-139` — `CapsByUserID` → `AllCapabilities()` | delete; else `/auth/me` renders capabilities by a different rule than enforcement uses |

**Definition of done is mechanical, not an inventory** (ME-09): the extinction lint of §8 F3. The list
above is a hand-synced enumeration and was already wrong once, by 40%, while reading as complete.

### 6.1 `recordBypass` — the system_admin arm dies, the function does not

v1 tabulated "delete short-circuit + `recordBypass`". Overbroad. `recordBypass` serves **two** kinds,
and only one is in scope:

- `BypassKindSystemAdmin` (`authz.go:126`) — dies with D2.
- `BypassKindBackground` (`authz.go:213`) — **live and out of scope.** `BypassSystem`
  (`authz.go:197`) is the scheduler-only bridge, fail-closed outside a background context, and it
  records every background bypass through this same function. Deleting it wholesale unaudits every
  scheduler privilege use — an F8-class regression on the path nobody watches.

So: delete the system_admin arm, keep the function and the background kind.

This also **refines** the read-only-tx reasoning: `Require` becomes RO-legal, but `BypassSystem` still
INSERTs in-tx. The `authz-require-rw-tx` lint relaxation applies to `Require` call sites only.

### 6.2 The audit disposition — mandatory ADR content

D2's claim that "the ordinary path already records it" is **false today**. `authz.Require:144-166`
records nothing; only the bypass arm INSERTs (`authz.go:125-133`, fail-closed in-tx, ADR 0022
Phase 11 F8, whose stated rationale is that a tenant-wide bypass must appear in `audit.read` at the
same fidelity as a normal grant). Deleting that arm moves `system_admin` privilege use from per-use
audited to unaudited.

ADR 0092 must dispose of F8 explicitly, with the reconstruction query shown: after unification the
grant is durable provenance-carrying data, so "why was this actor authorized" is
(operation audit event) ⋈ (binding history) — which the old model could not answer at all. Silently
dropping the INSERT is how a Part 11 finding is born.

### 6.3 Tripwire re-point — required by the gate, absent from v1

Invariant 6 is that the DB enforces invariants. The baseline arms
(`db/baseline/0001_current_schema.sql`) currently guard the tables being dropped:

| Arm | Table | Required capability |
|---|---|---|
| `:355` | `iam_user_roles` | `user.manage` |
| `:358` | `user_process_areas` | `membership.manage` |
| `:414` | `iam_group_members` | `user.manage` |

Migration 0318 drops the first two, so their arms must be removed **and `capability_bindings` needs a
new arm** — otherwise grant mutations end up *less* DB-guarded after a program about authorization
hardening.

**Design decision, made here rather than left to the implementer:** the required capability is
scope-discriminated, mirroring today's split — `scope_kind='tenant'` requires `user.manage`,
`scope_kind='area'` requires `membership.manage`. This preserves the existing authority boundary
(an area admin manages memberships in their area; granting tenant-wide authority is a different
power) instead of flattening two capabilities into one by accident of refactoring.

The `iam_group_members` arm is unchanged and stays. Arms are generator output (GMR M2) — the
generator is re-run, not hand-edited.

## 7. Migration 0318 — one transaction

Order: create `role_catalog` + `capability_bindings` → seed the catalog → backfill → **assert** →
re-point tripwire arms → drop `iam_user_roles`, `iam_group_roles`, `user_process_areas`.

Postgres DDL is transactional; `now()` is tx-stable so effective-interval evaluation cannot skew
mid-migration; with no production tenant, lock contention is irrelevant. Expand/contract was
considered and rejected: a dual-write window is two sources of truth — this program's defect, as a
rollout procedure.

**Backfill mapping**

| From | To |
|---|---|
| `iam_user_roles(user_id, role)` | `subject_kind='user'`, `scope_kind='tenant'` |
| `iam_group_roles` ⋈ `iam_group_members` | `subject_kind='group'`, `scope_kind='tenant'` |
| `user_process_areas(user_id, area_code, role)` | `subject_kind='user'`, `scope_kind='area'`, `scope_ref=area_code`, intervals and provenance carried across |

`iam_group_members` itself is **not** dropped (§3.5).

`iam_group_roles` has no CHECK, so it may hold role strings that exist nowhere and will violate the
new FK. They grant nothing today (they join to nothing in `role_capabilities`), so skipping them is
semantically safe — but it must be an explicit `WHERE role IN (SELECT code FROM role_catalog)` with a
**raised NOTICE and count**, never an accidental FK failure or a silent drop.

**Evidence before dropping.** The migration copies the three legacy tables verbatim into a
`_migration_0318_archive` schema before the DROPs. This is **not** a rollback mechanism — it is
evidence, at near-zero cost, and it is what makes §12's no-rollback position defensible in writing.

**Atomicity:** the `SeedWithCaps` re-point (gate constraint 6) lands in the same commit as 0318, or
the integration estate breaks against the template databases.

### 7.1 The equivalence assertion — three-part and directional

The in-migration assert is a **belt**. It can pass while the model is wrong, three ways:

1. **Vacuously** — the dev DB was wiped 2026-07-29, and old-empty equals new-empty for *any*
   backfill, including a broken one.
2. **By transcription** — the assert must re-express Go-embedded SQL in migration SQL. Drop one
   predicate (`g.tenant_id = $2`, `effective_to IS NULL`, the `IsValidCapability` pre-filter) and it
   compares the new model against a wrong rendering of the old one, and passes.
3. **By direction** — "the new source reproduces every old grant" is old ⊆ new and silent about
   new ∖ old, while the design *deliberately* creates new grants (`iam_user_roles` rows become
   tenant-scoped and are newly honored by tier-2). One-directional lets unintended escalation ride in
   beside the intended kind.

So the assert is three-part, all parts materialized and compared:

- tier-1: **exact equality**, old = new
- tier-2: old ⊆ new
- tier-2: **new ∖ old = the expected-escalation set**, which is **computed in the migration**
  (`SELECT … FROM iam_user_roles` at assert time), never hand-listed. A hand-listed expected set is a
  hand-sync against the backfill and would defeat the assert's purpose.

**The real acceptance evidence is a seeded Go parity test**, which calls the actual old evaluator
functions and so dodges (1) and (2) entirely. Corpus must cover: direct role; group-derived role;
`iam_group_roles` junk strings; expired `effective_to` rows; area grants for every area role;
the roles the `iam_user_roles` CHECK omits (`area_admin`, `qms_admin`, `signer`); `system_admin` via
both direct and group; **and a binding scoped to an archived area** (§3.4). The corpus's role list is
**derived from `role_catalog` in the test**, not typed out, so a future role is covered automatically.
Diff old-evaluator against new-evaluator for every (actor, capability, area).

### 7.2 The tenant-scope population needs a one-time human review

The backfill converts **every** legacy `iam_user_roles` row to `scope_kind='tenant'`. Day one, the
tenant-scope population is exactly the old sloppy population — now honored by tier-2, i.e. strictly
more power than it had. §7.1's escalation assert makes that *visible*; nothing makes it *reviewed*.

Acceptance item: every tenant-scope binding is deliberately re-affirmed or narrowed to an area, once,
after the migration. In a Part 11 shop that review is itself evidence. This is the one place the
program legitimately requires human judgment — and it is bounded, one-time, and recorded, rather
than standing discipline.

## 8. Firing mechanisms this program must land

Ranked per register doctrine, with levels stated honestly and **non-claims written into the guard**
(ME-02's standing instruction).

| # | Mechanism | Level | Kills | Does **not** cover |
|---|---|---|---|---|
| F1 | `role_code` FK to `role_catalog`; Go sets + OpenAPI enums generated from `role_catalog.yaml` | 1 (FK) / 3 (generation) | ME-01 / #74 | whether a role *should* exist |
| F2 | Sole-reader lint: `capability_bindings` + `iam_group_members` only inside the evaluator package, migrations scanned for views | 3 | evaluator/read-model drift | logic errors inside the builder |
| F3 | Extinction lint: the literal `'system_admin'` absent outside the catalog seed and generated files | 3 | ME-09 / #82 | **role-identity special-casing built from the generated constant** — see below |
| F4 | `system_admin` bundle is generator output with a parity lint, or a boot assert vs the registry | 2–3 | stale bundle | correctness of any *other* role's bundle |
| F5 | Tier-1's existential joins live areas (§3.4) | 1 | dangling scope | nothing — the state is unreachable by construction |
| F6 | Permanent property test `∀cap. tier1(cap) ⊇ tier2(cap, ·)` | 3 | tier divergence | — |
| F7 | `TestNoDeclaredOperationIsUnreachable` green | 3 | **the definition of done** | whether a reachable operation is *correctly* guarded |
| F8 | `role_capabilities.role` FK + capability referential parity vs the Go registry (§3.3) | 1 / 3 | typo'd seed rows that read as grants | bundle *correctness*, a human judgment |
| F9 | `seededCaps` excludes `system_admin` (§3.3.1) | 3 | D2 disarming an existing guard | — |

**F3's honest scope, stated because §9.6 of v1 got this wrong.** F1 generates Go role constants, so
generated code legitimately contains `system_admin` and F3 must allowlist generated files. From that
moment any hand-written `if role == iamtypes.RoleSystemAdmin { … }` is a brand-new bypass with zero
literals and a green F3. The lint's noun is *the string*; the defect's noun is *role-identity
special-casing in enforcement paths*. F3 claims **literal extinction only**, and constant-based
special-casing is named uncovered residue (§14). The stronger mechanism — enforcement packages may
consume roles only through the evaluator — is the right global maximum and is scoped in §14 rather
than half-built here.

F7 is gate constraint 3 and load-bearing: the red CI lane goes green because the unreachable set is
empty, never because the test was skipped, excluded, or baselined.

## 9. HTTP contract surface

Contract-first: routes change only by editing `api/openapi` + `oapi-codegen`. A design spec with no
contract section is unfinishable without improvisation, so:

**New IAM binding surface** (replacing the split role/membership endpoints):
- `POST /api/v1/iam/bindings` — create a binding `(subject, role, scope)`
- `DELETE /api/v1/iam/bindings/{id}` — revoke (soft; sets `effective_to`)
- `GET /api/v1/iam/subjects/{id}/bindings` — bindings held by a subject
- `GET /api/v1/iam/roles/{code}/subjects` — subjects holding a role (the "who can do X" query D3's
  auditability argument depends on)

Capabilities per §6.3: tenant-scope writes require `user.manage`, area-scope writes require
`membership.manage`. Existing role-assignment and area-membership endpoints are **deleted**, not
deprecated (legacy-fallback extermination doctrine).

**Frontend re-point** — in scope per §11, and it is real work: 18 files consume the role enums, and
the two membership screens collapse into one binding surface. The generated types carry the enum
change for free; the screens do not.

## 10. Cache invalidation (gate RF-3)

Closed by this program, per the gate's instruction. After unification exactly two tables invalidate
capability caches: `capability_bindings` and `iam_group_members` (§3.5). `WithCapCache` is
per-request, so live risk is small — which is precisely why the contract is worth writing now, while
there are two tables to watch instead of three plus a bypass.

## 11. ADR 0092 — required content

1. **Supersedes ADR 0007's Decision.** The two-tier split stays as *enforcement points*; the
   disjoint-grant-tables ruling is withdrawn.
2. **MUST-deviation from REQ-AUTHZ-4** — "`system_admin` tenant-wide bypass is the only inheritance
   short-circuit" becomes false when there is no bypass at all. Target spec amended in the same
   change.
3. **F8 audit disposition** with the reconstruction query (§6.2), and the `recordBypass` scoping
   (§6.1).
4. **`authz-require-rw-tx` relaxation**, scoped to `Require` call sites only (§6.1).
5. **DD-6's role-catalog ruling** and its inversion trigger (ME-10 / #83).
6. **DD-5's audit-surface consequence.** The floor holds — an authored revision is an immutable
   historical *fact*, not an administrable assignment, so no second grant regime and no D3 exception.
   But "who can read D?" is now one query over **bindings ∪ revision provenance**: two legs, one of
   which is not in `capability_bindings`. Any future "who can access X" audit tooling must include the
   provenance leg or it silently under-reports.
7. **DD-5 residual, accepted in writing.** An actor under a for-cause investigation retains retrieval
   of the revisions at issue — their own frozen acts, unseverable. Acceptable (the audit trail holds
   them regardless), but stated rather than silent.
8. **No rollback, and why that is acceptable.** Pre-commit the single tx aborts clean. Post-commit
   there is no path back — the legacy tables are gone. With no production tenant that is sound, and
   `_migration_0318_archive` (§7) preserves evidence. A default is not a decision; this must be a
   sentence.
9. **What these guards do not claim** (ME-02): the §8 non-claims column, especially F3's.

## 12. Out of scope, named so it is not smuggled in

- **ME-07 / #80** authentication build-vs-adopt — separate study, post-v1.
- **ME-08 / #81** MFA — routed; deleting the coverage endpoint is a standalone fix.
- **ME-06 / #79** relation-tuple engine — triggered only by per-document sharing.
- **ME-10 / #83** tenant-defined roles — triggered by the first customer requiring them.
- The `deferredCaps` hand-mirror between `permissions_test.go` and `scripts/api-lint/registry_rules.go`
  (§3.3.1) — a register entry, not this program.

## 13. Acceptance

1. `TestNoDeclaredOperationIsUnreachable` green; `./apps/...` restored to green in `test-full.yml`
   and `test-nightly.yml` **by the unreachable set being empty**.
2. Seeded Go parity test passes over the full corpus (§7.1), including the archived-area case.
3. F1–F9 landed and firing.
4. Full integration ladder per the selective policy (touched packages + guard suites; `./...` because
   this touches db/platform).
5. ADR 0092 written with §11's nine items; target spec amended; `wiki/modules/iam.md` and
   `wiki/architecture/backend-target-architecture.md` updated.
6. OpenAPI regenerated (full regen — partial is forbidden drift); frontend re-pointed; the two
   membership screens replaced by the binding surface.
7. The one-time tenant-scope review (§7.2) performed and recorded.
8. Issues #74, #76, #82 closed by the landed mechanisms; #75 closed if F1 subsumes it.

## 14. Accepted residue — what still relies on discipline after this lands

Written down because unnamed residue is indistinguishable from no residue.

1. **`role_capabilities` bundle *content*** — which capabilities a role should hold is a human
   judgment. F8 makes every row referentially valid; nothing makes it *right*. Accepted: this is
   genuinely a review object, which is D3's whole argument.
2. **Role-identity special-casing via generated constants** — uncovered by F3 (§8). The fix is a
   semantic lint (enforcement packages consume roles only through the evaluator). **Route to the
   register as a follow-up entry**; it is the one residue that could silently reconstruct a bypass.
3. **`subject_kind` / `scope_kind` vocabularies stated twice** — SQL CHECKs and Go builder constants.
   Two surfaces, hand-agreed, small and closed. Accepted as two-surface residue.
4. **F3's allowlist** — exactly two path patterns (the catalog seed, generated files), asserted
   inside the lint itself so the carve-out cannot quietly grow.
5. **The one-time tenant-scope review** (§7.2) — bounded, recorded, not standing.

Items 1, 3, 4, 5 are accepted permanently. Item 2 is a follow-up with a named mechanism.
