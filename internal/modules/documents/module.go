package documents

import (
	"context"
	"database/sql"
	"metaldocs/internal/platform/httprouter"
	"net/http"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/application"
	dhttp "metaldocs/internal/modules/documents/delivery/http"
	"metaldocs/internal/modules/documents/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	templatesdomain "metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/ratelimit"
)

// Module is the documents bounded-context module: it wires the repository,
// application services, and HTTP handlers built by New, and publishes them
// for the composition root to mount.
type Module struct {
	Handler                   *dhttp.Handler
	Service                   *application.Service
	ExportHandler             *dhttp.ExportHandler
	FillInHandler             *dhttp.FillInHandler
	PlaceholderOptionsHandler *dhttp.PlaceholderOptionsHandler
	ViewHandler               *dhttp.ViewHandler
	ReconstructHandler        *dhttp.ReconstructHandler
	repo                      *infrastructure.Repository
}

// Dependencies carries every external collaborator New needs to construct a
// Module: the DB handle, cross-module read ports, and infrastructure
// clients. Optional ports are nil-guarded in New with fail-closed Noop
// defaults.
type Dependencies struct {
	DB                           *sql.DB
	Presign                      application.Presigner
	TplRead                      application.TemplateReader
	FormVal                      application.FormValidator
	Audit                        application.Audit
	ControlledDocumentDuplicator application.ControlledDocumentDuplicator
	Caps                         application.CapabilityChecker
	ProfileDefaults              application.ProfileDefaultTemplateReader
	SnapshotReader               application.SnapshotTemplateReader
	ViewPresign                  application.ViewPresigner
	ExportPresign                application.ExportPresigner
	ExportDocgen                 application.DocgenPDFClient
	DocgenVer                    string
	GrammarVer                   string
	ReconstructRunner            application.ReconstructionRunner
	IAMUserOptions               application.IAMUserOptionsReader
	// DisplayNameReader is the iam-owned port for reading display_name off
	// metaldocs.iam_users without crossing module boundaries (M4/F4.1).
	// When nil, iamdomain.NoopUserDisplayNameReader{} is used automatically.
	DisplayNameReader iamdomain.UserDisplayNameReader
	// CDFieldReader is the controlleddocuments-owned read-port for individual
	// controlled_documents fields (profile_code) consumed without crossing the
	// module boundary (M2/F2.1; ADR-0039 D3(b)). When nil,
	// controlleddocumentsdomain.NoopCDFieldReader{} is used automatically.
	CDFieldReader controlleddocumentsdomain.CDFieldReader
	// AreaCatalogReader is the taxonomy-owned read-port for document_process_areas
	// (area name) consumed in-tx without crossing the module boundary (M2/F2.3;
	// ADR-0039 D3(b)). When nil, taxonomydomain.NoopAreaCatalogReader{} is used.
	AreaCatalogReader taxonomydomain.AreaCatalogReader
	// TemplateVersionPort is the templates-owned read-port for template version
	// state + placeholder schema (ADR-0030; extended M2/F2.4, ADR-0039 D3(b)).
	// documents reads fill-in placeholder schema through it instead of joining
	// templates_template_version to its own table. When nil, the template-schema
	// fill-in reader is not attached (matching the prior absent-reader behavior).
	TemplateVersionPort templatesdomain.TemplateVersionPort
	// PDFOutboxReader is the render/fanout-owned read-port for PDF-pipeline
	// liveness (dead-lettered materialize/pdf outbox events). ViewService uses
	// it to derive pdf_status='failed' honestly (QA-1 F13). When nil, dead
	// letters are invisible and pdf_status stays 'pending'.
	PDFOutboxReader application.PDFOutboxStateReader
}

