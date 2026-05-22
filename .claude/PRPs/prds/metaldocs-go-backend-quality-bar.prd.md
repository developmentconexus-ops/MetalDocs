# PRD — MetalDocs Go Backend Production-Grade Quality Bar

**Slug:** `metaldocs-go-backend-quality-bar`
**Date:** 2026-05-22
**Author:** Leandro (solo dev) + Claude
**Status:** Draft v1 — pending review
**Branch target:** fresh branch (not extending `codex/backend-invariant-slice`)

---

## 1. Problem

MetalDocs Go backend is in late development for an ISO 9001 / ISO 13485 QMS product. Current state across `apps/api/` + `internal/`:

- Mixed maturity: some modules pro-grade, others legacy wrappers, no consistent standard.
- Module #1 review surfaced 4 Critical + 5 High + 6 Medium + 6 Low findings.
- Module #2a (security boundary, 7 packages) surfaced 5 Critical + 11 High + 11 Medium + 8 Low findings.
- Recurring failure classes: stringly-typed boundaries (`string` for tenant/user/role/error-code), silent error swallowing, missing trusted-proxy semantics, single-phase idempotency with replay races, fail-open auth paths.
- No published Go bar. Each fix decided ad-hoc. Pattern wins do not survive into the next module.
- Without a bar, agent-driven reviews drift: 13 more modules + 18 platform packages still to audit at ~$15-30 each. Without a written standard, fixes regress.

**Cost of inaction:** ISO audit failure risk (no traceability of code-quality controls), production bugs in compliance-critical paths (audit log, RBAC, e-signature when built), unmaintainable codebase as it grows past current ~20k LoC.

## 2. Goal

Publish a written, enforced, MetalDocs-specific Go backend production-grade quality bar that:

1. Codifies patterns already proven in module #1 + #2a fixes (typed roles via `iamdomain.Role`, two-phase idempotency, trusted-proxy CIDR helper, fail-closed authn, RFC 9457 problem envelope).
2. Extends `~/.claude/rules/ecc/golang/{coding-style,hooks,patterns,security,testing}.md` with MetalDocs-specific patterns and anti-patterns — does **not** reinvent them.
3. Red-gates regressions via a curated `.golangci.yml` enforced in CI.
4. Gives every future module review a written rubric so findings are bar-violations, not novel inventions.
5. Gives the refactor work for #2a Highs + module #2b/#2c + modules #3-#10 a written playbook.

## 3. Hypothesis

> We believe a **published `wiki/standards/golang/` bar + enforced `.golangci.yml` red-gating CI + per-module refactor playbook** will let solo-dev ship MetalDocs to ISO 9001/13485 production grade
> **because** the bar captures patterns already proven in module #1 + #2a fixes (typed IDs, two-phase idempotency, trusted-proxy, fail-closed authn, structured logging discipline) and lint catches their regression class.
> **We'll know we're right when** the next 3 module reviews each cost <$15, produce zero new Critical findings of the same class as #1/#2a, and CI red-gates any PR that reintroduces the banned patterns.

## 4. Users

| Persona | Role | Interaction |
|---------|------|-------------|
| Leandro (solo dev) | Primary author + reviewer | Writes code; reads bar before non-trivial change; runs `.golangci.yml` locally |
| Claude / ECC agents | Review + implementation | Cites bar in findings; agents extend `~/.claude/rules/ecc/golang/*` automatically |
| ISO auditor (future) | External | Reviews `wiki/standards/golang/` + lint config as evidence of code-quality controls |

## 5. Scope

### In scope
- All Go code under `apps/api/` and `internal/` (hand-written).
- All platform packages, all `internal/modules/*`, `cmd/metaldocs-api`.
- Curated `.golangci.yml` covering: errcheck, govet, staticcheck, gosec, gocritic, revive, errorlint, exhaustive, sqlclosecheck, rowserrcheck, contextcheck, nilerr, bodyclose, gocyclo, gocognit (curated thresholds).
- Linkage from each bar section to existing `~/.claude/rules/ecc/golang/*` and to the evidence finding(s) that justified it.
- Per-module refactor playbook (`wiki/standards/golang/refactor-playbook.md`) describing the exact sequence to bring a legacy module up to bar.

