# Feature F1.1 — resolution.go typed status constants

> **Milestone:** 1 — Category A: typed status constants  ·  **Folder:** `f1.1-resolution-constants`
> **Status:** Implementing

## Source

- Milestone spec row: **What to implement —** `resolution.go:42,55,58` reference the templates-owned
  `domain.VersionStatusPublished`/`VersionStatusObsolete` constants instead of bare literals; pure
  refactor, no behavior/signature/struct change. **What to validate —** build exit 0; existing
  `resolution_test.go` (9) green unchanged; 0 bare-literal hits in `resolution.go`; cilint exit 0.
- Governing-spec reference: mission §5 row 3, §7 M1 F1.1; ADR-0039 worked-classification row 3; ADR 0030.

## Plan

1. **Baseline (done, captured):** `go build ./...`=0; `go test … -run TestResolve -v`=9 PASS;
   `go run ./tools/cilint ./...`=0; bare literals present at resolution.go:42,55,58.
2. **Regression-guard test (added before edit):** append `TestResolve_UsesTemplatesVocabulary` to
   `resolution_test.go` — import `templatesdomain`, feed `string(VersionStatusPublished)` →
   resolves (override + default), `string(VersionStatusObsolete)` → `ErrDefaultObsolete`. Run: green
   at baseline too (values equal), so this is a guard, not RED-first. Confirms the test compiles with
   the new import path before the production edit.
3. **Edit `resolution.go`:** add import `templatesdomain "metaldocs/internal/modules/templates/domain"`;
   replace the 3 literals:
   - `:42` `*candidate.Status != "published"` → `!= string(templatesdomain.VersionStatusPublished)`
   - `:55` `*candidate.Status == "obsolete"` → `== string(templatesdomain.VersionStatusObsolete)`
   - `:58` `*candidate.Status != "published"` → `!= string(templatesdomain.VersionStatusPublished)`
4. **Verify (green-after):** rerun all four proof commands from the Validation Gate; assert 9/9
   resolution tests + the new guard pass, 0 bare literals, build=0, cilint=0.
5. **Evidence:** record real command output (red/green, grep, build, cilint) in `evidence.md`.

Files touched: `internal/modules/controlleddocuments/domain/resolution.go` (production),
`internal/modules/controlleddocuments/domain/resolution_test.go` (regression guard). No others.

Test strategy: characterization suite is the parity lock (unchanged, must stay green); one added
boundary regression guard wires the owner's constants as inputs. No mocks, no SQL, no provider.

## Execution notes

- Model: main session (mechanical, single-file). No subagent dispatch needed for a 3-line swap.
- `string(...)` conversion kept (field stays `*string`) — retyping is M2/HS-2, see spec non-goals.
