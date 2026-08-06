//go:build integration

package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	authapp "metaldocs/internal/modules/auth/application"
	authdelivery "metaldocs/internal/modules/auth/delivery/http"
	authdomain "metaldocs/internal/modules/auth/domain"
	authpg "metaldocs/internal/modules/auth/infrastructure/postgres"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/tests/integration/testdb"

	"golang.org/x/crypto/bcrypt"
)

// TestSurfaceConformance is positive evidence against the ONE authority: for
// every declared operation it proves the table the server enforces is the
// table the spec declares. It is NOT a differential against routeRules — a
// peer table is not what the server should obey (§10.1).
//
// THE DISCRIMINATOR IS THE SENTINEL, NOT THE RESPONSE. Both PDP tiers deny
// with 403 by deliberate, test-pinned design, and the problem CODE does not
// separate them either: routes_memberships.go:312-318 maps a tier-2
// ErrCapDenied to the tier-1 code, and documents/handler.go:1300-1325 emits
// the tier-1 code from inside the handler. Those codes are decided by fifteen
// modules' error mappers — a discriminator this suite does not own.
//
// Tier-1 denies IN THE MIDDLEWARE, before next.ServeHTTP. So the suite mounts
// the real chain over a sentinel terminal handler and observes invocation:
//   negative = denied AND sentinel never invoked  => the middleware decided
//   positive = sentinel invoked                   => the middleware admitted
//
// This depends on one property it does not establish: that the pattern it
// registers is the pattern the real publisher mounts. That is §5 check 3, at
// boot, named here rather than assumed.
func TestSurfaceConformance(t *testing.T) {
	db := testdb.New(t)

	// One fixture serves every negative case. A truly role-less principal
	// (NewUser without WithRole) cannot serve this purpose: buildCurrentUser
	// (auth/application/service.go:983) calls the real RoleProvider
	// unconditionally while building the login response, and it hard-fails
	// (ErrUserNotFound / ErrNoRolesAssigned both propagate) for a user holding
	// no role anywhere — so a role-less user cannot complete a real login at
	// all, not even to reach a session-required (no-specific-capability)
	// operation. "viewer" is the least-privileged role the iam_user_roles
	// CHECK constraint (db/baseline: system_admin, approver, author, editor,
	// viewer) actually allows assigning, so it is the minimal REAL fixture.
	tenant := testdb.NewTenant(t, db)
	minimalUser := testdb.NewUser(t, db, testdb.WithTenantID(tenant.ID), testdb.WithRole("viewer"))

	for pattern, rule := range httpSurface {
		pattern, rule := pattern, rule
		t.Run(pattern, func(t *testing.T) {
			method, path := splitPattern(t, pattern)
			concrete := fillPathParams(path) // {id} -> a syntactically valid value

			var reached bool
			sentinel := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reached = true
				w.WriteHeader(http.StatusOK)
			})
			srv := chainOverSentinel(t, db, pattern, sentinel)

			switch rule.visibility {
			case iamdelivery.VisibilityPublic:
				reached = false
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, httptest.NewRequest(method, concrete, nil))
				if !reached {
					t.Fatalf("public operation was not admitted unauthenticated (status %d)", rr.Code)
				}

			case iamdelivery.VisibilitySessionRequired:
				reached = false
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, httptest.NewRequest(method, concrete, nil))
				if reached {
					t.Fatal("session-required operation admitted an unauthenticated caller")
				}
				reached = false
				rr = httptest.NewRecorder()
				srv.ServeHTTP(rr, withSession(t, httptest.NewRequest(method, concrete, nil), minimalUser))
				if !reached {
					t.Fatalf("session-required operation refused a valid session (status %d)", rr.Code)
				}

			case iamdelivery.VisibilityPermissionGuarded:
				// grantingRoles = the assignable roles whose role_capabilities
				// include this operation's capability. Its size decides which
				// sub-cases are runnable, per Ruling 6:
				//   0            -> unreachable by any real user; that is a
				//                   separate finding (TestNoDeclaredOperationIsUnreachable),
				//                   NOT this loop's job. Skip this operation.
				//   all 5        -> the tier-1 negative case is vacuous BY
				//                   CONSTRUCTION (ADR 0022: tier-1 is coarse
				//                   route->capability; a capability every
				//                   assignable role holds is, at tier-1,
				//                   equivalent to "any authenticated
				//                   principal"). Assert that equality as a
				//                   CHECKED CLAIM in place of the negative
				//                   fixture — it self-heals: the day any grant
				//                   narrows, this branch stops matching and
				//                   the real negative case below runs again.
				//   proper subset -> run both sub-cases; the negative fixture
				//                   is a real assignable role that does NOT
				//                   grant the capability.
				grantingRoles := rolesGranting(t, db, string(rule.capability))
				switch {
				case len(grantingRoles) == 0:
					t.Skipf("capability %q is granted by NO assignable role — tracked as a finding by TestNoDeclaredOperationIsUnreachable, not asserted per-operation here", rule.capability)
				case len(grantingRoles) == len(assignableRoleNames):
					if !equalRoleSets(grantingRoles, assignableRoleNames) {
						t.Fatalf("capability %q: granted by %d assignable roles but the sets differ: got %v, want %v", rule.capability, len(grantingRoles), grantingRoles, assignableRoleNames)
					}
					// Checked claim stands in for the negative case: every
					// assignable role holds this capability, so there is no
					// real principal to deny. Positive sub-case still proves
					// admission below.
				default:
					nonGranting := firstRoleNotIn(assignableRoleNames, grantingRoles)
					negUser := testdb.NewUser(t, db, testdb.WithTenantID(tenant.ID), testdb.WithRole(nonGranting))
					reached = false
					rr := httptest.NewRecorder()
					srv.ServeHTTP(rr, withSession(t, httptest.NewRequest(method, concrete, nil), negUser))
					if reached {
						t.Fatalf("capability %q was NOT required: the sentinel ran for a principal (role %q) holding it not", rule.capability, nonGranting)
					}
				}

				// POSITIVE: any assignable role granting the declared
				// capability — grantingRoles is non-empty in every remaining
				// branch.
				role := grantingRoles[0]
				holder := testdb.NewUser(t, db, testdb.WithTenantID(tenant.ID), testdb.WithRole(role))
				reached = false
				rr := httptest.NewRecorder()
				srv.ServeHTTP(rr, withSession(t, httptest.NewRequest(method, concrete, nil), holder))
				if !reached {
					t.Fatalf("capability %q was required but a holder (role %q) was refused (status %d)", rule.capability, role, rr.Code)
				}

			default:
				t.Fatalf("unexpected visibility %v", rule.visibility)
			}
		})
	}
}

