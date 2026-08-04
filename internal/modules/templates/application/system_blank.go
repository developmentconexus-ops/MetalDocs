package application

import (
	"context"
	"fmt"

	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/objectstore"
)

// The system blank template's identity. It is a real, system-owned template
// row seeded by db/reference-data (tenant ffffffff-…, template …0101, version
// 1, status 'published'), whose docx_storage_key/content_hash pair names the
// deterministic asset the deployment seeds at system/templates/blank.docx.
//
// These constants live HERE, in the owning module's application layer, and are
// consumed by delivery/http rather than re-declared there: the repo's named
// meta-defect is hand-synced enumerations, and two copies of a well-known UUID
// are exactly that. This is the only Go-side declaration of the identity — the
// storage KEY and the HASH are deliberately NOT declared in Go at all; they are
// read from the row (ADR 0088 §3), so the asset and its hash have a single
// source of truth in db/reference-data.
const (
	// SystemBlankTemplateTenantID is the system tenant that owns the blank
	// template. Matches objectstore.SystemTenantID.
	SystemBlankTemplateTenantID = objectstore.SystemTenantID
	// SystemBlankTemplateID is the fixed id of the system blank template row.
	SystemBlankTemplateID = "00000000-0000-0000-0000-000000000101"
	// SystemBlankTemplateVersionNumber is the version whose object every
	// from-scratch template is materialized from.
	SystemBlankTemplateVersionNumber = 1
)

// materializeFromSystemBlank stores the bytes a from-scratch template version
// must point at, and returns the VERIFIED hash of what now lives at dstKey.
//
// ADR 0088 §2/§3 — store-then-reference, PRE-TX: the object exists before the
// referencing row commits, so the only crash outcome is a safe orphan object
// (never a row without content). This is the same shape spawnNextDraft already
// used; blank creation was the one version-creation path that skipped it.
//
// The source key and its expected hash are read from the system blank
// template's own version row via the module's own repository — no hardcoded
// hash in Go, no cross-module edge, and no second generator of "what an empty
// template is" (the docx-renderer is deliberately not involved).
//
// Fails CLOSED on every branch: a missing row, a row without a real 64-hex
// hash, a missing object, or a hash that does not match the pinned one all
// abort creation with an error. There is no degraded path that produces a
// content-less version — that is the state this ADR exists to make unreachable.
func (s *Service) materializeFromSystemBlank(ctx context.Context, tenantID, dstKey string) (string, error) {
	src, err := s.repo.GetSystemBlankVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("templates: load system blank template source: %w: %w", domain.ErrContentMaterializationFailed, err)
	}
	if src.DocxStorageKey == "" || len(src.ContentHash) != 64 {
		return "", fmt.Errorf(
			"templates: system blank template version is not materialized (docx_storage_key=%q, content_hash length=%d): %w",
			src.DocxStorageKey, len(src.ContentHash), domain.ErrContentMaterializationFailed,
		)
	}
	return s.copyAndConfirm(ctx, tenantID, src.DocxStorageKey, dstKey, src.ContentHash)
}

// copyAndConfirm is the shared store-then-reference primitive behind every
// template-version creation path: copy srcKey's bytes to the new version's OWN
// canonical key, then Confirm that key against the source's hash so the value
// the row is born with is the VERIFIED hash of the object it points at — read
// back from the store, never assumed from the source row.
//
// Confirm doubles as the integrity gate on the copy: a truncated or wrong
// object yields ErrHashMismatch (and VerifiedStore deletes the bad copy) rather
// than a row that claims a hash its bytes do not have.
//
// EVERY failure here is ErrContentMaterializationFailed, deliberately. Both
// callers copy bytes the server already owns — the system blank asset, or the
// previous version's object — so no user input reaches this function and no
// failure inside it is a user error. The objectstore sentinel that actually
// fired is preserved in the wrapped chain for logs; it is not promoted to the
// classification, because no caller can act on the difference. Classifying
// these as the user-facing upload sentinels (as the code did before ADR 0088)
// told the user to fix an upload they never made, and hid an object-store
// failure behind a 4xx.
func (s *Service) copyAndConfirm(ctx context.Context, tenantID, srcKey, dstKey, expectedHash string) (string, error) {
	if err := s.presign.Copy(ctx, tenantID, srcKey, dstKey); err != nil {
		return "", fmt.Errorf("templates: copy docx %s -> %s: %w: %w", srcKey, dstKey, domain.ErrContentMaterializationFailed, err)
	}
	vp, err := s.presign.Confirm(ctx, tenantID, dstKey, expectedHash)
	if err != nil {
		return "", fmt.Errorf("templates: confirm copied docx %s: %w: %w", dstKey, domain.ErrContentMaterializationFailed, err)
	}
	if len(vp.ContentHash) != 64 {
		return "", fmt.Errorf("templates: confirm copied docx %s returned a non-sha256 hash %q: %w", dstKey, vp.ContentHash, domain.ErrContentMaterializationFailed)
	}
	return vp.ContentHash, nil
}
