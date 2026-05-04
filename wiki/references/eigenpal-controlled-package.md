# EigenPal Controlled Package

> **Last verified:** 2026-05-01
> **Scope:** What MetalDocs needs to know about the controlled EigenPal package.
> **Out of scope:** Internal EigenPal implementation details; keep those in the fork docs.
> **Key files:**
> - `vendor/eigenpal/README.md` - artifact source, fork branch, pack command
> - `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` - vendored package artifact
> - `packages/editor-ui/src/MetalDocsEditor.tsx` - React wrapper used by MetalDocs
> - `apps/docgen-v2/package.json` - server-side docgen dependency
> - `packages/editor-ui/package.json` - editor-ui dependency
> - `frontend/apps/web/package.json` - web dependency

---

## Current state

MetalDocs consumes a controlled EigenPal package:

```text
@eigenpal/docx-js-editor -> vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz
```

This artifact is built from the MetalDocs-controlled EigenPal fork, then vendored into this repository for deterministic installs.

## Why vendored

The EigenPal source is a monorepo. The package MetalDocs needs is produced from the EigenPal React package, not from the repository root. A vendored tarball keeps MetalDocs stable until the fork is published as a package or the upstream PR series is accepted.

## What belongs in MetalDocs docs

- Which package artifact MetalDocs consumes.
- Which MetalDocs modules import it.
- How to refresh/reinstall after a new EigenPal build.
- Which smoke checks prove the integration still works.

## What does not belong here

- Header/footer rendering internals.
- Table layout and ProseMirror table command internals.
- DOCX serialization implementation details.
- Historical debugging notes from the EigenPal lab.

Those details belong in the EigenPal fork docs and the local lab dossier referenced by `vendor/eigenpal/README.md`.

## Refresh checklist

1. Build/package the EigenPal fork package.
2. Replace the tarball in `vendor/eigenpal/`.
3. Reinstall dependencies at the MetalDocs root and in `frontend/apps/web`.
4. Commit lockfile updates together with the tarball update.
5. Run the editor validation checklist in `wiki/modules/editor-ui-eigenpal.md`.
