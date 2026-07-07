# F1 — Evidence

## Commands + real output

- `npm run test -- src/trackChanges.test.tsx src/MetalDocsEditor.test.tsx src/MetalDocsEditor.tokens.test.tsx`
  → `Test Files 3 passed (3) · Tests 16 passed (16)` (trackChanges 5, mode-mapping 2, tokens 9).
- `npm run typecheck` (`tsc -p tsconfig.json --noEmit`) → clean, no output.
- ACL-wall grep — `types.ts`/`index.ts` for `@eigenpal|prosemirror|EditorState|EditorView|Command|TrackedChangeEntry`
  → **zero** matches (reviewer-confirmed). Vendor imports confined to `MetalDocsEditor.tsx`.

## TDD proof

- Implementer subagent ran `src/trackChanges.test.tsx` **before** implementation → all 5 fail with
  `TypeError: ref.current.X is not a function` (methods absent). RED confirmed, toolchain healthy (no
  junction drift). GREEN after implementing the 6 members + helpers. Test authored before code.

## Runtime proof (observable change) + fixture-vs-real

- **Unit level (real vitest, mocked vendor):** the adapter's mapping + forwarding contract is asserted
  against a mutable fixture — neutral shape, string `revisionId`, a `replacement` accept resolving ids
  `8+9+10` together (integrity), accept-all emptying the set, `onTrackedChangesChange` firing with the
  post-accept set. Labeled **fixture/mock** — the vendor is stubbed at the module boundary because
  jsdom has no real ProseMirror.
- **Real end-to-end** (accept/reject actually mutating the docx in a live editor) is the vendor's own
  tested concern and is deferred to the **F8 live-QA walkthrough** (review a real instance, accept a
  suggestion, observe the change clear). Registered there, not claimed here.

## Review / QA disposition

- Independent reviewer subagent (separation from implementer): **APPROVE**, 0 Critical, 0 Major, 3
  Minor. All 3 Minors applied by the main session (type union M1, wiki doc M2, invariant comment M3) —
  see `plan.md` addendum. Reviewer re-confirmed GREEN 16/16 + typecheck clean + wall intact, and
  judged the test meaningful (not tautological): the replacement-resolves-all assertion is the one
  that matters and it is asserted precisely.

## Bounded defers

- None new. The server-authoritative suggestion-resolution gate remains the F0/HS-2 program-level
  bounded defer (client-authoritative eigenpal resolution + freeze hash chain covers today).
