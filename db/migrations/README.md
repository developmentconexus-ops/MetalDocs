# Post-Baseline Forward Migrations

Use this directory for forward-only database changes after the curated baseline cutoff.

Rules:

- Do not add local dev users or demo data here.
- Every migration must insert one `public.schema_migrations` row.
- Every migration must be safe to run exactly once.
- Destructive changes require ADR approval, rollback notes, and a maintenance-window plan.
- Update `wiki/database/` when a table, column, function, trigger, reference data rule, or ownership boundary changes.
