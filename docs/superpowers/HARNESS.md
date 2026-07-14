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
| **Orchestrator** | Standing dispatch session (operator's hub) | Builds per-unit context pack → dispatches the unit AGENT (Transport A, §2: Agent tool, Opus, worktree isolation, background — zero operator clicks) → answers its events → acceptance-reviews returned evidence → integrates → auto-advances to the next actionable unit. Transport B (spawn_task chip the operator launches) only on the fallback triggers in §2 | Writing large code diffs itself when a worker fits; reviewing its own dispatched work |
| **Unit agent/session** | **Opus** (hub sets `model: opus` on dispatch; Transport B: operator selects at launch, medium effort) | Runs one unit end-to-end per its dispatch prompt: spec→plan→TDD→verify→evidence→commit; dispatches its own nested sonnet/haiku workers (verified capability 2026-07-12) | Scope beyond its prompt; pushing; flipping statuses past HS-1; touching shared infra (:80 stack) |
| **Implement worker** | Sonnet subagent (inside the unit session) | One TDD slice per dispatch, per written plan task | Scope beyond its task card; touching files outside its slice |
| **Reviewer** | Sonnet subagent (independent — never the implementer) | Two-stage review per slice (see §4) | Generating new scope (anti-circle: reviews VERIFY, they don't renegotiate the plan) |
| **Mechanical** | Haiku | Renames, comment sweeps, format-preserving tweaks | Anything requiring judgment |
| **Browser QA persona** | Fresh background QA agent (zero inherited context, browser tools); `spawn_task` fallback if the agent env cannot render (F-UI-1 class) | §6 protocol — validates as a USER, from zero context | Reading implementation diffs first (would bias); fixing anything |

Concurrency ≤15 workers. Fable never a worker.

## 2. Unit execution loop (every ROADMAP unit)

**Dispatch mechanics v2 (2026-07-12) — DEFAULT REVERTED 2026-07-13 (operator-directed:
"vou querer retomar como estávamos antes — spawn_task e mensagem entre sessões eu preferia").
Transport B (chip) is the DEFAULT again; the operator accepts launch/approval clicks in
exchange for turn-by-turn observability of unit sessions. Transport A stays available but only
on explicit operator request per unit.** Two transports, same logic, same event grammar:

**Transport A — unit agent (opt-in only, fully autonomous).** The hub dispatches the unit as a
background subagent: `Agent(subagent_type: general-purpose, model: opus, isolation: worktree,
run_in_background: true)` with the same self-contained context pack a chip would get. Zero
operator clicks: no launch click, no message-approval clicks (agent events are turn-returns;
hub replies via the native `SendMessage` tool — both verified prompt-free 2026-07-12).
Mechanics facts (probe-verified 2026-07-12):
- Worktree is auto-created at `.claude/worktrees/agent-<agentId>` on branch
  `worktree-agent-<agentId>`; the path is DETERMINISTIC from the agentId in the spawn result —
  immediately after spawning, the hub copies `.env` (+`.env.local` if present) from the main
  checkout into that worktree (file copy only, contents never printed/read). The agent verifies
  `.env` exists before any DB/integration work; missing = `REQUEST env` event, never hand-set
  connection strings.
- `worktree.baseRef = "head"` in `.claude/settings.json` is MANDATORY — the default ("fresh")
  forks from `origin/main`, which is hundreds of commits stale here. `symlinkDirectories` shares
  `node_modules` from the main checkout: unit agents NEVER run `pnpm install`/`npm install`
  inside a symlinked node_modules (dep changes = `REQUEST` to the hub).
- Unchanged worktrees auto-clean on completion; changed ones persist for the hub to merge.
- The unit agent CAN dispatch nested sonnet/haiku workers (§4 obligations apply unreduced) —
  but ONLY synchronously (`run_in_background: false`). Confirmed mechanics (3.1a, 2026-07-13):
  a BACKGROUND nested child's completion routes to the HUB session's main loop, never to the
  parent agent — the parent stops "waiting" and deadlocks until the hub pokes it. Not a race;
  deterministic misrouting. Hub side: a child-agent notification arriving at the hub = relay
  its result to the parent via SendMessage immediately (parent must not re-implement). Unit-
  agent turns end ONLY with a §2 event, never with "waiting on my worker" narration.
- Crash model: background agents die with the hub session. Recovery = repo truth: commit per
  green slice on the worktree branch is mandatory, and hub boot (harness-hub skill) scans
  `worktree-agent-*` branches + `.claude/worktrees/agent-*` for orphaned in-flight work.

**Transport B — spawn_task chip (DEFAULT since 2026-07-13).** A real standalone session the
operator launches and can watch turn-by-turn. All chip rules from
the 2026-07-10/11 ratifications stay binding for B: operator launches on Opus; `.env` copy note
in the prompt; comms over `mcp__ccd_session_mgmt__send_message` to `HUB_SESSION_ID` (real
registry id verified via a previously-sent message's `from="local_…"` attribute, never the
scratchpad UUID; title-match fallback embedded) — accepting the confirmed client limitation that
the operator approves each send (Ctrl+Enter), messages few and batched.

**Advance (revised 2026-07-13, chip default):** when a unit's acceptance completes green (merge
+ ladder + rebuild + smoke + board + cleanup), the hub PREPARES the next actionable ROADMAP
unit as a spawn_task chip and surfaces it — the operator launches it (Opus). No agent
auto-dispatch. Queue holds unchanged: (a) `OPERATOR-GATE` rows; (b) ordering lock or pending
HS-1; (c) operator "pause dispatch". The operator retains: unit launches, HS-1 gates,
ratifications (AskUserQuestion), push decisions, pause/resume.

**Unit prompt template MUST carry, verbatim-strength (both transports):** (a) the §4 subagent
obligations 1–6 as numbered instructions, not a passing mention; (b) "read
docs/superpowers/HARNESS.md §4–§5 before implementing"; (c) the evidence.md dispatch-ledger
requirement — the hub rejects closures whose evidence lists no implementer/reviewer dispatches;
(d) the task-board obligation — create the native task board from the plan before any dispatch,
keep statuses live; (e) the comms contract below — Transport A: "end your turn with exactly ONE
event header when you have something for the hub; the hub's reply arrives as your next message";
Transport B: the `HUB_SESSION_ID` + fallback.

**Hub–unit comms protocol (event grammar identical on both transports).** Units are not
fire-and-forget. Transport A: an event = the agent ENDS ITS TURN with the event as its final
message (this pauses the unit — no budget burns past a block); the hub answers via `SendMessage`
(context resumes intact). Transport B: an event = one `mcp__ccd_session_mgmt__send_message`.
One event per message, header line first:

| Event | When | Payload (keep caveman-brief) |
|---|---|---|
| `CLOSED` | unit done, commits in place | unit id · branch + commit range · gates run + results · defers · HS-1 items. **Mandatory** — the closure report travels as the event itself, never sits waiting in a transcript. |
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
back **to the same unit** (Transport A: `SendMessage` to the agent, context resumes intact;
Transport B: `send_message` to the chip session) — remediation is a message, not a new unit
(a new corrective dispatch only on 2× reject = redesign). Reviewer
subagents dispatched by the hub are **git read-only** (diff/show/log; never checkout/apply/stash
— an acceptance reviewer once contaminated the main index via checkout). **Dual-reviewer parity
(2026-07-14):** hub acceptance runs the independent Claude cavecrew-reviewer AND an independent
GPT-5.6 SOL pass (Codex rescue, subagent fan-out, git-read-only) on the same diff; merge only when
both clear, disagreements reconciled in the acceptance note (§4 obligation 2 has the full rule). Transport B comms
caveat (confirmed client limitation 2026-07-12): the desktop client hard-enforces per-send
confirmation on `mcp__ccd_session_mgmt__send_message` regardless of permission mode — the
operator approves each with one click; never reroute comms through side-channel files to dodge
it; messages few and batched. Transport A has no such prompt — which is why it is the default.

**Hub bootstrap:** any fresh session becomes the hub via the `harness-hub` skill
(`.claude/skills/harness-hub/SKILL.md`) — it reconstructs hub state from repo truth (this file,
ROADMAP, git incl. orphaned `worktree-agent-*` branches, task board, live chip sessions, :80
stack health) and runs the loop below. One hub at a time; a new hub first confirms the old hub
session is not mid-merge.

**Hub instrumentation:** the hub mirrors the ROADMAP queue as its own native task board
(one task per unit, blockedBy = the roadmap's ordering locks), may inspect a running unit
agent's transcript (`TaskOutput`) or a chip's (`list_events`, operator-approved) for audit,
may inject mid-flight corrections (`SendMessage` to agents; `send_message` to chips), and
push-notifies the operator when an HS-1 gate or ratification is waiting on them.

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
   **Dual-reviewer parity at the final/CLOSED gate (ratified 2026-07-14, operator: "claude and
   gpt on equal terms"):** before a unit emits `CLOSED`, its final diff is reviewed by BOTH an
   independent Claude reviewer (cavecrew-reviewer) AND an independent **GPT-5.6 SOL** reviewer via
   the Codex rescue path (`codex:codex-rescue` agent / `codex:rescue` skill). The GPT reviewer is
   instructed to **use subagents** to fan out over the changed files/modules (not a single-pass
   skim) and is **git-read-only** (diff/show/log/`git show <rev>:<path>` — never checkout/edit/
   stash). Both reviewers are independent of the implementer and of each other. Gate: `CLOSED`
   only after BOTH clear; on disagreement, BOTH verdicts + the unit's reconciliation travel IN
   the `CLOSED` event — GPT findings are never silently dropped. evidence.md records both
   dispositions. (Per-slice mid-unit review stays single Claude reviewer for cost; parity applies
   at the final gate.)
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
| L1 | `go test ./...` + integration via `.\scripts\test-integration.ps1` (NEVER hand-set `DATABASE_URL` — the script derives it from `.env` `POSTGRES_*`, probes postgres, and fails loud; `testdb.Open` silently SKIPS when the var is missing or the DB is down, which reads as false green) · FE vitest. **Selective-integration policy (2026-07-13):** default = touched packages' integration tests + the cross-cutting guard suites (`./tests/integration/tenantdata/...`, `./tests/integration/scenarios/...`, `./tests/integration/iam/...`); FULL `./...` integration only when `db/migrations`, `db/baseline`, or `internal/platform/**` are touched — otherwise the full ~50-min sweep is waste. Do NOT set `-count=1` blanket: the Go test cache is sound here (testdb's schema fingerprint reads every migration file, and the cache tracks files read — schema changes auto-invalidate); use `-count=1` only when re-proving a specific flake. | green; flaky = fix or delete per legacy-test rule |
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

**Mechanics:** dispatch a FRESH QA persona with zero inherited context — that's the point: it
can't trust anything the implementer believed. Default: background QA agent (browser tools
verified available to agents 2026-07-12); the hub keeps its own browser pane hands-off while QA
runs (one pane per session — serialize). Fallback: `spawn_task` a fresh session. Either way the
persona must have working browser tooling; if its env can't render (F-UI-1 class), it reports
CANNOT immediately — never fakes a pass from curl.

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

**Login/personas rule (operator-ratified 2026-07-11):** QA state and identities are REPO
KNOWLEDGE, not operator favors. Before asking the operator for ANYTHING, QA reads
`wiki/references/local-dev-startup.md` (seeded personas table: identifiers, passwords, roles —
dev-only seeds, documented in-repo, NOT secrets) and `wiki/quality/qa-operating-system.md`, and
checks for seed fixtures that already bake the needed state (e.g. `e2e_seed.go` bakes the
reviewer-also-approver overlap). Login is scripted with the seeded dev credentials (API login →
session cookie, or driving the login form with the DOCUMENTED seed password). The old
"QA personas never type passwords" rule applies to REAL credentials only — seeded dev personas
are test fixtures. Asking the operator to type a password that is written in the wiki = FAIL of
QA self-sufficiency; operator-manual steps are a last resort for genuinely non-seeded state.

## 7. Context & handoff discipline

- Fresh Opus unit agent (or chip) per unit; fresh QA persona per QA. Handoff state lives ONLY in:
  ROADMAP.md row, unit evidence.md, milestone.md status, memory index. If it isn't written there,
  it didn't happen. Commit per green slice is the crash-recovery contract for Transport A.
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
- Any stop-rule hit inside a unit: emit `BLOCKED` (or `ESCALATION` for out-of-scope defects) to
  the hub immediately per the §2 comms protocol — Transport A: end the turn with the event now;
  Transport B: send the message now. Never wait silently for someone to notice.
- Never: skip hooks, force-push, fake evidence, widen scope to "fix while here" (out-of-scope
  findings go to the hub as `ESCALATION` for their own future unit dispatch instead).

## 9. Parallel-track execution (ratified 2026-07-14 — operator: finish faster without losing quality)

The "one unit at a time, top-to-bottom" rule of ROADMAP §0 was a *serialization by decree*, not by
dependency. It is now relaxed: **the hub runs independent units as concurrent tracks.** Order is
governed by the real dependency DAG, not by row number.

**The only true parallelism axis is hub-dispatched sibling units** — one worktree/branch/session
each. Sibling units report to the hub *by design* (that is the correct routing), so the nested-child
misrouting bug (§2) does NOT apply between them. Genuine wall-clock concurrency lives here.
Intra-unit nested workers stay SYNC (`run_in_background: false`) — used for help, never for
wall-clock concurrency. A unit that discovers internal parallelism does NOT background its own
children; it emits `SPLIT-REQUEST` (below) so the hub forks a sibling unit on the real axis.

### 9.1 When the hub may parallelize

Before dispatching, the hub builds a **collision matrix** across candidate units and runs two
concurrently ONLY when they are disjoint on ALL of:

| Axis | Collision test | Rule if they collide |
|---|---|---|
| **Files** | same source file edited by both | serialize, or split ownership so each owns disjoint files |
| **FE surface** | same `features/<domain>` component tree | serialize the FE-touching pair (merge hell otherwise) |
| **Contract** | both edit `api/openapi` | serialize contract edits (one owns the spec at a time) |
| **Migration** | both add a migration | pre-allocate disjoint numbers up front (§9.2) |
| **DB shape** | both alter the same table / same tenant-port registration | one owns it exclusively; the other `REQUEST`s |
| **Module** | both mutate the same module's domain/service internals | serialize; cross-module edges are fine (different modules = safe) |

Disjoint on all six → **parallel-safe**, dispatch as concurrent tracks. Any collision → serialize
that pair (the rest can still parallelize). Test-infra units (harness/testdb) and mechanical
sweeps are almost always orthogonal to feature units — parallelize them freely.

### 9.2 Coordination protocol (the guardrails that keep quality)

1. **Migration number pre-allocation.** The hub reserves a disjoint block per track BEFORE
   dispatch and writes it into each unit's prompt (e.g. "you own 0306+"). A track needing an
   unplanned migration emits `REQUEST migration-number` — never grabs the next free number blind
   (two blind grabs = duplicate number = merge break).
2. **Exclusive-owner files.** When two tracks must touch the same shared file (e.g.
   `TenantDataPort` registration), the hub names ONE owner in the prompts; the other emits
   `REQUEST` and the hub serializes that edit or relays a patch. No silent concurrent edits to a
   shared file.
3. **Contract lock.** Only one in-flight track may edit `api/openapi`. The hub does not dispatch a
   second contract-touching track until the first merges (or assigns disjoint path sections and
   accepts a regen conflict it resolves at merge).
4. **Same base, serial merge-back.** Every track branches off the *same* current main. Merge-back
   is serialized through the ONE acceptance gate (independent read-only review → merge --no-ff →
   post-merge ladder → board/ROADMAP → cleanup). **Merge order = smallest/lowest-risk first**
   (mechanical → test-infra → feature) to minimize rebase pain.
5. **Rebase-on-merged-main.** After track N merges, any still-in-flight track that shares moved
   infra (migrations, platform, a shared file) is told to rebase on the new main before its own
   merge. The hub decides per collision matrix whether a rebase is needed; a purely disjoint track
   needs none.
6. **Baseline unchanged.** Zero-new-RED gate holds per track. A track's post-merge ladder runs on
   the integrated main, not just its branch — a green branch that reddens main = REJECT, remediate.

### 9.3 New event: `SPLIT-REQUEST` (intra-milestone / intra-unit autonomy)

A unit or milestone agent that finds its own work is internally parallel (independent slices on
disjoint files/modules) does NOT background nested children. It emits:

| Event | When | Payload |
|---|---|---|
| `SPLIT-REQUEST` | unit's remaining slices are mutually independent and splitting buys real wall-clock | proposed split: slice groups + their disjoint file/module/migration ownership + which stay in this unit vs fork to a sibling |

Hub side: validate the split against the §9.1 collision matrix; if clean, fork the forked group as
a **sibling unit** (own worktree/chip, pre-allocated migration block, exclusive-owner assignments)
and let the requester keep the rest. If the split collides, deny with the reason (requester
proceeds serially). This is how a milestone gets "autonomia de paralelizar o que conseguir e
reorganizar se precisar" — the milestone owns the *proposal*; the hub owns the *collision
adjudication* (so two forks never corrupt each other). Reorganization of a unit's OWN slice order
needs no request — the agent reslices freely (P2 authority); only a FORK to a concurrent sibling
routes through `SPLIT-REQUEST`.

### 9.4 Observability under Transport B

Each track = one session the operator launches and watches (chip default). Degree of concurrency =
operator's attention budget, set per batch (the hub asks when it is non-obvious). The hub prepares
all parallel-safe chips at once and surfaces them together, each carrying its collision-matrix
ownership (migration block, exclusive files, contract-lock status) so a launched track cannot
stray into another's lane.
