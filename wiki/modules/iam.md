# Module: iam — LEGACY BOUNDARY REFERENCE

> **Status:** LEGACY / target boundary being replaced
> **Marked:** 2026-08-14

The current IAM module mixes people/tenant administration, groups, roles/grants, authorization evaluation and some tenant-lifecycle responsibilities. That runtime still exists, but the module boundary is not target authority.

Target concepts are being separated into:

```text
Organization
  Tenant / Area / User / Group / GroupMembership

Authorization
  Permission / Role / RoleAssignment / Check + scope filtering
```

Locked V1 roles: `tenant_owner`, `area_manager`, `author`, `approver`, `viewer`. Groups are flat principals receiving ordinary RoleAssignments. Tenant/Area scopes are typed. No role bypass. No OpenFGA/SpiceDB requirement in V1.

Read:

- `wiki/architecture/cohesive-platform-redesign.md`
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`

Do not continue historical `system_admin`, `area_admin`, `qms_admin`, `signer`, `editor`, dual-grant-source or magic-area-sentinel designs. Detailed old IAM architecture remains in Git history for migration archaeology.
