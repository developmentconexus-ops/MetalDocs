# Refactor Backlog — editor-ui-eigenpal

> Actionable rows. One row = one PR. Pulled from `wiki/modules/editor-ui-eigenpal-tech-debt.md`.

**Last verified:** 2026-05-17

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Restore vendored eigenpal tarball or pin to published @eigenpal package | T-001 | M | Critical | — | — | closed 2026-05-11 | Plan 3 |
| R-002 | Migrate `TemplateEditorPage` to consume `MetalDocsEditor` instead of `DocxEditor` directly | T-002 | M | Major | R-001 | — | open | — |
| R-003 | Rewrite `templatePlugin.wiring.test.tsx` against the current `template-draft` gate | T-003 | S | Major | - | - | closed 2026-05-11 | Plan 11 |
| R-004 | Delete `createOutlinePlugin` export (or re-register if outline panel is brought back) | T-004 | XS | Minor | — | — | open | — |
| R-005 | Rename `plugins/mergefieldPlugin.ts` to `sidebarModel.ts`; drop the `Plugin` suffix | T-005 | XS | Minor | — | — | open | — |
| R-006 | Wire `onLockLost` through to a real eigenpal lock-loss event, or remove the prop | T-006 | XS | Minor | — | — | open | — |
| R-007 | Write ADR for `templatePlugin` mode gating rule | T-007 | S | Minor | — | — | open | — |
| R-008 | Write ADR for wrapper-only consumption boundary (`@eigenpal/docx-js-editor` only via `@metaldocs/editor-ui`) | T-008 | S | Minor | — | — | open | — |
| R-009 | Refresh `wiki/references/eigenpal-controlled-package.md` and ADR 0001 § Consequences to reflect `vendor/eigenpal/` removal | maint:doc-cleanup | XS | Minor | R-001 | — | open | — |
| R-010 | Bump eigenpal package to next fork build once upstream PR series lands | maint:dep-bump | M | Major | R-001 | — | open | — |

## Notes

- R-001 is the gating row: until the install break is fixed, no other refactor lands cleanly on a fresh checkout.
- R-002 depends on R-001 because exercising the migrated `TemplateEditorPage` from a clean install requires the resolvable package.
- R-007 and R-008 are sibling ADR PRs; R-008 implicitly closes T-002's "missing-ADR" subclaim once the wrapper-only rule has a decision record.
