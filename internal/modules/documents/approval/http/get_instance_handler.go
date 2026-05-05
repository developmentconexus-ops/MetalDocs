package approvalhttp

import (
	"errors"
	"net/http"
	"time"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	"metaldocs/internal/modules/documents/approval/repository"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

func (h *Handler) GetInstanceHandler(w http.ResponseWriter, r *http.Request) {
	reqID := requestID(r)
	tenantID := tenantIDFromReq(r)
	actorID := iamdomain.UserIDFromContext(r.Context())
	instanceID := r.PathValue("instance_id")

	if h.readSvc == nil {
		WriteError(w, reqID, errors.New("read service not configured"))
		return
	}

	inst, err := h.readSvc.LoadInstance(r.Context(), h.db, tenantID, actorID, instanceID)
	if err != nil {
		if errors.Is(err, repository.ErrNoActiveInstance) {
			WriteError(w, reqID, repository.ErrNoActiveInstance)
			return
		}
		WriteError(w, reqID, err)
		return
	}

	resp := mapInstanceResponse(inst)
	w.Header().Set("ETag", "\"v1\"")
	WriteJSON(w, http.StatusOK, resp)
}

func mapInstanceResponse(inst *domain.Instance) contracts.InstanceResponse {
	var completedAt *string
	if inst.CompletedAt != nil {
		v := inst.CompletedAt.UTC().Format(time.RFC3339)
		completedAt = &v
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
		ETag:        "\"v1\"",
	}
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
