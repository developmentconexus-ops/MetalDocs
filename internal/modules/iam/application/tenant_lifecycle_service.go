// tenant_lifecycle_service.go — TenantLifecycleService (M7 F7.3 Task E,
// ADR 0070 decisions 3-5): the iam orchestrator for tenant export + crypto-
// shred erasure. Pattern anchor: OnboardTenantService (onboard_tenant_service.go)
// for the TxRunner.Do + authz.SeedTxIdentity/Require + audit.RecordTx shape;
// the erase run-side additionally follows the iam TenantDataPort's documented
// erase order (infrastructure/postgres/tenant_data_port.go) and the
// authz.BypassSystem/WithBackgroundBypass background-job bridge.
//
// Two call surfaces, matching spec item 5/6 + validation-contract.md async
// correctness row:
//   - RequestExport/RequestErase (HTTP-triggered, "enqueue side"): one tx —
//     authz.Require the matching capability, verify the target tenant exists
//     and is not already erased, INSERT tenant_lifecycle_jobs (the row IS the
//     transactional-outbox record: no separate metaldocs.outbox_events write
//     is needed because there is no further downstream fan-out consumer other
//     than the River worker this same tx enqueues), and a paired River job via
//     InsertTx sharing the same *sql.Tx (mirrors
//     render/fanout/dispatchjobs.Enqueuer.EnqueuePDFTx: business-tx dedup row
//   - River InsertTx, both inside tx — NOT the generic
//     platform/messaging/outbox/postgres pipeline, which feeds the separate
//     metaldocs-worker binary's poll loop and has no consumer registered for
//     a tenant-lifecycle event type).
//   - RunJob (River-worker-triggered, "run side"): loads the job row, fans
//     export or erase out across every registered tenantdata.Port, and
//     transitions the job to ready/failed. Idempotent by job id: a job
//     already ready is a no-op; a retried job overwrites the same artifact
//     key / re-runs the same (already-idempotent, explicit-tenant-predicate)
//     DELETEs.
package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	securitydomain "metaldocs/internal/modules/security/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/objectstore"
	"metaldocs/internal/platform/requesttrace"
	"metaldocs/internal/platform/tenantdata"
)

// Job kinds and statuses — mirrored from the tenant_lifecycle_jobs CHECK
// constraints (migration 0278). Kept as typed constants here (application
// layer) rather than in the domain package because F7.3 has not established
// a dedicated domain sub-package for tenant lifecycle; the DB CHECK is the
// enforced source of truth either way.
const (
	tenantJobKindExport = "export"
	tenantJobKindErase  = "erase"

	tenantJobStatusPending = "pending"
	tenantJobStatusRunning = "running"
	tenantJobStatusReady   = "ready"
	tenantJobStatusFailed  = "failed"
)

// tenantLookupTx is the narrow port TenantLifecycleService uses to check
// target-tenant existence/erasure state inside the enqueue tx. Declared here
// (application layer), satisfied structurally by
// *postgres.TenantRepository.TenantLookupTx — mirrors tenantRepository in
// onboard_tenant_service.go.
type tenantLookupTx interface {
	TenantLookupTx(ctx context.Context, tx *sql.Tx, tenantID string) (found, erased bool, err error)
}

// lifecycleJobInserter is the narrow tx-scoped INSERT port for
// metaldocs.tenant_lifecycle_jobs. Declared here rather than imported from
// infrastructure/postgres to keep this file's persistence surface minimal
// and independently fakeable in unit tests.
type lifecycleJobInserter interface {
	InsertLifecycleJobTx(ctx context.Context, tx *sql.Tx, jobID, tenantID, kind, requestedBy string) error
}

