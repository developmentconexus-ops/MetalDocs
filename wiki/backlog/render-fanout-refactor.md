# Refactor Backlog - render-fanout

**Last verified:** 2026-07-02 (APP-07 — R-002 closed)

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Promote render-fanout module page from stub to full living-doc shape | T-001 | M | major | - | - | open | - |
| R-002 | Document and test outbox retry/finalization contract end-to-end | T-002 | M | major | - | - | closed 2026-07-02 | - |
| R-003 | Record resolver registry/version compatibility policy in ADR | T-003 | S | minor | - | - | open | - |

## R-002 closure evidence (APP-07)

- Contract documented in `wiki/backend/flows/async-job-pipeline.md` §7 "Staging outbox tables — retry/terminal contract (APP-07)": state machine, claim→process→mark choreography, backoff formula + config source, `ResetStaleClaims` ownership, dead-letter visibility, idempotency expectations, ADR 0054 tenancy rules.
- Tests added in `internal/modules/render/fanout/pdf_outbox_repository_test.go`: `TestPDFOutboxRepository_MarkFailed_RetryPath_ResetsClaimAndBumpsAttempts`, `TestPDFOutboxRepository_MarkFailed_FinalizePath_SetsDeadLetteredAndFailedStatus`, `TestPDFOutboxRepository_MarkFailed_RowNotFound`, `TestPDFOutboxRepository_ResetStaleClaims_OnlyTouchesProcessingOlderThanCutoff`, `TestPDFOutboxRepository_ClaimPending_RespectsAttemptsAndRetryGate`, `TestPDFOutboxRepository_ClaimPending_NoEligibleRowsReturnsEmpty`.
- `go build/vet/test -count=1 ./internal/modules/render/...` green; `gofmt -l` clean.
