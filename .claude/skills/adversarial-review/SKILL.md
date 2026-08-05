---
name: adversarial-review
description: >-
  Run when dispatching Codex (or any independent model) to adversarially review a MetalDocs
  design, plan, or diff — and when disposing of what it returns. Turns "ask Codex to review
  this" into a bounded protocol: root cause before patch, Local Maximum vs Global Maximum,
  /simplify, module abstraction + published exposure standard, YAGNI, and a binding
  architecture checklist. Produces a per-round disposition ledger and a convergence verdict.
  Use it for every gate review, every plan review, and every review round after the first.
  Do NOT use it for reading code, for a first-pass self-check, or as a substitute for the
  developing-new-work pre-design gate.
---

# Adversarial Review — the review doctrine, not the transport

The transport is `harness:codex-dispatch` (model, effort, OS-process, stdin closed, tee +
`-o`). **That skill decides how to launch. This skill decides what to ask and what to do with
the answer.** They compose: dispatch is mechanical, review is doctrinal.

This exists because an unbounded reviewer reviews *do jeito dele* — it finds the nearest
defect, proposes the nearest patch, and the author applies it. Round after round the artifact
accretes patches, each individually correct, and the structure that generated every one of
them is never touched. **A review loop with no root-cause rule is a patch generator.**

## The three failure modes this skill exists to stop

| Failure | What it looks like | Rule that stops it |
|---|---|---|
| **Patch-on-patch** | Every finding closed at the location it was reported | §1 Root cause first |
| **Local-maximum ratification** | The reviewer improves *inside* a structure it never questioned | §2 Local vs Global Maximum |
| **Deference** | "The existing implementation presumably has a reason" | §5 Never accept what is implemented |

---

## §1 — Root cause before patch (the ordering rule)

**Binding order for every finding, no exceptions:**

1. **What is the root cause?** Not where the symptom is — what structural fact makes this
   symptom *possible*. State it in one sentence.
2. **What would make this finding impossible?** If the answer is a different structure, name
   the structure. If it is a bound/guard/test, name that.
3. **Only then**, propose the fix.

A finding that reports a symptom without step 1 is **incomplete** and is sent back, not
applied. A fix applied without step 2 is a patch, and it is labelled as one.

**The repeated-finding signal.** Three findings in one area, or three bounds on one guard,
is a *structural* signal, not a quality signal. It means the guard is in the wrong place.
Stop bounding and go to §2.

**Ratchet:** the author may not apply more than **two** consecutive patches to the same
construct without §2 being run on it explicitly and its verdict written down.

---

## §2 — Local Maximum vs Global Maximum (CLAUDE.md, made operational)

CLAUDE.md: *"If the current implementation is legacy, a patch, or a workaround, do NOT
optimize inside it — that locks in a local maximum."* Under review pressure this is the rule
most easily skipped, because a patch always closes the finding and a restructure never does.

**The four questions, answered in writing, whenever §1 step 2 names a structure:**

1. **What is the global-maximum structure?** Name it concretely — "a published `documents`
   write port", not "better separation of concerns". If you cannot name it, you have not
   found it.
2. **What does a proven system do here?** The house rule is *what would a senior engineer or
   an existing validated system do*. Cite one if you can.
3. **What does the global maximum cost, and what does the local maximum cost later?** Both
   numbers. A local maximum with an unstated future cost is how it becomes permanent.
4. **Which is chosen, by whom?** This is an **operator decision**, not the agent's and not the
   reviewer's.

**The three legal outcomes.** There is no fourth.

| Outcome | Requirement |
|---|---|
| **(a) Restructure now** | The global maximum is built in this change. |
| **(b) Local maximum, labelled, successor scheduled** | The artifact says *explicitly transitional*, names the global-maximum structure, names the milestone that lands it, and states that the local maximum is **deleted** by that milestone — as part of its definition-of-done, not as a follow-up. |
| **(c) Stop and surface** | The better answer crosses this change's boundary. Do not patch around it. |

**An unlabelled local maximum is a defect.** It is indistinguishable from an intended design
by every future reader, and that is exactly how it becomes permanent.

---

## §3 — /simplify and YAGNI (the subtractive pass)

Every round runs a pass whose only question is: **what can be deleted?** Additive review is
the default disposition of both models and humans; this pass is the counterweight.

**Delete-first checklist:**
- **Two paths doing one thing** → one path. (Catalog §9.)
- **A field, flag, or column that has one producer and one consumer** → inline it.
- **A compatibility layer, shim, or "for now" adapter** → house rule: *tudo fallback legacy é
  extermínio*. Migrating a consumer means DROPPING the old wire/DTO field, never relaxing it
  to optional for compatibility.
