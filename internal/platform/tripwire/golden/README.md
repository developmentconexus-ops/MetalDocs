# Tripwire golden render

`0319_capability_bindings_tripwire.sql` in this directory is a committed golden
artifact that pins `internal/platform/tripwire.RenderMigration()` byte-for-byte, so a
hand-edit of the generated SQL or a stale regeneration is caught by two blocking gates:

- `scripts/api-lint` rule `TRIPWIRE-ARM-PARITY` (`tripwire_arm_rules.go`, run with
  `-strict` by `.github/workflows/api-contract.yml`)
- `internal/platform/tripwire.TestRenderMigration_MatchesCommittedFile`

Regenerate it with `go run ./cmd/gen-tripwire` (no argument → this path).

## Why it lives here and not only in `db/migrations/`

Through 0301, this golden pinned a migration that had been folded into
`db/baseline/0001_current_schema.sql` (the 2026-07-29 fold squashed migrations
0257–0315 and archived the files under
`archive/migrations/post-baseline-2026-07-fold/`, a conceptually immutable historical
record) — so the regenerable golden lived only here, with no live counterpart in
`db/migrations/`.

0319 (issue #89/A8.1, ADR 0092 D1) is past that fold boundary: it is BOTH a real,
unarchived forward migration at `db/migrations/0319_capability_bindings_tripwire.sql`
AND pinned here byte-identically as the golden CI reads. The two copies must stay in
sync — regenerating one without the other is exactly the drift TRIPWIRE-ARM-PARITY
exists to catch.

## Changing the tripwire vocabulary

Editing `arms.go` (or `render.go`) is a **schema change**, and this file is not applied
to any database. Any future tripwire vocabulary change ships as a NEW forward migration
in `db/migrations/` — pick the next free 4-digit version, regenerate the SQL into it
(`go run ./cmd/gen-tripwire db/migrations/<NNNN>_<slug>.sql`), and then advance this
golden plus the three path references in lockstep:

- `cmd/gen-tripwire/main.go` (`defaultRelPath` — keeps pointing into this directory,
  only the basename advances)
- `scripts/api-lint/tripwire_arm_rules.go` (`tripwireGoldenPath`)
- `internal/platform/tripwire/arms_test.go` (`goldenRelPath`)

`render.go`'s `migrationHeader` carries the filename, so the golden basename and the
forward migration basename must agree.
