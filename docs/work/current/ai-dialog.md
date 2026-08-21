# Cross-Repository Engineering Alignment — MetalDocs ↔ Marketplace Central

> **TEMPORARY REVIEW EVIDENCE ONLY — NOT PRODUCT / ARCHITECTURE / ROADMAP AUTHORITY**
>
> Delete this file after convergence. Do not merge it into MetalDocs `main`.

## Exact review subjects

```text
METALDOCS
repository        developmentconexus-ops/MetalDocs
current main      82832cce62d11ea90575fb484b97e3c934c03e37
T1→T8-H           CLOSED / OPERATOR-RATIFIED / INTEGRATED
T9                NEXT / NOT STARTED — requires explicit operator authorization
T10→T12           NOT OPEN
implementation    BLOCKED

MARKETPLACE CENTRAL
repository        developmentconexus-ops/marketplace-central
candidate branch  stage/d6-frontend
candidate SHA     cb55238c1908b087989825ff4d2ad9ce6f08527b
candidate PR      #54
D6-B1             OPERATOR-RATIFIED
D6-B2             ACTIVE DECISION
D7–D9             BLOCKED
implementation    BLOCKED UNTIL D9
```

This review was deliberately rebased onto integrated MetalDocs `main` after PR #148 merged during setup. The stale cross-review PR #152 was closed unmerged.

## Operator goal

Align engineering decisions across the two products where the **protected property and failure class are genuinely the same**, while preserving justified differences. The operator explicitly does **not** require shared code or a shared platform today.

Use DevelopmentConexus Root Cause / Global Maximum / YAGNI / falsification discipline.

Required classifications:

```text
ALIGN
DIVERGE_JUSTIFIED
REOPEN_MARKETPLACE
REOPEN_METALDOCS
DEFER
STOP
```

## Interaction rule

1. Start fresh from MetalDocs authority:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ only the bounded task authority pack
```

2. Then inspect Marketplace current authority:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ docs/engineering/rebaseline/D6-FRONTEND.md + ARCHITECTURE.md
```

Switch to Marketplace D2 only for the concrete AuthN/identity question. Use the routed engineering-research guide only for concrete technology research.

3. Read the incoming Marketplace review at:

```text
repository: developmentconexus-ops/marketplace-central
branch:     review/d6b2-metaldocs-alignment
PR:         #57
file:       docs/work/current/ai-dialog.md
```

Do not assume its findings are correct. It is Evidence to attack.

4. Use current official docs/upstream repositories for technology claims that may have changed.

5. Write only under `## MetalDocs reciprocal response` in this file. Modify no other file.

## Review scope

Perform a reciprocal review of Marketplace Central and, where Marketplace evidence exposes a stronger solution, challenge already-ratified MetalDocs technology decisions too.

At minimum adjudicate:

- React SPA realization;
- TanStack Query server-state ownership;
- OpenAPI TypeScript generation;
- native `fetch` thin transport vs `openapi-fetch` / Orval / Hey API / other current alternatives;
- route tree and whether a router dependency is materially justified;
- feature/package topology: MetalDocs lens-first vs Marketplace's unapproved owner-first hypothesis vs bounded hybrid;
- form and UI component dependencies;
- external OIDC boundary;
- Keycloak as first concrete provider while architecture remains provider-neutral;
- browser ApplicationSession / cookie / CSRF model;
- Go modular-monolith topology and package-direction law;
- Go `net/http` vs framework;
- generated Go OpenAPI boundary;
- PostgreSQL + pgx / SQL generation / migration tooling;
- tenant isolation and whether Marketplace justifiably needs a stronger mechanism than MetalDocs;
- River or another durable-job mechanism;
- OpenTelemetry / OTLP / slog;
- configuration, test-environment and security tooling;
- same-origin SPA/API profile;
- dependency-version alignment policy;
- mechanical frontend/backend dependency boundaries;
- deliberate absences: BFF, SSR, Redis, microfrontends, realtime, generic event bus/workflow, ORM.

## Required adversarial questions

1. Is `openapi-typescript + thin native fetch + TanStack Query` still the Global Maximum shared frontend baseline, or does Marketplace expose a real gap?
2. Does Marketplace need a router library where MetalDocs did not? If yes, is that a justified divergence or should MetalDocs reopen too?
3. Is MetalDocs lens-first feature topology stronger than Marketplace owner-first folders, or does Marketplace expose a better hybrid?
4. Does Marketplace's owner/composition model expose any material weakness in MetalDocs T8-F?
5. Is Keycloak the best first concrete IdP for both products while keeping OIDC/provider-neutral architecture?
6. Which MetalDocs Go/runtime decisions should Marketplace D7 begin from rather than rediscovering from zero?
7. Which decisions must *not* be copied because Marketplace's tenant/provider workload differs?
8. Is River a real shared candidate or only familiarity bias?
9. Is MetalDocs' no-RLS Launch posture correctly different from Marketplace tenant-ready isolation?
10. Are any MetalDocs technology choices now materially inferior under current 2026 upstream evidence?
11. Are any Marketplace accepted decisions materially weaker than MetalDocs and worth reopening?
12. What is the smallest shared technology profile that reduces engineering/LLM cognitive drift without creating coupled releases or a generic internal platform?

## Required response shape

For each material item:

```text
ID
PROPERTY / FAILURE CLASS
METALDOCS CURRENT DECISION
MARKETPLACE CURRENT DECISION
CURRENT PRIMARY EVIDENCE
CLASSIFICATION = ALIGN | DIVERGE_JUSTIFIED | REOPEN_MARKETPLACE | REOPEN_METALDOCS | DEFER | STOP
RATIONALE
SMALLEST NEXT ACTION
REOPEN TRIGGER
```

End with exactly these sections:

```text
SHARED PROFILE — safe to align now
REPO-SPECIFIC DIFFERENCES — must remain different
DEFERRED PROFILE — align later only after owning stage proves need
METALDOCS REOPEN DECISION
MARKETPLACE REOPEN DECISION
CONTINUATION RECOMMENDATION
```

Do not open T9. Do not begin Marketplace D7. Do not implement code. Reviewer output is Evidence, not authority.

---

## MetalDocs reciprocal response

<!-- MetalDocs reviewer writes here only. -->
