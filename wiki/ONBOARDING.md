# Onboarding — Day 1

> **Last verified:** 2026-08-14
> **For:** Any engineer or agent new to MetalDocs.
> **Current program:** Cohesive Platform Redesign — design-only.

## 0. Read these first

1. [`AGENTS.md`](../AGENTS.md)
2. [`wiki/standards/root-cause-global-maximum-method.md`](standards/root-cause-global-maximum-method.md)
3. **[`wiki/architecture/cohesive-platform-redesign.md`](architecture/cohesive-platform-redesign.md)**
4. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
5. [`wiki/references/current-agent-handoff.md`](references/current-agent-handoff.md)

If your task touches product/domain architecture, **stop there and follow the active redesign**. No product implementation is authorized yet.

## 1. What MetalDocs is becoming

MetalDocs is a governed operational-information platform. The active target is converging on:

```text
Organization + Authorization
           ↓
Controlled Information
Document + DocumentRevision
           ↓
Approval V1
           ↓
Domain Governance + Release
           ↓
Effective Revision
```

Supporting audit, rendering/renditions, periodic review, distribution/read-ack, notifications, search, tokens and security consume those canonical truths.

Do not learn the product as `templates → controlleddocuments → documents → old approval route`. That is current/history evidence, not the target domain model.

## 2. Current implementation vs target

The repository still runs its current modular-monolith code and current Postgres schema. Use those to answer **what happens today**.

Core current module pages are explicitly marked `LEGACY`/`CURRENT-STATE` because their boundaries are being replaced:

- approval
- auth/IAM boundaries
- controlled-documents
- documents
- taxonomy
- templates
- jobs as a “module”

Read [`wiki/modules/index.md`](modules/index.md) for classification.

## 3. Get the current runtime running — only when needed

Canonical local startup:

```powershell
.\scripts\start-api.ps1
```

Use [`wiki/references/local-dev-startup.md`](references/local-dev-startup.md) as runtime startup truth. The current architecture reset normally needs repository/source analysis, not a running product stack, unless validating a specific current-state premise.

Frontend current runtime:

```powershell
cd frontend/apps/web
corepack pnpm install
corepack pnpm dev
```

## 4. Task-oriented reading

- **Whole-product/domain redesign:** active redesign stack above.
- **Current HTTP/OpenAPI evidence:** [`architecture/backend-api-structure.md`](architecture/backend-api-structure.md), [`architecture/api-contract.md`](architecture/api-contract.md), [`architecture/api-design-system.md`](architecture/api-design-system.md).
- **Current DB/schema evidence:** [`database/index.md`](database/index.md) + live schema/migrations.
- **Current frontend evidence:** [`architecture/frontend-structure.md`](architecture/frontend-structure.md).
- **Current module evidence:** [`modules/index.md`](modules/index.md); respect LEGACY markers.
- **QA evidence:** [`quality/qa-operating-system.md`](quality/qa-operating-system.md).
- **ADR evidence:** [`decisions/index.md`](decisions/index.md), which now classifies retained vs historical/reopened target semantics.

## 5. Design workflow while implementation gate is closed

1. State the product/domain question.
2. Inspect current code only as evidence.
3. Compare mature products/standards/libraries when useful.
4. Apply Root-Cause / Global-Maximum + YAGNI.
5. Present alternatives and the smallest correct target.
6. Record operator-approved decisions in the active ledger.
7. Continue to the next dependency.
8. **Do not implement product code.**

## 6. Historical docs

The old `docs/superpowers` roadmaps/milestones/plans/reports/specs/analyses were intentionally removed from the live tree on 2026-08-14. Git history is the archive.

Do not restore them or infer a current task from them. If a historical artifact is needed to answer a specific evidence question, inspect that historical commit and return to the active redesign authority.

## 7. Exact current next step

Continue the active design with:

```text
DocumentType
+ DocumentFamily
+ GovernanceClass
+ TemplateDesignation/default policy
```

Then proceed to the complete Document/Revision lifecycle only after that section is approved.
