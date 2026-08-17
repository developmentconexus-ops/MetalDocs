# MetalDocs — Single-Company Deployment / Tenancy Rebaseline — Independent Fable Review Request

> **Status:** CANDIDATE / INDEPENDENT REVIEW REQUEST — **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Authority baseline HEAD:** `b2926f5a2d885ea8cc8a48f1261a1d8750498020`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Scope:** bounded deployment/tenancy rebaseline after operator clarified the actual V1 product objective
> **Implementation gate:** **CLOSED — design/documentation only.**
> **Authority note:** this artifact is review evidence only. It does not amend R9.5, R10-A, R10-B1, B2, handoff, code, schema, OpenAPI, frontend or deployment.

---

# 0. Cold reviewer bootstrap

Reconstruct repository state fresh. Do not use prior conversation memory as authority.

Read `AGENTS.md` and follow its complete read order / authority chain. At minimum read:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. `wiki/architecture/r10-technical-architecture.md`
7. the promoted GCR evidence chain if needed to understand why current tenancy laws exist
8. this candidate packet
9. current code/schema/runtime only as claim-specific evidence
10. primary external sources where deployment-model claims need verification

Current implementation is evidence, never target entitlement.

Apply the DevelopmentConexus Engineering Method proportionally and aggressively:

```text
Evidence
→ Known / Inferred / Unknown / Deferred
→ Root Cause
→ Target Invariant
→ Constraints
→ Credible Alternatives
→ Local Maximum vs Global Maximum
→ Essential vs Accidental Complexity
→ YAGNI / Future Cost
→ Authority / Boundary
→ Enforcement
→ Proof Strategy
→ Adversarial Challenge
→ Decision
→ Reopen Triggers
```

Use Structural Inversion. If the legacy implementation had been single-company from day one with no `tenant_id`, which current target laws would still follow from the real product constraints?

---

# 1. Changed operator requirement — material reopen evidence

The operator has now clarified the actual V1 product objective:

> MetalDocs is being built first for the operator's own company, Metal Nobre. The immediate objective is to make the system work professionally for one company. Commercialization to other companies is a future possibility, but there is no current requirement to operate multiple customer companies inside the same application/database deployment.

The operator explicitly does **not** yet know the future commercial tenancy model and does not want current development complexity dominated by a hypothetical choice between:

```text
pooled SaaS
shared backend + database-per-customer
single-customer deployment stamps
on-prem/private deployments
hybrid models
```

The future product should remain structurally productizable without customer-specific code forks, but V1 should not implement shared multitenancy merely to preserve that option.

This is a changed requirement / clarified deployment constraint and is material reopen evidence under the Method.

---

# 2. External architecture evidence

Primary external architecture guidance confirms that dedicated deployment is a legitimate architecture, not an anti-pattern:

- Azure Architecture Center, **Architectural Approaches for a Multitenant Solution**: a deployment stamp may serve one tenant or several; single-tenant stamps are easier because the application does not require internal multitenancy logic and provide strong isolation.
  - https://learn.microsoft.com/en-us/azure/architecture/guide/multitenant/approaches/overview
- Azure Architecture Center, **Deployment Stamps pattern**: independent copies of application components and data stores can be deployed per tenant/customer and scaled operationally by creating more stamps.
  - https://learn.microsoft.com/en-us/azure/architecture/patterns/deployment-stamp
- Azure Architecture Center, **Tenancy Models**: when tenants share one deployment, application code and a tenant identifier are typically required to isolate them — this cost follows from shared deployment, not from the abstract existence of multiple customers.
  - https://learn.microsoft.com/en-us/azure/architecture/guide/multitenant/considerations/tenancy-models
- AWS Well-Architected SaaS Lens / tenant-isolation guidance distinguishes silo, bridge and pool models. Full-stack silo gives a tenant a dedicated application stack; pool uses shared resources and requires additional isolation constructs.
  - https://docs.aws.amazon.com/wellarchitected/latest/saas-lens/silo-pool-and-bridge-models.html
  - https://docs.aws.amazon.com/solutions/multi-tenant-architectures-on-aws/
