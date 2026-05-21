# Scheduled Supersede Runtime Prerequisite Repair Design

Date: 2026-05-20
Status: proposed
Classification: runtime prerequisite
Scope: repair migration-stream truth for scheduled supersede schema so normal supported bootstrap/startup can reach the governed `documents.superseded_document_id` field

## Goal

Repair the deploy/runtime prerequisite discovered during review:

- scheduled supersede code now depends on `documents.superseded_document_id`
- the schema change currently exists only as `migrations/0203_documents_scheduled_supersede.sql`
- supported bootstrap/startup consumes `db/migrations/*.sql`, not the root `migrations/` folder
- version `0203` is already taken in the real post-baseline stream

Result: runtime code can rely on a column that normal environments never receive.

This design fixes that prerequisite without widening into route, scheduler, frontend, or authz work.

## Confirmed Runtime Truth

- Fresh supported bootstrap uses curated baseline plus the `db/migrations/*.sql` forward tail.
- Current real tail under `db/migrations/` ends at `0208_documents_schedule_generation.sql`.
- The stray root-level migration file currently contains the intended schema change and self-records version `0203`, which conflicts with the real `db/migrations/0203_rename_templates_v2_objects.sql`.

## Root Cause

The scheduled supersede schema change was authored in the wrong migration stream and with a reused version number. That created a false sense that the runtime prerequisite existed when the supported bootstrap path could never apply it.

## Recommended Approach

Use the `minimal plus cleanup` repair shape:

1. add one real post-baseline forward migration at the live tail
2. remove the misleading active-looking root migration artifact
3. clean the nearest stale design truth that still describes the wrong schedule-publish request contract

This remains intentionally bounded to the prerequisite slice.

## Repair Boundary

### In Scope

- `db/migrations/0209_documents_scheduled_supersede.sql`
- retirement of `migrations/0203_documents_scheduled_supersede.sql`
- nearby design-truth cleanup in `docs/superpowers/specs/2026-05-20-scheduled-supersede-hardening-design.md`
- focused verification that the supported curated path lands the column and records the migration version

### Out of Scope

- scheduler cutover behavior changes
- OpenAPI edits
- generated code regeneration
- frontend or TanStack Query changes
- authz alignment
- broad wiki sweep unrelated to the prerequisite repair

## Migration Design

Create `db/migrations/0209_documents_scheduled_supersede.sql` as a forward-only, idempotent post-baseline migration.

Required effects:

- add `public.documents.superseded_document_id uuid REFERENCES public.documents(id)` when absent
- add an index for non-null supersede targets
- write one `public.schema_migrations` row for version `0209`

Recommended SQL shape:

- `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`
- `CREATE INDEX IF NOT EXISTS`
- `INSERT INTO public.schema_migrations (version, description) ... ON CONFLICT (version) DO NOTHING`

Why this shape:

- it works on the supported path
- it stays safe if a local environment was manually patched earlier
- it preserves forward-only post-baseline migration discipline

## Stray Artifact Handling

Remove `migrations/0203_documents_scheduled_supersede.sql` from the repo rather than keeping it as an executable-looking file.

Reason:

- the filename and location imply active migration truth
- keeping it increases the chance of future drift or duplicate application attempts
- the real historical evidence will exist in the new governed migration and git history

If later archaeology is needed, git history is the source, not a second executable-looking migration path.

## Nearby Design Truth Cleanup

Update `docs/superpowers/specs/2026-05-20-scheduled-supersede-hardening-design.md` so it no longer claims that `POST /api/v1/documents/{id}/schedule-publish` must carry `content_hash`.

Reason:

- current runtime/contract truth for the schedule-publish slice has already moved away from that request requirement
- leaving the stale statement in place invites another round of prerequisite or contract drift

This cleanup is documentation-only and does not widen the implementation boundary.

## Verification

Use the smallest supported gate that proves runtime truth through the curated path:

1. confirm `0209` is the next real file in `db/migrations/`
2. run a curated bootstrap/migration verification that exercises `db/migrations/*.sql`
3. verify `public.documents` contains `superseded_document_id`
4. verify `public.schema_migrations` contains version `0209`
5. verify the adjacent design note no longer contains the stale `content_hash` request claim for schedule-publish

Recommended commands:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-bootstrap.ps1 -WithDevSeed
```

If the full bootstrap gate is too expensive for this moment, a smaller accepted fallback is:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 -WithDevSeed
```

## Risks And Handling

### Manual local drift already exists

If a developer manually applied the stray root migration earlier, the real `0209` migration must still succeed cleanly. The idempotent DDL and `schema_migrations` insert behavior handle that.

### More stale docs exist elsewhere

If additional stale references are discovered outside the immediate adjacent design note, record them as follow-up unless they directly block the prerequisite boundary. Do not turn this repair into a broad documentation sweep.

### Broader bootstrap break appears

If the curated bootstrap gate fails for reasons outside this migration, stop and classify that as a larger runtime prerequisite instead of silently widening scope.

## Success Criteria

1. Supported bootstrap/startup now reaches `documents.superseded_document_id`.
2. `public.schema_migrations` records version `0209`.
3. The repo no longer presents the root `migrations/0203_documents_scheduled_supersede.sql` file as active executable migration truth.
4. The nearest local design note no longer contradicts current schedule-publish request truth.

## Implementation Notes For The Next Plan

- This prerequisite repair should land before any registry, publish-flow, scheduler, or frontend follow-up that assumes the supersede column exists.
- The later implementation plan may cite this spec as the prerequisite that restores runtime trust for scheduled supersede persistence.