// lifecycleJobStore is the run-side persistence port: load a job row, and
// transition it to running/ready/failed.
type lifecycleJobStore interface {
	LoadLifecycleJob(ctx context.Context, jobID string) (iamdomain.TenantLifecycleJob, error)
	MarkLifecycleJobRunning(ctx context.Context, jobID string) error
	MarkLifecycleJobReady(ctx context.Context, jobID, objectKey string) error
	MarkLifecycleJobFailed(ctx context.Context, jobID, errMsg string) error
}

// TenantLifecycleJob is an alias for the shared domain type, kept so
// existing references to `TenantLifecycleJob` in this file (and any test
// doubles written against this package) do not need an iamdomain-qualified
// name. The canonical definition lives in iamdomain (domain/tenant_lifecycle.go)
// so infrastructure/postgres can also return it without an
// infra-imports-application inversion.
type TenantLifecycleJob = iamdomain.TenantLifecycleJob

// TenantLifecycleService implements the F7.3 export/erase orchestrator.
type TenantLifecycleService struct {
	tenants tenantLookupTx
	jobs    lifecycleJobInserter
	store   lifecycleJobStore
	river   iamdomain.TenantLifecycleEnqueuer
	runner  db.TxRunner
	audit   auditdomain.Writer
	ports   []tenantdata.Port
	store2  *sql.DB // pool handle for export SELECTs (tenantdata.Port.ExportTenantData takes *sql.DB)
	blobs   *objectstore.VerifiedStore
	crypto  securitydomain.TenantCrypto
}

// NewTenantLifecycleService constructs the service. db is the shared pool
// (used for export-side reads, off any single tx); ports is the composition
// root's registry.AllTenantDataPorts(db) slice; crypto may be nil (tenant
// crypto disabled — DestroyTenantKeyTx becomes a no-op, matching F7.2/F7.3's
// established nil-safe wiring pattern) but blobs must be non-nil for RunJob
// to do anything (callers wire it from the shared VerifiedStore instance).
func NewTenantLifecycleService(
	tenants tenantLookupTx,
	jobs lifecycleJobInserter,
	store lifecycleJobStore,
	river iamdomain.TenantLifecycleEnqueuer,
	runner db.TxRunner,
	audit auditdomain.Writer,
	pool *sql.DB,
	ports []tenantdata.Port,
	blobs *objectstore.VerifiedStore,
	crypto securitydomain.TenantCrypto,
) *TenantLifecycleService {
	return &TenantLifecycleService{
		tenants: tenants,
		jobs:    jobs,
		store:   store,
		river:   river,
		runner:  runner,
		audit:   audit,
		ports:   ports,
		store2:  pool,
		blobs:   blobs,
		crypto:  crypto,
	}
}

// RequestExport enqueues a tenant data export job. See RequestErase for the
// shared enqueue mechanics; export never checks isBackgroundContext (it never
// deletes anything) and is permitted even if a PRIOR export exists (a fresh
// job id + fresh artifact key every time — the object key is job-id keyed).
func (s *TenantLifecycleService) RequestExport(ctx context.Context, tenantID, actorID, actorTenantID string) (string, error) {
	return s.request(ctx, tenantID, actorID, actorTenantID, tenantJobKindExport, string(iamdomain.CapTenantExport))
}

// RequestErase enqueues an irreversible tenant erasure job. 409 if the
// target tenant is already erased (idempotent-refusal, not idempotent-retry:
// a second erase request against an already-erased tenant is a client error,
// distinct from RunJob's own job-id-keyed idempotent retry of the SAME job).
func (s *TenantLifecycleService) RequestErase(ctx context.Context, tenantID, actorID, actorTenantID string) (string, error) {
	return s.request(ctx, tenantID, actorID, actorTenantID, tenantJobKindErase, string(iamdomain.CapTenantErase))
}

