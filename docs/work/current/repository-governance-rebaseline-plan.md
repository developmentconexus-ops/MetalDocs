# MetalDocs Repository Governance Rebaseline — Implementation Plan

> **For agentic workers:** execute inline, task-by-task, with verification before advancing.

**Goal:** Restore a local, selective-context repository operating model that keeps all prior MetalDocs Product/architecture authority reachable while reducing context and governance thrash.

**Architecture:** Keep the accepted Engineering and Frontend methods unchanged. Add one focused Repository Operating Method; make `AGENTS.md` the method/router bootstrap, `docs/index.md` the semantic task router, `docs/roadmap.md` a compact snapshot, and remove stale active claims from repository governance/provenance documents.

**Tech Stack:** Markdown governance, GitHub Actions YAML, Git/GitHub.

**Spec:** `docs/work/current/repository-governance-rebaseline-design.md`

## Global constraints

- Work only on `chore/repository-governance-rebaseline`, created from current `main`.
- Do not modify PR #173/B11 work.
- Do not change bytes of `engineering-method.md` or `frontend-product-experience-planning-method.md`.
- No branch/ref deletion in this increment.
- No Product/runtime implementation.
- Temporary `docs/work/**` files must be deleted before Ready.
- No merge without separate explicit operator authorization.

---

### Task 1 — Establish repository operating method

**Files:**
- Create `docs/development/repository-method.md`

- [ ] Write a concise method owning fresh-session recovery, authority hierarchy, selective context, documentation model, roadmap/decision index laws, Evidence/provenance, acceptance increments, review/Ready discipline, and Git/archive reachability.
- [ ] Explicitly state default `1–2` task owners, search relevant section first for large owners, and expand on material falsification rather than hard context caps.
- [ ] Explicitly separate method from MetalDocs-specific mechanics.
- [ ] Verify it does not reference external `ROUTER.md`, methodology pins, Product statuses, or B11 details.

### Task 2 — Rebuild bootstrap and semantic router

**Files:**
- Modify `AGENTS.md`
- Modify `docs/index.md`

- [ ] Rewrite `AGENTS.md` as compact bootstrap: state revalidation → roadmap → method selection → index → smallest owners → material expansion.
- [ ] Route engineering tasks to `engineering-method.md`, repository/Git/context tasks to `repository-method.md`, frontend planning to Engineering + Frontend methods.
- [ ] Preserve MetalDocs hard stops and explicit merge authorization.
- [ ] Rewrite `docs/index.md` around `Task / intention | Start with | Add when | Do not read by default`.
- [ ] Keep current Product/architecture/decision routes, but avoid requiring whole large owner files when a section/operation lookup is sufficient.
- [ ] Verify neither file owns mutable program status.

### Task 3 — Collapse repository-specific governance duplication

**Files:**
- Modify `docs/development/engineering-rules.md`
- Delete `docs/development/documentation.md` if all unique current meaning is absorbed by repository-method/engineering-rules

- [ ] Keep only MetalDocs implementation gate, temporary P8 Evidence rule, forward obligations, Git/CI/provenance specializations and verification entrypoint in `engineering-rules.md`.
- [ ] Remove external methodology/router wording and false current control claims.
- [ ] Absorb unique current documentation-placement/provenance rules into `repository-method.md` where they are general operating law or `engineering-rules.md` where MetalDocs-specific.
- [ ] Search live routers for references to `development/documentation.md`; update/remove them.
- [ ] Delete `documentation.md` only after no current consumer requires independent meaning.

### Task 4 — Reconcile current decision/provenance authorities

**Files:**
- Modify `docs/decisions/index.md`
- Modify `docs/decisions/repository-reset.md`

- [ ] Remove/replace the active `REPO-STD-V1` row so current repository operation points to local `repository-method.md` + `engineering-rules.md` rather than superseded external standard machinery.
- [ ] Keep the decision register compact; do not add governance chronology.
- [ ] In repository-reset, preserve exact archive/evidence refs and clean-slate decision, but remove claims that current CI SHA-checks those refs when it does not.
- [ ] State that current locators/refs provide reachability and are verified when a claim actually needs it, not globally on every CI run.
- [ ] Preserve secret-scanning reopen obligation and clean-slate implementation gate.

### Task 5 — Restore roadmap as snapshot

**Files:**
- Modify `docs/roadmap.md`

- [ ] Keep integrated `main` truth: B01–B10 integrated, B11/B12 not opened on main, implementation blocked.
- [ ] Add only a compact candidate note that repository governance is being rebaselined; do not copy B11/#173 history into this branch.
- [ ] Keep exact next action short and point to the governance PR lifecycle.
- [ ] Verify roadmap contains no review chronology or Evidence iteration journal.

### Task 6 — Align objective CI with the operating model

**Files:**
- Modify `.github/workflows/ci.yml`

- [ ] Add `docs/development/repository-method.md` to required operating files.
- [ ] Keep `required` as the only objective aggregate job.
- [ ] Preserve conflict-marker, implementation-block and `docs/work/**` merge-candidate protections.
- [ ] Do not add methodology-content grep, file-count budget, Evidence-SHA network checks, UX quality checks, or historical prose gates.
- [ ] Evaluate whether `fetch-depth: 0` is required by current script; if not, reduce checkout history while still ensuring the PR base SHA needed for `git diff BASE...HEAD` is available. If reliable base retrieval would add brittle shell/network machinery, defer CI checkout optimization to the later hygiene increment.

### Task 7 — Consistency and context proof

**Files:** all modified files above

- [ ] Verify exact hashes of Engineering and Frontend methods equal their `main` hashes.
- [ ] Search active bootstrap/governance surfaces for `ROUTER.md`, `conexus-methodology`, `REPOSITORY-STANDARD`, `ADVERSARIAL-REVIEW-METHOD` and classify any remaining hit as historical provenance or defect.
- [ ] Search for stale `development/documentation.md` links.
- [ ] Measure byte sizes of `AGENTS.md`, `docs/index.md`, `docs/roadmap.md` and compare with current main.
- [ ] Verify a frontend Access task can route from bootstrap to Engineering + Frontend methods + Access owners without opening whole Wire/Persistence by default.
- [ ] Verify `docs/work/**` remains allowed only while Draft.

### Task 8 — PR candidate cleanup

**Files:**
- Delete `docs/work/current/repository-governance-rebaseline-design.md`
- Delete `docs/work/current/repository-governance-rebaseline-plan.md`

- [ ] Create Draft PR against current `main` if not already open.
- [ ] Run/inspect `required` while Draft.
- [ ] Perform a final diff review for accidental Product/B11 changes.
- [ ] Delete the temporary spec/plan.
- [ ] Confirm merge-candidate diff contains only durable governance files and no `docs/work/**`.
- [ ] Mark Ready once, inspect final `required`, and stop for explicit operator merge authorization.

## Self-review

- Spec coverage: all approved governance surfaces are represented; branch hygiene/deletion and B11 rebaseline remain separate later increments.
- Placeholder scan: no TODO/TBD placeholders.
- Boundary check: no Product semantics or method-content rewrite is authorized.
- Critical risk: deleting `documentation.md` is conditional on proving no unique current consumer remains.
