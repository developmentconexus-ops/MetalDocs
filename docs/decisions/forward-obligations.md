---
id: forward-obligations
kind: authority
owner: architecture
summary: Preserves the unresolved, deferred, and proof-backed future obligations that remaining stages must consciously consume without restoring legacy implementation.
---

# Forward decision obligations

This page is the compact durable successor to the pre-reset Rebaseline Decision Registry **only for decisions that still carry forward work or constraints**.

It deliberately does **not** copy the old registry's `CURRENT` rows: current truth lives in the semantic Product/T1→T8 authorities plus bounded current decisions routed by `docs/decisions/index.md`. It deliberately does **not** copy `SUPERSEDED` rows: superseded implementation has been removed and receives no inheritance right.

Source provenance: PR #131, `wiki/architecture/rebaseline-decision-registry.md`, commit `d8b1c6d31e704e9552a14faa7764c634a29b081d`.

## Consumption law

Before every remaining architecture stage:

```text
read current owning authorities
→ consume PRESERVE obligations as baseline evidence unless materially disproved
→ deliberately decide the stage-relevant REOPEN obligations
→ keep DEFERRED obligations as future seams/counterexamples only
→ never create dormant implementation for a deferred capability
```

A later ratified authority wins over an older forward-obligation wording and must update this page when it closes or materially refines that obligation.

## Consumed during T11 bounded reopen

```text
ASY-02
  prior: DEFERRED — Notifications had no concrete Launch consumer
  trigger: stable-Document @Mention + persistent in-app Inbox became operator-required Launch V1
  current disposition: consumed / refined into current Launch authority
  authority: discussion-notifications-launch.md
```

ASY-02 is therefore no longer a forward DEFERRED obligation and is excluded from the counts below. Email/push/preferences remain deferred inside the new bounded authority; only persistent in-app Mention Notifications/Inbox are current Launch scope.

## PRESERVE — 21

- **AUTH-02 — PRESERVE** — Keycloak remains selected V1 AuthN provider evidence unless concrete deployment evidence reopens it; it stays replaceable behind an anti-corruption seam.
- **AUTH-06 — PRESERVE** — No atomic MetalDocs↔identity-provider transaction; provider effects remain post-commit/reconciled when explicitly required.
- **ORG-08 — PRESERVE** — Area retirement/inactivity blocks future use as appropriate while preserving existing references/history.
- **CNT-14 — PRESERVE** — EigenPal may remain selected DOCX adapter/provider evidence; it never owns semantic truth.
- **ASY-04 — PRESERVE** — Use one PostgreSQL-backed durable-job mechanism; River remains selected/reference implementation and replaceable without changing semantic meaning.
- **LP-02 — PRESERVE** — Future Group audience resolution must freeze concrete Users so membership drift never rewrites a historical denominator.
- **LP-03 — PRESERVE** — Future acknowledgement requires an explicit AcknowledgementRecord; view/download never silently equals acknowledgement.
- **FUT-04 — PRESERVE** — A future Dossier link never grants access.
- **FUT-05 — PRESERVE** — LegalHold business preservation is not ObjectLock/provider physical enforcement.
- **FUT-06 — PRESERVE** — Do not build a generic Record/BPM/object platform without changed requirements.
- **MIG-05 — PRESERVE** — Plan/dry-run, deterministic outcomes, reconciliation, idempotency, and atomic semantic import units remain strong cutover evidence if a real migration appears.
- **MIG-06 — PRESERVE** — `CURRENT_STATE` / `FULL_HISTORY` remain migration-mode evidence only; actual source completeness decides any future mode set.
- **DB-01 — PRESERVE** — One PostgreSQL product-state database remains the technical default absent a real distributed trust/scale boundary.
- **DB-02 — PRESERVE** — One database namespace remains a mechanism default; database namespace is not semantic ownership.
- **DB-03 — PRESERVE** — Use opaque UUID technical IDs; business/provider identities never become primary-key authority.
- **DB-04 — PRESERVE** — Use `TIMESTAMPTZ` for trusted business instants.
- **DB-05 — PRESERVE** — Prefer typed FKs and closed unions; avoid universal polymorphic business relations except genuinely generic semantics.
- **DB-06 — PRESERVE** — Cross-owner FKs do not mutate another owner's state through `CASCADE`/`SET NULL`; `RESTRICT`/`NO ACTION` is the safe default.
- **DB-07 — PRESERVE** — JSONB is not an unmodeled-state escape hatch; use it only for bounded snapshots/provenance where variability is semantically justified.
- **DB-10 — PRESERVE** — One local transaction may compose multiple semantic owners when one invariant/business transition requires it.
- **SEC-01 — PRESERVE** — Platform/operator/system principals receive no implicit company-content access; maintenance trust surfaces must be explicit and non-serving.

