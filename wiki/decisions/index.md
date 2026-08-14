# Decisions

> **Last verified:** 2026-08-14
> **Status:** ADR archive + retained-decision register during Cohesive Platform Redesign

## Active redesign rule

Historical ADRs remain valuable evidence, but **Accepted does not mean automatically retained in the new target** when the active Cohesive Platform Redesign is explicitly re-adjudicating that boundary.

Current target authority:

- [../architecture/cohesive-platform-redesign.md](../architecture/cohesive-platform-redesign.md)
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`

For target-design conflicts, operator-approved decisions in the active ledger win until replacement/amending ADRs are promoted at design closure.

Do not edit dozens of historical ADR status blocks during WIP design. This index is the **program-level reclassification layer** until the final ADR set is ready; at closure, affected ADRs will be formally amended/superseded or retained.

## Retained cross-cutting decisions / principles

These remain useful target constraints unless a later material finding explicitly reopens them:

- [0001 — Eigenpal adoption](0001-eigenpal-adoption.md) — editor technology choice; editor does not own governance semantics.
- [0009 — PDF dispatch transactional outbox](0009-pdf-dispatch-outbox.md) — async delivery pattern remains a valid infrastructure precedent.
- [0012 — Contract-first API](0012-contract-first-api.md) — OpenAPI/generated-contract discipline remains.
- [0014 — docx-renderer service naming/responsibility](0014-rename-docgen-v2-to-docx-renderer.md) — renderer remains supporting infrastructure.
- [0021 — tenant vs platform admin separation](0021-tenant-vs-platform-admin-separation.md) — consistent with new `tenant_owner` vs future platform operator distinction.
- [0025 — RFC 9457 error envelope](0025-error-envelope-rfc9457.md) — stable API behavior constraint.
- [0027 — RLS adoption](0027-rls-adoption-sequencing.md) — tenant isolation defense-in-depth remains; final table set changes with the new data model.
- [0034 — integration test fixture framework](0034-integration-test-fixture-framework.md) — engineering/test infrastructure, not product-domain authority.
- [0036 — DB tracing](0036-otelsql-db-tracing.md) — observability infrastructure.

## Retained concept, placement/semantics under redesign

- [0085 — Release Coordinator](0085-release-coordinator-approval-driven-publication.md) — **concept retained**: human approval does not directly publish; release/effectivity waits for mechanical/domain gates. Final placement and lifecycle contract will be re-specified with Controlled Information.
- [0093 — Controlled Information context / template role](0093-controlled-information-context-template-as-role.md) — **direction incorporated but being amended by the active redesign**: one Controlled Information context survives; template-as-role survives and is now explicitly versioned only through DocumentRevision; the target stable noun is converging on `Document`, not a separate public `ControlledDocument` object.
- [0015 — async freeze/pin/materialize](0015-async-freeze-pin-materialize.md) — async principle is useful, but the content-source contract must be re-specified so only the reviewed Revision can be frozen/renditioned.
- [0069 — periodic review / reason-for-change](0069-document-periodic-review-and-reason-for-change.md) — product requirement retained for design; exact owner/data model/permissions will be re-specified.

## HISTORICAL / target semantics superseded or reopened

These files remain in Git/wiki as evidence but must not drive new implementation:

- [0007 — Two-Tier Authorization](0007-two-tier-authz.md) — enforcement-depth lessons remain, but old grant/role/scope semantics are superseded by the active scoped-RBAC + Groups redesign.
- [0022 — AuthZ capability coherence](0022-authz-capability-coherence.md) — historical implementation program; old role vocabulary and capability surfaces are not target authority.
- [0077 — Approval Delegation](0077-approval-delegation.md) — sophisticated delegation is not a V1 requirement; audited reassignment is the approved baseline.
- [0081 — Per-Profile Signature/GovernanceClass route policy](0081-per-profile-signature-policy.md) — reopened. `GovernanceClass` survives only if independent business meaning is proven; ApprovalPolicy owns workflow shape.
- [0082 — Approval kernel extraction](0082-approval-kernel-extraction.md) — the separation of Approval as an authority survives, but its historical internal engine/model is superseded by Approval V1.
- [0087 — livre zero-stage route](0087-livre-configured-no-approval-route.md) — historical route-shape solution; the new design does not encode no-approval through a fake zero-step route unless the current redesign independently chooses that model.
- [0092 — AuthZ grant-model unification](0092-authz-grant-unification.md) — historical precursor. The new design keeps one assignment concept but changes the role set, scope semantics, Group treatment and removes speculative/custom-role assumptions from V1.

## All other ADRs

Treat ADRs not explicitly classified above as **historical/current-implementation evidence pending revalidation** when they touch product nouns, lifecycle, permissions, workflow, document/template behavior or module ownership.

Pure infrastructure/contract/engineering ADRs may continue to apply when they do not conflict with the new domain model.

At integrated-design closure we will produce the exact formal ADR amendment/supersession map rather than leave this WIP classification as the permanent record.
