# MetalDocs — Single-Company Deployment / Tenancy Rebaseline — Independent Fable Review

> **Status:** INDEPENDENT COLD REVIEW — evidence, **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Reviewer:** independent cold session (Fable), no prior-conversation authority used
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Reviewed HEAD:** `cba89d9da049998d0700be661cf9fab0fa10afba`
> **Review packet:** `docs/superpowers/analysis/2026-08-17-single-company-deployment-tenancy-rebaseline-fable-review-request.md`
> **Method:** DevelopmentConexus Engineering Method v1.0.0 (`docs/engineering/standards/root-cause-global-maximum-method.md`)
> **Implementation gate:** CLOSED — this review changes no authority, code, schema, OpenAPI, frontend or deployment.

---

## 0. Review basis

Read in full, per `AGENTS.md` read order: `AGENTS.md`; the Method mirror; `wiki/references/current-agent-handoff.md`; `wiki/architecture/cohesive-platform-redesign.md`; `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`; `wiki/architecture/r10-technical-architecture.md`; the candidate packet.

Claim-specific current-code evidence (evidence of carrying cost, never target entitlement):

```text
tenant_id in Go source              = 1,965 occurrences across 354 files
FORCE ROW LEVEL SECURITY (SQL)      = 91 statements
SeedTxTenant call sites             = 143
composite PRIMARY KEY (tenant_id, … = 6 SQL files
tenant-namespaced storage keys      = internal/modules/documents/application/keys.go:12
                                      ("tenants/%s/documents/%s/revisions/%s.docx")
```

External deployment-model sources cited by the packet (Azure multitenant approaches / Deployment Stamps / tenancy models; AWS SaaS Lens silo-bridge-pool; AWS full-stack silo-and-pool; Keycloak Organizations docs) were checked for consistency with the packet's characterization; the characterization is accurate: single-tenant stamps are a legitimate first-class architecture, tenant-identifier/isolation machinery is a cost that follows from **shared deployment**, not from the abstract existence of multiple customers, and the silo model's canonical failure mode is per-customer forks, not the stamp itself.

Adversarial posture: the candidate was treated as a claim to falsify, not a conclusion to confirm. The strongest counterarguments attempted are recorded in §2.4.

---

## 1. Verdict

```text
VERDICT = APPROVE SINGLE-COMPANY DEPLOYMENT/TENANCY REBASELINE
          WITH MATERIAL FIXES

BLOCKER = 0
MAJOR   = 3   (M1 deployment↔database identity handshake;
               M2 re-anchor erasure/PII/audit-skeleton obligations to data subjects;
               M3 explicit frozen-permission-catalog amendment)
LOW     = 6
```

The candidate's core claim survives every falsification attempt I could construct: the tenant-partition machinery in the promoted target (composite tenant-qualified PK/FK, tenant RLS stack, fail-closed tenant context, tenant enumeration/routing, tenant lifecycle/erasure/portability as product features, tenant-namespaced storage keys as invariant) derives from the **pooled-multitenancy premise**, not from the controlled-information domain. The operator's clarified requirement removes that premise. Under the Method this is legitimate material reopen evidence (changed requirement), not preference or hypothesis. The three MAJOR fixes prevent the rebaseline from silently deleting two real properties and one authority-coherence obligation that must survive it.

---

## 2. Method application

### 2.1 Root cause

The tenancy laws in B1 §9.1/9.2/9.5/9.6/9.10 are not domain truths; they are **consequences of one architectural premise** — "multiple customer companies coexist inside one backend/database deployment" — inherited from the legacy implementation's pooled model and carried into the target without a current requirement supporting it. The root cause of the current mismatch is a premise change, so the correct correction is at the premise (deployment tenancy model), not patches inside individual laws.

### 2.2 Target invariant (accepted as stated)

The packet §3 candidate invariant is correctly formed: single-company production-grade V1 + one common codebase/build/migrations + productization via deployment stamps + shared/pooled tenancy deferred until evidence selects a model. It separates the essential property (**productizability without forks**) from the accidental one (**shared-deployment isolation machinery**).

### 2.3 Structural Inversion Test

If MetalDocs had been built single-company from day one, which promoted laws would still follow from the real constraints?

