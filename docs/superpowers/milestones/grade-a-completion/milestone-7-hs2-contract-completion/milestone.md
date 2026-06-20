# Milestone 7 — HS-2 Contract Completion (typed-body parity; honest H-D gate)

> **Program:** grade-a-completion
> **Governing spec:** `../mission.md` (§8 terminal acceptance; §2 Non-Goals; §9 HS-2 signal)
> **Status:** In progress — spec operator-approved at Phase-2 gate (2026-06-20, README + commit `45a03fa6`); F7.x execution running in this fresh session
> **Authored:** 2026-06-20 — before any F7.x feature began.
> **Opened by:** HS-2 — third consecutive Contract/API miss on the post-M6 terminal re-audit
> (`wiki/backend/_artifacts/architecture-re-audit-2026-06-19-post-m6.md`, `qa/mission-validation.md`, HEAD `5650b328`).
> **Operator scope decision (2026-06-20):** **typed-body parity**, not a full codegen-first
> StrictServerInterface rewire. Rationale recorded in `../README.md` hard-stops (2026-06-20 HS-2 row):
> the A− bar was reached **twice without** StrictServerInterface (templates std-server wrappers + typed
> bodies; IAM hand-rolled typed structs per ADR 0012), and auth + search have **no codegen pipeline**,
> so a full rewire is disproportionate to what §8 requires (contract ≥ A−, 0 Majors, H-D=0 — not a
> specific framework).
> **Appetite:** ≤1 session. Bounded: typed-response swaps on ~10 sites + ≤4 OpenAPI 200 schema
> declarations (documents) + 1 documents BE codegen regen + 1 FE codegen regen + 1 honest-gate redefinition.

---

## Objective

Close the 3 §8 pass-bar gaps that survived the post-M6 terminal re-audit so the next re-run passes:

1. **Contract/API ≥ A−** (currently **B** — improved from B− but still short; one Major + 10 H-D sites open).
2. **0 skeptic-confirmed Critical/Major** (currently **1**: audit export status `map[string]any` vs the
   already-generated, unused `AuditExportStatusResponse`).
3. **H-D class = 0** — and **redefine the acceptance gate so it is no longer blind**. M6 reported
   Grep A = 0 while 10 H-D sites survived, because the one-liner pattern `writeJSON.*map\[string\]any`
   cannot see the `writeFillInJSON` alias or multi-line map construction. The honest gate must count
   **every response-literal map on a public route**, and explicitly allowlist only non-response uses.

H-G is already 0 (PASS in both prior verdicts) — **do not regress F5.1/F5.2/F5.7 fixes**.
Module-boundaries (A−), composition (A−), authz (A−) — already lifted, do not touch.

---

## Appetite / rabbit holes

**In scope:** exactly the 5 features below — no wider refactor.

**Out of scope (YAGNI-ruthless):**
- **No full codegen-first StrictServerInterface rewire.** Per the operator's 2026-06-20 scope decision,
  M7 does **not** stand up new oapi-codegen pipelines for auth or search, and does **not** rewire any
  module's routing through generated `ServerInterface`/`NewStrictHandler`. Auth + search use hand-rolled
  typed structs (ADR 0012 legacy posture). If a feature starts to grow into pipeline-standup or routing
  rewire, **stop** (HS-2-internal) and surface — that is a successor mission, not M7.
