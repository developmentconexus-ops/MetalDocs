# MetalDocs web — pnpm workspace build (repo root context).
# The repo is a pnpm monorepo (packageManager pnpm@11.8.0, workspace:* deps);
# @metaldocs/web depends on workspace packages (editor-ui, shared-tokens), so the
# install MUST run through pnpm at the workspace root — npm ci cannot resolve
# workspace:* and there is no package-lock.json.
FROM node:26.3.0-alpine AS builder
RUN npm install --global pnpm@11.8.0
WORKDIR /repo

# Manifests first for layer caching: lockfile + workspace config + every package.json.
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY frontend/apps/web/package.json ./frontend/apps/web/
COPY packages/docx-editor/package.json ./packages/docx-editor/
COPY packages/editor-ui/package.json ./packages/editor-ui/
COPY packages/eigenpal-adapter/package.json ./packages/eigenpal-adapter/
COPY packages/form-ui/package.json ./packages/form-ui/
COPY packages/shared-tokens/package.json ./packages/shared-tokens/
COPY packages/shared-types/package.json ./packages/shared-types/
COPY apps/docx-renderer/package.json ./apps/docx-renderer/

RUN pnpm install --frozen-lockfile --filter @metaldocs/web...

# Source: the web app + the workspace packages it consumes.
COPY packages/ ./packages/
COPY frontend/apps/web/ ./frontend/apps/web/
# The web build generates API types from the OpenAPI contract at gen time in dev;
# the committed src/lib/api-types are used as-is here (contract-first: spec is truth).

RUN pnpm --filter @metaldocs/web run build

FROM nginx:1.27-alpine
COPY --from=builder /repo/frontend/apps/web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /repo/frontend/apps/web/dist /usr/share/nginx/html
EXPOSE 4173
