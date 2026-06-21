# Feature F4.3 — Evidence

> **Milestone:** 4 — search consumes published visibility contracts  ·  **Feature:** `f4.3-search-consume`  ·  **Closed:** 2026-06-21
> **Contract:** `spec.md` (operator-approved 2026-06-21). search consumes the three published views (F4.1/F4.2); all five raw C4 base-table reads removed; H-G debt ledger drained to empty; baseline test realigned.

## What was implemented

- **EDIT** `internal/modules/search/infrastructure/v2documents/reader.go` — `ListDocuments` rewritten to read
  ONLY published contracts (ADR-0039 D3a):
  - `FROM metaldocs.v_document_search_facts d` (was `public.documents`)
  - `LEFT JOIN metaldocs.v_cd_search_facts cd ON cd.controlled_document_id = d.controlled_document_id AND cd.tenant_id = d.tenant_id` (was `public.controlled_documents`)
  - Visibility predicate composed from the views: `(d.controlled_document_id IS NULL AND d.created_by = $13) OR
    (cd.controlled_document_id IS NOT NULL AND (cd.is_company OR cd.owner_user_id = $13 OR EXISTS(SELECT 1 FROM
    metaldocs.v_cd_grantee g WHERE g.tenant_id = cd.tenant_id AND g.controlled_document_id =
    cd.controlled_document_id AND g.grantee_user_id = $13)))` — removes the inline reads of
    `controlled_document_area_grants`, `controlled_document_user_grants`, `user_process_areas` and the
    `'company'`/`'restricted'` scope-enum literals (`is_company` replaces them).
  - Projection columns, all filters, `$1..$14` param order, `ORDER BY d.created_at DESC, d.id DESC`,
    `LIMIT/OFFSET`, and the Go family-port resolution path are **unchanged** (seam only). Block comments updated
    to describe view consumption.
- **NEW** `internal/modules/search/infrastructure/v2documents/reader_contract_parity_integration_test.go` — the
  D6 raw==contract parity gate: freezes a verbatim copy of the pre-M4 raw query and asserts the rewritten reader
  returns the SAME ordered doc-id set per actor, incl. the revoked-member + ungranted-user discriminators.
- **EDIT** `internal/modules/search/infrastructure/v2documents/reader_test.go` — the two sqlmock contract guards
  (`TestListDocumentsFiltersByTenantID`, `TestListDocumentsBindsActorForVisibility`) realigned: the expected
  query regexp now asserts `v_cd_grantee ... LIMIT $11 OFFSET $12` (the published-contract visibility leg) +
  actor bound at $13. Contract/invariant guard repair, not a behavioral change.
- **EDIT** `tools/cilint/internal/analyzers/hgcrossmodule.go` — **drained** all five C4 entries from
  `hgPendingRemediation`; the slice is now **EMPTY** (mission.md §8 terminal acceptance), with a closing comment
  recording the C4 drain.
- **EDIT** `tools/cilint/internal/analyzers/hgcrossmodule_test.go` — `TestHGCrossModule_Negative_PendingBaseline`
  → `TestHGCrossModule_LedgerDrained_EmptyAtMissionEnd`: flips to the empty-ledger end-state (the formerly-pending
  C4d site now FLAGS). The suppression mechanism (`hgListed`) stays covered by `TestHGCrossModule_Negative_Exempt`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| D6 raw==contract parity (owner/areaMember/userGrant/revokedMem/none) | `go test -tags integration -run TestListDocuments_ContractParityWithFrozenRaw ./internal/modules/search/infrastructure/v2documents/` | PASS before AND after rewrite (raw==contract both ways) | real (PG :5434) |
| Behavioral visibility guard still GREEN | `…-run TestListDocuments_EnforcesUnifiedVisibility …` | PASS | real (PG :5434) |
| Full search module suite (incl. family + sqlmock unit guards) | `go test -tags integration ./internal/modules/search/...` | `ok` all packages | real (PG :5434) |
| No `'company'`/`'restricted'` literal, no raw C4 base-table read in reader.go | `grep` for literals + `FROM/JOIN public.{documents,controlled_documents,grant tables,user_process_areas}` | NONE; only `metaldocs.v_*` views referenced | real |
| H-G guard green, ledger empty, no raw C4 read remains | `go run ./tools/cilint ./...` | `cilint-exit=0` | real |
| Ledger unit suite green with realigned baseline | `go test ./tools/cilint/...` | `ok …/internal/analyzers 5.128s` | real |
| Static build | `go build ./...` | `BUILD-OK` (exit 0) | — |
| `go vet` (search + cilint) | `go vet ./internal/modules/search/... ./tools/cilint/...` | clean | — |
| Full unit suite | `go test ./...` | exit 0 (no failures) | real |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| D6 raw==contract parity incl. revoked + ungranted | yes | `TestListDocuments_ContractParityWithFrozenRaw` PASS |
| Existing behavioral guard still GREEN | yes | `TestListDocuments_EnforcesUnifiedVisibility` PASS |
| No `'company'`/`'restricted'` literal remains in reader.go | yes | grep NONE |
| H-G guard green with empty ledger; no raw C4 read | yes | `cilint-exit=0`; grep NONE |
| Ledger unit suite green with realigned baseline | yes | `go test ./tools/cilint/...` ok |
| Build + full suite | yes | `go build ./...` ok; `go test ./...` exit 0 |

## Review disposition

- Spec-compliance review: **PASS** — consumer reads only the three published views; scope-enum literals removed;
  param order / ordering / pagination / family-port path unchanged (seam only); ledger drained to empty; baseline
  realigned to the end-state; no port/service/handler/producer change.
- Code-quality review: **PASS** — the rewrite is a minimal seam proven behavior-preserving by a frozen-raw parity
  test over 5 actor scopes incl. the revoked-member anti-drift discriminator; sqlmock guards realigned to assert
  the new contract leg + actor binding; ledger drain documented; the realigned guard asserts the drained
  end-state (formerly-pending site now flags).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| 3 integration tests fail on the raw base DSN (`TestSequenceAllocatorNextAndIncrement_Concurrent`, `TestPostgresLimiter_Live`, `TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema`) with `relation … does not exist` | **Pre-existing, not M4** — confirmed via stash-and-rerun: all three fail IDENTICALLY with the F4.3 changes stashed. Root cause is environmental (these tests connect to a raw base DSN whose schema is only materialized inside testdb per-test clones), in modules F4.3 never touched (CD/documents). | Test-environment fix owned outside this mission; re-point these live-schema tests at the testdb template clone. Out of M4 boundary. |
