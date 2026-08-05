# HTTP Surface Protocol — adversarial review ledger

Artifact: `docs/superpowers/specs/2026-08-05-http-surface-protocol-design.md`
Reviewer arm: `gpt-5.6-sol` / medium, read-only, via `harness:codex-dispatch`.

---

## Round 1 — 4 BLOCKER, 5 MAJOR, 2 MINOR. VERDICT: DO NOT PROCEED.

Every finding was verified against source before disposition (`adversarial-review` §5,
symmetric duty). Dispositions below.

| # | Sev | Claim | Verified | Disposition |
|---|---|---|---|---|
| 1 | BLOCKER | `route_resolve` chain link violates locked constraint 6; an ADR cannot override an operator lock | **True.** `…-system-impact.md:205` — "The middleware chain (`chain.go:25`) is not touched." Constraint 0 says a constraint that appears to require otherwise is a contradiction to *surface*, not resolve locally. `chain.go:64` also shows login limiting already identifies login without any surface lookup. | **applied — root cause, not patch.** §6 deletes the link entirely. Each of the three consumers resolves on demand; the login limiter needed nothing. This also kills finding 9 (no context key ⇒ no trust boundary to argue) and the second ADR (§12). Accepted cost stated: two extra `mux.Handler(r)` calls per request. |
| 2 | BLOCKER | Step 5 never proves old and new tier-1 agree, though step 6 relies on it | **True.** `assertSurface` compares *keys* only; the gate demands exhaustive `(Capability, Visibility)` parity (`…-system-impact.md:134,140`, locked constraint 5). | **applied.** §10 gains an exhaustive parity property; §11 makes it its own step 5, gating step 6, deleted in step 6. |
| 3 | BLOCKER | The E2E rule overlay *is* the exception mechanism ruling C deleted | **True.** `router.go:33-38` + `main.go:839-848` — e2e is mounted after `buildRouter`, deliberately outside route truth. Merging a conditional overlay into `httpSurface` makes surface membership env-dependent. | **applied.** §5 moves the scaffolding to its own listener with the same chain. Governed surface stays total. Cost stated: 3 CI/Playwright configs gain a second base URL. |
| 4 | BLOCKER | Migration omits `/metrics`; counts already-qualified presence as legacy | **True on both.** `/api/v1/metrics` is a bare `mux.Handle` at `router.go:123-125`; `iam/presence/handler.go:73` is already method-qualified and is excluded from codegen deliberately (WebSocket). | **applied.** §8 replaces the presence row with observability, states 12 bare patterns / 15 operations, and states why `streamPresence` staying hand-mounted is not an exception (§5 asserts patterns, not provenance). |
| 5 | MAJOR | Unresolved routing turns authenticated 404/405 into 500 | **True.** `middleware.go:90-96` 500s any unrecognized visibility before reaching the mux; `method_not_allowed` is downstream (`chain.go:39`). The draft's own prose claimed the opposite. | **applied.** §6 splits the two cases: no pattern → `VisibilitySessionRequired` (byte-identical to today), pattern-without-rule → new `VisibilityUnresolved` landing in the existing `default:` 500. `middleware.go` needs no edit. |
| 6 | MAJOR | One-commit guard gap between steps 4 and 5; `TestRouteCoverage` removal undated; step 1 omits full regeneration | **True.** `permissions_test.go:369-481` consumes the types step 4 deletes. `.github/workflows/api-contract.yml:29-53` enforces all-artifact regeneration. | **applied.** §11 merges the publisher list + assertion + guard deletion into one commit; step 1 and step 3 now say full regeneration explicitly (locked constraint 7). |
| 7 | MAJOR | "~28 routes rely on the fallback" is false | **True, and independently confirmed twice** — my own probe (148 mounted, 135 covered, 0 fell through, 13 bare) and Codex's static first-match over all 147 operations. The number was fabricated. | **applied.** §12 risk row rewritten: the fallback is reached **zero** times; the real authoring risk is transcription, caught by the §10 parity test. §1 row count corrected 119 → 120. |
| 8 | MAJOR | `x-authz-visibility` restates what `security` + capability presence already say | **True.** Root `security: - sessionCookie: []` (`openapi.yaml:12-13`) + 4 `security: []` overrides already give two tiers; capability presence gives the third. The analysis left this explicitly open (`:130`). | **applied, and strengthened past the finding.** §2 drops `x-authz-visibility` and its two agreement-validation rules. Pure derivation would make "session-required" a silent default reachable by forgetting a field, so the third tier is stated by `x-authz-capability-none: '<reason>'`, mirroring the live `x-authz-area-none` convention (`openapi.yaml:1705`). Three mutually exclusive markers; none restates another; a missing marker is a generation failure. |
| 9 | MAJOR | Context-carried rule has no trust boundary or ownership model | **True of the draft.** | **closed by dissolution.** Finding 1's fix removes context storage entirely. |
| 10 | MINOR | Naming the role `Module` makes platform handlers look like bounded contexts | **True.** `CLAUDE.md:35` reserves the word for the 15 modules; health/observability/configuration are `internal/platform/*`. | **applied.** Renamed `httprouter.SurfacePublisher` throughout. |
| 11 | MINOR | Byte-exact key claim rests on three simple paths | **True** — and the reviewer found no actual mismatch, checking two-parameter paths (`templates/api/api.gen.go:1629-1638`, `iam/api/api.gen.go:1699-1700`). | **applied.** §3 keeps the spot checks but labels them as such; §10 adds an exhaustive generator-level comparison over all 147 operations. |

**Author-originated findings folded into the same revision** (found before round 1 returned, held
to avoid a moving target):

