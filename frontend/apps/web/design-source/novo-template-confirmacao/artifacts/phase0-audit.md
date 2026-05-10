# Phase 0 — Audit · novo-template-confirmacao

**Date:** 2026-05-09
**User confirmed:** backend wiring mocked, visual only

## Keep / Cut / Defer table

| Element | Maps to | Decision | Reason |
|---|---|---|---|
| Kicker "Etapa 5 de 5" | static | Keep | |
| H2 "Revise e confirme a criação" | static | Keep | |
| Intro caption | static | Keep | |
| Paper-mock thumbnail (120×152) | decorative DOM | Keep | visual only, no data dependency |
| Code chip "TPL-XXX-NNN" | mocked next-code pattern | Keep (mock) | next-code-preview backlog; reuses step 2 pattern |
| StatusPill "draft" | StatusPill primitive | Keep | always draft at creation |
| "v1.0" pill | static | Keep | always v1.0 |
| Template name | `state.name` | Keep (real) | |
| Metadata row "Perfil" | `state.profileCode` + profile name | Keep (real) | from selectedProfile prop |
| Metadata row "Família" | `selectedProfile.family` | Keep (real) | |
| Metadata row "Origem" | `state.startingPoint` + selectedDocxName | Keep (real) | |
| Metadata row "Auto-fill" | placeholder count | **Cut** | no extraction endpoint |
| Metadata row "Permissões" | derived from permissionsMode/selections | Keep (computed) | |
| Metadata row "Autor" | `user.displayName` from auth store | Keep (real) | |
| "Ao confirmar" ol (4 items) | static + adapted | Keep | item 4 generic (no QUA-COORD hardcode) |
| Checkbox "código definitivo" | `useState(false)` | Keep | gates submit |
| CTA "Criar e abrir editor →" | mocked → navigate('/templates-v2') | Keep (mock) | `confirmacao-backend-submit` backlog |
