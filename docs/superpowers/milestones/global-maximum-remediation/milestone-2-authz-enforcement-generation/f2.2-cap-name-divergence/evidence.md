# F2.2 evidence — tier-1 ↔ tier-2 capability-name coherence

> Closes F2.2 against `../validation-contract.md §2`. Implemented by subagent (TDD pin), reviewed +
> gates independently re-run by the main session. **No route/capability behavior changed** — test +
> doc annotations only.

## Runtime verification (contract §2.1 premise — CONFIRMED in code, no HS-2)

Both named divergences are already code-resolved (ADR 0022 Phase 11 F4); the 2026-07-03 review read
stale follow-up lines. Source-traced by the main session:

| Site | Tier-1 (route→cap) | Tier-2 (`authz.Require`) | Agree? |
|---|---|---|---|
| force-release session | `permissions.go:157` → `CapMembershipManage` | `repository.go:798,828` → `CapMembershipManage` | ✅ `membership.manage` |
| approval-route management | `permissions.go:238-240` (`GET/POST/PUT /approval/routes`) → `CapRouteManage`, ordered **before** the generic `/approval/` fallback (`:244-246` → `CapDocumentSubmit`) | `route_admin_service.go:165,290,445,527` → `CapRouteManage` | ✅ `route.manage` |

The remaining coarse/fine `/approval/` differences (tier-1 `document.submit` gate vs tier-2
`document.signoff` on decision, etc.) are the **deliberate** PDP split — recorded as an intentional
written exception, not "aligned" (contract §2.1 non-goal).

## Regression pin (contract §2.2 T2)

`apps/api/cmd/metaldocs-api/permissions_test.go` → `TestTier1Tier2CapabilityCoherence_F4Sites`
(~66 lines, appended). Narrow: asserts the tier-1 resolver output equals the typed tier-2 const for
**exactly** the two reconciled routes; explicit doc-comment forbids broadening to a blanket
tier-1==tier-2 assertion. A tier-1 route-cap rename fails at runtime; a const rename fails at compile.

**POSITIVE (GREEN on HEAD) — independently re-run by main session:**
```
--- PASS: TestTier1Tier2CapabilityCoherence_F4Sites (0.00s)
ok  	metaldocs/apps/api/cmd/metaldocs-api	7.114s
```
**NEGATIVE (RED on divergence) — independently re-run by main session** (temporarily flipped
`permissions.go:157` force-release cap → `CapDocumentView`, then reverted via `git checkout`):
```
--- FAIL: TestTier1Tier2CapabilityCoherence_F4Sites (0.00s)
    permissions_test.go:717: force-release session: tier-1 must equal tier-2 CapMembershipManage:
    POST /api/v1/documents/d1/session/force-release: tier-1 cap="document.view" diverges from
    tier-2 cap="membership.manage" (ADR 0022 F4 regression)
FAIL
```
Revert confirmed clean (`git checkout -- permissions.go`; re-run GREEN). The implementing subagent
independently captured the same RED with a `CapDocumentEdit` flip — two distinct flips both bite.

## Doc-truth restore (contract §2.2 T3) — 4 files, annotations only, no deletions

| File | Change |
|---|---|
| `wiki/decisions/0022-authz-capability-coherence.md` | 4 back-annotations (≈lines 197, 251, 342, 451) marking the stale Phase 7/8 follow-up + F4 lines **RESOLVED in Phase 11 F4 (§349-351)**, each citing the new test. |
| `wiki/backend/_artifacts/stage1/module-iam.md` | force-release "known pre-existing defect" bullet annotated RESOLVED-in-Phase-11-F4 with file/test pointers. |
| `wiki/decisions/0018-approval-route-lifecycle.md` | §6 deferred `CapRouteView` split marked **SUPERSEDED by ADR 0022 Phase 11 F4** (never implemented; `route.view` never created; actual resolution = explicit `route.manage` tier-1 rows). |

Subagent verified already-correct (no edit): `wiki/concepts/approval-routes.md`,
`wiki/modules/approval-tech-debt.md` (BE-9 closed), `wiki/concepts/authz-tiers.md`. Left untouched as
out-of-scope (a *different*, still-open template-publish tier mismatch): `wiki/backend/legacy-register.md`,
`stage2/security-secrets.md`, `stage1/synthesis-legacy.md` — recorded, not silently skipped.

## Gates (independently re-run by main session)

| Gate | Result |
|---|---|
| `go build ./...` | clean, exit 0 |
| `go test ./apps/api/cmd/metaldocs-api/ -run TestTier1Tier2CapabilityCoherence_F4Sites` | PASS |
| `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | **0 violation(s)** — all 5 authz lints green |
| Net code diff | `permissions.go` + all `authz.Require` sites byte-identical; only a test + 4 doc annotations added |

## Bounded defers
None for F2.2. (The separate template-publish tier mismatch noted above is pre-existing, out of this
feature's boundary, and left for a future coherence pass — not introduced here.)

## Acceptance (contract §2.3) — MET
Both sites verified aligned · pin RED-on-divergence / GREEN-on-HEAD · 5 authz lints green · ADR 0022 +
affected wiki reflect the Phase-11 resolution · coarse/fine difference recorded as intentional.
