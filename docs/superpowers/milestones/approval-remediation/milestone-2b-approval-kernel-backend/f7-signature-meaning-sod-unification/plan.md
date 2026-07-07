# Feature F7 — `signature-meaning-sod-unification` — plan.md

Ref: `spec.md` (this folder), plan.md `docs/superpowers/plans/2026-07-07-m2b-approval-kernel-backend.md` F7 section, design spec.md §5/§2.1 (W7/W8, 21 CFR 11.50(a)(3)).

## Task 1 — SoD unification: failing test first

- [ ] Grep-confirm today's exact duplication: `grep -rn "SubmittedBy ==" --include=*.go internal/modules/documents/approval` shows exactly one inline hit in `review_verdict_service.go` (decision_service.go already calls `domain.CheckSoD`).
- [ ] Extend `internal/modules/documents/approval/application/review_verdict_service_test.go` (or the existing integration test file) with a test asserting the self-verdict rejection still returns `domain.ErrAuthorCannotSign` — this should already pass pre-change (regression pin) so the subsequent refactor is provably behavior-preserving.

## Task 2 — Implement: call `domain.CheckSoD` from `review_verdict_service.go`

- [ ] Edit `internal/modules/documents/approval/application/review_verdict_service.go`: replace the inline `if instance.SubmittedBy == req.ActorUserID { return domain.ErrAuthorCannotSign }` block with `if err := domain.CheckSoD(instance.SubmittedBy, req.ActorUserID, nil); err != nil { return err }`. Update the surrounding comment to state this now shares the exact predicate `decision_service.go` calls (remove the old comment's "a second table-shape-specific duplicate helper here would be redundant" framing — this IS now the shared helper).
- [ ] Edit `internal/modules/documents/approval/domain/sod.go`: update `CheckSoD`'s doc comment to state explicitly it is called from both `RecordSignoff` (approval-stage signoffs) and `RecordVerdict` (review-stage verdicts), and that a `nil` `priorSignoffs` is the correct call shape for callers with no same-shape prior-record source.
- [ ] Run `go test ./internal/modules/documents/approval/...` — confirm the Task 1 regression pin and all existing SoD/review-verdict tests still pass.
- [ ] Grep-zero: `grep -rn "SubmittedBy == req.ActorUserID" --include=*.go .` → zero hits outside `domain/sod.go` itself (the comparison expression naturally lives inside `CheckSoD`'s own body under a different variable-name spelling — confirm by reading the file, not just grepping the literal string).

## Task 3 — DB-level SoD trigger parity for `approval_review_verdicts`

- [ ] Read `enforce_signoff_sod()`'s exact current body (`db/baseline/0001_current_schema.sql:682-701`) to mirror verbatim.
- [ ] Write failing integration test first: direct SQL `INSERT INTO approval_review_verdicts (...) VALUES (... actor_user_id = <author's user id> ...)` against a testdb-factory-seeded instance → expect a Postgres error (`P0001`/`check_violation`) — this must FAIL before the trigger exists.
- [ ] Write `db/migrations/0290_sod_unified_trigger.sql`: create (or re-use, if the function can be made table-agnostic without an ALTER on the existing trigger) a single `enforce_approval_sod()`-style function keyed off `NEW.approval_instance_id` (a column both tables already have), and attach it via `CREATE TRIGGER ... BEFORE INSERT ON approval_review_verdicts FOR EACH ROW EXECUTE FUNCTION ...`. Decide at implementation time (document the decision in evidence.md) whether to also re-point `approval_signoffs`'s existing `trg_signoff_sod` at the shared function (DRY, zero behavior change, same error text) or leave it calling `enforce_signoff_sod()` untouched and only add the new table's trigger against a differently-named but identical-body function — prefer the shared-function approach unless the existing trigger's exact name is relied upon elsewhere (grep `trg_signoff_sod`/`enforce_signoff_sod` for non-migration references first).
- [ ] Run the failing test from this task against a describe-only SQL review (cannot execute live — see Bounded defers) — confirm the migration SQL is syntactically correct and mirrors the existing trigger's idiom (same `RAISE EXCEPTION ... USING ERRCODE = 'check_violation'` shape, same `SET search_path` hardening).
- [ ] Ledger row: `INSERT INTO public.schema_migrations (version, description) VALUES ('0290', ...) ON CONFLICT (version) DO NOTHING;`.

## Task 4 — Signature meaning: failing test first

- [ ] Extend `internal/modules/documents/approval/application/decision_service_test.go` (or `decision_service_reauth_test.go`, wherever `RecordSignoff`'s domain-object construction is most directly testable) with two failing tests: `TestRecordSignoff_ApproveSetsSignatureMeaningApproval` and `TestRecordSignoff_RejectSetsSignatureMeaningRejection` — assert the persisted/constructed `Signoff.SignatureMeaning()` matches the decision, using the fake repo's captured `InsertSignoff` argument (existing fakes already capture the inserted `domain.Signoff` — confirm the capture point, extend if it currently discards the value).
- [ ] Run — expect FAIL (`SignatureMeaning()` returns default `"approval"` even for a reject decision today).

## Task 5 — Implement: derive `SignatureMeaning` from `Decision`

- [ ] Edit `internal/modules/documents/approval/application/decision_service.go`: immediately before the `domain.NewSignoff(domain.SignoffParams{...})` call (around line 308), compute `signatureMeaning := "approval"; if req.Decision == domain.DecisionReject { signatureMeaning = "rejection" }` (or a small helper function `signatureMeaningForDecision(req.Decision) string` if that reads cleaner given the file's existing style — check for a similar small mapping helper already in the file, e.g. `signoffOutcome`-style, and match that idiom). Add `SignatureMeaning: signatureMeaning` to the `SignoffParams` literal.
- [ ] Run Task 4's tests — expect PASS.
- [ ] Run full `decision_service_test.go` suite — confirm no regression (existing tests that don't assert on `SignatureMeaning` should be unaffected since the field defaults identically for approve).

## Task 6 — Expose `signature_meaning` in the read-path manifest

- [ ] Edit `internal/modules/documents/approval/http/contracts/instance_read.go`: add `SignatureMeaning string \`json:"signature_meaning"\`` to `SignoffRecord`.
- [ ] Edit `internal/modules/documents/approval/http/get_instance_handler.go`'s `mapInstanceResponse`: add `SignatureMeaning: sig.SignatureMeaning()` to the `contracts.SignoffRecord{}` literal built in the per-stage signoff loop.
- [ ] Write/extend a failing test first (e.g. `get_instance_handler_test.go`) asserting the JSON response's `stages[].signoffs[].signature_meaning` field is present and matches a fixture signoff's meaning, for both an approve and a reject fixture — then implement to green.

## Task 7 — Contract (openapi) + regen

- [ ] Edit `api/openapi/v1/openapi.yaml`'s `ApprovalSignoffRecordResponse` schema: add `signature_meaning: { type: string, enum: [approval, rejection] }` (not in `required`, since older rows before this feature's migration might theoretically read back a value — but migration 0286 already defaults every row, so consider marking required if the read path is confirmed to always populate it; decide and document at implementation time).
- [ ] Run the repo's oapi-codegen regen make target for the approval module; confirm `internal/modules/documents/approval/api/api.gen.go`'s generated `ApprovalSignoffRecordResponse` Go struct gains the new field with no other diff.
- [ ] `go build ./...` clean (confirm the regen doesn't break any existing generated-type consumer — expected zero consumers per spec.md Interview #6, but verify).

## Task 8 — Full verification sweep

- [ ] `go build ./...`
- [ ] `go build -tags integration ./...`
- [ ] `go test -count=1 ./internal/modules/documents/approval/...`
- [ ] `go test -count=1 ./...` (grep zero FAIL)
- [ ] `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` (0 violations)
- [ ] `go test ./scripts/api-lint/...`
- [ ] `grep -rn "SubmittedBy == req.ActorUserID" --include=*.go .` → zero hits outside `domain/sod.go`

## Task 9 — Self-review pass

- [ ] Grep-zero confirm ALL duplicate inline SoD call sites are gone, not just the one named in spec.md (re-scan the whole `approval` module for any OTHER `== req.ActorUserID` / `== instance.SubmittedBy` pattern that might be a second, previously-unnoticed duplicate).
- [ ] Confirm no column/field is reused across two unrelated semantic purposes (F4's caught-bug class) — specifically check the new SoD trigger function doesn't accidentally read/write a column shared with an unrelated concept.
- [ ] Confirm no silently swallowed error path in the new migration's trigger or the `signatureMeaning` derivation (every branch either sets a real value or returns a real error — no logged-and-continue).

## Task 10 — Evidence + commit

- [ ] Write `evidence.md` (implementation summary, verification table, judgment calls, bounded defers, grep-zero confirmation, scope confirmation).
- [ ] `git add` explicit touched paths only (never `-A` — repo has unrelated pending deletions/untracked files).
- [ ] Commit: `feat(approval): F7 signature meaning + unified SoD predicate (W7/W8, 21 CFR 11.50(a)(3))`.
