# Milestone 6 — HS-5 Contract Sweep (drive contract-api ≥ A−, H-D Grep A → 0)

> **Program:** grade-a-completion
> **Governing spec:** `../mission.md` (§8 terminal acceptance; §2 Non-Goals; §9 HS-2 watch)
> **Status:** Spec approved — Executing
> **Authored:** 2026-06-19 — before any F6.x feature began.
> **Opened by:** HS-5 — mission-validator FAIL on terminal re-audit `architecture-re-audit-2026-06-19.md` (`qa/mission-validation.md`, HEAD `ad8e6fc8`).
> **Appetite:** ≤1 session. All changes are bounded: typed-response swaps + 9 OpenAPI 200 schema declarations + 1 BE codegen regen + 1 FE codegen regen.

---

## Objective

Close the 3 §8 pass-bar gaps that survived the 2026-06-19 terminal re-audit so the next re-run passes:

1. **Contract/API ≥ A−** (currently **B−** — regressed from B+ on 2026-06-16 under the broader fresh read).
2. **0 skeptic-confirmed Critical/Major** (currently **5**, all in the Contract/API root-cause family).
3. **H-D Grep A = 0** (currently **24** sites — M5 closed the 2 prior-cited spot sites but did not sweep the class).

H-G is already 0 (PASS in the 2026-06-19 verdict) — **do not regress F5.1/F5.2 fixes**.
Module-boundaries (A−), composition (A−) — already lifted, do not touch.

---

## Appetite / rabbit holes

**In scope:** exactly the 6 features below — no wider refactor.

**Out of scope (YAGNI-ruthless):**
- No FE feature work beyond the codegen-type regen the contract fixes require.
- No new product capabilities.
- No Minor sweep from the 2026-06-19 re-audit §7 unless it sits inside a touched file (drive-by repair only — record in evidence).
- No re-litigating refuted findings (re-audit §5).
- No new authz/composition/observability work — those grades are lifted.
- No migration / schema work.
- **HS-2 watch:** Contract/API has missed twice (B+ on 2026-06-16, B− on 2026-06-19). If the M6 features grow into a generalized codegen-first StrictServerInterface adoption across modules, **stop** — that is HS-2 / next-mission territory, not M6.

---

## Features

