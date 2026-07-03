# ADR 0061 — Area hierarchy shape: self-FK parent pointer, no live cycle-prevention check (gap acknowledged)

- **Status:** Accepted (records current shape; flags an open gap — see Consequences)
- **Last verified:** 2026-07-02
- **Date:** 2026-07-02
- **Scope:** Records the existing shape of the taxonomy area hierarchy (`metaldocs.document_process_areas.parent_code`, a self-referencing FK) and documents, as verified runtime truth, that **no cycle-prevention check is currently wired** despite tech-debt row T-016's premise that one exists. Closes tech-debt T-016 (`wiki/modules/taxonomy-tech-debt.md`) as "ADR written," but reopens a corrected, more severe finding — see Consequences.
- **Depends on:** none.

---

## Context

Tech-debt row T-016 (`wiki/modules/taxonomy-tech-debt.md:146-152`) as briefed described the area hierarchy as "self-FK + application-layer acyclicity check," citing `area_service.go`'s `SetParent` invoking `ListAncestors` to reject cycles, and framed the gap as purely "undocumented as an ADR." Verifying this against current code surfaced a **different and more serious finding**: the described acyclicity check does not exist in the current codebase.

### Verified runtime facts

- **Schema shape: self-referencing FK, no cycle guard.** `archive/migrations/0123_taxonomy_extend_process_areas.sql:31-44` adds `fk_area_parent_tenant FOREIGN KEY (tenant_id, parent_code) REFERENCES metaldocs.document_process_areas (tenant_id, code)` — present in the baseline at `db/baseline/0001_current_schema.sql:4029-4033`. This FK enforces that a `parent_code` must reference an *existing* area row in the same tenant; it does not and cannot prevent a cycle (A→B→A), which is a graph-reachability property, not a referential-integrity property. No CHECK constraint, no trigger, and no recursive-CTE-backed constraint on `document_process_areas` enforces acyclicity at the DB layer — confirmed by grepping the baseline's trigger list for this table (`trg_process_areas_code_immutable` at `:3778-3781`, rejecting `code` mutation; `trg_require_cap_asserted` at `:3799-3802`, the authz tripwire — neither addresses `parent_code` cycles).
- **`ListAncestors` exists but `SetParent` does not, and no caller invokes `ListAncestors` before a parent-code write.** `internal/modules/taxonomy/domain/port.go:27-28` declares `ListAncestors`/`ListAncestorsTx` on the `AreaRepository` port; `internal/modules/taxonomy/infrastructure/repository.go:623-658,695-737` implements both via a `WITH RECURSIVE ancestors AS (...)` CTE walk. **There is no `SetParent` function anywhere in `internal/modules/taxonomy`.** The actual mutation path is `AreaService.Update` (`internal/modules/taxonomy/application/area_service.go:84-131`): it loads the existing row (`GetByCodeForUpdate`, line 99), overwrites `existing.ParentCode = normalized.ParentCode` unconditionally (line 105), and persists via `UpdateTx` (line 109) — **no call to `ListAncestors`, `ListAncestorsTx`, or any other reachability check appears in this function or anywhere else in `area_service.go`.**
- **A cycle-specific error type exists and is wired to an HTTP response, but is never raised.** `internal/modules/taxonomy/domain/area.go:24` declares `ErrAreaParentCycle = errors.New("area parent assignment creates cycle")`; `internal/modules/taxonomy/delivery/http/routes_areas.go:166-167` maps it to `400 AREA_PARENT_CYCLE`. Grepping the entire `internal/modules/taxonomy` tree for `ErrAreaParentCycle` returns exactly these two lines — the error is defined and has a response mapping, but **no code path returns it.** `area_service_test.go` has no cycle-rejection test case.
- **Net effect:** today, `PUT`-ing area `A`'s `parent_code` to `B` while `B.parent_code` is already (transitively) `A` will succeed at both the application and database layer — no self-loop or transitive-loop guard exists.

## Decision

**Record the shape as it is, not as the tech-debt row assumed it to be: the area hierarchy is a self-referencing-FK parent pointer with referential integrity only (parent must exist, same tenant) and no cycle-prevention enforcement anywhere in the stack.** This ADR does not implement a fix (out of scope for a decisions-only pass) — it corrects the record so the next implementer does not reuse a false premise.

The FK shape itself (self-FK on `(tenant_id, parent_code) → (tenant_id, code)`) is retained as designed and is not in question — it is the correct minimal shape for "each area optionally has one parent in the same tenant." What this ADR declines to affirm is the claim that acyclicity is enforced; it explicitly is not.

For any future fix, the two live options remain open and undecided by this ADR (a genuine follow-up decision, not resolved here):
- **Option A — application-layer check:** wire `ListAncestors`/`ListAncestorsTx` into the actual parent-mutation path (there is no `SetParent`; the real target is `AreaService.Update`, or a new dedicated method) and return `ErrAreaParentCycle` (already defined, already wired to HTTP 400) when the new parent's ancestor chain includes the area being updated.
- **Option B — DB-layer check:** a Postgres function/trigger (e.g. `assert_no_cycle`) invoked `BEFORE UPDATE` on `parent_code`, or a closure-table redesign. Higher cost, stronger guarantee (defends even non-application writers).

## Consequences

- T-016 (`wiki/modules/taxonomy-tech-debt.md`) is closed as "ADR written" in the narrow sense the register asked for, but the finding underneath it is upgraded: this is not a missing-ADR-for-an-enforced-rule (minor, per the register's own severity rubric) but a **missing enforcement entirely** for a rule the domain layer already believes exists (`ErrAreaParentCycle` + its HTTP mapping are dead code today, giving a false impression of protection to anyone reading `routes_areas.go` in isolation). The task brief's hard rule to flag runtime-truth contradictions applies here: **the tech-debt register's characterization of this item was inaccurate** — it described enforced-but-undocumented; the actual state is unenforced-and-partially-scaffolded.
- This is being recorded, not fixed, per this task's decisions-only scope. A follow-up backlog/tech-debt item should be opened (separate from this ADR) to either wire Option A (cheapest — the plumbing already exists: `ListAncestors`, `ErrAreaParentCycle`, and the 400 mapping are all present, just not connected) or select Option B.
- No migration, schema change, or code change is made by this ADR itself.

## References

- `wiki/modules/taxonomy-tech-debt.md` T-016 — tech-debt row this ADR closes (with the correction noted above).
- `internal/modules/taxonomy/application/area_service.go:84-131` — `AreaService.Update`, the actual (unguarded) parent-mutation path.
- `internal/modules/taxonomy/domain/area.go:24` — `ErrAreaParentCycle`, defined but never raised.
- `internal/modules/taxonomy/domain/port.go:27-28`, `internal/modules/taxonomy/infrastructure/repository.go:623-658,695-737` — `ListAncestors`/`ListAncestorsTx`, defined but never called from a mutation path.
- `internal/modules/taxonomy/delivery/http/routes_areas.go:166-167` — dead HTTP mapping for `ErrAreaParentCycle`.
- `archive/migrations/0123_taxonomy_extend_process_areas.sql:31-44`; `db/baseline/0001_current_schema.sql:4029-4033,3778-3802` — self-FK + the two triggers that exist on this table (neither addresses cycles).
