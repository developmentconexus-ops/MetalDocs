# Mission: Grade-A Completion (close the F5.1 re-audit gap)

> **Status:** Drafting (awaiting operator Phase-4 approval)
> **Date:** 2026-06-15  ·  **Branch of record:** main
> **Type:** remediation
> **Slug:** `grade-a-completion`  ·  **Owner / operator:** leandrotca.work
> **Evidence base:** `./discovery-brief.md`  ·  **Program index:** `./README.md`
> **Parent:** the `grade-a-architecture-remediation` program — this mission **is** that program's F5.2
> remediation under M5/HS-5. Its terminal acceptance (re-run F5.1) closes the parent's M5 and declares Grade A.
> **Governs:** Milestones M0..M4 below. Each milestone gets its own plan via the `milestone` skill, executed
> in a fresh session. This file is the **stable governing contract** — *what* the mission is and *what proves
> it done*; it contains **no execution detail**.

---

## 1. Problem / why now

The `grade-a-architecture-remediation` program's authoritative gate — the **F5.1 independent re-audit**
(`wiki/backend/_artifacts/architecture-re-audit-2026-06-15.md`, `Workflow wf_b0109977-23a`, 42 agents,
skeptic-per-finding, at HEAD `02ed1c24`) — returned **VERDICT: MICRO-WAVE NEEDED**. All four §6 pass-bar
checks fail: the 3 formerly-C dimensions are below A− (module-boundaries **B+**, contract-api **C+**,
composition **B+**), there are **21** skeptic-confirmed Critical/Major findings (target 0), **H-D = 4**, and
**H-G = 1** (plus one new boundary site the M4 census missed). The earlier per-milestone slices closed the
*instances they censused* but the fresh independent read found the **classes** were not fully swept and
surfaced new defects. Grade A is therefore **not** independently provable today. This mission closes that gap
at the class level so a re-run of the re-audit passes. Why now: it is the last thing standing between the
program and operator Grade-A sign-off. Every claim here traces to `./discovery-brief.md`.

## 2. Goals / Non-Goals

**Goals**
- All 3 formerly-C dimensions independently re-graded **≥ A−** (module-boundaries, contract-api, composition).
- **0** skeptic-confirmed Critical/Major on a fresh re-audit (the 21 closed by root-cause family).
- **H-D class = 0** (handler/contract tri-source drift) and **H-G class = 0** (cross-module reach-without-a-port + hardcoded domain-state), proven by the report §6 grep commands.
- The 3 HS-2 redesign-boundary candidates **fully closed** (D2): IAM role port (C2), MfaCoverage port retiring the M4 defer (C3), meaningful app-level OTel spans (D3).
- Minor findings swept in-scope (D4 decision) or carried as bounded defers with written triggers.

**Non-Goals** (YAGNI-ruthless)
- **No FE feature work** beyond the codegen-type regen the contract fixes require (no new screens/components).
- **No re-litigating the 8 refuted/downgraded findings** (report §5) — out of scope unless new evidence appears (e.g. the WriteTimeout-vs-WebSocket claim, the cross-tenant revoke claim).
- **No new product capabilities** — this is remediation to a bar, not enhancement.
- **No schema/migration redesign** — the legacy-cluster teardown is already done (M4b `071931c9`).
- **No gold-plating observability** beyond what an A− composition grade requires (meaningful spans on DB + critical flows, not exhaustive instrumentation).
- **No merge by the agent**; the operator merges (house rule).

## 3. Locked decisions (operator-approved)

| # | Decision | Value |
|---|----------|-------|
| D1 | Scope | **Full-close, ports last** — close all 21 confirmed findings + the H-D and H-G classes; systemic module-ports sequenced last (risk-isolation). |
| D2 | HS-2 candidates | **Full-close all 3** — build the IAM role port (C2), build the MfaCoverage port retiring the M4 accepted defer (C3), and add meaningful app-level OTel spans (D3, to a pragmatic A− bar). |
| D3 | Sequencing | **authz → contract → observability → quality → ports** (M0→M4). Security/correctness first; FE-visible contract second; observability + quality-tail middle; systemic module-ports **last** so they cannot regress an already-lifted grade. |
| D4 | Minors + terminal proof | **Fix Minors in-scope** within their family milestone; **terminal acceptance = re-run the same F5.1 10-dimension re-audit** method and pass §6. |
| D5 | Execution | One `mission.md` governs; per-milestone plans via `milestone`; **fresh session per milestone**; inter-milestone HS-1 operator gates; no agent merge. |
| D6 | Linkage | Terminal PASS closes the **parent program's M5** and is the basis for operator Grade-A sign-off; this mission's `mission-validator` verdict is the parent's terminal evidence. |

