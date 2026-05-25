package approvalhttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/lib/pq"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	"metaldocs/internal/modules/documents/approval/repository"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

func (h *Handler) GetInstanceHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := iamdomain.UserIDFromContext(r.Context())
	instanceID := r.PathValue("instance_id")

	if h.readSvc == nil {
		WriteError(w, errors.New("read service not configured"))
		return
	}

	inst, err := h.readSvc.LoadInstance(r.Context(), h.db, tenantID, actorID, instanceID)
	if err != nil {
		if errors.Is(err, repository.ErrNoActiveInstance) {
			WriteError(w, repository.ErrNoActiveInstance)
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
				ID:              sig.ID(),
				ActorUserID:     name,
				Decision:        string(sig.Decision()),
				Reason:          sig.Comment(),
				SignatureMethod: sig.SignatureMethod(),
				SignedAt:        sig.SignedAt().UTC().Format(time.RFC3339),
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
		Status:      string(inst.Status),
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
	if h.db == nil {
		return names, nil
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
	rows, err := h.db.QueryContext(
		ctx,
		`SELECT user_id, COALESCE(NULLIF(display_name, ''), user_id)
		   FROM metaldocs.iam_users
		  WHERE tenant_id = $1::uuid
		    AND user_id = ANY($2)`,
		tenantID,
		pq.Array(actorIDs),
	)
	if err != nil {
		return nil, fmt.Errorf("resolve eligible actors: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var actorID, displayName string
		if err := rows.Scan(&actorID, &displayName); err != nil {
			return nil, fmt.Errorf("resolve eligible actor row: %w", err)
		}
		names[actorID] = displayName
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("resolve eligible actors rows: %w", err)
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
		decision := string(sig.Decision())
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