- AWS **Full stack silo and pool** explicitly warns that dedicated tenant stacks should still use one common product rather than one-off tenant customizations.
  - https://docs.aws.amazon.com/whitepapers/latest/saas-architecture-fundamentals/full-stack-silo-and-pool.html
- Keycloak Organizations is explicitly a capability for multi-tenancy/organization context *within a realm*. It is therefore not automatically needed when one deployment serves one company.
  - https://www.keycloak.org/docs/latest/server_admin/

External sources are evidence about credible deployment patterns, not requirement authority.

---

# 3. Candidate target invariant

> **MetalDocs V1 is a production-grade single-company application for Metal Nobre, deployed as one organization/company per deployment stamp. The product codebase, domain model, migrations, provider seams and deployment artifact remain common and replicable so a future second customer can receive another deployment without a source fork. Shared/pooled multitenancy is deliberately deferred until a real commercialization/scale requirement proves which tenancy model is needed.**

This target deliberately separates:

```text
productizability  = required now
shared multitenancy = NOT required now
```

Prepare seams, not pooled tenancy machinery.

---

# 4. Candidate deployment model

V1 deployment:

```text
Metal Nobre deployment
  1 frontend deployment
  1 backend modular monolith deployment
  1 MetalDocs product-state PostgreSQL database
  1 Managed Artifact Store profile
  1 Keycloak realm/client configuration
  1 company / organization root
```

Future commercial default, until evidence changes it:

```text
same source/build/migrations
→ deployment stamp for Customer A
→ deployment stamp for Customer B
→ deployment stamp for Customer C
```

Customer-specific product forks/branches/repositories are forbidden as a commercialization strategy.

Customer differences should enter only through explicit configuration/provider/integration seams whose real consumers are proven.

---

# 5. Candidate decision set to challenge

## SC-R1 — deployment tenancy

```text
OUTCOME CANDIDATE = RESTRUCTURE

V1 shared pooled customer multitenancy = NO
one company/organization per deployment = YES
same product codebase/build = YES
customer-specific forks = NO
future commercial tenancy topology = DEFER
```

A second real customer is not itself enough to force pooled tenancy; it is enough to trigger an explicit deployment economics/operations review.

Reopen pooled/shared tenancy only when real evidence such as fleet size, infrastructure cost, operations burden, central analytics/control-plane need, shared integration requirements or commercial constraints makes dedicated stamps materially inferior.

---

## SC-R2 — meaning of `Tenant`

The current architecture uses `Tenant` both as product/organization root and as database isolation partition.

Candidate split:

```text
Tenant semantic root         = MAY remain
Tenant database partition    = REMOVE V1
```

Candidate meaning:

> `Tenant` is the single company/organization root configured for a deployment, used only where company-wide product semantics actually require a root object or company-wide scope.

Exactly one active Tenant/company root exists per deployment V1.

The review MUST compare at least:

A. retain `Tenant` as one singleton semantic root;
B. rename/collapse it into a more precise Organization/Company root;
C. remove the durable root entirely and express company-wide facts as deployment configuration.

Do not retain `Tenant` merely because current code already uses the noun. Do not rename it merely for aesthetic cleanliness if the semantic root still has real consumers such as company settings or company-wide Authorization scope.

---

## SC-R3 — B1 tenant-qualified identity / composite PK-FK law

Current promoted B1 assumes durable tenant-owned identity:

```text
PRIMARY KEY (tenant_id, id)
FOREIGN KEY (tenant_id, target_id)
```

Candidate:

```text
shared-customer tenant partition = absent V1
normal durable identity = id UUID PRIMARY KEY
normal references = typed FK by target id
```

Do **not** replace `tenant_id` with `company_id`, `organization_id`, `deployment_id`, or another repeated partition column unless a real invariant requires it.

