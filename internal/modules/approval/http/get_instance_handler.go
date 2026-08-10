package approvalhttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/http/contracts"
	"metaldocs/internal/modules/approval/infrastructure"
)

// GetInstanceHandler returns a single approval instance by ID, with an ETag
// header set to the document's current revision_version for OCC on subsequent writes.
func (h *Handler) GetInstanceHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	instanceID := r.PathValue("instance_id")

	if h.readSvc == nil {
		WriteError(w, r, errors.New("read service not configured"))
		return
	}

	inst, err := h.readSvc.LoadInstance(r.Context(), h.runner, tenantID, instanceID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrNoActiveInstance) {
			WriteError(w, r, infrastructure.ErrNoActiveInstance)
			return
		}
		WriteError(w, r, err)
		return
	}

	resp, err := h.mapInstanceResponse(r.Context(), tenantID, inst, nil, nil)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	w.Header().Set("ETag", resp.ETag)
	WriteJSON(w, r, http.StatusOK, resp)
}

// mapInstanceResponse builds the wire InstanceResponse from a loaded domain
// instance. viewer is nil for the by-id read (GetInstanceHandler, not scoped
// to a single viewer's workspace); non-nil for the by-document view
// (GetInstanceByDocumentHandler), which projects viewer's ViewerFacts onto
// the response (F2d.1, ADR 0078). The delegator display name (when
// viewer.ViaDelegationFromUserID is set) is resolved OFF-TX here via the
// same displayNameReader port resolveEligibleActorNames uses (H-PRE-1).
func (h *Handler) mapInstanceResponse(ctx context.Context, tenantID string, inst *domain.Instance, viewer *domain.ViewerFacts, verdicts []domain.ReviewVerdict) (contracts.InstanceResponse, error) {
	var completedAt *string
	if inst.CompletedAt != nil {
		v := inst.CompletedAt.UTC().Format(time.RFC3339)
		completedAt = &v
	}

	eligibleNames, err := h.resolveEligibleActorNames(ctx, tenantID, inst)
	if err != nil {
		return contracts.InstanceResponse{}, err
	}
	stages := buildStageInstanceContracts(inst.Stages, eligibleNames)

	viewerContract, err := h.buildViewerContract(ctx, tenantID, viewer)
	if err != nil {
		return contracts.InstanceResponse{}, err
	}

	// Verdict history (F2d.2, ADR 0079): projected only on the by-document view
	// (viewer != nil), matching the OpenAPI schema split. Always a non-nil slice
	// there so an instance with no verdicts serializes as [] (empty ⇒ []), never
	// omitted. actor_display_name is the immutable snapshot — no fallback.
	var verdictsContract *[]contracts.VerdictRecord
	if viewer != nil {
		verdictsContract = buildVerdictRecordContracts(verdicts)
	}

	return contracts.InstanceResponse{
		ID:                inst.ID,
		DocumentID:        inst.DocumentID,
		RouteID:           inst.RouteID,
		TenantID:          inst.TenantID,
		Status:            contracts.InstanceStatus(inst.Status),
		SubmittedBy:       inst.SubmittedBy,
		SubmittedAt:       inst.SubmittedAt.UTC().Format(time.RFC3339),
		CompletedAt:       completedAt,
		Stages:            stages,
		ETag:              instanceETag(inst.RevisionVersion),
		FrozenContentHash: inst.FrozenContentHash,
		Viewer:            viewerContract,
		Verdicts:          verdictsContract,
	}, nil
}

// buildStageInstanceContracts maps an instance's domain stages onto their
// wire StageInstance shape, resolving each stage's pending/signed actor list
// via eligibleNames.
func buildStageInstanceContracts(domainStages []domain.StageInstance, eligibleNames map[string]string) []contracts.StageInstance {
	stages := make([]contracts.StageInstance, len(domainStages))
	for i, s := range domainStages {
		recs := make([]contracts.SignoffRecord, 0, len(s.Signoffs))
		for _, sig := range s.Signoffs {
			// Snapshot-only (D2, ADR 0079): the signed actor's name is the
			// immutable value cast at sign time. The column is DB-enforced
			// non-empty (migration 0294), so no id/live fallback is needed.
			recs = append(recs, contracts.SignoffRecord{
				ID:               sig.ID(),
				ActorUserID:      sig.ActorDisplayNameSnapshot(),
				Decision:         contracts.SignoffDecision(sig.Decision()),
				Reason:           sig.Comment(),
				SignatureMethod:  contracts.SignatureMethod(sig.SignatureMethod()),
				SignedAt:         sig.SignedAt().UTC().Format(time.RFC3339),
				SignatureMeaning: sig.SignatureMeaning(),
			})
		}
		var dueAt *string
		if s.DueAt != nil {
			v := s.DueAt.UTC().Format(time.RFC3339)
			dueAt = &v
		}
		stages[i] = contracts.StageInstance{
			ID:         s.ID,
			StageIndex: s.StageOrder,
			Label:      s.NameSnapshot,
			Status:     mapStageStatus(s.Status),
			Signoffs:   recs,
			Actors:     buildStageActors(s, eligibleNames),
			StageKind:  string(s.Kind),
			DueAt:      dueAt,
		}
	}
	return stages
}

