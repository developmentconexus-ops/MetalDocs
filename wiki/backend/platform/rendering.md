# Platform Rendering — render & docgenv2 Packages

> **Last verified:** 2026-07-03 (DB-01 CLOSED: legacy `TemplateReader`, `FanoutTemplateReader`, and the `LegacyTemplateReadCount()` counter are deleted; `TemplatesTemplateReader` is the only template reader, wired directly as `docDeps.TplRead`. Run-window proof collected before deletion: zero legacy fallback reads across the full Goal-3 QA window, legacy tables empty on canonical bootstrap. Legacy `public.templates`/`public.template_versions` tables dropped by migration 0268.)
> **Prior:** 2026-07-01 (ARC-01 canonical-first flip), 2026-06-11
> **Scope:** `internal/platform/render/gotenberg/`, `internal/platform/docgenv2/` — the two platform packages that bridge domain-module rendering needs to external infrastructure (Gotenberg, MinIO, and the two template schemas). This page covers only the platform layer; the end-to-end flow lives in [../flows/render-pipeline.md](../flows/render-pipeline.md) and the TypeScript sidecar is documented in [../binaries/docx-renderer.md](../binaries/docx-renderer.md).
> **Key files:**
> - `internal/platform/render/gotenberg/client.go`
> - `internal/platform/docgenv2/templates_reader.go`
> - `internal/platform/docgenv2/templates_snapshot_reader.go`

---

## 1. Purpose

These two packages are platform-layer adapters (infrastructure side, in hexagonal terms): they implement interfaces defined by the `documents` and `templates` modules but own no business logic themselves. They exist so that the Gotenberg HTTP API and the two competing template database schemas remain isolated outside the domain modules.

---

## 2. `internal/platform/render/gotenberg/`

### Role