Cross-owner FK authority-neutrality and cross-owner `RESTRICT/NO ACTION` laws are independent of customer multitenancy and are expected to remain.

The reviewer must identify any concrete business relationship whose meaning genuinely requires the company root identity on every row even under one-company-per-deployment.

---

## SC-R4 — RLS / tenant context

Current B1 requires application tenant predicates + composite keys/FKs + ENABLE/FORCE RLS + fail-closed tenant context.

Candidate:

```text
cross-company RLS/customer tenant context = REMOVE V1
```

Reason: deployment boundary already isolates the one customer company. RLS was explicitly promoted as Tenant isolation only, not business Authorization. With no second customer in the same DB it protects against an unreachable cross-customer state while imposing permanent schema/repository/session/worker complexity.

Do not replace tenant RLS with Area/role RLS. Canonical business Authorization remains application/domain Authorization exactly as frozen.

Database role hardening such as non-superuser/non-owner serving roles may remain if independently justified as ordinary DB security, but it must not masquerade as multitenancy.

---

## SC-R5 — Keycloak tenancy posture

Keycloak remains V1 Authentication provider.

Candidate single-company posture:

```text
one ordinary Keycloak realm/trust domain for the deployment
one MetalDocs client/application
Metal Nobre users
Keycloak Organizations feature = NOT required V1
realm-per-customer question = irrelevant inside one stamp
```

If a future customer receives its own deployment stamp, its deployment can receive its own provider configuration/realm without adding multi-company logic to MetalDocs.

The review must ensure this does not cause code-level provider assumptions that prevent future federation/SSO/productization.

---

## SC-R6 — AuthenticationSubjectBinding / ApplicationSession

Because a deployment contains one company, candidate B2-1 becomes:

```text
AuthenticationSubjectBinding
  id
  organization_user_id
  issuer
  subject
  lifecycle/evidence

ApplicationSession
  id
  subject_binding_id
  authentication context / assurance / expiry / revocation
```

Candidate removals:

```text
tenant_id from binding/session solely for customer isolation
tenant-first routing before login
cross-tenant subject lookup logic
tenant selector / company switching within a live deployment
```

Stable `issuer + subject` binding, opaque application Session, no AuthZ snapshot, anti-corruption boundary and no email auto-binding remain independently justified.

The reviewer must decide exact uniqueness/cardinality without importing hypothetical cross-company personas into the V1 schema.

---

## SC-R7 — Authorization company-wide scope

Authorization remains essential even in one company.

Candidate:

```text
AreaScope = unchanged
TenantScope = reinterpret as whole-company/root scope OR rename later if clearer
```

Five frozen roles and RoleAssignment semantics remain unless a genuine contradiction is found.

Do not remove company-wide scope simply because there is one company; `tenant_owner`/company-wide permissions may still need a scope distinct from Area.

The reviewer should determine whether retaining the token/noun `TenantScope` creates meaningful future confusion or whether renaming it now is pure churn.

---

## SC-R8 — Tenant lifecycle / customer deletion / erasure

Current target includes:

```text
Tenant ACTIVE / SUSPENDED / ERASED
TenantDeletionRequest
TenantErasureRecord
erasure tombstones
restore reconciliation
```

Those facts were designed partly for SaaS customer lifecycle.

Candidate:

```text
customer/Tenant suspension-deletion-erasure product feature = DEFER V1
```

Reason: V1 is the company's own operational system; there is no real customer offboarding workflow requiring an administrator to erase the entire company from itself.

This MUST NOT weaken:

```text
Document/Evidence disposition
RetentionBinding
LegalHold
backup/restore
ordinary user offboarding
Artifact deletion when lawfully disposed
PII minimization
```

The reviewer must determine whether some deployment decommission/restore safety property still requires a durable tombstone or whether that belongs to operations/backup tooling rather than product domain state.

---

## SC-R9 — Tenant Portability Export

Current export contracts distinguish:

