# F4.5 evidence — authz soft-GUC NULL hardening

> **Contract:** `./spec.md`. Closes `task_e03a4383` inside M4 (root cause behind the F4.2 defer).
> **Outcome:** one shared NULL-correct GUC reader; fail-closed unchanged (stricter-or-equal);
> F4.2 now runs **live green** on real Postgres.

## Root cause (audit conclusion)

`current_setting('metaldocs.actor_id', true)` returns SQL **NULL** — not `''` — when the placeholder
GUC was never introduced on the connection (cold pooled connection, no `set_config`). Four readers of
these soft GUCs existed with **three** behaviors:

| Reader | Old scan | NULL behavior (the bug) |
|---|---|---|
| `MustActorID` (context.go) | plain `string` | **driver error** `converting NULL to string is unsupported` instead of `ErrActorContextMissing` |
| `MustTenantID` (context.go) | plain `string` | same, for tenant |
| `softGUC` (bypass_audit.go) | plain `string` | masked — any error → default (fail-soft, but swallowed the class) |
| `loadAssertedCaps` (authz.go) | `sql.NullString` | **correct** — the idiom the others should have used |

Production masks the crash because the request lifecycle always seeds identity before `authz.Require`;
the NULL path only appears on a never-seeded connection (exactly what a cold integration test hits).
Still a real defect: the guard failed with the **wrong** (opaque) error, degrading defense-in-depth.

## What shipped

One shared reader in `context.go`, all four sites unified onto it:

```go
func readSoftGUC(ctx context.Context, tx *sql.Tx, query string) (string, error) {
    var v sql.NullString
    if err := tx.QueryRowContext(ctx, query).Scan(&v); err != nil {
        return "", err
    }
    if !v.Valid { return "", nil } // NULL == unset == ""
    return v.String, nil
}
```

- `MustActorID` / `MustTenantID`: `readSoftGUC` → `if v=="" → sentinel`. Emitted SQL byte-identical.
- `softGUC`: `readSoftGUC`, keep fail-soft `if err!=nil || v=="" → def`.
- `loadAssertedCaps`: `readSoftGUC` (dropped its local `sql.NullString`).

**Stricter-or-equal:** accept path (non-empty GUC) unchanged; reject path only — driver crash replaced
by the documented sentinel. No guard weakened, no identity fabricated. Root-cause per ADR 0022 (no
symptom-patch): the divergent-readers *class* is collapsed to one helper, not the two symptoms alone.

## Commands (real output)

```
$ go build ./...                                                        → BUILD OK

$ go test ./internal/modules/iam/authz/ -run 'Null' -count=1 -v
=== RUN   TestMustActorID_ReturnsErrWhenGUCNull
--- PASS: TestMustActorID_ReturnsErrWhenGUCNull (0.00s)
=== RUN   TestMustTenantID_ReturnsErrWhenGUCNull
--- PASS: TestMustTenantID_ReturnsErrWhenGUCNull (0.00s)
PASS
ok  metaldocs/internal/modules/iam/authz  1.344s

$ go test ./internal/modules/iam/authz/ -count=1                       → ok (unit, 1.487s)

$ METALDOCS_DATABASE_URL=…  go test ./internal/modules/iam/authz/ -tags integration \
      -run 'SeedTxTenant|RLS|EffectiveFrom|Bypass|Require' -count=1   → ok (113.989s, real Postgres)
```

**TDD proof:** the two NULL pins fail against pre-fix code with
`converting NULL to string is unsupported` (the driver error), and pass after the `sql.NullString`
rewire — the exact behavior contract the fix installs.

**Downstream proof (the point of the fix):** F4.2 `TestPublishRace -tags integration` now runs
**live green** on real Postgres, all 4 subtests — see
[`../f4.2-publish-race/evidence.md`](../f4.2-publish-race/evidence.md).

## Review / QA disposition

Independent reviewer pass on the aggregate authz diff (security-sensitive identity path): confirmed
accept-path parity, fail-closed preserved, SQL byte-identical, no boundary breach
(`internal/platform/db` still does not import `iam`). Disposition: **clean**.

## Bounded defers

None. `task_e03a4383` — dismissed (resolved here, not deferred).
