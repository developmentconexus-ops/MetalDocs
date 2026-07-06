# Milestone 7 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (up-front spec) + `../validation-contract.md` (D4 binding) +
> each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-05  ·  **Verdict:** see C7 — **PASS**.
> Judge-only role: this file is the sole artifact written. No source, status, or spec was edited.

## Inputs loaded (fail-closed check)

All present and readable: `milestone.md`; `validation-contract.md` (§0–§6); `f7.1` evidence (gate feature — spec/plan
correctly absent, it IS the gate + ADR, per D7); `f7.2`/`f7.3`/`f7.4` spec+plan+evidence; program `README.md`;
governing `mission.md` §7 M7; ADR 0070 (Accepted 2026-07-05); ADR 0027 amendment (F7.4 stamps). Aggregate diff
`2486232e..HEAD` (M6 close → M7 tip, 111 files, +11,482). No missing input → no fail-fast.

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F7.1 | ✅ | ✅ | ✅ | Gate 🟡 Yellow (no AS-1/2/3) + ADR 0070 Accepted, cited by milestone. Gate feature: no spec/plan by design (it authors the contract rails). |
| F7.2 | ✅ producer (`POST /tenants` → `iam OnboardTenant`) matches consumer contract §1; source-of-truth `openapi.yaml` regen clean | ✅ all 6 gate rows re-proven (below C2) | ✅ no crypto impl (seam only), no export/erasure, no RLS flip | `f7.2-onboarding/evidence.md`; interview populated; approval line filled (2026-07-05, ADR-0070-locked) |
| F7.3 | ✅ per-module `TenantDataPort` fan-out (invariant-6 clean); census = live schema query (single source of truth); export/erase routes match §2/§3 | ✅ crypto/coverage/export/erasure/isolation/tripwire all re-proven | ✅ no KEK rotation, no plaintext-audit backfill, no RLS flip — all in Non-goals + declared defers | `f7.3-export-erasure/evidence.md`; interview #2a/#2b recorded |
| F7.4 | ✅ CI role NOSUPERUSER+NOBYPASSRLS+non-owner; §4.5 negative+positive proven under `metaldocs_ci`; source-of-truth = pg role catalog + pg_policy | ✅ all 11 gate rows re-proven; two deviations judged acceptable (see C5) | ✅ no `metaldocs_app` flip, no 63-table reassign, no policy-model rewrite, no speculative predicate retrofit | `f7.4-rls-truth-sweep/evidence.md`; both deviations surfaced (not patched) for HS-1 |

Every feature in `milestone.md`'s Features table has the required artifacts. Approval lines filled. Interview records
populated (F7.2/F7.3/F7.4 all have Q&A rows). `plan.md` files are execution-shaped (task graph, files, TDD ordering,
risks) — not re-specs. `evidence.md` acceptance tables map row-for-row to each `spec.md` Validation Gate. Non-goals and
the milestone rabbit-hole list respected. Every deviation carries a written rationale. **C1 PASS.**

## C2 — Gates re-run, isolated (validator ran these from clean state, not trusted from transcript)

| Feature / gate | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| Build (all) | `go build ./...` | exit 0 | ✅ |
| Regression: full unit suite | `go test ./...` | exit 0, no FAIL lines | ✅ |
| Contract lints | `go run ./scripts/api-lint -strict ./api/openapi/v1/openapi.yaml .` | `0 violation(s)` | ✅ |
| Integration vet | `go vet -tags=integration ./tests/...` | exit 0, 0 output | ✅ |
| Caps (registry 38) | `go test ./internal/modules/iam/domain/ -run 'RegistrySize\|Classified\|AreaGrade'` | `ok` (want=38 confirmed) | ✅ |
| Tripwire arms 18→20 (M2 dep) | `go test ./internal/platform/tripwire/...` | `ok`; arms map has `tenants`/INSERT + `tenant_lifecycle_jobs`/INSERT | ✅ |
| Crypto + api-lint units | `go test ./internal/platform/crypto/... ./scripts/api-lint/...` | both `ok` (incl. sole-RLS synthetic negative) | ✅ |
| **F7.4 §4.5 RLS proof** | `go test -tags=integration -run TestRLSTruth_NonOwnerRoleEnforcesIsolation ./tests/integration/security/...` | `--- PASS`, `ok 142.084s` — 4 cases (a positive / b wrong-GUC-blocks-0 / c bypass-leaks-under-owner / d null-GUC-pin) + role-attr f\|f\|0 + approval_signoffs catalog | ✅ **real Postgres, non-owner `metaldocs_ci` conn** |
| F7.3 coverage census | `go test -tags=integration ./tests/integration/tenantdata/...` | `ok 133.214s` (live schema census == registered ports ∪ allowlist) | ✅ real |
| F7.3 export/erase/seal | `-run TestTenantExport_CompleteArtifact\|TestTenantErasure_ChainStaysGreen\|_DoesNotTouchOtherTenant\|TestOnboardTenant_AuditPayloadSealedWhenCryptoWired` | 4× `--- PASS` (`ok 30.6s`): export complete, chain GREEN + auth_identities=0 post-erase, tenant-B untouched, payload sealed not plaintext | ✅ real |
| F7.2 onboard + tripwire | `-run TestOnboardTenant_*\|TestTenantsInsertTripwire` | 5 tests + 2 subtests `--- PASS` (`ok 15.7s`): e2e login+act, atomic create, dup-slug 409, requires-cap, tripwire reject-without-cap / accept-with-cap | ✅ real |
| System runnable | `.\scripts\check-system-runnable.ps1` | exit 0 — blank-template, login 200, session, auth/me 200, ready 200 all PASS | ✅ |

