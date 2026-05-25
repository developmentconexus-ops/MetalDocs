package taxonomy

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/taxonomy/application"
	thttp "metaldocs/internal/modules/taxonomy/delivery/http"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/modules/taxonomy/infrastructure"
)

type Module struct {
	Handler *thttp.Handler
}

type Dependencies struct {
	DB          *sql.DB
	TplChecker  application.TemplateVersionChecker
	AuditWriter auditdomain.Writer
}

func New(deps Dependencies) *Module {
	profileRepo := infrastructure.NewProfileRepository(deps.DB)
	areaRepo := infrastructure.NewAreaRepository(deps.DB)
	tplChecker := deps.TplChecker
	if tplChecker == nil {
		tplChecker = infrastructure.NewTemplateVersionChecker(deps.DB)
	}
	var govLogger domain.GovernanceLogger
	if deps.AuditWriter != nil {
		govLogger = application.NewAuditGovernanceAdapter(deps.AuditWriter)
	} else {
		slog.WarnContext(context.Background(), "taxonomy audit writer missing; using legacy DB governance logger")
		govLogger = application.NewDBGovernanceLogger(deps.DB)
	}

	familyRepo := infrastructure.NewFamilyRepository(deps.DB)

	profileService := application.NewProfileService(profileRepo, tplChecker, govLogger)
	areaService := application.NewAreaService(areaRepo, govLogger)
	familyService := application.NewFamilyService(familyRepo, govLogger)
	handler := thttp.NewHandler(profileService, areaService, familyService)

	return &Module{Handler: handler}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.Handler.RegisterRoutes(mux)
}
