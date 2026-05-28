# Backend/API QA Checklist

> **Last verified:** 2026-05-27
> **Scope:** Reusable QA checklist for backend HTTP behavior, contract-backed routes, and integration-visible API changes.

## When to use

Use this checklist for any change to public or shared backend behavior, including route handlers, OpenAPI-backed surfaces, authz-visible outcomes, validation, and response semantics.

Pair it with:

- [qa-operating-system.md](qa-operating-system.md)
- [release-readiness.md](release-readiness.md) for final merge/release gate

## Entry gate

Before QA starts, confirm:

- route ownership is known
- runtime/spec/generated-surface alignment was checked
- target environment is running fresh code

## Checklist

- happy-path request returns the expected status and shape
- validation failures return the intended code and problem semantics
- auth and authz allow/deny behavior is correct
- not-found, conflict, and invariant failures are honest and stable
- persistence side effects match the returned response
- no legacy route, prefix, or wrapper drift was reintroduced
- OpenAPI and generated surfaces still match runtime truth when the contract changed
- idempotency, retry, or duplicate-submit behavior is correct where relevant
- logging/observability-critical failures are still diagnosable
- shared consumers are not silently broken by a local route change

## Evidence expectation

Record:

- route and method exercised
- request identity/role used
- runtime API proof, or explicit fallback proof type when runtime is disproportionate
- persisted-state or downstream-side-effect proof when behavior mutates state
- exact command, fixture, or artifact used for validation

## Escalation rule

Stop local close-out and classify the issue when:

- runtime, OpenAPI, generated backend surfaces, or generated frontend types disagree
- the fix requires a shared contract or route-prefix redesign
- one handler fix would leave sibling consumers on a lying contract
