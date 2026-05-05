# Screen Redesign — Implementation Tracker

**Branch:** `feature/screen-redesign`
**Spec:** `docs/superpowers/specs/2026-05-05-screen-redesign-design.md`
**Design source:** `frontend/apps/web/design-source/`
**Last updated:** 2026-05-05 · Foundation ✅

---

## Flow

Foundation → Login → Library → Editor → Wizard → Templates → Registry → Dashboard → Approval Inbox → Signoff Detail

Each block gets its own detailed plan in `docs/superpowers/plans/`. Mark complete here when merged into `feature/screen-redesign`.

---

## Status

| Block | Description | Plan | Status |
|---|---|---|---|
| **Foundation** | Tokens rename, fonts, Zustand cleanup, UI primitives, queryKeys, AppShell, Rail, AppToolbar, AppRoot, Router restructure | `wiki/implementation/plan-foundation.md` | ✅ Complete |
| **Login** | Full-page split layout, auth form, no Rail | `wiki/implementation/plan-login.md` | 🔲 Not started |
| **Library** | Dense document table, stat cards, filter tabs, collapsible activity sidebar, SectionPanel | — | ⏳ Waiting on Login |
| **Editor** | Slim doc bar, mini toolbar, paper canvas, metadata sidebar | — | ⏳ Waiting on Library |
| **Wizard** | 4-step stepper, profile/area/visibility/template pickers | — | ⏳ Waiting on Editor |
| **Templates** | 3-col card grid, mini doc preview | — | ⏳ Waiting on Wizard |
| **Registry** | Sequence counter grid, controlled documents table | — | ⏳ Waiting on Templates |
| **Dashboard** | Hero greeting, stat row, pending approvals, shortcuts | — | ⏳ Waiting on Registry |
| **Approval Inbox** | Filter strip, multi-select table, pagination | — | ⏳ Waiting on Dashboard |
| **Signoff Detail** | A4 diff view, approval flow panel, decision form | — | ⏳ Waiting on Approval Inbox |

---

## Legend

| Symbol | Meaning |
|---|---|
| 🔲 | Not started |
| 🔄 | In progress |
| ✅ | Complete — merged into branch |
| ⏳ | Blocked — waiting on previous block |

---

## How to Update

When starting a block: change status to `🔄 In progress`, add plan link if missing.
When block is complete and committed: change to `✅ Complete`, update next block from `⏳` to `🔲`.

---

## Key Files Reference

| File | Purpose |
|---|---|
| `src/styles/tokens.css` | Wine palette CSS vars (`--brand`, `--bg`, `--rail`, etc.) |
| `src/lib/queryKeys.ts` | Centralized TanStack Query key constants |
| `src/components/ui/Icon.tsx` | SVG icon component |
| `src/components/ui/Avatar.tsx` | Initials avatar |
| `src/components/ui/StatusPill.tsx` | Document status pill |
| `src/components/ui/CodeChip.tsx` | Code chip wrapper |
| `src/features/shell/pages/AppRoot.tsx` | Auth guard + bootstrap |
| `src/features/shell/components/AppShell.tsx` | Rail + Toolbar + Outlet layout |
| `src/features/shell/components/Rail.tsx` | 56px dark nav sidebar |
| `src/features/shell/components/AppToolbar.tsx` | 52px top bar |
| `src/app/AppRouter.tsx` | Router (public `/login` + protected routes) |
| `src/features/auth/pages/LoginPage.tsx` | Full-page login, no Rail |

---

## Design System Reference

**Palette:** Wine — `--brand: #6b1f2a`, `--rail: #2a1418`, `--bg: #f4eeee`, `--accent: #c8364a`

**Fonts:** Inter Tight (sans) + JetBrains Mono (mono)

**Key classes:** `.card`, `.btn`, `.btn-primary`, `.btn-ghost`, `.btn-sm`, `.pill`, `.pill-draft/review/approved/frozen/rejected/archived`, `.code-chip`, `.mono`, `.kicker`, `.avatar`, `.avatar-sm`, `.row`, `.divider`, `.spacer`

**Shell dimensions:** Rail 56px · Toolbar 52px · SectionPanel 224px (Library only) · Activity sidebar 320px (Library only)
