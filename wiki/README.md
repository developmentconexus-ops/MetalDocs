# MetalDocs Wiki

> **Last verified:** 2026-05-03
> **Purpose:** Single source of truth for codebase knowledge. Read this first - drill into folders only after.

## How to use this wiki

- **Humans:** Browse by folder.
- **AI agents:** Read this index, then `Glob wiki/**/*.md` to discover. Each doc has `Last verified:` + `Key files:` block at the top - use those file:line anchors instead of re-grepping.
- **Drift policy:** When changing code referenced by a doc, update the doc's `Last verified` stamp. Stale stamps = trust nothing in that doc until verified.

---

## Index

### Bug Tracker
- [bugs/audit-2026-05-03.md](bugs/audit-2026-05-03.md) - **40 bugs in 5 groups (A–E)** found in pre-smoke deep audit 2026-05-03. Group A (8 blockers) in-flight; B–E queued. **Read this before starting any new session.**

### Vision
- [vision/product-vision.md](vision/product-vision.md) - what MetalDocs is, problem it solves (stub, Last verified: 2026-05-01)
- [vision/target-users.md](vision/target-users.md) - quality engineers, ISO-bound orgs, document control roles (stub, Last verified: 2026-05-01)

### Architecture
- [architecture/system-overview.md](architecture/system-overview.md) - services, ports, end-to-end flow
- [architecture/data-model.md](architecture/data-model.md) - Postgres tables, key relationships, document_families (global/is_active), metaldocs_app grants (stub, Last verified: 2026-05-02)
- [architecture/tech-stack.md](architecture/tech-stack.md) - Go, React, Postgres, MinIO, Gotenberg, eigenpal (stub, Last verified: 2026-05-01)
- [architecture/deployment.md](architecture/deployment.md) - Docker compose, env vars, dev setup (stub, Last verified: 2026-05-01)

### Modules (one per backend module / frontend feature)
- [modules/templates-v2.md](modules/templates-v2.md) - template authoring, schemas, versioning, approval (stub, Last verified: 2026-05-01)
- [modules/documents.md](modules/documents.md) - document instances, editing flow, session model, API; backend module `internal/modules/documents/`, table `public.documents` (Last verified: 2026-05-03)
- [modules/taxonomy.md](modules/taxonomy.md) - document families (global), profiles, areas; CRUD routes, scoping distinction, deactivation guards (Last verified: 2026-05-02)
- [modules/approval.md](modules/approval.md) - approval routes, signoffs, ISO segregation, idempotency store, known gaps D4/E4/outbox/revision-number (Last verified: 2026-05-02)
- [modules/render-fanout.md](modules/render-fanout.md) - DOCX -> PDF rendering, substitution engine (stub, Last verified: 2026-05-01)
- [modules/iam-rbac.md](modules/iam-rbac.md) - capabilities, roles (viewer/editor/author/approver/system_admin) + process-area roles (signer/area_admin/qms_admin), DB-backed CanDo, area-scoped authz.Require, group grants, migration 0162-0166 + 0169 + 0170 (Last verified: 2026-05-03)
- [modules/editor-ui-eigenpal.md](modules/editor-ui-eigenpal.md) - eigenpal integration layer, controlled package, plugin wiring (Last verified: 2026-05-01)
- [modules/search.md](modules/search.md) - cross-module search; v2 reader JOINs `controlled_documents` to populate `DocumentCode` (stub, Last verified: 2026-05-01)

#### documents snapshot note

Snapshot columns (`placeholder_schema_snapshot`, etc.) are populated at document creation by `application.SnapshotService`, wired via `documents.Dependencies.SnapshotReader`/`SnapshotWriter`. The `enforce_snapshot_on_submit_trg` trigger enforces these are non-NULL before draft -> under_review.

`public.documents_v2` was the W1 scaffold table (migration 0103); dropped by migration 0168. Use `public.documents` for all queries.

