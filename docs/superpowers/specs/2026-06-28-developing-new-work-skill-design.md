# `developing-new-work` — system-impact orientation skill (design)

**Date:** 2026-06-28
**Status:** Approved design. Build via `skill-creator`; this doc is the contract.
**Scope:** A new project-local Claude skill at `.claude/skills/developing-new-work/`. No runtime/code change to MetalDocs itself, except documentation wiring (CLAUDE.md Context Map + Orientation rule). First consumer: SP-1 `tokens` module.

---

## 1. Problem & intent

CLAUDE.md already states two meta-rules — the **Orientation rule** ("before planning any new feature or improvement, state (a) owning module(s), (b) invariants it must satisfy, (c) read the owning `wiki/modules/<name>.md`") and **Global Maximum, Not Local Maximum** ("judge the foundation first; do not optimize inside a patch"). They are prose, not actionable: nothing forces them, nothing produces an auditable record, and every new piece of work re-derives the same system map from scratch.

**Intent:** turn those two rules into a binding, **checklist-driven orientation gate** that runs *before* design. Given a one-line intent, it walks a **pre-baked, static** map of the system (modules, invariants, wiring touchpoints, frameworks, governance), judges fit and foundation, emits a written **system-impact analysis** + a **Green / Yellow / Red** verdict, then hands the analysis to brainstorming. Red is a **hard block** — design cannot start until the named redesign gate clears.

**Non-goal:** re-analyzing the codebase on every run. The MetalDocs backend base is mature and frozen (memory: *Backend is Grade-A, signed off 2026-06-21; module-boundaries A*). The skill is a checklist, not a fan-out analyzer. It reads code only to verify a single uncertain anchor.

## 2. Why a static checklist (binding decision)

The skill's engine is a **static, cached checklist**, not per-run discovery. Justification (evidence):

- **Stable base ⇒ static checklist pays off.** The 14 modules, middleware chain, invariants, and wiring touchpoints are settled contracts. Re-deriving them per run buys nothing and costs thousands of tokens (the 6-agent fan-out used to *build* this checklist is exactly what must not repeat).
- **The repo already governs itself by frozen-list-plus-guard.** Machine checklists pin the structure: `TestCapabilityRegistrySize` hard-pins the capability count (`internal/modules/iam/domain/model_test.go:90`), `TestEveryCapabilityClassified` (`internal/modules/iam/domain/capability_scope_test.go:10`), the `.github/workflows/module-boundaries.yml` CI guard, and `scripts/check-test-discipline.sh` (R1–R4). The skill's checklist is the agent-facing mirror of guards the repo already trusts.
- **Staleness is the only real risk, and the repo already has the antidote.** Mitigations, all lifted from existing governance:
  1. **Hybrid content** — checklist text is inline (zero reads on a normal run) **and** every item carries a `file:line` / wiki-REQ anchor.
  2. **Targeted verify** — when (and only when) an item is genuinely uncertain, read *that one anchor* (1–2 reads), never re-map the system.
  3. **`Last verified` stamp** on the checklist (same convention as `wiki/standards/documentation-governance.md`). Refresh only when the base structure changes (new module, invariant change) — rare, deliberate, never per-run.
  4. **Runtime wins ties** — if checklist and code disagree, code is truth (CLAUDE.md rule); the disagreement *triggers* a checklist refresh.
  5. **CI guards are the backstop** — an invariant violation the stale checklist missed still reds the build.

Rejected alternatives: *self-contained, no anchors* (silent drift, no cheap re-check); *wiki-pointer only* (every run opens wiki docs — violates the no-rescan requirement).

## 3. Placement in the pipeline

```
intent (one line)
   │
   ▼
[developing-new-work]  ── Red ──▶ redesign gate (no design until cleared)
   │ Green / Yellow
   ▼
superpowers:brainstorming  →  spec  →  superpowers:writing-plans  →  implementation
```

The skill runs **first**, before brainstorming. Its artifact is the rails brainstorming designs within. On Green/Yellow it ends by invoking `superpowers:brainstorming` with the analysis as input; on Red it stops and surfaces the redesign gate.

## 4. The artifact

**Path:** `docs/superpowers/analysis/YYYY-MM-DD-<slug>-system-impact.md`
**Template:** `templates/system-impact-analysis.md` (in the skill).

Ten sections. Same shape for **module** and **feature** work; a feature marks module-only rows (5, parts of 4/9) **N/A** with a one-line reason. Each section forces a whole-system answer before any design exists.

1. **Classify & own** — work type (module | feature); owning module(s); modules explicitly *not* owning it; cross-module edges and their direction. (CLAUDE.md Orientation rule)
2. **Foundation verdict** — is the base sound, or legacy/patch/workaround? If patchy, name the global-maximum structure + trade-off, or STOP. (Global-Maximum rule)
3. **Invariant alignment** — the 6 non-negotiables, each: *touched? how satisfied? helper to reuse?*
   - AuthZ = capabilities not roles · contract-first OpenAPI · multi-tenant (`tenant_id`/tx-local GUC) · async = transactional outbox · DB-enforced invariants · cross-module via published interface only.
