# Milestone 3 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` + `../validation-contract.md` (D4) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-03 (RE-VALIDATION) · **Verdict:** see C7.
>
> **Why re-validated:** The prior PASS (`5f7169b5`) treated the F3.2 negative-RLS integration proof as a
> *deferred* live run. The operator required the real run; it was **RED** on first execution (the proof
> targeted `public.documents`, whose M2 capability write-tripwire `trg_require_cap_asserted` raises `P0001`
> *before* RLS is the deciding control, masking the outcome). Fix-feature **F3.5** retargeted the proof to
> `metaldocs.notifications` (FORCE-RLS `tenant_isolation`, a real F3.2 async seed site, NOT tripwired) and
> ran it GREEN for real. The prior PASS is VOID; this verdict re-runs the full C1–C7 from clean state and
> does not assume any prior feature is still green.

## Inputs loaded

milestone.md · validation-contract.md (incl. F3.4 §0.3/§2.4/§4 erratum + §6 F3.5 defer-CLOSED note) ·
f3.1/f3.2/f3.3/f3.4/f3.5 spec+plan+evidence · program README.md · aggregate diff `da7ae95c..HEAD`
(commits `de94758a`, `bad74d86`, `a6c874a3`, `118e0196`, `5f7169b5`, `66af8f04`). All present and readable.

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F3.1 | ✅ platform carrier reads only `internal/platform/tenant` (no iam import); chokepoint seeds both `Do`/`DoReadOnly`, present→seed / absent→no-op | ✅ 21 `SeedTxIdentity` sites = 21 allowlist entries, 0 outside; SEED-CHOKEPOINT blocking | ✅ no RLS policy change | spec/plan/evidence + allowlist |
| F3.2 | ✅ `SeedTxTenant` tenant-only (no actor), 5 processing txs seeded; ASYNC-TENANT-SEED handler-scoped per §2.3 | ✅ lint 0 live violations, RED-on-synthetic captured; `emitStuckAlert` gap closed | ✅ policy unchanged; tx-local GUCs | spec/plan/evidence |
| F3.3 | ✅ ADR 0027 dated amendment (all 5 §3.1 points + per-binary §4 table); wiki tenancy pages updated | ✅ no stale "async no backstop"/"85 sites" claim; wiki-curator clean | ✅ docs-only, verified 0 code diff | spec/plan/evidence |
| F3.4 | ✅ `async-tenant-tables.txt` set-equal to the 33 FORCE tables; `idempotency_keys` a real entry, `job_leases` noted no-`tenant_id` | ✅ contract §0.3/§2.4/§4 corrected under operator-approved HS-7 re-open + dated erratum | ✅ docs+list only, RLS byte-identical | spec/plan/evidence |
| F3.5 | ✅ proof retargeted to `metaldocs.notifications` (non-tripwired FORCE-RLS, real async seed site) so RLS is the sole control | ✅ real green run; PG-1/2/3/4 all pass (see C2) | ✅ test-construction only; SeedTxTenant/policy/migration byte-identical | spec/plan/evidence |

- All five features have spec.md (approval recorded: F3.1–F3.4 dated `2026-07-03`; F3.5 opened by
  operator-mandated HS-4 real run, recorded in the README HS table + spec header), populated interview
  records (F3.4/F3.5 explicit Q&A tables), execution-shaped plan.md, and evidence.md whose acceptance
  tables match each spec's validation gate row-for-row.
- Minor doc note (not a fail): F3.5's spec records approval via its operator-mandated opener rather than a
  standalone dated "Approved" line as F3.1–F3.4 use. Intent of C1.1 (contract authored before the fix, not
  guessed) is met — F3.5 has a full consumer-contract + interview record and an operator-mandated opener.

**C1: PASS.**

## C2 — Gates re-run, isolated (validator ran these, not trusted from transcript)

