# Mission: Frontend Screen Completion

> **Status:** Operator-approved (2026-06-21) — ready for M0 in a fresh `/milestone` session
> **Date:** 2026-06-21  ·  **Branch of record:** main
> **Type:** enhancement (with embedded greenfield slices)
> **Slug:** `frontend-screen-completion`  ·  **Owner / operator:** leandrotca
> **Evidence base:** `./discovery-brief.md`  ·  **Program index:** `./README.md`
> **Governs:** Milestones M0..M5 below. Each milestone gets its own plan via the `milestone` skill,
> executed in a fresh session. This file is the **stable governing contract** — it says *what* the
> mission is and *what proves it done*; it contains **no execution detail**.

---

## 1. Problem / why now

The MetalDocs backend reached Grade-A and was signed off 2026-06-21 (HEAD `d477e9f0`). The frontend was paused mid-screen-redesign ~two weeks earlier to do that backend work, and its state was never reconciled. A 4-agent sweep this session (see `./discovery-brief.md`) found the app is **most-of-the-way done but not finishable as-is**: the home screen ships mock data (finding 6), two screens are dead empty stubs with a duplicate landing-route bug (findings 2–4), two screens are blocked on backend endpoints that don't exist (findings 8–11), one published-document screen carries six "em breve" placeholders (finding 12), and two screens (Documento Obsoleto, Detalhe Signoff) were never built (findings 13–14). The tracker that should record all this is two weeks stale and wrong on half its rows (finding 1).

The app cannot ship as a Professional-SaaS product while the landing screen is fake, routes render empty shells, and core document flows dead-end at "em breve". Now is the moment: the backend bar is set and green, the screen-implementation workflow + both reviewer agents exist (finding 16), and the gap is fully mapped. No solution here — just the gap.

## 2. Goals / Non-Goals

**Goals**
- Every in-scope user-facing screen is **production-complete**: real API data (no mocks), redesign design-system tokens, and an on-record **APPROVE from both** `frontend-screen-reviewer` and `frontend-code-reviewer`, with tests green.
- The missing backend endpoints that block screens (Distribuição fanout/coverage, Notifications) are **built to the Grade-A bar** — contract-first, ADR-backed, api-lint `-strict` = 0, integration-tested, all 6 CI guards still green.
- Zero mock-data screens, zero dead/duplicate routes, zero empty no-API shells in the routed app.
- The screen tracker reflects **verified reality** and stays the durable resume doc.
- Net-new screens (Documento Obsoleto, Detalhe Signoff) are implemented from their design sources through the same workflow + gates.

**Non-Goals**
- **Not** re-reviewing or re-styling screens already shipped DONE (Login, Library, Editor, wizards, Templates, Inbox, Route-admin, IAM Admin Center). They are out of scope unless a milestone provably regresses one.
- **Not** building `alternativas-inicio-caixa` or `catalogo-slots` — **CUT** (D3): no route, no NOTES, no product intent.
- **Not** a pixel-perfect re-audit of the whole design system — the design tokens and primitives are settled; this mission consumes them, it does not redesign them.
- **Not** expanding backend scope beyond the endpoints that directly unblock an in-scope screen. New backend is a means to finish a screen, not an open-ended platform program.
- **Not** merging or pushing anything — no-merge-by-agent stands.

## 3. Locked decisions (operator-approved)

| # | Decision | Value |
|---|----------|-------|
| D1 | Backend scope | **Full-stack.** Missing endpoints that block a screen (Distribuição fanout/coverage M2, Notifications M3, any Publicado-stub backend M4) are built as their own features to the Grade-A bar (contract-first, ADR, api-lint, integration tests, 6 CI guards green). |
| D2 | Per-screen gate | A screen counts done only when **both** `frontend-screen-reviewer` (visual/parity) **and** `frontend-code-reviewer` (architecture/maintainability) return **APPROVE** on record, **and** its tests are green. |
| D3 | Cut list | `alternativas-inicio-caixa` + `catalogo-slots` are **CUT** with rationale. `biblioteca` is already shipped as `LibraryPage` (not a gap). |
| D4 | Execution | One `mission.md` governs; each milestone runs via the `milestone` skill in a **fresh session**; per-feature consumer-contract spec gate + per-milestone `milestone-validator`; HS-1 operator gate at every milestone boundary. |
| D5 | Sequencing | **Clear-the-ground first** (M0 truth+routing), **wire-only quick win next** (M1 Dashboard), **backend-blocked screens in the isolated middle** (M2 Distribuição, M3 Notifications), **dependent + net-new last** (M4 Publicado+Obsoleto depends on M2 fanout; M5 Signoff+Taxonomy net-new/polish). Net-new and dependent work last so it can't regress earlier milestones. |
| D6 | Proof of done | Independent `mission-validator` judges a full-app screen audit (§8): every in-scope screen production-complete + 0 mock/dead-route/stub remaining + backend gates green. |
| D7 | Operations/Audit fate | **Delete both** `OperationsPage` + `AuditPage` dead stubs + their routes (M0/F0.3). IAM Admin Center (`/admin/*`) already owns metrics/audit/sessions; no separate top-level Operations or Audit screen. |

