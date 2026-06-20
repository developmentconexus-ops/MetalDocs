# Feature F7.5 — Honest H-D gate + FE codegen regen + final proof

> **Milestone:** 7 — HS-2 Contract Completion  ·  **Folder:** `f7.5-honest-hd-gate-and-final-proof`
> **Status:** Approved 2026-06-20 — code change may begin.
> **Approved before code:** 2026-06-20 / leandrotca.work — inherited from the M7 Phase-2 operator
> approval (commit `45a03fa6`).

## Interview record (B1.5)

| # | Question | Answer / source |
|---|----------|----------------|
| 1 | Why redefine the gate? | M6 reported Grep A (`writeJSON.*map\[string\]any`) = 0 while 10 H-D sites survived — the one-liner is blind to the `writeFillInJSON` alias, the `WriteJSON` capital, and built-then-written multi-line locals (`page :=`, `payload :=`). The acceptance gate must count **every response-literal map on a public route**, not just the one-liner. |
| 2 | What is the honest gate? | The two-part measurement already specified in `milestone.md` §"Milestone validation definition" §2: **Part A** (Grep A → 0, necessary) + **Part B** (`grep -rEn 'map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' \| grep -v _test.go` → every survivor on the **non-response allowlist**, zero response literals). It must live in `wiki/architecture/api-contract.md` so future milestones inherit it. |
| 3 | What is the allowlist? | Non-response uses only: domain-mirror struct fields (`audit AuditEventItem.Payload` via the decode buffer, `security signalItem.Evidence`), internal audit-emit params (`recordAudit(... payload map[string]any)` in auth/iam/audit), command inputs (`controlleddocuments formData`, `documents ContentFormData`). |
| 4 | What FE regen is needed? | `npm run gen:api` (openapi-typescript) in `frontend/apps/web` — reflects only F7.4's 4 documents 200 schema additions. Contract-first order: OpenAPI (F7.4) → BE `api.gen.go` (F7.4) → FE types (here). |
| 5 | Final proof? | Whole-repo: Part A = 0, Part B = allowlist-only (0 response literals); `go generate` documents api → no uncommitted diff; `go build ./...` clean; `go test -count=1 ./...` → 0 FAIL. |
| 6 | HS-2 risk? | None — documentation + FE type regen + proof only. No pipeline standup, no routing rewire. |

## Consumer contract (FIRST)

**Consumers:** future milestone validators + maintainers (the gate definition); FE TS consumers of the
documents types (the regenerated `index.d.ts`). The gate is the source of truth for "H-D = 0".

## What this feature implements

1. Add a "Response-body typing gate (H-D — honest two-part)" section to
   `wiki/architecture/api-contract.md` documenting Part A + Part B + the allowlist, replacing reliance
   on Grep A alone.
2. `npm run gen:api` in `frontend/apps/web`; commit the regenerated `src/lib/api-types/index.d.ts`.
3. Update the api-contract.md `Last verified` stamp to 2026-06-20 (M7).
4. Final whole-repo proof (commands + output in `evidence.md`).

## Non-goals (mandatory)

- No new OpenAPI changes (F7.4 already declared the documents schemas).
- No FE feature work beyond the type regen.
- No converting allowlisted non-response `map[string]any` uses (out of scope; gold-plating).
- No new gate tooling/CI wiring beyond documenting the measurement (CI drift guard already exists §7).

## Validation Gate (concrete — approved before code)

| # | Acceptance criterion | Proof command | Real vs fixture |
|---|----------------------|---------------|------------------|
| 1 | Honest two-part gate documented in api-contract.md | section present with Part A + Part B + allowlist | real |
| 2 | Part A = 0 whole-repo | `grep -rEn 'writeJSON.*map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' \| grep -v _test.go` → 0 | real |
| 3 | Part B = allowlist only (0 response literals) | `grep -rEn 'map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' \| grep -v _test.go` → each hit allowlisted | real |
| 4 | FE codegen regen clean | `npm run gen:api`; `git diff --stat src/lib/api-types/index.d.ts` shows the 4 documents types | real |
| 5 | BE codegen fresh | `GOFLAGS=-mod=mod go generate ./internal/modules/documents/api/...`; `git diff --exit-code api.gen.go` → clean | real |
| 6 | Whole-repo build + tests green | `go build ./...`; `go test -count=1 ./...` → 0 FAIL | real |
| 7 | Wiki stamp current | api-contract.md `Last verified: 2026-06-20 (M7 …)` | real |

## ADR needed?

- [x] No — the gate redefinition is a measurement clarification (honest H-D), not an architecture
  decision (per milestone.md quality goal note). Recorded in api-contract.md, which is the governing
  contract doc.
