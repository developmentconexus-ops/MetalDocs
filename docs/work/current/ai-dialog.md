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

> Bounded Round-2 challenge of candidate `arch/t8f-frontend-realization @ e54986904063c982315129635191ebade8f9b9ed`. Authority reconstructed fresh: `AGENTS.md → docs/index.md → docs/roadmap.md → docs/architecture/frontend.md → docs/decisions/frontend-read-symmetry.md`, with `docs/architecture/wire-contract.md` and `docs/architecture/persistence.md` consulted for specific verification. Evidence only.

### Verdict

**CONVERGED.** No MATERIAL finding survives against the F1–F6 corrections or their immediate blast radius. Non-material observations are listed for the record; none blocks ratification.

### A. Read-symmetry precision — attacked, holds

- **Resolution completeness (F1/F3 closure).** `open_revision?: {revision: RevisionIdentity, state: OpenRevisionState}` resolves `/documents/:document_id/work` directly from operation 47 on cold load, and the `state` member correctly disambiguates the two work sub-flows (DRAFT → `getRevisionDraft`; SUBMITTED → `getRevision`/`getSubmission`). `active_obsolescence_request_id?` resolves inspect/withdraw via operation 76/77. §6.7 now explicitly bans History/My Work as current-resource resolvers, closing the Round-1 archaeology hazard. No operation 79; census effect verified as zero (operation 47 response precision only).
- **Executability claim re-executed against T8-D.** The decision doc's "no persistence reopen" assertion depends on uniqueness that I verified exists verbatim in ratified persistence: `UNIQUE(document_id) WHERE state IN ('DRAFT','SUBMITTED')` (persistence.md §controlled_docs.revisions) and `UNIQUE(document_id) WHERE state='ACTIVE'` (persistence.md §controlled_docs.obsolescence_requests). Both members are therefore uniquely derivable read truth; persistence.md's "no current-status pointer on Document" law is untouched.
- **Disclosure attack.** Presence is gated on existence AND disclosure; absence is explicitly non-probative for callers lacking disclosure authority. I attacked the affordance consequence: could a caller entitled to act (create-next-Revision) be ambiguous on absence? No — every fixed T3 role bundle carrying `document.create`/`document.edit` also carries `document.read_working` (author, area_manager), so the acting population always has working-context disclosure and absence is authentic for them; pure viewers get no work affordances anyway. Race between a stale view and `createDocumentRevision` still lands on `409 state.conflict` → refetch, per §10.3. No leak: `DocumentSummary`/Library list shape is untouched, so open-work existence does not bleed into discovery results.
- **Smuggling attack.** No new operation, Problem code, header profile, Permission, T8-C contract class, or persisted pointer. The disclosure predicate is stated abstractly ("current disclosure/Authorization") — consistent with the wire's own law of not restating AuthZ predicates, and composable from existing predicates (`document.read_working`; operation-76 read disclosure). ETag safety: `DocumentOfficialView` is `JSON_NO_STORE`, not ETag-protected, so adding independently mutable members violates no concurrency-domain law.

### B. Navigation correction — holds

§6.1 is coherent with `SessionView {user, csrf_token}`: presence is not permission-filtering authority, server 403/404 stays authoritative, no permission matrix. No concrete usability or security contradiction found — SPA route names are public bundle knowledge, so always-present navigation discloses nothing.

### C. Recovery corrections — hold

- **`state.upload_expired`** (§6.4, §10.3): preserve local bytes → new allocation → re-upload → complete → attach under current DRAFT ETag; never revive expired capability. Matches the wire's closed expiry semantics; orphaned expired uploads remain reclaimable; a draft that changed meanwhile still lands on `412 precondition.draft_changed` reconciliation — no silent overwrite path opened.
- **`permission.csrf_failed`** (§10.3): re-bootstrap then same-command retry with preserved Idempotency-Key and conditional semantics. Duplicate-command attack fails structurally: the wire's request-processing precedence checks CSRF (stage 4) before idempotency replay and business effect, so a csrf-failed request had zero semantic effect and the retry cannot duplicate. Loop risk is bounded: persistent failure degrades to `getSession` → 401 → login flow.

### D. Audit subtraction — holds

§3 now says "page meaningful AuditEvent evidence"; §6.11 pins inspection/paging only and explicitly refuses inventing client- or server-side filters absent from operation 78. Library "search/filter" wording correctly remains (listDocuments has admitted filters).

### E. Regression checks — re-executed

```text
coverage partition      §5 ranges = 1–2, 3–26, 27–33, 34–43+50–51, 44–46,
                        47–49+52+72–73+75+77, 53–56+76, 57–66, 67–71, 74, 78
                        -> exact union 1..78, no dup/gap; identical to the
                           Round-1 mechanically verified partition
operation 79            absent; every mention is a prohibition
route set               §4 == accepted T6 ten routes, verbatim, unchanged
semantic owners         none added; new members are derived read references
frontend AuthZ engine   absent (§6.1/§14 unchanged in substance)
parallel DTO authority  absent (§9 unchanged in substance)
T8-G                    §17 consumers-only, unchanged; NOT OPEN
implementation          BLOCKED (roadmap + §18)
registry routing        decisions/index.md row T8-E-FR + docs/index.md router
                        point at frontend-read-symmetry.md; supersession is
                        explicit and bounded to the DocumentOfficialView
                        member set (census-precision precedent)
required CI             run 32495703178 SUCCESS on exact candidate HEAD
                        e54986904063c982315129635191ebade8f9b9ed (re-checked)
```

The Round-2 compression of §§6.8–6.10 (dropping restated per-operation ETag/idempotency detail) removes only text the wire already owns; the ETag-domain independence and role-fixity laws survive. No Round-1 property was lost in the rewrite.

### Non-material observations (no action required for ratification)

1. `journeys.md` §5 still reads "Audit when authorized / Administration when authorized" while §6.1 adopts always-present navigation. The lens-usability reading is defensible and operator-approved; noting it here only so the unamended Product wording is not re-litigated later as if unexamined.
2. The two new members' presence is AuthZ/disclosure-dependent rather than cross-field, so they fall outside the wire's `FIXTURE_PRESENCE` registry by the same precedent as `DocumentCreationOptionsView.responsible_owner_candidates`; a future OAS expansion should carry the disclosure law as prose + fixture, not schema.
3. `wire-contract.md` §3.5 still prints the pre-precision `DocumentOfficialView`; the supersession is explicit in the decision doc and routed in both indexes — acceptable, same pattern as the census supersession of journeys §29.
4. Candidate history contains two empty no-op commits (`e5a6c79d`, `81fc67fc`); final tree verified clean (5-file delta, one decision file). Squash-merge erases the noise.

### Strongest attacks attempted (all failed)

- T8-D derivability fraud: uniqueness constraints re-read from ratified persistence — both exist verbatim.
- Disclosure/affordance incoherence for acting roles: defeated by fixed T3 bundle composition (create/edit ⊂ read_working carriers).
- Existence leak into Library/list surfaces: `DocumentSummary` untouched.
- Duplicate-command via csrf retry: defeated by wire precedence (CSRF before replay/business effect).
- Expired-allocation revival or stale If-Match reuse in the upload restart law: text forbids both; OCC path preserved.
- Hidden operation/permission/Problem/T8-C class in the precision: census/problem/header/permission deltas all zero.
- Coverage/route/owner regressions from the §5/§6 rewrite: partition and route set re-verified exact.

Do not treat this Evidence as authority. Lead adjudication and operator ratification remain the exit; T8-G remains unopened.
