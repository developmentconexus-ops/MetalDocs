# MetalDocs Development Harness

**Ratified:** 2026-07-10 (operator-designed with Fable; binding for all ROADMAP.md units).
**Purpose:** one repeatable execution machine per roadmap unit — senior-level output, no AI slop,
no context loss, no unverified closure. The harness does not replace the skills/gates that exist
(`developing-new-work`, MNFS workflow validation, TDD superpowers); it fixes WHO runs them,
in WHAT order, with WHAT evidence.

---

## 1. Model matrix — who does what

| Role | Model | Does | Never does |
|---|---|---|---|
| **Design/synthesis** | Fable (rare, expensive) | Architecture decisions, system-impact gates, spec ratification support, roadmap surgery | Implementation, bulk file reading, worker duty |
| **Orchestrator** | Standing dispatch session (operator's hub) | Builds per-unit context pack → creates spawn_task chip (self-contained prompt: files, constraints, done-criteria, budget) → operator clicks to launch → acceptance-reviews returned evidence → integrates → next chip. Chips ALWAYS (operator-ratified 2026-07-10): every unit runs as its own observable session the operator can open and follow | Writing large code diffs itself when a worker fits; reviewing its own dispatched work |
| **Unit session** | **Opus** (operator selects at launch; medium effort default) | Runs one unit end-to-end per its chip prompt: spec→plan→TDD→verify→evidence→commit | Scope beyond its chip; pushing; flipping statuses past HS-1 |
| **Implement worker** | Sonnet subagent (inside the unit session) | One TDD slice per dispatch, per written plan task | Scope beyond its task card; touching files outside its slice |
| **Reviewer** | Sonnet subagent (independent — never the implementer) | Two-stage review per slice (see §4) | Generating new scope (anti-circle: reviews VERIFY, they don't renegotiate the plan) |
| **Mechanical** | Haiku | Renames, comment sweeps, format-preserving tweaks | Anything requiring judgment |
| **Browser QA persona** | Fresh session via `spawn_task` | §6 protocol — validates as a USER, from zero context | Reading implementation diffs first (would bias); fixing anything |

Concurrency ≤15 workers. Fable never a worker.

## 2. Unit execution loop (every ROADMAP unit)

**Dispatch mechanics (operator-ratified 2026-07-10):** every unit dispatches as a spawn_task
CHIP — a real standalone session the operator can open, watch turn-by-turn, and intervene in.
The orchestrator hub authors the chip's self-contained context pack (exact files, constraints,
done-criteria, budget); the operator launches it on Opus. On completion the hub runs per-unit
acceptance review (accept / reject / block; 2× reject = redesign, new chip). Rejection = new
corrective chip with the findings. Shared-resource rule: one owner of the :80 stack at a time.
Note: chip sessions run in fresh worktrees — untracked runtime files (.env, .env.local) must be
copied from the main checkout (never printed). Chip prompts must state this.

**Chip prompt template MUST carry, verbatim-strength:** (a) the §4 subagent obligations 1–6 as
numbered instructions, not a passing mention; (b) "read docs/superpowers/HARNESS.md §4–§5 before
implementing"; (c) the evidence.md dispatch-ledger requirement — the hub rejects closures whose
evidence lists no implementer/reviewer dispatches; (d) the task-board obligation — create the
native task board from the plan before any dispatch, keep statuses live; (e) the hub comms
contract below, including the hub's `HUB_SESSION_ID` so the chip can address it from turn one.

**Hub–chip comms protocol (operator-ratified 2026-07-11).** Chips are not fire-and-forget: every
chip session talks to the hub over `mcp__ccd_session_mgmt__send_message` (target =
`HUB_SESSION_ID` from the chip prompt; fallback: `list_sessions` and match the hub by cwd = repo
root + hub title). Messages are structured events, not chat — one message per event, header line
first:

| Event | When | Payload (keep caveman-brief) |
|---|---|---|
| `CLOSED` | unit done, commits in place | unit id · branch + commit range · gates run + results · defers · HS-1 items. **Mandatory** — the closure report travels as a message, never sits waiting in the chip transcript. |
| `BLOCKED` | stop-rule hit (architecture contradiction, broken prerequisite, budget ceiling) | what blocked, evidence, smallest unblock ask. Send IMMEDIATELY — do not burn budget waiting or patching around it. |
| `ESCALATION` | out-of-scope defect found (pre-existing debt, product defect) | classification + reproduction pointer; keep fixing in-scope work meanwhile if possible. |
| `REQUEST` | needs a shared resource | e.g. :80 stack rebuild/restart, DB reseed — the hub owns shared infra; chips never take it. |
| `ACK` | hub sent a mid-flight correction | one line: what was applied / why not applicable. |

Hub side of the contract: `CLOSED` triggers the acceptance pipeline (independent read-only
reviewer → verdict → merge or remediation) without operator relaying. **Post-merge cleanup is
part of acceptance:** once the merge lands and the post-merge ladder is green, the hub removes
the chip's worktree (`git worktree remove`, `--force` for detached-HEAD leftovers; verify no
running session occupies the path via `list_sessions` first) and safe-deletes the chip branch
(`git branch -d` — never `-D`; a branch that refuses deletion carries unmerged work and stays
until the operator decides). Worktrees are execution scratch, not archives — evidence lives in
merged docs, not in leftover checkouts. `BLOCKED`/`ESCALATION` get
a decision or an operator escalation, not silence; ACCEPT-WITH-CONDITIONS / REJECT findings go
back **to the same chip session** via `send_message` — context is intact there, so remediation is
a message, not a new chip (a new corrective chip only on 2× reject = redesign). Reviewer
subagents dispatched by the hub are **git read-only** (diff/show/log; never checkout/apply/stash
— an acceptance reviewer once contaminated the main index via checkout). Known friction, accepted:
hub→chip `send_message` prompts the operator to confirm each send; chip→hub arrives on the hub's
next turn. That keeps the operator in the loop by design.

