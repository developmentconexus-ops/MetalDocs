# ADR 0082 — Approval kernel extraction to a first-class module (supersedes ADR 0072)

> **Status:** Accepted 2026-07-12
> **Supersedes:** [ADR 0072](0072-approval-nested-exception-and-boundary-model.md) ruling (a) only.
> **Scope:** ROADMAP unit 3.1 / approval-remediation M3 — promote
> `internal/modules/documents/approval` to a top-level bounded context `internal/modules/approval`
> (the 15th module), realign the module-boundary guard to treat it as first-class.

## Context

ADR 0072 (Accepted 2026-07-06, M9 F9.5) **rejected** promoting `documents/approval` to its own
top-level module *at that time*, on the ground that it was one DDD aggregate with dense bidirectional
coupling to `documents`, and that splitting it mid-hygiene-milestone would be an interface redesign
with no functional gain. Crucially, ADR 0072 recorded a **named promotion trigger** verbatim:

> *"if approval ever needs an independent lifecycle — its own deploy cadence, its own owning team, or
> a **second bounded-context consumer that isn't `documents`** — the promotion plan starts from this
> ADR's coupling-edge inventory rather than re-deriving it."*

**The trigger has fired.** The ratified review/approval workflow model (spec
`docs/superpowers/specs/2026-07-08-approval-workflow-coherence-design.md` §5, Milestone C) rewires the
`templates` module's approval onto the same kernel — a **second bounded-context consumer that is not
`documents`**. Executing the promotion now is following ADR 0072's own recorded plan, not reversing
its reasoning.

### Live coupling census (2026-07-12, production `.go`, non-test)

ADR 0072's 2026-07-06 "dense, bidirectional, 100+ edges" estimate counted test edges and is stale.
The live production coupling is materially lighter:

