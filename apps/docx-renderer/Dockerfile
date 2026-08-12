# MetalDocs docx-renderer — pnpm workspace build (repo root context).
# Pre-pnpm history: this file once npm-installed a vendored eigenpal tarball from
# apps/docx-renderer/vendor. The repo has since migrated to pnpm workspaces and
# @eigenpal/* resolves from the registry via pnpm-lock.yaml — no vendor dir exists.
FROM node:22-alpine AS build
RUN corepack enable
WORKDIR /repo

# Manifests first for layer caching.
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/docx-renderer/package.json ./apps/docx-renderer/
COPY packages/docx-editor/package.json ./packages/docx-editor/
COPY packages/editor-ui/package.json ./packages/editor-ui/
COPY packages/eigenpal-adapter/package.json ./packages/eigenpal-adapter/
COPY packages/form-ui/package.json ./packages/form-ui/
COPY packages/shared-tokens/package.json ./packages/shared-tokens/
COPY packages/shared-types/package.json ./packages/shared-types/
COPY frontend/apps/web/package.json ./frontend/apps/web/

RUN pnpm install --frozen-lockfile --filter @metaldocs/docx-renderer...

# Source for the renderer + the workspace packages it imports.
COPY packages/shared-tokens/ ./packages/shared-tokens/
COPY packages/eigenpal-adapter/ ./packages/eigenpal-adapter/
COPY apps/docx-renderer/src ./apps/docx-renderer/src
COPY apps/docx-renderer/build.mjs apps/docx-renderer/tsconfig.json ./apps/docx-renderer/

# Bundle our code (+ workspace deps) into dist/index.js; registry deps stay
# external (see build.mjs). Then produce a lean prod-only deployment.
RUN pnpm --filter @metaldocs/docx-renderer run build \
 && pnpm --filter @metaldocs/docx-renderer --prod deploy --legacy /out \
 && cp -r apps/docx-renderer/dist /out/dist

FROM node:22-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /out ./
USER node
EXPOSE 3100
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:3100/health || exit 1
CMD ["node", "dist/index.js"]
