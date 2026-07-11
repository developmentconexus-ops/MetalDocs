# Unit 2.4 — Approver execution panel (frontend) — evidence

**Unit:** ROADMAP 2.4 (frontend follow-ups to review/approval workflow model, spec §4/§5).
**Branch:** `claude/amazing-saha-589663`.
**Contract:** FIXED — G1+G2+G3 merged on main. FE-only unit. Nothing under `internal/` or `api/openapi` changes.
**Owning surface:** `features/approval` decision-panel components, `features/shared/controlled-artifact` panel, `features/documents` signoff wiring, generated api-types.

---

## 1. Scope (4 tasks, spec §4 FE follow-ups + §5)

1. **Two actions only** on the approval stage (R3): "Assinar e aprovar" (password + legal e-sig ceremony) and
   "Solicitar mudanças" (comment-only, NO signature). Remove signed "Assinar e devolver" from the UI.
2. **Reject-default fix** (§5): `ArtifactDecisionPanel` no longer preselects any option; user must actively choose.
3. **Fast-forward "Aprovar já"** (G3): consume `ReviewVerdictResponse.fast_forward_eligible` + `next_stage_id`;
   offer the single-gesture fast-forward ceremony (`POST …/stages/{stage_id}/fast-forward`). Hint-only.
4. **Regen FE api-types** from current openapi (G2/G3 merged).

## 2. Foundation judgement (GM §3)

The consumed surface is sound, not legacy: `ArtifactDecisionPanel` is a tested kind-agnostic decision panel
reused by both documents (e-sig) and templates (review/publish); `DecisionFooter` already models the three-way
review/approve/observe gate; the API layer is contract-first codegen. Build on it — no rewrite.

## 3. GM fork record — approval "Solicitar mudanças" ceremony split (Slice C)

- **Option A (local max):** keep the two-radio `ArtifactDecisionPanel` and add per-option password/legal gating +
  per-option endpoint routing inside the shared panel. Bloats the kind-agnostic shared panel with approval-only
  dual-endpoint logic; templates would carry dead branches.
- **Option B (global max):** the approval stage's two actions hit **different endpoints with different ceremonies** —
  "Assinar e aprovar" → `signoff` approve (password+legal+content_hash); "Solicitar mudanças" → `reviewVerdict`
  `request_changes` (comment-only, no signature — G2 relaxes the guard to allow it on approval-kind stages). So the
  approval footer becomes symmetric with the review footer: the signature ceremony stays an **approve-only**
  `ArtifactDecisionPanel`, and request-changes reuses the existing comment-only dialog. The shared panel keeps its
  multi-option shape untouched for templates.
- **Boundary:** B is FE-only, inside the unit boundary; the fixed contract already supports both legs. → **B chosen.**

## 4. Fast-forward interpretation (Slice D) — reconciliation note for HS-1

The chip prose says "when `fast_forward_eligible` true **after recording a review verdict**, offer the ceremony".
Backend truth (`fast_forward_service.go`, `review_verdict_service.go`, `fast_forward_handler_test.go`): the
fast-forward endpoint records the `ready` verdict **and** the approve signoff itself in one tx (`If-Match` = the
**current** instance version, `stage_id` = the active **review** stage). Recording a normal `ready` verdict FIRST
completes the review stage, so a subsequent fast-forward call replays the verdict leg → `ErrActorAlreadySigned` /
`ErrFastForwardNotEligible`. Therefore fast-forward MUST be a **single standalone gesture**, not preceded by a
normal verdict.

**Implementation:** "Aprovar já" is offered **opportunistically** in the review footer (viewer eligible on the
active review stage AND a following approval stage the viewer is eligible on). The `fast_forward_eligible` flag is
honoured as a hint (never gates correctness — replay returns false). Backend is the authority; on
`ErrFastForward*` the UI shows a friendly fallback to the normal "Pronto para aprovação" path. Flagged for HS-1.

## 5. Slice plan / task board

