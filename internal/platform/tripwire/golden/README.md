# Tripwire golden render

`0301_tripwire_template_review_retired.sql` in this directory is **not a migration**.
It is the committed golden artifact that pins
`internal/platform/tripwire.RenderMigration()` byte-for-byte, so a hand-edit of the
generated SQL or a stale regeneration is caught by two blocking gates:

- `scripts/api-lint` rule `TRIPWIRE-ARM-PARITY` (`tripwire_arm_rules.go`, run with
  `-strict` by `.github/workflows/api-contract.yml`)
- `internal/platform/tripwire.TestRenderMigration_MatchesCommittedFile`

Regenerate it with `go run ./cmd/gen-tripwire` (no argument → this path).

## Why it lives here and not in `db/migrations/`

It used to be the newest forward tripwire migration. The 2026-07-29 fold squashed
migrations 0257–0315 into `db/baseline/0001_current_schema.sql` and archived the files
under `archive/migrations/post-baseline-2026-07-fold/`, which is a conceptually
immutable historical record — a regenerable golden cannot live there, and
`db/migrations/` no longer carries the file at all. The golden was therefore re-homed
next to the renderer that produces it. The archived copy at
`archive/migrations/post-baseline-2026-07-fold/0301_tripwire_template_review_retired.sql`
is byte-identical today and must stay untouched.

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