## 4. Discovery summary

A 4-agent parallel sweep + a main-session skeptic pass (OpenAPI grep + router read) mapped every routed screen and every design-source slug to a verified completeness verdict, and confirmed which screens are blocked on missing backend endpoints. Confidence is high: per-page verdicts and endpoint-existence are **verified** from files read this session; the only **assumed** finding (two unspecced slugs) was operator-confirmed CUT. Full evidence: `./discovery-brief.md`. Every milestone below traces to a numbered finding there.

## 5. Work / requirement inventory

| # | Item (site / requirement) | Class / kind | Milestone |
|---|---------------------------|--------------|-----------|
| 1 | Rewrite `wiki/implementation/screen-redesign-tracker.md` to verified 2026-06-21 state (finding 1) | truth-reset | M0 / F0.1 |
| 2 | Resolve duplicate `index: true` (dashboard vs operations `routes.tsx:5`) (finding 2) | routing bug | M0 / F0.2 |
| 3 | Dispose `OperationsPage` + `AuditPage` empty dead stubs (findings 3–4) | dead-code | M0 / F0.3 |
| 4 | Record CUT list + author per-screen Definition-of-Done + reviewer-gate wiring (findings 5, 16) | governance | M0 / F0.4 |
| 5 | Dashboard real stats — kill `MOCK_STATS` (findings 6–7) | wire-only | M1 / F1.1 |
| 6 | Dashboard real activity — kill `MOCK_ACTIVITY` (findings 6–7) | wire-only | M1 / F1.2 |
| 7 | Distribuição fanout/coverage backend endpoint, contract-first + ADR (finding 9) | greenfield BE | M2 / F2.1, F2.2 |
| 8 | Wire `DocumentDistributionPage` to real fanout; remove illustrative markers (finding 8) | screen | M2 / F2.3 |
| 9 | Notifications backend endpoints (list/stream/mark-read), contract-first + ADR (finding 11) | greenfield BE | M3 / F3.1, F3.2 |
| 10 | Wire `NotificationsPage` to real notifications API (finding 10) | screen | M3 / F3.3 |
| 11 | Close `DocumentPublishedPage` backlog stubs (PDF download, coverage, related docs, comments) (finding 12) | screen + partial BE | M4 / F4.1 |
| 12 | Build Documento Obsoleto as `obsolete` variant of Publicado (finding 13) | greenfield screen | M4 / F4.2 |
| 13 | Build Detalhe Signoff standalone screen (finding 14) | greenfield screen | M5 / F5.1 |
| 14 | Restyle `TaxonomyAdminPage` inline styles → redesign tokens (finding 15) | styling polish | M5 / F5.2 |
| — | `alternativas-inicio-caixa`, `catalogo-slots` | **out-of-scope — CUT (D3)** | — |
| — | `content-builder` wrapper internals | **out-of-scope** — thin wrapper, not a target screen; surfaces via HS-6 if it proves a real gap | — |
| — | Already-DONE screens (Login, Library, Editor, wizards, Templates, Inbox, Route-admin, IAM Admin Center) | **out-of-scope** — shipped; not re-reviewed | — |

## 6. Program architecture (by reference)

This mission executes via the `milestone` skill. The per-feature close-out loop, the per-feature consumer-contract spec gate, and the per-milestone `milestone-validator` gate are defined there — **not duplicated here**. See `.claude/skills/milestone/SKILL.md`. Per-screen execution additionally follows the documented `metaldocs-screen-implementation` workflow (Phase 0 audit → build → Phase 4 assembly) and its two reviewer-agent gates (`frontend-screen-reviewer`, `frontend-code-reviewer`). Backend features additionally follow the backend QA checklist (`wiki/quality/backend-api-qa-checklist.md`) and inherit the 6 CI guards + `api-lint -strict`.

