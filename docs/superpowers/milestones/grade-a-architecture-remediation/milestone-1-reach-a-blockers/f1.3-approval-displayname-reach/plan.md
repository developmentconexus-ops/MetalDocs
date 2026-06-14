# Feature F1.3 — approval signoff display-name reach (contain off-tx) — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: superpowers:subagent-driven-development.
> Steps use checkbox (`- [ ]`) syntax. Contract read from source; see `spec.md`.

**Goal:** Move the signoff display-name read (`SELECT display_name FROM metaldocs.iam_users`) out of
the lock-holding signoff transaction into a **contained** off-tx method on the approval module's own
`ApprovalRepository`, preserving the persisted `ActorDisplayNameSnapshot` value exactly (H-PRE-1).

**Architecture:** Add `LoadActorDisplayName(ctx, tenantID, userID) (string, error)` to
`ApprovalRepository`; implement on `postgresApprovalRepository` using the pool (`r.db`, off-tx),
mirroring the existing off-tx `ListRoutes` precedent. In `RecordSignoff`, call it **pre-flight**
(before `runner.Do`), capture the string into the tx closure, and delete the inline `tx.QueryRowContext`
read. Align the three RecordSignoff-driving test fakes; add a positive threading-proof test.

**Tech Stack:** Go `net/http`, `database/sql`, Postgres (RLS NULL-permissive per migration 0237),
`go test`.

**Why off-tx is safe (read, don't re-derive):** `metaldocs.iam_users` RLS `tenant_isolation` policy is
NULL-permissive (`GUC unset/empty → rows visible`, `db/migrations/0237_rls_all_tenant_tables.sql:11-14,
113-117`); the pool connection sets no `metaldocs.tenant_id` GUC, and the explicit `tenant_id = $2::uuid`
predicate keeps the read tenant-correct — identical row to today's in-tx read.

**Scope guard (HS-6):** contained method on `ApprovalRepository` ONLY. **No** shared cross-module
`UserDisplayNameReader` port — that is M4/F4.1.

---

### Task 1: Contain the signoff display-name read off-tx (TDD)

**Files:**
- Modify: `internal/modules/documents/approval/repository/approval_repository.go` (interface, ~line 90)
- Modify: `internal/modules/documents/approval/repository/postgres_approval_repository.go` (new impl)
- Modify: `internal/modules/documents/approval/application/decision_service.go` (hoist call, remove inline read)
- Modify (test): `internal/modules/documents/approval/application/decision_service_test.go` (fake field+override, remove dead branch, new test)
- Modify (test): `internal/modules/documents/approval/application/coverage_boost_test.go` (fake override)
- Modify (test): `internal/modules/documents/approval/application/phase5_integration_test.go` (fake override)

- [ ] **Step 1: Add the interface method.** In `approval_repository.go`, inside the
  `ApprovalRepository` interface, after `LoadActiveDocumentContentHash` (the read-helper group,
  ~line 90), add:

```go
	// LoadActorDisplayName returns metaldocs.iam_users.display_name for (tenantID,
	// userID), or "" when the user row is absent. It runs OFF the caller's
	// transaction (on the pool) so it never executes inside the signoff
	// advisory-lock atomic tx (H-PRE-1). Tenant scope is the explicit tenant_id
	// predicate; the metaldocs.tenant_id RLS GUC is unset on the pool connection,
	// which the NULL-permissive tenant_isolation policy (migration 0237) allows.
	LoadActorDisplayName(ctx context.Context, tenantID, userID string) (string, error)
```

- [ ] **Step 2: Implement on the postgres repo.** In `postgres_approval_repository.go`, add (mirror the
  off-tx `ListRoutes` precedent at `:421` — uses `r.db`, not `tx`):

```go
// LoadActorDisplayName reads the approver's display name off the connection pool
// (NOT inside the caller's signoff transaction) so the cross-module iam_users read
// is never held inside the signoff advisory-lock tx (H-PRE-1). Empty string when
// the user is absent (best-effort snapshot, matches the prior inline behavior).
func (r *postgresApprovalRepository) LoadActorDisplayName(ctx context.Context, tenantID, userID string) (string, error) {
	var displayName sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT display_name
		  FROM metaldocs.iam_users
		 WHERE user_id = $1
		   AND tenant_id = $2::uuid`,
		userID, tenantID,
	).Scan(&displayName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", MapPgError(err, MapHints{})
	}
	return displayName.String, nil
}
```

- [ ] **Step 3: Build the production packages** (interface assertion forces the impl):

Run: `go build ./internal/modules/documents/approval/...`
Expected: exit 0. (The compile-time `var _ ApprovalRepository = (*postgresApprovalRepository)(nil)` now
requires the new method; if it errors, the impl signature is wrong.)

- [ ] **Step 4: Add the fake field + override to `fakeDecisionRepo`.** In `decision_service_test.go`,
  add a field to the `fakeDecisionRepo` struct (after `instanceStatusFrom`, ~line 41):

```go
	actorDisplayName   string
