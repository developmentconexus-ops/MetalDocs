# Security Module — Architecture Notes

**Last verified:** 2026-06-12 (Wave Z, Z-8 boundary disposition)
**Owner:** Admin Center Sessions & Security tab.
**Related docs:**
- [`wiki/modules/security-signals.md`](security-signals.md) — signal inventory and endpoint detail
- [`wiki/modules/security-tech-debt.md`](security-tech-debt.md) — open tech-debt items
- [`wiki/decisions/0027-rls-adoption-sequencing.md`](../decisions/0027-rls-adoption-sequencing.md) — auth_identities tenant-global by design (T-008)
- [`wiki/backend/legacy-register.md`](../backend/legacy-register.md) — F-06 entry (cross-module boundary register)

---

## Accepted boundary: JOIN to iam_users for tenant scoping (F-06 residual, Z-8)

**Decision: accepted boundary — do not refactor.**

`internal/modules/security/infrastructure/postgres/repository.go` contains multiple queries
(`ListLockouts`, `CountRecentFailedLoginsByUser`, `ListNewDeviceLogins`) that JOIN
`metaldocs.auth_identities` (or `metaldocs.auth_sessions`) to `metaldocs.iam_users` on
`u.user_id = i.user_id WHERE u.tenant_id = $1`.

This JOIN is the **tenancy mechanism itself**, not an accidental cross-module read.

### Why the JOIN is structurally required

`auth_identities` is tenant-global by design (ADR 0027, T-008 closed as by-design,
`wiki/decisions/0027-rls-adoption-sequencing.md` §1). The table holds identity credentials;
the `user_id` is associated with exactly one tenant via `iam_users.tenant_id`. There is no
`tenant_id` column on `auth_identities` — adding one would be an incorrect data model.
The JOIN is therefore the only correct way to scope any auth_identities query to a caller's
tenant. This is the same pattern documented in ADR 0027:

> Tenant scoping of the identity happens via `JOIN auth_identities ON iam_users.user_id =
> auth_identities.user_id`, not by a `tenant_id` column on `auth_identities`.

The same applies to `auth_sessions`, which carries `tenant_id` directly on its own rows;
queries against it scope by `s.tenant_id = $1` and JOIN `iam_users` only to resolve
`display_name`.

### display_name reads

The queries also project `u.display_name` from the same JOIN. These display-field reads are
opportunistic — they ride a JOIN that is already required for tenant scoping. Extracting a
separate `iam` port for display-name resolution would require a second round-trip per query
or a batch N+1 pattern, with no correctness benefit. The `LoginContextPort` pattern
(commit `07f914e9`, F-06c) applies to cases where auth writes to an IAM-owned column
without a required structural JOIN; it does not apply here because the JOIN to `iam_users`
is mandatory for correctness regardless of the display field.

### F-06 legacy-register entry

The F-06 register (`wiki/backend/legacy-register.md:228`) lists this as:

> security module cross-tenant isolation via JOIN to iam_users only (no tenant_id on
> auth_identities) | `internal/modules/security/infrastructure/postgres/repository.go:84-100`
> | security → auth table structure

This entry is **closed as accepted boundary** by this note. The JOIN is the intended
cross-module seam, not a violation. No port extraction is required.