func (s *TenantLifecycleService) request(ctx context.Context, tenantID, actorID, actorTenantID, kind, capability string) (string, error) {
	tenantID, actorID, actorTenantID, err := normalizeLifecycleRequestInput(tenantID, actorID, actorTenantID)
	if err != nil {
		return "", err
	}
	if s.runner == nil || s.tenants == nil || s.jobs == nil {
		return "", fmt.Errorf("iam: TenantLifecycleService is not fully wired")
	}

	jobID := uuid.NewString()
	params := lifecycleRequestParams{
		jobID:         jobID,
		tenantID:      tenantID,
		actorID:       actorID,
		actorTenantID: actorTenantID,
		kind:          kind,
		capability:    capability,
	}
	if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		return s.enqueueLifecycleJobTx(ctx, tx, params)
	}); err != nil {
		return "", err
	}
	return jobID, nil
}

// normalizeLifecycleRequestInput trims tenantID/actorID/actorTenantID and
// requires all three be present before request opens its enqueue tx.
func normalizeLifecycleRequestInput(tenantID, actorID, actorTenantID string) (string, string, string, error) {
	tenantID = strings.TrimSpace(tenantID)
	actorID = strings.TrimSpace(actorID)
	actorTenantID = strings.TrimSpace(actorTenantID)
	if tenantID == "" {
		return "", "", "", fmt.Errorf("%w: tenant id required", iamdomain.ErrTenantOnboardValidation)
	}
	if actorID == "" || actorTenantID == "" {
		return "", "", "", fmt.Errorf("%w: actor identity required", iamdomain.ErrTenantOnboardValidation)
	}
	return tenantID, actorID, actorTenantID, nil
}

// lifecycleRequestParams bundles the values enqueueLifecycleJobTx needs to
// run the in-tx enqueue body, so the tx closure in request stays a thin call
// into that method.
type lifecycleRequestParams struct {
	jobID         string
	tenantID      string
	actorID       string
	actorTenantID string
	kind          string
	capability    string
}

// enqueueLifecycleJobTx runs the full in-tx enqueue sequence shared by
// RequestExport/RequestErase: tier-2 authz under the ACTOR's own tenant ->
// target-tenant lookup/erased-guard -> INSERT tenant_lifecycle_jobs -> paired
// River job -> audit. Kept as a single method (not split further): the
// tripwire ordering (assert caps before the target lookup, reseed the tenant
// GUC to the TARGET before the INSERT) is one atomicity invariant that
// splitting across functions would obscure rather than clarify. ctx is the
// caller's (non-tx-scoped) context, used only for requesttrace resolution in
// buildLifecycleRequestedEvent — every DB call below derives from txCtx.
func (s *TenantLifecycleService) enqueueLifecycleJobTx(ctx context.Context, tx *sql.Tx, p lifecycleRequestParams) error {
	// Tier-2 authz FIRST (before any lookup of the target row), mirroring
	// OnboardTenantService: grants are resolved under the ACTOR's own
	// tenant, never the target tenant (the actor may hold tenant.export /
	// tenant.erase in their own tenant while the target is a different
	// tenant entirely — this is an operator-tooling capability, not a
	// same-tenant action).
	txCtx := authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(txCtx, tx, p.actorTenantID, p.actorID); err != nil {
		return fmt.Errorf("seed %s authorization: %w", p.capability, err)
	}
	if err := authz.Require(txCtx, tx, p.capability, "tenant"); err != nil {
		return fmt.Errorf("require %s authorization: %w", p.capability, err)
	}

	found, erased, err := s.tenants.TenantLookupTx(txCtx, tx, p.tenantID)
	if err != nil {
		return err
	}
	if !found {
		return iamdomain.ErrTenantNotFound
	}
	if erased {
		// Export of an erased tenant is equally meaningless (its data is
		// gone) — both kinds refuse an erased target with the same 409.
		return iamdomain.ErrTenantAlreadyErased
	}

	// tenant_lifecycle_jobs is tenant-scoped (FORCE RLS) and its INSERT
	// tripwire arm (migration 0279) requires tenant.export/tenant.erase —
	// already asserted above by Require, which appended the cap to
	// metaldocs.asserted_caps on this same tx. Re-seed the tenant GUC to
	// the TARGET tenant (not the actor's) before the INSERT, mirroring
	// OnboardTenantService's SeedTxTenant-before-provisioning-write step,
	// so RLS's WITH CHECK compares row tenant_id against the row's own
	// tenant, not the actor's.
	if err := authz.SeedTxTenant(txCtx, tx, p.tenantID); err != nil {
		return fmt.Errorf("seed target tenant: %w", err)
	}

	if err := s.jobs.InsertLifecycleJobTx(txCtx, tx, p.jobID, p.tenantID, p.kind, p.actorID); err != nil {
		return err
	}

	if s.river != nil {
		if err := s.river.EnqueueTenantLifecycleJobTx(txCtx, tx, iamdomain.TenantLifecycleJobArgs{JobID: p.jobID, TenantID: p.tenantID, JobKind: p.kind}); err != nil {
			return fmt.Errorf("enqueue river job: %w", err)
		}
	}

	if s.audit == nil {
		return nil
	}
	ev, err := buildLifecycleRequestedEvent(ctx, p.jobID, p.tenantID, p.actorID, p.kind)
	if err != nil {
		return err
	}
	return s.audit.RecordTx(txCtx, tx, ev)
}

