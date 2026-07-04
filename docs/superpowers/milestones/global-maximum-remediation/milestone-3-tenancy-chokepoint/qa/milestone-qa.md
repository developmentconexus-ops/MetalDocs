# Milestone 3 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` + `../validation-contract.md` (D4, binding) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-03  ·  **HEAD:** `a6c874a3`  ·  **Baseline:** `e80c1800` (pre-impl contract commit)  ·  **Verdict:** see C7 → **FAIL**.
> The validator judges and writes this file only; it edited no source and flipped no status.

## Inputs loaded (none missing — did not judge blind)

milestone.md · validation-contract.md (§0–§6) · program `README.md` · governing `mission.md` (§7 M3) ·
all three features' `spec.md`/`plan.md`/`evidence.md` · aggregate diff `git diff e80c1800..a6c874a3`
(74 files, +3332/−239) across the three feature commits `de94758a` (F3.1) · `bad74d86` (F3.2) ·
`a6c874a3` (F3.3).

---

## C1 — Spec & plan conformance (per feature)

| Feature | spec approved before code | interview populated | consumer contract honored | acceptance met | non-goals respected | Evidence |
|---------|---------------------------|---------------------|---------------------------|----------------|---------------------|----------|
| F3.1 | ✅ 2026-07-03 | ✅ 5-row Q&A | ✅ FORCE-RLS + `SEED-CHOKEPOINT` | ✅ | ✅ | see C2/C3 below |
| F3.2 | ✅ 2026-07-03 | ✅ 5-row Q&A | ✅ FORCE-RLS + `ASYNC-TENANT-SEED` + neg-proof | ⚠️ (see finding) | ✅ | see C2/C3 below |
| F3.3 | ✅ 2026-07-03 | ✅ 4-row Q&A | ❌ **durable record does not match runtime truth** | ❌ | ✅ (docs-only diff confirmed) | **F-1 (below)** |

All nine artifacts (`spec.md`/`plan.md`/`evidence.md` × 3) exist and are execution-shaped (task lists,
files-touched, test strategy). Each `evidence.md` acceptance table maps to its `spec.md` Validation Gate.