## REOPEN — 3

- **CNT-03 — REOPEN** — EditorSession is not a correctness dependency; a bounded editor lease may be added only if concrete UX/integration evidence requires it.
- **AUD-06 — REOPEN** — No claim of indefinite/statutory Audit retention exists; a future Records/compliance requirement must define retention, pruning, and checkpoint semantics.
- **MIG-10 — REOPEN** — Detailed imported target families are not frozen; any future migration derives the smallest truthful shape from actual source evidence.

## DEFERRED — 27

- **AUTH-07 — DEFERRED** — Fresh-auth/eSignature evidence has no named Launch consumer; Authentication remains the future owner if promoted.
- **GOV-09 — DEFERRED** — Fresh-auth per governance Step.
- **GOV-10 — DEFERRED** — Bounded optional per-Step deadline truth is current under `governance-step-deadline.md`; SLA breach consequences, deadline extension and escalation remain deferred.
- **GOV-11 — DEFERRED** — Generic reassign/overseer/delegation; current Launch uses only bounded explicit exits.
- **GOV-12 — DEFERRED** — Anchored selected-range governance review, if promoted, binds to the exact immutable governed snapshot through a provider-neutral anchor; it never mutates Submission or silently remaps to returned DRAFT, and tracked-change/suggestion requires separate semantics. Current seam authority: `governance-review-layer-seam.md`.
- **DOC-10 — DEFERRED** — DocumentType category/taxonomy platform.
- **DOC-11 — DEFERRED** — Editable Dictionary/System Value platform.
- **CNT-12 — DEFERRED** — Structured TemplateSpec platform.
- **CNT-13 — DEFERRED** — DRAFT EditorialComment platform; stable-Document Discussion is current but does not alter the deferred DRAFT/editor-comment concept.
- **CNT-15 — DEFERRED** — Realtime Yjs/CRDT; the seam remains WorkingContent concurrency.
- **REL-04 — DEFERRED** — Scheduled/future-dated Release.
- **OBS-05 — DEFERRED** — Separate obsolescence route; current Launch reuses the DocumentType governance route.
- **OBS-06 — DEFERRED** — Reactivation of an OBSOLETE document.
- **AUD-05 — DEFERRED** — Generic Audit export permission/capability.
- **STO-09 — DEFERRED** — Application-layer Company DEK/crypto-erasure is not a default absent a named target-data/assurance requirement.
- **STO-19 — DEFERRED** — Whole-Submission canonical JCS/composite digest until a named signing/export/non-repudiation consumer exists.
- **LP-01 — DEFERRED** — Distribution / Read & Acknowledge attaches to Release + User/Group and never becomes AuthZ/effectivity authority.
- **LP-04 — DEFERRED** — Periodic Review attaches to stable Document + exact current EFFECTIVE Revision; due/overdue never silently changes effectivity.
- **LP-05 — DEFERRED** — Detailed PeriodicReviewPolicy/Record schema and outcomes.
- **FUT-01 — DEFERRED** — Dossier is documentary context, never content owner/access grant; stable Document is the seam.
- **FUT-02 — DEFERRED** — Evidence may own its own exact descriptor + managed-content handle with an independent lifecycle; no Artifact owner is required.
- **FUT-03 — DEFERRED** — Retention/Hold/Disposition stay separate from document lifecycle/provider storage; expiry does not imply delete.
- **MIG-08 — DEFERRED** — Governed Subject Export may derive future manifests/digests from semantic descriptors without provider identity becoming product truth.
- **MIG-09 — DEFERRED** — External Repository IMPORT/PUBLISH remains future and reuses exact-content copy/admission seams without making external object IDs MetalDocs identity.
- **SEC-05 — DEFERRED** — In-Area document confidentiality is decided as a future seam in `document-confidentiality-seam.md`: a governed ConfidentialityClass plus clearance grants, never a per-document ACL and never Area used as a secrecy mechanism. Launch implements no dormant capability.
- **SEC-04 — DEFERRED** — Pooled/shared multi-customer tenancy is not Launch; stable Company identity preserves the seam.
- **SEC-05 — DEFERRED** — Customer-company lifecycle/portability deletion/export is not Launch.
- **SEC-06 — DEFERRED** — Generic eDiscovery/PKI/TSA/HSM/signature/quarantine platform absent a concrete requirement.

## Count proof

```text
PRESERVE  21
REOPEN     3
DEFERRED  27
TOTAL     51
```

These are the current cross-stage forward obligations. Most preserve pre-reset registry decisions; `GOV-12` was added by the operator-ratified T11 B06-F2 future seam. Current semantic decisions live in their owning authorities; superseded legacy decisions remain deleted.