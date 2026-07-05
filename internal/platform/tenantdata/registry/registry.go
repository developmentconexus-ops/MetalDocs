// Package registry is the composition root for every module's
// TenantDataPort. It is a separate package from internal/platform/tenantdata
// (the interface + generic SQL helpers) so the interface package stays a
// true dependency-free leaf: internal/modules/render/fanout and
// internal/modules/documents/repository (among others) need to import
// internal/platform/tenantdata to declare `var _ tenantdata.Port = (...)`,
// and if the registry (which imports every implementing module) lived in
// that same package, those modules would transitively import themselves —
// an import cycle. Splitting the wiring into this sibling package keeps the
// dependency direction one-way: registry -> {tenantdata, every module},
// never the reverse.
package registry

import (
	"database/sql"

	"metaldocs/internal/platform/tenantdata"

	auditpg "metaldocs/internal/modules/audit/infrastructure/postgres"
	authpg "metaldocs/internal/modules/auth/infrastructure/postgres"
	cdinfra "metaldocs/internal/modules/controlleddocuments/infrastructure"
	approvalrepo "metaldocs/internal/modules/documents/approval/repository"
	docsrepo "metaldocs/internal/modules/documents/repository"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	jobstenantdata "metaldocs/internal/modules/jobs/tenantdata"
	notificationsinfra "metaldocs/internal/modules/notifications/infrastructure"
	renderfanout "metaldocs/internal/modules/render/fanout"
	taxonomyinfra "metaldocs/internal/modules/taxonomy/infrastructure"
	templatesrepo "metaldocs/internal/modules/templates/repository"
	tokensinfra "metaldocs/internal/modules/tokens/infrastructure"
)

// AllTenantDataPorts is the single composition-root list every module's
// TenantDataPort is registered into. Both apps/api (export/erase HTTP
// handlers) and apps/jobs (the River worker that runs the fan-out) construct
// their orchestrator from this same slice — one place, no duplicate wiring,
// no init-time magic (deliberately a plain function, not a package-level
// var/init registration pattern, so construction order and the db handle
// are explicit at every call site).
//
// Allowlisted tables intentionally NOT covered by any port here (enforced by
// the coverage test's explicit allowlist, tests/integration/tenantdata):
//   - metaldocs.tenants — the tenant root row; the erase orchestrator
//     tombstones it directly, it is never "erased" as tenant data.
//   - metaldocs.tenant_keys — the per-tenant crypto key row; destroyed by
//     the security module's crypto-shred step (TenantCrypto.DestroyTenantKeyTx),
//     not by a TenantDataPort (a different feature/task).
func AllTenantDataPorts(db *sql.DB) []tenantdata.Port {
	return []tenantdata.Port{
		auditpg.NewTenantDataPort(db),
		authpg.NewTenantDataPort(db),
		cdinfra.NewTenantDataPort(db),
		approvalrepo.NewTenantDataPort(db),
		docsrepo.NewTenantDataPort(db),
		iampg.NewTenantDataPort(db),
		jobstenantdata.NewTenantDataPort(db),
		notificationsinfra.NewTenantDataPort(db),
		renderfanout.NewTenantDataPort(db),
		taxonomyinfra.NewTenantDataPort(db),
		templatesrepo.NewTenantDataPort(db),
		tokensinfra.NewTenantDataPort(db),
	}
}
