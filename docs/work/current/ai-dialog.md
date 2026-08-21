# T8-E Fable independent review

> **Evidence only — non-authoritative.**
> Candidate authority remains `arch/t8e-wire-contract`; this review branch must never merge.

## Lead handoff

Repository: `developmentconexus-ops/MetalDocs`

Candidate branch: `arch/t8e-wire-contract`

Exact candidate HEAD under review: `ef329534fc9d5df3254d59c3787197fefa8435e6`

Review branch: `review/t8e-fable`

Gate: **T8-E — Executable Wire Contract**

Canonical Method: `developmentconexus-ops/conexus-methodology/METHOD.md` v1.0.0

Repository Standard: `developmentconexus-ops/conexus-methodology/REPOSITORY-STANDARD.md` v1.0.0

### Fresh-actor route

Reconstruct authority independently:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ only the smallest owning authority pack needed for a finding
```

Do not use this handoff as architecture authority. Repository current authority wins.

### Candidate target

Adversarially challenge whether the candidate is the **smallest sustainable executable wire** for the accepted 78-operation `/api/v1` Product/T6 census, with no Writer-visible semantic choices left that belong in T8-E.

Lead evidence already includes executable probes for:

```text
78-row census / unique operationId / method+path
10 durable Idempotency-Key creations
13 ETag read/mutation domains
4 exact-byte resources
OpenAPI 3.0.3 -> oapi-codegen v2.8.0
OpenAPI 3.0.3 -> openapi-typescript 7.13.0
kin-openapi strict-request split
S3 create-only presign + exact Content-Length reference profile
document admission ceilings / adversarial DOCX fixtures
whole-candidate Global Coherence PASS
```

Recent bounded corrections were operator-approved and are now in their owning authorities:

```text
T3      remove unreachable ProviderSubjectBinding-disabled Audit event
T8-D    persist Governance Step label + immutable attempt label_snapshot
T8-D    Audit/transaction + 24h idempotency precision
T4/T5/T8-C/T8-D
        already-PDF RequireOfficialRendition(PDF) reuses exact admitted bytes;
        no renderer/copy/River job unless transformation is required
T8-C/T8-D
        server-side per-session CSRF synchronizer secret is reconstructible for GET /session
```

### Review focus

Try to **falsify**, not confirm, the candidate. In particular attack:

1. **Authority leakage / duplication** — any wire rule that re-owns T1→T8-D semantics, or any upstream semantic requirement the wire fails to encode.
2. **Hidden Writer decisions** — required/optional/nullability, unions, cross-field presence, ordering, status/header/problem mapping, ETag/idempotency/replay behavior, upload/byte semantics, Audit projection.
3. **YAGNI / overengineering** — fields, headers, problems, normalization, limits, jobs, security mechanisms, projection data or generic abstractions with no current consumer.
4. **Security correctness** — session/CSRF, disclosure precedence, idempotency replay authorization/expiry, exact-byte integrity, direct upload admission, Problem information leakage.
5. **Persistence/internal-contract executability** — especially parity after the bounded corrections; no wire property should require state/contracts that accepted owners cannot realize.
6. **Generation/tooling feasibility** — identify any OpenAPI 3.0.3 shape the stated Go/TypeScript boundaries cannot represent without a material semantic compromise.
7. **Structural Inversion** — look for anything inherited from familiar API/platform patterns rather than current MetalDocs Product requirements.
8. **Global Maximum** — propose a materially smaller/stronger alternative only if it preserves all accepted properties; do not add platform capability by preference.

### Output contract

Write your independent review **only in this file** below `## Fable response`.

For each material finding, provide:

```text
ID
severity: MATERIAL | MINOR
claim
owning authority implicated
concrete counterexample/failure
smallest correction
whether it reopens an accepted authority
```

Separate non-material observations from findings. If no material finding survives, say so explicitly and list the strongest attacks attempted.

Do not edit the candidate branch or any other file. Reviewer output is Evidence; the Lead adjudicates every finding.

## Fable response

**Reviewer:** Fable, independent · **Candidate HEAD:** `ef329534fc9d5df3254d59c3787197fefa8435e6` · **Verdict:** **NOT CONVERGED** — 7 material findings survive falsification.

Authority was reconstructed independently (`AGENTS.md` → `docs/index.md` → `docs/roadmap.md`, then the smallest owning pack per finding). The Lead handoff was not used as authority. This output is Evidence only.

The candidate is strong. Every mechanical closure claim in §9.3 that I could re-execute independently held exactly (see *Attacks that failed*, below). The surviving findings are not about the 78-row ledger being wrong; they are about (a) the candidate having no durable owner after ratification, (b) three wire members/laws that no accepted owner can realize or that leave the generated boundary undetermined, and (c) two live contradictions with ratified T6 text that were never recorded as bounded precisions.

