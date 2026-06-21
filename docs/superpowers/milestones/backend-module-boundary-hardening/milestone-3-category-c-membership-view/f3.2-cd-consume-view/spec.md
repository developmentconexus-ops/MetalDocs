# Feature F3.2 — Spec: CD consumes `v_active_user_areas` (C1 + C2)

> **Milestone:** 3 — Category C  ·  **Folder:** `f3.2-cd-consume-view`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-21 / leandrotca (operator) — via operator-approved mission §7 M3 + ADR-0039 D3(a). Consumer call sites read this session; the change is a contained membership-leg swap. *No implementation begins until this line is filled — it is.*

> The feature's **contract**, approved **before any code**. The milestone-validator judges F3.2 against this file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | What exactly changes in CD's two visibility queries? | Only the **foreign membership leg** of the restricted-visibility EXISTS: `FROM user_process_areas upa … AND upa.effective_to IS NULL` becomes `FROM metaldocs.v_active_user_areas upa …` (the view encodes `effective_to IS NULL`, so that clause is dropped). The CD-**owned** legs — `controlled_document_area_grants`, `controlled_document_user_grants`, the `controlled_documents` base scan — are **unchanged**. |
| 2 | Both sites or one? | Both, identically: `List` (`repository.go:150`, the `ListControlledDocuments` keyset query) and `CanRead` (`:492`). They share the same membership-leg shape. |
| 3 | Is set-based SQL preserved (no N+1)? | Yes — the membership leg stays an `EXISTS` subquery inside the one list/visibility query. No per-row Go membership call is introduced. |
| 4 | What proves no authz drift? | A per-site parity test: `repo.CanRead` and `repo.List` (view form, post-repoint) return **exactly** what verbatim inline copies of the deleted raw `user_process_areas` SQL return, across company / restricted-area-grant / restricted-area-grant-but-**revoked**-membership / restricted-user-grant / owner / no-access scopes. The revoked-membership row is the discriminator (active view must exclude it, exactly as `effective_to IS NULL` did). |
| 5 | Any guard/ledger change? | Drop the `{controlleddocuments/infrastructure/repository.go, user_process_areas}` entry from `hgPendingRemediation` (C1+C2) once both raw reads are gone. Then **realign `TestHGCrossModule_Negative_PendingBaseline`** (currently points at this exact site) to a still-pending C4 `search/.../reader.go` row and re-green `go test ./tools/cilint/...`. |

## Consumer contract (FIRST)

- **Consumer(s):** CD's own `List` (`repository.go:150`) and `CanRead` (`:492`) — the two restricted-visibility membership legs. These are the *consumers* of iam's published view (F3.1).
- **Contract:** read membership from **`metaldocs.v_active_user_areas`** correlating `upa.tenant_id = cd.tenant_id AND upa.user_id = $actor AND upa.area_code = cdag.area_code`. Membership = row present in the view (active-now). No `effective_to` predicate in CD's SQL anymore.
- **Source of truth:** the two call sites (read this session) + F3.1's published view + ADR-0039 D3(a).

## What this feature implements

Repoint both CD restricted-visibility membership EXISTS legs from `user_process_areas` to
`metaldocs.v_active_user_areas`, dropping the now-redundant `effective_to IS NULL` clause. Update the
`upa.effective_to IS NULL` anchor comments at `:96-97` / `:479-480` to reflect that the predicate now lives
in the view. Drain the C1+C2 ledger entry; realign + re-green the cilint unit suite.

## Non-goals (mandatory)

- No change to the CD-owned grant-table legs or the `controlled_documents` base scan / pagination / filters.
- No change to `List`/`CanRead` Go signatures, scanning, or surrounding logic — only the membership-leg SQL.
- No Go-side membership loop (set-based EXISTS preserved).
- No touch to C3 (approval, F3.3) or C4 (search, M4); no touch to the `metaldocs.user_process_areas` passthrough view.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| `CanRead` view-form == raw-form across all scopes incl. revoked membership | `TestCanRead_ViewParityWithRaw` in `internal/modules/controlleddocuments/infrastructure/membership_view_parity_integration_test.go` | real (PG :5434) |
| `List` view-form visible-id set == raw-form set across all scopes incl. revoked membership | `TestList_ViewParityWithRaw` (same file) | real (PG :5434) |
| Parity green **before** raw deleted (D6) | run order in `evidence.md`: parity green, then raw SQL removed in same feature, re-green | real |
| Build + full test | `go build ./...`; `go test ./...` | real |
| Guard exit 0, C1+C2 entry drained | `go run ./tools/cilint ./...` (exit 0); `hgPendingRemediation` no longer lists `controlleddocuments/infrastructure/repository.go × user_process_areas` | real |
| `git grep user_process_areas` under `controlleddocuments/` → 0 | `git grep -n user_process_areas -- internal/modules/controlleddocuments/` (excluding the parity test's verbatim baseline) | real |
| Cilint unit suite green after fixture realign | `go test ./tools/cilint/...` | real |

> TDD: parity test seeded + asserted; run **green pre-repoint** (raw==raw baseline sanity) then **green
> post-repoint** (view==raw) — the revoked-membership case is the drift discriminator. Real PG only; HS-3 if down.

## ADR needed?

- [x] No — uses F3.1's view under ADR-0039 D3(a). No new durable decision.
