# F4.3 — concurrency-idiom unification

> **Contract:** `../validation-contract.md` §3 (full analysis + decision). Feature home.
> **Approved for code: 2026-07-04.**

## Consumer contract

- **Consumer = API clients + the generated FE client.** Require ONE optimistic-concurrency wire idiom
  across the two kernel modules. Decision (contract §3.4): **unify on the `If-Match` header** (RFC 7232 /
  AIP-154 / Zalando-endorsed HTTP-native precondition). Documents already use it; **templates migrate**
  from the body field `expected_lock_version` to the `If-Match` header, contract-first.
- The `templates_template_version.lock_version` column + its CAS semantics are UNCHANGED — only the
  transport moves from request body to header.

## Non-goals

- NOT renaming the `lock_version` column (internal; contract §6 defer).
- NOT changing documents (already on `If-Match`).
- NOT the state machine (F4.1) or race (F4.2).

## Validation gate

Per contract §3.6. Templates mutating write endpoints take `If-Match` (openapi + regen, zero hand-edits);
handler parses header → `lock_version` CAS preserved; FE template-edit consumers send the header; ADR
landed + cited; `tsc --noEmit` + targeted vitest + `go build ./...` + openapi lint green. Fallback
(HS-7-gated): if migration surfaces a hard blocker, ADR-record the split instead (contract §3.5).

## Interview record

| Q | Operator answer |
|---|---|
| Idiom choice | "Best solution — full analysis: what we have, what we want, a fresh professional impl." → delegated to the §3 analysis, which decided **unify on If-Match + migrate templates + ADR.** |
