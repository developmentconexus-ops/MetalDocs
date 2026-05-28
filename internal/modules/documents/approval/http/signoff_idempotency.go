package approvalhttp

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"metaldocs/internal/modules/documents/approval/domain"
)

func signoffPayloadHash(documentID, instanceID, stageID string, decision domain.Decision, reason, contentHash string) string {
	payload := strings.Join([]string{
		strings.TrimSpace(documentID),
		strings.TrimSpace(instanceID),
		strings.TrimSpace(stageID),
		string(decision),
		strings.TrimSpace(reason),
		strings.TrimSpace(contentHash),
	}, "\n")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
