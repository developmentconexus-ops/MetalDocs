---
name: harness-hub
description: Coordinate MetalDocs work. During the active Cohesive Platform Redesign this skill is DESIGN-ONLY: it may parallelize research/census/review, but MUST NOT revive the deleted docs/superpowers roadmap or dispatch product implementation.
---

# Harness Hub — Cohesive Redesign Override

## Hard stop

Before any hub/dispatch action read:

1. `AGENTS.md`
2. `wiki/standards/root-cause-global-maximum-method.md`
3. `wiki/architecture/cohesive-platform-redesign.md`
4. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
5. `wiki/references/current-agent-handoff.md`

The previous `docs/superpowers/HARNESS.md`, `ROADMAP.md`, milestones, plans, reports and specs were intentionally removed from the live tree on 2026-08-14. **Do not recover them from Git history to rebuild a queue.**

## Allowed while redesign implementation gate is closed

The hub may coordinate independent **read-only** work such as:

- repository/current-runtime census;
- impact mapping;
- comparison of mature products/standards/libraries;
- adversarial review of proposed domain/architecture decisions;
- documentation consistency checks.

All conclusions return to the **single active redesign ledger**. Do not create parallel design authorities.

## Forbidden while gate is closed

Do not dispatch:

- product code changes;
- DB migrations/schema work;
- OpenAPI changes;
- frontend implementation;
- old A8/Approval/ControlledDocuments/Templates milestone continuation;
- compatibility patches intended only to preserve concepts the target may delete.

## Execution restart trigger

Implementation orchestration becomes legal only after the redesign ledger records that:

1. whole-product domain design is closed;
2. final Organization/AuthZ/Approval/Controlled Information/supporting-concern semantics are closed;
3. build-vs-buy and technical boundaries/data model/contracts are closed;
4. migration/deletion map is closed;
5. final ADR/spec material is promoted to `wiki/`;
6. adversarial review passes;
7. operator approves the integrated design;
8. an implementation plan and new execution queue are authored from that accepted target.

At that point create a **new** queue. Never revive the pre-reset queue by inertia.
