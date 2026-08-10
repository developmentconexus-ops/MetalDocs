# Pass 05 — Database Ownership Graph

> **Date:** 2026-08-09
> **Baseline:** `main@418070bf`
> **Status:** reproduced-current
> **Scope:** `internal/modules/**/*.go` (production files only, `_test.go` excluded), cross-checked against `wiki/database/`, `db/baseline/0001_current_schema.sql`, `internal/composition/tenantdata/registry/registry.go`.
> **Method:** `grep -rnE "(FROM|JOIN|INSERT INTO|UPDATE|DELETE FROM)\s+[a-z_.]+" internal/modules --include=*.go`, comment lines stripped, then every table identifier resolved against the ownership source in §1 and read against `db/baseline/0001_current_schema.sql` for view vs. base-table identity. Raw evidence: `sql-edges-nocomment.txt` (583 matched lines) built during this pass, retained in the session scratchpad, not committed (mechanical intermediate, not durable evidence — the citations below are the durable artifact).

## 0. Answer up front

- **55 foreign-table READ statements**, **12 foreign-table WRITE statements** (10 of the 12 writes are the *same* pattern: `approval` executing raw `UPDATE documents` on `documents`' primary table). Counted at statement/call-site granularity — see §4 for the full enumeration and §5 for three alternative granularities so the number is reproducible under any counting convention.
- The worst offender is **`approval` writing directly into `documents`** (10 distinct call sites, all raw `UPDATE documents ... SET status = ...`) — this is **more severe than any read finding** in #93, which described the approval↔documents relationship as read-only. It is not: approval also owns part of the document status lifecycle via inline SQL, with no producer-owned port, while the *identical* problem for templates was already solved in-repo (`TemplateCompletionWriter`, §6).
- #93's "17+ foreign-table reads" **undercounts** at every granularity except the coarsest (directed module-pair edges, §5c = 10). At statement granularity the real number is 55. The three approval→documents / documents→approval / approval→controlleddocuments claims in #93 are all reproduced (§5), but #93 named only 3 of the 10 directed module-pair edges that actually exist in production code.

## 1. Ownership source assessment (task step 1)

The closest existing machine-readable ownership declaration is **`wiki/database/dictionary-index.md`**: a single Markdown table `Table | Schema | Owner | Page`, sourced from `db/baseline/0001_current_schema.sql`, last verified 2026-07-16. It is:

- **Structured enough to parse** (one row per table, one `Owner` column) — a script can `awk`/regex it today.
- **Not machine-generated** — hand-maintained, and it has drifted: **10 tables that exist in the baseline and have their own per-table dictionary page under `wiki/database/tables/*.md` (each with an explicit `> **Owner:**` line) are missing from the index table itself**: `release_generations`, `approval_review_verdicts`, `approval_route_stage_selectors`, `approval_delegations`, `audit_export_jobs`, `tenant_keys`, `tenant_lifecycle_jobs`, `tenant_plans`, `token_dictionary_entries`, `materialize_dispatch_outbox`. Their per-table pages are authoritative and were used below; the index page itself is the gap.
- **Silent on views** — `metaldocs.v_active_user_areas`, `v_cd_grantee`, `v_cd_search_facts`, `v_cd_obligated_readers`, `v_document_search_facts`, `v_process_area_name` are real, deliberately-designed read contracts (defined + `COMMENT ON VIEW`-documented in `db/baseline/0001_current_schema.sql:1687-2004`, citing ADR-0039/ADR-0037/ADR-0041 and the `backend-module-boundary-hardening` mission) but are absent from the ownership dictionary entirely, so a naive "grep table names, look up owner" scan misclassifies every legitimate projection read as a violation unless it special-cases views (this pass did — see §7).

**Decision:** do not hand-author a second ownership list. §2 below is `dictionary-index.md` plus the 10 missing rows pulled from their own (already-authoritative) per-table pages, presented as one corrected table for this pass's use — the correction belongs in the index itself (see §8), not as a new parallel artifact.

No ADR 0092 file exists in `wiki/decisions/` (only 0090, 0091, 0093) — #93's own text references "#92/A5" as an issue/work-item pairing, not an ADR; nothing to reconcile here.

## 2. Table → owning module (corrected)

Legend: **B** = platform/shared (not a bounded-context table, see §6), **D** = dead/unused in current Go code (zero readers/writers found).

| Table | Schema | Owner module | Note |
|---|---|---|---|
| `documents` | `public` | documents | the live, richly-governed table |
| `documents` | `metaldocs` | documents | **D** — zero references in `internal/`, `apps/` outside `_test.go`; legacy/dead schema object |
| `document_attachments`, `document_collaboration_presence`, `document_edit_locks`, `document_images`, `document_version_images`, `document_versions`, `document_versions_mddm`, `document_checkpoints`, `document_comments`, `document_exports`, `document_placeholder_values`, `document_revisions`, `editor_sessions`, `autosave_pending_uploads`, `mddm_shadow_diff_events` | mixed | documents | |
| `document_departments`, `document_families`, `document_process_areas`, `document_profiles`, `document_profile_governance`\*, `document_profile_schema_versions`\*, `document_profile_template_defaults`\*, `document_type_schema_versions`, `document_types` | mixed | taxonomy | \*dropped migration 0308 |
| `document_template_assignments`, `document_template_versions`, `document_template_versions_mddm`, `templates_template`, `templates_template_version`, `template_audit_log`, `template_drafts` | mixed | templates | |
| `templates`, `template_versions`, `templates_approval_config`, `templates_audit_log` | `public` | templates | **D** — retired/dropped, historical |
| `approval_instances`, `approval_routes`, `approval_route_stages`, `approval_route_stage_selectors`, `approval_signoffs`, `approval_stage_instances`, `approval_review_verdicts`, `approval_delegations`, `release_generations`, `auth_failure_counters` | `public` | approval | `release_generations` owner corrected here (missing from index; see own page, ADR 0085). `auth_failure_counters` is approval-owned despite the `auth` prefix — approval's own e-signature lockout counter, not `auth` module's. |
| `workflow_approvals` | `metaldocs` | approval | **D** — zero references outside `_test.go` |
| `controlled_documents`, `controlled_document_area_grants`, `controlled_document_user_grants`, `cd_sequence_counters` | `public` | controlleddocuments | |
| `auth_identities`, `auth_sessions` | `metaldocs` | auth | |
| `iam_users`, `iam_user_roles`, `iam_groups`, `iam_group_members`, `iam_group_roles`, `role_capabilities` | `metaldocs` | iam | |
| `user_process_areas` | `public` | iam | |
| `tenant_lifecycle_jobs` | `metaldocs` | iam | corrected here (missing from index; M7 F7.3) |
| `tenant_plans` | `metaldocs` | iam (Admin Center observability) | corrected here (missing from index) |
| `tenants` | `metaldocs` | **B** platform root | tombstoned directly by the erase orchestrator, never through a `TenantDataPort` — `registry.go:41-47` |
| `tenant_keys` | `metaldocs` | security | corrected here (missing from index; crypto-shred, M7 F7.3) |
| `audit_events`, `audit_export_jobs`, `governance_events` | mixed | audit | `audit_export_jobs` and `release_generations` both corrected here (missing from index) |
| `token_dictionary_entries` | `metaldocs` | tokens | corrected here (missing from index) |
| `notifications` | `metaldocs` | **B** platform/workers (de facto notifications module) | |
| `idempotency_keys`, `job_leases`, `outbox_events`, `schema_migrations` | mixed | **B** platform | |
| `pdf_dispatch_outbox` | `metaldocs` | **B** platform queue, written by render | |
| `materialize_dispatch_outbox` | `metaldocs` | **B** render/fanout's own dispatch queue | corrected here (missing from index) |

**Modules with zero owned tables:** `distribution`, `search`, `jobs` — both are pure consumers (see §7).

## 3. Deliberate projections (task step 6) — read BEFORE §4, these are NOT violations

`db/baseline/0001_current_schema.sql:1687-2004` defines 6 views, each with a `COMMENT ON VIEW` stating "non-owner modules read THIS view, never the base table" and citing the owning ADR/mission:

| View | Base table(s) | Owner | Declared consumers | Evidence |
|---|---|---|---|---|
| `metaldocs.v_active_user_areas` | `public.user_process_areas` | iam | approval, controlleddocuments | `postgres_approval_repository.go:1874,1984`, `controlleddocuments/infrastructure/repository.go:172,557` |
| `metaldocs.v_cd_grantee` | `controlled_documents` + grants + `v_active_user_areas` | controlleddocuments | search | `search/infrastructure/v2documents/reader.go:109` |
| `metaldocs.v_cd_search_facts` | `controlled_documents` | controlleddocuments | search, (transitively `v_cd_obligated_readers`) | `search/infrastructure/v2documents/reader.go:76` |
| `metaldocs.v_cd_obligated_readers` | grants + `v_active_user_areas` + `v_cd_search_facts` | controlleddocuments | distribution, notifications | `distribution/infrastructure/coverage_repository.go:51,154,175,199,311`; `notifications/infrastructure/fanout_worker.go:106` |
| `metaldocs.v_document_search_facts` | `public.documents` | documents | search | `search/infrastructure/v2documents/reader.go:75` |
| `metaldocs.v_process_area_name` | `metaldocs.document_process_areas` | taxonomy | distribution | `distribution/infrastructure/coverage_repository.go:155,176,200,312` |

This is the exact target pattern the rest of this document argues for extending to the raw-table edges in §4: a producer-owned, DB-enforced read contract instead of a raw cross-schema `SELECT`. It already covers **every** SQL read that `search`, `distribution`, and `controlleddocuments`'s IAM-area lookups perform — those three consumers have **zero** raw foreign-table reads in production code. None of these view reads are counted in §4/§5.

## 4. Foreign-edge enumeration (task steps 2–4)

### 4a. Foreign WRITES (12 statements — the more severe class)

| # | Writer module | Foreign table (owner) | Evidence | What it writes | Proposed contract |
|---|---|---|---|---|---|
| 1 | approval | `documents` (documents) | `internal/modules/approval/application/cancel_service.go:171` | `status='draft'` on cancel | consumer-owned port, see below |
| 2 | approval | `documents` (documents) | `internal/modules/approval/application/decision_service.go:767` | status on legacy signoff completion path | " |
| 3 | approval | `documents` (documents) | `internal/modules/approval/application/document_terminal_approval.go:129` | `status='approved', revision_version+1` (the shared `completeDocumentTerminalApproval` helper — comment at line 90 names this "the document-only `UPDATE documents` path", explicitly contrasted with the templates port) | " |
| 4 | approval | `documents` (documents) | `internal/modules/approval/infrastructure/postgres_approval_repository.go:916` | status transition | " |
| 5 | approval | `documents` (documents) | `internal/modules/approval/application/mark_reviewed_service.go:191` | `last_reviewed_at`/OCC CAS | " |
| 6 | approval | `documents` (documents) | `internal/modules/approval/application/obsolete_service.go:137` | `status='obsolete'` OCC CAS | " |
| 7 | approval | `documents` (documents) | `internal/modules/approval/application/release_coordinator.go:350` | publish-path status write (self-lock) | " |
| 8 | approval | `documents` (documents) | `internal/modules/approval/application/release_coordinator.go:484` | publish-path status write | " |
| 9 | approval | `documents` (documents) | `internal/modules/approval/application/submit_service.go:557` | submit-time status write | " |
| 10 | approval | `documents` (documents) | `internal/modules/approval/application/review_verdict_service.go:470` | review-verdict status write | " |
| 11 | approval | `governance_events` (audit) | `internal/modules/approval/application/events.go:84` | `INSERT INTO governance_events` — approval's own `EventEmitter`/`sqlEmitter`, the sole non-test literal `INSERT` into `governance_events` anywhere in `internal/modules` | producer query service: audit should own this table's writer, approval should call an `audit`-owned `GovernanceEventEmitter` port instead of holding the SQL text |
| 12 | iam | `governance_events` (audit) | `internal/modules/iam/application/tenant_lifecycle_service.go:601` | `DELETE FROM public.governance_events WHERE tenant_id=$1` during tenant erasure — **separate from** and **in addition to** audit's own `TenantDataPort.EraseTenantData`, which also deletes `governance_events` (`internal/modules/audit/infrastructure/postgres/tenant_data_port.go:123`) | route entirely through `registry.AllTenantDataPorts` / audit's port; iam should not hold a second, parallel raw-SQL delete path for a table it doesn't own |

**All 10 approval→documents writes are the same class**: raw `UPDATE documents` inline in approval application/infrastructure code, no producer-owned port. This is provably fixable in-repo — see §6.

### 4b. Foreign READS (55 statements)

Grouped by directed module pair; every row is a statement-level citation (one row = one SQL statement, its own `file:line`), not a table-name match.

**approval → documents (25):** `cancel_service.go:122`, `mark_reviewed_service.go:175`, `obsolete_service.go:119`, `read_service.go:504,711,1029`, `release_coordinator.go:183,724,742`, `release_facts.go:143`, `release_terminal_approval.go:185`, `postgres_approval_repository.go:366,440,505,571,623,647,669,690,714,1205,2025,2091`, `release_hold_reader.go:93,133`.
Reason: approval needs the document's current `status`, `process_area_code_snapshot`, `controlled_document_id`, `revision_version` to run its own OCC CAS checks and area-grade authz before writing its own instance/stage rows.

**approval → documents.document_comments (2):** `postgres_approval_repository.go:1775,1793` — reason: approval's `ErrApprovalBlockedByUnresolvedComments` check needs unresolved-comment counts before allowing a decision.

**approval → documents.document_revisions (1):** `postgres_approval_repository.go:2090` — reason: resolving the revision behind a document join for area/route derivation.

**approval → controlleddocuments.controlled_documents (1):** `read_service.go:717` (`LEFT JOIN controlled_documents cd`) — reason: worklist projection needs the CD identifier/label for template-subject and document-subject rows in one shape.

**documents → approval (9):** `context_builder.go:49`, `active_instance_reader.go:159`, `resolver_readers.go:102,142,144,179,180`, `repository.go:2053,2054` — reason: documents needs to know whether an active approval instance exists (to gate edit locks / active-session logic) and to resolve signoff counts for its own resolver read-model.

**documents → approval.release_generations (1):** `repository.go:380` (`LEFT JOIN LATERAL ... FROM release_generations rg`) — reason: documents' own read model needs the latest release-generation pointer for a document row.

**iam → auth.auth_identities (3):** `observability_repository.go:144-145,188`, `presence/repository.go:65` — reason: Admin Center / presence views need identity fields (email, display data) joined onto `iam_users`.

**iam → audit.audit_events (4):** `observability_repository.go:69,81,102,166` — reason: Admin Center observability surfaces recent audit activity per user/tenant.

**security → auth.auth_identities (3):** `repository.go:128,194,248` — reason: session-anomaly/failed-login detection needs identity rows.

**security → auth.auth_sessions (2):** `repository.go:277,284` — reason: same detection logic, prior/concurrent session comparison.

**security → audit.audit_events (1):** `repository.go:358` — reason: time-of-day anomaly detection reads recent audit events directly.

**templates → audit.audit_events (1):** `postgres.go:766` — reason: templates' own audit-trail surface reads `audit_events` directly instead of through an audit-owned query port.

**auth → iam.iam_users (1):** `tenant_data_port.go:81` (`SELECT user_id FROM metaldocs.iam_users WHERE tenant_id=$1` as a subquery inside auth's own tenant-erase `DELETE`) — reason: `auth_identities` has no `tenant_id` column; auth must resolve tenant scope through iam's user table to erase correctly.

**jobs → approval (1 statement, 2 tables):** `jobs/stuck_instance_watchdog/job.go:136-137` (`FROM approval_instances ai LEFT JOIN approval_stage_instances asi`) — reason: the watchdog (alert-only, ADR 0068) detects stuck in-progress instances. Read-only, `authz.BypassSystem` background path.

Total: 25+2+1+9+1+3+4+3+2+1+1+1+1(×2 tables) = **55 statements**.

## 5. Reproducing / correcting #93's "17+ foreign-table reads"

#93's finding (`docs/superpowers/analysis/2026-08-09-metaldocs-architecture-reproduced-inventory.md:100-108`) named exactly 3 bullet groups:

- "Approval → Documents: multiple `documents` / `document_comments` reads in `postgres_approval_repository.go`" — **reproduced**, and undercounted: it names only `postgres_approval_repository.go`, but the same edge also lives in `cancel_service.go`, `mark_reviewed_service.go`, `obsolete_service.go`, `read_service.go`, `release_coordinator.go`, `release_facts.go`, `release_terminal_approval.go`, `release_hold_reader.go` — 8 more files.
- "Documents → Approval: reads of `approval_instances`, `approval_signoffs`, `release_generations`" — **reproduced exactly**, all three tables confirmed (`release_generations` ownership corrected to approval in §2, so this is genuinely a foreign read as #93 states).
- "Approval → ControlledDocuments: direct `controlled_documents` join" — **reproduced exactly**, single site `read_service.go:717`.

#93 did **not** name: iam→auth, iam→audit, security→auth, security→audit, templates→audit, auth→iam, jobs→approval. Those 6 directed pairs (16 statements) are real and are new to this pass, not previously filed.

Three honest ways to state "the number":

- **(a) statement/call-site granularity** (this pass's primary count): **55 reads + 12 writes = 67 foreign-table SQL statements**.
- **(b) distinct foreign table names touched:** `documents`, `document_comments`, `document_revisions`, `controlled_documents`, `approval_instances`, `approval_signoffs`, `approval_stage_instances`, `release_generations`, `auth_identities`, `auth_sessions`, `audit_events`, `iam_users`, `governance_events` = **13 tables**.
- **(c) directed module-pair edges:** approval→documents, approval→controlleddocuments, documents→approval, iam→auth, iam→audit, security→auth, security→audit, templates→audit, auth→iam, jobs→approval = **10 pairs** (9 read-only pairs + approval→documents also carries the 10 writes).

**Correction:** "17+" is closest to, and still undershoots, granularity (a) restricted to only the 3 pairs #93 named (approval→documents 25+2+1=28, documents→approval 9+1=10, approval→controlleddocuments 1 → **39** statements just for the named pairs, already more than double "17+"). Under any granularity this pass measured, #93's count was an underestimate, not an overestimate — there is no dimension along which "17+" was too high.

## 6. Read-model vs. accidental leakage — and the proof that the target pattern already exists

§3's 6 views are unambiguously deliberate. The §4a writes and most of §4b's approval↔documents traffic are unambiguously not — and the repo already proves the fix is buildable, because it was built once, for a sibling case:

`internal/modules/approval/application/decision_service.go:51-75` defines `TemplateCompletionWriter`, an approval-owned interface (`MarkTemplateVersionApproved`/`MarkTemplateVersionRejected`) whose production implementation lives in `internal/modules/templates/infrastructure/approval_completion_writer.go` — templates owns the SQL against its own `templates_template_version` table, approval only calls the interface. The doc comment at `decision_service.go:56-63` is explicit: *"approval never imports templates infrastructure or writes `templates_template_version` directly — this is the ONLY seam a terminal approve/reject decision crosses the module boundary through."* Line 90 names the parallel document path as *"the document-only `UPDATE documents` path"* — the codebase's own comments already identify this as the exception, not the rule.

This means the correct target for the 10 approval→documents writes and the bulk of the approval↔documents reads is not a new abstraction to invent — it is applying the `TemplateCompletionWriter` pattern that already ships for templates: a `DocumentCompletionWriter` (or equivalent) interface owned by approval, implemented in `internal/modules/documents/infrastructure`, so `documents` executes its own status-transition SQL on the caller's `tx`. The reverse edge (documents reading `approval_instances`/`approval_signoffs`/`release_generations` for its own resolver/read-model) is a genuine consumer-owned-read-port candidate — declare a `documents`-owned reader interface (mirroring `documents/application`'s existing `DictionaryValueReader` pattern, cited in the parent inventory §5) and implement it once in `approval/infrastructure`, rather than `documents` holding `approval_instances`/`approval_signoffs` SQL directly.

## 7. Platform/shared vs. bounded-context tables (task step 5)

**Platform/shared (a module touching these is not a violation):**
`idempotency_keys`, `job_leases`, `outbox_events`, `schema_migrations`, `pdf_dispatch_outbox`, `materialize_dispatch_outbox`, `notifications`, `tenants`. River's own tables (not scanned — River owns its schema outside `internal/modules`) belong in this class too.

`audit_events` is the ambiguous one the parent inventory flagged with a "?": it has exactly one writer (`audit/infrastructure/postgres/writer.go`, hash-chained append-only ledger) and multiple cross-module *readers* (iam, security, templates, all §4b). That read pattern (single owner-writer, many foreign readers going straight at the base table instead of a view) is exactly what the `v_*` views in §3 exist to prevent for the CD/search/distribution family — `audit_events` never got the same treatment. Recommend: **treat `audit_events` as bounded-context (owned by audit)**, not platform-shared, specifically because the write side is genuinely single-owner business data (compliance ledger), and give it the same `v_*`-projection treatment as `v_document_search_facts`.

`governance_events` shows a writer/owner mismatch: dictionary-owned by audit, but the sole day-to-day writer is approval (§4a #11), and audit itself only touches it via the tenant-erasure `TenantDataPort`. **Ownership ruling: audit stays the owner.** ADR 0044 explicitly defines `governance_events` as the actor-centric audit log; the current writer location is implementation evidence, not ownership evidence (using it to relabel the owner would be exactly the ME-13 circular-evidence class this audit forbids). Classification of the mismatch: approval's direct `INSERT` is a **foreign write** to be replaced by an audit-owned `GovernanceEventEmitter`-style capability port; iam's erasure `DELETE` is a **foreign write** to be routed through audit's tenant-data port. No dictionary ownership relabel — target access goes through the owner's ports. Supersedes any earlier suggestion in this pass to correct the ownership label toward approval.

## 8. ADR 0093 (A9 Controlled Information) reclassification (task step 7)

ADR 0093 folds `documents` + `templates` + `controlleddocuments` into one future "Controlled Information" context. Applying that lens to every edge in §4:

| Edge | Becomes intra-context after A9? | Why |
|---|---|---|
| approval → documents (25 reads, 10 writes) | **No** — stays cross-context | approval is explicitly excluded from A9 and is architecturally required to stay subject-generic (`subject_kind`/`subject_key`) per ADR 0082/0083; it must never become documents-specific, so this is the highest-priority genuine seam to fix with §6's port pattern |
| approval → controlleddocuments (1 read) | **No** — stays cross-context | same reason; approval reading `controlled_documents` for a template-subject row must not become an argument for approval joining the merged context's tables directly |
| approval → document_comments / document_revisions (3 reads) | **No** | same reason (these are `documents`-owned tables) |
| documents → approval (9 reads + `release_generations` 1 read) | **No** — stays cross-context, same direction reversed | approval stays separate; this is the mirror-image of the same seam |
| iam → auth, security → auth, templates → audit, auth → iam, security → audit, iam → audit | **Not applicable to ADR 0093** | none of these modules are part of the Controlled Information consolidation; they are a separate cluster (identity/audit family) with no governing ADR for a similar merge today — noted as an observation only, not a recommendation to open one |
| jobs → approval (1 read) | **No** | `jobs` is cross-cutting infrastructure, not part of any bounded context |

**Net effect of A9 on this graph: zero of the measured 67 foreign-table statements become intra-context.** A9 removes Go-level import friction between documents/templates/controlleddocuments (the parent inventory's finding 5/8), but the SQL ownership graph measured here is untouched by it — the approval↔documents seam is the one that must be closed by producer/consumer ports regardless of whether or when A9 lands.

## 9. Minimal future enforcement property (task step 8)

Two artifacts, composed:

1. **Ownership as data** — promote §2's corrected table→module map (dictionary-index.md plus the 10 missing rows) into a single machine-readable source (e.g. `db/ownership.yaml` or a Go `map[string]string` literal consumed by both the wiki generator and the verifier), so it stops being hand-synced prose. Views (§3) get an explicit `is_projection: true` / `consumers: [...]` field instead of being absent.
2. **SQL identifier scan in the #87 verifier** — a Go AST or regex pass (this pass's method, made permanent) over `internal/modules/**/*.go`, per package: extract every table identifier following `FROM|JOIN|INSERT INTO|UPDATE|DELETE FROM`, resolve it against artifact (1), and fail if the package's owning module ≠ the table's owning module, *unless* the identifier is a `v_*` view listed as a projection. A negative fixture (a package that does raw `UPDATE` on a foreign table) proves the rule fires, per the plan's Task 6 requirement that every blocking guard have one.

This is a semantic property ("no package writes/reads a table it doesn't own, except through a declared view"), not a syntactic proxy — it directly targets the class of defect in §4, and would have caught all 67 statements measured here on day one.

## 10. Evidence summary for closure

- Foreign READ statements: **55** (§4b), spanning 10 directed module pairs.
- Foreign WRITE statements: **12** (§4a) — the more severe class; 10 of 12 are approval directly mutating `documents`' `status`/`revision_version` columns with no producer-owned port, while the identical templates case already has one (§6).
- Deliberate, DB-documented projections: **6 views**, **0** raw reads from their consumer modules (search, distribution, controlleddocuments' IAM-area lookups) — a clean subset of the graph, already at the target shape.
- Dead/unused tables found in passing (not remediated, just noted): `metaldocs.documents`, `workflow_approvals`.
- Ownership-index gaps found in passing (not remediated, just noted): 10 tables missing from `wiki/database/dictionary-index.md` despite having authoritative per-table pages. `governance_events` writer/owner mismatch: owner stays **audit** (ADR 0044 defines it as the audit log); approval's INSERT and iam's DELETE are classified as foreign writes to be re-routed through audit-owned ports (§7) — not an ownership relabel.
- No product/runtime code was modified in this pass.