- No FE feature work beyond the codegen-type regen the documents contract fix requires.
- No new product capabilities; no migration / schema work.
- No Minor sweep from the post-M6 re-audit §7 unless it sits inside a touched file (drive-by repair
  only — record in evidence). The 3 audit 405-Allow-header Minors (§7 #11–13) sit in the F7.1 file and
  MAY be repaired drive-by; they are not H-D and not required for the bar.
- No re-litigating downgraded findings (re-audit §5) — they are already Minor.
- No new authz/composition/observability/module-boundary work — those grades are lifted.

---

## Features

Executed in order. Each gets `spec.md → plan.md → evidence.md` before commit.

The surviving `map[string]any` uses split into two classes — **response literals** (kill all) and
**legit non-response uses** (keep, allowlisted in F7.5). The allowlist, enumerated at HEAD `5650b328`,
is: domain-mirror struct fields (`audit EventResponse.Payload`, `security signalItem.Evidence`),
internal audit-emit params (`recordAudit(... payload map[string]any)` in auth/iam/audit), and command
inputs (`controlleddocuments formData`, `documents ContentFormData`). None of these is a public response
body. Every feature below kills only **response literals**.

| Feature | What to implement | What to validate (acceptance) | Closes |
|---------|-------------------|-------------------------------|--------|
| **F7.1** `audit-typed` | Audit module **already has** generated `*Response` types in `internal/modules/audit/api/api.gen.go` but the hand-rolled mux handlers ignore them. Emit the generated types at 3 response sites in `audit/delivery/http/handler.go`: `:268-279` export-status → `AuditExportStatusResponse` (**the confirmed Major**); `:216-221` export-POST → `AuditExportResponse`; `:120-130` events-list (incl. the `page := map[string]any{...}` at `:120`) → `ListAuditEventsResponse` + `CursorPage`. Keep the `EventResponse.Payload map[string]any` field (`:51`) and the internal audit-emit `payload` (`:404`) — neither is a response literal. **Optional drive-by** (same file): fix the 3 405-Allow-header Minors (§7 #11–13) — record in evidence if done, skip cleanly if not. | The 3 response sites emit the generated types; wire JSON keys equal the OpenAPI declaration. Handler tests assert the typed shape (strict-decode). 0 response-literal `map[string]any` in `handler.go`. Build + tests green. | 1 confirmed Major + 2 H-D sites |
| **F7.2** `auth-typed` | Auth is **pre-codegen** (no `api.gen.go`, no `cfg.yaml`). Per ADR 0012 legacy posture, define hand-rolled typed Go response structs in `auth/delivery/http/` mirroring the OpenAPI `AuthLoginResponse` and `ChangePasswordResponse` schemas. Swap the 2 response sites: `handler.go:90-93` (login) and `:161-164` (change-password). Keep all 4 `recordAudit(... payload map[string]any)` internal audit-emit uses (`:83,94,112,189`) — not responses. | The 2 response sites emit hand-rolled typed structs whose JSON tags equal the OpenAPI `AuthLoginResponse`/`ChangePasswordResponse` schemas. Handler tests assert wire shape unchanged from baseline. 0 response-literal `map[string]any` in `auth/.../handler.go`. Build + tests green. | 2 H-D sites |
| **F7.3** `search-typed` | Search is **pre-codegen**. A local `SearchDocumentResponse` mirror struct already exists; the only gap is the envelope literal at `handler.go:134` (`map[string]any{"items": out}`). Define a hand-rolled `searchDocumentsResponse{ Items []SearchDocumentResponse }` typed struct and swap the one site. | The site emits the typed envelope; wire JSON `{"items":[...]}` unchanged. Handler test asserts shape. 0 response-literal `map[string]any` in `search/.../handler.go`. Build + tests green. | 1 H-D site |
| **F7.4** `documents-fillin-view-placeholder-typed` | Documents **has** a codegen pipeline (`internal/modules/documents/api/api.gen.go`). Contract-first: declare OpenAPI 200 body schemas for the 4 currently-undeclared routes (`getFillInSchema`, `putPlaceholderValue`, `getPlaceholderOptions`, `viewDocument` — only `viewDocument` has an operationId today), run `GOFLAGS=-mod=mod go generate ./internal/modules/documents/api/...`, then emit the generated types at the 4 response sites: `fillin_handler.go:58-62` and `:116-119`; `placeholder_options_handler.go:67,74`; `view_handler.go:46-51`. Note `view_handler.go:46` is `payload := map[string]any{"pdf_status":...}` built then written via `writeFillInJSON` — the multi-line/alias pattern Grep A is blind to. | OpenAPI declares 200 body schemas for all 4 ops. Generated `documents/api/api.gen.go` types match. Handlers emit the generated types. Handler tests assert wire JSON keys equal the OpenAPI declaration. 0 response-literal `map[string]any` (incl. via `writeFillInJSON`) in the 3 files. BE codegen fresh (no uncommitted diff after regen). Build + tests green. | 4 H-D sites (incl. the Grep-A-blind alias sites) |
| **F7.5** `honest-hd-gate-and-final-proof` | After F7.1–F7.4 commit clean: (a) **redefine the H-D acceptance gate** to be non-blind — the gate is the documented two-part measurement in "Milestone validation definition" §2 below (response-literal count = 0, allowlist re-derived from first principles). Record it in `wiki/architecture/api-contract.md` so future milestones inherit the honest gate, not Grep A alone. (b) Regen FE codegen (`npm run gen:api`) for the documents schema changes, contract-first order. (c) Update `wiki/architecture/api-contract.md` Last-verified stamp. (d) Final proof: response-literal H-D count = 0; whole-repo `go test -count=1 ./...` green. | The honest gate is documented + runs to 0 response-literal sites. FE codegen regen clean (no orphan types). Wiki stamp current. Whole-repo tests green. | Exit-bar proof + permanent gate-blindspot fix |

---

## Sequencing

```
F7.1  (audit — types already generated; emit)          ─┐
F7.2  (auth — hand-rolled typed structs; pre-codegen)   ├─ independent of each other
F7.3  (search — hand-rolled typed struct; pre-codegen)  ┘
F7.4  (documents — OpenAPI 200 decl + BE codegen regen + emit; contract-first)
F7.5  (honest gate redefinition + FE codegen regen + wiki stamp + final proof)
```

F7.1/F7.2/F7.3 are independent (different modules, no shared codegen). F7.4 is contract-first and
self-contained (documents OpenAPI → documents `api.gen.go` → handlers). F7.5 depends on all four prior
features being closed; the FE codegen regen in F7.5 reflects only F7.4's documents schema additions.

---

## Quality goals and constraints

1. **H-D response-literal count = 0** after F7.5 — proven by the honest two-part gate (validation
   definition §2), not by Grep A alone. The `writeFillInJSON`/multi-line blindspot must be closed.
2. **0 confirmed Majors** after F7.1 — the audit export-status site emits the generated
   `AuditExportStatusResponse`; handler test asserts wire-JSON parity.
3. **Contract-first regen order respected** (F7.4, F7.5) — OpenAPI → BE `api.gen.go` → FE codegen.
   Never the reverse. Auth/search hand-rolled structs are a deliberate ADR-0012 legacy exception, not a
   regen path.
4. **No prior-milestone regression** — H-G stays at 0; F5.1 / F5.2 / F5.7 untouched; M6's templates +
   IAM typed responses untouched.
5. **No scope creep** — fix only the cited response-literal sites + the documents contract-first regen
   they require + the gate redefinition. Mention-don't-fix anything else found. The allowlisted
   non-response `map[string]any` uses are **kept** — converting them is out of scope and would be
   gold-plating.
6. **HS-2-internal stop**: if any feature implies standing up a new codegen pipeline (auth/search) or
   rewiring routing through generated `ServerInterface`/`NewStrictHandler`, **STOP** and surface — that
   exceeds the operator's typed-body-parity scope and is successor-mission territory.
7. **Authz untouched** — ADR 0022 boundary respected; F0.1 / F5.6 fixes untouched. **H-PRE-1** untouched
   (no new authz-recording reads inside lock-holding tx).
8. **Never push** — commits authorized after verified work (CLAUDE.md §5.0), operator merges.

---

## Hard-stops

| ID | Trigger | Action |
|----|---------|--------|
| HS-1 | Milestone boundary | Operator gate before terminal re-audit re-run |
| HS-2 | A feature grows into new-pipeline standup (auth/search codegen) or routing rewire through generated `ServerInterface`/`NewStrictHandler` | STOP; surface boundary; propose as successor mission; do not exceed typed-body-parity scope within M7 |
| HS-3 | Build / `go generate` / `go test` prerequisite fails (e.g. stale codegen, vendor mismatch) | Repair (`runtime-contract-prereq`); rerun checkpoint; resume feature |
| HS-4 | milestone-validator returns FAIL | Open the validator's named fix feature; re-run its lifecycle; re-dispatch |
| HS-5 | Terminal re-audit (post-M7) misses §8 bar a **fourth** time | STOP — do not open M8 by default. Surface to operator: the bounded-sweep approach has now failed to close the dimension repeatedly; the full codegen-first rewire (the originally-framed HS-2 option) or a re-scoping of the §8 bar is the operator's call. |
| HS-6 | Off-plan discovery mid-milestone (e.g. a swap surfaces a new bug or runtime/spec drift outside the cited sites) | STOP; surface; replan before continuing |

---

## Milestone validation definition

Run by the independent `milestone-validator` subagent after all 5 features have `evidence.md`.

1. **Spec/plan conformance** — each `f7.x-<slug>/spec.md` carries a pre-commit approval line; each
   `evidence.md` Acceptance row maps to a re-runnable command.
2. **The honest H-D gate (two-part) re-run from clean state:**
   - **Part A — Grep A (necessary, not sufficient):**
     `grep -rEn 'writeJSON.*map\[string\]any' internal/modules/*/delivery/http/ --include='*.go'` → **0**.
   - **Part B — response-literal completeness (closes the blindspot):**
     `grep -rEn 'map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' | grep -v _test.go`
     → every surviving hit must be on the **allowlist** of non-response uses, re-derived from first
     principles by the validator (domain-mirror struct fields: `audit EventResponse.Payload`,
     `security signalItem.Evidence`; internal audit-emit params: `recordAudit(... payload map[string]any)`
     in auth/iam/audit; command inputs: `controlleddocuments formData`, `documents ContentFormData`).
     **Zero response-literal sites** — i.e. no `map[string]any` passed (directly or via a built local
     like `page :=` / `payload :=`, on one line or many) to `writeJSON` / `writeFillInJSON` / `WriteJSON`.
   - **H-G Grep B/C (re-audit §6)** → **0** (no regression of F5.1/F5.2/F5.7).
   - `GOFLAGS=-mod=mod go generate ./internal/modules/documents/api/...` → no uncommitted diff.
   - `go build ./...` → clean.
   - `go test -count=1 ./...` → 0 FAIL (full repo, from clean state).
   - Per-feature handler tests pass — wire-JSON keys equal the OpenAPI declaration (F7.1/F7.4) or
     unchanged from baseline (F7.2/F7.3).
3. **Senior review** of aggregate M7 diff — only the ~10 cited response-literal sites + the ≤4 documents
   OpenAPI 200 schema declarations + the 2 regen artifacts (documents BE `api.gen.go`, FE codegen) + the
   wiki gate-redefinition stamp. No pipeline standup, no routing rewire (would be an HS-2 violation).
4. **Regression** — M0–M6 sentinels green; H-G greps still 0; F5.1/F5.2/F5.6/F5.7 + M6 templates/IAM
   typed responses untouched.
5. **Quality bar:** H-D response-literal count = 0 absolute (allowlist is non-response uses only, no defer
   allowlist for response literals). Contract/API indicatively A− on dimension re-read. Forbidden list:
   - No `map[string]any` **response literal** surviving on any public delivery route, including via the
     `writeFillInJSON` alias or a built-then-written local (the M6 blindspot).
   - No symptom-patching (e.g. `writeJSON(any(map[string]any{}))` or other grep-evasion).
   - No fixture/mock proof passed off as real-provider proof.
   - No BE codegen regen committed without a matching OpenAPI source change.
   - No new codegen pipeline or routing rewire (HS-2 scope violation).

---

## Dependencies & constraints respected

- **Mission §2 Non-Goals:** no FE feature work, no new capabilities, no schema/migration redesign, no
  merge by agent.
- **Mission §3 D3 sequencing:** M7 is the bounded HS-2 contract-completion micro-milestone — does not
  reorder M0..M6.
- **Mission §9 HS-2 signal:** M7 **is** the operator's resolution of the third-miss HS-2 signal — a
  deliberate, evidence-backed choice of typed-body parity over full codegen-first rewire (README
  hard-stops 2026-06-20). The full-rewire option remains the documented fallback if M7's re-audit
  misses a fourth time (HS-5 above).
- **CLAUDE.md:** never read/print/commit `.env`; PowerShell for local startup; runtime truth beats
  docs; commits authorized but **never push**.
- **ADR 0012 (contract-first API):** F7.1/F7.4/F7.5 follow contract-first regen order. F7.2/F7.3 use
  hand-rolled typed structs (legacy posture preserved for pre-codegen auth + search modules).
- **ADR 0022 (authz boundary):** untouched.
- **H-PRE-1 (advisory-lock hazard):** untouched.
- **Memory `backend-target-architecture-governs-reviews`:** PRs cite REQ IDs; MUST deviations need an
  ADR. M7 introduces no MUST deviation (the gate-redefinition is a measurement clarification, not an
  architecture change).
- **Memory `workflow-model-balancing`:** sonnet for implementation/review, haiku mechanical, never
  fable, ≤15 concurrent.
- **Skill routing (CLAUDE.md §3):** backend HTTP/contract → `metaldocs-backend-api`; FE codegen →
  `metaldocs-tanstack-query`; prereq repair → `runtime-contract-prereq`; after structural change →
  `wiki-curator`.