| Edge | Count | Layer | Post-extraction disposition |
|---|---|---|---|
| `documents` → `approval` | 2 files | `application` (delivery/http handler), `domain` (infrastructure/active_instance_reader) | Allowed — published layers; import-path rename only |
| `approval` → `documents` | 24 | 17 `domain` + 7 `application` | Allowed — published layers (approval now depends on documents' published surface) |
| `jobs/approval_sla_surfacer` → `approval` | 1 | `domain` | Allowed |
| `jobs/stuck_instance_watchdog` → `approval` | 1 | `application` | Allowed |
| `templates` → `approval` (new, M3 P3) | — | `application`/`api` | Allowed — published surface only |
| `audit` → `approval/http/router` | **0** | — | **False positive** — the reference is a code *comment*, not a Go import (`sed -n '/^import (/,/^)/p'` confirms no approval import in the audit handler). No re-port required. |

**True cross-module violations after the pure relocate: 0.** `check-module-boundaries.ps1` →
`[module-boundaries] OK` on the relocated tree, because approval's published surface
(`domain`/`application`/`api`) is already covered by the guard's layer allow-list.

## Decision

### (a) `documents/approval` → `internal/modules/approval` — promoted (reverses ADR 0072 ruling (a))

`documents/approval` becomes the top-level bounded context `internal/modules/approval` (the 15th
module). The relocation is a **pure move** (`git mv`, 165 renamed files, byte-identical staged content;
import prefix `metaldocs/internal/modules/documents/approval` → `metaldocs/internal/modules/approval`
across 111 `.go` files + lint/staticcheck path-string consumers; one `//go:generate` relative-path
depth fix). ZERO behavior change — `go build`/`go vet`/unit suites green post-move.

The kernel generalizes from document-specific keying to `(subject_kind, subject_key)` (M3 phases 2–3;
`document+profile_code` existing rows backfilled, `template+doc_type` added). `documents` and
`templates` consume the kernel through its published application-service / api surface only.

### (b) Boundary-guard realigned — approval is first-class, nested-exception retired

`scripts/check-module-boundaries.ps1` no longer special-cases a `documents/approval` nested family.
`approval` is treated exactly like any other module: cross-module imports may target only its published
surface (`domain`/`application`/`api`). This is **stricter** than the old nested model, which allowed
edges between `documents` and `documents/approval` **at any layer**: `documents` may now no longer reach
`approval/http` or `approval/infrastructure`, only approval's published layers (and vice-versa). The
dead `$approvalPublishedExtra` variable and the `$bothInDocumentsFamily` bypass are removed.

**Proofs (P1.S4, mirroring ADR 0072's discipline):**
1. GREEN on the relocated tree with the realigned guard.
2. Negative plant: a blank import `_ "metaldocs/internal/modules/approval/infrastructure"` added to a
   genuinely-external module (`internal/modules/jobs/stuck_instance_watchdog/job.go`) → guard RED,
   naming exactly `internal/modules/jobs/stuck_instance_watchdog/job.go ->
   metaldocs/internal/modules/approval/infrastructure`.
3. Revert: `git diff --exit-code` on the planted file = clean; guard GREEN again.

### What ADR 0072 rulings survive

Rulings **(b)** (one `infrastructure/` persistence directory per module) and **(c)** (the boundary-guard
allow-model keyed to REQ-TOP-1: layer allow-list + `$publishedPackages` + empty `$debtAllowList`) remain
in force unchanged. Only ruling (a) — "approval stays nested" — is superseded.

## Consequences

- **Positive.** Approval is a clean one-module-one-directory bounded context; the boundary guard is now
  stricter around it (persistence/http no longer cross-reachable between documents and approval);
  templates and documents share one approval kernel instead of two parallel approval implementations.
- **Costs.** 111 import paths moved (mechanical, compiler-verified). `documents` and `approval` now have
  a legitimate cross-module dependency in both directions, each through published layers — this is
  normal inter-module coupling, guarded by the layer allow-list.
- **Superseded trigger.** ADR 0072's promotion trigger is now consumed; no further "promote approval"
  decision is pending.

### Transitional coexistence — M3 delivers the kernel, ROADMAP 3.1a retires the legacy path (operator-ratified 2026-07-12)

M3 makes the kernel the **backend truth** for template approval (subject-generic instances/signoffs,
subject-discriminated tripwire per [ADR 0083](0083-subject-discriminated-capability-tripwire.md), two
additive contract-first routes `POST /templates/{id}/versions/{n}/submit-for-approval` + `/signoff`,
validated by service- and repository-level real-DB integration). It does **not** yet make the kernel
the *sole* path. During #11 close-out, investigation verified two facts that make an in-M3 "delete all
legacy, no fallback" retirement non-executable without scope beyond this milestone:

1. **The role-based model is not confined to the 4 legacy approval routes.** `templates_approval_config`
   (`template_id` PK, `reviewer_role`, `approver_role`), `domain.ApprovalConfig`, the version
   `PendingReviewerRole`/`PendingApproverRole` fields, `RoleBindingFor`, and `CheckSegregation` are
   **still load-bearing** for two non-legacy paths: `CreateTemplate` seeds the approver/reviewer *role*
   and writes the config table at creation (`application/create.go:83`), and `PublishTemplateVersion`
   performs role-based SoD + role-binding tier-2 on the direct-publish path (`application/lifecycle.go:421-427`).
   The table/type therefore cannot be dropped without also migrating create + publish off the role
   model — a materially larger change than the 4 named routes.
2. **The frontend has zero kernel-route consumer.** `TemplateApprovalRoute.tsx` / `TemplateEditorPage.tsx`
   call only the legacy routes; no FE code calls the kernel routes. Deleting the legacy backend routes
   in M3 would 404 the entire template-approval UI with no replacement.

**Ratified resolution (Option A):** defer the full retirement to a sequenced dedicated follow-on milestone
(**ROADMAP 3.1a**, "template legacy-approval retirement" — distinct from 3.2 ActorSelector), which does,
atomically and in order: (a) migrate `CreateTemplate` (stop seeding roles) and `PublishTemplateVersion`
(kernel-driven completion, remove role-SoD) off the role model; (b) rebuild the template-approval frontend
onto the kernel routes; (c) then delete the legacy path (4 routes + handlers + `Service.SubmitForReview/Review/Approve/UpsertApprovalConfig`
+ `approval_config.go` + `GetApprovalConfig`/non-Tx `UpsertApprovalConfig` repo methods + legacy tests + FE
consumers) and drop `templates_approval_config` behind a pre-drop emptiness assert. Until M4, the kernel and
the legacy role-based path **coexist**; this is a named, tracked transitional debt, not a silent fallback —
the kernel is the target sole path (decision unchanged), M4 is where legacy dies. Orphan bookkeeping to fold
into M4: capability `CapTemplateReview` becomes functionally unreferenced once the legacy `review` route is
deleted (registry + `arms.go` allowlist cleanup), and `domain.ApprovalConfig`/`ErrInvalidApprovalConfig`
collapse once create + publish stop consuming them.

### Retirement executed — ROADMAP unit 3.1a (2026-07-13)

Phases (a), (b), and (c) of the ratified Option A retirement plan above were all executed atomically by
ROADMAP unit 3.1a (branch `worktree-agent-a0d0c8a51f43bb656`): `CreateTemplate` and `PublishTemplateVersion`
were migrated off the role model, the template-approval frontend was rebuilt onto the kernel routes, and the
legacy path (4 routes + handlers + `Service.SubmitForReview`/`Review`/`Approve`/`UpsertApprovalConfig` +
`domain.ApprovalConfig` + the `GetApprovalConfig`/non-Tx `UpsertApprovalConfig` repo methods + legacy tests +
FE consumers) was deleted.

**The transitional coexistence described above has ENDED.** The approval kernel
(`submit-for-approval` → `signoff` → `publish`) is now the sole template-approval path — there is no
remaining role-based fallback. Orphan bookkeeping named above was closed: capability `CapTemplateReview`
is retired (capability registry 40→39, registry + `arms.go` allowlist cleanup), and
`templates_approval_config` is dropped by forward migration `0302_drop_templates_approval_config.sql`
behind a pre-drop emptiness assert. The `pending_reviewer_role`/`pending_approver_role` columns on
`templates_template_version` are retained as named debt (write-never/read-never), ratified out of the
migration-0302 drop list rather than dropped alongside the table.

See `docs/superpowers/reports/2026-07-13-unit-3.1a-evidence.md` for full execution evidence.

## Alternatives considered

- **Keep the nested exception, add templates as a third intra-documents nest** — rejected: templates is
  not part of the documents aggregate; nesting it under documents to reuse approval would couple two
  unrelated bounded contexts and defeat REQ-TOP-1.
- **Two parallel approval implementations (status quo)** — rejected: the ratified model requires the
  same eQMS rigor (SoD, quorum, delegation, e-signature, instance state machine) for template versions
  as for documents; duplicating the kernel is the local maximum ADR 0072 warned against.
