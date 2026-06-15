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
> | `security/infrastructure/postgres/repository.go` (lockouts / active-sessions / MFA counts) | `iam_users` as **JOIN for tenant-scoping** on `auth_identities`/`auth_sessions` (no display-name read) | cross-module, but a **tenant-scope JOIN**, not a display-name read — a `UserDisplayNameReader` does not cover it; the correct port (if any) is a different one | **No** — different concern (see scope decision below) |
> | `documents/approval/infrastructure/signature/password_reauth.go` | `iam_users.password_hash` | already behind an `IamUserReader` interface; reauth, not display-name | **No** — already a port, different column |
>
> **Net F4.1 scope:** the **3 display-name reads** (decision_service, documents/repository,
> get_instance_handler). The other named sites are classified out-of-port above. **Scope decision
> RESOLVED by operator 2026-06-15: defer security.** `security`'s `iam_users` tenant-scoping JOINs are
> **not** an M4 target — they are a JOIN for tenant binding within security's own bounded read, not a
> display-name reach, and would need a *different* port than `UserDisplayNameReader`. Recorded as a
> **bounded defer** — *Trigger:* next structural touch of `security/infrastructure/postgres/repository.go`
> or an M5 re-audit finding that flags it as an H-G reach; *Owner:* backend; resolved via a dedicated
> iam tenant-scope/membership port, not by widening M4. F4.2 and the hardcode site verified exactly as
> the spec names.

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

1. **Per-feature acceptance** — F4.1, F4.2, F4.3 each meet their declared "what to validate", and each
   feature's **consumer contract** (`spec.md`) was honored: F4.1's port consumers get the same
   display-name semantics; F4.2 preserves CD's `TemplateVersionChecker` `(status, doc_type_code)`
   shape so override-validation is behavior-identical (producer matches consumer; the port is read
   from the consumer's existing contract, not invented).
2. **Workflow-class QA checklists** — `backend-api-qa-checklist` (both ports) **+**
   `workflow-async-qa-checklist` (F4.2 touches the CD-create lock-bearing path).
3. **Regression** — M0, M1, M2, M3 still pass their gates; the M1 test-infra-rebaseline full-HTTP
   `seed→finalize→signoff` E2E stays green (F4.1 touches the signoff display-name read; F4.2 touches
   CD-create).
4. **Quality-bar / root-cause check** — **focused audit slice** on **module-boundaries / DDD**:
   - **H-G reach-without-a-port = 0** in the swept surface: no module queries another module's owned
     tables (grep proof — no `iam_users` display-name SQL outside iam/; no `templates_*` SQL under
     `controlleddocuments/`);
   - **H-G hardcoded-domain-state = 0**: no fabricated domain value (`status := "published"`) in
     wiring/adapters;
   - root cause = the *class* is gone via owning-module ports, **not** a single instance patched and
     not a snapshot/denormalization shortcut (reads stay live, H-PRE-1 intact);
   - **ADRs present** for both ports;
   - dimension re-measured ≥ A−.
5. **No unplanned scope** — anything implemented beyond F4.1–F4.3 is recorded with rationale. The
   security tenant-scope JOIN decision and any presence/password_reauth handling are recorded as
   explicit bounded defers with triggers (not silent skips).

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
- **HS-6** — scope drift: if execution surfaces cross-module reaches beyond the classified set (the 3
  display-name reads + the F4.2 reach + the hardcode), **stop**, surface it, replan before absorbing
  it. The **security tenant-scope JOIN** is the known open scope decision — resolve it at Phase 2
  (operator), not mid-execution.