---

### Material findings

#### F-1 — The ratified artifact has no durable authority, and merging it violates documentation governance

```text
severity: MATERIAL
```

**claim** — T8-E "exits by operator ratification" (`docs/roadmap.md`, Remaining architecture program), but the entire executable contract lives in `docs/work/current/proposal.md`, which is `kind: work` and self-declares "Temporary / non-authoritative" (`proposal.md:10`). `docs/development/documentation.md:44` states: "`work` remains temporary and branch-only under `docs/work/`; it never enters a merge candidate or `main`." PR #136 is a merge candidate with `baseRefName: main`, and the candidate diff against `merge-base(main, arch/t8e-wire-contract)` adds exactly `docs/work/current/index.md` (+24) and `docs/work/current/proposal.md` (+1692). Neither §11 ("Remaining closure gate", A→F) nor the roadmap "Exact next action" contains a promotion/absorption/deletion step.

**owning authority implicated** — `docs/development/documentation.md` (documentation governance), `docs/roadmap.md` (stage exit), `docs/reference/t8e-checkpoint.md` (removal trigger), `docs/index.md` + `docs/decisions/index.md` (routing/register).

**concrete counterexample/failure** — Ratify and merge #136 as it stands. Result: `main` contains `docs/work/**` in direct violation of `documentation.md:44`. Now open T8-F ("Opens after T8-E ratification") and route a fresh actor through `AGENTS.md` → `docs/index.md`. The router has exactly one T8-E row — "Accepted T8-E design checkpoint → `reference/t8e-checkpoint.md`" (`docs/index.md:34`) — which preserves only the *previous four* layers and is explicitly "Non-authoritative until the final T8-E candidate is ratified". `docs/decisions/index.md` has rows T8-A, T8-B, T8-C, T8-D and **no T8-E row**. So the closed 78-row ledger, the component registry, the header/problem matrices and the admission ceilings are reachable only through a `kind: work` file that governance says must not exist on `main`. Worse, the checkpoint's own removal trigger ("when the owning stage is ratified and its durable authority fully absorbs the checkpoint") can never fire, because no durable authority exists to absorb it. The gate is circular.

**smallest correction** — Add one promotion step to `proposal.md` §11 and to the roadmap "Exact next action", between adjudication (E) and ratification (F):

```text
promote the wire contract to a durable kind: authority document
  (e.g. docs/architecture/wire-contract.md)
+ add its docs/index.md router row
+ add the T8-E row to docs/decisions/index.md
+ delete docs/reference/t8e-checkpoint.md per its own removal trigger
+ delete docs/work/current/** before the merge commit
```

No contract content changes. This is the same promotion pattern already used for T8-C/T8-D.

**reopens an accepted authority** — No. It executes existing governance and the checkpoint's own removal trigger.

---

#### F-2 — `SubmissionRepresentationGate.attention_required` is not realizable by any accepted owner and contradicts T5/T6

```text
severity: MATERIAL
```

**claim** — `attention_required:boolean` is a **required** member of `SubmissionRepresentationGate` (`proposal.md:739`), defined as true "only while rendition required + unsatisfied + terminal renderer attention exists" (`proposal.md:749`). No accepted owner holds that state, T6 forbids it as a business fact, and T8-D forbids the only query that could produce it.

**owning authority implicated** — `docs/product/journeys.md` §14 and §2.3; `docs/architecture/async-and-search.md` §6/§14/§16; `docs/architecture/persistence.md` §18 + DB-object ownership catalog; `docs/architecture/interfaces.md` §17; `docs/reference/t8e-checkpoint.md` layer 4 ("**may** surface", never "must").

**concrete counterexample/failure** — DOCX Submission under `RequireOfficialRendition(PDF)`. The River `official_rendition_render` job exhausts bounded retry and goes terminal. Client calls `getSubmission` (row 63). The response **must** serialize `representation_gate.attention_required`. Now try to answer it from accepted state:

1. `controlled_docs.official_renditions` is insert-only; there is no rendition-failure row and no `render_attention` fact anywhere in `persistence.md`.
2. `journeys.md:707` — "Renderer/job state is mechanism only. No `RENDERING`/`RENDER_FAILED` business lifecycle states exist." A required boolean on the Submission read model *is* a business-visible render-failure state under another name.
3. `async-and-search.md:226` — "A terminal renderer failure leaves the still-eligible Submission truthfully `SUBMITTED` with the Release gate unsatisfied." The accepted product answer is `required=true, satisfied=false` and nothing more.
4. The only state that answers the question is River job state. `persistence.md` classifies `river.*` as `THIRD_PARTY_MANAGED`, records "first-party raw SQL against `river.*` → FAIL" (line 137), "Raw first-party SQL against River objects is forbidden" (line 1168), and pins D34 "no first-party raw SQL against `river.*`". `interfaces.md` §17: "Renderer/provider job ids never become semantic identity."
5. `async-and-search.md` §16 assigns terminal-failure visibility and redrive to the **operations** surface, which `journeys.md` §2.3 says "never become[s] business authority".