| Feature / gate | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| F3.5 negative-RLS proof (**load-bearing, real DB**) | `run-rls-proof.ps1` → `go test -tags integration -count=1 -v -run 'SeedTxTenant_RLSBackstop' ./internal/modules/iam/authz/...` (live Postgres `metaldocs-postgres`, port 5433, NOBYPASSRLS role) | `--- PASS: ...LeakBeforeBlockedAfter (105.05s)` + all 4 subtests PASS: `leak_before_no_seed`, `.../select_update_delete_see_zero_rows`, `.../insert_or_update_producing_b_row_is_42501`; `ok metaldocs/internal/modules/iam/authz 107.899s`; `EXITCODE=0` | ✅ |
| F3.1/F3.2 api-lint (both new rules, live, blocking) | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` (exit 0) | ✅ |
| F3.1/F3.2 lint unit tests | `go test ./scripts/api-lint/ -run 'SeedChokepoint|AsyncTenantSeed'` | `ok metaldocs/scripts/api-lint 2.489s` | ✅ |
| F3.1 census | grep non-test `SeedTxIdentity(` outside chokepoint | 21 sites, all 21 in `seed-chokepoint-allowlist.txt`; 0 outside | ✅ |
| F3.4 table-set equality | `diff <lint-tables> <FORCE-RLS tables>` from `0001_current_schema.sql` | `SET_EQUAL` (33 == 33; `idempotency_keys`, `notifications` present) | ✅ |
| Build | `go build ./...` | exit 0, no output | ✅ |
| M2 regression lints | `go test ./scripts/api-lint/ -run 'Tripwire|CapName|Divergence|Arm'` | `ok metaldocs/scripts/api-lint 1.740s` | ✅ |

- **The load-bearing question is answered: the negative-RLS proof RAN GREEN for real** — my own live-DB
  execution, not the evidence transcript. It is not deferred, not skipped, not a mock/sqlmock: the
  `leak_before_no_seed` subtest reproduces the 1-row cross-tenant leak (proving RLS, not some other gate,
  is the sole control), and the seeded subtests give 0/0 rows + SQLSTATE `42501` on retenant.

**C2: PASS.**

## C3 — Senior review of the aggregate milestone diff

Reviewed `da7ae95c..HEAD` as one unit.

- **No split-brain:** F3.4 made `async-tenant-tables.txt` set-equal to the 33 FORCE tables and reconciled
  the three prior records that falsely claimed `idempotency_keys` has no `tenant_id`; the D4 contract
  §0.3/§2.4/§4 carry a dated erratum. One fact (which tables are tenant-scoped), one source of truth.
- **No production diff from F3.5:** `git show 66af8f04 --name-only` = the test file + M3 docs only; no
  `.go` (non-test) / `.sql` / migration. `SeedTxTenant` last touched in F3.2 (`bad74d86`) — byte-identical.
  The fix corrected the *proof*, not the thing under proof (correct separation).
- **No dead code / no feature breaking another:** the ASYNC-TENANT-SEED lint is handler-scoped (documented,
  mirrors M2 TRIPWIRE-ARM-DRIFT scoping); the F3.2 main-session audit closed the real `emitStuckAlert`
  unseeded-write gap the AST lint could not see — an honest coverage boundary, recorded.
- **Honest residual (not a defect):** the F3.1 allowlist retains 21 manual seeds — genuine distinct-actor
  (category A) and raw-`BeginTx` repository/taxonomy sites the TxRunner chokepoint structurally does not
  cover (they open their own `*sql.Tx`). Each is enumerated with a reason; the SEED-CHOKEPOINT lint blocks
  *new* drift. This is a bounded coverage limit, transparently recorded — not split-brain or symptom-patch.

Staff-engineer bar met? ✅

**C3: PASS.**

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api multi-tenant) | pass | Both new lints blocking + 0 live violations; targeted RLS integration proof green (real DB); `go build ./...` green. Full 20-min suite deliberately NOT run (mission §10). |
| Regression vs prior milestones | all still pass | M2 tripwire/cap-name lints: `go test ./scripts/api-lint/ -run 'Tripwire\|CapName\|Divergence\|Arm'` → `ok`. `go build ./...` green. RLS policy byte-identical (no migration in scope) → M0/M1 contract gates unaffected. |

**C4: PASS.**

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Manual-seed discipline (finding #2) | ~62 hand-seed acts, forgetting one silently absorbed by NULL-permissive RLS | Structural: chokepoint auto-seed (api) + `SeedTxTenant` per-message (async), guarded by 2 blocking lints | api-lint `0 violation(s)`; 21 residual sites all allowlisted-with-reason + drift-locked |
| Async RLS backstop (Dimension 4) | worker/jobs seeded nothing; zero RLS backstop | FORCE RLS engaged on async processing txs; **proven live** | My own real run: unseeded cross-tenant UPDATE leaks 1 row; `SeedTxTenant(A)` → 0/0 rows + `42501`. Not asserted — executed. |
| RLS policy integrity | NULL-permissive, FORCE, 33 tables | **byte-identical** | No policy/migration diff anywhere in `da7ae95c..HEAD`; bar moved by adding seeding, not weakening RLS |

- Root cause fixed, not symptom-patched: F3.5 did **not** weaken any assertion to force green — it moved the
  proof to a table where RLS is the sole control and kept the leak-before / 0-row / 42501 shape intact.
- Could it be built better? The 21-entry raw-`BeginTx` allowlist is the residual: those repository writes
  bypass the TxRunner chokepoint. A future consolidation onto the chokepoint (or a repo-tx wrapper that
  seeds) would shrink the allowlist toward 0. Recorded as next-milestone input; the current construction is
  sound (drift-locked), so it does not FAIL M3.

**C5: PASS.**

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean; each
  feature's acceptance mapped to a validator-run command (C2).*
- [ ] Fixture/mock passed off as real-provider proof — *clean; the load-bearing proof is a real live-DB run
  the validator executed; F3.1's sqlmock PG-1 is explicitly labeled fixture and the real backstop proof is
  F3.2/F3.5's integration drive.*
- [ ] Consumer contract guessed rather than read — *clean; contracts read from source/consumer.*
- [ ] Split-brain — *clean; F3.4 reconciled the `idempotency_keys` fact to one source of truth.*
- [ ] Self-judged close / validator edited or fixed code — *clean; validator only judged + wrote this file.*
- [ ] Scope drift — *clean; the `emitStuckAlert` fix and the allowlist are recorded with rationale.*
- [ ] Symptom-patch — *clean; F3.5 retargeted the proof without weakening assertions; RLS byte-identical.*

(All unchecked = clean.)

**C6: PASS.**

## C7 — Verdict

- **VERDICT: PASS**
- The load-bearing question — *did the negative-RLS proof actually run green for real (not deferred, not
  mocked)?* — is answered **YES** by the validator's own live-Postgres execution: all subtests PASS,
  leak-before reproduced (1 row), retenant `42501`, `EXITCODE=0`. F3.5 is a test-construction-only fix with
  no production/policy/migration diff. All other M3 themes (F3.1 chokepoint + census, F3.2 async seed +
  lint, F3.3 ADR/wiki, F3.4 table-set truth) re-verified green from clean state; M0/M1/M2 gates
  un-regressed; `go build ./...` green.
- Handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only on PASS + operator HS-1 approval