## 4. Discovery summary

Discovery was the program's own F5.1 re-audit — a fresh, cited, adversarially-verified read of all 10
dimensions at HEAD `02ed1c24`; **code is unchanged since** (`git diff 02ed1c24..5ce0cffb -- '*.go'` empty),
so every `file:line` is current. Confidence is **high**: only skeptic-**confirmed** findings are carried; the
8 refuted/downgraded are excluded. The findings cluster into 5 root-cause families (A contract, B
auth/authz/session, C module-ports, D observability, E quality). See `./discovery-brief.md`.

## 5. Work / requirement inventory

Every confirmed finding + class site, mapped to a milestone. (Family letters per the brief.)

| # | Item (site) | Class / kind | Milestone |
|---|-------------|--------------|-----------|
| B1 | `iam/authz/authz.go:123` — `Require` ignores `effective_from` (premature access) | authz correctness (security) | M0 / F0.1 |
| B2 | `controlleddocuments/application/service.go:173` — manual-code create no tx-identity seed | authz/functional | M0 / F0.2 |
| B3 | `documents/approval/application/read_service.go:68` — tenant-grade view narrowed to area-grade | authz correctness | M0 / F0.3 |
| B4 | `auth/delivery/http/handler.go:153` — ChangePassword no expired-cookie | session hygiene | M0 / F0.4 |
| A1 | `documents/delivery/http/handler.go:881` — checkpoints untagged PascalCase (**Critical**) | contract / H-D-adjacent | M1 / F1.1 |
| A2 | `documents/delivery/http/handler.go:519` — renameDocument raw domain body | contract | M1 / F1.2 |
| A4 | `templates/delivery/http/routes_create.go:36` — 201 vs spec 200 (H-D) | contract / H-D | M1 / F1.2 |
| A5 | `templates/delivery/http/routes_autosave.go:42` — 201 vs spec 200 (H-D) | contract / H-D | M1 / F1.2 |
| A3 | `templates/delivery/http/routes_generated.go:64` — undeclared `id`/`version_id` (H-D) | contract / H-D | M1 / F1.3 |
| A6 | `documents/delivery/http/handler.go:816`(+5) — pervasive `map[string]any` bypass | contract class | M1 / F1.4 |
| A7 | `taxonomy/delivery/http/routes_profiles.go:67,111,126,169` — raw `domain.DocumentProfile` (H-D) | contract / H-D | M1 / F1.4 |
| A8 | `documents/delivery/http/handler.go:317`; `audit/delivery/http/handler.go:81` — raw stats; 405 no Allow | contract (minor) | M1 / F1.4 |
| D1 | `jobs/scheduler/scheduler.go:131` — hardcoded text logger | observability | M2 / F2.1 |
| D2 | `jobs/scheduler/scheduler.go:273` — job metrics not exposed | observability | M2 / F2.2 |
| D3 | `platform/observability/otel.go:95` — no app-level spans | observability | M2 / F2.3 |
| D4 | `platform/bootstrap/api.go:99`; `cmd/metaldocs-api/permissions.go:95` — pool stats absent; misleading comment | observability (minor) | M2 / F2.4 |
| E1 | `cmd/metaldocs-api/main.go:413` — IAMUserOptions never wired (functional) | dead-wiring/functional | M3 / F3.1 |
| E2 | `documents/application/freeze_service.go:77` — `fanoutClient any` | code-quality | M3 / F3.2 |
| E3 | `documents/application/service.go:433` — dead `userID` param | code-quality | M3 / F3.2 |
| E4 | `documents/application/freeze_service.go:194` — `_ = snap` discard | code-quality | M3 / F3.3 |
| E5 | `platform/objectstore/template_keys.go:5` — dead exported keys | dead-code | M3 / F3.4 |
| E6 | report §7 minors (dup constructors, PT string, 8× `tenantIDFromRequest`, SHA-1 dedup, …) | code-quality (minor) | M3 / F3.5 |
| C1 | `documents/application/service.go:282` — hardcoded `"published"` (**H-G**) | module-boundary / H-G | M4 / F4.1 |
| C2 | `security/infrastructure/postgres/repository.go:345` — JOIN `iam_user_roles`, no port | module-boundary (new) | M4 / F4.2 |
| C3 | `security/infrastructure/postgres/repository.go:67` — reads `iam_users` (M4 defer) | module-boundary | M4 / F4.3 |
| C4 | `documents/repository/repository.go:16` — `templates/domain.Placeholder` cross-seam type | module-boundary | M4 / F4.4 |

