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

---

## Spec self-review (brainstorming checklist step 7) — 7 defects, all fixed inline

Run by an independent sub-agent over the committed revision-3 document with a scan-only brief
(placeholders, internal contradictions, stale self-references, ambiguity, plus three factual
anchors). It is not a design review and raised no design finding — which is the expected result
if the adversarial loop actually converged. It did raise seven **internal-consistency** defects
that three rounds of adversarial review had not, because none of them is a design flaw: they are
the artifact contradicting itself.

| # | Class | Defect | Fix |
|---|---|---|---|
| 1 | contradiction (count) | §5 said 137 generated + 9 legacy + `streamPresence` = 147, while §8 enumerates **12** legacy operations. 137 + 9 + 1 = 147 is internally consistent but incompatible with §8; the arithmetic forces 134 + 12 + 1. | §5 corrected to 134, and it now cites §8's per-family breakdown rather than restating a number. |
| 2 | contradiction (count) | §12's risk header said "three client-visible behavior changes" and then listed four. | Corrected to four, each lettered. |
| 3 | ordering | §5 assigned the `internal-e2e.yaml` lint/validation gates to §11 **step 2**, but the document does not exist until **step 4**. A gate authored before its subject passes vacuously. | Split: lint + generator validation → step 4; drift pathspec → step 2. The reason is stated in the doc, not just the assignment. |
| 4 | stale reference | §4 said the presence 501 delta is "listed in §8, tested in §10". §8 is the legacy-family migration and never mentions presence; §10 had no row for it. | Pointer corrected to §12, **and the missing §10 test row was added** — boot with `presence == nil`, assert 501 + `problem+json`. The stale pointer had concealed a genuinely absent test. |
| 5 | contradiction | §3 said "**exactly one file**" of output, while §5 requires a second generated declaration from `internal-e2e.yaml`. | §3 rewritten: one file **per input document**, both under the same drift gate. |
| 6 | uncounted rule | §4 required the generator to reject multi-tag operations, but §2's validation list had four rules and §10's test row said "four failures" — the rejection was enforced nowhere. | Added as §2 rule 5 with its rationale (check 4 is only well-formed if the tag is single-valued); §10's caption updated to five. |
| 7 | naming drift | §8's family table said `featureflags` (the Go package) where §4's tag inventory says `configuration` (the spec tag); the equivalence was never stated. | Both names given together, with the rule that the **tag** governs — check 4 compares tags, never package names. |

**Factual anchors re-verified:** `Makefile:3` — AGREES (five targets). `muxer.go:11-14` — AGREES
(quoted verbatim). `.github/workflows/api-contract.yml` — **DISAGREES**, and this is the program's
**fifth** fabricated-or-wrong citation: the doc labelled `:35` as "Redocly bundling" and `:62` as
`api-lint`. In fact `:35` is the backend codegen-drift step name, `:62` is the Redocly **lint**
command, the repo's own `api-lint` guard is at `:100`, and the workflow has **no bundling step at
all**. Corrected in §5. The drift-gate citation (`:37-38`) that the round-3 fix rests on was
re-checked and is correct.

**Why this pass found what three adversarial rounds did not.** The adversarial reviewer is briefed
to attack the *design*; a count restated two sections apart, or a pointer to the wrong section, is
invisible to that brief because neither changes what gets built. Both scans are needed and they do
not substitute for each other. Worth noting which class turned out to matter: defect 4 looked like
a typo and was in fact a **missing test** — the wrong pointer was the only visible symptom of it.

Six of the seven were arithmetic, pointer, or naming errors — rung-1 mechanical, exactly the
altitude §8 predicts at the end of a converged loop. This confirms the round-3 stop decision rather
than reopening it.

> **Superseded by round 4.** That last sentence is wrong. See "Correction to the round-3 record"
> below.

---

## Round 4 — the verdict round that round 3 should have been

**Why this round exists.** Round 3 was closed as converged. That call was wrong on both of §8's
axes and the operator caught it: the count went **8 → 9**, not down, and — the deeper error — the
DO NOT PROCEED verdict the loop closed on had been rendered against **revision 2**. Two revisions
had landed since. No round had ever judged the artifact that actually existed, so "converged" was
a statement about a document nobody had reviewed. The prompt for this round says so explicitly and
instructs the reviewer to judge the current revision and to not withhold PROCEED merely because a
prior round withheld it — a non-terminating loop is itself a defect.

**Dispatch:** `gpt-5.6-sol` / `medium`, OS-process, read-only. Artifacts: `agent__r4.log`,
`agent__r4.last.md`, prompt `prompt-r4.md`.