- **An abstraction with exactly one implementation and no second one planned in this
  milestone** → YAGNI. Delete the interface, keep the concrete type.
- **A configuration knob nobody sets** → delete the knob and the branch behind it.
- **A vocabulary entry whose copy is identical to another entry's** → the distinction is not
  real; delete one.

**YAGNI has a boundary.** It applies to *speculative capability*. It does **not** apply to:
- an invariant's enforcement point,
- a fail-closed default,
- a test for a state that is reachable today.

Deleting one of those is not YAGNI. It is removing a guard, and it must be argued on its own.

---

## §4 — Module abstraction and the exposure standard

Modules talk to each other. What a module *exposes* to be talked to is a published contract,
and what is behind it is nobody else's business. The review must check both directions.

**The exposure standard — what a module publishes, and nothing else:**

| Layer | Published? | Rule |
|---|---|---|
| Application service methods | **Yes** | The primary and preferred entry point. |
| Published Go interfaces (ports) | **Yes** | The seam a consumer depends on. |
| Domain value objects crossing the seam | **Yes** | Closed types; no anemic maps or `any`. |
| Domain internals, aggregates, invariant logic | **No** | — |
| Repositories | **Never** | — |
| SQL / tables / triggers | **Never** | — |

**Review questions, every round:**
- Does any consumer reach past a module's application service into its repository, SQL, or
  domain internals? Each occurrence is a boundary violation, not a shortcut. *(In this repo
  that is the CLAUDE.md cross-module rule; a foreign module writing an owner's table is the
  canonical instance — and it is what forces the owner's guards to reason about who is
  writing and why, a question the owner should be answering.)*
- Is a consumer redeclaring a concept the owner already publishes? (Catalog §10.)
- Does the module have a **discoverable** published surface — can a consumer answer "what do I
  call, what do I consult?" from one place, without reading the module's internals? If the
  answer is "read the code and find out", the module has no exposure standard, and that is
  the finding.
- Is a primitive that two modules now need still private to one of them? Promotion to
  platform is triggered by the **second** consumer, in that PR. (Catalog §11.)

---

## §5 — Adversarial posture: never accept what is implemented

**The reviewer's null hypothesis is that the artifact is wrong.** "Looks reasonable" is not a
disposition. Neither is "the existing code presumably had a reason".

- **Existing implementation is evidence of use, not of correctness.** Shipped, compiling, and
  unreported means it was never exercised under the caller set now being added.
- **The author's stated rationale is a claim to falsify**, not context to defer to.
- **Every load-bearing claim about existing code must carry a `file:line`.** No anchor → it is
  an open question, not a premise. This is the single highest-yield rule in the protocol
  (defect-class catalog §27, §29).
- **Anchor verification runs FIRST**, before any design critique. It is mechanical, cheap, and
  it kills the plausible-but-false premise class outright.

**Symmetric duty on the author.** Verify every finding against source before accepting it.
A reviewer is confidently wrong by exactly the same mechanism the author is. Sustained,
evidenced pushback is part of the protocol — a finding you disproved in code is closed by
disproving it, not by complying with it.

---

## §6 — The architecture checklist (binding, every round)

Run this list against the artifact. It is not a summary of the review — it is the floor
beneath it. Any **No** is a finding.

**Boundaries**
- [ ] Owning module named for every change; no consumer reaches past an application service.
- [ ] No module redeclares another's published concept; second consumer ⇒ platform promotion.
- [ ] Every new cross-module edge has a direction and a published port.

**Invariants** *(the six non-negotiables — CLAUDE.md; see `developing-new-work`)*
- [ ] AuthZ reasoned in **capabilities**, never roles; two-tier PDP intact; tier-1 is not
      treated as a floor.
- [ ] Request lifecycle unchanged — no route reinvents auth, validation, or errors.
- [ ] Contract-first: routes change only via `api/openapi` + `oapi-codegen`.
- [ ] Multi-tenant: `tenant_id` on every tenant table, tx-local GUCs, cross-tenant → 404.
- [ ] Async: no network call shares a tx with a state write; consumers idempotent.
- [ ] DB enforces the invariant; the app check is the friendly first line, not the guard.

**Guards and backstops**
- [ ] Every backstop answers: **what does this trust, and who can set it?** A DB backstop
      re-derives; it never accepts an attribution claim carried in the row. (Catalog §30.)
- [ ] Every read-then-write is one statement, or it is a race.
- [ ] Fail-closed defaults: a guard and the data that arms it land in the **same** commit.
- [ ] No guarded branch that no environment executes. (Catalog §22.)