### Concepts (cross-cutting)
- [concepts/placeholders.md](concepts/placeholders.md) - **CRITICAL:** fixed 7-token catalog, substitution at freeze; composition system deprecated 2026-04-27 (Last verified: 2026-04-27)
- [concepts/token-syntax.md](concepts/token-syntax.md) - `{name}` vs `{{uuid}}` - why it matters
- [concepts/controlled-documents.md](concepts/controlled-documents.md) - code generation, profile binding, sequence counters (stub, Last verified: 2026-05-01)
- [concepts/iso-segregation.md](concepts/iso-segregation.md) - why submitter cannot approve own submit (stub, Last verified: 2026-05-01)
- [concepts/freeze-and-hashing.md](concepts/freeze-and-hashing.md) - content_hash, values_hash, schema_hash, immutability (stub, Last verified: 2026-05-01)
- [concepts/authz-tiers.md](concepts/authz-tiers.md) - two-tier authz model: tier-1 CapabilityService (HTTP middleware) vs tier-2 authz.Require (in-tx area check); GUC setup, pitfalls (Last verified: 2026-05-03)

### Workflows (end-to-end flows)
- **[workflows/user-onboarding.md](workflows/user-onboarding.md)** - full user journey, non-technical: taxonomy -> template -> profile binding -> CD -> fill-in -> approval -> freeze -> PDF (Last verified: 2026-05-02) **Start here for QA/smoke testing.**
- [workflows/template-authoring.md](workflows/template-authoring.md) - create -> edit schema -> submit -> approve (stub, Last verified: 2026-05-01)
- [workflows/document-fillin.md](workflows/document-fillin.md) - pick CD -> wizard -> editor -> fill placeholders (stub, Last verified: 2026-05-01)
- [workflows/approval.md](workflows/approval.md) - submit, route, signoffs, idempotency; atomic finalize+submit, eligible_actor_ids fix, PostgresSignoffIdempStore (Last verified: 2026-05-02)
- [workflows/freeze-and-fanout.md](workflows/freeze-and-fanout.md) - approve -> freeze -> fanout -> PDF artifact

### Decisions (ADRs)
- [decisions/0001-eigenpal-adoption.md](decisions/0001-eigenpal-adoption.md) - why we picked eigenpal over CKEditor/BlockNote
- [decisions/0002-zone-purge.md](decisions/0002-zone-purge.md) - why we removed editable zones (2026-04-25)
- [decisions/0003-token-syntax-migration.md](decisions/0003-token-syntax-migration.md) - plan to move from `{{uuid}}` -> `{name}` (stub, Last verified: 2026-05-01)
- [decisions/0007-two-tier-authz.md](decisions/0007-two-tier-authz.md) - accept two distinct authz tiers (CapabilityService vs authz.Require); no schema migration needed (Last verified: 2026-05-03)
- [decisions/0008-placeholder-fixed-catalog.md](decisions/0008-placeholder-fixed-catalog.md) - replace user-fill placeholders with fixed 7-token computed catalog (2026-04-26)

### References
- [references/eigenpal-spike.md](references/eigenpal-spike.md) - pointer to spike repo + key findings (T1-T8)
- [references/eigenpal-controlled-package.md](references/eigenpal-controlled-package.md) - current vendored EigenPal package contract for MetalDocs
- [references/environment-setup.md](references/environment-setup.md) - local dev: compose, migrations, seed (stub, Last verified: 2026-05-01)
- [references/how-to-run-tests.md](references/how-to-run-tests.md) - Go tests, frontend vitest, e2e playwright (stub, Last verified: 2026-05-01)
- [references/local-dev-startup.md](references/local-dev-startup.md) - **START HERE** - PS script, port, credentials, common mistakes
- [references/local-dev-credentials.md](references/local-dev-credentials.md) - admin login details, DB access

### Glossary
- [GLOSSARY.md](GLOSSARY.md) - placeholder, zone (deprecated), fanout, freeze, eigenpal, controlled doc, profile, etc.

---

## Conventions

**Filename:** kebab-case, descriptive. ADRs prefix with 4-digit number.

**File header (every doc):**
```markdown
# Title

> **Last verified:** YYYY-MM-DD
> **Scope:** what this covers
> **Out of scope:** what it doesn't (link to where it does)
> **Key files:**
> - `path/to/file.go:42` - anchor description
> - `path/to/other.tsx:115` - anchor description
```

**Cross-refs:** Use full path + line numbers. Example: `internal/modules/templates_v2/application/schema.go:42`.

**Length:** Hard cap ~300 lines. Split if longer.

**Code blocks:** Always with language tag for highlighting.
