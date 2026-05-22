---
extends:
  - ~/.claude/rules/ecc/golang/coding-style.md
  - ~/.claude/rules/ecc/golang/security.md
  - ~/.claude/rules/ecc/golang/testing.md
  - ~/.claude/rules/ecc/golang/patterns.md
evidence:
  - wiki/reviews/2026-05-21-go-backend-review.md
  - wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md
  - wiki/reviews/2026-05-21-go-backend-review/cmd-metaldocs-api.md
enforced_by:
  - .golangci.yml
  - .github/workflows/golangci-lint.yml
---
# MetalDocs Go Quality Bar

> **Last verified:** 2026-05-22
> **Scope:** hand-written Go under `apps/api/` and `internal/`
> **Out of scope:** generated Go, frontend TypeScript, database migration policy

This bar turns the Critical and High findings from the first Go backend reviews into a reusable standard. Every new Critical or High review finding must cite one of these anchors.

| Section | Bar Doc | Failure Mode Prevented | Finding ID | Commit SHA | Lint Rule | Extends Rule |
|---|---|---|---|---|---|---|
| Typed boundaries | [typed-boundaries.md](typed-boundaries.md#the-rule) | raw strings crossing auth, tenant, role, and problem-code boundaries | C5, H9, H11, M3, M6, L8 | d2242313, e1daeeb3 | exhaustive, revive | coding-style.md |
| Errors and logging | [errors-and-logging.md](errors-and-logging.md#error-wrapping-rule) | swallowed errors, lost context, double logging | C3, H1, H4, L2, L6 | 12cae0f9 | errcheck, errorlint, nilerr | coding-style.md |
| Security boundaries | [security-boundaries.md](security-boundaries.md#fail-closed-authn-useridfromcontext) | fail-open authn, spoofed proxy headers, malformed error envelopes | C1, C2, C5, H2, H4, H7 | def24e4a, 2f8f6dcc, d2242313, 73a769aa | gosec, contextcheck | security.md |
| Idempotency and concurrency | [idempotency-and-concurrency.md](idempotency-and-concurrency.md#two-phase-write-pattern) | duplicate side effects, replay races, wedged in-flight rows | C3, C4, H1, H11, M10 | 12cae0f9, 07312d58 | manual-review | patterns.md |
| Persistence | [persistence.md](persistence.md#parameterized-queries-only) | SQL injection, leaked rows, ignored row errors | H11, M9, M10, M11 | 12cae0f9 | sqlclosecheck, rowserrcheck, bodyclose | patterns.md |
| HTTP handlers | [http-handlers.md](http-handlers.md#handler-anatomy) | public route drift, missing validation, wrong middleware order | Module #1 C1-C4, H1, H4, H7 | 6eb31ec7, 66fe1ee3 | contextcheck, gocyclo | patterns.md, security.md |
| Testing | [testing.md](testing.md#no-mock-db-rule) | mock/real DB divergence, order-dependent tests, missed races | process rule | manual-review | manual-review, go test -race | testing.md |
| Package layout | [package-layout.md](package-layout.md#import-direction-law) | import cycles, business logic in cmd, invalid zero-value config | Module #1 C2, C3, C4, H10 | 66fe1ee3 | depguard or manual-review | patterns.md |
| Refactor playbook | [refactor-playbook.md](refactor-playbook.md#overview) | inconsistent review/fix sequencing and missing evidence updates | Module #1, #2a | review tracker | manual-review | project process |

## How to use this bar

Agents reviewing Go backend code must cite `wiki/standards/golang/<doc>.md#<anchor>` in every Critical or High finding. If a real defect does not fit an existing anchor, fix the defect and update the closest bar doc in the same PR.
