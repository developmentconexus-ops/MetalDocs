# Milestone 3 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` + `../validation-contract.md` (D4, binding, re-opened once
> under HS-7 for F3.4) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-03 (HS-4 re-dispatch after F3.4)  ·  **HEAD:** `118e0196`  ·  **Baseline:** `e80c1800`
> (pre-impl contract commit)  ·  **Verdict:** see C7 → **PASS**.
> The validator judges and writes this file only; it edited no source, fixed no finding, flipped no status.
> (Synthetic RED-probe files inserted for lint negative-proof were removed; final tree clean.)

## Context — this is a re-validation

Prior run (@ `a6c874a3`) returned **FAIL** on finding **F-1**: three durable records falsely claimed
`idempotency_keys` has "no `tenant_id` column / RLS structurally N/A" (a split-brain contradicting the
schema and ADR 0027's own body). The main session opened + closed fix-feature **F3.4
(idempotency-keys-rls-truth)** — a docs + lint-data-only truth reconciliation, no behavior/policy change.
This run re-executes the full C1–C7 gate from clean state (all four features re-verified, not assumed).

## Inputs loaded (none missing — did not judge blind)

`milestone.md` · `validation-contract.md` (§0–§6 + the HS-7 F3.4 erratum head) · program `README.md` ·
governing `mission.md` (§7 M3) · all **four** features' `spec.md`/`plan.md`/`evidence.md`
(F3.1/F3.2/F3.3/F3.4) · aggregate diff `git diff e80c1800..118e0196` (79 files) across five commits
`de94758a` (F3.1) · `bad74d86` (F3.2) · `a6c874a3` (F3.3) · `118e0196` (F3.4).

---

## C1 — Spec & plan conformance (per feature)

| Feature | spec approved before code | interview populated | consumer contract honored | acceptance met | non-goals respected | Evidence |
|---------|---------------------------|---------------------|---------------------------|----------------|---------------------|----------|
| F3.1 | ✅ 2026-07-03 | ✅ Q&A | ✅ FORCE-RLS + `SEED-CHOKEPOINT` | ✅ | ✅ | C2/C3 |
| F3.2 | ✅ 2026-07-03 | ✅ Q&A | ✅ FORCE-RLS + `ASYNC-TENANT-SEED` + neg-proof | ✅ | ✅ | C2/C3 |
| F3.3 | ✅ 2026-07-03 | ✅ Q&A | ✅ **durable record now matches runtime truth** (F-1 closed) | ✅ | ✅ (docs-only) | C2/C3 |
| F3.4 | ✅ 2026-07-03 (operator, incl. HS-7 re-open) | ✅ 4-row Q&A | ✅ PG-1..4 (below) | ✅ | ✅ (no behavior/policy change) | this run |

All twelve artifacts (`spec.md`/`plan.md`/`evidence.md` × 4) exist and are execution-shaped (task lists,
files-touched, test strategy). Each `evidence.md` acceptance table maps to its `spec.md` Validation Gate.
Every approval line is filled (dated 2026-07-03). F3.4's spec/plan/evidence are consistent and scoped to
truth reconciliation.

**C1 result: PASS.** The single prior C1 miss (F3.3 acceptance §3.3/PG-2 "durable record matches runtime
truth") is now met — see F3.4 verification below.

### F-1 closure — verified (the fix under test)

F3.4's four gates re-verified independently by the validator, from source:

- **PG-1 — no residual false claim; `job_leases` intact.** Grepped `idempotency_keys` across all three
  durable records. Every occurrence now classifies it as a **`tenant_id`-bearing FORCE-RLS table** whose
  janitor TTL `DELETE` is a **sanctioned cross-tenant NULL-permissive maintenance sweep (audit-scan class)**:
  `validation-contract.md` §0.3 (line 77), §2.4 (line 228), §4 (line 310); ADR 0027 §4 residual list
  (line 189) + per-binary janitor row (line 212). No "no `tenant_id` / RLS N/A / cannot apply" is attributed
  to `idempotency_keys` anywhere. The `job_leases`-genuinely-no-`tenant_id` claim is intact and correctly
  distinct (contract §2.4 line 77 + §4; ADR 0027 line 187). ✅
- **PG-2 — set-equality (33 == 33).** Extracted FORCE tables from `db/baseline/0001_current_schema.sql`
  (schema-qualified, both `metaldocs.`/`public.`, normalized to bare name, sorted-unique) → **33**;
  non-comment entries of `scripts/api-lint/async-tenant-tables.txt` → **33**; `diff` **empty** (set-equal).
  `idempotency_keys` present in both. Ground truth confirmed at `0001_current_schema.sql:1330`
  (`tenant_id uuid NOT NULL`) + `:1347` (`FORCE ROW LEVEL SECURITY`). ✅
- **PG-3 — lint green + no false trip.** `go run ./scripts/api-lint -strict …` → **0 violation(s), exit 0**;
  `go test ./scripts/api-lint/ -run 'SeedChokepoint|AsyncTenantSeed'` → **14/14 PASS** (no hardcoded-32
  broke). Confirmed at source that `asyncHandlerRoots` (`async_tenant_seed_rule.go:45-52`) does **not**
  include `internal/modules/jobs/idempotency_janitor` nor `internal/platform/idempotency`, so listing
  `idempotency_keys` in the table set cannot false-trip the lint on the janitor's `DELETE`. ✅
- **PG-4 — no behavior diff.** `git show --name-only 118e0196` → only the two docs (contract + ADR 0027),
  the `.txt` lint data, the f3.4 folder, and the prior `qa/milestone-qa.md`. **No `.go`/`.sql`/migration
  file.** RLS byte-identical. ✅
- **HS-7 discipline.** The committed contract was re-opened **with operator approval** and the re-open
  recorded as a **dated erratum** at the contract head (`validation-contract.md:11-19`) and in the ADR
  amendment (`0027-…md:121-124`) — auditable, not a silent edit. The acceptance bar (§3.3/PG-2) was **not**
  weakened; only a false premise was corrected. ✅

---

## C2 — Gates re-run, isolated (validator ran these @ `118e0196`; not trusted from evidence)

| Gate | Command re-run | Real output | Pass? |
|------|----------------|-------------|-------|
| build | `go build ./...` | exit 0 | ✅ |
| build (integration tag) | `go vet -tags integration ./internal/modules/iam/authz/` | exit 0 (compiles + vets) | ✅ |
| core F3.1/F3.2/M2 pkgs | `go test ./internal/platform/db/... ./internal/platform/tenant/... ./internal/modules/iam/authz/... ./scripts/api-lint/... ./internal/modules/iam/domain/...` | all `ok` (api-lint 9.0s) | ✅ |
| live api-lint (both new lints blocking) | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | **0 violation(s)**, exit 0 | ✅ |
| lint unit tests | `go test ./scripts/api-lint/ -run 'SeedChokepoint\|AsyncTenantSeed' -v` | **14/14 PASS** (clean-green, stray-fires-named, after-removal-green, allowlist, stale-allowlist, sync-path-never-flagged, non-tenant-never-flagged) | ✅ |
| **F3.2 ASYNC-TENANT-SEED live RED→GREEN** | inserted throwaway unseeded `UPDATE documents` in `internal/platform/worker/zz_validator_probe.go` | `ASYNC-TENANT-SEED … zz_validator_probe.go:5 (op UPDATE, table documents)`, **1 violation, exit 1**; removed → **0, exit 0** | ✅ |
| **PG-2 set-equality (F3.4)** | `diff <(FORCE tables normalized sort -u) <(txt non-comment sort -u)` | **IDENTICAL, 33 == 33**; `idempotency_keys` in both | ✅ |
| **PG-4 no-behavior (F3.4)** | `git show --name-only 118e0196 \| grep -E '\.(go\|sql)$\|migration'` | **NONE** | ✅ |
| F3.2 negative-RLS proof | `ls` + `head` + `go vet -tags integration …` on `internal/modules/iam/authz/seed_tx_tenant_rls_integration_test.go` | **exists**, `//go:build integration`, `testdb` factory, §2.5 shape (2 tenants, NOBYPASSRLS role, leak-before → row visible + UPDATE affects 1, blocked-after `SeedTxTenant(A)` → 0 rows / `42501`); compiles + vets; **run deferred** (no DB env — accepted bounded defer §2.5/§6) | ✅ (authored) |
| M2 regression | `go test ./internal/modules/iam/domain/ -run CapabilityRegistrySize` (within core run) | `ok`; live api-lint 0 violations ⇒ 5 authz lints + M2 tripwire drift/parity lints still green | ✅ |

Throwaway probe file deleted; final `git status` shows no probe residue (only untracked `docs/release/`).
No gate regressed. All C2 commands passed.

---

## C3 — Senior review of the aggregate milestone diff

Reviewed `git diff e80c1800..118e0196` as one unit.

- **No schema/RLS change (whole milestone):** `git diff --name-only e80c1800..118e0196` yields **no**
  `.sql`, migration, or `db/**` schema file (the only `db/`-path hits are `internal/platform/db/runner.go`
  + its test — Go chokepoint code, not schema). The RLS policy is **byte-identical** before/after M3
  (NULL-permissive, FORCE, 33 tables). §0.1/§4 invariant held.
- **F3.1 chokepoint / collapse / F3.2 primitive / 5 seed sites / negative-RLS proof:** unchanged from the
  prior run's senior review — all senior-clean (module boundary honored, H-PRE-1 preserved, single-tenant
  per tx, no HS-2 tenant-mix). No regression introduced by F3.4 (docs/data only).
- **F-1 split-brain — RESOLVED.** The sole prior split-brain (the `idempotency_keys` exemption fact stated
  two contradictory ways) is closed: reality (`has tenant_id` + FORCE-RLS, sweeps cross-tenant via
  NULL-permissive) and the durable records (contract §0.3/§2.4/§4 + ADR 0027 §4/per-binary + `.txt`) now
  state **one** fact. The ADR no longer contradicts its own body (`0027-…md:91` lists `idempotency_keys`
  among `tenant_id`-carrying tables; the amendment now agrees). `async-tenant-tables.txt` is the honest
  33-table FORCE mirror.
- **Staff-engineer bar:** **met** on both mechanism (F3.1/F3.2 code) and durable-record truth (F3.3 as
  corrected by F3.4). No duplication, no dead code, no feature broke another.

**C3 result: PASS.** No blocking finding; the prior split-brain is closed.

---

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api multi-tenant) | pass-with-defers | both new lints blocking + 0 live violations; RED→GREEN captured; core tenant/db/authz pkgs green; full 20-min suite intentionally not run (mission §10) |
| RLS policy byte-identity | pass | aggregate diff touches **no** `db/**`/migration/`*.sql`; no `CREATE/DROP POLICY`/`ENABLE`/`FORCE` change (§0.1/§4 invariant held) |
| Regression vs M0/M1/M2 | all still pass | `TestCapabilityRegistrySize` `ok`; 5 authz lints + M2 tripwire drift/parity lints green (live api-lint 0); no route/contract shape regressed; F3.4 added no code path to regress |

