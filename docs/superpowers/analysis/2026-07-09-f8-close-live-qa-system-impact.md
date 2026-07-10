# System-impact analysis — F2d.8 milestone close (UI-driven live QA)

**Date:** 2026-07-09
**Intent (one line):** Close M2d by **UI-driven** live QA on the real running stack — drive the single
mode-adaptive screen (`/documents/:id`) through the full approval lifecycle on BOTH route shapes
(review+approval AND approval-only), plus `changes_requested` round-trip, delegation signoff, and observer
view — capturing per-step DOM/a11y UI evidence that proves the M2c deviation is closed at the root (review
stage → verdict CTAs, NEVER a signature panel; approval stage → signature panel only when eligible).
**Work type:** feature — **QA / verification only** (zero source, zero backend, zero contract change).
**Author:** developing-new-work skill
**Verdict:** 🟢 Green *(see §10)* — conditional on the HS-3 runnability prerequisite, verified below.

> This is a **close/verification** feature, not a code feature. It writes exactly one artifact —
> `qa/live-qa-log.md` — and touches no product source, no backend, no contract, no migration. The ten
> sections are therefore mostly **N/A with a one-line reason**; the load-bearing sections are §2
> (runnability foundation / HS-3), §8 (what gets QA'd and the validator's forbidden-list), and §10.

---

## 1. Classify & own

- **Work type:** feature — milestone-close live QA (evidence-producing, not code-producing).
- **Owning module(s):** none in the code sense — this is a **QA activity** over the FE surfaces built by
  F2d.1–F2d.7 (`features/documents` single screen, `features/approval` timeline/DecisionFooter/signature).
  The artifact lands under the milestone folder (`.../f8-close-live-qa/` + `.../qa/live-qa-log.md`).
- **Explicitly NOT owning / must NOT touch:** all product source. If live QA finds a defect, that is a
  **finding recorded in the log** (→ HS-4 fix-feature if blocking), NOT an inline patch inside this feature.
  Zero backend, zero FE source, zero contract diff — same discipline as a read-only audit.
- **Cross-module edges:** none created. This exercises existing edges through the UI; it adds none.
- **Ambiguity?** None. AS-3 not triggered.

## 2. Foundation verdict (the load-bearing section — HS-3)

- **Base you'd build on:** the single-screen destination (ADR 0080) is live; F2d.1–F2d.7 are all closed
  with evidence. The QA validates that ratified foundation end-to-end on the real stack.
- **Runnability (HS-3 prerequisite — VERIFIED before this verdict):** the full stack is up in docker —
  `metaldocs-api` (healthy, host :8081; `/healthz` → 200, `/api/v1/*` → 401 problem+json), `metaldocs-worker`,
  `metaldocs-jobs`, `metaldocs-web`, `metaldocs-gateway` (:80), `postgres` (:5433), `redis`, `minio`,
  `gotenberg`, `docx-renderer` — all `Up`. **Prerequisite MET.**
- **Critical foundation caveat (drives the method):** the running `metaldocs-web` container is a ~3h-old
  production build — it does **not** contain the not-pushed F2d.5/5b/6/7 FE commits (single screen, cockpit
  retirement). QA'ing that stale image would be a **false green** (the exact failure mode M2c's F8 curl-only
  QA hit). **Therefore live QA MUST run against the working-tree FE**, served by the vite **dev server**
  (`npm run dev`, port 4173, current source + HMR), proxying `/api/v1` → the running api on :8081
  (`.env.local` `VITE_API_PROXY_TARGET=http://127.0.0.1:8081`; 8080 is the unrelated marketplace backend).
  The backend/worker/jobs docker stack is reused as-is (no FE change touches backend).
- **Sound, or legacy/patch?** Sound. QA'ing the real foundation is the global-maximum close. Not AS-2.
- **Fallback if the dev server won't boot** (memory: pnpm junction drift can break vite): rebuild the
  `metaldocs-web` docker image from current source and QA the gateway (:80). Recorded so the method is not
  silently downgraded. Curl remains **corroborating** backend evidence only — never the primary proof.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied |
|-----------|----------|---------------|
| AuthZ = capabilities, never roles | No (exercised, not changed) | QA observes capability-gated UI; changes no gate. |
| Contract-first (OpenAPI + oapi-codegen) | No | No route/DTO/spec edit. |
| Multi-tenant pooled | No | No query/table touched. |
| Async = transactional outbox | No | No side effect changed (publish exercised through existing path). |
| DB enforces invariants | No | No DB/migration touched. |
| Cross-module via published interface only | No | No code; exercises existing edges via UI. |

No violation. AS-1 not triggered.

## 4. Capability wiring
**N/A** — no capability added/changed/removed. `TestCapabilityRegistrySize` unchanged.

