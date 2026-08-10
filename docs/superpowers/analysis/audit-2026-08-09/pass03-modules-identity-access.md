# Pass 03 — Module Architecture Maps: Identity & Access (iam, auth, security, taxonomy, tokens)

date: 2026-08-09
baseline: main@418070bf
status: reproduced-current

Scope: `internal/modules/{iam,auth,security,taxonomy,tokens}`. Cross-module edge
counts and reciprocal-edge/SCC facts are taken as given from
`docs/superpowers/analysis/audit-2026-08-09/pass02-go-dependency-graph.md` and
`module-edge-evidence.txt` (not recomputed). All other claims below are
independently verified with file:line evidence in this pass.

---

## 0. Two-tier PDP map (cross-cutting, read before the per-module sections)

Three enforcement tiers exist, split across two modules and one platform package:

| Tier | Mechanism | Owner | Evidence |
|---|---|---|---|
| 1 — route→capability | HTTP middleware, checks `CapabilityService` before the handler runs | **iam** (`delivery/http`, `application`) | `internal/modules/iam/delivery/http/middleware.go:58` `NewMiddleware(caps *iamapp.CapabilityService, roleProvider iamdomain.RoleProvider, enabled bool)`; `:80` `Wrap` |
| 1 data source | `CanDo`/`IsSystemAdmin`/`CapsByUserID` query `iam_user_roles` UNION `iam_group_members⋈iam_groups⋈iam_group_roles`, joined to `role_capabilities` | **iam** (`application`) | `internal/modules/iam/application/capability_service.go:48-81` (`CanDo`), `:101-113` (`IsSystemAdmin`), `:132-152` (`CapsByUserID`) |
| 2 — capability×area in-tx | `authz.Require(ctx, tx, capability, areaCode)`, must run in RW tx | **iam** (`authz` package — sibling of `domain`/`application`, not nested under either) | `internal/modules/iam/authz/authz.go:100` |
| 2 data source | `SELECT EXISTS(... FROM role_capabilities rc JOIN user_process_areas upa ON upa.role=rc.role AND upa.tenant_id=$4 AND upa.user_id=$3 AND upa.effective_from<=now() AND upa.effective_to IS NULL WHERE rc.capability=$1 AND ($2='tenant' OR upa.area_code=$2))` | **iam** (`authz`) | `internal/modules/iam/authz/authz.go:144-156` |
| 2 system-admin bypass | separate query over `iam_user_roles` + group chain | **iam** (`authz`) | `internal/modules/iam/authz/authz.go:25-40` `SystemAdminExistsSQL` |
| 3 — DB tripwire | `enforce_capability_asserted()` trigger, Go-side arm registry (24 `Table:` arms) | **platform/tripwire** (deliberately outside all 15 modules) | `internal/platform/tripwire/arms.go:1-18` package doc, one-way import of `iam/domain` only |

**#89/A8 finding confirmed precisely**: tier-1 and tier-2 read *different* grant
tables — tier-1 unions `iam_user_roles` with the group-membership chain
(`iam_group_members⋈iam_groups⋈iam_group_roles`); tier-2 reads
`role_capabilities JOIN user_process_areas` directly and has **no group path at
all** (a user granted a role only via group membership, never a direct
`iam_user_roles` row or a `user_process_areas` row, can pass tier-1 but fail
tier-2 for area-scoped capabilities, or vice versa for tenant-scope). Per the
task instruction this is not re-litigated here (ADR-0092 governs); this section
exists only to pin the exact evidence with line numbers, superseding any
looser prose version.

**Tripwire coverage gap (tier-3)**: `arms.go` has zero `Table:` arms for
`auth_identities`, `auth_sessions` (auth-owned), `tenant_keys` (security-owned),
or `token_dictionary_entries` (tokens-owned) — confirmed absent by grep of all
`Table:` literals in `internal/platform/tripwire/arms.go`. Every mutation to
these four tables is enforced by tiers 1–2 only, with no last-line DB trigger.
This is consistent with tripwire's stated scope (governance/lifecycle tables:
approval, documents, controlled_documents, templates, iam grant tables,
tenants) but means identity/session/crypto/dictionary tables sit one
application-layer bug away from an unauthorized write, with no fail-safe.

**Platform P-edge verdicts** (task-requested classification):

- **`platform/tripwire → iam/domain`** (1 edge) — **legitimate**. Package doc at
  `internal/platform/tripwire/arms.go:1-18` explicitly states "one-way import:
  this package imports iam/domain for the capability consts and
  IsValidCapability; iam/domain and its dependents never import this package.
  No cycle." This is vocabulary borrowing (capability enum), not business-logic
  coupling — tripwire owns its own arm registry and trigger semantics
  independently of iam's services. Verdict: **not a misplacement**.
- **`platform/authn → iam/domain`** (`internal/platform/authn/context.go:7`) —
  **legitimate**. `UserIDFromContext` (`context.go:21-27`) is a thin,
  presence-aware wrapper around `iamdomain.UserIDFromContext`, adding only a
  trim+empty-check. No business rule beyond "empty after trim = absent."
  Verdict: **not a misplacement**.
