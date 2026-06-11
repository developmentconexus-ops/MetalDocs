# docx-renderer — TypeScript DOCX Substitution Service

> **Last verified:** 2026-06-11
> **Scope:** `apps/docx-renderer/` — the TypeScript/Node.js sidecar binary that owns DOCX token substitution and composition-block injection. This page covers the service as a deployed binary: entrypoint, HTTP API, security, eigenpal integration, MinIO I/O, build, and Dockerfile. How the Go worker calls it and where it fits in the full pipeline is in [../flows/render-pipeline.md](../flows/render-pipeline.md).
> **Key files:**
> - `apps/docx-renderer/src/index.ts`
> - `apps/docx-renderer/src/routes/fanout.ts`
> - `apps/docx-renderer/src/render/fanout.ts`
> - `apps/docx-renderer/src/env.ts`
> - `apps/docx-renderer/src/service-auth.ts`
> - `apps/docx-renderer/Dockerfile`
> - `apps/docx-renderer/package.json`

---

## 1. Purpose and position

`docx-renderer` is the internal service that performs the actual DOCX artifact production. It is **never exposed at the edge** — the target architecture (REQ in [../../architecture/backend-target-architecture.md](../../architecture/backend-target-architecture.md)) specifies it is reached only via authenticated internal calls from the Go worker.

Its single responsibility: given a body DOCX template in MinIO and a set of resolved placeholder values + composition-block configs, apply token substitution and inject OOXML composition blocks, then write the frozen DOCX back to MinIO and return the content hash.

All business logic for *which* placeholders to resolve and *what* their values are lives in the Go side (`internal/modules/render/resolvers/`). The TypeScript side is a pure rendering engine: it does not read Postgres, does not call external APIs, and makes no business decisions.

---

## 2. Entrypoint and server

**`src/index.ts`** — builds the Fastify app, registers all cross-cutting hooks and routes, then calls `listen()` on `DOCX_RENDERER_PORT` (default `3100`).

Startup sequence:
1. Parse and validate env via `loadEnv()` (`src/env.ts`) — Zod schema; throws on invalid config with secrets redacted in error messages.
2. Create a lazy `s3Factory` closure (`src/index.ts:17-18`): `makeS3Client(env)` is **not** called at startup — the client is constructed on the first request via `cachedClient ??= makeS3Client(env)` and then reused.
3. Build Fastify instance with configured logger at `LOG_LEVEL`.
4. Register `onRequest` service-auth hook (`src/service-auth.ts`) — validates `X-Service-Token` via `node:crypto.timingSafeEqual`; exempts `/health`.
5. Register health route (`GET /health`).
6. Register fanout route via `registerRoutes(app, env, s3Factory)` (`src/routes/index.ts:6`, called from `src/index.ts:20`).
7. `app.listen({ port, host: '0.0.0.0' })`.

---

## 3. HTTP API

### Routes

| Method | Path | Auth | Purpose |
|---|---|---|---|
| `GET` | `/health` | None | Liveness probe; returns `{ status: "ok", version }` |
| `POST` | `/render/fanout` | `X-Service-Token` | DOCX rendering: read template DOCX from MinIO → substitute tokens + inject composition blocks → write frozen DOCX to MinIO → return hash |

### `POST /render/fanout` — request/response

**Request body** (Zod-validated in `src/routes/fanout.ts`):

```
{
  tenant_id: string (UUID),
  revision_id: string (UUID),
  body_docx_s3_key: string,          // MinIO key of the template DOCX
  placeholder_values: Record<string, string>,
  composition_config: {              // sub-block layout config (header, footer slots)
    header_sub_blocks: string[],
    footer_sub_blocks: string[],
    sub_block_params: Record<string, Record<string, unknown>>
  },
  resolved_values: Record<string, unknown>   // full resolved document values (fed to sub-block renderers)
}
```

**Handler flow** (`src/routes/fanout.ts`):
1. Validate body with Zod schema.
2. Fetch template DOCX bytes from MinIO via `getObjectBuffer(s3, bucket, bodyDocxS3Key)` (`s3.ts`).
3. Call `fanout(docxBuffer, body)` (`src/render/fanout.ts`).
4. Upload frozen DOCX to `tenants/{tenantId}/revisions/{revisionId}/frozen.docx` via `putObjectBuffer` (`routes/fanout.ts:65-71`).
5. Return `{ content_hash, final_docx_s3_key, unreplaced_vars, size_bytes }`.

