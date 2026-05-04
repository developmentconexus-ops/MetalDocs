# Vendored EigenPal package

This directory contains the Metal Docs controlled EigenPal package artifact used by the integration branch.

## Artifact

- `eigenpal-docx-js-editor-0.2.0.tgz`
- Package name: `@eigenpal/docx-js-editor`
- Package version inside tarball: `0.2.0`

## Source

- Fork: https://github.com/leandrotcawork/docx-editor
- Branch: `codex/eigenpal-professional-patch`
- Source commit used when packed: `7cbadeb77eaa97f2e0a07fd7d9a2671fe3f7a753`
- Internal PR: https://github.com/leandrotcawork/docx-editor/pull/1

## Why this is vendored

The EigenPal repository is a monorepo and the npm package lives in `packages/react`. A direct GitHub dependency on the repository root does not install the published package shape. The tarball keeps this integration deterministic until we publish the fork package or split the upstream PR series.

## Refresh command

From the Metal Docs repository root, after building the EigenPal fork package:

```powershell
npm pack "C:\Users\leandro.theodoro.MN-NTB-LEANDROT\Documents\MetalDocs\non_git\eigenpal-isolated-lab\analysis\eigenpal-upstream-source\packages\react" --pack-destination ".\vendor\eigenpal"
```

After refreshing, reinstall dependencies in the Metal Docs root and `frontend/apps/web` so the lockfiles capture the new tarball integrity.