```text
Backup
Tenant Portability Export
Governed Subject Export
External Repository PUBLISH_COPY
```

Candidate:

```text
Tenant Portability Export = DEFER
Backup = KEEP
Governed Subject Export = KEEP
PUBLISH_COPY = KEEP
```

Tenant portability was primarily a future SaaS/customer-exit requirement. V1 Metal Nobre still needs reliable backup and governed exports, which are different contracts.

A future commercial second customer, contractual portability requirement, or migration between deployment stamps is a reopen trigger.

---

## SC-R10 — background work / routing

Current B1 has tenant enumeration or globally claimable tenant-routing intents because one runtime may serve many companies.

Candidate single-company execution:

```text
worker/job
→ this deployment's product DB
→ discover/claim due mechanism state
→ perform normal canonical Authorization/system execution rules where applicable
```

Remove customer/Tenant routing metadata and cross-tenant discovery machinery when it exists solely to select among customer companies.

Keep independent async properties:

```text
outbox / durable intent
idempotency
lease / retry / DLQ
external effect truth
```

R10-D still owns execution details.

---

## SC-R11 — Artifact storage customer namespace

Current frozen storage uses tenant-namespaced provider keys.

Candidate:

```text
opaque immutable provider keys = KEEP
customer/Tenant prefix as isolation law = REOPEN / likely REMOVE
```

The deployment/storage boundary already isolates the company. Do not retain customer namespace merely for hypothetical future pooling.

The reviewer should distinguish harmless implementation key layout from a semantic/security invariant and avoid overconstraining R10-C.

---

# 6. Decisions explicitly expected to remain out of this reopen

Do not rediscover or reopen these merely because deployment tenancy changes:

```text
modular monolith
8+3 semantic ownership topology unless a fact is truly orphaned by Tenant simplification
Keycloak as AuthN provider
Organization/User/Area/Group/GroupMembership semantics except Tenant-root consequences
Authorization/Approval separation
five frozen roles unless TenantScope wording creates a genuine semantic issue
Document / Revision / WorkingContent / RevisionSubmission
OCC / immutable Submission
Approval / SoD / fresh-auth
Rendition / Release
Artifact immutable exact-byte identity
ManagedArtifactStore provider seam
malware inspection production invariant
Dossier / Evidence
Retention / LegalHold / Disposition
Audit append-only/tamper-evident ownership
non-PII post-erasure audit principle as applicable to actual data deletion
Interchange/Historical Migration except Tenant Portability
same-commit Audit + durable-intent laws inside the product DB
cross-owner ownership-neutral FK action laws
provider DB non-atomicity
Search/Notifications classifications
no BPM/ReBAC/ECM generic platforms
```

The whole product should not be redesigned because the deployment tenancy premise changed.

---

# 7. Productization seams that MUST remain

Single-company V1 must not become company-hardcoded software.

Required seams:

```text
one source repository / one product codebase
same migrations and release artifact across deployments
company identity/settings represented as data/config, not code branches
provider-independent Authentication boundary
provider-independent Artifact boundary
integration adapters remain replaceable
opaque technical UUIDs
no references to "Metal Nobre" in domain branching/business mechanics unless they are actual configured company data
no customer-specific forks as normal delivery model
```

Do not build a fleet control plane, billing, customer registry, tenant router, central SaaS admin plane, per-customer deployment orchestrator or IaC factory now. Those are future operational capabilities, not seams.

---

# 8. Hard questions for independent challenge

The reviewer MUST attack at least:

