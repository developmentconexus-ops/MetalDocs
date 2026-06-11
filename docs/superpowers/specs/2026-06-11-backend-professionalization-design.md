# Backend Professionalization — Design Spec

> **Date:** 2026-06-11
> **Status:** APPROVED (brainstorm complete, user sign-off in session)
> **Owners:** Leandro (product/ops decisions, credential rotation) · Claude (execution)
> **Scope:** Governs execution of the Stage-2 roadmap (`wiki/backend/stage2-evaluation.md`) — Waves 0–3. Every wave executes against THIS document. Deviations from a decision below require coming back to the user, not improvising.
> **Out of scope:** Frontend, new features, anything not traceable to a finding ID in `wiki/backend/legacy-register.md`.

---

## 1. Context and goal

Stage 1 mapped the backend completely (`wiki/backend/` atlas, 19 areas, 212 evidence-anchored flags). Stage 2 evaluated every finding against cited industry standards and produced 34 verdicts and a 4-wave roadmap (`wiki/backend/stage2-evaluation.md`). This spec records the **brainstorm decisions** that resolve the open judgment calls, and defines the **execution protocol** so the waves can run in fresh sessions without context loss and without creating new errors.

**Goal (program-level, unchanged):** structure the MetalDocs backend like a professional SaaS — industry standards, rules and best practices — so it can scale, evolve, and be used in production at the user's company. Simplicity-first: the smallest change that reaches the professional bar; over-engineering is a defect.

## 2. Decisions (brainstorm output — binding)

| ID | Decision | Rationale / consequences |
|----|----------|--------------------------|
| D-1 | **Deployment target: single server, Docker Compose.** | Compose is THE deployment manifest → F-19 jobs service is P0. Single replica → in-memory auth-failure limiter is mild (restart clears lockouts; no cross-replica issue) but the Postgres-backed fix stays (cheap, durable). WebSocket drain (F-16B) deferred. OTel full pipeline would be over-engineering today → Wave 3, trigger: "second host or first external-tenant SLA". |
| D-2 | **Compliance bar: audit-facing, no ceremony.** | MetalDocs is the company's real document-control system → build as if an ISO 9001 auditor samples the trail. F-07 fixed with in-transaction audit/governance writes everywhere + CI guard forbidding post-commit audit calls. NO regulated-industry paperwork (validation protocols, 21 CFR signature docs) until certification is in scope. |
| D-3 | **Tenancy: keep multi-tenant shared-schema design; one real tenant foreseeable.** | Shared-schema + tenant_id is the standard SaaS pattern (not a "bad monolith"). RLS = defense-in-depth against our own bugs: `controlled_documents` + `audit_events` get RLS in Wave 2; full per-table isolation program trigger-gated on "first external tenant onboarded". ADR 0027 records this sequencing. |
| D-4 | **F-18 secret remedy: FULL.** | The dev DB password (redacted — see `.env`) is committed via 3 channels: `cmd/seed-test-document/main.go:25`, 2 wiki reference docs quoting it verbatim, and 5 Stage-1/2 audit docs (audit propagation mistake — see D-4a). Remedy: (1) USER rotates the password when Docker stack is re-created (new random value in `.env`; also rotate anywhere personally reused); (2) delete dead binary + artifacts, REDACT the value from all 7 docs; (3) `git filter-repo --replace-text` across ALL history + force-push to origin (single-user repo: re-clone cost is minutes); (4) gitleaks secret scanning in CI. History rewrite is gated on rotation confirmation. |
| D-4a | **Documentation rule (new, permanent):** secrets are referenced by location, NEVER quoted, in any doc, report, or commit message. | The Stage-1/2 audit itself reproduced the secret in 5 committed files. Rule lands in `wiki/standards/documentation-governance.md`. |
| D-4b | **Residual history accepted (2026-06-11, post-Wave-0):** the old secret remains reachable from stale local refs (`refs/heads/main` pre-rewrite, tag `w5-complete`). NO further scrubbing — the password is rotated (true remedy done) and the user will re-baseline into a **fresh repository at first release**, erasing all history. | Do not resume filter-repo work on residual refs; it is superseded. The fresh-repo re-baseline at release is the closing act of F-18. |
| D-5 | **Execution style: wave-by-wave, review gate between waves, each wave in a FRESH session.** | Each wave session reads: `CLAUDE.md` → `wiki/README.md` → `wiki/references/current-agent-handoff.md` → THIS SPEC → its wave card (§5–§8). Wave completes only with the evidence block (§4) recorded. User reviews between waves. |

