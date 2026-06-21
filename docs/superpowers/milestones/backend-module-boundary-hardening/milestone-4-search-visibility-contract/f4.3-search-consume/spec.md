# Feature F4.3 — Spec: search consumes the published visibility + projection contracts

> **Milestone:** 4 — search consumes published visibility contracts  ·  **Folder:** `f4.3-search-consume`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-21 / leandrotca (operator ratified the F4.3 consumer rewrite + ledger drain + baseline realignment) — *no implementation begins until this line is filled.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Source / consumer contract (already published)

F4.3 is the **consumer adoption** of the contracts F4.1 and F4.2 already published — no new producer:
- `metaldocs.v_cd_search_facts` (F4.1, 0243) — 1 row/CD projection + `is_company` + `owner_user_id`.
- `metaldocs.v_cd_grantee` (F4.1, 0243) — bounded restricted-CD visibility edges.
- `metaldocs.v_document_search_facts` (F4.2, 0244) — 1:1 documents projection (14 cols).

The consumer is `internal/modules/search/infrastructure/v2documents/reader.go` `ListDocuments`. Today its
query reads five foreign base tables raw (the C4 ledger): `public.documents` (C4a), `public.controlled_documents`
(C4b), `controlled_document_area_grants` (C4c), `controlled_document_user_grants` (C4e), `user_process_areas`
(C4d). F4.3 repoints every one to the published views.

## What this feature implements

1. **Rewrite the `reader.go` `ListDocuments` SQL** to read only the three `metaldocs` views:
   - `FROM metaldocs.v_document_search_facts d`
   - `LEFT JOIN metaldocs.v_cd_search_facts cd ON cd.controlled_document_id = d.controlled_document_id AND cd.tenant_id = d.tenant_id`
   - Projection unchanged (same 13 output columns, same `COALESCE`s — the views expose the same column names).
   - Visibility predicate becomes (removing the `'company'`/`'restricted'` literals, Category-A coupling):
     ```
     (d.controlled_document_id IS NULL AND d.created_by = $13)
     OR (cd.controlled_document_id IS NOT NULL AND (
            cd.is_company
         OR cd.owner_user_id = $13
         OR EXISTS (SELECT 1 FROM metaldocs.v_cd_grantee g
                     WHERE g.tenant_id = cd.tenant_id
                       AND g.controlled_document_id = cd.controlled_document_id
                       AND g.grantee_user_id = $13)
     ))
     ```
   - All other filters (`archived_at IS NULL`, text/status/profile/family/area/department/owner/expiry),
     `ORDER BY d.created_at DESC, d.id DESC`, `LIMIT/OFFSET`, the `$1..$14` param order, and the Go family-port
     resolution path are **unchanged** (seam only). Comments updated to describe view consumption.
2. **Drain the H-G ledger:** remove all five C4 entries from `hgPendingRemediation` in
   `tools/cilint/internal/analyzers/hgcrossmodule.go` (the slice becomes **empty** — M4 is the last debt
   milestone; mission.md §8 requires an empty ledger at terminal acceptance), with a closing comment recording
   the C4 drain.
3. **Realign the ledger unit test:** `TestHGCrossModule_Negative_PendingBaseline` points at the now-drained
   C4d site; flip it to the **empty-ledger end-state** — assert the formerly-pending site (search reading
   `user_process_areas`) now **flags** (proving the ledger no longer suppresses it). The suppression mechanism
   (`hgListed`) stays covered by `TestHGCrossModule_Negative_Exempt`.

## Non-goals (mandatory)

- **No authz/visibility change** — the composed view decision is set-identical to the pre-M4 inline predicate
  (proven at the SQL level in F4.1 and end-to-end here). Seam only.
- **No new view, no producer change** — F4.1/F4.2 views are consumed as-is.
- **No change to the `searchdomain` port, the application service, the HTTP handler, or the family port.**
- **No param-order / pagination / ordering change** — `$1..$14` and `ORDER BY`/`LIMIT`/`OFFSET` are preserved.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| D6 raw==contract parity: the rewritten reader returns the SAME document set/order as a verbatim copy of the pre-M4 raw query, per actor, incl. **revoked-member + ungranted-user discriminators** | new `reader_contract_parity_integration_test.go` (runs frozen raw query vs `reader.ListDocuments`; owner/areaMember/userGrant/revokedMem/none) | real (PG :5434) |
| Existing behavioral guard still GREEN | `go test -tags integration -run TestListDocuments_EnforcesUnifiedVisibility ./internal/modules/search/infrastructure/v2documents/` | real (PG :5434) |
| No `'company'`/`'restricted'` string literal remains in `reader.go` | `grep -n "'company'\|'restricted'" reader.go` → none | real |
| H-G guard green with empty ledger; no raw C4 read remains | `go run ./tools/cilint ./...` exit 0; `grep` shows no `FROM public.documents`/`controlled_documents`/grant tables in reader.go | real |
| Ledger unit suite green with realigned baseline | `go test ./tools/cilint/...` | real |
| Build + full suite | `go build ./...`; `go test ./...` | real |

> TDD: write the D6 parity test first (it FAILs only if the rewrite drifts); rewrite the reader to green it;
> then drain the ledger and realign the baseline test (run `go test ./tools/cilint/...`). **HS-3:** if test PG
> :5434 is down, the integration steps are **not-run**, never false-green.

## ADR needed?

- [x] No *new* ADR — F4.3 consumes the already-Accepted ADR-0039 D3a views; the literal removal is the
  Category-A coupling cleanup the mission scoped. No new decision.
