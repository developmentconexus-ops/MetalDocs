# Current Agent Handoff

> **Last verified:** 2026-06-10
> **Scope:** Recovery context for a fresh agent session after the user reinstalls Claude/Codex. Captures workspace state, in-flight programs, the new architecture reference stack, and the agent-memory contents that do not survive a tool reinstall.
> **Out of scope:** Implementation details — follow the linked ADRs, backlog programs, and architecture docs.

## Why this exists (2026-06-10 rewrite)

The user is uninstalling Claude Code / Codex. The agent's persistent memory directory (`~/.claude/projects/.../memory/`) will be wiped, and session history will be gone. This file is the durable landing document. The previous version of this file (Plan 12 screen-finalization era, 2026-05-27) is superseded; Plan 12 context lives in git history if ever needed.

Start a fresh session by reading: `CLAUDE.md` → `wiki/README.md` → this file → the three-doc architecture stack (below).

## Workspace state at handoff

- Repository: `C:\Users\leandro.theodoro\Documents\MetalDocs`
- Branch: `qa/iam-area-membership` (main branch for PRs: `main`)
- Latest commit: `791cc0850 docs(wiki): sync route templates + query params to snake_case (Family 5)`
- **Uncommitted work in the tree (this session's deliverable — commit it):**
  - `wiki/standards/backend-canon.md` (new)
  - `wiki/architecture/backend-blueprint.md` (new)
  - `wiki/architecture/backend-target-architecture.md` (new)
  - `wiki/architecture/index.md`, `wiki/standards/index.md` (index entries)
  - `wiki/references/current-agent-handoff.md` (this rewrite)
  - Suggested commit: `docs(wiki): backend architecture reference stack (canon → blueprint → target) + handoff`

## The architecture reference stack (created 2026-06-10)

The user's directive: before further "industry-grade" refactoring, define everything — what a backend is composed of, how ours maps, and how it must behave. Three documents now form the canonical chain. **This is the reference for the whole refactoring effort:**

1. **[`wiki/standards/backend-canon.md`](../standards/backend-canon.md)** — implementation-independent definition of what a backend is composed of. Planes (data/control/management), layers, full concern catalog (edge, API, middleware, identity, domain/DDD, data, async, reliability, observability, security, config, integrations), exact identity vocabulary (authn vs authz, permission vs capability vs role vs scope, PEP/PDP, RBAC/ABAC/ReBAC), canonical sources, 10 litmus tests.
2. **[`wiki/architecture/backend-blueprint.md`](../architecture/backend-blueprint.md)** — current MetalDocs mapped onto the canon. Composition diagram, real middleware chain (`main.go:595-602`), async write path, per-concern maturity grades (✅/🟡/🔴), standards register, scoreboard with owning programs.
3. **[`wiki/architecture/backend-target-architecture.md`](../architecture/backend-target-architecture.md)** — **normative target spec.** ~60 REQ-* requirements (RFC 2119), target diagrams/workflows (topology, middleware chain, login, two-tier authz sequence, contract-first workflow, outbox/retry/DLQ), **Refactoring register RF-1..RF-10** with sequencing, litmus tests bound to REQ IDs as release gates, governance rules (PRs cite REQ IDs; MUST deviations need an ADR).

Flow: **canon defines target → blueprint locates gap → target specifies behavior → ADR records decision → program executes.**

## In-flight programs (status at handoff)

### 1. ADR 0022 — Authz capability coherence
`wiki/decisions/0022-authz-capability-coherence.md`. Phases 1-5 and 7-13 COMPLETE (registry SSOT, typed scope, membership area-scoping, directory SQL-scoping, CI binding guards, runtime scope binding, raw-string dialect closed, CI coherence-net revived). **Phase 6 (wiki sync via `wiki-curator`) still pending — deliberately sequenced LAST.** Residual: `authz-call-present` call-graph lint rewrite (deferred), authz-cache invalidation contract (RF-3). The original `area_admin` membership 403 is closed at root. **Never symptom-patch authz — this ADR is the boundary.**

### 2. API contract hardening
`wiki/backlog/api-contract-hardening.md`. 6 phases from the 2026-06-05 audit: A base-path ✅, B authz security gaps ✅, C dead-path prune ✅ (2026-06-07), D envelope / E hygiene / F cleanup — F partially done (FD-1 landed 2026-06-08). ~397 reported-only (non-blocking) api-lint spec-drift hits remain; blocking guards are at 0. Maps to RF-4 (fence/converge `spec2.yaml` + `internal/api/v2`) and RF-5.

### 3. Backend standardization (final polish)
6 families: auth hardening / pagination / error-vocab / casing / CI / dead-code. Executed as batched workflows on `qa/std-execution`, **merged into `qa/iam-area-membership`** (last session). Recent commits on this branch are its output: RFC 9457 error vocabulary (`ef696a177`), constant-time login paths (`f792d64de`), snake_case params spec-first (`16b72f81f`, `2369a02bf`), api-lint CI cleanup (`8e3b99727`), wiki sync (`791cc0850`). Batch A review findings addressed (`7faa7fb3d`). Status: Batches A+B landed; verify whether any family has open tail work before claiming the program closed.

## Next steps (agreed sequencing, from the target doc §10)

1. Commit the doc stack (above).
2. **Close ADR 0022 Phase 6** (wiki sync — runs last, now unblocked).
3. Finish api-contract-hardening tail (RF-5) + decide RF-4 (fence or converge the v2 spec surface).
4. **RF-1 + RF-2 as one program** — the biggest unowned gap: observability audit (OTel exporter wiring, trace propagation api→worker→docx-renderer through outbox rows, readiness-probe depth) + middleware chain realignment (panic recovery + trace context outermost, metrics outside auth so 401s/panics are visible, pre-auth IP-keyed rate limit on login). Current chain: `cors → origin → authn → iam → presence → obs → ratelimit → mux` (`apps/api/cmd/metaldocs-api/main.go:595-602`).
5. RF-3 (authz cache contract), RF-9 (graceful shutdown/timeout audit), RF-10 (idempotency coverage audit).
6. RF-7 (delete or implement empty `platform/cache`, `platform/storage`; document-or-fence `platform/messaging` servicebus), RF-8 (feature-flag lifecycle doc).

### Stage-1 backend audit — COMPLETE (2026-06-11)

Stage 1 (backend audit and mapping) is complete. The full atlas is at:

- **[`wiki/backend/index.md`](../backend/index.md)** — atlas root: binary map, repo topology, HTTP kernel, domain modules, platform packages, contract surface, flows, and navigation to all detail pages
- **[`wiki/backend/legacy-register.md`](../backend/legacy-register.md)** — complete legacy/duplication/smell register with RF-* cross-references and prioritized remediation sequencing

**Next step after the items above: Stage 2** — evaluate the as-mapped backend against professional SaaS and industry standards (Go conventions, PostgreSQL practice, ISO/RFC-grade API standards) using the atlas as the factual base.

## Recovered agent memory (wiped by the reinstall — re-create if a memory system returns)

These facts lived in the agent's persistent memory and are NOT derivable from the repo:

- **Machine:** C: SSD (XPG S50 Lite) has degraded writes (~7-16 MB/s; reads fine). Hardware/firmware issue. Prefer `D:` for heavy-write work (builds, caches) when possible.
- **Impeccable adoption plan:** adopt the "impeccable" design skill only AFTER IAM PR-2, as a guidance layer over MetalDocs design tokens (not a replacement); needs a bridge doc first.
- **Authz program intent:** the user explicitly chose root-cause coherence (ADR 0022) over symptom-patching the membership 403 — keep honoring that.
- **Backend standardization parameters:** approved as a 6-family batched-workflow program; H2 finding downgraded to LOW; NEW-1 was the real HIGH; idle timeout policy 30m sliding.
- **User working style:** wants industry-grade/professional standards, definitions-before-implementation, evidence-based closure; caveman (terse) chat mode was active in Claude sessions — irrelevant for docs/code, which stay normal.

## Operating rules that remain binding (pointers, not copies)

- `CLAUDE.md` — skill routing table, mandatory gates, hard-stop rule, evidence rule, close-out loop. Read first, always.
- `wiki/quality/qa-operating-system.md` — canonical QA/close-out policy.
- Startup: `.\scripts\start-api.ps1` only (script-truth policy). Dev login: `POST /api/v1/auth/login` `{"identifier":"admin","password":"AdminMetalDocs123!"}`.
- Wiki drift policy: code change touching a documented surface bumps `Last verified` same change; dispatch `wiki-curator` after refactors.
- **New since 2026-06-10:** backend work is reviewed against `backend-target-architecture.md` REQ IDs; deviating from a MUST requires an ADR.

## How to proceed in a fresh session

1. Read `CLAUDE.md` → `wiki/README.md` → this file → the three-doc stack.
2. `git status --short`; confirm branch; commit the doc stack if still uncommitted.
3. Pick up at "Next steps" item 2 (ADR 0022 Phase 6) unless the user redirects.
4. Do not re-execute anything from prior session summaries without verifying against the working tree — most of it is already done.
