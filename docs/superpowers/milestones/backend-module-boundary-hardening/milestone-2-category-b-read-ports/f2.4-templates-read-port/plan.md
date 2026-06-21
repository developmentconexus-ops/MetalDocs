# Feature F2.4 — Plan — templates placeholder-schema read-port

> Input: `spec.md` (approved). Engine: subagent-driven-development run inline (TDD).

## Plan

### Files touched
- **Edit** `internal/modules/templates/domain/template_version_port.go` — add `PlaceholderSchema(ctx,
  tenantID, versionID) ([]byte, error)` to `TemplateVersionPort`.
- **Edit** `internal/modules/templates/infrastructure/template_version_reader.go` — implement
  `PlaceholderSchema` (SELECT v.placeholder_schema FROM templates_template_version v JOIN
  templates_template t ON t.id=v.template_id WHERE v.id=$1 AND t.tenant_id=$2::uuid; ErrNoRows→nil).
- **Edit** `internal/modules/documents/application/fillin_service.go` — `TemplateVersionSchemaReader`
  gains `tplVersions templatesdomain.TemplateVersionPort`; `NewTemplateVersionSchemaReader(db,
  tplVersions)`; `LoadFillInSchema` resolves versionID from `documents` (own table), calls the port,
  unmarshals; delete the cross-module JOIN.
- **Edit** `internal/modules/documents/module.go` — `Dependencies.TemplateVersionPort`; pass to
  `NewTemplateVersionSchemaReader`; nil→guard (skip the template-schema reader, as today when absent).
- **Edit** `apps/api/cmd/metaldocs-api/main.go` — set `docDeps.TemplateVersionPort =
  templatesinfra.NewTemplateVersionReader(deps.SQLDB)`.
- **New** `internal/modules/documents/application/fillin_schema_parity_integration_test.go` —
  `TestLoadFillInSchema_ParityWithRawJoin` (raw JOIN baseline vs port path).
- **Edit** `tools/cilint/internal/analyzers/hgcrossmodule.go` — remove the N1 entry.

### Ordering (parity-before-delete, D6)
1. Extend `TemplateVersionPort` + implement `PlaceholderSchema`.
2. Write the parity test: seed templates_template(+version w/ placeholder_schema) + a document
   pointing at it; raw `tv JOIN documents` vs new split path. RED→GREEN.
3. Rewire `TemplateVersionSchemaReader` (own-table version resolve + port call); wire module + main.
4. **Parity green** → delete the raw `templates_template_version` JOIN.
5. Remove N1 from `hgPendingRemediation`; `cilint` exit 0.
6. `go build ./...`; targeted tests; `grep` proof; `evidence.md`.

### Test strategy
- Real PG (:5434). Cases: present (doc→version with non-empty placeholder_schema array → equal
  non-nil `[]Placeholder`); absent document (docID unknown → nil); version with NULL/`'null'`
  placeholder_schema → nil. Compare raw-JOIN result vs `LoadFillInSchema` via the real port.
- The N1 reader has no unit-test callers; the new ctor arg only affects module.go + the parity test.

### Import-cycle / boundary check
- `documents/application → templates/domain` already exists (templatesdomain.Placeholder). Adding the
  port interface dep stays within that edge. Adapter injected from main (`templates/infrastructure`).
  Verify `go build`.

### Risks
- **Two reads vs one JOIN.** Splitting doc→version and version→schema must preserve nil-on-missing for
  both the absent-document and absent-version cases. Parity test's absent-document case locks it.
- **Tenant-scope divergence.** Bounded, documented in spec; parity test seeds valid (same-tenant) data.