**Hub bootstrap:** any fresh session becomes the hub via the `harness-hub` skill
(`.claude/skills/harness-hub/SKILL.md`) — it reconstructs hub state from repo truth (this file,
ROADMAP, git, task board, live chip sessions, :80 stack health) and runs the loop below. One hub
at a time; a new hub first confirms the old hub session is not mid-merge.

**Hub instrumentation:** the hub mirrors the ROADMAP queue as its own native task board
(one task per unit, blockedBy = the roadmap's ordering locks), may read a chip session's
transcript mid-flight (`list_events`, operator-approved) for audit, may inject corrections into
a running chip (`send_message`, operator-confirmed), and push-notifies the operator when a unit
returns for acceptance or an HS-1 gate is waiting.

```
ROADMAP.md → take topmost actionable unit → open ONLY its listed context files
  ↓
P0 GATE      developing-new-work (new feature/module only; Red = STOP)
P1 SPEC      brainstorm within locked constraints → design doc → operator ratifies
P2 PLAN      writing-plans → slice-level task cards (each: goal, files, failing-test-first,
             done-criteria, authz/module checklist §5 pre-answered)
P3 IMPLEMENT subagent-driven TDD per slice (§4) — commit per green slice
P4 VERIFY    ladder L0→L2 (§5) run by orchestrator from clean state
P5 QA        fresh-session browser QA persona (§6) — milestone-close units only... 
             feature units: targeted UI walk of the changed surface
P6 JUDGE     milestone validation gate (milestone end only) — separation of powers. Cold,
             independent judge: `mnfs-workflow:milestone-validate` skill (dispatches the
             read-only milestone-reviewer crew). Legacy programs mid-flight (M2d, M5) whose
             worktrees still carry the old milestone-validator agent may finish on it.
P7 CLOSE     evidence.md (commands+outcomes+review disposition+bounded defers+REQ IDs)
             → commit → CLOSED message to hub (comms protocol above) → hub acceptance pipeline
             → HS-1 operator gate where the program requires it → update ROADMAP row
```

Stop rules (inherited, restated): architecture contradiction mid-implementation = AS-2 stop, not
adaptation. Prerequisite broken (startup/auth/contract drift) = repair first, feature work halts.
Budget ceiling hit = stop, flush state to evidence.md/ROADMAP, split the unit.

## 3. Planning standards (P1–P2)

- **Global maximum first.** Every plan names the foundation it builds on and judges it (sound vs
  patch). Optimizing inside a patch is a defect. If the right structure crosses the unit boundary —
  surface, don't absorb.

### GM fork procedure (binding when ≥2 implementations compete)

Trigger: at spec time OR mid-implementation, two viable shapes exist — typically A = improve
in-place on the current base (local maximum) vs B = the right structure, but it changes the base.
Never pick silently; run this:

1. **Judge the base first.** Is the current implementation sound, or legacy/patch/workaround?
   (Sound base → A and B are both legitimate candidates; judge on merit. Bad base → step 2 rules.)
2. **Write the fork record** (in the unit spec or evidence.md — 5 lines, not an essay): the
   options, which one is the global maximum and why, cost of each, and whether B crosses the unit
   boundary (module ownership, contract, DB shape, budget).
3. **Route:**
   - B fits inside the unit boundary + budget → **B, always.** "A is faster" never beats
     structure; speed is not a tiebreaker against the global maximum.
   - B crosses the boundary → STOP (AS-1/AS-2 shape) and send `BLOCKED` to the hub with the fork
     record. Operator/hub picks: (a) expand the unit to B; (b) file B as a named ROADMAP unit and
     land a **bridge slice** on the current base under the no-deepening rule; (c) accept A as
     named debt with an ADR-recorded reopen trigger (§ "Long-term over expedient").
4. **No-deepening rule.** A bridge on a base marked-for-replacement must be minimal and
   reversible: it may touch the bad base, it may NOT grow it — no new capabilities, abstractions,
   or consumers added to the structure that B will delete.
5. **Reversibility test (the veto).** If choosing A makes B materially more expensive later
   (new callers of the wrong seam, data written in the wrong shape, contract surface that must be
   broken), A is FORBIDDEN at chip discretion — operator sign-off only.
6. **Named debt or it didn't happen.** Any accepted local maximum gets a debt entry (ROADMAP row
   or evidence defers) naming the reopen trigger. An unnamed local max discovered at review =
   reject.
