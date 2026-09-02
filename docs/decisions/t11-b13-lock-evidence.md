---
id: t11-b13-lock-evidence
kind: evidence-locator
owner: architecture
summary: Durable locator for the exact operator-LOCKED B13 P8 R4 plus P9/P10 proof and the FP2-F3 confidentiality study preserved outside the merge candidate.
---

# T11 B13 LOCK Evidence Locator

> **Status:** ACTIVE DURABLE EVIDENCE LOCATOR.
> **Scope:** B13 — Document Creation operator-LOCKED P8 R4 + P9/P10 proof + the FP2-F3 study that produced the confidentiality reopen.
> **T11:** remains OPEN.
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Why this exists

MetalDocs merge candidates contain no `docs/work/**`, while later P11/P12 assembly requires the exact operator-LOCKED frontend Evidence to remain recoverable. The complete B13 tree is preserved under one immutable Evidence ref before temporary work is removed from the acceptance candidate.

This locator routes Evidence only. It is not Product authority, a second roadmap or permission to reopen B13.

## 2. Preserved Git identity

```text
repository   developmentconexus-ops/MetalDocs
source       claude/repo-context-technical-design-69t84i
evidence ref evidence/t11-b13-p8-r4-locks-20260827
exact commit 4fda88d72827e6735a5a60b1ee4d085a374c0616
```

The remote ref resolves to the exact commit above. It MUST NOT move while current T11/P11/P12/P13/P14 proof depends on B13.

## 3. Canonical B13 Evidence

| Evidence | Path on exact Evidence commit | Git blob |
|---|---|---|
| P8 R4 functional LOCK artifact (single self-contained HTML) | `docs/work/current/t11-b13-document-creation-p8.html` | `0931da989b7e915a2d6c197cea7cefc848998115` |
| P6/P7 planning + FP2-F3 absorption + LOCK/P9/P10 record | `docs/work/current/t11-b13-document-creation-planning.md` | `6ee40e051c749db026f98b40e22d7454c64462ef` |
| FP2-F3 confidentiality Global-Maximum study (reference research + falsified alternatives) | `docs/work/current/t11-fp-f3-document-confidentiality-study.md` | `9f52b67ebd5d6134150a077d2e38cd92305a266c` |
| FP2 coherence alignment worksheet | `docs/work/current/t11-fp-coherence-alignment.md` | `2c031f17553bcfd1019d9665e4664bbcf4bec1a6` |

Unlike B10–B12, B13 carries no separate Screen Contract / pattern-consolidation files: P9 and P10 are recorded as §8 of the planning document above, because B13 introduced no new pattern — it consumes the canonical wireframe pattern already graduated into `../architecture/frontend.md` §19.

The Evidence commit history also preserves the superseded R1 (rejected — wrong representation medium), R2 and R3 candidates; the R4 identities above are canonical.

## 4. Protected B13 meaning

Durable Product/architecture authority remains in current Product, architecture and bounded decision owners (`content-format-vocabulary.md` for formats, `document-confidentiality-launch.md` + `document-confidentiality-seam.md` for confidentiality). The LOCK protects the frontend structure proved by the exact P8/P9/P10 Evidence, including:

```text
op44 progressive disclosure: complete non-paginated arrays, each region revealed only once
  the server can answer it; an unrequested list is never rendered as an empty collection
three start modes (blank / from template / upload a source file) as one radio decision,
  with the template region bound to eligible + effective-revision templates only
source-file attachment composed honestly as ops 59 → 60 → 58 AFTER op46, with named
  partial-failure recovery (upload failure, 410 expired, malware, invalid structure)
op46 idempotent creation with same-key ambiguous replay proving a 1 → 1 mutation
FP2-F3 confidentiality: server-projected class set limited to the author's own clearances,
  materialized default class, conjunctive audience rule, live revocation, total
  non-disclosure, and the governance consequence that routing does not confer read
non-identity: the allocated Document.code never carries the confidentiality class (proved)
result dialog as the canonical creation confirmation, carrying code, REV000, class, read
  consequence, Idempotency-Key and the B04 boundary
403 denial panel / 404 non-disclosure / visible server-page traversal throughout
```

Operator dispositions preserved by this Evidence: B13-Q1 resolved to hypothesis B (hypothesis A blocked by B13-F2 — op44 projects neither representation policy nor template source format); B13-F1 remains an open, non-blocking upstream finding; FP2-F3 Q1–Q5 ratified 2026-08-27 and realized in R4.

## 5. Proof recorded on the Evidence commit

```text
R3 regression suite   17 / 17 PASS
R4 FP2-F3 suite       16 / 16 PASS
                      -----------
total                 33 / 33 PASS   (Chromium headless)
unresolved material findings   0
application operations         89 unchanged by B13
```