### Job 1 — round-3 dispositions (reviewer's own, verbatim)

| # | Disposition | Anchor |
|---|---|---|
| 1 | **OPEN** | `permissions.go:35` — §10.2 still samples only one wrong method although `matches` distinguishes every method-qualified rule |
| 2 | CLOSED | design `:251` — unconditional mounting + the step-5 presence fold makes declared ⊆ mounted total |
| 3 | CLOSED | design `:819` — E2E publication now precedes the recorder/assertion |
| 4 | CLOSED | design `:388` — step 4 adds both lint targets, step 2 widens the real drift gate |
| 5 | CLOSED | design `:123` — the sole HEAD authorization delta is accepted and has a regression test |
| 6 | CLOSED | design `:561` — the unwired HMAC handler and its tests are explicitly deleted |
| 7 | **PARTIAL** | `iam/.../observability_handler.go:31` — at least two more production-dead registration APIs remain outside the manifest |
| 8 | CLOSED | design `:877` — the deliverable consistently requires two ADRs |

Six closed, one partial, one open. Both survivors are the **same two** findings the sub-agent vote
layer had independently flagged in round 3 — which is the useful signal: two independent reviewers
converging on the same two gaps is evidence the gaps are real, not evidence of reviewer
persistence.

### Job 2 — findings and author disposition

All nine anchors were verified against source before acceptance (§5 symmetric author duty). **All
nine hold.** None was disproved.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| 1 | BLOCKER | §10.2's method domain is not total: `openapi.yaml:3361` defines GET-only `/documents/{id}/view`; `PATCH` reaches the broad row `permissions.go:175` → `document.edit`, `POST` reaches the fallback `permissions.go:336` → session-required. Sampling *one* wrong method hides the other. | **applied** — §10.2 now enumerates every distinct `routeRules` method (GET 42 / POST 53 / PUT 11 / PATCH 5 / DELETE 8) plus one no-collision method, and mandates the **full cross-product** of tokens × recorded patterns |
| 2 | BLOCKER | The boot assertion rejects the design's own topology: §4 `:285` permitted two publishers for `iam`, §5 `:325` check 1 rejects exactly that. | **applied** — the `:285` clause is **deleted**, not the check. §11 step 5's presence fold is what makes one-publisher-per-tag true before step 6 asserts it |
| 3 | MAJOR | The generated e2e descriptor was assigned a nonexistent `METALDOCS_E2E` **build tag**. Mounting is a runtime env branch (`main.go:155-166`); the only e2e tags are `integration && !production` (`internal/test/e2e_seed.go:1`) with a stub (`e2e_seed_stub.go:1`) — a different axis. | **applied** — both descriptors are untagged files in `package main`; `activeSurface()` selects at runtime with the **same predicate** the mounts use, which is what keeps check 2 total on both boot paths |
| 4 | MAJOR | `/healthz` is a second omitted **authentication** delta. Public today (`permissions.go:85`, `health.go:20`); after deletion the empty pattern is session-required and authn rejects at `auth/.../middleware.go:78`. §10.3 claimed "no delta to have". | **applied** — the row now states unauthenticated **200 → 401** and explains the error: authn runs **before** the mux, so deleting the route does not make the request unobservable. §12's count corrected **four → five** |
| 5 | MAJOR | Two more production-dead registration APIs: `iam/.../observability_handler.go:31` (superseded by `router.go:205-223`) and `documents/.../handler.go:144` (reachable only via the already-dead `module.go:157`). | **applied** — §7.1 is **13 → 15**, with both added and their callers enumerated |
| 6 | MINOR | The 501 accounting says eight and names nine. | **applied** — nine `writeIAMNotImplemented` sites: **eight** nil-dependency guards, **one** unconditional deprecated (`router.go:248`), counted separately as a different mechanism |
| 7 | MINOR | The deletion manifest double-counts the PDF webhook (separate row + row 13 of the table). | **applied** — one row: "15 orphaned methods, including `pdf_webhook_handler.go:41`, whose whole file and test go too" |
| 8 | MINOR | `:690` says `assertSurface` has "three checks" after four are defined. | **applied** — four, one negative test named per check |
| 9 | MINOR | The dependency graph omits `2 → 4`. | **applied** — edge added; 3/4/5 are parallelizable with each other, 4 is **not** parallelizable with 2 |

**VERDICT (round 4): DO NOT PROCEED** — the non-total parity gate cannot license deletion of
`routeRules`. Nine of nine applied; nothing deferred, nothing disproved.

