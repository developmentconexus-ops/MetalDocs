FROM node:20.11-alpine AS build
WORKDIR /app

# Copy manifests for the workspace packages this service needs
COPY package.json package-lock.json ./
COPY packages/shared-tokens/package.json ./packages/shared-tokens/
COPY apps/docx-renderer/package.json ./apps/docx-renderer/

# Vendor tarball lives alongside the app (not in Go's /vendor root, which is
# excluded by .dockerignore)
COPY apps/docx-renderer/vendor ./apps/docx-renderer/vendor

# Install workspace deps (includes esbuild devDep + shared-tokens transitive deps)
RUN npm ci --workspace @metaldocs/docx-renderer --include-workspace-root

# Copy source for the bundle step
COPY packages/shared-tokens/src ./packages/shared-tokens/src
COPY apps/docx-renderer/src ./apps/docx-renderer/src
COPY apps/docx-renderer/build.mjs apps/docx-renderer/tsconfig.json ./apps/docx-renderer/

# Bundle our code + @metaldocs/shared-tokens into dist/index.js; node_modules
# deps stay external (see build.mjs). Then drop devDeps for a lean runtime.
RUN npm run build --workspace @metaldocs/docx-renderer \
 && npm prune --omit=dev

FROM node:20.11-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production

# Ship the pruned workspace: dist bundle + the external node_modules deps it
# imports at runtime (fastify, eigenpal, minio, etc.)
COPY --from=build /app ./

USER node
EXPOSE 3100
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:3100/health || exit 1
CMD ["node", "apps/docx-renderer/dist/index.js"]
