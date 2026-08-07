# Lane: security-config

## Findings

| ID | Class | Finding | Evidence | Scale |
|----|-------|---------|----------|-------|
| SEC-01 | hazard | `password_reauth` (approval sign-off) has no timing-oracle mitigation for the "user not found" path — bcrypt compare is skipped entirely, unlike `auth.Authenticate` which always spends bcrypt-equivalent time. | `internal/modules/approval/infrastructure/signature/password_reauth.go:99-108` (no-user branch returns before any bcrypt call) vs `internal/modules/auth/application/service.go:266-273,884-900` (dummy-hash / always-compare) | 1 divergent path, 2 auth surfaces |
| SEC-02 | gap | No app-level or DB-tripwire tier-2 authz check in 4 of 15 modules (`audit`, `distribution`, `search`, `security`) — enforcement is tier-1 route→capability only, no defense in depth for these handlers' mutations/reads. | `grep -rln "authz.Require" internal/modules/{audit,distribution,search,security}` → 0 files; DB tripwire CASE arms (`db/baseline/0001_current_schema.sql:249-431`) list no table from these modules | 4/15 modules |
| SEC-03 | hazard | The Postgres role the app containers actually connect as (`${POSTGRES_USER}` = Docker's default superuser) is never verified non-superuser/non-BYPASSRLS anywhere in code; the `NOSUPERUSER NOBYPASSRLS` role that exists (`metaldocs_ci`) is a *test-only* role, not what `metaldocs-api`/`worker`/`jobs` connect as in `deploy/compose/docker-compose.yml`. | `deploy/compose/docker-compose.yml:18-21,PGUSER refs at 227,291,330`; `db/grants/0001_role_grants.sql:63,81-85` (creates `metaldocs_ci`, not the connecting role); no `rolsuper`/`rolbypassrls` assertion anywhere (`grep -rn "BYPASSRLS\|rolsuper" apps/ internal/` → 0 hits) | all 37 RLS-enabled tables, both dev and the only compose file that exists |
| SEC-04 | drift | The audit-events hardening `REVOKE UPDATE, DELETE, TRUNCATE ... FROM metaldocs_app` is inert against SEC-03: PostgreSQL REVOKE/GRANT is a no-op for a superuser role, so the tamper-evidence chain's DB-side backstop only holds if the app's actual connection role is non-superuser, which SEC-03 shows is unverified. | `db/grants/0001_role_grants.sql:101-102`; `db/baseline/0001_current_schema.sql:991` (`FORCE ROW LEVEL SECURITY` — same superuser bypass applies) | 1 control, whole audit chain |
| SEC-05 | gap | `govulncheck` is wired into both CI paths that reference it but defaults to **skipped** in each: the PR-blocking gate (`phase3-hardening-gate.yml`, runs on every PR to `main`) calls the script with its `$SkipGovulncheck = $true` default, and the only path that can run it for real (`release-readiness.yml`) is a manual `workflow_dispatch` whose `skip_govulncheck` input also defaults to `true`. `gosec` by contrast runs unconditionally and is blocking. | `scripts/phase3-hardening-gate.ps1:3` (default true, no override passed at `.github/workflows/phase3-hardening-gate.yml:26`); `.github/workflows/release-readiness.yml:12-14,44`; `scripts/security-baseline.ps1:1-4,58-83` | 2 workflows, 0 of them exercise govulncheck by default |
| SEC-06 | gap | No security response headers (CSP, HSTS, X-Frame-Options, X-Content-Type-Options) are set anywhere in the request path — not in the Go middleware chain, not in either nginx config. | `grep -n "Content-Security-Policy\|Strict-Transport-Security\|X-Frame-Options\|X-Content-Type-Options" deploy/nginx/nginx.conf frontend/apps/web/nginx.conf` → 0 hits; `apps/api/cmd/metaldocs-api/chain.go` chain link list has no headers middleware | whole HTTP surface, 2 nginx configs + Go chain |
| SEC-07 | gap | No concurrent-session-limit policy — a session grant has no cap and no eviction-of-oldest behaviour; login always mints a new session with no upper bound on live sessions per user. | `grep -rn "MaxConcurrentSessions\|SessionLimit" internal/modules/auth` → 0 hits; `internal/modules/auth/application/service.go:900-1000` session creation path has no count check | module-wide |
| SEC-08 | idiom | Tenant KEK (`METALDOCS_TENANT_KEK`) unset is treated as a valid first-class state that silently wires a no-op crypto path rather than failing closed at boot when tenant-crypto-shred is a stated invariant (GMR M7). Confirms and localizes the known release blocker ("backup w/o MinIO+KEK"). | `internal/platform/config/tenant_crypto.go:6-9,28-33` (comment: "Unset is a valid, first-class state") | 1 config path, all tenants until KEK is set |
| SEC-09 | gap | Zip-bomb / malformed-docx protection for uploaded `.docx` content is delegated entirely to the vendored `@eigenpal/docx-editor-core` package (`jszip` dependency of `apps/docx-renderer`); no size/entry-count cap was found in MetalDocs' own code before or after that boundary. | `apps/docx-renderer/package.json:16-25` (`jszip` present); `grep -rn "zip.NewReader\|archive/zip" internal apps --include=*.go` → 0 hits (no Go-side zip handling); `grep -rn "JSZip" apps/docx-renderer/src` → 0 hits (usage is inside the vendored package, not audited here) | unverified depth — flagged, not sized |

## The five heaviest, with detail

**SEC-03 / SEC-04 (superuser connection defeats both RLS and the audit REVOKE).** The only Postgres role wiring that exists for the deployable compose stack is the Docker-default `POSTGRES_USER`, which is a superuser by Postgres convention, and nothing in the codebase asserts otherwise before serving traffic. Because GRANT/REVOKE and `FORCE ROW LEVEL SECURITY` are both no-ops against a superuser connection, this single gap silently defeats two independently-built controls: the 37-table RLS tenant-isolation model and the audit-events hardening that revokes UPDATE/DELETE/TRUNCATE. It blocks trusting either control in the one environment shape (`deploy/compose/docker-compose.yml`) that ships.

**SEC-05 (govulncheck installed but never exercised).** Both CI surfaces that reference `govulncheck` default its skip flag to `true`, so the tool has never actually run against this codebase's dependency graph through CI as configured — it only *looks* wired, the same shape as ME-08's "reporting without enforcement." `gosec` is the real, blocking half of the intended pair; the vulnerability half is theater until someone flips a default.

**SEC-06 (no CSP/HSTS/frame-options anywhere).** For a session-cookie authenticated eQMS surface, the complete absence of response security headers across the Go chain and both nginx configs is a flat gap, not a partial one — there is nowhere in the request path currently capable of setting them.

**SEC-01 (reauth timing divergence).** The primary login path invests specifically in constant-time behaviour (named comment: "closes a TOCTOU window," "OWASP Authentication Cheat Sheet"); the approval-signature reauth path, which exists specifically to re-verify identity for a regulated signing action, skips the bcrypt-equivalent-time step on the not-found branch. The two paths were clearly held to different standards despite protecting comparable actions.

**SEC-02 (tier-2 gap in 4 modules).** `audit`, `distribution`, `search`, `security` have zero `authz.Require` calls and zero DB tripwire coverage. Tier-1 route capability (generated, boot-fatal via `assertSurface`) still gates these routes, so this is not "no enforcement" — it is "one layer instead of two," which is exactly the shape the two-tier model (ADR 0022) was built to avoid relying on.

## What is actually fine

- **Login/session core** (`internal/modules/auth/application/service.go`): bcrypt cost 12, dummy-hash constant-time path for unknown identifiers, per-identity lock closing the lockout TOCTOU window, atomic mutation+audit+session-revocation transactions, 32-byte `crypto/rand` session tokens that are HMAC-signed and stored hashed (never raw) — this is deliberate, well-commented, competent work, matching ME-07's own assessment ("not a criticism of the code... category").
- **Session cookie flags**: `HttpOnly`, `SameSite=Strict`, `Secure` gated on config — correct defaults, no evidence of a `Secure: false` fallback anywhere.
- **CORS config** (`internal/platform/config/cors.go`): fails closed if `AllowCredentials` and `AllowedOrigins` contain `*` together; env-driven with safe method/header defaults.
- **Origin-protection middleware**: default-deny posture on `TrustedProxyCIDRs` ("Empty (default) means no upstream is trusted") — a header-spoofing surface closed by construction rather than by discipline.
- **Config layer**: NOT scattered ad hoc `os.Getenv` — `internal/platform/config/*.go` is a deliberate per-domain typed-config package (cors, jobs, worker, retention, tenant_crypto, trusted_proxy, etc.), 15 files, each with validated parsing and explicit required-vs-optional handling. `METALDOCS_AUTH_SESSION_SECRET` fails closed (returns error, no fallback) if unset or under 32 chars.
- **`.gitignore` + gitleaks**: `.env*` patterns present; `.github/workflows/secret-scan.yml` runs full-git-history gitleaks on every push and PR with `--exit-code 1` (genuinely blocking, not advisory).
- **`gosec`**: runs unconditionally on every PR to `main` via `phase3-hardening-gate.yml`, no skip flag — the working half of the security-tooling pair (contrast SEC-05).
- **Audit hash chain**: computed by an `IMMUTABLE` SQL function (`metaldocs.audit_event_row_hash`, sha256 over `prev_hash` + all row fields with a `\x1f` separator) at INSERT time, not app-computed — correct choice of trust boundary, undermined only by SEC-03/SEC-04, not by its own design.
- **Supply-chain CVE gate**: Grype scan with `fail-build: true`, `severity-cutoff: high` on push to `main` and tags (`.github/workflows/supply-chain.yml`) — real blocking gate, distinct from the govulncheck gap.

## Unverified / needs judgment

- Does the `metaldocs_app` role referenced in `db/grants/0001_role_grants.sql` ever actually get created and used as the app's *connecting* role in any environment, or is it vestigial relative to `POSTGRES_USER`? Only the compose file was checked; no separate prod Terraform/Helm/Ansible was found in this pass — if one exists elsewhere it could resolve SEC-03.
- Depth of zip-bomb/entry-count/decompression-ratio protection inside the vendored `@eigenpal/docx-editor-core` (SEC-09) — out of this pass's budget; the vendor package itself was not opened.
- Whether Chromium-based HTML→PDF rendering (`internal/platform/render/gotenberg/client.go`) is exposed to any user-controlled external URL (SSRF via `<img src>`/`<link>` in rendered content) — the client itself has no egress restriction, but whether upstream content generation ever includes user-supplied absolute URLs was not traced end-to-end.
- Whether any environment outside the one `docker-compose.yml` in this repo (e.g., a customer's actual prod Postgres) uses a distinct, restricted role — asked as a question because the repo's own compose file is the only artifact that exists to check.
- Dependency freshness (majors-behind count) — `go list -m -u all` returned no output in this environment (likely network-gated); not independently sized. `go.mod` shows `golang.org/x/crypto v0.51.0`, `github.com/jackc/pgx/v5 v5.9.2`, `github.com/riverqueue/river v0.37.1`, `github.com/lib/pq v1.12.3` (a second, presumably legacy, Postgres driver alongside pgx — worth a question for the driver-duplication angle, not sized here).

## Commands run

```
grep -rn "GenerateFromPassword|bcrypt.Cost|DefaultCost" internal/modules/auth internal/modules/approval --include=*.go
grep -rn "http.Cookie{" internal/modules/auth --include=*.go
grep -rn "SessionTTL|sessionTTL|MaxAge|Expires" internal/modules/auth --include=*.go
grep -rln "authz.Require" internal/modules --include=*.go | grep -v _test
grep -rn "TG_TABLE_NAME = '" db/baseline/0001_current_schema.sql
grep -rn "SET LOCAL metaldocs.tenant_id|set_config.*tenant_id|SeedTxTenant" internal/platform --include=*.go
grep -c "FORCE ROW LEVEL SECURITY" db/baseline/0001_current_schema.sql
grep -c "ENABLE ROW LEVEL SECURITY" db/baseline/0001_current_schema.sql
grep -rn "metaldocs_app|BYPASSRLS|SUPERUSER" db/**/*.sql scripts/*.ps1
sed -n '1,50p' deploy/compose/docker-compose.yml
grep -rn "MaxBytesReader|LimitReader" internal apps --include=*.go
grep -rn "os.Getenv" --include=*.go . | grep -v _test | grep -v vendor | wc -l
cat .github/workflows/secret-scan.yml
cat .github/workflows/supply-chain.yml
cat scripts/phase3-hardening-gate.ps1
cat scripts/security-baseline.ps1
sed -n '1,50p' .github/workflows/release-readiness.yml
grep -n "audit_event_row_hash" -A 25 db/baseline/0001_current_schema.sql
grep -n "Content-Security-Policy|Strict-Transport-Security|X-Frame-Options|X-Content-Type-Options" deploy/nginx/nginx.conf frontend/apps/web/nginx.conf
sed -n '1090,1145p' internal/modules/auth/application/service.go
```