// assignableRolesFilter mirrors the iam_user_roles CHECK constraint
// (db/baseline/0001_current_schema.sql chk_iam_user_roles_role_code): the only
// role_code values a real row can hold. role_capabilities also seeds
// area_admin/qms_admin/signer, but no user can actually be assigned those
// roles today — a role-granting lookup that ignores this constraint would
// pick an unassignable role and fail testdb.NewUser's insert.
const assignableRolesFilter = `role IN ('system_admin', 'approver', 'author', 'editor', 'viewer')`

// assignableRoleNames is assignableRolesFilter's role set as a sorted Go
// slice, so the conformance loop can compare a capability's granting-role set
// against "all of them" without a second round trip.
var assignableRoleNames = []string{"approver", "author", "editor", "system_admin", "viewer"}

// rolesGranting returns every ASSIGNABLE role whose role_capabilities set
// contains capability, sorted. An empty result means the capability is
// unreachable by any real user — that is TestNoDeclaredOperationIsUnreachable's
// finding to report, not this function's to fail on.
func rolesGranting(t *testing.T, db *sql.DB, capability string) []string {
	t.Helper()
	rows, err := db.Query(
		`SELECT role FROM metaldocs.role_capabilities WHERE capability = $1 AND `+assignableRolesFilter+` ORDER BY role`,
		capability,
	)
	if err != nil {
		t.Fatalf("rolesGranting(%q): %v", capability, err)
	}
	defer rows.Close()
	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			t.Fatalf("rolesGranting(%q) scan: %v", capability, err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rolesGranting(%q) rows: %v", capability, err)
	}
	return roles
}

