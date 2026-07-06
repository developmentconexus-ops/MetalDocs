# ADR 0072 — `documents/approval` nested exception + boundary-guard published-surface model

> **Status:** Accepted 2026-07-06
> **Scope:** M9 F9.5 structure-hygiene — approval module-birth decision, `repository/`→`infrastructure/`
> rename, and `scripts/check-module-boundaries.ps1` realignment to REQ-TOP-1.

## Context

Three related structural questions surfaced during M9 governance-hygiene (F9.5):

1. `internal/modules/documents/approval/` is a full-layer subtree (`api application domain http
   infrastructure jobs repository`) nested inside the `documents` module rather than living at
   `internal/modules/approval/` as its own top-level bounded context. Should it be promoted?
2. `internal/modules/documents/` and `internal/modules/templates/` each had a `repository/`
   directory (persistence layer) alongside — in templates' case — an existing `infrastructure/`
   directory. Two names for the same layer concept is drift; `documents/approval` had its own
   third `repository/` copy.
3. `scripts/check-module-boundaries.ps1` (the CI/milestone-gate boundary guard) allowed cross-module
   imports to target **only** `<module>/domain`. This is the **inverse** of REQ-TOP-1 ("cross-module
   access goes through a module's application service or published Go interface — never another
   module's repository, SQL, or domain internals"): the old model let a caller reach into another
   module's `application`, `api`, `authz`, `fanout`, `resolvers` packages **only** because they
   weren't named `domain` and therefore fell outside what the script inspected as "non-domain,
   therefore forbidden" — meanwhile it never had a rule that explicitly sanctioned these adjacent
   real published surfaces at all. Running it on the pre-F9.5 tree produced 53 flagged lines, nearly
   all of them the intentionally-published `iam/authz`, `render/fanout`, `render/resolvers` and
   cross-module `application`/`api` edges the system relies on.

## Decision

### (a) `documents/approval` stays nested — explicit ADR-recorded exception

**Promotion to `internal/modules/approval/` is rejected.** Evidence (import sweep, 2026-07-06):
`documents` → `approval` production-code edges are dense and bidirectional — `approval/application`
(29), `approval/http` (20+), `approval/domain` (34) from the documents side, and `approval` itself
imports `documents/domain` (16) and `documents/application` (6) back. This is the approval *workflow
of the documents aggregate* (frozen eligible-actor snapshots, SoD, dialect-B area-scoped authz, all
keyed to document versions) — one DDD bounded context, not two. Splitting it across a module
boundary would require re-porting 100+ edges through published interfaces with no functional gain,
which is an interface redesign outside a hygiene milestone's boundary (HS-2) and would trade a real
local-maximum problem (nesting) for a worse one (aggregate-splitting).

`documents/approval` is therefore a **named, permanent nested exception**:
- Intra-family edges (`documents` ↔ `documents/approval`, either direction, any layer) are
  **internal** to the documents bounded context and are never flagged by the boundary guard.
- External consumers (any module other than `documents`) may import **only** approval's published
  surface: `documents/approval/domain`, `documents/approval/application`, `documents/approval/api`.
  `documents/approval/infrastructure` and `documents/approval/jobs` are off-limits to everyone
  outside the documents family, exactly like any other module's persistence/jobs layer.
- **Promotion trigger** (recorded, not decided now): if approval ever needs an independent lifecycle
  — its own deploy cadence, its own owning team, or a second bounded-context consumer that isn't
  `documents` — the promotion plan starts from this ADR's coupling-edge inventory rather than
  re-deriving it.

### (b) `repository/` → `infrastructure/` rename, mechanical

One persistence-layer directory name exists per module going forward: `infrastructure/`.
- `internal/modules/documents/repository/` → `internal/modules/documents/infrastructure/`
  (package `repository` → `infrastructure`); ~39 importer files updated (compiler-verified).
- `internal/modules/templates/repository/*` folded into the **existing**
  `internal/modules/templates/infrastructure/` (no `infrastructure/repository/` nesting; no
  filename collisions); package unified to `infrastructure`; importers updated.
- `internal/modules/documents/approval/repository/` folded into the existing
  `internal/modules/documents/approval/infrastructure/`; package unified to `infrastructure`.

**Deviation flagged during execution (mechanical-fold cycle, resolved without interface redesign):**
folding approval's `repository/` files directly into the *same* `infrastructure/` package as the
pre-existing `postgres_route_admin_idemp_store.go` / `postgres_signoff_idemp_store.go` created a
real Go import cycle: those idemp-store files depend on `approval/application`'s port interfaces
(`application.RouteAdminIdempStore`, `application.SignoffIdempStore` — correct dependency
inversion, infra implements an application-defined port), while `approval/application`'s services
depend on `ApprovalRepository`, an interface **defined in** the persistence package itself. Merging
both file sets into one package closes that into an infra→application→infra cycle. Redefining where
`ApprovalRepository` lives would be an interface redesign (forbidden, HS-2) for a hygiene milestone.
**Resolution, staying mechanical and inside the existing convention:** the two idemp-store files (+
test) moved into a new subpackage `internal/modules/documents/approval/infrastructure/idempotency/`,
mirroring the already-established `internal/modules/documents/approval/infrastructure/signature/`
subpackage pattern (itself proof this convention pre-dates F9.5). Zero interface signatures,
behavior, or SQL changed — only the physical file location and package name of two constructor
functions moved one directory deeper, with their one external caller (`apps/api/cmd/metaldocs-api/main.go`)
updated to the new import path.

### (c) Boundary-guard model realigned to REQ-TOP-1

`scripts/check-module-boundaries.ps1` now encodes the actual rule: a cross-module import may target
only the owning module's **published surface** —
- layer-name allow-list: `domain`, `application`, `api` (any module, first path segment after the
  module name);
- explicit published-package allow-list (interfaces intentionally exposed one level deeper):
  `iam/authz`, `render/fanout`, `render/fanout/dispatchjobs`, `render/resolvers`;
- the `documents`/`documents/approval` nested-family exception described in (a);
- an explicit, ADR-anchored `$debtAllowList` at the top of the script — the **only** sanctioned
  suppression mechanism, currently **empty** (see debt table below).

Everything else — another module's `repository`, `infrastructure`, `delivery`, `http`, `jobs` — is
forbidden. This is **stricter-or-equal** to the old model on every class REQ-TOP-1 names
(repository/SQL/domain-internal imports): the old model already forbade non-`domain` targets in
principle but had no mechanism to distinguish "sanctioned published package" from "genuine
violation," so every real violation the old model could have caught, the new model still catches,
plus it removes the false-negative gap where a persistence-layer import happened to alias through a
name the regex didn't inspect.

**Proofs (all captured in the feature evidence):**
1. RED on the pre-fix baseline tree (old script, old model): 53 lines, mostly the sanctioned
   `iam/authz`/`render/fanout`/`render/resolvers` edges plus cross-module `application` edges.
2. GREEN on the final tree (new script, new model, post-rename).
3. Negative plant: a blank import of `documents/approval/infrastructure` added to
   `internal/modules/jobs/stuck_instance_watchdog/job.go` (a genuinely external module) → script
   RED, correctly naming the planted edge.
4. Revert: `git diff --exit-code` on the planted file after reverting = clean; script GREEN again.

## Violation census (production code, non-test `.go` files, 2026-07-06 sweep)

| Class | Count | Disposition |
|---|---|---|
| Sanctioned published (`iam/authz`) | 42 | Allowed — explicit published-package entry |
| Sanctioned published (`render/fanout`, incl. `dispatchjobs`) | 4 | Allowed — explicit published-package entry |
| Sanctioned published (`render/resolvers`) | 3 | Allowed — explicit published-package entry |
| Cross-module `application`/`domain` (already allowed) | ~127+11 | Allowed — layer allow-list |
| `documents` ↔ `documents/approval` (nested-family, any layer) | 1 flagged edge in the old scan (`documents/delivery/http` → `approval/repository`, now `approval/infrastructure`) | Allowed — nested-exception rule (a) |
| **True cross-module violations (repository/infrastructure/delivery/http/jobs of another module, non-test code)** | **0** | — |

The mini-gate's system-impact analysis (`docs/superpowers/analysis/2026-07-06-f95-approval-structure-system-impact.md`)
named `internal/modules/jobs/stuck_instance_watchdog/job.go` importing `approval/repository` as a
known violation-class import. Re-verified against the actual import list at execution time: the
watchdog imports `metaldocs/internal/modules/documents/approval/application` only — an already-
sanctioned layer, not `approval/repository`/`infrastructure`. No true violation existed there; the
finding is corrected here rather than silently dropped.

## Debt table (true violations not mechanically fixable)

*(empty)* — the production-code sweep found zero true cross-module violations after the rename. No
`$debtAllowList` entries exist in the guard script as of this ADR. If a future change introduces a
genuine repository/infrastructure-layer cross-module import that cannot be resolved by consuming an
existing published port, it must be added to `$debtAllowList` in
`scripts/check-module-boundaries.ps1` with a row added to this table (edge, reason, fix-now-or-
trigger) — never silently allow-listed outside this mechanism.

## Consequences

- **Positive.** The boundary guard now measures the invariant it was always supposed to measure;
  future PRs that add a real cross-module persistence-layer import will fail CI instead of passing
  silently. One persistence-layer directory name (`infrastructure/`) exists per module, matching the
  layout convention documented in `wiki/architecture/backend-target-architecture.md`.
- **Costs.** `documents/approval` remains a nested exception rather than a clean one-module-one-
  directory layout; anyone reading the module tree needs this ADR to understand why. Mitigated by
  the exception being explicit and guarded (not silent).
- **Named triggers.** Approval promotion trigger in (a). Debt-list entries (if any appear later)
  carry their own fix-or-trigger per the mechanism in the debt table section.

## Alternatives considered

- **Promote `documents/approval` to `internal/modules/approval/`** — rejected: would require
  redesigning 100+ edges through published ports with no behavioral gain, splitting one DDD
  aggregate across a module boundary (structure-worship, not structure).
- **Allow-list all 53 baseline-red edges wholesale** — rejected: would have hidden the real
  boundary-guard defect (its allow-model, not the edges) and given a false sense of enforcement.
- **Rewrite the boundary guard in Go instead of PowerShell** — rejected as out of scope (YAGNI); the
  tool works and CI already invokes it, only its allow-model was wrong.