No test that passed in implementation failed on isolated re-run. Fixture vs real labeled: crypto/lint/domain are pure/AST;
every RLS/export/erasure/onboard proof ran on real Postgres (container `metaldocs-postgres`), the RLS proof under the
genuinely non-bypassing `metaldocs_ci` role. **C2 PASS.**

## C3 — Senior review of the aggregate milestone diff

- **Structure:** per-module `tenant_data_port.go` files (documents, controlleddocuments, templates, iam, auth, notifications,
  render, taxonomy, tokens, jobs, audit) — cross-module access via published `tenantdata.Port` only, never a god-query.
  Invariant-6 clean; ADR 0070 decision 2 honored.
- **Split-brain check — PASS.** The export/erasure census is a **live `information_schema` query** (`coverage_test.go`
  `censusQuery`), compared against the union of registered ports' `Tables()`. There is no hand-maintained parallel table
  list to drift — the guard *kills* the hand-sync rot class rather than adding to it. Tripwire arms are generated from the
  registry (M2), not hand-synced. Capability registry has one source (`model.go`), size guard at 38.
- **Live-QA fixes are root-cause, not patches.** (a) `auth_identities` erase: user_id-join DELETE ordered before iam
  deletes `iam_users` — closes a real "PII survives erasure" gap. (b) tx-aware seal (`EncryptForTenantTx`/`WrappedDEKTx`
  thread `*sql.Tx`): fixes a genuine same-tx key-visibility defect where onboarding audit payloads landed plaintext
  forever. (c) migration 0283: a **latent class-defect** (BEFORE-DELETE trigger `RETURN NEW`→NULL silently cancelling
  every armed-table DELETE) fixed **at the generator** (`render.go`), correct-by-construction — surfaced by erasure, fixed
  for the whole class. All three disclosed in evidence.
- **Orchestrator (`tenant_lifecycle_service.go`):** H-PRE-1 respected (erasure-tombstone audit via `Record` on its own tx
  after commit, never inside the lock-holding fan-out tx); state-write/network separation (blob delete outside DB tx);
  explicit tenant predicates; idempotent 3 phases; nil-safe crypto. Minor cosmetic (`store2` field name) — non-blocking.
- **No feature broke another:** full unit suite + all targeted integration suites green; audit chain validates GREEN after
  erasure (M6/audit-immutability preserved); M2 tripwire generation intact (arms 18→20); M3 tenancy chokepoint reinforced
  (F7.4 makes RLS real, adds a complementary read-side lint). No dead code left by a superseded approach.
- Staff-engineer bar met? ✅

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| authz (5-surface + tripwire arms) | pass | 3 caps × 10-touchpoint complete; tier-1 routes mapped (no silent escalation); erase = system_admin-ONLY seed; tripwire negative proof green |
| contract (openapi regen / lints) | pass | api-lint strict 0 violations; regen clean; `iam/api/api.gen.go` regenerated |
| multi-tenant isolation | pass | F7.3 erasure isolation (tenant-B untouched); F7.4 real RLS negative+positive under non-owner role |
| DB-invariant | pass | audit append-only survives erasure (chain GREEN, 0 audit-row mutations); FORCE RLS active under non-owner; `approval_signoffs` gap closed (0285) |
| docs | pass | ADR 0070 cited; ADR 0027 amended with F7.4 stamps; wiki `approval_signoffs.md`, integration-harness updated |
| Regression vs prior milestones | all still pass | full `go test ./...` green; M2 tripwire gen/drift intact; M3 chokepoint reinforced not regressed; audit suite (M6/erasure) chain-green; system-runnable green. F7.4 role flip makes prior isolation tests *more* real, breaks none |

**C4 PASS.**

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Kernel gap: tenant reachable end-to-end | tenants exist only via manual seed | onboard via `POST /tenants` → login as seeded admin → capability-gated action → cross-tenant 404 | `TestOnboardTenant_EndToEnd_LoginAndAct` PASS + F7.2 live drive; not a seam-only stub |
| GDPR-erasure ⊥ audit-immutability | unresolved tension | crypto-shred DEK; audit skeleton byte-intact; chain validates GREEN post-erasure | `TestTenantErasure_ChainStaysGreen` PASS — 0 audit-row mutations, payload undecryptable post-shred. Root cause (key destruction), not audit-immutability weakened |
| Tenant-isolation tests real (not false-green) | dev role SUPERUSER+BYPASSRLS → FORCE RLS inert | isolation proven under NOSUPERUSER+NOBYPASSRLS+non-owner `metaldocs_ci`; wrong-GUC blocks (0), bypass-owner leaks (contrast) | `TestRLSTruth_...` PASS 142s — **role fixed** (correct-by-construction), plus a forward lint, not per-query symptom hunt |

