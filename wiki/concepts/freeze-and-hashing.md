# Concept: Freeze and Hashing

> **Last verified:** 2026-05-03
> **Status:** Stub. Verify exact hash algorithm + canonicalization rules against domain code.
> **Scope:** What freeze produces, what the three hashes prove, why immutability matters.
> **Out of scope:** Full freeze pipeline (see `workflows/freeze-and-fanout.md`).

## What freeze does

When a document version's approval condition is satisfied, the freeze service runs (in the same transaction as the final signoff) and:

1. Resolves the 7 fixed tokens to their final values.
2. Substitutes the values into the eigenpal-native DOCX format.
3. Uploads the resulting frozen DOCX to MinIO.
4. Records `values_frozen_at` timestamp.
5. Computes three hashes (below).
6. Marks the version `frozen` and immutable.

## The three hashes

| Hash             | What it covers                                            | Why                                                |
|------------------|-----------------------------------------------------------|----------------------------------------------------|
| `content_hash`   | The frozen DOCX bytes (post-substitution)                 | Proves the artifact stored is the one approved     |
| `values_hash`    | The resolved token values (canonicalized JSON)            | Proves which values were substituted               |
| `schema_hash`    | The placeholder schema snapshot at document creation time | Proves which token catalog applied at freeze       |

Storing all three lets an auditor verify:

- The DOCX in MinIO matches what was approved (`content_hash`).
- The substitution used the right values (`values_hash`).
- The fixed-7-token catalog applied (`schema_hash`).

## Immutability

Once `values_frozen_at` is set, the version is immutable:

- API rejects content edits.
- UI puts the editor in read-only mode.
- Re-running freeze is a no-op (early return).

To produce a new approved version → create a new revision (which clones the frozen content into a new draft).

## See also

- [workflows/freeze-and-fanout.md](../workflows/freeze-and-fanout.md) — full pipeline
- [concepts/placeholders.md](placeholders.md) — what the 7 tokens are
- [decisions/0008-placeholder-fixed-catalog.md](../decisions/0008-placeholder-fixed-catalog.md)
