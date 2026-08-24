---
id: t11-b01-b09-lock-evidence
kind: evidence-locator
owner: architecture
summary: Durable locator for exact operator-LOCKED frontend P8 evidence preserved outside the merge candidate while T11 remains open.
---

# T11 B01–B09 LOCK Evidence Locator

> **Status:** ACTIVE DURABLE EVIDENCE LOCATOR.  
> **Scope:** operator-LOCKED frontend P8 evidence through B09 only.  
> **T11:** remains OPEN.  
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Why this exists

MetalDocs repository governance requires every merge candidate to contain no `docs/work/**` tree. Frontend planning, however, must preserve the exact operator-LOCKED P8 browser artifacts so later P11 assembly can re-use and fidelity-check the actual accepted interaction evidence rather than a prose reconstruction.

Therefore the B01–B09 checkpoint preserves the pre-cleanup planning tree by exact Git identity and records the canonical P8 blobs here before temporary work is removed from the merge candidate.

This file is an Evidence locator. It is not Product authority, a second roadmap, or permission to reopen a LOCK.

## 2. Preserved Git identity

```text
repository  developmentconexus-ops/MetalDocs
evidence ref evidence/t11-pr162-b01-b09-locks-20260824
exact commit adf58e448bc5bd3a20cae5b7228d729c031f94ac
source PR    #162
```

The Evidence ref must not be moved or deleted while T11/P11 still depends on these LOCK artifacts. The exact commit remains the authoritative locator even if a human-facing ref name later changes.

## 3. Canonical P8 LOCK artifacts

| Block | Canonical artifact path on exact Evidence commit | Git blob |
|---|---|---|
| B01 App Shell + Global IA + Home | `docs/work/current/t11-b01-app-shell-wireframe.html` | `6d4f5c25c50dd302e00614c67c9c54f3068bc751` |
| B01N Notification global chrome + Quick Inbox | `docs/work/current/t11-b01-notifications-wireframe.html` | `17dd35707e1820290de676a241a2faf6eb48004a` |
| B02 Library / Discovery | `docs/work/current/t11-b02-library-wireframe.html` | `3ac7217f1fd723f43e4feec38d22256994156189` |
| B03 Document Official | `docs/work/current/t11-b03-document-official-functional-wireframe.html` | `6bffa8c2bfbbff840d410385b69150e3bff52e79` |
| B04 Document Work | `docs/work/current/t11-b04-document-work-functional-wireframe.html` | `8da45d990f40921fdc0c484eef1d6e12698a90b1` |
| B05 My Work | `docs/work/current/t11-b05-my-work-functional-wireframe.html` | `0942dfb9189454b55f2f1eced547426dcbce65a0` |
| B06 Governance Case | `docs/work/current/t11-b06-governance-case-functional-wireframe.html` | `785473b822b35b1261925c7465dd5de4793dbdf1` |
| B07 Document History | `docs/work/current/t11-b07-document-history-functional-wireframe.html` | `20ec64d34085fbc9075b136a61e69c48c0cad981` |
| B08 Notifications Full Inbox | `docs/work/current/t11-b08-notifications-full-inbox-functional-wireframe.html` | `bb130535721b2381524763a4885ade5199a15596` |
| B09 Audit | `docs/work/current/t11-b09-audit-functional-wireframe.html` | `7daa6054851e617aeacb95a28d907d0d6d4bd3d6` |

These blob identities are immutable evidence bytes. P11 must consume the locked artifact matching the recorded blob, not a later reconstructed approximation.

## 4. Supporting planning Evidence

The same exact Evidence commit also preserves the block P7/P9/P10 records, bounded upstream findings/rebaselines, and falsification plans that were still located under `docs/work/current/` at checkpoint time.

For B09 specifically:

```text
P7 design                 docs/work/current/t11-b09-audit-r1.md
P7 exit                   docs/work/current/t11-b09-p7-exit.md
P8 realization plan       docs/work/current/t11-b09-p8-realization-plan.md
P9 Screen Contract        docs/work/current/t11-b09-screen-contract.md
  blob                    ece854b403e727ccc7b56cfbf482d49426138568
P10 Pattern Consolidation docs/work/current/t11-b09-pattern-consolidation.md
  blob                    8d8328be68b8ab06f78b6038f2dd676564c61a60
```

Reviewer transport remains non-authoritative Evidence and never enters `main`.

## 5. Durable meaning vs temporary Evidence

After PR #162 cleanup:

```text
main / durable docs
  Product meaning
  architecture decisions
  numeric census
  repository status
  this exact Evidence locator

Evidence ref @ exact commit
  rendered P8 artifacts
  detailed temporary block planning/proof records
```

The Evidence ref does not become a second current Product authority. If durable Product/architecture meaning conflicts with a temporary planning note, current accepted durable authority wins unless material Evidence opens the smallest owning Finding.

## 6. P11 retrieval law

When FP2/P11 later opens:

```text
read current accepted Product/architecture authority
→ read current roadmap
→ use this locator only for retained LOCK artifact identity
→ fetch exact P8 blobs from exact Evidence commit
→ assemble a new disposable P11 prototype
→ preserve each retained LOCK's protected structure
→ reopen only when integration materially falsifies it
```

Do not edit a locked blob in place. A material correction requires the normal smallest-owner reopen and a newly operator-LOCKED artifact identity.

## 7. Reopen / retirement trigger

Keep this locator and Evidence ref while any current T11/P11/P13/P14 proof relies on the listed artifacts.

It may be retired only after a later accepted authority explicitly proves that the exact artifacts are no longer required for frontend conformance/reconstruction. Repository cleanup preference alone is not a retirement trigger.
