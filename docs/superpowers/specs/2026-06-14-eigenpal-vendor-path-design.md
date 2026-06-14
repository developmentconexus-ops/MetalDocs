# Eigenpal Vendor Path — Design Spec

> **Date:** 2026-06-14
> **Status:** Approved (brainstorming)
> **Scope:** Fix broken/drifted `file:` references to the vendored `@eigenpal/docx-js-editor` tarball so a fresh checkout installs cleanly. Establish one canonical, Go-safe, app-neutral vendor home. Sequence the deeper architecture moves as post-v1 follow-ups.
> **Origin:** Discovered during the Grade-A remediation M0/F0.1 ADR audit; documented in `wiki/decisions/0001-eigenpal-adoption.md` (Path note 2026-06-14, deferred as HS-2).

## Problem

`@eigenpal/docx-js-editor` is consumed as a checked-in tarball (`eigenpal-docx-js-editor-0.2.0.tgz`) via npm/pnpm `file:` references. The tarball physically exists at **one** location but is referenced from **three** `package.json` files using **two** different relative-path intents — two of which point at a directory that does not exist. A fresh install fails; current `node_modules` survives only from a prior install.

### Current state (verified 2026-06-14)

Three direct consumers (all import eigenpal source directly — the ACL is leaky, so none can be dropped without a refactor):

| Consumer | `file:` ref | Resolves to | Status |
|---|---|---|---|
| `apps/docx-renderer/package.json` | `./vendor/eigenpal/...tgz` | `apps/docx-renderer/vendor/eigenpal/` | ✅ tarball physically here |
| `packages/editor-ui/package.json` | `../../vendor/eigenpal/...tgz` | repo-root `vendor/eigenpal/` | ❌ missing |
| `frontend/apps/web/package.json` | `../../../vendor/eigenpal/...tgz` | repo-root `vendor/eigenpal/` | ❌ missing |

Direct-import evidence:
- `frontend/apps/web/src/editor-adapters/eigenpal-template-mode.ts` (+ spike/test files) import `@eigenpal/docx-js-editor/core`.
- `apps/docx-renderer/src/render/fanout.ts` + `build.mjs` import eigenpal directly.
- `packages/editor-ui/src/*` (7 files) import eigenpal directly — the ACL wrapper.

### Two structural facts driving the design

1. **Repo-root `vendor/` is Go-owned.** It holds `vendor/github.com/...` + `modules.txt`; `go mod vendor` **prunes** anything not in `modules.txt`. A JS tarball placed at root `vendor/eigenpal/` would be deleted on the next `go mod vendor`. → The "establish root `vendor/eigenpal/`" option is unsafe and rejected.
2. **Two package-manager universes.** Root `package.json` (`metaldocs-monorepo`) is an **npm** workspace (`apps/docx-renderer`, `packages/*`, npm `-ws`). `frontend/apps/web` is a **standalone pnpm** project with its own `frontend/apps/web/pnpm-lock.yaml`, reaching up into root `packages/*` via `file:` links. Two lockfiles, hand-maintained cross-references — this is the root cause of the drift class, addressed by follow-up F1, not by this change.

## Decision

Establish **one canonical, Go-safe, app-neutral vendor home** at repo root and point all three references at it. Defer the package-manager unification and registry migration as sequenced post-v1 ADRs.

### Why `third_party/` (not `apps/docx-renderer/vendor/`)

- Root `vendor/` — unsafe (Go prunes). Rejected.
- `apps/docx-renderer/vendor/` — Go-safe and the current ADR canonical, but forces the other two packages to reach into a sibling app's internals (e.g. `frontend/apps/web` → `apps/docx-renderer/vendor`), bad coupling for a shared dependency.
- `third_party/eigenpal/` — conventional name for vendored upstream artifacts, Go-safe, owned by no app, and **forward-compatible with F1**: when the repo collapses to a single workspace, an app-neutral root `third_party/` is already where that workspace wants the artifact. Pay the one small ADR-path update now rather than moving the path twice.

## Plan — bounded install fix (do now)

1. **Relocate the tarball.** `git mv apps/docx-renderer/vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`. Remove the now-empty `apps/docx-renderer/vendor/eigenpal/` (and `apps/docx-renderer/vendor/` if nothing else lives there).
   - **Verify:** `apps/docx-renderer/vendor/` holds nothing but the eigenpal dir before deleting it.
