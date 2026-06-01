# Onboarding — Day 1

> **Last verified:** 2026-06-01 (docgen-v2 → docx-renderer rename)
> **For:** Any engineer new to the codebase.
> **Goal:** Get the system running locally, understand what's where, and ship your first PR by end of day 2.

If anything below is wrong or out of date, fix it in the same PR as your first change. This doc is treated as code.

---

## 0. Read these three things first (10 min)

| Doc | Why |
|---|---|
| [`README.md`](../README.md) (project root) | Elevator pitch + canonical commands |
| [`CLAUDE.md`](../CLAUDE.md) (project root) | Operating rules every contributor (human or agent) follows — script-truth policy, mandatory gates, hard-stop rule |
| [`wiki/diagrams/c4-context.md`](diagrams/c4-context.md) | What MetalDocs is from outside |

If you only have 30 seconds: MetalDocs is a multi-tenant SaaS for ISO 9001 controlled documents. Templates → drafts → approval → frozen artifact → PDF. Go API + React SPA + Postgres + MinIO + a Node side-service (docx-renderer) for server-side docx render. See the diagrams referenced below.

---

## 1. Get it running (≈15 min)

**Prerequisites**

- Windows + PowerShell (canonical), or POSIX with bash
- Docker Desktop running (Postgres, Redis, MinIO, Gotenberg, docx-renderer run in containers)
- Go 1.22+, Node 20+, `corepack` enabled (`corepack enable`)

**Bring up infra + API + worker + frontend**

```powershell
# from repo root
.\scripts\start-api.ps1            # canonical API startup on :8081 (rebuilds if needed)
```

That script is the **only supported entrypoint** — see [`wiki/references/local-dev-startup.md`](references/local-dev-startup.md). Do not invoke `go run` directly or `source .env`.

**Frontend dev server**

```powershell
cd frontend/apps/web
corepack pnpm install
corepack pnpm dev                  # vite on :4173 (or :4174 if 4173 busy)
```

**Smoke-test login (dev creds)**

```
POST http://localhost:8081/api/v1/auth/login
{"identifier":"admin","password":"AdminMetalDocs123!"}
```

Browser: open `http://localhost:4173` and log in with the same. If you see the inbox, you're up.

---

## 2. The 30-minute mental model

Read these in order. Total: about 30 minutes of reading.

1. [`wiki/diagrams/c4-context.md`](diagrams/c4-context.md) — system boundary, who uses it.
2. [`wiki/diagrams/c4-container-backend.md`](diagrams/c4-container-backend.md) — the moving parts inside.
3. [`wiki/architecture/system-overview.md`](architecture/system-overview.md) — ports, services, topology.
4. [`wiki/diagrams/sequence-create-document.md`](diagrams/sequence-create-document.md) — simplest end-to-end flow.
5. [`wiki/diagrams/sequence-edit-autosave.md`](diagrams/sequence-edit-autosave.md) — the scalability pattern (browser ↔ MinIO direct).
6. [`wiki/diagrams/sequence-signoff-freeze.md`](diagrams/sequence-signoff-freeze.md) — the compliance moment, **plus a known design issue** (sync HTTP in transaction; planned async refactor inside the doc).
7. [`wiki/diagrams/sequence-pdf-export.md`](diagrams/sequence-pdf-export.md) — async derivation via transactional outbox.

After this you'll be able to follow any module-level conversation.

---

## 3. Role-based deep dives

Pick the path that matches what you're about to work on. Each ends with "now you can ship a small PR".

### Backend engineer

1. [`wiki/architecture/backend-api-structure.md`](architecture/backend-api-structure.md) — module layout (`internal/modules/<name>/{domain,application,delivery,repository}`).
2. [`wiki/architecture/api-contract.md`](architecture/api-contract.md) + [`api-design-system.md`](architecture/api-design-system.md) — OpenAPI is source of truth; `oapi-codegen` generates handlers.
3. [`wiki/concepts/authz-tiers.md`](concepts/authz-tiers.md) — two-tier authz + Postgres tripwire (this surprises everyone the first time).
4. [`wiki/modules/<area-you-touch>.md`](modules/) — pick one (auth / templates / documents / approval / controlleddocuments / iam / taxonomy / render).
5. Skill: [`metaldocs-backend-api`](../.agents/skills/metaldocs-backend-api/SKILL.md) — the route truth-table workflow.

**First PR idea:** add a new field to an existing read endpoint. Touches OpenAPI + handler + repository + tests.

### Frontend engineer

1. [`wiki/architecture/frontend-structure.md`](architecture/frontend-structure.md) — feature-sliced layout, `createBrowserRouter`, **no** HashRouter, **no** legacy `src/api/`.
2. Browse `frontend/apps/web/src/features/<domain>/` — each feature has `pages/`, `components/`, `hooks/`, `api/`, `queries/`, `routes.tsx`.
3. [`wiki/modules/editor-ui-eigenpal.md`](modules/editor-ui-eigenpal.md) — the eigenpal Anti-Corruption Layer (your seam for a future SuperDoc swap).
4. Skill: [`metaldocs-frontend`](../.agents/skills/metaldocs-frontend/SKILL.md) + [`metaldocs-tanstack-query`](../.agents/skills/metaldocs-tanstack-query/SKILL.md).

