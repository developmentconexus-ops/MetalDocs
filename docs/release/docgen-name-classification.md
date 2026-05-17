# Docgen Name Classification

Phase 7 classification for release-facing `docgen-v2` / `docgen_v2` / `DocgenV2` names.

## Decision

Classification: `defer` with compatibility prerequisite.

Do not rename docgen service names in this PR. The current names are not just generated API labels; they cross deployment, environment, package, persisted metadata, and worker event boundaries.

## Evidence

| Surface | Examples | Classification | Rationale |
|---|---|---|---|
| Service/workspace path | `apps/docgen-v2`, `@metaldocs/docgen-v2`, Dockerfile paths | deployment compatibility | Renaming affects workspace/package resolution and image build paths. |
| Compose service | `deploy/compose/docker-compose.yml` service `docgen-v2` | deployment compatibility | API/worker env points at service DNS names in compose. |
| Environment variables | `DOCGEN_V2_*`, `METALDOCS_DOCGEN_V2_*` | runtime compatibility | Existing `.env`, compose, docs, and scripts depend on these names. |
| Go client/types | `DocgenV2Client`, `LoadDocgenV2Config` | runtime compatibility | Can be renamed only with env alias support and broad caller update. |
| Persisted/export metadata | `docgen_v2_ver`, `docgen_v2_version`, values like `docgen-v2@0.4.0` | persisted compatibility | Composite hashes and export cache semantics include the renderer version string. |
| Worker events | `docgen_v2_pdf` | persisted event compatibility | Existing outbox rows and worker dispatch require read compatibility. |
| S3 defaults | `metaldocs-docx-v2` | storage compatibility | Bucket rename requires migration/copy or alias strategy. |

## Required Follow-Up Before Rename

1. Define canonical service name, env aliases, package/workspace rename, Docker service rename, and migration path.
2. Keep read compatibility for existing env vars, outbox event types, export metadata, and bucket names.
3. Emit canonical names from new producers only after compatibility readers are in place.
4. Verify API, worker, docgen service tests, compose startup, PDF export, freeze/fanout, and existing outbox replay.

## Current PR Scope

This PR may rename generated API operation IDs and templates runtime objects, but it must not rename `docgen-v2` service/deployment/persisted names without the compatibility work above.