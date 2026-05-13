# IAM Surface Scan

## 1. File tree

### domain/
- `domain/capabilities.go` - capability constants for IAM/module permission checks.
- `domain/context.go` - auth context helpers for user ID and roles.
- `domain/errors.go` - domain-level IAM not-found/inactive errors.
- `domain/model.go` - core role/capability types and base constants.
- `domain/port.go` - role provider and role admin repository interfaces.
- `domain/role_capabilities.go` - role-to-capability map and version constant.
- `domain/user_area.go` - user-area membership model and active-window check.

### application/
- `application/admin_service.go` - role assignment orchestration for user admin.
- `application/area_membership_service.go` - area membership grant/revoke/list orchestration.
- `application/authorization.go` - authorization decision service and policy checks.
- `application/cached_role_provider.go` - TTL-cached role provider wrapper.
- `application/capability_service.go` - DB-backed tenant capability check.
- `application/dev_role_provider.go` - deterministic in-memory role provider.
- `application/startup.go` - role-capability version drift check and governance event write.

### area_membership/
- `area_membership/area_membership.go` - SQL transaction functions for grant/revoke/list memberships.

### authz/
- `authz/authz.go` - in-transaction capability enforcement helpers.
- `authz/context.go` - actor/tenant context extraction from DB GUCs.

### delivery/http/
- `delivery/http/admin_handler.go` - IAM admin HTTP routes and handlers.
- `delivery/http/middleware.go` - HTTP middleware wrapper for capability checks.
- `delivery/http/routes_memberships.go` - IAM area-membership HTTP routes.

### infrastructure/memory/
- `infrastructure/memory/role_admin_repository.go` - in-memory role admin repository.

### infrastructure/postgres/
- `infrastructure/postgres/role_admin_repository.go` - Postgres-backed role admin repository.
- `infrastructure/postgres/role_provider.go` - Postgres-backed role lookup provider.
- `infrastructure/postgres/user_area_repository.go` - Postgres-backed user-area membership repository.

## 2. Public surface

