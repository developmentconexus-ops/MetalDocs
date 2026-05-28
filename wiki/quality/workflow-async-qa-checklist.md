# Workflow / Async QA Checklist

> **Last verified:** 2026-05-27
> **Scope:** Reusable QA checklist for workflows that cross request boundaries, workers, outbox processing, delayed completion, or scheduled state changes.

## When to use

Use this checklist when "request succeeded" is not enough to prove the final product outcome.

Examples:

- outbox or worker-owned completion
- scheduled publish/unpublish
- multi-step approval or state-machine flows
- webhook or callback-driven completion

## Required proof split

QA must prove each stage separately:

1. request accepted
2. state persisted
3. async owner executed
4. user-visible final truth updated

## Checklist

- request acceptance is distinguishable from final completion
- persisted intermediate state is correct
- worker/outbox/scheduler ownership is known
- async execution happens once or with defined duplicate safety
- failure or retry states are visible and diagnosable
- user-visible status does not claim completion too early
- refresh/re-entry after async completion shows final truth
- authorization still holds across delayed execution
- any notifications, follow-on jobs, or downstream writes are correct
- time-based or scheduled behavior is validated with honest evidence labeling

## Evidence expectation

Record:

- trigger action
- persisted intermediate proof
- async execution proof
- final user-visible proof
- whether the proof was live-runtime, injected, or synthetic

## Escalation rule

Stop local closure and classify the issue when:

- the system cannot prove which async owner is responsible
- final truth depends on unverified background behavior
- a local UI/API patch would hide an outbox, worker, or scheduler defect