**C1 result: FAIL — on F3.3.** F3.3's acceptance §3.3 / spec PG-2 requires *"wiki tenancy docs match
runtime truth."* They do not (finding **F-1**). F3.1 and F3.2 conform code-wise; F3.2 carries the same
factual defect **only** in a documentation/rationale clause (it does not affect F3.2's runtime seeding).

### F-1 (the failing finding) — `idempotency_keys` "no tenant_id column" is false; carried into the durable ADR record

- **Contract clauses affected (load-bearing):** validation-contract.md §0.3 (async-fleet table),
  §2.4 (sanctioned allowlist: *"System tables with no `tenant_id` column — `idempotency_keys` … RLS is
  structurally N/A"*), and §4 janitor row (*"`idempotency_keys`/`job_leases` having no `tenant_id`"*).
- **Runtime truth (traced to source):** `db/baseline/0001_current_schema.sql` —
  `CREATE TABLE metaldocs.idempotency_keys ( tenant_id uuid NOT NULL, … )` **and** it carries
  `FORCE ROW LEVEL SECURITY` + the identical `tenant_isolation` policy (it is 1 of the 33 FORCE tables).
  The janitor's cross-tenant `DELETE FROM metaldocs.idempotency_keys` (`internal/modules/jobs/idempotency_janitor/job.go:33`)
  survives **only** via the NULL-permissive escape hatch (GUC-unset) — the exact **opposite** of
  "RLS cannot apply." `job_leases` genuinely has no `tenant_id` (correct); `idempotency_keys` does not.
- **Baked into the durable record (F3.3):** ADR `wiki/decisions/0027-rls-adoption-sequencing.md` amendment
  restates the false claim twice — **line 183** (§4 residual surface) and **line 202** (per-binary table
  janitor row). The amendment thereby **self-contradicts the same ADR's own body at line 91**, which lists
  `idempotency_keys` among tables that *carry* `tenant_id`.
- **Severity classification (both dimensions judged):**
  - *Function-wise / runtime:* **SAFE.** A system-wide TTL sweep MUST be cross-tenant; running it GUC-unset
    under NULL-permissive is the sanctioned design (§2.4 "cross-tenant / system-maintenance" category; §4
    janitor "Unseeded — NULL-permissive by design"). No tenancy leak; no unseeded tenant-scoped *business*
    write. RLS policy byte-identical (verified in C3). M3's real security objective is met.
  - *Code-wise / contract-truth:* **DEFECT.** A load-bearing contract clause (§4 per-binary table + §2.4
    allowlist) states a **false schema fact**, and F3.3 propagated it into the durable ADR record — the
    precise "stale/incorrect claim in the enforcement record" class that F3.3 exists to eliminate (its
    acceptance PG-2). This is an **HS-7** contract-accuracy divergence and a **C1(F3.3) acceptance miss**.
- **Why this fails and cannot be waived:** F3.3's *entire deliverable* is durable-record truth; shipping a
  record that misstates the enforcement surface and self-contradicts the same ADR defeats the feature. The
  correct rationale is available and trivial (exempt `idempotency_keys` as a *sanctioned cross-tenant
  system-maintenance sweep relying on the NULL-permissive hatch* — same class as the audit scan), so this
  is a minimal, nameable fix, not a redesign.

---

## C2 — Gates re-run, isolated (validator ran these from clean tree @ `a6c874a3`; not trusted from evidence)

| Gate | Command re-run | Real output | Pass? |
|------|----------------|-------------|-------|
| build | `go build ./...` | exit 0 | ✅ |
| build (integration tag) | `go build -tags integration ./...` | exit 0 | ✅ |
| core pkgs | `go test ./internal/platform/tenant/... ./internal/platform/db/... ./internal/modules/auth/... ./internal/modules/iam/authz/... ./scripts/api-lint/...` | all `ok` (api-lint 7.6s) | ✅ |
| async + collapsed-site pkgs | `go test ./internal/platform/worker/... ./internal/modules/documents/approval/... ./internal/modules/notifications/infrastructure/... ./internal/modules/render/fanout/... ./internal/modules/jobs/stuck_instance_watchdog/... ./internal/modules/controlleddocuments/application/... ./internal/modules/templates/application/... ./internal/modules/documents/repository/...` | all `ok` | ✅ |
| live api-lint (both new lints blocking) | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | **0 violation(s)**, exit 0 | ✅ |
| lint unit tests | `go test ./scripts/api-lint/ -run 'SeedChokepoint\|AsyncTenantSeed'` | 14/14 PASS (clean-green, stray-fires-named, after-removal-green, exemptions, stale-allowlist) | ✅ |
| **F3.1 PG-3 live RED→GREEN** | inserted throwaway `authz.SeedTxIdentity` in `internal/modules/documents/application/` | `SEED-CHOKEPOINT … zz_validator_probe.go:11`, **1 violation, exit 1**; removed → **0, exit 0** | ✅ |
| **F3.2 PG-3 live RED→GREEN** | inserted throwaway unseeded `UPDATE documents` in `internal/platform/worker/` | `ASYNC-TENANT-SEED … zz_validator_probe.go:10 (op UPDATE, table documents)`, **1 violation, exit 1**; removed → **0, exit 0** | ✅ |
| F3.1 census | `grep -rn "SeedTxIdentity(" --include=*.go internal apps scripts \| grep -v _test.go` | 21 live call sites (excl. `context.go:48` defn + `seed_chokepoint_rule.go:10` doc-comment) = **21 allowlist entries; 0 outside chokepoint+allowlist** | ✅ |
| F3.2 negative-RLS proof | `go vet -tags integration ./internal/modules/iam/authz/` | exit 0 (compiles + vets); run **deferred** (no DB env — accepted bounded defer §2.5/§6) | ✅ (authored) |
| F3.3 stale-claim grep | `grep -rniE "async .*no .*backstop\|~?85 sites\|only on controlled_documents\|no RLS backstop"` on the 3 touched pages | 0 matches | ✅ |
| M2 regression | `go test ./internal/modules/iam/domain/ -run CapabilityRegistrySize` | PASS; live api-lint 0 violations ⇒ the 5 authz lints + M2 tripwire drift/parity lints still green | ✅ |

Throwaway probe files were deleted; final `git status` clean (only untracked `docs/release/`). No gate I
ran regressed. All C2 commands passed — **the FAIL is not a broken gate; it is the C1/HS-7 truth defect F-1.**

---

## C3 — Senior review of the aggregate milestone diff

Reviewed `git diff e80c1800..a6c874a3` as one unit.

- **Chokepoint (F3.1, `internal/platform/db/runner.go`):** correct — `seedTxIdentityFromContext` in the
  shared `do()` reads the **platform** carrier (`platform/tenant`, no `internal/modules/iam` import →
  module boundary honored), seeds both GUCs only when **both** present, no-ops when either absent, runs
  **before** `fn` (before Require/locks, H-PRE-1), applies to both `Do`/`DoReadOnly`. SQL byte-matches
  `SeedTxIdentity`. Senior-clean.
- **Collapse (F3.1):** 21 residual `SeedTxIdentity` sites, each allowlisted with a category-A (distinct
  stored actor: `d.CreatedBy`/`assignedBy`/`grantedByActor`) or raw-`BeginTx`/system reason; census = 0
  outside chokepoint+allowlist; the lint confirms **no stale allowlist line**. The one review-caught
  over-removal (`SystemCancelInstance` watchdog path) is correctly restored + allowlisted (category B).
- **Primitive (F3.2, `authz.SeedTxTenant`):** tenant-only `set_config('metaldocs.tenant_id',$1,true)`, no
  actor — matches §2.1 exactly; orthogonal to `BypassSystem`.
- **Five seed sites (F3.2):** materialize / pdf (wrapped) / scheduled-publish (seed before `FOR UPDATE`) /
  notifications-fanout (wrapped) / staging-outbox processing — all present; pdf legacy constructor kept for
  back-compat; each tx single-tenant (no HS-2 tenant-mix). The **review-fix** `emitStuckAlert`
  (`stuck_instance_watchdog/job.go:192`) correctly adds `SeedTxTenant(ctx, tx, inst.TenantID)` before the
  `governance_events` write — the one gap the lint could not see, now closed.
- **Negative RLS proof (F3.2):** `seed_tx_tenant_rls_integration_test.go` shape matches §2.5 verbatim
  (NOBYPASSRLS role; two tenants; leak-before unseeded; blocked-after `SeedTxTenant(A)` → B invisible,
  0-row UPDATE/DELETE, A visible, re-tenant UPDATE → `42501`). Real-DB (testdb), not sqlmock.
- **Split-brain finding (F-1):** the sole split-brain — the async-fleet exemption fact for
  `idempotency_keys` exists in **two forms**: (a) code/reality = *has* `tenant_id`, FORCE-RLS, sweeps
  cross-tenant via NULL-permissive; (b) contract §2.4/§4 + ADR amendment = *"has no `tenant_id`"*. One fact,
  two contradictory sources (and the ADR contradicts its own body line 91). This is the C3 split-brain.
- **Staff-engineer bar:** met on mechanism (F3.1/F3.2 code); **not met** on durable-record truth (F3.3 /
  the §2.4/§4 rationale). A staff reviewer would block the ADR merge for the self-contradicting schema claim.

**C3 result:** one blocking finding (F-1, split-brain / stale claim in the enforcement record). No
duplication, no dead code, no feature broke another.

---

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api multi-tenant) | pass-with-defers | both new lints blocking + 0 live violations; targeted tenant/isolation drives compile+green; full 20-min suite intentionally not run (mission §10) |
| RLS policy byte-identity | pass | `git diff e80c1800..a6c874a3` touches **no** `db/**`, migration, or `*.sql`; no `CREATE/DROP POLICY`/`ENABLE`/`FORCE` change — policy byte-identical (§0.1/§4 invariant held) |
| Regression vs M0/M1/M2 | all still pass | `TestCapabilityRegistrySize` PASS; 5 authz lints + M2 tripwire drift/parity lints green (live api-lint 0 violations); no route/contract shape regressed; cross-tenant isolation suites compile+green |

---

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Manual-seed discipline (Dimension-4 cross-cutting #2) | ~62 hand `SeedTxIdentity` acts; forget-one silently absorbed | chokepoint auto-seed + 21 reasoned allowlist entries + `SEED-CHOKEPOINT` blocking lint (census-drift = build-fail) | census 0 outside chokepoint+allowlist; lint RED on synthetic; **structurally** closed for sync |
| Async RLS backstop (Dimension-4 DEBT→) | worker/jobs seeded nothing; one bad join = silent leak, no gate | 5 processing txs + emitStuckAlert seed `SeedTxTenant`; `ASYNC-TENANT-SEED` blocking lint; negative RLS proof authored (live-run deferred) | lint RED on synthetic; §2.5-shaped proof compiles/vets; backstop engaged in code |
| RLS policy not weakened | NULL-permissive/FORCE/33 tables | identical | no SQL/migration/policy diff (C4) |

- **Root-cause vs symptom:** the seeding class is closed at the chokepoint + per-message primitive (root),
  not patched per-site — genuinely structural, **not** a symptom-patch.
- **Could it be built better?** Yes — the fix for F-1: derive `async-tenant-tables.txt` and the §2.4
  exemption rationale from the actual FORCE-RLS table set (which *includes* `idempotency_keys`) and label
  `idempotency_keys` honestly as a *sanctioned cross-tenant system-maintenance sweep under NULL-permissive*,
  not as "no tenant_id column." The **live run of the negative RLS proof** remains the standing bounded
  defer (no DB env; reading `.env` forbidden) — correctly recorded, not counted against this verdict.

---

## C6 — Forbidden-list

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean* (per-feature acceptance mapped in C1/C2)
- [ ] Fixture/mock passed off as real-provider proof — *clean* (F3.1 PG-1 explicitly labeled sqlmock-class; F3.2 negative proof labeled real-DB/testdb, run-deferred, not claimed as run)
- [ ] Consumer contract guessed — *clean*
- [x] **Split-brain (one fact, two sources of truth)** — **HIT (F-1):** `idempotency_keys` tenant-scoping fact stated two contradictory ways (reality: has `tenant_id` + FORCE-RLS; contract §2.4/§4 + ADR: "no `tenant_id`"), ADR contradicts its own body line 91
- [ ] Self-judged close / validator edited code — *clean* (validator only judged; wrote this file only)
- [ ] Scope drift — *clean* (diff matches the three features; docs-only F3.3 confirmed)
- [ ] Symptom-patch — *clean* (seeding closed at root)

**C6 result: one hit (split-brain, F-1) → contributes to FAIL.**

---

## C7 — Verdict

- **VERDICT: FAIL**
- **Failed checks:** **C1 (F3.3 acceptance §3.3/PG-2 not met)**, **C3 (split-brain)**, **C6 (split-brain hit)** — one root cause (finding **F-1**), an **HS-7** contract-accuracy defect propagated into the durable ADR 0027 record. Runtime is safe; the failure is durable-record / contract truth, which is F3.3's entire deliverable.
- **Minimum fix feature to open:** `f3.4-idempotency-keys-rls-truth`
  - Correct the false "`idempotency_keys` has no `tenant_id` column" claim in **all three** places it now
    lives: `validation-contract.md` §2.4 + §4 (re-open the contract WITH operator approval per HS-7 — do
    **not** silently edit), the ADR `wiki/decisions/0027-rls-adoption-sequencing.md` amendment (§4 line 183,
    per-binary line 202), and the `scripts/api-lint/async-tenant-tables.txt` rationale/derivation comment.
  - Re-classify `idempotency_keys` honestly: it **is** a FORCE-RLS `tenant_id`-bearing table; the
    idempotency janitor's cross-tenant `DELETE` is a **sanctioned system-maintenance sweep** that relies on
    the NULL-permissive GUC-unset hatch (same category as the audit-integrity scan), **not** a table where
    "RLS cannot apply."
  - Decide + record whether `idempotency_keys` belongs in `async-tenant-tables.txt` (if added, the janitor
    site is allowlisted with the sanctioned-sweep reason — verify `ASYNC-TENANT-SEED` stays green and no
    real per-tenant idempotency *write* path is left unseeded).
  - No behavior/policy change expected (RLS stays byte-identical); this is a truth-reconciliation fix.
- Milestone stays **active**; the main session does **not** advance to M4 and does **not** push. Re-dispatch
  this validator after `f3.4` closes (HS-4).

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending (blocked by FAIL)
> - Status flipped in `README.md`: **no** (only on PASS)
