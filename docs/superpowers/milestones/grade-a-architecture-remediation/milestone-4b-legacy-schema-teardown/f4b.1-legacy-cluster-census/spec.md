# Feature F4b.1 — legacy `metaldocs.documents` cluster verify-dead census

> **Milestone:** 4b (Legacy schema cluster teardown)  ·  **Feature:** `f4b.1-legacy-cluster-census`
> **Approved:** 2026-06-15 (operator standing authorization; first feature of 4b, scope = read-only census).

## Consumer contract

The **consumer** is F4b.2's drop migration. Before it may `DROP` anything it requires a
**verified-dead manifest**: the exact, complete set of legacy objects to drop, each with proof of
**zero live dependents**. The contract:

1. **Completeness** — the manifest names every object in the `metaldocs.documents` legacy cluster
   (the anchor table, every FK-satellite at any depth, and the independent `metaldocs.template_audit_log`
   duplicate). A satellite discovered mid-census (e.g. a 2nd-level FK child) MUST be added, not ignored.
2. **Zero-dependent proof per surface** — for the whole set: zero runtime Go references; no inbound FK
   from any **kept** (non-dropped) table; no view, trigger, RLS policy, stored function, sequence, or
   reference-data seed referencing any cluster object.
3. **Fail-closed (HS-2)** — if **any** live dependent is found, the cluster is not dead: stop, report
   the boundary, do **not** authorize the drop.
4. **Public twins are out of scope** — `public.documents` and `public.template_audit_log` are the real,
   live tables and are never in the manifest. The census is `metaldocs.`-qualified throughout.

## Non-goals

- **Not** writing or running the drop migration (that is F4b.2).
- **Not** changing any code, schema, or test.
- **Not** auditing the live `public.documents` governance model or the MDDM export feature flag —
  only proving the *legacy metaldocs tables* are dead.

## Validation Gate

| # | Acceptance | Proof |
|---|-----------|-------|
| A | Manifest is complete (anchor + all satellites at every FK depth + `template_audit_log`) | inbound-FK sweep on the curated baseline closes (no `REFERENCES metaldocs.<cluster>` from outside the set) |
| B | Zero runtime Go refs for every manifest object | `grep` per table over `internal/`,`apps/` (non-test, non-vendor) = 0 |
| C | No inbound FK from a kept table | every `REFERENCES metaldocs.<cluster>` originates from a table that is itself in the manifest |
| D | No view / trigger / RLS / function / sequence / reference-data dependent | targeted `grep` over baseline + migrations + reference-data = none |
| E | HS-2 not tripped | A–D all clear; manifest recorded in `evidence.md` |

## Interview record

No operator interview needed — the census is read-only and the contract is read directly from F4b.2's
need (a verified-dead manifest). Operator did intervene mid-census to flag MDDM as known legacy and ask
for a sanity check; that disambiguation (MDDM legacy *tables* dead vs MDDM export *feature flag* alive)
is recorded in `evidence.md`.
