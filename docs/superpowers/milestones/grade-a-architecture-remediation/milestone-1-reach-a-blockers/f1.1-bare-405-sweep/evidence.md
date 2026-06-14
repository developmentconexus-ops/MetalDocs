# Feature F1.1 — Evidence (bare-405 sweep)

> **Milestone:** 1 (Reach-A Blockers)  ·  **Feature:** `f1.1-bare-405-sweep`  ·  **Closed:** 2026-06-14
> **Contract:** `spec.md` — RFC 9457 `application/problem+json` 405 + RFC 9110 `Allow` header + body
> `{"title":"Method not allowed","status":405,"code":"METHOD_NOT_ALLOWED"}`, read from source
> (`problem.go`, `httpresponse.WriteMethodNotAllowed`), governed by D-03. No new contract authored.

## What was implemented

Replaced every bare `w.WriteHeader(http.StatusMethodNotAllowed)` with a call to the existing canonical
helper `httpresponse.WriteMethodNotAllowed(w, "<allow>")` — killing the **class** across all four
delivery/platform packages that emitted it. Producer now matches the consumer contract in `spec.md`
(problem+json envelope + `Allow` header + `METHOD_NOT_ALLOWED` code), not a variant. No new helper,
no success-path change, no route/OpenAPI shape change, `audit/delivery/http` left untouched.

10 sites swept (Allow value = each route's real method guard, read from source):

| Package | Site | Allow |
|---------|------|-------|
| `auth/delivery/http` | `handler.go` login / logout / change-password | `POST` |
| `auth/delivery/http` | `handler.go` me | `GET` |
| `iam/delivery/http` | `admin_handler.go` overview | `GET` |
| `iam/delivery/http` | `admin_handler.go` role-upsert | `POST` |
| `iam/delivery/http` | `sessions_handler.go` list | `GET` |
| `iam/delivery/http` | `sessions_handler.go` by-id | `DELETE` |
| `platform/featureflags` | `handler.go` | `GET` |
| `platform/observability` | `http.go` MetricsHandler | `GET` |

Commits (all on `main`):
- `2c920678` — auth: problem+json 405 via shared helper + table test
- `50601a2b` — iam (admin + sessions): problem+json 405 via shared helper + table test
- `fa792149` — platform (featureflags + observability): problem+json 405 via shared helper + tests

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | per-site table tests `*_method_not_allowed_test.go` asserting status+CT+Allow+code | assertions are red against the prior `w.WriteHeader(405)` (no `Allow`, no problem+json body, no `code`) and green after the helper swap | real |
| Static (build) | `go build ./...` | `build exit=0` | — |
| Targeted test (fresh) | `go test -count=1 -v -run 'MethodNotAllowed\|405' ./internal/modules/auth/delivery/http/... ./internal/modules/iam/delivery/http/... ./internal/platform/featureflags/... ./internal/platform/observability/...` | all PASS — auth(login/logout/me/change-password), iam-admin(overview/role-upsert), iam-sessions(list/by-id), featureflags, observability | real |
| AC1 root-cause grep | `grep -rnE 'WriteHeader\((http\.StatusMethodNotAllowed\|.*405)' <4 swept pkgs> --include='*.go' \| grep -v _test.go` | empty (`exit=1`) — **class dead** | real |
| Runtime proof (live API on `:8081`) | logged in (`POST /api/v1/auth/login` → 200 + session), then wrong-method per package (session + `Origin` to clear the session-CSRF guard) | see runtime table below — 5 sites across all 4 packages return the exact contract | real |

### Runtime (live `:8081`, observed by us)

| Request | Response |
|---------|----------|
| `DELETE /api/v1/auth/login` (AC4 literal) | `405` · `Allow: POST` · `application/problem+json` · `{"title":"Method not allowed","status":405,"code":"METHOD_NOT_ALLOWED"}` |
| `POST /api/v1/auth/sessions` (iam sessions) | `405` · `Allow: GET` · `application/problem+json` |
| `DELETE /api/v1/iam/admin/overview` (iam admin) | `405` · `Allow: GET` · `application/problem+json` |
| `POST /api/v1/feature-flags` (featureflags) | `405` · `Allow: GET` · `application/problem+json` |
| `POST /api/v1/metrics` (observability) | `405` · `Allow: GET` · `application/problem+json` |

> Note (not a defect): the runtime middleware stack runs `auth` then a session-CSRF `Origin` guard
> **before** the handler's method branch, so the bare spec curl (no session/Origin) surfaces 401/403
> first. The literal AC4 405 is reached with a valid session + matching `Origin`; success paths are
> unchanged. This middleware-ordering observation is contract-adjacent, not in F1.1's scope.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| AC1 — **0** bare-405 in swept packages (class killed at root) | yes | AC1 grep row → empty |
| AC2 — wrong-method → 405 + problem+json + `code:"METHOD_NOT_ALLOWED"` + non-empty `Allow`, per-site test, red-first | yes | TDD + targeted-test rows (10 sites) |
| AC3 — build clean, no regression, no success-path change | yes | build `exit=0`; targeted tests PASS; only wrong-method branch changed |
| AC4 — live wrong-method returns the contract | yes | runtime table — `DELETE /api/v1/auth/login` → 405 + `Allow: POST` + problem+json, plus 4 more sites |

## Review disposition

- **Spec-compliance review** (general-purpose, sonnet): `SPEC-COMPLIANT`. Contract honored at all 10
  sites with correct per-route `Allow`; AC1 grep empty; AC2 tests assert the full contract; all five
  non-goals respected (no success-path change, audit untouched, no new helper, no route/OpenAPI shape
  change, only the 4 swept packages touched); nothing extra. No findings → no fix loop.
- **Code-quality review** (cavecrew-reviewer, sonnet): clean — no Critical/Major/Minor. Idiomatic Go,
  matches surrounding handler style, `Allow` values consistent with each route's real guard, tests use
  `httptest` correctly (featureflags routes through the mux, observability calls the returned handler
  directly — both meaningful), import hygiene correct, no dead code/debug. No findings → no fix loop.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Middleware ordering surfaces 401/403 before the handler 405 for session-bearing unsafe-method requests | Contract-adjacent; the 405 contract itself is correct and reachable. Whether the method guard should precede the CSRF/origin guard is a separate decision, not a bare-status defect | If pursued, fold into M2 contract-tail (H-D) review of the middleware chain; owner = backend agent |
| `sessions_handler.go` registers path-only routes while `iam/api/api.gen.go` also registers method-specific `GET/DELETE /auth/sessions` (two registration surfaces) | Pre-existing tri-source/route-truth question, explicitly **out of F1.1 scope** (F1.1 = mechanical class-kill). The swept handler is live and emits the contract (runtime-confirmed) | F1.2 (route-truth-table) / M2 H-D drift sweep; owner = backend agent |
