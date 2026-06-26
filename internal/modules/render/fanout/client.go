package fanout

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type FanoutRequest struct {
	TenantID          string            `json:"tenant_id"`
	RevisionID        string            `json:"revision_id"`
	BodyDocxS3Key     string            `json:"body_docx_s3_key"`
	PlaceholderValues map[string]string `json:"placeholder_values"`
	Composition       json.RawMessage   `json:"composition_config"`
	ResolvedValues    map[string]any    `json:"resolved_values"`
}

type FanoutResponse struct {
	ContentHash    string   `json:"content_hash"`
	FinalDocxS3Key string   `json:"final_docx_s3_key"`
	UnreplacedVars []string `json:"unreplaced_vars"`
}

// RenderError is a classified failure returned by the docx-renderer.
type RenderError struct {
	Status   int
	Kind     string
	Message  string
	Variable string
}

func (e *RenderError) Error() string {
	if e.Variable != "" {
		return fmt.Sprintf("render failed (%s, status %d): %s [variable=%s]", e.Kind, e.Status, e.Message, e.Variable)
	}
	return fmt.Sprintf("render failed (%s, status %d): %s", e.Kind, e.Status, e.Message)
}

// Retryable reports whether the worker should retry. Template defects (4xx) are
// permanent; unknown/5xx failures are transient.
func (e *RenderError) Retryable() bool { return e.Status >= 500 }

type Client struct {
	baseURL      string
	serviceToken string
	http         *http.Client
}

func NewClient(baseURL, serviceToken string, h *http.Client) *Client {
	if h == nil {
		h = http.DefaultClient
	}
	return &Client{baseURL: baseURL, serviceToken: serviceToken, http: h}
}

func (c *Client) Fanout(ctx context.Context, req FanoutRequest) (FanoutResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return FanoutResponse{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/render/fanout", bytes.NewReader(body))
	if err != nil {
		return FanoutResponse{}, err
	}
	httpReq.Header.Set("content-type", "application/json")
	if c.serviceToken != "" {
		httpReq.Header.Set("X-Service-Token", c.serviceToken)
	}
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return FanoutResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		var classified struct {
			Kind     string `json:"kind"`
			Message  string `json:"message"`
			Variable string `json:"variable"`
		}
		_ = json.Unmarshal(errBody, &classified)
		message := classified.Message
		if classified.Kind == "" && message == "" {
			// Unclassified body (non-JSON, or a shape we don't model): keep the raw
			// payload as the message instead of dropping it, so the failure is
			// diagnosable rather than an empty "render failed (, status 5xx):".
			message = strings.TrimSpace(string(errBody))
		}
		return FanoutResponse{}, &RenderError{
			Status:   resp.StatusCode,
			Kind:     classified.Kind,
			Message:  message,
			Variable: classified.Variable,
		}
	}
	var out FanoutResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return FanoutResponse{}, err
	}
	return out, nil
}
