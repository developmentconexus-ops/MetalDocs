# IAM module doc — sync log

One line per `metaldocs-module-doc-sync` run. Append-only.

- 2026-05-11 · Plan 3 (session-bound tenant resolution, post-merge sweep). Patched anchors shifted by ~3 lines in `admin_handler.go` and `middleware.go` / `routes_memberships.go` (file growth from `tenant.FromContext` migration). Files: `wiki/modules/iam.md` (§2 + §6.4 envelope anchors :129→:132, :137→:150); `wiki/modules/iam-tech-debt.md` (T-005 :316→:319/:457→:454; T-006 :129→:132/:137→:150; Last verified bump); `wiki/backlog/iam-refactor.md` (Last verified bump); `wiki/README.md` (iam-tech-debt + iam-refactor index stamps). T-NNN affected: T-005, T-006 (anchors only — severity unchanged, debt not resolved). R-NNN affected: none. Escalation: no — verified no Plan 3 ADR exists in `wiki/decisions/` (flagged to caller).