**Executability**
- [ ] Tasks walked as a **dependency** graph, not a topic list.
- [ ] Every task ends green (build, vet incl. `-tags integration`, its declared tests).
- [ ] For each commit boundary: *what is broken between these two commits, and for how long?*

**Truth and vocabulary**
- [ ] Every identifier in plan code is defined in the plan or cited `file:line`.
- [ ] No hand-synced enumeration introduced; shared vocabulary has one registry.
- [ ] Generated artifacts regenerated, not hand-edited; parity gate named.
- [ ] Wire vocabulary: no two codes with identical meaning; deleting a code lists its full
      fanout (Go, generated FE JSON, FE copy, wiki).

**Contradiction** *(separate search strategy — do not fold into the pass above)*
- [ ] Every imperative and prohibition extracted to a flat list and checked for conflicts.
- [ ] No task violates the Global Constraints section.

---

## §7 — Round protocol

**Two jobs per round, in this order. Never one.**

- **Job 1 — disposition of prior findings.** One line each:
  `N. CLOSED | PARTIAL | OPEN — file:line — one sentence.`
  Strict: **mentioning a problem is not closing it.** Most carry-overs in practice are fixes
  that acknowledged a finding without resolving it — which is precisely why this job is
  separate and comes first.
- **Job 2 — attack the new material only.** Do not re-litigate closed findings. Weight the
  prompt toward what *this revision* introduced: new SQL, new triggers, new task ordering,
  new type shapes, new capabilities.

**Prompt shape** (write to file; `harness:codex-dispatch` for the launch):

```
READ-ONLY. Do not edit, create or delete any file.
Round N. Round N-1 returned <counts>, VERDICT: <verdict>. The artifact has been revised.
Verify every claim against actual code; cite file:line.

## Read
<the artifact, the spec, anything the artifact claims about>

## Job 1 — are your round N-1 findings closed?
<the numbered list, verbatim>

## Job 2 — what did THIS revision break or miss?
<5-8 weighted attack targets, each a concrete question with a named file>

## Output
Job 1 dispositions, then Job 2 numbered, most severe first:
    SEVERITY (BLOCKER|MAJOR|MINOR) — one-line claim — file:line evidence — what must change.
Then exactly one closing line:
    VERDICT: PROCEED | PROCEED WITH FIXES | DO NOT PROCEED — <biggest remaining risk>
No praise, no summary, no restating the artifact. Findings only.
```

**Weighted attack targets are the highest-leverage part of the prompt.** A generic "review
this" returns generic findings. Name the specific new construct and ask the specific question
you are least sure of.

---

## §8 — Convergence: when to stop

Track two things per round: **count** and **altitude**.

| Round | Findings | Altitude |
|---|---|---|
| 1–2 | many | design: wrong structure, missing authorization, lost evidence |
| 3 | fewer | fixes that named the problem without closing it |
| 4 | fewer still | mechanical: non-existent APIs, wrong status codes |

**Converging** = count falling **and** altitude dropping toward things a compiler or lint
would catch. That means the design-level search is exhausted — stop, and let the remaining
class be caught by the build.

**Not converging** = findings stay at the same altitude round after round. The loop is
generating scope, not finding defects. **Stop it** and escalate to §2: same-altitude
recurrence means the structure is wrong, and no number of rounds will fix that.

**Stop conditions — any one is sufficient:**
- Both sides agree, and the verdict is PROCEED / PROCEED WITH FIXES with no BLOCKER.
- Findings are all rung-1–3 mechanical (a compiler, generator, or lint would catch them).
- Same-altitude recurrence → §2, operator decision, loop paused.

**Never** close a loop on a DO NOT PROCEED verdict by declaring the remaining risk acceptable
without writing down who accepted it.

---

## §9 — Output of the review, per round

Written to the work item's directory, not to chat:

- **Disposition ledger** — Job 1 lines, verbatim.
- **Findings** — Job 2 lines, verbatim.
- **Author disposition per finding** — `applied` / `disproved (file:line)` / `deferred
  (owner + milestone)`. A finding is never silently dropped.
- **§2 verdicts** — every local-vs-global judgment, with outcome (a)/(b)/(c) and, for (b),
  the successor milestone.
- **Convergence line** — count + altitude + stop/continue.

The dispatch artifacts (`.log`, `.last.md`) are the evidence the ledger points at.

---

## Relation to other skills

| Skill | Boundary |
|---|---|
| `harness:codex-dispatch` | **Transport.** Model, effort, path, ceremony. Always used to launch. |
| `developing-new-work` | **Pre-design gate.** Runs before design exists. This skill runs after. |
| `superpowers:writing-plans` | Produces the artifact this skill attacks. |
| `docs/engineering/defect-class-catalog.md` | Part II + Appendix C are the empirical base for §5 and §6. |