1. Does one-company-per-deployment actually reduce total complexity, or merely move tenancy complexity to deployment/operations?
2. Which current B1 laws remain essential independently of customer pooling?
3. Is removing tenant-qualified PK/FKs safe, or does `Tenant` have genuine data-ownership semantics beyond isolation?
4. Is retaining a singleton `Tenant` useful, or is it a misleading leftover noun?
5. Would removing `tenant_id` now create an expensive future migration if pooled SaaS later wins? Is that future cost justified by current complexity savings?
6. Is there a cheaper seam than tenant columns everywhere that preserves realistic future productization?
7. Does removing RLS weaken defense-in-depth against any currently reachable threat, given RLS was Tenant isolation only?
8. Should any non-tenant RLS survive for another independent security property?
9. Does one deployment per company make Keycloak Organizations unnecessary without harming future enterprise IdP federation?
10. Does Authentication binding/session no longer need any Tenant/company dimension?
11. Does company-wide Authorization still require a root scope; should `TenantScope` be retained/renamed?
12. Is Tenant lifecycle `ACTIVE/SUSPENDED/ERASED` still a real business concept in an internal deployment, or SaaS lifecycle machinery?
13. Does backup/restore require an erasure tombstone even when the whole deployment belongs to one company?
14. Can Tenant Portability Export be safely deferred while preserving Backup and Governed Subject Export?
15. Which background routing laws disappear and which async laws survive?
16. Is provider key tenant namespacing still useful as defense-in-depth or merely dead prefix complexity?
17. Does the proposed model preserve a clean path to a second customer through another deployment stamp?
18. At what concrete triggers should pooled/shared tenancy be reconsidered?
19. Is there any currently frozen domain fact that depends semantically on multiple Tenants coexisting in one DB?
20. Does this rebaseline accidentally optimize only for Metal Nobre through hardcoded assumptions?
21. Are there any other pieces of SaaS/customer-fleet machinery in R3–R10 that should be removed/deferred by the same changed requirement?
22. Conversely, is any proposed removal actually essential complexity masquerading as SaaS complexity?

Perform a subtractive pass:

> What can now be deleted from the target because the real V1 has exactly one company per deployment, without weakening a business, safety, audit, retention or productization property?

---

# 9. Candidate resulting stage implications

If approved, expected impacted authorities are bounded but material:

```text
R9.5:
  Tenant/customer lifecycle semantics
  Tenant Portability Export
  storage tenant namespace wording if promoted as invariant
  possibly TenantScope wording only if semantically necessary

R10-A:
  Organization Tenant fact-family/lifecycle scope
  possibly no topology change

R10-B1:
  MATERIAL STRUCTURAL REOPEN
  tenant-qualified identity law
  same-Tenant composite FK law
  Tenant RLS law
  Tenant context serving law
  background tenant discovery/routing law
  product DB and cross-owner transaction laws otherwise expected to survive

R10-B2:
  PAUSE / rederive from the single-company substrate before continuing

R10-D:
  remove future cross-customer routing assumptions

R10-E:
  no tenant selector/company switching required V1
```

Implementation remains blocked.

---

# 10. Required reviewer output

Required verdict:

```text
APPROVE SINGLE-COMPANY DEPLOYMENT/TENANCY REBASELINE
or
APPROVE ... WITH MATERIAL FIXES
or
DO NOT APPROVE ...
```

Required output:

1. verdict;
2. BLOCKER / MAJOR / LOW findings;
3. disposition of SC-R1 through SC-R11 individually;
4. precise answer: retain singleton `Tenant`, rename it, or remove it;
5. resulting R9.5 reopen set;
6. resulting R10-A reopen set;
7. resulting R10-B1 reopen set;
8. resulting B2 scope/decomposition implications;
9. anything else in R3–R10 that is customer-multitenancy machinery and should be deferred;
10. anything proposed for removal that actually must remain;
11. concrete future trigger for reconsidering pooled/shared tenancy;
12. whether another broad review is required after adjudication;
13. exact promotion conditions.

Write the independent result to a NEW artifact:

`docs/superpowers/analysis/2026-08-17-single-company-deployment-tenancy-rebaseline-independent-fable-review.md`

Authorization for the reviewer is limited to creating that review artifact, committing it on the existing design branch and pushing it. Do not amend authority, handoff, code/schema/OpenAPI/frontend/deployment or begin B2 implementation.