**Would still exist (essential, pooling-independent):** one product-state PostgreSQL DB; `metaldocs` schema; no cross-provider-DB atomicity; UUID technical identity; business/provider IDs never PKs; cross-owner FK authority-neutrality with closed `RESTRICT/NO ACTION` action set; no polymorphic business registry; TIMESTAMPTZ/BYTEA-SHA-256/TEXT+CHECK/NULL primitives; semantic-persistence-class × mutation-law classification; READ COMMITTED + narrowest-sufficient invariant mechanism; same-commit Audit append + durable intent; composition-owned cross-owner transactions; outbox/idempotency/lease/retry/DLQ; non-owner NOSUPERUSER serving role; separate non-serving maintenance trust surface; the entire R9.5 domain semantics (Document/Revision/Submission/Approval/SoD/Artifact/Dossier/Retention/Hold/Audit/malware gate).

**Would never have existed:** `(tenant_id, id)` composite PKs; same-Tenant composite FKs; ENABLE+FORCE RLS with fail-closed tenant GUC context; tenant enumeration/routing in background discovery; Tenant ACTIVE/SUSPENDED/ERASED as a product lifecycle; TenantDeletionRequest/TenantErasureRecord/tombstones; Tenant Portability Export; tenant-namespaced key **invariant**; realm-per-tenant deliberation; tenant selector UX.

The inversion cleanly partitions the target. That is the strongest available structural evidence that the candidate reopen boundary is correctly drawn.

### 2.4 Falsification attempts (strongest counterarguments, all rejected)

