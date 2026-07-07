# F10 — ADRs + Wiki Gates + Live QA (milestone close)

## Scope

Final feature of M2b `approval-kernel-backend`, before milestone-validator dispatch.
Not a code feature — a closing/verification feature. Three parts:

1. **ADR completeness + index sync** — confirm ADR 0074 (route versioning, F2),
   0075 (approval.review/oversee capabilities, F3), 0076 (freeze boundary +
   choke-point concurrency, F5), 0077 (delegation, F9) exist, are structurally
   complete (Context/Decision/Consequences/Alternatives/Rollback/References),
   and are indexed in `wiki/decisions/index.md`.
2. **Wiki sync** — `wiki/modules/approval.md` (or equivalent) reflects
   `approval.review`/`approval.oversee` (F3), `changes_requested` status (F4),
   freeze boundary (F5), no-fallback hash chain (F6), unified SoD (F7),
   SLA/visibility/worklist (F8), delegation (F9).
3. **Cross-feature verification sweep + live QA** — confirm zero regression
   across F1-F9's combined changes via build/test/lint, then exercise the
   approval lifecycle over real HTTP against a locally running API + Postgres
   (`.\scripts\start-api.ps1`), per the milestone's "compile ≠ work" mandate.

## Out of scope

- New production features or capability changes.
- Rewriting the `route_admin_service_test.go` fake-driver harness to model
  real Postgres transaction-abort semantics (a disproportionate new
  investment; noted as an observation, not fixed).
- Adding a `stage_kind` field to the HTTP `StageRequest` contract (would be a
  real F-something code feature, not a closing/verification task — the gap is
  documented as a bounded defer, consistent with F4's own evidence.md).
- Any raw SQL **write** against the shared Postgres DB outside the sanctioned
  API path (explicitly out of bounds per this session's tool-permission
  denial precedent — see evidence.md).

## Locked constraints re-checked (system-impact analysis)

Every hard constraint from
`docs/superpowers/analysis/2026-07-07-approval-remediation-m2b-system-impact.md`
must have a closing reference by the end of F10: implemented+verified here,
or explicitly bounded-deferred with a named trigger. See evidence.md
"Self-review" section for the full checklist.
