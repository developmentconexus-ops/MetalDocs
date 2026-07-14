## 1. File tree

```text
internal/modules/auth/
+-- application/
¦   +-- service.go - Defines auth application `Config` and `Service` behavior.
+-- delivery/
¦   +-- http/
¦       +-- handler.go - Defines HTTP auth handler, route registration, and response/error mapping.
¦       +-- middleware.go - Defines auth middleware, public-route checks, and context injection.
¦       +-- middleware_test.go - Tests middleware public/private route and cookie behavior.
+-- domain/
¦   +-- context.go - Defines context helpers for current authenticated user.
¦   +-- errors.go - Defines auth domain error variables.
¦   +-- model.go - Defines auth domain entities and DTO structs.
¦   +-- port.go - Defines auth repository interface contract.
+-- infrastructure/
    +-- memory/
    ¦   +-- repository.go - In-memory auth repository implementation with role-admin helpers.
    +-- postgres/
        +-- repository.go - PostgreSQL auth repository implementation.
        +-- repository_test.go - Tests PostgreSQL repository behavior.
```

## 2. Public surface

| File:line | Kind | Name | Signature | Doc-line |
|---|---|---|---|---|
| `internal/modules/auth/application/service.go:26` | type | `Config` | `type Config` | `(undocumented)` |
| `internal/modules/auth/application/service.go:45` | type | `Service` | `type Service` | `(undocumented)` |
| `internal/modules/auth/application/service.go:52` | func | `NewService` | `func NewService(repo authdomain.Repository, roleProvider iamdomain.RoleProvider, roleAdmin iamdomain.RoleAdminRepository, cfg Config) *Service` | `(undocumented)` |
| `internal/modules/auth/application/service.go:56` | func | `BootstrapLocalAdmin` | `func (s *Service) BootstrapLocalAdmin(ctx context.Context) error` | `(undocumented)` |
| `internal/modules/auth/application/service.go:100` | func | `Authenticate` | `func (s *Service) Authenticate(ctx context.Context, identifier, password string, r *http.Request) (authdomain.AuthenticatedSession, error)` | `(undocumented)` |
| `internal/modules/auth/application/service.go:166` | func | `ResolveSession` | `func (s *Service) ResolveSession(ctx context.Context, rawToken, tenantID string) (authdomain.CurrentUser, error)` | `(undocumented)` |
| `internal/modules/auth/application/service.go:193` | func | `Logout` | `func (s *Service) Logout(ctx context.Context, rawToken string) error` | `(undocumented)` |
| `internal/modules/auth/application/service.go:205` | func | `ChangePassword` | `func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error` | `(undocumented)` |
| `internal/modules/auth/application/service.go:209` | func | `ChangePasswordForUser` | `func (s *Service) ChangePasswordForUser(ctx context.Context, currentUser authdomain.CurrentUser, currentPassword, newPassword string) error` | `(undocumented)` |
| `internal/modules/auth/application/service.go:253` | func | `ListUsers` | `func (s *Service) ListUsers(ctx context.Context, tenantID string) ([]authdomain.ManagedUser, error)` | `(undocumented)` |
| `internal/modules/auth/application/service.go:272` | func | `ListOnlineUsers` | `func (s *Service) ListOnlineUsers(ctx context.Context, activeSince time.Time) ([]authdomain.OnlineUser, error)` | `(undocumented)` |
| `internal/modules/auth/application/service.go:279` | func | `CreateUser` | `func (s *Service) CreateUser(ctx context.Context, userID, username, email, displayName, password, tenantID string, roles []iamdomain.Role, createdBy string) error` | `(undocumented)` |
| `internal/modules/auth/application/service.go:328` | func | `UpdateUser` | `func (s *Service) UpdateUser(ctx context.Context, params authdomain.UpdateUserParams, newPassword string) error` | `(undocumented)` |
| `internal/modules/auth/application/service.go:343` | func | `AdminResetPassword` | `func (s *Service) AdminResetPassword(ctx context.Context, userID, newPassword string) error` | `(undocumented)` |
| `internal/modules/auth/application/service.go:364` | func | `UnlockUser` | `func (s *Service) UnlockUser(ctx context.Context, userID string) error` | `(undocumented)` |
| `internal/modules/auth/application/service.go:371` | func | `SessionCookie` | `func (s *Service) SessionCookie(rawToken string, expiresAt time.Time) *http.Cookie` | `(undocumented)` |
| `internal/modules/auth/application/service.go:384` | func | `SessionCookieName` | `func (s *Service) SessionCookieName() string` | `(undocumented)` |
| `internal/modules/auth/application/service.go:388` | func | `ExpiredSessionCookie` | `func (s *Service) ExpiredSessionCookie() *http.Cookie` | `(undocumented)` |
| `internal/modules/auth/application/service.go:401` | func | `CurrentUser` | `func (s *Service) CurrentUser(ctx context.Context, userID, tenantID string) (authdomain.CurrentUser, error)` | `(undocumented)` |
| `internal/modules/auth/application/service.go:405` | func | `buildCurrentUser` | `func (s *Service) buildCurrentUser(ctx context.Context, userID, tenantID string) (authdomain.CurrentUser, error)` | `(undocumented)` |
| `internal/modules/auth/application/service.go:424` | func | `validatePassword` | `func (s *Service) validatePassword(password string) error` | `(undocumented)` |
| `internal/modules/auth/application/service.go:431` | func | `hashPassword` | `func (s *Service) hashPassword(password string) (string, error)` | `(undocumented)` |
| `internal/modules/auth/application/service.go:439` | func | `newSessionToken` | `func (s *Service) newSessionToken() (string, string, error)` | `(undocumented)` |
| `internal/modules/auth/application/service.go:450` | func | `tokenHashFromCookieValue` | `func (s *Service) tokenHashFromCookieValue(raw string) (string, error)` | `(undocumented)` |
| `internal/modules/auth/application/service.go:461` | func | `signToken` | `func (s *Service) signToken(token string) string` | `(undocumented)` |
| `internal/modules/auth/delivery/http/handler.go:17` | type | `Handler` | `type Handler` | `(undocumented)` |
| `internal/modules/auth/delivery/http/handler.go:31` | func | `NewHandler` | `func NewHandler(service *authapp.Service) *Handler` | `(undocumented)` |
| `internal/modules/auth/delivery/http/handler.go:35` | func | `RegisterRoutes` | `func (h *Handler) RegisterRoutes(mux *http.ServeMux)` | `(undocumented)` |
| `internal/modules/auth/delivery/http/handler.go:42` | func | `handleLogin` | `func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request)` | `(undocumented)` |
| `internal/modules/auth/delivery/http/handler.go:68` | func | `handleLogout` | `func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request)` | `(undocumented)` |
| `internal/modules/auth/delivery/http/handler.go:80` | func | `handleMe` | `func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request)` | `(undocumented)` |
| `internal/modules/auth/delivery/http/handler.go:94` | func | `handleChangePassword` | `func (h *Handler) handleChangePassword(w http.ResponseWriter, r *http.Request)` | `(undocumented)` |
| `internal/modules/auth/delivery/http/handler.go:131` | func | `writeAuthError` | `func (h *Handler) writeAuthError(w http.ResponseWriter, err error, traceID string)` | `(undocumented)` |
| `internal/modules/auth/delivery/http/middleware.go:19` | type | `PublicPathChecker` | `type PublicPathChecker` | `PublicPathChecker returns true if the given method+path requires no session` |
| `internal/modules/auth/delivery/http/middleware.go:21` | type | `Middleware` | `type Middleware` | `(undocumented)` |
| `internal/modules/auth/delivery/http/middleware.go:28` | func | `NewMiddleware` | `func NewMiddleware(service *authapp.Service, cfg authapp.Config, enabled bool) *Middleware` | `(undocumented)` |
| `internal/modules/auth/delivery/http/middleware.go:35` | func | `WithPublicPathChecker` | `func (m *Middleware) WithPublicPathChecker(fn PublicPathChecker) *Middleware` | `WithPublicPathChecker replaces the built-in public-path list with the` |
| `internal/modules/auth/delivery/http/middleware.go:40` | func | `isPublic` | `func (m *Middleware) isPublic(method, path string) bool` | `(undocumented)` |
| `internal/modules/auth/delivery/http/middleware.go:47` | func | `Wrap` | `func (m *Middleware) Wrap(next http.Handler) http.Handler` | `(undocumented)` |
| `internal/modules/auth/domain/context.go:9` | func | `WithCurrentUser` | `func WithCurrentUser(ctx context.Context, user CurrentUser) context.Context` | `(undocumented)` |
| `internal/modules/auth/domain/context.go:13` | func | `CurrentUserFromContext` | `func CurrentUserFromContext(ctx context.Context) (CurrentUser, bool)` | `(undocumented)` |
| `internal/modules/auth/domain/errors.go:6` | var | `ErrInvalidCredentials` | `ErrInvalidCredentials` | `(undocumented)` |
| `internal/modules/auth/domain/errors.go:7` | var | `ErrSessionNotFound` | `ErrSessionNotFound` | `(undocumented)` |
| `internal/modules/auth/domain/errors.go:8` | var | `ErrSessionExpired` | `ErrSessionExpired` | `(undocumented)` |
| `internal/modules/auth/domain/errors.go:9` | var | `ErrSessionRevoked` | `ErrSessionRevoked` | `(undocumented)` |
| `internal/modules/auth/domain/errors.go:10` | var | `ErrPasswordPolicy` | `ErrPasswordPolicy` | `(undocumented)` |
| `internal/modules/auth/domain/errors.go:11` | var | `ErrPasswordChangeRequired` | `ErrPasswordChangeRequired` | `(undocumented)` |
| `internal/modules/auth/domain/errors.go:12` | var | `ErrIdentityLocked` | `ErrIdentityLocked` | `(undocumented)` |
| `internal/modules/auth/domain/errors.go:13` | var | `ErrIdentityInactive` | `ErrIdentityInactive` | `(undocumented)` |
| `internal/modules/auth/domain/errors.go:14` | var | `ErrIdentityNotFound` | `ErrIdentityNotFound` | `(undocumented)` |
| `internal/modules/auth/domain/errors.go:15` | var | `ErrUserAlreadyExists` | `ErrUserAlreadyExists` | `(undocumented)` |
| `internal/modules/auth/domain/model.go:9` | type | `Identity` | `type Identity` | `(undocumented)` |
| `internal/modules/auth/domain/model.go:26` | type | `Session` | `type Session` | `(undocumented)` |
| `internal/modules/auth/domain/model.go:37` | type | `OnlineUser` | `type OnlineUser` | `(undocumented)` |
| `internal/modules/auth/domain/model.go:44` | type | `ManagedUser` | `type ManagedUser` | `(undocumented)` |
| `internal/modules/auth/domain/model.go:59` | type | `CreateUserParams` | `type CreateUserParams` | `(undocumented)` |
| `internal/modules/auth/domain/model.go:72` | type | `UpdateUserParams` | `type UpdateUserParams` | `(undocumented)` |
| `internal/modules/auth/domain/model.go:82` | type | `BootstrapAdminParams` | `type BootstrapAdminParams` | `(undocumented)` |
| `internal/modules/auth/domain/model.go:92` | type | `CurrentUser` | `type CurrentUser` | `(undocumented)` |
| `internal/modules/auth/domain/model.go:101` | type | `AuthenticatedSession` | `type AuthenticatedSession` | `(undocumented)` |
| `internal/modules/auth/domain/port.go:8` | iface | `Repository` | `type Repository interface` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:13` | type | `Repository` | `type Repository` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:20` | func | `NewRepository` | `func NewRepository() *Repository` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:28` | func | `FindIdentityByIdentifier` | `func (r *Repository) FindIdentityByIdentifier(_ context.Context, identifier string) (authdomain.Identity, error)` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:43` | func | `FindIdentityByUserID` | `func (r *Repository) FindIdentityByUserID(_ context.Context, userID string) (authdomain.Identity, error)` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:54` | func | `CreateSession` | `func (r *Repository) CreateSession(_ context.Context, session authdomain.Session) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:61` | func | `FindSession` | `func (r *Repository) FindSession(_ context.Context, sessionID string) (authdomain.Session, error)` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:72` | func | `TouchSession` | `func (r *Repository) TouchSession(_ context.Context, sessionID string, seenAt time.Time) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:85` | func | `RevokeSession` | `func (r *Repository) RevokeSession(_ context.Context, sessionID string, revokedAt time.Time) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:98` | func | `RevokeSessionsByUserID` | `func (r *Repository) RevokeSessionsByUserID(_ context.Context, userID string, revokedAt time.Time) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:112` | func | `RecordSuccessfulLogin` | `func (r *Repository) RecordSuccessfulLogin(_ context.Context, userID string, loginAt time.Time) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:128` | func | `RecordFailedLogin` | `func (r *Repository) RecordFailedLogin(_ context.Context, userID string, failedAttempts int, lockedUntil *time.Time) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:143` | func | `CreateUser` | `func (r *Repository) CreateUser(_ context.Context, params authdomain.CreateUserParams) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:181` | func | `ListUsers` | `func (r *Repository) ListUsers(_ context.Context) ([]authdomain.ManagedUser, error)` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:205` | func | `ListOnlineUsers` | `func (r *Repository) ListOnlineUsers(_ context.Context, activeSince time.Time) ([]authdomain.OnlineUser, error)` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:243` | func | `UpdateUser` | `func (r *Repository) UpdateUser(_ context.Context, params authdomain.UpdateUserParams) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:283` | func | `BootstrapAdmin` | `func (r *Repository) BootstrapAdmin(_ context.Context, params authdomain.BootstrapAdminParams) (bool, error)` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:315` | func | `RolesByUserID` | `func (r *Repository) RolesByUserID(_ context.Context, userID, _ string) ([]iamdomain.Role, error)` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:331` | func | `UpsertUserAndAssignRole` | `func (r *Repository) UpsertUserAndAssignRole(_ context.Context, userID, displayName, _ string, role iamdomain.Role, _ string) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:360` | func | `HasAnyRole` | `func (r *Repository) HasAnyRole(_ context.Context, role iamdomain.Role, _ string) (bool, error)` | `(undocumented)` |
| `internal/modules/auth/infrastructure/memory/repository.go:374` | func | `ReplaceUserRoles` | `func (r *Repository) ReplaceUserRoles(_ context.Context, userID, displayName, _ string, roles []iamdomain.Role, _ string) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:13` | type | `Repository` | `type Repository` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:17` | func | `NewRepository` | `func NewRepository(db *sql.DB) *Repository` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:21` | func | `FindIdentityByIdentifier` | `func (r *Repository) FindIdentityByIdentifier(ctx context.Context, identifier string) (authdomain.Identity, error)` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:32` | func | `FindIdentityByUserID` | `func (r *Repository) FindIdentityByUserID(ctx context.Context, userID string) (authdomain.Identity, error)` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:43` | func | `CreateSession` | `func (r *Repository) CreateSession(ctx context.Context, session authdomain.Session) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:55` | func | `FindSession` | `func (r *Repository) FindSession(ctx context.Context, sessionID string) (authdomain.Session, error)` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:80` | func | `TouchSession` | `func (r *Repository) TouchSession(ctx context.Context, sessionID string, seenAt time.Time) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:93` | func | `RevokeSession` | `func (r *Repository) RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:106` | func | `RevokeSessionsByUserID` | `func (r *Repository) RevokeSessionsByUserID(ctx context.Context, userID string, revokedAt time.Time) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:120` | func | `RecordSuccessfulLogin` | `func (r *Repository) RecordSuccessfulLogin(ctx context.Context, userID string, loginAt time.Time) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:136` | func | `RecordFailedLogin` | `func (r *Repository) RecordFailedLogin(ctx context.Context, userID string, failedAttempts int, lockedUntil *time.Time) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:151` | func | `CreateUser` | `func (r *Repository) CreateUser(ctx context.Context, params authdomain.CreateUserParams) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:176` | func | `ListUsers` | `func (r *Repository) ListUsers(ctx context.Context) ([]authdomain.ManagedUser, error)` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:215` | func | `ListOnlineUsers` | `func (r *Repository) ListOnlineUsers(ctx context.Context, activeSince time.Time) ([]authdomain.OnlineUser, error)` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:252` | func | `UpdateUser` | `func (r *Repository) UpdateUser(ctx context.Context, params authdomain.UpdateUserParams) error` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:298` | func | `BootstrapAdmin` | `func (r *Repository) BootstrapAdmin(ctx context.Context, params authdomain.BootstrapAdminParams) (bool, error)` | `(undocumented)` |
| `internal/modules/auth/infrastructure/postgres/repository.go:329` | func | `loadIdentity` | `func (r *Repository) loadIdentity(ctx context.Context, query string, arg string) (authdomain.Identity, error)` | `(undocumented)` |

