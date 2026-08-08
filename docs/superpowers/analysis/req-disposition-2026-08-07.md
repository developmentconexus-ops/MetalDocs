# Disposition — the four uncovered MUST requirements

**Measured:** `go run ./scripts/req-trace` →
```
UNCOVERED MUST REQ(s) (4):
  REQ-AUTHN-1
  REQ-AUTHN-3
  REQ-SEARCH-1
  REQ-SEC-3
reported (SHOULD, non-blocking, no evidence): REQ-TOP-4
reported (SHOULD, non-blocking, no evidence): REQ-IAM-2
reported (SHOULD, non-blocking, no evidence): REQ-REL-4
reported (SHOULD, non-blocking, no evidence): REQ-REL-5
reported (SHOULD, non-blocking, no evidence): REQ-SEC-5
67 REQ IDs (61 MUST, 6 SHOULD, 0 MAY); 4 MUST uncovered; stale=false
exit status 1
```

**Finding:** none of the four is a traceability gap. Three REQs describe a system
MetalDocs did not build; one is a process requirement the gate's evidence model
cannot express.

**Not done here:** no entry was added to `wiki/architecture/req-trace-map.yaml`.
That file's header declares the anti-gaming boundary (`kind` MUST be `commit`;
doc-annotation evidence is auto-derived and never sourced from this map); using
it to silence a REQ whose behaviour is absent would be gaming the gate.

## REQ-AUTHN-1 — code defect, not spec defect

**REQ text** (`wiki/architecture/backend-target-architecture.md:121`):
> - **REQ-AUTHN-1** Passwords hashed with a memory-hard KDF (Argon2id family);
>   verification constant-time; failure responses identical for unknown-user vs
>   wrong-password. (MUST)

**Code evidence:**
- `internal/modules/auth/application/service.go:38-39` — `passwordAlgoBcrypt = "bcrypt"`,
  `bcryptCost = 12`.
- `internal/modules/auth/application/service.go:1025` and `:1038` — password hashing calls
  `bcrypt.GenerateFromPassword(password, bcryptCost)`.
- `internal/modules/auth/infrastructure/postgres/repository.go:544` —
  `password_algo = CASE WHEN $3 IS NULL THEN password_algo ELSE 'bcrypt' END` — the write
  path hard-codes the literal `'bcrypt'`, not a parameterized algorithm.
- `internal/modules/approval/infrastructure/signature/password_reauth.go:10,67,106-107` —
  the approval-signature reauth path also compares against bcrypt hashes
  (`bcrypt.CompareHashAndPassword`).
- `db/baseline/0001_current_schema.sql` already carries a `password_algo` column
  (referenced at `repository.go:56,62,77,417,544,621`), so an algorithm migration has a
  column to drive off already — no schema change needed to start.
- `grep -rn 'argon2\|bcrypt' --include=*.go internal/ apps/ | grep -v vendor` returns
  bcrypt hits only, across `auth`, `approval`, `iam`, and their test files. Zero `argon2`
  matches anywhere in first-party code.

**Verdict — this is a genuine gap, not a documentation mismatch.** The REQ demands a
memory-hard KDF; the running system uses bcrypt (not memory-hard) everywhere a password
is hashed or verified. The `password_algo` column being schema-ready but hard-coded to
`'bcrypt'` at the one write site shows the migration path was anticipated but never
finished.

**Recommended disposition:** implement Argon2id with a `password_algo`-driven
rehash-on-login migration (new logins/rehashes write `argon2id`; existing bcrypt hashes
verify against bcrypt until they naturally rehash on next successful login). Do **NOT**
amend REQ-AUTHN-1 to accept bcrypt — bcrypt is a legitimate password hash but is not
memory-hard, and relaxing the requirement to match the code is exactly the move the
operator's standing doctrine ("we will never bend a rule to fit something bad") forbids.
If the operator judges bcrypt-cost-12 adequate given other compensating controls
(rate limiting, timing-safe dummy-hash comparison already present at
`service.go:174-175,277`), that is a security-risk-acceptance call for the operator to
make explicitly — not something this task enacts by editing the REQ.

**Owner decision required:** schedule the Argon2id migration. Not a CI-restructure
blocker.

## REQ-AUTHN-3 — spec defect

**REQ text** (`wiki/architecture/backend-target-architecture.md:123`):
> - **REQ-AUTHN-3** Token handling follows RFC 8725 (alg pinning, no `none`,
>   audience/issuer checks, short TTL). (MUST)

