# Authz Industry Evidence Base (for ADR 0022)

> **Purpose:** External, cited reference base validating the MetalDocs authorization model and ADR [`0022`](../decisions/0022-authz-capability-coherence.md). Compiled 2026-06-03 from four parallel research sweeps (IAM models, policy engines, enforcement placement, anti-patterns). Use this when challenged on "is our authz designed right?" — every decision below maps to a named industry standard or system.
> **Last verified:** 2026-06-03

## TL;DR verdict

ADR 0022 is **squarely in line with industry norm.** The area-scoped `area_admin` + tenant-wide `system_admin` split is literally the Kubernetes `admin`-ClusterRole-via-RoleBinding and GCP grant-at-project-node patterns. The two-tier enforcement is textbook OWASP defense-in-depth. The old code violated **four** named items at once (CWE-863, OWASP A01, OWASP API5/BFLA, latent API1/BOLA + NIST AC-6/AC-5). Keep the custom typed-registry + CI lint — do **not** adopt OPA/Cedar at current scale.

---

## 1. The model: scoped admin is a textbook pattern

| MetalDocs | Industry equivalent | Source |
|---|---|---|
| `system_admin` (tenant-wide) | K8s `cluster-admin` via ClusterRoleBinding; GCP role at Org node | K8s RBAC docs; GCP IAM resource hierarchy |
| `area_admin` (area-scoped via `user_process_areas` row) | K8s `admin` ClusterRole via **RoleBinding** in one namespace; GCP role at project/folder node | K8s RBAC docs; GCP IAM |
| role→capability + per-area membership | hybrid RBAC + ReBAC (relationship-derived authority) | NIST INCITS-359; Google Zanzibar |

