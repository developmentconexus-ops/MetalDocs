# Phase 0 Audit — novo-template-escopo

> **Confirmed by user:** 2026-05-09
> **Status:** GATE PASSED

---

## Keep / Cut / Defer table

| Element | Maps to | Decision | Reason |
|---|---|---|---|
| Rail (56px nav) | AppShell — always present | Keep | Standard shell, not in page scope |
| Toolbar breadcrumb "Templates / Novo" | AppToolbar breadcrumb via route handle | Keep | Real nav context |
| Page kicker "Templates / Novo" | `WizardShell` kicker prop | Keep | Real |
| Page h1 "Novo template reutilizável" | `WizardShell` title prop | Keep | Real |
| Page description + `{{campo}}` mono | `WizardShell` description prop | Keep | Real UX copy |
| 5-step Stepper | `Stepper` component | Keep | Real wizard nav |
| Card "Etapa 1 de 5" kicker | Step counter in step component | Keep | Real |
| "Escopo do template" h2 | Step card heading | Keep | Real |
| Card intro description | Static copy | Keep | Real |
| Tab "Para um perfil" | `scope='profile'` state via `TabBar` | Keep | Primary flow, API exists |
| Tab "A partir de um documento" | — | **Cut** | Wrong concept. Replaced by "Escopo de aplicação" field in Step 2 (Identidade). Template scope is a filter hint set at creation, not a document-clone flow. |
| Profile cards grid 2×3 | `SelectableCard` + `DocumentProfile[]` from taxonomy API | Keep | API exists |
| Profile code badge (mono "POP") | `CodeChip` or mono span | Keep | Real |
| Profile name + description in card | `DocumentProfile.name` + `.description` | Keep | Real |
| Profile family label | `DocumentProfile.familyCode` | Keep | Real |
| "X templates publicados" count per card | No summary endpoint | **Defer** | No API. Show `—`. Same TODO pattern as doc wizard `StepProfile`. |
| "em breve" badge on CHK card | `disabled aria-disabled title="Em breve"` | Keep | CHK not yet implemented |
| Document search + list (Tab B content) | — | **Cut** | Tab B cut entirely |
| Info banner "O conteúdo será clonado…" | — | **Cut** | Tab B cut entirely |
| Footer "Cancelar" (step 1, no back) | `WizardFooter` `showBack=false` | Keep | Real |
| Footer step label caption | `WizardFooter` `stepLabel` prop | Keep | Real |
| Footer "Avançar →" | `WizardFooter` `primaryDisabled` when no selection | Keep | Real |

---

## Confirmed cuts

1. **"A partir de um documento" tab** — cut entirely. Wrong UX model (misconception in design). Replaced by `Escopo de aplicação` field in Step 2.
2. **Tab B content** (document search, document list, clone info banner) — cut, tied to cut tab.

## Confirmed defers

1. **Template count per profile card** — show `—` with TODO comment block referencing `wiki/backlog/novo-template-wizard.md`.

---

## Product decision recorded (backlog)

**Template scoping model (confirmed by user 2026-05-09):**

Templates have an optional `Escopo de aplicação` set at creation time (Step 2):
- `Todos os [perfil]` — default, shows for all docs of that profile
- `[perfil] — Subtipo` — shows only for docs matching profile+subtype (deferred until Taxonomy has subtypes)
- `Genérico` — shows for any document type

Scope is a **filter hint** (soft), not a hard binding. Actual template-to-document link happens when user picks a template during document creation (doc wizard Step 3). Template remains reusable across multiple documents.

See `wiki/backlog/novo-template-wizard.md` for full backlog item.
