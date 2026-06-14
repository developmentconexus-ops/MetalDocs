# Feature F1.1 — bare-405 sweep — Spec

> **Milestone:** 1 (Reach-A Blockers)  ·  **Folder:** `f1.1-bare-405-sweep`
> **Closes:** Grade-A blockers #1, #2 + error-contract (H-B) tail.

## Interview record (fail-closed gate)

The consumer contract here is **not ambiguous and was not guessed** — it is the pre-existing
RFC 9457 error envelope already in the codebase, governed by decision **D-03 ("no bare-status
error responses")**. No operator interview was required: the contract is read from source
(`problem.go`, `httpresponse.WriteMethodNotAllowed`), not invented.

| Q | A | Source |
|---|---|--------|
| What shape must a 405 take? | RFC 9457 `application/problem+json` body + RFC 9110 `Allow` header | `internal/platform/problem/problem.go`, `internal/platform/httpresponse/response.go:22` |
| Is there already a canonical helper? | Yes — `httpresponse.WriteMethodNotAllowed(w, allow)` | `response.go:20-25` |
| Is a new decision/ADR needed? | No — D-03 already governs; this closes instances of a decided class | governing spec §8 |

## Consumer contract (FIRST — before any producer change)

- **Consumer(s):** any HTTP client that hits a registered MetalDocs route with a disallowed method
  — the web FE, monitoring/health probes, `curl`, and any error-contract consumer that parses
  `application/problem+json`.
- **Contract shape (the producer must match this exactly):**
  - Status `405`.
  - Header `Content-Type: application/problem+json`.
  - Header `Allow: <comma-separated allowed methods>` (RFC 9110).
  - Body = `problem.Problem`: `{ "title": "Method not allowed", "status": 405, "code": "METHOD_NOT_ALLOWED" }`.
- **Source of truth (read, not guessed):** `internal/platform/problem/problem.go` (envelope + `Write`
  sets `application/problem+json`), `internal/platform/problem/codes.go:22` (`CodeMethodNotAllowed =
  "METHOD_NOT_ALLOWED"`), `internal/platform/httpresponse/response.go:22` (`WriteMethodNotAllowed`).
  The producer change is: **call this existing helper** at each bare site. No new contract is authored.

## What this implements

Replace every bare `w.WriteHeader(http.StatusMethodNotAllowed)` (and any hand-rolled 405) with
`httpresponse.WriteMethodNotAllowed(w, "<allowed methods for that route>")`, across the swept
delivery/platform packages — killing the **class**, not only the listed instances.

Known sites (10, verified by grep at spec time — implementer re-derives, does not trust this list blindly):

| Package | Site | Route / handler | Allowed method(s) |
|---------|------|-----------------|-------------------|
| `auth/delivery/http` | `handler.go:66` | `/api/v1/auth/login` | `POST` |
| `auth/delivery/http` | `handler.go:101` | `/api/v1/auth/logout` | `POST` |
| `auth/delivery/http` | `handler.go:121` | `/api/v1/auth/me` | `GET` |
| `auth/delivery/http` | `handler.go:134` | `/api/v1/auth/change-password` | `POST` |
| `iam/delivery/http` | `admin_handler.go:149` | admin/overview | (implementer reads the guard) |
| `iam/delivery/http` | `admin_handler.go:297` | (extra site beyond spec's 7 — class mandate covers it) | (implementer reads) |
| `iam/delivery/http` | `sessions_handler.go:67` | sessions | (implementer reads) |
| `iam/delivery/http` | `sessions_handler.go:143` | sessions | (implementer reads) |
| `platform/featureflags` | `handler.go:33` | `/api/v1/feature-flags` | `GET` |
| `platform/observability` | `http.go:149` | `MetricsHandler` | `GET` |

`platform/featureflags` and `platform/observability` do not yet import `httpresponse` — add the import.
`audit/delivery/http` is already canonical (`writeProblem` + `problem.New`) — **out of scope, do not touch.**

## Non-goals (mandatory)

- **No success-path change** — only the wrong-method branch changes; 2xx behavior is byte-identical.
- **No new routes, renames, or OpenAPI shape change** — 405 is the already-implied error contract.
- **No new helper / abstraction** — `WriteMethodNotAllowed` already exists; reuse it.
- **No touching 405s outside the swept packages** (audit is already canonical; leave it).
- **No new ADR** — D-03 governs.

## Validation Gate (acceptance — objectively checkable)

| # | Criterion | Named test / proof | Real vs fixture |
|---|-----------|--------------------|-----------------|
| AC1 | **0** bare-405 in swept packages (class killed, root cause) | `grep -rnE 'WriteHeader\((http\.StatusMethodNotAllowed\|.*405)' internal/modules/auth/delivery/http internal/modules/iam/delivery/http internal/platform/featureflags internal/platform/observability` excl. `_test` → empty | real |
| AC2 | Wrong-method → 405 + `application/problem+json` + `code:"METHOD_NOT_ALLOWED"` + non-empty `Allow` | new Go table test per swept handler (TDD: **red first**, then green) — e.g. `internal/modules/auth/delivery/http/handler_problem_test.go` extended; new tests for iam/featureflags/observability | real |
| AC3 | Build clean, no regression, no success-path change | `go build ./...`; `go test ./internal/modules/auth/... ./internal/modules/iam/... ./internal/platform/featureflags/... ./internal/platform/observability/...` | real |
| AC4 | Runtime: live wrong-method request returns the contract | API on `:8081` via `.\scripts\start-api.ps1`; `curl -i -X DELETE http://localhost:8081/api/v1/auth/login` → 405 + problem+json + `Allow: POST` | real |

> TDD note: AC2 tests are written **red first** (assert problem+json/Allow before the fix), then made
> green by the producer change. A suite-level "all green" without these per-site assertions is not acceptance.

## ADR needed?

- [x] No — reuses decided contract D-03; no new architectural decision.

## Approval

Consumer contract + validation gate **approved** — covered by the operator's M1 milestone-spec
sign-off (2026-06-14), which declared F1.1's acceptance verbatim; this spec adds only the
read-from-source contract detail and introduces no new choice. No implementation began before this line.
