# F4.5 — authz soft-GUC NULL hardening (root-cause fix behind the F4.2 defer)

> **Origin:** F4.2 live-green audit (operator-directed: "auditoria completa … não deixar defer
> nem gap em M4, mesmo um pouco fora do escopo"). Closes `task_e03a4383` inside M4 rather than
> deferring it to M8.
> **Approved for code: 2026-07-04** (operator instruction to close the gap, not defer).
> **Boundary note:** touches `internal/modules/iam/authz` (one module out from documents/M4). In
> scope by explicit operator direction; the change is additive robustness, fail-closed unchanged.

## Consumer contract

- **Consumer = every caller of `authz.MustActorID` / `authz.MustTenantID`** (the in-tx identity
  guards behind `authz.Require`), plus the soft readers `softActorID`/`softTenantID` and
  `loadAssertedCaps`.
- **Required behavior:** when the transaction-local placeholder GUC (`metaldocs.actor_id` /
  `metaldocs.tenant_id` / `metaldocs.asserted_caps`) was **never set** on the connection,
  `current_setting(name, true)` returns SQL **NULL** (PostgreSQL semantics for a never-introduced
  custom GUC — *not* empty string). The reader MUST treat NULL and `''` identically as "unset":
  - `MustActorID` / `MustTenantID` → return `ErrActorContextMissing` / `ErrTenantContextMissing`
    (the documented sentinel), **never** a raw driver error
    (`converting NULL to string is unsupported`).
  - `softActorID` / `softTenantID` → return the caller's default (`"system"` / `""`).
  - `loadAssertedCaps` → empty asserted-caps set.
- **Fail-closed invariant (security):** the change is **stricter-or-equal**. The accept path (GUC
  seeded to a non-empty value) is byte-for-byte unchanged; only the reject path swaps an opaque
  driver error for the correct semantic sentinel. No guard is weakened; no identity is fabricated.
- **Emitted SQL unchanged:** the `current_setting(...)` statements stay byte-identical (constant
  literals passed through), so the existing sqlmock statement-shape assertions keep matching.

## Non-goals

- NOT changing `authz.Require`, tier-1/tier-2 semantics, the DB write-tripwire, or RLS policies.
- NOT changing `SeedTxIdentity` / `SeedTxTenant` (the write side is already correct).
- NOT altering the soft path's fail-soft leniency (any read error → default) — only routing it
  through the shared NULL-correct reader.
- NOT a broad iam/authz refactor — exactly the four soft-GUC read sites, unified onto one helper.

## Validation gate

- **TDD pin:** new `TestMustActorID_ReturnsErrWhenGUCNull` / `TestMustTenantID_ReturnsErrWhenGUCNull`
  (sqlmock `AddRow(nil)` = SQL NULL) assert the sentinel. Before the fix these fail with the driver
  error; after, they pass. Existing empty-string + happy-path + seed tests stay green.
- `go build ./...` green; `go test ./internal/modules/iam/authz/ -count=1` green (unit).
- `go test ./internal/modules/iam/authz/ -tags integration -count=1` green on real Postgres (RLS
  backstop + effective-from + bypass exercise cold-connection GUC reads).
- **Downstream proof:** with F4.5 in place, F4.2's `TestPublishRace` runs **live green** on real
  Postgres (all 4 subtests) — the concrete consumer that the NULL-GUC crash previously blocked.

## Interview record

| Q | Answer |
|---|---|
| Fix now in M4, or defer to M8? | **Now, in M4** (operator: close the gap, mesmo fora de escopo). |
| Symptom-patch the two Must* or fix the class? | **Fix the class** — one shared `readSoftGUC` (`sql.NullString`) backs all four readers; kills the three-way split (`loadAssertedCaps` already had the right idiom). ADR 0022 = root-cause, never symptom-patch. |
| Any behavior change on the accept path? | **None.** Stricter-or-equal: reject path only, driver-error → sentinel. |
