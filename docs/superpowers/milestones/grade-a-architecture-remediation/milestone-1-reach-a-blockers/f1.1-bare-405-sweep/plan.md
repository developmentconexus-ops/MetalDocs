# F1.1 bare-405 sweep — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace every bare `w.WriteHeader(http.StatusMethodNotAllowed)` in the swept delivery/platform packages with the canonical `httpresponse.WriteMethodNotAllowed(w, allow)` so wrong-method requests return an RFC 9457 `application/problem+json` 405 with an `Allow` header — killing the error-contract (H-B) class, not just the listed instances.

**Architecture:** The contract and helper already exist (`internal/platform/problem` + `internal/platform/httpresponse/response.go:22`, decision D-03). This feature only *calls* the existing helper at each bare site; it authors no new contract. Each handler's method guard returns before its service-nil check, so 405 branches are unit-testable with zero-dependency handler construction. TDD per package: red test asserting problem+json/Allow, then the one-line producer change.

**Tech Stack:** Go (`net/http`, `httptest`), existing `problem` + `httpresponse` platform packages.

**Consumer contract (read, not guessed):** see `spec.md` — status 405, `Content-Type: application/problem+json`, `Allow: <methods>`, body `{title,status:405,code:"METHOD_NOT_ALLOWED"}`. Source: `problem.go`, `httpresponse.WriteMethodNotAllowed`, `problem.CodeMethodNotAllowed`.

**Swept packages (and ONLY these):** `internal/modules/auth/delivery/http`, `internal/modules/iam/delivery/http`, `internal/platform/featureflags`, `internal/platform/observability`. `audit/delivery/http` is already canonical — **do not touch.**

---

### Task 1: auth/delivery/http (4 sites)

**Files:**
- Create: `internal/modules/auth/delivery/http/handler_method_not_allowed_test.go`
- Modify: `internal/modules/auth/delivery/http/handler.go:66,101,121,134`

- [ ] **Step 1: Write the failing test**

```go
package httpdelivery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/problem"
)

// assertMethodNotAllowedProblem asserts the canonical RFC 9457 405 contract (D-03).
func assertMethodNotAllowedProblem(t *testing.T, rec *httptest.ResponseRecorder, wantAllow string) {
	t.Helper()
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", got)
	}
	if got := rec.Header().Get("Allow"); got != wantAllow {
		t.Fatalf("Allow = %q, want %q", got, wantAllow)
	}
	var body problem.Problem
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != problem.CodeMethodNotAllowed {
		t.Fatalf("code = %q, want %q", body.Code, problem.CodeMethodNotAllowed)
	}
}

func TestAuthHandler_MethodNotAllowedIsProblemJSON(t *testing.T) {
	h := &Handler{}
	cases := []struct {
		name      string
		method    string
		target    string
		fn        func(http.ResponseWriter, *http.Request)
		wantAllow string
	}{
		{"login", http.MethodGet, "/api/v1/auth/login", h.handleLogin, "POST"},
		{"logout", http.MethodGet, "/api/v1/auth/logout", h.handleLogout, "POST"},
		{"me", http.MethodPost, "/api/v1/auth/me", h.handleMe, "GET"},
		{"change-password", http.MethodGet, "/api/v1/auth/change-password", h.handleChangePassword, "POST"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.target, nil)
			rec := httptest.NewRecorder()
			tc.fn(rec, req)
			assertMethodNotAllowedProblem(t, rec, tc.wantAllow)
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/auth/delivery/http/ -run TestAuthHandler_MethodNotAllowedIsProblemJSON -v`
Expected: FAIL — Content-Type empty (bare `WriteHeader(405)` sets no problem+json body/Allow).

- [ ] **Step 3: Write minimal implementation**

In `handler.go`, replace each of the four bare branches. `httpresponse` is already imported.

```go
// handleLogin (line ~66)
	if r.Method != http.MethodPost {
		httpresponse.WriteMethodNotAllowed(w, "POST")
		return
	}
// handleLogout (line ~101)
	if r.Method != http.MethodPost {
		httpresponse.WriteMethodNotAllowed(w, "POST")
		return
	}
// handleMe (line ~121)
	if r.Method != http.MethodGet {
		httpresponse.WriteMethodNotAllowed(w, "GET")
		return
	}
// handleChangePassword (line ~134)
	if r.Method != http.MethodPost {
		httpresponse.WriteMethodNotAllowed(w, "POST")
		return
	}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/auth/delivery/http/ -run TestAuthHandler_MethodNotAllowedIsProblemJSON -v` → PASS.
Then full package: `go test ./internal/modules/auth/delivery/http/` → PASS (no regression).

- [ ] **Step 5: Commit**

