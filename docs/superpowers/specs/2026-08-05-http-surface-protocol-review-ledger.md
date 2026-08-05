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

---

## Round 3 — Job 1: 6 CLOSED, 1 PARTIAL, 1 OPEN. Job 2: 2 BLOCKER, 4 MAJOR, 2 MINOR. VERDICT: DO NOT PROCEED.

**Independent vote layer.** This round each load-bearing finding was additionally verified by a
sub-agent reading source, with no access to the reviewer's reasoning. Four verdicts came back, and
they did not simply ratify: two findings were **narrowed**, one was **tripled in scope**, one was
**confirmed including a fabrication in our own artifact**. The vote is recorded per finding below,
because a review protocol whose second opinion only ever agrees is not a second opinion.

**Job 1 dispositions (round-2 findings):** 1 OPEN, 2 CLOSED, 3 CLOSED, 4 PARTIAL, 5 CLOSED,
6 CLOSED, 7 CLOSED, 8 CLOSED. The OPEN (parity) and the PARTIAL (sequencing) are re-raised as
round 3's BLOCKER 1 and MAJOR 3 and dispositioned there.

| # | Sev | Claim | Verified | Independent vote | Disposition |
|---|---|---|---|---|---|
| 1 | BLOCKER | The parity gate still omits an entire matcher dimension — `matches` filters on **method** first, so HEAD, wrong-method, and methodless rows are unreachable when only paths are varied | **True.** `permissions.go:35` gates on method before any path predicate. Varying paths alone cannot reach a methodless row once `/healthz` is gone from the recorded set, and cannot reach the 405 case at all. | — (author-verified; the read is one file) | **§1 ratchet fired — restructured, not patched a third time.** This construct had been patched twice and BLOCKED three rounds, so §2 was run and its verdict written into §10.1. **Root cause: the gate compared two *tables*, but the new tier-1 is table + mux.** `mux.Handler(r)` is half the new function and it was outside the comparison; every sub-finding of this round (HEAD aliasing, method mismatch, methodless rows) is a consequence of that one omission. §10 is now a **request-level** differential — `oldCap, oldVis := newPermissionResolver()(r.Method, r.URL.Path)` against `newCap, newVis := newPermissionResolver(mux)(r)` — over a domain enumerated mechanically from `routeRules` itself (§10.2): every literal token any rule carries, plus a no-collision sentinel; and for each path, its own method, `HEAD` when that method is GET, and one method registered nowhere on that path. §2 outcome **(a) restructure now**. |
| 2 | BLOCKER | Check 3 is not total on the SQLDB-less boot path: the presence stream's mount is skipped, so a declared operation is unmounted and the server fails to boot | **Mechanism true, phrasing wrong.** `router.go:104-106` skips the mount when `h.presence == nil`, and `main.go:1173-1175` returns nil when `deps.SQLDB == nil`. | **PARTIAL.** The sub-agent disproved one clause: the finding says "the generated `streamPresence` declaration", but `streamPresence` has **no** generated Go declaration — it is excluded from server codegen. The mechanism survives; the sentence describing it did not. | **applied as a protocol rule, not a per-site guard.** §4 gains **"Mount is total — conditional mounting is unrepresentable"**: a publisher's `Mount` registers every route it owns, unconditionally, and availability is a handler-level `501`, never a routing-level absence. This is the repo's own existing majority convention (8 IAM operations already answer 501 on a nil dep; the presence stream was the single site doing the opposite), so the rule is a codification, not an invention. Three consequences recorded: `conditionalRouteFamilies` (`router.go:85`) is **deleted** — a held round-4 MAJOR dissolves rather than being patched; the SQLDB-less presence stream answers 501 instead of 404 (accepted, §12 risk row); and "excluded from codegen" ≠ "excluded from the surface". Also clarified: a tag may have more than one publisher (`iam` has two), so check 1 asserts *at least one* and check 4 asserts mounter-matches-tag — no bijection is required. |
| 3 | MAJOR | Step 4's recorder cannot see e2e routes mounted directly after `buildRouter`, so the numbered order leaves them outside the asserted surface for three commits — despite the doc's own graph saying `7 → 4` | **True, and it is a self-contradiction inside one section.** `mountE2EHandlersIfEnabled` runs after `buildRouter` (`main.go:839-848`); §11's prose numbered it last while §11's graph required it earlier. | — | **applied.** §11 renumbered so the order *is* the graph: the e2e publisher is step 4 and the presence fold is step 5, both before the assertion (now step 6). A numbered order that contradicts its own dependency graph is a defect, not a presentation choice — following it would have shipped an assertion reporting completeness over a surface it could not see. §11 also now answers the executability question explicitly: what is broken between each pair of commits, and for how long. |
| 4 | MAJOR | `internal-e2e.yaml` is cleanly excludable, but consequently gets **no** lint or drift gate — and the claimed `make openapi-verify` target does not exist | **True, and the fabrication is ours.** | **CONFIRMED, including the fabrication.** The sub-agent read the `Makefile`: five targets (`up down logs test test-watch`, `Makefile:3`), none OpenAPI-related. Our own design doc invented the target. | **applied, and recorded as the program's fourth fabricated fact** (after "~28 fallback routes" → 0, "119 rows" → 120, "15 spec operations" → 12). §3 now names the real gate — `.github/workflows/api-contract.yml:37-38`, `go generate ./...` then `git diff --exit-code -- '**/api.gen.go'` — and states the hole: that pathspec matches only files named exactly `api.gen.go`, so a generated file under any other name is regenerated and then never checked. §11 step 2 widens the pathspec; §5 adds the two remaining gates (api-lint over the second document, generator validation treating it as a first-class input). The exclusion that stays is the public bundle and the FE types, which is the point of the second document. |
| 5 | MAJOR | The generated per-operation `allowedDuringPasswordChange` changes live authorization for `HEAD /api/v1/auth/me` | **True for that one operation.** `middleware.go:131-135` gates on exact `GET`; a `GET` mux pattern absorbs `HEAD` (`server.go:2484-2486`), so the boolean is inherited. | **NARROWED.** The finding presented `/auth/login` and `/auth/logout` as instances of the same hazard. The sub-agent showed both clauses are `MethodPost`, and a POST pattern does not absorb HEAD — they are **structurally immune**. Only `getCurrentUser` is exposed. | **applied at the corrected scope.** §2 carries a delta table naming `getCurrentUser` as the sole affected operation, with `logout` and `changePassword` explicitly marked immune and why. Accepted with a regression test and a §12 risk row: HEAD carries no body, so nothing reaches a principal who must change their password. Scoping a finding down is not softening it — an over-broad finding applied over-broadly produces guards on paths that cannot fail. |
| 6 | MAJOR | An off-spec, unwired HMAC endpoint and its tests survive the extermination scope | **True.** `pdf_webhook_handler.go:28` concedes in its own header comment that it is UNWIRED. | folded into the vote on 7 | **applied.** Deleted outright, tests included. Contract-first makes an off-spec route illegal and the extermination ruling makes "documented as dead" not a resting state — a dead HMAC endpoint is a security surface that *looks* maintained. |
| 7 | MINOR | The deletion manifest misses production-dead sub-handler registration APIs, and a comment that becomes false when `TestRouteCoverage` is deleted | **True, and severely under-reported.** | **TRIPLED.** The reviewer named 4 dead registration methods. The independent sweep found **13**, plus the complete unwired endpoint of finding 6. | **applied, and promoted from MINOR to the clearest illustration in the document.** New §7.1 tables all 13 with their live replacement mount. They are one defect with one cause: **each past codegen migration added the generated `HandlerWithOptions` mount and left the hand-written mounting method in place.** The codebase says so about itself — `iam/.../router.go:102-113` names the exact five call sites `RegisterGenerated` replaces, and all five are still there. Why it belongs to *this* program: under §5 a method that mounts nothing cannot be reached by any check the protocol adds, so it would survive the restructure invisibly — which is precisely how it survived the last three migrations. The `muxer.go:11-14` comment is rewritten in the same commit as the deletion it describes, alongside the three already-known false comments. The sweep also **cleared** `mountE2EHandlersIfEnabled`/`METALDOCS_E2E` as not dead (`e2e_gate_test.go` drives both branches, CI sets the variable) — recorded so it is not re-litigated. |
| 8 | MINOR | §12 says "ADR required — one, not two" and then names a second ADR | **True.** A self-contradiction two sentences wide. | — | **applied.** §12 now states an unambiguous **two-ADR deliverable** in a table: an amendment to ADR 0022 (tier-1 mechanism only; tier-2 and the tripwire untouched) and a new ADR for the HTTP surface protocol as a platform framework (15 delivery seams — different owner, different lifetime). No third ADR for the middleware chain: §6 dropped the `route_resolve` link, so `chain.go:25` is byte-identical afterwards. |