**Out-of-scope:** the 8 refuted/downgraded findings (report §5) — excluded by D1/Non-Goals. The M1
full-HTTP E2E SKIP (F5.1 coverage caveat) is re-exercised by the terminal re-audit, not a separate item.

## 6. Program architecture (by reference)

Executes via the `milestone` skill. The per-feature close-out loop, the consumer-contract spec gate, and the
per-milestone `milestone-validator` gate are defined there — not duplicated here
(`.claude/skills/milestone/SKILL.md`). Program-scale shape:

```
Mission: grade-a-completion
└── Milestone (M0..M4)        ── each: features → milestone-validator gate → HS-1 operator gate
    └── Feature (Fx.y)        ── each: spec(consumer-contract) → plan → TDD → evidence
Terminal acceptance (§8)      ── main session runs the F5.1 re-audit fan-out; independent
                                 mission-validator judges it, after M4 passes
```

## 7. Milestones

Order per D3: dependencies/correctness first, systemic ports last. No execution detail.

### M0 — Auth / authz / session correctness
**Objective:** close the 4 confirmed auth/authz/session defects at their shared root, lifting the authz and
sessions dimensions and removing the latent premature-access security bug. Bar: each defect has a regression
test that fails before and passes after; no authz regression.

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F0.1 authz-effective-from | `authz.Require` honors `effective_from <= now()` (matching `ResolveEligibleActors`), at the shared authz layer | Integration test: a future-dated membership is **denied**; a current membership granted; existing authz tests green |
| F0.2 manual-code-create-identity | The manual-code CD-create branch seeds the tx identity so non-admin creates pass the PEP/PDP | Integration test: non-system-admin manual-code create **succeeds**; system-admin path still works |
| F0.3 tenant-grade-view | `CapDocumentView` no longer narrows tenant-grade to area-grade when an area code is present | Test: a tenant-role-only viewer can read a document that has a real area code |
| F0.4 changepassword-cookie | Self-service ChangePassword emits an expired session cookie (mirrors `AdminResetPassword`) | Handler test: response sets an expired cookie; sessions revoked |

**Milestone gate:** `backend-api-qa-checklist` + authz correctness; whole-repo `go test ./...` green;
root-cause criterion — F0.1 fixed at the shared authz predicate, not per-caller. `milestone-validator`.

### M1 — Contract / API integrity
**Objective:** drive the **contract-api** dimension to **≥ A−** and the **H-D class to 0** — every handler
emits its declared generated response type with the spec-declared status code; no raw domain structs or
`map[string]any` on public routes. FE codegen regenerated (contract-first order).

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F1.1 checkpoints-typed | List/create checkpoint endpoints serialize the generated snake_case response type | Wire JSON keys are snake_case matching FE codegen (`index.d.ts`); contract test passes |
| F1.2 status-and-body-conformance | `renameDocument` → 200 no-content per spec; `createNextVersion` + `presignTemplateAutosave` → 200 | Handler tests assert status code + body shape equal the OpenAPI declaration |
| F1.3 declared-fields-only | `createTemplate` returns only the declared schema (drop undeclared `id`/`version_id`) | Response ⊆ spec schema; FE codegen consumers get the declared shape |
| F1.4 typed-responses-class | Replace all `map[string]any` / raw-domain emits with generated response types across the 6+ documents sites, taxonomy `routes_profiles`, `documentStats`; add the audit 405 `Allow` header | **H-D grep (report §6 commands) returns 0**; FE codegen regen clean; no raw-domain/map emit on any public route |

**Milestone gate:** `backend-api-qa-checklist` + contract truth-table (runtime/spec/codegen/wiki) + FE
codegen regen order; **H-D class-zero** proven by the report §6 greps; contract-api indicatively A−.
`milestone-validator`.

### M2 — Composition / observability
**Objective:** drive the **composition** dimension to **≥ A−** — the composition root injects real
observability everywhere, scheduler honors the JSON logger, job metrics are scrapeable, and app-level traces
exist on DB + critical flows.

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F2.1 scheduler-slog | Scheduler uses the injected `slog.Default()` JSON logger, not a hardcoded text handler | Scheduler log lines are JSON at runtime; no `NewTextHandler` in scheduler |
| F2.2 scheduler-metrics | Wire `MetricsSnapshot()` (run/error/skip counters) to the `/api/v1/metrics` scrape path | Metrics endpoint exposes per-job counters at runtime |
| F2.3 otel-app-spans | Add meaningful app-level OTel spans on DB calls + critical request flows (pragmatic A− bar) | Traces show child spans on key paths (not just the HTTP envelope) |
| F2.4 metrics-completeness | DB pool stats in `/api/v1/metrics`; correct the misleading "Prometheus" comment/endpoint | Pool stats present in the metrics payload; comment/endpoint accurate |