**Response body**:

```
{
  content_hash: string,          // hex-encoded SHA-256 of frozen DOCX bytes
  final_docx_s3_key: string,     // MinIO path where frozen DOCX was written
  unreplaced_vars: string[],     // placeholder tokens not substituted (diagnostic)
  size_bytes: number
}
```

---

## 4. Rendering engine — `src/render/fanout.ts`

The `fanout()` function is the rendering core. It:

1. Builds a `SubBlockRegistry` with all 5 built-in composition-block renderers (`registerV1Builtins`).
2. Renders header and footer sub-blocks concurrently via `Promise.all` (`render/fanout.ts:27-46`), using `compositionConfig` to select which sub-block key goes in each slot.
3. Merges rendered OOXML sub-block strings with `placeholderValues` into a single `variables` map.
4. Calls `processTemplateDetailed(docxBuffer, variables)` from `@eigenpal/docx-js-editor` (vendored) — synchronous DOCX template substitution; blocks the Node.js event loop for the duration.
5. Computes SHA-256 of the resulting buffer via `node:crypto.createHash('sha256')`.
6. Returns `{ buffer, contentHash, unreplacedVars }`.

### Sub-block registry and built-in renderers

`SubBlockRegistry` (`src/render/subblocks/registry.ts`) is a keyed map of renderer functions. `register(key, fn)` adds a renderer; `render(key, params)` throws on unknown keys; `keys()` returns the registered set.

**Built-in sub-blocks registered by `registerV1Builtins`** (`src/render/subblocks/builtins.ts`):

| Key | File | Output |
|---|---|---|
| `doc_header_standard` | `doc_header_standard.ts` | OOXML table: title, doc code, effective date, revision number |
| `revision_box` | `revision_box.ts` | OOXML table from `revision_history` array in resolved values |
| `approval_signatures_block` | `approval_signatures_block.ts` | OOXML table: approver name + `signed_at` |
| `footer_page_numbers` | `footer_page_numbers.ts` | OOXML `PAGE`/`NUMPAGES` field elements |
| `footer_controlled_copy_notice` | `footer_controlled_copy_notice.ts` | "CONTROLLED COPY — WHEN PRINTED" notice; `params.notice_text` overrides the default |

---

## 5. Service-to-service authentication

`src/service-auth.ts` registers a Fastify `onRequest` hook that runs on every route except `/health`.

- Reads the `X-Service-Token` request header.
- Compares it to `DOCX_RENDERER_SERVICE_TOKEN` (env) using `node:crypto.timingSafeEqual` — constant-time comparison prevents timing attacks.
- Returns HTTP 401 if the header is absent or the comparison fails.

The matching Go side: `fanout.Client` (`internal/modules/render/fanout/client.go:51`) sends the token from `METALDOCS_DOCX_RENDERER_SERVICE_TOKEN` as the `X-Service-Token` header.

---

## 6. Object storage (MinIO)

`src/s3.ts` builds a `minio.Client` from env. Two helpers:

- `getObjectBuffer(client, bucket, key)` — streams the object into a `Buffer`.
- `putObjectBuffer(client, bucket, key, buffer, contentType)` — uploads a `Buffer` as an object.

All MinIO I/O uses the bucket configured in `DOCX_RENDERER_S3_BUCKET` (default `metaldocs-docx-v2`).

**Read:** `bodyDocxS3Key` from request body — the template DOCX before rendering.
**Write:** `tenants/{tenantId}/revisions/{revisionId}/frozen.docx` — the post-render frozen DOCX. Because the key is deterministic, a retry of the same fanout call is idempotent at the MinIO level.

---

## 7. Environment configuration

All env vars are validated at startup via Zod in `src/env.ts`. Fields marked `min(N)` throw on startup if under-length.