Executed in order. Each gets `spec.md → plan.md → evidence.md` before commit. F6.5 closes the iam routes_memberships sites **in-scope** (operator decision 2026-06-19 over the ADR-0012 bounded-defer option permitted by §5 / §7 Minor #30).

| Feature | What to implement | What to validate (acceptance) | Closes |
|---------|-------------------|-------------------------------|--------|
| **F6.1** `templates-lifecycle-typed` | Declare 200 body schemas in `api/openapi/v1/openapi.yaml` for `submitTemplateVersion`, `reviewTemplateVersion`, `archiveTemplate`, `upsertTemplateApprovalConfig`; align `approveTemplateVersion` to its already-declared `ApproveTemplateVersionResponse`. Run `GOFLAGS=-mod=mod go generate ./internal/modules/templates/api/...`. Swap the 5 `writeJSON(...map[string]any...)` sites at `routes_lifecycle.go:46,100,164,196,239` to emit the generated `*Response` types. | OpenAPI declares 200 body schema for all 5 ops. Generated `api.gen.go` types match. Handler emits the generated type. Handler tests assert the wire JSON keys equal the OpenAPI declaration. H-D grep on `routes_lifecycle.go` = 0. | 5 Major sites; OpenAPI tri-source drift |
| **F6.2** `templates-query-typed` | Declare 200 body schemas in OpenAPI for `listTemplates`, `getTemplate`, `getTemplateVersion`, `getSystemBlankTemplate` (if currently in spec), `getTemplateDocxUrl`, `listTemplateAudit`. Regen BE codegen. Swap the 5 sites at `routes_query.go:73,104,145,211,260` to emit generated types. | OpenAPI declares 200 body schemas. Handler emits generated type. Handler tests assert wire JSON keys ⊆ OpenAPI declaration. H-D grep on `routes_query.go` = 0. | 5 Major sites; OpenAPI tri-source drift |
| **F6.3** `iam-handlers-typed` | IAM module is pre-codegen on the BE side (ADR 0012 partial rollout). Define hand-rolled typed Go response structs in `iam/delivery/http/` for the response envelopes at: `admin_handler.go:341,378` (UpsertUserAndAssignRole, handleReplaceUserRoles); `sessions_handler.go:132,138,158` (sessions list rows + envelope + page meta); `observability_handler.go:81,109` (usageToJSON, kpiToJSON helpers). Swap `map[string]any` → struct. | No `map[string]any` literal in the 5 swapped functions. Handler tests assert wire JSON shape unchanged. Build + tests green. | 5 Major sites (admin + sessions + observability) |
| **F6.4** `class-sweep-typed` | Class-sweep the remaining H-D Grep A sites (none in scope as new Majors — but the class count must reach 0). Define typed Go response structs (or use module codegen where available) for: `security/delivery/http/handler.go:67,94,107,130,173`; `taxonomy/delivery/http/routes_areas.go:40`; `taxonomy/delivery/http/routes_families.go:31`; `templates/delivery/http/routes_catalog.go:35`; `templates/delivery/http/routes_schema.go:68`. | No `map[string]any` literal in the 9 swapped functions. Build + tests green. | 9 H-D class sites (security 5 + taxonomy 2 + templates catalog 1 + templates schema 1) |
| **F6.5** `iam-memberships-typed` | Hand-rolled typed Go structs for `iam/delivery/http/routes_memberships.go:168` (list memberships envelope) and `:235` (create-membership 201 envelope). Operator decision 2026-06-19: **close in-scope** rather than ADR-0012 defer (cleanest exit bar). | No `map[string]any` literal in the 2 swapped functions. Handler tests assert wire shape unchanged. H-D Grep A = 0 absolute (no allowlist). | 2 H-D sites (memberships list + create) |
| **F6.6** `fe-codegen-regen-final-proof` | After F6.1–F6.5 commit clean: regen FE codegen for templates module (and any other modules whose `api.gen.go` changed under F6.1/F6.2), in contract-first order (BE codegen first → FE codegen second). Update `wiki/architecture/api-contract.md` Last-verified stamp. Final §6 Grep A proof — must show 0 hits. | FE codegen regen clean (no orphan types, no breaking diffs left uncommitted). Wiki Last-verified stamp current. `grep -rEn 'writeJSON.*map\[string\]any' internal/modules/*/delivery/http/ --include='*.go'` returns 0. Whole-repo `go test -count=1 ./...` green. | Exit bar proof |

---

## Sequencing

```
F6.1 → F6.2                  (templates contract-first: OpenAPI + BE codegen)
F6.3 → F6.4 → F6.5           (hand-rolled / class-sweep typed structs; independent)
F6.6                          (FE codegen regen + wiki stamps + final grep proof)
```

F6.3/F6.4/F6.5 are independent of each other but depend on F6.1/F6.2's BE codegen regen being committed clean. F6.6 depends on all five prior features being closed.

---

## Quality goals and constraints

1. **H-D Grep A = 0** after F6.6 — verified with `grep -rEn 'writeJSON.*map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' | wc -l`.
2. **0 confirmed Majors** after F6.1–F6.3 — each Major hot-site is replaced by a typed response with a handler test asserting wire-JSON parity.
3. **Contract-first regen order respected** (F6.1, F6.2, F6.6) — OpenAPI → BE `api.gen.go` → FE codegen. Never the reverse.
4. **No prior-milestone regression** — H-G stays at 0; F5.1 (`templates/infrastructure/template_version_reader.go:45`) and F5.2 (`auth/infrastructure/postgres/repository.go:117`) untouched.
5. **No scope creep** — fix only the cited sites + the contract-first regen they require. Mention-don't-fix anything else found.
6. **HS-2 watch**: if any feature implies redesign beyond the assigned boundary (e.g. converting IAM to codegen-served, generalizing StrictServerInterface adoption), STOP and surface as HS-2.
7. **Authz untouched** — ADR 0022 boundary respected; F0.1 / F5.6 fixes untouched.
8. **Never push** — commits authorized after verified work (CLAUDE.md §5.0), operator merges.

---

## Hard-stops

| ID | Trigger | Action |
|----|---------|--------|
| HS-1 | Milestone boundary | Operator gate before terminal re-audit re-run |
| HS-2 | A feature grows into codegen-first redesign across modules, or generalized StrictServerInterface adoption beyond the assigned 24 sites | STOP; surface boundary; propose as successor mission; no symptom-patch within M6 |
| HS-3 | Build / `go generate` / `go test` prerequisite fails (e.g. stale codegen, vendor mismatch) | Repair (`runtime-contract-prereq`); rerun checkpoint; resume feature |
| HS-4 | milestone-validator returns FAIL | Open the validator's named fix feature; re-run its lifecycle; re-dispatch |
| HS-5 | Terminal re-audit (post-M6) misses §8 bar a third time | Per mission.md §9 HS-2 signal: surface to operator as HS-2 redesign boundary; do not open a bounded M7 by default |
| HS-6 | Off-plan discovery mid-milestone (e.g. a swap surfaces a new bug or a runtime/spec drift outside the 24 sites) | STOP; surface; replan before continuing |

---

## Milestone validation definition

Run by the independent `milestone-validator` subagent after all 6 features have `evidence.md`.

1. **Spec/plan conformance** — each `f6.x-<slug>/spec.md` carries a pre-commit approval line; each `evidence.md` Acceptance row maps to a re-runnable command.
2. **Gates re-run from clean state:**
   - H-D Grep A: `grep -rEn 'writeJSON.*map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' | wc -l` → **0**
   - H-G Grep B/C (re-audit §6) → **0** (no regression of F5.1/F5.2)
   - `GOFLAGS=-mod=mod go generate ./internal/modules/templates/api/...` → no uncommitted diff
   - `go build ./...` → clean
   - `go test -count=1 ./...` → 0 FAIL (full repo, from clean state)
   - Per-feature handler tests pass — wire-JSON keys equal the OpenAPI declaration (F6.1/F6.2) or unchanged from baseline (F6.3/F6.4/F6.5)
3. **Senior review** of aggregate M6 diff — only the 24 cited sites + the 9 OpenAPI 200 schema declarations + the 2 regen artifacts (BE `api.gen.go`, FE codegen).
4. **Regression** — M0–M5 sentinels green; H-G greps still 0; F5.1/F5.2/F5.6/F5.7 untouched.
5. **Quality bar:** H-D Grep A = 0 absolute (no defer allowlist). Contract/API indicatively A− on dimension re-read. Forbidden list:
   - No `map[string]any` literal surviving on any public delivery route (`internal/modules/*/delivery/http/`).
   - No symptom-patching (e.g. `writeJSON(...interface{}(map[string]any{}))` or other grep-evasion).
   - No fixture/mock proof passed off as real-provider proof.
   - No codegen regen committed without matching OpenAPI source change.

---

## Dependencies & constraints respected

- **Mission §2 Non-Goals:** no FE feature work, no new capabilities, no schema/migration redesign, no merge by agent.
- **Mission §3 D3 sequencing:** M6 is the bounded contract-sweep micro-milestone (HS-5 child of M5 terminal acceptance) — does not reorder M0..M5.
- **CLAUDE.md:** never read/print/commit `.env`; PowerShell for local startup; runtime truth beats docs; commits authorized but **never push**.
- **ADR 0012 (contract-first API):** F6.1/F6.2/F6.6 follow contract-first regen order. F6.3/F6.4/F6.5 use hand-rolled typed structs (legacy posture preserved for IAM + cross-module non-codegen routes).
- **ADR 0022 (authz boundary):** untouched.
- **H-PRE-1 (advisory-lock hazard):** untouched (no new authz-recording reads in this milestone).
- **Memory `backend-target-architecture-governs-reviews`:** PRs cite REQ IDs; MUST deviations need an ADR. M6 does not introduce a MUST deviation.
- **Memory `workflow-model-balancing`:** sonnet for implementation/review, haiku mechanical, never fable, ≤15 concurrent.
- **Skill routing (CLAUDE.md §3):** backend HTTP/contract → `metaldocs-backend-api` if invoked; FE codegen → `metaldocs-tanstack-query`; prereq repair → `runtime-contract-prereq`; after structural change → `wiki-curator`.
