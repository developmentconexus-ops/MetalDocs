# Feature F9.4 — widen noresponsemap to any map[string]<T>

> **Milestone:** M9 — HS-5 contract type-gate  ·  **Folder:** `f9.4-noresponsemap-widen`
> **Status:** Done

## Source

- Milestone spec row: widen `tools/cilint` `noresponsemap` to flag **any** `map[string]<T>` response
  literal reaching a 2xx writer (not just `map[string]any`); update `api-contract.md` §5b.
- Governing-spec reference: mission.md §8 H-D class + §8 scope amendment (close the class, not instances).

## Plan

1. In `noresponsemap.go`: rename/rewrite `isMapStringAnyLiteral` → `isMapStringLiteral`, matching any
   `map[string]<T>` (key == "string", any value type). Keep scope (`inRegisteredRoutePackage`),
   exemptions (health.go), and `//cilint:allow-responsemap` unchanged. Update finding message + doc comment.
2. TDD: positive `map[string]string` reaching a writer must flag; negative `map[string]string` not reaching
   a writer must pass; existing typed-struct / allow-directive / exempt cases stay green.
3. Run the widened analyzer repo-wide — expect it to surface any newly-covered site, then fix that site
   (do NOT suppress; a response-shaped exemption is forbidden by §5b).
4. Update §5b: rule now "no `map[string]<T>` response literal"; Part A/B greps widened; anti-evasion note.

## Execution notes

- Widening surfaced a 4th site, `finalizeDocument` (`map[string]string{"instance_id": ...}`). Closed it by
  adding `DocumentFinalizeResult {instance_id: uuid}` to the spec, regenerating, and routing the handler
  (both the idempotency-replay marshal and the final write) through the typed struct — parsing the
  service's instance-id string to uuid with a 500 on failure. This is the class-closure the widening exists
  to force, not an instance whack-a-mole.
- Test fakes returning non-uuid instance ids (`inst_1`, `inst-test`) updated to real uuids — runtime truth
  is "UUID of created approval_instance" (`submit_service.go`). The verbatim idempotency-replay test is
  unaffected (it replays stored bytes, not a re-marshal).