---

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Manual-seed discipline (Dim-4 cross-cutting #2) | ~62 hand `SeedTxIdentity` acts; forget-one silently absorbed | chokepoint auto-seed + reasoned allowlist + `SEED-CHOKEPOINT` blocking lint | census 0 outside chokepoint+allowlist; lint RED on synthetic; structurally closed for sync |
| Async RLS backstop (Dim-4 DEBT→) | worker/jobs seeded nothing; one bad join = silent leak, no gate | 5 processing txs seed `SeedTxTenant`; `ASYNC-TENANT-SEED` blocking lint; negative RLS proof authored | lint RED on synthetic (verified live); §2.5-shaped proof compiles/vets; backstop engaged in code (live run = bounded defer) |
| Durable-record truth (F-1, the fix under test) | 3 records falsely claimed `idempotency_keys` has no `tenant_id`/RLS N/A (split-brain vs schema + ADR body) | all 3 records reclassify it honestly as FORCE-RLS `tenant_id`-bearing, janitor sweep = sanctioned audit-scan-class NULL-permissive DELETE | PG-1 grep clean; PG-2 33==33 set-equal; ADR no longer self-contradicts; HS-7 erratum dated + auditable |
| RLS policy not weakened | NULL-permissive/FORCE/33 tables | identical | no SQL/migration/policy diff (C3/C4); PG-4 confirms F3.4 added none |

- **Root-cause vs symptom:** seeding closed at chokepoint + per-message primitive (root, sync + async);
  F-1 closed by correcting the false schema premise at every durable source of truth (not masked). The
  `.txt` is now derived from the actual 33-table FORCE set, so future drift is a set-equality failure, not
  a silent lie.
- **Could it be built better?** The one remaining standing item is the **live run of the negative RLS
  integration proof** — authored + shape-verified, execution deferred (no DB env; `.env` read forbidden),
  a correctly-recorded bounded defer (§2.5/§6), not counted against this verdict. Trigger: run on CI/a
  capable box before program close-out.

---

## C6 — Forbidden-list

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean* (each feature's acceptance mapped in C1/C2; per-gate commands + outputs recorded)
- [ ] Fixture/mock passed off as real-provider proof — *clean* (F3.2 negative proof explicitly labeled real-DB/testdb, run-deferred, **not** claimed as run; lint tests labeled AST-static)
- [ ] Consumer contract guessed — *clean* (F3.4 consumer = maintainer/validator reading the enforcement record; contract read from schema source, not guessed)
- [ ] Split-brain (one fact, two sources of truth) — **clean (prior HIT F-1 now RESOLVED)** — `idempotency_keys` stated one consistent way across schema, contract, ADR, and `.txt`
- [ ] Self-judged close / validator edited code — *clean* (validator only judged; wrote this file only; probe files removed)
- [ ] Scope drift — *clean* (F3.4 diff = docs + `.txt` only, matches its spec; PG-4 no code/SQL)
- [ ] Symptom-patch — *clean* (F-1 fixed at every durable source of truth + the `.txt` now derived from the real FORCE set; not masked)

**C6 result: all clean — no hit.**

---

## C7 — Verdict

- **VERDICT: PASS**
- **All checks C1–C6 pass.** The prior FAIL finding **F-1** (false `idempotency_keys` no-`tenant_id`
  claim / split-brain, an HS-7 contract-accuracy defect) is **closed** by fix-feature **F3.4**, verified
  independently against source: PG-1 (no residual false claim; `job_leases` intact), PG-2 (`.txt` == the
  33 FORCE-RLS tables, set-equal), PG-3 (api-lint 0 + 14/14 lint unit tests; janitor outside scanned roots
  → no false trip), PG-4 (no `.go`/`.sql`/migration diff; RLS byte-identical). HS-7 discipline honored
  (operator-approved re-open, dated auditable erratum, acceptance bar unchanged).
- **All four M3 acceptance themes met:** F3.1 (census 0 outside chokepoint+allowlist; `SEED-CHOKEPOINT`
  lint blocking/RED-on-synthetic) · F3.2 (5 `SeedTxTenant` sites; `ASYNC-TENANT-SEED` lint verified
  RED→GREEN live; negative-RLS integration proof exists + §2.5-shaped, live run a labeled bounded defer) ·
  F3.3 (ADR 0027 dated amendment + live tenancy wiki, now truth-accurate) · F3.4 (F-1 closed per PG-1..4).
- **Standing bounded defer (not a failure):** live execution of the F3.2 negative-RLS integration proof
  (no DB env) — authored + shape-verified; trigger = CI/capable box before program close-out.
- Handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending (main session to present)
> - Status flipped in `README.md`: main session's action, only now that the verdict is PASS