```
Mission: frontend-screen-completion
└── Milestone (M0..M5)          ── each: features → milestone-validator gate → HS-1 operator gate
    └── Feature (Fx.y)          ── each: spec(consumer-contract) → plan → TDD → evidence
                                   FE features also: screen-reviewer + code-reviewer APPROVE (D2)
                                   BE features also: api-lint -strict 0 + 6 CI guards + integration tests
Terminal acceptance (§8)        ── independent mission-validator, after M5 passes
```

## 7. Milestones

### M0 — Truth reset & structural cleanup
**Objective:** make the routed app honest — one landing route, no dead stubs, a correct tracker, and the per-screen quality bar written down — so later milestones build on solid ground.

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F0.1 `tracker-rewrite` | Rewrite the screen tracker to the verified 2026-06-21 state (done / partial / stub / not-started / cut) | Every tracker row matches a grep of implemented pages + this mission's inventory; no row contradicts reality |
| F0.2 `index-route-fix` | Resolve the duplicate `index: true`; `/` renders exactly one intended page | Exactly one index route in the router; a router test asserts `/` → intended component; `npm run build` clean |
| F0.3 `dead-stub-disposition` | **Delete both** `OperationsPage` + `AuditPage` (empty `OperationsCenter` shells) and their routes — IAM Admin Center already owns metrics/audit/sessions (D7) | No route renders an empty no-API shell; `OperationsCenter`/`OperationsPage`/`AuditPage` deleted; routes removed; build + FE tests green |
| F0.4 `cut-list-and-dod` | Record `alternativas-inicio-caixa` + `catalogo-slots` as CUT; author the per-screen Definition-of-Done checklist + reviewer-gate wiring used by every later milestone | CUT slugs documented + absent from router; a DoD doc exists enumerating the D2 gate (both reviewers + tests) |

**Milestone gate:** frontend build + test suite green; router has a single index route; no dead-stub route remains; tracker + cut-list recorded. Validated by the `milestone-validator`.

### M1 — Dashboard real data (frontend-only)
**Objective:** the home screen renders 100% live data — zero mocks — through existing backend endpoints.

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F1.1 `dashboard-stats-wire` | Replace `MOCK_STATS` with real queries (`/documents/stats`, `/iam/kpi`) | `grep MOCK_STATS` = 0 in dashboard; stat cards render live values; query-hook tests green |
| F1.2 `dashboard-activity-wire` | Replace `MOCK_ACTIVITY` with a real activity feed (`/audit/events`) | `grep MOCK_ACTIVITY` = 0; activity list renders live; tests green |

**Milestone gate:** `grep -rE "MOCK_" dashboard/` = 0; DashboardPage passes `frontend-screen-reviewer` + `frontend-code-reviewer` APPROVE (D2); tests green. Validated by the `milestone-validator`.

### M2 — Distribuição (full-stack)
**Objective:** the document distribution/fanout screen is backed by a real, Grade-A backend endpoint and renders live coverage — no illustrative data.

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F2.1 `fanout-contract` | ADR + OpenAPI spec for the distribution/fanout/coverage endpoint (consumer-contract first); regen FE types | `api-lint -strict` parses the new path = 0 violations; ADR merged; generated FE types present; spec review |
| F2.2 `fanout-backend` | Implement the endpoint to the Grade-A bar (module/repo/authz, in-tx audit where applicable) | Integration test passes against live PG; `api-lint -strict` = 0; all 6 CI guards green; `go build`/`vet`/test green |
| F2.3 `distribuicao-wire` | Wire `DocumentDistributionPage` to the real fanout API; remove "Dados ilustrativos · Em breve" + enable CTAs | `grep "Em breve"\|illustrative` = 0 in the page; live coverage renders; both reviewers APPROVE; tests green |

**Milestone gate:** backend gates (api-lint 0 + 6 guards + integration test) AND the screen passes both reviewers; no illustrative markers remain. Validated by the `milestone-validator`.