func buildLifecycleRequestedEvent(ctx context.Context, jobID, tenantID, actorID, kind string) (auditdomain.Event, error) {
	payloadJSON, err := json.Marshal(map[string]any{
		"job_id": jobID,
		"kind":   kind,
	})
	if err != nil {
		return auditdomain.Event{}, err
	}
	return auditdomain.Event{
		ID:           "evt_" + uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      actorID,
		Action:       "tenant.lifecycle.requested",
		ResourceType: "tenant",
		ResourceID:   tenantID,
		PayloadJSON:  string(payloadJSON),
		TraceID:      requesttrace.Resolve(ctx),
		TenantID:     tenantID,
	}, nil
}

// ── Run side (River worker) ─────────────────────────────────────────────────

// RunJob executes the fan-out for jobID: export builds and writes the
// artifact; erase runs the ordered per-port DELETE fan-out, blob deletion,
// crypto-shred, and tombstone. Idempotent: a job already in status "ready" is
// a no-op (returns nil without re-running); any other status re-runs the
// fan-out and overwrites the same artifact key / re-applies the same
// (already idempotent) DELETEs.
func (s *TenantLifecycleService) RunJob(ctx context.Context, jobID string) error {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return fmt.Errorf("iam: job id required")
	}
	if s.store == nil {
		return fmt.Errorf("iam: TenantLifecycleService.RunJob is not wired (no job store)")
	}

	job, err := s.store.LoadLifecycleJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job.Status == tenantJobStatusReady {
		return nil // idempotent no-op: already completed.
	}

	if err := s.store.MarkLifecycleJobRunning(ctx, jobID); err != nil {
		return fmt.Errorf("mark job running: %w", err)
	}

	var runErr error
	var objectKey string
	switch job.Kind {
	case tenantJobKindExport:
		objectKey, runErr = s.runExport(ctx, job)
	case tenantJobKindErase:
		runErr = s.runErase(ctx, job)
	default:
		runErr = fmt.Errorf("iam: unknown tenant lifecycle job kind %q", job.Kind)
	}

	if runErr != nil {
		if err := s.store.MarkLifecycleJobFailed(ctx, jobID, runErr.Error()); err != nil {
			return fmt.Errorf("mark job failed (after run error %w): %w", runErr, err)
		}
		return runErr
	}
	if err := s.store.MarkLifecycleJobReady(ctx, jobID, objectKey); err != nil {
		return fmt.Errorf("mark job ready: %w", err)
	}
	return nil
}

// ── Export ───────────────────────────────────────────────────────────────────

