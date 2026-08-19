# R10-T6 — External / Current-State Evidence Docket

> **Status:** ACTIVE STAGING / EVIDENCE ONLY — NOT TARGET AUTHORITY  
> **Date:** 2026-08-18  
> **Scope:** claim-relative evidence used by the T6 candidate/adjudication packet.  
> **Implementation:** BLOCKED

This docket exists so T6 decisions remain reproducible from primary/current sources rather than conversation memory. Sources are evidence; Product Contract REV001 + T1→T5 + Decision Registry remain requirement authority.

---

## E1 — HTTP error semantics

**Source:** RFC 9457 — Problem Details for HTTP APIs  
https://www.rfc-editor.org/rfc/rfc9457.html

Claim supported:

```text
application/problem+json is a standards-track machine-readable HTTP error shape
Problem Details permits extension members
```

Does NOT prove MetalDocs-specific `code` families; T6 owns those.

---

## E2 — HTTP idempotency / optimistic preconditions

**Source:** RFC 9110 — HTTP Semantics  
https://www.rfc-editor.org/info/rfc9110/

Claims supported:

```text
safe methods + PUT + DELETE are HTTP-idempotent by defined intent
If-Match uses strong validators and is intended to prevent lost updates
failed conditional state-changing request maps to 412 Precondition Failed
```

This supports T6 using natural PUT/DELETE idempotency before adding replay storage, and expressing T2 DRAFT OCC through ETag/If-Match.

---

## E3 — Idempotency-Key is useful evidence, not a current RFC

**Source:** IETF HTTPAPI draft `draft-ietf-httpapi-idempotency-key-header-07`  
https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/

Current status observed 2026-08-18: **expired Internet-Draft**, not binding standard.

Claims useful as design evidence:

```text
Idempotency-Key can make non-idempotent POST/PATCH fault-tolerant
keys must not be reused for materially different requests
resource should document expiry policy
fingerprinting is a credible mechanism
```

MetalDocs therefore defines its own explicit semantics instead of claiming RFC conformance.

---

## E4 — OpenAPI and Go generator maturity

**Sources:**

- OpenAPI Specification 3.1.2: https://spec.openapis.org/oas/v3.1.2.html
- oapi-codegen v2.8.0 release: https://github.com/oapi-codegen/oapi-codegen/releases/tag/v2.8.0

Claims supported:

```text
OAS 3.1.2 is a current published OAS version
current oapi-codegen v2.8.0 added initial OpenAPI 3.1 support in July 2026
```

Does NOT prove MetalDocs should upgrade. T6 conclusion is the opposite: no current consumer requires 3.1 semantics, so a simultaneous contract + description-language/toolchain migration is not justified.

---

## E5 — Keycloak browser flow

**Source:** Keycloak current OIDC securing-apps documentation  
https://www.keycloak.org/securing-apps/oidc-layers

Claims supported:

```text
Authorization Code is targeted to web applications
Resource Owner Password Credentials / Direct Grant exposes credentials to the application
current guidance says ROPC MUST NOT be used, preferring Authorization Code/device flows
```

This supports removal of MetalDocs local password login/change-password routes and use of server-side Authorization Code + ApplicationSession.

---

## E6 — Keycloak provider-directory lookup

**Source:** Keycloak Admin REST API  
https://www.keycloak.org/docs-api/latest/rest-api/index.html

Claims supported:

```text
provider users can be searched by username/email/name
provider returns stable provider user IDs and basic enabled/profile hints
```

This proves a provider-directory lookup is technically viable. It does NOT promote provider roles/groups into MetalDocs Authorization.

---

## E7 — ONLYOFFICE integration shape

**Sources:**

- Permissions: https://api.onlyoffice.com/docs/docs-api/usage-api/config/document/permissions/
- Callback handler: https://api.onlyoffice.com/docs/docs-api/usage-api/callback-handler/

Claims supported:

```text
ONLYOFFICE can expose edit/review/read behaviors
Document Server persistence/status uses callbackUrl integration with the application's storage/document manager
```

This proves ONLYOFFICE is capable but has a server/callback lifecycle that must be adapted behind T4/OCC if selected. It does NOT justify choosing it before fidelity evidence.

---

## E8 — EigenPal/docx-editor integration shape

**Source:** current upstream repository  
https://github.com/eigenpal/docx-editor

Claims supported:

```text
current React adapter loads DOCX from ArrayBuffer/documentBuffer
current upstream exposes browser editing and OOXML parser/serializer architecture
```

This supports using a browser-buffer adapter as the first integration candidate because it composes naturally with T4 upload/admission + WorkingContent OCC.

It does NOT prove production fidelity for MetalDocs. That remains a required representative-corpus proof gate.

---

## E9 — Browser session/CSRF guidance

**Sources:**

- OWASP CSRF Prevention Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html
- MDN secure cookie/session guidance: https://developer.mozilla.org/en-US/docs/Web/Security/Authentication/Session_management

Claims supported:

```text
Secure + HttpOnly + SameSite are appropriate session-cookie controls
SameSite is only part of a complete CSRF defense
custom request-header CSRF tokens are a recognized same-origin browser defense
```

This supports a session-bound `X-CSRF-Token` in addition to a `__Host-` Secure/HttpOnly/SameSite cookie.

---

# Current-repository evidence

## R1 — current public API encodes superseded ownership

`api/openapi/v1/openapi.yaml` currently contains legacy/current-state surface including old Auth/IAM/Tenant/Taxonomy/Tokens/Approval/Distribution/Notifications concepts.

Disposition:

```text
use as Structural-Inversion counterexample/evidence only
never use route existence as target requirement
```

## R2 — current frontend taxonomy is not target authority

Current router/features preserve legacy `approval`, taxonomy/tokens/IAM and related feature decomposition.

Disposition:

```text
preserve useful React/TanStack/generated-client mechanisms
rederive semantic feature/workspace vocabulary from Product Contract + T1→T5
```

## R3 — historical reviewer-editor cockpit is now an explicit counterexample

Historical design allowed review-mode edits/suggestions against mutable WorkingContent.

Current Product Contract/T1/T2 instead require:

```text
review exact immutable Submission
feedback does not mutate Submission
RETURN → same Revision DRAFT
Author edits → new Submission
```

Therefore editor feature capability does not justify reviewer mutation.

## R4 — historical numbering proves only the simple examples

Historical UI/runtime showed examples such as:

```text
PO-...
PO-RH-003
preview does not need to reserve
```

It does not prove a customer-defined token language or reset engine.

---

# Evidence boundary

No source above can override ratified MetalDocs product semantics.

A provider/standard can prove:

```text
capability exists
mechanism has a constraint
standard defines a transport semantic
```

It cannot prove:

```text
MetalDocs needs the feature
provider mechanism owns business truth
current implementation should be retained
future capability should enter Launch
```

That distinction is binding for T6 Global-Maximum reasoning.
