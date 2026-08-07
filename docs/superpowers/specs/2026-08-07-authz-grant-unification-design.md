# AuthZ grant unification — design spec

**Date:** 2026-08-07 · **Status:** design, pending operator review
**Gate:** `developing-new-work` passed 🟡 Yellow —
`docs/superpowers/analysis/2026-08-06-authz-grant-unification-system-impact.md`
**Inputs:** `docs/superpowers/analysis/2026-08-06-authz-grant-unification-decisions.md` (D1–D4),
`docs/superpowers/analysis/2026-08-07-authz-advisory-opinion.md` (independent advisory + DD-5, DD-6)
**Next ADR number:** 0092 · **Next migration:** 0318

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

Not open for re-debate in this program. Everything below is *how*.

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
two enums happen to preserve.

### 3.2 `capability_bindings` — the single assignment relation

```
capability_bindings(
  id           uuid primary key,
  tenant_id    uuid not null,
  subject_kind text not null check (subject_kind in ('user','group')),
  subject_id   text not null,          -- iam_users.user_id (TEXT) or group id
  role_code    text not null references role_catalog(code),
  scope_kind   text not null check (scope_kind in ('tenant','area')),
  scope_ref    text,                   -- null iff scope_kind='tenant'
  effective_from timestamptz not null default now(),
  effective_to   timestamptz,          -- soft revocation; never DELETE
  granted_by   text not null,
  revoked_by   text,
  reason       text
)
```

Kubernetes-shaped: `role_catalog` is the bundle, `capability_bindings` is the binding, scope rides on
the binding. `subject_kind` is what makes D4 expressible — **binding a group to an area** is a
capability the current schema cannot state at all.

`subject_id` is `text`, not `uuid` — `iam_users.user_id` is the TEXT PK, and typing actor columns as
uuid is a known repeat defect (memory: tokens actor-id text contract).

Constraints:
- `check ((scope_kind = 'tenant') = (scope_ref is null))` — a tenant binding cannot name a scope, an
  area binding must.
- `scope_ref` FK to the areas table with **explicit retire semantics** (§3.4).
- Revocation sets `effective_to`; rows are never deleted (ADR 0037 soft-delete model). Grant history
  is Part 11 evidence: "who could do X on date D" must stay answerable retroactively.

### 3.3 `role_capabilities` — unchanged

Already correct. The defect was never on the bundle side.

`system_admin`'s bundle is the full capability set, and per advisory P2 it is **not** a hand-seeded
list of N rows: it is generator output from the Go capability registry with a blocking parity lint,
or a boot check asserting `bundle(system_admin) == registry`. Without one of those, D2 trades a
visible bypass for an invisible drift the day capability N+1 is added.

### 3.4 Dangling scope

`tier1 := ∃scope. tier2` reintroduces unreachability through one door: a binding whose `scope_ref`
names a retired or deleted area. Tier-1 admits, tier-2 denies for every live area — the program's own
defect, per-user. **Ruling:** retiring an area revokes (`effective_to = now()`) every binding scoped
to it, in the same transaction. The FK plus that rule makes a dangling scope unrepresentable rather
than guarded.

## 4. The evaluator

One package owns the grant query. Two predicates over one source:

```go
// tier-2, the binding decision
Granted(ctx, tenantID, subjectID, capability, scope) bool

// tier-1, the cheap edge reject
GrantedAnyScope(ctx, tenantID, subjectID, capability) bool  // ≡ ∃scope. Granted(...)
```

`GrantedAnyScope` is literally `Granted` with the scope conjunct dropped, built by the same query
builder. That makes **tier-1 ⊇ tier-2 a theorem of the code shape**, not a property a test hopes
holds. The permanent property test stays anyway (gate constraint 4) — the theorem argues the test is
cheap, not absent.

**Sole-reader rule (advisory gap 1, ME-09-shaped).** The identifier `capability_bindings` may appear
in production Go and SQL only inside the evaluator package. Every read model composes the builder;
none re-implements the join. Without this, `AreaCapabilityReader`, `MembershipDirectoryScope` and the
document-list query drift from the evaluator and ME-03 recurs as "the UI offers areas the evaluator
denies." Lint shape: SOLE-RLS.

