package approvalhttp

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"metaldocs/internal/modules/documents/approval/http/contracts"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

func (h *Handler) InboxHandler(w http.ResponseWriter, r *http.Request) {
	reqID := requestID(r)
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, reqID, err)
		return
	}
	actorID := iamdomain.UserIDFromContext(r.Context())
	areaCode := strings.TrimSpace(r.URL.Query().Get("area_code"))

	limit, err := parseInboxLimit(r.URL.Query().Get("limit"))
	if err != nil {
		WriteError(w, reqID, err)
		return
	}

	offset, err := parseInboxOffset(r.URL.Query().Get("offset"))
	if err != nil {
		WriteError(w, reqID, err)
		return
	}

	if h.readSvc == nil {
		WriteError(w, reqID, errors.New("read service not configured"))
		return
	}

	views, err := h.readSvc.ListInboxItems(r.Context(), h.db, tenantID, actorID, areaCode, limit, offset)
	if err != nil {
		WriteError(w, reqID, err)
		return
	}

	total, err := h.readSvc.CountPendingForActor(r.Context(), h.db, tenantID, actorID, areaCode)
	if err != nil {
		WriteError(w, reqID, err)
		return
	}

	respItems := make([]contracts.InboxItem, 0, len(views))
	for i := range views {
		v := views[i]
		respItems = append(respItems, contracts.InboxItem{
			InstanceID:           v.InstanceID,
			DocumentID:           v.DocumentID,
			ControlledDocumentID: v.ControlledDocumentID,
			DocumentTitle:        v.DocumentTitle,
			AreaCode:             v.AreaCode,
			SubmittedBy:          v.SubmittedBy,
			SubmittedAt:          v.SubmittedAt.UTC().Format(time.RFC3339),
			StageLabel:           v.StageLabel,
			QuorumProgress:       v.QuorumProgress,
		})
	}

	WriteJSON(w, http.StatusOK, contracts.InboxResponse{
		Items: respItems,
		Total: total,
	})
}

func parseInboxLimit(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return 25, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("limit must be an integer")
	}
	if v <= 0 || v > 100 {
		return 0, fmt.Errorf("limit must be between 1 and 100")
	}
	return v, nil
}

func parseInboxOffset(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return 0, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("offset must be an integer")
	}
	if v < 0 {
		return 0, fmt.Errorf("offset must be >= 0")
	}
	return v, nil
}