// exportArtifact is the full JSON document written to
// tenants/{tenantID}/exports/tenant-export-{jobID}.json.
type exportArtifact struct {
	Header       exportHeader         `json:"header"`
	Tables       []exportTableRows    `json:"tables"`
	BlobManifest []exportBlobManifest `json:"blob_manifest"`
}

type exportHeader struct {
	TenantID string              `json:"tenant_id"`
	JobID    string              `json:"job_id"`
	Census   []exportCensusEntry `json:"census"`
}

// exportCensusEntry's Status is one of "included" (table owned by a port,
// row payload present below — possibly zero rows) or "out_of_scope" (never
// emitted today; every registered port's tables are always "included",
// reserved for a future allowlisted-but-unexported asset).
type exportCensusEntry struct {
	Table  string `json:"table"`
	Status string `json:"status"`
}

type exportTableRows struct {
	Table string            `json:"table"`
	Rows  []json.RawMessage `json:"rows"`
}

type exportBlobManifest struct {
	Key          string `json:"key"`
	ContentHash  string `json:"content_hash,omitempty"`
	SizeBytes    int64  `json:"size_bytes"`
	LastModified string `json:"last_modified"`
}

// buildExportArtifact fans ExportTenantData out across every registered
// port and assembles the completeness header (spec item 5 / consumer
// contract §2: "row manifest ... completeness header ... full census").
// Exported (lowercase-package-private in practice, but named without a
// leading underscore) as a standalone function so it is unit-testable
// without a live DB (validation gate row "artifact header/census
// construction... fakes").
func buildExportArtifact(ports []tenantdata.Port, exports map[string][]tenantdata.TableExport, tenantID, jobID string, blobs []objectstore.ObjectInfo) exportArtifact {
	var allTables []string
	for _, p := range ports {
		allTables = append(allTables, p.Tables()...)
	}
	sort.Strings(allTables)

	census := make([]exportCensusEntry, 0, len(allTables))
	tables := make([]exportTableRows, 0, len(allTables))
	for _, moduleExports := range exports {
		for _, te := range moduleExports {
			status := "included"
			tables = append(tables, exportTableRows{Table: te.Table, Rows: te.Rows})
			census = append(census, exportCensusEntry{Table: te.Table, Status: status})
		}
	}
	sort.Slice(census, func(i, j int) bool { return census[i].Table < census[j].Table })
	sort.Slice(tables, func(i, j int) bool { return tables[i].Table < tables[j].Table })

	manifest := make([]exportBlobManifest, 0, len(blobs))
	for _, b := range blobs {
		manifest = append(manifest, exportBlobManifest{
			Key:          b.Key,
			SizeBytes:    b.SizeBytes,
			LastModified: b.LastModified.UTC().Format(time.RFC3339),
		})
	}

	return exportArtifact{
		Header: exportHeader{
			TenantID: tenantID,
			JobID:    jobID,
			Census:   census,
		},
		Tables:       tables,
		BlobManifest: manifest,
	}
}

func (s *TenantLifecycleService) runExport(ctx context.Context, job TenantLifecycleJob) (string, error) {
	if s.store2 == nil {
		return "", fmt.Errorf("iam: export requires a pool *sql.DB")
	}
	exports := make(map[string][]tenantdata.TableExport, len(s.ports))
	for _, port := range s.ports {
		te, err := port.ExportTenantData(ctx, s.store2, job.TenantID)
		if err != nil {
			return "", fmt.Errorf("export tenant data (%s): %w", port.Module(), err)
		}
		exports[port.Module()] = te
	}

	var blobs []objectstore.ObjectInfo
	if s.blobs != nil {
		prefix := "tenants/" + job.TenantID + "/"
		var err error
		blobs, err = s.blobs.ListTenantObjects(ctx, job.TenantID, prefix)
		if err != nil {
			return "", fmt.Errorf("list tenant blobs: %w", err)
		}
	}

	artifact := buildExportArtifact(s.ports, exports, job.TenantID, job.ID, blobs)
	payload, err := json.Marshal(artifact)
	if err != nil {
		return "", fmt.Errorf("marshal export artifact: %w", err)
	}

	objectKey := "tenants/" + job.TenantID + "/exports/tenant-export-" + job.ID + ".json"
	if s.blobs == nil {
		return "", fmt.Errorf("iam: export requires a configured object store")
	}
	if err := s.blobs.WriteTenantObject(ctx, job.TenantID, objectKey, payload, "application/json"); err != nil {
		return "", fmt.Errorf("write export artifact: %w", err)
	}
	return objectKey, nil
}