// pickSubsetCarrierRoute finds a GET, path-parameterized, permission-guarded
// operation whose granting-role set is a PROPER NON-EMPTY SUBSET of the 5
// assignable roles — i.e. an operation that can express BOTH tier-1
// sub-cases (a real fixture that holds the capability, a real fixture that
// doesn't). Computed, not hardcoded: a route gated on a capability every
// assignable role holds (a "floor" capability, e.g. document.view) cannot
// express the negative half at all, so the carrier must be chosen for that
// property and not for convenience — a hardcoded route string could silently
// become a floor capability under a later seed change with nobody noticing.
// Iteration is over a sorted copy of httpSurface's keys so the choice is
// deterministic across runs. An empty result is itself a finding: it means
// the resolver's declared surface has no operation left able to express a
// real tier-1 negative case at all.
func pickSubsetCarrierRoute(t *testing.T, db *sql.DB) (pattern, path string) {
	t.Helper()
	patterns := make([]string, 0, len(httpSurface))
	for p := range httpSurface {
		patterns = append(patterns, p)
	}
	sort.Strings(patterns)

	for _, p := range patterns {
		rule := httpSurface[p]
		if rule.visibility != iamdelivery.VisibilityPermissionGuarded {
			continue
		}
		method, rulePath := splitPattern(t, p)
		if method != http.MethodGet || !strings.Contains(rulePath, "{") {
			continue
		}
		granting := rolesGranting(t, db, string(rule.capability))
		if len(granting) > 0 && len(granting) < len(assignableRoleNames) {
			return p, rulePath
		}
	}
	t.Fatal("pickSubsetCarrierRoute: no GET, path-parameterized, permission-guarded operation has a capability granted by a proper non-empty subset of assignable roles — TestHEADInheritsGETCapability and TestParameterContentCannotSteerThePolicy have no valid carrier to run against")
	return "", ""
}

// fillFirstPathParam substitutes value for path's FIRST {param} only, and
// fills any remaining {param}s with testdb.AnyUUID via fillPathParams — used
// to craft a suspicious parameter VALUE on the carrier's own param slot while
// leaving any other params syntactically valid.
func fillFirstPathParam(path, value string) string {
	open := strings.IndexByte(path, '{')
	if open < 0 {
		return path
	}
	closeRel := strings.IndexByte(path[open:], '}')
	if closeRel < 0 {
		return path
	}
	rest := path[open+closeRel+1:]
	return path[:open] + value + fillPathParams(rest)
}

// unassignableRolesGranting returns every role OUTSIDE assignableRolesFilter
// (area_admin/qms_admin/signer, in practice) whose role_capabilities set
// contains capability, sorted — the diagnostic column for
// TestNoDeclaredOperationIsUnreachable's table: it names who WOULD be able to
// invoke the operation if the role catalog were widened.
func unassignableRolesGranting(t *testing.T, db *sql.DB, capability string) []string {
	t.Helper()
	rows, err := db.Query(
		`SELECT role FROM metaldocs.role_capabilities WHERE capability = $1 AND NOT (`+assignableRolesFilter+`) ORDER BY role`,
		capability,
	)
	if err != nil {
		t.Fatalf("unassignableRolesGranting(%q): %v", capability, err)
	}
	defer rows.Close()
	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			t.Fatalf("unassignableRolesGranting(%q) scan: %v", capability, err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("unassignableRolesGranting(%q) rows: %v", capability, err)
	}
	return roles
}