1. **"Keep tenant_id: future pooled migration is prohibitive."** Rejected. The declared future default is deployment stamps, which never need it. Pooled is one of five unchosen models; the Method forbids converting imagined possibility into a present column on every row of every table. The present carrying cost is empirically large (§0 counts — and the target redesign would reproduce that mass across B2–B6, R10-C/D/E, tests, and every future feature). The future cost, if pooled ever wins, is a bounded well-understood migration (add column, backfill the single company UUID from the durable root — see M1 — re-key, re-add RLS) executed under a design stage that would be required anyway to build the pooled control plane. Asymmetric: certain permanent cost vs. contingent bounded cost.
2. **"Removing RLS weakens defense-in-depth."** Rejected. Promoted law (B1 §9.5) is explicit: RLS is Tenant isolation **only**, never Authorization. With exactly one company in the database there is no second tenant's row to leak; the protected state is structurally unreachable. Against every remaining real threat (SQL injection, compromised serving credential, authz bug) RLS with a single tenant predicate is inert — the attacker's reachable data set is identical with or without it. No property is lost. Compensating Area/role RLS would violate the frozen "RLS ≠ canonical AuthZ" law and is correctly forbidden by the packet.
3. **"Tenant lifecycle/erasure is needed for LGPD."** Rejected as stated, but it exposes M2. LGPD/GDPR obligations attaching to V1 are **data-subject-level** (Metal Nobre's employees/users), not company-self-erasure. A company erasing itself from its own operational system is not a product workflow; deployment decommission is an operations action. But the frozen erasure design also carried the *data-subject* PII obligations (PII-minimized immutable Audit skeleton, separately erasable enrichment, GCR-R4 DEK reopen trigger) — those must be re-anchored, not deferred with the vehicle that carried them. See M2.
4. **"This is the fifth material local maximum the GCR said didn't exist — reopening one day after freeze proves process instability."** Rejected. The GCR answered correctly **under the pooled premise then in force**. The operator subsequently changed the requirement. Changed requirement is the Method's canonical legitimate reopen evidence (§4). The bounded-reopen discipline is functioning, not failing.
5. **"Stamps just move the complexity to operations."** Partially true and correctly priced: at fleet scale, per-stamp ops cost grows with N and that is exactly the named reopen trigger (§11). For V1 there is exactly one deployment, so total system complexity strictly decreases — nothing is "moved" because the fleet does not exist. The one genuinely new failure class stamps introduce (wrong-database attachment/restore across stamps or environments) is real and unhandled by the candidate — that is M1, and it is cheaply closed structurally.

### 2.5 Local vs Global Maximum

Keeping the tenancy machinery "because it is already designed" is the local maximum — optimizing inside a structure whose premise no longer holds. The candidate is the global-maximum move: restructure the substrate to the real constraint set now, at the cheapest possible moment (B2 not yet started; implementation gate closed; zero code churn), while preserving the productization seam. Timing is optimal and the alternative only gets more expensive.

---

## 3. Findings

### M1 — MAJOR — Deployment ↔ database identity handshake must replace per-row tenancy as the wrong-data-attachment guard

**Claim.** Removing `tenant_id` from every row deletes the only structural, in-band record of *whose data a database contains*. In the promoted pooled target this also guarded a failure class the stamp model makes **more** reachable, not less: attaching the wrong database to a deployment. Concretely: Customer A's backup restored into Customer B's stamp (future), or a production backup restored into a dev/staging stamp whose profile has malware inspection disabled (reachable in V1 already). The candidate deletes the mechanism without naming the surviving property.

**Root cause.** The property "this deployment serves exactly this company's data" was implicitly enforced by pooled machinery; the rebaseline removes the machinery without restating the property at its new natural boundary (deployment identity).

**Required fix (material, cheap, structural).** The rebaselined substrate must require:

```text
exactly one durable company/organization root row (UUID id) per database    (see SC-R2)
deployment configuration pins the expected company root UUID (and profile)
startup/readiness verifies configured UUID == database root UUID
mismatch = FAIL CLOSED (refuse to serve)
```

This preserves the real property at ~zero marginal cost, composes with GCR-R3's single-sourced deployment-profile declaration (same deployment-identity-integrity family), and gives any future pooled backfill an unambiguous tenant value. It does **not** reintroduce per-row company columns; SC-R3's prohibition on `company_id`-by-reflex stands.

### M2 — MAJOR — Data-subject erasure obligations must be re-anchored, not deferred with tenant erasure

**Claim.** SC-R8 defers the Tenant lifecycle/deletion/erasure family. But three obligations currently ride on that family and are **not** SaaS machinery: (a) the B6 proof that the immutable Audit skeleton is PII-minimized/non-PII (ledger §5, R10 §2.2 Audit, handoff successor obligations); (b) separately-erasable/projection-only human-readable enrichment; (c) the GCR-R4 reopen trigger for DEK/crypto-erasure if a named immutable Target Data family emerges. Metal Nobre's employees are LGPD data subjects; user-level erasure/PII-minimization is a real single-company V1 requirement. If the tenant-erasure vehicle is deferred wholesale, these obligations silently evaporate.

**Required fix.** The adjudicated amendment must explicitly re-anchor (a), (b), (c) to **user/data-subject-level erasure and ordinary user offboarding** (Organization User lifecycle + Audit skeleton, B2/B6 scope), independent of any Tenant lifecycle. Deployment decommission (destroy stamp + backups per retention policy) replaces company-level erasure and belongs to operations, not domain state.

### M3 — MAJOR — Frozen permission catalog must be amended explicitly

**Claim.** Deferring Tenant lifecycle (SC-R8) and Tenant Portability Export (SC-R9) orphans two frozen R9 permissions: `tenant.export` and `tenant.deletion.request` (ledger §1, base catalog of 29). Leaving grants pointing at capabilities that no longer exist in the target is exactly the hand-synced-enumeration/dual-authority drift class this program exists to kill.

**Required fix.** The R9.5 amendment must record the bounded catalog change (29 → 27 base permissions; `tenant_owner` bundle updated accordingly) as an explicit adjudicated delta, with reinstatement named in the portability/lifecycle reopen triggers. No other permission is affected (`tenant.settings.manage` retains its consumer: company settings; see SC-R2/SC-R7).

### L1 — LOW — Singleton company root needs structural enforcement

"Exactly one ACTIVE company root per deployment" must be enforced structurally in B2 (e.g., constant-key single-row table or unique partial index), not by convention. A control that cannot fire is not a control.

### L2 — LOW — Retain the noun `Tenant` redefined; forbid a partial rename

See §5. Whichever noun the operator ratifies, a **partial** rename (e.g., `Company` entity coexisting with `tenant_owner`/`tenant.settings.manage` vocabulary) is forbidden — it creates two vocabularies for one concept, worse than either consistent option.

### L3 — LOW — B1 role-hardening law survives with rewritten rationale

`NOSUPERUSER` / non-owner serving role / separate non-serving maintenance trust surface remain as ordinary least-privilege DB security (independently justified). `NOBYPASSRLS` becomes vacuous with no RLS; drop the wording rather than keep a dead attribute masquerading as an isolation control. Per-Tenant-iteration language in the maintenance-surface law (§9.7) drops.

### L4 — LOW — Storage key layout is freed, not re-frozen

R10-C must inherit "opaque immutable keys; layout free" — not a tenant-prefix invariant and not a mandated prefix-removal either. Deployment-scoped bucket/container isolation is stamp configuration (ops), not product law. Current `tenants/%s/...` layout (keys.go:12) is harmless current-state evidence.

### L5 — LOW — Do not keep a vestigial three-state lifecycle for one hypothetical state

If the ACTIVE/SUSPENDED/ERASED family defers, defer it whole. "SUSPENDED as maintenance mode" is an ops concern (stop serving), not domain state. A one-state enum is not a lifecycle.

### L6 — LOW — Per-tenant uniqueness scopes must be mechanically re-derived

Every frozen "unique within tenant" law collapses to deployment-wide uniqueness (Dossier stable key within tenant+type → within type; DocumentType code per tenant → per deployment; numbering scopes; EvidenceType codes; Area names; Group names; `UNIQUE(issuer,subject)` becomes deployment-global). B2–B5 must sweep these deliberately, not by find-and-replace accident.

---

## 4. SC-R1 … SC-R11 dispositions

### SC-R1 — one company per deployment — **APPROVE**

Changed operator requirement is material reopen evidence; external sources confirm stamps are first-class architecture; structural inversion (§2.3) shows the pooled premise, not the domain, produced the tenancy laws. Forks-as-delivery correctly forbidden (this, per AWS silo guidance, is the real silo failure mode). A second customer triggers a deployment-economics review, not automatic pooling — correct. **Fix rider:** M1 handshake becomes part of the stamp model's definition.

### SC-R2 — meaning of `Tenant` — **APPROVE option A (retain singleton semantic root), redefined** — see §5

Option C (remove; deployment config only) is **rejected**: real durable consumers exist — operator-mutable company settings (transactional, audited, UI-edited data, not env config: `tenant.settings.manage`), company display identity, company-wide Authorization scope target, future Keycloak-Organizations projection anchor, and M1's handshake/backfill anchor. Option B (rename) is rejected for now on churn/blast-radius grounds (§5), with a named rename trigger.

### SC-R3 — tenant-qualified PK/FK law — **APPROVE removal; id-only UUID PK/FK**

With one company per database, `(tenant_id, id)` degenerates to a constant prefix and same-Tenant composite FKs prove a tautology — pure accidental complexity by the Method's definition. Target: `id UUID PRIMARY KEY`, ordinary typed FKs. No business relationship was found whose meaning requires the company root identity on every row under one-company-per-deployment (the wrong-attachment concern is real but is a *deployment-identity* property, closed by M1 at one row instead of every row). Survives unchanged: UUID technical identity, business-ID-never-PK, cross-owner FK authority-neutrality + `RESTRICT/NO ACTION` closed action set, no polymorphic registry, within-owner cascade restraint. The packet's prohibition on reflex `company_id`/`deployment_id` columns is correct and binding.

### SC-R4 — tenant RLS / tenant context — **APPROVE removal**

Rejection of counterargument in §2.4(2). RLS was promoted as Tenant isolation only; the isolated boundary no longer exists inside the database. Fail-closed tenant GUC context guards a distinction without a difference. Area/role RLS compensation correctly forbidden (would smuggle canonical AuthZ into the DB against frozen law). Survives independently: non-owner/NOSUPERUSER serving role, non-serving maintenance trust surface, DB constraints/triggers as invariant backstop (L3).

### SC-R5 — Keycloak posture — **APPROVE**

One realm/trust domain per deployment, one client, Metal Nobre users; Keycloak Organizations not required V1 — consistent with GCR-R1, which already rejected realm-per-Tenant and demoted Organizations to a possible future AuthN projection. Future enterprise federation (e.g., Metal Nobre AD) is realm-level IdP federation inside the same topology — no product change. Guard that survives: `issuer` remains explicit stored data in the binding (never a hardcoded constant), so IdP migration/federation stays open. No code-level provider assumption blocking productization was found; each future stamp carries its own provider configuration.

### SC-R6 — AuthenticationSubjectBinding / ApplicationSession — **APPROVE**

Remove the tenant dimension from binding/session uniqueness and drop tenant-first routing/tenant selector (which existed only as B2 open questions, never designed). Uniqueness becomes deployment-global `UNIQUE(issuer, subject)`. Independently justified and retained: stable issuer+subject identity, opaque application Session, no AuthZ snapshot in session, anti-corruption contract, no email auto-binding. One User ↔ one-or-many bindings remains a genuine **single-company** B2 decision (e.g., local-subject → federated-subject migration) and must not be collapsed by this reopen.

### SC-R7 — company-wide Authorization scope — **APPROVE retention; do not rename now**

The typed scope pair survives: company-wide scope ≠ AreaScope is real semantics (`tenant_owner`, `organization.manage`, `access.manage`, type management are company-wide). Removing it would break the frozen RoleAssignment shape for zero gain. `TenantScope` the *token* follows the §5 noun ruling (retain, redefined as whole-company/root scope). Representation simplification (scope kind constant; no tenant_id payload) is a B2 detail. Five frozen roles and additive/default-deny grants untouched.

### SC-R8 — Tenant lifecycle / deletion / erasure — **APPROVE deferral WITH M2**

ACTIVE/SUSPENDED/ERASED, TenantDeletionRequest, TenantErasureRecord, erasure tombstones and restore-reconciliation defer as SaaS customer-lifecycle machinery (L5: defer whole). Company decommission = operations. **Mandatory rider M2:** data-subject erasure obligations (PII-minimized Audit skeleton proof, separately-erasable enrichment, GCR-R4 DEK trigger) re-anchor to user-level erasure/offboarding and remain binding on B2/B6. Explicitly not weakened (verified against ledger §12): disposition, RetentionBinding, LegalHold, backup/restore, user offboarding, lawful Artifact deletion, PII minimization.

### SC-R9 — Tenant Portability Export — **APPROVE deferral**

Backup (DR), Governed Subject Export (per-subject packages), PUBLISH_COPY (connector) are distinct contracts with live V1 consumers — all KEEP, including the authorization-safe fail-closed completeness law (ledger §13), which is export semantics, not tenancy. Tenant Portability Export's consumer was customer exit/migration between products — no current consumer; stamp-to-stamp moves are whole-deployment backup/restore of an identical schema. Interchange is not orphaned (retains Historical Migration + GSE + IMPORT/PUBLISH_COPY). Reinstatement triggers: contractual portability demand; cross-stamp migration exceeding backup/restore; product-exit obligation.

### SC-R10 — background work / routing — **APPROVE**

Tenant enumeration (discovery shape 1) and the tenant_id routing field in claimable intents exist solely to select among coexisting customers — remove. B1 §9.10's two-shape law collapses to: claim due mechanism state from this deployment's DB → execute under canonical application AuthZ/system execution rules. Independently essential and retained: transactional outbox, durable intent in same commit, idempotent consumers, lease/retry/DLQ, external-effect truth, routing-metadata-only claim surfaces (minus tenant_id). R10-D still owns execution mechanics.

### SC-R11 — tenant-namespaced object-store keys — **APPROVE reopen/remove as invariant** (L4)

The prefix was defense-in-depth for a shared store serving multiple customers; a stamp's store is deployment-scoped. Keys remain opaque and immutable (that IS the invariant — provider location never business identity, keys never overwritten). Layout becomes an R10-C implementation freedom; per-stamp bucket/account isolation is ops configuration. Do not overconstrain R10-C in either direction.

---

## 5. Tenant: retain / rename / remove

**RETAIN the durable singleton root, keep the noun `Tenant`, redefine it. Do not rename now; never partially rename.**

Comparison the packet requires:

| Option | Verdict | Reasoning |
|---|---|---|
| A — retain singleton semantic root named `Tenant` | **CHOSEN** | Real consumers survive (settings, company identity, company-wide scope, M1 handshake anchor, future AuthN-projection anchor). Under the stamp model the noun is not even wrong: each deployment serves exactly one tenant of the product (Azure's own framing — "a stamp may serve one tenant"). One consistent vocabulary; zero churn in frozen texts beyond the semantic redefinition. |
| B — rename to Company/Organization root | REJECTED now | The frozen vocabulary is `Tenant`-saturated (`tenant_owner`, `tenant.settings.manage`, TenantScope, tenant-scoped types). A full rename reopens frozen R9 catalog wording across every authority for aesthetic precision; a partial rename creates two vocabularies for one concept (L2). Also `Organization` already names a bounded context — an `Organization` entity inside the `Organization` BC adds confusion, not clarity. |
| C — remove; deployment configuration only | REJECTED | Config is not transactional, not audited, not operator-editable product data. Company settings/identity are durable domain state. Removal also deletes M1's anchor and the future pooled backfill anchor. |

**Binding redefinition to promote:** *`Tenant` is the single company/organization root of a deployment. Exactly one ACTIVE Tenant root exists per deployment (structurally enforced, L1). It is a semantic root and scope target only — never a database partition dimension in V1.*

**Rename trigger (recorded, not hidden):** if commercialization produces a real vocabulary defect — e.g., a pooled model is chosen and "Tenant-as-company-root" collides with "tenant-as-isolation-partition," or customer-facing language forces it — rename wholesale in one deliberate sweep at that design stage.

**Anti-inertia guard:** the redefinition must state that B2+ designers may not infer pooled machinery (partition columns, RLS, routing) from the noun. The defect class "agent sees `Tenant`, mechanically adds tenant_id/RLS" is real in this repo's history; the guard fires at design review.

---

## 6. Hard questions — compact answers

Packet §8, numbered as there:

1. **Reduces or moves complexity?** V1: strictly reduces (fleet doesn't exist; one deployment). At fleet scale ops cost appears with N stamps — priced into the reopen triggers (§11), not hidden.
2. **B1 laws surviving independently:** full list in §2.3 "would still exist."
3. **Removing tenant PK/FK safe?** Yes; Tenant's genuine semantics are root/scope, not per-row data ownership (SC-R3). Wrong-attachment property re-homed by M1.
4. **Singleton Tenant useful or leftover?** Useful — real consumers (§5); leftover only if C had been chosen.
5. **Future migration cost acceptable?** Yes — asymmetry argued in §2.4(1); the durable root UUID (M1) makes any future backfill well-defined. Must be recorded as an accepted, triggered risk in the decision record.
6. **Cheaper seam than tenant columns?** Yes — the seam IS: stamp model + common build/migrations + durable root UUID + config pin. Columns-everywhere was the expensive seam for one specific unchosen future.
7. **RLS removal weakens reachable defense?** No — §2.4(2).
8. **Non-tenant RLS for another property?** None found; no current target law uses RLS for anything but tenant isolation, and inventing new RLS uses would violate frozen AuthZ boundaries.
9. **Keycloak Organizations unnecessary without harming federation?** Yes — SC-R5; federation is realm-level IdP config.
10. **AuthN binding/session tenant dimension?** Not needed — SC-R6.
11. **Company-wide root scope?** Still required — SC-R7.
12. **Tenant lifecycle real or SaaS machinery?** SaaS machinery V1 — SC-R8, with M2 rider.
13. **Backup/restore tombstone single-company?** No tenant-level tombstone needed; wrong-restore protection is M1's handshake; user-level erasure reconciliation after restore belongs to the re-anchored M2 family (restore must not resurrect erased data-subject PII — B6/R10-C obligation).
14. **Portability deferrable?** Yes — SC-R9.
15. **Routing laws vs async laws:** SC-R10 split confirmed.
16. **Key namespacing:** dead prefix as law; harmless as layout — SC-R11.
17. **Clean path to second customer?** Yes: new stamp = same artifact + own config/realm/store/DB + M1 identity pin. No code change.
18. **Concrete pooled triggers:** §11.
19. **Any frozen fact semantically requiring coexisting Tenants in one DB?** None found. Sweep performed: all "per-tenant" uniqueness collapses cleanly (L6); no cross-tenant product feature exists anywhere in R3–R9.5 (no cross-tenant dedup, sharing, analytics — all were already non-goals).
20. **Hardcoded Metal Nobre optimization?** None proposed; §7 seams (data/config-only company identity) are sufficient and binding. "Metal Nobre" must appear only as configured data.
21. **Other SaaS/fleet machinery in R3–R10:** §8 list — nothing beyond the packet's set except the items named there (permissions M3, realm-per-tenant deliberations, per-Tenant maintenance iteration, B1 routing tenant_id field, erasure-seam wording).
22. **Removal confusing essential with SaaS complexity?** Two near-misses caught: M1 (deployment-identity property hidden inside per-row tenancy) and M2 (data-subject privacy obligations hidden inside tenant erasure). With those fixes, no remaining proposal removes essential complexity.

---

## 7. Subtractive pass

**Deleted by this rebaseline without weakening any business/security/audit/retention/productization property** (given M1+M2+M3):

```text
(tenant_id, id) composite PKs and same-Tenant composite FKs
ENABLE + FORCE RLS + tenant-context GUC + fail-closed context plumbing
per-request/per-worker tenant seeding (SeedTxTenant class of machinery)
tenant enumeration + tenant_id routing metadata in async discovery
Tenant ACTIVE/SUSPENDED/ERASED product lifecycle
TenantDeletionRequest / TenantErasureRecord / erasure tombstones / restore reconciliation (tenant-level)
Tenant Portability Export (contract + Interchange process line)
tenant.export + tenant.deletion.request permissions
tenant-namespaced storage keys as isolation invariant
realm-per-Tenant deliberation; tenant selector / tenant-first routing UX questions
per-Tenant iteration language in maintenance-surface law
NOBYPASSRLS wording (vacuous without RLS)
```

**Duplicate-authority check:** the rebaseline creates no second authority; it removes a partition dimension. One risk was found and closed: partial rename would create dual vocabulary (L2).

---

## 8. Resulting reopen sets

### R9.5 (bounded amendments to the frozen ledger)

```text
§1  permission catalog: 29 → 27 base (drop tenant.export, tenant.deletion.request);
    tenant_owner bundle updated; TenantScope redefinition note        (M3, SC-R7)
§6  Tenant lifecycle/erasure family → DEFERRED (SaaS lifecycle);
    PlatformOperator/SystemPrincipal law unchanged;
    data-subject erasure / PII-minimization / audit-skeleton / GCR-R4
    trigger re-anchored to user-level erasure + offboarding            (M2)
§9  storage: tenant-namespaced key law → opaque immutable keys, layout free (SC-R11)
§13 Tenant Portability Export → DEFERRED; Backup/GSE/PUBLISH_COPY and
    completeness law unchanged                                          (SC-R9)
North Star, §2–§5, §7–§8, §10–§12, §14–§16: UNCHANGED
```

### R10-A (fact-list amendments; **no topology change** — 8+3 stands)

```text
Organization fact list: Tenant root (redefined singleton) + settings + Area/User/
  Group/GroupMembership retained; lifecycle/deletion/erasure/tombstone facts deferred
Coordination seam #9 "Tenant erasure/restore" → re-anchored to user/data-subject
  erasure + deployment decommission (ops)
Interchange fact list: Tenant Portability Export process line deferred
```

### R10-B1 (MATERIAL STRUCTURAL REOPEN — the packet's own classification, confirmed)

```text
§9.1  identity law → id UUID PRIMARY KEY; singleton Tenant root; NEW deployment↔DB
      identity handshake law (M1); business-ID-never-PK unchanged
§9.2  same-Tenant composite FK law → ordinary typed FKs; cross-owner
      authority-neutral RESTRICT/NO ACTION law UNCHANGED
§9.5  tenant isolation stack → removed V1; RLS removed; least-privilege serving
      role retained as ordinary DB security (L3)
§9.6  claim-surface metadata: drop tenant_id; durable-intent-in-same-commit law UNCHANGED
§9.7  maintenance trust surface retained; per-Tenant iteration + NOBYPASSRLS wording dropped
§9.10 discovery → direct due-work claim in this deployment's DB
§9.13 proof obligations: drop RLS-negative/composite-census proofs; ADD handshake
      fail-closed proof + singleton-root structural proof
§9.14 reopen triggers: ADD pooled-tenancy trigger set (§11)
UNCHANGED: §9.1 DB topology/schema/provider non-atomicity, §9.3 primitives,
§9.4 persistence classes + mutation law, §9.8 transaction law, §9.9 same-commit
Audit/intent, §9.11 package map, §9.12 namespace/migration provenance
```

### R10-B2 (scope rederivation — currently NEXT, pause confirmed)

```text
DROP from B2 scope: tenant dimension in binding/session uniqueness; Tenant
  lifecycle ACTIVE/SUSPENDED/ERASED; TenantDeletionRequest/TenantErasureRecord/
  tombstone/restore-reconciliation state; same-Tenant FK + RLS application sections
ADD to B2 scope: singleton Tenant-root representation + structural singleton
  enforcement (L1); deployment↔DB handshake consumer surface (M1); deployment-wide
  uniqueness re-derivation (L6); User lifecycle/offboarding + user-level erasure
  hooks re-anchored per M2
UNCHANGED in B2: provider binding/reconciliation choreography; Session/assurance;
  anti-corruption proof; Area/User/Group/GroupMembership; Permission/Role/
  RoleAssignment; company-wide + Area typed scopes; grant evidence + evaluation
  read surface; persistence/mutation classification; tx boundaries; same-commit
  Audit/intent points; all B2 "does not design" exclusions
R10-D: drop cross-customer routing assumptions (SC-R10)
R10-E: no tenant selector/company switching; provider-hosted journeys unchanged
R10-C: key-layout freedom (L4); restore correctness gains M1 handshake +
  re-anchored M2 reconciliation
```

### B2 restart condition

B2 reopens only after the amended B1/R9.5/R10-A texts are promoted (§12).

---

## 9. Other customer-multitenancy machinery found (beyond the packet's list)

```text
tenant.export / tenant.deletion.request permissions            → defer (M3)
B1 §9.6 tenant_id field in claimable routing metadata          → drop (SC-R10)
B1 §9.7 per-Tenant iteration maintenance language              → drop (L3)
B1 §9.10 two-shape tenant discovery law                        → collapse (SC-R10)
GCR-R1 realm-per-Tenant deliberation                           → moot inside a stamp;
                                                                  survives as stamp-level note
R10-A seam #9 tenant-erasure coordination wording              → re-anchor (M2)
Interchange "Tenant Portability Export package process"        → defer (SC-R9)
handoff/B2 checklist tenant-lifecycle lines                    → rederive (§8)
```

Nothing else in R3–R10 was found to encode customer-fleet machinery. Explicitly NOT reopened (verified no material contradiction): modular monolith; 8+3 topology; Keycloak selection; Document/Revision/WorkingContent/Submission; Approval/SoD; Artifact identity; ManagedArtifactStore; malware gate; Dossier/Evidence; Retention/Hold/Disposition; Audit ownership; Historical Migration; Distribution; Search/Notifications classifications.

---

## 10. Proposed for deletion but MUST survive

```text
1. Durable singleton company root entity                       (vs SC-R2 option C)
2. Deployment↔database identity binding property               (M1 — new handshake law)
3. Company-wide Authorization scope                            (SC-R7; candidate agrees)
4. Non-owner/NOSUPERUSER serving role + non-serving
   maintenance trust surface                                   (L3; candidate agrees)
5. Data-subject erasure / PII-minimized Audit skeleton /
   separately-erasable enrichment / GCR-R4 DEK trigger         (M2 — re-anchored)
6. Export authorization-safe completeness law                  (rides in §13; not tenancy)
7. Backup/restore correctness incl. restore not resurrecting
   erased data-subject PII                                     (M2/R10-C)
8. Opaque immutable provider keys / keys never overwritten     (SC-R11 keeps the real invariant)
```

---

## 11. Concrete pooled/shared-tenancy reopen triggers

A second customer alone triggers only a **deployment-economics review**. Pooled/shared tenancy (or DB-per-customer on shared backend) reopens on measured evidence:

```text
1. fleet economics: stamp count × (infra + ops + upgrade) cost measurably exceeds
   sustainable margin for the actual customer segment (e.g., many small low-
   utilization customers where a stamp is structurally oversized)
2. operational scaling failure: upgrade/patch/backup orchestration across stamps
   demonstrably exceeds operations capacity despite automation
3. product requirement for a genuine cross-customer plane: central analytics,
   cross-company features, shared control plane, aggregated reporting
4. commercial motion requiring instant self-serve signup where stamp provisioning
   latency/cost is a proven conversion blocker
5. a real customer/compliance constraint contractually requiring a specific
   shared-tenancy or pooled hosting model
```

On any trigger: a deliberate design stage selects among pooled / DB-per-customer / stamps / hybrid with then-current evidence; the durable root UUID (M1) is the defined backfill anchor if pooling wins. The trigger set must be embedded in the amended B1 §9.14 (not only in review evidence).

---

## 12. Promotion conditions and review posture

**Is another broad review required?** **NO.** The GCR just closed; this reopen is bounded and the packet's §6 no-reopen fence held under attack (no material contradiction found outside it). One **bounded delta review** of the amended authority texts suffices, verifying: (a) M1/M2/M3 incorporated; (b) no tenant-law residue contradicting the new substrate; (c) no scope creep beyond §8's reopen sets.

**Exact promotion conditions:**

```text
1. Operator adjudicates SC-R1..SC-R11 dispositions including the three MAJOR fixes
   (accept / modify with reasons under the Method).
2. One adjudicated corrected-target amendment updates, in the same change set:
   - ledger §1/§6/§9/§13 bounded amendments (M2, M3, SC-R8/R9/R11);
   - r10-technical-architecture B1 §9.1/9.2/9.5/9.6/9.7/9.10/9.13/9.14 + R10-A
     fact lists/seam #9 + rewritten B2 scope (§8 above);
   - program authority + handoff mirrors re-routed accordingly;
   - the Tenant redefinition + singleton law + M1 handshake law recorded in B1;
   - decision recorded as RESTRUCTURE NOW (design baseline) with §11 triggers.
3. Bounded delta review (cold) of the amended texts returns APPROVE.
4. Operator ratifies; handoff flips B2 to rederive-from-new-substrate.
5. Implementation gate remains CLOSED throughout.
```

---

## 13. Convergence statement

Finding count: 3 MAJOR + 6 LOW, zero BLOCKER, zero architecture-altitude contradiction with the candidate's structure. All MAJORs are property-preservation riders on an approved direction, not structural objections. Altitude is descending (substrate wording, catalog hygiene, one new one-row mechanism). **Stop condition met for this review round.**

```text
VERDICT = APPROVE SINGLE-COMPANY DEPLOYMENT/TENANCY REBASELINE WITH MATERIAL FIXES
```