## 3. Reference stack (read order for every wave session)

1. `CLAUDE.md` — gates, skill routing, hard-stop rule
2. `wiki/architecture/backend-target-architecture.md` — REQ-*/RF-* normative spec (commits cite IDs)
3. `wiki/backend/stage2-evaluation.md` — verdicts, master table, smallest-correct-fix per item
4. `wiki/backend/legacy-register.md` — evidence anchors per finding
5. This spec — decisions + wave cards + protocol

## 4. Execution protocol (applies to every wave)

- **Branch/commits:** stay on the integration branch (currently `qa/iam-area-membership`). One commit per roadmap item (or tightly-related pair), message cites finding ID + REQ IDs. Exception: Wave 0 history rewrite (rewrites all history; coordinated step).
- **Per-item loop:** read the finding's anchors → implement smallest-correct-fix → `go build ./...` + `go vet` + targeted tests for the touched slice → commit.
- **Per-wave close:** full test run for touched modules + `api-lint` (when contract touched) → `/code-review` of the wave diff → findings fixed by family → wiki sync (`metaldocs-module-doc-sync` for touched documented modules; stamps bumped) → evidence block appended to `wiki/references/current-agent-handoff.md`: commands run, outcomes, review disposition, bounded defers.
- **Living tracker (mandatory):** [`wiki/backend/roadmap.md`](../../../wiki/backend/roadmap.md) is updated IN THE SAME COMMIT as each item/wave close (statuses, commit hashes, evidence). A wave with stale tracker rows is not closed. (Located in the wiki because `docs/superpowers/plans/` is gitignored as ephemeral; this tracker is durable program record.)
- **Wave F (final full review)** closes the program: all CI guards green, full static + runtime QA, legacy-register sweep (every entry resolved-or-deferred-with-trigger), blueprint re-score, full-program code review, wiki coherence pass, closing report. Defined as the Wave F card in the roadmap.
- **Hard-stop rule (CLAUDE.md) stands:** if a fix turns out to imply cross-module redesign beyond its card, STOP and report — do not symptom-patch.
- **Docker caveat:** until the user's Docker reinstall is done, runtime verification is impossible. Items marked ⚠ runtime-gated get code+test verification now and a `[runtime-unverified]` note in the evidence block, re-checked when the stack is up. Do not block the wave on it.
- **Models:** wave sessions use balanced sub-agent models when fanning out (sonnet implement, haiku mechanical checks); no top-tier inheritance (see memory: workflow-model-balancing).

## 5. Wave 0 card — P0 prerequisites

| # | Item | Findings | Steps | Verify |
|---|------|----------|-------|--------|
| 0.1 | Secret removal + redaction | F-18, D-4a | Delete `cmd/seed-test-document/`; redact secret value from `wiki/references/local-dev-startup.md:22`, `wiki/references/local-dev-credentials.md:32`, `wiki/backend/legacy-register.md:527`, `wiki/backend/stage2-evaluation.md:28`, `wiki/backend/_artifacts/stage2/security-secrets.md`, `wiki/backend/_artifacts/stage1/synthesis-legacy.md` (replace with `<redacted — see .env>`); `git rm bin/metaldocs-api.exe`; add `bin/*.exe` to `.gitignore`; delete `scripts/start-spec1-api.ps1`; SHA-256-pin or rebuild `scripts/api-lint/api-lint.exe` | `go build ./...`; grep proves zero occurrences of the secret in working tree |
| 0.2 | Secret-scan guard | F-18 | Add gitleaks step to CI workflow + document the D-4a rule in `wiki/standards/documentation-governance.md` | CI config valid; gitleaks run on working tree is clean |
| 0.3 | USER: rotate password | F-18 | New random `POSTGRES_PASSWORD`/`PGPASSWORD` in `.env` at Docker re-creation; rotate anywhere reused personally | User confirms in chat |
| 0.4 | History rewrite (AFTER 0.3 confirmed) | F-18 | `git filter-repo --replace-text` (old secret → `***REDACTED***`) on a fresh mirror clone; force-push; user re-clones other machines | `git log --all -S '<secret>'` returns empty |
| 0.5 | jobs deployment ⚠ | F-19-deployment | `deploy/docker/jobs.Dockerfile` (clone worker.Dockerfile, target `metaldocs-jobs`); `jobs` service in `deploy/compose/docker-compose.yml` with `depends_on: postgres(healthy), api(healthy)` | `docker compose config` parses; runtime-gated |
| 0.6 | platform/observability layering fix | F-06a | Replace `authdomain.CurrentUserFromContext` import in `internal/platform/observability/http.go:15` with injected `func(*http.Request) string` callback (pattern: `platform/ratelimit.Middleware`); wire at `main.go` construction | `go build` + `go test ./internal/platform/observability/...`; grep: no `modules/` import under `platform/` |