// buildViewerContract projects viewer's ViewerFacts onto the wire shape,
// resolving the delegator's display name off-tx when the viewer signed via
// delegation. Returns (nil, nil) when viewer is nil (the by-id read, not
// scoped to a single viewer's workspace).
func (h *Handler) buildViewerContract(ctx context.Context, tenantID string, viewer *domain.ViewerFacts) (*contracts.ViewerFacts, error) {
	if viewer == nil {
		return nil, nil
	}
	vf := contracts.ViewerFacts{
		IsAuthor:               viewer.IsAuthor,
		EligibleForActiveStage: viewer.EligibleForActiveStage,
		HasSignedActiveStage:   viewer.HasSignedActiveStage,
	}
	if viewer.ViaDelegationFromUserID != "" {
		displayName := viewer.ViaDelegationFromUserID
		names, err := h.displayNameReader.DisplayNames(ctx, tenantID, []string{viewer.ViaDelegationFromUserID})
		if err != nil {
			return nil, fmt.Errorf("resolve viewer delegation display name: %w", err)
		}
		if name, ok := names[viewer.ViaDelegationFromUserID]; ok && name != "" {
			displayName = name
		}
		vf.ViaDelegationFrom = &contracts.ViewerDelegationFrom{
			UserID:      viewer.ViaDelegationFromUserID,
			DisplayName: displayName,
		}
	}
	return &vf, nil
}

// buildVerdictRecordContracts maps an instance's review-verdict history onto
// its wire shape. Always returns a non-nil slice pointer so an instance with
// no verdicts serializes as [] (empty ⇒ []), never omitted.
func buildVerdictRecordContracts(verdicts []domain.ReviewVerdict) *[]contracts.VerdictRecord {
	recs := make([]contracts.VerdictRecord, 0, len(verdicts))
	for i := range verdicts {
		v := &verdicts[i]
		var reason *string
		if c := v.Comment(); c != "" {
			reason = &c
		}
		recs = append(recs, contracts.VerdictRecord{
			ID:               v.ID(),
			StageInstanceID:  v.StageInstanceID(),
			ActorUserID:      v.ActorUserID(),
			ActorDisplayName: v.ActorDisplayNameSnapshot(),
			Verdict:          string(v.Verdict()),
			Reason:           reason,
			VerdictAt:        v.VerdictAt().UTC().Format(time.RFC3339),
		})
	}
	return &recs
}

func (h *Handler) resolveEligibleActorNames(ctx context.Context, tenantID string, inst *domain.Instance) (map[string]string, error) {
	names := make(map[string]string)
	for _, stage := range inst.Stages {
		for _, sig := range stage.Signoffs {
			if displayName := sig.ActorDisplayNameSnapshot(); displayName != "" {
				names[sig.ActorUserID()] = displayName
			}
		}
	}
	actorSet := make(map[string]struct{})
	for _, stage := range inst.Stages {
		for _, actorID := range stage.EligibleActorIDs {
			if _, ok := names[actorID]; ok {
				continue
			}
			actorSet[actorID] = struct{}{}
		}
	}
	if len(actorSet) == 0 {
		return names, nil
	}
	actorIDs := make([]string, 0, len(actorSet))
	for actorID := range actorSet {
		actorIDs = append(actorIDs, actorID)
	}
	// M4/F4.1: use the iam-owned port instead of raw iam_users SQL.
	// The port omits absent/empty-display_name users; the post-loop fallback
	// (missing → userID) reproduces the prior COALESCE(NULLIF(…),'') behavior.
	portNames, err := h.displayNameReader.DisplayNames(ctx, tenantID, actorIDs)
	if err != nil {
		return nil, fmt.Errorf("resolve eligible actors: %w", err)
	}
	for actorID, displayName := range portNames {
		names[actorID] = displayName
	}
	for _, actorID := range actorIDs {
		if _, ok := names[actorID]; !ok {
			names[actorID] = actorID
		}
	}
	return names, nil
}

func buildStageActors(stage domain.StageInstance, eligibleNames map[string]string) []contracts.StageActor {
	actors := make([]contracts.StageActor, 0, len(stage.Signoffs)+len(stage.EligibleActorIDs))
	seen := make(map[string]struct{}, len(stage.Signoffs))
	for _, sig := range stage.Signoffs {
		decision := contracts.SignoffDecision(sig.Decision())
		status := "approved"
		if sig.Decision() == domain.DecisionReject {
			status = "rejected"
		}
		// Snapshot-only for a signed actor (D2, ADR 0079): DB-enforced non-empty.
		actors = append(actors, contracts.StageActor{
			UserID:      sig.ActorUserID(),
			DisplayName: sig.ActorDisplayNameSnapshot(),
			Status:      status,
			Decision:    &decision,
		})
		seen[sig.ActorUserID()] = struct{}{}
	}

	pendingStatus := ""
	switch stage.Status {
	case domain.StageActive:
		pendingStatus = "active"
	case domain.StagePending:
		pendingStatus = "waiting"
	case domain.StageCompleted, domain.StageSkipped, domain.StageRejectedHere:
		// Terminal stage statuses never contribute a pending-actor entry —
		// every eligible actor is already accounted for via signoffs above.
	}
	if pendingStatus == "" {
		return actors
	}
	for _, actorID := range stage.EligibleActorIDs {
		if _, ok := seen[actorID]; ok {
			continue
		}
		displayName := eligibleNames[actorID]
		if displayName == "" {
			displayName = actorID
		}
		actors = append(actors, contracts.StageActor{
			UserID:      actorID,
			DisplayName: displayName,
			Status:      pendingStatus,
		})
	}
	return actors
}

func mapStageStatus(s domain.StageStatus) string {
	switch s {
	case domain.StageCompleted:
		return "passed"
	case domain.StageRejectedHere:
		return "failed"
	case domain.StageSkipped:
		return "cancelled"
	default:
		return string(s)
	}
}
