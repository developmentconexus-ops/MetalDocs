# Sync log - search

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

---

### 2026-06-10 — Stage-1 backend audit drift patch

**Mode:** lite patch (stale evidence anchors + aspirational qualifier)

**Changes applied:**

- `search-tech-debt.md` T-001: Retitled item to reflect the actual remaining gap (405 response emits no body). Removed stale evidence reference to non-existent `writeAPIError` helper and stale surface anchor `handler.go:141`. New surface anchor `handler.go:54-56`. Noted that all other error paths were migrated to `httpresponse.WriteError` in commit c4c4d95d2; the 405 path was not. Severity downgraded from major to minor (the legacy envelope is gone; only the bodyless 405 remains).
- `search.md` line 52 (formerly): Replaced aspirational qualifier "shared with other surfaces that may populate them via different paths" with a factual statement — no such surface exists in the current codebase; `Subject`, `BusinessUnit`, `Classification`, `Tags` are always zero values in v2 search results.
- `search.md` + `search-tech-debt.md`: Bumped "Last verified" to 2026-06-10.

**Skipped:** No structural changes. No backlog or ADR updates (T-001 severity change is within the existing item boundary).
