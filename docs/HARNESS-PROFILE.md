# Harness Profile — MetalDocs

> **Status:** DESIGN-ONLY OVERRIDE — 2026-08-14 Cohesive Platform Redesign

The previous mission/unit execution profile is preserved in Git history but is **not the active queue or execution authority**.

MetalDocs is currently designing the whole platform before the next product implementation wave. Do not launch implementation chips/lanes from old milestone/ROADMAP context.

## Active design authority

Read in order:

1. `AGENTS.md`
2. `wiki/standards/root-cause-global-maximum-method.md`
3. `wiki/architecture/cohesive-platform-redesign.md`
4. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
5. `wiki/references/current-agent-handoff.md`

`docs/superpowers/ROADMAP.md` and the former milestones/plans/specs/reports were removed from the live tree on 2026-08-14. Git history is their archive.

## Current harness rule

While the redesign implementation gate is closed:

- harness coordination MAY be used to parallelize **read-only research, repository census or adversarial design review**;
- harness coordination MUST NOT dispatch product implementation, migrations, API changes or frontend changes;
- no old unit/milestone status may be used as a continuation trigger;
- all approved decisions return to the single active redesign ledger.

## Stable repo/runtime bindings

These remain useful when current-state verification is explicitly needed:

- default branch: `main`;
- operator environment: Windows + PowerShell startup scripts;
- contract-first OpenAPI/oapi-codegen;
- `go build ./...`, `go test ./...` and `go run ./tools/verify ...` remain the verification mechanisms after implementation resumes;
- local stack/startup truth lives in `wiki/references/local-dev-startup.md`;
- QA truth lives in `wiki/quality/qa-operating-system.md` and the owning checklist;
- secrets are never read/printed/committed.

## Re-enable execution

A new execution profile/mission queue is created **only after**:

1. the whole-product redesign checklist closes;
2. final ADR/spec material is promoted to `wiki/`;
3. adversarial review finds no material ambiguity;
4. the operator approves the integrated design;
5. an implementation plan is authored from that accepted target.

At that point, build a new collision/dependency-aware execution queue from the new architecture rather than reviving the pre-reset queue.
