---
id: t11-b11-lock-evidence
kind: evidence-locator
owner: architecture
summary: Durable locator for exact operator-LOCKED B11 P8 and converged post-LOCK proof preserved outside the merge candidate.
---

# T11 B11 LOCK Evidence Locator

> **Status:** ACTIVE DURABLE EVIDENCE LOCATOR.  
> **Scope:** B11 — Access Administration operator-LOCKED P8 + P9/P10 + final challenge proof only.  
> **T11:** remains OPEN.  
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Why this exists

MetalDocs merge candidates contain no `docs/work/**`, while later P11 assembly/fidelity proof requires the exact operator-LOCKED frontend Evidence to remain recoverable. B11 also carries post-LOCK P9/P10 and fresh final-challenge proof that must remain inspectable without promoting temporary planning notes into Product authority.

The complete clean-R3 planning tree is therefore preserved under one immutable Evidence ref before temporary work is removed from the acceptance candidate.

This locator routes Evidence only. It is not Product authority, a second roadmap or permission to reopen B11.

## 2. Preserved Git identity

```text
repository   developmentconexus-ops/MetalDocs
source       codex/t11-b11-clean-rebaseline
evidence ref evidence/t11-b11-clean-r3-locks-20260825
exact commit 75e00f343e49e7e0748c1a86a4f904e9ef5464f2
```

The remote ref resolves to the exact commit above. It MUST NOT move while current T11/P11/P13/P14 proof depends on B11.

## 3. Canonical B11 Evidence

| Evidence | Path on exact Evidence commit | Git blob |
|---|---|---|
| P8 functional LOCK artifact | `docs/work/current/t11-b11-access-administration-p8.html` | `ea20912e5259f4f3f51df7ce09ee3f2e5cfc7540` |
| P8 behavior | `docs/work/current/t11-b11-access-administration-p8.js` | `670ff9b905d94014ff27698e2a23c868316030a4` |
| P8 presentation | `docs/work/current/t11-b11-access-administration-p8.css` | `9ce012007613777187ae70956c2bfa09e7066c16` |
| Operator R3 LOCK record | `docs/work/current/t11-b11-p8-r3-operator-relock.md` | `57e410fb14513e2063c3cba01488b0a0cb3e6ecb` |
| P9 Screen Contract | `docs/work/current/t11-b11-screen-contract.md` | `1743710bf4c985bfa3bc4a7718ea33e34639f951` |
| P10 pattern consolidation | `docs/work/current/t11-b11-pattern-consolidation.md` | `9d36d38bb53a480a1dfa9a13ff1b3321c0dccacb` |
| Fresh final challenge R3 | `docs/work/current/t11-b11-final-challenge-r3.md` | `c77ca0e190268dc650dd9229b938943387acbfd1` |

The exact Evidence commit also preserves superseded R1/R2 records and the clean-rebaseline planning record. Those files retain challenge provenance only; the R3 identities above are canonical.

## 4. Protected B11 meaning

Durable Product/architecture authority remains in current Product, architecture and bounded decision owners. The LOCK protects the frontend structure proved by the exact P8/P9/P10 Evidence, including:

```text
stable /admin/access route
Por Área / Grupos / Funções local Access workspace
Subject(User|Group) × fixed Role × Scope(Company|Area)
visible server-page traversal without hidden collection crawl
raw op6 page fidelity; DISABLED visible but unavailable
unknown Group membership reconciled through op28 201/204
op32 same-key replay and ambiguous recovery with mutation 1 → 1
server-side op31 filtering before pagination
403 protected-surface replacement and selected-identity 404 reconciliation
no Group.area_id, custom Role editing, operation 90 or effective-access engine
```

If current durable authority conflicts with a temporary planning note, current durable authority wins unless material Evidence invokes the bounded reopen law.

## 5. P11 retrieval law

When FP2/P11 later opens:

```text
read current accepted Product/architecture authority
→ read current roadmap
→ use this locator only for B11 LOCK/proof identity
→ fetch exact R3 P8 blobs from the exact Evidence commit
→ assemble a new disposable P11 prototype
→ fidelity-check protected B11 structure
→ reopen only if integration materially falsifies the LOCK
```

Do not edit the locked P8 blobs in place. A material B11 correction requires the smallest normal frontend/upstream reopen and a newly operator-LOCKED Evidence identity.

## 6. Reopen / retirement trigger

Keep this locator and Evidence ref while any current T11/P11/P13/P14 proof relies on B11. Retirement requires later accepted authority explicitly proving the exact B11 Evidence is no longer required; repository-cleanup preference alone is not a retirement trigger.