## 6. Wave 1 card — high-value / low-blast

Items 1.1–1.9 exactly as the Wave 1 table in `wiki/backend/stage2-evaluation.md` (middleware chain reorder + panic recovery + pre-auth login rate-limit [F-01, depends 0.6]; server timeouts [F-16]; delete `spec2.yaml` + `internal/api/v2` + fix capability-catalog CI gate [F-03]; idempotency raw codes [F-09]; dead search fields + camelCase param + bare-405 fix [F-13a/b, D-03]; River dual-migration [F-19, depends 0.5]; lease_reaper JOIN bug [F-19]; templates audit sink split [F-07-sub-split]; scaffold .gitkeep cleanup [F-08]).

Wave-1 addition per brainstorm: **import-boundary CI guard** — a lint/test asserting `internal/platform/**` never imports `internal/modules/**` (locks in 0.6 and prevents F-06a recurrence). Middleware reorder target chain is normative in `backend-target-architecture.md` §2.1; add the REQ-MW-7 chain-order test in the same commit.

## 7. Wave 2 card — structural refactors

Items exactly as the Wave 2 table in `wiki/backend/stage2-evaluation.md`, with brainstorm refinements:
- **F-07/D-01 atomicity:** `RecordTx`/`LogTx` interfaces; approval module is the template; **plus CI guard forbidding post-commit audit/governance calls** (D-2).
- **F-12 RLS:** `controlled_documents` + `audit_events` only; **write ADR 0027 first** (auth_identities tenant-global by design; RLS sequencing; remaining tables trigger-gated per D-3); `ExportJob.TenantID` → `uuid.UUID` type fix.
- **F-11 capabilities:** resolve `"template.admin"` (decide capability with user if a new constant is minted → ADR 0028, else no ADR); typed EventType constants; tier-1/tier-2 pairing audit.
- F-06b/c/d boundary extractions; F-05/D-04 rate-limiter swap (delete `platform/security/ratelimit.go`, activate `platform/ratelimit`, coordinates with F-01); F-10 N+1 fixes (authz-path first); F-20e Postgres-backed auth-failure counter; F-14 dead-code deletion PRs; F-07-sub deprecated logger removal (after atomicity fix).

## 8. Wave 3 — trigger-gated (do NOT execute proactively)

The Wave 3 table in `wiki/backend/stage2-evaluation.md` stands, with D-1 refinement: **F-17 OTel** trigger is "second host or first external-tenant SLA"; until then the bar is structured logs + propagated request IDs (covered by Wave 1 F-01). Each Wave 3 item executes only when its written trigger fires, typically piggybacked on the next touch of its area.

## 9. Error-containment guards (the "don't create more errors" answer)

1. Fresh session per wave + this spec = no context drift.
2. One commit per item = every change revertable in isolation.
3. CI guards added as waves land (gitleaks, import-boundary, post-commit-audit, chain-order test) = each fixed class of error becomes machine-blocked, not memory-dependent.
4. Wiki sync per wave (drift policy) = documentation never lags more than one wave.
5. Review gate between waves = user sees evidence before scope continues.
6. Hard-stop rule = redesign-grade surprises surface instead of being patched around.

## 10. Completion definition

The program is DONE when: Waves 0–2 executed with evidence blocks; Wave 3 triggers documented; ADR 0027 (and 0028 if applicable) merged; all four CI guards active; `wiki/backend/legacy-register.md` entries marked with resolution notes; `backend-blueprint.md` maturity grades re-scored; handoff updated. Then the backend is at the professional bar this program defined — further evolution goes through the normal REQ/ADR process, not through this spec.
