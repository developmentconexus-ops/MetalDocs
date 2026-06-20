# Mission Terminal Acceptance — Verdict (2026-06-20 post-M7 re-run)

> Written by: the mission-validator subagent (separation of powers). Validates against: ../mission.md §8.
> Run: 2026-06-20 · Code HEAD: `dadb8275` (judged at working HEAD `9da740d6`, a docs-only status flip; source identical).
> Judges artifact: `wiki/backend/_artifacts/architecture-re-audit-2026-06-20-post-m7.md`, with independent re-grep + file:line re-read.
> Verdict: see bottom.

## Pre-condition: program readiness for terminal validation

- README milestone status: M0–M6 = `passed` (validator PASS + HS-1 operator approval). M7 = validator PASS
  (2026-06-20), HS-1 operator gate pending. The re-audit was run at HEAD `dadb8275` which includes all M7
  source — the terminal validation is correctly positioned at the post-remediation HEAD.
- Build: `GOFLAGS=-mod=mod go build ./...` → exit 0 (clean). Re-run by validator.
- Tests: `GOFLAGS=-mod=mod go test -count=1 ./...` → 0 FAIL packages, 85 `ok`. Re-run by validator.
  (Green build/tests do not satisfy §8; the four gate checks are class/grade defects, not test failures.)

## Per-criterion results (the four §8 checks, ALL required)

