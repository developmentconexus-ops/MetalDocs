# System-impact analysis — AuthZ grant-model unification

**Date:** 2026-08-06
**Intent (one line):** Collapse `iam_user_roles`, `iam_group_roles` + `iam_group_members`, and `user_process_areas` into one binding relation carrying scope, so tier-1 and tier-2 answer one predicate over one source.
**Work type:** feature (a large one — no new module is born; `iam` already owns every surface touched)
**Author:** developing-new-work skill
**Verdict:** 🟡 **Yellow** — proceed to design, with four locked constraints and one mandatory ADR (§10)

> Inputs already ratified by the operator: `docs/superpowers/analysis/2026-08-06-authz-grant-unification-decisions.md` (D1–D4).
> This artifact does not re-open those decisions. It judges what they cost the system.

---

## 1. Classify & own

- **Work type:** feature. No module is born. Every table, service and port involved already belongs to `iam`.
- **Owning module:** `iam`. It owns the capability model (`domain/model.go`), both evaluators (`application/capability_service.go` = tier-1, `authz/authz.go` = tier-2), all three grant tables, and the published ports other modules read grants through.
- **Explicitly NOT owning:**
  - `auth` — owns authentication (who you are), never authorization. It imports `iamtypes.Role` as a value type only; that edge is why `Role` was moved to `internal/platform/iamtypes` in the first place.
  - `approval` — the largest downstream consumer, but it *configures* against the role vocabulary; it does not own it.
  - `security` — owns origin protection and rate limiting, not grants.
- **Cross-module edges (`A → B` = A depends on B). All must stay on published interfaces:**

| Edge | Through what | Status |
|---|---|---|
| `approval → iam` | `iamtypes.IsAreaRole` / `AreaRoles()` for stage actor-selector validation (`approval/http/contracts/route.go:284`) | **Interface. Verified** — the only `user_process_areas` mention in that file is a comment; no cross-module SQL. |
| `approval → iam` | `AreaCapabilityReader` port — "in which areas does this actor hold capability X?" (`iam/domain/area_capability_reader_port.go:19`) | Interface. Read model for UI narrowing, explicitly *not* an authorization decision. |
| `taxonomy → iam` | area codes (`taxonomy/infrastructure/tenant_data_port.go`) | Interface. |
| `documents`, `controlleddocuments`, `distribution` → `iam` | `authz.Require` at tier-2 | The tier-2 call itself; unchanged in shape by this work. |
| `platform/tripwire → iam` | tripwire arms generated from the Go capability registry (GMR M2) | Generated, not hand-synced. |

- **Ambiguity?** None. **AS-3 does not fire.**

## 2. Foundation verdict

- **Base you'd build on:** three assignment tables read by two evaluators that never consult each other, plus a role catalog declared across six surfaces (§3 below).
- **Sound, or legacy/patch/workaround?** **Legacy, and self-documented as such.** ADR 0007's Context records verbatim that the 2026-05-02 IAM unification "unified the middleware path … but left `authz.Require` reading `user_process_areas`", then its Decision rules the split is "not a unification gap." An unfinished migration ratified retroactively as architecture — an unlabelled local maximum under CLAUDE.md's Global Maximum rule.
- **Would this work optimize *inside* the patch?** **No — it deletes it.** D1 replaces the three tables with one relation; the two tiers become two predicates over that one source. This is the global-maximum structure, not a tweak within the local one.
- **Therefore AS-2 does not fire.** Recorded explicitly because the trigger reads similar: the foundation *is* a patch (AS-2's first clause), but the work does not build inside it (AS-2's second clause), and both clauses must hold to fire. Had the plan been "seed the missing grants into `iam_user_roles`", AS-2 would fire and the verdict would be Red.
- **Trade-off of the global-maximum structure:** a destructive data migration across three tables holding live tenant grants, and a frontend that must learn one binding model instead of two membership screens. Bought with: one query answers "who can do X", scope becomes expressible on a group, and the tier-1 ⊆ tier-2 invariant becomes testable instead of accidental.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|---|---|---|---|
| AuthZ = capabilities, never roles | **Yes — centrally** | Roles stay what ADR 0022 already allows: the *assignment/grouping unit*. Enforcement still asks "holds capability X". `role_capabilities` (role → capability) is correct today and survives untouched. D1 changes only the assignment side. | `authz.Require`, `CapabilityService.CanDo` — both keep their signatures; only their FROM clause changes. |
| Contract-first (OpenAPI + oapi-codegen) | **Yes** | New/changed IAM routes for binding management go in `api/openapi/v1/openapi.yaml` first. `UserRole` (8) and `AreaRole` (7) enums are affected by the role-catalog change — see §7. | per-module `cfg.yaml` + `gen.go`; full regen (partial regen is forbidden drift). |
| Multi-tenant pooled | **Yes** | `capability_bindings` carries `tenant_id` like every tenant table; the tenant predicate goes on every query; tx-local GUCs only. | `tenant.FromContext`, `authz.SeedTxIdentity`. |
| Async = transactional outbox | No | Grant changes are synchronous state writes with no external side effect. If grant-change *notifications* are ever wanted, they go through the outbox — out of scope here. | — |
| DB enforces invariants | **Yes — and more than today** | Role catalog becomes a table referenced by **FK**, replacing three hand-synced CHECKs. `iam_group_roles`' complete absence of a CHECK is closed by construction. | existing tripwire pattern in `db/baseline/0001_current_schema.sql`. |
| Cross-module via published interface only | **Yes** | Every edge in §1 already goes through a port. `AreaCapabilityReader`'s contract ("which areas does this actor hold capability X in?") must survive the model change semantically intact — its callers narrow UI on it. | `iam/domain/area_capability_reader_port.go`. |

