package jobs

import (
	"context"
	"log/slog"
	"time"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/repository"
	"metaldocs/internal/platform/objectstore"
)

// OrphanObjectRepo is the narrow read surface the sweeper needs from the
// templates repository: which tenants own templates, and which object keys each
// tenant's template versions reference. Defined here (consumer side) so the
// sweeper depends on an interface, not the concrete repository, and stays
// unit-testable without a database. *repository.Repository satisfies it.
type OrphanObjectRepo interface {
	TenantIDsWithTemplates(ctx context.Context) ([]string, error)
	ReferencedTemplateObjectRefs(ctx context.Context, tenantID string) ([]repository.ReferencedTemplateObjectRef, error)
}

// OrphanObjectStore is the narrow object-store surface the sweeper needs:
// tenant-guarded enumeration under a prefix, and key deletion. Both are existing
// VerifiedStore primitives; *objectstore.VerifiedStore satisfies it.
type OrphanObjectStore interface {
	ListTenantObjects(ctx context.Context, tenantID, prefix string) ([]objectstore.ObjectInfo, error)
	Delete(ctx context.Context, key string) error
}

// StartTemplateOrphanSweeper launches a background reconciliation janitor that
// deletes orphaned template objects left in the object store by spawnNextDraft's
// pre-tx Copy when the surrounding transaction rolls back (or the process
// crashes between the Copy and the commit). For each tenant that owns templates,
// it enumerates the objects under tenants/{tenantID}/templates/ and deletes any
// object that is (a) older than maxAge AND (b) absent from the set of keys the
// tenant's template versions reference.
//
// Data-safety invariants (a referenced key MUST NEVER be deleted):
//   - The referenced set is the UNION of every stored docx_storage_key and the
//     DERIVED schema key for every template-version row (the schema key is not a
//     DB column — see application.TemplateVersionSchemaKey). Missing either class
//     would delete live objects.
//   - The age threshold (maxAge) protects an in-flight spawn whose Copy has run
//     but whose tx has not yet committed: its object is young, so it survives.
//   - Enumeration and deletion are scoped to one tenant's prefix via the
//     VerifiedStore tenant guard, so the sweep can never cross the tenant
//     boundary (multi-tenant invariant).
//
// Mirrors documents/jobs/orphan_pending_sweeper.go: a ticker goroutine running
// under authz.WithBackgroundBypass, cancellable via the returned stop func.
func StartTemplateOrphanSweeper(ctx context.Context, repo OrphanObjectRepo, store OrphanObjectStore, interval, maxAge time.Duration) (stop func()) {
	// Background root: mark the context as a fail-closed background bypass, off any
	// HTTP path (ADR 0022 Phase 7, CWE-269), matching the documents sibling.
	ctx = authz.WithBackgroundBypass(ctx)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				deleted, err := sweepOnce(ctx, repo, store, maxAge)
				if err != nil {
					slog.Warn("template_orphan_sweeper error", "err", err)
					continue
				}
				if deleted > 0 {
					slog.Info("template_orphan_sweeper deleted", "deleted", deleted)
				}
			}
		}
	}()
	return cancel
}

// sweepOnce runs a single reconciliation pass over every tenant that owns
// templates and returns the number of orphan objects deleted. Extracted from the
// ticker loop so it is directly testable.
func sweepOnce(ctx context.Context, repo OrphanObjectRepo, store OrphanObjectStore, maxAge time.Duration) (int, error) {
	cutoff := time.Now().UTC().Add(-maxAge)
	tenantIDs, err := repo.TenantIDsWithTemplates(ctx)
	if err != nil {
		return 0, err
	}
	deleted := 0
	for _, tenantID := range tenantIDs {
		n, err := sweepTenant(ctx, repo, store, tenantID, cutoff)
		if err != nil {
			// Skip this tenant but keep reconciling the rest; a single tenant's
			// store/DB hiccup must not stall the whole sweep.
			slog.Warn("template_orphan_sweeper tenant error", "tenant_id", tenantID, "err", err)
			continue
		}
		deleted += n
	}
	return deleted, nil
}

// sweepTenant reconciles one tenant: builds the referenced-key set from the DB,
// lists the tenant's template objects, and deletes the aged-out orphans.
func sweepTenant(ctx context.Context, repo OrphanObjectRepo, store OrphanObjectStore, tenantID string, cutoff time.Time) (int, error) {
	refs, err := repo.ReferencedTemplateObjectRefs(ctx, tenantID)
	if err != nil {
		return 0, err
	}
	// Referenced set = (all stored docx keys) ∪ (derived schema key per version).
	// Both classes are kept; never delete a key in this set.
	referenced := make(map[string]struct{}, len(refs)*2)
	for _, ref := range refs {
		referenced[ref.DocxStorageKey] = struct{}{}
		referenced[application.TemplateVersionSchemaKey(tenantID, ref.TemplateID, ref.VersionNumber)] = struct{}{}
	}

	objects, err := store.ListTenantObjects(ctx, tenantID, application.TenantTemplatePrefix(tenantID))
	if err != nil {
		return 0, err
	}

	deleted := 0
	for _, obj := range objects {
		if _, ok := referenced[obj.Key]; ok {
			continue // referenced — never delete
		}
		if !obj.LastModified.Before(cutoff) {
			continue // younger than maxAge — protects an in-flight spawn
		}
		if err := store.Delete(ctx, obj.Key); err != nil {
			slog.Warn("template_orphan_sweeper delete failed", "tenant_id", tenantID, "key", obj.Key, "err", err)
			continue
		}
		slog.Info("template_orphan_sweeper deleted object", "tenant_id", tenantID, "key", obj.Key)
		deleted++
	}
	return deleted, nil
}
