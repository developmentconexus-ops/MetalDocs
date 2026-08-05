package tokens

import (
	"database/sql"
	"metaldocs/internal/platform/httprouter"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/tokens/application"
	tokenshttp "metaldocs/internal/modules/tokens/delivery/http"
	"metaldocs/internal/modules/tokens/domain"
	"metaldocs/internal/modules/tokens/infrastructure"
	"metaldocs/internal/platform/db"
)

// Module is the tokens bounded-context assembly. Handler is the HTTP delivery
// layer; Reader is the published provider port for SP-2 (render/fanout consumers).
type Module struct {
	Handler *tokenshttp.Handler
	Reader  domain.DictionaryReader // published for SP-2
}

// Dependencies carries everything the module needs from the composition root.
type Dependencies struct {
	DB            *sql.DB
	AuditWriter   auditdomain.Writer
	ReservedNames application.ReservedNames
}

// New assembles the tokens module. Panics on nil required deps so misconfiguration
// fails loudly at startup rather than at first request.
func New(deps Dependencies) *Module {
	if deps.DB == nil {
		panic("tokens.module: DB is required")
	}
	if deps.AuditWriter == nil {
		panic("tokens.module: AuditWriter is required")
	}
	if deps.ReservedNames == nil {
		panic("tokens.module: ReservedNames is required")
	}
	repo := infrastructure.NewPostgresRepository()
	svc := application.NewService(db.NewTxRunner(deps.DB), repo, deps.AuditWriter).
		WithReservedNames(deps.ReservedNames)
	return &Module{
		Handler: tokenshttp.NewHandler(svc),
		Reader:  svc,
	}
}

// RegisterRoutes mounts the tokens HTTP routes onto mux.
func (m *Module) RegisterRoutes(mux httprouter.Muxer) { m.Handler.RegisterRoutes(mux) }