// equalRoleSets compares two sorted role slices for equality.
func equalRoleSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// firstRoleNotIn returns the first role in all that does not appear in
// granting. Both must be sorted; all must have at least one role granting
// does not (i.e. granting must be a proper subset of all) — callers only
// reach this for that exact case.
func firstRoleNotIn(all, granting []string) string {
	grantSet := make(map[string]bool, len(granting))
	for _, r := range granting {
		grantSet[r] = true
	}
	for _, r := range all {
		if !grantSet[r] {
			return r
		}
	}
	panic("firstRoleNotIn: granting is not a proper subset of all — caller invariant violated")
}

// TestNoDeclaredOperationIsUnreachable is the finding, extracted: a
// permission-guarded operation whose capability is granted by no ASSIGNABLE
// role can be invoked by no user in any environment. Pre-existing:
// metaldocs.role_capabilities seeds area_admin, qms_admin and signer, which
// chk_iam_user_roles_role_code forbids assigning. This test is red until the
// role-catalog decision lands — it is not this suite's job to fix the seed
// data, only to make the gap loud, precise, and impossible to bury inside 10
// scattered per-operation subtest failures.
func TestNoDeclaredOperationIsUnreachable(t *testing.T) {
	db := testdb.New(t)

	type finding struct {
		unassignableGranters []string
		operations           []string
	}
	findings := map[string]*finding{}

	for pattern, rule := range httpSurface {
		if rule.visibility != iamdelivery.VisibilityPermissionGuarded {
			continue
		}
		capability := string(rule.capability)
		if f, ok := findings[capability]; ok {
			f.operations = append(f.operations, pattern)
			continue
		}
		if len(rolesGranting(t, db, capability)) > 0 {
			continue
		}
		findings[capability] = &finding{
			unassignableGranters: unassignableRolesGranting(t, db, capability),
			operations:           []string{pattern},
		}
	}

	if len(findings) == 0 {
		return
	}

	capabilities := make([]string, 0, len(findings))
	for capability := range findings {
		capabilities = append(capabilities, capability)
	}
	sort.Strings(capabilities)

	var b strings.Builder
	b.WriteString("permission-guarded operation(s) gated on a capability NO assignable role grants — unreachable by any real user in any environment (pre-existing: metaldocs.role_capabilities seeds area_admin/qms_admin/signer, which chk_iam_user_roles_role_code forbids assigning to a real user):\n")
	fmt.Fprintf(&b, "%-24s | %-30s | %s\n", "capability", "granting unassignable roles", "operations")
	for _, capability := range capabilities {
		f := findings[capability]
		sort.Strings(f.operations)
		granters := "(none — no role at all, assignable or not, grants this)"
		if len(f.unassignableGranters) > 0 {
			granters = strings.Join(f.unassignableGranters, ",")
		}
		fmt.Fprintf(&b, "%-24s | %-30s | %s\n", capability, granters, strings.Join(f.operations, ", "))
	}
	t.Errorf("%s", b.String())
}

// splitPattern turns "GET /api/v1/documents/{id}" into ("GET", "/api/v1/documents/{id}").
func splitPattern(t *testing.T, pattern string) (method, path string) {
	t.Helper()
	i := strings.IndexByte(pattern, ' ')
	if i <= 0 {
		t.Fatalf("pattern %q is not method-qualified; every generated key must be", pattern)
	}
	return pattern[:i], pattern[i+1:]
}

// fillPathParams replaces every {param} with a syntactically valid value. The
// value never needs to EXIST: tier-1 runs before the handler, and the sentinel
// is what runs after, so nothing ever looks the id up.
func fillPathParams(path string) string {
	for {
		open := strings.IndexByte(path, '{')
		if open < 0 {
			return path
		}
		close := strings.IndexByte(path[open:], '}')
		if close < 0 {
			return path
		}
		path = path[:open] + testdb.AnyUUID + path[open+close+1:]
	}
}