### M3 — Notifications (full-stack)
**Objective:** the notifications center shows real notifications from a real backend — the stub is gone end-to-end.

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F3.1 `notifications-contract` | ADR + OpenAPI spec for notifications (list / unread-count / mark-read; stream if warranted); regen FE types | `api-lint -strict` = 0 on the new paths; ADR merged; FE types present; spec review |
| F3.2 `notifications-backend` | Implement endpoints to the Grade-A bar | Integration test passes; api-lint 0; 6 CI guards green; build/vet/test green |
| F3.3 `notifications-wire` | Replace the noop `notifications.ts` + empty page with real queries; restyle to redesign tokens | Page renders live notifications + mark-read works; no empty-array stub; both reviewers APPROVE; tests green |

**Milestone gate:** backend gates AND the screen passes both reviewers; the noop stub is removed. Validated by the `milestone-validator`.

### M4 — Documento Publicado completion + Documento Obsoleto
**Objective:** the published-document screen has no "em breve" gaps for in-scope items, and its obsolete variant exists. (Depends on M2 fanout for the coverage card.)

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F4.1 `publicado-stubs` | Close the in-scope Publicado placeholders (PDF download via existing export/render; coverage via M2 fanout; related docs; comments) — any genuinely-out-of-scope item becomes a written defer-with-trigger, not a silent stub | No "em breve" placeholder for an in-scope item; each remaining gap is an explicit defer row with a trigger; both reviewers APPROVE; tests green |
| F4.2 `obsoleto-variant` | Implement Documento Obsoleto as the `obsolete` state variant of Publicado (banner + status logic + route) | Obsolete state renders correctly at its route; shares the Publicado component; both reviewers APPROVE; tests green |

**Milestone gate:** Publicado + Obsoleto both pass both reviewers; remaining gaps are deferred-with-trigger, not stubbed. Validated by the `milestone-validator`.

### M5 — Detalhe Signoff + Taxonomy Admin restyle (net-new / polish — last)
**Objective:** the last owed screen (sign-off detail) is built, and the one off-design-system screen (Taxonomy Admin) is brought onto the redesign tokens. Risk-isolating last per D5.

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F5.1 `signoff-detail` | Build the Detalhe Signoff screen from its design source (A4 diff view, approval-flow panel, decision form), wired to existing approval/sign-off APIs | Screen renders at its route; the decision flow works against the real API; both reviewers APPROVE; tests green |
| F5.2 `taxonomy-restyle` | Convert `TaxonomyAdminPage` inline styles to redesign tokens with no behavior change | `grep` inline `style=`/non-token CSS in the page = 0; existing taxonomy tests still green (no regression); both reviewers APPROVE |

**Milestone gate:** Signoff passes both reviewers; Taxonomy restyle passes both reviewers with zero behavior regression. Validated by the `milestone-validator`.

## 8. ★ Terminal acceptance — definition of done (written up front)

- **Pass bar (the mission shall be X):** every in-scope screen (Dashboard, Distribuição, Notifications, Documento Publicado, Documento Obsoleto, Detalhe Signoff, Taxonomy Admin — plus the M0 structural fixes) is **production-complete**: real API data, redesign tokens, **both** reviewer agents APPROVE on record, tests green. **Zero** mock-data screens, **zero** dead/duplicate routes, **zero** empty no-API shells in the routed app. The new backend endpoints (Distribuição fanout, Notifications) pass `api-lint -strict` = 0 with all **6 CI guards green** and integration tests passing — **0 backend regressions**. The CUT slugs are absent from the router. The screen tracker matches verified reality.
- **What to validate:**
  1. For each in-scope screen: a real query hook (no `MOCK_`/illustrative/"em breve" literal), redesign-token styling, and an APPROVE-on-record from both `frontend-screen-reviewer` and `frontend-code-reviewer` (the milestone evidence carries the verdicts).
  2. Router: exactly one index route; no route renders an empty `OperationsCenter`-style shell; `alternativas-inicio-caixa`/`catalogo-slots` not routed.
  3. Backend: `go build ./...` + `go vet ./...` clean; `api-lint -strict` = 0; `go run ./tools/cilint/...` exit 0 (all 6 guards); the new endpoints' integration tests pass.
  4. Frontend: `make test` (vitest) green; `npm run build` clean.
  5. Tracker (`screen-redesign-tracker.md`) rows match the implemented set.
