FROM node:20.11-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
COPY apps/docx-renderer/package.json ./apps/docx-renderer/
RUN npm ci --workspace @metaldocs/docx-renderer --include-workspace-root
COPY apps/docx-renderer ./apps/docx-renderer
RUN npm run build --workspace @metaldocs/docx-renderer

FROM node:20.11-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/apps/docx-renderer/dist ./dist
COPY --from=build /app/apps/docx-renderer/package.json ./package.json
USER node
EXPOSE 3100
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:3100/health || exit 1
CMD ["node", "dist/index.js"]