| Variable | Default | Constraint | Notes |
|---|---|---|---|
| `DOCX_RENDERER_PORT` | `3100` | — | HTTP listen port |
| `DOCX_RENDERER_SERVICE_TOKEN` | required | `>=16 chars` | Pre-shared secret for service-auth hook |
| `DOCX_RENDERER_S3_ENDPOINT` | `http://minio:9000` | — | MinIO endpoint |
| `DOCX_RENDERER_S3_ACCESS_KEY` | required | `>=3 chars` | MinIO access key |
| `DOCX_RENDERER_S3_SECRET_KEY` | required | `>=3 chars` | MinIO secret key |
| `DOCX_RENDERER_S3_BUCKET` | `metaldocs-docx-v2` | — | Bucket for all DOCX read/write |
| `DOCX_RENDERER_S3_USE_SSL` | `false` | boolean | SSL flag for MinIO client |
| `DOCX_RENDERER_GOTENBERG_URL` | `http://gotenberg:3000` | — | **Declared but not consumed** — no route handler reads this value; Gotenberg conversion was removed from this service; the Go worker calls Gotenberg directly. Dead configuration (see §10 legacy flags). |
| `LOG_LEVEL` | `info` | — | Fastify logger level |
| `VERSION` | `dev` | — | Returned in `/health` response |

---

## 8. Build and deployment

### Build (`build.mjs`)

esbuild single-file bundle script. Produces a bundled CommonJS artifact from `src/index.ts`.

### TypeScript config

`tsconfig.json` — standard TypeScript config for the service.

### Tests

| Suite | File |
|---|---|
| Sub-block unit tests (7) | `src/render/subblocks/__tests__/*.test.ts` |
| `fanout()` function unit tests | `src/render/__tests__/fanout.test.ts` |
| Fanout HTTP route unit tests | `src/routes/__tests__/fanout.test.ts` |
| Health smoke test | `test/health.test.ts` |
| S3 integration smoke test | `test/s3.smoke.test.ts` |
| Shared fixtures | `test/fixtures.ts` |

Test runner: Vitest (`vitest.config.ts`).

### Dockerfile

Multi-stage build:

1. **Builder stage** (`node:20.11-alpine`): installs dependencies from `package.json`, runs the esbuild bundle.
2. **Runtime stage** (`node:20.11-alpine`): copies bundled output only; no dev dependencies.
3. Exposes port `3100`; healthcheck on `GET /health`; process runs as `node` user (non-root).

### eigenpal dependency

`@eigenpal/docx-js-editor@0.2.0` is vendored as `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` and installed locally via `package.json`. It is not pulled from a registry. This is the only external DOCX processing dependency.

Direct runtime dependencies: `fastify@4.26.2`, `minio@^7.1.3`, `zod@3.23.8`, `jszip@3.10.1` (direct dependency, declared in `package.json:19`), `ajv@^8.16.0`, `@metaldocs/shared-tokens`, `node:crypto` (stdlib). `@eigenpal/docx-js-editor@0.2.0` is a separate direct dependency (vendored tarball).

---

## 9. Concurrency model

Fastify runs on Node.js's single-threaded async I/O event loop. The `fanout()` function is `async` and uses `Promise.all` for concurrent sub-block rendering (`render/fanout.ts:27-46`). Sub-block renderers implement `render(ctx): Promise<string>` (typed async in `SubBlockRenderer`; all five built-ins declare `async render`), and the two `Promise.all` fan-outs await them before merging results.

`processTemplateDetailed` (eigenpal) is synchronous and blocks the event loop for the duration of DOCX processing. Under concurrent load, requests queue behind each other at the eigenpal call.

---

## 10. Legacy and open flags

| Flag | Location | Description |
|---|---|---|
| `DOCX_RENDERER_GOTENBERG_URL` declared but not consumed | `src/env.ts:13` | Env var declared in Zod schema; no route handler reads it; Gotenberg was removed from the TS side; dead configuration risks operator confusion |
| `processTemplateDetailed` blocks the event loop | `src/render/fanout.ts` | Synchronous call from eigenpal; sub-block rendering is async (awaited via `Promise.all`) but the eigenpal substitution step itself is synchronous; throughput limited by single-threaded execution at that call |
| Vendored eigenpal tarball at `0.2.0` | `vendor/eigenpal/` | No registry; upgrade requires manual tarball replacement; version pinned in perpetuity until manually updated |

See also [../_artifacts/stage1/synthesis-legacy.md](../_artifacts/stage1/synthesis-legacy.md) for the full cross-cutting legacy register.

---

## Sources

Stage-1 artifact: `wiki/backend/_artifacts/stage1/render-pipeline.md` (§2 file inventory for `apps/docx-renderer/`, §3 public surface, §4 flow descriptions, §5 dependencies, §7 config, §8 concurrency, §10 legacy flags).
Strategic framing: [../../architecture/backend-blueprint.md](../../architecture/backend-blueprint.md) concern D6 (internal service-to-service) and C6 (async platform).
