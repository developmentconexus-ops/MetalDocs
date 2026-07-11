# Unit 2.5 — Route Builder v2 — Evidence

**Session:** orchestrator (Opus), worktree `clever-wiles-94d525`, budget ≤250k.
**Scope:** FE-only profile-governed approval route builder v2 (consumes G1 `governance_class`, merged main).
**Design spec:** `docs/superpowers/specs/2026-07-11-route-builder-v2-design.md`.
**Workflow spec:** `docs/superpowers/specs/2026-07-10-review-approval-workflow-model.md` (R1–R5).

## P0 gate
`developing-new-work` **skipped** — not a new backend module/feature/invariant. FE screen rebuild on an
already-merged contract (G1 ran its own gate at unit 2.1). Rationale recorded per HARNESS §2 (P0 = new
feature/module only).

## Bounded deviations (for HS-1)
1. v2 mock `route_builder_mock_v2` absent from repo — workflow spec §mock is the design authority;
   v1 `design-source/route-admin` is visual base. Noted; not a blocker.
2. L3 browser-pixel QA on :80 deferred to operator — deployed web image predates the commits (needs
   rebuild from branch) AND the admin-gated route needs a login password this agent may not enter.
   Interim proof = RTL rendered-DOM suite; exact operator script + rebuild request below.
3. Adjacent bug fixed in-scope: `useRouteAdminMutations.stageRequestToSummary` (optimistic cache) was
   dropping `stage_kind` — a defect this unit's field introduction exposed. Fixed here (not deferred)
   because it directly corrupts the feature; a regression test guards it.

## Dispatch ledger
| Slice | Implementer (sonnet) | Reviewer(s) (independent, sonnet) | Verdict | Commit |
|---|---|---|---|---|
| S0 regen+thread governance_class | general-purpose a837fa | cavecrew-reviewer adfc4a | clean (0 findings) | f75a175f |
| S1 governance policy logic | general-purpose ac156d | cavecrew-reviewer aa0e0b | clean (0 findings) | ee2fa3f2 |
| S2 presentational blocks | general-purpose afd5cc | cavecrew-reviewer a84151 (3🟡 copy/enum centralization) → fix general-purpose a3168a | 3🟡 resolved | b89b8f18 |
| S3 assemble RouteEditorDialog v2 | general-purpose a6eb5e | frontend-screen-reviewer a6827d (APPROVE w/ nits) + frontend-code-reviewer a553ce (APPROVE w/ nits) | remediated → clean | f23009bb |
| S3-fix reviewer remediation | general-purpose a5e8da | cavecrew-reviewer a6099d (all 5 fixes verified correct, no regression) | clean | f23009bb |

Implementer ≠ reviewer held on every slice. S2's reviewer found 3 centralization yellows; a separate
fix worker resolved them before commit (re-greened). S3's two page-level reviewers (independent, both
APPROVE-WITH-NITS) raised 4 Majors → all fixed in a scoped remediation worker (implementer ≠ prior
implementer), then re-reviewed clean:
- **Major (bug):** `stageRequestToSummary` dropped `stage_kind` → optimistic cache reverted kind on
  re-open. Fixed + regression test (never-resolving mock asserts mid-flight patch keeps kind).
- **Major (a11y):** `role="radiogroup"` lacked accessible name → `aria-labelledby` per-stage unique ids.
- **Major (test):** archived-profile unknown-gc edit path untested → added (no badge, neutral note,
  Save enabled, validation skipped, backend sole enforcer).
- **Major (rule >400 LOC):** `RouteEditorDialog` 567→**305 LOC** via `StageCard` + `routeDraft` extraction.
- Minors: aria-live badge, aria-hidden arrow, R2 rounds comment. One duplicate narration comment trimmed
  by orchestrator inline (trivial, no behavior).

