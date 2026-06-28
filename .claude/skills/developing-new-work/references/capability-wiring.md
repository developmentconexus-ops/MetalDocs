# Capability wiring — add a capability (10 touchpoints, ordered)

**Last verified:** 2026-06-28

A new IAM capability is wired across 10 surfaces. Miss one and you get either a silently-unreachable
route, a privilege escalation, or a red guard test. Walk them in order. Anchors are for targeted
verify only.

1. **const + `validCapabilities`** — declare the typed capability constant and add it to the registry.
   `internal/modules/iam/domain/model.go:90` (consts) / `:134` (`validCapabilities`).

2. **scope classify** — every capability is `ScopeTenant` or `ScopeArea`. Classify it.
   `internal/modules/iam/domain/capability_scope.go:36`.

3. **tier-1 route→cap rule** — map the new route to the capability in the tier-1 middleware rules.
   `apps/api/cmd/metaldocs-api/permissions.go`. **Forgetting this is silent privilege escalation** —
   default route visibility is `VisibilitySessionRequired`, so an unmapped route is reachable by any
   authenticated session.

4. **tier-2 in-tx enforcement** — call `authz.Require(ctx, tx, cap, areaCode)` inside the business tx,
   after `authz.SeedTxIdentity`. Pattern: `internal/modules/templates/application/create.go:67` (sig `iam/authz/authz.go:76`; ScopeTenant passes areaCode `"tenant"`).

5. **seed grants** — grant the capability to the roles that should hold it (system_admin bypasses).
   `db/reference-data/0001_product_reference_data.sql:17`.

6. **DB tripwire** — the format/legacy constraints must accept the new cap name.
   `db/baseline/0001_current_schema.sql` (`ck_cap_format`, `ck_cap_not_legacy`).

7. **guard tests stay green** — `TestEveryCapabilityClassified`, `TestAreaGradeCapabilitySet`.
   `internal/modules/iam/domain/capability_scope_test.go`.

8. **bump `TestCapabilityRegistrySize`** — increment the `const want` count. This is the one mandatory
   manual edit. `internal/modules/iam/domain/model_test.go:90`. **Targeted-verify the current count
   here** before stating the new value.

9. **CI capability-coherence (5-surface)** — the const/classify/tier-1/seed/test surfaces must agree;
   governed by REQ-AUTHZ-5 in `wiki/architecture/backend-target-architecture.md`.

10. **H-PRE-1** — never call an authz-recording read (e.g. `authz.Require`) inside a lock-holding
    atomic tx; hoist it off-tx. Deadlock constraint.

**Net:** "who can do X" is always "who holds capability X". If the design ever says "the editor role
can…", it has already violated invariant 1 — stop and reframe in capabilities.
