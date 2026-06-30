# System-impact analysis — F-CD7: capability `template.manage` for approval-config governance

**Date:** 2026-06-29
**Intent (one line):** Replace the role-string `isOperator` guard in `UpsertApprovalConfig` with a dedicated capability `template.manage` (ADR-0022: authz = capability, never role).
**Work type:** feature
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow *(see §10)*

---

## 1. Classify & own

- **Work type:** feature (no new module; adds one capability + rewires one existing service method).
- **Owning module(s):** `templates` — owns `UpsertApprovalConfig` (`internal/modules/templates/application/approval_config.go`). `iam` — owns the capability registry/scope/tripwire surfaces a new cap must be wired through.
- **Explicitly NOT owning:** `controlleddocuments`/`documents` — they have their own approval governance; this is template-version approval config, templates' concern. `auth` — authn only, not capability authoring.
- **Cross-module edges (with direction):** `templates → iam` (already exists: templates imports `iam/authz` + `iam/domain` at approval_config.go:8-9 and calls `authz.Require(CapTemplateEdit,"tenant")` in-tx). New cap is consumed the same way — through iam's published `authz` pkg + `domain.Capability` consts. No new edge, no repo/SQL reach-in.
- **Ambiguity?** None. AS-3 not triggered.

## 2. Foundation verdict