// ── Erase ────────────────────────────────────────────────────────────────────

// eraseOrder is the cross-module FK-safe fan-out sequence (spec item 6 /
// task instructions): documents-approval -> documents -> render ->
// notifications -> controlleddocuments -> templates -> taxonomy -> tokens ->
// auth -> jobs -> audit -> iam (iam LAST: iam_users is referenced by every
// other module's rows via actor/user columns). Matched against each port's
// Module() name; a port whose Module() is not listed here runs LAST (after
// iam) in registration order as a safety net so a future port is never
// silently skipped by the fan-out (though it SHOULD be added to this list
// explicitly — see the coverage test for the anti-rot guard on Tables(),
// there is no equivalent automatic guard for ordering, which is why this
// list is reviewed by hand).
var eraseOrder = []string{
	"documents-approval",
	"documents",
	"render",
	"notifications",
	"controlleddocuments",
	"templates",
	"taxonomy",
	"tokens",
	"auth",
	"jobs",
	"audit",
	"iam",
}

func orderedPorts(ports []tenantdata.Port) []tenantdata.Port {
	rank := make(map[string]int, len(eraseOrder))
	for i, name := range eraseOrder {
		rank[name] = i
	}
	out := make([]tenantdata.Port, len(ports))
	copy(out, ports)
	sort.SliceStable(out, func(i, j int) bool {
		ri, oki := rank[out[i].Module()]
		rj, okj := rank[out[j].Module()]
		if !oki {
			ri = len(eraseOrder)
		}
		if !okj {
			rj = len(eraseOrder)
		}
		return ri < rj
	})
	return out
}

// runErase executes the three-phase erasure (spec item 6 / task
// instructions): (1) one tx erases every port's rows in FK-safe order under
// the tenant-erasure GUC + scheduler bypass token; (2) after commit, delete
// tenant blobs (network I/O, deliberately outside any DB tx); (3) a second
// small tx destroys the crypto key and tombstones the tenants row. Each
// phase is idempotent so a retried RunJob is safe: phase 1's DELETEs are all
// explicit-tenant-predicate (no-op on a second pass), phase 2's Delete is a
// no-op on a missing key (VerifiedStore.Delete treats NoSuchKey as success),
// and phase 3's DestroyTenantKeyTx / tombstone UPDATE are both no-op-safe on
// an already-destroyed/tombstoned tenant.
func (s *TenantLifecycleService) runErase(ctx context.Context, job TenantLifecycleJob) error {
	// Phase 1: row erasure, one tx, background-bypass context (H-PRE-1: this
	// is a worker-triggered background job, never an HTTP request path).
	bgCtx := authz.WithBackgroundBypass(ctx)
	if err := s.eraseTenantRowsTx(bgCtx, job); err != nil {
		return fmt.Errorf("erase rows: %w", err)
	}

	// Phase 2: blob deletion — network I/O, deliberately outside the DB tx
	// above (state-write + network never share a tx, per CLAUDE.md
	// invariants).
	if err := s.eraseTenantBlobs(ctx, job); err != nil {
		return err
	}

	// Phase 3: crypto-shred + tombstone, a second small tx.
	if err := s.cryptoShredAndTombstoneTx(bgCtx, job); err != nil {
		return fmt.Errorf("crypto-shred + tombstone: %w", err)
	}

	// Phase 4: erasure-tombstone audit event, recorded after both txs above
	// have committed (H-PRE-1).
	return s.recordErasureAudit(ctx, job)
}

