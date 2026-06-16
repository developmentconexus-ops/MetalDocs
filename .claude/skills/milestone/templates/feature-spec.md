# Feature <id> — Spec

> **Milestone:** <n> — <title>  ·  **Folder:** `f<n>.x-<slug>`
> **Status:** Drafting | Approved (pre-code) | Superseded
> **Approved before code:** <date / operator> — *no implementation begins until this line is filled.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Before writing the contract below, resolve ambiguity. **Engine: `superpowers:brainstorming`** —
invoke it with the feature row from `milestone.md` as the seed; it drives the consumer-contract
discovery dialog (intent → constraints → tradeoffs → contract shape). Persist the resulting Q&A
into the table below. **Skill absent → run the dialog inline, one question at a time, and persist
the same Q&A.** Validator C1 reads this table to confirm the contract was discovered, not guessed.

If you genuinely needed no interview (trivial feature, contract already explicit in the consumer's
existing code), record a single row with `none needed` + the reason — that line IS the evidence
for C1.

| # | Question | Answer |
|---|----------|--------|
| 1 | <question, or "none needed — why"> | |

## Consumer contract (FIRST — before any producer)

What depends on this feature, and the exact shape that dependency requires. Define the **consumer's**
expectation, then build the producer to match it — never the reverse.

- **Consumer(s):** <who calls this / reads this — route, hook, module, screen, downstream job>
- **Contract:** <the exact interface / response shape / event / schema the consumer relies on —
  fields, types, status codes, ordering, error envelope. Read it from the consumer; do not invent it.>
- **Source of truth for the contract:** <where this shape is canonically defined — OpenAPI op,
  generated type, ADR, existing caller>

## What this feature implements

<The change, stated by outcome. Producer satisfies the consumer contract above.>

## Non-goals (mandatory)

Explicitly out of scope. Anything here that later appears in the diff is scope drift (validator C6).

- <non-goal>

## Validation Gate (concrete — approved before code)

How this feature is proven done. Must be objectively checkable — a named test that passes, a route
that responds with the contracted shape, a command whose output is asserted. No "works" / "looks right".

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| <criterion tied to the consumer contract> | `<test name / cmd>` | real / fixture |

> TDD: write the failing test first, then implement to green. Fixture-only proof must be labeled —
> it is not end-to-end proof of the real provider.

## ADR needed?

- [ ] No durable decision — skip.
- [ ] Durable decision made → record an ADR under `wiki/decisions/` and link it here: <adr link>
