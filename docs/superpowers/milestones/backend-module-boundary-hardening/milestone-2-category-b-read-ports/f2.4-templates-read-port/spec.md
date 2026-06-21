# Feature F2.4 — Spec — templates placeholder-schema read-port

> **Milestone:** M2 (category-b-read-ports) · **Feature:** `f2.4-templates-read-port`
> **Census site:** **N1** (`documents/application/fillin_service.go:235` — `SELECT tv.placeholder_schema
> FROM templates_template_version tv JOIN documents d ON d.template_version_id = tv.id WHERE d.id=$1
> AND d.tenant_id=$2`).
> **Owner:** templates owns `templates_template_version`. ADR-0039 D1 + D3(b); **extends** the existing
> ADR-0030 `TemplateVersionPort` — no parallel reader.

## Problem

`documents` `TemplateVersionSchemaReader.LoadFillInSchema` reads templates' owned
`templates_template_version.placeholder_schema` by **joining it to `documents`** on
`documents.template_version_id`. The JOIN reaches across the module boundary into a templates base
table (H-G violation, census N1). The read is off-tx and ErrNoRows-tolerant (missing → nil schema).

`documents` already owns `documents.template_version_id`, so the doc→version resolution can be done
on its own table; only the `placeholder_schema` lookup belongs to templates.

## Consumer contract (defined before the producer)

The consumer needs, for a `(tenantID, docID)`: the placeholder schema of the template version that
document points to, as the raw JSON it already `json.Unmarshal`s into `[]templatesdomain.Placeholder`,
with not-found → nil.

Rather than a new reader, **extend the canonical ADR-0030 `templates/domain.TemplateVersionPort`**
(already consumed by documents/controlleddocuments for version state) with one keyed-by-version
method:

```go
// added to TemplateVersionPort
// PlaceholderSchema returns templates_template_version.placeholder_schema (raw
// JSON) for versionID, scoped to tenantID via the owning template. nil when the
// version is absent or the column is NULL (caller treats nil as "no schema").
PlaceholderSchema(ctx context.Context, tenantID, versionID string) ([]byte, error)
```

- **Producer:** `templates/infrastructure.TemplateVersionReader` implements `PlaceholderSchema` using
  the same `templates_template_version v JOIN templates_template t … WHERE v.id=$1 AND t.tenant_id=$2`
  tenant-scoping shape as its sibling `templateVersionQuery`. ErrNoRows → `(nil, nil)`.
- **Consumer:** `documents` `TemplateVersionSchemaReader` gains the port; `LoadFillInSchema` resolves
  `versionID` from **its own** `documents` table (`SELECT template_version_id FROM documents WHERE
  id=$1 AND tenant_id=$2`; ErrNoRows → nil), then calls `tplVersions.PlaceholderSchema(ctx, tenantID,
  versionID)` and unmarshals exactly as today. The cross-module `templates_template_version` JOIN is
  deleted.

## Non-goals

- **No new reader / no parallel port.** Extend `TemplateVersionPort` (ADR-0030) only.
- **No behavior change for valid data.** Same rows, same nil-on-missing, same unmarshal. Seam only.
- **No tx change.** Stays off-tx (both the documents own-table read and the port read run on the pool).

## Bounded divergence (deliberate, documented)

The port tenant-scopes via `templates_template.tenant_id` (consistent with the sibling
`templateVersionQuery`), whereas the old JOIN scoped via `documents.tenant_id`. For valid data the
version a tenant's document points to belongs to that tenant's template, so results are identical.
The only difference is a cross-tenant-corrupt pointer (a tenant's document referencing another
tenant's template version) would now return nil instead of leaking the foreign schema — a
correctness tightening, not a regression. Mirrors the M4/F4.1 display-name tenant-scoping precedent.

## Interview record (consumer-contract discovery)

| Q | A |
|---|---|
| New reader or extend the port? | **Extend** `TemplateVersionPort` (ADR-0030) — it is the sanctioned cross-module templates read surface; a parallel reader would re-introduce the seam the program is closing. |
| Key by docID or versionID? | versionID. documents owns `documents.template_version_id`; it resolves the version on its own table, then asks templates only for the schema. Templates never names `documents`. |
| Return type? | Raw `[]byte` (the `placeholder_schema` JSON), so the consumer's existing `json.Unmarshal` into `[]templatesdomain.Placeholder` is byte-for-byte unchanged. nil = absent/NULL. |
| Tenant scoping? | Via `templates_template.tenant_id`, matching the sibling query (bounded tightening — see above). |

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Port path returns the same `[]Placeholder` as the raw `tv JOIN documents` query across present-schema / absent-document / version-with-null-schema | `TestLoadFillInSchema_ParityWithRawJoin` (integration, :5434) | real (PG) |
| Whole tree builds; targeted tests pass | `go build ./...`; documents + templates suites | real |
| Guard clears N1 | `go run ./tools/cilint ./...` exit 0 with `{documents/application/fillin_service.go, templates_template_version}` removed from `hgPendingRemediation` | real |
| 0 raw `templates_template_version` reads remain in `documents/` (non-test) | `grep` shows none | real |

> TDD: write the failing parity test first, port to green, then delete the raw JOIN (parity green
> **before** deletion — D6). PG unavailable ⇒ mark integration steps **not-run (HS-3)**, never
> false-green.

## ADR needed?

- [x] **No new ADR.** Extends the existing **ADR-0030** `TemplateVersionPort` under **ADR-0039** D3(b).
  Method shape recorded here + in `evidence.md`. No novel cross-module contract.