Wraps the [Gotenberg](https://gotenberg.dev/) document-conversion service. Gotenberg is an HTTP microservice that exposes LibreOffice and Chromium conversion routes. The `Client` in this package is the only place in the Go codebase that forms multipart requests to Gotenberg.

### File inventory

| File | Role |
|---|---|
| `client.go` | `Client` struct — `ConvertHTMLToPDF` (Chromium route), `ConvertDocxToPDF`, `ConvertDocxToPDFWithOptions` (LibreOffice route); 30 s timeout; 64 MiB response body cap; A4/Letter paper-size override support |
| `client_test.go` | Unit tests |

### Public surface

```
type Client struct { ... }
func NewClient(baseURL string) (*Client, error)
func (c *Client) ConvertHTMLToPDF(ctx context.Context, htmlBytes []byte, cssBytes []byte) ([]byte, error)
func (c *Client) ConvertDocxToPDF(ctx context.Context, docxContent []byte) ([]byte, error)
func (c *Client) ConvertDocxToPDFWithOptions(ctx context.Context, docxContent []byte, paperSize string, landscape bool) ([]byte, error)
```

- Constructed in `internal/platform/bootstrap/worker.go:83` and `internal/platform/bootstrap/api.go:68`.
- Consumed by `platform/servicebus.GotenbergPDFClient.ConvertPDF` (`servicebus/gotenberg_pdf.go:70`) which wraps it with MinIO open/write and SHA-256 hashing.
- `ConvertDocxToPDF` posts to Gotenberg's `/forms/libreoffice/convert` route as a multipart form.
- Error responses are capped at 4 KiB; non-200 responses are returned as a typed error containing status and body.

### Configuration

| Go env var | Parsed at | Default | Effect |
|---|---|---|---|
| `METALDOCS_GOTENBERG_URL` | `config/gotenberg.go:18` | `""` | Empty → Gotenberg disabled, PDF conversion path inactive. Non-empty → `Client` constructed and wired. |
| `METALDOCS_ATTACHMENTS_PROVIDER` | `config.LoadAttachmentsConfig()` | — | Must be `minio` for the Gotenberg path to activate (`bootstrap/worker.go:76`). |

---

## 3. `internal/platform/docgenv2/`

### Role

Provides template and snapshot readers for use by the `documents` application layer. Historically the package bridged two template schemas (legacy `template_versions`/`templates` vs canonical `templates_template_version`/`templates_template`); the migration completed 2026-07-03 (DB-01, migration 0268) — the legacy tables are dropped and only the canonical readers remain.

### Naming note — docgen v1 vs v2

There is no "docgen v1" remnant in the codebase. The package name `docgenv2` is a historical artifact from the naming era when the service was called `docgen v2`. The service it references is now called `docx-renderer`. The event type constant `EventTypePDFConvert = "docgen_v2_pdf"` carries the v2 name string on the wire (`internal/platform/messaging`). The package name itself is flagged as misleading in the legacy register (see §6 below).

### File inventory

| File | Role |
|---|---|
| `templates_reader.go` | `TemplatesTemplateReader` — reads published DOCX key from `templates_template_version`/`templates_template`; schema always returns `""`; carries `systemTemplateTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"` for system-owned templates (relocated here when the legacy reader was deleted, DB-01) |
| `templates_snapshot_reader.go` | `TemplatesSnapshotReader` — implements `documents/application.SnapshotTemplateReader`; loads `placeholder_schema` JSON and DOCX key from `templates_template_version`; `CompositionJSON` hardcoded to `{}` [runtime-unverified: whether this is intentional or a missing feature] |
| `templates_reader_test.go` | Unit tests for `TemplatesTemplateReader` (system-template tenant allowance) |
| `templates_snapshot_reader_test.go` | Unit tests for `TemplatesSnapshotReader` |

### Public surface

```
type TemplatesTemplateReader struct { ... }
func NewTemplatesTemplateReader(db *sql.DB) *TemplatesTemplateReader

type TemplatesSnapshotReader struct { ... }
func NewTemplatesSnapshotReader(db *sql.DB) *TemplatesSnapshotReader
```

Wired in `apps/api/cmd/metaldocs-api/main.go`:
- `TemplatesTemplateReader` constructed directly as `docDeps.TplRead` (~`main.go:448-455`).
- `TemplatesSnapshotReader` assigned as `docSnapshotReader` a few lines above the `docDeps` construction.

### Template reader flow (DB-01 closed 2026-07-03: canonical only)

`TemplatesTemplateReader.GetPublishedVersion()` reads `templates_template_version`/`templates_template` directly; `sql.ErrNoRows` surfaces as not-found, any other error surfaces unchanged. The former `FanoutTemplateReader` (canonical-first with legacy `sql.ErrNoRows` fallback, ARC-01) and its `LegacyTemplateReadCount()` counter + WARN log were deleted together with the legacy tables (migration 0268) after the run-window proof: zero legacy fallback reads across the full Goal-3 QA window, legacy tables empty on canonical bootstrap.

- Schema JSON: the canonical reader always returns `""` (the legacy MinIO-backed schema fetch died with the legacy reader).
- `TemplatesSnapshotReader` maps `sql.ErrNoRows` to `domain.ErrSnapshotTemplateNotFound` (`templates_snapshot_reader.go:40`).

### Persistence — tables accessed

| Table | Schema | Access type | Reader |
|---|---|---|---|
| `templates_template_version` | canonical | read | `TemplatesTemplateReader`, `TemplatesSnapshotReader` |
| `templates_template` | canonical | read | `TemplatesTemplateReader`, `TemplatesSnapshotReader` |

(Legacy `template_versions`/`templates` dropped by migration 0268; MinIO schema fetch removed with the legacy reader.)

---

## 4. Imports

### Outbound

**`internal/platform/render/gotenberg`**
- `net/http`, `mime/multipart` — multipart HTTP to Gotenberg
- No internal MetalDocs imports

**`internal/platform/docgenv2`**
- `database/sql` — template and snapshot queries
- `metaldocs/internal/modules/documents/application` — `SnapshotTemplateReader` interface
- `metaldocs/internal/modules/documents/domain` — `TemplateSnapshot`, `ErrSnapshotTemplateNotFound`

(minio-go import removed with the legacy reader, DB-01.)

### Inbound (who imports these packages)

| Importer | Package imported |
|---|---|
| `internal/platform/bootstrap/worker.go` | `platform/render/gotenberg` |
| `internal/platform/bootstrap/api.go` | `platform/render/gotenberg` |
| `internal/platform/worker/pdf_pipeline_test.go` | `platform/render/gotenberg` (test) |
| `apps/api/cmd/metaldocs-api/main.go` | `platform/docgenv2` (via bootstrap) |

---

## 5. Relation to blueprint and target architecture

These packages are domain-free platform adapters (REQ-TOP-2 in [../../architecture/backend-target-architecture.md](../../architecture/backend-target-architecture.md)) and correctly carry no business logic. The import of `documents/application` and `documents/domain` interfaces in `docgenv2` is the correct direction (platform implements a domain-defined port). The reverse (a platform package importing a module and knowing its internal state) would be a layering violation.

**Resolved (2026-07-03, DB-01):** the whole fanout/fallback design-debt cluster is gone — `FanoutTemplateReader`, the legacy `TemplateReader`, and their test harnesses were deleted once the run-window proof showed zero legacy reads. The surviving `TemplatesTemplateReader` is a single concrete type with sqlmock unit coverage (`templates_reader_test.go`).

---

## 6. Legacy and open flags

| Flag | Location | Description | RF ref |
|---|---|---|---|
| Package name `docgenv2` is misleading | `internal/platform/docgenv2/` | Name reflects the "docgen v2" era; the service it references is now `docx-renderer` | — |
| Magic sentinel UUID for system templates | `templates_reader.go` | `systemTemplateTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"` baked as a package constant; no named domain concept | — |
| Schema always `""` in new template path | `templates_snapshot_reader.go:48` | `CompositionJSON` hardcoded to `{}` [runtime-unverified: whether composition config is intentionally absent for the templates module] | — |
| ~~Dual-schema fallback is temporary~~ | — | CLOSED 2026-07-03 (DB-01): run-window proof collected (zero legacy fallback reads); `FanoutTemplateReader` + legacy `TemplateReader` deleted, legacy `template_versions`/`templates` tables dropped (migration 0268) | DB-01 |

See also [../_artifacts/stage1/synthesis-legacy.md](../_artifacts/stage1/synthesis-legacy.md) for the full cross-cutting legacy register.

---

## Sources

Stage-1 artifact: `wiki/backend/_artifacts/stage1/render-pipeline.md` (§2 file inventories, §4 logic flows, §5 dependencies, §6 persistence, §7 config, §10 legacy flags).
Strategic framing: [../../architecture/backend-blueprint.md](../../architecture/backend-blueprint.md) concern C6 (async) and D6 (internal HTTP clients).