// conformanceDB is the leased test database the currently-running top-level
// conformance/edge test bound its fixtures against. withSession needs a *sql.DB
// to mint a real session for a testdb.User, but Task 17b/18 fix its signature
// as (t, r, u) — no db parameter — since other tests reference it by that
// exact shape. testdb.Open/New leases a FRESH physical database on every call
// (see lease_pool.go), so withSession cannot simply call testdb.New(t) itself:
// a second lease would be a different, fixture-less database. Each top-level
// test instead binds this package-level var immediately after its own
// testdb.New(t) call, before running any subtest; none of Task 17a's tests use
// t.Parallel(), so there is exactly one active db per process at a time.
var conformanceDB *sql.DB

// testSessionSecret is the fixed HMAC key shared between the authService
// withSession uses to mint a session and the authService chainOverSentinel
// wires into the real chain to resolve it — both read/write the SAME
// physical database (conformanceDB), so two distinct *authapp.Service
// instances agreeing on this one secret is what makes a session minted by
// one resolvable by the other.
const testSessionSecret = "surface-conformance-suite-session-secret-32+"

// testLoginPassword is the fixed plaintext password withSession provisions
// every fixture identity with, so Authenticate can mint a real session "the
// same way the login handler mints one" without this suite owning a second,
// parallel session-construction path.
const testLoginPassword = "Surface-Conformance-Suite-Pw-1!"

// noopLoginContextPort satisfies iamdomain.LoginContextPort (auth.NewService
// requires a non-nil implementation). RecordLoginContext is best-effort in
// production (errors are swallowed) — this suite has no governance-metadata
// assertions to make, so a no-op is the honest minimum, not a shortcut.
type noopLoginContextPort struct{}

func (noopLoginContextPort) RecordLoginContext(context.Context, string, string, string, string, string) error {
	return nil
}

// newTestAuthService builds a real, Postgres-backed auth.Service against db —
// the same construction main.go performs (authpg.NewRepository wrapping
// iampg.NewUserTenantRepository), minus the module this suite's session
// minting never touches (role admin — no test path ever calls
// CreateUserWithInput/ReplaceUserRoles). roleProvider IS real and required: a
// nil one panics inside buildCurrentUser, which Authenticate calls
// unconditionally to build the login response's CurrentUser projection.
func newTestAuthService(t *testing.T, db *sql.DB) *authapp.Service {
	t.Helper()
	repo := authpg.NewRepository(db, iampg.NewUserTenantRepository(db))
	svc, err := authapp.NewService(repo, iampg.NewRoleProvider(db), nil, noopLoginContextPort{}, testAuthConfig())
	if err != nil {
		t.Fatalf("new test auth service: %v", err)
	}
	return svc
}

func testAuthConfig() authapp.Config {
	return authapp.Config{
		SessionCookieName:      "metaldocs_session",
		SessionTTL:             time.Hour,
		SessionSecret:          authapp.Secret(testSessionSecret),
		PasswordMinLength:      1,
		LoginMaxFailedAttempts: 1000,
		LoginLockDuration:      time.Minute,
		AllowDevTenantFallback: true,
	}
}