**Two F7.4 deviations — judged on merit against contract INTENT + ratified invariants:**

1. **§4.2 role mechanism (dedicated non-owner `metaldocs_ci` vs 63-table ownership-reassignment)** — **ACCEPTABLE,
   within-boundary.** §4.2's actual requirement is "app-connection role is a non-owner under NOBYPASSRLS; RLS genuinely
   filters." `metaldocs_ci` satisfies every stated property (NOSUPERUSER, NOBYPASSRLS, owns 0 tables — proven `f|f|0`, DML-only).
   RLS applies to a non-owner under plain ENABLE (FORCE is only needed for the owner), so filtering is genuine — proven live.
   Avoiding the schema-wide owner migration that would break dev bootstrap/migrations/janitors is a legitimate HS-2 call.
   Prod posture (baseline "Owner: -", separate migration identity) faithfully mirrored. Not a symptom-patch.
2. **§4.2/§4.5 "no-GUC → 0 rows" reconciled to wrong-GUC-blocks + bypass-contrast** — **ACCEPTABLE, meets §4.5 intent.**
   Literal "no-GUC → 0 rows" is false-by-design under the ratified M3 null-GUC-permissive idiom (ADR 0027 amendment; the
   leader-elected janitors scan cross-tenant with no GUC). Removing the null branch across 34 policies would break those
   janitors — squarely HS-2. F7.4 proves the *actual* security property (cross-tenant isolation) via a strictly stronger
   assertion: wrong-GUC returns 0 of the other tenant's rows under the non-bypassing role, while the same query leaks under
   BYPASSRLS (documenting the exact pre-F7.4 false-green). Null-GUC=all-rows is explicitly pinned (case d) so it can't
   silently regress into a leak. This is the §4.5 non-negotiable intent (RLS genuinely active, not bypassed).

   Neither deviation is a MUST-deviation from the *target architecture*; both uphold "app connects as a non-owner under
   NOBYPASSRLS; RLS genuinely filters." Both were surfaced (not silently patched) for the HS-1 operator gate. Neither is an
   unmet criterion nor a symptom-patch → no HS-4 fix feature warranted.

- Could it be built better? The full integration suite is not blanket-reconnected through `metaldocs_ci` (setup legitimately
  needs the owner) — census is (lint static coverage of async reads) + (live non-owner isolation proof). This is honest and
  correct-by-construction, disclosed as a bounded defer with trigger "a non-owner-capable setup path is built / CI runs the
  suite under the flipped role." A whole-suite role swap is a reasonable future increment, not an M7 unsoundness. Pre-F7.3
  plaintext audit rows remain non-shreddable (backfill would mutate immutable rows — forbidden) — documented limitation with
  trigger (audit re-baseline). No current construction is unsound. **C5 PASS.**

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — **clean** (each acceptance row
      mapped to a named test re-run in C2; note the F7.2 evidence itself self-corrected an earlier false-green defer where
      DATABASE_URL-unset t.Skip masked as `ok` — that honesty is a positive signal)
- [ ] Fixture/mock passed off as real-provider proof — **clean** (RLS/erasure/export/onboard all real Postgres, labeled)
- [ ] Consumer contract guessed rather than read from the consumer — **clean** (source-of-truth cited per feature)
- [ ] Split-brain — **clean** (census = live schema query; arms generated; one cap source)
- [ ] Self-judged close / validator edited or fixed code — **clean** (validator wrote only this file; status unflipped)
- [ ] Scope drift — **clean** (0282/0283 migrations + templates test change are erasure-required and disclosed; no work
      beyond F7.1–F7.4 without rationale)
- [ ] Symptom-patch — **clean** (erasure = crypto-shred at root; F7.4 = role fixed; 0283 = generator-level class fix)

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass: **code-wise** (senior-level, invariant-clean, no split-brain / dead code / guessed contract) and
  **function-wise** (end-to-end onboarding, complete export, crypto-shred erasure with GREEN audit chain, and genuinely-real
  RLS under a non-bypassing non-owner role — all proven on real Postgres from clean state).
- The two F7.4 deviations are within-boundary implementations of the contract's INTENT (same security property, lower blast
  radius, both surfaced for HS-1) — not unmet criteria, not symptom-patches.
- Handed back to the main session to flip status and present the **HS-1 operator gate** (which must also ratify the two F7.4
  deviations). Commits are local/unpushed — correct per mission constraint.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending — must ratify the two F7.4 §4.2/§4.5 deviations.
> - Status flipped in `README.md`: no — main session action, only on this PASS + operator approval.