- **How to validate:**
  - **Fan-out (main session runs, validator judges the artifact):** a per-screen completion re-audit — one agent per in-scope screen verifying real-API + tokens + reviewer-APPROVE-on-record — producing a single audit artifact. The main session runs this with `Workflow`/`Agent`; the `mission-validator` judges the artifact against the pass bar and independently **spot-checks** load-bearing claims (re-grep a sample of "0 MOCK_/em breve" sites; re-run one named backend integration test).
  - **Deterministic (validator runs itself with `Bash`/`Grep`):** `grep -rE "MOCK_|Dados ilustrativos|Em breve" frontend/apps/web/src/features` over in-scope screens = 0; single-index-route check; cut-slug-not-routed grep; `go build`/`go vet`/`api-lint -strict`/`cilint` exit-0; `make test` + `npm run build`.
- **Who validates:** the independent `mission-validator` subagent (`.claude/agents/mission-validator.md`). It judges and writes `qa/mission-validation.md` only — never edits code or flips status.
- **On miss (HS-5):** the missed criteria become a bounded remediation micro-milestone, run through `milestone`, then the `mission-validator` is re-dispatched. The operator decides continue vs replan at each loop.

## 9. Hard-stop catalog

| ID | Trigger | Action |
|----|---------|--------|
| HS-1 | Every milestone boundary | Operator review gate; no next milestone / no merge without approval |
| HS-2 | A screen fix implies redesign outside its assigned boundary (e.g. a shared primitive or design token must change) | Stop; report the boundary + minimum prerequisite plan; no symptom-patch |
| HS-3 | A prerequisite boundary fails (build / runnable / auth / route / contract↔generated truth, esp. new BE endpoints) | Repair the prerequisite first (contract-first regen order); rerun the checkpoint; resume |
| HS-4 | A `milestone-validator` returns FAIL | Open the named fix feature; re-run its lifecycle; re-dispatch the validator |
| HS-5 | The terminal `mission-validator` misses the §8 bar | Bounded remediation micro-milestone; re-dispatch; operator decides continue vs replan |
| HS-6 | Scope drift / off-plan discovery (e.g. `content-builder` proves a real gap; a screen needs unforeseen backend) | Stop; surface the deviation; re-interview before continuing |

## 10. Constraints respected

- **Contract-first regen order** for all new backend surface: spec → OpenAPI → codegen → FE types; FE never hand-authors response types (HS-3).
- **6 backend CI guards stay green** (gitleaks · platformboundary · PostCommitAudit · chain-order · nosqltxindomain · nodualmode) + `api-lint -strict` = 0 for any backend feature.
- **ADR for new endpoints/capabilities**; PRs cite REQ IDs; MUST-deviations need an ADR (`backend-target-architecture.md` governance).
- **Advisory-lock hazard (H-PRE-1):** no authz-recording read inside a lock-holding atomic tx.
- **Skill routing + FE/BE/DB boundaries** (CLAUDE.md): FE screens via `metaldocs-screen-implementation` + both reviewer agents; BE via backend QA checklist.
- **Design system is consumed, not redesigned:** use `tokens.css` + existing primitives; changing a shared primitive trips HS-2.
- **No merge / no push by the agent;** commits allowed after verified work (CLAUDE.md §5.0).
- **Model policy:** sonnet implement/review, haiku mechanical, never fable workers, ≤15 concurrent agents.

## 11. Execution model

One `mission.md` governs all milestones. Each milestone → its own plan via `milestone`, executed in a **fresh session**, subagent-driven, with a per-feature consumer-contract spec gate and a per-milestone `milestone-validator`. Operator gate between every milestone (HS-1); **no merge by the agent**. Per-screen FE work runs the `metaldocs-screen-implementation` workflow and ends on both reviewer agents' APPROVE (D2). Backend features inherit the Grade-A gates. Token discipline: parallel fan-out only where it pays (the terminal re-audit); everything else direct tools. Model policy per §10.

## 12. End-state / reconciliation

Fill only when M5 has passed and the terminal gate is green:
- [ ] Every planned feature (M0..M5) has a complete evidence row.
- [ ] Zero unplanned scope merged; anything added is recorded with rationale.
- [ ] Every bounded defer has a written trigger and an owner.
- [ ] Terminal acceptance (§8) passed — link `qa/mission-validation.md`.
- [ ] Operator sign-off: <date / name>
