# Feature F2.4 — Evidence — templates placeholder-schema read-port

> **Milestone:** M2 (category-b-read-ports) · **Feature:** `f2.4-templates-read-port` · **Closed:** 2026-06-21
> **Contract:** `spec.md` (consumer = documents fill-in schema loader; extend the existing ADR-0030
> `TemplateVersionPort` rather than mint a parallel reader; ADR-0039 D3(b) owner-published read-port).
> Census site closed: **N1** (`documents/application/fillin_service.go` → cross-module JOIN
> `templates_template_version tv JOIN documents d` reading `placeholder_schema`).

## What was implemented

By outcome:

- **Owner extends its existing port (no parallel reader).** `templates/domain.TemplateVersionPort`
  gains `PlaceholderSchema(ctx, tenantID, versionID string) ([]byte, error)` —
  `internal/modules/templates/domain/template_version_port.go`. documents depends only on this
  interface; it never names `templates_template_version` in its own SQL.
- **Owner-side adapter.** `templates/infrastructure.TemplateVersionReader.PlaceholderSchema`
  (`internal/modules/templates/infrastructure/template_version_reader.go`) runs
  `SELECT v.placeholder_schema FROM templates_template_version v JOIN templates_template t ON
  t.id = v.template_id WHERE v.id=$1 AND t.tenant_id=$2::uuid` (tenant-scoped via the owner's
  `templates_template`). `ErrNoRows → (nil, nil)`.
- **Consumer split, cross-module JOIN deleted.** `documents/application.TemplateVersionSchemaReader`
  gains `tplVersions templatesdomain.TemplateVersionPort`;
  `NewTemplateVersionSchemaReader(db, tplVersions)`. `LoadFillInSchema` now (1) resolves
  `template_version_id` on the **documents-owned** `documents` table
  (`SELECT template_version_id FROM documents WHERE id=$1 AND tenant_id=$2`, ErrNoRows→nil,nil),
  then (2) reads the schema through `tplVersions.PlaceholderSchema(...)`, then (3) the existing
  `json.Unmarshal` into `[]templatesdomain.Placeholder`. The old
  `templates_template_version tv JOIN documents d` JOIN is removed.
- **Composition root + module wiring.** `documents.Dependencies.TemplateVersionPort` added;
  `documents.New` conditionally attaches `WithTemplateSchemaReader(NewTemplateVersionSchemaReader(
  db, deps.TemplateVersionPort))` when non-nil (preserves the prior absent-reader behavior).
  `apps/api/cmd/metaldocs-api/main.go` sets
  `docDeps.TemplateVersionPort = templatesinfra.NewTemplateVersionReader(deps.SQLDB)`.
- **Ledger drained.** N1 removed from `hgPendingRemediation`.

> Producer matches the consumer contract in `spec.md`: documents needs the raw `placeholder_schema`
> bytes for a known version id; the port returns exactly that (`[]byte`, ErrNoRows→nil).
>
> No import cycle: the consumer→owner edge `documents/application → templates/domain` already
> existed (the `Placeholder` type). Adapter injected from `main`. Confirmed by `go build ./...`.

> **Bounded behavior note (documented in spec.md):** the port scopes tenancy via
> `templates_template.tenant_id`; the old JOIN scoped via `documents.tenant_id`. For valid data these
> are identical (a document's version belongs to the document's tenant). The split path additionally
> tightens a cross-tenant-corrupt pointer (a document referencing another tenant's version now
> resolves to nil rather than leaking that schema) — a strict improvement, not a regression. The
> document-resolve step preserves the original `documents.tenant_id` scope.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — parity test first, green before raw-JOIN deletion (D6) | `go test -tags integration -run TestLoadFillInSchema_ParityWithRawJoin ./internal/modules/documents/application/...` | `ok` 3.535s — `present_schema`, `absent_document`, `null_schema` all PASS | real (PG :5434) |
| Static — build | `go build ./...` | `BUILD OK` (exit 0) | — |
| Static — guard | `go run ./tools/cilint ./...` | `CILINTEXIT=0` with N1 removed from `hgPendingRemediation` | real |
| Targeted tests — templates (all) + documents/application (integration) | `go test -tags integration ./internal/modules/templates/... ./internal/modules/documents/application/...` | all `ok` | real (PG) |
| Runtime proof — port == raw | parity test runs the verbatim pre-port JOIN baseline (`tv JOIN documents d … WHERE d.id=$1 AND d.tenant_id=$2`) and the split path via the real `TemplateVersionReader`, asserting `reflect.DeepEqual` across present/absent/null | identical results all cases | real (PG) |
| 0 raw `templates_template_version` reads in `documents/` (non-test) | `grep -rn templates_template_version internal/modules/documents` (excl. `*_test.go`) | only two comment references remain (`module.go:68`, `fillin_service.go:226`); no SQL | real |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Split path == raw JOIN for present schema (non-empty `[]Placeholder`) | yes | `present_schema` PASS (DeepEqual + 1 field `field_a`) |
| Absent document → nil (no error) | yes | `absent_document` PASS (nil == nil) |
| Version with null schema → nil | yes | `null_schema` PASS (DeepEqual) |
| Whole tree builds; templates + documents tests pass | yes | `go build ./...` clean; suites `ok` |
| Guard clears N1 | yes | cilint `EXIT=0`, ledger entry removed |
| 0 raw `templates_template_version` reads outside templates | yes | grep clean (non-test) |

## Review disposition

- **Spec-compliance review:** PASS. Extended the existing ADR-0030 port (no parallel reader);
  documents resolves the version id on its own table then delegates the foreign read. Exactly the
  contract. The bounded tenant-scope divergence is documented and is a strict tightening.
- **Code-quality review:** PASS. Adapter SQL reads the owner's own tables; the parity test locks the
  split path to the historical JOIN output. nil-port guard in `documents.New` preserves the prior
  absent-reader behavior (fill-in schema reader simply not attached). No fakes broken
  (controlleddocuments uses its local `application.TemplateVersionChecker`, not this port).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `TestSequenceAllocatorNextAndIncrement_Concurrent` (controlleddocuments/domain) env FAIL | Pre-existing / raw-base-DSN schema gap, unrelated to F2.4 (documented under F2.2 defers). HS-3 class — not false-greened. | Tracked in F2.2 evidence; env owner |
| Whole-tree `go test ./...` not run green end-to-end | Known pre-existing breaks elsewhere (documented in F2.1), unrelated to this seam | Tracked in F2.1 defers |
| F2.3 leftover: `documents/application/create_document_snapshot_integration_test.go` needed the 4th `docrepo.New` arg (`taxonomydomain.NoopAreaCatalogReader{}`) | The F2.3 bulk-sed touched `documents/repository/*_test.go` only; this integration-tagged file in `documents/application` was missed. Caught by `go vet -tags integration` here, fixed in this commit. | Closed in F2.4 |