| Slice | Task | Files | Gate |
|---|---|---|---|
| A | Regen api-types | `src/lib/api-types/index.d.ts` (generated) | additive-only diff, tsc clean — **DONE** |
| B | Remove reject preselection | `ArtifactDecisionPanel.tsx`, `types.ts`, `documentSignoffDecision.ts`, `DocumentWorkspacePage.tsx`, `DecisionFooter.tsx` | vitest: mounts unselected |
| C | Two-action approval footer | `documentSignoffDecision.ts`, `DecisionFooter.tsx` (+css), tests | vitest: approve-only sig + request_changes dialog; no "Assinar e devolver" |
| D | Fast-forward "Aprovar já" | `approvalApi.ts`, `approvalTypes.ts`, `useFastForwardMutation.ts`, `DecisionFooter.tsx`, `DocumentWorkspacePage.tsx`, `WorkspaceSidebar.tsx`, tests | vitest: offer+ceremony+call args |
| E | Verify + UI QA :80 | — | L0/L1 green, browser QA GREEN |

## 6. Dispatch ledger (HARNESS §4.5 — required)

| Slice | Implementer (sonnet) | Reviewer (independent sonnet) | Verdict | Commit |
|---|---|---|---|---|
| A | orchestrator (generated file, contract-first codegen — no hand diff) | tsc + additive-diff check | PASS | (folded into B commit) |
| B | sonnet general-purpose `a6bf24795b8ec1000` | independent sonnet `ad122f359dd0a76f3` | **PASS** (1 MINOR stale JSDoc → fixed inline) | 1c6db8dc |
| C | sonnet general-purpose `a6baff83ebfeffc83` | independent sonnet `a44fa71294c654d2f` | **PASS** (1 MINOR: `MeaningOfSignatureLine` reject branch unreachable for current caller → **accepted-defensive**: `tone` flows from shared `ArtifactDecisionOption.tone` which templates emit as `reject`; narrowing would force a call-site cast, a worse smell) | 9418d091 |
| D | sonnet general-purpose `a645293032f34bbd6` | independent sonnet `ab6c5203427fb3eeb` | **PASS** (0 findings; derivation + review-stage-id correctness core confirmed) | 6e71e127 |

## 7. Verification ladder

- L0 tsc (`tsconfig.build.json`): **CLEAN** on the integrated A–D tree (post-Slice-D, HEAD `6e71e127`).
- L1 vitest: **643/643 passed (96 files)** across `approval` + `documents` + `shared/controlled-artifact` + `templates` — templates untouched and green (no regression from the two-action / fast-forward changes).
- L2/L3 UI QA on :80: **BLOCKED — operator/hub dependency.** The chip's `HUB_SESSION_ID` (`local_39f1f842-1a02-4275-a6c9-2023312cd979`) no longer exists (not in `list_sessions`; session ended before this resumed run), so the REQUEST for a web-container rebuild could not be delivered. The mandatory :80 QA also requires the operator to perform login (QA personas are forbidden from typing passwords). Both dependencies are the operator's to resolve. Code is complete + L0/L1 green; only the browser-evidenced :80 QA remains. See §8 HS-1 (c).

## 8. Defers / HS-1 items

- HS-1 (a): fast-forward "opportunistic single-gesture" interpretation vs chip's "after recording a verdict" prose (see §4).
- HS-1 (b): `?decision=` deep-link preselection removed (spec §5 mandate) — inbox→panel now lands unselected.
- (carry) 21 CFR/MP jurisdiction of the meaning-of-signature copy (pre-existing, spec §6) — untouched.
- HS-1 (c): **:80 UI QA not yet executed** — hub session stale/unreachable + operator login required. Operator must (1) rebuild+serve the web container off HEAD `6e71e127` on the gateway :80, and (2) perform the persona logins, then the browser QA can run. Two personas needed: (a) an **approval**-kind active stage to verify the two actions ("Assinar e aprovar" + "Solicitar mudanças"; no "Assinar e devolver"); (b) a reviewer who is **also the next-stage approver** to verify "Aprovar já". Until then, closure is code-complete + L0/L1-green but the mandatory browser evidence is outstanding.

## 9. Commits (branch `claude/amazing-saha-589663`, NOT pushed)

- `1c6db8dc` — S-A/B: regen G3 api-types + remove decision preselection.
- `9418d091` — S-C: two-action approval footer (sign+approve or request_changes).
- `6e71e127` — S-D: fast-forward "Aprovar já" one-gesture ceremony.
