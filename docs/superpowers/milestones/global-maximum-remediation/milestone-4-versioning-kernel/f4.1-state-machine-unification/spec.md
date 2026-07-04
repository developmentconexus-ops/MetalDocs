# F4.1 — state-machine unification + `rejected` removal

> **Contract:** `../validation-contract.md` §1 (the binding transition table + parity + removal scope).
> This spec is the feature home; the deep acceptance lives in the contract. **Approved for code:
> 2026-07-04** (operator locked the three decisions in the M4 interview; runtime-truth discovery done).

## Consumer contract (who depends on what)

- **Consumer = the approval services** (`documents/approval/application/*`). Today each re-encodes "is this
  status move legal" as an ad-hoc `if status != X` guard + an OCC `WHERE status='<cur>'` clause. They
  require **one** authority to ask "may this document move `cur → next`?" — the unified
  `documents/domain` transition function — and keep their OCC `WHERE` as the atomic CAS.
- **Consumer = the DB trigger** `enforce_document_transition` (the last line). It requires the app
  function to be its **exact mirror** (parity test, contract §1.4) so friendly-first-line and last-line
  never drift.
- **Consumer = the wire/UI** (OpenAPI status enum, FE status rendering). After `rejected` removal they
  require the enum + FE to no longer list/handle `rejected`.

## Non-goals (mandatory)

- NOT adding routes for the DB-legal-but-app-unused arcs (`approved→draft`, `scheduled→draft`); they are
  included as *legal* in the function for parity only (contract §1.2 note ²).
- NOT removing `archived` (ADR 0010 soft-archive flag; contract §6 defer).
- NOT moving transition enforcement out of the DB; the trigger stays authoritative.
- NOT touching the approval-**instance** state machine (`InstanceApproved` etc.) — different FSM.
- NOT the concurrency idiom (F4.3) or the publish race (F4.2).

## Validation gate

Per contract §1.6. Summary: unified exhaustive fn (typed error, total, mirrors templates) · legal set ==
§1.2 table · every §0.3 service routes through it, scattered-guard census = 0 (or allowlisted) · dead
`CanTransitionDocument` deleted · **app↔DB parity test green** · coverage test green · `rejected` removed
full-stack (enum + reader + CHECK + trigger + OpenAPI regen zero-hand-edit + FE + tests) each with proof ·
`go build ./...` green · targeted go + vitest green · openapi lint green · M0–M3 not regressed.

## Interview record

| Q | Operator answer |
|---|---|
| `rejected` disposition | Full removal now, inside M4 (it is dead: reject writes `draft`). |
| `scheduled` disposition | Keep it, one-way (`scheduled→published` cutover). |
| Authority for legality | One function mirroring the DB trigger (global-max: kill split-brain, not a 4th guard). |

## Task decomposition (see plan.md)

A = domain function + coverage test + wire services + delete dead FSM (Go).
B = `rejected` removal: DB migration (CHECK + trigger, row precheck) + parity test + Go enum/reader/tests.
C = `rejected` removal: OpenAPI enum + regen BE/FE + FE status files.
