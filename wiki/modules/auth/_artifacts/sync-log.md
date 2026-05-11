# Sync log — auth

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-11 · Plan 6a — close T-002

- **Context:** Plan 6a (commits 27c19011 + f27529e8) · wire audit writer in auth handler; emit on login/logout/password-change/createUser
- **Anchors moved:** 0
- **Symbols renamed:** 0
- **T-NNN closed:** T-002 · evidence: handler.go now has WithAudit setter + recordAudit helper; emit calls added in handleLogin, handleLogout, handleChangePassword; handleCreateUser emits via iam admin handler
- **R-NNN updated:** R-002 → merged · commits 27c19011 + f27529e8
- **§11 counts after:** Critical=2 Major=3 Minor=7 (unchanged)
- **Tally gate:** PASS
- **Patched files:** wiki/modules/auth-tech-debt.md · wiki/backlog/auth-refactor.md
- **Structural changes noted (sweep needed):** WithAudit setter added to Handler; new auditdomain OUT-edge from auth handler — §5 Key Files + §8 cross-deps not yet updated