**First PR idea:** add a column to an existing list screen. Touches generated FE types + a query hook + a component + tests.

### Approval / workflow engineer

1. [`wiki/diagrams/sequence-signoff-freeze.md`](diagrams/sequence-signoff-freeze.md) — both the current and the planned async design.
2. [`wiki/modules/approval.md`](modules/approval.md) + [`wiki/modules/approval-tech-debt.md`](modules/approval-tech-debt.md).
3. [`wiki/concepts/freeze-and-hashing.md`](concepts/freeze-and-hashing.md).
4. [ADR 0009 — PDF Dispatch Outbox](decisions/0009-pdf-dispatch-outbox.md) — the async-outbox pattern (model for freeze refactor).

### Database engineer

1. [`wiki/database/overview.md`](database/overview.md) + [`migration-policy.md`](database/migration-policy.md) + [`reference-data.md`](database/reference-data.md).
2. [`wiki/database/relationships.md`](database/relationships.md) — the relational graph.
3. Skill: [`metaldocs-database`](../.agents/skills/metaldocs-database/SKILL.md).

### QA engineer

1. [`wiki/quality/qa-operating-system.md`](quality/qa-operating-system.md) — the 5 truths, 7 gates, evidence rule, hard-stop rule.
2. [`wiki/quality/screen-qa-checklist.md`](quality/screen-qa-checklist.md) and friends in `wiki/quality/`.
3. Skill: `metaldocs-screen-qa` (see `.claude/skills/`).

---

## 4. Conventions you'll bump into immediately

- **Wiki is source of truth.** When you change code that a wiki page anchors via `path:line`, update the page in the same PR. Bump `Last verified:`.
- **Diagrams as code.** Mermaid only. No PNGs. New diagrams go in [`wiki/diagrams/`](diagrams/) and are linked from where they're used.
- **ADRs are immutable.** A new decision = a new file in `wiki/decisions/`. Never edit a past ADR.
- **Script-truth.** Local dev uses canonical scripts under `scripts/`. If you find yourself running `go run` or `pnpm exec foo` by hand, you've probably wandered off the rails. Check the script.
- **Contract-first API.** OpenAPI YAML → `oapi-codegen` → handlers + FE types. Hand-editing `api.gen.go` is forbidden.
- **No HashRouter, no `src/api/` legacy paths in FE.** See [`wiki/architecture/frontend-structure.md`](architecture/frontend-structure.md).
- **Quality gates.** Non-trivial work runs through the gates in [`CLAUDE.md`](../CLAUDE.md) §4 (Implementation truth → Review → QA → Regression → Evidence). Don't claim "done" without evidence.

---

## 5. Your first PR

Suggested ramp:

1. **Pick a "good first issue"** or a typo/wiki-rot you spotted while reading above.
2. Branch: `<type>/<short-slug>` (`fix/`, `feat/`, `docs/`, `chore/`, `refactor/`).
3. Run the relevant checks locally:
   - Backend: `go build ./... && go test ./...`
   - Frontend: `corepack pnpm tsc --noEmit -p tsconfig.build.json && corepack pnpm exec vitest run`
4. PR description follows the template you'll see on first push. Include the test plan.
5. Tag a reviewer who owns the touched module (see the module page).

---

## 6. Where to get unstuck

- **"How does X work?"** → start at the relevant `wiki/modules/<name>.md`, then the linked sequence diagram, then the code anchors.
- **"Why did we do it this way?"** → grep `wiki/decisions/` for the topic. If no ADR exists, that's a candidate to write one.
- **"My local dev is broken."** → [`wiki/references/local-dev-startup.md`](references/local-dev-startup.md) + [`local-dev-credentials.md`](references/local-dev-credentials.md). If the script lies, fix the script.
- **"The API/FE/backend agent says X."** → agents are scoped helpers, not authority. Verify with the wiki + the code before acting.

---

## 7. What's deliberately NOT in v1 (so you don't try to use it)

- **No real-time co-editing.** Single-editor-per-doc with optimistic concurrency (If-Match on revision). See [`sequence-edit-autosave.md`](diagrams/sequence-edit-autosave.md) §"What this is NOT."
- **No external SSO/IdP.** MetalDocs owns identity.
- **No server-side template fill-in for end users.** All filling happens in the browser editor; server-side `/render/docx` is dead code as of 2026-06-01.
- **No PDF route through docx-renderer.** PDF goes Go → Gotenberg directly (`internal/platform/servicebus/gotenberg_pdf.go`). docx-renderer's only live route is `/render/fanout` (used by freeze).

When in doubt, ask. The codebase is younger than it looks; conventions are written down because we still remember why we picked them.
