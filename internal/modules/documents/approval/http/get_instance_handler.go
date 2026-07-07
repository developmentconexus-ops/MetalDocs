package approvalhttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	"metaldocs/internal/modules/documents/approval/infrastructure"
)

// GetInstanceHandler returns a single approval instance by ID, with an ETag
// header set to the document's current revision_version for OCC on subsequent writes.
func (h *Handler) GetInstanceHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	instanceID := r.PathValue("instance_id")

	if h.readSvc == nil {
		WriteError(w, errors.New("read service not configured"))
		return
	}

	inst, err := h.readSvc.LoadInstance(r.Context(), h.runner, tenantID, instanceID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrNoActiveInstance) {
			WriteError(w, infrastructure.ErrNoActiveInstance)
			return
		}
		WriteError(w, err)
		return
	}

	resp, err := h.mapInstanceResponse(r.Context(), tenantID, inst)
	if err != nil {
		WriteError(w, err)
		return
	}
	w.Header().Set("ETag", resp.ETag)
	WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) mapInstanceResponse(ctx context.Context, tenantID string, inst *domain.Instance) (contracts.InstanceResponse, error) {
	var completedAt *string
	if inst.CompletedAt != nil {
		v := inst.CompletedAt.UTC().Format(time.RFC3339)
		completedAt = &v
	}

	eligibleNames, err := h.resolveEligibleActorNames(ctx, tenantID, inst)
	if err != nil {
		return contracts.InstanceResponse{}, err
	}

	stages := make([]contracts.StageInstance, len(inst.Stages))
	for i, s := range inst.Stages {
		recs := make([]contracts.SignoffRecord, 0, len(s.Signoffs))
		for _, sig := range s.Signoffs {
			name := sig.ActorDisplayNameSnapshot()
			if name == "" {
				name = sig.ActorUserID()
			}
			recs = append(recs, contracts.SignoffRecord{
				ID:               sig.ID(),
				ActorUserID:      name,
				Decision:         contracts.SignoffDecision(sig.Decision()),
				Reason:           sig.Comment(),
				SignatureMethod:  contracts.SignatureMethod(sig.SignatureMethod()),
				SignedAt:         sig.SignedAt().UTC().Format(time.RFC3339),
				SignatureMeaning: sig.SignatureMeaning(),
			})
		}
		stages[i] = contracts.StageInstance{
			ID:         s.ID,
			StageIndex: s.StageOrder,
			Label:      s.NameSnapshot,
			Status:     mapStageStatus(s.Status),
			Signoffs:   recs,
			Actors:     buildStageActors(s, eligibleNames),
		}
	}

	return contracts.InstanceResponse{
		ID:          inst.ID,
		DocumentID:  inst.DocumentID,
		RouteID:     inst.RouteID,
		TenantID:    inst.TenantID,
		Status:      contracts.InstanceStatus(inst.Status),
		SubmittedBy: inst.SubmittedBy,
		SubmittedAt: inst.SubmittedAt.UTC().Format(time.RFC3339),
		CompletedAt: completedAt,
		Stages:      stages,
		ETag:        instanceETag(inst.RevisionVersion),
	}, nil
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
		displayName := sig.ActorDisplayNameSnapshot()
		if displayName == "" {
			displayName = eligibleNames[sig.ActorUserID()]
		}
		if displayName == "" {
			displayName = sig.ActorUserID()
		}
		actors = append(actors, contracts.StageActor{
			UserID:      sig.ActorUserID(),
			DisplayName: displayName,
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