**Milestone gate:** `backend-api-qa-checklist` + observability; runtime proof of JSON logs + metrics +
spans; composition indicatively A−. `milestone-validator`.

### M3 — Code-quality & dead-code tail
**Objective:** lift the **code-quality** and **legacy/dead-code** dimensions by closing the confirmed
quality defects (one is functional) and sweeping the Minor tail.

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F3.1 wire-iam-user-options | Wire the `IAMUserOptions` dependency so placeholder user-type lookups return real data | Integration test: placeholder-options for user-type returns a non-empty list |
| F3.2 type-safety-deadparam | Give `NewFreezeService` a typed `FanoutClient` param; remove the dead `userID` from `ListDocumentComments` (verify no missing authz scope) | Compiles with the typed param; interface no longer carries the dead param; authz scope confirmed present |
| F3.3 snapshot-read | `Pin` stops fetching-then-discarding the snapshot blob | `Pin` reads only what it uses; no `_ = snap` discard of a heavy read |
| F3.4 dead-keys | Remove the unused `TemplateDocxKey`/`TemplateSchemaKey` (or align to the live key schema) | Zero production refs; build + tests green |
| F3.5 minor-sweep | Close report §7 Minors in scope (dup constructors, hardcoded PT string, 8× `tenantIDFromRequest` dedup, etc.) or record each as a bounded defer with a trigger | Each enumerated Minor is closed or has a written defer trigger |

**Milestone gate:** `backend-api-qa-checklist` (code-quality lens); whole-repo build/test green; explicit
defer list for any un-fixed Minor. `milestone-validator`.

### M4 — Module boundaries / systemic ports  *(LAST — risk-isolation)*
**Objective:** drive the **module-boundaries/DDD** dimension to **≥ A−** and the **H-G class to 0** — no
module issues raw SQL against another module's owned table and no hardcoded domain-state; the IAM role and
MfaCoverage ports are built (retiring the M4 accepted defer). Sequenced last so port work cannot regress an
already-lifted grade; H-G is re-grepped after.

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F4.1 published-constant | `documents` uses `templates/domain.VersionStatusPublished` instead of a hardcoded `"published"` | H-G grep for hardcoded status literals returns 0 |
| F4.2 iam-role-port | An IAM-owned role-membership port serves `security.ListOffHoursAdminActions`; the direct `iam_user_roles` JOIN is removed (read kept **off** any lock-holding tx, H-PRE-1) | No `iam_user_roles` JOIN outside `iam/`; live test proves parity; advisory-lock rule respected |
| F4.3 mfa-coverage-port | An IAM-owned port serves `security.MfaCoverage` (retire the M4 defer) | `security` repo reads no `iam_users`; MfaCoverage metric value parity with a live test |
| F4.4 placeholder-seam | Resolve the `templates/domain.Placeholder` cross-module type — confirm it is a legitimate port-typed dependency or route it through a port | Documented boundary decision; if a leak, the dependency goes via a port |

**Milestone gate:** `backend-api-qa-checklist` (module-boundaries lens); **H-G class-zero** proven by the
report §6 greps after all features; module-boundaries indicatively A−. `milestone-validator`.

## 8. ★ Terminal acceptance — definition of done (written up front)

- **Pass bar (the mission shall be X):** a fresh, independent re-run of the F5.1 10-dimension re-audit at the
  post-M4 HEAD passes the §6 bar — **(1)** module-boundaries, contract-api, and composition all **≥ A−**;
  **(2)** **0** skeptic-confirmed new Critical/Major; **(3)** **H-D = 0**; **(4)** **H-G = 0**.
- **What to validate:**
  - The re-audit report exists, is cited, and grades all 10 dimensions; the 3 formerly-C dimensions each carry an explicit ≥A− call with evidence.
  - Every previously-confirmed finding (the 21 in §5) is either fixed (cited) or carried as an operator-approved bounded defer with a trigger; no new confirmed Critical/Major appears.
  - H-D and H-G are **0**, proven by re-running the exact report §6 grep commands.
  - Whole-repo `go test ./...` green; no prior-milestone regression.
