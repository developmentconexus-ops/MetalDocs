# F1 — Plan

Seeded from master plan `docs/superpowers/plans/2026-07-07-m2c-approval-screen-fe.md` §F1.
TDD via a fresh implementer subagent + independent reviewer subagent (milestone Phase 3 steps 6).

## Tasks

1. **[TDD] Failing test** — `src/trackChanges.test.tsx`: mount `MetalDocsEditor mode="review"` with the
   vendor mocked at the module boundary (`extractTrackedChanges` + the command curries), fixture =
   one `insertion` + one coalesced `replacement`. Assert neutral shape, string id, replacement
   resolves all ids, accept-all empties, `onTrackedChangesChange` fires. RED first.
2. **types.ts** — neutral `TrackedChange` + MetalDocs-owned `TrackedChangeType` literal union; 6 ref
   methods on `MetalDocsEditorRef`; `onTrackedChangesChange` prop.
3. **index.ts** — export `TrackedChange` + `TrackedChangeType`.
4. **MetalDocsEditor.tsx** — vendor imports (core commands + `extractTrackedChanges`, INTERNAL only);
   `onTrackedChangesChangeRef` (stale-closure-safe); component-scope helpers `getBodyView`,
   `readTrackedChanges` (map vendor→neutral), `resolveIds` (union primary+insertion+coalesced,
   deduped), `notifyTrackedChanges`; the 6 imperative-handle members (each null-guards the body view);
   `handleChange` calls `notifyTrackedChanges` after `props.onChange?.()`.
5. **Verify** — `npm run test` (targeted 3 files) GREEN; `npm run typecheck` clean; ACL-wall grep zero.
6. **Review pass** — independent reviewer subagent: spec compliance + code quality + test-meaning.
   Apply accepted findings.

## Review-fix addendum (post-reviewer, main session)

Reviewer verdict **APPROVE**, 3 Minors — all applied:
- M1: `TrackedChange.type` widened from bare `string` → MetalDocs-owned `TrackedChangeType` literal
  union (exhaustiveness for consumers; wall stays closed — union re-declared, not vendor-imported).
- M2: body-only scope + the ref API documented in `wiki/modules/editor-ui-eigenpal.md` §8.4.
- M3: invariant comment on `resolveIds` (the id passed is always a primary entry-level `revisionId`).

## Files touched

- `packages/editor-ui/src/types.ts`
- `packages/editor-ui/src/index.ts`
- `packages/editor-ui/src/MetalDocsEditor.tsx`
- `packages/editor-ui/src/trackChanges.test.tsx` (new)
- `wiki/modules/editor-ui-eigenpal.md` (§8.4 subsection + changelog)