- **Whole-system orientation.** Plan states owning module(s), non-owning modules, cross-module
  edges (direction + published interface), and which of the 6 invariants are touched — before any
  task card is written. (This is the `developing-new-work` §1/§3 output carried into the plan.)
- **YAGNI-professional.** Build exactly the ratified scope with production structure: no
  speculative abstraction, no config for imagined futures, no "v2 hooks". BUT never below the
  platform floor — TxRunner, problem+json, authz seams, outbox, testdb are the floor, not
  gold-plating. Extension points only where a NAMED next unit needs them (e.g. route-builder-v2
  leaves the ActorSelector slot because M4 is on the roadmap — that's a named consumer, not YAGNI
  violation).
- **Long-term over expedient.** No workaround lands without an ADR-recorded reopen trigger. "Works
  but wrong layer" = rejected at review, not deferred.
- Slices are vertical (contract→service→repo→test green per slice), PR-sized, independently
  revertable.

## 4. Implementation standards (P3) — the anti-slop contract

**Subagent dispatch is MANDATORY, not advisory (amended 2026-07-10 — sessions were observed
implementing everything inline).** Unit-session obligations, checkable:

1. Code slices are implemented by a dispatched **sonnet** subagent (Agent tool / Task), one slice
   per dispatch. The unit session (Opus) writes code inline ONLY for trivial glue (≤ ~10 lines,
   no new behavior) — anything more inline is a harness violation to note in evidence.md.
2. Every slice gets an **independent reviewer subagent** (sonnet — e.g. cavecrew-reviewer, or the
   frontend reviewer agents for FE) BEFORE the next slice starts. The session may NEVER
   self-review as substitute — implementer ≠ reviewer is the non-negotiable, regardless of cost.
3. Mechanical work (renames, comment sweeps, format-only) → **haiku** subagent.
4. Bulk reading/inventory → sonnet investigator subagent returning compressed report; the unit
   session does not tree-crawl.
5. evidence.md MUST list dispatches: per slice — implementer agent, reviewer agent(s), verdicts.
   A closure with zero dispatches listed fails acceptance review at the hub.
6. **The unit session IS a milestone orchestrator** (amended 2026-07-10): its FIRST act after
   reading the plan is to create a native task board (TaskCreate) — one task per feature/slice,
   `blockedBy` edges for real dependencies — and keep it live: in_progress at dispatch,
   completed only at reviewed-green. The board is the session's execution truth; evidence.md
   snapshots its final state. The hub keeps its own board of UNITS (chips) the same way —
   two levels, same discipline: hub tracks units, unit tracks slices.

Every slice: **failing test first** (canonical framework — testdb for DB integration), implement to
green, then two-stage review before the next slice starts:

1. **Stage-1 code review** (independent sonnet / cavecrew-reviewer): correctness, idiom match with
   the surrounding module, slop checklist below.
2. **Stage-2 domain review**: the §5 authz/module checklist, invariant fit, contract parity.

**AI-slop checklist — any hit = REJECT the slice:**
- Speculative abstraction / unused params / interfaces with one impl and no named second consumer
- Comment narration ("// now we check X"), PR-voice comments, restating-the-code comments
- Defensive noise: blanket recover/try-catch, nil-checks on invariants the type system holds,
  fallback values on integrity-critical reads (no-fallback principle — fail closed)
- Idiom mismatch: new pattern where the module already has one (error wrapping style, port naming,
  test table shape — copy the neighborhood)
- Dead code kept "just in case", commented-out blocks, TODO landmines without a roadmap/ADR anchor
- Hand-rolled platform equivalents (frameworks-catalog is binding: TxRunner, problem.Write, outbox,
  strictjson, testdb…)
