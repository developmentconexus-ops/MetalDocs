package application

import (
	"encoding/json"
	"fmt"
)

func marshalGovernancePayload(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal governance payload: %w", err)
	}
	return b, nil
}