**Code evidence (proving absence, not asserting it):**
- `grep -rln 'jwt' --include=*.go internal/ apps/ | grep -v vendor` → empty output,
  exit code 1 (grep's "no matches" exit code). No file under `internal/` or `apps/`
  (excluding vendor) contains the string `jwt` in any Go source.
- Session/auth flow instead centers on `internal/modules/auth/application/service.go`,
  which issues and validates opaque server-side session state (`PasswordHash`,
  `PasswordAlgo`, lockout counters, `dummyHash` timing defense) — no token encoding,
  signing, or claim-validation logic of any kind is present in that file or its
  neighbors (`internal/modules/auth/infrastructure/postgres/repository.go`,
  `internal/modules/auth/delivery/http/middleware*.go`).

**Verdict — the REQ names a technology the architecture rejected.** There is nothing in
the tree for RFC 8725's alg-pinning, `none`-algorithm rejection, or audience/issuer
checks to apply to, because there are no JWTs to check. This is not an implementation
gap that can be closed by writing token-validation code — doing so would mean building
a JWT subsystem the system does not otherwise need, just to satisfy a REQ's letter. The
REQ's *intent* (short-lived, non-forgeable, revocable tokens) is arguably already met by
opaque server-side sessions, which are immune to the entire alg-confusion/`none`-algorithm
class of JWT bugs by construction.

**Recommended disposition:** an ADR amending REQ-AUTHN-3 to describe opaque server-side
session tokens, preserving the equivalent security properties the RFC-8725 text was
reaching for: short TTL, rotation on privilege change, and server-side revocation. The
REQ names a technology the architecture rejected; it is the REQ that is wrong, not the
code.

**Owner decision required:** ratify the ADR retiring/rewriting REQ-AUTHN-3. Not a
CI-restructure blocker.

## REQ-SEARCH-1 — spec defect, with a live design question

**REQ text** (`wiki/architecture/backend-target-architecture.md:225`):
> - **REQ-SEARCH-1** Search indexes are derived and rebuildable; a full reindex
>   procedure exists and is tested. Search is never consulted for authz decisions. (MUST)

**Code evidence:**
- `grep -rn -i 'reindex\|tsvector' --include=*.go internal/modules/search/` → no matches.
  No reindex procedure and no Postgres `tsvector`-backed derived index exist anywhere in
  the search module.
- `internal/modules/search/infrastructure/v2documents/reader.go:75` —
  `AND ($2 = '' OR LOWER(COALESCE(d.name, '')) LIKE '%' || $2 || '%' ESCAPE '\')` — search
  is a `LIKE`/`ILIKE`-style query directly against the live `documents` table, not a
  derived index.
- `reader.go:14,148` — the escaping is delegated to
  `internal/platform/sqlescape.LikeEscape(...)`, confirmed by the dedicated
  `internal/modules/search/infrastructure/v2documents/reader_like_escape_test.go`.
- Directory listing of `internal/modules/search/infrastructure/v2documents/`:
  `reader.go`, `reader_contract_parity_integration_test.go`,
  `reader_family_integration_test.go`, `reader_like_escape_test.go`, `reader_test.go`,
  `reader_visibility_integration_test.go` — there is no `reindex.go`,
  `index_builder.go`, or equivalent. Nothing exists to rebuild.
- The REQ's second clause — "Search is never consulted for authz decisions" — is
  independently testable and appears to hold: `reader_visibility_integration_test.go`
  and `reader_contract_parity_integration_test.go` exercise the reader against the same
  visibility/authz filtering as the primary document read path (parity test explicitly
  checks the seam changes "no authz/visibility/ordering" per its own comment at
  `reader_contract_parity_integration_test.go:24`), i.e. search reuses the authz-scoped
  query rather than deciding authz itself.

**Verdict — the REQ's first clause is a spec defect (describes an index that was never
built); the second clause is true and already tested.** Splitting the REQ is legitimate
since the two clauses have different truth values today.

**Two viable dispositions, operator's call:**
(a) Amend the REQ to describe the live-table `ILIKE` search actually built, keeping the
    testable clause "Search is never consulted for authz decisions" and citing
    `reader_visibility_integration_test.go` / `reader_contract_parity_integration_test.go`
    as its evidence.
(b) Build the derived index (`tsvector`/GIN) the REQ describes, plus a tested reindex
    procedure.

(a) is honest documentation of a deliberate choice; (b) is a feature with real
implementation cost (index maintenance triggers, backfill job, staleness handling). Do
not pick one silently — this is a product decision, not a CI decision.

**Owner decision required:** choose (a) or (b). Not a CI-restructure blocker.

## REQ-SEC-3 — category error in the gate

**REQ text** (`wiki/architecture/backend-target-architecture.md:283`):
> - **REQ-SEC-3** OWASP ASVS is the review checklist for any change touching auth,
>   input handling, file paths, crypto, or queries. (MUST)

**Evidence — the gate's evidence model, not the codebase:** `scripts/req-trace` accepts
exactly two evidence kinds:
- `test` — the REQ literal appears in a `*_test.go` file under `internal/` or `apps/`.
- `commit` — an entry in `wiki/architecture/req-trace-map.yaml` (itself constrained to
  cite a commit, per that file's own header, reproduced above).

`grep -n 'REQ-SEC-3' wiki/architecture/backend-target-architecture.md` returns only the
REQ's own definition line (283); it does not appear in any `req-trace-map.yaml` entry
(confirmed: `grep -rn 'REQ-SEC-3' wiki/architecture/req-trace-map.yaml` → no matches) or
in any test file.

**Verdict — this REQ is not false, it is unmeasurable by this gate's design.**
"OWASP ASVS governs code review for sensitive changes" is a statement about the review
*process*, not about a runtime behavior or a piece of code. There is no `*_test.go`
assertion that can prove a checklist was consulted during a human's review, and citing a
`commit` that merely points at a checklist document (rather than an implemented/enforced
behavior, as every other `commit`-kind entry in the map does per its header's convention)
would be a category-mismatched use of that evidence kind — a false `commit` in the sense
the map's anti-gaming header is protecting against.

**Recommended disposition:** `req-trace`'s evidence model (`test` + `commit`) only fits
code-behavior REQs. Either (a) classify process REQs like REQ-SEC-3 as a distinct class
the gate does not demand code evidence for (e.g. an explicit "process" REQ kind excluded
from the uncovered-MUST count), or (b) reclassify REQ-SEC-3 as SHOULD, given the gate
cannot enforce MUST-level process requirements at all. Adding a `commit` entry pointing
at a checklist document would misuse the evidence kind.

**Owner decision required:** rule on gate scope for process REQs (affects the gate's
design, not just this REQ). Not a CI-restructure blocker.

## Net

`req-trace` stays red until these are ruled on. That is the correct state: the control
is not broken, the world it measures is. Silencing it with a map entry for any of the
four would convert a known, load-bearing divergence into an undetectable one — the exact
failure this disposition exists to avoid.
