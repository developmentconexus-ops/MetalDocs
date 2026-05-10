# Phase 0 — Audit · novo-template-permissoes

**Date:** 2026-05-09
**User confirmed:** inline (no cuts; all elements map to mocked or real data)

## Keep / Cut / Defer table

| Element | Maps to | Decision | Reason |
|---|---|---|---|
| Kicker "Etapa 4 de 5" | static label | Keep | |
| H2 "Quem pode usar este template?" | static heading | Keep | |
| Intro caption | static text | Keep | |
| Segmented control (3 modes) | `permissionsMode` state | Keep | pure UI |
| Mode "roles" — 2-col role cards | mocked role data | Keep (mock) | no personnel-roles API |
| Mode "areas" — 3-col area cards | mocked area data | Keep (mock) | area names from design, counts mocked |
| Mode "all" — company banner + ~340 users | mocked count | Keep (mock) | no user-count endpoint |
| Coverage summary block | derived from selection + mocked counts | Keep (mock) | computed locally |
| Icon home / taxonomy / users / check | `Icon.tsx` | Keep | all supported in IconName union |
| Advance gate: roles mode + 0 selected | pure UI gate | Keep | |
| User counts per role (14, 8, 2, etc.) | mock constant | Keep (mock + TODO) | permissions-api backlog |
| Area user counts (28, 89, 34, etc.) | mock constant | Keep (mock + TODO) | permissions-area-counts backlog |
| ~340 company users | mock constant | Keep (mock + TODO) | permissions-user-count backlog |

## No cuts

All design elements are Keep or Keep (mock). No element implies unsupported server behavior beyond the mocked counts.
