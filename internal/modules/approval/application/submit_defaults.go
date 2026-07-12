package application

import (
	"context"

	"metaldocs/internal/platform/db"
)

// SubmitDefaultsResolver is the narrow in-tx read port SubmitService uses to
// resolve server-authoritative submit prerequisites when the client omits them
// (ADR 0073 §2 — canonical /submit owns in-tx prereq resolution). It is kept
// separate from the broad infrastructure.ApprovalRepository (ISP): the submit
// service asks only for what the wrapper-era GetFinalizePrereqs read, now moved
// inside the submit transaction to close the TOCTOU.
//
// Every method is a plain, non-recording SELECT executed on the caller's
// transaction (HS-PRE-1): no authz recording, no advisory lock.
type SubmitDefaultsResolver interface {
	// LoadControlledDocumentID returns documents.controlled_document_id for
	// (tenantID, documentID). ok is false when the document row is absent or the
	// controlled-document link is NULL. Tenant-scoped, in-tx.
	LoadControlledDocumentID(ctx context.Context, tx db.Tx, tenantID, documentID string) (cdID string, ok bool, err error)

	// LoadActiveRouteIDByProfile returns the id of the single active approval
	// route for (tenantID, profileCode), newest route version first (mirrors the
	// wrapper's repository.go:1801-1817 query, now in-tx). Returns
	// infrastructure.ErrNoActiveApprovalRoute when no active route exists — the
	// submit service maps that to documents/domain.ErrApprovalRouteMissing.
	// Document specialization of LoadActiveRouteIDBySubject (M3 P3.S2b-2).
	LoadActiveRouteIDByProfile(ctx context.Context, tx db.Tx, tenantID, profileCode string) (routeID string, err error)

	// LoadActiveRouteIDBySubject returns the id of the single active approval
	// route for (tenantID, subjectKind, subjectKey), newest route version
	// first — the subject-generic form (M3 kernel extraction, ADR 0082,
	// P3.S2b-2) that makes template routes selectable. Returns
	// infrastructure.ErrNoActiveApprovalRoute when no active route exists.
	LoadActiveRouteIDBySubject(ctx context.Context, tx db.Tx, tenantID, subjectKind, subjectKey string) (routeID string, err error)

	// LoadHeadContentHash returns the most recent autosaved revision's
	// content_hash for the document, or "" when no revision has a hash yet
	// (mirrors the wrapper's repository.go:1819-1828 COALESCE / no-rows
	// tolerance). Tenant-scoped, in-tx.
	LoadHeadContentHash(ctx context.Context, tx db.Tx, tenantID, documentID string) (string, error)
}
