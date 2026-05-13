# Tech Debt Register - frontend-primitives

> Companion to `wiki/modules/frontend-primitives.md`. Debt only; no fix prescriptions.

**Last verified:** 2026-05-13

## Items

### T-001 · Module page scope is narrower than actual `components/ui` surface
- **Severity:** major
- **Surface:** `frontend/apps/web/src/components/ui/index.ts`
- **Observation:** doc focuses on `SelectableCard` and `useRovingRadioGroup` while package exports many additional primitives.
- **Evidence:** index export surface and component list.
- **Linked backlog row:** `R-001`
- **Linked ADR:** missing-ADR

### T-002 · TabBar still uses bespoke roving logic instead of shared hook
- **Severity:** minor
- **Surface:** `frontend/apps/web/src/components/ui/TabBar.tsx`
- **Observation:** keyboard navigation logic is duplicated instead of unified on shared hook.
- **Evidence:** hook consumers and TabBar local behavior.
- **Linked backlog row:** `R-002`
- **Linked ADR:** missing-ADR

### T-003 · Primitive governance boundaries are convention-only
- **Severity:** minor
- **Surface:** `wiki/architecture/frontend-structure.md`
- **Observation:** "domain-agnostic only" rule exists in docs but no enforcement rule/check.
- **Evidence:** no lint or policy gate tied to primitive boundaries.
- **Linked backlog row:** `R-003`
- **Linked ADR:** missing-ADR

## Coverage stats

- Public symbols undocumented: n/a (not fully audited)
- Operations missing C4 placement: n/a (frontend component module)
- Cross-deps missing in section map: n/a (partial module doc)
- State transitions missing: n/a
- Decisions without ADR link: 3
