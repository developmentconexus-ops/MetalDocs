# ADR 0014 — Rename docgen-v2 service to docx-renderer

**Date:** 2026-06-01
**Status:** Accepted

## Context

`apps/docgen-v2/` was originally a multi-route Node service:

- `POST /render/docx` — render a single document
- `POST /validate/template` — validate template tokens
- `POST /convert/pdf` — proxy DOCX to Gotenberg for PDF
- `POST /render/fanout` — reconstruct frozen DOCX during approval freeze

After the PDF path was moved to a Go direct-to-Gotenberg adapter (`GotenbergPDFClient` in `internal/platform/servicebus/gotenberg_pdf.go`), and the render/validate routes were removed as dead code, **only `/render/fanout` remains**. The service now does a single job: accept a body DOCX + token values, run eigenpal headless substitution, and upload the frozen DOCX artifact.

The name "docgen-v2" no longer describes what the service does. It:
- Implies PDF generation ("docgen")
- Implies there is or was a v1
- Does not communicate that its sole remaining responsibility is DOCX token rendering

The service is also an anti-corruption layer: Go calls it over HTTP, and the underlying JS rendering engine (eigenpal) can be replaced (e.g. with SuperDoc) without touching backend logic.

## Decision

Rename the service from `docgen-v2` to `docx-renderer` everywhere:

- Directory: `apps/docgen-v2/` → `apps/docx-renderer/`
- Package: `@metaldocs/docgen-v2` → `@metaldocs/docx-renderer`
- Docker service: `docgen-v2` → `docx-renderer`
- Env var prefix: `DOCGEN_V2_*` → `DOCX_RENDERER_*`
- Go-side auth token env var: `METALDOCS_DOCGEN_V2_SERVICE_TOKEN` → `METALDOCS_DOCX_RENDERER_SERVICE_TOKEN`
- Dev script: `scripts/dev-docgen.ps1` → `scripts/dev-docx-renderer.ps1`

**Not renamed:**
- HTTP route path `/render/fanout` — names the operation, not the service
- Go package `fanout` — names the operation
- `METALDOCS_FANOUT_URL` — names the route, not the service
- DB column `docgen_v2_ver` — opaque storage; changing it would invalidate composite hashes for existing exports
- Version string `"docgen-v2@0.4.0"` — stored verbatim in `document_exports.docgen_v2_ver`; renaming it changes composite hash output and breaks cache coherence with historical exports
- Event type `docgen_v2_pdf` — DB/outbox identifier; DB schema rename is out of scope
- Go package `docgenv2` in `internal/platform/docgenv2/` — internal identifier, does not appear in env/config surface

## Consequences

- Service name now reflects single responsibility (DOCX rendering only).
- Env config surface is consistent: everything the service owns uses `DOCX_RENDERER_*`.
- If eigenpal is replaced by SuperDoc or another provider, the service name remains accurate — "docx-renderer" names the contract (render DOCX), not the implementation.
- Operators must update `.env` files: `DOCGEN_V2_SERVICE_TOKEN` → `DOCX_RENDERER_SERVICE_TOKEN`, `METALDOCS_DOCGEN_V2_SERVICE_TOKEN` → `METALDOCS_DOCX_RENDERER_SERVICE_TOKEN`.
- Compose service name changes; any external tooling that references `metaldocs-docgen-v2` container by name must update.