**Convergence:** 8 → 8 findings. Count flat, but **altitude dropped decisively**: the round's own
content is grep-findable dead code (13 methods), a fabricated Makefile target, a self-contradictory
ADR paragraph, and a step-ordering slip that the document's own graph already refuted. Every one is
in the band §8 names as the stop signal — *"a compiler, generator, or lint would catch them."* The
two BLOCKERs are the exception that proves it: both were re-raises of the same construct, and both
were closed by the **one** structural insight (the mux is half the new tier-1) rather than by eight
separate fixes.

Against §8's stop conditions: findings are at rung 1–3 mechanical altitude, and the same-altitude
recurrence that would mandate escalation did not occur — altitude fell every round (design →
soundness → mechanical). **The loop stops here.** Round 4 is not dispatched; the remaining class is
what the build, the generator, and `api-lint` catch, and dispatching another design round would be
generating scope rather than finding defects.

**Verdict handling.** Round 3 closed on DO NOT PROCEED, and §8 forbids declaring the residual risk
acceptable without naming who accepted it. The residual is therefore **not** declared acceptable
here. Both BLOCKERs were closed by restructure inside this same revision — the artifact the verdict
judged no longer exists — so the honest statement is: *the verdict was rendered against revision 2,
and both of its blocking findings are structurally closed in revision 3.* Whether that closure is
sufficient to leave the design phase is the **operator's** call at the brainstorming user-review
gate, not the author's. That gate is the acceptance record.

**Two MAJORs held from an earlier program remain held**, and must not be patched in the interim:
the `main.go:817-837` keyed literal (still open), and `conditionalRouteFamilies` fail-open
(`router.go:85`) — **now dissolved** by §4's Mount-is-total rule rather than patched, which is the
outcome §1 step 2 asks for.
