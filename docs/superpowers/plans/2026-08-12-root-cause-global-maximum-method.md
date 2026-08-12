# Root-Cause / Global-Maximum Engineering Method Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the approved root-cause/global-maximum doctrine a single canonical engineering method, route all agent entrypoints to it, and remove dead skill routing without creating a new framework.

**Architecture:** `docs/engineering/root-cause-global-maximum-method.md` becomes the only normative definition. `AGENTS.md`, `CLAUDE.md`, `developing-new-work`, `adversarial-review`, and live onboarding/operating-system docs become thin contextual bridges. Do not restore the deleted `.agents/skills/` or retired `metaldocs-*` skill trees; commits `c7f06f2e` and `02ed1c24` removed them intentionally during the pre-v1 re-baseline.

**Tech Stack:** Markdown repository instructions, existing `.claude/skills`, MetalDocs wiki governance, existing verifier/governance scripts.

## Global Constraints

- Always simplify code, never correctness.
- Root cause precedes patch selection.
- Local maximums are legal only when explicitly transitional with a named successor and deletion condition.
- Prefer the strongest reasonable enforcement boundary; do not replace structural correctness with documentation or a weaker guard.
- YAGNI removes speculative capability and accidental complexity, not invariants or reachable-state proof.
- Global maximum does not authorize endless redesign; reopen settled decisions only on a material finding or changed constraint.
- `docs/engineering/root-cause-global-maximum-method.md` is the single canonical definition.
- `AGENTS.md` and `CLAUDE.md` must route to the canonical method rather than duplicate it.
- `developing-new-work` and `adversarial-review` retain their workflow-specific mechanics but consume the canonical definitions.
- Do not create a new verifier/framework solely to prove agents read the method.
- Do not restore `.agents/skills/` or retired `metaldocs-*` skills; those trees were deliberately removed.
- Historical snapshots, old plans, and module sync logs are evidence and are not rewritten merely to remove historical path references.
- If PR #128 (Solo-Strong CI v1) merges before execution, rebase this branch on the new `main` before Task 1 implementation and use the then-current verification commands.

---

### Task 1: Promote the approved design into the canonical engineering method

**Files:**
- Create: `docs/engineering/root-cause-global-maximum-method.md`
- Read: `docs/superpowers/specs/2026-08-12-root-cause-global-maximum-method-design.md`

**Interfaces:**
- Consumes: the operator-approved design spec.
- Produces: the canonical definitions and decision flow referenced by every later task.

- [ ] **Step 1: Create the canonical method from the approved normative sections**

Use the approved design as source text. The canonical document must contain, in this order:

```markdown
# Root-Cause / Global-Maximum Engineering Method

## Binding principle
> Always simplify the code, never simplify correctness. Find the root cause, test whether the proposed solution is only a local maximum, and prefer the global structure that makes the defect class unrepresentable or mechanically impossible at the strongest reasonable boundary.

## Definitions
Symptom; Root cause; Target property / invariant; Local maximum; Global maximum; Essential vs accidental complexity; YAGNI.

## Enforcement hierarchy
1. Structure / API makes the state unrepresentable
2. Type system makes the state invalid
3. Database/schema constraint makes the state invalid
4. Runtime boundary fails closed
5. Test proves reachable behavior
6. Lint/static guard detects the violation
7. Documentation/convention

## Required decision flow
Observe/reproduce -> Root cause -> Target property -> Authority/boundary -> Local/global candidates -> Legal outcome -> Proof -> Implement/simplify -> Evidence close-out.

## Legal outcomes
Restructure now | Transitional solution | Stop and split prerequisite | Current structure confirmed.

## Mandatory-use rules
Non-trivial bug fixes; repeated findings; architecture-bearing refactors; simplification of enforcement; new guard/lint/verifier/framework/platform primitive; cross-module work; auth/authz/tenant/transaction/idempotency/async/persistence/contract/schema changes; transitional designs; remediation/root-cause consolidation.

## Engineering Decision Record
Symptom; Root cause; Target property; Authority and boundary; Local-maximum candidate; Global-maximum candidate; Decision; Enforcement; Proof; Transitional exit.

## Guard policy
A custom guard is justified only when the property is material/reachable, stronger boundaries cannot reasonably express it yet, standard tooling is insufficient, and maintenance cost is lower than recurring risk.

## Transitional enforcement
Property now; why final structure cannot land yet; named successor; named deletion slice/milestone; deletion in successor definition-of-done.

## Review/convergence
Repeated same-altitude findings trigger structural analysis; optional hardening is not a blocker without a material property; stop when root cause/property/boundary/proofs are settled and no material contradiction remains.
```

