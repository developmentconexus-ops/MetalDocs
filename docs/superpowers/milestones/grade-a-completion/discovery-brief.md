# Discovery Brief: Grade-A Completion (close the F5.1 re-audit gap)

> **Mission slug:** `grade-a-completion`  ·  **Type:** remediation
> **Date:** 2026-06-15  ·  **Branch:** main  ·  **Audited HEAD:** `02ed1c24` (code unchanged at current `5ce0cffb` — docs-only commits since; findings valid)
> **Agents / models used:** evidence base is the **F5.1 re-audit** (`Workflow wf_b0109977-23a`: 42 agents — 10× sonnet dimension auditors + sonnet adversarial skeptic per Critical/Major + 2× sonnet H-D/H-G class-counters + sonnet synthesis; ~2.1M tokens). This brief synthesizes that report; **no new fan-out** was spent.
> This is the **evidence base** the mission stands on. Every claim in `mission.md` traces to a finding here.

## Method

Discovery for this mission was already performed by the program's own authoritative gate: the F5.1
independent re-audit (`wiki/backend/_artifacts/architecture-re-audit-2026-06-15.md`). It read the **actual
code** at HEAD `02ed1c24`, graded all 10 dimensions, and sent **every** Critical/Major finding to an
independent skeptic (refute-by-default). Only skeptic-**confirmed** findings are carried here. Re-running a
fresh fan-out would re-derive the same cited inventory at ~2M tokens — deliberately skipped.

| Agent / lens | Scope swept | Verified how |
|--------------|-------------|--------------|
| 10 dimension auditors (F5.1) | whole backend at HEAD `02ed1c24`, per dimension | read actual code, cited `file:line` |
| Adversarial skeptic per Crit/Major (F5.1) | every Critical/Major finding | re-read cited code, confirm/downgrade/refute |
| H-D / H-G class-counters (F5.1) | delivery handlers vs OpenAPI/codegen; cross-module SQL | reproducible grep commands (report §6) |
| This brief (synthesis) | the confirmed-findings + class counts | cross-checked against report §4/§6; `git diff 02ed1c24..HEAD -- '*.go'` empty (code unchanged) |

**Skeptic-pass outcome:** 8 auditor findings were **refuted** or **downgraded to Minor** by the F5.1 skeptic
(report §5) — e.g. the WriteTimeout-vs-WebSocket claim (refuted: post-Hijack net/http stops enforcing
WriteTimeout), the cross-tenant session-revoke claim (downgraded: both callers require it). **Those must not
be re-raised** in this mission without new evidence. Survivors only are inventoried below.

**Count note:** report §3 headline says "18" confirmed Crit/Major; report §4 detail table totals **21**
(1 Critical + 20 Major). The §4 detail is authoritative — this brief uses 21.

## Findings

Confirmed Critical/Major from report §4, plus the H-D (4) and H-G (1) class sites from report §6, grouped by
**root-cause family** (so the mission fixes the class, not the instance). Severity = skeptic's final
severity. All citations are `file:line` at HEAD `02ed1c24`.

### Family A — Contract / API integrity (dimension contract-api C+ → must reach A−; closes H-D=4)

| # | Finding (citation) | Sev | Confidence | Proposed home |
|---|--------------------|-----|------------|---------------|
| A1 | Checkpoint endpoints serialize untagged `domain.Checkpoint` → PascalCase on wire, breaks FE snake_case contract — `internal/modules/documents/delivery/http/handler.go:881` | **Critical** | verified | M-contract |
| A2 | `renameDocument` writes raw `*domain.Document` (incl. base64 FormDataJSON); spec declares 200 no-content — `…/documents/delivery/http/handler.go:519` | Major | verified | M-contract |
| A3 | `createTemplate` emits undeclared top-level `id`/`version_id` (H-D DRIFT-1) — `…/templates/delivery/http/routes_generated.go:64` | Major | verified | M-contract |
| A4 | `createNextVersion` emits 201, spec declares 200 (H-D DRIFT-3) — `…/templates/delivery/http/routes_create.go:36` | Major | verified | M-contract |
| A5 | `presignTemplateAutosave` emits 201, spec declares 200 (H-D DRIFT-2) — `…/templates/delivery/http/routes_autosave.go:42` | Major | verified | M-contract |
| A6 | Pervasive `map[string]any` bypasses generated types at 6 sites (presignAutosave, commitAutosave, listRevisionHistory, restoreCheckpoint, duplicateDocument, sessions list) — `…/documents/delivery/http/handler.go:816` + | Major | verified | M-contract |
| A7 | `routes_profiles` emits raw `domain.DocumentProfile`, missing required spec fields + exposing extras (H-D DRIFT-4) — `…/taxonomy/delivery/http/routes_profiles.go:67,111,126,169` | Major (H-D) | verified | M-contract |
| A8 | (minor) `documentStats` raw `application.DocumentStats`; audit 405 missing `Allow` header — `…/documents/delivery/http/handler.go:317`, `…/audit/delivery/http/handler.go:81` | Minor | verified | M-contract |

