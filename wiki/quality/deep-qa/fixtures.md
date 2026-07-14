# Documents + Approval Fixture Registry

Date: 2026-05-20
Status: active
Canonical home: `wiki/quality/deep-qa/fixtures.md`
Compatibility path: `wiki/references/documents-approval-deep-qa/fixtures.md`
Scope: reusable and consumable fixtures for modern `documents + approval` QA

## 1. Rules

- A fixture must be explicitly marked reusable or consumable
- A fixture must name the represented state, not just the ID
- A fixture that is advanced during QA must be updated in this registry before session close

## 2. Fixture Entry Format

- fixture name
- controlled document ID
- document ID
- approval instance ID when relevant
- represented state
- creation path
- safe reuse
- advancement path
- discard rule
- caveats

## 3. Canonical Fixtures

### Fixture: controlled document mainline

- controlled document ID: `750afeba-6e35-4dd4-8a74-2b51b9f9090c`
- document ID: `1588c6ff-179d-4be5-a8f5-b9bad6f1727c`
- represented state: approved or later canonical revision under active QA
- creation path: real runtime flow through modern documents + approval
- safe reuse: conditional
- advancement path: may move into scheduled or published depending on active QA work
- discard rule: mint a fresh fixture if clean approved pre-publish proof is required and current row has already advanced
- caveats: not suitable as a permanent baseline if the session consumes it into later states

### Fixture: previously published head

- controlled document ID: `750afeba-6e35-4dd4-8a74-2b51b9f9090c`
- document ID: `621f27f0-5006-40b5-b84ba-1ba86f669625`
- represented state: previous published head before REV10 handoff
- creation path: real runtime lineage
- safe reuse: read-only reference
- advancement path: none
- discard rule: keep as lineage reference unless superseded truth changes
- caveats: use as historical lineage reference, not as a clean active-flow fixture

### Fixture class: cross-tenant live authz target

- controlled document ID: not yet provisioned
- represented state: wrong-tenant route access validation
- creation path: pending dedicated multi-tenant fixture support
- safe reuse: not applicable
- advancement path: none
- discard rule: not applicable
- caveats: current sessions must classify lack of live target as a tooling or fixture blocker, not as product proof

### Fixture class: snapshot-corruption finalize target

- controlled document ID: not yet provisioned
- represented state: draft with invalid or missing required snapshot
- creation path: pending controlled fault-injection fixture
- safe reuse: not applicable
- advancement path: none
- discard rule: not applicable
- caveats: use injected contract proof until a safe runtime fixture exists
