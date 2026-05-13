# Sync Log Entry Template

Append one block to `wiki/modules/<m>/_artifacts/sync-log.md` per sync run. If the file does not exist, create it with this header:

```markdown
# Sync log - <m>

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.
```

Prepend each entry directly below the header.

## Entry Format

```markdown
## <YYYY-MM-DD> - <one-line change context>

- **Context:** <git range | plan task | explicit file list | uncommitted diff>
- **Mode:** <lite patch | structural refresh>
- **Anchors moved:** <N; list or "none">
- **Public surface:** <symbols/ports/handlers/codegen changed or "none">
- **Routes/API:** <runtime/spec/codegen facts updated or "none">
- **Runtime flows:** <flow sections/artifacts updated or "none">
- **Persistence:** <tables/migrations/triggers/GUC facts updated or "none">
- **Dependencies:** <IN/OUT edges or wiring facts updated or "none">
- **T-NNN touched:** <ids + status/evidence or "none">
- **R-NNN touched:** <ids + status/evidence or "none">
- **Counts after:** Critical=<n> Major=<n> Minor=<n>; missing-ADR=<n or n/a>
- **Tally gate:** <PASS | FAIL pre-existing | FAIL fixed after rerun>
- **Patched files:** <wiki file list>
```

## Notes

- Keep this log factual and compact.
- Do not use it as a commit message replacement.
- Record pre-existing tally failures separately from failures caused by the sync patch.