// New constructs the documents Module from deps, wiring the repository,
// application services (document, export, fill-in, view, reconstruction),
// and their HTTP handlers. Panics if deps.Caps is nil, since handler
// authorization cannot function without it.
func New(deps Dependencies) *Module {
	if deps.Caps == nil {
		panic("documents.New: Caps (CapabilityChecker) is required for handler authorization")
	}
	displayNameReader := deps.DisplayNameReader
	if displayNameReader == nil {
		displayNameReader = iamdomain.NoopUserDisplayNameReader{}
	}
	cdFieldReader := deps.CDFieldReader
	if cdFieldReader == nil {
		cdFieldReader = controlleddocumentsdomain.NoopCDFieldReader{}
	}
	areaCatalog := deps.AreaCatalogReader
	if areaCatalog == nil {
		areaCatalog = taxonomydomain.NoopAreaCatalogReader{}
	}
	repo := infrastructure.New(deps.DB, displayNameReader, cdFieldReader, areaCatalog)
	var svc *application.Service
	if deps.SnapshotReader != nil {
		snapSvc := application.NewSnapshotService(deps.SnapshotReader)
		svc = application.NewServiceWithSnapshot(repo, deps.Presign, deps.TplRead, deps.FormVal, deps.Audit, deps.ProfileDefaults, snapSvc)
	} else {
		svc = application.NewService(repo, deps.Presign, deps.TplRead, deps.FormVal, deps.Audit, deps.ProfileDefaults)
	}
	svc.WithRunner(db.NewTxRunner(deps.DB))
	svc.WithEligibility(repo)
	svc.WithControlledDocumentDuplicator(deps.ControlledDocumentDuplicator)
	h := dhttp.NewHandler(svc).WithCaps(deps.Caps)

	var exportHandler *dhttp.ExportHandler
	if deps.ExportPresign != nil && deps.ExportDocgen != nil {
		docgenVer := deps.DocgenVer
		if docgenVer == "" {
			docgenVer = "docgen-v2@0.4.0"
		}
		grammarVer := deps.GrammarVer
		if grammarVer == "" {
			grammarVer = "grammar-v1"
		}
		exportSvc := application.NewExportService(repo, deps.ExportPresign, deps.ExportDocgen, deps.Audit, docgenVer, grammarVer)
		exportHandler = dhttp.NewExportHandler(exportSvc)
	}

	fillInRepo := infrastructure.NewFillInRepository(deps.DB)
	fillInSvc := application.NewFillInService(db.NewTxRunner(deps.DB), application.NewSnapshotSchemaReader(deps.DB), fillInRepo).
		WithReader(fillInRepo).
		WithCDFieldReader(cdFieldReader).
		WithIAMReader(deps.IAMUserOptions)
	if deps.TemplateVersionPort != nil {
		fillInSvc.WithTemplateSchemaReader(application.NewTemplateVersionSchemaReader(deps.DB, deps.TemplateVersionPort))
	}
	fillInHandler := dhttp.NewFillInHandler(fillInSvc)
	placeholderOptionsHandler := dhttp.NewPlaceholderOptionsHandler(
		application.NewSnapshotSchemaReader(deps.DB),
		newPlaceholderOptionsIAMAdapter(deps.IAMUserOptions),
	)

	var viewHandler *dhttp.ViewHandler
	if deps.ViewPresign != nil && deps.DB != nil {
		viewSvc := application.NewViewService(db.NewTxRunner(deps.DB), deps.ViewPresign, deps.PDFOutboxReader)
		viewHandler = dhttp.NewViewHandler(viewSvc)
	}

	var reconstructHandler *dhttp.ReconstructHandler
	if deps.ReconstructRunner != nil && deps.DB != nil {
		reconstructSvc := application.NewReconstructionService(db.NewTxRunner(deps.DB), deps.ReconstructRunner, cdFieldReader)
		reconstructHandler = dhttp.NewReconstructHandler(reconstructSvc)
	}

	h.WithSubHandlers(exportHandler, fillInHandler, placeholderOptionsHandler, viewHandler, reconstructHandler)

	return &Module{
		Handler:                   h,
		Service:                   svc,
		ExportHandler:             exportHandler,
		FillInHandler:             fillInHandler,
		PlaceholderOptionsHandler: placeholderOptionsHandler,
		ViewHandler:               viewHandler,
		ReconstructHandler:        reconstructHandler,
		repo:                      repo,
	}
}

// Name identifies this publisher in boot assertion messages.
func (m *Module) Name() string { return "documents" }

// Tag is the OpenAPI tag this publisher owns.
func (m *Module) Tag() string { return "documents" }

// Mount mounts the module's HTTP handler onto mux.
func (m *Module) Mount(mux httprouter.Muxer) {
	m.Handler.Mount(mux)
}

// WithRateLimit wires the rate limiter and user-identity extractor onto the
// underlying Handler as constructor fields (Task 15a Step 3): Mount(Muxer)
// takes exactly one argument, so documents' two extra composition-root
// dependencies move here instead of a routeFamilies closure.
func (m *Module) WithRateLimit(rl *ratelimit.Middleware, userFn func(*http.Request) string) *Module {
	m.Handler.WithRateLimit(rl, userFn)
	return m
}

// Repo returns the module's underlying Repository, for wiring other modules'
// cross-module read ports at the composition root.
func (m *Module) Repo() *infrastructure.Repository { return m.repo }

type placeholderOptionsIAMAdapter struct {
	reader application.IAMUserOptionsReader
}

func newPlaceholderOptionsIAMAdapter(reader application.IAMUserOptionsReader) *placeholderOptionsIAMAdapter {
	return &placeholderOptionsIAMAdapter{reader: reader}
}

func (a *placeholderOptionsIAMAdapter) ListUserOptions(ctx context.Context, tenantID string) ([]dhttp.UserOptionView, error) {
	if a.reader == nil {
		return []dhttp.UserOptionView{}, nil
	}
	opts, err := a.reader.ListUserOptions(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]dhttp.UserOptionView, 0, len(opts))
	for _, opt := range opts {
		out = append(out, dhttp.UserOptionView{
			UserID:      opt.UserID,
			DisplayName: opt.DisplayName,
		})
	}
	return out, nil
}
