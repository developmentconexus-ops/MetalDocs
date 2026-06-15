# Milestone 4 — Systemic Ports (H-G class) (Wave D)

> **Program:** grade-a-architecture-remediation  ·  **Governing spec:** `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md`
> **Status:** Spec (drafting) — *authored before any feature in this milestone began.* Awaiting operator agreement (Phase 2 / HS-1-open).
> **Authored:** 2026-06-15 — *before any feature in this milestone began.*

> **⚠ HS-6 scope reconciliation (2026-06-15).** Before authoring this spec, the named H-G sites
> were verified against the current tree. The governing-spec §5.2 line "`iam_users` direct-read
> systemic 5+ sites (`documents/repository.go:134`, `approval/get_instance_handler.go:127`, +
> security + presence)" is **looser than the F4.1 port it prescribes**. Classification of every
> cross-module `iam_users` reader in the current tree:
>
> | Site | Reads | Shape | In `UserDisplayNameReader` scope? |
> |------|-------|-------|-----------------------------------|
> | `documents/approval/application/decision_service.go:163` (via `s.repo.LoadActorDisplayName`, impl `postgres_approval_repository.go:446`) | `iam_users.display_name` | cross-module display-name read; **F1.3's contained fix** (lives on `ApprovalRepository`, marked "M4/F4.1") | **Yes** — the contained fix to generalize |
> | `documents/repository/repository.go:134` | `iam_users.display_name` (created_by) | cross-module display-name read | **Yes** |
> | `documents/approval/http/get_instance_handler.go:127` | `iam_users.display_name` batch (`COALESCE(display_name,user_id)`, `ANY($2)`) | cross-module display-name read, raw `h.db` in the HTTP handler | **Yes** |
> | `iam/presence/*` (`repository.go`, `model.go`, `hub.go`, `middleware.go`) | `iam_users` | **iam-owned** — same module reading its own tables; **not** a cross-module reach-without-a-port | **No** — out of class (intra-module) |
> | `security/infrastructure/postgres/repository.go` MfaCoverage (×2), CountRecentLockouts | `iam_users` aggregate / COUNT, **no display-name read** | cross-module **tenant-scope/aggregate JOIN**, not a display-name reach — `UserDisplayNameReader` does not cover it | **No** — non-display-name tenant-scope read (genuine defer) |
> | `security/infrastructure/postgres/repository.go` **ListLockouts:89, CountRecentFailedLoginsByUser:137, ListNewDeviceLogins:191** | `iam_users.display_name` (`COALESCE(NULLIF(display_name,''),user_id)`) | cross-module **display-name read** — same H-G shape as the 3 F4.1 sites | **YES** — *miscounted in the original census (corrected 2026-06-15)* |
> | `documents/approval/infrastructure/signature/password_reauth.go` | `iam_users.password_hash` | already behind an `IamUserReader` interface; reauth, not display-name | **No** — already a port, different column |
>
> **⚠⚠ CENSUS CORRECTION (2026-06-15, post-validator-FAIL).** The original row above asserted
> security's `iam_users` reads were *"tenant-scope JOINs, no display-name read"*. **That was wrong.**
> Verified against the tree (`grep -n "NULLIF(u.display_name"`): **3** of security's reads
> (`ListLockouts:89`, `CountRecentFailedLoginsByUser:137`, `ListNewDeviceLogins:191`) DO read
> `iam_users.display_name` — identical H-G shape to the F4.1 sites. Plus the validator caught a 4th
> missed site, `auth/infrastructure/postgres/sessions_admin.go:32`. **Authoritative count of
> cross-module `iam_users.display_name` reads outside `iam/` after F4.1: 4** (1 auth + 3 security).
> The class is therefore **not** closed by F4.1 alone, and the original "defer all security" decision
> rested on a false premise.
>
> **Operator scope decision 2026-06-15 (post-correction): OPTION 2 — FULL CLOSE.** Close all 4
> display-name reaches in M4, including building the previously-deferred **iam tenant-scope/membership
> port** required to separate the 2 `auth_identities`-coupled security sites (ListLockouts,
> CountRecentFailedLoginsByUser — `auth_identities` has no `tenant_id`; the `iam_users` JOIN is the
> only tenant scoping). Executed as fix-features **F4.4 (auth) + F4.5 (iam membership port) + F4.6
> (security)**, added to the Features table below. **Still genuinely out of class (deferred, accurate):**
> security's `MfaCoverage`/`CountRecentLockouts` aggregate JOINs (no display-name; they count over
> iam_users) and security's direct reads of `auth_identities` (auth-owned, a *separate* cross-module
> concern not in the H-G display-name class) — *Trigger:* M5 re-audit flag or next structural touch;
> *Owner:* backend. F4.2 and the hardcode site verified exactly as the spec names.

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates
> the milestone against *this* document.

