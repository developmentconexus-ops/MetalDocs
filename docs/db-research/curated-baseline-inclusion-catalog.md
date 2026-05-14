# Curated Baseline Inclusion Catalog (Task 8)

## Scope and Classification

- Classification: `runtime prerequisite` + `workflow/tooling gap`
- Phase boundary: research synthesis only (no DB artifact implementation changes)
- Inputs: Phase 1 Tasks 3-7 research agent evidence
- Approval status: approved for implementation on 2026-05-14

## Commands Used (Coordinator)

```powershell
git status --short --branch
Test-Path docs/superpowers/specs/2026-05-14-curated-database-baseline-and-dictionary-design.md
Test-Path scripts/dev-db-reset.ps1
Test-Path scripts/dev-bootstrap-baseline.ps1
Test-Path migrations_baseline/0001_baseline_2026_05.sql
Test-Path deploy/compose/docker-compose.yml
```

## Evidence Sources

- Runtime usage: `apps/api/cmd/metaldocs-api/main.go:140-145,213,414`, `internal/platform/migrate/migrate.go:85`, `internal/modules/**` SQL repositories (auth/iam/audit/templates/documents/approval/registry/taxonomy/jobs/render).
- Migration archaeology: `migrations/*.sql` inventory (`total=181`, `ledger_insert_files=34`, `begin_files=77`), range `0001..0201` with gaps `0078-0100`, `0119`, `0178-0179`.
- Live schema: `metaldocs-postgres` healthy; `public` tables=27, `metaldocs` tables=41; extensions include `pgcrypto`; `schema_migrations` has 34 rows (`0112..0201`) and no baseline marker.
- Seed/auth classification: `migrations/0158_fix_process_area_role_constraint.sql`, `migrations/0159_seed_dev_approver_user.sql`, `migrations/0170_dev_approver_role_correction.sql`, capability/template seeds, bootstrap admin runtime path.
- Governance: `AGENTS.md`, `CLAUDE.md`, `wiki/architecture/data-model.md`, design spec indicate DB governance/dictionary gaps and source-of-truth contradiction.

## Inclusion Decisions (Curated Baseline)

### Include in `db/prerequisites/`

- Required extensions and foundational DB prerequisites:
  - `pgcrypto` (live extension evidence + migration dependencies including auth/audit paths).

### Include in `db/baseline/` (schema + invariants only)

- Runtime-used table families and required constraints/indexes/triggers/functions for:
  - auth/iam (`auth_identities`, `auth_sessions`, `iam_users`, `iam_user_roles`)
  - audit/governance (`audit_events`, `governance_events` if currently runtime-used)
  - templates v2 (`templates_v2_template`, `templates_v2_template_version`, `templates_v2_approval_config`, `templates_v2_audit_log`)
  - documents + approval (`documents`, `document_revisions`, approval tables, status/state invariants)
  - registry (`controlled_documents`, grants, sequence counters)
  - taxonomy/process model (`document_profiles`, `document_process_areas`, `document_families`)
  - jobs/outbox (`job_leases`, `pdf_dispatch_outbox`, lease helper function[s])
- `public.schema_migrations` table definition (policy-compliant baseline+tail ledger model to be finalized in implementation phase).

### Include in `db/reference-data/` (product reference only)

- Capability mappings required by runtime authorization surface:
  - role capability seed sets from current canonical capability names.
- System template reference data required for product correctness:
  - system blank template baseline records (`migrations/0199_system_blank_template.sql` lineage).
- Any immutable, product-wide taxonomy/process reference records confirmed as non-dev and non-tenant-specific.

### Include in `db/dev-seeds/` (local/dev only)

- Dev user/bootstrap-only records and convenience data:
  - dev approver seed lineage (`0159`, `0170`), dev tenant role conveniences (`0158`), local e2e seed helpers.
- Explicitly local bootstrap admin conveniences when applicable to local-only flows.

## Exclusions from Curated Fresh Baseline

- Historical repair mechanics and destructive cleanup steps that only existed to transition prior states.
- Legacy-only/drop/destroy cleanup migrations once resulting final state is represented directly in curated schema.
- Direct replay dependence on Docker entrypoint-mounted historical migration chain.

## Defer / Needs Explicit Approval or Policy Clarification

- Baseline marker semantics in `schema_migrations`:
  - live DB currently numeric-only (`0112..0201`), plan expects marker like `baseline-2026-05-14`.
- Duplicate migration numeric prefixes (`0042`, `0070`, `0130`) ordering semantics in legacy replay policy.
- Classification of mixed data sets:
  - `0029_seed_metal_nobre_document_registry.sql` (product reference vs tenant-specific seed split)
  - `0055/0057/0066` template default seeds (product invariant vs environment-opinionated seed)
- Governance event stream classification:
  - treat under audit family or separate governance surface for dictionary ownership.

## Coordinator Decisions for Task 9 Approval

Approve before Phase 3 implementation:

1. Baseline inclusion boundary above is accepted (schema/runtime surfaces + separated reference/dev seed model).
2. Mixed seed files will be split by intent (product reference vs local dev) without patching historical migrations.
3. Ledger policy adopts explicit baseline+tail semantics (including whether baseline marker row is required).
4. Governance updates will establish `wiki/database/` as canonical DB policy/dictionary location.

## Runtime Backfill Decision

| Backfill | Decision | Reason | Verification |
|---|---|---|---|
| `registry.BackfillLegacyDocuments` | `documented-legacy-maintenance-job` | `internal/modules/registry/application/migration.go` only mutates rows where `documents.controlled_document_id IS NULL`; fresh curated baselines should not have these legacy-null rows. `internal/modules/registry/module.go` now exposes this as legacy maintenance intent. | Fresh curated bootstrap + API startup should complete without durable backfill mutation; backfill logs should indicate no-op/zero processed on clean baseline. |

## Unresolved Questions to Carry into Phase 3 (Post-Approval)

1. Should `governance_events` be owned in audit dictionary scope or separate governance scope?
2. Is production guaranteed to avoid local-only seed artifacts from the historical chain today?
3. Should bootstrap admin be hard-disabled outside local by policy/code guard?
4. Are skipped versions (`0178`, `0179`, `0198`) and missing ranges documented as intentional in legacy replay policy?