```

  and add the override method (next to the other `fakeDecisionRepo` methods, e.g. after
  `LoadActiveDocumentContentHash`):

```go
func (r *fakeDecisionRepo) LoadActorDisplayName(_ context.Context, _, _ string) (string, error) {
	return r.actorDisplayName, nil
}
```

- [ ] **Step 5: Add overrides to the other two RecordSignoff-driving fakes** (both use the nil no-op
  embed and call `RecordSignoff`, so they need an explicit override to avoid a nil-method panic — they
  do not assert on the display name, so return `""`).

  In `coverage_boost_test.go`, next to the other `fakeDecisionRepoWithCounter` methods (~`:2965`):

```go
func (r *fakeDecisionRepoWithCounter) LoadActorDisplayName(_ context.Context, _, _ string) (string, error) {
	return "", nil
}
```

  In `phase5_integration_test.go`, next to the other `phase5Repo` methods:

```go
func (r *phase5Repo) LoadActorDisplayName(_ context.Context, _, _ string) (string, error) {
	return "", nil
}
```

- [ ] **Step 6: Write the failing threading-proof test.** In `decision_service_test.go`, add (clone of
  `TestRecordSignoff_ContentHashEchoesInstanceSubmitHash:622`, but set `actorDisplayName` and assert the
  snapshot):

```go
func TestRecordSignoff_ThreadsActorDisplayNameFromRepo(t *testing.T) {
	const (
		instanceID = "inst-dn"
		stageID    = "stage-dn"
		actorID    = "approver-dn"
		authorID   = "author-dn"
	)

	inst := buildTwoApproverInstance(instanceID, stageID, authorID, []string{actorID, "approver-2"})
	conn := &decisionTestConn{
		authzGranted: true,
		areaCode:     "QA",
		actorID:      actorID,
	}
	repo := &fakeDecisionRepo{instance: inst, actorDisplayName: "Alice Approver"}
	svc := &DecisionService{
		repo:          repo,
		emitter:       &MemoryEmitter{},
		clock:         fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)},
		freezeInvoker: &fakeFreezeInvoker{},
	}
	db := newDecisionTestDB(t, conn)

	_, err := svc.RecordSignoff(context.Background(), newTxRunner(db), SignoffRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "approve",
		SignaturePayload: map[string]any{},
		ContentFormData:  map[string]any{"_content_hash": inst.ContentHashAtSubmit},
	})
	if err != nil {
		t.Fatalf("RecordSignoff: unexpected error: %v", err)
	}
	if repo.insertedSignoff == nil {
		t.Fatal("expected inserted signoff")
	}
	if got := repo.insertedSignoff.ActorDisplayNameSnapshot(); got != "Alice Approver" {
		t.Errorf("ActorDisplayNameSnapshot() = %q; want %q (must be threaded from repo.LoadActorDisplayName)", got, "Alice Approver")
	}
}
```

Run: `go test ./internal/modules/documents/approval/application/ -run TestRecordSignoff_ThreadsActorDisplayNameFromRepo -v`
Expected: **FAIL** — `ActorDisplayNameSnapshot() = ""; want "Alice Approver"`, because `RecordSignoff`
still reads the inline `tx.QueryRowContext` (the in-memory driver returns `nil` for that query at
`decision_service_test.go:358-359`), not the repo override. This is the RED that proves the hoist is
wired by Step 7.

- [ ] **Step 7: Hoist the read; delete the inline tx read.** In `decision_service.go` `RecordSignoff`:

  (a) Immediately **before** `err := runner.Do(ctx, func(tx *sql.Tx) error {` (`:158`), and after the
  `var result SignoffResult` / `var eligibilityEvent *GovernanceEvent` declarations (`:156-157`), insert:

```go
	// H-PRE-1: resolve the actor display-name snapshot OFF the signoff transaction.
	// This is a cross-module read of metaldocs.iam_users; running it inside the
	// lock-holding signoff tx (advisory lock + FOR UPDATE stage rows) on a fresh
	// connection risks deadlock. req.TenantID/req.ActorUserID are server-derived and
	// available pre-flight. Contained on ApprovalRepository (not a shared port — M4/F4.1).
	actorDisplayName, err := s.repo.LoadActorDisplayName(ctx, req.TenantID, req.ActorUserID)
	if err != nil {
		return SignoffResult{}, fmt.Errorf("recordSignoff: lookup actor display name: %w", err)
	}
```

  (b) Change the existing `err := runner.Do(...)` on `:158` to `err = runner.Do(...)` (the `err`
  variable is now already declared by (a) — `:=` would be a re-declaration/compile error).

  (c) **Delete** the inline read block (`:263-274` region): remove the `var actorDisplayName sql.NullString`
  declaration and the entire `if err := tx.QueryRowContext(ctx, ` … `recordSignoff: lookup actor display
  name: %w", err) }` statement. Keep the surrounding `// Step 7: build the domain Signoff …` comment.

  (d) At the `domain.NewSignoff(domain.SignoffParams{…})` call (`:289`), change
  `ActorDisplayNameSnapshot: actorDisplayName.String,` to `ActorDisplayNameSnapshot: actorDisplayName,`
  (the captured pre-flight string; the closure captures the outer `actorDisplayName` like it captures
  `result`/`eligibilityEvent`).

Run: `go test ./internal/modules/documents/approval/application/ -run TestRecordSignoff_ThreadsActorDisplayNameFromRepo -v`
Expected: **PASS**.

- [ ] **Step 8: Remove the now-dead in-memory driver interception** (orphan of this change). In
  `decision_service_test.go`, delete the branch at `:357-360`:

```go
	// IAM display name lookup for actor_display_name_snapshot.
	if strings.Contains(q, "from metaldocs.iam_users") && strings.Contains(q, "display_name") {
		return &decisionSingleValueRows{value: nil}, nil // NULL — best-effort
	}
```

  The signoff display-name read no longer flows through the tx driver (it is the repo override now), so
  this branch is unreachable. If `strings` becomes unused in the file after deletion, leave it — it is
  used elsewhere in the same file (other `strings.Contains` branches remain).

- [ ] **Step 9: Confirm no `iam_users` read remains inside the tx closure** (AC1):

Run: `grep -n 'iam_users' internal/modules/documents/approval/application/decision_service.go`
Expected: **no output** (the only signoff-path `iam_users` reference is gone; the read is now a
`s.repo.LoadActorDisplayName` call before `runner.Do`). If any match remains inside the closure, the
inline read was not fully removed — fix before continuing.

- [ ] **Step 10: Full verification** (AC4):

Run: `go build ./...`  → exit 0.
Run: `go vet ./internal/modules/documents/approval/...`  → clean.
Run: `go test ./internal/modules/documents/approval/...`  → **all pass** (no nil-method panic from any
fake; existing signoff tests unchanged in behavior since empty-on-missing is preserved).

- [ ] **Step 11: Commit**

```bash
git add internal/modules/documents/approval/repository/approval_repository.go \
        internal/modules/documents/approval/repository/postgres_approval_repository.go \
        internal/modules/documents/approval/application/decision_service.go \
        internal/modules/documents/approval/application/decision_service_test.go \
        internal/modules/documents/approval/application/coverage_boost_test.go \
        internal/modules/documents/approval/application/phase5_integration_test.go
git commit -m "$(cat <<'EOF'
refactor(approval): contain signoff display-name read off-tx (H-PRE-1)

Move the raw SELECT display_name FROM metaldocs.iam_users out of the
lock-holding signoff transaction into a contained off-tx ApprovalRepository
method (LoadActorDisplayName), resolved pre-flight before runner.Do. The
cross-module reach is unchanged (still iam_users) and intentionally NOT
generalized to a shared port (that is M4/F4.1). Persisted
ActorDisplayNameSnapshot value is preserved exactly; empty-on-missing kept.

Off-tx pool read is safe: iam_users RLS tenant_isolation is NULL-permissive
(GUC unset -> rows visible, migration 0237) and the explicit tenant_id
predicate keeps it tenant-correct.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 2 (controller-run): runtime 101-equivalent proof + evidence (AC5)

> Not an implementer subagent task — the controller runs this after Task 1's two-stage review is clean,
> exactly as F1.2's AC5 runtime proof was controller-run.

- [ ] **Step 1: AC5 — live signoff proof.** `.\scripts\start-api.ps1 -Build`; login
  (`POST /api/v1/auth/login`, `{"identifier":"admin","password":"AdminMetalDocs123!"}`); drive a real
  approval to a signoff (submit → signoff) via the API; then read back
  `metaldocs.approval_signoffs.actor_display_name_snapshot` for that signoff and confirm it equals the
  approver's `metaldocs.iam_users.display_name`. Confirm the signoff request returns promptly (no lock
  wait / deadlock). Capture the value + timing.

- [ ] **Step 2: AC1 structural proof.** Show `grep -n 'iam_users' decision_service.go` is empty and the
  `LoadActorDisplayName` call precedes `runner.Do` (line numbers).

- [ ] **Step 3: Write `evidence.md`** — ACs table (AC1–AC5 with real proofs), review disposition
  (spec-compliance then code-quality, both passes), bounded defers (e.g. M4/F4.1 shared-port
  generalization explicitly deferred; decision-record fold-in deferred to `wiki-curator` at M1 close).

---

## Self-Review (run before dispatching)

- **Spec coverage:** AC1 → Steps 7,9 + Task 2 Step 2. AC2 → Steps 1-2 (contained on repo, `r.db`, no
  shared port). AC3 → Steps 6-7 (threading test). AC4 → Step 10. AC5 → Task 2 Step 1. All covered.
- **No placeholders:** every step has exact code/command + expected output.
- **Type consistency:** method `LoadActorDisplayName(ctx, tenantID, userID) (string, error)` is identical
  across interface (Step 1), impl (Step 2), and all three fake overrides (Steps 4-5); query arg order is
  `userID, tenantID` → `$1=user_id, $2=tenant_id`, matching the original inline read. Getter
  `ActorDisplayNameSnapshot()` and param field `ActorDisplayNameSnapshot` match `domain/signoff.go`.
- **TDD:** Step 6 RED before Step 7 GREEN.
- **Non-goals guarded:** no shared port (HS-6); no other `iam_users` read touched; value unchanged;
  tx/lock semantics unchanged apart from removing the one inline read.
