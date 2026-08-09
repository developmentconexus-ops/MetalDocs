# Modules

> **Last verified:** 2026-08-09 (Phase G governance reconciliation — added `distribution.md` stub and linked the previously-unlinked `security` docs; index now covers all 15 code modules, PASS 14 finding D5)
> **Scope:** Durable per-module knowledge, tech-debt registers, and maturity state.

## Core product modules

- [approval.md](approval.md), [approval-tech-debt.md](approval-tech-debt.md)
- [audit.md](audit.md), [audit-tech-debt.md](audit-tech-debt.md)
- [auth.md](auth.md), [auth-tech-debt.md](auth-tech-debt.md)
- [controlled-documents.md](controlled-documents.md), [controlled-documents-tech-debt.md](controlled-documents-tech-debt.md)
- [documents.md](documents.md), [documents-tech-debt.md](documents-tech-debt.md)
- [iam.md](iam.md), [iam-tech-debt.md](iam-tech-debt.md)
- [taxonomy.md](taxonomy.md), [taxonomy-tech-debt.md](taxonomy-tech-debt.md)
- [templates.md](templates.md), [templates-tech-debt.md](templates-tech-debt.md)
- [tokens.md](tokens.md), [tokens-tech-debt.md](tokens-tech-debt.md) — SP-1 per-tenant author-defined `name → value` dictionary; capabilities `token.view` + `token_dictionary.manage`; published `DictionaryReader` port for SP-2 render substitution

## Frontend-focused modules

- [editor-chrome.md](editor-chrome.md), [editor-chrome-tech-debt.md](editor-chrome-tech-debt.md)
- [editor-ui-eigenpal.md](editor-ui-eigenpal.md), [editor-ui-eigenpal-tech-debt.md](editor-ui-eigenpal-tech-debt.md)
- [frontend-primitives.md](frontend-primitives.md), [frontend-primitives-tech-debt.md](frontend-primitives-tech-debt.md)
- [novo-documento-wizard.md](novo-documento-wizard.md), [novo-documento-wizard-tech-debt.md](novo-documento-wizard-tech-debt.md)
- [frontend/index.md](frontend/index.md) — per-feature frontend module pages (approval, auth, controlled-documents, documents, iam, templates)

## Supporting modules

- [distribution.md](distribution.md) — distribution/read-acknowledgement surface (Stage-1 stub, created 2026-08-09)
- [jobs.md](jobs.md) — background-job / async worker module (Stage-1 draft)
- [notifications.md](notifications.md) — per-recipient notification inbox; two delivery-only River workers
- [render-fanout.md](render-fanout.md), [render-fanout-tech-debt.md](render-fanout-tech-debt.md)
- [search.md](search.md), [search-tech-debt.md](search-tech-debt.md)
- [security.md](security.md), [security-tech-debt.md](security-tech-debt.md), [security-signals.md](security-signals.md) — MFA coverage, security settings (docs existed but were unlinked until 2026-08-09)

## Governance

- [maturity-audit-2026-05-13.md](maturity-audit-2026-05-13.md) - module maturity baseline

Module research artifacts remain beside the owning module and are supporting evidence, not first-stop canonical summaries.