- **`platform/authn → auth/application`** (`internal/platform/authn/config.go:12`)
  — **questionable, borderline misplaced**. `config.go` builds `authapp.Config`
  (auth's own business configuration type) via `LoadRuntimeConfig()`, and
  `DevRoleMap()`/`parseDevRoleMap()` (same file) hardcode dev-only role-mapping
  business logic using `iamdomain.Role`/`iamdomain.RoleSystemAdmin`
  (`internal/platform/authn/config.go:13`). This is env/config assembly for a
  specific module's business type, not a neutral platform concern — it reads
  as auth's composition-root wiring that was placed in `platform/` instead of
  `apps/api/cmd/metaldocs-api/` or `auth/module.go`. It does not create a
  module→module cycle (nothing imports it back), so it is not on the #93/A4
  SCC-breaking critical path, but it inverts the "platform is neutral, modules
  own their business config" boundary. **Recommend**: fold `DevRoleMap`/
  `LoadRuntimeConfig` into `auth/module.go` or the composition root; leave only
  the thin context accessor in `platform/authn`.

---

## 1. Module: iam

### 1. Current responsibility (code truth)
Identity/tenancy administration + the capability-based authorization decision
point (tiers 1–2). Owns: users, tenant onboarding/export/erasure lifecycle,
group/role/capability grant tables, presence, area-membership admin, KPI/usage
read models. Not auth (login/session) and not the DB tripwire (tier-3, lives
in `platform/tripwire`).

### 2. Owned domain concepts/aggregates/VOs
`iam/domain`: `Role` (alias of `platform/iamtypes.Role`), `Capability`,
`ProcessArea` membership (`UserProcessArea`), `Group`, `RoleProvider`,
`TenantLifecycleEnqueuer`, tenant onboarding/export/erasure state machine
(`tenant_lifecycle.go`), `membership_tx.go` (in-tx role/group mutation
port). Own error sentinels: `ErrUserNotFound`, `ErrUserInactive`,
`ErrNoRolesAssigned`, `ErrInvalidRole`, etc.

### 3. Owned DB tables (writes)
`iam_users`, `iam_user_roles`, `iam_groups`, `iam_group_members`,
`iam_group_roles`, `role_capabilities`, `user_process_areas`, `tenants`,
`tenant_lifecycle_jobs` (confirmed against `platform/tripwire/arms.go` arm
`Table:` list, which lists these as iam-owned governed tables, cross-checked
against `iam/infrastructure/postgres/*` repository file names).

### 4. Foreign tables read/written
- **`auth_identities`/`auth_sessions`** — read indirectly only via
  `auth/application`'s own service calls (people_service.go composes
  `*authapp.Service`), never raw SQL from iam. Clean — no S-seam.

### 5. Public application surface (key exported services)
`iam/application.CapabilityService` (`CanDo`, `IsSystemAdmin`,
`CapsByUserID`), `PeopleService` (people_service.go), `OnboardTenantService`
(onboard_tenant_service.go). `iam/authz.Require`, `SeedTxIdentity`,
`WithBackgroundBypass` (published cross-module PDP plumbing, consumed by
every other module's infra for GUC seeding — legitimate shared
infrastructure, distinct from the anti-pattern below).

### 6. Public domain surface consumed by others
`iamdomain.Capability`, `iamdomain.Role` (re-export of `iamtypes.Role`),
`iamdomain.RoleProvider`, `iamdomain.UserIDFromContext`,
`iamdomain.WithAuthContext`, `iamdomain.Err*` sentinels (`ErrInvalidRole`,
`ErrNoRolesAssigned`, `ErrUserNotFound`, ...) — all consumed directly (Go
import) by `auth`, `security`, `taxonomy`, `tokens`, and `platform/authn`,
`platform/tripwire`.

### 7/8. Inbound / outbound module dependencies (from `module-edge-evidence.txt`, not recomputed)
- Outbound: `iam/application → auth/application, auth/domain`;
  `iam/delivery/http → auth/domain`; `iam/infrastructure/postgres →
  taxonomy/domain`; `iam/application → security/domain`.
- Inbound: `auth → iam` (5 edges), `security → iam` (1 edge), `taxonomy → iam`
  (2 edges), `tokens → iam` (2 edges), `platform:authn → iam`, `platform:tripwire
  → iam`, `platform:bootstrap → iam`.
- iam is the dominant hub: fan-in 13 / fan-out 4 (pass02 §fan table). It sits
  inside the size-9 module SCC.

### 9. Seam classification (cross-module collaborations touching iam)

