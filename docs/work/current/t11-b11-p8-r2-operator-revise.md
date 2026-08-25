# T11 — B11 P8 R2 — Operator REVISE

> **Disposition:** REVISE / VISUAL TOPOLOGY REGRESSION.
> **Trigger:** operator walkthrough after P8 R2 publication.
> **Scope:** frontend Evidence only; no B11-F1/Product/Authorization authority reopen.

## Operator finding

The P8 R2 content direction (`Por Área / Grupos / Funções`) was not rejected. The defect was continuity of the wireframe itself: R2 replaced the previously established MetalDocs low-fidelity shell/layout/topology with a new visual composition.

The operator expected continuity with the earlier B10/B11 wireframes, including the established family of:

```text
Evidence strip / review controls
MetalDocs header
light global navigation/sidebar
route + page heading + boundary
underlined local tabs
list/detail panel-grid cards
dialog interaction family
responsive/mobile behavior
```

R2 changed those elements without Evidence or authorization. That is accidental visual/topological redesign, not a Product improvement.

## Root cause

R2 was authored from scratch rather than inheriting the existing frontend Evidence topology. The new access IA was therefore coupled to an unnecessary shell redesign.

## Protected correction

P8 R3 must:

```text
PRESERVE B11-F1 and P7 R2 semantics
PRESERVE Por Área / Grupos / Funções
PRESERVE multi-Area Group footprint
PRESERVE Area-vs-Company grant separation
PRESERVE fixed read-only Roles
RESTORE the established MetalDocs wireframe topology/layout family
```

No Product/API/Authorization change is authorized by this finding.

## Proof

Regression verifier against R2:

```text
continuity checks: 4 / 13 PASS
```

Corrected R3:

```text
continuity checks: 13 / 13 PASS
functional browser checks: 42 / 42 PASS
```

Only the operator may LOCK R3 after walkthrough.