**Read-model re-points** (all currently embed the grant join or a bypass):
- `internal/modules/iam/infrastructure/postgres/area_capability_reader.go:49-64`
- `internal/modules/iam/infrastructure/postgres/user_area_repository.go:159` (`MembershipDirectoryScope`)
- `internal/modules/iam/application/capability_service.go:132-139` (`CapsByUserID`)
- the new document-list visibility query (§5)

## 5. Document-list visibility (ME-03 / issue #76)

Today, `internal/modules/documents/delivery/http/handler.go:435-439` is the whole model: admin sees
everything, everyone else sees only what they created. An area `viewer` holding a real grant sees
nothing.

Replaced by, **inside the evaluator**, never in the handler:

```
visible(actor, D) :=
      D.area ∈ areas where actor holds document.read     -- incl. tenant scope ⇒ all areas
   ∪  D has a revision authored or signed by actor        -- DD-5 floor
```

The floor is revision-scoped (DD-5): it grants retrieval of *the actor's own act*, frozen, not the
living document at whatever classification it later reached. Because an authored revision is a fact
already in revision provenance rather than a per-user ACL row, "who can read D?" stays a single query
and D3 needs no named exception.

The `WHERE created_by` shortcut is forbidden. That shortcut is how ME-03 was born.

## 6. Bypass extinction (DD-1, ME-09 / issue #82)

Five sites, not the three originally listed:

| | Site | Disposition |
|---|---|---|
| a | `capability_service.go:56-68` — tier-1 UNION admin arms | delete; ordinary bindings |
| b | `authz.go:113-136` — tier-2 short-circuit + `recordBypass` | delete; see §6.1 |
| c | `handler.go:436` + `area_capability_reader.go:51` | delete; §5 |
| d | `user_area_repository.go:159` — `SystemAdminExistsSQL` branch | delete; compose the builder |
| e | `capability_service.go:132-139` — `CapsByUserID` → `AllCapabilities()` | delete; else `/auth/me` renders capabilities by a different rule than enforcement uses |

**Definition of done is mechanical, not an inventory** (ME-09): a lint asserting the literal
`'system_admin'` appears in no production Go or SQL outside the `role_catalog` seed. The list above is
a hand-synced enumeration and was already wrong once, by 40%, while reading as complete.

### 6.1 The audit disposition — mandatory ADR content

D2's claim that "the ordinary path already records it" is **false today**. `authz.Require:144-166`
records nothing; only the bypass arm INSERTs an audit event (`authz.go:125-133`, fail-closed in-tx,
ADR 0022 Phase 11 F8, whose stated rationale is that a tenant-wide bypass must appear in `audit.read`
at the same fidelity as a normal grant). Deleting `recordBypass` moves `system_admin` privilege use
from per-use audited to unaudited.

ADR 0092 must dispose of F8 explicitly, with the reconstruction query shown: after unification the
grant is durable provenance-carrying data, so "why was this actor authorized" is
(operation audit event) ⋈ (binding history) — which the old model could not answer at all. Silently
dropping the INSERT is how a Part 11 finding is born.

**Coupled decision, same ADR:** dropping the INSERT is what makes `authz.Require` legal in a
read-only transaction. Relaxing the `authz-require-rw-tx` lint is a decision, not a side effect —
make it here, not by drift. (Memory: `authz-require-readonly-tx-conflict`.)

## 7. Migration 0318 — one transaction

Order: create `role_catalog` + `capability_bindings` → seed the catalog → backfill → **assert** →
drop `iam_user_roles`, `iam_group_roles`, `user_process_areas`.

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

`iam_group_roles` has no CHECK, so it may hold role strings that exist nowhere and will violate the
new FK. They grant nothing today (they join to nothing in `role_capabilities`), so skipping them is
semantically safe — but it must be an explicit `WHERE role IN (SELECT code FROM role_catalog)` with a
**raised NOTICE and count**, never an accidental FK failure or a silent drop.

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

So the assert is three-part, and all three parts are materialized and compared:

- tier-1: **exact equality**, old = new
- tier-2: old ⊆ new
- tier-2: **new ∖ old = the enumerated expected-escalation set** (the tenant-scope conversions)