**No violation. AS-1 does not fire.** The work *restores* invariant 1's intent rather than bending it.

### 3b. Finding that emerged during targeted verify — the role catalog has six declaration surfaces

Not three, as the decisions document states. Correcting the record:

| Surface | Roles | Mirrored by |
|---|---|---|
| `iamtypes.validRoles` (`internal/platform/iamtypes/role.go:39`) | 8 | OpenAPI `UserRole` |
| `iamtypes.areaRoles` (`role.go:75`) | 7 | OpenAPI `AreaRole`, `user_process_areas` CHECK |
| OpenAPI `UserRole` (`openapi.yaml:4836`) | 8 | `iamtypes.validRoles` |
| OpenAPI `AreaRole` (`openapi.yaml:4845`) | 7 | `iamtypes.areaRoles` |
| `user_process_areas` role CHECK (`db/baseline/0001_current_schema.sql:1662`) | 7 | `iamtypes.areaRoles` |
| `iam_user_roles` role_code CHECK (`:1276`) | **5** | **nothing** |
| `iam_group_roles.role` (`:1245`) | any string | nothing |

The system canonically believes in 8 roles, 7 area-assignable, and four of the six surfaces agree in two consistent pairs. **The `iam_user_roles` CHECK of 5 is mirrored by no Go set and no OpenAPI enum** — and the three it omits are exactly `area_admin`, `qms_admin`, `signer`, the three the rest of the system treats as first-class. That table is what tier-1 reads.

Consequence for governance: **REQ-AUTHZ-5's five-surface CI binding covers capabilities, not roles.** No guard compares these seven surfaces, which is why the drift survived. D1's FK-to-a-catalog collapses all of them to one.

## 4. Capability wiring

**Partially N/A.** The work adds **no new capability** — the registry, the scope classification, and `role_capabilities` are all correct and untouched. Of the 10 touchpoints, only these are in play:

- **3. tier-1 route→cap** — unchanged in *shape*. ADR 0090 already made this a generated lookup (`httpSurface[pattern]`). What changes is the grant query behind `CanDo`, not the route→capability mapping.
- **4. tier-2 in-tx** — `authz.Require(ctx, tx, cap, areaCode)` keeps its signature. Its FROM clause changes from `role_capabilities ⋈ user_process_areas` to the binding relation.
- **5. seed grants** — **materially affected.** `db/reference-data/0001_product_reference_data.sql` must express D2: `system_admin`'s bundle becomes the full capability set, derived from the registry rather than hand-listed.
- **6. DB tripwire** — the subject-discriminated arms (ADR 0083) read the grant model; they must be re-pointed and re-verified. Generated from the Go registry since GMR M2, so this is a regeneration, not a hand edit.
- **8. `TestCapabilityRegistrySize`** — **no bump.** No capability is added. Stating this explicitly so the design does not touch it reflexively.
- **10. H-PRE-1** — **live and load-bearing here.** `authz.Require` records a system_admin bypass audit inside the caller's tx (`authz.go:119`). D2 deletes the bypass, which *removes* one H-PRE-1 hazard — but the replacement path must not introduce a new authz-recording read inside a lock-holding tx. Carry as a design constraint.

