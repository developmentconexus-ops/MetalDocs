# Arc42 Cheatsheet

Condensed reference for module-doc composers. Full template upstream: https://arc42.org/overview (v8.2).

## The 12 sections, plain English

| # | Section | One-line purpose | Skip if… |
|---|---|---|---|
| 1 | Introduction & Goals | what this is, who uses it, top quality goals | never skip |
| 2 | Architecture Constraints | bind decisions: language, persistence, contracts, regs | never skip |
| 3 | System Scope & Context | the C4 Context view + business + technical | never skip |
| 4 | Solution Strategy | 3–5 load-bearing choices, each linked to an ADR or constraint | never skip |
| 5 | Building Block View | C4 Container + public surface table | never skip |
| 6 | Runtime View | sequence per selected scenario; state machines | skip only if module is pure-data |
| 7 | Deployment View | how it ships (binary, container, migrations, env) | rarely skip |
| 8 | Cross-cutting Concepts | authz, errors, idempotency, logging, tx — pointers to canonical docs | never skip |
| 9 | Architecture Decisions | table of decisions ↔ ADRs (or missing-ADR flag) | never skip |
| 10 | Quality Requirements | concrete scenarios that prove §1 goals | never skip |
| 11 | Risks & Technical Debt | pointer to `<m>-tech-debt.md` + top 3 | never skip |
| 12 | Glossary | module-specific terms only (not project-wide) | skip if all terms are project-wide |

## Section-by-section composition tips

### §1 Quality Goals — top 3, RANKED, testable
Goals must be measurable. "Reliable" is not a goal; "P99 latency < 200 ms for `listX`" is.

### §3 Context diagram — only direct neighbors
If a system is two hops away, it is NOT in the Context view. Put it in §5 Container or §8.

### §4 Solution Strategy — every bullet links somewhere
Pattern: `<choice> — driver: <ADR link OR constraint name>`. No bare assertions.

### §5 Building Block View — Container + public surface
- Mermaid `C4Container` (NOT `C4Component` — that is level 3, usually overkill)
- Public surface table is the canonical list of exports; coverage gate (a) checks it

### §6 Runtime View — one scenario per major capability
- Reads, writes, state-transitions: one of each, minimum
- Sequence diagram + state table for state machines

### §8 Cross-cutting — link, don't re-explain
- Authz lives in `wiki/architecture/authz.md` — link, do not paraphrase
- Errors live in `wiki/concepts/error-ux.md` — link
- Idempotency lives in `wiki/architecture/api-design-system.md` — link
- If the module deviates from the canonical pattern, THAT is the §8 content for that subsection

### §9 ADR table — every load-bearing decision
Either link an ADR or mark `missing-ADR` and add a TD row. No invisible decisions.

### §11 Risks & Tech Debt — pointer-only
Body lives in `<m>-tech-debt.md`. Here: severity counts + top 3 one-liners.

## Common Arc42 anti-patterns

| Anti-pattern | Why it fails |
|---|---|
| §1 with vague goals ("be robust") | Cannot verify in §10 |
| §3 with 10 neighbors | Not Context — that's Container |
| §4 with "best practice" bullets | Not a decision; not a strategy |
| §5 with no public surface table | Coverage gate (a) cannot run |
| §6 with state machine implied but no table | Coverage gate (d) fails |
| §8 paraphrasing instead of linking | Drift surface ×N |
| §9 missing ADRs claimed as "obvious" | Future readers cannot reconstruct intent |
| §11 prescribing fixes | Belongs in `backlog/<m>-refactor.md`, not here |