## 5. Module wiring
**N/A** — no module born/retired. This is a QA activity, not a bounded context.

## 6. Frameworks to reuse, not reinvent
Reuse: the **browser-preview tool workflow** (`preview_start` the `metaldocs-web` dev config → drive with
`preview_snapshot`/`preview_inspect`/`preview_click`/`preview_fill`/`preview_console_logs`/`preview_network`),
the existing **dev-seed personas** (prior F8 runs used `f8_admin`/`f8_approver`/`f8_author-test` — reuse the
seeded tenant + approval routes rather than hand-crafting data), and the existing approval routes that carry
both a review stage and an approval stage. No new test harness, no new seed script. Nothing reinvented.

## 7. Contract & data
- **OpenAPI-first:** no route added/changed.
- **Migration:** none.
- **Destructive change?** None to source. The QA *creates* documents/instances in the dev tenant (expected
  test data), which is non-destructive to the schema and isolated to dev.

## 8. Test & QA plan (load-bearing — the acceptance + forbidden-list)

- **Canonical evidence:** `qa/live-qa-log.md` — per-step, **UI-driven**, with DOM/a11y snapshots
  (`preview_snapshot`) and targeted `preview_inspect` assertions; `preview_network`/`preview_console_logs`
  for corroboration; `preview_screenshot` for the visual proof points.
- **Scenarios that MUST appear (milestone F2d.8 row + C4 quality-bar gate):**
  1. **Route shape A — review + approval:** draft → submit → **review stage** verdict(s)
     (`ready` and a `request_changes`) → `changes_requested` round-trip (author edits + resubmits) →
     **approval stage** signoff → publish. Prove at the review stage: verdict CTAs present, **signature
     panel ABSENT** (the M2c 412 defect made structurally impossible). Prove at the approval stage:
     signature panel present, frozen content.
  2. **Route shape B — approval-only:** draft → submit → approval-stage signoff → publish (no review stage).
  3. **Delegation signoff:** an approver eligible only via active delegation signs on the approval stage
     (`viewer.via_delegation_from` disclosure visible).
  4. **Observer view:** a non-eligible actor sees read-only + timeline, no verdict/signature CTAs.
  5. **Single-destination proof:** every persona lands on `/documents/:id` (no cockpit); a stale
     `/approvals/:id?decision=…` bookmark redirects there with the decision preserved.
- **Explicit FAIL conditions (validator forbidden-list — carried as locked constraints):**
  - A **review-stage** screen rendering a **signature panel** = **FAIL**.
  - A **curl-only** walkthrough (no UI evidence) = **FAIL** — curl is corroborating backend evidence only.
  - QA'ing the **stale web image** instead of current source (would false-green the retirement) = invalid.
- **Regression corroboration:** the FE suite (`documents` + `approval`, 399/399) and tsc(0) are already
  green from F2d.7; live QA is the functional complement, not a re-run of unit gates.

## 9. Docs / ADR
- **Wiki:** none required (QA artifact, not an architecture change).
- **ADR:** **no new ADR.** This feature records closure of the **M2c deviation** (tracked in the program
  README + `approval-remediation-m2b-m2c` memory) and the milestone quality-bar claim — it changes no
  decision.
- **REQ IDs cited:** milestone C4 (quality-bar/root-cause), F4 `stage_kind` contract, ADR 0078
  (viewer-facts), ADR 0080 (single destination).

## 10. Verdict & locked constraints

- **Verdict:** 🟢 Green — proceed. QA/verification-only close; no invariant, capability, backend, contract,
  or migration change. HS-3 runnability prerequisite is **met** (stack up; verified :8081).
- **Open hard-stops:** none. (If the dev server fails to boot → §2 fallback, still UI-driven; if live QA
  surfaces a blocking defect → HS-4 fix-feature, not an inline patch here.)
- **Locked constraints handed to execution:**
  1. **UI-driven or it doesn't count** — every acceptance claim backed by a `preview_snapshot`/`_inspect`
     DOM/a11y capture. Curl only corroborates the backend. Curl-only = FAIL.
  2. **QA current source, not the stale image** — serve the working-tree FE via the vite dev server
     (:4173, proxy → :8081); never QA the ~3h-old `metaldocs-web` build.
  3. **Zero source/backend/contract diff** — the only file this feature writes is `qa/live-qa-log.md`
     (+ this analysis + evidence.md + the milestone DONE marker). `.env.local` is throwaway local config.
  4. **Both route shapes + all five scenarios** (§8) or the close is incomplete; any dropped scenario is
     logged as a bounded defer with reason, never silently omitted.
  5. **The two structural FAIL conditions are binding** — review-stage signature panel = FAIL;
     curl-only = FAIL.
  6. **Findings are recorded, not patched** — a defect goes in the log (→ HS-4 if blocking); this feature
     does not edit product source.
