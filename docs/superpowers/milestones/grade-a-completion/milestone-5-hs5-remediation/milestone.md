# Milestone 5 — HS-5 Remediation (close the re-audit gap)

> **Program:** grade-a-completion  
> **Status:** all 7 features CLOSED — awaiting milestone-validator (Phase 4) + HS-1 operator gate. F5.1–F5.5 CLOSED (H-G=0, H-D=0, Major #3+#4 closed); **F5.6 CLOSED as REFUTED** (Major #1 false-positive — effective_to is a soft-delete tombstone, not a future-dated interval; ADR 0037; HS-2 raised+operator-approved Option A, doc-only); **F5.7 CLOSED** (Major #2 fixed — role-admin iam_users upsert now carries tenant_id, was defaulting to the sentinel placeholder).  
> **Opened by:** HS-5 (mission-validator FAIL on terminal re-audit `architecture-re-audit-2026-06-16.md`)  
> **Re-audit artifact:** `wiki/backend/_artifacts/architecture-re-audit-2026-06-16.md`  
> **Governing spec:** `../mission.md §8`  
> **Appetite:** ≤1 session; all changes are surgical (port + literal replacement + typed response swap).

---

## Objective

Close the 4 §8 pass-bar gaps that survived the 2026-06-16 terminal re-audit so the next re-run passes:

1. **H-G → 0** (2 sites remaining)
2. **H-D → 0** (2 sites remaining)
3. **0 skeptic-confirmed Critical/Major** (4 remaining)
4. **module-boundaries ≥ A−** (currently B+) and **contract-api ≥ A−** (currently B+)

composition is already A− — do not touch.

---

## Appetite / rabbit holes

**In scope:** exactly the 7 features below — no wider refactor.  
**Out of scope:**
- No new product capabilities, no FE changes.
- Minor findings in `§7` of the re-audit (session rotation, audit handler 405, persistence minors, etc.) — carry as bounded defers; do not drag scope.
- Do not re-litigate refuted findings from F5.1 or the 2026-06-15 baseline.
- Do not touch composition/observability (already A−).

---

## Features

Executed in order. Each gets `spec.md → plan.md → evidence.md` before code.

| Feature | What to implement | What to validate (acceptance) | Closes |
|---------|-------------------|-------------------------------|--------|
| **F5.1** `templates-literal` | Replace hardcoded `"published"` at `internal/modules/templates/infrastructure/template_version_reader.go:44` with `templatesdomain.VersionStatusPublished` (same pattern as M4/F4.1 fix in `documents/application/service.go`) | H-G grep for hardcoded status literals returns 0 for this file; build + tests green | H-G site #2 |
| **F5.2** `auth-usertenant-port` | Create `UserTenantReader` port in `iamdomain` (interface + Noop); Postgres impl reads `iam_user_roles` (off-tx, H-PRE-1); wire into `auth/infrastructure/postgres/repository.go`; `GetUserTenants` consumes the port instead of direct JOIN | No `FROM metaldocs.iam_user_roles` in `auth/` outside `iam/`; H-G = 0; live or unit test proves parity | H-G site #1; module-boundaries → A− |
| **F5.3** `routes-generated-typed` | Replace `map[string]any{...}` at `routes_generated.go:128` and `:238` with the correct generated response struct(s) from `api.gen.go` | H-D grep returns 0 for these lines; typed responses match spec shape; build + tests green | H-D sites; H-D = 0 |
| **F5.4** `templates-routes-typed` | Replace `toTemplateResponse`/`toVersionResponse` helpers in `routes_create.go:44,67` with generated typed responses; or confirm helpers already produce correct typed output and the issue is the helper return type | No `map[string]any` on public template routes; contract-api indicatively A−; tests green | Major #3 |
| **F5.5** `iam-admin-typed` | Replace `map[string]any` admin overview response at `iam/delivery/http/admin_handler.go:262` with the generated IAM admin response type | No `map[string]any` on public IAM admin route; tests green | Major #4 |
| **F5.6** `authz-effective-to` | Fix `internal/modules/iam/authz/authz.go:124` — `effective_to IS NULL` predicate should be `(effective_to IS NULL OR effective_to > now())` to allow time-bounded active memberships (mirrors the M0/F0.1 `effective_from` fix on the same query) | Test: a time-bounded membership with `effective_to` in the future is **granted**; a membership with `effective_to` in the past is **denied**; existing authz tests green | Major #1 |
| **F5.7** `role-admin-tenant-id` | Fix `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:61-68` and `:109-116` — iam_users upsert must include `tenant_id`; identify the correct tenant source in the call path and pass it through | iam_users rows created via role_admin path carry the correct tenant_id; test proves it | Major #2 |

---

## Sequencing

```
F5.1 → F5.2    (H-G family — do together; both small)
F5.3 → F5.4 → F5.5    (H-D / contract family)
F5.6 → F5.7    (authz/persistence Majors)
```

All three families are independent — the session may interleave them if it helps, but each feature gets its own spec gate before code.

---

## Quality goals and constraints

1. **H-G = 0** after F5.1 + F5.2 — verify with the §6 grep commands from the re-audit.
2. **H-D = 0** after F5.3 — verify the same way.
3. **0 confirmed Majors** after F5.4–F5.7 — each fix has a test that would have caught it.
4. **H-PRE-1 respected** for F5.2: the new `UserTenantReader` port call must stay off any lock-holding tx (pool read only).
5. **ADR 0022**: F5.6 fixes the authz predicate at the shared layer, not per-caller.
6. **No scope creep**: fix only the 7 cited sites; mention-don't-fix any other issue found.

---

## Hard-stops

| ID | Trigger | Action |
|----|---------|--------|
| HS-1 | Milestone boundary | Operator gate before re-audit re-run |
| HS-2 | F5.2 growing into a shared IAM-API redesign | Stop; report boundary |
| HS-3 | Build / route / contract prerequisite fails | Repair first; then resume |
| HS-4 | milestone-validator FAIL | Open named fix feature; re-dispatch |
| HS-5 | Re-audit still misses bar | Another bounded micro-milestone; operator decides |

---

## Milestone validation definition

Run by the independent `milestone-validator` subagent after all 7 features have `evidence.md`.

1. **Spec/plan conformance:** each feature `spec.md` has an approval line before its commit; evidence acceptance rows map to re-runnable commands.
2. **Gates re-run from clean state:**
   - H-G grep: `grep -rn "FROM metaldocs\.iam_user_roles" --include="*.go" internal/modules/ | grep -v "internal/modules/iam/" | grep -v "_test\.go"` → 0
   - H-G literal: `grep -rn 'status.*:=.*"published"\|"published"' --include="*.go" internal/modules/templates/infrastructure/` | grep -v "_test\.go"` → 0
   - H-D grep: `grep -rn "map\[string\]any" internal/modules/templates/delivery/http/routes_generated.go` → 0 (or only in comments)
   - `go build ./...` → clean
   - `go test -count=1 ./...` → 0 FAIL
3. **Senior review** of the aggregate M5 diff (7 features, bounded surgical changes).
4. **Regression:** whole-repo `go test ./...` green; M0–M4 sentinels clean.
5. **Quality bar:** H-G = 0, H-D = 0, 0 confirmed Majors on these sites, module-boundaries ≥ A−, contract-api ≥ A−.
6. **Forbidden list:** no symptom-patching, no fixture-as-live-proof for live-SQL swaps (F5.2), no map[string]any still on public routes.