### Out of scope (v1)
- Frontend (any TypeScript / React / `frontend/apps/web/**`).
- Database migration policy (lives in `wiki/database/`).
- Wiki workflow / `metaldocs-module-doc` skill itself.
- OpenTelemetry / Prometheus instrumentation (not wired in repo — defer until the metrics platform exists).
- OpenAPI contract refactors (separate concern).
- E-signature module (not built yet — bar will mark it as a deferred section).
- Continuing #2b–#10 reviews until v1 of the bar is published (cursor stays at #2b).
- Pinning Go version upgrades (locked at the project's `go 1.25.0`).

## 6. Constraints / Reality

- **Solo dev, cost-conscious.** No team to socialize a bar with — bar is for me + Claude agents.
- **Late-stage development.** Refactor budget is real but not infinite. Bar must distinguish *new code (mandatory)* from *legacy code (incremental migration)*.
- **ISO 9001 / 13485.** Audit log, RBAC, idempotency, document state transitions, e-signature (when built) sit on the regulated path → defense-in-depth required.
- **Existing rules.** `~/.claude/rules/ecc/golang/{coding-style,hooks,patterns,security,testing}.md` already define language-agnostic golang baseline. Bar **extends**, does not duplicate.
- **Evidence base.** Bar must cite the 4+5+11=20 Critical+High findings from module #1 + #2a as the failure modes it prevents. No speculative rules.
- **Go 1.25.0.** Range-over-func, structured logging (`slog`) is in std, generics mature. Bar can assume these.
- **Branch hygiene.** Bar lands on a fresh branch off `main`, not on top of unmerged review/fix branches.

## 7. Success metrics (MVP)

| Metric | Target | Source |
|--------|--------|--------|
| Bar published | `wiki/standards/golang/` v1 exists with index + ≥6 topic docs + refactor playbook | Repo |
| Lint red-gating active | `.golangci.yml` in repo root; `golangci-lint run` exit 0 on `main` after baseline; CI step blocks on lint failure | Repo + CI |
| Module review cost | Next 3 module reviews (#2b, #2c, #3) each ≤ $15 | Cost report |
| Recurrence rate | Zero new Critical findings of same class as #1/#2a Criticals in #2b+#2c+#3 | Findings docs |
| Rule citation in findings | 100% of new Critical/High findings in next 3 reviews cite a bar section by anchor | Findings docs |
| Solo-dev signal | Leandro reports the bar shortened a non-trivial decision (e.g. "I knew which pattern to pick from bar") at least once per next 3 modules | Self-report |

## 8. Non-goals

- Pleasing a hypothetical future team. Optimize for solo dev + Claude agents.
- 100% greenfield rewrite. Incremental, evidence-driven.
- Replacing `~/.claude/rules/ecc/golang/*`. Bar references them.
- Producing exhaustive Go textbook material. Only document patterns where MetalDocs has been bitten or is at risk.

## 9. Must-have v1 deliverables

1. **`wiki/standards/golang/README.md`** — index with anchor table mapping each bar section to (a) the proven pattern, (b) the failure mode it prevents (cite finding ID + commit SHA where landed), (c) the lint rule that enforces it, (d) the `~/.claude/rules/ecc/golang/*` rule it extends.
2. **`wiki/standards/golang/typed-boundaries.md`** — string→typed migration: `iamdomain.Role`, `TenantID`, `UserID`, `ErrCode`, `IdempotencyKey`. Cites #2a M-series typing findings.
3. **`wiki/standards/golang/errors-and-logging.md`** — error wrapping, `errors.Is/As`, never-swallow rule, `slog` structured logging conventions, no `fmt.Errorf` without `%w` for wrap, no log-and-return without classification. Cites #2a H-series silent-failure findings + #2a Critical idempotency error-swallow.
4. **`wiki/standards/golang/security-boundaries.md`** — fail-closed authn (UserIDFromContext returns `(string, bool)`), trusted-proxy CIDR pattern, RFC 9457 problem envelope, header trust rules (`X-Forwarded-*` only behind trusted hop), CORS reject-disallowed. Cites #2a C1/C2/C5 + H7 + landed commits `def24e4a`, `2f8f6dcc`, `d2242313`.
5. **`wiki/standards/golang/idempotency-and-concurrency.md`** — two-phase write (BeginReplay/CompleteReplay/FailReplay), `ON CONFLICT DO NOTHING RETURNING`, no single-phase replay storage, retry-safe handler semantics. Cites #2a C3/C4 + idempotency schema v2 in commit `07312d58`.
6. **`wiki/standards/golang/persistence.md`** — pgx usage, parameterized queries, no string concat, `RowsAffected` checks, transaction boundaries, `Close`/`defer` discipline.
7. **`wiki/standards/golang/http-handlers.md`** — context discipline, request validation at boundary, problem-envelope responses, no `panic` in handlers, recovery middleware contract, rate-limit + idempotency middleware ordering.
8. **`wiki/standards/golang/testing.md`** — table-driven tests, `t.Helper`, `testdata/` fixtures, no shared mutable state, integration tests hit real Postgres (no DB mocks — extending `~/.claude/rules/ecc/golang/testing.md` with MetalDocs's "no mock DB" rule).
9. **`wiki/standards/golang/package-layout.md`** — `internal/modules/<name>/{api,service,store,domain}` template, `apps/api/cmd/*` for entrypoints, no import cycles, dependency direction (api → service → store → domain).
10. **`wiki/standards/golang/refactor-playbook.md`** — step-by-step legacy-module → bar-compliant sequence: review with parallel ECC agents → categorize findings → land Criticals → land Highs → run lint baseline → commit → update tracker.
11. **`.golangci.yml`** in repo root — curated linters with thresholds calibrated to the MetalDocs codebase (no green-from-day-one fantasy: produce baseline issue list, gate on regressions).
12. **CI wiring** — workflow step running `golangci-lint run` that blocks merge on failure. Minimal step, no new infra.

## 10. Stretch / v2

- Pre-commit hook running `golangci-lint run --new-from-rev=origin/main` for fast local feedback.
- Generated coverage report gating per-module.
- A `make audit` target that runs the full bar verification + spawns ECC review pipeline.
- E-signature section once that module exists.
- Auto-applied `gofmt -s` + `goimports` in PostToolUse hooks (extends `~/.claude/rules/ecc/golang/hooks.md`).

## 11. Risks

| Risk | Mitigation |
|------|-----------|
| Bar becomes shelfware — written but not consulted | Wire bar citations into ECC agent prompts; require `wiki/standards/golang/*` reference in any new finding |
| Lint config too strict → baseline noise blocks all PRs | Use `--new-from-rev` for incremental gating; allow baseline waiver file with TODO + finding link |
| Bar drifts from `~/.claude/rules/ecc/golang/*` | Each bar doc carries `Extends:` front-matter pointing at the upstream rule file; doc-updater checks alignment |
| Solo-dev fatigue from documentation overhead | v1 must be drafted in this PRP pipeline only; future modules update bar opportunistically, not as standalone work |
| ISO auditor disagrees with bar choices | Bar is internal code-quality control evidence, not a regulatory submission — auditor scope is QMS process, not Go style |
| Refactor playbook stalls because module sizes vary wildly | Playbook scales by LoC bucket; #2a-style split-and-conquer documented as the canonical move |

## 12. Open questions

- **Q1:** Use `gofmt`/`goimports` only, or adopt `gofumpt` (stricter)? *Lean: `gofumpt` for new code, `gofmt` for legacy until touched.*
- **Q2:** Adopt `golangci-lint` v2 config format directly or v1 for max compatibility? *Lean: v2; project on Go 1.25 already.*
- **Q3:** Do we treat `gocyclo`/`gocognit` as hard gates or warnings? *Lean: warnings v1, hard gate v2.*
- **Q4:** Where does CI run? `.github/workflows/` already in repo, or new workflow file? *To verify in Phase 8 of PRP-PLAN.*
- **Q5:** Should bar address generated code (`api.gen.go`)? *Lean: exclude from lint via `.golangci.yml` `skip-files`, document the exclusion explicitly.*

## 13. Dependencies

- `~/.claude/rules/ecc/golang/{coding-style,hooks,patterns,security,testing}.md` — bar extends these; if they change, bar may need re-alignment.
- `wiki/reviews/2026-05-21-go-backend-review.md` — evidence tracker; bar cites it.
- `wiki/reviews/2026-05-21-go-backend-review/cmd-metaldocs-api.md` + `platform-2a-security.md` — failure evidence cited by bar.
- `go.mod` (Go 1.25.0) — bar assumes this version.
- `golangci-lint` binary in dev + CI environment.
- ECC golang skills (`ecc:go-review`, `ecc:go-build`, `ecc:go-test`) — agents must learn to cite bar.

## 14. Evidence base (load-bearing findings)

From module #1 (`apps/api/cmd/metaldocs-api`):
- 4 Critical (commits `6eb31ec7`, `66fe1ee3`) — wiring + lifecycle issues feeding `package-layout.md` + `http-handlers.md`.

From module #2a (`platform/{authn,security,idempotency,ratelimit,tenant,problem,httpresponse}`):
- **C1** UserIDFromContext silent zero-value — landed `def24e4a` → `security-boundaries.md`
- **C2** trusted-proxy + header trust — landed `2f8f6dcc` → `security-boundaries.md`
- **C3** idempotency two-phase write — landed in branch tip (idempotency package) → `idempotency-and-concurrency.md`
- **C4** idempotency replay-race fix — landed with C3 → `idempotency-and-concurrency.md`
- **C5** fail-closed authn callsite audit — landed `d2242313` → `security-boundaries.md`
- **H1/H4/H7** CORS + rate-limit + cookie hardening — landed during fix sprint → `http-handlers.md`
- **H10** ratelimit quota validation via constructor — landed `4a6a9e8b` → `package-layout.md` (constructor invariant pattern)
- **H11** idempotency schema v2 hardening — landed `07312d58` → `idempotency-and-concurrency.md`
- Open Highs (H2/H3/H5/H6/H8/H9/H11-remainder) → feed `errors-and-logging.md` + `idempotency-and-concurrency.md` + `persistence.md`.

## 15. Linkage to existing rules

Each bar doc carries:

```yaml
---
extends:
  - ~/.claude/rules/ecc/golang/<rule>.md
evidence:
  - wiki/reviews/2026-05-21-go-backend-review/<module>.md#<anchor>
enforced_by:
  - .golangci.yml:<linter>
---
```

This makes the bar auditable, traceable, and lets `wiki-curator` agent verify alignment.

---

## 16. Next step

Hand off to `/ecc:prp-plan` to produce:
- File-by-file implementation plan for `wiki/standards/golang/` v1.
- `.golangci.yml` baseline + CI wiring plan.
- Migration plan for legacy modules (#2a remaining Highs, #2b, #2c).
- Test plan: lint baseline acceptance, evidence-citation completeness check, bar-reference smoke test on next ECC review.