```bash
git add internal/modules/auth/delivery/http/handler.go internal/modules/auth/delivery/http/handler_method_not_allowed_test.go
git commit -m "fix(auth): bare 405 -> canonical problem+json (F1.1)"
```

---

### Task 2: iam/delivery/http (4 sites — admin ×2, sessions ×2)

**Files:**
- Create: `internal/modules/iam/delivery/http/method_not_allowed_test.go`
- Modify: `internal/modules/iam/delivery/http/admin_handler.go:149,297`, `internal/modules/iam/delivery/http/sessions_handler.go:67,143`

> Construction: `NewAdminHandler(nil, nil)` and `NewSessionsHandler(nil)` — the method guard returns before the service-nil check, so nil deps are fine. `handleUserRoleUpsert` takes `(w, r, userID)`. Confirm `httpresponse` is imported in both files (add if missing). The reuse of helper name `assertMethodNotAllowedProblem` is safe — different package from Task 1.

- [ ] **Step 1: Write the failing test**

```go
package httpdelivery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/problem"
)

func assertMethodNotAllowedProblem(t *testing.T, rec *httptest.ResponseRecorder, wantAllow string) {
	t.Helper()
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", got)
	}
	if got := rec.Header().Get("Allow"); got != wantAllow {
		t.Fatalf("Allow = %q, want %q", got, wantAllow)
	}
	var body problem.Problem
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != problem.CodeMethodNotAllowed {
		t.Fatalf("code = %q, want %q", body.Code, problem.CodeMethodNotAllowed)
	}
}

func TestIAMAdminHandler_MethodNotAllowed(t *testing.T) {
	h := NewAdminHandler(nil, nil)
	t.Run("overview", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/iam/admin/overview", nil)
		rec := httptest.NewRecorder()
		h.handleAdminOverview(rec, req)
		assertMethodNotAllowedProblem(t, rec, "GET")
	})
	t.Run("role-upsert", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/iam/users/u-1/roles", nil)
		rec := httptest.NewRecorder()
		h.handleUserRoleUpsert(rec, req, "u-1")
		assertMethodNotAllowedProblem(t, rec, "POST")
	})
}

func TestIAMSessionsHandler_MethodNotAllowed(t *testing.T) {
	h := NewSessionsHandler(nil)
	t.Run("list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sessions", nil)
		rec := httptest.NewRecorder()
		h.handleSessions(rec, req)
		assertMethodNotAllowedProblem(t, rec, "GET")
	})
	t.Run("by-id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/sessions/s-1", nil)
		rec := httptest.NewRecorder()
		h.handleSessionByID(rec, req)
		assertMethodNotAllowedProblem(t, rec, "DELETE")
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/iam/delivery/http/ -run 'TestIAM(Admin|Sessions)Handler_MethodNotAllowed' -v`
Expected: FAIL — Content-Type empty on all four subtests.

- [ ] **Step 3: Write minimal implementation**

Replace the bare branch at each site with the helper and the correct Allow value:

```go
// admin_handler.go handleAdminOverview (~149)
		httpresponse.WriteMethodNotAllowed(w, "GET")
// admin_handler.go handleUserRoleUpsert (~297)
		httpresponse.WriteMethodNotAllowed(w, "POST")
// sessions_handler.go handleSessions (~67)
		httpresponse.WriteMethodNotAllowed(w, "GET")
// sessions_handler.go handleSessionByID (~143)
		httpresponse.WriteMethodNotAllowed(w, "DELETE")
```

If `httpresponse` is not yet imported in a file, add `"metaldocs/internal/platform/httpresponse"` to its import block.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/iam/delivery/http/ -run 'TestIAM(Admin|Sessions)Handler_MethodNotAllowed' -v` → PASS.
Then full package: `go test ./internal/modules/iam/delivery/http/` → PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/modules/iam/delivery/http/admin_handler.go internal/modules/iam/delivery/http/sessions_handler.go internal/modules/iam/delivery/http/method_not_allowed_test.go
git commit -m "fix(iam): bare 405 -> canonical problem+json across admin+sessions (F1.1)"
```

---

### Task 3: platform/featureflags + platform/observability (2 sites)

**Files:**
- Create: `internal/platform/featureflags/handler_method_not_allowed_test.go`
- Modify: `internal/platform/featureflags/handler.go:33`
- Create: `internal/platform/observability/http_method_not_allowed_test.go`
- Modify: `internal/platform/observability/http.go:149`

- [ ] **Step 1: Write the failing tests**

featureflags (external test package — go through the mux):