## Objective

Close the **H-G class** — *cross-module reach-without-a-port* and *hardcoded-domain-state* — by
generalizing to **shared, owning-module ports**. This is the last remediation milestone before the M5
re-audit, sequenced last so it cannot regress the grade. Two ports, one PR each; reads stay **live**
(no migrations, no snapshot/denormalization semantics — Approach 2 was explicitly rejected, design
D4/Approach-3).

**Bar moved & criterion:** the **module-boundaries / DDD** dimension reaches the §6 pass bar (≥ A−) by
eliminating the H-G class at the *class* level, not the instance level:
- **0 reach-without-a-port** — no module issues SQL against another module's owned tables; every such
  read goes through the owning module's domain port.
- **0 hardcoded-domain-state** — no wiring/adapter fabricates a domain value (`status := "published"`)
  that should be read from the owning module.

Root cause closed = the *class* is gone in the swept surface, verified by grep + build, not the single
instance patched.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F4.1 | `f4.1-user-display-name-reader` | Introduce an **iam/domain-owned** `UserDisplayNameReader` port (single-and-batch display-name lookup by `(tenantID, userID)`), implemented in iam infrastructure against `metaldocs.iam_users`. Migrate the **3 cross-module display-name read sites** to consume it: `documents/repository/repository.go:134`, `documents/approval/http/get_instance_handler.go:127`, and **generalize F1.3's contained `ApprovalRepository.LoadActorDisplayName`** (`decision_service.go:163`) onto the shared port. Reads stay live; existing off-tx placement (H-PRE-1) preserved for the signoff path. | The 3 named sites no longer issue raw `SELECT … FROM metaldocs.iam_users` for display names — each calls the iam-owned port (grep: **0** `iam_users` display-name SQL outside iam/ in those files). Port lives in iam `domain` (owned), impl in iam `infrastructure`. Empty-string-on-missing + `COALESCE(display_name,user_id)` semantics preserved per current behavior. Signoff display-name read stays **off** the lock-holding tx (H-PRE-1 intact — runtime/`pg_locks` evidence). `go build`+`go vet` clean; existing display-name tests (incl. `postgres_approval_repository_displayname_integration_test.go`, `decision_service_test.go` snapshot assertions) green or migrated; new port unit test. backend-api-qa-checklist green. **Out-of-port (documented, not migrated):** `iam/presence/*` (intra-module), `security` tenant-scope JOINs (scope decision), `password_reauth` (existing `IamUserReader`). |
| F4.2 | `f4.2-template-version-state-reader` | **[HS-6 amended 2026-06-15 — see note below]** **Extend** the **existing** templates/domain-owned port (`TemplateVersionPort`, impl `TemplateVersionReader`, Wave Z Z-7) with a raw-state read `GetTemplateVersionState(ctx, tenantID, versionID) (*string, string, error)` = `(status, doc_type_code)` (keep `IsPublished` for taxonomy). Reuse the single existing `templateVersionQuery`. Replace (a) CD's `PostgresTemplateVersionChecker` reach into `templates_template_version`+`templates_template` (`controlleddocuments/infrastructure/repository.go:702-711`) — **delete** it; CD's module wires the templates-owned reader as `tplCheck` (directly satisfies CD's `application.TemplateVersionChecker`), and (b) the `documents_adapters.go:113` hardcoded `status := "published"` so the adapter reads the **real** template version status via the port. Read stays **off** the lock-holding CD-create tx (H-PRE-1). | CD no longer issues SQL against `templates_*` tables (grep: **0** `templates_template` references under `controlleddocuments/`); the read flows through the templates-owned port. `documents_adapters.go` no longer hardcodes `"published"` — status is read from the port (grep: **0** `status := "published"` in `wiring/`). CD's `TemplateVersionChecker` consumer contract (`(*string,string,error)` = status, profileCode/doc_type_code) preserved so the override-validation call sites (`service.go:209,308`) are behavior-identical. Status read stays off the CD-create lock-holding tx (H-PRE-1 — no authz-recording read inside the lock; runtime/`pg_locks` evidence). `go build`+`go vet` clean; existing CD override + template-checker tests + taxonomy `IsPublished` tests green or migrated; new templates port `GetTemplateVersionState` unit test. backend-api-qa-checklist + workflow-async-qa-checklist (CD-create is lock-bearing) green. |
| F4.3 | `f4.3-port-adrs` | One **ADR per port** into the now-clean `wiki/decisions/` ledger: the `UserDisplayNameReader` boundary decision and the `TemplateVersionStateReader` boundary decision (each: context = H-G reach, decision = owning-module port, consequences = reads-live/no-snapshot, alternatives rejected incl. Approach 2). | Two ADRs exist under `wiki/decisions/` with canonical `Status:` headers, registered in the decisions `index.md`, cross-linked from F4.1/F4.2 `spec.md`, and referenced by the touched module wiki docs. Each ADR records the design D4/Approach-3 constraint (reads live, no migration). |
| F4.4 | `f4.4-auth-session-display-name-port` | **[HS-4 fix — validator-found missed site]** Migrate `auth/infrastructure/postgres/sessions_admin.go` `ListActiveSessions` off its `JOIN metaldocs.iam_users` display-name read. `auth_sessions.tenant_id` already scopes; drop the JOIN (auth no longer touches `iam_users`), return auth-owned session rows, and have the **iam** consumer (`iam/delivery/http/sessions_handler.go`) enrich display names via the existing `UserDisplayNameReader.DisplayNames` + `missing→user_id` fallback. Remove `DisplayName` from auth's `SessionListItem` (auth doesn't own it). | `auth/` issues **0** `iam_users` SQL (grep). Rendered `display_name` byte-identical (`COALESCE(NULLIF(display_name,''),user_id)` reproduced consumer-side via port + fallback) — handler unit test (port fake + fallback) + live-PG proof that `ListActiveSessions` returns tenant-scoped sessions without reaching `iam_users`. Bounded note: prior INNER JOIN dropped sessions whose user lacks an `iam_users` membership row; on the real path a session in tenant T implies login→membership, so no behavior delta (documented, not silently changed). `go build`+`go vet` clean; existing sessions tests green/migrated. |
| F4.5 | `f4.5-iam-tenant-membership-port` | **[Option-2 architecture piece — the previously-deferred port]** Introduce an **iam/domain-owned** tenant-scope/membership port (e.g. `TenantUserReader.TenantUserIDs(ctx, tenantID) ([]string, error)` — the set of `user_id`s with an `iam_users` row in the tenant; no `deactivated_at` filter, matching the current INNER-JOIN membership semantics), impl in iam/infrastructure on the pool (off-tx, H-PRE-1). This is the port the original census deferred; it exists to let `auth_identities`-coupled consumers tenant-scope without JOINing `iam_users`. | Port in iam `domain` (owned), impl in iam `infrastructure`, pool-backed. Returns exactly the tenant's member `user_id`s (membership semantics match the replaced JOIN — live-PG test: present members returned, other-tenant excluded, empty tenant → empty). `go build`+`go vet` clean; new port unit + integration test. ADR authored (see F4.3-style decision record; folded into the F4.3 ledger or its own ADR). |
| F4.6 | `f4.6-security-display-name-port` | **[Option-2 — close the 3 security display-name reaches]** Migrate `security/infrastructure/postgres/repository.go` off all 3 `iam_users.display_name` JOINs: `ListNewDeviceLogins:191` (separable — `auth_sessions.tenant_id` scopes; drop JOIN + enrich via `UserDisplayNameReader`), `ListLockouts:89` and `CountRecentFailedLoginsByUser:137` (coupled — replace the `iam_users` JOIN tenant-scope with the F4.5 `TenantUserReader.TenantUserIDs` filter on `auth_identities.user_id`, then enrich display names via `UserDisplayNameReader`). Reads live, off-tx. | `security/` issues **0** `iam_users.display_name` reads (grep — only the genuinely-deferred `MfaCoverage`/`CountRecentLockouts` aggregate JOINs remain, accurately characterized). Each migrated method behavior-identical: same rows (membership-filtered identically), same `display_name` values (port + `missing→user_id` fallback). Live-PG integration per method (locked user surfaces w/ name; tenant isolation; recent-failure window). `go build`+`go vet` clean; existing security tests green/migrated; backend-api-qa-checklist green. |