- Generated-file edits (`api.gen.go` etc.) — contract-first or nothing
- Test theater: tests that assert the mock, not the behavior; missing negative/cross-tenant case

**Reviews verify, never generate scope.** A reviewer wanting new scope files it as a finding for
the operator queue; it does not grow the slice.

## 5. Verification ladder (P4) — every level from clean state

| Level | What | Gate |
|---|---|---|
| L0 | `go build ./...` · `tsc/typecheck` · lints incl. api-lint, module-boundaries.yml, check-test-discipline, capability-coherence | zero findings |
| L1 | `go test ./...` + integration suite via `.\scripts\test-integration.ps1` (NEVER hand-set `DATABASE_URL` — the script derives it from `.env` `POSTGRES_*`, probes postgres, and fails loud; `testdb.Open` silently SKIPS when the var is missing or the DB is down, which reads as false green) · FE vitest | green; flaky = fix or delete per legacy-test rule |
| L2 | Full container stack via gateway :80 (coded compose path, [[docker-deploy-methodology]]) · gateway smoke: logins, target routes, RFC 9457 shapes | green, evidence captured |
| L3 | Browser QA persona (§6) | GREEN verdict artifact |
| L4 | milestone validation gate (mnfs-workflow:milestone-validate; legacy in-flight milestones may close on the old milestone-validator) | PASS verdict written to the milestone's qa/ artifact |

**Backend non-negotiables re-checked at L0–L2 for every touched endpoint** (the miss-nothing list):
tier-1 route→capability wired · tier-2 `authz.Require` in-tx after `SeedTxIdentity` · tripwire arm
exists · tenant predicate on every query, cross-tenant = 404 · problem+json errors · outbox for any
side effect · idempotency where the contract requires it · H-PRE-1 (no authz-recording read inside
lock-holding tx) · OpenAPI ↔ generated ↔ handler parity.

## 6. Fresh-session browser QA protocol (P5)

**Trigger:** every milestone close (mandatory); feature units get a targeted version of the same
protocol on the changed surface. Curl-only = FAIL (M2c lesson: F8's curl QA was false-green; the
stage_kind violation was only caught by rendered-UI operator QA).

**Mechanics:** `spawn_task` a FRESH session (zero inherited context — that's the point: it can't
trust anything the implementer believed). Session must have browser tooling; if the spawned env
can't render (F-UI-1 class), the task reports CANNOT immediately — never fakes a pass from curl.

**The spawned prompt carries (self-contained):**
1. Persona + mission: "You are a user-level QA operator. You did not build this. Validate by
   driving the rendered UI through the gateway on :80. Trust nothing you cannot see rendered."
2. Binding checklist: `wiki/quality/qa-operating-system.md` + the relevant
   `wiki/quality/*-checklist.md` for the surface class.
3. The milestone's acceptance script (from milestone.md): concrete user journeys, BOTH route
   shapes where approval is involved (review+approval AND approval-only), negative paths
   (SoD block, wrong-stage action, cross-tenant 404), all relevant personas (author, reviewer,
   approver, delegate, observer, admin).
4. Stack bring-up pointer: coded compose path + seed, `.\scripts\start-api.ps1` never bash.
5. Evidence contract: per-step screenshot/network/console capture; verdict file
   `<milestone>/qa/browser-qa-<date>.md` with GREEN/RED per journey; RED = exact repro, no fix
   attempts (separation of powers — QA judges, orchestrator fixes, re-run FULL ladder after fix).

## 7. Context & handoff discipline

- Fresh Opus session per milestone; fresh spawn per QA. Handoff state lives ONLY in: ROADMAP.md
  row, unit evidence.md, milestone.md status, memory index. If it isn't written there, it didn't
  happen.
- Sessions open only the unit's listed context files; bulk reads → sonnet inventory agent
  returning compressed report.
- Self-compact at 200k main-session tokens: flush durable state first.
- Every session's first act on program work: read ROADMAP.md; last act: update the unit row
  (status + actuals vs budget).

## 8. Failure protocol

- RED at any ladder level: fix at root cause (systematic-debugging skill), then re-run the FULL
  ladder from L0 — no partial re-verification of "just the failing bit".
- Same slice fails review twice: escalate to orchestrator redesign of the slice, not a third patch.
- Contradiction with an invariant/ADR: STOP, write the finding, operator decides (AS-1/AS-2 shape).
- Any stop-rule hit inside a chip session: send `BLOCKED` (or `ESCALATION` for out-of-scope
  defects) to the hub immediately per the §2 comms protocol — never wait silently in the chip
  transcript for the operator to notice.
- Never: skip hooks, force-push, fake evidence, widen scope to "fix while here" (spawn_task chip
  for out-of-scope findings instead).
