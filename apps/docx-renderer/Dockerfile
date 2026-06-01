FROM node:20.11-alpine AS build
WORKDIR /app

# Copy manifests for all workspace packages needed by this service
COPY package.json package-lock.json ./
COPY packages/shared-tokens/package.json ./packages/shared-tokens/
COPY apps/docx-renderer/package.json ./apps/docx-renderer/

# Vendor tarball lives alongside the app (not in Go's /vendor root)
COPY apps/docx-renderer/vendor ./apps/docx-renderer/vendor

# Install workspace deps (includes esbuild devDep)
RUN npm ci --workspace @metaldocs/docx-renderer --include-workspace-root

# Copy source files
COPY packages/shared-tokens/src ./packages/shared-tokens/src
COPY apps/docx-renderer/src ./apps/docx-renderer/src

# Bundle everything into a single self-contained ESM file.
# esbuild resolves @metaldocs/shared-tokens from the workspace and inlines it,
# eliminating the "unknown .ts extension" runtime error.
RUN npm run build --workspace @metaldocs/docx-renderer

FROM node:20.11-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production

# Only the bundle is needed — no node_modules copy required
COPY --from=build /app/apps/docx-renderer/dist ./dist
COPY --from=build /app/apps/docx-renderer/package.json ./package.json

USER node
EXPOSE 3100
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:3100/health || exit 1
CMD ["node", "dist/index.js"]
