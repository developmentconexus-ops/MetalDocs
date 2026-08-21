# T8-F Fable Round 2

> Evidence only. Candidate authority remains `arch/t8f-frontend-realization`; this review branch must never merge.

## Lead handoff

Repository: `developmentconexus-ops/MetalDocs`

Candidate under review:

```text
arch/t8f-frontend-realization
e54986904063c982315129635191ebade8f9b9ed
```

Review branch: `review/t8f-fable-r2`

Reconstruct authority fresh:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ docs/architecture/frontend.md
→ only the smallest additional owner needed for a finding
```

Round 1 Evidence PR #140 produced F1–F6. Lead accepted all six; operator approved the correction package. Do not re-review T8-F broadly unless a correction caused a concrete regression.

## Bounded Round-2 target

Try to falsify only these corrections and their immediate regressions.

### A. T8-E-FR read symmetry

Review `docs/decisions/frontend-read-symmetry.md` and the corrected Document Official contract.

Attack:

```text
open_revision?: { revision:RevisionIdentity, state:OpenRevisionState }
active_obsolescence_request_id?: Uuid
```

Verify:

- they solve direct `/documents/:document_id/work` resolution and active-obsolescence discovery;
- no operation 79 is needed;
- they are derived read truth, not persisted Document pointers;
- T8-D can derive each uniquely from existing state;
- disclosure does not leak DRAFT/SUBMITTED or active-request existence to callers lacking the relevant context;
- absence is not interpreted as proof of semantic non-existence when disclosure is denied;
- follow-up commands still reauthorize current truth;
- no new T8-C contract class or Product/T6 capability is smuggled in.

### B. Navigation / Authorization correction

Verify the corrected shell law is coherent with `SessionView {user, csrf_token}`:

```text
navigation presence is not permission-filtering authority
server 403/404 remains authoritative for direct lens entry
no browser-maintained permission matrix
```

Flag only a concrete usability/security contradiction; do not demand permission snapshots by convention.

### C. Recovery corrections

Verify:

```text
state.upload_expired
  -> preserve local intended bytes
  -> new allocation
  -> upload/complete again
  -> never revive/reuse expired allocation

permission.csrf_failed
  -> getSession / fresh csrf_token
  -> retry only the same safe logical command
  -> preserve Idempotency-Key and conditional semantics
```

Attack for accidental duplicate semantic commands, stale If-Match reuse, or retry loops.

### D. Audit subtraction

Verify `/audit` now promises inspection/paging only and does not invent filtering absent from operation 78.

### E. Regression checks

Reconfirm only:

```text
78 / 78 operation coverage
operation 79 absent
stable T6 route meanings unchanged
no new semantic owner/Permission/lifecycle
no frontend AuthZ engine or parallel DTO authority
T8-G remains unopened / runtime mechanics deferred
implementation remains blocked
```

## Output

Write only below `## Fable response`.

For each surviving finding:

```text
ID
severity: MATERIAL | MINOR
claim
owning authority implicated
concrete counterexample
smallest correction
reopen required? yes/no + exact owner
```

If no MATERIAL finding survives, say **CONVERGED** explicitly and list the strongest attacks attempted.

Do not edit any other file. Do not ratify T8-F. Do not open T8-G.

## Fable response
