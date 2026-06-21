# Feature F4.3 — search consumes the published visibility + projection contracts

> **Milestone:** 4 — search consumes published visibility contracts  ·  **Folder:** `f4.3-search-consume`
> **Status:** Closed (2026-06-21 — D6 parity + behavioral guard GREEN; ledger drained to empty; baseline realigned; evidence.md written)

## Source

- `spec.md` (operator-approved 2026-06-21). ADR-0039 D3a/D4; F4.1 (0243) + F4.2 (0244) views.

## Plan (D6 parity-before-delete)

**TDD order:**
1. **Capture + RED** — write `internal/modules/search/infrastructure/v2documents/reader_contract_parity_integration_test.go`
   (`//go:build integration`, package `v2documents`). Embed a **frozen verbatim copy** of the pre-M4 raw
   `ListDocuments` query (base tables). Seed the visibility scenario (reuse the F4.1/visibility-test shape:
   company CD doc, restricted CD doc, standalone doc; owner/areaMember/userGrant/none) **plus a revoked-member
   discriminator** (a `user_process_areas` row with `effective_to` set). For each actor, collect the ordered
   doc-id list from (a) the frozen raw query via `db.Query` and (b) `reader.ListDocuments`, and assert equal.
   Run BEFORE the rewrite → the test passes against the current raw reader (establishes the baseline equals
   itself); it is the regression guard that will FAIL if the rewrite drifts. (D6 "raw==contract before delete":
   the frozen raw query is the permanent baseline; the reader becomes the contract path.)
2. **GREEN (rewrite)** — repoint `reader.go` `ListDocuments` to the three views (FROM v_document_search_facts,
   LEFT JOIN v_cd_search_facts, EXISTS v_cd_grantee), remove the `'company'`/`'restricted'` literals, keep the
   projection / filters / params / ORDER / LIMIT / family-port path identical. Update the block comments to
   describe view consumption. Rerun the parity test + `TestListDocuments_EnforcesUnifiedVisibility` → PASS.
3. **Drain ledger** — remove the 5 C4 `hgSite` entries from `hgPendingRemediation` (slice becomes empty) with a
   closing "C4 drained (M4/F4.3)" comment. Run `go run ./tools/cilint ./...` (exit 0 — no raw read remains).
4. **Realign baseline test** — flip `TestHGCrossModule_Negative_PendingBaseline` to the empty-ledger end-state:
   the formerly-pending C4d fixture now FLAGS. Rename to reflect end-state. Run `go test ./tools/cilint/...`.

**Files touched:**
- EDIT `internal/modules/search/infrastructure/v2documents/reader.go` (query + comments)
- NEW `internal/modules/search/infrastructure/v2documents/reader_contract_parity_integration_test.go`
- EDIT `tools/cilint/internal/analyzers/hgcrossmodule.go` (drain 5 C4 entries)
- EDIT `tools/cilint/internal/analyzers/hgcrossmodule_test.go` (realign baseline test)

**Test strategy:** integration parity + behavioral guard on real PG :5434 (`127.0.0.1:5434`, `-tags integration`);
cilint unit suite + guard run for the ledger. Full `go build ./...` + `go test ./...`.

**Non-goals (guardrails):** no authz change; no new view; no port/service/handler change; no param-order or
ordering/pagination change.

## Execution notes

- The frozen raw query in the parity test is copied verbatim from `reader.go:54-121` BEFORE the rewrite so the
  baseline is preserved even after the production query changes.
- `cd.id IS NOT NULL` → `cd.controlled_document_id IS NOT NULL` (the facts view's PK column); equivalent on a
  LEFT JOIN miss (all view cols NULL → standalone leg).
- Pre-existing bounded defer (NOT M4): `TestSequenceAllocatorNextAndIncrement_Concurrent` may fail on the raw
  base DSN — confirm pre-existing via stash-and-rerun if seen during `go test ./...`.