So a Writer has exactly three options, all defective: read `river.*` (forbidden), hard-code `false` (a required wire field that is a constant lie), or invent a durable attention fact in `controlled_docs` (an unrequested T5/T8-D reopen). This is attack surfaces 1, 3 and 5 simultaneously.

**smallest correction** — Subtract the member:

```text
SubmissionRepresentationGate { required:boolean, satisfied:boolean }
```

`required=true, satisfied=false` already carries the complete accepted product meaning, and T5 §16 already owns terminal visibility + manual redrive. Also delete the §3.6 cross-field line at `proposal.md:749`.

**reopens an accepted authority** — No. The checkpoint permits (`may`) but never requires it, so removing it stays inside the accepted checkpoint. If the operator instead *wants* a governed attention signal, that is a deliberate T5 + T8-D reopen and must not be smuggled in as a wire field.

---

#### F-3 — The required/optional/nullable matrix is not closed: the presence-law encoding rule is a Writer fork, and `ObsolescenceRequestView` proves it is not uniformly decidable

```text
severity: MATERIAL
```

**claim** — `proposal.md:729` says presence laws "are encoded as closed OAS branches where OAS 3.0.3 can represent them; value-relational laws that would require distortion remain explicit contract-fixture assertions" — without saying, per schema, which side of that line it falls on. Item 2 of the checkpoint's own "Next design layer" is "required/nullable field matrix". It is not closed. Two conforming Writers produce incompatible generated Go and TypeScript.

