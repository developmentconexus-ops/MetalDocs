# ADR 0092 — AuthZ Grant-Model Unification: one binding relation, scope on the binding

> **Status:** Accepted (operator-ratified 2026-08-06); D1 schema partially implemented 2026-08-11 (A8.1 — see Status history below). Remaining scope (A8.2+ query-builder repoint, D2–D4) is not implemented — no further code/schema change is authorized by this ADR alone. Filed 2026-08-09 (Phase G governance reconciliation); records the 2026-08-06 ruling, does not reopen it.
> **Decision source:** `docs/superpowers/analysis/2026-08-06-authz-grant-unification-decisions.md` (operator-ratified decision record, D1–D4 + red-lane ruling) and `docs/superpowers/analysis/2026-08-06-authz-grant-unification-system-impact.md` (system-impact gate output). Those documents are the original analysis; this file is the numbered ADR 0092 decision artifact that ADR 0093 and issue #89 forward-reference.
> **Scope:** The grant/assignment side of authorization — which tables record who holds which role at which scope, and which single source both PDP tiers read.
> **Out of scope:** The capability model itself (roles → capability bundles via `role_capabilities` stays — ADR 0022); the two-tier PDP *shape* (two enforcement points stay — ADR 0007's tiers as defense-in-depth); tier-1's data source for route→capability mapping (ADR 0090); bounded-context boundaries (ADR 0093).
> **Relations:**
> - **ADR 0007** — partially superseded: the framing that the disjoint grant tables are "distinct tiers … not a unification gap" is withdrawn (see Context); the two *enforcement points* (edge + in-tx + DB tripwire) remain in force.
> - **ADR 0022** — unchanged: authz remains capabilities, never roles, at every decision point.
> - **ADR 0093** — sequencing dependency: both advisory arms in the 0093 review held the grant-model axis (this ADR) **must not be shelved behind** the Controlled Information merge. Execution owned by issue #89/A8, Phase 2b of the remediation roadmap (`docs/superpowers/analysis/audit-2026-08-09/final-synthesis.md` §I).

## Context

Tier-1 and tier-2 read **disjoint grant tables**:

| Tier | Function | Reads | Role CHECK |
|---|---|---|---|
| 1 — HTTP edge | `CapabilityService.CanDo` (`internal/modules/iam/application/capability_service.go`) | `iam_user_roles` ∪ `iam_group_members` ⋈ `iam_group_roles` | 5 roles / none at all |
| 2 — in-tx | `authz.Require` (`internal/modules/iam/authz/authz.go`) | `role_capabilities` ⋈ `user_process_areas` | 7 roles, incl. `area_admin`, `qms_admin`, `signer` |

There are **six role-declaration surfaces** (`iamtypes.validRoles`, `iamtypes.areaRoles`, OpenAPI `UserRole`, OpenAPI `AreaRole`, the `user_process_areas` CHECK, the `iam_user_roles` CHECK) plus `iam_group_roles` with no CHECK. The `iam_user_roles` CHECK of 5 is mirrored by no Go set and no OpenAPI enum; the three roles it omits are exactly `area_admin`, `qms_admin`, `signer` — and that unmirrored table is what tier-1 reads. Consequence: a principal holding only an area membership is refused at tier-1 before tier-2 ever runs; the area model is unreachable through HTTP except to narrow someone tier-1 already admitted. Surfaced by `TestNoDeclaredOperationIsUnreachable` (http-surface-protocol Task 17, red by design).

ADR 0007's own Context records that the 2026-05-02 unification "left `authz.Require` reading `user_process_areas`"; its Decision then ruled the result "distinct tiers … not a unification gap." That ruling ratified an incomplete migration retroactively as architecture — an unlabelled local maximum under CLAUDE.md's Global Maximum rule. This ADR replaces that framing.

**What is NOT wrong and stays:** two enforcement points (edge rejects early and cheap; in-tx is the binding decision; DB tripwire is the last line) — that is defense in depth. The defect is that the two points answer **different questions against different tables**, free to disagree. `role_capabilities` (role → capability bundle) also stays.

## Decision (operator-ratified 2026-08-06)

### D1 — One assignment relation, scope on the binding

Collapse `iam_user_roles`, `user_process_areas`, `iam_group_roles` + `iam_group_members` into a single binding relation carrying `(subject_kind, subject_id, role_code, scope_kind, scope_ref)` plus the existing effective-interval and grant/revoke provenance columns. One role catalog referenced by FK (adding a role becomes data, not a migration). Model precedent: Kubernetes RBAC — scope lives on the binding (`RoleBinding` referencing a `ClusterRole` inside one namespace), not in a parallel table.

The two tiers become two **predicates over one source**:
- tier-1: capability granted at **any** scope → cheap edge reject;
- tier-2: capability granted at `tenant` **or** at the resource's area.

Target invariant (the missing property that caused the finding): **tier-1 is a strict relaxation of tier-2** — testable with one source, accidental with two.

### D2 — No bypass: `system_admin` holds every capability as ordinary grants

The special branch in both tiers (+ `recordBypass` audit path) is removed. `system_admin` becomes a role like any other, bound at `scope=tenant`, whose bundle is the full capability set (derivable from the capability registry). A bypass is a second authorization regime; making it a grant makes audit uniform and "who can do X?" one query.

### D3 — No direct per-user capability grants

Roles only. Direct grants collapse auditability ("who holds `document.publish`?" stops having a cheap complete answer), have no reviewable unit, make revocation search-and-destroy, and reintroduce a second grant regime. Recorded trade-off: the genuine one-off is served by **creating a narrow role and binding it** — D1 makes roles cheap (an INSERT, not a migration).

### D4 — Groups and areas are orthogonal; both survive

Group = **who** (a subject). Area = **where** (a scope). Under D1 they are different columns of one binding; binding a group to an area becomes expressible.

### Standing ruling — the red CI lane stays red

`TestNoDeclaredOperationIsUnreachable` stays red in **whichever CI lane runs the `./apps/...` integration suite** until this program lands (operator ruling 2026-08-06). The ruling is workflow-name-independent: at materialization time (post-PR-#97 consolidated CI) that lane is `.github/workflows/nightly.yml`; the ruling as originally filed named `test-full.yml`/`test-nightly.yml`, which are the historical pre-PR-#97 workflow names. Do NOT skip, exclude, or baseline the test; do NOT remove `./apps/...` from the lane. The lane goes green when the unreachable set is empty — that is the fix.

## Consequences

- Issue #89/A8 owns execution (dual-source unrepresentable; tripwire parity suite; data migration across the three assignment tables; user-management frontend). Roadmap slot: Phase 2b, after the #87/A1 verifier spine.
- ADR 0007 carries a status-history entry pointing here; ADR 0093's forward-reference to "0092 (pending)" now resolves to this file.
- The DB tripwire and the capability×area in-tx check remain the enforcement backbone during and after migration; migration is reversible until the old relations are dropped.
- Reproduced-current evidence for the dual-source state at baseline `418070bf`: `docs/superpowers/analysis/audit-2026-08-09/final-synthesis.md` (G-08) and `pass03-modules-identity-access.md`.

## Status history

- **2026-08-06** — decisions D1–D4 + red-lane ruling ratified by the operator (`2026-08-06-authz-grant-unification-decisions.md`); ADR number 0092 reserved, file not yet created.
- **2026-08-09** — materialized as this filed ADR during Phase G governance reconciliation (architecture audit #100 / PR #101). Recording only; no alternative reopened, no implementation authorized.
- **2026-08-11** — A8.1 amendment: D1's schema shipped (migrations 0318/0319) with a named deviation from the literal column list below. See "Amendment 2026-08-11 (A8.1 shipped shape)" below in this document for the full text.

## Amendment 2026-08-11 (A8.1 shipped shape)

> **Last verified (amendment):** 2026-08-11. Does not reopen or alter D1's decision (one binding relation,
> scope on the binding) — it documents that the physical schema issue #89/A8.1 shipped is not a literal
> column-for-column match of D1's prose, and records why.
> Source: `db/migrations/0318_capability_bindings_schema_backfill.sql` (see its "NAMED DEVIATION FROM THE
> ADR'S LITERAL COLUMN LIST" section) and `wiki/database/tables/capability_bindings.md`.

D1 (above) names a single logical column: `(subject_kind, subject_id, role_code, scope_kind, scope_ref)`.
`metaldocs.capability_bindings`, as created by migration 0318, does **not** have a `subject_id` column.
Postgres cannot express one FK column that conditionally targets two different tables of two different
types (`iam_users.user_id` is `text`; `iam_groups.id` is `uuid`), so `subject_id` is physically split into
two nullable columns, discriminated by `subject_kind`:

- `subject_user_id text` — FK → `metaldocs.iam_users(tenant_id, user_id)`, `NOT NULL` iff `subject_kind = 'user'`
- `subject_group_id uuid` — FK → `metaldocs.iam_groups(tenant_id, id)`, `NOT NULL` iff `subject_kind = 'group'`
- `CONSTRAINT chk_capability_bindings_subject_shape` — enforces exactly one of the two is set, matching `subject_kind`

This is read as a **stronger** implementation of D1's own intent than a single polymorphic `subject_id`
column could deliver: a dangling subject reference is a FOREIGN KEY VIOLATION at the database line, not
merely an application-validated invariant — "unrepresentable by construction" is more literally true of the
shipped shape than of the ADR's own prose. The same reasoning gave `scope_ref` a matching
`chk_capability_bindings_scope_shape` CHECK (tenant-scope binding cannot name an area; area-scope binding
must). D1's logical relation — one binding, subject × role × scope on one row — is unchanged; only its
physical column list differs from the ADR text above.

A future reader diffing this ADR against the live schema should treat this Amendment, not line 31's column
list, as the decision of record for the physical shape. D2–D4 and the query-builder repoint (A8.2+) remain
undecided-in-schema and out of scope for this amendment.

**A8.1's boundary, stated explicitly (do not infer from "D1's schema shipped" above):** A8.1 implements the
binding-table **shape** (`metaldocs.capability_bindings`, `metaldocs.roles`) and a **grant backfill only** —
`iam_user_roles`, `user_process_areas`, and `iam_group_roles` rows are copied into the new relation, with
`source_relation` provenance, for future reads to consume. D1 also names `iam_group_members` as one of the
four relations D1's binding collapses; A8.1 does **not** touch it. `iam_group_members` (group *membership* —
who is in a group) remains the sole membership authority through A8.1; no group-membership row is
represented in `capability_bindings`, and nothing reads `capability_bindings` for membership. "D1's logical
relation … is unchanged" above describes the *shape* of the binding relation `capability_bindings` was built
to hold, not a claim that every source D1 names has been migrated into it. A later slice, not A8.1, decides
whether/how group membership joins this relation.

**`capability_bindings` is NOT AUTHORITATIVE through A8.1 (TRANSITIONAL — read before trusting this table):**
the relation is populated only by 0318's one-time backfill. The three write sites that issue grants today
(`role_admin_repository.go`, `user_area_repository.go`, `onboard_tenant_service.go`) still write exclusively
to `iam_user_roles`, `user_process_areas`, and `iam_group_roles` — none dual-writes into
`capability_bindings`. Every grant issued after the backfill runs is therefore invisible in
`capability_bindings`, and the relation starts drifting stale from the moment A8.1 merges. This is harmless
only because A8.1 ships no read path against it (see the amendment above). It becomes a correctness hole the
moment any read path is repointed at it — no read path may consult `capability_bindings`, and it must not be
treated as authoritative, until a write cutover (dual-write or repoint of the three sites above) lands. That
cutover is not owned by any slice in the canonical A8.1–A8.4 decomposition as of this writing; it is tracked
separately. Until it lands, `iam_user_roles`, `user_process_areas`, and `iam_group_roles` remain the sole
grant source of record.
