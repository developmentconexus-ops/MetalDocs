# Live Schema Inventory

## Commands

```powershell
docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -c "select current_database(), current_user;"
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -P pager=off -c "select count(*) as tables from information_schema.tables where table_schema in ('public','metaldocs') and table_type='BASE TABLE';"
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -P pager=off -c "select extname, extversion from pg_extension order by extname;"
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -P pager=off -c "select count(*), min(version), max(version) from schema_migrations;"
```

## Inventory Summary

- `metaldocs-postgres` container healthy.
- Schemas present include `public` and `metaldocs`.
- Base tables observed: `public=27`, `metaldocs=41`.
- Public-side object counts observed by agent:
  - columns `289`
  - constraints `324`
  - indexes `80`
  - triggers `32`
  - functions `53`
- Extensions: `plpgsql`, `pgcrypto`.
- `schema_migrations` exists with `34` rows from `0112` to `0201`.

## Drift Signals

- Live ledger is numeric-only (`0112..0201`) with no baseline marker row yet.
- Duplicate numeric migration prefixes exist in historical files (`0042`, `0070`, `0130`) and need explicit legacy ordering policy.
- Skipped versions (`0178`, `0179`, `0198`) require explicit legacy replay documentation as intentional.

## Open Questions

- Canonical ordering source for duplicate numeric prefixes during legacy replay.
- Baseline marker policy (`baseline-2026-05-14`) adoption timing for curated path.