4. **Capability wiring** (if new cap) — the 10-touchpoint ordered checklist (see §6.2).
5. **Module wiring** (if new module) — the ordered checklist (see §6.3).
6. **Frameworks to reuse, not reinvent** — `TxRunner`, `tenant.FromContext`, `authz.SeedTxIdentity`/`authz.Require`, `problem.Write`, `httpresponse`, `audit.NewEvent`, the outbox repo, `testdb` factory. (the "we create frameworks" standard)
7. **Contract & data** — OpenAPI-first flow; migration conventions; expand/contract for destructive changes.
8. **Test & QA plan** — which canonical framework (`testdb`, R1–R4 discipline), which of the 6 QA gates apply, the evidence shape required.
9. **Docs / ADR** — `wiki/modules/<name>.md` + `<name>-tech-debt.md` + index updates; ADR required? (MUST-deviation or policy change ⇒ yes); REQ IDs cited.
10. **Verdict** — Green (proceed) / Yellow (proceed; ADR or risk flagged) / Red (STOP; redesign gate first). Plus the locked constraints handed to brainstorming.

## 5. Skill workflow (`SKILL.md`)

Phases mirror the artifact; each is a TodoWrite item.

1. **Orient** — classify; load `references/` checklist (no code reads); choose module vs feature branch.
2. **Foundation** — apply §2 of the artifact; AS-2 hard-stop if patchy.
3. **Invariants** — walk `references/invariant-checklist.md`.
4. **Wiring** — walk `references/capability-wiring.md` and/or `references/module-wiring.md` for the branch.
5. **Frameworks** — walk `references/frameworks-catalog.md`.
6. **Test/QA + Docs/ADR** — walk `references/test-qa-gates.md` and `references/docs-adr-governance.md`.
7. **Targeted verify** — for any item the run cannot answer with confidence, read its single anchor (1–2 files max). Never wholesale.
8. **Verdict + handoff** — write the artifact; Green/Yellow → invoke `superpowers:brainstorming` with it; Red → stop at the redesign gate.

**Hard-stops:**

| ID | Trigger | Action |
|----|---------|--------|
| AS-1 | An invariant (§3) would be violated | STOP; record the violation; require ADR or redesign before design |
| AS-2 | Foundation is a patch/legacy and the work would optimize inside it | STOP; propose the global-maximum structure + trade-off; operator decides |
| AS-3 | Owning module is ambiguous | STOP; resolve the boundary (verify the candidate anchors) before continuing |

A Red verdict (any unresolved AS-*) **blocks** the brainstorming handoff.

## 6. `references/` content (pre-baked from 2026-06-28 research)

Each file opens with a `Last verified: 2026-06-28` stamp and is inline-complete with anchors. Source content (to be transcribed into the reference files at build time):

### 6.1 `invariant-checklist.md` — the 6 non-negotiables + reuse helper + anchor
- Multi-tenant: `tenant.FromContext` (`internal/platform/tenant/context.go:27`), `authz.SeedTxIdentity` (`internal/modules/iam/authz/context.go:58`); every tenant table has `tenant_id`; cross-tenant URL → 404.
- Middleware chain (inherited): `apps/api/cmd/metaldocs-api/chain.go:25`. New routes never re-wire auth/errors.
- Errors RFC 9457: `internal/platform/problem/problem.go:76` (`Write`), codes `internal/platform/problem/codes.go:9`. Never bare `http.Error`.
- Contract-first: `api/openapi/v1/openapi.yaml`; module `api/cfg.yaml` + `gen.go`; `go generate ./internal/modules/<m>/api/...`.
- Async outbox: `internal/modules/render/fanout/staging_outbox.go:29`; enqueue in the business tx; consumers idempotent.
- Cross-module access via published Go interface only (domain `port.go` / application `ports.go`); never another module's repo/SQL/domain.
- TxRunner: `internal/platform/db/runner.go:21` (`Do`/`DoReadOnly`); services depend on the port, not `*sql.DB`; nil tx rejected.

### 6.2 `capability-wiring.md` — add a capability (10 touchpoints, ordered)
1. const + `validCapabilities` — `internal/modules/iam/domain/model.go:90` / `:134`.
2. scope classify — `internal/modules/iam/domain/capability_scope.go:36` (ScopeTenant | ScopeArea).
3. tier-1 route→cap rule — `apps/api/cmd/metaldocs-api/permissions.go` (forgetting = silent privilege escalation, default VisibilitySessionRequired).
4. tier-2 in-tx — `authz.Require(ctx, tx, cap, areaCode)` after `authz.SeedTxIdentity`; pattern `internal/modules/templates/application/create.go:63`.
5. seed grants — `db/reference-data/0001_product_reference_data.sql:17` (per role; system_admin bypasses).
6. DB tripwire — `db/baseline/0001_current_schema.sql` constraints `ck_cap_format`, `ck_cap_not_legacy`.
7. guard tests stay green — `TestEveryCapabilityClassified`, `TestAreaGradeCapabilitySet` (`capability_scope_test.go`).
8. bump `TestCapabilityRegistrySize` `const want` — `model_test.go:90` (the one mandatory manual edit; **targeted-verify the current count here**).
9. CI capability-coherence (5-surface) — REQ-AUTHZ-5 (`wiki/architecture/backend-target-architecture.md`).
10. H-PRE-1 — never call an authz-recording read inside a lock-holding atomic tx; hoist off-tx (memory `advisory-lock-deadlock-constraint`).