### §2 judgement — is finding 1 a third patch, or a completion?

Findings 1 and 2 both landed on §10's parity gate, the third consecutive round to do so. §1's
two-patch ratchet therefore **fires**, and §8's same-altitude-recurrence rule points at an operator
decision. A sub-agent vote was dispatched to answer the question the ratchet asks — *is the
enumerated domain total, or is the structure wrong?* — with a per-dimension proof.

**Answer: TOTAL, with the method dimension as the only hole.**

| Dimension | Totality argument |
|---|---|
| `pathPrefix` | **Provably** unflippable. All 47 prefix rules end at a `/` **before** the first `{param}`, and `{param}` segments cannot contain `/`. So `HasPrefix` reads only the static head, which no token substitution can alter |
| `pathExact` | Finite; every literal is a token in the enumeration |
| `pathSuffix` / `contains` / `notSuffix` | Structure-determined except for **whole-token collisions**, which the cross-product generates by construction |
| `method` | A **closed finite set** — five values in the table, plus HEAD, plus one no-collision method |

Altitude therefore **did** drop between rounds 3 and 4: the method dimension went from *absent* to
*sampled at 1 of 6*. That is not same-altitude recurrence; it is an incomplete restructure being
completed. Outcome **(a) restructure now**, and it is the same restructure round 3 began — not a
third patch on top of it. No operator escalation is required, and none is claimed.

**One implementation caveat the vote surfaced, and it is load-bearing.** The totality proof holds
**only** for the full cross-product. The narrower construction — substitute each token only into
the pattern its own rule targets — reads identically in English and is **unsound**. The
counterexamples are real: a document id whose literal value is `approval-preview` makes
`GET /api/v1/documents/approval-preview` satisfy `permissions.go:163`'s
`pathSuffix: "/approval-preview"` (→ `document.submit`) while the mux routes it to the shorter
`/documents/{id}` pattern (→ `document.view`). This is now stated in §10.2 as a mandate, because a
future maintainer optimizing the loop would silently reintroduce the hole.

### What the sub-agent vote layer bought, across rounds 3 and 4

Dispatched as an independent verification tier over the reviewer's findings, not as a second
reviewer. It narrowed two findings, tripled one, confirmed one **including a fabrication of mine**,
refuted one of my own claims outright, and corrected three factual classifications
(`admin_handler.go:132`, `people_handler.go:54`, `pdf_webhook_handler.go:41` are **zero-caller**,
not "test only").

**The highest-yield lesson is about prompt shape, not about review.** My round-3 vote prompt handed
the reviewer *a fixed list of 13 methods to verify*. A prompt of that shape can only confirm or
deny what is already listed — it is structurally incapable of extending the list. Asked instead to
**sweep**, the same reviewer returned two more (finding 5). The missed methods were a defect in my
prompt, not in the reviewer. Same class as the artifact's own subject matter: an enumeration handed
to a checker instead of derived by it.

### Correction to the round-3 record

The paragraph closing the self-review section — "this confirms the round-3 stop decision rather
than reopening it" — **is wrong and is superseded here.** The self-review found only mechanical
defects because it was briefed to scan for mechanical defects; that is not evidence the design
search was exhausted. Round 4, briefed adversarially, returned two BLOCKERs. Altitude is judged
from a brief that could have found design defects, never from one that could not.

### Convergence line

Count 9 (2B / 3M / 4N) → all nine applied. Altitude: one dimension-completion (rung 4) + eight
mechanical (rung 1–2). Round 5 is dispatched as the **verdict round on the current revision** — the
round the operator asked for, and the one round 3 skipped.

---

## Round 5 — the verdict round, and the §2 escalation it forces

**Dispatch:** `gpt-5.6-sol` / `medium`, OS-process, read-only, against HEAD `a287e6de`. Artifacts:
`agent__r5.log`, `agent__r5.last.md`, prompt `prompt-r5.md`. The prompt made two obligations
binding: judge the revision at HEAD, and emit a **proceed map** — a per-step license, because a
bare DO NOT PROCEED at round 5 is not an acceptable output.

### Job 1 — round-4 dispositions

Seven CLOSED (2, 4, 5, 6, 7, 8, 9), one PARTIAL (3), one OPEN (1). The reviewer independently
re-verified the five method counts and confirmed them exact, and swept for a sixteenth orphan and
found none — so finding 5 is genuinely closed, not merely acknowledged.

### The proceed map — what the operator asked for