Preserve the approved spec's precise definitions and examples; strip design-only sections about document placement, acceptance, and implementation architecture because those belong to the spec/plan, not the permanent doctrine.

- [ ] **Step 2: Verify the canonical document is standalone and contains no design placeholders**

Run:

```bash
git grep -n -E 'TBD|TODO|Proposed for operator review' -- docs/engineering/root-cause-global-maximum-method.md
```

Expected: no matches.

Run:

```bash
git diff --check
```

Expected: exit 0.

- [ ] **Step 3: Commit the canonical method**

```bash
git add docs/engineering/root-cause-global-maximum-method.md
git commit -m "docs(engineering): add root-cause global-maximum method"
```

---

### Task 2: Turn `AGENTS.md` and `CLAUDE.md` into truthful routing bridges

**Files:**
- Modify: `AGENTS.md`
- Modify: `CLAUDE.md`
- Read: `docs/engineering/root-cause-global-maximum-method.md`
- Read: `.claude/skills/developing-new-work/SKILL.md`
- Read: `.claude/skills/adversarial-review/SKILL.md`

**Interfaces:**
- Consumes: canonical method from Task 1 and actual current skill inventory.
- Produces: model-agnostic entrypoint rules with no dead repository-local skill links.

- [ ] **Step 1: Add a short mandatory engineering gate to `AGENTS.md`**

Add one routing section, not a duplicate doctrine:

```markdown
## Root-Cause / Global-Maximum Engineering Gate

For every non-trivial bug fix, refactor, architecture change, remediation, simplification, new abstraction, new guard, repeated review finding, or cross-boundary change, read `docs/engineering/root-cause-global-maximum-method.md` before implementation.

Before implementation, record at least: symptom, root cause, target property, authority/boundary, local-maximum candidate, global-maximum candidate, chosen outcome, enforcement layer, proof strategy, and transitional exit when applicable.

Do not optimize inside a known patch/workaround. Do not remove enforcement merely to reduce code. Do not add a guard when the invalid state can instead be made unrepresentable at a stronger reasonable boundary.
```

Place it before the detailed MetalDocs operating-system workflow so every boundary-specific task inherits it.

- [ ] **Step 2: Remove dead `metaldocs-*` / `.agents/skills/...` routing from `AGENTS.md`**

The deleted trees are not to be recreated. Replace phantom skill requirements with direct canonical docs plus the actual existing generic workflows:

```text
Backend/API -> wiki/architecture/backend-api-structure.md + api-contract.md + api-design-system.md
Database -> wiki/database/index.md + relevant database docs
Frontend -> wiki/architecture/frontend-structure.md
Frontend API/query -> frontend-structure.md query/API sections + generated API types
Module docs/wiki -> wiki/standards/documentation-governance.md + .claude/agents/wiki-curator.md when curator workflow is needed
New feature/module -> .claude/skills/developing-new-work/SKILL.md
Adversarial design/plan/diff review -> .claude/skills/adversarial-review/SKILL.md
Impact tracing when needed -> .claude/skills/gitnexus/SKILL.md
Harness coordination when needed -> .claude/skills/harness-hub/SKILL.md
```

Remove `runtime-contract-prereq` as a named local skill. Preserve its useful behavior as a direct rule: startup/auth/route/contract contradiction is a prerequisite hard stop; repair the owning boundary before returning to feature work.

- [ ] **Step 3: Collapse `CLAUDE.md`'s competing Global Maximum definition into a bridge**

Replace the long standalone definition with a short project bridge that keeps MetalDocs-specific force but delegates definitions:

```markdown
## Root Cause / Global Maximum

Canonical method: `docs/engineering/root-cause-global-maximum-method.md`.

Before non-trivial work, identify the root cause and target invariant before choosing a patch. Do not optimize inside a known workaround or local maximum. MetalDocs-specific invariants below remain binding constraints on every candidate solution.
```

Keep the existing MetalDocs system facts and non-negotiable invariants unchanged.

- [ ] **Step 4: Verify entrypoints contain no dead local skill routes**

Run:

```bash
git grep -nE '\.agents/skills/|metaldocs-(backend-api|frontend|tanstack-query|screen-integration-audit|screen-implementation|module-doc|module-doc-sync|database)|runtime-contract-prereq' -- AGENTS.md CLAUDE.md
```

Expected: no matches.

Verify all real skill paths exist:

```bash
test -f .claude/skills/developing-new-work/SKILL.md
test -f .claude/skills/adversarial-review/SKILL.md
test -f .claude/skills/gitnexus/SKILL.md
test -f .claude/skills/harness-hub/SKILL.md
```

