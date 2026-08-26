---
id: t11-b12-lock-evidence
kind: evidence-locator
owner: architecture
summary: Durable locator for exact operator-LOCKED B12 P8 and post-LOCK P9/P10 proof preserved outside the merge candidate.
---

# T11 B12 LOCK Evidence Locator

> **Status:** ACTIVE DURABLE EVIDENCE LOCATOR.
> **Scope:** B12 — Document Governance Administration operator-LOCKED P8 + P9/P10 proof only.
> **T11:** remains OPEN.
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Why this exists

MetalDocs merge candidates contain no `docs/work/**`, while later P11 assembly/fidelity proof requires the exact operator-LOCKED frontend Evidence to remain recoverable. The complete B12 planning tree is preserved under one immutable Evidence ref before temporary work is removed from the acceptance candidate.

This locator routes Evidence only. It is not Product authority, a second roadmap or permission to reopen B12.

## 2. Preserved Git identity

```text
repository   developmentconexus-ops/MetalDocs
source       claude/repo-context-technical-design-69t84i
evidence ref evidence/t11-b12-p8-r4-locks-20260826
exact commit c9a4414bc71a9ad5f347e2127c6b9e5905ff8b37
```

The remote ref resolves to the exact commit above. It MUST NOT move while current T11/P11/P13/P14 proof depends on B12.

## 3. Canonical B12 Evidence

| Evidence | Path on exact Evidence commit | Git blob |
|---|---|---|
| P8 R4 functional LOCK artifact (single self-contained HTML) | `docs/work/current/t11-b12-document-governance-p8.html` | `80276b92ff0c8a747c0c9ef964ce62fe3d556486` |
| P6/P7 planning + walkthrough/finding/LOCK record | `docs/work/current/t11-b12-document-governance-planning.md` | `109643e0fd90cdcf3ec04b328a06947562ab1a5b` |
| P9 Screen Contract | `docs/work/current/t11-b12-screen-contract.md` | `21bbbc81cee33f292dc003fbd2e1f5934851b7e7` |
| P10 pattern consolidation | `docs/work/current/t11-b12-pattern-consolidation.md` | `cedba46eeffd470c9de52326d637db365f88f08c` |

The Evidence commit history also preserves superseded R1–R3 candidates; the R4 identities above are canonical.

## 4. Protected B12 meaning

Durable Product/architecture authority remains in current Product, architecture and bounded decision owners (`template-configuration-read.md` for the ratified op43 read precision). The LOCK protects the frontend structure proved by the exact P8/P9/P10 Evidence, including:

```text
stable /admin/document-governance route — third Admin Center section
Tipos de documento / Modelos local lenses
type detail with three separate If-Match write domains (base op36/37, governance+representation op38/39,
  eligible templates op40/41) — a stale 412 in one section never contaminates its siblings
"Publicação oficial" presented as its own section while remaining one op39 write truth
sequential route editor: ordered steps, named_user|group selector, optional due_in_days,
  keyboard reorder (never drag-only)
op35 idempotent create with duplicate-code 409 and same-key ambiguous replay
numbering preview op42 with reservation:false honesty and area_id validation
Modelos lens over refined op43: server-side q + eligible-type/role/effective filters before pagination,
  filter change = new first-page identity, empty page = ordinary result
template detail dialog carrying only the journeys-§23 projection, with boundary terminators
  "Abrir documento → B03" and "Histórico → B07" (revisions/approval live on the Document's own
  surfaces under the user's read permissions)
403 denial panel / 404 non-disclosure / visible server-page traversal throughout
```

Operator dispositions preserved by this Evidence: B12-F1 REJECTED (try-and-fail on code/scope immutability), B12-F2 RATIFIED (op43 read precision), inline revision data in the admin lens declined (no §23 reopen), "general template" remains a derived fact.