| Step | Status | Reason |
|---|---|---|
| 1 — contract annotations + full regen | **LICENSED** | does not depend on the parity proof |
| 2 — generator + widened drift gate | **LICENSED** | follows step 1 |
| 3 — 6 legacy families + `/healthz` repoint | **LICENSED** | independent; repoint is atomic |
| 4 — e2e publisher | BLOCKED BY MAJOR 2 | activation + total-mount semantics not executable as written |
| 5 — presence fold into IAM publisher | **LICENSED** | independent |
| 6 — SurfacePublisher + recorder + assertSurface | BLOCKED BY MAJOR 2 | assertion not total on both e2e boot paths |
| 7 — parity gate | BLOCKED BY BLOCKER 1, MAJOR 4 | parity not exhaustive; deltas not enumerated |
| 8 — flip resolver + all deletions | BLOCKED BY 1, 3, 4, 5 | no total parity license; test coverage loss; docs deliverables unclosed |

**VERDICT (round 5): DO NOT PROCEED — steps 1–3 and 5 are licensed; escaped-separator parity must
become total before step 7 can license deletion of `routeRules`.**

Half the program is executable today. That is the first actionable verdict the loop has produced,
and it is the direct product of demanding a per-step map instead of a single word.

### Job 2 findings, verified

| # | Sev | Finding | Verification |
|---|---|---|---|
| 1 | BLOCKER | `GET /api/v1/documents/foo%2Fapproval-preview` matches the mux's `{id}` pattern (→ `document.view`) while the old resolver, reading **decoded** `r.URL.Path`, sees the suffix `/approval-preview` (`permissions.go:163`, → `document.submit`). The cross-product basis generates slash-free tokens only and cannot reach this class. | **CONFIRMED at the Go source.** `ServeMux.findHandler` matches on `r.URL.EscapedPath()` (`$GOROOT/src/net/http/server.go:2659-2680`), where `%2F` stays inside one segment; `r.URL.Path` is decoded, where it becomes a real separator. The two sides **read different strings**. |
| 2 | MAJOR | The e2e surface passes all four checks on neither boot path. | **CONFIRMED, independently and in more detail, by sub-agent vote 2** — see the merged §2 entry below. Same root cause. |
| 3 | MAJOR | Moving the 15 orphans' tests "to business methods directly" loses generated-wrapper and mux coverage — `module_wrapper_test.go:176,184`, `handler.go:169` — including typed path validation, method dispatch and `r.Pattern`. | Accepted. Direct calls are sound only for business-only assertions; routing assertions must be retained through the generated `HandlerWithOptions` mount. |
| 4 | MAJOR | The "five client-visible changes" count omits the general HEAD authorization tightening and first-match ordering collisions. | Accepted, and it **merges into BLOCKER 1's restructure** — see below. |
| 5 | MAJOR | §11's graph has no nodes for the two ADRs, the three wiki updates, or the `BaseURL` promotion, yet §12 lists all three as deliverables. | **CONFIRMED** — grep over the §11 graph region returns no ADR / wiki / BaseURL node. |
| 6 | MINOR | `iam/.../router.go:102-107` says it replaces "the six `RegisterRoutes` call sites" and then names **five**, omitting observability. | **CONFIRMED.** Sharper than reported: the same comment also asserts the swap "changes no tier-1 behavior" because tier-1 "keys off `r.Method`/`r.URL.Path`, not off mux dispatch mechanics" — which is precisely the premise this program deletes. The comment must be rewritten, not just recounted. |

### Sub-agent vote 2 — the e2e claim is FALSIFIED (author's own round-4 fix #3)

Dispatched to falsify §3's claim that the runtime `activeSurface()` keeps the four checks total on
both boot paths. It did.

- Production builds carry **no build tags** — `deploy/docker/api.Dockerfile:6` runs plain
  `go build`, and the repo contains no `-tags production` anywhere. So `!integration && !production`
  holds and the **stub** (`e2e_seed_stub.go:11`, a no-op) is what ships.
- `E2EEnabled()` (`internal/test/e2e_enabled.go:14`) reads **only** the env var. No build tag.
- Therefore `METALDOCS_E2E=1` on a production binary: `e2eHandlersEnabled()` → true → the e2e
  publisher enters the list → `activeSurface()` returns the **merged** table declaring 5 e2e
  patterns → but the stub mounts **zero** → check 3 (declared ⊆ mounted) fails → `log.Fatalf` →
  **the server refuses to boot.** A leaked env var becomes an unrecoverable boot loop.

