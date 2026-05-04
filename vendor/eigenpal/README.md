# Vendored EigenPal package

This directory contains the MetalDocs controlled EigenPal package artifact.

## Artifact

- File: `eigenpal-docx-js-editor-0.2.0.tgz`
- Package: `@eigenpal/docx-js-editor`
- Version: `0.2.0`
- SHA256: `9bcfc833bc2c104a2d8baff6dd8c174eb7c97c113f06266109f3b05204998245`

## Source of truth

- Fork repository: `https://github.com/leandrotcawork/docx-editor`
- Stable branch: `main`
- Stable tag: `metaldocs-eigenpal-v0.2.0`
- Tag target commit: `7cbadeb77eaa97f2e0a07fd7d9a2671fe3f7a753`

## Why vendored

The EigenPal source is a monorepo, and MetalDocs consumes the packaged React editor shape.
Vendoring keeps installs deterministic while updates happen through explicit tag bumps.

## Refresh procedure

1. In the EigenPal fork, create and approve the next stable tag.
2. Pack from that exact tag commit (`npm pack ./packages/react`).
3. Replace this `.tgz` file.
4. Reinstall dependencies in MetalDocs root and `frontend/apps/web` to refresh lockfiles.
5. Run the checklist in `wiki/modules/editor-ui-eigenpal.md`.