- **How to validate (method + split):** the terminal validation is a **fan-out re-audit** (a subagent cannot
  fan out). So: the **main session** re-runs the F5.1 `Workflow` (10 sonnet dimension auditors +
  adversarial skeptic per Critical/Major + 2 H-D/H-G class-counters + synthesis; refute-by-default), writing
  a new `wiki/backend/_artifacts/architecture-re-audit-<date>.md`. Then the **`mission-validator`** subagent
  **judges that artifact** against this §8: it re-reads the report, independently **re-greps a sample of the
  H-D/H-G "0 remaining" claims** with the report §6 commands, spot-checks 2–3 confirmed-fix sites, and
  re-runs `go test ./...`. Fixture-only or unre-run claims = FAIL.
- **Who validates:** the independent **`mission-validator`** subagent (`.claude/agents/mission-validator.md`).
  It judges and writes `qa/mission-validation.md` only — never edits code or flips status. The main session
  flips status on PASS; the operator gives final Grade-A sign-off.
- **On miss (HS-5):** each missed §6 item (a dimension < A−, a surviving/new confirmed Critical/Major, a
  non-zero class) becomes a bounded remediation micro-milestone run through `milestone`; then the main
  session re-runs the re-audit and re-dispatches `mission-validator`. The operator decides continue vs
  replan at each loop. If any single dimension misses twice, treat as an HS-2 design-boundary signal.

## 9. Hard-stop catalog

| ID | Trigger | Action |
|----|---------|--------|
| HS-1 | Every milestone boundary (M0..M4) | Operator review gate; no next milestone / no merge without approval |
| HS-2 | A fix implies redesign outside the assigned boundary — esp. the IAM role/MfaCoverage ports (F4.2/F4.3) growing into a shared IAM-API redesign, or OTel (F2.3) implying a tracing-architecture overhaul | Stop; report the boundary + minimum prerequisite plan; no symptom-patch |
| HS-3 | A prerequisite boundary fails (build/runnable/auth-session/route/contract truth) | Repair (`runtime-contract-prereq`); rerun the failed checkpoint; resume the feature |
| HS-4 | A `milestone-validator` returns FAIL | Open the named fix feature; re-run its lifecycle; re-dispatch the validator |
| HS-5 | The terminal `mission-validator` (or the re-audit it judges) misses the §8 bar | Bounded remediation micro-milestone; re-run re-audit; re-dispatch; operator decides continue vs replan |
| HS-6 | Scope drift / off-plan discovery mid-milestone (e.g. a fix uncovers a finding F5.1 missed) | Stop; surface the deviation; replan before continuing |

## 10. Constraints respected

- **Skill routing (CLAUDE.md §3):** backend HTTP/contract → `metaldocs-backend-api`; FE codegen/query types →
  `metaldocs-tanstack-query`; DB (only if a port needs a new read) → `metaldocs-database`; prereq repair →
  `runtime-contract-prereq`. Module-wiki sync after structural change → `metaldocs-module-doc-sync`.
- **Contract-first regen order (M1):** build route truth-table → compare runtime/spec/codegen/wiki → regen
  `api.gen.go` then FE codegen. No editing generated wiring or OpenAPI shape from memory.
- **H-PRE-1 advisory-lock hazard (M4):** new port reads stay off any lock-holding atomic tx.
- **Authz root-cause (ADR 0022 / memory):** never symptom-patch authz; F0.1 fixes the shared predicate.
- **No-merge-by-agent**; **never push**; commit-without-asking is authorized (CLAUDE.md §5.0).
- **Drift policy:** update wiki `Last verified:` stamps for any code a wiki doc references; dispatch
  `wiki-curator` after structural changes.

## 11. Execution model

One `mission.md` governs all milestones. Each milestone → its own plan via `milestone` (writing-plans where
installed), executed in a **fresh session**, subagent-driven (implementer + spec-compliance + code-quality
review per feature, fixing by root-cause family). Operator gate between every milestone (HS-1); **no merge by
the agent**. Token discipline: parallel fan-out only where it pays (the terminal re-audit); everything else
direct tools. Model policy: sonnet analysis/review, haiku mechanical, never fable workers, ≤15 concurrent.

## 12. End-state / reconciliation

Fill only when M4 has passed and the terminal gate is green:
- [ ] Every planned feature (M0..M4) has a complete evidence row.
- [ ] Zero unplanned scope merged; anything added is recorded with rationale.
- [ ] Every bounded defer has a written trigger and an owner.
- [ ] Terminal acceptance (§8) passed — link `qa/mission-validation.md` + the new re-audit report.
- [ ] Parent program M5 closed; `grade-a-architecture-remediation/README.md` updated.
- [ ] Operator Grade-A sign-off: <date / name>