| Item | Disposition |
|---|---|
| §8's "migration is forced by the design" argument | **withdrawn.** Method-qualifying alone would satisfy the assertion — each legacy handler is 1:1 with one method. The conclusion stands on the extermination ruling instead: two mount regimes kept alive because one is cheaper is exactly the banned mask. |
| Dead spec extensions `x-rate-limit`, `x-websocket-message` | **added to §7 deletions.** No consumer; the real rate-limit config is `ratelimit/config.go:35`. `x-authz-area` stays — live, pinned by `permissions_authz_scope_test.go:76`. |
| `if r.Method != X` guards in migrated legacy handlers | **added to §7 deletions** — dead branches after method-qualified mounting. |
| Locked constraint 9 (15 new module→iam edges) never answered | **answered in §4:** zero new edges. Capability never enters a module; publishers declare routes, the root declares policy. |

**Convergence:** round 1, 11 findings, altitude = design-level (violated lock, missing gate,
self-contradicting exception, behavior change on a live path). Not converged. Round 2 dispatched.

---

## Round 2 — Job 1: 10 CLOSED, 1 PARTIAL. Job 2: 1 BLOCKER, 5 MAJOR, 2 MINOR. VERDICT: DO NOT PROCEED.

The PARTIAL (finding 2, parity) is re-raised as round 2's BLOCKER and dispositioned there.

| # | Sev | Claim | Verified | Disposition |
|---|---|---|---|---|
| 1 | BLOCKER | The parity gate *samples* value-dependent rules and so cannot license deleting `routeRules` | **True, with a concrete instance.** `permissions.go:120` carries `notSuffix: "/roles"`, so `PATCH /api/v1/iam/users/roles` is disqualified from the `user.manage` row and falls through to session-required today, while the pattern lookup gives it `user.manage`. One sentinel per `{param}` never visits that point. | **applied — the fix is exhaustiveness, not more samples.** §10 derives the candidate set from `routeRules` itself: the sentinel plus every literal token any rule carries in `pathExact`/`pathSuffix`/`contains`/`notSuffix`. `matches` can only distinguish paths through those literals, so the set covers its whole decision space. Tokens are enumerated from the table, not hand-listed — a hand-listed set inside this test would be the very defect the program removes. |
| 2 | MAJOR | Tag coverage does not enforce ownership: publishers can mount each other's operations and all checks pass | **True.** One global pattern set makes the three checks satisfiable by any partition of the routes. | **applied.** The recorder is now **per publisher**; `surfaceRule` carries `tag`; `assertSurface` gains check 4 (pattern's declared tag == mounting publisher's `Tag()`). |
| 3 | MAJOR | The panic probe proves nothing and is never executed | **True.** `GET /internal/test/panic` has exactly one occurrence in the repo — its own registration (`main.go:845`). No workflow, script, or Playwright flow requests it. | **applied.** Deleted, not moved. REQ-MW-1 moves to a Go test over the composed handler, where it can actually fail. Catalog §22: a guarded branch no environment executes. |
| 4 | MAJOR | Sequencing breaks `/healthz` before its consumers move | **True.** Migrating health removes the bare `/healthz` mount (`health.go:20`); keeping it fails step 4's check 2, deleting it strands the consumers until step 7. | **applied.** `/healthz` deletion + consumer repointing folded into step 3. §11 now carries an explicit dependency graph (`1→2`, `3→4`, `2→4`, `7→4`, `4→5→6`) instead of a numbered list. |
| 5 | MAJOR | Six-family inventory is 12 operations, not 15; health is not 1:1 with GET | **True on both.** The tag inventory sums to 12 (auth 4, security 3, health 2, search 1, configuration 1, observability 1). `health.go:23,32` check no method at all. | **applied.** Count corrected. The health change (authenticated non-GET: 200 → 405) is now an explicitly approved delta in §8 with a test in §10, not an accident discovered at QA. |
| 6 | MAJOR | `VisibilityUnresolved` hides in the unknown-value `default:` | **Already closed before the round — disproved.** The revision states the explicit case at `…-design.md:321-326`; the reviewer anchored on `:317` and stopped one paragraph short. | **not applied as reported; one clause added** — the text now also says the `default:` arm is retained for genuinely unknown values, and §10 tests both arms. That part of the finding was real. |
| 7 | MINOR | The extermination pass proposes *repointing* a Playwright config that is itself dead | **True.** `playwright.approval.config.ts` is selected by no workflow or package script; its `START_BACKEND` knob is set by nothing and its branch invokes a nonexistent `./cmd/api`. | **applied, and it is the sharpest finding of the round.** Repointing a dead file would have been the pure form of the defect this program removes: maintaining a copy of route truth inside an artifact nobody runs. Deleted; `/healthz` has four live consumers, not five. |
| 8 | MINOR | `surfaceRule` is never defined, yet `rule.allowedDuringPasswordChange` is consumed | **True.** | **applied.** §3 defines the struct with `visibility`, `capability`, `tag`, `allowedDuringPasswordChange`. |

**Author-originated correction folded in:** the second-listener answer for the e2e scaffolding was
rejected on cost once its consumers were actually read — five endpoints called from eight
Playwright sites as relative paths against the page `baseURL`. Replaced by an `internal-e2e`
publisher whose declaration is generated from a separate spec document, excluded from the public
bundle and FE codegen. Publisher and tag enter and leave `specTags` together, so all four checks
stay total on both boot paths. This is not an exception: an exception is *mounted but not
checked*.

**Convergence:** 11 → 8 findings. Altitude dropped from "violates a lock / has no gate" to "the
gate is unsound at one point / a count is wrong / a struct is undefined". Four of the eight were
arithmetic or definition-level. Round 3 dispatched; if it returns only mechanical findings the
loop stops there.
