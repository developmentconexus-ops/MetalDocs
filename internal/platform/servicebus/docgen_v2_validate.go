package servicebus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const maxValidateResponseBodyBytes = 64 * 1024

// ValidateTemplate calls docgen-v2 /validate/template.
// Returns (true, raw, nil) on HTTP 200, (false, raw, nil) on HTTP 422,
// or (false, raw, err) for any other status or transport error.
func (c *DocgenV2Client) ValidateTemplate(ctx context.Context, docxKey, schemaKey string) (bool, []byte, error) {
	body, err := json.Marshal(map[string]string{"docx_key": docxKey, "schema_key": schemaKey})
	if err != nil {
		return false, nil, fmt.Errorf("docgen-v2: validate: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/validate/template", bytes.NewReader(body))
	if err != nil {
		return false, nil, fmt.Errorf("docgen-v2: validate: create request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-service-token", c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return false, nil, fmt.Errorf("docgen-v2: validate: do request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := readLimitedBody(resp.Body, maxValidateResponseBodyBytes)
	if err != nil {
		return false, nil, fmt.Errorf("docgen-v2: validate: read body: %w", err)
	}
	if resp.StatusCode == http.StatusOK {
		return true, raw, nil
	}
	if resp.StatusCode == http.StatusUnprocessableEntity {
		return false, raw, nil
	}
	return false, raw, fmt.Errorf("docgen-v2: unexpected status %d", resp.StatusCode)
}

func readLimitedBody(body io.Reader, maxBytes int64) ([]byte, error) {
	limited := io.LimitReader(body, maxBytes+1)
	payload, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) > maxBytes {
		return nil, fmt.Errorf("body exceeds %d-byte limit", maxBytes)
	}
	return payload, nil
}