| File:line | Kind | Name | Signature / receiver | Doc comment first line |
|---|---|---|---|---|
| `internal/modules/iam/area_membership/area_membership.go:18` | var | `ErrInsufficientPrivilege` | `var ErrInsufficientPrivilege error` | `(undocumented)` |
| `internal/modules/iam/area_membership/area_membership.go:19` | var | `ErrMembershipNotFound` | `var ErrMembershipNotFound error` | `(undocumented)` |
| `internal/modules/iam/area_membership/area_membership.go:20` | var | `ErrInvalidArgument` | `var ErrInvalidArgument error` | `(undocumented)` |
| `internal/modules/iam/area_membership/area_membership.go:24` | type | `Membership` | `type Membership struct` | `(undocumented)` |
| `internal/modules/iam/area_membership/area_membership.go:53` | func | `Grant` | `func Grant(ctx context.Context, tx *sql.Tx, tenantID, userID, areaCode, role, grantedBy string) (correlationID string, err error)` | `(undocumented)` |
| `internal/modules/iam/area_membership/area_membership.go:65` | func | `Revoke` | `func Revoke(ctx context.Context, tx *sql.Tx, tenantID, userID, areaCode, role, revokedBy string) error` | `(undocumented)` |
| `internal/modules/iam/area_membership/area_membership.go:77` | func | `List` | `func List(ctx context.Context, tx *sql.Tx, tenantID, userID string) ([]Membership, error)` | `(undocumented)` |
| `internal/modules/iam/application/admin_service.go:10` | iface | `RoleCacheInvalidator` | `type RoleCacheInvalidator interface` | `(undocumented)` |
| `internal/modules/iam/application/admin_service.go:14` | type | `AdminService` | `type AdminService struct` | `(undocumented)` |
| `internal/modules/iam/application/admin_service.go:19` | func | `NewAdminService` | `func NewAdminService(repo domain.RoleAdminRepository, invalidator RoleCacheInvalidator) *AdminService` | `(undocumented)` |
| `internal/modules/iam/application/admin_service.go:23` | func | `UpsertUserAndAssignRole` | `func (s *AdminService) UpsertUserAndAssignRole(ctx context.Context, userID, displayName, tenantID string, role domain.Role, assignedBy string) error` | `(undocumented)` |
| `internal/modules/iam/application/admin_service.go:47` | func | `ReplaceUserRoles` | `func (s *AdminService) ReplaceUserRoles(ctx context.Context, userID, displayName, tenantID string, roles []domain.Role, assignedBy string) error` | `(undocumented)` |
| `internal/modules/iam/application/area_membership_service.go:13` | var | `ErrMembershipNotFound` | `var ErrMembershipNotFound error` | `(undocumented)` |
| `internal/modules/iam/application/area_membership_service.go:14` | var | `ErrUnknownRole` | `var ErrUnknownRole error` | `(undocumented)` |
| `internal/modules/iam/application/area_membership_service.go:17` | iface | `UserAreaWriteRepository` | `type UserAreaWriteRepository interface` | `(undocumented)` |
| `internal/modules/iam/application/area_membership_service.go:25` | iface | `MembershipGovernanceLogger` | `type MembershipGovernanceLogger interface` | `(undocumented)` |
| `internal/modules/iam/application/area_membership_service.go:29` | type | `AreaMembershipService` | `type AreaMembershipService struct` | `(undocumented)` |
| `internal/modules/iam/application/area_membership_service.go:35` | func | `NewAreaMembershipService` | `func NewAreaMembershipService(repo UserAreaWriteRepository, logger MembershipGovernanceLogger) *AreaMembershipService` | `(undocumented)` |
| `internal/modules/iam/application/area_membership_service.go:45` | func | `ListActive` | `func (s *AreaMembershipService) ListActive(ctx context.Context, userID, tenantID string) ([]domain.UserProcessArea, error)` | `(undocumented)` |
| `internal/modules/iam/application/area_membership_service.go:49` | func | `Grant` | `func (s *AreaMembershipService) Grant(ctx context.Context, userID, tenantID, areaCode string, role domain.Role, grantedBy string) error` | `(undocumented)` |
| `internal/modules/iam/application/area_membership_service.go:108` | func | `Revoke` | `func (s *AreaMembershipService) Revoke(ctx context.Context, userID, tenantID, areaCode string, revokedBy string) error` | `(undocumented)` |
| `internal/modules/iam/application/authorization.go:15` | var | `ErrAccessDenied` | `var ErrAccessDenied error` | `(undocumented)` |
| `internal/modules/iam/application/authorization.go:16` | var | `ErrSoDViolation` | `var ErrSoDViolation error` | `(undocumented)` |
| `internal/modules/iam/application/authorization.go:17` | var | `ErrAreaRequired` | `var ErrAreaRequired error` | `(undocumented)` |
| `internal/modules/iam/application/authorization.go:20` | iface | `UserAreaRepository` | `type UserAreaRepository interface` | `(undocumented)` |
| `internal/modules/iam/application/authorization.go:24` | type | `AccessPolicy` | `type AccessPolicy struct` | `(undocumented)` |
| `internal/modules/iam/application/authorization.go:29` | iface | `AccessPolicyRepository` | `type AccessPolicyRepository interface` | `(undocumented)` |
| `internal/modules/iam/application/authorization.go:33` | iface | `TemplateAuthorChecker` | `type TemplateAuthorChecker interface` | `(undocumented)` |
| `internal/modules/iam/application/authorization.go:37` | type | `ResourceCtx` | `type ResourceCtx struct` | `(undocumented)` |
| `internal/modules/iam/application/authorization.go:42` | type | `AuthorizationService` | `type AuthorizationService struct` | `(undocumented)` |
| `internal/modules/iam/application/authorization.go:49` | func | `NewAuthorizationService` | `func NewAuthorizationService(userAreas UserAreaRepository, accessPolicies AccessPolicyRepository, authorChecker TemplateAuthorChecker) *AuthorizationService` | `(undocumented)` |
| `internal/modules/iam/application/authorization.go:74` | func | `WithAuthzCache` | `func WithAuthzCache(ctx context.Context) context.Context` | `(undocumented)` |
| `internal/modules/iam/application/authorization.go:81` | func | `Check` | `func (s *AuthorizationService) Check(ctx context.Context, userID, tenantID string, capability domain.Capability, resource ResourceCtx) error` | `(undocumented)` |
| `internal/modules/iam/application/cached_role_provider.go:18` | type | `CachedRoleProvider` | `type CachedRoleProvider struct` | `CachedRoleProvider wraps a RoleProvider with TTL cache and explicit invalidation.` |
| `internal/modules/iam/application/cached_role_provider.go:25` | func | `NewCachedRoleProvider` | `func NewCachedRoleProvider(base domain.RoleProvider, ttl time.Duration) *CachedRoleProvider` | `(undocumented)` |
| `internal/modules/iam/application/cached_role_provider.go:40` | func | `RolesByUserID` | `func (c *CachedRoleProvider) RolesByUserID(ctx context.Context, userID, tenantID string) ([]domain.Role, error)` | `(undocumented)` |
| `internal/modules/iam/application/cached_role_provider.go:65` | func | `InvalidateUser` | `func (c *CachedRoleProvider) InvalidateUser(userID string)` | `InvalidateUser invalidates cache entries for a user across all tenants.` |
| `internal/modules/iam/application/cached_role_provider.go:75` | func | `InvalidateAll` | `func (c *CachedRoleProvider) InvalidateAll()` | `(undocumented)` |
| `internal/modules/iam/application/capability_service.go:10` | var | `ErrCapabilityDenied` | `var ErrCapabilityDenied error` | `(undocumented)` |
| `internal/modules/iam/application/capability_service.go:12` | type | `CapabilityService` | `type CapabilityService struct` | `(undocumented)` |
| `internal/modules/iam/application/capability_service.go:16` | func | `NewCapabilityService` | `func NewCapabilityService(db *sql.DB) *CapabilityService` | `(undocumented)` |
| `internal/modules/iam/application/capability_service.go:31` | func | `CanDo` | `func (s *CapabilityService) CanDo(ctx context.Context, userID, tenantID, capability string) error` | `CanDo enforces a tier-1 tenant-level capability check.` |
| `internal/modules/iam/application/dev_role_provider.go:11` | type | `DevRoleProvider` | `type DevRoleProvider struct` | `DevRoleProvider is a deterministic in-memory provider used for local memory mode.` |
| `internal/modules/iam/application/dev_role_provider.go:15` | func | `NewDevRoleProvider` | `func NewDevRoleProvider(rolesByUser map[string][]domain.Role) *DevRoleProvider` | `(undocumented)` |
| `internal/modules/iam/application/dev_role_provider.go:22` | func | `RolesByUserID` | `func (p *DevRoleProvider) RolesByUserID(_ context.Context, userID, _ string) ([]domain.Role, error)` | `(undocumented)` |
| `internal/modules/iam/application/startup.go:15` | func | `CheckRoleCapabilitiesVersion` | `func CheckRoleCapabilitiesVersion(ctx context.Context, db *sql.DB, tenantID string) error` | `(undocumented)` |
| `internal/modules/iam/authz/authz.go:11` | type | `ErrCapabilityDenied` | `type ErrCapabilityDenied struct` | `(undocumented)` |
| `internal/modules/iam/authz/authz.go:17` | func | `Error` | `func (e ErrCapabilityDenied) Error() string` | `(undocumented)` |
| `internal/modules/iam/authz/authz.go:28` | func | `WithCapCache` | `func WithCapCache(ctx context.Context) context.Context` | `(undocumented)` |
| `internal/modules/iam/authz/authz.go:44` | func | `Require` | `func Require(ctx context.Context, tx *sql.Tx, capability, areaCode string) error` | `(undocumented)` |
| `internal/modules/iam/authz/authz.go:99` | func | `BypassSystem` | `func BypassSystem(ctx context.Context, tx *sql.Tx) error` | `(undocumented)` |
| `internal/modules/iam/authz/context.go:13` | var | `ErrActorContextMissing` | `var ErrActorContextMissing error` | `ErrActorContextMissing indicates actor context was not set in DB session.` |
| `internal/modules/iam/authz/context.go:17` | var | `ErrTenantContextMissing` | `var ErrTenantContextMissing error` | `ErrTenantContextMissing indicates tenant context was not set in DB session.` |
| `internal/modules/iam/authz/context.go:21` | func | `MustActorID` | `func MustActorID(ctx context.Context, tx *sql.Tx) (string, error)` | `MustActorID resolves actor ID from DB transaction context.` |
| `internal/modules/iam/authz/context.go:34` | func | `MustTenantID` | `func MustTenantID(ctx context.Context, tx *sql.Tx) (string, error)` | `MustTenantID resolves tenant ID from DB transaction context.` |
| `internal/modules/iam/delivery/http/admin_handler.go:20` | iface | `UserAdminService` | `type UserAdminService interface` | `(undocumented)` |
| `internal/modules/iam/delivery/http/admin_handler.go:29` | type | `AdminHandler` | `type AdminHandler struct` | `(undocumented)` |
| `internal/modules/iam/delivery/http/admin_handler.go:36` | type | `UpsertUserRoleRequest` | `type UpsertUserRoleRequest struct` | `(undocumented)` |
| `internal/modules/iam/delivery/http/admin_handler.go:42` | type | `ReplaceUserRolesRequest` | `type ReplaceUserRolesRequest struct` | `(undocumented)` |
| `internal/modules/iam/delivery/http/admin_handler.go:48` | type | `CreateUserRequest` | `type CreateUserRequest struct` | `(undocumented)` |
| `internal/modules/iam/delivery/http/admin_handler.go:57` | type | `UpdateUserRequest` | `type UpdateUserRequest struct` | `(undocumented)` |
| `internal/modules/iam/delivery/http/admin_handler.go:65` | type | `ResetPasswordRequest` | `type ResetPasswordRequest struct` | `(undocumented)` |
| `internal/modules/iam/delivery/http/admin_handler.go:69` | func | `NewAdminHandler` | `func NewAdminHandler(service *iamapp.AdminService, authService UserAdminService, auditWriter ...auditdomain.Writer) *AdminHandler` | `(undocumented)` |
| `internal/modules/iam/delivery/http/admin_handler.go:77` | func | `WithAuditReader` | `func (h *AdminHandler) WithAuditReader(reader auditdomain.Reader) *AdminHandler` | `(undocumented)` |
| `internal/modules/iam/delivery/http/admin_handler.go:82` | func | `RegisterRoutes` | `func (h *AdminHandler) RegisterRoutes(mux *http.ServeMux)` | `(undocumented)` |
| `internal/modules/iam/delivery/http/middleware.go:21` | type | `PermissionResolver` | `type PermissionResolver func(method, path string) (string, bool)` | `(undocumented)` |
| `internal/modules/iam/delivery/http/middleware.go:23` | type | `Middleware` | `type Middleware struct` | `(undocumented)` |
| `internal/modules/iam/delivery/http/middleware.go:31` | func | `NewMiddleware` | `func NewMiddleware(caps *iamapp.CapabilityService, roleProvider iamdomain.RoleProvider, enabled bool, legacyHeader ...bool) *Middleware` | `(undocumented)` |
| `internal/modules/iam/delivery/http/middleware.go:44` | func | `WithPermissionResolver` | `func (m *Middleware) WithPermissionResolver(resolver PermissionResolver) *Middleware` | `(undocumented)` |
| `internal/modules/iam/delivery/http/middleware.go:49` | func | `Wrap` | `func (m *Middleware) Wrap(next http.Handler) http.Handler` | `(undocumented)` |
| `internal/modules/iam/delivery/http/routes_memberships.go:15` | type | `MembershipHandler` | `type MembershipHandler struct` | `(undocumented)` |
| `internal/modules/iam/delivery/http/routes_memberships.go:25` | func | `NewMembershipHandler` | `func NewMembershipHandler(svc *iamapp.AreaMembershipService) *MembershipHandler` | `(undocumented)` |
| `internal/modules/iam/delivery/http/routes_memberships.go:29` | func | `RegisterRoutes` | `func (h *MembershipHandler) RegisterRoutes(mux *http.ServeMux)` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:4` | const | `CapDocView` | `const CapDocView = "doc.view"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:5` | const | `CapDocCreate` | `const CapDocCreate = "doc.create"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:6` | const | `CapDocEdit` | `const CapDocEdit = "doc.edit"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:7` | const | `CapDocSubmit` | `const CapDocSubmit = "doc.submit"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:8` | const | `CapDocSignoff` | `const CapDocSignoff = "doc.signoff"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:9` | const | `CapTemplateView` | `const CapTemplateView = "template.view"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:10` | const | `CapTemplateCreate` | `const CapTemplateCreate = "template.create"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:11` | const | `CapTemplateEdit` | `const CapTemplateEdit = "template.edit"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:12` | const | `CapTemplateSubmit` | `const CapTemplateSubmit = "template.submit"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:13` | const | `CapTemplateApprove` | `const CapTemplateApprove = "template.approve"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:14` | const | `CapTemplatePublish` | `const CapTemplatePublish = "template.publish"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:15` | const | `CapRegistryCreate` | `const CapRegistryCreate = "registry.create"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:16` | const | `CapTaxonomyManage` | `const CapTaxonomyManage = "taxonomy.manage"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:17` | const | `CapMembershipManage` | `const CapMembershipManage = "membership.manage"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:18` | const | `CapRouteManage` | `const CapRouteManage = "route.manage"` | `(undocumented)` |
| `internal/modules/iam/domain/capabilities.go:19` | const | `CapUserManage` | `const CapUserManage = "user.manage"` | `(undocumented)` |
| `internal/modules/iam/domain/context.go:12` | func | `WithAuthContext` | `func WithAuthContext(ctx context.Context, userID string, roles []Role) context.Context` | `(undocumented)` |
| `internal/modules/iam/domain/context.go:17` | func | `UserIDFromContext` | `func UserIDFromContext(ctx context.Context) string` | `(undocumented)` |
| `internal/modules/iam/domain/context.go:25` | func | `RolesFromContext` | `func RolesFromContext(ctx context.Context) []Role` | `(undocumented)` |
| `internal/modules/iam/domain/errors.go:6` | var | `ErrUserNotFound` | `var ErrUserNotFound error` | `(undocumented)` |
| `internal/modules/iam/domain/errors.go:7` | var | `ErrUserInactive` | `var ErrUserInactive error` | `(undocumented)` |
| `internal/modules/iam/domain/model.go:3` | type | `Role` | `type Role string` | `(undocumented)` |
| `internal/modules/iam/domain/model.go:6` | const | `RoleApprover` | `const RoleApprover Role = "approver"` | `(undocumented)` |
| `internal/modules/iam/domain/model.go:7` | const | `RoleAuthor` | `const RoleAuthor Role = "author"` | `(undocumented)` |
| `internal/modules/iam/domain/model.go:8` | const | `RoleEditor` | `const RoleEditor Role = "editor"` | `(undocumented)` |
| `internal/modules/iam/domain/model.go:9` | const | `RoleSystemAdmin` | `const RoleSystemAdmin Role = "system_admin"` | `(undocumented)` |
| `internal/modules/iam/domain/model.go:10` | const | `RoleViewer` | `const RoleViewer Role = "viewer"` | `(undocumented)` |
| `internal/modules/iam/domain/model.go:13` | type | `Capability` | `type Capability string` | `(undocumented)` |
| `internal/modules/iam/domain/model.go:16` | const | `CapDocumentView` | `const CapDocumentView Capability = "document.view"` | `(undocumented)` |
| `internal/modules/iam/domain/model.go:17` | const | `CapDocumentCreate` | `const CapDocumentCreate Capability = "document.create"` | `(undocumented)` |
| `internal/modules/iam/domain/model.go:18` | const | `CapDocumentEdit` | `const CapDocumentEdit Capability = "document.edit"` | `(undocumented)` |
| `internal/modules/iam/domain/model.go:19` | const | `CapWorkflowReview` | `const CapWorkflowReview Capability = "workflow.review"` | `(undocumented)` |
| `internal/modules/iam/domain/model.go:20` | const | `CapWorkflowApprove` | `const CapWorkflowApprove Capability = "workflow.approve"` | `(undocumented)` |
| `internal/modules/iam/domain/port.go:6` | iface | `RoleProvider` | `type RoleProvider interface` | `RoleProvider resolves effective roles for a given user identity within a tenant.` |
| `internal/modules/iam/domain/port.go:11` | iface | `RoleAdminRepository` | `type RoleAdminRepository interface` | `RoleAdminRepository writes IAM user and role assignments.` |
| `internal/modules/iam/domain/role_capabilities.go:3` | const | `RoleCapabilitiesVersion` | `const RoleCapabilitiesVersion = 2` | `(undocumented)` |
| `internal/modules/iam/domain/role_capabilities.go:5` | var | `RoleCapabilities` | `var RoleCapabilities map[Role][]Capability` | `(undocumented)` |
| `internal/modules/iam/domain/user_area.go:5` | type | `UserProcessArea` | `type UserProcessArea struct` | `(undocumented)` |
| `internal/modules/iam/domain/user_area.go:15` | func | `IsActive` | `func (m UserProcessArea) IsActive(now time.Time) bool` | `(undocumented)` |
| `internal/modules/iam/infrastructure/memory/role_admin_repository.go:10` | type | `RoleAdminRepository` | `type RoleAdminRepository struct` | `(undocumented)` |
| `internal/modules/iam/infrastructure/memory/role_admin_repository.go:20` | func | `NewRoleAdminRepository` | `func NewRoleAdminRepository() *RoleAdminRepository` | `(undocumented)` |
| `internal/modules/iam/infrastructure/memory/role_admin_repository.go:24` | func | `HasAnyRole` | `func (r *RoleAdminRepository) HasAnyRole(_ context.Context, role domain.Role, _ string) (bool, error)` | `(undocumented)` |
| `internal/modules/iam/infrastructure/memory/role_admin_repository.go:36` | func | `UpsertUserAndAssignRole` | `func (r *RoleAdminRepository) UpsertUserAndAssignRole(_ context.Context, userID, displayName, _ string, role domain.Role, _ string) error` | `(undocumented)` |
| `internal/modules/iam/infrastructure/memory/role_admin_repository.go:50` | func | `ReplaceUserRoles` | `func (r *RoleAdminRepository) ReplaceUserRoles(_ context.Context, userID, displayName, _ string, roles []domain.Role, _ string) error` | `(undocumented)` |
| `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:12` | type | `RoleAdminRepository` | `type RoleAdminRepository struct` | `(undocumented)` |
| `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:16` | func | `NewRoleAdminRepository` | `func NewRoleAdminRepository(db *sql.DB) *RoleAdminRepository` | `(undocumented)` |
| `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:20` | func | `HasAnyRole` | `func (r *RoleAdminRepository) HasAnyRole(ctx context.Context, role domain.Role, tenantID string) (bool, error)` | `(undocumented)` |
| `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:33` | func | `UpsertUserAndAssignRole` | `func (r *RoleAdminRepository) UpsertUserAndAssignRole(ctx context.Context, userID, displayName, tenantID string, role domain.Role, assignedBy string) error` | `(undocumented)` |
| `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:72` | func | `ReplaceUserRoles` | `func (r *RoleAdminRepository) ReplaceUserRoles(ctx context.Context, userID, displayName, tenantID string, roles []domain.Role, assignedBy string) error` | `ReplaceUserRoles writes the user+role assignment.` |
| `internal/modules/iam/infrastructure/postgres/role_provider.go:11` | type | `RoleProvider` | `type RoleProvider struct` | `(undocumented)` |
| `internal/modules/iam/infrastructure/postgres/role_provider.go:15` | func | `NewRoleProvider` | `func NewRoleProvider(db *sql.DB) *RoleProvider` | `(undocumented)` |
| `internal/modules/iam/infrastructure/postgres/role_provider.go:19` | func | `RolesByUserID` | `func (p *RoleProvider) RolesByUserID(ctx context.Context, userID, tenantID string) ([]domain.Role, error)` | `(undocumented)` |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:13` | type | `UserAreaRepository` | `type UserAreaRepository struct` | `(undocumented)` |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:17` | func | `NewUserAreaRepository` | `func NewUserAreaRepository(db *sql.DB) *UserAreaRepository` | `(undocumented)` |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:21` | func | `ListActive` | `func (r *UserAreaRepository) ListActive(ctx context.Context, userID, tenantID string, now time.Time) ([]domain.UserProcessArea, error)` | `(undocumented)` |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:51` | func | `Insert` | `func (r *UserAreaRepository) Insert(ctx context.Context, membership domain.UserProcessArea) error` | `(undocumented)` |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:75` | func | `CloseActive` | `func (r *UserAreaRepository) CloseActive(ctx context.Context, userID, tenantID, areaCode string, effectiveTo time.Time) error` | `(undocumented)` |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:90` | func | `GrantAtomic` | `func (r *UserAreaRepository) GrantAtomic(ctx context.Context, oldMembership, newMembership domain.UserProcessArea) error` | `(undocumented)` |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:154` | func | `GetActiveByUserAndArea` | `func (r *UserAreaRepository) GetActiveByUserAndArea(ctx context.Context, userID, tenantID, areaCode string, now time.Time) (*domain.UserProcessArea, error)` | `(undocumented)` |

