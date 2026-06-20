# Feature F8.4 — Spec (SEED — approval pending)

> **Milestone:** 8 — Grade-A Contract & Boundary Completion  ·  **Folder:** `f8.4-deactivation-session-enforcement`
> **Status:** Drafting (seed from post-M7 re-audit Sessions Major #1, CWE-613; **interview + approval pending**)
> **Approved before code:** PENDING — *no implementation begins until this line is filled (Phase 3, fresh session).*

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Enforce at revoke-time, resolve-time, or both? | **Both** (defense-in-depth): revoke sessions when `is_active` flips false, AND re-check `identity.IsActive` in `buildCurrentUser` fail-closed. Either alone leaves a path open. |
| 2 | Reuse existing revoke primitive? | Yes — `RevokeSessionsByUserIDTx` exists (`auth/.../postgres/repository.go:200`); mirror change-password/reset pattern (`service.go:495,702`). Honor H-PRE-1 (no authz-recording read inside the lock-holding tx). |
| 3 | Scope creep risk? | Token rotation / absolute-TTL changes are **out** — this feature is deactivation enforcement only. |

## Consumer contract (FIRST)

- **Consumer(s):** every authenticated request resolving a session (`ResolveSession`); IAM admin deactivate action.
- **Contract:** after a user is deactivated (`is_active=false`), any existing session for that user is **rejected at the next resolve** (treated as revoked); a deactivated identity never yields a `CurrentUser`.
- **Source of truth:** the existing password-change/reset revocation behavior (`auth/application/service.go:495,702`) is the precedent; login already rejects inactive (`service.go:270`).

## What this feature implements

1. `auth.UpdateUser` (`service.go:618-631`) revokes all sessions in-tx when `IsActive` transitions to false (or `iam/application/people_service.go:664-666` deactivate calls revoke after UpdateUser).
2. `buildCurrentUser` (`service.go:845-882`) consults `identity.IsActive` and fails closed if inactive — closing the gap regardless of which path deactivated.

## Non-goals (mandatory)

- No session-token rotation, no idle/absolute-TTL change.
- No change to activate/unlock semantics beyond the symmetric revoke.
- No new IAM endpoint (force-logout PR-7 remains separate).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Deactivate kills live session | auth/iam test: create session → deactivate → next ResolveSession rejected | real |
| Inactive identity fails closed at resolve | buildCurrentUser test with `IsActive=false` | real |
| Change-password/reset revocation unchanged | existing auth tests still green | real |
| No advisory-lock hazard | review confirms no authz-recording read inside the revoke tx (H-PRE-1) | real |

## ADR needed?

- [x] No durable decision — skip (mirrors established revoke-on-credential-change pattern). If resolve-time re-check changes a documented session contract, record a short ADR.