// withSession attaches a real authenticated session for u, minted the same way
// the login handler mints one: an auth_identity is provisioned for u's
// existing iam_users row (testdb.NewUser seeds iam_users/iam_user_roles only —
// auth and iam meet at, and stay decoupled by, this identity/role boundary,
// same as production), then authapp.Service.Authenticate mints a real session
// row and signed token exactly as the login handler does, and the resulting
// cookie is attached to r.
//
// Idempotent per user: called multiple times for the SAME u (this suite reuses
// one "minimalUser" fixture across every operation's negative/session-required
// case), the second+ call finds the identity already provisioned
// (ErrUserAlreadyExists) and skips straight to minting a fresh session — never
// re-fails the whole call.
func withSession(t *testing.T, r *http.Request, u testdb.User) *http.Request {
	t.Helper()
	if conformanceDB == nil {
		t.Fatal("withSession: conformanceDB is nil — the calling test must set it right after testdb.New(t)")
	}
	ctx := context.Background()

	repo := authpg.NewRepository(conformanceDB, iampg.NewUserTenantRepository(conformanceDB))
	hash, err := bcrypt.GenerateFromPassword([]byte(testLoginPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("withSession: hash test password: %v", err)
	}
	err = repo.CreateUser(ctx, authdomain.CreateUserParams{
		UserID:       u.ID,
		Username:     u.ID,
		DisplayName:  u.DisplayName,
		PasswordHash: authdomain.PasswordHash(hash),
		PasswordAlgo: "bcrypt",
		IsActive:     true,
	})
	if err != nil && err != authdomain.ErrUserAlreadyExists {
		t.Fatalf("withSession: provision identity for %s: %v", u.ID, err)
	}

	svc := newTestAuthService(t, conformanceDB)
	// No X-Tenant-ID claim: every fixture this suite mints (minimalUser and
	// every capability holder) carries exactly one iam_user_roles row in its
	// own tenant — a role-less session cannot complete login at all (see
	// TestSurfaceConformance's comment on minimalUser) — so GetUserTenants
	// always returns exactly [u.TenantID] and resolveLoginTenant's
	// single-tenant branch picks it without needing an explicit claim.
	loginReq := httptest.NewRequest(http.MethodPost, authdelivery.PathLogin, nil)
	authSession, err := svc.Authenticate(ctx, u.ID, testLoginPassword, loginReq)
	if err != nil {
		t.Fatalf("withSession: authenticate %s: %v", u.ID, err)
	}

	r.AddCookie(svc.SessionCookie(authSession.RawToken, authSession.ExpiresAt))
	return r
}

// chainOverSentinel builds the REAL middleware chain (chain.go:25) with the
// real newPermissionResolver, over a mux carrying ONLY `pattern` -> sentinel.
// Using the real chain and the real resolver is what makes the suite evidence
// about production rather than about a hand-built stand-in.
//
// Only authn and iam_authz are real, non-nil links — mirroring
// health_delta_test.go's established minimal-real-chain pattern, extended
// here to also cover tier-2's CapabilityService, since VisibilityPermissionGuarded
// cases need a real capability check to discriminate. Every other link
// (panic_recovery, otel, http_obs, cors, origin_protection,
// pre_auth_login_rate_limit, presence_bump, rate_limit, method_not_allowed) is
// intentionally nil: buildChain skips nil links, and this suite's assertions
// are about sentinel invocation, never about those links' side effects.
func chainOverSentinel(t *testing.T, db *sql.DB, pattern string, sentinel http.Handler) http.Handler {
	t.Helper()
	conformanceDB = db

	mux := http.NewServeMux()
	mux.Handle(pattern, sentinel)

	permResolver := newPermissionResolver(mux)

	authService := newTestAuthService(t, db)
	authMiddleware := authdelivery.NewMiddleware(authService, testAuthConfig(), true).
		WithPublicPathChecker(newPublicPathChecker(permResolver)).
		WithPasswordChangeAllowedChecker(newPasswordChangeAllowedChecker(mux))

	capabilityService := iamapp.NewCapabilityService(db)
	roleProvider := iampg.NewRoleProvider(db)
	iamMiddleware := iamdelivery.NewMiddleware(capabilityService, roleProvider, true).
		WithPermissionResolver(permResolver)

	links := apiChain(
		nil, // panic_recovery
		nil, // otel
		nil, // http_obs
		nil, // cors
		nil, // origin_protection
		nil, // pre_auth_login_rate_limit
		authMiddleware.Wrap,
		iamMiddleware.Wrap,
		nil, // presence_bump
		nil, // rate_limit
		nil, // method_not_allowed
	)
	return buildChain(mux, links)
}