**And the real acceptance evidence is a seeded Go parity test**, which calls the actual old
evaluator functions and so dodges (1) and (2) entirely. Corpus must cover: direct role;
group-derived role; `iam_group_roles` junk strings; expired `effective_to` rows; area grants for each
of the 7 area roles; the 3 roles the `iam_user_roles` CHECK omits (`area_admin`, `qms_admin`,
`signer`); `system_admin` via both direct and group. Diff old-evaluator against new-evaluator for
every (actor, capability, area).

## 8. Firing mechanisms this program must land

Ranked per register doctrine. A program whose charter is deleting a construct is done when the
construct is unwritable — never when a list of instances is checked off.

| # | Mechanism | Level | Kills |
|---|---|---|---|
| F1 | `role_code` FK to `role_catalog`; Go sets + OpenAPI enums generated from it | 1 unrepresentable | ME-01 / #74 |
| F2 | Sole-reader lint: `capability_bindings` only inside the evaluator package | 3 red build | evaluator/read-model drift |
| F3 | Bypass-extinction lint: `'system_admin'` literal absent outside the catalog seed | 3 red build | ME-09 / #82 |
| F4 | `system_admin` bundle is generator output with a parity lint, or a boot assert vs the registry | 2–3 | advisory P2 stale bundle |
| F5 | `check ((scope_kind='tenant') = (scope_ref is null))` + area-retire revocation | 1 unrepresentable | dangling scope |
| F6 | Permanent property test: `∀cap. tier1(cap) ⊇ tier2(cap, ·)` | 3 red build | gate constraint 4 |
| F7 | `TestNoDeclaredOperationIsUnreachable` green | 3 red build | **the definition of done** |

F7 is gate constraint 3 and is load-bearing: the red CI lane goes green because the unreachable set is
empty, never because the test was skipped, excluded, or baselined.

## 9. ADR 0092 — required content

Mandatory (gate §10, ground 1 and 2):

1. **Supersedes ADR 0007's Decision.** The two-tier split stays as *enforcement points*; the
   disjoint-grant-tables ruling is withdrawn.
2. **MUST-deviation from REQ-AUTHZ-4** — "`system_admin` tenant-wide bypass is the only inheritance
   short-circuit" becomes false when there is no bypass at all. The target spec is amended in the
   same change, not left to drift.
3. **F8 audit disposition** with the reconstruction query (§6.1).
4. **`authz-require-rw-tx` lint relaxation** as an explicit coupled decision (§6.1).
5. **DD-6's role-catalog ruling** and its recorded inversion trigger (ME-10 / #83).
6. **What this guard does not claim** — per ME-02: the sole-reader and extinction lints prove the
   grant source is singular and the bypass is gone. They prove nothing about whether any individual
   `role_capabilities` bundle is *correct*.

## 10. Out of scope, named so it is not smuggled in

- **ME-07 / #80** authentication build-vs-adopt — a separate study, post-v1.
- **ME-08 / #81** MFA — routed, and its immediate action (delete the coverage endpoint) is a
  standalone fix, not this program.
- **ME-06 / #79** relation-tuple engine — triggered only by per-document sharing.
- **ME-10 / #83** tenant-defined roles — triggered by the first customer requiring them.
- Frontend user-management beyond re-pointing to the unified surface.

## 11. Acceptance

1. `TestNoDeclaredOperationIsUnreachable` green; `./apps/...` restored to green in `test-full.yml`
   and `test-nightly.yml` **by the unreachable set being empty**.
2. Seeded Go parity test passes over the full corpus (§7.1).
3. F1–F7 all landed and firing.
4. Full integration ladder per the selective policy (touched packages + guard suites; `./...` because
   this touches db/platform).
5. ADR 0092 written with §9's six items; target spec amended; `wiki/modules/iam.md` and
   `wiki/architecture/backend-target-architecture.md` updated.
6. Issues #74, #76, #82 closed by the landed mechanisms; #75 closed if F1 subsumes it.

## 12. Open items

None blocking. DD-5 and DD-6 closed the two operator decisions the advisory raised.