Expected: exit 0.

- [ ] **Step 5: Commit routing entrypoints**

```bash
git add AGENTS.md CLAUDE.md
git commit -m "docs(agents): route engineering work through canonical method"
```

---

### Task 3: Make `developing-new-work` and `adversarial-review` consume the canonical doctrine

**Files:**
- Modify: `.claude/skills/developing-new-work/SKILL.md`
- Modify: `.claude/skills/adversarial-review/SKILL.md`

**Interfaces:**
- Consumes: canonical definitions from Task 1.
- Produces: workflow-specific skills with no competing definitions.

- [ ] **Step 1: Update `developing-new-work`**

Keep its pre-design system-impact gate, references, Green/Yellow/Red verdict, and MetalDocs invariant checklist. Replace standalone root-cause/global-maximum/YAGNI definitions with an explicit dependency:

```markdown
## Canonical engineering doctrine

Before this skill judges foundation or design direction, read `docs/engineering/root-cause-global-maximum-method.md`. That document owns the definitions of root cause, local/global maximum, YAGNI, enforcement hierarchy, transitional design, and legal outcomes. This skill only applies those definitions to MetalDocs pre-design system-impact analysis.
```

In the Foundation phase, require the decision record fields from the canonical method rather than introducing a second vocabulary.

- [ ] **Step 2: Update `adversarial-review`**

Keep review-specific mechanics: prior-finding disposition, weighted attack targets, repeated-finding trigger, two-patch ratchet, architecture checklist, convergence/stop conditions.

At the top add:

```markdown
## Canonical engineering doctrine

`docs/engineering/root-cause-global-maximum-method.md` owns root-cause, local/global maximum, YAGNI, enforcement, and transitional-solution semantics. This skill does not redefine them; it applies them during adversarial review.
```

Rewrite §1-§3 so they reference the canonical decision flow instead of restating full definitions. Preserve review-specific constraints such as “finding without root cause is incomplete”, repeated-finding structural signal, operator ownership of a redesign decision, and the subtractive review pass.

- [ ] **Step 3: Verify both skills reference the canonical method and do not route to deleted skill trees**

Run:

```bash
git grep -n 'docs/engineering/root-cause-global-maximum-method.md' -- .claude/skills/developing-new-work/SKILL.md .claude/skills/adversarial-review/SKILL.md
```

Expected: one or more matches in each file.

Run:

```bash
git grep -n '\.agents/skills/' -- .claude/skills/developing-new-work/SKILL.md .claude/skills/adversarial-review/SKILL.md
```

Expected: no matches.

- [ ] **Step 4: Commit skill convergence**

```bash
git add .claude/skills/developing-new-work/SKILL.md .claude/skills/adversarial-review/SKILL.md
git commit -m "docs(skills): consume canonical engineering doctrine"
```

---

### Task 4: Repair live onboarding/operating-system routing while preserving historical evidence

**Files:**
- Modify: `wiki/references/ai-operating-system.md`
- Modify: `wiki/ONBOARDING.md`
- Inspect and modify only if currently operational (not historical snapshot): `wiki/architecture/system-map.md`, `wiki/architecture/frontend-structure.md`, `wiki/modules/frontend/index.md`, `wiki/modules/frontend/auth.md`, `wiki/modules/frontend/documents.md`, `wiki/modules/frontend/approval.md`, `wiki/modules/frontend/controlled-documents.md`, `wiki/modules/frontend/iam.md`, `wiki/modules/frontend/templates.md`, `wiki/concepts/design-workflow-audit.md`
- Do not modify: `docs/superpowers/**` historical plans/specs merely for old path references; `wiki/**/_artifacts/**` sync logs/snapshots merely for old path references.

**Interfaces:**
- Consumes: truthful routing from Tasks 2-3.
- Produces: path-stable live docs that no longer tell humans/agents to open deleted skills.

- [ ] **Step 1: Rewrite `wiki/references/ai-operating-system.md` skill routing**

Replace the retired `metaldocs-*` list with the current model:

```text
Canonical engineering method -> docs/engineering/root-cause-global-maximum-method.md
New feature/module pre-design -> .claude/skills/developing-new-work/SKILL.md
Adversarial review -> .claude/skills/adversarial-review/SKILL.md
Boundary rules -> AGENTS.md + owning wiki architecture/module docs
Startup/contract contradiction -> prerequisite hard stop; repair owning boundary directly
Wiki sync -> documentation-governance.md + wiki curator workflow when needed
```

Update the document's `Last verified` date to `2026-08-12`.

- [ ] **Step 2: Remove deleted skill links from `wiki/ONBOARDING.md`**