## 3. HTTP operations

| Method | Path | Handler symbol | Source file:line |
|---|---|---|---|
| `GET` | `/api/v1/iam/area-memberships` | `(*MembershipHandler).listMemberships` | `internal/modules/iam/delivery/http/routes_memberships.go:30` |
| `POST` | `/api/v1/iam/area-memberships` | `(*MembershipHandler).grantMembership` | `internal/modules/iam/delivery/http/routes_memberships.go:31` |
| `DELETE` | `/api/v1/iam/area-memberships` | `(*MembershipHandler).revokeMembership` | `internal/modules/iam/delivery/http/routes_memberships.go:32` |
| `GET` | `/api/v1/iam/users` | `(*AdminHandler).handleListUsers` | `internal/modules/iam/delivery/http/admin_handler.go:88` |
| `POST` | `/api/v1/iam/users` | `(*AdminHandler).handleCreateUser` | `internal/modules/iam/delivery/http/admin_handler.go:90` |
| `GET` | `/api/v1/iam/admin/overview` | `(*AdminHandler).handleAdminOverview` | `internal/modules/iam/delivery/http/admin_handler.go:85` |
| `POST` | `/api/v1/iam/users/{userId}/roles` | `(*AdminHandler).handleUserRoleUpsert` | `internal/modules/iam/delivery/http/admin_handler.go:196` |
| `PUT` | `/api/v1/iam/users/{userId}/roles` | `(*AdminHandler).handleReplaceUserRoles` | `internal/modules/iam/delivery/http/admin_handler.go:198` |
| `POST` | `/api/v1/iam/users/{userId}/reset-password` | `(*AdminHandler).handleResetPassword` | `internal/modules/iam/delivery/http/admin_handler.go:206` |
| `POST` | `/api/v1/iam/users/{userId}/unlock` | `(*AdminHandler).handleUnlockUser` | `internal/modules/iam/delivery/http/admin_handler.go:210` |
| `PATCH` | `/api/v1/iam/users/{userId}` | `(*AdminHandler).handlePatchUser` | `internal/modules/iam/delivery/http/admin_handler.go:214` |

## 4. Migration list

migrations: external — see persistence artifact
