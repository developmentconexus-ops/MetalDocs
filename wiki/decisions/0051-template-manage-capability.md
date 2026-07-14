# ADR 0051 — `template.manage` capability for published-template approval-config governance

- **Status:** Accepted
- **Last verified:** 2026-06-30
- **Date:** 2026-06-30
- **Scope:** Introduces the `template.manage` IAM capability and removes the role-string `isOperator` guard from `UpsertApprovalConfig`, bringing that path into full ADR 0022 compliance.
- **Supersedes:** The role-string `isOperator` check in `internal/modules/templates/application/approval_config.go` (pre-ADR 0051 defect).

---

## Context

### The defect (F-CD7)

`UpsertApprovalConfig` (the service method behind `PUT /templates/{id}/approval-config`) contained a pre-tx role-string guard:

```go
isOperator := containsRole(cmd.ActorRoles, string(iamdomain.RoleSystemAdmin)) ||
    containsRole(cmd.ActorRoles, string(iamdomain.RoleQmsAdmin))
if hasEverPublished {
    if !isOperator { return nil, domain.ErrForbidden }
} else {
    if template.CreatedBy != cmd.ActorUserID && !isOperator {
        return nil, domain.ErrForbidden
    }
}
```

This violates **ADR 0022 invariant 1** ("AuthZ = capabilities, never roles"): authorization is decided by inspecting role strings rather than by an in-tx `authz.Require` capability check. REQ-AUTHZ-5 (capability-coherence 5-surface) requires that const, scope, tier-1, seed, and tests all agree; the role-string path was outside all five surfaces.

The method already contained a correct in-tx `authz.Require(CapTemplateEdit, "tenant")` check at the base gate — the role-string block was an additional layer layered on top of it without using the PDP.

### Why a new capability rather than reusing an existing one

`template.edit` is held by `author` and `editor` roles — it is not an operator-level gate. The elevated right to reconfigure a *published* template's approval policy is a distinct, higher-privilege action. Reusing `template.edit` would widen the blast radius of that cap beyond its current semantics. A dedicated `template.manage` cap is the correct PDP expression.

---

## Decision

1. **Add `CapTemplateManage Capability = "template.manage"`** to `internal/modules/iam/domain/model.go`, `validCapabilities`, and `capability_scope.go` (ScopeTenant — all `CapTemplate*` are tenant-grade).

2. **Rewrite `UpsertApprovalConfig`** — delete the `isOperator` role-string block; introduce `requireManage bool` (pre-tx domain flag); in-tx call `authz.Require(CapTemplateManage, "tenant")` when `requireManage` is true, after the existing `CapTemplateEdit` check. Drop `ActorRoles []string` from `UpsertApprovalConfigCmd` (no longer read).

3. **Elevation semantics preserved:**
   - Published template → `requireManage = true` always (no creator shortcut — governance integrity of a published config warrants operator authority).
   - Unpublished template, creator → `requireManage = false` (domain-ownership shortcut, not a role check).
   - Unpublished template, non-creator → `requireManage = true`.

4. **Tier-1 (permissions.go):** the approval-config route already maps to `CapTemplateEdit` at tier-1 (line 137 at time of writing). `template.manage` is an **elevation within the same route** enforced at tier-2 only — consistent with ADR 0022 Phase 11 precedent (route.manage). No new tier-1 row is added; the route is already guarded (not left unmapped).

5. **Seed grant:** `qms_admin` receives `template.manage` in `db/reference-data/0001_product_reference_data.sql`. `system_admin` accesses via tier-2 bypass (not seeded explicitly — as per the existing convention for bypass-eligible roles).

6. **Registry guard:** `TestCapabilityRegistrySize` bumped 33 → 34 (`model_test.go`). `TestEveryCapabilityClassified` and `TestAreaGradeCapabilitySet` pass without change because the new cap is classified in `capability_scope.go`.

---

## Consequences

### Positive
- Removes the only remaining role-string authz reasoning in the templates module.
- `UpsertApprovalConfig` is now fully governed by the PDP (two in-tx `authz.Require` calls, both audited via the asserted-caps GUC).
- `ActorRoles` is no longer carried through the approval-config command struct, reducing surface for accidental role-based reasoning.
- REQ-AUTHZ-5 capability-coherence is maintained: const / scope-classify / tier-1 (verified present) / seed / guard tests all updated.

### Negative / trade-offs
- The route-level tier-1 capability for approval-config remains `CapTemplateEdit` (not `CapTemplateManage`). This means a caller with only `template.edit` reaches the handler but is denied at tier-2 for published templates. This is the correct pattern for elevation-within-a-route (ADR 0022 Phase 11 precedent); the alternative (adding a separate route) would require a contract change.
- `qms_admin` is the only non-bypass role receiving `template.manage` today. If future roles legitimately need this power, a new seed row and an ADR revision are required.

---

## References

- ADR 0022 (authz = capabilities, never roles) — governing boundary.
- REQ-AUTHZ-5 — capability-coherence 5-surface guard.
- `internal/modules/templates/application/approval_config.go` — implementation.
- `internal/modules/iam/domain/model.go` — const + registry.
- `internal/modules/iam/domain/capability_scope.go` — scope classification.
- `apps/api/cmd/metaldocs-api/permissions.go:137` — tier-1 mapping (verified present, unchanged).
- `db/reference-data/0001_product_reference_data.sql` — seed grant.
- `internal/modules/iam/domain/model_test.go` — registry size guard (bumped 33→34).