### Family B — Auth / authz / session correctness (correctness + security; raises authz/sessions dims)

| # | Finding (citation) | Sev | Confidence | Proposed home |
|---|--------------------|-----|------------|---------------|
| B1 | `authz.Require` checks only `effective_to IS NULL`, ignores `effective_from <= now()` → **future-dated memberships grant access early** — `internal/modules/iam/authz/authz.go:123` | Major (security) | verified | M-authz |
| B2 | Manual-code CD-create never seeds tx identity → `MustActorID` returns `ErrActorContextMissing` before system-admin bypass → **all non-admin manual-code creates fail** — `internal/modules/controlleddocuments/application/service.go:173` | Major (functional) | verified | M-authz |
| B3 | `CapDocumentView` tenant-grade silently narrowed to area-grade when areaCode != "" — `internal/modules/documents/approval/application/read_service.go:68` | Major | verified | M-authz |
| B4 | `ChangePassword` revokes sessions but sends no expired-cookie header → client holds dead cookie — `internal/modules/auth/delivery/http/handler.go:153` | Major | verified | M-authz |

### Family C — Module boundaries / DDD (dimension C → B+ → must reach A−; closes H-G=1 + new boundary site)

| # | Finding (citation) | Sev | Confidence | Proposed home |
|---|--------------------|-----|------------|---------------|
| C1 | `overrideStatus := "published"` hardcodes templates-owned domain state (H-G VIOLATION-1) — `internal/modules/documents/application/service.go:282` | Major (H-G) | verified | M-ports |
| C2 | `security.ListOffHoursAdminActions` JOINs IAM-owned `iam_user_roles` directly; **no IAM role port** (NEW — not in M4 census) — `internal/modules/security/infrastructure/postgres/repository.go:345` | Major | verified | M-ports |
| C3 | `security.MfaCoverage` reads IAM-owned `iam_users` directly (the M4 accepted bounded defer) — `internal/modules/security/infrastructure/postgres/repository.go:67` | Major | verified | M-ports (decide: build port vs keep defer — operator) |
| C4 | documents imports `templates/domain.Placeholder` as a production type across the module seam (repository/service/fillin/delivery) — `internal/modules/documents/repository/repository.go:16` | Major | verified | M-ports (assess: legitimate port-typed dependency vs leak) |

### Family D — Composition / observability (dimension C → B+ → must reach A−)

| # | Finding (citation) | Sev | Confidence | Proposed home |
|---|--------------------|-----|------------|---------------|
| D1 | Scheduler hardcodes `slog.NewTextHandler` instead of `slog.Default()` (JSON) → scheduler log lines unparseable by aggregators — `internal/modules/jobs/scheduler/scheduler.go:131` | Major | verified | M-observability |
| D2 | Scheduler per-job metrics (`MetricsSnapshot()`) never wired to any scrape endpoint — `internal/modules/jobs/scheduler/scheduler.go:273` | Major | verified | M-observability |
| D3 | OTel is HTTP-envelope only; zero app-level spans / DB instrumentation — `internal/platform/observability/otel.go:95` | Major | verified | M-observability |
| D4 | (minor) DB pool stats absent from `/api/v1/metrics`; "Prometheus" comment serves custom JSON — `internal/platform/bootstrap/api.go:99`, `apps/api/cmd/metaldocs-api/permissions.go:95` | Minor | verified | M-observability |

### Family E — Code-quality & dead/orphan tail (raises code-quality + legacy dims; one is functional)