| Consumer | Producer | Coupling | Who should own contract | Adapter present? | Reciprocal? | Verdict |
|---|---|---|---|---|---|---|
| `security/infrastructure/postgres` | `iam/domain` (`UserDisplayNameReader`, `TenantUserReader`, `AdminRoleMemberReader`, `MfaUserReader`) | **T — producer-owned type inversion** | security (consumer) should declare its own reader ports | No adapter — iam's postgres repos wired directly, structurally satisfying iam's own interfaces | No | Anti-pattern (see §10) |
| `iam/infrastructure/postgres` (`ProcessAreaCatalog`) | `taxonomy/domain.AreaCatalogReader` | **T — producer-owned type inversion**, but documented/intentional (ADR-0039 D3(b)) | taxonomy (producer) — accepted as deliberate shared catalog-read port | No adapter (direct interface satisfaction) | Partially — `taxonomy → iam` also exists for `authz.SeedTxIdentity` (P plumbing, not the same seam) | Anti-pattern-shaped but ADR-ratified; still worth an adapter for consistency (see §10) |
| `auth/{application,domain,delivery/http,infrastructure/*}` | `iam/domain` (`Capability`, `Role`, `RoleProvider`, `Err*`) | **G — Go import + E — sentinel coupling**, bidirectional | Neither cleanly — historical bidirectional coupling (ARC-06); `iamtypes` extraction is the precedent | Partial (`iamtypes.Role` extracted; `Capability`/`RoleProvider`/sentinels not) | **Yes** — `auth ↔ iam` is one of pass02's 7 reciprocal module edges | Confirmed live defect, not resolved; see §10/§11 |
| `taxonomy/infrastructure/authz_guc.go` | `iam/authz.SeedTxIdentity` | **P — platform-style plumbing**, not a business coupling | n/a — designated shared PDP call | n/a | n/a | **Legitimate**, not an anti-pattern (every module's infra calls this identically) |
| `tokens/application` | `iam/authz`, `iam/domain` | Same PDP-plumbing pattern as taxonomy | n/a | n/a | No | Legitimate |

### 10. Consumer-owned vs producer-owned ports — iam's role
iam is the **producer** in the two confirmed anti-pattern seams (security→iam,
taxonomy→iam-as-consumer via `ProcessAreaCatalog`) and simultaneously the
**over-imported dependency** in the auth↔iam bidirectional seam. See the
consolidated inventory in §"Producer-owned port inventory" below.

### 11. Transaction participation
`iam/authz.Require` **must** run inside a caller-supplied RW `db.Tx`
(`authz.go:100`; enforced repo-wide by the `authz-require-rw-tx` api-lint rule
per CLAUDE.md/memory). 35 TxRunner/BeginTx/DoReadOnly hits across the module
(grep count). iam owns its own `TxRunner` usage for people/tenant lifecycle
writes; `authz` package itself takes an injected `db.Tx`/`db.DB`, not a
runner — correct consumer-owned-tx shape (the caller's transaction boundary
governs, not authz's).

### 12. Events/outbox/async jobs
Only async-job owner among the 5 modules. `iam/jobs/tenant_lifecycle_enqueuer.go`
implements `iamdomain.TenantLifecycleEnqueuer` via River's `InsertTx`
same-tx pattern (transactional-outbox-equivalent), queue `"temporal"`,
mirroring `documents/approval/jobs/lifecycle_event_enqueuer.go` exactly
(per the file's own comment). No consumer-side subscriber lives in this
module (executed by `metaldocs-jobs` per CLAUDE.md system facts).

### 13. HTTP routes owned
`internal/modules/iam/delivery/http/router.go` (357 lines): ~20+ routes —
sessions, admin-overview, area-memberships, capabilities/roles/role-caps,
kpi/usage, presence, users (×10, including a dead `CreateManagedUser` →
501), tenant onboarding/export/erase.

### 14. Test shape
39 test files, 8 integration-tagged. Heaviest test surface of the 5 modules,
consistent with hub role (fan-in 13).

### 15. Findings → issues
- **#93/A4**: iam is a party to 2 of the module's confirmed type-inversion
  seams (as producer, both directions) plus the auth↔iam bidirectional Go-import
  edge — the single largest seam concentration in this pass.
- **#89/A8**: both tier-1 (`capability_service.go:48-81`) and tier-2
  (`authz.go:144-156`) data sources live entirely inside iam; the dual-source
  split is an **intra-module** design fact, not a cross-module seam — worth
  flagging precisely because #93/A4-style "move the port" fixes do not touch
  #89/A8 at all; they are orthogonal defects that happen to share a module.
- **#92/A5**: `authz` package's `db.Tx`/`db.DB` in port signatures (`authz.go`
  `Require(ctx, tx db.Tx, ...)`) is consumer-owned-tx-correct, not part of the
  "9 of 15 domain packages leak db types" stat — but `iam/domain/membership_tx.go`
  and `tenant_lifecycle.go` DO import `platform/db` into `iam/domain` proper,
  contributing to that stat.

---

## 2. Module: auth

### 1. Current responsibility (code truth)
Login/logout/session lifecycle, password policy, current-user resolution.
Owns `auth_identities`, `auth_sessions`. Distinct from iam (identity
*administration*) — auth is identity *authentication*.

### 2. Owned domain concepts/aggregates/VOs
`auth/domain`: `ManagedUser`, session model (`SessionListItem` — see §leakage
below), login context, `CreateUserInput`/`UpdateUserParams`. Own sentinels:
`ErrUserAlreadyExists`, `ErrIdentityNotFound`, `ErrPasswordPolicy`,
`ErrSessionNotFound`.

### 3. Owned DB tables (writes)
`auth_identities`, `auth_sessions`.

### 4. Foreign tables read/written
None found via raw SQL — auth consumes iam exclusively through Go types/ports
(`iamdomain.RoleProvider`, `iamdomain.RoleAdminRepository`,
`iamdomain.LoginContextPort`), not foreign SQL. Clean on the S axis.

### 5. Public application surface
`auth/application.Service` (`Login`, `Logout`, `GetCurrentUser`,
`ChangePassword`, password hashing — `HashPassword` used as iam's
`PasswordHashFunc` default per `onboard_tenant_service.go:132`).

### 6. Public domain surface consumed by others
`authdomain.ManagedUser`, `CreateUserInput`, `UpdateUserParams`,
`CurrentUserFromContext`, `Err*` sentinels — consumed directly by iam
(`people_service.go:212` `auth *authapp.Service`) and by
`iam/delivery/http/middleware.go:9,192` (`authdomain.CurrentUserFromContext`).

### 7/8. Inbound / outbound dependencies
Outbound: `auth → iam` (5 edges: application, domain, delivery/http,
infrastructure/memory, infrastructure/postgres, all → `iam/domain`).
Inbound: `iam → auth` (3 edges: `iam/application → auth/application,
auth/domain`; `iam/delivery/http → auth/domain`); `platform:authn → auth/application`;
`platform:bootstrap → auth/{domain,infra/postgres}`.

### 9. Seam classification
See the shared row in iam's §9 table (`auth ↔ iam` — the reciprocal edge).
Additional auth-side detail: `errors.Is(err, iamdomain.ErrNoRolesAssigned)`,
`errors.Is(err, iamdomain.ErrInvalidRole)` confirmed in auth application code
— **E-seam** (foreign sentinel coupling), 2 confirmed sites, layered on top
of the G-seam (direct type imports). Classification: **G+E, bidirectional,
neither side owns the contract** — this is the worst-shaped seam in the pass
because it is reciprocal (unlike the one-directional security/taxonomy
inversions, no single "flip the arrow" fix resolves it; it needs a shared
extraction like the `iamtypes.Role` precedent, generalized to
`Capability`/`RoleProvider`/session-existence checks).

### 10. Producer-owned vs consumer-owned ports
auth is not itself typed by a foreign-declared reader interface (no T-seam
as consumer), but it is the **historical co-author** of the iam bidirectional
coupling: pre-`iamtypes` extraction, iam imported `auth/domain.ManagedUser`
et al. while auth imported `iam/domain.Role` et al. — the `iamtypes.Role` fix
(`internal/platform/iamtypes/role.go:1-16`, doc comment self-documents ARC-06)
resolved exactly one of the shared types; `Capability`, `RoleProvider`,
`ErrInvalidRole`/`ErrNoRolesAssigned`, and `ManagedUser`/`CreateUserInput`
remain un-extracted and the edge is still live and reciprocal per pass02.

### 11. Transaction participation
14 TxRunner/BeginTx/DoReadOnly hits. `auth/domain/session_admin.go` imports
raw `"database/sql"` directly (not `platform/db`) and types
`SessionListItem.CreatedAt/LastSeenAt/ExpiresAt` as `sql.NullTime`
(`session_admin.go:3,34-36`) — **domain-layer DB-driver leakage more severe
than a port-signature `db.Tx` param**, since it is baked into the value type
itself and would require a mapping layer to ever swap drivers/represent the
type over a wire boundary cleanly. Direct #92/A5 evidence.

### 12. Events/outbox/async jobs
None.

### 13. HTTP routes owned
`internal/modules/auth/delivery/http/handler.go` — 4 routes: Login, Logout,
GetCurrentUser, ChangePassword.

### 14. Test shape
13 test files, 1 integration-tagged — noticeably thin integration coverage
relative to iam given the reciprocal coupling surface.

### 15. Findings → issues
- **#93/A4**: auth↔iam bidirectional G/E-seam is the clearest "7 reciprocal
  module edges" instance directly inside this pass's 5-module scope (3 of
  pass02's 7 total reciprocal edges touch these 5 modules: `iam<->taxonomy`,
  `iam<->security`... — but only `auth<->iam` is truly bidirectional Go-import;
  the iam<->taxonomy and iam<->security "reciprocal" edges are actually
  one T-seam each direction, not raw bidirectional import, see §"reciprocal
  edge disambiguation" below).
- **#92/A5**: `session_admin.go`'s `sql.NullTime` domain fields are a concrete,
  citable instance of the "9 of 15 domain packages" / DB-type-leakage stat.

---

## 3. Module: security

### 1. Current responsibility (code truth)
Security signal surface: MFA coverage, lockouts, failed-login counters,
off-hours admin-action reporting, tenant-key crypto (crypto-shred). Read-heavy
aggregator module over other modules' data plus its own `tenant_keys` table.

### 2. Owned domain concepts/aggregates/VOs
`security/domain`: `TenantCrypto` port (`tenant_crypto.go`), `ErrKeyNotFound`,
`ErrKeyDestroyed`. No independent "security event" aggregate — the module is
primarily a read/report layer plus the crypto-shred write path.

### 3. Owned DB tables (writes)
`tenant_keys` only (`security/infrastructure/postgres/tenant_key_repository.go`
— `InsertIfAbsentTx:40`, `WrappedDEKTx:89`, `DestroyTx:114`, all taking a
caller-injected `db.Tx`, no owned TxRunner). Confirmed **not** covered by the
generic `TenantDataPort` erasure registry — `internal/composition/tenantdata/registry/*.go:43-47`
documents this as a deliberate exclusion: `tenant_keys` is destroyed by
security's own `TenantCrypto.DestroyTenantKeyTx` crypto-shred step, "not by a
TenantDataPort (a different feature/task)." This also explains why security
has no `tenant_data_port.go` file, unlike iam/auth/taxonomy/tokens — a
documented exception, not a gap.

### 4. Foreign tables read/written — the primary finding of this section
Raw SQL against tables **owned by other modules, with no port at all**
(worse than a type-coupling inversion — direct S-seam):
- `metaldocs.auth_identities` — `security/infrastructure/postgres/repository.go:128,194,248`
- `metaldocs.auth_sessions` — `repository.go:277,284`
- `metaldocs.audit_events` — `repository.go:358` (in `ListOffHoursAdminActions`)

Contrast: security's reads of **iam-owned** data (`iam_users`,
`iam_user_roles`/group tables via display-name/tenant-user/admin-role
lookups) go through 4 injected reader ports (§10) — so security is
internally inconsistent: iam data is behind ports, auth and audit data is
raw SQL with zero abstraction.

### 5. Public application surface
`security/infrastructure/postgres.Repository`: `resolveNames`, `MfaCoverage`,
`ListLockouts`, `CountRecentFailedLoginsByUser`, `CountRecentLockouts`,
`ListOffHoursAdminActions`. No distinct `application` package layer was found
with business logic beyond the infra repository — thin service, mostly a
query aggregator (worth flagging as a layering question for #93/A4, since it
means the T/S-seams below sit directly in the infrastructure layer with no
intervening application-owned port to absorb them).

### 6. Public domain surface consumed by others
None found — nothing in this pass imports `security/domain` except iam
(`iam/application → security/domain`, 1 edge per module-edge-evidence.txt;
purpose not fully traced in this pass, flagged for follow-up).

### 7/8. Inbound / outbound dependencies
Outbound: `security/infrastructure/postgres → iam/domain` (the 4-reader-port
edge, §9/§10). Inbound: `iam/application → security/domain` (1 edge).

### 9. Seam classification

| Consumer | Producer | Coupling | Who should own contract | Adapter? | Reciprocal? | Verdict |
|---|---|---|---|---|---|---|
| `security/infrastructure/postgres.Repository` | `iam/domain` (`UserDisplayNameReader`, `TenantUserReader`, `AdminRoleMemberReader`, `MfaUserReader`) | **T — producer-owned type inversion** | security (consumer) | No — iam's own postgres repos (`iampg.NewUserDisplayNameRepository` etc.) wired directly at `apps/api/cmd/metaldocs-api/main.go:877-890`, satisfying iam's interfaces structurally | No | **Named anti-pattern #1**, confirmed exactly as task described |
| `security/infrastructure/postgres.Repository` | `metaldocs.auth_identities`, `auth_sessions` tables (auth-owned) | **S — foreign SQL coupling, no port** | security should declare its own consumer-owned reader port; auth or a composition adapter should implement it | No — none | No | Newly-identified sibling, more severe than the T-seam (bypasses Go's type system entirely, not caught by any interface-satisfaction check) |
| `security/infrastructure/postgres.Repository` | `metaldocs.audit_events` (audit-owned) | **S — foreign SQL coupling, no port** | security should declare a reader port; audit should own the query surface | No | No | Same class as above, third foreign table |

### 10. Consumer-owned vs producer-owned ports — security's role
security is the **consumer** in every seam above and owns **zero** of its own
declared reader ports for foreign data — it is typed entirely by producer
(iam)-declared interfaces for iam data, and has no interface abstraction at
all for auth/audit data. This is the purest instance of the named
anti-pattern: not one violation but a structural pattern across the whole
module — security was built by injecting whatever the producer module already
had.

### 11. Transaction participation
0 TxRunner/BeginTx/DoReadOnly hits for the report-surface repository (pure
off-tx pool reads via `*sql.DB`). The crypto repository (`tenant_key_repository.go`)
uses caller-injected `db.Tx` correctly (consumer-owned-tx shape) — this part
of the module is architecturally sound; the reporting part is not.

### 12. Events/outbox/async jobs
None.

### 13. HTTP routes owned
`internal/modules/security/delivery/http/handler.go` — 3 routes:
GetMfaCoverage, ListLockouts, ListSecuritySignals.

### 14. Test shape
5 test files, 2 integration-tagged — thinnest test surface of the 5 modules,
notable given the module has the most foreign-coupling surface area.

### 15. Findings → issues
- **#93/A4**: security is the cleanest, most complete instance of the named
  anti-pattern — both the T-seam (task-specified) and 2 new S-seams
  (auth_identities/auth_sessions/audit_events raw SQL) live in the same
  ~390-line file (`security/infrastructure/postgres/repository.go`),
  confirming the file is a coupling hotspot in its own right, not just a
  single-line defect.
- Sentinel coupling (E-seam) is **clean** — security only uses its own
  `ErrKeyNotFound`/`ErrKeyDestroyed`; the coupling is 100% concentrated in
  types (T) and raw SQL (S), none in error handling.

---

## 4. Module: taxonomy

### 1. Current responsibility (code truth)
Flat, code-keyed classification catalog: `DocumentFamily`, `DocumentProfile`,
`ProcessArea`. Per `internal/modules/taxonomy/module.go:1-5` doc comment:
"the flat, code-keyed classification catalog... that controlled-documents and
documents bind to."

### 2. Owned domain concepts/aggregates/VOs
`DocumentFamily`, `DocumentProfile`, `ProcessArea`/`AreaCatalog`. Own
sentinels only (`domain.Err*`) — confirmed clean of foreign `errors.Is`
coupling by grep.

### 3. Owned DB tables (writes)
Area, family, profile tables (via `infrastructure.NewProfileRepository`,
`NewAreaRepository`, `NewFamilyRepository` — `module.go:43-44,57`).

### 4. Foreign tables read/written
None via raw SQL found in this pass — taxonomy's foreign dependency is
entirely through published Go ports (`approvaldomain.RouteReadinessReader`,
`iam/authz.SeedTxIdentity`), not table access.

### 5. Public application surface
`application.ProfileService`, `AreaService`, `FamilyService` (`module.go:59-62`).
`ProfileService.WithRouteReadinessReader` — consumes approval's
**consumer-owned-by-taxonomy** port `approvaldomain.RouteReadinessReader`
(taxonomy declares the dependency need, approval implements — this is the
*correct* direction, worth noting as a second positive exemplar alongside
DictionaryValueReader: `module.go:32-35,52-54` panics loudly if not wired,
"taxonomy reports the badge honestly or errors; it never defaults it to
false").

### 6. Public domain surface consumed by others — the second named anti-pattern
`taxonomy/domain/area_catalog_reader_port.go` declares `AreaCatalogReader`
(producer-owned; methods `AreaName`, `AreaExists`, structural `db.DB`
executor param so it works both in-tx and off-tx) plus a
`NoopAreaCatalogReader` null object. **Consumed by iam**:
`internal/modules/iam/infrastructure/postgres/area_catalog_reader.go:7,24`
— `ProcessAreaCatalog` struct field `areaCatalog taxonomydomain.AreaCatalogReader`,
delegating `AreaCodeExists` (`:40-42`) straight through. This is the exact
inversion the task named: **iam's infrastructure is typed by an interface
taxonomy's domain declared.**

### 7/8. Inbound / outbound dependencies
Outbound: `taxonomy/infrastructure → iam/authz, iam/domain` (2 edges — the
`authz_guc.go` PDP-plumbing call, legitimate per iam §9).
Inbound: `iam/infrastructure/postgres → taxonomy/domain` (1 edge — the
anti-pattern above).

### 9. Seam classification
See iam §9 table row 2. Summary: **T-seam, producer(taxonomy)-owned,
consumed by iam, one-directional (not reciprocal in the Go-import sense —
though pass02 counts `iam<->taxonomy` as a "reciprocal module edge" because
edges exist in both directions at the module-adjacency level: taxonomy→iam
via the authz plumbing, iam→taxonomy via this port. These are two
*different* seams, not one bidirectional coupling of the same concern — see
disambiguation note below.**

**Reciprocal-edge disambiguation** (clarifying pass02's module-level
adjacency for this pass's 5 modules): pass02 counts `iam<->taxonomy` and
`iam<->security` as "reciprocal" because *some* edge exists each direction at
the module-collapsed level. Concretely:
- `iam<->taxonomy`: taxonomy→iam is PDP plumbing (legitimate, P-shaped);
  iam→taxonomy is the `AreaCatalogReader` T-seam (anti-pattern-shaped, ADR-ratified).
  These are unrelated concerns that happen to create a bidirectional
  module-adjacency edge — **not** a symmetric coupling problem like auth↔iam.
- `iam<->security`: iam→security is `iam/application → security/domain` (1
  edge, purpose not fully traced this pass); security→iam is the 4-reader-port
  T-seam. Also two unrelated concerns, not a symmetric problem.
- `auth<->iam` is the **only** truly symmetric case: both directions are the
  same kind of coupling (direct domain-type + sentinel imports for
  overlapping business concepts — roles, capabilities, user identity), which
  is why it alone needs a shared-kernel extraction fix rather than a
  one-directional port flip.

### 10. Producer-owned vs consumer-owned ports — taxonomy's role
taxonomy is the **producer** of the `AreaCatalogReader` inversion (iam is
consumer) but the **correct consumer** of `approvaldomain.RouteReadinessReader`
(approval is producer, taxonomy declares the need via `WithRouteReadinessReader`
and fails loudly without it) — taxonomy is a mixed case, one anti-pattern
seam and one healthy seam.

### 11. Transaction participation
35 TxRunner/BeginTx/DoReadOnly hits (tied with iam's raw count, though
taxonomy is a much smaller module — high tx-density). `taxonomy/domain/port.go`
imports `platform/db` into the domain layer (contributes to #92/A5's "9 of 15"
stat, consistent with `AreaCatalogReader`'s `db.DB` executor param design).

### 12. Events/outbox/async jobs
None.

### 13. HTTP routes owned
`internal/modules/taxonomy/delivery/http/routes_{areas,families,generated,profiles}.go`
— 16 generated routes (profiles ×6, areas ×5, families ×5).

### 14. Test shape
24 test files, 5 integration-tagged.

### 15. Findings → issues
- **#93/A4**: confirms the task's named anti-pattern #2 exactly, with the
  caveat that it is ADR-0039 D3(b)-ratified as intentional — this changes the
  remediation framing from "undocumented defect" to "documented local
  maximum that should still carry an adapter for consistency with the
  DictionaryValueReader exemplar," per CLAUDE.md's Global-Maximum-Not-Local
  rule (a ratified local maximum is only non-defective if labelled
  transitional with a named deletion milestone — ADR-0039 D3(b) should be
  checked for whether it makes that labelling explicit; not fully confirmed
  in this pass, flagged as a follow-up read of the ADR text itself).

---

## 5. Module: tokens

### 1. Current responsibility (code truth)
Tenant-scoped dictionary of named token values (`token_dictionary_entries`)
that documents/controlleddocuments pin at creation time. Smallest, most
self-contained module of the 5.

### 2. Owned domain concepts/aggregates/VOs
`Entry`. Own sentinels only: `ErrNotFound`, `ErrReservedName`,
`ErrImmutableName` — confirmed clean of foreign `errors.Is` coupling.

### 3. Owned DB tables (writes)
`token_dictionary_entries` only. `domain/port.go:9-11` doc comment is
explicit: "The repo touches ONLY token_dictionary_entries."

### 4. Foreign tables read/written
None.

### 5. Public application surface
`tokens/application.Service` implementing both `domain.Repository`-backed
CRUD and the published `DictionaryReader` port (`GetByName`, `List`).

### 6. Public domain surface consumed by others — the positive exemplar's other half
`domain/port.go:21-27`: `DictionaryReader` is explicitly documented as
"the provider port this module PUBLISHES for SP-2... documents/controlleddocuments
creation path reads dictionary values through it off-tx." Consumed via
`apps/api/cmd/metaldocs-api/dictionary_reader_adapter.go:1-28` —
`dictionaryValueReaderAdapter` wraps `tokensdomain.DictionaryReader` and
adapts to **documents' own consumer-owned** `DictionaryValueReader` interface
(declared in `internal/modules/documents/application/service.go:144`),
mapping `tokensdomain.ErrNotFound` → `found=false` at the adapter boundary so
documents never imports tokens' error type directly. Comment: "keeps the
documents module free of any tokens import (SP-2 §11, invariant #6). Lives
at the composition root." **This is the textbook-correct shape**: consumer
(documents) declares the interface it needs, producer (tokens) publishes its
own port, an adapter at the composition root bridges them, and even the
sentinel mapping happens at the boundary instead of leaking `errors.Is`
coupling into documents.

### 7/8. Inbound / outbound dependencies
Outbound: `tokens/application → iam/authz, iam/domain` (2 edges — same
PDP-plumbing pattern as taxonomy, legitimate). Inbound: none from the other
4 modules in this pass (documents consumes tokens, but documents is outside
this pass's scope) — tokens sits **outside** pass02's size-9 SCC (fan-in 0
within that SCC), confirming it as the cleanest-boundaried module examined.

### 9. Seam classification
`tokens → documents` (via the exemplar adapter): **C — consumer-owned
healthy contract.** No anti-pattern seams found for tokens in either
direction.

### 10. Producer-owned vs consumer-owned ports — tokens' role
tokens is the **model producer** for this whole pass: it publishes its own
interface (`DictionaryReader`) rather than being typed by a consumer's
interface, and the consumer (documents) independently declares its own
interface (`DictionaryValueReader`) rather than importing tokens' type
directly. Zero producer-owned-inversion seams involving tokens were found.

### 11. Transaction participation
10 TxRunner/BeginTx/DoReadOnly hits. `domain/port.go:13-18` — all
`Repository` methods take caller-supplied `db.Tx` ("the service owns the tx
boundary via platform/db.TxRunner") — correct consumer-owned-tx shape,
explicitly documented in the port's own doc comment.

### 12. Events/outbox/async jobs
None.

### 13. HTTP routes owned
`internal/modules/tokens/delivery/http/handler.go` — 5 routes: ListTokens,
CreateToken, GetToken, UpdateToken, DeleteToken.

### 14. Test shape
6 test files, 2 integration-tagged.

### 15. Findings → issues
- **#93/A4**: tokens is the **negative control** for this pass — it proves
  the codebase already knows how to build this correctly (matches the
  DictionaryValueReader exemplar named in the task) and that the anti-pattern
  in iam/security/auth is a module-specific defect, not an unavoidable
  consequence of the modular-monolith architecture itself.

---

## Producer-owned port inventory (task §10, exact list)

Confirmed instances of a module's **infrastructure** layer being typed by an
interface **declared in a different module's domain package** (the specific
anti-pattern class named in the task), plus the raw-SQL sibling that is
structurally the same defect without even the type-system's help:

1. **`internal/modules/security/infrastructure/postgres/repository.go:23-33`**
   — `Repository` struct fields `displayNames iamdomain.UserDisplayNameReader`,
   `members iamdomain.TenantUserReader`, `adminRoles iamdomain.AdminRoleMemberReader`,
   `mfaUsers iamdomain.MfaUserReader`. Import: `iamdomain "metaldocs/internal/modules/iam/domain"`
   (`repository.go:16`). Wired directly with iam's own postgres
   implementations, no adapter, at `apps/api/cmd/metaldocs-api/main.go:877-890`.
2. **`internal/modules/iam/infrastructure/postgres/area_catalog_reader.go:24`**
   — `ProcessAreaCatalog` struct field `areaCatalog taxonomydomain.AreaCatalogReader`.
   Import: `taxonomydomain "metaldocs/internal/modules/taxonomy/domain"`
   (`:7`). ADR-0039 D3(b)-documented as intentional. Wired at
   `apps/api/cmd/metaldocs-api/main.go:1017`.
3. **Sibling (S-seam, no port at all, worse than 1–2)**:
   `security/infrastructure/postgres/repository.go:128,194,248` (raw SQL vs
   `metaldocs.auth_identities`), `:277,284` (vs `metaldocs.auth_sessions`),
   `:358` (vs `metaldocs.audit_events`). No interface, no adapter, direct
   table names in security's SQL strings.
4. **Bidirectional G/E-seam (distinct class — mutual Go-import + sentinel
   coupling, not a single-direction port inversion)**: `auth ↔ iam`, both
   directions, confirmed reciprocal by pass02. auth-side: `auth/application`,
   `auth/domain`, `auth/delivery/http`, `auth/infrastructure/{memory,postgres}`
   all import `iam/domain` directly (5 edges per module-edge-evidence.txt) plus
   2 confirmed `errors.Is(err, iamdomain.Err*)` sites. iam-side:
   `iam/application → auth/application, auth/domain`; `iam/delivery/http →
   auth/domain` (3 edges), plus iam's own foreign `errors.Is(err, authdomain.Err*)`
   sites. Partial precedent for the fix already exists:
   `internal/platform/iamtypes/role.go` extracted `Role` alone; `Capability`,
   `RoleProvider`, and the cross-referenced error sentinels remain
   un-extracted.

**Not an instance of the anti-pattern (verified and excluded)**:
`taxonomy/infrastructure/authz_guc.go`'s call to `iam/authz.SeedTxIdentity`,
and `tokens/application`'s calls to `iam/authz`/`iam/domain` — both are calls
to iam's designated, published cross-cutting PDP plumbing, not a producer's
domain-declared reader interface backing another module's own repository.

---

## Positive exemplar comparison

| | Documents/tokens (exemplar) | Security/iam/taxonomy (anti-pattern) |
|---|---|---|
| Interface declared by | Consumer (`documents/application.DictionaryValueReader`) | Producer (`iam/domain.*Reader`, `taxonomy/domain.AreaCatalogReader`) |
| Adapter at composition root? | Yes — `apps/api/cmd/metaldocs-api/dictionary_reader_adapter.go`, explicit, documented, maps sentinels | No — producer's own postgres struct passed directly (main.go:877-890, :1017) |
| Foreign sentinel handling | Mapped to a boolean (`found=false`) at the adapter boundary — consumer never sees producer's error type | None needed for iam→taxonomy (Noop pattern instead); auth↔iam leaks sentinels directly across the boundary via `errors.Is` |
| Consumer module import-clean of producer? | Yes — comment states this explicitly ("keeps the documents module free of any tokens import") | No — security imports `iam/domain` directly; iam imports `taxonomy/domain` directly; auth and iam import each other's `domain` packages directly |

## Findings-to-issues summary table

| Issue | Evidence from this pass |
|---|---|
| **#93/A4** (module seams) | 2 named anti-patterns confirmed exact (security→iam ×4 ports; iam→taxonomy ×1 port); 1 new S-seam sibling found (security→auth_identities/auth_sessions/audit_events raw SQL, 6 sites); 1 bidirectional G/E-seam confirmed live (auth↔iam, partial iamtypes precedent for the fix); taxonomy's `RouteReadinessReader` and tokens' `DictionaryReader`/`DictionaryValueReader` confirmed as correct consumer-owned counter-examples |
| **#89/A8** (authz dual-source) | Tier-1 (`capability_service.go:48-81`, `iam_user_roles` UNION group chain) vs tier-2 (`authz.go:144-156`, `role_capabilities JOIN user_process_areas`, no group path) pinned with exact SQL and line numbers; both live inside iam (intra-module, not cross-module) |
| **#92/A5** (persistence/tx mechanics) | `auth/domain/session_admin.go:3,34-36` `sql.NullTime` in domain value type (worse than port-signature leakage); `iam/domain`, `taxonomy/domain` import `platform/db` directly; `authz`, `tokens/domain`, `security`'s crypto repo all show the *correct* caller-injected-`db.Tx` shape — mixed picture, not uniformly bad |
| **#90/A3** (contract/runtime convergence) | Out of this pass's primary evidence — no direct route-contract findings surfaced; taxonomy's 16 generated routes and iam's dead `CreateManagedUser` (→501) are the only contract-adjacent observations, flagged but not deeply investigated here |

---

*End of pass03. Module boundaries verified by direct file:line reads and
targeted greps against the worktree at main@418070bf; cross-module edge
counts sourced from pass02/module-edge-evidence.txt per the task's
evidence-reuse instruction.*