**Two unreachable capabilities close as a side effect:** `distribution.read` and `notification.read` hold no grant on any path today. They are the residue `TestNoDeclaredOperationIsUnreachable` reports.

## 5. Module wiring

**N/A** — no module is born. `iam` exists, owns all fifteen-module-boundary-compliant surfaces involved, and its wiki doc (`wiki/modules/iam.md`) already carries the canonical-roles section this work rewrites.

## 6. Frameworks to reuse, not reinvent

| Primitive | Reuse confirmed |
|---|---|
| `TxRunner.Do` | Yes. Note the standing constraint: `authz.Require` needs a **writable** tx (`DoReadOnly` rejects the bypass audit INSERT); the `api-lint` rule `authz-require-rw-tx` enforces it. D2's removal of the bypass may make the read-only case newly legal — **do not assume it; the lint is the contract, and relaxing it is its own decision.** |
| `tenant.FromContext` | Yes — never thread tenant id by hand into binding queries. |
| `authz.SeedTxIdentity` | Yes — unchanged; the tx-local GUCs the tripwire reads are orthogonal to which table holds the grant. |
| `authz.Require` | Yes — reused, not replaced. Signature stable. |
| `problem.New` / `Write` | Yes — binding-management endpoints return RFC 9457. |
| `audit.NewEvent` / `RecordTx` | Yes. D2 makes this *simpler*: with no bypass, `recordBypass`'s separate audit path disappears and the ordinary grant path records everything. |
| Outbox repo | N/A — no external side effect. |
| `testdb` factory + `SeedWithCaps` | **Yes, and it is a first-class migration target.** `SeedWithCaps` seeds today's grant tables; every integration suite in the repo depends on it. Re-pointing it is the single highest-leverage step of the whole program — it converts hundreds of tests at once. |

No hand-rolled equivalent is proposed. No genuinely new cross-cutting concern appeared, so no new platform framework is warranted.

## 7. Contract & data

**OpenAPI-first.** Binding management is IAM route surface: create/revoke a binding, list bindings for a subject, list subjects for a role. The `UserRole` (8) and `AreaRole` (7) enums become derived from the role catalog rather than hand-written. Full regen — partial regen is forbidden drift ([[openapi-embedded-spec-regen-churn]]).

**Migration.** Next number is **0318** (`db/migrations/` currently ends at `0317_template_version_content_hash_always.sql`). `capability_bindings` carries `tenant_id`; the role catalog gets a table; the three CHECKs are replaced by one FK.

**Destructive change — expand/contract is mandatory, not optional.** These tables hold live tenant grants; a one-step cutover risks locking real users out of a regulated eQMS.

1. **Expand** — create `capability_bindings` + `role_catalog`; dual-write; backfill from all three legacy tables. Legacy tables still authoritative for reads.
2. **Verify** — assert the new relation reproduces every grant the old evaluators would have returned, per tenant. The migration is provable: for every (actor, capability, area) the old two evaluators can be replayed against the new source.
3. **Contract** — flip both evaluators to the binding relation; drop the legacy tables and their CHECKs; delete the dual-write.

The backfill has a **semantic decision the design must make explicitly, not by default**: a row in `iam_user_roles` today means "tenant-wide at tier-1, invisible to tier-2." Under D1 it becomes `scope_kind='tenant'`, which is *strictly more* than it was — tier-2 will now honor it where it previously ignored it. That is a privilege change during migration and must be deliberate.

## 8. Test & QA plan

- **Canonical framework:** `testdb` integration factory, `//go:build integration`, R1–R4. `SeedWithCaps` is re-pointed rather than duplicated.
- **QA gates that apply:** contract (routes change), **authz (the whole point)**, multi-tenant isolation (bindings are tenant-scoped and must not leak), DB-invariant (FK replaces three CHECKs), docs. **N/A:** async/idempotency — no outbox path.
- **The acceptance test already exists and is already failing.** `TestNoDeclaredOperationIsUnreachable` is red by operator decision. **It going green is this program's definition of done** — not a test to fix, but the acceptance criterion, and the reason the red lane must not be skipped, excluded, or baselined.
- **The invariant to add as a permanent guard:** tier-1 must be a strict relaxation of tier-2. Assert it as a property over the binding relation — for any (actor, capability, area), tier-2 granting implies tier-1 granting. Today this is unrepresentable as a test because the two read different tables.
- **Evidence shape:** `go build ./...`, `go build -tags integration ./...`, `go vet -tags integration ./...`, `go test ./...`, the integration ladder, `.\scripts\check-system-runnable.ps1`, plus a **live authz drive** — compile ≠ works is a recorded lesson from GMR M2, and this program is exactly the class where a non-functional wiring passes every static check.
- **Bit QR-C:** after any seam signature change, `go vet -tags integration` before commit — untagged `go test` does not compile integration files.