2. **Repoint all three `package.json` refs** to the single copy:
   - `apps/docx-renderer/package.json` → `file:../../third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`
   - `packages/editor-ui/package.json` → `file:../../third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`
   - `frontend/apps/web/package.json` → `file:../../../third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`
   - **Verify:** each relative path, resolved from its package dir, lands on the real `third_party/eigenpal/...tgz`.
3. **Regenerate both lockfiles.**
   - Root npm workspace: `npm install` (regenerates root lockfile / `package-lock.json` per the npm workspace). Confirm the root install resolves editor-ui + docx-renderer eigenpal refs.
   - Web pnpm app: `pnpm install` in `frontend/apps/web` (regenerates `frontend/apps/web/pnpm-lock.yaml`; integrity + tarball path lines update).
   - **Verify:** both installs exit clean; the `@eigenpal/docx-js-editor` integrity hash in the pnpm lock is unchanged (same tarball, new path only).
4. **Update ADR `wiki/decisions/0001-eigenpal-adoption.md`.**
   - Lines ~19 and ~43: change canonical path from `apps/docx-renderer/vendor/eigenpal/` to `third_party/eigenpal/`.
   - Replace the 2026-06-14 Path note / HS-2 defer with a resolution note (canonical home = `third_party/eigenpal/`; all three refs aligned; HS-2 closed).
   - Bump `Last verified:` stamp to 2026-06-14.
   - **Verify:** no remaining doc text claims the tarball lives under `apps/docx-renderer/vendor/`.
5. **Grep for stragglers.** Search the repo for any other `vendor/eigenpal` or old-path references (scripts, Dockerfiles, CI config, build.mjs, vitest configs, docs) and update them.
   - **Verify:** `grep -rn "vendor/eigenpal"` returns only `third_party/eigenpal` matches (or none in the old form).

### Acceptance criteria

- From a **fresh checkout** (clean, no surviving `node_modules`): root `npm install` **and** `frontend/apps/web` `pnpm install` both succeed with no missing-tarball error.
- `@eigenpal/docx-js-editor` resolves in all three consumers; web typecheck/build (`pnpm build`) and docx-renderer build run against the relocated artifact.
- `go mod vendor` can run without touching/deleting the JS tarball (it lives outside `vendor/`).
- ADR 0001 reflects `third_party/eigenpal/` as canonical; HS-2 defer cleared.
- No old-path references remain anywhere in the repo.

### Out of scope / explicit non-goals (now)

- No refactor of the eigenpal ACL (web + docx-renderer keep their direct imports).
- No package-manager migration; the two-universe split stays as-is for v1.
- No registry work.

## Follow-ups (post-v1, each its own ADR + full regression)

- **F1 — Single pnpm workspace.** Collapse the npm root workspace + standalone pnpm web app into one pnpm workspace: one root `pnpm-workspace.yaml`, one lockfile, `workspace:` protocol for internal packages. Dissolves the `file:`-cross-link drift class entirely. This is the real Tier-2 move and the first architecture ADR after the v1 release.
- **F2 — Tighten the eigenpal ACL.** Route all eigenpal usage through `packages/editor-ui` so web + docx-renderer depend on `@metaldocs/editor-ui` (not the raw tarball). Reduces direct-consumer fan-out to one.
- **F3 — Publish the eigenpal fork to a private registry** (GitHub Packages / Artifactory / Verdaccio); consume by semver + integrity hash; retire vendoring. After the fork stabilizes; needs registry + CI-auth decisions.

## Industry context (why this sequencing)

- **Tier 1** (Google/Meta): monorepo build graph (Bazel/Buck), vendored *source*. Overkill here.
- **Tier 2** (most large product SaaS): single JS workspace (pnpm/yarn + Turborepo/Nx), one lockfile; forks via private registry, consumed by semver. ← target state = F1 + F3.
- **Tier 3** (transitional): checked-in `.tgz` via `file:`. Current state; drift-prone, as this bug demonstrated.

This change cleans up Tier 3 and chooses a location forward-compatible with the Tier-2 target, without destabilizing the v1 build on release day.
