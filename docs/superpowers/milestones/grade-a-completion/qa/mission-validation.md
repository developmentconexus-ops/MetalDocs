# Mission Terminal Acceptance — Verdict (2026-06-20 post-M8 re-run)

> Written by: the mission-validator subagent (separation of powers). Validates against: ../mission.md §8
> (incl. the §8 scope amendment / M8-widened H-D scope) + the post-M8 re-audit report.
> Run: 2026-06-20 · Code HEAD: `58dea742` (verified via `git rev-parse HEAD`) · Verdict: see bottom.
> Judges: `wiki/backend/_artifacts/architecture-re-audit-2026-06-20-post-m8.md`.

## Per-criterion results

| # | §8 criterion | Method run (command/agent) | Real evidence | Pass? |
|---|--------------|----------------------------|---------------|-------|
| 1a | module-boundaries ≥ A− | Read `search/.../v2documents/reader.go`; grep `document_profiles` in search | reader.go:19/27 inject `taxonomydomain.FamilyCodeResolver`; :143 `ProfileCodesForFamily`, :224 `ResolveFamilyCodes`; **0** raw `document_profiles` SQL in search. Report grades **A−**. | ✅ |
| 1b | composition ≥ A− | grep `type MetricsResponse` in `observability/http.go` | http.go:183 `type MetricsResponse struct` (typed envelope, F8.2). Report grades **A−**. | ✅ |
| 1c | contract-api ≥ A− | Read handler.go:624/674/1105/1122/1159/1193; api.gen.go; openapi.yaml | 3 skeptic-confirmed Majors live; generated types exist & unused; OpenAPI diverges. Report grades **B+**. | ❌ |
| 2 | 0 skeptic-confirmed NEW Critical/Major | Independent re-read of all 3 cited Major sites | (1) handler.go:674 `WriteJSON(201, map[string]string{document_id,initial_revision_id,session_id})` vs `DocumentCreateResult` (api.gen.go:208) + openapi.yaml:2576 `$ref DocumentCreateResult`. (2) handler.go:1105 `WriteJSON(200, map[string]string{"url":url})` vs openapi.yaml:2904 declares only `302`/no body. (3) handler.go:1122/1159/1193 emit local `commentResponse` (:1217, `id string`, `content json.RawMessage`) vs `DocumentCommentResponse` (api.gen.go:188, `Id openapi_types.UUID`, `Content []DocumentCommentContentNode`) + openapi.yaml:2930/2956/2988. **3 confirmed Majors.** | ❌ |
| 3 | H-D = 0 at M8-widened scope | `GOFLAGS=-mod=mod go run ./tools/cilint ./...` (rc=0); §5b Part-A grep; read `noresponsemap.go:172 isMapStringAnyLiteral` | cilint exits **0**; Part A only the 2 `health.go:24,33` probe exemptions. **BUT** `isMapStringAnyLiteral` matches value type `any`/`interface{}` only (returns false for `*ast.Ident{Name:"string"}`), so `map[string]string` is unguarded. Live `map[string]string` response literals at handler.go:624, 674, 1105 on spec-declared routes. **Mechanical gate = 0; H-D by §8 intent = 3.** | ❌ |
| 4 | H-G = 0 at full cross-module scope | `grep document_profiles` (widened, -_test, -taxonomy); `grep FROM iam_users\|iam_user_roles` outside iam | `document_profiles`: only a comment (`approval/.../route_admin_service.go:23`) + seed in `internal/test/e2e_seed.go` — no SQL reads. iam tables outside `iam/`: **0** (rc=1). **Honest H-G = 0.** | ✅ |
| build | whole-repo build green | `GOFLAGS=-mod=mod go build ./...` | rc=0 (clean) | ✅ |
| test | whole-repo `go test ./...` green, no regression | `GOFLAGS=-mod=mod go test ./...` | rc=0, zero FAIL lines (full unit suite green) | ✅ |

Confirmed-fix spot-checks (all real, not fixture): F8.1 presence (`map[string]any` only as a comment at presence/handler.go:90 — allowlisted), F8.2 metrics typed struct (http.go:183), F8.3 search FamilyCodeResolver port (reader.go:19/143/224, 0 raw SQL).

## Pass bar
- Bar (§8): "a fresh, independent re-run of the F5.1 10-dimension re-audit at the post-M4 HEAD passes the §6 bar — (1) module-boundaries, contract-api, and composition all ≥ A−; (2) 0 skeptic-confirmed new Critical/Major; (3) H-D = 0; (4) H-G = 0." (H-D/H-G measured at the M8-widened scope.)
- Met? **No.** 2 of 4 checks pass (composition + module-boundaries within Check 1; Check 4 H-G). Contract/API is B+ (< A−); 3 skeptic-confirmed Majors survive; H-D by §8 intent = 3 (`map[string]string` literals the `map[string]any`-only analyzer cannot see). All independently reproduced. The report's self-FAIL is correct and honest.

## Forbidden-list (any hit = FAIL)
- [ ] Fixture/mock passed off as real-provider proof — no; build/test/cilint/greps all re-run from clean state at HEAD `58dea742`.
- [ ] A criterion marked pass without a command actually run — no; every row cites a command + real output.
- [x] Split-brain / guessed contract surfaced — **yes, and it is the binding failure**: handler emits untyped bodies diverging from generated types + OpenAPI on 3 public routes (Check 1c/2/3).
- [ ] Self-judged / validator edited or fixed code — no; this validator wrote only this verdict file and edited no source.

## Verdict
- VERDICT: **FAIL**
- Failed criteria: **1c (contract-api < A−), 2 (3 confirmed Majors), 3 (H-D = 3 by §8 intent)**.
- This is Contract/API's **5th consecutive miss** (B+ → B− → B → B+ → B+). Per mission HS-5 5th-miss directive, the bounded remediation does NOT auto-open; the decision is the operator's. The bounded micro-milestone that would clear these criteria (HS-5): (a) route `duplicateDocument` (674) and the comment endpoints (1122/1159/1193) through `DocumentCreateResult`/`DocumentCommentResponse`; correct `signedRevisionURL` (1105) to the spec-declared 302 (or amend OpenAPI to the implemented 200+body); regen BE/FE codegen; and (b) widen `noresponsemap` to flag any `map[string]<T>` response literal so this exact type-scope evasion cannot recur. Option B (full `StrictServerInterface` typed-response rewire) is the true root-cause close per the report §8.
- The mission stays **open**. The main session does not declare the mission done and does not flip mission/program/roadmap status. Grade-A terminal sign-off is not reached at HEAD `58dea742`.
