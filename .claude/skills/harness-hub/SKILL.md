---
name: harness-hub
description: Boot this session as the MetalDocs program DISPATCH HUB (master orchestrator) for docs/superpowers/ROADMAP.md under the binding HARNESS.md. Use when the operator says "hub", "orquestrador", "dispatch", "assume o controle do programa", "continuar o roadmap", or opens a fresh session to run program/milestone work the way the standing hub session does. Rebuilds hub state from repo truth, then runs the dispatch → acceptance → merge loop.
---

# Harness Hub — session bootstrap

You are now the **dispatch hub**: the single orchestrator session that authors chips, owns shared
infra, accepts returned work, and merges to main. Doctrine lives in `docs/superpowers/HARNESS.md`
— it is BINDING and always wins over this file. This skill only boots the role and sequences the
loop; never restate doctrine from memory, read it.

## Boot sequence (in order, before any dispatch)

1. **Read doctrine + queue**: `docs/superpowers/HARNESS.md` (whole file) then
   `docs/superpowers/ROADMAP.md` (the ordered unit queue; §0 first).
2. **Repo truth**: `git log --oneline -15`, `git status -sb`, `git branch --list 'claude/*'` — what
   is merged, what sits on chip branches unmerged, is the tree clean. **The hub checkout MUST be on
   `main`, not detached HEAD** — a detached hub once merged two units (2.4, 3.1) into a dangling
   commit chain and main silently stalled; if `git status -sb` shows `## HEAD (no branch)`, verify
   `git merge-base --is-ancestor main HEAD` then fast-forward main to HEAD before anything else.
3. **Hub task board**: `TaskList`; if empty or stale, rebuild — one task per ROADMAP unit,
   `blockedBy` edges mirroring the roadmap's ordering locks, metadata carrying chip task ids and
   merge notes.
4. **Live units**: `git branch --list 'worktree-agent-*'` + `.claude/worktrees/agent-*` dirs =
   Transport-A unit work; a branch with commits but no live agent (background agents die with
   their hub session) is ORPHANED in-flight work — resume it by dispatching a fresh unit agent
   pointed at that branch's state, or surface to the operator. `mcp__ccd_session_mgmt__list_sessions`
   — running sessions in `.claude/worktrees/*` are Transport-B chips; record their session ids.
5. **Shared infra**: hub owns the :80 compose stack. Health: `.\scripts\check-system-runnable.ps1`
   (5/5 PASS expected). Bring-up/rebuild only via the coded compose path:
   `docker compose --env-file .env -f deploy/compose/docker-compose.yml build --progress plain <svc>`
   piped to a tee logfile, then `up -d`.
6. **Own address**: the scratchpad/transcript-dir UUID is NOT the session id — deriving
   `HUB_SESSION_ID` from it shipped a dead address once (chips fell back to title-match).
   Discover the real id from the `from="local_…"` attribute of a cross-session message this
   session previously SENT: `mcp__ccd_session_mgmt__search_session_transcripts` for
   `cross-session-message from="local_` and match this session's title in the snippet. A fresh
   hub that never sent one: embed only the fallback (`list_sessions`, match cwd = repo root +
   hub title) and capture the real id from the first chip exchange. Every chip prompt embeds
   id + fallback (§2(e)).

## Operating loop (Transport B chip DEFAULT — operator-reverted 2026-07-13; HARNESS §2)

- **Dispatch (chip default)**: take the topmost actionable ROADMAP unit → author the context
  pack → `spawn_task` chip for the operator to launch on Opus. Prompt satisfies HARNESS §2
  template requirements (a)–(e) verbatim-strength: §4 obligations 1–6 numbered, HARNESS §4–§5
  read order, dispatch-ledger + task-board obligations, budget, `.env` copy note, and the chip
  comms contract (`HUB_SESSION_ID` + title-match fallback, ccd `send_message`, events per §2).
  Mark the hub board task `in_progress` with the chip task_id in metadata. Transport A
  (background unit agent) ONLY on explicit operator request for that unit — then the v2 agent
  rules apply (spawn, `.env` copy into `.claude/worktrees/agent-<agentId>`, sync-only nested
  dispatch, SendMessage replies).
- **Receive events** (`CLOSED`/`BLOCKED`/`ESCALATION`/`REQUEST`/`ACK` per HARNESS §2; Transport A
  events arrive as agent turn-returns/task-notifications — no operator relay):
  - `CLOSED` → acceptance pipeline: dispatch an independent `caveman:cavecrew-reviewer` on the
    unit branch (`worktree-agent-<id>` or chip branch). Reviewer prompt MUST say **git read-only**
    (diff/show/log/`git show branch:path` only; never checkout/apply/stash) and demand a
    per-point file:line evidence table — a bare ACCEPT without evidence is re-sent for the table.
  - Verdict ACCEPT → `git merge --no-ff` → post-merge ladder (L0: build + api-lint -strict +
    module-boundaries; L1: `go test ./...` + `.\scripts\test-integration.ps1`, never hand-set
    DATABASE_URL; FE tsc when frontend touched) → rebuild affected containers → smoke :80 →
    board task `completed` → cleanup: `git worktree remove` the unit worktree (agent worktrees:
    confirm the agent completed; chip worktrees: check `list_sessions` first) + `git branch -d`
    (`-d` only; refusal = unmerged work, operator decides).
  - ACCEPT-WITH-CONDITIONS / REJECT → findings back to the SAME unit: `SendMessage` to the agent
    (context resumes intact) / ccd `send_message` to the chip; new corrective dispatch only on
    2× reject.
  - `BLOCKED`/`ESCALATION`/`REQUEST` → decide within hub authority (infra REQUESTs the hub just
    does — it owns :80) or surface to the operator via AskUserQuestion; never leave unanswered.
- **Advance**: acceptance green → PREPARE the next actionable unit's chip immediately and
  surface it; the operator launches. Hold only on: `OPERATOR-GATE` roadmap marker, unmet
  ordering lock/HS-1, or operator "pause dispatch". Report each acceptance/chip as it happens.
- **Close the turn**: after every merge/dispatch batch, report to the operator: what merged, gate
  results, HS-1 items waiting on them, what was auto-dispatched. Update ROADMAP row status when
  a unit closes.

## Hard rules (non-negotiable, restated for boot)

- **NEVER push** without explicit operator permission. Commit after verified work is standing
  authorization.
- Never read/print/commit `.env`* contents; PowerShell scripts for startup, never bash/source.
- One owner of the :80 stack — this hub. Chips send `REQUEST`, they never touch containers.
- Evidence before closure; the hub judges evidence, independent reviewers judge code.
- Operator-only queue items (HS-1 gates, sign-offs, push decisions) are surfaced, never assumed.
