# Quality

> **Last verified:** 2026-05-27
> **Scope:** Canonical home for QA operating-system material, scenario-proof governance, and release-quality standards.

## Canonical pages

- [qa-operating-system.md](qa-operating-system.md) - review, QA, root-cause remediation, evidence, and hard-stop operating model
- [release-readiness.md](release-readiness.md) - canonical merge/release Go/No-Go gate
- [deep-qa/index.md](deep-qa/index.md) - canonical deep-QA execution artifacts for documents + approval

## Reusable QA checklists

- [screen-qa-checklist.md](screen-qa-checklist.md) - default QA pass for user-facing screen and interaction work
- [backend-api-qa-checklist.md](backend-api-qa-checklist.md) - default QA pass for backend/API and contract-backed route work
- [workflow-async-qa-checklist.md](workflow-async-qa-checklist.md) - required proof split for worker-owned, delayed, and multi-stage workflows
- [release-closeout-checklist.md](release-closeout-checklist.md) - final close-out checklist before merge/release declaration

## Current source inputs and compatibility areas

- [deep-qa/index.md](deep-qa/index.md) - current canonical deep-QA execution artifact set
- [../references/documents-approval-deep-qa/README.md](../references/documents-approval-deep-qa/README.md) - compatibility breadcrumb for older deep-QA links
- [../references/ai-operating-system.md](../references/ai-operating-system.md) - path-stable compatibility bridge still referenced by repo instructions; this folder remains canonical for QA close-out behavior
- [../../docs/superpowers/specs/2026-05-20-documents-approval-product-plus-qa-system-design.md](../../docs/superpowers/specs/2026-05-20-documents-approval-product-plus-qa-system-design.md) - design input, not canonical runtime truth
- [../../docs/runbooks/release-readiness.md](../../docs/runbooks/release-readiness.md) - staging source now superseded by [release-readiness.md](release-readiness.md)

## Placement rule

- Durable QA operating knowledge belongs under `wiki/quality/`.
- Temporary QA design and exploratory material stays in `docs/` until promoted.
- Existing QA artifacts under `wiki/references/` may remain as compatibility breadcrumbs after bounded promotion into `wiki/quality/`.
