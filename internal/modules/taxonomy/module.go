// Package taxonomy is the composition root for the taxonomy bounded
// context: it owns the flat, code-keyed classification catalog
// (DocumentFamily, DocumentProfile, ProcessArea) that controlled-documents
// and documents bind to. New wires the repository, service, and handler
// layers together; RegisterRoutes mounts the resulting HTTP surface.
package taxonomy

import (
	"database/sql"
	"net/http"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/taxonomy/application"
	thttp "metaldocs/internal/modules/taxonomy/delivery/http"
	"metaldocs/internal/modules/taxonomy/infrastructure"
)

// Module is the assembled taxonomy module: its HTTP Handler plus the
// services and repositories closed over it.
type Module struct {
	Handler *thttp.Handler
}

// Dependencies are the externally-owned collaborators New requires to
// build a Module. TplChecker and AuditWriter must not be nil.
type Dependencies struct {
	DB          *sql.DB
	TplChecker  application.TemplateVersionChecker
	AuditWriter auditdomain.Writer
}

// New builds a Module: repositories over deps.DB, the audit-backed
// GovernanceLogger, the three application services, and the HTTP handler.
// It panics if deps.TplChecker or deps.AuditWriter is nil — taxonomy has
// no silent fallback for either (fail-loud by design).
func New(deps Dependencies) *Module {
	profileRepo := infrastructure.NewProfileRepository(deps.DB)
	areaRepo := infrastructure.NewAreaRepository(deps.DB)
	tplChecker := deps.TplChecker
	if tplChecker == nil {
		panic("taxonomy.module: TplChecker is required; wire templates/infrastructure.NewTemplateVersionReader at the composition root")
	}
	if deps.AuditWriter == nil {
		panic("taxonomy.module: AuditWriter is required; nil fallback to DBGovernanceLogger was removed (FE / Wave 2.12)")
	}
	logger := application.NewAuditGovernanceAdapter(deps.AuditWriter)

	familyRepo := infrastructure.NewFamilyRepository(deps.DB)

	profileService := application.NewProfileService(profileRepo, tplChecker, logger)
	areaService := application.NewAreaService(areaRepo, logger)
	familyService := application.NewFamilyService(familyRepo, logger)
	handler := thttp.NewHandler(profileService, areaService, familyService)

	return &Module{Handler: handler}
}

// RegisterRoutes mounts the module's HTTP handler onto mux.
func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.Handler.RegisterRoutes(mux)
}