**owning authority implicated** — `docs/reference/t8e-checkpoint.md` (Next design layer, items 1–2); T8-E itself (this is the gate's own definition of done: "no Writer-visible semantic choices left that belong in T8-E").

**concrete counterexample/failure** — Two parts.

*(a) The fork.* `RevisionView.current_submission_id` is "present **iff** `state=submitted`" (`proposal.md:729`). Encoding A: `RevisionView` is a `oneOf` discriminated on `state` with six branches, only the `submitted` branch requiring `current_submission_id`. Encoding B: one flat object with `current_submission_id` optional plus a named contract fixture. Both satisfy the prose exactly. Encoding A generates a Go union with `json.RawMessage` storage and six typed accessors and a TS union of six shapes; Encoding B generates `*uuid.UUID` / `current_submission_id?: string`. These are incompatible client boundaries produced from the same ratified sentence. The same fork exists for `DocumentSummary.status → official_revision`, `DocumentOfficialView.status → official`, `TemplateConfigurationItem.current_effective_title`, `DocumentHistoryItem.governance_decision.reason`, and the five `SubmissionView` cross-field laws. Note the candidate is already inconsistent with itself here: `SubmissionCreateResult` is a union while `SubmissionView` — carrying five presence laws — is flat with optional members.

*(b) The proof that a blanket "always use branches" rule is impossible.* `ObsolescenceRequestView` (`proposal.md:863-871`):

```text
returned / withdrawn / completed-human  → ... state + governance_attempt_id + ended_at
completed-no-human                      → ... state=completed + ended_at; governance_attempt_id absent
```

`ObsolescenceRequestState` has four values; **`completed` must map to two different schemas**. OAS 3.0.3 `discriminator.mapping` maps one property value to exactly one schema, so `state` is not a total discriminator and there is no second discriminator property available. Both branches are live: `journeys.md` §14 "NoHumanApproval obsolescence → zero human Step → no fake System approver" reaches `completed` with no attempt, while human ACCEPT reaches `completed` with one. The candidate's own §2.2 law ("True closed semantic unions use `oneOf` + required discriminator") therefore cannot be satisfied here, and the candidate never says so.

**smallest correction** — Add one closed rule to §3 plus a per-schema marker, e.g.:

```text
a presence law is encoded as an OAS oneOf branch set
  iff exactly one required closed-enum member is a TOTAL discriminator for it;
otherwise the member is schema-optional and the law is a NAMED contract fixture
```

Then mark each affected schema in §3 (`branch` / `fixture`), and name `ObsolescenceRequestView` explicitly as `fixture` (`governance_attempt_id?` optional, with `present iff a GovernanceAttempt exists` as a named fixture). This is a precision, not new capability.

**reopens an accepted authority** — No. It closes an item the checkpoint already assigned to T8-E.

---

#### F-4 — The durable Problem `type` namespace contradicts ratified T6, with no recorded bounded precision

```text
severity: MATERIAL
```

**claim** — `proposal.md:1024` freezes `type=https://errors.conexus.fun/metaldocs/{code}`. `docs/product/journeys.md:1172` — `CURRENT / RATIFIED` T6, per `docs/decisions/index.md` — states `type = https://errors.metaldocs.io/{code}` under "One machine authority". Both the host and the path shape differ. `type` is one of the seven required Problem members and is durable public contract identity. The only other document carrying the Conexus form is `docs/reference/t8e-checkpoint.md:83`, which is `kind: checkpoint` and self-declares "Non-authoritative until the final T8-E candidate is ratified" — a non-authoritative checkpoint cannot override a ratified T6 authority.

**owning authority implicated** — `docs/product/journeys.md` §25 (T6), with `docs/decisions/api-operation-census.md` as the precedent mechanism.

**concrete counterexample/failure** — A fresh actor routed by `docs/index.md` for "Product journeys / application API meaning" reads `journeys.md` §25 and emits `https://errors.metaldocs.io/permission.denied`. A Writer routed to the candidate emits `https://errors.conexus.fun/metaldocs/permission.denied`. Both followed current repository authority. Nothing in `docs/decisions/**` disambiguates: the register has no row for the problem namespace, and `forward-obligations.md` carries no obligation covering it. Note the repository already has the exact mechanism for this class — the two read-symmetry GETs got their own `docs/decisions/api-operation-census.md` page **and** a `T6-API` register row precisely so a bounded T6 override would be discoverable. That mechanism was not used here.

A weaker second instance of the same class: `journeys.md` §12 prints `ETag: "draft-<generation>"`, while `proposal.md` §2.4 requires ETags that "never expose raw DB version/generation". I treat this instance as illustrative rather than normative, but it is the same unrecorded-override pattern and should be closed in the same edit.

**smallest correction** — Record a bounded T6 precision (mirroring `api-operation-census.md`): a short `docs/decisions/problem-type-namespace.md` (or a `## Problem type namespace` section in the promoted authority from F-1) stating that the Conexus organizational namespace convention supersedes the `errors.metaldocs.io/{code}` form in `journeys.md` §25 **for the type URI only**, plus a register row. Optionally replace the §12 ETag literal with `ETag: "<opaque strong validator>"`.

**reopens an accepted authority** — Yes, bounded: `journeys.md` §25 type-URI sentence only (and optionally the §12 ETag literal). No Product operation, journey, error family, code, status or title changes. Operator-approvable at the same scale as the 76→78 census precision.

---

#### F-5 — Unbounded human text multiplied by 100-item pages makes several ratified list responses unbounded

```text
severity: MATERIAL
```

**claim** — §2.3 rules that "Human text is bounded by the aggregate JSON ceiling rather than unrelated guessed per-field maxima." 41 members in §3 are typed `nonblank string` with no `maxLength`. The only bounded strings in the whole registry are machine-generated or opaque (`SearchQuery ≤256`, `OpaqueCursor ≤2048`, `CsrfToken ≤512`, `ProviderSubjectRef ≤2048`, `display_hints ≤256`, `ReplaySnapshot ≤2048`). The aggregate ceiling bounds one **request** at 65,536 bytes; it does not bound a **response** that projects up to 100 such values.

**owning authority implicated** — T8-E itself (§2.2 request ceiling, §2.3 scalar normalization, §2.7 pagination). No upstream authority forbids per-field maxima; `journeys.md` §7 explicitly delegates the analogous decision ("bounded length; exact API maximum belongs implementation-contract design").

**concrete counterexample/failure** — An ordinary author with `document.edit` sets a ~65,400-byte `title` via `updateRevisionDraft` (row 58) — legal, since the only bound is the request ceiling and `title` is `nonblank string`. Repeat over 100 revisions. Then:

```text
GET /api/v1/work/authoring?limit=100                     WorkAuthoringItem.title  -> ~6.5 MB
GET /api/v1/documents/{id}/history?limit=100             message/reason/title     -> ~6.5 MB
GET /api/v1/governance-attempts/{id}/feedback?limit=100  message                  -> ~6.5 MB
GET /api/v1/governance-attempts/{id}                     20 embedded feedback     -> ~1.3 MB
```

`getDocumentHistory`, `listGovernanceFeedback`, `listAuthoringWork` and `getGovernanceAttempt` therefore have no response bound at all. Secondary damage: `journeys.md` §6 makes the current EFFECTIVE title a canonical searchable fact with `title prefix` / `title contains` ranking — a 65 KB title degrades the ratified PostgreSQL Search baseline, and the governance inbox and Library become unusable. The trigger is an authorized insider, not an attacker, and nothing in the contract prevents it. The candidate's stated principle is also applied inconsistently: it guessed `≤256` for `SearchQuery` and for `display_hints` without measured evidence, but declined to bound the values that actually fan out.

**smallest correction** — Bound exactly the human-text scalars that appear inside a `PAGED` item array or an embedded page. One shared pair is enough:

```text
ShortName   nonblank, maxLength  256   // display_name, company/area/type/group name, GovernanceRouteStep.label
LongText    nonblank, maxLength 4096   // title, reason, message
```

This is additive by exactly two scalar definitions and removes an unbounded response class. It does not touch operations, owners, lifecycle or the census.

**reopens an accepted authority** — No. `journeys.md` §7 already assigns exact API maxima to the implementation-contract design, i.e. to T8-E.

---

#### F-6 — The Audit wire drops the one T3 §16 fact class it names, leaving either an under-projected audit read or a dormant T3 requirement

```text
severity: MATERIAL
```

**claim** — T3 §16 "Minimum bounded Audit facts" names six classes. The candidate projects five of them verbatim (`RoleAssignmentAuditFacts`, `GroupMembershipAuditFacts`, `GovernanceDecisionAuditFacts`, `ReleaseAuditFacts`, `RevisionCancellationAuditFacts`, `ObsolescenceAuditFacts`) and drops the sixth — "DocumentType/configuration changes ... bounded changed configuration identifiers/codes/operation facts" — with `exposed wire facts: none` (`proposal.md:967`), justified at `proposal.md:954` as avoiding "a generic configuration-diff bag without a named UI consumer".

**owning authority implicated** — `docs/architecture/authorization-and-audit.md` §13 and §16 (T3).

**concrete counterexample/failure** — `GET /api/v1/audit/events` (row 78) is the **only** consumer of Audit in the entire 78-operation census. A `governance_viewer` investigating why a document released without human approval retrieves:

```json
{"operation_code":"document_governance.changed","resource_kind":"document_type","resource_id":"…"}
```

with no `facts`. They cannot tell whether the route changed from `use_governance_route` to `no_human_approval`, or which steps or selectors moved. The same holds for `template_eligibility.changed` (which templates?) and `document_type.reconfigured` (which field?). T3 §13 requires Audit to reference "the minimum bounded facts required to reconstruct the action"; T3 §16 names those facts for this class. So the candidate produces a dilemma it does not resolve: either the T3-required facts have a consumer (this read) and the wire under-projects, or they have none — in which case T3 §16's DocumentType/configuration clause is dormant persistence with no reader, which is exactly the class §8.2 correctly removed for `provider_binding.disabled`.

The candidate is right to reject a generic diff bag. It is not entitled to take neither position.

**smallest correction** — Take the subtractive branch, matching the §8.2 precedent: record a bounded T3 §16 precision removing the DocumentType/configuration facts requirement, on the evidence that no Launch consumer reads them, and state in `proposal.md` §4 that the closed operation code plus `resource_kind`/`resource_id` is the complete Launch evidence for that class. If the operator instead wants auditor-visible configuration evidence, the smallest additive form is one closed enum per code — e.g. `DocumentTypeReconfigurationAuditFacts { changed: unique enum[code,name,numbering_scope,active][] }` and `DocumentGovernanceAuditFacts { mode: GovernanceMode, step_count: integer }` — never a JSON diff.

**reopens an accepted authority** — Yes, bounded: `authorization-and-audit.md` §16, DocumentType/configuration clause only. Same scale and same evidence class as the already-approved §8.2 correction. No operation code, resource kind, permission or census entry changes.

---

#### F-7 — The `PROFILE_REPLACE` conditional matrix has a reachable unspecified cell, and the race that reaches it is the privacy path

```text
severity: MATERIAL
```

**claim** — `proposal.md:212` enumerates the profile conditional matrix:

```text
existing profile           If-Match required
absent profile recreation  If-None-Match:* required
both / neither             400
If-None-Match:* + existing profile -> 412
```

The cell **`If-Match` supplied + profile absent** is missing. Row 10 declares both `N` (404 `notfound.resource`) and `P` (412 `precondition.resource_changed`), so both are legal answers and the Writer must choose.

**owning authority implicated** — T8-E §2.4/§2.8 and ledger row 10; interacts with `authorization-and-audit.md` §15 (lawful `user_profile.erased`).

**concrete counterexample/failure** — The cell is reachable precisely because `deleteUserProfile` (row 11) is `no body / UNSAFE_CSRF` — no `If-Match`, no `P` in its problem set, i.e. deliberately unconditional. Sequence: actor A opens the profile editor and holds ETag `E`; a lawful erasure or another admin calls `DELETE /users/{id}/profile`; A saves with `If-Match: E`. Writer 1 returns `404` (target representation gone) and the SPA shows "user not found", losing the user's edits and hiding the recreate path. Writer 2 returns `412` and the SPA re-fetches, discovers absence, and retries with `If-None-Match: *` — which is the behavior the matrix was designed for. RFC 9110 §13.1.1 pins the answer (`If-Match` with an entity-tag against a target with no current representation evaluates to false → `412`), so this is pure precision, but the candidate's own bar is that no Writer-visible choice survives.

**smallest correction** — One row in the §2.4 matrix:

```text
If-Match + absent profile  -> 412 precondition.resource_changed
```

and one fixture line in §9.4 under "PROFILE_REPLACE If-Match/If-None-Match matrix" (which currently names the matrix without enumerating its cells).

**reopens an accepted authority** — No.

---

### Minor findings

```text
O-1  §2.9 mandates mechanism, not a wire property
     "load semantic descriptor -> OpenExact -> verify count + SHA-256 -> ONLY THEN commit HTTP 200"
     plus "may not stream unverified provider bytes ... after the 200 has begun" (proposal.md:404-429)
     forces a full spool (up to DOC_RAW_MAX_BYTES = 100 MiB) or a double read on every exact-byte GET.
     The wire-observable property is identical if the server streams while hashing and TERMINATES the
     response before Content-Length is satisfied on mismatch: a conforming client MUST treat a short
     response with a declared Content-Length as incomplete (RFC 9112), and Cache-Control private,no-store
     excludes intermediary retention. Runtime buffering strategy is T8-G, not T8-E.
     correction: state the property ("no COMPLETE 200 ever delivers bytes whose SHA-256 differs from the
     semantic descriptor; a detected mismatch MUST terminate the response before completion"), drop the
     spool/no-stream mandate. Subtractive; reopens nothing.

O-2  429 Retry-After is Writer-optional with no named consumer (proposal.md:402).
     "may add ... but never fabricates one" leaves presence to the Writer, which is the exact class T8-E
     closes. journeys.md §25 requires the frontend to branch on `code`, never on headers, and no journey
     names a backoff consumer. correction: subtract it, or make it REQUIRED whenever the limiter exposes a
     window and state that source.

O-3  Deterministic list order is inconsistently derived. listAreas orders by `code ASC, area_id ASC`
     (business key), but listDocumentTypes orders by `document_type_id ASC` even though DocumentTypeView
     carries the same always-present `code:CodeToken`; listGroups orders by `group_id ASC` though GroupView
     carries a required `name`; listGovernanceWork orders by raw `governance_attempt_id` though
     WorkGovernanceItem carries `created_at` — a human governance inbox in random UUID order.
     listUsers/listRoleAssignments legitimately have no total business key (display_name is optional and
     erasable), so UUID order is correct there. correction: order by the always-present business key with
     UUID tie-break where one exists; keep UUID-only where none does.

O-4  §8.6's rejection rationale is inconsistent with ratified T8-D. It rejects a "global CSRF-HMAC rotation
     subsystem", but persistence.md already ratifies exactly that subsystem one table family away:
     `platform.idempotency_keys.fingerprint_key_version INTEGER NOT NULL CHECK(>0)` plus the pinned control
     "idempotency HMAC rotation drain prevents honest-retry false conflict". §2.7 also mandates
     "integrity-protected" stateless cursors, which already requires a server-side MAC key. A keyed
     derivation `csrf_token = MAC(server_key, session_id)` is reconstructible on GET /session with ZERO new
     durable state and is OWASP's documented HMAC-based alternative to the synchronizer token.
     I do NOT recommend reopening §8.6: the approved stateful secret is correct and the delta is one column.
     But the stated reason for rejecting the alternative is factually wrong and should not survive into the
     durable authority as precedent.

O-5  `validation.failed` (422) has no defined trigger on rows 3 (`searchProviderSubjects`) and 45
     (`listDocuments`). §2.2 maps "malformed path/query/header/JSON" to 400 and §2.6 stage 5 covers "pure
     normalization", so the 400-vs-422 boundary for a syntactically valid but semantically rejected QUERY
     value is unstated. Row 42 (`numbering-preview`) has an obvious trigger (`area_id` supplied against
     `numbering_scope=document_type`); rows 3 and 45 do not. correction: name the trigger per row, or drop
     `validation.failed` from rows 3/45 and route those failures to `request.invalid`.

O-6  Registry gaps: `Rfc6901Pointer` (used at proposal.md §5) and `URI` (used for
     `DraftUploadAllocation.upload_url`) are referenced but absent from the §2.3 scalar registry.
     `required_headers:map<string,string>` — the single deliberate map — has no `maxProperties`, no key
     pattern and no value maxLength, while every other opaque scalar in §2.3 is bounded.

O-7  `AuditEventView` union cardinality is unspecified. "closed operation_code-discriminated union" over 37
     codes admits 37 branch schemas or 7 (one per exposed-facts shape) with a 37-entry `discriminator.mapping`
     — OAS 3.0.3 permits many mapping keys to point at one `$ref`. Both satisfy the prose; they generate very
     different Go/TS. This is F-3's class, called out separately because the audit union is the largest.
     correction: specify grouping by facts shape with a full 37-entry mapping.

O-8  §2.7's cursor law says the cursor authenticates `operationId + normalized filters + ordering + seek
     position`, but §3.7 states that `GovernanceCaseView.feedback`'s cursor "targets listGovernanceFeedback"
     — i.e. an operation OTHER than the one that produced it. The carve-out is stated only in §3.7.
     correction: state it in §2.7 as the general law (an embedded page's cursor authenticates the standalone
     list operation that continues it).

O-9  `ObsolescenceCreationState = governance_pending | obsolete` and
     `ObsolescenceRequestState = active | returned | withdrawn | completed` describe the same no-human
     outcome with two different tokens (`obsolete` in the create result, `completed` in the read view). A
     client mapping create-result to read-view must know `obsolete == completed`. That equivalence is
     nowhere stated. correction: state it, or align the tokens.

O-10 forward-obligations.md consumption law: "A later ratified authority wins over an older
     forward-obligation wording and must update this page when it closes or materially refines that
     obligation." The candidate materially refines DOC-12 (REOPEN — exact numbering grammar) by freezing
     `DocumentCode ^[A-Z0-9]+(?:-[A-Z0-9]+)?-[0-9]{3,}$` and `CodeInput` at `^[A-Z0-9]+$`, length 1..32 —
     the exact API maximum that journeys.md §7 explicitly delegated to the implementation contract. §11's
     closure gate does not include the forward-obligations reconciliation, so ratification would leave the
     register stale (and its 21/4/27 = 52 count proof wrong). correction: fold the DOC-12 disposition and
     recount into the F-1 promotion step.
```

---

### Non-material observations

- **Pre-scan bytes served `inline`.** `getRevisionDraftSource` (row 61) and `getSubmissionSource` (row 64) serve `Content-Disposition: inline` bytes that have not passed malware inspection — scanning is at the governed admission boundary (SUBMIT), per `journeys.md` §13 and the checkpoint. This is ratified Product risk, correctly implemented, and the inline disposition is required by the §20 viewer journeys. Recording it as a known accepted exposure rather than a finding.
- **`AuditActor {kind:system}` omits `system_actor_code`.** T3 §13 requires the code on the record; the candidate keeps it internal "until a named client needs to distinguish multiple system principals". Launch already has at least two system paths (`release.completed`, `official_rendition.completed`), so the single-principal premise is unproven — but T3 §16 does not name it as required wire facts, so this is weaker than F-6 and I do not raise it as a finding.
- **Two byte-identical endpoints after §8.5.** For an already-PDF `RequireOfficialRendition(PDF)`, `getReleaseSource` and `getOfficialRenditionContent` return identical bytes with identical `Content-Digest` and identical generated filenames, and `ReleaseView` carries two identical `ContentSummary` objects. This is semantically honest (the OfficialRendition fact genuinely exists) and clients can detect sameness via `sha256`. No correction proposed.
- **T8-E narrows a T6 implementation freedom.** `journeys.md` §20 permits "an authorization-checked short-lived provider/CDN redirect"; §2.9 forbids redirects at Launch. This is legitimate — the redirect is wire-observable, so T8-E owns it, and the checkpoint already ratified it — but it is a subtraction from a granted freedom and should be visible in the promoted authority.

---

### Strongest attacks that failed to falsify the candidate

I re-executed the mechanical claims rather than trusting §9.3. All held.

```text
ledger <-> census, mechanical diff of method+path against journeys.md §29
  78 ledger rows; row numbers exactly 1..78; 78 unique operationIds; 78 unique method+path
  set difference against the §29 census in BOTH directions: EMPTY
  family partition 3 / 26 / 4 / 10 / 34 / 1 = 78                                      PASS

10 Idempotency-Key rows (7,17,23,32,35,46,52,62,69,75) == journeys.md §18 list exactly PASS
13 JSON_ETAG concurrency domains == journeys.md §24 minimum lost-update list           PASS
4 exact-byte resources (61,64,73,74) == journeys.md §20 list exactly                   PASS
role bundles + 15-value PermissionCode + scope matrix == T3 §4/§5/§6 verbatim          PASS
37 AuditOperationCode values == T3 §15 census exactly, incl. the §8.2 removal          PASS
5 of 6 T3 §16 fact classes projected verbatim                                         PASS (6th -> F-6)
```

Attacks I pressed and had to abandon:

- **`NumberingPreviewView.reservation:false` as a constant-value YAGNI field.** Refuted: `journeys.md:365` mandates `reservation=false` in the preview response. T6 owns it.
- **`startRevisionDraftUpload` missing an `Idempotency-Key`** while creating durable `ManagedContent OPEN` + `AdmissionClaim`. Refuted: `journeys.md` §13's truth ladder makes provider upload success explicitly not a semantic fact, journeys §18 scopes durable keys to "another semantic fact/resource", and T4 admission-claim GC reclaims orphans at the shared 15-minute expiry.
- **§8.5 same-PDF rendition as a silent `SourceOnly` downgrade** (`async-and-search.md:217` forbids exactly that). Refuted: the OfficialRendition fact is still created, `official_rendition.completed` Audit still fires, and I verified the reconciliation actually landed in all four owners in the candidate diff — T4 `content-integrity.md` two-path law, T5 §6 already-PDF path, T8-C §17.2 "returns **no** rendition intent", T8-D SUBMIT sequence and the `official_renditions` same-`managed_content_id` clause.
- **§2.10 `PresignCreate(handle, maxBytes=expected_size_bytes, ttl)` as a hidden T8-C signature reopen.** Refuted: the portable property (`<=`) is genuinely weaker than the S3 profile (exact signed `Content-Length`), completion independently derives the descriptor, and the client bound is never promoted to content identity. The subtractive result is correct.
- **`DocumentSummary` / `DocumentOfficialView` presence laws as a second lifecycle authority.** Refuted: they match `journeys.md` §6's derived-status law and its `REV000 EFFECTIVE + REV001 CANCELLED -> effective` example exactly, and `official present iff >=1 Release` is consistent across both lenses.
- **Idempotency key scope excluding path identifiers.** Refuted: path identifiers are in the *fingerprint*, so a reused UUID across two attempts/documents produces `422 validation.idempotency_key_reused` rather than a false replay. Fail-closed is correct.
- **`createSubmission` replay bypassing the DRAFT precondition.** Refuted: the `If-Match` value is fingerprint input, so a differently-based retry cannot replay; and §2.6 stage 7 sits strictly after stage 6 AuthZ, matching `journeys.md` §19 verbatim.
- **CSRF secret stored in plaintext (§8.6).** Refuted: it is not a bearer credential, it is worthless without the `__Host-` HttpOnly session cookie, and it is OWASP's documented stateful synchronizer pattern. (The *rationale* defect is O-4, not the decision.)
- **Disclosure precedence leaking existence via 403-vs-404.** Refuted: §5's three-line rule plus the per-row problem sets are consistent, including the `governance.act`-without-blanket-read case (`getSubmission` outside an active step must be 404, not 403) and the `listDocuments status=obsolete|cancelled` 403.
- **Ledger row 47 `getDocument` carrying `B + N` while sibling reads carry `A + N`.** Refuted: correct — the document lens is disclosure-filtered (404), while the sub-resources gate on a permission over an already-visible document (403).

---

### Global Maximum assessment

I could not find a materially smaller *architecture* than "one component registry + one closed 78-row ledger". Structural Inversion holds: the operation set derives from Product/T6, not from REST habit; there is no envelope, no `/actions`, no filter DSL, no facts bag, no policy engine, no operation 79. §10's subtraction list is honest and I verified its non-introduction claims against the registry.

The available Global Maximum is **subtractive, and every material finding above except F-3, F-4 and F-7 moves in that direction**:

```text
- attention_required                       (F-2)  one required member, unrealizable
- T3 §16 DocumentType/configuration facts   (F-6)  dormant persistence with no reader
- verify-before-commit spool mandate        (O-1)  mechanism belonging to T8-G
- 429 Retry-After                           (O-2)  optional header with no consumer
```

The three non-subtractive corrections are all precisions, not capability: one presence-law encoding rule (F-3), one bounded T6 namespace precision (F-4), and one matrix cell (F-7). F-5 adds exactly two scalar definitions and removes an unbounded response class. F-1 adds no contract content at all.

Net: the candidate should shrink by one wire member, one persistence obligation, one mechanism mandate and one header, and gain one encoding rule, two bounded scalars, one matrix cell, one recorded T6 precision and one promotion step. I found no addition that would make it stronger.

**Recommendation to the Lead:** do not treat this gate as converged. F-1 alone would leave T8-F with no durable authority to consume, and F-2 would ship a required wire field that no accepted owner can compute.