- K8s: *"A RoleBinding can reference a ClusterRole to grant the permissions defined in that ClusterRole to resources inside the RoleBinding's namespace... The namespace of the RoleBinding determines where the permissions are granted."* — https://kubernetes.io/docs/reference/access-authn-authz/rbac/
- GCP scoped-admin example: Compute Instance Admin granted on `project_2` "doesn't let them make changes to network resources in other projects." — https://docs.cloud.google.com/iam/docs/resource-hierarchy-access-control
- OWASP: *"ABAC and ReBAC should typically be preferred for application development"* over pure RBAC. — https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html
- Primary refs: NIST SP 800-162 (ABAC), NIST SP 800-178 (RBAC vs ABAC), Zanzibar USENIX ATC 2019 (https://www.usenix.org/conference/atc19/presentation/pang).

**Norm guardrail (now an ADR 0022 acceptance criterion):** the tenant-wide admin must *short-circuit* the scoped check — in K8s/GCP the org/cluster grant inherits downward. `system_admin` passing tier-1 must NOT then be blocked by a missing per-area membership at tier-2. (`authz.Require` already bypasses system_admin — Phase 3 must assert this with a test.)

## 2. Enforcement placement: two-tier = defense-in-depth done right

- OWASP Microservices Cheat Sheet prescribes the exact granularity split: coarse authz at gateway/edge + fine-grained at service via shared lib + business-specific rules in service code. — https://cheatsheetseries.owasp.org/cheatsheets/Microservices_Security_Cheat_Sheet.html
- OWASP: *"Do not depend on any single framework, library, technology, or control to be the sole thing enforcing proper access control"* + *"deny by default"* + *"Centralize the logic for handling failed access control checks."* — Authorization Cheat Sheet.
- OWASP Multi-Tenant: *"Implement authorization checks at the data access layer, not just API layer"* + *"Always validate that requested resources belong to the current tenant."* — https://cheatsheetseries.owasp.org/cheatsheets/Multi_Tenant_Security_Cheat_Sheet.html
- AWS Well-Architected Security Pillar: *"Apply security at all layers"* + least privilege + separation of duties. — https://docs.aws.amazon.com/wellarchitected/latest/security-pillar/security.html

MetalDocs tier-1 (coarse tenant capability at middleware) + tier-2 (fine area-scoped capability at the DB write) maps 1:1. Redundancy across granularity tiers is *prescribed*, not duplication. The "bad smell" is only when the *same* fine decision is copy-pasted with divergent logic — which is exactly the role-string handler check ADR 0022 removes.

## 3. The Postgres GUC + tripwire is a recognized, sound pattern — with guardrails

- Session-GUC-driven isolation is the canonical SaaS RLS pattern: AWS (`current_setting('app.current_tenant')`, https://aws.amazon.com/blogs/database/multi-tenant-data-isolation-with-postgresql-row-level-security/) and Crunchy Data (`current_setting('rls.org_id')`, https://www.crunchydata.com/blog/row-level-security-for-tenants-in-postgres). MetalDocs reading `metaldocs.actor_id` / `metaldocs.tenant_id` GUCs is the same mechanism; the trigger that rejects mutations missing an asserted-capability GUC is a *stricter* fail-closed variant.
- **Documented footguns → ADR 0022 hardening acceptance criteria:**
  1. **Transaction-scoped GUCs under pooling.** PgBouncer transaction mode reuses connections; a session-level GUC leaks to the next client = cross-tenant breach. Must use `set_config(..., true)` / `SET LOCAL`. *(MetalDocs `authz.go` already uses `set_config('metaldocs.asserted_caps', $1, true)` — verify the actor/tenant GUC seeding is also transaction-local.)* — https://www.bytebase.com/blog/postgres-row-level-security-footguns/
  2. **Owner/superuser bypass.** Postgres docs: table owners + superusers bypass RLS unless `ALTER TABLE ... FORCE ROW LEVEL SECURITY` and the app does not run as owner/superuser. — https://www.postgresql.org/docs/current/ddl-rowsecurity.html
- BOLA (OWASP API1:2023, #1 API risk) is the cross-tenant/cross-area object-access failure the tier-2 area check defends against. — https://owasp.org/API-Security/editions/2023/en/0xa1-broken-object-level-authorization/

## 4. Single source of truth: keep custom registry + CI lint (don't adopt an engine yet)

- The "schema validates policy at author/CI time" guarantee (AWS Cedar: a policy referencing a non-existent attribute *fails at save time*; AWS Security Blog 2024-01-10) is reproduced by a typed Go const registry + lint that bans inline capability strings. Sound substitute at our scale. — https://docs.cedarpolicy.com ; https://aws.amazon.com/blogs/security/automate-cedar-policy-validation-with-aws-developer-tools/
- OpenFGA "single model file = source of truth ... cannot be a case where a developer updates the authorization logic in one place but forgets another." — https://openfga.dev/docs/best-practices/source-of-truth
- OPA test framework (`opa test --fail-on-empty --coverage`) = our CI parity tests. — https://www.openpolicyagent.org/docs/policy-testing
- **Correction logged:** the premise "Stripe/GitHub chose lint over an engine" is *unverified*. What IS documented: both treat the permission catalog as a closed, enumerated, versioned set — which is the part that backs our registry-as-source-of-truth direction.

**Revisit-an-engine triggers (record for the future):**
1. Need relationship/hierarchy inheritance (folder→doc cascade, group-of-groups, cross-tenant sharing) → adopt a Zanzibar engine (SpiceDB / OpenFGA / Ory Keto). A capability registry will strain.
2. Need attribute/context-rich policy editable by security without redeploys → adopt Cedar / Amazon Verified Permissions.
3. Need one policy reused across many services/PEPs → OPA/Cedar PDP sweet spot.

## 5. Anti-patterns the old code hit (and the fix closes)

| Old code violated | Source | Fixed by ADR 0022 |
|---|---|---|
| CWE-863 Incorrect Authorization (role-string instead of capability) | https://cwe.mitre.org/data/definitions/863.html | (a) authorize on capability |
| OWASP A01 Broken Access Control ("implement once, reuse") | https://owasp.org/Top10/2021/A01_2021-Broken_Access_Control/ | (a) remove bespoke handler check |
| OWASP API5:2023 BFLA (admin/non-admin via role string) | https://owasp.org/API-Security/editions/2023/en/0xa5-broken-function-level-authorization/ | (a) capability + deny-by-default |
| OWASP API1:2023 BOLA + NIST AC-6/AC-5 (no area scope = latent escalation) | https://owasp.org/API-Security/editions/2023/en/0xa1-broken-object-level-authorization/ ; https://csf.tools/reference/nist-sp-800-53/r5/ac/ac-6/ | (b) area-scoped admin |
| Catalog drift (referenced-but-undefined caps) | OWASP ASVS V4; Oso/OPA policy-as-code | (c) CI binding |

**Load-bearing finding:** fix-parts **(a) remove role check** and **(b) area-scope** are **co-dependent** — removing the role gate without adding scope converts a BFLA over-restriction into a BOLA privilege-escalation. They MUST land in the same change. (c) CI drift detection is the OWASP/Oso-endorsed control, not optional polish.

Auth0 (2024-07-04, https://auth0.com/blog/an-overview-of-commonly-used-access-control-paradigms/): check permissions over roles so "your application doesn't need changes when roles and their permissions change over time."
