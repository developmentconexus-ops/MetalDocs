package application

import (
	"context"
	"database/sql"
	"fmt"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
)

// resolveCurrentEligibleForDrift computes the live eligible set used by the
// drift evaluation of an active stage under Option C′ (unit 3.2 slice 6a). Pool
// selectors (role_in_fixed_area, role_in_document_area) re-resolve against
// current membership — role_in_document_area against the subject's CURRENT
// area (the "area follows the document" property) — while named_user selectors
// (the drift-exempt class, incl. materialized submit_choice picks) contribute
// their frozen ids UNCONDITIONALLY and are never re-filtered. Callers invoke it
// only when the stage's drift policy is not keep_snapshot. Single-tx,
// non-recording reads (H-PRE-1): no authz.Require inside.
func resolveCurrentEligibleForDrift(
	ctx context.Context,
	tx *sql.Tx,
	repo infrastructure.ApprovalRepository,
	cdRead controlleddocumentsdomain.CDFieldReader,
	tenantID string,
	stage domain.StageInstance,
	subject domain.Subject,
) ([]string, error) {
	var poolSelectors []domain.ActorSelector
	var exemptIDs []string
	needsLiveArea := false
	for _, sel := range stage.SelectorsSnapshot {
		switch sel.Kind {
		case domain.SelectorNamedUser:
			exemptIDs = append(exemptIDs, sel.UserID)
		case domain.SelectorRoleInFixedArea:
			poolSelectors = append(poolSelectors, sel)
		case domain.SelectorRoleInDocumentArea:
			poolSelectors = append(poolSelectors, sel)
			needsLiveArea = true
		}
	}
	liveArea := ""
	if needsLiveArea {
		a, err := resolveSubjectAreaCode(ctx, tx, cdRead, tenantID, subject)
		if err != nil {
			return nil, fmt.Errorf("drift: resolve live subject area: %w", err)
		}
		liveArea = a
	}
	poolIDs, err := repo.ResolveEligibleActorsForSelectors(ctx, tx, tenantID, poolSelectors, liveArea)
	if err != nil {
		return nil, fmt.Errorf("drift: resolve pool selectors: %w", err)
	}
	return unionUserIDs(poolIDs, exemptIDs), nil
}