**Root cause (§1 step 1):** the declared side is gated by **one** predicate (env) and the mounted
side by **two** (env AND build tag). §3's sentence "the table the assertion reads and the mounts it
compares against are selected by the same predicate" is false in exactly the cell production
occupies. Round 5's MAJOR 2 is the same defect reached from the other direction.

**Two more, from the same vote:** §5's code snippet literally passes bare `httpSurface`, not
`activeSurface()` — which breaks even the matched case. And nothing at **generation** time forbids
a key collision between the two spec documents; the only guard is a boot-time panic.

**Fix (§1 step 2 — what makes it impossible), not yet applied:** one boolean governs both sides.

```go
e2e := e2ePublisher()                       // build-tag-selected pair: real publisher, or nil
useE2E := e2e != nil && e2eHandlersEnabled()
surface := httpSurface
if useE2E {
	publishers = append(publishers, e2e)
	surface = mergedSurface(httpSurface, httpSurfaceE2E)
}
assertSurface(mounted, surface, specTags, publishers)
```

`e2ePublisher()` is a tag-selected pair mirroring `internal/test/e2e_seed{,_stub}.go` — a shape the
repo already proves — but with **total** tags (`integration && !production` /
`!integration || production`), which also closes a latent gap the existing pair has: a literal
`-tags production` build satisfies neither of its two files and would not compile. Plus generator
validation rule 6: a method+path collision between the two spec documents is a **build** failure,
not a boot panic.

### Sub-agent vote 1 — the cross-product is tractable, with one ambiguity

