---
name: runtime-contract-prereq
description: Use when startup, runtime wiring, migrations, route truth, OpenAPI truth, generated code truth, or frontend-wrapper truth may have drifted after a refactor or rename and feature work must stop until the system is trustworthy again.
---

# Runtime + Contract Prerequisite Audit

Use this before screen or feature work when local runtime truth is unreliable.

## Goal

Restore trust in:
- startup truth
- auth/session truth
- target route truth
- module contract truth
- workflow truth after the incident

## Audit procedure

Check in this order and keep evidence for the first failing boundary:
1. Startup truth: verify the supported startup path, current runtime wiring, and whether the running process/binary matches current source.
2. Auth/session truth: verify login, session creation, and the first authenticated call needed by the task.
3. Target route truth: verify the exact route, method, prefix, and reachable handler for the task boundary.
4. Contract truth: compare runtime route registration / owning handler files, the OpenAPI path, generated backend code if applicable, generated frontend API types, and feature API wrapper/client behavior.
5. Workflow gap truth: note whether stale scripts, missing checks, or outdated skill/wiki guidance helped the drift survive.

## Required outputs

- issue classification using one of:
  - runtime prerequisite
  - shared contract prerequisite
  - workflow/tooling gap
  - module-local implementation only if truly bounded
  - screen-local implementation only if truly bounded
- route/runtime/spec/generated/frontend wrapper comparison
- prerequisite repair boundary
- exit-gate verification
- workflow gap updates required before resuming feature work

## Stop rules

Do not continue into feature work while the prerequisite boundary is still failing.
Do not classify a shared runtime or contract mismatch as local unless the break is truly bounded to the current module or screen.

## Exit behavior

- write root cause
- bound repair scope
- rerun the failed checkpoint
- do not resume feature work until the repaired boundary passes
