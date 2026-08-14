# Module: auth — CURRENT-STATE REFERENCE

> **Status:** CURRENT-STATE / boundary under Cohesive Platform Redesign
> **Marked:** 2026-08-14

`internal/modules/auth` remains the running V1 authentication/session implementation while the platform redesign proceeds. This page no longer acts as target architecture for identity boundaries.

Target authority:

- `wiki/architecture/cohesive-platform-redesign.md`
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`

Locked direction: Authentication and Authorization remain separate; the current AuthN implementation is sufficient for V1; Keycloak/OIDC/SAML/MFA federation is a future adapter triggered by concrete enterprise identity requirements, not a current dependency.

Use current code/schema/OpenAPI to understand what runs today. Use Git history for the previous detailed living doc if implementation archaeology is needed later.
