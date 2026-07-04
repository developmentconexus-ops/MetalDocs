# F4.2 — scheduled-vs-manual publish race

> **Contract:** `../validation-contract.md` §2. Feature home. **Approved for code: 2026-07-04.**

## Consumer contract

- **Consumer = correctness reviewer / the invariant "a revision publishes at most once."** Requires a
  *real* concurrent integration test (testdb factory, real Postgres, NOT sqlmock) proving that when the
  manual-publish path (api binary, `approved→published`) and the scheduled-cutover path (jobs binary,
  `scheduled→published`) race ONE document revision, **exactly one wins**, the loser no-ops with the
  correct sentinel, and the terminal state is a single `published` revision with `revision_version`
  bumped exactly once.
- If the proof fails, the consumer requires a single `PublishRevision` choke point both paths route
  through, and the test green against it.

## Non-goals

- NOT running the full integration suite (targeted `-run` only).
- NOT changing publish business logic unless the race proves unsafe (then: minimal single choke point).
- NOT the state-machine work (F4.1) or idiom (F4.3).

## Validation gate

Per contract §2.4. Real concurrent integration test, both interleavings deterministic, exactly-one-winner
+ correct terminal state + correct loser sentinel + single publish side-effect. Expected (per contract
§0.4): passes as-is (mutually exclusive `status` predicates + OCC `revision_version` CAS). If not, add
`PublishRevision` choke point (HS-2-bounded) and re-prove.

## Interview record

Race safety is asserted-but-unverified in the codebase (review §6 RE-LITIGATE #2). Operator goal: prove
it real. Decision: prove-first; choke point only if proof fails.