| # | §8 criterion | Method run | Real evidence | Pass? |
|---|--------------|------------|---------------|-------|
| 1 | 3 formerly-C dims (module-boundaries, contract-api, composition) all ≥ A− | Independent file:line re-read of the two live contract/composition Majors + judged dimension grades | Module-boundaries **A−** (every prod IAM-table touch inside `iam/`; consumers via documented ports) → meets bar. **Contract/API B+** → FAIL: `iam/presence/handler.go:83` emits `map[string]any{"items": items}` after `WriteHeader(200)` on spec-declared `GET /iam/presence/snapshot` (openapi.yaml:527) while typed `PresenceSnapshotResponse` (api.gen.go:465) + strict `GetPresenceSnapshot200JSONResponse` (api.gen.go:2381) sit unused. **Composition B+** → FAIL: `observability/http.go:183-194` builds `payload := map[string]any{...}` json-encoded to wire on public `/api/v1/metrics`; no `MetricsResponse` Go type exists (grep over internal/+apps non-test = empty); `scheduler`/`db_pool` keys undeclared. | ❌ |
| 2 | 0 skeptic-confirmed Critical/Major | Independently re-read all 5 cited Majors at file:line | 5 confirmed Majors upheld (listed below). | ❌ |
| 3 | H-D class = 0 (zero response literals; handler ⊆ OpenAPI ⊆ FE codegen) | Re-ran Part A + Part B greps; re-read both gate-evasion sites | Path-scoped gate reports 0 (Part A empty; all 10 Part B survivors are recordAudit params / command-input / domain-mirror fields → correctly allowlisted per §5b). But honest H-D = **2**: `iam/presence/handler.go:83` + `observability/http.go:183-194` are live response literals on public spec-declared routes OUTSIDE the gated `internal/modules/*/delivery/http/` path. By the §8 intent ("handler emits ⊆ declared OpenAPI ⊆ FE codegen") these are H-D violations the path-scoped grep structurally cannot see. | ❌ |
| 4 | H-G class = 0 (no cross-module raw SQL on another module's table; no hardcoded domain-state) | Re-ran iam_users / iam_user_roles / "published" greps; re-read search reader; confirmed table ownership | Narrow greps = 0 (no cross-module IAM-table SQL; 7 `"published"` hits are doc-comments). But honest H-G = **1**: `search/infrastructure/v2documents/reader.go:40,65` issues raw SQL `FROM metaldocs.document_profiles` (a taxonomy-owned table — taxonomy/infrastructure/repository.go + family_repository.go own it), hardcoding `family_code`/`code`/`tenant_id` + sentinel-tenant fallback, with NO taxonomy port (search imports no taxonomy package). The §8 H-G definition ("no raw SQL against another module's owned table") is violated; the grep only tests the two IAM tables. | ❌ |

## The 3 formerly-C dimensions — explicit ≥ A− call

- **Module boundaries / DDD: A−** — meets the bar (the baseline-C IAM-table reach is fully resolved via ports). The one search→taxonomy Major caps it at A− and is the Check-4 honest-H-G violation.
- **Contract / API layer: B+ — DOES NOT reach A−** (one live untyped-200 Major: presence).
- **Composition / observability: B+ — DOES NOT reach A−** (one live untyped-200 Major: metrics).

## Skeptic-confirmed Critical/Major I independently uphold (all 5)

| # | Sev | Dimension | Finding | File:Line (re-read & confirmed) |
|---|-----|-----------|---------|---------------------------------|
| 1 | Major | Sessions / auth lifecycle | Account deactivation not enforced on live sessions (CWE-613); deactivate issues `is_active` UPDATE with no session revoke | `auth/application/service.go` (deactivate path; change-password/reset DO revoke, deactivate deliberately does not) |
| 2 | Major | Middleware / HTTP kernel | Go 1.22 method-routed 405s return stdlib `text/plain`, bypassing D-03 problem+json | `documents/delivery/http/handler.go:177-178` (+ iam, approval) |
| 3 | Major | Module boundaries / DDD | Search raw SQL on taxonomy-owned `document_profiles`, no port (honest H-G) | `search/infrastructure/v2documents/reader.go:40,65` |
| 4 | Major | Contract / API | Spec-declared presence-snapshot route emits `map[string]any`; typed `PresenceSnapshotResponse` unused (honest H-D, gate-evading) | `iam/presence/handler.go:83` |
| 5 | Major | Composition / observability | `/api/v1/metrics` emits untyped `map[string]any`; no `MetricsResponse` type; undeclared keys (honest H-D, gate-evading) | `observability/http.go:175-196` |

My independent reads agree with the re-audit artifact on every load-bearing claim. I found no basis to refute or downgrade any of the 5; I found no additional Critical/Major beyond them.

## Pass bar

- Bar (§8): "a fresh, independent re-run of the F5.1 10-dimension re-audit at the post-M4 HEAD passes the §6 bar — (1) module-boundaries, contract-api, and composition all ≥ A−; (2) 0 skeptic-confirmed new Critical/Major; (3) H-D = 0; (4) H-G = 0."
- Met? **No.** 0 of 4 checks pass. Deciding evidence: two live untyped-200 response literals on public spec-declared routes (presence, metrics) fail Checks 1, 2, and 3; one cross-module raw-SQL reach (search→taxonomy) fails Checks 1(implication), 2, and 4. The program's documented H-D/H-G gates are path-scoped to `internal/modules/*/delivery/http/` and the two IAM tables — a scope narrower than the §8 intent — and every failing site sits just outside that scope and is live at HEAD.

## Forbidden-list (any hit = FAIL)

- [x] Fixture/mock passed off as real-provider proof — NOT triggered (failures are real source defects, independently grepped/read at HEAD).
- [x] A criterion marked pass without a command actually run — NOT triggered (no criterion passed; all four FAIL on commands I ran).
- [x] Split-brain / guessed contract surfaced in the aggregate diff — PRESENT and is itself a failing finding: handler emits diverge from declared OpenAPI / generated types (presence, metrics).
- [x] Self-judged / validator edited or fixed code — NOT triggered (read-only; verdict file only; no source touched, no status flipped, no push).

## Verdict

- VERDICT: FAIL
- Failed criteria: 1, 2, 3, 4 (all four).
- On FAIL — the bounded remediation needed to clear them (HS-5; and per artifact §9 this is the operator's HS-2 decision, Contract/API having now missed A− a fourth consecutive time B+→B−→B→B+):
  - (a) Route presence-snapshot (`iam/presence/handler.go:83`) through the existing typed `GetPresenceSnapshot200JSONResponse` model — eliminate the `map[string]any` literal. Closes one Contract Major + one honest-H-D site.
  - (b) Introduce a declared `MetricsResponse` type (incl. `scheduler`/`db_pool`) and emit it typed at `observability/http.go`, with the OpenAPI schema updated to match runtime. Closes one Composition Major + one honest-H-D site. (Alternative: explicitly exempt operational/off-spec endpoints in §8 + §5b and widen the gate to the true in-scope surface — artifact §9 option B.)
  - (c) Add a taxonomy port for `document_profiles` family-code lookup and route `search/.../reader.go` through it (read-only, off any lock-holding tx per H-PRE-1). Closes the boundary Major + the honest-H-G site.
  - (d) Enforce deactivation session-revoke in the IAM deactivate path (mirror change-password/reset), and add a fallthrough/405 problem+json interceptor for Go-1.22 method-routed mux.
  - After remediation: re-run the F5.1 fan-out, re-dispatch `mission-validator`. The mission stays open; the main session does NOT declare done and the operator does NOT give Grade-A sign-off at HEAD `dadb8275`.

VERDICT: FAIL