// eraseTenantRowsTx runs runErase's phase 1: one tx erases every port's rows
// in FK-safe order under the tenant-erasure GUC + scheduler bypass token.
// bgCtx must already carry authz.WithBackgroundBypass.
func (s *TenantLifecycleService) eraseTenantRowsTx(bgCtx context.Context, job TenantLifecycleJob) error {
	return s.runner.Do(bgCtx, func(tx *sql.Tx) error {
		if err := authz.SeedTxTenant(bgCtx, tx, job.TenantID); err != nil {
			return fmt.Errorf("seed erasure tenant: %w", err)
		}
		// metaldocs.bypass_authz='scheduler': the tripwire's cap-assertion
		// check (enforce_capability_asserted) has no DELETE branch for most
		// erased tables, so every DELETE in this tx needs the scheduler
		// bypass token, exactly like other background fan-outs
		// (authz.BypassSystem requires WithBackgroundBypass, set above).
		if err := authz.BypassSystem(bgCtx, tx); err != nil {
			return fmt.Errorf("bypass tier-2 for erasure fan-out: %w", err)
		}
		// metaldocs.tenant_erasure: the erasure-lifecycle context GUC
		// migration 0282 added to public.reject_user_process_areas_delete()
		// — without it that trigger unconditionally rejects the
		// user_process_areas DELETE the iam port issues.
		if _, err := tx.ExecContext(bgCtx, `SELECT set_config('metaldocs.tenant_erasure', '1', true)`); err != nil {
			return fmt.Errorf("seed erasure lifecycle context: %w", err)
		}

		// Early phase (tenantdata.EarlyEraser, optional): runs BEFORE any
		// ordered port below, for the narrow set of tables that hold an
		// outbound FK into a module ranked LATER than their own owning
		// module in eraseOrder — a two-way cross-module dependency a flat
		// order cannot express in one direction alone. iam implements this
		// today for user_process_areas and capability_bindings, both FK'd
		// into taxonomy's document_process_areas (taxonomy is ranked before
		// iam) — see tenantdata.EarlyEraser's doc and the iam TenantDataPort
		// type doc for the full mechanism and the live-reproduced bug this
		// closes.
		for _, port := range s.ports {
			early, ok := port.(tenantdata.EarlyEraser)
			if !ok {
				continue
			}
			if _, err := early.EraseEarly(bgCtx, tx, job.TenantID); err != nil {
				return fmt.Errorf("erase tenant data early (%s): %w", port.Module(), err)
			}
		}

		for _, port := range orderedPorts(s.ports) {
			if _, err := port.EraseTenantData(bgCtx, tx, job.TenantID); err != nil {
				return fmt.Errorf("erase tenant data (%s): %w", port.Module(), err)
			}
		}
		// Final governance sweep: enforce_capability_asserted's scheduler-bypass
		// branch INSERTs an 'authz.bypass_used' governance_events row for every
		// armed-table DELETE in this tx, including the early phase above and
		// every ordered port's own DELETEs (modules after audit in eraseOrder,
		// notably) — both fire AFTER the audit port has already deleted
		// governance_events, leaving fresh rows for the erased tenant.
		// governance_events itself carries no tripwire trigger, so this sweep
		// (running after both phases) cannot self-feed.
		if _, err := tx.ExecContext(bgCtx, `DELETE FROM public.governance_events WHERE tenant_id = $1::uuid`, job.TenantID); err != nil {
			return fmt.Errorf("final governance_events sweep: %w", err)
		}
		return nil
	})
}

