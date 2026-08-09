package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// ComputeValuesHash computes a deterministic sha256 hex digest over a
// placeholder values map, sorting keys first so the result is stable
// regardless of map iteration order.
func ComputeValuesHash(values map[string]any) (string, error) {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha256.New()
	for _, k := range keys {
		v, err := json.Marshal(values[k])
		if err != nil {
			return "", fmt.Errorf("marshal values hash: %w", err)
		}
		h.Write([]byte(k))
		h.Write([]byte{0})
		h.Write(v)
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
