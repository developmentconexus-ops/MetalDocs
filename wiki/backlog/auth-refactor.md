# Refactor Backlog — auth

> Actionable rows. One row = one PR. Pulled from `wiki/modules/auth-tech-debt.md`.

**Last verified:** 2026-05-12 (Plan 7)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Remove LegacyHeaderEnabled X-User-Id authn bypass | T-001 | S | critical | — | — | open | — |
| R-002 | Wire identity-mutation audit emission (login/logout/pw-change/admin-reset/create-user) | T-002 | M | critical | — | — | merged | Plan 6a (2026-05-11, commits 27c19011 + f27529e8) |
| R-003 | Migrate auth error envelope to RFC 9457 problem+json | T-003 | M | major | — | — | merged | Plan 7 (2026-05-11, commit 95ebedfc) |
| R-004 | Wrap CreateUser identity+role writes in single outer transaction | T-004 | M | major | — | — | open | — |
| R-005 | Add IP-based rate limit on POST /api/v1/auth/login | T-005 | S | major | — | — | open | — |
| R-006 | Throttle TouchSession write (debounce per-session in-memory) | T-006 | S | minor | — | — | open | — |
| R-007 | Extract auth↔iam shared identity contract into platform package | T-007 | L | minor | — | — | open | — |
| R-008 | Add tenant_id to auth_identities with backfill (auth_sessions.tenant_id already added by Plan 3 / migration 0184) | T-008 | M | minor | — | — | partial (sessions done 2026-05-11) | — |
| R-009 | Distinguish malformed-cookie from no-session in Logout error path | T-009 | XS | minor | — | — | open | — |
| R-010 | Author ADR for session-cookie + bcrypt + lockout policy | T-010 | S | minor | — | — | open | — |
| R-011 | Add Go doc comments to all exported auth symbols | T-011 | M | minor | — | — | open | — |
| R-012 | Wire OriginProtection + TrustedOrigins into CSRF middleware | T-012 | S | minor | — | — | open | — |

## Notes

- R-007 + R-008 are `L`: split into reader-extraction PR + writer-migration PR before opening.
- R-001 + R-002 + R-005 are critical-path: prioritise before any new auth feature.