Backend, frontend, database, and QA deep dives must point to the actual architecture/quality docs and `AGENTS.md`, not removed skill trees. Add the canonical engineering method to the initial contributor reading path after `AGENTS.md`/`CLAUDE.md` for non-trivial work.

- [ ] **Step 3: Sweep only live routing docs for remaining `.agents/skills/` references**

Run:

```bash
git grep -n '\.agents/skills/' -- wiki/ONBOARDING.md wiki/references wiki/architecture wiki/modules/frontend wiki/concepts
```

For each match:
- if the file is current operational guidance, replace the dead route with the truthful bridge from Tasks 2-3;
- if it is explicitly a frozen/historical audit artifact, leave it unchanged and record it in the PR handoff as historical evidence.

Do not touch `_artifacts/sync-log.md` or point-in-time `docs/superpowers/**` solely to make grep return zero.

- [ ] **Step 4: Verify every newly referenced path exists**

Run:

```bash
test -f docs/engineering/root-cause-global-maximum-method.md
test -f wiki/standards/documentation-governance.md
test -f .claude/skills/developing-new-work/SKILL.md
test -f .claude/skills/adversarial-review/SKILL.md
```

Expected: exit 0.

- [ ] **Step 5: Commit live routing cleanup**

```bash
git add wiki/ONBOARDING.md wiki/references/ai-operating-system.md wiki/architecture wiki/modules/frontend wiki/concepts
git commit -m "docs(wiki): remove dead agent skill routing"
```

Before committing, inspect `git diff --cached --name-only` and unstage any historical snapshot changed only because of a stale path reference.

---

### Task 5: Full self-review, repository validation, and draft PR

**Files:**
- Review all files changed by Tasks 1-4.
- Update: `docs/superpowers/specs/2026-08-12-root-cause-global-maximum-method-design.md` only if implementation discovered a material contradiction with the approved design; otherwise leave the approved spec immutable on this branch.

**Interfaces:**
- Consumes: completed documentation/routing implementation.
- Produces: a reviewable branch with evidence that the doctrine has one authority and all live bridges are truthful.

- [ ] **Step 1: Run a duplication/contradiction review**

Confirm:
- canonical definitions exist only in `docs/engineering/root-cause-global-maximum-method.md`;
- `AGENTS.md`/`CLAUDE.md` contain short bridges, not copied doctrine;
- skills contain workflow mechanics, not competing definitions;
- no new verifier or custom framework was added;
- no deleted skill tree was restored.

- [ ] **Step 2: Run routing checks**

```bash
git grep -n '\.agents/skills/' -- AGENTS.md CLAUDE.md wiki/ONBOARDING.md wiki/references/ai-operating-system.md .claude/skills/developing-new-work/SKILL.md .claude/skills/adversarial-review/SKILL.md
```

Expected: no matches.

```bash
git grep -n 'docs/engineering/root-cause-global-maximum-method.md' -- AGENTS.md CLAUDE.md wiki/references/ai-operating-system.md .claude/skills/developing-new-work/SKILL.md .claude/skills/adversarial-review/SKILL.md
```

Expected: each named bridge appears.

- [ ] **Step 3: Run repository documentation/governance verification**

Use the current commands after rebasing on latest `main`. At minimum:

```powershell
pwsh -NoProfile -File ./scripts/check-governance.ps1
```

and:

```bash
go run ./tools/verify --audit
```

If Solo-Strong CI has merged, also run the current changed-profile verification documented by that merge rather than preserving an obsolete command from this plan.

Expected: zero material findings caused by this change.

- [ ] **Step 4: Run textual quality checks**

```bash
git diff --check
git grep -n -E 'TBD|TODO' -- docs/engineering/root-cause-global-maximum-method.md docs/superpowers/plans/2026-08-12-root-cause-global-maximum-method.md
```

Expected: `git diff --check` clean; no placeholder matches in the canonical method.

- [ ] **Step 5: Commit any review-only corrections**

```bash
git add -A
git commit -m "docs(engineering): align doctrine routing and live guidance"
```

Skip this commit if the working tree is already clean.

- [ ] **Step 6: Open a draft PR and stop**

PR title:

```text
docs(engineering): canonicalize root-cause global-maximum method
```

PR body must report:
- approved spec path;
- canonical method path;
- removed dead routing and the two removal commits (`c7f06f2e`, `02ed1c24`) that justify not restoring it;
- files acting as bridges;
- live docs corrected vs historical references intentionally preserved;
- verification commands/results;
- no new verifier/framework;
- future MNFS promotion explicitly deferred.

Do not merge.
