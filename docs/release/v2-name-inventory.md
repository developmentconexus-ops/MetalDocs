# Release V2 Name Inventory

- Generated: 2026-05-16 11:14:28 -03:00
- Scope: api, apps, db, docs, frontend/apps/web/src, internal, migrations, wiki
- Match regex: (?i)\bV2\b|_v2\b|/v2/|api/v2|templatesV2|templates_v2|documents_v2|docgenv2|docgen-v2|docgen_v2
- Total hits: 5082
- Allowed hits: 49
- Unexpected hits: 5033
- Allowlist rules: github.com/oapi-codegen/oapi-codegen/v2, oapi-codegen v2, openapi-typescript v7, MDDM/v2

## Unexpected Hits

| Status | File | Line | Snippet |
|---|---|---:|---|
| unexpected | api/openapi/v1/openapi.yaml | 2913 | summary: Publish draft (delegates to docgen-v2 /validate/template) |
| unexpected | api/openapi/v1/partials/templates.yaml | 140 | summary: Publish draft (delegates to docgen-v2 /validate/template) |
| unexpected | apps/api/cmd/metaldocs-api/main.go | 63 | docgenv2 "metaldocs/internal/platform/docgenv2" |
| unexpected | apps/api/cmd/metaldocs-api/main.go | 234 | serviceToken := strings.TrimSpace(os.Getenv("METALDOCS_DOCGEN_V2_SERVICE_TOKEN")) |
| unexpected | apps/api/cmd/metaldocs-api/main.go | 236 | slog.Warn("METALDOCS_DOCGEN_V2_SERVICE_TOKEN not set; fanout requests will be rejected with 401") |
| unexpected | apps/api/cmd/metaldocs-api/main.go | 263 | docSnapshotReader := docgenv2.NewTemplatesSnapshotReader(deps.SQLDB) |
| unexpected | apps/api/cmd/metaldocs-api/main.go | 269 | TplRead: docgenv2.NewFanoutTemplateReader( |
| unexpected | apps/api/cmd/metaldocs-api/main.go | 270 | docgenv2.NewTemplateReader(deps.SQLDB, deps.MinioClient, deps.MinioBucket), |
| unexpected | apps/api/cmd/metaldocs-api/main.go | 271 | docgenv2.NewTemplatesTemplateReader(deps.SQLDB), |
| unexpected | apps/api/cmd/metaldocs-api/main.go | 283 | if deps.DocgenV2Client != nil { |
| unexpected | apps/api/cmd/metaldocs-api/main.go | 284 | docDeps.ExportDocgen = deps.DocgenV2Client |
| unexpected | apps/docgen-v2/dist/env.js | 3 | DOCGEN_V2_PORT: z.coerce.number().int().min(0).max(65535).default(3100), |
| unexpected | apps/docgen-v2/dist/env.js | 4 | DOCGEN_V2_SERVICE_TOKEN: z.string().min(16, 'service token must be >= 16 chars'), |
| unexpected | apps/docgen-v2/dist/env.js | 7 | DOCGEN_V2_S3_ENDPOINT: z.string().default('http://minio:9000'), |
| unexpected | apps/docgen-v2/dist/env.js | 8 | DOCGEN_V2_S3_ACCESS_KEY: z.string().min(3), |
| unexpected | apps/docgen-v2/dist/env.js | 9 | DOCGEN_V2_S3_SECRET_KEY: z.string().min(3), |
| unexpected | apps/docgen-v2/dist/env.js | 10 | DOCGEN_V2_S3_BUCKET: z.string().default('metaldocs-docx-v2'), |
| unexpected | apps/docgen-v2/dist/env.js | 11 | DOCGEN_V2_S3_USE_SSL: z.coerce.boolean().default(false), |
| unexpected | apps/docgen-v2/dist/env.js | 12 | DOCGEN_V2_GOTENBERG_URL: z.string().url().default('http://gotenberg:3000'), |
| unexpected | apps/docgen-v2/dist/env.js | 18 | const safe = { ...flat, DOCGEN_V2_SERVICE_TOKEN: flat.DOCGEN_V2_SERVICE_TOKEN ? ['[redacted]'] : undefined, DOCGEN_V2_S3_SECRET_KEY: flat.DOCGEN_V2_S3_SECRET_KEY ? ['[redacted]'] : undefined }; |
| unexpected | apps/docgen-v2/dist/index.js | 9 | registerServiceAuth(app, env.DOCGEN_V2_SERVICE_TOKEN); |
| unexpected | apps/docgen-v2/dist/index.js | 19 | app.listen({ port: env.DOCGEN_V2_PORT, host: '0.0.0.0' }) |
| unexpected | apps/docgen-v2/dist/pdf/version.js | 1 | export const DOCGEN_V2_VERSION = 'docgen-v2@0.4.0'; |
| unexpected | apps/docgen-v2/dist/routes/__tests__/fanout.test.js | 44 | process.env.DOCGEN_V2_SERVICE_TOKEN = TOKEN; |
| unexpected | apps/docgen-v2/dist/routes/__tests__/fanout.test.js | 45 | process.env.DOCGEN_V2_S3_ACCESS_KEY = 'key'; |
| unexpected | apps/docgen-v2/dist/routes/__tests__/fanout.test.js | 46 | process.env.DOCGEN_V2_S3_SECRET_KEY = 'sec'; |
| unexpected | apps/docgen-v2/dist/routes/convert-pdf.js | 3 | import { DOCGEN_V2_VERSION } from '../pdf/version.js'; |
| unexpected | apps/docgen-v2/dist/routes/convert-pdf.js | 20 | if (typeof token !== 'string' \\|\\| token !== env.DOCGEN_V2_SERVICE_TOKEN) { |
| unexpected | apps/docgen-v2/dist/routes/convert-pdf.js | 29 | const docxBuffer = await getObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, docx_key); |
| unexpected | apps/docgen-v2/dist/routes/convert-pdf.js | 40 | const gotenbergUrl = `${env.DOCGEN_V2_GOTENBERG_URL.replace(/\/+$/, '')}/forms/libreoffice/convert`; |
| unexpected | apps/docgen-v2/dist/routes/convert-pdf.js | 53 | await putObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, output_key, pdfBuffer, 'application/pdf'); |
| unexpected | apps/docgen-v2/dist/routes/convert-pdf.js | 58 | docgen_v2_version: DOCGEN_V2_VERSION, |
| unexpected | apps/docgen-v2/dist/routes/fanout.js | 31 | const bodyBuf = await getObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, body_docx_s3_key); |
| unexpected | apps/docgen-v2/dist/routes/fanout.js | 38 | await putObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, output_key, Buffer.from(result.buffer), DOCX_MIME); |
| unexpected | apps/docgen-v2/dist/routes/render.js | 21 | getObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, template_docx_key), |
| unexpected | apps/docgen-v2/dist/routes/render.js | 22 | getObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, schema_key), |
| unexpected | apps/docgen-v2/dist/routes/render.js | 37 | await putObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, output_key, Buffer.from(buffer), DOCX_MIME); |
| unexpected | apps/docgen-v2/dist/routes/validate-template.js | 19 | getObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, docx_key), |
| unexpected | apps/docgen-v2/dist/routes/validate-template.js | 20 | getObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, schema_key), |
| unexpected | apps/docgen-v2/dist/s3.js | 3 | const url = new URL(env.DOCGEN_V2_S3_ENDPOINT); |
| unexpected | apps/docgen-v2/dist/s3.js | 6 | port: Number(url.port \\|\\| (env.DOCGEN_V2_S3_USE_SSL ? 443 : 80)), |
| unexpected | apps/docgen-v2/dist/s3.js | 7 | useSSL: env.DOCGEN_V2_S3_USE_SSL, |
| unexpected | apps/docgen-v2/dist/s3.js | 8 | accessKey: env.DOCGEN_V2_S3_ACCESS_KEY, |
| unexpected | apps/docgen-v2/dist/s3.js | 9 | secretKey: env.DOCGEN_V2_S3_SECRET_KEY, |
| unexpected | apps/docgen-v2/Dockerfile | 4 | COPY apps/docgen-v2/package.json ./apps/docgen-v2/ |
| unexpected | apps/docgen-v2/Dockerfile | 5 | RUN npm ci --workspace @metaldocs/docgen-v2 --include-workspace-root |
| unexpected | apps/docgen-v2/Dockerfile | 6 | COPY apps/docgen-v2 ./apps/docgen-v2 |
| unexpected | apps/docgen-v2/Dockerfile | 7 | RUN npm run build --workspace @metaldocs/docgen-v2 |
| unexpected | apps/docgen-v2/Dockerfile | 13 | COPY --from=build /app/apps/docgen-v2/dist ./dist |
| unexpected | apps/docgen-v2/Dockerfile | 14 | COPY --from=build /app/apps/docgen-v2/package.json ./package.json |
| unexpected | apps/docgen-v2/package.json | 2 | "name": "@metaldocs/docgen-v2", |
| unexpected | apps/docgen-v2/src/env.ts | 4 | DOCGEN_V2_PORT: z.coerce.number().int().min(0).max(65535).default(3100), |
| unexpected | apps/docgen-v2/src/env.ts | 5 | DOCGEN_V2_SERVICE_TOKEN: z.string().min(16, 'service token must be >= 16 chars'), |
| unexpected | apps/docgen-v2/src/env.ts | 8 | DOCGEN_V2_S3_ENDPOINT: z.string().default('http://minio:9000'), |
| unexpected | apps/docgen-v2/src/env.ts | 9 | DOCGEN_V2_S3_ACCESS_KEY: z.string().min(3), |
| unexpected | apps/docgen-v2/src/env.ts | 10 | DOCGEN_V2_S3_SECRET_KEY: z.string().min(3), |
| unexpected | apps/docgen-v2/src/env.ts | 11 | DOCGEN_V2_S3_BUCKET: z.string().default('metaldocs-docx-v2'), |
| unexpected | apps/docgen-v2/src/env.ts | 12 | DOCGEN_V2_S3_USE_SSL: z.enum(['true', 'false', '1', '0']).transform(v => v === 'true' \\|\\| v === '1').default('false'), |
| unexpected | apps/docgen-v2/src/env.ts | 13 | DOCGEN_V2_GOTENBERG_URL: z.string().url().default('http://gotenberg:3000'), |
| unexpected | apps/docgen-v2/src/env.ts | 22 | const safe = { ...flat, DOCGEN_V2_SERVICE_TOKEN: flat.DOCGEN_V2_SERVICE_TOKEN ? ['[redacted]'] : undefined, DOCGEN_V2_S3_SECRET_KEY: flat.DOCGEN_V2_S3_SECRET_KEY ? ['[redacted]'] : undefined }; |
| unexpected | apps/docgen-v2/src/index.ts | 13 | registerServiceAuth(app, env.DOCGEN_V2_SERVICE_TOKEN); |
| unexpected | apps/docgen-v2/src/index.ts | 27 | app.listen({ port: env.DOCGEN_V2_PORT, host: '0.0.0.0' }) |
| unexpected | apps/docgen-v2/src/pdf/version.ts | 1 | export const DOCGEN_V2_VERSION = 'docgen-v2@0.4.0'; |
| unexpected | apps/docgen-v2/src/routes/__tests__/fanout.test.ts | 62 | process.env.DOCGEN_V2_SERVICE_TOKEN = TOKEN; |
| unexpected | apps/docgen-v2/src/routes/__tests__/fanout.test.ts | 63 | process.env.DOCGEN_V2_S3_ACCESS_KEY = 'key'; |
| unexpected | apps/docgen-v2/src/routes/__tests__/fanout.test.ts | 64 | process.env.DOCGEN_V2_S3_SECRET_KEY = 'sec'; |
| unexpected | apps/docgen-v2/src/routes/convert-pdf.ts | 6 | import { DOCGEN_V2_VERSION } from '../pdf/version.js'; |
| unexpected | apps/docgen-v2/src/routes/convert-pdf.ts | 30 | if (typeof token !== 'string' \\|\\| token !== env.DOCGEN_V2_SERVICE_TOKEN) { |
| unexpected | apps/docgen-v2/src/routes/convert-pdf.ts | 42 | const docxBuffer = await getObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, docx_key); |
| unexpected | apps/docgen-v2/src/routes/convert-pdf.ts | 59 | const gotenbergUrl = `${env.DOCGEN_V2_GOTENBERG_URL.replace(/\/+$/, '')}/forms/libreoffice/convert`; |
| unexpected | apps/docgen-v2/src/routes/convert-pdf.ts | 77 | env.DOCGEN_V2_S3_BUCKET, |
| unexpected | apps/docgen-v2/src/routes/convert-pdf.ts | 87 | docgen_v2_version: DOCGEN_V2_VERSION, |
| unexpected | apps/docgen-v2/src/routes/fanout.ts | 54 | env.DOCGEN_V2_S3_BUCKET, |
| unexpected | apps/docgen-v2/src/routes/fanout.ts | 67 | env.DOCGEN_V2_S3_BUCKET, |
| unexpected | apps/docgen-v2/src/routes/render.ts | 32 | getObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, template_docx_key), |
| unexpected | apps/docgen-v2/src/routes/render.ts | 33 | getObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, schema_key), |
| unexpected | apps/docgen-v2/src/routes/render.ts | 56 | env.DOCGEN_V2_S3_BUCKET, |
| unexpected | apps/docgen-v2/src/routes/validate-template.ts | 29 | getObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, docx_key), |
| unexpected | apps/docgen-v2/src/routes/validate-template.ts | 30 | getObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, schema_key), |
| unexpected | apps/docgen-v2/src/s3.ts | 5 | const url = new URL(env.DOCGEN_V2_S3_ENDPOINT); |
| unexpected | apps/docgen-v2/src/s3.ts | 8 | port: Number(url.port \\|\\| (env.DOCGEN_V2_S3_USE_SSL ? 443 : 80)), |
| unexpected | apps/docgen-v2/src/s3.ts | 9 | useSSL: env.DOCGEN_V2_S3_USE_SSL, |
| unexpected | apps/docgen-v2/src/s3.ts | 10 | accessKey: env.DOCGEN_V2_S3_ACCESS_KEY, |
| unexpected | apps/docgen-v2/src/s3.ts | 11 | secretKey: env.DOCGEN_V2_S3_SECRET_KEY, |
| unexpected | apps/docgen-v2/test/convert-pdf.unit.test.ts | 22 | process.env.DOCGEN_V2_SERVICE_TOKEN = TOKEN; |
| unexpected | apps/docgen-v2/test/convert-pdf.unit.test.ts | 23 | process.env.DOCGEN_V2_S3_ACCESS_KEY = 'key'; |
| unexpected | apps/docgen-v2/test/convert-pdf.unit.test.ts | 24 | process.env.DOCGEN_V2_S3_SECRET_KEY = 'sec'; |
| unexpected | apps/docgen-v2/test/convert-pdf.unit.test.ts | 121 | docgen_v2_version: 'docgen-v2@0.4.0', |
| unexpected | apps/docgen-v2/test/health.test.ts | 8 | process.env.DOCGEN_V2_SERVICE_TOKEN = 'test-token-0123456789'; |
| unexpected | apps/docgen-v2/test/health.test.ts | 9 | process.env.DOCGEN_V2_PORT = '0'; |
| unexpected | apps/docgen-v2/test/health.test.ts | 10 | process.env.DOCGEN_V2_S3_ACCESS_KEY = 'minioadmin'; |
| unexpected | apps/docgen-v2/test/health.test.ts | 11 | process.env.DOCGEN_V2_S3_SECRET_KEY = 'minioadmin'; |
| unexpected | apps/docgen-v2/test/render.hash-stability.test.ts | 34 | process.env.DOCGEN_V2_SERVICE_TOKEN = TOKEN; |
| unexpected | apps/docgen-v2/test/render.hash-stability.test.ts | 35 | process.env.DOCGEN_V2_S3_ACCESS_KEY = 'key'; |
| unexpected | apps/docgen-v2/test/render.hash-stability.test.ts | 36 | process.env.DOCGEN_V2_S3_SECRET_KEY = 'sec'; |
| unexpected | apps/docgen-v2/test/render.smoke.test.ts | 37 | process.env.DOCGEN_V2_SERVICE_TOKEN = TOKEN; |
| unexpected | apps/docgen-v2/test/render.smoke.test.ts | 38 | process.env.DOCGEN_V2_S3_ACCESS_KEY = 'key'; |
| unexpected | apps/docgen-v2/test/render.smoke.test.ts | 39 | process.env.DOCGEN_V2_S3_SECRET_KEY = 'sec'; |
| unexpected | apps/docgen-v2/test/s3.smoke.test.ts | 7 | DOCGEN_V2_PORT: 0, |
| unexpected | apps/docgen-v2/test/s3.smoke.test.ts | 8 | DOCGEN_V2_SERVICE_TOKEN: 'test-token-0123456789', |
| unexpected | apps/docgen-v2/test/s3.smoke.test.ts | 11 | DOCGEN_V2_S3_ENDPOINT: 'http://minio:9000', |
| unexpected | apps/docgen-v2/test/s3.smoke.test.ts | 12 | DOCGEN_V2_S3_ACCESS_KEY: 'k', |
| unexpected | apps/docgen-v2/test/s3.smoke.test.ts | 13 | DOCGEN_V2_S3_SECRET_KEY: 's', |
| unexpected | apps/docgen-v2/test/s3.smoke.test.ts | 14 | DOCGEN_V2_S3_BUCKET: 'b', |
| unexpected | apps/docgen-v2/test/s3.smoke.test.ts | 15 | DOCGEN_V2_S3_USE_SSL: false, |
| unexpected | apps/docgen-v2/test/validate-template.test.ts | 28 | process.env.DOCGEN_V2_SERVICE_TOKEN = TOKEN; |
| unexpected | apps/docgen-v2/test/validate-template.test.ts | 29 | process.env.DOCGEN_V2_S3_ACCESS_KEY = 'key'; |
| unexpected | apps/docgen-v2/test/validate-template.test.ts | 30 | process.env.DOCGEN_V2_S3_SECRET_KEY = 'sec'; |
| unexpected | apps/worker/cmd/metaldocs-worker/main.go | 26 | if deps.DocgenV2Client != nil && deps.SQLDB != nil { |
| unexpected | apps/worker/cmd/metaldocs-worker/main.go | 28 | pdfRunner := workerapp.NewPDFJobRunner(deps.DocgenV2Client, snapRepo) |
| unexpected | db/baseline/0001_current_schema.sql | 2032 | docgen_v2_ver text NOT NULL, |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 5 | IF to_regclass('public.templates_v2_template') IS NOT NULL |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 7 | EXECUTE 'ALTER TABLE public.templates_v2_template RENAME TO templates_template'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 14 | IF to_regclass('public.templates_v2_template_version') IS NOT NULL |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 16 | EXECUTE 'ALTER TABLE public.templates_v2_template_version RENAME TO templates_template_version'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 23 | IF to_regclass('public.templates_v2_approval_config') IS NOT NULL |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 25 | EXECUTE 'ALTER TABLE public.templates_v2_approval_config RENAME TO templates_approval_config'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 32 | IF to_regclass('public.templates_v2_audit_log') IS NOT NULL |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 34 | EXECUTE 'ALTER TABLE public.templates_v2_audit_log RENAME TO templates_audit_log'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 41 | IF to_regclass('public.templates_v2_audit_log_id_seq') IS NOT NULL |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 43 | EXECUTE 'ALTER SEQUENCE public.templates_v2_audit_log_id_seq RENAME TO templates_audit_log_id_seq'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 50 | IF to_regclass('public.idx_templates_v2_audit_template_time') IS NOT NULL |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 52 | EXECUTE 'ALTER INDEX public.idx_templates_v2_audit_template_time RENAME TO idx_templates_audit_template_time'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 55 | IF to_regclass('public.idx_templates_v2_template_tenant_doctype') IS NOT NULL |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 57 | EXECUTE 'ALTER INDEX public.idx_templates_v2_template_tenant_doctype RENAME TO idx_templates_template_tenant_doctype'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 60 | IF to_regclass('public.idx_templates_v2_version_template_status') IS NOT NULL |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 62 | EXECUTE 'ALTER INDEX public.idx_templates_v2_version_template_status RENAME TO idx_templates_version_template_status'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 65 | IF to_regclass('public.ux_templates_v2_system_blank') IS NOT NULL |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 67 | EXECUTE 'ALTER INDEX public.ux_templates_v2_system_blank RENAME TO ux_templates_system_blank'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 78 | AND conname = 'templates_v2_template_pkey' |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 80 | EXECUTE 'ALTER TABLE public.templates_template RENAME CONSTRAINT templates_v2_template_pkey TO templates_template_pkey'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 87 | AND conname = 'templates_v2_template_tenant_id_key_key' |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 89 | EXECUTE 'ALTER TABLE public.templates_template RENAME CONSTRAINT templates_v2_template_tenant_id_key_key TO templates_template_tenant_id_key_key'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 96 | AND conname = 'fk_templates_v2_published_version' |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 98 | EXECUTE 'ALTER TABLE public.templates_template RENAME CONSTRAINT fk_templates_v2_published_version TO fk_templates_published_version'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 109 | AND conname = 'templates_v2_template_version_pkey' |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 111 | EXECUTE 'ALTER TABLE public.templates_template_version RENAME CONSTRAINT templates_v2_template_version_pkey TO templates_template_version_pkey'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 118 | AND conname = 'templates_v2_template_version_template_id_version_number_key' |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 120 | EXECUTE 'ALTER TABLE public.templates_template_version RENAME CONSTRAINT templates_v2_template_version_template_id_version_number_key TO templates_template_version_template_id_version_number_key'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 127 | AND conname = 'templates_v2_template_version_template_id_fkey' |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 129 | EXECUTE 'ALTER TABLE public.templates_template_version RENAME CONSTRAINT templates_v2_template_version_template_id_fkey TO templates_template_version_template_id_fkey'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 140 | AND conname = 'templates_v2_approval_config_pkey' |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 142 | EXECUTE 'ALTER TABLE public.templates_approval_config RENAME CONSTRAINT templates_v2_approval_config_pkey TO templates_approval_config_pkey'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 149 | AND conname = 'templates_v2_approval_config_template_id_fkey' |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 151 | EXECUTE 'ALTER TABLE public.templates_approval_config RENAME CONSTRAINT templates_v2_approval_config_template_id_fkey TO templates_approval_config_template_id_fkey'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 162 | AND conname = 'templates_v2_audit_log_pkey' |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 164 | EXECUTE 'ALTER TABLE public.templates_audit_log RENAME CONSTRAINT templates_v2_audit_log_pkey TO templates_audit_log_pkey'; |
| unexpected | db/migrations/0203_rename_templates_v2_objects.sql | 170 | VALUES ('0203', 'rename templates v2 object names to current naming') |
| unexpected | docs/CHANGELOG.md | 6 | - `@eigenpal/docx-js-editor`-based editor under `/api/v2/*`. |
| unexpected | docs/CHANGELOG.md | 7 | - Content-addressed revision + export pipeline (docgen-v2 + Gotenberg). |
| unexpected | docs/CHANGELOG.md | 16 | - Legacy docgen client (replaced by docgen-v2 + Gotenberg). |
| unexpected | docs/CHANGELOG.md | 20 | - OpenAPI partials renamed: `documents-v2.yaml` → `documents.yaml`, `templates-v2.yaml` → `templates.yaml`. |
| unexpected | docs/CHANGELOG.md | 21 | - CI workflow renamed: `docx-v2-ci.yml` → `ci.yml`. |
| unexpected | docs/CHANGELOG.md | 22 | - Templates-v2 and documents-v2 nav items always visible (no feature flag guard). |
| unexpected | docs/ck5-wiki/02-versioning.md | 131 | - Export-to-PDF defaults to API v2 (output may differ). |
| unexpected | docs/ck5-wiki/03-engine-model.md | 231 | **Licensing**: the Inspector repo ships under a **dual license — GPL v2+ or |
| unexpected | docs/ck5-wiki/15-lists.md | 9 | Document list plugin (v2), list properties (start, reversed, styles), multi-block list items. |
| unexpected | docs/ck5-wiki/29-template-instantiation.md | 353 | - Publishing template v2 does not touch existing v1-derived documents. |
| unexpected | docs/ck5-wiki/29-template-instantiation.md | 357 | `POST /documents/:id/upgrade-template` → re-instantiate from v2, attempt |
| unexpected | docs/ck5-wiki/29-template-instantiation.md | 363 | renamed between v1 and v2. Defer until a concrete user need exists; the |
| unexpected | docs/ck5-wiki/35-backend-contracts.md | 330 | - **Breaking HTML schema changes bump `data-mddm-schema`** (v1 → v2) and |
| unexpected | docs/ck5-wiki/35-backend-contracts.md | 331 | ship with a migration under `migrations/html/v1-to-v2.ts`. |
| unexpected | docs/ck5-wiki/35-backend-contracts.md | 339 | - **API versioning**: REST routes live under `/api/v1/…`. A v2 prefix is |
| unexpected | docs/db-research/curated-baseline-inclusion-catalog.md | 41 | - templates v2 (`templates_v2_template`, `templates_v2_template_version`, `templates_v2_approval_config`, `templates_v2_audit_log`) |
| unexpected | docs/db-research/runtime-usage-inventory.md | 23 | - templates: `templates_v2_template`, `templates_v2_template_version`, `templates_v2_approval_config`, `templates_v2_audit_log` |
| unexpected | docs/release/docgen-name-classification.md | 3 | Phase 7 classification for release-facing `docgen-v2` / `docgen_v2` / `DocgenV2` names. |
| unexpected | docs/release/docgen-name-classification.md | 15 | \\| Service/workspace path \\| `apps/docgen-v2`, `@metaldocs/docgen-v2`, Dockerfile paths \\| deployment compatibility \\| Renaming affects workspace/package resolution and image build paths. \\| |
| unexpected | docs/release/docgen-name-classification.md | 16 | \\| Compose service \\| `deploy/compose/docker-compose.yml` service `docgen-v2` \\| deployment compatibility \\| API/worker env points at service DNS names in compose. \\| |
| unexpected | docs/release/docgen-name-classification.md | 17 | \\| Environment variables \\| `DOCGEN_V2_*`, `METALDOCS_DOCGEN_V2_*` \\| runtime compatibility \\| Existing `.env`, compose, docs, and scripts depend on these names. \\| |
| unexpected | docs/release/docgen-name-classification.md | 18 | \\| Go client/types \\| `DocgenV2Client`, `LoadDocgenV2Config` \\| runtime compatibility \\| Can be renamed only with env alias support and broad caller update. \\| |
| unexpected | docs/release/docgen-name-classification.md | 19 | \\| Persisted/export metadata \\| `docgen_v2_ver`, `docgen_v2_version`, values like `docgen-v2@0.4.0` \\| persisted compatibility \\| Composite hashes and export cache semantics include the renderer version string. \\| |
| unexpected | docs/release/docgen-name-classification.md | 20 | \\| Worker events \\| `docgen_v2_pdf` \\| persisted event compatibility \\| Existing outbox rows and worker dispatch require read compatibility. \\| |
| unexpected | docs/release/docgen-name-classification.md | 21 | \\| S3 defaults \\| `metaldocs-docx-v2` \\| storage compatibility \\| Bucket rename requires migration/copy or alias strategy. \\| |
| unexpected | docs/release/docgen-name-classification.md | 32 | This PR may rename generated API operation IDs and templates runtime objects, but it must not rename `docgen-v2` service/deployment/persisted names without the compatibility work above. |
| unexpected | docs/release/v2-operationid-inventory.md | 1 | # V2 Operation ID Inventory |
| unexpected | docs/release/v2-operationid-inventory.md | 3 | Generated operation IDs still carrying release-facing `V2` before Phase 6 cleanup. |
| unexpected | docs/release/v2-operationid-inventory.md | 39 | `docgen-v2` service and package names are not renamed in Phase 6. They require Phase 7 compatibility classification because they may be deployment, environment, or service-bound names rather than generated API operation names. |
| unexpected | docs/runbooks/docx-v2-w1-scaffold.md | 1 | # Runbook — docx-v2 W1 scaffold |
| unexpected | docs/runbooks/docx-v2-w1-scaffold.md | 10 | - New Fastify service `apps/docgen-v2` exposing only `/health` (port 3100). |
| unexpected | docs/runbooks/docx-v2-w1-scaffold.md | 11 | - New Postgres tables 0101–0108 (templates, template_versions, documents_v2, |
| unexpected | docs/runbooks/docx-v2-w1-scaffold.md | 22 | export DOCGEN_V2_SERVICE_TOKEN=$(openssl rand -hex 24) |
| unexpected | docs/runbooks/docx-v2-w1-scaffold.md | 24 | up -d postgres minio gotenberg docgen-v2 |
| unexpected | docs/runbooks/docx-v2-w1-scaffold.md | 25 | bash scripts/docx-v2-verify-migrations.sh |
| unexpected | docs/runbooks/docx-v2-w1-scaffold.md | 26 | bash scripts/docx-v2-seed-minio.sh |
| unexpected | docs/runbooks/docx-v2-w1-scaffold.md | 37 | document_revisions, editor_sessions, documents_v2, |
| unexpected | docs/runbooks/docx-v2-w1-scaffold.md | 41 | …and remove `docgen-v2` from the compose file. No data loss beyond new-path |
| unexpected | docs/runbooks/docx-v2-w1-scaffold.md | 48 | - `/render/docx` and other docgen-v2 routes return 404 by design. |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 1 | # Runbook: docx-v2 W2 Templates Vertical |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 10 | \\| `GET` \\| `/api/v2/templates` \\| admin, template_author, template_publisher \\| List all templates for tenant \\| |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 11 | \\| `POST` \\| `/api/v2/templates` \\| admin, template_author \\| Create a new template (seeds draft v1) \\| |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 12 | \\| `GET` \\| `/api/v2/templates/{id}/versions/{n}` \\| admin, template_author, template_publisher \\| Fetch a specific version \\| |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 13 | \\| `PUT` \\| `/api/v2/templates/{id}/versions/{n}/draft` \\| admin, template_author \\| Save draft content (storage keys + hashes) \\| |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 14 | \\| `POST` \\| `/api/v2/templates/{id}/versions/{n}/publish` \\| admin, template_publisher \\| Validate + publish a draft version \\| |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 15 | \\| `POST` \\| `/api/v2/templates/{id}/versions/{n}/docx-upload-url` \\| admin, template_author \\| Get presigned S3 PUT URL for DOCX \\| |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 16 | \\| `POST` \\| `/api/v2/templates/{id}/versions/{n}/schema-upload-url` \\| admin, template_author \\| Get presigned S3 PUT URL for schema JSON \\| |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 17 | \\| `GET` \\| `/api/v2/signed?key=<storage_key>` \\| admin, template_author, template_publisher \\| Redirect to presigned S3 GET URL \\| |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 28 | Author                   API (templates handler)     docgen-v2            Postgres |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 55 | 2. Validation always runs before the DB write — a failed `docgen-v2` call (5xx) returns 500 and does **not** advance version state. |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 67 | 1. Re-fetch the version via `GET /api/v2/templates/{id}/versions/{n}`. |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 93 | ## docgen-v2 /validate/template error taxonomy |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 95 | docgen-v2 returns a JSON array of error objects when validation fails (HTTP 200 with `valid: false`). |
| unexpected | docs/runbooks/docx-v2-w2-templates.md | 141 | Controls whether the W2 templates routes (`/api/v2/templates/*`) are registered in the router. |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 1 | # Runbook: docx-v2 W3 Documents Vertical |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 5 | All routes are under `/api/v2/documents` and require auth headers (`X-Tenant-ID`, `X-User-Roles`). |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 10 | \\| 1 \\| `POST` \\| `/api/v2/documents` \\| allow \\| allow \\| |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 11 | \\| 2 \\| `GET` \\| `/api/v2/documents/{id}` \\| allow \\| allow (owner only) \\| |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 12 | \\| 3 \\| `POST` \\| `/api/v2/documents/{id}/autosave/presign` \\| allow \\| allow (owner + active session) \\| |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 13 | \\| 4 \\| `POST` \\| `/api/v2/documents/{id}/autosave/commit` \\| allow \\| allow (owner + session holder) \\| |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 14 | \\| 5 \\| `POST` \\| `/api/v2/documents/{id}/session/acquire` \\| allow \\| allow (owner only) \\| |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 15 | \\| 6 \\| `POST` \\| `/api/v2/documents/{id}/session/heartbeat` \\| allow \\| allow (owner + session holder) \\| |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 16 | \\| 7 \\| `POST` \\| `/api/v2/documents/{id}/session/release` \\| allow \\| allow (owner + session holder) \\| |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 17 | \\| 8 \\| `POST` \\| `/api/v2/documents/{id}/session/force-release` \\| allow \\| deny \\| |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 18 | \\| 9 \\| `GET` + `POST` \\| `/api/v2/documents/{id}/checkpoints` \\| allow \\| allow (owner + active session for POST) \\| |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 19 | \\| 10 \\| `POST` \\| `/api/v2/documents/{id}/checkpoints/{versionNum}/restore` \\| allow \\| allow (owner + active session) \\| |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 20 | \\| 11 \\| `POST` \\| `/api/v2/documents/{id}/finalize` \\| allow \\| allow (owner only) \\| |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 21 | \\| 12 \\| `POST` \\| `/api/v2/documents/{id}/archive` \\| allow \\| allow (owner-only draft/archive path) \\| |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 22 | \\| 13 \\| `GET` \\| `/api/v2/documents` \\| allow \\| allow (filtered to own docs) \\| |
| unexpected | docs/runbooks/docx-v2-w3-documents.md | 44 | Route: `POST /api/v2/documents/{id}/session/force-release` |
| unexpected | docs/runbooks/docx-v2-w4-dogfood.md | 44 | 2. User B (no access to D) attempts `POST /api/v2/documents/{D}/export/pdf`. |
| unexpected | docs/runbooks/docx-v2-w4-dogfood.md | 98 | `docs/runbooks/docx-v2-w4-soak-evidence.md` and obtain sign-off from |
| unexpected | docs/runbooks/docx-v2-w4-exports.md | 5 | W4 adds PDF export (`POST /api/v2/documents/{id}/export/pdf`) and DOCX signed-URL |
| unexpected | docs/runbooks/docx-v2-w4-exports.md | 6 | (`GET /api/v2/documents/{id}/export/docx-url`) to the docx-v2 platform. PDFs are |
| unexpected | docs/runbooks/docx-v2-w4-exports.md | 7 | generated via docgen-v2 → Gotenberg and cached by composite hash (SHA-256 over |
| unexpected | docs/runbooks/docx-v2-w4-exports.md | 19 | \\| docgen-v2 `/convert-pdf` route \\| DOCX → PDF via Gotenberg \\| |
| unexpected | docs/runbooks/docx-v2-w4-exports.md | 29 | 4. **Restart docgen-v2**: requires `DOCGEN_V2_GOTENBERG_URL` in env (default `http://gotenberg:3000`) |
| unexpected | docs/runbooks/docx-v2-w4-exports.md | 31 | 6. **Smoke**: `curl -X POST /api/v2/documents/{id}/export/pdf` → 200 with `signed_url` |
| unexpected | docs/runbooks/docx-v2-w4-exports.md | 84 | 2. Revert `docgen-v2` to previous release. |
| unexpected | docs/runbooks/docx-v2-w4-exports.md | 99 | \\| PDF cache miss ratio > 80% \\| `cached=false` rate > 80% in 5-min window \\| Investigate docgen-v2 errors \\| |
| unexpected | docs/runbooks/docx-v2-w4-exports.md | 101 | \\| Gotenberg latency P99 > 30s \\| docgen-v2 `/convert-pdf` slow \\| Scale Gotenberg or check LibreOffice \\| |
| unexpected | docs/runbooks/docx-v2-w4-soak-evidence.md | 4 | > `docx-v2-w4-dogfood.md`. Commit it to the feature branch before |
| unexpected | docs/runbooks/docx-v2-w5-rollback.md | 41 | 2. `make maintenance-mode ENABLE=true` — 503 all /api/v2 routes. |
| unexpected | docs/runbooks/docx-v2-w5-rollback.md | 59 | - `curl /api/v2/templates` → 404 or unreachable |
| unexpected | docs/runbooks/docx-v2-w5-rollback.md | 70 | /api/v2 for real production data. |
| unexpected | docs/runbooks/security-baseline.md | 11 | - `go install github.com/securego/gosec/v2/cmd/gosec@latest` |
| unexpected | docs/standards/ENGINEERING_STANDARDS.md | 11 | - Breaking change apenas em `/api/v2`. |
| unexpected | docs/superpowers/archive/2026-04-01-content-builder-fixes.md | 9 | **Tech Stack:** React 18, TypeScript, CSS Modules (`DynamicEditor.module.css`), global CSS (`styles.css`), Vite, TipTap v2. |
| unexpected | docs/superpowers/archive/2026-04-02-governed-document-canvas-pilot.md | 385 | t.Fatalf("resolved template = %+v, want po-doc-special v2", got) |
| unexpected | docs/superpowers/archive/2026-04-02-po-schema-unification.md | 210 | Expected: Three rows (v1, v2, v3). v3 should have the longest `schema_json`. |
| unexpected | docs/superpowers/archive/2026-04-07-mddm-foundational-design.md | 8884 | When v2 swaps from PostgresByteaStorage to S3Storage: |
| unexpected | docs/superpowers/archive/2026-04-07-mddm-foundational-design.md | 8949 | - **Action**: when MDDM v2 is created (post-v1 launch), this becomes the test pattern. No task needed in v1 sprint since there are no migrations to test. |
| unexpected | docs/superpowers/archive/2026-04-07-mddm-foundational-design.md | 8953 | - Same as above — only needed after v1 launch when v2 ships |
| unexpected | docs/superpowers/archive/2026-04-10-mddm-engine-foundation.md | 887 | expect(isAllowlistedAssetUrl("/api/images_v2/foo")).toBe(false); |
| unexpected | docs/superpowers/archive/2026-04-10-mddm-engine-foundation.md | 1345 | // Example: MIGRATIONS[1] upgrades a v1 envelope to v2. |
| unexpected | docs/superpowers/archive/2026-04-12-mddm-react-parity.md | 538 | The adapter handles old→new DataTable format conversion internally on read WITHOUT bumping `mddm_version`. The version bump (1→2) is deferred to Phase 6 (Task 37.5) after all phases stabilize, so v2 represents the full parity schema — not just DataTable. |
| unexpected | docs/superpowers/archive/2026-04-18-ck5-pagination-v3.md | 13 | **Branch:** current HEAD on `main` (v2 dirty tree in working copy — revert before starting or stash). |
| unexpected | docs/superpowers/archive/2026-04-18-ck5-pagination-v3.md | 23 | \\| 0 — Clean slate (stash v2 dirty tree) \\| Opus \\| — \\| none \\| |
| unexpected | docs/superpowers/archive/2026-04-18-ck5-pagination-v3.md | 49 | - None at plan level. (The v2 dirty tree introduced a `PageFrames.tsx` in the working copy; Task 0 stash removes it from the working tree. It was never committed to `main`, so nothing to delete after stash.) |
| unexpected | docs/superpowers/archive/2026-04-18-ck5-pagination-v3.md | 59 | **Goal:** Discard v2 dirty working tree so v3 starts from known-good `main`. |
| unexpected | docs/superpowers/archive/2026-04-18-ck5-pagination-v3.md | 72 | - [ ] **Step 2: Stash v2 dirty tree** |
| unexpected | docs/superpowers/archive/2026-04-18-ck5-pagination-v3.md | 75 | git stash push -u -m "v2-pagination-broken-for-reference" -- frontend/apps/web/src/features/documents/ck5 shared/mddm-pagination-types |
| unexpected | docs/superpowers/archive/2026-04-18-ck5-pagination-v3.md | 78 | Expected: `Saved working directory and index state On main: v2-pagination-broken-for-reference`. |
| unexpected | docs/superpowers/archive/2026-04-18-ck5-pagination-v3.md | 94 | Expected: all tests pass (pre-v2 counts). Note the baseline number. |
| unexpected | docs/superpowers/archive/2026-04-18-ck5-pagination-v3.md | 963 | git stash list  # Confirm v2 stash still there. |
| unexpected | docs/superpowers/archive/2026-04-18-ck5-pagination-v3.md | 966 | Leave v2 stash in place for comparison reference. Do NOT drop it — user may want to diff behavior later. |
| unexpected | docs/superpowers/archive/2026-04-18-ck5-pagination-v3.md | 1013 | git stash pop  # restore v2 tree if needed for reference |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 59 | - **Fix:** New endpoint `GET /api/v2/controlled-documents/{id}/active-document` returns linked document instance. `RegistryDetailPage` fetches instance separately, passes correct `documentId` to panel. |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 71 | - **Verified:** `go test ./internal/modules/documents_v2/approval/http/...` passes. Full `go test ./...` clean. |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 74 | - **File:** `internal/modules/templates_v2/application/schema.go:106` |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 79 | - **File:** `frontend/apps/web/src/features/templates_v2/TemplateAuthorPage.tsx` |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 84 | - **File:** `internal/modules/documents_v2/repository/repository.go:44` |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 118 | 1. Browser DevTools console: `fetch('/api/v2/...', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({...}) })` |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 132 | ### documents-v2 routing |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 133 | App uses state-based routing (`docsRoute` state + `window.history.pushState`). Direct URL navigation to `/documents-v2/{id}` doesn't work — must enter via workspace sidebar button first. |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 203 | - 9.1: `POST /api/v2/documents/{id}/submit` → 200, `instance_id` returned ✅ |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 215 | - Fanout pipeline requires docgen-v2 service. Document ACTIVE but no frozen DOCX/PDF artifact. |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 216 | - Can be re-run once docgen-v2 is available. |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 239 | \\| F-09 \\| Stage 6 \\| `controlled_document_id` not persisted in `documents_v2` repo INSERT → NULL → `active-document` 404 \\| 🔴 Bug \\| FIX-07: added column to `repository.go` INSERT + backfill \\| |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 248 | \\| F-18 \\| Stage 10 \\| `z.coerce.boolean()` parsed env string `"false"` as `true` → MinIO used SSL on plain HTTP → `EPROTO tls_get_more_records` \\| 🔴 Bug \\| FIX-14: replaced with `z.enum(['true','false','1','0']).transform()` in `apps/docgen-v2/src/env.ts` \\| |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 249 | \\| F-19 \\| Stage 10 \\| `METALDOCS_DOCGEN_V2_URL` missing from `.env` → export handler not registered → `404` on PDF export \\| 🔴 Bug \\| FIX-15: added `METALDOCS_DOCGEN_V2_URL=http://localhost:3001` to `.env` \\| |
| unexpected | docs/superpowers/e2e-test-log-2026-04-26.md | 258 | \\| `POST /api/v2/documents/{id}/export/pdf` \\| ✅ 200, signed URL returned, PDF 13.5KB \\| |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 54 | \\| 1.1 \\| Navigate to `/templates-v2` \\| `preview_eval window.location.hash = '#/templates-v2'` \\| List page renders \\| |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 61 | \\| 1.8 \\| Click "Create Template" \\| `preview_click` \\| `POST /api/v2/templates → 201` \\| |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 68 | FROM templates_v2_template WHERE key = 'e2e-full-v1'; |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 80 | **Note (2026-04-26):** Template authoring uses the built-in ProseMirror editor — no DOCX upload. Authors type `{token_name}` directly in the canvas. DOCX is only generated at freeze/finalize time by docgen-v2 on the backend. |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 107 | \\| 2.13 \\| `preview_network` \\| network log \\| `GET /api/v2/templates/{id}/versions/1 → 200` \\| |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 113 | -- GET /api/v2/templates/{id}/versions/1 → placeholders array |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 138 | FROM templates_v2_template_version WHERE template_id = '{id}' AND version_number = 1; |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 201 | \\| 6.1 \\| Navigate to `/documents-v2` \\| `preview_eval` \\| Documents list \\| |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 205 | \\| 6.5 \\| Click Create \\| `preview_click` \\| `POST /api/v2/documents → 201` \\| |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 207 | \\| 6.7 \\| `preview_snapshot` \\| snapshot \\| Redirected to `/documents-v2/{docID}` editor \\| |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 255 | **Goal:** Document approved → freeze runs → docgen-v2 substitutes all 7 catalog tokens → DOCX uploaded to S3. |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 283 | **Note:** If fanout fails (docgen-v2 down or token mismatch), `preview_network` shows 500 on approve. Do NOT proceed to Stage 10b. |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 296 | 1. `POST /api/v2/documents/{id}/signoff` receives `decision`, `password`, `content_hash` |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 300 | 5. Fanout call dispatched to docgen-v2 with substitution map (all 7 catalog tokens resolved) |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 301 | 6. docgen-v2 fetches template DOCX from `metaldocs-attachments` bucket, applies substitutions, uploads frozen DOCX to `metaldocs-docx-v2` bucket |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 306 | POST /api/v2/documents/{id}/signoff |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 320 | **Bug 3 — `composition_config` required fields in docgen-v2 fanout route** |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 321 | - docgen-v2 fanout route required `header_sub_blocks`, `footer_sub_blocks`, and `sub_block_params` even when empty, causing validation errors on minimal payloads. |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 322 | - Fixed in `apps/docgen-v2/src/routes/fanout.ts`: made all three fields optional with defaults `[]` / `{}`. |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 324 | **Bug 4 — Template DOCX missing from `metaldocs-docx-v2` bucket** |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 325 | - The `metaldocs-docx-v2` bucket was empty; the source DOCX lives in `metaldocs-attachments`. |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 327 | - **Dev setup requirement:** manually copy the template DOCX from `metaldocs-attachments/templates/{id}/versions/` into `metaldocs-docx-v2/templates/{id}/versions/` before running the freeze pipeline locally. |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 338 | - MinIO: `frozen.docx` exists in `metaldocs-docx-v2` bucket |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 353 | **Goal:** After signoff, worker generates PDF → `GET /api/v2/documents/{id}/view` returns presigned PDF URL. PDF renders in browser with all 7 substituted catalog tokens. |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 361 | 3. `PDFDispatcher` publishes `docgen_v2_pdf` event to `metaldocs.outbox_events` table |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 363 | 5. `PDFJobRunner` calls docgen-v2 `/convert/pdf` synchronously with `final_docx_s3_key` |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 364 | 6. docgen-v2 converts DOCX → PDF, uploads to MinIO at `tenants/{id}/revisions/{id}/final.pdf`, returns `OutputKey` + `ContentHash` |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 366 | 8. `GET /api/v2/documents/{id}/view` reads `final_pdf_s3_key`, returns presigned URL → PDF renders |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 372 | - Fixed in `apps/docgen-v2/src/routes/convert-pdf.ts`. |
| unexpected | docs/superpowers/e2e-workflow-test-plan-2026-04-25.md | 387 | GET /api/v2/documents/32152e2f-e9cf-4ce2-bf38-0e36ce979cb5/view |

_Output truncated to first 300 unexpected rows. Use a larger -MaxRows for full inventory._
