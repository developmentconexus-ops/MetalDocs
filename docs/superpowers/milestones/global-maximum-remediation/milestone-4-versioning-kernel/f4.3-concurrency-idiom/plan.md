# F4.3 plan

Executed via subagent; main reviews + commits. Contract-first.

## Task
Migrate templates optimistic-concurrency transport from body `expected_lock_version` to the `If-Match`
header, matching documents' idiom. Land an ADR.

Files:
- `api/openapi/**` — templates mutating write endpoint(s) that today declare a body `expected_lock_version`
  gain an `If-Match` header precondition; drop the body field. Then regenerate BE (`api.gen.go`) + FE
  (`api-types`) — ZERO hand-edits to generated files.
- `internal/modules/templates/delivery/http/routes_schema.go` (~32-57) + handlers — parse `If-Match`
  (reuse the documents `parseIfMatch` shape from `documents/approval/http/handler.go:145-164`, or extract
  a shared helper) → map to `lock_version`. CAS `WHERE lock_version=$N` UNCHANGED.
- FE template-edit consumers — send `If-Match` header instead of the body field.
- `wiki/decisions/NNNN-optimistic-concurrency-if-match.md` — ADR: decision, RFC 7232 / AIP-154 / Zalando
  basis, templates cutover, pre-v1 atomic-cutover exception. Cite it in the commit.

Header value grammar: documents use `If-Match: "vN"`. Apply the same grammar uniformly; the shared helper
handles it. ADR records the exact on-the-wire value format so both modules match.

## Gate
`oapi-codegen` + FE typegen clean, zero hand-edits · openapi lint green · templates handler parses header,
`lock_version` CAS intact · `tsc --noEmit` + targeted vitest (templates) green · `go build ./...` green.

## Fallback (HS-7-gated)
If a hard blocker makes the header cutover unsafe within M4, ADR-record the split as intentional instead,
and re-open contract §3 with operator approval. Not expected.