## 3. HTTP operations

| Method | Path | Handler | Source line |
|---|---|---|---|
| `POST` | `/api/v1/auth/login` | `h.handleLogin` | `internal/modules/auth/delivery/http/handler.go:36` |
| `POST` | `/api/v1/auth/logout` | `h.handleLogout` | `internal/modules/auth/delivery/http/handler.go:37` |
| `GET` | `/api/v1/auth/me` | `h.handleMe` | `internal/modules/auth/delivery/http/handler.go:38` |
| `POST` | `/api/v1/auth/change-password` | `h.handleChangePassword` | `internal/modules/auth/delivery/http/handler.go:39` |

## 4. Migrations

| Filename | Verb | Tables touched |
|---|---|---|
| `migrations/0002_init_iam_rbac.sql` | `CREATE` | `metaldocs.iam_users`, `metaldocs.iam_user_roles` |
| `migrations/0021_init_auth_identities_and_sessions.sql` | `CREATE` | `metaldocs.auth_identities`, `metaldocs.auth_sessions`, `metaldocs.iam_users` |
| `migrations/0036_decouple_auth_identity_from_iam_user_tables.sql` | `ALTER` | `metaldocs.auth_identities`, `metaldocs.auth_sessions`, `metaldocs.iam_users` |
| `migrations/0130_iam_users_tenant_deactivated.sql` | `ALTER` | `metaldocs.iam_users` |
| `migrations/0135_approval_instances.sql` | `CREATE` | `approval_instances`, `approval_stage_instances`, `approval_signoffs`, `metaldocs.iam_users` |
| `migrations/0136_user_process_areas_hardening.sql` | `ALTER` | `user_process_areas`, `metaldocs.iam_users` |
| `migrations/0137_db_roles_security_definer.sql` | `CREATE` | `metaldocs.iam_users`, `public.user_process_areas` |
| `migrations/0143_area_membership_fns.sql` | `CREATE` | `metaldocs.iam_users`, `public.user_process_areas`, `metaldocs.governance_events` |
| `migrations/0151_seed_dev_tenant_approval_data.sql` | `INSERT` | `metaldocs.role_capabilities`, `metaldocs.iam_users`, `public.user_process_areas` |
| `migrations/0159_seed_dev_approver_user.sql` | `INSERT` | `metaldocs.auth_identities`, `metaldocs.iam_users`, `metaldocs.iam_user_roles` |
| `migrations/0173_signoff_actor_displayname_snapshot.sql` | `ALTER` | `public.approval_signoffs`, `metaldocs.iam_users`, `public.schema_migrations` |
| `migrations/0174_documents_created_by_displayname_snapshot.sql` | `ALTER` | `public.documents`, `metaldocs.iam_users`, `public.schema_migrations` |