| # | Finding (citation) | Sev | Confidence | Proposed home |
|---|--------------------|-----|------------|---------------|
| E1 | `IAMUserOptions` dependency never wired → placeholder user-lookup silently returns empty (functional) — `apps/api/cmd/metaldocs-api/main.go:413` | Major (functional) | verified | M-quality |
| E2 | `NewFreezeService` accepts `fanoutClient any`, drops type-mismatch silently — `internal/modules/documents/application/freeze_service.go:77` | Major | verified | M-quality |
| E3 | `ListDocumentComments` carries dead `userID` param through interface + impl (may mask missing authz scope) — `internal/modules/documents/application/service.go:433` | Major | verified | M-quality |
| E4 | `_ = snap` discards a 4-column (incl. JSON blob) snapshot read in `Pin` — `internal/modules/documents/application/freeze_service.go:194` | Major | verified | M-quality |
| E5 | `TemplateDocxKey`/`TemplateSchemaKey` exported, zero prod callers, format diverges from live key schema — `internal/platform/objectstore/template_keys.go:5` | Major (dead) | verified | M-quality |
| E6 | (minors) duplicate `New`/`NewService` constructors; hardcoded Portuguese business string; 8× private `tenantIDFromRequest` copies; SHA-1 dedup key; etc. (report §7) | Minor | verified | M-quality |

## Constraints & risks surfaced

- **Contract-first regen order (house rule).** Family A changes touch handler output shapes / status codes
  → must follow build-route-truth-table → compare runtime/spec/codegen/wiki → regen `api.gen.go` + FE
  codegen. Skill: `metaldocs-backend-api` (+ `metaldocs-tanstack-query` if FE query types shift). A1/A3 change
  the FE-visible contract — FE codegen regen is in-scope.
- **H-PRE-1 advisory-lock hazard.** Any new port read (Family C) must stay **off** a lock-holding atomic tx
  (never call an authz-recording read inside the audit-lock tx). Applies to a new IAM role port for C2.
- **HS-2 redesign-boundary candidates (flag, don't patch through):**
  - **C3 `MfaCoverage`** — was an *accepted* M4 defer (aggregate JOIN, no display-name). Building a port for
    an aggregate metric query may be a shared-IAM-API design question, not a mechanical fix. Operator decides
    scope (build port vs keep documented defer) in Phase 2.
  - **C2 IAM role port** — security needs IAM role-membership data; no port exists. Designing it is an
    IAM-module API addition (additive, but a cross-module contract) — bounded, but name it.
  - **D3 OTel app-level spans** — full distributed tracing is potentially large; "A−" may not require full
    span coverage. Phase 2 must fix the bar (what observability level = A−).
- **B1 is a latent security defect** (premature access via future-dated membership). Highest fix priority;
  needs a regression test proving `effective_from > now()` denies.
- **Risk-isolation (grade-a precedent):** systemic ports (Family C) regressed the grade once already (M4 →
  re-audit residue). Sequence port work **last** so it cannot regress an already-lifted grade, and re-grep
  H-G after.

## Open questions for the operator (Phase 2 must lock)

1. **Scope of the 3 HS-2 candidates** — C3 (`MfaCoverage` port vs keep defer), C2 (build IAM role port), D3
   (how much OTel = A−). Full-close vs bounded-defer for each.
2. **Milestone sequencing** — recommend authz-correctness first (B, security), contract integrity second (A,
   FE-visible), observability + quality tail (D, E) in parallel-eligible middle, module ports **last** (C).
   Operator confirms or reorders.
3. **A− bar definition per dimension** — is "B+ → A−" satisfied by closing the listed findings, or does the
   operator want an explicit rubric the re-audit must score against?
4. **Minor findings (E6, A8, D4, report §7)** — fix-in-scope vs explicit bounded defer with triggers.
5. **Terminal proof** — confirm terminal acceptance = re-run the **same** F5.1 re-audit method and pass §6
   (vs a lighter targeted re-check).

## Coverage statement

- **Not re-swept:** this brief did **not** run a new audit — it inherits F5.1's coverage (all 10 dimensions
  at HEAD `02ed1c24`). Anything F5.1 missed, this brief also misses. F5.1's own coverage caveat: the M1
  full-HTTP E2E was a SKIP (not re-stood-up), and Minor findings were not skeptic-gated.
- **Excluded by design:** the 8 refuted/downgraded findings (report §5) are out-of-scope unless new evidence
  appears.
- **Validity window:** findings cited at `02ed1c24`; `git diff 02ed1c24..5ce0cffb -- '*.go'` is empty, so
  every `file:line` is current. If code changes before execution, re-anchor the cited lines.
