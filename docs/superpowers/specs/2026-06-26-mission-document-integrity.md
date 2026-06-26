# Mission — Document Contract & Storage Integrity (Grade-A)

- **Date:** 2026-06-26
- **Status:** Approved design, amended after adversarial review — pending implementation plans
- **Workstreams:**
  - A — [Contract Integrity](2026-06-26-contract-integrity-design.md)
  - B — [Storage Integrity](2026-06-26-storage-integrity-design.md)

This overview owns what neither workstream owns alone: sequencing, the coupled
deploy/rollout story, the terminal end-to-end acceptance gate, and governance.

---

## 1. Why this mission

Two structural cracks in document create / store / access, both pre-existing on
`main` (not eigenpal regressions):

- **A — the contract is asserted, not enforced.** A half-applied ADR-0035 flat
  migration + hand-rolled shape assertions crash the template editor.
- **B — pointers declare intent, not existence.** `docx_storage_key` is an eager
  derived constant; the object is written lazily, so the DB can advertise a docx
  that 404s.

## 2. Sequencing (verified)

**A first, then B.** Confirmed against code: the editor mount is blocked *only* by
`getTemplateSchemas` throwing (A). A missing docx (B) already degrades to a blank
buffer via `useTemplateDraft`'s `if (res.ok)` / `if (version.docx_storage_key)`
guards — it does not crash. So A alone re-mounts the editor; B then makes the
stored docx trustworthy. There is no hidden reverse dependency.

**One coupling to honor:** B's read-side "not-ready / not-found" response on the
docx-url path is a *contract* change. It MUST be declared in A's contract law
(OpenAPI + typed body) so B implements a handler against a shape A already froze.
`getDocxURL` (today enveloped) is therefore migrated under A, not B. See A §4.3.

## 3. Coupled deploy & migration ordering (mandatory)

Both workstreams change externally-observable contracts (A: wire shapes; B: DB
schema). Naive deploy re-creates the exact `undefined`-deref crash class we are
fixing. Rules:

1. **A wire-shape flips are breaking.** Ship backend + regenerated frontend types
   **in lockstep** (same release), OR introduce a tolerant reader on the frontend
   first. Do not deploy a new backend shape against an old frontend bundle. The
   migration is done endpoint-by-endpoint; each endpoint's BE+FE change is atomic.
2. **B migration is forward-only.** The `docx_storage_key` nullable migration +
   backfill nulls keys that were never committed. There is **no down-migration to
   NOT NULL** (impossible once nulls exist). Deploy code that tolerates null
   *before* the backfill runs.
3. **Backfill signal is `content_hash`, not an object-store scan.** A committed
   (verified) key always has `content_hash != ''`; a never-committed eager key has
   `content_hash = ''`. Backfill nulls `docx_storage_key` where `content_hash = ''`.
   This is O(rows) in the DB only — no per-row S3 Stat sweep. (See B §8.)

## 4. Terminal acceptance gate (mission-level, runs after B)

Neither workstream's unit ACs prove the pipeline composes. The mission is not done
until a real **create → store → render → download** E2E passes — the original
Claude-Preview QA intent that the crash blocked:

1. Create a template → editor mounts (proves A).
2. Autosave → commit persists a **server-verified** docx key (proves B write-side).
3. Reopen the version → docx bytes load and render in the editor (proves A+B at the
   autosave-commit boundary).
4. Instantiate a document from the published template → freeze → materialize →
   download the final docx/PDF succeeds (proves the clone/render path, B §5 #3).

This gate is owned here, executed via Claude Preview once both workstreams land.

## 5. Governance

- **ADRs:** A extends ADR-0035 (single=flat / collection=enveloped law + the
  enforcement gate). B writes a **new ADR** (templates adopt the two-phase
  pending→confirmed pattern; `storage_key` existence invariant; nullable column).
- **Design-system gate:** any new typed body (e.g. B's not-ready response) must be
  a typed struct, not `map[string]any` — `tools/cilint noresponsemap` enforces.
- **Test discipline (CLAUDE.md):** new tests use the canonical fixture framework;
  delete one-off scaffolding that breaks; repair only contract/invariant guards.
- **Wiki sync:** update `wiki/architecture/api-contract.md`, `wiki/modules/templates.md`,
  `wiki/modules/documents.md`, and the database docs via the wiki-curator after each
  workstream.

## 6. Adversarial-review disposition

Both specs were reviewed by three independent code-grounded reviewers (contract,
storage, whole-mission). All Critical/Major findings are folded into the amended
specs below. Decisions (flat/enveloped law, compile-time typed client, no frontend
zod, write-verified pointer) survived review unchanged; the amendments correct
scope, accuracy, and missing gates — not direction.
