package backfill

import (
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"

	approvaljobs "metaldocs/internal/modules/approval/jobs"
	docapp "metaldocs/internal/modules/documents/application"
	docrepo "metaldocs/internal/modules/documents/infrastructure"
	fanoutpkg "metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/fanout/dispatchjobs"
	"metaldocs/internal/modules/render/resolvers"
	"metaldocs/internal/platform/db"
)

// Wire builds the production object graph the per-document core needs, using
// the SAME constructors apps/worker and apps/api use — no parallel freeze
// implementation, no hand-rolled outbox writer.
//
// riverClient must be a client bound to the same *sql.DB, so the evaluation job
// and the outbox rows commit in one transaction.
//
// The FreezeService is constructed with its real dependencies even though
// RepairMaterialization only exercises the materialize outbox: narrowing the
// constructor for one caller would fork the freeze wiring, and forking it is
// exactly how the two paths drift.
func Wire(database *sql.DB, riverClient *river.Client[*sql.Tx]) (Deps, error) {
	if database == nil {
		return Deps{}, fmt.Errorf("backfill: wire requires a non-nil *sql.DB")
	}
	if riverClient == nil {
		return Deps{}, fmt.Errorf("backfill: wire requires a non-nil river client")
	}

	snapRepo := docrepo.NewSnapshotRepository(database)
	fillInRepo := docrepo.NewFillInRepository(database)
	schemaReader := docapp.NewSnapshotSchemaReader(database)
	resolverReg := resolvers.NewRegistry()
	resolvers.RegisterBuiltins(resolverReg)

	// nil ResolverContextBuilder / nil fanout client mirror apps/worker's
	// materialize wiring for the paths this tool never walks: the backfill
	// never resolves placeholders and never calls docx-renderer itself.
	freeze := docapp.NewFreezeService(schemaReader, fillInRepo, fillInRepo, resolverReg, snapRepo, nil, snapRepo, nil)

	pdfOutboxRepo := fanoutpkg.NewPDFOutboxRepository(database)
	materializeOutboxRepo := fanoutpkg.NewMaterializeOutboxRepository(database)
	// maxAttempts 0 = "let River's own default apply" (NewEnqueuer contract).
	dispatchEnqueuer := dispatchjobs.NewEnqueuer(riverClient, pdfOutboxRepo, materializeOutboxRepo, 0)
	freeze = freeze.WithMaterializeOutbox(dispatchEnqueuer)

	return Deps{
		Runner:   db.NewTxRunner(database),
		Freeze:   freeze,
		Evaluate: approvaljobs.NewReleaseEvaluationEnqueuer(riverClient),
	}, nil
}
