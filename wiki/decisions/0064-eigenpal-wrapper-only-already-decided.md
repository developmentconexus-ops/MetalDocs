# ADR 0064 — Wrapper-only eigenpal consumption: already decided by ADR 0046 (closure record)

- **Status:** Accepted (ratifies an already-Accepted decision; no new rule added)
- **Last verified:** 2026-07-02
- **Date:** 2026-07-02
- **Scope:** Confirms that the "all `@eigenpal/*` access in `frontend/apps/web` goes through `@metaldocs/editor-ui`" rule referenced by backlog row R-008 (`wiki/backlog/editor-ui-eigenpal-refactor.md`) and tech-debt row T-008 (`wiki/modules/editor-ui-eigenpal-tech-debt.md`) is **already fully decided and enforced** by ADR 0046 (Eigenpal Anti-Corruption Layer). This ADR records that finding and formally closes the tech-debt row's "missing-ADR" status; it does not add a new rule.
- **Depends on:** ADR 0046 (Eigenpal Anti-Corruption Layer) — this ADR adds nothing to it.

---

## Context

The task brief for this decision batch (DEC-07g) asked for a 1-page ADR recording "wrapper-only eigenpal consumption" as a decision, citing `wiki/modules/editor-ui-eigenpal-tech-debt.md` R-008 and noting `third_party/eigenpal` vendoring. Verifying this against the existing decisions index surfaced that the work is already done, under a different ADR number, closing a different-but-identically-scoped tech-debt row ID.

### Verified runtime facts

- **The tech-debt row is T-008, not R-008** (R-008 is its linked backlog row in `wiki/backlog/editor-ui-eigenpal-refactor.md`, not the tech-debt item ID). `wiki/modules/editor-ui-eigenpal-tech-debt.md:49-55` — "### T-008 · No ADR for Anti-Corruption Layer / wrapper-only consumption rule" — **"Linked ADR: ADR 0046 — decision recorded; implementation in Phase 3."**
- **ADR 0046 (`wiki/decisions/0046-eigenpal-anti-corruption-layer.md`) already states, as Status: Accepted, Last verified 2026-06-26:** a two-wall Anti-Corruption Layer — `@metaldocs/eigenpal-adapter` (server wall, `apps/docx-renderer`) and `@metaldocs/editor-ui` (browser wall, `frontend/apps/web`) — with `frontend/apps/web` depending **only** on `@metaldocs/editor-ui` (ADR 0046 §Decision, "Browser wall" bullet). ADR 0046's own Consequences section states explicitly: **"Supersedes the previous 'editor-ui is the wrapper' pass-through arrangement; closes tech-debt T-008."**
- **Enforcement is live, not aspirational.** `eslint.config.mjs:16-25` — a repo-wide `no-restricted-imports` rule bans `@eigenpal`/`@eigenpal/*`/`@eigenpal/**` everywhere except the two wall packages (`packages/eigenpal-adapter/**`, `packages/editor-ui/**`, exempted at `eslint.config.mjs:62-63`), with the ESLint error message itself citing "ADR 0046." A guard test, `packages/editor-ui/test/public-surface.test.ts`, keeps the browser wall's exported surface vendor-free (ADR 0046 §Decision.6).
- **Vendoring note (task brief asked this be recorded): `third_party/eigenpal/` holds only a `NOTICE` file today** — the tarball itself (`eigenpal-docx-js-editor-0.2.0.tgz`) was retired 2026-06-23 per `wiki/modules/editor-ui-eigenpal-tech-debt.md` T-001 (RESOLVED): `@eigenpal/docx-editor-react@1.9.0` is now installed from the npm registry directly, not from a vendored tarball. `third_party/eigenpal/` is retained as the canonical home for licensing/`NOTICE` provenance (per user memory `eigenpal-vendor-third-party`), not as an active vendoring mechanism — it does not participate in the wrapper-only enforcement described above, which is a build-time (ESLint) and package-boundary (two walls) concern, orthogonal to where the npm package is sourced from.

## Decision

**No new decision is made here.** ADR 0046 is confirmed to already be the complete, accepted, enforced answer to "wrapper-only eigenpal consumption": exactly two packages (`packages/eigenpal-adapter`, `packages/editor-ui`) may import `@eigenpal/*`; every other consumer (including all of `frontend/apps/web`) must go through `@metaldocs/editor-ui`'s public surface. This is enforced today by lint (`eslint.config.mjs`) plus a guard test, not merely documented.

The only actionable item from this verification pass is bookkeeping: `wiki/modules/editor-ui-eigenpal-tech-debt.md` T-008 should be marked **CLOSED**, since its own "Linked ADR" field already names the closing decision (ADR 0046) but the row heading lacks the `— CLOSED` marker other resolved rows in the same file use (compare T-001, T-002, T-003, all of which carry an explicit `— RESOLVED`/`— CLOSED` suffix in their heading).

## Consequences

- T-008 (`wiki/modules/editor-ui-eigenpal-tech-debt.md`) heading is updated (by this same change) to add the `— CLOSED` marker, consistent with the file's own convention; its body already correctly names ADR 0046.
- No new architectural surface, rule, or enforcement is introduced by this ADR.
- Anyone tracking DEC-07g in the grade-A simplification register should record it as "already satisfied by ADR 0046" rather than "new ADR written," to avoid the appearance of two ADRs governing the same boundary.

## References

- ADR [`0046-eigenpal-anti-corruption-layer.md`](0046-eigenpal-anti-corruption-layer.md) — the actual decision; Status Accepted, closes T-008 per its own Consequences section.
- `wiki/modules/editor-ui-eigenpal-tech-debt.md` T-008 — tech-debt row, heading updated to CLOSED by this ADR's companion edit.
- `wiki/backlog/editor-ui-eigenpal-refactor.md` R-008 — linked backlog row (task brief's "R-008" reference; the tech-debt item itself is T-008).
- `eslint.config.mjs:16-25,62-63` — live enforcement.
- `packages/editor-ui/test/public-surface.test.ts` — guard test.
- `third_party/eigenpal/NOTICE` — vendoring provenance record (tarball retired 2026-06-23; package now sourced from npm registry).
