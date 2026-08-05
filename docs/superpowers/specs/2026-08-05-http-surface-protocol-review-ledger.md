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