- **Base you'd build on:** `UpsertApprovalConfig` already has the *correct* tier-2 base — `runner.Do` → `SeedTxIdentity` → `authz.Require(CapTemplateEdit,"tenant")` in-tx (`:73-79`). The defect is the **legacy pre-tx role-string patch** layered on top (`isOperator = containsRole(system_admin||qms_admin)`, `:36-46`).
- **Sound, or legacy/patch/workaround?** The in-tx authz base is sound (canonical). The `isOperator` block is the patch — it reasons in role strings, violating invariant 1, and adds a false-negative restriction (only operators may edit a published template's config) that is expressed in roles instead of a capability.
- **If patchy:** the global-maximum structure is the existing capability PDP — the elevated "may govern a published template's approval policy" right becomes a **capability** (`template.manage`), checked in-tx via the same `authz.Require` the method already uses. We are *removing* a patch and folding the rule into the proven framework, not optimizing inside the patch. AS-2 not triggered.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | **YES (core)** | Delete `isOperator` role-string block; gate elevated path with `authz.Require(CapTemplateManage,"tenant")` in-tx (second Require alongside CapTemplateEdit). Creator-of-unpublished shortcut stays as a domain-ownership rule layered over CapTemplateEdit, not a role check. | `authz.Require`, `authz.SeedTxIdentity` |
| Contract-first (OpenAPI + oapi-codegen) | No | Route `PUT /templates/{id}/approval-config` already exists; no route/DTO change. Tier-1 rule for that route maps to a template cap (see §4.3). | — |
| Multi-tenant pooled | No (unchanged) | Method already tenant-scoped (`cmd.TenantID`, `SeedTxIdentity`, GUCs). | `tenant.FromContext` (caller) |
| Async = transactional outbox | No | No network side effect; pure DB write + in-tx audit. | — |
| DB enforces invariants | **YES** | `ck_cap_format` (dotted lower) accepts `template.manage`; `ck_cap_not_legacy` unaffected. Seed grant row in reference-data. DB tripwire remains last line. | `db/baseline` constraints; `db/reference-data` seed |
| Cross-module via published interface only | No (already correct) | templates consumes iam via `authz` pkg + `domain.Capability` consts only. | iam `authz` published pkg |

No invariant violated. AS-1 not triggered.

## 4. Capability wiring

Concrete cap: **`template.manage`** (`CapTemplateManage`). 10 touchpoints:

1. **const + `validCapabilities`** — add `CapTemplateManage Capability = "template.manage"` at `iam/domain/model.go:~108` (after CapTemplateArchive) and to `validCapabilities` (`:134`).
2. **scope classify** — `CapTemplateManage: ScopeTenant` in `capability_scope.go:~60` (all `CapTemplate*` are ScopeTenant).
3. **tier-1 route→cap** — `permissions.go`: the approval-config route currently maps to a template cap; confirm/keep its mapping (likely `template.edit`). `template.manage` is enforced **tier-2 only** (an elevation within the same route), so no new tier-1 row is strictly required — but verify the route is not left unmapped. *(Locked: targeted-verify permissions.go entry for the approval-config route during impl.)*
4. **tier-2 in-tx** — `authz.Require(ctx, tx, string(CapTemplateManage), "tenant")` in `approval_config.go` for the elevated branch.
5. **seed grants** — `db/reference-data/0001_product_reference_data.sql`: grant `template.manage` to `qms_admin` (and any role that legitimately held the old operator power). `system_admin` bypasses.
6. **DB tripwire** — `ck_cap_format`/`ck_cap_not_legacy` in `db/baseline/0001_current_schema.sql` accept the name (dotted, non-legacy) — no migration to the constraint needed; reference-data seed is the change.
7. **guard tests green** — `TestEveryCapabilityClassified`, `TestAreaGradeCapabilitySet` stay green once §2 classify lands.
8. **bump `TestCapabilityRegistrySize`** — current `const want = 33` (`model_test.go:91`) → **34**. (Targeted-verified count = 33 on 2026-06-29.)
9. **CI capability-coherence (REQ-AUTHZ-5)** — const/classify/tier-1/seed/test must agree; this walk keeps all five aligned.
10. **H-PRE-1** — `UpsertApprovalConfig` takes **no FOR UPDATE lock**; the in-tx `authz.Require` is already used safely there. Adding a second Require in the same tx is safe. ✓

## 5. Module wiring

**N/A** — no new module. `templates` and `iam` both exist and are already wired.

## 6. Frameworks to reuse, not reinvent

- `TxRunner.Do` — already used (`approval_config.go:73`). Keep.
- `authz.SeedTxIdentity` + `authz.Require` — the capability check primitive. Reuse for the new cap; do **not** hand-roll a role/cap comparison.
- `audit.NewEvent`/in-tx append — already used (`newAuditEvent` + `AppendAuditTx`). Keep.
- `problem.Write` — error mapping at delivery is unchanged (`ErrForbidden` → 403 already wired). Keep.
- No new primitive needed.

## 7. Contract & data

- **OpenAPI-first:** no route/DTO change (existing `PUT /templates/{id}/approval-config`). `ActorRoles` may be dropped from the command struct once `isOperator` is gone — internal refactor, not a contract change (roles aren't a request field; they come from auth context).
- **Migration:** none to schema. The only data change is the **reference-data seed** granting `template.manage` to the appropriate role(s). Reference-data is applied via the existing seed path, not a forward migration of tenant data.
- **Destructive change?** No. Behavior tightens to capability semantics; expand/contract not needed (single deploy).

## 8. Test & QA plan

- **Canonical framework:** unit test for `UpsertApprovalConfig` (existing pattern in `templates/application/*_test.go`) covering: (a) holder of `template.manage` may edit a published template's config; (b) non-holder → `ErrForbidden`; (c) creator may configure own unpublished template without manage; (d) non-creator non-holder → `ErrForbidden`. Plus the iam guard tests (registry size 34, classified, area-grade).
- **QA gates that apply:** authz gate (capability coherence, tier-2 enforcement); the rest (contract, multi-tenant isolation new-table, async, DB-migration) **N/A** — no contract/table/async change.
- **Evidence shape:** `go build ./...`, `go test ./...` (esp. `iam/domain` registry guards + templates application), capability-coherence CI guard. Report outcomes + two-stage review disposition.

## 9. Docs / ADR

- **Wiki:** refresh `wiki/modules/templates.md` (approval-config authz section) + `Last verified`. Note the cap in the iam capability catalog doc if one is enumerated.
- **REQ IDs cited:** REQ-AUTHZ-5 (capability-coherence 5-surface), and the ADR-0022 authz boundary.
- **ADR required?** **YES** — a new capability is a change to the capability set governed by ADR 0022; the handoff mandates an ADR for F-CD7. New sequential ADR documenting `template.manage` (rationale: de-role the published-template approval-config governance; supersedes the role-string operator check). This is the Yellow flag.

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — fits cleanly within the existing PDP; proceeds with two locked constraints (ADR + registry bump).
- **Open hard-stops:** none. AS-1/AS-2/AS-3 not triggered.
- **Locked constraints handed to implementation:**
  1. New cap `template.manage` ⇒ **bump `TestCapabilityRegistrySize` 33 → 34** (`model_test.go:91`).
  2. **Write an ADR** (next sequential number) for the new capability under ADR-0022; cite REQ-AUTHZ-5.
  3. Walk all 10 capability touchpoints; keep the 5-surface coherence guard green.
  4. Replace `isOperator` with `authz.Require(CapTemplateManage,"tenant")` **in-tx** (mirror `:77`); keep the creator-of-unpublished domain shortcut over CapTemplateEdit, not a role check.
  5. Seed-grant `template.manage` to `qms_admin` (and prior operator-power roles); `system_admin` bypasses.
  6. Targeted-verify the `permissions.go` tier-1 mapping for the approval-config route is present (no silent escalation).