Measured, not estimated: **80** distinct tokens (`pathExact` tails 38, `pathSuffix` 35, `contains`
7, `notSuffix` 0 new), **148** mux patterns (134 generated + 14 hand-registered; cross-checked
three ways against the spec's 147 operations, the 148th being the bare `/healthz`), and the five
method counts **confirmed exact** (GET 42, POST 53, PUT 11, PATCH 5, DELETE 8, plus **1 methodless
row** the design had not counted — `/healthz`, `permissions.go:85`).

Wall-clock at a padded 20 μs/iteration: **~1.2 s** literal, **~1.9 s** as §10.2 reads, **~20 s** on
the most adversarial reading. Tractable under every reading; no reduction needed. Round 5's "86
tokens / 82,563 cases" is the same order of magnitude reached by a slightly different extraction.

**One real ambiguity the vote surfaced:** §10.2 says "instantiate every `{param}` position in
turn", and **20 of the 148 patterns have two `{param}` positions**. One-at-a-time vs a joint
cross-product differ by ~10× in case count. The text does not say which. §10.2's own rule — "its
totality is re-verified against the code, not inherited from this paragraph" — makes this the
author's defect to close before the test is written.

### §2 — ESCALATION. The ratchet has fired twice; this is an operator decision.

**Round 5 is the fourth consecutive round to land on §10, and the third at the same altitude.**
Round 3: the method dimension is absent. Round 4: the method dimension is sampled, not enumerated.
Round 5: the **encoding** dimension is absent. Each finding is individually correct and each fix
was individually correct. That pattern is §1's explicit signal: *three findings on one construct is
a structural signal, not a quality signal.* §1's two-patch ratchet has now been exceeded, and §8's
same-altitude-recurrence rule says stop and escalate. **I am not applying a third basis extension.**

**Q1 — What is the global-maximum structure?**
Not a bigger enumeration. The parity gate is trying to prove *behavioral equality between two
implementations that read different inputs*: the old resolver classifies `r.URL.Path` (decoded);
the router routes `r.URL.EscapedPath()`. On that input space, equality is neither achievable nor
desirable — where they differ, **the old resolver is the wrong one**, because it classifies strings
the router never routed. Adding dimensions to chase equality is chasing a bug into its corners.

The global maximum is **a row-level derivation proof plus an enumerated delta list**, replacing the
request-level differential:

1. **Derivation** — each of the 120 `routeRules` rows is mapped to the generated table and compared
   row-by-row. That comparison is finite, total, and complete **by construction** (120 rows), with
   no input-space enumeration anywhere in it.
2. **Delta list** — every row that does not map identically becomes a named, approved delta with a
   regression test. Requests only the old resolver could classify (escaped separators, off-surface
   prefix matches, the methodless `/healthz` row) land here **automatically**, because the old
   resolver's domain is strictly larger than the router's.
3. **Boot assertion** — §5's four checks already prove mounted ≡ declared at every boot. That is
   the ongoing guard; the derivation proof is the one-time migration license.

This also absorbs round 5's MAJOR 4 (the delta count is incomplete): under a delta-list gate,
completeness of the delta list *is* the gate's output, not a claim made beside it.

**Q2 — What does a proven system do?** This is the standard policy-table migration: prove the new
artifact is **derived** from the source of truth and enumerate the intentional differences. It is
what every codegen migration in this repo already does (`oapi-codegen` + the drift gate), and what
§5 checks 2 and 3 already are. Fuzzing two implementations against each other is what you do when
you have no source of truth — but this program's entire premise is that the OpenAPI spec **is** the
source of truth.

**Q3 — Costs.**
*Global maximum:* rewrite §10 (one section of twelve); the deletion license then rests on row-level
derivation + the boot assertion instead of request equality. Steps 1–3 and 5 are unaffected and
stay licensed. *Local maximum (extend the basis a third time):* closes round 5 and buys nothing
structural — there is no argument that encoding is the **last** dimension. Trailing slashes, `..`
segments, `;`-parameters, unicode normalization, and duplicate slashes are all reachable and all
outside the current basis. **An enumeration-based proof over an infinite input space cannot be
closed by adding dimensions.** That is the sentence the last three rounds have been paying to
learn.

**Q4 — Who decides?** The operator. Recorded here, unresolved, blocking nothing that is licensed.

### Convergence line

Count 6 (1B / 4M / 1N), down from 9. Altitude did **not** drop on §10 — third same-altitude
recurrence — so §8's stop condition fires and the loop is **paused**, not closed. Everything else
did drop: findings 3, 5 and 6 are rung 1–2 mechanical, and the e2e defect is fully diagnosed with a
fix drafted. Steps 1–3 and 5 carry a clean license from an adversarial reviewer that was explicitly
invited to grant one. Nothing is deferred silently and no risk is declared acceptable by me.

---

## Round 6 — verdict round on the restructure (HEAD 53eff4a0, Sol / medium)

First round judging the derivation proof rather than the differential. Prompt explicitly barred
re-raising round 5's BLOCKER as an enumeration gap and redirected the attack to *decidability* and
*totality* of the new construction.

### Job 1 — disposition of round-5 findings

| # | Reviewer verdict | Author disposition |
|---|---|---|
| 1 (B) parity non-total, `%2F` | **OPEN** — reframed: skeleton evaluation misses param-dependent first-match | **applied**, see BLOCKER below |
| 2 (M) e2e cannot pass four boot checks | **PARTIAL** — tags fixed statically, but the fifth e2e route is still conditional on a non-nil scheduler callback (`e2e_seed.go:113-114`) | **applied** — §3 now states the e2e publisher's `Mount` is total; both the `E2EEnabled()` early return and the callback guard are deleted, 501 from the handler |
| 3 (M) orphan tests lose mux coverage | CLOSED | — |
| 4 (M) delta list omits HEAD/ordering | **PARTIAL** — param-sensitive ordering absent | **applied** — §10.3 class 9 |
| 5 (M) graph omits ADRs/wiki/BaseURL | CLOSED | — |
| 6 (N) IAM `router.go:102` comment | CLOSED | — |

### Job 2 findings

| # | Sev | Claim | Disposition |
|---|---|---|---|
| 1 | **BLOCKER** | A literal skeleton is not the route language: §10.2 cannot discover *partial* governance inside wildcard patterns. `GET /documents/{id}/placeholder-options/{pid}` with `pid=approval-preview` resolves to `CapDocumentSubmit` (`permissions.go:163`) while the skeleton reading picks the broad `CapDocumentView` row (`:164`) and emits no delta | **applied** — verified against source. §10.2's classifier is now three-valued (`none` / `uniform` / `partial`) with a per-field decision table; every `partial` pair is a mandatory delta (§10.3 class 9) |
| 2 | MAJOR | Validation rule 6 is impossible under one-document-per-invocation | **applied** — the generator now takes both documents in one run and emits both files with distinct symbols; rule 6 is vacuous before step 4, never skipped |
| 3 | MAJOR | `activeTags(publishers)` derives the expected set from the list under audit — vacuously true | **applied** — `expectedTags` comes from `specTags` (+ `specTagsE2E` when `useE2E`), fixed in both §3 and §5 |
| 4 | MINOR | The `pathPrefix` measurement is false: 80 rows, 20 ending in `/`, not 47 all ending in `/` | **applied** — independently re-measured before accepting (80 rows, 8 distinct non-slash values). The stated criterion was also wrong: it is "terminates at or before the end of the pattern's leading literal segments", and it is now asserted per-pair inside `classify` rather than written as a count |
| 5 | MINOR | Test table says five generator failures after §2 defines six | **applied** |
| 6 | MINOR | No per-step verification boundary despite the locked constraint mandating one | **applied** — §11 now names build + `go vet -tags integration` + the declared evidence per step |

### Author-side verification, run before accepting finding 4

A sonnet vote was dispatched to do nothing but *measure* `permissions.go`. It returned 120 rows
(confirmed), the real field names (`contains` / `notSuffix` — the artifact had invented
`pathContains` / `pathNotSuffix`), 80 `pathPrefix` rows, and the exact method distribution
(confirmed). I re-ran the prefix measurement myself rather than take either the reviewer's or the
vote's number, and additionally swept all 125 spec paths for sibling-prefix collisions — there are
none today, which is *why* the eight non-slash prefixes still classify `uniform`, and precisely why
that fact belongs in an assertion and not in a sentence.

Second-order result nobody asked for: `/healthz` is the table's **only** methodless row
(`permissions.go:85`). §7 deletes it, so after step 3 every rule is method-qualified and the
`method` classifier loses its "any" arm entirely.

### §2 judgement — is this same-altitude recurrence?

No, and the distinction matters. Round 6 is the **first** round against the chosen structure. Its
BLOCKER is not "the structure is wrong" but "the classification function has a missing arm, here is
the arm". Under the differential this would have been a seventh dimension to enumerate — unbounded.
Under the derivation proof it is a bounded, decidable case that the completeness assertions force.
The §2 ruling is not re-opened.

What the restructure did **not** do is make the design correct on the first attempt. §10.1's round
table now carries round 6 for that reason: it made the missing case *expressible*, which is a
smaller and more honest claim than the one the previous revision was drifting toward.

### Convergence

Count 6 (1B / 2M / 3N), same total as round 5 but the severity mix moved (round 5: 1B / 4M / 1N).
Altitude on §10 dropped from "the gate reads an infinite input space" to "the gate's classifier is
missing one arm" — a design fix, but a bounded and named one, not another dimension. Findings 4, 5
and 6 are rung 1–2 mechanical. Findings 2 and 3 are single-construct design fixes with the fix
named inside the finding.

Proceed map returned: steps 0, 1, 3 and 5 LICENSED. 2 blocked by finding 2 (now applied). 4 and 6
blocked by round-5 #2 (now applied). 7, 8, 9 blocked by the BLOCKER (now applied).

**Loop is not closed.** Every finding is applied, but the reviewer has not yet seen the three-valued
classifier — and a BLOCKER fix that no adversarial round has attacked is exactly what this ledger
exists to refuse to wave through. Round 7 judges it.

---

## Round 7 — judging the three-valued classifier (HEAD 93dbb997, Sol / medium)

### Job 1

| # | Verdict | Note |
|---|---|---|
| 1 (B) skeleton not the route language | **PARTIAL** — three values express partiality, but pairwise classification still does not reproduce ordered first-match resolution |
| 2 (M) rule 6 cross-document | CLOSED |
| 3 (M) `activeTags` circular | CLOSED |
| 4 (N) prefix measurement | CLOSED |
| 5 (N) five vs six generator failures | CLOSED |
| 6 (N) per-step evidence | CLOSED |
| r5 #2 e2e fifth route | CLOSED |
| r5 #4 param-sensitive ordering | **PARTIAL** — membership exists, but the proof still ignores which earlier rule wins |

### Job 2

| # | Sev | Claim |
|---|---|---|
| 1 | **BLOCKER** | The proof classifies rules independently instead of resolving the ordered decision list. `routeRules` is first-match (`permissions.go:342`); `:163` shadows `:164`, `:168` shadows `:169`. Pairwise comparison reports a false delta for every shadowed row. Fix: classify each rule's **effective language** `pattern ∩ rule ∩ ¬(all earlier rules)` and compare the resulting first-match partition; empty shadowed regions ignored |
| 2 | **BLOCKER** | The per-field table defines no composition algebra for the AND-ed predicate, and `notSuffix` is left as an unexplained "dual". The three labels are **not** a lattice: `uniform ∧ partial = partial`, `none ∧ partial = none`, but `partial ∧ partial` cannot be decided from labels at all. Fix: operate on the regular languages (intersect positives, complement `notSuffix`, then project to a label **last**) |
| 3 | MAJOR | Completeness assertion 2 and the zero-fallback claim are **false over request languages**: `PATCH /api/v1/iam/users/roles` routes to the `{user_id}` pattern but `notSuffix: "/roles"` (`permissions.go:120`) disqualifies its only explicit PATCH rule, reaching the fallback at `:336` |
| 4 | MAJOR | Class 9's golden file has no independent oracle — a classifier defect and its regenerated output commit together and the diff passes |
| 5 | MAJOR | Step 4 must create `api/openapi/internal-e2e.yaml` (absent today), and the design never assigns visibility markers, though unauthenticated seed callers need the first seed response to obtain session cookies (`frontend/apps/web/e2e/utils/seed.ts:27`) |
| 6 | MINOR | Step 3 deletes the sole methodless `/healthz` rule before step 7 claims to walk 120 rows — it is 119; risk table still says "Eight" after class 9 |
| 7 | MINOR | Per-commit-boundary audit covers only two spans; state the runnable/broken condition for every dependency edge |

Author verification: finding 3 confirmed against source. `/iam/users/{user_id}` PATCH exists
(`openapi.yaml:332`); there is no literal PATCH route at `/iam/users/{user_id}/roles`
(`openapi.yaml:998` is POST/PUT). The "fallback reached zero times" claim in §12 was measured over
the 147 **operations**, not over their request languages — the same skeleton-vs-language error as
the BLOCKER, third instance. This is the second fabricated measurement in the artifact (after the
47/80 prefix count).

### §2 escalation — the ratchet has fired on `classify`

`adversarial-review` §1 permits **two** consecutive patches to one construct before §2 runs
explicitly. `classify` has now been patched twice (round 6: two-valued → three-valued; round 7
would be the third: pairwise → ordered + language algebra). Rounds 5, 6 and 7 all landed on §10.2
and the last two on the same function. **No third patch is applied. The four §2 questions are put
to the operator.**

**Q1 — What is the global-maximum structure?** Two candidates, both nameable.

**(A) The ordered symbolic engine.** Represent each rule as a regular language over path segments.
Compute `E_i = pattern ∩ rule_i ∩ ¬(⋃_{j<i} rule_j)`, label the partition, and emit a delta for
every non-empty region where the old and new verdicts differ. This handles ordering, negation,
conjunction, `%2F` and param-spelling **uniformly** — it is the construction round 7 finding 1 and
2 jointly specify, and it is correct.

**(B) Stop proving equivalence; `routeRules` is not an oracle.** Rounds 5, 6 and 7 each found a new
instance of one fact: `routeRules` decides on **decoded substrings of the whole path**, so its
verdict can be steered by user-supplied parameter content. Where it disagrees with the router, it
is not a second opinion — it is **wrong**. The deletion license becomes: every one of the 147
operations carries an explicit reviewed `x-authz-*` annotation (the generator fails the build on
any missing one, so completeness is structural, not measured); one live integration test per
operation asserting the declared capability is required and its absence yields 403; §10.3 rows 1–8
keep their regression tests; and the param-spelling class is recorded once as a security
improvement rather than enumerated.

**Q2 — What does a proven system do here?** Neither Chi, Gin, nor `net/http` ships a policy-table
differ. The industry practice for replacing a hand-written authorization list with a generated one
is a positive conformance suite against the new source of truth plus a reviewed migration diff —
which is (B). Symbolic equivalence checking of two policy engines is done (AWS Zelkova for IAM
policies) but it is a funded product, not a throwaway test.

**Q3 — Costs, both sides.**
- (A) costs a small regular-language engine — realistically 400–700 lines of tricky code plus its
  own test suite — written to be **deleted at step 8**. It is itself unverified software used as a
  correctness oracle, which is exactly the objection round 7 finding 4 raises about the golden
  file, one level up. It also does not remove the need for the live conformance tests.
- (B) costs the equivalence proof. If a `routeRules` row encoded a capability decision that nobody
  transcribed correctly into the spec, only the row-by-row annotation review and the live test
  catch it. It does not prove the absence of an authoring slip; it makes one detectable by test
  rather than by proof.

**Q4 — Which is chosen, by whom?** **Operator decision. Pending.** Author recommendation: **(B)**.
Building sophisticated machinery to prove agreement with an object all three rounds have now agreed
is wrong-where-it-disagrees is the textbook definition of optimizing inside a local maximum, and
CLAUDE.md forbids it.

### Convergence

Count 7 (2B / 3M / 2N), up from 6. Altitude did **not** drop: both BLOCKERs are design-level and
both are on `classify`. §8's same-altitude-recurrence stop condition fires. **The loop is paused,
not closed.** Steps 0, 1, 2, 3 and 5 carry a clean license — one more than round 6, because the
single-invocation generator closed finding 2. Findings 5, 6 and 7 are applicable under either §2
outcome and are held only so no work is done that a (B) ruling would delete.