## 9. Docs / ADR

- **Wiki:** `wiki/modules/iam.md` (the canonical-roles section is rewritten), `wiki/modules/iam-tech-debt.md` (T-010 and the grant-model debt close), `wiki/architecture/backend-target-architecture.md` (REQ-AUTHZ-4 — see below), and the `Last verified` stamps on each.
- **REQ IDs cited:** REQ-AUTHZ-1 (capabilities never roles — preserved), REQ-AUTHZ-3 (area-grade caps get the real area — preserved and strengthened), **REQ-AUTHZ-4** (deny by default; `system_admin` bypass is the only short-circuit — **D2 deviates**), REQ-AUTHZ-5 (declaration surfaces CI-bound — extended to cover roles, see §3b), REQ-AUTHZ-7 (every deny audited — preserved), REQ-AUTHZ-8 / RF-3 (cache invalidation contract, still open — the binding model changes what invalidation must watch, so RF-3 should be closed *by* this program rather than left behind it).
- **ADR required? YES — two, and one is mandatory:**
  1. **Mandatory.** A new ADR that supersedes **ADR 0007**'s "distinct tiers, not a unification gap" ruling and amends **ADR 0022**. It must also record the **MUST-deviation from REQ-AUTHZ-4**: that requirement states the `system_admin` bypass *is* the only inheritance short-circuit; D2 removes the bypass entirely, making the sentence false as written. The target spec is amended in the same change, not left to drift.
  2. The ADR states the tier-1 ⊆ tier-2 invariant as a governed property, so a future reader cannot re-derive the split as intentional the way ADR 0007 did.

## 10. Verdict & locked constraints

**Verdict: 🟡 Yellow.** Proceed to design. The work fits the architecture, restores an invariant rather than bending one, and no hard-stop fires — but it carries a mandatory ADR, a MUST-deviation, and a destructive data migration, so it is not Green.

**Open hard-stops:** none.
- **AS-1** — does not fire. No invariant violated; invariant 1's intent is restored.
- **AS-2** — does not fire, and the reasoning is recorded in §2 because the trigger reads similar: the foundation *is* a patch, but this work replaces it rather than building inside it. The rejected alternative (seeding grants into `iam_user_roles`) *would* have fired it.
- **AS-3** — does not fire. `iam` owns unambiguously.

**Locked constraints handed to brainstorming:**

1. **Expand → verify → contract.** Live tenant grants; no one-step cutover. The backfill must state explicitly what `iam_user_roles` → `scope_kind='tenant'` does to tier-2 visibility, because it is a privilege change.
2. **The ADR is mandatory and supersedes ADR 0007**, amends ADR 0022, and amends REQ-AUTHZ-4 in the target spec in the same change. D2 makes REQ-AUTHZ-4's current wording false.
3. **`TestNoDeclaredOperationIsUnreachable` going green is the definition of done** — the acceptance criterion, not a test to repair. Do not skip, exclude, or baseline it; do not remove `./apps/...` from the integration lanes to restore green.
4. **Add the tier-1 ⊆ tier-2 property as a permanent guard**, and extend the REQ-AUTHZ-5 surface-coherence CI binding to cover the **role** catalog, which it does not cover today (§3b). Without this, the same class of drift returns in a different table.
5. **No new capability.** Do not touch `TestCapabilityRegistrySize`; do not add registry entries. `distribution.read` and `notification.read` become reachable through grants, not through new declarations.
6. **`SeedWithCaps` is the migration lever.** Re-point the `testdb` factory rather than writing a parallel seeding path; it converts the integration estate in one move and keeps the test-framework hard gate satisfied.
7. **H-PRE-1 stays live.** Removing the bypass removes one hazard instance; it does not retire the constraint. No authz-recording read inside a lock-holding tx.
8. **Preserve `AreaCapabilityReader`'s semantics.** Its contract is a UI-narrowing read model, explicitly not an authorization decision. Callers narrow on it; changing its meaning silently changes what users see.

**Handoff:** Green/Yellow ⇒ this analysis is the locked rail set for `superpowers:brainstorming`.