> **⚠ HS-6 scope/approach reconciliation (2026-06-15, F4.2).** Pre-F4.2 investigation found the
> templates module **already owns** a cross-module port for template-version state —
> `templates/domain.TemplateVersionPort.IsPublished(ctx, versionID) (bool, docTypeCode, error)`, impl
> `templates/infrastructure.TemplateVersionReader` (Wave Z task Z-7, `e50150506`), already consumed by
> taxonomy (`taxonomy/application/profile_service.go:155`). The F4.2 row's literal "**introduce** a new
> `TemplateVersionStateReader`" predated this knowledge. Engineering decision (operator delegated the
> call 2026-06-15 — "study it, reach the best solution by industry standards"): **extend the existing
> owning port, do not introduce a parallel one.** Rationale (DDD / consumer-driven-contracts /
> single-owning-adapter): one adapter must own the `templates_*` SQL + tenant-scoping; a second reader
> over the same tables is the duplication anti-pattern this milestone exists to kill. Raw
> `GetTemplateVersionState → (status, doc_type_code)` is the *primitive*; `IsPublished` is a *derived*
> predicate kept (unchanged) for taxonomy. CD's `Resolve` needs the raw status string
> (`resolution.go:42,55,58` distinguish published/obsolete/draft), and the `status := "published"`
> hardcode at `documents_adapters.go:113` currently **masks a real bug** — an *obsolete* default
> template wrongly passes `resolveDefaultTemplate`. F4.2 row amended above accordingly. The validation
> bar is unchanged (0 `templates_*` SQL under `controlleddocuments/`; 0 `status := "published"` in
> `wiring/`; consumer contract preserved; reads live + off-tx).

