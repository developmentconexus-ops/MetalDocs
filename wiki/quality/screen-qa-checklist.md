# Screen QA Checklist

> **Last verified:** 2026-05-27
> **Scope:** Reusable QA checklist for user-facing screen work in MetalDocs.

## When to use

Use this checklist for any screen, page, modal, wizard, or major interaction change where the user-visible result is part of the definition of done.

Pair it with:

- [qa-operating-system.md](qa-operating-system.md)
- [deep-qa/index.md](deep-qa/index.md) when a deeper module-specific matrix already exists

## Entry gate

Before QA starts, confirm:

- target route/runtime truth is trustworthy
- auth/session state is known
- touched backend/API/query prerequisites are either passing or explicitly classified

## Checklist

- initial load state is correct and honest
- happy path completes and the final visible state is correct
- empty state is intentional and not a broken-data placeholder
- validation states show the right message at the right time
- loading, disabled, and submitting states prevent misleading interaction
- server/network failure states are visible and recoverable
- stale cache, refresh, and return-navigation behavior stay correct
- auth and authz outcomes are correct for the acting role
- optimistic or delayed UI state settles to persisted truth
- copy, labels, badges, and status states do not overstate success
- linked navigation paths land on the correct destination
- browser refresh or route re-entry does not corrupt the workflow
- any touched accessibility-critical interaction still works at a basic level

## Evidence expectation

Record:

- screen or route exercised
- acting user/role
- runtime proof or focused artifact used
- persisted/API proof when the screen changes state
- explicit failures, defers, or skipped paths with reason

## Escalation rule

Stop the screen close-out loop and classify the issue when:

- the screen depends on a missing backend capability
- runtime and contract truth disagree
- user-visible success is shown before persisted truth exists