// eraseTenantBlobs runs runErase's phase 2: blob deletion. Idempotent:
// VerifiedStore.Delete treats a missing key as success, and
// ListTenantObjects on an already-emptied prefix returns an empty slice.
func (s *TenantLifecycleService) eraseTenantBlobs(ctx context.Context, job TenantLifecycleJob) error {
	if s.blobs == nil {
		return nil
	}
	prefix := "tenants/" + job.TenantID + "/"
	objs, err := s.blobs.ListTenantObjects(ctx, job.TenantID, prefix)
	if err != nil {
		return fmt.Errorf("list tenant blobs for erasure: %w", err)
	}
	for _, obj := range objs {
		if err := s.blobs.Delete(ctx, obj.Key); err != nil {
			return fmt.Errorf("delete tenant blob %s: %w", obj.Key, err)
		}
	}
	return nil
}

// cryptoShredAndTombstoneTx runs runErase's phase 3. Runs even if crypto is
// disabled (s.crypto nil) — DestroyTenantKeyTx becomes a no-op call skipped
// entirely below, matching the established nil-safe pattern; the tombstone
// UPDATE still must happen. bgCtx must already carry
// authz.WithBackgroundBypass.
func (s *TenantLifecycleService) cryptoShredAndTombstoneTx(bgCtx context.Context, job TenantLifecycleJob) error {
	return s.runner.Do(bgCtx, func(tx *sql.Tx) error {
		if s.crypto != nil {
			if err := s.crypto.DestroyTenantKeyTx(bgCtx, tx, job.TenantID); err != nil {
				return fmt.Errorf("destroy tenant key: %w", err)
			}
		}
		// Tombstone: metaldocs.tenants carries no UPDATE tripwire arm (the
		// arm is INSERT-only, migration 0279) so this UPDATE needs no
		// bypass token. slug is suffixed with the tenant id to satisfy the
		// tenants_slug_key uniqueness constraint across repeated erasures of
		// distinct tenants (each gets a distinct "erased-{id}" slug).
		if _, err := tx.ExecContext(bgCtx, `
UPDATE metaldocs.tenants
   SET name = 'erased', slug = 'erased-' || id::text, erased_at = COALESCE(erased_at, now())
 WHERE id = $1::uuid
`, job.TenantID); err != nil {
			return fmt.Errorf("tombstone tenant: %w", err)
		}
		return nil
	})
}

// recordErasureAudit runs runErase's phase 4. H-PRE-1: the erasure-tombstone
// audit event is recorded AFTER commit, off any lock-holding tx, via Record
// (its own new tx) — never RecordTx sharing either fan-out tx above. Written
// plaintext by design: the tenant's own DEK is already destroyed, so an
// enveloped write here would be unreadable the instant it lands;
// audit.RecordTx/Record never receive a TenantCrypto themselves for this
// call (composition-root wiring only touches the audit WRITER's own
// envelope decision, not this orchestrator's).
func (s *TenantLifecycleService) recordErasureAudit(ctx context.Context, job TenantLifecycleJob) error {
	if s.audit == nil {
		return nil
	}
	payloadJSON, err := json.Marshal(map[string]any{
		"job_id":    job.ID,
		"tenant_id": job.TenantID,
	})
	if err != nil {
		return fmt.Errorf("marshal erasure audit payload: %w", err)
	}
	if err := s.audit.Record(ctx, auditdomain.Event{
		ID:           "evt_" + uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      job.RequestedBy,
		Action:       "tenant.erased",
		ResourceType: "tenant",
		ResourceID:   job.TenantID,
		PayloadJSON:  string(payloadJSON),
		TraceID:      requesttrace.Resolve(ctx),
		TenantID:     job.TenantID,
	}); err != nil {
		return fmt.Errorf("record erasure audit event: %w", err)
	}
	return nil
}

// isTenantNotFoundOrErased is a small helper handlers can use if they need
// to distinguish the two enqueue-side sentinel errors from a generic
// failure; kept here (not exported error vars in this file) because the
// canonical error values already live in iamdomain (ErrTenantNotFound /
// ErrTenantAlreadyErased).
var _ = errors.Is // silence unused-import if the helper above is trimmed later
