# ADR 0060 — Audit events are append-only by database GRANT, not by trigger or RLS

- **Status:** Accepted
- **Last verified:** 2026-07-02
- **Date:** 2026-07-02
- **Scope:** Records the existing, deliberate mechanism by which `metaldocs.audit_events` is append-only: a database privilege grant (`GRANT INSERT`, no `UPDATE`/`DELETE`), not a trigger or a row-level-security policy. Also records the port-and-adapter shape (`Writer`/`Reader` as plain Go interfaces, same concrete `postgres.Writer` satisfying both, no tx-accepting variant). Closes tech-debt T-011 (`wiki/modules/audit-tech-debt.md`).
- **Depends on:** none.

---

## Context

Three load-bearing decisions in the `audit` module are encoded in code with no ADR backing: (a) `Writer`/`Reader` as plain Go ports with no transaction-accepting variant, (b) append-only enforced by `GRANT INSERT` only — not a trigger, not RLS, (c) the same concrete type satisfies both ports. Each is defensible individually but none was documented; future refactors (tamper-evidence hardening, tenant scoping, tx-bundled emit) need this ADR as a baseline to evaluate changes against.

### Verified runtime facts

- **Append-only is enforced by GRANT, not trigger/RLS.** `archive/migrations/0005_grant_workflow_audit_privileges.sql`:
  ```
  GRANT UPDATE ON TABLE metaldocs.documents TO metaldocs_app;
  GRANT INSERT ON TABLE metaldocs.audit_events TO metaldocs_app;
  ```
  The application's runtime DB role (`metaldocs_app`) is granted `INSERT` on `audit_events` and nothing else on that table — no `UPDATE`, no `DELETE`. There is no trigger rejecting UPDATE/DELETE on `audit_events` (unlike, e.g., `user_process_areas`, which has an explicit `trg_user_process_areas_no_delete` trigger — `db/baseline/0001_current_schema.sql:3967-3970` — a different module's belt-and-suspenders choice that `audit_events` does not mirror) and no RLS policy scopes or blocks mutation on it either. Append-only-ness is a consequence of the role's privilege set, full stop — a superuser or a role with broader grants is not blocked by anything at the schema level.
- **Ports are plain Go interfaces, no explicit-tx variant.** `internal/modules/audit/{domain,application,delivery,infrastructure}/` — `Writer` and `Reader` ports take no `*sql.Tx` parameter; callers cannot bundle an audit write into an existing caller-owned transaction through these ports.
- **Same concrete instance satisfies both ports.** DI wiring confirms a single `*postgres.Writer` instance is injected into both the `Writer` and `Reader` slots (verified via `_artifacts/03-deps.md` §3 DI wiring for the audit module).

## Decision

**Audit-event immutability is enforced exclusively by PostgreSQL `GRANT`: the application's runtime role holds `INSERT` on `metaldocs.audit_events` and no other DML privilege on that table.** This is the binding mechanism today — it is intentionally *not* a trigger and *not* an RLS policy. The two ports (`Writer`, `Reader`) are plain Go interfaces with no transaction-accepting variant, and the same concrete `postgres.Writer` type is wired to satisfy both — audit reads and writes go through one adapter, not two.

Binding rules for future work:
1. **Do not add UPDATE/DELETE grants on `audit_events`** to `metaldocs_app` or any application runtime role. If a legitimate redaction/right-to-erasure need arises (e.g. regulatory), that is a new ADR proposing a specific mechanism (e.g. a superuser-only maintenance path, not an application-role grant), not a silent grant widening.
2. **GRANT-based enforcement is accepted as sufficient for now**, on the basis that the only code path capable of writing to `audit_events` is the application's own `metaldocs_app`-authenticated connection, and that role's privilege set is itself under migration-controlled change review. A trigger or RLS policy would add defense-in-depth against a compromised/misconfigured role but is not required by this ADR; if that defense-in-depth is wanted later, propose it as an amendment or successor ADR, weighing it against the module's stated minimalism.
3. **`Writer`/`Reader` stay transaction-free ports** unless a concrete cross-module transactional-outbox need for audit emission is identified — until then, audit writes are fire-and-forget from the caller's perspective, consistent with an audit trail that must survive even if the caller's own transaction later rolls back for unrelated reasons (an audit write inside the caller's tx would be undone by that rollback, which is not always the desired semantic for "this action was attempted/authorized").
4. **One concrete adapter serving both ports is accepted**, not treated as a coupling defect — splitting into two adapter instances buys nothing while there is only one storage backend.

## Consequences

- T-011 (`wiki/modules/audit-tech-debt.md`) is closed by this ADR.
- Anyone proposing to widen `metaldocs_app`'s grants on `audit_events`, or to add a transactional-write variant to the audit ports, must treat that as an architecture change requiring a new/amending ADR, not a routine migration.
- No migration, schema change, or code change is required by this ADR — it documents and binds existing, verified runtime behavior.

## References

- `archive/migrations/0005_grant_workflow_audit_privileges.sql` — the `GRANT INSERT ... TO metaldocs_app` statement that is the entire enforcement mechanism.
- `internal/modules/audit/{domain,application,delivery,infrastructure}/` — `Writer`/`Reader` port shape.
- `db/baseline/0001_current_schema.sql:3967-3970` — contrasting example (`user_process_areas` no-delete trigger) showing a different module chose trigger-based enforcement; cited to make explicit that `audit_events`'s GRANT-only approach is a deliberate, not accidental, difference.
- `wiki/modules/audit-tech-debt.md` T-011 — tech-debt row closed by this ADR.
