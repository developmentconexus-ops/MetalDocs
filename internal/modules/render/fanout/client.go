package fanout

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Request is the payload sent to docx-renderer's /render/fanout endpoint:
// the frozen revision's placeholder inputs plus the composition config needed
// to re-render it.
type Request struct {
	TenantID          string            `json:"tenant_id"`
	RevisionID        string            `json:"revision_id"`
	BodyDocxS3Key     string            `json:"body_docx_s3_key"`
	PlaceholderValues map[string]string `json:"placeholder_values"`
	Composition       json.RawMessage   `json:"composition_config"`
	ResolvedValues    map[string]any    `json:"resolved_values"`
}

// Response is docx-renderer's successful /render/fanout result: the
// content hash and storage key of the rendered document plus any placeholder
// variables that were left unreplaced.
type Response struct {
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

// Client calls the docx-renderer service's fanout HTTP endpoint.
type Client struct {
	baseURL      string
	serviceToken string
	http         *http.Client
}

// NewClient builds a fanout client. h is a required dependency: an injected,
// timeout-bound *http.Client (e.g. httpclient.NewInternalClient). Passing nil is
// a programmer error and panics rather than silently falling back to
// http.DefaultClient (which has no timeout and can hang indefinitely). Mirrors
// db.NewTxRunner's non-nil dependency contract.
func NewClient(baseURL, serviceToken string, h *http.Client) *Client {
	if h == nil {
		panic("fanout: NewClient requires a non-nil *http.Client")
	}
	return &Client{baseURL: baseURL, serviceToken: serviceToken, http: h}
}

// Fanout posts req to docx-renderer and returns the rendered document's
// content hash and storage key, or a classified *RenderError on failure.
func (c *Client) Fanout(ctx context.Context, req Request) (Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/render/fanout", bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	httpReq.Header.Set("content-type", "application/json")
	if c.serviceToken != "" {
		httpReq.Header.Set("X-Service-Token", c.serviceToken)
	}
	// Inject the active span's W3C traceparent (REQ-OBS-3: edge → api → outbox
	// → worker → docx-renderer). otel.GetTextMapPropagator() is the composite
	// TraceContext+Baggage propagator installed by SetupOTel when enabled; when
	// OTel is not installed (unconfigured default), the global propagator is
	// the OTel no-op and Inject is a harmless no-op — no header is added, no
	// error, matching SetupOTel's inert-by-default contract.
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(httpReq.Header))
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return Response{}, err
	}
	defer func() { _ = resp.Body.Close() }()
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
		return Response{}, &RenderError{
			Status:   resp.StatusCode,
			Kind:     classified.Kind,
			Message:  message,
			Variable: classified.Variable,
		}
	}
	var out Response
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Response{}, err
	}
	return out, nil
}