### 6.3 `module-wiring.md` — birth a module (ordered; exemplars taxonomy/templates)
folders `{api,application,domain,delivery/http,infrastructure}` → domain entities + `port.go` interfaces → application service + `ports.go` (consumer ports) → infrastructure repo (own tables only; authz GUC + `authz.Require` in-tx) → delivery `Handler` + `RegisterRoutes(mux)` → api `cfg.yaml` (`include-tags`) + `gen.go` → OpenAPI `tags:` entry + every route tagged → optional `module.go` `New(Dependencies)` (panic on nil deps) → composition root wiring in `apps/api/cmd/metaldocs-api/main.go` (+ worker/jobs if async) → migration `db/migrations/0NNN_*.sql` → `wiki/modules/<name>.md` + `<name>-tech-debt.md` + `wiki/modules/index.md`.

### 6.4 `frameworks-catalog.md` — reuse table
`TxRunner` · `tenant.FromContext` · `authz.SeedTxIdentity`/`Require` · `problem.New`/`Write` · `httpresponse.WriteError` · `audit.NewEvent`/`Record`/`RecordTx` · outbox repo · `contracts.Decode` (strict JSON) · `testdb.Open`/factory builders. Each with import path + "use when".

### 6.5 `test-qa-gates.md`
Canonical integration framework `tests/integration/testdb/` (`Open(t)`, factory builders, `SeedWithCaps`, `Qualified`); `//go:build integration`; R1–R4 (`wiki/quality/test-discipline.md`, ADR 0034). The 6 QA gates (`wiki/quality/qa-operating-system.md`) and which apply to module vs feature. Evidence: commands + outcomes + review/QA disposition + bounded defers (CLAUDE.md). Commands: `go build ./...`, `go test ./...`, `.\scripts\check-system-runnable.ps1`.

### 6.6 `docs-adr-governance.md`
`wiki/modules/<name>.md` 12-section structure (exemplar `wiki/modules/taxonomy.md`) + tech-debt sister doc + index entry; `Last verified` header convention; REQ-ID citation against `wiki/architecture/backend-target-architecture.md` (MUST deviation ⇒ ADR); ADR template + numbering (`wiki/decisions/`), key ADRs 0007/0008/0022/0034.

## 7. Wiring into the repo

- **CLAUDE.md Context Map** — new row: *"Starting any new module or feature → `developing-new-work` skill."*
- **CLAUDE.md Orientation rule** — append: *"Operationalized by the `developing-new-work` skill — run it before brainstorming."*
- No `marketplace.json`/registry change needed; project-local skills under `.claude/skills/` are auto-discovered.

## 8. Build & verification

- **Build:** via `skill-creator`. Produce `SKILL.md` (frontmatter + workflow §5), `templates/system-impact-analysis.md` (§4), and the six `references/*.md` (§6).
- **Frontmatter `description`** must trigger on: "new module", "new feature", "add X / build Y / implement Z", "does this fit the architecture". Must NOT collide with `gitnexus-impact-analysis` (code blast-radius) — this is *architecture orientation, pre-design*.
- **Verification (dogfood):** run `developing-new-work` on SP-1 *"add a tenant token dictionary"*. Pass = a complete `…-tokens-system-impact.md` with the 10 sections, the cap-wiring and module-wiring checklists filled from the references (≤2 targeted reads), and a Yellow verdict naming the ADR-superseding-0008 constraint. That artifact becomes SP-1's pre-design doc.

## 9. Out of scope

- SP-1 `tokens` implementation (this skill only *orients* it).
- Automating the checklist refresh (manual, stamp-driven for now).
- A "fix/refactor" branch (Q1 scoped this to new module + new feature; revisit later).
- Any change to `gitnexus-impact-analysis`.

## 10. Risks & trade-offs

- **Checklist drift** — mitigated by §2 (hybrid anchors, `Last verified`, runtime-wins, CI backstop). Accepted residual: a base change not followed by a refresh leaves stale items until the next targeted-verify or CI red.
- **Duplication (checklist vs wiki)** — deliberate; the price of zero-read runs, and consistent with how the wiki already caches code facts.
- **Over-gating small features** — mitigated by the N/A-row scaling; a feature run is mostly invariants + test/QA + verdict.