```go
package featureflags_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/featureflags"
	"metaldocs/internal/platform/problem"
)

func TestHandler_MethodNotAllowedIsProblemJSON(t *testing.T) {
	h := featureflags.NewHandler(config.FeatureFlagsConfig{})
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/feature-flags", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", got)
	}
	if got := rr.Header().Get("Allow"); got != "GET" {
		t.Fatalf("Allow = %q, want GET", got)
	}
	var body problem.Problem
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Code != problem.CodeMethodNotAllowed {
		t.Fatalf("code = %q, want %q", body.Code, problem.CodeMethodNotAllowed)
	}
}
```

observability (external test package — exercise the returned handler):

```go
package observability_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/observability"
	"metaldocs/internal/platform/problem"
)

func TestMetricsHandler_MethodNotAllowedIsProblemJSON(t *testing.T) {
	o := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	h := o.MetricsHandler()

	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", got)
	}
	if got := rr.Header().Get("Allow"); got != "GET" {
		t.Fatalf("Allow = %q, want GET", got)
	}
	var body problem.Problem
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Code != problem.CodeMethodNotAllowed {
		t.Fatalf("code = %q, want %q", body.Code, problem.CodeMethodNotAllowed)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/platform/featureflags/ ./internal/platform/observability/ -run MethodNotAllowed -v`
Expected: FAIL — Content-Type empty on both.

- [ ] **Step 3: Write minimal implementation**

```go
// featureflags/handler.go handle (~33)
	if r.Method != http.MethodGet {
		httpresponse.WriteMethodNotAllowed(w, "GET")
		return
	}
// observability/http.go MetricsHandler (~149)
		if r.Method != http.MethodGet {
			httpresponse.WriteMethodNotAllowed(w, "GET")
			return
		}
```

Add `"metaldocs/internal/platform/httpresponse"` to the import block of each file (neither imports it yet). Verify no import cycle: `httpresponse` imports only `problem` + stdlib, so both platform packages can depend on it.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/platform/featureflags/ ./internal/platform/observability/ -run MethodNotAllowed -v` → PASS.
Then full packages: `go test ./internal/platform/featureflags/ ./internal/platform/observability/` → PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/platform/featureflags/handler.go internal/platform/featureflags/handler_method_not_allowed_test.go internal/platform/observability/http.go internal/platform/observability/http_method_not_allowed_test.go
git commit -m "fix(platform): bare 405 -> canonical problem+json in featureflags+observability (F1.1)"
```

---

### Task 4: Class-kill verification + runtime proof (acceptance gate)

**Files:** none modified (verification only).

- [ ] **Step 1: AC1 — prove the class is dead in all swept packages**

Run:
```bash
grep -rnE 'WriteHeader\(http\.StatusMethodNotAllowed\)|WriteHeader\([0-9]*405' \
  internal/modules/auth/delivery/http \
  internal/modules/iam/delivery/http \
  internal/platform/featureflags \
  internal/platform/observability | grep -v '_test.go'
```
Expected: **no output** (zero bare-405 remaining). If any line prints, fix that site and re-run.

- [ ] **Step 2: AC3 — build + regression across touched packages**

Run:
```bash
go build ./...
go test ./internal/modules/auth/... ./internal/modules/iam/... ./internal/platform/featureflags/... ./internal/platform/observability/...
```
Expected: build clean; all tests PASS (no success-path regression).

- [ ] **Step 3: AC4 — runtime proof on the live API**

Start (or confirm) the API: `.\scripts\start-api.ps1` (rebuild with `-Build` if the binary is stale). Then:
```bash
curl -i -X DELETE http://localhost:8081/api/v1/auth/login
```
Expected: `HTTP/1.1 405 Method Not Allowed`, `Content-Type: application/problem+json`, `Allow: POST`, body `{"title":"Method not allowed","status":405,"code":"METHOD_NOT_ALLOWED"}`. Capture the response for `evidence.md`.

- [ ] **Step 4: Record evidence**

Fill `../f1.1-bare-405-sweep/evidence.md` (copy `.claude/skills/milestone/templates/feature-evidence.md`): the four AC rows with real output, TDD red→green note, review disposition, and bounded defers (expected: none).

---

## Self-review (done at authoring)

- **Spec coverage:** AC1 (Task 4 grep), AC2 (Tasks 1-3 red→green tests), AC3 (Task 4 build+test), AC4 (Task 4 curl) — all four acceptance criteria from `spec.md` map to a task. ✅
- **Placeholder scan:** every code step has complete code; the only "implementer reads the guard" note is backed by the concrete Allow values in Tasks 2-3. ✅
- **Type consistency:** `assertMethodNotAllowedProblem(t, rec, wantAllow)` signature identical in Tasks 1 & 2; `problem.CodeMethodNotAllowed`, `httpresponse.WriteMethodNotAllowed(w, allow)` used consistently. ✅
- **Non-goal guard:** only the four swept packages are touched; `audit` is explicitly excluded; no success path is altered. ✅