For each feature, "what to validate" is **objectively checkable** — a grep that returns zero
cross-module table reads, a build that is clean, a preserved consumer-contract signature, an off-tx
`pg_locks` runtime proof. No "works" / "looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What that gate
enforces for this milestone:

1. **Per-feature acceptance** — F4.1, F4.2, F4.3, **F4.4, F4.5, F4.6** each meet their declared "what
   to validate", and each feature's **consumer contract** (`spec.md`) was honored: F4.1/F4.4/F4.6 port
   consumers get the same display-name semantics; F4.2 preserves CD's `TemplateVersionChecker`
   `(status, doc_type_code)` shape so override-validation is behavior-identical; F4.5's membership port
   reproduces the replaced INNER-JOIN membership set exactly (producer matches consumer; ports read
   from the consumers' existing behavior, not invented).
2. **Workflow-class QA checklists** — `backend-api-qa-checklist` (both ports) **+**
   `workflow-async-qa-checklist` (F4.2 touches the CD-create lock-bearing path).
3. **Regression** — M0, M1, M2, M3 still pass their gates; the M1 test-infra-rebaseline full-HTTP
   `seed→finalize→signoff` E2E stays green (F4.1 touches the signoff display-name read; F4.2 touches
   CD-create).
4. **Quality-bar / root-cause check** — **focused audit slice** on **module-boundaries / DDD**:
   - **H-G reach-without-a-port = 0** in the swept surface: no module queries another module's owned
     tables (grep proof — **true zero** `iam_users.display_name` reads outside iam/ across the whole
     tree, i.e. `auth` + `security` both clean after F4.4/F4.6; no `templates_*` SQL under
     `controlleddocuments/`). The only remaining cross-module `iam_users` reads are the accurately-
     characterized non-display-name aggregate JOINs in `security` (MfaCoverage/CountRecentLockouts),
     recorded as a bounded defer — **not** display-name reaches;
   - **H-G hardcoded-domain-state = 0**: no fabricated domain value (`status := "published"`) in
     wiring/adapters;
   - root cause = the *class* is gone via owning-module ports, **not** a single instance patched and
     not a snapshot/denormalization shortcut (reads stay live, H-PRE-1 intact);
   - **ADRs present** for both ports;
   - dimension re-measured ≥ A−.
5. **No unplanned scope** — anything implemented beyond F4.1–F4.6 is recorded with rationale. The
   security **non-display-name** aggregate JOINs (MfaCoverage/CountRecentLockouts), security's direct
   `auth_identities` reads (separate concern), and any presence/password_reauth handling are recorded
   as explicit bounded defers with triggers (not silent skips). The census-correction trail
   (F4.4–F4.6 added post-validator-FAIL under operator Option 2) is itself the rationale record.

## Dependencies & constraints

- Depends on: M0, M1, M2, M3 passed (this milestone runs on the contract-clean, mechanically-hardened
  baseline they established). F4.1 **generalizes F1.3's contained fix** — contained precedes
  generalization per design D4/Approach-3.
- Architectural constraints respected:
  - **H-PRE-1** advisory-lock deadlock rule (memory `[[advisory-lock-deadlock-constraint]]`): never
    call an authz-recording read on a fresh connection inside a lock-holding atomic tx. F4.1's signoff
    display-name read **stays off-tx** (already hoisted in F1.3 — must not regress); F4.2's template
    status read **stays off** the CD-create lock-holding tx.
  - **Reads stay live; no migrations; no snapshot/denormalization semantics** (D4/Approach-3 — Approach
    2 "freeze actor name" was explicitly rejected absent a separate audit/legal product decision).
  - **Owning-module ownership:** `UserDisplayNameReader` is **iam-owned** (port in iam/domain, impl in
    iam/infrastructure); `TemplateVersionStateReader` is **templates-owned**. Consumers depend on the
    port interface, never on the producer's tables.
  - **Contract-stability:** no OpenAPI/route shape changes expected (these are internal module ports).
    F4.2 preserves CD's existing `TemplateVersionChecker` Go interface contract. If any feature would
    force a public contract shape change → HS-2.
  - **Surgical-changes rule (CLAUDE.md §5.3):** each feature touches only its named sites + the new
    port files; no adjacent refactor, no opportunistic cleanup.
  - **One PR per port** (governing spec §M4).

## Applicable hard-stops

Default catalog HS-1..HS-6 in force. What would trip each here:

- **HS-1** — milestone close gate. On validator PASS, the **main session** flips status and presents
  to the operator; **no M5 open and no merge without explicit approval.** (M4 is the last remediation
  milestone; M5 is the authoritative re-audit.)
- **HS-2** — if introducing either port turns out to require redesign outside its boundary (e.g. the
  templates module has no clean domain seam to own `TemplateVersionStateReader` without a cross-module
  auth/authz model change, or F4.1 can't be done without changing the signoff tx/lock model):
  **stop**, report the boundary + minimum prerequisite plan, do **not** symptom-patch. Esp. guard
  F4.2 against an H-PRE-1-violating "read inside the lock" shortcut.
- **HS-3** — if a prerequisite boundary fails (build / runnable / auth-session / CD-create or signoff
  route / contract truth) while executing a feature: repair via `runtime-contract-prereq`, rerun the
  failed checkpoint, then resume.
- **HS-4** — validator returns FAIL (a reach not actually removed, a hardcode still present, a port
  placed in the wrong module, a snapshot shortcut taken, H-PRE-1 violated, or a consumer contract
  silently changed): open the named fix feature, re-run its lifecycle, re-dispatch the validator.
- **HS-5** — N/A mid-program (terminal acceptance is M5).
- **HS-6** — scope drift: if execution surfaces cross-module reaches beyond the classified set, **stop**,
  surface it, replan before absorbing it. **TRIPPED 2026-06-15** (post-validator-FAIL): the original
  census undercounted the H-G display-name class by 3 (security's `ListLockouts`/
  `CountRecentFailedLoginsByUser`/`ListNewDeviceLogins` were mischaracterized as non-display-name
  JOINs) plus the validator's 1 (`sessions_admin`). Surfaced to operator; **replanned under Option 2
  (full close)** → F4.4/F4.5/F4.6 added above; census corrected. Resolution recorded in program
  `README.md` hard-stops.