## Verification ladder (from clean state)
- **L0 typecheck** — `pnpm exec tsc --noEmit -p tsconfig.build.json` → **PASS** (zero errors). No FE
  eslint config exists in `frontend/apps/web` (build script's `tsc --noEmit` is the type gate); FE lint
  N/A. Backend Go lints (api-lint/module-boundaries) N/A — FE-only unit.
- **L1 vitest** — `pnpm exec vitest run` (full suite) post-remediation → **PASS 140 files / 882 tests**,
  58s. (Pre-remediation run was 139/874.) S0's new required `governanceClass` field rippled to one
  fixture (`StepConfirm.test.tsx`), fixed in S0. The new route-builder tests render the REAL dialog in
  jsdom and drive it (profile select → policy badge per class, livre block, controlado ≥1-signature
  validation, review-after-approval order rejection, stage_kind in submit payload, radiogroup
  accessible names, archived-profile edit degrade, optimistic stage_kind cache).
- **Production build** — `pnpm build` (`tsc --noEmit` + `vite build`) → **PASS** in 11s; `RouteAdminPage`
  chunk built (24 kB). Only warning is the pre-existing >500 kB editor chunk (`HyperlinkDialog`), not
  this unit. Confirms v2 is deployable.
- **L3 browser-pixel QA on :80 — DEFERRED to operator (HS-1). Reason (honest, not a skip):**
  1. The running `metaldocs-web:dev` container was built 2026-07-11 09:32, **before** this unit's commits
     — the deployed image does NOT contain v2. A rendered walk on :80 today would validate the OLD
     route-admin, i.e. a false green. **Requires a web container rebuild from branch
     `claude/clever-wiles-94d525`** (hub-owned; I must not rebuild/restart the shared stack).
  2. `/workspace/approval-routes` is `requiresAdmin`-gated; a browser walk needs an authenticated admin
     login. Entering a password to authenticate is a prohibited action for this agent, so I cannot
     self-drive the gated route. Per HARNESS §6 the rendered pass belongs to a fresh browser-QA persona /
     operator, not a curl fake — and curl-only would be a FAIL, so none was attempted.
  Interim behavioral proof = the RTL rendered-DOM suite above (real React tree, not curl). Formal
  browser-pixel verdict pending the rebuild + operator/QA-persona walk below.

## Operator L3 rendered-QA script (run after web rebuild from this branch)
Persona: admin user, gateway :80, rendered UI only, no curl. Route builder has NO signature ceremony
(that is unit 2.4's approver panel), so no signature-password step is involved here.
1. Login as an admin; open **Administração de Rotas** (`/workspace/approval-routes`).
2. **Nova rota** → select a **Controlado** profile → assert badge reads **"Obrigatório ≥1 assinatura"**.
   Add one stage, set its kind to **Revisão** only → Save → assert it BLOCKS with the ≥1-signature
   message. Flip the stage kind to **Assinatura** → Save succeeds.
3. New route → **Simples** profile → assert badge **"Assinatura opcional"**; a review-only route saves.
4. New route → **Livre** profile → assert the builder is **blocked** with the friendly perfil-livre
   message and Save disabled (and that a forced attempt yields the backend 409/422 friendly message).
5. Multi-stage: add Revisão → Revisão → Assinatura; confirm the **live flow preview** renders them in
   order with round separation; confirm the **SoD** ("autor excluído") and **"Aprovar já"** overlap
   notes are visible. Try ordering an approval stage BEFORE a review stage → assert it is rejected.
6. Edit an existing route whose profile is archived → assert no policy badge, the neutral
   "Perfil não classificado" note, and Save remains enabled (backend enforces).
7. Keyboard/AT spot-check: tab into a stage's kind + quorum pills → each radiogroup announces its
   per-stage name (aria-labelledby).
Capture per-step screenshot + console/network; write verdict to a `qa/browser-qa-<date>.md`.

## Commands + outcomes
- `pnpm gen:api` (S0) — regenerated `index.d.ts` with governance_class. OK.
- `pnpm exec tsc --noEmit -p tsconfig.build.json` — zero errors (twice: post-S3, post-remediation).
- `pnpm exec vitest run` — 140 files / 882 tests passed (post-remediation).
- `pnpm build` — vite production build OK, RouteAdminPage chunk emitted.
- `wc -l RouteEditorDialog.tsx` — 305 (under the 400-LOC rule).
- `docker ps` — `metaldocs-web:dev` created 09:32 (pre-commit) → confirms :80 lacks v2.
