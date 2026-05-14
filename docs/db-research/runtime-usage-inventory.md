# Runtime Usage Inventory

## Commands

```powershell
rg -n "FROM |JOIN |INSERT INTO|UPDATE |DELETE FROM|CALL |SELECT .* FROM|ExecContext|QueryContext|QueryRowContext" apps internal -S
rg -n "schema_migrations|migrate.Apply|RunStartupMigrations|Backfill|METALDOCS_SKIP_STARTUP_MIGRATIONS|METALDOCS_MIGRATIONS_DIR" apps internal scripts deploy -S
```

## Startup and Migration Mutation Paths

- `apps/api/cmd/metaldocs-api/main.go:140-145`: startup applies migrations unless `METALDOCS_SKIP_STARTUP_MIGRATIONS=true`; optional `METALDOCS_MIGRATIONS_DIR`.
- `internal/platform/migrate/migrate.go:85`: reads `public.schema_migrations`.
- `apps/api/cmd/metaldocs-api/main.go:213` + `internal/modules/registry/module.go:49-50`: registry startup backfill path.
- `internal/modules/registry/application/migration.go:19,62,78`: backfill mutates `controlled_documents` and `documents` with advisory lock.

## Runtime Table Usage By Area

- auth/iam: `metaldocs.auth_identities`, `metaldocs.auth_sessions`, `metaldocs.iam_users`, `metaldocs.iam_user_roles`
  - Evidence: `internal/modules/auth/infrastructure/postgres/repository.go:30,49,62,109,149,201,252-253`; `internal/modules/iam/infrastructure/postgres/role_provider.go:22,38`; `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:49,59,65`.
- audit/governance: `metaldocs.audit_events`, governance events stream
  - Evidence: `internal/modules/audit/infrastructure/postgres/writer.go:46,55,78,134`; `apps/api/cmd/metaldocs-api/main.go:414`.
- templates: `templates_v2_template`, `templates_v2_template_version`, `templates_v2_approval_config`, `templates_v2_audit_log`
  - Evidence: `internal/modules/templates/repository/postgres.go:42,182,452,473,523`.
- documents/approval: `documents`, `document_revisions`, `approval_instances`, `approval_stage_instances`, `approval_signoffs`, `approval_routes`, `approval_route_stages`
  - Evidence: `internal/modules/documents/repository/repository.go:108,127,142,190,481,800,971`; `internal/modules/documents/approval/repository/postgres_approval_repository.go:34,97,127,446,471,494`.
- registry: `controlled_documents`, grants tables, `cd_sequence_counters`
  - Evidence: `internal/modules/registry/infrastructure/repository.go:33,338,383,391,490,504,527`.
- taxonomy: `metaldocs.document_profiles`, `metaldocs.document_process_areas`, `metaldocs.document_families`
  - Evidence: `internal/modules/taxonomy/infrastructure/repository.go:26,116,200,287`; `internal/modules/taxonomy/infrastructure/family_repository.go:32,88,107`.
- jobs/outbox: `metaldocs.job_leases`, `metaldocs.pdf_dispatch_outbox`
  - Evidence: `internal/modules/jobs/scheduler/lease_reaper.go:19,51`; `internal/modules/render/fanout/pdf_outbox_repository.go:33,46,52,74,100`.

## Candidate Exclusions

No safe runtime exclusions found in this pass.

## Open Decisions

- Classify `governance_events` under audit or a separate governance dictionary scope.
- Decide whether registry startup backfill remains runtime maintenance or becomes governed migration artifact.
